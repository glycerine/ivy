# PLAN386: Missing Go/Web UI Features — TODO List

## Context

This document is the result of a "set subtraction" audit: the union of Python/Tcl GUI behaviors
inventoried in PLAN378 (all sections read in full) and PLAN383 (read in full) minus the behaviors
already confirmed as implemented in the Go/web UI (`~/ivy/goivy/webui/`). The Go/web UI
inventory was produced by reading the full frontend TypeScript source (`frontend/src/`), CSS
(`static/css/ivy.css`), and Go backend files (`webui_session.go`, `webui_cyrender.go`,
`webui_cystyles.go`, `webui_handlers.go`, `webui_graph_widget.go`, etc.), plus two targeted
feature-verification agent passes.

A feature is counted as **missing** if it is absent, stubbed, or lacks test coverage in the
Go/web UI. Each item below is a concrete TODO with file locations.

Source inventory references: PLAN378 §N.N (Claude inventory), PLAN383 §N (Codex inventory).

---

## SECTION 1 — Confirmed Missing: Must Implement

### 1.1 ARG Node Green Highlight on Successful Safety Check

**Python behavior (PLAN378 §12.3 / PLAN383 §22):** After `check_safety_node` succeeds,
`update_node_color` sets the node outline to **green** (`node_color` returns `"green"` when
`node.safe` is True). On failure the outline stays black.

**Go/web status:** `ARGStyle()` in `webui_cystyles.go` defines `bottom_state` (black) and
default gray only. No `safe` Cytoscape class exists. Backend tracks safety results
(`CheckLocalSafety`, `CheckBoundedSafety` in `webui_session.go`) but never emits a `safe` CSS
class in ARG element JSON.

**Files to change:**
- `webui_cystyles.go` — add `.safe` Cytoscape selector with green border-color
- `webui_cyrender.go` / ARG serialization — emit `safe` class on nodes whose state has `safe=true`
- `frontend/src/services/graphRuntime.ts` — apply class update when `arg_updated` SSE fires

**Test:** Run a safety check that passes; verify the ARG node's border turns green. Failure
case stays black.

---

### 1.2 ARG Node Red Fill for Marked State

**Python behavior (PLAN378 §11.1 / PLAN383 §26):** `mark_node` stores the node and calls
`show_mark(True)`, filling the node **red**. Rebuild restores the red fill (§11.2). Only one
node is red at a time (§11.3). The mark is consumed by Cover by Marked and Join with Marked.

**Go/web status:** `mark` action in `webui_session.go` (line 2130) calls `ui.MarkNode()` but
`ARGStyle()` in `webui_cystyles.go` has no `marked` class or red fill. No visual indicator is
applied.

**Files to change:**
- `webui_cystyles.go` — add `.marked` selector with `background-color: red`
- ARG JSON serialization — include `marked_state_id` field so frontend knows which node is marked
- `frontend/src/services/graphRuntime.ts` — toggle `.marked` class on ARG node matching
  `marked_state_id`; also re-apply after graph rebuild

**Test:** Mark a node; verify it fills red. Mark a second; first loses red, second gains it.
After a graph rebuild, the marked node retains red fill.

---

### 1.3 T Checkbox: Actual Transitive Reduction Computation

**Python behavior (PLAN378 §18.7 / PLAN383 §100):** When the T checkbox is active for a
relation with `all_to_all` facts, `get_transitive_reduction()` **removes** self-edges and edges
implied by a two-step path (A→C hidden when A→B and B→C both exist). This is a computed filter,
not CSS visibility.

**Go/web status:** `conceptVisibilityService.ts` (line 81) registers the `transitive` checkbox.
No `get_transitive_reduction` function found in Go or TypeScript source. The checkbox likely
toggles display only without computing the reduction.

**Files to change:**
- `webui_cyrender.go` or a new `webui_transitive_reduction.go` — implement
  `getTransitiveReduction(elements []CyElement, relName string) []CyElement` that removes
  self-edges and two-step-implied edges
- Wire into concept graph element computation when T toggle is active for a relation

**Test:** Create concept graph with relation R and edges A→B, B→C, A→C, A→A. Enable T for R.
Verify A→A and A→C are hidden while A→B and B→C remain.

---

### 1.4 Bulk Relation Toggle: Click Relation Name Toggles All Three Columns

**Python behavior (PLAN378 §19.1 / PLAN383 §85):** Clicking the relation-name button toggles
all three edge display checkboxes (`+`, `?`, `-`) for that relation simultaneously. Also, clicking
a column header (`+`, `?`, `-`) toggles that class across ALL relations at once.

**Go/web status:** `conceptVisibilityService.ts` (lines 99–105) attaches a click listener on
the relation-name `<a>` element that only calls `event.preventDefault()`. No toggle logic
implemented.

**Files to change:**
- `frontend/src/services/conceptVisibilityService.ts` — implement click handler on relation-name
  link: toggle the `none_to_none`, `edge_unknown`, and `all_to_all` checkbox states for that row
- Also implement column-header button click to toggle the same column across all relation rows

**Test:** Click a relation name; all three column boxes for that relation flip. Click `+` header;
all relations' positive boxes flip simultaneously.

---

### 1.5 Entry Dialogs: Bind Enter/Return Key to Confirm

**Python behavior (PLAN378 §22.8 / PLAN383 §15):** `entry_dialog` focuses the entry field and
binds `<Return>` to trigger the OK action. Applies to Add Relation, Remember goal, pattern
search, and integer (BMC bound) dialogs.

**Go/web status:** `ivyRuntime.ts` dialog code (lines 3935–3960, 3857–3867) only binds Escape
(line 3859). `_addDialogInput()` and `_addDialogIntInput()` have no `keydown` handler for Enter.

**Files to change:**
- `frontend/src/services/ivyRuntime.ts` — in `_addDialogInput()` and `_addDialogIntInput()`,
  add: `input.addEventListener('keydown', e => { if (e.key === 'Enter') confirmButton.click(); })`

**Test:** Open Add Relation, type a name, press Enter — relation is added without clicking OK.
Open Bounded Check, type a number, press Enter — check starts.

---

### 1.6 ARG Edge Context Menu: "Dismiss" Action

**Python behavior (PLAN378 §10.1):** Right-clicking an ARG edge shows "Dismiss" as the first
entry. Dismiss removes the edge from the view by calling `decompose_edge` and catching failure.
This is distinct from "Step in": Dismiss just removes the visual edge, Step in creates a new tab.

**Go/web status:** The ARG edge context menu is generated from `webui_session.go` actions for
edges. "Step in" / `decompose` is confirmed. "Dismiss" as a standalone remove-edge-from-view
action is not confirmed. Need to check if `dismiss` is a registered action in the ARG edge menu.

**Files to check/change:**
- `webui_session.go` — search for `"dismiss"` in ARG edge action cases
- If absent, add `case "dismiss"` that removes the edge from the ARG visualization without
  creating a new sheet

**Test:** Right-click an edge → Dismiss; edge disappears from ARG. ARG state is otherwise
unchanged.

---

### 1.7 Try Conjecture → Browse Source at Conjecture Definition Line

**Python behavior (PLAN378 §16.2):** After a conjecture is selected from the listbox, the
source browser opens pointing to the conjecture's `.ivy` source file and line number.

**Go/web status:** The `try_conjecture` action exists in `webui_session.go` (line 2180). However
the Go handler for "Try conjecture" may not return the conjecture's `lineno` for editor
navigation. Needs verification.

**Files to check/change:**
- `webui_session.go` — `case "try_conjecture"` handler; check if response includes `lineno`
- `frontend/src/services/argActionService.ts` — check if it calls `scrollEditorToLine` after
  conjecture selection

**Test:** Right-click ARG node → Try conjecture → select a conjecture; CodeMirror editor
scrolls to the conjecture's declaration line.

---

### 1.8 Subgraph Cluster Boxes (Rectangles Around Isolate Groups)

**Python behavior (PLAN378 §8.13 / §26.4):** When `subgraph_boxes=True`, DOT layout computes
bounding boxes around nodes in the same Graphviz cluster. These are rendered as rectangles
behind the nodes.

**Go/web status:** `webui_cyrender.go` renders `CyElements.add_shape` as rectangles (the shape
type exists in the data model). However, it is unclear whether the Go ARG rendering actually
sets `subgraph_boxes=True` and emits shape elements for isolate groups. No Cytoscape rectangle
elements for clusters were confirmed in the frontend inventory.

**Files to check/change:**
- `webui_cyrender.go` — check if `RenderARG` passes `subgraph_boxes=true` to layout
- If not, add cluster/isolate grouping and emit rectangular `shape` elements in the ARG JSON

**Test:** Open an Ivy file with multiple isolates. Verify rectangle outlines appear in the ARG
grouping nodes that belong to the same isolate.

---

## SECTION 2 — Likely Missing: Needs Investigation and Fix

### 2.1 Non-Inductive Conjecture Name Shown in Check-Induction Dialog

**Python behavior (PLAN378 §20.5 / PLAN383 §66):** When CTI is found, `ok_dialog` shows
`"An assertion failed..."` with the specific failing conjecture name. On success, shows
`"Inductive invariant found:"` with all conjectures joined as text.

**Go/web status:** Check result arrives as a `check_result` SSE event. It is unclear whether
the frontend opens a dialog with the specific conjecture text (not just pass/fail status).

**Files to investigate:**
- `frontend/src/services/ivyRuntime.ts` — search for `check_result` handler (around lines
  3382, 4193)
- `webui_session.go` — check `CheckInduction` return data format for conjecture text

**Expected fix:** The `check_result` response should include the failing conjecture text/name;
the frontend should display it in a modal dialog rather than only in the status bar.

---

### 2.2 CTI Two-State ARG Display After Induction Failure

**Python behavior (PLAN378 §20.6 / PLAN383 §66 §68):** After a CTI failure, `set_states(s0, s1)`
replaces the displayed ARG with a two-state pre/post view. `Diagram` then operates on this
two-state graph.

**Go/web status:** `GET /api/session/{id}/arg/cti` endpoint exists and is tested in
`webui_arg_step_in_test.go`. But it is unclear if the frontend **automatically switches the ARG
panel** to show the CTI ARG after a failed induction check event.

**Files to investigate:**
- `frontend/src/services/ivyRuntime.ts` — `check_result` handler: does it call `arg/cti`
  endpoint and refresh the ARG panel?
- `frontend/src/services/graphRuntime.ts` — does it fetch CTI ARG?

**Expected fix:** On `check_result` event with CTI, frontend should call `GET .../arg/cti` and
render that ARG, then auto-select the pre-state concept graph.

---

### 2.3 Event Viewer: `<<`/`>>` Directional Pattern Navigation Buttons

**Python behavior (PLAN378 §25.5 / PLAN383 §82):** The pattern list has `<<` (search backward
from current selection to previous match) and `>>` (search forward to next match) buttons.
Neither creates a new sheet; they navigate in the existing sheet.

**Go/web status:** `eventTraceService.ts` accepts a `reverse` parameter (line 126), suggesting
backend support. But no `<<`/`>>` button UI elements were confirmed in `static/index.html`.

**Files to investigate:**
- `static/index.html` — event trace panel markup
- `frontend/src/services/eventTraceService.ts` — lines 120–140

**Expected fix:** If `<<`/`>>` buttons are absent in the HTML, add them and wire to
`findEvent(pattern, reverse=true/false)` from the selected pattern.

---

### 2.4 Refine with Interpolant Dialog: PDR-Specific vs Induction-Specific Text

**Python behavior (PLAN378 §17.1–17.2 / PLAN383 §58):** When Reverse finds an infeasible
pre-image with an interpolant, the dialog says either `"PDR found the following invariant:"`
(PDR mode) or `"Found the following separating formula:"` (other modes). The button label is
`"Refine"` not `"OK"`. For a vacuous pre-state: `"The pre-state is vacuous..."`.

**Go/web status:** The Go/web UI handles the reverse step with interpolant via concept graph
Reverse action. Need to verify the specific dialog text and "Refine" button label are used rather
than a generic OK dialog.

**Files to investigate:**
- `frontend/src/services/conceptActionService.ts` — reverse action response handler
- `webui_session.go` — reverse action return structure

**Expected fix:** Ensure the dialog uses mode-specific text and "Refine" button label (not "OK").

---

### 2.5 Eliminated Conjectures Dialog from Reach (One-Step Reachability)

**Python behavior (PLAN378 §31.2):** When `one_step_reach` finds a reachable state that
violates a conjecture, a `listbox_dialog` is shown with the list of eliminated conjectures.
The user acknowledges them.

**Go/web status:** The "Reach" action in `webui_session.go` implements one-step reachability.
Need to verify if the response includes an eliminated-conjecture list and if the frontend shows
a dialog for it.

**Files to investigate:**
- `webui_session.go` — `case "reach"` / `one_step_reach` return value
- `frontend/src/services/conceptActionService.ts` or `analysisActionService.ts` — reach handler

**Expected fix:** If conjectures are eliminated, show a modal listing them before continuing.

---

### 2.6 Named Domain Snapshots: Save/Load Domain by Name

**Python behavior (PLAN378 §30.3):** `save_domain(name)` / `load_domain(name)` store and
restore named concept domain snapshots (separate from undo/redo stack). "Reset domain" restores
the `'initial'` snapshot (§33.13).

**Go/web status:** The `concept/reset` endpoint restores the initial domain. Undo/redo is
implemented. But named snapshots beyond `'initial'` may not exist. The "Remember" action stores
a *graph* copy but at the concept-session domain level, save/load-by-name is distinct.

**Files to investigate:**
- `webui_concept_domain.go` or `webui_session.go` — look for `save_domain`/`load_domain`

**Expected fix:** If absent, implement named domain snapshots distinct from undo/redo, accessible
via backend API.

---

## SECTION 3 — Missing Test Coverage (Feature Exists, Tests Absent)

These features are implemented in the backend but lack Playwright e2e tests or frontend unit
tests. Missing test coverage counts as "missing" per the audit criteria.

### 3.1 ARG Edge Right-Click: Step In (Decompose) — E2E Test
- Backend: `webui_arg_step_in_test.go` has Go-level test
- Missing: Playwright e2e in `pw_test/ivyweb_pilot.spec.mjs` for the full user flow
  (right-click edge → "Step in" → new tab appears with sub-ARG)

### 3.2 ARG Edge Right-Click: View Source (Editor Scroll) — E2E Test
- Backend: `scrollEditorToLine` in `argActionService.ts` confirmed
- Missing: Playwright test verifying CodeMirror editor scrolls to the correct line

### 3.3 Splatter Node Action — Unit + E2E Test
- Backend: `case "splatter"` (`webui_session.go` line 1887); frontend: `conceptActionService.ts`
  line 210
- Missing: unit test in `conceptActionService.test.ts` and Playwright e2e

### 3.4 Constraints Display: Clickable Toggle Below Concept Graph
- Backend: `detailsService.ts` `populateConstraintFacts` confirmed
- Missing: Playwright test — click a constraint text, verify it toggles grey and is removed from
  active facts; undo restores it

### 3.5 Try Remembered Goal from ARG Context Menu
- Backend: `case "try_remembered"` in `webui_session.go` (line 2180) confirmed
- Missing: integration test: (a) remember a goal, (b) ARG right-click → Try remembered goal,
  (c) verify concept graph loads the stored goal state

### 3.6 Cover Node: Dashed Cover Edge Appears After Success
- Backend: `case "cover"` in `webui_session.go` adds cover edge to ARG
- Missing: Playwright test verifying the dashed cover-edge CSS class appears in ARG

### 3.7 Join Node: New Join-Transition State and "join" Edge Labels
- Backend: `case "join"` in `webui_session.go`
- Missing: unit test verifying joined state has `transition_join`-class edges labeled "join"

### 3.8 Delete ARG Node: Renumbering After Deletion
- Backend: `case "delete"` in `webui_session.go` (line 2204) calls `ui.DeleteNode`
- Missing: test verifying remaining node IDs renumber and no transition references deleted state

### 3.9 Save Abstraction `.ivy` Format Round-Trip
- Backend: `saveInvariantContent()` in `webui_session.go` confirmed; format test exists
- Missing: e2e test: (a) add concept, (b) save abstraction, (c) reload file, verify concept
  syntax is valid Ivy

### 3.10 Recalculate Concept Graph from Parent ARG State
- Backend: `GraphWidget.Recalculate()` in `webui_graph_widget.go` (lines 377–393)
- Missing: unit test in TypeScript concept action layer confirming parent ARG state is re-queried

### 3.11 Extend (Find Non-Covered Extension): "State N is closed" Dialog
- Backend: `case "extend"` in `webui_session.go` (line 2094)
- Partial: `argActionService.test.ts` line 48 has basic coverage
- Missing: integration test for the "State N is closed." OK dialog path when node has no
  extensions

### 3.12 Recalculate All: Each Target Processed Once
- Backend: `ui.RecalculateAll()` in `webui_session.go`
- Missing: test verifying each unique target is recalculated exactly once when multiple
  transitions share the same target state

### 3.13 Bounded Check Failure: "View error trace?" OK/Cancel Dialog
- Python (PLAN378 §12.2): bounded safety failure shows OK/Cancel "View error trace?" dialog;
  Cancel suppresses new tab
- Go/web: the check result flow should show this dialog; missing explicit Playwright test that
  Cancel prevents the new tab

### 3.14 Try Conjecture: List Dialog Shows Unproven Conjectures
- Backend: `case "try_conjecture"` in `webui_session.go` confirmed
- Missing: unit test that the listbox dialog is populated from the unproven-conjecture list

### 3.15 ARG Node Color Persists Across Graph Rebuild
- Python (PLAN378 §11.2): mark (red) and safe (green) colors are restored after `rebuild()`
- Missing: test that after any graph-refreshing action, marked and safe node colors are
  re-applied

### 3.16 CTI Weaken: Conjectures Displayed Without Leading Universals
- Backend: `webui_session.go` uses `DropUniversals` confirmed
- Missing: test comparing displayed conjecture text to formula-minus-universals in Go unit test

### 3.17 Save Invariant Three-Section Format Round-Trip E2E
- Backend: `TestSaveInvariantUsesPythonKeptDroppedSections` Go unit test confirmed
- Missing: Playwright e2e: remove a conjecture, add one, save invariant, verify three sections
  in output file

### 3.18 Local Safety Check: Buttons Dialog with "View unsafe states" and "View error trace"
- Python (PLAN378 §12.4–12.5): local safety failure shows a `buttons_dialog_cancel` with two
  action buttons
- Go/web: need to verify both buttons are shown and functional; Playwright test missing

---

## SECTION 4 — Intentional Omissions (Not Required in Go/Web UI)

These Python behaviors are intentionally absent (Tk-specific, Jupyter-specific, or
replaced by web equivalents):

| Python Feature | Reason Omitted |
|---|---|
| `tkinter.tix.Tk()` root window, palette white | Replaced by browser window + CSS dark theme |
| Tix NoteBook, HList, TList widgets | Replaced by HTML tab bar and DOM lists |
| `DYLD_LIBRARY_PATH` setup in `ivy_shell.py` | Not relevant in browser context |
| `.a2g` pickle file loading | Intentionally unsupported; `python_a2g_equivalent: false` flag set in `webui_session.go` line 2422. Migration: user must re-run from source. |
| Separate floating `FileBrowser` window (§23) | Replaced by embedded CodeMirror editor with `scrollEditorToLine` |
| Pre-seeded dialog answers `TkUI.answer()` (§22.2) | Tk test hook; Playwright handles web UI testing |
| Busy cursor `watch`/`cursor=''` (§22.14) | Replaced by loading overlay spinner in Go/web UI |
| `ivy_check --diagnose=true` GUI entry (§27) | Python CLI entry; Go uses web session launch |
| `ivy_launch.py` xterm process launcher (§33.16) | Python distributed-process tool, not a web UI feature |
| `ivy2.py` Jupyter notebook entry (§28) | Jupyter-specific; web UI is the replacement |
| `ivy_show.py` compile helper (§95) | CLI only; no GUI component |
| Interactive UPDR modal `iupdr.py` (§88) | Very specialized notebook feature; no web equivalent yet |
| Abstractor selection dropdown (§33.5) | Jupyter `AnalysisSessionWidget`-specific; replaced by mode selector |
| BMC bound dropdown `[1,3,5,10,15]` (§33.6) | Jupyter `TransitionViewWidget`-specific; replaced by int dialog |
| `TransitionViewWidget` log file (§33.8) | Jupyter-only; no web equivalent |
| `apply_structure_renaming()` (§33.12) | Jupyter-only; no web equivalent |
| Analysis session history First/Prev/Next/Last (§21) | Jupyter history; no web equivalent; undo/redo is the web analog |
| Extension point API `arg_node_actions` (§29) | Server-provided actions replace this; no JS extension hook needed |
| Concept graph background `rgb(192,192,255)` (§18.5) | Design decision: Go/web uses dark theme |
| 27-color cycling palette for Tk lines (§32.18) | Tk rendering detail; web uses CSS color classes |
| `ivy_ui_none.py` empty compile mode (§96) | Compiler-only; no web surface |
| `ivy_launch.py` port allocation (§33.17) | Python distributed tool |
| `@interaction` decorator / ExecuteNewCell (§29.6) | Jupyter generator interaction pattern; not needed in Go/web |

---

## SECTION 5 — Summary Table

| # | Feature | Status | Priority |
|---|---|---|---|
| 1.1 | ARG node green border on safe state | MISSING | High |
| 1.2 | ARG node red fill when marked | MISSING | High |
| 1.3 | T checkbox: transitive reduction computed | MISSING | High |
| 1.4 | Bulk relation toggle by name/header | MISSING | High |
| 1.5 | Enter key confirms entry dialogs | MISSING | High |
| 1.6 | ARG edge "Dismiss" action | MISSING | Medium |
| 1.7 | Try Conjecture → browse source line | MISSING | Medium |
| 1.8 | Subgraph cluster boxes in ARG | MISSING | Low |
| 2.1 | Check induction: failing conjecture text in dialog | NEEDS VERIFY | High |
| 2.2 | CTI two-state ARG display auto-switch | NEEDS VERIFY | High |
| 2.3 | Event viewer `<<`/`>>` nav buttons | NEEDS VERIFY | Medium |
| 2.4 | Refine dialog: PDR vs induction text + "Refine" button | NEEDS VERIFY | Medium |
| 2.5 | Eliminated conjectures dialog from Reach | NEEDS VERIFY | Medium |
| 2.6 | Named domain snapshots (save/load by name) | NEEDS VERIFY | Low |
| 3.1 | Test: Step In (decompose) e2e | NO TEST | Medium |
| 3.2 | Test: View Source editor scroll e2e | NO TEST | Medium |
| 3.3 | Test: Splatter node unit+e2e | NO TEST | Medium |
| 3.4 | Test: Constraints clickable toggle e2e | NO TEST | Medium |
| 3.5 | Test: Try remembered goal flow e2e | NO TEST | Medium |
| 3.6 | Test: Cover node dashed edge e2e | NO TEST | Low |
| 3.7 | Test: Join node "join" edge labels | NO TEST | Low |
| 3.8 | Test: Delete ARG node renumbering | NO TEST | Low |
| 3.9 | Test: Save abstraction round-trip e2e | NO TEST | Low |
| 3.10 | Test: Recalculate concept from parent state | NO TEST | Low |
| 3.11 | Test: Extend "closed" dialog | NO TEST | Low |
| 3.12 | Test: Recalculate all (once per target) | NO TEST | Low |
| 3.13 | Test: Bounded fail OK/Cancel dialog | NO TEST | Medium |
| 3.14 | Test: Try conjecture listbox content | NO TEST | Low |
| 3.15 | Test: Node color persists after rebuild | NO TEST | Medium |
| 3.16 | Test: Weaken without universals display | NO TEST | Low |
| 3.17 | Test: Save invariant 3-section e2e | NO TEST | Low |
| 3.18 | Test: Local safety failure buttons dialog | NO TEST | Medium |

---

## Verification

After implementing each item:
- Run `cd ~/ivy/goivy && make test` to execute the full Go test suite
- Run Playwright tests in `pw_test/` for e2e UI verification
- For visual changes (green/red node colors, transitive reduction, cluster boxes), manually
  load a small `.ivy` file in the web UI and exercise the feature
