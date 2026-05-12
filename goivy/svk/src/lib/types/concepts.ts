import type { Id, Revision } from './ids';

export type EdgeDisplayClass = 'all_to_all' | 'edge_unknown' | 'none_to_none' | 'transitive';
export type NodeLabelDisplayClass = 'node_necessarily' | 'node_maybe' | 'node_necessarily_not';

export type Concept = {
	name: string;
	variables: string[];
	formula: string;
	sorts: string[];
	arity: number;
};

export type ConceptToggles = {
	edges: Record<string, Partial<Record<EdgeDisplayClass, boolean>>>;
	labels: Record<string, Partial<Record<NodeLabelDisplayClass, boolean>>>;
};

export type ConceptState = {
	id: Id;
	sessionId: Id;
	sheetId: Id;
	concepts: Record<string, Concept>;
	sortNodes: string[];
	relations?: string[];
	relationEdges: string[];
	nodeLabels: string[];
	abstractValue: Record<string, boolean>;
	toggles: ConceptToggles;
	selectedConcept?: string;
	revision: Revision;
};
