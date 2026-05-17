import { describe, expect, it } from 'vitest';
import {
	createContextMenuUiState,
	createDetailsUiState,
	createDialogUiState,
	createEditorUiState,
	createEventTraceUiState,
	createLayoutUiState,
	createMenuUiState,
	createRecentFilesUiState,
	createSessionUiState,
	createToastUiState,
	createWorkbenchSheetsState
} from './uiStores.svelte';
import type { TraceEventSheet } from '$lib/types';

describe('workbench ui stores', () => {
	it('tracks session status, mode, loading, and loaded file display', () => {
		expect.hasAssertions();
		const session = createSessionUiState();
		session.setSessionId('s1');
		session.setMode('induction');
		session.setStatus('Checking...', 'info');
		session.setLoadedFile('client.ivy', '/tmp/client.ivy');
		session.showLoading('Starting');

		expect(session.current).toMatchObject({
			sessionDisplay: 'Session: s1',
			mode: 'induction',
			status: 'Checking...',
			statusLevel: 'info',
			loadedFileDisplay: '/tmp/client.ivy',
			loading: true,
			loadingMessage: 'Starting'
		});

		session.setMode('unknown');
		session.hideLoading();
		expect(session.current.mode).toBe('pdr');
		expect(session.current.loading).toBe(false);
	});

	it('derives editor dirty and label state', () => {
		expect.hasAssertions();
		const editor = createEditorUiState({ path: 'model.ivy', content: 'type t' });
		expect(editor.dirty).toBe(false);
		expect(editor.label).toBe('model.ivy [saved]');

		editor.edit('type u');
		expect(editor.dirty).toBe(true);
		expect(editor.label).toBe('** model.ivy');

		editor.markSaving();
		expect(editor.label).toBe('model.ivy [saving...]');
		editor.markSaved();
		editor.setKeymap('vim');
		expect(editor.current.keymap).toBe('vim');
		expect(editor.dirty).toBe(false);
	});

	it('tracks tutorial navigation and panel sizes', () => {
		expect.hasAssertions();
		const layout = createLayoutUiState();
		layout.navigateTutorial('/static/tutorial/page-2.html');
		layout.navigateTutorial('/static/tutorial/page-3.html');
		expect(layout.current.tutorialUrl).toBe('/static/tutorial/page-3.html');
		expect(layout.canGoBack).toBe(true);

		layout.goTutorialBack();
		expect(layout.current.tutorialUrl).toBe('/static/tutorial/page-2.html');
		expect(layout.canGoForward).toBe(true);
		layout.setPanelSize('editor', 320);
		expect(layout.current.editorWidth).toBe(320);
	});

	it('tracks menus, dialogs, context menus, toasts, sheets, traces, details, and recents', () => {
		expect.hasAssertions();
		const menus = createMenuUiState();
		expect(menus.toggle('file')).toBe(true);
		menus.setDescriptor('arg', [{ id: 'arg.check', label: 'Check' }]);
		expect(menus.isOpen('file')).toBe(true);
		expect(menus.current.descriptors.arg).toHaveLength(1);

		const dialogs = createDialogUiState();
		dialogs.open({ id: 'd1', kind: 'message', title: 'Hi', message: 'There' });
		dialogs.open({ id: 'd2', kind: 'confirm', title: 'Confirm', message: 'Proceed?' });
		expect(dialogs.current.active?.id).toBe('d1');
		dialogs.resolve();
		expect(dialogs.current.active?.id).toBe('d2');

		const contextMenu = createContextMenuUiState();
		contextMenu.show(10, 20, [{ id: 'expand', label: 'Expand' }]);
		expect(contextMenu.current).toMatchObject({ visible: true, x: 10, y: 20 });

		const toasts = createToastUiState(() => 42);
		toasts.push({ id: 't1', level: 'success', message: 'Saved' });
		expect(toasts.current.toasts[0]).toMatchObject({ createdAt: 42, message: 'Saved' });

		const sheets = createWorkbenchSheetsState();
		sheets.upsert({ id: 'events-1', label: 'Trace', type: 'events', visualOnly: true });
		sheets.activate('events-1');
		expect(sheets.current.activeSheetId).toBe('events-1');

		const traceSheet: TraceEventSheet = {
			id: 'events-1',
			projectId: 'project-1',
			sessionId: 'session-1',
			label: 'Trace',
			events: [{ id: 'event-1', text: 'root', address: '', children: [{ id: 'event-2', text: 'child', address: '' }] }],
			patterns: [],
			expandedAddresses: [],
			revision: 1
		};
		const traces = createEventTraceUiState();
		traces.upsert(traceSheet);
		traces.select('events-1', '0/0');
		traces.setExpanded('events-1', '0', true);
		expect(traces.current.sheets['events-1'].events[0]?.children?.[0]?.address).toBe('0/0');
		expect(traces.current.sheets['events-1'].selectedAddress).toBe('0/0');

		const details = createDetailsUiState();
		details.setFacts([{ text: 'link(X,Y)' }]);
		details.toggleFact(0);
		expect(details.current.facts[0]).toMatchObject({ selected: false });

		const recent = createRecentFilesUiState();
		recent.remember({ id: 'a', label: 'a.ivy' });
		recent.remember({ id: 'a', label: 'a.ivy', sessionId: 's1' });
		expect(recent.current.files).toEqual([{ id: 'a', label: 'a.ivy', sessionId: 's1' }]);
	});
});
