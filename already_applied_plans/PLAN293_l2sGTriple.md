# Fix l2sGTriple.key() pointer-address dedup bug using NodeKey

**Created**: 2026-04-13 ~19:00 UTC

## Context

The Go `TestOrdLive` golden test diverges from Python at xtrace line 596106, in `l2s.SharedStep6 toG[1]`. Go produces an extra `Not(And(...))` entry where Python has `Not(Eq(...))`. The root cause is that `l2sGTriple.key()` uses `fmt.Sprintf("%v")` on the `Environ *string` field — Go's `%v` prints the pointer address (e.g., `0x3bfce4b33730`), not the string value. Structurally identical triples from different AST clones get different `Environ` pointers, producing different keys, so `cfg.L2sGs` keeps 14 entries instead of 6.

Python uses value-based structural equality via recstruct `__hash__`/`__eq__`. The Go equivalent is the `lg.NodeKey` system based on s-expressions (`logic/sexp.go:15`).

## Changes

### Change 1: Export `VarsSexp` — file `logic/sexp.go`

The `check` package needs `varsSexp` to build s-expression keys. Export it.

**Line 141** — rename function:
```go
// BEFORE:
func varsSexp(vars []*Variable) string {

// AFTER:
func VarsSexp(vars []*Variable) string {
```

**Lines 151, 155, 159, 167** — update all 4 internal call sites from `varsSexp(` to `VarsSexp(`:
- Line 151: `ForAll.Sexp()` — `varsSexp(f.Variables)` → `VarsSexp(f.Variables)`
- Line 155: `Exists.Sexp()` — `varsSexp(e.Variables)` → `VarsSexp(e.Variables)`
- Line 159: `Lambda.Sexp()` — `varsSexp(l.Variables)` → `VarsSexp(l.Variables)`
- Line 167: `NamedBinder.Sexp()` — `varsSexp(nb.Variables)` → `VarsSexp(nb.Variables)`

### Change 2: Change `key()` to return `lg.NodeKey` — file `check/l2s.go`

**Lines 162-164** — replace the entire key() method:
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

This follows the same pattern as `NamedBinder.Sexp()` at `logic/sexp.go:162-168`: dereference the `*string` Environ, build vars via `VarsSexp`, build body via `Sexp()`.

### Change 3: Change `L2sGs` map type — file `check/l2s_shared.go`

**Line 41** — struct field declaration:
```go
// BEFORE:
L2sGs             map[string]L2sGTriple

// AFTER:
L2sGs             map[lg.NodeKey]L2sGTriple
```

**Line 85-87** — `sortL2sGTriples` comment and signature:
```go
// BEFORE:
// sortL2sGTriples extracts values from a map[string]L2sGTriple and
// returns them sorted by Body.Canon() for deterministic cross-language ordering.
func sortL2sGTriples(m map[string]L2sGTriple) []L2sGTriple {

// AFTER:
// sortL2sGTriples extracts values from a map[lg.NodeKey]L2sGTriple and
// returns them sorted by Body.Canon() for deterministic cross-language ordering.
func sortL2sGTriples(m map[lg.NodeKey]L2sGTriple) []L2sGTriple {
```

**Line 104** — map initialization in `SharedStep1_ConvertTemporals`:
```go
// BEFORE:
cfg.L2sGs = make(map[string]L2sGTriple)

// AFTER:
cfg.L2sGs = make(map[lg.NodeKey]L2sGTriple)
```

### Change 4: Fix debug print — file `check/l2s.go`

**Lines 590-591** — show environ value instead of pointer address:
```go
// BEFORE:
for _, triple := range cfg.L2sGs {
    fmt.Printf("l2s_g: %v %v %v\n", triple.Vars, triple.Body, triple.Environ)
}

// AFTER:
for _, triple := range cfg.L2sGs {
    env := "<nil>"
    if triple.Environ != nil {
        env = *triple.Environ
    }
    fmt.Printf("l2s_g: %v %v %s\n", triple.Vars, triple.Body, env)
}
```

## Why this is the complete fix

The `l2sGTriple.key()` is called in exactly ONE place — `check/l2s_shared.go:110`:
```go
cfg.L2sGs[triple.key()] = triple
```

This is the only insertion into the `L2sGs` map. The map is read in three places:
- `SharedStep6_BuildTableau` line 294-296 — iterates values
- `SharedStep7_InstrumentActions` line 363 — calls `sortL2sGTriples`
- Debug print at line 590 — iterates values

All three just iterate values, so they only need the map type to match.

Other maps in the l2s code (`L2sWhensSet`, `eventProps`, `eventWhens`, `eventWaits`) use `NamedBinder.String()` as keys. While `String()` omits Environ, those maps work correctly because:
- `L2sWhensSet` entries are all created with `cfg.ProofLabel` (same pointer)
- `eventProps`/`eventWhens`/`eventWaits` are per-action local maps using the same proof context

The `dedupeVarBodyPairs` function uses `fmt.Sprintf("%v:%v", p.Vars, p.Body)` which is also value-based (Variable and Expr both have String() methods, and varBodyPair has no `*string` field), so it's not affected.

## Verification

Build:
```
cd /Users/jaten/ivy/goivy && go build ./...
```

Unit tests:
```
cd /Users/jaten/ivy/goivy/check && go test -run TestDedupeVarBodyPairs -v
cd /Users/jaten/ivy/goivy/check && go test -run TestL2s -v
```

Golden test:
```
cd /Users/jaten/ivy/goivy/parser && go test -run TestOrdLive -v -count=1 -timeout 300s
```

After the fix, the 14 raw `l2s_g` entries collapse to 6 unique entries (matching Python's `l2s_gs` set size). The Canon-sorted `toG` list produces the correct ordering, and `toG[1]` matches Python's `Not(Eq(ref.evs.l_req(_T), memc_l))`.
