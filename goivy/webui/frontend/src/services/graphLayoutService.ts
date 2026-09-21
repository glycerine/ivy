import { Graphviz } from '@hpcc-js/wasm-graphviz';
import {
  GraphPositionMap,
} from '../models/uiDataModel.ts';
import {
  selectArgGraphView,
  selectConceptGraphView,
  selectSheet,
} from '../models/uiDataSelectors.ts';

type RawRecord = Record<string, any>;

export interface DotLayoutGraph {
  dot: string;
  idByDotId: Record<string, string>;
}

function rawRecord(value: unknown): RawRecord {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as RawRecord : {};
}

function dotString(value: unknown): string {
  return `"${String(value ?? '').replace(/\\/g, '\\\\').replace(/"/g, '\\"').replace(/\n/g, '\\n')}"`;
}

function elementData(element: RawRecord): RawRecord {
  return rawRecord(element.data);
}

function elementId(element: RawRecord): string {
  const data = elementData(element);
  return String(data.id || data.obj || data.label || '');
}

function elementLabel(element: RawRecord): string {
  const data = elementData(element);
  return String(data.label || data.display_label || data.obj || data.id || '');
}

export function buildDotLayoutGraph(elements: RawRecord[] = []): DotLayoutGraph {
  const nodes = (elements || []).filter((element) => element && element.group === 'nodes');
  const dotIdById: Record<string, string> = {};
  const idByDotId: Record<string, string> = {};
  nodes.forEach((element, index) => {
    const id = elementId(element);
    if (!id) return;
    const dotId = `n${index}`;
    dotIdById[id] = dotId;
    idByDotId[dotId] = id;
  });

  const lines = [
    'digraph ivy_graph {',
    '  graph [rankdir=TB, margin=0.1, nodesep=0.5, ranksep=0.8];',
    '  node [shape=ellipse, margin="0.08,0.04"];',
    '  edge [];',
  ];

  for (const element of nodes) {
    const id = elementId(element);
    const dotId = dotIdById[id];
    if (!dotId) continue;
    lines.push(`  ${dotId} [label=${dotString(elementLabel(element))}];`);
  }

  for (const element of elements || []) {
    if (!element || element.group !== 'edges') continue;
    const data = elementData(element);
    const source = dotIdById[String(data.source || '')];
    const target = dotIdById[String(data.target || '')];
    if (!source || !target) continue;
    lines.push(`  ${source} -> ${target};`);
  }

  lines.push('}');
  return { dot: lines.join('\n'), idByDotId };
}

export function parseGraphvizPlainPositions(plain: string, idByDotId: Record<string, string>, {
  scale = 96,
} = {}): GraphPositionMap {
  const positions: GraphPositionMap = {};
  let graphHeight = 0;
  for (const line of String(plain || '').split(/\r?\n/)) {
    const parts = line.trim().split(/\s+/);
    if (parts[0] === 'graph') {
      graphHeight = Number(parts[3]) || 0;
      continue;
    }
    if (parts[0] !== 'node') continue;
    const id = idByDotId[parts[1]];
    const x = Number(parts[2]);
    const y = Number(parts[3]);
    if (!id || !Number.isFinite(x) || !Number.isFinite(y)) continue;
    positions[id] = {
      x: x * scale,
      y: (graphHeight > 0 ? graphHeight - y : y) * scale,
    };
  }
  return positions;
}

async function defaultGraphviz() {
  return Graphviz.load();
}

export async function computeGraphvizDotPositions(elements: RawRecord[] = [], {
  graphviz = null,
  scale = 96,
} = {}): Promise<GraphPositionMap> {
  const layoutGraph = buildDotLayoutGraph(elements);
  if (Object.keys(layoutGraph.idByDotId).length === 0) return {};
  const engine = graphviz || await defaultGraphviz();
  const plain = await engine.layout(layoutGraph.dot, 'plain', 'dot');
  return parseGraphvizPlainPositions(plain, layoutGraph.idByDotId, { scale });
}

export async function relayoutGraphWithDot(app, graphKind: 'arg' | 'concept' = 'concept', {
  sheetId = app && (app.activeSheetId || 'sheet-1'),
  graphviz = null,
  scale = 96,
} = {}): Promise<GraphPositionMap | null> {
  const sheet = selectSheet(app && app.uiDataModel, sheetId);
  if (!sheet || !app || !app.uiDataStore || typeof app.uiDataStore.setGraphPositions !== 'function') return null;
  const view = graphKind === 'arg' ? selectArgGraphView(sheet) : selectConceptGraphView(sheet);
  const positions = await computeGraphvizDotPositions(view.elements, { graphviz, scale });
  app.uiDataStore.setGraphPositions(sheetId, graphKind, positions);
  return positions;
}
