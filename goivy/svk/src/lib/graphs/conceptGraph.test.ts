import { describe, expect, it } from 'vitest';
import { conceptStateToGraphSnapshot } from './conceptGraph';
import type { ConceptState } from '$lib/types';

describe('conceptStateToGraphSnapshot', () => {
	it('turns concept state into an octagonal concept graph snapshot', () => {
		expect.hasAssertions();
		const concept: ConceptState = {
			id: 'concept-state-1',
			sessionId: 'session-1',
			sheetId: 'sheet-1',
			concepts: {
				reachable: { name: 'reachable', variables: ['X'], formula: 'reachable(X)', sorts: ['node'], arity: 1 },
				blocked: { name: 'blocked', variables: ['X'], formula: 'blocked(X)', sorts: ['node'], arity: 1 }
			},
			sortNodes: ['node'],
			relationEdges: [],
			nodeLabels: [],
			abstractValue: {},
			toggles: { edges: {}, labels: {} },
			revision: 3
		};

		const graph = conceptStateToGraphSnapshot(concept);

		expect(graph).toMatchObject({ kind: 'concept', sheetId: 'sheet-1', sourceRevision: 3 });
		expect(graph.nodeOrder).toEqual(['concept:blocked', 'concept:reachable']);
		expect(graph.nodes['concept:reachable']).toMatchObject({
			obj: 'reachable',
			label: 'reachable',
			shape: 'octagon',
			shortInfo: 'reachable(X)'
		});
	});
});
