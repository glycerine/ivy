import { afterEach, describe, expect, it, vi } from 'vitest';
import { createIvyPersist } from './legacyPersist.js';

afterEach(() => {
  localStorage.clear();
  window.history.replaceState(null, '', '/');
  delete window.__ivyVueBridge;
});

function makeApp(overrides = {}) {
  return {
    api: { sessionId: 'sess-1' },
    _persistedFileName: 'client.ivy',
    _persistedFilePath: '/tmp/client.ivy',
    _persistedFileContent: '#lang ivy1.7',
    selectedArgNode: '0',
    _edgeVisibility: { link: { all_to_all: true } },
    _labelVisibility: { link: { node_necessarily: true } },
    _persistedConceptRelations: { relations: [] },
    argGraph: { cy: { json: () => ({ elements: { nodes: [{ data: { id: 'a' } }] } }) } },
    conceptGraph: {
      cy: {
        json: () => ({ elements: { nodes: [{ data: { id: 'c' } }] } }),
        nodes: () => [{ id: () => 'selected-c' }],
      },
    },
    buildAnalysisState: () => ({ sheets: [] }),
    ...overrides,
  };
}

describe('legacyPersist', () => {
  it('saves and lists sessions through the Vue-bundled IvyPersist shim', () => {
    window.__ivyVueBridge = {
      getMode: vi.fn(() => 'bounded'),
      getStateRelationToggles: vi.fn(() => ({ 'link|all_to_all': true })),
    };
    const persist = createIvyPersist(window);

    persist.save(makeApp());

    expect(persist.load()).toMatchObject({
      sessionId: 'sess-1',
      fileName: 'client.ivy',
      filePath: '/tmp/client.ivy',
      mode: 'bounded',
      toggles: { 'link|all_to_all': true },
      selectedConceptNodes: ['selected-c'],
    });
    expect(persist.listSessions()).toEqual([
      expect.objectContaining({
        id: 'sess-1',
        fileName: 'client.ivy',
        filePath: '/tmp/client.ivy',
      }),
    ]);
  });

  it('routes restored mode, relation toggles, and loaded file into the Vue bridge', () => {
    window.__ivyVueBridge = {
      setMode: vi.fn(),
      setStateRelationToggles: vi.fn(),
      setLoadedFile: vi.fn(),
    };
    const persist = createIvyPersist(window);

    persist._setMode('pdr');
    persist._setToggles({ 'link|edge_unknown': true });
    persist.setFileName('client.ivy', '/tmp/client.ivy');

    expect(window.__ivyVueBridge.setMode).toHaveBeenCalledWith('pdr');
    expect(window.__ivyVueBridge.setStateRelationToggles).toHaveBeenCalledWith({ 'link|edge_unknown': true });
    expect(window.__ivyVueBridge.setLoadedFile).toHaveBeenCalledWith('client.ivy', '/tmp/client.ivy');
  });

  it('keeps URL session and path truncation behavior compatible with the old runtime', () => {
    const persist = createIvyPersist(window);

    persist.setSessionIdInURL('sess-2');

    expect(persist.getSessionIdFromURL()).toBe('sess-2');
    expect(persist.truncatePath('/very/long/parent/client.ivy', 12)).toBe('...ng/parent');
  });
});

