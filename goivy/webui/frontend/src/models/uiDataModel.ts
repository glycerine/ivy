export type JsonPrimitive = string | number | boolean | null;
export type JsonValue = JsonPrimitive | JsonObject | JsonValue[];
export interface JsonObject {
  [key: string]: JsonValue | undefined;
}

type RawRecord = Record<string, unknown>;
type Constructor<T> = new (raw?: unknown) => T;

function isRecord(value: unknown): value is RawRecord {
  return !!value && typeof value === 'object' && !Array.isArray(value);
}

function rawRecord(value: unknown): RawRecord {
  return isRecord(value) ? value : {};
}

function hasOwn(obj: RawRecord, key: string): boolean {
  return Object.prototype.hasOwnProperty.call(obj, key);
}

function pick(raw: unknown, ...keys: string[]): unknown {
  const obj = rawRecord(raw);
  for (const key of keys) {
    if (hasOwn(obj, key)) return obj[key];
  }
  return undefined;
}

function stringValue(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value : fallback;
}

function numberValue(value: unknown, fallback = 0): number {
  return typeof value === 'number' && Number.isFinite(value) ? value : fallback;
}

function intValue(value: unknown, fallback = 0): number {
  return Math.trunc(numberValue(value, fallback));
}

function boolValue(value: unknown, fallback = false): boolean {
  return typeof value === 'boolean' ? value : fallback;
}

function stringArray(value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === 'string') : [];
}

function rawArray(value: unknown): unknown[] {
  return Array.isArray(value) ? value : [];
}

function typedArray<T>(value: unknown, Type: Constructor<T>): T[] {
  return rawArray(value).map((item) => new Type(item));
}

function stringArrayMap(value: unknown): Record<string, string[]> {
  const obj = rawRecord(value);
  const out: Record<string, string[]> = {};
  for (const key of Object.keys(obj)) {
    out[key] = stringArray(obj[key]);
  }
  return out;
}

function stringMap(value: unknown): Record<string, string> {
  const obj = rawRecord(value);
  const out: Record<string, string> = {};
  for (const key of Object.keys(obj)) {
    out[key] = stringValue(obj[key]);
  }
  return out;
}

function boolMap(value: unknown): Record<string, boolean> {
  const obj = rawRecord(value);
  const out: Record<string, boolean> = {};
  for (const key of Object.keys(obj)) {
    out[key] = boolValue(obj[key]);
  }
  return out;
}

function cloneRecord(value: unknown): RawRecord {
  return { ...rawRecord(value) };
}

function optionalString(value: unknown): string | undefined {
  return typeof value === 'string' && value !== '' ? value : undefined;
}

function optionalNumber(value: unknown): number | undefined {
  return typeof value === 'number' && Number.isFinite(value) ? value : undefined;
}

function optionalBoolean(value: unknown): boolean | undefined {
  return typeof value === 'boolean' ? value : undefined;
}

function cloneJsonish(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(cloneJsonish);
  if (isRecord(value)) {
    const out: RawRecord = {};
    for (const key of Object.keys(value)) out[key] = cloneJsonish(value[key]);
    return out;
  }
  return value;
}

export class RawBackedModel<TRaw = unknown> {
  readonly raw: TRaw;

  constructor(raw?: TRaw) {
    this.raw = (raw ?? {}) as TRaw;
  }
}

export interface StateJSON extends RawRecord {
  id?: number;
  ID?: number;
  clauses?: unknown;
  Clauses?: unknown;
  safe?: boolean;
  Safe?: boolean;
  label?: string;
  Label?: string;
}

export class State extends RawBackedModel<unknown> {
  readonly id: number;
  readonly clauses: unknown;
  readonly safe: boolean | undefined;
  readonly label: string;
  readonly expr: unknown;
  readonly universe: unknown;

  constructor(raw: unknown = {}) {
    super(raw);
    this.id = intValue(pick(raw, 'id', 'ID'), -1);
    this.clauses = pick(raw, 'clauses', 'Clauses');
    const safe = pick(raw, 'safe', 'Safe');
    this.safe = typeof safe === 'boolean' ? safe : undefined;
    this.label = stringValue(pick(raw, 'label', 'Label'));
    this.expr = pick(raw, 'expr', 'Expr');
    this.universe = pick(raw, 'universe', 'Universe');
  }
}

export class AnalysisTransition extends RawBackedModel<unknown> {
  readonly sourceId: number;
  readonly targetId: number;
  readonly label: string;
  readonly operation: unknown;
  readonly pre: unknown;
  readonly post: unknown;

  constructor(raw: unknown = {}) {
    super(raw);
    this.sourceId = intValue(pick(raw, 'source_id', 'sourceId', 'SourceID'), -1);
    this.targetId = intValue(pick(raw, 'target_id', 'targetId', 'TargetID'), -1);
    this.label = stringValue(pick(raw, 'label', 'Label'));
    this.operation = pick(raw, 'operation', 'op', 'Op');
    this.pre = pick(raw, 'pre', 'Pre');
    this.post = pick(raw, 'post', 'Post');
  }
}

export class ARGCover extends RawBackedModel<unknown> {
  readonly coveredId: number;
  readonly coveringId: number;

  constructor(raw: unknown = {}) {
    super(raw);
    this.coveredId = intValue(pick(raw, 'covered_id', 'coveredId', 'CoveredID'), -1);
    this.coveringId = intValue(pick(raw, 'covering_id', 'coveringId', 'CoveringID'), -1);
  }
}

// AnalysisGraph is NOT populated from the current wire format (Go sends lightweight
// AnalysisGraphState render intermediates, not the full graph). These classes are
// reserved for future CTI inspection or a full-graph endpoint. For current GUI work
// use ARGSnapshot.analysisGraphState (ARGNode/ARGTransition/ARGCover).
export class AnalysisGraph extends RawBackedModel<unknown> {
  readonly states: State[];
  readonly transitions: AnalysisTransition[];
  readonly covering: ARGCover[];
  readonly pvars: unknown[];

  constructor(raw: unknown = {}) {
    super(raw);
    this.states = typedArray(pick(raw, 'states', 'States'), State);
    this.transitions = typedArray(pick(raw, 'transitions', 'Transitions'), AnalysisTransition);
    this.covering = typedArray(pick(raw, 'covering', 'Covering'), ARGCover);
    this.pvars = rawArray(pick(raw, 'pvars', 'PVars'));
  }
}

export class ARGNode extends RawBackedModel<unknown> {
  readonly id: number;
  readonly label: string;
  readonly isBottom: boolean;
  readonly isSafe: boolean;
  readonly isMarked: boolean;
  readonly info: string;

  constructor(raw: unknown = {}) {
    super(raw);
    this.id = intValue(pick(raw, 'id', 'ID'), -1);
    this.label = stringValue(pick(raw, 'label', 'Label'));
    this.isBottom = boolValue(pick(raw, 'is_bottom', 'isBottom', 'IsBottom'));
    this.isSafe = boolValue(pick(raw, 'is_safe', 'isSafe', 'IsSafe'));
    this.isMarked = boolValue(pick(raw, 'is_marked', 'isMarked', 'IsMarked'));
    this.info = stringValue(pick(raw, 'info', 'Info'));
  }
}

// FullARGNode is a superset of ARGNode — sent when ?full=true is requested.
// clauses: clean formula string (ToOpenFormula().String()).
// actionName: the action that produced this state ("" for initial state).
// universe: sort → concrete element strings from BMC model (null if no BMC run).
export class FullARGNode extends RawBackedModel<unknown> {
  readonly id: number;
  readonly label: string;
  readonly isBottom: boolean;
  readonly isSafe: boolean;
  readonly isMarked: boolean;
  readonly info: string;
  readonly clauses: string;
  readonly actionName: string;
  readonly universe: Record<string, string[]> | null;

  constructor(raw: unknown = {}) {
    super(raw);
    this.id = intValue(pick(raw, 'id', 'ID'), -1);
    this.label = stringValue(pick(raw, 'label', 'Label'));
    this.isBottom = boolValue(pick(raw, 'is_bottom', 'isBottom', 'IsBottom'));
    this.isSafe = boolValue(pick(raw, 'is_safe', 'isSafe', 'IsSafe'));
    this.isMarked = boolValue(pick(raw, 'is_marked', 'isMarked', 'IsMarked'));
    this.info = stringValue(pick(raw, 'info', 'Info'));
    this.clauses = stringValue(pick(raw, 'clauses', 'Clauses'));
    this.actionName = stringValue(pick(raw, 'action_name', 'actionName', 'ActionName'));
    const u = pick(raw, 'universe', 'Universe');
    this.universe = u != null ? stringArrayMap(u) : null;
  }
}

export class ARGTransition extends RawBackedModel<unknown> {
  readonly sourceId: number;
  readonly targetId: number;
  readonly label: string;
  readonly isJoin: boolean;

  constructor(raw: unknown = {}) {
    super(raw);
    this.sourceId = intValue(pick(raw, 'source_id', 'sourceId', 'SourceID'), -1);
    this.targetId = intValue(pick(raw, 'target_id', 'targetId', 'TargetID'), -1);
    this.label = stringValue(pick(raw, 'label', 'Label'));
    this.isJoin = boolValue(pick(raw, 'is_join', 'isJoin', 'IsJoin'));
  }
}

export class AnalysisGraphState extends RawBackedModel<unknown> {
  readonly states: ARGNode[];
  readonly transitions: ARGTransition[];
  readonly covering: ARGCover[];

  constructor(raw: unknown = {}) {
    super(raw);
    this.states = typedArray(pick(raw, 'states', 'States'), ARGNode);
    this.transitions = typedArray(pick(raw, 'transitions', 'Transitions'), ARGTransition);
    this.covering = typedArray(pick(raw, 'covering', 'Covering'), ARGCover);
  }
}

// FullAnalysisGraphState is a superset of AnalysisGraphState where each node
// carries formula and universe data for CTI inspection.
export class FullAnalysisGraphState extends RawBackedModel<unknown> {
  readonly states: FullARGNode[];
  readonly transitions: ARGTransition[];
  readonly covering: ARGCover[];

  constructor(raw: unknown = {}) {
    super(raw);
    this.states = typedArray(pick(raw, 'states', 'States'), FullARGNode);
    this.transitions = typedArray(pick(raw, 'transitions', 'Transitions'), ARGTransition);
    this.covering = typedArray(pick(raw, 'covering', 'Covering'), ARGCover);
  }
}

export class CyPosition extends RawBackedModel<unknown> {
  readonly x: number;
  readonly y: number;

  constructor(raw: unknown = {}) {
    super(raw);
    this.x = numberValue(pick(raw, 'x', 'X'));
    this.y = numberValue(pick(raw, 'y', 'Y'));
  }
}

export interface CyPositionData {
  x: number;
  y: number;
}

export type GraphPositionMap = Record<string, CyPositionData>;

function positionData(raw: unknown): CyPositionData | null {
  if (!isRecord(raw)) return null;
  const x = pick(raw, 'x', 'X');
  const y = pick(raw, 'y', 'Y');
  if (typeof x !== 'number' || typeof y !== 'number') return null;
  if (!Number.isFinite(x) || !Number.isFinite(y)) return null;
  return { x, y };
}

function graphPositionMap(value: unknown): GraphPositionMap {
  const out: GraphPositionMap = {};
  const obj = rawRecord(value);
  for (const id of Object.keys(obj)) {
    const position = positionData(obj[id]);
    if (position) out[id] = position;
  }
  return out;
}

export class GraphAction {
  readonly label: string;
  readonly action: string;
  readonly id: string;
  readonly name: string;
  readonly args: RawRecord;

  constructor(raw: unknown = {}) {
    const obj = rawRecord(raw);
    this.label = stringValue(pick(raw, 'label', 'Label', 'name', 'Name'));
    this.action = stringValue(pick(raw, 'action', 'Action', 'id', 'ID'));
    this.id = stringValue(pick(raw, 'id', 'ID'), this.action);
    this.name = stringValue(pick(raw, 'name', 'Name'), this.label);
    this.args = cloneRecord(pick(raw, 'args', 'Args'));
  }

  toClientAction(): RawRecord {
    const out: RawRecord = {};
    if (this.label) out.label = this.label;
    if (this.action) out.action = this.action;
    if (this.id && this.id !== this.action) out.id = this.id;
    if (this.name && this.name !== this.label) out.name = this.name;
    if (Object.keys(this.args).length > 0) out.args = cloneJsonish(this.args);
    return out;
  }
}

export class CyElementData {
  readonly id: string;
  readonly obj: string;
  readonly label: string;
  readonly source: string;
  readonly target: string;
  readonly sourceObj: string;
  readonly targetObj: string;
  readonly shortInfo: string;
  readonly longInfo: unknown;
  readonly actions: GraphAction[];
  readonly cluster: string;
  readonly displayLabel: string;
  readonly sort: string;
  readonly width: number | undefined;
  readonly height: number | undefined;
  readonly shape: string;
  readonly borderColor: string;
  readonly lineColor: string;
  readonly isSafe: boolean;
  readonly isMarked: boolean;
  readonly layoutSource: string;
  readonly layoutTarget: string;
  readonly layoutSourceObj: string;
  readonly layoutTargetObj: string;
  readonly layoutReversed: boolean;
  readonly layoutConstraint: boolean | undefined;
  readonly layoutFor: string;

  constructor(raw: unknown = {}) {
    this.id = stringValue(pick(raw, 'id', 'ID'));
    this.obj = stringValue(pick(raw, 'obj', 'Obj'));
    this.label = stringValue(pick(raw, 'label', 'Label'));
    this.source = stringValue(pick(raw, 'source', 'Source'));
    this.target = stringValue(pick(raw, 'target', 'Target'));
    this.sourceObj = stringValue(pick(raw, 'source_obj', 'sourceObj', 'SourceObj'));
    this.targetObj = stringValue(pick(raw, 'target_obj', 'targetObj', 'TargetObj'));
    this.shortInfo = stringValue(pick(raw, 'short_info', 'shortInfo', 'ShortInfo'));
    this.longInfo = cloneJsonish(pick(raw, 'long_info', 'longInfo', 'LongInfo'));
    this.actions = typedArray(pick(raw, 'actions', 'Actions'), GraphAction);
    this.cluster = stringValue(pick(raw, 'cluster', 'Cluster'));
    this.displayLabel = stringValue(pick(raw, 'display_label', 'displayLabel', 'DisplayLabel'));
    this.sort = stringValue(pick(raw, 'sort', 'Sort'));
    this.width = optionalNumber(pick(raw, 'width', 'Width'));
    this.height = optionalNumber(pick(raw, 'height', 'Height'));
    this.shape = stringValue(pick(raw, 'shape', 'Shape'));
    this.borderColor = stringValue(pick(raw, 'border_color', 'borderColor', 'BorderColor'));
    this.lineColor = stringValue(pick(raw, 'line_color', 'lineColor', 'LineColor', 'color', 'Color'));
    this.isSafe = boolValue(pick(raw, 'is_safe', 'isSafe', 'IsSafe'));
    this.isMarked = boolValue(pick(raw, 'is_marked', 'isMarked', 'IsMarked'));
    this.layoutSource = stringValue(pick(raw, 'layout_source', 'layoutSource', 'LayoutSource'));
    this.layoutTarget = stringValue(pick(raw, 'layout_target', 'layoutTarget', 'LayoutTarget'));
    this.layoutSourceObj = stringValue(pick(raw, 'layout_source_obj', 'layoutSourceObj', 'LayoutSourceObj'));
    this.layoutTargetObj = stringValue(pick(raw, 'layout_target_obj', 'layoutTargetObj', 'LayoutTargetObj'));
    this.layoutReversed = boolValue(pick(raw, 'layout_reversed', 'layoutReversed', 'LayoutReversed'));
    this.layoutConstraint = optionalBoolean(pick(raw, 'layout_constraint', 'layoutConstraint', 'LayoutConstraint'));
    this.layoutFor = stringValue(pick(raw, 'layout_for', 'layoutFor', 'LayoutFor'));
  }

  toCytoscapeData(): RawRecord {
    const out: RawRecord = {};
    if (this.id) out.id = this.id;
    if (this.obj) out.obj = this.obj;
    if (this.label) out.label = this.label;
    if (this.source) out.source = this.source;
    if (this.target) out.target = this.target;
    if (this.sourceObj) out.source_obj = this.sourceObj;
    if (this.targetObj) out.target_obj = this.targetObj;
    if (this.shortInfo) out.short_info = this.shortInfo;
    if (this.longInfo !== undefined) out.long_info = cloneJsonish(this.longInfo);
    if (this.actions.length > 0) out.actions = this.actions.map((action) => action.toClientAction());
    if (this.cluster) out.cluster = this.cluster;
    if (this.displayLabel) out.display_label = this.displayLabel;
    if (this.sort) out.sort = this.sort;
    if (this.width !== undefined) out.width = this.width;
    if (this.height !== undefined) out.height = this.height;
    if (this.shape) out.shape = this.shape;
    if (this.borderColor) out.border_color = this.borderColor;
    if (this.lineColor) out.line_color = this.lineColor;
    if (this.isSafe) out.is_safe = true;
    if (this.isMarked) out.is_marked = true;
    if (this.layoutSource) out.layout_source = this.layoutSource;
    if (this.layoutTarget) out.layout_target = this.layoutTarget;
    if (this.layoutSourceObj) out.layout_source_obj = this.layoutSourceObj;
    if (this.layoutTargetObj) out.layout_target_obj = this.layoutTargetObj;
    if (this.layoutReversed) out.layout_reversed = true;
    if (this.layoutConstraint !== undefined) out.layout_constraint = this.layoutConstraint;
    if (this.layoutFor) out.layout_for = this.layoutFor;
    return out;
  }
}

export type GraphSelectionKind = 'node' | 'edge';

export interface GraphSelection {
  kind: GraphSelectionKind;
  id: string;
  obj: string;
  label: string;
  sourceObj: string;
  targetObj: string;
}

function normalizeGraphSelection(raw: unknown, fallbackKind: GraphSelectionKind): GraphSelection | null {
  if (typeof raw === 'string') {
    const value = raw.trim();
    if (!value) return null;
    return { kind: fallbackKind, id: value, obj: value, label: value, sourceObj: '', targetObj: '' };
  }
  const obj = rawRecord(raw);
  const kind = stringValue(pick(raw, 'kind'), fallbackKind) as GraphSelectionKind;
  const data = rawRecord(pick(raw, 'data'));
  const source = Object.keys(data).length > 0 ? data : obj;
  const id = stringValue(pick(source, 'id', 'ID'));
  const selectedObj = stringValue(pick(source, 'obj', 'Obj'), id);
  const label = stringValue(pick(source, 'label', 'Label'), selectedObj || id);
  const sourceObj = stringValue(pick(source, 'source_obj', 'sourceObj', 'SourceObj'));
  const targetObj = stringValue(pick(source, 'target_obj', 'targetObj', 'TargetObj'));
  if (!id && !selectedObj && !label) return null;
  return { kind, id, obj: selectedObj, label, sourceObj, targetObj };
}

export function conceptNodeSelection(raw: unknown): GraphSelection | null {
  return normalizeGraphSelection(raw, 'node');
}

export function conceptEdgeSelection(raw: unknown): GraphSelection | null {
  return normalizeGraphSelection(raw, 'edge');
}

export function graphSelectionKey(selection: GraphSelection | null | undefined): string {
  if (!selection) return '';
  if (selection.kind === 'edge') {
    const tuple = [selection.obj, selection.sourceObj, selection.targetObj].filter(Boolean).join('|');
    return `edge:${selection.id || tuple || selection.label}`;
  }
  return `node:${selection.obj || selection.id || selection.label}`;
}

export function graphSelectionAliases(selection: GraphSelection | null | undefined): string[] {
  if (!selection) return [];
  const aliases = new Set<string>();
  const primary = graphSelectionKey(selection);
  if (primary) aliases.add(primary);
  if (selection.id) aliases.add(`${selection.kind}:${selection.id}`);
  if (selection.obj) aliases.add(`${selection.kind}:${selection.obj}`);
  if (selection.kind === 'edge' && selection.obj && selection.sourceObj && selection.targetObj) {
    aliases.add(`edge:${selection.obj}|${selection.sourceObj}|${selection.targetObj}`);
  }
  return Array.from(aliases);
}

export class CyElement {
  readonly group: string;
  readonly data: CyElementData;
  readonly classes: string;
  readonly locked: boolean;
  readonly position: CyPosition | null;

  constructor(raw: unknown = {}) {
    this.group = stringValue(pick(raw, 'group', 'Group'));
    this.data = new CyElementData(pick(raw, 'data', 'Data'));
    this.classes = stringValue(pick(raw, 'classes', 'Classes'));
    this.locked = boolValue(pick(raw, 'locked', 'Locked'));
    const position = pick(raw, 'position', 'Position');
    this.position = position ? new CyPosition(position) : null;
  }
}

export class CyElements {
  readonly elements: CyElement[];
  readonly positions: GraphPositionMap;

  constructor(raw: unknown = {}) {
    this.elements = typedArray(pick(raw, 'elements', 'Elements'), CyElement);
    this.positions = graphPositionMap(pick(raw, 'positions', 'Positions'));
    for (const element of this.elements) {
      if (element.group !== 'nodes' || !element.position) continue;
      const id = element.data.id || element.data.obj || element.data.label;
      if (id && !this.positions[id]) {
        this.positions[id] = { x: element.position.x, y: element.position.y };
      }
    }
    // NodeID and EdgeID are tagged json:"-" in Go and never appear on the wire.
    // Use typed CyElementData id/obj/tuple fields for lookups instead.
  }
}

function nodeIds(elements: CyElement[]): string[] {
  return elements
    .filter((element) => element.group === 'nodes')
    .map((element) => element.data.id || element.data.obj || element.data.label)
    .filter((id): id is string => !!id);
}

function clonePositionMap(positions: GraphPositionMap | null | undefined): GraphPositionMap {
  const out: GraphPositionMap = {};
  for (const [id, position] of Object.entries(positions || {})) {
    if (!position) continue;
    out[id] = { x: position.x, y: position.y };
  }
  return out;
}

export function positionsForElements(elements: CyElement[], incoming: GraphPositionMap | null | undefined, previous: GraphPositionMap | null | undefined): GraphPositionMap {
  const out: GraphPositionMap = {};
  const incomingMap = incoming || {};
  const previousMap = previous || {};
  for (const id of nodeIds(elements)) {
    const position = incomingMap[id] || previousMap[id];
    if (position) out[id] = { x: position.x, y: position.y };
  }
  return out;
}

export class FactSelection extends RawBackedModel<unknown> {
  readonly index: number;
  readonly text: string;
  readonly selected: boolean;

  constructor(raw: unknown = {}) {
    super(raw);
    this.index = intValue(pick(raw, 'index', 'Index'));
    this.text = stringValue(pick(raw, 'text', 'Text'));
    this.selected = boolValue(pick(raw, 'selected', 'Selected'));
  }
}

export class Option extends RawBackedModel<unknown> {
  readonly val: boolean;

  constructor(raw: unknown = false) {
    super(raw);
    this.val = isRecord(raw) ? boolValue(pick(raw, 'val', 'Val')) : boolValue(raw);
  }
}

export type CheckboxOptionMap = Record<string, Record<string, Option>>;

export class DisplayCheckboxes extends RawBackedModel<unknown> {
  readonly edgeDisplayCheckboxes: CheckboxOptionMap;
  readonly nodeLabelDisplayCheckboxes: CheckboxOptionMap;

  constructor(raw: unknown = {}) {
    super(raw);
    const edgePayload = pick(raw, 'edge_display_checkboxes', 'edges', 'EdgeDisplayCheckboxes') ?? {};
    const labelPayload = pick(raw, 'node_label_display_checkboxes', 'labels', 'NodeLabelDisplayCheckboxes') ?? {};
    this.edgeDisplayCheckboxes = this.optionMap(edgePayload);
    this.nodeLabelDisplayCheckboxes = this.optionMap(labelPayload);
  }

  edgeVisible(name: string, displayClass: string): boolean {
    return !!this.edgeDisplayCheckboxes[name]?.[displayClass]?.val;
  }

  nodeLabelVisible(name: string, displayClass: string): boolean {
    return !!this.nodeLabelDisplayCheckboxes[name]?.[displayClass]?.val;
  }

  private optionMap(rawMap: unknown): CheckboxOptionMap {
    const out: CheckboxOptionMap = {};
    const obj = rawRecord(rawMap);
    for (const name of Object.keys(obj)) {
      out[name] = {};
      const boxes = rawRecord(obj[name]);
      for (const key of Object.keys(boxes)) {
        out[name][key] = new Option(boxes[key]);
      }
    }
    return out;
  }
}

export class Concept extends RawBackedModel<unknown> {
  readonly name: string;
  readonly variables: string[];
  readonly formula: string;
  readonly sorts: string[];
  readonly arity: number;

  constructor(raw: unknown = {}) {
    super(raw);
    this.name = stringValue(pick(raw, 'name', 'Name'));
    this.variables = stringArray(pick(raw, 'variables', 'Variables'));
    this.formula = stringValue(pick(raw, 'formula', 'Formula'));
    this.sorts = stringArray(pick(raw, 'sorts', 'Sorts'));
    this.arity = intValue(pick(raw, 'arity', 'Arity'), this.variables.length);
  }
}

export class ConceptCombiner extends RawBackedModel<unknown> {
  readonly name: string;
  readonly label: string;
  readonly source: string;
  readonly target: string;
  readonly formula: string;

  constructor(raw: unknown = {}) {
    super(raw);
    this.name = stringValue(pick(raw, 'name', 'Name'));
    this.label = stringValue(pick(raw, 'label', 'Label'));
    this.source = stringValue(pick(raw, 'source', 'Source'));
    this.target = stringValue(pick(raw, 'target', 'Target'));
    this.formula = stringValue(pick(raw, 'formula', 'Formula'));
  }
}

export class ConceptDomain extends RawBackedModel<unknown> {
  readonly concepts: Record<string, Concept>;
  readonly combiners: ConceptCombiner[];
  readonly nodes: string[];
  readonly edges: string[];
  readonly nodeLabels: string[];

  constructor(raw: unknown = {}) {
    super(raw);
    const concepts = rawRecord(pick(raw, 'concepts', 'Concepts'));
    this.concepts = {};
    for (const name of Object.keys(concepts)) {
      this.concepts[name] = new Concept(concepts[name]);
    }
    this.combiners = typedArray(pick(raw, 'combiners', 'Combiners'), ConceptCombiner);
    this.nodes = stringArray(pick(raw, 'nodes', 'Nodes'));
    this.edges = stringArray(pick(raw, 'edges', 'Edges'));
    this.nodeLabels = stringArray(pick(raw, 'node_labels', 'nodeLabels', 'NodeLabels'));
  }
}

export class ConceptSession extends RawBackedModel<unknown> {
  readonly domain: ConceptDomain;
  readonly abstractValue: Record<string, boolean>;

  constructor(raw: unknown = {}) {
    super(raw);
    this.domain = new ConceptDomain(pick(raw, 'domain', 'Domain') ?? {});
    this.abstractValue = boolMap(pick(raw, 'abstract_value', 'abstractValue', 'AbstractValue'));
    // undoDepth/redoDepth are only present in ConceptInteractiveSession wire format.
  }
}

function tagValueArrayToMap(arr: unknown[]): Record<string, boolean> {
  const out: Record<string, boolean> = {};
  for (const item of arr) {
    if (!isRecord(item)) continue;
    const tag = item['Tag'] ?? item['tag'];
    const val = item['Value'] ?? item['value'];
    if (Array.isArray(tag) && typeof val === 'boolean')
      out[(tag as string[]).join('|')] = val;
  }
  return out;
}

export class ConceptInteractiveSession extends ConceptSession {
  readonly undoDepth: number;
  readonly redoDepth: number;
  readonly state: string;
  readonly axioms: string;
  readonly goalConstraints: string[];
  readonly supposeConstraints: string[];
  readonly cache: Record<string, boolean>;
  readonly info: string;

  constructor(raw: unknown = {}) {
    super(raw);
    // Go sends abstract_value as []TagValue (array of {Tag: string[], Value: bool}),
    // not as map[string]bool. Convert to Record<string,bool> keyed by "|"-joined tag.
    // The domain field inherited from ConceptSession is always empty: Go no longer sends
    // the "domain" key in the CIS payload (CDConceptDomain had unexported fields and
    // serialized to {}). Use snapshot.domain (from concept_domain key) instead.
    const avRaw = pick(raw, 'abstract_value', 'abstractValue', 'AbstractValue');
    if (Array.isArray(avRaw)) {
      (this as any).abstractValue = tagValueArrayToMap(avRaw);
    }
    this.undoDepth = intValue(pick(raw, 'undo_depth', 'undoDepth'));
    this.redoDepth = intValue(pick(raw, 'redo_depth', 'redoDepth'));
    this.state = stringValue(pick(raw, 'state', 'State'));
    this.axioms = stringValue(pick(raw, 'axioms', 'Axioms'));
    this.goalConstraints = stringArray(pick(raw, 'goal_constraints', 'goalConstraints', 'GoalConstraints'));
    this.supposeConstraints = stringArray(pick(raw, 'suppose_constraints', 'supposeConstraints', 'SupposeConstraints'));
    this.cache = boolMap(pick(raw, 'cache', 'Cache'));
    this.info = stringValue(pick(raw, 'info', 'Info'));
  }
}

export class GraphStack extends RawBackedModel<unknown> {
  readonly canUndo: boolean;
  readonly canRedo: boolean;
  readonly undoDepth: number;
  readonly redoDepth: number;

  constructor(raw: unknown = {}) {
    super(raw);
    this.canUndo = boolValue(pick(raw, 'can_undo', 'canUndo'));
    this.canRedo = boolValue(pick(raw, 'can_redo', 'canRedo'));
    this.undoDepth = intValue(pick(raw, 'undo_depth', 'undoDepth'));
    this.redoDepth = intValue(pick(raw, 'redo_depth', 'redoDepth'));
    // current/undoStack/redoStack are not emitted by conceptGraphStackPayload().
  }
}

export class ConceptGraphModel extends RawBackedModel<unknown> {
  readonly sorts: string[];
  readonly conceptSession: ConceptSession;
  readonly interactiveSession: ConceptInteractiveSession;
  readonly displayCheckboxes: DisplayCheckboxes;
  readonly graphStack: GraphStack;
  readonly newRelations: string[];
  readonly state: string;
  readonly concrete: string;
  readonly attributes: string[];
  readonly reverseResult: string[];

  constructor(raw: unknown = {}) {
    super(raw);
    this.sorts = stringArray(pick(raw, 'sorts', 'Sorts'));
    this.conceptSession = new ConceptSession(pick(raw, 'concept_session', 'conceptSession', 'ConceptSess') ?? {});
    this.interactiveSession = new ConceptInteractiveSession(pick(raw, 'concept_interactive_session', 'interactive_session', 'InteractiveSess') ?? {});
    this.displayCheckboxes = new DisplayCheckboxes(pick(raw, 'display_checkboxes', 'checks', 'Checks', 'toggles') ?? {});
    this.graphStack = new GraphStack(pick(raw, 'graph_stack', 'graphStack', 'GraphStack') ?? {});
    this.newRelations = stringArray(pick(raw, 'new_relations', 'newRelations', 'NewRelations'));
    this.state = stringValue(pick(raw, 'state', 'State'));
    this.concrete = stringValue(pick(raw, 'concrete', 'Concrete'));
    this.attributes = stringArray(pick(raw, 'attributes', 'Attributes'));
    this.reverseResult = stringArray(pick(raw, 'reverse_result', 'reverseResult', 'ReverseResult'));
  }
}

export class ARGSnapshot extends RawBackedModel<unknown> {
  readonly render: CyElements;
  readonly analysisGraphState: AnalysisGraphState;
  readonly fullAnalysisGraphState: FullAnalysisGraphState;

  constructor(raw: unknown = {}) {
    super(raw);
    // ARG payload is flat: {elements:[…], analysis_graph_state:{…}}.
    // CyElements reads 'elements' from the raw root; graph state from sub-key.
    this.render = new CyElements(raw);
    const ags = pick(raw, 'analysis_graph_state', 'analysisGraphState') ?? {};
    this.analysisGraphState = new AnalysisGraphState(ags);
    this.fullAnalysisGraphState = new FullAnalysisGraphState(ags);
    // Note: no 'analysis_graph' key exists in the Go wire format.
  }
}

// CTISnapshot is returned by GET /api/session/{id}/arg/cti.
// It extends ARGSnapshot with CTI-specific fields.
export class CTISnapshot extends ARGSnapshot {
  readonly haveCti: boolean;
  readonly currentConjecture: string;

  constructor(raw: unknown = {}) {
    super(raw);
    this.haveCti = boolValue(pick(raw, 'have_cti', 'haveCti'));
    this.currentConjecture = stringValue(pick(raw, 'current_conjecture', 'currentConjecture'));
  }
}

export class ConceptSnapshot extends RawBackedModel<unknown> {
  readonly render: CyElements;
  readonly domain: ConceptDomain;
  readonly session: ConceptSession;
  readonly interactiveSession: ConceptInteractiveSession;
  readonly displayCheckboxes: DisplayCheckboxes;
  readonly graphStack: GraphStack;
  readonly graph: ConceptGraphModel;
  readonly facts: FactSelection[];
  readonly nodes: string[];
  readonly edges: string[];
  readonly nodeLabels: string[];
  readonly relations: string[];
  readonly relationColors: Record<string, string>;
  readonly edgeSorts: Record<string, string[]>;
  readonly labelSorts: Record<string, string>;
  readonly abstractValue: Record<string, boolean>;
  readonly sheetId: string;
  readonly selectedNode: string | null;
  readonly stateLabel: string;

  constructor(raw: unknown = {}) {
    super(raw);
    this.render = new CyElements(raw);
    this.domain = new ConceptDomain(pick(raw, 'concept_domain', 'conceptDomain') ?? {});
    this.session = new ConceptSession(pick(raw, 'concept_session', 'conceptSession') ?? {});
    this.interactiveSession = new ConceptInteractiveSession(pick(raw, 'concept_interactive_session', 'interactive_session') ?? {});
    this.displayCheckboxes = new DisplayCheckboxes(pick(raw, 'display_checkboxes', 'toggles') ?? {});
    this.graphStack = new GraphStack(pick(raw, 'graph_stack', 'graphStack') ?? {});
    this.graph = new ConceptGraphModel(pick(raw, 'graph', 'Graph') ?? {});
    this.facts = typedArray(pick(raw, 'facts'), FactSelection);
    this.nodes = stringArray(pick(raw, 'nodes', 'Nodes'));
    this.edges = stringArray(pick(raw, 'edges', 'Edges'));
    this.nodeLabels = stringArray(pick(raw, 'node_labels', 'nodeLabels', 'NodeLabels'));
    this.relations = stringArray(pick(raw, 'relations', 'Relations'));
    this.relationColors = stringMap(pick(raw, 'relation_colors', 'relationColors', 'RelationColors'));
    this.edgeSorts = stringArrayMap(pick(raw, 'edge_sorts', 'edgeSorts', 'EdgeSorts'));
    const labelSortsRaw = rawRecord(pick(raw, 'label_sorts', 'labelSorts', 'LabelSorts'));
    this.labelSorts = {};
    for (const key of Object.keys(labelSortsRaw)) {
      this.labelSorts[key] = stringValue(labelSortsRaw[key]);
    }
    this.abstractValue = boolMap(pick(raw, 'abstract_value', 'abstractValue', 'AbstractValue'));
    this.sheetId = stringValue(pick(raw, 'sheet_id', 'sheetId'));
    const selected = pick(raw, 'selected_node', 'selectedNode');
    this.selectedNode = typeof selected === 'string' ? selected : null;
    this.stateLabel = stringValue(pick(raw, 'state_label', 'stateLabel'));
  }
}

export type SheetKind = 'analysis' | 'events' | string;

export class SheetModel extends RawBackedModel<unknown> {
  readonly id: string;
  readonly type: SheetKind;
  arg: ARGSnapshot | null;
  concept: ConceptSnapshot | null;
  cti: CTISnapshot | null;
  argPositions: GraphPositionMap;
  conceptPositions: GraphPositionMap;
  selectedArgNode: string | null;
  conceptSelections: GraphSelection[];
  visualOnly: boolean;
  reachabilityOnly: boolean;

  constructor(raw: unknown = {}) {
    super(raw);
    this.id = stringValue(pick(raw, 'id'));
    this.type = stringValue(pick(raw, 'type'), 'analysis');
    this.arg = null;
    this.concept = null;
    this.cti = null;
    this.argPositions = graphPositionMap(pick(raw, 'argPositions', 'arg_positions'));
    this.conceptPositions = graphPositionMap(pick(raw, 'conceptPositions', 'concept_positions'));
    const selected = pick(raw, 'selectedArgNode', 'selected_arg_node');
    this.selectedArgNode = typeof selected === 'string' ? selected : null;
    this.conceptSelections = rawArray(pick(raw, 'conceptSelections', 'concept_selections'))
      .map((item) => normalizeGraphSelection(item, 'node'))
      .filter((item): item is GraphSelection => !!item);
    this.visualOnly = boolValue(pick(raw, 'visualOnly', 'visual_only'));
    this.reachabilityOnly = boolValue(pick(raw, 'reachabilityOnly', 'reachability_only'));
  }
}

export interface SessionMetadata {
  id: string;
  mode: string;
  status: string;
  fileName: string;
  filePath: string;
}

export class UIDataModel extends RawBackedModel<unknown> {
  session: SessionMetadata;
  sheets: Record<string, SheetModel>;
  activeSheetId: string;

  constructor(raw: unknown = {}) {
    super(raw);
    this.session = {
      id: '',
      mode: '',
      status: '',
      fileName: '',
      filePath: '',
    };
    this.sheets = {};
    this.activeSheetId = 'sheet-1';
  }

  setSessionMetadata(metadata: Partial<SessionMetadata> = {}): SessionMetadata {
    this.session = { ...this.session, ...metadata };
    return this.session;
  }

  registerSheet(sheetId: string, raw: unknown = {}): SheetModel {
    if (!this.sheets[sheetId]) {
      this.sheets[sheetId] = new SheetModel({ ...rawRecord(raw), id: sheetId });
    }
    return this.sheets[sheetId];
  }

  removeSheet(sheetId: string): void {
    delete this.sheets[sheetId];
    if (this.activeSheetId === sheetId) this.activeSheetId = 'sheet-1';
  }

  setActiveSheet(sheetId: string): string {
    if (sheetId) this.activeSheetId = sheetId;
    return this.activeSheetId;
  }

  setSelectedArgNode(sheetId: string, nodeId: string | null | undefined): string | null {
    const sheet = this.registerSheet(sheetId || this.activeSheetId);
    sheet.selectedArgNode = nodeId || null;
    return sheet.selectedArgNode;
  }

  setGraphPositions(sheetId: string, graphKind: 'arg' | 'concept', positions: GraphPositionMap | null | undefined): GraphPositionMap {
    const sheet = this.registerSheet(sheetId || this.activeSheetId);
    const next = clonePositionMap(positions);
    if (graphKind === 'arg') {
      sheet.argPositions = next;
    } else {
      sheet.conceptPositions = next;
    }
    return next;
  }

  toggleConceptNodeSelection(sheetId: string, selectionInput: unknown): boolean {
    const sheet = this.registerSheet(sheetId || this.activeSheetId);
    const selection = conceptNodeSelection(selectionInput);
    if (!selection) return false;
    const aliases = new Set(graphSelectionAliases(selection));
    const existing = sheet.conceptSelections.findIndex((candidate) =>
      graphSelectionAliases(candidate).some((alias) => aliases.has(alias)));
    if (existing >= 0) {
      sheet.conceptSelections.splice(existing, 1);
      return false;
    }
    sheet.conceptSelections.push(selection);
    return true;
  }

  toggleConceptEdgeSelection(sheetId: string, selectionInput: unknown): boolean {
    const sheet = this.registerSheet(sheetId || this.activeSheetId);
    const selection = conceptEdgeSelection(selectionInput);
    if (!selection) return false;
    const aliases = new Set(graphSelectionAliases(selection));
    const existing = sheet.conceptSelections.findIndex((candidate) =>
      graphSelectionAliases(candidate).some((alias) => aliases.has(alias)));
    if (existing >= 0) {
      sheet.conceptSelections.splice(existing, 1);
      return false;
    }
    sheet.conceptSelections.push(selection);
    return true;
  }

  clearConceptSelection(sheetId: string): void {
    const sheet = this.registerSheet(sheetId || this.activeSheetId);
    sheet.conceptSelections = [];
  }

  setConceptSelections(sheetId: string, selections: unknown[] = []): GraphSelection[] {
    const sheet = this.registerSheet(sheetId || this.activeSheetId);
    sheet.conceptSelections = selections
      .map((item) => normalizeGraphSelection(item, 'node'))
      .filter((item): item is GraphSelection => !!item);
    return sheet.conceptSelections;
  }

  setFile(fileName: string, filePath?: string): void {
    this.session.fileName = fileName || '';
    this.session.filePath = filePath || fileName || '';
  }
}
