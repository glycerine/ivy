# Go-junior Spreadsheet Language Plan

Plan only. No implementation changes in this document.

## Goal

Build a TypeScript implementation of Go-junior, a deliberately small,
spreadsheet-oriented language with Go-like syntax and deterministic
single-threaded execution. Go-junior should compile to browser-runnable
JavaScript using source-level copy-and-patch templates, and it should be able
to call allowlisted Go packages compiled to WebAssembly through typed host
wrappers.

The first production target is the Ivy webui analysis spreadsheet. The current
spreadsheet-like pane can then evolve from an editable mock grid into a real
spreadsheet/calculation surface with formula cells, dependencies, diagnostics,
and host/WASM calls.

## Design Principles

- Treat Go-junior as its own DSL, not as "almost all of Go".
- Keep the language small enough that users can reason about every construct.
- Preserve Go's readable expression and statement syntax where it helps.
- Make spreadsheet semantics first-class: cells, ranges, dependencies,
  recalculation, error values, deterministic execution, and bounded work.
- Never emit raw user source into JavaScript. Parse, validate, typecheck, then
  emit compiler-owned JavaScript stencils with sanitized holes.
- Keep the compiler backend replaceable. Start with JavaScript source
  copy-and-patch. Leave a path to a future WebAssembly stencil backend.
- All external capability goes through typed host bindings. Go-junior code
  cannot import arbitrary JavaScript, global browser APIs, or arbitrary Go
  packages.

## Non-goals

- Full Go compatibility.
- Goroutines, `go`, `select`, channels, channel operations, `defer`, `panic`,
  `recover`, reflection, unsafe, cgo, package-level initialization order, build
  tags, or module/package loading.
- Native machine-code JIT in the browser.
- Arbitrary dynamic evaluation of user JavaScript.
- Mutating browser state from Go-junior programs.
- Long-running unbounded calculations on the main UI thread.

## High-level Architecture

```
Go-junior source
  -> scanner/parser
  -> AST with source spans
  -> resolver and typechecker
  -> spreadsheet dependency extractor
  -> typed IR
  -> JavaScript source copy-and-patch emitter
  -> function compiler/cache
  -> worker-backed runtime
  -> typed host/WASM bindings
```

The compiler package should be independent from the Ivy runtime. The Ivy webui
integration should sit on top of the compiler and runtime packages.

Suggested frontend layout:

```
goivy/webui/frontend/src/gojunior/
  ast.ts
  diagnostics.ts
  scanner.ts
  parser.ts
  resolver.ts
  types.ts
  ir.ts
  dependency.ts
  stencils.ts
  emitJs.ts
  compiler.ts
  runtime.ts
  host.ts
  workerProtocol.ts
  worker.ts
  __tests__ or adjacent *.test.ts

goivy/webui/frontend/src/services/
  goJuniorSpreadsheetService.ts
```

## Language Shape

### Program Forms

Support two source forms:

1. Cell expression form:

```go
A1 + B1 * 2
```

2. Function body form:

```go
x := cell("A1") + cell("B1")
if x > 10 {
    return math.Sqrt(x)
}
return x * 2
```

The compiler should normalize both into an implicit function:

```go
func cell(ctx Context) Value {
    ...
}
```

Do not require users to write package declarations or imports in the first
version.

### Initial Types

Start with a deliberately small static type system:

- `bool`
- `string`
- `int`
- `float64`
- `number` as an internal convenience type if needed
- `Value`, the spreadsheet value union
- `Range`, an opaque spreadsheet range type
- `Error`, an internal spreadsheet error type

Open question for implementation: whether user-visible numeric types should
distinguish `int` and `float64` in v1. A pragmatic first release can parse both
but lower both to JavaScript number unless a host binding requires an integer.

### Spreadsheet Values

Runtime cells should represent:

- blank
- bool
- number
- string
- error

Errors should propagate predictably. A cell formula that fails typechecking
should not compile. A cell formula that fails at runtime should produce a
spreadsheet error value with a stable code and message.

Initial error codes:

- `#ERROR!` for general runtime failures
- `#TYPE!` for runtime type conversion failures
- `#DIV/0!` for division by zero
- `#NAME?` for unknown symbols in formulas loaded from persisted state
- `#CYCLE!` for dependency cycles
- `#TIMEOUT!` for exhausted execution budget

### Expressions

Support:

- literals: numbers, strings, booleans
- identifiers
- unary operators: `+`, `-`, `!`
- binary arithmetic: `+`, `-`, `*`, `/`, `%`
- comparisons: `==`, `!=`, `<`, `<=`, `>`, `>=`
- boolean operators: `&&`, `||`
- parentheses
- function calls
- selector calls for host namespaces, for example `math.Sqrt(x)`
- cell references in explicit call form: `cell("A1")`
- range references in explicit call form: `range("A1:B10")`

Optional after v1:

- direct spreadsheet references like `A1` and `A1:B10`
- array/slice literals
- indexing
- struct literals

### Statements

Support:

- short variable declarations: `x := expr`
- assignments to locals: `x = expr`
- `if`, `else if`, `else`
- `return expr`
- expression statements only for calls whose return value can be ignored
- bounded `for` loops in a later stage

Avoid mutation outside local variables in v1. Cell writes should not be possible
from formulas.

### Functions

Stage v1 can compile one implicit cell function. Stage v2 can add local helper
functions:

```go
func clamp(x float64, lo float64, hi float64) float64 {
    if x < lo {
        return lo
    }
    if x > hi {
        return hi
    }
    return x
}
return clamp(cell("A1"), 0, 100)
```

No closures in the first implementation. Local helper functions should be
pure, deterministic, and compiled into private generated JavaScript functions.

### Bounded Loops

When loops are added, support only bounded forms:

```go
for i := 0; i < 10; i = i + 1 {
    sum = sum + i
}
```

Runtime should enforce a fuel counter regardless of static bounds. Range loops
may be added later if the range source is a spreadsheet range.

## Host and WebAssembly Calls

External calls must resolve through a typed host binding table:

```ts
const hostSpec = {
  math: {
    Sqrt: fn(["float64"], "float64"),
    Pow: fn(["float64", "float64"], "float64"),
  },
  ivy: {
    Parse: fn(["string"], "Value"),
  },
};
```

The compiler should reject calls that are not present in the host spec. The JS
emitter should emit calls through generated host slots, not through arbitrary
global names:

```js
const _h0 = host.math.Sqrt;
return _h0(_x);
```

Go packages compiled to `GOOS=js GOARCH=wasm` should be wrapped as services
behind this table. The Go WASM side should expose stable functions to
JavaScript through `wasm_exec.js` or a narrower wrapper. Go-junior should not
know whether a host function is implemented in JavaScript, Go WASM, TinyGo
WASM, or a web worker RPC.

For expensive host calls, the host binding can be asynchronous in a later
stage. The initial version should prefer synchronous calculation functions so
that spreadsheet recalculation remains simple.

## Source-level Copy-and-Patch JavaScript Emission

Use source stencils instead of native binary stencils. A stencil is a compiler
owned JavaScript text fragment with typed holes. For example:

```ts
const jsStencils = {
  addNumber: ["(", hole("left"), " + ", hole("right"), ")"],
  ifStmt: ["if (", hole("test"), ") {\n", hole("then"), "}\n"],
  returnStmt: ["return ", hole("value"), ";\n"],
};
```

Emission rules:

- User source text is never copied directly.
- Identifier names are generated, for example `_v0`, `_v1`, `_fn0`.
- String literals are emitted with `JSON.stringify`.
- Numeric literals are validated before emission.
- Host calls are emitted only from resolved binding IDs.
- Cell and range accesses are emitted only through runtime calls.
- Each compiled formula receives only `(ctx, host, budget)` as authority.

Example generated JavaScript:

```js
"use strict";
return function _gj_cell(ctx, host, budget) {
  let _v0 = ctx.cell("A1");
  let _v1 = ctx.cell("B1");
  let _v2 = _v0 + _v1;
  if (_v2 > 10) {
    return host.math.Sqrt(_v2);
  }
  return _v2 * 2;
}
```

Compilation can use `new Function(source)()` in the worker. If Content Security
Policy later forbids dynamic code generation, the same IR should be runnable by
a simple interpreter or a future WebAssembly backend.

## Runtime and Sandboxing

Run compiled formulas in a dedicated Web Worker whenever feasible.

Runtime responsibilities:

- compile formulas
- cache compiled functions by source hash plus host spec version
- evaluate cells in dependency order
- detect dependency cycles
- enforce execution budget/fuel
- convert thrown exceptions into spreadsheet error values
- serialize diagnostics back to the UI
- isolate host/WASM service calls behind narrow wrappers

The main UI thread should not evaluate user programs directly except in unit
tests or explicitly small trusted development paths.

## Spreadsheet Integration Model

Every formula cell should have:

- raw source text
- parsed AST or parser error
- typecheck diagnostics
- dependency list
- compiled function cache key
- last computed value
- last runtime error, if any

The spreadsheet engine should maintain a dependency graph:

```
cell -> direct dependencies
dependency -> reverse dependents
```

When a cell changes:

1. Parse and typecheck the changed formula.
2. Update dependency edges.
3. Mark dependents dirty.
4. Recalculate in topological order.
5. Publish value and diagnostics changes to the UI model.

Direct references like `A1` can be syntax sugar for `cell("A1")` after the
explicit API works.

## Stage Plan and Test Plan

### Stage 0: Specification and Fixtures

Write a small language spec in the new Go-junior package or docs area. Define
syntax, excluded Go features, types, errors, host binding shape, and runtime
limits.

Implementation tasks:

- Create fixture format for compiler tests.
- Define diagnostic object shape with code, message, severity, source span,
  and optional help text.
- Define source span conventions: line, column, offset, and length.

Tests:

- Golden fixture loader reads valid and invalid cases.
- Diagnostic spans round-trip line/column/offset accurately.
- Spec examples are copied into parser/typechecker tests so docs cannot drift.
- A "blocked Go feature" fixture list asserts that `go`, `select`, channels,
  `defer`, `panic`, `recover`, `unsafe`, `package`, and `import` are rejected
  with explicit diagnostic codes.

Acceptance criteria:

- The subset is explicit enough that parser and typechecker work can begin.
- Tests can assert both successful normalized output and exact diagnostics.

### Stage 1: Scanner

Implement a TypeScript scanner for Go-junior tokens.

Implementation tasks:

- Recognize identifiers, keywords, numbers, strings, operators, delimiters,
  comments, and whitespace.
- Track source spans.
- Support Go-like semicolon insertion only if needed. Prefer requiring braces
  and explicit statement boundaries in parser logic rather than reproducing all
  Go semicolon rules.
- Reject unsupported tokens early.

Tests:

- Tokenizes literals: decimal integers, floats, quoted strings, escaped
  strings, booleans.
- Tokenizes operators and delimiters.
- Preserves source spans across newlines and comments.
- Rejects unterminated strings with a useful span.
- Rejects unsupported rune literals if they are not in v1.
- Rejects channel operator `<-`.
- Rejects malformed numbers.
- Snapshot tests for representative snippets.

Acceptance criteria:

- Scanner never throws for user input; it returns tokens plus diagnostics.
- Every token has a stable span.

### Stage 2: Parser

Implement a recursive-descent parser with Pratt expression parsing.

Implementation tasks:

- Parse expression-form and function-body-form programs.
- Parse statements: declarations, assignments, if/else, returns, expression
  statements.
- Parse calls and selector expressions.
- Parse optional local helper functions after v1 expression/body support.
- Preserve AST spans for all nodes.
- Add error recovery sufficient to show multiple diagnostics.

Tests:

- Parses arithmetic precedence correctly.
- Parses nested parentheses.
- Parses unary and binary operations.
- Parses `if`, `else if`, and `else`.
- Parses short declarations and assignments.
- Parses return statements.
- Parses selector calls like `math.Sqrt(x)`.
- Parses explicit cell/range calls.
- Rejects `go f()`, `select {}`, channel sends, receives, imports, package
  declarations, labels, `goto`, `defer`, `panic`, `recover`, and `unsafe`.
- Error recovery reports multiple syntax errors in one source where possible.
- Golden AST tests for small valid programs.

Acceptance criteria:

- Valid v1 syntax produces a complete AST.
- Invalid or unsupported syntax produces stable diagnostics without crashing.

### Stage 3: AST Normalization

Normalize program forms into one internal function-like representation.

Implementation tasks:

- Expression form becomes `return expr`.
- Function body form becomes an implicit cell function.
- Insert explicit return requirement checks where needed.
- Normalize syntactic sugar into canonical AST nodes.

Tests:

- Expression source normalizes to one return statement.
- Function body source preserves user statements.
- Empty source produces a blank or diagnostic according to spec.
- Missing return in a non-void body is reported.
- Unreachable trailing code after unconditional return is diagnosed as warning
  or ignored according to spec.

Acceptance criteria:

- Later compiler stages consume one normalized AST shape.

### Stage 4: Resolver and Scope Analysis

Resolve local variables, function names, host symbols, and predeclared
spreadsheet functions.

Implementation tasks:

- Implement lexical scopes.
- Resolve local declarations and uses.
- Reject use before declaration unless explicitly allowed.
- Reject duplicate local declarations in the same scope.
- Resolve host namespaces and functions against a typed host spec.
- Resolve predeclared functions: `cell`, `range`, and conversion helpers.
- Generate stable symbol IDs for locals and functions.

Tests:

- Resolves locals in nested `if` blocks.
- Rejects unknown identifiers.
- Rejects duplicate declarations.
- Rejects assignment to undeclared variables.
- Rejects assignment to host symbols and predeclared functions.
- Resolves host calls with namespace and function names.
- Rejects unknown host namespaces/functions.
- Does not resolve against `window`, `document`, `globalThis`, or other browser
  globals.

Acceptance criteria:

- Every identifier use is bound to a known symbol or produces a diagnostic.

### Stage 5: Typechecker

Add static typing for expressions, statements, returns, and calls.

Implementation tasks:

- Define type model and assignability rules.
- Type arithmetic, comparisons, booleans, strings, and returns.
- Type predeclared cell/range helpers.
- Type host calls from the host spec.
- Decide numeric coercion policy and implement it consistently.
- Type local helper functions when added.

Tests:

- Arithmetic accepts compatible numeric operands.
- String concatenation policy is enforced, either allowed only for strings or
  rejected in v1.
- Boolean operators reject non-bool operands.
- Comparisons reject incompatible types.
- `if` conditions must be bool.
- Return expression must match the implicit cell return type or be convertible
  to `Value`.
- Host calls validate arity and argument types.
- Host return types flow into later expressions.
- Cell values require explicit conversion when needed, if the type policy uses
  `Value`.
- Diagnostics point at the offending expression.

Acceptance criteria:

- Typechecking rejects invalid programs before emission.
- Typechecked AST nodes carry resolved types.

### Stage 6: Dependency Extraction

Extract spreadsheet dependencies from typed AST.

Implementation tasks:

- Recognize `cell("A1")` and `range("A1:B10")`.
- Validate cell and range addresses.
- Record direct dependencies in a normalized address format.
- Mark dynamic references as requiring conservative dependency handling if
  later supported.
- Keep dependency extraction independent from JS emission.

Tests:

- Extracts one cell dependency.
- Extracts multiple dependencies from expressions and branches.
- Extracts range dependencies.
- Rejects invalid cell/range strings.
- Handles duplicate dependencies by deduplicating.
- Does not count strings passed to unrelated functions as dependencies.
- Produces stable output order.

Acceptance criteria:

- The spreadsheet engine can update dependency graph without executing code.

### Stage 7: Typed IR

Lower typed AST to a compact IR designed for both JS template emission and a
future WebAssembly backend.

Implementation tasks:

- Define IR nodes for constants, locals, arithmetic, comparisons, boolean ops,
  branches, returns, calls, cell/range reads, and later loops.
- Preserve diagnostic/source mapping metadata.
- Add optional constant folding for simple literals.
- Add explicit conversions where needed.
- Represent host calls by binding ID, not by text name.

Tests:

- Golden IR tests for expressions, branches, locals, and host calls.
- Constant folding tests if implemented.
- IR contains no raw user identifier names except for metadata.
- IR host calls contain host binding IDs.
- IR cell reads contain normalized addresses.

Acceptance criteria:

- JS emission can be implemented without consulting raw source text.

### Stage 8: JavaScript Source Copy-and-Patch Emitter

Emit JavaScript from IR using stencils.

Implementation tasks:

- Define stencil helpers for expressions, statements, functions, branches, and
  calls.
- Generate hygienic local names.
- Emit strict mode function source.
- Emit optional budget checks at function entry and loop backedges.
- Emit source map or diagnostic mapping metadata if practical.
- Expose compiler API returning generated source for debugging and tests.

Tests:

- Generated JS for simple arithmetic matches golden snapshots.
- Generated JS never contains raw user variable names.
- String literals are escaped with `JSON.stringify`.
- Host calls route through `host` using resolved binding paths or IDs.
- Cell reads route through `ctx.cell`.
- Range reads route through `ctx.range`.
- Unsupported IR nodes fail with internal compiler diagnostics.
- Generated source parses via `new Function` in tests.
- Malicious source snippets cannot break out through emitted JS:
  - `"); globalThis.evil = true; ("`
  - identifiers named like JS keywords
  - strings containing backticks and `${...}`
  - source containing `</script>`

Acceptance criteria:

- Safe, valid JavaScript is produced for all valid v1 IR programs.

### Stage 9: Function Compilation and Cache

Compile generated JavaScript to callable functions and cache them.

Implementation tasks:

- Create cache key from source hash, compiler version, host spec version, and
  selected runtime options.
- Compile with `new Function`.
- Return structured compile errors if browser/CSP rejects dynamic compilation.
- Add a development/debug option to expose generated source.

Tests:

- Same source and host spec hits cache.
- Host spec version changes invalidate cache.
- Compiler version changes invalidate cache.
- Compile failure returns diagnostic/error object.
- Cached function returns same result as newly compiled function.

Acceptance criteria:

- Repeated recalculation does not repeatedly parse and compile unchanged cells.

### Stage 10: Runtime Evaluation

Implement the evaluator context and spreadsheet value conversion.

Implementation tasks:

- Define `ctx.cell(address)`, `ctx.range(address)`, and value conversion helpers.
- Convert returned JS values to spreadsheet values.
- Catch runtime exceptions.
- Enforce fuel/budget for loops and optional call count.
- Define division by zero behavior.
- Add deterministic math behavior where possible.

Tests:

- Evaluates arithmetic formulas.
- Evaluates conditionals.
- Reads cells from a fake context.
- Reads ranges from a fake context.
- Calls fake host functions.
- Converts return values to spreadsheet values.
- Propagates runtime errors as spreadsheet error values.
- Enforces budget on artificial loops once loops exist.
- Division by zero gives expected error or value according to spec.

Acceptance criteria:

- A compiled formula can run against a fake spreadsheet context in unit tests.

### Stage 11: Worker Runtime

Move compile/evaluate operations into a Web Worker.

Implementation tasks:

- Define worker protocol: compile, evaluateCell, evaluateBatch, disposeCache,
  updateHostSpec, and diagnostics messages.
- Keep host/WASM services behind explicit RPC endpoints.
- Support cancellation by generation token.
- Ensure stale results are ignored by the main thread.

Tests:

- Worker compiles valid formulas.
- Worker reports diagnostics for invalid formulas.
- Worker evaluates a batch in dependency order provided by the main thread.
- Cancellation token suppresses stale results.
- Worker survives a formula runtime exception.
- Worker cache persists across messages.
- Worker can be terminated and restarted without losing main-thread state.

Acceptance criteria:

- User formulas do not execute on the main UI thread in production mode.

### Stage 12: Spreadsheet Dependency Graph and Recalculation

Implement spreadsheet-level formula management.

Implementation tasks:

- Track formula cells, literal cells, dirty cells, dependencies, and dependents.
- Topologically sort dirty subgraphs.
- Detect cycles before evaluation.
- Incrementally update dependencies when formulas change.
- Publish recalculation results to UI state.

Tests:

- Single formula cell recalculates from literals.
- Changing one literal marks dependents dirty.
- Recalculation order respects dependencies.
- Diamond dependency graph recalculates each formula once.
- Dependency cycle reports `#CYCLE!` for affected cells.
- Removing a formula removes old dependency edges.
- Invalid formula source preserves previous value or produces error according
  to spec.
- Batch updates coalesce recalculation.

Acceptance criteria:

- The spreadsheet engine has real formula semantics independent of the DOM.

### Stage 13: Ivy Webui Spreadsheet Integration

Integrate the Go-junior engine with the existing webui spreadsheet pane or its
replacement grid.

Implementation tasks:

- Add service layer for formula source, calculated value, selected cell, and
  diagnostics.
- Decide which cells mirror Ivy source lines and which cells are free formula
  cells.
- Wire formula bar editing to Go-junior source.
- Show calculated values in grid cells while preserving raw source in formula
  bar.
- Surface parse/type/runtime diagnostics near the selected cell.
- Keep editor-line synchronization for the existing "spec line" column if that
  workflow remains.

Tests:

- Editing a formula cell stores raw source and displays calculated value.
- Selecting a cell shows raw formula source in the formula bar.
- Editing formula bar updates selected cell.
- Formula diagnostics appear for invalid formulas.
- Editing Ivy source lines still updates the editor buffer where supported.
- Spreadsheet recalculation does not mutate Ivy source except through explicit
  spec-line cells.
- UI tests verify no overlapping text or broken layout in the spreadsheet pane.

Acceptance criteria:

- The visible spreadsheet pane performs real calculations with Go-junior
  formulas.

### Stage 14: Go WASM Host Binding Integration

Expose selected Go/WASM package functions through typed host wrappers.

Implementation tasks:

- Define host binding declaration format in TypeScript.
- Wrap existing Go WASM exports or `wasm_exec.js` globals behind stable async or
  sync functions.
- Version the host spec.
- Add marshalling for primitive values and structured spreadsheet values.
- Add error conversion from Go/WASM failures to spreadsheet errors.

Tests:

- Fake host binding tests for arity/type enforcement.
- Real smoke test loads a small Go WASM module if available in the existing
  webui runtime.
- Host function can be called from Go-junior formula.
- Host function errors become spreadsheet errors.
- Unknown host function is rejected at typecheck time.
- Host spec version invalidates compiled function cache.

Acceptance criteria:

- Go-junior formulas can call allowlisted Go/WASM functions safely.

### Stage 15: Persistence

Persist formulas, calculated values, diagnostics, and engine metadata.

Implementation tasks:

- Define persisted formula cell shape.
- Persist raw Go-junior source, not generated JavaScript.
- Persist compiler version and host spec version for cache invalidation.
- On load, reparse/retype/recompile formulas rather than trusting old compiled
  source.

Tests:

- Save/load preserves raw formula text.
- Save/load recalculates values from current literals.
- Loading old compiler version invalidates cache.
- Loading unknown host function reports `#NAME?` or diagnostic.
- Corrupt persisted formulas fail safely.

Acceptance criteria:

- Spreadsheet state can round-trip without persisting executable JavaScript.

### Stage 16: Diagnostics and Developer Tools

Add user and developer visibility into compilation.

Implementation tasks:

- Display selected cell diagnostics.
- Add optional developer panel or console hook for generated JS.
- Add trace mode for dependency graph and recalculation order.
- Add stable diagnostic codes.

Tests:

- Diagnostic codes are stable in golden tests.
- Diagnostic spans map to formula bar offsets.
- Generated JS debug output is disabled by default.
- Trace mode records dependency and recalculation events.

Acceptance criteria:

- Users see actionable formula errors, and developers can debug compiler output
  without exposing unsafe execution paths.

### Stage 17: Security Hardening

Audit all authority boundaries.

Implementation tasks:

- Confirm no raw user source enters generated JS.
- Confirm worker has no unnecessary host capabilities.
- Confirm host bindings are allowlisted and typed.
- Add CSP compatibility fallback decision: interpreter or explicit unsupported
  message.
- Add runtime budget defaults.
- Add maximum source length, AST size, dependency count, and recursion limits.

Tests:

- Fuzz scanner/parser with random input.
- Property tests ensure generated JS contains only generated identifiers for
  user variables.
- Malicious strings and identifiers cannot escape emission.
- Formulas cannot access browser globals.
- Long formulas hit size limits.
- Deeply nested expressions hit depth limits gracefully.
- Infinite or very long loops hit budget once loops exist.
- Worker restart recovers from fatal internal errors.

Acceptance criteria:

- The engine is safe enough for untrusted spreadsheet formulas under the
  intended browser threat model.

### Stage 18: Performance and Benchmarks

Measure compile and execution behavior before optimizing.

Implementation tasks:

- Add microbenchmarks for scanner, parser, typechecker, emitter, `new Function`
  compile, and evaluation.
- Add spreadsheet recalculation benchmarks for common graph shapes.
- Measure main-thread latency with worker enabled.
- Add cache hit/miss counters.

Tests and benchmarks:

- 1,000 simple formulas compile within target budget.
- 1,000 cached formulas recalculate without recompilation.
- Dependency graph update scales with changed subgraph size.
- Worker batch evaluation avoids per-cell message overhead.
- Generated JS outperforms a simple interpreter on arithmetic-heavy formulas.

Acceptance criteria:

- The source-level copy-and-patch backend is measurably fast enough for
  interactive spreadsheet editing.

### Stage 19: Optional Interpreter Backend

Add an IR interpreter if CSP or debugging requires it.

Implementation tasks:

- Interpret the same typed IR used by the JS emitter.
- Keep interpreter semantics identical to generated JS.
- Use interpreter for tests, CSP fallback, or debugging.

Tests:

- Differential tests compare interpreter and generated JS for valid formulas.
- Random expression tests produce identical results.
- Runtime errors match between backends.
- Host calls are routed through the same typed host layer.

Acceptance criteria:

- There is a non-dynamic-code fallback path if needed.

### Stage 20: Optional WebAssembly Stencil Backend

If JS source copy-and-patch is not enough, add a WebAssembly bytecode stencil
backend using the same IR.

Implementation tasks:

- Define Wasm-friendly subset of IR.
- Build Wasm stencils for arithmetic, locals, branches, calls, and returns.
- Patch locals, constants, function indexes, and branch depths.
- Compile with `WebAssembly.compile`.
- Keep JS backend as baseline/fallback.

Tests:

- Wasm backend matches JS backend on arithmetic and branch formulas.
- Host call ABI tests cover supported primitive signatures.
- Invalid or unsupported IR falls back to JS or reports a diagnostic.
- Benchmark compile/evaluate cost against JS backend.

Acceptance criteria:

- Wasm backend is optional and only adopted if it beats JS source templates for
  the target workloads.

## Cross-cutting Test Strategy

Use layered tests so failures identify the responsible compiler phase:

- Scanner unit tests.
- Parser AST golden tests.
- Resolver/typechecker diagnostic tests.
- IR golden tests.
- JS emitter golden and security tests.
- Runtime evaluation tests with fake contexts.
- Worker protocol tests.
- Spreadsheet dependency/recalculation tests.
- Webui service tests.
- Browser/Playwright tests only for real UI behavior.
- Differential tests between interpreter and JS backend if interpreter exists.
- Fuzz/property tests for parser and emitter hardening.

Every diagnostic-producing stage should test:

- diagnostic code
- message substring
- source span
- recovery behavior

Every successful compiler stage should test:

- normalized output
- source map/span preservation where applicable
- absence of raw unsafe source in emitted artifacts

## Initial Milestone Slice

The first useful implementation should be intentionally small:

1. Scanner and parser for expression form plus short declarations, `if`, and
   `return`.
2. Types for bool, string, and number.
3. Resolver/typechecker for locals, `cell`, and one fake host namespace.
4. Dependency extraction for `cell("A1")`.
5. JS source copy-and-patch emitter.
6. Runtime tests that evaluate formulas against a fake spreadsheet context.
7. A small service-level integration test that edits a formula cell and sees a
   calculated value.

Example first formulas:

```go
cell("A1") + cell("B1")
```

```go
x := cell("A1") + cell("B1")
if x > 10 {
    return math.Sqrt(x)
}
return x * 2
```

```go
if cell("A1") == "" {
    return "missing"
}
return cell("A1")
```

## Acceptance Criteria for the Full Project

- Go-junior has a documented, tested subset of Go-like syntax.
- Unsupported Go features are rejected explicitly.
- Valid formulas typecheck before execution.
- Dependencies are extracted without running formulas.
- Formulas compile to JavaScript generated only from compiler-owned stencils.
- Generated code runs in a worker-backed runtime.
- Formula cells recalculate incrementally and detect cycles.
- Go/WASM host calls work through typed allowlisted wrappers.
- Spreadsheet state persists raw Go-junior source, not executable JS.
- Security tests cover injection, globals access, runaway execution, and host
  authority boundaries.
- Performance tests show interactive compile/recalculate latency.

