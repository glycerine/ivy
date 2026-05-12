import type { Id, Revision } from './ids';

export type GraphKind = 'arg' | 'concept' | 'proof' | 'event-trace';

export type NodeAction = {
	label: string;
	action: string;
	args?: Record<string, unknown>;
};

export type GraphNode = {
	id: Id;
	obj: string;
	label: string;
	classes: string[];
	shape?: 'ellipse' | 'octagon' | 'rectangle' | string;
	shortInfo?: string;
	longInfo?: unknown;
	actions?: NodeAction[];
	width?: number;
	height?: number;
	color?: string;
};

export type GraphEdge = {
	id: Id;
	obj: string;
	source: Id;
	target: Id;
	label?: string;
	classes: string[];
	shortInfo?: string;
	longInfo?: unknown;
};

export type GraphSnapshot = {
	id: Id;
	sheetId: Id;
	kind: GraphKind;
	sourceRevision: Revision;
	nodes: Record<Id, GraphNode>;
	edges: Record<Id, GraphEdge>;
	nodeOrder: Id[];
	edgeOrder: Id[];
	layout?: Record<Id, { x: number; y: number }>;
	styleRevision: Revision;
	createdAt: string;
};

export type CytoscapePayload = {
	elements?: CytoscapeElement[];
};

export type CytoscapeElement = {
	group?: string;
	data?: Record<string, unknown>;
	classes?: string;
	position?: { x?: number; y?: number };
};
