# L2S Pipeline Comprehensive Conformance Audit

**Created:** 2026-04-12 ~late evening (updated ~night)

## Context

Golden test `TestOrdLive` diverges at trace line 260353. A comprehensive step-by-step audit of the Go l2s pipeline, ranking pipeline, and all SharedStep functions against Python revealed:

- **The l2s pipeline (l2s.go + SharedStep functions) is a faithful port** — no semantic divergences found across all 8 SharedStep functions, all 3 modPass calls, all axiom building, all instrumentation.
- **The ranking pipeline (ranking.go) has 5 bugs** — 2 already fixed (12-13), 3 remaining (14-16).
- **One cross-cutting bug** in SharedStep8 (Bug 15) affects both l2s and ranking.

## Guiding Principle

The point of tests is to find bugs. If a test uncovers a bug in production code, fix the production code — never weaken the test. Strengthen tests in that area.

---

## Already Fixed

### Bug 12: Ranking modPass missing property prem transformation — FIXED
`ranking.go:327-336` now transforms property prems via `list_transform` pattern.

### Bug 13: Ranking invars not merged into model.Invars — FIXED
`ranking.go:293` now does `model.Invars = append(model.Invars, invars...)`.

---

## Bug 14: Missing WhenOperator invariant generation in ranking

**Files:** `check/ranking.go` (missing entirely)
**Python:** `ivy_ranking.py:449-461`

Python collects property prems, scans `invars+pprems` for temporal operators, and generates invariants for `WhenOperator{name:"first"}`:
```python
pprems = list(x for x in prems if ipr.goal_is_property(x))
winvs = []
for tmprl in list(iu.unique(ilu.temporals_asts(invars+pprems))):
    if isinstance(tmprl, lg.WhenOperator):
        if tmprl.name == 'first':
            nws = lg.Or(lg.Not(l2s_waiting), lg.Not(l2s_w((), tmprl.t2)))
            tmp = lg.Implies(lg.Not(nws), lg.Eq(tmprl, lg.WhenOperator('next', tmprl.t1, tmprl.t2)))
            tmp = lg.Implies(l2s_init((), lg.Eventually(proof_label, tmprl.t2)), tmp)
            winvs.append(tmp)
for i, winv in enumerate(winvs):
    invars.append(mklf("l2s_when_"+str(i), winv))
```

This runs BEFORE the invars merge (Python line 511), so generated invariants get merged.

**Fix:** Add between desugar block (~line 290) and invars merge (~line 293):
```go
// Python ivy_ranking.py:449-461: WhenOperator invariant generation from pprems
var pprems []lg.Expr
for _, p := range prems {
    if lf, ok := p.(*ast.LabeledFormula); ok && proof.GoalIsProperty(lf) {
        if e, ok := lf.Formula.(lg.Expr); ok {
            pprems = append(pprems, e)
        }
    }
}
var allFmlas []lg.Expr
for _, inv := range invars {
    if e, ok := inv.Formula.(lg.Expr); ok {
        allFmlas = append(allFmlas, e)
    }
}
allFmlas = append(allFmlas, pprems...)
seen := make(map[string]bool)
var winvs []lg.Expr
for _, f := range allFmlas {
    for _, t := range lu.TemporalsAst(f) {
        if wo, ok := t.(*lg.WhenOperator); ok && wo.Name == "first" {
            key := wo.String()
            if seen[key] { continue }
            seen[key] = true
            nws := &lg.Or{Terms: []lg.Expr{
                &lg.Not{Body: L2SWaiting()},
                &lg.Not{Body: applyNB(l2sW(nil, wo.T2, proofLabel))},
            }}
            tmp := &lg.Implies{
                T1: &lg.Not{Body: nws},
                T2: &lg.Eq{T1: wo, T2: &lg.WhenOperator{Name: "next", T1: wo.T1, T2: wo.T2}},
            }
            tmp2 := &lg.Implies{
                T1: applyNB(l2sInit(nil, &lg.Eventually{Environ: strPtr(proofLabel), Body: wo.T2}, proofLabel)),
                T2: tmp,
            }
            winvs = append(winvs, tmp2)
        }
    }
}
acfg := cfg.Mod.Cfg.AstCfg
for i, winv := range winvs {
    label := lg.NewConst(fmt.Sprintf("l2s_when_%d", i), lg.Boolean)
    invars = append(invars, acfg.NewLabeledFormula(label, winv))
}
```

**Note:** This may not manifest in the ord_live.ivy test if no `WhenOperator{name:"first"}` nodes exist. Check after Bugs 12-13 fix whether golden test passes; if not, implement this.

**Depends on:** Verify `lg.WhenOperator` type exists and `lu.TemporalsAst` function exists. If not, this needs to be adapted.

---

## Bug 15: AddParamsToD extra type check in SharedStep8

**Files:** `check/l2s_shared.go:533-538`
**Python:** `ivy_l2s.py:1228-1231`

Python accepts ANY non-finite sort for addParamsToD:
```python
add_params_to_d = [
    AssignAction(l2s_d(p.sort)(p), lg.true)
    for p in b.action.inputs
    if p.sort.name not in finite_sorts
]
```

Go has an EXTRA type assertion restricting to `*lg.UninterpretedSort`:
```go
if p.CSort != nil && !cfg.FiniteSorts[p.CSort.String()] {
    if _, ok := p.CSort.(*lg.UninterpretedSort); ok {  // ← EXTRA CHECK
        addParamsToD = append(addParamsToD, ...)
    }
}
```

**Impact:** Parameters with non-finite sorts that aren't `*lg.UninterpretedSort` (e.g., function sorts with uninterpreted range) would be skipped in Go but included in Python. This could cause missed l2s_d assignments for some parameters.

**Fix:** Remove the inner type assertion:
```go
if p.CSort != nil && !cfg.FiniteSorts[p.CSort.String()] {
    addParamsToD = append(addParamsToD,
        setLineno(actions.NewAssignAction(mustApply(L2SD(p.CSort), p), lg.True), cfg.Lineno))
}
```

**Risk:** `L2SD(p.CSort)` creates a function sort from `p.CSort → Boolean`. If `p.CSort` is something unexpected, this could panic. Check if `L2SD` handles non-UninterpretedSort inputs.

---

## Bug 16: Missing ranking trace_hook setup

**Files:** `check/ranking.go:422-426`
**Python:** `ivy_ranking.py:1016`

Python attaches a trace_hook to the ranking goal:
```python
goal.trace_hook = lambda tr,fcs: auto_hook(tasks,triggers,subs,tr,fcs)
```

Go's ranking tactic returns from `SharedStep12_BuildGoal` without setting any trace_hook:
```go
if tm != nil {
    return SharedStep12_BuildGoal(cfg.Mod.Cfg.AstCfg, goal, cfg.Goals, prems, tm)
}
```

For comparison, Go's l2s tactic (l2s.go:781-820) DOES set trace_hook after SharedStep12. The ranking tactic should do the same.

**Fix:** After `SharedStep12_BuildGoal` returns, set trace_hook on `result[0]`:
```go
if tm != nil {
    result, err := SharedStep12_BuildGoal(cfg.Mod.Cfg.AstCfg, goal, cfg.Goals, prems, tm)
    if err != nil {
        return nil, err
    }
    if len(result) > 0 && result[0] != nil {
        subs := icfg.Subs
        tasks := icfg.Tasks
        triggers := icfg.Triggers
        result[0].TraceHook = TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
            if subs != nil {
                applyRenamingToHandler(handler, subs)
            }
            applyAutoDiagnosticsToHandler(handler, fcs, tasks, triggers)
        })
    }
    return result, nil
}
```

---

## Audit Confirmation: What IS Correct

The comprehensive audit confirmed these areas are faithful ports with **no divergences**:

| Component | Status |
|-----------|--------|
| l2s.go modPass (3 calls, same transforms) | ✓ Matches Python |
| SharedStep1_ConvertTemporals | ✓ Matches Python |
| SharedStep3_CollectNamedBinders | ✓ Matches Python |
| SharedBuildSaveAndWait (save_state, done_waiting, reset_w) | ✓ Matches Python |
| SharedStep6_BuildTableau (all 4 axiom types) | ✓ Matches Python |
| SharedStep7_InstrumentActions (split_returns, events) | ✓ Matches Python |
| SharedStep8_PatchExports (except Bug 15 type check) | ✓ Matches Python |
| SharedStep11_ReplaceNamedBinders | ✓ Matches Python |
| SharedStep12_BuildGoal | ✓ Matches Python |
| l2sAutoInvariants (all auto5 patterns) | ✓ Matches Python |
| InstrumentationConfig carries all needed state | ✓ Complete |
| Monitor state machine (wait→frozen→saved) | ✓ Matches Python |
| Fair cycle construction | ✓ Matches Python |
| Idle action construction | ✓ Matches Python |
| Init action construction | ✓ Matches Python |

---

## Implementation Order

1. **Bug 15** — Remove extra type check in SharedStep8 (1 line, zero cascade, affects both l2s+ranking)
2. **Bug 16** — Add trace_hook setup in ranking.go (small block, pattern from l2s.go)
3. **Bug 14** — WhenOperator invariant generation (medium, needs type verification first)

---

## Verification

1. `go build ./...` — compile clean
2. `go test ./check/... -count=1` — existing + new tests pass
3. `make test-ordlive` — golden test should now advance further or pass
4. For Bug 15: verify `L2SD` works with non-UninterpretedSort inputs before removing guard
5. For Bug 16: verify trace_hook output matches Python format
6. For Bug 14: verify `lg.WhenOperator` exists; if not, check what Go type represents Python's WhenOperator

---

## Unit Tests to Add

### Bug 15: TestAddParamsToD_NonUninterpSort
- Create a binding with a parameter of function sort (non-finite, non-UninterpretedSort)
- Verify addParamsToD includes it (after fix)

### Bug 16: TestRankingTraceHook_Set
- Verify that after SharedStep12, the result goal has a non-nil TraceHook
- Verify the hook calls applyRenamingToHandler and applyAutoDiagnosticsToHandler

### Bug 14: TestWhenOperatorInvarGeneration
- Create an expression with WhenOperator{name:"first"}
- Verify invariants are generated with correct structure

---

## Files to Modify

| File | Changes |
|------|---------|
| `check/l2s_shared.go` | Bug 15 (remove type check in AddParamsToD) |
| `check/ranking.go` | Bug 16 (add trace_hook), Bug 14 (WhenOperator invariants) |
| `check/ranking_test.go` | Unit tests for Bugs 14-16 |
