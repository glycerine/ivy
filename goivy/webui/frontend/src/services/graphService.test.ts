import { describe, expect, it, vi } from 'vitest';
import {
  currentSheet,
  refreshLayoutAfterPatch,
  refreshGraphsAndEditorLayout,
  registerSheet,
} from './graphService.ts';
import { UIDataModel } from '../models/uiDataModel.ts';

describe('graphService', () => {
  it('registers analysis sheets and reports the current one', () => {
    const app = {
      sheets: {},
      activeSheetId: 'sheet-1',
      uiDataModel: new UIDataModel(),
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
      reachabilityOnly: false,
    });
    expect(app.installConceptGraphVisibilityHook).toHaveBeenCalledWith(conceptGraph);
    expect(app.uiDataModel.sheets['sheet-1'].type).toBe('analysis');
  });

  it('registers reachability-only sheets through the model before mirroring runtime state', () => {
    const app = {
      sheets: {},
      activeSheetId: 'sheet-1',
      uiDataModel: new UIDataModel(),
      installConceptGraphVisibilityHook: vi.fn(),
    };

    registerSheet(app, 'sheet-3', {}, null, { reachabilityOnly: true });

    expect(app.uiDataModel.sheets['sheet-3'].reachabilityOnly).toBe(true);
    expect(app.sheets['sheet-3'].reachabilityOnly).toBe(true);
    expect(app.sheets['sheet-3'].conceptGraph).toBeNull();
  });

  it('refreshes visible graph layouts', () => {
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
