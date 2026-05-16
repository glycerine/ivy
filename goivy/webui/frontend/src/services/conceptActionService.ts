import {
  selectMaterializableEdges,
  selectPrimaryConceptNodeSelection,
} from '../models/uiDataSelectors.ts';

export async function executeConceptNodeAction(app, nodeData, action) {
  const actionID = action.action || action.id || action[0] || action.name || '';
  const actionName = actionID.toLowerCase();
  const actionArgs = action.args || {};
  const concept = nodeData.obj || nodeData.id;

  if (actionName === 'select') return app.selectConceptNode(concept);
  if (actionName === 'remove') return app.removeConcept(concept);
  if (actionName === 'empty' || actionName === 'suppose empty') return app.supposeEmpty(concept);
  if (actionName === 'materialize') return app.materializeNode(concept);
  if (actionName === 'materialize edge' || actionName === 'materialize_from_selected') return app.materializeEdgeFromSelected(concept);
  if (actionName.indexOf('split by ') === 0) return app.splitConcept(concept, actionName.substring('split by '.length));
  if (actionName.indexOf('split:') === 0) return app.splitConcept(concept, actionName.substring('split:'.length));
  if (actionName.indexOf('add ') === 0) return app.addProjection(actionName.substring('add '.length), concept);
  if (actionName === 'add_projection') return app.addProjection(actionArgs.name, actionArgs.concept || actionArgs.name);

  app.controls.setStatus(`Executing: ${actionName}...`);
  try {
    const result = await app.api.executeAction(actionName, { concept });
    await app.refreshConceptGraph();
    app.controls.setStatus(`Done: ${actionName}`, 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Error: ${err.message}`, 'error');
    return null;
  }
}

export async function executeConceptEdgeAction(app, edgeData, action) {
  const actionName = (action[0] || action.name || action.id || '').toLowerCase();
  const conceptId = edgeData.obj || edgeData.id;
  if (actionName === 'empty' || actionName === 'empty_edge') return app.supposeEmpty(conceptId);
  if (actionName === 'remove') return app.removeConcept(conceptId);
  if (actionName === 'materialize' || actionName === 'materialize +') return app.materializeEdge(edgeData, true);
  if (actionName === 'materialize -' || actionName === 'materialize \u2013') return app.materializeEdge(edgeData, false);
  if (actionName === 'dematerialize') return app.materializeEdge(edgeData, false);

  app.controls.setStatus(`Executing: ${actionName}...`);
  try {
    const result = await app.api.executeAction(actionName, { concept: conceptId });
    await app.refreshConceptGraph();
    app.controls.setStatus(`Done: ${actionName}`, 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Error: ${err.message}`, 'error');
    return null;
  }
}

export async function splitConcept(app, concept, splitBy) {
  app.controls.setStatus(`Splitting ${concept} by ${splitBy}...`);
  try {
    const result = await app.api.splitConcept(concept, splitBy);
    await app.refreshConceptGraph();
    app.controls.setStatus('Split complete', 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Split failed: ${err.message}`, 'error');
    return null;
  }
}

export async function supposeEmpty(app, concept) {
  app.controls.setStatus(`Supposing ${concept} is empty...`);
  try {
    const result = await app.api.supposeEmpty(concept);
    await app.refreshConceptGraph();
    app.controls.setStatus('Suppose empty applied', 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Suppose empty failed: ${err.message}`, 'error');
    return null;
  }
}

export async function removeConcept(app, concept) {
  app.controls.setStatus('Removing concept...');
  try {
    const result = await app.api.removeConcept(concept);
    await app.refreshConceptGraph();
    app.controls.setStatus('Concept removed', 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Remove failed: ${err.message}`, 'error');
    return null;
  }
}

export async function materializeNode(app, concept) {
  app.controls.setStatus('Materializing node...');
  try {
    const result = await app.api.materializeNode(concept);
    await app.refreshConceptGraph();
    app.controls.setStatus('Node materialized', 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Materialize failed: ${err.message}`, 'error');
    return null;
  }
}

export async function materializeEdge(app, edgeOrConcept, positive) {
  const dir = positive ? '+' : '\u2013';
  let relation = edgeOrConcept;
  let source = '';
  let target = '';
  if (edgeOrConcept && typeof edgeOrConcept === 'object') {
    relation = edgeOrConcept.obj || edgeOrConcept.label || edgeOrConcept.id;
    source = edgeOrConcept.source_obj || edgeOrConcept.source || '';
    target = edgeOrConcept.target_obj || edgeOrConcept.target || '';
  }
  app.controls.setStatus(`Materializing edge (${dir})...`);
  try {
    const result = await app.api.materializeEdge(relation, source, target, positive);
    await app.refreshConceptGraph();
    app.controls.setStatus(`Edge materialized (${dir})`, 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Materialize failed: ${err.message}`, 'error');
    return null;
  }
}

export async function addProjection(app, name, concept) {
  app.controls.setStatus(`Adding projection ${name}...`);
  try {
    const result = await app.api.addProjection(name, concept);
    await app.refreshConceptGraph();
    app.controls.setStatus('Projection added', 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Add projection failed: ${err.message}`, 'error');
    return null;
  }
}

export async function materializeEdgeFromSelected(app, targetConceptId) {
  const sourceSelection = selectPrimaryConceptNodeSelection(app.uiDataModel, app.activeSheetId || 'sheet-1');
  const sourceConceptId = sourceSelection && (sourceSelection.obj || sourceSelection.id || sourceSelection.label);
  if (!sourceConceptId) {
    app.controls.setStatus('Select a source node first', 'warning');
    return undefined;
  }
  let relations = selectMaterializableEdges(app.uiDataModel, app.activeSheetId || 'sheet-1', sourceConceptId, targetConceptId);
  if (relations.length === 0) {
    app.controls.setStatus('No matching binary relations', 'warning');
    return undefined;
  }
  const selected = await app.listboxDialog(
    'Materialize edge',
    'Materialize this relation from selected node:',
    relations.map((rel) => ({ label: rel, value: rel })),
    { cancel: true },
  );
  if (selected == null) {
    app.controls.setStatus('Materialize edge cancelled', 'warning');
    return undefined;
  }
  app.controls.setStatus(`Materializing edge ${selected}...`);
  try {
    const result = await app.api.materializeEdge(selected, sourceConceptId, targetConceptId, true);
    await app.refreshConceptGraph();
    app.controls.setStatus('Edge materialized (+)', 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Materialize edge failed: ${err.message}`, 'error');
    return null;
  }
}

export async function addRelationFromString(app) {
  const input = await app.entryDialog(
    'Add relation',
    'Add a relation [example: p(X,a,Y)]:',
    '',
    { okLabel: 'Add' },
  );
  if (!input) return undefined;
  try {
    const result = await app.api.executeAction('add_relation', { formula: input });
    await app.refreshConceptGraph();
    app.controls.setStatus('Relation added', 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Add relation failed: ${err.message}`, 'error');
    return null;
  }
}

export function selectConceptNode(app, conceptId) {
  if (!app.uiDataStore) throw new Error('UIDataModel store is required to select concept nodes');
  const selected = app.uiDataStore.toggleConceptNodeSelection(app.activeSheetId || 'sheet-1', conceptId);
  if (selected) {
    app.controls.setStatus(`Selected: ${conceptId}`);
  } else {
    app.controls.setStatus(`Deselected: ${conceptId}`);
  }
}

export async function splatterNode(app, conceptId) {
  app.controls.setStatus(`Splattering ${conceptId}...`);
  try {
    const result = await app.api.executeAction('splatter', { concept: conceptId });
    await app.refreshConceptGraph();
    app.controls.setStatus(`Splattered: ${conceptId}`, 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Splatter failed: ${err.message}`, 'error');
    return null;
  }
}
