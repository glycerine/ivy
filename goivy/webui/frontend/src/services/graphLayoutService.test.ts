import { describe, expect, it, vi } from 'vitest';
import {
  buildDotLayoutGraph,
  computeGraphvizDotPositions,
  parseGraphvizPlainPositions,
  relayoutGraphWithDot,
} from './graphLayoutService.ts';
import { UIDataModel } from '../models/uiDataModel.ts';
import { createUIDataModelStore } from '../models/uiDataModelStore.ts';

describe('graphLayoutService', () => {
  it('builds DOT from model-derived graph elements', () => {
    const layoutGraph = buildDotLayoutGraph([
      { group: 'nodes', data: { id: 'client', label: 'client' } },
      { group: 'nodes', data: { id: 'server', label: 'server' } },
      { group: 'edges', data: { id: 'link', source: 'client', target: 'server' } },
    ]);

    expect(layoutGraph.dot).toContain('digraph ivy_graph');
    expect(layoutGraph.dot).toContain('n0 [label="client"]');
    expect(layoutGraph.dot).toContain('n0 -> n1');
    expect(layoutGraph.idByDotId).toEqual({ n0: 'client', n1: 'server' });
  });

  it('parses Graphviz plain coordinates into Cytoscape positions', () => {
    const positions = parseGraphvizPlainPositions(
      [
        'graph 1 2 3',
        'node n0 0.5 2.5 1 0.5 client solid ellipse black lightgrey',
        'node n1 1.5 0.5 1 0.5 server solid ellipse black lightgrey',
        'stop',
      ].join('\n'),
      { n0: 'client', n1: 'server' },
      { scale: 100 },
    );

    expect(positions).toEqual({
      client: { x: 50, y: 50 },
      server: { x: 150, y: 250 },
    });
  });

  it('computes and stores explicit dot relayout positions through the model store', async () => {
    const uiDataModel = new UIDataModel();
    const uiDataStore = createUIDataModelStore(uiDataModel);
    uiDataStore.applyConceptSnapshot('sheet-1', {
      elements: [
        { group: 'nodes', data: { id: 'client', label: 'client' } },
        { group: 'nodes', data: { id: 'server', label: 'server' } },
        { group: 'edges', data: { id: 'link', source: 'client', target: 'server' } },
      ],
    });
    const graphviz = {
      layout: vi.fn(() => [
        'graph 1 2 2',
        'node n0 0.25 1.75 1 0.5 client solid ellipse black lightgrey',
        'node n1 1.25 0.75 1 0.5 server solid ellipse black lightgrey',
        'stop',
      ].join('\n')),
    };
    const app = { uiDataModel, uiDataStore, activeSheetId: 'sheet-1' };

    const positions = await relayoutGraphWithDot(app, 'concept', { graphviz, scale: 100 });

    expect(graphviz.layout).toHaveBeenCalledWith(expect.stringContaining('n0 -> n1'), 'plain', 'dot');
    expect(positions).toEqual({
      client: { x: 25, y: 25 },
      server: { x: 125, y: 125 },
    });
    expect(uiDataModel.sheets['sheet-1'].conceptPositions).toEqual(positions);
  });

  it('uses injected Graphviz engines for tests without loading WASM', async () => {
    const graphviz = {
      layout: vi.fn(() => 'graph 1 1 1\nnode n0 0.5 0.5 1 1 x solid ellipse black white\nstop\n'),
    };

    const positions = await computeGraphvizDotPositions([
      { group: 'nodes', data: { id: 'x', label: 'x' } },
    ], { graphviz, scale: 10 });

    expect(positions).toEqual({ x: { x: 5, y: 5 } });
  });
});
