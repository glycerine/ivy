# PLAN385 TODO: Missing Go/Web UI Features vs Python Tcl GUI

Audit date: 2026-05-13.

Sources subtracted:

- `~/ivy/already_applied_plans/PLAN378_cc_python_gui_inventory.md`
- `~/ivy/already_applied_plans/PLAN383_codex_tcl_gui_inventory.md`
- Current implementation under `~/ivy/goivy/webui`, especially `DATA_MODEL.md`, `webui_session.go`, `webui_ui_main.go`, `webui_ui_cti.go`, `webui_graph_widget.go`, `webui_cyrender.go`, `static/index.html`, and `frontend/src/services/*`.

Scope note: this is not a request to recreate Tk pixel-for-pixel. A behavior is counted as present when the Go/web UI exposes the same user workflow and state effect, even if the toolkit is different. A behavior is counted as missing or partial when the current port either has no workflow, has backend-only code with no reachable UI, has a web-only substitute that cannot load or save Python GUI state, or has a visible behavior that is materially weaker than the Python Tcl GUI.

## Mostly Subtracted

The current Go/web UI already covers a large core: it can load Ivy source into a hosted session, display Reachability (ARG; Abstract Reachability Graph) and Concept graphs in Cytoscape, maintain multiple analysis/event sheets, select ARG states, open concept graphs for states, run top-level checks, expose many ARG node and edge context actions, show concept checkboxes for `+`, `?`, `-`, and `T`, gather/select facts, perform several concept graph operations, run CTI/invariant actions, save/export several text artifacts, and load/filter/search event traces. Those covered areas are not repeated below except where the port is incomplete or the behavior differs enough to matter for tests.

## P1: Menus, Commands, And Dialogs

### 2. Backend Session Save Is Not Wired To Python File Menu Semantics

Inventory refs: PLAN383 items 19, 34, and 35; PLAN378 sections 5.1, 5.2, and 22.11.

The current web UI has browser-side "Save Analysis State" and backend `SaveState`, but File menu descriptors also include actions such as `save` and `save_abstraction` that do not all route through the Python-shaped menu semantics. Some static buttons work, while descriptor-driven menus can dispatch unsupported action names.

TODO: make File menu behavior explicit and complete: save current model, save analysis state, save abstraction, remove tab, and exit/close should each map to one controller command or be removed from the menu. Tests should click every File menu item exposed by the browser shell and assert a successful state effect or a deliberate disabled state, not an "unknown action" backend error.

Done 2026-07-05: Added focused red tests for explicit descriptor File commands in Go and frontend dispatcher routing. The ARG descriptor File menu now exposes `save_model`, `save_analysis_state`, `save_abstraction`, `remove_tab`, and `exit`, and JSDOM coverage clicks each dynamic File item to verify it routes through browser controller methods instead of backend `runAction`.

[x] ### 8. Descriptor Menus Are Static And Not Session/Mode Aware

Inventory refs: PLAN383 items 9, 37, 64, and 72; PLAN378 sections 4, 20.1, and 20.12 through 20.20.

The backend exposes browser menu descriptors, but `BuildBrowserMenuDescriptors()` constructs them from fresh default `AnalysisGraphUI` and `GraphWidget` instances rather than the active session, CTI UI, or active sheet. The static HTML has additional Invariant and Conjecture menus, so there are two overlapping menu systems with different coverage.

TODO: collapse menu ownership into one active-session menu model or make the static menus the only source of truth. Tests should load a model, enter CTI and non-CTI contexts, switch sheets, and assert that only valid current actions are enabled and dispatchable.

Done 2026-07-05: Added active menu context (`sheet_id`/`ui_mode`) through runtime, hosted Go, and browser-WASM snapshot paths. Go descriptors now come from the actual session: reachability mode exposes ARG analysis menus, CTI mode exposes Invariant/Conjecture menus, and event sheets expose only global File commands. Added focused Go and frontend tests for active sheet/workflow menu selection.

[x] ### 9. Some Menu Items Dispatch Unsupported Backend Actions

Inventory refs: PLAN383 items 9, 31, 33, 34, and 35; PLAN378 sections 5 and 7.

The descriptor menu includes actions such as `save`, `remove_tab`, `exit`, mode actions, and `recalculate_all`; the generic menu dispatcher sends them to `runAction`, but `Session.ExecuteAction` does not implement all of those action names. Related functionality exists elsewhere in the controller, so the gap is wiring and command naming.

TODO: create a command map for every descriptor/static menu item and add regression tests that no visible menu item produces an unknown-action response. Where a menu action is intentionally browser-only, it should dispatch to the controller rather than the backend action endpoint.

Done 2026-07-05: Added a frontend dispatcher command map for descriptor workflow actions that were falling through to backend `runAction`: mode changes, CTI invariant actions, reachability ARG actions, and CTI concept aliases. Added red/green regression coverage proving those visible descriptor actions route to controller methods instead of unknown backend actions.

[x] ### 10. Dialog Pre-Seeding For Tests Is Missing

Inventory refs: PLAN383 item 11; PLAN378 sections 22.2 and 28.9 through 28.10.

The Python GUI can pre-seed dialog answers to make workflows testable without manual input. The web controller has promise-based dialogs, but no shared test-answer queue or deterministic answer injection mechanism.

TODO: add a small dialog-answer harness owned by the controller or command registry. Tests should pre-seed entry, integer, listbox, multiple-selection, and button-list dialogs, run real commands, and assert the command receives the injected result.

Done 2026-07-05: Added controller-owned `preseedDialogAnswers`/`clearDialogAnswers` queue support in `IvyRuntime`. The real dialog methods now consume queued answers for entry, integer, listbox single/multiple, text, OK, OK/cancel, and button-list dialogs. Added tests that preseed all requested dialog kinds and verify a real `rememberGraph` command receives the injected entry result.

[x] ### 11. Dialog Return Semantics Differ From Tk In Several Places

Inventory refs: PLAN383 items 13 through 18; PLAN378 sections 22.3 through 22.10.

The web `listboxDialog` generally returns option values, while the Tk inventory distinguishes index-returning listboxes, multiple-selection indices, and button-list cancel behavior. Some callers work around this by using values that are already indices, but the dialog contract is not the same as Python's.

TODO: define and test dialog return semantics per dialog kind. Tests should cover single selection, multiple selection, cancel, Escape, Return, out-of-range integer input, and callers that need selected indices rather than displayed text.

Done 2026-07-05: Added focused dialog-semantic coverage for Return-submitted entry dialogs, Tk-compatible listbox index return via `returnIndex`/`returnIndices`, cancel/Escape behavior for listbox and button-list dialogs, and integer dialogs that stay open on out-of-range input. Implemented the matching controller behavior in `IvyRuntime`; full `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`.

[x] ### 12. RunContext/Error Modal Behavior Is Only Partially Ported

Inventory refs: PLAN383 item 10; PLAN378 sections 22.13 through 22.15 and 33.1.

Python wraps long-running UI callbacks in a run context that shows blocking Ivy error dialogs and restores the cursor. The web UI usually writes errors to the status bar/toast and sometimes uses loading overlays, so important errors can be non-modal and inconsistent across commands.

TODO: define a web equivalent of `RunContext` for command execution, including modal-vs-status policy and busy/ready visual state. Tests should force backend errors in check, ARG action, concept action, CTI action, and event filtering paths and assert consistent user-visible error handling.

Done 2026-07-06: Added a shared frontend run-context helper that marks command execution busy, restores the ready state, formats Ivy error messages, and opens blocking `Ivy error` dialogs for backend failures. Added focused red/green coverage forcing errors through check, ARG action, concept action, CTI action, and event-filter paths; those paths now show a modal error, set the error status, hide the loading overlay, and clear `aria-busy`/`data-ivy-run-context`. Full `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`.

[x] ### 13. Save-As Dialog Filters And Titles Are Incomplete

Inventory refs: PLAN383 item 19; PLAN378 section 22.11.

Python uses file dialogs with specific titles and filters for model files, abstractions, invariants, event patterns, and analysis-state files. The web UI uses File System Access or download fallbacks, but the filter/title coverage is uneven and not obviously tied to every export action.

TODO: inventory every save/open/export path and give it an explicit suggested name, MIME type, extension set, and cancel behavior. Tests should stub `showSaveFilePicker` and download fallback paths for `.ivy`, `.ivyweb.json`, invariant files, abstraction files, DOT exports, and event pattern files.

Done 2026-07-06: Added a shared save-dialog descriptor service with explicit suggested names, descriptions, MIME types, and extension sets for Ivy models, IvyWeb analysis state, invariants, abstractions, Graphviz DOT, and event-pattern files. Added focused red/green tests that stub `showSaveFilePicker` and fallback downloads across those paths. Runtime invariant/abstraction saves now use the common download fallback, analysis state accepts `.ivyweb.json`, and event pattern saves support the File System Access picker. Full `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`.

## P1: ARG Graph Behavior

[x] ### 16. Safety Error Trace Viewing From Node Safety Is Missing

Inventory refs: PLAN378 sections 12.5, 13.4, 24, and 27.2. ( ~/ivy/already_applied_plans/PLAN378_cc_python_gui_inventory.md )

Top-level checks can show trace actions, but node-level safety from the context menu only returns safe/message fields. Python can present a "View Error Trace" path when a safety or BMC check produces a counterexample.

TODO: return trace data or a trace ARG from node safety/BMC failures and add a visible "view trace" action. Tests should create an unsafe node, run node safety, click the trace action, and verify an event/ARG trace sheet opens with the counterexample.

(from ~/ivy/already_applied_plans/PLAN378_cc_python_gui_inventory.md 12.5 Safety Check: "View Error Trace" Button)

When the local safety check fails, one of the buttons in the `buttons_dialog_cancel` is labeled to view the error trace. Clicking it calls the trace viewing function and adds a new ARG tab.

**Testing:** Trigger a local safety failure. In the resulting dialog, click the view-trace button. Confirm a new tab with the counterexample trace ARG appears. Confirm the trace states are highlighted appropriately.

Done 2026-07-06: Verified the Go node-safety backend already returns `trace_arg`/`trace_sheet_id` for bounded safety failures and marks the final trace state, with existing Go coverage. Added focused red/green frontend coverage requiring the visible "View error trace" action to switch into reachability mode before opening the trace ARG sheet. Full `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`.


[x] ### 17. BMC Entry Points Are Not Fully Exposed As ARG Workflows

Inventory refs: PLAN378 sections 13.1 through 13.4 and 16.3; PLAN383 items 71 and 92.

There is a top-level bounded check and backend methods for BMC-like operations, but the Python ARG workflow includes bounded checking from an ARG node, bound entry, unreachable/counterexample dialogs, and optional trace viewing. Those behaviors are only partially reachable in the browser.

TODO: add the missing ARG/CTI BMC commands and result dialogs, or clearly merge them with the existing bounded-check button. Tests should cover user-entered bounds, unreachable results, reachable counterexamples, and trace-sheet creation.

Done 2026-07-06: Added an ARG node "Bounded check" command backed by `Session.ArgNodeAction("bmc")`. The backend now accepts user-supplied bounds and error conditions, returns explicit reachable/unreachable result payloads, registers counterexample trace ARG sheets, and marks the final trace state. The frontend prompts for bound/error condition, preserves the current bound, and shows a Python-style View dialog that opens the trace sheet. Added focused red/green Go and frontend tests for user-entered bounds, unreachable results, reachable counterexamples, and trace-sheet creation; full `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`.

[x] ### 18. Recalculate-State Behavior Is Missing Or Ambiguous

Inventory refs: PLAN383 item 32; PLAN378 sections 7.1, 10.2, and 16.4.

The port supports recalculating all transitions and recalculating an edge, but the Python inventory also calls out recalculate of the current state/concept graph state. The current UI does not make that state-level distinction clear.

TODO: identify the Python state-level recalculate behavior and add a matching command if it is distinct from edge/all recalculation. Tests should select a state, modify concept/domain information, run recalculate state, and verify the concept graph is updated without recalculating unrelated targets.

Done 2026-07-06: Matched Python's concept-graph Recalculate path by having `GraphWidget.Recalculate()` ask its parent `AnalysisGraphUI` to recalculate the selected ARG state via `AnalysisGraph.RecalculateState` before reloading the parent clauses into the concept graph. The backend `recalculate` action now targets the active sheet's current concept graph and returns a refreshed concept payload, while the frontend sends the active `sheet_id` and applies that payload directly. Added red/green Go tests for join-state recalculation without touching unrelated states and session action payloads, plus a focused frontend test for active-sheet dispatch. While running `make test`, fixed two build-contract gaps it exposed: `ConformBackend.GetMenus` now satisfies the backend interface, and `ivySample` is shared by web and non-web webui tests. Full `make test` and `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`.

[x] ### 19. `Extend` Closed-Node Reporting Is Too Weak

Inventory refs: PLAN383 item 30; PLAN378 sections 14.1 and 14.2.

The backend can attempt an extension and returns an error when the state is closed. Python shows an explicit closed-node dialog and otherwise executes the chosen extension and rebuilds the graph.

TODO: make closed-node and extension-success UX match the inventory. Tests should run Extend on a closed node and assert a visible closed-node message, then run it on an extendable node and assert the new state/edge appears.

Done 2026-07-06: Added a Python-shaped Extend result path. Closed nodes now return a non-error payload with `closed=true`, `result=closed`, and the explicit `State N is closed.` message that the frontend shows through an `ivyweb` OK dialog. Successful Extend now evaluates the chosen state equation, adds the resulting ARG state with action provenance, switches the concept graph to the new state, and returns both ARG and concept payloads. Added red/green Go tests for closed-node payloads and successful state/concept updates, plus a frontend test for the closed-node dialog. Full `make test-web` and `make test` passed with `GOCACHE=/tmp/goivy-gocache-codex`.

[x] ### 20. Try-Conjecture Source Browsing And Result Views Are Partial

Inventory refs: PLAN383 item 57; PLAN378 sections 16.1 through 16.4.

The web UI prompts for conjecture choices and calls the backend, but the Python workflow also browses conjecture source and chooses different result views depending on bounded, induction, or PDR mode. The current path mostly ends in status updates and backend effects.

TODO: complete the mode-specific try-conjecture flow, including source browsing and resulting concept/trace graph presentation. Tests should try a conjecture in bounded, induction, and PDR modes and assert the expected sheet/dialog/source behavior for each mode.

Done 2026-07-06: Completed the mode-specific Try conjecture workflow. The backend now resolves dialog-returned pretty conjectures back to module formulas, returns source filename/line/content for browsing, opens PDR conjectures as concept graphs with dual conjecture constraints, and returns bounded/induction result views as either unreachable message payloads or registered trace ARG sheets with explicit `reachable` status. The frontend now browses returned source, applies returned concept graphs, shows unreachable results in an `ivyweb` dialog, and opens trace-result sheets through a Python-style View dialog. Added focused red/green Go coverage for PDR concept/source behavior and bounded/induction result views, plus frontend coverage for concept, message, and trace try-conjecture results.

[x] ### 21. Interpolant Refinement Workflow Is Missing

Inventory refs: PLAN378 sections 17.1 through 17.3.

Python supports refinement with interpolants, including a vacuous-pre-state message and adding predicates to the concept domain. No equivalent web workflow is visible in the current controller or menus.

TODO: port the interpolant/refinement command path. Tests should trigger a refinement case, accept an interpolant, and verify the new predicate appears in the concept relation controls and graph.

Done 2026-07-06: Added an explicit interpolant refinement workflow. `pdr_step` now returns a Python-shaped refinement payload (`refinement_action`, `refinement_kind`, `refinement_message`, and interpolant text) when reverse produces an interpolant. The browser shows a text dialog with a `Refine` action and, when accepted, calls `refine_with_interpolant`. The backend action parses the interpolant in the compiled module context, adds PDR refinements to `AbstractionPredicates`, and adds non-PDR refinements as concept spaces plus a visible current-graph concept. Added red/green Go tests for PDR predicate refinement and non-PDR concept-space/graph-domain refinement, plus frontend coverage for accepting the Refine dialog from PDR step.

[x] ### 23. Direct Context-Popup Shortcut For Single `<>` Action Is Not Preserved

Inventory refs: PLAN383 item 105; PLAN378 section 9.2.

The web UI handles ARG left-click as "view state", which covers the main user effect. Python's popup helper also has a special shortcut for a single `<>` action in context-action lists.

TODO: decide whether the shortcut matters in the web context-menu implementation. If it matters, test a one-action context list and verify the action executes directly without rendering a redundant popup.

Done 2026-07-06: Preserved the Python `TkCyCanvas.make_popup` direct-action shortcut for web context actions. When a server-provided ARG node action list contains exactly one `<>` action, the runtime now dispatches it immediately instead of rendering a context menu; `view_state` routes through the same path as the normal ARG left-click state view. Added focused red/green frontend coverage asserting the action runs directly and no redundant menu is shown.

## P1: Graph Rendering Fidelity

[x] ### 25. Cluster/Subgraph Box Rendering Is Missing

Inventory refs: PLAN378 sections 8.13, 26.4, and 104.

The Python canvas renderer draws cluster/subgraph rectangles when graph elements contain them. The current Cytoscape ARG/concept renderer does not expose or draw cluster boxes.

TODO: either render cluster/subgraph boxes in Cytoscape or record them as intentionally omitted. Tests should load a graph with subgraph/cluster metadata and assert the web graph displays the same grouping affordance.

Done: Go concept rendering now emits Python-compatible `subgraphs` shape elements for visible clusters, and the web Cytoscape runtime translates those shape records into compound `subgraph_box` parent nodes so grouped concepts display with cluster boxes. Covered by focused Go renderer tests and a frontend graph runtime test.

[x] ### 26. Back Edge Reversal And Pending Edge Constraints Are Missing

Inventory refs: PLAN378 sections 26.3 and 26.6.

Python graph rendering handles DOT-specific back edge reversal and pending edge constraints. The current Cytoscape pipeline has no matching representation.

TODO: identify whether Ivy ARG/concept graphs still emit these edge cases in the Go port. Tests should construct a graph containing a back edge and pending edge constraint and verify the displayed direction and placement match the intended semantics.

Done: Concept graph rendering now annotates layout-only back-edge reversal while preserving the visible source/target, and `pending` edges are marked `layout_constraint:false`. The frontend preserves those fields, adds hidden `layout_only` edges for Dagre ranking, excludes visible ignored edges from the layout collection, and keeps pending edges visible but unconstrained. Covered by focused Go renderer, frontend runtime, and typed metadata tests.

[x] ### 27. Edge Label Post-Processing Is Incomplete

Inventory refs: PLAN383 item 44; PLAN378 section 8.3.

The Go renderer restores brace markers in ARG labels, and there is regression coverage for that. The inventory also calls out newline restoration and general label text transformation, which is not obviously implemented.

TODO: port the full label transformation rules from Python graph rendering. Tests should include labels with escaped braces, newline encodings, and action labels that combine transition and action names.

Done: Web ARG rendering now restores Python/DOT label encodings by converting `-[`/`]-` back to braces, decoding `\l` and `\n` line markers to real newlines, and removing the terminal left-justify marker added by Python graph rendering. Covered by an ARG renderer regression that includes brace markers, newline encodings, and a `trans -> action` label.

[x] ### 28. ARG Green/Black Outline And Concept Grey Background Styling Diverge

Inventory refs: PLAN378 sections 32.1 through 32.18; PLAN383 items 40 through 42 and 97.

The web styles intentionally use a dark UI and white concept node backgrounds, while Python concept nodes use grey backgrounds and specific outline/line color conventions. Some semantics are preserved by classes, but not all visual distinctions from the inventory are present.

TODO: decide which visual styling details are semantic test requirements. At minimum, tests should cover bottom states, safe states, marked states, cover edges, join edges, concept cardinality, edge truth classes, total/functional/injective/surjective arrows, selection overlays, and node text wrapping.

Done: The semantic contract is now selector/class affordances rather than exact palette parity: bottom/safe/marked ARG states, cover/join/action edges, concept cardinality borders, edge truth line styles, total/functional/injective/surjective arrows, selection overlays, and text wrapping are covered in frontend style tests. The web palette remains intentionally dark/white, but unknown concept nodes now match Python/Go semantics with no cardinality border.

## P1: Concept Graph And Domain Workflows

[x] ### 29. Relation Bulk-Toggling By Relation And Class Is Partial

Inventory refs: PLAN383 items 39 and 85; PLAN378 sections 19.1 through 19.7.

The web UI renders per-relation checkboxes for `+`, `?`, `-`, and `T` and syncs them through the backend. The Python/Tk and notebook widgets also support bulk toggling by relation/class controls, including edge class buttons that affect whole groups.

TODO: add bulk toggle controls or document their exclusion. Tests should toggle one class for all relations, one relation across all classes, and a single cell, then verify backend-owned toggle state and graph visibility remain consistent.

Done: Relation-name buttons now have regression coverage for toggling every edge display class for a relation, class headers have coverage for toggling one class across all relation rows, and single-cell toggles have coverage for graph visibility. The frontend optimistic model update now mirrors backend `set_checkbox` normalization by updating both full and bare relation names, and by mapping `+`, `?`, and `-` edge toggles to the corresponding unary node-label display classes, so rendered graph visibility stays consistent before the backend refresh completes.

[x] ### 30. Relation Color Assignment Is Partial

Inventory refs: PLAN383 item 40; PLAN378 sections 19.8 and 32.18.

The Go concept renderer computes a fixed palette for sort/node border colors, but relation controls and relation edge line colors do not fully mirror Python's 27-color relation palette behavior. The current Cytoscape frontend also overrides some backend color assumptions.

TODO: decide whether Python's relation color cycling is required for readability and tests. If required, carry relation color metadata through the render payload and test that controls, edges, and labels use stable colors across graph rebuilds.

Done: Python-compatible relation color cycling is now part of the web contract. The shared palette has all 27 Python `line_colors`, concept graph rendering assigns deterministic relation colors from sorted relation ids, edge elements carry `line_color` metadata, and concept payloads expose `relation_colors` for controls. The frontend model preserves relation color data, state checkbox relation buttons use the same color metadata, and the graph runtime applies `line_color` to Cytoscape edge lines and arrows. Tests cover palette parity, stable edge metadata across rebuilds, typed metadata preservation, row button styling, and runtime edge styling.

[x] ### 31. Constraint Text/Facts UI Is Not The Same As Python's Selectable Text Area

Inventory refs: PLAN383 item 45; PLAN378 sections 19.9, 20.28, 33.9, 33.10, and 33.11.

The web details panel renders facts as selectable buttons and persists selected state through backend `FactSelection`. Python displays constraints under the graph as selectable text and uses selected facts for conjectures, highlighting, and gather behavior.

TODO: finish the fact-selection contract: visual location, multi-select ergonomics, highlight-selected-facts behavior, and exact use in conjecture/CTI flows. Tests should gather facts, select/deselect facts, verify graph highlighting, and assert selected facts are the only facts used by conjecture/minimize/strengthen when appropriate.

Done: Fact selection now follows Python SelectMultiple semantics: newly gathered facts default selected, explicitly selected subsets remain active, and an empty selected subset falls back to all facts. `set_fact_selection` recomputes highlighted fact selections and returns an updated concept snapshot; concept payload rendering carries backend-selected nodes/edges as `selected_node`/`selected_edge`, and the frontend applies returned concept snapshots after fact toggles so highlighting refreshes immediately. Covered by backend active-fact fallback, payload highlight, regression/conjecture tests, and frontend details-service snapshot application tests.

[x] ### 32. Concept Domain Save/Load/Replace Is Backend-Tested But Not UI-Exposed

Inventory refs: PLAN378 sections 30.3 and 30.4.

The Go concept domain/session code has save/load/replace-style functionality and tests, but the browser UI does not expose a concept-domain save/load/replace workflow. Python concept sessions include domain persistence and replacement behavior.

TODO: add UI commands for save domain, load domain, and replace domain if these remain part of the GUI contract. Tests should save a modified domain, reset/replace it, reload the saved domain, and verify concepts, checkboxes, graph stack, and abstract value update correctly.

Done: Concept domain persistence is now exposed through browser actions and concept-menu commands. The backend supports `save_domain`, `load_domain`, and `replace_domain` for the active concept graph, stores browser-visible saved domains, mirrors full interactive-session saves when present, restores domains through the graph replacement path, and returns updated concept snapshots with graph-stack state. The frontend prompts for a domain name, calls the backend action with the active sheet, applies returned concept snapshots immediately, and routes menu descriptors to the new commands. Covered by a reachability-domain save/load/replace backend workflow using a modified diagram domain, plus frontend command and descriptor routing tests.

[x] ### 33. Add Projection And Ternary Relation UX Needs Full Fidelity

Inventory refs: PLAN383 item 56; PLAN378 section 20.21.

The backend and frontend have add-projection plumbing, but the Python UI presents projection choices derived from concepts and witnesses as part of node context menus. The web menu currently depends on server-provided descriptors and simple prompts, so it needs coverage against real ternary relation cases.

TODO: test and finish projection selection for ternary and higher-arity relations. Tests should right-click a concept node with available projections, add a projection, and verify the new binary concept appears with correct endpoints and relation controls.

Done: Ternary projection selection now follows the Python contract, which only offers binary projections from arity-3 concepts with a concrete witness. Backend coverage builds a real ternary relation, obtains the node projection descriptor, adds the projection through the session action, and verifies the interactive domain, render domain, edge list, arity, and relation controls all update for the new binary concept. The projection endpoint now returns an updated concept snapshot, and the frontend applies that snapshot immediately after `addProjection` while preserving the existing CTI projection context-menu filtering behavior.

[x] ### 34. Splatter, Materialize, And Empty Need End-To-End Browser Coverage

Inventory refs: PLAN383 items 51 through 55; PLAN378 sections 20.23 through 20.27 and 33.14.

The backend implements splatter, empty, materialize node, positive/negative edge materialization, and fresh witness names. The browser has handlers, but these workflows need end-to-end checks because they combine context menus, dialogs, backend domain mutation, fresh constants, and graph refresh.

TODO: add UI-level tests for each concept context action. Tests should verify witness names with `@` prefixes, negative edge facts, selected-source materialize edge prompts, graph-stack undo/redo after the operation, and relation checkbox preservation.

Done 2026-07-06: Added focused red/green Go and browser coverage for concept-node Splatter, Empty, Materialize, positive edge Materialize, negative edge Dematerialize, and selected-source Materialize edge prompts using the hosted Go backend and typed concrete constants. Node Materialize now routes through the graph-stack-aware session action and returns witness/concept payloads; edge Materialize seeds/syncs the interactive concept session, preserves it through graph-stack copies, returns witnesses plus concept snapshots, and keeps the legacy empty `/concept/undo` endpoint error contract. Browser tests verify fresh witness concepts/facts, negative edge facts, selected-source relation dialog behavior, relation-checkbox preservation, and graph-stack undo/redo. Full `make test` and `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`.

[x] ### 35. One-Step Reachability Eliminated-Conjecture Dialog Is Missing

Inventory refs: PLAN383 item 61; PLAN378 sections 31.1 and 31.2.

The web backend has reach/path-reach actions and can update graph state, but the Python one-step reachability flow can report eliminated conjectures in a dialog. That user-facing eliminated-conjecture feedback is not visible in the current browser workflow.

TODO: port eliminated-conjecture reporting for reach operations. Tests should run a reach case that eliminates conjectures and assert the dialog/content lists them before or while updating the graph.

Done 2026-07-06: Ported one-step reach eliminated-conjecture reporting. The backend now filters conjectures against the reached model and returns `eliminated_conjectures` plus the Python-shaped dialog message for both current-sheet and generic reach actions. The frontend `reachStep()` shows a non-cancel listbox dialog before completing the reach status update. Added focused red/green Go coverage for eliminated conjecture payloads, Vitest coverage for the runtime dialog behavior, and hosted-Go Playwright coverage for the browser dialog. Full `make test` and `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`.

## P1: CTI And Invariant Workflows

[x] ### 36. CTI Menu Switching Is Partial

Inventory refs: PLAN383 items 64 and 72; PLAN378 sections 20.1 and 20.12 through 20.20.

The web shell always shows static Invariant and Conjecture menus, and descriptor menus are generated from default non-CTI structures. Python CTI mode replaces the ART File/Action menus with CTI-specific invariant menus and CTI concept graph menus.

TODO: make CTI/non-CTI menu switching match the active sheet and state. Tests should load a model with conjectures, cause a CTI, switch between ARG, CTI, and event sheets, and assert the visible menu set and enabled actions match Python's CTI menus.

Done 2026-07-06: Completed active sheet/workflow menu switching for the browser shell. Sheet switches now publish the active sheet type and reachability-only state on the document body, refresh backend-owned menu descriptors, and render descriptor menus into the active sheet's panel instead of the first cloned panel ID. Workflow mode changes now refresh descriptors as well, and event sheets hide analysis-only workflow controls while preserving the File/event surface. Added focused red/green Vitest coverage for sheet menu context, workflow descriptor refresh, and active-sheet descriptor placement, plus hosted-Go Playwright coverage for CTI/reachability/event-sheet visible menu switching. Existing Go descriptor coverage verifies CTI/reachability/event menu payloads. Full `make test` and `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`; the first sandboxed `make test-web` attempt was rerun outside the sandbox because localhost binding was blocked.

[x] ### 37. CTI Used-Relations Display Is Partial

Inventory refs: PLAN383 item 67; PLAN378 sections 20.10 and 33.7.

The frontend can auto-check used relation rows from a failed check result, but Python CTI also has used-relation display and a "relations to minimize" input. The current UI does not expose the full relation-minimization control surface.

TODO: add CTI used-relations and relations-to-minimize UI. Tests should run a failing induction check, verify only relevant relation controls are enabled/checked, edit the relations-to-minimize input, and verify minimize uses that set.

Done 2026-07-06: Added the CTI relations-to-minimize control and wired it through the check/minimize paths. The state/relations pane now exposes the Python-style `relations to minimize` text input. Frontend checks send the edited value as `relations_to_minimize`, CTI Minimize sends the same value with its active sheet, the HTTP check endpoint decodes it, and the Go induction path stores it on `CTIUI.RelationsToMinimize` and passes it into `CheckFinalCond`. Added focused red/green Go coverage for the check option, static/Vitest coverage for the visible input and request payloads, and hosted-Go Playwright coverage proving the browser field is sent to both check and Minimize. Existing used-relation tests still verify only relevant relation rows are auto-checked. Full `make test` and `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`.

[x] ### 38. CTI Diagram Pre/Post State Presentation Is Partial

Inventory refs: PLAN383 item 68; PLAN378 sections 20.6 and 20.11.

The backend routes diagram through `CTIUI.Diagram()` and returns concept payloads, but the browser workflow does not clearly present CTI pre-state/post-state context as Python does. Users need to see which CTI state is being diagrammed and how it affects the concept graph.

TODO: add visible CTI state labels and verify diagram uses the intended pre-state. Tests should create a CTI, run Diagram, and assert the concept graph/facts correspond to the pre-state rather than a stale BMC or post-state value.

Done 2026-07-06: Added explicit CTI pre-state labeling to diagram responses and concept snapshots. CTI-scoped diagram actions now return top-level `cti_state_label` plus concept-level `state_label`/`cti_state_label`, derived from the CTI analysis graph pre-state when available; the older direct `ConceptDiagram` backend route emits the same labels. The browser now sends `ui_mode` with Diagram, prefers CTI labels when rendering concept snapshots, and shows `State: CTI pre-state 0` in the state panel. Added focused red/green Go, Vitest, and hosted-Go Playwright coverage for the backend label, runtime payload, model label preference, and visible browser label. Full `make test` and `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`.

[x] ### 39. CTI Strengthen Confirmation Is Missing

Inventory refs: PLAN383 item 75; PLAN378 section 20.18.

Python asks for confirmation before adding a selected conjecture as a strengthened invariant. The web `cti_strengthen` path can add to the compiled module without a visible confirmation step.

TODO: add a confirm dialog with the exact conjecture text before strengthening. Tests should cancel and accept the dialog, verifying cancel leaves conjectures unchanged and accept appends exactly one new conjecture.

Done 2026-07-06: Added a backend `cti_strengthen_preview` action that derives the exact selected conjecture without mutating CTI state, and routed browser `cti_strengthen` through a read-only confirmation dialog before the mutating action runs. Cancel now leaves conjectures unchanged and accept performs exactly one append using the same conjecture text. Added focused red/green Go coverage for preview-vs-append behavior, Vitest coverage for cancel/accept frontend flow, and hosted-Go Playwright coverage for the visible dialog text and cancel/accept dispatch. Full `make test` and `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`.

[x] ### 40. CTI Save-Invariant File Content Needs Full Section Fidelity

Inventory refs: PLAN383 item 70; PLAN378 section 20.8.

There is backend coverage for kept/dropped-style invariant output, but the browser save path should be checked against the Python sections and file-dialog behavior. This is high value because saved invariants become source artifacts.

TODO: add browser-level tests for CTI Save invariant, including kept, dropped, and new conjectures. Tests should verify the filename, extension, content sections, labels, and formula formatting.

Done 2026-07-06: Added hosted-Go browser coverage for CTI Save invariant using a labeled invariant model that drops one original invariant and adds one new conjecture before saving. The test captures the File System Access picker payload and saved text, verifying the backend filename `invariant.ivy`, `.ivy` picker extension, kept/dropped/new section ordering, preserved labels, commented dropped invariant, and formula formatting. The browser save path now honors the backend-provided filename before falling back to model-derived suggestions. Full `make test` and `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`.

[x] ### 41. CTI Bounded-Check Counterexample Viewing Is Partial

Inventory refs: PLAN383 item 71; PLAN378 section 20.9.

The current bounded check path prompts for a bound and calls the backend, and failed top-level checks can offer trace actions. Python CTI bounded check specifically offers to view a counterexample for the conjecture workflow.

TODO: implement the CTI-specific BMC result path, including view-counterexample affordance. Tests should run a CTI bounded check with a counterexample and assert the user can open the corresponding trace/ARG sheet.

Done 2026-07-06: Completed the CTI bounded-check counterexample affordance contract. CTI BMC results that register a counterexample trace now explicitly return `view: "trace"` along with the trace ARG, sheet id, and label. Added focused red/green Go coverage for that backend result shape and hosted-Go Playwright coverage for the browser flow: entering a bound, seeing the Python-style `View` dialog, switching to reachability mode, and opening the reachability-only trace sheet. Existing Vitest coverage verifies cancel does not open the trace. Full `make test` and `make test-web` passed with `GOCACHE=/tmp/goivy-gocache-codex`.

[x] ### 42. CTI Minimize Needs User-Visible BMC/Core Details

Inventory refs: PLAN383 item 76; PLAN378 sections 20.15 and 33.10.

The backend has a minimize action, but Python's CTI minimize first runs BMC and then reduces selected facts using an unsat core, with user-visible fact/core selection behavior. The current UI hides most of that reasoning behind a single action/status.

TODO: expose enough minimize detail for users and tests to validate the same facts were minimized. Tests should select facts, run minimize, verify the BMC bound used, inspect the resulting conjecture, and verify unselected facts do not appear.

Done: `cti_minimize` now reports the BMC bound, selected/input facts, core/minimized facts, removed facts, and resulting conjecture. The frontend opens a read-only details dialog for the minimize result, and browser coverage verifies the bound/core display while excluding unselected facts. Added focused Go, Vitest, and Playwright regressions; `make test` and `make test-web` pass with `GOCACHE=/tmp/goivy-gocache-codex`.

[x] ### 43. CTI Check-Sufficient And Check-Relative-Induction Result Dialogs Are Partial

Inventory refs: PLAN383 items 77 and 78; PLAN378 sections 20.16 and 20.17.

The backend returns ok/message for these checks, and the frontend can call the actions. Python presents these checks as conjecture workflow results tied to the selected conjecture.

TODO: make the result display explicit and tied to the active selected facts/conjecture. Tests should run both actions with sufficient, insufficient, inductive, and non-inductive selections and assert the displayed result text and backend state.

Done: CTI sufficient and relative-induction actions now return explicit check/result metadata, selected conjecture, target conjecture, and dialog text. The frontend displays read-only result dialogs for sufficient, insufficient, inductive, and non-inductive outcomes. Added focused Go, Vitest, and hosted-browser regressions; `make test` and `make test-web` pass with `GOCACHE=/tmp/goivy-gocache-codex`.

## P2: Event Viewer And Trace Display

[x] ### 44. Trace Serialization Display Is Missing

Inventory refs: PLAN378 sections 24.1 through 24.5 and 27.2.

The web UI can load event traces and show check-result details, but Python trace display has serialization-to-lines behavior, infinite-loop detection, function-call formatting, trace line numbers, and a detailed option. Those formatting choices are not exposed as a trace display feature in the web UI.

TODO: port trace formatting or explicitly fold it into the event trace viewer. Tests should feed traces with repeated states, function calls, detailed and non-detailed output, and line numbers, then verify the displayed text matches the Python behavior.

Done: event trace sheets now carry compact and detailed serialized trace text alongside the tree view. Trace events accept optional line, kind, state, state-key, and loop metadata; backend formatting covers Python-style compact call lines, detailed line/state blocks, and infinite-repeat markers. The browser renders a Trace pane with a Detailed toggle. Added focused Go and hosted-browser regressions; `make test` and `make test-web` pass with `GOCACHE=/tmp/goivy-gocache-codex`.

[x] ### 45. Event Viewer Launch As A Standalone Tool Is Missing

Inventory refs: PLAN383 item 79; PLAN378 sections 25.1 and 25.2.

The web UI can open event trace sheets from the main app. Python also has an event viewer launched as its own Tix tree UI/notebook, including `ivy_show.py` workflows.

TODO: decide whether `ivyweb` should support an event-viewer-only launch mode. Tests should open a `.iev` file directly into the intended web surface and verify no model session is required if standalone mode is in scope.

Done: `ivyweb` now supports an event-viewer-only launch mode through query parameters, opens `.iev` traces into an event sheet without restoring or requiring a model session, hides the default model sheet, and has focused Playwright coverage plus green `make test`/`make test-web`.

[x] ### 46. Event Tree Selection Style And Keyboard Ergonomics Are Partial

Inventory refs: PLAN378 section 25.7.

The web event tree supports lazy child expansion, filtering, find forward/reverse, and saved patterns. Tk HList selection style, keyboard traversal, and selection affordances are not fully represented.

TODO: define the required event tree interaction details for the web port. Tests should cover row selection, expansion/collapse, keyboard navigation if supported, preserved selection after filtering/finding, and visible selected-row styling.

Done: event trees now expose tree/treeitem roles, ARIA selected/expanded state, roving focus, visible selected-row styling, click focus, arrow-key navigation, keyboard expand/collapse, and selected-row preservation when filtering retains the selected event. Focused Playwright coverage was added and `make test`/`make test-web` are green.

[x] ### 47. Event Pattern File Dialog Semantics Need Coverage

Inventory refs: PLAN383 item 82; PLAN378 sections 25.5 and 25.6.

Pattern add/remove/load/save/clear exists, but save/load currently use web dialogs/downloads rather than Python file dialogs. The exact persistence behavior and malformed-pattern handling should be nailed down.

TODO: add end-to-end tests for loading and saving pattern files from the browser shell. Tests should verify newline handling, invalid patterns, duplicate patterns, selected pattern preservation, and backend-authoritative state after failed operations.

Done: browser coverage now drives the event pattern Load and Save buttons, including malformed pattern load failures, CRLF/newline input, duplicate pattern preservation, selected option preservation after backend-authoritative list refresh, and File System Access picker save content/metadata. Failed loads/saves now report status errors without mutating the pattern list, and `make test`/`make test-web` are green.

## P2: Notebook/Widget Workflows And Extensions

[x] ### 48. Analysis-Session History Navigation Is Missing

Inventory refs: PLAN383 item 90; PLAN378 sections 21.1 through 21.4 and 33.2.

Python notebook widgets expose first/prev/next/last history navigation, step info display, active-element auto-click, and modal messages. The current web UI has sheet switching and tutorial URL history, but not analysis-session step history.

TODO: implement or exclude analysis-session history navigation for web UI. Tests should drive a multi-step analysis session, navigate back and forward, and verify the ARG/concept/transition widgets reflect the selected history step.

Done: Go `AnalysisSessionWidget` history navigation now clamps to Python-style first/previous/next/last bounds, and browser analysis sheets now expose a history strip with first/prev/next/last controls plus step/transition info. Browser snapshot history restores ARG and concept graph payloads for the selected step. Focused Go and Playwright coverage was added, and `make test`/`make test-web` are green.

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

**Go/Web### 55. Text Fitting And Wrapping Need Semantic Regression Tests

Inventory refs: PLAN378 sections 32.16 and 32.17.

Python graph rendering sets wrapping and font sizes for ARG and concept nodes. The web style wraps text, but node sizes are heuristically derived and can diverge for long action/concept labels.

TODO: add graph-render tests for long state labels, relation labels, concept labels, and multi-line edge labels. Tests should assert text remains visible and does not overlap enough to hide the graph semantics.

## P3: Source, Files, And Recent State

[ ] ### 56. File Browser Raise/Re-use Semantics Are Missing

Inventory refs: PLAN383 item 20; PLAN378 sections 23.1 through 23.5.

The web UI has one editor and can highlight a line in it. Python has a separate source browser window that is reused, raised, and line-highlighted independently.

TODO: if source browsing remains in scope, implement a source browser surface with reuse and raise semantics. Tests should open source for two different edges and verify the same browser surface updates and remains separate from the editable model.
