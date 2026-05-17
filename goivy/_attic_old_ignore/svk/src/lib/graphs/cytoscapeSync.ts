import type cytoscape from 'cytoscape';
import type { GraphSnapshot } from '$lib/types';
import { graphSnapshotToCytoscapeElements, type GraphElementDefinition } from './cytoscapeElements';

export function syncCytoscapeSnapshot(cy: cytoscape.Core, snapshot: GraphSnapshot) {
	const elements = graphSnapshotToCytoscapeElements(snapshot);
	const desiredIds = new Set(elements.map((element) => element.data.id));

	cy.elements().forEach((element: cytoscape.SingularElementReturnValue) => {
		if (!desiredIds.has(element.id())) {
			element.remove();
		}
	});

	for (const element of elements) {
		const existing = cy.getElementById(element.data.id);
		if (existing?.length) {
			updateCytoscapeElement(existing, element);
		} else {
			cy.add(element);
		}
	}

	cy.layout({ name: 'preset', fit: true, padding: 24 }).run();
}

function updateCytoscapeElement(existing: cytoscape.Collection, element: GraphElementDefinition) {
	existing.data(element.data);
	existing.classes(element.classes ?? '');
	const singular = existing[0];
	if (element.position && singular?.isNode()) {
		singular.position(element.position);
	}
}
