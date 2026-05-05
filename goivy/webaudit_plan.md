# Web UI Implementation Plan

This plan topologically sorts the findings in `webaudit.md` into a bottom-up
implementation order. The rule is: build shared UI/API substrate first, then
wire individual Python behaviors, then tackle higher-level CTI and event-viewer
flows.

Each step should be implemented test-first where practical. The normal gate for
completed steps is:

```sh
make test && make test-web
```

## Dependency Principles

1. Browser primitives before browser workflows.
   Dialogs, errors, menus, selected-state routing, and graph-instance ownership
   are prerequisites for many visible Tcl/Tk behaviors.

2. Backend operation shape before UI buttons.
   If the Python operation needs relation/source/target/truth, selected facts,
   a bound, or a chosen conjecture, the HTTP/Go API must carry that data before
   the browser menu is wired.

3. Python-shaped Go methods before simplified session shims.
   Where `webui_graph_widget.go`, `webui_ui_main.go`, or `webui_ui_cti.go`
   already mirror Python, route browser actions through those methods instead
   of reimplementing a second simplified path in `Session.ExecuteAction`.

4. Avoid exposing nonfunctional menu entries.
   Dynamic menu rendering is useful only if actions can be disabled, hidden, or
   wired to real implementations as each phase lands.

## Topological Implementation Order

### 0. Preserve Web-Only Behaviors As Explicit Non-Conformance

Findings: 36

The CodeMirror editor, browser file handles, dirty-close dialog, external-change
dialog, and tutorial browser are web-only additions. Do not try to make them
look like Tcl/Tk behavior. Instead, mark them as web-only in tests and docs so
future conformance work does not confuse these with Python-port gaps.

Verification:
- Add or update browser tests only if they currently assert Python parity for
  these web-only features.

### 1. Fix Tiny Global Mismatches First

Findings: 2, 14

Set the browser default mode to PDR so it matches Python and Go's `DefaultMode`.
Add web-side ARG edge-label display post-processing for `-[` and `]-`, matching
the Tk rewrite.

Dependencies: none.

Why first: these are small, visible, low-risk fixes that reduce noise while
larger infrastructure is still changing.

Verification:
- Browser/unit test for initial mode value.
- Rendering test for ARG edge labels containing `{`/`}`.

### 2. Implement A Python-Style Web Run Context

Findings: 7

Introduce a shared action runner for browser-triggered operations:
- Shows busy/loading state.
- Surfaces backend errors consistently.
- Treats unknown/unwired actions as errors, not success.
- Produces structured results for dialogs, graph updates, file downloads, and
  tab opens.

On the Go side, stop accepting unknown generic actions as successful. Return a
clear error unless an action is explicitly declared as intentionally inert.

Dependencies: none.

Unlocks: every later menu command can report Python-like errors and avoid false
green UI states.

Verification:
- Unit/API test that an unknown action returns an error.
- Browser test that an errored action is visible in the UI and does not report
  success.

### 3. Build The Missing Dialog Primitives

Findings: 6

Implement reusable browser dialogs corresponding to Python:
- OK dialog.
- OK/Cancel dialog.
- Text dialog with optional custom command label and optional Cancel.
- Entry dialog.
- Integer dialog with min/max validation.
- Listbox dialog with single-select and multi-select.
- Button-list dialog with arbitrary commands.

These should return Promises so workflows can be written in the same shape as
Python: ask, then continue with the selected value.

Dependencies: step 2.

Unlocks: findings 10, 12, 17, 22, 27, 28, 32, and parts of 21/29/33.

Verification:
- JS/browser tests for each dialog type.
- At least one Go/browser integration test that drives a list selection and an
  integer prompt.

### 4. Establish A Dynamic Menu/Action Descriptor Pipeline

Findings: 1, 15

Create a single browser menu/action descriptor format that can be supplied by Go
and rendered by JS. Include:
- Menu groups and separators.
- Context-menu actions.
- Button/radio/list/entry/int-dialog action metadata.
- Enabled/disabled state.
- Click semantics metadata where web intentionally diverges from Tcl/Tk.

Then make the static browser menu consume descriptors for at least the ARG and
concept graph menus. Keep the web convention of right-click context menus if
desired, but encode the divergence explicitly.

Dependencies: steps 2 and 3.

Unlocks: safer incremental wiring of findings 9, 10, 17, 18, 22, and 32.

Verification:
- Unit test serializing Go menu specs to browser descriptors.
- Browser test that a known Go-supplied menu item appears and dispatches through
  the shared runner.

### 5. Fix Selected ARG State Routing

Findings: 8

Make `GET /concept?node=...` meaningful. The handler/backend should accept the
node id, call the Python-shaped `AnalysisGraphUI.ViewState`/concept graph path,
and return the concept graph for that ARG state.

Dependencies: step 2.

Unlocks: almost every graph workflow; otherwise commands appear to act on one
state while rendering another.

Verification:
- API test with at least two ARG states proving `?node=state_0` and
  `?node=state_1` can return different parent-state labels/data.
- Browser test clicking two ARG nodes and observing the state panel/concept
  graph update to the selected node.

### 6. Make Checkbox State Backend-Owned And Round-Trippable

Findings: 19

Move relation/node-label checkbox truth into the backend graph state, with JS as
a view/controller. Match Python `sync_checkboxes` and `reverse_sync_checkboxes`:
- Default labels get `+`.
- Relation toggles update graph rendering.
- Undo/redo restores checkbox state.
- The server returns current checkbox state with each concept graph response.

Dependencies: steps 4 and 5.

Unlocks: fact gathering, CTI relation display, diagram display, graph layout
conformance, and save/restore.

Verification:
- Unit test for default label `+` behavior.
- Unit/API test that toggles survive undo/redo.
- Browser test that toggling a relation changes edge visibility and persists
  through refresh of the concept graph.

### 7. Implement Constraint/Facts Text Selection

Findings: 20

Render concept graph constraints/facts as a selectable list, matching Python's
click-to-toggle constraints below the graph. Wire active facts to the backend
instead of leaving `GetActiveFacts`/`HighlightSelectedFacts` as stubs.

Dependencies: steps 3, 5, and 6.

Unlocks: CTI strengthen, minimize, sufficient/inductive checks, and faithful
gather behavior.

Verification:
- Unit test for `GetActiveFacts`.
- Browser test toggling a fact and then gathering/strengthening from only active
  facts.

### 8. Repair Source Browsing

Findings: 5

Return enough source data for "View Source" to work:
- `file`
- `lineno`
- source text, or a route to fetch source text by filename

Then update the editor/source pane and line highlight exactly when the backend
reports a source location.

Dependencies: steps 2 and 5.

Verification:
- API test for `view_source_edge` result shape.
- Browser test invoking View Source and verifying editor selection/highlight.

### 9. Build Real Web Sheet/Graph Instance Ownership

Findings: 4

Replace the single global `argGraph`/`conceptGraph` assumption with a sheet
model:
- Each sheet owns its ARG graph instance.
- Each sheet owns its concept graph instance or explicitly has none.
- Sheet removal follows Python semantics where possible, while preserving any
  web-only "main tab cannot close" decision as a documented divergence if kept.
- Backend session knows which sheet/ARG a browser action targets.

Dependencies: steps 2, 4, and 5.

Unlocks: decompose/step-in, trace viewing, reachable tree tabs, remembered goal
viewing, and analysis-state save/load.

Verification:
- Browser test opening two sheets and dispatching an action to the non-active
  then active sheet without cross-updating the wrong graph.

### 10. Make Step-In/Decompose A Full Analysis UI

Findings: 13

Route decompose through the Python-shaped `AnalysisGraphUI.DecomposeEdge`, then
open a real new sheet whose graph actions, selected-state behavior, source
browsing, and concept graph all work like the main sheet.

Dependencies: steps 5, 8, and 9.

Verification:
- Browser test "Step in" creates a new sheet and then a node click in that sheet
  loads that sheet's concept graph.

### 11. Wire ARG Execute-Action Menus

Findings: 9

Expose `AnalysisGraphUI.NodeExecuteCommands` through node action descriptors.
Dispatch selected execute actions to the backend with the selected state id and
actual action identity, not just a label string.

Dependencies: steps 4, 5, and 9.

Verification:
- Unit/API test that state actions appear in node actions sorted as Python.
- Browser test executing an exported action from a state and seeing a new ARG
  state/edge.

### 12. Implement ARG Choice-Backed Commands

Findings: 10, 22

Use the dialog primitives to port:
- Try conjecture: list undecided conjectures, browse source when available, then
  dispatch with the chosen conjecture.
- Remember graph: prompt for a name and store a copy under that name.
- Try remembered goal: list remembered names and restore the chosen graph.

Dependencies: steps 3, 4, 5, 9, and 11.

Verification:
- Unit/API tests for named remembered graphs.
- Browser test for Try conjecture dialog selection.
- Browser test remembering two names and recalling the chosen one.

### 13. Repair Safety/Counterexample Trace Viewing

Findings: 12

Port Python's safety result flow:
- Bounded safety failure offers a "View" action that opens the trace ARG in a
  sheet.
- Local safety failure offers "View unsafe states" and "View concrete trace"
  when available.
- Check results update node color and graph state consistently.

Dependencies: steps 3, 5, 9, and 10.

Verification:
- Unit test for safety action result metadata.
- Browser test opening a counterexample/unsafe trace sheet from a failed check.

### 14. Add Show Reachable States

Findings: 11

Implement Python `show_reachable_states` and `reachable_tree` behavior in the
web: create or reuse the reachable tree ARG and open it as a sheet.

Dependencies: steps 9 and 13.

Verification:
- Unit/API test for reachable tree creation.
- Browser test menu action opens a reachable-state sheet.

### 15. Fix Concept Edge Materialization API Shape

Findings: 16

Replace `/concept/materialize`'s single `concept` payload with distinct node and
edge operations, or a tagged operation that actually carries:
- relation id
- source node id
- target node id
- truth value

Route edge materialization through `GraphWidget.MaterializeEdge` and
`DematerializeEdge`, not `SimpleSess.Materialize`.

Dependencies: steps 2, 4, 5, and 6.

Verification:
- API test proving positive and negative edge materialization call different
  backend paths and update graph state.
- Browser test using Materialize + and Materialize - on an edge.

### 16. Implement Materialize Edge From Selected

Findings: 17

Port Python's selected-node edge materialization:
- Select source node.
- On target node, list binary concepts whose sorts match source/target.
- Materialize the chosen relation.

Dependencies: steps 3, 6, 7, and 15.

Verification:
- Browser test selecting two nodes, choosing a relation, and seeing the witness
  relation/node labels appear.

### 17. Implement Projection Discovery And Actions

Findings: 18

Replace the `GetNodeProjectionActions` stub with real projection discovery from
the concept graph/session. Show "Add projection..." only when projections exist.

Dependencies: steps 4, 5, and 6.

Verification:
- Unit test for projection action discovery.
- Browser test adding a projection from the node context menu.

### 18. Make Add Relation Use Full Python Concept Parsing

Findings: 23

Route Add relation through the real concept-domain parser/equivalent of
Python's `g.string_to_concept`, preserving variables, sorts, arity, and concept
metadata. Keep the entry dialog from step 3 instead of raw `prompt()`.

Dependencies: steps 3, 4, and 6.

Verification:
- Unit/API test adding a unary and binary relation and checking derived sort and
  arity metadata.
- Browser test Add relation makes a usable relation, not only a label.

### 19. Fix Backtrack

Findings: 24

Route browser Backtrack through `GraphWidget.Backtrack`, or implement the same
loop over backtrack points in the session path. Do not map it to one undo.

Dependencies: steps 5 and 6.

Verification:
- Unit test with multiple undo frames and one backtrack point.
- Browser test Backtrack skips intermediate non-backtrack undo frames.

### 20. Repair Concrete/Gather/Reach Semantics

Findings: 25

Replace simplified `Session.ExecuteAction` implementations with Python-shaped
concept graph operations:
- `concrete` modifies the current graph state with concrete constraints.
- `gather` uses displayed relation values and active fact/checkbox state.
- `reach` and `path_reach` follow Python parent-state/constraint behavior and
  dialog flow.

Dependencies: steps 3, 5, 6, 7, and 9.

Verification:
- Unit tests comparing concrete/gather/reach behavior to Python-facing expected
  structures where possible.
- Browser test Gather changes selected facts/goal in a way visible in the graph.

### 21. Implement DOT Export

Findings: 21

Make concept graph Export produce a DOT file/content equivalent to Python's
display export. Keep browser download/save-as mechanics, but the content should
be DOT, not a fact-count event.

Dependencies: steps 6 and 20.

Verification:
- Unit/API test export content begins with a DOT graph and includes expected
  nodes/edges.
- Browser test Export triggers a file/download path with `.dot` content.

### 22. Start CTI UI During Session Load

Findings: 26

When the browser is in the CTI/invariant workflow, instantiate and start
`CTIAnalysisGraphUI` rather than only generic `AnalysisGraphUI`. Load module
conjectures, view state 0, autodetect transitive relations, and preserve CTI
state.

Dependencies: steps 5 and 6.

Unlocks: CTI-specific diagram, used-relations, BMC, weaken, strengthen, and
save-invariant behavior.

Verification:
- Unit/API test that loading a model initializes CTI conjectures and transitive
  relation checkboxes like Python.

### 23. Fix Used-Relation Display After CTI Failure

Findings: 30

Port Python `show_used_relations` precisely:
- Clear edges.
- Enable `+` only for relations actually used by the failing clauses.
- Support `both=True` for `-`.
- Add generated concepts for three-argument numeral apps.

Dependencies: steps 6 and 22.

Verification:
- Unit test with clauses using one relation out of several.
- Browser/check test that only the used relation is auto-checked after CTI.

### 24. Route CTI Diagram Through CTIAnalysisGraphUI

Findings: 31

Make the browser CTI Diagram action call `CTIAnalysisGraphUI.Diagram`, not
`SimpleSess.Diagram`. It should check inductiveness if needed, view the diagram
state, show used relations with `both=True`, and gather facts.

Dependencies: steps 7, 22, and 23.

Verification:
- Unit/API test that CTI diagram invokes the CTI path and returns diagram facts.
- Browser test Diagram updates relation toggles and selected facts.

### 25. Implement CTI Bounded Check Bound Prompt

Findings: 27

Use the integer dialog to ask for the BMC bound, remember the current bound, and
send it to the backend. Remove the fixed hidden bound of 10 from the browser
path.

Dependencies: steps 3 and 22.

Verification:
- Browser test entering a bound and checking that the backend receives it.
- Unit test that repeated BMC prompts default to the last bound.

### 26. Implement CTI Weaken Selection

Findings: 28

Use a multi-select listbox to choose conjectures, pass selected indices or
stable conjecture ids to the backend, remove them, clear CTI state, and show the
removed conjectures in a text dialog.

Dependencies: steps 3 and 22.

Verification:
- Browser test weakening two selected conjectures.
- Unit/API test that old/current conjecture lists update correctly.

### 27. Wire CTI Concept-Graph Actions

Findings: 32

Expose and implement the CTI concept graph menu:
- Gather.
- Bounded check.
- Minimize.
- Check sufficient.
- Check relative induction.
- Strengthen.
- Export.

Use active facts from step 7 and dialogs/tabs from earlier steps. Counterexample
"View" buttons should open ARG sheets.

Dependencies: steps 3, 7, 9, 13, 22, 25, and 26.

Verification:
- Unit tests for each CTI concept action.
- Browser tests for Strengthen and at least one counterexample-producing action
  with a View flow.

### 28. Repair Save Invariant Format

Findings: 29

Track original conjectures, dropped conjectures, kept conjectures, and newly
added conjectures so web Save Invariant emits the same sections as Python:
- original conjectures kept
- original conjectures dropped, commented out
- new conjectures

Dependencies: steps 22, 26, and 27.

Verification:
- Golden-style unit test comparing saved invariant text to Python for kept,
  dropped, and new conjectures.

### 29. Repair Event Pattern Semantics

Findings: 34

Before building UI, make Go event filter/find semantics conform to Python's
event-pattern logic rather than substring matching. Port enough pattern parsing
and matching to make Filter, Find reverse, and Find forward meaningful.

Dependencies: step 3 if pattern entry dialogs are included in tests; otherwise
none.

Verification:
- Unit tests comparing pattern matching/filter/find to Python examples.

### 30. Port The Event Trace Viewer Browser UI

Findings: 33

Build the web equivalent of `ivy_ev_viewer.py`:
- Event tree with lazy open/close.
- Event notebook sheets.
- Filter.
- Find reverse/forward.
- Pattern list.
- Pattern save/load/clear.

Dependencies: steps 3 and 29.

Verification:
- Browser tests for filter creating a new sheet and find selecting/uncovering an
  event.

### 31. Repair Concept Graph Layout/Ordering

Findings: 35

After graph data and checkboxes are faithful, align layout behavior:
- Node sorting with enabled positive transitive relations.
- Subgraph/grouping behavior.
- Transitive reduction use in production rendering.
- DOT/Dagre differences documented or minimized.

Dependencies: steps 6, 20, and 23.

Verification:
- Rendering/unit tests for transitive sorting/reduction.
- Browser screenshot/pixel or graph-json tests for stable layout features.

### 32. Implement Analysis State Save/Load

Findings: 3

Once the sheet/session/graph model is stable, implement faithful analysis-state
save/load:
- Preserve ARGs.
- Preserve concept graphs/domains.
- Preserve tabs/sheets.
- Preserve checkbox state and selected state/facts.
- Decide whether to support Python `.a2g` directly or provide a web JSON format
  with an explicit non-equivalence note.

Dependencies: steps 6, 9, 10, 20, 22, and 28.

Verification:
- Round-trip test saving and loading a multi-sheet session.
- Browser test reloads a saved session and verifies graph/toggle/tab state.

## Suggested Milestones

### Milestone A: Web Substrate Is Trustworthy

Steps: 0-8

Result: the browser has the right defaults, reliable error reporting, Python-like
dialogs, backend-supplied menus, correct selected-state routing, backend-owned
checkboxes, selectable facts, and working source browsing.

### Milestone B: ARG Workflows Are Faithful Enough To Build On

Steps: 9-14

Result: sheets are real units of interaction, Step in opens a functional
analysis UI, ARG execute actions work, choice-backed ARG commands work, safety
trace viewing works, and reachable states are available.

### Milestone C: Generic Concept Graph Workflows Are Faithful

Steps: 15-21

Result: concept graph materialization, selected-edge materialization,
projections, add relation, backtrack, concrete/gather/reach, and DOT export
follow Python behavior.

### Milestone D: CTI/Invariants Are Faithful

Steps: 22-28

Result: CTI starts correctly, used-relation display is precise, diagram/BMC/
weaken/strengthen/minimize/sufficient/inductive actions are wired, and invariant
save format matches Python.

### Milestone E: Secondary UI Surfaces

Steps: 29-32

Result: event trace viewer behavior is ported, concept graph layout is closer to
Tcl/Tk/dot behavior, and complete analysis-state save/load is available.

