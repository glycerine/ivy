# Plan: Tcl GUI Behavior Inventory for Ivy

## Context

The Python/Tcl ivy tool has a rich interactive GUI built with Tkinter/Tix (Tk) that is being ported to a web UI. This plan captures the exhaustive inventory of all GUI behaviors from the Python source so that nothing is missed in the port. The inventory will be written to `~/ivy/goivy/tcl_gui_behavior_inventory_claude.md`.

## Approach

The inventory below was compiled by thoroughly reading all .py files in `~/ivy/pyivy/ivy/ivy/`, with special emphasis on the UI-relevant files. The output is organized by functional area (not by source file), with each behavior having a 2-3 sentence description and 3+ sentences on how to test it.

## Critical Files Modified

- New file: `~/ivy/goivy/tcl_gui_behavior_inventory_claude.md`

## Verification

After writing, verify:
1. File exists at destination path
2. File opens and renders as valid markdown
3. All major sections (application startup, menus, dialogs, graphs, CTI, etc.) are present

---

## Full Inventory Content to Write

The content below is the complete markdown for the inventory file.

---

# Ivy Tcl/Tk GUI Behavior Inventory

This document enumerates every behavior of the Ivy Tcl/Tk Python GUI (from `~/ivy/pyivy/ivy/ivy/`) for use in porting to a web interface. Each behavior has a description and testing guidance.

---

## 1. Application Startup & Initialization

### 1.1 Entry Point Execution
The main entry point (`ivy.py:main()`) reads command-line parameters via `ivy_init.read_params()`, loads the `.ivy` source file, and then calls `tk_ui.ui_main_loop(ag)` to display the GUI. The SIGINT signal is set to its default handler before launch so that Ctrl+C cleanly exits rather than raising a Python exception. The module context (`im.Module()`) wraps the entire session so that all loaded symbols are properly scoped.

**Testing:** Launch `ivy myfile.ivy` and confirm the GUI window appears. Verify that `Ctrl+C` at the terminal exits without a Python traceback. Confirm that a syntax error in `myfile.ivy` prints an error to stderr and exits without showing the GUI.

### 1.2 Tk Root Window Creation
`TkUI.__init__` creates a `tkinter.tix.Tk()` root window (not a plain `Tk()`—Tix extensions are required). The global Tk palette is set to white (`tk.tk_setPalette('white')`). The window title is set to `"ivy"` immediately after creation.

**Testing:** Observe that the window title bar reads "ivy". Confirm the background color of all widgets defaults to white. Confirm that Tix-specific widgets (NoteBook, HList, etc.) are available without error.

### 1.3 Window Title
The root window title is set to `"ivy"` via `self.tk.title("ivy")` in `TkUI.__init__`. No dynamic title changes happen during the session; the title remains "ivy" for the lifetime of the application.

**Testing:** Open the application and read the OS window title. Confirm it says "ivy". Maximize, minimize, and restore; the title should remain "ivy" throughout.

### 1.4 Tix NoteBook Initialization
A `tkinter.tix.NoteBook` widget is created as the only top-level child of the root and packed with `fill=BOTH, expand=1`. This notebook holds all analysis sheets as tabs. The `tab_counter` and `tabs` instance variables are initialized to 0.

**Testing:** Open a file and confirm a tabbed notebook widget appears filling the window. Resize the window; the notebook should resize with it. Verify the first tab is automatically selected.

### 1.5 Initial Sheet Creation
When `ui_main_loop(art)` is called, it calls `ui.add(art)` to add the first analysis graph as a tab named `"sheet_1"` with label `"Sheet 1"`. The new tab is immediately raised to the front. `gw.start()` is called on the newly created widget to initialize the display.

**Testing:** Open a file and confirm "Sheet 1" tab exists and is selected. The ARG graph should be visible. No empty or error state should be shown for a valid file.

### 1.6 DYLD_LIBRARY_PATH Setup (macOS)
`ivy_shell.py:main()` detects the macOS platform and exports `DYLD_LIBRARY_PATH` to include library paths needed for native dependencies (Z3, etc.). This happens before any GUI is shown. Without this, native library loading may fail on macOS.

**Testing:** On macOS, confirm that Z3 is loadable after the shell setup. Check `os.environ['DYLD_LIBRARY_PATH']` before and after `ivy_shell.main()` and confirm the paths were added.

---

## 2. Main Window Layout

### 2.1 PanedWindow Split View per Tab
Each tab contains a horizontal `PanedWindow` with two panes: a left pane (minimum width 50px) for the ARG graph canvas, and a right pane (minimum width 200px) for the state/concept graph display. The panes are resizable by the user dragging the sash.

**Testing:** Open an ivy file and confirm two side-by-side panes. Drag the sash left and right and confirm both panes resize. Confirm neither pane can be collapsed below its minimum size.

### 2.2 Left Pane: ARG Graph Canvas
The left pane contains the Analysis/Reachability Graph (ARG) canvas with both horizontal and vertical scrollbars. The canvas has the ARG graph widget filling the pane.

**Testing:** Open a file with multiple reachable states. Confirm the ARG is visible in the left pane. If the graph is larger than the pane, confirm scrollbars appear and scrolling works.

### 2.3 Right Pane: State Display Frame
The right pane is a frame (`f2`) that holds the state/concept graph widget when a node is clicked. Initially it may be empty or contain a placeholder. When a node is selected, a concept graph widget populates this frame.

**Testing:** Open a file, click a node in the ARG. Confirm the right pane populates with a concept graph. Confirm scrollbars in the right pane work if the concept graph is large.

### 2.4 Menu Bar Placement
The menu bar (`MenuBar`) is a Frame packed at the top of the analysis graph widget frame. Menu buttons are packed left-to-right. An empty frame is packed to the right to push buttons left.

**Testing:** Confirm menu buttons (File, Mode, Action, etc.) appear horizontally at the top of each tab's graph widget. Add more menu items and confirm they appear left of the spacer. Resize the window narrower and confirm buttons are not cut off.

### 2.5 Scrollbars for ARG Canvas
Two scrollbars are created: one vertical (right side) and one horizontal (bottom). They are linked to the graph canvas. The `xview` and `yview` commands of the canvas are bound to the scrollbars.

**Testing:** Create a large ARG with many nodes. Confirm both scrollbars appear. Drag the horizontal scrollbar and confirm the graph pans left/right. Drag the vertical scrollbar and confirm up/down panning.

---

## 3. Tab/Sheet Management

### 3.1 Adding a New Sheet Tab
`TkUI.add(art, name, label, ui_class)` creates a new tab in the NoteBook. `tab_counter` increments on each call. Default name is `"sheet_{N}"`, default label is `"Sheet {N}"`. The new tab is raised to front with `nb.raise_page(name)` immediately after creation.

**Testing:** Trigger an operation that adds a new tab (e.g., "Show reachable states" or "Step in"). Confirm a new tab labeled "Sheet 2" (or incremented) appears. Confirm the new tab is immediately selected/active.

### 3.2 Removing a Sheet Tab
`TkUI.remove(art)` deletes the tab by name from the NoteBook and decrements `tabs`. If `tabs` reaches 0 after removal, `tk.quit()` is called and the application exits.

**Testing:** With a single sheet open, select "Remove tab" from File menu. Confirm the application exits. With two sheets, remove one; confirm the other remains and the application continues running.

### 3.3 Tab Counter Persistence
`tab_counter` is only incremented, never decremented. Thus if Sheet 1 and Sheet 2 exist and Sheet 1 is removed, the next sheet added will be Sheet 3 (not Sheet 2).

**Testing:** Open a file (Sheet 1 created). Trigger addition of Sheet 2. Remove Sheet 2. Trigger addition of another sheet. Confirm it is named "Sheet 3", not "Sheet 2".

### 3.4 Tab Label Display
Each tab in the NoteBook shows its label string (e.g., "Sheet 1"). The label is set via the Tix NoteBook tab creation call. Labels are static after creation.

**Testing:** Open the application with two sheets. Confirm both tabs show their respective labels. Verify labels do not change dynamically as the user interacts with the content.

### 3.5 Raising Tab to Front
When `add()` creates a new tab, `nb.raise_page(name)` brings it to the foreground. When operations create new ARGs (e.g., "Show reachable states"), the new sheet's tab becomes active.

**Testing:** Be viewing Sheet 1 and trigger creation of a new ARG. Confirm the new Sheet 2 tab is automatically activated and visible.

---

## 4. Menu Bar System

### 4.1 Menu Bar Widget Construction
`WithMenuBar.__init__` reads the `menus()` method of the widget class, iterates through the menu structure list, and creates `Menubutton` widgets packed left in the `MenuBar` frame. Each `Menubutton` has `tearoff=0`.

**Testing:** Confirm menu buttons appear in the menu bar. Confirm no tearoff perforation is present (menus cannot be torn off). Confirm clicking a menu button opens a dropdown.

### 4.2 Menu Item: Button Type
Menu items of type `"button"` create a `Menu.add_command(label=..., command=...)` entry. Each button has a label string and a callback. Clicking executes the callback immediately.

**Testing:** Click "File" menu and then "Save". Confirm a save-as dialog appears. Click "Action" → "Recalculate all". Confirm recalculation runs.

### 4.3 Menu Item: Separator
Menu items of type `"separator"` add a visual divider line in the dropdown menu. They have no command or label.

**Testing:** Open the File menu. Confirm a horizontal rule appears between "Save abstraction" and "Remove tab". Confirm clicking on the separator does nothing.

### 4.4 Menu Item: Radiobuttons
Menu items of type `"radiobuttons"` create a set of `Menu.add_radiobutton()` entries sharing a `StringVar`. The variable stores the currently selected option name. The `radiobutton(name)` method on `WithMenuBar` returns the `StringVar` for a given set.

**Testing:** Open "Mode" menu. Confirm radio buttons for Concrete, Abstract, Bounded, Induction, Pdr. Select "Induction"; reopen the menu and confirm "Induction" has a bullet/checkmark. Confirm `self.mode` returns the selected value.

### 4.5 RadioButton Variable Storage
Each radiobutton set has its `StringVar` stored in `self.radios[name]` where `name` is the dict key of the radiobutton spec. This allows code to retrieve the current mode via `self.radiobutton('mode').get()`.

**Testing:** Programmatically set the mode variable and confirm the menu reflects the change. Confirm `AnalysisGraphUI.mode` property reads the correct value after a menu selection.

---

## 5. File Menu

### 5.1 Save Analysis State
"Save" in the File menu opens a "Save analysis state as..." save-as dialog filtered to `.a2g` files. If the user provides a filename and confirms, the ARG is pickled to that file. If the user cancels, nothing happens.

**Testing:** Open a file with some states. Click File → Save. Confirm a file dialog appears with `.a2g` filter. Save to a path. Confirm a `.a2g` file exists and can be unpickled as an ARG object. Cancel and confirm no file is written.

### 5.2 Save Abstraction
"Save abstraction" in the File menu opens a "Save abstraction as..." dialog filtered to `.ivy` files. The concept spaces (domain abstraction) are written to the file as ivy source. This allows reloading the proof state.

**Testing:** After adding concept predicates, click File → "Save abstraction". Provide a `.ivy` filename. Confirm the file is written and contains `conjecture` declarations corresponding to the active concepts. Open the saved file and confirm ivy can parse it.

### 5.3 Remove Tab
"Remove tab" in the File menu calls `self.ui_parent.remove(self.g)` to remove the current sheet. If it is the last sheet, the application exits.

**Testing:** With one sheet, File → "Remove tab". Confirm the app exits. With two sheets, remove one and confirm the other remains selected.

### 5.4 Exit
"Exit" in the File menu calls `self.ui_parent.exit()` which calls `self.tk.quit()`, ending the Tk main loop cleanly.

**Testing:** Click File → Exit. Confirm the window closes. Confirm no Python exception is printed. Confirm the process exits with code 0.

---

## 6. Mode Selection

### 6.1 Mode Radiobutton Set
The Mode menu contains radiobuttons for: `"Concrete"`, `"Abstract"`, `"Bounded"`, `"Induction"`, `"Pdr"`. The underlying `StringVar` stores the selected mode key. The default mode is set when the UI is first created.

**Testing:** Confirm all five modes appear as radiobuttons in the Mode menu. Select each one and confirm it becomes checked. Confirm only one mode is checked at a time.

### 6.2 Mode: Concrete
In Concrete mode, `init_alpha()` returns `None` (no abstraction). The ARG operates on concrete states. Safety checking uses bounded safety rather than inductive safety.

**Testing:** Select Concrete mode. Trigger a state action. Confirm states are concrete (not abstracted). Confirm the safety check uses bounded path analysis.

### 6.3 Mode: Abstract
In Abstract mode, `init_alpha()` returns `None`. Like Concrete but the abstraction function may be applied depending on context. The ARG can still operate abstractly through concept spaces.

**Testing:** Select Abstract mode. Confirm it is distinct from Concrete in behavior when concept spaces are defined. Add a concept and confirm it affects abstraction.

### 6.4 Mode: Bounded
In Bounded mode, `check_safety_node` calls `check_bounded_safety` instead of `check_local_safety`. Bounded model checking (BMC) is the primary verification strategy.

**Testing:** Select Bounded mode. Click "Check safety" on a node. Confirm BMC runs rather than local safety checking. Confirm a bound is used for the path length.

### 6.5 Mode: Induction
In Induction mode, `init_alpha()` returns `ivy_alpha.alpha`. This enables inductive generalization when computing post-states. The CTI workflow is the primary proof strategy.

**Testing:** Select Induction mode. Verify `init_alpha()` returns the induction alpha function. Trigger node extension and confirm inductive generalization is applied to resulting states.

### 6.6 Mode: PDR (Property-Directed Reachability)
In PDR mode, `init_alpha()` returns `ivy_alpha.predicate_alpha`. This uses predicate abstraction. The refine dialog uses PDR-specific messaging.

**Testing:** Select PDR mode. Verify `init_alpha()` returns the predicate alpha function. Trigger a safety check and confirm PDR-specific messages appear in any dialogs.

### 6.7 Mode Affects init_alpha()
The `init_alpha()` method reads `self.mode` and returns different alpha functions accordingly. Default (unknown mode) returns `top_alpha`. This controls how post-states are abstracted.

**Testing:** Change the mode and trigger a node extension. Confirm the post-state abstraction method changes. Test the default case (unknown mode string) and confirm `top_alpha` is used.

---

## 7. Action Menu

### 7.1 Recalculate All
"Recalculate all" in the Action menu iterates through all ARG transitions, recomputes each one, and rebuilds the graph. Duplicate target nodes (same as the edge target) are skipped.

**Testing:** Build an ARG with multiple states. Click Action → "Recalculate all". Confirm all edges are recomputed. Introduce a change to the Ivy model externally and confirm recalculation reflects the new model. Verify no duplicate states appear.

### 7.2 Show Reachable States
"Show reachable states" calls `self.show_reachable_states()` which adds a reachable-state tree to the UI as a new tab. The reachable tree is built lazily from the initial state.

**Testing:** Click Action → "Show reachable states". Confirm a new tab appears. Confirm it shows reachable states of the Ivy program. Confirm the new tab is labeled appropriately.

---

## 8. ARG Graph Canvas

### 8.1 Canvas Deletion and Rebuild
`TkAnalysisGraphWidget.rebuild()` calls `self.delete('all')` on the canvas, clearing all items, then recreates everything. This is called on startup and after any state-modifying operation.

**Testing:** Add a state to the ARG. Confirm the canvas updates. Trigger a delete, then confirm the deleted node disappears immediately. Confirm the canvas does not show stale data.

### 8.2 DOT Layout Computation
`rebuild()` creates `CyElements` from the ARG and calls `dot_layout(cy_elements)` (via `ivy_graphviz.AGraph`) to compute node positions using the Graphviz `dot` algorithm. The y-coordinate is flipped relative to y_origin.

**Testing:** Create an ARG with multiple states. Confirm nodes are laid out in a DAG structure (roots at top, leaves at bottom). Confirm positions are stable across rebuilds for the same graph structure.

### 8.3 Edge Label Notation Transformation
After layout, the string `"-["` in edge labels is replaced with `"{"` and `"]-"` is replaced with `"}"`. This converts the internal notation to a more human-readable brace notation.

**Testing:** Create a state transition whose action uses the `[-...]` notation. Observe the edge label in the graph and confirm it shows `{...}` instead of `[-...]`. Verify the substitution is visual-only and does not affect underlying data.

### 8.4 Node Shape Rendering
`TkCyCanvas.create_shape()` creates Tk canvas items for nodes based on their shape type: `ellipse`/`oval` uses `create_oval`, `octagon` uses `create_octagon`. The `double` parameter creates a second, inner shape for double-border effects.

**Testing:** Confirm ARG nodes render as ovals/ellipses. Confirm concept graph nodes with `__ID` sort render as octagons. Confirm nodes requiring double borders (e.g., `at_least_one` cardinality) display two concentric shapes.

### 8.5 Node Shape: Black Fill for Bottom States
In `get_node_styles()`, if the node is a `bottom_state` (representing the error/bottom state), `fill='black'` is used; otherwise `fill=''` (transparent/default background).

**Testing:** Trigger a safety violation that creates a bottom_state. Confirm the bottom state node renders black. Confirm normal states have white/default fill. Confirm the outline is always black.

### 8.6 Node Shape Outline Color
`update_node_color(node)` updates the outline color of the node's canvas shape(s) based on node status. `get_node_styles()` returns `outline='black'` and `width=2` by default.

**Testing:** Create a node. Confirm its outline is black. Mark the node (which should turn red). Confirm outline change. Unmark and confirm black outline returns.

### 8.7 Canvas Scroll Region Update
After `create_elements()` in `rebuild()`, `self.configure(scrollregion=self.bbox('all'))` sets the scroll region to encompass all canvas items. This ensures scrollbars reflect the actual content size.

**Testing:** Build a large ARG. Confirm the scroll region covers all nodes and edges (none are clipped). Drag the scrollbar to the edges and confirm no nodes are missing.

### 8.8 Bezier Curve Edge Rendering
`TkCyCanvas.create_elements()` renders edges as Bezier curves approximated by multiple line segments. The `approximate_cubic_bezier()` function subdivides the Graphviz B-spline output to a list of points, which are rendered as a polyline.

**Testing:** Confirm edges between non-adjacent nodes render as smooth curves, not straight lines. Confirm curves do not overlap with nodes. For a simple two-node graph, confirm a direct straight line or gentle curve.

### 8.9 Edge Arrow Shape
In `get_edge_styles()`, edges use `arrowshape="14 14 5"` (arrow length 14, arrow width 14, arrowhead width 5). The arrow is drawn at the target end of the edge.

**Testing:** Confirm every edge in the ARG has a visible arrowhead at its target node. Verify the arrowhead size is consistent. Confirm the arrow does not overlap the target node excessively.

### 8.10 Dashed Edges for Cover Relations
In `get_edge_styles()`, edges of type `"cover"` use `dash=(5,5)` (alternating 5-pixel dash/gap). Non-cover edges use a solid line.

**Testing:** Mark a node and cover it with another. Confirm the cover edge appears as a dashed line. Confirm regular transition edges are solid. Confirm join edges are solid.

### 8.11 Edge Canvas Event Binding
Each edge canvas item has `Button-1` bound to `click_edge('left', ...)` and `Button-3` bound to `click_edge('right', ...)` via `tag_bind`. These trigger the context menu for edges.

**Testing:** Left-click on an edge. Confirm the left-click action fires. Right-click on an edge and confirm the context menu appears with edge actions.

### 8.12 Node Canvas Event Binding
Each node canvas shape has `Button-1` bound to `click_node('left', ...)` and `Button-3` bound to `click_node('right', ...)`. These trigger node-specific actions.

**Testing:** Left-click a node. Confirm it triggers the view-state action (concept graph opens). Right-click a node and confirm a context menu with "Execute actions", "Check safety", etc. appears.

### 8.13 Subgraph/Cluster Rectangle Rendering
Subgraph shapes (from `CyElements.add_shape`) are rendered as rectangles on the canvas. They appear behind nodes and edges, grouping nodes that belong to the same cluster/isolate.

**Testing:** Open a file with isolates or subgraph structure. Confirm a rectangle appears around nodes belonging to the same cluster. Confirm the rectangle does not obscure node labels.

---

## 9. Node Context Menu (Right-Click)

### 9.1 Context Menu Pop-Up via make_popup()
`TkCyCanvas.make_popup(event, actions, arg)` creates a Tk `Menu` widget at the event location and calls `post(event.x_root, event.y_root)`. The menu is destroyed after selection.

**Testing:** Right-click any node. Confirm a context menu appears near the cursor. Click outside the menu to dismiss it. Confirm the menu disappears and no action is taken.

### 9.2 Auto-Execute Single Action with '<>' Label
If the action list contains exactly one item and its label is `'<>'`, `make_popup` executes that action immediately without showing a menu. This is an invisible "default" action.

**Testing:** Identify a case where a single auto-execute action is defined. Confirm clicking the target does not show a menu but immediately executes the action. Verify by checking for side effects of the action.

### 9.3 Separator via '---' Label
In `make_popup`, items with label `'---'` add a `Menu.add_separator()`. They have no command.

**Testing:** Right-click a node and confirm visual dividers appear between logical groups of menu items. Confirm clicking a separator does nothing. Confirm separators appear between "Execute actions" group and node command group.

### 9.4 Label-Only Menu Items (None Command)
Items with `command=None` are added as disabled labels (`add_command(state=DISABLED)`). They serve as section headers or informational text.

**Testing:** Right-click a node that has state equations available. Confirm "Execute actions" appears as a non-clickable header label. Confirm it is visually distinct from clickable items.

### 9.5 Node Context Menu: Execute Actions
The node context menu shows all available state equations at the top as clickable items. Each item is labeled with `state_equation_label(a)` (the action name). Clicking executes `do_state_action(a, node)`.

**Testing:** Right-click a node in a state with applicable actions. Confirm action names appear in the menu. Click one and confirm a new child state is added to the ARG. Confirm the graph rebuilds.

### 9.6 Node Context Menu: Check Safety
"Check safety" calls `check_safety_node(node)`. In Bounded mode this runs BMC; in other modes it runs local safety checking.

**Testing:** Right-click a node → "Check safety". Confirm the appropriate safety check runs. For a safe node, confirm no error dialog. For an unsafe node, confirm an error dialog appears.

### 9.7 Node Context Menu: Extend
"Extend" calls `find_extension(node)`. This finds an applicable action and executes it to extend the ARG. If no action is applicable, shows "State {id} is closed."

**Testing:** Right-click a node → "Extend". If the node has applicable actions, confirm a new child state appears. If fully extended, confirm the "closed" ok_dialog appears.

### 9.8 Node Context Menu: Mark
"Mark" calls `mark_node(node)`. The previously marked node is unhighlighted (red fill removed), and the new node gets red fill via `show_mark(True)`.

**Testing:** Mark a node. Confirm it turns red. Mark a different node. Confirm the first node returns to normal and the second turns red. Confirm only one node is red at a time.

### 9.9 Node Context Menu: Cover by Marked
"Cover by marked" calls `cover_node(covered_node)` with the right-clicked node as the argument. The marked node is used as the covering node. On success, a cover edge is added. On failure, an error message is displayed.

**Testing:** Mark node A. Right-click node B → "Cover by marked". If A subsumes B, confirm a dashed cover edge from B to A appears. If not, confirm an error dialog shows. Confirm the graph rebuilds on success.

### 9.10 Node Context Menu: Join with Marked
"Join with marked" calls `join_node(node2)` with the right-clicked node. The marked node (node1) and right-clicked node (node2) are joined. The result replaces both in the ARG.

**Testing:** Mark node A, right-click node B → "Join with marked". Confirm both A and B are replaced by a single joined state. Confirm the joined state has the abstract union of A and B. Confirm edges are rewired.

### 9.11 Node Context Menu: Try Conjecture
"Try conjecture" calls `try_conjecture(node)`. If multiple conjectures are available, a listbox dialog appears for selection. The selected conjecture initiates a BMC-based proof attempt.

**Testing:** Right-click a node → "Try conjecture". If conjectures exist, confirm a listbox dialog lists them. Select one. Confirm the proof attempt begins and a new ARG or dialog shows the result. If no conjectures, confirm a message is shown.

### 9.12 Node Context Menu: Try Remembered Goal
"Try remembered goal" calls `try_remembered_graph(node)`. If multiple remembered goals exist, a listbox dialog appears. The selected goal is used to set up the concept graph for proof.

**Testing:** Save a proof goal first (via "Remember"). Then right-click a node → "Try remembered goal". Confirm a listbox dialog shows saved goals. Select one and confirm the concept graph is set up for the goal.

### 9.13 Node Context Menu: Delete
"Delete" calls `delete_node(node)`. The node and all its dependents are removed from the ARG. The graph is rebuilt.

**Testing:** Delete a leaf node and confirm it disappears. Delete a node with children; confirm children are also removed. Confirm the graph is valid and consistent after deletion. Confirm undo is not available (delete is permanent in the base UI).

### 9.14 Left-Click on Node: View State
In `get_node_actions()`, a left-click returns a `view_state` action. `view_state(n, clauses, reset)` shows the state in an existing concept graph widget or creates a new one in the right pane.

**Testing:** Left-click a node. Confirm the right pane populates with a concept graph showing the node's abstract state. Left-click another node; confirm the concept graph updates.

---

## 10. Edge Context Menu (Right-Click)

### 10.1 Edge Right-Click: Dismiss
"Dismiss" calls `decompose_edge(transition)` and removes the edge from the view. If the edge cannot be decomposed, an `IvyError` is raised and shown in an error dialog.

**Testing:** Right-click an edge → "Dismiss". Confirm the edge disappears from the ARG. If the edge is non-decomposable, confirm an error dialog. Confirm the ARG is valid after dismissal.

### 10.2 Edge Right-Click: Recalculate
"Recalculate" calls `recalculate_edge(transition)`. The edge is recomputed from its source state using the associated action. The graph is rebuilt.

**Testing:** Right-click an edge → "Recalculate". Confirm the edge and its target state are recomputed. Confirm the new target state reflects the current model. Confirm the graph rebuilds.

### 10.3 Edge Right-Click: Step In (Decompose)
"Step in" calls `decompose_edge(transition)`. If the action can be decomposed into sub-actions, a new tab is added showing the decomposed ARG.

**Testing:** Right-click an edge whose action is composite → "Step in". Confirm a new tab appears with the decomposed action ARG. Confirm the decomposition reflects the action's internal structure.

### 10.4 Edge Right-Click: View Source
"View source" calls `view_source_edge(transition)`. This opens the file browser pointing to the source file and line number of the action definition.

**Testing:** Right-click an edge → "View source". Confirm the file browser window appears (or is brought to front). Confirm the source file is loaded and the relevant action line is highlighted in red. Confirm the browser scrolls to the highlighted line.

---

## 11. Node Marking Behavior

### 11.1 Mark Visualization (Red Fill)
`show_mark(on=True)` sets the fill of all shape items tagged with the marked node's tag to `'red'`. `show_mark(on=False)` sets fill to `''` (transparent).

**Testing:** Mark a node and confirm it shows a red fill. Unmark it (by marking another node or removing mark) and confirm it returns to its default fill. Confirm the red fill persists across rebuilds until explicitly cleared.

### 11.2 Mark Persistence Across Rebuilds
`mark_node(n)` stores the marked node in `self.mark`. When `rebuild()` runs, it calls `show_mark(on=True)` at the end to restore the mark highlight. The mark visually persists even after graph updates.

**Testing:** Mark a node. Trigger a recalculate-all operation (which rebuilds the graph). Confirm the marked node is still visually highlighted red after the rebuild.

### 11.3 Single Mark at a Time
Only one node can be marked at a time. When `mark_node(n)` is called, `show_mark(False)` is first called for the previous `self.mark` before updating `self.mark = n` and calling `show_mark(True)`.

**Testing:** Mark node A, then mark node B. Confirm A is no longer red, only B is red. Repeat for a third node.

---

## 12. Safety Checking

### 12.1 Check Safety Dispatch
`check_safety_node(node)` checks `self.mode`: if mode is `"bounded"` it calls `check_bounded_safety(node)`, otherwise it calls `check_local_safety(node)`.

**Testing:** In Bounded mode, trigger "Check safety" and confirm BMC runs. In Induction mode, trigger "Check safety" and confirm local safety checking runs.

### 12.2 Bounded Safety Check
`check_bounded_safety(node)` calls `self.g.check_bounded_safety(node)`. If the check returns a BMC result (counterexample), it sets `node.safe = False`, updates the node's color, and shows "The node is unsafe: View error trace?" as an ok_cancel dialog. If OK is clicked, the error trace is displayed as a new tab.

**Testing:** Create a safety violation scenario. Select Bounded mode, right-click the state → "Check safety". Confirm the "unsafe" dialog appears. Click OK and confirm a new tab with the error trace ARG opens. Click Cancel and confirm no new tab.

### 12.3 Node Color Update After Safety Check
`update_node_color(node)` is called after `check_bounded_safety`. `node_color(node)` returns `"green"` if `node.safe` is true, `"black"` otherwise. The node's canvas outline is updated accordingly.

**Testing:** After a successful safety check, confirm the node's outline turns green. After a failed safety check, confirm it turns black (or remains black). Confirm all other nodes remain unchanged.

### 12.4 Local Safety Check
`check_local_safety(node)` calls `self.g.check_safety(node)`. If unsafe states are found, it shows a `buttons_dialog_cancel` with the message "The node is not proved safe: {reason}", with buttons to view unsafe states or view the error trace.

**Testing:** Create a local safety violation. Trigger local safety check. Confirm a buttons dialog appears with the appropriate message. Click "View unsafe states" and confirm the states are shown. Click "View error trace" and confirm a trace dialog.

### 12.5 Safety Check: "View Error Trace" Button
When the local safety check fails, one of the buttons in the `buttons_dialog_cancel` is labeled to view the error trace. Clicking it calls the trace viewing function and adds a new ARG tab.

**Testing:** Trigger a local safety failure. In the resulting dialog, click the view-trace button. Confirm a new tab with the counterexample trace ARG appears. Confirm the trace states are highlighted appropriately.

---

## 13. Bounded Model Checking (BMC)

### 13.1 BMC Entry Point
`AnalysisGraphUI.bmc(state, err_cond, bound)` calls `self.g.bmc(state, err_cond, bound)`. If the result is `None` (no counterexample), shows an `ok_dialog` stating unreachable. Otherwise adds the result ARG as a new tab.

**Testing:** Run BMC on a safe property with a small bound. Confirm the "unreachable" ok_dialog appears. Run BMC on an unsafe property. Confirm a new tab with the counterexample ARG appears.

### 13.2 BMC Bound Specification
In `bmc_conjecture()`, if no bound is provided, `int_dialog` is called to ask the user for a bound value. The bound constrains the number of steps in the path search.

**Testing:** Trigger a BMC without specifying a bound. Confirm an integer input dialog appears asking for the bound. Enter a valid integer and confirm BMC runs with that bound. Enter a non-integer and confirm an error dialog. Enter an out-of-range value and confirm an error.

### 13.3 BMC Result Display (Unreachable)
When BMC finds no counterexample, `ok_dialog("State {id} is not reachable within {bound} steps")` is displayed. The user can acknowledge and continue.

**Testing:** Run BMC on an obviously safe property. Confirm the unreachable message dialog appears with the correct bound number. Click OK and confirm the dialog closes and the ARG is unchanged.

### 13.4 BMC Result Display (Counterexample)
When BMC finds a counterexample, `view_ag(res)` adds the result ARG as a new tab. The new tab shows the counterexample path as a sequence of states.

**Testing:** Create a property that is violated within 1 step. Run BMC. Confirm a new tab appears immediately. Confirm the tab shows a 2-state ARG (initial + violating state). Confirm each state shows its concrete values.

---

## 14. Find Extension

### 14.1 Closed Node Dialog
If `find_extension(node)` finds no applicable state equations, it shows `ok_dialog("State {node.id} is closed.")`. This informs the user that the node has no outgoing actions.

**Testing:** Fully extend a node until no more actions are available. Right-click → "Extend". Confirm the "closed" ok_dialog appears with the correct node ID.

### 14.2 Extension Execution
If applicable actions are found, `find_extension` calls `do_state_action(a, node)` for one of them. This adds a new child state to the ARG.

**Testing:** Right-click an unexpanded node → "Extend". Confirm a new child state appears connected to the selected node by a transition edge.

---

## 15. Decompose Edge / Step In

### 15.1 Decompose Success: New Tab
`decompose_edge(transition)` calls `self.g.decompose_edge(transition)`. On success, a new `AnalysisSubgraph` is added to the UI via `view_ag(res)`, which adds a new tab.

**Testing:** Right-click an edge whose action is a compound action → "Step in". Confirm a new tab appears containing the decomposed ARG. Confirm the sub-actions are visible as separate transitions.

### 15.2 Decompose Failure: Error Dialog
If `decompose_edge` raises `IvyError` (action cannot be decomposed), the `RunContext` catches it and shows an error dialog with the error message.

**Testing:** Right-click an edge whose action is atomic (cannot be decomposed) → "Step in". Confirm an error dialog appears. Click OK and confirm the dialog closes without any new tab.

---

## 16. Conjecture Workflow

### 16.1 Try Conjecture: No Conjecture Given (List Dialog)
If `try_conjecture(node, conj=None)` is called without a conjecture, it gets the list of unproven conjectures via `itp.undecided_conjectures(state)`, formats them as text, and shows a `listbox_dialog` titled "Choose a conjecture to prove:" with OK and Cancel.

**Testing:** With multiple unproven conjectures in the model, right-click a node → "Try conjecture". Confirm a listbox dialog appears with all unproven conjecture formulas listed. Select one and click OK; confirm the proof attempt starts. Click Cancel; confirm nothing happens.

### 16.2 Try Conjecture: Browse Source
After a conjecture is selected, `self.ui_parent.browse(filename, lineno)` opens the file browser pointing to the conjecture's source declaration.

**Testing:** Select a conjecture. Confirm the file browser window opens (or comes to front) with the source file loaded and the conjecture line highlighted.

### 16.3 Try Conjecture: BMC Mode
In modes other than PDR and induction, `try_conjecture` calls `self.bmc(state, dual_of_conjecture, bound)` to check the conjecture by bounded model checking.

**Testing:** In Bounded mode, try a conjecture. Confirm BMC runs. Confirm either the counterexample ARG or the "unreachable" dialog appears.

### 16.4 Try Conjecture: Show Concept Graph (Induction/PDR)
In Induction or PDR mode, `try_conjecture` calls `self.show_graph(sg)` to display a concept graph with the conjecture's predicate abstraction loaded.

**Testing:** In Induction mode, try a conjecture. Confirm the concept graph widget appears in the right pane (or as a new popup). Confirm the concept graph includes the conjecture's predicates.

### 16.5 Remember Graph
`remember_graph(name, graph)` stores the graph in `self.remembered_graphs[name]`. Names and graphs are stored for later retrieval by "Try remembered goal".

**Testing:** After performing a proof step, call remember_graph. Then right-click a node → "Try remembered goal" and confirm the saved graph name appears in the listbox.

### 16.6 Try Remembered Goal: List Dialog
If multiple goals are remembered, `try_remembered_graph(node, goal=None)` shows a `listbox_dialog` with names of remembered goals. The user selects one and the concept graph is set up.

**Testing:** Remember two different proof goals. Right-click a node → "Try remembered goal". Confirm both names appear in the listbox. Select each and confirm the correct concept graph is loaded.

---

## 17. Interpolant / Refinement

### 17.1 Refine with Interpolant Dialog
`refine_with_interpolant(interp)` shows a dialog with the interpolant formula. In PDR mode the message says "PDR found the following invariant...". In other modes it says "Found the following separating formula...". A "Refine" button (or custom label) is available.

**Testing:** Trigger an interpolant refinement step. Confirm a text dialog appears showing the interpolant formula. In PDR mode confirm PDR-specific message text. Click "Refine" and confirm the concept space is updated.

### 17.2 Pre-State Vacuous Interpolant Message
If the pre-state is vacuous (no counterexample), `refine_with_interpolant` shows the message "The pre-state is vacuous..." instead of the interpolant formula.

**Testing:** Set up a vacuous pre-state and trigger refinement. Confirm the "vacuous" message appears in the dialog rather than a formula.

### 17.3 Add Predicate to Domain
`add_predicate(interp)` adds the interpolant as a predicate to the abstract domain. This expands the concept space without user interaction.

**Testing:** Trigger an automatic predicate addition. Confirm the concept graph widget updates to show the new predicate. Confirm the abstract domain in the current sheet includes the new concept.

---

## 18. Concept Graph Display

### 18.1 Concept Graph Creation via view_state()
`view_state(n, clauses, reset)` creates a concept graph `sg` from `self.g.concept_graph(n, ...)`. If `reset=True` or no existing concept graph exists, `self.show_graph(sg)` is called. Otherwise the existing concept graph is updated via `update`.

**Testing:** Left-click a node. Confirm the concept graph appears. Left-click another node with `reset=False`; confirm the same concept graph widget updates rather than creating a new one. Left-click with `reset=True`; confirm a new concept graph replaces the old one.

### 18.2 Concept Graph Node Classes
Concept graph nodes have CSS classes: `non_existing` (hidden), `exactly_one` (4px solid border), `at_least_one` (8px double border), `at_most_one` (3px dotted border), `node_unknown` (no border). The class reflects the cardinality of the concept.

**Testing:** Open a concept graph for a state with multiple sorts. Confirm nodes are styled according to their cardinality. A singleton sort should show "exactly_one" (thick solid border). An empty sort shows "non_existing" (hidden).

### 18.3 Concept Node Labels
Each concept node has a label showing the node name and optional additional label lines. Labels are formatted with prefix notation: empty prefix for "necessarily true", `~` for "necessarily false", `?` for "unknown".

**Testing:** Open a concept graph. Hover over (or read) node labels. Confirm prefix characters (empty/~/?) appear. Confirm equality constraints show `=` or `≠` appropriately.

### 18.4 Concept Edge Classes
Concept graph edges have classes: `none_to_none` (dashed), `all_to_all` (solid), `edge_unknown` (dotted). Additional classes: `total` (circle source arrow), `functional` (square source arrow), `injective` (triangle-backcurve target), `surjective` (filled target arrow).

**Testing:** Open a concept graph with a functional relation. Confirm a square source arrow appears. For a total relation, confirm a circle source arrow. Verify dashed/solid/dotted line styles correspond to edge class.

### 18.5 Concept Graph Background Color
The concept graph background is `rgb(192,192,255)` (light blue/periwinkle). This distinguishes it from the white ARG canvas.

**Testing:** Open a concept graph. Confirm the background is a light blue color. Confirm the ARG canvas remains white. Confirm the blue background persists across updates.

### 18.6 Concept Node Shape
`get_shape(concept_name)` returns `'ellipse'` for most nodes and `'octagon'` for `'__ID'` concepts. The shape is set in the `CyElements.add_node` call.

**Testing:** Open a concept graph for a state that includes an `__ID` relation. Confirm nodes representing IDs render as octagons. All other nodes should be ellipses.

### 18.7 Transitive Reduction of Edges
`get_transitive_reduction(widget, a, edges)` hides edges that are implied by transitivity. For transitive relations where the checkbox is enabled, edges covered by a longer transitive path are suppressed. Self-edges on transitive relations are hidden first.

**Testing:** Add a transitive relation with multiple nodes. Confirm that only "direct" edges (not implied by transitivity) are shown. If A→B, B→C, A→C exists for a transitive relation, A→C should be hidden.

---

## 19. Concept Graph Controls (Right Pane)

### 19.1 Relation Button Panel
`create_relbuttons_window()` creates a separate Tix `HList` frame with relation display controls. Each relation has 5 controls: `[+]`, `[?]`, `[-]`, relation-label, `[T]`. These are `Checkbutton` widgets bound to `IntVar`s.

**Testing:** Open the relation buttons panel. Confirm every relation appears as a row with the 5 controls. Toggle the `[+]` button for a relation; confirm the "all_to_all" edges appear/disappear. Toggle `[T]` for a transitive relation and confirm transitive reduction behavior changes.

### 19.2 Relation Button: + (All-to-All / Positive)
The `+` button (first column) controls display of positive (`all_to_all`) edges for a relation. When checked, edges where all pairs hold are shown.

**Testing:** Uncheck `+` for a relation that has all-to-all edges. Confirm those edges disappear from the concept graph. Re-check and confirm they reappear.

### 19.3 Relation Button: ? (Unknown)
The `?` button controls display of `edge_unknown` edges (where the edge value is undetermined in the current abstraction).

**Testing:** Create a scenario with an unknown relation edge. Uncheck `?` for that relation and confirm the dotted unknown edges disappear. Re-check and confirm they return.

### 19.4 Relation Button: - (None-to-None)
The `-` button controls display of `none_to_none` edges (where no pairs hold). When checked, dashed "no edges" lines are shown.

**Testing:** Check `-` for a relation with known-absent edges. Confirm dashed lines appear representing the "none-to-none" class. Uncheck and confirm they disappear.

### 19.5 Relation Button: T (Transitive)
The `T` button marks a relation as transitive for transitive reduction purposes. When enabled, the transitive reduction algorithm is applied.

**Testing:** Enable `T` for a transitive relation with redundant edges. Confirm transitive-closure edges are hidden. Disable `T` and confirm all edges including implied ones reappear.

### 19.6 Relation Display Checkboxes (IPython Widget Version)
In the IPython/Jupyter version, `ConceptSessionControls` creates checkbox widgets for each relation/concept with `new_display_checkbox()`. Checking/unchecking updates the domain concepts and triggers `recompute()` and `render()`.

**Testing:** Toggle a relation checkbox in the concept session widget. Confirm the concept graph updates immediately. Confirm the domain's active concepts change accordingly. Confirm the checkbox state is persistent after re-rendering.

### 19.7 Edge Class Buttons (+, ?, -, ≤) in IPython Widget
`update_view_controls()` creates small header buttons labeled `+`, `?`, `-`, `≤` at the top of each edge group. Clicking a class button toggles all checkboxes in that class simultaneously.

**Testing:** Click the `+` class button. Confirm all positive-edge checkboxes toggle off (if all were on) or all toggle on (if any were off). Verify the logic: all-on → all-off, else all-on.

### 19.8 Color Assignment per Concept/Relation
`TkGraphWidget.choose_colors()` assigns distinct colors to relations and sorts, cycling through 27 predefined color names. Each relation gets a unique color for edge lines.

**Testing:** Open a concept graph with multiple relations. Confirm each relation uses a distinct line color. With more than 27 relations, confirm colors cycle without error.

### 19.9 Constraint Text in Concept Graph
`TkGraphWidget.rebuild()` renders constraint formulas as text items on the canvas. Each constraint text has a `left_click_constraint` callback. Selected constraints have black text; unselected have grey text.

**Testing:** Open a concept graph with constraints. Confirm formula text appears on the canvas. Click a formula text and confirm it toggles selected (black) / unselected (grey). Confirm the selection state is tracked.

---

## 20. CTI (Counterexample to Induction) Widget

### 20.1 CTI Menu: Invariant Menu
The CTI `AnalysisGraphUI` adds an "Invariant" menu with: "Check induction", "Bounded check", "Diagram", "Weaken". These appear in addition to the base File menu.

**Testing:** Open the CTI UI. Confirm the "Invariant" menu exists alongside "File". Confirm all four items are present. Click each and confirm the corresponding function is invoked.

### 20.2 CTI Start: Auto-detect Transitive
`start()` calls `autodetect_transitive()` which scans the module for transitive relations and populates `transitive_relations` and `transitive_relation_concepts`. These are displayed with `'T'` in the relation buttons.

**Testing:** Open an ivy file with defined transitive relations. Confirm the concept graph shows 'T' labels for those relations automatically on startup. Confirm relations not marked transitive in the model do not show 'T'.

### 20.3 CTI Check Inductiveness
`check_inductiveness(button)` builds a proof ARG by checking each conjecture for inductiveness. For each failed conjecture, the CTI (counterexample to induction) states are shown. Shows an error dialog if any conjecture has a counterexample. Shows a success dialog with the full invariant text if all pass.

**Testing:** With an inductive invariant loaded, click Invariant → "Check induction". Confirm a success dialog appears with the invariant text. With a non-inductive conjecture, confirm an error dialog with the CTI description. Check the concept graph updates to show the CTI states.

### 20.4 CTI Success Dialog: Invariant Text
When all conjectures are proved inductive, an `ok_dialog` shows the text of the invariant (all conjectures joined with newlines). The user can read and acknowledge.

**Testing:** Prove an inductive invariant. Confirm the success dialog contains the full invariant text as a readable formula. Confirm the text is correct by comparing to the `.ivy` source.

### 20.5 CTI Failure Dialog: Counterexample Description  
When a CTI is found, `ok_dialog` shows "An assertion failed..." with a description of the failed conjecture and the CTI. The `have_cti` flag is set to `True`.

**Testing:** Use a non-inductive conjecture. Click "Check induction". Confirm the dialog message contains the assertion failure. Confirm `have_cti` is set. Confirm the concept graph shows the pre/post CTI states.

### 20.6 CTI: Set Pre/Post States
`set_states(s0, s1)` on the `ConceptGraphUI` sets the pre and post states. The concept graph shows both states side-by-side or in a transition view.

**Testing:** After CTI check fails, confirm the concept graph shows the pre-state on the left and post-state on the right (in the Tk version, as two graphs; in the Jupyter version, as separate CyGraphWidgets).

### 20.7 CTI Weaken Invariant
`weaken(conjs, button)` shows a `listbox_dialog` with multiple-selection enabled, listing all current conjectures. Selected conjectures are removed from the invariant. A confirmation `ok_dialog` shows the remaining invariant text.

**Testing:** Click Invariant → "Weaken". Confirm a multi-selection listbox dialog appears. Select some conjectures and click OK. Confirm those conjectures are removed. Confirm the confirmation dialog shows the remaining invariant text.

### 20.8 CTI Save Invariant
`save_conjectures()` (from "Save invariant" in the File menu) writes the current conjecture set to a `.ivy` file. Old conjectures, new conjectures, and dropped conjectures are tracked and annotated with comments.

**Testing:** Make some changes to the conjecture set. Click File → "Save invariant". Provide a filename. Confirm the file is written. Check the file contains the correct conjecture declarations and appropriate comments marking old/new/dropped.

### 20.9 CTI Bounded Check on Conjecture
`bmc_conjecture(button, bound)` runs BMC on the currently selected conjecture. If no bound is given, `int_dialog` asks the user. Results are shown in a `text_dialog` displaying the BMC trace or unreachable message.

**Testing:** Select a conjecture in the CTI UI. Click Invariant → "Bounded check". Enter a bound. If a counterexample is found, confirm a text dialog appears with the trace. If none found, confirm a "safe within N steps" message.

### 20.10 CTI: Show Used Relations
`show_used_relations(clauses, both)` updates the concept graph to highlight relations used in the CTI clauses. It calls `clear_edges()` then enables checkboxes for used relations.

**Testing:** After a CTI, confirm the relation buttons show only the relations relevant to the failed conjecture highlighted. Unrelated relations should be unchecked.

### 20.11 CTI Diagram View
`diagram()` creates a diagram visualization from the current abstract state. It calls `autodetect_inductiveness()` if needed and shows the result in the concept graph widget.

**Testing:** Click Invariant → "Diagram". Confirm the concept graph updates to show the abstract state diagram. Confirm nodes and edges reflect the abstract value of the current state.

### 20.12 Conjecture Menu: Undo
The `ConceptGraphUI` Conjecture menu has "Undo" which calls `undo()` on the concept session. This reverts the last domain modification.

**Testing:** Make a domain change (e.g., add a relation). Click Conjecture → "Undo". Confirm the domain reverts. Verify the concept graph updates to reflect the reverted domain.

### 20.13 Conjecture Menu: Redo
"Redo" applies a previously undone operation. This re-applies the last undone domain change.

**Testing:** Undo a domain change, then click Conjecture → "Redo". Confirm the change is reapplied. Confirm undo/redo history is maintained correctly across multiple operations.

### 20.14 Conjecture Menu: Gather
"Gather" calls `gather_facts(button)`. It collects all visible facts from selected nodes and edges and presents them as candidates for conjecture strengthening.

**Testing:** Select some nodes and edges in the concept graph. Click Conjecture → "Gather". Confirm the gathered facts appear in the facts list. Confirm gathered facts reflect visible properties of the selection.

### 20.15 Conjecture Menu: Minimize
"Minimize" calls `minimize_conjecture(button)`. It reduces the currently selected facts to the minimal unsat core. A `text_dialog` shows the resulting minimal conjecture.

**Testing:** Select more facts than necessary. Click Conjecture → "Minimize". Confirm a text dialog shows a smaller set of facts that still achieves the proof goal. Confirm the minimal conjecture is a subset of the selected facts.

### 20.16 Conjecture Menu: Check Sufficient
"Check sufficient" calls `is_sufficient(button)`. It checks if the active conjecture implies the target. A text_dialog shows the result. If insufficient, offers to view the counterexample.

**Testing:** Select an insufficient set of facts. Click Conjecture → "Check sufficient". Confirm a text dialog says the conjecture is not sufficient. Confirm an option to view the counterexample. Select sufficient facts and confirm a success message.

### 20.17 Conjecture Menu: Check Relative Induction
"Check relative induction" calls `is_inductive(button)`. It checks if the conjecture is inductive relative to the current invariant. A text_dialog shows the result.

**Testing:** Select a relatively inductive conjecture. Confirm a success message. Select a non-inductive one and confirm a counterexample message with an option to view it.

### 20.18 Conjecture Menu: Strengthen
"Strengthen" calls `strengthen(button)`. It gets the selected conjecture and shows a `text_dialog` with the conjecture text, labeled "Add conjecture". If confirmed, the conjecture is added to the invariant.

**Testing:** Select some facts forming a valid conjecture. Click Conjecture → "Strengthen". Confirm a text dialog shows the conjecture formula. Click "Add conjecture" and confirm the conjecture is added to the invariant list.

### 20.19 Conjecture Menu: Export
"Export" exports the current concept graph as a DOT file via a save dialog.

**Testing:** Click Conjecture → "Export". Confirm a save-as dialog appears. Save to a `.dot` file. Confirm the file is written and contains valid DOT graph syntax. Open with Graphviz to verify the graph renders correctly.

### 20.20 View Menu: Add Relation
"Add relation" in the View menu shows a dialog allowing the user to type a relation name to add to the concept graph display.

**Testing:** Click View → "Add relation". Confirm a dialog appears (entry_dialog) asking for a relation name. Enter a valid relation name and confirm it appears in the concept graph. Enter an invalid name and confirm an error dialog.

### 20.21 Concept Node Right-Click: Projections Menu
`ConceptGraphUI.get_node_actions()` for right-click returns a "Projections" cascade menu. Each projection of the node is listed as a menu item that calls `add_projection()`.

**Testing:** Right-click a concept graph node that has ternary relations. Confirm a "Projections" menu appears. Click a projection item and confirm the projection is added to the concept graph display.

### 20.22 Concept Node Left-Click: Select
Left-click on a concept node calls `select(node)`. This selects the node for fact gathering or other operations.

**Testing:** Left-click a concept node. Confirm it becomes visually selected (grey fill in Tk version). Left-click another node and confirm the first is deselected. Confirm selected nodes contribute to "Gather" operations.

### 20.23 Node Context Action: Remove Concept
Right-clicking a concept node offers a "remove" action. This calls `remove_concepts(concept)` on the session, deleting the concept from the domain.

**Testing:** Right-click a concept node → "remove". Confirm the node disappears from the concept graph. Confirm the domain no longer includes that concept. Confirm undo restores it.

### 20.24 Node Context Action: Suppose Empty
"suppose_empty" in node context menu calls `suppose_empty(concept)`. This adds an emptiness constraint assuming the concept has no instances.

**Testing:** Right-click a node → "suppose empty". Confirm the node's cardinality class changes to indicate emptiness (e.g., `non_existing`). Confirm the suppose constraint is added to the session.

### 20.25 Node Context Action: Materialize
"materialize" calls `materialize_node(concept_name)`. This creates a concrete witness element for the concept.

**Testing:** Right-click an abstract node → "materialize". Confirm a new concrete witness node appears in the concept graph. Confirm the concept graph updates to show the materialized node.

### 20.26 Node Context Action: Split
For a node with label-based splitting, actions like "split by {label}" appear. These call `split(concept, by)` to refine the concept space.

**Testing:** Right-click a node that has applicable label concepts. Confirm "split by {label}" items appear. Click one and confirm the node is split into sub-nodes based on the label value.

### 20.27 Edge Context Action: Materialize+/−
Edge context menus offer "materialize +" (add a positive edge fact) and "materialize -" (add a negative edge fact). These call `materialize_edge()`.

**Testing:** Right-click an unknown edge → "materialize +". Confirm the edge class changes to all_to_all. Right-click → "materialize -". Confirm it changes to none_to_none. Verify the concept session reflects these constraints.

### 20.28 Get Selected Conjecture
`get_selected_conjecture()` builds a positive universal conjecture from currently selected facts in the concept graph. Numerals in the facts are substituted with universally quantified variables.

**Testing:** Select facts in the concept graph. Call get_selected_conjecture. Confirm the result is a universally quantified formula. Confirm numerals in the facts are replaced by variables. Confirm unselected facts are excluded.

---

## 21. Session History Navigation (Jupyter Widget)

### 21.1 History Navigation Buttons (First/Prev/Next/Last)
`AnalysisSessionWidget` has four navigation buttons: "first", "prev", "next", "last". These update `current_step` and re-render the display.

**Testing:** Navigate through a multi-step proof. Click "next" and confirm the display advances one step. Click "prev" and confirm it goes back. Click "first" and confirm it goes to step 0. Click "last" and confirm it goes to the latest step.

### 21.2 Step Info Display
The `step_box` Textarea displays information about the current step from `self.step_info`. This shows what operation was performed at each step in the proof history.

**Testing:** Navigate through proof steps. Confirm the step_box text changes at each step. Confirm step_info accurately reflects the operation (e.g., "recalculate", "new goal", etc.).

### 21.3 Auto-Click Active Element
`step()` calls `auto_click` on the active element at the current step. This automatically triggers the appropriate click handler (arg_node_click, crg_node_click, etc.) without user action.

**Testing:** Navigate to a step that has an active element. Confirm the concept widget automatically updates as if the element were clicked. Confirm this happens without user interaction beyond navigation.

### 21.4 Modal Message on Step
If `silent` is not set, `step()` sends a `new_message(title, body)` to the `ModalMessagesWidget`, displaying a modal popup with the step's message.

**Testing:** Navigate to a step with an associated message. Confirm a modal dialog appears with the message title and body. Dismiss it and continue navigation.

---

## 22. Dialog System

### 22.1 OK Dialog
`ok_dialog(msg)` creates a `Toplevel` window with a `Label` showing `msg` and an "OK" button. The window is centered on the parent. `tk.wait_window(dlg)` blocks until closed.

**Testing:** Trigger any action that shows an OK dialog (e.g., close a closed node). Confirm a centered popup with the message and single OK button appears. Click OK and confirm the popup closes and the application continues normally.

### 22.2 OK Dialog: Test Automation Answer
If `self.answers` is non-empty, `getans()` pops the last answer and auto-clicks the appropriate button without user interaction. This supports automated testing.

**Testing:** Call `ui.answer("ok")` before triggering an ok_dialog. Confirm the dialog auto-closes without user interaction. Verify this works for all dialog types.

### 22.3 OK/Cancel Dialog
`ok_cancel_dialog(msg, cmd)` creates a `Toplevel` with a `Label`, "OK" button (calls `cmd` then closes), and "Cancel" button (just closes). OK executes the command; Cancel does nothing.

**Testing:** Trigger a safety check failure that shows "View error trace?". Click OK; confirm the error trace opens. Trigger again; click Cancel; confirm no trace is shown. Test with `ui.answer("ok")` and `ui.answer("cancel")` for automation.

### 22.4 Listbox Dialog (Single Selection)
`listbox_dialog(msg, items, command, on_cancel, multiple=False)` creates a `Toplevel` with a `Label`, `Scrollbar`, `Listbox` (height=8, width=50, selectmode=SINGLE), "OK" button, and optional "Cancel" button. OK calls `command(selection_index)`.

**Testing:** Trigger a conjecture selection dialog. Confirm the listbox shows all items. Select one and click OK; confirm `command` is called with the correct index. Click Cancel; confirm `on_cancel` is called. Test with more items than fit to confirm scrollbar works.

### 22.5 Listbox Dialog (Multiple Selection)
When `multiple=True`, the listbox uses `selectmode=EXTENDED`. Clicking OK calls `command(list_of_indices)`.

**Testing:** Trigger a "Weaken" action that shows a multi-select dialog. Select multiple items using Ctrl+click or Shift+click. Click OK and confirm all selected indices are passed to the callback. Confirm Extended selection mode works (Ctrl+click for individual, Shift+click for range).

### 22.6 Listbox Dialog Centering
`center_window_on_window(toplevel, win)` centers the dialog over its parent window. The position is computed from the parent's geometry and the dialog's requested size.

**Testing:** Open a dialog from different positions of the main window. Confirm the dialog consistently appears centered over the parent window, not at a fixed screen position.

### 22.7 Text Dialog
`text_dialog(tk, root, msg, text, command, on_cancel, command_label)` creates a `Toplevel` with a `Label`, `Scrollbar`, `Text` widget (height=4, width=100), an "OK" (or `command_label`) button, and optional "Cancel" button. Initial text is inserted into the `Text` widget.

**Testing:** Trigger an action showing a text dialog (e.g., "Strengthen"). Confirm the text widget pre-populated with the conjecture text. Edit the text and click OK; confirm modified text is passed to the command. Confirm the text widget is scrollable.

### 22.8 Entry Dialog
`entry_dialog(tk, root, msg, command, on_cancel, command_label, initval)` creates a `Toplevel` with `Label`, `Entry` widget (with optional initial value), "OK" (or `command_label`) button, and optional "Cancel" button. The Entry widget gains focus. Pressing `<Return>` triggers OK.

**Testing:** Trigger an entry dialog (e.g., "Add relation"). Confirm the entry widget has focus immediately. Type text and press Enter; confirm OK is triggered without clicking the button. Delete and re-type, then click OK; confirm the command receives the text.

### 22.9 Integer Dialog
`int_dialog(tk, root, msg, minval, maxval, command)` wraps `entry_dialog` with a validator that converts the input to int and checks it is within `[minval, maxval]`. Non-integers show an `IvyError` dialog. Out-of-range values show an error dialog.

**Testing:** Trigger a BMC bound dialog. Enter "abc" and confirm "not an integer" error. Enter "-1" for a positive-only bound and confirm "out of range" error. Enter a valid integer and confirm the command fires with the correct value.

### 22.10 Buttons Dialog (Custom Buttons)
`buttons_dialog_cancel(tk, root, msg, button_commands, on_cancel)` creates a dialog with a `Label` and one button per `(label, callback)` pair, plus a "Cancel" button. Each custom button calls its callback.

**Testing:** Trigger a local safety failure that shows a buttons dialog. Confirm each button labeled action appears. Click one and confirm its callback fires. Click Cancel and confirm `on_cancel` is called. Confirm the dialog closes after any button click.

### 22.11 Save As Dialog
`saveas_dialog(msg, filetypes)` calls `tkinter.filedialog.asksaveasfile(title=msg, filetypes=filetypes)`. Returns the opened file object if confirmed, `None` if cancelled.

**Testing:** Click File → Save. Confirm the native OS save-as dialog appears with the correct title. Confirm file type filtering (e.g., `.a2g` files). Save a file and confirm the path is returned. Cancel and confirm `None` is returned.

### 22.12 Dialog Window Centering
`center_window(toplevel)` centers a dialog on screen by parsing its geometry, computing the center offset, and calling `geometry()`. This is used for dialogs not centered on a parent.

**Testing:** Open a standalone dialog (one that uses `center_window` rather than `center_window_on_window`). Confirm it appears in the center of the screen. Move the main window to a corner and confirm the dialog still centers on screen.

### 22.13 Error Dialog from RunContext
`RunContext.__exit__` catches `IvyError` exceptions and shows a `Toplevel` with the error message and an "OK" button. The error is swallowed (returns True to suppress). Other exceptions propagate normally.

**Testing:** Trigger an operation that raises `IvyError` (e.g., invalid covering). Confirm an error dialog appears with the error message. Click OK; confirm the application continues normally. Trigger a non-IvyError exception and confirm it propagates.

### 22.14 Busy Cursor During Long Operations
`RunContext.__enter__` calls `parent.busy()` which sets `cursor='watch'` on both `tk` and `frame` widgets and calls `tk.update()`. The watch cursor is shown during the entire duration of a context-managed operation.

**Testing:** Trigger a slow operation (large BMC). Confirm the cursor changes to a watch/hourglass during processing. Confirm it returns to the normal arrow cursor when done. Confirm `tk.update()` keeps the UI responsive enough to show the cursor change.

### 22.15 Ready Cursor After Operation
`RunContext.__exit__` calls `parent.ready()` which sets `cursor=''` (default) on both widgets. This restores the normal cursor.

**Testing:** After a long operation completes, confirm the cursor returns to the default arrow. If an exception is raised in the operation, confirm the cursor still returns to normal (the `ready()` call is in `__exit__`).

---

## 23. File Browser Widget

### 23.1 File Browser Window Creation
`new_file_browser(tk)` creates a `Toplevel` window and fills it with a `FileBrowser` frame. The `FileBrowser` contains a `Text` widget (width=100, height=20) and a vertical `Scrollbar`.

**Testing:** Trigger any "View source" action. Confirm a new file browser window appears. Confirm it contains a scrollable text area. Confirm the window can be independently resized, moved, and closed.

### 23.2 File Loading into Browser
`FileBrowser.set(filename, lineno)` opens and reads the file only if `filename` differs from the currently loaded file. The entire file contents are inserted into the Text widget.

**Testing:** Open the file browser for file A. Confirm A's contents are displayed. Trigger "View source" for a different file B. Confirm B's contents replace A's. Trigger "View source" for file A again. Confirm A is reloaded (not cached, since filename differs).

### 23.3 Line Highlight (Red Background)
`set()` configures a `'highlight'` tag with `background='red'` and adds it to the line at `lineno`. The previous highlight tag is removed before adding the new one.

**Testing:** Open the file browser pointing to line 10. Confirm line 10 has a red background. Trigger "View source" for line 25 of the same or a different file. Confirm the red highlight moves to line 25 and line 10 is no longer highlighted.

### 23.4 Scroll to Highlighted Line
After highlighting, `set()` scrolls the Text widget to the highlighted line using `see(lineno)`.

**Testing:** Open a file with 1000 lines and point to line 900. Confirm the text widget scrolls so line 900 is visible. Confirm the highlighted line is visible without manual scrolling.

### 23.5 File Browser Window Raise
`set()` calls `self.lift()` on the FileBrowser's parent window to bring it to the foreground.

**Testing:** Open the file browser, then bring the main application window to the front. Trigger "View source". Confirm the file browser window rises to the top of the window stack.

---

## 24. Trace Display

### 24.1 Trace Serialization to Lines
`Trace.to_lines()` converts the trace to a multi-line string with indented nested structure. State labels, action names, and parameter values are formatted with indentation for sub-calls.

**Testing:** Generate a counterexample trace. Confirm the trace is displayed as indented text. Confirm sub-actions are indented relative to their caller. Confirm parameter values appear inline.

### 24.2 Infinite Loop Detection in Trace
If the trace contains a repeating cycle, "--- the following repeats infinitely ---" is inserted before the cycle. This prevents the trace display from looping.

**Testing:** Create a trace with an intentional cycle (infinite loop). Confirm the "repeats infinitely" marker appears exactly once in the trace. Confirm the trace display terminates rather than looping.

### 24.3 Function Call Formatting
`do_return()` in Trace formats function call returns with parameter values shown inline. The trace distinguishes call entry from return.

**Testing:** Generate a trace involving a function call. Confirm the trace shows both the call (with arguments) and the return (with return value). Confirm proper indentation nesting.

### 24.4 Trace Line Numbers
`label_from_action(action)` uses `action.lineno` to show source line numbers alongside action names in the trace.

**Testing:** Generate a trace and confirm action lines include the source file line number. Click "View source" (if available) and confirm it opens at the correct line.

### 24.5 Trace: Detailed Option
`option_detailed` is a `BooleanParameter("detailed")` that controls the verbosity of trace output. When True, additional details are included in `to_lines()`.

**Testing:** Set `--detailed=true` before generating a trace. Confirm more details appear in the trace text (e.g., full formula for each state). Compare with `--detailed=false` and confirm reduced output.

---

## 25. Event Viewer (ivy_show.py)

### 25.1 EventTree Widget
`EventTree` inherits from `tkinter.tix.Tree` and `WithMenuBar`. It displays hierarchical events with lazy loading via `opencmd` and browsing via `browsecmd`.

**Testing:** Open the event viewer with a trace file. Confirm a tree widget appears. Click on a tree item to expand it. Confirm child items load lazily when the parent is expanded.

### 25.2 Event NoteBook
`EventNoteBook` manages multiple sheets of event trees. Each `new_sheet(evs)` call creates a new tab in the NoteBook.

**Testing:** Load a trace with multiple event types. Confirm separate sheets appear for different event categories. Confirm each sheet has its own independent tree.

### 25.3 Event Filter Dialog
The Events menu has a "Filter..." item. Clicking it shows an `ask_pat` entry dialog. The entered pattern is used to filter events shown in the tree.

**Testing:** Open the event viewer with a complex trace. Click Events → "Filter...". Enter a regex pattern. Confirm only matching events appear in the tree. Enter an empty pattern and confirm all events return.

### 25.4 Find Reverse Dialog
"Find reverse..." opens an entry dialog for pattern input and searches for events matching the pattern in reverse chronological order.

**Testing:** Click Events → "Find reverse...". Enter a pattern. Confirm matching events are highlighted or listed in reverse order. Confirm the search terminates even on large event logs.

### 25.5 Pattern List
`PatternList` is a frame with a `TList` widget displaying saved patterns. Buttons `+` and `-` add/remove patterns. "<<" reverses a pattern, ">>" forwards it.

**Testing:** Click `+` to add a pattern. Confirm it appears in the TList. Select a pattern and click `-`; confirm it is removed. Click `<<` on a pattern to confirm it is reversed (e.g., A→B becomes B→A).

### 25.6 Pattern List Save/Load/Clear
The PatternList has "Save", "Load", and "Clear" buttons for persisting pattern collections.

**Testing:** Add patterns to the list. Click "Save" and provide a filename. Click "Clear" and confirm the list is empty. Click "Load" with the saved file and confirm patterns are restored. Verify saved patterns survive application restart.

### 25.7 HList Selection Style
The `HList` in EventTree uses `red` foreground for selected items.

**Testing:** Select an event in the EventTree. Confirm the selected item shows red text. Deselect (click elsewhere) and confirm text returns to default color.

---

## 26. Graph Rendering Details

### 26.1 Graph Layout Algorithm (DOT)
`dot_layout(cy_elements)` uses the Graphviz `dot` algorithm for DAG layout. Nodes are sorted topologically with transitive relations taken into account. The `dot` algorithm assigns positions suitable for directed graphs.

**Testing:** Create an ARG with multiple nodes and edges. Confirm the layout is a proper DAG (root nodes at top, leaves at bottom). Confirm no node overlaps. Confirm the layout is consistent (stable for the same graph structure).

### 26.2 Y-Coordinate Inversion
`dot_layout` flips the y-coordinate: `y = y_origin - node.y`. Graphviz positions from bottom-left; this converts to top-left origin.

**Testing:** Confirm that ARG root (initial state) appears at the top and child states appear below it. This would be reversed if y-inversion were not applied.

### 26.3 Back Edge Reversal
If `node_gt(source, target)` returns True (target is "greater than" source in some ordering), the edge is reversed in the Graphviz graph with `dir='back'`, and the spline points are reversed on output.

**Testing:** Create a covering relation or back edge. Confirm the arrow direction is correct in the displayed graph (pointing from covered to covering, not reversed). Confirm the edge rendering looks correct visually.

### 26.4 Cluster/Subgraph Boxes
`dot_layout` with `subgraph_boxes=True` creates Graphviz subgraphs for clusters, which appear as boxes around grouped nodes. The box coordinates are extracted from the Graphviz output.

**Testing:** Open a file with declared isolates or subgraph annotations. Confirm rectangular boxes appear around related node groups. Confirm boxes do not overlap nodes' labels.

### 26.5 Edge Weight Configuration
The `weight` dict in `dot_layout` assigns weights to edge types: `reach`/`le` edges get weight 10 (preferring straight lines), `id` edges get weight 1.

**Testing:** Create a graph with both reach/le edges and id edges. Confirm reach/le edges are rendered more vertically (straighter) due to higher weight. Compare with a version where all weights are equal.

### 26.6 Pending Edge Constraint
The `constraint` dict marks `pending` edges as `constraint=False` in Graphviz. This allows pending edges to cross without affecting the layout hierarchy.

**Testing:** Add a pending edge. Confirm it is laid out without distorting the DAG structure of non-pending edges. Confirm the pending edge may cross other edges freely.

### 26.7 Node Width/Height from Graphviz
After layout, each node's `width` and `height` are extracted from Graphviz output and converted to pixels (`72 * graphviz_units`). These are stored in the element's position data.

**Testing:** Inspect node dimensions after layout. Confirm they match the expected Graphviz output dimensions. Confirm nodes with longer labels get wider bounding boxes. Confirm the canvas renders nodes at these dimensions.

---

## 27. Show Counterexample in Diagnose Mode

### 27.1 --diagnose Flag Behavior
When `--diagnose=true` is set as a command-line parameter, failing property checks open the GUI rather than just printing an error. `check_properties()` shows the UI in Induction mode when properties fail.

**Testing:** Run `ivy --diagnose=true myfile.ivy` with a file that has failing properties. Confirm the GUI opens automatically. Confirm it is in Induction mode. Confirm the failing property is shown in the UI.

### 27.2 show_counterexample() Creates Trace ARG
`show_counterexample(ag, state, bmc_res)` builds a copy of the ARG path and assigns concrete state values from the BMC result. It then calls `gui_art(other_art)` to display it.

**Testing:** Trigger a BMC counterexample in diagnose mode. Confirm the counterexample path is shown as a sequence of concrete states in the ARG. Confirm state values are populated from the BMC model.

### 27.3 check_conjectures() Shows CTI in Diagnose Mode
`check_conjectures(kind, msg, ag, state)` in diagnose mode opens the GUI with the ARG and calls `try_conjecture()` on the failed conjecture, showing it as a listbox for user selection.

**Testing:** Run with a failing conjecture and `--diagnose=true`. Confirm the CTI UI opens with the ARG displayed. Confirm a dialog lists the failed conjectures for investigation.

### 27.4 try_property() in UI
`IvyUI.try_property(prop)` calls `self.try_property()` and `bmc()` on the dual of the property. If no property is given, a listbox dialog lists all properties to try.

**Testing:** With multiple properties defined, call `try_property()` with no argument. Confirm a listbox dialog appears with all property names. Select one and confirm BMC runs on the negation of that property.

---

## 28. Jupyter Notebook Integration

### 28.1 ivy2.py: Notebook Generation
`ivy2.py:main()` reads a `.ivy` filename, generates a Jupyter notebook (`.ipynb`) with Ivy analysis cells, and opens it. Generated cells include: imports, module setup, session creation, and widget display.

**Testing:** Run `ivy2 myfile.ivy`. Confirm a `.ipynb` file is created (or opened if it exists). Confirm the notebook contains cells that set up the Ivy session. Confirm running the cells creates the analysis widgets.

### 28.2 PYTHONPATH Setting for Ivy
`ivy2.py` sets `PYTHONPATH` in the notebook environment to include the Ivy library location. This ensures all Ivy modules are importable within the notebook.

**Testing:** Open the generated notebook and run the import cell. Confirm no ImportError occurs. Confirm the Ivy modules are accessible.

### 28.3 CyGraphWidget JS Integration
`CyGraphWidget` communicates with its JavaScript counterpart via `send()` and `on_msg()`. It uses `_trait_to_json` and `_trait_from_json` to serialize Python objects to/from JSON while preserving object identity.

**Testing:** Create a CyGraphWidget, set elements, and confirm the JS view renders the correct graph. Modify elements and confirm the JS view updates. Verify object identity is preserved across serialization/deserialization.

### 28.4 DialogWidget jQuery UI Integration
`DialogWidget` renders as a jQuery UI dialog in the browser. The `options` dict supports 'max' for maximized dimensions and standard jQuery UI dialog options.

**Testing:** Create a DialogWidget with `options={'height': 'max'}`. Confirm the dialog fills the available height in the browser. Set a specific width and confirm the dialog respects it.

### 28.5 ModalWidget Bootstrap Integration
`ModalWidget` renders as a Bootstrap modal. `on_close(callback)` registers callbacks that fire with the button name when the modal is dismissed (OK or Cancel).

**Testing:** Create a ModalWidget and display it. Click OK; confirm the callback fires with 'ok'. Click Cancel; confirm it fires with 'cancel'. Close with the X button; confirm the callback fires appropriately.

### 28.6 ModalMessagesWidget: New Message
`ModalMessagesWidget.new_message(title, body)` sends a JSON message to the JS view to display a modal message. The JS shows the message as a Bootstrap modal or alert.

**Testing:** Call `new_message("Test Title", "Test body")`. Confirm a modal dialog appears in the browser with the correct title and body. Dismiss it; confirm no error occurs.

### 28.7 ExecuteNewCell Operation
`ExecuteNewCell(code)` queues Python code to execute in a new Jupyter cell. `submit(on_done)` registers a `post_run_cell` callback and executes the code. `on_done` is called with the execution result.

**Testing:** Submit `ExecuteNewCell("2+2")`. Confirm a new cell appears in the notebook and executes. Confirm `on_done` is called with the result. Confirm the cell is visible in the notebook history.

### 28.8 ShowModal Operation
`ShowModal(title, children)` creates a ModalWidget and displays it when `submit(on_done)` is called. `on_done` is called with True (OK) or False (Cancel).

**Testing:** Submit a `ShowModal`. Confirm the modal appears with the correct title and children. Click OK; confirm `on_done(True)` fires. Click Cancel; confirm `on_done(False)` fires.

### 28.9 UserSelect Operation
`UserSelect(options, title, prompt)` shows a modal with a single-selection dropdown. The selected value is passed to `on_done`.

**Testing:** Submit a `UserSelect` with three options. Confirm the modal shows a dropdown. Select the second option and click OK. Confirm `on_done` receives the second option's value.

### 28.10 UserSelectMultiple Operation
`UserSelectMultiple(options, title, prompt)` shows a modal with a `SelectMultiple` widget. A list of selected values is passed to `on_done`.

**Testing:** Submit a `UserSelectMultiple` with five options. Select two non-adjacent options. Click OK. Confirm `on_done` receives a list containing exactly those two options. Click Cancel; confirm `on_done(None)`.

---

## 29. Extension Points (ui_extensions_api.py)

### 29.1 ExtensionPoint Registration
`ExtensionPoint.register(function)` adds a callback to be called when the extension point fires. The `@ep.action("label")` decorator registers functions as UI actions. The decorated function is wrapped with `interaction()`.

**Testing:** Register a new callback on `arg_node_actions`. Trigger a node right-click. Confirm the registered action appears in the context menu. Click it and confirm the callback fires. Unregister and confirm it no longer appears.

### 29.2 arg_node_actions Extension Point
`arg_node_actions(node)` collects context menu actions from all registered callbacks for a given ARG node. Default registered callbacks include `execute_actions` and `try_conjectures`.

**Testing:** Right-click an ARG node. Confirm the default extension actions appear. Register a custom action and confirm it appears. Unregister it and confirm it disappears.

### 29.3 goal_node_actions Extension Point
`goal_node_actions(node)` collects context menu actions for proof goal nodes. Currently no default callbacks registered.

**Testing:** Right-click a proof goal node. Confirm the default empty menu (or whatever is registered). Register a custom callback and confirm the action appears in proof goal node menus.

### 29.4 execute_actions Default Callback
The `execute_actions(s)` callback returns the list of actions from the current analysis session that can be applied to state `s`. These appear in the ARG node context menu under "Execute actions".

**Testing:** Open a model with multiple actions. Right-click an ARG node. Confirm all applicable actions from the ivy module appear. Confirm unrelated actions do not appear.

### 29.5 try_conjectures Default Callback
The `try_conjectures(s)` callback returns the list of unproven conjectures that can be applied at state `s`. These appear in the ARG node context menu.

**Testing:** Define multiple conjectures in the ivy model. Right-click an ARG node. Confirm the unproven conjectures appear in the menu. Prove one and confirm it no longer appears.

### 29.6 @interaction Decorator
The `@interaction` decorator wraps generator functions to support async operations (yielding `FrontEndOperation` objects). It checks if the session is at the last history step.

**Testing:** Use a decorated interaction function that yields a `UserSelect`. Confirm the UI shows the selection dialog and waits for user input. Confirm the function resumes after the selection.

### 29.7 arg_check_cover Action
`arg_check_cover(node)` validates that exactly one node is selected (raises `InteractionError` otherwise), then yields a `UserSelect` for choosing the covering node. Then yields `ExecuteNewCell` to call `check_cover()`.

**Testing:** Select a single node and invoke `arg_check_cover`. Confirm a selection dialog for the covering node appears. Select a node and confirm `check_cover` is called. Try with zero or two nodes selected and confirm an `InteractionError`.

### 29.8 arg_remove_facts Action
`arg_remove_facts(node)` yields a `UserSelectMultiple` listing available facts for the node. The user selects facts to remove, then `remove_facts()` is called in a new cell.

**Testing:** Invoke `arg_remove_facts` on a node with multiple facts. Confirm a multi-select dialog appears with all facts. Select some and click OK. Confirm `remove_facts()` removes them. Click Cancel; confirm no facts are removed.

### 29.9 arg_join2 Action
`arg_join2(node)` validates selection, yields a `UserSelect` for the second join node, then yields `ExecuteNewCell` for `join2()`. The `InteractionError` is raised if selection is invalid.

**Testing:** Select a node and invoke `arg_join2`. Confirm a selection dialog for the second node appears. Select a compatible node; confirm join2 executes. Try with invalid selection and confirm `InteractionError`.

---

## 30. Concept Session Behaviors

### 30.1 Domain Undo/Redo
`ConceptInteractiveSession.push()` saves the current domain to `undo_stack`. `pop()` restores from it. `undo()` calls pop and recomputes.

**Testing:** Make a domain change (add concept). Click Undo. Confirm the domain reverts. Make another change. Confirm undo/redo history is maintained for multiple operations.

### 30.2 Recompute After Domain Change
After `undo()`, `split()`, `remove_concepts()`, `suppose_empty()`, `materialize_node()`, or `materialize_edge()`, `recompute()` is called to update `abstract_value`. The widget is then asked to `render()`.

**Testing:** Make a domain change. Confirm the concept graph immediately re-renders. Confirm the new abstract value is reflected in node/edge classes and labels.

### 30.3 Save/Load Domain
`save_domain(name)` stores a snapshot of the domain under `name`. `load_domain(name)` restores it. Multiple snapshots can be stored.

**Testing:** Save the domain as "checkpoint1". Make changes. Load "checkpoint1". Confirm the domain reverts exactly to the saved state. Confirm the concept graph reflects the restored domain.

### 30.4 Replace Domain
`replace_domain(new_domain, new_suppose_constraints)` swaps the entire concept domain and suppose constraints. This is used by "Reset domain" and "Diagram domain" operations.

**Testing:** Click "reset domain" in the concept widget. Confirm the domain reverts to the initial concept domain. Confirm suppose constraints are cleared. Confirm the concept graph re-renders.

### 30.5 Get Projections
`get_projections(node)` finds ternary relations where `node` appears as one dimension and returns projection concepts. These appear in the node context menu.

**Testing:** Open a concept graph for a model with ternary relations. Right-click a node. Confirm "Projections" submenu appears with applicable projections. Click a projection and confirm it is added to the concept display.

### 30.6 Add Custom Edge
`add_custom_edge(edge, source, target)` adds a specific (edge, source, target) combination to the domain display.

**Testing:** Add a custom edge via the UI or programmatically. Confirm the new edge appears in the concept graph. Confirm it has the correct source and target nodes.

---

## 31. One-Step Reachability

### 31.1 one_step_reach()
`one_step_reach(state, clauses)` computes states reachable in one step from known reachable states, constrained by `clauses`. It adds new states to the reachable tree and displays eliminated conjectures.

**Testing:** Call one_step_reach after marking some states reachable. Confirm new reachable states are added to the reachable tree. Confirm the reachable tree display updates.

### 31.2 Eliminated Conjectures Dialog
If conjectures are eliminated during one_step_reach (because a reachable state violates them), a `listbox_dialog` shows the eliminated conjectures. The on_cancel is `None` so only OK is shown.

**Testing:** Set up a conjecture that is violated by a reachable state. Call one_step_reach. Confirm a listbox dialog appears listing the eliminated conjectures. Click OK and confirm the dialog closes.

---

## 32. Visual Styling Details

### 32.1 ARG Node: Green for Safe
`node_color(node)` returns `"green"` if `node.safe` is True. This is used by `update_node_color()` to paint the node outline green after a successful safety check.

**Testing:** Run a safety check that succeeds. Confirm the verified node's outline turns green. Confirm unsafe nodes remain black. Confirm color changes are immediate and persistent.

### 32.2 ARG Node: Black Default Outline
All ARG nodes start with `outline='black'` (from `get_node_styles()`). The width is 2. This is the default visual state before any safety checking.

**Testing:** Open a fresh ARG. Confirm all node outlines are black. Confirm outline width is 2 (visually consistent, not thin or thick).

### 32.3 Concept Node: Grey Background
The default concept graph node background is `#888` (grey). Text is white with a 3px `#888` text outline for readability on the grey background.

**Testing:** Open a concept graph. Confirm nodes have a grey background. Confirm node labels are white and readable. Confirm the grey color matches `#888` exactly.

### 32.4 Concept Node: Non-Existing Hidden
Nodes with class `non_existing` have `display: none` in `concept_style`. They are not visible in the rendered graph.

**Testing:** Create a scenario where a sort is empty. Confirm no node appears for that sort. Confirm the hidden node does not affect layout (no gap where it would appear).

### 32.5 ARG Bottom State: Black Background
Nodes with class `bottom_state` have `background: black` and `text-outline-color: black` in `arg_style`. This makes them appear as solid black circles.

**Testing:** Trigger an error condition creating a bottom_state. Confirm it appears as a solid black node. Confirm the label (if any) is not visible against the black background.

### 32.6 Proof Goal: Refuted State
Proof goal nodes with class `refuted` have `background: black` in `proof_style`. This marks goals that have been shown to be unreachable.

**Testing:** Refute a proof goal via BMC. Confirm the corresponding node in the proof stack turns black. Confirm non-refuted goals remain grey.

### 32.7 Selection Overlay
`:selected` elements have `overlay-opacity: 0.2` in all style sheets. This adds a semi-transparent overlay when elements are selected.

**Testing:** Click to select a node or edge in a Cytoscape-based concept graph. Confirm a slight darkening or overlay appears on the selected element. Deselect and confirm it returns to normal.

### 32.8 Transition Action Edge: Triangle Arrow
Transition action edges have `target-arrow-shape: triangle` and `target-arrow-fill: filled` in `arg_style`.

**Testing:** Create an action transition in the ARG. Confirm the edge terminates with a filled triangle arrowhead. Confirm the arrowhead size is consistent.

### 32.9 Transition Join Edge: Backcurve Arrow
Join transition edges have `target-arrow-shape: triangle-backcurve`. This distinguishes joins from regular transitions.

**Testing:** Join two states in the ARG. Confirm the resulting join edge has a backcurve (concave) arrowhead distinct from normal triangle arrows.

### 32.10 Cover Edge: Dashed, No Label
Cover edges have `line-style: dashed` and no label (`content: ''`). The arrow shape is `triangle`.

**Testing:** Create a covering relation. Confirm the cover edge is dashed. Confirm no label text appears on the cover edge. Confirm it has a triangle arrowhead.

### 32.11 Edge Text Rotation
ARG edges have `edge-text-rotation: none` in `arg_style`. This keeps edge labels horizontal even for diagonal edges.

**Testing:** Create an ARG with angled edges. Confirm edge labels remain horizontal (not rotated to follow the edge angle). Confirm labels are readable regardless of edge angle.

### 32.12 Concept Edge: Total (Circle Source Arrow)
Edges with class `total` have `source-arrow-shape: circle`. This indicates a total function (every element has an image).

**Testing:** Add a total function concept. Confirm the edge has a circle at the source end. Confirm non-total functions do not have this circle.

### 32.13 Concept Edge: Functional (Square Source Arrow)
Edges with class `functional` have `source-arrow-shape: square`. This indicates at-most-one image per element.

**Testing:** Add a functional relation. Confirm the edge has a square at its source. Confirm non-functional edges lack the square.

### 32.14 Concept Edge: Injective (Backcurve Target)
Edges with class `injective` have `target-arrow-shape: triangle-backcurve`. Injective means at-most-one preimage per element.

**Testing:** Define an injective relation. Confirm the target end has a backcurve arrow. Confirm non-injective relations use a plain triangle.

### 32.15 Concept Edge: Surjective (Filled Target Arrow)
Edges with class `surjective` have `target-arrow-fill: filled` for the target arrow. Surjective means every element has a preimage.

**Testing:** Define a surjective relation. Confirm the target arrowhead is filled. Confirm non-surjective relations have an open (hollow) arrow.

### 32.16 Node Text Wrapping
Concept graph nodes have `text-wrap: wrap` in `concept_style`. Long labels are wrapped rather than overflowing the node boundary.

**Testing:** Create a concept with a very long name. Confirm the node label wraps to multiple lines. Confirm the node expands to accommodate wrapped text.

### 32.17 Font Size 14px
All concept graph nodes and edges use `font-size: 14px` in `concept_style`. The ARG graph uses the same size.

**Testing:** Open a concept graph and inspect font size. Confirm labels render at 14px. Zoom in to confirm the font size is consistent.

### 32.18 Concept Graph Line Colors (27 Colors)
`TkGraphWidget.line_colors` contains 27 named colors used cyclically for relation edge rendering in the Tk version. Colors cycle when more than 27 relations are present.

**Testing:** Open a model with 28+ relations. Confirm colors cycle (28th relation uses same color as 1st). With fewer relations, confirm each gets a distinct color. Colors include blue, green, red, yellow, magenta, pink, purple, black, white.

---

## 33. Miscellaneous UI Behaviors

### 33.1 tk.update_idletasks() Before Dialogs
In `ivy_check.py`, `gui.tk.update_idletasks()` is called before dialogs to ensure the main window is fully rendered before a dialog appears on top of it.

**Testing:** Trigger a diagnose-mode dialog. Confirm the main window is fully rendered before the dialog appears. Confirm the dialog appears above the main window, not behind it.

### 33.2 Analyze Session Widget set_context()
`set_context(analysis_session_widget)` sets global context variables used by the extension API. This allows extension point callbacks to reference the current session.

**Testing:** After `set_context()`, confirm that `_analysis_session_widget`, `_analysis_session`, `_concept_session_widget`, and `_concept_session` globals are populated. Confirm extension callbacks can access the current session through these globals.

### 33.3 Proof Goal Click Updates Concept Widget
`AnalysisSessionWidget.proof_node_click(goal)` updates the concept session widget with the clicked proof goal's information.

**Testing:** Click a proof goal node in the proof graph. Confirm the concept widget updates to display the goal's constraints. Confirm clicking different goals shows different constraints.

### 33.4 CRG Node Click Updates Concept and Transition Widgets
`crg_node_click(crg_node)` updates both the concept domain and, if the node has a transition, the transition view widget's pre/post states.

**Testing:** Click a CRG (Concrete Reachability Graph) node. Confirm the concept widget updates. If the node represents a transition, confirm the transition view widget shows the pre and post states. Click a non-transition node; confirm no transition view update.

### 33.5 Abastractor Selection Dropdown
In `AnalysisSessionWidget`, a dropdown `select_abstractor` offers abstraction strategies: 'top/bottom', 'concrete', 'propagate', 'propagate & conjectures', 'concept space'. The default is 'ta.Abstractors.top_bottom'.

**Testing:** Select each abstraction strategy. Confirm the analysis behavior changes accordingly. Select 'concrete' and confirm states are not abstracted. Select 'concept space' and confirm concept-based abstraction is used.

### 33.6 BMC Bound Dropdown in Transition Widget
`TransitionViewWidget` has a `Dropdown` widget with BMC bound options [1, 3, 5, 10, 15], defaulting to 3. The selected value is used as the default bound for `bmc_conjecture()`.

**Testing:** Set the BMC bound dropdown to 5. Trigger "bmc conjecture". Confirm BMC runs with bound 5 without showing an int_dialog. Change to 10 and confirm the bound changes.

### 33.7 Relations to Minimize Text Input
`TransitionViewWidget` has a `Text` widget for entering relation names to minimize. The default text is 'relations to minimize'. This input is used by `minimize_conjecture()`.

**Testing:** Enter specific relation names in the "relations to minimize" field. Click "minimize conjecture". Confirm the minimization focuses on those relations. Confirm the default value does not cause errors if not changed.

### 33.8 Transition View Widget Log File
`TransitionViewWidget.register_session(session)` creates a log file named `tvw_log_{filename}`. Operations like `check_inductiveness`, `bmc_conjecture`, etc. are logged with metadata via `log()`.

**Testing:** Register a session. Perform several operations (check inductiveness, bmc). Confirm a log file is created. Check the log file contains entries for each operation with appropriate metadata.

### 33.9 Gather Facts: Node vs Edge Distinction
`gather_facts()` in `ConceptStateViewWidget` distinguishes single-element (node) tuples from three-element (edge) tuples when collecting facts. Node facts are unary predicates; edge facts are binary.

**Testing:** Select a mix of nodes and edges in the concept graph. Click "gather facts". Confirm the facts list includes both node facts (formatted as `P(x)`) and edge facts (formatted as `R(x,y)`). Verify no type errors or mixing.

### 33.10 Facts List: SelectMultiple Widget
The `facts_list` in concept view widgets is a `SelectMultiple` widget. Multiple facts can be selected. `get_active_facts()` returns all selected facts, or all facts if none selected.

**Testing:** Gather multiple facts. Select a subset in the SelectMultiple. Call `get_active_facts()` and confirm only the selected subset is returned. Deselect all and confirm all facts are returned.

### 33.11 Fact Highlighting in Concept Graph
`highligh_selected_facts()` (in `TransitionViewWidget`) collects atoms from selected facts and adds custom edges and node labels to highlight them in the pre/post graphs.

**Testing:** Select some facts in the transition view. Confirm the corresponding nodes and edges in the concept graph become visually highlighted. Deselect facts and confirm highlights are removed.

### 33.12 apply_structure_renaming()
`apply_structure_renaming(st)` applies a dictionary of string substitutions to rename structures in the display. The `TransitionViewWidget` version applies `structure_renaming` dict; the base version is a no-op.

**Testing:** Set a structure renaming (e.g., rename 'node_0' to 'client'). Confirm node labels in the concept graph show the renamed version. Confirm the renaming is applied consistently across all labels.

### 33.13 Concept Domain Initial Snapshot
On startup, the initial concept domain is saved as `'initial'` in the domain snapshots. "Reset domain" restores this snapshot.

**Testing:** Make concept domain changes. Click "reset domain". Confirm the domain reverts to exactly the state at startup. Confirm all added concepts, removed concepts, and suppose constraints are cleared.

### 33.14 Witness Constants with '@' Prefix
In `check_inductiveness()` (TransitionViewWidget), witness constants are created with a `'@'` prefix character. These are unique constants used for BMC witnesses.

**Testing:** Run inductiveness check. Confirm witness constants in the resulting CTI states are prefixed with `@`. Confirm they do not conflict with user-defined constants. Confirm the concept graph checkboxes update to show the witness constant sorts.

### 33.15 Autodetect Transitive Relations
`autodetect_transitive()` scans `im.module` for relations with transitive axioms. Found relations are listed in `self.transitive_relations`. The concept graph UI shows 'T' markers for them.

**Testing:** Create a model with a declared transitive relation. Open the CTI UI. Confirm the transitive relation shows 'T' in the relation buttons without manual intervention.

### 33.16 Process Launch with xterm
`ivy_launch.py:run_in_terminal(cmd, name)` spawns `xterm` windows for distributed test processes. Each window has a descriptive title and runs the command.

**Testing:** Call `run_in_terminal("echo hello", "test_proc")`. Confirm an xterm window appears with the title containing "test_proc". Confirm the command runs in that window.

### 33.17 Sequential Port Allocation
`ivy_launch.py:get_unused_port()` returns sequential ports starting at 49123, incrementing `next_unused_port` each call. No port reuse detection.

**Testing:** Call `get_unused_port()` three times. Confirm ports 49123, 49124, 49125 are returned. Confirm global state tracks the next port correctly.

### 33.18 Diagnose Mode: --diagnose Parameter
`ivy_check.py` reads `diagnose = iu.BooleanParameter("diagnose", False)`. When True, failures open the GUI instead of raising exceptions.

**Testing:** Pass `--diagnose=true` to ivy. Trigger a property failure. Confirm the GUI opens rather than exiting with an error. Pass `--diagnose=false` (default); confirm the error is printed and the process exits.

---

## Summary Table of Menu Items

| Menu | Item | Callback | Mode |
|------|------|----------|------|
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
| View | Add relation | add_relation dialog | Concept |
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
| listbox_dialog | One-step reach | Eliminated conjectures | OK |
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
| refine_dialog | Interpolant | Formula + PDR/induction message | Refine, Cancel |
