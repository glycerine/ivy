# Audit: Eliminate All Remaining Mutable Globals

**Created:** 2026-03-27 23:45

## Context

CLAUDE.md Section C mandates: no package-level `var` for mutable state. After the BUG F work (which deleted `DefaultIvyUtilsConfig` and 6 ivyutils globals), a full audit reveals **45 remaining mutable globals** across 12 packages. Many already have Config struct equivalents — the globals are transitional duplicates that need caller migration and deletion.

Read-only state (compiled regexps, const maps, singleton sorts, `var _ =` assertions, goyacc parser tables) is exempt per CLAUDE.md Section C rule 9.

---

## Inventory of Mutable Globals

### GROUP A: Transitional Default Configs (4 globals — delete, migrate callers)

These are `var DefaultFooConfig = NewFooConfig()` patterns that violate Section C rule 6.

| # | Global | File | Callers |
|---|--------|------|---------|
| A1 | `DefaultIsolateConfig` | `isolate/isolate.go:65` | 0 external callers |
| A2 | `DefaultInterpConfig` | `interp/interp.go:283` | 3 legacy wrapper fns (lines 311,316,321) |
| A3 | `DefaultTransrelConfig` | `transrel/phase4.go:341` | 2 legacy wrapper fns (lines 361,366) |
| A4 | `DefaultCodegenConfig` | `codegen/codegen.go:117` | 0 external callers |

**Fix:** Delete globals. Delete or convert legacy wrapper functions to require `*Config` parameter. If wrappers have callers, trace them and pass config.

### GROUP B: Isolate Package Flags (16 globals — already duplicated on IsolateConfig)

`isolate/isolate.go:71-87` has 15 bare `var` flags that are exact duplicates of fields on `IsolateConfig`. Plus `IvyVersion` at `iter.go:567`.

| # | Global | IsolateConfig field |
|---|--------|-------------------|
| B1 | `ShowCompiled` | `.ShowCompiled` |
| B2 | `ConeOfInfluence` | `.ConeOfInfluence` |
| B3 | `FilterSymbols` | `.FilterSymbols` |
| B4 | `CreateImports` | `.CreateImports` |
| B5 | `EnforceAxioms` | `.EnforceAxioms` |
| B6 | `DoCheckInterference` | `.DoCheckInterference` |
| B7 | `Pedantic` | `.Pedantic` |
| B8 | `PreferImpls` | `.PreferImpls` |
| B9 | `KeepDestructors` | `.KeepDestructors` |
| B10 | `IsolateMode` | `.IsolateMode` |
| B11 | `CompileWithInvariants` | `.CompileWithInvariants` |
| B12 | `AssumeInvariants` | `.AssumeInvariants` |
| B13 | `InterpretAllSorts` | `.InterpretAllSorts` |
| B14 | `NumIsolateParams` | `.NumIsolateParams` |
| B15 | `StripAddedSymbols` | `.StripAddedSymbols` |
| B16 | `IvyVersion` (iter.go:567) | `.IvyVersion` |

**106 internal references** across 8 files within `isolate/`. No external callers (confirmed by grep).

**Fix:** All isolate functions that read these globals need access to an `*IsolateConfig`. The config should be threaded from `module.Config` (add `IsolateConfig *IsolateConfig` field to `module.Config`). Every isolate function that takes `*module.Module` already has `mod.Cfg` — so access is `mod.Cfg.IsolateCfg.ConeOfInfluence` etc. Functions that don't take a module need a config parameter added.

### GROUP C: Actions Package Globals (4 globals — ActionsConfig exists)

| # | Global | File | Notes |
|---|--------|------|-------|
| C1 | `determinize` | `actions/phase3.go:464` | Get/SetDeterminize read/write it |
| C2 | `choiceActionCtr` | `actions/action.go:591` | `cfg.NewChoiceAction` method already exists |
| C3 | `callActionCtr` | `actions/action.go:633` | `cfg.NewCallAction` method already exists |
| C4 | `SymexParams` | `actions/action.go:1620` | Module-level `symex_params = []` in Python |

**Fix:** C1: Already on ActionsConfig. Migrate callers of `Get/SetDeterminize()` to use config. C2-C3: The free `NewChoiceAction()`/`NewCallAction()` functions use global counters — find all callers and migrate to `cfg.NewChoiceAction()`. C4: Add `SymexParams []lg.Expr` to ActionsConfig, thread through SymExContext.

### GROUP D: ivyutils Globals (6 globals)

| # | Global | File | Refs |
|---|--------|------|------|
| D1 | `Filename` | `source_file.go:6` | `WithSourceFile` + 1 caller in `compiler/ivyinit.go` |
| D2 | `ComposeCharacter` | `names.go:14` | 38 refs across 10 files |
| D3 | `stdIncludeDir` | `names.go:180` | Cache used by `GetStdIncludeDir` method |
| D4 | `ParseErrorListVar` | `error.go:184` | `ParseWith`/`PError` |
| D5 | `GlobalRegistry` | `globals.go:8` | Only in globals.go + test |
| D6 | `uiModules` | `globals.go:40` | `RegisterUIModule`/`GetDefaultUIModule` |

**Fix:**
- D1: `IvyUtilsConfig` already has `Filename` field and `WithSourceFile` method. Delete global, redirect `iu.WithSourceFile()` callers to `cfg.WithSourceFile()`.
- D2: Add `ComposeCharacter` to `IvyUtilsConfig`. This is the largest migration (38 refs across 10 files). All callers already have access to config via `mod.Cfg.IuCfg`.
- D3: Already a field on `IvyUtilsConfig`. Delete the package-level cache `var`.
- D4: Add `ParseErrorList` to `IvyUtilsConfig`. Parser already has config access.
- D5: Add `Registry *ParameterRegistry` to `IvyUtilsConfig`. Parameters register on the per-session registry.
- D6: Add `UIModules map[string]*UIModule` to `IvyUtilsConfig`.

### GROUP E: Compiler Globals (2 globals)

| # | Global | File | Notes |
|---|--------|------|-------|
| E1 | `globalIncluded` | `ivyinit.go:32` | Reset per `SourceFile()` call, passed to parser |
| E2 | `AdmitDefinitionFactory` | `ivy_compile.go:46` | Already has `mod.Cfg.AdmitDefinitionFactory` equivalent |

**Fix:**
- E1: Move to `module.Config.GlobalIncluded map[string]bool`. `SourceFile` already has `mod.Cfg`. `ReadModule` already gets `cfg *module.Config`.
- E2: Delete global. `IvyCompile` already checks `mod.Cfg.AdmitDefinitionFactory` first (line 77-78). Remove the fallback to the global.

### GROUP F: mc Package Globals (4 globals — internal only)

| # | Global | File | Notes |
|---|--------|------|-------|
| F1 | `propAbsCtr` | `propabs.go:12` | Atomic counter |
| F2 | `iteCtr` | `qelim.go:13` | Atomic counter |
| F3 | `Verbose` | `checker.go:15` | Flag |
| F4 | `CheckedAssert` | `phase7.go:270` | Lineno filter |

**Fix:** Create `McConfig` struct. Add to `module.Config`. F3/F4 mirror fields already on `module.Config` (`CheckedAction`, `OptTrace`). Thread from `module.Config` at mc entry points.

### GROUP G: vmt Package Globals (2 globals — internal only)

| # | Global | File | Notes |
|---|--------|------|-------|
| G1 | `Verbose` | `vmt.go:26` | Flag |
| G2 | `CheckedAssertValue` | `vmt.go:31` | Lineno filter |

**Fix:** Create `VmtConfig` or add to `module.Config` directly (these are simple flags). Thread from entry points.

### GROUP H: Other Package Globals (6 globals)

| # | Global | File | Notes |
|---|--------|------|-------|
| H1 | `OptionAbsInit` | `art/art.go:1474` | Boolean flag |
| H2 | `DiagnoseMode` | `webui/ui_show.go:60` | Boolean flag |
| H3 | `CompileKwargs` | `webui/ui_show.go:64` | Map |
| H4 | `Verbose` | `autoinst/autoinst.go:20` | Boolean flag — already has `mod.Cfg.AutoinstVerbose` |
| H5 | `tempCounter` | `codegen/codegen.go:124` | Atomic counter |
| H6 | `currentContext` | `codegen/codegen.go:180` | Context pointer |

**Fix:**
- H1: Add to `module.Config` (already has similar flags).
- H2: `module.Config` already has `Diagnose` field. Use it.
- H3: Move to session config on webui.
- H4: Read from `mod.Cfg.AutoinstVerbose` (already exists). Delete global.
- H5-H6: Move to `CodegenConfig`. Thread config through codegen functions.

### GROUP I: ivylogic Globals (1 global)

| # | Global | File | Notes |
|---|--------|------|-------|
| I1 | `polymorphicSymbols` | `poly.go:59` | Map, written to dynamically at line 136 |

**Fix:** The writes are in `FindPolymorphicSymbol` which adds "bfe[...]" symbols. The initial map is built from a static definition (read-only). The dynamic adds are the problem. Move to `IvyLogicConfig.PolymorphicSymbols` (which already exists as a field — verify it's being used instead of the package-level map).

### GROUP J: Exempt / Edge Cases (not actionable now)

| # | Global | File | Reason |
|---|--------|------|--------|
| J1 | `suppressed` | `xtracer/xtracer.go:35` | Set once from env at init, effectively read-only |
| J2 | `debugNextDefinitionDeclSn` | `ast/decl_ast.go:15` | Debug counter (debug-only exempt) |
| J3 | `v17Debug`, `v17ErrorVerbose` | `lalr_full/grammar_v17.go` | goyacc-generated code |
| J4 | `HandleRangeSorts` | `solver/z3convert.go:1066` | Already has `mod.Cfg.HandleRangeSorts` — needs caller migration |

---

## Execution Plan

### Phase 1: Low-hanging fruit — delete unused transitional defaults (GROUP A)

Delete 4 `DefaultFooConfig` globals that have 0 or few callers:
- `DefaultIsolateConfig` — delete (0 callers)
- `DefaultCodegenConfig` — delete (0 callers)
- `DefaultInterpConfig` — delete, convert 3 legacy wrappers to take config param
- `DefaultTransrelConfig` — delete, convert 2 legacy wrappers to take config param

Also delete `AdmitDefinitionFactory` global (E2) — the config path is already primary.

**Files:** `isolate/isolate.go`, `interp/interp.go`, `transrel/phase4.go`, `codegen/codegen.go`, `compiler/ivy_compile.go`

### Phase 2: Actions globals (GROUP C)

- Delete `choiceActionCtr`, `callActionCtr` globals — migrate all `NewChoiceAction()`/`NewCallAction()` free function callers to `cfg.NewChoiceAction()`
- Delete `determinize` global + Get/SetDeterminize — migrate callers to `cfg.Determinize`
- Move `SymexParams` to `ActionsConfig`

**Files:** `actions/action.go`, `actions/phase3.go`, callers across compiler/isolate/etc.

### Phase 3: Isolate flags (GROUP B)

- Add `IsolateCfg *IsolateConfig` to `module.Config`
- Wire in `NewConfig()`
- Convert all 106 internal references from bare globals to `cfg.Field`
- Delete 16 globals + `DefaultIsolateConfig`

**Files:** `module/config.go`, `isolate/isolate.go`, `isolate/iter.go`, `isolate/create.go`, `isolate/helpers.go`, `isolate/strip.go`, `isolate/deps.go`

### Phase 4: ivyutils globals (GROUP D)

- D1: Delete `Filename` global, migrate `WithSourceFile` callers
- D2: Migrate `ComposeCharacter` (38 refs) to `IvyUtilsConfig.ComposeCharacter`
- D3: Delete `stdIncludeDir` package-level cache
- D4: Move `ParseErrorListVar` to config
- D5: Move `GlobalRegistry` to config
- D6: Move `uiModules` to config

**Files:** `ivyutils/source_file.go`, `ivyutils/names.go`, `ivyutils/error.go`, `ivyutils/globals.go`, `ivyutils/config.go`, plus 10 files for ComposeCharacter

### Phase 5: Compiler, mc, vmt, other globals (GROUPS E, F, G, H, I)

- E1: Move `globalIncluded` to `module.Config`
- F1-F4: Create `McConfig`, thread from `module.Config`
- G1-G2: Add vmt flags to `module.Config`
- H1-H6: Thread remaining flags through configs
- I1: Migrate `polymorphicSymbols` dynamic writes to `IvyLogicConfig`
- J4: Migrate `HandleRangeSorts` callers to `mod.Cfg.HandleRangeSorts`

### Phase 6: Cleanup — delete `var HandleRangeSorts` in solver

---

## Verification

After each phase:
```bash
cd ~/goivy && go build ./... && go test ./ivyutils/ ./ivylogic/ ./compiler/ ./module/ ./isolate/ ./actions/ ./solver/ ./proof/ ./mc/ ./check/ ./interp/ ./transrel/ ./art/ ./codegen/ ./autoinst/ ./vmt/ ./webui/
```

After all phases:
```bash
# Verify no mutable globals remain (excluding vprint.go, test files, exempt items)
grep -rn '^var ' ~/goivy --include='*.go' --exclude='*vprint.go' --exclude='*_test.go' | grep -v '/vendor/' | grep -v 'regexp.MustCompile\|= map\[string\]bool{\|= map\[string\]int{\|= \[\]string{'
```
