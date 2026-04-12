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

## Companion files NOT audited (future work)

- `check/l2s_auto.go` — vs Python ivy_l2s.py:222-704 (auto invariant generation, ~480 lines). Known to use `ilu.variables_ast` at lines 237, 586, 595, 672 — likely has Bug 7 equivalents.
- `check/l2s_hooks.go` — vs Python ivy_l2s.py:1349-1537 (trace hooks, ~190 lines)
- `check/l2s_shared.go` — vs Python l2s_tactic_int inline steps
