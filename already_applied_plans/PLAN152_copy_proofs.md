# Plan: Fix Module.Copy() Missing Proofs + Other Fields

**Created**: 2026-03-31 00:30, **Updated**: 2026-03-31 02:30

## Context

The golden test (`make golden`) diverges at line 152539 where Go emits `isolate.proofs n=0` but Python continues with `allSyms_pre_follow` symbols. The Vocab infrastructure (from prior work) is correct, but `mod.Proofs` is empty because `Module.Copy()` doesn't copy it.

**Root cause**: Python's `Module.copy()` (`ivy_module.py:229`) copies ALL attributes via:
```python
for x,y in self.__dict__.items():
    m.__dict__[x] = copy(y)
```

Go's `Module.Copy()` (`module/module.go:292-410`) explicitly lists fields to copy, but **misses several**, including the critical `Proofs` field. This means when `check/isolate_check.go:880` does `isoMod := mod.Copy()` before calling `CreateIsolate`, the copied module has `Proofs = nil`.

## Fix: Add Missing Copy Lines to `Module.Copy()`

**File**: `module/module.go`, inside the `Copy()` method (after line ~319, in the "Copy slices" section)

### Missing fields with Python equivalents

| Go Field | Python Field | Type | Copy Method |
|----------|-------------|------|-------------|
| `Proofs` | `self.proofs` | `[]ProofEntry` | `append([]ProofEntry{}, m.Proofs...)` |
| `Named` | `self.named` | `[]NamedEntry` | `append([]NamedEntry{}, m.Named...)` |
| `Subgoals` | `self.subgoals` | `[]SubgoalEntry` | `append([]SubgoalEntry{}, m.Subgoals...)` |
| `ConjSubgoals` | `self.conj_subgoals` | `[]*ast.LabeledFormula` | `copyLFSlice(m.ConjSubgoals)` |
| `ConceptSpaces` | `self.concept_spaces` | `[]ConceptSpace` | `append([]ConceptSpace{}, m.ConceptSpaces...)` |

### Missing Go-only fields (also need copying for correctness)

| Go Field | Type | Copy Method |
|----------|------|-------------|
| `VPrivates` | `map[string]bool` | `copyMapBool(m.VPrivates)` |
| `Macros` | `map[string]*ast.Definition` | shallow map copy |
| `CompCfg` | `*CompilerConfig` | shared pointer (just assign) |
| `CompileActionBodyFn` | `func(...)` | shared pointer (just assign) |
| `AdmitDefinitionFn` | `func(...)` | shared pointer (just assign) |
| `Instantiator` | `func(...)` | shared pointer (just assign) |
| `Theory` | `*co.Clauses` | shared pointer (just assign) |

Python copies `theory` when it exists (via `__dict__` iteration). The function pointers (`CompCfg`, `CompileActionBodyFn`, `AdmitDefinitionFn`, `Instantiator`) are Go-specific config threading — they should be shared (not deep-copied), matching how Python's `copy.copy()` does shallow copies.

### Exact code to add

Insert after line 319 (after `c.Delegates = ...`) in the "Copy slices" section:

```go
// Copy proofs, named, subgoals (Python copies all via __dict__ iteration)
c.Proofs = append([]ProofEntry{}, m.Proofs...)
c.Named = append([]NamedEntry{}, m.Named...)
c.Subgoals = append([]SubgoalEntry{}, m.Subgoals...)
c.ConjSubgoals = copyLFSlice(m.ConjSubgoals)
c.ConceptSpaces = append([]ConceptSpace{}, m.ConceptSpaces...)
```

Insert after line 386 (after `c.IsolateProofs = ...`) in the "Copy maps" section:

```go
// VPrivates: map[string]bool
if m.VPrivates != nil {
    c.VPrivates = copyMapBool(m.VPrivates)
}
// Macros: map[string]*ast.Definition
c.Macros = make(map[string]*ast.Definition, len(m.Macros))
for k, v := range m.Macros {
    c.Macros[k] = v
}
```

Insert after map copies, for shared pointers:

```go
// Shared pointers / callbacks (Python: copy.copy does shallow copy)
c.CompCfg = m.CompCfg
c.CompileActionBodyFn = m.CompileActionBodyFn
c.AdmitDefinitionFn = m.AdmitDefinitionFn
c.Instantiator = m.Instantiator
c.Theory = m.Theory
```

## Why This Fixes the Golden Test Divergence

1. Before fix: `mod.Copy()` → `isoMod.Proofs = nil` → `isolate.proofs n=0` → no proof symbols → missing `cf_pio_live.issued_pio` from `allSyms`
2. After fix: `mod.Copy()` → `isoMod.Proofs` has entries → Vocab infrastructure extracts symbol names → `cf_pio_live.issued_pio` added to `allSyms` → traces match Python

## Files Modified

| File | Changes |
|------|---------|
| `module/module.go` | Add ~15 lines to `Copy()` method for missing fields |

## Prior Work (already in working tree)

These were completed in the previous session and are already present:

- `ast/tactic.go`: VocabNames, Vocaber, VocabNode, IterSymbolsASTNode, 9 Vocab() methods
- `isolate/isolate.go`: Proof loop using `ast.VocabNode` + `allNames` InsMap + `allNamesMap` conversion

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...` — must compile
2. `cd ~/goivy && make test` — full test suite passes
3. `cd ~/goivy && make golden` — divergence at line 152539 should advance past `isolate.proofs n=0`
