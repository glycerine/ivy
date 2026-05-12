import type { EntityTable, Id, Project } from '$lib/types';
import { emptyEntityTable } from '$lib/types';

export function createProjectsState(initial: EntityTable<Project> = emptyEntityTable<Project>()) {
	const table = $state<EntityTable<Project>>(structuredClone(initial));

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
		replaceAll(projects: Project[]) {
			table.byId = {};
			table.order = [];
			for (const project of projects) {
				table.byId[project.id] = project;
				table.order.push(project.id);
			}
			bump();
		},
		upsert(project: Project) {
			if (!Object.hasOwn(table.byId, project.id)) {
				table.order.push(project.id);
			}
			table.byId[project.id] = project;
			bump();
		}
	};
}
