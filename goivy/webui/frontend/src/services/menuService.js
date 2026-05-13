export function closeAllDropdowns({
  doc = globalThis.document,
} = {}) {
  const all = doc.querySelectorAll('.dropdown.open');
  for (const dropdown of all) dropdown.classList.remove('open');
}

export function flashAndClose(app, el, callback, {
  win = globalThis.window,
} = {}) {
  const setTimeoutFn = win && typeof win.setTimeout === 'function' ? win.setTimeout.bind(win) : setTimeout;
  el.classList.add('menu-flash');
  setTimeoutFn(() => {
    el.classList.remove('menu-flash');
    app.closeAllDropdowns();
    if (callback) callback();
  }, 50);
}

export function dispatchMenuDescriptorAction(app, region, item) {
  void region;
  if (!item || item.enabled === false) return Promise.resolve({ ok: false, error: 'disabled action' });
  if (item.dispatch === 'action') {
    return app.runAction(item.action, {}, {
      runningMessage: `Running: ${item.action}...`,
      successMessage: `Done: ${item.action}`,
    });
  }
  return app.runAction(item.action, {});
}
