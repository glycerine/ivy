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

export type HostedGoEngineOptions = {
	baseUrl?: string;
	fetcher?: typeof fetch;
	createEventSource?: (url: string) => EventSourceLike;
};

export type EventSourceLike = {
	onmessage: ((event: MessageEvent<string>) => void) | null;
	onerror: ((event: Event) => void) | null;
	close(): void;
};

type JsonResponse<T> = {
	ok: true;
	value: T;
};

export class HostedGoEngine implements IvyEngine {
	readonly kind: EngineKind = 'hosted-go';

	private readonly baseUrl: string;
	private readonly fetcher: typeof fetch;
	private readonly createEventSource: (url: string) => EventSourceLike;
	private readonly eventSources = new Map<Id, EventSourceLike>();
	private readonly listeners = new Map<Id, Set<(event: EngineEvent) => void>>();

	constructor(options: HostedGoEngineOptions = {}) {
		this.baseUrl = options.baseUrl ?? '';
		this.fetcher = options.fetcher ?? fetch;
		this.createEventSource = options.createEventSource ?? ((url) => new EventSource(url));
	}

	async newSession(projectId: Id): Promise<EngineSession> {
		const { session } = await this.post<{ session: EngineSession }>('/api/engine/session', { projectId });
		return session;
	}

	async closeSession(sessionId: Id): Promise<void> {
		this.eventSources.get(sessionId)?.close();
		this.eventSources.delete(sessionId);
		this.listeners.delete(sessionId);
	}

	async loadModel(sessionId: Id, model: ModelDocument): Promise<VerificationJob> {
		const { job } = await this.post<{ job: VerificationJob }>(
			`/api/engine/session/${encodeURIComponent(sessionId)}/load`,
			{ model }
		);
		return job;
	}

	async runCommand(intent: CommandIntent): Promise<VerificationJob> {
		const { job } = await this.post<{ job: VerificationJob }>(
			`/api/engine/session/${encodeURIComponent(intent.sessionId)}/command`,
			{ intent }
		);
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
		return this.get<SnapshotBundle>(`/api/engine/session/${encodeURIComponent(sessionId)}/snapshot${suffix}`);
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

	private openEventStream(sessionId: Id) {
		const source = this.createEventSource(this.url(`/api/engine/session/${encodeURIComponent(sessionId)}/events`));
		source.onmessage = (event) => {
			const engineEvent = JSON.parse(event.data) as EngineEvent;
			for (const listener of this.listeners.get(sessionId) ?? []) {
				listener(engineEvent);
			}
		};
		source.onerror = () => {
			for (const listener of this.listeners.get(sessionId) ?? []) {
				listener({ type: 'diagnostic', severity: 'warning', message: 'Hosted engine event stream disconnected' });
			}
		};
		this.eventSources.set(sessionId, source);
	}

	private async get<T>(path: string): Promise<T> {
		const response = await this.fetcher(this.url(path), {
			credentials: 'include',
			headers: { accept: 'application/json' }
		});
		return readJsonResponse<T>(response);
	}

	private async post<T = JsonResponse<unknown>>(path: string, body: unknown): Promise<T> {
		const response = await this.fetcher(this.url(path), {
			method: 'POST',
			credentials: 'include',
			headers: { accept: 'application/json', 'content-type': 'application/json' },
			body: JSON.stringify(body)
		});
		return readJsonResponse<T>(response);
	}

	private url(path: string) {
		return `${this.baseUrl}${path}`;
	}
}

async function readJsonResponse<T>(response: Response): Promise<T> {
	if (!response.ok) {
		let message = `Hosted engine request failed with ${response.status}`;
		try {
			const payload = (await response.json()) as { error?: string };
			message = payload.error ?? message;
		} catch {
			// Keep the status-based message when the body is not JSON.
		}
		throw new Error(message);
	}
	return (await response.json()) as T;
}
