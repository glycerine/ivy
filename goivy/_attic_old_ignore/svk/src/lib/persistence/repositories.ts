import type {
	CheckResult,
	ConceptState,
	GraphSnapshot,
	ModelDocument,
	ModelRevision,
	Project,
	SyncOp,
	TraceEventSheet,
	VerificationJob
} from '$lib/types';
import { createSvkIndexedDb, type SvkDatabaseOptions, type SvkIndexedDb } from './indexedDb';

export class SvkRepository {
	constructor(private readonly db: SvkIndexedDb) {}

	close() {
		this.db.close();
	}

	saveProject(project: Project) {
		return this.db.put('projects', project);
	}

	getProject(id: string) {
		return this.db.get<Project>('projects', id);
	}

	listProjects() {
		return this.db.getAll<Project>('projects');
	}

	saveModel(model: ModelDocument) {
		return this.db.put('models', model);
	}

	getModel(id: string) {
		return this.db.get<ModelDocument>('models', id);
	}

	listModels(projectId: string) {
		return this.db.getByIndex<ModelDocument>('models', 'by_project', projectId);
	}

	saveModelRevision(revision: ModelRevision) {
		return this.db.put('model_revisions', revision);
	}

	listModelRevisions(projectId: string) {
		return this.db.getByIndex<ModelRevision>('model_revisions', 'by_project', projectId);
	}

	saveJob(job: VerificationJob) {
		return this.db.put('jobs', job);
	}

	getJob(id: string) {
		return this.db.get<VerificationJob>('jobs', id);
	}

	listJobs(projectId: string) {
		return this.db.getByIndex<VerificationJob>('jobs', 'by_project', projectId);
	}

	saveCheckResult(result: CheckResult) {
		return this.db.put('check_results', result);
	}

	getCheckResult(id: string) {
		return this.db.get<CheckResult>('check_results', id);
	}

	saveGraphSnapshot(snapshot: GraphSnapshot) {
		return this.db.put('graph_snapshots', snapshot);
	}

	getGraphSnapshot(id: string) {
		return this.db.get<GraphSnapshot>('graph_snapshots', id);
	}

	saveConceptState(state: ConceptState) {
		return this.db.put('concept_states', state);
	}

	getConceptState(id: string) {
		return this.db.get<ConceptState>('concept_states', id);
	}

	saveTraceEventSheet(sheet: TraceEventSheet) {
		return this.db.put('trace_event_sheets', sheet);
	}

	getTraceEventSheet(id: string) {
		return this.db.get<TraceEventSheet>('trace_event_sheets', id);
	}

	saveSyncOp(op: SyncOp) {
		return this.db.put('sync_ops', op);
	}

	getSyncOp(id: string) {
		return this.db.get<SyncOp>('sync_ops', id);
	}

	listSyncOps(status?: SyncOp['status']) {
		if (!status) {
			return this.db.getAll<SyncOp>('sync_ops');
		}
		return this.db.getByIndex<SyncOp>('sync_ops', 'by_status', status);
	}

	listProjectSyncOps(projectId: string, status?: SyncOp['status']) {
		if (!status) {
			return this.db.getByIndex<SyncOp>('sync_ops', 'by_project', projectId);
		}
		return this.db.getByIndex<SyncOp>('sync_ops', 'by_project_status', [projectId, status]);
	}

	deleteSyncOp(id: string) {
		return this.db.delete('sync_ops', id);
	}
}

export async function createSvkRepository(options: SvkDatabaseOptions): Promise<SvkRepository> {
	return new SvkRepository(await createSvkIndexedDb(options));
}
