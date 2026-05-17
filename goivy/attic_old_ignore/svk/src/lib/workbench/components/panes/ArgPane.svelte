<script lang="ts">
	import GraphSnapshotView from '$lib/graphs/GraphSnapshotView.svelte';
	import { ARG_INVARIANT_MENU } from '$lib/workbench/commands/menuDefinitions';
	import type { GraphNode, GraphSnapshot, NodeAction } from '$lib/types';

	type Props = {
		latestGraph: GraphSnapshot | null;
		selectedNodeId: string | null;
		onSelectNode: (nodeId: string) => void;
		onNodeAction: (action: NodeAction, node: GraphNode) => void | Promise<void>;
		onRunCommand?: (commandId: string) => void | Promise<void>;
	};

	let { latestGraph, selectedNodeId, onSelectNode, onNodeAction, onRunCommand }: Props = $props();
</script>

<section class="pane arg-pane graph-pane" aria-label="ARG graph">
	<div class="pane-title stacked">
		<strong>ARG (Abstract Reachability Graph)</strong>
		<span class="dropdown">
			<button type="button" class="panel-menu">Invariant</button>
			<span class="dropdown-content">
				{#each ARG_INVARIANT_MENU as item (item.id)}
					{#if item.separatorBefore}
						<span class="dropdown-sep"></span>
					{/if}
					<button type="button" onclick={() => void onRunCommand?.(item.commandId)}>{item.label}</button>
				{/each}
			</span>
		</span>
	</div>
	<div class="subtoolbar">
		<span>File</span>
		<span>Mode</span>
		<span>Action</span>
	</div>
	<GraphSnapshotView
		snapshot={latestGraph}
		{selectedNodeId}
		testId="arg-graph"
		onSelect={(node) => onSelectNode(node.id)}
		onAction={onNodeAction}
	/>
</section>
