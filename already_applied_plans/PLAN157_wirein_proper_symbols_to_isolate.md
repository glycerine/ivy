# Replace `collectUsedSymbolNames` with `SymbolsIluAst` in isolate.go

**Created:** 2026-03-31

## Context

The Go `IsolateComponent()` in `isolate/isolate.go` collects referenced symbols using `collectUsedSymbolNames` (helpers.go:404), which does a generic recursive walk of `Children()` and only special-cases `*lg.Apply` and `*lg.Const`. It does NOT replicate the Python's `symbols_ilu_ast` behavior of checking whether an app's rep is a binder and recursing into the binder's body.

The Python source of truth (ivy_isolate.py) uses `lu.used_symbols_asts(asts)` (which calls `symbols_ilu_ast`) at:
- **Line 1230**: first collection (allSyms) — formulas
- **Line 1237**: first collection — formals
- **Line 1243**: first collection — natives
- **Line 1300**: dropped axioms check
- **Line 1377**: second collection (allSyms2) — all asts
- **Line 1400**: property dependency check

The Go has 8 call sites of `collectUsedSymbolNames` in isolate.go that need to switch to the new `il.SymbolsIluAst`-based functions.

## Key observations

1. **`collectUsedSymbolNames` collects into a mutable `map[lg.NodeKey]lg.Expr`**. The new `il.UsedSymbolsAst` returns a fresh map. We need a variant that merges INTO an existing map, or we iterate the `SymbolsIluAst` generator directly.

2. **Formal params/returns**: Go adds them as direct Const entries (`allSyms[actions.ConstSymKey(p)] = p`). Python puts them in the asts list and walks them with `symbols_ilu_ast`. Since formals are bare Consts, `symbols_ilu_ast(Const)` yields the Const itself — so the Go approach is equivalent. No change needed for formals.

3. **The `ConstSymKey(c)` vs `lg.Key(n)` keying**: Both call `Sexp()`, so they produce identical keys for `*lg.Const`. For non-Const symbols (rare: would be Apply or other app-like nodes), `lg.Key` still works.

4. **`collectUsedSymbolNames` is also used in `helpers.go` for `pd.Prop.Formula` checks and `droppedAxioms`**. These also need fixing.

## Implementation Plan

### Step 1: Add `CollectSymbolsInto` helper to `isolate/helpers.go`

Add a small bridge function that collects `SymbolsIluAst` results into an existing map:

```go
// collectSymbolsInto walks node with il.SymbolsIluAst and adds
// all yielded symbols into the target map.
// This replaces collectUsedSymbolNames for Python-conformant symbol collection.
func collectSymbolsInto(node lg.Expr, syms map[lg.NodeKey]lg.Expr) {
    if node == nil {
        return
    }
    for sym := range il.SymbolsIluAst(node) {
        syms[lg.Key(sym)] = sym
    }
}
```

### Step 2: Replace all `collectUsedSymbolNames` call sites in `isolate.go`

There are 7 call sites in isolate.go:

| Line | Current | Replacement |
|------|---------|-------------|
| 925 | `collectUsedSymbolNames("lf.Formula", lf.Formula.(lg.Expr), allSyms)` | `collectSymbolsInto(lf.Formula.(lg.Expr), allSyms)` |
| 949 | `collectUsedSymbolNames("mod.Natives", expr, allSyms)` | `collectSymbolsInto(expr, allSyms)` |
| 1093 | `collectUsedSymbolNames("droppedAxioms.Formula", a.Formula.(lg.Expr), symsInAxiom)` | `collectSymbolsInto(a.Formula.(lg.Expr), symsInAxiom)` |
| 1280 | `collectUsedSymbolNames("lf.Formula", lf.Formula.(lg.Expr), allSyms2)` | `collectSymbolsInto(lf.Formula.(lg.Expr), allSyms2)` |
| 1291 | `collectUsedSymbolNames("mod.Actions", act, allSyms2)` | `collectSymbolsInto(act, allSyms2)` |
| 1302 | `collectUsedSymbolNames("mod.Natives", expr, allSyms2)` | `collectSymbolsInto(expr, allSyms2)` |
| 1309 | `collectUsedSymbolNames("mod.Proofs", n, allSyms2)` | `collectSymbolsInto(n, allSyms2)` |

### Step 3: Replace call site in `helpers.go`

| Line | Current | Replacement |
|------|---------|-------------|
| 1358 (isolate.go) | `collectUsedSymbolNames("pd.Prop.Formula", pd.Prop.Formula.(lg.Expr), propSyms)` | `collectSymbolsInto(pd.Prop.Formula.(lg.Expr), propSyms)` |

### Step 4: Keep old `collectUsedSymbolNames` but mark deprecated

Keep the old function for any other callers that might exist, but add a deprecation comment. (Or delete if no other callers exist.)

### Step 5: Add `il "github.com/glycerine/ivy/goivy/ivylogic"` import to `isolate.go`

The helpers.go already has the import. isolate.go needs it added.

### Step 6: Tests

**Unit tests** in `isolate/isolate_symbols_test.go`:

1. **TestCollectSymbolsInto_SimpleConst** — single Const yields itself
2. **TestCollectSymbolsInto_Apply** — Apply yields func + args
3. **TestCollectSymbolsInto_BinderFunc** — Apply with Lambda func: confirms binder body is expanded (the key behavioral difference)
4. **TestCollectSymbolsInto_Nil** — nil node doesn't crash
5. **TestCollectSymbolsInto_MergesIntoExisting** — confirms symbols are added to pre-populated map
6. **TestCollectSymbolsInto_ForAllBody** — ForAll formula: walks into body
7. **TestCollectSymbolsInto_NestedActions** — Action-like tree (Sequence of CallActions)

**Comparison tests** — verify old and new produce same results for non-binder cases, and new produces correct results for binder cases:

8. **TestOldVsNew_SimpleApply** — confirm same output for non-binder cases
9. **TestOldVsNew_BinderFunc** — confirm new handles binder correctly where old did not

**Randomized/fuzz test**:

10. **FuzzCollectSymbolsInto** — generate random AST trees (Apply, Const, And, ForAll, Lambda combinations), run both old and new, verify new always produces a superset of old (since new correctly expands binders that old might miss or double-count).

### Files to modify

| File | Action |
|------|--------|
| `isolate/helpers.go` | Add `collectSymbolsInto`, deprecation comment on old function |
| `isolate/isolate.go` | Add import, replace 8 call sites |
| `isolate/isolate_symbols_test.go` | **Create** — new test file |

## Verification

1. `go build ./isolate/` — must compile
2. `go test ./isolate/ -run TestCollectSymbols -v` — new tests pass
3. `go test ./isolate/ -run TestOldVsNew -v` — comparison tests pass
4. `go test ./isolate/ -run FuzzCollectSymbols -fuzz=. -fuzztime=10s` — fuzz passes
5. `go test ./isolate/ -v` — full package (no regressions)
6. `go vet ./isolate/` — clean
7. `go test ./...` — full project (no regressions)
