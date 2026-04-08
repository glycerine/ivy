# Plan: Port Python's `lookup_native` faithfully into Go

Created: 2026-04-07 ~01:45 UTC

## Context

Python has a single `lookup_native(thing, table, kind)` function (ivy_solver.py:311-347) that handles native interpretation lookups for sorts, relations, AND functions. It takes a `table` parameter — one of `sorts()`, `relations()`, or `functions()` — and a `kind` string.

Go currently splits this into two separate mechanisms:
- **For sorts**: `SortLookup` callback on Translator, with `lookup_native + sorts()` logic inlined in a closure at `solver/solver.go:125-181`
- **For functions/relations**: `LookupNative` method on Solver at `solver/z3convert.go:549-612`, wired via `NativeLookup` callback

This violates the mechanical port rules (one Python function = one Go function) and the existing `LookupNative` has bugs:
- Returns `nil` for EnumeratedSort/RangeSort interps (Python returns `z3name.to_z3()`)
- Missing `table` parameter — uses `isRelation` flag instead
- Different structure from Python

## Changes

### 1. Create `Sorts()` method on Solver

**File**: `solver/z3convert.go`

Port of Python's `sorts()` function (ivy_solver.py:120-145). Extract the logic currently inlined in the `SortLookup` closure at `solver/solver.go:135-166`.

```go
// Sorts resolves a sort interpretation name to a Z3 sort.
// Corresponds to Python sorts() (ivy_solver.py:120).
func (s *Solver) Sorts(name string) any {
    xtracer.Trace("ivy_solver.py:121 sorts() ENTER name=%s", name)
    ctx := s.tr.Ctx
    switch name {
    case "nat", "int":
        return ctx.IntSort()
    case "real":
        return ctx.RealSort()
    case "strlit":
        return ctx.StringSort()
    }
    base, params, ok := ParseIntParams(name)
    if ok && len(params) > 0 {
        switch base {
        case "bv", "strbv", "intbv":
            return ctx.BvSort(params[0])
        }
    }
    if dom, rng, ok2 := z3bridge.ParseArraySortName(name); ok2 {
        domSort, err1 := s.tr.TranslateSort(&lg.UninterpretedSort{Name: dom})
        rngSort, err2 := s.tr.TranslateSort(&lg.UninterpretedSort{Name: rng})
        if err1 == nil && err2 == nil {
            return ctx.ArraySort(domSort, rngSort)
        }
    }
    return nil
}
```

### 2. Create `Relations()` and `Functions()` methods on Solver

**File**: `solver/z3convert.go`

Ports of Python's `relations()` (ivy_solver.py:171) and `functions()` (ivy_solver.py:227).

```go
// Relations resolves a relation name to a native Z3 function.
// Corresponds to Python relations() (ivy_solver.py:171).
func (s *Solver) Relations(name string) any {
    xtracer.Trace("ivy_solver.py:172 relations() ENTER name=%s", name)
    return s.lookupBuiltinRelation(name)
}

// Functions resolves a function name to a native Z3 function.
// Corresponds to Python functions() (ivy_solver.py:227).
func (s *Solver) Functions(name string) any {
    xtracer.Trace("ivy_solver.py:228 functions() ENTER name=%s", name)
    return s.lookupBuiltinFunc(name, false)
}
```

### 3. Rewrite `LookupNative` to match Python

**File**: `solver/z3convert.go`

Replace `LookupNative(sym *lg.Const, isRelation bool) NativeFunc` with a faithful port matching Python's `lookup_native(thing, table, kind)`:

```go
// LookupNative resolves the native Z3 interpretation for an Ivy symbol.
// Corresponds to Python lookup_native(thing, table, kind) (ivy_solver.py:311).
func (s *Solver) LookupNative(thing *lg.Const, table func(string) any, kind string) any {
    xtracer.Trace("ivy_solver.py:312 lookup_native() ENTER name=%s kind=%s", thing.Name, kind)
    if s.sig == nil { return nil }

    z3name, hasInterp := s.sig.Interp[thing.Name]
    if !hasInterp {
        if strings.HasPrefix(thing.Name, "bfe[") { return s.bfeToZ3(thing) }
        if thing.Name == "arrcst" {
            // Python: sort = thing.sort.rng; if sort.name in sig.interp: ...
            if fs, ok := thing.CSort.(*lg.FunctionSort); ok {
                rngSort := fs.Range()
                rngName := lg.SortName(rngSort)
                if _, inInterp := s.sig.Interp[rngName]; inInterp {
                    ctx := s.tr.Ctx
                    z3arrSort, err := s.tr.TranslateSort(rngSort)
                    if err == nil && z3arrSort.Kind() == z3bridge.SortArray {
                        domSort := z3arrSort.ArrayDomain()
                        return NativeFunc(func(args ...z3bridge.Expr) z3bridge.Expr {
                            if len(args) == 1 { return ctx.ConstArray(domSort, args[0]) }
                            return ctx.BoolVal(false)
                        })
                    }
                }
            }
        }
        if isPolymorphicOp(thing.Name) {
            return s.lookupPolymorphicNative(thing, table, kind)
        }
        return nil
    }

    // Python line 342-343
    switch v := z3name.(type) {
    case *lg.EnumeratedSort:
        zs, err := s.tr.TranslateSort(v)
        if err != nil { return nil }
        return zs
    case *lg.RangeSort:
        zs, err := s.tr.TranslateSort(v)
        if err != nil { return nil }
        return zs
    case string:
        // Python line 344: z3val = table(z3name)
        return table(v)
    }
    return nil
}
```

### 4. Refactor `lookupPolymorphicNative` to take `table` and `kind`

**File**: `solver/z3convert.go`

Update signature from `(sym, isRelation)` to `(sym, table, kind)` to match Python's fall-through to `table(thing.name)` at line 337:

```go
func (s *Solver) lookupPolymorphicNative(sym *lg.Const, table func(string) any, kind string) any {
    // ... same domain sort + sig.interp checks ...
    // ... nat subtraction, RangeSort clamped arithmetic ...
    // Python line 337: z3val = table(thing.name)
    return table(sym.Name)
}
```

The internal logic stays the same; only the fall-through at the end changes from `lookupBuiltinRelation`/`lookupBuiltinFunc` dispatch to `table(sym.Name)`.

### 5. Replace dual callbacks on Translator with unified `LookupNative`

**File**: `z3bridge/translate.go`

Remove `NativeLookupFunc` type (line 16) and `SortLookupFunc` type (line 30).
Remove `NativeLookup` field (line 54) and `SortLookup` field (line 57).
Add:

```go
// LookupNativeFunc is a callback matching Python lookup_native(thing, table, kind).
// Returns any: z3bridge.Sort for sort lookups, func(args ...Expr) Expr for
// function/relation lookups, or nil.
type LookupNativeFunc func(name string, sort logic.Sort, kind string) any

type Translator struct {
    ...
    LookupNative LookupNativeFunc  // replaces NativeLookup + SortLookup
    ...
}
```

### 6. Update `wireNativeLookup`

**File**: `solver/solver.go`

Replace the two separate closures (lines 113-122 for NativeLookup, lines 125-181 for SortLookup) with one:

```go
s.tr.LookupNative = func(name string, sort lg.Sort, kind string) any {
    sym := lg.NewConst(name, sort)
    var table func(string) any
    switch kind {
    case "sort":
        table = s.Sorts
    case "relation":
        table = s.Relations
    case "function":
        table = s.Functions
    }
    return s.LookupNative(sym, table, kind)
}
```

### 7. Update callers in `z3bridge/translate.go`

**TranslateSort** (line 137-146) — `SortLookup` → `LookupNative`:
```go
// Python: s = lookup_native(us, sorts, "sort")
if t.LookupNative != nil {
    result := t.LookupNative(st.Name, s, "sort")
    if zs, ok := result.(Sort); ok {
        t.sorts[key] = zs
        t.sortsInv[zs.GetId()] = s
        return zs, nil
    }
}
```

Note: the `lookup_native` xtrace at line 138 should be REMOVED since it now lives inside `LookupNative` itself.

**atomToZ3** (line 478-483) — `NativeLookup` → `LookupNative`:
```go
if t.LookupNative != nil {
    if result := t.LookupNative(c.Name, c.CSort, "relation"); result != nil {
        if nativeFn, ok := result.(func(args ...Expr) Expr); ok {
            t.preds[predKey] = nativeFn
            return t.applyPred(nativeFn, app.Terms)
        }
    }
}
```

**translateCore non-Boolean Apply** (line 278-290) — `NativeLookup` → `LookupNative`:
```go
if t.LookupNative != nil {
    if result := t.LookupNative(c.Name, c.CSort, "function"); result != nil {
        if nativeFn, ok := result.(func(args ...Expr) Expr); ok {
            args := make([]Expr, len(node.Terms))
            for i, term := range node.Terms {
                a, err := t.TermToZ3(term)
                if err != nil { return Expr{}, err }
                args[i] = a
            }
            return nativeFn(args...), nil
        }
    }
}
```

**translateVariable** (line 582-595) — `SortLookup` → `LookupNative`:
```go
if t.LookupNative != nil {
    xtracer.Trace("ivy_solver.py:312 lookup_native() ENTER name=%s kind=sort", sortName)
    if result := t.LookupNative(sortName, sort, "sort"); result != nil {
        if zsVal, ok := result.(Sort); ok {
            zs = &zsVal
            sortKey := sort.Sexp()
            if _, ok := t.sorts[sortKey]; !ok {
                t.sorts[sortKey] = *zs
                t.sortsInv[zs.GetId()] = sort
            }
        }
    }
}
```

Wait — the `translateVariable` path emits its own `lookup_native` trace (line 585) BEFORE calling the callback. But `LookupNative` also emits the trace. This would double-emit. Fix: remove the trace from translateVariable since `LookupNative` handles it internally.

### 8. Update `BinaryInterpolant`

**File**: `solver/z3convert.go` (line 243)

```go
// Old:
itpTr.NativeLookup = s.tr.NativeLookup

// New:
itpTr.LookupNative = s.tr.LookupNative
```

Also should add `itpTr.SolverName = s.tr.SolverName` (already there at line 244).

### 9. Update `CheckNativeCompatSym`

**File**: `solver/compat.go` (line 69)

```go
// Old:
nf := s.LookupNative(sym, isRelation)

// New:
var table func(string) any
var kind string
if isRelation {
    table = s.Relations
    kind = "relation"
} else {
    table = s.Functions
    kind = "function"
}
result := s.LookupNative(sym, table, kind)
nf, _ := result.(NativeFunc)
```

### 10. Update tests

**`z3bridge/translate2_test.go`** (lines 97, 145): Change `tr.SortLookup = func(name string) *Sort {` to `tr.LookupNative = func(name string, sort logic.Sort, kind string) any {` — return `Sort` value (not pointer) for sort lookups, `nil` otherwise.

**`z3bridge/translate2_fuzz_test.go`** (lines 146, 250): Same callback update.

**`solver/solver_test.go`**: Tests calling `lookupBuiltinFunc` and `lookupBuiltinRelation` remain unchanged (those helpers survive as internal methods). `TestLookupNativeArrselDispatch` tests `lookupBuiltinFunc` directly, unchanged.

### 11. Remove redundant array sort check from TranslateSort

**File**: `z3bridge/translate.go` (lines 148-163)

The array sort pattern check on the raw sort name (`ParseArraySortName(st.Name)`) is done directly in TranslateSort AFTER `SortLookup` fails. Python's `uninterpretedsort()` does NOT have this check — it relies entirely on `lookup_native → sorts()`. With the unified `LookupNative` now handling the `sorts()` path (which includes array sort parsing), this redundant check should be removed. If the sort name is `arr[dom][rng]` and it's in `sig.interp`, `LookupNative → Sorts()` handles it. If it's NOT in `sig.interp`, Python creates an uninterpreted sort with that name (no array magic).

## Files Modified

| File | Changes |
|------|---------|
| `solver/z3convert.go` | Rewrite `LookupNative`, add `Sorts`/`Relations`/`Functions`, update `lookupPolymorphicNative` signature, fix EnumeratedSort/RangeSort, update `BinaryInterpolant` |
| `z3bridge/translate.go` | Replace `NativeLookupFunc`+`SortLookupFunc` with `LookupNativeFunc`, replace `NativeLookup`+`SortLookup` fields with `LookupNative`, update 4 call sites, remove redundant array check |
| `solver/solver.go` | Rewrite `wireNativeLookup` with single unified closure |
| `solver/compat.go` | Update `CheckNativeCompatSym` caller |
| `z3bridge/translate2_test.go` | Update 2 test callbacks |
| `z3bridge/translate2_fuzz_test.go` | Update 2 test callbacks |

## Verification

```bash
cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go build ./...
cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test ./z3bridge/... ./solver/...
cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy/parser && go test -v -run TestOrdLive -timeout 600s
```
