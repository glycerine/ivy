# ivy2cpp Go port audit, 2026-05-21

This audit compares the in-progress Go port in `goivy/ivy2cpp` with the Python
Ivy-to-C++ implementation under `pyivy/ivy/ivy`. The goal is a mechanical,
faithful port. The Go implementation currently covers a small direct subset of
the Python generator. The TODOs below focus on features that are missing,
mis-implemented, or shaped differently enough that generated C++ will not match
Python behavior.

Line numbers are from the workspace state audited on 2026-05-21.

Primary Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py`
- `pyivy/ivy/ivy/ivy_cpp_types.py`
- `pyivy/ivy/ivy/ivy_cpp.py`

Primary Go references:

- `goivy/ivy2cpp/generator.go`
- `goivy/ivy2cpp/action.go`
- `goivy/ivy2cpp/expr.go`
- `goivy/ivy2cpp/types.go`
- `goivy/ivy2cpp/native.go`
- `goivy/ivy2cpp/z3.go`
- `goivy/ivy2cpp/repl.go`
- `goivy/ivy2cpp/init.go`
- `goivy/ivy2cpp/tick.go`
- `goivy/ivy2cpp/definitions.go`
- `goivy/ivy2cpp/compile.go`
- `goivy/ivy2cpp/build.go`

## DONE 001 - Port the full Python command-line/driver surface

Status: completed 2026-05-21. The Go port now accepts the Python driver
surface for this item, supports `target=class`, normalizes Python-compatible
driver defaults, emits test iteration/run defaults, writes batch outputs and
descriptors, plans/executes class compile-only builds, and covers the behavior
with ivy2cpp unit/integration tests. Verified with
`XTRACE_OFF=1 go test ./ivy2cpp -count=1`.

Go locations:

- `goivy/ivy2cpp/compile.go:29-62`
- `goivy/ivy2cpp/compile.go:71-86`
- `goivy/ivy2cpp/generator.go:13-22`
- `goivy/ivy2cpp/generator.go:50-56`
- `goivy/ivy2cpp/build.go:27-69`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:4482-4491`
- `pyivy/ivy/ivy/ivy_to_cpp.py:4513-4645`
- `pyivy/ivy/ivy/ivy_to_cpp.py:4648-4730`

Gap:

- The Go `Config` and parser accept only `target`, `classname`, `main`,
  `outdir`, `trace`, `stdafx`, `build`, and `isolate`.
- Python supports additional user-visible parameters such as `test_iters`,
  `test_runs`, `compiler`, and target `class`; it also has the v2/isolate
  control flow, conjecture insertion, specification/lib path handling, output
  file handling, compiler selection, and build command construction.
- Go rejects target `class` in `Generate`; Python's default target is `gen` but
  `class` is a supported compile-only target.

Conformance work:

- Mirror Python's parameter set and defaults, including `test_iters`,
  `test_runs`, `compiler`, `target=class`, and the exact target validation.
- Port the Python `main_int` flow: isolate handling, v2 forwarding, conjecture
  promotion, include/spec/lib search, output naming, and the separation between
  C++ generation and compilation.
- For `target=class`, emit the C++ class/header body without a main/test/gen
  harness, matching Python.
- Rework `build.go` to match Python's compiler logic: GCC/Clang versus MSVC,
  `lib/specs.cpp`, `-pthread`, include directories, outdir behavior, and
  compile-only/class behavior.

Tests to add:

- CLI golden tests for every Python target: `impl`, `class`, `repl`, `test`,
  and `gen`.
- Tests that `test_iters` and `test_runs` affect emitted C++ exactly where the
  Python generator uses them.
- Build-command tests comparing Go output with Python output for GCC and MSVC
  configurations.

## DONE 002 - Restore the Python runtime/class skeleton

Status: completed 2026-05-21. The Go generator now emits the Python-style
runtime/class skeleton for all targets: global runtime declarations and streams,
`__argv`, mutex locking, reader/timer install hooks, thread cleanup in the
destructor, target-specific REPL/test subclasses for assert/assume diagnostics,
and generator-aware `___ivy_choose` with `___ivy_stack`/`___ivy_gen` plumbing.
The local Z3 support header now exposes the `ivy_gen`/`choose` protocol used by
generated gen/test action generators. Verified with
`XTRACE_OFF=1 go test ./ivy2cpp -count=1` in `goivy`, completing in 0.102s.

Go locations:

- `goivy/ivy2cpp/generator.go:130-155`
- `goivy/ivy2cpp/generator.go:160-170`
- `goivy/ivy2cpp/generator.go:210-234`
- `goivy/ivy2cpp/generator.go:745-798`
- `goivy/ivy2cpp/repl.go:10-280`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:1947-1973`
- `pyivy/ivy/ivy/ivy_to_cpp.py:1992-2024`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2030-2069`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2071-2176`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2262-2293`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2295-2347`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2379-2393`

Gap:

- The Go class has only a small set of methods: `ivy_assert`, `ivy_assume`,
  `ivy_check_progress`, `___ivy_choose`, `__init`, and `__tick`.
- Python emits a larger runtime skeleton: global runtime includes, lock support,
  reader/timer support, thread cancellation/joining, `__argv`, `__ivy_out`,
  `__ivy_modelfile`, gen/test stack tracking, action-gen helpers, native member
  initialization hooks, and target-specific runtime glue.
- Go `ivy_assert` and `ivy_assume` just abort. Python REPL/test code uses
  target-specific assert/assume behavior, output streams, tracing, and diagnostic
  text.
- Go `___ivy_choose` returns `0` unconditionally. Python's generated code has a
  random/generator-aware choose path and a stack used by gen/test modes.
- Go `ivy_check_progress` is a no-op, while Python wires progress tracking into
  emitted counters and checks.

Conformance work:

- Port the Python class skeleton generation before filling in individual
  features. This should include all class members and helper methods emitted by
  Python for the relevant targets.
- Make target-specific assert/assume behavior match Python, especially for REPL
  and test output.
- Implement the Python `___ivy_choose` protocol, including random choices,
  action-generator stack records, and trace/model support.
- Add destructor cleanup for threads/readers/timers.
- Add `__lock`/`__unlock`, reader installation, timer installation, and imported
  callback registration paths.

Tests to add:

- Golden generated C++ snippets for class members and runtime helpers for each
  target.
- Runtime smoke tests for a model with a reader, a timer, progress properties,
  and at least one `some`/local nondeterministic choice.

## DONE 003 - Implement Python's initial-state solver path

Status 2026-05-21:

- Added `goivy/ivy2cpp/initial_state.go` with the Python-style initial-state
  split: labeled init/axiom parameter checks, init+axiom+relevant-definition
  constraint collection, Z3 model construction for non-`gen`/`test` targets,
  model evaluation into scalar state and finite-domain function/relation state,
  and progress-counter resets after initial-state construction.
- Updated `goivy/ivy2cpp/generator.go` so constructors initialize runtime
  shell, constructor parameters, progress counters, native `init` blocks, and
  solver/default state, but no longer run explicit `__init()` actions directly.
  Generated REPL glue now calls `ivy.__init()` after construction and argv
  capture, matching the Python lifecycle split.
- Updated `goivy/ivy2cpp/z3.go` and `include2cpp/ivy_go_z3.hpp` so `gen` and
  `test` targets emit init-generator constraints, call `check()`, read used
  state symbols back from the Z3 model with `__from_solver`, reset progress
  counters, set `obj.___ivy_gen`, and then invoke `obj.__init()`.
- Parameter declarations remain class members and constructor parameters, but
  are excluded from solver/default state initialization and Z3 randomization;
  initial conditions/axioms that depend on stripped parameters now fail during
  generation.
- Added fast in-process tests for inconsistent logical init, stripped-parameter
  init rejection, solver-backed finite relation initialization, relevant
  definition initialization, and `gen`/`test` init-generator constraint/eval
  shape.

Verification:

- `XTRACE_OFF=1 go test ./ivy2cpp -count=1` from `goivy/`: PASS in 0.196s.

Go locations:

- `goivy/ivy2cpp/generator.go:120-127`
- `goivy/ivy2cpp/generator.go:605-639`
- `goivy/ivy2cpp/init.go:47-82`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:1642-1648`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2345-2377`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2952-2956`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2959-2992`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2995-3001`

Gap:

- Go rejects or ignores many valid initial conditions unless they can be reduced
  to direct assignments/set actions by `initialConditionActions`.
- Python can construct an initial state using Z3 when direct initial actions are
  not enough. It includes axioms, relevant definitions, parameter stripping,
  consistency checks, model evaluation, and assignment of all relevant state
  symbols.
- Go does not distinguish the Python cases for explicit initial actions,
  generated initial state, constructor parameters, native init hooks, and
  progress-counter initialization.

Conformance work:

- Port `check_init_cond`, `emit_one_initial_state`, and
  `emit_parameter_assignments` behavior.
- Keep the simple direct-initial-action optimization only where Python uses it;
  otherwise use the solver-backed initial-state generator.
- Include axioms and relevant definitions exactly as Python does.
- Support initialization of every state symbol class Python can initialize:
  functions, relations, native types, destructors, variants, finite sorts,
  bitvectors, and strings.
- Make invalid or inconsistent initial conditions fail at the same point and with
  comparable diagnostics to Python.

Tests to add:

- Initial conditions that require quantified constraints, axioms, and derived
  definitions.
- Initial states for functions over finite domains, hash-thunk-like large
  domains, destructors, variants, bitvectors, and native types.

## DONE 004 - Port Python's type lowering and function storage model

Status: completed 2026-05-21. The Go generator now follows Python's
cardinality-driven type/storage lowering for this TODO: bools, nonnumeric
enums, numeric enums, ranges, interpreted finite sorts, `nat`, `strlit`,
native/destructor/variant sorts, and ordinary uninterpreted sorts lower through
the Python-style `ctype`/`ctypefull` decision tree. Function-typed state,
progress counters, parameters, and generated accesses now use array storage for
small finite integer-like domains and `hash_thunk` storage for large or
non-array domains, with generated `ctuple` key structs, `__hash`, equality,
memo storage, and context-aware key naming. Range array dimensions intentionally
reserve index space through the upper bound, matching Python's `cards` behavior
for ranges such as `{2..4}`.

The surrounding emitters now consume the new storage model: state declarations,
constructor/default initialization, solved initial-state assignments, progress
counter resets/ticks, action/expression function applications, REPL parsing,
native type antiquote rendering, Z3 randomization, and generator action
parameters all use the generator-aware C++ type/storage helpers. Extensional
unknown-domain quantifier and `if some` paths now iterate `hash_thunk.memo`
instead of assuming the old `std::map` representation. Verified with
`XTRACE_OFF=1 go test ./ivy2cpp -count=1` in `goivy`, completing in 0.208s.

Go locations:

- `goivy/ivy2cpp/types.go:10-47`
- `goivy/ivy2cpp/types.go:91-168`
- `goivy/ivy2cpp/generator.go:272-332`
- `goivy/ivy2cpp/generator.go:527-561`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:238-241`
- `pyivy/ivy/ivy/ivy_to_cpp.py:245-261`
- `pyivy/ivy/ivy/ivy_to_cpp.py:282-309`
- `pyivy/ivy/ivy/ivy_to_cpp.py:364-389`
- `pyivy/ivy/ivy/ivy_to_cpp.py:396-432`
- `pyivy/ivy/ivy/ivy_to_cpp.py:445-466`
- `pyivy/ivy/ivy/ivy_cpp.py:238-296`

Gap:

- Go lowers every function sort to `std::map<key,value>` and primitive integer
  sorts mostly to `long long`.
- Python chooses storage and C++ types based on sort cardinality and sort kind:
  arrays for small finite domains, nested arrays for small multi-argument
  functions, `hash_thunk` for large/non-array maps, unsigned/ranged integer
  types where applicable, and special types for bitvectors and strings.
- Python has C++ type objects and parameter-passing policy (`ByValue`,
  `ConstRef`, `ReturnRef`, etc.). Go emits strings directly and passes almost
  everything by value except multiple returns.

Conformance work:

- Port Python's `ctype`, `ctypefull`, `ctype_constructor`, `ctype_function`,
  `large_thresh`, `is_large_type`, and parameter-passing logic.
- Add C++ declarations for `ctuple` and `hash_thunk`, including `__hash`, memo
  storage, default values, and equality semantics.
- Replace unconditional `std::map` lowering with Python's array/hash-thunk
  decision tree.
- Use Python's exact primitive representations for bool, enum, range, int, nat,
  uninterpreted, bitvector, string-literal, native, destructor, and variant
  sorts.

Tests to add:

- Golden type declarations for small finite functions, large finite functions,
  multi-argument functions, int/nat domains, and uninterpreted domains.
- Compile tests that exercise generated equality, hash, stream, solver, and
  randomization support for each storage class.

## DONE 005 - Port `ivy_cpp_types.py` bitvector and string C++ types

Status update, May 21, 2026:

- Implemented the C++ interpreted type model for `bv[N]`, `strbv[N]`, and
  `intbv[lo][hi][bits]` in `goivy/ivy2cpp/cpp_types.go`.
- Wired these interpretations through sort declaration emission, scalar type
  lowering, cardinality, zero values, array/hash-thunk storage choices, REPL
  parsers/writers, generated Z3 setup/randomization, and support-header Z3
  bitvector/string sort helpers.
- Added Python-style bitvector expression emission for masked numerals, `bvand`,
  `bvor`, `bvnot`, `cast`, arithmetic masking, `concat`, and `bfe[...]`
  extraction.
- Added fast in-process shape tests covering primitive bitvector lowering,
  `StrBV`/`IntBV` helper class skeletons, REPL hooks, Z3 solver/randomization
  specializations, generated setup, and TODO004 storage interactions.
- Verification: `XTRACE_OFF=1 go test ./ivy2cpp -count=1` from
  `/Users/jaten/go/src/github.com/glycerine/ivy/goivy` passed in `0.299s`.

Go locations:

- `goivy/ivy2cpp/types.go:10-47`
- `goivy/ivy2cpp/expr.go:155-200`
- `goivy/ivy2cpp/z3.go:181-221`

Python references:

- `pyivy/ivy/ivy/ivy_cpp_types.py:12-124`
- `pyivy/ivy/ivy/ivy_cpp_types.py:128-178`
- `pyivy/ivy/ivy/ivy_cpp_types.py:179-245`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3060-3103`

Gap:

- Go has no equivalent of Python's `XBV`, `StrBV`, or `IntBV` C++ helper
  classes.
- Go expression emission treats arithmetic as ordinary C++ infix operations and
  does not implement Python's bitvector masking, extraction, concatenation,
  signed/unsigned conversions, or string-bitvector behavior.
- Go Z3 sort registration has no bitvector or string-bitvector path.

Conformance work:

- Port the helper C++ declarations generated by `ivy_cpp_types.py`.
- Add expression emission for all Python special bitvector/string operators:
  bit extraction, concatenation, casts, shifts, arithmetic masking, comparisons,
  and string literal conversion.
- Add Z3 conversion, randomization, and parser/writer support for these types.

Tests to add:

- Fixtures using fixed-width numeric bitvectors, string bitvectors, casts,
  extraction, concatenation, arithmetic overflow, and solver-generated values.

## DONE 006 - Replace the Go variant implementation with Python's ref-counted wrapper

Status update, May 21, 2026:

- Replaced the generated flat variant-supertype representation with a
  Python-style ref-counted wrapper: `wrap`, `twrap<T>`, `tag`, `ptr`, copy
  constructor, assignment, destructor, `temp_counter`, `prepare`, `cleanup`,
  `__hash`, and `unwrap<T>`.
- Updated variant relation, `exists`, `some`, `if some`, assignments, call
  arguments, and single-return call assignment to use Python-style
  `isa`/downcast/upcast expressions instead of direct payload fields.
- Emitted variant stream output, `_arg`, `__ser`, `__deser`, and Z3
  `__from_solver`, `__to_solver`, and `__randomize` specializations with
  `*>:super:sub` relation names.
- Added the missing generated `__hash` method for scalar destructor/struct
  fields so struct variants can participate in wrapper hashing.
- Added shared Z3 `exists`/`forall` helpers to `include2cpp/ivy_go_z3.hpp` for
  the generated variant solver templates.
- Updated variant shape tests from the old flat `__tag`/payload-field layout to
  the wrapper API and added checks for parser/serializer/Z3 specialization
  shape.
- Verification: `XTRACE_OFF=1 go test ./ivy2cpp -count=1` from
  `/Users/jaten/go/src/github.com/glycerine/ivy/goivy` passed in `0.330s`.
- Slow compile-only smoke: `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp
  -run 'TestVariantSupertypeAssignmentCompiles$' -count=1` passed in `1.118s`.

Go locations:

- `goivy/ivy2cpp/generator.go:347-436`
- `goivy/ivy2cpp/types.go:127-144`
- `goivy/ivy2cpp/expr.go:202-223`
- `goivy/ivy2cpp/action.go:379-417`

Python references:

- `pyivy/ivy/ivy/ivy_cpp_types.py:247-497`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2420-2674`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3811-3886`

Gap:

- Go emits a flat struct with `__tag` and one value field per variant subtype.
- Python emits a `VariantType` wrapper with `tag`, a ref-counted `wrap *ptr`,
  copy/assignment/destructor logic, `isa`, `downcast`, `upcast`, equality,
  stream output, `_arg` parsing, serialization/deserialization, Z3 conversion,
  and randomization.
- Go's variant relation check only compares tag and stored value. It does not
  match Python's wrapper semantics and cannot support solver, REPL, or native
  conversions.
- Go call/assignment paths do not perform Python's variant upcast/downcast
  handling.

Conformance work:

- Port `VariantType` generation mechanically from `ivy_cpp_types.py`.
- Generate Python-compatible constructors, copy behavior, equality, streaming,
  parser, serializer, deserializer, Z3 conversion, and randomizer methods.
- Update assignment, call returns, derived definitions, and expression
  application to insert upcasts/downcasts where Python does.
- Ensure variant relation symbols use the Python naming and semantics, including
  subtype tests and solver encoding.

Tests to add:

- Variant construction, copying, assignment, equality, REPL parse/print,
  serialization, Z3 model generation, and action return values requiring upcast.

## DONE 007 - Implement full destructor/struct support

Status: completed 2026-05-22. `goivy/ivy2cpp/destructor.go` (517 lines) emits
in-struct `operator==` / `operator<` / `operator<<` (matching Python
`field_eq`, `ivy_to_cpp.py:222-226`), plus `_arg`, `__ser`, `__deser`,
`__from_solver`, `__to_solver`, and `__randomize` for destructor sorts.
Fields can be scalars, finite functions, nested struct/variant/native, or
hash-thunked large maps. Coverage includes
`TestDestructorStructDeclaration`,
`TestDestructorFunctionNotEmittedAsMutableState`,
`TestDestructorMultiArgFieldDeclaration`,
`TestDestructorHashThunkField`,
`TestDestructorStructStreamMultiArg`,
`TestDestructorMultiArgFieldRoundTrip`,
`TestDestructorSerDeserShape`,
`TestDestructorArgShape`,
`TestDestructorZ3ImplShape`,
`TestDestructorRandomizeSkipsUninterpretedRange`.

Go locations:

- `goivy/ivy2cpp/generator.go:334-345`
- `goivy/ivy2cpp/generator.go:438-478`
- `goivy/ivy2cpp/types.go:102-116`
- `goivy/ivy2cpp/repl.go:10-280`
- `goivy/ivy2cpp/z3.go:181-282`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:2213-2254`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2420-2674`
- `pyivy/ivy/ivy/ivy_cpp.py:238-296`

Gap:

- Go emits destructor sorts as flat C++ structs with scalar fields only.
- Python handles destructor fields whose types can be arrays/functions, emits
  nested C++ types, equality helpers, stream operators, argument parsers,
  serializers/deserializers, Z3 converters, randomizers, and hash helpers.
- Go does not emit `_arg`, `__ser`, `__deser`, `__from_solver`, `__to_solver`,
  or randomization helpers for destructor sorts.

Conformance work:

- Generate field types using Python's full C++ type object model, not just the
  result sort of the destructor.
- Port stream, parser, serializer, deserializer, solver, randomizer, equality,
  and hash support.
- Ensure destructor fields that are functions/arrays/hash-thunks get initialized
  and compared exactly like Python.

Tests to add:

- Struct sorts with scalar fields, finite-function fields, nested struct fields,
  variant fields, and bitvector/string fields.

## DONE 008 - Match Python's state-symbol and extensional-relation analysis

Status update, 2026-05-21:

- Added `allStateSymbols()` in `goivy/ivy2cpp/generator.go` (mirrors Python
  `all_state_symbols`): iterates `Sig.AllSymbols()` and excludes constructors
  + solver-interpreted symbols via `goivy.SolverName(...)`.
- Added `goivy/ivy2cpp/extensional.go` with `extensionalRelations()` matching
  Python `extensional_relations` (ivy_to_cpp.py:44-76): the "bad" set comes
  from scanning every action's `IterSubactions()` for assigns/havocs/call-
  returns whose LHS is not a simple point and not false; the "inited" set
  comes from recursing `Mod.Initializers` (with a fallback to init mixins for
  modules compiled with `create_isolate=false`) and picking up
  `r(X) := false` over all-variable LHS args. Cached on `*Generator`.
- Replaced the narrow `findExtensionalRelationBound` family in `expr.go` with
  `matchExtensionalBoundExprs` (Python ivy_to_cpp.py:3351-3377): full
  polarity tracking through Not/Implies/Or/And and derived-definition
  unfolding via `goivy.Substitute`.
- Reworked `emitExtensionalQuant` to peel multiple quantified variables
  sharing one extensional atom (Python ivy_to_cpp.py:3445-3453), with
  nested numeric loops for remaining unbound variables.
- Updated `emitIfSomeExtensional` to use the new matcher (with a small
  Const→Variable substitution since goivy compiles `some` parameters to
  locals).
- Added `emitExtensionalRelationClear` in `extensional.go` to translate
  `r(X) := false` over a hash_thunk-backed extensional relation to
  `r.memo.clear();`, matching the semantic effect of Python's
  `emit_assign_large` make_thunk path for this shape (full make_thunk
  emission is reserved for TODO 013).
- Updated the four existing extensional tests in `ivy2cpp_test.go` to
  include explicit `after init { r(X) := false; }` blocks — Python's
  algorithm requires init-to-false for a relation to be extensional, and
  the old tests inadvertently verified Go's over-permissive behavior.
- Added seven TODO 008 tests:
  `TestExtensionalRelationDetectedFromInitializer`,
  `TestNonExtensionalRelationDueToBadUpdate`,
  `TestUninitializedRelationNotExtensional`,
  `TestExtensionalThroughDerivedDefinition`,
  `TestExtensionalQuantifierMultipleVariables`,
  `TestExtensionalRelationViaPolarityNegation`,
  `TestExtensionalIfSomeBoundedByDerivedDefinition`.

Verification:

- `cd ~/ivy/goivy && make test`: PASS for every package, ivy2cpp in 0.25s.
- `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -count=1 -run
  'TestExtensionalRelationDetectedFromInitializer|TestExtensionalThroughDerivedDefinition|TestExtensionalQuantifierMultipleVariables|TestExtensionalRelationViaPolarityNegation|TestExtensionalIfSomeBoundedByDerivedDefinition'`:
  PASS in 1.94s (generated C++ compiles for each shape).

Go locations:

- `goivy/ivy2cpp/generator.go:526-563` (allStateSymbols)
- `goivy/ivy2cpp/extensional.go` (extensionalRelations + helpers)
- `goivy/ivy2cpp/expr.go:394-617` (rewritten emitExtensionalQuant +
  matchExtensionalBoundExprs)
- `goivy/ivy2cpp/action.go:137-160, 268-340` (emitAssign clear path +
  emitIfSomeExtensional rewire)

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:33-35`
- `pyivy/ivy/ivy/ivy_to_cpp.py:44-76`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3351-3377`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3380-3465`

## DONE 009 - Port Python method signatures, parameter passing, and return handling

Status update, 2026-05-21:

- Added `goivy/ivy2cpp/ptype.go` with the four passing-policy types
  (`ValueType`, `ConstRefType`, `RefType`, `ReturnRefType{Pos int}`)
  mirroring Python `ivy_to_cpp.py:396-412`, plus `annotateAction` /
  `getParamTypes` (Python 1479-1517) and `mayAlias` / `rootVar` /
  `isDestructorName` (Python 1522-1530). The (param_types, return_types)
  annotation is cached per-generation on a new `Generator.ptypeCache` field
  (per goivy/CLAUDE.md section C — Python attaches these to the action
  object; Go cannot monkey-patch `*Action`).
- `actionAssigns` walks `IterSubactions()` and uses a local
  `modifiedRootName` helper that strips destructor applications via
  `g.Mod.DestructorSorts` directly. `goivy.ModifiesSingle` cannot be relied
  upon here because `g.Mod.Cfg.ActCfg.Context.GetDomain()` is not always
  threaded in the ivy2cpp call path (e.g. tests using `compileIvySource`
  with `create_isolate=false`).
- `methodSignature` (`generator.go:633-699`) rewritten to consume the
  annotation: `rtypes[0].Make(ctype(rs[0]))` produces the return type
  (where `ReturnRefType.Make` returns `"void"`); each input param is
  `ptypes[i].Make(scalarType) + " " + varName`; function-sorted params
  retain the existing `cppFunctionStorageDecl` (matching Python's
  `sym_decl(p)` ternary at `ivy_to_cpp.py:1539`); trailing
  `ReturnRefType{Pos: pos >= len(formals)}` returns are appended as
  `RefType{}.Make(qualifiedType) + " " + varName`; the `virtual ` prefix
  fires only on declaration form for non-gen, non-inline (Python lines
  1559-1560). Multi-return validation rejects exported multi-output
  actions exactly as Python does (lines 1565-1567).
- `emitMethods` (`generator.go:683-707`) suppresses both the synthetic
  primary-return local and the trailing `return X;` when `rtypes[0]` is a
  `ReturnRefType` (the input slot IS the storage). For private multi-return
  actions whose primary return is `ValueType`, the synthetic local and
  return are still emitted.
- `emitCall` (`action.go:422-548`) replaced with a line-by-line port of
  Python `emit_call` (`ivy_to_cpp.py:3811-3886`). The unified path handles
  single-return, multi-return, alias-safety temporaries (per-output
  `mayAlias` check vs other inputs; pre-call `__tmp` save and post-call
  copy-back when `iparg != rv` OR an alias is detected), assignment prefix
  vs trailing-ref routing based on `rtypes[0]`, per-argument variant
  upcast at the formal-sort boundary, and `___ivy_stack` push/pop for
  gen/test. The previously incorrect *result-side* variant upcast at the
  old `action.go:459` is removed (Python upcasts only arguments).
- Updated `TestGeneratedMultipleReturnActionCompiles` to trim
  `PublicActions` to mirror `create_isolate`, and rewrote its expected
  substrings to the annotation-driven signature
  (`color split(color c, bool& good)`) and call-site
  (`saved = split(green, ok);`).
- Replaced `TestReplDispatchWritesMultipleReturns` with an
  expected-rejection test, since exporting a multi-return action is
  Python-forbidden at `ivy_to_cpp.py:1565-1567`.
- Added eight TODO 009 tests:
  `TestPrivateActionStructParamUsesConstRef`,
  `TestPrivateActionStructParamModifiedUsesValue`,
  `TestReturnAliasingInputUsesRefType`,
  `TestCallAliasSafetySwapInputAndOutput`,
  `TestCallNoAliasNoTempEmitted`,
  `TestPublicActionAllValueTypeSignature`,
  `TestVirtualKeywordOmittedForGenTarget`,
  `TestCallSiteVariantUpcastOnArgumentOnly`. Each verifies generated-output
  shape and (where applicable) compiles the C++ under `SLOW_CPP_TEST`.

Verification:

- `cd ~/ivy/goivy && make test`: PASS for every package, ivy2cpp in 0.18s.
- `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -count=1 -run
  'TestGeneratedSingleReturnActionCompiles|TestGeneratedMultipleReturnActionCompiles|TestVariantSupertypeAssignmentCompiles|TestPrivateActionStructParamUsesConstRef|TestPrivateActionStructParamModifiedUsesValue|TestReturnAliasingInputUsesRefType|TestCallAliasSafetySwapInputAndOutput|TestCallNoAliasNoTempEmitted|TestPublicActionAllValueTypeSignature|TestVirtualKeywordOmittedForGenTarget|TestCallSiteVariantUpcastOnArgumentOnly'`:
  PASS in 4.23s (generated C++ compiles for each shape).
- `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -count=1` (full slow
  sweep): PASS in 59.16s.

Go locations:

- `goivy/ivy2cpp/ptype.go` (Ptype + four types + annotation + mayAlias)
- `goivy/ivy2cpp/generator.go:41-71` (ptypeCache field on Generator)
- `goivy/ivy2cpp/generator.go:625-699` (rewritten methodSignature)
- `goivy/ivy2cpp/generator.go:683-707` (emitMethods annotation gating)
- `goivy/ivy2cpp/action.go:422-548` (rewritten emitCall)

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:396-412`
- `pyivy/ivy/ivy/ivy_to_cpp.py:1479-1530`
- `pyivy/ivy/ivy/ivy_to_cpp.py:1542-1571`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3811-3886`

## DONE 010 - Port derived definitions, constructors, and skolemized helpers

Status update, 2026-05-21:

- Extracted `emitSomeAction` and `emitMethodDeclLine` from the
  per-action body of `emitMethods`. Both helpers mirror Python
  `emit_some_action` (`ivy_to_cpp.py:1592-1625`) and the
  `emit_method_decl(header,...) + ';'` pair
  (`ivy_to_cpp.py:1595-1597`). Behavior for ordinary actions is
  unchanged.
- Replaced `emitDefinitionDecls` / `emitDefinitions` in
  `goivy/ivy2cpp/definitions.go` with the Python synthetic-action
  route: `derivedActionFor(d)` builds an `AssignAction(retval, rhs)`
  where each bound `*goivy.LogicVariable` is skolemized to a fresh
  `*goivy.Const` named `fml:<varname>` (Python `var_to_skolem`
  pattern, `ivy_logic_utils.py:1572`), substitutes the variables in
  the body via `goivy.Substitute`, attaches `formal_params` /
  `formal_returns`, and dispatches through `emitSomeAction` /
  `emitMethodDeclLine`. The old `definitionSignature` helper was
  deleted — derived definitions now share the ptype annotation
  pipeline with ordinary actions.
- Added `goivy/ivy2cpp/constructors.go` with `constructorActionFor`
  (Python `emit_constructor`, `ivy_to_cpp.py:1364-1375`),
  `emitConstructorDecls`, and `emitConstructors`. For each sort
  constructor, the helper builds a `Sequence(...)` of
  `AssignAction(d_i(retval), fml:X_i)` for each destructor `d_i` in
  `g.Mod.SortDestructors[sortName]`, again routing through the shared
  `emitSomeAction` / `emitMethodDeclLine` helpers.
- Wired both new emitters into `Generate` immediately after their
  derived-definition counterparts, matching Python's emission order at
  `ivy_to_cpp.py:2308-2313`.
- Fixed `goivy.Substitute` (`goivy/logicutil.go`) to handle
  `*LogicNativeExpr` (recurse into `CompiledChildren`) and `*NativeCode`
  (leaf). Python's `substitute_ast` walks `ast.args` generically; the
  Go switch was missing these cases, which broke derived definitions
  whose RHS contains a native expression (e.g.,
  `definition lt(x:idx,y:idx) = <<< `+"`x`"+` < `+"`y`"+` >>>`).
- Updated two legacy tests whose assertions tested the old direct
  `return rhs;` emission
  (`TestImplDerivedDefinitionEmitsMethodNotState` and
  `TestNativeDefinitionEmitsTemplateMethod`). The new emission matches
  Python's `val = rhs; return val;` shape from `emit_some_action`. The
  native-definition test now also asserts `const idx&` parameter
  passing because `interpret idx -> <<< int >>>` registers idx in
  `mod.NativeTypes` (`compiler_decl.go:1047`), so `isStructSort(idx)`
  returns true and `annotateAction` selects `ConstRefType`
  (`ptype.go:103`).
- Added six TODO 010 tests in `goivy/ivy2cpp/ivy2cpp_test.go`:
  `TestDerivedDefinitionEmitsMethod`,
  `TestZeroArgDerivedDefinitionEmitsMethod`,
  `TestDerivedDefinitionStructParamUsesConstRef`,
  `TestSortConstructorEmitsMethod`,
  `TestConstructorExcludedFromStateSymbols`, and the
  `SLOW_CPP_TEST`-gated `TestDerivedAndConstructorCompile`.

Verification:

- `cd ~/ivy/goivy && make test`: PASS for every package, ivy2cpp in
  0.19s.
- `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -count=1 -run
  'TestDerivedDefinitionEmitsMethod$|TestZeroArgDerivedDefinitionEmitsMethod|TestDerivedDefinitionStructParamUsesConstRef|TestSortConstructorEmitsMethod|TestConstructorExcludedFromStateSymbols|TestDerivedAndConstructorCompile'`:
  PASS in 2.08s (generated C++ compiles for each shape).
- `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -count=1` (full slow
  sweep): PASS in 61.07s.

Go locations:

- `goivy/ivy2cpp/generator.go:618-624` (emitMethodDeclLine)
- `goivy/ivy2cpp/generator.go:740-786` (emitSomeAction)
- `goivy/ivy2cpp/definitions.go:114-181` (derivedActionFor + rewritten
  emit{Definition,Definitions}Decls)
- `goivy/ivy2cpp/constructors.go` (constructorActionFor +
  emitConstructor{,Decl}s)
- `goivy/logicutil.go` (Substitute cases for LogicNativeExpr +
  NativeCode)

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:1351-1375`
- `pyivy/ivy/ivy/ivy_to_cpp.py:1592-1625`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2308-2313`
- `pyivy/ivy/ivy/ivy_logic_utils.py:1572` (var_to_skolem)

## DONE 011 - Complete expression emission

Done 2026-05-21 (UTC).

Closed gaps:

- `*goivy.LogicLet` now expanded by substitution (`emitLetExpr` in
  `expr.go`), mirroring Python's pre-emit let-removal.
- Macro expansion via `goivy.IsMacro` / `goivy.ExpandMacro` runs at the top
  of `emitApply` (matches Python `il.is_macro` at `ivy_to_cpp.py:3124`).
- Range-result arithmetic (`+`, `-`, `*`, `/`, `%`) clamps to `[lb, ub]`
  using the canonical `( x < lb ? lb : ub < x ? ub : x )` form
  (`emitRangeArithApply`).
- `nat`-typed `-` saturates at 0 (`emitNatMinusApply`).
- `cast` operator handles non-bitvector targets — cast to `int`/`nat`/
  `RangeSort` (`emitCastApply`); the existing `emitBVApply` continues to
  handle BV destinations.
- `strlit`-interpreted constants: numeral `0` → `""`; other numerals
  return an actionable error (matches Python emit_constant lines
  3019–3023).
- `if some X. fmla minimizing|maximizing idx` lowered in
  `emitIfSomeMinMax` (`action.go`): scans candidates, tracks
  best-so-far index and witness, dispatches THEN/ELSE.
- `*goivy.LogicNamedBinder` returns a clear error rather than a generic
  "unsupported expression %T" diagnostic.

Tests added (in `goivy/ivy2cpp/ivy2cpp_test.go`):

- `TestEmitLetExpression`
- `TestEmitMacroExpansionInApply`
- `TestEmitNatSaturationOnMinus`
- `TestEmitRangeSaturationOnArithmetic`
- `TestEmitCastToRange`
- `TestEmitCastToNat`
- `TestEmitStringInterpConstantZero`
- `TestEmitStringInterpConstantNonzeroErrors`
- `TestEmitNamedBinderUnsupportedIsActionable`
- `TestEmitIfSomeMinimizing`
- `TestEmitIfSomeMaximizing`

Out of scope for this commit (tracked as follow-up):

- Native-literal `__lit<...>` wrappers via `is_native_sym`.
- `delegate_methods_to` / `delegate_enums_to` prefixing of method/enum
  references.
- Action terms embedded inside expressions and lambda-as-expression
  (no module currently produces these); the actionable error paths
  catch them rather than emitting partial C++.
- Parametric `let p(X,Y) := body in ...` substitution — returns an
  explicit error rather than wrong C++.

## DONE 012 - Port quantifier bounds, iterable sorts, and `some`/min/max

Done 2026-05-22 (UTC).

Closed gaps:

- `matchBoundExprs` walks the body collecting `<`, `<=`, `>`, `>=`
  applications that constrain a quantified variable, tracking polarity
  through `Not`, `LogicLiteral`, `Implies`, `Or`, `And`, and unfolding
  derived definitions (mirrors Python `get_bound_exprs`
  ivy_to_cpp.py:3264-3289).
- `getBounds` / `getAllBounds` synthesize `(lo, hi)` pairs from the
  collected inequalities, adding `"0"` for non-negative sorts, the
  sort cardinality, range-sort interpretations, and the
  `<sort>.cardinality` attribute as upper-bound fallback (mirrors
  Python `get_bounds` / `get_all_bounds` ivy_to_cpp.py:3301-3349).
  Sibling-variable filtering matches Python's `variables[i+1:]`
  slicing in `get_all_bounds`.
- `loopHeaderForSortBounds` emits the Python-conformant cast loop for
  enumerated sorts (`for (T X = (T)0; (int) X < N; X = (T)(((int)X)+1))`)
  and half-open form for integer sorts (`for (T X = lo; X < hi; X++)`),
  mirroring `open_loop` (ivy_to_cpp.py:1697-1711).
- `iterableSortFor` recognizes `<sort>.iterable` attributes and the
  paired `iter` / `iter.t` sort; `quantIterableHeader` emits the
  `for (T x = iter__create(0); !iter__is_end(x); x = iter__next(x))`
  loop and recurses into the body for the remaining variables
  (mirrors ivy_to_cpp.py:3405-3427).
- `emitQuant` now tries iterable → inequality bounds → extensional
  relation → finite-value fallback in the same order Python does.
- `emitSome` / `emitSomeWithElse` route through a shared
  `someLoopHeaders` helper that prefers inequality-derived bounds.
- `emitIfSome` / `emitIfSomeMinMax` route through `someConditionLoopHeaders`,
  which lifts the `*Const` parameters to `*LogicVariable` so the bound
  walker can see them (matches the `emitIfSomeExtensional` pattern).
- `firstParamIsIndex` emits the `break;` optimization when
  `some.Params[0] == some.Index` in `if some X. ... minimizing X`
  (mirrors Python emit_some:3539-3540 — the first hit during ascending
  iteration is the minimum, so the loop exits early).

Tests added (in `goivy/ivy2cpp/ivy2cpp_test.go`):

- `TestEmitQuantInequalityBoundOverRange`
- `TestEmitSomeMinMaxBreakWhenIndexIsFirstParam`
- `TestEmitSomeMinMaxNoBreakWhenIndexIsExpression`
- `TestEmitIfSomeUsesInequalityBound`
- `TestEmitQuantMultiVarInequalityFiltersSibling`

Pre-existing quantifier tests were updated to expect the
Python-conformant cast/half-open form instead of the prior Go
range-init / closed-range shorthand: `TestEmitExprQuantifierFiniteEnumLoop`,
`TestEmitExprSomeFiniteEnumLoop`, `TestEmitExprSomeWithElseFiniteEnumLoop`,
`TestGeneratedIfSomeActionCompiles`, `TestEmitExprQuantifierFiniteRangeLoop`,
`TestEmitIfSomeMinimizing`.

Out of scope for this commit (tracked as follow-up):

- Expression-context `SomeMin` / `SomeMax`. `goivy.LogicSome` lacks a
  `Kind` field, so only statement-context `if some` carries the min/max
  discriminator. A future TODO should add `Kind` to `LogicSome` and the
  parser side, then mirror Python's expression-context SomeMinMax in
  `emitSome`.

## DONE 013 - Fix assignment/update semantics, especially quantified assignments

Status: completed 2026-05-21. emitAssign now dispatches to
emitAssignSimple / emitAssignTwoPhase / emitAssignLarge mirroring Python
emit_assign (ivy_to_cpp.py:3703-3764). Quantified assignments use a
function-typed temporary so self-referential RHS reads see pre-assignment
values. The bexpr trick (Python ivy_to_cpp.py:3717-3720) tightens loops
when the RHS is Ite(cond, then, lhs) and cond does not mention the
modified symbol. Bounds-error paths fall back to emit_assign_large +
makeThunk, which emits a C++ thunk struct (local to the method body)
wrapped in hash_thunk. emitAssignField was refactored to synthesize a
LogicAssignAction and re-enter emitAssign so any future synthesizer
inherits the new dispatch. The Z3 / gen-mode to_z3 path in makeThunk is
stubbed and documented inline (Python ivy_to_cpp.py:538-602).

Files added / modified:

- `goivy/ivy2cpp/assign.go` (new) — emitAssignSimple, emitAssignTwoPhase,
  assignBoundsExpr, openAssignmentLoopsBounded,
  canOpenAssignmentLoopsBounded.
- `goivy/ivy2cpp/thunk.go` (new) — makeThunk, emitAssignLarge,
  emitThunkBody, thunkEnvSymbols.
- `goivy/ivy2cpp/action.go` — emitAssign refactor; emitAssignField now
  routes through emitAssign.
- `goivy/ivy2cpp/generator.go` — thunkCtr field on Generator.
- `goivy/ivy2cpp/assign_test.go` (new) — eight new tests covering
  self-referential, multi-variable transpose, bexpr tightening,
  bexpr-skip-on-modified, thunk fallback, field-action routing,
  scalar-simple, and a Python oracle check for the flip case.
- `goivy/ivy2cpp/ivy2cpp_test.go` — updated TestEmitAfterInitEnumLoop,
  TestEmitAfterInitRangeLoop, TestRangeArrayDimensionUsesUpperBoundIndexSpace,
  and the smoke fixtures table entry to expect the new two-phase shape.

## DONE 014 - Correct havoc, local variables, choices, and old-value binding

Status: completed 2026-05-22. All four gap items closed:

- `emitHavoc` (`goivy/ivy2cpp/action.go:97-103`) mirrors Python `emit_havoc`'s
  `assert False` (`ivy_to_cpp.py:3768-3773`) by reporting via
  `g.unsupported` — havoc reaching emit is a bug because lowering should
  have eliminated it upstream.
- `emitLocal` (`action.go:785-796`) declares the local with
  `cppStorageDecl` and then calls `mkNondetSym` with the LocalAction's
  UniqueID, matching Python `local_start` + `emit_local`
  (`ivy_to_cpp.py:3893-3917`).
- `emitChoice` uses `mkNondet(..., "___branch", a.UniqueID, ...)` with
  the same hardcoded-0 quirk as Python `ivy_to_cpp.py:189`.
- `emitBindOlds` (`action.go:835-`) reports unsupported because Python
  has no `BindOldsAction.emit`; the wrapper must be eliminated by
  `bind_olds_action` upstream (`ivy_transrel.py:240`).

Coverage: `TestGeneratedChoiceActionCompiles`,
`TestGeneratedVarActionCompiles`, `TestGeneratedHavocReportsUnsupported`,
`TestGeneratedChoiceUsesIfElseChain`,
`TestGeneratedLocalActionUsesNondet`,
`TestGeneratedLocalFunctionUsesNondetLoop`.

Go locations:

- `goivy/ivy2cpp/action.go:14-73`
- `goivy/ivy2cpp/action.go:84-100`
- `goivy/ivy2cpp/action.go:419-429`
- `goivy/ivy2cpp/action.go:461-467`
- `goivy/ivy2cpp/generator.go:229-234`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:185-215`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3768-3773`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3893-3917`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3976-3996`

Gap:

- Go implements havoc by assigning zero values. Python's `HavocAction.emit`
  currently emits an assertion failure; Python's nondeterministic values are
  introduced through explicit nondet helpers such as local variables, initial
  state, and choose paths.
- Go local variables are zero-initialized. Python local variables without an
  initializer receive nondeterministic values via `mk_nondet`.
- Go choice actions use `___ivy_choose`, but `___ivy_choose` always returns `0`
  in the base class.
- Go `emitBindOlds` ignores old-value binding semantics.

Conformance work:

- Match Python's exact havoc behavior, even if that means emitting the same
  assertion failure until a broader nondet semantics is intentionally changed.
- Use Python's nondet helper logic for uninitialized locals and other
  nondeterministic symbols.
- Port Python's choice chain and the target-specific `___ivy_choose`
  implementation.
- Implement old-value binding by capturing old expressions before inner action
  execution, according to Python semantics.

Tests to add:

- Actions with local variables, choice actions, explicit havoc, and `old`
  references in postconditions/updates.

## DONE 015 - Port native declarations, native code blocks, and callback thunks

Status: completed 2026-05-22. `goivy/ivy2cpp/native.go` (636 lines) ports
`split_native` (`ivy_to_cpp.py:1378`), `emit_native`
(`ivy_to_cpp.py:1456`), and `native_reference` (`ivy_to_cpp.py:4032`),
covering `header`, `impl`, `member`, `init`, `inline`, and `encode`
tags plus antiquote substitution for `%`, `"`, symbol references, type
references, and action callbacks. `goivy/ivy2cpp/native_thunk.go` ports
the callback-thunk struct generation via `emitCallbackThunks` /
`emitCallbackThunk` (Python lines 4032-4074). Native type declarations
emit either typedef-style aliases or class-style wrappers, and the
test/gen paths consume them through the existing `_arg` / serializer /
solver helpers. Coverage: 20+ tests including
`TestEmitNativeActionAntiquotes`,
`TestNativePrimitiveTypeDeclarationFromInterpret`,
`TestNativeTypeAntiquoteReferencesSort`,
`TestNativeClassParameterDefaultCompiles`,
`TestNativeDefinitionEmitsTemplateMethod`,
`TestTopLevelNativeBlocksEmitHeaderMemberAndInit`,
`TestDuplicateOnceNativeHeaderEmitsOnce`.

Go locations:

- `goivy/ivy2cpp/native.go:40-66`
- `goivy/ivy2cpp/native.go:97-147`
- `goivy/ivy2cpp/native.go:244-273`
- `goivy/ivy2cpp/native.go:315-325`
- No equivalent callback-thunk scan found in Go.

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:1378-1459`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2187-2200`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2314-2328`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2395-2404`
- `pyivy/ivy/ivy/ivy_to_cpp.py:4032-4074`

Gap:

- Go parses native blocks and maps `once` to `header`, but it does not implement
  the full Python tag set and validation behavior.
- Go has partial antiquote support for `%`, `"`, and basic symbol references,
  but Python supports type substitutions, references, callback thunks, native
  action invocation, and target-specific placement.
- Go detects action references in `nativeTypeOf` as `thunk__name`, but it never
  emits the thunk structs/classes Python emits.
- Go native type declarations are just typedef-like members. Python also emits
  hash functions, constructors, solver converters, parsers/writers, and native
  member initialization logic where needed.

Conformance work:

- Port `split_native`, `translate_ref`, `emit_native_refs`, `emit_native`,
  `native_type`, and native-action emission.
- Add callback thunk struct generation and invocation for all native references
  to exported actions.
- Match Python's `header`, `impl`, `member`, `init`, `inline`, and encode tags,
  including errors for unknown tags.
- Generate native type support methods used by REPL, serializers, Z3, and
  hash-thunk storage.

Tests to add:

- Native header/member/init/inline snippets, native action callbacks, native
  types used in state, and native expressions with antiquoted references.

## DONE 016 - Replace the simple REPL with Python's REPL/server/test runtime

Status: completed 2026-05-22. `goivy/ivy2cpp/repl.go` (643 lines) and
`goivy/ivy2cpp/runtime.go` (415 lines) port Python's REPL/test/server
runtime: `emit_repl_boilerplate1a` builds the per-classname
`cmd_reader` subclass; `_arg<T>` overloads parse every emitted sort
(enums, ranges, numerics, destructors, variants, natives, bv/strings)
out of `ivy_value`; main-loop server vs. REPL branching mirrors
`emit_repl_boilerplate3` / `emit_repl_boilerplate3server`. Coverage:
`TestRuntimeSkeletonAcrossTargets`,
`TestRuntimeReplAndTestSubclassGlue`,
`TestRuntimeChoiceStackAndGeneratorPlumbing`,
`TestRuntimeNativeReaderTimerSkeletonShape`,
`TestReplDispatchForExportedAction`,
`TestReplMainReadsCommandsFromStdin`,
`TestReplIgnoresInternalAction`,
`TestReplDispatchForParameterizedExportCompiles`,
`TestReplDispatchParsesEnumBoolRangeArgs`,
`TestReplEmitsCmdReader`,
`TestReplCatchesSyntaxOutOfBoundsBadArity`,
`TestReplServerModeWhenNoPublicActions`, and several
`TestReplWrites…SupertypeReturn` cases for variant-typed returns.

Go locations:

- `goivy/ivy2cpp/generator.go:679-743`
- `goivy/ivy2cpp/repl.go:10-280`
- `goivy/ivy2cpp/generator.go:745-798`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:2676-2852`
- `pyivy/ivy/ivy/ivy_to_cpp.py:4079-4135`
- `pyivy/ivy/ivy/ivy_to_cpp.py:4137-4187`
- `pyivy/ivy/ivy/ivy_to_cpp.py:4189-4225`
- `pyivy/ivy/ivy/ivy_to_cpp.py:4228-4467`

Gap:

- Go emits a minimal `main` loop that reads a line, splits whitespace, and
  dispatches action names.
- Python emits a substantial REPL/test/server runtime around `ivy_value`,
  `ivy_repl`, parser errors, prompt/output handling, action lock management,
  out/modelfile streams, delays, wait commands, imported callbacks, readers,
  timers, Windows socket initialization, and target-specific action wrappers.
- Go's parsers/writers cover only simple bool/enums/ranges/numeric values.
  Python supports structs, variants, arrays/functions, native types, bitvectors,
  serializers, and rich error messages.

Conformance work:

- Port Python REPL class generation and command-reader logic.
- Replace ad hoc whitespace splitting with Python-compatible `ivy_value`
  parsing and `_arg` conversion methods for every type.
- Add output streams, prompt, errors, wait/delay, action lock/unlock, reader and
  timer dispatch, server mode, and test harness behavior.
- Use generated serializers/deserializers and parser helpers for complex values.

Tests to add:

- REPL parse/print for each type class.
- Action dispatch with inputs/outputs, bad arity, bad value syntax, wait/delay,
  reader callbacks, and timer callbacks.

## DONE 017 - Port action generation and solver-backed random testing

Go locations:

- `goivy/ivy2cpp/z3.go:11-23`
- `goivy/ivy2cpp/z3.go:234-282`
- `goivy/ivy2cpp/z3.go:383-494`
- `goivy/ivy2cpp/generator.go:745-798`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:905-977`
- `pyivy/ivy/ivy/ivy_to_cpp.py:979-1006`
- `pyivy/ivy/ivy/ivy_to_cpp.py:1027-1046`
- `pyivy/ivy/ivy/ivy_to_cpp.py:1068-1091`
- `pyivy/ivy/ivy/ivy_to_cpp.py:1093-1159`
- `pyivy/ivy/ivy/ivy_to_cpp.py:1161-1175`
- `pyivy/ivy/ivy/ivy_to_cpp.py:1197-1348`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2410-2418`

Gap:

- Go `emitZ3Randomize` randomizes state/helper variables independently. It does
  not encode current state, action preconditions, transition constraints, or
  solve for a preimage/model.
- Go action generator classes simply randomize inputs and call actions.
- Python's generator path builds reverse-image constraints, extracts defined
  inputs, expands field references, handles preconditions and before-export
  constraints, initializes solver variables, evaluates model results, records
  traces, handles action weights, and integrates with the gen/test runtime.

Conformance work:

- Port `emit_init_gen`, `emit_randomize`, `extract_defined_parameters`,
  `emit_defined_inputs`, field extraction, `expand_field_references`, and
  `emit_action_gen`.
- Encode state before solving and assign model values back into C++ state after
  solving.
- Respect preconditions, assumptions, before-export constraints, extensional
  relation constraints, field references, generated inputs, and action weights.
- Integrate with the same gen/test loop and stack/reporting code as Python.

Tests to add:

- Random test generation for actions with preconditions, derived fields,
  extensional relations, defined inputs, and output constraints.
- Compare generated traces against Python for deterministic seeds where
  possible.

## DONE 018 - Complete Z3 sort/declaration/eval/set conversion

Landed in four milestones (M1–M4); plan at
`~/.claude/plans/we-are-in-ivy-goivy-ivy2cpp-wise-wombat.md`.

Go locations:

- `goivy/ivy2cpp/solver_emit.go:emitSetField` (M1)
- `goivy/ivy2cpp/solver_emit.go:emitSetSolver` branch (2) + `isLargeType` (M2)
- `goivy/ivy2cpp/z3.go:emitZ3EnumSolverConversion` + `repl.go:emitEnumSortArgSpecDecls` (M3)
- `goivy/ivy2cpp/solver_emit.go:emitFromSolverLoop` + `isRecordRange` / `recordRangeType` (M4.b)
- `goivy/ivy2cpp/z3.go:emitZ3SortRegistrations` (M4.c — runtime-equivalence comment)

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:687-729` (`emit_sorts`)
- `pyivy/ivy/ivy/ivy_to_cpp.py:730-756` (`emit_decl`)
- `pyivy/ivy/ivy/ivy_to_cpp.py:772-799` (`emit_eval`)
- `pyivy/ivy/ivy/ivy_to_cpp.py:806-824` (`emit_set_field`)
- `pyivy/ivy/ivy/ivy_to_cpp.py:826-870` (`emit_set`)
- `pyivy/ivy/ivy/ivy_to_cpp.py:2654-2668` (enum-sort `__from_solver`/`__to_solver`/`__randomize`)
- `pyivy/ivy/ivy/ivy_to_cpp.py:4469-4480` (`emit_boilerplate1`)

What landed:

- **M1.** Ported `emit_set_field` faithfully. `emitSetSolver` branch (1)
  restructured to open loops per-destructor matching Python's flow; the
  recursive `add(__to_solver(*this, apply("<destr>", lhs, ...), rhs[idx...].field))`
  call is now emitted instead of the previous unsupported-marker stub.
- **M2.** Added the `forall`-quantified branch in `emitSetSolver` for
  function-sorted symbols whose domain `is_large_type` (non-integer
  domain element or product > 1024). Added `Generator.isLargeType`
  mirroring Python ivy_to_cpp.py:445-449.
- **M3.** Converted enum-sort solver conversion from `static` overloads
  to `template <>` specializations matching Python, delegating to
  `__from_solver<int>` / `__to_solver<int>` / `__randomize<int>`. Added
  matching forward declarations gated under `#ifdef Z3PP_H_`.
- **M4.a.** No-op — Python's `__pto__<dom0>__<dom1>` cname at
  ivy_to_cpp.py:735 is dead code (computed but never referenced); Python
  emits `*>` literally and Go already matches.
- **M4.b.** Split `emitFromSolverLoop` into the Python-faithful branches:
  destructor / native / cpp-interp ranges → `__from_solver<class::T>
  (*this, apply("name", ...), lvalue);`; primitive ranges → `lvalue =
  (ctype)eval_apply("name", ...);`. Added `recordRangeType` so cpp-typed
  bv / strbv / intbv sorts route to `<class::word>` rather than decaying
  to `<unsigned>`. `emitZ3EvaluateStateSymbol` now delegates to the
  shared helper.
- **M4.c.** Documented runtime divergence at `emitZ3SortRegistrations`:
  Python's `enum_sorts.insert(name, <class>::<sortvar>::z3_sort(ctx))`
  is functionally equivalent to the Go runtime's `mk_bv(name, bits)` /
  `mk_string(name)` — same Z3 sort constructors underneath, but
  `ivy_go_z3.hpp` uses a single `sorts` map rather than Python's separate
  `enum_sorts` map. The `z3_sort(ctx)` static method emitted by
  `cpp_types.go` is dead on the Go runtime path; mk_bv / mk_string is the
  runtime-supported equivalent.

Tests added:

- `TestEmitSetSolverDestructorRecordRange` — white-box assertion of the
  nested `add(__to_solver(...))` recursion (M1).
- `TestEmitSetSolverLargeTypeForall` — asserts the `forall(__quants, ...)`
  emission for a non-integer-domain state symbol (M2).
- `TestEnumSortSolverSpecsAreTemplateSpecializations` — asserts the new
  `template <>` form for enum sorts, including forward decls, and that
  the legacy `static` overloads are gone (M3).
- `TestEmitEvalBranchesByRangeKind` — asserts both eval branches:
  primitive `(ctype)eval_apply(...)` cast and cpp-typed
  `__from_solver<class::T>(*this, apply(...), x)` (M4.b).

`make test` green after each milestone.

## DONE 019 - Match Python progress/rely logic

Status: completed 2026-05-22. `emitRelyMax` (`goivy/ivy2cpp/tick.go:265-323`)
now alpha-renames extra rely-RHS variables by appending `__` and rewrites the
LHS-aligned progress args in a single substitution pass, matching Python
`ivy_to_cpp.py:1798-1804`. Bare relies are filtered through `hasBareRely`,
`ivy_check_progress(lhs, maxt)` is called for each progress declaration
(`tick.go:251`), and the per-target loop bounds use the same iterable-sort
machinery as Python `__tick`. Coverage:
`TestTickCallsIvyCheckProgressWithoutRely`,
`TestTickRelyImplicationComputesMaxAndChecksProgress`,
`TestTickRelySubstitutesProgressArgs`,
`TestTickRelyExtraFreeVariableLoops`,
`TestTickUnconditionalRelySkipsProgressCheckLikePython`,
`TestTickRelyExtraSharesProgressVarNameRenamed`.

Go locations:

- `goivy/ivy2cpp/tick.go:24-127`
- `goivy/ivy2cpp/tick.go:129-136`
- `goivy/ivy2cpp/tick.go:200-228`
- `goivy/ivy2cpp/tick.go:237-275`
- `goivy/ivy2cpp/generator.go:224-227`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:1749-1825`
- `pyivy/ivy/ivy/ivy_to_cpp.py:1798-1804`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2295-2347`

Gap:

- Go's tick generation resembles Python's shape, but it does not alpha-rename
  extra rely variables the way Python does.
- Go skips bare rely checks in a simplified way and identifies progress/rely
  definitions by expression-name heuristics.
- Go `ivy_check_progress` is empty, so generated counters are not enforced in
  the same runtime path as Python.
- Initialization of progress counters and interaction with generated
  constructor/init code is incomplete.

Conformance work:

- Port Python's progress/rely scan and variable handling, including the
  alpha-renaming of extra variables.
- Implement `ivy_check_progress` and any generated counter members exactly as
  Python.
- Ensure `__tick` includes the same loop bounds and max/zero behavior as Python
  for int/nat/range counters.

Tests to add:

- Progress properties with extra quantified variables, rely formulas, and
  counters requiring max updates.

## DONE 020 - Handle requires/ensures/subgoals and exported action semantics exactly

Status: completed 2026-05-22. Five focused fixes landed:

- `goivy/ivy2cpp/action_gen.go:66-78` — `ext_preconds` wrap now preserves
  `Lineno`, formal params, and formal returns on the new
  `Sequence(AssumeAction(pre), action)` (Python `ivy_to_cpp.py:1213-1217`).
- `goivy/ivy2cpp/action.go:200-208` — added `linenoStr` mirroring
  `iu.lineno_str` (strips the trailing `": "` that `Location.String()`
  appends) and used it for the assert/assume/requires/ensures/subgoal
  emit sites at `action.go:30-38` so labels match Python
  `ivy_to_cpp.py:3788-3808`.
- `goivy/ivy2cpp/generator.go:870-940` — added `importCallers()` and
  `emitTraceActionPrologue()` mirroring Python `find_import_callers`
  (`ivy_to_cpp.py:1888-1897`) and `trace_action`
  (`ivy_to_cpp.py:1576-1607`). For `target=test`, imported unscoped
  actions now open their method body with
  `__ivy_out << "< name(args)" << std::endl;`. Cached on `Generator`.
- `goivy/actions_transforms.go:71-86` — `AssertToAssume` no longer
  downgrades `LogicSubgoalAction` when `"assert"` is in kinds. Python's
  class-identity check (`ivy_actions.py:396-403` + `:416-421`) only
  converts when the caller explicitly opts in via `"subgoal"`.

Coverage: `TestActionGenExtPrecondsPreservesFormals`,
`TestAssertLabelStripsTrailingColonSpace`,
`TestSubgoalNeverConvertedToAssume`,
`TestRequiresInExternalBecomesAssume`,
`TestImportCallerTracePrologueInTest`,
`TestExtPrecondsAppearsInActionGenSMT`. Full `make test` green.

Go locations:

- `goivy/ivy2cpp/action.go:31-36`
- `goivy/ivy2cpp/action.go:379-417`
- `goivy/ivy2cpp/z3.go:383-494`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:1479-1517`
- `pyivy/ivy/ivy/ivy_to_cpp.py:1197-1348`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3788-3808`

Gap:

- Go emits `Requires`, `Ensures`, and `Subgoal` actions as direct runtime
  `ivy_assert` calls.
- Python action annotation and action-generator code use requires/ensures in
  several contexts: method declarations, exported action constraints, generated
  preconditions, and trace/test behavior.
- Go test/gen code does not include Python's `ext_preconds`,
  `before_export`/`after_export` handling, or exported-action bookkeeping.

Conformance work:

- Port Python's action annotation and exported-action constraint handling.
- Ensure runtime assertion/assumption choices match Python for each action
  wrapper and target.
- Feed requires/ensures/subgoals into action generation exactly like Python,
  instead of treating all of them as simple body assertions.

Tests to add:

- Exported actions with requires, ensures, assumptions, subgoals, and generated
  tests that must satisfy preconditions before execution.

## DONE 021 - Fix debug/trace output semantics

Status: completed 2026-05-22. Two layered fixes plus tests.

**emit_debug semantics.** `LogicDebugAction` now carries a `WithNames`
slice (`goivy/actions_phase3.go:494`) populated by `CompileDebugAction`
(`goivy/compiler_phase6.go:818-823`) so each `with name = expr` clause
keeps its name through compilation. `goivy/ivy2cpp/action.go:emitDebug`
emits each value under its real name and runs the new `emitPrintExpr`,
which opens a loop over the value's free variables and wraps the print
with `std::cout << "["` / `"]"` per Python `emit_print_expr`
(`ivy_to_cpp.py:3989-3996`).

**Trace integration.** Added `Generator.numberFormat()`
(`goivy/ivy2cpp/generator.go`) returning ` << std::hex << std::showbase`
when the module attribute `radix == "16"` (Python `ivy_to_cpp.py:1935-
1938`). Threaded into `emitTraceActionPrologue`,
`emitRuntimeReplAssertOverride`, `emitTracePrelude`, and the REPL
dispatch close-brace line. `emitActionGenExecute`
(`goivy/ivy2cpp/action_gen.go`) now matches Python lines 1331-1346:
emits the `> name(args)` trace line, opens `{` and closes `}` when
`Config.Trace`, and prints `= __res` for single-return actions.
`emitSomeAction` (`goivy/ivy2cpp/generator.go`) wraps the imported
action body with `{` / `}` braces when `Config.Trace` is on.
`emitAssignSimple` (`goivy/ivy2cpp/assign.go`) emits the
`__ivy_out << "  write(<lhs>," << (<rhs>) << ")"` line under
`Config.Trace`, gated by the same `':' not in name` rule as Python
(`ivy_to_cpp.py:3627`).

Coverage: `TestDebugActionEmitsNamedValues`,
`TestDebugActionEmitsQuantifiedLoop`,
`TestNumberFormatHexFromRadixAttribute`,
`TestActionGenExecuteWrapsTraceBraces`,
`TestImportCallerBodyWrappedInBracesUnderTrace`,
`TestAssignSimpleEmitsWriteTraceUnderTrace`, plus the updated
`TestGeneratedDebugActionCompiles`. Full `make test` green.

Go locations:

- `goivy/ivy2cpp/action.go:550-563`
- `goivy/ivy2cpp/generator.go:13-22`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:3998-4026`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2262-2293`

Gap:

- Go debug emission prints generic keys such as `value0` and `value1`.
- Python prints the supplied debug names, handles quantified debug values, and
  integrates trace output with generated choose/action stacks.
- Go's `Trace` config is not wired into all the Python places that use tracing.

Conformance work:

- Port Python's `DebugAction.emit` behavior, including named values and
  quantified output loops.
- Wire trace output into choose, call, gen/test, REPL, and action wrappers like
  Python.

Tests to add:

- Debug actions with explicit names, multiple values, quantified values, and
  tracing enabled under `impl`, `repl`, `test`, and `gen`.

## DONE 022 - Add full serialization/deserialization support

Status: ported field-for-field from Python. Per-sort `operator<<`, `_arg<T>`,
`__ser<T>`, `__deser<T>` are emitted for enum, destructor, and variant sorts.
Z3 `__from_solver`/`__to_solver`/`__randomize` specializations are emitted for
gen/test targets. StrBV/IntBV interpreted-type templates already covered the
bv/strbv/intbv sort kinds. Three gaps closed in this pass:

1. **Destructor sort forward declarations.** `emitDestructorSortArgSpecDecls`
   in `goivy/ivy2cpp/destructor.go:258` mirrors Python `ivy_to_cpp.py:2232-2254`
   and is invoked from `goivy/ivy2cpp/runtime.go:148` right after the enum
   forward decls.
2. **Zero-init for destructor `_arg<T>`.** `emitDestructorZeroInit` and
   `emitZeroAssign` in `goivy/ivy2cpp/destructor.go:487, 513` mirror Python
   `assign_zero_symbol` (`ivy_to_cpp.py:216-219`). They are called from
   `emitDestructorArgImpl` at `goivy/ivy2cpp/destructor.go:406` right after the
   `T res;` declaration. Strlit, native, cpptype, and variant-super fields are
   skipped per the Python guard.
3. **`to_solver_class<hash_thunk<D,R>>` specializations.** `allHashThunkDomains`
   in `goivy/ivy2cpp/types.go:612` mirrors Python `all_hash_thunk_domains`.
   `emitHashThunkToSolver` and `emitAllCtuplesToSolver` in
   `goivy/ivy2cpp/solver_emit.go:736, 770` mirror Python
   `emit_hash_thunk_to_solver` / `emit_all_ctuples_to_solver`
   (`ivy_to_cpp.py:1841-1871`), and the wire-in is at
   `goivy/ivy2cpp/generator.go:255` inside the existing `usesZ3()` block.
   The runtime primary `to_solver_class<T>` template and the
   `z3_thunk<D,R>` abstract subclass of `thunk<D,R>` are now declared in
   `include2cpp/ivy_go_z3.hpp` (lines 379-396).

Go locations:

- Enum impls: `goivy/ivy2cpp/repl.go:60-160`; decl wire at `goivy/ivy2cpp/runtime.go:145`.
- Destructor impls: `goivy/ivy2cpp/destructor.go:244-517`.
- Destructor forward decls: `goivy/ivy2cpp/destructor.go:258` (`emitDestructorSortArgSpecDecls`);
  wire at `goivy/ivy2cpp/runtime.go:148`.
- Destructor zero-init: `goivy/ivy2cpp/destructor.go:487, 513`
  (`emitDestructorZeroInit`, `emitZeroAssign`).
- Variant impls: `goivy/ivy2cpp/variant.go:122-310`.
- StrBV/IntBV impls: `goivy/ivy2cpp/cpp_types.go:272-365`.
- Hash-thunk to_solver: `goivy/ivy2cpp/solver_emit.go:736, 770`
  (`emitHashThunkToSolver`, `emitAllCtuplesToSolver`); domain enumerator at
  `goivy/ivy2cpp/types.go:612` (`allHashThunkDomains`); wire at
  `goivy/ivy2cpp/generator.go:255`.
- Runtime template support: `include2cpp/ivy_go_z3.hpp:379-396` (`to_solver_class`
  primary template, `z3_thunk` abstract class).

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:2213-2254` (forward decls).
- `pyivy/ivy/ivy/ivy_to_cpp.py:216-219` (`assign_zero_symbol`).
- `pyivy/ivy/ivy/ivy_to_cpp.py:1841-1871, 2673` (`emit_hash_thunk_to_solver`,
  `emit_all_ctuples_to_solver`).
- `pyivy/ivy/ivy/ivy_to_cpp.py:2420-2674` (per-sort impls).
- `pyivy/ivy/ivy/ivy_cpp_types.py:355-496` (variant/StrBV/IntBV template emission).
- `pyivy/ivy/ivy/ivy_z3_helpers.hpp:42-44, 135-138` (`to_solver_class`, `z3_thunk`).

Tests:

- Existing: `TestEmitsArgSpecForEnum` (ivy2cpp_test.go:3885),
  `TestDestructorSerDeserShape` (5852), `TestDestructorArgShape` (5881),
  `TestDestructorZ3ImplShape` (5909),
  `TestDestructorRandomizeSkipsUninterpretedRange` (5940).
- New: `TestDestructorForwardDeclsInImplPreamble`,
  `TestDestructorZ3ForwardDeclsInImplPreamble`,
  `TestDestructorArgZeroInitsPrimitiveFields`,
  `TestDestructorZeroInitSkipsStringField`,
  `TestHashThunkToSolverSpecializationEmittedForTestTarget`,
  `TestHashThunkToSolverNotEmittedForReplTarget`
  (all at ivy2cpp_test.go:5974+).

## DONE 023 - Restore Python include ordering and support headers

Status: header and impl preambles now mirror Python's emission. Three gaps
closed:

1. **Windows host preamble.** `emitRuntimeHeaderPreamble` in
   `goivy/ivy2cpp/runtime.go:26` emits `#define WIN32_LEAN_AND_MEAN` then
   `#include <windows.h>` before all other includes when the build host
   is Windows (mechanical port of Python `ivy_to_cpp.py:1949-1951` —
   `platform.system() == 'Windows'`). Host detection lives in
   `hostOS()` at `runtime.go:19`, which reads the new `Config.HostOS`
   field (`generator.go:14` `Config`) with `runtime.GOOS` fallback so
   tests can exercise both branches on any host.
2. **`_HAS_ITERATOR_DEBUGGING` define.** Same site, unconditional, ahead
   of all `#include` lines — mirrors Python `ivy_to_cpp.py:1952`.
3. **Z3 helper include moved to the impl preamble.** Go's
   `ivy_go_z3.hpp` is now emitted at `goivy/ivy2cpp/runtime.go:157`
   immediately after `#include "ivy_value.hpp"`/`#include "ivy_repl.hpp"`
   and before any inline `__from_solver`/`__to_solver`/`__randomize`
   template definitions — mirroring Python `ivy_to_cpp.py:2210-2211`
   (`ivy_z3_helpers.hpp`). The duplicate emission inside
   `emitZ3Runtime` (formerly `z3.go:30-33`) is removed; that helper
   is gone and `emitZ3Support` in `goivy/ivy2cpp/z3.go:12` no longer
   calls it.

Deliberate divergence (documented, not changed in this TODO): Go keeps
`ivy_threads.hpp` in the *header* rather than the impl preamble (where
Python places it at `ivy_to_cpp.py:2068`) because the generated Go class
body declares `std::vector<HANDLE>` / `std::vector<pthread_t>` members
directly — `pthread_t` must be visible at class-definition time. Existing
test `TestGoOutputUsesSharedSupportIncludes` pins this placement and
exercises a real C++ compile via `compileGeneratedCPP`.

Go locations:

- `goivy/ivy2cpp/runtime.go:19` (`hostOS()`).
- `goivy/ivy2cpp/runtime.go:26-67` (`emitRuntimeHeaderPreamble`).
- `goivy/ivy2cpp/runtime.go:119-174` (`emitRuntimeImplPreamble`); Z3
  helper include at line 157.
- `goivy/ivy2cpp/generator.go:14` (`Config.HostOS` field).
- `goivy/ivy2cpp/z3.go:12-25` (`emitZ3Support` no longer emits the
  runtime include).

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:1947-1973` (header preamble).
- `pyivy/ivy/ivy/ivy_to_cpp.py:2030-2069` (impl preamble).
- `pyivy/ivy/ivy/ivy_to_cpp.py:2206-2211` (late impl support includes).

Tests:

- Existing (pre-existing constraints preserved):
  `TestGoOutputUsesSharedRuntimeIncludes`,
  `TestGoOutputUsesSharedSupportIncludes` (+ `compileGeneratedCPP`),
  `TestGoOutputUsesSharedZ3RuntimeIncludes` (now also pins
  `ivy_go_z3.hpp` to impl-only).
- New: `TestHeaderPreambleEmitsIteratorDebuggingDefine`,
  `TestHeaderPreambleEmitsWindowsHostIncludesWhenHostOSIsWindows`,
  `TestHeaderPreambleSkipsWindowsHostIncludesByDefaultOnNonWindowsHost`,
  `TestImplPreambleEmitsZ3HelperIncludeAfterReplInclude`,
  `TestImplPreambleOmitsZ3HelperIncludeForReplTarget`.

Future work (separate TODOs): feature gates for sockets/timers/REPL/
serialization/native snippets become meaningful only once those runtime
features land. Python currently emits the platform-specific
network/socket headers unconditionally in the impl preamble and Go
already mirrors that.

## DONE 024 - Match Python's member-name collision checks

Status: ported `check_member_names` faithfully. `(*Generator).checkMemberNames`
at `goivy/ivy2cpp/generator.go:185` collects the varName-lowered names of all
signature symbols (`g.Mod.Sig.Symbols.All()`), sorts (`g.Mod.Sig.Sorts.All()`),
and actions (`g.Mod.Actions.All()`), and returns an `ivy2cpp:` error when the
generated C++ class name appears among them. The check runs at the top of
`(*Generator).generate` at `generator.go:169`, after `prepareModuleForCPP`
in `Generate` (`generator.go:135`) has had its chance to register
`_generating` for the test target. The error message mirrors Python's
two-line text — including the `Use command line option classname=...` hint —
so user-facing diagnostics match the Python toolchain.

Reused utilities:

- `varName` (`goivy/ivy2cpp/names.go:15`) — already the Python `varname`
  mirror; accepts string keys via `fmt.Sprint`.
- Map iteration via `.Symbols.All()` / `.Sorts.All()` / `.Actions.All()` —
  the idiom used throughout this package.

Go locations:

- `goivy/ivy2cpp/generator.go:169` (call site at top of `generate`).
- `goivy/ivy2cpp/generator.go:181-210` (`checkMemberNames` method, with
  Python-origin comment block).

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:1830-1834` (`check_member_names`).
- `pyivy/ivy/ivy/ivy_to_cpp.py:1904` (call site inside `module_to_cpp_class`).

Tests:

- New (`goivy/ivy2cpp/ivy2cpp_test.go:2395-2442`):
  - `TestCheckMemberNamesRejectsActionCollision`
  - `TestCheckMemberNamesRejectsSymbolCollision`
  - `TestCheckMemberNamesRejectsSortCollision`
  - `TestCheckMemberNamesAllowsDistinctClassname` (regression guard against
    the check firing on legitimate inputs).

Out of scope (potential future TODO): Python's `check_member_names` itself
only guards the *classname* against module-declared names. Reserved C++
keywords, runtime helper names, lock/stream/solver names, etc. are not
checked by Python either; broadening to those would be net-new behaviour
and belongs in its own audit item.

Reminder:

- [x] When this lands, rename to `## DONE 024 - …`, add a `Status:` paragraph
  citing the new Go locations and test names, and update this audit doc in
  the same commit as the implementation.

## DONE 025 - Port conjecture/property/isolate integration

Status: ported the per-session flag setup, isolate-specific module
preparation, and lineno-preserving conjecture insertion from Python's
`main_int` (`pyivy/ivy/ivy/ivy_to_cpp.py:4513-4645`) and
`add_conjs_to_actions` (`pyivy/ivy/ivy/ivy_to_cpp.py:4495-4503`).

The new `applySessionParameters` helper
(`goivy/ivy2cpp/compile.go:applySessionParameters`) runs once per
session and mirrors Python lines 4514-4530: `SetDeterminize(true)`,
`SolverOpts.UseZ3Enums=true`, `IsolateCfg.InterpretAllSorts=true`,
`SetVerifyingOnMod(false)` (auto-initializing `CompCfg` when nil), and
the `IsolateCfg` toggles `ConeOfInfluence=false`, `CreateImports=true`,
`EnforceAxioms=true`, `AssumeInvariants=false`, plus the
target-dependent `IsolateMode` and `FilterSymbols`/`KeepDestructors`
choices. It is wired into both `CompileAndGenerateAll` (before
`goivy.SourceFile`) and `Generate` (after config normalization), so
direct-Generate callers see the same session state.

Per-isolate setup inside the `CompileAndGenerateAll` loop now mirrors
Python 4612-4622: for `target=repl` in language ≥1.7, a non-extract
isolate named on the command line is rewritten in place to an extract
(`iso.Kind="extract"; iso.WithArgs=len(iso.Elems)`); and
`isoMod.Cfg.IsolateCfg.CompileWithInvariants` is set to true exactly
when `target=="test" && languageVersionAtLeast(isoMod,"1.7")`.

Early `_generating` registration (Python 4550-4551) happens before
`goivy.SourceFile` for the `test` target so cone-of-influence sees the
symbol; the existing late add in `prepareModuleForCPP` stays as a
safety net.

`addConjsToActions` (`goivy/ivy2cpp/compile.go`) now calls
`a.SetLineno(conj.GetLineno())` on each appended `LogicAssertAction`,
mirroring Python's `set_lineno(conj.lineno)`. The line number flows
through the existing `linenoStr` formatter (`action.go:200-207`) into
the emitted `ivy_assert(..., "<file>: line N")` label.

Reused utilities:

- `goivy.SetDeterminize` (`goivy/actions_phase3.go:467`).
- `goivy.SetVerifyingOnMod` (`goivy/module_compiler_config.go:35`)
  with `goivy.NewCompilerConfig` (`goivy/module_compiler_config.go:12`).
- `(*IsolateDef).IsExtract` (`goivy/ast_decl_ast.go:1552`) for
  discrimination; in-place mutation of `Kind`/`WithArgs` matches the
  Python `im.module.isolates[isolate] = the_iso` rebind.
- `(*ActionBase).SetLineno` (`goivy/module_action.go:53`) and the
  `*LabeledFormula` embedded `Base.GetLineno()` (`goivy/ast.go:292`).

Go locations:

- `goivy/ivy2cpp/compile.go` — `applySessionParameters`, early
  `_generating` add in `CompileAndGenerateAll`, per-isolate extract
  conversion + `CompileWithInvariants` toggle, and the lineno
  assignment in `addConjsToActions`.
- `goivy/ivy2cpp/generator.go` — `applySessionParameters` wiring in
  `Generate`.

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:4495-4503` (`add_conjs_to_actions`).
- `pyivy/ivy/ivy/ivy_to_cpp.py:4513-4645` (`main_int`).

Tests:

- New (`goivy/ivy2cpp/ivy2cpp_test.go`, appended after the TODO 023
  repl-target Z3 include test):
  - `TestConjectureAppendsAssertToPublicActions`
  - `TestConjectureAddsCheckInvariantsInitializer`
  - `TestPropertyMovedIntoAxioms`
  - `TestConjectureLinenoPropagatesToAssert`

Reminder:

- [x] When this lands, rename to `## DONE 025 - …`, add a `Status:` paragraph
  citing the new Go locations and test names, and update this audit doc in
  the same commit as the implementation.

## TODO 026 - Fill out unsupported action forms

Go locations:

- `goivy/ivy2cpp/action.go:14-73`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:3775-4074`

Gap:

- Go supports a subset of Python action classes: sequence, assert/assume,
  assignment, if, if-some, while, choice, call, local, let, old-bind wrapper
  placeholder, crash, field actions, debug, and native.
- Missing or incomplete forms include logic thunk actions, instantiation
  actions, ranking/progress-related actions, pattern-based updates, full native
  action refs, and any Python action classes not represented in the Go switch.
- Some supported forms are semantic stubs: `BindOlds`, `Crash`, `Havoc`, and
  `Local` do not match Python.

Conformance work:

- Build a Python action-class matrix and add one Go test per class.
- Port every Python `emit` method mechanically, even for classes that emit
  `pass` or assertion failures in Python.
- Replace comment-only unsupported output with an early generation error that
  identifies the missing Python action class.

Tests to add:

- AST/action fixtures that exercise every Python action emission class.

Reminder:

- [ ] When this lands, rename to `## DONE 026 - …`, add a `Status:` paragraph
  citing the new Go locations and test names, and update this audit doc in
  the same commit as the implementation.

## TODO 027 - Add Python-compatible parser/writer and value conversion for every sort

Go locations:

- `goivy/ivy2cpp/repl.go:10-280`
- `goivy/ivy2cpp/types.go:91-168`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:2420-2674`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2676-2852`

Gap:

- Go parsers are simple hand-written conversions for bool, enums, numeric
  ranges, and numeric uninterpreted sorts.
- Python parser/writer support is generated per type and integrates with
  `ivy_value`, `_arg`, `_arg_seq`, serializers, destructors, variants, arrays,
  native values, bitvectors, and strings.

Conformance work:

- Delete the ad hoc parser/writer model once the Python-generated parser helpers
  are ported.
- Use the same parse failure categories and messages as Python where possible:
  syntax errors, out-of-bounds values, enum constant failures, and arity errors.

Tests to add:

- Bad and good REPL values for every sort category.

Reminder:

- [ ] When this lands, rename to `## DONE 027 - …`, add a `Status:` paragraph
  citing the new Go locations and test names, and update this audit doc in
  the same commit as the implementation.

## TODO 028 - Implement Python-compatible randomization for all generated types

Go locations:

- `goivy/ivy2cpp/z3.go:234-282`
- `goivy/ivy2cpp/types.go:91-168`
- `goivy/ivy2cpp/generator.go:272-478`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:979-1006`
- `pyivy/ivy/ivy/ivy_cpp_types.py:355-496`

Gap:

- Go randomization is mostly helper-variable assignment through simple Z3 value
  helpers.
- Python-generated types provide per-type randomization hooks, and the solver
  generator path uses these together with constraints and model evaluation.

Conformance work:

- Generate `__randomize` helpers for all Python-supported sorts.
- Use the same randomization hooks in init/test/gen paths and in complex nested
  values.

Tests to add:

- Randomization for enums, ranges, uninterpreted sorts, destructors, variants,
  bitvectors, strings, arrays/functions, and native types.

Reminder:

- [ ] When this lands, rename to `## DONE 028 - …`, add a `Status:` paragraph
  citing the new Go locations and test names, and update this audit doc in
  the same commit as the implementation.

## TODO 029 - Align emitted C++ expression scoping and temporary management

Go locations:

- `goivy/ivy2cpp/expr.go:18-120`
- `goivy/ivy2cpp/action.go:137-180`
- `goivy/ivy2cpp/action.go:419-429`

Python references:

- `pyivy/ivy/ivy/ivy_cpp.py:72-135`
- `pyivy/ivy/ivy/ivy_cpp.py:185-223`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3243-3256`

Gap:

- Go manually writes strings with indentation. Python has a C++ context model
  with scopes, class members, globals, one-time declarations, impl sections, and
  temporary naming.
- This affects correctness for temps, duplicate helper declarations, native
  inserts, class/impl separation, and target-specific code placement.

Conformance work:

- Either port the Python C++ context model directly or build an equivalent Go
  abstraction with the same observable output sections.
- Ensure temporaries, helper declarations, and native snippets are emitted in
  the same scopes as Python.

Tests to add:

- Golden tests for helpers that must be emitted once, class members, impl-only
  blocks, inline native snippets, and expression temporaries inside nested
  actions.

Reminder:

- [ ] When this lands, rename to `## DONE 029 - …`, add a `Status:` paragraph
  citing the new Go locations and test names, and update this audit doc in
  the same commit as the implementation.

## TODO 030 - Add a Python/Go oracle test suite before filling feature gaps

Go locations:

- No dedicated oracle suite found under `goivy/ivy2cpp`.

Python references:

- Entire `pyivy/ivy/ivy/ivy_to_cpp.py`
- Entire `pyivy/ivy/ivy/ivy_cpp_types.py`
- Entire `pyivy/ivy/ivy/ivy_cpp.py`

Gap:

- The port is large enough that feature-by-feature manual inspection is risky.
  There should be an oracle suite that runs the Python generator and the Go
  generator on the same Ivy fixtures, then compares generated C++ at a chosen
  level of strictness.

Conformance work:

- Build fixtures grouped by feature: basic actions, type lowering, arrays,
  hash-thunks, variants, destructors, natives, bitvectors, REPL, solver init,
  gen/test, progress, and isolates.
- Start with exact text comparison for small helpers and structural comparison
  for full files if exact formatting differs.
- Add compile-and-run tests for representative generated C++.
- Mark unsupported Go cases with expected failures until each TODO is ported,
  then flip them to passing.

Tests to add:

- This TODO is the test harness itself. It should become the gate for claiming
  the Go generator is a faithful mechanical port.

Reminder:

- [ ] When this lands, rename to `## DONE 030 - …`, add a `Status:` paragraph
  citing the new Go locations and test names, and update this audit doc in
  the same commit as the implementation.

## Suggested porting order

1. Port Python's C++ context/type model (`ivy_cpp.py` and type lowering) so later
   features have the right foundation.
2. Port runtime class skeleton, includes, method declarations, and CLI targets.
3. Port expression and action emission class-by-class with oracle fixtures.
4. Port serializers/parsers, destructor/variant helpers, and native support.
5. Port solver initial-state generation, Z3 conversions, and gen/test action
   generation.
6. Port REPL/server/test runtime.
7. Add broad compile/run fixtures and close remaining feature-specific gaps.
