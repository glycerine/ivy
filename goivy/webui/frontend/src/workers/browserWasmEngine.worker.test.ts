import { afterEach, describe, expect, it, vi } from 'vitest';

function response(value, init = {}) {
  return new Response(value, { status: 200, statusText: 'OK', ...init });
}

function installWorkerHarness() {
  const pending = new Map();
  const messages = [];
  vi.stubGlobal('self', {
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
  });
  return {
    messages,
    send(request) {
      const onmessage = (globalThis as any).self.onmessage;
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

describe('browserWasmEngine worker runtime lifecycle', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.resetModules();
  });

  it('restarts the Go runtime instead of reusing a dispatcher after wasm_exec reports exit', async () => {
    const assetBaseUrl = './__fixtures__/';
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

    await expect(harness.send({
      type: 'init',
      requestId: 'init-1',
      assetBaseUrl,
      includeRoot: '/include',
    })).resolves.toEqual({ ok: true });

    await expect(harness.send({
      type: 'new-session',
      requestId: 'new-session-1',
      projectId: 'webui',
    })).resolves.toEqual({ id: 'browser-s2' });
    expect((globalThis as any).__fakeGoRunCount).toBe(2);
  });
});
