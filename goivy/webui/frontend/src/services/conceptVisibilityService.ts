import { ConceptSnapshot } from '../models/uiDataModel.ts';
import {
  EDGE_DISPLAY_CLASSES,
  displayConceptName,
  selectConceptGraphView,
  selectSheet,
  selectStateCheckboxRows,
  selectStateToggles,
  toggleChecked as selectorToggleChecked,
} from '../models/uiDataSelectors.ts';

function snapshotFrom(app, conceptData = null) {
  if (conceptData instanceof ConceptSnapshot) return conceptData;
  if (conceptData) return new ConceptSnapshot(conceptData);
  const sheet = selectSheet(app && app.uiDataModel, app && app.activeSheetId);
  return sheet ? sheet.concept : null;
}

function activeSheet(app) {
  return selectSheet(app && app.uiDataModel, app && app.activeSheetId);
}

export function hydrateBackendToggleState(app, conceptData) {
  return snapshotFrom(app, conceptData);
}

export function toggleChecked(app, name, displayClass) {
  return selectorToggleChecked(snapshotFrom(app), name, displayClass);
}

export function stateRelationRows(app, conceptData) {
  const snapshot = snapshotFrom(app, conceptData);
  if (!snapshot) return [];
  const shimSheet: any = { concept: snapshot, conceptSelections: [] };
  return selectStateCheckboxRows(shimSheet);
}

function headerLabels() {
  return [
    ['all_to_all', '+', 'Show definite edges'],
    ['edge_unknown', '?', 'Show unknown edges'],
    ['none_to_none', '-', 'Show absent edges'],
    ['transitive', 'T', 'Transitive reduction'],
  ];
}

function headerRows(table) {
  return Array.prototype.slice.call(table.querySelectorAll('thead'));
}

function headerCellAt(doc, tr, index) {
  let cell = tr.children[index];
  if (!cell) {
    cell = doc.createElement('th');
    tr.appendChild(cell);
  }
  return cell;
}

function wireDisplayClassHeader(app, rows, cell, displayClass, label, title) {
  if (!cell.textContent || !cell.textContent.trim()) {
    cell.textContent = label;
  }
  cell.setAttribute('data-state-toggle-class', displayClass);
  cell.setAttribute('role', 'button');
  cell.setAttribute('tabindex', '0');
  cell.setAttribute('title', title);
  cell.setAttribute('aria-label', title);
  cell.onclick = () => {
    const shouldCheck = (rows || []).some((row) => !(row.checked && row.checked[displayClass]));
    onDisplayClassToggle(app, rows || [], displayClass, shouldCheck);
  };
  cell.onkeydown = (event) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      cell.click();
    }
  };
}

function ensureStateCheckboxHeader(app, tbody, rows, doc) {
  const table = tbody && tbody.closest ? tbody.closest('table') : null;
  if (!table || !doc) return;

  const theads = headerRows(table);
  const staticThead = theads.find((candidate) => !candidate.hasAttribute('data-state-checkbox-header'));
  const generatedTheads = theads.filter((candidate) => candidate.hasAttribute('data-state-checkbox-header'));
  if (staticThead) {
    generatedTheads.forEach((candidate) => candidate.remove());
  }

  let thead = staticThead || generatedTheads[0];
  if (!thead) {
    thead = doc.createElement('thead');
    thead.setAttribute('data-state-checkbox-header', 'true');
    table.insertBefore(thead, table.firstChild);
  }

  let tr = thead.querySelector('tr');
  if (!tr) {
    tr = doc.createElement('tr');
    thead.appendChild(tr);
  }

  headerLabels().forEach(([displayClass, label, title], index) => {
    wireDisplayClassHeader(app, rows, headerCellAt(doc, tr, index), displayClass, label, title);
  });

  const nameHeader = headerCellAt(doc, tr, headerLabels().length);
  if (!nameHeader.textContent || !nameHeader.textContent.trim()) {
    nameHeader.textContent = 'Relation';
  }
}

export function logStateRelationTableClear(reason, app, details: any = {}) {
  try {
    const sheetId = app && (app.activeSheetId || 'sheet-1');
    const sheet = app && app.uiDataModel && app.uiDataModel.sheets && app.uiDataModel.sheets[sheetId];
    const stack = new Error().stack || '';
    console.log('[ivyweb state-relations] clear table', {
      reason,
      activeSheetId: sheetId || '',
      rowCount: details.rowCount,
      hadRows: details.hadRows,
      modelStateInvalid: !!(app && app._modelStateInvalid),
      modelStateInvalidReason: (app && app._modelStateInvalidReason) || '',
      reachabilityOnly: !!(sheet && sheet.reachabilityOnly),
      visualOnly: !!(sheet && sheet.visualOnly),
      hasConceptSnapshot: !!(sheet && sheet.concept),
      selectedArgNode: sheet ? sheet.selectedArgNode : null,
      details,
      stack,
    });
  } catch (err) {
    console.log('[ivyweb state-relations] clear table', { reason, details, logError: String(err) });
  }
}

export function renderStateCheckboxes(app, rows, {
  doc = globalThis.document,
} = {}) {
  const tbody = doc && doc.getElementById('state-checkbox-body');
  if (!tbody) return;
  ensureStateCheckboxHeader(app, tbody, rows, doc);

  const existingByName = new Map();
  for (const tr of Array.prototype.slice.call(tbody.querySelectorAll('tr[data-state-toggle-row]'))) {
    existingByName.set(tr.getAttribute('data-state-toggle-row'), tr);
  }
  const desiredNames = new Set((rows || []).map((row) => row.name));
  for (const [name, tr] of existingByName.entries()) {
    if (!desiredNames.has(name)) tr.remove();
  }

  const emptyRow = tbody.querySelector('tr[data-state-toggle-empty]');
  if (rows && rows.length > 0 && emptyRow) emptyRow.remove();

  for (const row of rows || []) {
    let tr = existingByName.get(row.name);
    if (!tr) {
      tr = doc.createElement('tr');
      tr.setAttribute('data-state-toggle-row', row.name);
    }
    while (tr.firstChild) tr.removeChild(tr.firstChild);
    [
      ['all_to_all', 'Show definite edges'],
      ['edge_unknown', 'Show unknown edges'],
      ['none_to_none', 'Show absent edges'],
      ['transitive', 'Transitive reduction'],
    ].forEach(([displayClass, title]) => {
      const td = doc.createElement('td');
      const cb = doc.createElement('input');
      cb.type = 'checkbox';
      cb.name = row.name;
      cb.value = displayClass;
      cb.title = `${title} (${row.name})`;
      cb.checked = !!(row.checked && row.checked[displayClass]);
      cb.addEventListener('change', () => {
        app.onEdgeToggle(row.name, displayClass, cb.checked);
      });
      td.appendChild(cb);
      tr.appendChild(td);
    });

    const td = doc.createElement('td');
    td.className = 'name-col';
    const relationButton = doc.createElement('button');
    relationButton.type = 'button';
    relationButton.textContent = row.name;
    relationButton.setAttribute('data-state-toggle-relation', row.name);
    relationButton.addEventListener('click', () => {
      const shouldCheck = EDGE_DISPLAY_CLASSES.some((displayClass) => !(row.checked && row.checked[displayClass]));
      onRelationToggle(app, row.name, shouldCheck);
    });
    td.appendChild(relationButton);
    tr.appendChild(td);
    tbody.appendChild(tr);
  }

  if ((!rows || rows.length === 0) && snapshotFrom(app)) {
    const tr = emptyRow || doc.createElement('tr');
    tr.setAttribute('data-state-toggle-empty', 'true');
    while (tr.firstChild) tr.removeChild(tr.firstChild);
    const td = doc.createElement('td');
    td.colSpan = 5;
    td.style.color = '#666';
    td.style.fontStyle = 'italic';
    td.textContent = 'No relations loaded';
    tr.appendChild(td);
    tbody.appendChild(tr);
  }
}

export function populateStateCheckboxes(app, conceptData, {
  doc = globalThis.document,
} = {}) {
  if (conceptData && app && typeof app.applyConceptSnapshot === 'function') {
    return app.applyConceptSnapshot((conceptData && conceptData.sheet_id) || app.activeSheetId || 'sheet-1', conceptData || {});
  }
  renderStateCheckboxes(app, stateRelationRows(app, conceptData), { doc });
  if (app && typeof app.populateConstraintFacts === 'function') {
    app.populateConstraintFacts(conceptData);
  }
  return snapshotFrom(app, conceptData);
}

export async function onEdgeToggle(app, edgeName, displayClass, checked) {
  return applyToggleBatch(app, [{ edge: edgeName, display_class: displayClass, value: checked }]);
}

export function applyModelToggleUpdates(app, updates) {
  if (!app || !app.uiDataStore || typeof app.uiDataStore.setConceptToggles !== 'function') return null;
  return app.uiDataStore.setConceptToggles(app.activeSheetId || 'sheet-1', updates || []);
}

export async function applyToggleBatch(app, updates) {
  const changes = (updates || []).filter(Boolean);
  if (changes.length === 0) return;
  try {
    for (const update of changes) {
      await app.api.setToggles(update);
    }
    applyModelToggleUpdates(app, changes);
    await app.refreshConceptGraph();
  } catch (err) {
    console.error('Toggle error:', err);
    if (app && typeof app.refreshConceptGraph === 'function') {
      try {
        await app.refreshConceptGraph();
      } catch (refreshErr) {
        console.error('Toggle resync error:', refreshErr);
      }
    }
  }
}

export async function onRelationToggle(app, edgeName, checked) {
  return applyToggleBatch(app, EDGE_DISPLAY_CLASSES.map((displayClass) => ({
    edge: edgeName,
    display_class: displayClass,
    value: checked,
  })));
}

export async function onDisplayClassToggle(app, rows, displayClass, checked) {
  return applyToggleBatch(app, (rows || []).map((row) => ({
    edge: row.name,
    display_class: displayClass,
    value: checked,
  })));
}

export function findEdgeVisibility(app, obj, label) {
  const snapshot = snapshotFrom(app);
  if (!snapshot) return null;
  const names = [obj, label].filter(Boolean);
  for (const name of names) {
    for (const displayClass of EDGE_DISPLAY_CLASSES) {
      if (selectorToggleChecked(snapshot, name, displayClass)) {
        return Object.fromEntries(EDGE_DISPLAY_CLASSES.map((klass) => [klass, selectorToggleChecked(snapshot, name, klass)]));
      }
    }
  }
  return null;
}

export function applyEdgeVisibility(app, conceptGraph, view = null) {
  const graph = conceptGraph || (app && app.conceptGraph);
  if (!graph || !graph.cy) return;
  const conceptView = view || selectConceptGraphView(activeSheet(app));
  graph.cy.edges().forEach((edge) => {
    const id = edge.id ? edge.id() : edge.data('id');
    const aliases = [
      id,
      `edge:${id}`,
      `edge:${edge.data('obj') || ''}`,
      `edge:${edge.data('obj') || ''}|${edge.data('source_obj') || ''}|${edge.data('target_obj') || ''}`,
    ];
    const visible = aliases.some((alias) => conceptView.edgeVisibilityById[alias]);
    edge.style('display', visible ? 'element' : 'none');
  });
}

export function applyNodeLabels(app) {
  if (app && typeof app.renderUIDataChange === 'function') {
    app.renderUIDataChange({ sheetId: app.activeSheetId || 'sheet-1', changed: ['concept'] });
  }
}

export function updateStateLabel(nodeId, {
  doc = globalThis.document,
} = {}) {
  const label = doc && doc.getElementById('state-label');
  if (label) {
    label.textContent = `State: ${nodeId != null && nodeId !== '' ? nodeId : '\u2014'}`;
  }
}

export async function onEdgeToggleChange(app, edgeName, className, checked) {
  return onEdgeToggle(app, edgeName, className, checked);
}

export async function onLabelToggleChange(app, labelName, className, checked) {
  try {
    await app.api.setToggles({
      label: labelName,
      display_class: className,
      value: checked,
    });
    await app.refreshConceptGraph();
  } catch (err) {
    console.error('Toggle update error:', err);
  }
}

export function persistedToggles(app) {
  return selectStateToggles(activeSheet(app));
}

export { displayConceptName };
