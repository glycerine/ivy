# Plan: Complete De-globalization — Remove All Transitional Globals

**Created:** 2026-03-25T10:30

## Context

The previous session created per-package Config structs and added `cfg.NewFoo()` method versions of all constructors, but left behind `Default*Config` transitional globals and legacy wrapper functions. Every package still uses the old `ast.NewAtom()` / `ast.SetReferenceLineno()` / etc. APIs that delegate to `DefaultAstConfig`. This plan completes the migration by:

1. Threading `*AstConfig` through all callers so they use `cfg.NewFoo()` instead of `ast.NewFoo()`
2. Removing `DefaultAstConfig` and all legacy wrappers once no callers remain
3. Doing the same for interp, transrel, codegen, isolate, ivyutils, actions, and lalr_full

---

## Inventory of Transitional Globals to Remove

### A. `ast.DefaultAstConfig` (ast/ast.go:59)
The biggest — ~920 external callers across lalr_full (442), compiler (393), parser (189), and others.

**Legacy wrapper functions to delete (in ast/):**
- `SetReferenceLineno()` → callers use `cfg.SetReferenceLineno()`
- `GetReferenceLineno()` → callers use `cfg.GetReferenceLineno()`
- `LinenoAddRef()` → callers use `cfg.LinenoAddRef()` or extract from node
- `SetAlwaysCloneWithFreshID()` → callers use `cfg.SetAlwaysCloneWithFreshID()`
- All 113 standalone `NewFoo()` functions that wrap `DefaultAstConfig`

**Internal ast/ uses of `LinenoAddRef` (must extract cfg from node):**
- `rewrite.go:209` — `CopyAttributesAstRef`: use `src.GetAstConfig().LinenoAddRef()`
- `rewrite.go:729` — AstRewrite default case: use `x.GetAstConfig().LinenoAddRef()`
- `formula.go:345` — `WhenOperator.Clone`: use `w.GetAstConfig().LinenoAddRef()`
- `ast.go:461` — `Variable.Resort`: use `v.GetAstConfig().LinenoAddRef()`

### B. `interp.DefaultInterpConfig` (interp/interp.go:275)
3 wrapper functions, ~3 internal callers + tests.

### C. `transrel.DefaultTransrelConfig` (transrel/phase4.go:341)
2 wrapper functions, low external usage.

### D. `codegen.DefaultCodegenConfig` (codegen/codegen.go:117)
Plus globals: `tempCounter`, `currentContext`.

### E. `isolate.DefaultIsolateConfig` (isolate/isolate.go:65)
Plus 15 transitional global vars (ShowCompiled, ConeOfInfluence, etc.) and 3 more in iter.go/create.go.

### F. `ivyutils.DefaultIvyUtilsConfig` (ivyutils/config.go:68)
Plus 18 transitional globals (Filename, ComposeCharacter, etc.).

### G. `actions` counter globals (actions/action.go)
`choiceActionCtr`, `callActionCtr`, `localActionCtr` — plus `determinize`.

### H. `lalr_full` globals (grammar_v17.y)
`lalrLabelCounter`, `checkUnprovable` — currently synced with ParserConfig.

### I. `compiler` globals (ivy_compile.go)
`OptMutax`, `AdmitDefinitionFactory` — already on module.Config but legacy globals still exist.

---

## Implementation Strategy

**Key insight:** The cfg flows top-down from entry points:
- **Parse entry:** `ParseV17` creates `ParserConfig` with `AstCfg` → grammar actions get it via `lex.cfg.AstCfg`
- **Compile entry:** `IvyCompile` gets `*module.Module` → `mod.Cfg.AstCfg`
- **Check entry:** `StartCheck` gets `*module.Module` → same path
- **Tests:** Create `ast.NewAstConfig()` locally

The migration proceeds in dependency order: fix ast-internal uses first, then lalr_full (biggest caller), then compiler, then remaining packages. Finally delete the globals.

---

## Step 1: Fix ast-internal LinenoAddRef calls

These 4 call sites currently call the global `LinenoAddRef()`. Change them to extract cfg from the node being operated on.

**Files:** `ast/rewrite.go`, `ast/formula.go`, `ast/ast.go`

| Call site | Current | New |
|-----------|---------|-----|
| `rewrite.go:209` CopyAttributesAstRef | `LinenoAddRef(src.GetLineno())` | `src.GetAstConfig().LinenoAddRef(src.GetLineno())` |
| `rewrite.go:729` AstRewrite default | `LinenoAddRef(x.GetLineno())` | `x.GetAstConfig().LinenoAddRef(x.GetLineno())` |
| `formula.go:345` WhenOperator.Clone | `LinenoAddRef(w.GetLineno())` | `w.GetAstConfig().LinenoAddRef(w.GetLineno())` |
| `ast.go:461` Variable.Resort | `LinenoAddRef(v.GetLineno())` | `v.GetAstConfig().LinenoAddRef(v.GetLineno())` |

**Safety:** If `GetAstConfig()` returns nil (node created before config system), add nil guard that falls back to identity (no wrapping). This matches the behavior when `referenceLineno` is zero.

```go
func safeLinenoAddRef(n Node, loc Location) Location {
    if cfg := n.GetAstConfig(); cfg != nil {
        return cfg.LinenoAddRef(loc)
    }
    return loc
}
```

---

## Step 2: Migrate lalr_full (442 calls) — the biggest caller

### 2a. Thread AstCfg into grammar actions

The grammar actions run inside `v17Parse(lex)` where `lex` is the `*v17LexAdapter` and `lex.cfg.AstCfg` is the `*ast.AstConfig`. Grammar actions access `lex` as `v17lex.(*v17LexAdapter)`.

**Approach:** Define a local helper at the top of the grammar's Go preamble:

```go
// In grammar_v17.y preamble:
func astCfg(lex v17Lexer) *ast.AstConfig {
    return lex.(*v17LexAdapter).cfg.AstCfg
}
```

Then in every grammar action, replace:
- `ast.NewAtom(...)` → `astCfg(v17lex).NewAtom(...)`
- `ast.NewSequence(...)` → `astCfg(v17lex).NewSequence(...)`
- etc.

This is mechanical: find-replace `ast.New` with `astCfg(v17lex).New` in grammar_v17.y, then regenerate.

**Exception:** Some helper functions called from grammar actions (like `newLabel`, `addLabel`, `makeMixinName`, `handleMixin`, `createObject`) also call `ast.New*`. These need the config passed in. Since they're already in the grammar preamble, they can take `*ast.AstConfig` as a parameter, or take `v17Lexer` and call `astCfg()`.

### 2b. Eliminate lalrLabelCounter and checkUnprovable globals

Currently `ParseV17` syncs globals ↔ config. Instead, make grammar actions access `lex.cfg.LabelCounter` directly:
- `lalrLabelCounter++` → `v17lex.(*v17LexAdapter).cfg.LabelCounter++`
- `checkUnprovable` → `v17lex.(*v17LexAdapter).cfg.CheckUnprovable`

Or define helpers:
```go
func parserCfg(lex v17Lexer) *ParserConfig {
    return lex.(*v17LexAdapter).cfg
}
```

### 2c. Migrate inst_mod.go

`instMod` calls `ast.SetReferenceLineno()` and `ast.SetAlwaysCloneWithFreshID()`. These should use the AstConfig. The ivyAccum already exists in the function — thread `*ast.AstConfig` through:

```go
func instMod(ivy *ivyAccum, module *ivyAccum, pref *ast.Atom, subst map[string]string, vsubst map[string]*ast.Variable, modname string, lineno ...ast.Location) {
    cfg := ast.DefaultAstConfig // TODO: get from ivy or caller
    cfg.SetAlwaysCloneWithFreshID(true)
    defer cfg.SetAlwaysCloneWithFreshID(false)
    ...
    cfg.SetReferenceLineno(refLineno)
    ...
}
```

Better: add `astCfg *ast.AstConfig` field to `ivyAccum`, set it when the accumulator is created in `ParseV17`, and access as `ivy.astCfg` in `instMod`.

### 2d. Migrate autoinstance.go

Similar to inst_mod — thread AstCfg through `expandAutoInstances` from the lex adapter.

### 2e. Migrate test files

Tests in lalr_full/ that call `ast.New*()` or `ast.SetReferenceLineno()` need a local `cfg := ast.NewAstConfig()` and use `cfg.NewAtom()` etc.

---

## Step 3: Migrate parser/ package (189 calls)

Add `cfg *ast.AstConfig` field to the `Parser` struct. Initialize it in `New()`:

```go
type Parser struct {
    ...
    cfg *ast.AstConfig
}

func New(input string, version lexer.Version) *Parser {
    return &Parser{
        ...
        cfg: ast.NewAstConfig(),
    }
}
```

Then replace all `ast.NewAtom(...)` → `p.cfg.NewAtom(...)` in parser/decl.go, parser/expr.go, parser/action.go, parser/parser.go.

---

## Step 4: Migrate compiler/ package (393 calls)

Compiler functions receive `*module.Module`. Access config as `mod.Cfg.AstCfg`.

**Production files (small count):** compiler.go, phase6.go, decl.go, helpers.go, ivy_compile.go — extract `cfg := mod.Cfg.AstCfg` and use `cfg.NewFoo()`.

**Test files (large count):** Each test creates a module via `module.New()` which gives `mod.Cfg.AstCfg`. Extract and use.

Also remove the legacy `compiler.OptMutax` and `compiler.AdmitDefinitionFactory` globals — update the single caller in `check/check.go:37` to set on `mod.Cfg` instead.

---

## Step 5: Migrate remaining packages

### 5a. isolate/ (1 direct ast.New* call + many isolate globals)

Replace `ast.NewAtom("iso")` etc. with config-based call. The bigger task is migrating the 15 transitional bool globals to `IsolateConfig` fields:

**Approach:** Each isolate function that reads a global gets an `*IsolateConfig` parameter (or accesses it via `mod.IsolateCfg` on Module). Then internal references change from `ConeOfInfluence` to `cfg.ConeOfInfluence`.

Since isolate is self-contained (only isolate/ reads these globals), this is a package-internal refactor. Add `IsolateCfg *IsolateConfig` to `module.Config`, and thread it through isolate functions.

### 5b. actions/ (counter globals)

The `actions.NewChoiceAction()` etc. still use `atomic.AddInt64(&choiceActionCtr, ...)`. Callers (compiler, isolate, l2s, gogen tests) need to migrate to `cfg.NewChoiceAction()` where `cfg` is `*ActionsConfig`. Thread it through from module.

### 5c. interp/ (3 calls)

`CurrentContext()` in eval.go → pass `*InterpConfig` to eval functions.

### 5d. transrel/ (0-1 external calls)

May already be done. Verify and remove wrappers.

### 5e. codegen/ (tempCounter, currentContext)

Thread `*CodegenConfig` through codegen functions. The `Enter()`/`Exit()` pattern on CodeContext needs the config.

### 5f. ivyutils/ (18 globals, widest reach)

`iu.ComposeNames()` is used in ~18 files. Add `*IvyUtilsConfig` to `module.Config` (already done as `IuCfg`), then migrate callers to use `mod.Cfg.IuCfg.ComposeNames()`. Some internal-to-ivyutils functions can use the config directly.

---

## Step 6: Delete all transitional globals and wrappers

Once all callers are migrated, delete:

### ast/ast.go:
- `var DefaultAstConfig`
- `func SetReferenceLineno(lineno Location)` wrapper
- `func GetReferenceLineno() Location` wrapper
- `func LinenoAddRef(loc Location) Location` wrapper
- `func SetAlwaysCloneWithFreshID(val bool)` wrapper
- All 113 standalone `func NewFoo(...)` that wrap DefaultAstConfig

### interp/interp.go:
- `var DefaultInterpConfig`
- `func (ec *EvalContext) Enter()` legacy wrapper
- `func (ec *EvalContext) Exit()` legacy wrapper
- `func CurrentContext() *EvalContext` legacy wrapper

### transrel/phase4.go:
- `var DefaultTransrelConfig`
- `func UseNumerals() bool` legacy wrapper
- `func SetUseNumerals(v bool)` legacy wrapper

### codegen/codegen.go:
- `var DefaultCodegenConfig`
- `var tempCounter int64`
- `var currentContext *CodeContext`

### isolate/isolate.go:
- `var DefaultIsolateConfig`
- All 15 transitional `var` declarations (ShowCompiled, etc.)

### isolate/iter.go + create.go:
- `var VPrivates`, `var IvyVersion`, `var ExtAction`

### ivyutils/:
- `var DefaultIvyUtilsConfig`
- All 18 transitional globals in names.go, source_file.go, error.go, globals.go

### actions/action.go:
- `var choiceActionCtr`, `var callActionCtr`, `var localActionCtr`

### actions/phase3.go:
- `var determinize`

### lalr_full/grammar_v17.y:
- `var lalrLabelCounter`, `var checkUnprovable`

### compiler/ivy_compile.go:
- `var OptMutax`, `var AdmitDefinitionFactory`

### module/config.go:
- No deletions — this is the destination, not a source

---

## Implementation Order

1. **Step 1** — ast-internal LinenoAddRef (4 edits, trivial)
2. **Step 2** — lalr_full migration (biggest, ~450 mechanical edits + helper refactoring)
3. **Step 3** — parser/ migration (189 mechanical edits)
4. **Step 4** — compiler/ migration (393 edits, mostly tests)
5. **Step 5a-f** — Remaining packages (isolate, actions, interp, transrel, codegen, ivyutils)
6. **Step 6** — Delete all transitional globals and legacy wrappers

Each step should leave the code building and tests passing.

---

## Files to Modify (by step)

### Step 1:
- `ast/rewrite.go` (2 edits)
- `ast/formula.go` (1 edit)
- `ast/ast.go` (1 edit + add `safeLinenoAddRef` helper)

### Step 2:
- `lalr_full/grammar_v17.y` (~174 edits + helper functions)
- `lalr_full/grammar_v17.go` (regenerated)
- `lalr_full/inst_mod.go` (~5 edits)
- `lalr_full/autoinstance.go` (~3 edits)
- `lalr_full/ivy_module.go` (add astCfg field)
- `lalr_full/lalr_parser.go` (remove sync-globals pattern)
- `lalr_full/*_test.go` (many test files)

### Step 3:
- `parser/parser.go` (add cfg field)
- `parser/decl.go` (~121 edits)
- `parser/expr.go` (~26 edits)
- `parser/action.go` (~28 edits)

### Step 4:
- `compiler/ivy_compile.go` (remove globals, use mod.Cfg)
- `compiler/compiler.go`, `compiler/phase6.go`, `compiler/decl.go`, `compiler/helpers.go`
- `compiler/*_test.go` (many test files)
- `check/check.go` (update AdmitDefinitionFactory setter)

### Step 5:
- `isolate/*.go` (thread IsolateConfig)
- `actions/action.go`, `actions/phase3.go`, `actions/extra_actions.go`
- `interp/interp.go`, `interp/eval.go`
- `transrel/phase4.go`
- `codegen/codegen.go`, `codegen/types.go`
- `ivyutils/names.go`, `ivyutils/source_file.go`, `ivyutils/error.go`, `ivyutils/globals.go`
- `module/config.go` (add IsolateCfg)

### Step 6:
- All files listed in the deletion inventory above

---

## Verification

After each step:
1. `go build ./...` — clean build
2. `go test ./ast/ ./lalr_full/ ./compiler/ ./parser/ ./module/ ./interp/ ./transrel/ ./codegen/ ./isolate/ ./actions/ ./ivyutils/` — all pass
3. `go vet ./...` — no warnings
4. After Step 6: `grep -r 'DefaultAstConfig\|DefaultInterpConfig\|DefaultTransrelConfig\|DefaultCodegenConfig\|DefaultIsolateConfig\|DefaultIvyUtilsConfig' --include='*.go'` returns nothing
5. Final: `go test -race ./...` — no data races
