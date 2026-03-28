# Fix: CompileCall divergence at line 130675
Created: 2026-03-28

## Context

At line 130675, Go's `compile_call` (CompileCall) fails immediately with "call to non-action" because it only handles `*ast.Atom` callee nodes, but the LALR parser produces `*ast.App` for bare action calls like `retire_hook(p,t);` (line 510 of ord_live.ivy). Additionally, even after fixing the type handling, Go uses `Thing()` to compile callee args while Python uses `cmpl()` — producing extra traces that would cause the next divergence.

## Divergence trace
```
130674  go : compiler.compile_call ENTER
        py : compiler.compile_call ENTER
130675  go : compiler.Thing return type=CallAction        ← Go returns immediately (error)
        py : compiler.CompileNode return case=App         ← Python compiles callee arg via cmpl()
```

## Root cause analysis

### Bug A: CompileCall only handles `*ast.Atom` callee, not `*ast.App`

Grammar rule `simpleact → term %prec TOK_SEMI` (grammar_v17.y:4136-4142) wraps the term in a CallAction. But `term → appelem → SYMBOLx(terms)` creates `*ast.App` (grammar_v17.y:1746-1753), not `*ast.Atom`. So `CallAction.Elems[0]` is an `*ast.App`.

Go's CompileCall (action.go:720-726):
```go
if atom, ok := calleeNode.(*ast.Atom); ok {
    name = atom.Rep        // string
    calleeArgs = atom.Terms
} else {
    return nil, lg.NewIvyError(calleeNode, "call to non-action")  ← FAILS HERE
}
```

Python's compile_call (ivy_compiler.py:700-702) uses duck-typing:
```python
name = self.args[0].rep   # works for both App (rep is string) and Atom
```

Go `ast.App.Rep` is a `Node` (usually `*ast.Symbol`), not a string. Need: `app.Rep.(*ast.Symbol).Rep`.

### Bug B: Go uses `Thing()` where Python uses `cmpl()` for callee args

Python's dispatch:
- `a.compile()` = `thing()` → emits `Thing ENTER`, `CompileNode ENTER`, case trace, `Thing return`
- `a.cmpl()` = type-specific handler directly → emits ONLY the case trace (e.g., `CompileNode return case=Variable`)

In `compile_call`'s "in actions" path (ivy_compiler.py:714-715):
```python
with ctx:
    args = [a.cmpl() for a in self.args[0].args]   # cmpl, NOT compile
```

In `compile_call`'s "not in actions" path (ivy_compiler.py:704-706):
```python
with ReturnContext([a.cmpl() for a in self.args[1:]]):         # returns: cmpl
    res = compile_field_reference(name, [a.compile() for a in self.args[0].args], ...)  # callee: compile (thing)
```

And return targets in "in actions" path (ivy_compiler.py:724):
```python
res = CallAction(*([ivy_ast.Atom(name,mas)] + [a.cmpl() for a in self.args[1:]]))  # returns: cmpl
```

Summary:
| Path | Callee args | Return targets |
|------|------------|----------------|
| in actions | `cmpl()` | `cmpl()` |
| not in actions | `compile()` (thing) | `cmpl()` |

Go currently uses `Thing()` for everything.

## Fix plan

### Step 1: Add `Cmpl()` method — CompileNode without ENTER trace

Python's `cmpl()` dispatches to the type-specific handler which emits its own `CompileNode return case=X` trace. Go's `CompileNode()` adds a `CompileNode ENTER` trace before dispatching. `Cmpl()` should match Python by skipping that ENTER trace.

In `compiler/compiler.go`:
1. Rename current `CompileNode` body to `compileNodeCore(node, emitEnter bool)`
2. `CompileNode(node)` calls `compileNodeCore(node, true)`
3. `Cmpl(node)` calls `compileNodeCore(node, false)`

```go
func (c *Compiler) Cmpl(node ast.Node) (lg.Expr, error) {
    return c.compileNodeCore(node, false)
}

func (c *Compiler) CompileNode(node ast.Node) (lg.Expr, error) {
    return c.compileNodeCore(node, true)
}

func (c *Compiler) compileNodeCore(node ast.Node, emitEnter bool) (lg.Expr, error) {
    if node == nil { ... }
    if emitEnter {
        xtracer.Trace(fmt.Sprintf("compiler.CompileNode ENTER type=%s", typeName(node)))
    }
    switch n := node.(type) { ... same cases ... }
}
```

### Step 2: Handle `*ast.App` callee in CompileCall

In `compiler/action.go` lines 720-726, change to:
```go
switch v := calleeNode.(type) {
case *ast.Atom:
    name = v.Rep
    calleeArgs = v.Terms
case *ast.App:
    if sym, ok := v.Rep.(*ast.Symbol); ok {
        name = sym.Rep
    } else {
        c.ExprCtx = savedCtx
        return nil, lg.NewIvyError(calleeNode, "call to non-action")
    }
    calleeArgs = v.Terms
default:
    c.ExprCtx = savedCtx
    return nil, lg.NewIvyError(calleeNode, "call to non-action")
}
```

### Step 3: Use `Cmpl()` in CompileCall where Python uses `cmpl()`

In `compiler/action.go` CompileCall:

- "In actions" path — callee args (line 784): `c.Thing(a)` → `c.Cmpl(a)`
- "In actions" path — return targets (line 832): `c.Thing(r)` → `c.Cmpl(r)`
- "Not in actions" path — return targets (line 736): `c.Thing(r)` → `c.Cmpl(r)`
- "Not in actions" path — callee args (line 748): `c.Thing(a)` → keep as `c.Thing(a)` (Python uses `a.compile()`)

## Files to modify
- `compiler/compiler.go` — refactor CompileNode into compileNodeCore, add Cmpl()
- `compiler/action.go` — fix type assertion in CompileCall, replace Thing→Cmpl for callee args and return targets

## Verification
```
cd ~/goivy && make golden
```
