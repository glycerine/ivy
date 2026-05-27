import { afterEach, describe, expect, it, vi } from 'vitest';

function response(value, init = {}) {
  return new Response(value, { status: 200, statusText: 'OK', ...init });
}

function installWorkerHarness() {
  const pending = new Map();
  const messages = [];
  const scope = {
    onmessage: null,
    postMessage(message) {
      messages.push(message);
      const waiter = message && message.requestId ? pending.get(message.requestId) : null;
      if (!waiter) return;
      pending.delete(message.requestId);
      if (message.type === 'error') {
        waiter.reject(new Error(message.error || 'worker failed'));
      } else {
        waiter.resolve(message.value);
      }
    },
  };
  vi.stubGlobal('self', scope);
  return {
    messages,
    send(request) {
      const onmessage = scope.onmessage;
      if (typeof onmessage !== 'function') {
        return Promise.reject(new Error('worker did not install onmessage'));
      }
      return new Promise((resolve, reject) => {
        pending.set(request.requestId, { resolve, reject });
        onmessage({ data: request });
      });
    },
  };
}

function installFakeGoRuntime() {
  (globalThis as any).__fakeGoRunCount = 0;
  vi.stubGlobal('Go', class FakeGo {
    argv = [];
    env = {};
    importObject = {};

    run() {
      (globalThis as any).__fakeGoRunCount += 1;
      if ((globalThis as any).__fakeGoRunCount === 1) {
        const fs = (globalThis as any).fs;
        if (fs && typeof fs.writeSync === 'function') {
          fs.writeSync(2, new TextEncoder().encode('panic: fake Go wasm crash\nfake stack frame\n'));
        }
        (globalThis as any).goivyWebEngineDispatch = () => {
          throw new Error('Go program has already exited');
        };
        return Promise.resolve();
      }
      (globalThis as any).goivyWebEngineDispatch = (raw) => {
        const request = JSON.parse(raw);
        return JSON.stringify({
          type: 'session',
          requestId: request.requestId,
          value: { id: 'browser-s2' },
        });
      };
      return new Promise(() => {});
    }
  });
}

async function flushAsync() {
  await Promise.resolve();
  await Promise.resolve();
  await Promise.resolve();
}

describe('browserWasmEngine worker runtime lifecycle', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.resetModules();
  });

  it('restarts the Go runtime instead of reusing a dispatcher after wasm_exec reports exit', async () => {
    installFakeGoRuntime();
    const harness = installWorkerHarness();
    vi.stubGlobal('fetch', vi.fn(async (url) => {
      const text = String(url);
      if (text.endsWith('include-tree.json')) {
        return response(JSON.stringify({ root: '/include', files: [] }), {
          headers: { 'Content-Type': 'application/json' },
        });
      }
      if (text.endsWith('z3-471-api.js')) {
        return response('function initZ3() { return Promise.resolve({}); }');
      }
      if (text.endsWith('goivy-webengine.wasm')) {
        return response(new Uint8Array([0]), {
          headers: { 'Content-Type': 'application/wasm' },
        });
      }
      throw new Error(`unexpected fetch: ${text}`);
    }));
    vi.stubGlobal('WebAssembly', {
      ...WebAssembly,
      instantiateStreaming: vi.fn(async () => ({ instance: { exports: { mem: {} } } })),
      instantiate: vi.fn(async () => ({ instance: { exports: { mem: {} } } })),
    });

    await import('./browserWasmEngine.worker.js');
    (globalThis as any).self = undefined;

    await expect(harness.send({
      type: 'init',
      requestId: 'init-1',
      assetBaseUrl: '/static/wasm/',
      includeRoot: '/include',
    })).resolves.toEqual({ ok: true });
    await flushAsync();

    await expect(harness.send({
      type: 'new-session',
      requestId: 'new-session-1',
      projectId: 'webui',
    })).resolves.toEqual({ id: 'browser-s2' });
    expect((globalThis as any).__fakeGoRunCount).toBe(2);
    const crashEvent = harness.messages.find((message) => (
      message.type === 'event' && message.event?.type === 'browser_wasm_runtime_crash'
    ));
    expect(crashEvent?.event.data.message).toContain('goivy webengine wasm exited');
    expect(crashEvent?.event.data.recent_output).toContain('panic: fake Go wasm crash');
    expect(crashEvent?.event.data.recent_output).toContain('fake stack frame');
  });
});
