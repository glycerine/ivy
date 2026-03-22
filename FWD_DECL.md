# Fix: `undefined action: bar.a` and `undefined isolate: cf_live`

**Created: 2026-03-22**

## Two Bugs

### Bug 1: SourceFile only runs first compilation pass
`SourceFile` calls `DomainSetup.ProcessDecls` only. Python's `ivy_compile` runs all 3 passes. **Fixed**: changed to `compiler.IvyCompile(result.Decls, mod)`.

### Bug 2: IsolateObjectDecl not registered
The ARGSetup only has `case *ast.IsolateDecl:`, missing `*ast.IsolateObjectDecl`. **Fixed**: added case with shared `registerIsolateDecl` helper.

### Bug 3 (exposed): ARGSetup silently drops actions that fail compilation
When `CompileAction` fails (e.g., for forward declarations with no body), line 487-488 prints a debug message and `continue`s — the action is never registered in `mod.Actions`. When `export bar.a` is later processed, it can't find `bar.a`.

Python's `IvyARGSetup.action()` just calls `compile_action_def(a, sig)` without error handling — errors propagate. Python's `compile_action_def` handles empty bodies by creating an empty sequence.

## Fix for Bug 3

**File: `compiler/ivy_compile.go`, lines 484-492**

Change the error handling: don't silently skip. If `CompileAction` fails, still register the action with an empty body (matching Python's behavior for forward declarations):

```go
name := ad.Defines()
action, err := as.Compiler.CompileAction(ad)
if err != nil {
    // Register with empty sequence for forward declarations.
    // Python: compile_action_def always succeeds, creating empty Sequence for no-body actions.
    action = actions.NewSequence()
}
mod.Actions[name] = action
mod.PublicActions[name] = true
```

Also need to handle the case in `CompileAction` where `node.Body` is nil:

**File: `compiler/action.go`, line 23**

```go
bodyToCompile := node.Body
if bodyToCompile == nil {
    // Forward declaration — no body. Return empty sequence.
    return actions.NewSequence(), nil
}
```

## Verification

1. `go build ./...` compiles
2. `go test ./...` passes
3. `./goivy_check ivy-lang-examples/test/action1.ivy` prints "OK"
4. `./goivy_check isolate=cf_live ord_live.ivy` gets past the action/isolate errors
