Code Review: TypeScript Data Model vs Python/Go Data Model

  Overall Assessment

  The TypeScript port is structurally sound and covers all three Python layers. The pick() multi-key
  pattern correctly handles Go's PascalCase default serialization, camelCase JS convention, and
  snake_case wire format in one call. The RawBackedModel.raw retention is excellent for forward
  compatibility. The Option class correctly handles Go's flat bool (from Toggles.Edges/Labels) vs.
  {val: bool} objects.

  However there are several concrete bugs and gaps.

  ---
  Bug-Level Issues

  1. CyElements.nodeId / edgeId will always be {}

  art_cyrender.go:27-28:
  NodeID   map[string]string `json:"-"`  // excluded from JSON
  EdgeID   map[string]string `json:"-"`  // excluded from JSON
  Both are tagged json:"-". The TypeScript model reads pick(raw, 'node_id', 'nodeId', 'NodeID') and
  will always get undefined → empty {}. If the browser needs object-to-Cytoscape-ID lookup (e.g., to
  highlight a selected ARG node by state object key), this is broken. Fix: either change Go's tags to
  json:"node_id" / json:"edge_id", or remove these fields from the TS model and derive IDs from
  data.id inside each element's data map.

  2. GraphStack computed fields missing from Go wire format

  Go GraphStack (webui_graph_model.go:732):
  type GraphStack struct {
      Current   *Graph
      UndoStack []*Graph
      RedoStack []*Graph
  }
  No JSON tags, no can_undo/undo_depth/redo_depth fields. TypeScript reads pick(raw, 'can_undo',
  'canUndo') etc. — all will be false/0. The test in uiDataModel.test.ts expects can_undo: true and
  undo_depth: 2, confirming this is a required gap to fill. The Go side needs a snapshot struct:
  type GraphStackSnapshot struct {
      CanUndo   bool `json:"can_undo"`
      CanRedo   bool `json:"can_redo"`
      UndoDepth int  `json:"undo_depth"`
      RedoDepth int  `json:"redo_depth"`
  }

  3. ConceptSession.undoStack is unexported — depth never serialized

  webui_concept_session.go:12:
  undoStack  []*undoEntry  // lowercase = unexported
  TypeScript ConceptSession reads undo_depth and redo_depth which are absent from the wire. Undo state
   in the concept panel UI will be stuck at 0. The undoStack must either be exported with a JSON tag,
  or a serialization snapshot that computes len(undoStack) / len(redoStack) must be used.

  ---
  Structural Gaps

  4. ARGSnapshot.render reads CyElements from payload root — needs verification

  TypeScript:
  constructor(raw: unknown = {}) {
      this.render = new CyElements(raw);           // reads 'elements' from root
      this.analysisGraphState = new AnalysisGraphState(pick(raw, 'analysis_graph_state', ...) ?? {});
  This assumes the Go GetARG() response is a flat object like:
  { "elements": [...], "analysis_graph_state": {...}, "analysis_graph": {...} }
  If elements is under a "render" or "cy_elements" sub-key in the actual response, render.elements
  will be empty. The same assumption applies to ConceptSnapshot. This needs to be verified against the
   Go session's GetARG / GetConcept response structure and a TS test written to validate it
  end-to-end.

  5. ConceptSnapshot has redundant/duplicate fields

  TypeScript ConceptSnapshot declares:
  - domain, session, interactiveSession, displayCheckboxes, graphStack at top level
  - AND graph: ConceptGraphModel which re-contains all of these under ConceptGraphModel

  Depending on what Go actually sends, one set will always be empty defaults. This is dead weight and
  will cause confusion — callers won't know which to read from. Pick one canonical location and remove
   the other. If Go sends the Graph struct directly, use graph.*; if Go sends a flat concept snapshot,
   use the top-level fields.

  6. Python State.universe and State.expr missing from TypeScript State

  Python ivy_interp.State:
  - expr — action-application expression (how the state was reached; needed for CTI display)
  - universe — model universe set from BMC result (needed for concrete/CTI visualization)

  TypeScript State only has id, clauses, safe, label. The CTI flow (webui_ui_cti.go) needs universe to
   show counterexample witnesses. Add:
  readonly expr: unknown;      // how state was reached
  readonly universe: unknown;  // BMC model universe
  with pick(raw, 'expr', 'Expr') and pick(raw, 'universe', 'Universe').

  ---
  Minor Issues

  7. AnalysisGraph (Layer 1 full model) — verify it's actually sent

  ARGSnapshot.analysisGraph reads from analysis_graph. The full AnalysisGraph (Python Layer 1)
  contains formula expressions that are not easily serialized. Verify that the Go backend actually
  emits analysis_graph in the ARG response. If not, this field is always empty and can be removed to
  avoid confusion.

  8. ConceptInteractiveSession field types need verification against Go serialization

  webui_concept_isession.go:17:
  State              goivy.Expr
  Axioms             goivy.Expr
  GoalConstraints    []goivy.Expr
  SupposeConstraints []goivy.Expr
  These are goivy.Expr objects. TypeScript reads them as string and string[]. This works only if the
  Go session serializes them via .String() before sending. Confirm with a Go JSON marshal test. If
  they come through as complex expression objects, the TypeScript types are wrong.

  9. State.id fallback of -1 differs from Python

  Python State.id is set when added to the graph; unset states have no ID. TypeScript uses -1 as
  sentinel. This could mask bugs if a state ID is genuinely -1. Consider readonly id: number | null
  with default null.

  10. ConceptGraphModel reads ConceptSess/InteractiveSess/Checks from PascalCase

  Go Graph (webui_graph_model.go:212) has no JSON tags so fields serialize as PascalCase. TypeScript
  correctly tries 'ConceptSess', 'InteractiveSess', 'Checks' as fallback keys in pick(). This is
  correct but fragile — if the Go side ever adds explicit json:"..." tags the camelCase aliases will
  take priority. Consider adding explicit snake_case JSON tags on Go Graph fields to stabilize the
  wire format.

  ---
  What's Done Well

  - The pick(raw, 'snake_case', 'camelCase', 'PascalCase') pattern is the right approach for a
  polyglot codebase.
  - Option handling of bool vs {val: bool} correctly adapts Go's flat Toggles map.
  - Test coverage in uiDataModel.test.ts for ConceptSnapshot, ARGSnapshot, and DisplayCheckboxes is
  the right approach — more tests at the JSON boundary will catch wire format mismatches before the
  browser UI.
  - The inheritance ConceptInteractiveSession extends ConceptSession correctly mirrors the Python
  class hierarchy.
  - RawBackedModel.raw makes the model forward-compatible with new server fields.

  ---
  Priority Order for Fixes

  1. nodeId/edgeId always empty — affects node selection (Bug #1)
  2. GraphStack undo state always false/0 — affects undo button enable state (Bug #2)
  3. ConceptSession undo depth always 0 — affects undo button enable state (Bug #3)
  4. ConceptSnapshot redundancy — architectural clarity (Gap #5)
  5. State.universe/State.expr — required for CTI visualization (Gap #6)
  6. ARGSnapshot.render reads from root — needs structural test (Gap #4)
  7. AnalysisGraph verify or remove — cleanup (Minor #7)

