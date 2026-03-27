# BUG F: Eliminate Transitional Globals and DefaultIvyUtilsConfig

**Created:** 2026-03-27 22:30

## Context

Seven transitional mutable globals in `ivyutils/names.go` and two in `isolate/` violate CLAUDE.md Section C ("No Global Variables"). The `DefaultIvyUtilsConfig` global exists as a crutch for unmigrated callers — it must be deleted, not leaned on further.

The Config infrastructure already exists: `module.Config.IuCfg *iu.IvyUtilsConfig` is created by `NewConfig()`. But `ivylogic/` functions read the package-level globals directly instead of accessing config. The fix is to thread `*IvyUtilsConfig` through `Sig`, `IvyLogicConfig`, and the 5 free functions that read globals, so every caller gets its flags from a real per-session config, not a shared global.

Python uses module-level globals for these flags — that's fine for Python's single-threaded model. Go must use Config for multi-tenancy.

---

## Scope

### Globals to delete

| Global | File | Readers |
|--------|------|---------|
| `ivyLanguageVersion` | `ivyutils/names.go:129` | `GetStringVersion`, `GetNumericVersion`, `GetStdIncludeDir` |
| `IvyHavePolymorphism` | `ivyutils/names.go:132` | `sig.go:97`, `poly.go:96,119` |
| `IvyUsePolymorphicMacros` | `ivyutils/names.go:135` | `ivylogic.go:386`, `classify_ext.go:346` |
| `IvyForbidGhostInit` | `ivyutils/names.go:138` | tests only |
| `IvyLatestLanguageVersion` | `ivyutils/names.go:141` | none found |
| `SymbolCharsParser` | `ivyutils/names.go:144` | none found |
| `DefaultIvyUtilsConfig` | `ivyutils/config.go:68` | never referenced |
| `ExtAction` | `isolate/create.go:21` | `create.go:303,328,330` |
| `VPrivates` | `isolate/iter.go:23` | `iter.go:28`, `helpers.go:133,136,153` |

### Free functions to retire (ivyutils)

The free functions `SetStringVersion()`, `GetStringVersion()`, `GetNumericVersion()`, `GetStdIncludeDir()` currently read/write the globals. After the globals are deleted, these functions must either be deleted (callers use config methods directly) or converted to require a `*IvyUtilsConfig` parameter. The config already has method versions: `cfg.SetStringVersion()`, `cfg.GetStringVersion()`.

---

## Phase 1: Thread `*IvyUtilsConfig` into `Sig`

### 1a. Add `IuCfg` field to `Sig`

**File:** `ivylogic/sig.go`

```go
type Sig struct {
    IuCfg              *iu.IvyUtilsConfig     // per-session config
    Sorts              map[string]lg.Sort
    // ... rest unchanged
}
```

### 1b. Add `NewSigOn(iuCfg)` constructor, update `NewSig()`

**File:** `ivylogic/sig.go`

```go
// NewSigOn creates a Sig with the given config. Production code should use this.
func NewSigOn(iuCfg *iu.IvyUtilsConfig) *Sig {
    s := &Sig{
        IuCfg:  iuCfg,
        Sorts:  make(map[string]lg.Sort),
        // ... same as current NewSig
    }
    return s
}

// NewSig creates a Sig with a fresh default config. For tests only.
func NewSig() *Sig {
    return NewSigOn(iu.NewIvyUtilsConfig())
}
```

Also update `Sig.Copy()` (line 69) to copy `IuCfg`:
```go
func (s *Sig) Copy() *Sig {
    res := &Sig{
        IuCfg: s.IuCfg,
        // ... rest unchanged
    }
```

### 1c. `Sig.AddSymbol` reads config instead of global

**File:** `ivylogic/sig.go:97`

```go
// BEFORE:
if iu.IvyHavePolymorphism && IsPolymorphicName(name) {

// AFTER:
if s.IuCfg.HavePolymorphism && IsPolymorphicName(name) {
```

### 1d. Update production callers of `NewSig()` → `NewSigOn(iuCfg)`

Key production sites that already have access to `mod.Cfg.IuCfg`:

| File | Line | Current | After |
|------|------|---------|-------|
| `module/module.go` | 282 | `m.Sig = il.NewSig()` | `m.Sig = il.NewSigOn(m.Cfg.IuCfg)` |
| `compiler/compiler.go` | 171 | `c.Sig = il.NewSig()` | `c.Sig = il.NewSigOn(c.Module.Cfg.IuCfg)` |
| `compiler/ivyinit.go` | 248 | `sig := il.NewSig()` | `sig := il.NewSigOn(mod.Cfg.IuCfg)` |
| `compiler/phase6.go` | 1296 | `schemaSig := il.NewSig()` | `schemaSig := il.NewSigOn(c.Module.Cfg.IuCfg)` |
| `proof/phase5_matching.go` | 130 | `return il.NewSig()` | Need to check if config is available |
| `solver/solver.go` | 58 | `sig: il.NewSig()` | Leave as `NewSig()` (solver creates standalone contexts) |
| `webui/session.go` | 104 | `sig := il.NewSig()` | Check config access |

Test call sites (~60) keep using `NewSig()` which creates a fresh config internally.

Isolate test sites that do `m.Sig = il.NewSig()` after `m := module.NewModule(...)` should use `m.Sig = il.NewSigOn(m.Cfg.IuCfg)` — but only if `m.Cfg` is non-nil in the test. Most tests create modules with `NewModule()` which sets `Cfg: NewConfig()`.

---

## Phase 2: Thread `*IvyUtilsConfig` into `IvyLogicConfig` and free functions

### 2a. Add `IuCfg` to `IvyLogicConfig`

**File:** `ivylogic/poly.go`

```go
type IvyLogicConfig struct {
    IuCfg              *iu.IvyUtilsConfig     // per-session version flags
    PolymorphicSymbols map[string]*lg.Symbol
    // ... rest unchanged
}
```

Update `NewIvyLogicConfig()` to accept config:
```go
func NewIvyLogicConfigOn(iuCfg *iu.IvyUtilsConfig) *IvyLogicConfig {
    return &IvyLogicConfig{
        IuCfg:              iuCfg,
        PolymorphicSymbols: buildPolymorphicSymbols(),
        Equals:             lg.NewSymbol("=", RelationSort([]lg.Sort{lg.TopS, lg.TopS})),
    }
}

func NewIvyLogicConfig() *IvyLogicConfig {
    return NewIvyLogicConfigOn(iu.NewIvyUtilsConfig())
}
```

### 2b. `FindPolymorphicSymbolOn` reads config

**File:** `ivylogic/poly.go:96`

```go
// BEFORE:
if !iu.IvyHavePolymorphism {

// AFTER:
if cfg.IuCfg == nil || !cfg.IuCfg.HavePolymorphism {
```

### 2c. Convert `FindPolymorphicSymbol` (free function) to use Sig's config

The free function `FindPolymorphicSymbol` at `poly.go:118` uses a package-level `polymorphicSymbols` map (line 59) AND reads `iu.IvyHavePolymorphism`. This is the hardest one — it has no config access.

**Approach:** Add `iuCfg *iu.IvyUtilsConfig` parameter:

```go
func FindPolymorphicSymbol(name string, iuCfg *iu.IvyUtilsConfig) (*lg.Symbol, bool) {
    if iuCfg == nil || !iuCfg.HavePolymorphism {
        return nil, false
    }
    // ... rest uses package-level polymorphicSymbols map (read-only, safe)
}
```

Update 5 callers:
- `ivylogic/context.go:135,143` — these are methods, need to carry config
- `compiler/helpers.go:118` — has access to `c.Module.Cfg.IuCfg`
- `compiler/compiler.go:560,1129` — has access to `c.Module.Cfg.IuCfg`

### 2d. Convert `NormalizeSymbol` — add config parameter

**File:** `ivylogic/ivylogic.go:385`

```go
func NormalizeSymbol(sym *lg.Symbol, iuCfg *iu.IvyUtilsConfig) *lg.Symbol {
    if iuCfg != nil && iuCfg.UsePolymorphicMacros {
        if canonical, ok := PolymorphicMacrosMap[sym.Name]; ok {
            return lg.NewSymbol(canonical, sym.CSort)
        }
    }
    return sym
}
```

Currently 0 production callers — change is free.

### 2e. Convert `IsMacro` — add config parameter

**File:** `ivylogic/classify_ext.go:345`

```go
func IsMacro(term lg.Expr, iuCfg *iu.IvyUtilsConfig) bool {
    if iuCfg == nil || !iuCfg.UsePolymorphicMacros {
        return false
    }
    // ... rest unchanged
}
```

Update 2 production callers + 2 test callers:
- `mc/phase7.go:64` — check config access
- `autoinst/autoinst.go:211` — check config access
- `ivylogic/new_funcs_test.go:287,293` — pass `iu.NewIvyUtilsConfig()` with flags set

---

## Phase 3: Delete ivyutils globals and free functions

### 3a. Delete globals from `ivyutils/names.go`

Delete lines 129-144:
- `var ivyLanguageVersion`
- `var IvyHavePolymorphism`
- `var IvyUsePolymorphicMacros`
- `var IvyForbidGhostInit`
- `var IvyLatestLanguageVersion`
- `var SymbolCharsParser`

### 3b. Delete or convert free functions

The free functions `GetStringVersion()`, `SetStringVersion()`, `GetNumericVersion()`, `GetStdIncludeDir()` read/write the globals. After globals are deleted:

- **Delete** `SetStringVersion()` (free function) — callers use `cfg.SetStringVersion()` directly
- **Delete** `GetStringVersion()` (free function) — callers use `cfg.GetStringVersion()`
- **Delete** `GetNumericVersion()` (free function) — callers use `cfg.GetNumericVersion()`
- **Convert** `GetStdIncludeDir()` to method `(cfg *IvyUtilsConfig) GetStdIncludeDir()` using `cfg.LanguageVersion` and `cfg.StdIncludeDir`

Find all callers of these free functions and redirect to config methods. The callers are primarily in:
- `compiler/` — has `mod.Cfg.IuCfg`
- `lalr_full/` — has parser config with `AstCfg` → need to check IuCfg access
- `isolate/` — has `mod.Cfg.IuCfg`
- `actions/` — has config access

### 3c. Delete `DefaultIvyUtilsConfig`

**File:** `ivyutils/config.go:67-68`

Delete:
```go
// DefaultIvyUtilsConfig is a transitional default for unmigrated callers.
var DefaultIvyUtilsConfig = NewIvyUtilsConfig()
```

---

## Phase 4: Isolate globals

### 4a. `ExtAction` → `mod.Cfg.ExtAction`

**File:** `isolate/create.go`

`CreateIsolate(iso string, mod *module.Module)` already receives `mod`. Replace:
```go
// Line 303: if ExtAction != ""
if mod.Cfg != nil && mod.Cfg.ExtAction != "" {
    extAction := mod.Cfg.ExtAction
    // Lines 328, 330: use extAction
}
```

Delete `var ExtAction = ""` at line 21.

Update test `isolate/batch_fixes_test.go:651-653` to set `mod.Cfg.ExtAction`.

### 4b. `VPrivates` → `mod.VPrivates`

Add `VPrivates map[string]bool` field to `module.Module`.

**Writer** (`isolate/helpers.go:133-153`): Write to `mod.VPrivates` instead of global.
**Reader** (`isolate/iter.go:28`): Read `mod.VPrivates` instead of global.

Delete `var VPrivates = make(map[string]bool)` at `iter.go:23`.

---

## Execution order

1. Phase 1 (Sig) — self-contained, testable independently
2. Phase 2 (IvyLogicConfig + free functions) — depends on Phase 1
3. Phase 3 (delete globals) — depends on Phases 1-2
4. Phase 4 (isolate globals) — independent of Phases 1-3

## Verification

After each phase:
```bash
go build ./...
go test ./ivylogic/ ./ivyutils/ ./compiler/ ./module/ ./isolate/ ./solver/ ./proof/ ./mc/ ./autoinst/ ./actions/ ./check/ ./webui/ ./interp/
cd ~/goivy && make golden
```
