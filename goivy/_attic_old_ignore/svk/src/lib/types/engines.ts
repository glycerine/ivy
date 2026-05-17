import type { Id, Revision } from './ids';
import type { CommandIntent } from './commands';
import type { ConceptState, ConceptToggles } from './concepts';
import type { GraphSnapshot } from './graphs';
import type { CheckResult, VerificationJob } from './jobs';
import type { ModelDocument } from './models';

export type EngineKind = 'browser-js-wasm' | 'hosted-go';

export type EngineCapabilities = {
	offline: boolean;
	persistentJobs: boolean;
	cancelJob: boolean;
	eventStream: boolean;
	parallelJobs: boolean;
};

export type EngineSession = {
	id: Id;
	kind: EngineKind;
	status: 'idle' | 'starting' | 'ready' | 'busy' | 'failed' | 'closed';
	capabilities: EngineCapabilities;
	projectId: Id;
	currentModelId?: Id;
	currentModelRevision?: Revision;
	error?: string;
	createdAt: string;
	updatedAt: string;
};

export type DiagnosticSeverity = 'info' | 'warning' | 'error';

export type EngineEvent =
	| { type: 'session-ready'; session: EngineSession }
	| { type: 'job-created'; job: VerificationJob }
	| { type: 'job-progress'; jobId: Id; progress: VerificationJob['progress'] }
	| { type: 'job-succeeded'; jobId: Id; result: unknown }
	| { type: 'job-failed'; jobId: Id; error: string }
	| { type: 'check-updated'; result: CheckResult }
	| { type: 'graph-updated'; snapshot: GraphSnapshot }
	| { type: 'concept-updated'; concept: ConceptState }
	| { type: 'toggles-updated'; sheetId: Id; toggles: ConceptToggles }
	| { type: 'diagnostic'; severity: DiagnosticSeverity; message: string };

export type SnapshotRequest = {
	graphIds?: Id[];
	conceptIds?: Id[];
};

export type SnapshotBundle = {
	graphs?: GraphSnapshot[];
	concepts?: ConceptState[];
};

export type Unsubscribe = () => void;

export interface IvyEngine {
	readonly kind: EngineKind;
	newSession(projectId: Id): Promise<EngineSession>;
	closeSession(sessionId: Id): Promise<void>;
	loadModel(sessionId: Id, model: ModelDocument): Promise<VerificationJob>;
	runCommand(intent: CommandIntent): Promise<VerificationJob>;
	cancelJob(jobId: Id): Promise<void>;
	getSnapshot(sessionId: Id, request: SnapshotRequest): Promise<SnapshotBundle>;
	subscribe(sessionId: Id, onEvent: (event: EngineEvent) => void): Unsubscribe;
}
