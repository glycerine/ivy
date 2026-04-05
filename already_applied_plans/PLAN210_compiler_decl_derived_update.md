# PLAN207: Fix DerivedUpdate symbol sort — TopSort instead of concrete sort

**Created:** 2026-04-05 ~02:30 UTC

## Context

Tracing revealed the true divergence at xtrace line 161761: `FrameDefConst` for `cmpl_live.retired` shows Go has `sort=alpha0 -> alpha1` (TopSort) while Python has `sort=lclock -> Boolean` (concrete). The symbol comes from `DerivedUpdate.GetUpdateAxioms` which appends `a.Symbol` to the Modified list.

## Root Cause

In `compiler/decl.go:644-684`, the `Derived()` function:
1. Creates `sym` with `TopFunctionSort` (generic alpha sorts) at line 644
2. Compiles the definition (which infers concrete sorts) at line 654
3. Re-adds concretely-sorted symbol to `Sig` at line 665
4. **BUT passes the original `sym` (still TopSort) to `NewDerivedUpdate` at line 684**

Python's `DerivedUpdate(df)` stores the COMPILED definition and extracts the symbol via `defn.args[0].rep`, which has concrete sorts.

## Fix

At line 683-684, extract the concretely-sorted symbol from the compiled definition instead of using the old `sym`:

```go
// Python: self.domain.updates.append(DerivedUpdate(df))
// Extract concretely-sorted symbol from compiled definition (matches Python's defn.args[0].rep)
derivedSym := sym // fallback
if def, ok := compiled.(*il.Definition); ok {
    if cnst, ok := def.Defines().(*lg.Const); ok {
        derivedSym = cnst
    }
}
mod.Updates = append(mod.Updates,
    module.NewDerivedUpdate(derivedSym, compiled))
```

Also fix line 675 and 679 which pass the old `sym` too:
```go
mod.SymbolOrder = append(mod.SymbolOrder, derivedSym)
mod.AllRelations = append(mod.AllRelations, derivedSym)
```

## Files Modified

- `compiler/decl.go` — `Derived()` function, lines 675-684

## Verification

```bash
go build ./...
go test ./compiler/... -count=1
make golden
```
