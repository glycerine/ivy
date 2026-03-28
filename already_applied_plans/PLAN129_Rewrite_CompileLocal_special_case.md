# Fix: CompileLocal divergence at line 134628
Created: 2026-03-28

## Context

At line 134628, Go's `CompileLocal` special case fails because it checks the WRONG node for the assignment pattern. Python checks `isinstance(ls[0], AssignAction)` (the first local declaration), but Go checks `body.(*ast.Atom).Rep == ":="` (the body). When `LowerVarStatements` produces `LocalAction(AssignAction(lsym, rhs), body)`, the AssignAction is in `localDecls[0]`, not in `body`. Go then passes the raw AssignAction to `CompileConst`, which has no case for it and errors out.

## Divergence trace
```
134627  go : compiler.CompileConst ENTER
        py : compiler.CompileConst ENTER
134628  go : compiler.Thing return type=LocalAction    ← CompileConst errors, returns up to Thing
        py : compiler.GetFunctionSort ENTER            ← CompileConst succeeds, processes LHS args
```

## Root cause

**Go (action.go:892-893) — checks body, wrong:**
```go
if assignAtom, ok := body.(*ast.Atom); ok && assignAtom.Rep == ":=" {
    sym, err := c.CompileConst(localDecls[0], sigCopy)  // localDecls[0] is *ast.AssignAction!
```

**Python (ivy_compiler.py:594-607) — checks localDecls[0], correct:**
```python
if len(ls) == 1 and isinstance(ls[0], AssignAction):
    v = ls[0]           # the AssignAction
    lhs = v.args[0]     # the variable declaration (Atom or App)
    rhs = v.args[1]     # the initial value
    sym = compile_const(lhs, sig)  # passes LHS only, not the AssignAction
```

**AST structure from LowerVarStatements (ast/lower_var.go:58-70):**
```
VarAction(x:T, value)  →  LocalAction(AssignAction(loc:x:T, value), body)
                             Elems[0] = AssignAction     Elems[1] = body(Sequence)
```

So `localDecls[0]` = `AssignAction(loc:x:T, value)`, `body` = continuation Sequence.
`CompileConst` type-switch (compiler.go:1060) handles Atom/App/Symbol but NOT AssignAction → falls to default → error.

## Structural comparison: Python vs Go

### Python flow (ivy_compiler.py:594-632):
```python
v = ls[0]                              # AssignAction
lhs = v.args[0]                        # variable declaration
if not hasattr(lhs, "sort"):           # ensure sort annotation exists
    lhs.sort = 'S'
rhs = v.args[1]                        # initial value
tmp_lhs = lhs
with ExprContext(code, local_syms):
    with top_sort_as_default():
        sym = compile_const(tmp_lhs, sig)  # register symbol
        ctmp_lhs = tmp_lhs.compile()       # compile LHS via thing()
        crhs = rhs.compile()              # compile RHS via thing()
with ASTContext(self):
    if is_variant(ctmp_lhs.sort, crhs.sort):
        teq = sort_infer(pto(ctmp_lhs.sort, crhs.sort)(ctmp_lhs, crhs))
    else:
        teq = sort_infer(Equals(ctmp_lhs, crhs))
clhs, crhs = list(teq.args)
asgn = v.clone([clhs, crhs])
remove_symbol(sym)
if clhs.rep.name in sig.symbols:
    del sig.symbols[clhs.rep.name]      # shadow
add_symbol(clhs.rep.name, clhs.rep.sort)
body = sortify(self.args[-1])           # compile body SEPARATELY
lines = body.args if isinstance(body, Sequence) else [body]
body = Sequence(*([asgn] + lines))      # prepend assignment to body
code.append(LocalAction(clhs.rep, body))
# ... extract pattern
```

### Current Go flow (action.go:892-998) — WRONG:
```go
// Checks body for ":=" — body is NOT the assignment, it's the continuation!
if assignAtom, ok := body.(*ast.Atom); ok && assignAtom.Rep == ":=" {
    sym := CompileConst(localDecls[0], ...)  // localDecls[0] is AssignAction, NOT the LHS!
    lhs := Thing(assignAtom.Terms[0])        // lhs/rhs from BODY, not from localDecls
    rhs := Thing(assignAtom.Terms[1])
    // ... sort inference ...
    // NO separate body compilation — body IS the assignment in this wrong model
}
```

## Fix plan

### Rewrite the special case to match Python

Replace `action.go:892-998` with code that:

1. **Checks `localDecls[0]` for `*ast.AssignAction`** (not body for Atom ":=")
2. **Extracts LHS/RHS** from `assignAction.Elems[0]` and `assignAction.Elems[1]`
3. **Ensures LHS has sort annotation** — Python: `if not hasattr(lhs, "sort"): lhs.sort = 'S'`
   - Go equivalent: if `ASort` is nil on Atom/App, set to `&ast.Symbol{Rep: "S"}`
   - `extractSortRep` on `&ast.Symbol{Rep: "S"}` returns `"S"` → `CmplSort("S")` finds universe sort
   - NOTE: this is actually redundant in Python too (both paths in CompileConst resolve to the same default sort), but we implement it faithfully
4. **Passes LHS to CompileConst** (not the AssignAction)
5. **Compiles LHS via Thing**, **compiles RHS via Thing** (within ExprContext + top_sort_as_default)
6. **Sort inference** via Equals or variant pto (same as current Go code)
7. **Extracts local variable symbol** from compiled LHS:
   - Python: `clhs.rep` which is `self` for Symbol, `self.func` for Apply (ivy_logic.py:129,283)
   - Go: if `*lg.Symbol` → use directly; if `*lg.Apply` → use `.Func`
8. **Remove/shadow/add symbol** (same as current Go code, using inferred sort)
9. **Compiles body SEPARATELY** via `c.Sortify(body)` — this is the continuation body, NOT the assignment
10. **Extracts body lines**: if compiled body is `*actions.Sequence` → use `.Elems`; else wrap in single-element list
11. **Prepends assignment** to body: `actions.NewSequence(append([]lg.Expr{asgn}, bodyLines...)...)`
12. **Wraps in LocalAction**: `NewLocalAction(localVar, bodyWithAsgn)` where `localVar` is the symbol from step 7

### Pseudocode for the rewrite

```go
if len(localDecls) == 1 {
    if assignAction, ok := localDecls[0].(*ast.AssignAction); ok && len(assignAction.Elems) >= 2 {
        lhsNode := assignAction.Elems[0]  // variable declaration
        rhsNode := assignAction.Elems[1]  // initial value

        // Step 3: ensure sort annotation (Python: if not hasattr(lhs,"sort"): lhs.sort='S')
        ensureSortAnnotation(lhsNode)

        // Set up ExprContext + top_sort_as_default (same as current)
        code := make([]lg.Expr, 0)
        localSyms := make([]*lg.Symbol, 0)
        savedExprCtx := c.ExprCtx
        loc := assignAction.GetLineno()
        c.ExprCtx = &ExprContext{Code: code, LocalSyms: localSyms, Lineno: &loc, ActCfg: c.ActCfg}
        savedSig := c.Sig
        c.Sig = sigCopy
        tsDefault := il.TopSortAsDefault(sigCopy)
        tsDefault.Enter()

        // Step 4: compile_const(lhs, sig) — register symbol
        sym, err := c.CompileConst(lhsNode, sigCopy)
        if err != nil { /* restore + return */ }

        // Step 5: compile LHS and RHS via Thing
        lhs, lhsErr := c.Thing(lhsNode)
        rhs, rhsErr := c.Thing(rhsNode)

        tsDefault.Exit()
        exprCtx := c.ExprCtx
        c.ExprCtx = savedExprCtx
        if lhsErr != nil { /* error */ }
        if rhsErr != nil { /* error */ }

        // Step 6: sort inference (Equals or variant pto — same as current)
        // ... existing sort inference code ...

        // Step 7: extract local variable symbol
        var localVar lg.Expr
        if app, ok := lhs.(*lg.Apply); ok {
            localVar = app.Func  // Python: Apply.rep = self.func
        } else {
            localVar = lhs       // Python: Symbol.rep = self
        }

        // Step 8: remove/shadow/add symbol
        localName := sym.Name
        localSort := lhs.NodeSort()
        sigCopy.RemoveSymbol(localName, sym.CSort)
        delete(sigCopy.Symbols, localName)
        sigCopy.AddSymbol(localName, localSort)

        c.Sig = savedSig

        // Step 9: compile body SEPARATELY
        c.Sig = sigCopy
        compiledBody, bodyErr := c.Sortify(body)
        c.Sig = savedSig
        if bodyErr != nil { /* error */ }

        // Step 10: extract body lines
        var bodyLines []lg.Expr
        if seq, ok := compiledBody.(*actions.Sequence); ok {
            bodyLines = seq.Elems
        } else {
            bodyLines = []lg.Expr{compiledBody}
        }

        // Step 11: prepend assignment, build combined body
        asgn := actions.NewAssignAction(lhs, rhs)
        asgn.SetLineno(assignAction.GetLineno())
        allLines := append([]lg.Expr{asgn}, bodyLines...)
        bodyWithAsgn := actions.NewSequence(allLines...)

        // Step 12: wrap in LocalAction
        exprCtx.Code = append(exprCtx.Code, c.ActCfg.NewLocalAction(localVar, bodyWithAsgn))

        // Set lineno on all code items
        for _, codeItem := range exprCtx.Code {
            if act, ok := codeItem.(actions.Action); ok {
                act.SetLineno(assignAction.GetLineno())
            }
        }

        // Extract pattern (same as current)
        if len(exprCtx.Code) == 1 {
            if act, ok := exprCtx.Code[0].(actions.Action); ok {
                return act, nil
            }
        }
        args := make([]lg.Expr, 0, len(exprCtx.LocalSyms)+1)
        for _, s := range exprCtx.LocalSyms {
            args = append(args, s)
        }
        args = append(args, actions.NewSequence(exprCtx.Code...))
        result := c.ActCfg.NewLocalAction(args...)
        result.SetLineno(assignAction.GetLineno())
        return result, nil
    }
}
```

### ensureSortAnnotation helper

```go
func ensureSortAnnotation(n ast.Node) {
    // Python: if not hasattr(lhs, "sort"): lhs.sort = 'S'
    // In Go, ASort is always a struct field (possibly nil).
    // 'S' is the universe sort string — extractSortRep returns "S",
    // CmplSort("S") resolves it to the universe sort.
    switch v := n.(type) {
    case *ast.Atom:
        if v.ASort == nil {
            v.ASort = &ast.Symbol{Rep: "S"}
        }
    case *ast.App:
        if v.ASort == nil {
            v.ASort = &ast.Symbol{Rep: "S"}
        }
    }
}
```

## Files to modify
- `compiler/action.go` — rewrite CompileLocal special case (lines 892-998), add `ensureSortAnnotation` helper

## Key differences from current Go code
| Aspect | Current Go (wrong) | Python / Fixed Go |
|--------|-------------------|-------------------|
| What is checked | `body.(*ast.Atom).Rep == ":="` | `localDecls[0].(*ast.AssignAction)` |
| LHS source | `assignAtom.Terms[0]` (from body) | `assignAction.Elems[0]` (from decl) |
| RHS source | `assignAtom.Terms[1]` (from body) | `assignAction.Elems[1]` (from decl) |
| CompileConst arg | `localDecls[0]` (full AssignAction) | `lhsNode` (just the LHS variable) |
| Body compilation | body IS the assignment | body compiled SEPARATELY via Sortify |
| LocalAction body | just the assignment | Sequence([assignment] + bodyLines) |

## Verification
```
cd ~/goivy && make golden
```
