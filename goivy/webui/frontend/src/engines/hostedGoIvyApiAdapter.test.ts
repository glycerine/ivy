import { describe, expect, it, vi } from 'vitest';
import { IvyHttpClient } from '../api/ivyHttpClient.ts';
import { HostedGoIvyApiAdapter } from './hostedGoIvyApiAdapter.ts';

function jsonResponse(data, init = {}) {
  return new Response(JSON.stringify(data), {
    status: 200,
    headers: { 'content-type': 'application/json' },
    ...init,
  });
}

describe('HostedGoIvyApiAdapter', () => {
  it('creates and scopes a hosted Go session', async () => {
    const fetchImpl = vi.fn(async (url, options) => {
      expect(url).toBe('/api/session/new');
      expect(options.method).toBe('POST');
      return jsonResponse({ session_id: 's123' });
    });
    const api = new HostedGoIvyApiAdapter({ client: new IvyHttpClient({ fetchImpl }) });

    await expect(api.createSession()).resolves.toBe('s123');
    expect(api.sessionId).toBe('s123');
  });

  it('loads model content as multipart form data', async () => {
    const fetchImpl = vi.fn(async () => jsonResponse({ ok: true }));
    const api = new HostedGoIvyApiAdapter({ client: new IvyHttpClient({ fetchImpl }) });
    api.sessionId = 'abc';

    await api.loadModel({ filename: 'client.ivy', content: '#lang ivy1.7' });

    expect(fetchImpl).toHaveBeenCalledTimes(1);
    const [url, options] = fetchImpl.mock.calls[0];
    expect(url).toBe('/api/session/abc/load');
    expect(options.method).toBe('POST');
    expect(options.body).toBeInstanceOf(FormData);
  });

  it('sends selected isolate with model loads', async () => {
    const fetchImpl = vi.fn(async () => jsonResponse({ ok: true }));
    const api = new HostedGoIvyApiAdapter({ client: new IvyHttpClient({ fetchImpl }) });
    api.sessionId = 'abc';

    await api.loadModel({ filename: 'client.ivy', content: '#lang ivy1.7', isolate: 'cf_live' });

    const form = fetchImpl.mock.calls[0][1].body;
    expect(form.get('isolate')).toBe('cf_live');
  });

  it('keeps graph calls behind snapshot requests', async () => {
    const fetchImpl = vi.fn(async (url) => jsonResponse({ url }));
    const api = new HostedGoIvyApiAdapter({ client: new IvyHttpClient({ fetchImpl }) });
    api.sessionId = 'abc';

    await api.getARG({ sheetId: 'trace-1' });
    await api.getConceptGraph('state_2', 'trace-1');
    await api.getMenus({ sheetId: 'sheet-2', uiMode: 'reachability' });

    expect(fetchImpl.mock.calls[0][0]).toBe('/api/session/abc/arg?sheet=trace-1');
    expect(fetchImpl.mock.calls[1][0]).toBe('/api/session/abc/concept?sheet=trace-1&node=state_2');
    expect(fetchImpl.mock.calls[2][0]).toBe('/api/session/abc/menus?sheet=sheet-2&ui_mode=reachability');
  });

  it('maps check commands to the hosted check endpoint', async () => {
    const fetchImpl = vi.fn(async () => jsonResponse({ ok: true }));
    const api = new HostedGoIvyApiAdapter({ client: new IvyHttpClient({ fetchImpl }) });
    api.sessionId = 'abc';
    const controller = new AbortController();

    await api.runCheck('bounded', { bound: 7, trace: true }, { signal: controller.signal });

    const [url, options] = fetchImpl.mock.calls[0];
    expect(url).toBe('/api/session/abc/check');
    expect(JSON.parse(options.body)).toEqual({ bound: 7, trace: true, mode: 'bounded' });
    expect(options.signal).toBe(controller.signal);
  });

  it('maps command intents to current webui endpoints', async () => {
    const fetchImpl = vi.fn(async () => jsonResponse({ ok: true }));
    const api = new HostedGoIvyApiAdapter({ client: new IvyHttpClient({ fetchImpl }) });
    api.sessionId = 'abc';

    await api.splitConcept('client', 'server');
    await api.argNodeAction('state_1', 'view_source', { target: 'state_2' });
    await api.executeAction('redo', {});

    expect(fetchImpl.mock.calls[0][0]).toBe('/api/session/abc/concept/split');
    expect(JSON.parse(fetchImpl.mock.calls[0][1].body)).toEqual({ concept: 'client', split_by: 'server' });
    expect(fetchImpl.mock.calls[1][0]).toBe('/api/session/abc/arg/action');
    expect(JSON.parse(fetchImpl.mock.calls[1][1].body)).toEqual({
      node: 'state_1',
      action: 'view_source',
      args: { target: 'state_2' },
    });
    expect(fetchImpl.mock.calls[2][0]).toBe('/api/session/abc/action');
    expect(JSON.parse(fetchImpl.mock.calls[2][1].body)).toEqual({ action: 'redo', args: {} });
  });

  it('preserves save failure handling', async () => {
    const fetchImpl = vi.fn(async () => new Response('nope', {
      status: 500,
      statusText: 'Internal Server Error',
    }));
    const api = new HostedGoIvyApiAdapter({ client: new IvyHttpClient({ fetchImpl }) });
    api.sessionId = 'abc';

    await expect(api.saveSession()).rejects.toThrow('Save failed: Internal Server Error');
    expect(fetchImpl).toHaveBeenCalledWith('/api/session/abc/save', {});
  });

  it('returns an unsubscribe function for server events', () => {
    const close = vi.fn();
    const factory = vi.fn(() => ({ close }));
    const api = new HostedGoIvyApiAdapter({ eventSourceFactory: factory });
    api.sessionId = 'abc';

    const unsubscribe = api.connectEvents(vi.fn());

    expect(factory).toHaveBeenCalledWith('/api/session/abc/events');
    unsubscribe();
    expect(close).toHaveBeenCalled();
  });
});
