import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createIvyPersist } from './persistenceService.ts';
import { UIDataModel } from '../models/uiDataModel.ts';
import { createUIDataModelStore } from '../models/uiDataModelStore.ts';

function makeLocalStorage() {
  const items = new Map<string, string>();
  return {
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

function makeApp(overrides: any = {}) {
  const uiDataModel = overrides.uiDataModel || makeModel(overrides.selectedArgNode || '0');
  return {
    api: { sessionId: 'sess-1' },
    activeSheetId: 'sheet-1',
    _persistedFileName: 'client.ivy',
    _persistedFilePath: '/tmp/client.ivy',
    _persistedFileContent: '#lang ivy1.7',
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
      mode: 'bounded',
      toggles: { 'link|all_to_all': true },
      conceptSelections: [
        { kind: 'node', id: 'selected-c', obj: 'selected-c', label: 'selected-c', sourceObj: '', targetObj: '' },
      ],
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

  it('keeps URL session and path truncation behavior compatible with the old runtime', () => {
    const persist = createIvyPersist(window);

    persist.setSessionIdInURL('sess-2');

    expect(persist.getSessionIdFromURL()).toBe('sess-2');
    expect(persist.truncatePath('/very/long/parent/client.ivy', 12)).toBe('...ng/parent');
  });
});
