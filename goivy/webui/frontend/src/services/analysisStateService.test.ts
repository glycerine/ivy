import { describe, expect, it, vi } from 'vitest';
import {
  analysisStateLimits,
  buildAnalysisState,
  loadAnalysisStateFile,
  loadAnalysisStateObject,
  removeAnalysisStateExtraSheets,
  saveAnalysisState,
  validateAnalysisStateObject,
} from './analysisStateService.ts';

function makeGraph(elements = []) {
  return {
    update: vi.fn(),
    cy: {
      json: vi.fn(() => ({ elements })),
    },
  };
}

describe('analysisStateService', () => {
  it('builds a serializable analysis state from explicit app dependencies', () => {
    const app = {
      sheets: {
        'sheet-1': {
          type: 'analysis',
          selectedArgNode: 'n0',
          argGraph: { cy: { json: () => ({ elements: ['arg'] }) } },
          conceptGraph: { cy: { json: () => ({ elements: ['concept'] }) } },
        },
        'events-1': {
          type: 'events',
          events: [{ text: 'root(a)', address: '0' }],
          patterns: ['root($x)'],
          selectedEventAddress: '0',
        },
      },
      tabLabelForSheet: vi.fn((sheetId) => (sheetId === 'events-1' ? 'Trace' : 'ARG')),
      graphElementsSnapshot: (graph) => graph.cy.json().elements,
      _persistedFileName: 'client.ivy',
      _persistedFilePath: '/tmp/client.ivy',
      _editorContent: () => 'ivy',
      getMode: () => 'pdr',
      activeSheetId: 'events-1',
      selectedArgNode: 'n0',
    };
    const persist = {
      _getToggles: vi.fn(() => { return { 'link|all_to_all': true }; }),
    };

    expect(buildAnalysisState(app, persist)).toMatchObject({
      analysis_state_format: 'ivyweb-json',
      fileName: 'client.ivy',
      fileContent: 'ivy',
      activeSheetId: 'events-1',
      sheets: [
        {
          id: 'sheet-1',
          arg: { elements: ['arg'] },
          concept: { elements: ['concept'] },
        },
        {
          id: 'events-1',
          type: 'events',
          label: 'Trace',
          events: [{ text: 'root(a)', address: '0' }],
          patterns: ['root($x)'],
          selectedEventAddress: '0',
        },
      ],
    });
  });

  it('validates sheet shape and limits', () => {
    const app = {
      analysisStateLimits,
      isValidSheetId: (id) => /^[A-Za-z]/.test(id),
    };

    expect(() => validateAnalysisStateObject(app, {
      analysis_state_format: 'ivyweb-json',
      activeSheetId: 'sheet-1',
      sheets: [{ id: 'sheet-1', type: 'analysis', arg: { elements: [] }, concept: { elements: [] } }],
    })).not.toThrow();

    expect(() => validateAnalysisStateObject(app, {
      analysis_state_format: 'ivyweb-json',
      sheets: [{ id: '1-bad', type: 'analysis' }],
    })).toThrow('invalid sheet id');

    expect(() => validateAnalysisStateObject(app, {
      analysis_state_format: 'ivyweb-json',
      sheets: Array.from({ length: 101 }, (_, i) => ({ id: `events-${i}`, type: 'events', events: [], patterns: [] })),
    })).toThrow('too many sheets');

    expect(() => validateAnalysisStateObject(app, {
      analysis_state_format: 'ivyweb-json',
      sheets: [{ id: 'events-1', type: 'events', events: [{ text: 'root(a)', address: '0"]' }], patterns: [] }],
    })).toThrow('invalid event address');
  });

  it('removes every restored sheet except the root analysis sheet', () => {
    const app = {
      sheets: { 'sheet-1': {}, 'sheet-2': {}, events: {} },
      removeSheet: vi.fn(),
    };

    removeAnalysisStateExtraSheets(app);

    expect(app.removeSheet).toHaveBeenCalledWith('sheet-2');
    expect(app.removeSheet).toHaveBeenCalledWith('events');
  });

  it('saves analysis state through browser download fallback', async () => {
    const state = { analysis_state_format: 'ivyweb-json', sheets: [] };
    const app = {
      _persistedFileName: 'client.ivy',
      buildAnalysisState: vi.fn(() => state),
      downloadTextFile: vi.fn(),
      controls: { setStatus: vi.fn() },
    };

    await expect(saveAnalysisState(app, { win: {} })).resolves.toBe(state);

    expect(app.downloadTextFile).toHaveBeenCalledWith(
      'client.ivyweb.json',
      `${JSON.stringify(state, null, 2)}\n`,
      'application/json',
    );
    expect(app.controls.setStatus).toHaveBeenCalledWith('Analysis state downloaded: client.ivyweb.json', 'success');
  });

  it('rejects oversized analysis state files before reading them', async () => {
    const app = {
      analysisStateLimits,
      readFileText: vi.fn(),
      loadAnalysisStateObject: vi.fn(),
    };
    const file = {
      size: 26 * 1024 * 1024,
      text: vi.fn(async () => '{"analysis_state_format":"ivyweb-json"}'),
    };

    await expect(loadAnalysisStateFile(app, file)).rejects.toThrow('analysis state file too large');
    expect(app.readFileText).not.toHaveBeenCalled();
  });

  it('restores visual analysis state into root, event, and extra analysis sheets', async () => {
    const extraConcept = makeGraph();
    const app = {
      sheets: { 'sheet-1': {} },
      argGraph: makeGraph(),
      conceptGraph: makeGraph(),
      validateAnalysisStateObject: vi.fn(),
      setEditorContent: vi.fn(),
      setMode: vi.fn(),
      removeAnalysisStateExtraSheets: vi.fn(),
      openEventTraceSheet: vi.fn((label, data, sheetId) => {
        app.sheets[sheetId] = { id: sheetId, type: 'events' };
      }),
      openARGSheet: vi.fn((label, arg, sheetId) => {
        app.sheets[sheetId] = { id: sheetId, type: 'analysis', conceptGraph: extraConcept };
      }),
      uiDataStore: {
        applyArgSnapshot: vi.fn(),
        applyConceptSnapshot: vi.fn(),
        setSelectedArgNode: vi.fn(),
      },
      setVisualOnlySheet: vi.fn((sheetId, visualOnly) => {
        app.sheets[sheetId].visualOnly = visualOnly;
      }),
      sheetExists: vi.fn((sheetId) => sheetId === 'events-1'),
      switchSheet: vi.fn(),
      controls: { setStatus: vi.fn() },
      api: { reloadContent: vi.fn(async () => ({ status: 'ok' })) },
    };
    const persist = {
      _setToggles: vi.fn(),
      setFileName: vi.fn(),
    };
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
      toggles: { 'link|all_to_all': true },
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
          id: 'sheet-2',
          type: 'analysis',
          label: 'Saved Sheet',
          selectedArgNode: 'state_2',
          arg: { elements: [{ data: { id: 'n2', obj: 'state_2' } }] },
          concept: { elements: [{ data: { id: 'c2', obj: 'client' } }] },
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

    await loadAnalysisStateObject(app, state, persist);

    expect(app.api.reloadContent).toHaveBeenCalledWith('ivy source', 'client.ivy');
    expect(app.setEditorContent).toHaveBeenCalledWith('ivy source');
    expect(app.setMode).toHaveBeenCalledWith('bounded');
    expect(app.uiDataStore.applyArgSnapshot).toHaveBeenCalledWith('sheet-1', state.sheets[0].arg);
    expect(app.uiDataStore.applyConceptSnapshot).toHaveBeenCalledWith('sheet-1', state.sheets[0].concept);
    expect(app.openARGSheet).toHaveBeenCalledWith('Saved Sheet', state.sheets[1].arg, 'sheet-2');
    expect(app.uiDataStore.applyConceptSnapshot).toHaveBeenCalledWith('sheet-2', state.sheets[1].concept);
    expect(app.openEventTraceSheet).toHaveBeenCalledWith('Trace', {
      sheet_id: 'events-1',
      events: state.sheets[2].events,
      patterns: state.sheets[2].patterns,
      selected_address: null,
    }, 'events-1');
    expect(app.sheets['sheet-1'].visualOnly).toBe(true);
    expect(app.sheets['sheet-2'].visualOnly).toBe(true);
    expect(app.sheets['events-1'].visualOnly).toBe(true);
    expect(persist._setToggles).toHaveBeenCalledWith({ 'link|all_to_all': true });
    expect(app.switchSheet).toHaveBeenCalledWith('events-1');
    expect(persist.setFileName).toHaveBeenCalledWith('client.ivy', '/tmp/client.ivy');
  });

  it('validates analysis state before mutating existing sheets or editor content', async () => {
    const app = {
      validateAnalysisStateObject: vi.fn(() => {
        throw new Error('invalid sheet id: bad"sheet');
      }),
      setEditorContent: vi.fn(),
      api: { reloadContent: vi.fn() },
    };

    await expect(loadAnalysisStateObject(app, { analysis_state_format: 'ivyweb-json' }, null)).rejects.toThrow('invalid sheet id');
    expect(app.setEditorContent).not.toHaveBeenCalled();
    expect(app.api.reloadContent).not.toHaveBeenCalled();
  });
});
