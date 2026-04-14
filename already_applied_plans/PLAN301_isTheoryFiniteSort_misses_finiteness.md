# Plan: Add missing `IsFinite()` methods on sort types and fix `isTheoryFiniteSort`

Created: 2026-04-14 ~22:45 UTC

## Context

**Divergence**: `make golden` still diverges at xtrace line 736175 (in `TestOrdLive`, same line as before the `SortName` fix). The `SortName` fix (applied in prior session) was necessary but insufficient — the root cause is deeper.

Both Go and Python enter `replaceNamedBindersAst` with an ActionTerm for binding `ext:cfabric.step` (binding[7]). The ActionTerm's `stmt` field differs:

- **Go**: First child is `assignAction(l2s_d(ph_type -> bool)(fml:ph), true)` — an l2s "done" variable assignment
- **Python**: First child is `assumeAction(Implies(NamedBinder(l2s_g, ...)))` — an l2s "guarantee" assume

This means Go's `addParamsToD` list in Step 8 (PatchExports) is non-empty for `fml:ph` (EnumeratedSort ph_type), while Python's `add_params_to_d` is empty.

**Root cause**: `isTheoryFiniteSort` in `l2s.go:259-272` calls `theory.GetSortTheory(s, interp)` and then ONLY checks if the result is `*theory.Theory` with `t.Finite`. For EnumeratedSort, `GetSortTheory` returns the sort itself (not a Theory), so `isTheoryFiniteSort` returns `false`.

In Python (`ivy_logic.py:797`), `EnumeratedSort.is_finite = lambda self: True`. When `get_sort_theory(sort)` returns the sort itself, calling `.is_finite()` returns True. Go's sort types are missing `IsFinite()` methods for `EnumeratedSort`, `UninterpretedSort`, `RangeSort`, and `TopSort`.

**Secondary bug**: `GetSortTheory` (`theory/theory.go:218`) uses `sort.String()` for the interp lookup, but Python's `get_sort_theory` uses `sort.name`. For EnumeratedSort, `.String()` returns `"{nop_ph,wr_ph,rd_ph,cpl_ph}"` while `.name` is `"ph_type"`. Same issue in `HasIntegerInterp` at line 239.

## Fix

### 1. Add `IsFinite()` methods to sort types in `logic/sort.go`

Match Python `ivy_logic.py` monkey-patched lambdas:

| Sort type | Python `is_finite` | Go `IsFinite()` | Status |
|-----------|-------------------|------------------|--------|
| BooleanSort | True (line 168) | true | Already exists (line 44) |
| FunctionSort | True (line 838) | true | Already exists (line 83) |
| UninterpretedSort | False (line 790) | false | **ADD** after line 35 |
| EnumeratedSort | True (line 797) | true | **ADD** after line 110 |
| RangeSort | True (line 801) | true | **ADD** after line 155 |
| TopSort | (not in Python) | false | **ADD** after line 172 |

### 2. Fix `isTheoryFiniteSort` in `check/l2s.go:259-272`

When `GetSortTheory` returns a sort (not a `*theory.Theory`), check its `IsFinite()` method:

```go
// Python: sort.is_finite() — each sort type has an is_finite method.
if ifc, ok := th.(interface{ IsFinite() bool }); ok {
    return ifc.IsFinite()
}
```

### 3. Fix `GetSortTheory` interp lookup in `theory/theory.go:218`

Change `name := sort.String()` to `name := lg.SortName(sort)` to match Python's `name = sort.name`.

### 4. Fix `HasIntegerInterp` interp lookup in `theory/theory.go:239`

Same: change `name := sort.String()` to `name := lg.SortName(sort)`.

## Critical files to modify

1. `logic/sort.go` — add `IsFinite()` to 4 sort types
2. `check/l2s.go` — fix `isTheoryFiniteSort` to check sort-level finiteness
3. `theory/theory.go` — fix interp lookups in `GetSortTheory` and `HasIntegerInterp`

## Verification

1. `go build ./...` compiles clean
2. `make golden` — divergence advances past line 736175
