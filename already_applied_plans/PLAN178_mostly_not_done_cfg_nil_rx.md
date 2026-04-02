# Fix: `Module.Copy()` doesn't copy `Cfg.ActCfg` — nil cfg panic in `NewLocalActionOn`

NOTE: most of this besides *s.Cfg = *m.Cfg is a bad idea, and was skipped.

**Created**: 2026-04-01 23:00
**Previous fixes**: Consolidated `SubstituteConstantsAST` (moved divergence from 157358→157372)

## Context

The golden test panics at `actions/action.go:862` with `panic: cfg must not be nil`. The panic stack trace (`~/ivy/goivy/log.panic`) shows:

```
NewLocalActionOn(cfg=nil)        ← action.go:862
  IfAction.subactionsSome(actCfg=0x0)  ← action.go:596
    IfAction.Subactions(actCfg=nil)     ← action.go:503
      intUpdateWithSubactions           ← update.go:1559: a.Subactions(ctx.ActCfg)
        IfAction.IntUpdate              ← update.go:1527
          ...
          GetUpdate                     ← update.go:2143
            fragment.makeFmlaPairFromAction ← fragment.go:951: ctx.ActCfg = m.Cfg.ActCfg
              ...
              CheckIsolate(isoMod)      ← isolate_check.go:955
                CheckModule(mod)        ← isolate_check.go:880: isoMod := mod.Copy()
```

**Root cause**: `Module.Copy()` (`module/module.go:293`) calls `c := New()` which creates a fresh `Config` via `NewConfig()`. `NewConfig()` does NOT set `ActCfg` (it's nil). Copy() never copies `m.Cfg.ActCfg` to `c.Cfg.ActCfg`. So the copied module has `Cfg.ActCfg == nil`.

When `fragment.makeFmlaPairFromAction` reads `m.Cfg.ActCfg` at line 951, it gets nil. This nil propagates through `UpdateContext.ActCfg` → `intUpdateWithSubactions` → `Subactions(nil)` → `subactionsSome(nil)` → `NewLocalActionOn(nil)` → **PANIC**.

Python doesn't have this bug because `module.__copy__` shallow-copies `__dict__` which preserves all fields including the action config.

## Changes

### Step 1: Fix `Module.Copy()` — copy config from source

**File: `module/module.go`**, in `func (m *Module) Copy()`, after `c := New()`:

Add config inheritance. Python's `__copy__` preserves all module attributes. The key sub-configs (AstCfg, IuCfg, ActCfg) contain shared counters (matching Python globals), so sharing the pointers is correct:

```go
// After c := New() (around line 295):
// Inherit config from source. Python __copy__ preserves __dict__.
// Sub-config pointers are shared (matching Python's global counters).
c.Cfg.AstCfg = m.Cfg.AstCfg
c.Cfg.IuCfg = m.Cfg.IuCfg
c.Cfg.ActCfg = m.Cfg.ActCfg
c.Cfg.ProofCfg = m.Cfg.ProofCfg
c.Cfg.IsolateCfg = m.Cfg.IsolateCfg
// Copy value-type config fields that were set during compilation:
c.Cfg.Coverage = m.Cfg.Coverage
c.Cfg.MacroFinder = m.Cfg.MacroFinder
c.Cfg.OptSummary = m.Cfg.OptSummary
c.Cfg.OptTrusted = m.Cfg.OptTrusted
c.Cfg.OptSeparate = m.Cfg.OptSeparate
c.Cfg.OptAction = m.Cfg.OptAction
c.Cfg.OptUncheckedProps = m.Cfg.OptUncheckedProps
c.Cfg.GlobalIncluded = m.Cfg.GlobalIncluded
```

**Simpler alternative** — shallow-copy the entire Config struct:
```go
*c.Cfg = *m.Cfg
```
This copies all fields at once. Pointer fields (AstCfg, IuCfg, ActCfg, etc.) are shared, value fields (bools, strings) are independent copies. This exactly matches Python's shallow __dict__ copy.

### Step 2: Defense-in-depth — `NewConfig()` creates default `ActCfg`

**File: `module/config.go`**, in `func NewConfig()`:

Add `ActCfg` creation so any module always has a non-nil ActCfg, even if not explicitly set. This prevents panics from other `module.New()` call sites:

```go
func NewConfig() *Config {
    iuCfg := iu.NewIvyUtilsConfig()
    astCfg := ast.NewAstConfig()
    astCfg.IuCfg = iuCfg
    return &Config{
        Coverage:         true,
        MacroFinder:      true,
        GlobalIncluded:   make(map[string]bool),
        AstCfg:           astCfg,
        IuCfg:            iuCfg,
        ActCfg:           NewActionsConfig(), // defense-in-depth
        IsolateCfg:       NewIsolateConfig(),
        HandleRangeSorts: true,
        AlphaTestBottom:  true,
        AutoinstVerbose:  true,
        TraceDetailed:    true,
        ProofCfg:         TacticNewConfig(),
    }
}
```

Note: `NewActionsConfig()` creates an ActCfg with `IuCfg: iu.NewIvyUtilsConfig()`. But the compiler at `ivy_compile.go:105-108` later creates a NEW ActCfg with `IuCfg` shared from the module's AstCfg. This override is correct and expected — the default just prevents nil panics for code paths that don't go through the compiler.

For proper counter sharing when the compiler IS used, `ivy_compile.go:105-108` should also wire the new ActCfg's IuCfg from the module's config:
```go
actCfg := module.NewActionsConfig()
if mod.Cfg != nil && mod.Cfg.AstCfg != nil {
    actCfg.IuCfg = mod.Cfg.AstCfg.IuCfg
}
```
This already exists and is correct.

### Step 3: Audit other `module.New()` call sites

These production code sites call `module.New()` without ensuring `ActCfg` is set. After Step 2 (default ActCfg in NewConfig), they all get a default ActCfg automatically. But they should be checked to ensure proper counter sharing:

| File | Line | Context | Status after fix |
|------|------|---------|-----------------|
| `compiler/phase6.go:2633` | `IvyLoadFile()` | Creates mod, later compiled (sets ActCfg) | OK |
| `compiler/phase6.go:2671` | `CompileModule()` | Creates mod, later compiled | OK |
| `compiler/compiler.go:184` | `Ensure()` | Fallback when Module is nil | OK (line 188 creates ActCfg) |
| `compiler/ivy_compile.go:66` | `IvyCompile()` | Creates mod when nil, then compiles | OK (line 105 creates ActCfg) |
| `check/check.go:1163` | `CheckModule()` | Creates mod | **Check**: does it go through compiler? |
| `check/check.go:1238` | `CheckModuleWithConfig()` | Creates mod | **Check** |
| `check/isolate_check.go:605` | `CheckSubgoals()` | Creates mod when nil | **Check** |
| `trace/trace.go:103,457` | `NewTraceBase/NewTrace` | Creates mod | OK if no actions used |
| `interp/interp.go:102` | `NewState()` | Creates mod when nil | **Check** |
| `webui/session.go:109` | `Session.Compile()` | Creates mod, then compiles | OK |
| `proof/phase5_matching.go:56,115` | `compile/compileVocab` | Creates mod when nil | **Check** |
| `art/art.go:281` | `NewAnalysisGraph()` | Creates mod when nil | **Check** |

After Step 2, all these get a default ActCfg. The default has its own IuCfg (separate counters from the main compilation), which is acceptable for:
- Code paths that don't create LocalActions (no counter mismatch)
- Code paths that are independent of the main compilation

## Key files

- `module/module.go:293-440` — `Module.Copy()` (missing Cfg copy)
- `module/config.go:185-202` — `NewConfig()` (missing ActCfg default)
- `module/config.go:229-231` — `NewActionsConfig()` (creates IuCfg)
- `compiler/ivy_compile.go:105-112` — ActCfg creation and wiring
- `fragment/fragment.go:950-953` — reads `m.Cfg.ActCfg` (nil panic site)
- `check/isolate_check.go:880` — `mod.Copy()` call site

## Verification

```bash
cd ~/ivy/goivy && go build ./... && go test ./... && make golden
```

Check that:
1. `go build` and `go test` pass
2. No panic in `log.panic` when running golden test
3. The divergence in `log.red` advances past line 157372
