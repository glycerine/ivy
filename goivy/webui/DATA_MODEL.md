# Ivy Python GUI Data Model

Yes — there is a clear, coherent three-layer data model. It is not tangled with the
view code. Here it is:

Layer 1: Semantic/Computation Model (ivy_art.py, ivy_interp.py)

ivy_art.AnalysisGraph (ivy_art.py:72) — the ARG itself:
states       list[State]           ivy_art.py:77   — the graph nodes
transitions  list[(State,op,label,State)]  :78   — directed edges
covering     list[(State,State)]   :79            — covering pairs
pvars        predicate variables   :80
state_graphs list                  :81

ivy_interp.State (ivy_interp.py:83) — one ARG node:
domain    module/interpreter       :90
moded, clauses, precond  (from value tuple)  :91–92
expr      action-application expr  :95   (encodes how state was reached)
label     display label            :96
cached_update  transition relation :97
cached_pred    predecessor state   :98
id        int (set when added)     ivy_art.py:254
safe      bool (safety flag)       set in check_bounded_safety
universe  model universe           set from BMC result

State.value (ivy_interp.py:101) is a (moded, clauses, precond) triple where clauses is
the abstract formula representing the state. State.conjs and State.unders are global on
the domain (not per-node) — a design wart explicitly commented at :107–114.

Layer 2: Concept/Proof Model (concept.py, concept_interactive_session.py)

ConceptInteractiveSession (concept_interactive_session.py:26) — proof state for one
concept graph panel:
domain              ConceptDomain   :41   — active predicates
state               formula         :42   — current abstract state formula
axioms              formula         :43
goal_constraints    list[formula]   :44
suppose_constraints list[formula]   :45   — user-added emptiness constraints
abstract_value      dict            :93   — cached result of alpha()
undo_stack          list[(domain,suppose_constraints)]  :47
cache               dict            :51   — formula cache (shallow-copied)

concept.ConceptDomain (concept.py) — the set of predicates the user has chosen to
track:
- concepts: dict of name → Concept (quantifier-free formula)
- combiners: second-order formulas (node/edge combiners like at_least_one, all_to_all,
etc.)
- combinations: the instantiations to evaluate

Layer 3: Render Model (cy_elements.py, cy_render.py, dot_layout.py)

CyElements (cy_elements.py:5) — the serializable intermediate representation, totally
independent of both computation and browser:
elements  list[dict]     :42   — JSON-serializable list
node_id   dict[obj→str]  :43   — object → 'n{i}'
edge_id   dict[(obj,src,tgt)→str]  :44
shape_id  dict[obj→str]  :45

Each element dict has group ('nodes'/'edges'), data (id, label, classes, short_info,
long_info, events, actions), and positional data added later by dot_layout.

The rendering pipeline is a pure function chain:
AnalysisGraph
  → cy_render.render_rg()     (ivy_art.py:as_cy_elements, cy_render.py:276)
  → dot_layout.dot_layout()   (dot_layout.py:136) — adds x,y,width,height,bspline
  → CyElements (now with positions)
  → JSON to JS/browser

dot_layout is a pure function: given CyElements with no positions, it shells out to
graphviz dot, parses the output, and returns CyElements with positions. Entirely
testable without a browser.

cy_render.render_rg() (cy_render.py:276) is also pure: 
given an AnalysisGraph and callback lists, it produces 
CyElements. Testable with just the graph data.

Summary: The Python Data Model in One Diagram

AnalysisGraph (ivy_art.py:72)
  └─ []State (ivy_interp.py:83)
       ├─ clauses   (abstract formula)
       ├─ expr      (how we got here)
       └─ id, safe, universe
  └─ []transition  (State, op, label, State)
  └─ []covering    (State, State)

ConceptInteractiveSession (concept_interactive_session.py:26)
  └─ ConceptDomain (concept.py)
       └─ concepts: dict[name → Concept]
  └─ abstract_value: dict  (output of alpha())
  └─ undo_stack

CyElements (cy_elements.py:5)         ← render-side, serializable
  └─ elements: []dict                 ← sent as JSON to browser

This is a solid data model. 

# port of the data model to Go

Layer 1 — Semantic model (art.go, goivy package):
- AnalysisGraph { States []*State; Transitions []Transition; Covering []CoveringPair;
PVars []Expr }
- State { ID; Clauses; IsBottom(); Label; ... }
- Transition { Pre *State; Op ActionsAction; Label string; Post *State }

Layer 1b — Typed render intermediates (art_cyrender.go):
- AnalysisGraphState { States []ARGNode; Transitions []ARGTransition; Covering
[]ARGCover }
- ARGNode { ID int; Label string; IsBottom bool; Info string }
- ARGTransition { SourceID; TargetID; Label; IsJoin }
- ARGCover { CoveredID; CoveringID }

Layer 2 — Concept/proof model (webui_graph_model.go, webui_concept_isession.go,
webui_concept_domain.go):
- Graph { Sorts; ConceptSess *ConceptSession; Checks *DisplayCheckboxes; GraphStack }
- DisplayCheckboxes { EdgeDisplayCheckboxes map[string]map[string]*Option;
NodeLabelDisplayCheckboxes ... }
- Full ConceptInteractiveSession with domain, undo/redo stack, suppose constraints

Layer 3 — Cytoscape render (art_cyrender.go):
- CyElements { Elements []CyElement; NodeID map[string]string; EdgeID map[string]string
 }
- CyElement { Group; Data map[string]interface{}; Classes; Position *CyPosition }

# port of the data model to webui/ typescript

## TypeScript Class Inventory

The TypeScript port lives in `frontend/src/models/uiDataModel.ts`.  Every class
extends `RawBackedModel<TRaw>` which stores the original JSON so callers can always
fall through to `raw` for unparsed fields.  A shared set of helper functions
(`pick`, `stringValue`, `intValue`, `boolValue`, `boolMap`, `typedArray`, …) provide
safe, zero-default parsing across snake_case, camelCase, and PascalCase key variants
so that the same class works whether the JSON came from Go or Python.

### Layer 1 — Semantic / ARG (render-ready typed intermediates)

| TypeScript class    | Go counterpart              | Wire key (in ARGSnapshot) |
|---------------------|-----------------------------|---------------------------|
| `ARGNode`           | `ARGNode` in art_cyrender.go | `analysis_graph_state.states[]` |
| `ARGTransition`     | `ARGTransition`             | `analysis_graph_state.transitions[]` |
| `ARGCover`          | `ARGCover`                  | `analysis_graph_state.covering[]` |
| `AnalysisGraphState`| `AnalysisGraphState`        | `analysis_graph_state` sub-object |
| `ARGSnapshot`       | result of `WebUIARGPayload` | top-level ARG HTTP response |

`ARGSnapshot` reads the flat `{elements, analysis_graph_state?}` shape exactly as
`WebUIARGPayload` emits it.  `analysis_graph_state` is absent when the graph is
empty (Go omits it; TypeScript falls back to empty typed arrays).

The `State`, `AnalysisGraph`, and `AnalysisTransition` classes also exist in the
file for future save/restore and CTI state annotation use.  They are **not** wired
to the current HTTP response (Go does not send a full `AnalysisGraph` object on the
wire — only `AnalysisGraphState`).  `State` carries `expr` and `universe` for CTI
visualization once that pathway is wired.

### Layer 2 — Concept / Proof model

| TypeScript class              | Go counterpart                              | Wire key |
|-------------------------------|---------------------------------------------|----------|
| `Concept`                     | `webui.Concept`                             | inside `concept_domain.concepts` |
| `ConceptCombiner`             | `webui.ConceptCombiner`                     | `concept_domain.combiners[]` |
| `ConceptDomain`               | `webui.ConceptDomain`                       | `concept_domain` |
| `ConceptSession`              | `webui.ConceptSession`                      | `concept_session` |
| `ConceptInteractiveSession`   | `webui.ConceptInteractiveSession` via payload| `concept_interactive_session` |
| `GraphStack`                  | result of `conceptGraphStackPayload()`      | `graph_stack` |
| `DisplayCheckboxes`           | `webui.Toggles` (= `DisplayCheckboxes.Snapshot()`) | `display_checkboxes` / `toggles` |
| `ConceptGraphModel`           | result of `conceptGraphPayload()`           | `graph` sub-key |
| `ConceptSnapshot`             | full `GetConcept` response                  | top-level concept HTTP response |

`ConceptDomain` JSON is snake_case throughout (`concepts`, `combiners`, `nodes`,
`edges`, `node_labels`) and parses cleanly.

`ConceptSession` wire is `{domain, abstract_value}` only — `undoStack` is unexported
in Go and never serialized.  TypeScript `ConceptSession` correctly has no
`undoDepth`/`redoDepth`.

`ConceptInteractiveSession` is the most complex class.  Key points:
- `abstract_value` on the wire is `[]TagValue` (array of `{Tag:string[], Value:bool}`),
  not `Record<string,bool>`.  The helper `tagValueArrayToMap()` converts it to
  `Record<string,bool>` keyed by `"|"`-joined tag parts (e.g.
  `"node_info|at_least_one|client"`, `"edge_info|all_to_all|link|client|server"`).
- `undoDepth` and `redoDepth` are present here (from `undo_depth`/`redo_depth`).
- `state`, `axioms`, `goalConstraints`, `supposeConstraints`, `cache`, `info` are
  all populated from the matching snake_case wire keys.

`GraphStack` has exactly the 4 scalar fields that `conceptGraphStackPayload()` emits:
`canUndo`, `canRedo`, `undoDepth`, `redoDepth`.  No graph/stack content is sent.

`DisplayCheckboxes` accepts both `display_checkboxes` and `toggles` as source keys
(Go sends both, sometimes only `toggles`), and both `edges`/`labels` sub-maps.

### Layer 3 — Render

| TypeScript class | Go counterpart                  | Notes |
|------------------|---------------------------------|-------|
| `CyElement`      | `WebUICyElement`                | `group`, `data`, `classes`, `locked`, `position?` |
| `CyPosition`     | `WebUICyPosition`               | `x`, `y` |
| `CyElements`     | `WebUICyElements`               | `elements[]` only; `NodeID`/`EdgeID` are `json:"-"` in Go and absent from wire |

### Top-level snapshot classes

`ARGSnapshot` and `ConceptSnapshot` are the two entry points that the service layer
calls.  `UIDataModel.acceptArgSnapshot()` and `UIDataModel.acceptConceptSnapshot()`
dispatch to them and store the result per-sheet.

`ConceptSnapshot` field map vs `GetConcept` response keys:

| `ConceptSnapshot` field | `GetConcept` JSON key          |
|-------------------------|--------------------------------|
| `domain`                | `concept_domain`               |
| `session`               | `concept_session`              |
| `interactiveSession`    | `concept_interactive_session`  |
| `displayCheckboxes`     | `display_checkboxes` or `toggles` |
| `graphStack`            | `graph_stack`                  |
| `graph`                 | `graph`                        |
| `render`                | `elements` (root-level array)  |
| `selectedNode`          | `selected_node`                |
| `stateLabel`            | `state_label`                  |
| `facts`                 | `facts`                        |

---

## Faithfulness assessment

**High-fidelity mappings (no known divergence):**
- ARG render intermediates (`ARGNode`, `ARGTransition`, `ARGCover`, `AnalysisGraphState`) match Go struct JSON tags exactly.
- `Concept`, `ConceptCombiner`, `ConceptDomain` parse Go's snake_case JSON cleanly.
- `ConceptSession` wire shape (`domain` + `abstract_value`) matches exactly.
- `GraphStack` four-scalar contract matches `conceptGraphStackPayload()` exactly; locked by `TestConceptGraphStackPayload`.
- `CyElements` correctly excludes `nodeId`/`edgeId` (Go `json:"-"`); locked by `TestCyElementsNodeIDNotSerialized`.
- `DisplayCheckboxes` handles both `display_checkboxes` and `toggles` key aliases.
- `ARGSnapshot` conditional `analysis_graph_state` matches `WebUIARGPayload()` (absent when empty); locked by `TestConformARG`.
- TagValue array → bool-map conversion tested in `TestConceptInteractiveSessionPayloadTagValues` (Go) and multiple TypeScript tests.

**Fixed limitations** (code changes applied):

1. ~~CIS `domain` field is always empty~~ — **Fixed.**  Removed `"domain"` from
   `conceptInteractiveSessionPayload()` (`webui_session.go`).  `CDConceptDomain` had
   unexported fields and serialized to `{}`.  The authoritative domain is always
   `snapshot.domain` from the top-level `concept_domain` key.
   Locked by `TestCISPayloadNoDomainKey`.

2. ~~`State.clauses` and `State.expr` are `unknown`~~ — **Fixed.**  `ArtToGraphState()`
   (`webui_ui_main.go`) now populates `ARGNode.Info` with `state.Clauses.String()`
   instead of the hardcoded `"State %d"` fallback.  The formula string is therefore
   reachable via `ARGNode.info` in TypeScript (already parsed correctly).
   Locked by `TestARGNodeInfoIsFormula`.

5. ~~`facts` array is `unknown[]`~~ — **Fixed.**  `FactSelection` TypeScript class added
   to `uiDataModel.ts`; `ConceptSnapshot.facts` retyped from `unknown[]` to
   `FactSelection[]`.  Wire format: `{index:int, text:string, selected:bool}`.
   Locked by three TypeScript tests in `uiDataModel.test.ts`.

7. ~~`concept_domain` is null in fresh sessions~~ — **Fixed.**  `GetConcept()`
   (`webui_backend_go.go`) now initialises `conceptDomain` to `NewConceptDomain()`
   (empty but non-nil) before the `SimpleSess` guard, eliminating the `nil` JSON value.
   Locked by `TestConceptDomainNeverNull`.

3. **Full CTI graph is now on the wire via two new endpoints** — **Implemented.**
   `GET /api/session/{id}/arg?full=true` returns `FullARGNode` entries with `clauses`,
   `action_name`, and `universe` per state.
   `GET /api/session/{id}/arg/cti` returns the CTI analysis graph with those same
   full nodes plus `have_cti` (bool) and `current_conjecture` (string).
   TypeScript: `ARGSnapshot.fullAnalysisGraphState` (type `FullAnalysisGraphState`)
   is always populated from the same `analysis_graph_state` sub-key; `CTISnapshot`
   extends `ARGSnapshot` with `haveCti` and `currentConjecture`.
   All new Go endpoints go through the `IvyApiAdapter` choke point:
   `IvyApiAdapter.getCTIARG()` → `getSnapshot({ctiArg:{}})`; 
   `HostedGoIvyApiAdapter.getArgSnapshot({full:true})` adds `?full=true` to the ARG URL.
   Locked by `TestFullARGPayloadStructure`, `TestCTIARGPayloadShape`, `TestFullARGNodeUniverse*`
   (Go) and 7 new TypeScript tests in `uiDataModel.test.ts`.
   `UIDataModel.acceptCtiSnapshot()` stores the result in `SheetModel.cti`.

**Accepted as by-design** (no code change needed):

4. **`abstract_value` at top level is node-label keys only.**  This is by design: the
   top-level `abstract_value` feeds the node-label checkbox panel (node_label| keys);
   `ConceptInteractiveSession.abstractValue` (TagValue-converted map) carries the full
   picture including node_info and edge_info keys.  Use the CIS field for concept-panel
   truth-value display.

6. **`ConceptGraphModel` is zero-valued before file load.**  `GetConcept` returns
   `conceptGraphPayload(nil, nil)` until a `GraphWidget` is associated with the session.
   All fields zero-default cleanly in TypeScript.  Guard with `graph.sorts.length > 0`
   before rendering sort-dependent UI.

8. **`pvars` / `state_graphs` not serialized.**  `AnalysisGraph.PVars` is `[]Expr` (an
   interface slice with no JSON tags) and `StateGraphs` is `[]interface{}`.  Neither is
   used in the webui package nor required by any GUI behavior in PLAN378.  Out of scope.

---

## Guidance for next GUI porting work

The following Python GUI behaviors (from `PLAN378_cc_python_gui_inventory.md`) depend
on this data model and need to be ported to Go/TypeScript:

### ARG panel behaviors

- **Node selection** — User clicks ARG node; browser sends `nodeId` to
  `GET /api/session/{id}/concept?node={nodeId}&sheet={sheetID}`.  Go calls
  `sess.selectConceptARGNode()` which resolves to `selectedNode`/`stateLabel` in the
  `ConceptSnapshot` response.  TypeScript: read `snapshot.selectedNode` and
  `snapshot.stateLabel`; highlight the node via `ARGNode.id` lookup in
  `analysisGraphState.states`.

- **Safety status coloring** — Each `ARGNode.isBottom` flag marks an unsafe/bottom
  state.  TypeScript `ARGNode.isBottom` is correctly wired.  For full formula data,
  use `getARG({full:true})` (via `IvyApiAdapter`) to get `FullARGNode` entries with
  `clauses`, `actionName`, and `universe` populated.  `ARGSnapshot.fullAnalysisGraphState`
  is parsed from the same response key; `analysisGraphState` (lightweight) is always
  present alongside it.

- **Covering arcs** — `ARGCover.coveredId` / `CoveringId` pairs are on the wire.
  Render them as dashed/dotted edges in Cytoscape distinct from transition edges.

### Concept panel behaviors

- **Checkbox toggles** — `DisplayCheckboxes.edgeVisible(name, cls)` and
  `nodeLabelVisible(name, cls)` are the read side.  The write side (user toggles a
  checkbox) sends `POST /api/session/{id}/toggle` with `{edge, display_class}` or
  `{node_label, display_class}`.  After any toggle, refresh the concept render by
  calling `GetConcept`.

- **Undo/redo** — `GraphStack.canUndo`, `canRedo`, `undoDepth`, `redoDepth` drive
  button enable/disable.  Post to the undo/redo action endpoints; the response is a
  fresh `ConceptSnapshot`.  The `ConceptInteractiveSession` also has its own
  `undoDepth`/`redoDepth` (from the CIS undo stack, separate from the Graph undo
  stack).

- **Abstract value display** — The concept panel shows each node/edge concept's
  truth value in the current abstract state.  Read `interactiveSession.abstractValue`
  (full TagValue-converted map).  Keys are `"node_info|<combiner>|<concept>"` for
  nodes and `"edge_info|<combiner>|<relation>|<src_sort>|<tgt_sort>"` for edges.
  The display combiner classes (`at_least_one`, `none`, `all_to_all`, `none_to_none`,
  `node_necessarily`) are the middle tag component.

- **Suppose constraints** — `interactiveSession.supposeConstraints` is the list of
  user-added emptiness constraints.  Display these in the suppose panel; post additions/
  removals to the corresponding action endpoint.

- **Goal constraints** — `interactiveSession.goalConstraints`; same pattern as suppose.

- **Split concept** — Triggered from concept-graph node right-click.  After a split,
  `ConceptDomain.concepts` gains two new entries (name+"+" and name+"-").  Requires
  a full `GetConcept` refresh; `snapshot.domain.concepts` is the authoritative source.

- **CTI (counterexample inspection)** — Call `getCTIARG()` (via `IvyApiAdapter`) to
  fetch the CTI analysis graph.  `CTISnapshot.haveCti` is the gate; when true,
  `CTISnapshot.fullAnalysisGraphState.states[i].clauses` has the abstract-state formula,
  `.actionName` has the triggering action, and `.universe` (if non-null) has the BMC
  concrete universe (`Record<string, string[]>`, sort → element names).
  `CTISnapshot.currentConjecture` is the conjecture currently being checked.
  The `ConceptInteractiveSession.state` string is the current abstract state formula
  (already on the wire as a formatted string).

- **Facts / active-facts panel** — `ConceptSnapshot.facts` is the list of
  `FactSelection` objects (active suppose-constraints shown as checkboxes in the
  concept panel).  Currently `unknown[]`; model `FactSelection` when porting this
  panel.

### Rendering pipeline

The Go render pipeline mirrors Python exactly:

```
AnalysisGraph / ConceptSession
  → RenderARG() / RenderConceptGraph()  (art_cyrender.go / webui_cyrender.go)
  → WebUICyElements (with octagon nodes, edge classes, per-sort colors)
  → GetARG / GetConcept HTTP response
  → ARGSnapshot.render / ConceptSnapshot.render (TypeScript CyElements)
  → Cytoscape.js instance
```

Go does **not** shell out to graphviz for layout (unlike Python's `dot_layout.py`).
Position data is expected to come from Cytoscape's built-in layout algorithms on the
browser side.  The `positions` field in the ARG payload has been removed; the browser
is responsible for layout.

When porting Python layout-dependent behaviors (e.g., edge label placement,
bspline routing), use Cytoscape layout options rather than pre-computed dot positions.

### Key invariants to preserve during porting

- `ARGSnapshot.analysisGraphState.states[i].id` is the canonical integer ID used to
  link ARG nodes to concept state selection.  Preserve this as the `data.id` key in
  the matching `CyElement` (Go stores it as `"state_{id}"`).

- `ConceptInteractiveSession.abstractValue` keys use `"|"` as separator — never `","`.
  Five-part edge keys are `"edge_info|<combiner>|<rel>|<src>|<tgt>"`.

- `DisplayCheckboxes` and `Toggles` are the same data serialized as a `Snapshot()`.
  TypeScript should always read `display_checkboxes` first, then fall back to
  `toggles`, never merge both.

- `GraphStack.undoDepth` counts `Graph`-level undo entries (concept split/join/undo).
  `ConceptInteractiveSession.undoDepth` counts CIS-level undo entries
  (suppose/goal/alpha refinement steps).  These are independent stacks with separate
  undo action endpoints.

