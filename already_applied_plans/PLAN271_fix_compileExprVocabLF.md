# Plan: Clone LabeledFormula after sort inference in compileExprVocabLF

Created: 2026-04-12 ~07:20 UTC
Updated: 2026-04-12 ~12:30 UTC

## Prior fixes (DONE)

1. TopSortAsDefault fix in `proof/goal.go:574-577` — moved divergence from 258992 to 258998.
2. Merkle state sharing (Module.SigMerkle) — moved divergence from 258998 to 259072.

## Context

Golden test `TestOrdLive` diverges at xtrace line 259072:

```
259069  go : ivylogic.SortAsDefault.Exit RESTORED ...    ← both match
259070  go : ast.LF.clone PRESERVE origid=599 counter=2250  ← both match
259071  go : compiler.Thing return type=LabeledFormula      ← both match
259072  go : ivylogic.SortAsDefault.Exit DELETED_S ...      ← DIVERGE
        py : ast.LF.clone PRESERVE origid=599 counter=2250
```

Go emits an **extra** `SortAsDefault.Exit DELETED_S` where Python emits a **second** `ast.LF.clone PRESERVE`. Go is missing a LabeledFormula clone that Python produces during sort inference.

## Root Cause

**Python** `compile_expr_vocab` (`ivy_proof.py:740-743`):
```python
with il.top_sort_as_default():
    expr = il.sort_infer_list([expr.compile()] + vocab.variables)[0]
```
1. `expr.compile()` → `_labeled_formula_cmpl` → `self.clone(...)` → **first** `ast.LF.clone PRESERVE` (line 259070)
2. `sort_infer_list([LabeledFormula] + vars)` → `concretize_terms` → `infer_sorts` sees `hasattr(t, 'clone')` → returns lambda `t.clone(inferred_args)` → **second** `ast.LF.clone PRESERVE` (line 259072)
3. Then `top_sort_as_default.__exit__` → `SortAsDefault.Exit DELETED_S`

**Go** `compileExprVocabLF` (`proof/goal.go:574-608`):
```go
tsDefault.Enter()
defer tsDefault.Exit()

compiled := c.CompileLF(lf)        // first ast.LF.clone PRESERVE (259070) ✓
// ...
if formula, ok := compiled.Formula.(lg.Expr); ok {
    // passes inner formula (lg.Expr), NOT the LabeledFormula
    inferred := il.SortInferList([formula, ...vars], nil, nil)
    compiled.Formula = inferred[0]  // MUTATES — no clone!
}
return compiled  // defer: tsDefault.Exit() → DELETED_S at 259072 ✗
```

Go extracts `compiled.Formula` (a `lg.Expr`) and passes that to `SortInferList`. Python passes the full `LabeledFormula` which has `clone`, so `concretize_terms` → `infer_sorts` clones it. Go never clones the LabeledFormula during sort inference → missing second trace.

Key fact: `lg.Expr` embeds `ast.Node` (`logic/node.go:9`), so logic expressions satisfy `ast.Node` and can be passed as `Clone` args.

## Fix

**File:** `proof/goal.go` — in `compileExprVocabLF`, replace the formula mutation with a LabeledFormula clone.

Change lines ~602-605 from:
```go
    if err == nil && len(inferred) > 0 {
        compiled.Formula = inferred[0]
    }
```
to:
```go
    if err == nil && len(inferred) > 0 {
        // Python: concretize_terms returns t.clone(inferred_args) for
        // AST nodes with clone. Match by cloning the LabeledFormula
        // with the sort-inferred formula, producing the matching
        // ast.LF.clone PRESERVE trace.
        cloneArgs := compiled.Args()  // []ast.Node{label, formula}
        cloneArgs[1] = inferred[0]    // replace formula with sort-inferred version
        compiled = compiled.Clone(cloneArgs).(*ast.LabeledFormula)
    }
```

This produces the **second** `ast.LF.clone PRESERVE` trace that matches Python, BEFORE the deferred `tsDefault.Exit()` runs.

### Files to modify

1. `proof/goal.go` — ~line 602-605 in `compileExprVocabLF`

## Verification

1. `go build ./...` — must compile
2. Run `TestOrdLive` — check that line 259072 now matches (`ast.LF.clone PRESERVE`) and divergence moves forward or test passes
