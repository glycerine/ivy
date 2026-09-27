import { applyArgSnapshot, applyConceptSnapshot } from './uiDataRenderService.ts';
import { addTraceResultViewAction, analysisControllerOptions, openTraceArgFromResult } from './checkService.ts';
import { runWithContext } from './runContextService.ts';
import { openSourceBrowser as openSourceBrowserViaService } from './sourceBrowserService.ts';
import { clearDetailsLog } from './detailsService.ts';

async function applyDescriptorDialog(app, actionName, args) {
  const dialog = args && args.dialog;
  if (!dialog) return args;
  delete args.dialog;
  const dialogType = dialog.type || dialog.kind || 'select';
  const title = dialog.title || actionName || 'Choose';
  const prompt = dialog.prompt || dialog.message || 'Choose:';
  const argName = dialog.arg || dialog.arg_name || 'selection';
  const items = Array.isArray(dialog.options) ? dialog.options : [];
  if (dialogType === 'select_multiple' || dialogType === 'multi_select' || dialog.multiple === true) {
    const selected = await app.listboxDialog(title, prompt, items, {
      multiple: true,
      okLabel: dialog.okLabel || dialog.ok_label || 'OK',
    });
    if (selected === null) return null;
    args[argName] = selected || [];
    return args;
  }
  const selected = await app.listboxDialog(title, prompt, items, {
    okLabel: dialog.okLabel || dialog.ok_label || 'OK',
    cancel: dialog.cancel !== false,
  });
  if (selected === null) return null;
  args[argName] = selected;
  return args;
}

export async function prepareArgNodeActionArgs(app, nodeData, actionName, args, sheetId) {
  args = await applyDescriptorDialog(app, actionName, args);
  if (args === null) return null;
  if (actionName === 'try_conjecture' && !args.conjecture) {
    const conjChoices = await app.api.argNodeAction(nodeData.obj || nodeData.id, 'try_conjecture_choices', { sheet_id: sheetId });
    const selectedConj = await app.listboxDialog(
      'Try conjecture',
      'Choose a conjecture to prove:',
      conjChoices.choices || [],
      { cancel: true },
    );
    if (selectedConj == null) return null;
    args.conjecture = selectedConj;
  } else if (actionName === 'try_remembered' && !args.goal) {
    const goalChoices = await app.api.argNodeAction(nodeData.obj || nodeData.id, 'try_remembered_choices', { sheet_id: sheetId });
    const selectedGoal = await app.listboxDialog(
      'Try remembered goal',
      'Choose a remembered goal:',
      goalChoices.choices || [],
      { cancel: true },
    );
    if (selectedGoal == null) return null;
    args.goal = selectedGoal;
  } else if (actionName === 'bmc') {
    if (args.bound == null) {
      const initialBound = Number.isInteger(app.currentBound) && app.currentBound >= 0 ? app.currentBound : 10;
      const bound = await app.integerDialog('Bounded check', 'Enter bound:', initialBound, {
        min: 0,
        okLabel: 'Check',
      });
      if (bound === null) return null;
      args.bound = bound;
      app.currentBound = bound;
    }
    if (!args.err_cond && !args.error_condition) {
      const errCond = await app.entryDialog('Bounded check', 'Enter error condition:', 'true', {
        okLabel: 'Check',
      });
      if (errCond === null) return null;
      args.err_cond = errCond || 'true';
    }
  } else if (actionName === 'check_safety' && !args.mode && typeof app.getMode === 'function') {
    args.mode = app.getMode();
  }
  return args;
}

function showArgNodeSafetyResult(app, result) {
  const message = (result && result.message) || 'Safety check complete';
  if (result && result.safe === true) {
    app.controls.setStatus(message, 'success');
    return true;
  }
  app.controls.setStatus(message, 'error');
  if (app.controls && typeof app.controls.showInfo === 'function') {
    app.controls.showInfo('Safety Check', message);
    addTraceResultViewAction(app, result);
  }
  return true;
}

function splitBmcMessage(message) {
  const lines = String(message || 'Bounded check complete').split('\n');
  const dialogMessage = lines.shift() || 'Bounded check complete';
  return {
    dialogMessage,
    dialogText: lines.join('\n'),
  };
}

async function showArgNodeBmcResult(app, result) {
  const message = (result && result.message) || 'Bounded check complete';
  const found = !!(result && (result.reachable === true || result.found === true || result.result === 'fail'));
  app.controls.setStatus(message, found ? 'error' : 'success');
  const { dialogMessage, dialogText } = splitBmcMessage(message);
  if (found && result && result.trace_arg && typeof app.textDialog === 'function') {
    const action = await app.textDialog('ivyweb', dialogMessage, dialogText, {
      okLabel: 'View',
      cancel: true,
      primaryFirst: true,
    });
    if (action !== null) {
      openTraceArgFromResult(app, result, { label: result.trace_label || 'BMC counterexample' });
    }
    return true;
  }
  if (typeof app.okDialog === 'function') {
    await app.okDialog('ivyweb', message);
  } else if (app.controls && typeof app.controls.showInfo === 'function') {
    app.controls.showInfo('Bounded Check', message);
  }
  return true;
}

async function showArgNodeExtendResult(app, result) {
  if (!result || !result.closed) return false;
  const message = result.message || 'State is closed.';
  if (typeof app.okDialog === 'function') {
    await app.okDialog('ivyweb', message);
  } else if (app.controls && typeof app.controls.showInfo === 'function') {
    app.controls.showInfo('Extend', message);
  }
  app.controls.setStatus(message, 'warning');
  return true;
}

function showReturnedSource(app, result) {
  if (!result || !result.source) return false;
  if (typeof app.openSourceBrowser === 'function') {
    app.openSourceBrowser(result);
  } else {
    openSourceBrowserViaService(app, result);
  }
  if (app.controls && typeof app.controls.showInfo === 'function') {
    app.controls.showInfo(
      `Source: ${result.file || ''}${result.lineno ? ` line ${result.lineno}` : ''}`,
      '',
    );
  }
  return true;
}

function showArgNodeExhaustedResult(app, result) {
  if (!result || result.exhausted !== true) return false;
  const message = result.message || 'No valid action could be found.';
  app.controls.setStatus(message, 'warning');
  return true;
}

function fieldValue(obj, names, fallback = '') {
  if (!obj) return fallback;
  for (const name of names) {
    if (Object.prototype.hasOwnProperty.call(obj, name) && obj[name] !== undefined && obj[name] !== null) {
      return obj[name];
    }
  }
  return fallback;
}

function safeStableString(value) {
  if (value === undefined || value === null) return '';
  if (typeof value === 'string') return value;
  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
}

function snapshotElements(snapshot) {
  if (!snapshot) return null;
  if (Array.isArray(snapshot.elements)) return snapshot.elements;
  if (snapshot.elements && typeof snapshot.elements === 'object') {
    const nodes = Array.isArray(snapshot.elements.nodes) ? snapshot.elements.nodes : [];
    const edges = Array.isArray(snapshot.elements.edges) ? snapshot.elements.edges : [];
    if (nodes.length || edges.length) return [...nodes, ...edges];
  }
  if (snapshot.render && Array.isArray(snapshot.render.elements)) return snapshot.render.elements;
  return null;
}

function renderedARGSignature(snapshot) {
  const elements = snapshotElements(snapshot);
  if (!elements) return null;
  const keys = elements.map((element) => {
    const data = element && element.data ? element.data : {};
    return JSON.stringify({
      group: fieldValue(element, ['group']),
      classes: fieldValue(element, ['classes']),
      id: fieldValue(data, ['id']),
      obj: fieldValue(data, ['obj']),
      source: fieldValue(data, ['source']),
      target: fieldValue(data, ['target']),
      sourceObj: fieldValue(data, ['source_obj', 'sourceObj']),
      targetObj: fieldValue(data, ['target_obj', 'targetObj']),
      label: fieldValue(data, ['label']),
      shortInfo: fieldValue(data, ['short_info', 'shortInfo']),
      longInfo: safeStableString(fieldValue(data, ['long_info', 'longInfo'], '')),
    });
  });
  keys.sort();
  return JSON.stringify(keys);
}

function currentARGSignature(app, sheetId) {
  const sheet = app && app.uiDataModel && app.uiDataModel.sheets && app.uiDataModel.sheets[sheetId];
  if (!sheet || !sheet.arg) return null;
  return renderedARGSignature(sheet.arg);
}

function argGraphForSheet(app, sheetId) {
  const sheet = app && app.sheets && app.sheets[sheetId];
  if (sheet && sheet.argGraph) return sheet.argGraph;
  if (app && app.activeSheetId === sheetId && app.argGraph) return app.argGraph;
  return null;
}

function graphRenderedARGSignature(graph) {
  const cy = graph && graph.cy;
  if (!cy) return null;
  if (typeof cy.json === 'function') {
    const raw = cy.json();
    const sig = renderedARGSignature(raw);
    if (sig !== null) return sig;
  }
  if (typeof cy.elements === 'function') {
    const collection = cy.elements();
    if (collection && typeof collection.jsons === 'function') {
      return renderedARGSignature({ elements: collection.jsons() });
    }
  }
  return null;
}

function currentVisibleARGSignature(app, sheetId) {
  return graphRenderedARGSignature(argGraphForSheet(app, sheetId));
}

function cleanActionDisplayName(name) {
  return String(name || '')
    .replace(/^ext:/, '')
    .replace(/^call:\s*/, '')
    .replace(/^call\s+/, '')
    .trim();
}

function concreteActionDisplayName(actionName, args) {
  if (actionName === 'execute_action' && args && args.action_name) {
    return cleanActionDisplayName(args.action_name) || 'action';
  }
  return cleanActionDisplayName(actionName) || 'action';
}

function showUnchangedExecuteActionResult(app, actionName, args, beforeSignature, result, appliedSnapshot, afterSignature) {
  if (actionName !== 'execute_action') return false;
  if (beforeSignature === null || beforeSignature === undefined) return false;
  const visibleOrResultSignature = afterSignature || renderedARGSignature(appliedSnapshot || (result && result.arg));
  if (visibleOrResultSignature === null || visibleOrResultSignature !== beforeSignature) return false;
  const message = (result && result.message) || `No new visible transition was added for ${concreteActionDisplayName(actionName, args)}.`;
  app.controls.setStatus(message, 'warning');
  return true;
}

function highlightReturnedSourceInEditor(app, result) {
  if (!result || !result.lineno || typeof app.scrollEditorToLine !== 'function') return false;
  app.scrollEditorToLine(result.lineno);
  return true;
}

async function showArgNodeTryConjectureResult(app, result) {
  const message = (result && result.message) || 'Try conjecture complete';
  showReturnedSource(app, result);
  if (result && result.view === 'trace' && result.trace_arg) {
    app.controls.setStatus(message, 'error');
    if (typeof app.textDialog === 'function') {
      const action = await app.textDialog('ivyweb', message, '', {
        okLabel: 'View',
        cancel: true,
        primaryFirst: true,
      });
      if (action !== null) {
        openTraceArgFromResult(app, result, { label: result.trace_label || 'Try conjecture trace' });
      }
      return true;
    }
    openTraceArgFromResult(app, result, { label: result.trace_label || 'Try conjecture trace' });
    return true;
  }
  if (result && result.view === 'message' && typeof app.okDialog === 'function') {
    await app.okDialog('ivyweb', message);
  }
  app.controls.setStatus(message, 'success');
  return true;
}

export async function executeArgNodeAction(app, nodeData, action, sheetId) {
  const targetSheetId = sheetId || app.activeSheetId || 'sheet-1';
  if (app.isVisualOnlySheet(targetSheetId)) {
    app.controls.setStatus(app.visualOnlyMessage('analysis'), 'warning');
    return null;
  }
  const actionName = action.action || action.id || action[0] || action.name;
  clearDetailsLog(app);
  app.controls.setStatus(`Executing: ${actionName}...`);
  try {
    let args = {
      ...analysisControllerOptions({ includeBound: false, includeRelations: false }),
      ...(action.args || {}),
      sheet_id: targetSheetId,
    };
    args = await app.prepareArgNodeActionArgs(nodeData, actionName, args, targetSheetId);
    if (args === null) {
      app.controls.setStatus(`Action cancelled: ${actionName}`, 'warning');
      return null;
    }
    const beforeArgSignature = actionName === 'execute_action'
      ? (currentVisibleARGSignature(app, targetSheetId) || currentARGSignature(app, targetSheetId))
      : null;
    const result = await runWithContext(app, {
      busyMessage: `Executing: ${actionName}...`,
      failurePrefix: 'Action failed',
    }, () => app.api.argNodeAction(nodeData.obj || nodeData.id, actionName, args));
    if (!result) return null;
    let appliedArgSnapshot = null;
    if (result && result.arg) {
      appliedArgSnapshot = applyArgSnapshot(app, targetSheetId, result.arg);
    }
    if (result && result.concept) {
      applyConceptSnapshot(app, targetSheetId, result.concept);
    }
    if (showArgNodeExhaustedResult(app, result)) {
      return result;
    }
    const afterArgSignature = actionName === 'execute_action' ? currentVisibleARGSignature(app, targetSheetId) : null;
    if (showUnchangedExecuteActionResult(app, actionName, args, beforeArgSignature, result, appliedArgSnapshot, afterArgSignature)) {
      return result;
    }
    if (actionName === 'check_safety') {
      showArgNodeSafetyResult(app, result);
      return result;
    }
    if (actionName === 'bmc') {
      await showArgNodeBmcResult(app, result);
      return result;
    }
    if (actionName === 'try_conjecture') {
      await showArgNodeTryConjectureResult(app, result);
      return result;
    }
    if (actionName === 'find_extension' || actionName === 'extend') {
      if (await showArgNodeExtendResult(app, result)) {
        return result;
      }
    }
    app.controls.setStatus(`Action complete: ${actionName}`, 'success');
    return result;
  } catch (err) {
    await runWithContext(app, {
      busyMessage: `Executing: ${actionName}...`,
      failurePrefix: 'Action failed',
    }, () => {
      throw err;
    });
    console.error('ARG action error:', err);
    return null;
  }
}

export async function executeArgEdgeAction(app, edgeData, actionName, sheetId) {
  const targetSheetId = sheetId || app.activeSheetId || 'sheet-1';
  app.controls.setStatus(`Executing: ${actionName}...`);
  try {
    const result = await runWithContext(app, {
      busyMessage: `Executing: ${actionName}...`,
      failurePrefix: 'Edge action failed',
    }, () => app.api.argNodeAction(
        edgeData.source_obj || edgeData.source || edgeData.obj,
        actionName,
        {
          ...analysisControllerOptions({ includeBound: false, includeRelations: false }),
          target: edgeData.target_obj || edgeData.target,
          sheet_id: targetSheetId,
        },
      ));
    if (!result) return null;
    if (actionName === 'decompose' && result && result.decomposed) {
      const label = `Step: ${edgeData.label || actionName}`;
      app.openARGSheet(label, result.sub_arg, result.sheet_id);
    }
    if (result && result.arg) {
      applyArgSnapshot(app, targetSheetId, result.arg);
    }
    if (actionName === 'view_source') {
      highlightReturnedSourceInEditor(app, result);
    }
    app.controls.setStatus(`Done: ${actionName}`, 'success');
    return result;
  } catch (err) {
    await runWithContext(app, {
      busyMessage: `Executing: ${actionName}...`,
      failurePrefix: 'Edge action failed',
    }, () => {
      throw err;
    });
    console.error('ARG edge action error:', err);
    return null;
  }
}
