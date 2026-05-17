import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { createSvkRepository } from './repositories';
import { deleteSvkDatabase } from './indexedDb';
import { SyncQueue } from './syncQueue';
import type {
	CheckResult,
	ConceptState,
	GraphSnapshot,
	ModelDocument,
	ModelRevision,
	Project,
	TraceEventSheet,
	VerificationJob
} from '$lib/types';

const databaseName = `svk-persistence-test-${crypto.randomUUID()}`;
let repository: Awaited<ReturnType<typeof createSvkRepository>>;

const createdAt = '2026-05-12T00:00:00.000Z';

const project: Project = {
	id: 'project-1',
	accountId: 'account-1',
	ownerKind: 'user',
	ownerId: 'user-1',
	name: 'Offline proof lab',
	slug: 'offline-proof-lab',
	storageMode: 'local_indexeddb',
	createdAt,
	updatedAt: createdAt
};

const model: ModelDocument = {
	id: 'model-1',
	projectId: 'project-1',
	filename: 'demo.ivy',
	text: 'relation r(X,Y)',
	dirty: false,
	parseRevision: 1,
	engineRevision: 1,
	createdAt,
	updatedAt: createdAt
};

const revision: ModelRevision = {
	id: 'revision-1',
	modelId: 'model-1',
	projectId: 'project-1',
	revision: 1,
	textHash: 'hash-1',
	text: model.text,
	createdAt,
	createdBy: 'user-1'
};

const job: VerificationJob = {
	id: 'job-1',
	sessionId: 'session-1',
	engineId: 'engine-1',
	projectId: 'project-1',
	modelId: 'model-1',
	modelRevision: 1,
	kind: 'check-induction',
	status: 'queued',
	createdAt,
	updatedAt: createdAt
};

const checkResult: CheckResult = {
	id: 'result-1',
	jobId: 'job-1',
	sessionId: 'session-1',
	mode: 'induction',
	z3Contacted: true,
	result: 'pass',
	message: 'ok',
	createdAt
};

const graphSnapshot: GraphSnapshot = {
	id: 'graph-1',
	sheetId: 'sheet-1',
	kind: 'arg',
	sourceRevision: 1,
	nodes: {},
	edges: {},
	nodeOrder: [],
	edgeOrder: [],
	styleRevision: 1,
	createdAt
};

const conceptState: ConceptState = {
	id: 'concept-state-1',
	sessionId: 'session-1',
	sheetId: 'sheet-1',
	concepts: {},
	sortNodes: [],
	relationEdges: [],
	nodeLabels: [],
	abstractValue: {},
	toggles: { edges: {}, labels: {} },
	revision: 1
};

const traceEventSheet: TraceEventSheet = {
	id: 'trace-sheet-1',
	projectId: 'project-1',
	sessionId: 'session-1',
	label: 'Trace',
	events: [{ id: 'event-1', text: 'start', address: '0' }],
	patterns: [],
	expandedAddresses: ['0'],
	revision: 1
};

describe('SvkRepository IndexedDB persistence', () => {
	beforeEach(async () => {
		repository = await createSvkRepository({ name: databaseName });
	});

	afterEach(async () => {
		repository.close();
		await deleteSvkDatabase(databaseName);
	});

	it('persists and indexes the core offline project records', async () => {
		expect.hasAssertions();

		await repository.saveProject(project);
		await repository.saveModel(model);
		await repository.saveModelRevision(revision);
		await repository.saveJob(job);
		await repository.saveCheckResult(checkResult);
		await repository.saveGraphSnapshot(graphSnapshot);
		await repository.saveConceptState(conceptState);
		await repository.saveTraceEventSheet(traceEventSheet);

		await expect(repository.getProject('project-1')).resolves.toMatchObject({ slug: 'offline-proof-lab' });
		await expect(repository.listModels('project-1')).resolves.toHaveLength(1);
		await expect(repository.listModelRevisions('project-1')).resolves.toHaveLength(1);
		await expect(repository.listJobs('project-1')).resolves.toMatchObject([{ kind: 'check-induction' }]);
		await expect(repository.getCheckResult('result-1')).resolves.toMatchObject({ result: 'pass' });
		await expect(repository.getGraphSnapshot('graph-1')).resolves.toMatchObject({ kind: 'arg' });
		await expect(repository.getConceptState('concept-state-1')).resolves.toMatchObject({ revision: 1 });
		await expect(repository.getTraceEventSheet('trace-sheet-1')).resolves.toMatchObject({
			events: [{ id: 'event-1', text: 'start', address: '0' }]
		});
	});

	it('queues sync operations with deterministic ids and status transitions', async () => {
		expect.hasAssertions();

		const queue = new SyncQueue(repository, {
			createId: () => 'sync-op-1',
			now: () => '2026-05-12T00:00:01.000Z'
		});

		const op = await queue.enqueue({
			projectId: 'project-1',
			entity: 'model',
			op: 'update',
			baseRevision: 1,
			payload: { id: 'model-1', text: 'updated' }
		});

		expect(op).toMatchObject({
			id: 'sync-op-1',
			status: 'pending',
			baseRevision: 1
		});
		await expect(queue.listPending('project-1')).resolves.toHaveLength(1);
		await expect(queue.markSyncing('sync-op-1')).resolves.toMatchObject({ status: 'syncing' });
		await expect(queue.listPending('project-1')).resolves.toHaveLength(0);
		await expect(queue.markFailed('sync-op-1', 'conflict')).resolves.toMatchObject({
			status: 'failed',
			error: 'conflict'
		});
	});
});
