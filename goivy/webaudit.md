# Web UI Audit Against Python Tcl/Tk UI

This audit compares the Python Tcl/Tk GUI behavior in `pyivy/ivy/ivy/ivy_ui.py`,
`ivy_graph_ui.py`, `ivy_ui_cti.py`, `tk_ui.py`, `tk_graph_ui.py`,
`ivy_ui_util.py`, and `ivy_ev_viewer.py` against the Go/web UI in
`goivy/webui` and `goivy/webui/static`.

Scope: this is a source audit only. It intentionally records confirmed gaps,
bugs, and behavior divergences; it does not attempt fixes.

## Source Map

The Python README describes the intended GUI layers: `ivy_graph_ui.py` owns
toolkit-independent concept graph operations, `ivy_ui.py` owns the top-level
ARG UI, `tk_graph_ui.py` and `tk_ui.py` are the Tcl/Tk backends, and the main
UI contains a collection of tabs each with an ARG and optional concept graph
(`pyivy/ivy/ivy/README.md:94-121`).

The web UI has a static HTML/JS shell in `goivy/webui/static/index.html`,
`ivyweb_app.js`, `ivyweb_api.js`, and `ivyweb_controls.js`; a Go HTTP/API layer
in `webui_handlers.go` and `webui_backend.go`; session behavior in
`webui_session.go`; and several Python-inspired Go UI structs such as
`webui_ui_main.go`, `webui_graph_widget.go`, `webui_ui_cti.go`, and
`webui_ui_evviewer.go`.

## Findings

### 1. Menu Specs Are Not A Single Source Of Truth

Python builds menus from toolkit-independent menu specs. The top-level ARG UI,
generic concept graph UI, and CTI concept graph UI each define behavior-bearing
menu structures (`ivy_ui.py:34-50`, `ivy_graph_ui.py:34-53`,
`ivy_ui_cti.py:36-47`, `ivy_ui_cti.py:367-381`).

The web UI hardcodes the visible menus in HTML and binds individual ids in JS
(`static/index.html:34-137`, `static/js/ivyweb_app.js:544-567`). Go also has
menu-spec ports (`webui_graph_widget.go:71-101`, `webui_ui_cti.go:70-90`,
`webui_ui_cti.go:582-597`), but the browser does not render those Go specs.

Impact: adding or correcting a Python-style menu in Go does not make it
available in the browser. This is already visible for CTI concept-graph actions
such as Minimize, Check sufficient, Check relative induction, and Strengthen,
which exist in Python and Go structs but not in the static browser menu.

### 2. Default Mode Is Not Faithful

Python's default mode is `pdr`, and startup alpha is chosen from the active mode
(`ivy_ui.py:27`, `ivy_ui.py:62-73`). Go also records `DefaultMode = ModePDR`
(`webui_ui_main.go:33-34`).

The browser select defaults to the first option, `induction`, and `runCheck`
reads the DOM value directly (`static/index.html:53-58`,
`static/js/ivyweb_app.js:2504-2517`). The `/check` handler also defaults to
`induction` if JSON decode fails (`webui_handlers.go:403-410`).

Impact: a fresh web UI check can run induction where Python and the Go UI model
would default to PDR. More importantly, any initialization behavior that depends
on mode can diverge immediately.

### 3. Analysis State Save/Load Is Missing

Python `File -> Save` saves the current analysis graph, and possibly concept
graph state, as a pickle `.a2g` (`ivy_ui.py:76-82`). This is not just source
text; it preserves interactive ARG/domain state.

The web File menu exposes Load, Save as, Download current model, and Save
Invariant (`static/index.html:34-48`). The API save endpoint returns
`ivy_session.json` (`webui_handlers.go:380-393`), and `Session.SaveState`
serializes only session id, file path/content, and toggles
(`webui_session.go:970-979`).

Impact: the web UI cannot faithfully persist or reload the Python GUI's analysis
state. Current save behavior is a browser/source/session convenience, not a
port of Python's `.a2g` feature.

### 4. Tab Semantics Are Only Partially Ported

Python `TkUI.add` creates a real notebook tab, wires a fresh ARG widget, restores
any `art.state_graphs`, raises the page, and starts the UI
(`tk_ui.py:79-111`). `remove` deletes the tab and quits when the last tab is
gone (`tk_ui.py:113-117`).

The web starts with a single `Sheet 1` (`static/index.html:78-85`). JS can clone
the sheet DOM for "Step in" (`static/js/ivyweb_app.js:908-955`), but `Sheet 1`
cannot be removed (`static/js/ivyweb_app.js:957-976`), and the application still
has single `this.argGraph` and `this.conceptGraph` fields for the main graph
objects.

Impact: browser sheets are not equivalent to Python analysis tabs. A decomposed
sub-ARG can be displayed, but it is not a full independent `AnalysisGraphUI`
with its own state graph, concept graph, menus, and lifecycle.

### 5. Source Browsing Is Broken/Inert

Python "View Source" extracts an action source location and opens a reusable
file browser highlighting that line (`ivy_ui.py:289-295`,
`tk_ui.py:119-124`, `ivy_ui_util.py:48-72`).

The web JS expects `result.source` before it updates and scrolls the editor
(`static/js/ivyweb_app.js:1659-1669`). The server-side ARG action only returns
`file` and `lineno`, not source text (`webui_session.go:859-878`).

Impact: "View Source" can complete successfully without showing source or
highlighting the line. It is not faithful to the Python file browser behavior.

### 6. Python Dialog Primitives Are Mostly Missing

Python has reusable modal primitives for listbox selection, multi-selection,
text dialogs with custom OK labels and optional Cancel, entry prompts, integer
prompts, OK/Cancel dialogs, and arbitrary button lists
(`ivy_ui_util.py:120-229`, `tk_ui.py:127-181`).

The web has one text dialog with a single OK button
(`static/index.html:219-227`, `static/js/ivyweb_app.js:2887-2921`), browser
file dialogs, and a raw `prompt()` for Add relation
(`static/js/ivyweb_app.js:3195-3205`).

Impact: workflows that depend on choosing a conjecture, choosing a remembered
goal, choosing relations for edge materialization, entering a BMC bound, or
viewing a result with a custom "View" button cannot behave like Python until
the missing dialog types exist.

### 7. Run Context And Error Reporting Are Not Faithful

Python wraps slow/error-prone GUI actions in `run_context`, which sets a busy
cursor and converts `IvyError` into a modal dialog (`ivy_ui_util.py:239-257`,
`tk_ui.py:183-199`).

The web mostly writes status strings. Worse, unknown generic actions are
accepted and reported as "not yet wired to engine" rather than failing
(`webui_session.go:656-699`).

Impact: a broken or unported browser command can look successful. Python's modal
error boundary is part of the interactive semantics because many actions depend
on it for user-visible failures.

### 8. Selecting An ARG Node Does Not Select The Backend State

Python `view_state` sets the concept graph's parent state to the clicked ARG
node, recomputing or creating the corresponding graph (`ivy_ui.py:124-134`).

The web sends `GET /concept?node=...` on ARG node click
(`static/js/ivyweb_app.js:1483-1508`, `static/js/ivyweb_api.js:90-100`), but
the HTTP handler ignores the `node` query and calls `GetConcept(sessionID)` with
no node (`webui_handlers.go:125-137`). `GoBackend.GetConcept` renders
`sess.SimpleSess` directly (`webui_backend_go.go:165-253`).

Impact: clicking ARG node 1 can update the browser label while the backend
concept graph is still the session's current simple concept graph. This is a
real behavioral bug relative to Python.

### 9. ARG Node Execute-Action Menu Is Missing In The Browser

Python right-clicking an ARG node shows "Execute action:" followed by sorted
state-action commands (`ivy_ui.py:100-108`, `ivy_ui.py:159-162`,
`ivy_ui.py:227-245`).

Go has a port of `NodeExecuteCommands` (`webui_ui_main.go:270-315`), but
`RenderARG` adds nodes with `nil` actions (`goivy/art_cyrender.go:210-224`).
The browser fallback menu includes only fixed node commands, not execute-action
entries (`static/js/ivyweb_app.js:1516-1565`).

Impact: the web UI cannot execute available state actions from a node the way
the Tcl/Tk UI can.

### 10. ARG Commands That Need User Choices Are Called Without Arguments

Python `Try conjecture` opens a listbox of undecided conjectures when no
conjecture is passed, browses the source for a chosen conjecture, and then
continues according to mode (`ivy_ui.py:395-418`). Python `Try remembered goal`
opens a list of remembered graph names when no goal is passed
(`ivy_ui.py:419-435`).

The web JS sends only the node and action id (`static/js/ivyweb_app.js:1576-1581`).
The Go session expects `args["conjecture"]` or `args["goal"]`
(`webui_session.go:799-814`).

Impact: these menu entries cannot reproduce the Python UI. They either fail,
no-op, or operate with empty strings.

### 11. Show Reachable States Is Missing

Python top-level Action menu includes "Show reachable states", and it opens a
reachable-state tree as a new tab (`ivy_ui.py:48-50`, `ivy_ui.py:249-313`).

The web top menu has no corresponding action (`static/index.html:51-70`), and
the available "Reach"/"Path reach" concept menu actions are different
operations (`static/index.html:113-130`, `static/js/ivyweb_app.js:3092-3115`).

Impact: a Python reachability inspection workflow has no browser equivalent.

### 12. Safety-Check Result Dialogs And Trace Viewing Are Missing

Python bounded safety failure asks "View error trace?" and can add the trace ARG
as a tab (`ivy_ui.py:341-350`, `ivy_ui.py:377-380`). Python local safety can
offer "View unsafe states" and "View concrete trace" buttons
(`ivy_ui.py:357-367`).

The web ARG safety action returns fields/status only
(`webui_session.go:751-759`), and the browser check display writes to the
Details panel without Python-style View buttons (`static/js/ivyweb_app.js:2582-2600`).

Impact: counterexample navigation is substantially weaker than the Tcl/Tk UI.

### 13. Step-In/Decompose Creates A Display, Not A Full Analysis UI

Python `decompose_edge` obtains a sub-ART and calls `ui_parent.add` with
`AnalysisGraphUI`, producing a real new tab (`ivy_ui.py:269-277`).

The web decompose action creates a cloned sheet, then instantiates a local
`IvyGraph` for the sub-ARG container (`static/js/ivyweb_app.js:1632-1655`).
That local graph is not registered as the app's active ARG UI.

Impact: the user can see a decomposed graph, but follow-on Python behaviors
such as node actions, concept graph viewing, source browsing, and tab state are
not faithfully available on that sub-ARG.

### 14. ARG Edge Label Brace Post-Processing Is Missing

Go/Python ARG labels encode braces as `-[` and `]-` for dot safety
(`goivy/art.go:1389-1390`). Python Tk rewrites those back to `{` and `}` after
rendering (`tk_ui.py:262-268`).

The web `RenderARG` passes transition labels through directly
(`goivy/art_cyrender.go:227-241`), and no corresponding web rewrite was found.

Impact: ARG edge labels can show internal escaping in the browser where Tcl/Tk
shows user-facing braces.

### 15. Concept Graph Click Semantics Differ

Python concept graph node and edge action menus are left-click menus; right
click intentionally has no actions (`ivy_graph_ui.py:126-138`,
`ivy_graph_ui.py:164-173`). CTI concept graphs also use left-click selection
actions (`ivy_ui_cti.py:384-397`).

The web uses left click to toggle selection and right click for context menus
(`static/js/ivyweb_app.js:624-680`, `static/js/ivyweb_app.js:1682-1819`).

Impact: this may be an acceptable web UI convention, but it is a direct
interaction divergence. It matters because Python distinguishes "select" from
"open the action menu" in later workflows.

### 16. Concept Edge Materialization API Drops The Essential Parameters

Python materializes an edge tuple `(relation, source, target)` and a truth value;
negative materialization calls the same path with `truth=False`
(`ivy_graph_ui.py:488-529`). Go `GraphWidget` has the matching shape:
`MaterializeEdge(relID, headID, tailID, truth)` and `DematerializeEdge`
(`webui_graph_widget.go:265-279`).

The browser API sends `{concept, type: "edge", positive}` for edge
materialization (`static/js/ivyweb_api.js:180-199`), but the HTTP handler decodes
only `Concept` (`webui_handlers.go:217-230`). `GoBackend.ConceptMaterialize`
then always calls node/simple materialization and has no relation/source/target
or truth parameter (`webui_backend_go.go:343-354`).

Impact: positive edge materialization, negative edge materialization, and
dematerialization cannot be faithful through the current browser API.

### 17. "Materialize Edge From Selected" Is Missing

Python node menu includes "Materialize edge"; it uses the previously selected
node, filters binary relations by source/target sorts, prompts with a listbox,
and materializes the chosen relation (`ivy_graph_ui.py:126-136`,
`ivy_graph_ui.py:488-507`).

The Go `GraphWidget.GetNodeActions` includes an action placeholder
(`webui_graph_widget.go:386-400`), but the browser fallback node menu omits it
(`static/js/ivyweb_app.js:1707-1753`). The dialog infrastructure needed to
choose the relation is also absent.

Impact: a core concept-domain construction workflow is unavailable in the web
UI.

### 18. Projection Actions Are Stubbed/Missing

Python adds projection actions from `g.get_projections(node)` to the node menu
(`ivy_graph_ui.py:152-162`).

Go `GraphWidget.GetNodeProjectionActions` is an explicit stub returning nil
(`webui_graph_widget.go:427-431`). The browser can call `addProjection` only if
an action somehow reaches it (`static/js/ivyweb_app.js:1840-1843`,
`static/js/ivyweb_app.js:1946-1954`).

Impact: projection discovery is not ported. The browser lacks the Python
"Add projection..." behavior.

### 19. Checkbox State Is Not Faithful Across Defaults And Undo/Redo

Python concept graph checkboxes are stored in `rel_enabled`, copied into the
graph before rendering, and copied back after undo/redo
(`tk_graph_ui.py:62-87`). Python also enables default labels' `+` checkbox
(`tk_graph_ui.py:76-79`).

The web builds relation rows client-side (`static/js/ivyweb_app.js:1187-1250`)
and tracks visibility in JS (`static/js/ivyweb_app.js:1268-1305`). The server
API only stores a single edge/display/value per call (`webui_handlers.go:254-263`),
and the main client-side defaults initialize visibility maps to false
(`static/js/ivyweb_app.js:18-25`).

Impact: display checkbox state can diverge from the backend graph state, default
labels are not guaranteed to match Python, and undo/redo does not restore
checkboxes the way Python `reverse_sync_checkboxes` does.

### 20. Constraint/Facts Text Selection Is Missing

Python renders graph constraints below the concept graph, and each constraint
can be clicked to toggle whether it is an active fact. `get_active_facts` then
returns only selected constraints (`tk_graph_ui.py:168-218`).

The web has a generic Details panel (`static/index.html:144-148`). Go
`GraphWidget.GetActiveFacts` and `HighlightSelectedFacts` are stubs
(`webui_graph_widget.go:503-515`).

Impact: CTI workflows that build conjectures from selected facts cannot match
Python from the browser.

### 21. Concept Graph Export Does Not Produce A DOT File

Python concept graph Export asks for a `.dot` path and writes the displayed Tcl
graph (`tk_graph_ui.py:254-258`).

The web "Export" menu calls generic action `export`
(`static/js/ivyweb_app.js:3181-3188`). The backend action only emits an event
with the number of facts and returns no file/content (`webui_session.go:572-577`).

Impact: Export is currently not a port of Python's DOT export behavior.

### 22. Remembered Goals Are Not Usable Faithfully

Python `GraphWidget.remember` prompts for a user-entered name and stores a copy
of the graph in the parent; `try_remembered_graph` prompts from those names and
restores the chosen goal (`ivy_graph_ui.py:331-339`, `ivy_ui.py:419-443`).

The web `rememberGraph` sends no name (`static/js/ivyweb_app.js:3167-3175`).
The backend stores a fixed domain name `"remembered"` (`webui_session.go:346-350`),
while `try_remembered` expects an explicit `goal` argument
(`webui_session.go:807-814`).

Impact: multiple named remembered goals, and even a reliable single remembered
goal recall path, are not faithfully implemented.

### 23. Add Relation Is A Shallow Port

Python `add_concept_from_string` parses the user's text through
`g.string_to_concept` and then adds the resulting concept relation
(`ivy_graph_ui.py:320-328`).

The web uses `prompt()` and the backend only calls `ToFormula` as a parse check,
then inserts a simple `Concept{Name, Formula}` without deriving the full concept
metadata (`static/js/ivyweb_app.js:3191-3205`,
`webui_session.go:600-621`).

Impact: relations added in the web UI may render but can lack the sort,
variable, and concept-domain behavior Python obtains from `string_to_concept`.

### 24. Backtrack Is Wired To A Single Undo In The Browser Path

Python backtrack undoes until the most recent graph marked with
`backtrack_point` (`ivy_graph_ui.py:194-202`). Go `GraphWidget.Backtrack`
contains the same loop (`webui_graph_widget.go:128-138`).

The browser menu calls generic action `backtrack`, and `Session.ExecuteAction`
maps that to a single `ConceptSess.Undo()` (`webui_session.go:332-335`).

Impact: browser Backtrack is not the Python backtrack operation.

### 25. Concrete/Gather/Reach Operations Use Simplified Session Paths

Python `GraphWidget.concrete` modifies the current concept graph state with
`g.state + g.concrete` (`ivy_graph_ui.py:204-210`). Python `gather` computes
facts from displayed relation values (`ivy_graph_ui.py:231-238`). Python `reach`
shows success/failure dialogs and can let the user view the reached state
(`ivy_graph_ui.py:373-389`).

The web generic actions route through `Session.ExecuteAction`. `concrete`
computes a Z3 model from the last ARG state (`webui_session.go:381-425`),
`gather` calls `ConceptSess.GetFacts(nil)` without displayed relation values
(`webui_session.go:327-331`), and `reach/path_reach` return status/result data
without the Python dialog flow (`webui_session.go:461-495`,
`static/js/ivyweb_app.js:3043-3115`).

Impact: these action names exist in the web menu, but they are not necessarily
the same operations as Python's concept graph UI methods.

### 26. CTI Top-Level Start Is Not Ported Into Session Load

Python CTI startup calls the ARG start routine, initializes transitive tracking,
loads `im.module.conjs`, views state 0, autodetects transitive relations, and
records whether the current graph is a CTI (`ivy_ui_cti.py:49-62`).

`Session.LoadFileContent` builds a generic `AnalysisGraphUI` and simple concept
session, but does not instantiate/start `CTIAnalysisGraphUI`
(`webui_session.go:64-230`). Go has `StartCTI`, including transitive detection
(`webui_ui_cti.go:93-179`), but it is not part of the browser load path.

Impact: CTI-specific initialization can be skipped even though the browser shows
Invariant/CTI-style controls.

### 27. CTI Bounded Check Has No Bound Prompt And Uses A Fixed Bound

Python CTI bounded check prompts for an integer bound when none is supplied and
remembers the current bound (`ivy_ui_cti.py:280-294`).

The web "Bounded check" calls `runCheck('bounded')` with no bound
(`static/js/ivyweb_app.js:2924-2934`). Backend bounded mode uses `nSteps := 10`
(`webui_session.go:1178-1186`).

Impact: users cannot reproduce Python's selected bound behavior from the web UI.

### 28. CTI Weaken Menu Is Currently Broken

Python Weaken opens a multi-select listbox of conjectures, removes the selected
ones, clears `have_cti`, and shows a text dialog listing removals
(`ivy_ui_cti.py:235-247`).

The web calls `executeAction('weaken', {})` with no indices
(`static/js/ivyweb_app.js:2937-2950`). The backend requires
`args["indices"]` and errors if none are provided (`webui_session.go:497-507`).

Impact: the visible Weaken command cannot work like Python.

### 29. Save Invariant Format Is Not Faithful

Python `save_conjectures` preserves original conjectures kept, comments out
original conjectures dropped, and separates new conjectures
(`ivy_ui_cti.py:249-278`).

The web `saveInvariant` asks for `get_conjectures`, falls back to `gather`, or
copies source invariant/conjecture declarations (`static/js/ivyweb_app.js:2096-2148`).
The Go session `save_abstraction` emits concepts and current conjectures only
(`webui_session.go:536-570`), and `CTIAnalysisGraphUI.SaveConjectures` emits a
simple current-conjecture file (`webui_ui_cti.go:468-483`).

Impact: saved invariant files cannot be mechanically compared to Python's
kept/dropped/new output.

### 30. Used-Relation Display After CTI Failure Is Over-Approximated

Python `show_used_relations` clears edges, enables `+` only for relations whose
formulas mention symbols used in the failing clauses, handles three-argument
numeral apps by creating concepts, and optionally enables `-`
(`ivy_ui_cti.py:182-205`).

The web induction check returns every boolean relation symbol as `used_relations`
(`webui_session.go:1132-1146`), and JS auto-checks `+` for all returned names
(`static/js/ivyweb_app.js:2533-2577`).

Impact: the CTI graph can show too many relations, hiding the critical
diagnostic signal Python tries to preserve.

### 31. CTI Diagram Uses A Different Browser Path

Python CTI Diagram checks inductiveness if needed, computes a diagram model,
calls `view_state` with the diagram, calls `show_used_relations(..., both=True)`,
and gathers facts (`ivy_ui_cti.py:207-227`).

Go `CTIAnalysisGraphUI.Diagram` has a closer backend implementation
(`webui_ui_cti.go:376-443`), but the browser "Diagram Domain" path calls
`/concept/diagram`, which invokes `sess.SimpleSess.Diagram()` instead
(`static/js/ivyweb_app.js:2646-2658`, `webui_backend_go.go:397-405`).

Impact: the browser diagram action is not the Python CTI diagram behavior.

### 32. CTI Concept-Graph Actions Are Not Browser-Wired

Python CTI concept graph menus include Gather, Bounded check, Minimize, Check
sufficient, Check relative induction, Strengthen, and Export
(`ivy_ui_cti.py:367-381`). Python implements these through active facts and
counterexample dialogs (`ivy_ui_cti.py:447-610`, `ivy_ui_cti.py:675-721`).

Go has a `CTIConceptGraphWidget` with matching method names
(`webui_ui_cti.go:600-855` and following), but the static browser menu only has
the generic concept menu (`static/index.html:113-137`) and lacks bindings for
Minimize, Check sufficient, Check relative induction, and Strengthen.

Impact: major CTI refinement workflow operations are not reachable from the web
UI.

### 33. Event Trace Viewer Is Not Ported To The Browser

Python `ivy_ev_viewer.py` provides a Tcl/Tk event tree with Filter, Find
reverse/forward, event sheets, a pattern list, pattern save/load, and lazy tree
expansion (`ivy_ev_viewer.py:35-229`).

Go has an `EventTraceViewer` data/model port (`webui_ui_evviewer.go:1-272`),
but no browser DOM, JS menu, route, or rendering integration was found. Static
web references to "events" are SSE infrastructure, not the Python event trace
viewer.

Impact: users cannot inspect/filter/find event traces in the web UI as they can
in the Tcl/Tk tool.

### 34. Event Pattern Semantics Are Simplified In Go

Python event viewer filtering and finding delegates to event-pattern logic
(`ivy_ev_viewer.py:52-87`) and supports pattern list operations/save/load
(`ivy_ev_viewer.py:113-175`).

Go `FilterEvents` and `FindEvent` perform string containment over flattened
events (`webui_ui_evviewer.go:149-188`), while pattern storage is newline string
save/load (`webui_ui_evviewer.go:236-271`).

Impact: even before browser wiring, the Go event viewer model is not a faithful
behavioral port of Python's pattern-based viewer.

### 35. Concept Graph Layout/Ordering Is Not Faithful

Python concept graph display is driven by `IvyGraph`, Cytoscape elements, and
Graphviz dot layout (`README.md:82-89`). `GraphWidget.sort_nodes` uses enabled
positive transitive relations to topologically sort nodes for display, and
`make_subgraphs` returns true (`ivy_graph_ui.py:59-78`, `ivy_graph_ui.py:116-119`).

The web uses Cytoscape/Dagre in the browser (`static/index.html:24-28`), and
`RenderConceptGraph` walks the current domain nodes/edges without the Python
transitive sorting/subgraph grouping behavior (`webui_cyrender.go:89-218`).
`GetTransitiveReduction` exists only in tests and is not called by production
web rendering (`webui_graph_model.go:725-771`).

Impact: graph shape, ordering, grouping, and transitive-reduction display can
diverge even when the underlying abstract value is the same.

### 36. Web UI Has Browser-Only File Editing Semantics

The web adds CodeMirror editing, file handles, dirty-close handling, external
file change detection, browser download/save-as behavior, and tutorial iframe
navigation (`static/index.html:180-209`, `static/index.html:231-255`,
`static/js/ivyweb_app.js:160-430`).

Python Tcl/Tk has a source browser and save dialogs, but not an always-live
source editor with browser-local persistence (`tk_ui.py:119-181`).

Impact: this is not necessarily a bug, but it is a deliberate divergence. Tests
for Python GUI conformance should distinguish these web-only behaviors from
backend verification/UI parity.

## High-Risk Summary

The most concerning confirmed backend/access divergences are:

1. `GET /concept?node=...` ignores the node parameter, so ARG node selection is
   not faithful.
2. Concept edge materialization drops relation/source/target/truth and routes to
   node/simple materialization.
3. Browser menus call many Python-named actions through simplified generic
   `Session.ExecuteAction` paths instead of the closer Go `GraphWidget` or
   `CTIAnalysisGraphUI` methods.
4. CTI workflows need Python-style dialogs and active-fact selection before
   Weaken, BMC bound selection, Strengthen, Minimize, Check sufficient, and Check
   relative induction can be faithful.
5. Save/load/export behavior is not currently a faithful port of `.a2g`,
   invariant kept/dropped/new output, or DOT export.

