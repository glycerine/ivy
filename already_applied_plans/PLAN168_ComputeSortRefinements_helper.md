# Fix CanonizeTypes No-Op Causing Trace Divergence at Line 156456

**Created**: 2026-04-01 (current session)

## Context

At xtrace line 156456, Go exits `CreateIsolate` while Python continues doing work:
- Go: `XTRACE: check.CreateIsolate EXIT name=cf_live`
- Python: `XTRACE: ast.LF.clone PRESERVE origid=226 counter=1822`

Python's `canonize_types()` dynamically computes sort refinements via `il.sort_refinement()` and, when non-empty, resorts all labeled formulas (which triggers LF clone operations). Go calls `mod.CanonizeTypes(nil)`, making it a no-op. Additionally, Go's `resortLabeledFormulas` uses `NewLabeledFormulaFrom` instead of `Clone`, so even if sort refinements were provided, the PRESERVE xtrace wouldn't fire and the LF counter would drift.

## Changes

### 1. Add `ComputeSortRefinements` helper — `module/canonize.go`

Add a new exported function after the existing `CanonizeTypes` function (after line 96):

```go
// ComputeSortRefinements computes sort refinements from the signature.
// Corresponds to Python's il.sort_refinement() (ivy_logic.py:1471).
func ComputeSortRefinements(sig *il.Sig) []SortRefinement {
    if sig == nil {
        return nil
    }
    raw := il.GetSortRefinement(sig)
    if len(raw) == 0 {
        return nil
    }
    result := make([]SortRefinement, 0, len(raw))
    for _, s := range sig.Sorts {
        key := lg.SortKey(s)
        if newSort, ok := raw[key]; ok {
            result = append(result, SortRefinement{Old: s, New: newSort})
        }
    }
    return result
}
```

### 2. Fix `resortLabeledFormulas` to use Clone — `module/canonize.go`

Change lines 253-265 from using `NewLabeledFormulaFrom` to using `Clone`:

```go
func resortLabeledFormulas(lfs []*ast.LabeledFormula, rn map[lg.NodeKey]*SortRefinement) []*ast.LabeledFormula {
    if len(lfs) == 0 {
        return lfs
    }
    result := make([]*ast.LabeledFormula, len(lfs))
    for i, lf := range lfs {
        newFormula := ResortAST(lf.Formula.(lg.Expr), rn)
        // Python: ast.clone([ast.args[0], lu.resort_ast(ast.args[1], sort_refinement)])
        cloned := lf.Clone([]ast.Node{lf.Label, newFormula})
        result[i] = cloned.(*ast.LabeledFormula)
    }
    return result
}
```

Type safety: `lg.Expr` embeds `ast.Node` (logic/node.go:9), so `newFormula` is assignable to `ast.Node`. `LabeledFormula.Clone` returns `*LabeledFormula` wrapped as `ast.Node`.

This matches Python's behavior: decrements `LfCounter`, preserves ID, emits `ast.LF.clone PRESERVE` xtrace.

### 3. Pass computed refinements — `isolate/create.go`

Change line 270 from:
```go
mod.CanonizeTypes(nil)
```
to:
```go
mod.CanonizeTypes(module.ComputeSortRefinements(mod.Sig))
```

## Files to Modify

| File | Change |
|------|--------|
| `module/canonize.go` | Add `ComputeSortRefinements` function; fix `resortLabeledFormulas` to use Clone |
| `isolate/create.go` line 270 | Pass computed sort refinements instead of nil |

## Verification

```bash
cd ~/ivy/goivy && make golden
```

Check that line 156456 now matches between Go and Python (both should show `ast.LF.clone PRESERVE` traces from the canonize_types resort operation).

## Note: Ordering Difference (Not Part of This Fix)

The post-`isolate_component` operation order differs between Python and Go:
- **Python**: label → ext → check_compat → update_conjs → cone → fix_initializers → canonize_types → bracket → EXIT
- **Go**: fix_initializers → canonize_types → bracket → check_compat → update_conjs → label → ext → cone → EXIT

This may cause future divergences but is a separate concern. The canonize_types fix here should resolve the immediate divergence at line 156456.
