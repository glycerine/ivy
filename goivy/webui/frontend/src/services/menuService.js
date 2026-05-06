export function closeAllDropdowns({
  bridge = globalThis.window && globalThis.window.__ivyVueBridge,
  doc = globalThis.document,
} = {}) {
  if (bridge && typeof bridge.closeDropdownMenus === 'function') {
    bridge.closeDropdownMenus();
    return;
  }
  const all = doc.querySelectorAll('.dropdown.open');
  for (const dropdown of all) dropdown.classList.remove('open');
}

export function flashAndClose(app, el, callback, {
  bridge = globalThis.window && globalThis.window.__ivyVueBridge,
  win = globalThis.window,
} = {}) {
  const setTimeoutFn = win && typeof win.setTimeout === 'function' ? win.setTimeout.bind(win) : setTimeout;
  if (bridge && typeof bridge.flashMenuItem === 'function' && el && el.id) {
    bridge.flashMenuItem(el.id, 50);
    setTimeoutFn(() => {
      app.closeAllDropdowns();
      if (callback) callback();
    }, 50);
    return;
  }
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
