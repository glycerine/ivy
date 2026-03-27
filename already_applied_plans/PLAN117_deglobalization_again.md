# BUG F & NC 1: Transitional Globals Migration + UsedSymbolsAST Return Type

**Created:** 2026-03-27 21:00

## Context

BUG F: Seven transitional mutable globals still exist despite Config fields being ready. These violate the "No Global Variables" rule (CLAUDE.md Section C) and break multi-tenancy when running concurrent Ivy sessions. The Config system is fully in place — fields exist, constructor methods exist — the globals just haven't been deleted and callers haven't been redirected.

NC 1: `UsedSymbolsAST()`, `Clauses.Symbols()`, and `logicutil.UsedConstants()` return `map[NodeKey]lg.Expr` but only ever store `*lg.Symbol` values. This forces 44 callers to bare-cast values to `*lg.Symbol`. Python's `used_symbols_ast` returns a set of Symbol objects directly. Changing the Go return type aligns with Python, adds compile-time safety, and eliminates all 44 bare casts.

---

## BUG F: Transitional Globals Migration

### Phase F1: ivyutils globals (5 variables, EASY)

These 5 globals in `ivyutils/names.go` already have corresponding `IvyUtilsConfig` fields with method-based `SetStringVersion()`. The globals are written only by the free-function `SetStringVersion()` (lines 155-174) and read by 6 production callers plus tests.

#### F1.1: Make free functions delegate to DefaultIvyUtilsConfig

**File:** `ivyutils/names.go`

The free functions `GetStringVersion()`, `SetStringVersion()`, `GetNumericVersion()`, `GetStdIncludeDir()` currently read/write the globals directly. Change them to delegate to `DefaultIvyUtilsConfig`:

```go
// GetStringVersion delegates to DefaultIvyUtilsConfig.
func GetStringVersion() string {
    return DefaultIvyUtilsConfig.GetStringVersion()
}

// SetStringVersion delegates to DefaultIvyUtilsConfig.
func SetStringVersion(version string) {
    DefaultIvyUtilsConfig.SetStringVersion(version)
}

// GetNumericVersion delegates to DefaultIvyUtilsConfig.
func GetNumericVersion() []int {
    return DefaultIvyUtilsConfig.GetNumericVersion()
}
```

Then **delete** the 5 globals:
- `ivyLanguageVersion` (line 129) — replaced by `DefaultIvyUtilsConfig.LanguageVersion`
- `IvyHavePolymorphism` (line 132) — replaced by `DefaultIvyUtilsConfig.HavePolymorphism`
- `IvyUsePolymorphicMacros` (line 135) — replaced by `DefaultIvyUtilsConfig.UsePolymorphicMacros`
- `IvyForbidGhostInit` (line 138) — replaced by `DefaultIvyUtilsConfig.ForbidGhostInit`
- `SymbolCharsParser` (line 144) — replaced by `DefaultIvyUtilsConfig.SymbolCharsParser`

Also delete `IvyLatestLanguageVersion` (line 141) — same pattern.

#### F1.2: Redirect ivylogic/ callers to use config

6 production callers in ivylogic/ read the exported globals directly:

| File | Line | Global | Replacement |
|------|------|--------|-------------|
| `ivylogic/sig.go` | 97 | `iu.IvyHavePolymorphism` | `iu.DefaultIvyUtilsConfig.HavePolymorphism` |
| `ivylogic/poly.go` | 96 | `iu.IvyHavePolymorphism` | `iu.DefaultIvyUtilsConfig.HavePolymorphism` |
| `ivylogic/poly.go` | 119 | `iu.IvyHavePolymorphism` | `iu.DefaultIvyUtilsConfig.HavePolymorphism` |
| `ivylogic/ivylogic.go` | 386 | `iu.IvyUsePolymorphicMacros` | `iu.DefaultIvyUtilsConfig.UsePolymorphicMacros` |
| `ivylogic/classify_ext.go` | 346 | `iu.IvyUsePolymorphicMacros` | `iu.DefaultIvyUtilsConfig.UsePolymorphicMacros` |

These 5 callers continue using `DefaultIvyUtilsConfig` for now. **Later**, when these functions gain config parameters (or become methods on `IvyLogicConfig`), they can be threaded properly.

#### F1.3: Update tests

- `ivyutils/names_extra_test.go` — Tests that read globals (`IvyHavePolymorphism`, `IvyUsePolymorphicMacros`, `IvyForbidGhostInit`) should read from `DefaultIvyUtilsConfig` fields instead.
- `ivylogic/new_funcs_test.go:278,299` — Tests that write `iu.IvyUsePolymorphicMacros = true` should write `iu.DefaultIvyUtilsConfig.UsePolymorphicMacros = true`.

Also add `GetNumericVersion()` and `GetStdIncludeDir()` methods to `IvyUtilsConfig` if they don't already exist (check config.go).

---

### Phase F2: isolate globals (2 variables, HARDER)

These globals require threading Config through function signatures because `CreateIsolate` and helpers don't currently receive Config.

#### F2.1: `ExtAction` — `isolate/create.go:21`

**Current state:**
- Global `var ExtAction = ""` at `create.go:21`
- Read in `CreateIsolate()` at lines 303, 328, 330
- Config field already exists: `module.Config.ExtAction` (config.go:109) and `IsolateConfig.ExtAction` (isolate.go:48)
- `Module.Cfg` is `*module.Config` — accessible from `CreateIsolate`'s `mod` parameter

**Fix:** Replace all reads of the global with reads from `mod.Cfg.ExtAction`:

```go
// Line 303: if ExtAction != ""  →  if mod.Cfg != nil && mod.Cfg.ExtAction != ""
// Line 328: mod.Actions[ExtAction]  →  mod.Actions[mod.Cfg.ExtAction]
// Line 330: mod.PublicActions[ExtAction]  →  mod.PublicActions[mod.Cfg.ExtAction]
```

Then delete `var ExtAction = ""` at line 21.

Guard with `mod.Cfg != nil` check to avoid nil panic when called without config.

Update test `isolate/batch_fixes_test.go:651-653` to set `mod.Cfg.ExtAction = "ext"` instead of the global.

#### F2.2: `VPrivates` — `isolate/iter.go:23`

**Current state:**
- Global `var VPrivates = make(map[string]bool)` at `iter.go:23`
- Written in `helpers.go:133,136,153` (inside `SetupVerifiedNames()` or similar)
- Read in `iter.go:28` by `VStartsWithSomeRec(name, prefixes, mod)`
- Config field: `IsolateConfig.VPrivates` at `isolate.go:46`

**Fix:** Add an `IsolateConfig` parameter (or use `mod.Cfg`) to both writer and reader functions:

1. **Writer** (`helpers.go:128-157`): The function that populates `VPrivates` already receives `mod *module.Module`. Store on `mod` instead of global:

```go
// helpers.go — instead of writing to global VPrivates:
if mod.Cfg != nil {
    // Use a local map, store it on module or isolate config
    vprivates := make(map[string]bool)
    for _, isol := range mod.Isolates {
        for _, v := range isol.VerifiedNames() {
            vprivates[v] = true
        }
    }
    // Store on module for later access
    mod.VPrivates = vprivates  // or pass through IsolateConfig
}
```

2. **Reader** (`iter.go:27-28`): `VStartsWithSomeRec` already receives `mod *module.Module`. Read from `mod` instead of global:

```go
func VStartsWithSomeRec(name string, prefixes map[string]bool, mod *module.Module) bool {
    vprivates := mod.VPrivates  // or get from IsolateConfig if stored there
    if mod.Privates[name] || vprivates[name] {
        return false
    }
```

**Module struct change needed:** Add `VPrivates map[string]bool` field to `module.Module` (or use existing `IsolateConfig.VPrivates` if the config is accessible). Since `Module.Cfg` has no `IsolateConfig` field, the simplest approach is to add `VPrivates` directly to `Module`.

Then delete `var VPrivates` at `iter.go:23`.

---

## NC 1: UsedSymbolsAST Return Type Change

### Current state

Three functions return `map[NodeKey]lg.Expr` but only store `*lg.Symbol`:

| Function | File | Line |
|----------|------|------|
| `usedSymbolsAST` | `clauseops/clauses.go` | 301 |
| `UsedSymbolsAST` | `clauseops/astutil.go` | 27 |
| `Clauses.Symbols()` | `clauseops/clauses.go` | 179 |
| `logicutil.UsedConstants` | `logicutil/logicutil.go` | 195 |

44 callers bare-cast values to `*lg.Symbol`. Python returns `set` of Symbol objects.

### NC 1.1: Change internal helpers

**File:** `clauseops/clauses.go`

```go
// Change symbolsASTRec signature:
func symbolsASTRec(node lg.Expr, result map[lg.NodeKey]*lg.Symbol) {
    switch t := node.(type) {
    case *lg.Symbol:
        result[lg.Key(t)] = t  // now stores *lg.Symbol directly
    case *lg.Apply:
        if c, ok := t.Func.(*lg.Symbol); ok {
            result[lg.Key(c)] = c  // now stores *lg.Symbol directly
        } else {
            symbolsASTRec(t.Func, result)
        }
        for _, arg := range t.Terms {
            symbolsASTRec(arg, result)
        }
        return
    }
    for _, c := range node.Children() {
        symbolsASTRec(c, result)
    }
}

// Change usedSymbolsAST:
func usedSymbolsAST(node lg.Expr) map[lg.NodeKey]*lg.Symbol {
    result := make(map[lg.NodeKey]*lg.Symbol)
    symbolsASTRec(node, result)
    return result
}
```

**File:** `clauseops/astutil.go`

```go
func UsedSymbolsAST(node lg.Expr) map[lg.NodeKey]*lg.Symbol {
    return usedSymbolsAST(node)
}
```

**File:** `clauseops/clauses.go` — `Symbols()` method:

```go
func (c *Clauses) Symbols() map[lg.NodeKey]*lg.Symbol {
    result := make(map[lg.NodeKey]*lg.Symbol)
    for _, f := range c.Fmlas {
        for s, sym := range usedSymbolsAST(f) {
            result[s] = sym
        }
    }
    for _, d := range c.Defs {
        for s, sym := range usedSymbolsAST(d) {
            result[s] = sym
        }
    }
    return result
}
```

### NC 1.2: Change logicutil.UsedConstants

**File:** `logicutil/logicutil.go`

```go
func UsedConstants(t logic.Expr) map[logic.NodeKey]*logic.Symbol {
    result := make(map[logic.NodeKey]*logic.Symbol)
    usedConstantsRec(t, result)
    return result
}

func usedConstantsRec(t logic.Expr, result map[logic.NodeKey]*logic.Symbol) {
    switch n := t.(type) {
    case *logic.Symbol:
        result[logic.Key(n)] = n
    default:
        for _, c := range t.Children() {
            usedConstantsRec(c, result)
        }
        // Binder variables...
        switch b := n.(type) {
        case *logic.ForAll:
            for _, v := range b.Variables {
                usedConstantsRec(v, result)
            }
        // ... same for Exists, Lambda, NamedBinder
        }
    }
}
```

### NC 1.3: Update 44 callers — remove bare casts

All callers that iterate values and cast `v.(*lg.Symbol)` can now use `v` directly since the map value type IS `*lg.Symbol`. This eliminates 44 bare type assertions across:

| Package | Files | Caller Count |
|---------|-------|-------------|
| `mc/` | `toaiger.go`, `mine.go` | 11 |
| `proof/` | `phase5_matching.go`, `phase5_goals.go`, `checker.go`, `goal.go` | 7 |
| `actions/` | `action.go` | 2 |
| `transrel/` | `transrel.go`, `interpolant.go` | 3 |
| `clauseops/` | `litclause.go`, `subsume.go`, `ops.go` | 9 |
| `solver/` | `solver.go`, `clauses.go`, `model.go`, `herbrand.go`, `compat.go` | 8 |
| `vmt/` | `vmt.go` | 1 |
| `check/` | `check.go` | 1 |
| `ivylogic/` | `classify_ext.go`, `sortinfer.go` | 3 |
| `webui/` | `concept_isession.go`, `concept_domain.go` | 4 |

**Special case — `proof/phase5_goals.go:199-210` (`FmlaVocab`):**

This function returns `map[lg.NodeKey]lg.Expr` because it merges both Symbols and Variables. The merge at line 202-203 still works because `*lg.Symbol` satisfies `lg.Expr`:
```go
result := make(map[lg.NodeKey]lg.Expr)
for k, v := range co.UsedSymbolsAST(fmla) {
    result[k] = v  // *lg.Symbol → lg.Expr: implicit interface satisfaction
}
```
No change needed here except removing the cast.

**Special case — `proof/phase5_goals.go:415,433` and `solver/compat.go:277-280`:**

These use `:=` to capture the return. After the change, the inferred type becomes `map[NodeKey]*lg.Symbol`. Merging two such maps works. At line 428, `usedSyms[lg.Key(sym)]` is a key-existence check — works regardless.

### NC 1.4: Update tests

- `clauseops/clauseops_test.go:404` — adjust if test explicitly types the return
- `logicutil/logicutil_test.go:144,156` — adjust if test explicitly types the return

---

## EXECUTION ORDER

1. **NC 1** first (return type changes) — touches many files but each change is mechanical (remove cast)
   - NC 1.1: Change `symbolsASTRec`, `usedSymbolsAST`, `UsedSymbolsAST`, `Symbols()` return types
   - NC 1.2: Change `UsedConstants`, `usedConstantsRec` return types
   - NC 1.3: Update all 44 callers — remove bare `.(*lg.Symbol)` casts
   - NC 1.4: Update tests

2. **BUG F Phase F1** (ivyutils globals) — straightforward delegation
   - F1.1: Make free functions delegate to `DefaultIvyUtilsConfig`
   - F1.2: Redirect ivylogic/ callers
   - F1.3: Update tests
   - Delete globals

3. **BUG F Phase F2** (isolate globals) — requires adding fields
   - F2.1: `ExtAction` — read from `mod.Cfg.ExtAction` instead of global
   - F2.2: `VPrivates` — add field to Module, thread through
   - Delete globals

## VERIFICATION

After each phase:
```bash
go build ./...
go test ./clauseops/ ./logicutil/ ./solver/ ./proof/ ./mc/ ./vmt/ ./check/ ./actions/ ./transrel/ ./ivylogic/ ./webui/ ./isolate/ ./ivyutils/ ./module/
cd ~/goivy && make golden   # should not regress from line 124642
```
