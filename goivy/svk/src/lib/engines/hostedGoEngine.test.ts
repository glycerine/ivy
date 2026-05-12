import { describe, expect, it } from 'vitest';
import { HostedGoEngine, type EventSourceLike } from './hostedGoEngine';
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

class FakeEventSource implements EventSourceLike {
	onmessage: ((event: MessageEvent<string>) => void) | null = null;
	onerror: ((event: Event) => void) | null = null;
	closed = false;

	close(): void {
		this.closed = true;
	}

	emit(event: EngineEvent) {
		this.onmessage?.({ data: JSON.stringify(event) } as MessageEvent<string>);
	}

	fail() {
		this.onerror?.(new Event('error'));
	}
}

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
				return jsonResponse({});
			}
		});

		await expect(engine.newSession('project-1')).resolves.toMatchObject({ id: 'session-1' });
		await expect(engine.loadModel('session-1', model)).resolves.toMatchObject({ id: 'job-1' });
		expect(requests).toEqual([
			'POST https://example.test/api/engine/session',
			'POST https://example.test/api/engine/session/session-1/load'
		]);
	});

	it('routes SSE messages and reports stream disconnect diagnostics', () => {
		expect.hasAssertions();

		const source = new FakeEventSource();
		const engine = new HostedGoEngine({
			createEventSource: () => source,
			fetcher: async () => jsonResponse({})
		});
		const events: EngineEvent[] = [];
		const unsubscribe = engine.subscribe('session-1', (event) => events.push(event));

		source.emit({ type: 'job-created', job });
		source.fail();
		unsubscribe();

		expect(events).toEqual([
			{ type: 'job-created', job },
			{ type: 'diagnostic', severity: 'warning', message: 'Hosted engine event stream disconnected' }
		]);
		expect(source.closed).toBe(true);
	});

	it('maps hosted HTTP errors to rejected engine operations', async () => {
		expect.hasAssertions();

		const engine = new HostedGoEngine({
			fetcher: async () => jsonResponse({ error: 'unauthorized' }, 401)
		});

		await expect(engine.newSession('project-1')).rejects.toThrow('unauthorized');
	});
});
