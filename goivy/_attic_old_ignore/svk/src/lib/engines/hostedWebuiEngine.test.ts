import { describe, expect, it } from 'vitest';
import { HostedWebuiEngine, type WebuiEventSourceLike } from './hostedWebuiEngine';
import type { EngineEvent, ModelDocument } from '$lib/types';

const createdAt = '2026-05-12T00:00:00.000Z';

const model: ModelDocument = {
	id: 'model-1',
	projectId: 'project-1',
	filename: 'client_server_example.ivy',
	text: '#lang ivy1.7\ntype client',
	dirty: false,
	parseRevision: 1,
	engineRevision: 3,
	createdAt,
	updatedAt: createdAt
};

const argPayload = {
	elements: [
		{
			group: 'nodes',
			data: { id: '0', obj: '0', label: '0' },
			position: { x: 120, y: 110 },
			classes: 'state'
		}
	]
};

const conceptPayload = {
	concepts: {
		link: { name: 'link', variables: ['X', 'Y'], formula: 'link(X,Y)', sorts: ['client', 'server'], arity: 2 }
	},
	nodes: ['client', 'server'],
	relations: ['link(X,Y)', 'semaphore'],
	edges: ['link(X,Y)'],
	node_labels: ['=@X'],
	toggles: {
		edges: { 'link(X,Y)': { edge_unknown: true } },
		labels: { '=@X': { node_maybe: true } }
	},
	abstract_value: {}
};

class StubEventSource implements WebuiEventSourceLike {
	onmessage: ((event: MessageEvent<string>) => void) | null = null;
	onerror: ((event: Event) => void) | null = null;
	closed = false;

	close(): void {
		this.closed = true;
	}

	emit(raw: unknown) {
		this.onmessage?.({ data: JSON.stringify(raw) } as MessageEvent<string>);
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

describe('HostedWebuiEngine', () => {
	it('loads a model through the existing webui API and emits normalized snapshots', async () => {
		expect.hasAssertions();

		const requests: Array<{ url: string; method: string; body?: BodyInit | null }> = [];
		const engine = new HostedWebuiEngine({
			baseUrl: 'http://webui.test',
			createId: (prefix) => `${prefix}-1`,
			now: () => createdAt,
			createEventSource: () => new StubEventSource(),
			fetcher: async (input, init) => {
				requests.push({ url: String(input), method: init?.method ?? 'GET', body: init?.body });
				if (String(input).endsWith('/api/session/new')) {
					return jsonResponse({ session_id: 's1' });
				}
				if (String(input).endsWith('/load')) {
					expect(init?.body).toBeInstanceOf(FormData);
					return jsonResponse({ status: 'ok' });
				}
				if (String(input).endsWith('/arg')) {
					return jsonResponse(argPayload);
				}
				if (String(input).endsWith('/concept')) {
					return jsonResponse(conceptPayload);
				}
				return jsonResponse({});
			}
		});

		const session = await engine.newSession('project-1');
		const events: EngineEvent[] = [];
		engine.subscribe(session.id, (event) => events.push(event));
		const job = await engine.loadModel(session.id, model);

		expect(job).toMatchObject({ kind: 'load', modelRevision: 3 });
		expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
			'POST http://webui.test/api/session/new',
			'POST http://webui.test/api/session/s1/load',
			'GET http://webui.test/api/session/s1/arg',
			'GET http://webui.test/api/session/s1/concept'
		]);
		expect(events.some((event) => event.type === 'graph-updated')).toBe(true);
		expect(events.some((event) => event.type === 'concept-updated')).toBe(true);
		expect(events.some((event) => event.type === 'toggles-updated')).toBe(true);
	});

	it('posts induction checks and emits normalized check results', async () => {
		expect.hasAssertions();

		const postedBodies: unknown[] = [];
		const engine = new HostedWebuiEngine({
			createId: (prefix) => `${prefix}-${postedBodies.length + 1}`,
			now: () => createdAt,
			createEventSource: () => new StubEventSource(),
			fetcher: async (input, init) => {
				if (String(input).endsWith('/api/session/new')) {
					return jsonResponse({ session_id: 's1' });
				}
				if (String(input).endsWith('/check')) {
					postedBodies.push(JSON.parse(String(init?.body)));
					return jsonResponse({
						status: 'ok',
						result: 'fail',
						mode: 'induction',
						message: 'counterexample found',
						failed_conjecture: '~link(X,Y)',
						z3_contacted: true,
						trace_arg: argPayload
					});
				}
				if (String(input).endsWith('/arg')) {
					return jsonResponse(argPayload);
				}
				if (String(input).endsWith('/concept')) {
					return jsonResponse(conceptPayload);
				}
				return jsonResponse({ status: 'ok' });
			}
		});
		const session = await engine.newSession('project-1');
		const events: EngineEvent[] = [];
		engine.subscribe(session.id, (event) => events.push(event));
		await engine.loadModel(session.id, model);
		await engine.runCommand({
			id: 'intent-1',
			sessionId: session.id,
			engineId: session.id,
			commandId: 'check.induction'
		});

		expect(postedBodies).toEqual([{ mode: 'induction' }]);
		expect(events).toContainEqual({
			type: 'check-updated',
			result: expect.objectContaining({
				result: 'fail',
				message: 'counterexample found',
				z3Contacted: true,
				failedConjecture: '~link(X,Y)'
			})
		});
	});

	it('translates webui SSE lifecycle events without leaking raw event types', async () => {
		expect.hasAssertions();

		const source = new StubEventSource();
		const engine = new HostedWebuiEngine({
			createId: (prefix) => `${prefix}-1`,
			now: () => createdAt,
			createEventSource: () => source,
			fetcher: async (input) => {
				if (String(input).endsWith('/api/session/new')) {
					return jsonResponse({ session_id: 's1' });
				}
				return jsonResponse({});
			}
		});

		const session = await engine.newSession('project-1');
		const events: EngineEvent[] = [];
		const unsubscribe = engine.subscribe(session.id, (event) => events.push(event));
		source.emit({ type: 'file_loaded', data: { path: 'model.ivy' } });
		source.fail();
		unsubscribe();

		expect(events).toEqual([
			{ type: 'diagnostic', severity: 'info', message: 'Model loaded by hosted webui backend' },
			{ type: 'diagnostic', severity: 'warning', message: 'Hosted webui event stream disconnected' }
		]);
		expect(source.closed).toBe(true);
	});

	it('maps webui HTTP errors to rejected engine operations', async () => {
		expect.hasAssertions();

		const engine = new HostedWebuiEngine({
			fetcher: async () => jsonResponse({ error: 'bad session' }, 404)
		});

		await expect(engine.newSession('project-1')).rejects.toThrow('bad session');
	});
});
