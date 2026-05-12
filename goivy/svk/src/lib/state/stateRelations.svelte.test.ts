import { describe, expect, it } from 'vitest';
import { createStateRelationsState, rowsFromConcept, rowsFromToggles } from './stateRelations.svelte';
import type { ConceptState, ConceptToggles } from '$lib/types';

const toggles: ConceptToggles = {
	edges: {
		'link(X,Y)': { all_to_all: true, edge_unknown: false, none_to_none: false, transitive: true }
	},
	labels: {
		'=@X': { node_necessarily: true, node_maybe: false, node_necessarily_not: false }
	}
};

const concept: ConceptState = {
	id: 'concept-1',
	sessionId: 'session-1',
	sheetId: 'sheet-1',
	concepts: {},
	sortNodes: ['client'],
	relations: ['semaphore'],
	relationEdges: ['link(X,Y)'],
	nodeLabels: ['=@X'],
	abstractValue: {},
	toggles,
	revision: 1
};

describe('state relation rows', () => {
	it('derives stable rows from concept payloads', () => {
		expect.hasAssertions();

		const rows = rowsFromConcept(concept);

		expect(rows.map((row) => `${row.kind}:${row.name}`)).toEqual([
			'edge:link(X,Y)',
			'edge:semaphore',
			'label:=@X'
		]);
		expect(rows[0].columns.map((column) => [column.label, column.checked])).toEqual([
			['+', true],
			['?', false],
			['-', false],
			['T', true]
		]);
	});

	it('can reduce concept and toggle updates into state', () => {
		expect.hasAssertions();

		const state = createStateRelationsState();
		state.updateFromConcept(concept);
		state.updateFromToggles('sheet-2', {
			edges: { edge: { edge_unknown: true } },
			labels: {}
		});

		expect(state.current.revision).toBe(2);
		expect(state.rowsForSheet('sheet-2')).toHaveLength(1);
		expect(state.current.rows[0]).toMatchObject({ sheetId: 'sheet-2', name: 'edge' });
	});

	it('derives rows directly from toggles', () => {
		expect.hasAssertions();

		const rows = rowsFromToggles('sheet-1', toggles);

		expect(rows.map((row) => row.id)).toEqual(['sheet-1:edge:link(X,Y)', 'sheet-1:label:=@X']);
	});

	it('toggles an existing relation column in place', () => {
		expect.hasAssertions();
		const state = createStateRelationsState([
			{
				id: 'sheet-1:edge:link',
				sheetId: 'sheet-1',
				kind: 'edge',
				name: 'link',
				columns: [{ key: 'all_to_all', label: '+', checked: false }]
			}
		]);

		expect(state.toggle('sheet-1:edge:link', 'all_to_all', true)).toBe(true);
		expect(state.current.rows[0]?.columns[0]?.checked).toBe(true);
		expect(state.toggle('missing', 'all_to_all', true)).toBe(false);
	});
});
