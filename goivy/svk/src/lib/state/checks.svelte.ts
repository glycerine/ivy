import type { CheckResult, EntityTable, Id } from '$lib/types';
import { emptyEntityTable } from '$lib/types';

export function createChecksState(initial: EntityTable<CheckResult> = emptyEntityTable<CheckResult>()) {
	const table = $state<EntityTable<CheckResult>>(structuredClone(initial));
	const latestBySession = $state<Record<Id, Id>>({});

	for (const id of table.order) {
		const result = table.byId[id];
		if (result) {
			latestBySession[result.sessionId] = result.id;
		}
	}

	function bump() {
		table.revision += 1;
	}

	return {
		get table() {
			return table;
		},
		get latestBySession() {
			return latestBySession;
		},
		upsert(result: CheckResult) {
			if (!Object.hasOwn(table.byId, result.id)) {
				table.order.push(result.id);
			}
			table.byId[result.id] = result;
			latestBySession[result.sessionId] = result.id;
			bump();
		},
		latestForSession(sessionId: Id): CheckResult | null {
			const id = latestBySession[sessionId];
			return id ? table.byId[id] ?? null : null;
		},
		statusText(sessionId: Id): string {
			const id = latestBySession[sessionId];
			const result = id ? table.byId[id] : null;
			if (!result) {
				return 'No check has run';
			}
			const z3 = result.z3Contacted ? 'Z3: yes' : 'Z3: no';
			return `Check ${result.result.toUpperCase()} (${result.mode}) [${z3}]`;
		}
	};
}
