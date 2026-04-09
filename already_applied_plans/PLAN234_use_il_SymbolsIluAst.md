# Comprehensive: Replace all unfaithful symbol collection with faithful `SymbolsIluAst` port

**Created:** 2026-04-09 ~02:30 UTC

## Context

Go has THREE implementations of "collect symbols from AST":

1. **`lu.UsedConstants`/`lu.UsedConstantsList`** (logicutil.go) — **BROKEN**. Traverses via `Children()` which for `Apply` returns only `Terms`, missing `Func`. Misses all function/relation symbols.

2. **`module.UsedSymbolsAST`/`clauses.Symbols()`/etc.** (module/clauses.go, astutil.go, ops.go) — **Works but unfaithful**. Handles `Apply.Func` explicitly but doesn't match Python's `symbols_ilu_ast` algorithm (no binder-body recursion). Note: `collectSymbolsOrdered` in ops.go uses `Children()` so is also broken.

3. **`il.SymbolsIluAst`/`il.UsedSymbolsAst`** (ivylogic/symbols.go) — **Faithful port** of Python's `symbols_ilu_ast`. Uses `IsApp`/`NodeRep`/`NodeArgs`. Only 6 call sites currently use it.

Python uses ONE implementation (`symbols_ilu_ast`) everywhere. Go should too.

## Strategy

All callers need `*lg.Const`. The faithful `il.SymbolsIluAst` yields `lg.Expr` (Go type-system artifact — in Python, Symbol IS the only type). So:

1. **Add** `*lg.Const`-filtered wrappers in `ivylogic/symbols.go`
2. **Rewrite** module-level wrapper functions to delegate to faithful traversal (fixes all 47 Category 2 callers automatically — their signatures don't change)
3. **Replace** 11 Category 1 call sites (`lu.UsedConstants*`) with the new faithful functions
4. **Delete** all unfaithful implementations so they can't be reused by mistake

---

## Phase 1: Add Const-filtered helpers in `ivylogic/symbols.go`

```go
// UsedConstantsAst returns the set of constant symbols in an AST.
// Const-filtered version of UsedSymbolsAst. In Python, used_symbols_ast
// returns Symbol objects directly; this filter is a Go type-system artifact.
func UsedConstantsAst(node lg.Expr) map[lg.NodeKey]*lg.Const {
    result := make(map[lg.NodeKey]*lg.Const)
    for sym := range SymbolsIluAst(node) {
        if c, ok := sym.(*lg.Const); ok {
            result[lg.Key(c)] = c
        }
    }
    return result
}

// UsedConstantsListAst returns used constants as a slice (convenience).
func UsedConstantsListAst(node lg.Expr) []*lg.Const {
    m := UsedConstantsAst(node)
    result := make([]*lg.Const, 0, len(m))
    for _, c := range m {
        result = append(result, c)
    }
    return result
}
```

---

## Phase 2: Rewrite module-level wrappers (fixes all Category 2 callers)

### 2a. `module/astutil.go` — `UsedSymbolsAST` and `SymbolsAST`

**Before:** delegates to `usedSymbolsAST` (the unfaithful local function)
**After:** delegates to `il.UsedConstantsAst`

```go
func UsedSymbolsAST(node lg.Expr) map[lg.NodeKey]*lg.Const {
    return il.UsedConstantsAst(node)
}

func SymbolsAST(node lg.Expr) []*lg.Const {
    return il.UsedConstantsListAst(node)
}
```

This fixes all 29 callers of `module.UsedSymbolsAST` and any callers of `SymbolsAST`.

### 2b. `module/clauses.go` — `Symbols()` method

**Before:** iterates fmlas/defs calling `usedSymbolsAST` (unfaithful)
**After:** iterates fmlas/defs calling `il.UsedConstantsAst` (faithful)

```go
func (c *Clauses) Symbols() map[lg.NodeKey]*lg.Const {
    result := make(map[lg.NodeKey]*lg.Const)
    for _, f := range c.Fmlas {
        for k, v := range il.UsedConstantsAst(f) {
            result[k] = v
        }
    }
    for _, d := range c.Defs {
        for k, v := range il.UsedConstantsAst(d) {
            result[k] = v
        }
    }
    return result
}
```

This fixes all 12 callers of `clauses.Symbols()`.

### 2c. `module/clauses.go` — `IsUniversalFirstOrder()`

**Before:** calls `usedSymbolsAST(f)`
**After:** calls `il.UsedConstantsAst(f)`

### 2d. `module/clauses.go` — `usesSymbolsAST()`

**Before:** calls `usedSymbolsAST(node)`
**After:** calls `il.UsedConstantsAst(node)`

Actually, keep the function but change its internals. The key check (`used[s]`) works with NodeKey which both implementations use.

### 2e. `module/ops.go` — `UsedSymbolsClauses()`

**Before:** calls `UsedSymbolsAST(f)` (from astutil.go)
**After:** already fixed by 2a (astutil.go delegates to faithful port). No change needed here.

### 2f. `module/ops.go` — `UsedSymbolsClausesOrdered()`

**Before:** calls `collectSymbolsOrdered` which uses broken `Children()` traversal
**After:** rewrite to use `il.SymbolsIluAst` iterator, preserving insertion order

```go
func UsedSymbolsClausesOrdered(c *Clauses) *iu.InsMap[lg.NodeKey, *lg.Const] {
    result := iu.NewInsMap[lg.NodeKey, *lg.Const]()
    if c == nil {
        return result
    }
    for _, f := range c.Fmlas {
        for sym := range il.SymbolsIluAst(f) {
            if c, ok := sym.(*lg.Const); ok {
                k := lg.Key(c)
                if !result.Has(k) {
                    result.Set(k, c)
                }
            }
        }
    }
    for _, d := range c.Defs {
        for sym := range il.SymbolsIluAst(d) {
            if c, ok := sym.(*lg.Const); ok {
                k := lg.Key(c)
                if !result.Has(k) {
                    result.Set(k, c)
                }
            }
        }
    }
    return result
}
```

### 2g. `module/ops.go` — `UsedSymbolNamesClauses()`

**Before:** calls `collectSymbolNamesFromNode` (custom traversal)
**After:** use `il.SymbolsIluAst` for name extraction

```go
func UsedSymbolNamesClauses(clauses *Clauses) map[string]bool {
    if clauses == nil {
        return make(map[string]bool)
    }
    result := make(map[string]bool)
    for _, f := range clauses.Fmlas {
        for sym := range il.SymbolsIluAst(f) {
            if c, ok := sym.(*lg.Const); ok {
                result[c.Name] = true
            }
        }
    }
    for _, d := range clauses.Defs {
        for sym := range il.SymbolsIluAst(d) {
            if c, ok := sym.(*lg.Const); ok {
                result[c.Name] = true
            }
        }
    }
    return result
}
```

### 2h. `module/ops.go` — `ConstantsClauses()`

**Before:** calls `lu.UsedConstants` (the broken logicutil version)
**After:** calls `il.UsedConstantsAst`

```go
func ConstantsClauses(clauses *Clauses) []*lg.Const {
    if clauses == nil {
        return nil
    }
    seen := make(map[string]bool)
    var result []*lg.Const
    for _, f := range clauses.Fmlas {
        for _, cc := range il.UsedConstantsAst(f) {
            if !seen[cc.Name] {
                seen[cc.Name] = true
                result = append(result, cc)
            }
        }
    }
    for _, d := range clauses.Defs {
        for _, cc := range il.UsedConstantsAst(d) {
            if !seen[cc.Name] {
                seen[cc.Name] = true
                result = append(result, cc)
            }
        }
    }
    return result
}
```

### 2i. `module/ops.go` — `collectUsedNames()`

**Before:** calls `cls.Symbols()` → automatically fixed by 2b. No change needed.

### 2j. `module/ops.go` — `ClausesUsingSymbols()` (line 625)

**Before:** calls `usesSymbolsAST()` → fixed by 2d. No change needed.

---

## Phase 3: Replace Category 1 call sites (`lu.UsedConstants*`)

| # | File:Line | Before | After |
|---|-----------|--------|-------|
| 1 | `compiler/ivy_compile.go:1421` | `lu.UsedConstantsList(expr)` | `il.UsedConstantsListAst(expr)` |
| 2 | `compiler/ivy_compile.go:1602` | `lu.UsedConstantsList(rhs)` | `il.UsedConstantsListAst(rhs)` |
| 3 | `ivylogic/classify_ext.go:186` | `lu.UsedConstants(term)` | `UsedConstantsAst(term)` (same pkg) |
| 4 | `ivylogic/classify_ext.go:196` | `lu.UsedConstants(term)` | `UsedConstantsAst(term)` (same pkg) |
| 5 | `ivylogic/sortinfer.go:90` | `lu.UsedConstants(term)` | `UsedConstantsAst(term)` (same pkg) |
| 6 | `proof/goal.go:219` | `lu.UsedConstants(fmla)` | `il.UsedConstantsAst(fmla)` |
| 7 | `module/ops.go:1101` | `lu.UsedConstants(f)` | `il.UsedConstantsAst(f)` (handled in 2h) |
| 8 | `module/ops.go:1109` | `lu.UsedConstants(d)` | `il.UsedConstantsAst(d)` (handled in 2h) |
| 9 | `webui/concept_isession.go:152` | `logicutil.UsedConstants(formula)` | `il.UsedConstantsAst(formula)` |
| 10 | `webui/concept_isession.go:158` | `logicutil.UsedConstants(c.Formula)` | `il.UsedConstantsAst(c.Formula)` |
| 11 | `webui/concept_isession.go:307` | `logicutil.UsedConstantsList(...)` | `il.UsedConstantsListAst(...)` |
| 12 | `webui/concept_domain.go:1428` | `logicutil.UsedConstants(diagram)` | `il.UsedConstantsAst(diagram)` |
| 13 | `webui/concept_domain.go:1534` | `logicutil.UsedConstants(stateFormula)` | `il.UsedConstantsAst(stateFormula)` |

For `compiler/ivy_compile.go`, add import: `il "github.com/glycerine/ivy/goivy/ivylogic"`
For `proof/goal.go`, add import: `il "github.com/glycerine/ivy/goivy/ivylogic"`
For `webui/*.go`, change import from `logicutil` to `il` (or add `il` if `logicutil` is still needed for other functions).
For `ivylogic/*.go`, no import needed — same package.

---

## Phase 4: Delete unfaithful functions

### 4a. `logicutil/logicutil.go` — delete:
- `UsedConstants` (lines 195-199)
- `usedConstantsRec` (lines 201-229)
- `UsedConstantsList` (lines 662-670)

### 4b. `module/clauses.go` — delete:
- `usedSymbolsAST` (lines 339-343)
- `symbolsASTRec` (lines 345-363)
- `IterSymbolsAST` (lines 369-373)
- `iterSymbolsRec` (lines 375-402)

### 4c. `module/ops.go` — delete:
- `collectSymbolsOrdered` (lines 931-939)
- `collectSymbolNamesFromNode` (lines 613-623)

---

## Phase 5: Clean up imports

Remove `lu "...logicutil"` imports from files that no longer use any `lu.*` functions:
- `ivylogic/classify_ext.go` — check if `lu` is used elsewhere
- `ivylogic/sortinfer.go` — check if `lu` is used elsewhere
- `module/astutil.go` — remove `lu` import if unused
- `module/ops.go` — remove `lu` import if unused
- `webui/concept_isession.go` — remove `logicutil` if unused
- `webui/concept_domain.go` — remove `logicutil` if unused

Add `il "...ivylogic"` imports where needed:
- `compiler/ivy_compile.go`
- `proof/goal.go`
- `webui/concept_isession.go`
- `webui/concept_domain.go`

---

## Files to modify

| File | Changes |
|------|---------|
| `ivylogic/symbols.go` | Add `UsedConstantsAst`, `UsedConstantsListAst` |
| `module/astutil.go` | Rewrite `UsedSymbolsAST`, `SymbolsAST` to delegate to `il.UsedConstantsAst` |
| `module/clauses.go` | Rewrite `Symbols()`, `IsUniversalFirstOrder()`, `usesSymbolsAST()`; delete `usedSymbolsAST`, `symbolsASTRec`, `IterSymbolsAST`, `iterSymbolsRec` |
| `module/ops.go` | Rewrite `UsedSymbolsClausesOrdered`, `UsedSymbolNamesClauses`, `ConstantsClauses`; delete `collectSymbolsOrdered`, `collectSymbolNamesFromNode` |
| `logicutil/logicutil.go` | Delete `UsedConstants`, `usedConstantsRec`, `UsedConstantsList` |
| `compiler/ivy_compile.go` | Replace `lu.UsedConstantsList` → `il.UsedConstantsListAst` (2 sites) |
| `ivylogic/classify_ext.go` | Replace `lu.UsedConstants` → `UsedConstantsAst` (2 sites) |
| `ivylogic/sortinfer.go` | Replace `lu.UsedConstants` → `UsedConstantsAst` (1 site) |
| `proof/goal.go` | Replace `lu.UsedConstants` → `il.UsedConstantsAst` (1 site) |
| `webui/concept_isession.go` | Replace `logicutil.UsedConstants*` → `il.UsedConstantsAst` (3 sites) |
| `webui/concept_domain.go` | Replace `logicutil.UsedConstants` → `il.UsedConstantsAst` (2 sites) |
| `z3bridge/solver.go` | Already fixed (lines 290, 295, 492) |

## Verification

1. `cd ~/go/src/github.com/glycerine/ivy/goivy && go build ./...` — must compile
2. `grep -r 'UsedConstants\b\|usedConstantsRec\|UsedConstantsList\|usedSymbolsAST\b\|symbolsASTRec\|IterSymbolsAST\|iterSymbolsRec\|collectSymbolsOrdered\|collectSymbolNamesFromNode' --include='*.go' .` — must return zero matches (confirms deletion)
3. `cd ~/ivy/goivy && make golden` — confirm divergence at 236241 is resolved
