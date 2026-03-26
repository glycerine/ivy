# Plan: Merkle-chained Sig conformance auditing via Canon + HASH

**Created:** 2026-03-27T06:00

## Context

We just found that Go's `CompileSchemaBody` was replacing `c.Sig` with a fresh sig, losing the outer scope's sorts (like `index`). Python never replaces the global `ivy_logic.sig` — it accumulates schema-local state in a parameter sig and temporarily merges it via `WithSorts`/`WithSymbols`. The bug was invisible until xtracer diagnostics revealed `sigSorts=[bool q]` (missing `index`) in Go vs `sigSorts=['bool', 'index']` in Python.

**Goal:** Add a Merkle-chained Sig canon check at every key signature state transition point. When both sides emit `HASH` + `canon=` lines, the golden test (line 544) automatically runs `DiffSexp` to show exactly which sorts/symbols diverge. The rolling Merkle root catches the *first* divergence point — if the chain breaks, we know the exact transition that went wrong.

## Design

### 1. Sig Canon format

Both sides emit a canonical s-expression for the Sig:

```
(sig sorts:[bool index q] symbols:[base:FunctionSort fun:FunctionSort step:FunctionSort])
```

- `sorts:` — alphabetically sorted list of sort names
- `symbols:` — alphabetically sorted list of `name:sortKind` pairs (sortKind is the sort type name, not full sort canon — keeps it concise)

### 2. Merkle chain on the Compiler

**Go:** Add `SigMerkle iu.MerkleState` field to `Compiler` struct.

**Python:** Add `sig_merkle = xtracer.MerkleState()` as a module-level global in `ivy_compiler.py` (Python has no Compiler struct — compiler state is module globals).

Both sides feed Sig canon into the same chain:
```
leaf, root = merkle.AddLeaf(sigCanon)
xtracer.Trace("compiler.SigCheck HASH leaf=%s root=%s canon=%s", leaf, root, sigCanon)
```

The golden test sees both lines with `HASH` + `canon=`, extracts and diffs them.

### 3. Audit points — where to emit SigCheck

These are all points where `c.Sig` state transitions (set, copied, restored, or mutated). Each gets a matching trace on both sides.

| # | Label | Go location | Python location | When |
|---|-------|-------------|-----------------|------|
| 1 | `SigCheck@CompileDefnImpl.entry` | `compiler.go` compileDefnImpl top | `ivy_compiler.py` compile_defn after `with sig:` | Before any sig copy/mutation |
| 2 | `SigCheck@CompileDefnImpl.afterCopy` | `compiler.go` after `c.Sig = sigCopy` | `ivy_compiler.py` after `with sig:` enters | After sig is copied for defn |
| 3 | `SigCheck@CompileDefnImpl.restored` | `compiler.go` after `c.Sig = savedSig` | `ivy_compiler.py` after `with sig:` exits | Sig restored after defn |
| 4 | `SigCheck@SchemaBody.entry` | `phase6.go` CompileSchemaBody top | `ivy_compiler.py` compile_schema_body top | Before fresh sig creation |
| 5 | `SigCheck@SchemaBody.afterPrems` | `phase6.go` after premise loop | `ivy_compiler.py` after prems list comp | Schema sig after premises |
| 6 | `SigCheck@SchemaConc.entry` | `phase6.go` CompileSchemaConcWithSig | `ivy_compiler.py` compile_schema_conc entry | Before WithSorts/WithSymbols |
| 7 | `SigCheck@SchemaConc.withCtx` | `phase6.go` after ws/wss.Enter() | `ivy_compiler.py` inside `with` blocks | After sorts/symbols added |
| 8 | `SigCheck@SortifyWithInference` | `compiler.go` SortifyWithInference entry | `ivy_compiler.py` sortify_with_inference entry | Before compilation + sort infer |
| 9 | `SigCheck@DomainSetup.derived` | `decl.go` Derived() entry | `ivy_compiler.py` DomainSetup derived dispatch | At derived decl compilation |
| 10 | `SigCheck@DomainSetup.definition` | `decl.go` DefinitionDecl() entry | `ivy_compiler.py` DomainSetup definition dispatch | At definition decl compilation |
| 11 | `SigCheck@DomainSetup.schema` | `decl.go` Schema() entry | `ivy_compiler.py` DomainSetup schema dispatch | At schema decl compilation |
| 12 | `SigCheck@IvyCompileTheory` | `phase6.go` IvyCompileTheory entry | `ivy_compiler.py` IvyCompileTheory entry | At theory compilation start |
| 13 | `SigCheck@CompileConst.after` | `compiler.go` CompileConst after AddSymbol | `ivy_compiler.py` compile_const after add_symbol | After constant added to sig |

### 4. SigCheck helper function

**Go** — method on `*Compiler`:
```go
func (c *Compiler) SigCheck(label string) {
    if !xtracer.Enabled { return }
    canon := c.Sig.Canon()
    leaf, root := c.SigMerkle.AddLeaf(canon)
    xtracer.Trace("compiler.SigCheck@%s HASH leaf=%s root=%s canon=%s", label, leaf, root, string(canon))
}
```

**Python** — module function:
```python
def sig_check(label):
    if not __debug__: return
    canon = sig_canon(ivy_logic.sig)
    leaf, root = sig_merkle.add_leaf(canon)
    xtracer.trace("compiler.SigCheck@%s HASH leaf=%s root=%s canon=%s" % (label, leaf, root, canon))
```

### 5. Sig.Canon() implementation

**Go** — `ivylogic/sig.go`:
```go
func (s *Sig) Canon() iu.Canonical {
    // Sort names alphabetically
    sortNames := s.SortNames()
    sort.Strings(sortNames)
    // Symbol names + sort kind alphabetically
    symParts := make([]string, 0, len(s.Symbols))
    for name, entry := range s.Symbols {
        symParts = append(symParts, name+":"+lg.SortName(entry.Sort))
    }
    sort.Strings(symParts)
    return iu.Canonical(fmt.Sprintf("(sig sorts:[%s] symbols:[%s])",
        strings.Join(sortNames, " "),
        strings.Join(symParts, " ")))
}
```

**Python** — `ivy_compiler.py`:
```python
def sig_canon(sig):
    sort_names = sorted(sig.sorts.keys())
    sym_parts = sorted("%s:%s" % (name, type(sym.sort).__name__) for name, sym in sig.symbols.items())
    return "(sig sorts:[%s] symbols:[%s])" % (" ".join(sort_names), " ".join(sym_parts))
```

**Key:** sort KIND (type name like `FunctionSort`, `UninterpretedSort`) rather than full sort canon — keeps the strings comparable without needing to canonize every sort's internals. Sort kind names must match between Go and Python (e.g. both emit `FunctionSort`).

## Files to modify

1. **`/Users/jaten/goivy/ivylogic/sig.go`** — Add `Canon()` method on `*Sig`
2. **`/Users/jaten/goivy/compiler/compiler.go`** — Add `SigMerkle` field to `Compiler`, `SigCheck()` helper, audit points in `compileDefnImpl`, `SortifyWithInference`, `CompileConst`
3. **`/Users/jaten/goivy/compiler/phase6.go`** — Audit points in `CompileSchemaBody`, `CompileSchemaConcWithSig`, `IvyCompileTheory`
4. **`/Users/jaten/goivy/compiler/decl.go`** — Audit points in `Derived`, `DefinitionDecl`, `Schema`
5. **`/Users/jaten/pyivy/ivy/ivy/ivy_compiler.py`** — Add `sig_canon()`, `sig_merkle`, `sig_check()`, matching audit points
6. **`/Users/jaten/pyivy/ivy/ivy/ivy_logic.py`** — (no changes needed — Python sig_canon reads sig from outside)

## Sort representation in symbol entries (verified matching)

For the `symbols:` part of the canon, use `il.SortName(entry.Sort)` in Go and `str(sym.sort)` in Python. These produce identical strings:

| Sort type | Go `il.SortName()` / `.String()` | Python `str(sort)` | Example |
|---|---|---|---|
| UninterpretedSort | `.Name` | `.name` | `index` |
| FunctionSort | `"dom1 * dom2 -> rng"` | `"dom1 * dom2 -> rng"` | `index * index -> q` |
| BooleanSort | `"bool"` | `"Boolean"` | **MISMATCH — needs fix** |
| TopSort | `.Name` ("TopSort") | `.name` ("TopSort") | `TopSort` |
| EnumeratedSort | `.Name` | `"{ext1,ext2}"` via `__str__` | **MISMATCH — use .name** |

**Fixes needed:** For BooleanSort, Go returns "bool" but Python returns "Boolean". For EnumeratedSort, Go uses `.Name` but Python `str()` uses `{ext}` format. Solution: use `.name` attribute on Python side instead of `str()` for the symbol sort representation — or add a `sort_name()` helper that matches Go's `SortName()`.

Python `sort_name()` helper:
```python
def sort_name(s):
    if hasattr(s, 'name'):
        return s.name
    if isinstance(s, lg.BooleanSort):
        return 'bool'
    if isinstance(s, lg.FunctionSort):
        return str(s)
    return str(s)
```

This matches Go's `il.SortName()` exactly.

## Verification

```bash
cd ~/goivy && go build ./... && make golden
```

When Sig state diverges, the golden test output will show:
```
=== S-expression diff (go '-' vs py '+') ===
 (sig
-  sorts:[bool q]
+  sorts:[bool index q]
   symbols:[...])
```

The Merkle root divergence pinpoints the exact SigCheck where Go and Python first disagree.
