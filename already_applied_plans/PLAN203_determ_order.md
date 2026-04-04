# PLAN203: Fix Non-Deterministic Def Ordering in Clauses Operations

**Created:** 2026-04-05 ~04:30 UTC

## Context

The `TestOrdLive` golden test diverges at xtrace line 159811 in `ComposeUpdates` result. The Go and Python s-expression outputs for the `TR` Clauses differ — the `defs:[]` list has the same definitions but in **different order**. This causes the positional diff to show apparently different definitions at the same index, when in fact the content is equivalent but reordered.

**Root cause:** Several functions in `module/ops.go` and `actions/transrel.go` iterate over Go `map` types to build ordered output (definition slices) or to call `UniqueRenamer.Rename()` where call order determines generated names. Go map iteration is explicitly randomized, while Python 3.7+ dicts preserve insertion order.

**Evidence:** The diff at the divergence point:
- Go's 2nd def: `(Def lhs:(Symbol name:__ts0__ts0_a_a sort:(BooleanSort)) rhs:(Eq ...))`
- Python's 2nd def: `(Def lhs:(Symbol name:new_loc:op_ack sort:(EnumeratedSort ...)) rhs:(Ite ...))`

Both definitions likely exist in both outputs, but at different positions in the `defs:[]` array.

## Bug Sites (5 locations)

### Site A: `module/ops.go` — `iteClausesInt` (line 380-403) — PRIMARY

```go
defIdx := make(map[lg.NodeKey]*il.Definition)  // line 380
// ... populate ...
for _, d := range defIdx {   // line 399 — NON-DETERMINISTIC
    defs = append(defs, d)
}
```

**Impact:** Directly produces the `Defs` slice in the resulting `Clauses`. Called by `IteClauses` which is called by `iteUpdate` → `IteAction` → `IfAction.IntUpdate`. This is the exact code path producing the divergent `ComposeUpdates` result.

**Fix:** Replace `map[lg.NodeKey]*il.Definition` with `*iu.InsMap[lg.NodeKey, *il.Definition]`. Iterate with `.All()` to get insertion-order iteration.

### Site B: `module/ops.go` — `orClausesInt` (line 312-336) — SECONDARY

Same pattern as Site A but in the `Or` (disjunction) path:
```go
defIdx := make(map[lg.NodeKey]*il.Definition)  // line 312
// ... populate ...
for _, d := range defIdx {   // line 332 — NON-DETERMINISTIC
    defs = append(defs, d)
}
```

**Fix:** Same as Site A — use `InsMap`.

### Site C: `module/ops.go` — `elimDeadDefinitions` (line 772)

```go
defined := make(map[lg.NodeKey]*lg.Const)  // line 758
// ... populate ...
for key := range defined {   // line 772 — NON-DETERMINISTIC
    // builds captured slice, which feeds toRename, which feeds rn.Rename()
}
```

**Impact:** The `captured` slice order determines the order of `rn.Rename()` calls in the skolem renaming loop (line 804-806). Different call order → different generated fresh names. Called by both `iteClausesInt` and `orClausesInt`.

**Fix:** Use `InsMap` for `defined`, or sort `captured` keys before processing.

### Site D: `actions/transrel.go` — `RenameDistinctClauses` (line 477)

```go
for s := range used1 {   // line 477 — NON-DETERMINISTIC
    if IsSkolem(s) && !IsGlobalSkolem(s) {
        nameMap[s] = rn.Rename(s)
    }
}
```

**Impact:** When multiple skolems conflict with `used2`, the generated fresh names depend on call order. Affects `ComposeUpdates` step 1.

**Fix:** Sort the filtered skolem names before calling `rn.Rename()`.

### Site E: `actions/transrel.go` — `RenameDistinct` (line 454)

Same pattern as Site D but for the `lg.Expr` (non-Clauses) version:
```go
for s := range used1 {   // line 454 — NON-DETERMINISTIC
    if IsSkolem(s) && !IsGlobalSkolem(s) {
        nameMap[s] = rn.Rename(s)
    }
}
```

**Fix:** Sort the filtered skolem names before calling `rn.Rename()`.

## Non-Issues (verified OK)

- `collectUsedNames` (ops.go:742): Output feeds into `NewUniqueRenamer` which converts to a set — order irrelevant.
- `nameSetToSlice` (transrel.go:347): Same — feeds into `NewUniqueRenamer`.
- `ComposeUpdates` mid-variable loop: Iterates over `mid []*lg.Const` (a slice from `updated1`), deterministic.

## Fix Strategy

Use `InsMap` at all 5 sites. This matches Python's insertion-ordered dict semantics exactly. InsMap insertion is O(log n) — cheaper than collecting into a slice and sorting O(n log n), especially since `elimDeadDefinitions` (Site C) is called from every `iteClausesInt`/`orClausesInt` invocation, and `RenameDistinctClauses` (Site D) can have large symbol sets.

### Step 1: Fix `iteClausesInt` (Site A) — `module/ops.go`

Replace:
```go
defIdx := make(map[lg.NodeKey]*il.Definition)
```
With:
```go
defIdx := iu.NewInsMap[lg.NodeKey, *il.Definition]()
```

Update the populate loop (lines 381-396) to use `defIdx.Set(key, d)` and `defIdx.Get2(key)`.

Update the extract loop (lines 398-401):
```go
var defs []*il.Definition
for _, d := range defIdx.All() {
    defs = append(defs, d)
}
```

### Step 2: Fix `orClausesInt` (Site B) — `module/ops.go`

Same transformation as Step 1 for the `defIdx` at line 312-336.

### Step 3: Fix `elimDeadDefinitions` (Site C) — `module/ops.go`

Replace `defined := make(map[lg.NodeKey]*lg.Const)` (line 758) with `InsMap[lg.NodeKey, *lg.Const]`.
Update population (lines 760-768) to use `.Set(key, c)`.
Update iteration (line 772) to use `.All()` for deterministic `captured` ordering.

### Step 4: Fix `RenameDistinctClauses` (Site D) — `actions/transrel.go`

Change `usedSymbolNamesClauses` to return `*iu.InsMap[string, bool]` (or change just the call site). Replace:
```go
for s := range used1 {
    if IsSkolem(s) && !IsGlobalSkolem(s) {
        nameMap[s] = rn.Rename(s)
    }
}
```
With InsMap iteration via `.All()` so `rn.Rename()` is called in insertion order (matching Python's deterministic set traversal).

Alternatively, if changing `usedSymbolNamesClauses` return type is too invasive, collect the skolem names into a local `InsMap[string, bool]` during the filter loop, then iterate that.

### Step 5: Fix `RenameDistinct` (Site E) — `actions/transrel.go`

Same transformation as Step 4 but for `usedSymbolNames` (the `lg.Expr` version).

### Step 6: Verify

```bash
go build ./...
go test ./module/...
go test ./actions/...
go test ./lalr_full/ -run TestOrdLive -count=1
```

The `TestOrdLive` golden test should pass or advance past xtrace line 159811.

## Files Modified

- `module/ops.go` — Sites A, B, C: `iteClausesInt`, `orClausesInt`, `elimDeadDefinitions`
- `actions/transrel.go` — Sites D, E: `RenameDistinctClauses`, `RenameDistinct`
