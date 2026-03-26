# Plan: Extend Merkle-chained Canon framework — Module state + Z3 request hashing

**Created:** 2026-03-27T14:30

## Context

The Merkle-chained SigCheck framework successfully catches Go/Python divergences in the compilation signature (sorts + symbols). But Sig is only one piece of compiler state. Two extensions will catch bugs earlier and in more places:

1. **Module state**: The Module accumulates definitions, axioms, props, inits, conjectures, schemata, etc. during compilation. Hashing this alongside Sig in every SigCheck call catches divergences in accumulated declarations, not just type-system state.

2. **Z3 API requests**: When Ivy sends formulas to Z3 for checking, the translated expressions must match between Go and Python. Hashing the Z3-bound formula at the translation boundary catches solver divergences.

## What exists today

- **SigCheck** (12 call sites, Go+Python matched): Hashes `Sig.Canon()` into `c.SigMerkle` chain. Emits `compiler.SigCheck@label HASH leaf=... root=... canon=...`
- **Module.CanonSnapshot** (Go: `module/canon.go:17`, Python: `ivy_module.py:289`): Emits individual xtracer lines for each module field (axioms, defs, props, inits, conjs, schemata, initCond, theory). Called at 3 points: after-domain-setup, after-conj-setup, after-arg-setup. **NOT Merkle-chained** — just raw traces.
- **Module has no `Canon()` method** — only `CanonSnapshot()` which emits multiple traces.
- **Z3 translation** — Go: `z3bridge/translate.go` `Translator.Translate(logic.Expr)`. Python: `ivy_solver.py` with `to_z3()` methods. **No canon/hash points exist.**

## Phase A: Add Module.Canon() and hash it into SigCheck

### A1: Add `Module.Canon()` → single canonical s-expression (Go)

**File**: `/Users/jaten/goivy/module/canon.go`

Add a `Canon()` method that returns a single `iu.Canonical` string combining the key module fields. Reuse the existing `canonLFSlice`, `canonSortMap`, `canonSchemaMap` helpers.

```go
func (m *Module) Canon() iu.Canonical {
    var parts []string
    parts = append(parts, "axioms:"+canonLFSlice(m.LabeledAxioms))
    parts = append(parts, "defs:"+canonLFSlice(m.Definitions))
    parts = append(parts, "props:"+canonLFSlice(m.LabeledProps))
    parts = append(parts, "inits:"+canonLFSlice(m.LabeledInits))
    parts = append(parts, "conjs:"+canonLFSlice(m.LabeledConjs))
    parts = append(parts, "schemata:"+canonSchemaMap(m.Schemata))
    return iu.Canonical(fmt.Sprintf("(module %s)", strings.Join(parts, " ")))
}
```

### A2: Add `Module.canon()` → single canonical s-expression (Python)

**File**: `/Users/jaten/pyivy/ivy/ivy/ivy_module.py`

```python
def canon(self):
    parts = []
    parts.append("axioms:" + _canon_lf_slice(self.labeled_axioms))
    parts.append("defs:" + _canon_lf_slice(self.definitions))
    parts.append("props:" + _canon_lf_slice(self.labeled_props))
    parts.append("inits:" + _canon_lf_slice(self.labeled_inits))
    parts.append("conjs:" + _canon_lf_slice(self.labeled_conjs))
    parts.append("schemata:" + _canon_schema_map(self.schemata))
    return "(module %s)" % " ".join(parts)
```

### A3: Extend SigCheck to also hash Module (Go)

**File**: `/Users/jaten/goivy/compiler/compiler.go` ~line 152

Change `SigCheck` to hash both Sig and Module into the Merkle chain:

```go
func (c *Compiler) SigCheck(label string) {
    sigCanon := c.Sig.Canon()
    modCanon := iu.Canonical("")
    if c.Module != nil {
        modCanon = c.Module.Canon()
    }
    combined := iu.Canonical(string(sigCanon) + string(modCanon))
    leaf, root := c.SigMerkle.AddLeaf(combined)
    xtracer.Trace("compiler.SigCheck@%s HASH leaf=%s root=%s canon=%s",
        label, leaf, root, string(combined))
}
```

### A4: Extend sig_check to also hash module (Python)

**File**: `/Users/jaten/pyivy/ivy/ivy/ivy_compiler.py` ~line 56

```python
def sig_check(label):
    sig_c = sig_canon(ivy_logic.sig)
    mod_c = ""
    if ivy_module.module is not None:
        mod_c = ivy_module.module.canon()
    combined = sig_c + mod_c
    leaf, root = sig_merkle.add_leaf(combined)
    xtracer.trace("compiler.SigCheck@%s HASH leaf=%s root=%s canon=%s" % (label, leaf, root, combined))
```

## Phase B: Merkle-chain the Z3 solver check calls

Both sides already emit `z3.check seq=N result=... smt2=...` traces via:
- **Go**: `z3bridge/xtrace.go:59` `TraceCheck()` called from `Solver.Check()` (quantifier.go:1443)
- **Python**: `ivy_solver.py:1183` `_trace_z3_check()` called from monkey-patched `z3.Solver.check` (line 1192-1196)

Both sides also have matching `NormalizeZ3VarNames` / `_normalize_z3_varnames` for deterministic variable naming.

**What's missing**: These traces are not Merkle-chained. A divergence in solver call N is only caught if the golden test happens to compare that exact line. By feeding the normalized SMT2 into a Merkle chain, any earlier divergence corrupts the root for all subsequent checks.

### B1: Add Z3Merkle to Go's xtrace.go

**File**: `/Users/jaten/goivy/z3bridge/xtrace.go`

Add a package-level `z3Merkle` (this is debug-only state, exempt from the no-globals rule per CLAUDE.md section C.10):

```go
var z3Merkle iu.MerkleState

func TraceCheck(s *Solver, result CheckResult) {
    if !xtracer.Enabled {
        return
    }
    seq := atomic.AddInt64(&z3CheckCounter, 1)
    smt2 := NormalizeZ3VarNames(s.String())

    var rs string
    switch result {
    case Sat:  rs = "sat"
    case Unsat: rs = "unsat"
    default:   rs = "unknown"
    }

    // Merkle-chain the solver state + result
    canon := iu.Canonical(fmt.Sprintf("(z3check seq=%d result=%s smt2=%s)", seq, rs, smt2))
    leaf, root := z3Merkle.AddLeaf(canon)
    xtracer.Trace("z3.check seq=%d result=%s HASH leaf=%s root=%s smt2=%s", seq, rs, leaf, root, smt2)
}
```

### B2: Add z3_merkle to Python's _trace_z3_check

**File**: `/Users/jaten/pyivy/ivy/ivy/ivy_solver.py`

Add module-level `_z3_merkle = xtracer.MerkleState()` near existing `_z3_check_counter`:

```python
_z3_merkle = xtracer.MerkleState()

def _trace_z3_check(s, res):
    _z3_check_counter[0] += 1
    smt2 = _normalize_z3_varnames(s.to_smt2())
    rs = 'sat' if res == z3.sat else ('unsat' if res == z3.unsat else 'unknown')
    # Merkle-chain the solver state + result
    canon = "(z3check seq=%d result=%s smt2=%s)" % (_z3_check_counter[0], rs, smt2)
    leaf, root = _z3_merkle.add_leaf(canon)
    xtracer.trace("z3.check seq=%d result=%s HASH leaf=%s root=%s smt2=%s" % (
        _z3_check_counter[0], rs, leaf, root, smt2))
```

### B3: Also Merkle-chain the Translate() input (Go)

**File**: `/Users/jaten/goivy/z3bridge/translate.go`

Add a `TranslateMerkle iu.MerkleState` field to `Translator` struct. Hash the input expression at translation entry:

```go
func (t *Translator) Translate(n logic.Expr) (Expr, error) {
    if xtracer.Enabled {
        canon := n.Canon()
        leaf, root := t.TranslateMerkle.AddLeaf(canon)
        xtracer.Trace("z3bridge.Translate HASH leaf=%s root=%s canon=%s", leaf, root, string(canon))
    }
    // ... existing code ...
}
```

### B4: Match Translate input hashing in Python

**File**: `/Users/jaten/pyivy/ivy/ivy/ivy_solver.py`

Add `_translate_merkle = xtracer.MerkleState()` and hash at the `formula_to_z3` entry point (the Python equivalent of `Translator.Translate`):

```python
_translate_merkle = xtracer.MerkleState()

# In formula_to_z3 or the main to_z3 dispatch:
if __debug__:
    canon = fmla.sexp()
    leaf, root = _translate_merkle.add_leaf(canon)
    xtracer.trace("z3bridge.Translate HASH leaf=%s root=%s canon=%s" % (leaf, root, canon))
```

**Note**: Need to identify the exact 1:1 Python entry point matching Go's `Translator.Translate()`. Likely `formula_to_z3()` in `ivy_solver.py`.

## Files to modify

| File | Change |
|------|--------|
| `/Users/jaten/goivy/module/canon.go` | Add `Module.Canon()` method returning single `iu.Canonical` |
| `/Users/jaten/pyivy/ivy/ivy/ivy_module.py` | Add `Module.canon()` method returning single string |
| `/Users/jaten/goivy/compiler/compiler.go` | Extend `SigCheck` to hash Module alongside Sig |
| `/Users/jaten/pyivy/ivy/ivy/ivy_compiler.py` | Extend `sig_check` to hash module alongside sig |
| `/Users/jaten/goivy/z3bridge/xtrace.go` | Add `z3Merkle`, Merkle-chain `TraceCheck` output |
| `/Users/jaten/goivy/z3bridge/translate.go` | Add `TranslateMerkle` field, hash input in `Translate()` |
| `/Users/jaten/pyivy/ivy/ivy/ivy_solver.py` | Add `_z3_merkle` + `_translate_merkle`, Merkle-chain both `_trace_z3_check` and formula translation entry |

After Python changes, sync venv:
```bash
rsync -a ~/pyivy/ivy/ivy/ivy_module.py ~/pyivy/venv/lib/python3.10/site-packages/ivy/
rsync -a ~/pyivy/ivy/ivy/ivy_compiler.py ~/pyivy/venv/lib/python3.10/site-packages/ivy/
rsync -a ~/pyivy/ivy/ivy/ivy_solver.py ~/pyivy/venv/lib/python3.10/site-packages/ivy/
```

## Verification

```bash
cd ~/goivy && go build ./... && make golden
```

The golden test should show:
1. **SigCheck HASH** lines now contain both Sig and Module canon data — any Module-level divergence (wrong axiom, missing definition, extra schema) breaks the Merkle root immediately at the point of divergence.
2. **z3bridge.Translate HASH** lines appear at each formula-to-Z3 translation — any divergence in the logic IR expression sent to Z3 shows up as a leaf/root mismatch with canon= showing exactly which subexpression differs.
3. **z3.check HASH** lines replace the old z3.check traces — any divergence in accumulated solver state or check result propagates through the Merkle root, catching solver-level bugs even if individual assertions look correct.
