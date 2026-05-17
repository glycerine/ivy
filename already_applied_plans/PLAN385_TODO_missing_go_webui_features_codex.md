# PLAN385 TODO: Missing Go/Web UI Features vs Python Tcl GUI

Audit date: 2026-05-13.

Sources subtracted:

- `~/ivy/already_applied_plans/PLAN378_cc_python_gui_inventory.md`
- `~/ivy/already_applied_plans/PLAN383_codex_tcl_gui_inventory.md`
- Current implementation under `~/ivy/goivy/webui`, especially `DATA_MODEL.md`, `webui_session.go`, `webui_ui_main.go`, `webui_ui_cti.go`, `webui_graph_widget.go`, `webui_cyrender.go`, `static/index.html`, and `frontend/src/services/*`.

Scope note: this is not a request to recreate Tk pixel-for-pixel. A behavior is counted as present when the Go/web UI exposes the same user workflow and state effect, even if the toolkit is different. A behavior is counted as missing or partial when the current port either has no workflow, has backend-only code with no reachable UI, has a web-only substitute that cannot load or save Python GUI state, or has a visible behavior that is materially weaker than the Python Tcl GUI.

## Mostly Subtracted

The current Go/web UI already covers a large core: it can load Ivy source into a hosted session, display ARG and concept graphs in Cytoscape, maintain multiple analysis/event sheets, select ARG states, open concept graphs for states, run top-level checks, expose many ARG node and edge context actions, show concept checkboxes for `+`, `?`, `-`, and `T`, gather/select facts, perform several concept graph operations, run CTI/invariant actions, save/export several text artifacts, and load/filter/search event traces. Those covered areas are not repeated below except where the port is incomplete or the behavior differs enough to matter for tests.

## P0: Compatibility And State Ownership

### 1. Python `.a2g` Analysis-State Load/Save Compatibility

Inventory refs: PLAN383 items 3 and 34; PLAN378 sections 5.1, 27.1, and 27.2.

The Python GUI loads `.a2g` analysis-state files before source files and saves analysis graphs as pickle protocol 2. The Go/web UI instead serializes `analysis_state_format: "ivyweb-json"` and explicitly marks `python_a2g_equivalent: false`, so Python GUI state is not round-trippable.

TODO: decide whether to implement native `.a2g` compatibility, a one-way importer/exporter, or a documented non-goal. Tests should use a Python-created `.a2g` with multiple ARG sheets, covers, current tab, and concept state, then verify the web UI reconstructs equivalent sheets or rejects the file with an explicit compatibility message.

### 2. Backend Session Save Is Not Wired To Python File Menu Semantics

Inventory refs: PLAN383 items 19, 34, and 35; PLAN378 sections 5.1, 5.2, and 22.11.

The current web UI has browser-side "Save Analysis State" and backend `SaveState`, but File menu descriptors also include actions such as `save` and `save_abstraction` that do not all route through the Python-shaped menu semantics. Some static buttons work, while descriptor-driven menus can dispatch unsupported action names.

TODO: make File menu behavior explicit and complete: save current model, save analysis state, save abstraction, remove tab, and exit/close should each map to one controller command or be removed from the menu. Tests should click every File menu item exposed by the browser shell and assert a successful state effect or a deliberate disabled state, not an "unknown action" backend error.

### 3. Mode Selection Is Not Backend-Authoritative Everywhere

Inventory refs: PLAN383 item 8; PLAN378 sections 6.1 through 6.7 and 12.1.

The browser has a mode `<select>` and top-level checks pass a mode to `/check`, but ARG node actions such as node safety still use the backend `AnalysisGraphUI.Mode`, which is not synchronized by the browser mode control. Python's radio menu stores mode in the UI object and uses it to choose `init_alpha()` and safety-check dispatch.

TODO: add an explicit mode command or include mode in all relevant ARG actions so backend behavior always matches the visible mode. Tests should set each mode in the UI, run node safety and extension-related actions, and verify the backend branch and resulting graph/state match the selected mode.

### 4. FIXED. Typed `UIDataModel` Is Present But Not The Driver For Most Behaviors

Inventory refs: DATA_MODEL.md and PLAN383 items 83 through 86.

The TypeScript data model preserves typed ARG, CTI, concept, graph-stack, and fact-selection payloads, but most UI behavior still updates Cytoscape and DOM directly from service results. That is acceptable for a controller-owned design, but tests that want to avoid full browser rendering need stable model-level entry points for each behavior.

TODO: for each newly ported behavior, define whether the test target is backend state, `UIDataModel`, direct controller state, or DOM. Tests should be able to validate graph snapshots, selected ARG state, active facts, toggles, and sheets without needing Playwright unless the behavior is specifically visual.

## P0: Launch, Entry Points, And Configuration

### 5. Python Launch Parameters Are Not Fully Represented

Inventory refs: PLAN383 items 1, 2, 8, 91, 92, and 96; PLAN378 sections 1.1, 1.6, 27.1, 27.3, and 33.18.

The Python GUI startup path honors UI module selection, initial mode, diagnose mode, no-UI mode, target settings, extension settings, and macOS dynamic-library path adjustments. The web UI starts a hosted session and exposes some runtime controls, but it does not provide the same launch-time GUI configuration contract.

TODO: enumerate which Python CLI/UI parameters should have web equivalents and which are obsolete. Tests should start `ivyweb` with representative flags, load a model, and verify the first rendered session has the requested mode, diagnostic behavior, extension action set, and compile settings.

### 6. Diagnostic GUI Entry Points Are Missing

Inventory refs: PLAN383 items 91 and 92; PLAN378 sections 27.1 through 27.4.

Python diagnostic flows such as `ivy_check --diagnose`, `show_counterexample()`, and `try_property()` open GUI views for counterexamples or false properties. The Go/web UI has check results and can show a trace ARG for some failures, but there is no complete diagnostic entrypoint parity.

TODO: implement or explicitly retire web equivalents for diagnose-mode counterexample display and try-property workflows. Tests should run a failing property through the intended web diagnostic entrypoint and verify a CTI or trace ARG sheet opens with the failed property and source context.

### 7. Notebook, `ivy2.py`, `ivy_launch.py`, And Distributed Process GUI Behaviors Are Unclassified

Inventory refs: PLAN383 items 93 through 96; PLAN378 sections 28.1, 28.2, 33.16, and 33.17.

The Python GUI inventory includes notebook generation, generated JavaScript widget integration, terminal process launch, and sequential port allocation. The current web UI does not expose these workflows as part of `webui`.

TODO: decide which of these are in scope for the Go/web GUI port. If any are in scope, add web equivalents and tests; if not, record them as intentionally excluded so future audits do not keep rediscovering them.

## P1: Menus, Commands, And Dialogs

[ ] ### 8. Descriptor Menus Are Static And Not Session/Mode Aware

Inventory refs: PLAN383 items 9, 37, 64, and 72; PLAN378 sections 4, 20.1, and 20.12 through 20.20.

The backend exposes browser menu descriptors, but `BuildBrowserMenuDescriptors()` constructs them from fresh default `AnalysisGraphUI` and `GraphWidget` instances rather than the active session, CTI UI, or active sheet. The static HTML has additional Invariant and Conjecture menus, so there are two overlapping menu systems with different coverage.

TODO: collapse menu ownership into one active-session menu model or make the static menus the only source of truth. Tests should load a model, enter CTI and non-CTI contexts, switch sheets, and assert that only valid current actions are enabled and dispatchable.

[ ] ### 9. Some Menu Items Dispatch Unsupported Backend Actions

Inventory refs: PLAN383 items 9, 31, 33, 34, and 35; PLAN378 sections 5 and 7.

The descriptor menu includes actions such as `save`, `remove_tab`, `exit`, mode actions, and `recalculate_all`; the generic menu dispatcher sends them to `runAction`, but `Session.ExecuteAction` does not implement all of those action names. Related functionality exists elsewhere in the controller, so the gap is wiring and command naming.

TODO: create a command map for every descriptor/static menu item and add regression tests that no visible menu item produces an unknown-action response. Where a menu action is intentionally browser-only, it should dispatch to the controller rather than the backend action endpoint.

### 10. Dialog Pre-Seeding For Tests Is Missing

Inventory refs: PLAN383 item 11; PLAN378 sections 22.2 and 28.9 through 28.10.

The Python GUI can pre-seed dialog answers to make workflows testable without manual input. The web controller has promise-based dialogs, but no shared test-answer queue or deterministic answer injection mechanism.

TODO: add a small dialog-answer harness owned by the controller or command registry. Tests should pre-seed entry, integer, listbox, multiple-selection, and button-list dialogs, run real commands, and assert the command receives the injected result.

[ ] ### 11. Dialog Return Semantics Differ From Tk In Several Places

Inventory refs: PLAN383 items 13 through 18; PLAN378 sections 22.3 through 22.10.

The web `listboxDialog` generally returns option values, while the Tk inventory distinguishes index-returning listboxes, multiple-selection indices, and button-list cancel behavior. Some callers work around this by using values that are already indices, but the dialog contract is not the same as Python's.

TODO: define and test dialog return semantics per dialog kind. Tests should cover single selection, multiple selection, cancel, Escape, Return, out-of-range integer input, and callers that need selected indices rather than displayed text.

[ ] ### 12. RunContext/Error Modal Behavior Is Only Partially Ported

Inventory refs: PLAN383 item 10; PLAN378 sections 22.13 through 22.15 and 33.1.

Python wraps long-running UI callbacks in a run context that shows blocking Ivy error dialogs and restores the cursor. The web UI usually writes errors to the status bar/toast and sometimes uses loading overlays, so important errors can be non-modal and inconsistent across commands.

TODO: define a web equivalent of `RunContext` for command execution, including modal-vs-status policy and busy/ready visual state. Tests should force backend errors in check, ARG action, concept action, CTI action, and event filtering paths and assert consistent user-visible error handling.

[ ] ### 13. Save-As Dialog Filters And Titles Are Incomplete

Inventory refs: PLAN383 item 19; PLAN378 section 22.11.

Python uses file dialogs with specific titles and filters for model files, abstractions, invariants, event patterns, and analysis-state files. The web UI uses File System Access or download fallbacks, but the filter/title coverage is uneven and not obviously tied to every export action.

TODO: inventory every save/open/export path and give it an explicit suggested name, MIME type, extension set, and cancel behavior. Tests should stub `showSaveFilePicker` and download fallback paths for `.ivy`, `.ivyweb.json`, invariant files, abstraction files, DOT exports, and event pattern files.

## P1: ARG Graph Behavior

[x] DONE. ### 14. ARG Safe-Node Coloring Is Missing From Rendered State

Inventory refs: PLAN383 items 22 and 66; PLAN378 sections 8.6, 12.3, and 32.1.

`AnalysisGraphUI.NodeColor()` knows about green safe nodes, but the render payload's `ARGNode` type has no safe flag and the Cytoscape style has no safe class. Node safety checks return a message but do not update the node's visible outline/fill status.

TODO: carry safety status through the ARG data model and Cytoscape element classes. Tests should run a safety check, inspect the backend ARG payload for safety metadata, and verify the rendered node class/style changes without relying only on a status message.

[x] DONE. ### 15. Marked ARG Node Visualization Is Missing

Inventory refs: PLAN383 item 26; PLAN378 section 11.

The backend can set `AnalysisGraphUI.Mark`, but the rendered ARG does not mark the node red and the frontend does not update the graph after a mark action unless an ARG payload is returned. Python fills the marked node red and preserves a single mark across rebuilds.

TODO: include marked-node state in the ARG payload and render it as a stable class. Tests should mark one node, rebuild/recalculate the graph, mark another node, and verify only the current marked node is visually marked and cover/join uses the same backend mark.

### 16. Safety Error Trace Viewing From Node Safety Is Missing

Inventory refs: PLAN378 sections 12.5, 13.4, 24, and 27.2.

Top-level checks can show trace actions, but node-level safety from the context menu only returns safe/message fields. Python can present a "View Error Trace" path when a safety or BMC check produces a counterexample.

TODO: return trace data or a trace ARG from node safety/BMC failures and add a visible "view trace" action. Tests should create an unsafe node, run node safety, click the trace action, and verify an event/ARG trace sheet opens with the counterexample.

### 17. BMC Entry Points Are Not Fully Exposed As ARG Workflows

Inventory refs: PLAN378 sections 13.1 through 13.4 and 16.3; PLAN383 items 71 and 92.

There is a top-level bounded check and backend methods for BMC-like operations, but the Python ARG workflow includes bounded checking from an ARG node, bound entry, unreachable/counterexample dialogs, and optional trace viewing. Those behaviors are only partially reachable in the browser.

TODO: add the missing ARG/CTI BMC commands and result dialogs, or clearly merge them with the existing bounded-check button. Tests should cover user-entered bounds, unreachable results, reachable counterexamples, and trace-sheet creation.

[ ] ### 18. Recalculate-State Behavior Is Missing Or Ambiguous

Inventory refs: PLAN383 item 32; PLAN378 sections 7.1, 10.2, and 16.4.

The port supports recalculating all transitions and recalculating an edge, but the Python inventory also calls out recalculate of the current state/concept graph state. The current UI does not make that state-level distinction clear.

TODO: identify the Python state-level recalculate behavior and add a matching command if it is distinct from edge/all recalculation. Tests should select a state, modify concept/domain information, run recalculate state, and verify the concept graph is updated without recalculating unrelated targets.

[ ] ### 19. `Extend` Closed-Node Reporting Is Too Weak

Inventory refs: PLAN383 item 30; PLAN378 sections 14.1 and 14.2.

The backend can attempt an extension and returns an error when the state is closed. Python shows an explicit closed-node dialog and otherwise executes the chosen extension and rebuilds the graph.

TODO: make closed-node and extension-success UX match the inventory. Tests should run Extend on a closed node and assert a visible closed-node message, then run it on an extendable node and assert the new state/edge appears.

[ ] ### 20. Try-Conjecture Source Browsing And Result Views Are Partial

Inventory refs: PLAN383 item 57; PLAN378 sections 16.1 through 16.4.

The web UI prompts for conjecture choices and calls the backend, but the Python workflow also browses conjecture source and chooses different result views depending on bounded, induction, or PDR mode. The current path mostly ends in status updates and backend effects.

TODO: complete the mode-specific try-conjecture flow, including source browsing and resulting concept/trace graph presentation. Tests should try a conjecture in bounded, induction, and PDR modes and assert the expected sheet/dialog/source behavior for each mode.

[ ] ### 21. Interpolant Refinement Workflow Is Missing

Inventory refs: PLAN378 sections 17.1 through 17.3.

Python supports refinement with interpolants, including a vacuous-pre-state message and adding predicates to the concept domain. No equivalent web workflow is visible in the current controller or menus.

TODO: port the interpolant/refinement command path or explicitly classify it as deferred. Tests should trigger a refinement case, accept an interpolant, and verify the new predicate appears in the concept relation controls and graph.

[ ] ### 22. ARG Source Browser Should Not Depend On The Main Editor

Inventory refs: PLAN383 item 20; PLAN378 section 23.

Python source browsing reuses a separate file browser window, loads a file, highlights a line red, scrolls to it, and raises the browser. The web UI currently loads returned source into the main editor and scrolls it, which can overwrite or disturb the editable model buffer.

TODO: add a source peek/browser panel or modal that is independent of the model editor. Tests should view source for an edge while the editor has unsaved changes and verify the edit buffer is preserved, the source line is highlighted, and repeated source views reuse/raise the same browser surface.

[ ] ### 23. Direct Context-Popup Shortcut For Single `<>` Action Is Not Preserved

Inventory refs: PLAN383 item 105; PLAN378 section 9.2.

The web UI handles ARG left-click as "view state", which covers the main user effect. Python's popup helper also has a special shortcut for a single `<>` action in context-action lists.

TODO: decide whether the shortcut matters in the web context-menu implementation. If it matters, test a one-action context list and verify the action executes directly without rendering a redundant popup.

## P1: Graph Rendering Fidelity

### FIXED (partly at least) by adding in graphviz. 24. DOT/Tk Layout Fidelity Is Replaced By Cytoscape Dagre

Inventory refs: PLAN383 item 43; PLAN378 sections 8.2, 26.1, 26.2, 26.5, and 26.7.

Python uses Graphviz DOT and maps DOT coordinates, dimensions, edge weights, and inverted Y coordinates into Tk canvas coordinates. The web UI uses Cytoscape dagre/grid and does not preserve DOT coordinates or node dimensions from Graphviz.

TODO: decide which DOT layout behaviors are required for port fidelity. Tests should compare layout invariants that matter to users, such as state ordering, cluster placement, back edge handling, and stable positions after recalculation.

[ ] ### 25. Cluster/Subgraph Box Rendering Is Missing

Inventory refs: PLAN378 sections 8.13, 26.4, and 104.

The Python canvas renderer draws cluster/subgraph rectangles when graph elements contain them. The current Cytoscape ARG/concept renderer does not expose or draw cluster boxes.

TODO: either render cluster/subgraph boxes in Cytoscape or record them as intentionally omitted. Tests should load a graph with subgraph/cluster metadata and assert the web graph displays the same grouping affordance.

[ ] ### 26. Back Edge Reversal And Pending Edge Constraints Are Missing

Inventory refs: PLAN378 sections 26.3 and 26.6.

Python graph rendering handles DOT-specific back edge reversal and pending edge constraints. The current Cytoscape pipeline has no matching representation.

TODO: identify whether Ivy ARG/concept graphs still emit these edge cases in the Go port. Tests should construct a graph containing a back edge and pending edge constraint and verify the displayed direction and placement match the intended semantics.

[ ] ### 27. Edge Label Post-Processing Is Incomplete

Inventory refs: PLAN383 item 44; PLAN378 section 8.3.

The Go renderer restores brace markers in ARG labels, and there is regression coverage for that. The inventory also calls out newline restoration and general label text transformation, which is not obviously implemented.

TODO: port the full label transformation rules from Python graph rendering. Tests should include labels with escaped braces, newline encodings, and action labels that combine transition and action names.

[ ] ### 28. ARG Green/Black Outline And Concept Grey Background Styling Diverge

Inventory refs: PLAN378 sections 32.1 through 32.18; PLAN383 items 40 through 42 and 97.

The web styles intentionally use a dark UI and white concept node backgrounds, while Python concept nodes use grey backgrounds and specific outline/line color conventions. Some semantics are preserved by classes, but not all visual distinctions from the inventory are present.

TODO: decide which visual styling details are semantic test requirements. At minimum, tests should cover bottom states, safe states, marked states, cover edges, join edges, concept cardinality, edge truth classes, total/functional/injective/surjective arrows, selection overlays, and node text wrapping.

## P1: Concept Graph And Domain Workflows

[ ] ### 29. Relation Bulk-Toggling By Relation And Class Is Partial

Inventory refs: PLAN383 items 39 and 85; PLAN378 sections 19.1 through 19.7.

The web UI renders per-relation checkboxes for `+`, `?`, `-`, and `T` and syncs them through the backend. The Python/Tk and notebook widgets also support bulk toggling by relation/class controls, including edge class buttons that affect whole groups.

TODO: add bulk toggle controls or document their exclusion. Tests should toggle one class for all relations, one relation across all classes, and a single cell, then verify backend-owned toggle state and graph visibility remain consistent.

[ ] ### 30. Relation Color Assignment Is Partial

Inventory refs: PLAN383 item 40; PLAN378 sections 19.8 and 32.18.

The Go concept renderer computes a fixed palette for sort/node border colors, but relation controls and relation edge line colors do not fully mirror Python's 27-color relation palette behavior. The current Cytoscape frontend also overrides some backend color assumptions.

TODO: decide whether Python's relation color cycling is required for readability and tests. If required, carry relation color metadata through the render payload and test that controls, edges, and labels use stable colors across graph rebuilds.

[ ] ### 31. Constraint Text/Facts UI Is Not The Same As Python's Selectable Text Area

Inventory refs: PLAN383 item 45; PLAN378 sections 19.9, 20.28, 33.9, 33.10, and 33.11.

The web details panel renders facts as selectable buttons and persists selected state through backend `FactSelection`. Python displays constraints under the graph as selectable text and uses selected facts for conjectures, highlighting, and gather behavior.

TODO: finish the fact-selection contract: visual location, multi-select ergonomics, highlight-selected-facts behavior, and exact use in conjecture/CTI flows. Tests should gather facts, select/deselect facts, verify graph highlighting, and assert selected facts are the only facts used by conjecture/minimize/strengthen when appropriate.

[ ] ### 32. Concept Domain Save/Load/Replace Is Backend-Tested But Not UI-Exposed

Inventory refs: PLAN378 sections 30.3 and 30.4.

The Go concept domain/session code has save/load/replace-style functionality and tests, but the browser UI does not expose a concept-domain save/load/replace workflow. Python concept sessions include domain persistence and replacement behavior.

TODO: add UI commands for save domain, load domain, and replace domain if these remain part of the GUI contract. Tests should save a modified domain, reset/replace it, reload the saved domain, and verify concepts, checkboxes, graph stack, and abstract value update correctly.

[ ] ### 33. Add Projection And Ternary Relation UX Needs Full Fidelity

Inventory refs: PLAN383 item 56; PLAN378 section 20.21.

The backend and frontend have add-projection plumbing, but the Python UI presents projection choices derived from concepts and witnesses as part of node context menus. The web menu currently depends on server-provided descriptors and simple prompts, so it needs coverage against real ternary relation cases.

TODO: test and finish projection selection for ternary and higher-arity relations. Tests should right-click a concept node with available projections, add a projection, and verify the new binary concept appears with correct endpoints and relation controls.

[ ] ### 34. Splatter, Materialize, And Empty Need End-To-End Browser Coverage

Inventory refs: PLAN383 items 51 through 55; PLAN378 sections 20.23 through 20.27 and 33.14.

The backend implements splatter, empty, materialize node, positive/negative edge materialization, and fresh witness names. The browser has handlers, but these workflows need end-to-end checks because they combine context menus, dialogs, backend domain mutation, fresh constants, and graph refresh.

TODO: add UI-level tests for each concept context action. Tests should verify witness names with `@` prefixes, negative edge facts, selected-source materialize edge prompts, graph-stack undo/redo after the operation, and relation checkbox preservation.

[ ] ### 35. One-Step Reachability Eliminated-Conjecture Dialog Is Missing

Inventory refs: PLAN383 item 61; PLAN378 sections 31.1 and 31.2.

The web backend has reach/path-reach actions and can update graph state, but the Python one-step reachability flow can report eliminated conjectures in a dialog. That user-facing eliminated-conjecture feedback is not visible in the current browser workflow.

TODO: port eliminated-conjecture reporting for reach operations. Tests should run a reach case that eliminates conjectures and assert the dialog/content lists them before or while updating the graph.

## P1: CTI And Invariant Workflows

[ ] ### 36. CTI Menu Switching Is Partial

Inventory refs: PLAN383 items 64 and 72; PLAN378 sections 20.1 and 20.12 through 20.20.

The web shell always shows static Invariant and Conjecture menus, and descriptor menus are generated from default non-CTI structures. Python CTI mode replaces the ART File/Action menus with CTI-specific invariant menus and CTI concept graph menus.

TODO: make CTI/non-CTI menu switching match the active sheet and state. Tests should load a model with conjectures, cause a CTI, switch between ARG, CTI, and event sheets, and assert the visible menu set and enabled actions match Python's CTI menus.

[ ] ### 37. CTI Used-Relations Display Is Partial

Inventory refs: PLAN383 item 67; PLAN378 sections 20.10 and 33.7.

The frontend can auto-check used relation rows from a failed check result, but Python CTI also has used-relation display and a "relations to minimize" input. The current UI does not expose the full relation-minimization control surface.

TODO: add CTI used-relations and relations-to-minimize UI. Tests should run a failing induction check, verify only relevant relation controls are enabled/checked, edit the relations-to-minimize input, and verify minimize uses that set.

[ ] ### 38. CTI Diagram Pre/Post State Presentation Is Partial

Inventory refs: PLAN383 item 68; PLAN378 sections 20.6 and 20.11.

The backend routes diagram through `CTIUI.Diagram()` and returns concept payloads, but the browser workflow does not clearly present CTI pre-state/post-state context as Python does. Users need to see which CTI state is being diagrammed and how it affects the concept graph.

TODO: add visible CTI state labels and verify diagram uses the intended pre-state. Tests should create a CTI, run Diagram, and assert the concept graph/facts correspond to the pre-state rather than a stale BMC or post-state value.

[ ] ### 39. CTI Strengthen Confirmation Is Missing

Inventory refs: PLAN383 item 75; PLAN378 section 20.18.

Python asks for confirmation before adding a selected conjecture as a strengthened invariant. The web `cti_strengthen` path can add to the compiled module without a visible confirmation step.

TODO: add a confirm dialog with the exact conjecture text before strengthening. Tests should cancel and accept the dialog, verifying cancel leaves conjectures unchanged and accept appends exactly one new conjecture.

[ ] ### 40. CTI Save-Invariant File Content Needs Full Section Fidelity

Inventory refs: PLAN383 item 70; PLAN378 section 20.8.

There is backend coverage for kept/dropped-style invariant output, but the browser save path should be checked against the Python sections and file-dialog behavior. This is high value because saved invariants become source artifacts.

TODO: add browser-level tests for CTI Save invariant, including kept, dropped, and new conjectures. Tests should verify the filename, extension, content sections, labels, and formula formatting.

[ ] ### 41. CTI Bounded-Check Counterexample Viewing Is Partial

Inventory refs: PLAN383 item 71; PLAN378 section 20.9.

The current bounded check path prompts for a bound and calls the backend, and failed top-level checks can offer trace actions. Python CTI bounded check specifically offers to view a counterexample for the conjecture workflow.

TODO: implement the CTI-specific BMC result path, including view-counterexample affordance. Tests should run a CTI bounded check with a counterexample and assert the user can open the corresponding trace/ARG sheet.

### 42. CTI Minimize Needs User-Visible BMC/Core Details

Inventory refs: PLAN383 item 76; PLAN378 sections 20.15 and 33.10.

The backend has a minimize action, but Python's CTI minimize first runs BMC and then reduces selected facts using an unsat core, with user-visible fact/core selection behavior. The current UI hides most of that reasoning behind a single action/status.

TODO: expose enough minimize detail for users and tests to validate the same facts were minimized. Tests should select facts, run minimize, verify the BMC bound used, inspect the resulting conjecture, and verify unselected facts do not appear.

### 43. CTI Check-Sufficient And Check-Relative-Induction Result Dialogs Are Partial

Inventory refs: PLAN383 items 77 and 78; PLAN378 sections 20.16 and 20.17.

The backend returns ok/message for these checks, and the frontend can call the actions. Python presents these checks as conjecture workflow results tied to the selected conjecture.

TODO: make the result display explicit and tied to the active selected facts/conjecture. Tests should run both actions with sufficient, insufficient, inductive, and non-inductive selections and assert the displayed result text and backend state.

## P2: Event Viewer And Trace Display

### 44. Trace Serialization Display Is Missing

Inventory refs: PLAN378 sections 24.1 through 24.5 and 27.2.

The web UI can load event traces and show check-result details, but Python trace display has serialization-to-lines behavior, infinite-loop detection, function-call formatting, trace line numbers, and a detailed option. Those formatting choices are not exposed as a trace display feature in the web UI.

TODO: port trace formatting or explicitly fold it into the event trace viewer. Tests should feed traces with repeated states, function calls, detailed and non-detailed output, and line numbers, then verify the displayed text matches the Python behavior.

[ ] ### 45. Event Viewer Launch As A Standalone Tool Is Missing

Inventory refs: PLAN383 item 79; PLAN378 sections 25.1 and 25.2.

The web UI can open event trace sheets from the main app. Python also has an event viewer launched as its own Tix tree UI/notebook, including `ivy_show.py` workflows.

TODO: decide whether `ivyweb` should support an event-viewer-only launch mode. Tests should open a `.iev` file directly into the intended web surface and verify no model session is required if standalone mode is in scope.

### 46. Event Tree Selection Style And Keyboard Ergonomics Are Partial

Inventory refs: PLAN378 section 25.7.

The web event tree supports lazy child expansion, filtering, find forward/reverse, and saved patterns. Tk HList selection style, keyboard traversal, and selection affordances are not fully represented.

TODO: define the required event tree interaction details for the web port. Tests should cover row selection, expansion/collapse, keyboard navigation if supported, preserved selection after filtering/finding, and visible selected-row styling.

### 47. Event Pattern File Dialog Semantics Need Coverage

Inventory refs: PLAN383 item 82; PLAN378 sections 25.5 and 25.6.

Pattern add/remove/load/save/clear exists, but save/load currently use web dialogs/downloads rather than Python file dialogs. The exact persistence behavior and malformed-pattern handling should be nailed down.

TODO: add end-to-end tests for loading and saving pattern files from the browser shell. Tests should verify newline handling, invalid patterns, duplicate patterns, selected pattern preservation, and backend-authoritative state after failed operations.

## P2: Notebook/Widget Workflows And Extensions

### 48. Analysis-Session History Navigation Is Missing

Inventory refs: PLAN383 item 90; PLAN378 sections 21.1 through 21.4 and 33.2.

Python notebook widgets expose first/prev/next/last history navigation, step info display, active-element auto-click, and modal messages. The current web UI has sheet switching and tutorial URL history, but not analysis-session step history.

TODO: implement or exclude analysis-session history navigation for web UI. Tests should drive a multi-step analysis session, navigate back and forward, and verify the ARG/concept/transition widgets reflect the selected history step.

[ ] ### 49. Proof Goal And CRG Widget Interactions Are Missing From The Browser

Inventory refs: PLAN378 sections 33.3 and 33.4.

The backend has proof stack data and proof actions, but the browser has no proof graph/pane and `proof_updated` is marked future work. Python widgets update concept and transition views when proof goals or CRG nodes are clicked.

TODO: add a proof/goal graph UI or mark proof widgets out of scope. Tests should click proof goals and CRG nodes and verify the concept graph, transition view, and selected goal state update.

[ ] ### 50. Abstractor, BMC Bound, Relations-To-Minimize, And Transition Log Controls Are Missing

Inventory refs: PLAN378 sections 33.5 through 33.8.

The Python widget layer includes an abstractor selection dropdown, BMC bound dropdown, relations-to-minimize input, and transition view log file behavior. The current web UI has a mode select and a currentBound field, but not those specific controls.

TODO: decide which controls should be surfaced in the web controller. Tests should set each control, run the dependent command, and verify the backend receives the selected abstractor/bound/relation/log setting.

### 51. UI Extension Points Are Not Browser-Extensible Yet

Inventory refs: PLAN383 items 7, 87, and 88; PLAN378 section 29.

The Go code has some registered ARG helper actions, but the Python `ui_extensions_api.py` model supports extension-point registration, dynamic ARG/goal node actions, interaction decorators, and modal fact/core selection. The browser cannot currently discover and render arbitrary extension-provided actions in the same way.

TODO: design a web extension action descriptor API for ARG nodes, proof goals, and modal interactions. Tests should register a fake extension action, verify it appears in the right context menu, exercise its dialog interactions, and confirm it mutates backend state.

[ ] ### 52. Interactive UPDR Modal Fact/Core Selection Is Missing

Inventory refs: PLAN383 item 88; PLAN378 sections 28.8 through 28.10 and 29.6.

The backend can run PDR/UPDR-like actions, but Python interactive UPDR uses modal user selection for facts and cores. The current web PDR step is a direct action with status messaging.

TODO: port interactive UPDR selection dialogs or provide a non-interactive mode with explicit limitations. Tests should run a PDR case that requires user selection and verify the modal choices drive the backend tactic result.

## P3: Lower-Level Rendering And Styling Details

[ ] ### 53. Generic Canvas Element Mapping Is Not Ported

Inventory refs: PLAN378 section 26 and PLAN383 item 104.

Python maps generic graph elements such as polygons, splines, text, clusters, and toolkit callbacks to Tk canvas objects. The web port maps semantic ARG/concept/proof payloads into Cytoscape elements, which does not cover the whole generic graph-element model.

TODO: determine whether any remaining Python GUI behavior depends on generic graph elements rather than semantic ARG/concept elements. If so, add a generic render layer or targeted conversions and tests for every element kind.

### 54. Scrollbar/Scroll-Region Semantics Are Replaced By Pan/Zoom

Inventory refs: PLAN378 sections 2.5, 8.7, and 23.4.

Tk canvases expose scrollbars and update scroll regions after rendering. Cytoscape uses pan/zoom/fit, so the direct behavior is not present.

TODO: decide whether this is an acceptable toolkit substitution. Tests should at least verify large graphs remain navigable, fit/reset works, and selected/highlighted nodes can be centered or scrolled into view.

### 55. Text Fitting And Wrapping Need Semantic Regression Tests

Inventory refs: PLAN378 sections 32.16 and 32.17.

Python graph rendering sets wrapping and font sizes for ARG and concept nodes. The web style wraps text, but node sizes are heuristically derived and can diverge for long action/concept labels.

TODO: add graph-render tests for long state labels, relation labels, concept labels, and multi-line edge labels. Tests should assert text remains visible and does not overlap enough to hide the graph semantics.

## P3: Source, Files, And Recent State

[ ] ### 56. File Browser Raise/Re-use Semantics Are Missing

Inventory refs: PLAN383 item 20; PLAN378 sections 23.1 through 23.5.

The web UI has one editor and can highlight a line in it. Python has a separate source browser window that is reused, raised, and line-highlighted independently.

TODO: if source browsing remains in scope, implement a source browser surface with reuse and raise semantics. Tests should open source for two different edges and verify the same browser surface updates and remains separate from the editable model.

### 57. Recent Files And Browser Persistence Do Not Match Python Analysis-State Semantics

Inventory refs: PLAN383 items 3, 5, and 34; PLAN378 sections 3 and 5.1.

The web UI has recent sessions/files and local browser persistence, which is useful but not equivalent to Python notebook tab and `.a2g` state persistence. Restored web analysis sheets can become visual-only, which is a web-specific divergence.

TODO: document and test persistence boundaries: browser session restore, web JSON state restore, Python `.a2g` restore, and live backend session restore. Tests should make clear which restored sheets are editable/live versus visual-only.

## P3: Deliberate Non-Goals To Confirm

### 58. Tk Root Window Title/Palette And Tix Notebook Details

Inventory refs: PLAN383 items 4 and 5; PLAN378 sections 1.2 through 1.5.

The browser shell replaces Tk/Tix root-window behavior. Exact root title, palette, Tix widget creation, and Tk update behavior should likely be non-goals, but they should be marked that way.

TODO: record these as accepted toolkit substitutions unless a user-visible web equivalent is desired. Tests should focus on browser title/session labels, tab labels, and layout rather than Tk internals.

### 59. Dynamic Tk Class Mixing Is Not A Web Requirement Unless Extensions Depend On It

Inventory refs: PLAN383 item 7.

Python dynamically mixes UI classes into Tk widgets. The Go/web port uses typed Go structs and a TypeScript controller, so dynamic class mixing itself is not required unless it is the mechanism by which extension behaviors appear.

TODO: close this item by tying it to the web extension API work. Tests should validate extension behavior, not Python's implementation mechanism.

### 60. Jupyter Widget JavaScript, jQuery UI, And Bootstrap Modal Plumbing Need A Scope Decision

Inventory refs: PLAN378 sections 28.3 through 28.10.

The Python inventory includes notebook-specific JavaScript widget integration that is not the same product surface as the hosted Go web UI. None of the jQuery UI/Bootstrap widget plumbing exists in the current web app.

TODO: decide whether notebook widget parity belongs in `goivy/webui`, a separate package, or out of scope. If in scope, add separate widget-contract tests rather than mixing them with the hosted web shell tests.

