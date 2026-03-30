# Plan: Make CallAction Use AstCallee (Atom) for Tree-Walking, Matching Python

**Created**: 2026-03-30 04:00

## Context

Python's `CallAction.args[0]` is always an `ivy_ast.Atom`. Go's `CallAction` stores TWO representations: a compiled `Callee` (`lg.Expr` — either `*lg.Const` or `*lg.Apply`) and an `AstCallee` (`*ast.Atom`). The tree-walking interface (`Args()`/`Clone()`) currently exposes the compiled form, which differs structurally from Python's Atom. This causes divergences in:

1. `substitute_constants_action`: Go recurses into `*lg.Apply` (3 args), Python into `Atom` (3 args) — different type names in XTRACE
2. `prefix_calls`: Go only handles `*lg.Const` callees, silently skips `*lg.Apply` callees; Python always prefixes the Atom
3. Any future recursive operation

**Key insight**: Go's `AstCallee.Terms` already stores the SAME compiled `lg.Expr` objects as the `lg.Apply.Terms` (since `lg.Expr` embeds `ast.Node`). The data is already there — we just need to expose it through `Args()`/`Clone()`.

## Step 1: Change `CallAction.Args()` to return AstCallee

**File**: `actions/action_expr.go` (line 342)

**Current**:
```go
func (a *CallAction) Args() []ast.Node { return actionArgsToNodes(a.ActionArgs()) }
```

**New**:
```go
func (a *CallAction) Args() []ast.Node {
    // Python: CallAction.args = [Atom(name, compiled_args), *actual_returns]
    // Return AstCallee (Atom) as first child, matching Python's .args[0].
    first := ast.Node(a.AstCallee)
    if a.AstCallee == nil {
        // Fallback for CallActions without AstCallee (e.g., from tests)
        first = a.Callee
    }
    result := make([]ast.Node, 0, 1+len(a.ActualReturns))
    result = append(result, first)
    for _, r := range a.ActualReturns {
        result = append(result, r)
    }
    return result
}
```

## Step 2: Change `CallAction.Clone()` to handle Atom input

**File**: `actions/action_expr.go` (line 343-345)

**Current**:
```go
func (a *CallAction) Clone(args []ast.Node) ast.Node {
    return a.ActionClone(nodesToExprs(args)).(ast.Node)
}
```

**New**:
```go
func (a *CallAction) Clone(args []ast.Node) ast.Node {
    // When Args() returns [AstCallee(Atom), ...returns], the recursive
    // functions (substituteConstantsAST, etc.) process the Atom's children
    // and clone it, producing a new Atom as args[0].
    var newCallee lg.Expr
    var newAstCallee *ast.Atom

    if atom, ok := args[0].(*ast.Atom); ok {
        newAstCallee = atom
        // Reconstruct compiled callee from atom
        newCallee = calleeFromAtom(atom)
    } else {
        // Fallback: args[0] is an lg.Expr (from ActionClone path or tests)
        newCallee = args[0].(lg.Expr)
        newAstCallee = a.AstCallee
    }

    returns := make([]lg.Expr, len(args)-1)
    for i, arg := range args[1:] {
        returns[i] = arg.(lg.Expr)
    }

    if a.ActCfg == nil {
        panic("we should have a.ActCfg set!")
    }
    r := NewCallActionOn(a.ActCfg, newCallee, returns...)
    r.ActionBase = a.ActionBase
    r.AstCallee = newAstCallee
    return r
}
```

## Step 3: Add `calleeFromAtom` helper

**File**: `actions/action.go` (near CallAction definition)

This reconstructs the compiled `lg.Expr` from an `*ast.Atom`, which is the inverse of what the compiler does.

```go
// calleeFromAtom reconstructs a compiled lg.Expr callee from an ast.Atom.
// This is the inverse of the compiler's pattern:
//   lg.NewConst(rep) or lg.NewApply(lg.NewConst(rep), terms...)
// The Atom's Terms already hold the compiled lg.Expr objects.
func calleeFromAtom(atom *ast.Atom) lg.Expr {
    nameConst := lg.NewConst(atom.Rep, lg.TopS)
    if len(atom.Terms) == 0 {
        return nameConst
    }
    terms := make([]lg.Expr, len(atom.Terms))
    for i, t := range atom.Terms {
        terms[i] = t.(lg.Expr)
    }
    applied, err := lg.NewApply(nameConst, terms...)
    if err != nil {
        // Fallback: construct Apply without sort checking
        return &lg.Apply{Func: nameConst, Terms: terms}
    }
    return applied
}
```

## Step 4: Simplify `PrefixCallsFunc` CallAction case

**File**: `actions/transforms.go` (lines 344-362)

Now that Args() returns the Atom, PrefixCallsFunc can use Atom.Rename() directly, matching Python's `self.args[0].rename(pref(self.args[0].rep))`:

```go
case *CallAction:
    // Python: CallAction.prefix_calls always creates new CallAction
    // with self.args[0].rename(pref(self.args[0].rep))
    if a.AstCallee != nil {
        newAtom := a.AstCallee.Rename(renamer(a.AstCallee.Rep))
        newCallee := calleeFromAtom(newAtom)
        if a.ActCfg == nil {
            panic("we should have a.ActCfg set!")
        }
        newCall := NewCallActionOn(a.ActCfg, newCallee, a.ActualReturns...)
        newCall.ActionBase = a.ActionBase
        newCall.AstCallee = newAtom
        a.ActionBase.CopyFormalsTo(newCall)
        return newCall
    }
    // Fallback for no AstCallee
    if c, ok := a.Callee.(*lg.Const); ok {
        newCallee := lg.NewConst(renamer(c.Name), c.CSort)
        newCall := NewCallActionOn(a.ActCfg, newCallee, a.ActualReturns...)
        newCall.ActionBase = a.ActionBase
        a.ActionBase.CopyFormalsTo(newCall)
        return newCall
    }
    return a
```

## Step 5: Fix missing `copy_formals` in transform default cases

**File**: `actions/transforms.go`

Python's `Action.prefix_calls`, `Action.assert_to_assume`, `Action.drop_invariants`, and `Action.unroll_loops` all call `self.copy_formals(res)` after cloning. Go's default cases don't. Add `CopyFormalsTo` to each:

1. **`PrefixCallsFunc` default** (~line 374): add `CopyFormalsTo(action, res)` before return
2. **`assertToAssumeChildren`** (~line 89): add `CopyFormalsTo(action, res)` before return
3. **`DropInvariants` default** (~line 403): add `CopyFormalsTo(action, res)` before return
4. **`UnrollLoops` default** (~line 439): add `CopyFormalsTo(action, res)` before return

## Step 6: Update tests

**File**: `actions/audit51_test.go`

Tests that check `call.Callee.(*lg.Const)` after PrefixCalls will still work since PrefixCallsFunc creates a new CallAction with a compiled callee. Tests that create CallActions without AstCallee should still work via the fallback paths.

## Files to modify

1. **`actions/action_expr.go`** — Change `Args()` and `Clone()` for CallAction
2. **`actions/action.go`** — Add `calleeFromAtom()` helper
3. **`actions/transforms.go`** — Simplify PrefixCallsFunc CallAction case; add CopyFormalsTo in default cases
4. **Remove diagnostic traces** added earlier (the `callee_type` trace in transforms.go)

## What stays unchanged

- `Callee` field type (`lg.Expr`) — kept for internal use by SplitReturns, update.go, isolate.go
- `AstCallee` field type (`*ast.Atom`) — already correct
- `ActionArgs()` / `ActionClone()` — used by Sexp, Canon, and other internal consumers
- `Children()` — returns ActionArgs() for lg.Expr consumers
- Compiler creation sites — already set both Callee and AstCallee correctly
- `CalleeName()`, `IterCalls()`, `String()` — use Callee directly, still work

## Verification

`go build ./...` then `cd ~/goivy && make golden`
