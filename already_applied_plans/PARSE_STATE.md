# Plan: Incremental Merkle Hashing for Parser State Verification

**Created:** 2026-03-23 23:45
**Updated:** 2026-03-24 01:45

## Context

Golden test (`make golden`) now matches 50171 production traces between Go and Python. This confirms the same grammar rules fire in the same order, but does NOT verify that semantic actions produce identical AST nodes or accumulator state. A rule could fire and build a wrong AST node — the error only manifests later or gets silently overwritten.

**Goal:** Verify internal state equivalence incrementally during parsing, using deterministic serialization + cryptographic hashing + Merkle trees.

## Mutable State to Track

### Per-scope accumulator (Python `Ivy` class / Go `ivyAccum`):
- `decls` — ordered list of declarations (primary output)
- `modules` — map: name -> ModuleDecl
- `macros` — map: name -> MacroDecl
- `actions` — map: name -> ActionDecl
- `included` — set of module names
- `attributes` — tuple/slice of strings
- `static` — set of names
- `defined` — map: name -> [(lineno, class)]
- `objects` — map: name -> defined-dict
- `params` — list of parameter nodes
- `is_module` — bool

### Globals:
- `parent_object` / `parentObject`
- `label_counter` / `lalrLabelCounter`
- `special_attribute`, `global_attribute`, `common_attribute` (Python only; Go uses different mechanism)

## Approach: Hybrid — Leaf Hashing + Boundary Snapshots

### Phase 1: Declare-granularity leaf hashing

After every `declare()` call in both Python and Go:

1. Compute canonical string: `decl.canon()` (Python) / `decl.Canon()` (Go)
2. Leaf hash: `canon_blake3(canonical)` (Python) / `canonical.Blake3()` (Go)
3. Rolling Merkle root: `root = canon_blake3(prev_root + leaf)`
4. Emit: `XTRACE: parser.declare HASH leaf=<blake3hash> root=<blake3hash>`

The golden test compares these lines automatically — no test harness changes needed.

### Phase 2: Full-state boundary snapshots

At `inst_mod EXIT`, `create_object EXIT`, `do_insts EXIT`:

1. Serialize full accumulator state (decls in order, maps with sorted keys)
2. Hash entire state
3. Emit: `XTRACE: parser.inst_mod STATEHASH <blake3hash> decls=<N> modules=<N>`

## Hash Function Choice

**BLAKE3** — already implemented on both sides:

- Go: `~/goivy/ivyutils/canon.go` — `Canonical.Blake3()` method
  - Uses `github.com/glycerine/blake3` and `github.com/cristalhq/base64`
  - Computes 64-byte (512-bit) un-keyed blake3 hash
  - Takes first 33 bytes, URL-base64 encodes, prepends `"blake3.33B-"`

- Python: `~/pyivy/ivy/ivy/canon.py` — `canon_blake3()` function
  - Uses `import blake3` and `import base64`
  - Same algorithm: 64-byte hash → first 33 bytes → URL-base64 → `"blake3.33B-"` prefix
  - Also available as `node.blake3()` method on all AST nodes (installed by `canon_ast.install()`)

- **Verified identical output:** Both sides produce `blake3.33B-I7uNfQahZ1pogEdyvUu7EzTtRW4lRD9gVXVWeDnJaU16` for the same input.

## Canonical Serialization Strategy

Both sides use `Canon()`/`canon()` methods that produce flattened s-expression strings.
Go struct embedding is promoted (flattened) into the parent so Python's flat class
inheritance can produce identical output.

**Format:** `(typeName lineno:42 field:value field2:[elem1 elem2])` — see CLAUDE.md section E.

**Key files:**
- Go: `ast/ast.go` (canonFields), `ast/canon_decl.go`, `ast/formula.go`, `ast/tactic.go`, `ast/sort.go`, `logic/canon.go`
- Python: `~/pyivy/ivy/ivy/canon.py` (helpers), `~/pyivy/ivy/ivy/canon_ast.py` (monkey-patches canon() onto all AST classes)

For maps: sort keys lexicographically before serialization. `decls` list keeps insertion order (order is semantic).

### Drill-down on divergence

Environment variable `XTRACE_HASH_VERBOSE=1` causes both sides to emit the full canonical string instead of just hashes, making diffs human-readable.

## Implementation Steps

### Step 1: Canonical s-expression alignment — DONE

Both Go and Python produce identical flattened s-expressions for all AST types.
Verified with smoke tests on Symbol, Atom, Variable, Not, Forall, This, NoneAST.

### Step 2: Blake3 hashing — DONE

Both sides produce identical blake3 hashes for the same canonical string.
- Go: `canonical.Blake3()` in `ivyutils/canon.go`
- Python: `canon_blake3(s)` in `canon.py`, or `node.blake3()` via `canon_ast.install()`

### Step 3: Add Merkle hash infrastructure to Go xtracer

**File:** `/Users/jaten/goivy/xtracer/xtracer.go`

Add:
```go
type MerkleState struct {
    prevRoot string // blake3 hash string
}

func (ms *MerkleState) AddLeaf(c iu.Canonical) (leafB3, rootB3 string) {
    leafB3 = c.Blake3()
    combined := iu.Canonical(ms.prevRoot + leafB3)
    ms.prevRoot = combined.Blake3()
    rootB3 = ms.prevRoot
    return
}
```

Gated by the existing `xtracer` build tag.

### Step 4: Add Merkle hash infrastructure to Python xtracer

**File:** `/Users/jaten/pyivy/ivy/ivy/xtracer.py`

Add:
```python
from .canon import canon_blake3

_merkle_root = ''

def add_leaf(canonical_str):
    """Add a leaf to the Merkle tree and return (leaf_hash, root_hash)."""
    global _merkle_root
    leaf = canon_blake3(canonical_str)
    _merkle_root = canon_blake3(_merkle_root + leaf)
    return leaf, _merkle_root
```

### Step 5: Instrument Python `Ivy.declare()`

**File:** `/Users/jaten/pyivy/ivy/ivy/ivy_parser.py` — after `self.decls.append(decl)` (~line 309)

```python
if hasattr(decl, 'canon'):
    canonical = decl.canon()
    leaf, root = xtracer.add_leaf(canonical)
    xtracer.trace("parser.declare HASH leaf=%s root=%s", leaf, root)
```

### Step 6: Instrument Go `ivyAccum.declare()`

**File:** `/Users/jaten/goivy/lalr_full/ivy_module.go` — after `m.decls = append(m.decls, decl)` (line 108)

```go
if xtracer.Enabled {
    canonical := decl.Canon()
    leaf, root := xtracer.GlobalMerkle.AddLeaf(canonical)
    xtracer.Trace("parser.declare HASH leaf=%s root=%s", leaf, root)
}
```

### Step 7: Run `make golden` and iterate

The golden test will show mismatches in hash lines. Use `XTRACE_HASH_VERBOSE=1` to see full canonical strings and fix differences one type at a time.

### Step 8 (Phase 2): Full-state boundary snapshots

Add `canonicalState()` method to both `Ivy` and `ivyAccum` that serializes all fields deterministically. Instrument `inst_mod EXIT`, `create_object EXIT`, `do_insts EXIT`.

## Verification

```bash
# Phase 1: Run golden test — hash lines should match
cd ~/goivy && make golden

# Debug mode: see canonical strings when hashes diverge
XTRACE_HASH_VERBOSE=1 make golden
```

## Options Considered but Not Chosen

1. **Hash after every production** (~50K) — too noisy, most productions don't mutate accumulator state
2. **End-of-file only** — defeats the purpose; bad state gets overwritten before check
3. **JSON serialization** — more verbose than canon s-expressions, no advantage
4. **SHA256** — originally considered but blake3 is faster and already available on both sides
