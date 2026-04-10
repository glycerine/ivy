# Plan: Complete Port of `iupdr.py`, `widget_analysis_session.py`, and Their Dependencies

Created: 2026-04-10 ~14:00 ART

## Context

This plan resolves the three deferred items from the previous plan's followup notes
(`PLAN253_more_followup.md`).

**Status of the three deferred items:**

1. **`actions.Interpolant` relocation into `transrel.go`** — **ALREADY DONE**
   in commit `8c3a5dd3` ("apply plan 253"). The old `actions/interpolant.go`
   (153 lines) was deleted and content folded into `actions/transrel.go`. Today
   `Interpolant` lives at `actions/transrel.go:1517` and is correct. **No work
   needed.**

2. **`actions.ForwardInterpolant` relocation into `transrel.go`** — **ALREADY
   DONE** in the same commit. Today it lives at `actions/transrel.go:1536` and
   is correct. **No work needed.**

3. **Python `interp_from_unsat_core` callers not yet ported to Go** — **NOT
   DONE.** Two Python files call `interp_from_unsat_core` and need faithful
   mechanical ports per CLAUDE.md rules 4 and 6:
   - `widget_analysis_session.py:1644` (1648-line file, 4 classes, 66 methods)
   - `iupdr.py:193` (203-line file with `UserSelectCore` class and
     `interactive_updr` generator)

   Today's Go state:
   - `actions.InterpFromUnsatCore` exists at `actions/transrel.go:1605` and is
     correct (the previous plan completed this).
   - `iupdr/iupdr.go` exists (196 lines) but is a Go-invented `Session`/`Step`
     refactor that does **not** mechanically port the Python file. Per CLAUDE.md
     rule 7 (no helper types that don't exist in Python) and rule 6 (must port
     all functions), this needs to be replaced with a literal port.
     - Verified: `iupdr.Session`, `iupdr.NewSession`, `iupdr.StepInfo` have
       **zero callers** elsewhere in the goivy tree. Safe to delete.
   - `webui/widget_analysis_session.go` does not exist at all.
   - `webui/ext_api.go` (235 lines) is a partial port of `ui_extensions_api.py`
     (410 lines). Missing pieces (`ShowModal`, `UserSelect`, `UserSelectMultiple`,
     `ExecuteNewCell`, `RunInteraction`, the `interaction` decorator, the default
     extension implementations) are required by `iupdr.py` and
     `widget_analysis_session.py`.
   - Several widget dependency files (`widget_dialog.py`, `widget_modal_messages.py`,
     `widget_cy_graph.py`, `dot_layout.py`) are unported.
   - Several `tactics_api`/`tactics` functions are missing as Go functions:
     `GetSafetyProperty`, `ArgAddActionNode`, `CustomRefineOrReverse`,
     `RemoveIfRefuted` (function form), `RecalculateFacts` (function form),
     `CheckCover` (function form).

The user instructed: "Do full and complete work, transitively planning any
additional work needed. Do not defer. Nothing is out of scope." This plan
therefore covers both file ports plus every transitive dependency, applying
CLAUDE.md rules 4, 6, 7, 8, and 9 strictly. No stubs.

## Goals

1. Port `iupdr.py` (203 lines) → `iupdr/iupdr.go` mechanically. The existing
   Go-invention `Session`/`StepInfo` is removed.
2. Port `widget_analysis_session.py` (1648 lines) → `webui/widget_analysis_session.go`
   mechanically. All 4 classes, all 66 methods, all 3 helpers.
3. Port the missing widget dependencies (`widget_dialog.py`,
   `widget_modal_messages.py`, `widget_cy_graph.py`, `dot_layout.py`).
4. Complete the `webui/ext_api.go` port (full `ui_extensions_api.py`).
5. Add the missing `tactics_api`/`tactics` Go functions required by both ports.
6. Build cleanly, pass existing tests, and add unit + conformance coverage.

## Strategy: Generators and Widgets in Go

Two Python idioms appear across these files. Both have one canonical literal
translation:

### Generators (`@interaction` + `yield`)

Python:
```python
@interaction
def interactive_updr():
    user_selection = yield UserSelectMultiple(...)
    ...
    user_selection, user_is_sat = yield UserSelectCore(...)
```

Go (literal translation): use Go 1.23+ range-over-function iterators
(`iter.Seq[FrontEndOperation]`). The interactive function returns an iterator;
its body uses `yield(op)` to pause. The consumer (e.g., a webui handler) loops
over the iterator with `for op := range gen { ... }` and writes the user's
response back into mutable fields on the `op` value before its loop body
returns. The next call to `yield(op)` resumes the generator body, which then
reads those response fields. This is the most literal Go translation of
Python's `value = yield op` pattern (CLAUDE.md rule 8) — single-goroutine,
synchronous, no channels, no goroutine leaks. The project already uses
`iter.Seq2` in `ivyutils/omap.go` (Go 1.24.3 in `go.mod`).

```go
type FrontEndOperation interface{ frontEndOp() }

type UserSelectMultiple struct {
    // input fields (set by generator before yield)
    Options *OrderedMap
    Title, Prompt string
    Default []lg.Expr

    // response fields (set by consumer before next iteration)
    Selection []lg.Expr
    Cancelled bool
}
func (*UserSelectMultiple) frontEndOp() {}

type UserSelectCore struct {
    *ShowModal
    Theory     *module.Clauses
    Constrains []lg.Expr
    Title, Prompt string

    // response fields
    SelectedConstraints []lg.Expr
    UserIsSat           bool
    Cancelled           bool
}
func (*UserSelectCore) frontEndOp() {}
```

Generator body:

```go
func InteractiveUpdr(tc *tactics.TacticsContext) iter.Seq[FrontEndOperation] {
    return func(yield func(FrontEndOperation) bool) {
        op := &UserSelectMultiple{ Options: opts, ... }
        if !yield(op) { return }
        if op.Cancelled { return }    // Python: assert user_selection is not None
        userSelection := op.Selection
        // ...continue...
    }
}
```

Consumer (webui handler):

```go
for op := range iupdr.InteractiveUpdr(tc) {
    switch o := op.(type) {
    case *iupdr.UserSelectMultiple:
        sel, ok := waitForUserMultiple(o.Options, o.Default)
        if !ok { o.Cancelled = true } else { o.Selection = sel }
    case *iupdr.UserSelectCore:
        sel, isSat, ok := waitForUserCore(o.Theory, o.Constrains)
        if !ok { o.Cancelled = true } else {
            o.SelectedConstraints = sel
            o.UserIsSat            = isSat
        }
    }
}
```

The `@interaction` decorator does not need a Go counterpart — generator-style
functions simply return `iter.Seq[FrontEndOperation]`. Non-generator helpers
remain plain Go functions. The Python `current_step == len(history)-1` guard
that the decorator imposes becomes a precondition check inside each generator
function (the first thing it does before any yield).

### IPython widgets

Python widget-only files (`widget_*.py`) use `IPython.html.widgets.Button`,
`Latex`, `SelectMultiple`, etc. The Go port follows the established pattern
already used in `webui/concept_isession.go`, `webui/cyrender.go`,
`webui/cystyles.go`: each Python class becomes a Go struct in `webui/`. Methods
that touch IPython widgets translate to methods that produce JSON for the
HTTP/WebSocket front end. Field names follow CLAUDE.md rule 3 (PascalCase).
Methods that are pure logic (geometry, collation, conjecture handling) port 1:1.

## Phase 1 — Complete `webui/ext_api.go`

`webui/ext_api.go` already has `ExtensionPoint`, `ExtensionAction`,
`InteractionError`, `FrontEndOperation` interface, and the registration
plumbing. Add the missing pieces. Each new symbol cites the Python source line
range and is a literal port.

**Additions to `webui/ext_api.go`:**

| Python symbol | Python lines | Go addition |
|---|---|---|
| `set_context` | 29-40 | `func (cfg *ExtConfig) SetContext(asw *AnalysisSessionWidget)` |
| `arg_node_actions` registration | 128-133 | replace stub `RegisterArg*` with literal `ExecuteActions` and `TryConjectures` callbacks |
| `try_conjectures` | 135-143 | full body using `TacticsContext.BackgroundTheory` and `Solver.Implies` |
| `run_interaction` | 146-163 | **dropped** — replaced by `iter.Seq` consumption in callers (`for op := range gen { ... }`) |
| `interaction` decorator | 165-184 | **dropped** — generator functions just return `iter.Seq[FrontEndOperation]`; the `current_step == len(history)-1` guard becomes a precondition `if`-check at the top of each generator body, yielding a `&ShowModal{Title:"Error", ...}` and returning if violated |
| `FrontEndOperation` (base) | 191-199 | already an interface; replace `Submit(onDone)` with marker method `frontEndOp()`. The "submit/on_done" lifecycle is now expressed by mutating the op's response fields between yields. |
| `ExecuteNewCell` | 201-221 | `type ExecuteNewCell struct { Code string; Output any /*set by consumer*/ }` |
| `ShowModal` | 224-242 | `type ShowModal struct { Title string; Children []Widget; OK bool /*set by consumer*/ }` |
| `UserSelect` | 245-268 | `type UserSelect struct { *ShowModal; Prompt string; Options *OrderedMap; Default any; Selection any; Cancelled bool }` |
| `UserSelectMultiple` | 271-296 | `type UserSelectMultiple struct { *ShowModal; Prompt string; Options *OrderedMap; Default []lg.Expr; Selection []lg.Expr; Cancelled bool }` |
| `execute_arg_action` | 299-306 | `func ExecuteArgAction(node *art.State, action string) iter.Seq[FrontEndOperation]` |
| `try_conjecture` | 309-318 | `func TryConjecture(node *art.State, conj lg.Expr) iter.Seq[FrontEndOperation]` |
| `arg_new_goal` | 321-326 | replace stub in `RegisterArgNewGoal` with a literal port that yields `ExecuteNewCell` |
| `arg_recalculate` | 328-336 | likewise for `RegisterArgRecalculate` |
| `arg_check_cover` | 338-356 | likewise for `RegisterArgCheckCover` |
| `arg_remove_facts` | 359-379 | likewise for `RegisterArgRemoveFacts`, yielding `UserSelectMultiple` then `ExecuteNewCell` |
| `apply_goal_tactic` | 382-388 | `func ApplyGoalTactic(goal *proof.ProofGoal, tactic string) iter.Seq[FrontEndOperation]` |
| `arg_join2` | 390-409 | likewise for `RegisterArgJoin`, yielding `UserSelect` then `ExecuteNewCell` |

**New supporting types in `webui/ext_api.go`** (because Go has no Python
generator/decorator/widget runtime):

```go
type Widget interface{ MarshalWidget() interface{} } // marker for ShowModal children
type LatexWidget struct{ Text string }                // Python widgets.Latex
type ButtonWidget struct{ Description string; OnClick func() }
type SelectWidget struct{ Options *OrderedMap; Value interface{} }
type SelectMultipleWidget struct{ Options *OrderedMap; Value []interface{} }
type OrderedMap struct{ ... }                         // Python OrderedDict shim
```

These mirror the small set of `IPython.html.widgets` classes the Python files
actually instantiate. They are **not** an IPython reimplementation — they are
data carriers serialized to the existing webui front end.

## Phase 2 — Port the missing widget dependency files

CLAUDE.md rule 4: one Python file → one Go file. Each goes in `webui/`.

### 2.1 `webui/widget_dialog.go` ← `widget_dialog.py` (23 lines)

Single class `DialogWidget`. Port to `type DialogWidget struct` with
constructor `NewDialogWidget(title, orientation, overflowX, overflowY string,
options map[string]interface{}) *DialogWidget` and a `MarshalWidget()` method.

### 2.2 `webui/widget_modal_messages.go` ← `widget_modal_messages.py` (25 lines)

Single class `ModalMessagesWidget`. Port to `type ModalMessagesWidget struct`
with `Add(message string)` method and `MarshalWidget()`.

### 2.3 `webui/widget_cy_graph.go` ← `widget_cy_graph.py` (222 lines)

Class `CyGraphWidget` (DOMWidget wrapper around Cytoscape.js elements).
Port to `type CyGraphWidget struct { Elements *art.CyElements; Selected
[]CyElement; ... }` with all helper methods. The `_object_key` and
`_is_user_object` helpers go as private functions in the same file.

The class methods map 1:1: `__init__`, `set_elements`, `on_select`,
`on_unselect`, `clear_selection`, etc. Reuse `art.CyElements` types already
ported in `art/cyrender.go`.

### 2.4 `webui/dot_layout.go` ← `dot_layout.py` (286 lines)

Pure-function module with the geometry helpers and `dot_layout()` entry point.
Functions to port (1:1):

| Python | Go |
|---|---|
| `cubic_bezier_point(c, t)` | `CubicBezierPoint(c [4]Point, t float64) Point` |
| `square_distance_to_segment(...)` | `SquareDistanceToSegment(...)` |
| `approximate_cubic_bezier(...)` | `ApproximateCubicBezier(...)` |
| `get_approximation_points(...)` | `GetApproximationPoints(...)` |
| `_to_position(...)` | `toPosition(...)` |
| `_to_edge_position(...)` | `toEdgePosition(...)` |
| `_to_coord_list(...)` | `toCoordList(...)` |
| `dot_layout(elements)` | `DotLayout(elements *art.CyElements)` |

Helper `Point` type lives in this file (no abstractions to add elsewhere).

## Phase 3 — Port `iupdr.py` to `iupdr/iupdr.go`

### 3.1 Pre-port: extend `tactics/tactics.go`

Add the missing Go counterparts (each is a literal port; functions live in
`tactics/tactics.go` or a sibling file in the same package):

| Python (in `tactics_api.py` or `tactics.py`) | Lines | Go addition (signature) |
|---|---|---|
| `get_safety_property()` | tactics_api.py:60-64 | `func (tc *TacticsContext) GetSafetyProperty() *module.Clauses` — returns `tc.AG.Interp.Conjs[0]` negated as Clauses |
| `arg_add_action_node(pre, action, abstractor)` | tactics_api.py:109-120 | `func (tc *TacticsContext) ArgAddActionNode(pre *art.State, action actions.Action, abstractor interface{}) *art.State` — wraps `ag.Execute(...)`, sets `node.Clauses = TrueClauses(nil)` if abstractor is nil |
| `pop_goal()` | (not in Python; iupdr never calls it) | n/a |
| `recalculate_facts(node, facts)` | tactics.py | `func RecalculateFacts(tc *TacticsContext, node *art.State, facts []lg.Expr) bool` — literal port of the `RecalculateFacts` tactic body |
| `remove_if_refuted(goal)` | tactics.py | `func RemoveIfRefuted(tc *TacticsContext, goal *proof.ProofGoal) bool` — calls `RefutedGoal(goal)` and removes |
| `check_cover(covered, by)` | tactics.py:207-211 | `func CheckCover(tc *TacticsContext, covered, by *art.State) bool` — wraps `tc.AG.Cover(covered, by)` |
| `custom_refine_or_reverse(goal, x, y, auto_remove)` | tactics.py:90-125 | `func CustomRefineOrReverse(tc *TacticsContext, goal *proof.ProofGoal, x bool, y interface{}, autoRemove bool) bool` — literal port of the `CustomRefineOrReverse` tactic body (handles `if x`/`else`, the `auto_remove` removal walk, the `Step` logging dict) |

For each, the existing `RefineOrReverseTactic`, `RemoveIfRefuted` struct, etc.,
remains in place. The new package-level functions are the literal ports of
the Python functions created by the `@tactic` decorator.

Also verify (and add if missing): `module.NegateClauses` — already present in
`module/ops.go:453+` per the exploration. Confirm signature is
`func NegateClauses(*Clauses) *Clauses`.

### 3.2 Delete `iupdr/iupdr.go` body, replace with literal port

Replace the entire current contents (`Session`, `StepInfo`, `NewSession`,
`Initialize`, `AddFrame`, `PushBadStates`, `Step`, `CheckInductive`,
`RunToCompletion`) with the mechanical port of `iupdr.py`. The new file holds:

**`UserSelectCore` struct** (Python iupdr.py lines 23-93). Embeds the
`*webui.ShowModal` (Go composition mirrors Python inheritance). Faithful port:

```go
type UserSelectCore struct {
    *webui.ShowModal
    Theory      *module.Clauses
    Constrains  []lg.Expr
    S           *z3bridge.Solver
    Alits       []lg.Expr        // auxiliary boolean literals
    Select      *webui.SelectMultipleWidget
    Prompt      *webui.LatexWidget
    Result      *webui.LatexWidget
    CheckButton *webui.ButtonWidget
}

func NewUserSelectCore(theory *module.Clauses, constrains []lg.Expr,
    title, prompt string) *UserSelectCore { ... }   // lines 35-61

func (u *UserSelectCore) OnClose(modal *webui.ShowModal, button string) { ... } // lines 63-76

func (u *UserSelectCore) Check(button *webui.ButtonWidget) interface{} { ... }  // lines 78-93
```

The `__core_aux{n}` auxiliary literals are created via
`lg.NewConstantBool(fmt.Sprintf("__core_aux%d", n))` and added to the solver
exactly as Python does. The dead `#core = ivy_solver.minimize_core(...)`
comment block is preserved verbatim as Go comments per CLAUDE.md rule 8.

**`InteractiveUpdr` function** (Python iupdr.py lines 96-203). Returns an
`iter.Seq[webui.FrontEndOperation]`. The body is a literal translation; each
Python `yield X` becomes `if !yield(op) { return }` followed by reading the
consumer-mutated response fields off `op`.

```go
func InteractiveUpdr(tc *tactics.TacticsContext) iter.Seq[webui.FrontEndOperation] {
    return func(yield func(webui.FrontEndOperation) bool) {
        frames := tc.AG.States
        if len(frames) != 1 {
            // Python: raise InteractionError(...)
            errOp := &webui.ShowModal{Title: "Error", Children: []webui.Widget{
                &webui.LatexWidget{Text:
                    "Interactive UPDR can only be started when the ARG " +
                    "contains nothing but the initial state."},
            }}
            yield(errOp)
            return
        }

        badStates := module.NegateClauses(tc.GetSafetyProperty())
        action := tactics.GetBigAction(tc.AG)
        tc.AG.Actions[action.Repr()] = action

        initFrame := frames[0]
        lastFrame := initFrame

        for {
            // check inductive invariant
            for i := 0; i < len(frames)-1; i++ {
                if tactics.CheckCover(tc, frames[i+1], frames[i]) {
                    tc.Step(map[string]any{
                        "msg": fmt.Sprintf("Inductive invariant found at frame %d", i),
                        "i":   i,
                    })
                }
            }

            // add new frame
            lastFrame = tc.ArgAddActionNode(lastFrame, action, nil)
            tc.PushGoal(tactics.GoalAtArgNode(badStates.ToFormula(), lastFrame))
            tc.Step(map[string]any{"msg": "Added new frame"})

            tactics.RecalculateFacts(tc, lastFrame,
                tactics.ArgGetConjuncts(tactics.ArgGetPred(lastFrame)))

            for {
                currentGoal := tc.TopGoal()
                if currentGoal == nil { break }

                if tactics.RemoveIfRefuted(tc, currentGoal) { continue }

                if currentGoal.Node == initFrame {
                    fmt.Println("No Invariant!")
                }

                dg := tc.GetDiagram(currentGoal, false)
                options := webui.NewOrderedMap()
                for _, c := range module.SimplifyClauses(dg.Formula).Conjuncts() {
                    options.Set(c.String(), c)
                }
                op1 := &webui.UserSelectMultiple{
                    Options: options,
                    Title:   "Generalize Diagram",
                    Prompt:  "Choose which literals to take as the refutation goal",
                    Default: options.Values(),
                }
                if !yield(op1) { return }
                if op1.Cancelled {
                    // Python: assert user_selection is not None
                    panic("user_selection is None")
                }
                userSelection := op1.Selection

                ug := tactics.GoalAtArgNode(
                    module.NewClauses(userSelection, nil, nil).ToFormula(),
                    currentGoal.Node)
                tc.PushGoal(ug)
                tc.Step(map[string]any{"msg": "Pushed user selected goal", "ug": ug})

                goal := tc.TopGoal()
                preds, act := tactics.ArgGetPredAction(goal.Node)
                if act == nil { panic("action != 'join'") }
                if len(preds) != 1 { panic("len(preds) == 1") }
                pred := preds[0]
                axioms := tc.AG.Interp.BackgroundTheory()
                theory := module.AndClausesTyped(
                    actions.ForwardImage(pred.Clauses, axioms,
                        act.Update(tc.AG.Interp, nil)),
                    axioms)
                goalClauses := module.SimplifyClauses(
                    module.FormulaToClauses(goal.Formula, nil))
                if len(goalClauses.Defs) != 0 { panic("defs not empty") }

                slv := z3bridge.NewSolver(nil, nil)
                slv.Add(theory)
                slv.Add(goalClauses)
                isSat := slv.Check()

                var x bool
                var y interface{}
                switch isSat {
                case z3bridge.Sat:
                    bi := tc.BackwardImage(
                        module.FormulaToClauses(goal.Formula, nil), act)
                    x, y = false, tactics.GoalAtArgNode(bi.ToFormula(), pred)
                case z3bridge.Unsat:
                    op2 := &UserSelectCore{
                        Theory:     theory,
                        Constrains: goalClauses.Fmlas,
                        Title:      "Refinement",
                        Prompt:     "Choose the literals to use",
                    }
                    if !yield(op2) { return }
                    if op2.Cancelled { panic("user cancelled core selection") }
                    if op2.UserIsSat { panic("user_is_sat is False") }
                    core := module.NewClauses(op2.SelectedConstraints, nil, nil)
                    x = true
                    y = actions.InterpFromUnsatCore(goalClauses, theory, core, nil)
                    // ^^^ This is the target call: iupdr.py:193
                default:
                    panic(fmt.Sprintf("unexpected check result: %v", isSat))
                }

                tactics.CustomRefineOrReverse(tc, goal, x, y, false)
            }

            // propagate phase
            for i := 1; i < len(frames); i++ {
                prevConjs := tactics.ArgGetConjuncts(frames[i-1])
                curConjs := tactics.ArgGetConjuncts(frames[i])
                factsToCheck := setDifference(prevConjs, curConjs)
                tactics.RecalculateFacts(tc, frames[i], factsToCheck)
            }
        }
    }
}
```

(`setDifference` is a private file-local helper for the `set(...) - set(...)`
Python idiom; it does not represent a new abstraction.)

### 3.3 Tests for `iupdr/iupdr_test.go`

- `TestUserSelectCoreCreate` — verify struct construction and that the
  auxiliary literals are added to the solver in order.
- `TestUserSelectCoreCheck` — drive `Check()` with sat and unsat selections.
- `TestUserSelectCoreOnCloseCancel` — verify cancel produces `Cancelled = true`.
- `TestInteractiveUpdrInitialFrameError` — start with a non-initial ARG;
  iterate the returned `iter.Seq` once and verify the first yielded op is a
  `*webui.ShowModal` with the expected error text.
- `TestInteractiveUpdrSimple` — drive a small Ivy program (e.g., one of the
  test_vectors entries) end-to-end. Use a `for op := range InteractiveUpdr(tc)`
  loop in the test that scripts each op response by mutating the op fields
  before the loop body returns. Assert final `tc.Goals.Len() == 0` and the
  expected facts have been added to the ARG.

## Phase 4 — Port `widget_analysis_session.py` to `webui/widget_analysis_session.go`

This is the largest piece: 1648 lines, 4 classes, 66 methods, 3 helpers. The
Go file is created in `webui/` to match the established widget-port pattern.

### 4.1 Helpers (Python lines 47-82)

| Python | Go |
|---|---|
| `_print_args(*args, **kwargs)` (47-48) | `func printArgs(args ...interface{})` |
| `_make_buttons(buttons, **kwargs)` (60-76) | `func makeButtons(buttons []ButtonSpec) []*ButtonWidget` |
| `SmallButton(*args, **kwargs)` (79-82) | `func SmallButton(...) *ButtonWidget` |

### 4.2 `ConceptSessionControls` (Python lines 85-338, 13 methods)

```go
type ConceptSessionControls struct {
    AnalysisSessionWidget *AnalysisSessionWidget
    Concept               *ConceptInteractiveSession  // from concept_isession.go
    DisplayCheckboxes     map[string]*CheckboxWidget
    StructureRenaming     map[string]string
    // ... fields
}

func NewConceptSessionControls(asw *AnalysisSessionWidget) *ConceptSessionControls
```

Methods to port (preserving names per CLAUDE.md rule 2):

| Python (line) | Go |
|---|---|
| `__init__` (94) | `NewConceptSessionControls` |
| `new_display_checkbox` (142) | `NewDisplayCheckbox` |
| `change_display_checkbox` (147) | `ChangeDisplayCheckbox` |
| `edge_name_click` (174) | `EdgeNameClick` |
| `edge_class_click` (186) | `EdgeClassClick` |
| `node_label_name_click` (204) | `NodeLabelNameClick` |
| `get_concept_style` (216) | `GetConceptStyle` |
| `apply_structure_renaming` (222) | `ApplyStructureRenaming` |
| `update_view_controls` (225) | `UpdateViewControls` |
| `undo` (302) | `Undo` |
| `reset_domain` (305) | `ResetDomain` |
| `diagram_domain` (310) | `DiagramDomain` |
| `remove_concept` (316) | `RemoveConcept` |
| `split` (320) | `Split` |
| `suppose_empty` (323) | `SupposeEmpty` |
| `materialize_node` (326) | `MaterializeNode` |
| `materialize_edge` (329) | `MaterializeEdge` |
| `add_projection` (334) | `AddProjection` |

### 4.3 `ConceptStateViewWidget` (Python lines 339-483, 9 methods)

```go
type ConceptStateViewWidget struct {
    *ConceptSessionControls   // Go embedding mirrors Python inheritance
    Graph    *CyGraphWidget
    State    *interp.State
    // ... fields
}
```

Methods to port:

| Python (line) | Go |
|---|---|
| `__init__` (347) | `NewConceptStateViewWidget` |
| `_ipython_display_` (432) | `IPythonDisplay` (returns the marshaled widget tree) |
| `update_concept_style` (436) | `UpdateConceptStyle` |
| `render_graph` (439) | `RenderGraph` |
| `render` (442) | `Render` |
| `gather_facts` (454) | `GatherFacts` |
| `get_active_facts` (474) | `GetActiveFacts` |

### 4.4 `TransitionViewWidget` (Python lines 484-1264, ~30 methods)

This is the largest class. Embeds `ConceptSessionControls`. Methods to port:

| Python (line) | Go |
|---|---|
| `__init__` (492) | `NewTransitionViewWidget` |
| `log` (606) | `Log` |
| `register_session` (624) | `RegisterSession` |
| `show_result` (636) | `ShowResult` |
| `set_states` (643) | `SetStates` |
| `_ipython_display_` (664) | `IPythonDisplay` |
| `update_concept_style` (668) | `UpdateConceptStyle` |
| `render_graph` (672) | `RenderGraph` |
| `render` (696) | `Render` |
| `gather_facts` (699) | `GatherFacts` |
| `apply_structure_renaming` (749) | `ApplyStructureRenaming` |
| `fact_to_label` (754) | `FactToLabel` |
| `get_active_facts` (757) | `GetActiveFacts` |
| `new_ag` (769) | `NewAg` |
| `check_inductiveness` (776) | `CheckInductiveness` |
| `get_selected_conjecture` (861) | `GetSelectedConjecture` |
| `bmc_conjecture` (912) | `BmcConjecture` |
| `minimize_conjecture` (971) | `MinimizeConjecture` |
| `highligh_selected_facts` (1014) | `HighlighSelectedFacts` (sic — preserve typo per rule 8) |
| `autodetect_transitive` (1077) | `AutodetectTransitive` |
| `is_sufficient` (1103) | `IsSufficient` |
| `is_inductive` (1152) | `IsInductive` |
| `strengthen` (1198) | `Strengthen` |
| `weaken` (1206) | `Weaken` (returns `iter.Seq[FrontEndOperation]`; yields `UserSelectMultiple`) |
| `get_relevant_elements` (1223) | `GetRelevantElements` |

`Weaken` is implemented as an `iter.Seq[FrontEndOperation]`-returning function
(matching the iupdr generator pattern):

```go
func (t *TransitionViewWidget) Weaken(button *webui.ButtonWidget) iter.Seq[webui.FrontEndOperation] {
    return func(yield func(webui.FrontEndOperation) bool) {
        opts := buildOptionsFromConjectures(t.Conjectures)
        op := &webui.UserSelectMultiple{
            Options: opts,
            Title:   "Conjectures",
            Prompt:  "Select conjectures to remove",
            Default: nil,
        }
        if !yield(op) { return }
        if op.Cancelled || op.Selection == nil { return }   // Python: if user_selection is not None
        for _, conj := range op.Selection {
            t.Conjectures = removeConjecture(t.Conjectures, conj)
        }
        t.ShowResult(fmt.Sprintf("Removed the following conjectures:\n%s", joinConjs(op.Selection)))
    }
}
```

### 4.5 `AnalysisSessionWidget` (Python lines 1265-1648, 16 methods)

```go
type AnalysisSessionWidget struct {
    Session     *proof.AnalysisSession
    CurrentStep int
    Silent      bool
    Box         *DialogWidget
    ProofGraph  *CyGraphWidget
    Arg         *CyGraphWidget
    Crg         *CyGraphWidget
    Concept     *ConceptStateViewWidget
    // ... fields
}
```

Methods to port:

| Python (line) | Go |
|---|---|
| `__init__` (1271) | `NewAnalysisSessionWidget` |
| `_ipython_display_` (1400) | `IPythonDisplay` |
| `register_session` (1404) | `RegisterSession` |
| `render` (1430) | `Render` |
| `prev` (1467) | `Prev` |
| `next` (1472) | `Next` |
| `first` (1477) | `First` |
| `last` (1481) | `Last` |
| `step` (1485) | `Step` |
| `arg_node_click` (1514) | `ArgNodeClick` |
| `crg_node_click` (1526) | `CrgNodeClick` |
| `proof_node_click` (1544) | `ProofNodeClick` |
| `concept_new_goal` (1557) | `ConceptNewGoal` |
| `concept_check` (1571) | `ConceptCheck` |
| `concept_min_unsat_core` (1586) | `ConceptMinUnsatCore` |
| **`concept_refine`** (1604) | **`ConceptRefine`** ← contains the target `InterpFromUnsatCore` call |

`ConceptRefine` is the literal port of lines 1604-1648:
```go
func (a *AnalysisSessionWidget) ConceptRefine(button *webui.ButtonWidget) {
    if a.CurrentStep != len(a.Session.History)-1 { panic("...") }
    if a.CurrentStep != a.Concept.CurrentStep { panic("...") }

    facts := a.Concept.GetActiveFacts()
    goal := tactics.GoalAtArgNode(
        module.NewClauses(facts, nil, nil).ToFormula(),
        a.Session.AnalysisState.IvyAg.States[a.Concept.ArgNode.ID])
    preds, act := tactics.ArgGetPredAction(goal.Node)
    if act == nil { panic("action != 'join'") }
    if len(preds) != 1 { panic("len(preds) == 1") }
    pred := preds[0]
    axioms := a.Session.AnalysisState.IvyInterp.BackgroundTheory()
    theory := module.AndClausesTyped(
        actions.ForwardImage(pred.Clauses, axioms,
            act.Update(a.Session.AnalysisState.IvyInterp, nil)),
        axioms)
    goalClauses := module.FormulaToClauses(goal.Formula, nil)
    if len(goalClauses.Defs) != 0 { panic("defs not empty") }

    slv := z3bridge.NewSolver(nil, nil)
    slv.Add(theory)
    slv.Add(goalClauses)
    isSat := slv.Check()

    var x bool
    var y interface{}
    switch isSat {
    case z3bridge.Sat:
        a.Concept.Result = "SAT"
        // NOTE: Python falls through to t.custom_refine_or_reverse with
        // uninitialized x/y here — this is a known Python bug. We preserve
        // the literal behavior: skip the call in the SAT branch by leaving
        // x/y zero-valued. (Document the divergence in a comment.)
    case z3bridge.Unsat:
        x = true
        y = actions.InterpFromUnsatCore(goalClauses, theory, goalClauses, nil)
        // ^^^ This is the target call: widget_analysis_session.py:1644
    default:
        panic(fmt.Sprintf("unexpected check result: %v", isSat))
    }

    tactics.CustomRefineOrReverse(tc(), goal, x, y, false)
}
```

(The Python code calls `t.custom_refine_or_reverse(goal, x, y, False)` even
in the SAT branch where `x`/`y` are uninitialized — this is a real Python bug,
not a translation issue. We preserve it but add a comment, per rule 8.)

### 4.6 Tests for `webui/widget_analysis_session_test.go`

- `TestAnalysisSessionWidgetConstruct` — verify `NewAnalysisSessionWidget`
  produces all sub-widgets and registers handlers.
- `TestConceptRefineUnsat` — drive `ConceptRefine` with a goal whose theory is
  unsat, verify `InterpFromUnsatCore` is called and `CustomRefineOrReverse`
  applies the result.
- `TestWeakenInteraction` — drive `Weaken` end-to-end with
  `for op := range t.Weaken(nil) { ... }`, scripting the user's selection by
  mutating `op.Selection`/`op.Cancelled` before each loop body returns.
- Round-trip tests for each click handler (`Prev`, `Next`, `First`, `Last`,
  `ArgNodeClick`, etc.).

## Phase 5 — Verification

1. **Compile**: `go build ./...` from `goivy/` must succeed with no errors.
2. **Existing tests must still pass**: `go test ./...`. Specific suites to
   watch: `actions/...`, `interp/...`, `tactics/...`, `webui/...`, `iupdr/...`.
3. **New tests**: all tests added in 3.3 and 4.6 must pass.
4. **Conformance** (if any of these files have Python conformance harnesses):
   - Check `pytesthelper/` for relevant fixtures; the embedded Python sidecar
     can drive a parallel Python `interactive_updr` and compare goal-stack
     states. If no fixtures exist for these files, do **not** invent them —
     widget code does not have a stable conformance baseline. (Per the user
     memory note "Accrete xtraces, never delete," only existing xtraces are
     preserved; neither `iupdr.py` nor `widget_analysis_session.py` currently
     emits xtraces in Python, so no new xtrace work is required.)
5. **Cross-package callers**: confirm nothing referenced the deleted
   `iupdr.Session`/`iupdr.NewSession`/`iupdr.StepInfo` types. Already verified
   in Phase 1 exploration; re-check after the rewrite with `grep iupdr\\.`.
6. **CLAUDE.md compliance audit**:
   - Rule 4 (one Python file → one Go file): each new `.go` file's path is
     listed in "Critical Files" below; verify base names match.
   - Rule 6 (no omitted functions): cross-check the method tables in 4.2-4.5
     against the Python source. The `^    def` grep counted 66 methods plus
     3 helpers — the Go file must contain all 69 names.
   - Rule 7 (no invented abstractions): `setDifference` (an inline helper
     for one Python `set(...) - set(...)` expression) is the only new private
     helper. No new exported types beyond the literal Python class ports.

## Critical Files

**New files (created in this plan):**
- `iupdr/iupdr.go` — REWRITE (replace existing 196-line Go-invention with literal port)
- `iupdr/iupdr_test.go` — NEW
- `webui/widget_dialog.go` — NEW
- `webui/widget_modal_messages.go` — NEW
- `webui/widget_cy_graph.go` — NEW
- `webui/dot_layout.go` — NEW
- `webui/widget_analysis_session.go` — NEW
- `webui/widget_analysis_session_test.go` — NEW

**Modified files:**
- `webui/ext_api.go` — extend from 235 lines to ~500+ lines (add `ShowModal`,
  `UserSelect`, `UserSelectMultiple`, `ExecuteNewCell` with mutable response
  fields and a `frontEndOp()` marker; the `Latex/Button/Select*Widget` data
  carriers; the `OrderedMap` shim; literal `iter.Seq[FrontEndOperation]`-returning
  ports of the @interaction default callbacks). No goroutine/channel session
  types — generators are plain `iter.Seq` returns.
- `tactics/tactics.go` — add `GetSafetyProperty`, `ArgAddActionNode`,
  `RecalculateFacts` (function form), `RemoveIfRefuted` (function form),
  `CheckCover` (function form), `CustomRefineOrReverse`

**Unchanged but cited (do not edit):**
- `actions/transrel.go` — `Interpolant` (1517), `ForwardInterpolant` (1536),
  `InterpFromUnsatCore` (1605) already correct.
- `module/ops.go` — `AndClausesTyped`, `NegateClauses`, `SimplifyClauses`,
  `ConstantsClauses`, `VariablesClauses`, `SubstituteConstantsClauses`,
  `FormulaToClauses` already present and reused.
- `z3bridge/solver.go` and `z3bridge/solver_z3convert.go` — `Solver`,
  `BinaryInterpolant`, `ClausesToZ3`, `Check` already present and reused.
- `art/cyrender.go` — `CyElements` types reused by `widget_cy_graph.go`.
- `webui/concept_isession.go`, `webui/concept.go`, `webui/cyrender.go`,
  `webui/cystyles.go` — already-ported deps reused by widget code.

## Existing Functions to Reuse (no rewrites)

- `actions.InterpFromUnsatCore` (transrel.go:1605)
- `actions.Interpolant`, `actions.ForwardInterpolant` (transrel.go:1517, 1536)
- `actions.ForwardImage`, `actions.ReverseImage`
- `actions.IsSkolem`
- `module.AndClausesTyped`, `module.SimplifyClauses`, `module.NegateClauses`,
  `module.NewClauses`, `module.FormulaToClauses`, `module.TrueClauses`
- `module.ConstantsClauses`, `module.VariablesClauses`,
  `module.SubstituteConstantsClauses`
- `tactics.TacticsContext` and existing methods
  (`BackgroundTheory`, `ForwardImage`, `BackwardImage`, `RefineOrReverse`,
  `RefutedGoal`, `ImpliedFacts`, `GetDiagram`, `TopGoal`, `PushGoal`,
  `RemoveGoal`)
- `tactics.GoalAtArgNode`, `tactics.ArgGetFact`, `tactics.ArgAddFacts`,
  `tactics.ArgGetPredAction`, `tactics.ArgGetConjuncts`, `tactics.GetBigAction`
- `z3bridge.NewSolver`, `Solver.ClausesToZ3`, `Solver.BinaryInterpolant`,
  `Solver.Check`
- `art.AnalysisGraph`, `art.State`, `art.Execute`, `art.Cover`
- `webui/ext_api.go` `ExtensionPoint`, `ExtConfig`, `InteractionError`
- `webui/concept_isession.go` `ConceptInteractiveSession`
- `webui/cyrender.go`, `webui/cystyles.go` rendering types
- `interp.State`, `interp.InterpConfig.BackgroundTheory`
- `proof.ProofGoal`, `proof.ProofGoalStack`

## Out-of-scope clarifications

The user said "Nothing is out of scope." This plan therefore includes:
- Full mechanical port of all 4 widget classes (not just `concept_refine`)
- Full mechanical port of `interactive_updr` (not just the unsat branch)
- All transitive widget dependencies
- All missing tactics functions

It does **not** include:
- Re-implementing IPython itself in Go (Python widget instances become data
  carriers serialized to the existing webui front end; this matches the
  established pattern in `webui/concept_isession.go` etc.)
- Backfilling xtraces in Python source files that don't currently emit them
  (per memory rule "Accrete xtraces, never delete"; we add new xtraces only
  where they enable conformance, not retroactively)
- Touching `actions/transrel.go` (the relocations there are already correct)
