import { describe, expect, it } from 'vitest';
import { createWorkspaceState } from './workspace.svelte';

describe('createWorkspaceState', () => {
	it('clears model, session, and sheet selection when project changes', () => {
		expect.hasAssertions();

		const workspace = createWorkspaceState({
			activeProjectId: 'project-1',
			activeModelId: 'model-1',
			activeSessionId: 'session-1',
			activeSheetId: 'sheet-1'
		});

		workspace.selectProject('project-2');

		expect(workspace.current.activeProjectId).toBe('project-2');
		expect(workspace.current.activeModelId).toBeNull();
		expect(workspace.current.activeSessionId).toBeNull();
		expect(workspace.current.activeSheetId).toBeNull();
	});

	it('tracks hydration and online status without side effects', () => {
		expect.hasAssertions();

		const workspace = createWorkspaceState({ online: true });
		workspace.hydrate();
		workspace.setOnline(false);

		expect(workspace.current.hydrated).toBe(true);
		expect(workspace.current.online).toBe(false);
	});
});
