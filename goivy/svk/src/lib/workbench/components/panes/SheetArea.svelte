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
</script>

<section class="workspace-grid" aria-label="Ivy workspace">
	<TabBar />
	<ArgPane {latestGraph} {selectedNodeId} {onSelectNode} {onRunCommand} onNodeAction={onNodeAction} />
	<ConceptPane {latestConcept} {onRunCommand} />
	<StateRelationsPane stateLabel="0" rows={stateRelationRows} onToggle={onToggleRelation} />
	<EditorPane {activeModel} {editorText} keymap={editorKeymap} onUpdateEditor={onUpdateEditor} onSetKeymap={onSetEditorKeymap} />
	<DetailsPane {latestCheck} {selectedNode} {jobs} />
	<StatusBar {latestCheck} {sessionLabel} message={statusMessage} level={statusLevel} />
</section>
