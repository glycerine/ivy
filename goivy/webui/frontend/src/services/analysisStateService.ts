import { applyArgSnapshot, applyConceptSnapshot } from './uiDataRenderService.ts';
import {
  selectConceptSelections,
  selectSheet,
  selectStateToggles,
} from '../models/uiDataSelectors.ts';
import { analysisStateSuggestedName, saveMimeType, savePickerOptions } from './saveDialogService.ts';

function rawRecord(value) {
  return value && typeof value === 'object' && !Array.isArray(value) ? value : {};
}

function graphPayload(snapshot, positions) {
  const payload = snapshot ? { ...rawRecord(snapshot.raw) } : { elements: [] };
  payload.positions = positions && Object.keys(positions).length > 0 ? positions : null;
  return payload;
}

export function analysisStateLimits() {
  return {
    maxFileBytes: 25 * 1024 * 1024,
    maxSheets: 100,
    maxGraphElements: 50000,
    maxEvents: 100000,
    maxEventDepth: 200,
  };
}

export function buildAnalysisState(app, persist) {
  const sheets = [];
  const ids = Object.keys(app.sheets || {});
  ids.sort((a, b) => {
    if (a === 'sheet-1') return -1;
    if (b === 'sheet-1') return 1;
    return a.localeCompare(b);
  });
  for (const sheetId of ids) {
    const sheet = app.sheets[sheetId];
    if (!sheet) continue;
    if (sheet.type === 'events') {
      sheets.push({
        id: sheetId,
        type: 'events',
        label: app.tabLabelForSheet(sheetId),
        events: sheet.events || [],
        patterns: sheet.patterns || [],
        selectedEventAddress: sheet.selectedEventAddress || null,
      });
    } else {
      const modelSheet = selectSheet(app.uiDataModel, sheetId);
      const reachabilityOnly = modelSheet ? modelSheet.reachabilityOnly : !!sheet.reachabilityOnly;
      sheets.push({
        id: sheetId,
        type: 'analysis',
        label: app.tabLabelForSheet(sheetId),
        ...(reachabilityOnly ? { reachabilityOnly: true } : {}),
        selectedArgNode: modelSheet ? modelSheet.selectedArgNode : null,
        conceptSelections: selectConceptSelections(modelSheet),
        arg: modelSheet ? graphPayload(modelSheet.arg, modelSheet.argPositions) : { elements: [], positions: null },
        concept: modelSheet ? graphPayload(modelSheet.concept, modelSheet.conceptPositions) : { elements: [], positions: null },
      });
    }
  }
  const activeSheet = selectSheet(app.uiDataModel, app.activeSheetId || 'sheet-1');
  return {
    analysis_state_format: 'ivyweb-json',
    analysis_state_version: 1,
    python_a2g_equivalent: false,
    fileName: app._persistedFileName || '',
    filePath: app._persistedFilePath || app._persistedFileName || '',
    fileContent: app._editorContent ? app._editorContent() : (app._persistedFileContent || ''),
    activeIsolate: app.activeIsolate || '',
    availableIsolates: Array.isArray(app.availableIsolates) ? app.availableIsolates.slice() : [],
    editorKeymap: app.getEditorKeymap ? app.getEditorKeymap() : (app._editorKeymap || 'emacs'),
    mode: app.getMode(),
    activeSheetId: app.activeSheetId || 'sheet-1',
    selectedArgNode: activeSheet ? activeSheet.selectedArgNode : null,
    toggles: persist && persist.getToggles ? persist.getToggles(app) : selectStateToggles(activeSheet),
    sheets,
  };
}

export function validateAnalysisStateEvents(events, depth, count, limits) {
  if (!Array.isArray(events)) {
    throw new Error('event children must be an array');
  }
  if (depth > limits.maxEventDepth) {
    throw new Error('event tree too deep');
  }
  for (const ev of events) {
    count.events += 1;
    if (count.events > limits.maxEvents) {
      throw new Error('too many events in analysis state');
    }
    const event = ev || {};
    if (event.address != null && !/^\d+(\/\d+)*$/.test(String(event.address))) {
      throw new Error(`invalid event address: ${event.address}`);
    }
    if (event.subs != null) {
      validateAnalysisStateEvents(event.subs, depth + 1, count, limits);
    }
  }
}

export function validateAnalysisStateGraphPayload(graph, name, limits) {
  if (!graph) return;
  if (graph.elements != null && !Array.isArray(graph.elements)) {
    throw new Error(`${name} graph elements must be an array`);
  }
  if (graph.elements && graph.elements.length > limits.maxGraphElements) {
    throw new Error(`${name} graph has too many elements`);
  }
}

export function validateAnalysisStateSheet(sheet, seen, limits, isValidSheetId) {
  if (!sheet || typeof sheet !== 'object') {
    throw new Error('invalid sheet entry');
  }
  if (!isValidSheetId(sheet.id)) {
    throw new Error(`invalid sheet id: ${sheet.id}`);
  }
  if (seen[sheet.id]) {
    throw new Error(`duplicate sheet id: ${sheet.id}`);
  }
  seen[sheet.id] = true;
  if (sheet.type !== 'analysis' && sheet.type !== 'events') {
    throw new Error(`invalid sheet type: ${sheet.type}`);
  }
  if (sheet.reachabilityOnly != null && typeof sheet.reachabilityOnly !== 'boolean') {
    throw new Error('sheet reachabilityOnly must be a boolean');
  }
  if (sheet.type === 'events') {
    if (sheet.events != null && !Array.isArray(sheet.events)) {
      throw new Error('event sheet events must be an array');
    }
    if (sheet.patterns != null && !Array.isArray(sheet.patterns)) {
      throw new Error('event sheet patterns must be an array');
    }
    validateAnalysisStateEvents(sheet.events || [], 0, { events: 0 }, limits);
    return;
  }
  validateAnalysisStateGraphPayload(sheet.arg, 'arg', limits);
  validateAnalysisStateGraphPayload(sheet.concept, 'concept', limits);
}

export function validateAnalysisStateObject(app, state) {
  if (!state || state.analysis_state_format !== 'ivyweb-json') {
    throw new Error('unsupported analysis state format');
  }
  const limits = app.analysisStateLimits();
  const sheets = state.sheets || [];
  if (!Array.isArray(sheets)) {
    throw new Error('analysis state sheets must be an array');
  }
  if (sheets.length > limits.maxSheets) {
    throw new Error('too many sheets in analysis state');
  }
  if (state.activeSheetId && !app.isValidSheetId(state.activeSheetId)) {
    throw new Error(`invalid active sheet id: ${state.activeSheetId}`);
  }
  const seen = {};
  for (const sheet of sheets) {
    validateAnalysisStateSheet(sheet, seen, limits, app.isValidSheetId.bind(app));
  }
}

export function removeAnalysisStateExtraSheets(app) {
  const ids = Object.keys(app.sheets || {});
  for (const id of ids) {
    if (id !== 'sheet-1') {
      app.removeSheet(id);
    }
  }
}

export async function saveAnalysisState(app, {
  win = globalThis.window,
} = {}) {
  try {
    const state = app.buildAnalysisState();
    const text = `${JSON.stringify(state, null, 2)}\n`;
    const suggestedName = analysisStateSuggestedName(app._persistedFileName);
    if (win && win.showSaveFilePicker) {
      const handle = await win.showSaveFilePicker(savePickerOptions('analysisState', suggestedName));
      const writable = await handle.createWritable();
      await writable.write(text);
      await writable.close();
      app.controls.setStatus(`Analysis state saved: ${handle.name}`, 'success');
    } else {
      app.downloadTextFile(suggestedName, text, saveMimeType('analysisState'));
      app.controls.setStatus(`Analysis state downloaded: ${suggestedName}`, 'success');
    }
    return state;
  } catch (err) {
    if (err.name === 'AbortError') {
      app.controls.setStatus('Save analysis state cancelled');
    } else {
      app.controls.setStatus(`Save analysis state failed: ${err.message}`, 'error');
    }
    return null;
  }
}

export async function loadAnalysisStateFile(app, file) {
  if (!file) return false;
  if (typeof file.size === 'number' && file.size > app.analysisStateLimits().maxFileBytes) {
    throw new Error('analysis state file too large');
  }
  const text = await app.readFileText(file);
  return app.loadAnalysisStateObject(JSON.parse(text));
}

export async function loadAnalysisStateObject(app, state, persist, options: any = {}) {
  app.validateAnalysisStateObject(state);
  if (!state || state.analysis_state_format !== 'ivyweb-json') {
    throw new Error('unsupported analysis state format');
  }
  const preservePrimarySheetModel = !!options.preservePrimarySheetModel;
  app._persistedFileName = state.fileName || '';
  app._persistedFilePath = state.filePath || state.fileName || '';
  app._persistedFileContent = state.fileContent || '';
  app._savedFileContent = app._persistedFileContent;

  if (app.setEditorContent) {
    app.setEditorContent(app._persistedFileContent);
  }
  if (state.mode) app.setMode(state.mode);
  if (state.editorKeymap && app.setEditorKeymap) app.setEditorKeymap(state.editorKeymap, { save: false });
  if (app.setIsolates) app.setIsolates(state.availableIsolates || [], state.activeIsolate || '');
  if (!options.skipReloadContent && app.api && app.api.reloadContent && app._persistedFileContent) {
    const loadResult = await app.api.reloadContent(
      app._persistedFileContent,
      app._persistedFilePath || app._persistedFileName || 'restored.ivy',
      { isolate: state.activeIsolate || '' },
    );
    if (app.setIsolates) {
      app.setIsolates(
        (loadResult && loadResult.isolates) || state.availableIsolates || [],
        (loadResult && loadResult.isolate) || state.activeIsolate || '',
      );
    }
  }

  app.removeAnalysisStateExtraSheets();
  const sheets = state.sheets || [];
  for (const sheet of sheets) {
    if (sheet.type === 'events') {
      app.openEventTraceSheet(sheet.label || 'Events', {
        sheet_id: sheet.id,
        events: sheet.events || [],
        patterns: sheet.patterns || [],
        selected_address: sheet.selectedEventAddress || null,
      }, sheet.id);
      app.setVisualOnlySheet(sheet.id, true);
      continue;
    }
    if (sheet.id === 'sheet-1') {
      if (!preservePrimarySheetModel && sheet.arg && sheet.arg.elements && app.argGraph) {
        applyArgSnapshot(app, sheet.id, sheet.arg);
      }
      if (!preservePrimarySheetModel && sheet.concept && sheet.concept.elements && app.conceptGraph) {
        applyConceptSnapshot(app, sheet.id, sheet.concept);
      }
      if (app.sheets && app.sheets['sheet-1']) {
        app.sheets['sheet-1'].selectedArgNode = sheet.selectedArgNode || null;
        app.sheets['sheet-1'].visualOnly = !preservePrimarySheetModel;
      }
      if (app.uiDataStore) app.uiDataStore.setSelectedArgNode(sheet.id, sheet.selectedArgNode || null);
      if (app.uiDataStore) app.uiDataStore.setConceptSelections(sheet.id, sheet.conceptSelections || []);
    } else if (sheet.arg && sheet.arg.elements) {
      if (sheet.reachabilityOnly) {
        app.openARGSheet(sheet.label || sheet.id, sheet.arg, sheet.id, { reachabilityOnly: true });
      } else {
        app.openARGSheet(sheet.label || sheet.id, sheet.arg, sheet.id);
      }
      const opened = app.sheets && app.sheets[sheet.id];
      if (opened && opened.conceptGraph && sheet.concept && sheet.concept.elements) {
        applyConceptSnapshot(app, sheet.id, sheet.concept);
      }
      if (opened) {
        opened.selectedArgNode = sheet.selectedArgNode || null;
        opened.visualOnly = true;
      }
      if (app.uiDataStore) app.uiDataStore.setSelectedArgNode(sheet.id, sheet.selectedArgNode || null);
      if (app.uiDataStore) app.uiDataStore.setConceptSelections(sheet.id, sheet.conceptSelections || []);
    }
  }

  if (state.toggles && persist && persist.applyToggles) {
    await persist.applyToggles(app, state.toggles);
  }
  if (state.activeSheetId && app.sheetExists(state.activeSheetId)) {
    app.switchSheet(state.activeSheetId);
  } else {
    app.switchSheet('sheet-1');
  }
  if (persist && persist.setFileName) {
    persist.setFileName(app._persistedFileName, app._persistedFilePath);
  }
  if (!options.preserveStatus) {
    app.controls.setStatus(`Visual analysis state loaded: ${app._persistedFileName || 'state'}`, 'warning');
  }
  return true;
}
