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

function ensureStateCheckboxHeader(app, tbody, rows, doc) {
  const table = tbody && tbody.closest ? tbody.closest('table') : null;
  if (!table || !doc) return;
  let thead = table.querySelector('thead[data-state-checkbox-header]');
  if (!thead) {
    thead = doc.createElement('thead');
    thead.setAttribute('data-state-checkbox-header', 'true');
    table.insertBefore(thead, table.firstChild);
  }
  thead.innerHTML = '';
  const tr = doc.createElement('tr');
  for (const [displayClass, label, title] of headerLabels()) {
    const th = doc.createElement('th');
    const button = doc.createElement('button');
    button.type = 'button';
    button.textContent = label;
    button.title = title;
    button.setAttribute('data-state-toggle-class', displayClass);
    button.addEventListener('click', () => {
      const shouldCheck = (rows || []).some((row) => !(row.checked && row.checked[displayClass]));
      onDisplayClassToggle(app, rows || [], displayClass, shouldCheck);
    });
    th.appendChild(button);
    tr.appendChild(th);
  }
  const nameHeader = doc.createElement('th');
  nameHeader.textContent = 'Relation';
  tr.appendChild(nameHeader);
  thead.appendChild(tr);
}

export function renderStateCheckboxes(app, rows, {
  doc = globalThis.document,
} = {}) {
  const tbody = doc && doc.getElementById('state-checkbox-body');
  if (!tbody) return;
  ensureStateCheckboxHeader(app, tbody, rows, doc);
  tbody.innerHTML = '';
  for (const row of rows || []) {
    const tr = doc.createElement('tr');
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
    const tr = doc.createElement('tr');
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

export async function applyToggleBatch(app, updates) {
  const changes = (updates || []).filter(Boolean);
  if (changes.length === 0) return;
  try {
    for (const update of changes) {
      await app.api.setToggles(update);
    }
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
