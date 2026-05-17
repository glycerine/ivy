import { describe, expect, it } from 'vitest';
import { normalizeGraphPayload } from './graphs';

describe('normalizeGraphPayload', () => {
	it('normalizes an empty graph snapshot', () => {
		expect.hasAssertions();

		const graph = normalizeGraphPayload(null, {
			id: 'graph-1',
			sheetId: 'sheet-1',
			kind: 'arg',
			now: new Date('2026-05-12T00:00:00.000Z')
		});

		expect(graph).toMatchObject({
			id: 'graph-1',
			sheetId: 'sheet-1',
			kind: 'arg',
			nodes: {},
			edges: {},
			nodeOrder: [],
			edgeOrder: [],
			createdAt: '2026-05-12T00:00:00.000Z'
		});
	});

	it('maps Cytoscape nodes and edges to stable normalized records', () => {
		expect.hasAssertions();

		const graph = normalizeGraphPayload(
			{
				elements: [
					{
						group: 'nodes',
						classes: 'state selected',
						data: {
							id: 'n0',
							obj: 'state_0',
							label: '0',
							short_info: 'State 0',
							long_info: 'state details',
							shape: 'ellipse',
							actions: [{ label: 'Execute', action: 'arg.execute', args: { action: 'ext:connect' } }]
						},
						position: { x: 10, y: 20 }
					},
					{
						group: 'nodes',
						classes: 'bottom_state',
						data: { id: 'n1', obj: 'state_1', label: '1' }
					},
					{
						group: 'edges',
						classes: 'transition_action',
						data: {
							id: 'e0',
							obj: 'tr_0_1',
							source: 'n0',
							target: 'n1',
							source_obj: 'state_0',
							target_obj: 'state_1',
							label: 'connect',
							short_info: 'connect'
						}
					}
				]
			},
			{
				id: 'graph-1',
				sheetId: 'sheet-1',
				kind: 'arg',
				sourceRevision: 7,
				now: new Date('2026-05-12T00:00:00.000Z')
			}
		);

		expect(graph.nodeOrder).toEqual(['node_state_0', 'node_state_1']);
		expect(graph.edgeOrder).toEqual(['edge_tr_0_1_state_0_state_1']);
		expect(graph.nodes.node_state_0).toMatchObject({
			obj: 'state_0',
			label: '0',
			classes: ['state', 'selected'],
			shape: 'ellipse',
			shortInfo: 'State 0',
			longInfo: 'state details'
		});
		expect(graph.nodes.node_state_0.actions).toEqual([
			{ label: 'Execute', action: 'arg.execute', args: { action: 'ext:connect' } }
		]);
		expect(graph.edges.edge_tr_0_1_state_0_state_1).toMatchObject({
			obj: 'tr_0_1',
			source: 'node_state_0',
			target: 'node_state_1',
			label: 'connect',
			classes: ['transition_action']
		});
		expect(graph.layout).toEqual({ node_state_0: { x: 10, y: 20 } });
	});
});
