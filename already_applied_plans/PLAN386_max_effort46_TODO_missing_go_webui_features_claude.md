# PLAN386: Missing Go/Web UI Features — TODO List

## Context

This document is the result of a "set subtraction" audit: the union of Python/Tcl GUI behaviors
inventoried in PLAN378 and PLAN383 minus the behaviors already confirmed as implemented in the
Go/web UI (`~/ivy/goivy/webui/`). The Go/web UI inventory was produced by reading the full
frontend TypeScript source (`frontend/src/`), CSS (`static/css/ivy.css`), and Go backend files
(`webui_session.go`, `webui_cyrender.go`, `webui_cystyles.go`, `webui_handlers.go`, etc.).

A feature is counted as **missing** if it is absent, stubbed, or lacks test coverage in the
Go/web UI. Each item below is a concrete TODO with file locations.

---

## SECTION 1 — Confirmed Missing: Must Implement

### 1.1 ARG Node Green Highlight on Successful Safety Check

**Python behavior (PLAN383 §22 / PLAN378 §2.1):** After `check_safety_node` succeeds, the ARG
node's outline is filled **green** by `AnalysisGraphUI.update_node_color` (reads `node.safe`).
On failure the outline remains black.

**Go/web status:** `ARGStyle()` in `webui_cystyles.go` defines only `bottom_state` (black) and
default gray. There is no `safe` Cytoscape class or color rule. The backend does track safety
results (`CheckLocalSafety`, `CheckBoundedSafety` in `webui_session.go`) but never sends a
`safe` CSS class in the ARG element response.

**Files to change:**
- `webui_cystyles.go` — add `.safe` selector with green outline
- `webui_cyrender.go` — set `safe` class on ARG nodes whose state has `safe=true`
- `frontend/src/services/graphRuntime.ts` — apply class update when arg_updated event arrives

**Test:** After running a safety check that passes, the ARG node's border should be green.
After a failure, it should remain black/default.

---

### 1.2 ARG Node Red Fill for Marked State

**Python behavior (PLAN383 §26 / PLAN378 §4.1):** `mark_node` stores the marked node and calls
`show_mark(True)`, which fills the node's canvas shape **red**. Unmarking clears the fill.
The mark is consumed by "Cover by marked" and "Join with marked".

**Go/web status:** The `mark` action in `webui_session.go` (line 2130) calls `ui.MarkNode()`
on the backend, but `ARGStyle()` in `webui_cystyles.go` has no `marked` Cytoscape class or red
fill. The ARG node has no visual indicator that it is currently marked.

**Files to change:**
- `webui_cystyles.go` — add `.marked` selector with red background-color
- `webui_cyrender.go` or ARG serialization — emit `marked` class on the currently marked state
- `webui_session.go` — include `marked_state_id` in ARG JSON so the frontend can apply the class

**Test:** Mark node 0 (right-click → Mark), verify it fills red. Mark node 1, verify node 0
loses red fill and node 1 becomes red. Use Cover/Join with marked and verify they reference the
correct node.

---

### 1.3 T Checkbox: Actual Transitive Reduction (Not Just Visibility Toggle)

**Python behavior (PLAN383 §100 / PLAN378 §3.7):** When the T checkbox for a relation is on and
the relation has `all_to_all` facts, `get_transitive_reduction()` hides self-edges AND edges
implied by a two-step path (e.g., A→C is hidden if A→B and B→C exist). This is computed, not
merely CSS-toggled.

**Go/web status:** `conceptVisibilityService.ts` (line 81) registers a `transitive` checkbox
labelled "Transitive reduction". No call to any `get_transitive_reduction` function was found in
the Go or TypeScript source; the checkbox likely controls only display-level visibility without
removing implied edges from the rendered element set.

**Files to change:**
- `webui_cyrender.go` or a new helper — implement `getTransitiveReduction(elements, relName)`
  that removes self-edges and two-step-implied edges when T is active for a relation
- Wire result into concept graph element computation when `T` toggle is set

**Test:** Create a concept graph with relation R and facts A→B, B→C, A→C, A→A. Enable T for R.
Verify A→A and A→C are hidden; A→B and B→C remain visible.

---

### 1.4 Bulk Relation Toggle: Click Relation Name Toggles All Three Columns

**Python behavior (PLAN383 §85 / PLAN378 §4.5):** In `ConceptSessionControls`, clicking a
relation-name button toggles all three edge display checkboxes (`+`, `?`, `-`) for that relation
simultaneously. Clicking a column header button toggles that display class across **all**
relations.

**Go/web status:** `conceptVisibilityService.ts` (line 99–105) attaches a click listener on the
relation-name `<a>` element that calls only `event.preventDefault()`. No toggle logic is
implemented; the click is silently absorbed.

**Files to change:**
- `frontend/src/services/conceptVisibilityService.ts` — implement the click handler to toggle
  the `none_to_none`, `edge_unknown`, and `all_to_all` checkboxes for the clicked relation row
- Also implement the column-header button to toggle that column across all relations

**Test:** With multiple relations visible in the checkbox panel, click a relation name and
verify all three boxes for that relation flip state simultaneously. Click the `+` header and
verify all relations' positive boxes flip.

---

### 1.5 Entry Dialogs: Bind Enter/Return Key to Confirm Action

**Python behavior (PLAN383 §15 / PLAN378 §3.5):** `entry_dialog` focuses the entry field and
binds the Return key to run the action with the current entry contents. This applies to: Add
Relation, Remember goal, pattern search, and integer dialogs.

**Go/web status:** `ivyRuntime.ts` dialogs (lines 3935–3960, 3857–3867) only bind the Escape
key (line 3859). Input elements (lines 3940–3945, 3968–3975) have no `keydown`/`keypress`
handlers. Pressing Enter in any entry dialog does nothing.

**Files to change:**
- `frontend/src/services/ivyRuntime.ts` — in `_addDialogInput()` and `_addDialogIntInput()`,
  add `input.addEventListener('keydown', e => { if (e.key === 'Enter') confirmButton.click(); })`

**Test:** Open "Add Relation" dialog, type a relation name, press Enter, and verify the relation
is added (same as clicking "Add"). Open bounded check dialog, type a number, press Enter, verify
the check starts.

---

## SECTION 2 — Likely Missing: Needs Investigation and Fix

### 2.1 Non-Inductive Conjecture Name Shown in Check-Induction Dialog

**Python behavior (PLAN383 §66):** When `check_induction` finds a non-inductive conjecture, it
opens a **text dialog** displaying the specific failing conjecture's name/formula. On success it
shows `"Inductive invariant found:"` with the list of conjectures.

**Go/web status:** The check result reaches the frontend as a `check_result` SSE event. It is
unclear whether the frontend opens a dialog with the specific failing conjecture text, or just
updates a status bar. This needs code-reading verification.

**Files to investigate:**
- `frontend/src/services/ivyRuntime.ts` — search for `check_result` handler (~line 3382, 4193)
- `webui_session.go` — search `CheckInduction` return data format

**Expected behavior:** A modal/toast should display the conjecture name (not just "FAILED").
If the current display omits the conjecture name, add it to the response JSON and display it.

---

### 2.2 CTI Two-State ARG Display After Induction Failure

**Python behavior (PLAN383 §66, §68):** When check-induction finds a CTI, the main ARG display
is **replaced with a two-state graph** showing the pre-state and post-state of the
counterexample, and the pre-state's concept graph is automatically loaded. `Diagram` then uses
this two-state graph.

**Go/web status:** The `cti_arg` endpoint (`GET /api/session/{id}/arg/cti`) exists in
`webui_handlers.go`, and `webui_arg_step_in_test.go` tests it, but it is unclear if the
frontend automatically switches the ARG panel to the CTI ARG after a failed induction check.

**Files to investigate:**
- `frontend/src/services/ivyRuntime.ts` — how `check_result` handler updates the ARG
- `frontend/src/services/graphRuntime.ts` — does it call the CTI ARG endpoint?

**Expected behavior:** After a failing induction check, the ARG panel should show the pre/post
CTI states and auto-select the pre-state concept graph.

---

### 2.3 Event Viewer: `<<` / `>>` Directional Pattern Navigation Buttons

**Python behavior (PLAN383 §82 / PLAN378 §5.4):** The pattern list panel has `<<` (find reverse
from current selection) and `>>` (find forward) buttons. These navigate to the previous/next
event matching the selected pattern without creating a new filter sheet.

**Go/web status:** `eventTraceService.ts` has a `reverse` parameter in the search call (line
126), suggesting the backend supports it, but no `<<`/`>>` button UI elements were confirmed.
The `static/index.html` event trace panel markup needs to be inspected.

**Files to investigate:**
- `static/index.html` — event trace panel buttons
- `frontend/src/services/eventTraceService.ts` — lines 120–140

**Expected behavior:** A `<<` button triggers `findEvent(pattern, reverse=true)` from current
position; `>>` triggers `findEvent(pattern, reverse=false)`. The found event is selected and
scrolled into view. If buttons are absent, add them.

---

## SECTION 3 — Missing Test Coverage (Feature Exists, Tests Absent)

These features are implemented in the backend but lack Playwright e2e tests or frontend unit
tests. Missing test coverage counts as "missing" per the audit criteria.

### 3.1 ARG Edge Right-Click: Step In (Decompose Into New Tab)
- Backend: `webui_arg_step_in_test.go` has Go-level test `TestArgStepInClientServerDiagnosticEdge`
- Missing: Playwright e2e test in `pw_test/ivyweb_pilot.spec.mjs` for the full user flow
  (right-click edge → "Step in" → new tab appears with sub-ARG)

### 3.2 ARG Edge Right-Click: View Source (Scroll Editor to Line)
- Backend: confirmed via `scrollEditorToLine` in `argActionService.ts`
- Missing: Playwright test verifying the CodeMirror editor scrolls to the correct line

### 3.3 Splatter Node Action
- Backend: `case "splatter"` in `webui_session.go` (line 1887); frontend in
  `conceptActionService.ts` (line 210)
- Missing: unit test in `conceptActionService.test.ts` and Playwright e2e test

### 3.4 Constraints Display: Clickable Toggle Below Concept Graph
- Backend: `detailsService.ts` `populateConstraintFacts` creates buttons
- Missing: Playwright test verifying clicking a constraint toggles it grey and removes it from
  active facts; Playwright test for undo restoring constraint state

### 3.5 Try Remembered Goal from ARG Context Menu
- Backend: `webui_session.go` (line 2180) `case "try_remembered"` confirmed
- Missing: integration test that (a) Remembers a goal, (b) clicks ARG node → "Try remembered
  goal", (c) verifies concept graph switches to the stored goal's state

### 3.6 Cover Node: Visual Dashed Cover Edge After Success
- Backend: `case "cover"` in `webui_session.go` adds a cover edge to the ARG
- Missing: Playwright test verifying the dashed cover edge class appears in the ARG after Cover

### 3.7 Join Node: New Join-Transition State in ARG
- Backend: `case "join"` in `webui_session.go`
- Missing: unit test verifying the joined state's transitions are labeled "join"

### 3.8 Delete ARG Node: Renumbering After Deletion
- Backend: `case "delete"` in `webui_session.go` (line 2204) calls `ui.DeleteNode`
- Missing: test verifying remaining node IDs are renumbered and no transition points to deleted
  state

### 3.9 Save Abstraction `.ivy` Format Round-Trip
- Backend: `saveInvariantContent()` in `webui_session.go` confirmed; format test exists
- Missing: e2e test that (a) adds a concept, (b) saves abstraction, (c) reloads file, verifies
  concept syntax is valid Ivy

### 3.10 Recalculate Concept Graph from Parent ARG State
- Backend: `GraphWidget.Recalculate()` in `webui_graph_widget.go` (lines 377–393)
- Missing: unit test in the TypeScript concept action layer confirming the parent ARG state is
  re-queried

### 3.11 Extend (Find Non-Covered Extension) Action
- Backend: `case "extend"` in `webui_session.go` (line 2094)
- Tests: `argActionService.test.ts` line 48 has basic coverage
- Missing: integration test verifying "State N is closed." dialog when node has no extensions

### 3.12 Recalculate All ARG Transitions
- Backend: `ui.RecalculateAll()` in `webui_session.go`
- Missing: test verifying each unique target state is recalculated exactly once when multiple
  transitions share the same target

---

## SECTION 4 — Intentional Omissions (Not Required in Go/Web UI)

The following Python behaviors are intentionally absent because they are Tk-specific, Python-
specific, or replaced by equivalent web mechanisms:

| Python Feature | Reason Omitted |
|---|---|
| `tkinter.tix.Tk()` root window, palette white | Replaced by browser window + CSS dark theme |
| Tix NoteBook widget | Replaced by HTML tab bar |
| `DYLD_LIBRARY_PATH` setup in `ivy_shell.py` | Not relevant in browser context |
| `.a2g` pickle file loading | Intentionally unsupported; `python_a2g_equivalent: false` flag set in `webui_session.go` (line 2422) |
| `ivy_check diagnose=true` GUI entry | Python CLI entry point; Go uses web session |
| `ivy_launch.py` xterm process launcher | Python distributed-process tool; not a web UI feature |
| `ivy2.py` Jupyter notebook entry | Jupyter-specific; web UI is the replacement |
| `ivy_show.py` non-interactive compile helper | CLI-only; no GUI component |
| `TkUI.answer(string)` pre-seeded dialog answers | Tk test hook; Playwright handles web testing |
| Interactive UPDR (`iupdr.py`) modal | Very specialized notebook feature; no web equivalent yet |
| FileBrowser top-level reuse singleton | Web uses CodeMirror editor tab; no separate window |
| Tk tab-counter-only-increments quirk | Not replicated; web tab management is cleaner |
| `ivy_ui_none.py` empty compile mode | Compiler-only; no web surface |

---

## SECTION 5 — Summary Table

| # | Feature | Status | Priority |
|---|---|---|---|
| 1.1 | ARG node green on safe | MISSING | High |
| 1.2 | ARG node red on marked | MISSING | High |
| 1.3 | T checkbox: transitive reduction computation | MISSING | High |
| 1.4 | Bulk relation toggle by name/header | MISSING | High |
| 1.5 | Enter key in entry dialogs | MISSING | High |
| 2.1 | Non-inductive conjecture name in dialog | NEEDS VERIFY | Medium |
| 2.2 | CTI two-state ARG display | NEEDS VERIFY | Medium |
| 2.3 | Event viewer `<<`/`>>` nav buttons | NEEDS VERIFY | Medium |
| 3.1 | Test: Step In (decompose) e2e | NO TEST | Medium |
| 3.2 | Test: View Source editor scroll e2e | NO TEST | Medium |
| 3.3 | Test: Splatter node | NO TEST | Medium |
| 3.4 | Test: Constraints clickable toggle | NO TEST | Medium |
| 3.5 | Test: Try remembered goal flow | NO TEST | Medium |
| 3.6 | Test: Cover node dashed edge | NO TEST | Low |
| 3.7 | Test: Join node labels | NO TEST | Low |
| 3.8 | Test: Delete ARG node renumbering | NO TEST | Low |
| 3.9 | Test: Save abstraction round-trip | NO TEST | Low |
| 3.10 | Test: Recalculate concept from parent | NO TEST | Low |
| 3.11 | Test: Extend "State N is closed" dialog | NO TEST | Low |
| 3.12 | Test: Recalculate all (once per target) | NO TEST | Low |

---

## Verification

After implementing each item:
- Run `cd ~/ivy/goivy && make test` to execute the full Go test suite
- Run Playwright tests in `pw_test/` for e2e UI verification
- For visual changes (green/red node colors, transitive reduction), manually load a small `.ivy`
  file in the web UI and exercise the feature
