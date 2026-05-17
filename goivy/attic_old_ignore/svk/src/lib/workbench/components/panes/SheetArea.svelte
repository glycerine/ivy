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
	import ArgPane from './ArgPane.svelte';
	import ConceptPane from './ConceptPane.svelte';
	import DetailsPane from './DetailsPane.svelte';
	import EditorPane from './EditorPane.svelte';
	import StateRelationsPane from './StateRelationsPane.svelte';
	import TabBar from './TabBar.svelte';
	import StatusBar from '../StatusBar.svelte';

	type Props = {
		activeModel: ModelDocument;
		editorText: string;
		editorSaveState: 'idle' | 'saving' | 'saved';
		editorKeymap: 'sublime' | 'emacs' | 'vim';
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
		onRunCommand: (commandId: string) => void | Promise<void>;
		onUpdateEditor: (text: string) => void;
		onSetEditorKeymap: (keymap: 'sublime' | 'emacs' | 'vim') => void;
		onSelectNode: (nodeId: string) => void;
		onNodeAction: (action: NodeAction, node: GraphNode) => void | Promise<void>;
		onToggleRelation: (rowId: string, displayClass: string, checked: boolean) => void | Promise<void>;
	};

	let {
		activeModel,
		editorText,
		editorSaveState,
		editorKeymap,
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
		onRunCommand,
		onUpdateEditor,
		onSetEditorKeymap,
		onSelectNode,
		onNodeAction,
		onToggleRelation
	}: Props = $props();

	let gridElement: HTMLElement;
	let argWidth = $state(14);
	let conceptWidth = $state(32);
	let stateWidth = $state(16);
	let detailsHeight = $state(170);
	const editorWidth = $derived(Math.max(22, 100 - argWidth - conceptWidth - stateWidth));
	const gridStyle = $derived(
		`--arg-col: ${argWidth}%; --concept-col: ${conceptWidth}%; --state-col: ${stateWidth}%; --editor-col: ${editorWidth}%; --details-row: ${detailsHeight}px;`
	);

	function clamp(value: number, min: number, max: number) {
		return Math.min(max, Math.max(min, value));
	}

	function startColumnResize(splitter: 'arg-concept' | 'concept-state' | 'state-editor', event: PointerEvent) {
		if (!gridElement) {
			return;
		}
		event.preventDefault();
		const startX = event.clientX;
		const rect = gridElement.getBoundingClientRect();
		const startArg = argWidth;
		const startConcept = conceptWidth;
		const startState = stateWidth;
		const originalCursor = document.body.style.cursor;
		const originalSelect = document.body.style.userSelect;
		document.body.style.cursor = 'col-resize';
		document.body.style.userSelect = 'none';

		const move = (moveEvent: PointerEvent) => {
			const delta = ((moveEvent.clientX - startX) / Math.max(1, rect.width)) * 100;
			if (splitter === 'arg-concept') {
				const nextArg = clamp(startArg + delta, 8, 36);
				const transfer = nextArg - startArg;
				const nextConcept = clamp(startConcept - transfer, 18, 58);
				argWidth = startArg + (startConcept - nextConcept);
				conceptWidth = nextConcept;
				return;
			}
			if (splitter === 'concept-state') {
				const nextConcept = clamp(startConcept + delta, 18, 58);
				const transfer = nextConcept - startConcept;
				const nextState = clamp(startState - transfer, 10, 32);
				conceptWidth = startConcept + (startState - nextState);
				stateWidth = nextState;
				return;
			}
			const maxState = Math.max(10, 100 - startArg - startConcept - 22);
			stateWidth = clamp(startState + delta, 10, maxState);
		};
		const stop = () => {
			window.removeEventListener('pointermove', move);
			window.removeEventListener('pointerup', stop);
			window.removeEventListener('pointercancel', stop);
			document.body.style.cursor = originalCursor;
			document.body.style.userSelect = originalSelect;
		};
		window.addEventListener('pointermove', move);
		window.addEventListener('pointerup', stop);
		window.addEventListener('pointercancel', stop);
	}

	function startDetailsResize(event: PointerEvent) {
		if (!gridElement) {
			return;
		}
		event.preventDefault();
		const startY = event.clientY;
		const startHeight = detailsHeight;
		const rect = gridElement.getBoundingClientRect();
		const maxHeight = Math.max(120, rect.height - 160);
		const originalCursor = document.body.style.cursor;
		const originalSelect = document.body.style.userSelect;
		document.body.style.cursor = 'row-resize';
		document.body.style.userSelect = 'none';
		const move = (moveEvent: PointerEvent) => {
			detailsHeight = clamp(startHeight + (startY - moveEvent.clientY), 72, maxHeight);
		};
		const stop = () => {
			window.removeEventListener('pointermove', move);
			window.removeEventListener('pointerup', stop);
			window.removeEventListener('pointercancel', stop);
			document.body.style.cursor = originalCursor;
			document.body.style.userSelect = originalSelect;
		};
		window.addEventListener('pointermove', move);
		window.addEventListener('pointerup', stop);
		window.addEventListener('pointercancel', stop);
	}
</script>

<section class="workspace-grid" aria-label="Ivy workspace" style={gridStyle} bind:this={gridElement}>
	<TabBar />
	<ArgPane {latestGraph} {selectedNodeId} {onSelectNode} {onRunCommand} onNodeAction={onNodeAction} />
	<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
	<div
		class="splitter vertical arg-concept-splitter"
		role="separator"
		aria-label="Resize ARG and concept graph"
		tabindex="0"
		data-testid="arg-concept-splitter"
		onpointerdown={(event) => startColumnResize('arg-concept', event)}
	></div>
	<ConceptPane {latestConcept} {onRunCommand} />
	<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
	<div
		class="splitter vertical concept-state-splitter"
		role="separator"
		aria-label="Resize concept graph and state relations"
		tabindex="0"
		data-testid="concept-state-splitter"
		onpointerdown={(event) => startColumnResize('concept-state', event)}
	></div>
	<StateRelationsPane stateLabel="0" rows={stateRelationRows} onToggle={onToggleRelation} />
	<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
	<div
		class="splitter vertical state-editor-splitter"
		role="separator"
		aria-label="Resize state relations and editor"
		tabindex="0"
		data-testid="state-editor-splitter"
		onpointerdown={(event) => startColumnResize('state-editor', event)}
	></div>
	<EditorPane
		{activeModel}
		{editorText}
		saveState={editorSaveState}
		keymap={editorKeymap}
		onUpdateEditor={onUpdateEditor}
		onSetKeymap={onSetEditorKeymap}
	/>
	<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
	<div
		class="splitter horizontal details-splitter"
		role="separator"
		aria-label="Resize details panel"
		tabindex="0"
		data-testid="details-splitter"
		onpointerdown={startDetailsResize}
	></div>
	<DetailsPane {latestCheck} {selectedNode} {jobs} />
	<StatusBar {latestCheck} {sessionLabel} message={statusMessage} level={statusLevel} />
</section>
