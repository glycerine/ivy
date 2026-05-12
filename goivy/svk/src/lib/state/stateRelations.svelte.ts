import type {
	ConceptState,
	ConceptToggles,
	EdgeDisplayClass,
	Id,
	NodeLabelDisplayClass
} from '$lib/types';

export type StateRelationKind = 'edge' | 'label';

export type StateRelationColumn = {
	key: EdgeDisplayClass | NodeLabelDisplayClass;
	label: string;
	checked: boolean;
};

export type StateRelationRow = {
	id: string;
	sheetId: Id;
	kind: StateRelationKind;
	name: string;
	columns: StateRelationColumn[];
};

const edgeColumns: Array<{ key: EdgeDisplayClass; label: string }> = [
	{ key: 'all_to_all', label: '+' },
	{ key: 'edge_unknown', label: '?' },
	{ key: 'none_to_none', label: '-' },
	{ key: 'transitive', label: 'T' }
];

const labelColumns: Array<{ key: NodeLabelDisplayClass; label: string }> = [
	{ key: 'node_necessarily', label: '+' },
	{ key: 'node_maybe', label: '?' },
	{ key: 'node_necessarily_not', label: '-' }
];

export function createStateRelationsState(initialRows: StateRelationRow[] = []) {
	const current = $state({ rows: structuredClone(initialRows), revision: 0 });

	function bump() {
		current.revision += 1;
	}

	return {
		get current() {
			return current;
		},
		replaceRows(rows: StateRelationRow[]) {
			current.rows = rows;
			bump();
		},
		updateFromConcept(concept: ConceptState) {
			current.rows = rowsFromConcept(concept);
			bump();
		},
		updateFromToggles(sheetId: Id, toggles: ConceptToggles) {
			current.rows = rowsFromToggles(sheetId, toggles);
			bump();
		},
		rowsForSheet(sheetId: Id) {
			return current.rows.filter((row) => row.sheetId === sheetId);
		},
		toggle(rowId: string, key: EdgeDisplayClass | NodeLabelDisplayClass, checked: boolean) {
			const row = current.rows.find((candidate) => candidate.id === rowId);
			const column = row?.columns.find((candidate) => candidate.key === key);
			if (!column) {
				return false;
			}
			column.checked = checked;
			bump();
			return true;
		}
	};
}

export function rowsFromConcept(concept: ConceptState): StateRelationRow[] {
	const edgeNames = sortedUnique([
		...(concept.relations ?? []),
		...concept.relationEdges,
		...Object.keys(concept.toggles.edges)
	]);
	const labelNames = sortedUnique([...concept.nodeLabels, ...Object.keys(concept.toggles.labels)]);
	return [
		...edgeNames.map((name) => edgeRow(concept.sheetId, name, concept.toggles.edges[name] ?? {})),
		...labelNames.map((name) => labelRow(concept.sheetId, name, concept.toggles.labels[name] ?? {}))
	];
}

export function rowsFromToggles(sheetId: Id, toggles: ConceptToggles): StateRelationRow[] {
	return [
		...Object.keys(toggles.edges)
			.sort()
			.map((name) => edgeRow(sheetId, name, toggles.edges[name] ?? {})),
		...Object.keys(toggles.labels)
			.sort()
			.map((name) => labelRow(sheetId, name, toggles.labels[name] ?? {}))
	];
}

function edgeRow(
	sheetId: Id,
	name: string,
	values: Partial<Record<EdgeDisplayClass, boolean>>
): StateRelationRow {
	return {
		id: `${sheetId}:edge:${name}`,
		sheetId,
		kind: 'edge',
		name,
		columns: edgeColumns.map((column) => ({ ...column, checked: values[column.key] ?? false }))
	};
}

function labelRow(
	sheetId: Id,
	name: string,
	values: Partial<Record<NodeLabelDisplayClass, boolean>>
): StateRelationRow {
	return {
		id: `${sheetId}:label:${name}`,
		sheetId,
		kind: 'label',
		name,
		columns: labelColumns.map((column) => ({ ...column, checked: values[column.key] ?? false }))
	};
}

function sortedUnique(values: string[]) {
	return Array.from(new Set(values.filter(Boolean))).sort();
}
