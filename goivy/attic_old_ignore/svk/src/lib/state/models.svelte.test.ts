import { describe, expect, it } from 'vitest';
import { createModelsState } from './models.svelte';
import type { ModelDocument } from '$lib/types';

const baseModel: ModelDocument = {
	id: 'model-1',
	projectId: 'project-1',
	filename: 'client_server.ivy',
	text: 'before',
	savedTextHash: 'before',
	dirty: false,
	parseRevision: 0,
	engineRevision: 0,
	createdAt: '2026-05-12T00:00:00.000Z',
	updatedAt: '2026-05-12T00:00:00.000Z'
};

describe('createModelsState', () => {
	it('upserts models and preserves first-seen order', () => {
		expect.hasAssertions();

		const models = createModelsState();
		models.upsert(baseModel);
		models.upsert({ ...baseModel, id: 'model-2', filename: 'other.ivy' });
		models.upsert({ ...baseModel, filename: 'renamed.ivy' });

		expect(models.table.order).toEqual(['model-1', 'model-2']);
		expect(models.table.byId['model-1'].filename).toBe('renamed.ivy');
		expect(models.table.revision).toBe(3);
	});

	it('marks text updates dirty and saved revisions clean', () => {
		expect.hasAssertions();

		const models = createModelsState();
		models.upsert(baseModel);
		models.updateText('model-1', 'after', '2026-05-12T00:00:01.000Z');

		expect(models.table.byId['model-1']).toMatchObject({
			text: 'after',
			dirty: true,
			parseRevision: 1,
			updatedAt: '2026-05-12T00:00:01.000Z'
		});

		models.markSaved('model-1', 'after', '2026-05-12T00:00:02.000Z');

		expect(models.table.byId['model-1']).toMatchObject({
			savedTextHash: 'after',
			dirty: false,
			updatedAt: '2026-05-12T00:00:02.000Z'
		});
	});
});
