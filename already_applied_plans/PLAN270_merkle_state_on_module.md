# Plan: Share Merkle state across compiler instances via Module

Created: 2026-04-12 ~07:20 UTC
Updated: 2026-04-12 ~10:45 UTC

## Prior fix (DONE)

TopSortAsDefault fix in `proof/goal.go:574-577` — already applied and building. Moved divergence from 258992 to 258998.

## Context

Golden test diverges at xtrace line 258998. Traces match through 258997, then:

```
258998  go : compiler.SigCheck@SortifyWithInference HASH leaf=blake3.33B-2Lce...  root=blake3.33B-QXxa...
        py : compiler.SigCheck@SortifyWithInference HASH leaf=blake3.33B-2Lce...  root=blake3.33B-UqT6...
```

**Leaf hashes match** (same sig+module canon) but **root hashes differ** (different Merkle chain history).

## Root Cause

**Python** (`ivy_compiler.py:38`): `sig_merkle = xtracer.MerkleState()` is MODULE-LEVEL — one chain for the entire compilation session. All `sig_check()` calls (DomainSetup, IvyCompileTheory, SortifyWithInference, CompileDefnImpl, etc.) accumulate into the same chain.

**Go** (`compiler/compiler.go:142`): `SigMerkle iu.MerkleState` is PER-COMPILER. When `compileExprVocabLF` (`proof/goal.go:584`) creates `compiler.New(sig, mod)`, it gets a fresh Merkle chain with `PrevRoot=""`. So:
- Go: `root = hash("" + leaf)` (fresh chain)
- Python: `root = hash(accumulated_prev_root + leaf)` (session chain)

The leaf matches because the sig+module state is identical. The root differs because Go lost the history of all prior SigCheck calls from earlier compilation phases (DomainSetup, IvyCompileTheory, etc.).

## Fix

Move the Merkle state from the Compiler struct to the Module, which persists across the entire compilation session and is passed to all compiler instances.

### Step 1: Add SigMerkle field to Module

**File:** `module/module.go`

In the Module struct (after line ~110, near `CompCfg`):
```go
// SigMerkle is the rolling Merkle hash for compiler conformance auditing.
// Matches Python's module-level sig_merkle in ivy_compiler.py.
// Lives on Module so all Compiler instances share one chain per session.
SigMerkle *iu.MerkleState
```

In `New()` (after `m.Clear()`, before `return m`):
```go
m.SigMerkle = &iu.MerkleState{}
```

No change to `Clear()` — the Merkle state must NOT be reset mid-session. `Clear()` is only called from `New()` anyway.

### Step 2: Wire Compiler to use Module's Merkle state

**File:** `compiler/compiler.go`

Change the field type (line 142):
```go
SigMerkle *iu.MerkleState  // was: iu.MerkleState (value)
```

In `New()` (after mod is ensured non-nil, ~line 185), add:
```go
if c.Module.SigMerkle != nil {
    c.SigMerkle = c.Module.SigMerkle
} else {
    c.SigMerkle = &iu.MerkleState{}
}
```

`SigCheck` (line 165) already calls `c.SigMerkle.AddLeaf(combined)` — pointer receiver on `*MerkleState`, works unchanged.

### Files to modify

1. `module/module.go` — add `SigMerkle` field + init in `New()`
2. `compiler/compiler.go` — change `SigMerkle` to pointer, wire from Module in `New()`

No import changes needed — `module` already imports `iu "github.com/glycerine/ivy/goivy/ivyutils"`.

## Verification

1. `go build ./...` — must compile
2. Run `TestOrdLive` — check that line 258998 root hashes now match (same Merkle chain), and the divergence moves forward or the test passes
