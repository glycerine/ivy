import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createIvyPersist } from './persistenceService.ts';
import { UIDataModel } from '../models/uiDataModel.ts';
import { createUIDataModelStore } from '../models/uiDataModelStore.ts';

function makeLocalStorage() {
  const items = new Map<string, string>();
  return {
    get length() {
      return items.size;
    },
    key: vi.fn((index: number) => Array.from(items.keys())[index] || null),
    getItem: vi.fn((key: string) => (items.has(key) ? items.get(key) : null)),
    setItem: vi.fn((key: string, value: string) => {
      items.set(key, String(value));
    }),
    removeItem: vi.fn((key: string) => {
      items.delete(key);
    }),
    clear: vi.fn(() => {
      items.clear();
    }),
  };
}

beforeEach(() => {
  const storage = makeLocalStorage();
  Object.defineProperty(window, 'localStorage', { value: storage, configurable: true });
  Object.defineProperty(globalThis, 'localStorage', { value: storage, configurable: true });
});

afterEach(() => {
  window.localStorage.clear();
  window.history.replaceState(null, '', '/');
});

function makeModel(selectedArgNode = '0') {
  const uiDataModel = new UIDataModel();
  const store = createUIDataModelStore(uiDataModel);
  store.applyArgSnapshot('sheet-1', { elements: [{ data: { id: 'a' } }] });
  store.applyConceptSnapshot('sheet-1', {
    relations: ['link'],
    toggles: { edges: { link: { all_to_all: true } } },
    elements: [{ data: { id: 'c', obj: 'selected-c' } }],
  });
  store.setSelectedArgNode('sheet-1', selectedArgNode);
  store.setConceptSelections('sheet-1', [
    { kind: 'node', id: 'selected-c', obj: 'selected-c', label: 'selected-c', sourceObj: '', targetObj: '' },
  ]);
  return uiDataModel;
}

function makeGraph() {
  return {
    update: vi.fn(),
    highlightNode: vi.fn(),
    clearHighlights: vi.fn(),
  };
}

function makeApp(overrides: any = {}) {
  const uiDataModel = overrides.uiDataModel || makeModel(overrides.selectedArgNode || '0');
  return {
    api: { sessionId: 'sess-1' },
    activeSheetId: 'sheet-1',
    _persistedFileName: 'client.ivy',
    _persistedFilePath: '/tmp/client.ivy',
    _persistedFileContent: '#lang ivy1.7',
    getEditorKeymap: () => overrides.editorKeymap || 'emacs',
    buildAnalysisState: () => ({ sheets: [] }),
    ...overrides,
    uiDataModel,
  };
}

describe('persistenceService', () => {
  it('saves and lists sessions from model-owned state', () => {
    document.body.innerHTML = '<select id="mode-select"><option value="bounded" selected>bounded</option></select><table><tbody id="state-checkbox-body"><tr><td><input type="checkbox" name="link" value="all_to_all" checked></td></tr></tbody></table>';
    const persist = createIvyPersist(window);

    persist.save(makeApp());

    expect(persist.load()).toMatchObject({
      sessionId: 'sess-1',
      fileName: 'client.ivy',
      filePath: '/tmp/client.ivy',
      editorKeymap: 'emacs',
      mode: 'bounded',
      toggles: { 'link|all_to_all': true },
      conceptSelections: [
        { kind: 'node', id: 'selected-c', obj: 'selected-c', label: 'selected-c', sourceObj: '', targetObj: '' },
      ],
    });
    expect(persist.listSessions()).toEqual(expect.arrayContaining([
      expect.objectContaining({
        id: 'sess-1',
        fileName: 'client.ivy',
        filePath: '/tmp/client.ivy',
      }),
    ]));
  });

  it('clears saved localStorage session data without touching unrelated keys', () => {
    const persist = createIvyPersist(window);
    window.localStorage.setItem('ivy_sessions', JSON.stringify(['sess-1']));
    window.localStorage.setItem('ivy_last_session', 'sess-1');
    window.localStorage.setItem('ivy_sess_sess-1', '{"fileName":"client.ivy"}');
    window.localStorage.setItem('ivy_sess_orphan', '{"fileName":"old.ivy"}');
    window.localStorage.setItem('other_app_key', 'keep');

    const removed = persist.clearSavedSessions();

    expect(removed).toBe(4);
    expect(window.localStorage.getItem('ivy_sessions')).toBeNull();
    expect(window.localStorage.getItem('ivy_last_session')).toBeNull();
    expect(window.localStorage.getItem('ivy_sess_sess-1')).toBeNull();
    expect(window.localStorage.getItem('ivy_sess_orphan')).toBeNull();
    expect(window.localStorage.getItem('other_app_key')).toBe('keep');
  });

  it('round-trips recent session metadata through localStorage with the URL session id', () => {
    window.history.replaceState(null, '', '/#stable-session');
    const persist = createIvyPersist(window);

    persist.save(makeApp({
      api: { sessionId: 'server-session' },
      _persistedFilePath: '/tmp/ivy/client.ivy',
      _persistedFileContent: 'ivy content',
      selectedArgNode: 'node0',
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
    expect(JSON.parse(window.localStorage.getItem('ivy_sessions'))).toEqual(expect.arrayContaining([
      'stable-session',
      'recent-spec:/tmp/ivy/client.ivy',
    ]));
    expect(persist.listSessions()).toEqual(expect.arrayContaining([
      expect.objectContaining({
        id: 'stable-session',
        fileName: 'client.ivy',
        filePath: '/tmp/ivy/client.ivy',
        timestamp: state.timestamp,
      }),
    ]));
  });

  it('keeps recently opened specs even when one browser session loads several files', () => {
    window.history.replaceState(null, '', '/#stable-session');
    const persist = createIvyPersist(window);

    persist.save(makeApp({
      api: { sessionId: 'server-session' },
      _persistedFileName: 'client.ivy',
      _persistedFilePath: '/tmp/ivy/client.ivy',
      _persistedFileContent: 'client content',
    }));
    persist.save(makeApp({
      api: { sessionId: 'server-session' },
      _persistedFileName: 'server.ivy',
      _persistedFilePath: '/tmp/ivy/server.ivy',
      _persistedFileContent: 'server content',
    }));

    const sessions = persist.listSessions();
    expect(sessions.map((session) => session.filePath)).toContain('/tmp/ivy/client.ivy');
    expect(sessions.map((session) => session.filePath)).toContain('/tmp/ivy/server.ivy');
    expect(persist.loadSession('recent-spec:/tmp/ivy/client.ivy')).toMatchObject({
      fileName: 'client.ivy',
      fileContent: 'client content',
    });
  });

  it('restores mode, loaded file, and applies relation toggles through the backend', async () => {
    document.body.innerHTML = '<select id="mode-select"><option value="pdr">pdr</option></select><span id="loaded-file"></span><span id="model-editor-label"></span>';
    const persist = createIvyPersist(window);
    const app = {
      api: { setToggles: vi.fn() },
      refreshConceptGraph: vi.fn(),
    };

    persist._setMode('pdr');
    await persist.applyToggles(app, { 'link|edge_unknown': true });
    persist.setFileName('client.ivy', '/tmp/client.ivy');

    expect(document.getElementById('mode-select').value).toBe('pdr');
    expect(app.api.setToggles).toHaveBeenCalledWith({
      edge: 'link',
      display_class: 'edge_unknown',
      value: true,
    });
    expect(app.refreshConceptGraph).toHaveBeenCalled();
    expect(document.getElementById('loaded-file').textContent).toBe('/tmp/client.ivy');
    expect(document.getElementById('model-editor-label').textContent).toBe('/tmp/client.ivy');
  });

  it('restores checkbox state into the model and keeps column headers clickable', async () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});
    document.body.innerHTML = `
      <select id="mode-select"><option value="pdr">pdr</option></select>
      <span id="loaded-file"></span>
      <span id="model-editor-label"></span>
      <span id="state-label"></span>
      <div id="info-content"></div>
      <table id="state-checkbox-table">
        <thead>
          <tr>
            <th class="chk-col">+</th>
            <th class="chk-col">?</th>
            <th class="chk-col">-</th>
            <th class="chk-col">T</th>
            <th class="name-col">Relation</th>
          </tr>
        </thead>
        <tbody id="state-checkbox-body"></tbody>
      </table>
    `;
    const persist = createIvyPersist(window);
    const uiDataModel = new UIDataModel();
    const app = {
      uiDataModel,
      activeSheetId: 'sheet-1',
      argGraph: makeGraph(),
      conceptGraph: makeGraph(),
      sheets: {
        'sheet-1': { id: 'sheet-1', type: 'analysis', argGraph: makeGraph(), conceptGraph: makeGraph() },
      },
      api: {
        sessionId: 'server-session',
        loadFile: vi.fn(),
        getARG: vi.fn(async () => ({ elements: [] })),
        getConceptGraph: vi.fn(async () => ({
          elements: [],
          relations: ['link', 'semaphore'],
          toggles: {
            edges: {
              link: { all_to_all: false },
              semaphore: { all_to_all: false },
            },
          },
        })),
        setToggles: vi.fn(),
      },
      refreshConceptGraph: vi.fn(),
      controls: { setStatus: vi.fn() },
      setEditorContent: vi.fn(),
      _updateEditorLabel: vi.fn(),
      _applyEdgeVisibility: vi.fn(),
      setEditorKeymap: vi.fn(),
    };

    try {
      await persist.restore(app, {
        sessionId: 'persisted-session',
        fileName: 'client.ivy',
        filePath: '/tmp/client.ivy',
        fileContent: '#lang ivy1.7',
        editorKeymap: 'vim',
        toggles: {
          'link|all_to_all': true,
          'semaphore|all_to_all': true,
        },
      });
    } finally {
      warn.mockRestore();
    }

    expect(Array.from(document.querySelectorAll('input[value="all_to_all"]')).map((input: HTMLInputElement) => input.checked)).toEqual([
      true,
      true,
    ]);
    expect(app.setEditorKeymap).toHaveBeenCalledWith('vim', { save: false });
    expect(document.querySelectorAll('thead')).toHaveLength(1);

    app.api.setToggles.mockClear();
    (document.querySelector('th[data-state-toggle-class="all_to_all"]') as HTMLElement).click();
    await Promise.resolve();
    await Promise.resolve();

    expect(app.api.setToggles).toHaveBeenCalledTimes(2);
    expect(app.api.setToggles).toHaveBeenCalledWith({
      edge: 'link',
      display_class: 'all_to_all',
      value: false,
    });
    expect(app.api.setToggles).toHaveBeenCalledWith({
      edge: 'semaphore',
      display_class: 'all_to_all',
      value: false,
    });
    expect(Array.from(document.querySelectorAll('input[value="all_to_all"]')).map((input: HTMLInputElement) => input.checked)).toEqual([
      false,
      false,
    ]);
  });

  it('restores saved ARG selection only when the loaded ARG still has that node', async () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});
    document.body.innerHTML = '<span id="loaded-file"></span><span id="model-editor-label"></span><table><tbody id="state-checkbox-body"></tbody></table>';
    const persist = createIvyPersist(window);
    const makeRestoringApp = (argPayload: any) => {
      const uiDataModel = new UIDataModel();
      return {
        uiDataModel,
        activeSheetId: 'sheet-1',
        argGraph: makeGraph(),
        conceptGraph: makeGraph(),
        api: {
          sessionId: 'server-session',
          loadFile: vi.fn(),
          getARG: vi.fn(async () => argPayload),
          getConceptGraph: vi.fn(async () => ({ elements: [] })),
        },
        controls: { setStatus: vi.fn() },
        setEditorContent: vi.fn(),
        _updateEditorLabel: vi.fn(),
        _applyEdgeVisibility: vi.fn(),
      };
    };
    const savedState = {
      sessionId: 'persisted-session',
      fileName: 'client.ivy',
      filePath: '/tmp/client.ivy',
      fileContent: '#lang ivy1.7',
      selectedArgNode: '0',
    };

    try {
      const staleApp = makeRestoringApp({ elements: [] });
      await persist.restore(staleApp, savedState);
      expect(staleApp.uiDataModel.sheets['sheet-1'].selectedArgNode).toBeNull();
      expect(staleApp.argGraph.highlightNode).not.toHaveBeenCalled();

      const validApp = makeRestoringApp({ elements: [{ group: 'nodes', data: { id: 'state_0', label: '0' } }] });
      await persist.restore(validApp, savedState);
      expect(validApp.uiDataModel.sheets['sheet-1'].selectedArgNode).toBe('state_0');
      expect(validApp.argGraph.highlightNode).toHaveBeenCalledWith('state_0');
    } finally {
      warn.mockRestore();
    }
  });

  it('keeps URL session and path truncation behavior compatible with the old runtime', () => {
    const persist = createIvyPersist(window);

    persist.setSessionIdInURL('sess-2');

    expect(persist.getSessionIdFromURL()).toBe('sess-2');
    expect(persist.truncatePath('/very/long/parent/client.ivy', 12)).toBe('...ng/parent');
  });
});
