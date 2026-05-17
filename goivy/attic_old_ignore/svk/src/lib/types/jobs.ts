import type { Id, Revision } from './ids';

export type VerificationJobKind =
	| 'load'
	| 'check-induction'
	| 'check-bounded'
	| 'check-pdr'
	| 'check-concrete'
	| 'check-abstract'
	| 'arg-action'
	| 'concept-action'
	| 'proof-action'
	| 'event-action';

export type VerificationJobStatus = 'queued' | 'running' | 'succeeded' | 'failed' | 'cancelled';

export type VerificationJob = {
	id: Id;
	sessionId: Id;
	engineId: Id;
	projectId: Id;
	modelId: Id;
	modelRevision: Revision;
	kind: VerificationJobKind;
	status: VerificationJobStatus;
	progress?: {
		phase: string;
		done?: number;
		total?: number;
	};
	resultId?: Id;
	error?: string;
	startedAt?: string;
	finishedAt?: string;
	createdAt: string;
	updatedAt: string;
};

export type CheckMode = 'induction' | 'bounded' | 'pdr' | 'concrete' | 'abstract';

export type CheckResult = {
	id: Id;
	jobId: Id;
	sessionId: Id;
	mode: CheckMode;
	z3Contacted: boolean;
	result: 'pass' | 'fail' | 'error';
	message: string;
	failedConjecture?: string;
	failedLabel?: string;
	usedRelations?: string[];
	counterexampleTrace?: string;
	counterexampleDetails?: string;
	graphSnapshotId?: Id;
	conceptStateId?: Id;
	createdAt: string;
};
