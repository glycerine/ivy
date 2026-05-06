import { afterEach, describe, expect, it, vi } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import { FakeAPI, FakeControls, FakeGraph, makePersist } from './helpers/fakes.mjs';

function makeGraph(elements) {
  const graph = {
    updateCalls: [],
    cy: { json: vi.fn(() => ({ elements })) },
    update(elementsArg, positionsArg) {
      this.updateCalls.push([elementsArg, positionsArg]);
    },
    highlightNode: vi.fn(),
    resize: vi.fn(),
  };
  return {
    ...graph,
  };
}

function installStateDom() {
  document.body.innerHTML = [
    '<select id="mode-select"><option value="pdr" selected>PDR</option><option value="bounded">Bounded</option></select>',
    '<tbody id="state-checkbox-body"></tbody>',
    '<div id="sheet-area">',
    '  <div id="tab-bar"><button class="sheet-tab active" data-sheet="sheet-1"><span>Sheet 1</span></button></div>',
    '  <div id="sheet-1" class="sheet-content active"></div>',
    '</div>',
    '<span id="loaded-file"></span>',
    '<div id="model-editor-label"></div>',
  ].join('');
}

function makeStateApp() {
  installStateDom();
  const IvyApp = loadIvyApp({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
    IvyPersist: makePersist(),
  });
  const app = new IvyApp();
  app.controls = new FakeControls();
  app.argGraph = makeGraph([{ data: { id: 'n0', obj: 'state_0' } }]);
  app.conceptGraph = makeGraph([{ data: { id: 'c0', obj: 'client' } }]);
  app.registerSheet('sheet-1', app.argGraph, app.conceptGraph);
  app.setupTabs();
  app.cmEditor = { getValue: vi.fn(() => 'ivy source'), setValue: vi.fn() };
  app._persistedFileName = 'client.ivy';
  app._persistedFilePath = '/tmp/client.ivy';
  app._persistedFileContent = 'ivy source';
  app._edgeVisibility = { link: { all_to_all: true, transitive: true } };
  app._labelVisibility = { semaphore: { node_necessarily: true } };
  return app;
}

afterEach(() => {
  window.__ivyVueBridge = undefined;
});

describe('IvyApp analysis state save/load', () => {
  it('captures root graphs, event sheets, mode, file content, and visibility state', () => {
    const app = makeStateApp();
    app.openEventTraceSheet('Trace', {
      sheet_id: 'events-1',
      events: [{ text: 'root(a)', address: '0' }],
      patterns: ['root($x)'],
    });

    const state = app.buildAnalysisState();

    expect(state.analysis_state_format).toBe('ivyweb-json');
    expect(state.fileName).toBe('client.ivy');
    expect(state.mode).toBe('pdr');
    expect(state.edgeVisibility.link.transitive).toBe(true);
    expect(state.sheets.map((sheet) => sheet.id)).toEqual(['sheet-1', 'events-1']);
    expect(state.sheets[0].arg.elements[0].data.obj).toBe('state_0');
    expect(state.sheets[1].type).toBe('events');
    expect(state.sheets[1].events[0].text).toBe('root(a)');
  });

  it('uses the Vue bridge as the source of truth for mode', () => {
    window.__ivyVueBridge = {
      getMode: vi.fn(() => 'abstract'),
      setMode: vi.fn(),
    };
    const app = makeStateApp();

    expect(app.buildAnalysisState().mode).toBe('abstract');

    app.setMode('bounded');
    expect(window.__ivyVueBridge.setMode).toHaveBeenCalledWith('bounded');
    expect(document.getElementById('mode-select').value).toBe('pdr');
  });

  it('keeps the mode select fallback for non-Vue harnesses', () => {
    const app = makeStateApp();

    app.setMode('bounded');

    expect(document.getElementById('mode-select').value).toBe('bounded');
  });

  it('does not mutate Vue-owned sheet or tab active classes during bridge tab activation', () => {
    const app = makeStateApp();
    document.getElementById('tab-bar').insertAdjacentHTML(
      'beforeend',
      '<button class="sheet-tab" data-sheet="sheet-2"><span>Sheet 2</span></button>',
    );
    document.getElementById('sheet-area').insertAdjacentHTML(
      'beforeend',
      '<div id="sheet-2" class="sheet-content"></div>',
    );
    const dynamicSheet = document.getElementById('sheet-2');
    dynamicSheet.__ivyVueRenderedSheet = true;
    app.sheets['sheet-2'] = {
      id: 'sheet-2',
      type: 'analysis',
      argGraph: { resize: vi.fn() },
      conceptGraph: { resize: vi.fn() },
      selectedArgNode: null,
    };
    window.__ivyVueBridge = {
      activateSheetTab: vi.fn(),
      setActiveGraphSheet: vi.fn(),
    };

    app.switchSheet('sheet-2');

    expect(window.__ivyVueBridge.activateSheetTab).toHaveBeenCalledWith('sheet-2');
    expect(document.querySelector('[data-sheet="sheet-2"]').classList.contains('active')).toBe(false);
    expect(document.getElementById('sheet-1').classList.contains('active')).toBe(true);
    expect(dynamicSheet.classList.contains('active')).toBe(true);
  });

  it('restores a saved visual analysis state into sheets and graph instances', async () => {
    const app = makeStateApp();
    app.api = { reloadContent: vi.fn(async () => ({ status: 'ok' })) };
    const state = {
      analysis_state_format: 'ivyweb-json',
      fileName: 'client.ivy',
      filePath: '/tmp/client.ivy',
      fileContent: 'ivy source',
      mode: 'bounded',
      activeSheetId: 'events-1',
      selectedArgNode: 'state_1',
      edgeVisibility: { link: { all_to_all: true } },
      labelVisibility: { semaphore: { node_maybe: true } },
      sheets: [
        {
          id: 'sheet-1',
          type: 'analysis',
          label: 'Sheet 1',
          selectedArgNode: 'state_1',
          arg: { elements: [{ data: { id: 'n1', obj: 'state_1' } }] },
          concept: { elements: [{ data: { id: 'c1', obj: 'server' } }] },
        },
        {
          id: 'events-1',
          type: 'events',
          label: 'Trace',
          events: [{ text: 'root(a)', address: '0' }],
          patterns: ['root($x)'],
        },
      ],
    };

    await app.loadAnalysisStateObject(state);

    expect(app.api.reloadContent).toHaveBeenCalledWith('ivy source', 'client.ivy');
    expect(app.argGraph.updateCalls).toEqual([[state.sheets[0].arg.elements, undefined]]);
    expect(app.conceptGraph.updateCalls).toEqual([[state.sheets[0].concept.elements, undefined]]);
    expect(app._edgeVisibility).toEqual(state.edgeVisibility);
    expect(app.selectedArgNode).toBe('state_1');
    expect(app.activeSheetId).toBe('events-1');
    expect(document.querySelector('#events-1 [data-event-address="0"]').textContent).toContain('root(a)');
  });

  it('restores an active Vue-owned event sheet before its DOM exists', async () => {
    const app = makeStateApp();
    app.api = { reloadContent: vi.fn(async () => ({ status: 'ok' })) };
    window.__ivyVueBridge = {
      upsertSheetTab: vi.fn(),
      upsertEventTraceSheet: vi.fn(),
      removeSheetTab: vi.fn(),
      activateSheetTab: vi.fn(),
      setActiveGraphSheet: vi.fn(),
      setLoadedFile: vi.fn(),
    };
    const state = {
      analysis_state_format: 'ivyweb-json',
      fileName: 'client.ivy',
      fileContent: 'ivy source',
      activeSheetId: 'events-1',
      sheets: [
        { id: 'sheet-1', type: 'analysis', label: 'Sheet 1' },
        { id: 'events-1', type: 'events', label: 'Trace', events: [], patterns: [] },
      ],
    };

    await app.loadAnalysisStateObject(state);

    expect(document.getElementById('events-1')).toBeNull();
    expect(app.activeSheetId).toBe('events-1');
    expect(window.__ivyVueBridge.activateSheetTab).toHaveBeenCalledWith('events-1');
  });

  it('marks restored analysis sheets visual-only and blocks backend graph actions', async () => {
    const app = makeStateApp();
    app.api = {
      reloadContent: vi.fn(async () => ({ status: 'ok' })),
      getConceptGraph: vi.fn(async () => ({ elements: [] })),
    };
    const state = {
      analysis_state_format: 'ivyweb-json',
      fileName: 'client.ivy',
      filePath: '/tmp/client.ivy',
      fileContent: 'ivy source',
      activeSheetId: 'sheet-1',
      sheets: [
        {
          id: 'sheet-1',
          type: 'analysis',
          label: 'Sheet 1',
          selectedArgNode: 'state_1',
          arg: { elements: [{ data: { id: 'n1', obj: 'state_1' } }] },
          concept: { elements: [{ data: { id: 'c1', obj: 'server' } }] },
        },
      ],
    };

    await app.loadAnalysisStateObject(state);
    await app.onArgNodeClick({ id: 'n1', obj: 'state_1', label: '1', short_info: 'state', long_info: 'saved' }, 'sheet-1');

    expect(app.sheets['sheet-1'].visualOnly).toBe(true);
    expect(app.api.getConceptGraph).not.toHaveBeenCalled();
    expect(app.controls.lastStatus.kind).toBe('warning');
    expect(app.controls.lastStatus.message).toContain('visual-only');
  });

  it('rejects oversized analysis state files before reading them', async () => {
    const app = makeStateApp();
    const file = {
      size: 26 * 1024 * 1024,
      text: vi.fn(async () => '{"analysis_state_format":"ivyweb-json"}'),
    };

    await expect(app.loadAnalysisStateFile(file)).rejects.toThrow(/too large/);
    expect(file.text).not.toHaveBeenCalled();
  });

  it('validates analysis state before mutating existing sheets', async () => {
    const app = makeStateApp();
    app.api = { reloadContent: vi.fn(async () => ({ status: 'ok' })) };
    const state = {
      analysis_state_format: 'ivyweb-json',
      fileName: 'client.ivy',
      fileContent: 'ivy source',
      sheets: [
        { id: 'events-1', type: 'events', label: 'Trace', events: [], patterns: [] },
        { id: 'bad"sheet', type: 'events', label: 'Bad', events: [], patterns: [] },
      ],
    };

    await expect(app.loadAnalysisStateObject(state)).rejects.toThrow(/invalid sheet id/);
    expect(document.getElementById('events-1')).toBeNull();
    expect(app.api.reloadContent).not.toHaveBeenCalled();
  });

  it('rejects excessive sheets and invalid event addresses', async () => {
    const app = makeStateApp();
    const tooManySheets = {
      analysis_state_format: 'ivyweb-json',
      sheets: Array.from({ length: 101 }, (_, i) => ({
        id: 'events-' + i,
        type: 'events',
        label: 'Trace',
        events: [],
        patterns: [],
      })),
    };
    await expect(app.loadAnalysisStateObject(tooManySheets)).rejects.toThrow(/too many sheets/);

    const badAddress = {
      analysis_state_format: 'ivyweb-json',
      sheets: [
        {
          id: 'events-1',
          type: 'events',
          label: 'Trace',
          events: [{ text: 'root(a)', address: '0"]' }],
          patterns: [],
        },
      ],
    };
    await expect(app.loadAnalysisStateObject(badAddress)).rejects.toThrow(/invalid event address/);
  });

  it('rejects duplicate preferred ARG sheet ids without duplicating DOM', () => {
    const app = makeStateApp();

    app.openARGSheet('First copy', { elements: [] }, 'sheet-2');
    expect(() => app.openARGSheet('Duplicate copy', { elements: [] }, 'sheet-2')).toThrow(/duplicate sheet id/);

    expect(document.querySelectorAll('#sheet-2')).toHaveLength(1);
    expect(Array.from(document.querySelectorAll('.sheet-tab')).filter(
      (tab) => tab.getAttribute('data-sheet') === 'sheet-2',
    )).toHaveLength(1);
  });
});
