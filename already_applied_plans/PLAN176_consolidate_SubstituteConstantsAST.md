# Fix: Consolidate `SubstituteConstantsAST` into one function in clauseops

**Created**: 2026-04-01 21:15

## Context

The golden test diverges at XTRACE line 157358 because Go's `clauseops.SubstituteConstantsAST` lacks the xtrace tracing that Python's `substitute_constants_ast` has. Go has TWO implementations of the same Python function:

1. `actions/helpers.go:substituteConstantsAST(ast.Node, subs)` — traces correctly, well-tested
2. `clauseops/astutil.go:SubstituteConstantsAST(lg.Expr, subs)` — no trace, untested

Python has ONE function (`ivy_logic_utils.py:173`). Go should too. Delete the clauseops version, move the actions version to clauseops, and have actions call through.

## Deep Analysis of Subtleties

### S1: `Args()/Clone()` vs `Children()/CloneNode()` equivalence

The two implementations use different traversal interfaces:
- **Actions version**: `node.Args()` → `node.Clone(newArgs)` (ast.Node interface)
- **Clauseops version**: `node.Children()` → `il.CloneNode(node, newChildren)` (lg.Expr interface)

**Verified equivalent** for all lg types in `logic/ast_compat.go`:
- `Apply.Args()` returns Terms only (same as `Children()`), `Clone()` preserves Func and aSort ✓
- `And/Or.Args()` returns Terms (same as `Children()`) ✓
- `Not.Args()` returns `[Body]` (same as `Children()`) ✓
- `Implies/Iff.Args()` returns `[T1,T2]` (same as `Children()`) ✓
- `Eq.Args()` returns `[T1,T2]` (same as `Children()`) ✓
- `ForAll/Exists/Lambda.Args()` returns `[Body]` only — `Clone()` copies Variables from original, sets Body from args[0] ✓
- `Ite.Args()` returns `[Cond,Then,Else]` ✓
- `Definition.Args()` returns `[Lhs,Rhs]` ✓ (`il.Definition = lg.Definition`, type alias, same type)
- `Variable.Args()` returns `nil`, `Clone()` returns self ✓
- `Const.Args()` → N/A (handled before Args() is called)

### S2: Variable tracing behavior change

- **Old clauseops**: `*lg.Variable` → return immediately, NO trace
- **New (from actions)**: traces `type=Var nargs=0`, then returns `node.Clone(nil)` = same variable
- **Python**: traces for Variables (they're not constants — `is_constant(Var(...))` is False)
- **Verdict**: New behavior is CORRECT and matches Python. The old version was wrong.
- `shortTypeName` maps Go's `Variable` → Python's `Var` ✓ (Python class is `lg.Var`, `type(v).__name__` = `"Var"`)

### S3: `changed` optimization removed

- Old clauseops tracks whether children changed, avoids cloning if not (preserves pointer identity)
- New version always clones (matching Python's `ast.clone(...)`)
- **Risk**: code using pointer equality (`==`) after substitution could break
- **Verdict**: Python doesn't guarantee identity; neither should Go. Matches mechanical port rules.

### S4: Apply sort validation removed

- Old clauseops: `lg.NewApply(t.Func, newTerms...)` which validates sort compatibility, returns error
- New version: `Clone()` → `&Apply{Func: a.Func, Terms: terms, aSort: a.aSort}` — no validation
- **Verdict**: Python doesn't validate in `substitute_constants_ast`. Matches Python. The sort was already correct before substitution; substituting a constant with a same-sorted variable shouldn't break it.

### S5: `len(subs) == 0` early return removed

- Old clauseops returns immediately for empty subs (no trace, no recursion)
- Python doesn't have this optimization — still traces and recurses
- **Verdict**: Remove it. Matches Python. In the current divergence subs is non-empty anyway.

### S6: Leaf node (nargs=0) tracing

- Old clauseops: returns `node` without trace for non-Const/Variable/Apply/Definition leaves with 0 children
- New version: traces `nargs=0`, then calls `node.Clone([])`
- **Python**: traces, then clones
- **Verdict**: New behavior matches Python ✓

### S7: Type assertion safety for `.(lg.Expr)`

- All `Clone()` implementations for lg types return types that implement `lg.Expr`
- `And.Clone()` → `*And` (implements Expr) ✓
- `Apply.Clone()` → `*Apply` (implements Expr) ✓
- All others similarly ✓
- Type assertion will never panic for lg inputs

### S8: `shortTypeName` must move to clauseops

Currently in `actions/helpers.go`, used by:
- `substituteConstantsAST` (moving to clauseops)
- `actions/transforms.go:515` (prefix_calls trace)
- Exported as `ShortTypeName` (line 16)

**Plan**: Move `shortTypeName` to clauseops, export as `ShortTypeName`. In actions, keep a one-line redirect: `func shortTypeName(v interface{}) string { return co.ShortTypeName(v) }`. This avoids touching `transforms.go`.

New imports needed in clauseops: `"fmt"` (already present), `"strings"` (add).

### S9: `compiler/phase6.go:699` — body is an Action

The `body` variable is an `actions.Action`. The old clauseops version would fall through to the generic `Children()/CloneNode()` path. `il.CloneNode` may or may not handle Action types correctly. The new version uses `Args()/Clone()` which Action types DO support (proven by `SubstituteConstantsAction` already working). So the new version is **more correct** here.

After substitution, `newBody` is `ast.Node` (was `lg.Expr`). Line 703 does `newBody.(actions.Action)` — this type assertion works since `Clone()` returns the same concrete Action type.

### S10: No circular dependency concerns

- `clauseops` currently imports: `ast`, `ivylogic`, `logic`, `logicutil`
- New imports needed: `xtracer`, `strings`
- `ast` does NOT import `clauseops` — no cycle ✓
- `xtracer` is a leaf package — no cycle ✓
- All caller packages already import `clauseops` — no new import edges ✓

## Changes

### Step 1: Add `ShortTypeName` and new imports to clauseops

**File: `clauseops/astutil.go`**

Add to imports: `"strings"`, `"github.com/glycerine/ivy/goivy/xtracer"`, `"github.com/glycerine/ivy/goivy/ast"`

Add exported function (from `actions/helpers.go:18-32`):
```go
func ShortTypeName(v interface{}) string {
    s := fmt.Sprintf("%T", v)
    if i := strings.LastIndex(s, "."); i >= 0 {
        s = s[i+1:]
    }
    if s == "Variable" {
        return "Var"
    }
    if s == "App" {
        return "Apply"
    }
    return s
}
```

### Step 2: Replace clauseops `SubstituteConstantsAST` + add `SubstituteConstantsExpr`

**File: `clauseops/astutil.go`**

Delete `SubstituteConstantsAST(lg.Expr, subs) lg.Expr` and `substituteConstantsRec`.

Add in their place (from `actions/helpers.go:235-259`, exported):
```go
// SubstituteConstantsAST substitutes terms for constants.
// Python: substitute_constants_ast (ivy_logic_utils.py:173)
func SubstituteConstantsAST(node ast.Node, subs map[lg.NodeKey]lg.Expr) ast.Node {
    if sym, ok := node.(*lg.Const); ok {
        if rep, found := subs[lg.Key(sym)]; found {
            return rep
        }
        return sym
    }
    args := node.Args()
    xtracer.Trace("actions.substitute_constants_action ENTER type=%s nargs=%d",
        ShortTypeName(node), len(args))
    if len(args) == 0 {
        return node.Clone(args)
    }
    newArgs := make([]ast.Node, len(args))
    for i, arg := range args {
        newArgs[i] = SubstituteConstantsAST(arg, subs)
    }
    return node.Clone(newArgs)
}

// SubstituteConstantsExpr is SubstituteConstantsAST for lg.Expr callers.
func SubstituteConstantsExpr(node lg.Expr, subs map[lg.NodeKey]lg.Expr) lg.Expr {
    return SubstituteConstantsAST(node, subs).(lg.Expr)
}
```

### Step 3: Update `actions/helpers.go`

Delete the local `substituteConstantsAST` function (lines 235–259).

Redirect `shortTypeName` to clauseops (keeps `transforms.go:515` working without changes):
```go
func shortTypeName(v interface{}) string { return co.ShortTypeName(v) }
func ShortTypeName(v interface{}) string { return co.ShortTypeName(v) }
```

Update `SubstituteConstantsAction`:
```go
func SubstituteConstantsAction(action Action, subs map[lg.NodeKey]lg.Expr) Action {
    return co.SubstituteConstantsAST(action, subs).(Action)
}
```

### Step 4: Update callers — use `SubstituteConstantsExpr` for lg.Expr callers

These callers pass `lg.Expr` and expect `lg.Expr` back. Change to `SubstituteConstantsExpr`:

- **`actions/action.go:102`**: `co.SubstituteConstantsAST(...)` → `co.SubstituteConstantsExpr(...)`
- **`actions/action.go:531`**: same
- **`actions/action.go:573`**: same
- **`actions/action.go:620`**: same
- **`ranking/tactic.go:193`**: same
- **`l2s/l2s_auto.go:205`**: same
- **`module/theory.go:341`**: same

For `compiler/phase6.go:699`: change to `co.SubstituteConstantsAST(body, subs)` (returns `ast.Node`, line 703 already does `.(actions.Action)` assertion).

### Step 5: Remove unused imports

After deletion, check for unused imports in:
- `clauseops/astutil.go` — `il` (ivylogic) is still used by `renameASTRec`, keep it
- `actions/helpers.go` — check if `ast` import is still needed; `os` and `strings` may be unused

## Verification

```bash
cd ~/ivy/goivy && go build ./... && go test ./... && make golden
```

Check that:
1. `go build` and `go test` pass
2. The divergence in `log.red` advances past line 157358 (substitute traces now match)
