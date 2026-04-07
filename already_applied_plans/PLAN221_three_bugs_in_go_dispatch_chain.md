# PLAN220: Fix Z3 Sort Mismatch for Comparison Operators on Uninterpreted Sorts

**Created:** 2026-04-07 ~24:00 UTC

## Context

At golden test line 236047 (TestOrdLive), Go produces a Z3 Sort mismatch panic while Python succeeds:
```
Sort mismatch at argument #1 for function (declare-fun < (Int Int) Bool) supplied sort is lclock
```
This occurs when translating `ref.prevents`, which uses `<` on `lclock` arguments. `lclock` is an uninterpreted sort (no entry in `sig.Interp`). Python creates a sort-qualified uninterpreted Z3 function `<:lclock:lclock` with signature `(lclock_z3, lclock_z3) -> Bool`. Go incorrectly uses Z3's built-in integer `Lt`.

## Root Cause — Three bugs in dispatch chain

### Bug 1: `translateBuiltinOp` hardcodes integer comparisons (PRIMARY)
**File:** `z3bridge/translate.go:696-728`

`translateBuiltinOp` handles `<`, `<=`, `>`, `>=` by calling `ctx.Lt()`, `ctx.Le()`, etc. — Z3 arithmetic-only functions. Returns `handled=true` for ALL sorts, preventing the NativeLookup callback (line 233) from being reached.

### Bug 2: `isPolymorphicOp` excludes comparison operators
**File:** `solver/z3convert.go:1035-1041`

`isPolymorphicOp` only includes `+`, `-`, `*`, `/`. Excludes `<`, `<=`, `>`, `>=`. So even if NativeLookup were reached, `LookupNative` (line 576) wouldn't recognize comparisons as polymorphic and would return nil.

### Bug 3: `lookupPolymorphicNative` fallback ignores relations
**File:** `solver/z3convert.go:672`

The fallback calls `lookupBuiltinFunc(name, isRelation)` which has `+`, `-`, `*`, `/` but NOT `<`, `<=`, `>`, `>=`. For interpreted sorts, it should call `lookupBuiltinRelation(name)` which has them with BV-awareness.

## How Python Works (source of truth)

Python `ivy_solver.py:solver_name()` (line 61): For polymorphic symbols on uninterpreted sorts, mangles the name: `<` on `(lclock, lclock)` becomes `<:lclock:lclock`.

Python `ivy_solver.py:atom_to_z3()` (line 484):
1. `lookup_native("<", relations, "relation")` → returns None (lclock is uninterpreted)
2. `<` is NOT in `polymacs` (only `<=`, `>`, `>=` are)
3. Creates `z3.Function(solver_name(sym), *sig)` → uninterpreted function `<:lclock:lclock`

For interpreted sorts: `lookup_native` returns `relations_dict["<"]` → BV-aware lambda.

Go already has the correct infrastructure: `SolverName` (z3convert.go:974) correctly produces `<:lclock:lclock`, `lookupBuiltinRelation` (z3convert.go:768) has BV-aware comparisons. Only the dispatch routing is wrong.

## Detailed Changes

### A. `z3bridge/translate.go` — Remove comparison operators from `translateBuiltinOp`

Remove the four comparison cases (lines 696-728):
```go
// DELETE these cases:
case "<":
    args, err := translateArgs()
    ...
    return t.Ctx.Lt(args[0], args[1]), true, nil
case "<=":
    ...
case ">":
    ...
case ">=":
    ...
```

After removal, comparison operators return `handled=false` from `translateBuiltinOp`, falling through to:
- NativeLookup (line 233) — handles interpreted sorts via `lookupBuiltinRelation`
- `getFuncDecl` (line 260) — creates sort-qualified uninterpreted function for uninterpreted sorts

Arithmetic operators (`+`, `-`, `*`, `/`) and BV operators remain unchanged.

### B. `solver/z3convert.go` — Add comparison operators to `isPolymorphicOp`

**Line 1035-1041.** Change:
```go
func isPolymorphicOp(name string) bool {
    switch name {
    case "+", "-", "*", "/", "<", "<=", ">", ">=":
        return true
    }
    return false
}
```

Matches Python's `iu.polymorphic_symbols` which includes all eight operators. Called only at z3convert.go:576.

### C. `solver/z3convert.go` — Fix `lookupPolymorphicNative` fallback for relations

**Line 672.** Change:
```go
// OLD:
return s.lookupBuiltinFunc(name, isRelation)

// NEW:
if isRelation {
    return s.lookupBuiltinRelation(name)
}
return s.lookupBuiltinFunc(name, isRelation)
```

This ensures `<` on interpreted sorts dispatches to `lookupBuiltinRelation` (has BV-aware comparisons), matching the sibling function `lookupNamedNative` (line 676-681) which already does this.

### D. `solver/solver_test.go` — Add tests

**D1. TestTranslateComparisonUninterpretedSort** — Create `<` relation on `lclock` (no sig.Interp entry). Translate via solver. Verify no panic/error. Verify Z3 output contains `<:lclock:lclock` (uninterpreted function).

**D2. TestTranslateComparisonInterpretedSort** — Create `<` relation on sort with `sig.Interp["mysort"] = "int"`. Verify Z3 uses built-in Lt (not uninterpreted function).

**D3. TestTranslateComparisonBVSort** — Create `<` relation on sort with `sig.Interp["mysort"] = "bv[32]"`. Verify Z3 uses BvUlt.

## Dispatch Flow After Fix

### `<` on uninterpreted sort (e.g., lclock):
1. `translateBuiltinOp` → no match → `handled=false`
2. NativeLookup → `LookupNative` → `isPolymorphicOp("<")` → true
3. `lookupPolymorphicNative`:  `sig.Interp["lclock"]` → not found → returns nil
4. NativeLookup returns nil
5. `getFuncDecl` → `makeFuncDecl` → `z3Name` → `SolverName` → `"<:lclock:lclock"`
6. Creates Z3 uninterpreted function `<:lclock:lclock(lclock_z3, lclock_z3) -> Bool`

### `<` on interpreted sort (e.g., int/nat):
1. `translateBuiltinOp` → no match → `handled=false`
2. NativeLookup → `LookupNative` → `isPolymorphicOp("<")` → true
3. `lookupPolymorphicNative`: `sig.Interp["int"]` → `"int"` → not EnumeratedSort
4. Fallback: `lookupBuiltinRelation("<")` → returns BV-aware lambda
5. Lambda: `ctx.IsBvExpr(args[0])` → false → `ctx.Lt(args[0], args[1])`

### `<` on BV sort:
Same as interpreted, but at step 5: `ctx.IsBvExpr(args[0])` → true → `ctx.BvUlt(args[0], args[1])`

## Execution Order

All three fixes (A, B, C) must be applied together. Individually:
- A alone: breaks interpreted sorts (NativeLookup returns nil since isPolymorphicOp excludes `<`)
- B alone: no effect (translateBuiltinOp still intercepts)
- C alone: no effect (fallback unreachable for comparisons)

## Files Modified

- `z3bridge/translate.go` — remove comparison cases from `translateBuiltinOp`
- `solver/z3convert.go` — expand `isPolymorphicOp`, fix `lookupPolymorphicNative` fallback
- `solver/solver_test.go` — add 3 new tests

## Verification

```bash
cd ~/go/src/github.com/glycerine/ivy/goivy && go build ./... && go test ./solver/... && go test ./z3bridge/... && go test ./tactics/...
```

Then golden test:
```bash
cd ~/ivy/goivy && make golden
```

Expected: Sort mismatch at line 236047 is resolved. Golden test progresses further.
