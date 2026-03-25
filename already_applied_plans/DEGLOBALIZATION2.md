# Plan: Eliminate Mutable Global Variables — Refactor into Config System

**Created:** 2026-03-25T06:00

## Context

The goivy port intentionally avoids global/package-level variables to enable multi-tenancy: running thread pools of ivy models safely on multi-core machines. Recent mechanical porting efforts introduced new globals (acceptable to get basics working). This refactoring pass re-establishes the no-globals discipline by moving all mutable package-level state into the existing Config system.

**Design principles:**
1. Each package has its own Config struct.
2. Config is preferably the **method receiver** — turn `func Foo(args)` into `func (cfg *Config) Foo(args)`. This avoids changing function signatures and guides future use.
3. `module.Config` is the top-level config; other package configs are embedded in or referenced from it.

---

## Exclusions (leave as-is)

- **All vprint.go globals** (every package) — debug facilities, not used in production
- **goyacc-generated parser tables** (v17Act, v17Chk, v17R1, v17R2, v17Tok*, v17Pact, v17Pgo, v17Def, v17Exca, v17Toknames, v17Statenames, v17ErrorMessages, v17Debug, v17ErrorVerbose) — read-only after init or goyacc infrastructure
- **Compile-time interface checks** (`var _ Node = (*Type)(nil)`) — zero-value
- **Compiled regexps that never change** (incDirPat in names.go:231, ast/rewrite.go symbolCharsParser) — immutable after init
- **Singleton constants** (ast.Equals, logic.Boolean, logic.TopS, logic.True, logic.False, ivylogic.Alpha/Beta/Gamma, ivylogic.Equals) — immutable after init
- **Read-only maps** (lexer.allReserved, lexer.tokenNames, logic.infixSymbols, logic.precSymbols, compiler.DefinedAttributes, compiler.KnownLogics, ivylogic.Logics, ivylogic.DecidableLogics, ivylogic.DefaultLogics, ivylogic.macroExpansions, ivylogic.PolymorphicMacrosMap, ivylogic.InfixSymbols, ivylogic.PrecSymbols, ivylogic.UninterpretedPolymorphicSymbols, ivylogic.polymorphicSymbols, ivylogic.polymorphicSymbolsDef, solver.z3Builtins, dafnygen.OpMap, dafnygen.ArithOps, compose.RequiredFields, parser.DefaultVersion) — populated once at init
- **Test-only data** (pyIvyNoError in golden_test.go)
- **Blank imports** (`var _ = fmt.Sprint`)
- **xtracer.Enabled, xtracer.HashVerbose** — build-tag controlled debug facility (same rationale as vprint.go)
- **ivyutils.PolymorphicSymbols** — populated once at init, read-only after
- **cppgen package** — deleted, no longer exists

---

## Phase 1: `ast` package — AstConfig

**New file:** `ast/config.go`

The `ast` package cannot import `module`, so it needs its own self-contained Config.

```go
type AstConfig struct {
    ReferenceLineno        Location
    ChoiceActionCounter    int64
    LocalActionCtr         int
    CallActionCtr          int
    LfCounter              int64
    AlwaysCloneWithFreshID bool
}

func NewAstConfig() *AstConfig {
    return &AstConfig{}
}
```

### Globals to migrate (6 variables):

| Global | File:Line | New home | Notes |
|--------|-----------|----------|-------|
| `referenceLineno` | ast.go:43 | `AstConfig.ReferenceLineno` | Set/cleared around instMod calls |
| `choiceActionCounter` | ast.go:814 | `AstConfig.ChoiceActionCounter` | Incremented in NewChoiceAction, Clone |
| `localActionCtr` | ast.go:1113 | `AstConfig.LocalActionCtr` | Incremented in NewLocalAction |
| `callActionCtr` | ast.go:1158 | `AstConfig.CallActionCtr` | Incremented in NewCallAction |
| `lfCounter` | ast.go:1377 | `AstConfig.LfCounter` | Currently uses atomic; method can use atomic on field |
| `alwaysCloneWithFreshID` | ast.go:1382 | `AstConfig.AlwaysCloneWithFreshID` | Flag during instMod |

### Migration pattern — Methodize + Store `*AstConfig` on `Base`:

**Key approach: Constructors become methods on `*AstConfig`.** Instead of changing `NewAtom(rep string, terms ...Node)` to `NewAtom(cfg *AstConfig, rep string, terms ...Node)`, we make it a method:

```go
// BEFORE (function):
func NewAtom(rep string, terms ...Node) *Atom { ... }

// AFTER (method on *AstConfig):
func (cfg *AstConfig) NewAtom(rep string, terms ...Node) *Atom {
    a := &Atom{Rep: rep, Terms: terms}
    a.Cfg = cfg
    return a
}
```

**Store `*AstConfig` on `Base` struct** so Clone methods (which don't change signature) can access cfg via their receiver:

```go
type Base struct {
    Loc Location
    Cfg *AstConfig  // NEW — pointer to shared config
}
```

Clone methods access cfg via receiver's Base (no signature change):
```go
func (c *ChoiceAction) Clone(args []Node) Node {
    c.Cfg.ChoiceActionCounter++
    return &ChoiceAction{Base: c.Base, Branches: args, UniqueID: c.Cfg.ChoiceActionCounter}
}
```

LinenoAddRef and CopyAttributesAstRef access cfg from the node's Base:
```go
func CopyAttributesAstRef(src, dst Node) {
    cfg := src.GetAstConfig()
    dst.SetLineno(cfg.LinenoAddRef(src.GetLineno()))
}
```

**Node interface addition:**
```go
type Node interface {
    // ... existing ...
    GetAstConfig() *AstConfig
}

// Default implementation via Base:
func (b *Base) GetAstConfig() *AstConfig { return b.Cfg }
```

### Functions that become methods on `*AstConfig`:

**Constructors (all become `cfg.New*()`):**
- `NewAtom(rep, terms...)` → `cfg.NewAtom(rep, terms...)`
- `NewApp(rep, args...)` → `cfg.NewApp(rep, args...)`
- `NewVariable(rep, sort)` → `cfg.NewVariable(rep, sort)`
- `NewSequence(stmts...)` → `cfg.NewSequence(stmts...)`
- `NewForall(bounds, body)` → `cfg.NewForall(bounds, body)`
- `NewChoiceAction(branches...)` → `cfg.NewChoiceAction(branches...)`
- `NewLocalAction(args...)` → `cfg.NewLocalAction(args...)`
- `NewCallAction(args...)` → `cfg.NewCallAction(args...)`
- All other `New*` constructors — same pattern

**State methods:**
- `SetReferenceLineno(loc)` → `cfg.SetReferenceLineno(loc)`
- `GetReferenceLineno()` → `cfg.GetReferenceLineno()`
- `LinenoAddRef(loc)` → `cfg.LinenoAddRef(loc)`
- `SetAlwaysCloneWithFreshID(val)` → `cfg.SetAlwaysCloneWithFreshID(val)`
- `nextLFID()` → `cfg.NextLFID()`

**Functions that extract cfg from node (no signature change):**
- `CopyAttributesAstRef(src, dst)` — extracts cfg from src node
- `AstRewrite(x, rewriter)` — extracts cfg from x; also store on rewriter structs as backup
- Clone methods — access cfg from receiver's Base (no change to Node interface Clone signature)

### Files to modify:
- `ast/config.go` (NEW) — AstConfig struct
- `ast/ast.go` — Remove 6 globals; all `New*` functions → methods on `*AstConfig`; add Cfg to Base; add GetAstConfig to Node interface
- `ast/rewrite.go` — CopyAttributesAstRef extracts cfg from node; AstRewrite extracts cfg
- `ast/formula.go` — All `New*` functions → methods; WhenOperator.Clone accesses cfg from Base
- `ast/decl.go` — All `New*` functions → methods; LabeledFormula.Clone uses cfg from Base
- `ast/sort.go` — All `New*` functions → methods
- `ast/tactic.go` — All `New*` functions → methods

### Callers to update (outside ast/):

Call sites change from `ast.NewAtom(rep)` to `cfg.NewAtom(rep)` where `cfg` is the `*ast.AstConfig`. The cfg flows from:
- **lalr_full:** `ParserConfig.AstCfg` → stored on `v17LexAdapter` → grammar actions use `lex.cfg.AstCfg.NewAtom(...)`
- **compiler/isolate/actions:** Get `*AstConfig` from `module.Config.AstCfg`
- **tests:** Create `ast.NewAstConfig()` in test setup

Key call sites:
- `lalr_full/grammar_v17.y` — ALL `ast.New*()` calls (hundreds)
- `lalr_full/inst_mod.go` — SetReferenceLineno, SetAlwaysCloneWithFreshID
- `lalr_full/autoinstance.go` — AstRewrite calls
- `compiler/` — many ast.New* calls, AstRewrite calls
- `actions/` — ast.New*Action calls
- `isolate/` — AstRewrite calls, ast.New* calls
- `ast/*_test.go` — all test files that create nodes

---

## Phase 2: `lalr_full` package — ParserConfig

**New file:** `lalr_full/config.go`

```go
type ParserConfig struct {
    LabelCounter    int
    CheckUnprovable bool
    AstCfg          *ast.AstConfig
}

func NewParserConfig() *ParserConfig {
    return &ParserConfig{
        AstCfg: ast.NewAstConfig(),
    }
}
```

### Globals to migrate (2 variables):

| Global | File:Line | New home |
|--------|-----------|----------|
| `lalrLabelCounter` | grammar_v17.y:25 | `ParserConfig.LabelCounter` |
| `checkUnprovable` | grammar_v17.y:30 | `ParserConfig.CheckUnprovable` |

### Migration pattern:

Store `*ParserConfig` on `v17LexAdapter`. Grammar actions access via `v17lex.(*v17LexAdapter).cfg.LabelCounter`. Helper functions like `newLabel()` and `makeMixinName()` become methods on `*ParserConfig`. `ParseV17` creates and initializes the config instead of resetting globals.

### Files to modify:
- `lalr_full/config.go` (NEW)
- `lalr_full/grammar_v17.y` — Remove 2 `var` declarations; reference via lex.cfg
- `lalr_full/lalr_parser.go` — Add `cfg *ParserConfig` to v17LexAdapter; init in ParseV17

---

## Phase 3: `ivyutils` package — IvyUtilsConfig

**New file:** `ivyutils/config.go`

```go
type IvyUtilsConfig struct {
    Filename         string
    ComposeCharacter string
    ParseErrorList   []error
    StdIncludeDir    string
    RuntimeCaller    func(skip int) (uintptr, string, int, bool)

    // Language version state
    LanguageVersion       string
    HavePolymorphism      bool
    UsePolymorphicMacros  bool
    ForbidGhostInit       bool
    LatestLanguageVersion string
    SymbolCharsParser     *regexp.Regexp

    // Parameter registry
    Registry    *ParameterRegistry
    UseNumerals *BooleanParameter
    UseNewUI    *BooleanParameter
    Catch       *BooleanParameter
    DefaultUI   *Parameter
    EnableDebug *BooleanParameter

    // UI modules
    UIModules   map[string]*UIModule
}

func NewIvyUtilsConfig() *IvyUtilsConfig {
    cfg := &IvyUtilsConfig{
        ComposeCharacter:      ".",
        LanguageVersion:       "1.8",
        HavePolymorphism:      true,
        LatestLanguageVersion: "1.8",
        SymbolCharsParser:     regexp.MustCompile(`[^\[\]\.]*`),
        RuntimeCaller:         runtime.Caller,
        UIModules:             make(map[string]*UIModule),
    }
    cfg.Registry = NewParameterRegistry()
    cfg.UseNumerals = NewBooleanParameterOn(cfg.Registry, "use_numerals", true)
    cfg.UseNewUI = NewBooleanParameterOn(cfg.Registry, "new_ui", false)
    cfg.Catch = NewBooleanParameterOn(cfg.Registry, "catch", true)
    cfg.DefaultUI = NewParameterOn(cfg.Registry, "ui", "cti")
    cfg.EnableDebug = NewBooleanParameterOn(cfg.Registry, "debug", false)
    return cfg
}
```

### Globals to migrate (18 variables):

| Global | File:Line | New home |
|--------|-----------|----------|
| `Filename` | source_file.go:7 | `IvyUtilsConfig.Filename` |
| `ComposeCharacter` | names.go:14 | `IvyUtilsConfig.ComposeCharacter` |
| `ParseErrorListVar` | error.go:185 | `IvyUtilsConfig.ParseErrorList` |
| `stdIncludeDir` | names.go:227 | `IvyUtilsConfig.StdIncludeDir` |
| `runtimeCaller` | names.go:345 | `IvyUtilsConfig.RuntimeCaller` |
| `ivyLanguageVersion` | names.go:129 | `IvyUtilsConfig.LanguageVersion` |
| `IvyHavePolymorphism` | names.go:134 | `IvyUtilsConfig.HavePolymorphism` |
| `IvyUsePolymorphicMacros` | names.go:139 | `IvyUtilsConfig.UsePolymorphicMacros` |
| `IvyForbidGhostInit` | names.go:144 | `IvyUtilsConfig.ForbidGhostInit` |
| `IvyLatestLanguageVersion` | names.go:148 | `IvyUtilsConfig.LatestLanguageVersion` |
| `SymbolCharsParser` | names.go:153 | `IvyUtilsConfig.SymbolCharsParser` |
| `GlobalRegistry` | globals.go:9 | `IvyUtilsConfig.Registry` |
| `UseNumerals` | globals.go:15 | `IvyUtilsConfig.UseNumerals` |
| `UseNewUI` | globals.go:19 | `IvyUtilsConfig.UseNewUI` |
| `Catch` | globals.go:23 | `IvyUtilsConfig.Catch` |
| `DefaultUI` | globals.go:27 | `IvyUtilsConfig.DefaultUI` |
| `EnableDebug` | globals.go:31 | `IvyUtilsConfig.EnableDebug` |
| `uiModules` | globals.go:54 | `IvyUtilsConfig.UIModules` |

### Migration pattern — methodize:

Functions become methods on `*IvyUtilsConfig`:
- `ComposeNames(names...)` → `cfg.ComposeNames(names...)`
- `GetStringVersion()` → `cfg.GetStringVersion()`
- `SetStringVersion(v)` → `cfg.SetStringVersion(v)`
- `GetStdIncludeDir()` → `cfg.GetStdIncludeDir()`
- `WithSourceFile(fname, fn)` → `cfg.WithSourceFile(fname, fn)`
- `ParseWith(parseFn, s)` → `cfg.ParseWith(parseFn, s)`
- `PError(lineno, value, msg)` → `cfg.PError(lineno, value, msg)`
- `RegisterUIModule(name, mod)` → `cfg.RegisterUIModule(name, mod)`
- etc.

**This is the widest-reaching phase.** Nearly every package imports ivyutils. The `IvyUtilsConfig` will be stored on `module.Config.IuCfg`.

### Files to modify in ivyutils/:
- `ivyutils/config.go` (NEW)
- `ivyutils/names.go` — Remove ~11 globals; functions → methods
- `ivyutils/source_file.go` — Remove Filename global
- `ivyutils/error.go` — Remove ParseErrorListVar
- `ivyutils/globals.go` — Remove GlobalRegistry and parameter globals

### Files to modify outside ivyutils/ (broad impact):
- Every package that calls `iu.ComposeNames`, `iu.GetStringVersion`, `iu.Filename`, `iu.ParseErrorListVar`, `iu.IvyHavePolymorphism`, etc.
- `module/config.go` — Add `IuCfg *ivyutils.IvyUtilsConfig` field

---

## Phase 4: `compiler` package

| Global | File:Line | New home |
|--------|-----------|----------|
| `OptMutax` | ivy_compile.go:40 | `module.Config.OptMutax` (bool field) |
| `AdmitDefinitionFactory` | ivy_compile.go:46 | `module.Config.AdmitDefinitionFactory` (func field) |

These move to `module.Config` since compiler already imports module.

### Files to modify:
- `module/config.go` — Add 2 fields
- `compiler/ivy_compile.go` — Remove 2 globals; access via mod.Cfg

---

## Phase 5: `actions` package — ActionsConfig

**Extend existing:** `actions/action.go:1033` already defines `ActionsConfig struct`.

Globals become fields; functions that use them become methods on `*ActionsConfig`:

| Global | File:Line | New home |
|--------|-----------|----------|
| `choiceActionCtr` | action.go:593 | `ActionsConfig.ChoiceActionCtr` |
| `callActionCtr` | action.go:629 | `ActionsConfig.CallActionCtr` |
| `localActionCtr` | action.go:732 | `ActionsConfig.LocalActionCtr` |
| `determinize` | phase3.go:464 | `ActionsConfig.Determinize` |
| `SymexParams` | extra_actions.go:200 | `ActionsConfig.SymexParams` |

---

## Phase 6: `isolate` package

| Global | File:Line | New home |
|--------|-----------|----------|
| `ShowCompiled` etc. | isolate.go:28 | `module.Config` fields (BooleanParameter → bool) |
| `VPrivates` | iter.go:24 | Per-module state or IsolateConfig on Module |
| `IvyVersion` | iter.go:572 | `IvyUtilsConfig.LanguageVersion` (covered by Phase 3) |
| `ExtAction` | create.go:21 | `module.Config.ExtAction` |

---

## Phase 7: `interp` package — InterpConfig

The `interp` package has a global mutex+context pair that must become per-instance:

```go
// interp/interp.go:260-263 — CURRENT:
var (
    contextMu sync.Mutex
    context   = &EvalContext{Check: true}
)
```

**New:** `interp/config.go`

```go
type InterpConfig struct {
    mu      sync.Mutex
    context *EvalContext
}

func NewInterpConfig() *InterpConfig {
    return &InterpConfig{
        context: &EvalContext{Check: true},
    }
}

func (ic *InterpConfig) CurrentContext() *EvalContext {
    ic.mu.Lock()
    defer ic.mu.Unlock()
    return ic.context
}
```

The `Enter()`/`Exit()` methods on `EvalContext` need access to the InterpConfig instead of the global mutex+context. Approach: store `*InterpConfig` on `EvalContext`:

```go
func (ec *EvalContext) Enter(ic *InterpConfig) {
    ic.mu.Lock()
    defer ic.mu.Unlock()
    ec.oldContext = ic.context
    ic.context = ec
}
```

Or better — methodize on InterpConfig:
```go
func (ic *InterpConfig) Enter(ec *EvalContext) {
    ic.mu.Lock()
    defer ic.mu.Unlock()
    ec.oldContext = ic.context
    ic.context = ec
}

func (ic *InterpConfig) Exit(ec *EvalContext) {
    ic.mu.Lock()
    defer ic.mu.Unlock()
    ic.context = ec.oldContext
}
```

| Global | File:Line | New home |
|--------|-----------|----------|
| `contextMu` | interp.go:261 | `InterpConfig.mu` |
| `context` | interp.go:262 | `InterpConfig.context` |

### Files to modify:
- `interp/config.go` (NEW)
- `interp/interp.go` — Remove globals; Enter/Exit/CurrentContext → methods on InterpConfig
- Callers of `interp.CurrentContext()`, `ec.Enter()`, `ec.Exit()`

---

## Phase 8: `transrel` package — TransrelConfig

The `transrel` package has a global mutex+value pair:

```go
// transrel/phase4.go:328-331 — CURRENT:
var (
    useNumeralsMu  sync.RWMutex
    useNumeralsVal = true
)
```

**New:** `transrel/config.go`

```go
type TransrelConfig struct {
    mu             sync.RWMutex
    useNumeralsVal bool
}

func NewTransrelConfig() *TransrelConfig {
    return &TransrelConfig{useNumeralsVal: true}
}

func (tc *TransrelConfig) UseNumerals() bool {
    tc.mu.RLock()
    defer tc.mu.RUnlock()
    return tc.useNumeralsVal
}

func (tc *TransrelConfig) SetUseNumerals(v bool) {
    tc.mu.Lock()
    defer tc.mu.Unlock()
    tc.useNumeralsVal = v
}
```

| Global | File:Line | New home |
|--------|-----------|----------|
| `useNumeralsMu` | phase4.go:328 | `TransrelConfig.mu` |
| `useNumeralsVal` | phase4.go:330 | `TransrelConfig.useNumeralsVal` |

### Files to modify:
- `transrel/config.go` (NEW)
- `transrel/phase4.go` — Remove globals; UseNumerals/SetUseNumerals → methods on TransrelConfig
- Callers of `transrel.UseNumerals()`, `transrel.SetUseNumerals()`

---

## Phase 9: `codegen` package

```go
// codegen/config.go
type CodegenConfig struct {
    TempCounter    int64
    CurrentContext *CodeContext
}
```

Functions using these globals become methods on `*CodegenConfig`.

---

## Phase 10: Remaining packages

| Package | Global | New home |
|---------|--------|----------|
| `tactics` | `UsedSorry` (bool) | `module.Config.UsedSorry` |
| `unionfind` | `ufIDCounter` (int64) | Per-instance counter on UnionFind struct |
| `ranking` | `Debug` (bool) | `module.Config.RankingDebug` |
| `ranking` | `RegisteredTactics` (map) | `proof.Config.Tactics` (already exists) or `module.Config` |
| `l2s` | `Debug` (bool) | `module.Config.L2SDebug` |
| `l2s` | `CurrentModule` (*Module) | Already `module.Config.CurrentModule` |
| `solver` | `HandleRangeSorts` (bool) | `module.Config.HandleRangeSorts` |
| `alpha` | `TestBottom` (bool) | `module.Config.AlphaTestBottom` |
| `alpha` | `Log` (bool) | `module.Config.AlphaLog` |
| `autoinst` | `Verbose` (bool) | `module.Config.AutoinstVerbose` |
| `trace` | `OptionDetailed` (bool) | `module.Config.TraceDetailed` |
| `compose` | `Debug` (bool) | `module.Config.ComposeDebug` |
| `module` | `fallbackPropIDCounter` (int64) | Eliminate — require CompilerConfig (panic if nil) |

---

## Wire-up: module.Config as hub

After all phases, `module.Config` gains references to sub-configs:

```go
type Config struct {
    // ... existing fields ...

    // Sub-package configs
    AstCfg      *ast.AstConfig
    IuCfg       *ivyutils.IvyUtilsConfig
    InterpCfg   *interp.InterpConfig
    TransrelCfg *transrel.TransrelConfig
    CodegenCfg  *codegen.CodegenConfig

    // Moved from compiler package
    OptMutax               bool
    AdmitDefinitionFactory func(mod *Module) func(defn *ast.LabeledFormula, proof ast.Node) error

    // Moved from various packages
    UsedSorry        bool
    HandleRangeSorts bool
    ExtAction        string
    RankingDebug     bool
    L2SDebug         bool
    ComposeDebug     bool
    AlphaTestBottom  bool
    AlphaLog         bool
    AutoinstVerbose  bool
    TraceDetailed    bool
}
```

The `lalr_full.ParserConfig` is NOT on module.Config (lalr_full doesn't import module). Instead, `ParseV17` creates its own `ParserConfig` with a fresh `AstConfig`, and the caller passes the `AstConfig` in via a `WithAstConfig` `ParseOption`.

---

## Implementation Order

1. **Phase 1 (ast)** — Most foundational; all other phases depend on AstConfig existing.
2. **Phase 2 (lalr_full)** — Small, self-contained. Right after Phase 1.
3. **Phase 4 (compiler)** — Small (2 globals). Quick win.
4. **Phase 5 (actions)** — Extend existing ActionsConfig.
5. **Phase 7 (interp)** — Mutex+context elimination.
6. **Phase 8 (transrel)** — Mutex+value elimination.
7. **Phase 10 (remaining small packages)** — Many quick wins: bool flags → module.Config fields.
8. **Phase 6 (isolate)** — Medium complexity.
9. **Phase 9 (codegen)** — Self-contained.
10. **Phase 3 (ivyutils)** — LAST because widest impact. Do after everything else is stable so this massive change is the final step.

---

## Verification

After each phase:
1. `go build ./...` — clean build
2. `go test ./...` — all tests pass (or at least no regressions)
3. `go vet ./...` — no vet warnings
4. Grep `^var ` in modified packages to confirm no mutable globals remain (excluding exclusions)
5. Final: `go test -race ./...` — verify no data races
