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
  try {
    const result = await app.api.getConceptGraph(app.selectedArgNode);
    if (result && result.elements) {
      app.conceptGraph.update(result.elements, result.positions);
    }
    if (result) {
      app.populateStateCheckboxes(result);
    }
  } catch (err) {
    console.error('Concept graph refresh error:', err);
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
