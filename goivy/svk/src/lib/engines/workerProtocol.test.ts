import { describe, expect, it } from 'vitest';
import { BrowserWasmEngine, type WorkerLike } from './browserWasmEngine';
import {
	parseWorkerRequest,
	parseWorkerResponse,
	serializeWorkerRequest,
	type WorkerRequest,
	type WorkerResponse
} from './workerProtocol';
import type { EngineSession, ModelDocument, VerificationJob } from '$lib/types';

const createdAt = '2026-05-12T00:00:00.000Z';

class ScriptedWorker implements WorkerLike {
	private listener: ((event: MessageEvent<string>) => void) | null = null;
	readonly requests: WorkerRequest[] = [];

	constructor(private readonly failInit = false) {}

	postMessage(message: string): void {
		const request = parseWorkerRequest(message);
		this.requests.push(request);

		if (this.failInit && request.type === 'init') {
			this.emit({ type: 'error', requestId: request.requestId, error: 'z3 wasm failed' });
			return;
		}
		if (request.type === 'init') {
			this.emit({ type: 'ready', requestId: request.requestId });
			return;
		}
		if (request.type === 'new-session') {
			const session: EngineSession = {
				id: 'session-1',
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
				createdAt,
				updatedAt: createdAt
			};
			this.emit({ type: 'session', requestId: request.requestId, session });
			return;
		}
		if (request.type === 'load-model') {
			const job = jobFor(request.requestId, request.sessionId, request.model);
			this.emit({ type: 'job', requestId: request.requestId, job });
			this.emit({
				type: 'event',
				sessionId: request.sessionId,
				event: { type: 'job-succeeded', jobId: job.id, result: { ok: true } }
			});
		}
	}

	terminate(): void {}

	addEventListener(type: 'message', listener: (event: MessageEvent<string>) => void): void {
		void type;
		this.listener = listener;
	}

	removeEventListener(): void {
		this.listener = null;
	}

	private emit(response: WorkerResponse) {
		queueMicrotask(() => {
			this.listener?.({ data: JSON.stringify(response) } as MessageEvent<string>);
		});
	}
}

function jobFor(id: string, sessionId: string, model: ModelDocument): VerificationJob {
	return {
		id: `job-${id}`,
		sessionId,
		engineId: sessionId,
		projectId: model.projectId,
		modelId: model.id,
		modelRevision: model.engineRevision,
		kind: 'load',
		status: 'queued',
		createdAt,
		updatedAt: createdAt
	};
}

const model: ModelDocument = {
	id: 'model-1',
	projectId: 'project-1',
	filename: 'model.ivy',
	text: 'type t',
	dirty: false,
	parseRevision: 1,
	engineRevision: 1,
	createdAt,
	updatedAt: createdAt
};

describe('worker protocol', () => {
	it('serializes and parses worker requests and responses', () => {
		expect.hasAssertions();

		const request: WorkerRequest = { type: 'init', requestId: 'request-1', assetBaseUrl: '/wasm/' };
		const encoded = serializeWorkerRequest(request);

		expect(parseWorkerRequest(encoded)).toEqual(request);
		expect(parseWorkerResponse(JSON.stringify({ type: 'ready', requestId: 'request-1' }))).toEqual({
			type: 'ready',
			requestId: 'request-1'
		});
	});

	it('routes worker events to browser engine subscribers', async () => {
		expect.hasAssertions();

		let sequence = 0;
		const worker = new ScriptedWorker();
		const engine = new BrowserWasmEngine({
			createWorker: () => worker,
			createId: () => {
				sequence += 1;
				return `request-${sequence}`;
			}
		});
		const session = await engine.newSession('project-1');
		const events: string[] = [];
		engine.subscribe(session.id, (event) => events.push(event.type));
		const job = await engine.loadModel(session.id, model);

		expect(job.kind).toBe('load');
		expect(worker.requests.map((request) => request.type)).toEqual(['init', 'new-session', 'load-model']);
		expect(events).toEqual(['job-succeeded']);
	});

	it('rejects pending browser engine requests when the worker reports an error', async () => {
		expect.hasAssertions();

		const engine = new BrowserWasmEngine({
			createWorker: () => new ScriptedWorker(true),
			createId: () => 'request-error'
		});

		await expect(engine.init()).rejects.toThrow('z3 wasm failed');
	});
});
