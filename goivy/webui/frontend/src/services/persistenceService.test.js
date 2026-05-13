import { afterEach, describe, expect, it, vi } from 'vitest';
import { createIvyPersist } from './persistenceService.js';

afterEach(() => {
  localStorage.clear();
  window.history.replaceState(null, '', '/');
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

describe('persistenceService', () => {
  it('saves and lists sessions from controller-owned DOM state', () => {
    document.body.innerHTML = '<select id="mode-select"><option value="bounded" selected>bounded</option></select><table><tbody id="state-checkbox-body"><tr><td><input type="checkbox" name="link" value="all_to_all" checked></td></tr></tbody></table>';
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

  it('round-trips recent session metadata through localStorage with the URL session id', () => {
    window.history.replaceState(null, '', '/#stable-session');
    const persist = createIvyPersist(window);

    persist.save(makeApp({
      api: { sessionId: 'server-session' },
      _persistedFilePath: '/tmp/ivy/client.ivy',
      _persistedFileContent: 'ivy content',
      selectedArgNode: 'node0',
      _edgeVisibility: { link: { all_to_all: true } },
      _labelVisibility: { semaphore: { node_maybe: true } },
      argGraph: null,
      conceptGraph: null,
    }));

    const state = persist.load();
    expect(state).toMatchObject({
      sessionId: 'stable-session',
      fileName: 'client.ivy',
      filePath: '/tmp/ivy/client.ivy',
      fileContent: 'ivy content',
      selectedArgNode: 'node0',
    });
    expect(window.localStorage.getItem('ivy_last_session')).toBe('stable-session');
    expect(JSON.parse(window.localStorage.getItem('ivy_sessions'))).toEqual(['stable-session']);
    expect(persist.listSessions()).toEqual([
      {
        id: 'stable-session',
        fileName: 'client.ivy',
        filePath: '/tmp/ivy/client.ivy',
        timestamp: state.timestamp,
      },
    ]);
  });

  it('restores mode, loaded file, and relation toggles into DOM state', () => {
    document.body.innerHTML = '<select id="mode-select"><option value="pdr">pdr</option></select><span id="loaded-file"></span><span id="model-editor-label"></span><table><tbody id="state-checkbox-body"><tr><td><input type="checkbox" name="link" value="edge_unknown"></td></tr></tbody></table>';
    const persist = createIvyPersist(window);

    persist._setMode('pdr');
    persist._setToggles({ 'link|edge_unknown': true });
    persist.setFileName('client.ivy', '/tmp/client.ivy');

    expect(document.getElementById('mode-select').value).toBe('pdr');
    expect(document.querySelector('input[name="link"][value="edge_unknown"]').checked).toBe(true);
    expect(document.getElementById('loaded-file').textContent).toBe('/tmp/client.ivy');
    expect(document.getElementById('model-editor-label').textContent).toBe('/tmp/client.ivy');
  });

  it('keeps URL session and path truncation behavior compatible with the old runtime', () => {
    const persist = createIvyPersist(window);

    persist.setSessionIdInURL('sess-2');

    expect(persist.getSessionIdFromURL()).toBe('sess-2');
    expect(persist.truncatePath('/very/long/parent/client.ivy', 12)).toBe('...ng/parent');
  });
});
