# Fix: Use faithful `UsedSymbolsAst` port in `formulaToZ3` instead of broken `UsedConstantsList`

**Created:** 2026-04-09 ~01:50 UTC

## Context

The golden test diverges at line 236241:
```
236241  go : XTRACE: ivy_solver.py:603 type_constraints() ENTER nsyms=2
        py : XTRACE: ivy_solver.py:603 type_constraints() ENTER nsyms=14
```

Both sides now call `type_constraints()` (previous plan fixed that), but Go passes only 2 symbols while Python passes 14. The call originates from `formula_to_z3(dfn)` (called from within `clauses_to_z3` for a definition), which calls `type_constraints(used_symbols_ast(fmla))`.

Python stack trace confirms the path:
```
clauses_to_z3(clauses)           # ivy_solver.py:641
  → formula_to_z3(dfn)           # for each definition
    → type_constraints(used_symbols_ast(fmla))  # line 725
```

## Root cause

Go's `formulaToZ3` (solver.go:477) collects symbols with:
```go
usedSyms := lu.UsedConstantsList(fmla)
```

`UsedConstantsList` calls `usedConstantsRec` (logicutil.go:201-229) which traverses nodes via `t.Children()`. But **`Apply.Children()` returns only `Terms` (arguments), NOT the `Func` field** (logic/term.go:203-210):
```go
func (a *Apply) Children() []Expr {
    // Returns Terms only — matches Python's Apply.args property
    cp := make([]Expr, len(a.Terms))
    copy(cp, a.Terms)
    return cp
}
```

This means `UsedConstantsList` **misses all function/relation symbols** that appear as `Apply.Func`. In a formula like `f(g(x,a), h(b,c))`, it would find only the leaf `*lg.Const` nodes (`a`, `b`, `c`) but miss `f`, `g`, `h`.

## Python source of truth

Python `used_symbols_ast` (ivy_logic_utils.py:610) calls `symbols_ilu_ast` (line 547-558):
```python
def symbols_ilu_ast(ast):
    if is_app(ast):
        if is_binder(ast.rep):
            for x in symbols_ilu_ast(ast.rep.body):
                yield x
        else:
            yield ast.rep          # ← yields the function symbol
    for arg in ast.args:
        for x in symbols_ilu_ast(arg):  # ← recurses into args
            yield x
```

This correctly yields the function symbol (`ast.rep`) of each Apply, PLUS recurses into arguments.

## Existing faithful port (already written, just not used)

Go already has a faithful port at `ivylogic/symbols.go`:
```go
func UsedSymbolsAst(node lg.Expr) map[lg.NodeKey]lg.Expr  // line 82
```
which uses `SymbolsIluAst` (line 39-64) — a mechanical port of Python's `symbols_ilu_ast`. It correctly calls `NodeRep(node)` to get the function symbol, then `NodeArgs(node)` to recurse into arguments.

## Plan

### 1. Replace `lu.UsedConstantsList(fmla)` with `il.UsedSymbolsAst(fmla)` in `formulaToZ3`

**File:** `z3bridge/solver.go` line 477

**Before:**
```go
// Python formula_to_z3 line 725: tcs = type_constraints(used_symbols_ast(fmla))
usedSyms := lu.UsedConstantsList(fmla)
tcs, tcErr := s.typeConstraints(usedSyms)
```

**After:**
```go
// Python formula_to_z3 line 725: tcs = type_constraints(used_symbols_ast(fmla))
symMap := il.UsedSymbolsAst(fmla)
usedSyms := make([]*lg.Const, 0, len(symMap))
for _, sym := range symMap {
    if c, ok := sym.(*lg.Const); ok {
        usedSyms = append(usedSyms, c)
    }
}
tcs, tcErr := s.typeConstraints(usedSyms)
```

`il.UsedSymbolsAst` returns `map[lg.NodeKey]lg.Expr`. We filter to `*lg.Const` because `typeConstraints` checks `.CSort` on each symbol — only `*lg.Const` has that field. This matches Python where `type_constraints` filters on `sym.sort.rng.name`.

### Why `ClausesToZ3` is OK

`ClausesToZ3` (solver.go:286) uses `clauses.Symbols()` which calls `module/clauses.go:usedSymbolsAST` → `symbolsASTRec`. This function **explicitly handles Apply.Func** (line 349-354):
```go
case *lg.Apply:
    if c, ok := t.Func.(*lg.Const); ok {
        result[lg.Key(c)] = c
    }
```
So the clauses path already collects function symbols correctly. No change needed there.

### Files to modify
- `z3bridge/solver.go` — line 477: replace `lu.UsedConstantsList(fmla)` with `il.UsedSymbolsAst(fmla)` + `*lg.Const` filter

## Verification

Run `cd ~/ivy/goivy && make golden` and confirm the nsyms count matches (14 = 14) and the divergence at 236241 is resolved.
