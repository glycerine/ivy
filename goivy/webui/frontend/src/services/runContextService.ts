function errorText(err) {
  if (!err) return 'Unknown error';
  if (err.message) return String(err.message);
  return String(err);
}

export function runContextErrorMessage(prefix, err) {
  const label = prefix || 'Command failed';
  return `${label}: ${errorText(err)}`;
}

export function beginRunContext(app, {
  busyMessage = 'Running...',
} = {}) {
  if (app) app._runContextDepth = Math.max(0, Number(app._runContextDepth) || 0) + 1;
  const doc = globalThis.document;
  if (doc && doc.body) {
    doc.body.setAttribute('data-ivy-run-context', 'busy');
    doc.body.setAttribute('aria-busy', 'true');
  }
  if (app && app.controls && typeof app.controls.showLoading === 'function') {
    app.controls.showLoading(busyMessage);
  }
  return function finishRunContext() {
    endRunContext(app);
  };
}

export function endRunContext(app) {
  if (app) app._runContextDepth = Math.max(0, (Number(app._runContextDepth) || 0) - 1);
  if (app && app._runContextDepth > 0) return;
  const doc = globalThis.document;
  if (doc && doc.body) {
    doc.body.removeAttribute('data-ivy-run-context');
    doc.body.removeAttribute('aria-busy');
  }
  if (app && app.controls && typeof app.controls.hideLoading === 'function') {
    app.controls.hideLoading();
  }
}

export async function reportRunContextError(app, err, {
  prefix = 'Command failed',
  modal = true,
  title = 'ivyweb',
  dialogMessage = 'Ivy error',
} = {}) {
  const message = runContextErrorMessage(prefix, err);
  if (app && app.controls && typeof app.controls.setStatus === 'function') {
    app.controls.setStatus(message, 'error');
  }
  if (modal !== false) {
    if (app && typeof app.textDialog === 'function') {
      await app.textDialog(title, dialogMessage, message, {
        readOnly: true,
        okLabel: 'OK',
        cancel: false,
      });
    } else if (app && typeof app.okDialog === 'function') {
      await app.okDialog(title, message);
    } else if (app && app.controls && typeof app.controls.showInfo === 'function') {
      app.controls.showInfo(dialogMessage, message);
    }
  }
  return { ok: false, error: errorText(err), message };
}

export async function runWithContext(app, options, callback) {
  const finish = beginRunContext(app, {
    busyMessage: options && options.busyMessage,
  });
  let finished = false;
  const finishOnce = function () {
    if (finished) return;
    finished = true;
    finish();
  };
  try {
    const result = await callback();
    finishOnce();
    return result;
  } catch (err) {
    finishOnce();
    await reportRunContextError(app, err, {
      prefix: options && options.failurePrefix,
      modal: !(options && options.modal === false),
      title: options && options.title,
      dialogMessage: options && options.dialogMessage,
    });
    return null;
  } finally {
    finishOnce();
  }
}
