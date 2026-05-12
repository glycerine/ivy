export type ParityKind =
	| 'component'
	| 'menu'
	| 'command'
	| 'dialog'
	| 'store'
	| 'service'
	| 'workflow';

export type ParityStatus = 'implemented' | 'planned';

export type WebuiParityItem = {
	id: string;
	kind: ParityKind;
	label: string;
	source: string;
	status: ParityStatus;
	implementation?: string;
};

export const WEBUI_PARITY_ITEMS = [
	{
		id: 'component.workspace-shell',
		kind: 'component',
		label: 'Workspace shell',
		source: 'goivy/webui/frontend/src/components/WorkspaceShell.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/WorkbenchShell.svelte'
	},
	{
		id: 'component.menubar',
		kind: 'component',
		label: 'Menubar',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/Menubar.svelte'
	},
	{
		id: 'component.status-bar',
		kind: 'component',
		label: 'Status bar',
		source: 'goivy/webui/frontend/src/components/StatusBar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/StatusBar.svelte'
	},
	{
		id: 'component.sheet-area',
		kind: 'component',
		label: 'Sheet area',
		source: 'goivy/webui/frontend/src/components/panes/SheetArea.vue',
		status: 'planned'
	},
	{
		id: 'component.tab-bar',
		kind: 'component',
		label: 'Tab bar',
		source: 'goivy/webui/frontend/src/components/panes/TabBar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/panes/TabBar.svelte'
	},
	{
		id: 'component.arg-pane',
		kind: 'component',
		label: 'ARG pane',
		source: 'goivy/webui/frontend/src/components/panes/ArgPane.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/panes/ArgPane.svelte'
	},
	{
		id: 'component.concept-pane',
		kind: 'component',
		label: 'Concept pane',
		source: 'goivy/webui/frontend/src/components/panes/ConceptPane.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/panes/ConceptPane.svelte'
	},
	{
		id: 'component.state-relations-pane',
		kind: 'component',
		label: 'State relations pane',
		source: 'goivy/webui/frontend/src/components/panes/StateRelationsPane.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/panes/StateRelationsPane.svelte'
	},
	{
		id: 'component.details-pane',
		kind: 'component',
		label: 'Details pane',
		source: 'goivy/webui/frontend/src/components/panes/DetailsPane.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/panes/DetailsPane.svelte'
	},
	{
		id: 'component.editor-pane',
		kind: 'component',
		label: 'Editor pane',
		source: 'goivy/webui/frontend/src/components/panes/EditorPane.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/panes/EditorPane.svelte'
	},
	{
		id: 'component.event-trace-sheet',
		kind: 'component',
		label: 'Event trace sheet',
		source: 'goivy/webui/frontend/src/components/panes/EventTraceSheet.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/panes/EventTraceSheet.svelte'
	},
	{
		id: 'component.tutorial-pane',
		kind: 'component',
		label: 'Tutorial pane',
		source: 'goivy/webui/frontend/src/components/panes/TutorialPane.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/panes/TutorialPane.svelte'
	},
	{
		id: 'component.dialog-host',
		kind: 'component',
		label: 'Dialog host',
		source: 'goivy/webui/frontend/src/components/DialogHost.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/overlays/DialogHost.svelte'
	},
	{
		id: 'component.context-menu-host',
		kind: 'component',
		label: 'Context menu host',
		source: 'goivy/webui/frontend/src/components/ContextMenuHost.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/overlays/ContextMenuHost.svelte'
	},
	{
		id: 'component.toast-host',
		kind: 'component',
		label: 'Toast host',
		source: 'goivy/webui/frontend/src/components/ToastHost.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/overlays/ToastHost.svelte'
	},
	{
		id: 'component.session-overlay-host',
		kind: 'component',
		label: 'Session overlay host',
		source: 'goivy/webui/frontend/src/components/SessionOverlayHost.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/overlays/SessionOverlayHost.svelte'
	},
	{
		id: 'component.file-input-host',
		kind: 'component',
		label: 'File input host',
		source: 'goivy/webui/frontend/src/components/FileInputHost.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/overlays/FileInputHost.svelte'
	},
	{
		id: 'menu.file.load',
		kind: 'menu',
		label: 'File / Load',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.file.open-event-trace',
		kind: 'menu',
		label: 'File / Open Event Trace',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.file.save-as',
		kind: 'menu',
		label: 'File / Save as',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.file.download-current-model',
		kind: 'menu',
		label: 'File / Download current model',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.file.save-analysis-state',
		kind: 'menu',
		label: 'File / Save Analysis State',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.file.load-analysis-state',
		kind: 'menu',
		label: 'File / Load Analysis State',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.file.save-invariant',
		kind: 'menu',
		label: 'File / Save Invariant',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.file.recent-files',
		kind: 'menu',
		label: 'File / Recent files',
		source: 'goivy/webui/frontend/src/components/RecentFilesMenu.vue',
		status: 'planned'
	},
	{
		id: 'menu.file.new-model',
		kind: 'menu',
		label: 'File / New Model',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.toolbar.check',
		kind: 'menu',
		label: 'Toolbar / Check',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/Menubar.svelte'
	},
	{
		id: 'menu.toolbar.show-reachable',
		kind: 'menu',
		label: 'Toolbar / Show Reachable',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/Menubar.svelte'
	},
	{
		id: 'menu.toolbar.undo',
		kind: 'menu',
		label: 'Toolbar / Undo',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/Menubar.svelte'
	},
	{
		id: 'menu.toolbar.reset-domain',
		kind: 'menu',
		label: 'Toolbar / Reset Domain',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/Menubar.svelte'
	},
	{
		id: 'menu.toolbar.diagram-domain',
		kind: 'menu',
		label: 'Toolbar / Diagram Domain',
		source: 'goivy/webui/frontend/src/components/Menubar.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/components/Menubar.svelte'
	},
	{
		id: 'menu.arg.check-induction',
		kind: 'menu',
		label: 'ARG Invariant / Check induction',
		source: 'goivy/webui/frontend/src/components/panes/ArgPane.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.arg.bounded-check',
		kind: 'menu',
		label: 'ARG Invariant / Bounded check',
		source: 'goivy/webui/frontend/src/components/panes/ArgPane.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.arg.diagram',
		kind: 'menu',
		label: 'ARG Invariant / Diagram',
		source: 'goivy/webui/frontend/src/components/panes/ArgPane.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.arg.weaken',
		kind: 'menu',
		label: 'ARG Invariant / Weaken',
		source: 'goivy/webui/frontend/src/components/panes/ArgPane.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.arg.save-abstraction',
		kind: 'menu',
		label: 'ARG Invariant / Save Abstraction',
		source: 'goivy/webui/frontend/src/components/panes/ArgPane.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.concept.conjecture',
		kind: 'menu',
		label: 'Concept / Conjecture actions',
		source: 'goivy/webui/frontend/src/components/panes/ConceptPane.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.concept.add-relation',
		kind: 'menu',
		label: 'Concept View / Add relation',
		source: 'goivy/webui/frontend/src/components/panes/ConceptPane.vue',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/menuDefinitions.ts'
	},
	{
		id: 'menu.event-trace.controls',
		kind: 'menu',
		label: 'Event trace controls',
		source: 'goivy/webui/frontend/src/components/panes/EventTraceSheet.vue',
		status: 'planned'
	},
	{
		id: 'menu.editor.keymaps',
		kind: 'menu',
		label: 'Editor keymaps',
		source: 'goivy/webui/frontend/src/components/panes/EditorPane.vue',
		status: 'planned'
	},
	{
		id: 'command.registry',
		kind: 'command',
		label: 'Command registry',
		source: 'goivy/webui/frontend/src/services/commandRegistry.js',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/commands/commandRegistry.ts'
	},
	{
		id: 'command.file',
		kind: 'command',
		label: 'File commands',
		source: 'goivy/webui/frontend/src/services/fileCommands.js',
		status: 'planned'
	},
	{
		id: 'command.analysis-action',
		kind: 'command',
		label: 'Analysis action commands',
		source: 'goivy/webui/frontend/src/services/analysisActionCommands.js',
		status: 'planned'
	},
	{
		id: 'command.arg',
		kind: 'command',
		label: 'ARG commands',
		source: 'goivy/webui/frontend/src/services/argCommands.js',
		status: 'planned'
	},
	{
		id: 'command.concept',
		kind: 'command',
		label: 'Concept commands',
		source: 'goivy/webui/frontend/src/services/conceptCommands.js',
		status: 'planned'
	},
	{
		id: 'command.event-trace',
		kind: 'command',
		label: 'Event trace commands',
		source: 'goivy/webui/frontend/src/services/eventTraceCommands.js',
		status: 'planned'
	},
	{
		id: 'dialog.all-types',
		kind: 'dialog',
		label: 'All webui dialog types',
		source: 'goivy/webui/frontend/src/services/dialogCommands.js',
		status: 'planned'
	},
	{
		id: 'store.session',
		kind: 'store',
		label: 'Session store',
		source: 'goivy/webui/frontend/src/stores/sessionStore.js',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/state/uiStores.svelte.ts'
	},
	{
		id: 'store.layout',
		kind: 'store',
		label: 'Layout store',
		source: 'goivy/webui/frontend/src/stores/layoutStore.js',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/state/uiStores.svelte.ts'
	},
	{
		id: 'store.editor',
		kind: 'store',
		label: 'Editor store',
		source: 'goivy/webui/frontend/src/stores/editorStore.js',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/state/uiStores.svelte.ts'
	},
	{
		id: 'store.details',
		kind: 'store',
		label: 'Details store',
		source: 'goivy/webui/frontend/src/stores/detailsStore.js',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/state/uiStores.svelte.ts'
	},
	{
		id: 'store.event-trace',
		kind: 'store',
		label: 'Event trace store',
		source: 'goivy/webui/frontend/src/stores/eventTraceStore.js',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/workbench/state/uiStores.svelte.ts'
	},
	{
		id: 'service.engine-neutrality',
		kind: 'service',
		label: 'Neutral engine service',
		source: 'goivy/svk/src/lib/services/engineService.ts',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/services/engineService.ts'
	},
	{
		id: 'workflow.local-first-wasm',
		kind: 'workflow',
		label: 'Local-first browser wasm execution',
		source: 'goivy/svk/src/lib/engines/browserWasmEngine.ts',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/engines/browserWasmEngine.ts'
	},
	{
		id: 'workflow.hosted-webui',
		kind: 'workflow',
		label: 'Hosted webui execution',
		source: 'goivy/svk/src/lib/engines/hostedWebuiEngine.ts',
		status: 'implemented',
		implementation: 'goivy/svk/src/lib/engines/hostedWebuiEngine.ts'
	}
] as const satisfies readonly WebuiParityItem[];

export function paritySummary(items: readonly WebuiParityItem[] = WEBUI_PARITY_ITEMS) {
	return items.reduce(
		(summary, item) => {
			summary.total += 1;
			summary.byStatus[item.status] += 1;
			summary.byKind[item.kind] = (summary.byKind[item.kind] ?? 0) + 1;
			return summary;
		},
		{
			total: 0,
			byStatus: { implemented: 0, planned: 0 },
			byKind: {} as Record<ParityKind, number>
		}
	);
}
