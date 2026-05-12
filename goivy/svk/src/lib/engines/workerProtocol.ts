import type {
	CommandIntent,
	EngineEvent,
	EngineSession,
	Id,
	ModelDocument,
	VerificationJob
} from '$lib/types';

export type WorkerRequest =
	| { type: 'init'; requestId: Id; assetBaseUrl: string }
	| { type: 'new-session'; requestId: Id; projectId: Id }
	| { type: 'load-model'; requestId: Id; sessionId: Id; model: ModelDocument }
	| { type: 'run-command'; requestId: Id; intent: CommandIntent }
	| { type: 'cancel-job'; requestId: Id; jobId: Id };

export type WorkerResponse =
	| { type: 'ready'; requestId: Id }
	| { type: 'session'; requestId: Id; session: EngineSession }
	| { type: 'job'; requestId: Id; job: VerificationJob }
	| { type: 'cancelled'; requestId: Id; jobId: Id }
	| { type: 'error'; requestId?: Id; error: string }
	| { type: 'event'; sessionId: Id; event: EngineEvent };

export function serializeWorkerRequest(request: WorkerRequest): string {
	return JSON.stringify(request);
}

export function parseWorkerRequest(value: unknown): WorkerRequest {
	const decoded = typeof value === 'string' ? JSON.parse(value) : value;
	if (!isRecord(decoded) || typeof decoded.type !== 'string') {
		throw new Error('Worker request is missing type');
	}
	if (decoded.type !== 'init' && typeof decoded.requestId !== 'string') {
		throw new Error(`Worker request ${decoded.type} is missing requestId`);
	}
	return decoded as WorkerRequest;
}

export function parseWorkerResponse(value: unknown): WorkerResponse {
	const decoded = typeof value === 'string' ? JSON.parse(value) : value;
	if (!isRecord(decoded) || typeof decoded.type !== 'string') {
		throw new Error('Worker response is missing type');
	}
	return decoded as WorkerResponse;
}

export function normalizeWorkerError(value: unknown): string {
	if (value instanceof Error) {
		return value.message;
	}
	if (typeof value === 'string') {
		return value;
	}
	return 'Unknown worker error';
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null;
}
