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

## TODO 005 - Port `ivy_cpp_types.py` bitvector and string C++ types

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

## TODO 006 - Replace the Go variant implementation with Python's ref-counted wrapper

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

## TODO 007 - Implement full destructor/struct support

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

## TODO 008 - Match Python's state-symbol and extensional-relation analysis

Go locations:

- `goivy/ivy2cpp/generator.go:527-561`
- `goivy/ivy2cpp/expr.go:403-477`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:33-35`
- `pyivy/ivy/ivy/ivy_to_cpp.py:44-76`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3351-3377`

Gap:

- Go collects state symbols from relations/functions with ad hoc exclusions for
  definitions, destructors, constructors, and mixin initializers.
- Python starts from all logic symbols, excludes constructors and solver
  interpreted symbols, and computes extensional relations by scanning
  initializers and exported action bodies.
- Go's quantifier/extensional handling only recognizes one simple mutable
  relation bound and does not use Python's global extensional-relation analysis.

Conformance work:

- Port `all_state_symbols` and `extensional_relations` exactly.
- Exclude solver/interpreted symbols via the Python equivalent of
  `slv.solver_name(None)`.
- Use the extensional relation set consistently in expression emission,
  quantifier emission, randomization, and action generation.

Tests to add:

- Models with extensional relations updated in actions and initializers.
- Quantifiers bounded by extensional relations where the relation is not a simple
  direct mutable state symbol.

## TODO 009 - Port Python method signatures, parameter passing, and return handling

Go locations:

- `goivy/ivy2cpp/generator.go:576-603`
- `goivy/ivy2cpp/generator.go:641-665`
- `goivy/ivy2cpp/action.go:379-417`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:396-412`
- `pyivy/ivy/ivy/ivy_to_cpp.py:1479-1517`
- `pyivy/ivy/ivy/ivy_to_cpp.py:1551-1571`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3811-3886`

Gap:

- Go emits action parameters mostly by value and multiple returns by non-const
  reference.
- Python annotates action parameters, distinguishes inputs/outputs, return-ref
  cases, const-ref cases, public actions, and generated temporaries for aliased
  returns.
- Go initializes single returns to zero and returns the local; Python has
  generated output-reference handling and special cases for action calls that
  return into variables, fields, or expressions.
- Go call emission lacks Python's `___ivy_stack` push/pop logic for gen/test
  traces.

Conformance work:

- Port `annotate_action` and `emit_method_decl`.
- Port Python's call-action lowering, including variable returns, temporary
  returns, return refs, alias-safety, stack push/pop, and variant upcasts.
- Use the full C++ type-passing policy from the type port.

Tests to add:

- Actions with multiple outputs, output aliases, large value parameters, variant
  returns, destructor returns, and nested action calls under gen/test target.

## TODO 010 - Port derived definitions, constructors, and skolemized helpers

Go locations:

- `goivy/ivy2cpp/definitions.go:17-56`
- `goivy/ivy2cpp/definitions.go:98-140`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:1351-1375`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2295-2347`

Gap:

- Go emits definitions as direct C++ methods when the body can be emitted.
- Python emits derived functions plus constructor methods for sort constructors.
  It also skolemizes unsupported derived forms with `fml:` symbols, creates
  nondeterministic returns, and adds constraints via assumptions/assertions where
  appropriate.
- Go does not emit constructor functions for declared sort constructors.

Conformance work:

- Port `emit_derived` and `emit_constructor`.
- Implement skolemized/nondeterministic derived definitions exactly as Python,
  including naming and constraints.
- Ensure constructors are excluded from state symbols but available as emitted
  C++ functions where Python emits them.

Tests to add:

- Derived definitions that are direct expressions, quantified formulas, native
  definitions, and constructor-backed sorts.

## TODO 011 - Complete expression emission

Go locations:

- `goivy/ivy2cpp/expr.go:18-120`
- `goivy/ivy2cpp/expr.go:155-200`
- `goivy/ivy2cpp/expr.go:202-223`
- `goivy/ivy2cpp/names.go:15-64`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:3004-3043`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3060-3078`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3086-3103`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3123-3223`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3243-3256`

Gap:

- Go supports a narrow set of constants, variables, equality/boolean connectives,
  basic arithmetic/comparison operators, applications, simple variants, simple
  quantifiers, and plain `some`.
- Missing or incomplete Python expression features include macros, `LogicLet`,
  lambdas, interpreted operators, native literals, string literals through
  `__strlit`, bitvector operators, casts, natural-number saturation, finite
  `SomeMin`/`SomeMax`, action terms, destructor field access where storage is
  non-scalar, variant upcast/downcast, and large-function/hash-thunk operations.
- Python frequently emits temporaries for complex expressions; Go usually emits
  direct inline expressions.

Conformance work:

- Port `emit_constant`, `emit_special_op`, `emit_bv_op`, `emit_app`, and
  `temp` behavior.
- Add macro expansion and let/lambda handling in the same places Python does.
- Respect Python's expression-level type conversions and storage-class-specific
  read/write behavior.
- Audit every default `unsupported expression` path and either implement the
  Python case or fail before writing a partial C++ file.

Tests to add:

- Golden expression tests for each AST expression class used by Python's
  `emit_expr`.
- Compile/runtime tests for nested lets, native constants, string literals,
  casts, bitvector arithmetic, and variant/destructor field reads.

## TODO 012 - Port quantifier bounds, iterable sorts, and `some`/min/max

Go locations:

- `goivy/ivy2cpp/expr.go:283-328`
- `goivy/ivy2cpp/expr.go:403-477`
- `goivy/ivy2cpp/expr.go:479-628`
- `goivy/ivy2cpp/expr.go:630-664`
- `goivy/ivy2cpp/action.go:217-220`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:1655-1711`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3264-3290`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3301-3336`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3351-3377`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3380-3465`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3486-3555`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3919-3946`

Gap:

- Go loops only over bool/enumerated/range sorts and a very limited
  one-variable extensional relation case.
- Python extracts bounds from formulas, inequalities, derived constraints,
  extensional relations, iterable attributes, and sort cardinality information.
- Go rejects `some_min` and `some_max` in `emitIfSome`; Python implements
  `Some`, `SomeMin`, and `SomeMax` with bound-aware loops and return values.
- Go's plain `some` often returns the first parameter or zero when no witness is
  found, while Python's generated code tracks witness existence and default
  behavior more carefully for the context.

Conformance work:

- Port `is_iterable_sort`, `is_any_integer_type`, `sort_bounds`, `open_loop`,
  `get_bound_exprs`, `get_bounds`, `get_extensional_bound_exprs`, `emit_quant`,
  and `emit_some`.
- Add support for inequality-derived bounds on integer/nat/range variables.
- Add `SomeMin` and `SomeMax` for both expression and action contexts.
- Ensure witness variables, found flags, breaks, and defaults match Python.

Tests to add:

- Quantifiers over finite sorts, ranged integers, cardinality-bounded sorts,
  extensional relations, and derived-bound formulas.
- `some`, `some_min`, and `some_max` in expressions and actions.

## TODO 013 - Fix assignment/update semantics, especially quantified assignments

Go locations:

- `goivy/ivy2cpp/action.go:137-180`
- `goivy/ivy2cpp/action.go:469-509`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:3624-3652`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3654-3663`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3665-3701`
- `pyivy/ivy/ivy/ivy_to_cpp.py:3703-3764`

Gap:

- Go writes assignment targets in-place while looping over free variables.
- Python uses a two-phase temporary for quantified assignments so RHS reads see
  the old value, then copies the temporary back. For large functions it can emit
  thunk-based fallback updates.
- Go does not implement Python's guarded/bounded update loops or storage-class
  specific assignment rules.
- Go's field-action updates assume direct mutable struct fields and do not cover
  Python's complex types, array fields, variants, or hash-thunk storage.

Conformance work:

- Port Python's `emit_assign_simple`, `emit_assign_large`,
  `emit_assign_bounded`, and `emit_assign`.
- Use temporaries whenever Python would use them, especially for assignments with
  free variables or self-referential RHS expressions.
- Add storage-specific write logic for arrays, hash-thunks, destructors,
  variants, and native types.

Tests to add:

- Self-referential map updates such as `f(X) := f(X) + 1`.
- Quantified assignments with multiple variables, guarded updates, large-domain
  updates, and field updates inside destructor/variant values.

## TODO 014 - Correct havoc, local variables, choices, and old-value binding

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

## TODO 015 - Port native declarations, native code blocks, and callback thunks

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

## TODO 016 - Replace the simple REPL with Python's REPL/server/test runtime

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

## TODO 017 - Port action generation and solver-backed random testing

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

## TODO 018 - Complete Z3 sort/declaration/eval/set conversion

Go locations:

- `goivy/ivy2cpp/z3.go:30-56`
- `goivy/ivy2cpp/z3.go:181-221`
- `goivy/ivy2cpp/z3.go:234-282`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:687-729`
- `pyivy/ivy/ivy/ivy_to_cpp.py:730-756`
- `pyivy/ivy/ivy/ivy_to_cpp.py:772-799`
- `pyivy/ivy/ivy/ivy_to_cpp.py:806-824`
- `pyivy/ivy/ivy/ivy_to_cpp.py:826-870`
- `pyivy/ivy/ivy/ivy_to_cpp.py:4469-4480`

Gap:

- Go emits generic solver helper templates and some sort declarations for enum,
  range, and uninterpreted sorts.
- Python emits full sort declarations, symbol declarations, eval helpers, field
  setters, and state setters across the complete type system.
- Go lacks solver support for arrays/hash-thunks, destructors, variants, native
  types, bitvectors, strings, and CPPTYPE custom conversions.

Conformance work:

- Port `emit_sorts`, `emit_decl`, `emit_eval`, `emit_set_field`, and `emit_set`.
- Add generated `__to_solver` and `__from_solver` helpers for every Python type
  class.
- Ensure solver names, enum constants, uninterpreted values, and range values
  match Python's generated C++ exactly.

Tests to add:

- Solver round-trip tests for every type class and nested type combination.
- Initial-state and action-generation fixtures that require model evaluation
  into complex state values.

## TODO 019 - Match Python progress/rely logic

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

## TODO 020 - Handle requires/ensures/subgoals and exported action semantics exactly

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

## TODO 021 - Fix debug/trace output semantics

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

## TODO 022 - Add full serialization/deserialization support

Go locations:

- `goivy/ivy2cpp/repl.go:10-280`
- `goivy/ivy2cpp/generator.go:391-436`
- `goivy/ivy2cpp/generator.go:438-478`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:2213-2254`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2420-2674`
- `pyivy/ivy/ivy/ivy_cpp_types.py:355-496`

Gap:

- Go emits stream operators for enums/destructors/variants but does not emit the
  Python serializer/deserializer API.
- Python emits `__ser`, `__deser`, `_arg`, stream output, and solver conversion
  helpers for generated and custom types.
- The REPL/test/server runtime depends on these helpers for values crossing text,
  network, and model boundaries.

Conformance work:

- Port Python's serialization and argument parsing method generation for every
  sort kind.
- Ensure nested functions, arrays, hash-thunks, destructors, variants, native
  types, bitvectors, and strings serialize exactly like Python.

Tests to add:

- Round-trip serialization for all supported type classes, including nested
  values.

## TODO 023 - Restore Python include ordering and support headers

Go locations:

- `goivy/ivy2cpp/generator.go:130-155`
- `goivy/ivy2cpp/z3.go:11-23`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:1947-1973`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2030-2069`
- `pyivy/ivy/ivy/ivy_to_cpp.py:2206-2211`

Gap:

- Go emits a minimal set of standard includes and only adds Z3 includes for
  `test`/`gen`.
- Python includes target-specific runtime headers such as `ivy_value.hpp`,
  `ivy_repl.hpp`, `ivy_z3_helpers.hpp`, platform/network/thread headers, and
  other support headers based on generated features.
- Missing includes are a compile break once Python-compatible runtime features
  are ported.

Conformance work:

- Port include selection and ordering from Python.
- Add feature gates for threads, sockets, timers, REPL, Z3, serialization, and
  native snippets.
- Preserve Python's `stdafx` behavior.

Tests to add:

- Compile generated C++ for each target and platform mode after every runtime
  feature port.

## TODO 024 - Match Python's member-name collision checks

Go locations:

- No equivalent check found in `goivy/ivy2cpp`.

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:1830-1834`

Gap:

- Python checks whether module symbols collide with generated class member
  names. Go does not perform this validation.

Conformance work:

- Port `check_member_names` before generation.
- Ensure reserved generated names such as runtime helpers, locks, streams, and
  solver helpers cannot be shadowed by Ivy declarations.

Tests to add:

- Ivy declarations that collide with generated C++ members should fail with a
  Python-compatible error.

## TODO 025 - Port conjecture/property/isolate integration

Go locations:

- `goivy/ivy2cpp/compile.go:29-86`
- `goivy/ivy2cpp/generator.go:38-115`

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:4495-4503`
- `pyivy/ivy/ivy/ivy_to_cpp.py:4513-4645`

Gap:

- Python adds conjectures to action behavior and has isolate/v2-specific
  generation control flow.
- Go exposes an `Isolate` field but does not port the full Python behavior
  around isolates, conjectures, and module preparation.

Conformance work:

- Port conjecture insertion and isolate handling from Python's main path.
- Verify that generated actions and tests include the same property checks as
  Python.

Tests to add:

- Isolated modules with conjectures/properties, including generated tests that
  fail when a conjecture is violated.

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
