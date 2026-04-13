# Fix l2sGTriple.key() pointer-address dedup bug using NodeKey

**Created**: 2026-04-13 ~19:00 UTC

## Context

The Go `TestOrdLive` golden test diverges from Python at xtrace line 596106, in `l2s.SharedStep6 toG[1]`. Go produces an extra `Not(And(...))` entry while Python has `Not(Eq(...))` at that index. The root cause is that Go's `l2sGTriple.key()` uses `fmt.Sprintf("%v")` which prints the `Environ *string` as a pointer address instead of the string value. This prevents structurally identical triples from deduplicating in `cfg.L2sGs`, leaving 14 entries instead of the correct 6.

Python uses value-based equality via recstruct `__hash__`/`__eq__`. The Go equivalent is the `lg.NodeKey` system based on s-expressions (see `logic/sexp.go:15`).

## Changes

### 1. Export `VarsSexp` from `logic/sexp.go`

Rename the unexported `varsSexp` to `VarsSexp` so the `check` package can use it.

**File**: `logic/sexp.go` line 141
- Rename `func varsSexp(` → `func VarsSexp(`
- Update all 4 call sites within `sexp.go` (ForAll, Exists, Lambda, NamedBinder)

### 2. Change `key()` to return `lg.NodeKey` using Sexp convention

**File**: `check/l2s.go` lines 162-164

```go
// BEFORE:
func (t l2sGTriple) key() string {
    return fmt.Sprintf("%v|%v|%v", t.Vars, t.Body, t.Environ)
}

// AFTER:
func (t l2sGTriple) key() lg.NodeKey {
    env := "nil"
    if t.Environ != nil {
        env = *t.Environ
    }
    return lg.NodeKey("(l2sGTriple environ:" + env + " vars:" + lg.VarsSexp(t.Vars) + " body:" + string(t.Body.Sexp()) + ")")
}
```

### 3. Change map type from `map[string]` to `map[lg.NodeKey]`

**File**: `check/l2s_shared.go`

- Line 41: `L2sGs map[string]L2sGTriple` → `L2sGs map[lg.NodeKey]L2sGTriple`
- Line 87: `func sortL2sGTriples(m map[string]L2sGTriple)` → `func sortL2sGTriples(m map[lg.NodeKey]L2sGTriple)`
- Line 104: `cfg.L2sGs = make(map[string]L2sGTriple)` → `cfg.L2sGs = make(map[lg.NodeKey]L2sGTriple)`

### 4. Fix debug print to show environ value

**File**: `check/l2s.go` line 591

```go
// BEFORE:
fmt.Printf("l2s_g: %v %v %v\n", triple.Vars, triple.Body, triple.Environ)

// AFTER:
env := "<nil>"
if triple.Environ != nil {
    env = *triple.Environ
}
fmt.Printf("l2s_g: %v %v %s\n", triple.Vars, triple.Body, env)
```

## Files to modify

1. `logic/sexp.go` — export `VarsSexp`, update 4 internal call sites
2. `check/l2s.go` — change `key()` return type and implementation, fix debug print
3. `check/l2s_shared.go` — change map types from `map[string]` to `map[lg.NodeKey]`

## Verification

Run the golden test:
```
cd /Users/jaten/ivy/goivy/parser && go test -run TestOrdLive -v -count=1 -timeout 300s
```

The trace should now match Python through at least line 596106 (and ideally to completion).
