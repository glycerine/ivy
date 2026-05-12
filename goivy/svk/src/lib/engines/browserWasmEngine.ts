import type {
	CommandIntent,
	EngineEvent,
	EngineKind,
	EngineSession,
	Id,
	IvyEngine,
	ModelDocument,
	SnapshotBundle,
	SnapshotRequest,
	Unsubscribe,
	VerificationJob
} from '$lib/types';
import {
	parseWorkerResponse,
	serializeWorkerRequest,
	type WorkerRequest,
	type WorkerResponse
} from './workerProtocol';

export type WorkerLike = {
	postMessage(message: string): void;
	terminate(): void;
	addEventListener(type: 'message', listener: (event: MessageEvent<string>) => void): void;
	removeEventListener(type: 'message', listener: (event: MessageEvent<string>) => void): void;
};

export type BrowserWasmEngineOptions = {
	assetBaseUrl?: string;
	createWorker?: () => WorkerLike;
	createId?: () => Id;
};

type PendingRequest = {
	resolve: (value: WorkerResponse) => void;
	reject: (reason: Error) => void;
};

export class BrowserWasmEngine implements IvyEngine {
	readonly kind: EngineKind = 'browser-js-wasm';

	private readonly assetBaseUrl: string;
	private readonly createId: () => Id;
	private readonly worker: WorkerLike;
	private readonly pending = new Map<Id, PendingRequest>();
	private readonly listeners = new Map<Id, Set<(event: EngineEvent) => void>>();
	private initialized: Promise<void> | null = null;

	constructor(options: BrowserWasmEngineOptions = {}) {
		this.assetBaseUrl = options.assetBaseUrl ?? '/wasm/';
		this.createId = options.createId ?? (() => crypto.randomUUID());
		this.worker = options.createWorker?.() ?? createDefaultWorker();
		this.worker.addEventListener('message', (event) => this.handleMessage(event));
	}

	async init(): Promise<void> {
		this.initialized ??= this.send({ type: 'init', requestId: this.createId(), assetBaseUrl: this.assetBaseUrl }).then(() => undefined);
		return this.initialized;
	}

	async newSession(projectId: Id): Promise<EngineSession> {
		await this.init();
		const response = await this.send({ type: 'new-session', requestId: this.createId(), projectId });
		if (response.type !== 'session') {
			throw new Error(`Expected session response, got ${response.type}`);
		}
		return response.session;
	}

	async closeSession(sessionId: Id): Promise<void> {
		this.listeners.delete(sessionId);
	}

	async loadModel(sessionId: Id, model: ModelDocument): Promise<VerificationJob> {
		await this.init();
		const response = await this.send({ type: 'load-model', requestId: this.createId(), sessionId, model });
		if (response.type !== 'job') {
			throw new Error(`Expected job response, got ${response.type}`);
		}
		return response.job;
	}

	async runCommand(intent: CommandIntent): Promise<VerificationJob> {
		await this.init();
		const response = await this.send({ type: 'run-command', requestId: this.createId(), intent });
		if (response.type !== 'job') {
			throw new Error(`Expected job response, got ${response.type}`);
		}
		return response.job;
	}

	async cancelJob(jobId: Id): Promise<void> {
		await this.init();
		await this.send({ type: 'cancel-job', requestId: this.createId(), jobId });
	}

	async getSnapshot(sessionId: Id, request: SnapshotRequest): Promise<SnapshotBundle> {
		void sessionId;
		void request;
		return {};
	}

	subscribe(sessionId: Id, onEvent: (event: EngineEvent) => void): Unsubscribe {
		const listeners = this.listeners.get(sessionId) ?? new Set();
		listeners.add(onEvent);
		this.listeners.set(sessionId, listeners);
		return () => {
			listeners.delete(onEvent);
			if (listeners.size === 0) {
				this.listeners.delete(sessionId);
			}
		};
	}

	terminate() {
		this.worker.terminate();
		this.pending.clear();
		this.listeners.clear();
	}

	private send(request: WorkerRequest): Promise<WorkerResponse> {
		return new Promise((resolve, reject) => {
			this.pending.set(request.requestId, { resolve, reject });
			this.worker.postMessage(serializeWorkerRequest(request));
		});
	}

	private handleMessage(event: MessageEvent<string>) {
		const response = parseWorkerResponse(event.data);
		if (response.type === 'event') {
			for (const listener of this.listeners.get(response.sessionId) ?? []) {
				listener(response.event);
			}
			return;
		}

		if (response.type === 'error') {
			if (response.requestId) {
				const pending = this.pending.get(response.requestId);
				this.pending.delete(response.requestId);
				pending?.reject(new Error(response.error));
			}
			return;
		}

		const pending = this.pending.get(response.requestId);
		this.pending.delete(response.requestId);
		pending?.resolve(response);
	}
}

function createDefaultWorker(): WorkerLike {
	return new Worker(new URL('../../workers/ivyEngine.worker.ts', import.meta.url), { type: 'module' });
}
