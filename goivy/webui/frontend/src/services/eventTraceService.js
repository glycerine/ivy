export function readFileText(file, win = globalThis.window) {
  if (file && typeof file.text === 'function') {
    return file.text();
  }
  return new Promise((resolve, reject) => {
    const reader = new win.FileReader();
    reader.onload = () => resolve(String(reader.result || ''));
    reader.onerror = () => reject(reader.error || new Error('failed to read file'));
    reader.readAsText(file);
  });
}

export async function loadEventTraceFile(app, file) {
  if (!file) return null;
  app.controls.setStatus('Loading event trace...');
  try {
    const text = await app.readFileText(file);
    const result = await app.api.executeAction('events_parse', { content: text, filename: file.name || '' });
    app.openEventTraceSheet(file.name || result.label || 'Event trace', result || {}, result && result.sheet_id);
    app.controls.setStatus(`Loaded event trace: ${file.name || 'trace'}`, 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Event trace load failed: ${err.message}`, 'error');
    throw err;
  }
}

export function lookupEventTrace(events, address) {
  if (!address && address !== '0') return null;
  const parts = String(address).split('/');
  let current = null;
  let list = events || [];
  for (const part of parts) {
    const idx = Number(part);
    if (!Number.isInteger(idx) || idx < 0 || idx >= list.length) return null;
    current = list[idx];
    list = current.subs || [];
  }
  return current;
}

export function toggleEventTraceNode(app, sheetId, address, {
  bridge = globalThis.window && globalThis.window.__ivyVueBridge,
  doc = globalThis.document,
} = {}) {
  if (bridge && typeof bridge.setEventTraceExpanded === 'function') {
    let expanded = false;
    if (typeof bridge.isEventTraceExpanded === 'function') {
      expanded = !!bridge.isEventTraceExpanded(sheetId, address);
    }
    bridge.setEventTraceExpanded(sheetId, address, !expanded);
    return;
  }
  const row = app.eventTraceRow(sheetId, address);
  const li = row ? row.closest('.event-tree-node') : null;
  const sheetState = app.sheets && app.sheets[sheetId];
  if (!li || !sheetState) return;
  const existing = li.querySelector(':scope > ul.event-tree-list');
  const toggle = row.querySelector('.event-toggle');
  if (existing) {
    existing.remove();
    if (toggle) toggle.textContent = '+';
    return;
  }
  const ev = app.lookupEventTrace(sheetState.events, address);
  if (!ev || !ev.subs || ev.subs.length === 0) return;
  const list = doc.createElement('ul');
  list.className = 'event-tree-list';
  for (let i = 0; i < ev.subs.length; i += 1) {
    list.appendChild(app.renderEventTreeNode(ev.subs[i], sheetId, `${address}/${i}`));
  }
  li.appendChild(list);
  if (toggle) toggle.textContent = '-';
}

export function uncoverEventTraceAddress(app, sheetId, address) {
  const parts = String(address || '').split('/');
  let prefix = '';
  for (let i = 0; i < parts.length - 1; i += 1) {
    prefix = prefix === '' ? parts[i] : `${prefix}/${parts[i]}`;
    if (!app.eventTraceRow(sheetId, prefix)) break;
    if (!app.eventTraceRow(sheetId, `${prefix}/${parts[i + 1]}`)) {
      app.toggleEventTraceNode(sheetId, prefix);
    }
  }
}

export function selectEventTraceRow(app, sheetId, address, {
  bridge = globalThis.window && globalThis.window.__ivyVueBridge,
  doc = globalThis.document,
} = {}) {
  const sheetState = app.sheets && app.sheets[sheetId];
  if (bridge && typeof bridge.selectEventTraceRow === 'function') {
    if (sheetState) sheetState.selectedEventAddress = address;
    bridge.selectEventTraceRow(sheetId, address);
    return;
  }
  app.uncoverEventTraceAddress(sheetId, address);
  const sheet = doc.getElementById(sheetId);
  if (!sheetState || !sheet) return;
  for (const row of sheet.querySelectorAll('.event-row.selected')) row.classList.remove('selected');
  const row = app.eventTraceRow(sheetId, address);
  if (row) {
    row.classList.add('selected');
    if (typeof row.scrollIntoView === 'function') {
      row.scrollIntoView({ block: 'nearest' });
    }
  }
  sheetState.selectedEventAddress = address;
}

export function activeEventSheet(app) {
  const sheet = app.sheets && app.sheets[app.activeSheetId];
  return sheet && sheet.type === 'events' ? sheet : null;
}

export async function filterEventTrace(app, pattern) {
  const sheet = app.activeEventSheet();
  if (!sheet) {
    app.controls.setStatus('No event sheet selected', 'error');
    return null;
  }
  if (sheet.visualOnly) {
    app.controls.setStatus(app.visualOnlyMessage('events'), 'warning');
    return null;
  }
  try {
    const result = await app.api.executeAction('events_filter', {
      sheet_id: sheet.id,
      pattern,
    });
    const label = (result && result.label) || 'Filtered events';
    app.openEventTraceSheet(label, result || {}, result && result.sheet_id);
    return result;
  } catch (err) {
    app.controls.setStatus(`Filter failed: ${err.message}`, 'error');
    return null;
  }
}

export async function findEventTrace(app, pattern, reverse) {
  const sheet = app.activeEventSheet();
  if (!sheet) {
    app.controls.setStatus('No event sheet selected', 'error');
    return null;
  }
  if (sheet.visualOnly) {
    app.controls.setStatus(app.visualOnlyMessage('events'), 'warning');
    return null;
  }
  let result;
  try {
    result = await app.api.executeAction('events_find', {
      sheet_id: sheet.id,
      pattern,
      reverse: !!reverse,
      anchor: sheet.selectedEventAddress || '',
    });
  } catch (err) {
    app.controls.setStatus(`Find failed: ${err.message}`, 'error');
    return null;
  }
  if (!result || !result.address) {
    app.controls.setStatus('Pattern not found', 'error');
    return result;
  }
  app.selectEventTraceRow(sheet.id, result.address);
  return result;
}

export function selectedEventPattern(app, sheetId, {
  bridge = globalThis.window && globalThis.window.__ivyVueBridge,
  doc = globalThis.document,
} = {}) {
  if (bridge && typeof bridge.getSelectedEventPattern === 'function') {
    return bridge.getSelectedEventPattern(sheetId) || '';
  }
  const sheet = doc.getElementById(sheetId);
  const select = sheet ? sheet.querySelector('.event-pattern-list') : null;
  return select && select.value ? select.value : '';
}
