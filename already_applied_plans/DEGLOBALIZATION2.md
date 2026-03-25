# Plan: Eliminate Mutable Global Variables — Refactor into Config System

**Created:** 2026-03-25T06:00

## Context

The goivy port intentionally avoids global/package-level variables to enable multi-tenancy: running thread pools of ivy models safely on multi-core machines. Recent mechanical porting efforts introduced new globals (acceptable to get basics working). This refactoring pass re-establishes the no-globals discipline by moving all mutable package-level state into the existing Config system.

**Design principle:** Each package has its own Config struct. Config is preferably the method receiver (creating methods); alternatively passed as a parameter. `module.Config` is the top-level config; other package configs are embedded in or referenced from it.

---

## Exclusions (leave as-is)

- **All vprint.go globals** (every package) — debug facilities, not used in production
- **goyacc-generated parser tables** (v17Act, v17Chk, v17R1, v17R2, v17Tok1, v17Tok2, v17Tok3, v17Pact, v17Pgo, v17Def, v17Exca, v17Toknames, v17Statenames, v17ErrorMessages, v17Debug, v17ErrorVerbose) — read-only after init or goyacc infrastructure
- **Compile-time interface checks** (`var _ Node = (*Type)(nil)`) — zero-value
- **Compiled regexps that never change** (incDirPat in names.go:231, puncsRe in cppgen) — immutable after init
- **Singleton constants** (ast.Equals, logic.Boolean, logic.TopS, logic.True, logic.False, ivylogic.Alpha/Beta/Gamma, ivylogic.Equals) — immutable after init
- **sync.Mutex instances** (TsPrintfMut, contextMu, useNumeralsMu) — synchronization primitives
- **Read-only maps** (lexer.allReserved, lexer.tokenNames, logic.infixSymbols, logic.precSymbols, compiler.DefinedAttributes, compiler.KnownLogics, ivylogic.Logics, ivylogic.DecidableLogics, ivylogic.DefaultLogics, ivylogic.macroExpansions, ivylogic.PolymorphicMacrosMap, ivylogic.InfixSymbols, ivylogic.PrecSymbols, ivylogic.UninterpretedPolymorphicSymbols, ivylogic.polymorphicSymbols, ivylogic.polymorphicSymbolsDef, solver.z3Builtins, cppgen.SpecialNames, cppgen.CppTypesByTitle, dafnygen.OpMap, dafnygen.ArithOps, compose.RequiredFields, parser.DefaultVersion) — populated once at init
- **Test-only data** (pyIvyNoError in golden_test.go)
- **Blank imports** (`var _ = fmt.Sprint`)
- **xtracer.Enabled, xtracer.HashVerbose** — build-tag controlled debug facility (same rationale as vprint.go)
- **ivyutils.PolymorphicSymbols** — populated once at init, read-only after

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

### Migration pattern — Store `*AstConfig` on `Base` struct:

**Decision:** Every AST node carries a `*AstConfig` pointer via its embedded `Base` struct. This adds 8 bytes per node but makes cfg universally accessible in Clone methods without changing the `Node` interface.

```go
// In Base struct (ast.go):
type Base struct {
    Loc Location
    Cfg *AstConfig  // NEW — pointer to shared config
}
```

Constructors set cfg:
```go
func NewChoiceAction(cfg *AstConfig, branches ...Node) *ChoiceAction {
    cfg.ChoiceActionCounter++
    ca := &ChoiceAction{Branches: branches, UniqueID: cfg.ChoiceActionCounter}
    ca.Cfg = cfg
    return ca
}
```

Clone methods access cfg via receiver's Base:
```go
func (c *ChoiceAction) Clone(args []Node) Node {
    c.Cfg.ChoiceActionCounter++
    return &ChoiceAction{Base: c.Base, Branches: args, UniqueID: c.Cfg.ChoiceActionCounter}
}
```

LinenoAddRef and CopyAttributesAstRef access cfg from the node's Base:
```go
func CopyAttributesAstRef(src, dst Node) {
    cfg := src.GetBase().Cfg  // access config from source node
    dst.SetLineno(cfg.LinenoAddRef(src.GetLineno()))
}
```

**Node interface change:** Add `GetBase() *Base` method (or `GetAstConfig() *AstConfig`). Since Base is already embedded in every node type, `GetBase()` is trivially implementable and many types already have access.

Alternatively, add a `GetAstConfig()` method to the Node interface:
```go
type Node interface {
    // ... existing ...
    GetAstConfig() *AstConfig
}

// Default implementation via Base:
func (b *Base) GetAstConfig() *AstConfig { return b.Cfg }
```

### Functions that change signature:

- `NewChoiceAction(branches)` → `NewChoiceAction(cfg *AstConfig, branches ...Node)`
- `NewLocalAction(args)` → `NewLocalAction(cfg *AstConfig, args ...Node)`
- `NewCallAction(args)` → `NewCallAction(cfg *AstConfig, args ...Node)`
- `NewAtom(rep)` → `NewAtom(cfg *AstConfig, rep string)` (and all other constructors that create nodes)
- `SetReferenceLineno(loc)` → `cfg.SetReferenceLineno(loc)` (method on *AstConfig)
- `GetReferenceLineno()` → `cfg.GetReferenceLineno()`
- `LinenoAddRef(loc)` → `cfg.LinenoAddRef(loc)` (method on *AstConfig)
- `SetAlwaysCloneWithFreshID(val)` → `cfg.SetAlwaysCloneWithFreshID(val)`
- `nextLFID()` → `cfg.NextLFID()`
- `CopyAttributesAstRef(src, dst)` — extracts cfg from src node's Base
- `AstRewrite(x, rewriter)` — extracts cfg from x's Base; also store on rewriter structs as backup
- Clone methods — unchanged signature, access cfg from receiver's Base

**Important:** ALL node constructors (NewAtom, NewApp, NewVariable, NewSequence, NewForall, etc.) must accept `*AstConfig` and store it on Base. This is a large but mechanical change.

### Files to modify:
- `ast/config.go` (NEW) — AstConfig struct
- `ast/ast.go` — Remove 6 globals; update constructors, LinenoAddRef, etc.
- `ast/rewrite.go` — Add AstCfg to rewriter structs, update CopyAttributesAstRef, AstRewrite
- `ast/formula.go` — WhenOperator.Clone uses cfg from rewriter context
- `ast/decl.go` — LabeledFormula.Clone uses cfg

### Callers to update (outside ast/):

Since ALL node constructors now take `*AstConfig`, every call site in the codebase that creates AST nodes must pass cfg. The cfg flows from:
- **lalr_full:** `ParserConfig.AstCfg` → stored on `v17LexAdapter` → grammar actions pass to constructors
- **compiler/isolate/actions:** Get `*AstConfig` from `module.Config.AstCfg`
- **tests:** Create a `NewAstConfig()` in test setup

Key call sites:
- `lalr_full/grammar_v17.y` — ALL `ast.New*()` calls (hundreds), SetReferenceLineno, SetAlwaysCloneWithFreshID
- `lalr_full/inst_mod.go` — SetReferenceLineno, SetAlwaysCloneWithFreshID
- `lalr_full/autoinstance.go` — AstRewrite calls
- `compiler/` — New*Action calls, AstRewrite calls, many ast.New* calls
- `actions/` — New*Action calls
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

Store `*ParserConfig` on `v17LexAdapter`. Grammar actions access via `v17lex.(*v17LexAdapter).cfg.LabelCounter`. `ParseV17` creates and initializes the config instead of resetting globals.

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

### Migration pattern:

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

**New file or extend existing:** `actions/config.go`

Note: `actions/action.go:1033` already defines `ActionsConfig struct`. Extend it.

```go
// Extend existing ActionsConfig with counter fields:
type ActionsConfig struct {
    // ... existing fields ...
    ChoiceActionCtr int64
    CallActionCtr   int64
    LocalActionCtr  int64
    Determinize     bool
    SymexParams     []lg.Expr
}
```

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

## Phase 7: `codegen` + `cppgen` packages

```go
// codegen/config.go
type CodegenConfig struct {
    TempCounter    int64
    CurrentContext *CodeContext
}

// cppgen/config.go
type CppgenConfig struct {
    ThunkCounter int64
    TempCtr      int64
    NondetCnt    int64
    TheClassname string
    SkipZ3       bool
    IndentLevel  int
}
```

---

## Phase 8: Remaining packages

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
    AstCfg     *ast.AstConfig
    IuCfg      *ivyutils.IvyUtilsConfig

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
5. **Phase 8 (remaining small packages)** — Many quick wins: bool flags → module.Config fields.
6. **Phase 6 (isolate)** — Medium complexity.
7. **Phase 7 (codegen/cppgen)** — Self-contained.
8. **Phase 3 (ivyutils)** — LAST because widest impact. Do after everything else is stable so this massive change is the final step.

---

## Verification

After each phase:
1. `go build ./...` — clean build
2. `go test ./...` — all tests pass (or at least no regressions)
3. `go vet ./...` — no vet warnings
4. Grep `^var ` in modified packages to confirm no mutable globals remain (excluding exclusions)
5. Final: `go test -race ./...` — verify no data races
