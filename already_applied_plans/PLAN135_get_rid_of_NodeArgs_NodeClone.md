# Plan: Fix Args()/Clone() on 5 LF-bearing Actions, Eliminate NodeArgs/NodeClone

**Created**: 2026-03-30 02:30

## Context

Python's recursive functions use `.args`/`.clone()` uniformly. Go's logic types already have correct `Args()`/`Clone()` in `logic/ast_compat.go`. The only broken types are the 5 LF-bearing actions whose `Args()` unwraps the LabeledFormula. Fix them directly on the types (like `ast_compat.go` does), then the recursive functions just call `node.Args()`/`node.Clone()` with no intermediary dispatch.

## Step 1: Fix Args() on 5 types in `actions/action_expr.go`

Each type gets its own `Args()` that returns `[LF]` when LF is present, matching Python's `self.args = [LabeledFormula]`. No embedding inheritance — each is explicit.

### AssumeAction.Args() (line 77)

**Current**: `return actionArgsToNodes(a.ActionArgs())` → returns `[Formula]`

**New**:
```go
func (a *AssumeAction) Args() []ast.Node {
    if a.LF != nil {
        return []ast.Node{a.LF}
    }
    return []ast.Node{a.Formula}
}
```

### AssertAction.Args() (line 97)

**Current**: `return actionArgsToNodes(a.ActionArgs())` → returns `[Formula]` or `[Formula, Proof]`

**New**:
```go
func (a *AssertAction) Args() []ast.Node {
    first := ast.Node(a.Formula)
    if a.LF != nil {
        first = a.LF
    }
    if a.Proof != nil {
        return []ast.Node{first, a.Proof}
    }
    return []ast.Node{first}
}
```

### RequiresAction.Args() (line 117)

**Current**: `return actionArgsToNodes(a.ActionArgs())` → delegates to embedded AssertAction.ActionArgs()

**New**: Same logic as AssertAction — explicit, no embedding:
```go
func (a *RequiresAction) Args() []ast.Node {
    first := ast.Node(a.Formula)
    if a.LF != nil {
        first = a.LF
    }
    if a.Proof != nil {
        return []ast.Node{first, a.Proof}
    }
    return []ast.Node{first}
}
```

### EnsuresAction.Args() (line 137)
Same as RequiresAction.

### SubgoalAction.Args() (line 438)
Same as RequiresAction.

## Step 2: Fix Clone() on 5 types in `actions/action_expr.go`

Each type's `Clone()` detects if `args[0]` is a LabeledFormula. If so, it sets both `Formula` and `LF`. If not, it falls back to the current behavior. Each concrete type preserves its own fields.

### AssumeAction.Clone() (line 78-80)

**Current**: `return a.ActionClone(nodesToExprs(args)).(ast.Node)` — panics if args[0] is LF

**New**:
```go
func (a *AssumeAction) Clone(args []ast.Node) ast.Node {
    r := &AssumeAction{
        ActionBase: a.ActionBase,
        Unprovable: a.Unprovable,
    }
    if len(args) >= 1 {
        if lf, ok := args[0].(*ast.LabeledFormula); ok {
            r.Formula = lf.Formula.(lg.Expr)
            r.LF = lf
        } else {
            r.Formula = args[0].(lg.Expr)
            r.LF = a.LF // preserve existing LF if args[0] is plain formula
        }
    }
    return r
}
```

Wait — when Python's `clone([new_lf])` is called, the new LF is the cloned/substituted version. The `r.LF` should be the new LF, not the old one. In the `else` branch (args[0] is a plain formula), it means the caller passed a formula directly (not wrapped in LF). This happens in non-LF code paths. We should NOT carry forward the old LF in that case since the formula might have changed. Actually, looking at when `Clone()` is called via `node.Clone(newArgs)` from `substituteConstantsAST`: `newArgs` comes from recursing into `Args()`. If `Args()` returned `[LF]`, then after substitution into LF's children and cloning the LF, `newArgs[0]` will be the cloned LF. If `Args()` returned `[Formula]` (no LF), then `newArgs[0]` will be the substituted formula.

So the logic should be:
- If `args[0]` is `*ast.LabeledFormula` → use it as both source of Formula and as LF
- If `args[0]` is `lg.Expr` → use as Formula, set LF = nil (LF wasn't present)

```go
func (a *AssumeAction) Clone(args []ast.Node) ast.Node {
    r := &AssumeAction{
        ActionBase: a.ActionBase,
        Unprovable: a.Unprovable,
    }
    if len(args) >= 1 {
        if lf, ok := args[0].(*ast.LabeledFormula); ok {
            r.Formula = lf.Formula.(lg.Expr)
            r.LF = lf
        } else {
            r.Formula = args[0].(lg.Expr)
        }
    }
    return r
}
```

### AssertAction.Clone() (line 98-100)

**New**:
```go
func (a *AssertAction) Clone(args []ast.Node) ast.Node {
    r := &AssertAction{
        ActionBase: a.ActionBase,
        Kind:       a.Kind,
        Unprovable: a.Unprovable,
    }
    if len(args) >= 1 {
        if lf, ok := args[0].(*ast.LabeledFormula); ok {
            r.Formula = lf.Formula.(lg.Expr)
            r.LF = lf
        } else {
            r.Formula = args[0].(lg.Expr)
        }
    }
    if len(args) >= 2 {
        r.Proof = args[1].(lg.Expr)
    }
    return r
}
```

### RequiresAction.Clone() (line 118-120)

**New** — preserves concrete type:
```go
func (a *RequiresAction) Clone(args []ast.Node) ast.Node {
    r := &RequiresAction{}
    r.ActionBase = a.ActionBase
    r.Kind = a.Kind
    r.Unprovable = a.Unprovable
    if len(args) >= 1 {
        if lf, ok := args[0].(*ast.LabeledFormula); ok {
            r.Formula = lf.Formula.(lg.Expr)
            r.LF = lf
        } else {
            r.Formula = args[0].(lg.Expr)
        }
    }
    if len(args) >= 2 {
        r.Proof = args[1].(lg.Expr)
    }
    return r
}
```

### EnsuresAction.Clone() (line 138-140)
Same pattern as RequiresAction.

### SubgoalAction.Clone() (line 439-441)

**New** — preserves `SubgoalKind`:
```go
func (a *SubgoalAction) Clone(args []ast.Node) ast.Node {
    r := &SubgoalAction{SubgoalKind: a.SubgoalKind}
    r.ActionBase = a.ActionBase
    r.Kind = a.Kind
    r.Unprovable = a.Unprovable
    if len(args) >= 1 {
        if lf, ok := args[0].(*ast.LabeledFormula); ok {
            r.Formula = lf.Formula.(lg.Expr)
            r.LF = lf
        } else {
            r.Formula = args[0].(lg.Expr)
        }
    }
    if len(args) >= 2 {
        r.Proof = args[1].(lg.Expr)
    }
    return r
}
```

## Step 3: Remove NodeArgs/NodeClone, use Args()/Clone() directly

Delete `NodeArgs()` and `NodeClone()` from `actions/helpers.go`. They're vestigial now.

Update all callers to use `node.Args()` / `node.Clone(newArgs)` directly:

- `substituteConstantsAST` in `actions/helpers.go`: `NodeArgs(node)` → `node.Args()`, `NodeClone(node, args)` → `node.Clone(args)`
- `assertToAssumeChildren` in `actions/transforms.go`: `NodeArgs(action)` → `action.Args()`, `NodeClone(action, newArgs).(Action)` → `action.Clone(newArgs).(Action)`
- `PrefixCallsFunc` default in `actions/transforms.go`: same
- `DropInvariants` default in `actions/transforms.go`: same
- `UnrollLoops` default in `actions/transforms.go`: same
- `recur` P23 in `compiler/phase6.go`: `actions.NodeArgs(act)` → `act.Args()`, `actions.NodeClone(act, newArgs).(actions.Action)` → `act.Clone(newArgs).(actions.Action)`

## Files to modify

1. **`actions/action_expr.go`** — Replace `Args()` and `Clone()` for 5 types: AssumeAction (lines 77-80), AssertAction (lines 97-100), RequiresAction (lines 117-120), EnsuresAction (lines 137-140), SubgoalAction (lines 438-441)

2. **`actions/helpers.go`** — Delete `NodeArgs`, `NodeClone`. Update `substituteConstantsAST` to call `node.Args()`/`node.Clone()` directly.
3. **`actions/transforms.go`** — Update `assertToAssumeChildren`, `PrefixCallsFunc`, `DropInvariants`, `UnrollLoops` to call `action.Args()`/`action.Clone()` directly.
4. **`compiler/phase6.go`** — Update `recur` P23 to call `act.Args()`/`act.Clone()` directly.

## What stays unchanged

- `ActionArgs()` / `ActionClone()` — used by Sexp(), Canon(), Modifies(), and direct field access
- `Children()` — used by Sexp(), Equal(), lg.Expr consumers
- All other action types' Args()/Clone() — already correct (they don't have LF)
- All logic types' Args()/Clone() in `ast_compat.go` — already correct
- `substituteConstantsAST` and all recursive functions — already use NodeArgs/NodeClone

## Verification

`go build ./...` then `cd ~/goivy && make golden`
