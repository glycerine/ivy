import { describe, expect, it, vi } from 'vitest';
import { UIDataModel } from '../models/uiDataModel.ts';
import { installUIDataModelStore } from './uiDataRenderService.ts';

function makeGraph() {
  return {
    update: vi.fn(),
    highlightNode: vi.fn(),
    clearHighlights: vi.fn(),
  };
}

describe('uiDataRenderService', () => {
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
});
