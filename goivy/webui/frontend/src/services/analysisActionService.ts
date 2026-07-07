import { applyConceptSnapshot } from './uiDataRenderService.ts';
import { saveMimeType, savePickerOptions } from './saveDialogService.ts';
import { normalizeArgNodeId } from '../models/uiDataModel.ts';

function activeSheetId(app) {
  return (app && app.activeSheetId) || 'sheet-1';
}

function sheetForRefresh(app, sheetId) {
  const modelSheet = app && app.uiDataModel && app.uiDataModel.sheets && app.uiDataModel.sheets[sheetId];
  const runtimeSheet = app && app.sheets && app.sheets[sheetId];
  return {
    modelSheet,
    runtimeSheet,
    reachabilityOnly: !!((modelSheet && modelSheet.reachabilityOnly) || (runtimeSheet && runtimeSheet.reachabilityOnly)),
    visualOnly: !!((modelSheet && modelSheet.visualOnly) || (runtimeSheet && runtimeSheet.visualOnly)),
  };
}

function selectedArgNodeForRefresh(app, sheetId, modelSheet, runtimeSheet) {
  if (modelSheet && modelSheet.selectedArgNode) return modelSheet.selectedArgNode;
  if (runtimeSheet && runtimeSheet.selectedArgNode) return runtimeSheet.selectedArgNode;
  return app && app.selectedArgNode;
}

export async function runAction(app, actionName, args, options) {
  const opts = options || {};
  const runningMessage = opts.runningMessage || `Running: ${actionName}...`;
  const successMessage = opts.successMessage || `Done: ${actionName}`;
  const failurePrefix = opts.failurePrefix || 'Action failed';
  if (opts.showLoading !== false) {
    app.controls.showLoading(runningMessage);
  }
  app.controls.setStatus(runningMessage, 'info');
  try {
    const result = await app.api.executeAction(actionName, args || {});
    if (opts.successStatus !== false) {
      app.controls.setStatus(successMessage, 'success');
    }
    return { ok: true, result };
  } catch (err) {
    const message = `${failurePrefix}: ${err.message}`;
    app.controls.setStatus(message, 'error');
    if (opts.showDialog) {
      app.showTextDialog('ivyweb', failurePrefix, err.message);
    }
    return { ok: false, error: err.message };
  } finally {
    if (opts.showLoading !== false) {
      app.controls.hideLoading();
    }
  }
}

export async function refreshConceptGraph(app) {
  const sheetId = activeSheetId(app);
  const { modelSheet, runtimeSheet, reachabilityOnly, visualOnly } = sheetForRefresh(app, sheetId);
  if (reachabilityOnly || visualOnly) return null;
  const selectedArgNode = normalizeArgNodeId(selectedArgNodeForRefresh(app, sheetId, modelSheet, runtimeSheet)) || undefined;
  try {
    const result = await app.api.getConceptGraph(selectedArgNode, sheetId);
    if (result && result.elements) {
      applyConceptSnapshot(app, sheetId, result);
    } else if (result) {
      app.populateStateCheckboxes(result);
    }
    return result;
  } catch (err) {
    console.error('Concept graph refresh error:', err);
    return null;
  }
}

export async function executeAndRefresh(app, {
  action,
  args = {},
  running,
  success,
  failure,
}) {
  app.controls.setStatus(running);
  try {
    const result = await app.api.executeAction(action, args);
    await app.refreshConceptGraph();
    app.controls.setStatus(success, 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`${failure}: ${err.message}`, 'error');
    return null;
  }
}

export async function rememberGraph(app) {
  app.controls.setStatus('Remembering graph...');
  try {
    const name = await app.entryDialog('Remember graph', 'Enter a name for this goal:', '', { okLabel: 'Remember' });
    if (name === null) {
      app.controls.setStatus('Remember cancelled', 'warning');
      return undefined;
    }
    const result = await app.api.executeAction('remember', { name, sheet_id: app.activeSheetId || 'sheet-1' });
    app.controls.setStatus('Graph remembered', 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Remember failed: ${err.message}`, 'error');
    return null;
  }
}

export async function exportConjecture(app, {
  win = globalThis.window,
  doc = globalThis.document,
  urlApi = globalThis.URL,
} = {}) {
  app.controls.setStatus('Exporting conjecture...');
  try {
    const result = await app.api.executeAction('export', { sheet_id: app.activeSheetId || 'sheet-1' });
    const content = (result && result.content) || '';
    const filename = (result && result.filename) || 'concept_graph.dot';
    const mimeType = (result && result.mime_type) || saveMimeType('dot');
    if (content) {
      if (win && win.showSaveFilePicker) {
        const handle = await win.showSaveFilePicker(savePickerOptions('dot', filename));
        const writable = await handle.createWritable();
        await writable.write(content);
        await writable.close();
      } else if (typeof app.downloadTextFile === 'function') {
        app.downloadTextFile(filename, content, mimeType);
      } else {
        const blob = new Blob([content], { type: mimeType });
        const url = urlApi.createObjectURL(blob);
        const a = doc.createElement('a');
        a.href = url;
        a.download = filename;
        doc.body.appendChild(a);
        a.click();
        doc.body.removeChild(a);
        urlApi.revokeObjectURL(url);
      }
    }
    app.controls.setStatus('Graph exported', 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Export failed: ${err.message}`, 'error');
    return null;
  }
}
