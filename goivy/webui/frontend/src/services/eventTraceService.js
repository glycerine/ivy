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
  doc = globalThis.document,
} = {}) {
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
  doc = globalThis.document,
} = {}) {
  const sheetState = app.sheets && app.sheets[sheetId];
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
  doc = globalThis.document,
} = {}) {
  const sheet = doc.getElementById(sheetId);
  const select = sheet ? sheet.querySelector('.event-pattern-list') : null;
  return select && select.value ? select.value : '';
}

export function applyEventPatternResult(app, sheetId, result, fallbackPatterns, {
} = {}) {
  const sheet = app.sheets && app.sheets[sheetId];
  if (!sheet) return;
  if (result && Array.isArray(result.patterns)) {
    sheet.patterns = result.patterns.slice();
  } else if (fallbackPatterns) {
    sheet.patterns = fallbackPatterns.slice();
  }
  app.renderEventPatternList(sheetId);
}

export function renderEventPatternList(app, sheetId, {
  doc = globalThis.document,
} = {}) {
  const sheetState = app.sheets && app.sheets[sheetId];
  const sheet = doc.getElementById(sheetId);
  const select = sheet ? sheet.querySelector('.event-pattern-list') : null;
  if (!select || !sheetState) return;
  select.innerHTML = '';
  for (const pattern of sheetState.patterns || []) {
    const option = doc.createElement('option');
    option.value = pattern;
    option.textContent = pattern;
    select.appendChild(option);
  }
}

export async function addEventPattern(app, sheetId, pattern) {
  const sheet = app.sheets && app.sheets[sheetId];
  if (!sheet) return undefined;
  if (app.api && app.api.executeAction && !sheet.visualOnly) {
    const result = await app.api.executeAction('events_add_pattern', { sheet_id: sheetId, pattern });
    app.applyEventPatternResult(sheetId, result);
    return result;
  }
  sheet.patterns = (sheet.patterns || []).concat([pattern]);
  app.renderEventPatternList(sheetId);
  return sheet.patterns;
}

export async function removeSelectedEventPattern(app, sheetId, {
  doc = globalThis.document,
} = {}) {
  const sheet = app.sheets && app.sheets[sheetId];
  const sheetEl = doc.getElementById(sheetId);
  const select = sheetEl ? sheetEl.querySelector('.event-pattern-list') : null;
  const idx = select ? select.selectedIndex : -1;
  if (!sheet || idx < 0) return undefined;
  if (app.api && app.api.executeAction && !sheet.visualOnly) {
    const result = await app.api.executeAction('events_remove_pattern', { sheet_id: sheetId, index: idx });
    app.applyEventPatternResult(sheetId, result);
    return result;
  }
  sheet.patterns.splice(idx, 1);
  app.renderEventPatternList(sheetId);
  return sheet.patterns;
}

export async function clearEventPatterns(app, sheetId) {
  const sheet = app.sheets && app.sheets[sheetId];
  if (!sheet) return undefined;
  if (app.api && app.api.executeAction && !sheet.visualOnly) {
    const result = await app.api.executeAction('events_clear_patterns', { sheet_id: sheetId });
    app.applyEventPatternResult(sheetId, result, []);
    return result;
  }
  sheet.patterns = [];
  app.renderEventPatternList(sheetId);
  return sheet.patterns;
}

export async function loadEventPatterns(app, sheetId, text) {
  const sheet = app.sheets && app.sheets[sheetId];
  if (!sheet) return undefined;
  if (app.api && app.api.executeAction && !sheet.visualOnly) {
    const result = await app.api.executeAction('events_load_patterns', { sheet_id: sheetId, patterns: text });
    app.applyEventPatternResult(sheetId, result);
    return result;
  }
  const patterns = String(text || '').split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
  sheet.patterns = sheet.patterns.concat(patterns);
  app.renderEventPatternList(sheetId);
  return sheet.patterns;
}

export async function saveEventPatterns(app, sheetId) {
  const sheet = app.sheets && app.sheets[sheetId];
  if (!sheet) return '';
  let content = (sheet.patterns || []).join('\n');
  if (content !== '') content += '\n';
  if (app.api && app.api.executeAction && !sheet.visualOnly) {
    const result = await app.api.executeAction('events_save_patterns', { sheet_id: sheetId });
    if (result && typeof result.content === 'string') {
      content = result.content;
    }
  }
  app.downloadTextFile('event_patterns.pats', content, 'text/plain');
  return content;
}
