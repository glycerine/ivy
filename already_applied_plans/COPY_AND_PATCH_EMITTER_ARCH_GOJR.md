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

Copy-and-patch remains the primary GoJr backend architecture, but the primary hot-code target is now WebAssembly, not JavaScript source text.

The revised architecture is Wasm-first and JavaScript-hosted:

- JavaScript remains the package loader, spreadsheet graph orchestrator, dynamic host-binding layer, browser/DOM/chart integration layer, and fallback path for highly dynamic Go semantics.
- Wasm is the preferred target for hot straight-line numeric code, dense slice/array loops, fixed-layout structs, and spreadsheet kernels where CPU cost dominates.
- Copy-and-patch stencils should therefore be binary Wasm stencils first, with JavaScript stencils kept for orchestration and dynamic glue.
- C compiled to Wasm with Clang/LLVM is the default stencil generation pipeline. Other sources may be used for research, but checked-in production stencils should be small, deterministic, runtime-free Wasm fragments extracted from Clang output.

For GoJr, "copy-and-patch" now means:

- maintain a stencil library of prewritten Wasm binary fragments and JS host fragments;
- select stencil variants from typed AST and effect information;
- copy the selected Wasm or JS stencil;
- patch holes with validated function indices, type indices, LEB128 immediates, memory offsets, branch depths, import paths, type descriptors, and generated sub-fragments;
- use supernode stencils for common high-value shapes such as giant `map[string][]byte` literals and dense numeric loops.

The artifact boundary must become:

```text
package.a
├── __.PKGDEF   export/type/cache metadata only
├── _gojr.js    package-local JavaScript host/orchestration code
└── _gojr.wasm  package-local executable Wasm code, when generated
```

`__.PKGDEF` must never contain executable AST, runtime plans, full source text, dependency payloads, or JavaScript bodies.

`_gojr.js` must contain executable package-local JavaScript host code. It may contain compact package-local metadata needed for runtime execution, but not a serialized AST interpreter payload.

`_gojr.wasm` must contain package-local Wasm code. The JavaScript host should instantiate it, wire imports, pass pointers/state/fuel, and expose package functions through the normal package object.

## Paper Takeaways Applied To GoJr

The copy-and-patch paper matters to GoJr in four concrete ways:

1. Stencils are fragments with holes. For us, holes are source slots such as `__GOJR_NAME__`, `__GOJR_BODY__`, `__GOJR_TYPE__`, `__GOJR_LABEL__`, and `__GOJR_LITERAL__`.
2. Variant selection is the optimization. The emitter should choose direct stencils for known concrete typed operations and helper stencils only for genuinely dynamic Go semantics.
3. Supernodes are essential. We should not emit a node per byte for giant literals; we should emit a single compact literal-data stencil.
4. Compilation should be cheaper than AST construction where possible. The emitter should avoid building a second large IR when a typed traversal can select and patch stencils directly.

The binary paper uses CPS and register/stack stencil variants. In GoJr, the equivalent design pressure is:

- use Wasm for hot numeric kernels so `i64`, `f64`, dense memory, and predictable control flow map naturally;
- use JavaScript for orchestration, dynamic spreadsheet references, package loading, host APIs, async scheduling, and browser integration;
- avoid high-frequency JS-to-Wasm micro-calls by batching loops/kernels inside Wasm;
- avoid Wasm helper calls for individual arithmetic operations; compile whole kernels/chunks so JS-to-Wasm crossings amortize over meaningful work;
- insert cooperative preemption at JS-owned chunk boundaries rather than calling out from Wasm on every loop iteration;
- keep allocation and lifetime policy in JavaScript/GoJr runtime code; Wasm kernels receive integer offsets into GoJr-managed linear memory, not pointers to ordinary JavaScript objects;
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
9. Do not emit Wasm helper calls for individual arithmetic operations. Wasm calls should run whole kernels or fuel-bounded chunks.
10. Wasm kernels do not own general allocation or garbage collection. JavaScript/GoJr owns allocation, lifetime, arenas, and view refresh after `WebAssembly.Memory.grow()`.
11. Wasm-visible arrays and structs are passed as offsets into GoJr-managed linear memory. Ordinary JavaScript object arrays are not passed to Wasm kernels.
12. The default Wasm stencil foundry is C to Wasm through Clang/LLVM with freestanding, no-standard-library output.

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
  -> executable _gojr.wasm for hot/static kernels
  -> executable _gojr.js for host/orchestration/dynamic glue
  -> thin __.PKGDEF
```

The typed lowering should prefer the Go-shaped front AST plus `go/types.Info`. The simplified `ProgramAst` can remain for REPL/interpreter compatibility during transition, but it must not be the long-term artifact payload.

## Generated `_gojr.js` / `_gojr.wasm` Shape

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
  const wasm = await runtime.instantiatePackageWasm(ctx, gojrPackageWasmBytes, imports);

  // type descriptors and package variables
  // JS host wrappers around Wasm functions
  // JS fallback functions and dynamic methods
  // package variable initialization in go/types InitOrder
  // init functions in file/source order

  return runtime.finishPackage(ctx);
}

export default { artifact: gojrPackageArtifact, instantiateGoJrPackage };
```

There should be no `runtime.ast` and no call to `evaluatePackageArtifact` in production generated package artifacts.

For tiny inline modules, the JS host may use synchronous `new WebAssembly.Module(bytes)` where browser limits allow it. The async instantiation path must also exist for larger modules and conservative browser environments. The benchmark harness must measure both once the real emitter reaches that point.

Loop preemption policy:

- Wasm loops should be emitted as chunkable kernels when they can run long enough to affect browser responsiveness.
- JS owns the outer loop and calls Wasm with a fuel budget.
- Wasm runs up to `fuel` iterations or until complete, stores resumable state, and returns.
- JS measures elapsed time and yields to the scheduler/event loop when the configured budget is exceeded, initially targeting about 10 ms.
- Do not call JS time/preemption helpers from every Wasm loop iteration.

## Generated-Code Runtime API

The existing `runtime.ts` already has much of the semantic machinery, but it is interpreter-internal. We need a narrow compiler-facing API.

The API should be explicit and boring:

```ts
createPackageContext(artifact, options)
finishPackage(ctx)
instantiatePackageWasm(ctx, wasmBytes, imports)
callWasmKernel(ctx, kernel, state, fuel)
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

### Stencil Generation Source

The default stencil-generation source is C compiled directly to WebAssembly with Clang/LLVM.

Rationale:

- GoJr is lowering from a typed Go AST to Wasm. The source language used to manufacture reusable Wasm shapes does not have to be Go.
- C is the cleanest stencil source for pure copy-and-patch simplicity. Small C functions can describe node behaviors, loop kernels, memory loads/stores, struct-field access, numeric conversions, and branch/control-flow shapes without pulling in a language runtime.
- Clang can target Wasm directly. With flags such as `--target=wasm32`, `-ffreestanding`, and `-nostdlib`, the output can be kept to the literal Wasm instructions and relocations needed by the stencil, without libc, allocator, scheduler, or panic machinery.
- The Wasm function-body boundary is parseable and deterministic: function bodies are length-delimited in the code section, and object/export/symbol metadata can identify the function to extract. The stencil extractor must parse the Wasm object/module, locate the intended function body, record relocations and patch holes, strip names/debug/custom baggage that is not needed, and reject output that contains unexpected imports/runtime dependencies.
- C is familiar enough that stencil behavior remains easy to review, but low-level enough that it maps closely to Wasm's `i32`, `i64`, `f32`, `f64`, locals, branches, and linear-memory operations.

The C/Clang pipeline should be a development-time stencil foundry, not a runtime dependency for ordinary `gojr build`. The checked-in or generated stencil artifact is what matters:

```text
stencil.c
  -> clang --target=wasm32 -ffreestanding -nostdlib ...
  -> parse Wasm object/module
  -> extract validated function bodies/data fragments
  -> record patch sites, type signatures, imports, memory assumptions, and benchmark metadata
  -> emit deterministic GoJr stencil table
```

Other stencil sources remain acceptable for investigation:

- handwritten WAT or hand-emitted binary Wasm for tiny primitives and tests;
- Rust `no_std` for comparison when it produces cleaner Wasm than C;
- TinyGo as a semantic oracle for Go-shaped kernels;
- the standard Go Wasm compiler as a whole-program reference, not as the normal stencil source.

None of these alternatives should add a required runtime dependency to normal package compilation unless benchmarks and artifact-size checks prove that the dependency buys enough value to justify it.

### Stencil Representation

The emitter needs two stencil families:

1. Binary Wasm stencils for hot kernels, memory operations, and low-level typed fragments.
2. JavaScript host stencils for package orchestration, dynamic semantics, runtime helper calls, imports, init sequencing, and browser/spreadsheet integration.

The Wasm stencil representation is primary for compiled kernels:

```ts
interface WasmStencil {
  name: string;
  signature: WasmFunctionSignature;
  locals: WasmLocalDecl[];
  bodyBytes: Uint8Array;
  holes: WasmPatchHole[];
  imports: WasmImportRequirement[];
  memory: WasmMemoryRequirement;
  metadata: WasmStencilMetadata;
}

type WasmPatchHoleKind =
  | "typeIndex"
  | "functionIndex"
  | "localIndex"
  | "globalIndex"
  | "branchDepth"
  | "lebI32"
  | "lebI64"
  | "memoryOffset"
  | "dataOffset"
  | "callTarget";

interface WasmPatchHole {
  kind: WasmPatchHoleKind;
  offset: number;
  width: number;
  signed: boolean;
  semanticName: string;
}
```

Wasm patching must validate every hole before writing bytes:

- patched LEB128 values fit the reserved width or trigger a controlled re-emit with a larger variant;
- branch depths target valid blocks/loops;
- function/type/local/global indices refer to the assembled module tables;
- memory offsets match GoJr layout descriptors;
- imports match the declared JS host ABI;
- extracted Clang output contains no unexpected runtime imports or allocator dependencies.

JavaScript stencils should still be structured enough to prevent unsafe string pasting:

```ts
interface JsStencil {
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

Do not use ad hoc string concatenation for semantic code. Raw JS should be allowed only for already-emitted trusted fragments. Do not represent Wasm stencils as strings except in test fixtures or human-readable WAT diagnostics.

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

### Wasm Linear Memory And Allocation

Browsers garbage collect JavaScript-visible Wasm objects such as modules, instances, functions, memories, typed-array views, and ordinary JavaScript wrappers. They do not garbage collect arbitrary objects allocated inside Wasm linear memory. A `WebAssembly.Memory` backing store is reclaimed only when the memory object becomes unreachable; while it is reachable, allocations inside it are GoJr's responsibility.

Therefore GoJr allocation policy lives in JavaScript/GoJr runtime code:

- JavaScript/GoJr owns arenas, free lists, slice growth, object lifetime, and temporary-kernel scratch allocation.
- Wasm kernels receive `i32` offsets into GoJr-managed linear memory plus lengths, capacities, strides, or descriptor pointers.
- Wasm kernels should allocate little or nothing. Temporary allocations should come from a JS-owned arena that can be reset at the orchestration boundary.
- Hot Wasm-visible data should live in `WebAssembly.Memory`, not ordinary JavaScript object arrays.
- Host-only values may still use ordinary JavaScript arrays or objects when they are never passed into a Wasm kernel.
- If `WebAssembly.Memory.grow()` occurs, JavaScript must refresh affected `DataView`, `Uint8Array`, `Float64Array`, `BigInt64Array`, and `BigUint64Array` views because the backing buffer can change.

For a Go struct slice:

```go
type Point struct {
    X float64
    Y float64
    N int64
}
```

the Wasm-visible representation should be canonical GoJr-layout bytes:

```text
Point.X offset 0
Point.Y offset 8
Point.N offset 16
sizeof(Point) = 24
[]Point = { dataPtr, len, cap }
```

JavaScript writes or aliases the records through `DataView` or generated typed accessors, then calls a kernel with either `(dataPtr, len)` or a pointer to the slice descriptor. Wasm sees only integer offsets into linear memory. It never receives a pointer to an ordinary JavaScript object array.

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

Every implementation stage ends with a template expansion pass. This is not optional cleanup. It is the mechanism that keeps copy-and-patch from quietly degenerating into generic helper calls. Each pass should inspect the code just implemented, identify shapes that are common, bloated, or paying unnecessary dynamic/runtime cost, and promote those shapes into explicit stencils or supernodes with tests.

### Pre-Implementation Build Steps

Before Stage 1 proper, build the narrow infrastructure slice that removes the remaining unknowns:

1. Add `src/emitter/wasm/` with a small Wasm binary parser/extractor for Clang-produced Wasm modules. It must parse enough of the module format to find type/import/function/export/code/custom sections, extract function bodies by export name, decode local declarations, identify imported memory, and reject unexpected imports or malformed bodies.
2. Add Clang discovery/configuration for stencil generation. Discovery order: `GOJR_CLANG`, known LLVM/Homebrew paths such as `/usr/local/opt/llvm/bin/clang` and `/opt/homebrew/opt/llvm/bin/clang`, then `clang` only if it can actually compile `--target=wasm32`. Apple `/usr/bin/clang` must not be accepted merely because it exists.
3. Add a checked-in/generated stencil table so ordinary `gojr build` does not require Clang. Clang is a development-time stencil foundry; normal package compilation consumes deterministic TypeScript stencil bytes and metadata.
4. Add the first mixed JS/Wasm package artifact fixture: thin `__.PKGDEF`, executable `_gojr.js` host, executable `_gojr.wasm` kernel, and a Node test that imports the JS host, instantiates the Wasm, and executes the kernel.
5. Add benchmark/size regression tests for this first artifact shape so the implementation cannot regress to serialized AST or interpreter payloads.

### Stage 1: Emitter Scaffold And Artifact Boundary

Deliverables:

- Create `src/emitter/` scaffold.
- Define `Stencil`, `SlotKind`, `ExprFragment`, `StmtFragment`, `EmitterContext`.
- Create a minimal package JS host emitter and Wasm binary emitter.
- Remove runtime payload from `__.PKGDEF` in the new layout.
- Make generated `_gojr.js` instantiate without `evaluatePackageArtifact` for supported packages.
- Add `_gojr.wasm` archive member for packages that generate Wasm kernels.

Initial supported constructs:

- package metadata;
- imports binding by full import path;
- top-level constants with literal values;
- top-level vars with literal or zero values;
- one exported numeric function lowered to Wasm;
- one fuel-chunked numeric loop lowered to Wasm with JS orchestration;
- `[]string`;
- `[]byte`;
- `map[string][]byte`;
- empty functions and simple return-literal functions.

Tests:

- Build a tiny package and assert `__.PKGDEF` has no `runtime`, `ast`, `source`, or function body text.
- Assert `_gojr.js` has `instantiateGoJrPackage` and does not call `evaluatePackageArtifact`.
- Assert `_gojr.wasm` is present for a package with a supported numeric kernel.
- Import generated `_gojr.js` in Node and execute a simple exported function.
- Execute a generated Wasm numeric function through the JS host wrapper.
- Execute a generated fuel-chunked Wasm loop through the JS host wrapper.
- Build the `4d63.com/tz` data shape fixture and assert artifact size is proportional to source size.
- Assert `__.PKGDEF` for the data fixture is below a small threshold, for example 128 KB.

Template Expansion Pass:

- Add the initial package, JS host wrapper, Wasm function, Wasm memory, Wasm loop chunk, variable, zero-value, `[]string`, `[]byte`, and `map[string][]byte` stencils.
- Add size snapshot tests proving literal supernodes beat serialized AST size.
- Record any fallback helper calls emitted by Stage 1 and classify them as deliberate runtime semantics or future stencil candidates.

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

Template Expansion Pass:

- For each helper introduced, decide whether it should remain a helper or become a type-directed stencil in later stages.
- Add direct construction stencils for helper-backed values that are common and cheap to emit, such as byte slices, dense 64-bit typed arrays, empty maps, empty slices, and nil-able zero values.
- Add tests that generated code calls narrow helpers, not broad interpreter-shaped helpers.

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

Template Expansion Pass:

- Add stencil variants for each high-frequency typed expression shape found while implementing expression lowering: integer arithmetic, float arithmetic, string concatenation, bool comparisons, nil checks, direct selector loads, direct index loads, and tuple-producing expressions.
- Add supernodes for constant-foldable expression patterns and compact literal expression trees.
- Audit generated JS for generic `runtime.binary`, `runtime.assign`, or `runtime.evaluate*` style calls; replace all statically proven cases with direct typed stencils.

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

Template Expansion Pass:

- Add statement stencils for common structured control-flow shapes: simple `if`, `if/else`, counted `for`, `for range` over slices, `for range` over maps, direct `return`, short declaration from tuple, map comma-ok assignment, and receive comma-ok assignment.
- Add state-machine stencils only for functions that need hard `goto`; keep normal structured code on structured JS stencils.
- Add output tests proving common control flow does not lower through a generic statement interpreter.

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

Template Expansion Pass:

- Add function stencils for common effect classes: no-await/no-defer, await-capable, defer-capable, named-return, method with value receiver, method with pointer receiver, closure with captures, and init function.
- Add capture stencils that box only captured or address-taken variables, not every local.
- Add package-init supernodes for common init-order shapes, including pure constant/package var initialization and multiple init functions in source order.

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

Template Expansion Pass:

- Add descriptor stencils for named types, structs, interfaces, pointers, slices, arrays, maps, chans, and functions.
- Add interface stencils for concrete-to-interface boxing, nil interface values, typed nil interface values, method dispatch, and type switch dispatch.
- Add reflection descriptor deduplication tests so repeated type uses patch references to shared descriptors instead of duplicating descriptor source.

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

Template Expansion Pass:

- Add generic function stencils for dictionary/type-descriptor passing, generic method receivers, and generic type constructors.
- Add specialization stencils only after measuring common instantiated shapes; do not preemptively explode the template library.
- Add tests comparing generic emitted source size for reusable dictionary lowering versus specialized lowering.

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

Template Expansion Pass:

- Use the standard-library and zygo dependency closure as the first real template expansion corpus.
- Rank emitted helper calls and large emitted fragments by frequency, size, and startup cost.
- Promote the highest-payoff shapes into new stencils or supernodes before declaring the stage complete.
- Add regression tests for every promoted stencil so later work does not collapse back to generic helper-heavy output.

## Current Implementation Progress

As of June 26, 2026, the `copy-patch-wasm-stage1` backend has moved past the initial fixture stage and now writes package-local executable artifacts for a useful subset of GoJr packages.

Implemented and covered by focused tests:

- Thin `__.PKGDEF` plus executable `_gojr.js`, with `_gojr.wasm` when an `int64` add kernel is selected.
- Checked-in Clang-produced Wasm stencil extraction and mixed JS/Wasm fixture execution.
- Direct generated package functions, methods, package vars, constants, imports by full import path, and source-order `init` calls.
- `gojr build` / `buildPackages()` default to the executable `copy-patch-wasm-stage1` backend rather than the old source-envelope backend; ordinary and ambient dependency artifacts now use the resolved backend consistently.
- Stage 3 expression lowering for literals, selectors, indexing, slicing, concrete arithmetic/string/bool/complex expressions, calls, conversions, method values, method calls, type assertions, map/slice/string indexing, and tuple returns.
- Stage 4 statement lowering for blocks, declarations, returns, if/else, switch/fallthrough, type switch, for, range, select, assignment, short declaration, inc/dec, labels, simple goto state machines, defer, goroutines, sends, receives, and named returns.
- Function literals, lexical captures, recursive and mutually recursive functions.
- Variadic generated functions and function literals, including spread calls such as `xs...`.
- First-pass generated interface descriptors: concrete method dispatch, pointer receiver dispatch, interface argument/return boxing, nil interface versus typed-nil interface preservation, and concrete type assertions through `any`.
- First-pass package-local type descriptor tables for named structs and interfaces, with generated reflect-style `TypeOf`, `String`, `Name`, `Kind`, `NumField`, `Field`, `Elem`, pointer descriptors, and typed nil pointer reflection.
- First-pass imported package descriptor lookup through `importsByPath`, preserving full import-path identity on generated values so reflection can distinguish same-named local and dependency types without embedding dependency descriptors.
- First-pass composite named type descriptors for slices, maps, arrays, channels, and functions, including `Elem`, `Key`, `Len`, `In`, and `Out` metadata used by generated reflection.
- Go-correct nil-able zero values in generated code for pointers, slices, maps, channels, and functions, including named nil-able types, nil map zero/false reads, nil range behavior, nil-map assignment rejection, and reflect-visible typed nil identity.
- Package-qualified struct field descriptors, including ambiguous same-named imported field types from multiple packages.
- First-pass generated `reflect.Value` host support for `ValueOf`, `IsValid`, `IsNil`, `Kind`, `Type`, `Field`, `Interface`, `String`, `Int`, and `Bool`.
- First-pass generic function instantiation erasure for simple generic functions.
- First-pass inferred generic function calls for direct identifier calls, with type dictionaries inferred from parameter-to-argument patterns including `T`, `[]T`, `*T`, `map[K]V`, and instantiated named type arguments.
- First-pass generic named type method lowering by generic receiver base, covering instantiated values such as `Box[int64]` calling methods declared on `Box[T]` and `*Box[T]`.
- First-pass generated descriptor substitution for instantiated generic named struct zero values, so `var b Box[int64]` constructs a typed `Box[int64]` value with `T` fields zeroed as `int64`.
- First-pass generic function type-argument dictionaries for runtime type-sensitive lowering, currently covering zero values of `T` and `make([]T, n)` element initialization.
- First-pass generic receiver method dictionaries inferred from receiver runtime type names such as `Box[int64]`, covering method bodies that need zero values of receiver type parameters.
- First-pass generated pointer/addressability cells for `new(T)`, address-of locals, package variables, struct fields, array/slice elements, composite literals, dereference reads/writes, pointer field selectors, and pointer-receiver calls on addressable values.
- Descriptor-backed named struct zero values in generated code, so `var s S` and `new(S)` construct Go-correct zeroed structs rather than `null`.
- Generated `make` resolves named underlying map, slice, and channel types such as `make(Values)` where `type Values map[interface{}]interface{}`.
- Literal supernodes for large `[]byte{...}` and `map[string][]byte{...}` data, using base64 payloads instead of huge element-by-element JavaScript.
- Narrow generated helpers for common builtins: `len`, `cap`, `append`, `copy`, `delete`, and `panic`.

Current focused scoreboard:

```text
emitterWasm.test.ts: 24 pass
targeted runtime nil-map compatibility slice: 6 pass
adjacent build/bench/generated-runtime/Wasm POC suites: 67 pass
```

Still incomplete:

- The backend still lowers from `ProgramAst`; the long-term target remains direct Go-shaped AST plus `go/types.Info`.
- Type descriptors are not yet complete enough for imported private/helper types or full `reflect.Value` parity.
- Generic constraints, inference-heavy contextual return inference, full generic type descriptors for reflection, dictionary use beyond zero-value construction, and specialization are only smoke-tested.
- Pointer/addressability now has first-pass generated cells, but still needs broader coverage for unsafe pointer conversions, pointer-shaped imported descriptors, pointer receiver copy-vs-address subtleties, addressability diagnostics, and Wasm linear-memory lowering.
- Package artifacts are executable by default for the supported subset, but the explicit legacy source-envelope backend and runtime evaluator still exist elsewhere and must be removed when compiled artifacts become fully authoritative.
- Standard-library and `zygo` cutover still need broader lowering coverage and warm-cache startup work.

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

### Implemented Benchmark Harness

The initial V8/JavaScript benchmark harness now exists and should be kept as the scoreboard for copy-and-patch work.

Mechanism of action:

- `gojr/src/bench.ts` is the runtime-neutral benchmark core.
- `gojr/src/nodeBench.ts` is the Node/V8 host layer.
- The native `gojr` binary exposes the same JavaScript implementation through `gojr bench`; the Go side only reads local source files and passes JSON into the embedded runtime.
- The npm-facing CLI also exposes `node dist/src/cli.js bench`.
- The embedded runtime installs `__gojrBench`, which calls `benchmarkGoJuniorOnNode` and returns a host JSON payload.
- Optional V8 CPU profiles are captured through Node's `inspector.Session` and written as Chrome DevTools-compatible `.cpuprofile` files.
- The harness measures phases independently enough to show where time moves as the emitter changes:
  - `parse`
  - `ast-lower`
  - `typecheck`
  - `package-build`
- The harness also records artifact bloat counters:
  - `artifact_bytes_max`
  - `pkgdef_bytes_max`
  - `js_bytes_max`
  - `runtime_ast_json_bytes_max`

The first built-in benchmark cases are deliberately small and diagnostic:

- `tiny-function`: minimal function/package overhead.
- `loop-and-branch`: common statement lowering pressure.
- `byte-literal-4k`: literal bloat pressure and a small stand-in for the `4d63.com/tz` explosion.

The Wasm-first proof-of-concept benchmark is available as `gojr bench -wasm-poc`.

It measures:

- binary Wasm byte generation for tiny `i64`, `f64` memory, and fuel-chunked `f64` memory modules;
- synchronous `WebAssembly.Module` compile cost;
- synchronous `WebAssembly.Instance` instantiate cost;
- JS `number` loop execution;
- JS `BigInt` loop execution;
- whole-call Wasm `i64` loop execution;
- JS `Float64Array` loop execution;
- whole-call Wasm `Float64Array` memory loop execution;
- fuel-chunked Wasm `Float64Array` loop execution with JS-owned loop orchestration.

The current POC lives in:

```text
gojr/src/wasmPoc.ts
```

It is intentionally small and hand-emits Wasm binary bytes so we can validate the copy-and-patch target without committing to a large emitter design too early.

Usage:

```bash
cd ~/ivy/gojr
make
gojr bench -n 10
gojr bench -n 10 -case tiny-function
gojr bench -n 10 -case byte-literal-4k
gojr bench -n 10 -cache warm -case tiny-function
gojr bench -n 5 -warmup 2 -phase front-end
gojr bench -n 3 -cpuprofile /tmp/gojr-copy-patch.cpuprofile
gojr bench -n 1 ./path/to/package
gojr bench --json -n 1 -case tiny-function
gojr bench -wasm-poc -n 3 -warmup 1 -work 100000 -fuel 8192
gojr bench -wasm-poc -n 3 -cpuprofile /tmp/gojr-wasm-poc.cpuprofile
```

For fast JavaScript-only development without rebuilding the native wrapper:

```bash
cd ~/ivy/gojr
npm run build
node dist/src/cli.js bench -n 10
node dist/src/cli.js bench -n 3 -case byte-literal-4k --cpuprofile /tmp/gojr-node.cpuprofile
node dist/src/cli.js bench -wasm-poc -n 3 -warmup 1 -work 100000 -fuel 8192
```

Profile workflow:

1. Run a representative benchmark with `-cpuprofile /tmp/name.cpuprofile`.
2. Open Chrome or Chromium DevTools.
3. Load the `.cpuprofile` in the Performance/JavaScript profiler view.
4. Compare parse/typecheck/emit/build costs before and after each template expansion pass.

Benchmark policy:

- Every copy-and-patch stage must add or update at least one benchmark case when it introduces a new lowering family.
- Every template expansion pass must be justified by benchmark data or a size snapshot.
- `byte-literal-4k` should trend sharply down once literal supernodes replace serialized AST payloads.
- A package target benchmark should be used before and after changing package-cache loading so we can distinguish cold compile cost from warm cache-hit startup cost.
- Do not optimize against the current interpreter-shaped artifact as if it were the final design; use the bloat counters to prove that `runtime_ast_json_bytes_max` is disappearing.
- Wasm POC results should guide batching decisions. If a shape only wins when control stays inside Wasm, lower it as a kernel/chunk rather than as a high-frequency JS-to-Wasm micro-call.
- Fuel should be tuned by elapsed time in the JS host. The POC uses fixed fuel to expose the mechanism; the real scheduler should adapt fuel toward the browser responsiveness target.

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
- JavaScript remains the host/orchestration layer even when hot kernels are emitted to Wasm.
- Long-running Wasm kernels must support JS-owned fuel/preemption boundaries so browser main-thread execution can yield.
- The cache loader must never key packages by short package name.
- The output target remains `GOOS=js`, `GOARCH=gojr`.
- Standard library source selection still uses the native host GOOS/GOARCH/build tags for reading Go source, as already decided.

## Definition Of Done

This architecture is complete when:

- `gojr build` writes thin `__.PKGDEF` plus executable `_gojr.js` and `_gojr.wasm` when Wasm kernels are generated.
- Cached package load does not parse source or typecheck source for cache hits.
- Cached package load does not deserialize executable AST.
- Hot numeric loops can be executed from generated Wasm through JS host wrappers.
- Long-running generated Wasm loops can be chunked and preempted from JS.
- `gojr run cmd/zygo` works from compiled cached artifacts.
- A second cached `gojr run cmd/zygo` is fast enough for interactive use.
- `4d63.com/tz.a` is no longer hundreds of MB.
- Existing unit tests and selected Go distribution tests pass through the compiled backend.
- The old normal artifact path using `runtime.ast` is deleted, not left as a misleading compatibility path.
