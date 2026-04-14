# Plan: Convert `Sig.Sorts` from `map` to `InsMap` for insertion-order iteration

Created: 2026-04-15 ~00:10 UTC

## Context

**Divergence**: `make golden` now diverges at xtrace line 844423 (previously 736175 — good progress). The divergence is in `TestOrdLive`, binding[24] `idle`. The `l2s_a`/`l2s_d` assignment actions are generated in a different order:

- **Go**: `addr_type`, `index`, `lclock`, ... (alphabetical)
- **Python**: `index`, `lclock`, `proc`, ... (dict insertion order)

**Root cause**: In `check/l2s.go:467`, Go iterates `m.Sig.Sorts` (a `map[string]lg.Sort` — random order) then sorts alphabetically at line 475. Python iterates `ilg.sig.sorts.values()` (`ivy_l2s.py:224`) which preserves dict insertion order (Python 3.7+). The alphabetical sort in Go doesn't match Python's insertion-order semantics.

**Fix approach**: Convert `Sig.Sorts` from `map[string]lg.Sort` to `*iu.InsMap[string, lg.Sort]`, matching the established pattern used by `Sig.Symbols`. Then remove the alphabetical `sort.Slice` in l2s.go so iteration naturally follows insertion order.

## Changes

### 1. Core: `ivylogic/sig.go` — change field type and update methods

**Field definition** (line 36):
```go
// Before:
Sorts map[string]lg.Sort
// After:
Sorts *iu.InsMap[string, lg.Sort]
```

**NewSigOn()** (line 58):
```go
Sorts: iu.NewInsMap[string, lg.Sort](),
```
Line 65: `s.Sorts["bool"] = lg.Boolean` → `s.Sorts.Set("bool", lg.Boolean)`

**Copy()** (lines 78, 86-88):
```go
// remove Sorts from struct literal (no pre-sized make)
// add after struct literal:
res.Sorts = iu.NewInsMap[string, lg.Sort]()
for k, v := range s.Sorts.All() {
    res.Sorts.Set(k, v)
}
```

**String()** (line 234): `range s.Sorts` → `range s.Sorts.All()`

**SortNames()** (lines 322-326): `len(s.Sorts)` → `s.Sorts.Len()`, `range s.Sorts` → `range s.Sorts.All()` (key-only iteration: `for n := range s.Sorts.All()`)

Note: Go range over `iter.Seq2` with a single variable captures the key only, but we need to verify this compiles. If not, use `for n, _ := range s.Sorts.All()`.

**Canon()** — already calls `SortNames()` then sorts alphabetically. No change needed.

**AddSort()** (line 350): `s.Sorts[name] = sort` → `s.Sorts.Set(name, sort)`

**WithSorts.Enter()** (lines 434-437):
- `ws.sig.Sorts[name]` → `ws.sig.Sorts.Get2(name)`
- `ws.sig.Sorts[name] = s` → `ws.sig.Sorts.Set(name, s)`

**WithSorts.Exit()** (lines 445-448):
- `delete(ws.sig.Sorts, name)` → `ws.sig.Sorts.Delkey(name)`
- `ws.sig.Sorts[s.name] = s.sort` → `ws.sig.Sorts.Set(s.name, s.sort)`

### 2. `ivylogic/globals.go` — 3 lines
- Line 217: `sd.sig.Sorts["S"] = sd.sort` → `.Set("S", sd.sort)`
- Line 224: `sd.sig.Sorts["S"] = sd.oldSort` → `.Set("S", sd.oldSort)`
- Line 227: `delete(sd.sig.Sorts, "S")` → `sd.sig.Sorts.Delkey("S")`

### 3. `ivylogic/ivylogic.go` — 2 lines
- Line 418: `sig.Sorts["S"] = ds` → `sig.Sorts.Set("S", ds)`
- Line 428: `for _, s := range sig.Sorts {` → `for _, s := range sig.Sorts.All() {`

### 4. `check/l2s.go` — the triggering fix
- Line 467: `for name, s := range m.Sig.Sorts {` → `for name, s := range m.Sig.Sorts.All() {`
- Lines 475-477: **Remove** the `sort.Slice(uninterpretedSorts, ...)` — insertion order from InsMap now provides the correct Python-matching order
- Remove `"sort"` import if no longer used (check other usages first)

### 5. `check/ranking.go`
- Line 246: `for name, s := range m.Sig.Sorts {` → `for name, s := range m.Sig.Sorts.All() {`

### 6. `module/resort.go` — full map replacement
Lines 76-82: Replace `newSorts` map pattern with InsMap rebuild:
```go
newSorts := iu.NewInsMap[string, lg.Sort]()
for _, sort := range sig.Sorts.All() {
    newSort := lu.ResortSort(sort, ss)
    newName := il.SortName(newSort)
    newSorts.Set(newName, newSort)
}
sig.Sorts = newSorts
```

### 7. `module/module.go`
- Line 665: `if _, inSig := m.Sig.Sorts[rep]; inSig {` → use `.Get2(rep)`

### 8. `module/theory.go`
- Line 418: `if _, ok := m.Sig.Sorts[vname]; ok {` → `.Get2(vname)`
- Line 426: `m.Sig.Sorts[sname]` → `.Get2(sname)`

### 9. `compiler/phase6.go`
- Line 38: `for _, s := range sig.Sorts {` → `for _, s := range sig.Sorts.All() {`
- Line 943: `c.Sig.Sorts[resolved]` → `c.Sig.Sorts.Get2(resolved)`
- Line 1224: `schemaSig.Sorts[name] = sort` → `schemaSig.Sorts.Set(name, sort)`
- Line 2539: `mod.Sig.Sorts[sortname]` → `mod.Sig.Sorts.Get2(sortname)`
- Line 2569: `mod.Sig.Sorts[name]` → `mod.Sig.Sorts.Get2(name)`
- Line 2582: `mod.Sig.Sorts[name]` → `mod.Sig.Sorts.Get(name)`

### 10. `compiler/helpers.go`
- Line 101: `c.Sig.Sorts[symbolName]` → `c.Sig.Sorts.Get2(symbolName)`

### 11. `compiler/compiler.go`
- Line 400: `c.Sig.Sorts[sortName]` → `c.Sig.Sorts.Get2(sortName)`

### 12. `compiler/ivy_compile.go`
- Line 1838: `len(sig.Sorts)` → `sig.Sorts.Len()`

### 13. `isolate/strip.go`
- Line 40: `mod.Sig.Sorts[name]` → `mod.Sig.Sorts.Get2(name)`
- Line 566: `mod.Sig.Sorts[v.VSort]` → `mod.Sig.Sorts.Get2(v.VSort)`
- Line 689: `mod.Sig.Sorts[sortAtom.Rep]` → `mod.Sig.Sorts.Get2(sortAtom.Rep)` or `.Get(sortAtom.Rep)`
- Line 717: `mod.Sig.Sorts[paramName]` → `mod.Sig.Sorts.Get2(paramName)`
- Line 736: `mod.Sig.Sorts[paramName]` → `mod.Sig.Sorts.Get2(paramName)`
- Line 932: `delete(mod.Sig.Sorts, sortName)` → `mod.Sig.Sorts.Delkey(sortName)`
- Line 944-945: `range mod.Sig.Sorts` → check context, probably needs to filter by SortOrder

### 14. `isolate/create.go`
- Line 517: `mod.Sig.Sorts[name]` → `mod.Sig.Sorts.Get2(name)`

### 15. `isolate/isolate.go`
- Line 1630: `for name := range mod.Sig.Sorts {` → `for name := range mod.Sig.Sorts.All() {`
- Line 1632: `delete(mod.Sig.Sorts, name)` → `mod.Sig.Sorts.Delkey(name)`
- Line 1637: `mod.Sig.Sorts[s]` → `mod.Sig.Sorts.Get2(s)`

### 16. `isolate/deps.go`
- Line 691: `delete(mod.Sig.Sorts, name)` → `mod.Sig.Sorts.Delkey(name)`
- Lines 696-697: `range mod.Sig.Sorts` → `range mod.Sig.Sorts.All()`

### 17. `z3bridge/solver_compat.go`
- Line 297: `for name, sort := range s.sig.Sorts {` → `for name, sort := range s.sig.Sorts.All() {`

### 18. `autoinst/schemata.go`
- Line 27: `for sortName := range m.Sig.Sorts {` → `for sortName := range m.Sig.Sorts.All() {`

### 19. `gogen/generator.go`
- Line 130: `g.Module.Sig.Sorts[sortName]` → `g.Module.Sig.Sorts.Get2(sortName)`

### 20. `gogen/types.go`
- Line 160: `mod.Sig.Sorts[sortName]` → `mod.Sig.Sorts.Get2(sortName)`

### 21. `webui/backend_go.go`
- Line 390: `for name, sort := range sess.CompiledSig.Sorts {` → `for name, sort := range sess.CompiledSig.Sorts.All() {`

### 22. Test files — mechanical `.Sorts[name] = val` → `.Sorts.Set(name, val)` and lookup → `.Get2()`
- `compiler/assert_like_test.go` (~2 writes)
- `compiler/batch_e_test.go` (~25 writes)
- `compiler/compiler_test.go` (~11 writes, ~5 reads)
- `compiler/decl_missing_test.go` (~11 writes)
- `compiler/decl_missing2_test.go` (~10 writes)
- `compiler/expr_test.go` (~12 writes)
- `compiler/theorem_to_property_test.go` (~1 read, 1 len, 1 range)
- `isolate/isolate_test.go` (~2 writes, 1 range)
- `gogen/generator_test.go` (~1 write)
- `gogen/types_test.go` (~1 write)
- `module/module_test.go` (~3 reads)
- `end2end/pipeline_test.go` (~4 reads)

## Critical files to modify

1. `ivylogic/sig.go` — core type change + method updates
2. `check/l2s.go` — remove alphabetical sort, use `.All()` iteration
3. `module/resort.go` — InsMap rebuild instead of map replacement
4. All files listed in sections 2-22 above

## Verification

1. `go build ./...` compiles clean
2. `go test ./...` passes (especially `ivylogic/`, `compiler/`, `isolate/` tests)
3. `make golden` — divergence advances past line 844423
