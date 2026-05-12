import { describe, expect, it } from 'vitest';
import { createGraphsState } from './graphs.svelte';
import { createSelectionState } from './selection.svelte';
import type { GraphSnapshot } from '$lib/types';

const graph: GraphSnapshot = {
	id: 'graph-1',
	sheetId: 'sheet-1',
	kind: 'arg',
	sourceRevision: 1,
	styleRevision: 0,
	createdAt: '2026-05-12T00:00:00.000Z',
	nodeOrder: ['node-1'],
	edgeOrder: ['edge-1'],
	nodes: {
		'node-1': {
			id: 'node-1',
			obj: 'state_0',
			label: '0',
			classes: ['state'],
			shortInfo: 'State 0',
			longInfo: 'Long state text'
		}
	},
	edges: {
		'edge-1': {
			id: 'edge-1',
			obj: 'tr_0_1',
			source: 'node-1',
			target: 'node-1',
			label: 'self',
			classes: ['transition_action'],
			shortInfo: 'self edge'
		}
	}
};

describe('graph and selection state', () => {
	it('stores graphs by sheet', () => {
		expect.hasAssertions();

		const graphs = createGraphsState();
		graphs.upsert(graph);

		expect(graphs.graphsForSheet('sheet-1')).toEqual([graph]);
		expect(graphs.graphsForSheet('missing')).toEqual([]);
	});

	it('derives active details from selected graph nodes', () => {
		expect.hasAssertions();

		const selection = createSelectionState();
		selection.selectGraphNode(graph, 'node-1');

		expect(selection.current.graphSelections['graph-1']).toEqual({ nodeIds: ['node-1'], edgeIds: [] });
		expect(selection.current.activeDetails).toEqual({
			source: 'arg',
			id: 'node-1',
			shortInfo: 'State 0',
			longInfo: 'Long state text'
		});
	});
});
