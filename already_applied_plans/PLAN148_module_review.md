# Plan: module/ Package Post-Refactor Review Fixes

**Created**: 2026-03-30 23:30

## Context

After refactoring `proof.Config`/`proof.Tactic` to `module.ProofConfig`/`module.ProofTactic` and adding `ProofCheckerInterface` to `module/proofapi.go`, we reviewed the module/ package and its interaction with proof/ against the Python source of truth (`~/pyivy/ivy/ivy/ivy_proof.py`).

## Verdict: module/ package types are correct; two bugs found in proof/ wiring

The types in module/proofapi.go (ProofCheckerInterface, ProofTactic, ProofConfig) are correctly defined and match Python semantics. The config.go wiring (ProofCfg field, TacticNewConfig in NewConfig) is correct. Two bugs were found in how proof/ uses these types.

## Issue 1: ProofChecker.Mod never set — nil deref in schema matching (CRITICAL)

**Problem**: `NewProofChecker` does not accept a `*module.Module` parameter. The `Mod` field defaults to nil. All production callers in check/ and compiler/ create ProofCheckers without ever setting Mod:

```go
// check/check.go:46, 303, 412; check/isolate_check.go:68
pc := proof.NewProofChecker(mod.Cfg.ProofCfg, pcAxioms, ...)
// pc.Mod is nil!
```

The factory in `proof/register.go` also creates ProofCheckers without Mod:
```go
return NewProofChecker(proofCfg, axioms, definitions, schemata, modCfg.AstCfg)
// Mod not set!
```

**Impact**: `CompileMatchList` (`proof/phase5_matching.go:601`) dereferences `mod.Cfg.AstCfg.NewDefinition(x, y)`. When `pc.Mod` is nil (which it always is in production), this panics for any schema proof with explicit match terms. The call chain:

```
MatchSchema → SetupMatching(goal, proof, pc.Mod=nil)
  → SetupSchemaMatching(..., mod=nil)
    → CompileMatchFull(..., mod=nil)
      → CompileMatchList(..., mod=nil)
        → mod.Cfg.AstCfg.NewDefinition(x, y)  ← PANIC
```

This is currently latent because:
- `getSigFrom(mod)` handles nil gracefully (returns fresh Sig)
- `CompileExprVocab(expr, vocab, mod)` handles nil mod (uses fresh Sig)
- The crash only triggers when `proofMatch` is non-empty (user provided explicit match terms in a SchemaInstantiation)

In Python, tactics/matching code accesses the module via the global `ivy_module.module` (set by context management). Go replaced globals with explicit threading — but the threading is incomplete here.

**Fix**: Thread `*module.Module` through the factory and constructor.

### Step 1a: Add Module parameter to NewProofCheckerFn factory signature

**File**: `module/config.go:20`

```go
// Old:
NewProofCheckerFn func(axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula) ProofCheckerInterface

// New:
NewProofCheckerFn func(mod *Module, axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula) ProofCheckerInterface
```

### Step 1b: Add Module parameter to NewProofChecker

**File**: `proof/checker.go:35`

```go
// Old:
func NewProofChecker(cfg *module.ProofConfig, axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula, astCfgs ...*ast.AstConfig) *ProofChecker {

// New — add mod parameter before astCfgs:
func NewProofChecker(cfg *module.ProofConfig, mod *module.Module, axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula, astCfgs ...*ast.AstConfig) *ProofChecker {
```

Inside the constructor, set `Mod: mod` in the struct literal (line 49).

### Step 1c: Update RegisterFactories closure

**File**: `proof/register.go:13-16`

```go
// Old:
modCfg.NewProofCheckerFn = func(axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula) module.ProofCheckerInterface {
    return NewProofChecker(proofCfg, axioms, definitions, schemata, modCfg.AstCfg)
}

// New:
modCfg.NewProofCheckerFn = func(mod *module.Module, axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula) module.ProofCheckerInterface {
    return NewProofChecker(proofCfg, mod, axioms, definitions, schemata, modCfg.AstCfg)
}
```

### Step 1d: Update all callers of NewProofCheckerFn (add mod argument)

| File:Line | Old | New |
|-----------|-----|-----|
| `compiler/phase6.go:2277` | `mod.Cfg.NewProofCheckerFn(mod.LabeledAxioms, mod.Definitions, schemataTyped)` | `mod.Cfg.NewProofCheckerFn(mod, mod.LabeledAxioms, mod.Definitions, schemataTyped)` |
| `compiler/ivy_compile.go:1601` | `mod.Cfg.NewProofCheckerFn(mod.LabeledAxioms, nil, schemataTyped)` | `mod.Cfg.NewProofCheckerFn(mod, mod.LabeledAxioms, nil, schemataTyped)` |

### Step 1e: Update all direct callers of NewProofChecker (add mod argument)

| File:Line | Old | New |
|-----------|-----|-----|
| `check/check.go:46` | `proof.NewProofChecker(m.Cfg.ProofCfg, m.LabeledAxioms, nil, typedSchemata)` | `proof.NewProofChecker(m.Cfg.ProofCfg, m, m.LabeledAxioms, nil, typedSchemata)` |
| `check/check.go:303` | `proof.NewProofChecker(mod.Cfg.ProofCfg, pcAxioms, mod.Definitions, ...)` | `proof.NewProofChecker(mod.Cfg.ProofCfg, mod, pcAxioms, mod.Definitions, ...)` |
| `check/check.go:412` | `proof.NewProofChecker(mod.Cfg.ProofCfg, pcAxioms, pcDefs, ...)` | `proof.NewProofChecker(mod.Cfg.ProofCfg, mod, pcAxioms, pcDefs, ...)` |
| `check/isolate_check.go:68` | `proof.NewProofChecker(mod.Cfg.ProofCfg, pcAxioms, mod.Definitions, ...)` | `proof.NewProofChecker(mod.Cfg.ProofCfg, mod, pcAxioms, mod.Definitions, ...)` |

### Step 1f: Update test callers

Tests that call `NewProofChecker(nil, ...)` or `NewProofChecker(cfg, ...)` need the new mod parameter:

| File | Old | New |
|------|-----|-----|
| `proof/proof_test.go` (multiple) | `NewProofChecker(nil, ...)` | `NewProofChecker(nil, nil, ...)` |
| `proof/checker_test.go` (multiple) | `NewProofChecker(nil, ...)` | `NewProofChecker(nil, nil, ...)` |
| `check/check_port_test.go:299` | `proof.NewProofChecker(nil, nil, nil, nil)` | `proof.NewProofChecker(nil, nil, nil, nil, nil)` |

Tests that construct ProofChecker via struct literal (e.g., `tactics/ivy_tactics_test.go:20`) already set `Mod: mod` and don't need changes.

## Issue 2: Callers use `== nil` to detect match failure instead of checking error (MINOR)

**Problem**: Several callers of `ApplyProof` use `subgoals == nil` to detect "match failed." In Go, nil and empty slices are functionally equivalent for `len()` and `range`. The proper signal for failure is the `error` return. Using `== nil` as a semantic signal creates fragility and forces unnecessary empty-slice allocations elsewhere.

**Affected sites** (all in `proof/checker.go`):

1. **Line 207-210** (ApplyProof empty-goals): Currently allocates `[]*ast.LabeledFormula{}` to avoid nil. Should return `nil, nil` — no allocation needed.

2. **Line 440** (AdmitDefinition, recursive case): `if subgoals == nil` after ApplyProof — redundant with `if err != nil` above it.

3. **Line 470** (AdmitProposition): `if subgoals == nil` after ApplyProof — redundant.

4. **Line 495** (GetSubgoals): `if subgoals == nil` after ApplyProof — redundant.

5. **Line 525** (composeProofs): `if decls == nil || len(decls) == 0` — the `decls == nil` check would misinterpret nil (success, no goals) as failure if returned from a caller that does `subgoals == nil`. Should use `len(decls) == 0` only.

**Fix**: Trust the error return. Remove nil-as-failure checks and revert the unnecessary allocation.

### Step 2a: Revert ApplyProof empty-goals to return nil

**File**: `proof/checker.go:207-210`

```go
// Old (our previous fix):
if len(goals) == 0 {
    return []*ast.LabeledFormula{}, nil
}

// New:
if len(goals) == 0 {
    return nil, nil
}
```

### Step 2b: Fix composeProofs nil check

**File**: `proof/checker.go:525`

```go
// Old:
if decls == nil || len(decls) == 0 {

// New:
if len(decls) == 0 {
```

### Step 2c: Remove redundant `subgoals == nil` checks

**File**: `proof/checker.go:440-442` (AdmitDefinition)

```go
// Remove these 3 lines:
if subgoals == nil {
    return nil, &NoMatch{Node: defn, Msg: "recursive definition does not match the given schema"}
}
```

All failure paths in ApplyProof already return errors. If `err == nil`, the result is valid regardless of nil/empty.

**File**: `proof/checker.go:470-472` (AdmitProposition)

```go
// Remove these 3 lines:
if subgoals == nil {
    return nil, &NoMatch{Node: proof, Msg: "goal does not match the given schema"}
}
```

**File**: `proof/checker.go:495-497` (GetSubgoals)

```go
// Remove these 3 lines:
if subgoals == nil {
    return nil, &NoMatch{Node: proof, Msg: "goal does not match the given schema"}
}
```

## Files Modified

| File | Changes |
|------|---------|
| `module/config.go:20` | Issue 1: Add `mod *Module` parameter to NewProofCheckerFn signature |
| `proof/checker.go:35` | Issue 1: Add `mod *module.Module` parameter to NewProofChecker; set `pc.Mod = mod` |
| `proof/checker.go:207-210` | Issue 2: Revert ApplyProof empty-goals to `return nil, nil` |
| `proof/checker.go:440-442` | Issue 2: Remove `if subgoals == nil` in AdmitDefinition |
| `proof/checker.go:470-472` | Issue 2: Remove `if subgoals == nil` in AdmitProposition |
| `proof/checker.go:495-497` | Issue 2: Remove `if subgoals == nil` in GetSubgoals |
| `proof/checker.go:525` | Issue 2: Change `decls == nil \|\| len(decls) == 0` to `len(decls) == 0` in composeProofs |
| `proof/register.go:13` | Issue 1: Pass `mod` through factory closure |
| `compiler/phase6.go:2277` | Issue 1: Pass `mod` to factory call |
| `compiler/ivy_compile.go:1601` | Issue 1: Pass `mod` to factory call |
| `check/check.go:46,303,412` | Issue 1: Pass `mod`/`m` to NewProofChecker |
| `check/isolate_check.go:68` | Issue 1: Pass `mod` to NewProofChecker |
| `proof/proof_test.go` | Issue 1: Add nil mod arg to NewProofChecker calls |
| `proof/checker_test.go` | Issue 1: Add nil mod arg to NewProofChecker calls |
| `check/check_port_test.go:299` | Issue 1: Add nil mod arg to NewProofChecker call |

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...` — must compile
2. `cd ~/go/src/github.com/glycerine/goivy && go test ./proof/... ./module/... ./check/... ./compiler/... ./tactics/...` — all tests pass
3. `cd ~/goivy && make golden` — golden test passes
