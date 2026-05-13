export function currentSheet(app) {
  return app.sheets ? app.sheets[app.activeSheetId] : null;
}

export function registerSheet(app, sheetId, argGraph, conceptGraph) {
  app.installConceptGraphVisibilityHook(conceptGraph);
  if (app.uiDataModel) app.uiDataModel.registerSheet(sheetId, { type: 'analysis' });
  app.sheets[sheetId] = {
    id: sheetId,
    type: 'analysis',
    argGraph,
    conceptGraph,
    selectedArgNode: null,
    visualOnly: false,
  };
}

export function graphElementsSnapshot(graph) {
  if (!graph || !graph.cy || typeof graph.cy.json !== 'function') return null;
  const json = graph.cy.json();
  return json ? json.elements : null;
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
