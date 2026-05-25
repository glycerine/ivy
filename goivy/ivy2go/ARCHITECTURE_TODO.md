# ivy2go — Architecture and Implementation TODO

Created: 2026-05-25 06:06:56 UTC

## 1. Context

The Go port `~/ivy/goivy/ivy2cpp/` is now a feature-complete mechanical
port of pyivy's `ivy_to_cpp.py` (32,913 LOC across 38 production + 18
test files, oracle-verified against pyivy on 14 fixtures). It emits C++
source — including an "automatic test generation" target (`target=test`
/ `target=gen`) that uses Z3 at runtime to synthesise action arguments
satisfying preconditions, then logs traces for behavioral comparison.

We now want `~/ivy/goivy/ivy2go/`: a sibling Go package, written in Go,
that emits **Go** source code with the same semantics as ivy2cpp's
output. Strategically this enables:

- Standalone Go binaries for Ivy protocols (no C++ toolchain, no MSVC
  registry lookup, no Z3 link gymnastics — `go build` suffices).
- Reuse of the mature, well-tested `goivy.Solver` / `goivy.Translator`
  / `goivy.smt.Z3Solver` stack already in goivy: emitted Go programs
  `import "github.com/glycerine/ivy/goivy"` and call into that facade
  for SMT, rather than linking Z3's C++ API.
- A second oracle perspective on goivy itself: divergences between
  ivy2cpp-emitted and ivy2go-emitted programs surface bugs in either.

The Python source of truth remains `~/ivy/pyivy/ivy/ivy/ivy_to_cpp.py`.
Per `goivy/CLAUDE.md` section A, ivy2cpp mirrors that file mechanically.
**ivy2go in turn mirrors ivy2cpp**, then substitutes Go-flavored
emission for C++-flavored emission at the leaves of the tree (the
emit-* functions). Architectural structure, function names, struct
names, file names, and call graph all match ivy2cpp.

## 2. Confirmed design decisions

| # | Decision | Choice |
|---|----------|--------|
| D1 | Z3 at *generator* time | Use `goivy.Solver` / `Translator` (already comprehensive in `~/ivy/goivy/z3bridge_*.go`). Never call Z3 directly. |
| D2 | Z3 at *generated-program* runtime | Generated `.go` files `import "github.com/glycerine/ivy/goivy"` and call `goivy.NewSolver`, `goivy.Translator`, `Solver.GetSmallModel`, etc. directly. No separate `ivy2go/runtime` re-export shim. |
| D3 | Relation to existing `~/ivy/goivy/gogen/` | Coexist independently. `gogen` is a separate older/simpler experiment; `ivy2go` is the ivy2cpp-equivalent and the long-term path. No code shared, no migration. |
| D4 | Output layout per emitted module | Generated **Go package directory** with multiple files: `types.go`, `state.go`, `actions.go`, `runtime.go`, `main.go` (and `init.go`, plus per-isolate sub-files when needed). |
| D5 | Test/oracle strategy | **In-process unit tests on emitted text only.** Mirror `ivy2cpp_test.go`'s pattern: parse Ivy fixture → `Generate()` → assert string-shape properties on emitted Go. No behavioral oracle vs ivy2cpp at MVP. No golden files. (May add `go vet` / `gofmt` static checks; see §3.9.) |
| D6 | Generated Go style | Go 1.25 features are fair game. **Avoid generics** (slow at runtime); prefer concrete types or interface dispatch. **Iterators** (range-over-func, `iter.Seq`) are fine. |

These are firm; the rest of this document is built on top of them.

## 3. Architecture overview

### 3.1 Scope and non-goals

**In scope.**

- A new Go package `github.com/glycerine/ivy/goivy/ivy2go` that takes
  a `*goivy.Module` (already compiled by goivy's front-end) and emits
  a directory of Go source files implementing the module's runtime
  semantics.
- All five targets ivy2cpp supports: `impl`, `class`, `repl`, `test`,
  `gen`. The semantics carry over unchanged; only the emitted language
  changes.
- A `Build()` step that runs `go build` on the emitted package to
  produce a binary (analogue of ivy2cpp's `BuildOutput()` → g++/cl).
- Comprehensive in-process unit tests that mirror `ivy2cpp_test.go`'s
  organisation and per-feature coverage.
- File-for-file structural parity with ivy2cpp so audit cross-references
  stay tractable.

**Out of scope (initial release).**

- Behavioral oracle harness comparing ivy2go and ivy2cpp binaries on
  shared fixtures. (Architected for as a follow-on; see §3.9.)
- Golden Go fixture files.
- WASM/wasm-build targets (orthogonal; goivy already has `z3vendor/wasm_lib`).
- Migration of `gogen/` (independent per D3).
- Cross-pollination with `dafnygen/` (separate target language).

**Non-goals (permanent).**

- Direct Z3 use (always through `goivy.Solver`).
- Generated code depending on generics for hot-path operations.
- Emitting Go that requires manual post-edit before `go build`.

### 3.2 Top-level pipeline & entry points

Mirrors `ivy2cpp/compile.go` and `ivy2cpp/generator.go`:

```
filename
   │
   ▼
CompileAndGenerate(filename, params, Config)
   │  ├─ goivy.New() + goivy.SourceFile() — parse + frontend
   │  ├─ mergeParams() — split goivy params from ivy2go cfg
   │  ├─ selectedIsolates() — fan out per isolate
   │  └─ for each isolate:
   │       isoMod = mod.Copy(); CreateIsolate(); prepareModuleForGo()
   │       Generate(isoMod, outCfg) → *Output
   ▼
*Output { Package string, Files map[string]string, BaseName, PackageName, Target, Config, ExtraFiles, LibSpecs }
   │
   ▼
WriteOutput(out, outDir)        // emit .go files to disk
   │
   ▼
Build(out, outDir) → BuildPlan  // invoke `go build`; returns binary path
```

The two principal types — `Config` and `Output` — mirror ivy2cpp's
`Config` and `Output` exactly, **substituting Go concepts for C++ ones**:

| ivy2cpp `Output` field | ivy2go `Output` field | Notes |
|------------------------|------------------------|-------|
| `Header string`         | (removed)              | Go has no header/impl split. |
| `Impl string`           | `Files map[string]string` | Generated `.go` files keyed by basename. |
| `BaseName string`       | `BaseName string`      | Module base name. |
| `ClassName string`      | `PackageName string`   | Go has packages, not classes. (Also `StateTypeName string` for the `*State` receiver.) |
| `Target string`         | `Target string`        | Same vocabulary. |
| `EffectiveTarget string`| `EffectiveTarget string` | Same. |
| `EmitMain bool`         | `EmitMain bool`        | Same. |
| `Config Config`         | `Config Config`        | Same. |
| `ExtraFiles map[string]string` | `ExtraFiles map[string]string` | Same. |
| `LibSpecs []string`     | `LibSpecs []string`    | Same (informational; `go build` doesn't need them). |

`Config` mirrors ivy2cpp's `Config` 1:1 with these substitutions:

- `ClassName` → `PackageName` (a Go identifier; default derived from
  base name via `varName()` lowering).
- `Compiler` → unused; deleted. Go has one official toolchain.
- `HostOS` is retained (some generated code may still want
  `runtime.GOOS` branches, e.g., for socket I/O in the REPL target).
- `Stdafx bool` — deleted (Windows precompiled-header artefact).
- New: `GoModule string` — when non-empty, written as `module <path>`
  in a generated `go.mod`; default empty (caller may supply or skip).
- New: `GoivyImportPath string` — defaults to
  `"github.com/glycerine/ivy/goivy"`; allows tests/forks to override.

### 3.3 Output package layout

Per D4, each Ivy module → one Go package directory. Files emitted:

| File | Contents | Counterpart in ivy2cpp |
|------|----------|------------------------|
| `go.mod` | Module declaration if `Config.GoModule != ""`. Else not emitted (caller manages). | (none; ivy2cpp lacks an analogue) |
| `types.go` | Sort declarations: enum types as `type Foo int` + `const (FooA Foo = iota; …)`; range types as `type Foo int`; uninterpreted as `type Foo int`; destructor records as `type Bar struct { … }`; variants as tagged-union structs. | `types.go` + `cpp_types.go` + `destructor.go` + `variant.go` |
| `state.go` | `State` struct with one field per signature symbol (scalar / `[N]T` array / `map[K]V` / hash-thunk). `NewState() *State`. | `generator.emitHeader` member emission (generator.go:232–289) |
| `actions.go` | Methods on `*State` for every Ivy action: `func (s *State) ActFoo(arg1 T1) T2 { … }`. Loops, ifs, assigns, calls, asserts/assumes. | `action.go` + `assign.go` + `action_gen.go` (test/gen variant) |
| `init.go` | `(s *State) Init()` — initial state; nondeterministic via `s.choose…()` for repl/impl/class, solver-driven for test/gen. | `init.go` + `initial_state.go` |
| `runtime.go` | Free helpers: `ivyAssert`, `ivyAssume`, `ivyChoose`, `ivyPrintWrite`, `ivyReadCmd`. For test/gen: solver helpers wrapping `goivy.NewSolver` + `goivy.Translator.FormulaToZ3`. For repl: command parser, dispatch loop. | `runtime.go` + `repl.go` + `tick.go` + `vprint.go` + `solver_emit.go` + `z3.go` |
| `nondet.go` | Nondeterministic helpers (havoc, choose). Per-sort sampling functions. | `nondet.go` |
| `extensional.go` | Extensional relation iteration helpers. | `extensional.go` |
| `definitions.go` | Pure-function emission for definitional axioms. | `definitions.go` |
| `native.go` | User-supplied native Go blocks (Ivy `<<< ... >>>` with a `go` tag — see §3.5.10). | `native.go` + `native_thunk.go` |
| `main.go` | Only when `EmitMain && PackageName == "main"`. `func main()` invoking repl loop or test loop. | (mixed into repl.go / runtime.go in ivy2cpp) |
| `_test.go` files in generated package | **Not emitted by ivy2go**. The generated package is shipped; tests live in `ivy2go/` itself. | n/a |

Implementation files **inside** the `ivy2go` package (not in the output)
mirror ivy2cpp's file layout 1:1 to make audits and grep cross-tractable:

| ivy2go source file | Role | ivy2cpp counterpart |
|--------------------|------|---------------------|
| `writer.go` | `goWriter` — indented Go output buffer with `line/linef/raw/blank/open/close`. Knows Go's brace style. | `writer.go` |
| `go_context.go` | `GoContext` — multi-stream output (per `.go` file in the output package). | `cpp_context.go` |
| `go_types.go` | `goType()` — Ivy sort → Go type string. | `cpp_types.go` |
| `types.go` | Sort declaration emission. | `types.go` |
| `expr.go` | `emitExpr()` — Ivy formula/term → Go expression string. | `expr.go` |
| `bv_expr.go` | Bitvector lowering (see §3.5.7). | `bv_expr.go` |
| `action.go` | `emitAction()` — Ivy action → Go statements. | `action.go` |
| `assign.go` | Assignment lowering (scalar / quantified / two-phase / hash-thunk). | `assign.go` |
| `nondet.go` | `___ivy_choose` Go analogue. | `nondet.go` |
| `extensional.go` | Extensional relation handling. | `extensional.go` |
| `variant.go` | Variant (tagged union) emission. | `variant.go` |
| `destructor.go` | Destructor / record struct emission. | `destructor.go` |
| `constructors.go` | Sort constructor emission. | `constructors.go` |
| `definitions.go` | Definitional axiom emission. | `definitions.go` |
| `initial_state.go` | Initial state population. | `initial_state.go` |
| `init.go` | `Init()` method emission. | `init.go` |
| `runtime.go` | Runtime preamble (imports, helpers). | `runtime.go` |
| `repl.go` | REPL emission (command reader, dispatch). | `repl.go` |
| `tick.go` | Progress / `Tick` method. | `tick.go` |
| `vprint.go` | Variable/trace printing. | `vprint.go` |
| `native.go` | Go-tagged native blocks. | `native.go` |
| `native_thunk.go` | Thunked native action wrappers. | `native_thunk.go` |
| `thunk.go` | Hash-thunk struct generation. | `thunk.go` |
| `action_gen.go` | Solver-driven action generator class emission (test/gen). | `action_gen.go` |
| `solver_emit.go` | Constraint encoding into `goivy.Solver`. | `solver_emit.go` |
| `z3.go` | Z3 helper emission (templates → Go funcs). | `z3.go` |
| `oracle_compare.go` | (Stub initially; behavioral oracle is post-MVP.) | `oracle_compare.go` |
| `clauses_helpers.go` | Pure-logic CNF / substitution helpers (ported as-is). | `clauses_helpers.go` |
| `ptype.go` | Param-type wrappers (`ValueType`, `RefType`) → in Go just `T` vs `*T` decisions. | `ptype.go` |
| `names.go` | Identifier mangling for Go (rules differ — see §3.5.5). | `names.go` |
| `compile.go` | `CompileAndGenerate*` entries. | `compile.go` |
| `generator.go` | `Generator` struct + `Generate()` + orchestration. | `generator.go` |
| `build.go` | `BuildPlanFor()` + `BuildOutput()` — wraps `go build`. | `build.go` (+ deletes `build_findvs*.go`) |
| `config.go` | `Config` + helpers (split out from generator.go for clarity). | (inline in ivy2cpp/generator.go) |

Note: ivy2cpp's `build_findvs*.go` set (Windows MSVC registry lookup)
has **no analogue** in ivy2go — `go build` is uniform across platforms.
Saves ~480 LOC.

### 3.4 File-by-file ivy2cpp → ivy2go correspondence

§3.3 above gives the table. The mechanical rule: each ivy2cpp file
maps to exactly one ivy2go file with the same name (or close
equivalent), housing the **same set of Go functions** with the **same
names** but emitting Go syntax instead of C++ syntax. Same callgraph,
same memoization caches, same Python-line-number comments preserved.

This means a search like `grep -n "emitExpr" ivy2cpp/expr.go` and the
analogous `grep -n "emitExpr" ivy2go/expr.go` produce parallel results.
Reviewers can compare patches across both packages line-by-line.

The few **structural deviations** from ivy2cpp:

1. No `Header` + `Impl` two-stream model; instead `GoContext` holds
   one stream per output `.go` file (`typesStream`, `stateStream`,
   `actionsStream`, `runtimeStream`, etc.) plus a `mainStream`.
   `Output.Files` is then `{ "types.go": typesStream.String(), ... }`.
2. `Output.ExtraFiles` continues to carry `.dsc` descriptors.
3. `cppWriter` becomes `goWriter`; the API is unchanged, but
   `open(s)` and `close(suffix)` understand Go brace conventions
   (`{` at end of line; closing `}` without trailing semicolon).
4. `ptype.go`'s `ValueType` / `RefType` / `ReturnRefType` collapse: Go
   passes structs by value or pointer; references vanish. We retain
   the names as a thin layer that just decides `T` vs `*T`.

### 3.5 Code-emission translation rules (C++ → Go)

This is the substance — what each emission decision becomes. Each
sub-section names the ivy2cpp file/function pair we're tracking and
the Go output we substitute.

#### 3.5.1 Type mapping (types.go, go_types.go)

| Ivy sort | ivy2cpp output | ivy2go output |
|----------|----------------|---------------|
| `Bool`   | `bool`         | `bool` |
| `Enum {a,b,c}` | scoped `enum class Foo : int { a, b, c }` | `type Foo int` + `const ( FooA Foo = iota; FooB; FooC )` |
| `Range[lo..hi]` | smallest fitting `unsigned`/`unsigned int`/etc. | `int` (with `// Range[lo..hi]` comment); range constraints enforced in `ivyAssume` at action boundaries |
| `Uninterpreted` (no native) | `int` | `int` |
| `Uninterpreted` with native `__cpptype<T>` | user-supplied C++ type | user-supplied Go type from `<<< go ... >>>` block (§3.5.10) |
| `BitVec[N]`, N≤32 | `unsigned` | `uint32` |
| `BitVec[N]`, 33≤N≤64 | `unsigned long long` | `uint64` |
| `BitVec[N]`, 65≤N≤128 | `unsigned __int128` | `Uint128` struct (two `uint64` halves) from `runtime.go`; arithmetic via methods |
| `BitVec[N]`, N>128 | `ivy_uint<N>` template | `*big.Int` masked to N bits (no generics; bare `*big.Int` is acceptable since wide BVs are rare/cold) |
| `String` (`__strlit`) | `std::string` | `string` |
| `Nat` interp | `unsigned long long` | `uint64` |
| Function `D₁×…×Dₖ → R`, small | C array `R[d1][d2]…` | `[d1][d2]…R` Go array (only when each Dᵢ is fixed-bounded enum/range) |
| Function, large | `hash_space::hash_map<K, V>` | `map[K]V` (K is a comparable Go type or a struct key) |
| Function, hash-thunk | `hash_thunk<K, V>` | dedicated thunk struct (see §3.5.6) |

The decision tree for "scalar / array / hash_map / thunk" is taken
*unchanged* from `ivy2cpp/cpp_types.go:cppFunctionStorageFor`. The
mechanical port preserves it.

#### 3.5.2 Expression emission (expr.go)

`emitExpr` dispatches on the same goivy AST node types as ivy2cpp; the
substitutions are mostly trivial:

| ivy2cpp emits | ivy2go emits |
|---------------|--------------|
| `(l && r)`    | `(l && r)` (same) |
| `(l \|\| r)`  | `(l \|\| r)` (same) |
| `!(b)`        | `!(b)` (same) |
| `(l == r)`    | `(l == r)` (same; for struct types, `equalsT(l, r)` helper) |
| `(c ? t : f)` | `ite(c, t, f)` helper from `runtime.go` (Go lacks ternary) |
| `__strlit`    | `string` |
| `forall x:S . p(x)` over finite `S` | `forAllS(func(x S) bool { return p(x) })` helper |
| `exists`       | `existsS(...)` helper |
| `let x = e in body` | substitute `e` into body (same as ivy2cpp; no runtime artefact) |
| native expression `<<< ... >>>` with `go` tag | verbatim Go expression text |
| temporal / `LogicNamedBinder` | rejected — same as ivy2cpp |

The `ite` and `forAllS` / `existsS` helpers are emitted once per sort
into `runtime.go`. We do NOT use Go generics for these — they get
specialised per sort during emission to keep call sites cheap (per D6).

#### 3.5.3 Action / statement emission (action.go, assign.go)

| Ivy construct | ivy2cpp emits | ivy2go emits |
|---------------|---------------|--------------|
| Sequence `{ s1; s2 }` | `{ s1; s2; }` block | `{ s1; s2 }` block (no semicolons) |
| Assign `lhs := rhs` (scalar) | `lhs = rhs;` | `lhs = rhs` |
| Assign quantified | nested loops | nested loops (`for i := 0; i < N; i++ { … }`) |
| Assign large (hash) | two-phase via thunk | two-phase via Go thunk closure |
| `if cond { s }` | `if (cond) { … }` | `if cond { … }` |
| `if some x. p(x) { s }` | helper macro | helper function + closure capture |
| `while cond inv { s }` | `while (cond) { … }` (invariant elided in impl) | `for cond { … }` (Go's `for` is its `while`) |
| `*` (nondet choice) | `___ivy_choose(...)` | `ivyChoose(...)` (returns randomized value) |
| `assume cond` | `ivy_assume(cond, "src:line")` | `ivyAssume(cond, "src:line")` |
| `assert cond` | `ivy_assert(cond, "src:line")` | `ivyAssert(cond, "src:line")` (panics with labeled error) |
| Call `call f(x)` | `f(x);` or method invocation | `s.f(x)` (method on `*State`) |
| Local `local x:T { s }` | `{ T x; … }` | `{ var x T; _ = x; … }` (or `func() { var x T; … }()`) |
| Return | `return` | `return` |
| Native action `<<< ... >>>` with `go` tag | C++ block | Go block |

The `emitAssertLike` helper at `action.go:231` (and analogue) emits
`ivyAssert(cond, "filename:lineno")`; in ivy2go the runtime helper
takes a `panic`-on-fail policy.

#### 3.5.4 State representation (state.go, initial_state.go)

ivy2cpp builds a C++ class; ivy2go builds a Go struct + methods. Per
state symbol:

- Scalar: field with the Go type chosen in §3.5.1.
- Small function: array field, e.g., `Slot [3][N]bool`.
- Large function: `map` field, lazily allocated in `NewState`.
- Thunk function: dedicated thunk struct field (see §3.5.6).

`NewState() *State` allocates maps, initialises arrays, primes random
state from `Init()`. Initial state (per ivy2cpp/initial_state.go) is:

- For `impl`/`repl`/`class` targets: each unconstrained symbol is set
  via `ivyChoose` (deterministic on `--seed`).
- For `test`/`gen` targets: a `goivy.NewSolver(mod, opts)` solves the
  initial-state formula and populates `State` via the model.

Destructors map to Go structs whose `Equal`, `Less`, `Hash` methods
replace ivy2cpp's `operator==`, `operator<`, `__hash` (so they remain
usable as map keys when needed).

#### 3.5.5 Naming (names.go)

Ivy names contain characters Go forbids in identifiers (`:`, `.`, `[`,
`]`, `<`, `>`). ivy2cpp's `varName()` regex chain is ported as-is; the
output character set is a subset of valid Go identifiers, so no
additional escaping is needed.

Go-specific adjustments:

- Public exposure: top-level types use `goExportedName` (PascalCase);
  intra-method locals use `varName` (lower_snake or mangled). The
  generated package's exported surface is then `State`, `NewState`,
  `(*State).Init`, `(*State).ActFoo`, plus enum types.
- Reserved-word avoidance: Go has different reserved words than C++.
  Add a Go-keywords set; if `varName(name)` lands in it, suffix `_`.
- Package name: `varName(base)` then strip leading digits / replace
  with `pkg_<base>` if it would start with a digit.

#### 3.5.6 Hash-thunk emission (thunk.go)

ivy2cpp emits anonymous `struct __thunk__N : z3_thunk<K, V> { … }`
classes. ivy2go emits:

```go
type thunk0 struct {
    memo map[keyT]valT
    env  envType   // captured state fields
}

func (t *thunk0) get(k keyT) valT {
    if v, ok := t.memo[k]; ok { return v }
    v := /* lambda body referencing t.env */
    t.memo[k] = v
    return v
}

// For test/gen targets only: a Z3 binding emission produced by the
// solver-aware path in thunk.go (see §3.5.9).
func (t *thunk0) toZ3(s *goivy.Solver, /* … */) error { … }
```

Counter naming (`__thunk__N`) is preserved verbatim for cross-package
audit alignment.

#### 3.5.7 Bitvector / arithmetic (bv_expr.go)

For widths ≤ 64, lowering is straightforward (mask & shift in Go's
fixed-width integer types). For 65 ≤ width ≤ 128, ivy2cpp uses
`unsigned __int128`; ivy2go ships a small `Uint128` struct in
`runtime.go` with methods `Add`, `Sub`, `Mul`, `And`, `Or`, `Xor`,
`Shl`, `Shr`, `Eq`, `Lt`, plus `Uint128FromString`, `Uint128Random`.
For widths > 128, use `*big.Int` directly with a mask applied after
each operation. No generics anywhere; method dispatch is monomorphic
in `Uint128` and inlinable.

Operator name mapping (`bvadd`, `bvsub`, `bvmul`, `bvudiv`, `bvurem`,
`bvshl`, `bvlshr`, `bvashr`, `bvand`, `bvor`, `bvxor`, `bvnot`,
`bvneg`, `concat`) mirrors ivy2cpp/bv_expr.go entry-by-entry. Shift
saturation logic copies the same out-of-range guards.

#### 3.5.8 Runtime emission (runtime.go, repl.go, tick.go, vprint.go)

What ivy2cpp ships as static C++ headers (`ivy_repl.hpp`,
`ivy_value.hpp`, `ivy_threads.hpp`, `ivy_hash.hpp`,
`ivy_wide_uint.hpp`, `ivy_z3_helpers.hpp`) becomes:

- A Go `runtime.go` emitted inline with helpers (`ivyAssert`,
  `ivyAssume`, `ivyChoose`, `ivyPrint`, `ivyWriteTrace`, `Uint128`,
  trace I/O, `Tick`).
- Concurrency: `pthread` / Windows threads → goroutines + `sync.Mutex`
  / `sync.WaitGroup`. The mutex around `State` becomes a `sync.Mutex`
  field on `*State`.
- I/O: `std::ofstream __ivy_out` → an `io.Writer` (default
  `os.Stdout`) stored on `*State` or as a package-level (set once at
  init) writer.
- REPL: ivy2cpp's `repl.go` (934 LOC) becomes a Go REPL emitted into
  the generated package — a token reader, action-name dispatch table,
  argument parsing per parameter sort, optional history. Mirrors
  ivy2cpp's command grammar precisely so the same `.in` transcripts
  produced for ivy2cpp work against ivy2go binaries.

For `target=class`, none of the REPL / main / signal-handling is
emitted — only the `State` type and its methods, suitable for
embedding in another Go program.

#### 3.5.9 Z3 / test-generation flow (action_gen.go, solver_emit.go, z3.go, thunk.go)

This is the heaviest substitution. ivy2cpp's flow:

1. For each action, compute the *reverse image* of its update — a
   formula over `pre_state` + `inputs` + `post_state`.
2. Emit an `action_gen_foo` C++ class whose constructor pushes the
   `pre_state` into a Z3 solver, then `execute(post)` randomizes
   inputs, solves, and reads back into actual C++ variables via the
   `__from_solver<T>` specialisations.

ivy2go's flow uses `goivy.Solver` / `goivy.Translator` end-to-end:

```go
type actionGenFoo struct {
    sol *goivy.Solver
    tr  *goivy.Translator
    // captured pre-state expressions, inputs, etc.
}

func newActionGenFoo(s *State) *actionGenFoo {
    sol := goivy.NewSolver(modRef, goivy.DefaultSolverOptions())
    // encode pre-state assertions:
    sol.Assert(/* goivy.Expr built from s */)
    return &actionGenFoo{sol: sol, tr: sol.Translator()}
}

func (g *actionGenFoo) execute(s *State) (in1 T1, in2 T2, err error) {
    // randomize inputs (push value assertions into sol)
    res, err := g.sol.GetSmallModel(/* ... */)
    if err != nil { return }
    // extract values
    in1 = /* model evaluation of input1 */
    in2 = /* model evaluation of input2 */
    // apply action
    s.ActFoo(in1, in2)
    return
}
```

Crucially, every Z3 call goes through `goivy.Solver` per D1/D2 — no
direct `smt.Z3Solver` use from generated code. The complexity of
`solver_emit.go` (1,195 LOC: per-cell `add(__to_solver(…))`, large-
function forall thunking, destructor recursion) maps almost mechanically
to `Solver.Assert(...)` calls built via `goivy.Translator.FormulaToZ3`.

For thunks (large-domain assignments), `to_z3` becomes a method that
captures env symbols (as `goivy.Expr` values), builds a forall-
quantified equality, and asserts it through `Solver.Assert`.

#### 3.5.10 Native blocks (native.go)

Ivy `<<< ... >>>` blocks can be tagged. ivy2cpp emits blocks tagged
with `cpp`, `header`, `member`, `impl`, `init`, `inline`, `encode`.
ivy2go reuses the same tag vocabulary but **also accepts a new
language tag set**: `go`, `go_header`, `go_init`, `go_inline`,
`go_encode`. The dispatch in `native.go` picks Go-tagged blocks when
emitting; non-Go-tagged blocks are warnings (skipped) so a single
`.ivy` file can carry both `cpp` and `go` native blocks side by side.

This is the only place where the Ivy source itself may differ
between targets. It's an intentional, documented extension.

### 3.6 Runtime support: emitted programs import goivy directly (D2)

Generated `runtime.go` opens with:

```go
package <pkg>

import (
    "fmt"
    "math/rand"
    "os"
    "sync"

    "github.com/glycerine/ivy/goivy"           // for Z3 in test/gen
    "github.com/glycerine/ivy/goivy/smt"       // for Z3 expr types
)
```

For `target=impl|class|repl`, the goivy imports may be elided (a
build-time flag in `Config` decides). For `target=test|gen`, they
are mandatory.

Tradeoffs accepted by D2:

- Every test/gen binary depends transitively on the full goivy
  package (≈large). Not a concern: these binaries are tooling, not
  production payloads; build cost dominates link cost.
- Cyclic import risk: goivy must not, in turn, import any package
  under `ivy2go/`. The `Config.GoivyImportPath` knob lets the
  ivy2go package's own tests substitute a stub if needed.

### 3.7 Build & integration (build.go)

`BuildPlanFor(out *Output, outDir string) (*BuildPlan, error)` returns:

```go
type BuildPlan struct {
    GoBin       string   // `go` from $PATH; or $GOEXE
    Args        []string // ["build", "-o", outBinary, "."]
    Env         []string // optional GOFLAGS overrides
    OutputPath  string   // path to produced binary
    CompileOnly bool     // for target=class, skip the binary build
}
```

`BuildOutput(out, outDir)` writes `out.Files` to `outDir`, then `exec.Command(plan.GoBin, plan.Args...)` from `outDir`.

We do **not** generate a `Makefile`, do **not** detect MSVC, do **not**
shell out to a separate package manager. `go.mod` is either emitted by
us (when `Config.GoModule` is set) or assumed to be supplied by the
caller (e.g., a workspace `go.work` includes the generated dir).

Vendored `goivy` resolution is solved automatically once `go.mod`
points at the right thing (or the caller uses a `replace` directive).

### 3.8 Testing strategy (the most important section)

Per D5, **in-process unit tests on emitted text** is the primary
strategy. Concretely, mirror `ivy2cpp_test.go` (8,940 LOC, 200+ test
functions) in `ivy2go_test.go`, with the following coverage tiers:

#### Tier 1 — Generator unit tests (must-have, fast, in-process)

For each translation rule in §3.5, a focused test that builds an
in-process `*goivy.Module` and asserts the emitted Go matches a set
of substring shapes. Patterns:

```go
func TestEmit_EnumDeclProducesIotaBlock(t *testing.T) {
    mod := compileIvySource(t, "type color = {red, green, blue}")
    out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
    require.NoError(t, err)
    text := out.Files["types.go"]
    requireHasLineWithAllTerms(t, text, "type", "Color", "int")
    requireHasLineWithAllTerms(t, text, "Red", "Color", "iota")
    requireHasLineWithAllTerms(t, text, "Green", "Color")
    requireHasLineWithAllTerms(t, text, "Blue", "Color")
}
```

Use the same `compileIvySource`, `hasLineWithAllTerms`,
`normalizeGo` helpers ivy2cpp does for C++. The full inventory:

- `TestEmit_*` per §3.5.1 type case (enum, range, uninterp,
  bitvec≤32 / 64 / 128 / >128, function-storage decision).
- `TestEmit_*` per §3.5.2 expression case (And/Or/Not/Iff/Implies,
  Eq, Ite, Apply, ForAll, Exists, Let, Numeral, BV numeral,
  Variable, NativeExpr).
- `TestEmit_*` per §3.5.3 action case (Sequence, Assign scalar /
  quantified / two-phase, If, While, Choice, Call, Local, Return,
  Assume, Assert, IfSome, NativeAction).
- `TestEmit_*` per §3.5.4 state case (NewState scalar fields, array
  fields, map fields, thunk fields, destructor structs).
- `TestEmit_*` per §3.5.5 naming case (Go-reserved-word collision,
  digit-prefix package name, dotted module path).
- `TestEmit_*` per §3.5.6 thunk case (env capture, multi-arg key,
  Z3 to_z3 emission under test/gen).
- `TestEmit_*` per §3.5.7 BV case (each operator, mask correctness
  per width band).
- `TestEmit_*` per §3.5.8 runtime case (REPL command grammar
  emission, `Tick` emission, trace LHS emission with `Trace: true`).
- `TestEmit_*` per §3.5.9 test/gen case (action_gen struct shape,
  solver assert calls, model extraction).
- `TestEmit_*` per §3.5.10 native case (go-tag dispatch, mixed
  cpp+go block in same source).
- `TestEmit_*` per target (impl, class, repl, test, gen).

Target volume: parity with ivy2cpp's 200+ functions. Tracking via a
running counter in `ivy2go/AUDIT_TODO_LIVE.md` (analogue of ivy2cpp's
`AUDIT2_*.md`).

#### Tier 2 — Build smoke (opt-in, slow)

Gated by `SLOW_GO_TEST=1`:

```go
func TestSmoke_BuildEmittedPackage_basicAssign(t *testing.T) {
    if os.Getenv("SLOW_GO_TEST") == "" { t.Skip("SLOW_GO_TEST not set") }
    out := generateFixture(t, "test_vec/basic_assign.ivy", Config{Target:"repl"})
    dir := t.TempDir()
    require.NoError(t, WriteOutput(out, dir))
    require.NoError(t, runGoBuild(t, dir)) // exec.Command("go", "build", "./...")
}
```

One smoke test per fixture in `test_vec/` (initially the 14 we
inherit from ivy2cpp). Catches whole classes of regressions (`go vet`
errors, undefined references, type mismatches) without comparing
behavior.

#### Tier 3 — Static checks (always-on, very fast)

For every generated file in every in-process test, run:

- `go/format.Source(text)` — assert the emitter produced syntactically
  valid, gofmt-clean Go.
- Optionally `go/parser.ParseFile()` to catch parse errors with rich
  location info.

This catches 90% of "I forgot a brace" / "wrong keyword" bugs before
a single `go build` runs, and it's near-free per test.

#### Tier 4 — Behavioral oracle vs ivy2cpp (post-MVP, future)

Out of scope for the initial deliverable per D5, but the file
`oracle_compare.go` exists from day one as a stub so the harness can
land later without renames. When added, the flow will be:

1. For each fixture: build ivy2cpp's C++ binary AND ivy2go's Go binary.
2. Run both with the same `.test.args` / `.in` transcripts.
3. Diff stdout lines (modulo a normalisation of trace line ordering).

This is gated by `BEHAVIOR_ORACLE=1` and is *advisory* — divergences
trigger investigation but don't necessarily block merges (since
either side could be the source of truth on a given divergence).

#### Tier 5 — Hygiene (always-on)

Port `ivy2cpp/comments_test.go`'s pattern: a test scans `ivy2go/*.go`
source for `TODO`/`DEFER`/`XXX` markers and requires each one to
reference an item in `ivy2go/AUDIT_TODO_LIVE.md`. Keeps the
audit-trail discipline alive in the new package.

#### Test data

`ivy2go/test_vec/` is populated by **symlinking** to
`ivy2cpp/test_vec/oracle/` initially:

```
ivy2go/test_vec/ → ../ivy2cpp/test_vec/oracle/
```

This is read-only sharing; both packages exercise the same 14
fixtures. When fixtures specific to ivy2go arise (e.g., to exercise
the `go`-tagged native blocks), they land directly under
`ivy2go/test_vec/` with no symlink involvement.

### 3.9 Risks / open questions

| # | Risk / question | Mitigation |
|---|------------------|-----------|
| R1 | Generated test/gen binaries pull entire goivy as a dep — link bloat, build time. | Acceptable per D2. Could later split out a `goivy/smtfacade` sub-package; not now. |
| R2 | Cyclic-import risk if goivy ever imports ivy2go. | Forbid in CI by adding a guard test (`go list -deps github.com/glycerine/ivy/goivy` must not contain `ivy2go`). |
| R3 | Avoiding generics constrains Uint128 / forAll helpers; emission size grows. | Accepted per D6. Helpers are per-sort, emitted once per program. |
| R4 | No behavioral oracle at MVP means semantic regressions may slip past unit text-shape checks. | Tier 2 (build smoke) catches type errors; Tier 4 (post-MVP) catches behaviour. M9 in §4 is dedicated to designing Tier 4. |
| R5 | Wide BV (>128) via `*big.Int` is slower than ivy2cpp's `ivy_uint<N>`. | Acceptable; wide BV is uncommon. Document the perf gap. |
| R6 | REPL semantics drift between ivy2cpp and ivy2go. | M7 ports the REPL grammar test-by-test from `ivy2cpp/repl_parser_test.go`; the same `.test.args` must produce equivalent dispatch. |
| R7 | `go.mod` resolution in generated dirs (replace directives, vendored goivy). | M5's smoke tests pin a known-working `go.mod`. The default is "caller manages `go.mod`"; we document this. |
| R8 | Native `go` blocks let users write arbitrary Go; security/static-check posture? | Same as ivy2cpp — native blocks are trusted source. Document in `ivy2go/CLAUDE.md`. |

---

## 4. Sequential TODO milestones (M0 – M10)

Each milestone is independently shippable and testable. The intent is
to land them in order; each builds on its predecessor.

### M0 — Bootstrap (≈1 day)

- Create `ivy2go/ARCHITECTURE_TODO.md` (this document).
- Create `ivy2go/CLAUDE.md` with sub-package rules:
  - Mirror `ivy2cpp` structure file-by-file.
  - Preserve Python-line-number comments where ivy2cpp has them
    (rephrased to reference `ivy_to_cpp.py` *via* ivy2cpp).
  - Generated code must `gofmt` clean; tests enforce it.
  - No generics in generated code (per D6).
  - Z3 access through `goivy.Solver` only.
- Create `ivy2go/AUDIT_TODO_LIVE.md` (initially empty; grows).
- Stub `ivy2go/doc.go` with package doc.

### M1 — Skeleton: writer, context, config, generator, compile (≈2 days)

Port the framing:

- `ivy2go/writer.go` — `goWriter`.
- `ivy2go/go_context.go` — `GoContext` with multi-stream output.
- `ivy2go/config.go` — `Config`, `Output`, `BatchOutput`.
- `ivy2go/generator.go` — `Generator` struct, `Generate(mod, cfg)`,
  `generate()`, `checkMemberNames()`, plus stubs for `emitTypes`,
  `emitState`, `emitActions`, `emitInit`, `emitRuntime`, `emitRepl`,
  `emitMain`.
- `ivy2go/compile.go` — `CompileAndGenerate`, `CompileAndGenerateAll`,
  `mergeParams`, `selectedIsolates`, `prepareModuleForGo`.

Tests: parse an empty module, run `Generate`, assert `Output.Files`
contains the expected keys (`types.go`, `state.go`, …) all with valid
Go syntax (`format.Source` round-trips).

### M2 — Type emission (≈3 days)

- `ivy2go/go_types.go` — `goType()`, `goScalarType()`, function
  storage classification.
- `ivy2go/types.go` — `emitSortDecls`, `emitCTupleDecls` (struct keys
  for multi-arg map keys), enum / range / uninterp / native / variant /
  destructor / BV emission.
- `ivy2go/destructor.go` — Destructor record structs with `Equal`,
  `Hash`, `Less`.
- `ivy2go/variant.go` — Tagged-union structs.
- `ivy2go/constructors.go` — Sort constructor emission.

Tests per type case (§3.5.1). Tier 3 static check on every output.

### M3 — Expression emission (≈3 days)

- `ivy2go/expr.go` — `emitExpr` dispatch matching ivy2cpp.
- `ivy2go/bv_expr.go` — Bitvector lowering + `Uint128` helpers.
- `ivy2go/runtime.go` — Initial helper emission (`Uint128`, `ite`,
  `forAllS`/`existsS` per-sort skeletons).

Tests per expression case (§3.5.2, §3.5.7).

### M4 — Action emission (≈4 days)

- `ivy2go/action.go` — `emitAction` dispatch.
- `ivy2go/assign.go` — Scalar / quantified / two-phase / large
  assignment lowerings.
- `ivy2go/nondet.go` — `ivyChoose` emission and per-sort randomizers.
- `ivy2go/extensional.go` — Extensional relation handling.
- `ivy2go/definitions.go` — Definitional axiom emission.

Tests per action case (§3.5.3). At end of M4: a non-empty `.ivy`
fixture with sorts + actions emits to a complete-looking package.

### M5 — State + Init + impl target (≈3 days)

- `ivy2go/initial_state.go`, `ivy2go/init.go`.
- `ivy2go/generator.go` — Wire `emitState`, `emitInit`.
- `ivy2go/build.go` — `BuildPlanFor`, `BuildOutput` for `target=impl`.

Tier 2 smoke test: build the 14 fixtures' `target=impl` outputs;
require `go build` to succeed.

### M6 — Class target + `class` smoke (≈1 day)

- Verify `target=class` produces no `main()`, exports `*State` API,
  and is `go build`-able as a library.

### M7 — REPL target (≈4 days)

- `ivy2go/repl.go` — Command reader, dispatch, parameter parsing.
- `ivy2go/tick.go` — `Tick` method, progress counters.
- `ivy2go/vprint.go` — Trace LHS / RHS printing.
- `ivy2go/generator.go` — Wire `emitRepl`, `emitMain`.

Tests: port `ivy2cpp/repl_parser_test.go` patterns; verify `.in`
transcripts dispatch the same actions.

### M8 — Thunks (≈3 days)

- `ivy2go/thunk.go` — Hash-thunk struct emission (runtime path only).
- `ivy2go/native_thunk.go` — Native action thunks.

Tests: emit thunks for `target=impl`; verify struct shape, env capture,
counter `__thunk__N` parity with ivy2cpp.

### M9 — Test/Gen target (Z3 via goivy.Solver) (≈6 days)

- `ivy2go/z3.go` — Setup helpers using `goivy.Translator` /
  `goivy.NewSolver`.
- `ivy2go/solver_emit.go` — Constraint emission for state-into-solver.
- `ivy2go/action_gen.go` — Per-action generator struct emission.
- `ivy2go/thunk.go` — Add Z3-aware `toZ3` method emission for
  test/gen targets.

Tests:
- `TestEmit_ActionGen_StructShape` — assert emitted Go has
  `newActionGen<Foo>`, captures `*goivy.Solver`, calls
  `Solver.GetSmallModel`.
- `TestEmit_SolverEmit_LargeFunctionForall` — assert large functions
  produce a forall-quantified `Solver.Assert`.
- `TestEmit_Thunk_Z3Path_Captures_Env` — solver-aware thunks include
  the env-capture forall.
- Tier 2 smoke: build the 14 fixtures' `target=test` outputs;
  require `go build` to succeed.

### M10 — Native Go blocks, hygiene, docs (≈2 days)

- `ivy2go/native.go` — `go` / `go_header` / etc. tag dispatch.
- `ivy2go/comments_test.go` — TODO/DEFER discipline test.
- `ivy2go/oracle_compare.go` — Stub the future behavioral oracle
  (so M11 can land without renames).
- Update `ivy2go/CLAUDE.md` with finalised rules.
- Update `ivy2go/AUDIT_TODO_LIVE.md` listing all known parity gaps
  vs ivy2cpp.

### M11+ — Future (post-MVP)

- Behavioral oracle harness (Tier 4 in §3.8) comparing ivy2go and
  ivy2cpp binaries on shared fixtures.
- Wasm target (orthogonal; revisit if a user needs it).
- `go.mod` auto-generation with `replace` directives pointing at the
  in-tree goivy.

---

## 5. Verification

End-to-end checks for each milestone (`cd ~/ivy/goivy`):

```sh
# Fast unit tests (Tier 1 + 3 + 5)
env XTRACE_OFF=1 go test ./ivy2go -count=1

# Build smoke (Tier 2)
env XTRACE_OFF=1 SLOW_GO_TEST=1 go test ./ivy2go -run TestSmoke -count=1 -v

# Hygiene (Tier 5)
env XTRACE_OFF=1 go test ./ivy2go -run TestComments -count=1
```

Per the project-wide rule in `goivy/CLAUDE.md` section 9:
**never** run `go test ./...` (XTRACE blows up the runtime). Always
scope to `./ivy2go` or use `make test`.

Acceptance gate for the MVP (end of M10):

1. All Tier 1 + 3 + 5 tests pass under `go test ./ivy2go -count=1`.
2. Under `SLOW_GO_TEST=1`, every fixture in `test_vec/` produces an
   emitted package that `go build`s cleanly for `target=impl`,
   `target=class`, `target=repl`, `target=test`, and `target=gen`.
3. `format.Source` round-trips every emitted file.
4. `AUDIT_TODO_LIVE.md` enumerates every known gap vs ivy2cpp, each
   with a Python-line citation and a Go file:line "land here" marker.

Post-MVP gate (M11+): behavioral oracle agrees with ivy2cpp on at
least 12 of the 14 baseline fixtures; the other 2 have documented
divergence reasons.
