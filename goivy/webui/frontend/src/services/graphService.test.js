import { describe, expect, it, vi } from 'vitest';
import {
  currentSheet,
  graphElementsSnapshot,
  installGraphStoreHook,
  refreshGraphsAndEditorLayout,
  registerSheet,
} from './graphService.js';

describe('graphService', () => {
  it('registers analysis sheets and reports the current one', () => {
    const app = {
      sheets: {},
      activeSheetId: 'sheet-1',
      installConceptGraphVisibilityHook: vi.fn(),
      installGraphStoreHook: vi.fn(),
    };
    const argGraph = {};
    const conceptGraph = {};

    registerSheet(app, 'sheet-1', argGraph, conceptGraph);

    expect(currentSheet(app)).toEqual({
      id: 'sheet-1',
      type: 'analysis',
      argGraph,
      conceptGraph,
      selectedArgNode: null,
      visualOnly: false,
    });
    expect(app.installGraphStoreHook).toHaveBeenCalledWith('sheet-1', 'arg', argGraph);
    expect(app.installGraphStoreHook).toHaveBeenCalledWith('sheet-1', 'concept', conceptGraph);
  });

  it('hooks graph updates into the Vue graph store bridge', () => {
    const graph = {
      update: vi.fn(() => 'updated'),
    };
    const bridge = {
      updateGraphSnapshot: vi.fn(),
    };

    installGraphStoreHook('sheet-2', 'concept', graph, { bridge });

    expect(graph.update([{ data: { id: 'n' } }], { n: { x: 1, y: 2 } })).toBe('updated');
    expect(bridge.updateGraphSnapshot).toHaveBeenCalledWith('sheet-2', 'concept', {
      elements: [{ data: { id: 'n' } }],
      positions: { n: { x: 1, y: 2 } },
    });
  });

  it('creates graph element snapshots and refreshes visible graph layouts', () => {
    const elements = [{ data: { id: 'n' } }];
    expect(graphElementsSnapshot({ cy: { json: () => ({ elements }) } })).toBe(elements);

    const app = {
      argGraph: { resize: vi.fn() },
      conceptGraph: { resize: vi.fn() },
      _refreshEditorLayout: vi.fn(),
    };
    refreshGraphsAndEditorLayout(app);

    expect(app.argGraph.resize).toHaveBeenCalled();
    expect(app.conceptGraph.resize).toHaveBeenCalled();
    expect(app._refreshEditorLayout).toHaveBeenCalled();
  });
});
