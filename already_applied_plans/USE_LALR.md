# Plan: Switch all top-level parsing from hand-rolled parser to `lalr_full` (golden test divergence at 78236)

**Created:** 2026-03-26T01:15

## Context

Golden test diverges at line 78236 because Go's `CompileTheory` uses the hand-rolled parser (`parser` package) instead of `lalr_full`. Since `lalr_full` is now the validated parser (golden test compares its output against Python), all production code that parses `.ivy` source should use `lalr_full.Parse`. Switching piecemeal would create inconsistencies, so we switch ALL top-level parsing call sites at once.

### What uses the hand-rolled parser today

Three categories:

**A. Top-level `.Parse()` — must switch to `lalr_full`:**
1. `ivyinit/ivyinit.go:150` — `ReadModule` fallback path (when `!cfg.UseLALRParser`)
2. `compiler/phase6.go:2280` — `CompileTheory` theory parsing
3. `compiler/phase6.go:2571` — `IvyCompileTheoryFromString` theory parsing
4. `webui/session.go:92` — Web UI file loading

**B. Expression-level `.ParseExpr()` / `.ParseActionBody()` — NOT switching:**
5. `logicparser/logicparser.go:74` — `parseStringHandWritten` for v1.7+ formulas
6. Test files in `lalr_logicparser/` — cross-validation tests

Category B uses the Pratt parser for individual expressions/formulas, not full file parsing. This is orthogonal to the LALR switch and should stay as-is.

**C. Test code using `.Parse()` — switch to `lalr_full`:**
7. `end2end/conform_test.go:73`
8. `end2end/pipeline_test.go:32`
9. `end2end/debug_init_test.go:62`
10. `compiler/compiler_test.go:11` (import only)
11. `compiler/regression_test.go:33`
12. `compiler/golden_ast_test.go:21` (import only)

### ParseResult compatibility

Both parsers define `ParseResult` with identical fields:
```go
type ParseResult struct {
    Decls    []ast.Node
    Modules  map[string]*ast.ModuleDecl
    Included map[string]bool
}
```
They're in different packages (`parser.ParseResult` vs `lalr_full.ParseResult`). The switchover means changing the return type of `ReadModule` and all callers to use `lalr_full.ParseResult`.

### Theory compilation must also switch to `instMod`

Python's `ivy_compile_theory_from_string` calls `inst_mod(ivy, module, None, {'t':sortname}, dict())`.
Go's `CompileTheory` uses `substituteAtomName` (simple atom rename). After switching to `lalr_full.Parse`, we also need to use `instMod` for theory rewriting, matching Python exactly.

## Implementation Plan

### Step 1: Add public `lalr_full.InstModSubst` wrapper

Add to `lalr_full/inst_mod.go`:
```go
// InstModSubst applies inst_mod with a substitution map and no prefix.
// Matches Python: ivy = Ivy(); inst_mod(ivy, module, None, subst, dict())
func InstModSubst(decls []ast.Node, subst map[string]string, cfg *ast.AstConfig) []ast.Node {
    module := newIvyAccum(nil, "")
    module.decls = decls
    module.astCfg = cfg
    // init maps to avoid nil panics
    module.modules = make(map[string]*ast.ModuleDecl)
    module.macros = make(map[string]ast.Node)
    module.actions = make(map[string]ast.Node)
    module.included = make(map[string]bool)

    ivy := newIvyAccum(nil, "")
    ivy.astCfg = cfg
    instMod(ivy, module, nil, subst, nil, "", 0)
    return ivy.decls
}
```

### Step 2: Change `ReadModule` return type to `lalr_full.ParseResult`

In `ivyinit/ivyinit.go`:
- Change `ReadModule` signature: return `*lalr_full.ParseResult` instead of `*parser.ParseResult`
- Remove the `cfg.UseLALRParser` conditional — always use `lalr_full.Parse`
- Remove the conversion code (lines 140-144) that converts between ParseResult types
- Remove the hand-rolled parser fallback (lines 149-162)
- Remove the `parser` import

### Step 3: Update `ReadModule` callers

`ReadModule` is called from:
- `ivyinit/ivyinit.go` itself (in `ImportModule`)
- `ivyinit/ivyinit.go:220` (in `SourceFile`)

Check if `SourceFile` passes the result to `compiler.IvyCompile` — it accesses `.Decls` which is the same field on both ParseResult types. Update type annotations as needed.

Also update `ImportModule` which returns `*parser.ParseResult` — change to `*lalr_full.ParseResult`.

### Step 4: Update `compiler/phase6.go` — theory compilation

Replace the `ivyparser` import with `lalr_full`:

**`CompileTheory`** (line 2263):
```go
func CompileTheory(mod *module.Module, sortname string, theoryname string) error {
    // ... get theoryStr (unchanged) ...
    return IvyCompileTheoryFromString(mod, theoryStr, sort, sortname)
}
```

**`IvyCompileTheoryFromString`** (line 2565):
```go
func IvyCompileTheoryFromString(mod *module.Module, source string, sort lg.Sort, sortName string) error {
    body, version := parseIvySource(source)

    // Parse with lalr_full (matching Python's read_module)
    result, err := lalr_full.Parse(body, version)
    if err != nil {
        return err
    }

    // Apply inst_mod substitution (matching Python's inst_mod(ivy, module, None, {'t':sortname}, {}))
    decls := lalr_full.InstModSubst(result.Decls, map[string]string{"t": sortName}, mod.Cfg.AstCfg)

    // Compile via DomainSetup
    return IvyCompileTheory(mod, decls)
}
```

Remove `substituteAtomName` function (line 2588) if no longer used.
Remove `ivyparser` import.

### Step 5: Update `webui/session.go`

Replace hand-rolled parser with `lalr_full.Parse`:
```go
import lalr_full "github.com/glycerine/goivy/lalr_full"

// In LoadFileContent:
result, err := lalr_full.Parse(src, version)
if err != nil { ... }
decls := result.Decls
```

Remove `parser` import.

### Step 6: Update test files

For each test file that uses `parser.New(src, version)` + `p.Parse()`:

**`end2end/conform_test.go`:**
```go
result, err := lalr_full.Parse(src, version)
```

**`end2end/pipeline_test.go`:**
```go
result, err := lalr_full.Parse(src, version)
```

**`end2end/debug_init_test.go`:**
```go
result, err := lalr_full.Parse(src, version)
```

**`compiler/regression_test.go`:**
```go
result, err := lalr_full.Parse(src, lexer.Version{1, 7})
```

**`compiler/compiler_test.go`** and **`compiler/golden_ast_test.go`:** update imports if they reference `parser.ParseResult`.

### Step 7: Remove `UseLALRParser` config flag

Since `lalr_full` is now always used, remove the `UseLALRParser` field from `module.Config` and all references to it.

### Step 8: Keep `logicparser` unchanged

The `logicparser` package uses `parser.ParseExpr()` for expression-level parsing (not full file). This stays as-is — it's a different concern (Pratt expression parsing vs. LALR file parsing). The `lalr_logicparser` cross-validation tests also stay unchanged.

## Files to Modify

| File | Change |
|------|--------|
| `lalr_full/inst_mod.go` | Add `InstModSubst()` public wrapper (~15 lines) |
| `ivyinit/ivyinit.go` | Always use `lalr_full.Parse`; change return types; remove parser import |
| `compiler/phase6.go` | Use `lalr_full.Parse` + `InstModSubst` for theory; remove `ivyparser` import and `substituteAtomName` |
| `webui/session.go` | Use `lalr_full.Parse`; remove parser import |
| `end2end/conform_test.go` | Switch to `lalr_full.Parse` |
| `end2end/pipeline_test.go` | Switch to `lalr_full.Parse` |
| `end2end/debug_init_test.go` | Switch to `lalr_full.Parse` |
| `compiler/regression_test.go` | Switch to `lalr_full.Parse` |
| `compiler/compiler_test.go` | Update imports if needed |
| `compiler/golden_ast_test.go` | Update imports if needed |
| `module/config.go` | Remove `UseLALRParser` field |

**NOT modified:**
- `logicparser/logicparser.go` — uses `ParseExpr`, not `Parse`
- `lalr_logicparser/` test files — cross-validation tests, keep both parsers
- `parser/` package itself — keep it for `ParseExpr`/`ParseActionBody` use

## AstConfig threading for theory parsing

`lalr_full.Parse` accepts `WithAstConfig(cfg)` option. For theory compilation, we need to pass the module's AstConfig so counters stay in sync. Update `CompileTheory` / `IvyCompileTheoryFromString` to pass `lalr_full.WithAstConfig(mod.Cfg.AstCfg)`.

## Verification

```bash
# Build everything
go build ./...

# Run compiler tests
go test ./compiler/... -count=1

# Run end2end tests
go test ./end2end/... -count=1

# Run the golden test (the primary validation)
cd ~/goivy && make golden

# Run full test suite
go test ./... -count=1
```

Golden test should advance past line 78236 because theory compilation now uses the same parser and `instMod` path as Python.
