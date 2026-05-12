export type Id = string;
export type Revision = number;

export type EntityTable<T> = {
	byId: Record<Id, T>;
	order: Id[];
	revision: Revision;
};

export function emptyEntityTable<T>(): EntityTable<T> {
	return {
		byId: {},
		order: [],
		revision: 0
	};
}

export function upsertEntity<T extends { id: Id }>(table: EntityTable<T>, entity: T): EntityTable<T> {
	const exists = Object.hasOwn(table.byId, entity.id);
	return {
		byId: {
			...table.byId,
			[entity.id]: entity
		},
		order: exists ? table.order : [...table.order, entity.id],
		revision: table.revision + 1
	};
}
