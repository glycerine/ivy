# Plan: Fix Section 9 (ivy_check.py -> check/) Audit Issues

## Context

The AUDIT18MARCH.md section 9 lists 20 issues (11 MISSING, 5 STUB, 4 BEHAVIORAL_DIFFERENCE) in the `check/` package port from Python's `ivy_check.py`. Many have been **partially or fully fixed** since the audit was written. After cross-referencing the audit against current Go code, **8 substantive issues remain**. The key insight is that most required infrastructure (`NormalProgramFromModule`, `AdmitProposition`, `ClausesModelToClauses`, `EvalToConstant`, `NewTrace`, `MatchAnnotation`) already exists in Go — the remaining work is mostly **wiring**.

## Triage: Already Fixed vs Still Open

### Already Fixed (no work needed)
| Audit # | Item | Status |
|---------|------|--------|
| 9.1 #1 | `check_properties()` with `itp.false_properties()` | FIXED — `CheckProperties` calls `interp.FalseProperties`, promotes to axioms, calls `UpdateTheory` |
| 9.1 #2 | `check_conjectures` no-op | FIXED — calls `interp.UndecidedConjectures` |
| 9.1 #6 | `check_subgoals` trivial stub | FIXED — full implementation at `isolate_check.go:514`, handles both temporal and non-temporal branches |
| 9.1 #9 | `convert_postconds` returns unchanged | FIXED — `ConvertPostcondsWithUpdate` at `check.go:728` does proper old/new variable renaming |
| 9.2 #1 | `CheckProperties` promotes without checking | FIXED — calls `interp.FalseProperties` first |
| 9.2 #2 | `CheckConjectures` returns nil | FIXED — calls `interp.UndecidedConjectures` |
| 9.2 #4 | `ApplyConjProofs` proof=no-proof identical | FIXED — proof branch calls `pc.ApplyProof` at `check.go:415` |
| 9.2 #5 | `PreprocessAssumedIgnoredProperties` temporal incorrect | FIXED — correctly filters temporal from assumed at `check.go:872` |
| 9.3 #1 | `Checker.Sat()` always calls Fail | FIXED — inverts on `OnlyCheckUnprovable` at `check.go:136` |
| 9.3 #2 | `CheckIsolate` init pattern | N/A — Go uses function call pattern, not constructor; functionally equivalent |

### Still Open (8 items, grouped below)

---

## Grouped Action Plan

### Group A: CheckTemporals Full Wiring (Audit 9.1 #5, 9.2 #3)
**Complexity: Medium** | **File: `check/check.go:313-348`**

**Problem:** `CheckTemporals` prints stub messages instead of actually verifying temporal properties. The proof path (`NormalizeGoal` -> `NormalProgramFromModule` -> `TemporalModels` -> `AdmitProposition` -> `CheckSubgoals`) is commented out.

**All required infrastructure exists:**
- `proof.NormalizeGoal()` at `proof/goal.go:92`
- `temporal.NormalProgramFromModule()` at `temporal/temporal.go:225`
- `ast.TemporalModels` type exists (checked in ast package)
- `proof.ProofChecker.AdmitProposition()` at `proof/checker.go:388`
- `check.CheckSubgoals()` at `isolate_check.go:514`

**Fix:** Replace the stub body (lines 336-345) with:
```go
propn := proof.NormalizeGoal(prop)
model := temporal.NormalProgramFromModule(mod)
subgoal := prop.Clone(prop.Args[0], ast.NewTemporalModels(model, propn.Args[1]))
subgoals := []*ast.LabeledFormula{subgoal}
subgoals, err := pc.AdmitProposition(prop, proof.(ast.Node), subgoals...)
if err != nil {
    return err
}
if err := CheckSubgoals(subgoals, nil, mod); err != nil {
    return err
}
```
Also need to build the `ProofChecker` (`pc`) at the top of the function, and call `pc.AdmitAxiom(prop)` for assumed temporal properties.

**Python reference:** `ivy_check.py:127-157`

---

### Group B: MC/VMT Tactic Temporal Chain (Audit 9.1 #7)
**Complexity: Medium** | **File: `check/check.go:980-1006`**

**Problem:** `MCTactic` and `VMTTactic` skip the temporal tactic chain (`tempind` -> `skolemizenp` -> `l2s_tactic_full`) and go straight to `CheckSubgoals`. Also, `VMTTactic` calls `CheckSubgoals` with `nil` method instead of `vmt.CheckIsolate`.

**Infrastructure status:**
- `l2s.L2STacticFull` EXISTS at `l2s/l2s.go:207`
- `proof.SkolemizeGoal` EXISTS at `proof/skolem.go:17` (equivalent of `skolemizenp`)
- `Tempind` (temporal induction tactic) does NOT exist — needs porting from Python `ivy_tactics.py`

**Fix (three parts):**

1. **Method dispatch** (easy): `MCTactic` should pass `mc.CheckIsolate` as method; `VMTTactic` should pass `vmt.CheckIsolate`. Currently both pass `nil`.

2. **Wire existing tactics**: For TemporalModels goals where `!lg.IsTrue(conc.Fmla)`:
   ```go
   // When tempind is ported:
   // goals = tactics.Tempind(prover, goals, proof)
   goals = proof.SkolemizeGoalList(goals)  // skolemizenp equivalent
   goals = l2s.L2STacticFull(prover, goals, l2sPf)
   ```

3. **Port Tempind** (separate task): The `tempind` tactic from `ivy_tactics.py` needs porting to `goivy/tactics/`. This is a dependency for full temporal verification but the other parts can proceed without it.

**Python reference:** `ivy_check.py:805-831`

---

### Group C: MatchHandler Model Integration (Audit 9.1 #8)
**Complexity: Medium** | **File: `check/helpers.go:100-227`**

**Problem:** Two stubs in MatchHandler:

1. **`NewMatchHandler` (line 133-138):** TODO comment says "when solver.ClausesModelToClauses is available" — but it IS available at `solver/model.go:403`. Need to call it to populate `h.Eqs` with ground equalities from the model.

2. **`Eval` (line 162-164):** Always returns true. `solver.HerbrandModel.EvalToConstant()` exists at `solver/herbrand.go:216`. Need to evaluate condition against model and return boolean result.

**Fix for NewMatchHandler:**
```go
// Need a solver instance to call ClausesModelToClauses
// The model passed in should be a *solver.ModelResult or *solver.HerbrandModel
if slvModel, ok := model.(*solver.HerbrandModel); ok {
    modClauses, err := slvModel.Solver.ClausesModelToClauses(clauses.(*clauseops.Clauses), nil)
    if err == nil {
        for _, fmla := range modClauses.Formulas() {
            // Python: if is_eq: eqs[lhs.rep].append(fmla)
            // elif is_not: eqs[app.rep].append(Equals(app, Or()))
            // elif is_app: eqs[fmla.rep].append(Equals(fmla, And()))
            // Parse equality/negation/application patterns
        }
    }
}
```

**Fix for Eval:**
```go
func (h *MatchHandler) Eval(cond lg.Expr) bool {
    if model, ok := h.Model.(*solver.HerbrandModel); ok {
        result := model.EvalToConstant(cond)
        return lg.IsTrue(result)  // or !lg.IsFalse(result)
    }
    return true // fallback
}
```

**Prerequisite:** Need to type-narrow `MatchHandler.Model` from `interface{}` to a concrete type. Consider changing the field type to `*solver.HerbrandModel` or adding a `trace.Model` interface.

**Python reference:** `ivy_check.py:281-364`

---

### Group D: Trace Path in check_fcs_in_state (Audit 9.1 #10, 9.3 #4)
**Complexity: Medium** | **File: `check/check.go:484-551`**

**Problem:** The trace path in `checkFcsTracePath` doesn't use the `trace.Trace` type or `actions.MatchAnnotation`. It just prints action names. The Python version builds a proper `ivy_trace.Trace`, calls `match_annotation` to walk annotations, and either launches GUI or prints trace.

**All infrastructure exists:**
- `trace.NewTrace()` at `trace/trace.go:442`
- `actions.MatchAnnotation()` at `actions/match.go:33`
- `clauseops.UsedSymbols*` for vocab extraction

**Fix:** Replace the simplified trace output (lines 530-547) with:
```go
vocab := clauseops.UsedSymbolsClauses(mclauses)
handler := trace.NewTrace(mclauses, model, vocab, true)
thing := failed[len(failed)-1].GetAnnot()
if thing == nil {
    // Build sequence from history actions
    var actionSlice []actions.Action
    for _, a := range history.Actions {
        // resolve string->action via mod.Actions
    }
    action := actions.NewSequence(actionSlice...)
    annot := clauses.Annot
    actions.MatchAnnotation(action, annot, handler, mod)
} else {
    action, annot := thing.(ActionAnnotPair)
    actions.MatchAnnotation(action, annot, handler, mod)
}
handler.End()
```

Also need to handle `mod.TraceHook` if present, and the `IsCti` field for conjecture checkers.

**Python reference:** `ivy_check.py:373-416`

---

### Group E: Start() Entry Point (Audit 9.1 #11)
**Complexity: Low** | **File: `check/check.go:1021-1028`**

**Problem:** `Start()` returns error "not yet fully integrated". Needs to parse args, load file via ivyinit, call `CheckModule`.

**Fix:** `ivyinit.SourceFile` exists at `ivyinit/ivyinit.go:145` with signature `(filename string, mod *module.Module, sig *il.Sig, kwargs map[string]interface{}) error`. Wire up:
```go
func Start(args []string) error {
    if len(args) < 1 || !strings.HasSuffix(args[0], ".ivy") {
        return fmt.Errorf(Usage())
    }
    mod := module.New()
    sig := il.NewSig()
    kwargs := map[string]interface{}{"create_isolate": false}
    if err := ivyinit.SourceFile(args[0], mod, sig, kwargs); err != nil {
        return err
    }
    return CheckModule(mod)
}
```

Also handle the Python logic for `checked_assert` == none.ivy:0 -> "NOT CHECKED" exit, and the `some_bounded`/"OK"/"OK, but used 'sorry'" epilogue.

**Python reference:** `ivy_check.py:975-999`

---

### Group F: ConjChecker.GetAnnot (Audit 9.1 #8 sub-item)
**Complexity: Low** | **File: `check/check.go:193-196`**

**Problem:** `ConjChecker.GetAnnot()` returns nil with comment "annotations not yet ported". In Python, it returns `lf.annot` which provides the (action, annotation) pair needed by the trace path (Group D).

**Fix:** Return the annotation from the labeled formula:
```go
func (c *ConjChecker) GetAnnot() interface{} {
    if c.LF != nil && c.LF.Annot != nil {
        return c.LF.Annot
    }
    return nil
}
```

This is closely connected to Group D (trace path) — the annotation returned here feeds into `MatchAnnotation`.

---

### Group H: CheckModule macro_finder Save/Restore (Audit 9.3 #3)
**Complexity: Low** | **File: `check/isolate_check.go:726+`**

**Problem:** Python saves/restores `macro_finder` setting around each isolate check (lines 929-935, 969-972). The Go `CheckModule` has no reference to `macro_finder` at all. When an isolate has `macro_finder` attribute set, Python disables it for that isolate's check and restores it afterward.

**Fix:** Add macro_finder save/restore logic in the isolate loop:
```go
// Before isolate check:
saveMacroFinder := false
if isolate != "" {
    attrKey := ivyutils.ComposeNames(isolate, "macro_finder")
    if _, ok := mod.Attributes[attrKey]; ok {
        saveMacroFinder = mod.Cfg.MacroFinder
        if saveMacroFinder {
            fmt.Println("Turning off macro_finder")
            mod.Cfg.MacroFinder = false
        }
    }
}
// ... check isolate ...
// After:
if saveMacroFinder {
    fmt.Println("Turning on macro_finder")
    mod.Cfg.MacroFinder = true
}
```

**Python reference:** `ivy_check.py:929-935, 969-972`

---

## Execution Order (by dependency and value)

| Phase | Groups | Rationale |
|-------|--------|-----------|
| 1 | **F** (GetAnnot), **H** (macro_finder) | Quick fixes, no dependencies, unblock Group D |
| 2 | **C** (MatchHandler model) | Enables proper trace output |
| 3 | **D** (trace path wiring) | Depends on F and C; high correctness impact |
| 4 | **A** (CheckTemporals) | Core verification gap; all infra exists |
| 5 | **B** (MC/VMT tactics) | Depends on A; may be blocked on l2s port |
| 6 | **E** (Start entry point) | Low priority; users call CheckModule directly |

## Files to Modify

| File | Groups |
|------|--------|
| `check/check.go` | A, B, E, F |
| `check/helpers.go` | C |
| `check/isolate_check.go` | H |
| `check/phase7.go` | (none — G is deferred) |

## Key Reusable Functions (already ported)

| Function | Location | Used By |
|----------|----------|---------|
| `proof.NormalizeGoal` | `proof/goal.go:92` | Group A |
| `temporal.NormalProgramFromModule` | `temporal/temporal.go:225` | Group A |
| `proof.ProofChecker.AdmitProposition` | `proof/checker.go:388` | Group A |
| `solver.Solver.ClausesModelToClauses` | `solver/model.go:403` | Group C |
| `solver.HerbrandModel.EvalToConstant` | `solver/herbrand.go:216` | Group C |
| `trace.NewTrace` | `trace/trace.go:442` | Group D |
| `actions.MatchAnnotation` | `actions/match.go:33` | Group D |
| `clauseops.AndClausesTyped` | clauseops package | Group D |

## Verification

1. **Existing tests:** `go test ./check/...` must pass after each phase
2. **Integration test:** Create a small `.ivy` file with:
   - A property that can be checked (tests Group A path)
   - A conjecture with a proof (tests ApplyConjProofs + CheckSubgoals)
   - A failing assertion (tests trace output, Groups C/D)
3. **Temporal test:** If l2s tactics are available, test a temporal property with `proof mc` annotation
4. **Manual verification:** Run `go vet ./check/...` and confirm no new warnings
