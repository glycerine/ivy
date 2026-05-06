import { afterEach, describe, expect, it, vi } from 'vitest';
import { ARG_STYLE, IvyGraph, installGraphGlobalsForCompatibility } from './graphRuntime.js';

function makeFakeCy() {
  const nodes = [];
  const layouts = [];
  const cy = {
    added: [],
    handlers: [],
    elements: vi.fn(() => ({ remove: vi.fn() })),
    add: vi.fn((elements) => {
      const list = Array.isArray(elements) ? elements : [elements];
      cy.added.push(...list);
      list.filter((element) => element.group === 'nodes').forEach((element) => {
        nodes.push(makeNode(element.data));
      });
    }),
    nodes: vi.fn(() => {
      const collection = [...nodes];
      collection.removeClass = vi.fn();
      return collection;
    }),
    getElementById: vi.fn((id) => nodes.find((node) => node.id() === id) || { length: 0 }),
    fit: vi.fn(),
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
  return cy;
}

function makeNode(data) {
  return {
    length: 1,
    id: () => data.id,
    data: vi.fn((key) => (key ? data[key] : data)),
    style: vi.fn(),
    position: vi.fn(),
    addClass: vi.fn(),
    remove: vi.fn(),
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
    expect(cy.fit).toHaveBeenCalledWith(undefined, 30);
  });
});
