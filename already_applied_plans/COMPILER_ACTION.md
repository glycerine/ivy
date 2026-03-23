# Fix nil pointer dereference caused by failed action compilations

**Created: 2026-03-22**

## Context

`make trt` crashes with nil pointer dereference in the fragment checker. The crash chain:

1. **Root cause**: Bare action calls (e.g., `complete_hook(t);`) in action bodies are parsed as plain atoms, not call actions. When `CompileActionBody` processes these, it falls through to `CompileNode` without `ExprCtx`, so the `TopCtx.Actions` check in `CompileAtom` (`compiler.go:438`, gated by `ExprCtx != nil`) is skipped. This sends the call to `CompileFieldReference` which fails with "unknown symbol".

2. **Cascade**: Failed action compilations → fallback actions with 0 formals → mixin param count mismatches (the 44 "warning: mixin has wrong number of input parameters" messages) → incomplete formulas with nil sorts → nil pointer crash in fragment checker.

**Python behavior**: Python's parser wraps ALL bare function calls in action bodies as `CallAction(callatom)`. Python's `compile_call` then checks `top_context.actions[name]` — always finding the action. Go's parser returns bare atoms for implicit calls (see `parser/action.go:158`), so they never reach `CompileCall`.

**Evidence**: Go trace shows `cfabric.complete_hook` is successfully collected (line 252-253) and IS in `TopCtx.Actions`. But `cfabric.step`'s body fails (line 255) because the bare call bypasses the `TopCtx.Actions` lookup.

## Change

### Fix bare action calls in `CompileActionBody` (`compiler/action.go:323`)

In `CompileActionBody`'s `case *ast.Atom:` → `default:` case, before falling through to `CompileNode`, check if the atom name is a known action in `TopCtx.Actions`. If so, compile it as a call.

At line 323 (`default:`), add:

```go
default:
    // Bare action call without "call" keyword.
    // Python parser wraps these in CallAction; Go parser leaves them as atoms.
    // Check TopCtx.Actions before falling through to expression compilation.
    if c.TopCtx != nil {
        if _, ok := c.TopCtx.Actions[n.Rep]; ok {
            return c.CompileCall(n, nil)
        }
    }
```

This matches Python where `compile_call` handles all action calls via `top_context.actions[name]`.

### Previously applied fixes (keep)

- `SortEqual` nil guard in `logic/sort.go` — defense in depth
- `defToConstraint` delegation in `clauseops/clauses.go` — correct fix
- `DefinitionToConstraint` IsIndividual check in `ivylogic/constraint.go` — faithful to Python
- `NewSymbol` nil guard in `logic/term.go` — defense in depth

## Files to modify

- `/Users/jaten/goivy/compiler/action.go` — line 323: add `TopCtx.Actions` check for bare atoms

## Verification

1. Run `make trt`
2. The "warning: mixin has wrong number of input parameters" warnings should be greatly reduced
3. The nil pointer crash in the fragment checker should be gone
