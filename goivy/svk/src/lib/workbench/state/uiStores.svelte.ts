import type { Id, TraceEvent, TraceEventSheet } from '$lib/types';

export const SESSION_MODES = ['induction', 'pdr', 'concrete', 'abstract', 'bounded'] as const;
export type SessionMode = (typeof SESSION_MODES)[number];
export type StatusLevel = '' | 'info' | 'success' | 'warning' | 'error';
export const EDITOR_KEYMAPS = ['sublime', 'emacs', 'vim'] as const;
export type EditorKeymap = (typeof EDITOR_KEYMAPS)[number];

export type MenuDescriptorItem = {
	id: string;
	label: string;
	action?: string;
	enabled?: boolean;
	separator?: boolean;
	dispatch?: 'action' | 'command';
};

export type DialogRequest =
	| { id: string; kind: 'message'; title: string; message: string }
	| { id: string; kind: 'confirm'; title: string; message: string }
	| { id: string; kind: 'entry'; title: string; message: string; value: string }
	| { id: string; kind: 'list'; title: string; message: string; choices: string[] };

export type ContextMenuItem = {
	id: string;
	label: string;
	enabled?: boolean;
	separator?: boolean;
};

export type Toast = {
	id: string;
	level: Exclude<StatusLevel, ''>;
	message: string;
	createdAt: number;
};

export type WorkbenchSheet = {
	id: Id;
	label: string;
	type: 'analysis' | 'events';
	closable: boolean;
	visualOnly: boolean;
};

const ROOT_SHEET: WorkbenchSheet = {
	id: 'sheet-1',
	label: 'Sheet 1',
	type: 'analysis',
	closable: false,
	visualOnly: false
};

export function createSessionUiState() {
	const current = $state({
		sessionId: '',
		sessionDisplay: '',
		mode: 'pdr' as SessionMode,
		loadedFileDisplay: '',
		loadedFileTitle: '',
		status: 'Ready',
		statusLevel: '' as StatusLevel,
		loading: false,
		loadingMessage: 'Loading...'
	});

	return {
		get current() {
			return current;
		},
		setStatus(message: string, level: StatusLevel = '') {
			current.status = message;
			current.statusLevel = level;
		},
		setSessionId(sessionId = '') {
			current.sessionId = sessionId;
			current.sessionDisplay = sessionId ? `Session: ${sessionId}` : '';
		},
		setMode(mode: string) {
			current.mode = isSessionMode(mode) ? mode : 'pdr';
		},
		setLoadedFile(fileName = '', filePath = '') {
			const display = filePath || fileName;
			current.loadedFileDisplay = display;
			current.loadedFileTitle = display;
		},
		showLoading(message = 'Loading...') {
			current.loading = true;
			current.loadingMessage = message;
		},
		hideLoading() {
			current.loading = false;
		}
	};
}

export function createEditorUiState(initial = { path: '', content: '' }) {
	const current = $state({
		path: initial.path,
		content: initial.content,
		savedContent: initial.content,
		saveState: 'idle' as 'idle' | 'saving' | 'saved',
		keymap: 'sublime' as EditorKeymap,
		reopenLastVisible: false,
		reopenLastLabel: ''
	});

	return {
		get current() {
			return current;
		},
		get dirty() {
			return current.content !== current.savedContent;
		},
		get label() {
			const base = current.path || '(unsaved file)';
			if (current.saveState === 'saving') return `${base} [saving...]`;
			if (current.content !== current.savedContent) return `** ${base}`;
			return current.path ? `${base} [saved]` : base;
		},
		load(path: string, content: string) {
			current.path = path;
			current.content = content;
			current.savedContent = content;
			current.saveState = 'idle';
		},
		edit(content: string) {
			current.content = content;
			if (current.saveState === 'saved') current.saveState = 'idle';
		},
		markSaving() {
			current.saveState = 'saving';
		},
		markSaved(content = current.content) {
			current.savedContent = content;
			current.saveState = 'saved';
		},
		setKeymap(keymap: string) {
			current.keymap = isEditorKeymap(keymap) ? keymap : 'sublime';
		},
		setReopenLast(visible: boolean, label = '') {
			current.reopenLastVisible = visible;
			current.reopenLastLabel = visible ? label || 'Re-open last file' : '';
		}
	};
}

export function createLayoutUiState() {
	const current = $state({
		tutorialVisible: false,
		tutorialUrl: '/static/tutorial/kenmcmil.github.io/ivy/language.html',
		tutorialHistory: ['/static/tutorial/kenmcmil.github.io/ivy/language.html'],
		tutorialHistoryIndex: 0,
		argPanelWidth: 0,
		statePanelWidth: 0,
		detailsHeight: 0,
		editorWidth: 0,
		tutorialHeight: 0
	});

	return {
		get current() {
			return current;
		},
		get canGoBack() {
			return current.tutorialHistoryIndex > 0;
		},
		get canGoForward() {
			return current.tutorialHistoryIndex < current.tutorialHistory.length - 1;
		},
		setTutorialVisible(visible: boolean) {
			current.tutorialVisible = visible;
		},
		navigateTutorial(url: string) {
			if (!url) return;
			current.tutorialHistory = current.tutorialHistory.slice(0, current.tutorialHistoryIndex + 1);
			current.tutorialHistory.push(url);
			current.tutorialHistoryIndex = current.tutorialHistory.length - 1;
			current.tutorialUrl = url;
		},
		goTutorialBack() {
			if (current.tutorialHistoryIndex <= 0) return;
			current.tutorialHistoryIndex -= 1;
			current.tutorialUrl = current.tutorialHistory[current.tutorialHistoryIndex] ?? current.tutorialUrl;
		},
		goTutorialForward() {
			if (current.tutorialHistoryIndex >= current.tutorialHistory.length - 1) return;
			current.tutorialHistoryIndex += 1;
			current.tutorialUrl = current.tutorialHistory[current.tutorialHistoryIndex] ?? current.tutorialUrl;
		},
		setPanelSize(panel: 'arg' | 'state' | 'details' | 'editor' | 'tutorial', size: number) {
			const clamped = Math.max(0, Math.trunc(size));
			if (panel === 'arg') current.argPanelWidth = clamped;
			if (panel === 'state') current.statePanelWidth = clamped;
			if (panel === 'details') current.detailsHeight = clamped;
			if (panel === 'editor') current.editorWidth = clamped;
			if (panel === 'tutorial') current.tutorialHeight = clamped;
		}
	};
}

export function createMenuUiState() {
	const current = $state({
		openId: '',
		flashingItemId: '',
		descriptors: {} as Record<string, MenuDescriptorItem[]>
	});

	return {
		get current() {
			return current;
		},
		isOpen(id: string) {
			return current.openId === id;
		},
		toggle(id: string) {
			current.openId = current.openId === id ? '' : id;
			return current.openId === id;
		},
		closeAll() {
			current.openId = '';
		},
		flash(id: string) {
			current.flashingItemId = id;
		},
		clearFlash() {
			current.flashingItemId = '';
		},
		setDescriptor(region: string, items: MenuDescriptorItem[]) {
			current.descriptors[region] = items;
		}
	};
}

export function createDialogUiState() {
	const current = $state({ active: null as DialogRequest | null, queue: [] as DialogRequest[] });
	return {
		get current() {
			return current;
		},
		open(dialog: DialogRequest) {
			if (current.active) current.queue.push(dialog);
			else current.active = dialog;
		},
		resolve() {
			current.active = current.queue.shift() ?? null;
		},
		clear() {
			current.active = null;
			current.queue = [];
		}
	};
}

export function createContextMenuUiState() {
	const current = $state({ visible: false, x: 0, y: 0, items: [] as ContextMenuItem[] });
	return {
		get current() {
			return current;
		},
		show(x: number, y: number, items: ContextMenuItem[]) {
			current.visible = true;
			current.x = x;
			current.y = y;
			current.items = items;
		},
		hide() {
			current.visible = false;
			current.items = [];
		}
	};
}

export function createToastUiState(now = () => Date.now()) {
	const current = $state({ toasts: [] as Toast[] });
	return {
		get current() {
			return current;
		},
		push(toast: Omit<Toast, 'createdAt'>) {
			current.toasts.push({ ...toast, createdAt: now() });
		},
		dismiss(id: string) {
			current.toasts = current.toasts.filter((toast) => toast.id !== id);
		},
		clear() {
			current.toasts = [];
		}
	};
}

export function createWorkbenchSheetsState() {
	const current = $state({ tabs: [ROOT_SHEET], activeSheetId: ROOT_SHEET.id });
	return {
		get current() {
			return current;
		},
		upsert(sheet: Partial<WorkbenchSheet> & { id: Id }) {
			const next: WorkbenchSheet = { ...ROOT_SHEET, ...sheet, closable: sheet.id !== ROOT_SHEET.id && sheet.closable !== false };
			const index = current.tabs.findIndex((tab) => tab.id === sheet.id);
			if (index >= 0) current.tabs[index] = { ...current.tabs[index], ...next };
			else current.tabs.push(next);
		},
		activate(sheetId: Id) {
			if (current.tabs.some((tab) => tab.id === sheetId)) current.activeSheetId = sheetId;
		},
		remove(sheetId: Id) {
			if (sheetId === ROOT_SHEET.id) return;
			current.tabs = current.tabs.filter((tab) => tab.id !== sheetId);
			if (current.activeSheetId === sheetId) current.activeSheetId = ROOT_SHEET.id;
		},
		reset() {
			current.tabs = [ROOT_SHEET];
			current.activeSheetId = ROOT_SHEET.id;
		}
	};
}

export function createEventTraceUiState() {
	const current = $state({ sheets: {} as Record<Id, TraceEventSheet> });
	return {
		get current() {
			return current;
		},
		upsert(sheet: TraceEventSheet) {
			current.sheets[sheet.id] = normalizeSheet(sheet);
		},
		select(sheetId: Id, address: string) {
			const sheet = current.sheets[sheetId];
			if (sheet) sheet.selectedAddress = address;
		},
		setExpanded(sheetId: Id, address: string, expanded: boolean) {
			const sheet = current.sheets[sheetId];
			if (!sheet) return;
			const withoutAddress = sheet.expandedAddresses.filter((candidate) => candidate !== address);
			sheet.expandedAddresses = (expanded ? [...withoutAddress, address] : withoutAddress).sort();
		},
		remove(sheetId: Id) {
			delete current.sheets[sheetId];
		}
	};
}

export function createDetailsUiState() {
	const current = $state({
		text: 'Select a node or edge to see details',
		facts: [] as Array<{ index: number; text: string; selected: boolean }>,
		traceActionVisible: false
	});
	return {
		get current() {
			return current;
		},
		setText(text: string) {
			current.text = text || 'Select a node or edge to see details';
			current.facts = [];
			current.traceActionVisible = false;
		},
		setFacts(facts: Array<{ index?: number; text: string; selected?: boolean }>) {
			current.facts = facts.map((fact, index) => ({
				index: fact.index ?? index,
				text: fact.text,
				selected: fact.selected !== false
			}));
			current.text = current.facts.length ? '' : 'Select a node or edge to see details';
		},
		toggleFact(index: number) {
			const fact = current.facts.find((candidate) => candidate.index === index);
			if (fact) fact.selected = !fact.selected;
		},
		setTraceActionVisible(visible: boolean) {
			current.traceActionVisible = visible;
		}
	};
}

export function createRecentFilesUiState() {
	const current = $state({ files: [] as Array<{ id: string; label: string; sessionId?: string }> });
	return {
		get current() {
			return current;
		},
		remember(file: { id: string; label: string; sessionId?: string }) {
			current.files = [file, ...current.files.filter((candidate) => candidate.id !== file.id)].slice(0, 20);
		},
		clear() {
			current.files = [];
		}
	};
}

function isSessionMode(mode: string): mode is SessionMode {
	return (SESSION_MODES as readonly string[]).includes(mode);
}

function isEditorKeymap(keymap: string): keymap is EditorKeymap {
	return (EDITOR_KEYMAPS as readonly string[]).includes(keymap);
}

function normalizeSheet(sheet: TraceEventSheet): TraceEventSheet {
	return { ...sheet, events: normalizeEvents(sheet.events), expandedAddresses: sheet.expandedAddresses ?? [] };
}

function normalizeEvents(events: TraceEvent[], prefix = ''): TraceEvent[] {
	return events.map((event, index) => {
		const address = event.address || (prefix ? `${prefix}/${index}` : String(index));
		return { ...event, address, children: normalizeEvents(event.children ?? [], address) };
	});
}
