# Tcl GUI Behavior Inventory

Codex version.

This inventory is based on a source-reading pass over the Python Ivy UI code under `~/ivy/pyivy/ivy/ivy`, with the likely-UI file list used as the starting spine and the remaining `.py` files checked for UI entry points, display behavior, callbacks, launch behavior, and UI-adjacent helpers. The most relevant Tcl/Tk implementation files are `tk_ui.py`, `tk_graph_ui.py`, `tk_cy.py`, `ivy_ui.py`, `ivy_ui_cti.py`, `ivy_graph_ui.py`, `ivy_ui_util.py`, `ivy_ev_viewer.py`, `ivy_init.py`, `ivy_check.py`, `ivy_art.py`, `ivy_graph.py`, `dot_layout.py`, `cy_elements.py`, `cy_render.py`, and the older IPython widget UI files that document parallel behavior.

## Launch, Mode, And Top-Level Window Behavior

### 1. `ivy.py` launches the Tk UI around `ivy_init()`

The script installs the default `SIGINT` handler, opens an Ivy module context, compiles or loads the requested Ivy artifact through `ivy_init()`, and then calls `tk_ui.ui_main_loop` with the resulting analysis graph. This makes the default app a single Tk event loop whose initial content is the compiled analysis graph. The UI must therefore start after command-line parameter parsing and source loading, not before.

To test this, run the GUI launcher on a small `.ivy` file and assert that a window titled `ivy` appears. Verify that the first displayed analysis graph reflects the compiled file rather than an empty placeholder. Send an interrupt signal while the GUI is running and verify it uses the process default signal behavior rather than swallowing it silently.

### 2. Command-line parameters select UI behavior before compilation

`ivy_init.read_params()` consumes leading `key=value` arguments, applies them through `ivy_utils.set_parameters`, and removes them from `sys.argv` before file processing. The `ui` parameter defaults to `cti`, maps `ui=art` to `ivy_ui`, and maps other values to modules named `ivy_ui_<value>`. The selected UI module can contribute `compile_kwargs`, which are merged into source loading; the CTI UI sets `{'ext': 'ext'}`.

To test this, launch with `ui=art` and confirm the menu set is the ART-oriented File/Mode/Action UI rather than the CTI invariant UI. Launch with `ui=cti` and confirm the CTI menus and the initial concept graph are shown. Use an invalid parameter value and assert that startup reports the parameter error and exits before creating the main window.

### 3. `.a2g` analysis-state files can be loaded before source files

`ivy_init.ivy_init()` accepts an optional `.a2g` file first, unpickles it as an analysis graph, restores the module and signature from the graph domain, and then optionally loads a `.ivy` or `.dfy` source file to update the graph module. If no `.a2g` file is supplied, it creates a fresh analysis graph through `ivy_new()`. This is part of the UI contract because the File/Save behavior writes `.a2g` files for later GUI sessions.

To test this, save an analysis graph from the GUI, restart the GUI with the `.a2g` file, and verify the same states, transitions, and covering edges appear. Then launch with both `.a2g` and `.ivy` arguments and verify the graph can still execute actions from the updated module. Also verify that a non-`.a2g`, non-`.ivy`, non-`.dfy` argument prints usage instead of opening the GUI.

### 4. The root window title and palette are fixed

When `TkUI` creates its own root, it uses `tkinter.tix.Tk()`, sets the Tk palette background to white, and sets the window title to `ivy`. If an existing Tk root is supplied, a new `Toplevel` is created for the UI frame instead. The port should preserve the distinction between owning the root and embedding in another root.

To test this, start the standalone UI and check the window title and default white background. Embed the UI in a test Tk root and verify it creates a child top-level rather than replacing the root. Confirm that the notebook and all graph panes still attach to the supplied frame when embedding.

### 5. The main UI is a notebook of analysis tabs

`TkUI` creates a `tix.NoteBook` that fills the root frame and owns one tab per analysis graph. Each call to `add` increments a tab counter, names tabs as `sheet_N` by default, labels them as `Sheet N`, and raises the newly added tab. The UI quits automatically when the last tab is removed.

To test this, add two analysis graphs and verify two tabs labeled `Sheet 1` and `Sheet 2` appear with the second raised. Remove one tab and check that the other remains active. Remove the last tab and assert the event loop exits.

### 6. Each analysis tab is split horizontally into ARG and state/concept graph panes

`TkUI.add` constructs a horizontal `tix.PanedWindow` with a left frame for the analysis graph and a right frame for state/concept graphs. The left frame has horizontal and vertical scrollbars wired to the ARG canvas. The right frame is reused when the user views states or opens concept graphs associated with the selected analysis tab.

To test this, launch a graph large enough to need scrolling and verify the left canvas scrollbars move the ARG. Click or otherwise view a state and verify the state/concept graph appears in the right pane, not in a separate unmanaged root. Resize the paned window and verify both panes remain visible and scroll regions are updated.

### 7. UI classes are mixed into Tk classes dynamically

`tk_ui.new_ui()` gets the selected Ivy UI class, creates a dynamic class combining `TkUI` with that Ivy UI class, instantiates it, and stores it in the global `ivy_ui.ui`. `TkUI.add` similarly combines `TkAnalysisGraphWidget` with the selected analysis graph UI class. This means port behavior should be interface-based: the UI objects are both toolkit widgets and Ivy action providers.

To test this, select `ui=art` and verify methods from `ivy_ui.AnalysisGraphUI` are callable from the rendered Tk widget. Select `ui=cti` and verify the graph widget uses `ivy_ui_cti.AnalysisGraphUI` and its `ConceptGraphUI`. Add a decomposed edge subgraph and verify it uses the requested UI class override.

### 8. The `mode` parameter and Mode radio menu control abstraction behavior

The ART UI exposes `Concrete`, `Abstract`, `Bounded`, `Induction`, and `Pdr` radio choices backed by `ivy_ui.default_mode`. The selected mode changes the initial abstraction operator and later action execution: concrete uses no abstraction, abstract and induction use `ivy_alpha.alpha`, pdr uses `ivy_alpha.predicate_alpha`, and bounded uses `top_alpha`. Historical behavior intentionally initializes abstract mode with a concrete initial state.

To test this, switch each mode before creating or executing states and verify the resulting state clauses match the corresponding abstraction. Check that the selected radio value persists while using node commands. Verify that bounded and induction modes route safety checks through bounded safety, while other modes use local safety.

## Menu Bars, Dialogs, And Error Handling

### 9. Menu bars are declarative and rendered with `Menubutton`

Classes inheriting `WithMenuBar` define `menus()` as tuples for menus, buttons, separators, and radio groups. `WithMenuBar` renders a top `MenuBar` frame, adds each menu as a `Menubutton`, converts buttons to menu commands, separators to separators, and radio groups to `StringVar`-backed radio menu items. Unknown menu item types assert.

To test this, inspect the File, Mode, Action, Invariant, Conjecture, View, and Events menus in the contexts where they should appear. Verify separators and radio selections match the declarative order. Add a test menu entry with an unknown type in a controlled fixture and verify it fails rather than being silently ignored.

### 10. `RunContext` reports Ivy errors as blocking dialogs

The Tk `RunContext` catches `iu.IvyError`, opens a `Toplevel` dialog showing `repr(exc_val)`, adds an `OK` button, waits for the dialog to close, and suppresses the exception. Non-Ivy exceptions are not swallowed. The utility `RunContext` variant also switches the cursor to busy before the operation and restores it afterward.

To test this, trigger a known Ivy error from a menu action and verify an OK dialog appears and the GUI remains usable afterward. Trigger a non-Ivy exception in a test-only action and verify it propagates to the test harness rather than appearing as a normal Ivy error dialog. For the utility context, assert the cursor changes to `watch` during a long operation and returns to normal.

### 11. Dialog answers can be pre-seeded for tests

`TkUI.answer(string)` pushes expected dialog answers into `self.answers`, and `getans()` pops from that list for OK/text/OK-cancel dialogs that support automated invocation. If an answer is present, `text_dialog`, `ok_cancel_dialog`, and `ok_dialog` invoke the matching button immediately and return without waiting. This behavior is a built-in testing hook.

To test this, push `"OK"` before an OK dialog and assert no modal wait blocks the test. Push `"Cancel"` before an OK/cancel dialog and verify the cancel callback runs. Push a custom command label used by `text_dialog`, such as `"Refine"` or `"View"`, and verify that command is invoked.

### 12. OK dialogs are centered and modal

`ok_dialog` creates a `Toplevel`, displays a message label and an `OK` button, centers it over the parent window, and waits for the window to close. The dialog is simple but important: many verification outcomes use it for terminal states such as "Cannot reverse" or "PDR terminated". The port should preserve both the blocking interaction and the exact button semantics.

To test this, invoke an action that calls `ok_dialog` and assert focus remains blocked until OK is clicked. Verify the dialog appears centered relative to the active UI frame. Confirm that closing it returns to the previous graph state without triggering any extra graph update.

### 13. OK/cancel dialogs run callbacks only for OK by default

`ok_cancel_dialog` displays a message with `OK` and `Cancel` buttons. The OK button runs the supplied command before closing; the Cancel button runs the supplied cancel callback, which defaults to a no-op. Safety failures and trace-view prompts rely on this behavior.

To test this, trigger a bounded safety failure and choose OK; verify the error trace graph is added. Repeat and choose Cancel; verify no new graph is added. Confirm both buttons close the dialog and restore focus to the same tab.

### 14. Text dialogs show read-only-looking text plus configurable command labels

`text_dialog` displays a label, a scrollable `Text` widget of height 4 and width 100, an action button whose label defaults to `OK`, and optionally a `Cancel` button. Several workflows change the action label to `Refine`, `View`, or `Remember`. The text is inserted into a normal Tk `Text`, so selection and scrolling are available.

To test this, open an interpolant refinement dialog and verify the button label is `Refine`. Open a BMC counterexample dialog and verify the action label is `View`. Feed long text and assert the vertical scrollbar scrolls the text without resizing the dialog unexpectedly.

### 15. Entry dialogs focus the entry and bind Return

`entry_dialog` shows a message, an entry field, optional initial value, an action button, and optional cancel. It focuses the entry and binds the Return key to run the action with the current entry contents. Add-relation, remember-goal, pattern-search, and integer dialogs build on this primitive.

To test this, open Add relation and verify the cursor is in the entry field immediately. Type a relation and press Return; assert it behaves the same as clicking `Add`. Open an entry dialog with an initial value in a test harness and verify the field is pre-populated.

### 16. Integer dialogs validate integer and range constraints

`int_dialog` wraps `entry_dialog` and converts the submitted string to an integer. Invalid integers and out-of-range values raise `IvyError` with messages describing the entered value. BMC bound selection uses this dialog and passes `minval=0`.

To test this, start bounded conjecture checking and enter a valid nonnegative bound; verify BMC runs for that bound. Enter a non-integer string and verify an Ivy error dialog appears. Enter a negative value when `minval=0` and verify the out-of-range message appears and no BMC run is started.

### 17. Listbox dialogs pass selected indices, not item text

`listbox_dialog` shows a message, a scrollable `Listbox`, an OK button, and optional Cancel button. When OK is clicked, it converts the selected rows to integers, destroys the dialog, and calls the command with either the single selected index or a list of indices when `multiple=True`; if nothing is selected, no command is called. Although it accepts `multiple=True`, the listbox is constructed with `selectmode=SINGLE`, a quirk the port should account for deliberately.

To test this, open a conjecture-selection dialog, choose the second item, and verify the command receives index `1`. Click OK with nothing selected and verify no command runs. Exercise a multi-select caller such as Weaken and decide whether to match the original single-select quirk or intentionally fix it with a regression note.

### 18. Button-list dialogs always include Cancel

`buttons_dialog_cancel` displays a message, one button per `(label, command)` pair, plus a `Cancel` button. Each action button runs its command and closes the dialog. The `TkUI.buttons_dialog_cancel` wrapper currently passes `on_cancel=lambda:None` instead of the caller's cancel callback, preserving a bug/quirk where caller-provided cancel behavior is ignored.

To test this, trigger local safety failure where multiple remedial buttons may be shown and verify each named button runs the expected action. Press Cancel and verify it only closes the dialog. Add a test-specific cancel callback and confirm whether the port intentionally preserves or fixes the ignored-cancel behavior.

### 19. Save-as dialogs use specific titles and file filters

The UI uses `asksaveasfile` for saving analysis states, abstractions, invariants, pattern files, and generated files, with specific titles and file type filters. Analysis saves use `analysis files`/`.a2g`, abstraction and invariant saves use `ivy files`/`.ivy`, graph export uses `dot files`/`.dot`, and event patterns use `event pattern files`/`.pats`. Canceling returns a false value and leaves state unchanged.

To test this, open each save action and verify the title and default file-type filter. Cancel the dialog and assert no file is created and no error is reported. Save to a temporary path and inspect the file contents for the expected format.

### 20. Source browsing reuses one file browser window

`TkUI.browse` creates a single `FileBrowser` top-level on first use, then reuses it for later source locations while the window still exists. `FileBrowser.set` reloads file contents only when the filename changes, highlights the requested line in red, scrolls it into view, and lifts the browser window. The browser uses a `Text` widget with a vertical scrollbar and dimensions 20 lines by 100 columns.

To test this, invoke View Source on two transitions from the same file and verify the same browser window updates the highlight without reopening. Invoke a second source in a different file and verify the text content is replaced. Confirm the highlighted line is visible and red after each browse call.

## Analysis Graph Tab Behavior

### 21. Startup initializes an empty ARG only once

`AnalysisGraphUI.start` initializes the analysis graph when `self.g.states` is empty, choosing the abstraction through `init_alpha()`, then rebuilds the display. If states already exist, it only rebuilds. This matters when loading saved `.a2g` graphs or adding decomposed subgraphs.

To test this, load a fresh Ivy file and assert exactly one initial state is created. Load a saved graph with several states and verify startup does not add another initial state. Add a decomposed edge tab and verify its existing states are preserved.

### 22. ARG node color represents safety status

`AnalysisGraphUI.node_color` returns green when the node has a truthy `safe` attribute and black otherwise. `TkAnalysisGraphWidget.update_node_color` applies that color to the outline of the node shape items tagged for the state. Safety checks therefore produce a persistent outline color update without requiring a full graph rebuild.

To test this, run a safety check that succeeds and verify the checked node outline turns green. Run or simulate a failing check and verify the outline is black. Rebuild the graph and verify the visual status remains consistent with the node's `safe` attribute.

### 23. ARG left-click directly views a state

For ART ARG nodes, left-click actions are the special `("<>", self.view_state)` action, and `TkCyCanvas.make_popup` immediately invokes a sole `<>` action rather than showing a popup. This means left-clicking a state opens or updates the concept/state graph directly. Right-clicking a node opens the execution and node-operation menu.

To test this, left-click an ARG node and verify no context menu appears. Confirm that the right pane shows the node's state graph. Right-click the same node and verify a popup menu appears with execution actions and node commands.

### 24. ARG right-click node menu lists executable actions before node commands

The node context menu starts with a disabled label `Execute action:`, a separator, one entry per state action sorted by `state_equation_label`, another separator, then fixed commands: Check safety, Extend, Mark, Cover by marked, Join with marked, Try conjecture, Try remembered goal, and Delete. The state-action labels show `label` for unlabeled equations or `label -> action` when an action component is present. Commands are functions of the clicked node.

To test this, right-click a node and verify the popup order and separators. Add multiple public actions and assert the action entries are sorted by their display label. Click each fixed command in a controlled graph and verify it receives the node that was clicked.

### 25. ARG edge right-click menu supports transition operations

For ART ARG edges, right-click offers Dismiss, Recalculate, Step in, and View Source. Dismiss is a no-op; Recalculate recomputes the target state for that transition; Step in decomposes the edge into a new analysis tab when possible; View Source browses the action source line if the action has `lineno`. Left-clicking ARG edges has no action.

To test this, right-click a transition edge and verify the four menu entries. Choose Recalculate and verify the target state clauses update and the graph rebuilds. Choose Step in for a decomposable action and assert a new tab appears; choose it for a non-decomposable action and verify an Ivy error dialog.

### 26. Marking an ARG node fills it red

`mark_node` clears the previous mark by calling `show_mark(False)`, stores the clicked node in `self.mark`, and calls `show_mark(True)`. `TkAnalysisGraphWidget.show_mark` fills the marked node's shape red when on and clears the fill when off. Other operations such as Cover by marked and Join with marked use the stored mark.

To test this, mark node A and verify it fills red. Mark node B and verify A is no longer red while B is red. Use Cover by marked and Join with marked after marking and verify they use the marked node rather than the currently selected or clicked node.

### 27. Cover by marked reports success or failure and rebuilds

`cover_node` attempts to cover the clicked node by the marked node. It prints an attempt message, calls `g.cover`, raises `IvyError("Covering failed")` when coverage fails, and rebuilds on success. Successful coverage adds a dashed cover edge to the ARG render.

To test this, create a graph where one state covers another, mark the covering state, and invoke Cover by marked on the covered state. Verify a dashed cover edge appears and no error dialog is shown. Repeat with a non-covering state and verify an Ivy error dialog appears and no cover edge is added.

### 28. Join with marked creates a join successor

`join_node` retrieves the marked node and joins it with the clicked node using the current abstraction operator. The resulting joined state is added to the ARG through `AnalysisGraph.join`, and the display rebuilds. If no marked node is available, the method silently returns.

To test this, mark one node, invoke Join with marked on another node, and verify a new state with incoming join edges appears. Confirm the join transition labels are `join`. Invoke the command with no mark in a test setup and verify no crash and no graph change.

### 29. Delete removes a state and dependent descendants

`AnalysisGraph.delete` marks the requested state by setting `id=-1`, also marks later states whose dependencies include a deleted state, then compacts states, transitions, and covering edges. `AnalysisGraphUI.delete_node` calls this and rebuilds. IDs are renumbered after removal.

To test this, delete a leaf node and verify only that node and its incident edges disappear. Delete an intermediate node and verify dependent descendants also disappear. Confirm remaining node labels are renumbered from zero and no transition points to a deleted state.

### 30. Extend searches for a non-covered extension

`find_extension` asks `g.state_extensions(node)` for the next available extension, performs that state action, and views the new state. If there are no extensions, it shows `State <id> is closed.` in an OK dialog. The extension search uses the fixed-point candidate over uncovered states.

To test this, run Extend on a node with an available action and verify a successor state is added and displayed. Run Extend on a closed node and verify the exact closed-state dialog appears. Confirm the action chosen matches the next yielded extension from the analysis graph.

### 31. Recalculate all processes each target state once

`recalculate_all` iterates through transitions and calls `recalculate_edge` for transitions whose target state ID has not already been processed. This avoids recalculating the same target multiple times in graphs with multiple incoming transitions. Each edge recalculation rebuilds the display.

To test this, create a graph with a join target that has multiple incoming transitions. Invoke Recalculate all and count recalculation calls per target; each target should be recalculated once. Verify the final graph display matches individually recalculating every unique target.

### 32. Recalculate state updates current concept graph state

Concept graph Recalculate calls the parent analysis UI's `recalculate_state` on the parent ARG node, then sets the concept graph state to the parent state's new clauses and updates. This is available from the concept graph Action menu. It depends on the concept graph having a parent state.

To test this, view a state, mutate or execute an action that makes recalculation observable, and click Recalculate in the concept graph. Verify the parent ARG node's clauses update and the concept graph redraws with the new facts. Invoke it on a graph with no parent state and verify it is a no-op.

### 33. Show reachable states opens or reuses a reachable-tree graph

`reachable_tree` is lazily created as a concrete reachability `AnalysisGraph` and initialized with the first under-approximation state. `show_reachable_states` adds that graph as a new tab. Later reachability operations add states and transitions to this tree.

To test this, invoke Show reachable states and verify a new tab appears. Run a successful one-step reach and verify the reachable tree gains the reached state and transition. Invoke Show reachable states again and verify it uses the same stored tree object rather than resetting progress.

### 34. Saving analysis state pickles the graph as protocol 2

The File/Save command opens `Save analysis state as...` and writes `self.g` through `pickle.dump(..., protocol=2)`. It closes the chosen file and does nothing if the user cancels. The saved file is the input format handled by `.a2g` startup.

To test this, save an analysis state and verify the file can be loaded by `ivy_init` as a `.a2g`. Inspect the first bytes or load behavior to confirm it is a pickle, not textual Ivy. Cancel the save dialog and verify no graph mutation or file write occurs.

### 35. Saving abstraction writes concept declarations

The File/Save abstraction command writes one `concept <name> = <space>` line per concept space in the graph domain. It uses the current abstract domain, not the full analysis graph. The output is a textual `.ivy` fragment.

To test this, refine the domain with a new concept and save the abstraction. Verify the file contains a `concept` line for the new concept. Load or parse the saved fragment in a suitable Ivy context and verify the concept syntax is valid.

## Concept Graph Rendering And Controls

### 36. Viewing a state reuses the current concept graph when possible

`view_state` uses the clicked state's clauses unless an override is supplied. If `current_concept_graph` already exists, it calls `set_parent_state` on it and returns; otherwise it builds a standard concept graph from the analysis graph and displays it in the right pane. A `reset` flag controls whether concepts are reset to defaults.

To test this, left-click node 0 and then node 1 and verify only one concept graph widget is reused. Confirm the parent-state label and displayed facts change to node 1. Repeat with `reset=True` through action execution and verify custom concepts are removed.

### 37. Concept graph has its own Action and View menus

The base concept graph menu includes Action commands Undo, Redo, PDR step, Concrete, Gather, Reverse, Path reach, Reach, Conjecture, Backtrack, Recalculate, Diagram, Remember, and Export. The View menu includes Add relation. These commands operate on the current graph stack and parent analysis state.

To test this, open a state concept graph and verify all Action entries appear in order. Verify View/Add relation opens the relation entry dialog. Disable or remove the parent state in a test and assert parent-dependent commands do not crash.

### 38. Concept graph undo, redo, and backtrack use graph stack checkpoints

Before graph-changing operations, `GraphWidget.checkpoint` stores a copy of the current graph and clears the redo stack. Undo swaps the current graph with the last undo snapshot, redo reverses that, and both reverse-sync checkbox states before updating. Backtrack repeatedly undoes to the most recent graph snapshot marked with `backtrack_point`, removes that marker, and updates.

To test this, materialize a node, then Undo and verify the materialization disappears and checkboxes match the restored graph. Redo and verify it reappears. Run Reverse to create a backtrack point, make additional changes, invoke Backtrack, and verify the graph returns to the marked state.

### 39. Relation checkboxes have four columns: `+`, `?`, `-`, and `T`

`tk_graph_ui` creates a right-side legend with a state label and a scrollable `tix.ScrolledHList`. The header row contains labels `+`, `?`, `-`, and `T`, and each relation row has colored checkbuttons for true/all-to-all, unknown, false/none-to-none, a colored relation label, and a transitive checkbox. Changing any checkbox calls `gw.update()`.

To test this, view a concept graph and verify the legend row layout and labels. Toggle each checkbox for a relation and verify the graph redraws to include or exclude the corresponding facts or edges. Toggle `T` on a transitive relation and verify transitive reduction affects displayed edges.

### 40. Relation colors cycle through a fixed palette

`tk_graph_ui.line_colors` defines a fixed list of Tk colors used for relation rows and edge colors. Node outlines are colored by sort, and relation edges by relation index. The palette cycles modulo the list length.

To test this, create a graph with more relations than the palette length and verify colors cycle. Check that a relation row label and that relation's edges use the same color. Verify nodes of the same sort share outline color and different sorts get different colors until the palette cycles.

### 41. Concept graph node styles encode cardinality

Concept graph nodes use styles based on classes: `at_least_one` is width 4 plus a double outline gap, `at_most_one` is width 2, `exactly_one` is width 4, and `node_unknown` is width 2 plus a double outline gap. Non-existing nodes are not drawn in the Tk renderer. Fill is normally empty, and outline color is the sort color.

To test this, create states producing each node cardinality class and verify the outline styling differs as specified. Confirm non-existing nodes do not appear. Select or mark a node and verify selection fill overlays the existing outline style without changing relation colors.

### 42. Concept graph edge styles encode truth status

Edges classified `none_to_none` are dashed with width 2, `all_to_all` are solid with width 2, and `edge_unknown` are short-dashed with width 2. All edges have black or relation-colored fill depending on widget type and use the arrowshape `"14 14 5"`. Cover edges in the ARG use a separate dashed style.

To test this, toggle `+`, `?`, and `-` for a relation with known true, unknown, and false statuses and verify line style changes. Verify arrowheads appear on directed edges. Compare ARG cover edges to concept graph negative edges to ensure the styles are not conflated.

### 43. Graph layout is produced by DOT and transformed to Tk coordinates

`dot_layout` sends nodes and edges to Graphviz `dot`, gets node positions, dimensions, edge splines, labels, and optional subgraph boxes, then flips y coordinates to a stable top-left origin. Transitive edges are used to order nodes top-to-bottom, clusters group nodes by sort, and back edges may be reversed before layout. Tk rendering uses these positions directly.

To test this, render a graph with multiple sorts and verify nodes cluster by sort and do not jump vertically when the graph's height changes. Render a transitive order and verify order-related edges tend to align top-to-bottom. Turn on subgraph boxes and verify the calculated boxes surround the clustered nodes.

### 44. Edge label text is post-processed to restore braces and newlines

ARG transition labels replace `{` with `-[`, `}` with `]-`, and newlines with `\l` before DOT layout so Graphviz preserves them. `TkAnalysisGraphWidget.rebuild` and `tk_cy.get_label_text` restore `-[` to `{`, `]-` to `}`, and `\l` to newline in rendered text. This preserves readable multi-line action labels.

To test this, create an action label containing braces and multiple lines. Verify the rendered edge label shows literal braces and line breaks. Export or inspect the intermediate DOT behavior only in a low-level test; the user-facing canvas should never show the escape digraphs.

### 45. Concept graph constraints are displayed as selectable text under the graph

If `g.constraints` is not true, `TkGraphWidget.rebuild` adds a `Constraints:` label below the graph and one text item per conjunct. Each constraint starts selected, is tagged `cnst<idx>`, and left-clicking toggles it between black and grey. Active constraints feed `get_active_facts()`.

To test this, open a goal with multiple constraints and verify they appear below the graph. Click one constraint and verify it turns grey and is omitted from `get_active_facts()`. Click it again and verify it turns black and is included.

### 46. Node and edge selections are visual grey overlays

Concept graph node selection toggles the node ID in `node_selection` and fills the shape grey. Edge selection stores the tuple of concept IDs and uses `highlight_edge` to draw or remove a thick grey line under the edge spline. Selection state is cleared on rebuild.

To test this, in CTI concept graph mode left-click a node and verify it fills grey. Left-click an edge and verify a thick grey highlight appears beneath it. Rebuild the graph and verify selections clear unless explicitly re-highlighted by a workflow such as gather/minimize.

### 47. Add relation parses user text into a new concept

Add relation opens an entry dialog with `Add a relation [example: p(X,a,Y)]:`. When submitted, it extends the graph signature with constants from the current state and constraints, parses the text into a formula, creates a concept from it, adds it to the domain, updates relation controls, and redraws. Parser errors are caught through `RunContext`.

To test this, add a valid unary or binary relation and verify a new relation row appears in the legend. Toggle its checkboxes and verify the graph can show the derived labels or edges. Enter invalid syntax and verify an Ivy error dialog without adding a row.

### 48. Gather replaces current constraints with visible definite facts

The base concept graph Gather command checkpoints, collects facts for all visible relations using the displayed true/false values, stores them as graph constraints, and updates. It ignores unknown facts and uses the relation checkbox state as its projection. This turns a visual graph pattern into a proof goal.

To test this, toggle a small set of positive and negative relation boxes, click Gather, and verify constraints appear below the graph. Toggle off a relation and gather again; verify facts from that relation are absent. Use Undo to verify the previous constraint set is restored.

### 49. Concrete adds concrete state facts to the graph state

The Concrete command checkpoints the current concept graph, appends `g.concrete` to `g.state`, and updates. The parent analysis graph sets `concrete` to an empty list in the standard graph path, but other graph construction paths can populate it. This is a visual refinement action rather than a solver query by itself.

To test this, create or mock a graph with non-empty concrete clauses and run Concrete. Verify the state formula includes the concrete clauses and the graph redraws with extra facts. Run Undo and verify the concrete clauses are removed.

### 50. Split with a unary relation refines a node into subnodes

Left-clicking a concept graph node opens actions including `Split with...` followed by each unary relation whose variable sort matches the node sort. Choosing one checkpoints the graph, splits the node concept by that relation, shows the relation label, and updates. The underlying domain replaces the original node with positive and negative split concepts.

To test this, create a unary predicate over the node sort and split a node by it. Verify the original node is replaced by two split nodes. Confirm the splitting relation's node-label checkbox is enabled and the graph redraws.

### 51. Splatter splits a node by known constants

The Splatter node action checkpoints and splits the clicked node using an enum concept built from constants in the current constraints unless explicit constants are supplied. It creates equality concepts for each constant and splits the node by that set. The resulting display separates possible witnesses by constant equality.

To test this, open a graph whose constraints mention several constants and invoke Splatter on a node of the matching sort. Verify the node is split into one child per constant equality. Check that constants of unrelated sorts do not create ill-sorted visible children.

### 52. Empty asserts that a node concept has no witnesses

The Empty node action checkpoints, adds a universal negation of the node concept to `suppose_constraints`, and updates. Empty edge similarly materializes a negative edge fact through the materialization path. These actions are proof-goal refinements rather than deletions from the graph domain.

To test this, invoke Empty on a node and verify the graph constraints include a no-witness condition. Verify the node class changes to non-existing or otherwise reflects emptiness after recomputation. Undo and verify the node and constraints return.

### 53. Materialize node creates or reuses a witness

Materialize checks whether the unary concept already has a witness constant implied by its formula; if so it uses the first one. Otherwise it creates a fresh constant, adds an equality concept for it, splits the node by that equality concept, supposes the concept holds for the witness, and returns the witness concept. The UI enables the witness label and redraws.

To test this, materialize a node whose formula already pins it to a known constant and verify no extra fresh constant is introduced. Materialize a generic node and verify a fresh equality label appears. Confirm the witness label checkbox is enabled and the graph shows the witness.

### 54. Materialize edge can assert positive or negative facts

Materialize edge creates witnesses for source and target node concepts, reusing the same witness when source and target are the same node. It supposes either the edge concept or its negation, then enables witness labels and the relevant edge checkbox. Dematerialize is implemented as materialize edge with `truth=False`.

To test this, materialize a positive edge and verify witness nodes and a positive edge become visible. Dematerialize the same edge and verify a negative fact is added and the `-` display can show it. Materialize a self-edge and verify only one witness constant is used for both endpoints.

### 55. Materialize edge from selected node prompts for a matching binary relation

`materialize_from_selected` requires a marked concept graph node, computes the source and target sorts from the mark and clicked node, filters relations whose sorts match exactly, and opens a listbox with their labels. The selected relation is then materialized as an edge from the marked node to the clicked node. If no mark exists, nothing happens.

To test this, select/mark node A, invoke Materialize edge on node B, and verify the dialog lists only binary relations with sorts `(sort(A), sort(B))`. Select a relation and verify the edge materializes from A to B. Invoke the command without a marked node and verify no dialog appears and no crash occurs.

### 56. Add projection creates binary concepts from ternary relations and witnesses

When a node has a witness, `get_projections` finds ternary concepts containing a variable of the witness sort and substitutes the witness to create a binary projection. Node menus include `Add projection...` entries for available projections. Choosing one adds the projected concept to the graph domain and relation controls.

To test this, create a ternary relation and materialize a node to provide a witness. Open the node menu and verify projection entries appear. Add a projection and verify a new binary relation row and corresponding edges can be displayed.

### 57. Remember stores a copy of the current goal under a user name

Remember opens an entry dialog titled `Enter a name for this goal:` and stores `self.g.copy()` in the parent analysis UI's `remembered_graphs` map. Try remembered goal lists stored names, copies the chosen graph, sets its parent state to the clicked ARG node, sets its state to the node clauses plus stored constraints, appends it to `state_graphs`, and displays it. The remembered graph is detached by copy so later edits do not mutate the stored template.

To test this, create a goal, remember it as `g1`, and verify `Try remembered goal` lists `g1`. Choose it on a different ARG node and verify the displayed goal uses that node as parent. Modify the displayed copy and then recall `g1` again; verify the stored copy was not altered.

### 58. Reverse pushes goals to predecessor states and may refine

Concept graph Reverse checkpoints with a backtrack point, asks the parent analysis UI to reverse-update concrete clauses from the graph's parent state, and if successful moves the graph to the predecessor state with constraints cleared into state. If reverse is impossible, it shows `Cannot reverse.` If reverse is infeasible but yields an interpolant, the parent UI offers refinement as a predicate in PDR mode or a concept in other modes.

To test this, view a non-initial state goal and click Reverse; verify the graph parent changes to the predecessor and constraints are incorporated. Reverse an initial-state goal and verify the Cannot reverse dialog. Construct an infeasible reverse with interpolant and verify the appropriate Refine dialog appears based on mode.

### 59. PDR step chains Reverse, Backtrack, Recalculate, and Diagram

`pdr_step` first calls Reverse. If the reverse result is false, it backtracks, recalculates, and may backtrack again or show `PDR terminated`; otherwise it diagrams the reverse result. This encodes a multi-action workflow behind one menu command. The graph's `reverse_result` attribute is used as the handoff between Reverse and Diagram.

To test this, create a reverse step with a non-false result and verify PDR step ends in a diagrammed goal. Create or mock a false reverse result and verify it backtracks and recalculates. Continue until no further undo/reverse result exists and verify `PDR terminated` is shown.

### 60. Diagram converts states or reverse results into diagram facts

Diagram checkpoints and calls `ivy_interp.diagram` on either the parent state plus current graph state or the stored reverse result. If the diagram is vacuous, it shows `The current state is vacuous.` or backtracks and retries when handling a reverse result. Successful diagrams are reskolemized, set as facts and state, and displayed.

To test this, run Diagram on a satisfiable goal and verify constraints are replaced by diagram facts. Run it on a vacuous state and verify the vacuous-state dialog. Run it after Reverse with a vacuous reverse result and verify it backtracks before retrying.

### 61. Reach checks one-step reachability from known reachable states

Reach asks the parent UI to find a reachable state satisfying the current goal constraints from known reachable states. On success it adds the reachable state to the reachable tree and opens a dialog saying `Goal reached! A reachable state has been added.` with `View state` and `OK` buttons. On failure it suggests trying Reverse.

To test this, create a one-step reachable goal and click Reach; verify the success dialog and reachable-tree update. Click View state and verify the concept graph changes to the reached state's clauses. Test an unreachable goal and verify the failure message mentions one step and Reverse.

### 62. Path reach runs bounded reachability along the current path

Path reach chooses current constraints if present, otherwise the graph state, and calls the parent analysis graph's `bmc` from the parent state. If a path is found, the parent adds the resulting analysis graph as a new tab. If no path is found, no tab is added.

To test this, create a path-reachable constraint and run Path reach; verify a new counterexample/reachability tab appears. Run it with an unreachable constraint and verify no new tab appears. Confirm that when constraints are true, the state formula is used as the condition.

### 63. Conjecture proposes invariants from known reachable states

The concept graph Conjecture command checkpoints, computes a goal from constraints or state, asks the parent UI for `case_conjecture`, and if successful shows a text dialog with the proposed invariant. It then reskolemizes the core, sets it as both facts and state, and updates. If no conjecture can be formed, it shows a message suggesting manual materialization.

To test this, create a known reachable-state setup where a separator exists and run Conjecture; verify the proposed invariant dialog appears and the graph updates to the core goal. Create a setup with no separator and verify the failure message. Undo after a successful conjecture and verify the previous graph returns.

## CTI / Invariant UI Behavior

### 64. CTI UI replaces the ART File/Action menus with invariant workflow menus

`ivy_ui_cti.AnalysisGraphUI.menus` exposes File entries Remove tab, Save invariant, Exit, and Invariant entries Check induction, Bounded check, Diagram, and Weaken. It omits the ART Mode menu and general Action menu. The CTI UI still inherits many ART behaviors but presents a narrower invariant-oriented surface.

To test this, launch with default `ui=cti` and verify the File and Invariant menus exactly match this set. Confirm the ART Mode menu is absent. Invoke inherited graph actions through concept graph menus and verify they still work where used.

### 65. CTI startup immediately views the initial state and detects transitive relations

CTI `start` calls the base startup, initializes transitive relation lists, sets current conjectures from `im.module.conjs`, views node 0, autodetects transitive binary relations from background theory, and records CTI state if the graph has `is_cti`. Detected transitive concepts have their `T` checkbox enabled and trigger an update. This makes the first screen a concept graph rather than just an ARG.

To test this, open a file with a transitive relation implied by axioms and verify its transitive checkbox is enabled at startup. Open a file with no such relation and verify no extra `T` boxes are set except intended defaults. Verify the initial state's concept graph is visible immediately after launch.

### 66. Check induction finds failing assertions or non-relatively-inductive conjectures

CTI Check induction builds a check ART, tests safety (`None`) and conjectures in order, constructs witness constants for universal conjectures, chooses relations to minimize if still at the placeholder text, and calls `ivy_trace.check_final_cond`. On a counterexample it replaces the displayed graph with the two-state result, views the pre-state, shows used relations, and displays either an assertion-failed OK dialog or a text dialog naming the non-inductive conjecture. If all checks pass, it shows `Inductive invariant found:` with the conjectures.

To test this, run Check induction on a safe inductive file and verify the inductive-invariant text dialog. Run it on a file with an assertion failure and verify a two-state graph is shown and the assertion-failed dialog appears. Run it on a file with a non-inductive conjecture and verify the failing conjecture text is displayed.

### 67. CTI used-relation display enables only relevant relation controls

`show_used_relations` clears edge controls, computes symbols used in a clause set, enables positive displays for matching relations, optionally enables negatives too, and adds special numeral-indexed relation concepts for ternary applications when needed. It updates relation controls before redrawing if new relation concepts were added. This keeps CTI counterexamples focused on the relevant vocabulary.

To test this, produce a failing conjecture involving a subset of relations and verify only those relation rows are enabled. Run with `both=True` through Diagram and verify negative boxes are also enabled for non-enumerated relations. Include a ternary relation with numeral first argument and verify a derived concept row is added.

### 68. CTI Diagram diagrams the pre-state of a CTI

If no CTI is currently available, Diagram first runs Check induction and stops if it passes or does not produce two states. It then computes the reverse image of the current conjecture or safety condition, conjoins pre-state, axioms, and universe constraints, obtains a model, converts it to a diagram, views the pre-state with the diagram clauses, shows used relations in both polarities, and gathers facts. This turns a CTI into a candidate strengthening pattern.

To test this, run Diagram after a failing induction check and verify the displayed graph switches to the pre-state diagram. Confirm relevant positive and negative relation boxes are enabled. Verify gathered facts appear as active facts/constraints for strengthening.

### 69. CTI Weaken removes selected conjectures

Weaken with no arguments lists current conjectures, displaying each without leading universals, under `Select conjecture to remove:` and requests multiple selection. The selected conjectures are removed from `self.conjectures`, `have_cti` is reset, and a text dialog lists removed conjectures. Because the underlying Tk listbox uses single selection, the legacy UI may effectively remove one at a time despite passing `multiple=True`.

To test this, open Weaken with several conjectures and verify the displayed formulas drop universals. Select one conjecture and verify it is removed from the invariant set and reported. Attempt multiple selection in the port and decide whether to preserve legacy single-select behavior or implement true multi-select with a compatibility test.

### 70. CTI Save invariant writes kept, dropped, and new conjecture sections

Save invariant writes `# This file was generated by ivy.` followed by sections for original conjectures kept, original conjectures dropped, and new conjectures. Kept originals are emitted as active `invariant` lines with labels preserved, dropped originals are commented out, and new conjectures are emitted unlabeled. It compares conjectures by string form against `im.module.conjs`.

To test this, remove one original invariant, add one new strengthening, and save. Verify the file has all three sections and the dropped invariant is commented. Verify labeled original invariants preserve their `[label]` syntax.

### 71. CTI bounded check asks for a bound and may offer to view a counterexample

`bmc_conjecture` asks for `Number of steps to check:` if no bound is supplied, remembers the current bound, builds a fresh analysis graph, optionally runs `initialize`, then checks the chosen conjecture or all conjectures up to the bound. On a counterexample it shows a text dialog with a `View` command that adds the counterexample graph using the ART UI class. On no counterexample it shows a text dialog reporting the bound.

To test this, invoke Bounded check and verify the integer dialog appears with the previous bound as initial value after the first run. Use a file with a bounded counterexample and click View; verify a new ART-style tab appears. Use a safe bound and verify the no-counterexample dialog includes the bound and conjecture text.

### 72. CTI concept graph has Conjecture-specific menus

`ivy_ui_cti.ConceptGraphUI` replaces the base concept-graph menu with Conjecture entries Undo, Redo, Gather, Bounded check, Minimize, Check sufficient, Check relative induction, Strengthen, and Export, plus View/Add relation. Node left-click toggles selection directly; edge left-click toggles selection directly. Right-click node menus only expose projections.

To test this, open a CTI concept graph and verify the Conjecture menu entries and absence of the base Action entries. Left-click nodes and edges and verify selection toggles without a popup. Right-click a node with projections and verify the projection menu appears.

### 73. CTI selected facts become a universal conjecture

`get_selected_conjecture` reads active facts, rejects facts with free variables, replaces uninterpreted numerals with generated variables, negates facts, simplifies the disjunction, and rewrites an OR of negated facts into the negation of a conjunction. This produces a positive universal conjecture from selected diagram facts. The generated conjecture is used by BMC, minimization, sufficiency, induction, and strengthening.

To test this, select a set of ground facts and run Strengthen; verify the proposed conjecture generalizes constants to variables. Include a fact with a free variable and verify the assertion/error path prevents conjecture creation. Compare the displayed conjecture to the selected facts to ensure it is the negation of their conjunction.

### 74. CTI Gather uses selected nodes and visible edges

CTI `gather_facts` uses selected nodes, or all graph nodes if none are selected. It gathers node facts for selected nodes and edge facts only between selected nodes, using visible all-to-all, none-to-none, transitive, and equality rules; it filters duplicate/self disequalities. It stores fact-to-element mappings, updates the graph, and highlights selected facts.

To test this, select two nodes, enable one relation, click Gather, and verify only facts involving those nodes and visible relation states are gathered. Select no nodes and verify facts from all nodes are gathered. Verify disequality duplicates such as reversed equality facts are filtered.

### 75. CTI Strengthen confirms before adding a conjecture

Strengthen builds the selected conjecture, drops universals for display, and opens a text dialog titled `Add the following conjecture:`. Pressing the dialog's action adds the conjecture to the parent invariant list and marks `have_cti=False`. Cancel leaves the invariant list unchanged.

To test this, select facts and click Strengthen; verify the confirmation dialog text. Accept and verify the conjecture is appended and later appears in Check induction. Cancel and verify no conjecture is appended.

### 76. CTI Minimize first runs BMC, then reduces facts with an unsat core

Minimize calls BMC for the selected conjecture with `tell_unsat=False`; if a counterexample is found, it stops. Otherwise it executes the system to the current bound, conjoins post-state clauses with axioms, computes an unsat core from active facts, keeps only facts in the core, highlights them, and displays the possible conjecture. If the unsat core is `None`, it treats it as an empty clause set.

To test this, select an overly large fact set with no bounded counterexample and run Minimize; verify the selected/highlighted facts shrink. Create a bounded counterexample and verify minimization stops after reporting the counterexample. Test a trivially true conjecture path and verify the empty-core case does not crash.

### 77. CTI Check sufficient compares selected conjecture to current CTI conjecture

Check sufficient builds a pre-state from the selected conjecture plus existing conjectures, executes one environment step, and checks whether it implies the current failing conjecture at the next time. If not, it offers `View counterexample?` with a View button; otherwise it reports implication. The dialog text labels the selected conjecture as `(1)` and target as `(2)`.

To test this, run Check sufficient after an induction failure and select facts insufficient to prove the target; verify the View counterexample dialog and new graph when accepted. Select stronger facts and verify the success dialog. Confirm the dialog text includes both numbered formulas.

### 78. CTI Check relative induction checks the selected conjecture against itself

Check relative induction uses the selected conjecture as both precondition and target, adds existing conjectures to the pre-state, executes one environment step, and checks final condition. It reports either `(1) is not relatively inductive. View counterexample?` with View or `(1) is relatively inductive:`. This lets the user validate a candidate strengthening before adding it.

To test this, select a non-inductive fact pattern and verify a counterexample can be viewed. Select an inductive pattern and verify the success dialog. Confirm existing conjectures are included in the pre-state by comparing behavior with and without them.

## Event Viewer Behavior

### 79. Event viewer launches its own Tix tree UI

`ivy_ev_viewer.main` parses a `.iev` event file, creates a `tkinter.tix.Tk`, calls `RunSample`, and starts `mainloop`. The UI has a left notebook of event-tree sheets and a right pattern-list panel. It also installs the default `SIGINT` handler.

To test this, launch the event viewer with a `.iev` file and verify a Tix window opens with an event tree and pattern panel. Launch with the wrong number of arguments and verify usage is printed. Send an interrupt and verify default signal behavior.

### 80. Event tree sheets display nested events lazily

Each `EventNoteBook.new_sheet` adds a tab labeled `Sheet N`, creates an `EventTree`, and configures `opencmd` to call `opendir`. `opendir` only loads children the first time a directory is opened; later opens show already-loaded children. The tree uses `/`-separated addresses such as `0/1/2`.

To test this, open an event with nested children and verify child rows appear only when expanded. Collapse and expand again and verify the same entries reappear without duplication. Check that tab labels increment for filtered result sheets.

### 81. Event tree menu supports Filter and Find reverse

The Event menu contains `Filter...` and `Find reverse...`. Both ask for a pattern string using an entry dialog titled `Pattern:`; Filter creates a new sheet containing events matching the parsed pattern, while Find reverse searches backward from the current selection and selects the found event. Pattern parse errors are reported as Ivy syntax errors.

To test this, apply a filter pattern and verify a new sheet contains only matching events. Select an event, run Find reverse, and verify the matching prior event is uncovered, selected, and scrolled into view. Enter invalid pattern syntax and verify an Ivy error dialog.

### 82. Pattern list supports saved reusable search patterns

The pattern panel has `<<`, `>>`, `+`, `-`, `Save`, `Load`, and `Clear` buttons. `+` prompts for a pattern and appends its parsed representation to a `TList`; `<<` searches reverse with the selected pattern and `>>` searches forward. Save writes one pattern per line to a `.pats` file and Load parses each line back through the same pattern parser; `-` and Clear are currently stubs.

To test this, add a pattern and verify it appears in the pattern list. Select it and use `<<` and `>>` to navigate matches in the current sheet. Save patterns to a file, clear in a controlled fixture if implemented, load the file, and verify patterns are restored; verify legacy `-` and Clear no-op behavior if preserving it.

## Notebook / Widget UI Behaviors That Inform The Port

### 83. Cytoscape elements carry callbacks, context actions, tooltips, and long info

`CyElements` represents nodes, edges, and shapes with stable IDs, labels, class strings, short info, long info, event callback tuples, and context-menu action tuples. Tk rendering ignores most web-widget metadata, but the same data model drives graph behavior and is serialized by newer sidecar paths. Once assigned to `CyGraphWidget.cy_elements`, a `CyElements` instance is invalidated by setting `elements=None`.

To test this, create nodes and edges with callbacks/actions and verify the web widget can call the callback with the original Python object references. Assign a `CyElements` instance and verify later mutation or reuse fails. In the Tk port, verify equivalent object identity mappings are preserved for click dispatch.

### 84. Web widget graph selection is represented as tuples

`CyGraphWidget.elements` exposes nodes as `(obj,)` and edges as `(obj, source_obj, target_obj)`, and `selected` is a tuple of those element tuples. The widget resets `selected` to an empty list when graph elements change. Concept fact gathering in the notebook UI uses this exact shape.

To test this, select a node and edge in the widget and verify the selected property contains the expected tuple shapes. Replace graph elements and verify selection clears. Use selection to gather facts and confirm node versus edge facts are distinguished by tuple length.

### 85. Concept-session view controls support bulk toggling by relation and class

The notebook `ConceptSessionControls` builds edge display checkboxes for `all_to_all`, `edge_unknown`, `none_to_none`, and `transitive`, and node-label checkboxes for necessary, maybe, and necessarily-not. Clicking a relation-name button toggles all three edge display classes for that relation; clicking a header button toggles that display class across all relations. Any checkbox change rebuilds the concept domain's visible edges and labels before recomputing.

To test this in a port that includes equivalent controls, click a relation label and verify its positive, unknown, and negative boxes all toggle together. Click the `+` header and verify all relations' positive boxes toggle. Confirm that changing checkboxes changes the computed domain before rendering, not merely CSS visibility after rendering.

### 86. Notebook dialogs provide asynchronous modal interactions

`ui_extensions_api` defines frontend operations such as `ShowModal`, `UserSelect`, `UserSelectMultiple`, and `ExecuteNewCell`. Generator-based interaction functions yield these operations and resume when the user responds or a generated cell finishes executing. The wrapper refuses to run generator interactions unless the analysis session is on the latest history step.

To test this, invoke a generator-based action from a non-latest history step and verify a modal error says it must be on the last step. Invoke an action that asks the user to select values and verify cancel returns `None` while OK returns selected values. Invoke an action that executes a new cell and verify the callback resumes only after the expected cell completes.

### 87. UI extension points add node actions dynamically

`arg_node_actions` and `goal_node_actions` are extension points that collect callback lists and return action tuples for context menus. Built-in ARG actions include executing every available action, trying unproved conjectures, new goal, recalculate, check cover, remove facts, and join with selection. `tactics_api.goal_tactic` registers proof-goal tactics automatically as UI actions.

To test this, register a test action through an extension point and verify it appears in the appropriate node context menu. Trigger the built-in Execute action entry and verify it creates/executed the expected command. Define a dummy goal tactic with the decorator and verify the corresponding action appears on proof-goal nodes.

### 88. Interactive UPDR uses modal fact/core selection

`iupdr.interactive_updr` is a generator interaction that repeatedly asks users to select diagram literals for generalization and unsat-core literals for refinement. `UserSelectCore` includes a SelectMultiple list, a `Check SAT` button, and a result label that updates to SAT or UNSAT. OK returns selected constraints plus whether the selected set was satisfiable; Cancel returns `(None, None)`.

To test this, start interactive UPDR in a state where it is allowed and verify the Generalize Diagram modal lists literals with all selected by default. In the refinement modal, select a subset and click Check SAT; verify the result label updates. Cancel the modal and verify the interaction receives `(None, None)` and handles it consistently with legacy behavior.

### 89. Modal message widget groups tactic messages

`ModalMessagesWidget.new_message` sends a frontend message with method `new_message`, title, and body. `AnalysisSessionWidget.step` uses it for tactic messages unless `silent` is true. Multiple messages may be grouped by the frontend implementation.

To test this, execute a tactic step that includes a `msg` field and verify a modal message appears with title `Tactic <name> says:`. Set the session widget to silent and verify no modal is shown. Trigger multiple messages and verify the frontend groups or queues them without dropping bodies.

### 90. Analysis-session history navigation changes rendered graphs

The notebook `AnalysisSessionWidget` tracks `current_step` and has First, Prev, Next, and Last buttons. Rendering updates proof graph, ARG, concrete reachability graph, and step text from the selected history entry. `step()` jumps to the latest step, optionally opens the active proof goal or state in the concept viewer.

To test this, create a session with several history steps and use each navigation button; verify graph content and step text change. Trigger a new step and verify the current step moves to the latest. Set the active item to a proof goal, ARG state, and CRG state in separate tests and verify the concept viewer opens the right context.

## Diagnostic And Non-Default GUI Entry Points

### 91. `ivy_check` opens the GUI for diagnostics

When `diagnose=true`, property failures, conjecture failures, and explicit counterexample display paths open the Tk UI rather than just raising an Ivy error. `display_cex`, `check_properties`, `check_conjectures`, `show_counterexample`, and `gui_art` create or reuse analysis graphs, set `mode=induction` for some flows, update idletasks so dialogs sit above the main window, and start the main loop. These paths exit the process after the diagnostic GUI closes.

To test this, run `ivy_check` with `diagnose=true` on a property failure and verify a GUI opens to choose a property/counterexample. Run with `diagnose=false` and verify it raises or prints the error without opening a GUI. Confirm diagnostic GUI exit terminates the process with the expected failure status.

### 92. `try_property` displays false properties and opens BMC traces

`IvyUI.try_property` lists false properties with `Choose a property to see counterexample:`. Selecting a property browses its source if a line number is present, computes the dual of the property under its context, runs BMC from a fresh analysis graph, and adds the resulting graph to the UI. It prints the property type for debugging.

To test this, create a file with a false property and invoke try_property; verify the list dialog appears. Select the property and verify the source browser highlights its line. Confirm a new counterexample graph tab is added.

### 93. `ivy2.py` creates or opens a notebook-based UI file

The `ivy2.py` script creates a notebook next to a given `.ivy` file if one does not already exist, otherwise opens the existing notebook. The generated first cell seeds random/z3, imports proof/session/widget/tactics modules, creates `AnalysisSessionWidget`, starts an `AnalysisSession`, sets context, initializes CTI conjectures, ensures `initialize` and `step` actions exist, autodetects transitive relations, and displays the widget. It then starts IPython notebook with the file.

To test this, run `ivy2.py sample.ivy` when no notebook exists and verify `sample.ivy.ipynb` is created with the expected first cell. Run it again after creating `sample.ipynb` and verify it opens the existing notebook instead. Execute the generated cell and verify the analysis widget appears with transition-view conjectures initialized.

### 94. `ivy_launch.py` runs generated distributed processes in terminals

`ivy_launch` reads a `.dsc` descriptor, parses command-line parameter values, computes process dimensions, allocates loopback endpoint IDs and ports, and runs each process command. If more than one process is launched, each command is started in an `xterm` with a large courier font and title based on the process instance; a shell read prompt keeps the terminal open. For single-process test runs with `runs`, output is redirected to a log and event counts are summarized.

To test this, launch a descriptor with multiple process instances and verify one titled terminal per process instance appears. Verify endpoint parameters are auto-filled with unique loopback ports. Run with `runs=N` and verify seeds are passed, logs are written, and nonzero return codes cause a failure message.

### 95. `ivy_show.py` is a non-interactive compile/show helper, not a GUI

`ivy_show.py` parses parameters, forces `show_compiled=true`, loads the Ivy file without creating isolates initially, creates the requested isolate, and prints compiled output. It imports `tk_ui` but does not open a Tk main loop. This file is UI-adjacent only because it shares parameter and compile paths.

To test this, run `ivy_show.py file.ivy` and verify no GUI window appears. Confirm compiled output is printed. Pass invalid arguments and verify usage text is printed.

### 96. `ivy_ui_none.py` disables UI compile extras

The `ivy_ui_none` module exports only `compile_kwargs = {}`. It is selected via the default UI module resolution if `ui=none` is supplied. This provides a way to compile/run without UI-specific source loading extensions.

To test this, launch or compile with `ui=none` and verify no CTI `ext` compile option is injected. Confirm code paths that request `get_default_ui_compile_kwargs()` receive an empty dict. If a GUI class is requested from this module, verify the port reports the missing class clearly or avoids that path.

## Rendering Details From Shared Graph Code

### 97. ARG rendering includes states, action edges, join edges, and cover edges

`ivy_art.render_rg` turns each state into a node labeled by state ID, classing bottom states as `bottom_state` and others as `state`. Transitions become edges classed `transition_action` or `transition_join`, and covering relation entries become unlabeled dashed `cover` edges. Long info for state nodes is the open-formula clauses of the state.

To test this, create an ARG with a normal state, a bottom state, an action transition, a join transition, and a cover relation. Verify each node and edge is rendered with the correct class and label. Open tooltip/long-info equivalents in the port and verify state formulas and action text are available.

### 98. Concept graph vocabulary updates from state symbols

When graph state changes, `Graph.state_changed` clears the concept-session cache, rebuilds the domain vocabulary from node formulas, current formula symbols, global signature symbols, and non-skolem symbols used by state/constraints. It then re-adds user-added new relations and recomputes if requested. This keeps relation controls synchronized with whatever symbols the new state mentions.

To test this, view a state that introduces a relation not visible in the previous state and verify a new relation row appears. Add a custom relation, change state, and verify the custom relation survives the vocabulary refresh. Verify skolem-only symbols are not treated as normal signature vocabulary.

### 99. Default label concepts for numerals are shown automatically

`Graph.default_labels` returns node-label concepts that are numeral equality concepts. During checkbox sync, default labels get checkbox index 0 enabled regardless of user relation state. This makes numeral witness labels visible by default.

To test this, materialize or diagram a state with numeral equality labels and verify those labels appear without manually toggling their rows. Toggle other labels off and verify numeral defaults still show after update. Confirm non-numeral equality labels do not get this automatic treatment.

### 100. Transitive reduction hides self edges and implied transitive edges

Both `ivy_graph.get_transitive_reduction` and `cy_render.get_transitive_reduction` hide all self edges and edges implied by a two-step path when a relation's transitive checkbox is on and the all-to-all fact is true. Custom edges are exempt from hiding. This is display-only; the underlying facts remain present.

To test this, create a relation with edges A->A, A->B, B->C, and A->C and enable `T`. Verify A->A and A->C are hidden while A->B and B->C remain. Add a custom A->C edge and verify it remains visible despite transitive reduction.

### 101. Node labels encode necessary, maybe, and negative facts

Rendering examines node-label facts in priority order: necessary, necessarily-not, else maybe. The modern renderer prefixes labels with nothing, `~`, or `?` in Tk/legacy graph code, and the notebook renderer uses logical symbols for negation/equality variants. Labels whose display checkbox is off are hidden unless they are custom labels.

To test this, create a node with one necessary label, one impossible label, and one maybe label. Toggle each label class and verify only matching prefixes appear. Add a custom label and verify it appears even when the corresponding checkbox would normally hide it.

### 102. Equality labels are abbreviated for readability

`Graph.concept_label` abbreviates unary formulas such as `p(X)`, `X=e`, and `f(X)=e` by dropping the variable and removing spaces and empty parentheses. Numeral equality concepts are treated specially for default display. More complex concepts fall back to full formula strings.

To test this, create labels for `p(X)`, `X=a`, `f(X)=b`, and a more complex formula. Verify the first three render in abbreviated form and the complex one renders fully. Confirm tests compare user-facing labels, not internal concept names.

### 103. Concept graph shape defaults differ between Tk graph variants

In `ivy_graph.get_shape`, concept graph nodes are rendered as octagons in the Tcl/Tk concept graph path. In `cy_render.get_shape`, notebook rendering uses special shapes for concept names such as `__Node` and `__ID`, defaulting to ellipse. The Tk port should preserve the Tcl app's octagon default where matching the original GUI is the goal.

To test this, render a standard Tcl concept graph and verify concept nodes are octagons. Render ARG nodes and verify they remain ovals/ellipses rather than octagons. If notebook-style rendering is also supported, verify its special shape map separately.

### 104. Canvas rendering maps generic graph elements to toolkit objects

`TkCyCanvas.create_elements` maps node elements to ovals/octagons plus centered text, edge elements to smooth bezier lines plus separate arrowhead lines and optional labels, and shape elements to rectangles. It binds both left and right mouse buttons on each element tag. It stores `elem_ids` and `edge_points` for later selection and highlighting.

To test this, render a graph with nodes, edges, labels, and subgraph boxes and verify all corresponding canvas items are created. Click the label text and the shape body and verify both dispatch to the same element action through shared tags. Highlight an edge and verify the stored spline points are used.

### 105. Context popup behavior has a special direct-action shortcut

`TkCyCanvas.make_popup` immediately invokes the action if the action list has exactly one entry labeled `<>`; otherwise it builds a right-click-style popup menu at the event screen coordinates. A menu entry with label `---` becomes a separator, and an entry with `cmd is None` becomes a disabled-looking command label with no callback. Other entries call the command with the clicked object as the first argument.

To test this, left-click an ARG node and verify the single `<>` action runs directly. Right-click a node with section labels and separators and verify the popup structure matches. Click a normal menu entry and verify the command receives the clicked node or edge.

## Audit Coverage

The files explicitly listed as likely UI files were read and traced for behavior: `concept.py`, `concept_alpha.py`, `concept_interactive_session.py`, `concept_space_parsetab.py`, `cy_elements.py`, `cy_render.py`, `cy_styles.py`, `dot_layout.py`, `ev_parsetab.py`, `general.py`, `interrupt_context.py`, `ivy.py`, `ivy2.py`, `ivy_art.py`, `ivy_compose.py`, `ivy_concept_space.py`, `ivy_congclos.py`, `ivy_core.py`, `ivy_formulatab.py`, `ivy_graph.py`, `ivy_graph_ui.py`, `ivy_graphviz.py`, `ivy_init.py`, `ivy_launch.py`, `ivy_shell.py`, `ivy_show.py`, `ivy_termtab.py`, `ivy_ui.py`, `ivy_ui_cti.py`, `ivy_ui_none.py`, `ivy_ui_util.py`, `tk_cy.py`, `tk_graph_ui.py`, `tk_ui.py`, `ui_extensions_api.py`, `widget_analysis_session.py`, `widget_cy_graph.py`, `widget_dialog.py`, `widget_modal.py`, and `widget_modal_messages.py`.

The remaining Python files under `~/ivy/pyivy/ivy/ivy` were checked for UI-facing imports, launch hooks, callbacks, display/menu/dialog behavior, or graph-rendered behavior. Files with additional UI-adjacent behavior beyond the likely list include `ivy_check.py`, `ivy_ev_viewer.py`, `iupdr.py`, `sidecar.py`, `tactics_api.py`, and `ivy_utils.py`; their relevant behaviors are captured above. Parser tables, solver/core logic, compiler modules, Z3 vendored files, tests, and utility data-structure modules did not add Tcl/Tk GUI behavior beyond supporting the state, formula, concept, or graph data consumed by the UI.

All `.py` files audited in this pass:

`__init__.py`, `canon.py`, `canon_ast.py`, `canon_fragment.py`, `client_server_example.py`, `concept.py`, `concept_alpha.py`, `concept_interactive_session.py`, `concept_space_parsetab.py`, `cy_elements.py`, `cy_render.py`, `cy_styles.py`, `dot_layout.py`, `ev_parsetab.py`, `general.py`, `interrupt_context.py`, `iupdr.py`, `ivy.py`, `ivy2.py`, `ivy2/stage2.py`, `ivy2/stage3.py`, `ivy2/stage4.py`, `ivy2/stage5.py`, `ivy2/stage6.py`, `ivy2/stage7.py`, `ivy2/test1.py`, `ivy_acl.py`, `ivy_actions.py`, `ivy_alpha.py`, `ivy_art.py`, `ivy_ast.py`, `ivy_auto_inst.py`, `ivy_bmc.py`, `ivy_check.py`, `ivy_compiler.py`, `ivy_compose.py`, `ivy_concept_space.py`, `ivy_congclos.py`, `ivy_core.py`, `ivy_cpp.py`, `ivy_cpp_types.py`, `ivy_dafny_ast.py`, `ivy_dafny_compiler.py`, `ivy_dafny_grammar.py`, `ivy_dafny_lexer.py`, `ivy_dafny_parser.py`, `ivy_dafny_parsetab.py`, `ivy_dump.py`, `ivy_ev_parser.py`, `ivy_ev_viewer.py`, `ivy_formulatab.py`, `ivy_fragment.py`, `ivy_graph.py`, `ivy_graph_ui.py`, `ivy_graphviz.py`, `ivy_init.py`, `ivy_interp.py`, `ivy_isolate.py`, `ivy_l2s.py`, `ivy_launch.py`, `ivy_lexer.py`, `ivy_libs.py`, `ivy_logic.py`, `ivy_logic_parser.py`, `ivy_logic_parser_gen.py`, `ivy_logic_utils.py`, `ivy_lsp.py`, `ivy_lsp_client.py`, `ivy_mc.py`, `ivy_module.py`, `ivy_parser.py`, `ivy_parsetab.py`, `ivy_printer.py`, `ivy_proof.py`, `ivy_ranking.py`, `ivy_resolution.py`, `ivy_shell.py`, `ivy_show.py`, `ivy_smtlib.py`, `ivy_solver.py`, `ivy_tactics.py`, `ivy_temporal.py`, `ivy_termtab.py`, `ivy_theory.py`, `ivy_to_cpp.py`, `ivy_to_lean.py`, `ivy_to_md.py`, `ivy_trace.py`, `ivy_transrel.py`, `ivy_ui.py`, `ivy_ui_cti.py`, `ivy_ui_none.py`, `ivy_ui_util.py`, `ivy_union_find.py`, `ivy_union_find2.py`, `ivy_unitres.py`, `ivy_utils.py`, `ivy_vmt.py`, `logic.py`, `logic_sexp.py`, `logic_util.py`, `proof.py`, `sidecar.py`, `tactics.py`, `tactics_api.py`, `tests/test_base.py`, `tests/test_ivy_union_find2.py`, `tk_cy.py`, `tk_graph_ui.py`, `tk_ui.py`, `token_counter.py`, `type_inference.py`, `ui_extensions_api.py`, `utils/__init__.py`, `utils/immutables.py`, `utils/recstruct_object.py`, `utils/rectagtuple.py`, `utils/try1.py`, `utils/try11.py`, `utils/try2.py`, `utils/try3.py`, `widget_analysis_session.py`, `widget_cy_graph.py`, `widget_dialog.py`, `widget_modal.py`, `widget_modal_messages.py`, `xtracer.py`, `z3/__init__.py`, `z3/z3.py`, `z3/z3consts.py`, `z3/z3core.py`, `z3/z3num.py`, `z3/z3poly.py`, `z3/z3printer.py`, `z3/z3rcf.py`, `z3/z3types.py`, `z3/z3util.py`, and `z3_utils.py`.
