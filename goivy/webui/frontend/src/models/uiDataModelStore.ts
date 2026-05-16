import {
  ARGSnapshot,
  CTISnapshot,
  ConceptSnapshot,
  GraphPositionMap,
  UIDataModel,
  positionsForElements,
} from './uiDataModel.ts';

export type UIDataChangeKind =
  | 'arg'
  | 'concept'
  | 'cti'
  | 'sheet'
  | 'argSelection'
  | 'conceptSelection'
  | 'argLayout'
  | 'conceptLayout'
  | 'visualOnly'
  | 'interaction';

export interface UIDataChange {
  sheetId: string;
  changed: UIDataChangeKind[];
}

type Listener = (change: UIDataChange) => void;
type RawRecord = Record<string, any>;

export interface ToggleUpdate {
  edge?: string;
  label?: string;
  display_class?: string;
  displayClass?: string;
  value?: boolean;
}

function isRecord(value: unknown): value is RawRecord {
  return !!value && typeof value === 'object' && !Array.isArray(value);
}

function cloneRaw(value: unknown): RawRecord {
  return isRecord(value) ? { ...value } : {};
}

function checkboxMapToBooleans(map: Record<string, Record<string, { val: boolean }>> = {}) {
  const out: Record<string, Record<string, boolean>> = {};
  for (const [name, checks] of Object.entries(map || {})) {
    out[name] = {};
    for (const [displayClass, option] of Object.entries(checks || {})) {
      out[name][displayClass] = !!(option && option.val);
    }
  }
  return out;
}

function conceptSnapshotWithToggles(snapshot: ConceptSnapshot, updates: ToggleUpdate[]): ConceptSnapshot {
  const edges = checkboxMapToBooleans(snapshot.displayCheckboxes.edgeDisplayCheckboxes);
  const labels = checkboxMapToBooleans(snapshot.displayCheckboxes.nodeLabelDisplayCheckboxes);
  for (const update of updates || []) {
    if (!update) continue;
    const displayClass = update.display_class || update.displayClass;
    if (!displayClass) continue;
    if (update.edge) {
      edges[update.edge] = edges[update.edge] || {};
      edges[update.edge][displayClass] = !!update.value;
    }
    if (update.label) {
      labels[update.label] = labels[update.label] || {};
      labels[update.label][displayClass] = !!update.value;
    }
  }
  const toggles = { edges, labels };
  return new ConceptSnapshot({
    ...cloneRaw(snapshot.raw),
    display_checkboxes: toggles,
    toggles,
  });
}

export class UIDataModelStore {
  readonly model: UIDataModel;
  private listeners: Listener[];

  constructor(model: UIDataModel) {
    this.model = model;
    this.listeners = [];
  }

  subscribe(listener: Listener): () => void {
    this.listeners.push(listener);
    return () => {
      this.listeners = this.listeners.filter((candidate) => candidate !== listener);
    };
  }

  emit(change: UIDataChange): void {
    for (const listener of this.listeners.slice()) {
      listener(change);
    }
  }

  registerSheet(sheetId: string, raw: unknown = {}) {
    const sheet = this.model.registerSheet(sheetId || this.model.activeSheetId, raw);
    this.emit({ sheetId: sheet.id, changed: ['sheet'] });
    return sheet;
  }

  removeSheet(sheetId: string): void {
    this.model.removeSheet(sheetId);
    this.emit({ sheetId, changed: ['sheet'] });
  }

  setActiveSheet(sheetId: string): string {
    const active = this.model.setActiveSheet(sheetId);
    this.emit({ sheetId: active, changed: ['sheet'] });
    return active;
  }

  applyArgSnapshot(sheetId: string, payload: unknown = {}) {
    const id = sheetId || this.model.activeSheetId;
    const sheet = this.model.registerSheet(id);
    const snapshot = new ARGSnapshot(payload);
    sheet.arg = snapshot;
    sheet.argPositions = positionsForElements(snapshot.render.elements, snapshot.render.positions, sheet.argPositions);
    sheet.selectedArgNode = null;
    this.emit({ sheetId: id, changed: ['arg', 'argSelection'] });
    return snapshot;
  }

  applyConceptSnapshot(sheetId: string, payload: unknown = {}) {
    const id = sheetId || this.model.activeSheetId;
    const sheet = this.model.registerSheet(id);
    const snapshot = new ConceptSnapshot(payload);
    sheet.concept = snapshot;
    sheet.conceptPositions = positionsForElements(snapshot.render.elements, snapshot.render.positions, sheet.conceptPositions);
    sheet.conceptSelections = [];
    this.emit({ sheetId: id, changed: ['concept', 'conceptSelection'] });
    return snapshot;
  }

  applyCtiSnapshot(sheetId: string, payload: unknown = {}) {
    const id = sheetId || this.model.activeSheetId;
    const sheet = this.model.registerSheet(id);
    const snapshot = new CTISnapshot(payload);
    sheet.cti = snapshot;
    this.emit({ sheetId: id, changed: ['cti'] });
    return snapshot;
  }

  setSelectedArgNode(sheetId: string, nodeId: string | null | undefined): string | null {
    const id = sheetId || this.model.activeSheetId;
    const selected = this.model.setSelectedArgNode(id, nodeId);
    this.emit({ sheetId: id, changed: ['argSelection'] });
    return selected;
  }

  setGraphPositions(sheetId: string, graphKind: 'arg' | 'concept', positions: GraphPositionMap | null | undefined, {
    emit = true,
  } = {}): GraphPositionMap {
    const id = sheetId || this.model.activeSheetId;
    const next = this.model.setGraphPositions(id, graphKind, positions);
    if (emit) {
      this.emit({ sheetId: id, changed: [graphKind === 'arg' ? 'argLayout' : 'conceptLayout'] });
    }
    return next;
  }

  toggleConceptNodeSelection(sheetId: string, node: unknown): boolean {
    const id = sheetId || this.model.activeSheetId;
    const selected = this.model.toggleConceptNodeSelection(id, node);
    this.emit({ sheetId: id, changed: ['conceptSelection'] });
    return selected;
  }

  toggleConceptEdgeSelection(sheetId: string, edge: unknown): boolean {
    const id = sheetId || this.model.activeSheetId;
    const selected = this.model.toggleConceptEdgeSelection(id, edge);
    this.emit({ sheetId: id, changed: ['conceptSelection'] });
    return selected;
  }

  clearConceptSelection(sheetId: string): void {
    const id = sheetId || this.model.activeSheetId;
    this.model.clearConceptSelection(id);
    this.emit({ sheetId: id, changed: ['conceptSelection'] });
  }

  setConceptSelections(sheetId: string, selections: unknown[] = []) {
    const id = sheetId || this.model.activeSheetId;
    const selected = this.model.setConceptSelections(id, selections);
    this.emit({ sheetId: id, changed: ['conceptSelection'] });
    return selected;
  }

  setConceptToggles(sheetId: string, updates: ToggleUpdate[] = []) {
    const id = sheetId || this.model.activeSheetId;
    const sheet = this.model.registerSheet(id);
    if (!sheet.concept) return null;
    sheet.concept = conceptSnapshotWithToggles(sheet.concept, updates);
    this.emit({ sheetId: id, changed: ['concept'] });
    return sheet.concept;
  }

  setVisualOnly(sheetId: string, visualOnly: boolean, raw: unknown = {}): boolean {
    const id = sheetId || this.model.activeSheetId;
    const sheet = this.model.registerSheet(id, raw);
    sheet.visualOnly = !!visualOnly;
    this.emit({ sheetId: id, changed: ['visualOnly'] });
    return sheet.visualOnly;
  }
}

export function createUIDataModelStore(model: UIDataModel): UIDataModelStore {
  return new UIDataModelStore(model);
}
