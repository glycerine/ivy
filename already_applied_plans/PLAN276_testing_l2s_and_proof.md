# Reverse-Conformance Audit: proof/goal.go and check/l2s.go

**Created:** 2026-04-12 ~afternoon

## Context

`proof/goal.go` and `check/l2s.go` have been a recurring source of non-conformance bugs. A line-by-line reverse audit against the Python originals (`ivy_proof.py` and `ivy_l2s.py`) identified **7 confirmed divergences** where Go behavior differs from Python.

The audit also confirmed that many functions initially suspected as missing (e.g., `l2sAutoInvariants`, trace hooks, shared steps, `GoalSubgoals`, `CheckNameClash`, `SkolemizeGoal`, etc.) DO exist in companion Go files (`l2s_auto.go`, `l2s_hooks.go`, `l2s_shared.go`, `phase5_goals.go`, `phase5_matching.go`, `skolem.go`). The core l2s flow (entry points, monitor construction, fair cycle, tableau, action instrumentation, named binder replacement, goal building) matches Python faithfully modulo these 7 bugs.

---

## Bug 1: GoalIsDefn missing UninterpretedSort

**Files:** `proof/goal.go:178-192`
**Python:** `ivy_proof.py:607-610`

Python returns `True` for `UninterpretedSort`:
```python
def goal_is_defn(x):
    if isinstance(x,ia.ConstantDecl):
        return not il.is_lambda(x.args[0])
    return isinstance(x,il.UninterpretedSort)   # ← Go misses this
```

Go returns `false`:
```go
func GoalIsDefn(x ast.Node) bool {
    if cd, ok := x.(*ast.ConstantDecl); ok { ... return true }
    return false  // BUG: should also check *lg.UninterpretedSort
}
```

**Fix:** Add `if _, ok := x.(*lg.UninterpretedSort); ok { return true }` before `return false`.

**Impact:** `NormalizeGoal` and `GoalSubgoals` (via `GetUnprovidedDefns`) call this. Wrong result causes definitions to be unnecessarily normalized or missed during schema instantiation.

**No signature change. No caller updates needed.**

---

## Bug 2: GoalSubst missing CheckNameClash call

**Files:** `proof/goal.go:345-350`
**Python:** `ivy_proof.py:506-508`

Python validates before substituting:
```python
def goal_subst(g1,g2,lineno):
    check_name_clash(g1,g2)               # ← Go skips this
    return make_goal(lineno, g2.label, goal_prems(g1) + goal_prems(g2), goal_conc(g2))
```

Go:
```go
func GoalSubst(cfg *ast.AstConfig, g1, g2 *ast.LabeledFormula, loc ast.Location) *ast.LabeledFormula {
    prems := append(GoalPrems(g1), GoalPrems(g2)...)
    return MakeGoal(cfg, loc, g2.Label, prems, GoalConc(g2))
}
```

**Fix:** Change signature to return `(*ast.LabeledFormula, error)`. Call `CheckNameClash(g1, g2)` (already exists at `proof/phase5_goals.go:83-92`) at the top.

**Callers to update (4):**
- `proof/matching.go:197` — `ApplySchemaToGoal`
- `proof/phase5_goals.go:50` — `GoalsSubst`
- `proof/phase5_goals.go:184` — `GoalSubgoals`
- `proof/phase5_goals.go:198` — `GoalSubgoals`

Each caller already returns error (or can be made to), so propagation is contained.

---

## Bug 3: Desugar missing parameter validation for was/happened

**Files:** `check/l2s.go:956-988`
**Python:** `ivy_l2s.py:726-733`

Python validates that `was` and `happened` binders have no parameters:
```python
if expr.name == 'was':
    if len(expr.variables) > 0:
        raise iu.IvyError(expr,"operator 'was' does not take parameters")
```

Go silently ignores parameterized was/happened.

**Fix:** Change `Desugar` signature to `func Desugar(expr lg.Expr, proofLabel string) (lg.Expr, error)`. Add checks for `len(nb.Variables) > 0` returning error for both `"was"` and `"happened"` cases.

**Callers to update (3):**
- `check/l2s.go:478` — desugar loop in `l2sTacticInt`
- `check/ranking.go:274` — ranking desugarFn closure
- `check/l2s.go:978` — recursive self-call

---

## Bug 4: GoalVocab missing `bound` parameter

**Files:** `proof/goal.go:220-283`
**Python:** `ivy_proof.py:572-583`

Python supports `bound=True` to include bound variables from the conclusion:
```python
def goal_vocab(goal, bound=False):
    ...
    if bound:
        conc_fmla = conc.fmla if isinstance(conc,ia.TemporalModels) else conc
        variables = variables + [x for x in logic_util.bound_variables(conc_fmla)
                                 if x not in variables]
```

Go has no equivalent.

**Fix:** Add `GoalVocabBound(goal *ast.LabeledFormula) *Vocab` that calls `GoalVocab` then appends bound variables from conclusion using `lu.BoundVariables` (verify this exists or port `logic_util.bound_variables`).

**Callers to update (1):**
- `tactics/ivy_tactics.go:178` — has comment `// Python: vocab = pr.goal_vocab(goal, bound=True)` but currently calls `GoalVocab` without bound support. Switch to `GoalVocabBound`.

---

## Bug 5: GoalDefns accepts any Expr instead of only Const/Symbol

**Files:** `proof/goal.go:196-214`
**Python:** `ivy_proof.py:540-547`

Python checks specifically for `il.Symbol` (= Go's `*lg.Const`):
```python
if isinstance(x,ia.ConstantDecl) and isinstance(x.args[0],il.Symbol):
    res.add(x.args[0])
```

Go accepts any `lg.Expr`:
```go
if c, ok := args[0].(lg.Expr); ok {
    res[lg.Key(c)] = c
}
```

**Fix:** Narrow the type assertion from `lg.Expr` to `*lg.Const`:
```go
if c, ok := args[0].(*lg.Const); ok {
    res[lg.Key(c)] = c
}
```

**No signature change. No caller updates needed.**

---

## Bug 6: l2s model source — Go rebuilds from module instead of cloning from goal

**Files:** `check/l2s.go:304`
**Python:** `ivy_l2s.py:121`

Python clones the NormalProgram embedded in the goal's TemporalModels:
```python
conc = ipr.goal_conc(goal)
model = conc.model.clone([])    # ← clone from the goal's embedded model
```

Go ignores `tm.Model` and rebuilds from the module:
```go
model := extractNormalProgram(m)   // ← rebuilds from module.Module
```

Even though both originate from `normal_program_from_module`, the behavioral difference is:
- **Ordering**: `NormalProgramFromModule` sorts bindings alphabetically (line 236-238 of `temporal.go`). The model embedded in the goal was built at a different time and may have a different binding order.
- **Content**: The module could have been mutated between goal construction and l2s execution. The Python clone preserves the snapshot from goal-construction time.
- **Trace output**: Different binding order = different action instrumentation order = different trace output.

**Fix:** Replace `extractNormalProgram(m)` with clone from `tm.Model`:
```go
np, ok := tm.Model.(*temporal.NormalProgram)
if !ok {
    return nil, fmt.Errorf("l2s: TemporalModels.Model is not a NormalProgram")
}
model := temporal.NormalProgramClone(np)
```

`NormalProgramClone` already exists at `temporal/temporal.go:399` and does a shallow copy matching Python's `clone([])`.

**No signature change. 1 line replacement in l2s.go:304. Can remove `extractNormalProgram` helper if no other callers.**

---

## Bug 7: FreeVariablesList returns alphabetical order, not DFS first-occurrence order

**Files:** `check/l2s.go:963, 1011` and `logicutil/logicutil.go:617`
**Python:** `ivy_l2s.py:721, 724` using `ilu.variables_ast` + `iu.unique`

Python's `variables_ast` (`ivy_logic_utils.py:474-486`) yields free variables in **DFS tree-traversal order**. Wrapped with `iu.unique()`, it produces unique variables in **first-occurrence** order.

Go's `FreeVariablesList` (`logicutil.go:617`) uses `FreeVariables` which stores in an `Omap` backed by a red-black tree. `Omap.All()` iterates in **sorted NodeKey order** (alphabetical), not DFS order.

**Example:** For expression `And(p(Y), q(X))`:
- Python: `list(iu.unique(ilu.variables_ast(expr)))` → `[Y, X]` (DFS order)
- Go: `lu.FreeVariablesList(expr)` → `[X, Y]` (alphabetical)

This directly affects named binder parameter order in `l2s_s(vs, expr)(*vs)` and `l2s_w(vs, expr)(*vs)`, which affects canon() output and trace conformance.

**Affected call sites in l2s.go:**
- `l2s.go:963` — `Desugar` "happened" case: `vs := lu.FreeVariablesList(nb.Body)`
- `l2s.go:1011` — `applyWasRec` base case: `vs := lu.FreeVariablesList(expr)`

**Affected call sites in l2s_auto.go** (Python lines 237, 586, 595, 672 use `ilu.variables_ast`):
- Need to grep l2s_auto.go for corresponding Go calls.

**Fix:** Add `VariablesAstList` to `logicutil/logicutil.go` — a DFS free-variable collector that preserves first-occurrence order:
```go
// VariablesAstList returns the free variables of t in DFS first-occurrence
// order, matching Python's list(iu.unique(ilu.variables_ast(t))).
func VariablesAstList(t logic.Expr) []*logic.Variable {
    var result []*logic.Variable
    seen := make(map[logic.NodeKey]bool)
    variablesAstRec(t, &result, seen, nil)
    return result
}

func variablesAstRec(t logic.Expr, result *[]*logic.Variable, seen map[logic.NodeKey]bool, bound map[logic.NodeKey]bool) {
    switch n := t.(type) {
    case *logic.Variable:
        k := logic.Key(n)
        if !bound[k] && !seen[k] {
            seen[k] = true
            *result = append(*result, n)
        }
    case *logic.ForAll:
        newBound := copyBoolKeySet(bound)
        for _, v := range n.Variables { newBound[logic.Key(v)] = true }
        variablesAstRec(n.Body, result, seen, newBound)
    case *logic.Exists:
        newBound := copyBoolKeySet(bound)
        for _, v := range n.Variables { newBound[logic.Key(v)] = true }
        variablesAstRec(n.Body, result, seen, newBound)
    case *logic.Lambda:
        newBound := copyBoolKeySet(bound)
        for _, v := range n.Variables { newBound[logic.Key(v)] = true }
        variablesAstRec(n.Body, result, seen, newBound)
    case *logic.NamedBinder:
        newBound := copyBoolKeySet(bound)
        for _, v := range n.Variables { newBound[logic.Key(v)] = true }
        variablesAstRec(n.Body, result, seen, newBound)
    default:
        for _, c := range t.Children() {
            variablesAstRec(c, result, seen, bound)
        }
    }
}
```

Then replace `lu.FreeVariablesList` with `lu.VariablesAstList` at the affected call sites.

---

## Implementation Order

1. **Bug 5** — one-line type narrowing, zero cascade
2. **Bug 1** — add one branch, zero cascade
3. **Bug 6** — one-line replacement in l2s.go
4. **Bug 7** — new function in logicutil + call site replacements
5. **Bug 4** — new function + one caller switch
6. **Bug 3** — signature change, 3 callers
7. **Bug 2** — signature change, 4 callers with possible cascading

---

## Verification

1. `go build ./...` — ensure all changes compile
2. `go test ./proof/... ./check/... ./tactics/... ./logicutil/...` — run existing tests
3. Run any existing golden test vectors in `test_vectors/` that exercise proof/l2s paths
4. For Bug 1: verify that `GoalIsDefn(anUninterpretedSort)` now returns `true`
5. For Bug 2: verify that `GoalSubst` with clashing premise names returns an error
6. For Bug 3: verify that `Desugar` with `was(X)` (parameterized) returns an error
7. For Bug 4: verify that `GoalVocabBound` includes bound variables from conclusion
8. For Bug 5: verify that `GoalDefns` only collects `*lg.Const` premises, not other Expr types
9. For Bug 6: verify l2s uses `tm.Model` clone, not module rebuild. Check trace output matches Python.
10. For Bug 7: verify `VariablesAstList(And(p(Y), q(X)))` returns `[Y, X]` not `[X, Y]`

---

## Files to Modify

| File | Bugs |
|------|------|
| `proof/goal.go` | 1, 2, 4, 5 |
| `check/l2s.go` | 3, 6, 7 |
| `check/ranking.go` | 3 (caller) |
| `logicutil/logicutil.go` | 7 (new function) |
| `proof/matching.go` | 2 (caller) |
| `proof/phase5_goals.go` | 2 (caller) |
| `tactics/ivy_tactics.go` | 4 (caller) |

---

---

# Phase 2: Companion File Audit (l2s_auto.go, l2s_shared.go, l2s_hooks.go)

**Created:** 2026-04-12 ~evening

Bugs 1-7 were implemented in Phase 1. The three companion files now audited yield 4 additional bugs. NOTE: The agent initially flagged `collectVarsSlice` (l2s_auto.go) as using alphabetical order — this was INCORRECT. `modpkg.VariablesAST` → `collectFreeVarsOrdered` already uses DFS first-occurrence order, equivalent to `lu.VariablesAstList`.

---

## Bug 8: l2s_not_all_done emitted conditionally — Python always emits

**Files:** `check/l2s_auto.go:710-712`
**Python:** `ivy_l2s.py:567-568`

Python ALWAYS emits the invariant (empty `Or()` = false):
```python
tmp = lg.Or(*not_all_done_preds)
invars.append(ivy_ast.LabeledFormula(ivy_ast.Atom("l2s_not_all_done"),tmp).sln(proof.lineno))
```

Go conditionally skips it:
```go
if len(notAllDonePreds) > 0 {
    invars = appendLF(autoAcfg, invars, "l2s_not_all_done", buildOrExpr(notAllDonePreds))
}
```

**Fix:** Remove the `if len(notAllDonePreds) > 0` guard. When empty, emit with `lg.False` (matching Python's empty Or → false).

**No signature change. No caller updates.**

---

## Bug 9: when.body structure mismatch in assume_when_axioms (Step 6)

**Files:** `check/l2s_shared.go:264-268`
**Python:** `ivy_l2s.py:1005-1008`

Python decomposes `when.body` as a `Cond` node (T1=condition, T2=value):
```python
AssumeAction(forall(when.variables,
    lg.Implies(when.body.t1, lg.Eq(when(*when.variables), when.body.t2))))
```

Go uses `when.Body` (the entire Cond) as both T1 of Implies and T2 of Eq:
```go
inner := forall(when.Variables, &lg.Implies{
    T1: when.Body,                        // BUG: should be when.Body.(*lg.Cond).T1
    T2: &lg.Eq{T1: applyNB(...), T2: when.Body},  // BUG: should be .T2
})
```

When binders are created at `ivy_logic_utils.py:300-305`: the body is `lg.Cond(cond, val)`. Go's `lg.Cond` struct has `T1` and `T2` fields (`logic/formula.go:222-227`).

**Fix:** Decompose when.Body:
```go
cond, ok := when.Body.(*lg.Cond)
if ok {
    inner := forall(when.Variables, &lg.Implies{
        T1: cond.T1,
        T2: &lg.Eq{T1: applyNB(when, varsToNodes(when.Variables)...), T2: cond.T2},
    })
}
```

---

## Bug 10: when.body structure mismatch in whenEventsFunc (Step 7)

**Files:** `check/l2s_shared.go:360-390`
**Python:** `ivy_l2s.py:1070-1085`

Same root cause as Bug 9. Python line 1074-1075:
```python
name, vs, t = when.name, when.variables, when.body
cond, val = t.t1, t.t2
```

Go uses `when.Body` directly as `cond` (4 locations):
- Line 365: `cond := when.Body` in l2s_whennext — should be `when.Body.(*lg.Cond).T1`
- Line 374: `cond := when.Body` in l2s_whenprev — should be `when.Body.(*lg.Cond).T1`
- Line 385: `T1: when.Body` in post assume — should be `cond.T1`
- Line 386: `T2: when.Body` in post assume — should be `cond.T2`

**Fix:** At the top of whenEventsFunc, extract T1/T2 from each when's body:
```go
for _, when := range whens {
    condVal, _ := when.Body.(*lg.Cond)
    cond := condVal.T1
    // val := condVal.T2  // (used in post assume)
    ...
}
```

---

## Bug 11: l2s_hooks.go diagnostic stubs — missing evaluation logic

**Files:** `check/l2s_hooks.go:72-128`
**Python:** `ivy_l2s.py:1364-1537`

Go's `diagnoseAutoFailure` has only stub `fmt.Printf` messages for 6 of 8 cases. Python provides detailed evaluation of predicates against trace states, printing "Note:" diagnostic messages. The missing pieces are:

**A. Missing helpers (port from Python):**
1. `temporal_and_l2s(sym)` — Python line 1352-1354: `return sym.name.startswith('l2s') and not sym.name.startswith('l2s_g') or sym.name.startswith('_old_l2s')`
2. `ls2_g_to_globally(ast)` — Python lines 1356-1362: expand l2s_g named binders to Globally operators

**B. Missing infrastructure integration:**
3. `applyAutoDiagnosticsToHandler` needs access to the trace/MatchHandler's equation map to evaluate symbols. The MatchHandler already has `Eqs map[lg.NodeKey][]lg.Expr` — this corresponds to Python's iteration over `state.clauses.fmlas`.
4. Justice predicate map extraction (Python lines 1379-1388) — extract justice predicates from l2s_progress_invar formulas.
5. `hidden_symbols` support — add a `HiddenSymbols func(string) bool` field to MatchHandler, set it in each diagnostic case.

**C. Full diagnostic cases to port (6 cases):**
- `l2s_created` (Python 1393-1405): extract work_created, evaluate Skolem symbols, print Note
- `l2s_needed_when_start` (Python 1407-1423): extract work_needed + work_created, evaluate, print Note
- `l2s_work_preserved` (Python 1425-1437): extract work_needed, evaluate, print Note
- `l2s_needed_are_frozen` (Python 1439-1451): same pattern as work_preserved
- `l2s_progress_made` (Python 1453-1509): 56 lines — extract helpful/trigger maps, justice conditions, two evaluation loops
- `l2s_sched_stable` (Python 1511-1525): extract work_progress + work_helpful, evaluate

Each case follows the same pattern:
1. Extract the relevant task predicate definition from `tasks[sfx]`
2. Get variables from the predicate's LHS args
3. Create Skolem symbols `@v.name` for each variable
4. Evaluate Skolem symbols in the handler's model (use `handler.Eqs` to look up values)
5. If all values resolved, build the predicate application and print a "Note:" message
6. Set `handler.HiddenSymbols = temporal_and_l2s`

**Fix approach:** Change `diagnoseAutoFailure` signature to accept `*MatchHandler` (it currently only gets name/tasks/triggers/lf). Implement each case by looking up equations in `handler.Eqs`, add `HiddenSymbols` field to MatchHandler, port `temporal_and_l2s` as a method.

---

## Implementation Order (Phase 2)

1. **Bug 8** — remove conditional guard, 1 line
2. **Bug 9** — decompose when.Body in assume_when_axioms
3. **Bug 10** — decompose when.Body in whenEventsFunc (4 locations)
4. **Bug 11A** — add `temporal_and_l2s` and `ls2_g_to_globally` helpers
5. **Bug 11B** — add `HiddenSymbols` field to MatchHandler, pass handler to diagnoseAutoFailure
6. **Bug 11C** — port 6 diagnostic cases with evaluation logic

---

## Verification (Phase 2)

1. `go build ./...` — ensure all changes compile
2. `go test ./check/... ./logicutil/...` — run existing tests
3. For Bug 8: verify l2s_not_all_done is always emitted (even with empty predicates)
4. For Bugs 9-10: verify assume_when_axioms and when events correctly decompose Cond body
5. For Bug 11: verify diagnostic output matches Python format for each case

---

## Files to Modify (Phase 2)

| File | Bugs |
|------|------|
| `check/l2s_auto.go` | 8 |
| `check/l2s_shared.go` | 9, 10 |
| `check/l2s_hooks.go` | 11 (full rewrite of diagnoseAutoFailure) |
| `check/helpers.go` | 11B (add HiddenSymbols field to MatchHandler) |

---

# Phase 3: Unit Tests and Fuzz Tests for l2s*.go and goal.go

**Created:** 2026-04-12 ~late evening

## Context

Bugs 1-11 are implemented but have no unit tests. The goal.go functions have only skeletal coverage (basic happy-path) in `proof/proof_test.go`. The check/l2s*.go files have zero test coverage. This phase adds comprehensive unit tests and fuzz tests that:
1. Serve as regression tests for Bugs 1-11
2. Exercise all pure/testable functions
3. Provide fuzz coverage for functions with tricky input space

## Guiding Principle

**The point of tests is to find bugs. If a test uncovers a bug in production code, fix the production code — never weaken or "fix" the test to make it pass. Then strengthen tests in that area, since a bug there signals a weak spot needing more coverage.**

## Existing test infrastructure

| File | Package | Helpers available |
|------|---------|-------------------|
| `proof/proof_test.go` | proof | `mkSort`, `mkVar`, `mkConst`, `mkLF`, `testAstCfg` |
| `check/check_port_test.go` | check | `testAstCfg` |
| `check/ranking_test.go` | check | `boolVar`, `boolConst`, `unintSort` |

Uses standard `testing` package. Fuzz patterns from `proof/proof_test.go` (FuzzMergeMatches, FuzzMatch).

---

## New Files

| File | Tests |
|------|-------|
| `proof/goal_test.go` | goal.go functions (Bugs 1, 2, 4, 5 regression + coverage) |
| `check/l2s_test.go` | l2s.go pure functions (Bug 3 regression + helpers) |
| `check/l2s_auto_test.go` | l2s_auto.go pure functions (Bug 8 related) |
| `check/l2s_hooks_test.go` | l2s_hooks.go pure functions (Bug 11 related) |

---

## File 1: `proof/goal_test.go`

Reuses `mkSort`, `mkVar`, `mkConst`, `mkLF`, `testAstCfg` from `proof_test.go` (same package).

### A. GoalIsDefn — Bug 1 regression

| Test | Input | Expected |
|------|-------|----------|
| `TestGoalIsDefn_ConstantDecl` | `cfg.NewConstantDecl(mkConst("f", S))` | `true` |
| `TestGoalIsDefn_ConstDeclLambda` | `cfg.NewConstantDecl(lambda)` where lambda is `*lg.Lambda` | `false` |
| `TestGoalIsDefn_UninterpretedSort` | `&lg.UninterpretedSort{Name: "S"}` | `true` (**Bug 1**) |
| `TestGoalIsDefn_PlainConst` | `mkConst("c", S)` | `false` |
| `TestGoalIsDefn_ConstDeclNoArgs` | `cfg.NewConstantDecl()` (empty args) | `true` |

### B. GoalDefns — Bug 5 regression

| Test | Premises | Expected |
|------|----------|----------|
| `TestGoalDefns_ConstPrem` | ConstantDecl(`*lg.Const`) | map contains it |
| `TestGoalDefns_VariablePrem` | ConstantDecl(`*lg.Variable`) | map does NOT contain it (**Bug 5**) |
| `TestGoalDefns_UninterpSortPrem` | `*lg.UninterpretedSort` as premise | map contains it |
| `TestGoalDefns_NoPremises` | no SchemaBody | empty map |

Construction: `mkLF(label, cfg.NewSchemaBody(prem, conc))`.

### C. GoalSubst — Bug 2 regression

| Test | Setup | Expected |
|------|-------|----------|
| `TestGoalSubst_OK` | g1 has prem `ConstantDecl(f)`, g2 has prem `ConstantDecl(g)` (no clash) | result non-nil, err nil, combined prems |
| `TestGoalSubst_NameClash` | g1 and g2 both have `ConstantDecl(f)` with same key | err != nil (**Bug 2**) |

### D. GoalVocabBound — Bug 4 regression

| Test | Conclusion | Expected |
|------|------------|----------|
| `TestGoalVocabBound_IncludesBound` | `ForAll{X, body}` | `Variables` includes X |
| `TestGoalVocabBound_NoDuplicates` | Free X in prem + bound X in conc | X appears once |
| `TestGoalVocabBound_TemporalModels` | `&ast.TemporalModels{Fmla: ForAll{X, body}}` | includes X (**Bug 4**) |
| `TestGoalVocabBound_NilConc` | Non-Expr conclusion | same as `GoalVocab` |

### E. ConcAsExpr / GoalConcUnwrap

| Test | Input | Expected |
|------|-------|----------|
| `TestConcAsExpr_Expr` | `mkConst("c", S)` | returns the Const |
| `TestConcAsExpr_TemporalModels` | `&ast.TemporalModels{Fmla: mkConst("c", S)}` | returns inner Const |
| `TestConcAsExpr_NonExpr` | `testAstCfg.NewAtom("a")` | returns nil |
| `TestGoalConcUnwrap` | goal with TemporalModels conclusion | returns inner Expr |

### F. ApplyToConc / GoalApplyToConc

| Test | Input | Expected |
|------|-------|----------|
| `TestApplyToConc_PlainExpr` | Const, fn wraps in Not | `*lg.Not{Body: Const}` |
| `TestApplyToConc_TemporalModels` | TM wrapping Const, fn wraps in Not | TM wrapping `*lg.Not{Body: Const}` |
| `TestApplyToConc_NonExpr` | non-Expr node, fn wraps in Not | node unchanged |
| `TestGoalApplyToConc` | goal with Const conc, fn adds Not | cloned goal with Not conclusion |

### G. GoalAddPrem / GoalRemovePrem

| Test | Expected |
|------|----------|
| `TestGoalAddPrem_Empty` | premless goal → 1 prem |
| `TestGoalAddPrem_Existing` | goal with 1 prem → 2 prems |
| `TestGoalRemovePrem_Found` | remove named prem → gone |
| `TestGoalRemovePrem_NotFound` | remove non-existent → unchanged |

### H. TrivialGoal positive case

| Test | Setup | Expected |
|------|-------|----------|
| `TestTrivialGoal_Positive` | Build goal where prem has same conc as outer | `true` |

Construction: inner prem goal `mkLF(label, sameConc)` wrapped in SchemaBody.

### I. Fuzz

| Target | Input | Invariant |
|--------|-------|-----------|
| `FuzzGoalIsDefn` | random node type choice | no panic |
| `FuzzGoalDefns` | random prem type combos | no panic, map values are either `*lg.Const` or `*lg.UninterpretedSort` |

---

## File 2: `check/l2s_test.go`

New helpers: `var l2sTestCfg = ast.NewAstConfig()`.
Reuses `boolVar`, `boolConst`, `unintSort` from `ranking_test.go` (same package).

### A. Named constant constructors

| Test | Func | Check name and sort |
|------|------|---------------------|
| `TestL2SWaiting` | `L2SWaiting()` | Name="l2s_waiting", CSort=Boolean |
| `TestL2SFrozen` | `L2SFrozen()` | Name="l2s_frozen", CSort=Boolean |
| `TestL2SSaved` | `L2SSaved()` | Name="l2s_saved", CSort=Boolean |
| `TestL2SD` | `L2SD(s)` | Name="l2s_d", sort is function sort from s→Boolean |
| `TestL2SA` | `L2SA(s)` | Name="l2s_a", sort is function sort from s→Boolean |

### B. Named binder constructors (unexported, within-package tests)

| Test | Func | Check Name field |
|------|------|-----------------|
| `TestL2sW` | `l2sW(vs, body, "lbl")` | Name="l2s_w" |
| `TestL2sS` | `l2sS(vs, body, "lbl")` | Name="l2s_s" |
| `TestL2sG` | `l2sG(vs, body, env)` | Name="l2s_g" |
| `TestOldL2sG` | `oldL2sG(vs, body, env)` | Name="_old_l2s_g" |
| `TestL2sInit` | `l2sInit(vs, body, "lbl")` | Name="l2s_init" |
| `TestL2sWhen` | `l2sWhen("foo", vs, body, "lbl")` | Name="l2s_whenfoo" |
| `TestL2sOld` | `l2sOld(vs, body, "lbl")` | Name="l2s_old" |

### C. Pure helpers

| Test | Func | Input → Expected |
|------|------|-----------------|
| `TestApplyNB_NoArgs` | `applyNB(nb)` | returns nb itself |
| `TestApplyNB_WithArgs` | `applyNB(nb, x)` | returns `*lg.Apply` |
| `TestVarsToNodes_Empty` | `varsToNodes(nil)` | empty slice |
| `TestVarsToNodes` | `varsToNodes([x, y])` | `[]lg.Expr{x, y}` |
| `TestForall_Empty` | `forall(nil, body)` | body |
| `TestForall_WithVars` | `forall([x], body)` | `*lg.ForAll` |
| `TestExists_Empty` | `exists(nil, body)` | body |
| `TestExists_WithVars` | `exists([x], body)` | `*lg.Exists` |
| `TestMakeAnd_Zero` | `makeAnd()` | `lg.True` |
| `TestMakeAnd_One` | `makeAnd(a)` | `a` |
| `TestMakeAnd_Multi` | `makeAnd(a, b)` | `*lg.And{[a, b]}` |

### D. dedupeVarBodyPairs

| Test | Input | Expected len |
|------|-------|-------------|
| `TestDedupeVarBodyPairs_NoDupes` | 3 unique pairs | 3 |
| `TestDedupeVarBodyPairs_WithDupes` | 3 pairs, 2 identical | 2 |
| `TestDedupeVarBodyPairs_Empty` | nil | 0 |

### E. findTemporalModels

| Test | Input | Expected |
|------|-------|----------|
| `TestFindTemporalModels_Nil` | nil goal | nil |
| `TestFindTemporalModels_DirectTM` | goal.Formula = TemporalModels | returns TM |
| `TestFindTemporalModels_SchemaTM` | goal.Formula = SchemaBody, last elem = TM | returns TM |
| `TestFindTemporalModels_NoTM` | goal.Formula = Const | nil |

### F. String helpers

| Test | Func | Expected |
|------|------|----------|
| `TestCreateSavedCopy` | `CreateSavedCopy("foo")` | `"l2s_saved_foo"` |
| `TestIsSavedSymbol_True` | `IsSavedSymbol("l2s_saved_x")` | true |
| `TestIsSavedSymbol_False` | `IsSavedSymbol("l2s_x")` | false |
| `TestIsL2SSymbol_True` | `IsL2SSymbol("l2s_foo")` | true |
| `TestIsL2SSymbol_False` | `IsL2SSymbol("foo")` | false |
| `TestPyBool` | `pyBool(true)`, `pyBool(false)` | `"True"`, `"False"` |

### G. Desugar — Bug 3 regression

| Test | Input | Expected |
|------|-------|----------|
| `TestDesugar_Was_OK` | NB name="was", no vars, body=Const | `(And [l2s_saved, l2s_s([], body)()])`, err nil |
| `TestDesugar_Was_Error` | NB name="was", vars=[v] | err contains "does not take parameters" (**Bug 3**) |
| `TestDesugar_Happened_OK` | NB name="happened", no vars, body=Const | `(And [l2s_saved, Not(l2s_w(...))])`, err nil |
| `TestDesugar_Happened_Error` | NB name="happened", vars=[v] | err contains "does not take parameters" (**Bug 3**) |
| `TestDesugar_PlainConst` | `*lg.Const` | same Const back, err nil |
| `TestDesugar_NestedWas` | `And{NB(was, body)}` | recursion applies desugar to child |

### H. applyWasRec

| Test | Input | Expected output type |
|------|-------|---------------------|
| `TestApplyWasRec_And` | `And{a, b}` | `And{l2sS(a), l2sS(b)}` |
| `TestApplyWasRec_Or` | `Or{a, b}` | `Or{l2sS(a), l2sS(b)}` |
| `TestApplyWasRec_Not` | `Not{a}` | `Not{l2sS(a)}` |
| `TestApplyWasRec_Implies` | `Implies{a, b}` | `Implies{l2sS(a), l2sS(b)}` |
| `TestApplyWasRec_Leaf` | `Const` | Apply of l2s_s binder |

### I. transformAction

| Test | Input | Expected |
|------|-------|----------|
| `TestTransformAction_Nil` | nil | nil |
| `TestTransformAction_Simple` | action with lg.True, fn=negate | action with lg.Not{lg.True} |

### J. Fuzz

| Target | Seed | Invariant |
|--------|------|-----------|
| `FuzzDesugar` | (was, 0), (happened, 0), (was, 1), (other, 0) | no panic; vars>0 with was/happened returns error |
| `FuzzMakeAnd` | 0, 1, 5 | no panic; result type matches expected |

---

## File 3: `check/l2s_auto_test.go`

### A. buildOrExpr

| Test | Input | Expected |
|------|-------|----------|
| `TestBuildOrExpr_Empty` | `nil` | `*lg.Or{nil}` |
| `TestBuildOrExpr_Single` | `[a]` | `a` (not wrapped) |
| `TestBuildOrExpr_Multiple` | `[a, b]` | `*lg.Or{[a, b]}` |

### B. isGloballyOrNotEventually

| Test | Input | Expected |
|------|-------|----------|
| `TestIsGloballyOrNotEventually_Globally` | `*lg.Globally{body}` | true |
| `TestIsGloballyOrNotEventually_NotEventually` | `Not{Eventually{body}}` | true |
| `TestIsGloballyOrNotEventually_Eventually` | `*lg.Eventually{body}` | false |
| `TestIsGloballyOrNotEventually_Plain` | `*lg.Const` | false |

### C. isEventuallyOrNotGlobally

| Test | Input | Expected |
|------|-------|----------|
| `TestIsEventuallyOrNotGlobally_Eventually` | `*lg.Eventually{body}` | true |
| `TestIsEventuallyOrNotGlobally_NotGlobally` | `Not{Globally{body}}` | true |
| `TestIsEventuallyOrNotGlobally_Globally` | `*lg.Globally{body}` | false |
| `TestIsEventuallyOrNotGlobally_Plain` | `*lg.Const` | false |

### D. sameLHSSort

| Test | Input | Expected |
|------|-------|----------|
| `TestSameLHSSort_BothNil` | nil, nil | true |
| `TestSameLHSSort_SameSort` | Eq{Apply(f:S→Bool, x), _}, Eq{Apply(g:S→Bool, y), _} | true |
| `TestSameLHSSort_DiffSort` | Eq{Apply(f:S→Bool, x), _}, Eq{Apply(g:T→Bool, y), _} | false |
| `TestSameLHSSort_ConstLHS` | Eq{Const("c", S), _}, Eq{Const("d", S), _} | true |

### E. cloneLHS

| Test | Input | Expected |
|------|-------|----------|
| `TestCloneLHS_Const` | Const("f", S), "g" | Const("g", S) |
| `TestCloneLHS_Apply` | Apply(Const("f", fs), [x]), "g" | Apply(Const("g", fs), [x]) |
| `TestCloneLHS_Other` | Variable, "g" | same Variable unchanged |

### F. appendLF

| Test | Input | Expected |
|------|-------|----------|
| `TestAppendLF` | invars=nil, name="foo", fmla=True | len(result)==1, label name=="foo" |

---

## File 4: `check/l2s_hooks_test.go`

### A. TemporalAndL2S — table-driven

```go
tests := []struct{ name string; want bool }{
    {"l2s_waiting", true},
    {"l2s_frozen", true},
    {"l2s_saved", true},
    {"l2s_g", false},        // starts with l2s_g
    {"l2s_g_foo", false},    // starts with l2s_g
    {"_old_l2s", true},      // starts with _old_l2s
    {"_old_l2s_g", true},    // _old_l2s prefix matches
    {"foo", false},
    {"", false},
    {"l2s", true},           // "l2s" but not "l2s_g"
    {"l2", false},
    {"l2s_", true},          // "l2s_" but not "l2s_g"
    {"l2s_ga", false},       // starts with "l2s_g"
}
```

### B. applyRenamingToHandler

| Test | Handler | Subs | Expected |
|------|---------|------|----------|
| `TestApplyRenaming_Nil` | nil | any | no panic |
| `TestApplyRenaming_Empty` | Lines=["hello"] | {} | Lines=["hello"] |
| `TestApplyRenaming_Replace` | Lines=["val=_c42"] | {"_c42":"l2s_w(X)"} | Lines=["val=l2s_w(X)"] |
| `TestApplyRenaming_Multi` | Lines=["_c1 and _c2"] | {"_c1":"a", "_c2":"b"} | Lines=["a and b"] |

### C. predLHSArgs / predLHSRep

| Test | Eq.T1 | predLHSArgs | predLHSRep |
|------|-------|-------------|------------|
| Apply(f, [x, y]) | `[x, y]` | `f` |
| Const("c", S) | nil | `Const("c", S)` |
| nil pred | nil | nil |

### D. makeSkolems

| Test | Input | Expected |
|------|-------|----------|
| `TestMakeSkolems_Vars` | `[Variable("X", S)]` | `[Const("@X", S)]` |
| `TestMakeSkolems_Consts` | `[Const("c", S)]` | `[Const("@c", S)]` |
| `TestMakeSkolems_Empty` | nil | empty |

### E. exprNameSort

| Test | Input | Expected |
|------|-------|----------|
| `TestExprNameSort_Variable` | Variable("X", S) | ("X", S) |
| `TestExprNameSort_Const` | Const("c", S) | ("c", S) |
| `TestExprNameSort_Panic` | `*lg.And{}` | panics with "unhandled type" |

Use `defer func() { recover() }()` pattern for panic test.

### F. evalSkolemInHandler

| Test | Handler.Eqs | Skolem | Expected |
|------|-------------|--------|----------|
| `TestEvalSkolem_Found` | `{key: [Eq{sk, val}]}` | sk | val |
| `TestEvalSkolem_NotFound` | `{}` | sk | nil |
| `TestEvalSkolem_NilHandler` | nil handler | sk | nil |

### G. evalSkolems

| Test | Skolems → values | Expected |
|------|------------------|----------|
| `TestEvalSkolems_AllFound` | all present | `[]lg.Expr` with values |
| `TestEvalSkolems_MissingOne` | one absent | nil |

### H. applyPredToVals

| Test | Input | Expected |
|------|-------|----------|
| `TestApplyPredToVals_NoVals` | rep="f", vals=[] | `"f"` |
| `TestApplyPredToVals_WithVals` | rep="f", vals=[c1, c2] | `"f(c1,c2)"` |
| `TestApplyPredToVals_NilRep` | nil | `"<nil>"` |

### I. extractJusticePred

| Test | Formula | Expected |
|------|---------|----------|
| `TestExtractJP_ForAllImplies` | ForAll{v, Implies{_, Eq{Apply(J, _), _}}} | J |
| `TestExtractJP_NoMatch` | And{a, b} | nil |
| `TestExtractJP_NestedForAll` | ForAll{v, ForAll{w, Implies{_, Eq{Apply(J, _), _}}}} | J |

### J. lfName

| Test | Input | Expected |
|------|-------|----------|
| `TestLfName_Nil` | nil LF | `""` |
| `TestLfName_ConstLabel` | LF with Const("myname") label | `"myname"` |
| `TestLfName_NilLabel` | LF with nil label | `""` |

### K. Fuzz

| Target | Invariant |
|--------|-----------|
| `FuzzTemporalAndL2S` | no panic, returns bool |
| `FuzzExprNameSort` | Variable/Const don't panic; other types do |

---

## Implementation Order

1. **`check/l2s_hooks_test.go`** — all pure functions, no AST config needed for most
2. **`check/l2s_auto_test.go`** — pure functions, lightweight fixtures
3. **`check/l2s_test.go`** — pure helpers first, then Desugar (needs NamedBinder construction)
4. **`proof/goal_test.go`** — needs SchemaBody/ConstantDecl/TemporalModels fixtures

## Verification

```bash
go test ./proof/... ./check/... -v -count=1       # unit tests
go test ./proof/... ./check/... -fuzz=Fuzz -fuzztime=10s  # fuzz (per target)
go build ./...                                      # no build regressions
```

---

## Files to Create

| File | Package | Approx tests |
|------|---------|-------------|
| `proof/goal_test.go` | proof | ~25 unit + 2 fuzz |
| `check/l2s_test.go` | check | ~30 unit + 2 fuzz |
| `check/l2s_auto_test.go` | check | ~15 unit |
| `check/l2s_hooks_test.go` | check | ~25 unit + 2 fuzz |
