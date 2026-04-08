# Plan: Comprehensive Unit Tests for z3bridge/z3_utils.go

**Created:** 2026-04-08

## Context

We just created `z3bridge/z3_utils.go` — a faithful port of Python `z3_utils.py`. It needs comprehensive unit tests in a new file `z3bridge/z3_utils_test.go`. The Python file has a `__main__` block (lines 174-197) with excellent test scenarios involving transitivity, antisymmetry, Iff, and Ite that we should port. We also need tests for every code path in the translator.

## File to Create

`z3bridge/z3_utils_test.go` (package `z3bridge`, same as existing `z3bridge_test.go`)

## Test Design

Reuse existing helper `mustFS(t, sorts...)` from `z3bridge_test.go`.

### Group 1: ToZ3 Sort Translation

- **TestToZ3BooleanSort** — `*logic.BooleanSort` → returns `Sort` (z3 BoolSort)
- **TestToZ3UninterpretedSort** — `*logic.UninterpretedSort{Name:"S"}` → returns `Sort`
- **TestToZ3UninterpretedSortCache** — same sort twice → same Z3 object (cache hit)
- **TestToZ3FunctionSortError** — `*logic.FunctionSort` → returns error

### Group 2: ToZ3 Term Translation

- **TestToZ3Variable** — `Variable{Name:"X", VSort:S}` → returns Expr, const named `"X:S"`
- **TestToZ3Const** — `Const{Name:"c", CSort:S}` → returns Expr, const named `"c:S"`
- **TestToZ3ConstBoolSort** — `Const{Name:"p", CSort:Boolean}` → returns Expr, named `"p:Boolean"`
- **TestToZ3ConstNullaryFunc** — `Const{Name:"c", CSort:FunctionSort(S)}` (arity 0) → returns Expr, named `"c:S"`
- **TestToZ3ConstHigherOrder** — `Const{Name:"f", CSort:FunctionSort(S,S,Boolean)}` → returns FuncDecl
- **TestToZ3VariableHigherOrderError** — `Variable` with FunctionSort arity>=1 → error

### Group 3: ToZ3 Apply Translation

- **TestToZ3ApplyNullary** — `Apply{Func:c, Terms:[]}` → delegates to ToZ3(c)
- **TestToZ3ApplyWithArgs** — `Apply{Func:f, Terms:[x,y]}` where f has FunctionSort → applies FuncDecl to args

### Group 4: ToZ3 Formula Translation

- **TestToZ3Eq** — `Eq{T1:x, T2:y}` → z3 Eq
- **TestToZ3Not** — `Not{Body:p}` → z3 Not
- **TestToZ3And** — `And{Terms:[p,q]}` → z3 And
- **TestToZ3Or** — `Or{Terms:[p,q]}` → z3 Or
- **TestToZ3AndEmpty** — `And{Terms:[]}` (logic.True) → z3 BoolVal(true)
- **TestToZ3OrEmpty** — `Or{Terms:[]}` (logic.False) → z3 BoolVal(false)
- **TestToZ3Implies** — `Implies{T1:p, T2:q}` → z3 Implies
- **TestToZ3Iff** — `Iff{T1:p, T2:q}` → z3 **Eq** (not Iff! matching Python)
- **TestToZ3Ite** — `Ite{Cond:b, Then:x, Else:y}` → z3 Ite

### Group 5: ToZ3 Quantifier Translation

- **TestToZ3ForAll** — `ForAll{Variables:[X,Y], Body:...}` → z3 ForAll, NO quant constraints
- **TestToZ3Exists** — `Exists{Variables:[X], Body:...}` → z3 Exists
- **TestToZ3ForAllEmpty** — `ForAll{Variables:[], Body:p}` → returns ToZ3Expr(p) directly
- **TestToZ3ExistsEmpty** — same for Exists

### Group 6: ToZ3 Caching

- **TestToZ3CacheHit** — translate same expr twice, verify cache is populated
- **TestToZ3ClearResetsCache** — translate, Clear(), translate again → cache miss (fresh result)

### Group 7: Z3Implies (port of Python z3_implies)

- **TestZ3UtilsImpliesValid** — `p ∧ q ⊨ p` → true
- **TestZ3UtilsImpliesInvalid** — `p ⊭ q` → false
- **TestZ3UtilsImpliesCache** — call twice, verify ImpliesCache populated
- **TestZ3UtilsImpliesTautology** — `true ⊨ p ∨ ¬p` → true
- **TestZ3UtilsImpliesTimeout** — same as valid but with timeout=true

### Group 8: Z3ImpliesBatch (port of Python z3_implies_batch)

- **TestZ3UtilsBatchValid** — `p ∧ q ⊨ {p, q}` → [true, true]
- **TestZ3UtilsBatchInvalid** — `p ⊨ {q}` → [false]
- **TestZ3UtilsBatchMixed** — `p ∧ q ⊨ {p, r, q}` → [true, false, true]
- **TestZ3UtilsBatchEmpty** — `p ⊨ {}` → []
- **TestZ3UtilsBatchCacheHit** — call twice, second uses cache

### Group 9: Python __main__ Block (lines 174-197)

This is the most important group — ports the exact test scenarios from Python z3_utils.py.

- **TestZ3UtilsTransitivityEquivalences** — Three equivalent formulations of transitivity:
  ```
  transitive1 = ForAll(X,Y,Z, Implies(And(leq(X,Y), leq(Y,Z)), leq(X,Z)))
  transitive2 = ForAll(X,Y,Z, Or(Not(leq(X,Y)), Not(leq(Y,Z)), leq(X,Z)))
  transitive3 = Not(Exists(X,Y,Z, And(leq(X,Y), leq(Y,Z), Not(leq(X,Z)))))
  ```
  Verify: t1⊨t2, t2⊨t3, t3⊨t1 (all true)

- **TestZ3UtilsTransitivityNotAntisymmetric** — transitivity does NOT imply antisymmetry:
  ```
  antisymmetric = ForAll(X,Y, Implies(And(leq(X,Y), leq(Y,X), true), Eq(Y,X)))
  ```
  Verify: t3⊭antisymmetric (false)

- **TestZ3UtilsIffEquivalence** — `true ⊨ Iff(transitive1, transitive2)` → true

- **TestZ3UtilsIteImplications** — Ite (if-then-else) properties:
  ```
  b ⊨ Eq(Ite(b,x,y), x)           → true
  ¬b ⊨ Eq(Ite(b,x,y), y)          → true
  ¬Eq(x,y) ⊨ Iff(Eq(Ite(b,x,y),x), b) → true
  ```

### Group 10: Free Variable Sharing

- **TestZ3UtilsFreeVarsShared** — Free variable X in premise and formula resolves to same Z3 constant:
  `r(X) ⊨ r(X)` → true (confirms no closing, shared constants)

## Critical Files

- `z3bridge/z3_utils.go` — code under test (just created)
- `z3bridge/z3bridge_test.go` — existing test file with `mustFS` helper to reuse
- `z3bridge/z3_utils_test.go` — **NEW** test file
- Python source: `~/ivy/pyivy/ivy/ivy/z3_utils.py` lines 174-197 (__main__ block)

## Verification

```
go test ./z3bridge/ -run "TestZ3Utils|TestToZ3" -v -count=1
```
