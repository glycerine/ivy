# Plan: Switch all top-level parsing to `lalr_full` + fix remaining `LabeledFormula` issues (golden test at 78236)

**Created:** 2026-03-26T01:30

## Context

### Previous fixes (COMPLETED — advanced golden test from 76364 → 78236)

The following were implemented in the previous session:

- **CompileLF method** added to `compiler/compiler.go:725` — matches Python's `LabeledFormula.cmpl` (ivy_compiler.py:411-414), calls `lf.Clone()` to preserve metadata and emit PRESERVE traces
- **DomainSetup.Axiom** updated to use `CompileLF` instead of manual `&ast.LabeledFormula{}`
- **DomainSetup.Property** updated to use `CompileLF`
- **DomainSetup.Derived** updated to use `lf.Clone([]ast.Node{lf.Label, compiled})`
- **DomainSetup.DefinitionDecl** updated to use `lf.Clone([]ast.Node{lf.Label, compiled})`
- **Dispatch traces** added on both Go and Python sides: `compiler.IvyDomainSetup.dispatch`, `compiler.IvyConjectureSetup.dispatch`, `compiler.IvyARGSetup.dispatch`
- **inst_mod traces** added on both sides: `parser.inst_mod.body`, `parser.inst_mod.iter`, `parser.inst_mod.declare`
- **`ast.DeclName()`** added for Python-compatible dispatch name lookup
- **Debug `Sn` field** added to `DefinitionDecl` for tracing

### Current divergence (line 78236)

```
78235  go : XTRACE: compiler.IvyDomainSetup.dispatch name=interpret
       py : XTRACE: compiler.IvyDomainSetup.dispatch name=interpret

78236  go : XTRACE: compiler.IvyDomainSetup.dispatch name=definition
       py : XTRACE: init.ReadModule ENTER file=? nested=False
```

Root cause: `DomainSetup.Interpret` calls `CompileTheory` which parses theory strings. Go uses the hand-rolled parser + `substituteAtomName`. Python uses `read_module` (PLY parser) + `inst_mod`. Different parser + different rewrite → different declaration ordering in nested DomainSetup → trace mismatch.

This is the first point where we switch to `lalr_full`, so we switch ALL top-level parsing at once.

## Part 1: Switch all top-level parsing from hand-rolled to `lalr_full`

### What uses the hand-rolled parser today

**A. Top-level `.Parse()` — MUST switch:**

| # | File | Line | Context |
|---|------|------|---------|
| 1 | `ivyinit/ivyinit.go` | 150 | `ReadModule` fallback (when `!cfg.UseLALRParser`) |
| 2 | `compiler/phase6.go` | 2280 | `CompileTheory` theory parsing |
| 3 | `compiler/phase6.go` | 2571 | `IvyCompileTheoryFromString` theory parsing |
| 4 | `webui/session.go` | 92 | Web UI file loading |
| 5 | `end2end/conform_test.go` | 73 | Conformance test |
| 6 | `end2end/pipeline_test.go` | 32 | Pipeline test |
| 7 | `end2end/debug_init_test.go` | 62 | Debug init test |
| 8 | `compiler/regression_test.go` | 33 | Regression test |

**B. Expression-level `.ParseExpr()` / `.ParseActionBody()` — NOT switching:**

| # | File | Line | Context |
|---|------|------|---------|
| 9 | `logicparser/logicparser.go` | 74 | Pratt parser for v1.7+ formulas |
| 10 | `lalr_logicparser/` tests | various | Cross-validation tests |

Category B is orthogonal — expression-level Pratt parsing stays.

### ParseResult compatibility

Both parsers define identical `ParseResult` structs (different packages):
- `parser.ParseResult{Decls, Modules, Included}`
- `lalr_full.ParseResult{Decls, Modules, Included}`

Switching means changing return types from `*parser.ParseResult` to `*lalr_full.ParseResult`.

### Step 1.1: Add public `lalr_full.InstModSubst` wrapper

In `lalr_full/inst_mod.go`, add a public wrapper for theory substitution:

```go
// InstModSubst applies inst_mod with a substitution map and no prefix.
// Matches Python: ivy = Ivy(); inst_mod(ivy, module, None, subst, dict())
func InstModSubst(decls []ast.Node, subst map[string]string, cfg *ast.AstConfig) []ast.Node {
    module := newIvyAccum(nil, "")
    module.decls = decls
    module.astCfg = cfg
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

### Step 1.2: Change `ReadModule` to always use `lalr_full`

In `ivyinit/ivyinit.go`:
- Change `ReadModule` return type: `*lalr_full.ParseResult` (was `*parser.ParseResult`)
- Remove the `cfg.UseLALRParser` conditional — always use the LALR path
- Remove the hand-rolled parser fallback (lines 149-162)
- Remove `parser` import
- Change `ImportModule` return type similarly

### Step 1.3: Update `ReadModule` callers

`SourceFile` calls `ReadModule` and accesses `.Decls` — this works with either ParseResult type. Update type annotations.

### Step 1.4: Rewrite theory compilation in `compiler/phase6.go`

**`IvyCompileTheoryFromString`:**
```go
func IvyCompileTheoryFromString(mod *module.Module, source string, sort lg.Sort, sortName string) error {
    body, version := parseIvySource(source)
    result, err := lalr_full.Parse(body, version, lalr_full.WithAstConfig(mod.Cfg.AstCfg))
    if err != nil {
        return err
    }
    decls := lalr_full.InstModSubst(result.Decls, map[string]string{"t": sortName}, mod.Cfg.AstCfg)
    return IvyCompileTheory(mod, decls)
}
```

**`CompileTheory`:** Simplify to delegate to `IvyCompileTheoryFromString`.

Remove `substituteAtomName` function and `ivyparser` import.

### Step 1.5: Update `webui/session.go`

Replace `parser.New(src, version)` + `p.Parse()` with `lalr_full.Parse(src, version)`.
Remove `parser` import, add `lalr_full` import.

### Step 1.6: Update test files

Switch `end2end/conform_test.go`, `end2end/pipeline_test.go`, `end2end/debug_init_test.go`, `compiler/regression_test.go` from `parser.New` + `Parse()` to `lalr_full.Parse(src, version)`.

Update imports in `compiler/compiler_test.go` and `compiler/golden_ast_test.go` if they reference `parser.ParseResult`.

### Step 1.7: Remove `UseLALRParser` config flag

Remove `UseLALRParser` field from `module.Config` (or wherever it's defined) and all references.

## Part 2: Fix enum sort Z3 verification crash (TestVerify_EnumExhaustive)

### Symptom
`TestVerify_EnumExhaustive` panics with:
```
z3 bridge panic: Sort mismatch at argument #1 for function (declare-fun and (Bool Bool) Bool) supplied sort is color
```

### What xtracer revealed
Added traces to `solver.ClausesToZ3`. The clauses contain 2 formulas both with `sort={red,green,blue}` (color sort) instead of Boolean:
```
solver.ClausesToZ3 fmla[0] sort={red,green,blue}
 type=*logic.Symbol val=c
```
The formula is bare symbol `c` (color sort), not a Boolean equality like `c = red`.

### Root cause investigation needed
The `after init { c := red }` compiles to an action (ActionDef + MixinAfterDef), not an InitDecl. The init condition is assembled during `CreateIsolate`, not during ARGSetup. The `CreateIsolate` path extracts formulas from action updates and applies them as init conditions. Something in this path produces a bare `color`-sorted symbol instead of a Boolean formula.

### Plan
1. Add xtracer traces to `CreateIsolate`'s init condition assembly path (both Go and Python)
2. Add traces to `transrel/phase4.go` SmallModelClauses
3. Add traces to `check/check.go` checkFcsNormalPath
4. Compare Go vs Python trace output to find where the init clause diverges
5. Fix the compilation or clause assembly bug

### Files to instrument
| File | Trace points |
|------|-------------|
| `compiler/phase6.go` | CreateIsolate init condition assembly |
| `isolate/isolate.go` | get_isolate_exports, CreateIsolate |
| `transrel/phase4.go` | SmallModelClauses clause construction |
| `check/check.go` | checkFcsNormalPath, CheckConjsInState |
| Python equivalents | Matching traces on Python side |

## Part 3: Architecture notes for future reference

### Parser architecture (DO NOT change)
- `lalr_full` = file parser (declarations). Used for all `.ivy` file parsing. **High confidence** — validated by golden test.
- `lalr_logicparser` = formula parser (expressions/terms/actions in isolation). Separate by design, matching Python's `ivy_logic_parser.py` vs `ivy_parser.py` split.
- `logicparser` = dispatcher that routes to lalr_logicparser (v1.6−) or hand-rolled Pratt parser (v1.7+)
- The hand-rolled Pratt parser is still used for v1.7+ formula parsing — switching to lalr_logicparser's v17 grammar for this would be a future task

### Version gating
- Python: single parser with 42 version-gated grammar rules (30 in ivy_parser, 12 in ivy_logic_parser)
- Go: separate grammar files per version band (v1.2, v1.6, v1.7), no conditional rules inside grammars
- Full version gating implementation is a separate future task

## Part 4: Remaining `&ast.LabeledFormula{}` bare struct literals (D5 from previous plan)

These bare struct constructions bypass `cfg.NewLabeledFormula()` and miss metadata/counter management. They fall into two categories:

**New constructions** (should use `cfg.NewLabeledFormula()`):
- `compiler/phase6.go:1938` — `goal := &ast.LabeledFormula{Formula: cond}`
- `compiler/phase6.go:2071` — `newProp := &ast.LabeledFormula{...}`
- `compiler/ivy_compile.go:450` — `lf := &ast.LabeledFormula{Formula: compiled}`
- `compiler/ivy_compile.go:586` — `lf := &ast.LabeledFormula{Formula: compiled}`
- `compiler/ivy_compile.go:674` — `mlf := &ast.LabeledFormula{...}`
- `compiler/ivy_compile.go:1788` — `result := &ast.LabeledFormula{...}`
- `compiler/decl.go:1259` — `clf := &ast.LabeledFormula{...}`
- `compiler/decl.go:1311` — `proofLF := &ast.LabeledFormula{...}`
- `compiler/decl.go:1338` — `lastLF := &ast.LabeledFormula{...}`
- `compiler/decl.go:1390` — `lastLF := &ast.LabeledFormula{Formula: d.LastFact}`
- `compiler/decl.go:1422` — `mlf := &ast.LabeledFormula{...}`

Each of these should be checked against the Python equivalent to determine if Python uses `LabeledFormula(label, formula)` (constructor → use `cfg.NewLabeledFormula()`) or `lf.clone([...])` (clone → use `lf.Clone()`).

**Note:** This is lower priority than Part 1. The golden test divergence is caused by Part 1. Part 2 fixes will be needed when those code paths are reached during golden test progression.

## Files to Modify

| File | Change |
|------|--------|
| `lalr_full/inst_mod.go` | Add `InstModSubst()` wrapper (~15 lines) |
| `ivyinit/ivyinit.go` | Always use `lalr_full.Parse`; change return types |
| `compiler/phase6.go` | Use `lalr_full.Parse` + `InstModSubst`; remove `substituteAtomName` and `ivyparser` import |
| `webui/session.go` | Use `lalr_full.Parse` |
| `end2end/conform_test.go` | Switch to `lalr_full.Parse` |
| `end2end/pipeline_test.go` | Switch to `lalr_full.Parse` |
| `end2end/debug_init_test.go` | Switch to `lalr_full.Parse` |
| `compiler/regression_test.go` | Switch to `lalr_full.Parse` |
| `compiler/compiler_test.go` | Update imports |
| `compiler/golden_ast_test.go` | Update imports |
| `module/config.go` (or equivalent) | Remove `UseLALRParser` field |
| `compiler/phase6.go:1938,2071` | Part 2: bare LF construction → `cfg.NewLabeledFormula()` |
| `compiler/ivy_compile.go:450,586,674,1788` | Part 2: bare LF construction |
| `compiler/decl.go:1259,1311,1338,1390,1422` | Part 2: bare LF construction |

**NOT modified:**
- `logicparser/logicparser.go` — uses `ParseExpr`, not `Parse`
- `lalr_logicparser/` test files — cross-validation tests
- `parser/` package itself — kept for `ParseExpr`/`ParseActionBody`

## Verification

```bash
# Build everything
go build ./...

# Run compiler tests
go test ./compiler/... -count=1

# Run end2end tests
go test ./end2end/... -count=1

# Run the golden test (primary validation)
cd ~/goivy && make golden

# Run full test suite
go test ./... -count=1
```

Golden test should advance past line 78236 because theory compilation now uses `lalr_full` + `instMod`.
