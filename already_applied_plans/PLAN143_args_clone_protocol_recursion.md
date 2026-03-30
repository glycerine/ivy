# Plan: Rewrite Recursive Functions to Use Args()/Clone() from ast.Node

NOTE: ABANDONDED. NOT DONE.

**Created**: 2026-03-30 00:30

## Context

Python's recursive tree-walking functions (`substitute_constants_ast`, `assert_to_assume`, `prefix_calls`, etc.) all use the universal `.args`/`.clone()` protocol. Every node type — Actions, LabeledFormulas, formulas — participates uniformly. Go split this into `ActionArgs()`/`ActionClone()` for actions and `Children()`/`cloneExpr()` for expressions, with LabeledFormula falling between the cracks. This causes repeated recursion-order divergences.

The fix: make `AssumeAction.Args()`/`AssertAction.Args()` return the LF when present (matching Python's `.args = [LabeledFormula]`), fix their `Clone()` methods accordingly, and rewrite the recursive functions to dispatch: Action → `Args()`/`Clone()`, then Expr → `Children()`/`cloneExpr()`, then pure Node → `Args()`/`Clone()`.

## Key Observation: Dispatch Order

Go actions implement BOTH `ast.Node` and `lg.Expr`. The dispatch must check Action BEFORE Expr:

```go
if sym, ok := node.(*lg.Const); ok { ... }       // leaf constant
if _, ok := node.(actions.Action); ok { ... }      // action: Args()/Clone()
if expr, ok := node.(lg.Expr); ok { ... }          // expression: Children()/cloneExpr()
// else: pure ast.Node (LabeledFormula): Args()/Clone()
```

## Step 1: Fix `Args()` for LF-bearing action types

**File**: `actions/action_expr.go`

### AssumeAction.Args()
```go
func (a *AssumeAction) Args() []ast.Node {
    if a.LF != nil {
        return []ast.Node{a.LF}
    }
    return actionArgsToNodes(a.ActionArgs())
}
```

### AssertAction.Args() (inherited by RequiresAction, EnsuresAction, SubgoalAction)
```go
func (a *AssertAction) Args() []ast.Node {
    if a.LF != nil {
        if a.Proof != nil {
            return []ast.Node{a.LF, a.Proof}
        }
        return []ast.Node{a.LF}
    }
    return actionArgsToNodes(a.ActionArgs())
}
```

Remove the generated `Args()` for RequiresAction, EnsuresAction, SubgoalAction if they just delegate — they should inherit from AssertAction's embedded method.

## Step 2: Fix `Clone()` for LF-bearing action types

**File**: `actions/action_expr.go`

### AssumeAction.Clone()
```go
func (a *AssumeAction) Clone(args []ast.Node) ast.Node {
    if len(args) >= 1 {
        if lf, ok := args[0].(*ast.LabeledFormula); ok {
            r := &AssumeAction{
                ActionBase: a.ActionBase,
                Formula:    lf.Formula.(lg.Expr),
                LF:         lf,
                Unprovable: a.Unprovable,
            }
            return r
        }
    }
    return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
```

### AssertAction.Clone() (and RequiresAction, EnsuresAction, SubgoalAction)
Same pattern — detect LF in args[0], set both Formula and LF on the result.

## Step 3: Rewrite SubstituteConstantsAction

**File**: `actions/helpers.go`

Replace three functions (`SubstituteConstantsAction`, `substituteConstantsExpr`, `substituteConstantsNode`) with one unified function matching Python's `substitute_constants_ast`:

```go
// substituteConstantsAST matches Python's substitute_constants_ast from
// ivy_logic_utils.py:172. One function handles all node types uniformly.
func substituteConstantsAST(node ast.Node, subs map[lg.NodeKey]lg.Expr) ast.Node {
    // Python: if is_constant(ast): return subs.get(ast.rep, ast)
    if sym, ok := node.(*lg.Const); ok {
        if rep, found := subs[lg.Key(sym)]; found {
            return rep
        }
        return sym
    }

    // Actions: use Args()/Clone() — includes LF as first-class child
    if _, ok := node.(Action); ok {
        args := node.Args()
        xtracer.Trace("actions.substitute_constants_action ENTER type=%s nargs=%d",
            shortTypeName(node), len(args))
        newArgs := make([]ast.Node, len(args))
        for i, arg := range args {
            newArgs[i] = substituteConstantsAST(arg, subs)
        }
        return node.Clone(newArgs)
    }

    // lg.Expr: use Children()/cloneExpr() — logic formulas
    if expr, ok := node.(lg.Expr); ok {
        children := expr.Children()
        xtracer.Trace("actions.substitute_constants_action ENTER type=%s nargs=%d",
            shortTypeName(node), len(children))
        if len(children) == 0 {
            return node // leaf non-constant (e.g. Variable)
        }
        newChildren := make([]lg.Expr, len(children))
        for i, c := range children {
            result := substituteConstantsAST(c, subs)
            newChildren[i] = result.(lg.Expr)
        }
        return cloneExpr(expr, newChildren)
    }

    // Pure ast.Node (LabeledFormula, Atom labels, etc.): use Args()/Clone()
    args := node.Args()
    xtracer.Trace("actions.substitute_constants_action ENTER type=%s nargs=%d",
        shortTypeName(node), len(args))
    if len(args) == 0 {
        return node.Clone(args) // leaf, clone with empty args
    }
    newArgs := make([]ast.Node, len(args))
    for i, arg := range args {
        newArgs[i] = substituteConstantsAST(arg, subs)
    }
    return node.Clone(newArgs)
}

// SubstituteConstantsAction is the entry point for callers expecting Action return type.
func SubstituteConstantsAction(action Action, subs map[lg.NodeKey]lg.Expr) Action {
    return substituteConstantsAST(action, subs).(Action)
}
```

Remove: `substituteConstantsExpr`, `substituteConstantsNode`, all LFBearer handling in SubstituteConstantsAction.

## Step 4: Rewrite other recursive functions

**File**: `actions/transforms.go`

### assertToAssumeChildren
```go
func assertToAssumeChildren(action Action, kinds map[string]bool, ...) Action {
    args := action.Args()  // includes LF when present
    newArgs := make([]ast.Node, len(args))
    for i, arg := range args {
        if child, ok := arg.(Action); ok {
            newArgs[i] = AssertToAssume(child, kinds, ...)
        } else {
            newArgs[i] = arg  // LF, formulas pass through unchanged
        }
    }
    return action.Clone(newArgs).(Action)
}
```

### PrefixCallsFunc (default case)
```go
    args := action.Args()
    newArgs := make([]ast.Node, len(args))
    for i, arg := range args {
        if child, ok := arg.(Action); ok {
            newArgs[i] = PrefixCallsFunc(child, renamer)
        } else {
            newArgs[i] = arg
        }
    }
    return action.Clone(newArgs).(Action)
```

### DropInvariants, UnrollLoops
Same pattern.

### recur (compiler/phase6.go, generic path P23)
```go
    args := act.Args()
    newArgs := make([]ast.Node, len(args))
    for i, arg := range args {
        if subAct, ok := arg.(actions.Action); ok {
            newArgs[i] = recur(subAct)
        } else {
            newArgs[i] = arg
        }
    }
    return act.Clone(newArgs).(actions.Action)
```

## Step 5: Remove LFBearer workaround

**File**: `actions/helpers.go`

Remove `LFBearer` interface and all `LFBearer`-related code from `SubstituteConstantsAction`. The LF is now handled as a natural child via `Args()`.

## Files to Modify

1. `actions/action_expr.go` — Fix `Args()` and `Clone()` for AssumeAction, AssertAction, RequiresAction, EnsuresAction, SubgoalAction
2. `actions/helpers.go` — Rewrite `SubstituteConstantsAction` as unified `substituteConstantsAST`, remove `substituteConstantsExpr`, `substituteConstantsNode`, LFBearer handling
3. `actions/transforms.go` — Rewrite `assertToAssumeChildren`, `PrefixCallsFunc`, `DropInvariants`, `UnrollLoops` to use `Args()`/`Clone()`
4. `compiler/phase6.go` — Rewrite `recur` generic path to use `Args()`/`Clone()`

## ActionArgs()/ActionClone() are NOT removed

These remain for:
- Non-recursive callers that need typed `[]lg.Expr` access (e.g., reading `a.Formula`)
- `Sexp()`, `Canon()`, and other serialization methods
- `Modifies()`, `References()` (analysis-only recursive functions — convert later if needed)

## Verification

`go build ./...` then `cd ~/goivy && make golden`

The trace at line 147244 should match (LabeledFormula nargs=2 on both sides), and the golden test should advance significantly past 147682.
