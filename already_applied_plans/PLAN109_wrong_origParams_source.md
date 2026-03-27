# Fix xtrace divergence at line 124588: wrong origParams source + missing sortify trace

**Created:** 2026-03-27

## Context

Golden test diverges at xtrace line 124588:
- **Go:** `compiler.CompileConst ENTER` (3rd CompileConst)
- **Python:** `compiler.sortify ENTER`

Two issues:
1. Go emits 3 CompileConst calls where Python emits 2 (extra CompileConst)
2. Go is missing the `compiler.sortify ENTER` trace before body compilation

## Root Cause: Wrong origParams source

**Python** (`compile_action_def` in ivy_compiler.py:926-939):
```python
params = a.args[0].args          # object params from the NAME atom's children
pformals = [v.to_const('prm:') for v in params]
formals = [compile_const(v,sig) for v in pformals + a.formal_params]
returns = [compile_const(v,sig) for v in a.formal_returns]
res = sortify(a.args[1])
```
For `index.next`: Name atom has no children → `params=[]` → `pformals=[]` → formals = 0 + 1 fml:param = 1 CompileConst, returns = 1 CompileConst → total 2.

**Go** (`CompileAction` in compiler/action.go:30-92):
```go
origParams, _ := node.Formals()  // BUG: strips fml: from FormalParams
pformals = prm:-prefix origParams
// Compiles: pformals(1) + FormalParams(1) + FormalReturns(1) = 3 CompileConst
```
Go derives `origParams` from `node.Formals()` which returns unprefixed `FormalParams` — the SAME conceptual params. So each param gets compiled twice (once as prm:, once as fml:).

## Fix

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/action.go`

### Fix 1: Get origParams from node.Name's args (matching Python's `a.args[0].args`)

Change line 30-31 from:
```go
// Get original (unprefixed) params for prm: renaming (Python line 803)
origParams, _ := node.Formals()
```
To:
```go
// Get object params from the action name atom's children (Python line 926: params = a.args[0].args)
var origParams []Node
if node.Name != nil {
    origParams = node.Name.Args()
}
```

This matches Python where `params = a.args[0].args` gets the Name atom's term-arguments (object/receiver params), NOT the formal params.

### Fix 2: Add sortify ENTER trace before CompileActionBody

At line ~131 (before `c.CompileActionBody(bodyToCompile)`), add:
```go
xtracer.Trace("compiler.sortify ENTER")
body, err := c.CompileActionBody(bodyToCompile)
```

## Verification

Run `cd ~/goivy && make golden` and confirm the test advances past line 124588.
