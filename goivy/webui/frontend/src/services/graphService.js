export function currentSheet(app) {
  return app.sheets ? app.sheets[app.activeSheetId] : null;
}

export function installGraphStoreHook(sheetId, kind, graph, {
  bridge = globalThis.window && globalThis.window.__ivyVueBridge,
} = {}) {
  if (!graph || graph._ivyGraphStoreHooked) return;
  const origUpdate = graph.update.bind(graph);
  graph.update = function updateWithStoreSnapshot(elements, positions) {
    const result = origUpdate(elements, positions);
    if (bridge && typeof bridge.updateGraphSnapshot === 'function') {
      bridge.updateGraphSnapshot(sheetId, kind, {
        elements: elements || [],
        positions: positions || null,
      });
    }
    return result;
  };
  graph._ivyGraphStoreHooked = true;
}

export function registerSheet(app, sheetId, argGraph, conceptGraph) {
  app.installConceptGraphVisibilityHook(conceptGraph);
  app.installGraphStoreHook(sheetId, 'arg', argGraph);
  app.installGraphStoreHook(sheetId, 'concept', conceptGraph);
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

export function refreshLayoutAfterVuePatch(app, {
  bridge = globalThis.window && globalThis.window.__ivyVueBridge,
  win = globalThis.window,
} = {}) {
  const refresh = () => {
    app._refreshGraphsAndEditorLayout();
  };
  if (bridge && typeof bridge.afterLayoutSettled === 'function') {
    bridge.afterLayoutSettled(() => {
      refresh();
      win.setTimeout(refresh, 60);
    });
    return;
  }
  if (win && typeof win.requestAnimationFrame === 'function') {
    win.requestAnimationFrame(() => {
      win.requestAnimationFrame(refresh);
    });
  } else {
    const setTimeoutFn = win && typeof win.setTimeout === 'function' ? win.setTimeout.bind(win) : setTimeout;
    setTimeoutFn(refresh, 0);
  }
}
