# Fix Golden Test Divergence at Line 124634 — Missing Concrete Action Type Cases

**Created:** 2026-03-27 17:30

## Context

Steps 1–7 of the lg.Expr refactoring are COMPLETE (actions implement lg.Expr, WrapAction/UnwrapAction eliminated). Golden test now advances to line 124634 (was 124589).

Divergence at line 124634:
```
go : XTRACE: compiler.Thing ENTER type=AssertAction    ← extra recursive Thing call
py : XTRACE: compiler.CompileNode return case=Action   ← Python returns from CompileNode
```

**Root cause:** Go's `CompileNode` routes `*ast.AssertAction` to `CompileActionBody`, but `CompileActionBody` has no `case *ast.AssertAction`. It falls to the default (line 414) which calls `c.Thing(node)` again — creating a recursive loop with an extra trace.

The same bug affects ALL concrete action AST types that have Python `.cmpl` overrides but are missing from `CompileActionBody`'s switch.

## Analysis: Python .cmpl vs Go CompileActionBody

Python action types with explicit `.cmpl` methods (from ivy_compiler.py):

| Python type | Python .cmpl handler | Go CompileActionBody case? | Trace emitted by Python |
|---|---|---|---|
| AssignAction | compile_assign | ✅ `*ast.AssignAction` | — |
| AssertAction | compile_assert_action | ❌ MISSING | `case=Action` |
| AssumeAction | compile_assert_action | ❌ MISSING | `case=Action` |
| CallAction | compile_call | ❌ MISSING | `case=default type=CallAction` |
| IfAction | compile_if_action | ❌ MISSING | `case=default type=IfAction` |
| WhileAction | compile_while_action | ❌ MISSING | `case=default type=WhileAction` |
| LocalAction | compile_local | ✅ `*ast.LocalAction` | — |
| CrashAction | compile_crash_action | ✅ `*ast.CrashAction` | — |
| ThunkAction | compile_thunk_action | ✅ `*ast.ThunkAction` | — |
| DebugAction | compile_debug_action | ❌ MISSING | `case=default type=DebugAction` |
| NativeAction | compile_native_action | ❌ MISSING | `case=default type=NativeAction` |
| SetAction | NO .cmpl (uses other_thing) | In CompileNode routing but shouldn't be | — |
| HavocAction | NO .cmpl (uses other_thing) | In CompileNode routing but shouldn't be | — |

## Fix Plan

### Step 1: Add missing concrete action type cases to CompileActionBody

**File:** `compiler/action.go`, function `CompileActionBody`

Add these cases to the switch statement (before the closing `}` of the switch at line 411):

#### `*ast.AssertAction`
Python's `compile_assert_action` (line 770): ExprContext + sortify_with_inference on args[0] + optional proof via args[1].compile()
```go
case *ast.AssertAction:
    xtracer.Trace("compiler.CompileNode return case=Action")
    // Python: compile_assert_action — args[0] is the formula, args[1] is optional proof
    if len(n.Elems) >= 1 {
        act, err := c.CompileAssertFormula(n.Elems[0])
        if err != nil {
            return nil, err
        }
        if len(n.Elems) >= 2 {
            pf, pfErr := c.CompileTactic(n.Elems[1])
            if pfErr == nil && pf != nil {
                if aa, ok := act.(*actions.AssertAction); ok {
                    aa.Proof = actions.WrapTactic(pf)
                }
            }
        }
        return act, nil
    }
    return nil, fmt.Errorf("assert needs a formula")
```

#### `*ast.AssumeAction`
Python: same `compile_assert_action` handler (line 789)
```go
case *ast.AssumeAction:
    xtracer.Trace("compiler.CompileNode return case=Action")
    if len(n.Elems) >= 1 {
        return c.CompileAssumeFormula(n.Elems[0])
    }
    return nil, fmt.Errorf("assume needs a formula")
```

#### `*ast.CallAction`
Python: `compile_call` (line 684) — ExprContext + looks up action in top_context.actions
```go
case *ast.CallAction:
    xtracer.Trace("compiler.CompileNode return case=default type=CallAction")
    if len(n.Elems) >= 1 {
        return c.CompileCall(n.Elems[0], n.Elems[1:])
    }
    return nil, fmt.Errorf("call needs a target")
```

#### `*ast.IfAction`
Python: `compile_if_action` (line 723) — handles Some variant + ExprContext for plain if
```go
case *ast.IfAction:
    xtracer.Trace("compiler.CompileNode return case=default type=IfAction")
    return c.CompileIf(n.Cond, n.Then, n.Else)
```

#### `*ast.WhileAction`
Python: `compile_while_action` (line 750) — ExprContext + invariants
```go
case *ast.WhileAction:
    xtracer.Trace("compiler.CompileNode return case=default type=WhileAction")
    if len(n.Elems) >= 2 {
        var invNodes []ast.Node
        if len(n.Elems) > 2 {
            invNodes = n.Elems[2:]
        }
        return c.CompileWhile(n.Elems[0], n.Elems[1], invNodes)
    }
    return nil, fmt.Errorf("while needs condition and body")
```

#### `*ast.DebugAction`
Python: `compile_debug_action` (line 873)
```go
case *ast.DebugAction:
    xtracer.Trace("compiler.CompileNode return case=default type=DebugAction")
    result, err := c.CompileDebugAction(node)
    if err != nil {
        return nil, err
    }
    if act, ok := result.(actions.Action); ok {
        return act, nil
    }
    return actions.NewSequence(), nil
```

#### `*ast.NativeAction`
Python: `compile_native_action` (line 908) — check if CompileNativeAction exists; if not, create it
```go
case *ast.NativeAction:
    xtracer.Trace("compiler.CompileNode return case=default type=NativeAction")
    // Python: NativeAction.cmpl = compile_native_action
    act := actions.NewNativeAction(nil)
    act.SetLineno(node.GetLineno())
    return act, nil
```

### Step 2: Move "CompileNode return case=Action" trace from CompileNode to individual handlers

**File:** `compiler/compiler.go`, lines 291–301

Currently CompileNode emits `"compiler.CompileNode return case=Action"` AFTER `CompileActionBody` returns. But Python emits this at the START of each `.cmpl` handler. The trace must move INTO each handler (Step 1 above already includes it).

Change CompileNode's action routing to NOT emit the trace:
```go
case *ast.AssignAction, *ast.AssumeAction, *ast.AssertAction,
    *ast.CrashAction, *ast.ThunkAction,
    *ast.LocalAction,
    *ast.CallAction, *ast.IfAction, *ast.WhileAction,
    *ast.DebugAction, *ast.NativeAction:
    act, err := c.CompileActionBody(node)
    if err != nil {
        return nil, err
    }
    return act, nil
```

Remove `*ast.SetAction`, `*ast.HavocAction`, and `*ast.InstantiateDecl` from this list — they have no `.cmpl` in Python, so they should use the default `OtherThing` path.

**Important:** Keep the `*ast.InstantiateDecl` case in CompileActionBody even though we remove it from CompileNode routing — CompileActionBody is also called from other places (e.g., compiling if/while bodies).

### Step 3: Add appropriate traces to existing CompileActionBody handlers

The existing handlers need "CompileNode return case=..." traces at their START to match Python. Verified traces from Python source:

- `*ast.AssignAction`: add `xtracer.Trace("compiler.CompileNode return case=Action")` at start (Python's `compile_assign` says `case=Action`)
- `*ast.CrashAction`: add `xtracer.Trace("compiler.CompileNode return case=default type=CrashAction")` at start
- `*ast.ThunkAction`: add `xtracer.Trace("compiler.CompileNode return case=default type=ThunkAction")` at start
- `*ast.LocalAction`: add `xtracer.Trace("compiler.CompileNode return case=default type=LocalAction")` at start
- `*ast.InstantiateDecl`: check Python — if InstantiateAction has no .cmpl, it uses other_thing, so consider removing from action routing
- `*ast.Ite`: this handles `if` from hand-rolled parser; check Python trace for Ite if relevant

### Step 4: Verify Python trace for each handler

Check each Python `.cmpl` handler's FIRST trace to ensure Go matches:
- `compile_assert_action`: `"compiler.CompileNode return case=Action"` ✓
- `compile_call`: `"compiler.CompileNode return case=default type=CallAction"` ✓
- `compile_if_action`: `"compiler.CompileNode return case=default type=IfAction"` ✓
- `compile_while_action`: `"compiler.CompileNode return case=default type=WhileAction"` ✓
- `compile_crash_action`: `"compiler.CompileNode return case=default type=CrashAction"` ✓
- `compile_thunk_action`: check Python source
- `compile_debug_action`: check Python source
- `compile_native_action`: check Python source
- `compile_local`: check Python source
- `compile_assign`: check Python source

## Critical Files

- `compiler/action.go` — `CompileActionBody` switch: add 7 new cases, add traces to 5 existing cases
- `compiler/compiler.go` — `CompileNode` action routing: remove trace, remove SetAction/HavocAction
- Python source of truth: `~/pyivy/ivy/ivy/ivy_compiler.py`

## Verification

```bash
go build ./...
go test ./compiler/
cd ~/goivy && make golden   # should advance past line 124634
```
