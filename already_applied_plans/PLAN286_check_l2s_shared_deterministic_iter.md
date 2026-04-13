# Fix l2s_consts_d non-deterministic symbol ordering

Created: 2026-04-13 ~13:00 UTC

## Context

TestOrdLive diverges at xtrace line 321375. The `l2s_consts_d` invariant's `And` terms appear in different order between Go and Python because Go iterates `map[string]*SymbolEntry` (random order) while Python iterates `dict.values()` (insertion order).

Go produces: `l2s_d(_T) & l2s_d(cfabric.t_rd_min) & l2s_d(ref.lt) & ...`
Python produces: `l2s_d(ref.lt) & l2s_d(cfabric.t_rd_min) & l2s_d(_T) & ...`

Go already tracks insertion order via `module.Module.SymbolOrder []*lg.Const` — it just isn't used at these sites.

## Fix: 3 files, 4 changes

### 1. Add `insertionOrderSymbols` helper to `check/l2s_shared.go`

Place near line 635 (after `BuildAddConstsToD`). Returns `mod.SymbolOrder` deduplicated by name (first occurrence wins), matching Python dict semantics.

```go
func insertionOrderSymbols(mod *module.Module) []*lg.Const {
    if mod == nil || len(mod.SymbolOrder) == 0 {
        return nil
    }
    seen := make(map[string]bool, len(mod.SymbolOrder))
    result := make([]*lg.Const, 0, len(mod.SymbolOrder))
    for _, sym := range mod.SymbolOrder {
        if !seen[sym.Name] {
            seen[sym.Name] = true
            result = append(result, sym)
        }
    }
    return result
}
```

### 2. Fix `check/l2s_auto.go` line 879

Replace `for _, sym := range m.Sig.Symbols` with `insertionOrderSymbols(m)`.
Change `sym.Sort` → `sym.CSort` (field name differs on `*lg.Const` vs `*SymbolEntry`).
Remove intermediate `lg.NewConst` call — `sym` is already a `*lg.Const`.

### 3. Fix `check/ranking_tactic.go` line 341

Same pattern as #2: replace `mod.Sig.Symbols` → `insertionOrderSymbols(mod)`, `sym.Sort` → `sym.CSort`, remove redundant `NewConst`.

### 4. Fix `check/l2s_shared.go` line 626 (`BuildAddConstsToD`)

Replace `sortedSymbols(mod.Sig)` (alphabetical) with `insertionOrderSymbols(mod)` (insertion order). Rest of body unchanged since `sortedSymbols` already returned `[]*lg.Const`.

## Files to modify

- `/Users/jaten/ivy/goivy/check/l2s_shared.go` — add helper; fix `BuildAddConstsToD`
- `/Users/jaten/ivy/goivy/check/l2s_auto.go` — fix map iteration at line 879
- `/Users/jaten/ivy/goivy/check/ranking_tactic.go` — fix map iteration at line 341

## Verification

Run the TestOrdLive test. The `l2s_consts_d` invariant at xtrace line 321375 should now match Python's term order. The `BuildAddConstsToD` fix prevents a subsequent divergence in action construction.
