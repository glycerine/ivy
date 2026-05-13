import { describe, expect, it } from 'vitest';
import { ConceptSnapshot, UIDataModel } from './uiDataModel.ts';

describe('UIDataModel', () => {
  it('preserves typed ARG payloads separately from Cytoscape render elements', () => {
    const model = new UIDataModel();
    const snapshot = model.acceptArgSnapshot('sheet-1', {
      analysis_graph_state: {
        states: [{ id: 0, label: '0', is_bottom: false, info: 'initial' }],
        transitions: [{ source_id: 0, target_id: 1, label: 'connect', is_join: true }],
        covering: [{ covered_id: 2, covering_id: 1 }],
      },
      elements: [{ group: 'nodes', data: { id: 'n0', obj: 'state_0' } }],
    });

    expect(snapshot.analysisGraphState.states[0].id).toBe(0);
    expect(snapshot.analysisGraphState.transitions[0].isJoin).toBe(true);
    expect(snapshot.analysisGraphState.covering[0].coveringId).toBe(1);
    expect(snapshot.render.elements[0].data.obj).toBe('state_0');
  });

  it('does not infer ARG state from Cytoscape elements when typed fields are absent', () => {
    const model = new UIDataModel();
    const snapshot = model.acceptArgSnapshot('sheet-1', {
      elements: [{ group: 'nodes', data: { id: 'n0', obj: 'state_0' } }],
    });

    expect(snapshot.analysisGraphState.states).toEqual([]);
    expect(snapshot.render.elements).toHaveLength(1);
  });

  it('keeps concept domain, sessions, checkboxes, and graph stack as typed data', () => {
    const snapshot = new ConceptSnapshot({
      concept_domain: {
        concepts: {
          client: { name: 'client', variables: ['X'], formula: 'X = X', sorts: ['client'], arity: 1 },
          link: { name: 'link', variables: ['X', 'Y'], formula: 'link(X,Y)', sorts: ['client', 'server'], arity: 2 },
        },
        nodes: ['client'],
        edges: ['link'],
        node_labels: ['client'],
      },
      concept_session: {
        domain: {
          concepts: {
            client: { name: 'client', variables: ['X'], formula: 'X = X', sorts: ['client'], arity: 1 },
          },
          nodes: ['client'],
        },
        abstract_value: { 'node_label|node_necessarily|client|client': true },
      },
      display_checkboxes: {
        edges: { link: { all_to_all: true } },
        labels: { client: { node_necessarily: true } },
      },
      graph_stack: { can_undo: true, undo_depth: 2 },
    });

    expect(snapshot.domain.concepts.link.arity).toBe(2);
    expect(snapshot.session.abstractValue['node_label|node_necessarily|client|client']).toBe(true);
    expect(snapshot.displayCheckboxes.edgeVisible('link', 'all_to_all')).toBe(true);
    expect(snapshot.graphStack.canUndo).toBe(true);
    expect(snapshot.graphStack.undoDepth).toBe(2);
  });
});
