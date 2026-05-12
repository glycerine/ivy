import type { Id, Revision } from './ids';

export type SyncEntity =
	| 'model'
	| 'model_revision'
	| 'job'
	| 'job_result'
	| 'graph_snapshot'
	| 'concept_state'
	| 'trace_event_sheet';

export type SyncOp = {
	id: Id;
	projectId: Id;
	entity: SyncEntity;
	op: 'create' | 'update' | 'delete';
	baseRevision?: Revision;
	payload: unknown;
	status: 'pending' | 'syncing' | 'synced' | 'conflict' | 'failed';
	error?: string;
	createdAt: string;
	updatedAt: string;
};
