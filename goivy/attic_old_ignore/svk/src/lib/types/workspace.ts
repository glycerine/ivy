import type { Id } from './ids';

export type WorkspaceState = {
	activeProjectId: Id | null;
	activeModelId: Id | null;
	activeEngineId: Id | null;
	activeSessionId: Id | null;
	activeSheetId: Id | null;
	online: boolean;
	hydrated: boolean;
};
