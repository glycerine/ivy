# Plan: Complete ivy_check.py → check/ Port — ALL GROUPS COMPLETED

**Status: All groups (Phase 0, F, C, D, I, A, B, H, E) implemented and tests passing.**
**Also fixed: Annotation type mismatch (string → lg.Expr/lg.NodeKey), SmallModelClauses returns solver.**
**Moved l2s/hooks.go → check/l2s_hooks.go to break import cycle (no factory indirection).**

## Context

The `check/` package is the Go port of Python's `ivy_check.py` — the top-level verification
driver for Ivy. Most functions are ported, but 8 stub/incomplete areas remain. The `tactics/`
package is now fully ported, unblocking the remaining work. This plan completes all stubs with
faithful-to-Python logic, fixes a discovered porting mistake in the annotation subsystem, and
adds comprehensive tests.

## Pre-requisite Fix: Annotation Type Mismatch (porting bug)

**Discovery:** The Go `actions.AnnotationHandler` interface uses `string` for conditions and
environments, but the Python source of truth uses `lg.Symbol` (logic expressions). This is a
porting mistake that cascades through the annotation system and blocks Group D (trace path).

**Python source of truth:**
- `IteAnnotation.cond` = `lg.Symbol` (set via `a.ite(v, annot)` in `ivy_logic_utils.py:1255`)
- `handler.eval(rncond)` → receives `lg.Symbol`, calls `model.eval_to_constant(cond)`
- `env` = `dict[lg.Symbol, lg.Symbol]`
- `RenameAnnotation.map` = `dict[lg.Symbol, lg.Symbol]`

**Go mistake:** `IteAnnotation.Cond` is `string`, `env` is `map[string]string`, `Eval(cond string)`

### Fix (actions/annotation.go, actions/match.go)

1. **`IteAnnotation.Cond`**: `string` → `lg.Expr`
2. **`RenameAnnotation.Map`**: `map[string]string` → `map[string]lg.Expr` (keys stay string since
   `lg.Symbol.Name` is the lookup key in env, matching Python dict keyed by symbol — but values
   must be `lg.Expr`). Actually re-examining Python: `env[x] = env.get(y, y)` where x,y are
   symbols. The env maps symbol→symbol. For Go, use `map[string]lg.Expr` where string key =
   symbol name, value = symbol expression. This avoids needing `lg.Expr` as map key.
3. **`AnnotationHandler`**:
   ```go
   type AnnotationHandler interface {
       Eval(cond lg.Expr) bool
       Handle(action Action, env map[string]lg.Expr)
       DoReturn(action Action, env map[string]lg.Expr)
       Fail()
   }
   ```
4. **`AnnotBranch.Cond`**: `string` → `lg.Expr`
5. **`envGet`**: `envGet(env map[string]lg.Expr, key string) lg.Expr`
6. Update all callers in `matchAnnotationRecur` and `UniteAnnot`.

**Files:** `actions/annotation.go`, `actions/match.go`

---

## Group F: ConjChecker.GetAnnot

**File:** `check/check.go:195-197`
**Python:** `return self.lf.annot if hasattr(self.lf,'annot') else None`

### Steps
1. Add `Annot interface{}` field to `ast.LabeledFormula` (`ast/decl.go:11-22`)
2. Preserve `Annot` in `Clone` method (`ast/decl.go:33-47`)
3. Fix `ConjChecker.GetAnnot()` to return `c.LF.Annot`

**Reuse:** `ast.LabeledFormula.Clone` at `ast/decl.go:33`

---

## Group C: MatchHandler Model Integration

**File:** `check/helpers.go:124-165`

### Steps

**NewMatchHandler (line 124-143):**
1. Change `Clauses` field from `interface{}` to `*clauseops.Clauses`
2. Change `Model` field from `interface{}` to `*solver.ModelResult` (or the concrete type returned by `tr.SmallModelClauses`)
3. Accept `*solver.Solver` parameter (needed for `ClausesModelToClauses`)
4. Call `solver.ClausesModelToClausesWithModel(clauses, model, nil, true)` → `modClauses`
5. Populate `h.Eqs` matching Python:
   ```go
   for _, fmla := range modClauses.Fmlas {
       if lg.IsEq(fmla) {
           lhs := fmla.Children()[0]
           if lg.IsApp(lhs) { h.Eqs[lhs.(*lg.Apply).Func.String()] = append(..., fmla) }
       } else if _, ok := fmla.(*lg.Not); ok {
           app := fmla.Children()[0]
           if lg.IsApp(app) { h.Eqs[app.Rep()] = append(..., lg.NewEquals(app, &lg.Or{})) }
       } else if lg.IsApp(fmla) {
           h.Eqs[fmla.Rep()] = append(..., lg.NewEquals(fmla, &lg.And{}))
       }
   }
   ```

**Eval (line 162-165):**
```go
func (h *MatchHandler) Eval(cond lg.Expr) bool {
    truth := h.Model.EvalToConstant(cond)
    if lg.IsFalse(truth) { return false }
    if lg.IsTrue(truth) { return true }
    panic(fmt.Sprintf("unexpected truth value: %v", truth))
}
```

**ShowSym (line 147-158):** Add `lut.RenameAst(fmla, rmap)` and current-value dedup matching Python.

**Reuse:**
- `solver.Solver.ClausesModelToClausesWithModel` at `solver/model.go:417`
- `solver.HerbrandModel.EvalToConstant` at `solver/herbrand.go:216`

---

## Group D: Trace Path (checkFcsTracePath)

**File:** `check/check.go:484-551`
**Python:** `ivy_check.py:373-416`

### Steps
Replace lines 530-548 with:
1. Extract vocab: `vocab := clauseops.UsedSymbolsClauses(mclauses)`
2. Create trace: `handler := trace.NewTrace(mclauses, model, vocab, true)`
3. Get annotation: `thing := failed[len(failed)-1].GetAnnot()`
4. If nil: build Sequence from `history.Actions`, get `annot` from `clauses.Annot`
5. If non-nil: unpack `(action, annot)` pair
6. Call `actions.MatchAnnotation(action, annot, handler, mod)` — works after annotation fix
7. Call `handler.End()`
8. Handle `mod.TraceHook` if present
9. Set `handler.IsCti` for conjecture failures
10. Print trace or launch GUI

**Note:** `trace.Trace` already has `Eval(cond lg.Expr) (bool, error)` at `trace/trace.go:474`.
After the annotation fix, `trace.Trace` needs to implement `actions.AnnotationHandler` — add a
thin wrapper that adapts `Eval(lg.Expr) (bool, error)` to `Eval(lg.Expr) bool` (panics on error,
matching Python's `assert False, truth`).

**Reuse:**
- `trace.NewTrace` at `trace/trace.go:442`
- `actions.MatchAnnotation` at `actions/match.go:33`
- `clauseops.UsedSymbolsClauses` (or equivalent)

---

## Group I: IsolateProof Handling

**File:** `check/isolate_check.go:54-57`
**Python:** `ivy_check.py:506-516`

### Steps
Replace stub with:
```go
pcAxioms := append(mod.LabeledAxioms[:len(mod.LabeledAxioms):len(mod.LabeledAxioms)],
    mod.AssumedInvs...)
pc := proof.NewProofChecker(nil, pcAxioms, mod.Definitions,
    ModuleSchemataToAst(mod.Schemata))
model := temporal.NormalProgramFromModule(mod)
safetyLabel := ast.NewAtom("safety")
prop := ast.NewLabeledFormula(safetyLabel, &lg.And{})
tm := &ast.TemporalModels{Model: model, Fmla: &lg.And{}}
subgoal := ast.NewLabeledFormula(safetyLabel, tm)
subgoal.Lineno = extractLineno(mod.IsolateProof)
subgoals := []*ast.LabeledFormula{subgoal}
subgoals, err := pc.AdmitProposition(prop, mod.IsolateProof.(ast.Node), subgoals...)
if err != nil { return err }
return CheckSubgoals(subgoals, nil, mod)
```

**Reuse:**
- `temporal.NormalProgramFromModule` at `temporal/temporal.go:225`
- `proof.ProofChecker.AdmitProposition` at `proof/checker.go:388`

---

## Group A: CheckTemporals Full Wiring

**File:** `check/check.go:314-348`
**Python:** `ivy_check.py:127-157`

### Steps
Replace stub body with full implementation:
```go
func CheckTemporals(mod *module.Module) error {
    pmap := make(map[int64]interface{})
    for _, pe := range mod.Proofs { pmap[pe.Formula.ID] = pe.Proof }

    pcAxioms := append(mod.LabeledAxioms[:len(mod.LabeledAxioms):len(mod.LabeledAxioms)],
        mod.AssumedInvs...)
    pc := proof.NewProofChecker(nil, pcAxioms, mod.Definitions,
        ModuleSchemataToAst(mod.Schemata))

    for _, prop := range mod.LabeledProps {
        if !prop.Temporal { continue }
        if prop.Assumed || (mod.Cfg.OptUncheckedProps != "" &&
            acl.IsAssumed(fmt.Sprint(prop.Label))) {
            fmt.Println("  ivy_check temporal: admitting axiom...", PrettyLF(prop, 0))
            pc.AdmitAxiom(prop)
        } else {
            fmt.Print("\n    The following temporal property is being proved:\n")
            fmt.Print(PrettyLF(prop, 4) + " ... ")
            pf, _ := pmap[prop.ID]
            propn := proof.NormalizeGoal(prop)
            model := temporal.NormalProgramFromModule(mod)
            subgoal := prop.Clone([]ast.Node{
                prop.Label,
                &ast.TemporalModels{Model: model, Fmla: propn.Formula},
            }).(*ast.LabeledFormula)
            subgoals := []*ast.LabeledFormula{subgoal}
            var pfNode ast.Node
            if pf != nil { pfNode, _ = pf.(ast.Node) }
            var err error
            subgoals, err = pc.AdmitProposition(prop, pfNode, subgoals...)
            if err != nil { return err }
            if err := CheckSubgoals(subgoals, nil, mod); err != nil { return err }
        }
    }
    return nil
}
```

**Reuse:**
- `proof.NormalizeGoal` at `proof/goal.go:92`
- `proof.ProofChecker.AdmitAxiom` at `proof/checker.go:129`
- `proof.ProofChecker.AdmitProposition` at `proof/checker.go:388`
- `temporal.NormalProgramFromModule` at `temporal/temporal.go:225`

---

## Group B: MCTactic/VMTTactic Temporal Chain

**File:** `check/check.go:981-1007`
**Python:** `ivy_check.py:805-831`

### Steps

**MCTactic:**
```go
func MCTactic(prover interface{}, goals []*ast.LabeledFormula,
    proofNode ast.Node, mod *module.Module) ([]*ast.LabeledFormula, error) {
    if len(goals) == 0 { return nil, nil }
    goal := goals[0]
    conc := proof.GoalConc(goal)
    if tm, ok := conc.(*ast.TemporalModels); ok {
        if fmlaExpr, ok := tm.Fmla.(lg.Expr); ok && !lg.IsTrue(fmlaExpr) {
            pc, _ := prover.(*proof.ProofChecker)
            var err error
            goals, err = tactics.Tempind(pc, goals, proofNode)
            if err != nil { return nil, err }
            goals, err = tactics.Skolemizenp(pc, goals, proofNode)
            if err != nil { return nil, err }
            // Python: l2s_pf = proof.clone([proof.args[0], TacticLets()] + list(proof.args[2:]))
            l2sPf := cloneProofWithTacticLets(proofNode)
            goals, err = l2s.L2STacticFull(pc, goals, l2sPf)
            if err != nil { return nil, err }
        }
    }
    mcMethod := func() error {
        res, err := mc.CheckIsolate(mod, "mc")
        if err != nil { return err }
        if res != nil && !res.Proved { return fmt.Errorf("model checking failed") }
        return nil
    }
    err := CheckSubgoals(goals[0:1], mcMethod, mod)
    return goals[1:], err
}
```

**VMTTactic:** Same structure, pass `vmt.CheckIsolate` as method.

**Helper `cloneProofWithTacticLets`:** Clone proof node with `TacticLets{}` inserted as second arg.

**Reuse:**
- `tactics.Tempind` at `tactics/ivy_tactics.go`
- `tactics.Skolemizenp` at `tactics/ivy_tactics.go`
- `l2s.L2STacticFull` at `l2s/l2s.go:207`
- `ast.TacticLets` at `ast/tactic.go:261`

---

## Group H: CheckModule macro_finder Save/Restore

**File:** `check/isolate_check.go` around line 747 (inside isolate loop)
**Python:** `ivy_check.py:928-938, 962-965`

### Steps
Add before/after each isolate check:
```go
var saveMacroFinder bool
hasMFAttr := false
if isolate != "" {
    attrKey := iu.ComposeNames(isolate, "macro_finder")
    if _, ok := mod.Attributes[attrKey]; ok {
        hasMFAttr = true
        saveMacroFinder = mod.Cfg.MacroFinder
        if saveMacroFinder {
            fmt.Println("Turning off macro_finder")
            solver.SetMacroFinder(false) // or isoMod.Solver.SetMacroFinder(false)
        }
    }
}
// ... check isolate ...
if hasMFAttr && saveMacroFinder {
    fmt.Println("Turning on macro_finder")
    solver.SetMacroFinder(true)
}
```

**Reuse:** `solver.Solver.SetMacroFinder` at `solver/encoding.go:227`

---

## Group E: Start() Entry Point

**File:** `check/check.go:1024-1031`
**Python:** `ivy_check.py:975-999`

### Steps
```go
func Start(args []string) error {
    if len(args) < 1 || !strings.HasSuffix(args[0], ".ivy") {
        return fmt.Errorf(Usage())
    }
    someBounded := false
    mod := module.New()
    // ivyinit.SourceFile(args[0], mod, ...)
    if err := ivyinit.SourceFile(args[0], mod, nil, map[string]interface{}{
        "create_isolate": false,
    }); err != nil { return err }
    // NOT CHECKED check
    if mod.Cfg.CheckLineno == "none.ivy:0" {
        fmt.Println("NOT CHECKED"); return nil
    }
    if err := CheckModule(mod); err != nil { return err }
    if someBounded { fmt.Println("BOUNDED") }
    if tactics.UsedSorry {
        fmt.Println("OK, but used 'sorry'")
    } else {
        fmt.Println("OK")
    }
    return nil
}
```

Also update `Main()` to match Python's `main()` (set recursion limit analog, read params, etc.)

---

## Implementation Order

| Phase | Groups | Rationale |
|-------|--------|-----------|
| 0 | Annotation fix | Unblocks Group D; fixes porting bug in actions/ |
| 1 | F (GetAnnot) | Trivial, unblocks Group D |
| 2 | C (MatchHandler) | Needed for trace path |
| 3 | D (trace path) | Depends on 0, 1, 2 |
| 4 | I (IsolateProof) | Independent, uses temporal + proof |
| 5 | A (CheckTemporals) | Core temporal verification |
| 6 | B (MC/VMT tactics) | Depends on tactics package |
| 7 | H (macro_finder) | Low complexity |
| 8 | E (Start entry point) | Final integration |

---

## Files to Modify

| File | Changes |
|------|---------|
| `actions/annotation.go` | Fix `IteAnnotation.Cond`, `RenameAnnotation.Map`, `AnnotBranch` types |
| `actions/match.go` | Fix `AnnotationHandler` interface, `matchAnnotationRecur`, `envGet` |
| `ast/decl.go` | Add `Annot` field to `LabeledFormula`, update `Clone` |
| `check/check.go` | Groups A, B, D, E, F |
| `check/helpers.go` | Group C (MatchHandler) |
| `check/isolate_check.go` | Groups H, I |
| `trace/trace.go` | Implement `AnnotationHandler` interface on `Trace` |

---

## Testing Plan

### Unit Tests (check/check_test.go additions)

- `TestConjCheckerGetAnnot` — with/without Annot field
- `TestNewMatchHandlerPopulatesEqs` — mock model, verify Eqs
- `TestMatchHandlerEvalTrueFalse` — model evaluation
- `TestCheckTemporalsAssumed` — admitted as axiom path
- `TestCheckTemporalsWithProof` — full pipeline
- `TestMCTacticTemporalChain` — temporal tactic chain fires
- `TestMCTacticNonTemporal` — bypass chain
- `TestVMTTacticUsesVMTMethod` — correct method dispatch
- `TestIsolateProofHandling` — proof pipeline runs
- `TestCheckModuleMacroFinderToggle` — save/restore
- `TestStartInvalidArgs` — error handling
- `TestAnnotationHandlerTypes` — verify lg.Expr flows through annotation system

### Fuzz Tests (check/check_fuzz_test.go)

- `FuzzMatchHandlerEqs` — random formulas through Eqs extraction, no panics
- `FuzzDualClauses` — random clause sets through DualClauses
- `FuzzCheckTemporalsProps` — random properties through CheckTemporals
- `FuzzAnnotationMatchRoundTrip` — random annotation trees through match_annotation

### Integration Tests (check/regression_test.go additions)

- End-to-end with safety property (exercises property checking + axiom promotion)
- End-to-end with failing conjecture (exercises trace path, Groups C+D)
- End-to-end with temporal property + proof (exercises Groups A+B)
- End-to-end with isolate proof (exercises Group I)

### Annotation System Tests (actions/match_test.go additions)

- `TestMatchAnnotationWithLgExprCond` — IteAnnotation with `lg.Symbol` condition
- `TestRenameAnnotationWithLgExprValues` — env mapping with `lg.Expr` values
- `TestUniteAnnotWithLgExprConds` — flattened branches have correct types

## Verification

1. `go build ./...` — all packages compile
2. `go vet ./...` — no warnings
3. `go test ./check/...` — all existing + new tests pass
4. `go test ./actions/...` — annotation fix doesn't break existing tests
5. `go test ./trace/...` — trace tests still pass
6. `go test ./ast/...` — LabeledFormula tests still pass
