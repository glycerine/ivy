import { afterEach, describe, expect, it, vi } from 'vitest';
import { IvyHttpClient } from './ivyHttpClient.ts';

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

  it('surfaces JSON API error messages without burying them in raw response text', async () => {
    const fetchImpl = vi.fn(async () => new Response(JSON.stringify({
      error: 'compile: unknown type: time',
    }), {
      status: 400,
      headers: { 'content-type': 'application/json' },
    }));
    const client = new IvyHttpClient({ fetchImpl });

    await expect(client.request('/api/session/s1/load')).rejects.toMatchObject({
      message: 'compile: unknown type: time',
    });
  });
});
