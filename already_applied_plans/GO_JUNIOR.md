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
- Include core Go function-structuring idioms that matter for real libraries,
  including `defer`, multiple return values, and named return values.
- Make spreadsheet semantics first-class: cells, ranges, dependencies,
  recalculation, error values, deterministic execution, and bounded work.
- Never emit raw user source into JavaScript. Parse, validate, typecheck, then
  emit compiler-owned JavaScript stencils with sanitized holes.
- Keep the compiler backend replaceable. Start with JavaScript source
  copy-and-patch. Leave a path to a future WebAssembly stencil backend.
- All external capability goes through typed host bindings. Go-junior code
  cannot import arbitrary JavaScript or global browser APIs. Go-junior package
  imports resolve through an explicit package resolver, package manifest, and
  compiled-code cache.

## Non-goals

- Full Go compatibility.
- Goroutines, `go`, `select`, channels, channel operations, `panic`, `recover`,
  reflection, unsafe, cgo, full Go package initialization semantics, build
  tags, or full Go module semantics.
- Arbitrary ambient package loading. Source packages are allowed, but they must
  be Go-junior-compatible packages resolved through the in-browser package
  registry/cache or a trusted host package provider.
- Native machine-code JIT in the browser.
- Ambient access to arbitrary JavaScript or browser globals. Dynamic
  JavaScript-backed helpers are allowed only through explicit typed
  capabilities.
- Unmediated browser mutation. Go-junior may produce graphs, sheet updates, or
  UI effects through declared host capabilities, but it should not reach
  `window`, `document`, `globalThis`, or DOM APIs directly.
- Long-running unbounded calculations on the main UI thread.

## High-level Architecture

```
Go-junior source
  -> Chevrotain lexer/token stream
  -> Chevrotain CST parser with source spans
  -> Go-junior AST builder
  -> resolver and typechecker
  -> package resolver and package compiler
  -> spreadsheet dependency extractor
  -> typed IR
  -> JavaScript source copy-and-patch emitter
  -> function/package compiler cache
  -> worker-backed runtime
  -> typed package and host/WASM bindings
```

The compiler package should be independent from the Ivy runtime. The Ivy webui
integration should sit on top of the compiler and runtime packages.

## Lexer and Parser Technology

Use Chevrotain for the Go-junior lexer and parser instead of a hand-rolled
scanner/parser. The goal is to rely on a mature TypeScript parsing toolkit for
tokenization, source spans, grammar validation, error recovery, and CST
generation while keeping all Go-junior semantics in our own compiler stages.

Chevrotain should own:

- token definitions and lexing
- keyword/operator/delimiter recognition
- source location tracking
- parser grammar rules
- concrete syntax tree production
- syntax-level recovery and parser diagnostics
- optional grammar diagrams for developer debugging

Go-junior should still own:

- CST-to-AST conversion
- spreadsheet namespace resolution
- typechecking
- package resolution and package-state effects
- dependency extraction
- IR lowering
- JavaScript source copy-and-patch emission
- runtime evaluation and recalculation semantics

The Chevrotain grammar should describe the Go-junior subset directly. It should
not try to become a full Go parser by gradual accident. If the project later
needs to ingest arbitrary full-Go packages, that should be a separate import
path using the official Go parser or a Go parser compiled for the target
environment, not a promise that the Chevrotain grammar can parse all Go.

Implementation sequencing should be CLI-first:

1. Build the shared language/compiler/recalculation core.
2. Build the Node runtime and `gojunior` CLI.
3. Use CLI tests and manual bash sessions to refine syntax, package calls,
   dynamic dependencies, circular-reference behavior, and graph-value semantics.
4. Add the browser worker adapter after the language semantics are stable.
5. Integrate the browser worker into the Ivy webui spreadsheet pane.

Suggested frontend layout:

```
goivy/webui/frontend/src/gojunior/
  ast.ts
  diagnostics.ts
  tokens.ts
  lexer.ts
  parser.ts
  cstToAst.ts
  grammarDiagrams.ts
  resolver.ts
  types.ts
  ir.ts
  dependency.ts
  packageManifest.ts
  packageResolver.ts
  packageCompiler.ts
  packageCache.ts
  stencils.ts
  emitJs.ts
  compiler.ts
  runtime.ts
  host.ts
  runtimeAdapter.ts
  browserRuntime.ts
  nodeRuntime.ts
  nodeCli.ts
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
number(sheet.A1) + number(sheet.B1) * 2
```

2. Function body form:

```go
x := number(sheet.A1) + number(sheet.B1)
if x > 10 {
    return math.Sqrt(x)
}
return x * 2
```

The compiler should normalize both into an implicit function:

```go
func formula(ctx Context) Value {
    ...
}
```

Cell snippets should not require users to write package declarations. They may
use prebound packages from the sheet/package environment. Source package units
should support explicit package declarations and imports using the
Go-junior-compatible package resolver.

### Initial Types

Start with a deliberately small static type system:

- `bool`
- `string`
- `int`
- `float64`
- `number` as an internal convenience type if needed
- `Value`, the spreadsheet value union
- `Range`, an opaque spreadsheet range type
- tuple types for multiple return values
- `error`, for Go-style library APIs that return `(T, error)`
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
- `#REF!` for invalid cell or range references, including invalid dynamic
  references produced at runtime
- `#CYCLE!` for dependency cycles when iterative calculation is disabled
- `#DIVERGE!` for iterative components that fail to converge
- `#OSCILLATE!` for iterative components that repeat prior states or switch
  between values instead of converging
- `#UNSTABLE_DEPS!` for dynamic dependency sets that fail to stabilize within
  the recalculation budget
- `#TIMEOUT!` for exhausted execution budget

### Runtime Targets

Go-junior should run in both browser and Node.js environments.

Browser target:

- production webui execution
- Web Worker isolation
- IndexedDB package artifact cache
- browser-safe host/WASM bindings

Node target:

- fast command-line unit and integration tests
- manual interactive usability tests from bash
- worker_threads isolation when desired
- filesystem-backed package source and compiled artifact cache
- Node-safe host bindings
- optional CLI for evaluating formulas, compiling packages, inspecting
  generated JavaScript, and running spreadsheet fixtures

The compiler core, AST, resolver, typechecker, IR, JS emitter, package
compiler, and spreadsheet recalculation engine should be shared across both
targets. Environment-specific code should sit behind adapters:

- `RuntimeAdapter`
- `PackageProvider`
- `PackageArtifactCache`
- `HostBindingProvider`
- `WorkerAdapter`

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
- current-sheet cell references through the reserved pseudo package `sheet`,
  for example `sheet.A1`, `sheet.$A$1`, `sheet.A$1`, and `sheet.$A1`
- current-sheet range references through `sheet`, with `sheet.A1:B10` as the
  canonical static range syntax
- cross-sheet cell and range references through sheet-name namespaces, for
  example `Data.A1`, `Data.$A$1`, and `Data.A1:B10`
- dynamic cell and range references through explicit sheet namespace calls
  where the address expression has type `string`, for example
  `sheet.Cell(prefix + string(row))`, `sheet.Range("A1:B10")`,
  `Data.Cell(addr)`, and `Data.Range(start + ":" + end)`

Optional after v1:

- array/slice literals
- indexing
- struct literals

### Spreadsheet Reference Syntax

Spreadsheet references must be explicit so local variables, package names, and
cell names are never ambiguous. The reserved pseudo package `sheet` denotes the
current sheet. Other sheets are exposed as sheet-name namespaces, using the
same selector shape as Go package selectors:

```go
sheet.A1
sheet.$A$1
sheet.A$1
sheet.$A1
sheet.A1:B10
sheet.$A$1:$B$10
Data.A1
Data.$A$1:$B$10
sheet.Cell("A" + string(row))
sheet.Range(start + ":" + end)
Data.Cell(addr)
Data.Range(start + ":" + end)
```

Rules:

- Bare `A1` is an ordinary identifier candidate, not a cell reference.
- The resolver reserves `sheet`; users cannot declare a local variable, source
  package import alias, or package named `sheet` in Go-junior formula scope.
- Sheet names are package namespaces. A selector namespace such as `Data` may
  resolve to either a sheet namespace or an imported/source package namespace,
  but not both in the same formula environment.
- The current sheet is always available through the stable alias `sheet`; it
  may also be available by its real sheet namespace when that namespace is valid
  and not ambiguous.
- A collision between a sheet name and a package import alias/source package
  name is a compile-time diagnostic. A future alias mechanism can relax this,
  but the initial implementation should reject ambiguity rather than guessing.
- Sheet names used directly in formulas must be valid Go-junior identifiers.
  Workbook UI can later provide aliases for display names that contain spaces,
  punctuation, or other non-identifier characters.
- Imported package exports keep normal Go selector syntax, for example
  `stats.Mean` or `pkg.VariableName`.
- `sheet.A1`, `Data.A1`, `sheet.A1:B10`, and `Data.A1:B10` are parsed by a
  special spreadsheet-reference grammar after a resolved sheet namespace. They
  are not general Go identifiers.
- `$` marks absolute column and/or row, matching the familiar Excel idiom.
- Non-`$` row/column parts are relative for formula copy/fill operations but
  resolve to concrete addresses in the formula's current anchor cell.
- The sheet namespace part is fixed during copy/fill. Relative row/column
  parts remain relative to the formula's anchor cell, even when the target is a
  different sheet namespace.
- Static `sheet` references lower to the same runtime path as dynamic
  `sheet.Cell` / `sheet.Range`, so observed dependency tracking is uniform.
- Static cross-sheet references lower to the same runtime path as dynamic
  `Namespace.Cell` / `Namespace.Range`, so dependency tracking always records
  the concrete sheet namespace plus address or range.

### Statements

Support:

- short variable declarations: `x := expr`
- assignments to locals: `x = expr`
- assignments to exported mutable package variables: `pkg.VariableName = expr`
- `if`, `else if`, `else`
- `return expr`
- `return expr, expr` for multiple return values
- naked `return` only inside functions with named return values
- `defer call(...)`
- expression statements only for calls whose return value can be ignored
- bounded `for` loops in a later stage

Keep ordinary formula cells calculation-oriented by default: they calculate and
return values. Go-junior source packages should still support Go-like mutable
package variables from the start. Any formula that calls a package function
known to read or write mutable package state should be marked `package-state`
and recalculated with that effect classification rather than treated as a pure
cacheable helper. Direct assignment to exported mutable package variables is
also `package-state`. Assigning to `sheet.A1` or `Data.A1` is not package
variable assignment; spreadsheet mutation must go through declared sheet-effect
capabilities. Browser, sheet, graph, and UI mutations should be modeled as
command/action cells or explicit spreadsheet actions with declared
capabilities. Those actions may write cells, produce graph artifacts, or
request UI updates through the host effect API; formulas should not silently
mutate browser or sheet state during ordinary dependency recalculation.

### Functions

Stage v1 can compile one implicit cell function. Stage v2 can add local helper
functions and source package functions. Go-junior functions should support
single returns, multiple returns, and named returns.

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
return clamp(number(sheet.A1), 0, 100)
```

Multiple returns:

```go
func DivMod(x int, y int) (int, int) {
    return x / y, x % y
}

q, r := DivMod(17, 5)
```

Named returns:

```go
func SplitTotal(x float64) (half float64, rest float64) {
    half = x / 2
    rest = x - half
    return
}
```

`defer` should follow Go's essential semantics:

- a `defer` statement defers a function or method call
- deferred call arguments are evaluated immediately
- deferred calls execute in LIFO order when the surrounding function exits
- deferred return values are ignored
- named return values are assigned before defers run; future closure/pointer
  support may allow deferred code to observe or mutate them exactly like Go
- the first implementation can exclude closures while still supporting
  `defer someCall(args...)`

No closures in the first implementation. Local helper functions should be pure,
deterministic, and compiled into private generated JavaScript functions.

### Source Packages

Go-junior should support Go-junior-compatible source packages so formulas can
call reusable libraries without those libraries being built into the core host
API. These packages are source Go-like packages constrained to the Go-junior
language subset.

Supported package shape:

```go
package stats

import "mathx"

func Mean(r Range) float64 {
    total := 0.0
    n := 0
    for _, v := range r {
        total = total + number(v)
        n = n + 1
    }
    return total / float64(n)
}
```

Package support should include:

- `package name` declarations for package source units.
- explicit imports by package path, for example `import "stats"`.
- exported functions, constants, mutable package variables, and later exported
  types.
- package-private helpers.
- package-level `var` declarations with deterministic initializers.
- Go-like mutable package variable reads and writes inside package functions.
- a simplified deterministic package initialization order defined by the
  Go-junior package graph, not the full Go specification.
- topological compilation of package import graphs.
- rejection of import cycles.
- support for `defer`, multiple returns, and named returns inside package
  functions.
- rejection of excluded Go features inside packages.
- package ABI metadata so formulas can typecheck calls and package variable
  reads/writes before loading code.
- package export metadata that marks functions and variables as pure,
  package-state-reading, or package-state-mutating.

Cell snippets should be able to call packages through the sheet environment:

```go
return stats.Mean(sheet.A1:A10)
```

The package resolver should distinguish three sources:

- built-in packages shipped with the webui
- user/workbook packages stored with the sheet or in browser storage
- trusted host packages that are already compiled to JavaScript or WASM

Full arbitrary Go packages are not a v1 target. A package can be used as source
only if it stays inside the Go-junior subset. Existing full-Go libraries that
use goroutines, channels, reflection, unsafe, or unsupported runtime features
should be exposed through typed host/WASM bindings instead.

### Bounded Loops

When loops are added, support only bounded forms:

```go
for i := 0; i < 10; i = i + 1 {
    sum = sum + i
}
```

Runtime should enforce a fuel counter regardless of static bounds. Range loops
over finite spreadsheet ranges should be supported as an early package-library
feature:

```go
for _, v := range sheet.A1:A10 {
    sum = sum + number(v)
}
```

The runtime must bound range iteration by the concrete range size and the same
fuel counter used for all loops.

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

## Capability and Effect Model

Go-junior should support controlled dynamic JavaScript and controlled browser
mutation through explicit capabilities. The distinction is between ambient
authority, which is disallowed, and declared authority, which is part of the
typed host contract.

Capability categories:

- `pure`: deterministic calculation helpers. Safe in ordinary formula cells.
- `wasm`: calls into allowlisted Go/WASM services. Safe in formulas if the
  binding is deterministic and side-effect-free.
- `package-state`: reads or writes mutable Go-junior package variables. Legal
  from the start so source packages match ordinary Go semantics, but formulas
  that call package-state exports must be treated as effectful/volatile for
  cache and recalculation purposes.
- `dynamic-js`: JavaScript-backed helper functions supplied by the host. The
  helper implementation may use dynamic JavaScript, but Go-junior code sees
  only a typed function signature. A dynamic-JS binding can still be
  formula-safe if it is declared deterministic, side-effect-free, and approved
  for formula context.
- `sheet-effect`: explicit mutations to spreadsheet state, such as setting a
  cell, adding a sheet, or creating a named output object.
- `graph-effect`: graph production and rendering requests, such as emitting
  nodes/edges, chart specs, DOT, Vega-lite specs, Cytoscape elements, or an Ivy
  concept/ARG visualization payload.
- `ui-effect`: explicit UI requests, such as opening a panel, displaying a
  diagnostic, or focusing a result.

Formula recalculation should run `pure`, approved deterministic `wasm`,
approved formula-safe `dynamic-js`, and explicitly allowed `package-state`
exports. A formula that uses `package-state` is not pure: it should not be
common-subexpression cached, it should be evaluated in the scheduler's stable
order, and its package state lifecycle must be deterministic. Browser and
workbook effect capabilities should require an action context, for example a
user-triggered command, button, menu action, or explicitly marked command cell.
This prevents normal dependency recomputation from repeatedly mutating sheets
or redrawing browser state.

Example host declaration shape:

```ts
const hostSpec = {
  math: {
    Sqrt: fn(["float64"], "float64", { capability: "pure" }),
  },
  js: {
    EvalTrustedHelper: fn(["string"], "Value", {
      capability: "dynamic-js",
      formulaSafe: true,
    }),
  },
  graph: {
    Show: fn(["GraphSpec"], "Value", { capability: "graph-effect" }),
  },
  sheet: {
    SetCell: fn(["string", "Value"], "Value", { capability: "sheet-effect" }),
  },
};
```

The resolver and typechecker should know the current execution context. A
formula context rejects `sheet-effect`, `graph-effect`, `ui-effect`, and
non-formula-safe `dynamic-js` calls. An action context may allow them if the
spreadsheet host grants those capabilities for that invocation.

## Source Package Compilation and Code Cache

The runtime should be able to JIT compile Go-junior-compatible source packages
in the browser and cache the compiled output. This lets the spreadsheet behave
like a scripting environment with reusable libraries.

Package compilation pipeline:

```
package sources
  -> parse each file
  -> validate package declaration
  -> resolve imports
  -> typecheck package exports and internals
  -> emit JS package module or future Wasm package module
  -> persist compiled package artifact in browser cache
```

The package cache should live in IndexedDB in the browser and in a filesystem
cache for Node.js. Cache Storage is optional in the browser for large auxiliary
artifacts. Node's default cache can be a content-addressed directory under a
user-provided path or a temp test directory; later it can move to SQLite if
querying cache metadata becomes useful. Cache keys must include:

- package path and package version/name
- content hashes of all package source files
- transitive dependency cache keys
- compiler version
- target backend, for example `js-source` or `wasm-stencil`
- package ABI version
- host spec version for host/WASM calls used by the package
- capability policy, because formula-safe and action-capable packages have
  different authority

Cached artifacts should include:

- exported signature table
- exported package variable table with mutability and effect metadata
- package diagnostics, if compilation failed
- generated JavaScript source or compiled function factory
- future Wasm bytes or module metadata
- source map or source-span metadata
- dependency metadata for imported packages

Package compilation should run in the browser worker for the webui and in the
Node runtime or a Node `worker_threads` worker for command-line use. Formula
compilation may request a package by path; the runtime resolves, compiles,
caches, and links it before compiling the formula that imports or references
it. Stale package compilations must be ignored by generation token just like
stale formula evaluations.

Package linking rules:

- formulas call package exports through generated package slots, not globals
- packages call imported package exports through generated slots
- package initialization should be minimal and deterministic
- package-level constants are allowed
- package-level variables are mutable by default, matching Go semantics
- package variable initialization runs once per package instance using a
  deterministic topological order over the package import graph
- a workbook/runtime generation owns one package instance graph; source/cache
  invalidation creates a new package instance graph
- compiled package artifacts must not persist current mutable variable values as
  executable artifacts; runtime state is separate from compiled code
- formulas that call functions which read or write mutable package variables are
  marked `package-state`, evaluated in stable scheduler order, and never
  optimized as pure/idempotent helpers

This model allows existing source libraries to be loaded when they are written
in the Go-junior subset, while still keeping full-Go/WASM libraries available
through typed host bindings.

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
- Each compiled formula receives only declared `(ctx, host, budget)` authority.
- Effectful actions receive an explicit effect-capable host object; value
  formulas receive a pure host view.
- Literal and dynamic cell/range reads both route through `ctx` so dependency
  observation is uniform.

Example generated JavaScript:

```js
"use strict";
return function _gj_cell(ctx, host, budget) {
  let _v0 = ctx.cell("sheet", "A1");
  let _v1 = ctx.cell("sheet", "B1");
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
- compile source packages and cache compiled package artifacts
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
- declared dependency set
- observed dependency set from the most recent evaluation
- compiled function cache key
- last computed value
- last runtime error, if any

The spreadsheet engine should maintain a dependency graph:

```
cell -> direct dependencies
dependency -> reverse dependents
```

Dependency records should always include the target sheet namespace or stable
sheet ID plus the normalized address/range. Current-sheet `sheet.A1` and
cross-sheet `Data.A1` should therefore have the same graph representation apart
from the sheet namespace field.

Each formula should track two dependency sets:

- `declaredDeps`: dependencies that can be seen statically, such as
  `sheet.A1`, `sheet.A1:B10`, `Data.A1`, and `Data.A1:B10` references.
- `observedDeps`: exact cells and ranges read during the most recent
  evaluation, including dynamic references.

The scheduler should treat `observedDeps` as authoritative after a successful
or diagnosable evaluation, because branch choices and dynamic addresses can
change the active dependency set. `declaredDeps` remain useful for validation,
initial scheduling, diagnostics, and conservative fallback.

When a cell changes:

1. Parse and typecheck the changed formula.
2. Extract `declaredDeps`.
3. Mark the changed cell and all reverse dependents from current
   `observedDeps` dirty.
4. Recalculate the affected graph. Each evaluation records all `ctx.cell` and
   `ctx.range` reads as the next `observedDeps`.
5. If a formula's `observedDeps` changed, update dependency edges and reschedule
   any newly affected work.
6. Repeat until values and dependency edges are stable, or until a dependency
   stabilization budget is exhausted.
7. Publish value and diagnostics changes to the UI model.

This makes dynamic references a first-class recalculation feature rather than a
later retrofit.

Bare direct references like `A1` are intentionally not cell references. Use
explicit `sheet.A1` syntax so locals, packages, and cells remain easy for users
and the compiler to distinguish.

## Circular References and Iterative Calculation

Circular references should be rejected by default. This matches the desired
Excel-like model: most circular references are mistakes, so a cycle produces
`#CYCLE!` unless the user explicitly enables iterative calculation.

Iterative calculation can be enabled at sheet/workbook level and later refined
to selected cells or selected strongly connected components if the UI needs
that granularity.

Initial option shape:

```ts
interface IterativeCalculationOptions {
  enabled: boolean;          // default false
  maxIterations: number;     // default 100
  numericTolerance: number;  // default 0.001
  detectOscillation: boolean;// default true
  maxObservedStates: number; // default small, for oscillation detection
}
```

When iterative calculation is disabled:

- Any strongly connected component with more than one formula cell reports
  `#CYCLE!`.
- Any self-dependent formula cell reports `#CYCLE!`.
- The engine should surface the cycle members for diagnostics and UI tracing.

When iterative calculation is enabled:

1. Partition the affected graph into strongly connected components.
2. Evaluate acyclic components in topological order.
3. For each cyclic component, repeatedly evaluate all cells in a stable order
   from the previous iteration's values.
4. After each iteration, compare previous and next values.
5. Stop successfully when values converge:
   - numeric values differ by no more than `numericTolerance`
   - non-numeric values are structurally equal
   - dynamic `observedDeps` have stabilized
6. Stop with `#DIVERGE!` when numeric deltas grow or the iteration limit is
   reached without convergence.
7. Stop with `#OSCILLATE!` when the component repeats a prior value vector or
   alternates between value vectors.
8. Stop with `#UNSTABLE_DEPS!` when dynamic dependency edges keep changing
   beyond the dependency stabilization budget.

The scheduler should publish the final converged values only after the
component settles. During iteration, intermediate graph or UI effects are not
published. Graph cells inside an iterative component should produce a derived
graph value after convergence, then the UI renderer replaces the displayed
graph with that final value.

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
  `panic`, `recover`, and `unsafe` are rejected with explicit diagnostic codes.
- A "supported Go-junior Go idiom" fixture list asserts that `defer`, multiple
  returns, named returns, and source package declarations/imports are accepted
  in the contexts where the language supports them.
- A "selector versus spreadsheet reference" fixture list asserts that bare
  `A1`, `pkg.A1`, `sheet.A1`, `Data.A1`, and `Data.A1:B10` are parsed and
  resolved according to namespace rules.

Acceptance criteria:

- The subset is explicit enough that parser and typechecker work can begin.
- Tests can assert both successful normalized output and exact diagnostics.

### Stage 1: Chevrotain Lexer

Implement the Go-junior token vocabulary with Chevrotain.

Implementation tasks:

- Add Chevrotain as a frontend dependency for the webui TypeScript package.
- Define token classes for identifiers, keywords, numbers, strings, operators,
  delimiters, comments, whitespace, and newlines.
- Use Chevrotain token categories for identifier-like names, keywords, and
  spreadsheet-reference parts where useful.
- Define lexer behavior for spreadsheet address fragments such as `A1`, `$A$1`,
  `A$1`, `$A1`, and range separator `:`.
- Treat `A1`-style tokens as both address-like and identifier-like where
  needed, because bare `A1` is a legal ordinary identifier candidate until the
  resolver proves a surrounding selector namespace is a sheet.
- Keep enough newline information to support explicit statement boundaries or a
  small Go-like semicolon insertion/token-normalization pass.
- Decide whether comments and whitespace are skipped or retained for diagnostic
  and formatting metadata.
- Normalize Chevrotain token location data into Go-junior source spans.
- Convert Chevrotain lexing errors into Go-junior diagnostics.
- Reject unsupported tokens early, including channel receive/send syntax.

Tests:

- Tokenizes literals: decimal integers, floats, quoted strings, escaped
  strings, booleans.
- Tokenizes operators and delimiters.
- Tokenizes address-like fragments used in spreadsheet references, including
  `$A$1`, `A$1`, `$A1`, and range separators.
- Does not accidentally tokenize `A1foo` or `A1_foo` as a cell reference plus
  trailing identifier.
- Tokenizes `Data.A1:B10`, `sheet.$A$1`, and `Data.$A$1:$B$10` into stable
  token sequences.
- Allows bare `A1` to remain parseable as an identifier-like token.
- Tokenizes selector-looking package references like `pkg.VariableName` without
  assuming they are spreadsheet references.
- Preserves source spans across newlines and comments.
- Rejects unterminated strings with a useful span.
- Rejects unsupported rune literals if they are not in v1.
- Rejects channel operator `<-`.
- Rejects malformed numbers.
- Converts Chevrotain lexing errors into stable Go-junior diagnostic codes.
- Snapshot tests for representative snippets.

Acceptance criteria:

- Lexer never throws for user input; it returns tokens plus diagnostics.
- Every token has a stable span.
- The token vocabulary is explicit enough for Chevrotain parser self-analysis
  to run during parser construction.

### Stage 2: Chevrotain Parser and AST Builder

Implement a Chevrotain `CstParser` for Go-junior and convert the CST into the
compiler-owned AST.

Implementation tasks:

- Define Chevrotain parser rules for the Go-junior subset.
- Run Chevrotain parser self-analysis at construction time and treat grammar
  ambiguities as test failures.
- Parse expression-form and function-body-form programs.
- Parse statements: declarations, assignments, if/else, returns, expression
  statements.
- Parse `defer` call statements.
- Parse multiple return expressions.
- Parse named return function signatures.
- Parse calls and selector expressions.
- Represent operator precedence with layered Chevrotain rules, for example
  `or -> and -> equality -> compare -> add -> mul -> unary -> primary`.
- Parse special current-sheet and cross-sheet cell/range references.
- Parse ambiguous statement prefixes, such as assignment versus expression
  statement, with explicit lookahead or factored grammar rules.
- Represent `Name.A1` as a selector-or-cell-reference candidate in the AST
  until resolver determines whether `Name` is a sheet namespace or a package
  namespace.
- Represent `Name.A1:B10` as a spreadsheet range candidate because the colon
  form is not ordinary Go selector syntax.
- Parse optional local helper functions after v1 expression/body support.
- Convert Chevrotain CST nodes into Go-junior AST nodes.
- Preserve AST spans for all nodes by merging token/CST source ranges.
- Convert Chevrotain parser errors into Go-junior diagnostics.
- Enable and tune Chevrotain recovery where it improves multi-error reporting.
- Optionally emit Chevrotain syntax diagrams in developer/debug builds.

Tests:

- Parses arithmetic precedence correctly.
- Parses nested parentheses.
- Parses unary and binary operations.
- Parses `if`, `else if`, and `else`.
- Parses short declarations and assignments.
- Parses selector assignments like `pkg.VariableName = expr`.
- Parses return statements.
- Parses multiple return values.
- Parses named return values and naked returns inside named-return functions.
- Parses `defer` call statements.
- Parses selector calls like `math.Sqrt(x)`.
- Parses `sheet.A1`, `sheet.$A$1`, `sheet.A$1`, and `sheet.$A1`.
- Parses `sheet.A1:B10` and `sheet.$A$1:$B$10`.
- Parses cross-sheet `Data.A1`, `Data.$A$1`, `Data.A1:B10`, and
  `Data.$A$1:$B$10`.
- Parses dynamic `sheet.Cell(addr)` and `sheet.Range(addr)`.
- Parses dynamic cross-sheet `Data.Cell(addr)` and `Data.Range(addr)`.
- Parses spreadsheet references only when the selector namespace is syntactically
  eligible; resolver later decides whether the namespace is a sheet or package.
- Parses `pkg.A1` as a selector-or-cell candidate so the resolver can treat it
  as a package export when `pkg` is a package namespace.
- Rejects `go f()`, `select {}`, channel sends, receives, imports, package
  declarations in cell snippets, labels, `goto`, `panic`, `recover`, and
  `unsafe`.
- Chevrotain self-analysis succeeds with no unresolved ambiguities.
- Parser diagnostics wrap Chevrotain errors with stable Go-junior diagnostic
  codes and source spans.
- Error recovery reports multiple syntax errors in one source where possible.
- CST-to-AST conversion preserves spans for expressions, statements,
  spreadsheet references, and function signatures.
- Golden AST tests for small valid programs.

Acceptance criteria:

- Valid v1 syntax produces a complete AST.
- Invalid or unsupported syntax produces stable diagnostics without crashing.
- Later compiler stages do not depend on Chevrotain CST shape; they consume
  only Go-junior AST nodes and diagnostics.

### Stage 3: AST Normalization

Normalize program forms into one internal function-like representation.

Implementation tasks:

- Expression form becomes `return expr`.
- Function body form becomes an implicit cell function.
- Named returns normalize to explicit return locals.
- Insert explicit return requirement checks where needed.
- Normalize syntactic sugar into canonical AST nodes.

Tests:

- Expression source normalizes to one return statement.
- Function body source preserves user statements.
- Multiple return values normalize to tuple return IR shape.
- Named return values normalize to local return slots.
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
- Resolve the reserved `sheet` pseudo package and conversion helpers.
- Resolve workbook sheet namespaces as package-like selector namespaces.
- Reject collisions between sheet namespaces and imported/source package
  namespaces unless a future explicit alias mechanism is present.
- Reject local declarations, import aliases, or package names that collide with
  reserved `sheet` in formula scope.
- Generate stable symbol IDs for locals and functions.

Tests:

- Resolves locals in nested `if` blocks.
- Rejects unknown identifiers.
- Rejects duplicate declarations.
- Rejects assignment to undeclared variables.
- Rejects assignment to host symbols and predeclared functions.
- Resolves assignment to exported mutable package variables.
- Rejects assignment to package constants, unexported package variables, and
  read-only host bindings.
- Rejects assignment to `sheet.A1`, `Data.A1`, or other spreadsheet references
  outside explicit sheet-effect APIs.
- Resolves host calls with namespace and function names.
- Resolves imported package exported variables with normal package selector
  syntax.
- Resolves `pkg.A1` as an exported package symbol, not a cell reference, when
  `pkg` is a package namespace.
- Resolves `sheet` references as spreadsheet references, not package variables.
- Resolves cross-sheet references like `Data.A1` and `Data.A1:B10` as
  spreadsheet references when `Data` is a known sheet namespace.
- Rejects ambiguous selector namespaces when a sheet namespace and package alias
  share the same name.
- Rejects unknown host namespaces/functions.
- Rejects unknown sheet namespaces in spreadsheet references.
- Rejects local variables named `sheet`.
- Rejects package imports or aliases named `sheet`.
- Does not resolve against `window`, `document`, `globalThis`, or other browser
  globals.

Acceptance criteria:

- Every identifier use is bound to a known symbol or produces a diagnostic.

### Stage 5: Typechecker

Add static typing for expressions, statements, returns, and calls.

Implementation tasks:

- Define type model and assignability rules.
- Type arithmetic, comparisons, booleans, strings, and returns.
- Type current-sheet and cross-sheet cell/range references plus dynamic
  `Namespace.Cell` / `Namespace.Range` helpers.
- Type assignment to locals and exported mutable package variables.
- Type host calls from the host spec.
- Type tuple return values and tuple assignment from multiple-return calls.
- Type named return values and naked returns.
- Type `defer` statements; deferred expressions must be calls and their return
  values are ignored.
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
- Multiple return expression count and types match the function signature.
- Named return functions allow naked `return`.
- Non-named-return functions reject naked `return`.
- Tuple assignment from a multiple-return call binds each target type.
- Assignment to an exported package variable requires assignability to that
  package variable's declared type and marks the formula `package-state`.
- Assignment to a package constant or read-only binding is rejected.
- Assignment to a spreadsheet reference is rejected unless it goes through a
  declared sheet-effect capability.
- Deferred calls validate argument types and ignore return values.
- Deferred non-call expressions are rejected.
- Host calls validate arity and argument types.
- Host return types flow into later expressions.
- Cell values require explicit conversion when needed, if the type policy uses
  `Value`.
- `sheet.A1` and `sheet.$A$1` produce `Value`.
- `sheet.A1:B10` produces `Range`.
- `Data.A1` and `Data.$A$1` produce `Value`.
- `Data.A1:B10` produces `Range`.
- `Data.Cell(addr)` and `Data.Range(addr)` require string addresses and produce
  `Value` and `Range`.
- Diagnostics point at the offending expression.

Acceptance criteria:

- Typechecking rejects invalid programs before emission.
- Typechecked AST nodes carry resolved types.

### Stage 6: Dependency Extraction

Extract statically visible spreadsheet dependencies from typed AST and mark
which formulas require runtime dependency observation.

Implementation tasks:

- Recognize static current-sheet and cross-sheet cell/range references.
- Validate cell and range addresses.
- Record statically known direct dependencies in normalized
  `(sheet namespace, address/range)` format.
- Recognize dynamic `Namespace.Cell(expr)` and `Namespace.Range(expr)` calls
  where `Namespace` is a sheet namespace and `expr` has type `string`.
- Mark formulas with dynamic references as requiring observed dependency
  tracking.
- Preserve enough source metadata to diagnose invalid literal references and
  invalid runtime references.
- Keep dependency extraction independent from JS emission.

Tests:

- Extracts one cell dependency.
- Extracts multiple dependencies from expressions and branches.
- Extracts range dependencies.
- Extracts absolute and mixed absolute/relative references from `sheet.$A$1`,
  `sheet.A$1`, and `sheet.$A1`.
- Extracts range dependencies from `sheet.A1:B10` and
  `sheet.$A$1:$B$10`.
- Extracts cross-sheet dependencies from `Data.A1`, `Data.$A$1`, and
  `Data.A1:B10`.
- Preserves the sheet namespace in each dependency record.
- Marks `sheet.Cell(addr)` as dynamic when `addr` is a variable.
- Marks `sheet.Range(start + ":" + end)` as dynamic when the address is
  computed.
- Marks `Data.Cell(addr)` and `Data.Range(addr)` as dynamic cross-sheet
  references.
- Rejects invalid cell/range strings.
- Handles duplicate dependencies by deduplicating.
- Does not count strings passed to unrelated functions as dependencies.
- Distinguishes literal dependencies from dynamic-reference markers.
- Produces stable output order.

Acceptance criteria:

- The spreadsheet engine gets `declaredDeps` for static cases and an explicit
  runtime-observation requirement for dynamic cases.

### Stage 7: Typed IR

Lower typed AST to a compact IR designed for both JS template emission and a
future WebAssembly backend.

Implementation tasks:

- Define IR nodes for constants, locals, arithmetic, comparisons, boolean ops,
  branches, returns, tuple returns, named-return slots, defers, calls,
  package-state reads/writes, cell/range reads, and later loops.
- Preserve diagnostic/source mapping metadata.
- Add optional constant folding for simple literals.
- Add explicit conversions where needed.
- Represent host calls by binding ID, not by text name.

Tests:

- Golden IR tests for expressions, branches, locals, and host calls.
- Constant folding tests if implemented.
- IR contains no raw user identifier names except for metadata.
- IR host calls contain host binding IDs.
- IR cell reads contain normalized sheet namespace plus address.
- IR dynamic cell/range reads preserve the address expression and lower to
  runtime-observed `ctx.cell` / `ctx.range` calls.
- IR package variable reads and writes carry package slot IDs, not raw selector
  text.
- IR multiple returns use an explicit tuple shape.
- IR named returns use explicit local return slots.
- IR defers preserve LIFO execution order and immediate argument evaluation.

Acceptance criteria:

- JS emission can be implemented without consulting raw source text.

### Stage 7A: Source Package Resolver, Compiler, and Cache

Implement package-shaped Go-junior libraries before treating external package
calls as an integration detail.

Implementation tasks:

- Define package manifest and package source file shapes.
- Parse package declarations and import declarations for source package units.
- Resolve package imports through built-in, workbook, browser-cache, and trusted
  host providers.
- Build a package import graph and reject import cycles.
- Typecheck package exports and internals.
- Produce an exported signature table for each package.
- Produce an exported package variable table with mutability, type, and effect
  metadata.
- Support package-level `var` declarations with deterministic initializers.
- Support package variable reads and assignments inside package functions.
- Mark package functions that read or write mutable package variables as
  `package-state` exports.
- Lower package functions to IR.
- Emit package JavaScript using the same source copy-and-patch backend as
  formulas.
- Link formulas to package exports through generated package slots.
- Persist compiled package artifacts in IndexedDB.
- Invalidate package artifacts when source, transitive dependency, compiler
  version, backend, ABI version, host spec, or capability policy changes.
- Keep compiled package artifacts separate from runtime package variable state.
- Keep package compilation in the worker.

Tests:

- Parses a single-file package with one exported function.
- Parses a multi-file package and merges package scope correctly.
- Rejects mixed package names in one package unit.
- Rejects unsupported Go features in package source.
- Resolves imports from a fake built-in package provider.
- Resolves imports from a fake workbook package provider.
- Rejects unknown imports with source-span diagnostics.
- Rejects import cycles across two and three packages.
- Typechecks exported functions before formula compilation.
- Formula can call an exported package function.
- Formula can read an exported package constant or mutable variable using
  normal Go selector syntax, for example `pkg.VariableName`.
- Formula can assign to an exported mutable package variable using normal Go
  selector syntax, for example `pkg.VariableName = 10`.
- Formula can call an exported package function that mutates package state, and
  the formula is marked `package-state`.
- Package functions can read and write exported and unexported package
  variables according to Go visibility rules.
- Package variable state persists across calls within one runtime/package
  instance.
- Package variable state resets when the runtime/package instance is recreated
  by source or dependency invalidation.
- Formula cannot call an unexported package helper.
- Formula cannot read an unexported package variable.
- Formula cannot assign to an unexported package variable or package constant.
- Package can call an imported package export.
- Package function using `Range` iterates over a finite range with fuel limits.
- Package compilation emits a stable exported signature table.
- Package compilation emits stable package variable metadata.
- Package cache hits when source and transitive dependencies are unchanged.
- Package cache misses when any package source file changes.
- Package cache misses when a transitive dependency changes.
- Package cache misses when compiler version, target backend, ABI version, host
  spec version, or capability policy changes.
- Failed package compilation stores diagnostics but not executable artifacts.
- Worker package compile results are ignored if their generation token is stale.
- Cached package artifacts never persist raw executable authority beyond the
  generated JS/Wasm and typed metadata.
- Cached package artifacts do not persist current mutable package variable
  values unless a separate explicit workbook-session state format is added.

Acceptance criteria:

- Go-junior formulas can call Go-junior-compatible source packages compiled and
  cached in the browser.
- Full-Go packages outside the Go-junior subset remain available only through
  typed host/WASM bindings or trusted precompiled package providers.

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
- Generated JS for multiple return functions returns a stable tuple/array
  representation.
- Generated JS for named returns assigns return locals and emits a final return
  tuple/value.
- Generated JS for `defer` emits a LIFO defer stack and `try/finally` or
  equivalent control flow.
- Generated JS for a package function matches golden snapshots.
- Formula-generated JS calls package exports through generated package slots.
- Package-generated JS calls imported package exports through generated package
  slots.
- Package-generated JS reads and writes package variables through generated
  package state slots, not globals.
- Generated JS never contains raw user variable names.
- String literals are escaped with `JSON.stringify`.
- Host calls route through `host` using resolved binding paths or IDs.
- Static current-sheet and cross-sheet cell reads route through `ctx.cell`.
- Static current-sheet and cross-sheet range reads route through `ctx.range`.
- Dynamic `Namespace.Cell` reads route through `ctx.cell`.
- Dynamic `Namespace.Range` reads route through `ctx.range`.
- Unsupported IR nodes fail with internal compiler diagnostics.
- Generated source parses via `new Function` in tests.
- Malicious source snippets cannot break out through emitted JS:
  - `"); globalThis.evil = true; ("`
  - identifiers named like JS keywords
  - strings containing backticks and `${...}`
  - source containing `</script>`

Acceptance criteria:

- Safe, valid JavaScript is produced for all valid v1 IR programs.

### Stage 9: Function and Package Artifact Compilation Cache

Compile generated JavaScript to callable formula/package functions and cache
them.

Implementation tasks:

- Create cache key from source hash, compiler version, host spec version, and
  selected runtime options.
- Include package artifact keys for formulas that link source packages.
- Include transitive package dependency keys for compiled package artifacts.
- Store formula function artifacts in the worker memory cache.
- Store package artifacts in worker memory and IndexedDB.
- Store mutable package variable runtime state separately from compiled package
  artifacts.
- Compile with `new Function`.
- Return structured compile errors if browser/CSP rejects dynamic compilation.
- Add a development/debug option to expose generated source.

Tests:

- Same source and host spec hits cache.
- Host spec version changes invalidate cache.
- Compiler version changes invalidate cache.
- Formula cache misses when a linked package artifact changes.
- Package artifact cache survives worker restart by reloading from IndexedDB.
- Package artifact cache survives Node process restart by reloading from the
  filesystem cache.
- Package artifact cache ignores incompatible ABI/backend artifacts.
- Package artifact cache reload does not accidentally restore stale mutable
  package variable values.
- Package runtime state resets or restores only through an explicit runtime
  state policy, never as a side effect of loading compiled code.
- Compile failure returns diagnostic/error object.
- Cached function returns same result as newly compiled function.

Acceptance criteria:

- Repeated recalculation does not repeatedly parse and compile unchanged cells
  or unchanged Go-junior-compatible packages.

### Stage 10: Runtime Evaluation

Implement the evaluator context and spreadsheet value conversion.

Implementation tasks:

- Define `ctx.cell(sheetNamespace, address)`,
  `ctx.range(sheetNamespace, address)`, and value conversion helpers.
- Make `ctx.cell` and `ctx.range` validate normalized addresses at runtime.
- Make every `ctx.cell` and `ctx.range` call record an observed dependency
  before returning or raising a reference diagnostic when possible.
- Maintain package instance state for mutable package variables.
- Evaluate `package-state` formulas in the scheduler's stable order and avoid
  pure-result memoization for them.
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
- Reads cells and ranges from a fake cross-sheet context.
- Records observed dependencies for literal and dynamic cell reads.
- Records observed dependencies for literal and dynamic range reads.
- Observed dependencies include sheet namespace for current-sheet and
  cross-sheet reads.
- Package variable mutation persists across calls in one runtime instance.
- Direct assignment to an exported package variable from formula code mutates
  package state and marks the formula `package-state`.
- Package variable mutation is reset when the package instance graph is
  recreated.
- Produces `#REF!` for invalid dynamic addresses while preserving any
  dependencies read before the invalid reference was constructed.
- Calls fake host functions.
- Converts return values to spreadsheet values.
- Propagates runtime errors as spreadsheet error values.
- Enforces budget on artificial loops once loops exist.
- Division by zero gives expected error or value according to spec.

Acceptance criteria:

- A compiled formula can run against a fake spreadsheet context in unit tests.

### Stage 11: Node.js Runtime and CLI

Implement the Node.js runtime target first so language design can be exercised
quickly from bash before browser integration begins.

Implementation tasks:

- Implement `NodeRuntimeAdapter` using the shared compiler/recalculation core.
- Implement a filesystem-backed `PackageArtifactCache`.
- Implement filesystem/workspace package providers for source packages.
- Implement Node-safe host binding providers for tests.
- Support optional `worker_threads` isolation, but allow direct in-process
  execution for fast unit tests.
- Add a `gojunior` CLI entry point under the frontend package scripts or a
  small Node bin.
- CLI subcommands should include:
  - `eval`: evaluate one formula with JSON-provided cells and packages
  - `compile`: compile formula or package and print diagnostics
  - `run-fixture`: run a spreadsheet/recalculation fixture
  - `inspect-js`: print generated JavaScript for a formula or package
  - `cache`: inspect, clear, or warm package cache entries
- Keep CLI output machine-readable with a JSON mode and human-readable by
  default.

Tests:

- Node runtime evaluates a simple formula without DOM or browser APIs.
- Node runtime evaluates a formula with dynamic dependencies and returns
  observed deps.
- Node runtime evaluates formulas against multi-sheet fixture data.
- Node runtime compiles and calls a source package.
- Node runtime preserves mutable package variable state across calls in one
  runtime session.
- Node runtime reuses filesystem package cache on second run.
- Node runtime invalidates cache when package source changes.
- Node runtime can run without worker_threads for fast unit tests.
- Node runtime can run with worker_threads for isolation tests.
- CLI `eval` returns value, diagnostics, and observed dependencies.
- CLI `compile` returns package/formula diagnostics without evaluation.
- CLI `run-fixture` executes a multi-cell dependency graph.
- CLI `inspect-js` prints generated JS without executing it.
- CLI `cache clear` removes package artifacts from the chosen cache directory.
- Node host bindings cannot access browser-only APIs.

Acceptance criteria:

- Developers can run Go-junior formulas, packages, and spreadsheet fixtures
  from bash through Node.js.
- The Node target uses the same compiler, IR, emitter, typechecker, and
  recalculation engine as the browser target.

### Stage 11A: Browser Worker Runtime

After the CLI/runtime semantics are useful, move the same compile/evaluate
protocol into a Web Worker for browser integration.

Implementation tasks:

- Define browser worker protocol: compile, evaluateCell, evaluateBatch,
  disposeCache, updateHostSpec, compilePackage, resolvePackage,
  disposePackageCache, and diagnostics messages.
- Include `observedDeps`, runtime diagnostics, value version, and dependency
  generation in evaluation responses.
- Keep host/WASM services behind explicit RPC endpoints.
- Support cancellation by generation token.
- Ensure stale results are ignored by the main thread.

Tests:

- Browser worker compiles valid formulas.
- Browser worker compiles valid source packages.
- Browser worker links formula compilation against cached package artifacts.
- Browser worker reports diagnostics for invalid formulas.
- Browser worker reports diagnostics for invalid source packages.
- Browser worker evaluates a batch in dependency order provided by the main
  thread.
- Browser worker evaluation returns observed dependencies for each evaluated
  formula.
- Main thread can ignore stale observed dependency results by generation token.
- Cancellation token suppresses stale results.
- Browser worker survives a formula runtime exception.
- Browser worker cache persists across messages.
- Browser worker can be terminated and restarted without losing main-thread
  state.

Acceptance criteria:

- User formulas do not execute on the main UI thread in production mode.

### Stage 12: Spreadsheet Dependency Graph and Recalculation

Implement spreadsheet-level formula management.

Implementation tasks:

- Track formula cells, literal cells, dirty cells, `declaredDeps`,
  `observedDeps`, and reverse dependents.
- Track dependencies across `(sheet namespace, cell address)` keys from the
  start.
- Topologically sort dirty acyclic subgraphs using current `observedDeps`.
- Detect strongly connected components before evaluation.
- Keep iterative calculation disabled by default.
- Support explicit iterative recalculation for opted-in cyclic components or
  sheets using `IterativeCalculationOptions`.
- Compare iterative value vectors with numeric tolerance and structural
  equality.
- Detect non-convergence, divergence, oscillation, and dynamic-dependency
  instability as separate outcomes.
- Incrementally update static dependencies when formulas change.
- Atomically update observed dependencies after each evaluation.
- If observed dependencies change, update reverse edges and reschedule the
  affected graph until both values and observed edges stabilize.
- Enforce a dependency-stabilization budget separate from formula execution
  fuel.
- Publish recalculation results to UI state.

Tests:

- Single formula cell recalculates from literals.
- Changing one literal marks dependents dirty.
- Recalculation order respects dependencies.
- Diamond dependency graph recalculates each formula once.
- Dynamic reference formula observes the concrete cell it reads.
- Cross-sheet static reference formula observes the concrete sheet namespace and
  cell it reads.
- Cross-sheet dynamic reference formula updates observed dependencies when the
  computed address changes.
- Changing the address-driving cell reschedules the formula and moves the
  reverse dependency edge from the old referenced cell to the new one.
- Branch-dependent formula observes only the executed branch plus the guard
  dependencies.
- Changing an unobserved branch dependency does not recalculate until the guard
  changes.
- Dynamic range formula updates observed range edges when the range expression
  changes.
- Observed dependency changes trigger one additional scheduling pass and then
  settle.
- A dependency set that toggles indefinitely reports `#UNSTABLE_DEPS!`.
- Dependency cycle reports `#CYCLE!` when iterative calculation is disabled.
- Self-reference reports `#CYCLE!` when iterative calculation is disabled.
- Enabling iterative calculation for the sheet allows the same cycle to run.
- Opted-in iterative cycle converges when numeric values stabilize within
  tolerance.
- Opted-in iterative cycle converges when non-numeric values become
  structurally equal.
- Opted-in iterative cycle reports `#DIVERGE!` when values fail to converge
  within `maxIterations` or deltas grow according to the configured policy.
- Opted-in iterative cycle reports `#OSCILLATE!` when value vectors repeat or
  alternate.
- Dynamic dependencies inside an iterative component must stabilize along with
  values.
- Dynamic dependencies inside an iterative component that do not stabilize
  report `#UNSTABLE_DEPS!`.
- Package-state formulas in an iterative component are evaluated in stable order
  and report diagnostics if package-state effects make convergence
  non-deterministic under the configured policy.
- Removing a formula removes old dependency edges.
- Invalid formula source preserves previous value or produces error according
  to spec.
- Batch updates coalesce recalculation.

Acceptance criteria:

- The spreadsheet engine has real formula semantics independent of the DOM.
- Dynamic references are part of the first recalculation architecture, not a
  later extension.

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

Expose selected Go/WASM package functions through typed host wrappers and
trusted precompiled package providers.

Implementation tasks:

- Define host binding declaration format in TypeScript.
- Wrap existing Go WASM exports or `wasm_exec.js` globals behind stable async or
  sync functions.
- Define a trusted precompiled package provider interface for packages that are
  delivered as already-compiled JavaScript or Wasm plus ABI metadata.
- Version the host spec.
- Add marshalling for primitive values and structured spreadsheet values.
- Add error conversion from Go/WASM failures to spreadsheet errors.

Tests:

- Fake host binding tests for arity/type enforcement.
- Real smoke test loads a small Go WASM module if available in the existing
  webui runtime.
- Host function can be called from Go-junior formula.
- Formula can call an export from a trusted precompiled package provider.
- Trusted precompiled package ABI is typechecked like source package ABI.
- Host function errors become spreadsheet errors.
- Unknown host function is rejected at typecheck time.
- Host spec version invalidates compiled function cache.
- Effectful host bindings are rejected in formula context.
- Effectful host bindings are allowed in action context when the host grants
  the required capability.
- Graph-producing bindings can return or publish a graph spec through the
  declared host API.

Acceptance criteria:

- Go-junior formulas can call allowlisted Go/WASM functions safely.
- Go-junior formulas can call trusted precompiled Go package exports through
  the same package-call path used by source packages.
- Go-junior actions can call allowlisted sheet, graph, UI, and dynamic-JS
  capabilities without ambient browser access.

### Stage 15: Persistence

Persist formulas, calculated values, diagnostics, and engine metadata.

Implementation tasks:

- Define persisted formula cell shape.
- Define persisted workbook package source shape.
- Persist raw Go-junior source, not generated JavaScript.
- Persist source package text/manifests, not generated package JavaScript or
  Wasm as authoritative workbook state.
- Define whether mutable package variable current values are workbook-session
  state, saved workbook state, or reset-on-open state. The initial
  implementation should make this policy explicit and test it rather than
  letting package caches decide implicitly.
- Persisting `observedDeps` is optional cache data only; load must recompute
  observed dependencies from source and current sheet values before trusting
  recalculation edges.
- Persisted package artifact cache entries are optional acceleration only and
  must be validated by cache key before use.
- Persist compiler version and host spec version for cache invalidation.
- On load, reparse/retype/recompile formulas rather than trusting old compiled
  source.

Tests:

- Save/load preserves raw formula text.
- Save/load preserves workbook package source text and manifests.
- Save/load handles mutable package variable state according to the explicit
  workbook policy.
- Save/load recalculates values from current literals.
- Save/load recalculates cross-sheet dependencies from current sheet values.
- Save/load with dynamic references rebuilds observed dependencies on load.
- Save/load can reuse a valid package artifact cache but rebuilds packages when
  cache keys do not match.
- Loading old compiler version invalidates cache.
- Loading unknown host function reports `#NAME?` or diagnostic.
- Loading a workbook package with diagnostics surfaces those package diagnostics
  before dependent formulas run.
- Corrupt persisted formulas fail safely.

Acceptance criteria:

- Spreadsheet state can round-trip without persisting executable JavaScript.
- Workbook package state can round-trip without trusting persisted executable
  artifacts.

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
- Confirm formula contexts receive only pure, approved deterministic WASM,
  approved formula-safe dynamic-JS, and explicitly allowed package-state
  capabilities.
- Confirm action contexts receive only explicitly granted effect capabilities.
- Confirm package imports resolve only through approved providers.
- Confirm package cache artifacts are keyed and validated before execution.
- Confirm mutable package variable state is isolated per workbook/runtime
  generation and is not mixed across cache keys.
- Add CSP compatibility fallback decision: interpreter or explicit unsupported
  message.
- Add runtime budget defaults.
- Add maximum source length, AST size, dependency count, and recursion limits.

Tests:

- Fuzz Chevrotain lexer/parser and CST-to-AST conversion with random input.
- Property tests ensure generated JS contains only generated identifiers for
  user variables.
- Malicious strings and identifiers cannot escape emission.
- Formulas cannot access browser globals directly.
- Formula contexts reject graph, sheet, UI, and dynamic-JS capabilities unless
  explicitly configured as formula-safe.
- Action contexts can produce graph and sheet effects only through the typed
  effect host.
- Dynamic JavaScript-backed helpers cannot receive raw Go-junior source unless
  the host binding explicitly declares and tests that behavior.
- Source packages cannot access browser globals directly.
- Source packages cannot smuggle extra capabilities through package imports.
- Mutable package variables cannot smuggle host, DOM, worker, or workbook
  authority outside the typed capability model.
- Tampered package cache entries are rejected by key/hash/ABI validation.
- Tampered package variable runtime state is rejected or reset according to the
  explicit state persistence policy.
- A package compiled with action capabilities is rejected in formula context
  unless its exported function is explicitly formula-safe.
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

- Add microbenchmarks for Chevrotain lexing, Chevrotain parsing, CST-to-AST
  conversion, typechecker, package resolver, package compiler, emitter,
  `new Function` compile, and evaluation.
- Add spreadsheet recalculation benchmarks for common graph shapes.
- Measure main-thread latency with worker enabled.
- Measure Node CLI cold-start, warm-cache, and fixture-run latency.
- Add cache hit/miss counters.

Tests and benchmarks:

- 1,000 simple formulas compile within target budget.
- 1,000 cached formulas recalculate without recompilation.
- A workbook package graph compiles within target budget on cold cache.
- The same workbook package graph links from IndexedDB cache within target
  budget after worker restart.
- The same workbook package graph links from Node filesystem cache within
  target budget after process restart.
- Node CLI can run representative fixtures fast enough for ordinary unit-test
  loops.
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

- Chevrotain lexer token snapshot tests.
- Chevrotain parser self-analysis tests.
- Parser CST-to-AST golden tests.
- Resolver/typechecker diagnostic tests.
- Package resolver/compiler/cache tests.
- IR golden tests.
- JS emitter golden and security tests.
- Runtime evaluation tests with fake contexts.
- Worker protocol tests.
- Node runtime and CLI tests.
- Spreadsheet dependency/recalculation tests.
- Webui service tests.
- Browser/Playwright tests only for real UI behavior.
- Differential tests between interpreter and JS backend if interpreter exists.
- Fuzz/property tests for lexer, parser, CST-to-AST conversion, and emitter
  hardening.

Every diagnostic-producing stage should test:

- diagnostic code
- message substring
- source span
- recovery behavior

Every successful compiler stage should test:

- normalized output
- source map/span preservation where applicable
- absence of raw unsafe source in emitted artifacts

## Initial Milestone Slice: CLI First

The first useful implementation should run under Node.js before any browser UI
integration. This gives fast unit tests and lets users manually exercise the
language from bash while the design is still fluid.

1. Chevrotain lexer/parser for expression form plus short declarations, `if`,
   and `return`, with CST-to-AST conversion.
2. Types for bool, string, and number.
3. Parser/typechecker coverage for `defer`, multiple returns, and named
   returns.
4. Resolver/typechecker for locals, reserved `sheet`, sheet-name namespaces,
   and one fake host namespace.
5. Dependency extraction for literal current-sheet and cross-sheet references,
   plus runtime observed dependency tracking for dynamic references.
6. One tiny Go-junior-compatible source package compiled by the Node runtime
   and called from a formula, including a mutable package variable.
7. JS source copy-and-patch emitter.
8. Runtime tests that evaluate formulas against a fake spreadsheet context in
   Node.
9. A minimal Node CLI `eval`, `compile`, `run-fixture`, and `inspect-js` path
   for manual bash testing.
10. Browser worker and webui service integration are deferred until the language
   and CLI semantics have been exercised.

Example first formulas:

```go
number(sheet.A1) + number(sheet.B1)
```

```go
return number(Data.A1) + number(sheet.B1)
```

```go
row := number(sheet.B1)
return sheet.Cell("A" + string(row))
```

```go
addr := "A" + string(number(sheet.B1))
return Data.Cell(addr)
```

```go
x := number(sheet.A1) + number(sheet.B1)
if x > 10 {
    return math.Sqrt(x)
}
return x * 2
```

```go
if string(sheet.$A$1) == "" {
    return "missing"
}
return sheet.$A$1
```

Example first source package:

```go
package counter

var Count int

func Next() int {
    Count = Count + 1
    return Count
}
```

Formula calling it:

```go
return counter.Next()
```

## Acceptance Criteria for the Full Project

- Go-junior has a documented, tested subset of Go-like syntax.
- Unsupported Go features are rejected explicitly.
- Valid formulas typecheck before execution.
- Statically visible dependencies are extracted without running formulas.
- Dynamic dependencies are observed during evaluation and update the dependency
  graph automatically.
- Current-sheet and cross-sheet dependencies use the same graph model with
  explicit sheet namespaces.
- Go-junior-compatible source packages compile, link, and cache in the browser.
- Go-junior-compatible source packages compile, link, and cache in Node.js.
- Mutable package variables follow documented Go-like runtime semantics without
  being stored inside compiled package artifacts.
- Formulas compile to JavaScript generated only from compiler-owned stencils.
- Generated code runs in both browser worker-backed and Node.js runtimes.
- Formula cells recalculate incrementally and detect cycles.
- Go/WASM host calls work through typed allowlisted wrappers.
- Spreadsheet state persists raw Go-junior source, not executable JS.
- Security tests cover injection, globals access, runaway execution, and host
  authority boundaries.
- Performance tests show interactive compile/recalculate latency.
