# Plan: Fix allSyms2 Divergence at Golden Test Line 155136

**Created: 2026-03-30 ~17:00**

## Context

The golden test (`make golden` / `TestOrdLive`) compares Go and Python execution traces line-by-line. They match through 155135 lines (the `isolate.oldActions` phase), then diverge at line 155136:

```
155136  go : XTRACE: actions.referencesRec ENTER type=sequence n_before=61
        py : XTRACE: isolate.allSyms2.sym 0:index
```

Go enters the action-specific `referencesRec` code path while Python has already moved to emitting `allSyms2` symbol traces. The divergence is in the `allSyms2` construction phase of `isolate_component`.

## Root Cause

Three bugs in `isolate/isolate.go` lines 1267-1302 (the `allSyms2` block):

### Bug 1 (Primary — causes the trace divergence)

**Line 1284**: Go calls `actions.GetReferencesInto(act, allSyms2, mod.DestructorSorts)` — the action-specific `referencesRec` code path with specialized per-type logic and its own traces.

**Python line 1360**: Adds actions to a flat `asts` list and calls `lu.used_symbols_asts(asts)` — the generic AST symbol walker (`ivy_logic_utils.symbols_ast`), which simply walks `.args` recursively yielding `Const` symbols. No `get_references()` is called.

Note: The first `allSyms` phase (Go line ~967) correctly uses `GetReferencesInto` because Python's first phase also uses `action.get_references()`. Only the second phase (`allSyms2`) differs — Python switches to the generic walker.

### Bug 2 (SchemaBody filter missing)

**Go lines 1268-1275**: No filter for `SchemaBody` formulas.
**Python line 1359**: `if not isinstance(y.formula, ivy_ast.SchemaBody)` — skips SchemaBody.

### Bug 3 (Natives collection broken)

**Go lines 1291-1294**: `nat.(*ast.LabeledFormula)` — wrong type. Natives are `*ast.NativeDef`, not `*ast.LabeledFormula`. Type assertion always fails, collecting nothing.
**Python line 1366-1367**: `asts.extend(tmp.args[2:])` — walks NativeDef args from index 2.
The first allSyms phase (Go line ~944) already has the correct pattern.

## Fix

All changes in one file: `/Users/jaten/go/src/github.com/glycerine/goivy/isolate/isolate.go`

### Fix 1: Add SchemaBody filter (lines 1271-1274)

Add before the `collectUsedSymbolNames` call inside the formula loop:

```go
if _, isSchema := lf.Formula.(*ast.SchemaBody); isSchema {
    continue
}
```

### Fix 2: Replace GetReferencesInto with generic walker (line 1284)

Replace:
```go
actions.GetReferencesInto(act, allSyms2, mod.DestructorSorts)
```
With:
```go
collectUsedSymbolNames(act, allSyms2)
```

`module.Action` embeds `lg.Expr`, so `act` passes directly to `collectUsedSymbolNames(node lg.Expr, ...)`. The walker uses `Children()` which returns `ActionArgs()` for actions — matching Python's `.args` traversal.

### Fix 3: Fix natives collection (lines 1291-1294)

Replace:
```go
for _, nat := range mod.Natives {
    if lf, ok := nat.(*ast.LabeledFormula); ok && lf.Formula != nil {
        collectUsedSymbolNames(lf.Formula.(lg.Expr), allSyms2)
    }
}
```
With:
```go
for _, nat := range mod.Natives {
    args := nat.Args()
    for i := 2; i < len(args); i++ {
        if expr, ok := args[i].(lg.Expr); ok {
            collectUsedSymbolNames(expr, allSyms2)
        }
    }
}
```

## Key Files

- `isolate/isolate.go` lines 1267-1302 — all three fixes
- `isolate/helpers.go` lines 404-417 — `collectUsedSymbolNames` (the correct generic walker, already exists)
- Python reference: `ivy_isolate.py` lines 1357-1377

## Verification

```bash
cd ~/goivy && make golden
cd ~/goivy && make test
```

The golden test should now match past line 155136, with `isolate.allSyms2.sym` traces appearing where `referencesRec` traces previously diverged.
