<script lang="ts">
	import type {
		CheckResult,
		ConceptState,
		GraphNode,
		GraphSnapshot,
		ModelDocument,
		NodeAction,
		VerificationJob
	} from '$lib/types';
	import type { StateRelationRow } from '$lib/state/stateRelations.svelte';
	import type { SessionMode } from '$lib/workbench/state';
	import Menubar from './Menubar.svelte';
	import ContextMenuHost from './overlays/ContextMenuHost.svelte';
	import DialogHost from './overlays/DialogHost.svelte';
	import FileInputHost from './overlays/FileInputHost.svelte';
	import SessionOverlayHost from './overlays/SessionOverlayHost.svelte';
	import ToastHost from './overlays/ToastHost.svelte';
	import SheetArea from './panes/SheetArea.svelte';
	import TutorialPane from './panes/TutorialPane.svelte';

	type EngineChoice = 'hosted-go' | 'hosted-webui' | 'browser-wasm';

	type Props = {
		engineChoice: EngineChoice;
		mode: SessionMode;
		activeModel: ModelDocument;
		editorText: string;
		editorSaveState: 'idle' | 'saving' | 'saved';
		editorKeymap: 'sublime' | 'emacs' | 'vim';
		tutorialVisible: boolean;
		tutorialUrl: string;
		tutorialInput: string;
		tutorialCanGoBack: boolean;
		tutorialCanGoForward: boolean;
		tutorialFrameKey: number;
		tutorialButtonFlashing: boolean;
		sessionLabel: string;
		statusMessage: string;
		statusLevel: '' | 'info' | 'success' | 'warning' | 'error';
		latestGraph: GraphSnapshot | null;
		latestConcept: ConceptState | null;
		latestCheck: CheckResult | null;
		selectedNode: GraphNode | null;
		selectedNodeId: string | null;
		jobs: VerificationJob[];
		stateRelationRows: StateRelationRow[];
		onActivateEngine: (choice: EngineChoice) => void | Promise<void>;
		onSetMode: (mode: SessionMode) => void;
		onToggleTutorial: () => void;
		onSetTutorialInput: (value: string) => void;
		onNavigateTutorial: (url: string) => void;
		onTutorialBack: () => void;
		onTutorialForward: () => void;
		onTutorialReload: () => void;
		onCloseTutorial: () => void;
		onRunCommand: (commandId: string) => void | Promise<void>;
		onUpdateEditor: (text: string) => void;
		onSetEditorKeymap: (keymap: 'sublime' | 'emacs' | 'vim') => void;
		onSelectNode: (nodeId: string) => void;
		onNodeAction: (action: NodeAction, node: GraphNode) => void | Promise<void>;
		onToggleRelation: (rowId: string, displayClass: string, checked: boolean) => void | Promise<void>;
	};

	let {
		engineChoice,
		mode,
		activeModel,
		editorText,
		editorSaveState,
		editorKeymap,
		tutorialVisible,
		tutorialUrl,
		tutorialInput,
		tutorialCanGoBack,
		tutorialCanGoForward,
		tutorialFrameKey,
		tutorialButtonFlashing,
		sessionLabel,
		statusMessage,
		statusLevel,
		latestGraph,
		latestConcept,
		latestCheck,
		selectedNode,
		selectedNodeId,
		jobs,
		stateRelationRows,
		onActivateEngine,
		onSetMode,
		onToggleTutorial,
		onSetTutorialInput,
		onNavigateTutorial,
		onTutorialBack,
		onTutorialForward,
		onTutorialReload,
		onCloseTutorial,
		onRunCommand,
		onUpdateEditor,
		onSetEditorKeymap,
		onSelectNode,
		onNodeAction,
		onToggleRelation
	}: Props = $props();
</script>

<main class="workspace-shell">
	<Menubar
		{engineChoice}
		{mode}
		{tutorialVisible}
		{tutorialButtonFlashing}
		{onActivateEngine}
		{onSetMode}
		{onToggleTutorial}
		{onRunCommand}
	/>

	<div class="workbench-body">
		<SheetArea
			{activeModel}
			{editorText}
			{editorSaveState}
			{editorKeymap}
			{sessionLabel}
			{statusMessage}
			{statusLevel}
			{latestGraph}
			{latestConcept}
			{latestCheck}
			{selectedNode}
			{selectedNodeId}
			{jobs}
			{stateRelationRows}
			{onRunCommand}
			{onUpdateEditor}
			{onSetEditorKeymap}
			{onSelectNode}
			{onNodeAction}
			{onToggleRelation}
		/>

		{#if tutorialVisible}
			<div class="tutorial-divider"></div>
		{/if}
		<TutorialPane
			visible={tutorialVisible}
			url={tutorialUrl}
			input={tutorialInput}
			canGoBack={tutorialCanGoBack}
			canGoForward={tutorialCanGoForward}
			frameKey={tutorialFrameKey}
			onSetInput={onSetTutorialInput}
			onNavigate={onNavigateTutorial}
			onBack={onTutorialBack}
			onForward={onTutorialForward}
			onReload={onTutorialReload}
			onClose={onCloseTutorial}
		/>
	</div>
	<DialogHost />
	<ContextMenuHost />
	<ToastHost />
	<SessionOverlayHost />
	<FileInputHost />
</main>
