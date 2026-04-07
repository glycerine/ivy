# PLAN209: Fix Apply aSort loss during AST renaming

**Created:** 2026-04-06 ~23:50 UTC

## Context

At xtrace line 179848, Go and Python diverge:
```
go : defToConstraint lhsSort=<nil>  resultType=Eq
py : defToConstraint lhsSort=Boolean resultType=Iff
```

Both sides reach `ToOpenFormula` with matching counts (nFmlas=286 nDefs=27). The 27 definitions match through ComposeUpdates and Hide. But when `defToConstraint` converts defs to constraints, Go's Definition LHS has `NodeSort()=nil` instead of `Boolean`, so it produces `Eq` instead of `Iff`.

**Root cause**: When `Hide` calls `ExistQuantClauses` → `RenameClauses` → `RenameAST` → `renameASTRec`, the Apply case at `astutil.go:180` creates `&lg.Apply{Func: newFunc, Terms: newTerms}` — a struct literal that does NOT set the `aSort` field. Since `aSort` is unexported (lowercase), code outside the `logic` package can't access it. The zero value is `nil`, so `NodeSort()` returns nil.

The same bug exists in two other AST reconstruction paths:
- `ivylogic/util.go:20` — `CloneNode` (called from `renameASTRec` line 205 for fallthrough nodes)
- `actions/transrel.go:1332` — `renameNode` (used in compose_updates renaming path)

## Fix: Use `lg.NewApply` instead of struct literals

`NewApply` properly computes `aSort` from the function's sort (FunctionSort.Range() → the correct result sort). Renaming preserves sorts on the function Const (verified in `renameASTRec` Const case, lines 152-163), so `NewApply` will always succeed after renaming.

### Change 1: `module/astutil.go:180`
```go
// OLD:
return &lg.Apply{Func: newFunc, Terms: newTerms}

// NEW:
app, err := lg.NewApply(newFunc, newTerms...)
if err != nil {
    panic(fmt.Sprintf("renameASTRec: NewApply failed after rename: %v", err))
}
return app
```

### Change 2: `ivylogic/util.go:20`
```go
// OLD:
return &lg.Apply{Func: t.Func, Terms: args, }

// NEW:
app, err := lg.NewApply(t.Func, args...)
if err != nil {
    panic(fmt.Sprintf("CloneNode: NewApply failed: %v", err))
}
return app
```

### Change 3: `actions/transrel.go:1332`
```go
// OLD:
return &lg.Apply{Func: newFunc, Terms: newTerms}

// NEW:
app, err := lg.NewApply(newFunc, newTerms...)
if err != nil {
    panic(fmt.Sprintf("renameNode: NewApply failed after rename: %v", err))
}
return app
```

All three files already import `fmt` and `lg`.

## Files

- `module/astutil.go:180` — PRIMARY (Hide → ExistQuantClauses → RenameClauses path)
- `ivylogic/util.go:20` — CloneNode fallthrough path
- `actions/transrel.go:1332` — renameNode in compose_updates path

## Note: Systemic issue

There are ~80 other `&lg.Apply{...}` struct literals across the codebase that also don't set `aSort`. Many are in tests (acceptable) or in code that doesn't depend on correct sorts (solver internals). The three above are on the critical path causing the current divergence. Others can be fixed incrementally as they surface.

## Verification

```bash
cd ~/ivy/goivy && go build ./...
make golden
```

Expected: The divergence at line 179848 resolves. `defToConstraint` will now produce `Iff` (matching Python) because the renamed Definition LHS retains its Boolean sort through the renaming pipeline.
