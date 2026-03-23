# Solver Package Code Review: Bug Fixes & Python Conformance

## Context
Deep line-by-line comparison of Go solver (`~/goivy/solver/`) against Python source (`~/pyivy/ivy/ivy/ivy_solver.py`). Found 9 bugs/divergences, ranked by severity.

---

## Critical Bugs

### Bug 1: Range sort type constraints never generated
**File:** `solver/solver.go:321-397` (`typeConstraintsForSymbol`)
**Python:** `ivy_solver.py:550-576` (`type_constraints`)

The Go code at lines 344-347 does:
```go
interpStr, isStr := interp.(string)
if !isStr { return nil }   // <-- BLOCKS RangeSort!
```
Since `*lg.RangeSort` is not a string, the function returns `nil` before reaching the RangeSort check at line 379. **Range sort bounds constraints are completely broken.**

**Fix:** Remove the early `return nil` and restructure:
```go
if interpStr, ok := interp.(string); ok && interpStr == "nat" {
    // nat constraints (lines 368-376)
}
if rs, ok := interp.(*lg.RangeSort); ok {
    // range sort constraints (lines 379-394) — now reachable
}
```

---

### Bug 2: GetSmallModel finalCond — Pop before callbacks
**File:** `solver/model.go:162-185` (`GetSmallModelWithCond`)
**Python:** `ivy_solver.py:1240-1260`

Go does `Push, Assert, Check, Pop, then Sat()/Unsat()`. Python does `push, add, check, Sat()/Unsat(), then pop`. Callbacks may inspect the solver/model state, which is destroyed after Pop.

**Fix:** Move `z3solver.Pop()` to AFTER `Sat()/Unsat()` calls:
```go
z3solver.Push()
z3solver.Assert(zCond)
res := z3solver.Check()
if res != z3bridge.Unsat {
    overallResult = res
    if fc.Sat() {
        overallResult = z3bridge.Unsat
        z3solver.Pop()
        continue
    }
    z3solver.Pop()
    break
} else {
    fc.Unsat()
}
z3solver.Pop()
```

---

### Bug 3: SortedSortUniverse ignores user-defined orderings
**File:** `solver/herbrand.go:90-162` (`SortedSortUniverse`, `evalLt`)
**Python:** `ivy_solver.py:776-786, 839-855`

Go uses `ctx.Lt(a, b)` — Z3 built-in `<` that only works on Int/Real/BV. Python translates the Ivy `<` symbol via `atom_to_z3(order(*vs))` and uses `substitute + model.eval`, which works for **user-defined** `<` on uninterpreted sorts.

**Fix:** Rewrite to match Python:
1. Create Ivy `<` symbol and application over variables X, Y
2. Translate to Z3 via `h.tr.Translate(orderApp)`
3. Translate X, Y to Z3: `z3X, z3Y`
4. Sort using comparator: `SubstituteZ3(ctx, orderZ3, {{z3X, a}, {z3Y, b}})` then `model.Eval`
5. Wrap in `defer recover()` matching Python's `except IndexError: pass`

---

### Bug 4: mineInterpretedConstants — inverted + incomplete
**File:** `solver/herbrand.go:356-376`
**Python:** `ivy_solver.py:803-818`

Two problems:
- **Inverted filter:** Go skips interpreted sorts (`continue`), Python only processes them
- **Missing collect_model_values:** Python creates `sym(V0,V1,...)`, evaluates in model, recursively collects numerals from ITE branches. Go just evaluates the symbol directly.

**Fix:**
1. Invert: process symbols whose range sort IS interpreted
2. Implement full `collectModelValues`: build placeholder term, translate, eval in model, collect numerals from ITE tree
3. Store results in `h.constants[sortName]` as Z3 expressions

---

### Bug 5: constantFromZ3 returns wrong types for True/False
**File:** `solver/herbrand.go:380-388`
**Python:** `ivy_solver.py:908-913`

Python returns `ivy_logic.And()` for true (empty conjunction) and `ivy_logic.Or()` for false (empty disjunction). Go returns `lg.NewSymbol("true/false", lg.Boolean)`. Callers may type-switch on And vs Or.

**Fix:** Return `lg.True` / `lg.False` (or `&lg.And{}` / `&lg.Or{}` if that's what the Go logic package uses). May need to change return type from `*lg.Symbol` to `lg.Expr` and update callers.

---

### Bug 6: TermsMatch — no variable substitution matching
**File:** `solver/compat.go:167-177`
**Python:** `ivy_solver.py:753-766`

Go does structural equality. Python does unification-style pattern matching: variables in `tl1` are bound to corresponding terms in `tl2`, and consistency is checked.

**Fix:** Implement with substitution environment:
```go
env := map[string]string{}
for i := range tl1 {
    if v, ok := tl1[i].(*lg.Variable); ok {
        if prev, exists := env[v.Name]; exists {
            if exprName(tl2[i]) != prev { return false }
        } else {
            env[v.Name] = exprName(tl2[i])
        }
    } else {
        if exprName(tl1[i]) != exprName(tl2[i]) { return false }
    }
}
```

---

## Medium-Priority Divergences

### Bug 7: Non-incremental mode missing in GetSmallModel
**File:** `solver/model.go`
**Python:** `ivy_solver.py:1226-1230`

Python supports `opt_incremental=False` which creates a fresh solver per checker, re-adding accumulated assumes. Go only has incremental mode.

**Fix:** Add `Incremental` to Options. When false, create fresh solver + re-add base clauses + accumulated assumes for each check.

---

### Bug 8: ClausesCase returns nil on UNSAT (should return FalseClauses)
**File:** `solver/herbrand.go:737`
**Python:** `ivy_solver.py:1037` returns `[[]]` (false clauses)

**Fix:** Change `return nil, nil` to `return FalseClauses(), nil` (one line).

---

### Bug 9: CheckCube missing memo/caching
**File:** `solver/model.go:303-314`
**Python:** `ivy_solver.py:714-733`

Python caches by Z3 AST ID with `memo` and `memo_unsat_only` parameters.

**Fix:** Add optional `memo map[uint32]*memoEntry` parameter. Requires `z3bridge.GetId(e) uint32`.

---

## Implementation Order (by dependency and impact)

| Priority | Bug | File | Effort |
|----------|-----|------|--------|
| 1 | Bug #1 — RangeSort constraints | solver.go | Small (restructure if-chain) |
| 2 | Bug #2 — Pop timing | model.go | Small (reorder 3 lines) |
| 3 | Bug #8 — ClausesCase nil | herbrand.go | Trivial (1 line) |
| 4 | Bug #6 — TermsMatch | compat.go | Small (rewrite function) |
| 5 | Bug #5 — constantFromZ3 types | herbrand.go | Medium (return type + caller updates) |
| 6 | Bug #3 — SortedSortUniverse | herbrand.go | Medium (new sort comparator) |
| 7 | Bug #4 — mineInterpreted | herbrand.go | Medium (new collect logic) |
| 8 | Bug #9 — CheckCube memo | model.go | Medium (new z3bridge API) |
| 9 | Bug #7 — Non-incremental mode | model.go | Large (fresh solver path) |

## Verification

After each fix:
1. `make test` from goivy root (uses DYLD_LIBRARY_PATH for Z3)
2. Run any existing solver tests
3. For Bug #1: verify with a test that creates a RangeSort interp, calls ClausesToZ3, and checks that the Z3 output contains bound constraints
4. For Bug #2: verify FinalCond callbacks can access the model
5. For Bug #3: verify uninterpreted sort elements are sorted by user-defined `<`
