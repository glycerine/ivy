import { describe, expect, it } from 'vitest';
import {
  ARGSnapshot,
  ARGNode,
  CTISnapshot,
  ConceptInteractiveSession,
  ConceptSession,
  ConceptSnapshot,
  CyElement,
  CyElementData,
  CyElements,
  FactSelection,
  FullARGNode,
  FullAnalysisGraphState,
  GraphStack,
  State,
  UIDataModel,
  graphSelectionKey,
} from './uiDataModel.ts';
import { createUIDataModelStore } from './uiDataModelStore.ts';

describe('UIDataModel', () => {
  it('preserves typed ARG payloads separately from Cytoscape render elements', () => {
    const snapshot = new ARGSnapshot({
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
    const snapshot = new ARGSnapshot({
      elements: [{ group: 'nodes', data: { id: 'n0', obj: 'state_0' } }],
    });

    expect(snapshot.analysisGraphState.states).toEqual([]);
    expect(snapshot.render.elements).toHaveLength(1);
  });

  it('keeps reachability-only sheet layout as typed model state', () => {
    const model = new UIDataModel();
    const sheet = model.registerSheet('sheet-3', {
      type: 'analysis',
      reachabilityOnly: true,
    });

    expect(sheet.reachabilityOnly).toBe(true);
    expect(sheet.visualOnly).toBe(false);
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

describe('UIDataModel — wire format coverage', () => {
  // -------------------------------------------------------------------------
  // ARG payload tests
  // -------------------------------------------------------------------------

  it('ARGSnapshot: flat payload → all fields parsed', () => {
    const snapshot = new ARGSnapshot({
      analysis_graph_state: {
        states: [
          { id: 0, label: 'init', is_bottom: false, is_safe: true, info: 'initial state' },
          { id: 1, label: 'err', is_bottom: true, is_marked: true, info: 'bottom' },
        ],
        transitions: [{ source_id: 0, target_id: 1, label: 'step', is_join: false }],
        covering: [{ covered_id: 3, covering_id: 0 }],
      },
      elements: [
        { group: 'nodes', data: { id: 'n0', obj: 'state_0', label: 'init' } },
        { group: 'edges', data: { id: 'e0', source: 'n0', target: 'n1' } },
      ],
      positions: null,
    });

    expect(snapshot.analysisGraphState.states).toHaveLength(2);
    expect(snapshot.analysisGraphState.states[0].id).toBe(0);
    expect(snapshot.analysisGraphState.states[0].isSafe).toBe(true);
    expect(snapshot.analysisGraphState.states[1].isBottom).toBe(true);
    expect(snapshot.analysisGraphState.states[1].isMarked).toBe(true);
    expect(snapshot.analysisGraphState.transitions[0].sourceId).toBe(0);
    expect(snapshot.analysisGraphState.transitions[0].isJoin).toBe(false);
    expect(snapshot.analysisGraphState.covering[0].coveredId).toBe(3);
    expect(snapshot.analysisGraphState.covering[0].coveringId).toBe(0);
    expect(snapshot.render.elements).toHaveLength(2);
    expect(snapshot.render.elements[0].data.label).toBe('init');
  });

  it('ARGSnapshot: missing analysis_graph_state → empty typed arrays', () => {
    const snapshot = new ARGSnapshot({
      elements: [{ group: 'nodes', data: { id: 'n0' } }],
    });
    expect(snapshot.analysisGraphState.states).toEqual([]);
    expect(snapshot.analysisGraphState.transitions).toEqual([]);
    expect(snapshot.analysisGraphState.covering).toEqual([]);
    expect(snapshot.render.elements).toHaveLength(1);
  });

  it('ARGSnapshot has no analysisGraph property', () => {
    const snapshot = new ARGSnapshot({});
    expect('analysisGraph' in snapshot).toBe(false);
  });

  it('CyElement.position is null when position key absent', () => {
    const el = new CyElement({ group: 'nodes', data: { id: 'n0' } });
    expect(el.position).toBeNull();
  });

  it('CyElement.position parsed when present', () => {
    const el = new CyElement({ group: 'nodes', data: { id: 'n0' }, position: { x: 10, y: 20 } });
    expect(el.position).not.toBeNull();
    expect(el.position!.x).toBe(10);
    expect(el.position!.y).toBe(20);
  });

  it('CyElements parses graph-level and node-level positions', () => {
    const elements = new CyElements({
      positions: { n0: { x: 1, y: 2 } },
      elements: [
        { group: 'nodes', data: { id: 'n0' } },
        { group: 'nodes', data: { id: 'n1' }, position: { x: 3, y: 4 } },
      ],
    });

    expect(elements.positions).toEqual({
      n0: { x: 1, y: 2 },
      n1: { x: 3, y: 4 },
    });
  });

  it('CyElementData parses typed metadata and emits only the render contract', () => {
    const data = new CyElementData({
      id: 'e0',
      obj: 'link',
      source: 'n0',
      target: 'n1',
      source_obj: 'client',
      target_obj: 'server',
      short_info: 'link(client, server)',
      long_info: ['formula'],
      actions: [{ label: 'Remove', action: 'remove', args: { concept: 'link' } }],
      is_safe: true,
      is_marked: true,
      ignored_raw_field: 'nope',
    });

    expect(data.sourceObj).toBe('client');
    expect(data.actions[0].action).toBe('remove');
    expect(data.isSafe).toBe(true);
    expect(data.isMarked).toBe(true);
    expect(data.toCytoscapeData()).toEqual({
      id: 'e0',
      obj: 'link',
      source: 'n0',
      target: 'n1',
      source_obj: 'client',
      target_obj: 'server',
      short_info: 'link(client, server)',
      long_info: ['formula'],
      actions: [{ label: 'Remove', action: 'remove', args: { concept: 'link' } }],
      is_safe: true,
      is_marked: true,
    });
  });

  it('State.expr and State.universe preserved from raw', () => {
    const s = new State({ id: 0, label: 'q0', expr: 'assume(pre)', universe: { client: ['c0'] } });
    expect(s.expr).toBe('assume(pre)');
    expect(s.universe).toEqual({ client: ['c0'] });
  });

  it('State.expr and State.universe are undefined when absent', () => {
    const s = new State({ id: 0, label: 'q0' });
    expect(s.expr).toBeUndefined();
    expect(s.universe).toBeUndefined();
  });

  // -------------------------------------------------------------------------
  // CyElements structural test
  // -------------------------------------------------------------------------

  it('CyElements has no nodeId or edgeId fields', () => {
    const cy = new CyElements({ elements: [{ group: 'nodes', data: { id: 'n0' } }] });
    expect('nodeId' in cy).toBe(false);
    expect('edgeId' in cy).toBe(false);
    expect('raw' in cy).toBe(false);
    expect(cy.elements).toHaveLength(1);
  });

  // -------------------------------------------------------------------------
  // ConceptInteractiveSession tests
  // -------------------------------------------------------------------------

  it('CIS.abstractValue converts TagValue array → Record<string,boolean>', () => {
    const cis = new ConceptInteractiveSession({
      abstract_value: [
        { Tag: ['node_info', 'at_least_one', 'client'], Value: true },
        { Tag: ['node_info', 'none', 'client'], Value: false },
      ],
    });
    expect(cis.abstractValue['node_info|at_least_one|client']).toBe(true);
    expect(cis.abstractValue['node_info|none|client']).toBe(false);
  });

  it('CIS.abstractValue handles five-part edge_info tag', () => {
    const cis = new ConceptInteractiveSession({
      abstract_value: [
        { Tag: ['edge_info', 'all_to_all', 'link', 'client', 'server'], Value: true },
      ],
    });
    expect(cis.abstractValue['edge_info|all_to_all|link|client|server']).toBe(true);
  });

  it('CIS.abstractValue accepts lowercase tag/value keys', () => {
    const cis = new ConceptInteractiveSession({
      abstract_value: [{ tag: ['a', 'b'], value: true }],
    });
    expect(cis.abstractValue['a|b']).toBe(true);
  });

  it('CIS.abstractValue falls back to empty map when field absent', () => {
    const cis = new ConceptInteractiveSession({ state: 'foo' });
    expect(cis.abstractValue).toEqual({});
  });

  it('CIS.abstractValue still works when given a plain bool map (map format)', () => {
    const cis = new ConceptInteractiveSession({
      abstract_value: { 'node_info|at_least_one|client': true },
    });
    expect(cis.abstractValue['node_info|at_least_one|client']).toBe(true);
  });

  it('CIS.undoDepth and redoDepth parsed from wire', () => {
    const cis = new ConceptInteractiveSession({ undo_depth: 3, redo_depth: 1 });
    expect(cis.undoDepth).toBe(3);
    expect(cis.redoDepth).toBe(1);
  });

  it('CIS.goalConstraints and supposeConstraints are string arrays', () => {
    const cis = new ConceptInteractiveSession({
      goal_constraints: ['forall X. p(X)', 'q(a)'],
      suppose_constraints: ['~r(b)'],
    });
    expect(cis.goalConstraints).toEqual(['forall X. p(X)', 'q(a)']);
    expect(cis.supposeConstraints).toEqual(['~r(b)']);
  });

  it('CIS.axioms parses as string, CIS.cache parses as bool map', () => {
    const cis = new ConceptInteractiveSession({
      axioms: 'forall X,Y. link(X,Y) => connected(X)',
      cache: { 'node_info|at_least_one|client': true, 'node_info|none|server': false },
    });
    expect(cis.axioms).toBe('forall X,Y. link(X,Y) => connected(X)');
    expect(cis.cache['node_info|at_least_one|client']).toBe(true);
    expect(cis.cache['node_info|none|server']).toBe(false);
  });

  // -------------------------------------------------------------------------
  // ConceptSession structural tests
  // -------------------------------------------------------------------------

  it('ConceptSession base class has no undoDepth or redoDepth', () => {
    const cs = new ConceptSession({ domain: {}, abstract_value: {} });
    expect('undoDepth' in cs).toBe(false);
    expect('redoDepth' in cs).toBe(false);
  });

  // -------------------------------------------------------------------------
  // GraphStack tests
  // -------------------------------------------------------------------------

  it('GraphStack has only four scalar fields — no current/undoStack/redoStack', () => {
    const gs = new GraphStack({ can_undo: true, can_redo: false, undo_depth: 2, redo_depth: 0 });
    expect('current' in gs).toBe(false);
    expect('undoStack' in gs).toBe(false);
    expect('redoStack' in gs).toBe(false);
    expect(gs.canUndo).toBe(true);
    expect(gs.canRedo).toBe(false);
    expect(gs.undoDepth).toBe(2);
    expect(gs.redoDepth).toBe(0);
  });

  it('GraphStack zero-values when payload is empty', () => {
    const gs = new GraphStack({});
    expect(gs.canUndo).toBe(false);
    expect(gs.canRedo).toBe(false);
    expect(gs.undoDepth).toBe(0);
    expect(gs.redoDepth).toBe(0);
  });

  // -------------------------------------------------------------------------
  // ConceptSnapshot tests
  // -------------------------------------------------------------------------

  it('ConceptSnapshot reads domain from concept_domain key', () => {
    const snap = new ConceptSnapshot({
      concept_domain: {
        concepts: {
          node: { name: 'node', variables: ['X'], formula: 'X=X', sorts: ['node'], arity: 1 },
        },
        nodes: ['node'],
        edges: [],
        node_labels: [],
      },
    });
    expect(snap.domain.concepts.node.name).toBe('node');
    expect(snap.domain.nodes).toEqual(['node']);
  });

  it('ConceptSnapshot.displayCheckboxes reads from toggles key when display_checkboxes absent', () => {
    const snap = new ConceptSnapshot({
      toggles: {
        edges: { link: { all_to_all: true, none_to_none: false } },
        labels: {},
      },
    });
    expect(snap.displayCheckboxes.edgeVisible('link', 'all_to_all')).toBe(true);
    expect(snap.displayCheckboxes.edgeVisible('link', 'none_to_none')).toBe(false);
  });

  it('ConceptSnapshot.displayCheckboxes prefers display_checkboxes over toggles', () => {
    const snap = new ConceptSnapshot({
      display_checkboxes: {
        edges: { link: { all_to_all: true } },
        labels: {},
      },
      toggles: {
        edges: { link: { all_to_all: false } },
        labels: {},
      },
    });
    expect(snap.displayCheckboxes.edgeVisible('link', 'all_to_all')).toBe(true);
  });

  it('ConceptSnapshot.graphStack reads can_undo/can_redo/undo_depth/redo_depth', () => {
    const snap = new ConceptSnapshot({
      graph_stack: { can_undo: true, can_redo: true, undo_depth: 4, redo_depth: 1 },
    });
    expect(snap.graphStack.canUndo).toBe(true);
    expect(snap.graphStack.canRedo).toBe(true);
    expect(snap.graphStack.undoDepth).toBe(4);
    expect(snap.graphStack.redoDepth).toBe(1);
  });

  it('full concept payload round-trip', () => {
    const snap = new ConceptSnapshot({
      concept_domain: {
        concepts: {
          client: { name: 'client', variables: ['X'], formula: 'X=X', sorts: ['client'], arity: 1 },
          link: { name: 'link', variables: ['X', 'Y'], formula: 'link(X,Y)', sorts: ['client', 'server'], arity: 2 },
        },
        nodes: ['client'],
        edges: ['link'],
        node_labels: [],
      },
      concept_session: {
        domain: { concepts: {}, nodes: [], edges: [], node_labels: [] },
        abstract_value: { 'node_label|node_necessarily|client|=client': true },
      },
      concept_interactive_session: {
        abstract_value: [
          { Tag: ['edge_info', 'all_to_all', 'link', 'client', 'client'], Value: false },
          { Tag: ['node_info', 'at_least_one', 'client'], Value: true },
        ],
        goal_constraints: ['safety'],
        suppose_constraints: [],
        state: 'init_state_formula',
        undo_depth: 2,
        redo_depth: 0,
        info: 'checking',
      },
      display_checkboxes: {
        edges: { link: { all_to_all: true, none_to_none: false } },
        labels: { client: { node_necessarily: true } },
      },
      graph_stack: { can_undo: true, can_redo: false, undo_depth: 2, redo_depth: 0 },
      elements: [{ group: 'nodes', data: { id: 'n0', obj: 'client', label: 'client' } }],
      selected_node: 'n0',
      state_label: 'state 0',
    });

    // domain
    expect(snap.domain.concepts.link.arity).toBe(2);
    expect(snap.domain.nodes).toEqual(['client']);

    // session abstractValue (plain map format from concept_session)
    expect(snap.session.abstractValue['node_label|node_necessarily|client|=client']).toBe(true);

    // interactiveSession: TagValue array converted to map
    expect(snap.interactiveSession.abstractValue['node_info|at_least_one|client']).toBe(true);
    expect(snap.interactiveSession.abstractValue['edge_info|all_to_all|link|client|client']).toBe(false);
    expect(snap.interactiveSession.undoDepth).toBe(2);
    expect(snap.interactiveSession.goalConstraints).toEqual(['safety']);

    // display checkboxes
    expect(snap.displayCheckboxes.edgeVisible('link', 'all_to_all')).toBe(true);
    expect(snap.displayCheckboxes.nodeLabelVisible('client', 'node_necessarily')).toBe(true);

    // graph stack
    expect(snap.graphStack.canUndo).toBe(true);
    expect(snap.graphStack.undoDepth).toBe(2);

    // render elements
    expect(snap.render.elements).toHaveLength(1);
    expect(snap.render.elements[0].data.label).toBe('client');

    // selected_node and state_label
    expect(snap.selectedNode).toBe('n0');
    expect(snap.stateLabel).toBe('state 0');
  });

  // -------------------------------------------------------------------------
  // FactSelection tests
  // -------------------------------------------------------------------------

  it('FactSelection: three-field parse from wire', () => {
    const f = new FactSelection({ index: 2, text: 'forall X. p(X)', selected: true });
    expect(f.index).toBe(2);
    expect(f.text).toBe('forall X. p(X)');
    expect(f.selected).toBe(true);
  });

  it('FactSelection: zero-value defaults when fields absent', () => {
    const f = new FactSelection({});
    expect(f.index).toBe(0);
    expect(f.text).toBe('');
    expect(f.selected).toBe(false);
  });

  it('ConceptSnapshot.facts parsed as FactSelection[]', () => {
    const snap = new ConceptSnapshot({
      facts: [
        { index: 0, text: '~r(X)', selected: true },
        { index: 1, text: 'p(a)', selected: false },
      ],
    });
    expect(snap.facts).toHaveLength(2);
    expect(snap.facts[0]).toBeInstanceOf(FactSelection);
    expect(snap.facts[0].text).toBe('~r(X)');
    expect(snap.facts[0].selected).toBe(true);
    expect(snap.facts[1].index).toBe(1);
    expect(snap.facts[1].selected).toBe(false);
  });

  it('ConceptSnapshot.facts is empty array when facts absent', () => {
    const snap = new ConceptSnapshot({});
    expect(snap.facts).toEqual([]);
  });

  it('FullARGNode: clauses, action_name, universe parsed', () => {
    const n = new FullARGNode({
      id: 3,
      label: '3',
      is_bottom: false,
      is_safe: true,
      is_marked: true,
      info: 'some info',
      clauses: 'pre(x) & inv(y)',
      action_name: 'send',
      universe: { node: ['n0', 'n1'], data: ['d0'] },
    });
    expect(n.id).toBe(3);
    expect(n.isSafe).toBe(true);
    expect(n.isMarked).toBe(true);
    expect(n.clauses).toBe('pre(x) & inv(y)');
    expect(n.actionName).toBe('send');
    expect(n.universe).toEqual({ node: ['n0', 'n1'], data: ['d0'] });
  });

  it('ARGNode: safety and marked metadata parsed', () => {
    const n = new ARGNode({ id: 4, label: '4', is_safe: true, is_marked: true });
    expect(n.isSafe).toBe(true);
    expect(n.isMarked).toBe(true);
  });

  it('FullARGNode: universe is null when absent', () => {
    const n = new FullARGNode({ id: 0, label: '0', clauses: '', action_name: '' });
    expect(n.universe).toBeNull();
  });

  it('FullARGNode: universe Record<string,string[]> when present', () => {
    const n = new FullARGNode({ universe: { sort: ['a', 'b'] } });
    expect(n.universe).not.toBeNull();
    expect(n.universe!['sort']).toEqual(['a', 'b']);
  });

  it('FullAnalysisGraphState: reuses ARGTransition and ARGCover', () => {
    const fs = new FullAnalysisGraphState({
      states: [{ id: 0, label: '0', is_bottom: false, info: '', clauses: '', action_name: '' }],
      transitions: [{ source_id: 0, target_id: 1, label: 'step', is_join: false }],
      covering: [{ covered_id: 2, covering_id: 0 }],
    });
    expect(fs.states[0]).toBeInstanceOf(FullARGNode);
    expect(fs.transitions[0].sourceId).toBe(0);
    expect(fs.covering[0].coveredId).toBe(2);
  });

  it('ARGSnapshot: fullAnalysisGraphState populated alongside analysisGraphState', () => {
    const snap = new ARGSnapshot({
      elements: [],
      analysis_graph_state: {
        states: [{ id: 7, label: '7', is_bottom: true, info: 'x', clauses: 'false', action_name: 'act' }],
        transitions: [],
        covering: [],
      },
    });
    expect(snap.analysisGraphState.states[0].id).toBe(7);
    expect(snap.fullAnalysisGraphState.states[0]).toBeInstanceOf(FullARGNode);
    expect(snap.fullAnalysisGraphState.states[0].clauses).toBe('false');
    expect(snap.fullAnalysisGraphState.states[0].actionName).toBe('act');
  });

  it('CTISnapshot: haveCti and currentConjecture parsed', () => {
    const snap = new CTISnapshot({
      elements: [],
      have_cti: true,
      current_conjecture: 'forall X. inv(X)',
    });
    expect(snap.haveCti).toBe(true);
    expect(snap.currentConjecture).toBe('forall X. inv(X)');
  });

  it('CTISnapshot: extends ARGSnapshot — render elements present', () => {
    const snap = new CTISnapshot({
      elements: [{ group: 'nodes', data: { id: 'n0' } }],
      have_cti: false,
      current_conjecture: '',
      analysis_graph_state: {
        states: [{ id: 0, label: '0', is_bottom: false, info: '', clauses: '', action_name: '' }],
        transitions: [],
        covering: [],
      },
    });
    expect(snap.render.elements).toHaveLength(1);
    expect(snap.fullAnalysisGraphState.states[0]).toBeInstanceOf(FullARGNode);
    expect(snap.haveCti).toBe(false);
  });

  it('stores concept node and edge selections as typed graph tuples', () => {
    const model = new UIDataModel();
    const store = createUIDataModelStore(model);
    store.applyConceptSnapshot('sheet-1', { elements: [] });

    expect(store.toggleConceptNodeSelection('sheet-1', { id: 'n0', obj: 'client', label: 'client' })).toBe(true);
    expect(store.toggleConceptEdgeSelection('sheet-1', {
      id: 'e0',
      obj: 'link',
      label: 'link',
      source_obj: 'client',
      target_obj: 'server',
    })).toBe(true);

    const selections = model.sheets['sheet-1'].conceptSelections;
    expect(selections).toEqual([
      { kind: 'node', id: 'n0', obj: 'client', label: 'client', sourceObj: '', targetObj: '' },
      { kind: 'edge', id: 'e0', obj: 'link', label: 'link', sourceObj: 'client', targetObj: 'server' },
    ]);
    expect(graphSelectionKey(selections[1])).toBe('edge:e0');

    store.applyConceptSnapshot('sheet-1', { elements: [] });
    expect(model.sheets['sheet-1'].conceptSelections).toEqual([]);
  });

  it('keeps graph positions as model state across edge-only snapshot changes', () => {
    const model = new UIDataModel();
    const store = createUIDataModelStore(model);
    store.applyConceptSnapshot('sheet-1', {
      elements: [{ group: 'nodes', data: { id: 'client' } }],
      positions: { client: { x: 20, y: 30 } },
    });

    store.applyConceptSnapshot('sheet-1', {
      elements: [
        { group: 'nodes', data: { id: 'client' } },
        { group: 'edges', data: { id: 'self', source: 'client', target: 'client' } },
      ],
    });

    expect(model.sheets['sheet-1'].conceptPositions).toEqual({
      client: { x: 20, y: 30 },
    });
  });

  it('invalidates cached analysis model state without clearing event sheets', () => {
    const model = new UIDataModel();
    const store = createUIDataModelStore(model);
    store.applyArgSnapshot('sheet-1', {
      elements: [{ group: 'nodes', data: { id: 's0' } }],
      positions: { s0: { x: 1, y: 2 } },
    });
    store.applyConceptSnapshot('sheet-1', {
      elements: [{ group: 'nodes', data: { id: 'client' } }],
      positions: { client: { x: 3, y: 4 } },
    });
    store.applyCtiSnapshot('sheet-1', { elements: [{ group: 'nodes', data: { id: 'cti0' } }] });
    store.setSelectedArgNode('sheet-1', 's0');
    store.setConceptSelections('sheet-1', [{ kind: 'node', id: 'client', obj: 'client' }]);
    model.registerSheet('events-1', { id: 'events-1', type: 'events' }).argPositions = { kept: { x: 9, y: 9 } };

    store.invalidateModelState();

    expect(model.sheets['sheet-1'].arg).toBeNull();
    expect(model.sheets['sheet-1'].concept).toBeNull();
    expect(model.sheets['sheet-1'].cti).toBeNull();
    expect(model.sheets['sheet-1'].argPositions).toEqual({});
    expect(model.sheets['sheet-1'].conceptPositions).toEqual({});
    expect(model.sheets['sheet-1'].selectedArgNode).toBeNull();
    expect(model.sheets['sheet-1'].conceptSelections).toEqual([]);
    expect(model.sheets['events-1'].argPositions).toEqual({ kept: { x: 9, y: 9 } });
  });
});
