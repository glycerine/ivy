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
  const modelSheet = selectSheet(app && app.uiDataModel, sheetId);
  const kind = (runtimeSheet && runtimeSheet.type) || (modelSheet && modelSheet.type);
  if (kind === 'events') {
    return { runtimeSheet, argGraph: null, conceptGraph: null };
  }
  if (runtimeSheet) {
    return {
      runtimeSheet,
      argGraph: runtimeSheet.argGraph || null,
      conceptGraph: runtimeSheet.conceptGraph || null,
    };
  }
  return {
    runtimeSheet,
    argGraph: app && app.argGraph,
    conceptGraph: app && app.conceptGraph,
  };
}

function isAnalysisSheet(app, sheetId) {
  const runtimeSheet = app && app.sheets && app.sheets[sheetId];
  const modelSheet = selectSheet(app && app.uiDataModel, sheetId);
  const kind = (runtimeSheet && runtimeSheet.type) || (modelSheet && modelSheet.type) || 'analysis';
  return kind !== 'events';
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
  if (!isAnalysisSheet(app, sheetId)) return;
  const sheet = selectSheet(app && app.uiDataModel, sheetId);
  const view = selectArgGraphView(sheet);
  const { argGraph } = sheetGraphs(app, sheetId);
  if (argGraph && typeof argGraph.update === 'function') {
    const positions = argGraph.update(view.elements, view.positions);
    if (positions && app && app.uiDataStore && typeof app.uiDataStore.setGraphPositions === 'function') {
      app.uiDataStore.setGraphPositions(sheetId, 'arg', positions, { emit: false });
    }
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
  if (!isAnalysisSheet(app, sheetId)) return;
  const sheet = selectSheet(app && app.uiDataModel, sheetId);
  const view = selectConceptGraphView(sheet);
  const { conceptGraph } = sheetGraphs(app, sheetId);
  if (conceptGraph && typeof conceptGraph.update === 'function') {
    const positions = conceptGraph.update(view.elements, view.positions);
    if (positions && app && app.uiDataStore && typeof app.uiDataStore.setGraphPositions === 'function') {
      app.uiDataStore.setGraphPositions(sheetId, 'concept', positions, { emit: false });
    }
  }
  if (conceptGraph && typeof app._applyEdgeVisibility === 'function') {
    app._applyEdgeVisibility(conceptGraph, view);
  }
}

function renderActivePanels(app, sheetId, {
  doc = globalThis.document,
} = {}) {
  if (!app || app.activeSheetId !== sheetId) return;
  if (!isAnalysisSheet(app, sheetId)) return;
  const sheet = selectSheet(app.uiDataModel, sheetId);
  renderStateCheckboxes(app, selectStateCheckboxRows(sheet), { doc });
  populateConstraintFacts(app, { facts: selectConstraintFacts(sheet) }, { doc });
  updateStateLabel(selectStateLabel(sheet), { doc });
}

export function renderUIDataChange(app, change, options = {}) {
  if (!app || !change || !change.sheetId) return;
  const changed = new Set(change.changed || []);
  syncRuntimeSheetMirrors(app, change.sheetId);
  if (changed.has('arg') || changed.has('argLayout')) {
    renderArgGraph(app, change.sheetId);
  } else if (changed.has('argSelection')) {
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
  if (changed.has('concept') || changed.has('conceptSelection') || changed.has('conceptLayout')) {
    renderConceptGraph(app, change.sheetId);
  }
  if (changed.has('concept') || changed.has('conceptSelection') || changed.has('sheet')) {
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
  const store = app && (app.uiDataStore || installUIDataModelStore(app));
  if (!store) throw new Error('UIDataModel store is required to apply ARG snapshots');
  return store.applyArgSnapshot(id, payload || {});
}

export function applyConceptSnapshot(app, sheetId, payload) {
  const id = sheetId || (app && app.activeSheetId) || 'sheet-1';
  const store = app && (app.uiDataStore || installUIDataModelStore(app));
  if (!store) throw new Error('UIDataModel store is required to apply concept snapshots');
  return store.applyConceptSnapshot(id, payload || {});
}

export function applyCtiSnapshot(app, sheetId, payload) {
  const id = sheetId || (app && app.activeSheetId) || 'sheet-1';
  const store = app && (app.uiDataStore || installUIDataModelStore(app));
  if (!store) throw new Error('UIDataModel store is required to apply CTI snapshots');
  return store.applyCtiSnapshot(id, payload || {});
}
