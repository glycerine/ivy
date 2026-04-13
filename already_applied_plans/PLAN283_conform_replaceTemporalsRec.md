# Full conformance audit: replaceTemporalsRec

Created: 2026-04-13 ~12:30 UTC

## Context

Two bugs already fixed in `replaceTemporalsRec` (leaf shortcut, Not in switch). A full audit reveals more divergences from the Python source of truth. The root cause is structural: Go handles Apply, Not, and NamedBinder in the switch with their own recursion logic, while Python handles them in the `else` branch after a COMMON recursion of `ast.args`.

## Python structure (ivy_logic_utils.py:295-349)

```
1. if Globally → special
2. elif Eventually → special
3. elif WhenOperator → special
4. else:
   a. args = [recurse(x) for x in ast.args]     ← COMMON recursion first
   b. if Apply:
      - if func is NamedBinder l2s_init:
        - body = recurse(ast.func.body)           ← body AFTER terms
        - if body is Not → l2s_init_not (uses nb.clone)
        - else → l2s_init (uses nb.clone)
      - else: func = recurse(ast.func)            ← func AFTER terms
        → app
   c. elif Not and args[0] is Not → doubleNeg
   d. elif NamedBinder l2s_init and args[0] is Not → nb_init_not (uses ast.clone)
   e. else → ast.clone(args) tracing "cloned"
```

## All remaining divergences

### D1: Apply l2s_init — recursion ORDER
- **Python**: recurses terms first (common step), THEN body
- **Go**: recurses body FIRST, then terms
- **Impact**: xtrace ENTER/EXIT order for children differs

### D2: Apply general — recursion ORDER
- **Python**: recurses terms first (common step), THEN func
- **Go**: recurses func FIRST, then terms
- **Impact**: xtrace ENTER/EXIT order for children differs

### D3: Apply l2s_init — NB construction method
- **Python**: `ast.func.clone([body])` — uses NamedBinder.clone() which COPIES the Variables slice
- **Go**: manual `&logic.NamedBinder{...Variables: nb.Variables...}` — shares Variables slice
- **Impact**: subtle potential aliasing bug

### D4: NamedBinder l2s_init non-Not body — trace label
- **Python**: falls to `ast.clone(args)` → traces "cloned"
- **Go**: traces "nb_init" (separate code path in switch)
- **Impact**: xtrace divergence

### D5: NamedBinder l2s_init Not body — construction method
- **Python**: `lg.Not(ast.clone([args[0].args[0]]))` — uses NamedBinder.clone()
- **Go**: manual `&logic.NamedBinder{...}` — no clone
- **Impact**: same aliasing issue as D3

## Fix: Restructure to match Python flow exactly

**File**: `/Users/jaten/ivy/goivy/logicutil/logic_utils.go`

Remove Apply and NamedBinder from the switch. Keep only Globally/Eventually/WhenOperator (matching Python's outer if/elif). Move everything else to a unified fall-through that mirrors Python's else branch exactly:

```go
func replaceTemporalsRec(n ast.Node, g GloballyBinderFunc, when WhenBinderFunc) ast.Node {
    xtracer.Trace("ENTER ...")
    switch t := n.(type) {
    case *logic.Globally:   // Python line 302 — unchanged
        ...
    case *logic.Eventually: // Python line 308 — unchanged
        ...
    case *logic.WhenOperator: // Python line 313 — unchanged
        ...
    }

    // === Python else branch (line 321) ===

    // Step 1: Common recursion of all args (Python line 322)
    children := n.Args()
    newChildren := make([]ast.Node, len(children))
    for i, c := range children {
        newChildren[i] = replaceTemporalsRec(c, g, when)
    }

    // Step 2a: Apply (Python line 323)
    if t, ok := n.(*logic.Apply); ok {
        if nb, ok := t.Func.(*logic.NamedBinder); ok && nb.Name == "l2s_init" {
            // Python line 325: recurse body AFTER terms
            body := replaceTemporalsRec(nb.Body, g, when).(logic.Expr)
            newArgs := nodesToExprs(newChildren)
            if notBody, ok := body.(*logic.Not); ok {
                // Python line 327: ast.func.clone([body.body])
                clonedNB := nb.Clone([]ast.Node{notBody.Body}).(logic.Expr)
                inner := logic.MustApply(clonedNB, newArgs...)
                result := &logic.Not{Body: inner}
                xtracer.Trace("... l2s_init_not ...")
                return result
            }
            // Python line 331: ast.func.clone([body])
            clonedNB := nb.Clone([]ast.Node{body}).(logic.Expr)
            result := logic.MustApply(clonedNB, newArgs...)
            xtracer.Trace("... l2s_init ...")
            return result
        }
        // Python line 334: recurse func AFTER terms
        newFunc := replaceTemporalsRec(t.Func, g, when).(logic.Expr)
        newArgs := nodesToExprs(newChildren)
        result := logic.MustApply(newFunc, newArgs...)
        xtracer.Trace("... app ...")
        return result
    }

    // Step 2b: Not double negation (Python line 338)
    if _, ok := n.(*logic.Not); ok && len(newChildren) > 0 {
        if inner, ok := newChildren[0].(*logic.Not); ok {
            xtracer.Trace("... doubleNeg ...")
            return inner.Body
        }
    }

    // Step 2c: NamedBinder l2s_init with Not body (Python line 342)
    if t, ok := n.(*logic.NamedBinder); ok && t.Name == "l2s_init" && len(newChildren) > 0 {
        if notChild, ok := newChildren[0].(*logic.Not); ok {
            // Python line 343: lg.Not(ast.clone([args[0].args[0]]))
            cloned := n.Clone([]ast.Node{notChild.Body})
            result := &logic.Not{Body: cloned.(logic.Expr)}
            xtracer.Trace("... nb_init_not ...")
            return result
        }
    }

    // Step 2d: Default clone (Python line 346)
    result := n.Clone(newChildren)
    xtracer.Trace("... cloned ...")
    return result
}
```

Also add a small helper (used for converting recursed terms to `[]logic.Expr`):

```go
func nodesToExprs(nodes []ast.Node) []logic.Expr {
    exprs := make([]logic.Expr, len(nodes))
    for i, n := range nodes {
        exprs[i] = n.(logic.Expr)
    }
    return exprs
}
```

## Key design notes

- `logic.Expr` embeds `ast.Node` (confirmed in `logic/node.go:8-9`), so Expr values are directly assignable to `ast.Node`
- `NamedBinder.Clone()` copies the Variables slice (confirmed in `logic/ast_compat.go`), matching Python's clone behavior — manual construction does NOT copy, so we must use Clone
- `MustApply` recomputes aSort from the func's sort — matches Python's `Apply()` constructor. `Apply.Clone()` preserves old aSort, which would NOT match
- `Apply.Args()` returns Terms only (not Func) — matches Python's `Apply.args` property

## Verification

```
cd ~/ivy/goivy && go test -run TestOrdLive -timeout 300s ./parser/
```
