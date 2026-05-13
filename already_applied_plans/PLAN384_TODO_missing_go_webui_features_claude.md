# PLAN384: Missing Go/Web UI Features — TODO List

## Context

This document is the result of a set-subtraction audit: the union of Python GUI
behaviors from PLAN378 (Claude's Tcl/Tk inventory, 31 sections × ~100 items) and
PLAN383 (Codex Tcl/Tk inventory, 105 numbered items) was compared against the
current Go/web UI in `~/ivy/goivy/webui/`.  The result is the TODO list below —
behaviors present in the Python GUI but absent or incomplete in the Go/web UI.

Items that are clearly **Tk-paradigm-only** (busy cursor, Tix NoteBook, `.a2g`
pickle format, `ivy2.py` notebook generation, `ivy_launch.py`, pre-seeded dialog
test answers, `--diagnose` CLI flag) are excluded as non-applicable to a web UI.
Items already fully implemented in the Go/web UI are also excluded.

---

## Already Confirmed Implemented (excluded from TODO)

For reference, the following Python behaviors were verified present in the Go/web UI
by code audit of `~/ivy/goivy/webui/frontend/src/`:

- ARG node context menu: Mark, Cover by marked, Join with marked, Extend, Delete
  (found in `ivyRuntime.ts` ~2226)
- ARG edge context menu: Dismiss, Recalculate, Step in, View Source
  (found in `ivyRuntime.ts` ~2322)
- Bottom-state black fill (`graphRuntime.ts` ~106)
- Constraint fact click-to-toggle active/inactive (`detailsService.ts`)
- Splatter node action (`ivyRuntime.ts` ~2476)
- BMC bound remembered across invocations (`checkService.ts` ~149–157)
- Transitive reduction checkbox
- All concept graph operations: Concrete, Gather, CTI Gather, Minimize, Check
  Sufficient, Check Relative Induction, Strengthen, Reverse, PDR step, Diagram,
  Reach, Path Reach, Conjecture, Backtrack, Undo/Redo, Remember, Export
- CTI check induction, weaken, save invariant, bounded check
- Event viewer tree with lazy loading, filter, find forward/backward, pattern
  save/load
- Sheet/tab system, session persistence, undo/redo stacks

---

## TODO List

### T1 — ARG Node Visual: Mark Indicator (Red Fill)

**Source**: PLAN378 §11, PLAN383 §26  
**Python behavior**: `show_mark(True)` fills the marked node's canvas shape **red**;
`show_mark(False)` clears it.  The red fill is restored after every graph rebuild
(`rebuild()` calls `show_mark(True)` at its end if a mark exists).  Only one node
is red at a time; marking a new node clears the previous one.  
**Go/web gap**: The `mark` action exists (it sends `mark` to the backend), but no
red fill is applied to the Cytoscape node element after marking.  A user performing
Cover-by-marked or Join-with-marked has no visual feedback about which node is
currently marked.  
**Fix**: After a successful `mark` action response, add/remove a Cytoscape CSS class
(e.g. `marked_node`) on the affected node.  Add a CSS rule
`node.marked_node { border-color: red; border-width: 4px; }`.  On every ARG graph
rebuild that preserves the graph structure (re-render without full reset), re-apply
the class to the stored marked node ID.

---

### T2 — Dialog: "Cannot reverse"

**Source**: PLAN378 §17, PLAN383 §58  
**Python behavior**: When `Reverse` is impossible (no predecessor state), an
`ok_dialog("Cannot reverse.")` is shown.  
**Go/web gap**: The Reverse action is present but the frontend has no handler for a
"cannot reverse" result from the backend; the condition may be silently dropped or
only appear in the status bar.  
**Fix**: Backend should return a structured result type for the Reverse action.  On
`{type: "cannot_reverse"}`, the frontend should show an OK dialog with
"Cannot reverse."

---

### T3 — Dialog: "PDR terminated"

**Source**: PLAN378 §16, PLAN383 §59  
**Python behavior**: When `pdr_step` exhausts all proof avenues (backtracks and
recalculates with no further undo/reverse result), it shows
`ok_dialog("PDR terminated")`.  
**Go/web gap**: Not found in frontend source.  
**Fix**: Backend PDR step should emit `{type: "pdr_terminated"}` when the workflow
terminates; frontend shows OK dialog "PDR terminated."

---

### T4 — Dialog: "Goal reached!" with View state

**Source**: PLAN378 §31, PLAN383 §61  
**Python behavior**: When `Reach` (one-step reachability) succeeds, shows
`ok_cancel_dialog("Goal reached!  A reachable state has been added.", view_cmd)`.
Clicking OK does nothing; clicking "View state" displays the reached state's concept
graph.  
**Go/web gap**: The Reach action exists but no success dialog with "View state"
found.  
**Fix**: Backend `reach` success response should include the reached state ID.
Frontend shows OK/Cancel dialog; "View state" triggers `view_state` for the returned
state ID.

---

### T5 — Dialog: "The current state is vacuous"

**Source**: PLAN383 §60  
**Python behavior**: During `Diagram`, if the current state is vacuous (no satisfying
model), shows `ok_dialog("The current state is vacuous.")`.  When this arises from a
reverse-result diagram attempt, Python backtracks before retrying.  
**Go/web gap**: Not found.  
**Fix**: Backend `diagram` response should include `{type: "vacuous"}`.  Frontend
shows OK dialog; if the diagram was called as part of a PDR-step sequence, also
trigger backtrack+retry.

---

### T6 — Dialog: "The pre-state is vacuous" (interpolant refinement)

**Source**: PLAN378 §17, PLAN383 §58  
**Python behavior**: During interpolant-based reverse, if the pre-state is vacuous,
`refine_with_interpolant` shows "The pre-state is vacuous..." instead of the
interpolant formula.  
**Go/web gap**: Interpolant refinement dialog not present at all (see T7).  
**Fix**: Address as part of T7.

---

### T7 — Interpolant Refinement Dialog

**Source**: PLAN378 §17, PLAN383 §58–59  
**Python behavior**: When `Reverse` finds the goal infeasible but produces an
interpolant / separating formula, a **text dialog** shows the formula with a **Refine**
action button.  The message differs by mode:
- PDR mode: "PDR found the following invariant: ..."
- Other modes: "Found the following separating formula: ..."
Clicking Refine adds the interpolant as a predicate to the abstract domain (PDR) or
as a concept (other modes).  
**Go/web gap**: No such dialog found in frontend.  
**Fix**: Backend Reverse response should include `{type: "refine", formula: "...", mode: "pdr"|"other"}`.
Frontend shows a text dialog with mode-appropriate heading and a "Refine" button that
calls `addPredicate` / `addConcept`.

---

### T8 — Eliminated Conjectures Listbox Dialog (one_step_reach)

**Source**: PLAN378 §31, PLAN383 §61  
**Python behavior**: When one-step reachability eliminates conjectures (a reachable
state violates them), a `listbox_dialog("The following conjectures were eliminated:",
conj_list)` is shown.  OK dismisses; no cancel is offered.  
**Go/web gap**: Not found.  
**Fix**: Backend `reach` response should include `{eliminated: [...]}` when
conjectures are violated.  Frontend shows read-only listbox dialog.

---

### T9 — Session / Proof History Step Navigation

**Source**: PLAN378 §21, PLAN383 §90  
**Python behavior** (Jupyter `AnalysisSessionWidget`): **First / Prev / Next / Last**
navigation buttons let the user walk through the history of proof steps.  Each step
shows: rendered ARG, concept graph (CRG), reachable-state graph, and a `step_box`
description of the operation at that step.  Clicking a navigation button auto-clicks
the active element of the step (arg_node, crg_node, or CRG state).  
**Note**: The Tk UI does not have this; it is a Jupyter-widget feature.  
**Go/web gap**: No step history navigation found.  The Go/web UI has undo/redo for
concept graph domain changes but not a sequential read-only walkthrough of all proof
actions taken in the session.  
**Fix** (medium complexity): Maintain an ordered `history[]` array in the session
model.  Each proof-mutating action appends a history record containing: action name,
timestamp, ARG snapshot, concept graph snapshot, step description.  Add First / Prev
/ Next / Last toolbar buttons that render the selected history entry read-only,
restoring the ARG and concept graph view without re-executing computations.

---

### T10 — Interactive UPDR Dialog (UserSelectCore)

**Source**: PLAN378 §29, PLAN383 §88  
**Python behavior**: `iupdr.interactive_updr` is a generator interaction that
repeatedly asks the user to:
1. Select diagram literals for generalization (all selected by default, `SelectMultiple` list).
2. Select unsat-core literals for refinement; the dialog has a **Check SAT** button
   that updates a result label to SAT / UNSAT without closing the dialog.
OK returns `(selected_constraints, is_sat_bool)`; Cancel returns `(None, None)`.  
**Go/web gap**: Interactive UPDR dialog not found.  
**Fix**: Add `UserSelectCore` dialog component: a modal with a multi-select list,
a "Check SAT" button (triggers backend check without closing), and a result label.
Wire into the `iupdr` backend interaction protocol.

---

### T11 — Tactic Message Modal (ModalMessagesWidget)

**Source**: PLAN378 §28, PLAN383 §89  
**Python behavior**: When a tactic step emits a message (the tactic's `msg` field is
set), a modal dialog titled **"Tactic \<name\> says:"** is shown with the message body.
Multiple messages may be shown per step; `silent=True` suppresses them.  
**Go/web gap**: Tactic message modal not found.  
**Fix**: Backend tactic step response should include `{tactic_messages: [{title, body}]}`.
Frontend iterates the list and shows each as a dismissible modal or toast.  A
`silent` flag on the session can suppress them.

---

### T12 — Transitive Reduction: Self-Edges (A→A) Hidden

**Source**: PLAN378 §18, PLAN383 §100  
**Python behavior**: `get_transitive_reduction` first hides all **self-edges** (A→A)
for a relation whose `T` checkbox is on, then hides edges implied by a two-step
path.  Custom edges are exempt.  
**Go/web gap**: The transitive reduction toggle exists, but whether self-edges are
explicitly hidden is unclear.  
**Fix**: In the transitive reduction pass, add an initial filter: for any edge where
`source === target`, exclude it from the rendered edge set when `T` is enabled.

---

### T13 — Numeral Equality Default Labels Shown Automatically

**Source**: PLAN383 §99  
**Python behavior**: `Graph.default_labels` returns the set of **numeral equality
concepts** (e.g., `X=0`, `X=1`).  During checkbox sync, these always get their
display checkbox enabled (index 0 enabled) regardless of user state.  This makes
witness numeral labels visible by default without the user toggling them.  
**Go/web gap**: No evidence of numeral equality concepts being auto-enabled.  
**Fix**: After graph state update, identify node-label concepts whose name matches
the numeral equality pattern (e.g., `=<numeral>:<sort>`).  Force their visibility
checkbox on in the concept visibility service regardless of user toggle state.

---

### T14 — Equality Label Abbreviation (Full Form)

**Source**: PLAN378 §18, PLAN383 §102  
**Python behavior**: `Graph.concept_label` abbreviates:
- `p(X)` → `p`
- `X=e` → `=e` (equality concept)
- `f(X)=e` → `f=e`
- Complex formulas → full string

The Go/web `displayConceptName` handles the `=<body>` prefix case but the full
abbreviation for `p(X)→p` and `f(X)=e→f=e` patterns is unclear.  
**Go/web gap**: Partial.  
**Fix**: Extend `displayConceptName` (in `conceptVisibilityService.ts`) to cover
all three cases per `Graph.concept_label`.

---

### T15 — Concept Graph: Relation Color Consistency

**Source**: PLAN378 §19, PLAN383 §40  
**Python behavior**: Relation row labels and that relation's edges use the **same
color**.  Nodes of the same sort share an outline color; different sorts get different
colors.  Colors cycle through a fixed palette.  
**Go/web gap**: Need to verify that the relation row label color and the matching
Cytoscape edge color are always drawn from the same palette index.  
**Fix**: Audit `conceptVisibilityService.ts` color assignment to confirm row-label
color and edge style color are derived from the same palette slot for each relation.

---

### T16 — Concept Graph Node Shape: Octagons for `__ID` Concepts

**Source**: PLAN378 §18, PLAN383 §103  
**Python behavior**: `get_shape(concept_name)` returns `'octagon'` for `__ID`
concepts and `'ellipse'` for all others.  
**Go/web gap**: Not confirmed.  
**Fix**: In the Cytoscape node style rules, add
`{ selector: 'node[id *= "__ID"]', style: { shape: 'octagon' } }`.

---

### T17 — ARG: "State N is closed" When Extend Finds No Extensions

**Source**: PLAN378 §14, PLAN383 §30  
**Python behavior**: `find_extension` shows `ok_dialog("State {node.id} is closed.")`
when no applicable state actions exist.  
**Go/web gap**: The backend likely returns an empty-extension result for `extend`;
the frontend may only show a status message rather than a blocking OK dialog.  
**Fix**: When backend `extend` response indicates no extensions available, show an
OK dialog with "State \<id\> is closed."

---

### T18 — Save Invariant: Kept / Dropped / New Sections

**Source**: PLAN378 §20, PLAN383 §70  
**Python behavior**: The saved `.ivy` invariant file has three annotated sections:
1. `# Original conjectures kept` — active `invariant [label] formula;` lines
2. `# Original conjectures dropped` — commented-out originals
3. `# New conjectures` — unlabeled new strengthenings

Labels from original conjectures are preserved in the `[label]` syntax.  
**Go/web gap**: The Go/web "Save Invariant" action exists.  Need to verify the
output format matches this three-section layout.  
**Fix**: Audit the invariant serializer in the Go backend to confirm it writes all
three sections.  If not, update it to match Python's format.

---

### T19 — CTI Startup: Immediately View Initial State (Node 0)

**Source**: PLAN383 §65  
**Python behavior**: CTI `start()` calls `view_state(node_0)` so the **concept
graph for the initial state is visible immediately** on launch, without the user
needing to click a node.  
**Go/web gap**: Not confirmed.  The Go/web CTI start may leave the concept graph pane
empty until the user clicks a node.  
**Fix**: After CTI session initialization completes, automatically dispatch
`view_state` for node 0 to populate the concept graph pane.

---

### T20 — ARG Node Safety Color: Green on Safe

**Source**: PLAN378 §12, PLAN383 §22  
**Python behavior**: After a successful safety check, `update_node_color(node)` sets
the node outline to **green**.  Failed/unchecked nodes remain black.  
**Go/web gap**: Need to verify that the Cytoscape ARG node border turns green after
a successful `check_safety` response.  
**Fix**: Backend `check_safety` success response should include `{node_id, safe: true}`.
Frontend adds/removes a `safe_node` Cytoscape class with CSS rule
`{ border-color: green, border-width: 3px }`.

---

## Summary Table

| # | Feature | Confidence | Effort |
|---|---------|-----------|--------|
| T1 | Mark node red fill | High — action exists, fill missing | Small |
| T2 | "Cannot reverse" dialog | High — dialog not found | Small |
| T3 | "PDR terminated" dialog | High — not found | Small |
| T4 | "Goal reached!" dialog + View state | High — not found | Small |
| T5 | "Current state vacuous" dialog | High — not found | Small |
| T6 | "Pre-state vacuous" message | High — part of T7 | Small |
| T7 | Interpolant refinement dialog | High — not found | Medium |
| T8 | Eliminated conjectures listbox | High — not found | Small |
| T9 | Session history navigation | High — not found | Large |
| T10 | Interactive UPDR dialog | High — not found | Large |
| T11 | Tactic message modal | High — not found | Medium |
| T12 | Self-edge hiding in transitive reduction | Medium — partial | Small |
| T13 | Numeral equality auto-labels | Medium — not found | Small |
| T14 | Equality label abbreviation (full) | Medium — partial | Small |
| T15 | Relation color consistency audit | Low — may be fine | Small |
| T16 | `__ID` octagon shape | Medium — not confirmed | Small |
| T17 | "State N is closed" dialog | Medium — may route correctly | Small |
| T18 | Save invariant 3-section format | Medium — needs audit | Small |
| T19 | CTI auto-view node 0 on start | Medium — not confirmed | Small |
| T20 | ARG node green on safe | Medium — needs audit | Small |

---

## Not In Scope (Different Paradigm)

| Python feature | Reason excluded |
|---|---|
| `--diagnose` CLI flag opens GUI on failure | Web UI is always interactive |
| Source file `FileBrowser` window (separate, reusable) | Go/web uses in-editor line jump (functionally equivalent) |
| Busy cursor (`cursor='watch'`) | Go/web has loading spinner overlay |
| Pre-seeded dialog answers for automated testing | Web testing done differently |
| `ivy2.py` Jupyter notebook generation | Separate product |
| `ivy_launch.py` distributed process launcher | Separate product |
| `.a2g` pickle protocol-2 format | Go/web uses JSON (deliberate choice) |
| Tk palette white background | Browser CSS handles theming |
| Dynamic Tk class mixing | Go/web uses command registry pattern |
| `ConceptGraphUI` left-click = direct select toggle (no popup) | Go/web has equivalent context-menu select |

---

## Source Files to Modify (Primary)

- `~/ivy/goivy/webui/frontend/src/services/ivyRuntime.ts` — ARG action handlers, mark visual (T1, T17, T20)
- `~/ivy/goivy/webui/frontend/src/services/checkService.ts` — dialog wiring for verification results (T2–T8, T17)
- `~/ivy/goivy/webui/frontend/src/services/conceptVisibilityService.ts` — self-edge hiding, numeral defaults, label abbreviation (T12, T13, T14, T15)
- `~/ivy/goivy/webui/frontend/src/services/graphRuntime.ts` — Cytoscape CSS rules (T1, T16, T20)
- `~/ivy/goivy/webui/frontend/src/services/conceptActionService.ts` — CTI startup auto-view (T19)
- `~/ivy/goivy/webui/static/css/ivy.css` — mark node red, safe node green, octagon CSS (T1, T16, T20)
- `~/ivy/goivy/webui/webui_ui_main.go` — backend response types for new dialogs
- New component: tactic message modal (T11)
- New component: interactive UPDR dialog (T10)
- New feature: session history navigation (T9)
