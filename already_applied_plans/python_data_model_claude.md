  c) The Python GUI's Data Model

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

  cy_render.render_rg() (cy_render.py:276) is also pure: given an AnalysisGraph and
  callback lists, it produces CyElements. Testable with just the graph data.

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

  This is a solid data model. The Go port already mirrors most of this — the question is
  whether the Go structs are complete and cover all three layers.

----

-----------------

The data model audit is complete. Ready for your next task.

Q: a compaction interrupted you and may have prevented you from actually answering the
  question, which is: tell me how much of the data model that you described just now is
  present a) in javascript files and b) in Go files. Answer now in this discussion.

Claude: Let me verify the current state of the key files before answering.

Here is the complete picture:

---
Go side — all three layers fully present

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

---
JS side — partial coverage only

┌─────────────────────────┬──────────┬─────────────────────────────────────────────┐
│          What           │ Present? │                    Where                    │
├─────────────────────────┼──────────┼─────────────────────────────────────────────┤
│ Session metadata (mode, │ ✅       │ sessionStore.js                             │
│  status, sessionId)     │          │                                             │
├─────────────────────────┼──────────┼─────────────────────────────────────────────┤
│ Toggle/checkbox state   │          │ stateRelationsStore.js —                    │
│ Toggle/checkbox state         │ ✅       │ rows[].checked.{all_to_all,           │
│ mirroring DisplayCheckboxes   │          │ edge_unknown, none_to_none,           │
│                               │          │ transitive}                           │
├───────────────────────────────┼──────────┼───────────────────────────────────────┤
│ Cytoscape style constants     │ ✅       │ graphRuntime.js — CONCEPT_STYLE,      │
│                               │          │ ARG_STYLE, PROOF_STYLE                │
├───────────────────────────────┼──────────┼───────────────────────────────────────┤
│ Raw Cytoscape element arrays  │          │ graphStore.js —                       │
│ (opaque dicts from server)    │ ✅       │ argBySheet[sheetId].elements,         │
│                               │          │ conceptBySheet[sheetId].elements      │
├───────────────────────────────┼──────────┼───────────────────────────────────────┤
│ Typed ARGNode / ARGTransition │ ❌       │ absent                                │
│  / ARGCover                   │          │                                       │
├───────────────────────────────┼──────────┼───────────────────────────────────────┤
│ State with id, clauses, safe, │ ❌       │ absent                                │
│  label                        │          │                                       │
├───────────────────────────────┼──────────┼───────────────────────────────────────┤
│ AnalysisGraph with typed      │ ❌       │ absent                                │
│ states/transitions/covering   │          │                                       │
├───────────────────────────────┼──────────┼───────────────────────────────────────┤
│ ConceptDomain,                │          │                                       │
│ ConceptInteractiveSession,    │ ❌       │ absent                                │
│ undo stack                    │          │                                       │
├───────────────────────────────┼──────────┼───────────────────────────────────────┤
│ GraphStack (undo/redo for     │ ❌       │ absent                                │
│ concept domain)               │          │                                       │
└───────────────────────────────┴──────────┴───────────────────────────────────────┘

---
Bottom line: The JS client treats all graph data as opaque server-sent Cytoscape JSON.
It has Layer 2 toggle state and Layer 1 session metadata, but no typed representations
of the ARG nodes/transitions, no State objects, and no concept session model. Layers 1
and 2 exist exclusively on the server (Go). The JS would need typed counterparts — at
minimum ARGNode/Transition/Cover and a concept-domain checkbox model — to support
client-side graph operations like selection, marking, undo, or any computation that
shouldn't require a round-trip.
