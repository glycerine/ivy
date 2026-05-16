import { describe, expect, it, vi } from 'vitest';
import { BrowserWasmIvyApiAdapter } from './browserWasmIvyApiAdapter.ts';

class FakeWorker {
  onmessage = null;
  onerror = null;
  terminated = false;
  requests = [];

  postMessage(message) {
    this.requests.push(message);
    if (message.type === 'init') {
      this.reply(message, 'ready', { ok: true });
    }
  }

  reply(request, type, value) {
    this.onmessage && this.onmessage({ data: { type, requestId: request.requestId, value } });
  }

  terminate() {
    this.terminated = true;
  }
}

function makeAdapter() {
  const workers = [];
  const api = new BrowserWasmIvyApiAdapter({
    workerFactory: () => {
      const worker = new FakeWorker();
      workers.push(worker);
      return worker as unknown as Worker;
    },
  });
  return { api, workers };
}

async function flushAsync() {
  await Promise.resolve();
  await Promise.resolve();
  await Promise.resolve();
}

describe('BrowserWasmIvyApiAdapter', () => {
  it('creates browser wasm sessions through the worker protocol', async () => {
    const { api, workers } = makeAdapter();
    const pending = api.createSession();
    await flushAsync();

    expect(workers[0].requests.map((request) => request.type)).toEqual(['init', 'new-session']);
    workers[0].reply(workers[0].requests[1], 'session', { id: 'browser-s1' });

    await expect(pending).resolves.toBe('browser-s1');
    expect(api.sessionId).toBe('browser-s1');
  });

  it('loads a complete model snapshot including isolate before commands run', async () => {
    const { api, workers } = makeAdapter();
    const session = api.createSession();
    await flushAsync();
    workers[0].reply(workers[0].requests[1], 'session', { id: 'browser-s1' });
    await session;

    const load = api.reloadContent('ivy source', 'model.ivy', { isolate: 'cf_live' });
    await flushAsync();
    const loadRequest = workers[0].requests[2];
    expect(loadRequest).toMatchObject({
      type: 'load-model',
      sessionId: 'browser-s1',
      model: {
        filename: 'model.ivy',
        text: 'ivy source',
        isolate: 'cf_live',
        engineRevision: 1,
      },
    });
    workers[0].reply(loadRequest, 'load-result', { status: 'ok' });
    await expect(load).resolves.toEqual({ status: 'ok' });
  });

  it('hard-cancels browser jobs by terminating the worker and hydrates the next worker from the last model', async () => {
    const { api, workers } = makeAdapter();
    const session = api.createSession();
    await flushAsync();
    workers[0].reply(workers[0].requests[1], 'session', { id: 'browser-s1' });
    await session;

    const load = api.reloadContent('ivy source', 'model.ivy');
    await flushAsync();
    workers[0].reply(workers[0].requests[2], 'load-result', { status: 'ok' });
    await load;

    const controller = new AbortController();
    const running = api.runCheck('induction', {}, { signal: controller.signal });
    await flushAsync();
    const commandRequest = workers[0].requests[3];
    expect(commandRequest.type).toBe('run-command');

    controller.abort();
    await expect(running).rejects.toMatchObject({ name: 'AbortError' });
    expect(workers[0].terminated).toBe(true);
    expect(api.sessionId).toBeNull();

    const nextSnapshot = api.getSnapshot({ arg: true });
    await flushAsync();
    expect(workers).toHaveLength(2);
    workers[1].reply(workers[1].requests[1], 'session', { id: 'browser-s2' });
    await flushAsync();
    expect(workers[1].requests[2]).toMatchObject({
      type: 'load-model',
      sessionId: 'browser-s2',
      model: { text: 'ivy source' },
    });
    workers[1].reply(workers[1].requests[2], 'load-result', { status: 'ok' });
    await flushAsync();
    workers[1].reply(workers[1].requests[3], 'snapshot', { arg: { elements: [] } });
    await expect(nextSnapshot).resolves.toEqual({ arg: { elements: [] } });
  });

  it('emits worker events to subscribers', async () => {
    const { api, workers } = makeAdapter();
    const onEvent = vi.fn();
    api.subscribe(onEvent);

    const session = api.createSession();
    await flushAsync();
    workers[0].reply(workers[0].requests[1], 'session', { id: 'browser-s1' });
    await session;

    workers[0].onmessage({
      data: {
        type: 'event',
        event: { type: 'job-progress', data: { message: 'working' } },
      },
    });

    expect(onEvent).toHaveBeenCalledWith({ type: 'job-progress', data: { message: 'working' } });
  });
});
