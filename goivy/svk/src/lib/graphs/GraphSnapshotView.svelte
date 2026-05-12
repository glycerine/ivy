<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import type cytoscape from 'cytoscape';
	import type { GraphNode, GraphSnapshot, NodeAction } from '$lib/types';
	import { graphSnapshotToCytoscapeElements } from './cytoscapeElements';
	import { syncCytoscapeSnapshot } from './cytoscapeSync';
	import { graphStyleFor } from './graphStyles';

	type Props = {
		snapshot: GraphSnapshot | null;
		selectedNodeId?: string | null;
		testId?: string;
		onSelect?: (node: GraphNode) => void;
		onAction?: (action: NodeAction, node: GraphNode) => void;
	};

	let {
		snapshot,
		selectedNodeId = null,
		testId = 'cytoscape-graph',
		onSelect,
		onAction
	}: Props = $props();

	let container = $state<HTMLDivElement | null>(null);
	let cy = $state<cytoscape.Core | null>(null);

	const nodes = $derived(snapshot ? snapshot.nodeOrder.map((id) => snapshot.nodes[id]) : []);
	const selectedNode = $derived(snapshot && selectedNodeId ? snapshot.nodes[selectedNodeId] : null);

	onMount(() => {
		if (!browser || !container || !snapshot) {
			return;
		}

		let destroyed = false;
		void (async () => {
			const cytoscapeFactory = (await import('cytoscape')).default;
			if (destroyed || !container || !snapshot) {
				return;
			}
			cy = cytoscapeFactory({
				container,
				elements: graphSnapshotToCytoscapeElements(snapshot),
				style: graphStyleFor(snapshot.kind),
				layout: { name: 'preset', fit: true, padding: 24 },
				wheelSensitivity: 0.2
			});
			cy.on('tap', 'node', (event: cytoscape.EventObjectNode) => {
				const node = snapshot?.nodes[event.target.id()];
				if (node) {
					onSelect?.(node);
				}
			});
		})();

		return () => {
			destroyed = true;
			cy?.destroy();
			cy = null;
		};
	});

	$effect(() => {
		if (!cy || !snapshot) {
			return;
		}
		syncCytoscapeSnapshot(cy, snapshot);
		cy.style(graphStyleFor(snapshot.kind));
	});

	$effect(() => {
		if (!cy) {
			return;
		}
		cy.elements().unselect();
		if (selectedNodeId) {
			cy.getElementById(selectedNodeId).select();
		}
	});

	function nodeStyle(node: GraphNode, index: number) {
		const layout = snapshot?.layout?.[node.id];
		const x = layout?.x ?? 120 + index * 140;
		const y = layout?.y ?? 110 + (index % 2) * 72;
		return `left:${x}px;top:${y}px`;
	}

	function selectNode(node: GraphNode) {
		onSelect?.(node);
	}
</script>

<div class="graph-view" data-testid={testId}>
	<div class="cy-container" bind:this={container} aria-hidden="true"></div>
	<div class="graph-hit-layer" aria-label="Graph nodes">
		{#each nodes as node, index (node.id)}
			<button
				type="button"
				class:selected={selectedNodeId === node.id}
				class="graph-node-hit"
				data-testid="graph-node"
				style={nodeStyle(node, index)}
				onclick={() => selectNode(node)}
			>
				{node.label}
			</button>
		{/each}
	</div>
	{#if selectedNode?.actions?.length}
		<div class="graph-actions" data-testid="graph-actions">
			{#each selectedNode.actions as action (action.label)}
				<button type="button" onclick={() => onAction?.(action, selectedNode)}>{action.label}</button>
			{/each}
		</div>
	{/if}
</div>

<style>
	.graph-view {
		position: relative;
		min-height: 0;
		height: 100%;
		overflow: hidden;
		background: #11101d;
	}

	.cy-container {
		position: absolute;
		inset: 0;
	}

	.graph-hit-layer {
		position: absolute;
		inset: 0;
		pointer-events: none;
	}

	.graph-node-hit {
		position: absolute;
		transform: translate(-50%, -50%);
		border: 0;
		background: #858585;
		color: #ffffff;
		border-radius: 999px;
		min-width: 124px;
		min-height: 124px;
		padding: 0 12px;
		cursor: pointer;
		pointer-events: auto;
		font-size: 2rem;
		box-shadow: none;
	}

	.graph-node-hit.selected {
		outline: 3px solid #d7d7d7;
		background: #858585;
	}

	.graph-actions {
		position: absolute;
		right: 10px;
		bottom: 10px;
		display: flex;
		gap: 8px;
		padding: 8px;
		border: 1px solid #393b44;
		border-radius: 4px;
		background: rgba(24, 26, 28, 0.94);
	}

	.graph-actions button {
		border: 1px solid #424550;
		background: #24262b;
		color: #d8d8df;
		border-radius: 3px;
		min-height: 30px;
		padding: 0 10px;
		cursor: pointer;
	}

	@media (max-height: 720px) {
		.graph-node-hit {
			min-width: 92px;
			min-height: 92px;
			font-size: 1.5rem;
		}
	}
</style>
