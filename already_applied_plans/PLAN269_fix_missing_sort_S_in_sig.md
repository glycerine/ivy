# Plan: Fix missing sort "S" in Sig at SigCheck@SortifyWithInference

Created: 2026-04-12 ~06:30 UTC

## Context

At golden test xtrace line 255406, `compiler.SigCheck@SortifyWithInference` diverges. Python's Sig contains sort `S` in its sorts list; Go's does not:

```
Go:  sorts:[addr_type bool index lclock loc_type ...]
Py:  sorts:[S addr_type bool index lclock loc_type ...]
```

Sort "S" is the default sort for unsorted constants in Ivy version <= 1.2. Python's `default_sort()` lazily creates `UninterpretedSort("S")` and permanently adds it to `sig.sorts`. Go has `GetDefaultSort()` that does the same thing, but it is **dead code** -- never called from anywhere.

## Root Cause

Two Go code paths fail to call `GetDefaultSort()`:

1. **`CompileConst()`** (compiler/compiler.go:1244-1251) -- when a constant has no explicit sort, Go falls back to `lg.TopS` instead of calling `GetDefaultSort(sig)`. Python calls `ivy_logic.default_sort()` here (ivy_compiler.py:582).

2. **`FindSort()`** (ivylogic/sig.go:300-304) -- when looking up sort name "S" and it's not in the map, Go returns an error if `DefaultSort` is nil. Python calls `default_sort()` here (ivy_logic.py:330-332).

Because "S" is never permanently added to `sig.Sorts`, the `SortAsDefault` context manager (which temporarily sets/restores "S") always deletes it on `Exit()` (`hadOld=false`). In Python, since "S" was permanently added earlier, the context manager restores it instead.

## Changes

### 1. `ivylogic/ivylogic.go` -- Add version check to `GetDefaultSort`

Modify `GetDefaultSort` (lines 403-415) to return an error for version > 1.2, matching Python:

```go
func GetDefaultSort(sig *Sig) (lg.Sort, error) {
    if sig.DefaultSort != nil {
        return sig.DefaultSort, nil
    }
    if sig.IuCfg != nil && !iu.VersionLE(sig.IuCfg.LanguageVersion, "1.2") {
        return nil, &lg.IvyError{Msg: "unspecified type"}
    }
    ds := &lg.UninterpretedSort{Name: "S"}
    sig.Sorts["S"] = ds
    sig.DefaultSort = ds
    return ds, nil
}
```

### 2. `compiler/compiler.go` -- Fix `CompileConst` fallback (lines 1244-1251)

Replace:
```go
if rng == nil {
    if sig.DefaultSort != nil {
        rng = sig.DefaultSort
    } else {
        rng = lg.TopS
    }
}
```

With:
```go
if rng == nil {
    var err error
    rng, err = il.GetDefaultSort(sig)
    if err != nil {
        return nil, err
    }
}
```

### 3. `ivylogic/sig.go` -- Fix `FindSort` "S" handling (lines 300-304)

Replace:
```go
if name == "S" {
    if s.DefaultSort != nil {
        return s.DefaultSort, nil
    }
    return nil, &lg.IvyError{Msg: "unspecified type"}
}
```

With:
```go
if name == "S" {
    return GetDefaultSort(s)
}
```

## Files Modified

- `/Users/jaten/ivy/goivy/ivylogic/ivylogic.go` -- `GetDefaultSort` signature change + version check
- `/Users/jaten/ivy/goivy/compiler/compiler.go` -- `CompileConst` fallback
- `/Users/jaten/ivy/goivy/ivylogic/sig.go` -- `FindSort` "S" branch

## Why This Works

Once `GetDefaultSort()` is actually called, it:
1. Creates `UninterpretedSort("S")` and adds it to `sig.Sorts["S"]` permanently
2. Caches it in `sig.DefaultSort`
3. Subsequent `SortAsDefault.Enter()` sees `hadOld=true`, so `Exit()` restores "S" rather than deleting it
4. `Canon()` iterates `sig.Sorts` and "S" appears in the sorted output

This matches Python's flow exactly.

## Verification

1. `go build ./...` -- confirms compilation
2. Run the TestOrdLive golden test to check line 255406 matches
3. Verify no other tests regress: `go test ./ivylogic/... ./compiler/...`
