import { describe, expect, it, vi } from 'vitest';
import { IvyHttpClient } from '../api/ivyHttpClient.js';
import { HostedGoEngine } from './hostedGoEngine.js';

function jsonResponse(data, init = {}) {
  return new Response(JSON.stringify(data), {
    status: 200,
    headers: { 'content-type': 'application/json' },
    ...init,
  });
}

describe('HostedGoEngine', () => {
  it('creates and scopes a hosted Go session', async () => {
    const fetchImpl = vi.fn(async (url, options) => {
      expect(url).toBe('/api/session/new');
      expect(options.method).toBe('POST');
      return jsonResponse({ session_id: 's123' });
    });
    const engine = new HostedGoEngine({ client: new IvyHttpClient({ fetchImpl }) });

    await expect(engine.createSession()).resolves.toBe('s123');
    expect(engine.sessionId).toBe('s123');
  });

  it('loads model content as multipart form data', async () => {
    const fetchImpl = vi.fn(async () => jsonResponse({ ok: true }));
    const engine = new HostedGoEngine({ client: new IvyHttpClient({ fetchImpl }) });
    engine.sessionId = 'abc';

    await engine.loadModel({ filename: 'client.ivy', content: '#lang ivy1.7' });

    expect(fetchImpl).toHaveBeenCalledTimes(1);
    const [url, options] = fetchImpl.mock.calls[0];
    expect(url).toBe('/api/session/abc/load');
    expect(options.method).toBe('POST');
    expect(options.body).toBeInstanceOf(FormData);
  });

  it('keeps graph calls behind the engine boundary', async () => {
    const fetchImpl = vi.fn(async (url) => jsonResponse({ url }));
    const engine = new HostedGoEngine({ client: new IvyHttpClient({ fetchImpl }) });
    engine.sessionId = 'abc';

    await engine.getArg({ sheetId: 'trace-1' });
    await engine.getConcept({ sheetId: 'trace-1', stateId: 2 });

    expect(fetchImpl.mock.calls[0][0]).toBe('/api/session/abc/arg?sheet=trace-1');
    expect(fetchImpl.mock.calls[1][0]).toBe('/api/session/abc/concept?sheet=trace-1&state=2');
  });

  it('returns an unsubscribe function for server events', () => {
    const close = vi.fn();
    const factory = vi.fn(() => ({ close }));
    const engine = new HostedGoEngine({ eventSourceFactory: factory });
    engine.sessionId = 'abc';

    const unsubscribe = engine.subscribeEvents(vi.fn(), vi.fn());

    expect(factory).toHaveBeenCalledWith('/api/session/abc/events');
    unsubscribe();
    expect(close).toHaveBeenCalled();
  });
});
