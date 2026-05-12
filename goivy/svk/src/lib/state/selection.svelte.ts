import type { GraphSnapshot, Id } from '$lib/types';

export type ActiveDetails = {
	source: 'arg' | 'concept' | 'proof' | 'event';
	id: Id;
	shortInfo?: string;
	longInfo?: unknown;
};

export type SelectionState = {
	graphSelections: Record<Id, { nodeIds: Id[]; edgeIds: Id[] }>;
	activeDetails?: ActiveDetails;
};

export function createSelectionState(initial: Partial<SelectionState> = {}) {
	const state = $state<SelectionState>({
		graphSelections: {},
		...initial
	});

	return {
		get current() {
			return state;
		},
		selectGraphNode(graph: GraphSnapshot, nodeId: Id) {
			state.graphSelections[graph.id] = { nodeIds: [nodeId], edgeIds: [] };
			const node = graph.nodes[nodeId];
			state.activeDetails = node
				? {
						source: graph.kind === 'concept' ? 'concept' : graph.kind === 'proof' ? 'proof' : 'arg',
						id: nodeId,
						shortInfo: node.shortInfo,
						longInfo: node.longInfo
					}
				: undefined;
		},
		selectGraphEdge(graph: GraphSnapshot, edgeId: Id) {
			state.graphSelections[graph.id] = { nodeIds: [], edgeIds: [edgeId] };
			const edge = graph.edges[edgeId];
			state.activeDetails = edge
				? {
						source: graph.kind === 'concept' ? 'concept' : graph.kind === 'proof' ? 'proof' : 'arg',
						id: edgeId,
						shortInfo: edge.shortInfo,
						longInfo: edge.longInfo
					}
				: undefined;
		},
		clearGraphSelection(graphId: Id) {
			delete state.graphSelections[graphId];
			state.activeDetails = undefined;
		}
	};
}
