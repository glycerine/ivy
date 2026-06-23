# Go-junior Spreadsheet Language Plan

Plan only. No implementation changes in this document.

## Goal

Build a TypeScript implementation of Go-junior, now targeting full Go language
compatibility for source parsing, typechecking, and execution, plus explicit
spreadsheet syntax extensions. Go-junior should compile to browser-runnable
JavaScript using a compiler-owned CPS/state-machine backend with source-level
copy-and-patch templates where useful. It should be able to call Go-compatible
source packages compiled in-browser or in Node, and it should be able to call
allowlisted Go packages compiled to WebAssembly through typed host wrappers.

The first production target is the Ivy webui analysis spreadsheet. The current
spreadsheet-like pane can then evolve from an editable mock grid into a real
spreadsheet/calculation surface with formula cells, dependencies, diagnostics,
and host/WASM calls.

## Design Principles

- Treat Go-junior as full Go language semantics plus deliberate spreadsheet
  extensions, not as a smaller Go-like DSL. Any deviation from Go must be
  named, tested, and justified by the spreadsheet/browser target.
- Keep execution deterministic where the spreadsheet host needs determinism,
  while still supporting Go's concurrency surface (`go`, channels, `select`,
  `close`, and `recover`) through a deterministic runtime scheduler.
- Preserve Go's readable expression and statement syntax where it helps.
- Include core Go function-structuring idioms that matter for real libraries,
  including `defer`, multiple return values, and named return values.
- Make spreadsheet semantics first-class: cells, ranges, dependencies,
  recalculation, error values, deterministic execution, and bounded work.
- Never emit raw user source into JavaScript. Parse, validate, typecheck, then
  emit compiler-owned JavaScript stencils with sanitized holes.
- Keep the compiler backend replaceable. Start with a JavaScript CPS/state
  machine backend that preserves source spans and types. Leave a path to a
  future WebAssembly stencil backend.
- Go-junior code cannot directly import arbitrary JavaScript or reach ambient
  browser globals; JavaScript, browser, sheet, graph, UI, and WASM
  functionality must be exposed through explicit typed host capabilities.
  Go-junior package imports resolve through an explicit package resolver,
  package manifest, and compiled-code cache.

## Rationale

Spreadsheet formulas are not ordinary scripts. The spreadsheet engine may rerun
them automatically, many times, in dependency order, in a worker, from Node
runtime tests, through the Go `gojr` CLI backed by embedded Node/V8, or inside
iterative recalculation for circular references. The runtime therefore needs
to know what a formula can read, what it can mutate, and whether a call is
pure, stateful, volatile, or effectful.

Typed host capabilities preserve useful power without granting ambient
authority. Go-junior should be able to call JavaScript-backed helpers, draw
graphs, request UI actions, mutate sheets through explicit actions, call Go/WASM
services, and use compiled package libraries. The rule is that those operations
must pass through a typed, auditable contract instead of reaching invisible
global state directly.

This protects:

- Security: workbook code should not directly reach `window`, `document`,
  `globalThis`, cookies, local storage, network APIs, or other browser globals.
- Determinism: the recalculation engine needs to know whether a call is pure,
  package-stateful, volatile, or effectful.
- Dependency tracking: if code can read arbitrary browser state, the
  spreadsheet graph cannot know what should trigger recomputation.
- Worker portability: formulas should run in Node tests and browser workers;
  direct DOM/global APIs break that portability.
- Testing: typed bindings are easy to fake in the Node runtime, the Go `gojr`
  CLI, and unit tests.
- Caching: compiled formulas and packages can be cached only if their host
  capability contract is explicit.
- Backend flexibility: the same typed IR can later run in an interpreter or
  Wasm backend if external calls are abstracted.

Direct ambient access should be rejected:

```go
window.document.body.innerHTML = "..."
globalThis.fetch(...)
eval(userText)
```

Declared host capabilities are allowed:

```go
graph.Show(spec)
sheet.SetCell("A1", value)
js.TrustedRenderChart(spec)
math.Sqrt(x)
ivy.Parse(source)
```

Each binding should declare its type and authority:

```ts
{
  args: ["GraphSpec"],
  returns: "Value",
  capability: "graph-effect",
  formulaSafe: false,
}
```

The distinction is not "no JavaScript". It is "no ambient JavaScript authority".

## Non-goals

- Native machine-code JIT in the browser.
- Ambient access to arbitrary JavaScript or browser globals. Dynamic
  JavaScript-backed helpers are allowed only through explicit typed
  capabilities.
- Unmediated browser mutation. Go-junior may produce graphs, sheet updates, or
  UI effects through declared host capabilities, but it should not reach
  `window`, `document`, `globalThis`, or DOM APIs directly.
- Long-running unbounded calculations on the main UI thread.
- Browser-native cgo execution. cgo packages may be represented through typed
  host/WASM bindings or trusted precompiled adapters, but the TypeScript
  runtime does not execute arbitrary native cgo.
- Exact Go runtime implementation details that are not observable by ordinary
  source programs, such as native stack layout, OS thread scheduling,
  preemptive scheduling timing, or machine pointer identity. Observable Go
  language semantics remain in scope.
- Arbitrary ambient package loading. Source packages are allowed, but they must
  be resolved through the in-browser package registry/cache or a trusted host
  package provider.

## High-level Architecture

```
Go-junior source
  -> TypeScript Go-style scanner/token stream
  -> TypeScript Go-style parser and AST with source spans
  -> TypeScript Go-style resolver/typechecker front end
  -> Go language compatibility validator and spreadsheet syntax validator
  -> resolver and typechecker
  -> package resolver and package compiler
  -> spreadsheet dependency extractor
  -> typed IR
  -> CPS/state-machine lowering for suspendable code
  -> JavaScript source copy-and-patch emitter and runtime scheduler
  -> function/package compiler cache
  -> worker-backed runtime
  -> typed package and host/WASM bindings
```

The compiler package should be independent from the Ivy runtime. The Ivy webui
integration should sit on top of the compiler and runtime packages.

## Front-end Technology Pivot

Do not continue expanding the Chevrotain prototype into a Go parser. The
browser target still requires a TypeScript implementation, but the implementation
strategy should now follow the Go standard front-end architecture directly:

- scanner/token stream modeled on `go/scanner` and `go/token`
- parser and AST modeled on `go/parser` and `go/ast`
- resolver/typechecker modeled on `go/types`, including interfaces, method
  sets, constants, generics, channels, `go`, `select`, `recover`, and package
  initialization semantics needed by normal Go source files
- package/export metadata sufficient for browser and Node package resolution
- Go compatibility validation after parsing/typechecking, with diagnostics only
  for runtime/environment limitations such as unsupported cgo/native adapters
- spreadsheet syntax handled as deliberate scanner/parser extensions

The rationale for the pivot is semantic correctness. A CST-only parser or an
off-the-shelf JavaScript grammar is not enough for Go-junior because selectors,
method expressions/calls, conversions, type assertions, interfaces, constants,
imports, package scopes, and package export data all require type information.
Continuing to grow a Go-like Chevrotain grammar would recreate the Go compiler
front end one ambiguity at a time.

Spreadsheet syntax should be integrated into the scanner/parser rather than
rewritten as source text:

- `sheet.A1` is a sheet-cell selector candidate.
- `sheet.$A$1`, `sheet.A$1`, and `sheet.$A1` are sheet-cell selectors with
  explicit absolute row/column flags.
- `sheet.A1:B10` and `sheet.$A$1:$B$10` are range expressions.
- Cross-sheet forms such as `Budget.A1` and `Budget.$A$1:$B$10` use the same
  AST shape with a different namespace.
- The resolver decides whether the selector namespace is a sheet namespace, an
  imported package, or an error.

The existing Chevrotain-based code is a prototype and may remain temporarily as
a runtime/REPL bootstrap. New language front-end work should go into the
standard-Go-style TypeScript front end and migrate existing parser tests toward
that implementation.

## Full Go and CPS Runtime Pivot

The project now needs enough full Go support to parse, typecheck, run, and test
existing Go source packages in the CLI and browser worker. The old
"single-threaded subset" strategy is no longer sufficient. The implementation
must support the Go language surface first, then layer spreadsheet semantics and
capability policy on top.

Required language additions:

- `go` statements and goroutine scheduling
- channel types, send statements, receive expressions, two-value receives,
  `close`, closed-channel behavior, nil-channel blocking semantics, buffered
  and unbuffered channels
- `select`, including default cases, receive cases, send cases, closed-channel
  behavior, nil-channel disabling, fairness policy, and deterministic test mode
- `defer`, `panic`, and `recover` across goroutine and CPS suspension
- labels and `goto`, including Go's goto restrictions
- generics: type parameter lists, constraints, type sets, instantiation,
  inference, method sets involving type parameters, and export metadata
- full package initialization order: package variable initialization,
  `init` functions, import graph order, and panic behavior during init
- `unsafe` package surface sufficient for typechecking normal source; runtime
  support is phased and may be restricted by target capability policy

The runtime backend must be a compiler-owned CPS/state-machine transform over
typed Go-junior AST/IR, not a TypeScript AST transform. TypeScript compiler API
or ts-morph may be studied for implementation patterns, but Go-junior's
transform must preserve Go-junior source filenames, spans, types, package
symbols, and spreadsheet dependency metadata.

CPS lowering requirements:

- Every potentially suspending operation lowers to an explicit state transition:
  channel send, channel receive, select, goroutine yield, async host call,
  worker/package RPC, and future asynchronous sheet/graph capabilities.
- Non-suspending code may remain direct JS inside a state body for speed.
- Function calls are annotated as direct or may-suspend from type/effect data.
- Defers are represented in explicit runtime frames and run correctly on
  return, panic, and goroutine exit.
- `recover` observes the innermost active deferred call exactly where Go permits
  it.
- Named returns, multiple returns, panics, gotos, labeled break/continue, and
  loops lower to explicit state-machine control edges.
- Each state and emitted operation carries source span and static type metadata
  for diagnostics and runtime stack traces.
- The scheduler owns goroutine queues, channel wait queues, timers/future async
  waits, panic propagation, deadlock detection, fuel/time budgets, and
  deterministic test hooks.
- Browser and Node runtimes share the same scheduler semantics. Node runtime
  tests and the Go `gojr` CLI can run the scheduler without a browser.

Initial concurrency runtime semantics:

- Goroutines are cooperative tasks scheduled by the Go-junior runtime.
- The default production scheduler should be deterministic for a fixed
  workbook/runtime seed. A separate randomized/fairness stress mode should be
  available in tests.
- Unbuffered channels pair one sender with one receiver.
- Buffered channels maintain FIFO element order.
- Sends to closed channels panic.
- Receives from closed drained channels return zero value and `ok == false`.
- `close(nil)` and closing an already closed channel panic.
- Send/receive on nil channels block forever unless disabled in `select`.
- `select` with no ready cases blocks; with default it runs default
  immediately; with multiple ready cases it chooses via the scheduler policy.
- Runtime deadlock detection reports a structured Go-junior deadlock diagnostic
  when all goroutines are blocked and no host/worker wakeup can occur.

Implementation strategy:

1. Keep the current direct interpreter/backend for non-concurrent smoke tests
   while building the typed CPS IR in parallel.
2. Add a typed effect pass that marks expressions/statements/functions as
   direct, may-panic, may-defer, may-suspend, package-state, diagnostic-effect,
   sheet-effect, graph-effect, or UI-effect.
3. Lower only functions containing `go`, channel operations, `select`,
   `recover`, or may-suspend calls to CPS at first.
4. Once stable, lower all package/formula code through the same state-machine
   backend so one runtime path handles defers, panics, channels, and stack
   traces.
5. Preserve the existing Node REPL as the fastest manual test surface. The REPL
   must parse/typecheck/evaluate with source filename `gojr-repl.go`.

Test strategy for the pivot:

- Start with Go distribution language tests that do not require native cgo or
  target-specific OS behavior.
- Import `/usr/local/go1.26.4/test` fixtures incrementally, tagging expected
  unsupported environment features separately from language failures.
- Add Go-junior-specific spreadsheet/reference tests on top of the Go language
  fixtures.
- Add scheduler model tests for every channel/select edge case before browser
  integration.
- Add randomized scheduler stress tests under a deterministic seed so failures
  are reproducible.

Implementation sequencing should be CLI-first:

1. Build the shared language/compiler/recalculation core.
2. Build the Node runtime and the Go `gojr` CLI backed by embedded Node/V8.
3. Use CLI tests and manual bash sessions to refine syntax, package calls,
   dynamic dependencies, circular-reference behavior, and graph-value semantics.
4. Add the browser worker adapter after the language semantics are stable.
5. Integrate the browser worker into the Ivy webui spreadsheet pane.

Suggested frontend layout:

```
goivy/webui/frontend/src/gojunior/
  ast.ts
  diagnostics.ts
  front/token.ts
  front/scanner.ts
  front/ast.ts
  front/parser.ts
  front/resolver.ts
  front/checker.ts
  front/exportData.ts
  front/compatibility.ts
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
sheet.A1 + sheet.B1 * 2
```

2. Function body form:

```go
x := sheet.A1 + sheet.B1
if x > 10 {
    return math.Sqrt(x)
}
return x * 2
```

Both forms may begin with Go-style imports, including grouped imports and
import aliases:

```go
import (
    "fmt"
    stats "workbook/stats"
)

fmt.Printf("A1=%v\n", sheet.A1)
return stats.Mean(sheet.A1:A10)
```

The compiler should normalize both into an implicit function:

```go
func formula(ctx Context) T {
    ...
}
```

The implicit return type `T` is inferred from the formula body and then boxed
into the spreadsheet runtime value representation. Ordinary formulas should not
need explicit `number(...)` casts for statically known numeric cells.

Go-junior functions live in spreadsheet cells. Cell snippets should not require
users to write package declarations, but they can declare their own imports and
import aliases using Go syntax. Cell imports resolve through the same Go
package resolver used by source package units.

### Initial Types

Use a Go-compatible static type system, extended with spreadsheet namespaces
and values:

- `bool`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr`
- `byte` as an alias for `uint8`
- `rune` as an alias for `int32`
- `string`
- `int64`; untyped integer constants default to `int64`
- `float32`
- `float64`; untyped floating-point constants default to `float64`
- `complex64` and `complex128`; imaginary literals are supported and untyped
  complex constants default to `complex128` when a concrete type is required
- arrays, for example `[3]float64`
- slices, for example `[]float64`
- maps, for example `map[string]int64`
- structs and named struct types
- pointers, for example `*Point`
- interfaces and method sets
- type parameters, constraints, instantiated named types/functions, and type
  inference metadata
- function types, function literals, and closures
- channel types, including directional channels
- tuple-like multiple return values
- `error`, the single Go-style error type
- `Value`, an internal spreadsheet value union for dynamic/unknown host
  boundaries, not a type users should need in ordinary formulas

There should be no Go-junior `number` type. Numeric lowering may still use
JavaScript's numeric representation internally where appropriate, but the
language-level types should be `int64` and `float64`.

Before source package compilation work is treated as complete, Go-junior should
complete a full Go language compatibility pass. Ordinary Go library code should
not be blocked by avoidable syntax, typechecker, package initialization,
concurrency, or runtime gaps. Required compatibility work includes:

- explicit conversions such as `int(x)`, `float64(i)`, `string(b)`, named type
  conversions, and conversion diagnostics for unsupported cases
- the full Go operator surface: `|`, `^`, binary `&`, `&^`, `<<`,
  `>>`, unary `^`, and compound assignments such as `+=`, `-=`, `*=`, `/=`,
  `%=`, `&=`, `|=`, `^=`, `&^=`, `<<=`, and `>>=`
- Go numeric literal forms: binary, octal, hexadecimal, underscores,
  hexadecimal floating-point literals, rune literals, imaginary literals, and
  the standard Go string/rune escape forms
- the predeclared built-ins `new`, `delete`, `copy`, `clear`, `min`, `max`,
  `complex`, `real`, `imag`, `close`, `recover`, plus `print` and `println`
  for Go compatibility
- Go-style `if init; condition {}` and validation of Go control-flow rules,
  including short-declaration redeclaration, `for` post-statement restrictions,
  `fallthrough` placement, and `goto` restrictions
- methods on any named non-pointer, non-interface type, not only named struct
  types
- embedded struct fields, promoted fields/methods, interface embedding, and
  struct tags
- two-value type assertions, for example `v, ok := x.(T)`
- array and struct comparability, map-key comparability enforcement, and Go's
  special nil-comparison rules for slices, maps, functions, pointers, and
  interfaces
- address-of composite literals, for example `&Point{X: 1}`
- three-index slicing with capacity semantics
- blank imports and dot imports
- package `init` functions for Go source packages
- current Go range forms over integers and iterator functions
- channel sends, receives, two-value receives, `go`, `select`, `close`, and
  `recover`
- generic type/function declarations, constraint interfaces/type sets,
  instantiation, and type inference
- package initialization order, including package variable dependencies and
  multiple `init` functions

Go-junior intentionally keeps one spreadsheet ergonomics deviation from Go:
declared but uninitialized maps may auto-initialize on first assignment or map
literal-like use. This avoids a common spreadsheet-scripting annoyance while
the type system still records `map[K]V` and map reads still produce Go-like
zero values plus optional presence booleans.

Static spreadsheet references should be strongly typed. The compiler should
receive a workbook/sheet type environment that maps literal cells, formula
cells, ranges, named regions, and sheet namespaces to declared or inferred
types. `sheet.A1` should have the type recorded for `A1`, such as `int64`,
`float64`, `string`, or `bool`; it should not default to `Value` merely because
it came from the spreadsheet. `sheet.A1:B10` is a spreadsheet range expression
with internal element-type metadata. It can be ranged over directly, or
contextually converted to an ordinary Go-junior slice such as `[]float64` when
all cells in the range share a compatible element type.

`Value` remains useful at dynamic and integration boundaries: dynamic
`Namespace.Cell(addr)` calls whose address cannot be statically constrained,
mixed/untyped ranges, host APIs that deliberately traffic in spreadsheet
values, and runtime error propagation. The common static formula path should be
typed enough that users can write `sheet.A1 + sheet.B1` directly.

### Spreadsheet Values

Runtime cells should represent both a value and a declared or inferred static
cell type. The runtime value domain includes:

- blank
- bool
- int64
- float64
- string
- error
- Go-junior function

The type environment should be updated as formulas are edited and typechecked:

- literal cells infer their type from their literal value unless the user or
  schema declares a stronger type
- formula cells infer their output type from the formula body unless the user
  or schema declares an expected result type
- Go-junior function cells store editable Go-junior source plus a compiled
  callable function value
- declared cell/range types are checked against literal values and formula
  return types
- blank cells can be typed by schema or remain untyped until written
- runtime errors are values but do not erase the cell's declared/inferred type

Errors should propagate predictably. A cell formula that fails typechecking
should not compile. A cell formula that fails at runtime should produce a
spreadsheet error value with a stable code and message.

### Go-junior Function Cells

One spreadsheet cell type should be "Go-junior function". A function cell
contains editable Go-junior source and compiles to a callable function value
that other formulas can reference and call.

Example:

```go
// Cell Lib.Double, type: Go-junior function
func Double(x float64) float64 {
    return x * 2
}
```

```go
// Another cell
return Lib.Double(sheet.A1)
```

Function-cell rules:

- The cell stores raw editable source as authoritative workbook state.
- The cell's runtime value is a callable function pointer/closure produced by
  compiling that source.
- The type environment records the function signature so callers can typecheck
  before evaluation.
- Function cells can declare their own imports and aliases using the same
  cell-import rules as formula cells.
- Function cells can capture only explicit lexical state from their source and
  declared imports, not arbitrary workbook globals.
- Changing a function cell's source invalidates callers that depend on its
  signature or implementation.
- Function cells participate in dependency tracking like formulas: calls to the
  function record a dependency on the function cell, and any cell/range reads
  performed by the function are observed during evaluation.
- Persisted workbooks store the function source and declared cell type, not the
  compiled function object.

Initial error codes:

- `#ERROR!` for general runtime failures
- `#TYPE!` for runtime type conversion failures
- `#DIV/0!` for division by zero
- `#NAME?` for unknown symbols in formulas loaded from persisted state
- `#REF!` for invalid cell or range references, including invalid dynamic
  references produced at runtime
- `#PANIC!` for an unrecovered `panic` value escaping a formula, function cell,
  or package call
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
- OPFS package artifact cache, with IndexedDB only for optional indexes
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

- literals: integer, floating-point, string, boolean, array, slice, map, and
  struct literals
- identifiers
- unary operators: `+`, `-`, `!`
- binary arithmetic: `+`, `-`, `*`, `/`, `%`
- comparisons: `==`, `!=`, `<`, `<=`, `>`, `>=`
- boolean operators: `&&`, `||`
- parentheses
- function calls
- function literals and closures
- channel receive expressions: `<-ch`
- address-of and dereference expressions: `&x` and `*p`
- indexing and slicing for arrays, slices, maps, strings where supported, and
  spreadsheet ranges where appropriate
- field selectors and method calls on structs
- interface method calls
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

Deliberately excluded:

- cgo-only syntax/build behavior in browser execution
- ambient JavaScript/global access

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

### Cell Imports

Spreadsheet-cell Go-junior functions may declare imports before the expression
or function body:

```go
import "fmt"
```

```go
import (
    f "fmt"
    stats "workbook/stats"
)
```

Rules:

- Cell imports are scoped to the containing cell function.
- Cell imports use Go import syntax, including grouped imports and explicit
  aliases.
- The imported package name or alias enters the same selector namespace as
  sheet namespaces and prebound packages.
- `sheet` remains reserved; an import alias named `sheet` is rejected.
- A cell import alias that collides with a sheet namespace or another package
  namespace is a compile-time diagnostic.
- If no alias is provided, the default selector name is the imported package's
  declared package name, matching Go.
- Imported packages resolve through the built-in, workbook, browser-cache, and
  trusted host package providers used by source packages.

### Statements

Support:

- short variable declarations: `x := expr`
- `var` declarations, including grouped declarations
- Go-style `const` declarations, including grouped declarations, implicit
  expression repetition, and `iota`
- `type` declarations for structs, interfaces, named arrays, named slices, and
  named maps
- assignments to locals: `x = expr`
- assignments to exported mutable package variables: `pkg.VariableName = expr`
- `if`, `else if`, `else`
- expression `switch` statements
- type `switch` statements when interfaces are available
- `return expr`
- `return expr, expr` for multiple return values
- naked `return` only inside functions with named return values
- `defer call(...)`
- `go f(...)`
- channel send statements: `ch <- value`
- `select` statements with send, receive, assignment, and default cases
- expression statements only for calls whose return value can be ignored
- full Go-style `for` loops: `for {}`, `for cond {}`,
  `for init; cond; post {}`
- Go-style `for range` loops over arrays, slices, maps, strings where
  supported, and spreadsheet ranges
- `break`, `continue`, and `fallthrough`, matching Go switch and loop
  semantics
- labels and `goto`, matching Go restrictions

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

### Constants

Go-junior should support Go-style `const` declarations:

```go
const Pi = 3.14159

const (
    A = iota
    B
    C
)
```

Constants are compile-time values. Grouped declarations, implicit expression
repetition, and `iota` should match Go's useful semantics. Untyped integer
constants default to `int64`; untyped floating-point constants default to
`float64` when a concrete type is required.

### Predeclared Built-ins

Go-junior should include a small predeclared built-in surface. Required early
built-ins include:

```go
func panic(v interface{})

func panicOn(err error) {
    if err != nil {
        panic(err)
    }
}
```

`panic` and `recover` should follow Go's essential unwinding semantics:
deferred calls run in LIFO order as the stack unwinds, and `recover` succeeds
only when called directly by a deferred function in the panicking goroutine. A
panic that escapes the top-level formula, function cell, package call, or
goroutine root becomes a `#PANIC!` spreadsheet error or a structured goroutine
panic diagnostic, depending on execution context. The panic diagnostic should
include the panic value, source span if available, goroutine ID, and a
Go-junior stack trace.

`panicOn(err)` is a required predeclared helper because `(T, error)` APIs are
common and spreadsheet formulas need a compact way to fail fast during
interactive exploration.

The rest of the Go built-in surface should be added before source package
compilation is treated as complete: `append`, `cap`, `clear`, `close`,
`complex`, `copy`, `delete`, `imag`, `len`, `make`, `max`, `min`, `new`,
`panic`, `print`, `println`, `real`, and `recover`.

### Functions

Stage v1 can compile one implicit cell function. Go-junior functions should
support local helper functions, source package functions, function literals,
closures, methods on structs, single returns, multiple returns, and named
returns. Variadic parameters are required so built-ins such as `fmt.Printf`
can use Go-style `args ...interface{}` signatures.

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
return clamp(sheet.A1, 0, 100)
```

Multiple returns:

```go
func DivMod(x int64, y int64) (int64, int64) {
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
- named return values are assigned before defers run; closures can observe or
  mutate captured named return values according to the Go-junior closure model
- deferred closures and deferred ordinary function/method calls are supported

Closures are available. The implementation must define capture semantics
clearly and test mutations through captured locals, including captured named
return values used with `defer`.

Variadic functions should follow Go's call-shape rules closely enough for
`fmt.Printf(format, args...)` style APIs and ordinary variadic calls.

### Goroutines, Channels, Select, and Recover

Full Go concurrency is required. Go-junior should support goroutines, channels,
`select`, `close`, and `recover` using a deterministic cooperative scheduler in
the TypeScript runtime.

Required syntax and type support:

```go
ch := make(chan int)
done := make(chan struct{})

go func() {
    defer close(done)
    ch <- 3
}()

select {
case v := <-ch:
    fmt.Printf("value=%v\n", v)
case <-done:
    return
default:
    fmt.Printf("not ready\n")
}
```

Semantics:

- `make(chan T)` creates an unbuffered channel; `make(chan T, n)` creates a
  buffered channel with capacity `n`.
- Directional channel types (`<-chan T`, `chan<- T`) are supported and enforced
  by the typechecker.
- Send to a nil channel blocks; receive from a nil channel blocks.
- Send to a closed channel panics.
- Receive from a closed and drained channel returns the element zero value; the
  two-value receive returns `ok == false`.
- `close` wakes blocked receivers and future receives observe closed-channel
  semantics.
- Closing nil or already closed channels panics.
- `select` disables nil-channel cases, runs `default` immediately if no case is
  ready, and otherwise chooses one ready case through the scheduler's policy.
- The default scheduler policy should be deterministic for reproducible
  spreadsheet recalculation and CLI tests. A seeded randomized/fairness mode
  should exist for stress testing.
- Deadlock detection reports a structured diagnostic when all goroutines are
  blocked and no host or worker wakeup can occur.
- Goroutine roots must report panics with goroutine stack traces and source
  spans.
- `recover` works only during deferred-call unwinding in the same goroutine.

Implementation notes:

- Channel operations and `select` are may-suspend operations in the typed effect
  pass.
- Functions containing may-suspend operations, or calling may-suspend
  functions, must lower through CPS/state-machine IR.
- Defers live in explicit runtime frames so they survive suspension.
- The runtime scheduler must be shared between Node tests and browser workers.
- Spreadsheet formula contexts may restrict goroutine/channel use by policy
  later, but the language/runtime must support them so source packages can be
  tested and run.

Tests:

- Unbuffered send/receive handoff between two goroutines.
- Buffered channel FIFO behavior and capacity blocking.
- Two-value receive from open, closed non-drained, and closed drained channels.
- Send to closed channel panics and runs defers.
- Close nil and double close panic.
- Nil channel send/receive block and are disabled inside `select`.
- `select` with default runs default when no case is ready.
- `select` receives from ready closed channels.
- Multiple ready `select` cases follow deterministic scheduler policy.
- Deadlock is detected with useful goroutine stack diagnostics.
- `recover` succeeds in a deferred function and fails outside that context.
- Goroutine panic reports goroutine ID, source file/line, and stack.
- Channel element types and directional channel assignments are enforced by the
  typechecker.

### Structs, Interfaces, Methods, and Closures

Go-junior should support struct and interface declarations, struct literals,
field selection, methods on named struct types, interface method-set checking,
and interface method calls:

```go
type Point struct {
    X float64
    Y float64
}

func (p Point) Len2() float64 {
    return p.X*p.X + p.Y*p.Y
}

func (p *Point) Scale(k float64) {
    p.X = p.X * k
    p.Y = p.Y * k
}

type HasLen2 interface {
    Len2() float64
}
```

Struct values should have deterministic value semantics. Interface satisfaction
should be structural by method set, following the Go model. Methods on structs
and pointer receivers are required. Go-junior should support pointer types,
address-of expressions, dereference expressions, field/method selection through
pointers, and Go-like method-set rules for `T` and `*T`. Nil pointer values
should be representable and nil dereference should produce a runtime
spreadsheet error rather than crashing the worker.

Method and field selection should use Go's single selector syntax. Values and
pointers both use `x.Method()` and `x.Field`; Go-junior must not introduce a
C/C++-style `x->Method()` or `x->Field` syntax. The typechecker should perform
the same automatic dereference and address-taking for method calls that Go
permits.

Go-style type switches on interface values are required:

```go
func Describe(x interface{}) string {
    switch v := x.(type) {
    case string:
        return "string: " + v
    case float64:
        return fmt.Sprintf("float64: %v", v)
    case HasLen2:
        return fmt.Sprintf("len2: %v", v.Len2())
    default:
        return "unknown"
    }
}
```

Switches on interface-typed expressions are also required, subject to
Go-junior's comparable-type rules.

Function literals and closures are required:

```go
scale := 2.0
double := func(x float64) float64 {
    return x * scale
}
```

Captured locals should have clear mutation semantics and must work with
`defer` and named returns.

### Arrays, Slices, Maps, and Indexing

Arrays, slices, indexing, slicing, and composite literals are required:

```go
xs := []float64{1, 2, 3}
ys := [3]int64{1, 2, 3}
return xs[0] + float64(ys[1])
```

Maps are required with Go-style type and literal syntax:

```go
counts := map[string]int64{"a": 1, "b": 2}
counts["c"] = 3
```

Map iteration differs intentionally from Go: Go-junior maps iterate in
deterministic insertion order like Python `dict`. Updating an existing key does
not move it; deleting and reinserting a key appends it at the new insertion
position.

Map keys must be Go-comparable. Slices, maps, functions, and other
non-comparable values are rejected as map keys, except for Go's explicit nil
comparison rules where applicable. Declared zero-value maps may auto-initialize
on first assignment for spreadsheet ergonomics, but `delete`, `clear`,
two-value lookup, range order, and missing-key zero values should otherwise be
Go-like.

### Source Packages

Go-junior should support Go-compatible source packages so formulas can call
reusable libraries without those libraries being built into the core host API.
These packages are parsed and typechecked as Go source with spreadsheet-aware
extensions only where explicitly allowed by formula/cell context.

Supported package shape:

```go
package stats

func Mean(xs []float64) float64 {
    total := 0.0
    n := 0
    for _, v := range xs {
        total = total + v
        n = n + 1
    }
    return total / float64(n)
}
```

Package support should include:

- `package name` declarations for package source units.
- explicit imports by package path, for example `import "stats"`.
- exported functions, constants, mutable package variables, types, structs,
  interfaces, methods, arrays, slices, maps, and closures.
- package-private helpers.
- package-level `const` declarations, including grouped declarations and
  `iota`.
- package-level `var` declarations with deterministic initializers.
- package-level `type` declarations for structs, interfaces, arrays, slices,
  maps, pointers, and function types.
- methods on named struct types, pointer receiver methods, plus interface
  method-set checking.
- Go-like mutable package variable reads and writes inside package functions.
- a simplified deterministic package initialization order defined by the
  Go-junior package graph, not the full Go specification.
- topological compilation of package import graphs.
- rejection of import cycles.
- support for `defer`, multiple returns, and named returns inside package
  functions.
- support for `go`, channels, channel operations, `select`, `close`, `panic`,
  and `recover` inside package functions.
- support for generics, constraints, type sets, instantiation, and inference.
- support for closures, `switch`, full `for`, `for range`, array/slice/map
  literals, indexing, and struct literals inside package functions.
- rejection only of environment features that cannot run in the selected target,
  such as cgo/native execution in browser, with explicit diagnostics and host
  adapter alternatives.
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

Full arbitrary Go modules are still a packaging and environment project, but
the language implementation target is full Go. A package can be used as source
when its imports and environment requirements are available to the Go-junior
runtime. Packages requiring cgo, native syscalls, OS-specific services, or
unimplemented standard-library internals should be exposed through typed
host/WASM bindings until those runtime services exist.

### Built-in fmt Package

`fmt` should be one of the first built-in Go-junior packages implemented
because printing is the diagnostic keystone for CLI usability, spreadsheet-cell
debugging, and language development.

Initial package surface:

```go
package fmt

func Printf(format string, args ...interface{}) (int64, error)
func Sprintf(format string, args ...interface{}) string
func Println(args ...interface{}) (int64, error)
```

`fmt.Printf` writes to the runtime diagnostic/log sink rather than directly to
browser globals or Node process globals. In the Go `gojr` CLI, that sink can
render to stdout or structured JSON output. In the browser worker, it should
be returned with the cell evaluation result and displayed in the UI diagnostics
or trace panel.

The implementation can lean on trusted JavaScript reflection internally to
format Go-junior values, structs, maps, slices, interfaces, errors, spreadsheet
values, and host values. That reflection is implementation authority of the
built-in package, not ambient authority exposed to user code.

Diagnostic output rules:

- `fmt.Printf` is a `diagnostic-effect`, not a sheet/UI mutation.
- Output is tagged with workbook, sheet, cell, evaluation generation, and
  recalculation pass.
- Output is bounded by per-evaluation byte and call-count limits.
- Recalculation may rerun a formula and therefore rerun `fmt.Printf`; the UI
  should make generation/pass information visible enough to avoid confusion.
- `fmt.Sprintf` is pure and returns a string.
- Formatting of maps must use Go-junior's deterministic insertion order.

### Loops and Ordered Maps

Go-junior should support the ordinary Go `for` family:

```go
for i := 0; i < 10; i = i + 1 {
    sum = sum + i
}
```

```go
for i < limit {
    i = i + 1
}
```

```go
for {
    break
}
```

Range loops over finite spreadsheet ranges, arrays, slices, maps, and strings
where supported should be available:

```go
for _, v := range sheet.A1:A10 {
    sum = sum + v
}
```

The typechecker should infer `v` from the static range/slice/array/map element
type, for example `float64` for a spreadsheet range contextually typed as
`[]float64`.

Maps use Go-style `map[K]V` type syntax and map literal syntax, but their
iteration order is deterministic insertion order, matching Python `dict`
behavior rather than Go's intentionally randomized map iteration. Updating an
existing key does not move it; deleting and reinserting a key appends it at the
new insertion position.

The runtime must enforce a fuel counter for all loops, including `for {}` and
data-dependent loops. Loop syntax is full Go-style; loop execution is still
bounded by runtime fuel so spreadsheet recalculation cannot hang indefinitely.

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

Host calls may be synchronous or asynchronous. Asynchronous host calls are
may-suspend operations and must use the same CPS/runtime scheduler path as
channel operations and goroutine blocking. Pure synchronous calculation helpers
remain preferable for ordinary formulas, but the compiler/runtime must be able
to suspend and resume a goroutine around host, worker, package, or WASM calls
from the start of the CPS backend.

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
- `diagnostic-effect`: writes to the evaluation diagnostic/log sink, such as
  `fmt.Printf`. This is allowed in formula context because diagnostics are a
  first-class development workflow, but output should be tagged with cell ID,
  evaluation generation, and recalculation pass so repeated formula evaluation
  is understandable and bounded.
- `sheet-effect`: explicit mutations to spreadsheet state, such as setting a
  cell, adding a sheet, or creating a named output object.
- `graph-effect`: graph production and rendering requests, such as emitting
  nodes/edges, chart specs, DOT, Vega-lite specs, Cytoscape elements, or an Ivy
  concept/ARG visualization payload.
- `ui-effect`: explicit UI requests, such as opening a panel, displaying a
  diagnostic, or focusing a result.

Formula recalculation should run `pure`, approved deterministic `wasm`,
approved formula-safe `dynamic-js`, approved `diagnostic-effect`, and
explicitly allowed `package-state` exports. A formula that uses `package-state`
is not pure: it should not be common-subexpression cached, it should be
evaluated in the scheduler's stable order, and its package state lifecycle must
be deterministic. Browser and workbook mutation capabilities should require an
action context, for example a user-triggered command, button, menu action, or
explicitly marked command cell. This prevents normal dependency recomputation
from repeatedly mutating sheets or redrawing browser state while still allowing
formula-local diagnostic printing.

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
  diag: {
    Printf: fn(["string", "...interface{}"], "int64,error", {
      capability: "diagnostic-effect",
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

The runtime should be able to JIT compile Go source packages
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
  -> persist compiled package artifact in the target artifact cache
```

The package artifact cache has two durable targets:

- Node/CLI: `~/go/pkg/gojr_js/` by default. This mirrors Go's compiled package
  cache layout under directories such as `~/go/pkg/darwin_amd64/`, except the
  package artifacts are JavaScript source/module files with a `.js` suffix
  instead of `.a` archives.
- Browser: OPFS (Origin Private File System). OPFS is the durable artifact store
  for generated `.js` package artifacts and metadata inside the browser.

IndexedDB may still be used in the browser for small indexes, package manifests,
or queryable metadata, but the generated package artifacts themselves should
live in OPFS. Worker memory remains the first-level cache in both browser and
Node runtimes.

CLI artifact examples:

```text
~/go/pkg/gojr_js/golang.org/x/crypto/scrypt.js
~/go/pkg/gojr_js/github.com/user/project/pkg/name.js
```

Cache keys must include:

- package path and package version/name
- content hashes of all package source files
- transitive dependency cache keys
- compiler version
- target backend, for example `js-source` or `wasm-stencil`
- package ABI version
- artifact layout version
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

CLI `gojr build` artifact layout:

- `gojr build <pkg>` compiles the package and any stale dependencies into
  `~/go/pkg/gojr_js/` by default.
- Tests and explicit developer workflows may override the root with a CLI flag
  such as `-pkgdir`, an environment variable, or an injected runtime option.
- The output path is derived from the package import path, preserving the
  package namespace directory layout and ending in `.js`.
- A package artifact may store metadata in an embedded header, a sibling
  `.gojr.json` sidecar, or a cache manifest, but the stable user-facing artifact
  is the import-path-derived `.js` file.
- Rebuilds are required when source hashes, dependency artifact hashes,
  compiler version, package ABI version, host spec, capability policy, target
  backend, or artifact layout version change.
- Writes must be atomic: write to a temp artifact in the target directory,
  fsync where appropriate, rename into place, then update sidecar/manifest
  metadata after the `.js` artifact succeeds.
- Compiled artifacts must not persist mutable package variable values. They
  contain generated code, export metadata, type metadata, source-span metadata,
  and dependency metadata only.
- The build report should be machine-readable and include built/skipped
  packages, artifact paths, cache keys, dependency edges, and diagnostics.

Package compilation should run in the browser worker for the webui and in the
Node runtime or a Node `worker_threads` worker for command-line use. Formula
compilation may request a package by path; the runtime resolves, compiles,
caches, and links it before compiling the formula that imports or references
it. Stale package compilations must be ignored by generation token just like
stale formula evaluations.

The JavaScript compiler/runtime owns the build API, for example
`gojr.buildPackage(...)` or `gojr.buildPackages(...)`. The Go `gojr` binary
invokes this JavaScript method through embedded Node/V8 for `gojr build`,
`.load`, `.source`, and `.test` workflows. The Go binary should marshal CLI
arguments, package roots, cache roots, stdin/stdout, and diagnostics across the
embedding boundary; it should not reimplement package parsing, typechecking,
code generation, or artifact-cache invalidation in Go.

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

This model allows existing Go source libraries to be loaded when their imports
and runtime environment requirements are available to Go-junior, while keeping
cgo/native or otherwise target-specific libraries available through typed
host/WASM bindings until equivalent runtime services exist.

## CPS State-machine and Source-level Copy-and-Patch JavaScript Emission

Use a compiler-owned CPS/state-machine backend. JavaScript source stencils are
still useful for emitting individual state bodies and runtime helper calls, but
the generated program is not merely a direct JavaScript function when it may
suspend. A stencil is a compiler-owned JavaScript text fragment with typed
holes. For example:

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
- May-suspend operations lower to explicit scheduler calls and continuation
  states.
- Direct non-suspending runs may be optimized to straight-line JavaScript, but
  the observable runtime behavior must match the state-machine path.
- Generated runtime frames include source filenames, spans, static types,
  function names, goroutine IDs, defer stacks, and panic/recover state.

Example generated JavaScript:

```js
"use strict";
return function _gj_entry(runtime, ctx, host) {
  return runtime.start(function _gj_step(frame) {
    switch (frame.pc) {
      case 0:
        frame.v0 = ctx.cell("sheet", "A1", "float64");
        frame.v1 = ctx.cell("sheet", "B1", "float64");
        frame.v2 = frame.v0 + frame.v1;
        if (frame.v2 > 10) { frame.pc = 10; return runtime.cont(frame); }
        return runtime.return(frame, frame.v2 * 2);
      case 10:
        return runtime.return(frame, host.math.Sqrt(frame.v2));
    }
  });
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
- run the cooperative goroutine scheduler
- manage channel queues, select waits, close wakeups, async host wakeups, and
  goroutine lifecycle
- detect goroutine deadlock and convert it into structured diagnostics
- preserve source filename/line/type metadata in runtime frames and stack
  traces
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
syntax, Go compatibility scope, spreadsheet extensions, types, errors, host
binding shape, scheduler behavior, and runtime limits.

Implementation tasks:

- Create fixture format for compiler tests.
- Define diagnostic object shape with code, message, severity, source span,
  and optional help text.
- Define source span conventions: line, column, offset, and length.
- Define workbook/sheet type environment shape for cells, ranges, named
  regions, sheet namespaces, and formula result types.

Tests:

- Golden fixture loader reads valid and invalid cases.
- Diagnostic spans round-trip line/column/offset accurately.
- Spec examples are copied into parser/typechecker tests so docs cannot drift.
- A "Go concurrency" fixture list asserts that `go`, `select`, channels,
  `close`, send/receive, and `recover` parse, typecheck, and execute according
  to the scheduler model.
- An "environment-limited Go feature" fixture list asserts that cgo/native-only
  packages and target-unavailable `unsafe` operations produce explicit
  environment diagnostics, not parser/typechecker gaps.
- A "panic" fixture list asserts that `panic(value)` parses, typechecks, runs
  defers during unwind, and becomes `#PANIC!` when unrecovered.
- A "recover" fixture list asserts that `recover()` works only inside a
  deferred call during panic unwinding and returns nil elsewhere.
- A "panicOn" fixture list asserts that `panicOn(nil)` returns normally and
  `panicOn(err)` becomes `#PANIC!`.
- A "supported Go-junior Go idiom" fixture list asserts that `defer`, multiple
  returns, named returns, and source package declarations/imports are accepted
  in the contexts where the language supports them.
- A "cell import" fixture list asserts that spreadsheet-cell functions accept
  `import "fmt"`, grouped imports, and Go-style import aliases such as
  `import ( f "fmt" )`.
- A "fmt diagnostic" fixture list asserts that `fmt.Printf` output is captured
  as evaluation diagnostics in Node and browser-style runtimes.
- A "selector versus spreadsheet reference" fixture list asserts that bare
  `A1`, `pkg.A1`, `sheet.A1`, `Data.A1`, and `Data.A1:B10` are parsed and
  resolved according to namespace rules.
- Typed-cell fixtures assert that `sheet.A1 + sheet.B1` typechecks when both
  cells are numeric and fails with a span diagnostic when a referenced cell is
  typed as `string` or `bool`.

Acceptance criteria:

- The Go compatibility target and spreadsheet extensions are explicit enough
  that parser and typechecker work can begin.
- Tests can assert both successful normalized output and exact diagnostics.

### Stage 1: TypeScript Go-style Scanner

Implement the Go-junior scanner in TypeScript, modeled on Go's scanner/token
front end rather than on Chevrotain.

Implementation tasks:

- Define token kinds for identifiers, keywords, integer literals,
  floating-point literals, strings, operators, delimiters, comments,
  whitespace, and newlines.
- Include Go-junior keywords for `const`, `type`, `struct`, `interface`,
  `switch`, `case`, `default`, `fallthrough`, `for`, `range`, `break`,
  `continue`, `map`, and `func`.
- Implement Go-like semicolon insertion in the scanner/token stream.
- Define lexer behavior for spreadsheet address fragments such as `A1`, `$A$1`,
  `A$1`, `$A1`, and range separator `:`.
- Treat `A1`-style tokens as both address-like and identifier-like where
  needed, because bare `A1` is a legal ordinary identifier candidate until the
  resolver proves a surrounding selector namespace is a sheet.
- Keep enough newline information to support explicit statement boundaries or a
  small Go-like semicolon insertion/token-normalization pass.
- Decide whether comments and whitespace are skipped or retained for diagnostic
  and formatting metadata.
- Normalize scanner token location data into Go-junior source spans.
- Convert scanner errors into Go-junior diagnostics.
- Tokenize channel receive/send syntax (`<-`) and all operators required by
  full Go. Scanner diagnostics should be reserved for malformed tokens, not
  supported language constructs.

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
- Tokenizes rune literals and preserves their spans.
- Tokenizes channel operator `<-` with source spans.
- Rejects malformed numbers.
- Converts scanner errors into stable Go-junior diagnostic codes.
- Snapshot tests for representative snippets.

Acceptance criteria:

- Lexer never throws for user input; it returns tokens plus diagnostics.
- Every token has a stable span.
- The token vocabulary is close enough to Go's token model that parser and
  typechecker work can follow the Go front-end architecture.

### Stage 2: TypeScript Go-style Parser and AST

Implement a recursive-descent parser and AST in TypeScript, modeled on Go's
parser and AST. The parser owns a Go-compatible syntax shape plus Go-junior's
spreadsheet extensions.

Implementation tasks:

- Define AST node types aligned with Go's AST, plus explicit spreadsheet
  extension nodes.
- Parse with Go-like precedence and statement grammar instead of a CST
  conversion layer.
- Keep parser diagnostics stable and source-spanned.
- Parse expression-form and function-body-form programs.
- Parse optional cell-level import declarations before expression-form or
  function-body-form programs, including grouped imports and explicit aliases.
- Parse statements: declarations, assignments, if/else, switch, full for loops,
  for-range loops, break/continue, returns, and expression statements.
- Parse `defer` call statements.
- Parse `go` statements.
- Parse channel types, directional channel types, send statements, receive
  expressions, and `select` statements.
- Parse `close` and `recover` calls as ordinary predeclared built-in calls.
- Parse multiple return expressions.
- Parse named return function signatures.
- Parse variadic parameter syntax, for example `args ...interface{}`.
- Parse generic type/function declarations, type parameter lists, constraints,
  type sets, instantiations, and generic method receiver forms where Go
  permits them.
- Parse type expressions used in signatures, including arrays, slices, maps,
  pointers, structs, interfaces, named types, and function types.
- Parse `const`, `var`, and `type` declarations, including grouped
  declarations.
- Parse struct type declarations, interface type declarations, and method
  declarations with receivers.
- Parse function literals and closures.
- Parse calls and selector expressions.
- Parse address-of and dereference expressions: `&x` and `*p`.
- Parse indexing and slicing expressions.
- Parse array, slice, map, and struct composite literals.
- Represent operator precedence with Go-style expression parsing.
- Parse special current-sheet and cross-sheet cell/range references.
- Parse ambiguous statement prefixes, such as assignment versus expression
  statement, with explicit lookahead or factored grammar rules.
- Represent `Name.A1` as a selector-or-cell-reference candidate in the AST
  until resolver determines whether `Name` is a sheet namespace or a package
  namespace.
- Represent `Name.A1:B10` as a spreadsheet range candidate because the colon
  form is not ordinary Go selector syntax.
- Preserve AST spans for all nodes from token source ranges.
- Convert parser errors into Go-junior diagnostics.
- Add parser recovery where it improves multi-error reporting without
  compromising correctness.

Tests:

- Parses arithmetic precedence correctly.
- Parses cell-level `import "fmt"` before expression form and function-body
  form.
- Parses grouped cell imports with aliases, for example
  `import ( f "fmt"; stats "workbook/stats" )`.
- Parses nested parentheses.
- Parses unary and binary operations.
- Parses `if`, `else if`, and `else`.
- Parses short declarations and assignments.
- Parses `var`, grouped `var`, `const`, grouped `const`, and `iota` syntax.
- Parses `type` declarations for structs, interfaces, arrays, slices, maps,
  pointers, and function types.
- Parses method declarations with value and pointer receivers.
- Parses address-of and dereference expressions.
- Parses method and field selection through values and pointers using `.`.
- Rejects C/C++-style `->` selector syntax.
- Parses function literals and closures.
- Parses selector assignments like `pkg.VariableName = expr`.
- Parses indexing and slicing expressions.
- Parses array, slice, map, and struct literals.
- Parses expression switches and type switches.
- Parses full `for` loops: `for {}`, `for cond {}`, and
  `for init; cond; post {}`.
- Parses `for range` loops over range expressions.
- Parses `break`, `continue`, and `fallthrough` in legal contexts.
- Rejects `fallthrough` outside a switch case or in the final switch case,
  matching Go semantics.
- Parses return statements.
- Parses multiple return values.
- Parses named return values and naked returns inside named-return functions.
- Parses variadic signatures such as
  `func Printf(format string, args ...interface{}) (int64, error)`.
- Parses slice signatures like `func Mean(xs []float64) float64`.
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
- Parses calls to predeclared `panic(...)` and `panicOn(err)`.
- Parses `go f()`, `select {}`, channel sends, receives, labels, `goto`,
  `recover`, generic type parameters, and generic instantiations.
- Rejects package declarations in cell snippets when the snippet form does not
  allow package units.
- Parses `unsafe` package imports and selectors as ordinary imports/selectors;
  target/runtime policy later decides whether a specific unsafe operation can
  execute.
- Parser diagnostics use stable Go-junior diagnostic codes and source spans.
- Error recovery reports multiple syntax errors in one source where possible.
- AST construction preserves spans for expressions, statements, spreadsheet
  references, and function signatures.
- Golden AST tests for small valid programs.

Acceptance criteria:

- Valid v1 syntax produces a complete AST.
- Invalid syntax and target/environment-unavailable constructs produce stable
  diagnostics without crashing.
- Later compiler stages consume only Go-junior AST nodes, type information, and
  diagnostics.

### Stage 3: AST Normalization

Normalize program forms into one internal function-like representation.

Implementation tasks:

- Expression form becomes `return expr`.
- Function body form becomes an implicit cell function.
- Cell import declarations attach to the implicit cell function compile unit.
- Named returns normalize to explicit return locals.
- Const groups expand implicit expression repetition and `iota` values.
- Switch statements normalize to explicit branch structures while preserving
  Go switch semantics.
- Function literals receive explicit closure capture metadata.
- Insert explicit return requirement checks where needed.
- Normalize syntactic sugar into canonical AST nodes.

Tests:

- Expression source normalizes to one return statement.
- Function body source preserves user statements.
- Cell import declarations are preserved as compile-unit imports.
- Multiple return values normalize to tuple return IR shape.
- Named return values normalize to local return slots.
- Const declarations normalize repeated expressions and `iota` correctly.
- Switch statements normalize without changing case order or default behavior.
- Function literals normalize with explicit capture lists.
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
- Resolve const, var, type, function, method, field, and interface names.
- Resolve cell-level imports and import aliases through the package resolver.
- Resolve closure captures and distinguish captured locals from package state.
- Reject use before declaration unless explicitly allowed.
- Reject duplicate local declarations in the same scope.
- Resolve host namespaces and functions against a typed host spec.
- Resolve the reserved `sheet` pseudo package and conversion helpers.
- Resolve predeclared built-ins, including `panic` and `panicOn`.
- Resolve workbook sheet namespaces as package-like selector namespaces.
- Resolve static cell and range references against the workbook/sheet type
  environment.
- Reject collisions between sheet namespaces and imported/source package
  namespaces unless a future explicit alias mechanism is present.
- Reject local declarations, import aliases, or package names that collide with
  reserved `sheet` in formula scope.
- Generate stable symbol IDs for locals and functions.

Tests:

- Resolves locals in nested `if` blocks.
- Resolves locals and captures inside function literals.
- Resolves const declarations, grouped consts, and `iota`.
- Resolves cell-level imports, grouped imports, and import aliases.
- Resolves `import f "fmt"` so `f.Printf` binds to the built-in `fmt` package.
- Rejects unknown cell imports with source-span diagnostics.
- Resolves predeclared `panic` and `panicOn`.
- Resolves struct fields, method declarations, method calls, and interface
  method names.
- Resolves value receiver and pointer receiver methods into the appropriate
  method sets.
- Resolves array, slice, map, pointer, and function type names.
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
- Resolves static cell/range references to type-environment entries.
- Reports unknown static cell/range references with source-span diagnostics.
- Rejects ambiguous selector namespaces when a sheet namespace and package alias
  share the same name.
- Rejects unknown host namespaces/functions.
- Rejects unknown sheet namespaces in spreadsheet references.
- Rejects local variables named `sheet`.
- Rejects package imports or aliases named `sheet`.
- Rejects cell import aliases that collide with sheet namespaces, package
  namespaces, or other imports.
- Does not resolve against `window`, `document`, `globalThis`, or other browser
  globals.

Acceptance criteria:

- Every identifier use is bound to a known symbol or produces a diagnostic.

### Stage 5: Typechecker

Add static typing for expressions, statements, returns, and calls.

Implementation tasks:

- Define type model and assignability rules.
- Type arithmetic, comparisons, booleans, strings, and returns.
- Type untyped integer constants as `int64` by default and untyped
  floating-point constants as `float64` by default.
- Type arrays, slices, maps, structs, pointers, interfaces, methods, function
  literals, closures, indexing, slicing, address-of, dereference, and composite
  literals.
- Type channel types, directional channels, `make(chan T, cap)`, send
  statements, receive expressions, two-value receives, `select` cases, and
  `close`.
- Type `go` statements and mark launched calls as goroutine roots.
- Type generic declarations, constraints, type sets, instantiations, and
  inference following Go's type parameter model.
- Type Go-style const declarations, including grouped declarations, implicit
  expression repetition, and `iota`.
- Type expression switches, type switches, full `for` loops, and `for range`
  loops.
- Type Go-style type switches over interface values and switch cases involving
  interface-typed expressions.
- Type variadic calls and variadic function declarations.
- Type built-in `fmt` package calls, including `fmt.Printf`,
  `fmt.Sprintf`, and `fmt.Println`.
- Type predeclared `panic(v interface{})` and `panicOn(err error)`.
- Type predeclared `recover()` and its `interface{}` result; runtime legality
  is enforced by defer/panic state.
- Type current-sheet and cross-sheet cell/range references from the
  workbook/sheet type environment.
- Type dynamic `Namespace.Cell` / `Namespace.Range` helpers with contextual
  typing when the expected type is known, otherwise as `Value` or an untyped
  spreadsheet range value.
- Type assignment to locals and exported mutable package variables.
- Type host calls from the host spec.
- Type tuple return values and tuple assignment from multiple-return calls.
- Type Go-junior function cells as callable function values using their
  recorded signatures.
- Type named return values and naked returns.
- Type `defer` statements; deferred expressions must be calls and their return
  values are ignored.
- Decide numeric coercion policy and implement it consistently.
- Type local helper functions when added.

Tests:

- Arithmetic accepts compatible numeric operands.
- Integer literals default to `int64`; floating-point literals default to
  `float64`.
- There is no accepted `number` type annotation or conversion helper.
- String concatenation policy is enforced, either allowed only for strings or
  rejected in v1.
- Boolean operators reject non-bool operands.
- Comparisons reject incompatible types.
- `if` conditions must be bool.
- Return expression must match the declared cell result type when one exists,
  or establish the inferred formula result type when the cell is inferred.
  Runtime boxing into spreadsheet values happens after typechecking.
- Multiple return expression count and types match the function signature.
- A formula can call a Go-junior function cell through its sheet namespace and
  typechecks against the function cell signature.
- Changing a function cell signature invalidates dependent formula typechecks.
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
- Arrays and slices typecheck literals, indexing, slicing, append-like helper
  calls if supported, and range loops.
- Maps typecheck literals, indexing, assignment, deletion if supported, and
  insertion-order range loops.
- Struct literals typecheck field names and field types.
- Pointer types typecheck address-of, dereference, nil assignability, field
  selection through pointers, and nil pointer checks.
- Method calls typecheck value receivers, pointer receivers, automatic
  address-taking where Go permits it, and argument types.
- Values and pointers both call methods with `x.Method()`; pointer receiver
  calls through addressable values and value receiver calls through pointers
  follow Go selector rules.
- Interfaces typecheck method sets and assignments from implementing structs
  and pointers to structs according to Go method-set rules.
- Closures typecheck captured locals and captured mutation.
- Const groups and `iota` produce expected `int64` or contextual constant
  values.
- Switch cases typecheck against the switch expression or type-switch guard.
- Type switches require an interface-typed guard and narrow the case variable
  according to the selected case type.
- Switches on interface-typed expressions compare according to the Go-junior
  comparable-type rules and reject non-comparable dynamic cases.
- Variadic calls typecheck ordinary arguments and `args...` expansion.
- `fmt.Printf` accepts `format string, args ...interface{}` and returns
  `(int64, error)`.
- `fmt.Sprintf` is pure and returns `string`.
- `fmt.Println` emits a diagnostic effect and returns `(int64, error)`.
- `panic(v)` accepts any value assignable to `interface{}` and is treated as a
  non-returning call for control-flow analysis.
- `panicOn(err)` accepts exactly `error`, returns normally only when `err ==
  nil`, and is treated as potentially non-returning for control-flow analysis.
- Channel send requires an assignable element value and a send-capable channel.
- Channel receive requires a receive-capable channel and returns either the
  element value or `(element, ok bool)` in two-value receive contexts.
- `close(ch)` requires a bidirectional or send-capable channel.
- `select` case guards must be send or receive operations; default appears at
  most once.
- `go f(args...)` requires a valid call expression; return values are ignored.
- Generic functions/types instantiate with explicit type arguments or inferred
  type arguments and reject unsatisfied constraints.
- Full `for` and `for range` loops typecheck init/condition/post statements,
  range variables, `break`, `continue`, and Go-style `fallthrough` legality in
  switch statements.
- `break`, `continue`, and `fallthrough` match Go semantics, including
  rejected `fallthrough` in type switches and final switch clauses.
- Static numeric cell references can be used directly in arithmetic without
  `number(...)` casts.
- `sheet.A1` and `Data.A1` produce the declared or inferred type for that cell,
  for example `float64`, `string`, or `bool`.
- `sheet.A1:B10` and `Data.A1:B10` produce typed ranges such as
  `[]float64` in a contextual slice position when the range has a compatible
  common element type.
- Mixed static ranges produce diagnostics unless the callee explicitly accepts
  dynamic spreadsheet `Value` elements.
- `Data.Cell(addr)` and `Data.Range(addr)` require string addresses and produce
  contextually typed values when possible, otherwise `Value` and
  an untyped spreadsheet range value.
- Unknown or untyped static cell references produce diagnostics when used where
  a concrete type is required.
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
  branches, switches, loops, range loops, returns, tuple returns,
  named-return slots, defers, calls, closures/captured environments, structs,
  fields, pointers, address-of, dereference, methods, interface dispatch,
  arrays, slices, maps, indexing, package-state reads/writes,
  diagnostic-effect calls, variadic calls, type-switch narrowing,
  panic/unwind, recover, goroutine launch, channels, send, receive,
  two-value receive, close, select, suspend/resume, Go-junior function cell
  calls, and cell/range reads.
- Define a CPS/state-machine IR layer with explicit frames, program counters,
  continuation edges, goroutine roots, defer stacks, panic/recover slots,
  channel wait records, and scheduler operations.
- Preserve diagnostic/source mapping metadata.
- Add optional constant folding for simple literals.
- Add explicit conversions where needed.
- Represent host calls by binding ID, not by text name.

Tests:

- Golden IR tests for expressions, branches, locals, and host calls.
- Constant folding tests if implemented.
- IR contains no raw user identifier names except for metadata.
- IR host calls contain host binding IDs.
- IR cell reads contain normalized sheet namespace, address, and resolved cell
  type.
- IR range reads contain normalized sheet namespace, address bounds, and
  resolved range element type.
- IR dynamic cell/range reads preserve the address expression and lower to
  runtime-observed `ctx.cell` / `ctx.range` calls.
- IR package variable reads and writes carry package slot IDs, not raw selector
  text.
- IR multiple returns use an explicit tuple shape.
- IR named returns use explicit local return slots.
- IR defers preserve LIFO execution order and immediate argument evaluation.
- IR closures carry explicit captured environment slots.
- IR pointer operations carry explicit lvalue/reference metadata rather than
  raw JavaScript object/property access.
- IR map iteration preserves insertion-order semantics.
- IR loops carry fuel-check points for entry and backedges.
- IR type switches carry explicit interface type-test cases and narrowed case
  bindings.
- IR variadic calls carry ordinary argument lists plus optional spread
  expansion metadata.
- IR diagnostic-effect calls carry diagnostic sink IDs, not direct console or
  process globals.
- IR panic nodes carry panic value expression, source span, and defer-unwind
  metadata.
- IR recover nodes carry deferred-call context metadata and return nil when not
  in a recoverable panic frame.
- IR goroutine launch nodes carry callee, arguments, source span, parent
  goroutine metadata, and ignored return metadata.
- IR channel operations carry channel element type, direction, buffer metadata,
  source span, and may-suspend effect.
- IR select nodes carry ordered cases, default case metadata, send/receive case
  operations, nil-channel disabling behavior, and scheduler selection policy.
- CPS IR golden tests assert state labels, continuation targets, and source
  spans for send/receive/select/defer/panic/recover combinations.
- IR function-cell calls carry the target sheet namespace/cell ID and expected
  function signature.

Acceptance criteria:

- JS emission can be implemented without consulting raw source text.

### Stage 7A: Source Package Resolver, Compiler, and Cache

Implement package-shaped Go-junior libraries before treating external package
calls as an integration detail.

Implementation tasks:

- Define package manifest and package source file shapes.
- Parse package declarations and import declarations for source package units.
- Parse and resolve Go-style import aliases in package source units.
- Resolve package imports through built-in, workbook, browser-cache, and trusted
  host providers.
- Provide the built-in `fmt` package through the built-in package provider as
  an early required package.
- Build a package import graph and reject import cycles.
- Typecheck package exports and internals.
- Produce an exported signature table for each package.
- Produce exported type metadata for structs, interfaces, methods, arrays,
  slices, maps, function types, and constants.
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
- Persist compiled package artifacts in OPFS for browser runtimes and under
  `~/go/pkg/gojr_js/` for Node/CLI builds by default.
- Invalidate package artifacts when source, transitive dependency, compiler
  version, backend, ABI version, host spec, or capability policy changes.
- Keep compiled package artifacts separate from runtime package variable state.
- Keep package compilation in the worker.

Tests:

- Parses a single-file package with one exported function.
- Parses a multi-file package and merges package scope correctly.
- Rejects mixed package names in one package unit.
- Accepts full Go language constructs in package source, including generics,
  channels, `go`, `select`, `close`, `panic`, and `recover`.
- Rejects only target/environment-unavailable requirements such as browser cgo
  execution with explicit diagnostics and host/WASM adapter guidance.
- Resolves imports from a fake built-in package provider.
- Resolves the real built-in `fmt` package and its `Printf`, `Sprintf`, and
  `Println` exports.
- Resolves aliased imports such as `import f "fmt"` and grouped import aliases.
- Resolves imports from a fake workbook package provider.
- Rejects unknown imports with source-span diagnostics.
- Rejects import cycles across two and three packages.
- Typechecks exported functions before formula compilation.
- Typechecks exported structs, interfaces, methods, constants, arrays, slices,
  maps, pointers, pointer receiver methods, and function types before formula
  compilation.
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
- Package function accepting `[]float64` can be called with a compatible static
  spreadsheet range such as `sheet.A1:A10`.
- Package function using `for range` iterates over arrays, slices, maps, and
  finite spreadsheet ranges with fuel limits.
- Package map iteration is deterministic insertion order.
- Package functions can call predeclared `panic` and `panicOn`.
- Package panic escapes to the calling formula as `#PANIC!` after defers run.
- Package functions can launch goroutines and communicate with channels through
  the runtime scheduler.
- Package functions can use `select` and recover panics according to Go
  semantics.
- Package generic functions/types export instantiated signature metadata and
  constraint metadata so formula callers can typecheck them.
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

- Go-junior formulas can call Go source packages compiled and cached in the
  browser when their imports and runtime requirements are available.
- cgo/native or target-specific packages remain available through typed
  host/WASM bindings or trusted precompiled package providers until equivalent
  runtime services exist.

### Stage 8: JavaScript CPS State-machine Copy-and-Patch Emitter

Emit JavaScript from direct IR and CPS/state-machine IR using compiler-owned
stencils.

Implementation tasks:

- Define stencil helpers for expressions, statements, functions, branches, and
  calls.
- Generate hygienic local names.
- Emit strict mode function source.
- Emit scheduler-compatible frame/state-machine source for may-suspend
  functions.
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
- Generated JS for `panic` unwinds defers in LIFO order and converts escaping
  panics into structured runtime panic objects.
- Generated JS for `recover` observes only valid deferred panic frames.
- Generated JS for `go f()` creates a scheduler goroutine root and ignores
  return values.
- Generated JS for channel send/receive/select uses scheduler/channel helpers
  and never busy-waits.
- Generated JS for nil-channel operations, closed-channel receives, close, and
  send-to-closed behavior delegates to tested runtime helpers.
- Generated JS for may-suspend calls stores all live locals in frame slots and
  resumes at the correct continuation state.
- Generated JS for closures emits captured environments with correct mutation
  semantics.
- Generated JS for structs, methods, and interface dispatch uses generated
  slots and tables, not raw user property names where that would be unsafe.
- Generated JS for pointer operations uses compiler-owned reference cells or
  equivalent safe lvalue objects for address-of, dereference, field selection
  through pointers, and pointer receiver calls.
- Generated JS for type switches over interfaces performs runtime type tests
  through compiler-owned helpers.
- Generated JS for variadic calls packs ordinary arguments and expands `args...`
  consistently.
- Generated JS for `fmt.Printf` routes output through the diagnostic sink, not
  `console.log`, `process.stdout`, or browser globals directly.
- Generated JS for arrays, slices, maps, indexing, slicing, switch, full loops,
  and for-range loops matches golden snapshots.
- Generated JS for maps preserves insertion-order iteration.
- Generated JS for a package function matches golden snapshots.
- Formula-generated JS calls package exports through generated package slots.
- Package-generated JS calls imported package exports through generated package
  slots.
- Package-generated JS reads and writes package variables through generated
  package state slots, not globals.
- Generated JS never contains raw user variable names.
- String literals are escaped with `JSON.stringify`.
- Host calls route through `host` using resolved binding paths or IDs.
- Static current-sheet and cross-sheet cell reads route through `ctx.cell` with
  resolved expected cell type metadata.
- Static current-sheet and cross-sheet range reads route through `ctx.range`
  with resolved expected element type metadata.
- Dynamic `Namespace.Cell` reads route through `ctx.cell` with contextual type
  metadata when available.
- Dynamic `Namespace.Range` reads route through `ctx.range` with contextual
  element type metadata when available.
- Unsupported IR nodes fail with internal compiler diagnostics.
- Generated source parses via `new Function` in tests.
- Malicious source snippets cannot break out through emitted JS:
  - `"); globalThis.evil = true; ("`
  - identifiers named like JS keywords
  - strings containing backticks and `${...}`
  - source containing `</script>`

Acceptance criteria:

- Safe, valid JavaScript is produced for all valid v1 IR programs.

### Stage 8A: Goroutine Scheduler and Channel Runtime

Implement the shared Node/browser runtime that executes CPS state machines,
goroutines, channels, `select`, `close`, panic/recover, and asynchronous host
wakeups.

Implementation tasks:

- Define runtime frame shape: function ID, goroutine ID, program counter, live
  locals, return slots, defer stack, panic state, source span stack, and budget.
- Define scheduler queues: runnable goroutines, blocked senders, blocked
  receivers, blocked selects, async host waits, and completed goroutines.
- Implement goroutine launch and lifecycle.
- Implement unbuffered channel handoff.
- Implement buffered channel FIFO queueing and capacity blocking.
- Implement channel close, closed receive, send-to-closed panic, nil channel
  blocking, and double-close/nil-close panic.
- Implement `select` registration, wakeup, cancellation of losing cases,
  default cases, nil-channel disabling, and deterministic ready-case selection.
- Implement panic propagation through frames and goroutine roots.
- Implement `recover` during deferred calls.
- Implement deadlock detection.
- Implement deterministic scheduler seed/policy and randomized stress policy.
- Implement structured runtime diagnostics with source filenames, line/column,
  static types, goroutine stacks, and scheduler state summaries.
- Expose a Node test harness for running scheduler programs without the REPL.
- Expose a browser worker adapter that runs the same scheduler and reports
  structured events/results.

Tests:

- A goroutine starts, runs, returns, and is reaped.
- Goroutine return values are ignored for `go f()`.
- Unbuffered send blocks until receive; receive blocks until send.
- Buffered channel preserves FIFO order and blocks when full.
- `len(ch)` and `cap(ch)` report buffer length/capacity.
- Receive from closed buffered channel drains buffered values before zero/false.
- Receive from closed unbuffered channel returns zero/false.
- Send to closed channel panics and runs defers.
- Close nil and double close panic and run defers.
- Nil channel send/receive block and participate in deadlock detection.
- `select` with default runs default when nothing is ready.
- `select` with one ready receive runs that receive.
- `select` with multiple ready cases follows deterministic policy under test
  seed and exercises randomized policy in stress tests.
- `select` unregisters losing cases so later sends/receives do not wake stale
  waiters.
- `select` on nil channel cases ignores those cases.
- Deadlock reports all blocked goroutines and source spans.
- Panic in child goroutine reports goroutine stack and source span.
- `recover` in deferred call stops panic and resumes normal return.
- `recover` outside deferred panic returns nil.
- Deferred calls run LIFO across suspension points.
- Named return values can be modified by deferred closures after suspension.
- Scheduler budget/fuel stops runaway goroutine creation or infinite loops with
  a structured timeout diagnostic.

Acceptance criteria:

- Compiled CPS programs with channels, goroutines, select, defer, panic, and
  recover run identically in Node and browser worker adapters for deterministic
  scheduler seeds.

### Stage 9: Function and Package Artifact Compilation Cache

Compile generated JavaScript to callable formula/package functions and cache
them.

Implementation tasks:

- Create cache key from source hash, compiler version, host spec version, and
  selected runtime options.
- Include package artifact keys for formulas that link source packages.
- Include transitive package dependency keys for compiled package artifacts.
- Store formula function artifacts in the worker memory cache.
- Store Node/CLI package artifacts under `~/go/pkg/gojr_js/` by default, using
  import-path directory layout and `.js` file suffixes.
- Support a package artifact root override for tests, CI, and explicit CLI
  workflows.
- Store browser package artifacts in OPFS, using IndexedDB only for optional
  indexes/manifests.
- Store hot package artifacts in worker memory as a first-level cache.
- Store mutable package variable runtime state separately from compiled package
  artifacts.
- Expose a JavaScript build API such as `buildPackage` and `buildPackages` that
  owns package graph loading, typechecking, emission, artifact writes, and
  invalidation.
- Have the Go `gojr build` command invoke the JavaScript build API through the
  embedded Node/V8 runtime.
- Compile with `new Function`.
- Return structured compile errors if browser/CSP rejects dynamic compilation.
- Add a development/debug option to expose generated source.

Tests:

- Same source and host spec hits cache.
- Host spec version changes invalidate cache.
- Compiler version changes invalidate cache.
- Formula cache misses when a linked package artifact changes.
- Node package artifacts write to
  `~/go/pkg/gojr_js/<import/path>.js` by default.
- Node package artifact root override writes to a temp fake GOPATH/pkg root in
  tests.
- Node package artifacts use atomic write/rename and never leave a partial
  `.js` file visible after a simulated failure.
- `gojr build <pkg>` builds the requested package and stale dependencies into
  the artifact root.
- `gojr build <pkg>` skips fresh artifacts and reports them as cache hits.
- `gojr build <pkg>` rebuilds when source content, dependency content,
  compiler version, backend ABI, host spec, capability policy, or artifact
  layout version changes.
- Package artifact metadata maps import paths to generated `.js` artifacts and
  includes dependency edges.
- Browser package artifact cache survives worker restart by reloading from
  OPFS.
- Browser OPFS artifact metadata remains consistent after interrupted writes.
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
  or unchanged Go source packages.

### Stage 10: Runtime Evaluation

Implement the evaluator context and spreadsheet value conversion.

Implementation tasks:

- Define `ctx.cell(sheetNamespace, address, expectedType)`,
  `ctx.range(sheetNamespace, address, expectedElementType)`, and value
  conversion helpers.
- Make `ctx.cell` and `ctx.range` validate normalized addresses at runtime.
- Make every `ctx.cell` and `ctx.range` call record an observed dependency
  before returning or raising a reference diagnostic when possible.
- Validate returned runtime cell values against resolved static cell/range
  types and convert mismatches into `#TYPE!` values.
- Maintain package instance state for mutable package variables.
- Evaluate `package-state` formulas in the scheduler's stable order and avoid
  pure-result memoization for them.
- Convert returned JS values to spreadsheet values.
- Catch runtime exceptions.
- Implement Go-junior panic runtime objects, defer unwinding, and conversion of
  escaping panics to `#PANIC!` spreadsheet errors.
- Enforce fuel/budget for loops and optional call count.
- Implement runtime representations for arrays, slices, ordered maps, structs,
  pointers, interfaces, methods, pointer receivers, closures, and captured
  environments.
- Implement runtime interface type tests and type-switch dispatch.
- Implement variadic call packing and `args...` expansion.
- Implement runtime representation for Go-junior function cell values and
  callable function pointers.
- Implement the diagnostic/log sink used by `fmt.Printf` and `fmt.Println`.
- Preserve insertion-order map iteration.
- Define division by zero behavior.
- Add deterministic math behavior where possible.

Tests:

- Evaluates arithmetic formulas.
- Evaluates conditionals.
- Reads cells from a fake context.
- Reads ranges from a fake context.
- Reads cells and ranges from a fake cross-sheet context.
- Reads statically typed numeric cells and uses them directly in arithmetic.
- Produces `#TYPE!` if a runtime value violates the resolved static cell type.
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
- `panic(value)` runs deferred calls, aborts the current evaluation, and
  produces `#PANIC!` with panic value and stack metadata when it escapes.
- `panicOn(nil)` returns normally.
- `panicOn(err)` panics with that error value and produces `#PANIC!` if it
  escapes.
- Deferred calls run during panic unwind and can themselves panic; the runtime
  records the final panic according to the Go-junior panic policy.
- Enforces budget on artificial loops once loops exist.
- Evaluates closures with captured reads and writes.
- Evaluates structs, methods, interface calls, arrays, slices, indexing,
  slicing, maps, switch, full for loops, and for-range loops.
- Evaluates address-of, dereference, field selection through pointers, pointer
  receiver calls, nil pointer comparisons, and nil pointer runtime errors.
- Evaluates `x.Method()` consistently for value and pointer receivers without
  any separate pointer-call syntax.
- Evaluates Go-style type switches over interface values and switch statements
  on interface-typed expressions.
- Evaluates `break`, `continue`, and `fallthrough` with Go switch/loop
  semantics.
- Evaluates variadic calls and `args...` expansion.
- Evaluates calls to Go-junior function cells and records dependencies on the
  function cell plus any observed reads performed by the called function.
- Captures `fmt.Printf` output as tagged diagnostic output with cell ID,
  generation, and recalculation pass metadata.
- Map `for range` order follows insertion order, including update,
  delete-if-supported, and reinsert cases.
- Division by zero gives expected error or value according to spec.

Acceptance criteria:

- A compiled formula can run against a fake spreadsheet context in unit tests.

### Stage 11: Node.js Runtime and Go `gojr` CLI

Implement the Node.js runtime target first so language design can be exercised
quickly from bash before browser integration begins.

Implementation tasks:

- Implement `NodeRuntimeAdapter` using the shared compiler/recalculation core.
- Implement a filesystem-backed `PackageArtifactCache`.
- Implement filesystem/workspace package providers for source packages.
- Implement Node-safe host binding providers for tests.
- Support optional `worker_threads` isolation, but allow direct in-process
  execution for fast unit tests.
- Expose JavaScript runtime/build methods callable from embedded Node/V8:
  `evaluate`, `compile`, `buildPackage`, `buildPackages`, `testPackage`,
  `testFile`, and `inspectGeneratedJS`.
- Keep the Go `gojr` binary as the primary user-facing CLI. It should invoke
  JavaScript runtime/build methods through the embedded Node/V8 runtime rather
  than duplicating compiler logic in Go.
- CLI subcommands should include:
  - `eval`: evaluate one formula with JSON-provided cells and packages
  - `compile`: compile formula or package and print diagnostics
  - `build`: compile a package graph into the artifact cache root, defaulting
    to `~/go/pkg/gojr_js/`
  - `test`: run Go-junior tests for a file or package path, equivalent in
    spirit to `go test`
  - `run-fixture`: run a spreadsheet/recalculation fixture
  - `inspect-js`: print generated JavaScript for a formula or package
  - `cache`: inspect, clear, or warm package cache entries
- `gojr build` should accept an override for the package artifact root so tests
  can use a temp directory instead of the real `~/go/pkg/gojr_js/`.
- Keep CLI output machine-readable with a JSON mode and human-readable by
  default.

Tests:

- Node runtime evaluates a simple formula without DOM or browser APIs.
- Node runtime evaluates a formula with dynamic dependencies and returns
  observed deps.
- Node runtime evaluates formulas against multi-sheet fixture data.
- Node runtime compiles and calls a source package.
- Node runtime evaluates fixtures covering const/iota, structs, interfaces,
  methods, pointer receivers, pointers, closures, arrays, slices, maps, switch,
  full for loops, and for-range loops.
- Node runtime evaluates fixtures covering type switches over interfaces and
  switches on interface-typed expressions.
- Node runtime evaluates formulas with cell-level imports and import aliases.
- Node runtime captures `fmt.Printf` and `fmt.Println` output in CLI human mode
  and JSON mode.
- Node runtime preserves insertion-order map iteration.
- Node runtime preserves mutable package variable state across calls in one
  runtime session.
- Node runtime reuses filesystem package cache on second run.
- Node runtime invalidates cache when package source changes.
- Node runtime can run without worker_threads for fast unit tests.
- Node runtime can run with worker_threads for isolation tests.
- CLI `eval` returns value, diagnostics, and observed dependencies.
- CLI `eval` returns captured `fmt.Printf` output with source cell/evaluation
  metadata.
- CLI `compile` returns package/formula diagnostics without evaluation.
- CLI `build` invokes the JavaScript `buildPackage`/`buildPackages` API through
  embedded Node/V8.
- CLI `build` writes generated package artifacts under
  `<pkgroot>/gojr_js/<import/path>.js` when a test package root override is
  supplied.
- CLI `build` writes generated package artifacts under
  `~/go/pkg/gojr_js/<import/path>.js` by default.
- CLI `build` emits a build report with built packages, skipped packages,
  artifact paths, cache keys, dependency edges, and diagnostics.
- CLI `build` leaves mutable package state out of generated artifacts.
- CLI `build` invalidates generated artifacts on source, dependency, compiler,
  ABI, host spec, capability policy, and artifact-layout changes.
- CLI `test` runs tests from a single file.
- CLI `test` runs tests from a package directory.
- CLI `run-fixture` executes a multi-cell dependency graph.
- CLI `inspect-js` prints generated JS without executing it.
- CLI `cache clear` removes package artifacts from the chosen cache directory.
- Node host bindings cannot access browser-only APIs.

Acceptance criteria:

- Developers can run Go-junior formulas, packages, tests, builds, and
  spreadsheet fixtures from bash through the Go `gojr` binary backed by the
  embedded JavaScript runtime.
- The Node target uses the same compiler, IR, emitter, typechecker, and
  recalculation engine as the browser target.

### Stage 11A: Browser Worker Runtime

After the CLI/runtime semantics are useful, move the same compile/evaluate
protocol into a Web Worker for browser integration.

Implementation tasks:

- Define browser worker protocol: compile, evaluateCell, evaluateBatch,
  disposeCache, updateHostSpec, compilePackage, resolvePackage,
  disposePackageCache, and diagnostics messages.
- Include `observedDeps`, runtime diagnostics, captured `fmt` diagnostic output,
  value version, and dependency generation in evaluation responses.
- Keep host/WASM services behind explicit RPC endpoints.
- Support cancellation by generation token.
- Ensure stale results are ignored by the main thread.
- Implement an OPFS-backed browser `PackageArtifactCache` for generated `.js`
  package artifacts and metadata.
- Use IndexedDB only for optional browser-side artifact indexes/manifests when
  OPFS directory scans are not enough.

Tests:

- Browser worker compiles valid formulas.
- Browser worker compiles valid source packages.
- Browser worker links formula compilation against cached package artifacts.
- Browser worker persists generated package artifacts in OPFS.
- Browser worker reloads OPFS package artifacts after worker termination and
  restart.
- Browser worker reports diagnostics for invalid formulas.
- Browser worker reports diagnostics for invalid source packages.
- Browser worker evaluates a batch in dependency order provided by the main
  thread.
- Browser worker evaluation returns observed dependencies for each evaluated
  formula.
- Browser worker evaluation returns captured `fmt.Printf` output for each
  evaluated formula.
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
- Track declared/inferred cell types and update the workbook/sheet type
  environment as formulas and literals change.
- Track Go-junior function cells, their editable source, compiled callable
  values, signatures, and dependents.
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
- Changing a literal or formula result type marks type-dependent formulas dirty.
- Changing a Go-junior function cell implementation marks call dependents dirty.
- Changing a Go-junior function cell signature reparses/retypechecks dependent
  callers before evaluation.
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
- Add service layer support for declared/inferred cell types, typed ranges, and
  workbook/sheet type-environment updates.
- Add service layer support for Go-junior function cells: editable source,
  compiled callable value, signature metadata, and call dependents.
- Decide which cells mirror Ivy source lines and which cells are free formula
  cells.
- Provide UI affordances or schema hooks for declaring cell/range types where
  inference is not enough.
- Wire formula bar editing to Go-junior source.
- Show calculated values in grid cells while preserving raw source in formula
  bar.
- Surface parse/type/runtime diagnostics near the selected cell.
- Surface captured `fmt.Printf` / `fmt.Println` diagnostic output near the
  selected cell and in any developer trace panel.
- Keep editor-line synchronization for the existing "spec line" column if that
  workflow remains.

Tests:

- Editing a formula cell stores raw source and displays calculated value.
- Selecting a cell shows raw formula source in the formula bar.
- Editing formula bar updates selected cell.
- Editing a Go-junior function cell updates its raw source, recompiles the
  callable value, and invalidates callers.
- Selecting a Go-junior function cell shows its editable source and signature.
- Formula diagnostics appear for invalid formulas.
- `fmt.Printf` output from a formula appears with cell and evaluation metadata.
- Numeric typed cells can be referenced directly in formulas without
  `number(...)` casts.
- Type mismatches between declared cell types, literals, and formula results
  surface as diagnostics.
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
- Define persisted Go-junior function cell shape, including declared cell type,
  raw editable source, inferred/declared signature metadata, and diagnostics.
- Define persisted workbook package source shape.
- Persist raw Go-junior source, not generated JavaScript.
- Persist Go-junior function cell source and signature metadata, not compiled
  function objects.
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
- Save/load preserves Go-junior function cell source, declared cell type, and
  signature metadata.
- Save/load recompiles Go-junior function cells before dependent callers run.
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
  approved formula-safe dynamic-JS, approved diagnostic-effect, and explicitly
  allowed package-state capabilities.
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

- Fuzz the TypeScript scanner, parser, and AST construction with random input.
- Property tests ensure generated JS contains only generated identifiers for
  user variables.
- Malicious strings and identifiers cannot escape emission.
- Formulas cannot access browser globals directly.
- Formula contexts reject graph, sheet, UI, and dynamic-JS capabilities unless
  explicitly configured as formula-safe.
- Formula contexts allow `fmt.Printf` only through the built-in diagnostic sink,
  not through direct `console`, `process`, `window`, or DOM access.
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

- Add microbenchmarks for scanning, parsing, AST construction, typechecker,
  package resolver, package compiler, emitter,
  `new Function` compile, and evaluation.
- Add spreadsheet recalculation benchmarks for common graph shapes.
- Measure main-thread latency with worker enabled.
- Measure Go `gojr` CLI cold-start, warm-cache, build-cache, and fixture-run
  latency.
- Add cache hit/miss counters.

Tests and benchmarks:

- 1,000 simple formulas compile within target budget.
- 1,000 cached formulas recalculate without recompilation.
- A workbook package graph compiles within target budget on cold cache.
- The same workbook package graph links from OPFS package artifact cache within
  budget after worker restart.
- The same workbook package graph links from Node/CLI
  `~/go/pkg/gojr_js/` filesystem cache within target budget after process
  restart.
- The Go `gojr` CLI can run representative fixtures fast enough for ordinary
  unit-test loops.
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

- Scanner token snapshot tests.
- Parser AST golden tests.
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
- Fuzz/property tests for scanner, parser, AST construction, and emitter
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
language from bash while the design is still fluid. This slice should not be
considered usable for current source-package testing until the CLI can parse,
typecheck, and execute representative Go files that use goroutines, channels,
`select`, `defer`, `panic`, `recover`, package `init`, and `fmt.Printf`.

1. TypeScript Go-style scanner/parser for package files and cell snippets,
   including filenames on every token/span/diagnostic.
2. Go-compatible resolver/typechecker for locals, packages, constants, method
   sets, interfaces, generics, channels, `go`, `select`, `panic`, `recover`,
   and package initialization.
3. Spreadsheet extensions layered on top: reserved `sheet`, sheet-name
   namespaces, typed cell/range references, imports/import aliases, and
   dependency extraction.
4. Built-in `fmt` package with `fmt.Printf`, `fmt.Sprintf`, and `fmt.Println`
   working in the Node runtime diagnostic sink.
5. Typed effect analysis marking direct, may-panic, may-defer, may-suspend,
   package-state, diagnostic-effect, sheet-effect, graph-effect, and UI-effect
   code paths.
6. CPS/state-machine IR for may-suspend functions, preserving source filenames,
   spans, static types, live locals, defer stacks, panic/recover state, and
   goroutine metadata.
7. Shared scheduler/channel runtime in Node: goroutine launch, unbuffered and
   buffered channels, `select`, `close`, deadlock detection, deterministic
   scheduler policy, panic/recover, and runtime stack diagnostics.
8. JavaScript CPS copy-and-patch emitter that can run the scheduler tests and
   direct formula tests.
9. One representative Go source package compiled by `gojr build` through the
   JavaScript build API into `~/go/pkg/gojr_js/`, then called from a formula,
   including mutable package state and a channel/goroutine smoke path.
10. Runtime tests that evaluate formulas against a fake spreadsheet context in
    Node and package tests against named source files.
11. A minimal Go `gojr` CLI/REPL path for manual bash testing: `.load`,
    `.source`, `.test`, `eval`, `compile`, `build`, `run-fixture`, and
    `inspect-js`.
12. Browser worker and webui service integration remain deferred until the full
    Go language and CLI scheduler semantics have been exercised.

Example first formulas:

```go
import "fmt"

fmt.Printf("A1=%v B1=%v\n", sheet.A1, sheet.B1)
sheet.A1 + sheet.B1
```

```go
// Cell Lib.Double, type: Go-junior function
func Double(x float64) float64 {
    return x * 2
}
```

```go
return Lib.Double(sheet.A1)
```

```go
import (
    f "fmt"
    stats "workbook/stats"
)

f.Printf("mean input starts at %v\n", sheet.A1)
return stats.Mean(sheet.A1:A10)
```

```go
return Data.A1 + sheet.B1
```

```go
row := sheet.B1
return sheet.Cell("A" + string(row))
```

```go
addr := "A" + string(sheet.B1)
return Data.Cell(addr)
```

```go
x := sheet.A1 + sheet.B1
if x > 10 {
    return math.Sqrt(x)
}
return x * 2
```

```go
if sheet.$A$1 == "" {
    return "missing"
}
return sheet.$A$1
```

Example first source package:

```go
package counter

var Count int64

func Next() int64 {
    Count = Count + 1
    return Count
}
```

Formula calling it:

```go
return counter.Next()
```

## Acceptance Criteria for the Full Project

- Go-junior has documented, tested full Go language syntax and semantics plus
  explicit spreadsheet extensions.
- Go-junior has no user-facing `number` type; integer defaults are `int64` and
  floating-point defaults are `float64`.
- Go-junior has one Go-style `error` type.
- Go-junior supports Go consts, structs, interfaces, methods, closures,
  arrays, slices, maps, indexing, switch, full for loops, for-range loops,
  conversions, all Go operators, compound assignments, Go literal forms,
  complex numbers, predeclared built-ins, `if init; condition`, methods on
  named non-struct types, embedded fields/method promotion, interface
  embedding, struct tags, two-value type assertions, comparability, map-key
  comparability enforcement, address-of composite literals, three-index
  slicing, blank/dot imports, package `init`, and current Go range forms over
  integers and iterator functions.
- Go-junior supports goroutines, channels, channel operations, `select`,
  `close`, `panic`, and `recover` through the shared CPS scheduler.
- Go-junior supports Go generics: type parameters, constraints, type sets,
  instantiation, and inference.
- Go-junior supports pointer types, address-of, dereference, pointer receiver
  methods, and Go-like method-set rules.
- Go-junior supports predeclared `panic`, `recover`, and
  `panicOn(err error)`, with defers running during unwind, valid recover calls
  stopping panics, and escaping panics reported as `#PANIC!` or structured
  goroutine panic diagnostics.
- `break`, `continue`, and `fallthrough` match Go semantics.
- Go-junior function cells store editable source, expose typed callable
  function values, and can be called by other formulas.
- Go-junior supports Go-style type switches over interfaces and switches on
  interface-typed expressions.
- Spreadsheet-cell Go-junior functions support Go-style imports and import
  aliases.
- The built-in `fmt` package supports `Printf`, `Sprintf`, and `Println` early,
  with `Printf`/`Println` routed through the diagnostic sink.
- Go-junior maps iterate in deterministic insertion order.
- Declared zero-value maps auto-initialize on first assignment for spreadsheet
  ergonomics while preserving Go-like typed map reads, delete/clear behavior,
  and deterministic insertion-order iteration.
- Target/environment limitations such as browser cgo/native execution are
  reported explicitly with diagnostics and host/WASM adapter guidance.
- Valid formulas typecheck before execution.
- Statically visible dependencies are extracted without running formulas.
- Dynamic dependencies are observed during evaluation and update the dependency
  graph automatically.
- Current-sheet and cross-sheet dependencies use the same graph model with
  explicit sheet namespaces.
- Go source packages compile, link, and cache in the browser when their imports
  and runtime requirements are available.
- Go source packages compile, link, and cache in Node.js.
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
