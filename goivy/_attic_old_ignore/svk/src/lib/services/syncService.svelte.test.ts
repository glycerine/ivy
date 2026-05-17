import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { HostedProjectClient } from '$lib/api/projectClient';
import { createSvkRepository, deleteSvkDatabase, SyncQueue } from '$lib/persistence';
import { SyncService } from './syncService';

const databaseName = `svk-sync-service-test-${crypto.randomUUID()}`;
let repository: Awaited<ReturnType<typeof createSvkRepository>>;

describe('SyncService', () => {
	beforeEach(async () => {
		repository = await createSvkRepository({ name: databaseName });
	});

	afterEach(async () => {
		repository.close();
		await deleteSvkDatabase(databaseName);
	});

	it('uploads pending model revisions and marks them synced', async () => {
		expect.hasAssertions();

		const queue = new SyncQueue(repository, { createId: () => 'sync-1', now: () => '2026-05-12T00:00:00.000Z' });
		await queue.enqueue({
			projectId: 'project-1',
			entity: 'model_revision',
			op: 'update',
			baseRevision: 1,
			payload: { filename: 'demo.ivy', text: 'type t' }
		});
		const client = new HostedProjectClient(async () => new Response(JSON.stringify({ model: { id: 'model-1' } }), { status: 200 }));
		const service = new SyncService(queue, client);

		const processed = await service.uploadPendingModelRevisions('project-1');

		expect(processed).toMatchObject([{ id: 'sync-1', status: 'synced' }]);
	});

	it('turns hosted conflicts into conflict sync ops', async () => {
		expect.hasAssertions();

		const queue = new SyncQueue(repository, { createId: () => 'sync-1', now: () => '2026-05-12T00:00:00.000Z' });
		await queue.enqueue({
			projectId: 'project-1',
			entity: 'model_revision',
			op: 'update',
			baseRevision: 1,
			payload: { filename: 'demo.ivy', text: 'type t' }
		});
		const client = new HostedProjectClient(async () => new Response('conflict', { status: 409 }));
		const service = new SyncService(queue, client);

		const processed = await service.uploadPendingModelRevisions('project-1');

		expect(processed).toMatchObject([{ id: 'sync-1', status: 'conflict', error: 'model revision conflict' }]);
	});
});
