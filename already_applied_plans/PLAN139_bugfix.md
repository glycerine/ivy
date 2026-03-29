# Plan: Wire ProofChecker and Fix check_definitions Prover Reuse

**Created:** 2026-03-29 03:15

## Context

After fixing the `modifiesRec()` Apply→Func bug, ActionInterferenceCheck passes (146,153 matching lines). The `golden-nonstop` run reveals **445 divergent lines**, all caused by two root issues:

### Bug 1: `proof.RegisterFactories()` is never called

`proof/register.go` defines `RegisterFactories(modCfg, proofCfg)` which wires `NewProofCheckerFn` onto the module config. **No code ever calls it.** Result:
- `mod.Cfg.NewProofCheckerFn` is nil
- `CheckProperties` (phase6.go:2217) skips prover creation entirely
- The entire property-checking loop is a no-op — no axiom admission, no subgoal creation, no proof processing
- Go's `CheckProperties` runs in 4 xtrace lines (ENTER, ReorderProps ENTER/EXIT, EXIT). Python's runs in 650+.

### Bug 2: `check_definitions` creates fresh ProofChecker per definition call

Python (ivy_compiler.py:2028) creates ONE `ProofChecker(mod.labeled_axioms, [], mod.schemata)` before the SCC loop and reuses it — accumulating admitted definitions across iterations. Go's `wireAdmitDefinitionFactory` (check.go:43-50) creates a **new ProofChecker per call**, losing prior state.

Also: Python creates the ProofChecker **unconditionally** (even when there are zero recursive definitions), causing `normalize_goal()` → `clone_with_fresh_id()` → LF counter advancement on axioms/schemata. Go only creates provers when `AdmitDefinitionFn` is invoked for a recursive definition, which may be never.

### Impact

Python's two ProofChecker instances (one in check_definitions, one in check_properties) normalize all axioms/definitions/schemata, creating ~209 extra LabeledFormulas. Go creates zero. The resulting LF counter offset makes all subsequent traces diverge permanently.

## Plan

### Step 1: Call `proof.RegisterFactories()` during check setup

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/check/check.go`

At lines ~1164 and ~1235 (after `wireAdmitDefinitionFactory(mod)`), add:

```go
proof.RegisterFactories(mod.Cfg, proof.NewConfig())
```

Add `"github.com/glycerine/goivy/proof"` to the imports.

This wires `NewProofCheckerFn` and `GoalConcFn` so `CheckProperties` can create a prover and process properties.

### Step 2: Add `AdmitDefinition` to `ProofCheckerInterface`

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/module/proofapi.go`

The interface currently has `AdmitProposition`, `GetSubgoals`, `SetLastAxiom`, `SetSchema`. Add:

```go
AdmitDefinition(defn *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error)
```

The `proof/checker.go` `ProofChecker` already implements `AdmitDefinition` (line 401), so no change needed there.

### Step 3: Create shared ProofChecker in CheckDefinitions

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/ivy_compile.go` (~line 1587)

Replace the `mod.AdmitDefinitionFn` call in the SCC loop with a shared prover created via `NewProofCheckerFn`, matching Python's pattern:

```go
// Before SCC loop — matches Python line 2028:
// prover = ivy_proof.ProofChecker(mod.labeled_axioms, [], mod.schemata)
var defProver module.ProofCheckerInterface
if mod.Cfg != nil && mod.Cfg.NewProofCheckerFn != nil {
    schemataTyped := make(map[string]*ast.LabeledFormula)
    for k, v := range mod.Schemata {
        if lf, ok := v.(*ast.LabeledFormula); ok {
            schemataTyped[k] = lf
        }
    }
    defProver = mod.Cfg.NewProofCheckerFn(mod.LabeledAxioms, nil, schemataTyped)
}

// In SCC loop, replace:
//   if mod.AdmitDefinitionFn != nil { mod.AdmitDefinitionFn(d, proof) }
// with:
if defProver != nil {
    if _, err := defProver.AdmitDefinition(d, proof); err != nil {
        return err
    }
}
```

This creates ONE prover before the loop (matching Python), reuses it across definitions (accumulating state), and also ensures the prover is created unconditionally (matching Python's unconditional construction on line 2028).

### Step 4: Clean up `wireAdmitDefinitionFactory` / `AdmitDefinitionFn`

Since Step 3 replaces the factory pattern with direct prover usage, `wireAdmitDefinitionFactory` and `mod.AdmitDefinitionFn` become unused for check_definitions. Keep them for now (they may be used elsewhere) but the SCC loop no longer calls them.

## Critical Files

| File | Change |
|------|--------|
| `check/check.go` | Call `proof.RegisterFactories()` at ~lines 1164 and 1235; add `proof` import |
| `compiler/ivy_compile.go` | Create shared ProofChecker in CheckDefinitions SCC loop (~line 1587) |
| `module/proofapi.go` | Add `AdmitDefinition` to `ProofCheckerInterface` |

## Verification

1. `cd ~/goivy && make golden` — Check that ActionInterferenceCheck still passes AND that CheckProperties now shows work
2. `cd ~/goivy && make golden-nonstop &> nonstop.xtrace` — Count remaining divergences (should be dramatically reduced)
3. `grep -c "LF.__init__" out.go.xtrace out.py.xtrace` — LF counts should be close to matching
4. `grep "CheckProperties" out.go.xtrace` — Go should now show property processing between ENTER and EXIT (not just 4 lines)
