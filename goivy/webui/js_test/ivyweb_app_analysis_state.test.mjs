import { describe, expect, it, vi } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import { FakeAPI, FakeControls, FakeGraph, makePersist } from './helpers/fakes.mjs';

function makeGraph(elements) {
  const graph = {
    updateCalls: [],
    cy: { json: vi.fn(() => ({ elements })) },
    update(elementsArg, positionsArg) {
      this.updateCalls.push([elementsArg, positionsArg]);
    },
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
});
