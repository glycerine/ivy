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

### Phase 3: Change `IterSymbolsASTNode` to Yield `any` (matching Python's mixed-type yields)

Python's `symbols_ast` yields `str` from `Atom.rep` and `Symbol` objects from `App.rep`. To faithfully mirror this, Go's `IterSymbolsASTNode` should yield `any` — returning `string` for Atom and `*Symbol` (or `Node`) for App.

**File: `ast/tactic.go`** — Change signatures and implementation:

```go
// IterSymbolsASTNode yields values from an AST node tree.
// Port of Python symbols_ast (ivy_ast.py:1879) as iter.Seq[any].
//
// Yields string for *Atom (matching Python str from Atom.rep)
// and Node for *App (matching Python Symbol from App.rep).
// Both Atom and App (and all other nodes) recurse on Args() children.
func IterSymbolsASTNode(node Node) iter.Seq[any] {
    return func(yield func(any) bool) {
        iterSymbolsASTNodeRec(node, yield)
    }
}

func iterSymbolsASTNodeRec(node Node, yield func(any) bool) bool {
    if node == nil {
        return true
    }
    switch v := node.(type) {
    case *Atom:
        if v.Rep != "" {
            if !yield(v.Rep) {  // yields string — matches Python str
                return false
            }
        }
    case *App:
        if v.Rep != nil {
            if !yield(v.Rep) {  // yields Node (usually *Symbol) — matches Python Symbol
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
```

### Phase 4: Change VocabNames to `map[any]bool` (matching Python's `set()`)

Python's `all_names` is a `set()` containing both `str` and `Symbol` objects as separate entries (`Symbol.__eq__` checks type). To match this in Go, change VocabNames from `Omap[string, bool]` to a plain `map[any]bool`.

**File: `ast/tactic.go`**:

```go
// VocabNames is a set of mixed-type values (string and Node),
// matching Python's set() which stores both str and Symbol objects.
type VocabNames = map[any]bool

func NewVocabNames() *VocabNames {
    m := make(VocabNames)
    return &m
}

// VocabNamesUpdate consumes an iter.Seq[any] and adds to the set.
func VocabNamesUpdate(vn *VocabNames, seq iter.Seq[any]) {
    for val := range seq {
        (*vn)[val] = true
    }
}
```

Note: `*ast.Symbol` implements the necessary interface for use as a map key (pointer identity) which matches Python's `Symbol.__eq__` (compares by type + rep). However, we need `Symbol` to be comparable by value (rep), not pointer. Options:

- **Use `fmt.Sprintf("Symbol:%s", sym.Rep)`** as the key for Symbol entries (string representation in map)
- **Or** wrap in a struct: `type vocabKey struct { kind string; name string }` where kind is "str" or "sym"

Cleanest approach — use a small wrapper struct as key:

```go
// VocabKey distinguishes str-sourced vs Symbol-sourced names,
// matching Python's set() where str("foo") != Symbol("foo").
type VocabKey struct {
    Name  string
    IsSym bool  // true = from App.Rep (*Symbol in Go, Symbol in Python)
}

type VocabNames struct {
    M map[VocabKey]bool
}

func NewVocabNames() *VocabNames {
    return &VocabNames{M: make(map[VocabKey]bool)}
}
```

And `VocabNamesUpdate`:
```go
func VocabNamesUpdate(vn *VocabNames, seq iter.Seq[any]) {
    for val := range seq {
        switch v := val.(type) {
        case string:
            vn.M[VocabKey{Name: v, IsSym: false}] = true
        case *Symbol:
            vn.M[VocabKey{Name: v.Rep, IsSym: true}] = true
        case *This:
            vn.M[VocabKey{Name: "this", IsSym: true}] = true
        }
    }
}
```

Downstream consumer at `isolate.go:1016`:
```go
// allNames.Get2(c.Name) → check string-sourced names only (matching Python)
_, found := allNames.M[VocabKey{Name: c.Name, IsSym: false}]
```

This also checks only str-sourced names — matching Python where `"foo" in all_names` matches str entries but not Symbol entries.

For traces, iterate sorted by Name:
```go
// Collect all names as strings for tracing
var traceNames []string
for k := range allNames.M {
    traceNames = append(traceNames, k.Name)
}
sort.Strings(traceNames)
for _, name := range traceNames {
    xtracer.Trace("isolate.allNames_from_proofs.name %s", name)
}
```

### Phase 5: Update All VocabNames Consumers

Consumers that call `vn.Set(name, true)` or `vn.Get2(name)` need to use the new `VocabKey`/map API. Check:
- `isolate/isolate.go` — `allNames.Get2(c.Name)` and `allNamesMap` construction
- Any other callers of `VocabNamesUpdate`, `NewVocabNames`

## Files Modified

| File | Changes |
|------|---------|
| `ast/tactic.go` | `IterSymbolsASTNode` yields `any` (string or Node); `VocabNames` uses `VocabKey` struct + `map`; `VocabNamesUpdate` type-switches on `string`/`*Symbol`/`*This` |
| `isolate/isolate.go` | Add per-proof delta traces; update `allNames.Get2()` → `allNames.M[VocabKey{...}]`; update trace loop to sort by Name |
| `ivy_isolate.py` (lines 1260-1268) | Add per-proof delta traces; add type-breakdown diagnostic trace |

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...`
2. `cd ~/goivy && make test`
3. `cd ~/goivy && make golden` — divergence should advance past line 152573 with matching n=329 count
