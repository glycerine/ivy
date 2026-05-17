import type { Id, Revision } from './ids';

export type ModelDocument = {
	id: Id;
	projectId: Id;
	filename: string;
	path?: string;
	text: string;
	savedTextHash?: string;
	dirty: boolean;
	parseRevision: Revision;
	engineRevision: Revision;
	createdAt: string;
	updatedAt: string;
};

export type ModelRevision = {
	id: Id;
	modelId: Id;
	projectId: Id;
	revision: Revision;
	textHash: string;
	text?: string;
	createdAt: string;
	createdBy: Id;
};
