import type { EntityTable, GraphSnapshot, Id } from '$lib/types';
import { emptyEntityTable } from '$lib/types';

export function createGraphsState(initial: EntityTable<GraphSnapshot> = emptyEntityTable<GraphSnapshot>()) {
	const table = $state<EntityTable<GraphSnapshot>>(structuredClone(initial));

	function bump() {
		table.revision += 1;
	}

	return {
		get table() {
			return table;
		},
		getById(id: Id) {
			return table.byId[id];
		},
		upsert(snapshot: GraphSnapshot) {
			if (!Object.hasOwn(table.byId, snapshot.id)) {
				table.order.push(snapshot.id);
			}
			table.byId[snapshot.id] = snapshot;
			bump();
		},
		graphsForSheet(sheetId: Id) {
			return table.order.map((id) => table.byId[id]).filter((graph) => graph.sheetId === sheetId);
		}
	};
}
