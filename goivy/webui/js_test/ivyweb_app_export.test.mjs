import { describe, expect, it, vi } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import { FakeAPI, FakeControls, FakeGraph, makePersist } from './helpers/fakes.mjs';

function makeExportApp() {
  document.body.innerHTML = '<div id="statusbar"></div>';
  const IvyApp = loadIvyApp({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
    IvyPersist: makePersist(),
  });
  const app = new IvyApp();
  app.controls = new FakeControls();
  app.activeSheetId = 'sheet-2';
  return app;
}

describe('IvyApp concept graph export', () => {
  it('requests DOT export for the active sheet', async () => {
    const app = makeExportApp();
    Object.defineProperty(URL, 'createObjectURL', {
      configurable: true,
      value: vi.fn(() => 'blob:concept-graph'),
    });
    Object.defineProperty(URL, 'revokeObjectURL', {
      configurable: true,
      value: vi.fn(),
    });
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {});
    app.api = {
      executeAction: vi.fn(async () => ({
        content: 'digraph concept_graph {\n}\n',
        filename: 'concept_graph.dot',
      })),
    };

    await app.exportConjecture();

    expect(app.api.executeAction).toHaveBeenCalledWith('export', { sheet_id: 'sheet-2' });
    expect(app.controls.lastStatus).toEqual({ message: 'Graph exported', kind: 'success' });
  });
});
