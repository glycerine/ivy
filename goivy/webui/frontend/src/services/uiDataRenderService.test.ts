import { afterEach, describe, expect, it, vi } from 'vitest';
import { UIDataModel } from '../models/uiDataModel.ts';
import { applyArgSnapshot, installUIDataModelStore } from './uiDataRenderService.ts';

function makeGraph() {
  return {
    update: vi.fn(),
    highlightNode: vi.fn(),
    clearHighlights: vi.fn(),
    resize: vi.fn(),
  };
}

describe('uiDataRenderService', () => {
  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('renders ARG graph updates only after the typed model accepts a snapshot', () => {
    const argGraph = makeGraph();
    const app = {
      uiDataModel: new UIDataModel(),
      activeSheetId: 'sheet-1',
      selectedArgNode: null,
      sheets: {
        'sheet-1': { id: 'sheet-1', type: 'analysis', argGraph, conceptGraph: makeGraph() },
      },
      _applyEdgeVisibility: vi.fn(),
    };
    installUIDataModelStore(app);

    const payload = {
      elements: [{ group: 'nodes', data: { id: 'state_0', label: '0' } }],
      analysis_graph_state: { states: [{ id: 0, label: '0' }] },
    };
    app.uiDataStore.applyArgSnapshot('sheet-1', payload);

    expect(app.uiDataModel.sheets['sheet-1'].arg.analysisGraphState.states[0].id).toBe(0);
    expect(argGraph.update).toHaveBeenCalledWith([
      { group: 'nodes', data: { id: 'state_0', label: '0' } },
    ], null);
  });

  it('derives concept graph labels, edge visibility, and selection from the model', () => {
    const conceptGraph = makeGraph();
    const app = {
      uiDataModel: new UIDataModel(),
      activeSheetId: 'sheet-1',
      selectedArgNode: null,
      sheets: {
        'sheet-1': { id: 'sheet-1', type: 'analysis', argGraph: makeGraph(), conceptGraph },
      },
      _applyEdgeVisibility: vi.fn(),
    };
    installUIDataModelStore(app);

    app.uiDataStore.applyConceptSnapshot('sheet-1', {
      elements: [
        { group: 'nodes', data: { id: 'client', obj: 'client', cluster: 'client', display_label: 'client' } },
        { group: 'edges', data: { id: 'e1', obj: 'link', source: 'client', target: 'server' }, classes: 'all_to_all' },
      ],
      node_labels: ['p(X)'],
      label_sorts: { 'p(X)': 'client' },
      toggles: {
        edges: { link: { all_to_all: true } },
        labels: { p: { node_maybe: true } },
      },
    });
    app.uiDataStore.toggleConceptNodeSelection('sheet-1', 'client');

    const latestElements = conceptGraph.update.mock.calls.at(-1)[0];
    expect(latestElements[0].data.label).toBe('client\n?p');
    expect(latestElements[0].classes).toContain('selected_node');
    expect(app._applyEdgeVisibility.mock.calls.at(-1)[1].edgeVisibilityById.e1).toBe(true);
  });

  it('installs the store before applying snapshots through service helpers', () => {
    const argGraph = makeGraph();
    const app = {
      uiDataModel: new UIDataModel(),
      activeSheetId: 'sheet-1',
      selectedArgNode: null,
      sheets: {
        'sheet-1': { id: 'sheet-1', type: 'analysis', argGraph, conceptGraph: makeGraph() },
      },
      _applyEdgeVisibility: vi.fn(),
    };

    applyArgSnapshot(app, 'sheet-1', {
      elements: [{ group: 'nodes', data: { id: 'state_0', label: '0' } }],
    });

    expect(app.uiDataStore).toBeTruthy();
    expect(argGraph.update).toHaveBeenCalledWith([
      { group: 'nodes', data: { id: 'state_0', label: '0' } },
    ], null);
  });

  it('derives reachability-only sheet layout from the model', () => {
    document.body.innerHTML = '<div id="sheet-3" class="sheet-content"></div>';
    const app = {
      uiDataModel: new UIDataModel(),
      activeSheetId: 'sheet-1',
      selectedArgNode: null,
      sheets: {
        'sheet-3': { id: 'sheet-3', type: 'analysis', argGraph: makeGraph(), conceptGraph: null },
      },
      _applyEdgeVisibility: vi.fn(),
    };
    installUIDataModelStore(app);

    app.uiDataStore.registerSheet('sheet-3', { type: 'analysis', reachabilityOnly: true });

    const sheet = document.getElementById('sheet-3');
    expect(app.uiDataModel.sheets['sheet-3'].reachabilityOnly).toBe(true);
    expect(app.sheets['sheet-3'].reachabilityOnly).toBe(true);
    expect(sheet.classList.contains('reachability-only-sheet')).toBe(true);
    expect(sheet.getAttribute('data-sheet-layout')).toBe('reachability-only');
  });

  it('captures graph-rendered positions back into model state without a render loop', () => {
    const argGraph = makeGraph();
    argGraph.update.mockReturnValue({ state_0: { x: 5, y: 6 } });
    const app = {
      uiDataModel: new UIDataModel(),
      activeSheetId: 'sheet-1',
      selectedArgNode: null,
      sheets: {
        'sheet-1': { id: 'sheet-1', type: 'analysis', argGraph, conceptGraph: makeGraph() },
      },
      _applyEdgeVisibility: vi.fn(),
    };
    installUIDataModelStore(app);

    app.uiDataStore.applyArgSnapshot('sheet-1', {
      elements: [{ group: 'nodes', data: { id: 'state_0', label: '0' } }],
    });

    expect(app.uiDataModel.sheets['sheet-1'].argPositions).toEqual({ state_0: { x: 5, y: 6 } });
    expect(argGraph.update).toHaveBeenCalledTimes(1);
  });

  it('does not render event sheets into the active analysis graphs', () => {
    const analysisArg = makeGraph();
    const analysisConcept = makeGraph();
    const app = {
      uiDataModel: new UIDataModel(),
      activeSheetId: 'events-1',
      selectedArgNode: null,
      argGraph: analysisArg,
      conceptGraph: analysisConcept,
      sheets: {
        'events-1': { id: 'events-1', type: 'events' },
      },
      _applyEdgeVisibility: vi.fn(),
    };
    installUIDataModelStore(app);
    app.uiDataStore.registerSheet('events-1', { type: 'events' });

    app.uiDataStore.applyConceptSnapshot('events-1', {
      elements: [{ group: 'nodes', data: { id: 'client', obj: 'client' } }],
    });

    expect(analysisArg.update).not.toHaveBeenCalled();
    expect(analysisConcept.update).not.toHaveBeenCalled();
    expect(app._applyEdgeVisibility).not.toHaveBeenCalled();
    expect(app.uiDataModel.sheets['events-1'].type).toBe('events');
  });

  it('routes ARG and concept selection changes independently', () => {
    const argGraph = makeGraph();
    const conceptGraph = makeGraph();
    const app = {
      uiDataModel: new UIDataModel(),
      activeSheetId: 'sheet-1',
      selectedArgNode: null,
      sheets: {
        'sheet-1': { id: 'sheet-1', type: 'analysis', argGraph, conceptGraph },
      },
      _applyEdgeVisibility: vi.fn(),
    };
    installUIDataModelStore(app);
    app.uiDataStore.applyArgSnapshot('sheet-1', {
      elements: [{ group: 'nodes', data: { id: 'state_0', label: '0' } }],
    });
    app.uiDataStore.applyConceptSnapshot('sheet-1', {
      elements: [{ group: 'nodes', data: { id: 'client', obj: 'client' } }],
    });
    argGraph.update.mockClear();
    conceptGraph.update.mockClear();

    app.uiDataStore.setSelectedArgNode('sheet-1', 'state_0');
    expect(argGraph.highlightNode).toHaveBeenCalledWith('state_0');
    expect(conceptGraph.update).not.toHaveBeenCalled();

    app.uiDataStore.toggleConceptNodeSelection('sheet-1', { id: 'client', obj: 'client' });
    expect(conceptGraph.update).toHaveBeenCalled();
    expect(argGraph.update).not.toHaveBeenCalled();
  });
});
