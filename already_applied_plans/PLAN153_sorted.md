# Plan: Add Matching Proof Traces to Python + Sort Both Sides via Omap

**Created**: 2026-03-31 02:45, **Updated**: 2026-03-31 03:00

## Context

Golden test diverges at line 152539: Go emits `isolate.proofs n=33` but Python goes straight to `allSyms_pre_follow.sym`. Go has proof-related traces that Python lacks. Per project rule: never remove Go traces — add matching traces to Python instead.

The `allNames_from_proofs.name` per-name traces need sorted output on both sides because Python's `set()` doesn't preserve insertion order. Use `Omap[string, bool]` for `VocabNames` (instead of `InsMap`) — it maintains sorted order incrementally via red-black tree, so iteration is always sorted with no extra sort step.

## Fix

### 1. Change `VocabNames` type from `InsMap` to `Omap`

**File**: `ast/tactic.go`

```go
// VocabNames is a sorted set of symbol name strings,
// used as the container for Vocab methods.
// Uses Omap (red-black tree) for deterministic sorted iteration.
type VocabNames = iu.Omap[string, bool]

func NewVocabNames() *VocabNames {
    return iu.NewOmap[string, bool]()
}
```

`Omap` has the same `Set()`, `Get2()`, `Len()`, `All()` API as `InsMap` — all consumers work unchanged.

### 2. Go: Add `typeName` helper to isolate, change `%T` → `typeName()`

**File**: `isolate/isolate.go`

Add local `typeName` (copy from `compiler/compiler.go:27` — isolate can't import compiler):

```go
// typeName returns the struct name without package prefix.
// e.g. *ast.ProofTactic → "ProofTactic". Matches Python's type(x).__name__.
func typeName(v interface{}) string {
    t := reflect.TypeOf(v)
    for t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    return t.Name()
}
```

Change line 980: `xtracer.Trace("isolate.proof type=%T", pe.Proof)` → `xtracer.Trace("isolate.proof type=%s", typeName(pe.Proof))`

Need `"reflect"` import (check if already present).

Remove the `sort.Strings` step from the per-name trace — `Omap.All()` already iterates in sorted order, so the existing loop works as-is.

### 3. Python: Add matching proof traces

**File**: `/Users/jaten/pyivy/ivy/ivy/ivy_isolate.py` (between lines 1259-1263)

Replace:
```python
    all_names = set()
    for x in mod.proofs:
        x[1].vocab(all_names)
```

With:
```python
    all_names = set()
    if __debug__: xtracer.trace("isolate.proofs n=%d" % len(mod.proofs))
    for x in mod.proofs:
        if __debug__: xtracer.trace("isolate.proof type=%s" % type(x[1]).__name__)
        x[1].vocab(all_names)
    if __debug__:
        xtracer.trace("isolate.allNames_from_proofs n=%d" % len(all_names))
        for name in sorted(str(x) for x in all_names):
            xtracer.trace("isolate.allNames_from_proofs.name %s" % name)
```

Python `type(x[1]).__name__` matches Go `typeName()` (e.g. `"ProofTactic"`).
Python `sorted(str(x) for x in all_names)` matches Go `Omap.All()` sorted order.

## Files Modified

| File | Changes |
|------|---------|
| `ast/tactic.go` | Change `VocabNames` type alias from `InsMap` to `Omap`, `NewVocabNames` uses `NewOmap` |
| `isolate/isolate.go` | Add `typeName` helper, change `%T` → `typeName()`. Add `"reflect"` import if needed |
| `ivy_isolate.py` (lines 1259-1263) | Add 4 matching trace calls with sorted per-name output |

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...`
2. `cd ~/goivy && make test`
3. `cd ~/goivy && make golden` — divergence should advance past proof traces
