# GoJr Copy-And-Patch Emitter Architecture

This plan records the pivot from interpreter-payload package artifacts to real generated JavaScript package artifacts.

## Current Finding

The current package artifact architecture is the wrong shape for fast cached loads.

For `4d63.com/tz.a`, the local artifact investigation showed:

```text
archive:    493,829,906 bytes
__.PKGDEF:  491,094,448 bytes
_gojr.js:     2,735,330 bytes
```

Inside `__.PKGDEF`:

```text
runtime:              491,092,610 bytes
runtime.ast:          491,092,105 bytes
runtime.runtimePlan:         471 bytes
runtime.ast.body:     491,065,618 bytes
```

The `runtimePlan` is not the source of the bloat. The source is the serialized executable GoJr AST. The worst offender is a generated declaration like:

```go
var files = map[string][]byte{
    "zoneinfo/Africa/Abidjan": []byte{84, 90, 105, ...},
}
```

The 2.7 MB generated `zoneinfo.go` expands into hundreds of MB because every byte literal becomes an object with `kind`, `literalKind`, `value`, `raw`, and `span`.

The deeper root cause is that `_gojr.js` is not yet real generated JavaScript. It is an envelope that calls back into:

```ts
evaluatePackageArtifact(ast, runtimePlan, options)
```

That means cached package load still depends on an interpreter-shaped payload.

## Decision

Copy-and-patch source emission is the primary GoJr backend architecture.

For GoJr, "copy-and-patch" means JavaScript source stencils rather than binary machine-code stencils. The core idea remains the same:

- maintain a stencil library of prewritten source fragments;
- select stencil variants from typed AST and effect information;
- copy the selected stencil;
- patch holes with validated identifiers, literals, type descriptors, branch labels, import paths, and generated sub-fragments;
- use supernode stencils for common high-value shapes such as giant `map[string][]byte` literals.

The artifact boundary must become:

```text
package.a
├── __.PKGDEF   export/type/cache metadata only
└── _gojr.js    package-local executable JavaScript only
```

`__.PKGDEF` must never contain executable AST, runtime plans, full source text, dependency payloads, or JavaScript bodies.

`_gojr.js` must contain executable package-local JavaScript. It may contain compact package-local metadata needed for runtime execution, but not a serialized AST interpreter payload.

## Paper Takeaways Applied To GoJr

The copy-and-patch paper matters to GoJr in four concrete ways:

1. Stencils are fragments with holes. For us, holes are source slots such as `__GOJR_NAME__`, `__GOJR_BODY__`, `__GOJR_TYPE__`, `__GOJR_LABEL__`, and `__GOJR_LITERAL__`.
2. Variant selection is the optimization. The emitter should choose direct stencils for known concrete typed operations and helper stencils only for genuinely dynamic Go semantics.
3. Supernodes are essential. We should not emit a node per byte for giant literals; we should emit a single compact literal-data stencil.
4. Compilation should be cheaper than AST construction where possible. The emitter should avoid building a second large IR when a typed traversal can select and patch stencils directly.

The binary paper uses CPS and register/stack stencil variants. In JavaScript source, the equivalent design pressure is different:

- use ordinary structured JS control flow where it is clearer and easier for V8;
- use `async` functions universally because goroutines/channels/select need suspension;
- use local JS variables for direct values;
- box only address-taken or captured variables;
- select stencils based on static type/effect facts rather than doing broad runtime type checks.

## Non-Negotiable Invariants

1. `__.PKGDEF` is export metadata only.
2. `_gojr.js` is package-local and executable.
3. A package artifact references dependencies by full import path and cache key; it never embeds dependency AST, source, JS, or runtime state.
4. Runtime helpers do not re-typecheck ordinary assignments or operations that `go/types` has already proven.
5. Runtime helpers exist for representation, dynamic semantics, and host integration.
6. The Go wrapper remains thin. Browser and native `gojr` use the same JavaScript compiler/runtime implementation.
7. All package cache keys include the GoJr compiler/toolchain version and artifact layout version.
8. Emitted output must be deterministic for the same source, build context, and toolchain version.

## Target Pipeline

Current package build roughly does:

```text
Go source files
  -> front parser
  -> go/types checker
  -> simplified ProgramAst
  -> runtime.ast + runtimePlan
  -> _gojr.js envelope calling evaluatePackageArtifact
```

Target package build:

```text
Go source files
  -> front parser producing Go-shaped AST
  -> go/types checker producing Info.Types, Info.Defs, Info.Uses, Info.Selections, Info.InitOrder
  -> copy-and-patch typed lowering
  -> executable _gojr.js
  -> thin __.PKGDEF
```

The typed lowering should prefer the Go-shaped front AST plus `go/types.Info`. The simplified `ProgramAst` can remain for REPL/interpreter compatibility during transition, but it must not be the long-term artifact payload.

## Generated `_gojr.js` Shape

Each generated package module should look conceptually like:

```js
// Code generated by gojr build; DO NOT EDIT.

export const gojrPackageArtifact = {
  layoutVersion: "...",
  compilerVersion: "...",
  importPath: "example.com/pkg",
  packageName: "pkg",
  cacheKey: "...",
  dependencies: ["fmt"],
  dependencyCacheKeys: ["fmt:..."]
};

export async function instantiateGoJrPackage(runtime, options = {}) {
  const ctx = runtime.createPackageContext(gojrPackageArtifact, options);
  const pkg = ctx.package;
  const imports = ctx.imports;

  // type descriptors and package variables
  // functions and methods
  // package variable initialization in go/types InitOrder
  // init functions in file/source order

  return runtime.finishPackage(ctx);
}

export default { artifact: gojrPackageArtifact, instantiateGoJrPackage };
```

There should be no `runtime.ast` and no call to `evaluatePackageArtifact` in production generated package artifacts.

## Generated-Code Runtime API

The existing `runtime.ts` already has much of the semantic machinery, but it is interpreter-internal. We need a narrow compiler-facing API.

The API should be explicit and boring:

```ts
createPackageContext(artifact, options)
finishPackage(ctx)
declareType(ctx, name, descriptor)
declarePackageVar(ctx, name, initial, typeDescriptor)
makeMap(ctx, keyType, valueType, entries?)
mapGet(map, key)
mapGetOk(map, key)
mapSet(map, key, value)
mapDelete(map, key)
makeSlice(elementType, length, capacity?)
sliceGet(slice, index)
sliceSet(slice, index, value)
sliceRange(slice, low, high, max?)
append(slice, ...values)
copy(dst, src)
makeChan(ctx, elementType, capacity)
chanSend(ctx, channel, value)
chanRecv(ctx, channel)
select(ctx, cases)
go(ctx, fn)
defer(ctx, fn)
deferScope(ctx, body)
panic(value)
recover(ctx)
toInterface(value, interfaceType)
typeAssert(value, targetType)
methodValue(receiver, methodDescriptor)
newPointer(get, set, typeDescriptor)
zeroValue(typeDescriptor)
```

These helpers are not a generic interpreter. They are primitive operations the emitted JS cannot or should not inline.

Normal concrete assignments should usually emit direct JS:

```js
x = y;
```

not:

```js
x = runtime.assign(x, y, "int");
```

Runtime checks are still required for interface assignment, type assertions, type switches, reflection, sheet values, unsafe boundaries, map key canonicalization, channel operations, and panic/recover semantics.

## Stencil Library Design

Add a new source package under:

```text
gojr/src/emitter/
```

Recommended files:

```text
stencil.ts       Stencil, Slot, Fragment, patching and validation
context.ts       EmitterContext, symbol allocation, import binding, labels
types.ts         wrappers around go/types Info and type descriptor emission
expr.ts          typed expression lowering
stmt.ts          statement/control-flow lowering
decl.ts          package declarations, functions, methods, init
literals.ts      compact literal/data supernodes
package.ts       package module assembly
runtimeApi.ts    names and contracts for generated-code runtime helpers
```

### Stencil Representation

A stencil should be structured enough to prevent unsafe string pasting:

```ts
interface Stencil {
  name: string;
  template: string;
  slots: Record<string, SlotKind>;
}

type SlotKind =
  | "identifier"
  | "property"
  | "expression"
  | "statementList"
  | "typeDescriptor"
  | "json"
  | "stringLiteral"
  | "label"
  | "rawTrustedJs";
```

Every patch slot must validate or escape according to its kind.

Do not use ad hoc string concatenation for semantic code. Raw JS should be allowed only for already-emitted trusted fragments.

### Fragment Metadata

Expression and statement lowering should return fragments with metadata:

```ts
interface ExprFragment {
  code: string;
  precedence: number;
  type: GoTypesType;
  valueMode: "value" | "address" | "tuple" | "package" | "type";
  resultCount: number;
  mayAwait: boolean;
  mayPanic: boolean;
  mayCall: boolean;
}

interface StmtFragment {
  code: string;
  mayAwait: boolean;
  completesNormally: boolean;
}
```

This lets the emitter choose direct operators, insert `await`, preserve precedence, and reject illegal contexts early.

## Type-Directed Lowering Rules

The emitter must use `go/types.Info` as the source of truth.

### Concrete Values

If the checker proves a value has a concrete non-interface type, emit direct representation operations:

```go
var x int
x = y + 1
```

Possible JS:

```js
x = y + 1n;
```

No broad assignability helper.

### Interfaces

Interface assignment and dynamic dispatch remain runtime operations:

```go
var w io.Writer
w = b
```

Possible JS:

```js
w = runtime.toInterface(b, TD_io_Writer);
```

### Maps

Map operations need helpers for Go map identity, insertion order, key canonicalization, comma-ok lookup, NaN/zero behavior, and mutation:

```js
runtime.mapSet(m, k, v);
const [value, ok] = runtime.mapGetOk(m, k);
```

The helper should not recheck that `k` is assignable to the key type when the key is statically typed; it should canonicalize the key and preserve Go semantics.

### Slices And Arrays

Use direct JS arrays or typed byte buffers where safe, but preserve Go slice semantics:

- length and capacity;
- shared backing storage;
- three-index slicing;
- append growth;
- addressability of elements;
- byte/rune/string conversions.

For Go `[]byte`, the JavaScript representation should use native byte storage:

```text
ArrayBuffer   raw backing memory
Uint8Array    8-bit unsigned read/write view over that memory
```

This is structurally close to Go slices: a backing store plus a view with offset/length/capacity metadata. GoJr should wrap `ArrayBuffer`/`Uint8Array` in a slice descriptor when Go semantics require capacity, slicing, append, aliasing, or addressability. Literal byte payloads should decode directly into this representation.

For dense 64-bit integer slices and tables, use fixed-width native typed arrays:

```text
BigInt64Array   dense []int64 storage
BigUint64Array  dense []uint64 storage
```

JavaScript scalar reads from these arrays are still `bigint`, but the storage is exact 8-byte signed/unsigned integer storage. The emitter should use these for dense `[]int64`, `[]uint64`, `[N]int64`, `[N]uint64`, and large literal tables when Go slice/array semantics allow it.

### Pointers And Addressability

This is one of the sharpest areas.

Use direct locals for ordinary values:

```js
let x = 0n;
```

Use boxes only when needed:

- address-taken locals;
- captured variables in closures;
- package vars that can be referenced by pointer;
- struct fields or array/slice elements exposed through reflection/addressing.

Pointer stencils should patch getter/setter closures:

```js
const p = runtime.ptr(() => x, (next) => { x = next; }, TD_int);
```

### Multi-Result Calls

Multi-result calls should lower to tuple temporaries:

```go
a, ok := m[k]
x, y := f()
```

Possible JS:

```js
const [a, ok] = runtime.mapGetOk(m, k);
const [x, y] = await f();
```

Assignment lowering must handle:

- short declarations;
- mixed reassign/new declaration;
- blank identifiers;
- map comma-ok;
- type assertion comma-ok;
- receive comma-ok;
- function tuple results;
- named returns.

### Defer, Panic, Recover

Generated functions should be wrapped by a defer scope stencil when they contain `defer` or `recover`:

```js
async function f(args) {
  return await runtime.deferScope(ctx, async () => {
    // body
  });
}
```

Functions without defer/recover should not pay this cost.

### Async, Goroutines, Channels, Select

All generated Go functions can be emitted as `async` initially. This is simple and correct with the scheduler.

Later optimization can emit non-async functions only when effect analysis proves they cannot suspend. This is optional and not part of the first cut.

Channel operations lower to runtime scheduler helpers:

```js
await runtime.chanSend(ctx, c, value);
const [value, ok] = await runtime.chanRecv(ctx, c);
await runtime.select(ctx, cases);
runtime.go(ctx, async () => { ... });
```

Select must use the single deterministic global PRNG for ready-case selection.

### Labels, Break, Continue, Goto

Structured loops can use native JS labels for many cases:

```js
top:
for (...) {
  break top;
}
```

Go `goto` is harder because it can jump across statement boundaries that JavaScript labels cannot express in every case. The plan should support two lowering strategies:

1. direct JS labels for structured legal cases;
2. state-machine lowering for functions that contain `goto` patterns not expressible as direct JS labels.

The emitter should detect when a function needs state-machine lowering and isolate that complexity to those functions.

### Generics

Initial generic lowering should use dictionary/type-descriptor passing:

```js
async function Identity__generic(typeArgs, value) {
  return value;
}
```

Calls with type arguments pass compact descriptors. Later, hot generic instantiations can be specialized and cached, but the first backend should prioritize correctness.

### Reflection

Reflection requires runtime descriptors. The emitter must generate package-local type descriptors for:

- named types;
- struct fields and tags;
- methods and receiver types;
- interfaces and method sets;
- pointer/slice/map/chan/array types used by the package.

Descriptors should be deduplicated in `_gojr.js` and referenced by local constants.

## Literal Supernodes

Literal-heavy generated Go source is the first high-value proof path.

Required supernodes:

### `[]byte{...}`

Emit:

```js
runtime.bytesBase64("VFppZjIA...")
```

or, for small literals:

```js
runtime.bytes([84, 90, 105, 102])
```

`runtime.bytesBase64` and `runtime.bytes` should materialize an `ArrayBuffer` plus `Uint8Array`-backed Go slice representation. Choose base64 above a small threshold.

### `map[string][]byte{...}`

Emit:

```js
runtime.mapStringBytesBase64([
  ["zoneinfo/Africa/Abidjan", "VFppZjIA..."],
  ...
])
```

This avoids one object per byte and should make `4d63.com/tz.a` proportional to source/data size.

### `[]string{...}`

Emit a direct JS array when values are strings and no addressability mutation semantics are needed at construction:

```js
["a", "b", "c"]
```

Then wrap in a Go slice representation if the value is a slice rather than an array.

### Large Scalar Tables

Add specialized stencils for:

- `[]int`, `[]int32`, `[]uint32`;
- `[]int64` via `BigInt64Array`;
- `[]uint64` via `BigUint64Array`;
- `[N]int64` and `[N]uint64` via fixed-width BigInt typed arrays where addressability/array semantics allow it;
- fixed arrays of numeric literals;
- map literals with string keys and scalar values.

Use compact typed-array payloads when Go semantics allow it. Use ordinary JS numbers for smaller integer widths where statically safe, but prefer native fixed-width typed-array storage for dense 64-bit data.

## Artifact Layout

Bump artifact layout version when the new backend lands, for example:

```text
gojr-js-v5
```

`GoJuniorPackageExportData` should remove:

```ts
runtime?: GoJuniorPackageRuntimePayload;
```

`GoJuniorPackageArchive` should remove or deprecate:

```ts
runtime?: GoJuniorPackageRuntimePayload;
```

`parseGoJuniorPackageArchive` should parse `__.PKGDEF` only as export metadata.

`nodeHost.ts` should always instantiate from `_gojr.js`; it should not prefer `archive.runtime`.

The artifact cache should reject old layout versions and rebuild them.

## Transitional Policy

Do not implement the final fix by merely moving `runtime.ast` from `__.PKGDEF` into `_gojr.js`.

That would hide the bloat and keep cached execution interpreter-shaped.

During development, a fallback interpreter backend may exist only behind an explicit development/test flag, and it must not be used for normal `gojr build`, `gojr run`, or package cache writes once the compiled backend stage is enabled.

Unsupported lowering should produce a clear diagnostic:

```text
error GOJR_EMIT001: unsupported compiled lowering for <construct>
```

not silently write an interpreter artifact.

## Implementation Stages

### Stage 1: Emitter Scaffold And Artifact Boundary

Deliverables:

- Create `src/emitter/` scaffold.
- Define `Stencil`, `SlotKind`, `ExprFragment`, `StmtFragment`, `EmitterContext`.
- Create a minimal package module emitter.
- Remove runtime payload from `__.PKGDEF` in the new layout.
- Make generated `_gojr.js` instantiate without `evaluatePackageArtifact` for supported packages.

Initial supported constructs:

- package metadata;
- imports binding by full import path;
- top-level constants with literal values;
- top-level vars with literal or zero values;
- `[]string`;
- `[]byte`;
- `map[string][]byte`;
- empty functions and simple return-literal functions.

Tests:

- Build a tiny package and assert `__.PKGDEF` has no `runtime`, `ast`, `source`, or function body text.
- Assert `_gojr.js` has `instantiateGoJrPackage` and does not call `evaluatePackageArtifact`.
- Import generated `_gojr.js` in Node and execute a simple exported function.
- Build the `4d63.com/tz` data shape fixture and assert artifact size is proportional to source size.
- Assert `__.PKGDEF` for the data fixture is below a small threshold, for example 128 KB.

### Stage 2: Runtime Helper Facade

Deliverables:

- Extract compiler-facing helpers from `runtime.ts`.
- Keep interpreter internals working while exposing stable generated-code helpers.
- Add runtime API tests independent of AST interpretation.

Tests:

- `makeMap`, `mapGet`, `mapGetOk`, `mapSet`, `delete` with string, int, float, NaN, and signed-zero keys.
- `makeSlice`, append, copy, slicing, three-index slicing.
- Pointer getter/setter identity for locals, package vars, struct fields, and slice elements.
- Interface boxing and type assertion.
- Channel send/receive/select with deterministic PRNG.
- Defer/panic/recover ordering.

### Stage 3: Expression Lowering

Deliverables:

- Lower identifiers, selectors, indexes, slices, literals, unary, binary, calls, conversions, type assertions.
- Use type-directed stencils for numeric/string/bool operations.
- Use runtime helpers only for dynamic or representation-sensitive operations.

Tests:

- Direct concrete arithmetic does not emit generic `runtime.binary`.
- String concatenation works.
- Integer/float/complex operations match Go behavior for representative cases.
- Map/slice/string indexing and slicing.
- Function calls with single and multi-result returns.
- Method calls and method values.
- Type assertions and comma-ok assertions.
- Compile output snapshot tests for key stencils.

### Stage 4: Statement Lowering

Deliverables:

- Lower blocks, if, switch, type switch, for, range, select, return, assign, short var, inc/dec, branch, labels, goto, defer, go statement, send statement.
- Add state-machine fallback for functions that contain hard `goto` shapes.

Tests:

- Existing parser/runtime control-flow tests against generated artifacts.
- Labeled break/continue/goto cases.
- For/range over arrays, slices, strings, maps, channels, integers, iterator functions.
- Switch fallthrough.
- Select with nil/default/ready cases and deterministic PRNG.
- Defer LIFO and panic/recover behavior.
- Named returns with defers.

### Stage 5: Functions, Methods, Closures, And Packages

Deliverables:

- Emit package functions and methods as real JS functions.
- Capture closures correctly.
- Box captured/address-taken variables.
- Emit init functions in file order and package initialization in Go spec order.
- Ensure each imported package initializes once.

Tests:

- Multiple `init` functions in file/source order.
- Shared dependency initialized once.
- Pointer receiver and value receiver method calls.
- Method values and method expressions.
- Closures capturing locals and package vars.
- Recursive functions.
- Mutually recursive functions.
- Package variable initialization dependencies from `Info.InitOrder`.

### Stage 6: Type Descriptors, Interfaces, And Reflection

Deliverables:

- Generate runtime type descriptors from `go/types`.
- Support interface method sets and dynamic dispatch.
- Support reflection over generated types.

Tests:

- Assign concrete pointers/values to interfaces.
- Interface nil vs typed nil behavior.
- Type switches over interfaces.
- `reflect.TypeOf`, `Elem`, fields, tags, methods.
- `reflect.Value` operations already covered by zygo dependencies.
- Imported private/helper types crossing package boundaries.

### Stage 7: Generics

Deliverables:

- Dictionary/type-descriptor passing for generic functions and generic types.
- Runtime representation for instantiated named generic types.
- Optional specialization cache later.

Tests:

- Generic identity and containers.
- Generic methods on generic types.
- Constraints involving `~`, unions, comparable, byte slices, strings.
- Generic callbacks crossing package boundaries.
- Current Go 1.27rc1 generic-method cases that motivated the latest parser/typechecker update.

### Stage 8: Standard Library And Real Project Cutover

Deliverables:

- Build core standard library packages with compiled backend.
- Build and run `github.com/glycerine/zygomys/v9/cmd/zygo` from cache.
- Remove normal interpreter artifact path.

Tests:

- `gojr build 4d63.com/tz` produces a small artifact.
- `ar -x tz.a` yields thin `__.PKGDEF` and executable `_gojr.js`.
- `gojr run cmd/zygo` first run builds artifacts.
- `gojr run cmd/zygo` second run loads cached artifacts without source recompile.
- Cache size remains reasonable, target under 150 MB for the current zygo dependency closure unless source/data size justifies more.
- Second cached zygo start target: initially under 5 seconds, later under 1 second.

## Test Plan By Concern

### Artifact Tests

- `__.PKGDEF` contains only export metadata.
- `_gojr.js` contains executable JS.
- No artifact member contains serialized `ProgramAst`.
- No artifact member contains repeated dependency payloads.
- `dependencyCacheKeys` use full import paths.
- Cache invalidates on compiler version/layout version/source hash/dependency cache key changes.

### Size Regression Tests

- Tiny package: artifact below small fixed threshold.
- `map[string][]byte` fixture: generated JS less than 10 percent of serialized AST size.
- `4d63.com/tz` fixture: `__.PKGDEF` below 128 KB; archive below a practical threshold, initially 10 MB.
- Whole zygo closure: cache size tracked and reported.

### Execution Tests

- Existing runtime tests should be migrated to run through compiled package artifacts where applicable.
- Add paired interpreter-vs-compiled tests during development only, then retire the interpreter comparison once compiled is authoritative.
- Go distribution tests should run through compiled artifacts.

### Emission Snapshot Tests

Use stable snapshot-like assertions for important stencil choices:

- direct integer addition emits direct BigInt operation;
- interface assignment emits `toInterface`;
- map assignment emits `mapSet`;
- `[]byte` emits base64 helper;
- package init does not call `evaluatePackageArtifact`.

### Negative Tests

- Unsupported compiled lowering reports `GOJR_EMIT001`.
- Invalid lvalues report compile diagnostics.
- Invalid goto crossing variable initialization remains illegal per Go rules.
- Name collisions in generated JS are escaped via symbol allocator.

## Performance Instrumentation

Add compiler timing hooks:

```text
parse
typecheck
emit
write artifact
read artifact
instantiate package
run init
```

Expose them in progress output when requested, without changing normal user output too much.

Keep `gojr test: starting ...` visibility.

The immediate priority is to eliminate interpreter-shaped artifacts before optimizing around them.

## Hardest Areas

The hardest lowering is not stencil patching. The hardest parts are semantic:

1. Addressability and pointers without boxing everything.
2. Defer/panic/recover with named returns.
3. Interfaces, typed nils, method sets, and reflection.
4. Multi-result assignment forms.
5. Goto and labels that need state-machine lowering.
6. Async suspension through channels/select/goroutines.
7. Generics and type descriptors.
8. Exact Go numeric constant and conversion behavior.

Each of these should receive focused tests before broad standard-library cutover.

## Migration Notes

- Keep the interpreter for the REPL and as a development oracle until compiled packages are ready, but do not write interpreter payloads into package cache artifacts as the final design.
- The native Go wrapper should not know about emitter details.
- Browser and native package loading must share the same JavaScript path.
- The cache loader must never key packages by short package name.
- The output target remains `GOOS=js`, `GOARCH=gojr`.
- Standard library source selection still uses the native host GOOS/GOARCH/build tags for reading Go source, as already decided.

## Definition Of Done

This architecture is complete when:

- `gojr build` writes thin `__.PKGDEF` plus executable `_gojr.js`.
- Cached package load does not parse source or typecheck source for cache hits.
- Cached package load does not deserialize executable AST.
- `gojr run cmd/zygo` works from compiled cached artifacts.
- A second cached `gojr run cmd/zygo` is fast enough for interactive use.
- `4d63.com/tz.a` is no longer hundreds of MB.
- Existing unit tests and selected Go distribution tests pass through the compiled backend.
- The old normal artifact path using `runtime.ast` is deleted, not left as a misleading compatibility path.
