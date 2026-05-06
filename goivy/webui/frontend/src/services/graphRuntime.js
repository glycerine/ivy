export const CONCEPT_STYLE = [
  {
    selector: 'node',
    style: {
      content: 'data(label)',
      'text-wrap': 'wrap',
      'font-size': '14px',
      'text-outline-width': '0px',
      'text-valign': 'center',
      color: '#000',
      width: 'data(width)',
      height: 'data(height)',
      'border-color': '#000',
      shape: 'data(shape)',
      'background-color': '#fff',
    },
  },
  { selector: 'node.non_existing', style: { display: 'none' } },
  {
    selector: 'node.exactly_one',
    style: { 'border-width': '4px', 'border-style': 'solid', 'border-color': '#000' },
  },
  {
    selector: 'node.at_least_one',
    style: { 'border-width': '8px', 'border-style': 'double', 'border-color': '#000' },
  },
  {
    selector: 'node.at_most_one',
    style: { 'border-width': '3px', 'border-style': 'dotted', 'border-color': '#000' },
  },
  {
    selector: 'node.node_unknown',
    style: { 'border-width': '5px', 'border-style': 'double', 'border-color': '#000' },
  },
  {
    selector: 'edge',
    style: {
      width: '3px',
      'line-color': '#888',
      'target-arrow-color': '#888',
      'source-arrow-color': '#888',
      'target-arrow-shape': 'triangle',
      'target-arrow-fill': 'filled',
      'source-arrow-fill': 'filled',
      'curve-style': 'bezier',
    },
  },
  {
    selector: 'edge.none_to_none',
    style: {
      width: '4px',
      'line-style': 'dashed',
      'target-arrow-shape': 'triangle',
      'target-arrow-fill': 'filled',
      'source-arrow-fill': 'filled',
    },
  },
  { selector: 'edge.all_to_all', style: { width: '4px', 'line-style': 'solid' } },
  { selector: 'edge.edge_unknown', style: { width: '4px', 'line-style': 'dotted' } },
  {
    selector: 'edge.total',
    style: { 'source-arrow-shape': 'circle', 'source-arrow-fill': 'filled' },
  },
  { selector: 'edge.functional', style: { 'source-arrow-shape': 'square' } },
  { selector: 'edge.injective', style: { 'target-arrow-shape': 'triangle-backcurve' } },
  { selector: 'edge.surjective', style: { 'target-arrow-fill': 'filled' } },
  { selector: 'node:selected', style: { 'overlay-opacity': 0 } },
  { selector: 'edge:selected', style: { 'overlay-opacity': 0 } },
  {
    selector: 'edge.selected_edge',
    style: {
      width: '6px',
      'line-color': '#007acc',
      'target-arrow-color': '#007acc',
      'source-arrow-color': '#007acc',
    },
  },
  {
    selector: 'node.highlighted',
    style: {
      'overlay-color': '#007acc',
      'overlay-opacity': 0.25,
      'overlay-padding': 6,
    },
  },
  { selector: 'node.selected_node', style: { 'background-color': '#999' } },
];

export const ARG_STYLE = [
  {
    selector: 'node',
    style: {
      content: 'data(label)',
      'text-wrap': 'wrap',
      'text-outline-width': '3px',
      'text-outline-color': '#888',
      'text-valign': 'center',
      color: '#fff',
      width: '50',
      height: '50',
      'border-color': '#000',
      'background-color': '#888',
    },
  },
  { selector: 'node.state', style: {} },
  {
    selector: 'node.bottom_state',
    style: {
      'background-color': '#000',
      'text-outline-color': '#000',
    },
  },
  {
    selector: 'edge',
    style: {
      content: 'data(label)',
      width: '4px',
      'line-style': 'solid',
      'edge-text-rotation': 'none',
      'curve-style': 'bezier',
      color: '#c8c8c8',
      'text-outline-width': '2px',
      'text-outline-color': '#1a1a2e',
      'font-size': '11px',
    },
  },
  { selector: 'edge.transition_join', style: { 'target-arrow-shape': 'triangle-backcurve' } },
  { selector: 'edge.transition_action', style: { 'target-arrow-shape': 'triangle' } },
  {
    selector: 'edge.cover',
    style: { content: '', 'target-arrow-shape': 'triangle', 'line-style': 'dashed' },
  },
  { selector: 'node:selected', style: { 'overlay-opacity': 0.2 } },
  { selector: 'edge:selected', style: { 'overlay-opacity': 0.2 } },
  {
    selector: 'node.highlighted',
    style: {
      'overlay-color': '#007acc',
      'overlay-opacity': 0.3,
      'overlay-padding': 8,
    },
  },
];

export const PROOF_STYLE = [
  {
    selector: 'node',
    style: {
      content: 'data(label)',
      'text-wrap': 'wrap',
      'text-outline-width': '3px',
      'text-outline-color': '#888',
      'text-valign': 'center',
      color: '#fff',
      width: '50',
      height: '50',
      'border-color': '#000',
      'background-color': '#888',
    },
  },
  { selector: 'edge', style: { 'target-arrow-shape': 'triangle', 'curve-style': 'bezier' } },
  {
    selector: 'node.refuted',
    style: {
      'background-color': '#000',
      'text-outline-color': '#000',
    },
  },
  { selector: 'node:selected', style: { 'overlay-opacity': 0.2 } },
  { selector: 'edge:selected', style: { 'overlay-opacity': 0.2 } },
];

function getWindow(win) {
  return win || globalThis.window;
}

function getCytoscape(win) {
  return (win && win.cytoscape) || globalThis.cytoscape;
}

function getDagre(win) {
  return (win && win.cytoscapeDagre) || globalThis.cytoscapeDagre;
}

function copyOptions(defaults, options) {
  return { ...defaults, ...(options || {}) };
}

export class IvyGraph {
  constructor(containerId, style, { win = globalThis.window } = {}) {
    this.containerId = containerId;
    this.win = getWindow(win);
    this.doc = this.win && this.win.document;
    const cytoscape = getCytoscape(this.win);
    try {
      this.cy = cytoscape({
        container: this.doc && this.doc.getElementById(containerId),
        style,
        layout: { name: 'preset' },
        userZoomingEnabled: true,
        userPanningEnabled: true,
        boxSelectionEnabled: false,
        minZoom: 0.2,
        maxZoom: 5,
      });
    } catch (err) {
      const msg = `FATAL: Cytoscape initialization failed for "${containerId}": ${err.message}\n`
        + 'This usually means a stylesheet has an invalid data() mapper for a property that requires a concrete value.';
      console.error(msg, err);
      if (this.win && typeof this.win.alert === 'function') this.win.alert(msg);
      throw err;
    }

    this._cxttapOk = false;
    this.cy.on('cxttap', () => {
      this._cxttapOk = true;
    });

    const dagre = getDagre(this.win);
    if (cytoscape && dagre && typeof cytoscape.use === 'function') {
      cytoscape.use(dagre);
    }

    this._clickCallbacks = {};
    this._rightClickCallbacks = {};
  }

  update(elements, positions) {
    this.cy.elements().remove();
    if (!elements || elements.length === 0) return;

    const toAdd = elements.map((element) => {
      if (element.group === 'nodes') {
        if (!element.data.width) {
          element.data.width = element.data.label ? Math.max(50, element.data.label.length * 8 + 20) : 50;
        }
        if (!element.data.height) element.data.height = 50;
        if (!element.data.shape) element.data.shape = 'ellipse';
        if (!element.data.border_color) element.data.border_color = '#000';
      }
      return element;
    });

    this.cy.add(toAdd);
    this.cy.nodes().forEach((node) => {
      const borderColor = node.data('border_color');
      if (borderColor) node.style('border-color', borderColor);
    });

    if (positions) {
      Object.entries(positions).forEach(([nodeId, position]) => {
        const node = this.cy.getElementById(nodeId);
        if (node.length > 0) node.position(position);
      });
      this.cy.fit(undefined, 30);
    } else {
      this.runLayout();
    }
  }

  runLayout(options) {
    const layoutOptions = copyOptions({
      name: 'dagre',
      rankDir: 'TB',
      nodeSep: 50,
      rankSep: 80,
      edgeSep: 10,
      animate: false,
      fit: true,
      padding: 30,
    }, options);

    try {
      this.cy.layout(layoutOptions).run();
    } catch (err) {
      console.warn('Dagre layout failed, falling back to grid:', err);
      this.cy.layout({ name: 'grid', fit: true, padding: 30 }).run();
    }
  }

  onNodeClick(callback) {
    this.cy.on('tap', 'node', (evt) => {
      callback(evt.target.data(), evt);
    });
  }

  onNodeRightClick(callback) {
    this.cy.on('cxttap', 'node', (evt) => {
      const pos = evt.renderedPosition || evt.target.renderedPosition();
      callback(evt.target.data(), pos, evt);
    });
  }

  onEdgeClick(callback) {
    this.cy.on('tap', 'edge', (evt) => {
      callback(evt.target.data(), evt);
    });
  }

  onEdgeRightClick(callback) {
    this.cy.on('cxttap', 'edge', (evt) => {
      const pos = evt.renderedPosition || evt.target.renderedMidpoint();
      callback(evt.target.data(), pos, evt);
    });
  }

  onBackgroundClick(callback) {
    this.cy.on('tap', (evt) => {
      if (evt.target === this.cy) callback(evt);
    });
  }

  getSelectedNodes() {
    const result = [];
    this.cy.$('node:selected').forEach((node) => {
      result.push(node.data());
    });
    return result;
  }

  highlightNode(id) {
    this.cy.nodes().removeClass('highlighted');
    const node = this.cy.getElementById(id);
    if (node.length > 0) node.addClass('highlighted');
  }

  clearHighlights() {
    this.cy.nodes().removeClass('highlighted');
  }

  centerOnNode(id) {
    const node = this.cy.getElementById(id);
    if (node.length > 0) {
      this.cy.animate({
        center: { eles: node },
        duration: 300,
      });
    }
  }

  fit() {
    this.cy.fit(undefined, 30);
  }

  resize() {
    this.cy.resize();
    this.cy.fit(undefined, 30);
  }

  destroy() {
    if (this.cy) {
      this.cy.destroy();
      this.cy = null;
    }
  }

  healthCheck() {
    const id = this.containerId;
    try {
      if (!this.cy) throw new Error('cy instance is null');
      this.cy.add({
        group: 'nodes',
        data: {
          id: '__healthcheck__',
          label: 'test',
          border_color: '#000',
          shape: 'ellipse',
          width: 10,
          height: 10,
        },
      });
      const testNode = this.cy.getElementById('__healthcheck__');
      if (testNode.length === 0) throw new Error('test node not found after add');
      testNode.style('border-color');
      testNode.remove();

      const handler = () => {};
      this.cy.on('tap', handler);
      this.cy.off('tap', handler);

      this.win[`__ivyGraphHealthy_${id}`] = true;
      return true;
    } catch (err) {
      const msg = `IvyGraph health check FAILED for "${id}": ${err.message}`;
      console.error(msg, err);
      if (this.win && typeof this.win.alert === 'function') this.win.alert(msg);
      this.win[`__ivyGraphHealthy_${id}`] = false;
      return false;
    }
  }
}

export function installLegacyGraphGlobals(win = globalThis.window) {
  if (!win) return;
  if (!win.IvyGraph) win.IvyGraph = IvyGraph;
  if (!win.CONCEPT_STYLE) win.CONCEPT_STYLE = CONCEPT_STYLE;
  if (!win.ARG_STYLE) win.ARG_STYLE = ARG_STYLE;
  if (!win.PROOF_STYLE) win.PROOF_STYLE = PROOF_STYLE;
}

