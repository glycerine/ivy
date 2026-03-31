# Port `symbols_ilu_ast` from Python to Go

**Created:** 2026-03-31 (current session)

## Context

The Python `symbols_ilu_ast` (ivy_logic_utils.py:537) is a generator that recursively walks an AST and yields all function symbols. Its critical behavior: when an app node's "rep" (function head) is a binder (ForAll, Exists, Lambda, NamedBinder, Some), it recurses into the binder's body instead of yielding the binder itself.

The Go isolate package currently uses `collectUsedSymbolNames` (isolate/helpers.go:404) which does NOT handle this binder expansion — it just recurses into `app.Func` and `node.Children()` generically. This is the bug identified in commit f501e6b: the Go side doesn't replicate the Python behavior where `lu.used_symbols_asts` calls `symbols_ilu_ast` which expands binder reps.

### Python source (ivy_logic_utils.py:537-548)

```python
def symbols_ilu_ast(ast):
    if is_app(ast):
        if is_binder(ast.rep):
            for x in symbols_ilu_ast(ast.rep.body):
                yield x
        else:
            yield ast.rep
    for arg in ast.args:
        if isinstance(arg,str):
            print(arg)
        for x in symbols_ilu_ast(arg):
            yield x
```

### Derived functions (ivy_logic_utils.py:591-601)

```python
symbols_asts = apply_gen_to_list(symbols_ilu_ast)   # generator over list of ASTs
used_symbols_ast = gen_to_set(symbols_ilu_ast)       # set from one AST
used_symbols_asts = gen_to_set(symbols_clause)       # set from list of ASTs
```

### Python `.rep` mapping to Go

| Python type | `.rep` returns | Go equivalent |
|---|---|---|
| `lg.Apply` | `self.func` | `app.Func` |
| `lg.Const` (= Symbol) | `self` | the Const itself |
| `lg.NamedBinder` (0 vars) | `self` | the NamedBinder itself |

### Existing Go helpers to reuse

| Function | File | Purpose |
|---|---|---|
| `IsApp(n lg.Expr) bool` | ivylogic/ivylogic.go:118 | checks Apply, Const, or 0-var NamedBinder |
| `IsBinder(n lg.Expr) bool` | ivylogic/ivylogic.go:177 | checks ForAll, Exists, Lambda, NamedBinder, Some |
| `BinderBody(n lg.Expr) lg.Expr` | ivylogic/util.go:175 | returns body of a binder |
| `NodeArgs(n lg.Expr) []lg.Expr` | ivylogic/util.go:192 | returns args (mirrors Python `.args`) |

## Implementation Plan

### Step 1: Create `ivylogic/symbols.go`

New file with these functions:

**1a. `NodeRep(n lg.Expr) lg.Expr`** — Returns the "rep" (function head) of an app-like node.
- `*lg.Apply` -> `app.Func`
- `*lg.Const` -> the Const itself
- `*lg.NamedBinder` with 0 variables -> the NamedBinder itself
- Otherwise -> nil

**1b. `SymbolsIluAst(node lg.Expr) iter.Seq[lg.Expr]`** — Core generator. Mechanical port of Python `symbols_ilu_ast`. Uses Go 1.23+ `iter.Seq`. Internal recursive helper `symbolsIluAstRec(node lg.Expr, yield func(lg.Expr) bool) bool`.

Logic:
```
if IsApp(node):
    rep := NodeRep(node)
    if IsBinder(rep):
        recurse into BinderBody(rep)
    else:
        yield rep
for each arg in NodeArgs(node):
    recurse into arg
```

**1c. `SymbolsAsts(nodes []lg.Expr) iter.Seq[lg.Expr]`** — Applies `SymbolsIluAst` to each element of a slice. Matches Python `symbols_asts = apply_gen_to_list(symbols_ilu_ast)`.

**1d. `UsedSymbolsAst(node lg.Expr) map[lg.NodeKey]lg.Expr`** — Collects `SymbolsIluAst` into a set (map). Matches Python `used_symbols_ast = gen_to_set(symbols_ilu_ast)`.

**1e. `UsedSymbolsAsts(nodes []lg.Expr) map[lg.NodeKey]lg.Expr`** — Collects `SymbolsAsts` into a set. Matches Python `used_symbols_asts`.

**1f. `UsedSymbolsInOrderAst(node lg.Expr) []lg.Expr`** — Returns unique symbols in order of first occurrence. Matches Python `used_symbols_in_order_ast = gen_unique(symbols_ilu_ast)`.

### Step 2: Create `ivylogic/symbols_test.go`

Comprehensive tests:

1. **TestNodeRep_Apply** — Apply with Const func returns the Const
2. **TestNodeRep_Const** — Bare Const returns itself
3. **TestNodeRep_NamedBinder0Vars** — 0-var NamedBinder returns itself
4. **TestNodeRep_NamedBinderWithVars** — NamedBinder with variables returns nil
5. **TestNodeRep_Variable** — Variable returns nil (not an app)
6. **TestSymbolsIluAst_SimpleConst** — Bare constant yields itself
7. **TestSymbolsIluAst_SimpleApply** — `f(a, b)` yields f, a, b
8. **TestSymbolsIluAst_NestedApply** — `f(g(a), b)` yields f, g, a, b
9. **TestSymbolsIluAst_BinderFunc** — `Apply{Func: Lambda{body: f(x)}}` should NOT yield the Lambda, should recurse into Lambda.Body and yield f, x
10. **TestSymbolsIluAst_ForAllFormula** — `ForAll(vars, f(a) & g(b))` yields f, a, g, b (ForAll args = [body], recurse into body)
11. **TestSymbolsIluAst_NestedBinders** — ForAll containing Apply with Lambda func
12. **TestSymbolsIluAst_Variable** — Variable alone yields nothing
13. **TestSymbolsIluAst_And** — `And(f(a), g(b))` yields f, a, g, b
14. **TestSymbolsIluAst_Eq** — `Eq(a, b)` yields a, b (Eq is not an app)
15. **TestSymbolsIluAst_NamedBinder0VarsAsApp** — 0-var NamedBinder treated as app, but since IsBinder returns true, recurses into body
16. **TestUsedSymbolsAst_Dedup** — `f(a, a)` returns set with f and a (no duplicates)
17. **TestUsedSymbolsAsts_MultipleNodes** — Collects from list of ASTs
18. **TestSymbolsAsts_PreservesOrder** — Order matches depth-first traversal
19. **TestUsedSymbolsInOrderAst** — Returns unique symbols preserving first-occurrence order

### Files to create/modify

| File | Action |
|---|---|
| `ivylogic/symbols.go` | **Create** — all functions from Step 1 |
| `ivylogic/symbols_test.go` | **Create** — all tests from Step 2 |

No existing files need modification.

## Verification

1. `cd ~/ivy/goivy && go build ./ivylogic/` — must compile
2. `cd ~/ivy/goivy && go test ./ivylogic/ -run TestNodeRep -v` — NodeRep tests pass
3. `cd ~/ivy/goivy && go test ./ivylogic/ -run TestSymbolsIluAst -v` — all generator tests pass
4. `cd ~/ivy/goivy && go test ./ivylogic/ -run TestUsedSymbols -v` — set-collection tests pass
5. `cd ~/ivy/goivy && go test ./ivylogic/ -v` — full package passes (no regressions)
6. `cd ~/ivy/goivy && go vet ./ivylogic/` — no vet warnings
