# Plan: Bug Fixes, init() Elimination, and GoIvyConfig

## jea: I discarded this. Too many corrections to the GoIvyConfig were needed

Instead, just did BUGFIXES.md first.

## Context

Code review of CheckProperties (proof/checker.go, compiler/phase6.go) revealed 7 bugs. Additionally, Z3 requires single-threaded access per context, so all `init()`-based registration must move to explicit setup called after Go's init phase completes. A `GoIvyConfig` struct will hold all mutable global state to enable running a thread pool of goivy instances.

---

## Part A: Bug Fixes (immediate)

### BUG 1 (HIGH): Wrong key for Definitions map
**File**: `proof/checker.go:59`
**Problem**: Uses `d.LabelName()` instead of `d.Formula.Defines().Name` for definition map key.
**Fix**:
```go
name := ""
if def, ok := d.Formula.(*lg.Definition); ok {
    if sym, ok := def.Defines().(*lg.Symbol); ok {
        name = sym.Name
    }
}
if name == "" {
    name = d.LabelName()
}
pc.Definitions[name] = norm
```

### BUG 3 (HIGH): Type assertion panic on prop.Label
**File**: `compiler/phase6.go:1975`
**Fix**: Safe type assertion before `ComposeAtoms`.

### BUG 4 (MEDIUM): AdmitProposition missing subgoals parameter
**File**: `proof/checker.go:368`
**Fix**: Add variadic `existingSubgoals ...*ast.LabeledFormula`. Update `ProofCheckerInterface` to match.

### BUG 5 (MEDIUM): GetSubgoals missing definition assertion
**File**: `proof/checker.go:397`
**Fix**: Add `if _, isDef := prop.Formula.(*lg.Definition); isDef { return error }` before normalize.

### BUG 6 (MEDIUM): Errors from AdmitProposition silently ignored
**File**: `compiler/phase6.go:2011,2016`
**Fix**: Log errors with `pp()`.

### BUG 7 (LOW): sga.Kind vs sga.SubgoalKind
**File**: `compiler/phase6.go:1811`
**Fix**: Change `sga.Kind = a.Kind` to `sga.SubgoalKind = a.Name()`.

---

## Part B: Eliminate init() Registration (immediate)

### Problem

These `init()` functions register factories/tactics into global mutable maps. Z3's thread-local storage means all goivy work for a given Z3 context must happen on one goroutine. init() runs before main() and can't be controlled.

### init() Functions to Eliminate (7 registration inits)

| File | Line | What it registers |
|------|------|-------------------|
| `proof/register.go` | 9 | `compiler.NewProofCheckerFn`, `compiler.GoalConcFn` |
| `check/check.go` | 1021 | `proof.RegisterTactic("mc", ...)`, `proof.RegisterTactic("vmt", ...)` |
| `l2s/l2s.go` | 831 | 7 l2s tactics into `proof.RegisteredTactics` |
| `temporal/temporal.go` | 656 | `proof.RegisterTactic("invariance", ...)` |
| `actions/annotation.go` | 11 | `co.AnnotConjFunc` callback |
| `actions/action.go` | 1166 | `GlobalContext = &ActionContext{}` |
| `acl/acl.go` | 22 | Initializes `ignores`/`assumes` maps |

### init() Functions to KEEP (benign, no mutable verification state)

| File | Line | Why keep |
|------|------|----------|
| `ivylogic/poly.go` | 60 | Builds read-only `polymorphicSymbols` map from static data. No Z3, no cross-goroutine issue. |
| `compiler/vprint.go` | 31 | Timezone/debug setup — infrastructure, not verification state |
| `z3bridge/vprint.go` | 31 | Same |
| `parser/vprint.go` | 31 | Same |
| `end2end/vprint.go` | 31 | Same |
| `webui/vprint.go` | 31 | Same |
| `webui/ext_api.go` | 135 | UI extension points — not verification |

### Approach: Move ProofCheckerInterface to `module` package, add Setup() functions

**Step B1**: Create `module/proofapi.go` — define `ProofCheckerInterface`, `NewProofCheckerFn`, `GoalConcFn` vars here (both `proof` and `compiler` already import `module`). This replaces BUG 2's original fix.

**Step B2**: Each package with a registration init() gets an exported `Setup()` or `RegisterTactics()` function containing the same code that was in init(). The init() body is deleted.

| Package | New function | Replaces init() at |
|---------|-------------|-------------------|
| `proof` | `RegisterFactories()` | `register.go:9` — sets `module.NewProofCheckerFn`, `module.GoalConcFn` |
| `check` | `RegisterTactics()` | `check.go:1021` — registers mc/vmt tactics |
| `l2s` | `RegisterTactics()` | `l2s.go:831` — registers l2s tactics |
| `temporal` | `RegisterTactics()` | `temporal.go:656` — registers invariance tactic |
| `actions` | `InitAnnotConj()` | `annotation.go:11` — sets `co.AnnotConjFunc` |
| `actions` | `InitGlobalContext()` | `action.go:1166` — sets `GlobalContext` |
| `acl` | `InitACL()` | `acl.go:22` — initializes maps |

**Step B3**: Create orchestrator that calls all Setup functions in order. This goes in an existing entry point (e.g. `check.Start()` or a new `GoIvySetup()` function in the `check` package, since `check` already imports all these packages).

```go
// check/setup.go (or similar)
func GoIvySetup() {
    acl.InitACL()
    actions.InitAnnotConj()
    actions.InitGlobalContext()
    proof.RegisterFactories()    // sets module.NewProofCheckerFn, module.GoalConcFn
    check.RegisterTactics()      // mc, vmt
    l2s.RegisterTactics()        // l2s, l2s_full, l2s_auto, etc.
    temporal.RegisterTactics()   // invariance
}
```

**Step B4**: Update `compiler/phase6.go` to read `module.NewProofCheckerFn` / `module.GoalConcFn` instead of local vars. Remove the local `ProofCheckerInterface`, `NewProofCheckerFn`, `GoalConcFn` declarations from phase6.go.

**Step B5**: Update all tests that depend on these registrations to call `GoIvySetup()` in TestMain or test init.

---

## Part C: GoIvyConfig Struct (action item — phased)

### Goal
Encapsulate all mutable global state into a `GoIvyConfig` struct so multiple goivy instances can run concurrently in a thread pool, each with its own config (and its own Z3 context pinned to one goroutine).

### Where it lives: `ivyutils/config.go`
`ivyutils` is a leaf package imported by everything. No new import edges.

### What goes in GoIvyConfig

**Group 1 — Factory functions** (currently in module after Part B):
- `NewProofCheckerFn` (interface{} to avoid cycle)
- `GoalConcFn` (interface{})

**Group 2 — Registries** (currently package-level maps):
- `RegisteredTactics` (from proof/checker.go — map[string]interface{})
- `AnnotConjFunc` (from clauseops)

**Group 3 — Compiler state** (currently package-level vars in compiler):
- `PropIDCounter` int64
- `OptionVerifying` bool
- `OptMutax` *BooleanParameter

**Group 4 — Check state** (currently package-level in check):
- `Failures` int
- `CheckedActionFound` bool
- `CheckLineno` string
- `Diagnose`, `Coverage`, `OptTrusted`, `OptMC`, `OptTrace` *BooleanParameter
- `CheckedAction` *Parameter

**Group 5 — ivylogic state**:
- `DefaultSort` interface{} (lg.Sort)
- `AllowUnsorted` bool
- `ReasonText` string

**Group 6 — Module/tactic state**:
- `L2SDebug` bool
- `L2SCurrentModule` interface{}
- `RankingDebug` bool
- `GlobalContext` interface{} (actions.IActionContext)

**Group 7 — ACL state** (already mutex-protected):
- `Ignores`, `Assumes` maps
- `IgnoresRegex`, `AssumesRegex`

### Migration Strategy (phased, one package at a time)

**Phase C1**: Create `ivyutils/config.go` with `GoIvyConfig` struct and `NewGoIvyConfig()`.

**Phase C2**: Add `Config *iu.GoIvyConfig` field to key structs:
- `compiler.Compiler`
- `proof.ProofChecker`
- `module.Module`

**Phase C3**: Migrate one package at a time (compiler → proof → check → ivylogic → l2s → temporal → acl → actions → ranking). For each:
1. Change functions to read `cfg.Field` instead of package global
2. Thread `cfg` through call chains
3. Remove the package-level global var
4. Verify with `make test`

**Phase C4**: Update `GoIvySetup()` to accept a `*GoIvyConfig` parameter and populate it. Each goroutine in a thread pool creates its own config:
```go
cfg := iu.NewGoIvyConfig()
check.GoIvySetup(cfg)
// run verification on this goroutine with cfg
```

### Note on interface{} fields
Many GoIvyConfig fields must be `interface{}` because `ivyutils` can't import `logic`, `ast`, `module`, etc. Each consuming package provides typed accessor helpers that do the type assertion.

---

## Files to Modify (Parts A + B)

| File | Changes |
|------|---------|
| `proof/checker.go` | BUG 1 (line 59), BUG 4 (line 368 variadic), BUG 5 (line 397 assertion) |
| `module/proofapi.go` | **NEW** — `ProofCheckerInterface`, `NewProofCheckerFn`, `GoalConcFn` |
| `proof/register.go` | Replace init() with `RegisterFactories()` setting module vars |
| `compiler/phase6.go` | Remove local interface+vars (use module's); BUG 3 (line 1975), BUG 6 (lines 2011,2016), BUG 7 (line 1811) |
| `check/check.go` | Replace init() at line 1021 with `RegisterTactics()` |
| `check/setup.go` | **NEW** — `GoIvySetup()` orchestrator |
| `l2s/l2s.go` | Replace init() at line 831 with `RegisterTactics()` |
| `temporal/temporal.go` | Replace init() at line 656 with `RegisterTactics()` |
| `actions/annotation.go` | Replace init() at line 11 with `InitAnnotConj()` |
| `actions/action.go` | Replace init() at line 1166 with `InitGlobalContext()` |
| `acl/acl.go` | Replace init() at line 22 with `InitACL()` |
| `proof/checker_test.go` | Tests for BUGs 1, 4, 5; call GoIvySetup() |
| `end2end/*_test.go` | Update TestMain to call GoIvySetup() |

## Files to Create Later (Part C)

| File | Purpose |
|------|---------|
| `ivyutils/config.go` | GoIvyConfig struct + NewGoIvyConfig() |

---

## Verification

1. `cd /Users/jaten/go/src/github.com/glycerine/goivy && go build ./...`
2. `make test`
3. New tests in `proof/checker_test.go`:
   - Definition keyed by defines-name (BUG 1)
   - GetSubgoals rejects definitions (BUG 5)
   - AdmitProposition with pre-supplied subgoals (BUG 4)
4. Verify no init() calls remain for the 7 registration functions (grep)
5. Verify end2end tests pass with explicit GoIvySetup()
