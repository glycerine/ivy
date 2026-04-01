# Fix allSyms.add Trace Ordering: Insertion-Ordered Containers

**Created:** 2026-04-01 (current session)

## Context

At XTRACE line 151058, Go and Python diverge:
- Go: `isolate.allSyms.add index.succ`
- Python: `isolate.allSyms.add <`

The root cause is that Python's `_traced_add_syms` iterates a `set()` (hash-table order)
while Go's `collectSymbolsInto` iterates the `SymbolsIluAst` generator (AST traversal order).

**Fix**: Use insertion-ordered containers throughout:
- Go: `*iu.InsMap[lg.NodeKey, lg.Expr]` everywhere `map[lg.NodeKey]lg.Expr` is used for symbol sets
- Python: `OrderedSymSet` (dict subclass with `.add()`) instead of `set()`; pass generators instead of sets to `_traced_add_syms`

Additionally, change `traceSymSet` (Go) and `_trace_sym_set` (Python) to iterate in
**insertion order** (not sorted). This is the "stronger check": it verifies the sequence
of symbol collection, not just the final sorted content.

## File 1: Python — `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_isolate.py`

### 1a. Add `OrderedSymSet` class near line 43 (before `_traced_add_syms`):

```python
class OrderedSymSet(dict):
    """Insertion-ordered symbol container with set-like .add() API.
    Keys are the symbols; values are None. Iteration yields keys in insertion order.
    Compatible with Python Ivy internals that call .add() on symbol sets."""
    def add(self, sym):
        self.setdefault(sym, None)
    def update(self, syms):
        for s in syms:
            self.setdefault(s, None)
```

### 1b. Change `_trace_sym_set` to iterate in insertion order (NOT sorted):

Old (line 38-41):
```python
for x in sorted(syms, key=lambda x: str(x)):
    xtracer.trace("%s.sym %s" % (label, str(x)))
```
New:
```python
for x in syms:
    xtracer.trace("%s.sym %s" % (label, str(x)))
```

### 1c. Change allSyms1 accumulator and callers:

Line 1258: `all_syms_raw = set()` → `all_syms_raw = OrderedSymSet()`
Line 1266: `lu.used_symbols_ast(y.formula)` → `lu.symbols_ilu_ast(y.formula)`
Line 1270: `lu.used_symbols_asts(a.formal_params)` → `lu.symbols_asts(a.formal_params)`
Line 1271: `lu.used_symbols_asts(a.formal_returns)` → `lu.symbols_asts(a.formal_returns)`
Line 1275: `lu.used_symbols_asts(tmp.args[2:])` → `lu.symbols_asts(tmp.args[2:])`
Line 1278: `all_syms = set()` → `all_syms = OrderedSymSet()`
Line 1284: `all_syms.add(nsym)` stays (OrderedSymSet.add works)

### 1d. Change allSyms2 accumulator and callers:

Line 1398: `all_syms = set()` → `all_syms = OrderedSymSet()`
Line 1404: `lu.used_symbols_ast(y.formula)` → `lu.symbols_ilu_ast(y.formula)`
Line 1408: `lu.used_symbols_ast(action)` → `lu.symbols_ilu_ast(action)`
Line 1412: `lu.used_symbols_asts(mod.params)` → `lu.symbols_asts(mod.params)`
Line 1416: `lu.used_symbols_asts(a.formal_params)` → `lu.symbols_asts(a.formal_params)`
Line 1417: `lu.used_symbols_asts(a.formal_returns)` → `lu.symbols_asts(a.formal_returns)`
Line 1421: `lu.used_symbols_asts(tmp.args[2:])` → `lu.symbols_asts(tmp.args[2:])`
Line 1425: `lu.used_symbols_ast(x[1])` → `lu.symbols_ilu_ast(x[1])`

Note: `follow_definitions_rec` (line 875) calls `all_syms.add(sym)` — works via OrderedSymSet.add.
Note: `collect_sort_destructors` (line 903) calls `res.add(dstr)` — works via OrderedSymSet.add.
Note: `action.get_references(all_syms)` (line 1289) calls `refs.add(sym)` — works via OrderedSymSet.add.

---

## File 2: Go — `isolate/helpers.go`

Change ALL `map[lg.NodeKey]lg.Expr` parameters/returns to `*iu.InsMap[lg.NodeKey, lg.Expr]`.
Change all map operations to InsMap API. Drop `"sort"` import.

### Function signatures to update:
- `traceSymSet(label string, syms *iu.InsMap[lg.NodeKey, lg.Expr])`
- `collectSymbolsInto(label string, node lg.Expr, syms *iu.InsMap[lg.NodeKey, lg.Expr])`
- `normalizeSymbolKeys(syms *iu.InsMap[lg.NodeKey, lg.Expr], ...)`
- `allSymsNameSet(syms *iu.InsMap[lg.NodeKey, lg.Expr]) map[string]bool`
- `symSetContainsName(syms *iu.InsMap[lg.NodeKey, lg.Expr], name string) bool`
- `copySymSet(s *iu.InsMap[lg.NodeKey, lg.Expr]) *iu.InsMap[lg.NodeKey, lg.Expr]`
- `FollowDefinitions(ldfs []*ast.LabeledFormula, allSyms *iu.InsMap[lg.NodeKey, lg.Expr])`
- `FollowDefinitionsLabeled(label string, ldfs []*ast.LabeledFormula, allSyms *iu.InsMap[lg.NodeKey, lg.Expr])`
- `followDefinitionsRec(key lg.NodeKey, expr lg.Expr, dmap map[lg.NodeKey]lg.Expr, allSyms *iu.InsMap[lg.NodeKey, lg.Expr], memo map[lg.NodeKey]bool)`
- `usedSymbolExprs(node lg.Expr) *iu.InsMap[lg.NodeKey, lg.Expr]`

Note: `followDefinitionsRec`'s `dmap` stays as `map[lg.NodeKey]lg.Expr` (local, not the accumulator).

### Key operation translations:
- `syms[key] = sym` → `syms.Set(key, sym)`
- `_, exists := syms[key]` → `_, exists := syms.Get2(key)`
- `for k, v := range syms` → `for k, v := range syms.All()`
- `len(syms)` → `syms.Len()`
- `delete(syms, key)` → `syms.Delkey(key)`
- `make(map[lg.NodeKey]lg.Expr)` → `iu.NewInsMap[lg.NodeKey, lg.Expr]()`

### `traceSymSet`: iterate in insertion order (NOT sorted), remove sort.Strings:
```go
func traceSymSet(label string, syms *iu.InsMap[lg.NodeKey, lg.Expr]) {
    for _, v := range syms.All() {
        xtracer.Trace("%s.sym %s", label, lg.PrettyFmla(v))
    }
    xtracer.Trace("%s n=%d", label, syms.Len())
}
```

### `normalizeSymbolKeys`: collect changes first (safe for InsMap mutation):
```go
func normalizeSymbolKeys(syms *iu.InsMap[lg.NodeKey, lg.Expr], usePolymorphicMacros bool, iuCfg *iu.IvyUtilsConfig) {
    if !usePolymorphicMacros { return }
    type change struct{ oldKey, newKey lg.NodeKey; expr lg.Expr }
    var changes []change
    for key, expr := range syms.All() {
        if c, ok := expr.(*lg.Const); ok {
            if canonical, ok := il.PolymorphicMacrosMap[c.Name]; ok {
                normalized := il.NormalizeSymbol(c, iuCfg)
                if normalized.Name == c.Name { normalized = lg.NewConst(canonical, c.CSort) }
                changes = append(changes, change{key, actions.ConstSymKey(normalized), normalized})
            }
        }
    }
    for _, ch := range changes {
        syms.Delkey(ch.oldKey)
        syms.Set(ch.newKey, ch.expr)
    }
}
```

### `copySymSet`: return new InsMap in insertion order:
```go
func copySymSet(s *iu.InsMap[lg.NodeKey, lg.Expr]) *iu.InsMap[lg.NodeKey, lg.Expr] {
    c := iu.NewInsMap[lg.NodeKey, lg.Expr]()
    for k, v := range s.All() { c.Set(k, v) }
    return c
}
```

### `collectSymbolsInto`: use `.Get2()` and `.Set()`:
```go
func collectSymbolsInto(label string, node lg.Expr, syms *iu.InsMap[lg.NodeKey, lg.Expr]) {
    if node == nil { return }
    for sym := range il.SymbolsIluAst(node) {
        key := lg.Key(sym)
        if xtracer.Enabled {
            if _, exists := syms.Get2(key); !exists {
                xtracer.Trace("%s.add %s", label, lg.PrettyFmla(sym))
            }
        }
        syms.Set(key, sym)
    }
}
```

---

## File 3: Go — `actions/transforms.go`

Change these function signatures from `map[lg.NodeKey]lg.Expr` to `*iu.InsMap[lg.NodeKey, lg.Expr]`:
- `GetReferencesInto`
- `referencesRec`
- `collectSymbols`
- `assignRefs`
- `EraseUnrefed`

Update all direct map operations inside these functions using the same translation table as above.

`iu` is already imported in this file.

---

## File 4: Go — `isolate/isolate.go`

- Line 959: `allSyms := make(map[lg.NodeKey]lg.Expr)` → `allSyms := iu.NewInsMap[lg.NodeKey, lg.Expr]()`
- Line 1351: `allSyms2 := make(map[lg.NodeKey]lg.Expr)` → `allSyms2 := iu.NewInsMap[lg.NodeKey, lg.Expr]()`
- All `allSyms[key]` / `allSyms2[key]` operations → use InsMap API
- All `len(allSyms)` / `len(allSyms2)` → `.Len()`
- All `for _, sym := range allSyms` → `for _, sym := range allSyms.All()`
- All lookups `_, exists := allSyms[key]` → `_, exists := allSyms.Get2(key)`

Also verify `iu` is already imported (it is — used for `IvyUtilsConfig`).

---

## File 5: Go — `isolate/isolate_symbols_test.go`

Update test helper `syms := make(map[lg.NodeKey]lg.Expr)` → `syms := iu.NewInsMap[lg.NodeKey, lg.Expr]()`
Pass InsMap to `collectSymbolsInto`. Update any map-style accesses in tests.

---

## Verification

```
cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy/lalr_full && go test -v -run TestVerboseOrdLive
```

Trace comparison should pass through line 151058 (and the `_trace_sym_set` / `traceSymSet` dumps
should now be in insertion order on both sides, providing a stronger structural check).

Also run:
```
cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate && go test ./...
cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions && go test ./...
```
