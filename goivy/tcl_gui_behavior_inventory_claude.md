# Ivy Tcl/Tk GUI Behavior Inventory

This document enumerates every behavior of the Ivy Tcl/Tk Python GUI (from `~/ivy/pyivy/ivy/ivy/`) for use in porting to a web interface. Each behavior has a 2-3 sentence description and 3+ sentences on how to test it.

Source files read: `tk_ui.py`, `ivy_ui.py`, `ivy_ui_cti.py`, `ivy_ui_util.py`, `ivy_ui_none.py`, `ui_extensions_api.py`, `widget_analysis_session.py`, `widget_cy_graph.py`, `widget_dialog.py`, `widget_modal.py`, `widget_modal_messages.py`, `ivy_graph.py`, `ivy_graph_ui.py`, `ivy_graphviz.py`, `tk_cy.py`, `tk_graph_ui.py`, `cy_elements.py`, `cy_render.py`, `cy_styles.py`, `dot_layout.py`, `ivy_shell.py`, `ivy_launch.py`, `ivy_show.py`, `ivy_ev_viewer.py`, `ivy_trace.py`, `concept.py`, `concept_interactive_session.py`, `ivy_concept_space.py`, `ivy_art.py`, `ivy_compose.py`, `ivy_congclos.py`, `ivy_core.py`, `ivy_interp.py`, `ivy_check.py`, `ivy_init.py`, `ivy2.py`, `ivy.py`, `general.py`, `interrupt_context.py`.

---

## 1. Application Startup & Initialization

### 1.1 Entry Point Execution
The main entry point (`ivy.py:main()`) reads command-line parameters via `ivy_init.read_params()`, loads the `.ivy` source file, and calls `tk_ui.ui_main_loop(ag)` to display the GUI. SIGINT is set to default before launch so Ctrl+C exits cleanly. The module context (`im.Module()`) wraps the entire session so all loaded symbols are properly scoped.

**Testing:** Launch `ivy myfile.ivy` and confirm the GUI window appears. Verify that Ctrl+C at the terminal exits without a Python traceback. Confirm a syntax error in `myfile.ivy` prints an error to stderr and exits without showing the GUI.

### 1.2 Tk Root Window Creation
`TkUI.__init__` creates a `tkinter.tix.Tk()` root window (not plain `Tk()`—Tix extensions are required). The global Tk palette is set to white (`tk.tk_setPalette('white')`). The window title is set to `"ivy"` immediately after creation.

**Testing:** Observe that the window title bar reads "ivy". Confirm the background color of all widgets defaults to white. Confirm that Tix-specific widgets (NoteBook, HList, etc.) are available without error.

### 1.3 Window Title
The root window title is set to `"ivy"` via `self.tk.title("ivy")` in `TkUI.__init__`. No dynamic title changes occur during the session; the title remains "ivy" for the lifetime of the application.

**Testing:** Open the application and read the OS window title. Confirm it says "ivy". Maximize, minimize, and restore; the title should remain "ivy" throughout.

### 1.4 Tix NoteBook Initialization
A `tkinter.tix.NoteBook` widget is created as the only top-level child of the root and packed with `fill=BOTH, expand=1`. This notebook holds all analysis sheets as tabs. `tab_counter` and `tabs` instance variables are initialized to 0.

**Testing:** Open a file and confirm a tabbed notebook widget appears filling the window. Resize the window; the notebook should resize with it. Verify the first tab is automatically selected.

### 1.5 Initial Sheet Creation
When `ui_main_loop(art)` is called, it calls `ui.add(art)` to add the first analysis graph as a tab named `"sheet_1"` with label `"Sheet 1"`. The new tab is immediately raised to front. `gw.start()` is called on the new widget to initialize the display.

**Testing:** Open a file and confirm "Sheet 1" tab exists and is selected. The ARG graph should be visible. No empty or error state should appear for a valid file.

### 1.6 DYLD_LIBRARY_PATH Setup (macOS)
`ivy_shell.py:main()` detects the macOS platform and exports `DYLD_LIBRARY_PATH` to include library paths needed for native dependencies (Z3, etc.). This happens before any GUI is shown. Without this, native library loading may fail on macOS.

**Testing:** On macOS, confirm that Z3 is loadable after the shell setup. Check `os.environ['DYLD_LIBRARY_PATH']` before and after `ivy_shell.main()` and confirm the paths were added.

---

## 2. Main Window Layout

### 2.1 PanedWindow Split View per Tab
Each tab contains a horizontal `PanedWindow` with two panes: a left pane (minimum width 50px) for the ARG graph canvas, and a right pane (minimum width 200px) for the state/concept graph display. The panes are resizable by dragging the sash.

**Testing:** Open an ivy file and confirm two side-by-side panes. Drag the sash left and right and confirm both panes resize. Confirm neither pane can be collapsed below its minimum size.

### 2.2 Left Pane: ARG Graph Canvas
The left pane contains the Analysis/Reachability Graph (ARG) canvas with both horizontal and vertical scrollbars. The canvas widget fills the pane.

**Testing:** Open a file with multiple reachable states. Confirm the ARG is visible in the left pane. If the graph is larger than the pane, confirm scrollbars appear and scrolling works.

### 2.3 Right Pane: State Display Frame
The right pane is a frame (`f2`) that holds the state/concept graph widget when a node is clicked. When a node is selected, a concept graph widget populates this frame.

**Testing:** Open a file, click a node in the ARG. Confirm the right pane populates with a concept graph. Confirm scrollbars in the right pane work if the concept graph is large.

### 2.4 Menu Bar Placement
The menu bar (`MenuBar`) is a Frame packed at the top of the analysis graph widget. Menu buttons are packed left-to-right. An empty frame is packed to the right to push buttons left.

**Testing:** Confirm menu buttons (File, Mode, Action, etc.) appear horizontally at the top of each tab's graph widget. Resize the window narrower and confirm buttons remain visible.

### 2.5 Scrollbars for ARG Canvas
Two scrollbars are created: one vertical (right side) and one horizontal (bottom). They are linked to the graph canvas via `xview` and `yview`.

**Testing:** Create a large ARG with many nodes. Confirm both scrollbars appear. Drag the horizontal scrollbar and confirm the graph pans left/right. Drag the vertical scrollbar and confirm up/down panning.

---

## 3. Tab/Sheet Management

### 3.1 Adding a New Sheet Tab
`TkUI.add(art, name, label, ui_class)` creates a new tab in the NoteBook. `tab_counter` increments on each call. Default name is `"sheet_{N}"`, default label is `"Sheet {N}"`. The new tab is raised to front with `nb.raise_page(name)` immediately after creation.

**Testing:** Trigger an operation that adds a new tab (e.g., "Show reachable states" or "Step in"). Confirm a new tab labeled "Sheet 2" appears. Confirm the new tab is immediately selected/active.

### 3.2 Removing a Sheet Tab
`TkUI.remove(art)` deletes the tab by name from the NoteBook and decrements `tabs`. If `tabs` reaches 0 after removal, `tk.quit()` is called and the application exits.

**Testing:** With a single sheet open, select "Remove tab" from File menu. Confirm the application exits. With two sheets, remove one; confirm the other remains and the application continues running.

### 3.3 Tab Counter Persistence
`tab_counter` is only incremented, never decremented. If Sheet 1 and Sheet 2 exist and Sheet 1 is removed, the next sheet added will be Sheet 3 (not Sheet 2).

**Testing:** Open a file (Sheet 1 created). Add Sheet 2. Remove Sheet 2. Add another sheet. Confirm it is named "Sheet 3", not "Sheet 2".

### 3.4 Tab Label Display
Each tab in the NoteBook shows its label string (e.g., "Sheet 1"). Labels are static after creation.

**Testing:** Open the application with two sheets. Confirm both tabs show their respective labels. Verify labels do not change dynamically during interaction.

### 3.5 Raising Tab to Front
When `add()` creates a new tab, `nb.raise_page(name)` brings it to the foreground. When operations create new ARGs, the new sheet's tab becomes active.

**Testing:** Be viewing Sheet 1 and trigger creation of a new ARG. Confirm the new Sheet 2 tab is automatically activated and visible.

---

## 4. Menu Bar System

### 4.1 Menu Bar Widget Construction
`WithMenuBar.__init__` reads the `menus()` method of the widget class, iterates through the menu structure list, and creates `Menubutton` widgets packed left in the `MenuBar` frame. Each `Menubutton` has `tearoff=0`.

**Testing:** Confirm menu buttons appear in the menu bar. Confirm no tearoff perforation is present. Confirm clicking a menu button opens a dropdown.

### 4.2 Menu Item: Button Type
Menu items of type `"button"` create a `Menu.add_command(label=..., command=...)` entry. Clicking executes the callback immediately.

**Testing:** Click "File" menu then "Save". Confirm a save-as dialog appears. Click "Action" → "Recalculate all". Confirm recalculation runs.

### 4.3 Menu Item: Separator
Menu items of type `"separator"` add a visual divider line in the dropdown. They have no command or label.

**Testing:** Open the File menu. Confirm a horizontal rule appears between "Save abstraction" and "Remove tab". Confirm clicking on the separator does nothing.

### 4.4 Menu Item: Radiobuttons
Menu items of type `"radiobuttons"` create `Menu.add_radiobutton()` entries sharing a `StringVar`. The variable stores the currently selected option name.

**Testing:** Open "Mode" menu. Confirm radio buttons for Concrete, Abstract, Bounded, Induction, Pdr. Select "Induction"; reopen the menu and confirm "Induction" has a bullet/checkmark. Confirm only one mode is selected at a time.

### 4.5 RadioButton Variable Storage
Each radiobutton set has its `StringVar` stored in `self.radios[name]`. This allows code to retrieve the current mode via `self.radiobutton('mode').get()`.

**Testing:** Programmatically set the mode variable and confirm the menu reflects the change. Confirm `AnalysisGraphUI.mode` property reads the correct value after a menu selection.

---

## 5. File Menu

### 5.1 Save Analysis State
"Save" in the File menu opens a "Save analysis state as..." dialog filtered to `.a2g` files. The ARG is pickled to the file. Canceling does nothing.

**Testing:** Open a file with some states. Click File → Save. Confirm a file dialog appears with `.a2g` filter. Save to a path. Confirm a `.a2g` file exists and can be unpickled as an ARG. Cancel and confirm no file is written.

### 5.2 Save Abstraction
"Save abstraction" opens a "Save abstraction as..." dialog filtered to `.ivy` files. The concept spaces (domain abstraction) are written as ivy source, enabling proof state reload.

**Testing:** After adding concept predicates, click File → "Save abstraction". Provide a `.ivy` filename. Confirm the file contains `conjecture` declarations. Open the saved file and confirm ivy can parse it.

### 5.3 Remove Tab
"Remove tab" calls `self.ui_parent.remove(self.g)` to remove the current sheet. If it is the last sheet, the application exits.

**Testing:** With one sheet, File → "Remove tab". Confirm the app exits. With two sheets, remove one and confirm the other remains.

### 5.4 Exit
"Exit" calls `self.ui_parent.exit()` which calls `self.tk.quit()`, ending the Tk main loop cleanly.

**Testing:** Click File → Exit. Confirm the window closes. Confirm no Python exception is printed. Confirm the process exits with code 0.

---

## 6. Mode Selection

### 6.1 Mode Radiobutton Set
The Mode menu contains radiobuttons for: `"Concrete"`, `"Abstract"`, `"Bounded"`, `"Induction"`, `"Pdr"`. The default mode is set when the UI is first created. Only one mode can be selected at a time.

**Testing:** Confirm all five modes appear as radiobuttons. Select each one and confirm it becomes checked. Confirm selecting one deselects the previous.

### 6.2 Mode: Concrete
In Concrete mode, `init_alpha()` returns `None` (no abstraction). The ARG operates on concrete states. Safety checking uses bounded safety.

**Testing:** Select Concrete mode. Trigger a state action. Confirm states are concrete (not abstracted). Confirm the safety check uses bounded path analysis.

### 6.3 Mode: Abstract
In Abstract mode, `init_alpha()` returns `None`. Like Concrete but abstraction may be applied via concept spaces.

**Testing:** Select Abstract mode. Add a concept and confirm it affects abstraction. Confirm behavior differs from Concrete when concept spaces are defined.

### 6.4 Mode: Bounded
In Bounded mode, `check_safety_node` calls `check_bounded_safety` instead of `check_local_safety`. BMC is the primary verification strategy.

**Testing:** Select Bounded mode. Click "Check safety" on a node. Confirm BMC runs rather than local safety checking. Confirm a bound is used for the path length.

### 6.5 Mode: Induction
In Induction mode, `init_alpha()` returns `ivy_alpha.alpha`. This enables inductive generalization when computing post-states.

**Testing:** Select Induction mode. Verify `init_alpha()` returns the induction alpha function. Trigger node extension and confirm inductive generalization is applied.

### 6.6 Mode: PDR
In PDR mode, `init_alpha()` returns `ivy_alpha.predicate_alpha`. The refine dialog uses PDR-specific messaging.

**Testing:** Select PDR mode. Trigger a safety check and confirm PDR-specific messages appear. Verify `init_alpha()` returns the predicate alpha function.

### 6.7 Mode Affects init_alpha()
`init_alpha()` reads `self.mode` and returns different alpha functions. Default (unknown mode) returns `top_alpha`.

**Testing:** Change the mode and trigger a node extension. Confirm the post-state abstraction method changes. Test the default case and confirm `top_alpha` is used.

---

## 7. Action Menu

### 7.1 Recalculate All
"Recalculate all" iterates through all ARG transitions, recomputes each, and rebuilds the graph. Duplicate target nodes are skipped.

**Testing:** Build an ARG with multiple states. Click Action → "Recalculate all". Confirm all edges are recomputed. Verify no duplicate states appear.

### 7.2 Show Reachable States
"Show reachable states" adds a reachable-state tree to the UI as a new tab. The reachable tree is built lazily from the initial state.

**Testing:** Click Action → "Show reachable states". Confirm a new tab appears. Confirm it shows reachable states of the Ivy program.

---

## 8. ARG Graph Canvas

### 8.1 Canvas Deletion and Rebuild
`TkAnalysisGraphWidget.rebuild()` calls `self.delete('all')` on the canvas, clearing all items, then recreates everything from scratch. This is called on startup and after any state-modifying operation.

**Testing:** Add a state to the ARG. Confirm the canvas updates. Delete a node; confirm it disappears immediately. Confirm the canvas never shows stale data.

### 8.2 DOT Layout Computation
`rebuild()` creates `CyElements` from the ARG and calls `dot_layout(cy_elements)` using the Graphviz `dot` algorithm. The y-coordinate is flipped relative to `y_origin`.

**Testing:** Create an ARG with multiple states. Confirm nodes are laid out in a DAG structure (roots at top, leaves at bottom). Confirm positions are stable across rebuilds.

### 8.3 Edge Label Notation Transformation
After layout, `"-["` in edge labels is replaced with `"{"` and `"]-"` is replaced with `"}"`. This converts internal notation to human-readable brace notation for display.

**Testing:** Create a state transition using the `[-...]` notation. Observe the edge label and confirm it shows `{...}` instead of `[-...]`.

### 8.4 Node Shape Rendering
`TkCyCanvas.create_shape()` creates Tk canvas items: `ellipse`/`oval` uses `create_oval`, `octagon` uses `create_octagon`. The `double` parameter creates a second inner shape for double-border effects.

**Testing:** Confirm ARG nodes render as ovals. Confirm `__ID`-sort nodes render as octagons. Confirm `at_least_one` nodes display two concentric shapes.

### 8.5 Node Shape: Black Fill for Bottom States
In `get_node_styles()`, if the node is a `bottom_state` (error/bottom), `fill='black'`; otherwise `fill=''`.

**Testing:** Trigger a safety violation creating a bottom_state. Confirm it renders black. Confirm normal states have transparent fill.

### 8.6 Node Shape Outline Color
`update_node_color(node)` updates the node outline color. `get_node_styles()` returns `outline='black'` and `width=2` by default.

**Testing:** Create a node. Confirm its outline is black with width 2. Mark the node (which turns it red). Unmark and confirm black outline returns.

### 8.7 Canvas Scroll Region Update
After `create_elements()`, `self.configure(scrollregion=self.bbox('all'))` sets the scroll region to encompass all canvas items.

**Testing:** Build a large ARG. Confirm the scroll region covers all nodes and edges. Drag the scrollbar to the edges and confirm no nodes are missing.

### 8.8 Bezier Curve Edge Rendering
`TkCyCanvas.create_elements()` renders edges as Bezier curves approximated by multiple line segments via `approximate_cubic_bezier()`.

**Testing:** Confirm edges between non-adjacent nodes render as smooth curves. Confirm curves do not overlap nodes. For a two-node graph, confirm a straight or gently curved line.

### 8.9 Edge Arrow Shape
In `get_edge_styles()`, edges use `arrowshape="14 14 5"` (length 14, width 14, arrowhead width 5). Arrow is at the target end.

**Testing:** Confirm every ARG edge has a visible arrowhead at its target node. Verify the arrowhead size is consistent across all edges.

### 8.10 Dashed Edges for Cover Relations
In `get_edge_styles()`, edges of type `"cover"` use `dash=(5,5)`. Non-cover edges use a solid line.

**Testing:** Create a covering relation. Confirm the cover edge appears as a dashed line. Confirm regular transition edges are solid.

### 8.11 Edge Canvas Event Binding
Each edge canvas item has `Button-1` bound to `click_edge('left', ...)` and `Button-3` bound to `click_edge('right', ...)` via `tag_bind`.

**Testing:** Left-click on an edge. Confirm the left-click action fires. Right-click on an edge and confirm the context menu appears.

### 8.12 Node Canvas Event Binding
Each node shape has `Button-1` bound to `click_node('left', ...)` and `Button-3` bound to `click_node('right', ...)`.

**Testing:** Left-click a node. Confirm it triggers the view-state action (concept graph opens). Right-click a node and confirm a context menu appears.

### 8.13 Subgraph/Cluster Rectangle Rendering
Subgraph shapes from `CyElements.add_shape` are rendered as rectangles on the canvas, appearing behind nodes and edges, grouping nodes by cluster/isolate.

**Testing:** Open a file with isolates. Confirm a rectangle appears around nodes belonging to the same cluster. Confirm the rectangle does not obscure node labels.

---

## 9. Node Context Menu (Right-Click)

### 9.1 Context Menu Pop-Up via make_popup()
`TkCyCanvas.make_popup(event, actions, arg)` creates a Tk `Menu` at the event location and calls `post(event.x_root, event.y_root)`. The menu is destroyed after selection.

**Testing:** Right-click any node. Confirm a context menu appears near the cursor. Click outside to dismiss it. Confirm the menu disappears without taking action.

### 9.2 Auto-Execute Single Action with '<>' Label
If the action list has exactly one item with label `'<>'`, `make_popup` executes that action immediately without showing a menu.

**Testing:** Identify a case where a single auto-execute action is defined. Confirm clicking the target executes immediately without showing a menu. Verify by checking for side effects.

### 9.3 Separator via '---' Label
Items with label `'---'` add a `Menu.add_separator()`.

**Testing:** Right-click a node and confirm visual dividers appear between logical menu groups. Confirm clicking a separator does nothing.

### 9.4 Label-Only Menu Items (None Command)
Items with `command=None` are added as disabled labels. They serve as section headers.

**Testing:** Right-click a node with available state equations. Confirm "Execute actions" appears as a non-clickable header label, visually distinct from clickable items.

### 9.5 Node Context Menu: Execute Actions
The node context menu shows all available state equations at the top labeled with `state_equation_label(a)`. Clicking executes `do_state_action(a, node)`.

**Testing:** Right-click a node with applicable actions. Confirm action names appear in the menu. Click one and confirm a new child state is added to the ARG.

### 9.6 Node Context Menu: Check Safety
"Check safety" calls `check_safety_node(node)`. In Bounded mode this runs BMC; in other modes it runs local safety checking.

**Testing:** Right-click a node → "Check safety". Confirm the appropriate safety check runs. For a safe node, confirm no error dialog. For an unsafe node, confirm an error dialog.

### 9.7 Node Context Menu: Extend
"Extend" calls `find_extension(node)`. If applicable, executes an action to extend the ARG. If no action is applicable, shows "State {id} is closed."

**Testing:** Right-click a node → "Extend". If the node has applicable actions, confirm a new child state appears. If fully extended, confirm the "closed" ok_dialog appears.

### 9.8 Node Context Menu: Mark
"Mark" calls `mark_node(node)`. The previously marked node is unhighlighted, and the new node gets red fill via `show_mark(True)`.

**Testing:** Mark a node. Confirm it turns red. Mark a different node. Confirm the first returns to normal and the second turns red. Confirm only one node is red at a time.

### 9.9 Node Context Menu: Cover by Marked
"Cover by marked" calls `cover_node(covered_node)` using the marked node as covering. On success, a cover edge is added; on failure, an error message is shown.

**Testing:** Mark node A, right-click node B → "Cover by marked". If A subsumes B, confirm a dashed cover edge appears. If not, confirm an error dialog. Confirm the graph rebuilds on success.

### 9.10 Node Context Menu: Join with Marked
"Join with marked" calls `join_node(node2)`. The marked node and right-clicked node are joined, replacing both in the ARG.

**Testing:** Mark node A, right-click node B → "Join with marked". Confirm both A and B are replaced by a single joined state. Confirm the joined state is the abstract union of A and B.

### 9.11 Node Context Menu: Try Conjecture
"Try conjecture" calls `try_conjecture(node)`. If multiple conjectures exist, a listbox dialog appears for selection.

**Testing:** Right-click a node → "Try conjecture". If conjectures exist, confirm a listbox dialog lists them. Select one; confirm the proof attempt begins. If no conjectures, confirm a message is shown.

### 9.12 Node Context Menu: Try Remembered Goal
"Try remembered goal" calls `try_remembered_graph(node)`. If multiple remembered goals exist, a listbox dialog appears.

**Testing:** Save a proof goal first. Then right-click a node → "Try remembered goal". Confirm a listbox dialog shows saved goals. Select one and confirm the concept graph is set up.

### 9.13 Node Context Menu: Delete
"Delete" calls `delete_node(node)`. The node and all dependents are removed from the ARG.

**Testing:** Delete a leaf node and confirm it disappears. Delete a node with children; confirm children are also removed. Confirm the graph is valid and consistent after deletion.

### 9.14 Left-Click on Node: View State
A left-click returns a `view_state` action. `view_state(n, clauses, reset)` shows the state in an existing concept graph widget or creates a new one in the right pane.

**Testing:** Left-click a node. Confirm the right pane populates with a concept graph. Left-click another node; confirm the concept graph updates.

---

## 10. Edge Context Menu (Right-Click)

### 10.1 Edge Right-Click: Dismiss
"Dismiss" calls `decompose_edge(transition)` and removes the edge. If the edge cannot be decomposed, an `IvyError` is raised and shown in an error dialog.

**Testing:** Right-click an edge → "Dismiss". Confirm the edge disappears. If non-decomposable, confirm an error dialog. Confirm the ARG is valid after dismissal.

### 10.2 Edge Right-Click: Recalculate
"Recalculate" calls `recalculate_edge(transition)`. The edge is recomputed from its source state using the associated action.

**Testing:** Right-click an edge → "Recalculate". Confirm the edge and target state are recomputed. Confirm the graph rebuilds.

### 10.3 Edge Right-Click: Step In (Decompose)
"Step in" calls `decompose_edge(transition)`. If the action can be decomposed, a new tab is added showing the decomposed ARG.

**Testing:** Right-click an edge whose action is composite → "Step in". Confirm a new tab appears with the decomposed action ARG. Confirm sub-actions are visible as separate transitions.

### 10.4 Edge Right-Click: View Source
"View source" calls `view_source_edge(transition)`. This opens the file browser pointing to the source file and line number of the action.

**Testing:** Right-click an edge → "View source". Confirm the file browser window appears. Confirm the source file is loaded and the relevant line is highlighted in red.

---

## 11. Node Marking Behavior

### 11.1 Mark Visualization (Red Fill)
`show_mark(on=True)` sets the fill of all shapes tagged with the marked node's tag to `'red'`. `show_mark(on=False)` sets fill to `''` (transparent).

**Testing:** Mark a node and confirm it shows a red fill. Unmark it and confirm it returns to default fill. Confirm the red fill persists across rebuilds until explicitly cleared.

### 11.2 Mark Persistence Across Rebuilds
`mark_node(n)` stores the marked node in `self.mark`. When `rebuild()` runs, it calls `show_mark(on=True)` to restore the highlight.

**Testing:** Mark a node. Trigger recalculate-all (which rebuilds). Confirm the marked node is still red after the rebuild.

### 11.3 Single Mark at a Time
Only one node can be marked at a time. `mark_node(n)` calls `show_mark(False)` for the previous mark before setting the new one.

**Testing:** Mark node A, then mark node B. Confirm A is no longer red, only B is red. Repeat for a third node.

---

## 12. Safety Checking

### 12.1 Check Safety Dispatch
`check_safety_node(node)` checks `self.mode`: if `"bounded"` calls `check_bounded_safety(node)`, otherwise calls `check_local_safety(node)`.

**Testing:** In Bounded mode, trigger "Check safety" and confirm BMC runs. In Induction mode, trigger it and confirm local safety checking runs.

### 12.2 Bounded Safety Check
`check_bounded_safety(node)` calls `self.g.check_bounded_safety(node)`. If unsafe, sets `node.safe = False`, updates node color, and shows "The node is unsafe: View error trace?" as an ok_cancel dialog.

**Testing:** Create a safety violation. Select Bounded mode, right-click → "Check safety". Confirm the "unsafe" dialog appears. Click OK; confirm a new tab with the error trace ARG opens. Click Cancel; confirm no new tab.

### 12.3 Node Color Update After Safety Check
`update_node_color(node)` is called after `check_bounded_safety`. `node_color(node)` returns `"green"` if `node.safe`, `"black"` otherwise.

**Testing:** After a successful safety check, confirm the node's outline turns green. After a failed check, confirm it turns black. Confirm other nodes are unchanged.

### 12.4 Local Safety Check
`check_local_safety(node)` calls `self.g.check_safety(node)`. If unsafe, shows a `buttons_dialog_cancel` with "The node is not proved safe: {reason}", with buttons to view unsafe states or view error trace.

**Testing:** Create a local safety violation. Trigger local check. Confirm a buttons dialog appears. Click "View unsafe states" to see them. Click "View error trace" to confirm a trace dialog.

### 12.5 Safety Check: "View Error Trace" Button
When local safety fails, a button in the `buttons_dialog_cancel` views the error trace and adds a new ARG tab.

**Testing:** Trigger a local safety failure. In the resulting dialog, click the view-trace button. Confirm a new tab with the counterexample trace ARG appears.

---

## 13. Bounded Model Checking (BMC)

### 13.1 BMC Entry Point
`AnalysisGraphUI.bmc(state, err_cond, bound)` calls `self.g.bmc(state, err_cond, bound)`. If no counterexample, shows `ok_dialog` stating unreachable. Otherwise adds the result ARG as a new tab.

**Testing:** Run BMC on a safe property. Confirm the "unreachable" ok_dialog appears. Run on an unsafe property. Confirm a new tab with the counterexample ARG appears.

### 13.2 BMC Bound Specification
If no bound is provided in `bmc_conjecture()`, `int_dialog` asks the user for a bound value. The bound constrains the number of steps.

**Testing:** Trigger BMC without specifying a bound. Confirm an integer input dialog appears. Enter a valid integer and confirm BMC runs. Enter a non-integer; confirm an error dialog. Enter an out-of-range value; confirm an error.

### 13.3 BMC Result Display (Unreachable)
When no counterexample is found, `ok_dialog("State {id} is not reachable within {bound} steps")` is displayed.

**Testing:** Run BMC on a safe property with a small bound. Confirm the unreachable message dialog appears with the correct bound number.

### 13.4 BMC Result Display (Counterexample Found)
When BMC finds a counterexample, `view_ag(res)` adds the result ARG as a new tab showing the counterexample path.

**Testing:** Create a property violated in 1 step. Run BMC. Confirm a new tab appears with a 2-state ARG (initial + violating). Confirm each state shows concrete values.

---

## 14. Find Extension

### 14.1 Closed Node Dialog
If `find_extension(node)` finds no applicable state equations, it shows `ok_dialog("State {node.id} is closed.")`.

**Testing:** Fully extend a node until no more actions are available. Right-click → "Extend". Confirm the "closed" ok_dialog appears with the correct node ID.

### 14.2 Extension Execution
If applicable actions are found, `find_extension` calls `do_state_action(a, node)` for one action, adding a new child state.

**Testing:** Right-click an unexpanded node → "Extend". Confirm a new child state appears connected to the selected node by a transition edge.

---

## 15. Decompose Edge / Step In

### 15.1 Decompose Success: New Tab
`decompose_edge(transition)` on success adds a new `AnalysisSubgraph` to the UI via `view_ag(res)`, creating a new tab.

**Testing:** Right-click an edge whose action is compound → "Step in". Confirm a new tab appears with the decomposed ARG. Confirm sub-actions are visible as separate transitions.

### 15.2 Decompose Failure: Error Dialog
If `decompose_edge` raises `IvyError`, the `RunContext` catches it and shows an error dialog.

**Testing:** Right-click an edge whose action is atomic → "Step in". Confirm an error dialog appears. Click OK; confirm no new tab is created.

---

## 16. Conjecture Workflow

### 16.1 Try Conjecture: No Conjecture Given (List Dialog)
If `try_conjecture(node, conj=None)` is called without a conjecture, it shows a `listbox_dialog` "Choose a conjecture to prove:" listing all unproven conjectures.

**Testing:** With multiple unproven conjectures, right-click a node → "Try conjecture". Confirm a listbox dialog lists them. Select one and click OK; confirm the proof attempt starts. Cancel; confirm nothing happens.

### 16.2 Try Conjecture: Browse Source
After conjecture selection, `self.ui_parent.browse(filename, lineno)` opens the file browser at the conjecture's source declaration.

**Testing:** Select a conjecture. Confirm the file browser opens with the source file loaded and the conjecture line highlighted.

### 16.3 Try Conjecture: BMC Mode
In modes other than PDR and induction, `try_conjecture` calls `self.bmc(state, dual_of_conjecture, bound)`.

**Testing:** In Bounded mode, try a conjecture. Confirm BMC runs. Confirm either the counterexample ARG or the "unreachable" dialog appears.

### 16.4 Try Conjecture: Show Concept Graph
In Induction or PDR mode, `try_conjecture` calls `self.show_graph(sg)` to display a concept graph with the conjecture's abstraction.

**Testing:** In Induction mode, try a conjecture. Confirm the concept graph widget appears in the right pane. Confirm it includes the conjecture's predicates.

### 16.5 Remember Graph
`remember_graph(name, graph)` stores the graph in `self.remembered_graphs[name]` for later retrieval.

**Testing:** After a proof step, call remember_graph. Then right-click a node → "Try remembered goal" and confirm the saved graph name appears in the listbox.

### 16.6 Try Remembered Goal: List Dialog
If multiple goals are remembered, `try_remembered_graph(node, goal=None)` shows a `listbox_dialog` with their names.

**Testing:** Remember two different proof goals. Right-click a node → "Try remembered goal". Confirm both names appear. Select each and confirm the correct concept graph is loaded.

---

## 17. Interpolant / Refinement

### 17.1 Refine with Interpolant Dialog
`refine_with_interpolant(interp)` shows a dialog with the interpolant formula. In PDR mode: "PDR found the following invariant...". In other modes: "Found the following separating formula...". A "Refine" button is available.

**Testing:** Trigger an interpolant refinement step. Confirm a text dialog appears. In PDR mode confirm PDR-specific message text. Click "Refine" and confirm the concept space updates.

### 17.2 Pre-State Vacuous Message
If the pre-state is vacuous, `refine_with_interpolant` shows "The pre-state is vacuous..." instead of a formula.

**Testing:** Set up a vacuous pre-state and trigger refinement. Confirm the "vacuous" message appears in the dialog.

### 17.3 Add Predicate to Domain
`add_predicate(interp)` adds the interpolant as a predicate to the abstract domain, expanding the concept space.

**Testing:** Trigger automatic predicate addition. Confirm the concept graph updates to show the new predicate. Confirm the abstract domain includes the new concept.

---

## 18. Concept Graph Display

### 18.1 Concept Graph Creation via view_state()
`view_state(n, clauses, reset)` creates a concept graph from `self.g.concept_graph(n, ...)`. If `reset=True` or no existing concept graph, `self.show_graph(sg)` is called. Otherwise the existing concept graph updates.

**Testing:** Left-click a node. Confirm the concept graph appears. Click another node with `reset=False`; confirm the same widget updates. With `reset=True`; confirm a new concept graph replaces the old one.

### 18.2 Concept Graph Node Classes
Nodes have classes: `non_existing` (hidden), `exactly_one` (4px solid border), `at_least_one` (8px double border), `at_most_one` (3px dotted border), `node_unknown` (no border).

**Testing:** Open a concept graph with multiple sorts. Confirm nodes are styled according to cardinality. A singleton sort shows "exactly_one". An empty sort is hidden.

### 18.3 Concept Node Labels
Concept node labels use prefix notation: empty prefix for "necessarily true", `~` for "necessarily false", `?` for "unknown". Equality constraints show `=` or `≠`.

**Testing:** Open a concept graph. Confirm prefix characters (empty/~/?) appear on node labels. Confirm equality constraints show `=` or `≠` appropriately.

### 18.4 Concept Edge Classes
Edges have classes: `none_to_none` (dashed), `all_to_all` (solid), `edge_unknown` (dotted). Additional: `total` (circle source), `functional` (square source), `injective` (triangle-backcurve target), `surjective` (filled target).

**Testing:** Open a concept graph with a functional relation. Confirm a square source arrow appears. For a total relation, confirm a circle source arrow. Verify line styles match edge classes.

### 18.5 Concept Graph Background Color
The concept graph background is `rgb(192,192,255)` (light blue/periwinkle), distinguishing it from the white ARG canvas.

**Testing:** Open a concept graph. Confirm the background is light blue. Confirm the ARG canvas remains white.

### 18.6 Concept Node Shape
`get_shape(concept_name)` returns `'ellipse'` for most nodes and `'octagon'` for `'__ID'` concepts.

**Testing:** Open a concept graph for a state with an `__ID` relation. Confirm those nodes render as octagons. All other nodes should be ellipses.

### 18.7 Transitive Reduction of Edges
`get_transitive_reduction(widget, a, edges)` hides edges implied by transitivity. Self-edges on transitive relations are hidden first, then redundant transitive edges.

**Testing:** Add a transitive relation with nodes A→B, B→C, A→C. Confirm A→C is hidden. Disable the 'T' checkbox and confirm A→C reappears.

---

## 19. Concept Graph Controls

### 19.1 Relation Button Panel
`create_relbuttons_window()` creates a Tix `HList` frame with 5-column relation display controls per relation: `[+]`, `[?]`, `[-]`, relation-label, `[T]`. All are `Checkbutton` widgets bound to `IntVar`s.

**Testing:** Open the relation buttons panel. Confirm every relation appears as a row. Toggle `[+]` for a relation; confirm `all_to_all` edges appear/disappear. Toggle `[T]` for a transitive relation and confirm transitive reduction changes.

### 19.2 Relation Button: + (All-to-All)
The `+` button controls display of positive (`all_to_all`) edges for a relation.

**Testing:** Uncheck `+` for a relation with all-to-all edges. Confirm those edges disappear. Re-check and confirm they reappear.

### 19.3 Relation Button: ? (Unknown)
The `?` button controls display of `edge_unknown` edges (undetermined in current abstraction).

**Testing:** Create an unknown relation edge. Uncheck `?` for that relation and confirm dotted unknown edges disappear. Re-check and confirm they return.

### 19.4 Relation Button: - (None-to-None)
The `-` button controls display of `none_to_none` edges (no pairs hold). When checked, dashed "no edges" lines are shown.

**Testing:** Check `-` for a relation with known-absent edges. Confirm dashed lines appear. Uncheck and confirm they disappear.

### 19.5 Relation Button: T (Transitive)
The `T` button marks a relation as transitive for the transitive reduction algorithm.

**Testing:** Enable `T` for a transitive relation with redundant edges. Confirm transitive-closure edges are hidden. Disable `T` and confirm all edges including implied ones reappear.

### 19.6 Relation Display Checkboxes (IPython Widget Version)
`ConceptSessionControls` creates checkbox widgets for each relation/concept. Checking/unchecking updates domain concepts and triggers `recompute()` and `render()`.

**Testing:** Toggle a relation checkbox in the concept session widget. Confirm the concept graph updates immediately. Confirm the domain's active concepts change accordingly.

### 19.7 Edge Class Buttons (+, ?, -, ≤) in IPython Widget
`update_view_controls()` creates header buttons `+`, `?`, `-`, `≤` at the top of each edge group. Clicking toggles all checkboxes in that class. Logic: if all on → all off; else all on.

**Testing:** Click the `+` class button. Confirm all positive-edge checkboxes toggle off (if all were on) or all toggle on (if any were off).

### 19.8 Color Assignment per Concept/Relation
`TkGraphWidget.choose_colors()` assigns distinct colors to relations from 27 predefined color names, cycling when more than 27 relations are present.

**Testing:** Open a concept graph with multiple relations. Confirm each relation uses a distinct line color. With 28+ relations, confirm colors cycle without error.

### 19.9 Constraint Text in Concept Graph
`TkGraphWidget.rebuild()` renders constraint formulas as text items on the canvas. Selected constraints have black text; unselected have grey text. `left_click_constraint` toggles selection.

**Testing:** Open a concept graph with constraints. Confirm formula text appears on canvas. Click a formula text; confirm it toggles selected (black) / unselected (grey).

---

## 20. CTI (Counterexample to Induction) Widget

### 20.1 CTI Menu: Invariant Menu
The CTI `AnalysisGraphUI` adds an "Invariant" menu with: "Check induction", "Bounded check", "Diagram", "Weaken".

**Testing:** Open the CTI UI. Confirm the "Invariant" menu exists. Confirm all four items are present. Click each and confirm the corresponding function is invoked.

### 20.2 CTI Start: Auto-detect Transitive
`start()` calls `autodetect_transitive()` which scans the module for transitive relations. These are displayed with `'T'` in relation buttons.

**Testing:** Open an ivy file with transitive relations. Confirm the concept graph shows 'T' labels for those relations automatically on startup.

### 20.3 CTI Check Inductiveness
`check_inductiveness(button)` builds a proof ARG, checking each conjecture for inductiveness. Shows success dialog if all pass; error dialog if any fail. Sets `have_cti` flag.

**Testing:** With an inductive invariant, click Invariant → "Check induction". Confirm a success dialog appears. With a non-inductive conjecture, confirm an error dialog with the CTI. Confirm the concept graph updates.

### 20.4 CTI Success Dialog: Invariant Text
When all conjectures are proved inductive, `ok_dialog` shows the full invariant text (all conjectures joined with newlines).

**Testing:** Prove an inductive invariant. Confirm the success dialog contains the full invariant text as a readable formula.

### 20.5 CTI Failure Dialog: Counterexample Description
When a CTI is found, `ok_dialog` shows "An assertion failed..." with a description. The `have_cti` flag is set to True.

**Testing:** Use a non-inductive conjecture. Click "Check induction". Confirm the dialog message contains the assertion failure. Confirm `have_cti` is set.

### 20.6 CTI: Set Pre/Post States
`set_states(s0, s1)` sets the pre and post states for the concept graph. The concept graph shows both states for CTI analysis.

**Testing:** After a CTI check fails, confirm the concept graph shows both the pre-state and post-state for the failed conjecture.

### 20.7 CTI Weaken Invariant
`weaken(conjs, button)` shows a multi-select `listbox_dialog` listing all conjectures. Selected ones are removed. A confirmation `ok_dialog` shows the remaining invariant.

**Testing:** Click Invariant → "Weaken". Confirm a multi-selection dialog appears. Select some conjectures and click OK. Confirm they are removed. Confirm the confirmation dialog shows remaining invariant text.

### 20.8 CTI Save Invariant
`save_conjectures()` writes the conjecture set to a `.ivy` file. Old, new, and dropped conjectures are tracked and annotated with comments.

**Testing:** Make changes to the conjecture set. Click File → "Save invariant". Provide a filename. Confirm the file contains correct declarations and comments marking old/new/dropped.

### 20.9 CTI Bounded Check on Conjecture
`bmc_conjecture(button, bound)` runs BMC on the selected conjecture. If no bound, `int_dialog` asks. Results shown in `text_dialog`.

**Testing:** Select a conjecture. Click Invariant → "Bounded check". Enter a bound. Confirm a text dialog shows trace (if counterexample) or "safe within N steps" message.

### 20.10 CTI: Show Used Relations
`show_used_relations(clauses, both)` clears edge display then enables checkboxes for relations used in the CTI clauses.

**Testing:** After a CTI, confirm relation buttons show only relations relevant to the failed conjecture. Unrelated relations should be unchecked.

### 20.11 CTI Diagram View
`diagram()` creates a diagram visualization from the current abstract state. It calls `autodetect_inductiveness()` if needed and shows the result in the concept graph.

**Testing:** Click Invariant → "Diagram". Confirm the concept graph updates to show the abstract state diagram.

### 20.12 Conjecture Menu: Undo
The `ConceptGraphUI` Conjecture menu has "Undo" which calls `undo()` on the concept session.

**Testing:** Make a domain change (e.g., add a relation). Click Conjecture → "Undo". Confirm the domain reverts. Confirm the concept graph updates.

### 20.13 Conjecture Menu: Redo
"Redo" re-applies the last undone operation.

**Testing:** Undo a domain change, then click Conjecture → "Redo". Confirm the change is reapplied. Confirm undo/redo history is maintained across multiple operations.

### 20.14 Conjecture Menu: Gather
"Gather" calls `gather_facts(button)`. It collects all visible facts from selected nodes and edges.

**Testing:** Select nodes and edges in the concept graph. Click Conjecture → "Gather". Confirm gathered facts appear in the facts list, reflecting visible properties.

### 20.15 Conjecture Menu: Minimize
"Minimize" calls `minimize_conjecture(button)`. Reduces selected facts to minimal unsat core. A `text_dialog` shows the result.

**Testing:** Select more facts than necessary. Click Conjecture → "Minimize". Confirm a text dialog shows a smaller set that still achieves the proof goal.

### 20.16 Conjecture Menu: Check Sufficient
"Check sufficient" calls `is_sufficient(button)`. Checks if active conjecture implies target. Text_dialog shows result.

**Testing:** Select an insufficient set of facts. Click Conjecture → "Check sufficient". Confirm a dialog says the conjecture is not sufficient with an option to view counterexample. Select sufficient facts; confirm a success message.

### 20.17 Conjecture Menu: Check Relative Induction
"Check relative induction" calls `is_inductive(button)`. Checks inductiveness relative to current invariant.

**Testing:** Select a relatively inductive conjecture; confirm a success message. Select a non-inductive one; confirm a counterexample message with an option to view it.

### 20.18 Conjecture Menu: Strengthen
"Strengthen" calls `strengthen(button)`. Shows a `text_dialog` with the conjecture labeled "Add conjecture". If confirmed, the conjecture is added.

**Testing:** Select facts forming a valid conjecture. Click Conjecture → "Strengthen". Confirm a text dialog shows the formula. Click "Add conjecture" and confirm it is added to the invariant.

### 20.19 Conjecture Menu: Export
"Export" exports the concept graph as a DOT file via a save dialog.

**Testing:** Click Conjecture → "Export". Confirm a save-as dialog appears. Save to a `.dot` file. Confirm the file contains valid DOT graph syntax.

### 20.20 View Menu: Add Relation
"Add relation" shows an `entry_dialog` asking for a relation name to add to the concept graph.

**Testing:** Click View → "Add relation". Confirm an entry dialog appears. Enter a valid relation name; confirm it appears in the concept graph. Enter an invalid name; confirm an error dialog.

### 20.21 Concept Node Right-Click: Projections Menu
`ConceptGraphUI.get_node_actions()` right-click returns a "Projections" cascade menu listing applicable projections. Clicking calls `add_projection()`.

**Testing:** Right-click a concept node with ternary relations. Confirm a "Projections" menu appears. Click a projection item and confirm it is added to the concept graph.

### 20.22 Concept Node Left-Click: Select
Left-click on a concept node calls `select(node)`. This selects the node (grey fill in Tk).

**Testing:** Left-click a concept node. Confirm it becomes visually selected. Left-click another node and confirm the first is deselected. Confirm selected nodes contribute to "Gather" operations.

### 20.23 Node Context Action: Remove Concept
Right-clicking a concept node offers "remove". This calls `remove_concepts(concept)`.

**Testing:** Right-click a concept node → "remove". Confirm the node disappears. Confirm undo restores it.

### 20.24 Node Context Action: Suppose Empty
"suppose_empty" calls `suppose_empty(concept)`. Adds an emptiness constraint.

**Testing:** Right-click a node → "suppose empty". Confirm the node's cardinality changes to indicate emptiness. Confirm the suppose constraint is added.

### 20.25 Node Context Action: Materialize
"materialize" calls `materialize_node(concept_name)`. Creates a concrete witness element.

**Testing:** Right-click an abstract node → "materialize". Confirm a new concrete witness node appears in the concept graph.

### 20.26 Node Context Action: Split
"split by {label}" actions call `split(concept, by)` to refine the concept space.

**Testing:** Right-click a node with applicable label concepts. Confirm "split by {label}" items appear. Click one; confirm the node is split into sub-nodes.

### 20.27 Edge Context Action: Materialize+/−
Edge context menus offer "materialize +" (positive edge fact) and "materialize -" (negative edge fact).

**Testing:** Right-click an unknown edge → "materialize +". Confirm the edge class changes to all_to_all. Right-click → "materialize -". Confirm it changes to none_to_none.

### 20.28 Get Selected Conjecture
`get_selected_conjecture()` builds a positive universal conjecture from selected facts. Numerals are substituted with universally quantified variables.

**Testing:** Select facts in the concept graph. Call get_selected_conjecture. Confirm the result is a universally quantified formula. Confirm numerals are replaced by variables.

---

## 21. Session History Navigation (Jupyter Widget)

### 21.1 History Navigation Buttons
`AnalysisSessionWidget` has "first", "prev", "next", "last" buttons that update `current_step` and re-render.

**Testing:** Navigate through a multi-step proof. Click "next" and confirm the display advances one step. Click "prev" and confirm it goes back. Click "first" and "last" buttons and confirm they go to step 0 and the latest.

### 21.2 Step Info Display
The `step_box` Textarea displays `self.step_info` for the current step.

**Testing:** Navigate through proof steps. Confirm the step_box text changes at each step and accurately reflects the operation performed.

### 21.3 Auto-Click Active Element
`step()` calls `auto_click` on the active element at the current step, automatically triggering appropriate click handlers.

**Testing:** Navigate to a step with an active element. Confirm the concept widget updates automatically without user interaction beyond navigation.

### 21.4 Modal Message on Step
If not `silent`, `step()` sends `new_message(title, body)` to `ModalMessagesWidget`.

**Testing:** Navigate to a step with an associated message. Confirm a modal dialog appears with the correct title and body. Dismiss it and continue navigation.

---

## 22. Dialog System

### 22.1 OK Dialog
`ok_dialog(msg)` creates a `Toplevel` with a `Label` and "OK" button. Centered on parent. Blocks with `wait_window(dlg)`.

**Testing:** Trigger any OK dialog. Confirm a centered popup with the message and single OK button appears. Click OK; confirm it closes and the application continues.

### 22.2 OK Dialog: Test Automation Answer
If `self.answers` is non-empty, `getans()` pops the last answer and auto-clicks the appropriate button.

**Testing:** Call `ui.answer("ok")` before triggering an ok_dialog. Confirm the dialog auto-closes without user interaction.

### 22.3 OK/Cancel Dialog
`ok_cancel_dialog(msg, cmd)` creates a `Toplevel` with `Label`, "OK" button (calls `cmd` then closes), and "Cancel" button (just closes).

**Testing:** Trigger a safety check failure showing "View error trace?". Click OK; confirm the error trace opens. Click Cancel; confirm no trace. Test with `ui.answer("ok")` and `ui.answer("cancel")`.

### 22.4 Listbox Dialog (Single Selection)
`listbox_dialog` with `multiple=False` creates a `Toplevel` with `Label`, `Scrollbar`, `Listbox` (height=8, width=50, selectmode=SINGLE), "OK" button, and optional "Cancel". OK calls `command(selection_index)`.

**Testing:** Trigger a conjecture selection dialog. Confirm the listbox shows all items. Select one and click OK; confirm command is called with the correct index. Click Cancel; confirm `on_cancel` is called.

### 22.5 Listbox Dialog (Multiple Selection)
With `multiple=True`, the listbox uses `selectmode=EXTENDED`. OK calls `command(list_of_indices)`.

**Testing:** Trigger a "Weaken" action showing a multi-select dialog. Select multiple items using Ctrl+click/Shift+click. Click OK; confirm all selected indices are passed. Confirm Extended selection mode works.

### 22.6 Listbox Dialog Centering
`center_window_on_window(toplevel, win)` centers the dialog over its parent window.

**Testing:** Open a dialog from different main window positions. Confirm the dialog consistently appears centered over the parent.

### 22.7 Text Dialog
`text_dialog` creates a `Toplevel` with `Label`, `Scrollbar`, `Text` widget (height=4, width=100), "OK" (or `command_label`) button, and optional "Cancel". Initial text is pre-inserted.

**Testing:** Trigger a "Strengthen" action dialog. Confirm the text widget is pre-populated with the conjecture text. Edit it and click OK; confirm modified text is passed. Confirm the text widget is scrollable.

### 22.8 Entry Dialog
`entry_dialog` creates a `Toplevel` with `Label`, `Entry` widget, "OK" button, and optional "Cancel". The Entry gets focus. `<Return>` triggers OK.

**Testing:** Trigger an entry dialog. Confirm the entry widget has focus immediately. Press Enter; confirm OK is triggered. Delete and re-type, then click OK; confirm the command receives the text.

### 22.9 Integer Dialog
`int_dialog` wraps `entry_dialog` with a validator that checks for int in `[minval, maxval]`. Non-integers and out-of-range values show `IvyError` dialogs.

**Testing:** Trigger a BMC bound dialog. Enter "abc"; confirm "not an integer" error. Enter "-1" for a positive-only bound; confirm "out of range" error. Enter a valid integer; confirm the command fires.

### 22.10 Buttons Dialog (Custom Buttons)
`buttons_dialog_cancel` creates a dialog with `Label`, one button per `(label, callback)` pair, and a "Cancel" button.

**Testing:** Trigger a local safety failure showing a buttons dialog. Confirm each labeled action appears. Click one; confirm its callback fires. Click Cancel; confirm `on_cancel` is called.

### 22.11 Save As Dialog
`saveas_dialog(msg, filetypes)` calls `tkinter.filedialog.asksaveasfile(...)`. Returns file object if confirmed, `None` if cancelled.

**Testing:** Click File → Save. Confirm the native OS save-as dialog appears with correct title and file type filter. Save a file; confirm the path is returned. Cancel; confirm `None`.

### 22.12 Dialog Window Centering
`center_window(toplevel)` centers a dialog on screen by computing center offset from geometry.

**Testing:** Open a standalone dialog (using `center_window`). Confirm it appears in the center of the screen regardless of main window position.

### 22.13 Error Dialog from RunContext
`RunContext.__exit__` catches `IvyError` and shows a `Toplevel` with the error message and "OK" button. Returns True to suppress the exception.

**Testing:** Trigger an operation raising `IvyError` (e.g., invalid covering). Confirm an error dialog appears. Click OK; confirm the application continues. Trigger a non-IvyError and confirm it propagates.

### 22.14 Busy Cursor During Long Operations
`RunContext.__enter__` calls `parent.busy()` which sets `cursor='watch'` on `tk` and `frame` and calls `tk.update()`.

**Testing:** Trigger a slow operation. Confirm the cursor changes to a watch/hourglass during processing. Confirm it returns to normal when done.

### 22.15 Ready Cursor After Operation
`RunContext.__exit__` calls `parent.ready()` which sets `cursor=''` on both widgets.

**Testing:** After a long operation, confirm the cursor returns to the default arrow. If an exception occurs, confirm the cursor still returns to normal.

---

## 23. File Browser Widget

### 23.1 File Browser Window Creation
`new_file_browser(tk)` creates a `Toplevel` with a `FileBrowser` frame. Contains `Text` widget (width=100, height=20) and vertical `Scrollbar`.

**Testing:** Trigger any "View source" action. Confirm a file browser window appears with a scrollable text area. Confirm it can be independently resized, moved, and closed.

### 23.2 File Loading into Browser
`FileBrowser.set(filename, lineno)` opens and reads the file only if `filename` differs from the currently loaded file. The entire file is inserted into the Text widget.

**Testing:** Open the browser for file A. Trigger "View source" for file B. Confirm B's contents replace A's. Trigger "View source" for file A again; confirm A is reloaded.

### 23.3 Line Highlight (Red Background)
`set()` configures a `'highlight'` tag with `background='red'` and adds it to line `lineno`. Previous highlight is removed before adding the new one.

**Testing:** Open the browser pointing to line 10. Confirm line 10 has a red background. Navigate to line 25 in the same or different file. Confirm the highlight moves to line 25 and line 10 is no longer highlighted.

### 23.4 Scroll to Highlighted Line
After highlighting, `set()` scrolls the Text widget to the highlighted line using `see(lineno)`.

**Testing:** Open a file with 1000 lines pointing to line 900. Confirm the text widget scrolls so line 900 is visible without manual scrolling.

### 23.5 File Browser Window Raise
`set()` calls `self.lift()` on the parent window to bring it to the foreground.

**Testing:** Open the file browser, then bring the main window to front. Trigger "View source". Confirm the file browser window rises to the top of the window stack.

---

## 24. Trace Display

### 24.1 Trace Serialization to Lines
`Trace.to_lines()` converts the trace to a multi-line string with indented nested structure. State labels, action names, and parameter values are formatted with indentation for sub-calls.

**Testing:** Generate a counterexample trace. Confirm the trace is displayed as indented text. Confirm sub-actions are indented relative to their caller. Confirm parameter values appear inline.

### 24.2 Infinite Loop Detection in Trace
If the trace contains a repeating cycle, "--- the following repeats infinitely ---" is inserted before the cycle.

**Testing:** Create a trace with an intentional cycle. Confirm the "repeats infinitely" marker appears exactly once. Confirm the trace display terminates.

### 24.3 Function Call Formatting
`do_return()` in Trace formats function call returns with parameter values shown inline.

**Testing:** Generate a trace involving a function call. Confirm the trace shows both the call (with arguments) and the return (with return value), with proper indentation nesting.

### 24.4 Trace Line Numbers
`label_from_action(action)` uses `action.lineno` to show source line numbers alongside action names.

**Testing:** Generate a trace and confirm action lines include the source file line number. Click "View source" and confirm it opens at the correct line.

### 24.5 Trace: Detailed Option
`option_detailed` is a `BooleanParameter("detailed")` controlling trace verbosity. When True, additional details are included in `to_lines()`.

**Testing:** Set `--detailed=true` before generating a trace. Confirm more details appear (e.g., full formula for each state). Compare with `--detailed=false` and confirm reduced output.

---

## 25. Event Viewer (ivy_show.py)

### 25.1 EventTree Widget
`EventTree` inherits from `tkinter.tix.Tree` and `WithMenuBar`. It displays hierarchical events with lazy loading via `opencmd` and browsing via `browsecmd`.

**Testing:** Open the event viewer with a trace file. Confirm a tree widget appears. Click to expand items. Confirm child items load lazily when the parent is expanded.

### 25.2 Event NoteBook
`EventNoteBook` manages multiple sheets of event trees. Each `new_sheet(evs)` call creates a new tab.

**Testing:** Load a trace with multiple event types. Confirm separate sheets appear for different event categories. Confirm each sheet has its own independent tree.

### 25.3 Event Filter Dialog
The Events menu "Filter..." shows an `ask_pat` entry dialog. The pattern is used to filter events shown in the tree.

**Testing:** Click Events → "Filter...". Enter a regex pattern. Confirm only matching events appear. Enter an empty pattern and confirm all events return.

### 25.4 Find Reverse Dialog
"Find reverse..." opens an entry dialog for pattern input and searches events in reverse chronological order.

**Testing:** Click Events → "Find reverse...". Enter a pattern. Confirm matching events are found in reverse order. Confirm the search terminates on large logs.

### 25.5 Pattern List
`PatternList` is a frame with a `TList` widget for saved patterns. Buttons `+` and `-` add/remove patterns. `<<` reverses a pattern, `>>` forwards it.

**Testing:** Click `+` to add a pattern. Confirm it appears in the TList. Select and click `-`; confirm removal. Click `<<` on a pattern to confirm it is reversed (e.g., A→B becomes B→A).

### 25.6 Pattern List Save/Load/Clear
The PatternList has "Save", "Load", and "Clear" buttons.

**Testing:** Add patterns. Click "Save" and provide a filename. Click "Clear"; confirm the list is empty. Click "Load" with the saved file; confirm patterns are restored. Verify saved patterns survive restart.

### 25.7 HList Selection Style
The `HList` in EventTree uses `red` foreground for selected items.

**Testing:** Select an event. Confirm the selected item shows red text. Deselect and confirm text returns to default color.

---

## 26. Graph Rendering Details

### 26.1 Graph Layout Algorithm (DOT)
`dot_layout(cy_elements)` uses Graphviz `dot` for DAG layout. Nodes are sorted topologically with transitive relations considered.

**Testing:** Create an ARG with multiple nodes and edges. Confirm the layout is a proper DAG (root at top, leaves at bottom). Confirm no node overlaps. Confirm layout is stable for the same structure.

### 26.2 Y-Coordinate Inversion
`dot_layout` flips y-coordinates: `y = y_origin - node.y`. Graphviz uses bottom-left origin; this converts to top-left.

**Testing:** Confirm the ARG root (initial state) appears at the top and child states appear below. This would be reversed if y-inversion were absent.

### 26.3 Back Edge Reversal
If `node_gt(source, target)` returns True, the edge is reversed in Graphviz with `dir='back'`, and the spline points are reversed on output.

**Testing:** Create a covering relation or back edge. Confirm the arrow direction is correct in the displayed graph. Confirm edge rendering looks correct visually.

### 26.4 Cluster/Subgraph Boxes
`dot_layout` with `subgraph_boxes=True` creates Graphviz subgraphs for clusters, appearing as boxes around grouped nodes.

**Testing:** Open a file with declared isolates. Confirm rectangular boxes appear around related node groups. Confirm boxes do not overlap node labels.

### 26.5 Edge Weight Configuration
`weight` dict in `dot_layout` assigns weight 10 to `reach`/`le` edges (straighter layout) and weight 1 to `id` edges.

**Testing:** Create a graph with both reach/le and id edges. Confirm reach/le edges are rendered more vertically due to higher weight.

### 26.6 Pending Edge Constraint
`constraint` dict marks `pending` edges as `constraint=False`. These edges can cross without affecting the layout hierarchy.

**Testing:** Add a pending edge. Confirm it is laid out without distorting the DAG structure. Confirm the pending edge may cross other edges freely.

### 26.7 Node Width/Height from Graphviz
After layout, each node's `width` and `height` are extracted from Graphviz output and converted to pixels (`72 * graphviz_units`).

**Testing:** Inspect node dimensions after layout. Confirm nodes with longer labels get wider bounding boxes. Confirm the canvas renders nodes at these dimensions.

---

## 27. Diagnose Mode Behaviors

### 27.1 --diagnose Flag
When `--diagnose=true`, failing property checks open the GUI instead of raising exceptions. `check_properties()` shows the UI in Induction mode when properties fail.

**Testing:** Run `ivy --diagnose=true myfile.ivy` with failing properties. Confirm the GUI opens automatically in Induction mode. Pass `--diagnose=false` and confirm the error is printed and the process exits.

### 27.2 show_counterexample() Creates Trace ARG
`show_counterexample(ag, state, bmc_res)` builds a copy of the ARG path and assigns concrete state values from the BMC result. Then calls `gui_art(other_art)` to display it.

**Testing:** Trigger a BMC counterexample in diagnose mode. Confirm the counterexample path is shown as concrete states in the ARG. Confirm state values are populated from the BMC model.

### 27.3 check_conjectures() Shows CTI in Diagnose Mode
`check_conjectures(kind, msg, ag, state)` in diagnose mode opens the GUI with the ARG and calls `try_conjecture()` on the failed conjecture.

**Testing:** Run with a failing conjecture and `--diagnose=true`. Confirm the CTI UI opens with the ARG displayed. Confirm a dialog lists the failed conjectures.

### 27.4 try_property() in UI
`IvyUI.try_property(prop)` calls `bmc()` on the dual of the property. Without a property, a listbox dialog lists all properties.

**Testing:** With multiple properties, call `try_property()` without argument. Confirm a listbox dialog appears. Select one; confirm BMC runs on the negation of that property.

### 27.5 update_idletasks() Before Dialogs
`gui.tk.update_idletasks()` is called before dialogs in diagnose mode to ensure the main window renders before the dialog appears on top.

**Testing:** Trigger a diagnose-mode dialog. Confirm the main window is fully rendered before the dialog appears. Confirm the dialog appears above (not behind) the main window.

---

## 28. Jupyter Notebook Integration

### 28.1 ivy2.py: Notebook Generation
`ivy2.py:main()` reads a `.ivy` filename, generates a Jupyter notebook with Ivy analysis cells, and opens it. Generated cells include imports, module setup, session creation, and widget display.

**Testing:** Run `ivy2 myfile.ivy`. Confirm a `.ipynb` file is created. Confirm the notebook contains cells that set up the Ivy session. Confirm running the cells creates the analysis widgets.

### 28.2 PYTHONPATH Setting for Ivy
`ivy2.py` sets `PYTHONPATH` in the notebook environment to include the Ivy library location.

**Testing:** Open the generated notebook and run the import cell. Confirm no ImportError occurs. Confirm all Ivy modules are accessible.

### 28.3 CyGraphWidget JS Integration
`CyGraphWidget` communicates with its JavaScript counterpart via `send()` and `on_msg()`. Uses `_trait_to_json` and `_trait_from_json` to serialize Python objects to/from JSON while preserving object identity.

**Testing:** Create a CyGraphWidget, set elements, and confirm the JS view renders the correct graph. Modify elements and confirm the JS view updates. Verify object identity is preserved.

### 28.4 DialogWidget jQuery UI Integration
`DialogWidget` renders as a jQuery UI dialog. The `options` dict supports `'max'` for maximized dimensions and standard jQuery UI options.

**Testing:** Create a DialogWidget with `options={'height': 'max'}`. Confirm the dialog fills available height. Set a specific width; confirm the dialog respects it.

### 28.5 ModalWidget Bootstrap Integration
`ModalWidget` renders as a Bootstrap modal. `on_close(callback)` registers callbacks that fire with the button name when dismissed.

**Testing:** Create a ModalWidget and display it. Click OK; confirm callback fires with 'ok'. Click Cancel; confirm 'cancel'. Close with X; confirm callback fires appropriately.

### 28.6 ModalMessagesWidget: New Message
`ModalMessagesWidget.new_message(title, body)` sends JSON `{"method": "new_message", "title": title, "body": body}` to the JS view to display a modal message.

**Testing:** Call `new_message("Test Title", "Test body")`. Confirm a modal dialog appears in the browser with the correct title and body. Dismiss it; confirm no error occurs.

### 28.7 ExecuteNewCell Operation
`ExecuteNewCell(code)` queues Python code to execute in a new Jupyter cell. `submit(on_done)` registers a `post_run_cell` callback and executes the code.

**Testing:** Submit `ExecuteNewCell("2+2")`. Confirm a new cell appears in the notebook and executes. Confirm `on_done` is called with the result.

### 28.8 ShowModal Operation
`ShowModal(title, children)` creates a `ModalWidget` and displays it. `on_done` is called with True (OK) or False (Cancel).

**Testing:** Submit a `ShowModal`. Confirm the modal appears with correct title and children. Click OK; confirm `on_done(True)`. Click Cancel; confirm `on_done(False)`.

### 28.9 UserSelect Operation
`UserSelect(options, title, prompt)` shows a modal with a single-selection dropdown. The selected value is passed to `on_done`.

**Testing:** Submit a `UserSelect` with three options. Confirm a modal shows a dropdown. Select the second option and click OK. Confirm `on_done` receives the second option's value.

### 28.10 UserSelectMultiple Operation
`UserSelectMultiple(options, title, prompt)` shows a modal with `SelectMultiple`. A list of selected values is passed to `on_done`.

**Testing:** Submit with five options. Select two non-adjacent options. Click OK. Confirm `on_done` receives exactly those two options. Cancel; confirm `on_done(None)`.

---

## 29. Extension Points (ui_extensions_api.py)

### 29.1 ExtensionPoint Registration
`ExtensionPoint.register(function)` adds a callback. The `@ep.action("label")` decorator registers functions as UI actions and wraps them with `interaction()`.

**Testing:** Register a new callback on `arg_node_actions`. Trigger a node right-click. Confirm the registered action appears. Click it; confirm the callback fires. Unregister; confirm it no longer appears.

### 29.2 arg_node_actions Extension Point
`arg_node_actions(node)` collects context menu actions for ARG nodes from all registered callbacks. Default callbacks: `execute_actions` and `try_conjectures`.

**Testing:** Right-click an ARG node. Confirm default extension actions appear. Register a custom action; confirm it appears. Unregister; confirm it disappears.

### 29.3 goal_node_actions Extension Point
`goal_node_actions(node)` collects context menu actions for proof goal nodes. No default callbacks registered.

**Testing:** Register a custom callback. Right-click a proof goal node. Confirm the action appears. Unregister; confirm it disappears.

### 29.4 execute_actions Default Callback
The `execute_actions(s)` callback returns the list of applicable actions from the current analysis session.

**Testing:** Open a model with multiple actions. Right-click an ARG node. Confirm all applicable actions from the ivy module appear.

### 29.5 try_conjectures Default Callback
The `try_conjectures(s)` callback returns unproven conjectures applicable at state `s`.

**Testing:** Define multiple conjectures. Right-click an ARG node. Confirm unproven conjectures appear in the menu. Prove one; confirm it no longer appears.

### 29.6 @interaction Decorator
The `@interaction` decorator wraps generator functions to support async operations. It checks if the session is at the last history step.

**Testing:** Use a decorated interaction function that yields a `UserSelect`. Confirm the UI shows the selection dialog and waits for user input. Confirm the function resumes after selection.

### 29.7 arg_check_cover Action
`arg_check_cover(node)` validates exactly one node is selected (raises `InteractionError` otherwise), yields `UserSelect` for covering node, then yields `ExecuteNewCell` to call `check_cover()`.

**Testing:** Select a single node and invoke `arg_check_cover`. Confirm a selection dialog for covering node appears. Try with zero or two selected; confirm `InteractionError`.

### 29.8 arg_remove_facts Action
`arg_remove_facts(node)` yields `UserSelectMultiple` listing facts. User selects facts to remove, then `remove_facts()` is called in a new cell.

**Testing:** Invoke on a node with multiple facts. Confirm a multi-select dialog appears. Select some and click OK; confirm `remove_facts()` removes them. Cancel; confirm no facts are removed.

### 29.9 arg_join2 Action
`arg_join2(node)` validates selection, yields `UserSelect` for second join node, then yields `ExecuteNewCell` for `join2()`.

**Testing:** Select a node and invoke `arg_join2`. Confirm a selection dialog appears. Select a compatible node; confirm join2 executes. Try with invalid selection; confirm `InteractionError`.

---

## 30. Concept Session Behaviors

### 30.1 Domain Undo/Redo
`ConceptInteractiveSession.push()` saves the current domain to `undo_stack`. `undo()` calls `pop()` and recomputes.

**Testing:** Make a domain change (add concept). Click Undo. Confirm the domain reverts. Verify undo/redo history is maintained for multiple operations.

### 30.2 Recompute After Domain Change
After `undo()`, `split()`, `remove_concepts()`, `suppose_empty()`, `materialize_node()`, or `materialize_edge()`, `recompute()` is called. The widget then renders.

**Testing:** Make a domain change. Confirm the concept graph immediately re-renders. Confirm the new abstract value is reflected in node/edge classes and labels.

### 30.3 Save/Load Domain
`save_domain(name)` stores a snapshot. `load_domain(name)` restores it. Multiple snapshots can exist simultaneously.

**Testing:** Save the domain as "checkpoint1". Make changes. Load "checkpoint1". Confirm the domain reverts exactly to the saved state.

### 30.4 Replace Domain
`replace_domain(new_domain, new_suppose_constraints)` swaps the entire concept domain and suppose constraints. Used by "Reset domain" and "Diagram domain".

**Testing:** Click "reset domain". Confirm the domain reverts to the initial concept domain. Confirm suppose constraints are cleared.

### 30.5 Get Projections
`get_projections(node)` finds ternary relations where `node` appears and returns projection concepts for the context menu.

**Testing:** Open a concept graph for a model with ternary relations. Right-click a node. Confirm "Projections" submenu appears. Click a projection; confirm it is added to the display.

### 30.6 Add Custom Edge
`add_custom_edge(edge, source, target)` adds a specific (edge, source, target) combination to the domain display.

**Testing:** Add a custom edge. Confirm the new edge appears in the concept graph with the correct source and target nodes.

---

## 31. One-Step Reachability

### 31.1 one_step_reach()
`one_step_reach(state, clauses)` computes states reachable in one step from known reachable states. Adds new states to the reachable tree.

**Testing:** Call one_step_reach after marking some states reachable. Confirm new reachable states are added to the reachable tree. Confirm the display updates.

### 31.2 Eliminated Conjectures Dialog
If conjectures are eliminated (a reachable state violates them), a `listbox_dialog` shows the eliminated conjectures with only OK (no Cancel, since `on_cancel=None`).

**Testing:** Set up a conjecture violated by a reachable state. Call one_step_reach. Confirm a listbox dialog lists the eliminated conjectures. Click OK; confirm dialog closes.

---

## 32. Visual Styling Details

### 32.1 ARG Node: Green for Safe
`node_color(node)` returns `"green"` if `node.safe` is True. Used by `update_node_color()` to paint the node outline green.

**Testing:** Run a safety check that succeeds. Confirm the verified node's outline turns green. Confirm unsafe nodes remain black.

### 32.2 ARG Node: Black Default Outline
All ARG nodes start with `outline='black'` and `width=2` from `get_node_styles()`.

**Testing:** Open a fresh ARG. Confirm all node outlines are black with width 2.

### 32.3 Concept Node: Grey Background
Default concept graph node background is `#888` (grey). Text is white with a 3px `#888` text outline.

**Testing:** Open a concept graph. Confirm nodes have a grey background. Confirm node labels are white and readable.

### 32.4 Concept Node: Non-Existing Hidden
Nodes with class `non_existing` have `display: none` in `concept_style`. They are invisible and do not affect layout.

**Testing:** Create a scenario where a sort is empty. Confirm no node appears for that sort. Confirm no gap in the layout where the node would be.

### 32.5 ARG Bottom State: Black Background
Nodes with class `bottom_state` have `background: black` and `text-outline-color: black` in `arg_style`.

**Testing:** Trigger an error condition creating a bottom_state. Confirm it appears as a solid black node.

### 32.6 Proof Goal: Refuted State
Proof goal nodes with class `refuted` have `background: black` in `proof_style`.

**Testing:** Refute a proof goal via BMC. Confirm the corresponding node in the proof stack turns black.

### 32.7 Selection Overlay
`:selected` elements have `overlay-opacity: 0.2` in all style sheets.

**Testing:** Click to select a node or edge in a Cytoscape-based widget. Confirm a slight darkening/overlay appears on the selected element.

### 32.8 Transition Action Edge: Triangle Arrow
Transition action edges have `target-arrow-shape: triangle` and `target-arrow-fill: filled`.

**Testing:** Create an action transition. Confirm the edge terminates with a filled triangle arrowhead.

### 32.9 Transition Join Edge: Backcurve Arrow
Join transition edges have `target-arrow-shape: triangle-backcurve`.

**Testing:** Join two states. Confirm the resulting join edge has a concave (backcurve) arrowhead distinct from normal triangle arrows.

### 32.10 Cover Edge: Dashed, No Label
Cover edges have `line-style: dashed` and `content: ''` (no label).

**Testing:** Create a covering relation. Confirm the cover edge is dashed with no label text.

### 32.11 Edge Text Rotation
ARG edges have `edge-text-rotation: none`. Edge labels remain horizontal even on diagonal edges.

**Testing:** Create an ARG with angled edges. Confirm edge labels remain horizontal, not rotated.

### 32.12 Concept Edge: Total (Circle Source Arrow)
Edges with class `total` have `source-arrow-shape: circle`.

**Testing:** Add a total function concept. Confirm a circle appears at the source end. Confirm non-total functions lack this circle.

### 32.13 Concept Edge: Functional (Square Source Arrow)
Edges with class `functional` have `source-arrow-shape: square`.

**Testing:** Define a functional relation. Confirm a square appears at the source end.

### 32.14 Concept Edge: Injective (Backcurve Target)
Edges with class `injective` have `target-arrow-shape: triangle-backcurve`.

**Testing:** Define an injective relation. Confirm the target has a backcurve arrow.

### 32.15 Concept Edge: Surjective (Filled Target Arrow)
Edges with class `surjective` have `target-arrow-fill: filled` at the target.

**Testing:** Define a surjective relation. Confirm the target arrowhead is filled. Confirm non-surjective relations have an open arrow.

### 32.16 Node Text Wrapping
Concept graph nodes have `text-wrap: wrap`. Long labels are wrapped rather than overflowing.

**Testing:** Create a concept with a very long name. Confirm the label wraps to multiple lines. Confirm the node expands to accommodate wrapped text.

### 32.17 Font Size 14px
All concept graph nodes and edges use `font-size: 14px` in `concept_style`.

**Testing:** Open a concept graph and inspect font size. Confirm labels render at 14px consistently.

### 32.18 Concept Graph Line Colors (27 Colors)
`TkGraphWidget.line_colors` contains 27 named colors used cyclically for relation edge rendering.

**Testing:** Open a model with 28+ relations. Confirm colors cycle. With fewer relations, confirm each gets a distinct color.

---

## 33. Miscellaneous UI Behaviors

### 33.1 update_idletasks() Before Dialogs
`gui.tk.update_idletasks()` is called before dialogs to ensure the main window renders before the dialog appears on top.

**Testing:** Trigger a diagnose-mode dialog. Confirm the main window is fully rendered before the dialog appears above it.

### 33.2 Analysis Session Widget set_context()
`set_context(analysis_session_widget)` sets global context variables: `_analysis_session_widget`, `_analysis_session`, `_concept_session_widget`, `_concept_session`.

**Testing:** After `set_context()`, confirm the four globals are populated. Confirm extension callbacks can access the current session through these globals.

### 33.3 Proof Goal Click Updates Concept Widget
`AnalysisSessionWidget.proof_node_click(goal)` updates the concept session widget with the clicked proof goal's information.

**Testing:** Click a proof goal node in the proof graph. Confirm the concept widget updates to display the goal's constraints. Confirm different goals show different constraints.

### 33.4 CRG Node Click Updates Concept and Transition Widgets
`crg_node_click(crg_node)` updates both the concept domain and, if the node has a transition, the transition view widget's pre/post states.

**Testing:** Click a CRG node. Confirm the concept widget updates. If the node represents a transition, confirm the transition view shows pre/post states. Click a non-transition node; confirm no transition view update.

### 33.5 Abstractor Selection Dropdown
In `AnalysisSessionWidget`, a dropdown `select_abstractor` offers: 'top/bottom', 'concrete', 'propagate', 'propagate & conjectures', 'concept space'. Default: `'ta.Abstractors.top_bottom'`.

**Testing:** Select each abstraction strategy. Confirm analysis behavior changes accordingly. Select 'concrete' and confirm states are not abstracted.

### 33.6 BMC Bound Dropdown in Transition Widget
`TransitionViewWidget` has a `Dropdown` with BMC bound options [1, 3, 5, 10, 15], defaulting to 3.

**Testing:** Set the BMC bound dropdown to 5. Trigger "bmc conjecture". Confirm BMC runs with bound 5 without showing an int_dialog.

### 33.7 Relations to Minimize Text Input
`TransitionViewWidget` has a `Text` widget for relation names to minimize. Default: 'relations to minimize'.

**Testing:** Enter specific relation names in the field. Click "minimize conjecture". Confirm the minimization focuses on those relations.

### 33.8 Transition View Widget Log File
`TransitionViewWidget.register_session(session)` creates a log file named `tvw_log_{filename}`. Operations are logged with metadata.

**Testing:** Register a session. Perform several operations. Confirm a log file is created with entries for each operation.

### 33.9 Gather Facts: Node vs Edge Distinction
`gather_facts()` distinguishes single-element (node) tuples from three-element (edge) tuples. Node facts are unary predicates; edge facts are binary.

**Testing:** Select a mix of nodes and edges. Click "gather facts". Confirm the facts list includes both node facts (formatted as `P(x)`) and edge facts (formatted as `R(x,y)`).

### 33.10 Facts List: SelectMultiple Widget
The `facts_list` is a `SelectMultiple` widget. `get_active_facts()` returns selected facts, or all facts if none selected.

**Testing:** Gather multiple facts. Select a subset. Call `get_active_facts()` and confirm only the selected subset is returned. Deselect all; confirm all facts are returned.

### 33.11 Fact Highlighting in Concept Graph
`highligh_selected_facts()` collects atoms from selected facts and adds custom edges and node labels to highlight them in the pre/post graphs.

**Testing:** Select some facts in the transition view. Confirm corresponding nodes and edges in the concept graph become visually highlighted. Deselect facts; confirm highlights are removed.

### 33.12 apply_structure_renaming()
`apply_structure_renaming(st)` applies a dictionary of string substitutions to rename structures in the display.

**Testing:** Set a structure renaming (e.g., 'node_0' → 'client'). Confirm node labels show the renamed version. Confirm renaming is applied consistently.

### 33.13 Concept Domain Initial Snapshot
On startup, the initial concept domain is saved as `'initial'` in domain snapshots. "Reset domain" restores this snapshot.

**Testing:** Make concept domain changes. Click "reset domain". Confirm the domain reverts to startup state. Confirm all added concepts and suppose constraints are cleared.

### 33.14 Witness Constants with '@' Prefix
In `check_inductiveness()`, witness constants are created with a `'@'` prefix. These are unique constants for BMC witnesses.

**Testing:** Run inductiveness check. Confirm witness constants in CTI states are prefixed with `@`. Confirm they don't conflict with user-defined constants.

### 33.15 Autodetect Transitive Relations
`autodetect_transitive()` scans `im.module` for relations with transitive axioms. Found relations populate `self.transitive_relations`.

**Testing:** Create a model with a declared transitive relation. Open the CTI UI. Confirm the transitive relation shows 'T' in relation buttons without manual intervention.

### 33.16 Process Launch with xterm
`ivy_launch.py:run_in_terminal(cmd, name)` spawns `xterm` windows for distributed test processes with descriptive titles.

**Testing:** Call `run_in_terminal("echo hello", "test_proc")`. Confirm an xterm window appears with a title containing "test_proc". Confirm the command runs in that window.

### 33.17 Sequential Port Allocation
`ivy_launch.py:get_unused_port()` returns sequential ports starting at 49123, incrementing `next_unused_port` each call.

**Testing:** Call `get_unused_port()` three times. Confirm ports 49123, 49124, 49125 are returned.

### 33.18 Diagnose Mode: --diagnose Parameter
`diagnose = iu.BooleanParameter("diagnose", False)`. When True, failures open the GUI.

**Testing:** Pass `--diagnose=true` to ivy. Trigger a property failure. Confirm the GUI opens. Pass `--diagnose=false`; confirm the error is printed and the process exits.

### 33.19 ConceptStateViewWidget: Info Area
`ConceptStateViewWidget` has a `Textarea` for `info_area` (margin='5px') that displays current node/edge info.

**Testing:** Click a node in the concept widget. Confirm the info_area updates with the node's label and metadata. Click an edge; confirm edge info is shown.

### 33.20 ConceptStateViewWidget: State Text
A `Textarea` for `state_text` (margin='5px') shows the formula representation of the current state.

**Testing:** Select a node in the ARG. Confirm the state_text Textarea shows the formula of the selected state. Navigate between nodes and confirm the state text updates.

### 33.21 ConceptStateViewWidget: Constraints Text
A `Textarea` for `constraints_text` (margin='5px') shows the active suppose constraints.

**Testing:** Add a suppose_empty constraint. Confirm the constraints_text Textarea shows the constraint formula. Remove it; confirm the text clears.

### 33.22 TransitionViewWidget: Two Concept Graphs
`TransitionViewWidget` has two `CyGraphWidget` instances: `pre_graph` and `post_graph`, each with `cy_layout={'name': 'preset'}`.

**Testing:** After a CTI check failure, confirm two concept graphs appear side by side. Confirm the left shows the pre-state and the right shows the post-state. Confirm both graphs update when the transition changes.

### 33.23 AnalysisSessionWidget: Three Graphs
`AnalysisSessionWidget` has three `CyGraphWidget` instances: `proof_graph` (with `proof_style`), `arg` (with `arg_style`), `crg` (with `arg_style`).

**Testing:** Open the analysis session widget. Confirm three graph areas are visible. Confirm `proof_graph` shows the proof goal stack. Confirm `arg` shows the abstract reachability graph. Confirm `crg` shows the concrete reachability graph.

### 33.24 pre_state and post_state Initialization
`TransitionViewWidget` initializes `pre_state` and `post_state` as `(None, Or(), None)` tuples (None state, empty formula, None abstract value).

**Testing:** Open the transition view widget without selecting a CTI. Confirm both graphs render as empty (no nodes). Select a CTI; confirm both graphs populate with the CTI states.

### 33.25 Structure Renaming in TransitionViewWidget
`TransitionViewWidget.apply_structure_renaming(st)` applies string substitutions from `self.structure_renaming`. Used to rename witness constants for display.

**Testing:** Set `structure_renaming = {'@node_0': 'client'}`. Display a CTI with a witness constant `@node_0`. Confirm it appears as 'client' in the concept graph labels.

### 33.26 Concept Domain Saved as 'initial' on Register
When `TransitionViewWidget.register_session(session)` is called, it saves the domain as 'initial' via `concept_session.save_domain('initial')`.

**Testing:** Register a session. Make domain changes. Click "reset domain". Confirm the domain reverts to exactly the state at registration time.

### 33.27 Concept Domain Saved as 'diagram' for Diagram Mode
`diagram_domain()` in `ConceptSessionControls` calls `replace_domain(...)` with a diagram concept domain, replacing the current domain.

**Testing:** Click "diagram domain". Confirm the domain is replaced with the diagram-specific concepts. Confirm the concept graph re-renders with the new domain.

---

## Summary Table of Menu Items

| Menu | Item | Callback | Context |
|------|------|----------|---------|
| File | Save | save() | All |
| File | Save abstraction | save_abstraction() | All |
| File | *(separator)* | — | All |
| File | Remove tab | remove(self.g) | All |
| File | Exit | exit() | All |
| File | Save invariant | save_conjectures() | CTI |
| Mode | Concrete | mode=Concrete | Radiobutton |
| Mode | Abstract | mode=Abstract | Radiobutton |
| Mode | Bounded | mode=Bounded | Radiobutton |
| Mode | Induction | mode=Induction | Radiobutton |
| Mode | Pdr | mode=Pdr | Radiobutton |
| Action | Recalculate all | recalculate_all() | All |
| Action | Show reachable states | show_reachable_states() | All |
| Invariant | Check induction | check_inductiveness() | CTI |
| Invariant | Bounded check | bmc_conjecture() | CTI |
| Invariant | Diagram | diagram() | CTI |
| Invariant | Weaken | weaken() | CTI |
| Conjecture | Undo | undo() | Concept |
| Conjecture | Redo | redo() | Concept |
| Conjecture | Gather | gather_facts() | Concept |
| Conjecture | Bounded check | bmc_conjecture() | Concept |
| Conjecture | Minimize | minimize_conjecture() | Concept |
| Conjecture | Check sufficient | is_sufficient() | Concept |
| Conjecture | Check rel. induction | is_inductive() | Concept |
| Conjecture | Strengthen | strengthen() | Concept |
| Conjecture | Export | export() | Concept |
| View | Add relation | entry_dialog | Concept |
| Events | Filter... | ask_pat() | EventViewer |
| Events | Find reverse... | ask_pat(reverse) | EventViewer |

---

## Summary Table of Dialogs

| Dialog Type | Trigger | Message | Buttons |
|-------------|---------|---------|---------|
| ok_dialog | Closed node | "State {id} is closed." | OK |
| ok_dialog | BMC unreachable | "State {id} not reachable in N steps" | OK |
| ok_dialog | CTI success | "Inductive invariant found: {text}" | OK |
| ok_dialog | CTI failure | "An assertion failed..." | OK |
| ok_cancel_dialog | Bounded unsafe | "The node is unsafe: View error trace?" | OK, Cancel |
| buttons_dialog | Local unsafe | "The node is not proved safe: {}" | View states, View trace, Cancel |
| listbox_dialog | Try conjecture | "Choose a conjecture to prove:" | OK, Cancel |
| listbox_dialog | Try remembered goal | "Choose a remembered goal:" | OK, Cancel |
| listbox_dialog | one_step_reach | Eliminated conjectures | OK |
| listbox_dialog | Weaken | Conjecture list (multi-select) | OK, Cancel |
| text_dialog | Strengthen | Conjecture formula | Add conjecture, Cancel |
| text_dialog | BMC trace | Trace text | OK, Cancel |
| text_dialog | Is sufficient | Sufficient/not result | OK |
| text_dialog | Is inductive | Inductive/not result | OK |
| entry_dialog | Add relation | "Relation name:" | OK, Cancel |
| int_dialog | BMC bound | "Enter bound:" | OK, Cancel |
| saveas_dialog | Save analysis | "Save analysis state as..." | (OS dialog) |
| saveas_dialog | Save abstraction | "Save abstraction as..." | (OS dialog) |
| saveas_dialog | Export DOT | "Export graph as..." | (OS dialog) |
| error_dialog | IvyError | Error message | OK |
| text_dialog | Refine | Formula + PDR/induction message | Refine, Cancel |
| ok_dialog | Pre-state vacuous | "The pre-state is vacuous..." | OK |
