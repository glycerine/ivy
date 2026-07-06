import {
  ConceptSnapshot,
  CyElement,
  FactSelection,
  GraphPositionMap,
  GraphSelection,
  SheetModel,
  UIDataModel,
  conceptEdgeSelection,
  conceptNodeSelection,
  graphSelectionAliases,
} from './uiDataModel.ts';

export const EDGE_DISPLAY_CLASSES = ['all_to_all', 'edge_unknown', 'none_to_none', 'transitive'];
export const EDGE_TO_LABEL_CLASS = {
  all_to_all: 'node_necessarily',
  edge_unknown: 'node_maybe',
  none_to_none: 'node_necessarily_not',
};

const LABEL_PREFIXES = {
  node_necessarily: '',
  node_maybe: '?',
  node_necessarily_not: '\u00AC',
};

type RawRecord = Record<string, any>;

function rawRecord(value: unknown): RawRecord {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as RawRecord : {};
}

function classesArray(classes: unknown): string[] {
  return typeof classes === 'string' ? classes.split(/\s+/).filter(Boolean) : [];
}

function setClasses(element: RawRecord, classes: string[]): void {
  element.classes = Array.from(new Set(classes.filter(Boolean))).join(' ');
}

function addClass(element: RawRecord, className: string): void {
  const classes = classesArray(element.classes);
  if (!classes.includes(className)) classes.push(className);
  setClasses(element, classes);
}

function cloneCyElement(element: CyElement): RawRecord {
  const out: RawRecord = {
    group: element.group,
    data: element.data.toCytoscapeData(),
  };
  if (element.classes) out.classes = element.classes;
  if (element.locked) out.locked = element.locked;
  if (element.position) out.position = { x: element.position.x, y: element.position.y };
  return out;
}

function cyElementsToDefinitions(elements: CyElement[]): RawRecord[] {
  return elements.map(cloneCyElement);
}

function clonePositionMap(positions: GraphPositionMap | null | undefined): GraphPositionMap | null {
  const out: GraphPositionMap = {};
  for (const [id, position] of Object.entries(positions || {})) {
    if (!position) continue;
    out[id] = { x: position.x, y: position.y };
  }
  return Object.keys(out).length > 0 ? out : null;
}

function bareRelationName(name: string): string {
  const idx = String(name || '').indexOf('(');
  return idx >= 0 ? name.slice(0, idx) : name;
}

function elementId(element: RawRecord): string {
  const data = rawRecord(element.data);
  return String(data.id || data.obj || data.label || '');
}

function optionVisible(snapshot: ConceptSnapshot, name: string, displayClass: string): boolean {
  if (!name) return false;
  const base = bareRelationName(name);
  return snapshot.displayCheckboxes.edgeVisible(name, displayClass)
    || snapshot.displayCheckboxes.edgeVisible(base, displayClass);
}

function labelVisible(snapshot: ConceptSnapshot, name: string, displayClass: string): boolean {
  if (!name) return false;
  const base = bareRelationName(name);
  return snapshot.displayCheckboxes.nodeLabelVisible(name, displayClass)
    || snapshot.displayCheckboxes.nodeLabelVisible(base, displayClass);
}

function relationColor(snapshot: ConceptSnapshot, name: string): string {
  if (!name) return '';
  const base = bareRelationName(name);
  return snapshot.relationColors[name] || snapshot.relationColors[base] || '';
}

export function displayConceptName(name: string): string {
  if (typeof name === 'string' && name.charAt(0) === '=') {
    const body = name.slice(1);
    const idx = body.lastIndexOf(':');
    if (idx > 0) return `=${body.slice(0, idx)}`;
  }
  return name;
}

function applyConceptNodeLabel(snapshot: ConceptSnapshot, element: RawRecord): void {
  const data = rawRecord(element.data);
  const nodeId = String(data.obj || '');
  if (!nodeId) return;
  const sortName = String(data.cluster || data.sort || nodeId);
  const topLabel = String(data.display_label || sortName);
  const labelParts = [topLabel];

  for (const labelName of snapshot.nodeLabels) {
    const baseLabelName = bareRelationName(labelName);
    const labelSort = snapshot.labelSorts[labelName] || snapshot.labelSorts[baseLabelName];
    if (labelSort && labelSort !== sortName) continue;

    const necKey = `node_label|node_necessarily|${nodeId}|${baseLabelName}`;
    const necNotKey = `node_label|node_necessarily_not|${nodeId}|${baseLabelName}`;
    let displayClass = 'node_maybe';
    if (snapshot.abstractValue[necKey]) {
      displayClass = 'node_necessarily';
    } else if (snapshot.abstractValue[necNotKey]) {
      displayClass = 'node_necessarily_not';
    }

    if (!labelVisible(snapshot, labelName, displayClass)) continue;
    labelParts.push(`${LABEL_PREFIXES[displayClass]}${displayConceptName(baseLabelName)}`);
  }

  data.label = labelParts.join('\n');
  data.height = Math.max(Number(data.height) || 0, 30 + labelParts.length * 20, 50);
}

function conceptEdgeVisible(snapshot: ConceptSnapshot, element: RawRecord): boolean {
  const data = rawRecord(element.data);
  const obj = String(data.obj || '');
  const label = String(data.label || '');
  const classes = classesArray(element.classes);
  const names = [obj, label, bareRelationName(obj), bareRelationName(label)].filter(Boolean);
  for (const className of classes) {
    if (!EDGE_DISPLAY_CLASSES.includes(className)) continue;
    if (names.some((name) => optionVisible(snapshot, name, className))) return true;
  }
  return false;
}

export function selectSheet(model: UIDataModel | null | undefined, sheetId?: string): SheetModel | null {
  if (!model) return null;
  return model.sheets[sheetId || model.activeSheetId] || null;
}

export function selectArgGraphView(sheet: SheetModel | null | undefined) {
  const elements = sheet && sheet.arg ? cyElementsToDefinitions(sheet.arg.render.elements) : [];
  return {
    elements,
    positions: clonePositionMap(sheet && sheet.argPositions),
    selectedNodeId: sheet ? sheet.selectedArgNode : null,
  };
}

export function selectConceptGraphView(sheet: SheetModel | null | undefined) {
  const snapshot = sheet && sheet.concept;
  const elements = snapshot ? cyElementsToDefinitions(snapshot.render.elements) : [];
  const selectedAliases = new Set<string>();
  for (const selection of sheet ? sheet.conceptSelections : []) {
    for (const alias of graphSelectionAliases(selection)) selectedAliases.add(alias);
  }
  const edgeVisibilityById: Record<string, boolean> = {};

  for (const element of elements) {
    const data = rawRecord(element.data);
    if (element.group === 'nodes') {
      if (snapshot) applyConceptNodeLabel(snapshot, element);
      const selection = conceptNodeSelection(data);
      if (graphSelectionAliases(selection).some((alias) => selectedAliases.has(alias))) addClass(element, 'selected_node');
    } else if (element.group === 'edges') {
      const selection = conceptEdgeSelection(data);
      if (graphSelectionAliases(selection).some((alias) => selectedAliases.has(alias))) addClass(element, 'selected_edge');
      const visible = snapshot ? conceptEdgeVisible(snapshot, element) : true;
      const id = elementId(element);
      if (id) edgeVisibilityById[id] = visible;
      for (const alias of graphSelectionAliases(selection)) edgeVisibilityById[alias] = visible;
    }
  }

  return {
    elements,
    positions: clonePositionMap(sheet && sheet.conceptPositions),
    edgeVisibilityById,
  };
}

export function toggleChecked(snapshot: ConceptSnapshot | null | undefined, name: string, displayClass: string): boolean {
  if (!snapshot) return false;
  if (optionVisible(snapshot, name, displayClass)) return true;
  const labelClass = EDGE_TO_LABEL_CLASS[displayClass];
  return !!(labelClass && labelVisible(snapshot, name, labelClass));
}

export function selectStateCheckboxRows(sheet: SheetModel | null | undefined) {
  const snapshot = sheet && sheet.concept;
  if (!snapshot) return [];
  const names = (snapshot.relations.length > 0 ? snapshot.relations : snapshot.edges).slice().sort();
  return names.map((name) => {
    const row: Record<string, any> = {
      name,
      checked: {
      all_to_all: toggleChecked(snapshot, name, 'all_to_all'),
      edge_unknown: toggleChecked(snapshot, name, 'edge_unknown'),
      none_to_none: toggleChecked(snapshot, name, 'none_to_none'),
      transitive: toggleChecked(snapshot, name, 'transitive'),
      },
    };
    const color = relationColor(snapshot, name);
    if (color) row.color = color;
    return row;
  });
}

export function selectStateToggles(sheet: SheetModel | null | undefined): Record<string, boolean> {
  const toggles: Record<string, boolean> = {};
  for (const row of selectStateCheckboxRows(sheet)) {
    for (const displayClass of EDGE_DISPLAY_CLASSES) {
      toggles[`${row.name}|${displayClass}`] = !!row.checked[displayClass];
    }
  }
  return toggles;
}

export function selectConceptSelections(sheet: SheetModel | null | undefined): GraphSelection[] {
  return sheet ? sheet.conceptSelections.map((selection) => ({ ...selection })) : [];
}

export function selectPrimaryConceptNodeSelection(model: UIDataModel | null | undefined, sheetId?: string): GraphSelection | null {
  const sheet = selectSheet(model, sheetId);
  if (!sheet) return null;
  for (let i = sheet.conceptSelections.length - 1; i >= 0; i -= 1) {
    if (sheet.conceptSelections[i].kind === 'node') return { ...sheet.conceptSelections[i] };
  }
  return null;
}

export function selectConceptNodeSelected(model: UIDataModel | null | undefined, sheetId: string | undefined, node: unknown): boolean {
  const sheet = selectSheet(model, sheetId);
  const selection = conceptNodeSelection(node);
  if (!sheet || !selection) return false;
  const aliases = new Set(graphSelectionAliases(selection));
  return sheet.conceptSelections.some((candidate) =>
    candidate.kind === 'node' && graphSelectionAliases(candidate).some((alias) => aliases.has(alias)));
}

export function selectConstraintFacts(sheet: SheetModel | null | undefined): FactSelection[] {
  return sheet && sheet.concept ? sheet.concept.facts : [];
}

export function selectStateLabel(sheet: SheetModel | null | undefined): string | null {
  if (!sheet) return null;
  if (sheet.concept && sheet.concept.stateLabel) return sheet.concept.stateLabel;
  return sheet.selectedArgNode;
}

export function selectConceptSnapshot(model: UIDataModel | null | undefined, sheetId?: string): ConceptSnapshot | null {
  const sheet = selectSheet(model, sheetId);
  return sheet ? sheet.concept : null;
}

export function selectMaterializableEdges(model: UIDataModel | null | undefined, sheetId: string | undefined, source: string, target: string): string[] {
  const snapshot = selectConceptSnapshot(model, sheetId);
  if (!snapshot) return [];
  let relations = Object.keys(snapshot.edgeSorts).filter((rel) => {
    const sorts = snapshot.edgeSorts[rel] || [];
    return sorts.length >= 2 && sorts[0] === source && sorts[1] === target;
  });
  if (relations.length === 0) relations = snapshot.edges.slice();
  return relations;
}
