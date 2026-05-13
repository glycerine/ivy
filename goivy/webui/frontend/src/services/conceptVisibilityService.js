const EDGE_DEFAULTS = {
  all_to_all: false,
  edge_unknown: false,
  none_to_none: false,
  transitive: false,
};

const LABEL_DEFAULTS = {
  node_necessarily: false,
  node_maybe: false,
  node_necessarily_not: false,
};

const EDGE_TO_LABEL_CLASS = {
  all_to_all: 'node_necessarily',
  edge_unknown: 'node_maybe',
  none_to_none: 'node_necessarily_not',
};

export function hydrateBackendToggleState(app, conceptData) {
  app._edgeVisibility = {};
  app._labelVisibility = {};
  const toggles = (conceptData && conceptData.toggles) || {};
  const edges = toggles.edges || {};
  const labels = toggles.labels || {};
  for (const edge of Object.keys(edges)) {
    app._edgeVisibility[edge] = { ...EDGE_DEFAULTS, ...edges[edge] };
  }
  for (const label of Object.keys(labels)) {
    app._labelVisibility[label] = { ...LABEL_DEFAULTS, ...labels[label] };
  }
}

export function toggleChecked(app, name, displayClass) {
  const base = name.split('(')[0];
  const vis = app._edgeVisibility[name] || app._edgeVisibility[base];
  if (vis && Object.prototype.hasOwnProperty.call(vis, displayClass)) {
    return !!vis[displayClass];
  }
  const labelKey = EDGE_TO_LABEL_CLASS[displayClass];
  const labelVis = app._labelVisibility[name] || app._labelVisibility[base];
  if (labelKey && labelVis && Object.prototype.hasOwnProperty.call(labelVis, labelKey)) {
    return !!labelVis[labelKey];
  }
  return false;
}

export function stateRelationRows(app, conceptData) {
  const names = conceptData && conceptData.relations ? conceptData.relations.slice().sort() : [];
  return names.map((name) => ({
    name,
    checked: {
      all_to_all: toggleChecked(app, name, 'all_to_all'),
      edge_unknown: toggleChecked(app, name, 'edge_unknown'),
      none_to_none: toggleChecked(app, name, 'none_to_none'),
      transitive: toggleChecked(app, name, 'transitive'),
    },
  }));
}

export function populateStateCheckboxes(app, conceptData, {
  doc = globalThis.document,
} = {}) {
  app._lastConceptData = conceptData;
  const tbody = doc && doc.getElementById('state-checkbox-body');
  hydrateBackendToggleState(app, conceptData);
  const rows = stateRelationRows(app, conceptData);
  const names = rows.map((row) => row.name);

  if (!tbody) return;
  tbody.innerHTML = '';
  for (const name of names) {
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
      cb.name = name;
      cb.value = displayClass;
      cb.title = `${title} (${name})`;
      cb.checked = toggleChecked(app, name, displayClass);
      cb.addEventListener('change', () => {
        app.onEdgeToggle(name, displayClass, cb.checked);
      });
      td.appendChild(cb);
      tr.appendChild(td);
    });

    const td = doc.createElement('td');
    td.className = 'name-col';
    const a = doc.createElement('a');
    a.textContent = name;
    a.href = '#';
    a.addEventListener('click', (event) => {
      event.preventDefault();
    });
    td.appendChild(a);
    tr.appendChild(td);
    tbody.appendChild(tr);
  }

  if (names.length === 0 && conceptData) {
    const tr = doc.createElement('tr');
    const td = doc.createElement('td');
    td.colSpan = 5;
    td.style.color = '#666';
    td.style.fontStyle = 'italic';
    td.textContent = 'No relations loaded';
    tr.appendChild(td);
    tbody.appendChild(tr);
  }

  app._applyEdgeVisibility();
  app._applyNodeLabels();
  app.populateConstraintFacts(conceptData);
}

export async function onEdgeToggle(app, edgeName, displayClass, checked) {
  if (!app._edgeVisibility[edgeName]) {
    app._edgeVisibility[edgeName] = { ...EDGE_DEFAULTS };
  }
  app._edgeVisibility[edgeName][displayClass] = checked;

  const labelKey = EDGE_TO_LABEL_CLASS[displayClass];
  if (labelKey) {
    const bareName = edgeName.split('(')[0];
    if (!app._labelVisibility[bareName]) {
      app._labelVisibility[bareName] = { ...LABEL_DEFAULTS };
    }
    app._labelVisibility[bareName][labelKey] = checked;
  }

  app._applyEdgeVisibility();
  app._applyNodeLabels();

  try {
    await app.api.setToggles({
      edge: edgeName,
      display_class: displayClass,
      value: checked,
    });
    await app.refreshConceptGraph();
  } catch (err) {
    console.error('Toggle error:', err);
  }
}

export function findEdgeVisibility(app, obj, label) {
  const ev = app._edgeVisibility;
  if (ev[obj]) return ev[obj];
  if (ev[label]) return ev[label];
  for (const key of Object.keys(ev)) {
    const base = key.split('(')[0];
    if (base === obj || base === label) {
      return ev[key];
    }
  }
  return null;
}

export function applyEdgeVisibility(app, conceptGraph) {
  const graph = conceptGraph || app.conceptGraph;
  if (!graph || !graph.cy) return;
  graph.cy.edges().forEach((edge) => {
    const obj = edge.data('obj') || '';
    const label = edge.data('label') || '';
    const vis = findEdgeVisibility(app, obj, label);
    if (!vis) {
      edge.style('display', 'none');
      return;
    }
    const classList = edge.classes();
    const show = classList.some((className) => vis[className]);
    edge.style('display', show ? 'element' : 'none');
  });
}

export function displayConceptName(name) {
  if (typeof name === 'string' && name.charAt(0) === '=') {
    const body = name.slice(1);
    const idx = body.lastIndexOf(':');
    if (idx > 0) {
      return `=${body.slice(0, idx)}`;
    }
  }
  return name;
}

export function applyNodeLabels(app) {
  if (!app.conceptGraph || !app.conceptGraph.cy) return;
  if (!app._lastConceptData) return;

  const labelPrefixes = {
    node_necessarily: '',
    node_maybe: '?',
    node_necessarily_not: '\u00AC',
  };
  const nodeLabels = app._lastConceptData.node_labels || [];
  const labelSorts = app._lastConceptData.label_sorts || {};
  const abstractValue = app._lastConceptData.abstract_value || {};

  app.conceptGraph.cy.nodes().forEach((node) => {
    const nodeID = node.data('obj') || '';
    if (!nodeID) return;
    const sortName = node.data('cluster') || node.data('sort') || nodeID;
    const topLabel = node.data('display_label') || sortName;
    const labelParts = [topLabel];

    for (const labelName of nodeLabels) {
      const baseLabelName = labelName.split('(')[0];
      const labelSort = labelSorts[labelName] || labelSorts[baseLabelName];
      if (labelSort && labelSort !== sortName) continue;

      const necKey = `node_label|node_necessarily|${nodeID}|${baseLabelName}`;
      const necNotKey = `node_label|node_necessarily_not|${nodeID}|${baseLabelName}`;
      let k = 'node_maybe';
      if (abstractValue[necKey]) {
        k = 'node_necessarily';
      } else if (abstractValue[necNotKey]) {
        k = 'node_necessarily_not';
      }

      const vis = app._labelVisibility[labelName] || app._labelVisibility[baseLabelName];
      if (!vis || !vis[k]) continue;
      labelParts.push(`${labelPrefixes[k]}${displayConceptName(baseLabelName)}`);
    }

    const newLabel = labelParts.join('\n');
    if (node.data('label') !== newLabel) {
      node.data('label', newLabel);
      node.data('height', Math.max(50, 30 + labelParts.length * 20));
    }
  });
}

export function updateStateLabel(nodeId, {
  doc = globalThis.document,
} = {}) {
  const label = doc && doc.getElementById('state-label');
  if (label) {
    label.textContent = `State: ${nodeId != null ? nodeId : '\u2014'}`;
  }
}

export async function onEdgeToggleChange(app, edgeName, className, checked) {
  try {
    await app.api.setToggles({
      edge: edgeName,
      display_class: className,
      value: checked,
    });
    await app.refreshConceptGraph();
  } catch (err) {
    console.error('Toggle update error:', err);
  }
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
