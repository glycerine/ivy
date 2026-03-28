# Fix CompileAction to Route Through Sortify Instead of Direct CompileActionBody

**Created:** 2026-03-28 ~current time

## Context

The golden xtrace test (`make golden`) diverges at line 124757:
- **Go:** `XTRACE: compiler.CompileNode return case=Action`
- **Python:** `XTRACE: compiler.Thing ENTER type=AssertAction`

Python's `compile_action_def` calls `sortify(a.args[1])` which calls `.compile()` -> `thing()`, producing `Thing ENTER/return` traces around the compilation. Go's `CompileAction` calls `CompileActionBody` directly, skipping the `Thing` wrapper entirely.

## Fix: Single File Change

**File:** `compiler/action.go`, lines 132-148 (in `CompileAction`)

### Current code:
```go
xtracer.Trace("compiler.sortify ENTER")
savedSig := c.Sig
c.Sig = sigCopy
body, err := c.CompileActionBody(bodyToCompile)
c.Sig = savedSig
if err != nil {
    fallback := actions.NewSequence()
    fallback.SetFormalParams(formals)
    fallback.SetFormalReturns(returns)
    return fallback, err
}
```

### Replace with:
```go
savedSig := c.Sig
c.Sig = sigCopy
result, err := c.Sortify(bodyToCompile)
c.Sig = savedSig
if err != nil {
    fallback := actions.NewSequence()
    fallback.SetFormalParams(formals)
    fallback.SetFormalReturns(returns)
    return fallback, err
}

// Convert lg.Expr to actions.Action (same pattern as CompileActionBody's Sequence case)
var body actions.Action
switch v := result.(type) {
case actions.Action:
    body = v
case *lg.And:
    seq := actions.NewSequence(v.Terms...)
    seq.SetLineno(bodyToCompile.GetLineno())
    body = seq
default:
    body = actions.NewSequence()
}
```

### Why each change:
1. **Remove `xtracer.Trace("compiler.sortify ENTER")`** -- `Sortify` (phase6.go:507) already emits it
2. **Replace `CompileActionBody` with `Sortify`** -- Routes through `Sortify` -> `Thing` -> `CompileNode` -> `CompileActionBody`, producing correct traces
3. **Add type-switch conversion** -- `Sortify` returns `lg.Expr`; rest of function needs `actions.Action`. Three cases: direct action, `*lg.And` from Sequence, fallback empty

### Trace flow after fix (for AssertAction body):
```
sortify ENTER         <- Sortify
Thing ENTER type=...  <- Thing
CompileNode ENTER     <- CompileNode
CompileNode return    <- CompileActionBody (via CompileNode Action case)
compile_assert_action <- CompileAssertFormula
Thing return type=... <- Thing
```

## Verification

```bash
cd ~/goivy && make golden
```

Check that line 124757 now matches. Test should progress past this point (likely to a later divergence from other direct `CompileActionBody` calls in `CompileIf`, `CompileWhile`, `CompileLocal`, etc. -- those are separate fixes).
