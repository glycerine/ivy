import { installUIDataModelStore } from './uiDataRenderService.ts';
import { selectSheet } from '../models/uiDataSelectors.ts';

function activeSheetMenuContext(app, sheetId) {
  const runtimeSheet = app && app.sheets && app.sheets[sheetId];
  const modelSheet = selectSheet(app && app.uiDataModel, sheetId);
  const type = (runtimeSheet && runtimeSheet.type) || (modelSheet && modelSheet.type) || 'analysis';
  const reachabilityOnly = type !== 'events' && !!(
    (runtimeSheet && runtimeSheet.reachabilityOnly) ||
    (modelSheet && modelSheet.reachabilityOnly)
  );
  return { type, reachabilityOnly };
}

function updateActiveSheetMenuContext(app, sheetId, doc) {
  const body = doc && doc.body;
  if (!body) return;
  const context = activeSheetMenuContext(app, sheetId);
  body.setAttribute('data-active-sheet-type', context.type);
  body.setAttribute('data-active-sheet-reachability-only', context.reachabilityOnly ? 'true' : 'false');
}

function refreshActiveSheetMenus(app) {
  if (!app || typeof app.loadMenuDescriptors !== 'function') return;
  try {
    const refresh = app.loadMenuDescriptors();
    if (refresh && typeof refresh.catch === 'function') refresh.catch(() => {});
  } catch {
    // loadMenuDescriptors reports user-visible failures itself when available.
  }
}

export function isVisualOnlySheet(app, sheetId) {
  const modelSheet = app.uiDataModel && app.uiDataModel.sheets && app.uiDataModel.sheets[sheetId];
  if (modelSheet) return !!modelSheet.visualOnly;
  const sheet = app.sheets && app.sheets[sheetId];
  return !!(sheet && sheet.visualOnly);
}

export function setVisualOnlySheet(app, sheetId, visualOnly) {
  const sheet = app.sheets && app.sheets[sheetId];
  if (sheet) sheet.visualOnly = !!visualOnly;
  if (app.uiDataModel) {
    const store = app.uiDataStore || installUIDataModelStore(app);
    if (store) store.setVisualOnly(sheetId, !!visualOnly, { type: sheet && sheet.type });
  }
}

export function visualOnlyMessage(kind) {
  if (kind === 'events') {
    return 'Restored event trace is visual-only; reload or rerun analysis before event backend actions';
  }
  return 'Restored analysis state is visual-only; reload or rerun analysis before graph actions';
}

export function isValidSheetId(sheetId) {
  return /^[A-Za-z][A-Za-z0-9_-]*$/.test(String(sheetId || ''));
}

export function assertValidSheetId(sheetId) {
  if (!isValidSheetId(sheetId)) {
    throw new Error(`invalid sheet id: ${sheetId}`);
  }
  return sheetId;
}

export function sheetTab(sheetId, doc = globalThis.document) {
  const tabs = doc ? doc.querySelectorAll('.sheet-tab') : [];
  for (const tab of tabs) {
    if (tab.getAttribute('data-sheet') === sheetId) {
      return tab;
    }
  }
  return null;
}

export function sheetExists(app, sheetId, doc = globalThis.document) {
  return !!((app.sheets && app.sheets[sheetId]) || (doc && doc.getElementById(sheetId)) || sheetTab(sheetId, doc));
}

export function switchSheet(app, sheetId, {
  doc = globalThis.document,
} = {}) {
  const tabs = doc.querySelectorAll('.sheet-tab');
  const sheets = doc.querySelectorAll('.sheet-content');
  for (const tab of tabs) tab.classList.remove('active');
  for (const sheet of sheets) {
    sheet.classList.remove('active');
  }

  const tab = app.sheetTab(sheetId);
  const sheet = doc.getElementById(sheetId);
  if (tab) tab.classList.add('active');
  if (sheet) sheet.classList.add('active');
  if (app.sheets && app.sheets[sheetId]) {
    const runtimeSheet = app.sheets[sheetId];
    if (runtimeSheet.type !== 'events') {
      app.argGraph = runtimeSheet.argGraph;
      app.conceptGraph = runtimeSheet.conceptGraph;
      const modelSheet = app.uiDataModel && app.uiDataModel.sheets[sheetId];
      app.selectedArgNode = modelSheet ? modelSheet.selectedArgNode : runtimeSheet.selectedArgNode;
    } else {
      app.argGraph = null;
      app.conceptGraph = null;
    }
    app.activeSheetId = sheetId;
    if (app.uiDataModel) {
      const store = app.uiDataStore || installUIDataModelStore(app);
      if (store) {
        store.registerSheet(sheetId, {
          type: runtimeSheet.type,
          reachabilityOnly: !!runtimeSheet.reachabilityOnly,
        });
        store.setActiveSheet(sheetId);
      }
    }
    if (runtimeSheet.type !== 'events') {
      if (runtimeSheet.argGraph) runtimeSheet.argGraph.resize();
      if (runtimeSheet.conceptGraph) runtimeSheet.conceptGraph.resize();
    }
    updateActiveSheetMenuContext(app, sheetId, doc);
    refreshActiveSheetMenus(app);
    return;
  }
}
