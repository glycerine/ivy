# Plan: Audit and Fix TranslateSort to Faithfully Port Python's to_z3() Methods

**Created:** 2026-04-08 ~00:45 UTC

## Context

Go's `TranslateSort` (z3bridge/translate.go:124) is the centralized equivalent of Python's polymorphic `sort.to_z3()` methods (ivy_solver.py:257-295). Multiple previous sessions have hacked traces into TranslateSort without ensuring the Go code faithfully matches the Python logic. This audit compares every Python `to_z3()` method against the corresponding Go case, identifies all divergences, and fixes them.

**Source of truth:** Python ivy_solver.py lines 257-295 (sort to_z3 methods) and lines 311-347 (lookup_native).

---

## Audit Results

### 1. BooleanSort — CORRECT ✅

| | Python (line 294) | Go (line 126-128) |
|---|---|---|
| Code | `lambda self: z3.BoolSort()` | `return t.Ctx.BoolSort(), nil` |
| Trace | None | None |
| Caching | None | None |
| sortsInv | None | None |

No issues.

### 2. UninterpretedSort — 1 issue

**Python** (lines 257-266):
```python
def uninterpretedsort(us):
    trace("...258 uninterpretedsort() ENTER name=%s" % us.rep)
    s = z3_sorts.get(us.rep, None)       # cache key = us.name (plain string)
    if s is not None: return s
    s = lookup_native(us, sorts, "sort") # ALWAYS called; emits lookup_native trace
    if s == None:
        s = z3.DeclareSort(us.rep)
    z3_sorts[us.rep] = s                 # ALWAYS stored
    z3_sorts_inv[get_id(s)] = us         # ALWAYS stored
    return s
```

**Go** (lines 130-168):
- Trace: ✅ matches
- Cache check: ✅ functional equivalent (uses Sexp() key, Python uses name string — both unique)
- lookup_native call: ✅ calls SortLookup
- DeclareSort fallback: ✅ `t.Ctx.UninterpretedSort(st.Name)`
- z3_sorts cache store: ✅
- z3_sorts_inv store: ✅ via `zs.GetId()`
- Array sort handling: in wrong location (see Issue A below)

**Issue A — Array sort resolution path differs:** Python resolves array sorts inside `sorts()` (called from `lookup_native`), triggered when `sig.interp["mymap"] = "arr[int][bool]"` → `lookup_native` finds the string → calls `sorts("arr[int][bool]")` → `parse_array_theory` → ArraySort. Go's `ParseArraySortName` checks the sort NAME directly (line 150), not the interpreted value from sig.Interp. If a sort named "mymap" is interpreted as "arr[int][bool]", Python resolves it but Go would miss it.

**Fix A:** Move array sort resolution into `SortLookup` (solver/solver.go), since that's where sig.Interp is checked — matching where Python's `sorts()` function handles it.

### 3. FunctionSort — Structural divergence, functionally OK

**Python** (lines 268-273): Returns a **list** of Z3 sorts (domain + range). Callers unpack with `z3.Function(name, *sig)`.

**Go** (lines 182-184): Returns **error**. Go uses `makeFuncDecl` instead, which combines sort translation and FuncDecl creation in one step.

This is a structural difference, not a bug. Both produce the same Z3 Function declarations. The trace output matches (both emit `functionsort() ENTER`, then domain sort traces, then range sort trace). **No fix needed.**

### 4. EnumeratedSort — 2 issues

**Python** (lines 275-284):
```python
def enumeratedsort(es):
    trace("...276 enumeratedsort() ENTER name=%s" % es.name)
    s = z3_sorts.get(es.rep, None)
    if s is not None: return s
    res, consts = z3.EnumSort(es.name, es.extension)
    z3_sorts[es.rep] = res
    for c in consts:
        z3_constants[str(c)] = c         # key = Z3 string repr of constant
    # NOTE: NO z3_sorts_inv entry
    return res
```

**Go** (lines 186-202):
- Trace: ✅
- Cache: ✅
- EnumSort call: ✅
- z3_sorts store: ✅

**Issue B — Extra sortsInv entry:** Go stores `t.sortsInv[zs.GetId()] = s` (line 196). **Python does NOT** store EnumeratedSort in `z3_sorts_inv`. This means `SortFromZ3` would find enumerated sorts in Go but not in Python, potentially causing different behavior in Herbrand model processing.

**Fix B:** Remove line 196 (`t.sortsInv[zs.GetId()] = s`) from the EnumeratedSort case.

**Issue C — SortLookup missing EnumeratedSort case:** Python's `lookup_native` (line 342-343) handles the case where `sig.interp[name]` maps to an EnumeratedSort or RangeSort object: `if isinstance(z3name, (EnumeratedSort, RangeSort)): return z3name.to_z3()`. Go's `SortLookup` (solver.go:134-163) has a case for `*lg.RangeSort` but **no case for `*lg.EnumeratedSort`**. If `sig.Interp[sortName]` maps to an EnumeratedSort, Go returns nil, potentially creating a wrong uninterpreted sort.

**Fix C:** Add `*lg.EnumeratedSort` case to SortLookup in solver.go.

### 5. RangeSort — 2 issues

**Python** (line 293):
```python
ivy_logic.RangeSort.to_z3 = lambda self: z3.IntSort()
```
No trace. No caching. No sortsInv. Just returns IntSort.

**Go** (lines 204-210):
```go
case *logic.RangeSort:
    xtracer.Trace("ivy_solver.py:258 uninterpretedsort() ENTER name=%s", st.Name)  // WRONG
    zs := t.Ctx.IntSort()
    t.sortsInv[zs.GetId()] = s  // WRONG
    return zs, nil
```

**Issue D — WRONG TRACE:** Go emits `uninterpretedsort() ENTER` for RangeSort. Python's RangeSort.to_z3 is a plain lambda with **no trace at all**. This will cause trace divergence any time a RangeSort flows through TranslateSort.

**Fix D:** Remove the xtracer.Trace line from the RangeSort case.

**Issue E — Extra sortsInv entry:** Go stores `t.sortsInv[zs.GetId()] = s`. Python's RangeSort lambda does NOT. Same issue as EnumeratedSort.

**Fix E:** Remove `t.sortsInv[zs.GetId()] = s` from the RangeSort case.

### 6. TopSort — No Python equivalent

Go (lines 170-180) treats TopSort as uninterpreted. Python has **no** `TopSort.to_z3` assignment. If Python ever tries to call `.to_z3()` on a TopSort, it raises AttributeError. Go's treatment as uninterpreted is a reasonable defensive fallback. **No fix needed** — unlikely to affect traces and would crash Python if reached.

---

## Changes

### File 1: `z3bridge/translate.go`

#### Fix D — Remove wrong RangeSort trace (line 207)

```go
// BEFORE:
case *logic.RangeSort:
    xtracer.Trace("ivy_solver.py:258 uninterpretedsort() ENTER name=%s", st.Name)
    zs := t.Ctx.IntSort()
    t.sortsInv[zs.GetId()] = s
    return zs, nil

// AFTER:
case *logic.RangeSort:
    // Python: lambda self: z3.IntSort() — no trace, no caching, no sortsInv
    return t.Ctx.IntSort(), nil
```

#### Fix B — Remove extra sortsInv from EnumeratedSort (line 196)

```go
// BEFORE:
zs, constExprs := t.Ctx.EnumSort(st.Name, st.Extension)
t.sorts[key] = zs
t.sortsInv[zs.GetId()] = s   // DELETE THIS LINE

// AFTER:
zs, constExprs := t.Ctx.EnumSort(st.Name, st.Extension)
t.sorts[key] = zs
// Python enumeratedsort() does NOT store in z3_sorts_inv
```

### File 2: `solver/solver.go`

#### Fix C — Add EnumeratedSort handling to SortLookup (after line 161)

```go
case *lg.EnumeratedSort:
    // Python lookup_native line 342-343:
    // if isinstance(z3name, (EnumeratedSort, RangeSort)): return z3name.to_z3()
    zs, err := s.tr.TranslateSort(v)
    if err != nil {
        return nil
    }
    return &zs
```

#### Fix A — Add array sort resolution to SortLookup (in the string case default)

In the `case string:` / `default:` branch (after bv/strbv/intbv handling), add array sort resolution:

```go
default:
    // Check for bv[N], strbv[N], intbv[N]
    base, params, ok := ParseIntParams(v)
    if ok && len(params) > 0 {
        switch base {
        case "bv", "strbv", "intbv":
            zs := ctx.BvSort(params[0])
            return &zs
        }
    }
    // Python sorts() lines 125-130: array sort names "arr[dom][rng]"
    if dom, rng, ok2 := z3bridge.ParseArraySortName(v); ok2 {
        domSort, err1 := s.tr.TranslateSort(&lg.UninterpretedSort{Name: dom})
        rngSort, err2 := s.tr.TranslateSort(&lg.UninterpretedSort{Name: rng})
        if err1 == nil && err2 == nil {
            zs := ctx.ArraySort(domSort, rngSort)
            return &zs
        }
    }
```

---

## Verification

1. Build: `cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go build ./...`
2. Run solver tests: `go test ./solver/ -run TestTranslate -count=1 -v`
3. Run sort-specific tests: `go test ./z3bridge/ -run TestTranslateSort -count=1 -v`
4. Run SortLookup tests: `go test ./solver/ -run TestSortLookup -count=1 -v`
5. Run golden update: `cd ~/ivy/goivy && make golden`
6. Check log.red: RangeSort trace divergences should be gone
