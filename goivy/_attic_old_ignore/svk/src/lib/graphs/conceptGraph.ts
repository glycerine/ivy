import type { ConceptState, GraphSnapshot } from '$lib/types';

export function conceptStateToGraphSnapshot(concept: ConceptState): GraphSnapshot {
	const names = Object.keys(concept.concepts).sort();
	const nodes = Object.fromEntries(
		names.map((name) => [
			`concept:${name}`,
			{
				id: `concept:${name}`,
				obj: name,
				label: name,
				classes: ['concept'],
				shape: 'octagon',
				shortInfo: concept.concepts[name]?.formula
			}
		])
	);
	const layout = Object.fromEntries(
		names.map((name, index) => [
			`concept:${name}`,
			{
				x: 140 + (index % 3) * 170,
				y: 120 + Math.floor(index / 3) * 150
			}
		])
	);

	return {
		id: `concept-graph:${concept.id}`,
		sheetId: concept.sheetId,
		kind: 'concept',
		sourceRevision: concept.revision,
		nodes,
		edges: {},
		nodeOrder: names.map((name) => `concept:${name}`),
		edgeOrder: [],
		layout,
		styleRevision: 1,
		createdAt: new Date(0).toISOString()
	};
}
