import {
  ARGSnapshot,
  CTISnapshot,
  ConceptSnapshot,
  UIDataModel,
} from './uiDataModel.ts';

export type UIDataChangeKind =
  | 'arg'
  | 'concept'
  | 'cti'
  | 'sheet'
  | 'argSelection'
  | 'conceptSelection'
  | 'visualOnly'
  | 'interaction';

export interface UIDataChange {
  sheetId: string;
  changed: UIDataChangeKind[];
}

type Listener = (change: UIDataChange) => void;

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
    sheet.selectedArgNode = null;
    this.emit({ sheetId: id, changed: ['arg', 'argSelection'] });
    return snapshot;
  }

  applyConceptSnapshot(sheetId: string, payload: unknown = {}) {
    const id = sheetId || this.model.activeSheetId;
    const sheet = this.model.registerSheet(id);
    const snapshot = new ConceptSnapshot(payload);
    sheet.concept = snapshot;
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
