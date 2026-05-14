import { describe, expect, it, vi } from 'vitest';
import {
  addRelationFromString,
  executeConceptNodeAction,
  materializeEdge,
  materializeEdgeFromSelected,
  removeConcept,
} from './conceptActionService.ts';
import { UIDataModel } from '../models/uiDataModel.ts';
import { createUIDataModelStore } from '../models/uiDataModelStore.ts';

describe('conceptActionService', () => {
  it('dispatches named concept node actions to high-level workflows', async () => {
    const app = {
      removeConcept: vi.fn(),
    };

    await executeConceptNodeAction(app, { obj: 'client=0' }, { id: 'remove' });

    expect(app.removeConcept).toHaveBeenCalledWith('client=0');
  });

  it('materializes edge objects through the API and refreshes', async () => {
    const app = {
      api: {
        materializeEdge: vi.fn(),
      },
      refreshConceptGraph: vi.fn(),
      controls: { setStatus: vi.fn() },
    };

    await materializeEdge(app, {
      obj: 'link',
      source_obj: 'client=0',
      target_obj: 'server=0',
    }, true);

    expect(app.api.materializeEdge).toHaveBeenCalledWith('link', 'client=0', 'server=0', true);
    expect(app.refreshConceptGraph).toHaveBeenCalled();
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Edge materialized (+)', 'success');
  });

  it('materializes an edge from the selected source node after relation choice', async () => {
    const uiDataModel = new UIDataModel();
    const store = createUIDataModelStore(uiDataModel);
    store.applyConceptSnapshot('sheet-1', {
      edge_sorts: {
        link: ['client', 'server'],
        other: ['server', 'client'],
      },
      edges: ['link', 'other'],
    });
    store.setConceptSelections('sheet-1', [
      { kind: 'node', id: 'client', obj: 'client', label: 'client', sourceObj: '', targetObj: '' },
    ]);
    const app = {
      activeSheetId: 'sheet-1',
      uiDataModel,
      listboxDialog: vi.fn(async () => 'link'),
      api: {
        materializeEdge: vi.fn(async () => ({ status: 'ok' })),
      },
      refreshConceptGraph: vi.fn(),
      controls: { setStatus: vi.fn() },
    };

    await materializeEdgeFromSelected(app, 'server');

    expect(app.listboxDialog).toHaveBeenCalledWith(
      'Materialize edge',
      'Materialize this relation from selected node:',
      [{ label: 'link', value: 'link' }],
      { cancel: true },
    );
    expect(app.api.materializeEdge).toHaveBeenCalledWith('link', 'client', 'server', true);
    expect(app.refreshConceptGraph).toHaveBeenCalled();
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Edge materialized (+)', 'success');
  });

  it('dispatches backend projection descriptors to addProjection', async () => {
    const app = {
      addProjection: vi.fn(async () => undefined),
    };

    await executeConceptNodeAction(
      app,
      { id: 'server', obj: 'server' },
      { label: 'p(0,Y,Z)', action: 'add_projection', args: { name: 'p(0,Y,Z)', concept: 'p(0,Y,Z)' } },
    );

    expect(app.addProjection).toHaveBeenCalledWith('p(0,Y,Z)', 'p(0,Y,Z)');
  });

  it('uses the entry dialog and backend action for Add relation', async () => {
    const app = {
      entryDialog: vi.fn(async () => 'link(X,Y)'),
      api: {
        executeAction: vi.fn(async () => ({ status: 'ok' })),
      },
      refreshConceptGraph: vi.fn(),
      controls: { setStatus: vi.fn() },
    };

    await addRelationFromString(app);

    expect(app.entryDialog).toHaveBeenCalledWith(
      'Add relation',
      'Add a relation [example: p(X,a,Y)]:',
      '',
      { okLabel: 'Add' },
    );
    expect(app.api.executeAction).toHaveBeenCalledWith('add_relation', { formula: 'link(X,Y)' });
    expect(app.refreshConceptGraph).toHaveBeenCalled();
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Relation added', 'success');
  });

  it('removes concepts through the API', async () => {
    const app = {
      api: { removeConcept: vi.fn() },
      refreshConceptGraph: vi.fn(),
      controls: { setStatus: vi.fn() },
    };

    await removeConcept(app, 'server=0');

    expect(app.api.removeConcept).toHaveBeenCalledWith('server=0');
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Concept removed', 'success');
  });
});
