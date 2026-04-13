# Plan: Widen modPass/transform from logic.Expr to ast.Node

Created: 2026-04-13 ~00:30

## Context

**Problem**: Go/Python xtrace divergence at golden test line 319097 (log.red:3663-3664):
- Go: `ilu.replaceTemporalsRec ENTER type=Not HASH canon=(Not body:(Symbol ...))`
- Python: `ilu.replaceTemporalsRec ENTER type=LabeledFormula HASH canon=(labeledFormula label:(...) formula:(Not ...) ...)`

Go's `modPass` (l2s.go:539) manually unwraps `LabeledFormula.Formula` before calling `transform`:
```go
model.Invars[i] = inv.Clone([]ast.Node{inv.Label, transform(inv.Formula.(lg.Expr))}).(*ast.LabeledFormula)
```
Python's `mod_pass` (ivy_l2s.py:766) passes the whole LabeledFormula:
```python
model.invars[i] = transform(inv)
```

**Root cause**: Go's three transform functions (`replaceTemporalsRec`, `NormalizeNamedBinders`, `ReplaceNamedBindersAst`) only accept `logic.Expr`. `*ast.LabeledFormula` doesn't implement `logic.Expr`, so `modPass` was forced to unwrap/re-wrap. Python's functions are duck-typed and accept any AST node — the generic fallback uses `ast.args` / `ast.clone(args)` which work on all node types.

**Why type assertions are safe**: When recursing within specific logic-type cases (Globally, Not, Apply, etc.), the input is known `logic.Expr` and the transform preserves concrete type. `LabeledFormula` only enters the generic fallback path which uses `ast.Node` methods.

## Changes

### File 1: `logicutil/logic_utils.go`

Add import: `"github.com/glycerine/ivy/goivy/ast"` (no cycle — `ast` doesn't import `logicutil`).

**A. `ReplaceTemporalsByNamedBinder` (line 986)**: Change signature:
```go
func ReplaceTemporalsByNamedBinder(n ast.Node, g GloballyBinderFunc, when WhenBinderFunc) ast.Node
```

**B. `replaceTemporalsRec` (line 996)**: Change signature:
```go
func replaceTemporalsRec(n ast.Node, g GloballyBinderFunc, when WhenBinderFunc) ast.Node
```
- Rename param from `ast` to `n` to avoid shadowing the new `ast` package import.
- Xtrace (line 997): `n.Canon()` instead of `ast.Sexp()` (same output for logic types: `Canon() = Canonical(Sexp())`).
- Type-switch: `switch t := n.(type)` — all concrete cases still match.
- Inside specific cases, add `.(logic.Expr)` on recursive calls where result feeds `logic.Expr`-typed operations. Examples:
  - `body := replaceTemporalsRec(t.Body, g, when).(logic.Expr)` (Globally, Not, NamedBinder cases)
  - `val := replaceTemporalsRec(t.T1, g, when).(logic.Expr)` (WhenOperator)
  - `newFunc := replaceTemporalsRec(t.Func, g, when).(logic.Expr)` (Apply)
  - `newArgs[i] = replaceTemporalsRec(a, g, when).(logic.Expr)` (Apply terms)
  - `newTerms[i] = replaceTemporalsRec(a, g, when).(logic.Expr)` (Apply terms)
- **Generic fallback** (lines 1077-1097): Replace `Children()` / `cloneNode()` with `Args()` / `Clone()`:
  ```go
  args := n.Args()
  if len(args) == 0 { ... return n }
  newArgs := make([]ast.Node, len(args))
  for i, a := range args {
      na := replaceTemporalsRec(a, g, when)
      newArgs[i] = na
      if na != a { changed = true }
  }
  if !changed { return n }
  result := n.Clone(newArgs)
  ```
- Update all remaining xtrace `.Sexp()` → `.Canon()` in this function.

**C. `NormalizeNamedBinders` (line 888)**: Same widening pattern:
```go
func NormalizeNamedBinders(n ast.Node, names map[string]bool) ast.Node
```
- Xtrace: `.Sexp()` → `.Canon()`.
- Type assertions on recursive calls within NamedBinder/Apply cases:
  - `body := NormalizeNamedBinders(nb.Body, names).(logic.Expr)` (line 910)
  - `newFunc := NormalizeNamedBinders(app.Func, names).(logic.Expr)` (line 919)
  - `newTerms[i] = NormalizeNamedBinders(t, names).(logic.Expr)` (line 922)
- Generic fallback: `Args()` / `Clone()` instead of `Children()` / `cloneNode()`.

**D. `ReplaceNamedBindersAst` (line 1166)**: Same pattern:
```go
func ReplaceNamedBindersAst(n ast.Node, subs map[string]logic.Expr) ast.Node
```
- Type assertions in Apply case:
  - `newTerms[i] = ReplaceNamedBindersAst(t, subs).(logic.Expr)` (line 1180)
- Generic fallback: `Args()` / `Clone()`.

**E. Internal caller** in `reduceNamedBindersRec` (line 1134):
```go
newFunc := NormalizeNamedBinders(app.Func, nil).(logic.Expr)
```

### File 2: `check/l2s_shared.go`

- **Line 50** `InstrumentationConfig.ReplaceTemporals`:
  `func(lg.Expr) lg.Expr` → `func(ast.Node) ast.Node`
- **Line 76** `SharedStep1_ConvertTemporals` modPass param:
  `func(string, func(lg.Expr) lg.Expr)` → `func(string, func(ast.Node) ast.Node)`
- **Line 97** `cfg.ReplaceTemporals` lambda:
  `func(n ast.Node) ast.Node { return lu.ReplaceTemporalsByNamedBinder(n, ...) }`
- **Line 109** `cfg.NotLf` — add type assertion:
  `cfg.NotLf = cfg.ReplaceTemporals(&lg.Not{Body: cfg.Fmla}).(lg.Expr)`
- **Line 112** NormalizeNamedBinders callback:
  `func(n ast.Node) ast.Node { return lu.NormalizeNamedBinders(n, nil) }`
- **Line 182** `normNotLf` — add type assertion:
  `normNotLf := lu.NormalizeNamedBinders(cfg.NotLf, nil).(lg.Expr)`
- **Line 565** `SharedStep11_ReplaceNamedBinders` modPass param: same change.
- **Line 590** ReplaceNamedBindersAst callback:
  `func(n ast.Node) ast.Node { return lu.ReplaceNamedBindersAst(n, subs) }`

### File 3: `check/l2s.go`

- **Line 526** `modPass` closure — change transform type:
  `transform func(ast.Node) ast.Node`
- **Line 539** invars — match Python:
  `model.Invars[i] = transform(inv).(*ast.LabeledFormula)`
- **Line 544** asms — match Python:
  `model.Asms[i] = transform(asm).(*ast.LabeledFormula)`
- **Lines 559-569** prems — simplify to:
  ```go
  prems[i] = transform(lf).(*ast.LabeledFormula)
  ```
  Remove the `if e, ok := lf.Formula.(lg.Expr); ok` guard.
- **Line 893** `transformAction` — change transform param:
  `func(ast.Node) ast.Node`
  Line 906: `newArgs[i] = transform(a).(lg.Expr)` (type assert since ActionArgs returns `[]lg.Expr`)

### File 4: `check/ranking.go`

- **Line 367** `modPass` closure: same changes as l2s.go.
- **Line 380** invars: `transform(inv).(*ast.LabeledFormula)`
- **Line 385** asms: `transform(asm).(*ast.LabeledFormula)`
- **Lines 400-409** prems: `transform(lf).(*ast.LabeledFormula)`, remove guard.
- **Lines 412-415** postconds: `transform(pc).(*ast.LabeledFormula)`
- **Line 574** `ModelPass`: change transform param to `func(ast.Node) ast.Node`.
  Lines 580, 585: `transform(inv).(*ast.LabeledFormula)`, `transform(asm).(*ast.LabeledFormula)`.

### File 5: `check/ranking_test.go`

- **Line 507**: `func(n ast.Node) ast.Node { return n }`
- **Line 521**: `func(n ast.Node) ast.Node { ... return n }`
- **Line 586** (`TestRankingModPass_PropertyPremsTransformed`): The negate function and the simulated modPass loop need updating to pass whole LF to transform. Since the test simulates modPass inline, it should use the new pattern: `transform(lf).(*ast.LabeledFormula)`.

### File 6: `check/l2s_test.go`

- **Line 535**: `func(e ast.Node) ast.Node { return e }`
- **Line 543**: `func(n ast.Node) ast.Node { return &lg.Not{Body: n.(lg.Expr)} }`

## Files summary

| File | Changes |
|------|---------|
| `logicutil/logic_utils.go` | Widen 3 public + 1 private function signatures; generic fallback: `Args()`/`Clone()` |
| `check/l2s_shared.go` | `InstrumentationConfig` field, 2 `SharedStep*` params, callbacks, type assertions |
| `check/l2s.go` | `modPass` closure, invar/asm/prem loops, `transformAction` |
| `check/ranking.go` | `modPass` closure, all loops, `ModelPass` |
| `check/ranking_test.go` | Test lambda signatures |
| `check/l2s_test.go` | Test lambda signatures |

## Verification

1. `go build ./...`
2. `go test ./logicutil/...`
3. `go test ./check/...`
4. `go test ./...`
