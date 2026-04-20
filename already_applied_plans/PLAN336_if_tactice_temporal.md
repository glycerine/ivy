# Plan: Fix `if_tactic` and `let_tactic` TemporalModels descent divergence

Created: 2026-04-20 (date + time only; exact clock time omitted per CLAUDE.md rule E)

## Context

Go's `ifTactic` (proof/tactics.go:384-447) and `letTactic` (proof/tactics.go:27-91) both call `ApplyToConc` (proof/goal.go:77-88) to wrap a goal's conclusion in an `Implies(cond, ...)`. `ApplyToConc` descends into `*ast.TemporalModels`, re-wrapping the result inside the temporal wrapper.

Python's `if_tactic` (ivy_proof.py:412-420) and `let_tactic` (ivy_proof.py:221-231) do the opposite — they wrap `decls[0].formula` *directly* with `il.Implies(cond, decls[0].formula)`, without any descent, regardless of whether the formula is a `TemporalModels`, a `SchemaBody`, or a plain formula. Python's `il.Implies` is duck-typed and accepts any formula as its second argument.

Structural divergence on a goal whose formula is `TemporalModels(M, phi)`:
- Python: `Implies(cond, TemporalModels(M, phi))`
- Go (now): `TemporalModels(M, Implies(cond, phi))`

Same class of mismatch for `SchemaBody`. The `letTactic` comment at tactics.go:78 even asserts the descent is "correct" — it is not; it is a divergence from the source of truth.

Per CLAUDE.md rule B9, the current note at tactics.go:407 ("best approximation") is exactly the kind of shortcut that must be replaced with a faithful port. Per the user's `feedback_port_match_everywhere.md` memory, every mirror site must be fixed — `letTactic` shares the same pattern and must be fixed alongside `ifTactic`.

## Type-system constraint

`lg.Implies` (logic/formula.go:320) has `T1, T2 lg.Expr`. It cannot hold `*ast.TemporalModels` or `*ast.SchemaBody` — neither implements `lg.Expr` (missing `NodeSort`, `Children`, `Equal(Expr)`, `Sexp`). Making them implement `lg.Expr` would require `ast` to import `logic` types, but `logic` already imports `ast` — circular. Moving `TemporalModels` into `logic` is a multi-file refactor with cascading effects.

`ast.Implies` (ast/formula.go:90-108) already exists with `T1, T2 Node` (ast.Node). It can hold anything. It is constructed by `cfg.NewImplies(t1, t2 Node) *Implies`. Production code rarely uses it; downstream type-switches target `*lg.Implies`.

## Recommended approach — hybrid typed `Implies`

Wrap with `*lg.Implies` when `goal.Formula` is `lg.Expr`; fall back to `*ast.Implies` when it is not. This preserves the common-case downstream type (most goal formulas are `lg.Expr`, so every `*lg.Implies` type-switch keeps matching) while still producing the Python structure for the edge cases. The asymmetry is an honest concession to Go's type system; it is not a "shortcut" because both branches structurally wrap the *entire* formula, never descend.

A shared helper keeps the choice in one place:

```go
// proof/goal.go — new helper, placed next to ApplyToConc for visibility.
// Mirrors Python's il.Implies(cond, formula) duck-typed wrapping: accepts any
// formula (lg.Expr, *ast.TemporalModels, *ast.SchemaBody). Produces *lg.Implies
// when the RHS is lg.Expr so downstream type-switches still fire; otherwise
// *ast.Implies via cfg.NewImplies.
func WrapImplies(cfg *ast.AstConfig, cond lg.Expr, formula ast.Node) ast.Node {
    if e, ok := formula.(lg.Expr); ok {
        return &lg.Implies{T1: cond, T2: e}
    }
    return cfg.NewImplies(cond, formula)
}
```

## Files to modify

Only three files:

1. `/Users/jaten/ivy/goivy/proof/goal.go` — add `WrapImplies` helper.
2. `/Users/jaten/ivy/goivy/proof/tactics.go` — rewrite `letTactic` body and `ifTactic` body to use `WrapImplies` on `goal.Formula` directly; delete the `ApplyToConc` + `CloneGoal(..., GoalPrems(goal), ...)` dance; delete the `goal.Formula.(lg.Expr)` split in `ifTactic`.
3. Potentially `/Users/jaten/ivy/goivy/proof/tactics_test.go` (or new file) — add coverage for the three shapes (lg.Expr, TemporalModels, SchemaBody). Check if file already exists first.

Do NOT touch `ApplyToConc` itself — it is still used correctly by `phase5_goals.go:398, 723, 837`, `phase5_matching.go:1255, 1393, 1461`, `skolem.go:303`, and `goal.go:189`. Those correspond to Python sites that legitimately descend (`apply_to_conc` on line 1369, 1374, etc., and `goal_apply_to_conc` on line 401).

## Edits

### Edit 1 — `proof/tactics.go` letTactic body (lines 76-90)

Before (current):
```go
// Build subgoal: cond -> original formula.
// Python uses goal.formula directly (wraps entire formula including SchemaBody).
// Go uses ApplyToConc to handle *ast.TemporalModels wrapping correctly.
if GoalConc(goal) == nil {
    return nil, &ProofError{Msg: "let tactic: goal has no conclusion"}
}
newConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
    return &lg.Implies{T1: cond, T2: c}
})
subgoal := CloneGoal(pc.astCfg(), goal, GoalPrems(goal), newConc)
subgoal.SetLineno(goal.GetLineno())
```

After (matches Python ivy_proof.py:226 `subgoal = ia.LabeledFormula(decls[0].label, il.Implies(cond, decls[0].formula))` exactly):
```go
// Python ivy_proof.py:226:
//   subgoal = ia.LabeledFormula(decls[0].label, il.Implies(cond, decls[0].formula))
// Wrap the ENTIRE goal.Formula (SchemaBody or TemporalModels included) — no descent.
subgoal := pc.astCfg().NewLabeledFormula(goal.Label, WrapImplies(pc.astCfg(), cond, goal.Formula))
subgoal.SetLineno(goal.GetLineno())
```

The stale "Go uses ApplyToConc to handle *ast.TemporalModels wrapping correctly" comment and the `GoalConc(goal) == nil` guard both go away — Python has no such guard; `goal.Formula` is always present.

### Edit 2 — `proof/tactics.go` ifTactic body (lines 397-416)

Before (current):
```go
// Python: true_goal = ia.LabeledFormula(decls[0].label, il.Implies(cond, decls[0].formula))
// Uses decls[0].formula directly (not goal_conc). If formula is lg.Expr, wrap it.
// For SchemaBody (rare), fall back to wrapping just the conclusion.
var trueGoal, falseGoal *ast.LabeledFormula
if goalExpr, ok := goal.Formula.(lg.Expr); ok {
    trueGoal = pc.astCfg().NewLabeledFormula(goal.Label,
        lg.Expr(&lg.Implies{T1: cond, T2: goalExpr}))
    falseGoal = pc.astCfg().NewLabeledFormula(goal.Label,
        lg.Expr(&lg.Implies{T1: &lg.Not{Body: cond}, T2: goalExpr}))
} else {
    // SchemaBody or other non-expr: wrap the conclusion (best approximation)
    trueConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
        return &lg.Implies{T1: cond, T2: c}
    })
    trueGoal = CloneGoal(pc.astCfg(), goal, GoalPrems(goal), trueConc)
    falseConc := ApplyToConc(GoalConc(goal), func(c lg.Expr) lg.Expr {
        return &lg.Implies{T1: &lg.Not{Body: cond}, T2: c}
    })
    falseGoal = CloneGoal(pc.astCfg(), goal, GoalPrems(goal), falseConc)
}
trueGoal.SetLineno(goal.GetLineno())
falseGoal.SetLineno(goal.GetLineno())
```

After (matches Python ivy_proof.py:414,416):
```go
// Python ivy_proof.py:414-416:
//   true_goal  = ia.LabeledFormula(decls[0].label, il.Implies(cond,      decls[0].formula))
//   false_goal = ia.LabeledFormula(decls[0].label, il.Implies(Not(cond), decls[0].formula))
// Wrap the ENTIRE goal.Formula (SchemaBody or TemporalModels included) — no descent.
notCond := &lg.Not{Body: cond}
trueGoal := pc.astCfg().NewLabeledFormula(goal.Label, WrapImplies(pc.astCfg(), cond, goal.Formula))
falseGoal := pc.astCfg().NewLabeledFormula(goal.Label, WrapImplies(pc.astCfg(), notCond, goal.Formula))
trueGoal.SetLineno(goal.GetLineno())
falseGoal.SetLineno(goal.GetLineno())
```

### Edit 3 — `proof/goal.go` add `WrapImplies` helper

Add the helper shown above, next to `ApplyToConc` (around line 88). Docstring must reference ivy_proof.py:414 and note that Python is duck-typed while Go uses the hybrid typing because of `lg.Implies`'s `lg.Expr` field constraint.

### Edit 4 — tests

Check for an existing `proof/tactics_test.go`. If present, add:
- `TestIfTacticWrapsLgExprDirectly` — `goal.Formula = lg.Expr`; assert result is `*lg.Implies{T1: cond, T2: same-expr}`.
- `TestIfTacticWrapsTemporalModelsDirectly` — `goal.Formula = *ast.TemporalModels`; assert result top-level is `*ast.Implies` and its T2 is the *same* TemporalModels (no clone, no descent).
- `TestIfTacticWrapsSchemaBodyDirectly` — same shape check with `*ast.SchemaBody`.
- Three parallel cases for `letTactic`.

If no `tactics_test.go` exists, create `proof/tactics_let_if_test.go`. Use `test_vectors/` directory convention per `feedback_test_vectors_dir.md` if any file fixtures are needed (they are not for these tests).

## Downstream-impact audit

`*ast.Implies` is a *new* top-level shape for subgoals whose formula is `TemporalModels` or `SchemaBody`. The next tactic in the proof will see `*ast.Implies` at the top of `goal.Formula`. Grep shows no production site currently handles `*ast.Implies`, but this is OK because:

- Python's downstream is duck-typed: it accesses `.args[0]`, `.args[1]`. Nothing Python does with `Implies(cond, TemporalModels(...))` inspects the wrapper Implies type.
- Go's downstream `*lg.Implies` type-switches (check/ranking.go, check/l2s_hooks.go, ivylogic/, gogen/, leangen/, fragment/, compiler/) all run on goals whose formulas are plain logic expressions, NOT on SchemaBody/TemporalModels goals. The edge cases that produce `*ast.Implies` do not flow into those passes.
- If a downstream consumer *does* need to see through the edge-case `*ast.Implies`, add a parallel `case *ast.Implies` arm where needed. Do NOT widen `*lg.Implies.T1/T2` to `ast.Node` — that would ripple through every existing `case *lg.Implies` in the codebase.

## Verification

1. `cd ~/ivy/goivy && make test` (NEVER `go test ./...` — XTRACE on by default, per `feedback_use_make_test.md`). All existing tests must stay green.
2. Snapshot the `log.golden.2hr` output before and after. The change should produce identical output for tests that don't exercise if/let on temporal goals, and aligned output for tests that do.
3. If `test_vectors/` has an `.ivy` file with a temporal property + if/let proof (grep under `test_vectors/` for `proof {` combined with `if ` / `let ` and `temporal` / `|=` ), run it through goivy and compare the Canon() output of the resulting subgoals to the Python `.canon()` output per CLAUDE.md section F.
4. Run the new unit tests added in Edit 4.
5. Manual spot-check: construct a `LabeledFormula` with a `*ast.TemporalModels` body, run `ifTactic`, print `trueGoal.Formula.Canon()`. Expected: `(implies ... t1:... t2:(temporalModels ...))`.

## What I deliberately did NOT do

- Did not change `ApplyToConc` (used correctly elsewhere).
- Did not widen `lg.Implies.T1/T2` to `ast.Node` (would ripple across the codebase).
- Did not move `TemporalModels` out of `ast` into `logic` (large refactor; out of scope).
- Did not add `lg.Expr` methods to `*ast.TemporalModels` or `*ast.SchemaBody` (circular import).
