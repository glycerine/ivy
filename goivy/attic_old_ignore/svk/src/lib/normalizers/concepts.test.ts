import { describe, expect, it } from 'vitest';
import { normalizeConceptPayload } from './concepts';

describe('normalizeConceptPayload', () => {
	it('normalizes concept domains and display toggles', () => {
		expect.hasAssertions();

		const concept = normalizeConceptPayload(
			{
				concepts: {
					client: {
						name: 'client',
						variables: ['X'],
						formula: 'X = X',
						sorts: ['client'],
						arity: 1
					},
					link: {
						variables: ['X', 'Y'],
						formula: 'link(X, Y)',
						sorts: ['client', 'server'],
						arity: 2
					}
				},
				nodes: ['client', 'server'],
				edges: ['link'],
				node_labels: ['ready'],
				abstract_value: {
					'node_label|node_necessarily|client|ready': true
				},
				toggles: {
					edges: {
						link: { all_to_all: true, edge_unknown: false }
					},
					labels: {
						ready: { node_necessarily: true }
					}
				}
			},
			{ sessionId: 's1', sheetId: 'sheet-1', revision: 3 }
		);

		expect(concept.id).toBe('concept_s1_sheet-1_3');
		expect(concept.sortNodes).toEqual(['client', 'server']);
		expect(concept.relationEdges).toEqual(['link']);
		expect(concept.nodeLabels).toEqual(['ready']);
		expect(concept.concepts.link).toMatchObject({
			name: 'link',
			variables: ['X', 'Y'],
			sorts: ['client', 'server'],
			arity: 2
		});
		expect(concept.abstractValue['node_label|node_necessarily|client|ready']).toBe(true);
		expect(concept.toggles.edges.link.all_to_all).toBe(true);
		expect(concept.toggles.labels.ready.node_necessarily).toBe(true);
	});

	it('uses empty collections for missing optional payload fields', () => {
		expect.hasAssertions();

		const concept = normalizeConceptPayload({}, { sessionId: 's1', sheetId: 'sheet-1' });

		expect(concept.concepts).toEqual({});
		expect(concept.sortNodes).toEqual([]);
		expect(concept.relationEdges).toEqual([]);
		expect(concept.nodeLabels).toEqual([]);
		expect(concept.abstractValue).toEqual({});
		expect(concept.toggles).toEqual({ edges: {}, labels: {} });
	});
});
