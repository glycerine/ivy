import { beforeEach, describe, expect, it, vi } from 'vitest';
import { LegacyApiAdapter } from './legacyApiAdapter.js';

function makeEngine() {
  return {
    sessionId: 's1',
    eventSource: null,
    eventSourceFactory: vi.fn(() => ({ close: vi.fn() })),
    sessionPath: (suffix) => `/api/session/s1${suffix}`,
    createSession: vi.fn(async () => 's1'),
    loadModel: vi.fn(async () => ({ ok: true })),
    getArg: vi.fn(async () => ({ nodes: [] })),
    getConcept: vi.fn(async () => ({ nodes: [] })),
    getMenus: vi.fn(async () => ({ arg: [], concept: [] })),
    check: vi.fn(async () => ({ result: 'ok' })),
    runAction: vi.fn(async () => ({ status: 'ok' })),
    runArgAction: vi.fn(async () => ({ status: 'ok' })),
    getToggles: vi.fn(async () => ({ edges: {}, labels: {} })),
    setToggles: vi.fn(async () => ({ status: 'ok' })),
    requestSession: vi.fn(async () => ({ status: 'ok' })),
    fetchSession: vi.fn(async () => ({
      ok: true,
      blob: vi.fn(async () => new Blob(['state'])),
    })),
  };
}

describe('LegacyApiAdapter', () => {
  beforeEach(() => {
    vi.useRealTimers();
  });

  it('presents the old IvyAPI graph/action shape over the engine interface', async () => {
    const engine = makeEngine();
    const api = new LegacyApiAdapter(engine);

    await api.reloadContent('ivy source', 'client.ivy');
    await api.getConceptGraph('state_1', 'sheet-2');
    await api.executeAction('gather', { sheet_id: 'sheet-1' });
    await api.argNodeAction('state_1', 'view_source', { target: 'state_2' });
    await api.runCheck('bounded', { bound: 7 });

    expect(engine.loadModel).toHaveBeenCalledWith({ content: 'ivy source', filename: 'client.ivy' });
    expect(engine.getConcept).toHaveBeenCalledWith({ nodeId: 'state_1', sheetId: 'sheet-2' });
    expect(engine.runAction).toHaveBeenCalledWith({ action: 'gather', args: { sheet_id: 'sheet-1' } });
    expect(engine.runArgAction).toHaveBeenCalledWith({
      node: 'state_1',
      action: 'view_source',
      args: { target: 'state_2' },
    });
    expect(engine.check).toHaveBeenCalledWith({ mode: 'bounded', bound: 7 });
  });

  it('keeps concept mutation endpoints compatible with the old IvyAPI', async () => {
    const engine = makeEngine();
    const api = new LegacyApiAdapter(engine);

    await api.splitConcept('client', 'server');
    await api.materializeEdge('link', 'client', 'server', false);

    expect(engine.requestSession).toHaveBeenNthCalledWith(1, '/concept/split', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ concept: 'client', split_by: 'server' }),
    });
    expect(engine.requestSession).toHaveBeenNthCalledWith(2, '/concept/materialize', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        relation: 'link',
        source: 'client',
        target: 'server',
        type: 'edge',
        positive: false,
      }),
    });
  });

  it('preserves old IvyAPI save failure handling', async () => {
    const engine = makeEngine();
    engine.fetchSession = vi.fn(async () => ({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
      blob: vi.fn(),
    }));
    const api = new LegacyApiAdapter(engine);

    await expect(api.saveSession()).rejects.toThrow('Save failed: Internal Server Error');
    expect(engine.fetchSession).toHaveBeenCalledWith('/save');
  });

  it('uses a generic engine event subscription when EventSource internals are unavailable', () => {
    const unsubscribe = vi.fn();
    const engine = {
      sessionId: 'wanix-session',
      eventSource: null,
      subscribeEvents: vi.fn((_onEvent, onError) => {
        onError(new Error('lost'));
        return unsubscribe;
      }),
    };
    const api = new LegacyApiAdapter(engine);
    api.onConnectionLost = vi.fn();
    const onEvent = vi.fn();

    api.connectEvents(onEvent);
    api.disconnectEvents();

    expect(engine.subscribeEvents).toHaveBeenCalledWith(onEvent, expect.any(Function));
    expect(api.onConnectionLost).toHaveBeenCalled();
    expect(unsubscribe).toHaveBeenCalled();
  });
});
