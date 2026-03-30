# Plan: Fix allNames_from_proofs Count Divergence (295 vs 329)

**Created**: 2026-03-31 04:10

## Context

Golden test diverges at line 152573:
```
152573  go : XTRACE: isolate.allNames_from_proofs n=295
        py : XTRACE: isolate.allNames_from_proofs n=329
```

Go collects 295 names from proofs, Python collects 329. The 34 missing names are caused by a fundamental difference in how `symbols_ast` / `IterSymbolsASTNode` work.

## Root Cause

**Python `symbols_ast`** (`ivy_ast.py:1879`) yields `ast.rep` for BOTH `App` AND `Atom`:
```python
def symbols_ast(ast):
    if isinstance(ast, (App, Atom)):
        yield ast.rep          # App.rep is a Symbol object; Atom.rep is a str
    ...recurse into ast.args...
```

**Go `IterSymbolsASTNode`** (`ast/tactic.go:486`) only yields from `*Atom`, intentionally skipping `*App`:
```go
if atom, ok := node.(*Atom); ok && atom.Rep != "" {
    if !yield(atom.Rep) { return false }
}
// App is SKIPPED — comment says "Symbol.__eq__ rejects string comparisons"
```

### Why counts differ

Python's `all_names` is a `set()` containing **both** `str` objects (from `Atom.rep`) and `Symbol` objects (from `App.rep`). These are separate set entries because `Symbol.__eq__` checks `type(self) == type(other)` — so `"foo" != Symbol("foo")` even though they have the same hash.

Go's `allNames` is `Omap[string, bool]` — only strings from Atom. So it has 34 fewer entries.

### Key type details

| Side | `Atom.rep` type | `App.rep` type | Storage |
|------|-----------------|----------------|---------|
| Python | `str` | `Symbol` object | `set()` — both coexist |
| Go | `string` | `Node` (usually `*Symbol`) | `Omap[string, bool]` — only Atom strings |

Python's `This` class also has a `rep` property returning `"this"` (str), so `This` nodes inside an `App` tree could contribute names too.

## Fix

### Phase 1: Add Diagnostic Per-Proof Delta Traces (both sides)

These traces show, for each proof, how many names it adds. This narrows down WHICH proofs contribute the extra names in Python.

**File: `isolate/isolate.go`** — Change the proof loop (~line 988):

```go
allNames := ast.NewVocabNames()
if xtracer.Enabled {
    xtracer.Trace("isolate.proofs n=%d", len(mod.Proofs))
}
for i, pe := range mod.Proofs {
    if pe.Proof != nil {
        before := allNames.Len()
        if xtracer.Enabled {
            xtracer.Trace("isolate.proof type=%s", typeName(pe.Proof))
        }
        ast.VocabNode(pe.Proof, allNames)
        if xtracer.Enabled {
            xtracer.Trace("isolate.proof[%d].vocab delta=%d total=%d", i, allNames.Len()-before, allNames.Len())
        }
    }
}
```

**File: `/Users/jaten/pyivy/ivy/ivy/ivy_isolate.py`** — Change lines 1260-1268:

```python
    all_names = set()
    if __debug__: xtracer.trace("isolate.proofs n=%d" % len(mod.proofs))
    for i, x in enumerate(mod.proofs):
        before = len(all_names)
        if __debug__: xtracer.trace("isolate.proof type=%s" % type(x[1]).__name__)
        x[1].vocab(all_names)
        if __debug__: xtracer.trace("isolate.proof[%d].vocab delta=%d total=%d" % (i, all_names_len - before, len(all_names)))
    if __debug__:
        xtracer.trace("isolate.allNames_from_proofs n=%d" % len(all_names))
        for name in sorted(str(x) for x in all_names):
            xtracer.trace("isolate.allNames_from_proofs.name %s" % name)
```

### Phase 2: Add Type-Breakdown Trace in Python

Shows how many of the 329 are strings vs Symbol objects, confirming the hypothesis:

```python
    if __debug__:
        from ivy_ast import Symbol as _Sym
        n_str = sum(1 for x in all_names if isinstance(x, str))
        n_sym = sum(1 for x in all_names if isinstance(x, _Sym))
        n_other = len(all_names) - n_str - n_sym
        xtracer.trace("isolate.allNames_from_proofs n=%d (str=%d sym=%d other=%d)" % (len(all_names), n_str, n_sym, n_other))
```

This trace is Python-only (diagnostic). Expected: `n=329 (str=295 sym=34 other=0)` confirming the 34 extra are Symbol objects from App nodes.

### Phase 3: Fix Go `IterSymbolsASTNode` to Also Yield from App

**File: `ast/tactic.go`** — Change `iterSymbolsASTNodeRec` to also yield from `*App`:

```go
func iterSymbolsASTNodeRec(node Node, yield func(string) bool) bool {
    if node == nil {
        return true
    }
    switch v := node.(type) {
    case *Atom:
        if v.Rep != "" {
            if !yield(v.Rep) {
                return false
            }
        }
    case *App:
        // Python yields App.rep (a Symbol object). Extract the string name.
        // In Go, App.Rep is a Node (usually *Symbol).
        if name := nodeRepString(v.Rep); name != "" {
            if !yield(name) {
                return false
            }
        }
    }
    for _, child := range node.Args() {
        if !iterSymbolsASTNodeRec(child, yield) {
            return false
        }
    }
    return true
}

// nodeRepString extracts the string name from a Node used as App.Rep.
// Python: App.rep is a Symbol (has .rep str) or This (has .rep property = "this").
func nodeRepString(n Node) string {
    switch v := n.(type) {
    case *Symbol:
        return v.Rep
    case *This:
        return "this"
    default:
        return ""
    }
}
```

### Phase 4: Handle Python's Mixed-Type Set in VocabNames

Python's `set()` stores both `str("foo")` and `Symbol("foo")` as **separate entries** (different types → not equal). Go's `Omap[string, bool]` merges them into one.

To match Python's count, change VocabNames to distinguish string-sourced vs symbol-sourced names. Use a struct key:

```go
// VocabEntry distinguishes names from Atom (string source) vs App (symbol source).
// Matches Python's set() which stores str and Symbol objects as separate entries.
type VocabEntry struct {
    Name   string
    FromApp bool  // true = came from App.Rep (Symbol in Python), false = from Atom.Rep (str in Python)
}

type VocabNames = iu.Omap[VocabEntry, bool]
```

**Alternative simpler approach**: Use a tagged string key — e.g., prefix App-sourced names with `"\x00"` (a byte that never appears in symbol names) to distinguish them from Atom-sourced strings. This avoids changing the Omap key type:

```go
// In IterSymbolsASTNode, yield App-sourced names with a "\x00" prefix:
case *App:
    if name := nodeRepString(v.Rep); name != "" {
        if !yield("\x00" + name) { return false }
    }
```

Then traces strip the prefix: `strings.TrimPrefix(name, "\x00")`.
And downstream `Get2(c.Name)` only matches unprefixed (Atom-sourced) — matching Python where `str in set` only matches str entries, not Symbol entries.

This is the cleanest approach because:
- `Omap[string, bool]` type doesn't change
- App-sourced names never collide with Atom-sourced names (different keys)
- Downstream `allNames.Get2(c.Name)` naturally ignores App-sourced names (prefix mismatch) — exactly matching Python where `"foo" in all_names` doesn't match `Symbol("foo")`
- `allNames.Len()` counts both types — matching Python's `len(all_names)`
- Traces use `strings.TrimPrefix` to show clean names — matching Python's `str(x)`

## Files Modified

| File | Changes |
|------|---------|
| `ast/tactic.go` | Fix `IterSymbolsASTNode` to yield from `*App` with `"\x00"` prefix; add `nodeRepString` helper |
| `isolate/isolate.go` | Add per-proof delta traces; strip `"\x00"` prefix in name traces |
| `ivy_isolate.py` (lines 1260-1268) | Add per-proof delta traces; add type-breakdown diagnostic trace |

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...`
2. `cd ~/goivy && make test`
3. `cd ~/goivy && make golden` — run with diagnostic traces to confirm hypothesis
4. Once confirmed, remove Python-only diagnostic trace (type breakdown) and verify golden advances past line 152573
