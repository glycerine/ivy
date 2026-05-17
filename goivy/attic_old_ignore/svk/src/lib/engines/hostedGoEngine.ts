import type {
	CheckMode,
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
	normalizeCheckResult,
	normalizeConceptPayload,
	normalizeGraphPayload,
	stableId
} from '$lib/normalizers';

export type HostedGoEngineOptions = {
	baseUrl?: string;
	fetcher?: typeof fetch;
};

type JsonResponse<T> = {
	ok: true;
	value: T;
};

export class HostedGoEngine implements IvyEngine {
	readonly kind: EngineKind = 'hosted-go';

	private readonly baseUrl: string;
	private readonly fetcher: typeof fetch;
	private readonly listeners = new Map<Id, Set<(event: EngineEvent) => void>>();
	private readonly sessions = new Map<Id, EngineSession>();
	private readonly latestRevision = new Map<Id, number>();

	constructor(options: HostedGoEngineOptions = {}) {
		this.baseUrl = options.baseUrl ?? '';
		this.fetcher = options.fetcher ?? fetch;
	}

	async newSession(projectId: Id): Promise<EngineSession> {
		const { session } = await this.post<{ session: EngineSession }>('/api/engine/session', { projectId });
		this.sessions.set(session.id, session);
		return session;
	}

	async closeSession(sessionId: Id): Promise<void> {
		this.listeners.delete(sessionId);
		this.sessions.delete(sessionId);
		this.latestRevision.delete(sessionId);
	}

	async loadModel(sessionId: Id, model: ModelDocument): Promise<VerificationJob> {
		const { job } = await this.post<{ job: VerificationJob }>(
			`/api/engine/session/${encodeURIComponent(sessionId)}/load`,
			{ model }
		);
		this.latestRevision.set(sessionId, model.engineRevision);
		this.emit(sessionId, { type: 'job-created', job });
		this.emit(sessionId, { type: 'job-progress', jobId: job.id, progress: { phase: 'loading', done: 1, total: 1 } });
		await this.refreshSnapshots(sessionId, model.engineRevision);
		this.emit(sessionId, { type: 'job-succeeded', jobId: job.id, result: { ok: true } });
		return job;
	}

	async runCommand(intent: CommandIntent): Promise<VerificationJob> {
		const { job, result } = await this.post<{ job: VerificationJob; result?: unknown }>(
			`/api/engine/session/${encodeURIComponent(intent.sessionId)}/command`,
			{ intent }
		);
		this.emit(intent.sessionId, { type: 'job-created', job });
		this.emit(intent.sessionId, { type: 'job-progress', jobId: job.id, progress: { phase: 'running', done: 0, total: 1 } });
		if (intent.commandId.startsWith('check.') && result) {
			this.emit(intent.sessionId, {
				type: 'check-updated',
				result: normalizeCheckResult(result, {
					jobId: job.id,
					sessionId: intent.sessionId,
					mode: checkModeFromCommand(intent.commandId)
				})
			});
		}
		await this.refreshSnapshots(intent.sessionId, job.modelRevision);
		this.emit(intent.sessionId, { type: 'job-progress', jobId: job.id, progress: { phase: 'running', done: 1, total: 1 } });
		this.emit(intent.sessionId, { type: 'job-succeeded', jobId: job.id, result: { ok: true } });
		return job;
	}

	async cancelJob(jobId: Id): Promise<void> {
		await this.post(`/api/engine/jobs/${encodeURIComponent(jobId)}/cancel`, {});
	}

	async getSnapshot(sessionId: Id, request: SnapshotRequest): Promise<SnapshotBundle> {
		const params = new URLSearchParams();
		for (const graphId of request.graphIds ?? []) {
			params.append('graph', graphId);
		}
		for (const conceptId of request.conceptIds ?? []) {
			params.append('concept', conceptId);
		}
		const suffix = params.size ? `?${params.toString()}` : '';
		const payload = await this.get<{ arg?: unknown; concept?: unknown }>(
			`/api/engine/session/${encodeURIComponent(sessionId)}/snapshot${suffix}`
		);
		const revision = this.latestRevision.get(sessionId) ?? this.sessions.get(sessionId)?.currentModelRevision ?? 0;
		return {
			graphs: payload.arg
				? [
						normalizeGraphPayload(payload.arg as never, {
							id: stableId('graph', sessionId, 'arg', revision),
							sheetId: 'sheet-1',
							kind: 'arg',
							sourceRevision: revision
						})
					]
				: [],
			concepts: payload.concept
				? [normalizeConceptPayload(payload.concept, { sessionId, sheetId: 'sheet-1', revision })]
				: []
		};
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

	private async refreshSnapshots(sessionId: Id, revision: number) {
		const snapshot = await this.getSnapshot(sessionId, {});
		for (const graph of snapshot.graphs ?? []) {
			graph.sourceRevision = revision;
			this.emit(sessionId, { type: 'graph-updated', snapshot: graph });
		}
		for (const concept of snapshot.concepts ?? []) {
			concept.revision = revision;
			this.emit(sessionId, { type: 'concept-updated', concept });
			this.emit(sessionId, { type: 'toggles-updated', sheetId: concept.sheetId, toggles: concept.toggles });
		}
	}

	private emit(sessionId: Id, event: EngineEvent) {
		for (const listener of this.listeners.get(sessionId) ?? []) {
			listener(event);
		}
	}

	private async get<T>(path: string): Promise<T> {
		const response = await this.fetcher(this.url(path), {
			credentials: 'include',
			headers: { accept: 'application/json' }
		});
		return readJsonResponse<T>(response);
	}

	private async post<T = JsonResponse<unknown>>(path: string, body: unknown): Promise<T> {
		const headers: Record<string, string> = { accept: 'application/json', 'content-type': 'application/json' };
		const csrf = csrfTokenFromCookie();
		if (csrf) {
			headers['x-csrf-token'] = csrf;
		}
		const response = await this.fetcher(this.url(path), {
			method: 'POST',
			credentials: 'include',
			headers,
			body: JSON.stringify(body)
		});
		return readJsonResponse<T>(response);
	}

	private url(path: string) {
		return `${this.baseUrl}${path}`;
	}
}

function checkModeFromCommand(commandId: string): CheckMode {
	const raw = commandId.replace(/^check\./, '');
	if (raw === 'bounded' || raw === 'pdr' || raw === 'concrete' || raw === 'abstract') {
		return raw;
	}
	return 'induction';
}

function csrfTokenFromCookie(doc = globalThis.document) {
	const cookie = doc?.cookie ?? '';
	for (const part of cookie.split(';')) {
		const [name, ...value] = part.trim().split('=');
		if (name === 'ivysvk_csrf') {
			return decodeURIComponent(value.join('='));
		}
	}
	return '';
}

async function readJsonResponse<T>(response: Response): Promise<T> {
	if (!response.ok) {
		let message = `Hosted engine request failed with ${response.status}`;
		try {
			if ((response.headers.get('content-type') ?? '').includes('application/json')) {
				const payload = (await response.json()) as { error?: string };
				message = payload.error ?? message;
			} else {
				const text = (await response.text()).trim();
				if (text) {
					message = text;
				}
			}
		} catch {
			// Keep the status-based message when the body is not JSON.
		}
		throw new Error(message);
	}
	return (await response.json()) as T;
}
