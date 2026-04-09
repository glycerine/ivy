# Remove unfaithful Const-only filtering — pass lg.Expr like Python does

**Created:** 2026-04-09 ~03:40 UTC

## Context

Python's `used_symbols_ast` / `symbols_ilu_ast` yields both `Const` and `Var` objects. `gen_to_set` does NO type filtering. Callers like `type_constraints` receive a mixed set and work uniformly because both types have `.sort` and `.name`.

Go's `UsedConstantsAst` filters to `*lg.Const` only, discarding `*lg.Variable` entries. This is unfaithful — Python doesn't do this filtering. The faithful Go equivalents (`UsedSymbolsAst`, `SymbolsIluAst`) already return `lg.Expr` correctly.

## Problem: `.Name` access

Python duck-types: `s.name` works on both Const and Var. Go's `lg.Expr` interface has `NodeSort()` (replaces `.CSort`/`.VSort`) but NO name accessor. Both `*lg.Const` and `*lg.Variable` have `.Name` as a field, but callers accessing it through `lg.Expr` need a type switch.

Many callers (29 for `module.UsedSymbolsAST`, 12 for `Clauses.Symbols()`) access `.Name`. Adding type switches at every site is noisy.

## Strategy: Add `ExprName(lg.Expr) string` helper function

In Python, `.name` is a duck-typed field on many `recstruct` types: `Var`, `Const`, `UninterpretedSort`, `EnumeratedSort`, `RangeSort`, `WhenOperator`, `NamedBinder`. Callers like `symbols_ilu_ast` yield objects and access `.name` without caring about the concrete type.

In Go, we add a helper function `ExprName(x lg.Expr) string` — the Go equivalent of Python's duck-typed `.name`. This is clearer and easier to audit than an interface method, since only 9 of 28 `lg.Expr` types have a `.Name` field.

```go
// ExprName returns the name of an Expr node.
// Matches Python's duck-typed .name field on Var, Const, sorts,
// WhenOperator, and NamedBinder.
func ExprName(x Expr) string {
    switch t := x.(type) {
    case *Const:
        return t.Name
    case *Variable:
        return t.Name
    case *UninterpretedSort:
        return t.Name
    case *EnumeratedSort:
        return t.Name
    case *RangeSort:
        return t.Name
    case *TopSort:
        return t.Name
    case *WhenOperator:
        return t.Name
    case *NamedBinder:
        return t.Name
    default:
        panic(fmt.Sprintf("ExprName not implemented for %T", x))
    }
}
```

Place in `logic/expr_name.go` (new file in the logic package itself, so callers use `lg.ExprName(x)`).

Similarly, Python's `.sort` works on both Const and Var. Go already has `NodeSort()` on the `Expr` interface for this.

---

## Phase 1: Add `ExprName(Expr) string` helper function in `logic/`

Add `logic/expr_name.go` with the function shown in the Strategy section above (using unqualified types since it's in the `logic` package). No interface changes, no changes to action types or other Expr implementors outside the type switch.

## Phase 2: Change `module.UsedSymbolsAST` and `Clauses.Symbols()` return types

**`module/astutil.go`:**
```go
func UsedSymbolsAST(node lg.Expr) map[lg.NodeKey]lg.Expr {
    return il.UsedSymbolsAst(node)
}

func SymbolsAST(node lg.Expr) []lg.Expr {
    m := il.UsedSymbolsAst(node)
    result := make([]lg.Expr, 0, len(m))
    for _, sym := range m {
        result = append(result, sym)
    }
    return result
}
```

**`module/clauses.go` — `Symbols()`:**
```go
func (c *Clauses) Symbols() map[lg.NodeKey]lg.Expr {
    result := make(map[lg.NodeKey]lg.Expr)
    for _, f := range c.Fmlas {
        for k, v := range il.UsedSymbolsAst(f) {
            result[k] = v
        }
    }
    for _, d := range c.Defs {
        for k, v := range il.UsedSymbolsAst(d) {
            result[k] = v
        }
    }
    return result
}
```

## Phase 3: Change `module/ops.go` wrapper functions

**`UsedSymbolsClauses`** → return `map[lg.NodeKey]lg.Expr`
**`ConstantsClauses`** → use `il.UsedSymbolsAst`, access name via `.ExprName()`
**`UsedSymbolsClausesOrdered`** → return `*iu.InsMap[lg.NodeKey, lg.Expr]`
**`UsedSymbolsExprOrdered`** → return `*iu.InsMap[lg.NodeKey, lg.Expr]`
**`SymbolsClauses`** → return `[]lg.Expr`
**`UsedSymbolNamesClauses`** → use `.ExprName()` on each sym (already returns `map[string]bool`, signature unchanged)
**`collectUsedNames`** → use `.ExprName()`

## Phase 4: Change `z3bridge/solver.go` — `typeConstraints` and helpers

**`typeConstraints(syms []lg.Expr)`** — change parameter type
**`buildConstraintTerm(sym lg.Expr)`** — use `sym.NodeSort()` instead of `sym.CSort`
**`natConstraintForSymbol(sym lg.Expr)`** — use `sym.NodeSort()`, `.ExprName()`
**`rangeConstraintsForSymbol(sym lg.Expr)`** — use `sym.NodeSort()`, `.ExprName()`

The callers in `formulaToZ3` and `ClausesToZ3` that currently filter to `*lg.Const` would pass `[]lg.Expr` directly with no filter.

## Phase 5: Change `IsInterpretedSymbol` in `ivylogic/globals.go`

Currently takes `*lg.Const`. Change to `lg.Expr`:
```go
func IsInterpretedSymbol(sig *Sig, s lg.Expr) bool {
    name := lg.ExprName(s)
    sort := s.NodeSort()
    // ... same logic using name and sort instead of s.Name and s.CSort
}
```

## Phase 6: Update all callers of changed functions

Every caller of `module.UsedSymbolsAST`, `Clauses.Symbols()`, `UsedSymbolsClauses`, `UsedSymbolsClausesOrdered`, `SymbolsClauses`, `ConstantsClauses` that accesses `.Name` or `.CSort` on the returned symbols:

- Replace `sym.Name` → `lg.ExprName(sym)` (or just `ExprName(sym)` if in ivylogic package)
- Replace `sym.CSort` → `sym.NodeSort()`
- Where callers need `*lg.Const` specifically (e.g., `lg.NewConst(sym.Name, sym.CSort)` for renaming), use type assertion: `if c, ok := sym.(*lg.Const); ok { ... }`

## Phase 7: Remove `UsedConstantsAst` and `UsedConstantsListAst`

Delete from `ivylogic/symbols.go`. These were the unfaithful Const-only filters.

Also remove any remaining inline `*lg.Const` filters in z3bridge/solver.go.

---

## Files to modify

| File | Changes |
|------|---------|
| `logic/expr_name.go` | New file: `ExprName(Expr) string` helper function |
| `ivylogic/symbols.go` | Remove `UsedConstantsAst`, `UsedConstantsListAst` |
| `module/astutil.go` | Return `lg.Expr` from `UsedSymbolsAST`, `SymbolsAST` |
| `module/clauses.go` | Return `lg.Expr` from `Symbols()`; update `IsUniversalFirstOrder`, `usesSymbolsAST` |
| `module/ops.go` | Return `lg.Expr` from `UsedSymbolsClauses`, `UsedSymbolsClausesOrdered`, `UsedSymbolsExprOrdered`, `ConstantsClauses`, `SymbolsClauses`; use `.ExprName()` in `UsedSymbolNamesClauses`, `collectUsedNames` |
| `z3bridge/solver.go` | Change `typeConstraints`, `buildConstraintTerm`, `natConstraintForSymbol`, `rangeConstraintsForSymbol` to take `lg.Expr`; remove Const filter in `formulaToZ3` and `ClausesToZ3` |
| `ivylogic/globals.go` | Change `IsInterpretedSymbol` to take `lg.Expr` |
| 29 callers of `module.UsedSymbolsAST` | Replace `.Name` → `ExprName()`, `.CSort` → `.NodeSort()`, add type assertions where `*lg.Const` is specifically needed |
| 12 callers of `Clauses.Symbols()` | Same pattern |
| 5 callers of `UsedSymbolsClauses` | Same pattern |
| 1 caller of `UsedSymbolsClausesOrdered` | Same pattern |
| `compiler/phase6.go` | Already uses `il.SymbolsIluAst` (yielding lg.Expr) — just remove Const filter |

## Verification

1. `go build ./...` — must compile
2. `go vet ./...` — no errors
3. `cd ~/ivy/goivy && make golden` — confirm divergence resolved
