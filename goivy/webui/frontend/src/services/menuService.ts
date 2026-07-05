function dropdownContent(dropdown) {
  const content = dropdown && dropdown.querySelector('.dropdown-content');
  return content instanceof HTMLElement ? content : null;
}

function resetDropdownPosition(dropdown) {
  const content = dropdownContent(dropdown);
  if (!content) return;
  content.classList.remove('dropdown-floating');
  content.style.left = '';
  content.style.top = '';
  content.style.minWidth = '';
  content.style.maxHeight = '';
}

export function closeAllDropdowns({
  doc = globalThis.document,
} = {}) {
  const all = doc.querySelectorAll('.dropdown.open');
  for (const dropdown of all) {
    dropdown.classList.remove('open');
    resetDropdownPosition(dropdown);
  }
}

export function positionDropdownContent(dropdown, {
  win = globalThis.window,
} = {}) {
  if (!dropdown || !win) return;
  const trigger = dropdown.querySelector('.panel-menu');
  const content = dropdownContent(dropdown);
  if (!(trigger instanceof HTMLElement) || !content) return;

  const viewportWidth = win.innerWidth || document.documentElement.clientWidth || 0;
  const viewportHeight = win.innerHeight || document.documentElement.clientHeight || 0;
  if (!viewportWidth || !viewportHeight) return;

  const padding = 8;
  const triggerRect = trigger.getBoundingClientRect();
  content.classList.add('dropdown-floating');
  content.style.left = '0px';
  content.style.top = '0px';
  content.style.minWidth = `${Math.max(160, Math.ceil(triggerRect.width))}px`;
  content.style.maxHeight = `${Math.max(48, viewportHeight - padding * 2)}px`;

  const contentRect = content.getBoundingClientRect();
  const width = Math.max(contentRect.width, 160);
  const height = contentRect.height;

  let left = triggerRect.left;
  if (left + width > viewportWidth - padding) {
    left = viewportWidth - padding - width;
  }
  left = Math.max(padding, left);

  let top = triggerRect.bottom + 1;
  const topAbove = triggerRect.top - height - 1;
  if (top + height > viewportHeight - padding && topAbove >= padding) {
    top = topAbove;
  } else if (top + height > viewportHeight - padding) {
    top = Math.max(padding, viewportHeight - padding - height);
    content.style.maxHeight = `${Math.max(48, viewportHeight - top - padding)}px`;
  }

  content.style.left = `${Math.round(left)}px`;
  content.style.top = `${Math.round(top)}px`;
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
  if (!item || item.enabled === false) return Promise.resolve({ ok: false, error: 'disabled action' });
  if (item.action === 'save') {
    return app.saveAnalysisState();
  }
  if (item.action === 'save_abstraction') {
    return app.saveAbstraction();
  }
  if (item.action === 'save_conjectures') {
    return app.saveInvariant();
  }
  if (item.action === 'remove_tab') {
    if (app.activeSheetId) app.removeSheet(app.activeSheetId);
    return Promise.resolve({ ok: true });
  }
  if (item.action === 'exit') {
    return app.closeCurrentFile();
  }
  if (item.action === 'bmc_conjecture') {
    if (region === 'concept' && typeof app.ctiBoundedCheck === 'function') {
      return app.ctiBoundedCheck();
    }
    if (typeof app.boundedCheck === 'function') {
      return app.boundedCheck();
    }
  }
  if (item.dispatch === 'action') {
    return app.runAction(item.action, {}, {
      runningMessage: `Running: ${item.action}...`,
      successMessage: `Done: ${item.action}`,
    });
  }
  return app.runAction(item.action, {});
}
