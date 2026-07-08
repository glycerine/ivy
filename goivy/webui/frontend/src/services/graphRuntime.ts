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
    selector: 'node.subgraph_box',
    style: {
      content: 'data(label)',
      shape: 'roundrectangle',
      'background-opacity': 0,
      'border-width': '2px',
      'border-style': 'dashed',
      'border-color': '#777',
      padding: '18px',
      events: 'no',
      'z-index': 0,
    },
  },
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
    style: { 'border-width': '0px' },
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
      'text-wrap': 'wrap',
    },
  },
  {
    selector: 'edge[label]',
    style: {
      content: 'data(label)',
      color: '#000',
      'text-outline-width': '3px',
      'text-outline-color': '#ffd400',
    },
  },
  { selector: 'edge[text_max_width]', style: { 'text-max-width': 'data(text_max_width)' } },
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
  { selector: 'edge.layout_only', style: { display: 'none' } },
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
    selector: 'node.safe_state',
    style: {
      'background-color': '#2f8f46',
      'text-outline-color': '#1f6b33',
    },
  },
  {
    selector: 'node.marked_state',
    style: {
      'background-color': '#b73535',
      'text-outline-color': '#7f1f1f',
    },
  },
  {
    selector: 'edge',
    style: {
      width: '4px',
      'line-style': 'solid',
      'edge-text-rotation': 'none',
      'curve-style': 'bezier',
      'text-wrap': 'wrap',
      color: '#c8c8c8',
      'text-outline-width': '2px',
      'text-outline-color': '#1a1a2e',
      'font-size': '11px',
    },
  },
  { selector: 'edge[label]', style: { content: 'data(label)' } },
  { selector: 'edge[text_max_width]', style: { 'text-max-width': 'data(text_max_width)' } },
  { selector: 'edge.transition_join', style: { 'target-arrow-shape': 'triangle-backcurve' } },
  { selector: 'edge.transition_action', style: { 'target-arrow-shape': 'triangle' } },
  { selector: 'edge.layout_only', style: { display: 'none' } },
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

function getWindow(win: any) {
  return win || globalThis.window;
}

function getCytoscape(win: any) {
  return (win && win.cytoscape) || globalThis.cytoscape;
}

function getDagre(win: any) {
  return (win && win.cytoscapeDagre) || globalThis.cytoscapeDagre;
}

function copyOptions(defaults: any, options: any) {
  return { ...defaults, ...(options || {}) };
}

function positionValue(position: any) {
  return position
    && typeof position.x === 'number'
    && typeof position.y === 'number'
    && Number.isFinite(position.x)
    && Number.isFinite(position.y)
    ? { x: position.x, y: position.y }
    : null;
}

function nodeId(element: any): string {
  const data = (element && element.data) || {};
  return String(data.id || data.obj || data.label || '');
}

function classesArray(element: any): string[] {
  const classes = element && typeof element.classes === 'string' ? element.classes : '';
  return classes.split(/\s+/).filter(Boolean);
}

function hasClass(element: any, className: string): boolean {
  return classesArray(element).includes(className);
}

function classStringWith(classes: string | string[], className: string): string {
  const current = Array.isArray(classes) ? classes : String(classes || '').split(/\s+/);
  return Array.from(new Set([...current.filter(Boolean), className])).join(' ');
}

function withClass(element: any, className: string): string {
  return classStringWith(classesArray(element), className);
}

function isSubgraphShape(element: any): boolean {
  return element && element.group === 'shapes' && hasClass(element, 'subgraphs');
}

function isSubgraphBoxNode(element: any): boolean {
  return element && element.group === 'nodes' && hasClass(element, 'subgraph_box');
}

function isGenericShapeNode(element: any): boolean {
  return element && element.group === 'nodes' && hasClass(element, 'generic_shape');
}

function shapeCluster(element: any): string {
  const data = (element && element.data) || {};
  const explicit = String(data.cluster || '');
  if (explicit) return explicit;
  const obj = String(data.obj || '');
  return obj.startsWith('cluster_') ? obj.slice('cluster_'.length) : obj;
}

function shapeMembers(element: any): Set<string> {
  const members = new Set<string>();
  const raw = element && element.data && Array.isArray(element.data.nodes) ? element.data.nodes : [];
  raw.forEach((value) => {
    if (typeof value === 'string' && value) members.add(value);
  });
  return members;
}

function shapeBounds(element: any) {
  const coords = element && element.data && Array.isArray(element.data.coords) ? element.data.coords : [];
  const points = coords.map(positionValue).filter(Boolean);
  if (points.length === 0) return null;
  const xs = points.map((point) => point.x);
  const ys = points.map((point) => point.y);
  const minX = Math.min(...xs);
  const maxX = Math.max(...xs);
  const minY = Math.min(...ys);
  const maxY = Math.max(...ys);
  return {
    position: { x: (minX + maxX) / 2, y: (minY + maxY) / 2 },
    width: Math.max(1, maxX - minX),
    height: Math.max(1, maxY - minY),
  };
}

function nodeMatchesSubgraphShape(node: any, cluster: string, members: Set<string>): boolean {
  if (!node || node.group !== 'nodes' || isSubgraphBoxNode(node)) return false;
  const data = node.data || {};
  const id = String(data.id || '');
  const obj = String(data.obj || '');
  if (members.size > 0) {
    return members.has(id) || members.has(obj);
  }
  return !!cluster && String(data.cluster || '') === cluster;
}

function expandSubgraphShapes(elements: any[]): any[] {
  const nodesAndEdges = [];
  const boxes = [];
  for (const element of elements || []) {
    if (isSubgraphShape(element)) {
      boxes.push(element);
    } else {
      nodesAndEdges.push({
        ...element,
        data: { ...((element && element.data) || {}) },
      });
    }
  }
  if (boxes.length === 0) return nodesAndEdges;

  const parentNodes = [];
  for (const [index, box] of boxes.entries()) {
    const data = box.data || {};
    const cluster = shapeCluster(box);
    const members = shapeMembers(box);
    const childNodes = nodesAndEdges.filter((node) => nodeMatchesSubgraphShape(node, cluster, members));
    const bounds = shapeBounds(box);
    if (childNodes.length === 0 && !bounds) continue;
    const parentId = String(data.id || data.obj || `subgraph_box_${index}`);
    const parentData: any = {
      id: parentId,
      obj: data.obj || parentId,
      label: data.label || '',
      shape: data.shape || 'rectangle',
      cluster,
    };
    if (bounds) {
      parentData.width = bounds.width;
      parentData.height = bounds.height;
      parentData.generic_shape = childNodes.length === 0;
    }
    let parentClasses = withClass(box, 'subgraph_box');
    if (bounds && childNodes.length === 0) {
      parentClasses = classStringWith(parentClasses, 'generic_shape');
    }
    const parentNode: any = {
      group: 'nodes',
      data: parentData,
      classes: parentClasses,
      locked: !!box.locked,
    };
    if (bounds) {
      parentNode.position = bounds.position;
    }
    parentNodes.push(parentNode);
    for (const child of childNodes) {
      child.data.parent = parentId;
    }
  }
  return [...parentNodes, ...nodesAndEdges];
}

function edgeLayoutEndpoint(value: any): string {
  return typeof value === 'string' && value ? value : '';
}

function layoutEdgeClasses(element: any): string {
  return Array.from(new Set([...classesArray(element).filter((klass) => klass !== 'layout_ignored'), 'layout_only'])).join(' ');
}

function expandLayoutEdges(elements: any[]): any[] {
  const out = [];
  const layoutOnly = [];
  for (const element of elements || []) {
    const next = {
      ...element,
      data: { ...((element && element.data) || {}) },
    };
    if (next.group === 'edges') {
      const data = next.data || {};
      const layoutSource = edgeLayoutEndpoint(data.layout_source);
      const layoutTarget = edgeLayoutEndpoint(data.layout_target);
      const displayedSource = edgeLayoutEndpoint(data.source);
      const displayedTarget = edgeLayoutEndpoint(data.target);
      const hasLayoutProxy = !!layoutSource
        && !!layoutTarget
        && (layoutSource !== displayedSource || layoutTarget !== displayedTarget);
      const unconstrained = data.layout_constraint === false;
      if (hasLayoutProxy || unconstrained) {
        next.classes = withClass(next, 'layout_ignored');
      }
      if (hasLayoutProxy && !unconstrained) {
        const id = String(data.id || data.obj || `edge_${layoutOnly.length}`);
        layoutOnly.push({
          group: 'edges',
          data: {
            id: `layout_${id}`,
            source: layoutSource,
            target: layoutTarget,
            label: '',
            layout_for: id,
          },
          classes: layoutEdgeClasses(element),
        });
      }
    }
    out.push(next);
  }
  return [...out, ...layoutOnly];
}

function clonePositionMap(positions: any = {}) {
  const out = {};
  for (const [id, position] of Object.entries(positions || {})) {
    const value = positionValue(position);
    if (id && value) out[id] = value;
  }
  return out;
}

function mergePositionMaps(...maps: any[]) {
  return Object.assign({}, ...maps.map((map) => clonePositionMap(map)));
}

function elementPositionMap(elements: any[]) {
  const out = {};
  for (const element of elements || []) {
    if (!element || element.group !== 'nodes') continue;
    const id = nodeId(element);
    const position = positionValue(element.position);
    if (id && position) out[id] = position;
  }
  return out;
}

function graphNodeIds(elements: any[]): string[] {
  return (elements || [])
    .filter((element) => element && element.group === 'nodes' && (!isSubgraphBoxNode(element) || isGenericShapeNode(element)))
    .map(nodeId)
    .filter(Boolean);
}

function averageKnownNeighborPosition(node: string, elements: any[], positions: any) {
  const neighbors = [];
  for (const element of elements || []) {
    if (!element || element.group !== 'edges') continue;
    const data = element.data || {};
    if (data.source === node && positions[data.target]) neighbors.push(positions[data.target]);
    if (data.target === node && positions[data.source]) neighbors.push(positions[data.source]);
  }
  if (neighbors.length === 0) return null;
  const sum = neighbors.reduce((acc, position) => ({
    x: acc.x + position.x,
    y: acc.y + position.y,
  }), { x: 0, y: 0 });
  return { x: sum.x / neighbors.length, y: sum.y / neighbors.length };
}

function fallbackPosition(positions: any, index: number) {
  const known = Object.values(positions || {}).filter(Boolean) as any[];
  if (known.length === 0) return null;
  const maxX = Math.max(...known.map((position) => position.x));
  const minY = Math.min(...known.map((position) => position.y));
  const col = index % 4;
  const row = Math.floor(index / 4);
  return { x: maxX + 100 + col * 90, y: minY + row * 80 };
}

function fillMissingNodePositions(elements: any[], positions: any) {
  const out = clonePositionMap(positions);
  if (Object.keys(out).length === 0) return out;
  let missingIndex = 0;
  for (const id of graphNodeIds(elements)) {
    if (out[id]) continue;
    const neighbor = averageKnownNeighborPosition(id, elements, out);
    const fallback = fallbackPosition(out, missingIndex);
    const position = neighbor || fallback;
    if (position) {
      out[id] = {
        x: position.x + 40 + (missingIndex % 3) * 30,
        y: position.y + 40 + Math.floor(missingIndex / 3) * 30,
      };
    }
    missingIndex += 1;
  }
  return out;
}

function allNodesHavePositions(elements: any[], positions: any): boolean {
  const ids = graphNodeIds(elements);
  return ids.length > 0 && ids.every((id) => !!positions[id]);
}

export class IvyGraph {
  [key: string]: any;

  constructor(containerId: string, style: any, { win = globalThis.window }: { win?: any } = {}) {
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
    const currentPositions = this.getNodePositions();
    this.cy.elements().remove();
    if (!elements || elements.length === 0) return;

    const toAdd = expandLayoutEdges(expandSubgraphShapes(elements)).map((element) => {
      const next = {
        ...element,
        data: { ...(element.data || {}) },
      };
      if (element.group === 'nodes') {
        if (!next.data.width) {
          next.data.width = next.data.label ? Math.max(50, next.data.label.length * 8 + 20) : 50;
        }
        if (!next.data.height) next.data.height = 50;
        if (!next.data.shape) next.data.shape = 'ellipse';
        if (!next.data.border_color) next.data.border_color = '#000';
      }
      return next;
    });

    const mergedPositions = fillMissingNodePositions(toAdd, mergePositionMaps(elementPositionMap(toAdd), currentPositions, positions));
    const usePresetPositions = allNodesHavePositions(toAdd, mergedPositions);
    if (usePresetPositions) {
      for (const element of toAdd) {
        if (element.group !== 'nodes') continue;
        const id = nodeId(element);
        if (mergedPositions[id]) element.position = mergedPositions[id];
      }
    }

    this.cy.add(toAdd);
    this.cy.nodes().forEach((node) => {
      const borderColor = node.data('border_color');
      if (borderColor) node.style('border-color', borderColor);
    });
    this.cy.edges().forEach((edge) => {
      const lineColor = edge.data('line_color') || edge.data('color');
      if (!lineColor) return;
      edge.style('line-color', lineColor);
      edge.style('target-arrow-color', lineColor);
      edge.style('source-arrow-color', lineColor);
    });

    if (usePresetPositions) {
      Object.entries(mergedPositions).forEach(([id, position]) => {
        const node = this.cy.getElementById(id);
        if (node.length > 0) node.position(position);
      });
    } else {
      this.runLayout();
    }
    return this.getNodePositions();
  }

  runLayout(options: any = undefined) {
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
      const elements = typeof this.cy.elements === 'function' ? this.cy.elements() : null;
      if (elements && typeof elements.not === 'function' && typeof elements.layout === 'function') {
        elements.not('.layout_ignored').layout(layoutOptions).run();
      } else {
        this.cy.layout(layoutOptions).run();
      }
    } catch (err) {
      console.warn('Dagre layout failed, falling back to grid:', err);
      this.cy.layout({ name: 'grid', fit: true, padding: 30 }).run();
    }
  }

  getNodePositions() {
    const out = {};
    if (!this.cy || typeof this.cy.nodes !== 'function') return out;
    this.cy.nodes().forEach((node) => {
      const id = node.id ? node.id() : node.data && node.data('id');
      const position = node.position ? node.position() : null;
      const value = positionValue(position);
      if (id && value) out[id] = value;
    });
    return out;
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

  onNodePositionChange(callback) {
    this.cy.on('free', 'node', (evt) => {
      callback(this.getNodePositions(), evt);
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
    if (node.length > 0) {
      node.addClass('highlighted');
      this.centerOnNode(id);
    }
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

  resetView() {
    if (this.cy && typeof this.cy.reset === 'function') {
      this.cy.reset();
    }
    this.fit();
  }

  reset() {
    this.resetView();
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

export function installGraphGlobalsForCompatibility(win = globalThis.window) {
  if (!win) return;
  if (!win.IvyGraph) win.IvyGraph = IvyGraph;
  if (!win.CONCEPT_STYLE) win.CONCEPT_STYLE = CONCEPT_STYLE;
  if (!win.ARG_STYLE) win.ARG_STYLE = ARG_STYLE;
  if (!win.PROOF_STYLE) win.PROOF_STYLE = PROOF_STYLE;
}
