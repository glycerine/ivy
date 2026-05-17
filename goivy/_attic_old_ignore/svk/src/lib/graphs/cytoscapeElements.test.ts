import { describe, expect, it } from 'vitest';
import { graphSnapshotToCytoscapeElements } from './cytoscapeElements';
import type { GraphSnapshot } from '$lib/types';

const snapshot: GraphSnapshot = {
	id: 'graph-1',
	sheetId: 'sheet-1',
	kind: 'arg',
	sourceRevision: 1,
	nodes: {
		'n-1': {
			id: 'n-1',
			obj: 'node.one',
			label: 'One',
			classes: ['reachable', 'selected'],
			actions: [{ label: 'Expand', action: 'arg.expand', args: { depth: 1 } }]
		},
		'n-2': {
			id: 'n-2',
			obj: 'node.two',
			label: 'Two',
			classes: ['maybe']
		}
	},
	edges: {
		'e-1': {
			id: 'e-1',
			obj: 'edge.one',
			source: 'n-1',
			target: 'n-2',
			label: 'link',
			classes: ['transition']
		}
	},
	nodeOrder: ['n-1', 'n-2'],
	edgeOrder: ['e-1'],
	layout: {
		'n-1': { x: 10, y: 20 },
		'n-2': { x: 80, y: 20 }
	},
	styleRevision: 1,
	createdAt: '2026-05-12T00:00:00.000Z'
};

describe('graphSnapshotToCytoscapeElements', () => {
	it('converts nodes and edges in snapshot order', () => {
		expect.hasAssertions();

		const elements = graphSnapshotToCytoscapeElements(snapshot);

		expect(elements.map((element) => element.data.id)).toEqual(['n-1', 'n-2', 'e-1']);
		expect(elements[0]).toMatchObject({ group: 'nodes', data: { label: 'One' } });
		expect(elements[2]).toMatchObject({ group: 'edges', data: { source: 'n-1', target: 'n-2' } });
	});

	it('preserves classes, action descriptors, and layout positions', () => {
		expect.hasAssertions();

		const [node] = graphSnapshotToCytoscapeElements(snapshot);

		expect(node.classes).toBe('reachable selected');
		expect(node.data.actions).toEqual([{ label: 'Expand', action: 'arg.expand', args: { depth: 1 } }]);
		expect(node.position).toEqual({ x: 10, y: 20 });
	});
});
