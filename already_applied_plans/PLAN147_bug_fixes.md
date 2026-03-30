# Plan: Post-Refactor Code Review — proof/ Package

**Created**: 2026-03-30 20:15

## Context

After moving `proof.Config`/`proof.Tactic` to `module.ProofConfig`/`module.ProofTactic`, we reviewed the `proof/` package against the Python source of truth (`~/pyivy/ivy/ivy/ivy_proof.py`) to check for bugs introduced by the refactor and any pre-existing divergences.

## Verdict: Refactor introduced NO bugs

The refactor was clean: all changes were type renames, signature widenings (concrete → interface), and trivially correct getter methods. No logic was altered.

## Pre-existing issues found during review

### Issue 1: Un-wired RegisterTactics calls (CRITICAL)

**Problem**: `check.RegisterTactics`, `temporal.RegisterTactics`, and `l2s.RegisterTactics` are **defined but never called**. No tactics are ever registered on any ProofConfig. This means all `TacticTactic` proof nodes fail with "unknown tactic".

In Python, tactic registration happens at module import time (top-level `pf.register_tactic(...)` calls). In Go, the equivalent `RegisterTactics()` functions exist but have no callers.

**Affected tactics (all dead-registered)**:
- `vcgen`, `skolemize`, `skolemizenp`, `tempind`, `tempcase`, `sorry` (from `tactics.RegisterProofTactics`)
- `mc`, `vmt` (from `check.RegisterTactics`)
- `invariance` (from `temporal.RegisterTactics`)
- `l2s`, `l2s_full`, `l2s_auto`, `l2s_auto2`..`5` (from `l2s.RegisterTactics`)

**Fix**: Wire `check.RegisterTactics(mod.Cfg.ProofCfg, mod)` in both `Start()` (line ~1165) and `StartWithConfig()` (line ~1237), after `proof.RegisterFactories`. Also wire `temporal.RegisterTactics` and `l2s.RegisterTactics` from within `check.RegisterTactics`.

**Files**:
- `check/check.go:1153` (`Start`) — add call after line 1165
- `check/check.go:1223` (`StartWithConfig`) — add call after line 1237
- `check/check.go:1140` (`RegisterTactics`) — add calls to `temporal.RegisterTactics(proofCfg)` and `l2s.RegisterTactics(proofCfg)`

### Issue 2: ApplyProof empty-goals returns nil (MINOR)

**Problem**: Python `apply_proof` returns `[]` (empty list, distinct from `None` = "match failed") when `len(decls) == 0`. Go returns `nil, nil`, which conflates "no goals remain" with "match failed" since callers check `subgoals == nil`.

**Fix**: `checker.go:215-217` — change `return nil, nil` to `return []*ast.LabeledFormula{}, nil`.

### Issue 3: GetSubgoals normalize-before-assert ordering (MINOR)

**Problem**: Python asserts `not isinstance(prop.formula, il.Definition)` BEFORE normalizing. Go normalizes FIRST, then checks. If normalization could change the formula type, the ordering matters.

**Fix**: `checker.go:491-496` — swap: check Definition type before calling NormalizeGoal.

### Issue 4: Stale symbols misses premise symbols (MINOR)

**Problem**: In `NewProofChecker`, Go calls `GoalConc(lf)` then `UsedSymbolsAST(conc)` — only processes the conclusion. Python calls `used_symbols_ast(lf.formula)` on the whole formula, including premises in SchemaBody.

**Fix**: `checker.go:93-108` — use `lf.Formula` (the whole formula, not just conclusion) as input to `UsedSymbolsAST`, matching Python.

### Issue 5: Simplified tactic implementations (MAJOR, separate effort)

The following tactics are dramatically simplified vs Python and are missing critical logic. **Not in scope for this fix plan** — they need their own dedicated porting effort:

| Tactic | Missing from Go |
|--------|----------------|
| `assumeTactic` | `setup_schema_matching`, witness separation, `close_unmatched`, `apply_match_goal`, `drop_supplied_prems`, `AssumeGlobalTactic` handling, premise removal |
| `propertyTactic` | `compile_expr_vocab`, `goal_subst`, existential witness handling, `ConstantDecl` premise |
| `functionTactic` | `goal_vocab`, `goal_free`, `TopFunctionSort`, `WithSymbols`, `ForAll` wrapping, recursive check, dual `goal_add_prem` |
| `unfoldTactic` | `has_premise` path, `goal_apply_to_prem`, rename support via `unfspec.renamings` |
| `letTactic` | `compile_expr_vocab`, `goal_vocab`, `attrib_goals` |

## Implementation Steps (Issues 1-4 only)

### Step 1: Wire RegisterTactics in check/check.go

In `Start()`, after `proof.RegisterFactories(...)`:
```go
check.RegisterTactics(mod.Cfg.ProofCfg, mod)
```

In `StartWithConfig()`, after `proof.RegisterFactories(...)`:
```go
RegisterTactics(mod.Cfg.ProofCfg, mod)
```

In `RegisterTactics`, add temporal and l2s registration:
```go
func RegisterTactics(proofCfg *module.ProofConfig, mod *module.Module) {
    proofCfg.RegisterTactic("mc", ...)
    proofCfg.RegisterTactic("vmt", ...)
    tactics.RegisterProofTactics(proofCfg)
    temporal.RegisterTactics(proofCfg)  // ADD
    l2s.RegisterTactics(proofCfg)       // ADD
}
```

### Step 2: Fix ApplyProof empty-goals return

```go
// checker.go:215-217
if len(goals) == 0 {
    return []*ast.LabeledFormula{}, nil  // was: return nil, nil
}
```

### Step 3: Fix GetSubgoals ordering

```go
// checker.go:491-496
func (pc *ProofChecker) GetSubgoals(prop *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error) {
    // Python: assert not isinstance(prop.formula, il.Definition)  — BEFORE normalize
    if _, isDef := prop.Formula.(*lg.Definition); isDef {
        return nil, &ProofError{Msg: "GetSubgoals: prop may not be a definition"}
    }
    prop = NormalizeGoal(pc.astCfg(), prop)
    ...
```

### Step 4: Fix stale symbols to include premise symbols

```go
// checker.go:93-108 — change GoalConc(lf) to lf.Formula for stale tracking
for _, lf := range axioms {
    if lf.Formula != nil {
        if fmla, ok := lf.Formula.(lg.Expr); ok {
            for _, c := range clauseops.UsedSymbolsAST(fmla) {
                pc.Stale[c.Name] = true
            }
        }
    }
}
// same for definitions loop
```

## Verification

1. `cd ~/goivy && go build ./...` — must compile
2. `cd ~/goivy && make test` — all tests pass
3. `cd ~/goivy && make golden` — golden test passes
4. Verify registered tactics: add a test that creates a module via `Start`-like flow and checks that `mod.Cfg.ProofCfg.Tactics` contains expected keys

## Previous Plan (completed)

The original plan to move proof.Config/Tactic to module.ProofConfig/ProofTactic has been fully implemented and verified. See git history for details.

## Key Design Decision: ProofChecker stays in proof/

Moving ProofChecker to module/ would require moving all 24 methods across 3 files (checker.go, matching.go, tactics.go) — essentially the entire proof package. Instead, we:
- Move only `Config` and `Tactic` to module/
- Expand `ProofCheckerInterface` with getter methods
- Change `ProofTactic` signature to use `ProofCheckerInterface` instead of `*ProofChecker`

## Implementation Steps

### Step 1: Expand ProofCheckerInterface (`module/proofapi.go`)

Add getter methods that external tactic implementations need:

```go
type ProofCheckerInterface interface {
    // existing methods...
    AdmitDefinition(defn *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error)
    AdmitProposition(prop *ast.LabeledFormula, proof ast.Node, existingSubgoals ...*ast.LabeledFormula) ([]*ast.LabeledFormula, error)
    GetSubgoals(prop *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error)
    SetLastAxiom(prop *ast.LabeledFormula)
    SetSchema(name string, prop *ast.LabeledFormula)
    // NEW getters:
    GetModule() *Module
    GetAstCfg() *ast.AstConfig
    GetAxioms() []*ast.LabeledFormula
}
```

### Step 2: Add new types to `module/proofapi.go`

```go
// ProofTactic is a function that applies a proof tactic to a goal.
type ProofTactic func(checker ProofCheckerInterface, goals []*ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error)

// ProofConfig holds per-session proof state (tactic registry).
type ProofConfig struct {
    Tactics map[string]ProofTactic
}

// TacticNewConfig creates a new ProofConfig with an empty tactic registry.
func TacticNewConfig() *ProofConfig {
    return &ProofConfig{Tactics: make(map[string]ProofTactic)}
}

// RegisterTactic registers a named tactic on this config.
func (cfg *ProofConfig) RegisterTactic(name string, t ProofTactic) {
    cfg.Tactics[name] = t
}
```

### Step 3: Add `ProofCfg` field to `module.Config` (`module/config.go`)

```go
type Config struct {
    // ... existing fields ...
    ProofCfg *ProofConfig `json:"-"`
}
```

Wire it in `NewConfig()`:
```go
func NewConfig() *Config {
    // ... existing ...
    return &Config{
        // ... existing fields ...
        ProofCfg: TacticNewConfig(),
    }
}
```

### Step 4: Implement interface getters on `proof.ProofChecker` (`proof/checker.go`)

```go
func (pc *ProofChecker) GetModule() *module.Module       { return pc.Mod }
func (pc *ProofChecker) GetAstCfg() *ast.AstConfig       { return pc.AstCfg }
func (pc *ProofChecker) GetAxioms() []*ast.LabeledFormula { return pc.Axioms }
```

### Step 5: Update ProofChecker to use module.ProofConfig (`proof/checker.go`)

- Change field: `Cfg *Config` → `Cfg *module.ProofConfig`
- Change constructor param: `NewProofChecker(cfg *Config, ...)` → `NewProofChecker(cfg *module.ProofConfig, ...)`
- Change nil fallback: `cfg = NewConfig()` → `cfg = module.TacticNewConfig()`
- In `tacticTactic`: `tactic(pc, decls, proof)` still works — `*ProofChecker` satisfies `ProofCheckerInterface`
- **Delete** old `type Tactic`, `type Config`, `func NewConfig()`, `func (cfg *Config) RegisterTactic()`

### Step 6: Update `proof/register.go`

```go
func RegisterFactories(modCfg *module.Config, proofCfg *module.ProofConfig) {
    modCfg.ProofCfg = proofCfg  // Store on module config
    modCfg.NewProofCheckerFn = func(axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula) module.ProofCheckerInterface {
        return NewProofChecker(proofCfg, axioms, definitions, schemata, modCfg.AstCfg)
    }
    modCfg.GoalConcFn = func(g *ast.LabeledFormula) lg.Expr {
        return GoalConc(g)
    }
}
```

### Step 7: Update external tactic function signatures

**7a. `tactics/ivy_tactics.go`** (6 tactic funcs + 1 helper + registration)

| Old | New |
|-----|-----|
| `pcAstCfg(pc *proof.ProofChecker)` | `pcAstCfg(pc module.ProofCheckerInterface)` using `pc.GetAstCfg()` — also fixes infinite recursion bug |
| `Vcgen(pc *proof.ProofChecker, ...)` | `Vcgen(pc module.ProofCheckerInterface, ...)` |
| `Skolemize(pc *proof.ProofChecker, ...)` | `Skolemize(pc module.ProofCheckerInterface, ...)` |
| `Skolemizenp(pc *proof.ProofChecker, ...)` | `Skolemizenp(pc module.ProofCheckerInterface, ...)` |
| `Tempind(pc *proof.ProofChecker, ...)` | `Tempind(pc module.ProofCheckerInterface, ...)` |
| `Tempcase(pc *proof.ProofChecker, ...)` | `Tempcase(pc module.ProofCheckerInterface, ...)` |
| `Sorry(pc *proof.ProofChecker, ...)` → uses `pc.Mod.Cfg.UsedSorry` | `Sorry(pc module.ProofCheckerInterface, ...)` → `pc.GetModule().Cfg.UsedSorry` |
| `RegisterProofTactics(cfg *proof.Config)` | `RegisterProofTactics(cfg *module.ProofConfig)` |
| `&proof.Config{Tactics: make(map[string]proof.Tactic)}` (test) | `&module.ProofConfig{Tactics: make(map[string]module.ProofTactic)}` |

**7b. `temporal/temporal.go`** (1 tactic func + registration)

| Old | New |
|-----|-----|
| `InvarianceTactic(pc *proof.ProofChecker, ...)` | `InvarianceTactic(pc module.ProofCheckerInterface, ...)` |
| Access `pc.Mod`, `pc.AstCfg`, `pc.Axioms` | Use `pc.GetModule()`, `pc.GetAstCfg()`, `pc.GetAxioms()` |
| `RegisterTactics(proofCfg *proof.Config)` | `RegisterTactics(proofCfg *module.ProofConfig)` |

**7c. `l2s/l2s.go`** (3 entry + 1 internal func + registration)

| Old | New |
|-----|-----|
| `L2STactic(pc *proof.ProofChecker, ...)` | `L2STactic(pc module.ProofCheckerInterface, ...)` |
| `L2STacticFull(pc *proof.ProofChecker, ...)` | `L2STacticFull(pc module.ProofCheckerInterface, ...)` |
| `L2STacticAuto(pc *proof.ProofChecker, ...)` | `L2STacticAuto(pc module.ProofCheckerInterface, ...)` |
| `l2sTacticInt(pc *proof.ProofChecker, ...)` | `l2sTacticInt(pc module.ProofCheckerInterface, ...)` |
| Access `pc.Mod`, `pc.AstCfg`, `pc.Axioms` | Use `pc.GetModule()`, `pc.GetAstCfg()`, `pc.GetAxioms()` |
| `RegisterTactics(proofCfg *proof.Config)` | `RegisterTactics(proofCfg *module.ProofConfig)` |

**7d. `check/check.go`** (closures + registration + callers)

| Old | New |
|-----|-----|
| `RegisterTactics(proofCfg *proof.Config, mod *module.Module)` | `RegisterTactics(proofCfg *module.ProofConfig, mod *module.Module)` |
| Closure: `func(pc *proof.ProofChecker, ...)` | `func(pc module.ProofCheckerInterface, ...)` |
| `proof.NewConfig()` at lines 1165, 1237 | `module.TacticNewConfig()` |
| `proof.RegisterFactories(mod.Cfg, proof.NewConfig())` | `proof.RegisterFactories(mod.Cfg, module.TacticNewConfig())` |
| Line 1101: `pc, _ := prover.(*proof.ProofChecker)` | `pc, _ := prover.(module.ProofCheckerInterface)` |
| Lines 1103, 1107: `tactics.Tempind(pc, ...)` | No change needed — pc is now ProofCheckerInterface, matching new signatures |

### Step 8: Fix nil Config call sites (the whole point!)

All `proof.NewProofChecker(nil, ...)` become `proof.NewProofChecker(mod.Cfg.ProofCfg, ...)`:

| File:Line | Old | New |
|-----------|-----|-----|
| `check/isolate_check.go:68` | `proof.NewProofChecker(nil, pcAxioms, ...)` | `proof.NewProofChecker(mod.Cfg.ProofCfg, pcAxioms, ...)` |
| `check/check.go:46` | `proof.NewProofChecker(nil, m.LabeledAxioms, ...)` | `proof.NewProofChecker(mod.Cfg.ProofCfg, m.LabeledAxioms, ...)` — note: need to thread `mod` into `wireAdmitDefinitionFactory` |
| `check/check.go:303` | `proof.NewProofChecker(nil, pcAxioms, ...)` | `proof.NewProofChecker(mod.Cfg.ProofCfg, pcAxioms, ...)` |
| `check/check.go:412` | `proof.NewProofChecker(nil, pcAxioms, ...)` | `proof.NewProofChecker(mod.Cfg.ProofCfg, pcAxioms, ...)` |

### Step 9: Update test files

| File | Change |
|------|--------|
| `proof/isolation_test.go` | `NewConfig()` → `module.TacticNewConfig()`, add `module` import |
| `proof/proof_test.go` | `NewConfig()` → `module.TacticNewConfig()`, `RegisterTactic` stays same |
| `proof/checker_test.go` | `NewProofChecker(nil, ...)` → `NewProofChecker(module.TacticNewConfig(), ...)` or keep nil (still handled) |
| `tactics/ivy_tactics_test.go:17-23` | `*proof.ProofChecker` → keep (struct literal still valid), but field `Cfg` needs `module.ProofConfig` |
| `tactics/ivy_tactics_test.go:297` | `&proof.Config{...}` → `&module.ProofConfig{Tactics: make(map[string]module.ProofTactic)}` |
| `check/check_port_test.go:299` | `proof.NewProofChecker(nil, ...)` → `proof.NewProofChecker(module.TacticNewConfig(), ...)` |

### Step 10: Update imports

Packages that used `proof` ONLY for Config/Tactic types may be able to drop the `proof` import. However, most still need `proof` for helper functions (`proof.GoalConc`, `proof.MakeGoal`, etc.) or `proof.NewProofChecker`.

Packages that can potentially drop `proof` import:
- None clearly — all packages that used `proof.Config` also use other proof exports.

Packages that need NEW `module` import (if not already present):
- `proof/isolation_test.go` — add `"github.com/glycerine/goivy/module"`

## Files Modified (Summary)

| File | Changes |
|------|---------|
| `module/proofapi.go` | Add ProofTactic, ProofConfig, TacticNewConfig, RegisterTactic; expand ProofCheckerInterface |
| `module/config.go` | Add ProofCfg field to Config; wire in NewConfig() |
| `proof/checker.go` | Delete old types; change Cfg field type; add 3 getter methods; update NewProofChecker param |
| `proof/register.go` | Change param type; store ProofCfg on modCfg |
| `proof/isolation_test.go` | Use module.TacticNewConfig() |
| `proof/proof_test.go` | Use module.TacticNewConfig() |
| `proof/checker_test.go` | Minor updates |
| `tactics/ivy_tactics.go` | Change all 6 tactic sigs + pcAstCfg + RegisterProofTactics param; fix recursion bug |
| `tactics/ivy_tactics_test.go` | Update Config/Tactic references |
| `temporal/temporal.go` | Change InvarianceTactic sig + RegisterTactics param; use getters |
| `l2s/l2s.go` | Change 4 tactic sigs + RegisterTactics param; use getters |
| `check/check.go` | Change RegisterTactics param + closures + NewConfig calls + line 1101 type assert |
| `check/isolate_check.go` | Pass mod.Cfg.ProofCfg instead of nil |
| `check/check_port_test.go` | Update NewProofChecker call |

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...` — must compile clean
2. `cd ~/go/src/github.com/glycerine/goivy && go test ./proof/... ./tactics/... ./check/... ./temporal/... ./l2s/...` — all tests pass
3. `cd ~/goivy && make golden` — golden test still passes (no behavioral change)
4. Verify no remaining references to `proof.Config`, `proof.Tactic`, or `proof.NewConfig` (except in already_applied_plans/ docs)
