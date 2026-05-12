import type { Id, Revision } from './ids';

export type TraceEvent = {
	id: Id;
	text: string;
	address: string;
	children?: TraceEvent[];
};

export type TraceEventSheet = {
	id: Id;
	projectId: Id;
	sessionId: Id;
	label: string;
	events: TraceEvent[];
	patterns: string[];
	selectedAddress?: string;
	expandedAddresses: string[];
	revision: Revision;
};
