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

## Issue 2: AdmitDefinition returns nil instead of [] for non-recursive (MINOR)

**Problem**: Python's `admit_definition` returns `[]` (empty list = success, no subgoals) for non-recursive definitions. Go returns `nil` (nil slice). Same pattern as the Issue 2 fix we already applied to `ApplyProof`.

```python
# Python (ivy_proof.py:93-95)
else:
    subgoals = []
self.definitions[sym.name] = defn
return subgoals  # returns [] for non-recursive
```

```go
// Go (checker.go:430-445)
var subgoals []*ast.LabeledFormula  // nil!
if recursive {
    ...
}
pc.Definitions[symSym.Name] = defn
return subgoals, nil  // returns (nil, nil) for non-recursive
```

**Fix**: `proof/checker.go:430` — initialize to empty slice:

```go
// Old:
var subgoals []*ast.LabeledFormula

// New:
subgoals := []*ast.LabeledFormula{}
```

This ensures the non-recursive path returns `([], nil)` matching Python's `[]`.

## Files Modified

| File | Changes |
|------|---------|
| `module/config.go:20` | Add `mod *Module` parameter to NewProofCheckerFn signature |
| `proof/checker.go:35` | Add `mod *module.Module` parameter to NewProofChecker; set `pc.Mod = mod` |
| `proof/checker.go:430` | Change `var subgoals` to `subgoals := []*ast.LabeledFormula{}` |
| `proof/register.go:13` | Pass `mod` through factory closure |
| `compiler/phase6.go:2277` | Pass `mod` to factory call |
| `compiler/ivy_compile.go:1601` | Pass `mod` to factory call |
| `check/check.go:46,303,412` | Pass `mod`/`m` to NewProofChecker |
| `check/isolate_check.go:68` | Pass `mod` to NewProofChecker |
| `proof/proof_test.go` | Add nil mod arg to NewProofChecker calls |
| `proof/checker_test.go` | Add nil mod arg to NewProofChecker calls |
| `check/check_port_test.go:299` | Add nil mod arg to NewProofChecker call |

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...` — must compile
2. `cd ~/go/src/github.com/glycerine/goivy && go test ./proof/... ./module/... ./check/... ./compiler/... ./tactics/...` — all tests pass
3. `cd ~/goivy && make golden` — golden test passes
