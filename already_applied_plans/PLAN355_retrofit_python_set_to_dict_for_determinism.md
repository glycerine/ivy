# Plan: Fix Non-Determinism in mc/toaiger Pipeline

**Created:** 2026-04-30 18:50 UTC

## Context

The rfn test (`make rfn`) compares Go XTRACE output line-by-line against Python for `ord_live.ivy` isolate `rfn.abs.iso`. Both diverge at line 225142 (`mc.ToAiger postQelim`). Between two Go runs (log.rfn vs log.rfn.prev), Go itself produces **different output** — Eq argument ordering flips and variable names change. This makes debugging the Go/Python functional divergence (nTRfmlas=782 vs 512) impossible since each run produces different Go output.

**Root cause:** Go `map` iteration is deliberately randomized. `UsedSymbolsAst` returns `map[lg.NodeKey]lg.Expr` and callers iterate it non-deterministically. This feeds into `MineConstants` → `Qelim.GetConsts` → `cartesianProduct` → substitution order → formula structure.

**Note:** The nTRfmlas count difference (782 Go vs 512 Python) is a separate functional divergence bug, not caused by non-determinism. It remains after this fix but becomes debuggable.

## Approach: InsMap for Insertion-Order Determinism

Per project convention: use `ivyutils.InsMap` to match Python dict insertion-ordering. Since Python's `used_symbols_ast` currently returns `set()` (non-deterministic), retrofit Python side to `dict.fromkeys()` first, then use InsMap in Go. Both sides then iterate in DFS first-occurrence order.

---

## Phase 0: Python-Side Retrofit

**Goal:** Change `used_symbols_ast` from `set()` to insertion-ordered `dict`.

### 0.1 Add `gen_to_ordered_dict` helper
**File:** `~/ivy/pyivy/ivy/ivy/ivy_utils.py` (after line 42)
```python
def gen_to_ordered_dict(gen):
    """Apply generator, return result as ordered dict (deduped, insertion order)."""
    return lambda *args: dict.fromkeys(gen(*args))
```

### 0.2 Change `used_symbols_ast` and siblings
**File:** `~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py` (line 659-661)
```python
# was: used_symbols_ast = gen_to_set(symbols_ilu_ast)
used_symbols_ast = gen_to_ordered_dict(symbols_ilu_ast)
used_symbols_asts = used_symbols_clause = gen_to_ordered_dict(symbols_clause)
used_symbols_clauses = gen_to_ordered_dict(symbols_clauses)
```

All callers are compatible: `for sym in dict`, `dict.update(dict)`, `sym in dict` all work like `set`.

### 0.3 Fix `funs` in ivy_mc.py (lines 1209-1216)
Change from `set()` to `dict()` since it calls `.update(used_symbols_ast(...))`:
```python
funs = dict()
for df in trans.defs:
    funs.update(used_symbols_ast(df.args[1]))
for fmla in trans.fmlas:
    funs.update(used_symbols_ast(fmla))
funs.update(used_symbols_ast(invariant))
funs = dict((sym, None) for sym in funs if il.is_function_sort(sym.sort))
```

---

## Phase 1: Core Go Return Type Change

**Goal:** Change `UsedSymbolsAst` and wrappers to return `*InsMap[lg.NodeKey, lg.Expr]`.

### 1.1 `ivylogic/symbols.go:82` — `UsedSymbolsAst`
Change return type from `map[lg.NodeKey]lg.Expr` to `*iu.InsMap[lg.NodeKey, lg.Expr]`. Use `InsMap.Set(k, sym)` instead of map assignment. DFS generator `SymbolsIluAst` already produces symbols in first-occurrence order; InsMap preserves this.

Same for `UsedSymbolsAsts` (line 92).

### 1.2 `module/astutil.go:29` — `UsedSymbolsAST` wrapper
Change return type to `*iu.InsMap[lg.NodeKey, lg.Expr]`. Pass through from `il.UsedSymbolsAst`.

### 1.3 `module/clauses.go` — `Clauses.Symbols()`
Change return type to InsMap. Use `.All()` for iteration when merging symbols from sub-calls.

### 1.4 `module/ops.go:1173` — `SymbolsClauses`
Iterate `clauses.Symbols().All()` to build the `[]lg.Expr` result.

### 1.5 All other callers (mechanical)
Every `for k, v := range il.UsedSymbolsAst(...)` becomes `for k, v := range il.UsedSymbolsAst(...).All()`. The compiler catches every site. Key files:
- `actions/transrel.go` (lines 329, 379, 1648)
- `compiler/ivy_compile.go` (lines 1447, 1633)
- `proof/checker.go`, `proof/phase5_goals.go`, `proof/phase5_matching.go`, `proof/goal.go`
- `z3bridge/solver.go`, `z3bridge/solver_clauses.go`
- `check/check.go`, `check/vmt.go`
- `module/ops.go` (lines 830, 1149, 1159)

---

## Phase 2: Update mc/ Callers

### 2.1 `mc/mine.go` — MineConstants (line 32)
Iterate `syms.All()` instead of bare map range.

### 2.2 `mc/mine.go` — MineConstants2 (lines 63, 81)
Same: iterate `.All()`. `SymbolsClauses` returns `[]lg.Expr` (already deterministic after Phase 1.4).

### 2.3 `mc/mine.go` — PrevExpr (lines 119, 122, 147)
`symsMap` is now InsMap. Iterate `.All()`.

### 2.4 `mc/toaiger.go` — invarSyms (line 119)
Type changes from `map[lg.NodeKey]lg.Expr` to `*InsMap`. Update `invarSyms[k] = sym` → `invarSyms.Set(k, sym)` at line 285. Update iteration at line 329 to `.All()`.

### 2.5 `mc/toaiger.go` — origSyms collection (lines 200-204)
Iterate `.All()` for `UsedSymbolsAST(invariant)`.

### 2.6 `mc/toaiger.go` — allSyms/funs collection (lines 212-233)
Iterate `.All()` on each `UsedSymbolsAST` call. Change `allSyms` to `*InsMap[string, lg.Expr]` and `funs` to `*InsMap[string, *lg.Const]` so downstream iteration in `InstantiateAxioms`/`matchSchemaPrems` is deterministic.

### 2.7 `mc/toaiger.go` — fromAssertTerms (line 277)
Iterate `module.UsedSymbolsAST(errCondsConj).All()`.

### 2.8 `mc/toaiger.go` — isImmutableExpr (line 334)
Iterate `syms.All()`.

---

## Phase 3: PropAbs.Map to InsMap

### 3.1 `mc/propabs.go:27`
Change `Map map[string]*lg.Const` to `Map *iu.InsMap[string, *lg.Const]`.

### 3.2 Update NewPropAbs (line 44)
Use `iu.NewInsMap[string, *lg.Const]()`.

### 3.3 Update newProp (line 59)
`pa.Map[key]` → `pa.Map.Get2(key)`. `pa.Map[key] = res` → `pa.Map.Set(key, res)`.

### 3.4 Update toaiger.go iterations (lines 357, 557)
`for exprKey, v := range propAbs.Map` → `for exprKey, v := range propAbs.Map.All()`.

---

## Phase 4: iteCtr Global to Config

### 4.1 Remove global from `mc/qelim.go:14`
Delete `var iteCtr int64` and `func NextIteCtr()`.

### 4.2 Add `McIteCtr int64` to `module.Config`
**File:** `module/config.go` — add field to Config struct.

### 4.3 Thread config through ElimIte/ElimIteKey
Change signatures to accept `*module.Config`. Update call sites in `toaiger.go` (lines 259-268) to pass `mod.Cfg`. Update tests in `mc/mc_test.go` to create fresh config.

---

## Phase 5: Sort MineConstants Results

**File:** `mc/mine.go`

After collecting constants in `MineConstants` and `MineConstants2`, sort each `[]*lg.Const` slice by `sym.Name`:
```go
for _, consts := range res {
    sort.Slice(consts, func(i, j int) bool {
        return consts[i].Name < consts[j].Name
    })
}
```
This is defensive — even if InsMap provides deterministic order, sorting guarantees canonical quantifier instantiation order.

---

## Implementation Order

```
Phase 0 (Python retrofit)        — independent, do first
Phase 1.1-1.4 (return type)      — must be first on Go side  
Phase 1.5 + Phase 2 (callers)    — same commit as Phase 1 (compile errors)
Phase 3 (PropAbs.Map)            — independent of Phase 1
Phase 4 (iteCtr)                 — independent of Phase 1
Phase 5 (sort mine results)      — after Phase 2
```

Phases 1+1.5+2 should be one commit (changing return type breaks all callers).

---

## Verification

1. `cd ~/ivy/goivy && make test` — compilation and existing tests pass
2. `cd ~/ivy/goivy && make rfn` twice — diff log.rfn vs log.rfn.prev should show identical Go output (no Eq ordering or variable name differences)
3. The Go/Python divergence at 225142 should now be stable and debuggable (nTRfmlas=782 vs 512 is a separate bug)

---

## Critical Files

- `ivylogic/symbols.go` — core return type change
- `module/astutil.go` — wrapper return type
- `mc/mine.go` — MineConstants determinism
- `mc/toaiger.go` — pipeline map iterations
- `mc/propabs.go` — PropAbs.Map
- `mc/qelim.go` — iteCtr global
- `~/ivy/pyivy/ivy/ivy/ivy_utils.py` — Python gen_to_ordered_dict
- `~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py` — Python used_symbols_ast
