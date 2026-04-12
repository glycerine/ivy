# Fix ActionDef.Rewrite Dispatch Bug — Extra AstRewrite on Formal Params

**Created:** 2026-04-13 ~12:15am

## Context

Golden test `TestOrdLive` diverges at trace line 929 (log.red:117). Go emits `AstRewrite type(x)=App canon=(app rep:fml:x terms:[] aSort:t)` — an extra recursive AstRewrite on a formal parameter — while Python has already finished and emits `parser.spaa EXIT`. The extra App is `fml:x`, a formal parameter of ActionDef.

**Root cause:** Two bugs in Go's `ActionDef`:

1. **`ActionDef.Rewrite()` returns `*ActionDef` instead of `Node`** (decl_ast.go:506). This means `*ActionDef` does NOT satisfy the `AstRewritable` interface (which requires `Rewrite(AstRewriter) Node`). So when `AstRewrite` hits an ActionDef in the default case, the `x.(AstRewritable)` type assertion fails, and it falls through to the generic `x.Args()` path.

2. **`ActionDef.Args()` includes FormalParams and FormalReturns** (decl_ast.go:446-451): returns `[Name, Body, FormalParams..., FormalReturns...]`. The generic path then calls `AstRewriteSlice` on ALL of them — including the formals — producing extra AstRewrite XTRACE traces that Python doesn't produce.

**Python's behavior:** `ast_rewrite` checks `hasattr(x, 'rewrite')` (ivy_ast.py:1729) → True for ActionDef → calls `ActionDef.rewrite(rewrite)` (ivy_ast.py:1408) which:
- Only passes `self.args = [atom, action]` (Name + Body, NOT formals) through `ast_rewrite`
- Rewrites formals separately via `rewrite_param(p, rewrite)` (ivy_ast.py:1421) — a simple constructor call + sort rewrite, NOT a recursive `ast_rewrite` (no AstRewrite XTRACE traces)

**Secondary bug:** Go's `rewriteParams()` (decl_ast.go:514-523) calls full `AstRewrite(p, rw)` instead of matching Python's simpler `rewrite_param`. Even when the Rewrite dispatch is fixed, the formal param rewrite would still differ.

---

## Fix Plan

### Fix 1: Change `ActionDef.Rewrite()` return type to `Node` (decl_ast.go:506)

**File:** `ast/decl_ast.go`

```go
// BEFORE (line 506):
func (a *ActionDef) Rewrite(rw AstRewriter) *ActionDef {

// AFTER:
func (a *ActionDef) Rewrite(rw AstRewriter) Node {
```

This satisfies `AstRewritable` interface, so `AstRewrite`'s default case dispatches to it instead of falling through to the generic `x.Args()` path.

### Fix 2: Fix `Rewrite()` body — only rewrite Name+Body, not formals (decl_ast.go:507)

**File:** `ast/decl_ast.go`

```go
// BEFORE (line 507):
res := a.Clone(AstRewriteSlice(a.Args(), rw)).(*ActionDef)

// AFTER — match Python: ast_rewrite(self.args, rewrite) where self.args = [atom, action]
rewrittenNameBody := AstRewriteSlice([]Node{a.Name, a.Body}, rw)
allArgs := make([]Node, 0, 2+len(a.FormalParams)+len(a.FormalReturns))
allArgs = append(allArgs, rewrittenNameBody...)
allArgs = append(allArgs, a.FormalParams...)
allArgs = append(allArgs, a.FormalReturns...)
res := a.Clone(allArgs).(*ActionDef)
```

Clone expects `[Name, Body, params..., returns...]`. We pass rewritten Name+Body with OLD formals (which Clone will copy), then overwrite formals via rewriteParams.

### Fix 3: Fix `rewriteParams` to match Python's `rewrite_param` (decl_ast.go:514-523)

**File:** `ast/decl_ast.go`

Python's `rewrite_param` (ivy_ast.py:1421-1424):
```python
def rewrite_param(p, rewrite):
    res = type(p)(p.rep)          # fresh node with same rep, no args
    res.sort = rewrite_sort(rewrite, p.sort)  # rewrite sort only
    return res
```

Go's current `rewriteParams` calls full `AstRewrite(p, rw)` — wrong. Fix:

```go
func rewriteParams(params []Node, rw AstRewriter) []Node {
    if len(params) == 0 {
        return nil
    }
    result := make([]Node, len(params))
    for i, p := range params {
        result[i] = rewriteParam(p, rw)
    }
    return result
}

// rewriteParam matches Python rewrite_param (ivy_ast.py:1421-1424):
//   res = type(p)(p.rep)
//   res.sort = rewrite_sort(rewrite, p.sort)
func rewriteParam(p Node, rw AstRewriter) Node {
    switch n := p.(type) {
    case *App:
        repStr := NodeRep(n.Rep)
        repSym := &Symbol{Rep: repStr}
        repSym.Cfg = n.Cfg
        res := &App{Rep: repSym}
        res.Cfg = n.Cfg
        if n.ASort != nil {
            sortStr := fmt.Sprint(n.ASort)
            newSort := RewriteSort(rw, sortStr, n.Cfg)
            ss := &Symbol{Rep: newSort}
            ss.Cfg = n.Cfg
            res.ASort = ss
        }
        return res
    case *Atom:
        res := &Atom{Rep: n.Rep}
        res.Cfg = n.Cfg
        if n.ASort != nil {
            sortStr := fmt.Sprint(n.ASort)
            newSort := RewriteSort(rw, sortStr, n.Cfg)
            ss := &Symbol{Rep: newSort}
            ss.Cfg = n.Cfg
            res.ASort = ss
        }
        return res
    default:
        // Fallback — formal params should always be App or Atom
        return AstRewrite(p, rw)
    }
}
```

### Fix 4: Check all callers of `ActionDef.Rewrite()`

The return type changes from `*ActionDef` to `Node`. Any callers that expect `*ActionDef` must add a type assertion. Search for `.Rewrite(` calls on ActionDef values.

### Fix 5: Add xtracer traces for ActionDef.rewrite dispatch (both Go and Python)

**Go:** `ast/decl_ast.go` — Add trace inside Rewrite:
```go
func (a *ActionDef) Rewrite(rw AstRewriter) Node {
    xtracer.Trace("ActionDef.rewrite ENTER nParams=%d nReturns=%d", len(a.FormalParams), len(a.FormalReturns))
    // ... body ...
    xtracer.Trace("ActionDef.rewrite EXIT nParams=%d nReturns=%d", len(res.FormalParams), len(res.FormalReturns))
    return res
}
```

**Python:** `ivy_ast.py:1408` — Add matching trace:
```python
def rewrite(self, rewrite):
    if __debug__: xtracer.trace("ActionDef.rewrite ENTER nParams=%d nReturns=%d" % (len(self.formal_params), len(self.formal_returns)))
    res = self.clone(ast_rewrite(self.args, rewrite))
    if hasattr(self, 'formal_params'):
        res.formal_params = [rewrite_param(p, rewrite) for p in self.formal_params]
    if hasattr(self, 'formal_returns'):
        res.formal_returns = [rewrite_param(p, rewrite) for p in self.formal_returns]
    if __debug__: xtracer.trace("ActionDef.rewrite EXIT nParams=%d nReturns=%d" % (len(res.formal_params), len(res.formal_returns)))
    return res
```

---

## Files to Modify

| File | Changes |
|------|---------|
| `ast/decl_ast.go` | Fix Rewrite return type, fix body, fix rewriteParams, add traces |
| `~/pyivy/ivy/ivy/ivy_ast.py` | Add matching ActionDef.rewrite traces |

---

## Verification

1. `go build ./...` — compile clean
2. `cd ~/ivy/goivy/parser && go test -v -count=1 -run TestOrdLive` — golden test
3. Confirm line 929 no longer diverges (both sides should show ActionDef.rewrite ENTER/EXIT instead of extra App AstRewrite)
4. Check the next divergence point (if any) — the underlying l2s pipeline divergence from the previous plan may now be visible further along
