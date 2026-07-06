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
  if (typeof item.action === 'string' && item.action.startsWith('mode_')) {
    const mode = item.action.substring('mode_'.length);
    if (typeof app.setMode === 'function') {
      return Promise.resolve(app.setMode(mode));
    }
  }
  if (item.action === 'save_model') {
    return app.save();
  }
  if (item.action === 'save_analysis_state') {
    return app.saveAnalysisState();
  }
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
  if (item.action === 'check_inductiveness' && typeof app.checkInduction === 'function') {
    return app.checkInduction();
  }
  if (item.action === 'bmc_conjecture') {
    if (region === 'concept' && typeof app.ctiBoundedCheck === 'function') {
      return app.ctiBoundedCheck();
    }
    if (typeof app.boundedCheck === 'function') {
      return app.boundedCheck();
    }
  }
  if (item.action === 'diagram' && typeof app.diagramCurrentState === 'function') {
    return app.diagramCurrentState();
  }
  if (item.action === 'weaken' && typeof app.weakenInvariant === 'function') {
    return app.weakenInvariant();
  }
  if (item.action === 'recalculate_all' && typeof app.recalculateAll === 'function') {
    return app.recalculateAll();
  }
  if ((item.action === 'show_reachable' || item.action === 'show_reachable_states') && typeof app.showReachableStates === 'function') {
    return app.showReachableStates();
  }
  const ctiConceptActions = {
    gather_facts: 'cti_gather',
    cti_gather: 'cti_gather',
    minimize_conjecture: 'cti_minimize',
    cti_minimize: 'cti_minimize',
    is_sufficient: 'cti_check_sufficient',
    cti_check_sufficient: 'cti_check_sufficient',
    is_inductive: 'cti_check_inductive',
    cti_check_inductive: 'cti_check_inductive',
    strengthen: 'cti_strengthen',
    cti_strengthen: 'cti_strengthen',
  };
  if (region === 'concept' && ctiConceptActions[item.action] && typeof app.ctiConceptAction === 'function') {
    return app.ctiConceptAction(ctiConceptActions[item.action]);
  }
  if (item.dispatch === 'action') {
    return app.runAction(item.action, {}, {
      runningMessage: `Running: ${item.action}...`,
      successMessage: `Done: ${item.action}`,
    });
  }
  return app.runAction(item.action, {});
}
