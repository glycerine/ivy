# PLAN204: Fix RenameDistinct — name-only vs per-symbol renaming + set ordering

**Created:** 2026-04-04 ~23:00 UTC

## Context

PLAN203 fixed 5 non-deterministic map iteration sites. The golden test advanced from xtrace line 159811 to **166773**. The divergence:

```
Go:     name:__fml:y_a   (sort: tar_cf_clock)
Python: name:__fml:y_b   (sort: tar_cf_clock)
```

## Root Cause: TWO bugs in `RenameDistinctClauses`

### Bug 1: Name-only vs per-symbol renaming

**Python** `rename_distinct` (ivy_transrel.py:289-301):
```python
used1 = used_symbols_clauses(clauses1)  # set of Symbol objects (name+sort)
for s in used1:
    if is_skolem(s) and not is_global_skolem(s):
        map1[s] = rename(s, rn)          # per-Symbol key — separate entry per sort
return rename_clauses(clauses1, map1)    # structural rename (matches by name+sort)
```

**Go** `RenameDistinctClauses` (transrel.go:474-499):
```go
used1 := usedSymbolNamesClauses(c1)      // map[string]bool — name strings only!
for s := range used1 { ... }             // one entry per unique NAME
nameMap[s] = rn.Rename(s)               // one rename per unique name
return mod.RenameClausesByName(c1, nameMap) // name-only rename — WRONG
```

When clauses contain `__fml:y:lclock` AND `__fml:y:tar_cf_clock`:
- **Python**: calls `rn('__fml:y')` **twice** → first gets `_a`, second gets `_b`. Each sort variant gets its own new name.
- **Go**: calls `rn.Rename('__fml:y')` **once** → gets `_a`. Both sort variants get the same name.

### Bug 2: Set hash order vs deterministic order

Python iterates `used1` (a `set`) in hash order. Go iterates a Go `map` (random order, then sorted by PLAN203 fix). These differ. After fixing Bug 1 to use per-symbol renaming, the iteration order of Symbol objects must match.

**Fix**: Change Python's `used1` from `set(symbols_clauses(...))` to `dict.fromkeys(symbols_clauses(...))` (insertion order = AST depth-first traversal order). Change Go to collect symbols in the same depth-first traversal order using InsMap.

## Fix Plan

### Step 1: Add `UsedSymbolsClausesOrdered` — `module/ops.go`

New function, modeled after existing `UsedVariablesOrdered` (line 845):

```go
// UsedSymbolsClausesOrdered collects constant symbols from clauses in
// depth-first AST traversal order, deduplicating by structural identity.
// Matches Python's dict.fromkeys(symbols_clauses(c)) insertion order.
func UsedSymbolsClausesOrdered(c *Clauses) *iu.InsMap[lg.NodeKey, *lg.Const] {
    result := iu.NewInsMap[lg.NodeKey, *lg.Const]()
    if c == nil { return result }
    for _, f := range c.Fmlas {
        collectSymbolsOrdered(f, result)
    }
    for _, d := range c.Defs {
        collectSymbolsOrdered(d, result)
    }
    return result
}

func collectSymbolsOrdered(n lg.Expr, result *iu.InsMap[lg.NodeKey, *lg.Const]) {
    if n == nil { return }
    if c, ok := n.(*lg.Const); ok {
        result.Set(lg.Key(c), c)
    }
    for _, child := range n.Children() {
        collectSymbolsOrdered(child, result)
    }
}
```

### Step 2: Rewrite `RenameDistinctClauses` — `actions/transrel.go`

Change from name-only to per-symbol renaming:

```go
func RenameDistinctClauses(c1, c2 *mod.Clauses) *mod.Clauses {
    if c1 == nil { return c1 }
    // Collect symbols (name+sort) in AST traversal order
    used1 := mod.UsedSymbolsClausesOrdered(c1)
    // Build name-string set for renamer from c2
    used2 := usedSymbolNamesClauses(c2)
    used2Slice := nameSetToSlice(used2)
    rn := iu.NewUniqueRenamer("", used2Slice)
    // Iterate symbols in insertion order, building structural rename map
    constMap := make(map[lg.NodeKey]*lg.Const)
    for key, sym := range used1.All() {
        if IsSkolem(sym.Name) && !IsGlobalSkolem(sym.Name) {
            newName := rn.Rename(sym.Name)
            constMap[key] = lg.NewConst(newName, sym.CSort)
        }
    }
    if len(constMap) == 0 { return c1 }
    return mod.RenameClauses(c1, constMap)
}
```

Key changes:
- `usedSymbolNamesClauses(c1)` → `mod.UsedSymbolsClausesOrdered(c1)` (symbols not just names)
- Iterate InsMap `.All()` (depth-first order, not random map or sorted names)
- Build `constMap` with `lg.NodeKey` keys (structural identity) instead of `nameMap` with string keys
- Use `mod.RenameClauses` (structural) instead of `mod.RenameClausesByName` (name-only)

### Step 3: Rewrite `RenameDistinct` (non-Clauses version) — `actions/transrel.go`

Same transformation for the `lg.Expr` version. Need a `usedSymbolsOrdered(node lg.Expr)` helper returning InsMap, then per-symbol renaming with `mod.RenameAST`.

### Step 4: Fix Python `rename_distinct` — `ivy_transrel.py:294`

Change `used1` from set (hash order) to dict (insertion order):

```python
def rename_distinct(clauses1, clauses2):
    used1 = dict.fromkeys(symbols_clauses(clauses1))   # was: used_symbols_clauses(clauses1)
    used2 = used_symbols_clauses(clauses2)
    rn = UniqueRenamer('', used2)
    map1 = dict()
    for s in used1:
        if is_skolem(s) and not is_global_skolem(s):
            map1[s] = rename(s, rn)
    return rename_clauses(clauses1, map1)
```

Only line 294 changes: `used_symbols_clauses(clauses1)` → `dict.fromkeys(symbols_clauses(clauses1))`.

Note: `used2` stays as set — it feeds into `UniqueRenamer` which converts to a name-string set via `str()`. Order doesn't matter for the renamer's initial Used set.

### Step 5: Also fix Python `elim_dead_definitions` — `ivy_logic_utils.py:1329`

Already changed in prior edit: `set(...)` → `dict.fromkeys(...)`. Keep this change.

### Step 6: Regenerate golden test + verify

```bash
cd ~/ivy/goivy && make golden-refresh
go build ./...
go test ./module/... -count=1
go test ./actions/... -count=1
go test ./lalr_full/ -run TestOrdLive -count=1
```

## Files Modified

- `module/ops.go` — add `UsedSymbolsClausesOrdered`, `collectSymbolsOrdered`
- `actions/transrel.go` — rewrite `RenameDistinctClauses` and `RenameDistinct`
- `~/ivy/pyivy/ivy/ivy/ivy_transrel.py` — line 294: `used_symbols_clauses` → `dict.fromkeys(symbols_clauses(...))`
- `~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py` — line 1329: already changed
