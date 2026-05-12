import type { CommandIntent, EngineSession, ModelDocument, VerificationJob, VerificationJobKind } from '$lib/types';
import { parseWorkerRequest, type WorkerResponse } from '$lib/engines/workerProtocol';

let initialized = false;
let assetBaseUrl = '/wasm/';
let sequence = 0;
const sessions = new Map<string, EngineSession>();

self.onmessage = (event: MessageEvent<string>) => {
	void handleMessage(event.data);
};

async function handleMessage(message: string) {
	try {
		const request = parseWorkerRequest(message);
		switch (request.type) {
			case 'init':
				assetBaseUrl = request.assetBaseUrl;
				assertStableBrowserAssets(assetBaseUrl);
				initialized = true;
				post({ type: 'ready', requestId: request.requestId });
				break;
			case 'new-session': {
				requireInit();
				const timestamp = new Date().toISOString();
				const session: EngineSession = {
					id: nextId('browser-session'),
					kind: 'browser-js-wasm',
					status: 'ready',
					capabilities: {
						offline: true,
						persistentJobs: false,
						cancelJob: true,
						eventStream: true,
						parallelJobs: false
					},
					projectId: request.projectId,
					createdAt: timestamp,
					updatedAt: timestamp
				};
				sessions.set(session.id, session);
				post({ type: 'session', requestId: request.requestId, session });
				post({ type: 'event', sessionId: session.id, event: { type: 'session-ready', session } });
				break;
			}
			case 'load-model': {
				requireInit();
				const session = requireSession(request.sessionId);
				session.currentModelId = request.model.id;
				session.currentModelRevision = request.model.engineRevision;
				const job = createJob(session, request.model, 'load');
				post({ type: 'job', requestId: request.requestId, job });
				postJobLifecycle(session.id, job);
				break;
			}
			case 'run-command': {
				requireInit();
				const session = requireSession(request.intent.sessionId);
				const job = createJobFromIntent(session, request.intent);
				post({ type: 'job', requestId: request.requestId, job });
				postJobLifecycle(session.id, job);
				break;
			}
			case 'cancel-job':
				post({ type: 'cancelled', requestId: request.requestId, jobId: request.jobId });
				break;
		}
	} catch (error) {
		post({ type: 'error', error: error instanceof Error ? error.message : 'Worker failure' });
	}
}

function assertStableBrowserAssets(baseUrl: string) {
	if (baseUrl.includes('wasip1')) {
		throw new Error('wasip1 assets are not supported; use stable GOOS=js GOARCH=wasm assets');
	}
}

function requireInit() {
	if (!initialized) {
		throw new Error('Browser WASM engine is not initialized');
	}
}

function requireSession(sessionId: string) {
	const session = sessions.get(sessionId);
	if (!session) {
		throw new Error(`Unknown browser WASM session: ${sessionId}`);
	}
	return session;
}

function createJob(session: EngineSession, model: ModelDocument, kind: VerificationJobKind): VerificationJob {
	const timestamp = new Date().toISOString();
	return {
		id: nextId('browser-job'),
		sessionId: session.id,
		engineId: session.id,
		projectId: session.projectId,
		modelId: model.id,
		modelRevision: model.engineRevision,
		kind,
		status: 'queued',
		createdAt: timestamp,
		updatedAt: timestamp
	};
}

function createJobFromIntent(session: EngineSession, intent: CommandIntent): VerificationJob {
	const timestamp = new Date().toISOString();
	return {
		id: nextId('browser-job'),
		sessionId: session.id,
		engineId: session.id,
		projectId: session.projectId,
		modelId: session.currentModelId ?? 'model-unknown',
		modelRevision: session.currentModelRevision ?? 0,
		kind: commandToJobKind(intent.commandId),
		status: 'queued',
		createdAt: timestamp,
		updatedAt: timestamp
	};
}

function postJobLifecycle(sessionId: string, job: VerificationJob) {
	post({ type: 'event', sessionId, event: { type: 'job-created', job } });
	post({ type: 'event', sessionId, event: { type: 'job-progress', jobId: job.id, progress: { phase: 'browser-wasm', done: 1, total: 1 } } });
	post({ type: 'event', sessionId, event: { type: 'job-succeeded', jobId: job.id, result: { ok: true } } });
}

function commandToJobKind(commandId: string): VerificationJobKind {
	switch (commandId) {
		case 'check.induction':
			return 'check-induction';
		case 'check.bounded':
			return 'check-bounded';
		case 'check.pdr':
			return 'check-pdr';
		case 'check.concrete':
			return 'check-concrete';
		default:
			return 'arg-action';
	}
}

function nextId(prefix: string) {
	sequence += 1;
	return `${prefix}-${sequence}`;
}

function post(response: WorkerResponse) {
	self.postMessage(JSON.stringify(response));
}
