import { describe, expect, it, vi } from 'vitest';
import {
  currentSheet,
  graphElementsSnapshot,
  refreshLayoutAfterPatch,
  refreshGraphsAndEditorLayout,
  registerSheet,
} from './graphService.js';

describe('graphService', () => {
  it('registers analysis sheets and reports the current one', () => {
    const app = {
      sheets: {},
      activeSheetId: 'sheet-1',
      installConceptGraphVisibilityHook: vi.fn(),
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
    expect(app.installConceptGraphVisibilityHook).toHaveBeenCalledWith(conceptGraph);
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

  it('defers graph/editor layout refreshes until animation frames settle', () => {
    const win = {
      requestAnimationFrame: vi.fn((callback) => callback()),
    };
    const app = {
      _refreshGraphsAndEditorLayout: vi.fn(),
    };

    refreshLayoutAfterPatch(app, { win });

    expect(win.requestAnimationFrame).toHaveBeenCalledTimes(2);
    expect(app._refreshGraphsAndEditorLayout).toHaveBeenCalledTimes(1);
  });
});
