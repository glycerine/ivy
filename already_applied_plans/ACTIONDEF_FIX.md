# Fix NewActionDef to match Python's `fml:` prefix behavior

## Context

Python's `ActionDef.__init__` (ivy_ast.py:1375-1383) renames formal params/returns
with an `fml:` prefix and substitutes those names into the action body at
construction time. This prevents name capture between formals and state symbols.

Go's `NewActionDef` does none of this — it stores params/returns and body as-is.
This created a bug when we added `prm:` prefix support in `CompileAction`: the body
still referenced original param names (e.g., `x`, `y`) but only `prm:x`, `prm:y`
were added to the sig copy. We patched it with a workaround (adding original names
to sigCopy), but the real fix is making `NewActionDef` match Python.

---

## Python reference (ivy_ast.py:1375-1408)

```python
class ActionDef(Definition):
    def __init__(self, atom, action, formals=[], returns=[]):
        self.formal_params = [s.prefix('fml:') for s in formals]
        self.formal_returns = [s.prefix('fml:') for s in returns]
        if formals or returns:
            subst = dict((x.rep,y.rep) for x,y in zip(
                formals+returns, self.formal_params+self.formal_returns))
            action = subst_prefix_atoms_ast(action, subst, None, None)
        self.args = [atom, action]

    def clone(self, args):
        res = ActionDef(args[0], args[1])       # empty formals → no re-prefixing
        res.formal_params = self.formal_params   # copy already-prefixed
        res.formal_returns = self.formal_returns
        return res

    def formals(self):
        return ([s.drop_prefix('fml:') for s in self.formal_params],
                [s.drop_prefix('fml:') for s in self.formal_returns])

    def rewrite(self, rewrite):
        res = self.clone(ast_rewrite(self.args, rewrite))
        res.formal_params = [rewrite_param(p, rewrite) for p in self.formal_params]
        res.formal_returns = [rewrite_param(p, rewrite) for p in self.formal_returns]
        return res
```

Key behaviors:
1. `FormalParams`/`FormalReturns` always store `fml:`-prefixed nodes
2. Body is rewritten via `subst_prefix_atoms_ast` so body atoms use `fml:` names
3. `clone()` does NOT re-apply prefix — preserves already-prefixed values
4. `formals()` strips `fml:` to recover original names
5. `rewrite()` applies AST rewriters to formals too

---

## Changes

### 1. Add `Atom.DropPrefix` method — `ast/ast.go`

Python's `formals()` calls `s.drop_prefix('fml:')`. `App.DropPrefix` exists (line 183)
but `Atom.DropPrefix` does not.

```go
func (a *Atom) DropPrefix(s string) *Atom {
    if !strings.HasPrefix(a.Rep, s) {
        return a
    }
    c := a.Clone(a.Terms).(*Atom)
    c.Rep = a.Rep[len(s):]
    return c
}
```

### 2. Modify `NewActionDef` — `ast/decl.go:164-166`

Apply `fml:` prefix to formals/returns and substitute into body, matching Python.

```go
func NewActionDef(name, body Node, params, returns []Node) *ActionDef {
    // Rename formals with "fml:" prefix to avoid name capture (Python line 1377-1382)
    fmlParams := prefixNodes(params, "fml:")
    fmlReturns := prefixNodes(returns, "fml:")
    if len(params) > 0 || len(returns) > 0 {
        subst := make(map[string]string)
        for i, p := range params {
            subst[nodeRep(p)] = nodeRep(fmlParams[i])
        }
        for i, r := range returns {
            subst[nodeRep(r)] = nodeRep(fmlReturns[i])
        }
        body = SubstPrefixAtomsAst(body, subst, nil, nil, nil)
    }
    return &ActionDef{Name: name, Body: body, FormalParams: fmlParams, FormalReturns: fmlReturns}
}
```

Helper functions (in `ast/decl.go`):

```go
// prefixNodes applies Prefix(s) to each node, returning fml:-prefixed copies.
func prefixNodes(nodes []Node, s string) []Node {
    if len(nodes) == 0 {
        return nil
    }
    result := make([]Node, len(nodes))
    for i, n := range nodes {
        switch a := n.(type) {
        case *Atom:
            result[i] = a.Prefix(s)
        case *App:
            result[i] = a.Prefix(s)
        default:
            result[i] = n  // leave as-is if not Atom/App
        }
    }
    return result
}

// nodeRep extracts the name string from an AST node.
func nodeRep(n Node) string {
    switch a := n.(type) {
    case *Atom:
        return a.Rep
    case *App:
        return a.Relname()
    case *Variable:
        return a.Rep
    case *Symbol:
        return a.Rep
    default:
        return fmt.Sprint(n)
    }
}
```

### 3. Add `Formals()` method — `ast/decl.go`

Matches Python's `formals()` (line 1406-1408): strips `fml:` prefix.

```go
func (a *ActionDef) Formals() (params []Node, returns []Node) {
    params = make([]Node, len(a.FormalParams))
    for i, p := range a.FormalParams {
        if atom, ok := p.(*Atom); ok {
            params[i] = atom.DropPrefix("fml:")
        } else if app, ok := p.(*App); ok {
            params[i] = app.DropPrefix("fml:")
        } else {
            params[i] = p
        }
    }
    returns = make([]Node, len(a.FormalReturns))
    for i, r := range a.FormalReturns {
        if atom, ok := r.(*Atom); ok {
            returns[i] = atom.DropPrefix("fml:")
        } else if app, ok := r.(*App); ok {
            returns[i] = app.DropPrefix("fml:")
        } else {
            returns[i] = r
        }
    }
    return
}
```

### 4. Add `Rewrite()` method — `ast/decl.go`

Matches Python's `rewrite()` (line 1397-1405).

```go
func (a *ActionDef) Rewrite(rw AstRewriter) *ActionDef {
    res := a.Clone(AstRewriteSlice(a.Args(), rw)).(*ActionDef)
    res.FormalParams = rewriteParams(a.FormalParams, rw)
    res.FormalReturns = rewriteParams(a.FormalReturns, rw)
    return res
}
```

With helper (matches Python's `rewrite_param`):
```go
func rewriteParams(params []Node, rw AstRewriter) []Node {
    result := make([]Node, len(params))
    for i, p := range params {
        result[i] = AstRewrite(p, rw)
    }
    return result
}
```

### 5. Update `CompileAction` — `compiler/action.go`

Now that `NewActionDef` handles `fml:` prefixing, `CompileAction` needs to:
- Get original (unprefixed) params via `node.Formals()` for `prm:` prefix creation
- Compile `node.FormalParams` (already `fml:`-prefixed) into sigCopy on line 812
- Remove the workaround that adds original params to sigCopy

Python line 803-812:
```python
params = a.args[0].args          # original params from signature
pformals = [v.to_const('prm:') for v in params]
subst = dict(...)
a = substitute_ast(a, subst)     # substitute prm: in signature Variables
formals = [compile_const(v,sig) for v in pformals + a.formal_params]
#                                        ^^^^^^^^   ^^^^^^^^^^^^^^^^
#                                        prm:-prefixed  fml:-prefixed
```

Go changes:
```go
func (c *Compiler) CompileAction(node *ast.ActionDef) (actions.Action, error) {
    sigCopy := c.Sig.Copy()

    // Get original (unprefixed) params for prm: renaming (Python line 803)
    origParams, _ := node.Formals()

    // Rename signature params with "prm:" prefix (Python lines 804-807)
    // ... create pformals from origParams (same as before) ...
    // ... substitute prm: into body (same as before, but body already has fml: names) ...

    // Python line 812: formals = [compile_const(v,sig) for v in pformals + a.formal_params]
    var formals []*lg.Symbol
    for _, p := range paramsToCompile {  // prm:-prefixed
        sym, err := c.CompileConst(p, sigCopy)
        formals = append(formals, sym)
    }
    for _, p := range node.FormalParams {  // fml:-prefixed (a.formal_params)
        sym, err := c.CompileConst(p, sigCopy)
        formals = append(formals, sym)
    }

    // REMOVE the old workaround that added original params to sigCopy
    // (no longer needed — fml:-prefixed names are now in sigCopy)

    // Compile return parameters
    var returns []*lg.Symbol
    for _, r := range node.FormalReturns {
        sym, err := c.CompileConst(r, sigCopy)
        returns = append(returns, sym)
    }

    // ... rest unchanged ...
}
```

The `prm:` substitution in the body now substitutes Variables named `x` → `prm:x`.
But the body's atom references are already `fml:x` (from NewActionDef). So the prm:
substitution only affects the signature's Variable nodes, not the body — matching
Python's behavior exactly.

### 6. Update tests — `compiler/expr_test.go`, `ast/ast_test.go`

Tests that create `ActionDef` directly need to account for `fml:` prefixing:

- `TestExpr6_CompileActionDef_PrmPrefixSubstitution`: The body atoms will now be
  `fml:x` instead of `x`. The test only checks that formals have `prm:` prefix,
  which should still pass since `Formals()` returns unprefixed params for prm:
  creation.

- `TestExpr6_CompileActionDef_FreeVarCheckInCalls`: Body uses `Variable("X", ...)`
  which SubstPrefixAtomsAst would try to rename, but Variable X is not in the
  formals subst map (only `a` is). Need to verify this still works.

- `ast/ast_test.go:TestActionDef`: May need updating if it checks FormalParams
  directly — they'll now have `fml:` prefix.

---

## Files Modified

| File | Changes |
|------|---------|
| `ast/ast.go` | Add `Atom.DropPrefix` method |
| `ast/decl.go` | Modify `NewActionDef` (fml: prefix + body subst), add `Formals()`, add `Rewrite()`, add helpers `prefixNodes`, `nodeRep`, `rewriteParams` |
| `compiler/action.go` | Update `CompileAction` to use `node.Formals()` for prm: params, compile fml:-prefixed `node.FormalParams` into sigCopy, remove workaround |
| `compiler/expr_test.go` | Update tests if needed for fml:-prefixed body atoms |
| `ast/ast_test.go` | Update TestActionDef if it checks FormalParams directly |

---

## Verification

```bash
cd /Users/jaten/go/src/github.com/glycerine/goivy

# 1. All compiler tests pass (including Batch B)
go test ./compiler/ -v -count=1

# 2. AST tests pass
go test ./ast/ -v -count=1

# 3. Parser tests pass (NewActionDef callers)
go test ./parser/ -v -count=1

# 4. WebUI conformance test passes (the regression that prompted this)
cd /Users/jaten/goivy
DYLD_LIBRARY_PATH=/Users/jaten/goivy/ivy/z3 go test -v -run TestConformCheck -tags web ./webui/ -count=1

# 5. Full test suite (no regressions)
cd /Users/jaten/go/src/github.com/glycerine/goivy
go test ./... -count=1
```
