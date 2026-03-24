# Plan: Incremental Merkle Hashing for Parser State Verification

**Created:** 2026-03-23 23:45

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

1. Compute canonical string: `typeName + "|" + repr(decl) + "|attrs=" + repr(decl.attributes) + "|common=" + repr(decl.common)`
2. Leaf hash: `leaf = sha256(canonical)`
3. Rolling Merkle root: `root = sha256(prev_root || leaf)`
4. Emit: `XTRACE: parser.declare HASH leaf=<first16hex> root=<first16hex>`

The golden test compares these lines automatically — no test harness changes needed.

### Phase 2: Full-state boundary snapshots

At `inst_mod EXIT`, `create_object EXIT`, `do_insts EXIT`:

1. Serialize full accumulator state (decls in order, maps with sorted keys)
2. Hash entire state
3. Emit: `XTRACE: parser.inst_mod STATEHASH <first16hex> decls=<N> modules=<N>`

## Hash Function Choice

**SHA256** — zero-dependency on both sides (`crypto/sha256` in Go, `hashlib.sha256` in Python). ~200 declare events per file = ~0.2ms total overhead. Blake3 can be swapped in later if needed.

## Canonical Serialization Strategy

Use Python `repr()` as the canonical format. Go `String()` methods must produce identical output.

**Key alignment needed:** Python `Decl.__repr__` uses `','` (no space after comma) while Go may use `", "`. Go must match Python exactly.

For maps: sort keys lexicographically before serialization. `decls` list keeps insertion order (order is semantic).

### Drill-down on divergence

Environment variable `XTRACE_HASH_VERBOSE=1` causes both sides to emit the full canonical string instead of just hashes, making diffs human-readable.

## Implementation Steps

### Step 1: Canonical repr alignment (PREREQUISITE)

Verify and fix Go AST `String()` methods to exactly match Python `repr()` output for all declaration types.

**Files:**
- `/Users/jaten/goivy/ast/ast.go` — base `String()` methods
- `/Users/jaten/goivy/ast/decl.go` — declaration `String()` methods
- Compare against `/Users/jaten/pyivy/ivy/ivy/ivy_ast.py` lines 53-54, 275-280, 579-581

### Step 2: Add Merkle hash infrastructure to Go xtracer

**File:** `/Users/jaten/goivy/xtracer/xtracer.go`

Add:
```go
type MerkleState struct {
    prevRoot [32]byte
}

func (ms *MerkleState) AddLeaf(canonical string) (leafHex, rootHex string) {
    leaf := sha256.Sum256([]byte(canonical))
    combined := append(ms.prevRoot[:], leaf[:]...)
    ms.prevRoot = sha256.Sum256(combined)
    return hex.EncodeToString(leaf[:8]), hex.EncodeToString(ms.prevRoot[:8])
}
```

Gated by the existing `xtracer` build tag.

### Step 3: Add Merkle hash infrastructure to Python xtracer

**File:** `/Users/jaten/pyivy/ivy/ivy/xtracer.py`

Add:
```python
import hashlib
_merkle_root = b'\x00' * 32

def add_leaf(canonical):
    global _merkle_root
    leaf = hashlib.sha256(canonical.encode()).digest()
    _merkle_root = hashlib.sha256(_merkle_root + leaf).digest()
    return leaf[:8].hex(), _merkle_root[:8].hex()
```

### Step 4: Instrument Python `Ivy.declare()`

**File:** `/Users/jaten/pyivy/ivy/ivy/ivy_parser.py` — after `self.decls.append(decl)` (~line 309)

```python
canonical = type(decl).__name__ + "|" + repr(decl) + "|attrs=" + repr(getattr(decl, 'attributes', None)) + "|common=" + repr(getattr(decl, 'common', None))
leaf, root = xtracer.add_leaf(canonical)
xtracer.trace("parser.declare HASH leaf=%s root=%s", leaf, root)
```

### Step 5: Instrument Go `ivyAccum.declare()`

**File:** `/Users/jaten/goivy/lalr_full/ivy_module.go` — after `m.decls = append(m.decls, decl)` (line 108)

```go
if xtracer.Enabled {
    typeName := reflect.TypeOf(decl).Elem().Name()
    canonical := typeName + "|" + decl.String() + "|attrs=" + attrsRepr(decl) + "|common=" + commonRepr(decl)
    leaf, root := xtracer.GlobalMerkle.AddLeaf(canonical)
    xtracer.Trace("parser.declare HASH leaf=%s root=%s", leaf, root)
}
```

### Step 6: Run `make golden` and iterate on repr alignment

The golden test will show mismatches in hash lines. Use `XTRACE_HASH_VERBOSE=1` to see full canonical strings and fix repr differences one type at a time.

### Step 7 (Phase 2): Full-state boundary snapshots

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
3. **JSON serialization** — more verbose than repr, no advantage
4. **Blake3** — better performance but adds dependency; SHA256 sufficient for ~200 events/file
