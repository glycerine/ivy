import type { EngineSession, EntityTable, Id } from '$lib/types';
import { emptyEntityTable } from '$lib/types';

export function createEnginesState(initial: EntityTable<EngineSession> = emptyEntityTable<EngineSession>()) {
	const table = $state<EntityTable<EngineSession>>(structuredClone(initial));

	function bump() {
		table.revision += 1;
	}

	return {
		get table() {
			return table;
		},
		upsert(session: EngineSession) {
			if (!Object.hasOwn(table.byId, session.id)) {
				table.order.push(session.id);
			}
			table.byId[session.id] = session;
			bump();
		},
		setStatus(id: Id, status: EngineSession['status'], error?: string, updatedAt = new Date().toISOString()) {
			const session = table.byId[id];
			if (!session) {
				return;
			}
			session.status = status;
			session.error = error;
			session.updatedAt = updatedAt;
			bump();
		}
	};
}
