import { UIDataModel } from './uiDataModel.ts';

export type UIDataChangeKind =
  | 'arg'
  | 'concept'
  | 'cti'
  | 'sheet'
  | 'selection'
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
    this.emit({ sheetId: active, changed: ['sheet', 'selection'] });
    return active;
  }

  applyArgSnapshot(sheetId: string, payload: unknown = {}) {
    const id = sheetId || this.model.activeSheetId;
    const snapshot = this.model.acceptArgSnapshot(id, payload);
    this.emit({ sheetId: id, changed: ['arg', 'selection'] });
    return snapshot;
  }

  applyConceptSnapshot(sheetId: string, payload: unknown = {}) {
    const id = sheetId || this.model.activeSheetId;
    const snapshot = this.model.acceptConceptSnapshot(id, payload);
    this.emit({ sheetId: id, changed: ['concept', 'selection'] });
    return snapshot;
  }

  applyCtiSnapshot(sheetId: string, payload: unknown = {}) {
    const id = sheetId || this.model.activeSheetId;
    const snapshot = this.model.acceptCtiSnapshot(id, payload);
    this.emit({ sheetId: id, changed: ['cti'] });
    return snapshot;
  }

  setSelectedArgNode(sheetId: string, nodeId: string | null | undefined): string | null {
    const id = sheetId || this.model.activeSheetId;
    const selected = this.model.setSelectedArgNode(id, nodeId);
    this.emit({ sheetId: id, changed: ['selection'] });
    return selected;
  }

  toggleConceptNodeSelection(sheetId: string, nodeId: string | null | undefined): boolean {
    const id = sheetId || this.model.activeSheetId;
    const selected = this.model.toggleConceptNodeSelection(id, nodeId);
    this.emit({ sheetId: id, changed: ['selection'] });
    return selected;
  }

  toggleConceptEdgeSelection(sheetId: string, edgeId: string | null | undefined): boolean {
    const id = sheetId || this.model.activeSheetId;
    const selected = this.model.toggleConceptEdgeSelection(id, edgeId);
    this.emit({ sheetId: id, changed: ['selection'] });
    return selected;
  }

  clearConceptSelection(sheetId: string): void {
    const id = sheetId || this.model.activeSheetId;
    this.model.clearConceptSelection(id);
    this.emit({ sheetId: id, changed: ['selection'] });
  }

  setVisualOnly(sheetId: string, visualOnly: boolean): boolean {
    const id = sheetId || this.model.activeSheetId;
    const sheet = this.model.registerSheet(id);
    sheet.visualOnly = !!visualOnly;
    this.emit({ sheetId: id, changed: ['visualOnly'] });
    return sheet.visualOnly;
  }
}

export function createUIDataModelStore(model: UIDataModel): UIDataModelStore {
  return new UIDataModelStore(model);
}
