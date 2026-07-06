import { beforeEach, describe, expect, it, vi } from 'vitest';
import { UIDataModel } from '../models/uiDataModel.ts';
import { IvyRuntime } from './ivyRuntime.ts';

function installGraphContainer(id = 'concept-graph') {
  document.body.innerHTML = `<div id="${id}"></div>`;
  const graph = document.getElementById(id);
  graph.getBoundingClientRect = () => ({
    left: 10,
    top: 20,
    right: 310,
    bottom: 220,
    width: 300,
    height: 200,
    x: 10,
    y: 20,
    toJSON: () => ({}),
  });
}

function makeRuntime(uiMode = 'cti') {
  const runtime = Object.create(IvyRuntime.prototype) as any;
  runtime.uiMode = uiMode;
  runtime.activeSheetId = 'sheet-1';
  runtime.uiDataModel = new UIDataModel();
  runtime.sheets = {};
  runtime.controls = {
    showContextMenu: vi.fn(),
    hideContextMenu: vi.fn(),
    setStatus: vi.fn(),
  };
  return runtime;
}

function shownLabels(runtime) {
  return runtime.controls.showContextMenu.mock.calls[0][2].map((item) => (
    item.header || item.name || (item.separator ? '---' : '')
  ));
}

describe('IvyRuntime context menus', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
  });

  it('filters concept node right-click actions to CTI projections only', () => {
    installGraphContainer();
    const runtime = makeRuntime('cti');

    runtime.onConceptNodeRightClick({
      id: 'server',
      obj: 'server',
      actions: [
        { label: 'Select', action: 'select' },
        { label: 'Empty', action: 'empty' },
        { label: 'Materialize', action: 'materialize' },
        { label: 'Split with...', action: '' },
        { label: '---', action: '' },
        { label: 'Add projection...', action: '' },
        { label: '---', action: '' },
        { label: 'p(0,Y)', action: 'add_projection', args: { name: 'p(0,Y)', concept: 'p(0,Y)' } },
      ],
    }, { x: 1, y: 2 });

    expect(shownLabels(runtime)).toEqual(['Projections...', '---', 'Add projection...', '---', 'p(0,Y)']);
  });

  it('shows the reachability concept node menu from ivy_graph_ui.py', () => {
    installGraphContainer();
    const runtime = makeRuntime('reachability');

    runtime.onConceptNodeRightClick({ id: 'server', obj: 'server' }, { x: 1, y: 2 });

    expect(shownLabels(runtime)).toEqual([
      'Select',
      'Empty',
      'Materialize',
      'Materialize edge',
      'Splatter',
      'Split with...',
      '---',
    ]);
  });

  it('hides concept edge right-click menus in CTI mode', () => {
    installGraphContainer();
    const runtime = makeRuntime('cti');

    runtime.onConceptEdgeRightClick({ id: 'link', obj: 'link' }, { x: 1, y: 2 });

    expect(runtime.controls.hideContextMenu).toHaveBeenCalled();
    expect(runtime.controls.showContextMenu).not.toHaveBeenCalled();
  });

  it('shows the reachability concept edge menu from ivy_graph_ui.py', () => {
    installGraphContainer();
    const runtime = makeRuntime('reachability');

    runtime.onConceptEdgeRightClick({ id: 'link', obj: 'link' }, { x: 1, y: 2 });

    expect(shownLabels(runtime)).toEqual(['Empty', 'Materialize', 'Dematerialize']);
  });

  it('keeps ARG node right-click as the shared analysis menu', () => {
    installGraphContainer('arg-graph');
    const runtime = makeRuntime('cti');
    runtime.sheets = { 'sheet-1': { argGraph: { containerId: 'arg-graph' } } };

    runtime.onArgNodeRightClick({ id: 'state_0', obj: 'state_0' }, { x: 1, y: 2 }, 'sheet-1');

    expect(shownLabels(runtime)).toEqual([
      'Execute action:',
      '---',
      '---',
      'Check safety',
      'Extend',
      'Mark',
      'Cover by marked',
      'Join with marked',
      'Try conjecture',
      'Try remembered goal',
      'Delete',
    ]);
  });

  it('executes a single <> ARG node action directly without showing a menu', () => {
    installGraphContainer('arg-graph');
    const runtime = makeRuntime('reachability');
    runtime.sheets = { 'sheet-1': { argGraph: { containerId: 'arg-graph' } } };
    runtime.onArgNodeClick = vi.fn();
    runtime.executeArgNodeAction = vi.fn();

    runtime.onArgNodeRightClick({
      id: 'state_0',
      obj: 'state_0',
      actions: [{ label: '<>', action: 'view_state' }],
    }, { x: 1, y: 2 }, 'sheet-1');

    expect(runtime.onArgNodeClick).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'state_0', obj: 'state_0' }),
      'sheet-1',
    );
    expect(runtime.executeArgNodeAction).not.toHaveBeenCalled();
    expect(runtime.controls.showContextMenu).not.toHaveBeenCalled();
  });
});
