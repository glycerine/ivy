import { describe, expect, it } from 'vitest';
import { FakeEngine } from './fakeEngine';
import { compareObservations, observeEngine } from './parityHarness';
import type { ModelDocument } from '$lib/types';

const model: ModelDocument = {
	id: 'model-1',
	projectId: 'project-1',
	filename: 'parity.ivy',
	text: 'type node',
	dirty: false,
	parseRevision: 1,
	engineRevision: 4,
	createdAt: '2026-05-12T00:00:00.000Z',
	updatedAt: '2026-05-12T00:00:00.000Z'
};

describe('engine parity harness', () => {
	it('observes normalized load and graph shapes', async () => {
		expect.hasAssertions();

		const observation = await observeEngine(new FakeEngine({ now: () => '2026-05-12T00:00:00.000Z' }), 'project-1', model);

		expect(observation.loadJob).toMatchObject({ kind: 'load', modelRevision: 4 });
		expect(observation.graphShape).toMatchObject({ kind: 'arg' });
	});

	it('compares observations and records known differences explicitly', () => {
		expect.hasAssertions();

		const differences = compareObservations(
			{
				engine: 'browser-js-wasm',
				loadJob: { kind: 'load', status: 'succeeded', modelRevision: 1 },
				graphShape: { kind: 'arg', nodeOrder: ['a'], edgeOrder: [] }
			},
			{
				engine: 'hosted-go',
				loadJob: { kind: 'load', status: 'succeeded', modelRevision: 1 },
				graphShape: { kind: 'arg', nodeOrder: ['a', 'b'], edgeOrder: [] }
			},
			{ 'graphShape.nodeCount': 'hosted engine currently includes root summary node' }
		);

		expect(differences).toEqual([
			{
				field: 'graphShape.nodeCount',
				left: 1,
				right: 2,
				note: 'hosted engine currently includes root summary node'
			}
		]);
	});
});
