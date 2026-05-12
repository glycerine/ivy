import type {
	CommandIntent,
	ConceptState,
	EngineEvent,
	EngineSession,
	GraphSnapshot,
	Id,
	IvyEngine,
	ModelDocument,
	SnapshotBundle,
	SnapshotRequest,
	Unsubscribe,
	VerificationJob,
	VerificationJobKind
} from '$lib/types';

export type FakeEngineOptions = {
	createId?: (prefix: string) => Id;
	now?: () => string;
	failCommands?: Set<string>;
};

export class FakeEngine implements IvyEngine {
	readonly kind = 'fake' as const;

	private readonly createId: (prefix: string) => Id;
	private readonly now: () => string;
	private readonly failCommands: Set<string>;
	private readonly listeners = new Map<Id, Set<(event: EngineEvent) => void>>();
	private readonly sessions = new Map<Id, EngineSession>();
	private sequence = 0;

	constructor(options: FakeEngineOptions = {}) {
		this.createId =
			options.createId ??
			((prefix) => {
				this.sequence += 1;
				return `${prefix}-${this.sequence}`;
			});
		this.now = options.now ?? (() => new Date().toISOString());
		this.failCommands = options.failCommands ?? new Set();
	}

	async newSession(projectId: Id): Promise<EngineSession> {
		const timestamp = this.now();
		const session: EngineSession = {
			id: this.createId('session'),
			kind: this.kind,
			status: 'ready',
			capabilities: {
				offline: true,
				persistentJobs: false,
				cancelJob: true,
				eventStream: true,
				parallelJobs: true
			},
			projectId,
			createdAt: timestamp,
			updatedAt: timestamp
		};
		this.sessions.set(session.id, session);
		return session;
	}

	async closeSession(sessionId: Id): Promise<void> {
		const session = this.sessions.get(sessionId);
		if (session) {
			session.status = 'closed';
			session.updatedAt = this.now();
			this.emit(sessionId, { type: 'session-ready', session });
		}
		this.listeners.delete(sessionId);
	}

	async loadModel(sessionId: Id, model: ModelDocument): Promise<VerificationJob> {
		const session = this.requireSession(sessionId);
		session.currentModelId = model.id;
		session.currentModelRevision = model.engineRevision;
		session.updatedAt = this.now();
		this.emit(sessionId, { type: 'session-ready', session });

		const job = this.createJob(session, model, 'load');
		this.emitJobLifecycle(sessionId, job, {
			graph: this.createGraphSnapshot('arg', 'sheet-arg', model.engineRevision),
			concept: this.createConceptState(sessionId, 'sheet-concept')
		});
		return job;
	}

	async runCommand(intent: CommandIntent): Promise<VerificationJob> {
		const session = this.requireSession(intent.sessionId);
		const modelId = session.currentModelId ?? 'model-unknown';
		const revision = session.currentModelRevision ?? 0;
		const kind = commandToJobKind(intent.commandId);
		const model: ModelDocument = {
			id: modelId,
			projectId: session.projectId,
			filename: 'fake.ivy',
			text: '',
			dirty: false,
			parseRevision: revision,
			engineRevision: revision,
			createdAt: this.now(),
			updatedAt: this.now()
		};
		const job = this.createJob(session, model, kind);

		if (this.failCommands.has(intent.commandId)) {
			this.emit(intent.sessionId, { type: 'job-created', job });
			this.emit(intent.sessionId, { type: 'job-failed', jobId: job.id, error: 'fake command failed' });
			return job;
		}

		this.emitJobLifecycle(intent.sessionId, job, {
			graph: this.createGraphSnapshot(intent.target?.kind ?? 'arg', intent.target?.graphId ?? 'sheet-arg', revision)
		});
		return job;
	}

	async cancelJob(jobId: Id): Promise<void> {
		for (const sessionId of this.listeners.keys()) {
			this.emit(sessionId, { type: 'job-failed', jobId, error: 'cancelled' });
		}
	}

	async getSnapshot(sessionId: Id, request: SnapshotRequest): Promise<SnapshotBundle> {
		void request;
		this.requireSession(sessionId);
		return {
			graphs: [this.createGraphSnapshot('arg', 'sheet-arg', 1)],
			concepts: [this.createConceptState(sessionId, 'sheet-concept')]
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

	private requireSession(sessionId: Id): EngineSession {
		const session = this.sessions.get(sessionId);
		if (!session) {
			throw new Error(`Unknown fake engine session: ${sessionId}`);
		}
		return session;
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

	private emitJobLifecycle(
		sessionId: Id,
		job: VerificationJob,
		snapshots: { graph?: GraphSnapshot; concept?: ConceptState }
	) {
		this.emit(sessionId, { type: 'job-created', job });
		this.emit(sessionId, { type: 'job-progress', jobId: job.id, progress: { phase: 'fake-check', done: 1, total: 2 } });
		if (snapshots.graph) {
			this.emit(sessionId, { type: 'graph-updated', snapshot: snapshots.graph });
		}
		if (snapshots.concept) {
			this.emit(sessionId, { type: 'concept-updated', concept: snapshots.concept });
		}
		this.emit(sessionId, { type: 'job-succeeded', jobId: job.id, result: { ok: true } });
	}

	private createGraphSnapshot(kind: GraphSnapshot['kind'], sheetId: Id, sourceRevision: number): GraphSnapshot {
		const firstNodeId = this.createId('node');
		const secondNodeId = this.createId('node');
		const edgeId = this.createId('edge');
		return {
			id: this.createId('graph'),
			sheetId,
			kind,
			sourceRevision,
			nodes: {
				[firstNodeId]: {
					id: firstNodeId,
					obj: 'fake.node',
					label: '0',
					classes: ['fake'],
					shape: 'ellipse',
					actions: [{ label: 'Expand', action: 'arg.expand', args: { depth: 1 } }]
				},
				[secondNodeId]: {
					id: secondNodeId,
					obj: 'fake.node.next',
					label: '1',
					classes: ['fake'],
					shape: 'ellipse',
					actions: [{ label: 'Expand', action: 'arg.expand', args: { depth: 1 } }]
				}
			},
			edges: {
				[edgeId]: {
					id: edgeId,
					obj: 'fake.edge',
					source: firstNodeId,
					target: secondNodeId,
					label: 'call ext',
					classes: ['fake']
				}
			},
			nodeOrder: [firstNodeId, secondNodeId],
			edgeOrder: [edgeId],
			layout: { [firstNodeId]: { x: 120, y: 110 }, [secondNodeId]: { x: 120, y: 286 } },
			styleRevision: 1,
			createdAt: this.now()
		};
	}

	private createConceptState(sessionId: Id, sheetId: Id): ConceptState {
		return {
			id: this.createId('concept-state'),
			sessionId,
			sheetId,
			concepts: {
				reachable: {
					name: 'reachable',
					variables: ['X'],
					formula: 'reachable(X)',
					sorts: ['node'],
					arity: 1
				}
			},
			sortNodes: ['node'],
			relationEdges: [],
			nodeLabels: ['reachable'],
			abstractValue: {},
			toggles: { edges: {}, labels: {} },
			revision: 1
		};
	}

	private emit(sessionId: Id, event: EngineEvent) {
		for (const listener of this.listeners.get(sessionId) ?? []) {
			listener(event);
		}
	}
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
		case 'concept.action':
			return 'concept-action';
		case 'proof.action':
			return 'proof-action';
		case 'event.action':
			return 'event-action';
		default:
			return 'arg-action';
	}
}
