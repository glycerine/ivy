import { beforeEach, describe, expect, it, vi } from 'vitest';
import { IvyApiAdapter } from './ivyApiAdapter.ts';

class TestIvyApiAdapter extends IvyApiAdapter {
  constructor() {
    super();
    this.sessionId = 's1';
    this.calls = [];
    this.unsubscribe = vi.fn();
  }

  async loadModel(model) {
    this.calls.push(['loadModel', model]);
    return { ok: true };
  }

  async runCommand(intent) {
    this.calls.push(['runCommand', intent]);
    return { status: 'ok' };
  }

  async getSnapshot(request) {
    this.calls.push(['getSnapshot', request]);
    if (request.arg) return { arg: { nodes: [] } };
    if (request.concept) return { concept: { elements: [] } };
    if (request.menus) return { menus: { arg: [], concept: [] } };
    if (request.toggles) return { toggles: { edges: {}, labels: {} } };
    if (request.proof) return { proof: { goals: [] } };
    return {};
  }

  subscribe(onEvent, onError) {
    this.calls.push(['subscribe', onEvent, onError]);
    this.onError = onError;
    return this.unsubscribe;
  }

  async saveSession() {
    this.calls.push(['saveSession']);
    return new Blob(['state']);
  }
}

describe('IvyApiAdapter', () => {
  beforeEach(() => {
    vi.useRealTimers();
  });

  it('presents the old IvyAPI graph/action shape over neutral primitives', async () => {
    const api = new TestIvyApiAdapter();

    await api.reloadContent('ivy source', 'client.ivy');
    await api.getConceptGraph('state_1', 'sheet-2');
    await api.getMenus({ sheetId: 'sheet-2', uiMode: 'reachability' });
    await api.executeAction('gather', { sheet_id: 'sheet-1' });
    await api.argNodeAction('state_1', 'view_source', { target: 'state_2' });
    await api.runCheck('bounded', { bound: 7 });

    expect(api.calls).toEqual([
      ['loadModel', { content: 'ivy source', filename: 'client.ivy' }],
      ['getSnapshot', { concept: { nodeId: 'state_1', sheetId: 'sheet-2' } }],
      ['getSnapshot', { menus: { sheetId: 'sheet-2', uiMode: 'reachability' } }],
      ['runCommand', { commandId: 'gather', args: { sheet_id: 'sheet-1' } }],
      ['runCommand', {
        commandId: 'view_source',
        target: { kind: 'arg', nodeId: 'state_1' },
        args: { target: 'state_2' },
      }],
      ['runCommand', { commandId: 'check.bounded', args: { bound: 7, mode: 'bounded' } }],
    ]);
  });

  it('maps concept convenience methods into command intents', async () => {
    const api = new TestIvyApiAdapter();

    await api.splitConcept('client', 'server');
    await api.materializeEdge('link', 'client', 'server', false);
    await api.resetDomain();

    expect(api.calls).toEqual([
      ['runCommand', { commandId: 'concept.split', args: { concept: 'client', splitBy: 'server' } }],
      ['runCommand', {
        commandId: 'concept.materializeEdge',
        args: {
          relation: 'link',
          source: 'client',
          target: 'server',
          positive: false,
        },
      }],
      ['runCommand', { commandId: 'concept.reset', args: {} }],
    ]);
  });

  it('uses the neutral event subscription and disconnect hook', () => {
    const api = new TestIvyApiAdapter();
    api.onConnectionLost = vi.fn();
    const onEvent = vi.fn();

    const unsubscribe = api.connectEvents(onEvent);
    api.onError(new Error('lost'));
    api.disconnectEvents();

    expect(api.calls[0][0]).toBe('subscribe');
    expect(api.calls[0][1]).toBe(onEvent);
    expect(api.onConnectionLost).toHaveBeenCalled();
    expect(unsubscribe).toBe(api.unsubscribe);
    expect(api.unsubscribe).toHaveBeenCalled();
  });

  it('does not expose raw session endpoint escape hatches', () => {
    const api = new TestIvyApiAdapter();

    expect(api.requestSession).toBeUndefined();
    expect(api.fetchSession).toBeUndefined();
  });
});
