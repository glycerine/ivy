import { IvyApiAdapter } from './ivyApiAdapter.ts';

type AnyRecord = Record<string, any>;
type PendingRequest = {
  resolve: (value: any) => void;
  reject: (reason?: any) => void;
  abortHandler?: () => void;
  signal?: AbortSignal;
};

const DEFAULT_ASSET_BASE_URL = '/static/wasm/';

function abortError(message = 'Browser WASM job cancelled') {
  if (typeof DOMException !== 'undefined') {
    return new DOMException(message, 'AbortError');
  }
  const err = new Error(message);
  (err as any).name = 'AbortError';
  return err;
}

function workerMessageData(data: any): AnyRecord {
  if (typeof data === 'string') return JSON.parse(data);
  return data || {};
}

function isCheckCommand(commandId: string): boolean {
  return !!commandId && commandId.startsWith('check.');
}

function modeFromCommand(commandId: string, args: AnyRecord = {}) {
  if (isCheckCommand(commandId)) return commandId.slice('check.'.length);
  return args.mode || 'pdr';
}

async function normalizeModelDocument(model: AnyRecord = {}, revision: number): Promise<AnyRecord> {
  if (model.file && typeof model.file.text === 'function') {
    const text = await model.file.text();
    return {
      id: model.id || 'model',
      projectId: model.projectId || 'webui',
      filename: model.filename || model.file.name || 'model.ivy',
      text,
      isolate: model.isolate || '',
      engineRevision: revision,
    };
  }
  return {
    id: model.id || 'model',
    projectId: model.projectId || 'webui',
    filename: model.filename || model.path || 'model.ivy',
    text: model.content || model.text || '',
    isolate: model.isolate || '',
    engineRevision: revision,
  };
}

export class BrowserWasmIvyApiAdapter extends IvyApiAdapter {
  [key: string]: any;
  assetBaseUrl: string;
  includeRoot: string;
  projectId: string;
  workerFactory: () => Worker;
  worker: Worker | null;
  readyPromise: Promise<void> | null;
  pending: Map<string, PendingRequest>;
  requestSeq: number;
  modelRevision: number;
  lastModel: AnyRecord | null;
  eventListeners: Set<(event: any) => void>;
  sessionId: string | null;

  constructor({
    assetBaseUrl = DEFAULT_ASSET_BASE_URL,
    includeRoot = '/include',
    projectId = 'webui',
    workerFactory,
  }: {
    assetBaseUrl?: string;
    includeRoot?: string;
    projectId?: string;
    workerFactory?: () => Worker;
  } = {}) {
    super();
    this.kind = 'browser-wasm';
    this.capabilities = {
      offline: true,
      persistentJobs: false,
      cancelJob: true,
      eventStream: true,
      parallelJobs: false,
    };
    this.assetBaseUrl = assetBaseUrl.endsWith('/') ? assetBaseUrl : `${assetBaseUrl}/`;
    this.includeRoot = includeRoot;
    this.projectId = projectId;
    this.workerFactory = workerFactory || (() => new Worker(
      new URL('../workers/browserWasmEngine.worker.js', import.meta.url),
      { type: 'module' },
    ));
    this.worker = null;
    this.readyPromise = null;
    this.pending = new Map();
    this.requestSeq = 0;
    this.modelRevision = 0;
    this.lastModel = null;
    this.eventListeners = new Set();
    this.sessionId = null;
  }

  async createSession(): Promise<string | null> {
    await this.ensureWorker();
    const session = await this.send('new-session', { projectId: this.projectId });
    this.sessionId = session && (session.id || session.session_id || session.sessionId);
    if (!this.sessionId) throw new Error('Browser WASM engine returned no session id');
    return this.sessionId;
  }

  async loadModel(model: AnyRecord = {}, requestOptions: RequestInit = {}): Promise<any> {
    this.modelRevision += 1;
    this.lastModel = await normalizeModelDocument(model, this.modelRevision);
    await this.ensureSession({ hydrate: false });
    return this.send('load-model', {
      sessionId: this.sessionId,
      model: this.lastModel,
    }, requestOptions);
  }

  async runCommand(intent: AnyRecord = {}, requestOptions: RequestInit = {}): Promise<any> {
    await this.ensureSession({ hydrate: true });
    const commandId = intent.commandId || '';
    const args = intent.args || {};
    const normalizedIntent = {
      ...intent,
      args: isCheckCommand(commandId) ? { ...args, mode: modeFromCommand(commandId, args) } : args,
    };
    return this.send('run-command', {
      sessionId: this.sessionId,
      intent: normalizedIntent,
    }, requestOptions);
  }

  async getSnapshot(request: AnyRecord = {}, requestOptions: RequestInit = {}): Promise<any> {
    await this.ensureSession({ hydrate: true });
    return this.send('get-snapshot', {
      sessionId: this.sessionId,
      snapshot: request || {},
    }, requestOptions);
  }

  subscribe(onEvent: (event: any) => void, _onError?: (error?: any) => void) {
    this.eventListeners.add(onEvent);
    return () => {
      this.eventListeners.delete(onEvent);
    };
  }

  async cancelJob(_jobId?: any): Promise<any> {
    this.terminateWorker(abortError());
    return { status: 'cancelled' };
  }

  async saveSession(): Promise<Blob> {
    const payload = {
      backend: this.kind,
      sessionId: this.sessionId,
      model: this.lastModel,
    };
    return new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
  }

  async ensureSession({ hydrate = true } = {}) {
    if (!this.sessionId) {
      await this.createSession();
      if (hydrate && this.lastModel) {
        await this.send('load-model', {
          sessionId: this.sessionId,
          model: this.lastModel,
        });
      }
    }
  }

  async ensureWorker() {
    if (!this.worker) {
      this.worker = this.workerFactory();
      this.worker.onmessage = (event) => this.handleWorkerMessage(event);
      this.worker.onerror = (event) => {
        const message = event && event.message ? event.message : 'Browser WASM worker failed';
        this.terminateWorker(new Error(message));
      };
      this.readyPromise = this.send('init', {
        assetBaseUrl: this.assetBaseUrl,
        includeRoot: this.includeRoot,
      }).then(() => undefined);
    }
    return this.readyPromise;
  }

  send(type: string, payload: AnyRecord = {}, requestOptions: RequestInit = {}): Promise<any> {
    const signal = requestOptions && requestOptions.signal as AbortSignal | undefined;
    if (signal && signal.aborted) {
      return Promise.reject(abortError());
    }
    if (!this.worker && type !== 'init') {
      return this.ensureWorker().then(() => this.send(type, payload, requestOptions));
    }
    if (!this.worker) {
      throw new Error('Browser WASM worker is not available');
    }
    const requestId = this.nextRequestId(type);
    const message = { type, requestId, ...payload };
    return new Promise((resolve, reject) => {
      const pending: PendingRequest = { resolve, reject, signal };
      if (signal) {
        pending.abortHandler = () => {
          this.terminateWorker(abortError());
        };
        signal.addEventListener('abort', pending.abortHandler, { once: true });
      }
      this.pending.set(requestId, pending);
      this.worker!.postMessage(message);
    });
  }

  handleWorkerMessage(event: MessageEvent) {
    const message = workerMessageData(event.data);
    if (message.type === 'event') {
      this.emitEvent(message.event || message.data || message);
      return;
    }
    if (message.type === 'stdout' || message.type === 'stderr') {
      this.emitEvent({
        type: 'job_log',
        data: {
          stream: message.type,
          message: message.value || '',
        },
      });
      return;
    }
    const requestId = message.requestId;
    if (!requestId) return;
    const pending = this.pending.get(requestId);
    if (!pending) return;
    this.pending.delete(requestId);
    if (pending.signal && pending.abortHandler) {
      pending.signal.removeEventListener('abort', pending.abortHandler);
    }
    if (message.type === 'error') {
      pending.reject(new Error(message.error || 'Browser WASM worker failed'));
      return;
    }
    pending.resolve(message.value);
  }

  emitEvent(event: any) {
    for (const listener of Array.from(this.eventListeners)) {
      listener(event);
    }
  }

  terminateWorker(reason: any = abortError()) {
    if (this.worker && typeof this.worker.terminate === 'function') {
      this.worker.terminate();
    }
    this.worker = null;
    this.readyPromise = null;
    this.sessionId = null;
    const pending = Array.from(this.pending.values());
    this.pending.clear();
    for (const request of pending) {
      if (request.signal && request.abortHandler) {
        request.signal.removeEventListener('abort', request.abortHandler);
      }
      request.reject(reason);
    }
  }

  nextRequestId(prefix: string) {
    this.requestSeq += 1;
    return `${prefix}-${this.requestSeq}`;
  }
}
