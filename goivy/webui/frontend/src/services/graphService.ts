import { installUIDataModelStore } from './uiDataRenderService.ts';
import { selectSheet } from '../models/uiDataSelectors.ts';

export function currentSheet(app) {
  return app.sheets ? app.sheets[app.activeSheetId] : null;
}

export function registerSheet(app, sheetId, argGraph, conceptGraph, raw = {}) {
  const rawRecord: any = raw && typeof raw === 'object' && !Array.isArray(raw) ? raw : {};
  app.installConceptGraphVisibilityHook(conceptGraph);
  let modelSheet = null;
  if (app.uiDataModel) {
    const store = app.uiDataStore || installUIDataModelStore(app);
    if (store) modelSheet = store.registerSheet(sheetId, { type: 'analysis', ...rawRecord });
    if (!modelSheet) modelSheet = selectSheet(app.uiDataModel, sheetId);
  }
  const reachabilityOnly = !!(modelSheet ? modelSheet.reachabilityOnly : rawRecord.reachabilityOnly);
  app.sheets[sheetId] = {
    id: sheetId,
    type: 'analysis',
    argGraph,
    conceptGraph,
    selectedArgNode: null,
    visualOnly: false,
    reachabilityOnly,
  };
}

export function refreshGraphsAndEditorLayout(app) {
  if (app.argGraph) app.argGraph.resize();
  if (app.conceptGraph) app.conceptGraph.resize();
  app._refreshEditorLayout();
}

export function refreshLayoutAfterPatch(app, {
  win = globalThis.window,
} = {}) {
  const refresh = () => {
    app._refreshGraphsAndEditorLayout();
  };
  if (win && typeof win.requestAnimationFrame === 'function') {
    win.requestAnimationFrame(() => {
      win.requestAnimationFrame(refresh);
    });
  } else {
    const setTimeoutFn = win && typeof win.setTimeout === 'function' ? win.setTimeout.bind(win) : setTimeout;
    setTimeoutFn(refresh, 0);
  }
}
