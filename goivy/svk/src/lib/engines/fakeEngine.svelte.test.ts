import { describe, expect, it } from 'vitest';
import { FakeEngine } from './fakeEngine';
import { createConceptsState, createEnginesState, createGraphsState, createJobsState, createWorkspaceState } from '$lib/state';
import { createEngineService } from '$lib/services';
import type { ModelDocument } from '$lib/types';

const createdAt = '2026-05-12T00:00:00.000Z';

const model: ModelDocument = {
	id: 'model-1',
	projectId: 'project-1',
	filename: 'model.ivy',
	text: 'type t',
	dirty: false,
	parseRevision: 1,
	engineRevision: 7,
	createdAt,
	updatedAt: createdAt
};

function makeEngine() {
	let sequence = 0;
	return new FakeEngine({
		createId(prefix) {
			sequence += 1;
			return `${prefix}-${sequence}`;
		},
		now: () => createdAt
	});
}

function makeStores() {
	return {
		workspace: createWorkspaceState(),
		engines: createEnginesState(),
		jobs: createJobsState(),
		graphs: createGraphsState(),
		concepts: createConceptsState()
	};
}

describe('FakeEngine', () => {
	it('creates a ready offline-capable session', async () => {
		expect.hasAssertions();

		const engine = makeEngine();
		const session = await engine.newSession('project-1');

		expect(session).toMatchObject({
			id: 'session-1',
			kind: 'fake',
			status: 'ready',
			projectId: 'project-1',
			capabilities: { offline: true, eventStream: true }
		});
	});

	it('emits a deterministic job lifecycle through EngineService', async () => {
		expect.hasAssertions();

		const stores = makeStores();
		const service = createEngineService({ engine: makeEngine(), stores });
		const session = await service.startSession('project-1');
		const job = await service.loadModel(model);

		expect(session.id).toBe('session-1');
		expect(stores.workspace.current).toMatchObject({
			activeEngineId: 'session-1',
			activeSessionId: 'session-1'
		});
		expect(stores.jobs.table.byId[job.id]).toMatchObject({
			kind: 'load',
			status: 'succeeded',
			progress: { phase: 'fake-check', done: 1, total: 2 }
		});
		expect(stores.graphs.table.order).toHaveLength(1);
		expect(stores.concepts.table.order).toHaveLength(1);
	});

	it('reduces command graph updates into graph state', async () => {
		expect.hasAssertions();

		const stores = makeStores();
		const service = createEngineService({ engine: makeEngine(), stores });
		await service.startSession('project-1');
		await service.loadModel(model);
		const job = await service.runCommand({
			id: 'intent-1',
			commandId: 'check.induction',
			target: { kind: 'proof', graphId: 'proof-sheet' }
		});

		expect(stores.jobs.table.byId[job.id]).toMatchObject({
			kind: 'check-induction',
			status: 'succeeded'
		});
		const latestGraph = stores.graphs.table.byId[stores.graphs.table.order.at(-1) ?? ''];
		expect(latestGraph).toMatchObject({ kind: 'proof', sheetId: 'proof-sheet' });
	});

	it('stores failed command errors on the job record', async () => {
		expect.hasAssertions();

		const stores = makeStores();
		const engine = new FakeEngine({
			createId: (prefix) => `${prefix}-fixed`,
			now: () => createdAt,
			failCommands: new Set(['check.induction'])
		});
		const service = createEngineService({ engine, stores });
		await service.startSession('project-1');
		const job = await service.runCommand({ id: 'intent-1', commandId: 'check.induction' });

		expect(stores.jobs.table.byId[job.id]).toMatchObject({
			status: 'failed',
			error: 'fake command failed'
		});
	});
});
