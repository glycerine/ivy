import { afterEach, describe, expect, it, vi } from 'vitest';
import { IvyHttpClient } from './ivyHttpClient.js';

describe('IvyHttpClient', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('binds the default browser fetch receiver', async () => {
    const fetchImpl = vi.fn(function fetchWithReceiverCheck() {
      expect(this).toBe(globalThis);
      return Promise.resolve(new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { 'content-type': 'application/json' },
      }));
    });
    vi.stubGlobal('fetch', fetchImpl);
    const client = new IvyHttpClient();

    await expect(client.request('/api/session/new')).resolves.toEqual({ ok: true });
    expect(fetchImpl).toHaveBeenCalledWith('/api/session/new', {});
  });
});
