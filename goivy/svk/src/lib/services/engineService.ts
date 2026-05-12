import type {
	CommandIntent,
	EngineEvent,
	EngineSession,
	Id,
	IvyEngine,
	ModelDocument,
	Unsubscribe,
	VerificationJob
} from '$lib/types';
import type { createConceptsState } from '$lib/state/concepts.svelte';
import type { createEnginesState } from '$lib/state/engines.svelte';
import type { createGraphsState } from '$lib/state/graphs.svelte';
import type { createJobsState } from '$lib/state/jobs.svelte';
import type { createWorkspaceState } from '$lib/state/workspace.svelte';

export type EngineServiceStores = {
	workspace: ReturnType<typeof createWorkspaceState>;
	engines: ReturnType<typeof createEnginesState>;
	jobs: ReturnType<typeof createJobsState>;
	graphs: ReturnType<typeof createGraphsState>;
	concepts: ReturnType<typeof createConceptsState>;
};

export type EngineServiceOptions = {
	engine: IvyEngine;
	stores: EngineServiceStores;
};

export class EngineService {
	private activeSessionId: Id | null = null;
	private unsubscribe: Unsubscribe | null = null;
	private diagnostics: EngineEvent[] = [];

	constructor(private readonly options: EngineServiceOptions) {}

	get diagnosticEvents() {
		return this.diagnostics;
	}

	async startSession(projectId: Id): Promise<EngineSession> {
		this.unsubscribe?.();
		const session = await this.options.engine.newSession(projectId);
		this.activeSessionId = session.id;
		this.unsubscribe = this.options.engine.subscribe(session.id, (event) => this.reduceEvent(event));
		this.options.stores.engines.upsert(session);
		this.options.stores.workspace.selectEngine(session.id);
		this.options.stores.workspace.selectSession(session.id);
		return session;
	}

	async closeSession(): Promise<void> {
		if (!this.activeSessionId) {
			return;
		}
		const sessionId = this.activeSessionId;
		this.unsubscribe?.();
		this.unsubscribe = null;
		this.activeSessionId = null;
		await this.options.engine.closeSession(sessionId);
		this.options.stores.engines.setStatus(sessionId, 'closed');
		this.options.stores.workspace.selectSession(null);
	}

	async loadModel(model: ModelDocument): Promise<VerificationJob> {
		const sessionId = this.requireActiveSession();
		const job = await this.options.engine.loadModel(sessionId, model);
		if (!this.options.stores.jobs.table.byId[job.id]) {
			this.options.stores.jobs.upsert(job);
		}
		return job;
	}

	async runCommand(intent: Omit<CommandIntent, 'id' | 'sessionId' | 'engineId'> & { id: Id }): Promise<VerificationJob> {
		const sessionId = this.requireActiveSession();
		const fullIntent: CommandIntent = {
			...intent,
			sessionId,
			engineId: sessionId
		};
		const job = await this.options.engine.runCommand(fullIntent);
		if (!this.options.stores.jobs.table.byId[job.id]) {
			this.options.stores.jobs.upsert(job);
		}
		return job;
	}

	reduceEvent(event: EngineEvent) {
		switch (event.type) {
			case 'session-ready':
				this.options.stores.engines.upsert(event.session);
				this.options.stores.workspace.selectSession(event.session.id);
				break;
			case 'job-created':
				this.options.stores.jobs.upsert(event.job);
				break;
			case 'job-progress':
				this.options.stores.jobs.setProgress(event.jobId, event.progress);
				break;
			case 'job-succeeded':
				this.options.stores.jobs.transition(event.jobId, 'succeeded');
				break;
			case 'job-failed':
				this.options.stores.jobs.fail(event.jobId, event.error);
				break;
			case 'graph-updated':
				this.options.stores.graphs.upsert(event.snapshot);
				break;
			case 'concept-updated':
				this.options.stores.concepts.upsert(event.concept);
				break;
			case 'diagnostic':
				this.diagnostics = [...this.diagnostics, event];
				break;
		}
	}

	private requireActiveSession(): Id {
		if (!this.activeSessionId) {
			throw new Error('No active engine session');
		}
		return this.activeSessionId;
	}
}

export function createEngineService(options: EngineServiceOptions): EngineService {
	return new EngineService(options);
}
