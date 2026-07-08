import { afterEach, describe, expect, it, vi } from 'vitest';
import { ARG_STYLE, CONCEPT_STYLE, IvyGraph, installGraphGlobalsForCompatibility } from './graphRuntime.ts';

function makeFakeCy() {
  const nodes = [];
  const edges = [];
  const layouts = [];
  let collection;
  const cy = {
    added: [],
    handlers: [],
    elements: vi.fn(() => collection),
    add: vi.fn((elements) => {
      const list = Array.isArray(elements) ? elements : [elements];
      cy.added.push(...list);
      list.filter((element) => element.group === 'nodes').forEach((element) => {
        nodes.push(makeNode(element));
      });
      list.filter((element) => element.group === 'edges').forEach((element) => {
        edges.push(makeEdge(element));
      });
    }),
    nodes: vi.fn(() => {
      const collection = [...nodes];
      collection.removeClass = vi.fn();
      return collection;
    }),
    edges: vi.fn(() => [...edges]),
    getElementById: vi.fn((id) => nodes.find((node) => node.id() === id) || { length: 0 }),
    fit: vi.fn(),
    reset: vi.fn(),
    resize: vi.fn(),
    destroy: vi.fn(),
    animate: vi.fn(),
    layout: vi.fn((options) => {
      layouts.push(options);
      return { run: vi.fn() };
    }),
    on: vi.fn((...args) => {
      cy.handlers.push(args);
    }),
    off: vi.fn(),
    $: vi.fn(() => [{ data: () => ({ id: 'selected' }) }]),
    layouts,
  };
  collection = {
    remove: vi.fn(),
    not: vi.fn(() => collection),
    layout: vi.fn((options) => {
      layouts.push(options);
      return { run: vi.fn() };
    }),
  };
  return cy;
}

function makeNode(element) {
  const data = element.data || {};
  let currentPosition = element.position || { x: 0, y: 0 };
  return {
    length: 1,
    id: () => data.id,
    data: vi.fn((key) => (key ? data[key] : data)),
    style: vi.fn(),
    position: vi.fn((next) => {
      if (next) currentPosition = next;
      return currentPosition;
    }),
    addClass: vi.fn(),
    remove: vi.fn(),
  };
}

function makeEdge(element) {
  const data = element.data || {};
  return {
    length: 1,
    id: () => data.id,
    data: vi.fn((key) => (key ? data[key] : data)),
    style: vi.fn(),
  };
}

afterEach(() => {
  delete window.cytoscape;
  delete window.cytoscapeDagre;
  delete window.IvyGraph;
  delete window.ARG_STYLE;
  delete window.CONCEPT_STYLE;
  delete window.PROOF_STYLE;
});

describe('graphRuntime', () => {
  it('installs graph globals for compatibility harnesses', () => {
    installGraphGlobalsForCompatibility(window);

    expect(window.IvyGraph).toBe(IvyGraph);
    expect(window.ARG_STYLE).toBeTruthy();
    expect(window.CONCEPT_STYLE).toBeTruthy();
    expect(window.PROOF_STYLE).toBeTruthy();
  });

  it('styles safe and marked ARG node classes', () => {
    expect(ARG_STYLE.some((entry) => entry.selector === 'node.safe_state')).toBe(true);
    const marked = ARG_STYLE.find((entry) => entry.selector === 'node.marked_state');
    expect(marked?.style['background-color']).toBe('#b73535');
  });

  it('keeps semantic graph style affordances covered', () => {
    const styleFor = (style, selector) => style.find((entry) => entry.selector === selector)?.style || {};

    expect(styleFor(ARG_STYLE, 'node')['text-wrap']).toBe('wrap');
    expect(styleFor(ARG_STYLE, 'node.bottom_state')['background-color']).toBe('#000');
    expect(styleFor(ARG_STYLE, 'node.safe_state')['background-color']).toBe('#2f8f46');
    expect(styleFor(ARG_STYLE, 'node.marked_state')['background-color']).toBe('#b73535');
    expect(styleFor(ARG_STYLE, 'edge.cover')['line-style']).toBe('dashed');
    expect(styleFor(ARG_STYLE, 'edge.transition_join')['target-arrow-shape']).toBe('triangle-backcurve');
    expect(styleFor(ARG_STYLE, 'edge.transition_action')['target-arrow-shape']).toBe('triangle');
    expect(styleFor(ARG_STYLE, 'node:selected')['overlay-opacity']).toBe(0.2);
    expect(styleFor(ARG_STYLE, 'edge:selected')['overlay-opacity']).toBe(0.2);

    expect(styleFor(CONCEPT_STYLE, 'node')['text-wrap']).toBe('wrap');
    expect(styleFor(CONCEPT_STYLE, 'node.exactly_one')['border-style']).toBe('solid');
    expect(styleFor(CONCEPT_STYLE, 'node.at_least_one')['border-style']).toBe('double');
    expect(styleFor(CONCEPT_STYLE, 'node.at_most_one')['border-style']).toBe('dotted');
    expect(styleFor(CONCEPT_STYLE, 'node.node_unknown')['border-width']).toBe('0px');
    expect(styleFor(CONCEPT_STYLE, 'edge.none_to_none')['line-style']).toBe('dashed');
    expect(styleFor(CONCEPT_STYLE, 'edge.all_to_all')['line-style']).toBe('solid');
    expect(styleFor(CONCEPT_STYLE, 'edge.edge_unknown')['line-style']).toBe('dotted');
    expect(styleFor(CONCEPT_STYLE, 'edge.total')['source-arrow-shape']).toBe('circle');
    expect(styleFor(CONCEPT_STYLE, 'edge.functional')['source-arrow-shape']).toBe('square');
    expect(styleFor(CONCEPT_STYLE, 'edge.injective')['target-arrow-shape']).toBe('triangle-backcurve');
    expect(styleFor(CONCEPT_STYLE, 'edge.surjective')['target-arrow-fill']).toBe('filled');
    expect(styleFor(CONCEPT_STYLE, 'node:selected')['overlay-opacity']).toBe(0);
    expect(styleFor(CONCEPT_STYLE, 'edge:selected')['overlay-opacity']).toBe(0);
  });

  it('wraps semantic graph edge labels for long actions and relations', () => {
    const styleFor = (style, selector) => style.find((entry) => entry.selector === selector)?.style || {};

    expect(styleFor(ARG_STYLE, 'edge')['text-wrap']).toBe('wrap');
    expect(styleFor(ARG_STYLE, 'edge').content).toBeUndefined();
    expect(styleFor(ARG_STYLE, 'edge[label]').content).toBe('data(label)');
    expect(styleFor(ARG_STYLE, 'edge')['text-max-width']).toBeUndefined();
    expect(styleFor(ARG_STYLE, 'edge[text_max_width]')['text-max-width']).toBe('data(text_max_width)');
    expect(styleFor(CONCEPT_STYLE, 'edge').content).toBeUndefined();
    expect(styleFor(CONCEPT_STYLE, 'edge[label]').content).toBe('data(label)');
    expect(styleFor(CONCEPT_STYLE, 'edge[label]').color).toBe('#000');
    expect(styleFor(CONCEPT_STYLE, 'edge[label]')['text-outline-width']).toBe('3px');
    expect(styleFor(CONCEPT_STYLE, 'edge[label]')['text-outline-color']).toBe('#ffd400');
    expect(styleFor(CONCEPT_STYLE, 'edge')['text-wrap']).toBe('wrap');
    expect(styleFor(CONCEPT_STYLE, 'edge')['text-max-width']).toBeUndefined();
    expect(styleFor(CONCEPT_STYLE, 'edge[text_max_width]')['text-max-width']).toBe('data(text_max_width)');
  });

  it('renders concept subgraph shapes as compound cluster boxes', () => {
    expect(CONCEPT_STYLE.some((entry) => entry.selector === 'node.subgraph_box')).toBe(true);

    const cy = makeFakeCy();
    window.cytoscape = vi.fn(() => cy);
    document.body.innerHTML = '<div id="concept-graph"></div>';
    const graph = new IvyGraph('concept-graph', CONCEPT_STYLE);

    graph.update([
      { group: 'nodes', data: { id: 'n0', obj: '0:client', label: 'client', cluster: 'client' } },
      { group: 'nodes', data: { id: 'n1', obj: '1:client', label: 'client', cluster: 'client' } },
      { group: 'shapes', classes: 'subgraphs', data: { id: 's0', obj: 'cluster_client', shape: 'rectangle', cluster: 'client' } },
    ], null);

    const parent = cy.added.find((element) => element.data.id === 's0');
    expect(parent).toMatchObject({
      group: 'nodes',
      classes: expect.stringContaining('subgraph_box'),
      data: expect.objectContaining({ cluster: 'client', shape: 'rectangle' }),
    });
    expect(cy.added.filter((element) => element.data.parent === 's0').map((element) => element.data.id)).toEqual(['n0', 'n1']);
    expect(cy.added.some((element) => element.group === 'shapes')).toBe(false);
  });

  it('maps Python generic shape coordinates to locked Cytoscape box nodes', () => {
    const cy = makeFakeCy();
    window.cytoscape = vi.fn(() => cy);
    document.body.innerHTML = '<div id="concept-graph"></div>';
    const graph = new IvyGraph('concept-graph', CONCEPT_STYLE);

    graph.update([
      {
        group: 'shapes',
        classes: 'subgraphs',
        locked: true,
        data: {
          id: 'cluster_0',
          obj: 'cluster_0',
          label: 'client',
          shape: 'rectangle',
          coords: [
            { x: 10, y: 20 },
            { x: 110, y: 20 },
            { x: 110, y: 80 },
            { x: 10, y: 80 },
          ],
        },
      },
    ], null);

    const box = cy.added.find((element) => element.data.id === 'cluster_0');
    expect(box).toMatchObject({
      group: 'nodes',
      locked: true,
      position: { x: 60, y: 50 },
      data: expect.objectContaining({
        obj: 'cluster_0',
        label: 'client',
        shape: 'rectangle',
        width: 100,
        height: 60,
        generic_shape: true,
      }),
    });
    expect(box.classes).toContain('subgraph_box');
    expect(box.classes).toContain('generic_shape');
    expect(cy.added.some((element) => element.group === 'shapes')).toBe(false);
    expect(cy.layout).not.toHaveBeenCalled();
  });

  it('uses layout-only edges for reversed and unconstrained layout edges', () => {
    expect(CONCEPT_STYLE.some((entry) => entry.selector === 'edge.layout_only')).toBe(true);

    const cy = makeFakeCy();
    window.cytoscape = vi.fn(() => cy);
    document.body.innerHTML = '<div id="concept-graph"></div>';
    const graph = new IvyGraph('concept-graph', CONCEPT_STYLE);

    graph.update([
      { group: 'nodes', data: { id: 'n0', obj: 'A', label: 'A' } },
      { group: 'nodes', data: { id: 'n1', obj: 'B', label: 'B' } },
      {
        group: 'edges',
        classes: 'edge_unknown',
        data: {
          id: 'e0',
          obj: 'rel',
          source: 'n1',
          target: 'n0',
          layout_source: 'n0',
          layout_target: 'n1',
          layout_reversed: true,
        },
      },
      {
        group: 'edges',
        classes: 'edge_unknown',
        data: {
          id: 'e1',
          obj: 'pending',
          source: 'n0',
          target: 'n1',
          layout_constraint: false,
        },
      },
    ], null);

    const visibleRel = cy.added.find((element) => element.data.id === 'e0');
    expect(visibleRel.data).toMatchObject({ source: 'n1', target: 'n0' });
    expect(visibleRel.classes).toContain('layout_ignored');

    const layoutRel = cy.added.find((element) => element.data.id === 'layout_e0');
    expect(layoutRel).toMatchObject({
      group: 'edges',
      classes: expect.stringContaining('layout_only'),
      data: expect.objectContaining({ source: 'n0', target: 'n1', layout_for: 'e0' }),
    });

    const pending = cy.added.find((element) => element.data.id === 'e1');
    expect(pending.classes).toContain('layout_ignored');
    expect(cy.added.some((element) => element.data.id === 'layout_e1')).toBe(false);
    expect(cy.elements().not).toHaveBeenCalledWith('.layout_ignored');
  });

  it('updates Cytoscape elements with Ivy defaults and supplied positions', () => {
    const cy = makeFakeCy();
    window.cytoscape = vi.fn(() => cy);
    document.body.innerHTML = '<div id="arg-graph"></div>';
    const graph = new IvyGraph('arg-graph', ARG_STYLE);

    graph.update([
      { group: 'nodes', data: { id: 'n1', label: 'state' } },
      { group: 'edges', data: { id: 'e1', source: 'n1', target: 'n1' } },
    ], { n1: { x: 12, y: 34 } });

    expect(window.cytoscape).toHaveBeenCalledWith(expect.objectContaining({
      container: document.getElementById('arg-graph'),
      style: ARG_STYLE,
    }));
    expect(cy.added[0].data).toMatchObject({
      width: 60,
      height: 50,
      shape: 'ellipse',
      border_color: '#000',
    });
    expect(cy.getElementById('n1').position).toHaveBeenCalledWith({ x: 12, y: 34 });
    expect(cy.fit).not.toHaveBeenCalled();
  });

  it('applies relation line colors to concept graph edges', () => {
    const cy = makeFakeCy();
    window.cytoscape = vi.fn(() => cy);
    document.body.innerHTML = '<div id="concept-graph"></div>';
    const graph = new IvyGraph('concept-graph', CONCEPT_STYLE);

    graph.update([
      { group: 'nodes', data: { id: 'n0', obj: 'Client', label: 'Client' } },
      { group: 'nodes', data: { id: 'n1', obj: 'Server', label: 'Server' } },
      {
        group: 'edges',
        classes: 'all_to_all',
        data: { id: 'e0', obj: 'link', source: 'n0', target: 'n1', line_color: '#0000ff' },
      },
    ], null);

    const edge = cy.edges()[0];
    expect(edge.style).toHaveBeenCalledWith('line-color', '#0000ff');
    expect(edge.style).toHaveBeenCalledWith('target-arrow-color', '#0000ff');
    expect(edge.style).toHaveBeenCalledWith('source-arrow-color', '#0000ff');
  });

  it('fits and resets the graph viewport as the scrollbar replacement', () => {
    const cy = makeFakeCy();
    window.cytoscape = vi.fn(() => cy);
    document.body.innerHTML = '<div id="arg-graph"></div>';
    const graph = new IvyGraph('arg-graph', ARG_STYLE);

    graph.fit();
    graph.resetView();

    expect(cy.fit).toHaveBeenCalledWith(undefined, 30);
    expect(cy.reset).toHaveBeenCalledTimes(1);
    expect(cy.fit).toHaveBeenCalledTimes(2);
  });

  it('centers highlighted graph nodes so large graphs remain navigable', () => {
    const cy = makeFakeCy();
    window.cytoscape = vi.fn(() => cy);
    document.body.innerHTML = '<div id="arg-graph"></div>';
    const graph = new IvyGraph('arg-graph', ARG_STYLE);

    graph.update([
      { group: 'nodes', data: { id: 'state_99', label: '99' } },
    ], { state_99: { x: 2400, y: 1800 } });
    const node = cy.getElementById('state_99');
    cy.animate.mockClear();

    graph.highlightNode('state_99');

    expect(node.addClass).toHaveBeenCalledWith('highlighted');
    expect(cy.animate).toHaveBeenCalledWith({
      center: { eles: node },
      duration: 300,
    });
  });

  it('preserves existing node positions when an edge-only update arrives', () => {
    const cy = makeFakeCy();
    window.cytoscape = vi.fn(() => cy);
    document.body.innerHTML = '<div id="arg-graph"></div>';
    const graph = new IvyGraph('arg-graph', ARG_STYLE);

    graph.update([
      { group: 'nodes', data: { id: 'n1', label: 'state' } },
    ], { n1: { x: 12, y: 34 } });
    cy.layout.mockClear();

    graph.update([
      { group: 'nodes', data: { id: 'n1', label: 'state' } },
      { group: 'edges', data: { id: 'e1', source: 'n1', target: 'n1' } },
    ], null);

    expect(cy.getElementById('n1').position).toHaveBeenCalledWith({ x: 12, y: 34 });
    expect(cy.layout).not.toHaveBeenCalled();
  });
});
