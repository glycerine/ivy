import { createUIDataModelStore } from '../models/uiDataModelStore.ts';
import {
  selectArgGraphView,
  selectConceptGraphView,
  selectConstraintFacts,
  selectSheet,
  selectStateCheckboxRows,
  selectStateLabel,
} from '../models/uiDataSelectors.ts';
import { renderStateCheckboxes, updateStateLabel } from './conceptVisibilityService.ts';
import { populateConstraintFacts } from './detailsService.ts';

function sheetGraphs(app, sheetId) {
  const runtimeSheet = app && app.sheets && app.sheets[sheetId];
  return {
    runtimeSheet,
    argGraph: (runtimeSheet && runtimeSheet.argGraph) || (app && app.argGraph),
    conceptGraph: (runtimeSheet && runtimeSheet.conceptGraph) || (app && app.conceptGraph),
  };
}

function syncRuntimeSheetMirrors(app, sheetId) {
  const sheet = selectSheet(app && app.uiDataModel, sheetId);
  const runtimeSheet = app && app.sheets && app.sheets[sheetId];
  if (!sheet) return;
  if (runtimeSheet) {
    runtimeSheet.selectedArgNode = sheet.selectedArgNode;
    runtimeSheet.visualOnly = sheet.visualOnly;
  }
  if (app && app.activeSheetId === sheetId) {
    app.selectedArgNode = sheet.selectedArgNode;
  }
}

function renderArgGraph(app, sheetId) {
  const sheet = selectSheet(app && app.uiDataModel, sheetId);
  const view = selectArgGraphView(sheet);
  const { argGraph } = sheetGraphs(app, sheetId);
  if (argGraph && typeof argGraph.update === 'function') {
    argGraph.update(view.elements, view.positions);
  }
  if (argGraph && typeof argGraph.highlightNode === 'function') {
    if (view.selectedNodeId) {
      argGraph.highlightNode(view.selectedNodeId);
    } else if (typeof argGraph.clearHighlights === 'function') {
      argGraph.clearHighlights();
    }
  }
}

function renderConceptGraph(app, sheetId) {
  const sheet = selectSheet(app && app.uiDataModel, sheetId);
  const view = selectConceptGraphView(sheet);
  const { conceptGraph } = sheetGraphs(app, sheetId);
  if (conceptGraph && typeof conceptGraph.update === 'function') {
    conceptGraph.update(view.elements, view.positions);
  }
  if (typeof app._applyEdgeVisibility === 'function') {
    app._applyEdgeVisibility(conceptGraph, view);
  }
}

function renderActivePanels(app, sheetId, {
  doc = globalThis.document,
} = {}) {
  if (!app || app.activeSheetId !== sheetId) return;
  const sheet = selectSheet(app.uiDataModel, sheetId);
  renderStateCheckboxes(app, selectStateCheckboxRows(sheet), { doc });
  populateConstraintFacts(app, { facts: selectConstraintFacts(sheet) }, { doc });
  updateStateLabel(selectStateLabel(sheet), { doc });
}

export function renderUIDataChange(app, change, options = {}) {
  if (!app || !change || !change.sheetId) return;
  const changed = new Set(change.changed || []);
  syncRuntimeSheetMirrors(app, change.sheetId);
  if (changed.has('arg')) {
    renderArgGraph(app, change.sheetId);
  } else if (changed.has('selection')) {
    const sheet = selectSheet(app.uiDataModel, change.sheetId);
    const view = selectArgGraphView(sheet);
    const { argGraph } = sheetGraphs(app, change.sheetId);
    if (argGraph && typeof argGraph.highlightNode === 'function') {
      if (view.selectedNodeId) {
        argGraph.highlightNode(view.selectedNodeId);
      } else if (typeof argGraph.clearHighlights === 'function') {
        argGraph.clearHighlights();
      }
    }
  }
  if (changed.has('concept') || changed.has('selection')) {
    renderConceptGraph(app, change.sheetId);
  }
  if (changed.has('concept') || changed.has('selection') || changed.has('sheet')) {
    renderActivePanels(app, change.sheetId, options);
  }
}

export function installUIDataModelStore(app, options = {}) {
  if (!app || !app.uiDataModel) return null;
  const store = app.uiDataStore || createUIDataModelStore(app.uiDataModel);
  app.uiDataStore = store;
  if (!app._uiDataRenderUnsubscribe) {
    app._uiDataRenderUnsubscribe = store.subscribe((change) => {
      renderUIDataChange(app, change, options);
    });
  }
  return store;
}

export function applyArgSnapshot(app, sheetId, payload) {
  const id = sheetId || (app && app.activeSheetId) || 'sheet-1';
  if (app && app.uiDataStore) return app.uiDataStore.applyArgSnapshot(id, payload || {});
  if (app && typeof app.acceptArgSnapshot === 'function') return app.acceptArgSnapshot(id, payload || {});
  return null;
}

export function applyConceptSnapshot(app, sheetId, payload) {
  const id = sheetId || (app && app.activeSheetId) || 'sheet-1';
  if (app && app.uiDataStore) return app.uiDataStore.applyConceptSnapshot(id, payload || {});
  if (app && typeof app.acceptConceptSnapshot === 'function') return app.acceptConceptSnapshot(id, payload || {});
  return null;
}

export function applyCtiSnapshot(app, sheetId, payload) {
  const id = sheetId || (app && app.activeSheetId) || 'sheet-1';
  if (app && app.uiDataStore) return app.uiDataStore.applyCtiSnapshot(id, payload || {});
  if (app && typeof app.acceptCtiSnapshot === 'function') return app.acceptCtiSnapshot(id, payload || {});
  return null;
}
