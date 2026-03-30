# Plan: Uniform Child-Access Protocol via Two Dispatch Functions

**Created**: 2026-03-30 01:30

## Context

Python's recursive tree-walking functions use `.args`/`.clone()` uniformly on every node type. Go has three separate protocols (`ActionArgs`/`ActionClone`, `Children`/`cloneExpr`, `Args`/`Clone`), causing the LabeledFormula recursion-order bug.

**Key finding from research**: Go's `Args()`/`Clone()` on all logic types (Apply, And, Or, ForAll, etc.) **already match Python's `.args`/`.clone()` exactly**:
- `Apply.Args()` returns Terms only (not Func) — matches Python's `Apply.args`
- `Apply.Clone(newTerms)` preserves Func — matches Python's `Apply.clone()`
- `ForAll.Args()` returns `[Body]` only (not Variables) — matches Python
- `ForAll.Clone([newBody])` preserves Variables — matches Python
- `Const.Args()` returns nil, `Clone()` returns self — leaf, matches Python
- Same for all other logic types (verified in `logic/ast_compat.go`)

**The only broken types** are the 5 LF-bearing action types whose `Args()` unwraps the LabeledFormula:
- `AssumeAction.Args()` → returns `[Formula]` instead of `[LF]`
- `AssertAction.Args()` → returns `[Formula]` or `[Formula, Proof]` instead of `[LF]` or `[LF, Proof]`
- `RequiresAction`, `EnsuresAction`, `SubgoalAction` — same issue via embedding

## Solution: Two package-level dispatch functions

**File**: `actions/helpers.go` (new functions near top of file)

These encapsulate ALL the special-case logic in one place. Every recursive function calls these instead of `node.Args()` / `node.Clone()` directly.

### `NodeArgs(node ast.Node) []ast.Node`

```go
// NodeArgs returns the children of node matching Python's .args protocol.
// For all types except 5 LF-bearing actions, node.Args() is already correct.
// For LF-bearing actions, we return the LF instead of the unwrapped Formula.
func NodeArgs(node ast.Node) []ast.Node {
    switch a := node.(type) {
    case *AssumeAction:
        if a.LF != nil {
            return []ast.Node{a.LF}
        }
    case *SubgoalAction:
        // Must check BEFORE *AssertAction since SubgoalAction embeds it
        if a.LF != nil {
            if a.Proof != nil {
                return []ast.Node{a.LF, a.Proof}
            }
            return []ast.Node{a.LF}
        }
    case *RequiresAction:
        if a.LF != nil {
            if a.Proof != nil {
                return []ast.Node{a.LF, a.Proof}
            }
            return []ast.Node{a.LF}
        }
    case *EnsuresAction:
        if a.LF != nil {
            if a.Proof != nil {
                return []ast.Node{a.LF, a.Proof}
            }
            return []ast.Node{a.LF}
        }
    case *AssertAction:
        // Plain AssertAction (checked after subclasses)
        if a.LF != nil {
            if a.Proof != nil {
                return []ast.Node{a.LF, a.Proof}
            }
            return []ast.Node{a.LF}
        }
    }
    // Everything else: logic types, other actions, LabeledFormula, ast.Atom, etc.
    // node.Args() already returns the correct Python-compatible children.
    return node.Args()
}
```

**Type switch ordering**: SubgoalAction, RequiresAction, EnsuresAction MUST come before AssertAction because they embed it. Go's type switch matches the most specific type first only if listed first.

### `NodeClone(node ast.Node, newArgs []ast.Node) ast.Node`

```go
// NodeClone clones node with newArgs, matching Python's .clone() protocol.
// For LF-bearing actions, if newArgs[0] is a LabeledFormula, we set both
// Formula and LF on the result, preserving the concrete action type.
func NodeClone(node ast.Node, newArgs []ast.Node) ast.Node {
    // Check if first arg is LF — only matters for LF-bearing actions
    var lf *ast.LabeledFormula
    if len(newArgs) > 0 {
        lf, _ = newArgs[0].(*ast.LabeledFormula)
    }

    if lf != nil {
        switch a := node.(type) {
        case *AssumeAction:
            return &AssumeAction{
                ActionBase: a.ActionBase,
                Formula:    lf.Formula.(lg.Expr),
                LF:         lf,
                Unprovable: a.Unprovable,
            }
        case *SubgoalAction:
            r := &SubgoalAction{
                AssertAction: AssertAction{
                    ActionBase: a.ActionBase,
                    Formula:    lf.Formula.(lg.Expr),
                    LF:         lf,
                },
                SubgoalKind: a.SubgoalKind,
            }
            if len(newArgs) > 1 {
                r.Proof = newArgs[1].(lg.Expr)
            }
            return r
        case *RequiresAction:
            r := &RequiresAction{
                AssertAction: AssertAction{
                    ActionBase: a.ActionBase,
                    Formula:    lf.Formula.(lg.Expr),
                    LF:         lf,
                },
            }
            if len(newArgs) > 1 {
                r.Proof = newArgs[1].(lg.Expr)
            }
            return r
        case *EnsuresAction:
            r := &EnsuresAction{
                AssertAction: AssertAction{
                    ActionBase: a.ActionBase,
                    Formula:    lf.Formula.(lg.Expr),
                    LF:         lf,
                },
            }
            if len(newArgs) > 1 {
                r.Proof = newArgs[1].(lg.Expr)
            }
            return r
        case *AssertAction:
            r := &AssertAction{
                ActionBase: a.ActionBase,
                Formula:    lf.Formula.(lg.Expr),
                LF:         lf,
            }
            if len(newArgs) > 1 {
                r.Proof = newArgs[1].(lg.Expr)
            }
            return r
        }
    }

    // Everything else: node.Clone() already works correctly.
    return node.Clone(newArgs)
}
```

## Rewriting the recursive functions

With `NodeArgs`/`NodeClone` in place, `SubstituteConstantsAction` becomes one unified function:

### `substituteConstantsAST` (replaces 3 functions)

**File**: `actions/helpers.go`

```go
// substituteConstantsAST matches Python's substitute_constants_ast
// (ivy_logic_utils.py:172). One function handles all node types.
func substituteConstantsAST(node ast.Node, subs map[lg.NodeKey]lg.Expr) ast.Node {
    // Python: if is_constant(ast): return subs.get(ast.rep, ast)
    if sym, ok := node.(*lg.Const); ok {
        if rep, found := subs[lg.Key(sym)]; found {
            return rep
        }
        return sym
    }

    args := NodeArgs(node)
    xtracer.Trace("actions.substitute_constants_action ENTER type=%s nargs=%d",
        shortTypeName(node), len(args))

    if len(args) == 0 {
        // Leaf non-constant (Variable, Atom label, etc.)
        // Python traces then clones with empty args.
        return NodeClone(node, args)
    }

    newArgs := make([]ast.Node, len(args))
    for i, arg := range args {
        newArgs[i] = substituteConstantsAST(arg, subs)
    }
    return NodeClone(node, newArgs)
}

// SubstituteConstantsAction is the entry point returning Action type.
func SubstituteConstantsAction(action Action, subs map[lg.NodeKey]lg.Expr) Action {
    return substituteConstantsAST(action, subs).(Action)
}
```

**Remove**: `substituteConstantsExpr`, `substituteConstantsNode`, all LFBearer code, `substituteLookup`, `cloneExpr`.

Wait — `cloneExpr` is still needed by `NodeClone`'s fallback `node.Clone(newArgs)` path. Actually no — `node.Clone(newArgs)` calls the type's own `Clone()` method which handles reconstruction. `cloneExpr` was only needed when we had a separate `substituteConstantsExpr` that returned `lg.Expr` and needed to clone expressions. With the unified approach, `node.Clone(newArgs)` handles everything.

Actually, `cloneExpr` IS still used? Let me check. The `Clone()` methods on logic types (in `ast_compat.go`) handle their own reconstruction:
- `Apply.Clone(args) → &Apply{Func: a.Func, Terms: cast(args)}`
- `And.Clone(args) → &And{Terms: cast(args)}`
- etc.

So `cloneExpr` is NOT needed. `NodeClone` calls `node.Clone(newArgs)` which dispatches to the correct type-specific method.

**Remove**: `cloneExpr` (dead code after this change).

### Other recursive functions

**File**: `actions/transforms.go`

#### assertToAssumeChildren
```go
func assertToAssumeChildren(action Action, kinds map[string]bool, iuCfg ...*iu.IvyUtilsConfig) Action {
    args := NodeArgs(action)
    newArgs := make([]ast.Node, len(args))
    for i, arg := range args {
        if child, ok := arg.(Action); ok {
            newArgs[i] = AssertToAssume(child, kinds, iuCfg...)
        } else {
            newArgs[i] = arg // LF, formulas pass through unchanged
        }
    }
    return NodeClone(action, newArgs).(Action)
}
```

#### PrefixCallsFunc (default case)
```go
    default:
        args := NodeArgs(action)
        newArgs := make([]ast.Node, len(args))
        for i, arg := range args {
            if child, ok := arg.(Action); ok {
                newArgs[i] = PrefixCallsFunc(child, renamer)
            } else {
                newArgs[i] = arg
            }
        }
        return NodeClone(action, newArgs).(Action)
```

#### DropInvariants (default case), UnrollLoops (default case)
Same pattern.

### recur (compiler/phase6.go, generic path P23)

```go
    args := actions.NodeArgs(act)
    newArgs := make([]ast.Node, len(args))
    for i, arg := range args {
        if subAct, ok := arg.(actions.Action); ok {
            newArgs[i] = recur(subAct)
        } else {
            newArgs[i] = arg
        }
    }
    return actions.NodeClone(act, newArgs).(actions.Action)
```

## What gets removed

1. `substituteConstantsExpr` — replaced by unified `substituteConstantsAST`
2. `substituteConstantsNode` — replaced by unified `substituteConstantsAST`
3. `cloneExpr` — no longer needed; `node.Clone()` handles all types
4. `LFBearer` interface and all related code
5. `substituteLookup` helper — inlined into `substituteConstantsAST`

## What stays unchanged

- `ActionArgs()` / `ActionClone()` on all types — used by Sexp(), Canon(), Modifies(), References(), and non-recursive callers
- `Args()` / `Clone()` implementations in `action_expr.go` — still delegate to ActionArgs/ActionClone (the LF fix is in NodeArgs/NodeClone, not in the methods)
- `Children()` on all types — used by Sexp(), Equal(), and other lg.Expr consumers
- All logic type `Args()`/`Clone()` in `ast_compat.go` — already correct

## Files to modify

1. `actions/helpers.go` — Add `NodeArgs`, `NodeClone`, `substituteConstantsAST`. Remove `substituteConstantsExpr`, `substituteConstantsNode`, `cloneExpr`, `LFBearer`, `substituteLookup`.
2. `actions/transforms.go` — Rewrite `assertToAssumeChildren`, `PrefixCallsFunc` default, `DropInvariants` default, `UnrollLoops` default to use `NodeArgs`/`NodeClone`.
3. `compiler/phase6.go` — Rewrite `recur` generic path (P23) to use `actions.NodeArgs`/`actions.NodeClone`.
4. `actions/action.go` — Remove `LFBearer` interface definition and `GetLF`/`SetLF` methods (if no other callers).

## Verification

`go build ./...` then `cd ~/goivy && make golden`

Expected: golden test advances significantly past 147682.
