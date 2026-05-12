import { IvyHttpClient } from '../api/ivyHttpClient.js';
import { IvyApiAdapter } from './ivyApiAdapter.js';

function toFormDataFromModel(model = {}) {
  const formData = new FormData();
  if (model.file) {
    formData.append('file', model.file);
    return formData;
  }
  const content = model.content || '';
  const filename = model.filename || model.path || 'model.ivy';
  const blob = new Blob([content], { type: 'text/plain' });
  formData.append('file', blob, filename);
  return formData;
}

function jsonPost(body = {}) {
  return {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  };
}

function conceptNodeParam(request = {}) {
  if (request.nodeId != null) return request.nodeId;
  if (request.stateId != null) return request.stateId;
  return undefined;
}

function commandMode(commandId, args = {}) {
  if (commandId && commandId.startsWith('check.')) {
    return commandId.slice('check.'.length);
  }
  return args.mode || 'pdr';
}

export class HostedGoIvyApiAdapter extends IvyApiAdapter {
  constructor({
    baseURL = '',
    client = new IvyHttpClient({ baseURL }),
    eventSourceFactory = (url) => new EventSource(url),
  } = {}) {
    super();
    this.kind = 'hosted-go';
    this.capabilities = {
      offline: false,
      persistentJobs: false,
      cancelJob: false,
      eventStream: true,
      parallelJobs: false,
    };
    this.client = client;
    this.eventSourceFactory = eventSourceFactory;
    this.sessionId = null;
    this.eventSource = null;
    this._sseRetries = 0;
    this._sseMaxRetries = 5;
    this._sseClosed = true;
    this._sseTimer = null;
    this._sseOnEvent = null;
    this._sseOnError = null;
  }

  sessionPath(suffix) {
    if (!this.sessionId) {
      throw new Error('No Ivy session has been created yet');
    }
    return `/api/session/${encodeURIComponent(this.sessionId)}${suffix}`;
  }

  async createSession() {
    const data = await this.client.request('/api/session/new', { method: 'POST' });
    this.sessionId = data.session_id;
    return this.sessionId;
  }

  async loadModel(model = {}) {
    return this.client.request(this.sessionPath('/load'), {
      method: 'POST',
      body: toFormDataFromModel(model),
    });
  }

  async runCommand(intent = {}) {
    const { commandId = '', args = {}, target = {} } = intent;

    if (commandId.startsWith('check.')) {
      const mode = commandMode(commandId, args);
      const body = { ...args, mode };
      return this.client.request(this.sessionPath('/check'), jsonPost(body));
    }

    switch (commandId) {
      case 'concept.split':
        return this.client.request(this.sessionPath('/concept/split'), jsonPost({
          concept: args.concept,
          split_by: args.splitBy != null ? args.splitBy : args.split_by,
        }));
      case 'concept.empty':
        return this.client.request(this.sessionPath('/concept/empty'), jsonPost({ concept: args.concept }));
      case 'concept.remove':
        return this.client.request(this.sessionPath('/concept/remove'), jsonPost({ concept: args.concept }));
      case 'concept.undo':
        return this.client.request(this.sessionPath('/concept/undo'), { method: 'POST' });
      case 'concept.materializeNode':
        return this.client.request(this.sessionPath('/concept/materialize'), jsonPost({
          concept: args.concept,
          type: 'node',
        }));
      case 'concept.materializeEdge':
        return this.client.request(this.sessionPath('/concept/materialize'), jsonPost({
          relation: args.relation,
          source: args.source,
          target: args.target,
          type: 'edge',
          positive: args.positive,
        }));
      case 'concept.projection':
        return this.client.request(this.sessionPath('/concept/projection'), jsonPost({
          name: args.name,
          concept: args.concept,
        }));
      case 'concept.reset':
        return this.client.request(this.sessionPath('/concept/reset'), { method: 'POST' });
      case 'concept.diagram':
        return this.client.request(this.sessionPath('/concept/diagram'), { method: 'POST' });
      case 'toggles.set':
        return this.client.request(this.sessionPath('/toggles'), jsonPost(args));
      case 'proof.action':
        return this.client.request(this.sessionPath('/proof/action'), jsonPost({
          goal: target.goalId,
          action: args.action,
        }));
      default:
        if (target.kind === 'arg' || target.nodeId != null || target.obj != null) {
          return this.client.request(this.sessionPath('/arg/action'), jsonPost({
            node: target.obj != null ? target.obj : target.nodeId,
            action: commandId,
            args,
          }));
        }
        return this.client.request(this.sessionPath('/action'), jsonPost({ action: commandId, args }));
    }
  }

  async getSnapshot(request = {}) {
    const snapshot = {};
    const includeDefault = Object.keys(request || {}).length === 0;

    if (includeDefault || request.arg) {
      const argRequest = typeof request.arg === 'object' ? request.arg : {};
      snapshot.arg = await this.getArgSnapshot(argRequest);
    }

    if (includeDefault || request.concept) {
      const conceptRequest = typeof request.concept === 'object' ? request.concept : {};
      snapshot.concept = await this.getConceptSnapshot(conceptRequest);
    }

    if (request.menus) {
      snapshot.menus = await this.client.request(this.sessionPath('/menus'));
    }

    if (request.toggles) {
      snapshot.toggles = await this.client.request(this.sessionPath('/toggles'));
    }

    if (request.proof) {
      snapshot.proof = await this.client.request(this.sessionPath('/proof'));
    }

    return snapshot;
  }

  async getArgSnapshot({ sheetId } = {}) {
    const suffix = sheetId ? `/arg?sheet=${encodeURIComponent(sheetId)}` : '/arg';
    return this.client.request(this.sessionPath(suffix));
  }

  async getConceptSnapshot(request = {}) {
    const params = new URLSearchParams();
    if (request.sheetId) params.set('sheet', request.sheetId);
    const selectedNode = conceptNodeParam(request);
    if (selectedNode != null) params.set('node', String(selectedNode));
    const query = params.toString();
    return this.client.request(this.sessionPath(`/concept${query ? `?${query}` : ''}`));
  }

  async saveSession() {
    const response = await this.client.fetchImpl(this.client.baseURL + this.sessionPath('/save'), {});
    if (!response.ok) {
      throw new Error(`Save failed: ${response.statusText || response.status || 'unknown error'}`);
    }
    return response.blob();
  }

  subscribe(onEvent, onError) {
    if (!this.sessionId) {
      throw new Error('No Ivy session has been created yet');
    }
    this._closeEventSource();
    this._sseRetries = 0;
    this._sseClosed = false;
    this._sseOnEvent = onEvent;
    this._sseOnError = onError;
    this._sseConnect();
    return () => this._closeEventSource();
  }

  _sseConnect() {
    if (this._sseClosed || !this.sessionId) return;
    const source = this.eventSourceFactory(this.sessionPath('/events'));
    this.eventSource = source;

    source.onopen = () => {
      this._sseRetries = 0;
    };

    source.onmessage = (event) => {
      try {
        if (this._sseOnEvent) this._sseOnEvent(JSON.parse(event.data));
      } catch (err) {
        console.error('Failed to parse SSE event:', err, event.data);
      }
    };

    source.onerror = () => {
      if (typeof source.close === 'function') source.close();
      if (this._sseClosed) return;
      this._sseRetries += 1;
      if (this._sseRetries > this._sseMaxRetries) {
        console.error('SSE: max retries exceeded, giving up');
        if (this._sseOnError) this._sseOnError();
        return;
      }
      const delay = Math.min(1000 * Math.pow(2, this._sseRetries - 1), 16000);
      console.warn(`SSE: reconnect attempt ${this._sseRetries}/${this._sseMaxRetries} in ${delay}ms`);
      this._sseTimer = globalThis.setTimeout(() => {
        this._sseConnect();
      }, delay);
    };
  }

  _closeEventSource() {
    this._sseClosed = true;
    if (this._sseTimer) {
      globalThis.clearTimeout(this._sseTimer);
      this._sseTimer = null;
    }
    if (this.eventSource && typeof this.eventSource.close === 'function') {
      this.eventSource.close();
    }
    this.eventSource = null;
  }
}
