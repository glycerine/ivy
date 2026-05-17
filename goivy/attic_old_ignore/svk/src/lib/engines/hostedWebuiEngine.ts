import {
	normalizeCheckResult,
	normalizeConceptPayload,
	normalizeGraphPayload
} from '$lib/normalizers';
import { nowIso, stableId } from '$lib/normalizers/primitives';
import type {
	CheckMode,
	CommandIntent,
	CytoscapePayload,
	EngineEvent,
	EngineKind,
	EngineSession,
	Id,
	IvyEngine,
	ModelDocument,
	SnapshotBundle,
	SnapshotRequest,
	Unsubscribe,
	VerificationJob,
	VerificationJobKind
} from '$lib/types';

export type HostedWebuiEngineOptions = {
	baseUrl?: string;
	fetcher?: typeof fetch;
	createEventSource?: (url: string) => WebuiEventSourceLike;
	createId?: (prefix: string) => Id;
	now?: () => string;
};

export type WebuiEventSourceLike = {
	onmessage: ((event: MessageEvent<string>) => void) | null;
	onerror: ((event: Event) => void) | null;
	close(): void;
};

type RawWebuiEvent = {
	type: string;
	data?: unknown;
};

export class HostedWebuiEngine implements IvyEngine {
	readonly kind: EngineKind = 'hosted-go';

	private readonly baseUrl: string;
	private readonly fetcher: typeof fetch;
	private readonly createEventSource: (url: string) => WebuiEventSourceLike;
	private readonly createId: (prefix: string) => Id;
	private readonly now: () => string;
	private readonly eventSources = new Map<Id, WebuiEventSourceLike>();
	private readonly listeners = new Map<Id, Set<(event: EngineEvent) => void>>();
	private readonly sessions = new Map<Id, EngineSession>();
	private readonly latestModel = new Map<Id, ModelDocument>();
	private readonly activeJobBySession = new Map<Id, Id>();
	private sequence = 0;

	constructor(options: HostedWebuiEngineOptions = {}) {
		this.baseUrl = options.baseUrl ?? '';
		this.fetcher = options.fetcher ?? fetch;
		this.createEventSource = options.createEventSource ?? ((url) => new EventSource(url));
		this.createId =
			options.createId ??
			((prefix) => {
				this.sequence += 1;
				return `${prefix}-${this.sequence}`;
			});
		this.now = options.now ?? (() => nowIso());
	}

	async newSession(projectId: Id): Promise<EngineSession> {
		const data = await this.postJson<{ session_id?: string }>('/api/session/new');
		const timestamp = this.now();
		const session: EngineSession = {
			id: requireString(data.session_id, 'webui session_id'),
			kind: this.kind,
			status: 'ready',
			capabilities: {
				offline: false,
				persistentJobs: false,
				cancelJob: false,
				eventStream: true,
				parallelJobs: false
			},
			projectId,
			createdAt: timestamp,
			updatedAt: timestamp
		};
		this.sessions.set(session.id, session);
		return session;
	}

	async closeSession(sessionId: Id): Promise<void> {
		this.eventSources.get(sessionId)?.close();
		this.eventSources.delete(sessionId);
		this.listeners.delete(sessionId);
		this.sessions.delete(sessionId);
		this.latestModel.delete(sessionId);
		this.activeJobBySession.delete(sessionId);
	}

	async loadModel(sessionId: Id, model: ModelDocument): Promise<VerificationJob> {
		const session = this.requireSession(sessionId);
		this.latestModel.set(sessionId, model);
		session.currentModelId = model.id;
		session.currentModelRevision = model.engineRevision;
		session.updatedAt = this.now();
		this.emit(sessionId, { type: 'session-ready', session });

		const job = this.createJob(session, model, 'load');
		this.startJob(sessionId, job, 'uploading');
		try {
			const form = new FormData();
			form.append('file', new Blob([model.text], { type: 'text/plain' }), model.filename);
			await this.postForm(`/api/session/${encodeURIComponent(sessionId)}/load`, form);
			await this.refreshSnapshots(sessionId, model.engineRevision);
			this.finishJob(sessionId, job);
			return job;
		} catch (error) {
			this.failJob(sessionId, job, error);
			throw error;
		}
	}

	async runCommand(intent: CommandIntent): Promise<VerificationJob> {
		const session = this.requireSession(intent.sessionId);
		const model = this.latestModel.get(intent.sessionId) ?? modelFromSession(session);
		const job = this.createJob(session, model, commandToJobKind(intent.commandId));
		this.startJob(intent.sessionId, job, 'dispatching');
		try {
			await this.dispatchCommand(intent, job);
			this.finishJob(intent.sessionId, job);
			return job;
		} catch (error) {
			this.failJob(intent.sessionId, job, error);
			throw error;
		}
	}

	async cancelJob(jobId: Id): Promise<void> {
		void jobId;
		throw new Error('The current webui backend does not support job cancellation');
	}

	async getSnapshot(sessionId: Id, request: SnapshotRequest): Promise<SnapshotBundle> {
		void request;
		const revision = this.latestModel.get(sessionId)?.engineRevision ?? 0;
		const [arg, concept] = await Promise.all([
			this.getJson<unknown>(`/api/session/${encodeURIComponent(sessionId)}/arg`),
			this.getJson<unknown>(`/api/session/${encodeURIComponent(sessionId)}/concept`)
		]);
		return {
			graphs: [
				normalizeGraphPayload(arg as CytoscapePayload, {
					id: stableId('graph', sessionId, 'arg', revision),
					sheetId: 'sheet-1',
					kind: 'arg',
					sourceRevision: revision
				})
			],
			concepts: [normalizeConceptPayload(concept, { sessionId, sheetId: 'sheet-1', revision })]
		};
	}

	subscribe(sessionId: Id, onEvent: (event: EngineEvent) => void): Unsubscribe {
		const listeners = this.listeners.get(sessionId) ?? new Set();
		listeners.add(onEvent);
		this.listeners.set(sessionId, listeners);

		if (!this.eventSources.has(sessionId)) {
			this.openEventStream(sessionId);
		}

		return () => {
			listeners.delete(onEvent);
			if (listeners.size === 0) {
				this.listeners.delete(sessionId);
				this.eventSources.get(sessionId)?.close();
				this.eventSources.delete(sessionId);
			}
		};
	}

	private async dispatchCommand(intent: CommandIntent, job: VerificationJob) {
		const sessionId = intent.sessionId;
		switch (intent.commandId) {
			case 'check.induction':
				await this.runCheck(sessionId, job, 'induction');
				break;
			case 'check.bounded':
				await this.runCheck(sessionId, job, 'bounded', numberArg(intent.args?.bound));
				break;
			case 'check.pdr':
				await this.runCheck(sessionId, job, 'pdr');
				break;
			case 'check.concrete':
				await this.runCheck(sessionId, job, 'concrete');
				break;
			case 'check.abstract':
				await this.runCheck(sessionId, job, 'abstract');
				break;
			case 'concept.reset':
				await this.postJson(`/api/session/${encodeURIComponent(sessionId)}/concept/reset`);
				await this.refreshSnapshots(sessionId, job.modelRevision);
				break;
			case 'concept.diagram':
				await this.postJson(`/api/session/${encodeURIComponent(sessionId)}/concept/diagram`);
				await this.refreshSnapshots(sessionId, job.modelRevision);
				break;
			default:
				await this.runAction(intent);
				await this.refreshSnapshots(sessionId, job.modelRevision);
				break;
		}
	}

	private async runCheck(sessionId: Id, job: VerificationJob, mode: CheckMode, bound = 0) {
		const payload: Record<string, unknown> = { mode };
		if (bound > 0) {
			payload.bound = bound;
		}
		const resultPayload = await this.postJson<unknown>(`/api/session/${encodeURIComponent(sessionId)}/check`, payload);
		const result = normalizeCheckResult(resultPayload, { jobId: job.id, sessionId, mode });
		this.emit(sessionId, { type: 'check-updated', result });
		const traceArg = getTraceArg(resultPayload);
		if (traceArg) {
			this.emit(sessionId, {
				type: 'graph-updated',
				snapshot: normalizeGraphPayload(traceArg, {
					id: stableId('graph', sessionId, 'trace', job.id),
					sheetId: 'trace-arg',
					kind: 'arg',
					sourceRevision: job.modelRevision
				})
			});
		}
		await this.refreshSnapshots(sessionId, job.modelRevision);
	}

	private async runAction(intent: CommandIntent) {
		if (intent.target?.nodeId || intent.target?.obj || intent.commandId.startsWith('arg.')) {
			await this.postJson(`/api/session/${encodeURIComponent(intent.sessionId)}/arg/action`, {
				node: intent.target?.obj ?? intent.target?.nodeId ?? '',
				action: intent.commandId,
				args: intent.args ?? {}
			});
			return;
		}
		await this.postJson(`/api/session/${encodeURIComponent(intent.sessionId)}/action`, {
			action: intent.commandId,
			args: intent.args ?? {}
		});
	}

	private async refreshSnapshots(sessionId: Id, revision: number) {
		const bundle = await this.getSnapshot(sessionId, {});
		for (const graph of bundle.graphs ?? []) {
			graph.sourceRevision = revision;
			this.emit(sessionId, { type: 'graph-updated', snapshot: graph });
		}
		for (const concept of bundle.concepts ?? []) {
			concept.revision = revision;
			this.emit(sessionId, { type: 'concept-updated', concept });
			this.emit(sessionId, { type: 'toggles-updated', sheetId: concept.sheetId, toggles: concept.toggles });
		}
	}

	private openEventStream(sessionId: Id) {
		const source = this.createEventSource(this.url(`/api/session/${encodeURIComponent(sessionId)}/events`));
		source.onmessage = (event) => {
			this.translateRawEvent(sessionId, JSON.parse(event.data) as RawWebuiEvent);
		};
		source.onerror = () => {
			this.emit(sessionId, {
				type: 'diagnostic',
				severity: 'warning',
				message: 'Hosted webui event stream disconnected'
			});
		};
		this.eventSources.set(sessionId, source);
	}

	private translateRawEvent(sessionId: Id, raw: RawWebuiEvent) {
		switch (raw.type) {
			case 'check_started': {
				const jobId = this.activeJobBySession.get(sessionId);
				if (jobId) {
					this.emit(sessionId, { type: 'job-progress', jobId, progress: { phase: 'checking', done: 0, total: 1 } });
				}
				break;
			}
			case 'check_completed': {
				const jobId = this.activeJobBySession.get(sessionId);
				if (jobId) {
					this.emit(sessionId, { type: 'job-progress', jobId, progress: { phase: 'checking', done: 1, total: 1 } });
				}
				break;
			}
			case 'file_loaded':
				this.emit(sessionId, { type: 'diagnostic', severity: 'info', message: 'Model loaded by hosted webui backend' });
				break;
			default:
				this.emit(sessionId, { type: 'diagnostic', severity: 'info', message: `Hosted webui event: ${raw.type}` });
		}
	}

	private createJob(session: EngineSession, model: ModelDocument, kind: VerificationJobKind): VerificationJob {
		const timestamp = this.now();
		return {
			id: this.createId('job'),
			sessionId: session.id,
			engineId: session.id,
			projectId: session.projectId,
			modelId: model.id,
			modelRevision: model.engineRevision,
			kind,
			status: 'queued',
			createdAt: timestamp,
			updatedAt: timestamp
		};
	}

	private startJob(sessionId: Id, job: VerificationJob, phase: string) {
		this.activeJobBySession.set(sessionId, job.id);
		this.emit(sessionId, { type: 'job-created', job });
		this.emit(sessionId, { type: 'job-progress', jobId: job.id, progress: { phase } });
	}

	private finishJob(sessionId: Id, job: VerificationJob) {
		this.emit(sessionId, { type: 'job-succeeded', jobId: job.id, result: { ok: true } });
		this.activeJobBySession.delete(sessionId);
	}

	private failJob(sessionId: Id, job: VerificationJob, error: unknown) {
		this.emit(sessionId, {
			type: 'job-failed',
			jobId: job.id,
			error: error instanceof Error ? error.message : 'Hosted webui operation failed'
		});
		this.activeJobBySession.delete(sessionId);
	}

	private emit(sessionId: Id, event: EngineEvent) {
		for (const listener of this.listeners.get(sessionId) ?? []) {
			listener(event);
		}
	}

	private requireSession(sessionId: Id): EngineSession {
		const session = this.sessions.get(sessionId);
		if (!session) {
			throw new Error(`Unknown hosted webui session: ${sessionId}`);
		}
		return session;
	}

	private async getJson<T>(path: string): Promise<T> {
		const response = await this.fetcher(this.url(path), {
			credentials: 'include',
			headers: { accept: 'application/json' }
		});
		return readResponse<T>(response);
	}

	private async postJson<T = unknown>(path: string, body?: unknown): Promise<T> {
		const response = await this.fetcher(this.url(path), {
			method: 'POST',
			credentials: 'include',
			headers: body === undefined ? { accept: 'application/json' } : { accept: 'application/json', 'content-type': 'application/json' },
			body: body === undefined ? undefined : JSON.stringify(body)
		});
		return readResponse<T>(response);
	}

	private async postForm<T = unknown>(path: string, body: FormData): Promise<T> {
		const response = await this.fetcher(this.url(path), {
			method: 'POST',
			credentials: 'include',
			headers: { accept: 'application/json' },
			body
		});
		return readResponse<T>(response);
	}

	private url(path: string) {
		return `${this.baseUrl}${path}`;
	}
}

async function readResponse<T>(response: Response): Promise<T> {
	if (!response.ok) {
		let message = `Hosted webui request failed with ${response.status}`;
		try {
			const payload = (await response.json()) as { error?: string };
			message = payload.error ?? message;
		} catch {
			// Keep the status message when a failing response is not JSON.
		}
		throw new Error(message);
	}
	const contentType = response.headers.get('content-type') ?? '';
	if (contentType.includes('application/json')) {
		return (await response.json()) as T;
	}
	return (await response.text()) as T;
}

function commandToJobKind(commandId: string): VerificationJobKind {
	switch (commandId) {
		case 'check.induction':
			return 'check-induction';
		case 'check.bounded':
			return 'check-bounded';
		case 'check.pdr':
			return 'check-pdr';
		case 'check.concrete':
			return 'check-concrete';
		case 'check.abstract':
			return 'check-abstract';
		case 'concept.reset':
		case 'concept.diagram':
		case 'concept.action':
			return 'concept-action';
		case 'proof.action':
			return 'proof-action';
		default:
			return 'arg-action';
	}
}

function modelFromSession(session: EngineSession): ModelDocument {
	const timestamp = session.updatedAt;
	return {
		id: session.currentModelId ?? 'model-unknown',
		projectId: session.projectId,
		filename: 'model.ivy',
		text: '',
		dirty: false,
		parseRevision: session.currentModelRevision ?? 0,
		engineRevision: session.currentModelRevision ?? 0,
		createdAt: timestamp,
		updatedAt: timestamp
	};
}

function requireString(value: unknown, label: string): string {
	if (typeof value !== 'string' || value === '') {
		throw new Error(`Missing ${label}`);
	}
	return value;
}

function numberArg(value: unknown): number {
	return typeof value === 'number' && Number.isFinite(value) ? value : 0;
}

function getTraceArg(payload: unknown): unknown | null {
	return payload && typeof payload === 'object' && 'trace_arg' in payload
		? (payload as { trace_arg?: unknown }).trace_arg ?? null
		: null;
}
