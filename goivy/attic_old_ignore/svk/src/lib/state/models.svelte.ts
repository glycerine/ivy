import type { EntityTable, Id, ModelDocument } from '$lib/types';
import { emptyEntityTable } from '$lib/types';
import { nowIso } from '$lib/time';

export function createModelsState(initial: EntityTable<ModelDocument> = emptyEntityTable<ModelDocument>()) {
	const table = $state<EntityTable<ModelDocument>>(structuredClone(initial));

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
		upsert(model: ModelDocument) {
			if (!Object.hasOwn(table.byId, model.id)) {
				table.order.push(model.id);
			}
			table.byId[model.id] = model;
			bump();
		},
		updateText(id: Id, text: string, updatedAt = nowIso()) {
			const model = table.byId[id];
			if (!model) {
				return;
			}
			model.text = text;
			model.dirty = model.savedTextHash === undefined || model.savedTextHash !== text;
			model.updatedAt = updatedAt;
			model.parseRevision += 1;
			bump();
		},
		markSaved(id: Id, textHash: string, updatedAt = nowIso()) {
			const model = table.byId[id];
			if (!model) {
				return;
			}
			model.savedTextHash = textHash;
			model.dirty = false;
			model.updatedAt = updatedAt;
			bump();
		}
	};
}
