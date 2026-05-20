import { applyArgSnapshot, applyConceptSnapshot } from './uiDataRenderService.ts';
import { addTraceResultViewAction } from './checkService.ts';

export async function prepareArgNodeActionArgs(app, nodeData, actionName, args, sheetId) {
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

export async function executeArgNodeAction(app, nodeData, action, sheetId) {
  const targetSheetId = sheetId || app.activeSheetId || 'sheet-1';
  if (app.isVisualOnlySheet(targetSheetId)) {
    app.controls.setStatus(app.visualOnlyMessage('analysis'), 'warning');
    return null;
  }
  const actionName = action.action || action.id || action[0] || action.name;
  app.controls.setStatus(`Executing: ${actionName}...`);
  try {
    let args = { ...(action.args || {}), sheet_id: targetSheetId };
    args = await app.prepareArgNodeActionArgs(nodeData, actionName, args, targetSheetId);
    if (args === null) {
      app.controls.setStatus(`Action cancelled: ${actionName}`, 'warning');
      return null;
    }
    const result = await app.api.argNodeAction(nodeData.obj || nodeData.id, actionName, args);
    if (result && result.arg) {
      applyArgSnapshot(app, targetSheetId, result.arg);
    }
    if (result && result.concept) {
      applyConceptSnapshot(app, targetSheetId, result.concept);
    }
    if (actionName === 'check_safety') {
      showArgNodeSafetyResult(app, result);
      return result;
    }
    app.controls.setStatus(`Action complete: ${actionName}`, 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Action failed: ${err.message}`, 'error');
    console.error('ARG action error:', err);
    return null;
  }
}

export async function executeArgEdgeAction(app, edgeData, actionName, sheetId) {
  const targetSheetId = sheetId || app.activeSheetId || 'sheet-1';
  app.controls.setStatus(`Executing: ${actionName}...`);
  try {
    const result = await app.api.argNodeAction(
      edgeData.source_obj || edgeData.source || edgeData.obj,
      actionName,
      { target: edgeData.target_obj || edgeData.target, sheet_id: targetSheetId },
    );
    if (actionName === 'decompose' && result && result.decomposed) {
      const label = `Step: ${edgeData.label || actionName}`;
      app.openARGSheet(label, result.sub_arg, result.sheet_id);
    }
    if (result && result.arg) {
      applyArgSnapshot(app, targetSheetId, result.arg);
    }
    if (result && result.source && actionName === 'view_source') {
      app.setEditorContent(result.source);
      if (result.lineno) {
        app.scrollEditorToLine(result.lineno);
      }
      app.controls.showInfo(
        `Source: ${result.file || ''}${result.lineno ? ` line ${result.lineno}` : ''}`,
        '',
      );
    }
    app.controls.setStatus(`Done: ${actionName}`, 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Edge action failed: ${err.message}`, 'error');
    console.error('ARG edge action error:', err);
    return null;
  }
}
