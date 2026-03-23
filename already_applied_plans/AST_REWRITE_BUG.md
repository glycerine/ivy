# Fix: Action Compilation Failures — `local needs variables and body` + `unknown type: t`

**Created: 2026-03-22**

## Context

xtracer revealed Go registers only ~26 actions vs Python's ~50 for ord_live.ivy. The missing actions fail `CompileAction` with two categories of error:

1. **`local needs variables and body`** — affects `cfabric.step`, `arm.step_north`, all user actions with `var` in body
2. **`unknown type: t`** — affects `index.next`, `lclock.prev`, all module-expanded actions with alias types

## Bug A: Missing `lowerVarStatements` (Critical — 17 actions affected)

### Root Cause

Python has `lower_var_stmts()` (ivy_parser.py:2324-2350) that runs after parsing a `{ sequence }` block. It transforms:
```
var p : proc;
var m : mem_type;
wr := false;
```
into:
```
local p : proc {
  local m : mem_type {
    wr := false
  }
}
```

Each `var` becomes a `local` that wraps all subsequent statements as its body. Python does this in the parser's `p_sequence_lcb_actseq_rcb` rule (line 2357-2369): `stmts = lower_var_stmts(p[2])`.

Go's `parseSequence()` does NOT call any lowering. So `var p : proc` produces `Atom("var", Variable("p"))` with 1 term. The compiler expects `Atom("local", varDecl, body)` with 2+ terms → fails with "local needs variables and body".

### Fix

**File: `parser/action.go`** — Add `lowerVarStatements()` function matching Python's `lower_var_stmts`:

```go
// lowerVarStatements transforms var declarations into nested local scopes.
// Python: lower_var_stmts (ivy_parser.py:2324-2350)
//
// Input:  [VarAction(p), VarAction(m), Assign(wr, false)]
// Output: [Local(p, Local(m, Assign(wr, false)))]
func lowerVarStatements(stmts []ast.Node) []ast.Node {
    for i, stmt := range stmts {
        if isVarAtom(stmt) {
            // var with init becomes: local var { init_assign; rest }
            // var without init becomes: local var { rest }
            rest := lowerVarStatements(stmts[i+1:])
            body := makeSequence(maybeInitAssign(stmt), rest)
            local := ast.NewAtom("local", extractVarDecl(stmt), body)
            return append(stmts[:i], local)
        }
    }
    return stmts
}
```

Call it in `parseSequence()` after collecting all statements:
```go
func (p *Parser) parseSequence() ast.Node {
    // ... parse stmts ...
    stmts = lowerVarStatements(stmts)  // NEW: lower var to local
    // ... build And node ...
}
```

### Regression Test

```go
func TestRegression_VarInActionBody(t *testing.T) {
    src := `
type bool
type proc

object cfabric = {
    individual rd_pio_fair : bool

    action step(sel:bool) = {
        var p : proc;
        rd_pio_fair := sel
    }
}
export cfabric.step
`
    p := parser.New(src, lexer.Version{1, 8})
    result, _ := p.Parse()
    mod := module.New()
    mod.Cfg = module.NewConfig()
    err := IvyCompile(result.Decls, mod, false)
    if err != nil {
        if strings.Contains(err.Error(), "local needs variables and body") {
            t.Errorf("var inside action body not lowered to local: %v", err)
        } else {
            t.Logf("IvyCompile error: %v", err)
        }
    }
    // cfabric.step must be registered with non-empty body
    if act, ok := mod.Actions["cfabric.step"]; !ok {
        t.Errorf("cfabric.step missing from Actions")
    } else if fmt.Sprint(act) == "true" || fmt.Sprint(act) == "" {
        t.Errorf("cfabric.step has empty body")
    }
}
```

---

## Bug B: Action Parameter Types Not Rewritten During Module Expansion (8 actions affected)

### Root Cause

When `module unbounded_sequence = { type this; alias t = this; action next(x:t) ... }` is expanded as `instance index : unbounded_sequence`:

1. The AST rewriter transforms `type this` → `type index` ✅
2. The AST rewriter transforms `alias t = this` → `alias index.t = index` ✅
3. The action `next(x:t)` is rewritten to `index.next(x:t)` — the name is prefixed but the parameter type annotation `t` is NOT rewritten to `index.t` ❌

At compile time:
- `CmplSort("t")` → `ResolveAlias("t")` → not found (alias is `"index.t"`, not `"t"`)
- Falls through to `FindSort("t")` → "unknown type: t"

The issue is that the AST rewriter's `RewriteSort` function should transform type annotations in `Variable` nodes. Looking at `AstRewrite` case `*Variable` (ast/rewrite.go:417-424):

```go
case *Variable:
    sortStr := ""
    if n.VSort != nil {
        sortStr = fmt.Sprint(n.VSort)
    }
    newSort := RewriteSort(rewrite, sortStr)
    return n.Resort(NewSymbol(newSort, nil))
```

`RewriteSort` calls `rewrite.RewriteName(sortStr)` which only does subscript substitution — it does NOT apply `PrefixStr` to the sort name. So `t` stays as `t`.

Python handles this because `rewrite_sort` in `ivy_ast.py` does:
```python
def rewrite_sort(rewrite, sort):
    if isinstance(sort, str):
        return rewrite.rewrite_name(sort)
```
And Python's `rewrite_name` applies `subst_subscripts` which includes the formal→actual substitution. But crucially, Python's sort annotation is a string that gets substituted in the name tree which applies prefix_str. The Go version's `RewriteSort` only calls `RewriteName` which does subscript substitution but NOT prefix_str.

### Fix

**File: `ast/rewrite.go`** — In the `*Variable` case of `AstRewrite`, after rewriting the sort name, also apply prefix transformation:

```go
case *Variable:
    sortStr := ""
    if n.VSort != nil {
        sortStr = fmt.Sprint(n.VSort)
    }
    newSort := RewriteSort(rewrite, sortStr)
    // Also apply prefix transformation to sort names
    // (e.g., "t" → "index.t" when inside module expansion)
    if sp, ok := rewrite.(*AstRewriteSubstPrefix); ok {
        newSort = sp.PrefixStr(newSort, false)
    }
    return n.Resort(NewSymbol(newSort, nil))
```

### Regression Test

```go
func TestRegression_AliasTypeInModuleAction(t *testing.T) {
    src := `
type nat
module mymod = {
    type this
    alias t = this
    action next(x:t) returns (y:t)
}
instance idx : mymod
`
    p := parser.New(src, lexer.Version{1, 8})
    result, _ := p.Parse()
    mod := module.New()
    mod.Cfg = module.NewConfig()
    err := IvyCompile(result.Decls, mod, false)
    if err != nil {
        if strings.Contains(err.Error(), "unknown type: t") {
            t.Errorf("alias type 't' not resolved after module expansion: %v", err)
        } else {
            t.Logf("IvyCompile error: %v", err)
        }
    }
    // idx.next must be registered
    if _, ok := mod.Actions["idx.next"]; !ok {
        keys := make([]string, 0)
        for k := range mod.Actions { keys = append(keys, k) }
        t.Errorf("idx.next missing from Actions; have: %v", keys)
    }
}
```

---

## Files to Modify

| File | Change |
|------|--------|
| `parser/action.go` | Add `lowerVarStatements()`, call from `parseSequence()` |
| `ast/rewrite.go` | Apply prefix to Variable sort in `AstRewrite` case `*Variable` |
| `compiler/regression_test.go` | Add 2 regression tests |

## Verification

1. `go test ./...` passes
2. `TestRegression_VarInActionBody` passes (var lowered to local)
3. `TestRegression_AliasTypeInModuleAction` passes (alias type resolved)
4. xtracer shows Go registering the same ~50 actions as Python
5. `./goivy_check isolate=cf_live ord_live.ivy` gets past "immutable symbol assigned"
