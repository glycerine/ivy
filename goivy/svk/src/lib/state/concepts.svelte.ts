import type { ConceptState, EntityTable, Id } from '$lib/types';
import { emptyEntityTable } from '$lib/types';

export function createConceptsState(initial: EntityTable<ConceptState> = emptyEntityTable<ConceptState>()) {
	const table = $state<EntityTable<ConceptState>>(structuredClone(initial));

	function bump() {
		table.revision += 1;
	}

	return {
		get table() {
			return table;
		},
		upsert(concept: ConceptState) {
			if (!Object.hasOwn(table.byId, concept.id)) {
				table.order.push(concept.id);
			}
			table.byId[concept.id] = concept;
			bump();
		},
		selectConcept(id: Id, selectedConcept: string | undefined) {
			const concept = table.byId[id];
			if (!concept) {
				return;
			}
			concept.selectedConcept = selectedConcept;
			concept.revision += 1;
			bump();
		}
	};
}
