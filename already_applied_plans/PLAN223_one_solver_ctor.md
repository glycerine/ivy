# PLAN222: Fix SetMacroFinder Divergence — Move SolverOptions to module.Config, Consolidate Constructors

**Created:** 2026-04-07 ~20:30 UTC

## Context

`~/ivy/goivy/log.red` shows the first xtracer line diverges because:
1. Python calls `set_macro_finder(True)` at import time (`ivy_solver.py:53`), emitting a trace before `check.start ENTER`. Go never does.
2. Go's `SetMacroFinder()` stores a flag but never applies it to Z3.
3. `solver.New()` uses `il.NewSig()` (empty sig) even when real sigs are available — likely wrong.
4. There are 3 solver constructors but only `New()` is called externally. `NewWithOptions()` is never used outside tests.

**Design constraints:**
- No globals — thread `module.Config` through
- Always set `smt.macro_finder` on Z3 solvers (don't assume Z3's default matches ours)
- Conform to Python, don't diverge

## Changes

### 1. module/config.go — Add `SolverOptions` type

Move `solver.Options` here, renamed `SolverOptions`. Add `DefaultSolverOptions()`. Add field to `Config`.

```go
// SolverOptions controls per-solver Z3 behavior.
// Moved from solver.Options so the full config is accessible
// from module.Config without import cycles.
type SolverOptions struct {
    Seed        int
    Incremental bool
    MacroFinder bool
    ShowVCs     bool
    UseZ3Enums  bool
}

func DefaultSolverOptions() *SolverOptions {
    return &SolverOptions{
        Seed:        0,
        Incremental: true,
        MacroFinder: true,
        ShowVCs:     false,
        UseZ3Enums:  true,
    }
}
```

In `Config` struct: replace standalone `MacroFinder bool` field (line 58-60) with:
```go
SolverOpts *SolverOptions
```

In `NewConfig()`: replace `MacroFinder: true` (line 192) with:
```go
SolverOpts: DefaultSolverOptions(),
```

### 2. solver/solver.go — Delete `Options`, `New()`, `NewWithSig()`, `NewWithOptions()`; replace with `NewSolver()`

Delete:
- `type Options struct` (lines 27-33)
- `func DefaultOptions()` (lines 35-43)
- `func New()` (lines 64-74)
- `func NewWithSig()` (lines 76-87)
- `func NewWithOptions()` (lines 89-103)

Replace with `NewSolver` referencing `module.SolverOptions` and handling nil sig:
```go
func NewSolver(sig *il.Sig, opts *module.SolverOptions) *Solver {
    if opts == nil {
        opts = module.DefaultSolverOptions()
    }
    if sig == nil {
        sig = il.NewSig()
    }
    s := &Solver{
        tr:               z3bridge.NewTranslator(),
        opts:             opts,
        sig:              sig,
        HandleRangeSorts: true,
        impliesCache:     make(map[[2]lg.NodeKey]bool),
    }
    s.wireNativeLookup()
    return s
}
```

Change `opts` field type: `opts *module.SolverOptions`

Add `newZ3Solver()` helper that **always** sets `smt.macro_finder`:
```go
func (s *Solver) newZ3Solver() *z3bridge.Solver {
    zs := s.tr.Ctx.NewSolver()
    if s.opts.MacroFinder {
        zs.SetParam("smt.macro_finder", "true")
    } else {
        zs.SetParam("smt.macro_finder", "false")
    }
    return zs
}
```

Replace all 7 `s.tr.Ctx.NewSolver()` calls (lines 616, 653, 667, 690, 743, 780, 1072) with `s.newZ3Solver()`.

### 3. solver/encoding.go — Fix `SetMacroFinder` trace capitalization

```go
func (s *Solver) SetMacroFinder(enabled bool) {
    truthStr := "False"
    if enabled {
        truthStr = "True"
    }
    xtracer.Trace("ivy_solver.py:45 set_macro_finder() ENTER truth=%s", truthStr)
    s.opts.MacroFinder = enabled
}
```

### 4. External call sites — Change `solver.New()` → `solver.NewSolver(sig, opts)`

**25 call sites across 13 files.** For each, choose the right sig and opts:

#### Sites with `*module.Module` available (pass `m.Sig` and `m.Cfg.SolverOpts`):
| File | Line | Current | New |
|------|------|---------|-----|
| `vmt/vmt.go` | 625 | `solver.New()` | `solver.NewSolver(m.Sig, m.Cfg.SolverOpts)` |
| `actions/phase4.go` | 317 | `solver.New()` | `solver.NewSolver(m.Sig, m.Cfg.SolverOpts)` (nil-guard on m) |
| `tactics/tactics.go` | 132, 200, 267 | `solver.New()` | `solver.NewSolver(tc.Mod.Sig, tc.Mod.Cfg.SolverOpts)` |

#### Sites with `*State` available (pass `state.Domain.Sig`):
| File | Line | Current | New |
|------|------|---------|-----|
| `interp/helpers.go` | 378 | `solver.New()` | `solver.NewSolver(state.Domain.Sig, nil)` |
| `interp/phase4.go` | 89, 144 | `solver.New()` | `solver.NewSolver(state.Domain.Sig, nil)` |

#### Sites without sig/opts (pass nil, nil — gets defaults):
| File | Line |
|------|------|
| `alpha/alpha.go` | 238, 767 |
| `actions/interpolant.go` | 27 |
| `actions/phase4.go` | 110, 168, 264 |
| `actions/transrel.go` | 1256, 1981, 2005 |
| `trace/trace.go` | 648 |
| `webui/concept_isession.go` | 342 |
| `webui/concept_alpha.go` | 31 |
| `end2end/verify_test.go` | 144 |

### 5. Internal solver test call sites

Update all `New()` and `NewWithSig(sig)` calls in solver/*_test.go:
- `New()` → `NewSolver(nil, nil)`
- `NewWithSig(sig)` → `NewWithOptions(sig, nil)`

Files: `solver_test.go`, `solver2_test.go`, `solver2_fuzz_test.go`, `herbrand_unitres_test.go`

### 6. References to `mod.Cfg.MacroFinder` — Update to `mod.Cfg.SolverOpts.MacroFinder`

| File | Line | Old | New |
|------|------|-----|-----|
| `cmd/goivy_check/main.go` | 158 | `cfg.MacroFinder = parseBool(val)` | `cfg.SolverOpts.MacroFinder = parseBool(val)` |
| `check/isolate_check.go` | 883 | `mod.Cfg.MacroFinder` | `mod.Cfg.SolverOpts.MacroFinder` |
| `check/isolate_check.go` | 886 | `mod.Cfg.MacroFinder = false` | `mod.Cfg.SolverOpts.MacroFinder = false` |
| `check/isolate_check.go` | 985 | `mod.Cfg.MacroFinder = true` | `mod.Cfg.SolverOpts.MacroFinder = true` |
| `check/check_port_test.go` | 255 | `cfg.MacroFinder` | `cfg.SolverOpts.MacroFinder` |

### 7. check/check.go — Emit startup trace (before `check.start ENTER`)

```go
func StartWithConfig(args []string, cfg *module.Config) error {
    // Python: set_macro_finder(True) at ivy_solver.py:53 during module import
    {
        truthStr := "True"
        if !cfg.SolverOpts.MacroFinder {
            truthStr = "False"
        }
        xtracer.Trace("ivy_solver.py:45 set_macro_finder() ENTER truth=%s", truthStr)
    }
    if len(args) >= 1 {
        xtracer.Trace("check.start ENTER file=%s", args[0])
    }
    ...
```

### 8. check/isolate_check.go — Emit trace when toggling MacroFinder

Turn off (~line 886):
```go
if saveMacroFinder {
    fmt.Println("Turning off macro_finder")
    mod.Cfg.SolverOpts.MacroFinder = false
    xtracer.Trace("ivy_solver.py:45 set_macro_finder() ENTER truth=False")
}
```

Turn on (~line 985):
```go
if hasMFAttr && saveMacroFinder {
    fmt.Println("Turning on macro_finder")
    mod.Cfg.SolverOpts.MacroFinder = true
    xtracer.Trace("ivy_solver.py:45 set_macro_finder() ENTER truth=True")
}
```

Remove the TODO comments at lines 887-889 and 986-988.

## Data Flow

```
CLI param "macro_finder=false"
  ↓
cfg.SolverOpts.MacroFinder = false   (cmd/goivy_check/main.go)
  ↓
solver.NewSolver(sig, cfg.SolverOpts)   (actions/phase4.go SmallModelClauses, etc.)
  ↓
s.opts.MacroFinder = false   (stored on *Solver)
  ↓
s.newZ3Solver() → zs.SetParam("smt.macro_finder", "false")   (solver/solver.go)
  ↓
Z3 C API: Z3_solver_set_params()
```

## Files Modified

| File | Changes |
|------|---------|
| `module/config.go` | Add `SolverOptions` type + `DefaultSolverOptions()`; replace `MacroFinder` with `SolverOpts` field |
| `solver/solver.go` | Delete `Options`, `DefaultOptions`, `New()`, `NewWithSig()`; update `NewWithOptions`; add `newZ3Solver()` |
| `solver/encoding.go` | Fix `SetMacroFinder` trace, change `opts` type ref |
| `solver/solver_test.go` | Update constructor calls and type refs |
| `solver/solver2_test.go` | Update constructor calls |
| `solver/solver2_fuzz_test.go` | Update constructor calls |
| `solver/herbrand_unitres_test.go` | Update constructor calls |
| `cmd/goivy_check/main.go` | `cfg.MacroFinder` → `cfg.SolverOpts.MacroFinder` |
| `check/check.go` | Emit startup `set_macro_finder` trace |
| `check/isolate_check.go` | Emit trace at toggle points; update MacroFinder refs |
| `check/check_port_test.go` | Update MacroFinder ref |
| `vmt/vmt.go` | `solver.New()` → `solver.NewSolver(m.Sig, m.Cfg.SolverOpts)` |
| `actions/phase4.go` | 4 call sites: thread sig/opts where available |
| `actions/interpolant.go` | `solver.New()` → `solver.NewSolver(nil, nil)` |
| `actions/transrel.go` | 3 call sites → `solver.NewSolver(nil, nil)` |
| `tactics/tactics.go` | 3 call sites → `solver.NewSolver(tc.Mod.Sig, tc.Mod.Cfg.SolverOpts)` |
| `interp/helpers.go` | `solver.New()` → `solver.NewSolver(state.Domain.Sig, nil)` |
| `interp/phase4.go` | 2 call sites → `solver.NewSolver(state.Domain.Sig, nil)` |
| `alpha/alpha.go` | 2 call sites → `solver.NewSolver(nil, nil)` |
| `trace/trace.go` | `solver.New()` → `solver.NewSolver(nil, nil)` |
| `webui/concept_isession.go` | `solver.New()` → `solver.NewSolver(nil, nil)` |
| `webui/concept_alpha.go` | `solver.New()` → `solver.NewSolver(nil, nil)` |
| `end2end/verify_test.go` | `solver.New()` → `solver.NewSolver(nil, nil)` |

## Known Issue (not fixed in this plan)

`Solver.HandleRangeSorts` is also hardcoded to `true` in the constructor despite `module.Config.HandleRangeSorts` existing. This should be moved to `SolverOptions` in a future plan for consistency.

## Verification

```bash
cd ~/go/src/github.com/glycerine/ivy/goivy && go build ./... && go test ./solver/... && go test ./z3bridge/... && go test ./check/... && go test ./actions/... && go test ./tactics/... && go test ./interp/... && go test ./vmt/... && go test ./end2end/...
cd ~/ivy/goivy && make golden
```
