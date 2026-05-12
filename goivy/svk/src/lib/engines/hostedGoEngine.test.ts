import { describe, expect, it } from 'vitest';
import { HostedGoEngine } from './hostedGoEngine';
import type { EngineEvent, EngineSession, ModelDocument, VerificationJob } from '$lib/types';

const createdAt = '2026-05-12T00:00:00.000Z';

const session: EngineSession = {
	id: 'session-1',
	kind: 'hosted-go',
	status: 'ready',
	capabilities: {
		offline: false,
		persistentJobs: true,
		cancelJob: true,
		eventStream: true,
		parallelJobs: true
	},
	projectId: 'project-1',
	createdAt,
	updatedAt: createdAt
};

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

const job: VerificationJob = {
	id: 'job-1',
	sessionId: 'session-1',
	engineId: 'session-1',
	projectId: 'project-1',
	modelId: 'model-1',
	modelRevision: 1,
	kind: 'load',
	status: 'queued',
	createdAt,
	updatedAt: createdAt
};

function jsonResponse(value: unknown, status = 200) {
	return new Response(JSON.stringify(value), {
		status,
		headers: { 'content-type': 'application/json' }
	});
}

describe('HostedGoEngine', () => {
	it('maps hosted HTTP responses to IvyEngine records', async () => {
		expect.hasAssertions();

		const requests: string[] = [];
		const engine = new HostedGoEngine({
			baseUrl: 'https://example.test',
			fetcher: async (input, init) => {
				requests.push(`${init?.method ?? 'GET'} ${String(input)}`);
				if (String(input).endsWith('/api/engine/session')) {
					return jsonResponse({ session });
				}
				if (String(input).endsWith('/load')) {
					return jsonResponse({ job });
				}
				if (String(input).endsWith('/snapshot')) {
					return jsonResponse({ arg: { elements: [] }, concept: { concepts: {} } });
				}
				return jsonResponse({});
			}
		});

		await expect(engine.newSession('project-1')).resolves.toMatchObject({ id: 'session-1' });
		await expect(engine.loadModel('session-1', model)).resolves.toMatchObject({ id: 'job-1' });
		expect(requests).toEqual([
			'POST https://example.test/api/engine/session',
			'POST https://example.test/api/engine/session/session-1/load',
			'GET https://example.test/api/engine/session/session-1/snapshot'
		]);
	});

	it('emits local lifecycle events from native HTTP operations', async () => {
		expect.hasAssertions();

		const engine = new HostedGoEngine({
			fetcher: async (input) => {
				if (String(input).endsWith('/load')) {
					return jsonResponse({ job });
				}
				return jsonResponse({ arg: { elements: [] }, concept: { concepts: {} } });
			}
		});
		const events: EngineEvent[] = [];
		const unsubscribe = engine.subscribe('session-1', (event) => events.push(event));

		await engine.loadModel('session-1', model);
		unsubscribe();

		expect(events.map((event) => event.type)).toEqual([
			'job-created',
			'job-progress',
			'graph-updated',
			'concept-updated',
			'toggles-updated',
			'job-succeeded'
		]);
	});

	it('maps hosted HTTP errors to rejected engine operations', async () => {
		expect.hasAssertions();

		const engine = new HostedGoEngine({
			fetcher: async () => jsonResponse({ error: 'unauthorized' }, 401)
		});

		await expect(engine.newSession('project-1')).rejects.toThrow('unauthorized');
	});
});
