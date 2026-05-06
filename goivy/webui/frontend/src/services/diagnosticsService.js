import { runCommand } from './commandRegistry.js';
import { currentAppServices } from './appServices.js';

function graphFor(runtime, kind) {
  if (!runtime) return null;
  return kind === 'arg' ? runtime.argGraph : runtime.conceptGraph;
}

function graphElements(graph) {
  if (!graph || !graph.cy || typeof graph.cy.elements !== 'function') return [];
  const elements = graph.cy.elements();
  if (elements && typeof elements.jsons === 'function') return elements.jsons();
  if (elements && typeof elements.toArray === 'function') {
    return elements.toArray().map((element) => (typeof element.json === 'function' ? element.json() : element.data && element.data()));
  }
  return [];
}

export function createIvyDiagnostics({
  services = currentAppServices(),
} = {}) {
  const runtime = () => (services && typeof services.runtime === 'function' ? services.runtime() : null);
  return {
    runtime,

    sessionId() {
      const app = runtime();
      return (app && app.api && app.api.sessionId) || '';
    },

    runCommand(name, ...args) {
      return runCommand(name, ...args);
    },

    graphNodeCount(kind = 'concept') {
      const graph = graphFor(runtime(), kind);
      if (!graph || !graph.cy || typeof graph.cy.nodes !== 'function') return 0;
      return graph.cy.nodes().length;
    },

    graphSnapshot(kind = 'concept') {
      return graphElements(graphFor(runtime(), kind));
    },

    editorContent() {
      const app = runtime();
      if (!app) return '';
      if (app.cmEditor && typeof app.cmEditor.getValue === 'function') {
        return app.cmEditor.getValue();
      }
      return app._persistedFileContent || '';
    },

    sheetSnapshot() {
      const app = runtime();
      return app ? {
        activeSheetId: app.activeSheetId,
        sheets: app.sheets || {},
      } : { activeSheetId: '', sheets: {} };
    },
  };
}

export function installIvyDiagnostics({
  win = globalThis.window,
  services = currentAppServices(),
} = {}) {
  if (!win) return () => {};
  const diagnostics = createIvyDiagnostics({ services });
  win.__ivyDiagnostics = diagnostics;
  return () => {
    if (win.__ivyDiagnostics === diagnostics) {
      delete win.__ivyDiagnostics;
    }
  };
}
