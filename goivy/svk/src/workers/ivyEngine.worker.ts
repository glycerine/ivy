import { normalizeCheckResult, normalizeConceptPayload, normalizeGraphPayload } from '$lib/normalizers';
import { stableId } from '$lib/normalizers/primitives';
import type {
	CommandIntent,
	EngineSession,
	ModelDocument,
	SnapshotBundle,
	VerificationJob,
	VerificationJobKind
} from '$lib/types';
import { parseWorkerRequest, type WorkerResponse } from '$lib/engines/workerProtocol';

type GoRuntime = {
	importObject: WebAssembly.Imports;
	run(instance: WebAssembly.Instance): Promise<void>;
};

type WasmDispatchResponse = {
	type: string;
	requestId?: string;
	error?: string;
	value?: unknown;
};

declare global {
	var Go: new () => GoRuntime;
	var goivyWebEngineDispatch: ((request: string) => string) | undefined;
}

let initialized = false;
let assetBaseUrl = '/wasm/';
let wasmReady: Promise<void> | null = null;
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
				await loadWebEngineWasm();
				initialized = true;
				post({ type: 'ready', requestId: request.requestId });
				break;
			case 'new-session': {
				requireInit();
				const wasmSession = await callWasm<{ id: string }>({
					type: 'new-session',
					requestId: request.requestId,
					projectId: request.projectId
				});
				const timestamp = new Date().toISOString();
				const session: EngineSession = {
					id: wasmSession.id,
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
				post({ type: 'event', sessionId: session.id, event: { type: 'job-created', job } });
				post({ type: 'event', sessionId: session.id, event: { type: 'job-progress', jobId: job.id, progress: { phase: 'browser-wasm-load' } } });
				await callWasm({
					type: 'load-model',
					requestId: request.requestId,
					sessionId: request.sessionId,
					model: request.model
				});
				await emitSnapshot(session.id, request.model.engineRevision);
				post({ type: 'event', sessionId: session.id, event: { type: 'job-succeeded', jobId: job.id, result: { ok: true } } });
				break;
			}
			case 'run-command': {
				requireInit();
				const session = requireSession(request.intent.sessionId);
				const job = createJobFromIntent(session, request.intent);
				post({ type: 'job', requestId: request.requestId, job });
				post({ type: 'event', sessionId: session.id, event: { type: 'job-created', job } });
				post({ type: 'event', sessionId: session.id, event: { type: 'job-progress', jobId: job.id, progress: { phase: 'browser-wasm-command' } } });
				const value = await callWasm({
					type: 'run-command',
					requestId: request.requestId,
					sessionId: request.intent.sessionId,
					intent: request.intent
				});
				if (request.intent.commandId.startsWith('check.')) {
					post({
						type: 'event',
						sessionId: session.id,
						event: {
							type: 'check-updated',
							result: normalizeCheckResult(value, {
								jobId: job.id,
								sessionId: session.id,
								mode: checkMode(request.intent.commandId)
							})
						}
					});
				}
				await emitSnapshot(session.id, job.modelRevision);
				post({ type: 'event', sessionId: session.id, event: { type: 'job-succeeded', jobId: job.id, result: { ok: true } } });
				break;
			}
			case 'get-snapshot': {
				requireInit();
				const session = requireSession(request.sessionId);
				const bundle = await getSnapshot(session.id, session.currentModelRevision ?? 0);
				post({ type: 'snapshot', requestId: request.requestId, bundle });
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

async function loadWebEngineWasm() {
	wasmReady ??= (async () => {
		const [{ createSmtZ3Imports }] = await Promise.all([
			import('./smtZ3Imports.js') as Promise<{
				createSmtZ3Imports(options: { z3: unknown; getGoMemory: () => WebAssembly.Memory | null }): WebAssembly.Imports[string];
			}>
		]);
		const z3 = await loadZ3();
		await import(/* @vite-ignore */ `${assetBaseUrl}wasm_exec-go1.25.6.js`);
		const go = new globalThis.Go();
		const importObject = {
			...go.importObject,
			smt_z3: createSmtZ3Imports({ z3, getGoMemory: () => null })
		};
		const wasmUrl = `${assetBaseUrl}goivy-webengine.wasm`;
		const result = await WebAssembly.instantiateStreaming(fetch(wasmUrl), importObject);
		void go.run(result.instance);
		if (typeof globalThis.goivyWebEngineDispatch !== 'function') {
			throw new Error('goivy webengine wasm did not register its dispatcher');
		}
	})();
	return wasmReady;
}

async function loadZ3(): Promise<unknown> {
	const source = await (await fetch(`${assetBaseUrl}z3-471-api.js`)).text();
	const initZ3 = new Function(`${source}; return initZ3;`)() as (options: {
		locateFile(path: string): string;
	}) => Promise<unknown>;
	return initZ3({
		locateFile(path: string) {
			if (path.endsWith('.wasm')) {
				return `${assetBaseUrl}z3-471-api.wasm`;
			}
			return `${assetBaseUrl}${path}`;
		}
	});
}

async function callWasm<T = unknown>(request: Record<string, unknown>): Promise<T> {
	await loadWebEngineWasm();
	const dispatch = globalThis.goivyWebEngineDispatch;
	if (!dispatch) {
		throw new Error('goivy webengine wasm dispatcher is unavailable');
	}
	const response = JSON.parse(dispatch(JSON.stringify(request))) as WasmDispatchResponse;
	if (response.type === 'error') {
		throw new Error(response.error ?? 'goivy webengine wasm failed');
	}
	return response.value as T;
}

async function emitSnapshot(sessionId: string, revision: number) {
	const bundle = await getSnapshot(sessionId, revision);
	for (const graph of bundle.graphs ?? []) {
		post({ type: 'event', sessionId, event: { type: 'graph-updated', snapshot: graph } });
	}
	for (const concept of bundle.concepts ?? []) {
		post({ type: 'event', sessionId, event: { type: 'concept-updated', concept } });
		post({ type: 'event', sessionId, event: { type: 'toggles-updated', sheetId: concept.sheetId, toggles: concept.toggles } });
	}
}

async function getSnapshot(sessionId: string, revision: number): Promise<SnapshotBundle> {
	const value = await callWasm<{ arg?: unknown; concept?: unknown }>({
		type: 'get-snapshot',
		requestId: nextId('snapshot-request'),
		sessionId
	});
	return {
		graphs: value.arg
			? [
					normalizeGraphPayload(value.arg as never, {
						id: stableId('graph', sessionId, 'arg', revision),
						sheetId: 'sheet-1',
						kind: 'arg',
						sourceRevision: revision
					})
				]
			: [],
		concepts: value.concept ? [normalizeConceptPayload(value.concept, { sessionId, sheetId: 'sheet-1', revision })] : []
	};
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

function checkMode(commandId: string) {
	switch (commandId) {
		case 'check.induction':
			return 'induction';
		case 'check.bounded':
			return 'bounded';
		case 'check.pdr':
			return 'pdr';
		case 'check.concrete':
			return 'concrete';
		default:
			return 'pdr';
	}
}

function nextId(prefix: string) {
	sequence += 1;
	return `${prefix}-${sequence}`;
}

function post(response: WorkerResponse) {
	self.postMessage(JSON.stringify(response));
}
