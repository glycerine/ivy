# Fix: "mixin has wrong number of input parameters: 0 vs 2" panic

**Created: 2026-03-22**

## Context

Running `goivy_check isolate=cf_live ord_live.ivy` panics:
```
panic: mixin has wrong number of input parameters: 0 vs 2
```

Stack: `ApplyMixin` → `AddMixinsExt` → `IsolateComponent` → `CreateIsolate`

## Root Cause

When `CompileAction` fails (e.g., "unknown symbol: cfabric.complete_hook"), the caller at `ivy_compile.go:527` creates a bare `actions.NewSequence()` with **0 formal params**:

```go
// ivy_compile.go:522-528
action, err := as.Compiler.CompileAction(ad)
if err != nil {
    action = actions.NewSequence()  // 0 formal params!
}
mod.Actions[name] = action
```

Later, `IsolateComponent` applies before/after mixins to this action. `ApplyMixin` (actions/helpers.go:145) requires matching param counts — the mixin has 2 params, the fallback has 0 → panic.

**In Python**, `compile_action_def` doesn't fail for these actions (they resolve all types/symbols). But even if it did, the formal params are compiled **before** the body — they'd be available regardless of body failure.

## Fix

**File: `compiler/action.go`** — Restructure `CompileAction` to return a fallback action with formal params even when the body fails to compile.

Currently, `CompileAction` compiles formals (lines 60-84), then the body (line 102). If the body fails (line 104), it returns `nil, err`, discarding the already-compiled formals.

The fix: when the body fails, return an empty sequence **with the formals** alongside the error.

```go
// compiler/action.go — after compiling body (around line 102-106)
body, err := c.CompileActionBody(bodyToCompile)
c.Sig = savedSig
if err != nil {
    // Body failed, but formals are compiled. Return fallback with formals
    // so callers can register an action with the correct parameter signature.
    fallback := actions.NewSequence()
    fallback.SetFormalParams(formals)
    fallback.SetFormalReturns(returns)
    return fallback, err
}
```

**File: `compiler/ivy_compile.go`** — Update ARGSetup to use the returned fallback action instead of creating a bare `NewSequence()`:

```go
action, err := as.Compiler.CompileAction(ad)
if err != nil {
    xtracer.Trace("compiler.ARGSetup.action COMPILE_FAIL name=%s err=%v", name, err)
    pp("ARGSetup: compiling action %s: %v (registering fallback)", name, err)
    // action already has formals from CompileAction's fallback.
    // Only create bare sequence if CompileAction returned nil.
    if action == nil {
        action = actions.NewSequence()
    }
}
```

**Edge case: formal param compilation also fails** (e.g., "unknown type: nat"). In this case, `CompileAction` returns `nil, err` at lines 62-73 (before formals are complete). For these, the bare `NewSequence()` fallback is unavoidable. But the panic won't occur because those actions (with truly unknown types) won't have before/after mixins that reference them — the mixins also fail.

## Files to Modify

| File | Change |
|------|--------|
| `compiler/action.go` line 104 | On body compilation failure, return fallback sequence with already-compiled formals |
| `compiler/ivy_compile.go` line 527 | Use returned fallback action instead of bare `NewSequence()` |

## Verification

1. `go test ./...` passes
2. `goivy_check isolate=cf_live ord_live.ivy` no longer panics with "mixin has wrong number of input parameters"
3. Compilation proceeds past the IsolateComponent phase
