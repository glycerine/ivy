# Plan: T.1 Integration Tests for Full Verification Pipeline

## Context

The goivy port has ~25K lines of Go across 73 packages with 60 unit test files, but zero integration tests that run the full verify pipeline: parse → compile → isolate → check → Z3. Task T.1 from MISSING.md requires adding these tests.

During investigation, we found and fixed a critical bug: the parser's `parseOneRel` function wasn't setting `ASort = "bool"` on relation declarations (matching Python `ivy_parser.py:970`), causing relation sorts to have `TopSort` range instead of `BooleanSort`. **This is now fixed** in `parser/decl.go:270`.

A remaining conformance issue exists: `TestConformCheck` in `webui/backend_conform_test.go` — the Go induction checker finds a false counterexample for `link(X,Y) -> ~semaphore(Y)` while Python correctly proves it inductive. This is a separate semantic bug in the action/transition-relation pipeline (`trace.MakeCheckArt` / `trace.CheckFinalCond`), to be investigated as a follow-on.

## What was fixed so far

1. **TestPrune** (updr) — was failing due to Z3 version mismatch; now passes with the custom Z3 4.7.1 fork.
2. **Relation sort bug** — `parser/decl.go:270`: added `atom.ASort = ast.NewSymbol("bool", nil)` matching Python's `p[1].sort = 'bool'`. Relations like `relation link(X:client, Y:server)` now correctly get `FunctionSort(client, server -> Boolean)` instead of `FunctionSort(client, server -> TopSort)`.

## Integration Test Plan

### File: `/Users/jaten/go/src/github.com/glycerine/goivy/integ_test/pipeline_test.go`

New package `integ_test` (external test package) to avoid circular imports. Tests cut across `ivyinit`, `compiler`, `isolate`, `check`, `solver`, `z3bridge`.

### Test Categories

#### Category A: Parse + Compile (no Z3)
Verify .ivy source → AST → compiled module with correct sorts/signatures.

1. **TestParseCompile_ClientServer** — parse `client_server_example.ivy`, verify module has 2 sorts, 2 relations with Boolean range, 2 actions, 1 conjecture
2. **TestParseCompile_EnumTypes** — `type color = {red,green,blue}`, verify EnumeratedSort
3. **TestParseCompile_RelationSort** — verify `relation r(X:t)` produces `FunctionSort(t -> Boolean)` not TopSort
4. **TestParseCompile_MultipleConjectures** — verify conjecture formulas have proper sorts

#### Category B: Full Pipeline (with Z3)
Parse → compile → check → Z3 verification.

5. **TestVerify_TrivialPass** — simple invariant that trivially holds (e.g., `invariant true`)
6. **TestVerify_TrivialFail** — `assert false` in an exported action, should fail
7. **TestVerify_ClientServer** — the full client-server example (requires fix to MakeCheckArt first)
8. **TestVerify_PropertyFromAxioms** — property provable from axioms alone

#### Category C: Randomized Go-vs-Python Conformance
Cross-validate Go and Python verification results on the same .ivy inputs.

9. **TestConformRandomized** — generate random .ivy files with relations, actions, and conjectures; run both Go and Python pipelines; compare pass/fail. Uses the existing `PyBackend` sidecar infrastructure.
10. **TestConformPythonTestSuite** — run Go pipeline against a curated subset of the Python test files from `~/pyivy/ivy/test/*.ivy`; compare against expected results.

### Key files to modify/create

| File | Action |
|------|--------|
| `integ_test/pipeline_test.go` | New — Categories A & B tests |
| `integ_test/conform_test.go` | New — Category C randomized conformance |
| `integ_test/testdata/*.ivy` | New — curated .ivy test fixtures |
| `integ_test/gen_ivy.go` | New — random .ivy file generator for fuzzing |

### Reusable functions (already exist)

- `ivyinit.ReadModule()` / `ivyinit.SourceFile()` — file loading (`ivyinit/ivyinit.go`)
- `compiler.New()` + `compiler.NewDeclInterp()` + `di.ProcessDecl()` — compilation (`compiler/compiler.go`, `compiler/decl.go`)
- `parser.New(src, version).Parse()` — parsing (`parser/parser.go`)
- `check.CheckIsolate(mod, traceHook)` — verification (`check/isolate_check.go`)
- `check.CheckFcsInStateWithAG()` — formula checking with Z3 (`check/check.go:361`)
- `solver.New()` + `slv.ClausesToZ3()` — solver (`solver/solver.go`)
- `webui.NewPyBackend()` / `webui.NewGoBackend()` — conformance infrastructure (`webui/backend_*.go`)
- `trace.MakeCheckArt()` / `trace.CheckFinalCond()` — check art builder (`trace/trace.go`)

### Test helper pattern

```go
func compileIvy(t *testing.T, src string) *module.Module {
    t.Helper()
    version := lexer.Version{1, 7}
    p := parser.New(src, version)
    decls, err := p.Parse()
    require.NoError(t, err)
    sig := il.NewSig()
    mod := module.New()
    mod.Sig = sig
    cmplr := compiler.New(sig, mod)
    di := compiler.NewDeclInterp(cmplr)
    for _, decl := range decls {
        require.NoError(t, di.ProcessDecl(decl))
    }
    return mod
}
```

### Randomized .ivy generator spec

Generate .ivy files with:
- 1-3 uninterpreted sorts
- 1-5 relations over those sorts (using `relation` keyword)
- 0-2 individual constants
- 1-3 actions with `require`/`ensure`/assignments
- `export` for all actions
- 1-3 conjectures (some valid, some intentionally invalid)

Run both Go and Python pipelines, compare `pass`/`fail` results.

### Prerequisite: fix TestConformCheck semantic bug

Before Category B and C tests will be meaningful, the remaining conformance bug must be fixed. The Go `MakeCheckArt`/`CheckFinalCond` pipeline is returning SAT (counterexample found) for conjectures that Python correctly proves UNSAT (inductive). Root cause investigation needed in:
- `trace/trace.go:MakeCheckArt()` — how it builds the pre/post state
- `art/art.go:Execute()` + `PostState()` — how actions compute transition relations
- `actions/update.go` — action update semantics

## Verification

1. `go test ./integ_test/... -v` — all integration tests pass
2. `go test ./... -count=1` — no regressions in existing tests
3. `go test ./integ_test -run TestConformRandomized -count=100` — randomized conformance
4. Compare Go vs Python on `~/pyivy/ivy/test/*.ivy` files
