import type { GraphEdge, GraphNode, GraphSnapshot, Id } from '$lib/types';

export type GraphElementData = Record<string, unknown> & {
	id: Id;
	obj: string;
	label?: string;
	actions?: GraphNode['actions'];
	source?: Id;
	target?: Id;
};

export type GraphElementDefinition = {
	group: 'nodes' | 'edges';
	data: GraphElementData;
	classes?: string;
	position?: { x: number; y: number };
};

export function graphSnapshotToCytoscapeElements(snapshot: GraphSnapshot): GraphElementDefinition[] {
	return [
		...snapshot.nodeOrder.map((id) => nodeToElement(snapshot.nodes[id], snapshot.layout?.[id])),
		...snapshot.edgeOrder.map((id) => edgeToElement(snapshot.edges[id]))
	];
}

export function nodeToElement(
	node: GraphNode,
	position?: { x: number; y: number }
): GraphElementDefinition {
	return {
		group: 'nodes',
		data: {
			id: node.id,
			obj: node.obj,
			label: node.label,
			actions: node.actions,
			shortInfo: node.shortInfo,
			longInfo: node.longInfo,
			shape: node.shape,
			width: node.width,
			height: node.height,
			color: node.color
		},
		classes: node.classes.join(' '),
		position
	};
}

export function edgeToElement(edge: GraphEdge): GraphElementDefinition {
	return {
		group: 'edges',
		data: {
			id: edge.id,
			obj: edge.obj,
			source: edge.source,
			target: edge.target,
			label: edge.label,
			shortInfo: edge.shortInfo,
			longInfo: edge.longInfo
		},
		classes: edge.classes.join(' ')
	};
}
