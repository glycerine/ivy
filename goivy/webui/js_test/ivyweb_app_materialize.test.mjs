import { describe, expect, it, vi } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import { FakeAPI, FakeControls, FakeGraph, makePersist } from './helpers/fakes.mjs';

function makeMaterializeApp() {
  document.body.innerHTML = '<div id="statusbar"></div>';
  const IvyApp = loadIvyApp({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
    IvyPersist: makePersist(),
  });
  const app = new IvyApp();
  app.controls = new FakeControls();
  app.refreshConceptGraph = vi.fn(async () => undefined);
  return app;
}

describe('IvyApp concept edge materialization', () => {
  it('sends relation, source, target, and polarity for edge materialization', async () => {
    const app = makeMaterializeApp();
    app.api = {
      materializeEdge: vi.fn(async () => ({ status: 'ok' })),
    };

    await app.materializeEdge({
      obj: 'link',
      source_obj: 'client',
      target_obj: 'server',
    }, false);

    expect(app.api.materializeEdge).toHaveBeenCalledWith('link', 'client', 'server', false);
    expect(app.refreshConceptGraph).toHaveBeenCalled();
    expect(app.controls.lastStatus).toEqual({ message: 'Edge materialized (\u2013)', kind: 'success' });
  });

  it('materializes an edge from the selected source node after relation choice', async () => {
    const app = makeMaterializeApp();
    app.selectedConceptNode = 'client';
    app._lastConceptData = {
      edge_sorts: {
        link: ['client', 'server'],
        other: ['server', 'client'],
      },
      edges: ['link', 'other'],
    };
    app.api = {
      materializeEdge: vi.fn(async () => ({ status: 'ok' })),
    };

    const result = app.materializeEdgeFromSelected('server');
    document.querySelector('[data-ivy-dialog-list]').value = 'link';
    const ok = Array.from(document.querySelectorAll('[data-ivy-dialog-button]')).find((btn) => btn.textContent === 'OK');
    ok.click();
    await result;

    expect(app.api.materializeEdge).toHaveBeenCalledWith('link', 'client', 'server', true);
    expect(app.refreshConceptGraph).toHaveBeenCalled();
    expect(app.controls.lastStatus).toEqual({ message: 'Edge materialized (+)', kind: 'success' });
  });

  it('dispatches backend projection descriptors to addProjection', async () => {
    const app = makeMaterializeApp();
    app.addProjection = vi.fn(async () => undefined);

    await app.executeConceptNodeAction(
      { id: 'server', obj: 'server' },
      { label: 'p(0,Y,Z)', action: 'add_projection', args: { name: 'p(0,Y,Z)', concept: 'p(0,Y,Z)' } }
    );

    expect(app.addProjection).toHaveBeenCalledWith('p(0,Y,Z)', 'p(0,Y,Z)');
  });

  it('uses the entry dialog for Add relation', async () => {
    const app = makeMaterializeApp();
    app.api = {
      executeAction: vi.fn(async () => ({ status: 'ok' })),
    };

    const result = app.addRelationFromString();
    document.querySelector('[data-ivy-dialog-entry]').value = 'link(X,Y)';
    const add = Array.from(document.querySelectorAll('[data-ivy-dialog-button]')).find((btn) => btn.textContent === 'Add');
    add.click();
    await result;

    expect(app.api.executeAction).toHaveBeenCalledWith('add_relation', { formula: 'link(X,Y)' });
    expect(app.refreshConceptGraph).toHaveBeenCalled();
    expect(app.controls.lastStatus).toEqual({ message: 'Relation added', kind: 'success' });
  });
});
