# Reverse-Conformance Audit: proof/goal.go and check/l2s.go

**Created:** 2026-04-12 ~afternoon

## Context

`proof/goal.go` and `check/l2s.go` have been a recurring source of non-conformance bugs. A line-by-line reverse audit against the Python originals (`ivy_proof.py` and `ivy_l2s.py`) identified **5 confirmed divergences** where Go behavior differs from Python.

The audit also confirmed that many functions initially suspected as missing (e.g., `l2sAutoInvariants`, trace hooks, shared steps, `GoalSubgoals`, `CheckNameClash`, `SkolemizeGoal`, etc.) DO exist in companion Go files (`l2s_auto.go`, `l2s_hooks.go`, `l2s_shared.go`, `phase5_goals.go`, `phase5_matching.go`, `skolem.go`). The core l2s flow (entry points, monitor construction, fair cycle, tableau, action instrumentation, named binder replacement, goal building) matches Python faithfully at ~95% fidelity.

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

## Implementation Order

1. **Bug 5** — one-line type narrowing, zero cascade
2. **Bug 1** — add one branch, zero cascade
3. **Bug 4** — new function + one caller switch
4. **Bug 3** — signature change, 3 callers
5. **Bug 2** — signature change, 4 callers with possible cascading

---

## Verification

1. `go build ./...` — ensure all changes compile
2. `go test ./proof/... ./check/... ./tactics/...` — run existing tests
3. Run any existing golden test vectors in `test_vectors/` that exercise proof/l2s paths
4. For Bug 1: verify that `GoalIsDefn(anUninterpretedSort)` now returns `true`
5. For Bug 2: verify that `GoalSubst` with clashing premise names returns an error
6. For Bug 3: verify that `Desugar` with `was(X)` (parameterized) returns an error
7. For Bug 4: verify that `GoalVocabBound` includes bound variables from conclusion
8. For Bug 5: verify that `GoalDefns` only collects `*lg.Const` premises, not other Expr types

---

## Files to Modify

| File | Bugs |
|------|------|
| `proof/goal.go` | 1, 2, 4, 5 |
| `check/l2s.go` | 3 |
| `check/ranking.go` | 3 (caller) |
| `proof/matching.go` | 2 (caller) |
| `proof/phase5_goals.go` | 2 (caller) |
| `tactics/ivy_tactics.go` | 4 (caller) |

---

## Items NOT bugs (confirmed matching)

These were initially suspected but confirmed correct after deeper analysis:

- **`extractNormalProgram(m)` vs `conc.model.clone([])`**: Both originate from `normal_program_from_module`. Python embeds it in the goal then clones; Go rebuilds. Equivalent since module state is unchanged between construction and l2s execution.
- **`FreeVariablesList` vs `variables_ast`**: Both compute FREE variables. Python's `variables_ast` excludes binder-bound variables (line 478-482 of `ivy_logic_utils.py`).
- **`GoalFree.recFmla`**: Uses same free-variable semantics as Python.
- **Desugar/applyWasRec core logic**: Matches Python's desugar/apply_was structurally.
- **Fair cycle construction**: Logic matches Python's inline list comprehensions.
- **Temporal premises collection**: Same filter and combination logic.
- **l2s entry points**: l2sTactic/L2STacticFull/L2STacticAuto match Python.

## Companion files NOT audited (future work)

- `check/l2s_auto.go` — vs Python ivy_l2s.py:222-704 (auto invariant generation, ~480 lines)
- `check/l2s_hooks.go` — vs Python ivy_l2s.py:1349-1537 (trace hooks, ~190 lines)
- `check/l2s_shared.go` — vs Python l2s_tactic_int inline steps
