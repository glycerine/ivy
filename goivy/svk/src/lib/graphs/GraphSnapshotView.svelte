<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import type cytoscape from 'cytoscape';
	import type { GraphEdge, GraphNode, GraphSnapshot, NodeAction } from '$lib/types';
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
	const edges = $derived(snapshot ? snapshot.edgeOrder.map((id) => snapshot.edges[id]) : []);
	const selectedNode = $derived(snapshot && selectedNodeId ? snapshot.nodes[selectedNodeId] : null);
	const arrowId = $derived(`${testId}-arrowhead`);

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

	function nodeCenter(nodeId: string) {
		const index = snapshot?.nodeOrder.indexOf(nodeId) ?? -1;
		const node = snapshot?.nodes[nodeId];
		const layout = snapshot?.layout?.[nodeId];
		return {
			x: layout?.x ?? 120 + Math.max(index, 0) * 140,
			y: layout?.y ?? 110 + (Math.max(index, 0) % 2) * 72,
			node
		};
	}

	function edgeLine(edge: GraphEdge) {
		const source = nodeCenter(edge.source);
		const target = nodeCenter(edge.target);
		const radius = 58;
		const dx = target.x - source.x;
		const dy = target.y - source.y;
		const distance = Math.max(Math.hypot(dx, dy), 1);
		const unitX = dx / distance;
		const unitY = dy / distance;
		return {
			x1: source.x + unitX * radius,
			y1: source.y + unitY * radius,
			x2: target.x - unitX * radius,
			y2: target.y - unitY * radius,
			labelX: (source.x + target.x) / 2 - unitY * 24,
			labelY: (source.y + target.y) / 2 + unitX * 24
		};
	}

	function selectNode(node: GraphNode) {
		onSelect?.(node);
	}
</script>

<div class="graph-view" class:concept-graph={snapshot?.kind === 'concept'} data-testid={testId}>
	<div class="cy-container" bind:this={container} aria-hidden="true"></div>
	<svg class="graph-edge-layer" aria-hidden="true">
		<defs>
			<marker id={arrowId} viewBox="0 0 10 10" refX="9" refY="5" markerWidth="8" markerHeight="8" orient="auto">
				<path d="M 0 0 L 10 5 L 0 10 z"></path>
			</marker>
		</defs>
		{#each edges as edge (edge.id)}
			{@const line = edgeLine(edge)}
			<line x1={line.x1} y1={line.y1} x2={line.x2} y2={line.y2} marker-end={`url(#${arrowId})`}></line>
			{#if edge.label}
				<text x={line.labelX} y={line.labelY}>{edge.label}</text>
			{/if}
		{/each}
	</svg>
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
	{#if !snapshot}
		<div class="empty-graph-state">No graph loaded</div>
	{/if}
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
		opacity: 0;
		pointer-events: none;
	}

	.graph-edge-layer {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		overflow: visible;
		pointer-events: none;
		z-index: 1;
	}

	.graph-edge-layer line {
		stroke: #8d8d8d;
		stroke-width: 9;
	}

	.graph-edge-layer path {
		fill: #8d8d8d;
	}

	.graph-edge-layer text {
		fill: #c7c7cf;
		font-size: 1.35rem;
		font-family:
			Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
		text-anchor: middle;
	}

	.graph-hit-layer {
		position: absolute;
		inset: 0;
		pointer-events: none;
		z-index: 2;
	}

	.graph-node-hit {
		position: absolute;
		transform: translate(-50%, -50%);
		border: 0;
		background: #858585;
		color: #ffffff;
		border-radius: 999px;
		min-width: 112px;
		min-height: 112px;
		padding: 0 12px;
		cursor: pointer;
		pointer-events: auto;
		font-size: 1.8rem;
		box-shadow: none;
	}

	.graph-node-hit.selected {
		outline: 3px solid #d7d7d7;
		background: #858585;
	}

	.concept-graph .graph-node-hit {
		border-radius: 0;
		clip-path: polygon(28% 0, 72% 0, 100% 28%, 100% 72%, 72% 100%, 28% 100%, 0 72%, 0 28%);
		background: #f8f8f6;
		color: #111111;
		box-shadow: inset 0 0 0 8px #001eff;
	}

	.concept-graph .graph-node-hit.selected {
		outline: 3px solid #ffffff;
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
		z-index: 3;
	}

	.empty-graph-state {
		position: absolute;
		inset: 0;
		display: grid;
		place-items: center;
		color: #777a84;
		font-size: 0.9rem;
		z-index: 2;
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
