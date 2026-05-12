import type { Id } from './ids';
import type { GraphKind } from './graphs';

export type CommandDescriptor = {
	id: string;
	label: string;
	icon?: string;
	enabled: boolean;
	argsSchema?: unknown;
};

export type CommandIntent = {
	id: Id;
	sessionId: Id;
	engineId: Id;
	commandId: string;
	target?: {
		kind: GraphKind;
		graphId?: Id;
		nodeId?: Id;
		edgeId?: Id;
		obj?: string;
	};
	args?: Record<string, unknown>;
};
