import type { Id, WorkspaceState } from '$lib/types';

export function createWorkspaceState(initial: Partial<WorkspaceState> = {}) {
	const state = $state<WorkspaceState>({
		activeProjectId: null,
		activeModelId: null,
		activeEngineId: null,
		activeSessionId: null,
		activeSheetId: null,
		online: true,
		hydrated: false,
		...initial
	});

	return {
		get current() {
			return state;
		},
		hydrate() {
			state.hydrated = true;
		},
		setOnline(online: boolean) {
			state.online = online;
		},
		selectProject(projectId: Id | null) {
			state.activeProjectId = projectId;
			state.activeModelId = null;
			state.activeSessionId = null;
			state.activeSheetId = null;
		},
		selectModel(modelId: Id | null) {
			state.activeModelId = modelId;
		},
		selectEngine(engineId: Id | null) {
			state.activeEngineId = engineId;
			state.activeSessionId = null;
		},
		selectSession(sessionId: Id | null) {
			state.activeSessionId = sessionId;
		},
		selectSheet(sheetId: Id | null) {
			state.activeSheetId = sheetId;
		}
	};
}
