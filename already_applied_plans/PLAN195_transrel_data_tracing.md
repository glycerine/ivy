# Fix incomplete `simpIte` in clauseops causing Ite-vs-Or divergence

**Created:** 2026-04-04 ~10:00 UTC

## Context

The TestOrdLive golden test diverges at XTRACE line 167247. Go's `CloseEpr` receives `Ite(cond, And([]), x)` while Python receives `Or(cond, x)`. These are semantically equivalent (`And()` = true, so `Ite(cond, true, x)` = `Or(cond, x)`), but structurally different.

The root cause: `clauseops/ops.go` has its own **incomplete** `simpIte` function (line 687) that only handles the `then == else` case. Meanwhile, the **complete** `SimpIte` in `ivylogic/classify.go` (line 186) matches Python's `simp_ite` exactly — including constant-folding for true/false branches. But `clauseops/ops.go` uses its own local `simpIte` instead of `il.SimpIte`.

The `simpIte` is called at two sites in `clauseops/ops.go`:
- Line 255: in `orClausesInt` (definition merging for disjunction)
- Line 323: in `iteClausesInt` (definition merging for if-then-else)

Both need the full simplification to match Python.

The CanonSnapshot at `fragment/fragment.go:1001` matches because it captures **module state** (axioms, props, actions as stored). The divergence happens later when `GetAssumesAndAsserts` calls `GetUpdate` → action `int_update` → the transrel pipeline builds clauses using `ite_clauses`/`or_clauses` which call `simpIte`. The unsimplified Ite survives into `CloseEpr`'s input. So the CanonSnapshot is correct — it just doesn't cover the **computed** transition-relation data. Adding deeper data traces in transrel will help catch such divergences earlier in the future.

## Plan

### Step 1: Fix `simpIte` in `clauseops/ops.go`

Replace the incomplete local `simpIte` with a call to the complete `il.SimpIte`:

**File:** `clauseops/ops.go`

Delete lines 685-697 (the `simpIte` function definition).

Replace the two call sites:
- Line 255: `simpIte(vs[i], d.Rhs, existing.Rhs)` → `il.SimpIte(vs[i], d.Rhs, existing.Rhs)`
- Line 323: `simpIte(v, existing.Rhs, d.Rhs)` → `il.SimpIte(v, existing.Rhs, d.Rhs)`

`clauseops` already imports `il "github.com/glycerine/ivy/goivy/ivylogic"`, so no import changes needed.

### Step 2: Add data-level xtracer traces in transrel

Currently transrel traces only execution flow (function enter/exit, list of modified symbols) but not the actual formula data flowing through. Add HASH canon traces for:

**File:** `transrel/transrel.go`

1. **`Hide` function** (~line 925-926): After `ExistQuantClauses` produces `newTR` and `newPre`, trace their canonical forms:
   ```go
   xtracer.Trace("transrel.Hide result HASH canon= newTR=%s", newTR.Canon())
   xtracer.Trace("transrel.Hide result HASH canon= newPre=%s", newPre.Canon())
   ```

2. **`ComposeUpdates`** (~line 689): After computing the composed result, trace the TR/Pre canon:
   ```go
   xtracer.Trace("transrel.ComposeUpdates result HASH canon= TR=%s Pre=%s", result.TR.Canon(), result.Pre.Canon())
   ```

**File:** `ivy_transrel.py` (Python side) — add matching traces at corresponding points.

### Step 3: Add corresponding Python xtracer traces in transrel

**File:** `~/ivy/pyivy/ivy/ivy/ivy_transrel.py`

Add matching `xtracer.trace()` calls at the same points as Go:

1. After `hide` produces `new_tr` and `new_pre`.
2. After `compose_updates` produces the result.

### Step 4: Add unit test for `simpIte` / `SimpIte` constant folding

**File:** `clauseops/ops_test.go` (or `ivylogic/classify_test.go` if more appropriate)

Test that `il.SimpIte` correctly simplifies:
- `SimpIte(cond, And{}, x)` → `Or{cond, x}` (true then-branch → Or)
- `SimpIte(cond, Or{}, x)` → `And{Not(cond), x}` (false then-branch → And)
- `SimpIte(cond, x, And{})` → `Or{Not(cond), x}` (true else-branch → Or)
- `SimpIte(cond, x, Or{})` → `And{cond, x}` (false else-branch → And)
- `SimpIte(And{}, x, y)` → `x` (true condition)
- `SimpIte(Or{}, x, y)` → `y` (false condition)
- `SimpIte(cond, x, x)` → `x` (equal branches)
- `SimpIte(cond, x, y)` → `Ite{cond, x, y}` (no simplification)

Also test the helper functions `SimpAnd`, `SimpOr`, `SimpNot` for completeness.

### Step 5: Add CanonSnapshot coverage for computed data

The user correctly notes that the CanonSnapshot at `fragment/fragment.go:1001` didn't catch this. The snapshot covers **stored** module state but not the **computed** data (action updates, transition relations). We will add:

**File:** `fragment/fragment.go`

After `GetAssumesAndAsserts` returns (around line 1010), trace the assumes/asserts data:
```go
xtracer.Trace("fragment.CheckFragment HASH canon= assumes count=%d", len(assumes))
for i, a := range assumes {
    xtracer.Trace("fragment.CheckFragment HASH canon= assume[%d]=%s", i, a.fmla.Canon())
}
```

This matches what Python already does at `ivy_fragment.py:472-477`.

## Critical files to modify

1. `clauseops/ops.go` — delete `simpIte`, replace calls with `il.SimpIte`
2. `transrel/transrel.go` — add data-level HASH canon traces
3. `~/ivy/pyivy/ivy/ivy/ivy_transrel.py` — add matching Python traces
4. `ivylogic/classify_test.go` (new) — unit tests for SimpIte/SimpAnd/SimpOr/SimpNot
5. `fragment/fragment.go` — add assumes/asserts data traces after GetAssumesAndAsserts

## Existing code to reuse

- `il.SimpIte` at `ivylogic/classify.go:186` — already complete, matches Python exactly
- `il.SimpAnd`, `il.SimpOr`, `il.SimpNot` at `ivylogic/classify.go:138-184`
- `lg.IsTrue`, `lg.IsFalse` at `logic/formula.go:17-28`
- `clauseops.Clauses.Canon()` — if it exists, for tracing
- `xtracer.Trace` / `xtracer.trace` — existing tracing infrastructure

## Verification

1. `go build ./...` — ensure compilation
2. `go test ./ivylogic/...` — run the new SimpIte unit tests
3. `go test ./clauseops/...` — ensure clauseops still works
4. Re-run `TestOrdLive` in `lalr_full/` — the divergence at line 167247 should disappear (or move to a later, different divergence point)
5. Check that the new transrel data traces appear in both Go and Python logs and match
