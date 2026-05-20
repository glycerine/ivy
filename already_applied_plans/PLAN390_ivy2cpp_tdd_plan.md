prompt:

The original python Ivy in ~/ivy/pyivy/ivy/ivy has a facility for generating C++ code and tests from specifications in the ivy_cpp_types.py, ivy_cpp.py, ivy_to_cpp.py files. While we did not initially port the "Ivy spec to C++" translation code, we now want to plan to port the python to Go. Make a plan in plan mode to create a Go library in ~/ivy/goivy/ivy2cpp and a Go command line tool in ~/ivy/goivy/cmd/ivy2cpp that uses that ~/ivy/goivy/ivy2cpp library. Both are completely new. Deeply trace and study the python ivy C++ generation code and design our approach to port it to go while using the already complete port of the the Go package ~/ivy/goivy. I specify that ivy2cpp is its own library as an aspiration. It may turn out that the Go requirement for no circular imports makes it difficult to create the Go package ~/ivy/goivy/ivy2cpp without difficulty, in which case we will have to put it into the ~/ivy/goivy package instead. However since that package is already very large, we would like to try as a preference to have the ivy-to-C++ code in its own Go package that simply imports "github.com/glycerine/ivy/goivy" as a dependency.

Please note that the gogen/ package is incomplete and untested and may not prove a viable pattern to imitate.

Agreed. `goivy/gogen` should not be treated as a viable model beyond the very high-level fact that a generator package can import `github.com/glycerine/ivy/goivy`. For `ivy2cpp`, we should design from the Python C++ generator and the already-ported core `goivy` APIs, not from `gogen`.

**Port Plan**

1. **Create package shape**
   - Add `goivy/ivy2cpp` as a new package importing `github.com/glycerine/ivy/goivy`.
   - Add `goivy/cmd/ivy2cpp` as a thin CLI wrapper.
   - Keep all C++ generation state inside `ivy2cpp.Generator`, not package globals.
   - If `ivy2cpp` needs new accessors/helpers from `goivy`, add those helpers to `goivy`; do not make `goivy` import `ivy2cpp`.

2. **Port the Python architecture in layers**
   - `ivy_cpp.py` becomes internal C++ emission infrastructure: writers, indentation, header/impl buffers, temp names, class/member scopes.
   - `ivy_cpp_types.py` becomes C++ Ivy type support: `strbv`, `intbv`, variant types, special serialization/Z3/randomization hooks.
   - `ivy_to_cpp.py` becomes the real generator: module-to-class, type mapping, symbol declarations, expressions, actions, initial state, REPL/test/gen boilerplate.

3. **Start with a narrow MVP**
   - CLI parses Ivy-style `key=value` params plus one `.ivy` file.
   - Compile using existing `goivy` parser/compiler/isolate pipeline.
   - Emit `<basename>.h` and `<basename>.cpp`.
   - First supported target: likely `target=impl` or minimal `target=repl` without build.
   - Do not initially support automatic C++ compilation, sockets, full test generation, or Z3 random test generation.

4. **Map Python globals to explicit Go state**
   - Python globals like `target`, `emit_main`, `global_classname`, `the_classname`, `sort_to_cpptype`, `encoded_sorts`, `field_names`, `indent_level`, `delegate_enums_to`, and `native_classname` should become fields on `Generator`.
   - Python monkey-patched `emit` methods on AST/action classes should become explicit dispatcher methods:
     - `EmitExpr(goivy.Expr)`
     - `EmitAction(goivy.ActionsAction)`
     - `EmitSort(goivy.Sort)`
     - `EmitNative(...)`
     - `EmitInitialState(...)`

5. **Use existing Go core APIs where possible**
   - Module data: `goivy.Module`, `Sig`, `Relations`, `Functions`, `Actions`, `PublicActions`, `Initializers`, `Params`, `ParamDefaults`.
   - Sorts: `UninterpretedSort`, `BooleanSort`, `LogicEnumeratedSort`, `RangeSort`, `LogicFunctionSort`.
   - Constructors/destructors/variants: `SortConstructors`, `SortDestructors`, `Variants`, `IsVariant`, `VariantIndex`, `SortCard`.
   - Isolates: use `goivy.CreateIsolate` rather than re-porting Python isolate handling inside `ivy2cpp`.
   - Theory/model helpers may need audit before initial-state parity.

6. **Likely new `goivy` helper needs**
   - Stable access to solver names equivalent to Python `slv.solver_name`.
   - Ordered state-symbol enumeration equivalent to `all_state_symbols`.
   - Public helpers for function/domain/range sort inspection if current APIs are awkward.
   - Native declaration accessors, because Python native C++ anti-quotation is central.
   - Possibly exported wrappers around relevant definition/theory/model utilities for initial-state emission.

7. **Feature phases**
   - **Phase 1:** package scaffolding, CLI, options, compiled module loading, emit empty but valid `.h/.cpp`.
   - **Phase 2:** name mangling, C++ type mapping, sort declarations, symbols/state fields.
   - **Phase 3:** expression emitter: constants, applications, equality, booleans, arithmetic/interpreted ops, ITE, quantifiers with bounded loops.
   - **Phase 4:** action emitter: assign, sequence, assert/assume, call, local, if, while, choice, native action.
   - **Phase 5:** module class generation: constructor, destructor, `__init`, methods, params, progress/tick.
   - **Phase 6:** structs/enums/destructors/variants serialization and parsing.
   - **Phase 7:** REPL target and command parsing.
   - **Phase 8:** `target=test` / `target=gen` with Z3 random generation.
   - **Phase 9:** optional `build=true` support invoking `g++`/`cl`.

8. **Testing strategy**
   - Golden tests compare Go-generated `.h/.cpp` against Python-generated output for small Ivy specs.
   - Semantic tests compile generated C++ for a small subset once output stabilizes.
   - Add fixtures incrementally: enum, range, relation, function, action call, initializer, native block, struct destructor, variant, REPL public action.
   - Expect formatting differences; normalize only harmless whitespace, not symbol names or emitted logic.

9. **Big risks**
   - The Python generator depends heavily on global mutable compiler state and monkey-patching; Go needs explicit state and dispatch.
   - `target=test/gen` depends on Z3 C++ boilerplate and solver-name parity, so it should come after plain C++ emission.
   - Native C++ anti-quotation may expose missing AST/native APIs in `goivy`.
   - Initial state generation may need solver/model functionality that is present but not packaged for this use yet.
   - Circular imports should be avoidable if `ivy2cpp` imports `goivy` only, and any missing core helpers are added downward into `goivy`.


# TDD Plan For `goivy/ivy2cpp`

## Summary

Create a new Go library `github.com/glycerine/ivy/goivy/ivy2cpp` and command `goivy/cmd/ivy2cpp` to port Python Ivy’s C++ generator incrementally. Do not imitate `goivy/gogen` beyond the package-boundary idea. Drive the port with tests in `goivy/ivy2cpp`, using a hybrid oracle strategy: small exact/normalized Go tests for individual emitters, plus Python `ivy_to_cpp` integration oracles for selected fixtures.

## Key Changes

- Add library API:
  - `type Config struct { Target, ClassName, MainName, OutDir string; EmitMain, Trace, Stdafx bool }`
  - `type Output struct { Header string; Impl string; BaseName string; ClassName string }`
  - `func Generate(mod *goivy.Module, cfg Config) (*Output, error)`
  - `func CompileAndGenerate(filename string, params map[string]string, cfg Config) (*Output, error)`
- Add CLI behavior:
  - Accept Ivy-style `key=value` params and one `.ivy` file.
  - Initially support `target=impl` and `target=repl` generation only, with `build=false`.
  - Write `<basename>.h` and `<basename>.cpp` into `outdir` or the current/build directory, matching Python naming rules where practical.
- Add internal generator structure:
  - Explicit `Generator` state replaces Python globals such as target, class name, encoded sorts, C++ type registry, temp counters, and indentation.
  - Explicit dispatchers replace Python monkey-patched `.emit` methods: expressions, actions, types, native code, module/class generation.
  - Add small C++ writer/scoping helpers equivalent to `ivy_cpp.py`.

## Incremental TDD Sequence

1. **Scaffold tests before implementation**
   - Create `goivy/ivy2cpp` with tests that fail because the package does not exist yet.
   - Add test helpers:
     - `compileIvySource(t, src)` using `goivy.Parse` and `goivy.IvyCompile`.
     - `runPythonIvyToCpp(t, src, params)` using existing Python-oracle conventions and `t.TempDir`.
     - `normalizeCPP(s)` for stable comparisons: line endings, trailing whitespace, generated temp-number normalization only where necessary.
   - First tests:
     - `TestGenerateEmptyModuleProducesHeaderAndImpl`
     - `TestCLIParamParsingAcceptsIvyStyleKeyValues`

2. **Emitter primitives**
   - Test and implement name mangling equivalent to Python `varname`, `funname`, `memname`.
   - Test and implement C++ writer indentation and header/impl buffers.
   - Test and implement C++ type mapping for bool, enum, range/int, uninterpreted, functions, and tuple keys.
   - Example tests:
     - `TestVarNameMatchesPythonCases`
     - `TestCTypeBoolEnumRangeFunction`
     - `TestTupleTypeForBinaryRelation`

3. **Module shell generation**
   - Test generation of a minimal class, includes, constructor declaration, destructor declaration, and impl include.
   - Assert structural output first, not full parity.
   - Add one Python oracle fixture once the shell is stable.
   - Example fixtures:
     - Empty module
     - One enum sort
     - One relation
     - One function

4. **State symbols and sort declarations**
   - Test state-symbol discovery against compiled `goivy.Module`.
   - Implement declarations for relations, functions, constants, enum sorts, range-cardinality fields, constructors/destructors.
   - Add Python parity tests for small fixtures using normalized snippets.
   - Example tests:
     - `TestStateSymbolDeclarationsRelationAndFunction`
     - `TestEnumSortDeclarationMatchesPythonShape`
     - `TestDestructorStructDeclaration`

5. **Expression emitter**
   - Test constants, variables, applications, equality, boolean connectives, arithmetic/interpreted ops, ITE, quantifiers, and bounded loops.
   - Start with unit tests over constructed `goivy.Expr`; add integration tests from Ivy source once each family works.
   - Unsupported expression forms must return explicit errors containing the Go type and Ivy expression string.
   - Example tests:
     - `TestEmitExprBooleanAndEquality`
     - `TestEmitExprUnaryAndBinaryRelationLookup`
     - `TestEmitExprQuantifierFiniteEnumLoop`
     - `TestEmitExprUnsupportedIsActionable`

6. **Action emitter**
   - Test and implement actions in this order: sequence, assign, assert, assume, call, local, if, while, choice, native action.
   - Include tests for return parameters and variant assignment once basic calls work.
   - Example tests:
     - `TestEmitAssignNullaryAndIndexed`
     - `TestEmitIfWhileChoice`
     - `TestEmitCallWithReturnOutParam`
     - `TestEmitNativeActionAntiquotes`

7. **Initial state and constructor**
   - Start with `after init` action emission and parameter assignment.
   - Defer solver-model initial state parity until the core generator works.
   - For unsupported initial constraints, return a clear error rather than silently generating wrong C++.
   - Example tests:
     - `TestEmitAfterInitAssignment`
     - `TestConstructorInitializesParams`
     - `TestUnsupportedModelInitialStateReturnsError`

8. **REPL target**
   - Add parser/serialization boilerplate and public action dispatch for `target=repl`.
   - Test generated code contains the expected command reader dispatch for exported actions.
   - Add optional C++ compile smoke tests gated by compiler availability.
   - Example tests:
     - `TestReplDispatchForExportedAction`
     - `TestReplIgnoresInternalAction`
     - `TestGeneratedCppCompilesForSmallReplFixture`

9. **CLI integration**
   - Test CLI in a temp directory via `go run ./cmd/ivy2cpp ...` only after the library tests pass.
   - Verify output filenames, `classname=...`, `outdir=...`, `target=...`, and errors.
   - Do not implement `build=true` in v1; make it return an explicit unsupported error.

## Test Plan

- Unit tests live in `goivy/ivy2cpp` and run with `go test ./ivy2cpp`.
- Python-oracle tests skip when Python Ivy is unavailable, following existing repo helper behavior.
- C++ compile smoke tests skip when `g++` is unavailable or Z3 headers/libs are required.
- Golden/parity tests compare normalized output for small fixtures only; larger fixtures use structural assertions to avoid brittle formatting failures.
- Regression fixtures grow one feature at a time: empty, enum, relation, function, initializer, action call, local/if/while/choice, native, struct/destructor, variant, REPL.

## Assumptions

- Use the selected hybrid oracle strategy.
- Keep `ivy2cpp` as a separate package importing `goivy`; add missing accessors/helpers to `goivy` only when required to avoid circular imports.
- Do not use or modify `goivy/gogen` for this work.
- Do not use git commands.
- Initial v1 excludes `target=test`, `target=gen`, automatic C++ building, and full Z3 random test generation unless later explicitly added.
