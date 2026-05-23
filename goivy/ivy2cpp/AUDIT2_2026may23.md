# ivy2cpp Go port AUDIT2, 2026-05-23 04:28:44 UTC

This is a second-pass, independent audit of the Go port of the Python
Ivy-to-C++ translator. It was produced by:

1. Cataloguing the public surface of the Python source of truth in
   `~/ivy/pyivy/ivy/ivy/` (primarily `ivy_to_cpp.py`, `ivy_cpp.py`,
   `ivy_cpp_types.py`, with helpers from `ivy_solver.py`).
2. Cataloguing the Go port surface in `~/ivy/goivy/ivy2cpp/` (34 Go
   source files, 3 test files, 24,083 LOC; 270+ test functions).
3. Comparing the two and recording every Python-vs-Go divergence,
   without yet consulting the prior audit `TODO_AUDIT2026may21.md`.
4. Only after step 3, reading the May 21 audit and cross-referencing
   each finding so net-new items are clearly distinguished from
   refinements of items already known.

The May 21 audit declared 001-028 DONE and left two open:
- TODO 029 — C++ context/scope/temporary model parity.
- TODO 030 — Python/Go oracle test suite.

This audit appends items **031-046** below. Each item is
self-contained — gap, Python references, Go file:line locations,
conformance work, a comprehensive unit testing plan, and a reminder
checklist. The numbering continues the May 21 series so the two
audits can be read side by side.

Numbering scheme:
- 031-036, 045: net-new findings.
- 037, 040, 042, 043: refinements that sharpen TODO 029 or 030 with
  concrete code/test prescriptions.
- 038, 039, 041: net-new findings adjacent to DONE items where the
  May 21 audit declared completion but specific Python features are
  still missing.
- 033: directly contradicts the DONE 014/026 status — surfaced by
  the still-present `recover()` scaffold in `action_gen.go`.
- 044, 046: informational hygiene items.

The cross-reference table at the bottom maps every AUDIT2 item to its
relationship with the May 21 audit.

A standing acceptance gate: when items 031-046 all land and Item 043
(the oracle harness) reports 100% pass, the mechanical port is
finished.

---

## DONE 031 - Z3-aware thunk generation (gen/test target)

Created: 2026-05-23 04:28:44 UTC

Gap:

- `thunk.go:28-31` explicitly defers the Z3/gen-mode path of
  `make_thunk`. The Python implementation emits a thunk struct
  whose `to_z3` method adds a forall-quantified constraint binding
  the thunk's evaluation to the solver's interpretation of the
  function symbol. The Go port emits only the runtime `operator()`
  body, omitting `to_z3` entirely.
- Observable consequence: under `target=gen` and `target=test`, any
  state symbol whose assignment is lowered to a thunk (i.e., any
  large-domain function assignment) is invisible to Z3. The solver
  produces models that ignore the thunk's value, the generator
  prints counter-examples that disagree with the runtime, and
  `init_gen` cannot constrain initial state through such symbols.
- Comment in `thunk.go:30-31` says "this is sufficient for
  `impl`/`repl`/`class` targets". That is true for runtime, but it
  is not "sufficient" for the gen/test targets where the May 21
  audit DONE 017 claims solver-backed random testing works.

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:499-604` — `make_thunk` (entire
  function; the Z3 emission is at lines 538-602).
- `pyivy/ivy/ivy/ivy_to_cpp.py:511-514` — environment-symbol
  filtering used by `to_z3`.
- `pyivy/ivy/ivy/ivy_to_cpp.py:3654-3662` — `emit_assign_large`,
  the caller.
- `pyivy/ivy/ivy/ivy_solver.py:formula_to_z3` — used to render the
  thunk body into a Z3 expression.

Go locations:

- `goivy/ivy2cpp/thunk.go:28-31` — comment marking the deferral.
- `goivy/ivy2cpp/thunk.go:36-89` — `makeThunk`; struct emission
  body where `to_z3` should be added.
- `goivy/ivy2cpp/thunk.go:91-115` — `emitThunkBody`; substitution
  routine that will need a Z3-encoded variant.
- `goivy/ivy2cpp/solver_emit.go:125-146` — branch (2) of
  `emitSetSolver` consumes the thunk's `to_z3` via the
  `__to_solver` template; depends on this item.

Conformance work:

- Add a `g.usesZ3()` check inside `makeThunk`. When true:
  - After the `operator()` member, emit:
    `void to_z3(gen &__g, const z3::expr &__v) { ... }`
  - The body must call `__g.add(forall(<dom-quants>, <pred>))`
    where `<dom-quants>` are `ctx.constant("arg.argN", sort(D))`
    for each domain slot, and `<pred>` is
    `__to_solver(__g, __v(<args>), <substituted body>)`.
  - The substitution mirrors `emitThunkBody` but renders into a
    Z3 expression string via the existing `formulaToSmtlib`
    helper at `action_gen.go:formulaToSmtlib`.
- Hook the captured environment symbols: each must be re-encoded
  via `__to_solver(__g, ctx.constant("env_name", env_sort), env_name)`
  before the forall, so the solver sees the captured state.
- Update the `__to_solver<thunk<D,R>>` template specialization in
  the runtime header (under `z3.go:emitZ3SolverTemplates`) to call
  `t->to_z3(g, v)` instead of the current best-effort fallback.

Unit testing plan:

Create `thunk_z3_test.go` with these tests:

- `TestMakeThunkEmitsToZ3MethodUnderGen` — fixture: an Ivy model
  with `var f : T -> int` and an init action `f(X) := 0`. Generate
  with `target=test`. Assert the emitted impl contains
  `void to_z3(gen &__g, const z3::expr &__v)` and that the
  enclosing struct is named `__thunk__0`.
- `TestMakeThunkSkipsZ3MethodUnderImpl` — same fixture but
  `target=impl`. Assert the substring `to_z3` does NOT appear.
- `TestMakeThunkSkipsZ3MethodUnderClass` — same fixture but
  `target=class`. Assert `to_z3` absent.
- `TestMakeThunkZ3QuantBindsAllDomainSlots` — fixture: state
  `f : (T1, T2, T3) -> int`. Assert the emitted body contains, in
  order, `ctx.constant(\"arg.arg0\"`, `ctx.constant(\"arg.arg1\"`,
  `ctx.constant(\"arg.arg2\"`, then `forall(__quants, ...)`.
- `TestMakeThunkZ3CapturesEnvSyms` — fixture with state `g : int`
  referenced inside `f(X) := g + 1`. Assert the Z3 body re-encodes
  `g` via `__to_solver(...g...)` before the `forall`.
- `TestMakeThunkSingleVarUsesArgNotArg0` — fixture with one-arg
  domain. Per Python `make_thunk` lines 533-535, the substitution
  uses bare `arg`, not `arg.arg0`. Assert the body says `arg`.
- `TestThunkZ3CompileSmoke` (gated by `SLOW_CPP_TEST=1`) — fixture:
  `mc_test_z3_thunk.ivy` with `target=test`, build, run for one
  iteration, assert the binary exits 0 and the printed model
  satisfies the forall.

Status:

- Implemented the full Python `make_thunk` Z3 path in
  `goivy/ivy2cpp/thunk.go`: gen/test thunks now inherit
  `z3_thunk<D,R>`, carry `__ident`, initialize
  `z3_thunk_counter`, emit the primitive and constant fast paths,
  declare local arg/env/result symbols, encode captured env symbols
  under dynamic `__loc_<ident>__<name>` solver names, parse the
  equality SMT expression, rename env symbols, and substitute domain
  args/result against the caller's Z3 application.
- Generalized solver env emission in `goivy/ivy2cpp/solver_emit.go`
  so `emit_set` can be reused with Python's `prefix='g.'`,
  `gen='g'`, dynamic `csname`, scalar/function/destructor branches,
  and `g.slvr.add(...)`.
- Added runtime support in `include2cpp/ivy_go_z3.hpp`:
  `z3_thunk_counter`, `gen::parse_expr`, string `int_to_z3` /
  `__to_solver<std::string>` support, and `__z3_rename`.
- Tests added in `goivy/ivy2cpp/thunk_z3_test.go`:
  `TestMakeThunkZ3GeneralSingleArgEnvEncoding`,
  `TestMakeThunkZ3GeneralMultiArgSubstitutionAndFunctionEnv`,
  `TestMakeThunkZ3ConstantNumericFastPath`,
  `TestMakeThunkZ3ConstantCPPInterpFastPath`,
  `TestMakeThunkZ3PrimitiveSortReturnsTrue`,
  `TestMakeThunkSkipsZ3MethodOutsideGenAndTest`,
  `TestIvyGoZ3RuntimeThunkHelpers`, and
  `TestMakeThunkZ3GeneralGeneratedCPPCompiles`.
- Verified with `XTRACE_OFF=1 go test ./ivy2cpp -count=1`,
  `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -run
  TestMakeThunkZ3GeneralGeneratedCPPCompiles -count=1`, and
  `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -run
  TestHashThunkToSolverSpecializationEmittedForTestTarget -count=1`.

---

## DONE 032 - Thunk environment-symbol filtering vs `is_derived`

Created: 2026-05-23 04:28:44 UTC

Gap:

- `thunk.go:117-159` (`thunkEnvSymbols`) captures every non-loop
  `*goivy.Const` referenced in the thunk body as a field. Python
  `make_thunk` excludes symbols in the `is_derived` table because
  derived definitions are inlined separately and do not need to be
  captured.
- Observable consequence #1: derived predicates and functions get
  captured as thunk fields, even though they have no storage. The
  resulting C++ either fails to compile (no member of that name)
  or shadows the class member with an uninitialized field.
- Observable consequence #2: under gen/test (after Item 031),
  derived-symbol fields produce redundant `__to_solver` constraints
  duplicating the derived definition, potentially making the
  solver query unsatisfiable.
- The Go file admits the gap: comment `thunk.go:122-127` says
  "derived/function filtering is a follow-up."

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:511-514` — `make_thunk` env
  filtering: `env = [sym for sym in syms if sym not in is_derived]`.
- `pyivy/ivy/ivy/ivy_to_cpp.py` (search for `is_derived = {}` in
  `module_to_cpp_class`) — global initialization of the derived
  table.

Go locations:

- `goivy/ivy2cpp/thunk.go:117-159` — `thunkEnvSymbols`.
- `goivy/ivy2cpp/definitions.go:1-179` — `derivedDefinitions`,
  `definitionNames`, the source of the equivalent set.
- `goivy/ivy2cpp/generator.go` (Generator struct) — needs a
  `derivedSymNames` cached set if not already populated.

Conformance work:

- Extend `thunkEnvSymbols` to consult `g.definitionNames()`
  (already present at `definitions.go`). Skip any `Const` whose
  `Name` is in that set.
- Additionally, skip function-typed `Const` symbols. Python skips
  these because capturing a function value across a thunk struct
  boundary would require recursion into more thunks. The Python
  code does not explicitly check, but the AST shape prevents
  function refs from appearing in expression position outside
  apply contexts; Go should add a defensive type check on
  `c.CSort.(*LogicFunctionSort)` and emit an unsupported diagnostic
  if one slips through.
- Update the `thunk.go:122-127` comment to describe the new
  filter and reference this audit item.

Unit testing plan:

Tests, added to `thunk_z3_test.go`:

- `TestThunkEnvSymbolsExcludesDerivedDefinition` — fixture: a
  module with `definition p(X) = q(X)`. Inside the thunk body
  reference `p(arg)`. Assert the emitted struct has NO field named
  `p`, and that the body inlines the definition's RHS or calls
  the class method `p(...)` directly.
- `TestThunkEnvSymbolsIncludesPlainState` (control) — fixture
  with `var s : int`. Assert `s` IS a field of the thunk struct.
- `TestThunkEnvSymbolsSkipsNumeralsAndBooleans` — regression
  pinning the current `thunk.go:148-153` behavior.
- `TestThunkEnvSymbolsRejectsFunctionSymbol` — synthesized
  expression referencing a free function symbol; assert the
  generator records an unsupported diagnostic and does NOT add
  the function as a field.
- `TestThunkEnvSymbolsStableOrder` — capture order must be
  deterministic across runs; assert two consecutive generations
  produce identical field lists.

Status:

- Implemented derived-aware thunk preparation in
  `goivy/ivy2cpp/thunk.go`: thunk bodies now expand derived
  applications before env capture, derived heads are excluded from
  captured fields, numerals/booleans remain skipped, and bare
  higher-order function values emit an unsupported diagnostic instead
  of becoming bogus fields. Applied state functions remain capturable,
  preserving self-referential thunk assignments.
- Added `Head` to `derivedDefinition` in
  `goivy/ivy2cpp/definitions.go` so thunk expansion can key
  parametric definitions by the exact defining symbol.
- Tests added in `goivy/ivy2cpp/thunk_z3_test.go`:
  `TestThunkEnvSymbolsExcludesDerivedDefinition`,
  `TestThunkEnvSymbolsIncludesPlainState`,
  `TestThunkEnvSymbolsSkipsNumeralsAndBooleans`,
  `TestThunkEnvSymbolsRejectsBareFunctionSymbol`, and
  `TestThunkEnvSymbolsStableOrder`.
- Verified with `XTRACE_OFF=1 go test ./ivy2cpp -run
  'TestThunkEnvSymbols|TestEmitAssignLargeThunkFallback|TestMakeThunkZ3'
  -count=1`, `XTRACE_OFF=1 go test ./ivy2cpp -count=1`, and
  `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -run
  TestThunkEnvSymbolsExcludesDerivedDefinition -count=1`.

---

## DONE 033 - Remove `Action.Update` panic-recover scaffold

Created: 2026-05-23 04:28:44 UTC

Status 2026-05-23: Removed the `defer/recover` wrapper around
`goivy.GetUpdateForArt` in `action_gen.go`; illegal action-update
panics now propagate instead of silently demoting a solver-backed
action generator to weak random-input mode. Added
`action_gen_update_test.go`, whose matrix covers the legal update
dispatch surface (assume/assert/subgoal/requires/ensures, assignment,
havoc, set, native/debug/return, field updates, sequence, choice/env
including determinized branches, boolean and `some` ifs, while, local,
let, calls with input and return formals, bind_olds, crash, fail,
schema instantiate, and default no-op action rows). The same matrix
also exercises `buildActionGenPlan` with `plan.fallback == false`,
and a negative test proves illegal action trees are no longer recovered
inside action-gen analysis. No incomplete legal `Update` implementation
was surfaced by this pass.

Gap:

- `action_gen.go:83-94` wraps `goivy.GetUpdateForArt` in a
  `defer/recover()`. When a panic fires, `plan.fallback` flips to
  true and the generator silently demotes to weak (random-input)
  mode. The comment at `action_gen.go:85` cites "TODO 014/026
  territory" — but both items are marked DONE in the May 21 audit.
- Real observable consequence: any action subtype whose `Update`
  method panics will cause the strong solver-backed test harness
  to be replaced by random-input fuzzing for that specific action,
  with no visible warning. This contradicts the DONE 017 claim
  that solver-backed random testing works.
- Python's equivalent code path has no recover. A panic in the
  Python Update chain would surface immediately, forcing a fix.

Python references:

- `pyivy/ivy/ivy/ivy_actions.py:Action.update` and the family of
  `update_of_X` methods. There is no equivalent of `recover`.
- `pyivy/ivy/ivy/ivy_to_cpp.py:emit_action_gen` — caller; assumes
  `update_of_action` always returns a valid `Update`.

Go locations:

- `goivy/ivy2cpp/action_gen.go:83-94` — the recover block.
- `goivy/ivy2cpp/action_gen.go:85` — stale comment citing DONE
  items.
- `~/ivy/goivy/...` — every Action subtype's `Update` method; use
  `grep -rn "func.*Update(" ~/ivy/goivy/ast ~/ivy/goivy/actions`
  to enumerate.

Conformance work:

1. Enumerate every `Action` subtype with `grep -rnE
   'func \([a-z]+ \*?\w+\) Update\('` across the goivy tree.
2. For each, cross-check the Python `update_of_X` (in
   `ivy_actions.py` and `ivy_compiler.py`). Identify panics or
   nil-derefs that fire on legal AST inputs.
3. Fix each Update implementation. Where the Python returns a
   trivial empty `Update`, return one in Go too (do not panic).
4. After all Updates are panic-free, **delete** the `recover()` at
   `action_gen.go:83-94`. Replace with a direct call:
   `upd := goivy.GetUpdateForArt(plan.act, g.Mod, nil)`.
5. Replace `plan.fallback = true; plan.fallbackReason = ...` paths
   that depended on the recover with a clear unsupported
   diagnostic, so future regressions surface loudly.
6. Update the stale comment at `action_gen.go:85`.

Unit testing plan:

Tests, in new `action_gen_panic_test.go`:

- `TestGetUpdateForArtNeverPanicsOnSupportedActions` — table-
  driven; one row per Action subtype. Build a minimal synthetic
  instance, call `GetUpdateForArt`, assert no panic and result is
  either non-nil or an explicit empty update.
- `TestActionGenStrongPathForEveryActionShape` — for each
  supported action shape, run `buildActionGenPlan` and assert
  `plan.fallback == false`. This is the regression net: if a new
  Action subtype is added and we forget its Update, this test
  must fail.
- `TestActionGenFallbackReasonIsHumanReadable` — for any shape
  that legitimately falls back (e.g., truly unsupported), assert
  the reason does not start with `"GetUpdate panicked: "`.
- `TestNoRecoverInActionGen` — static check; grep the file for
  `recover()` and assert it's absent.

Reminder:

- [x] When this lands, rename to `## DONE 033 - …`, mention the
  removal of the recover, and update.
- [x] If this work surfaces any incomplete Update in goivy core,
  open a follow-up audit item in this file before closing.

---

## DONE 034 - Large-function `__to_solver` end-to-end coverage

Created: 2026-05-23 04:28:44 UTC

Status 2026-05-23: Updated `emitSetSolver` so destructor-record ranges
only take the recursive field path when the whole function sort is not
large, matching Python's `if sort.rng.name in sort_destructors and not
is_large_type(sort)` branch. Large destructor-range functions now use the
forall-quantified `__to_solver` path instead of falling into an
unsupported unenumerable-domain diagnostic. `emitHashThunkToSolver` now
emits concrete `__to_solver` overloads for both mutable and const
`hash_thunk<D,R>` references, delegating to the Python-shaped
`to_solver_class<hash_thunk<D,R>>` specialization so the Go runtime's
primitive primary `__to_solver` no longer steals hash-thunk calls.
Added `solver_emit_test.go` coverage for large cardinality functions,
large destructor-range functions, single-key and ctuple hash-thunk
`to_z3` dispatch, per-domain deduplication, and the stale-comment sweep.
Verification: focused TODO 034 tests, slow C++ compile smoke for
single-key/ctuple hash-thunk solver generation, and full
`XTRACE_OFF=1 go test ./ivy2cpp -count=1` all pass.

Gap:

- `solver_emit.go:65-67` comment claims the large-function branch
  is "deferred to milestone 5", but `solver_emit.go:129-146` does
  emit a forall-quantified `add(forall(__quants, __to_solver(...)))`
  call. The comment is stale.
- However, the forall body invokes `__to_solver(*this, apply(...),
  obj.name[X0,...])`, and the third argument resolves to a
  `hash_thunk<D,R>` value when the symbol is stored as a thunk.
  Without Item 031, `__to_solver<hash_thunk<...>>` has no body
  that meaningfully constrains the solver — it falls through to
  the runtime hash table, which Z3 cannot reason about.
- Observable consequence: `init_gen` produces constraints that
  Z3 satisfies trivially because the thunk side of the equation
  is unconstrained. The runtime then disagrees with the model.

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:844-853` — large-function branch.
- `pyivy/ivy/ivy/ivy_to_cpp.py:emit_hash_thunk_to_solver` — the
  `__to_solver<hash_thunk<...>>` specialization that Python emits
  globally.

Go locations:

- `goivy/ivy2cpp/solver_emit.go:65-67` — stale comment.
- `goivy/ivy2cpp/solver_emit.go:125-146` — branch (2)
  implementation.
- `goivy/ivy2cpp/z3.go:emitZ3SolverTemplates` — where the
  per-hash_thunk specialization should be emitted.
- `goivy/ivy2cpp/thunk.go` — see Item 031 (this item depends on
  it).

Conformance work:

1. Land Item 031 first. The `thunk->to_z3(...)` call site
   in the new specialization is the load-bearing piece.
2. Rewrite `solver_emit.go:65-67` to describe the *current* state
   (implemented, depends on Item 031 for soundness).
3. Add an emission for the global
   `__to_solver<hash_thunk<D,R>>` template that calls
   `t.fun->to_z3(g, v)`, gated on `g.usesZ3()`.
4. Audit Python's `emit_hash_thunk_to_solver` line-by-line and
   confirm every helper call (in particular the de-duplication
   per (D, R) tuple) is mirrored on the Go side using the
   `nativeOnceMemo`-style mechanism.

Unit testing plan:

Tests, in new `solver_emit_test.go`:

- `TestEmitSetSolverLargeFunctionEmitsForall` — fixture with state
  `f : int -> int` where the domain exceeds `largeThresh`. Assert
  the emitted impl contains `add(forall(__quants,
  __to_solver(*this, apply(...), <obj-access>)))`.
- `TestEmitSetSolverLargeFunctionUsesThunkToZ3` — same fixture
  after Item 031 lands. Assert the runtime header emits a
  `__to_solver<hash_thunk<D,R>>` specialization that calls
  `v.fun->to_z3(g, ...)`.
- `TestEmitSetSolverFallbackOnUnEnumerableDomain` — fixture with
  an uninterpreted domain. Assert the unsupported diagnostic is
  byte-identical to Python's wording (currently Go says "domain
  not enumerable"; verify Python uses the same).
- `TestEmitSetSolverHashThunkSpecializationEmittedOnce` — fixture
  with two state symbols sharing the same domain/range. Assert
  exactly one `__to_solver<hash_thunk<D,R>>` template specialization.
- `TestEmitSetSolverLargeFunctionCompileSmoke`
  (`SLOW_CPP_TEST=1`) — compile and run a tiny gen-mode model,
  assert it terminates.
- Documentation fix verified by `TestSolverEmitCommentNotStale`
  (a string-match assertion).

Reminder:

- [x] When this lands, rename to `## DONE 034 - …` and update.

---

## DONE 035 - BV widths > 64 bits

Created: 2026-05-23 04:28:44 UTC
Completed: 2026-05-23

Gap:

- `cpp_types.go:151` rejects `bv[N]` whenever `N > 64`. Python
  `ivy_cpp_types.py:XBV` supports arbitrary widths (Python int is
  unbounded) and emits C++ that uses `__int128` for 65-128, with
  a helper class for wider widths.
- Observable consequence: any Ivy model declaring a cryptographic
  hash, a 128-bit counter, or a 256-bit identifier is rejected at
  generation time. The Go port currently emits the diagnostic
  "bitvector width for X: bv[N] exceeds 64-bit C++ lowering",
  which surfaces to the user without further guidance.

Python references:

- `pyivy/ivy/ivy/ivy_cpp_types.py:XBV.__init__` — initialization
  of bit-vector classes, no width upper bound.
- `pyivy/ivy/ivy/ivy_cpp_types.py:XBV.emit_templates` — `__from_solver`,
  `__to_solver`, `__randomize` specializations. These use Python
  unbounded integers; the C++ side needs an explicit wider type.
- `pyivy/ivy/ivy/ivy_to_cpp.py:ctype` — type dispatch for large bvs.

Go locations:

- `goivy/ivy2cpp/cpp_types.go:130-160` — `cppInterpType.primitiveType`
  and the reject branch.
- `goivy/ivy2cpp/bv_expr.go:emitBVNumeral` — width-dependent
  numeral masking; will need 128-bit literal support.
- `goivy/ivy2cpp/runtime.go` — for the helper class declaration.

Conformance work:

1. Extend `primitiveType()`:
   - 1-32 bits  → `unsigned`
   - 33-64 bits → `unsigned long long`
   - 65-128 bits → `unsigned __int128`
   - >128 bits  → a helper class `ivy_uint<N>` declared at file
     scope in the runtime header.
2. Update `bvMask(int)` to produce widening-aware mask literals:
   for 65-128, mask is `(((__uint128_t)1 << N) - 1)`.
3. For widths >128, the helper class should expose `+`, `-`, `*`,
   `&`, `|`, `^`, `<<`, `>>`, `==`, `<`, `<=`, `>=`, `>`, `<<` on
   ostream, `>>` on istream. Mirror Python's `XBV.emit_templates`
   exactly.
4. Update `build.go` to ensure GCC/Clang flags are present for
   `__int128` (no extra flag needed on GCC ≥ 4.6 and Clang ≥ 3.1)
   and refuse to use MSVC for widths > 64 (MSVC has no `__int128`).
5. Audit Z3 integration: the bv sort name remains `bv[N]`; the
   C++/Z3 conversion already widens through string round-trip in
   the Python version. Mirror that.

Unit testing plan:

Tests, in new `cpp_types_test.go`:

- `TestBVWidth128GeneratesInt128Typedef` — fixture: `var x : bv[128]`.
  Assert the typedef line says `typedef unsigned __int128`.
- `TestBVWidth128MaskExpression` — assert the mask literal used
  for the type matches `(((__uint128_t)1<<128)-1)` or an
  equivalent canonical form.
- `TestBVWidth256GeneratesHelperClass` — fixture: `var x : bv[256]`.
  Assert a `class ivy_uint<256>` declaration appears at file scope.
- `TestBVWidth128OnMSVCRejected` — when `Config.Compiler` resolves
  to MSVC and a `bv[128]` is present, assert the diagnostic explains
  the limitation.
- `TestBVWidthUnsupportedHasUsefulMessage` — fixture: `var x : bv[513]`
  on GCC. Assert the diagnostic names the sort and the declaration
  site, not just "unsupported".
- `TestBVCompileBVOps` (`SLOW_CPP_TEST=1`) — fixture: a model that
  XORs two `bv[128]` values; compile and run; assert the printed
  result equals the expected XOR.

Status:

- Implemented the general wide-BV path rather than a narrow slice:
  `bv[1..32]` lowers to `unsigned`, `bv[33..64]` to `unsigned long long`,
  `bv[65..128]` to `unsigned __int128`, and `bv[N]` for `N > 128`
  to `ivy_uint<N>`.
- Added `include2cpp/ivy_wide_uint.hpp` with arbitrary-width unsigned
  arithmetic, bitwise, shift, comparison, stream, parse, decimal
  conversion, random, and hash support. Generated headers include it
  when a model uses a wide bit-vector.
- Updated numeral masking, Z3 string round-trips, `_arg`, `__ser`,
  `__deser`, `__from_solver`, `__to_solver`, and `__randomize` paths
  for both `unsigned __int128` and `ivy_uint<N>`.
- `BuildPlanFor` now rejects `compiler=cl` for generated outputs that
  use widths greater than 64, because MSVC lacks `unsigned __int128`.
- Added coverage in `cpp_types_wide_bv_test.go` for 128-bit, 256-bit,
  513-bit, MSVC rejection, and the helper operator surface.

Verification:

- `XTRACE_OFF=1 go test ./ivy2cpp -run 'TestBVWidth(128|256|513)|TestBVWidth256SupportHeaderExposesGeneralOperators' -count=1`
- `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -run 'TestBVWidth(128|256|513)' -count=1`
- `XTRACE_OFF=1 go test ./ivy2cpp -count=1`

---

## DONE 036 - `nondet.go` domain-synthesis fallbacks

Created: 2026-05-23 04:28:44 UTC
Completed: 2026-05-23

Gap:

- `nondet.go:99-110` records `unsupported` on two synthesis paths:
  (a) `goivy.NewVariable` failure when synthesizing a loop
  variable, (b) the domain-sort iterator returning an error.
- Python `mk_nondet_sym` (`ivy_to_cpp.py:3470-3525`) has no such
  escape paths. The AST helpers in Python always succeed for legal
  sort inputs; if they failed, generation would crash, not silently
  skip.
- Observable consequence: nondeterministic initialization for
  structured sorts (variants, destructor records, bitvector
  domains) can silently lose state. The generated C++ leaves the
  symbol at default-constructed values rather than `___ivy_choose`.

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:mk_nondet` (around line 3470).
- `pyivy/ivy/ivy/ivy_to_cpp.py:mk_nondet_sym`.
- `pyivy/ivy/ivy/ivy_to_cpp.py:emit_randomize` (sibling routine for
  randomization, mirrors the same expectations on the iterator).

Go locations:

- `goivy/ivy2cpp/nondet.go:1-216` — entire file.
- `goivy/ivy2cpp/nondet.go:99-110` — the two unsupported sites.
- `~/ivy/goivy/...` — `goivy.NewVariable` and the iterator helpers
  used by `mkNondetSym`. Find with `grep -rn "func NewVariable" ~/ivy/goivy`.

Conformance work:

1. For each failure cause, trace into the underlying helper:
   - `NewVariable` failure: identify which sort kind triggers it.
     If `NewVariable` legitimately requires a valid sort and is
     being called with nil/zero sort, the bug is in
     `mkNondetSym`'s caller. Fix at the caller.
   - Domain-iterator failure: the iterator should cover bitvectors,
     enums, ranges, native sorts, and uninterpreted (via the
     randomizer). Extend missing cases.
2. After fixes, remove the two `g.unsupported` calls at
   `nondet.go:99-110`. The generator should never fall through to
   them for legal inputs.
3. Add a guard at the entry to `mkNondetSym` that documents the
   contract: "Domain sorts must be iterable; non-iterable domains
   should be routed to the randomizer upstream."

Unit testing plan:

Tests, in new `nondet_test.go`:

- `TestMkNondetSymStructWithNestedDestructors` — fixture: a sort
  with two destructor fields (one bv, one enum). Assert no
  unsupported diagnostic and the emitted code initializes both
  fields with `___ivy_choose(...)`.
- `TestMkNondetSymBitvectorDomain` — fixture: `var f : bv[8] -> int`.
  Nondet init. Assert the emitted loop iterates 0..256.
- `TestMkNondetSymVariantDomain` — fixture: a variant sort with
  three branches. Assert the tag is chosen first via
  `___ivy_choose(3, ...)`, then each branch's body is initialized
  via the per-branch routine.
- `TestMkNondetSymUninterpretedSort` — assert the call routes to
  the randomizer rather than the iterator.
- `TestMkNondetSymNoUnsupportedOnLegalInputs` — table-driven; one
  row per sort kind. Assert every row produces zero unsupported
  diagnostics.
- `TestMkNondetSymBoundaryWidthsBitVector` — bitvector widths of 1,
  2, 8, 16, 32, 64. Assert the loop bound is the correct
  `1<<width`.

Status:

- Removed the soft `g.unsupported` fallback paths from `mkNondetSym`.
  If the generator reaches a non-iterable bounded-array domain now,
  it panics with a contract violation instead of emitting partial C++.
- Extended `loopHeaderForSort` to cover positive-cardinality
  integer-like domains, including pure `bv[N]` sorts and card-bounded
  uninterpreted sorts that lower to integer storage.
- Fixed `nondetSkipSort` so pure `bv[N]` sorts are nondet-initialized
  like Python's plain `bv[...]` path; only real cpptype helper sorts
  (`strbv`, `intbv`) are skipped.
- Centralized nondet initialization through `mkNondetValue`, so scalar
  locals, bounded-array cells, struct fields, thunk-produced values,
  destructor records, and variant supertypes all recurse through the
  same logic.
- Added variant nondet construction: choose the variant tag, nondet
  initialize the selected subtype payload, then upcast into the
  supertype wrapper.
- Added focused coverage in `nondet_test.go` for destructor fields,
  bv domains, variant locals, uninterpreted-domain thunks, boundary bv
  loop headers, and no-unsupported legal inputs.

Verification:

- `XTRACE_OFF=1 go test ./ivy2cpp -run 'TestMkNondetSym' -count=1`
- `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -run 'TestMkNondetSym' -count=1`
- `XTRACE_OFF=1 go test ./ivy2cpp -count=1`

---

## DONE 037 - Thunk struct emission scope (file vs local)

Created: 2026-05-23 04:28:44 UTC

(Sharpens TODO 029 with a concrete thunk-specific scope question.)

Gap:

- `thunk.go:14-23` (block comment) admits that Python emits thunk
  structs at impl/file scope via `thunks = impl`, whereas Go emits
  them inline at the assignment site. The Go choice is legal C++11
  (local classes are template-arg-friendly) but diverges from the
  mechanical-port mandate.
- Three concrete divergences:
  (a) Repeated thunks (same body, different call sites) get
      emitted multiple times in Go, once in Python.
  (b) Local-class names are scoped to the function; Python uses
      file-scope names that can be referenced by other emitted
      helpers (notably the Z3 `__to_solver` specialization in
      Item 034).
  (c) Some compilers warn under `-Wshadow` on repeated local
      classes named `__thunk__N` if N counters are not unique.

Python references:

- `pyivy/ivy/ivy/ivy_cpp.py:152-161` — `add_once_global`,
  `add_global`. Memoization by content.
- `pyivy/ivy/ivy/ivy_to_cpp.py:make_thunk` (around line 499) —
  uses the impl buffer (`thunks = impl`).

Go locations:

- `goivy/ivy2cpp/thunk.go:14-23` — comment.
- `goivy/ivy2cpp/thunk.go:36-89` — `makeThunk` writes to the
  caller's `*cppWriter`, which is the action body buffer.
- `goivy/ivy2cpp/writer.go` — `cppWriter`; needs a file/impl
  buffer companion (Item 042).
- `goivy/ivy2cpp/generator.go` — needs a `nativeOnceMemo`-style
  thunk dedupe table.

Conformance work:

1. Depends on Item 042 (context model) landing first.
2. Move `makeThunk`'s struct emission target from the per-action
   writer to the impl/file buffer obtained via the new C++ context
   model.
3. Memoize by a content hash of (domain sort tuple, range sort,
   captured env tuple, body s-expression). Mirror Python's
   `add_once_global`.
4. Keep `__thunk__N` numbering stable across the run; the counter
   advances only when a new (non-deduped) struct is emitted.
5. Update the block comment at `thunk.go:14-23` to reflect the
   new behavior.

Unit testing plan:

Tests, in new `thunk_emission_test.go`:

- `TestThunkStructEmittedAtFileScope` — fixture: a forall-X
  assignment. Assert the `struct __thunk__0` line appears in the
  impl section but NOT inside any function body.
- `TestIdenticalThunksEmittedOnce` — fixture with two identical
  forall assignments in two different actions. Assert exactly one
  `struct __thunk__` line appears in the output.
- `TestDifferentBodiesGetDifferentStructs` — fixture: two
  syntactically-different forall assignments. Assert two distinct
  structs.
- `TestThunkNumberingStableAcrossRuns` — run `Generate` twice on
  the same module; assert byte-identical impl output.
- `TestThunkScopeNoShadowWarning` (`SLOW_CPP_TEST=1`) — build with
  `-Wshadow -Werror`; assert success.

Status:

- Implemented before Item 042 by adding a dedicated impl/file-scope
  thunk buffer to `Generator` and flushing it before constructor and
  method bodies. Full `Generate` now mirrors Python's `thunks = impl`
  behavior; direct white-box calls still preserve their old inline
  writer behavior unless file-scope thunk emission is explicitly
  enabled.
- Added content memoization by thunk class, qualified domain/range
  types, loop variables, captured environment tuple, and expanded body
  expression. The `__thunk__N` counter advances only for new emitted
  structs, giving stable numbering across runs.
- Moved nondet-generated thunk structs into the same file-scope path and
  added scoped nondet construction so uninterpreted, struct, and variant
  domains/ranges can be emitted outside class method bodies.
- Added `thunk_emission_test.go` coverage for file-scope placement,
  identical-thunk one-time emission, distinct-body emission, byte-stable
  generation, and `-Wshadow -Werror` generated-C++ compilation.

Verification:

- `XTRACE_OFF=1 go test ./ivy2cpp -run 'TestThunk|TestEmitAssignLargeThunkFallback|TestMkNondetSymUninterpretedDomainUsesThunk' -count=1`
- `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -run 'TestThunk(StructEmittedAtFileScope|IdenticalThunksEmittedOnce|DifferentBodiesGetDifferentThunkStructs|ScopeNoShadowWarning)' -count=1`
- `XTRACE_OFF=1 go test ./ivy2cpp -count=1`

---

## DONE 038 - `emit_special_op` parity audit (bitvector + string ops)

Created: 2026-05-23 04:28:44 UTC

Gap:

- Python `emit_special_op` (in `ivy_to_cpp.py` around lines
  2660-2750+) dispatches on operator name for at least the
  following: `bvand`, `bvor`, `bvxor`, `bvnot`, `bvneg`, `bvshl`,
  `bvlshr`, `bvashr`, `<<`, `>>`, `cast`, `bfe[lo:hi]`, plus
  string `+`, `<`, `<=`.
- Go `bv_expr.go:emitBVApply` (lines 34+) explicitly covers only:
  `bvand`, `bvor`, `bvnot`, `cast`, concatenation, `bfe`.
- Missing from Go: `bvxor`, `bvneg`, `bvshl`, `bvlshr`, `bvashr`,
  and `<<`/`>>` lowering. Plus any string operators emitted by
  Python outside the strbv class. Observable consequence: any Ivy
  model using bitwise XOR, shifts, or arithmetic shift right will
  hit the default branch and likely emit incorrect C++.

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:emit_special_op` (search for
  `'bvand'` in the file to land at the dispatch table).
- `pyivy/ivy/ivy/ivy_to_cpp.py:emit_bv_op` — neighbor function.

Go locations:

- `goivy/ivy2cpp/bv_expr.go:34-160` — `emitBVApply` and helpers.
- `goivy/ivy2cpp/expr.go:emitApply` (search "bvand") — dispatch
  entry point.

Conformance work:

1. Read Python's `emit_special_op` table top to bottom, recording
   every operator name. Cross-check this audit's list.
2. For each missing operator, add a case in `emitBVApply`:
   - `bvxor x y` → `((x) ^ (y)) & mask`
   - `bvneg x` → `(-(x)) & mask`
   - `bvshl x y` → `((x) << (y)) & mask`
   - `bvlshr x y` → `((unsigned-cast x) >> (y)) & mask`
   - `bvashr x y` → `((signed-cast x) >> (y))` (no mask: sign
     bits propagate; cast back to unsigned only if needed by
     consumer)
   - `<<`, `>>` (if used standalone outside bv operators):
     mirror Python.
3. For string operators (if Python emits them at this level),
   add the corresponding string-typed branch. Cross-reference
   with `strbv` template specializations in `cpp_types.go`.
4. Move the unknown-operator fallback to call
   `g.unsupported(w, "unknown BV operator %s", op)` with the
   operator name and arity reported.

Unit testing plan:

Tests, in new `bv_expr_test.go`:

- `TestBVXorEmitsXor` — fixture: `f := g ^ h` over `bv[8]`. Assert
  emitted body contains `((g) ^ (h)) & 0xff` (or canonical form).
- `TestBVShiftLeftMasksWidth` — `x << 3` over `bv[8]`. Assert
  mask present.
- `TestBVLogicalShiftRightUsesUnsignedCast` — `bvlshr` over a
  signed-looking expression. Assert cast to unsigned before shift.
- `TestBVArithmeticShiftRightSignExtends` — `bvashr` over `bv[8]`.
  Assert cast to signed int8 before shift.
- `TestBVNegEmitsTwoComp` — `bvneg(x)` over `bv[8]`. Assert
  `(-(x)) & 0xff`.
- `TestUnknownBVOperatorReportsUnsupported` — synthesize a fake
  operator `bvfoo`; assert diagnostic names it.
- `TestStringOpAddEmitsPlus` (if applicable to Python's surface)
  — fixture: string concatenation; assert `+` is emitted.
- `TestBVOperatorTableCompleteVsPython` — meta-test: parse
  `bv_expr.go` source and assert every Python operator name
  appears as a case label.

Status:

- Audited the live Python `emit_special_op`/`emit_bv_op` path. The
  checked-out Python source has a smaller explicit table than this audit
  item described (`concat`, `bfe[...]`, and `bvand`/`bvor`/`bvnot`), but
  the Go tree already exposes the broader Z3 BV operator family. The Go
  implementation now covers the whole family used by the port rather
  than only the narrow Python table.
- Added lowering for `bvxor`, `bvneg`, `bvshl`, `bvlshr`, `bvashr`,
  symbolic `<<`/`>>`, and named arithmetic synonyms `bvadd`, `bvsub`,
  `bvmul`, `bvudiv`, `bvurem`, while preserving existing `+`, `-`, `*`,
  `/`, `%`, `bvand`, `bvor`, `bvnot`, `concat`, `bfe[...]`, and `cast`.
- Made BV casts/concat/extract explicit so the generated C++ remains
  type-correct for primitive BV widths, `unsigned __int128`, and
  arbitrary-width `ivy_uint<N>`.
- Added guarded shift lowering so large shift amounts do not rely on C++
  undefined behavior. Arithmetic right shift now sign-extends for all
  supported BV widths, including arbitrary-width `ivy_uint<N>`.
- Extended `ivy_wide_uint.hpp` with cross-width constructors,
  primitive conversion operators, unary negation, and saturating shift
  amount helpers needed by the general lowering.
- Unknown reserved-looking BV operators now report a diagnostic naming
  the operator and arity instead of falling through to an incorrect
  storage access expression.
- Added `bv_expr_test.go` coverage for XOR, negation, guarded shifts,
  arithmetic sign extension, symbolic shift aliases, wide BV shift
  helpers, unknown-operator diagnostics, string `+`, and operator-table
  coverage.

Verification:

- `XTRACE_OFF=1 go test ./ivy2cpp -run 'TestBV|TestBitvectorTypesAndExpressionsShape|TestStringOpAddEmitsPlus' -count=1`
- `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -run 'TestBV(Xor|Shift|Logical|Arithmetic|Neg|Wide|Operator|Unknown)|TestSymbolicShiftAliases|TestBitvectorTypesAndExpressionsShape|TestStringOpAddEmitsPlus' -count=1`
- `XTRACE_OFF=1 go test ./ivy2cpp -count=1`

---

## DONE 039 - `find_vs` / MSVC toolchain detection

Created: 2026-05-23 04:28:44 UTC

Gap:

- Python `ivy_to_cpp.py:find_vs()` locates the Visual Studio
  installation on Windows by reading the registry and/or running
  `vswhere.exe`. The resolved bin/include/lib directories are
  injected into the build environment so `ivyc` works out of the
  box on a fresh Windows install.
- Go `build.go:msvcBuildPlan` builds the command line correctly
  for MSVC but assumes `cl.exe` and its include/lib paths are on
  PATH. There is no `find_vs` equivalent.
- Observable consequence: a fresh Windows Claude/CI run with no
  prior `vcvars64.bat` invocation will fail at the link step with
  errors about missing standard library paths.

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:find_vs` — full function.
- `pyivy/ivy/ivy/ivy_to_cpp.py` — calls to `find_vs` from
  `main_int` / build planning.

Go locations:

- `goivy/ivy2cpp/build.go:msvcBuildPlan` (around line 200+).
- `goivy/ivy2cpp/build.go:cxxCompiler` and
  `goivy/ivy2cpp/build.go:cxxCompilerFor` — compiler detection.
- `goivy/ivy2cpp/build.go:1-50` — preamble (for adding the
  Windows-only build tag if extracted to a separate file).

Conformance work:

1. Create a new file `build_findvs.go` with build tag
   `//go:build windows`. Implement `findVS() (vsInfo, error)`:
   - Try `vswhere.exe -latest -format json -property installationPath`.
   - Parse the JSON.
   - Construct paths for `<install>/VC/Tools/MSVC/<ver>/bin/Hostx64/x64`
     and the corresponding include / lib paths.
2. On non-Windows: stub `findVS()` returns
   `vsInfo{}, errors.New("vswhere unavailable")`.
3. `msvcBuildPlan` calls `findVS()`; if successful, prepends the
   bin dir to PATH in the command's `Env`, and appends include /
   lib paths to the compile/link args.
4. Failure: build plan still returns a usable plan (assuming
   `vcvars` was sourced); the user gets a clear error from cl.exe
   if not.

Unit testing plan:

Tests, in new `build_findvs_test.go` (build tag `windows` for
some, generic for others):

- `TestFindVsParsesVswhereOutput` (generic) — feed a captured
  `vswhere -format json` payload via a test helper that
  bypasses the actual exec. Assert installation root extracted.
- `TestMsvcBuildPlanIncludesToolchainPaths` (generic) — with a
  synthetic `vsInfo` injected, assert the build plan's compile
  args include `/I<install>/VC/.../include`.
- `TestFindVsAbsentReturnsClearError` (generic, mocked) — assert
  the error mentions `vswhere`.
- `TestFindVsLinuxStubIsNoOp` (`//go:build !windows`) — assert
  the function returns the documented error string.
- `TestMsvcBuildPlanWithoutFindVSStillBuilds` — assert that when
  `findVS()` fails, the command line is still valid (the user is
  expected to have sourced `vcvars`).
- Integration smoke (manual, not CI): on a Windows VM with a
  fresh VS install, run `make test`; assert pass.

Status:

- Added `findVS()` plumbing with a non-Windows stub and a Windows
  implementation that prefers `vswhere.exe`, falls back to the old
  Python-style Visual Studio directory scan, selects the latest
  `VC/Tools/MSVC/<version>` toolset, and constructs bin/include/lib
  paths.
- On Windows, discovery also attempts to run `vcvarsall.bat amd64 &&
  set` and captures the resulting environment. This preserves the
  Python behavior of building under the Visual Studio environment, while
  also giving `BuildPlan` explicit PATH/INCLUDE/LIB updates.
- `BuildPlan` now carries an optional `Env`; `BuildOutput` applies it
  when invoking the compiler.
- `msvcBuildPlan` uses discovered include/lib directories and prepends
  the toolchain bin directory to PATH. If discovery fails, it still
  returns the existing usable `cl` command line so callers with an
  already-sourced `vcvars` shell keep working.
- Added mocked cross-platform tests for vswhere JSON parsing, latest
  MSVC toolset selection, MSVC include/lib/PATH injection, absent
  vswhere diagnostics, non-Windows stub behavior, and fallback build
  plans.
- No Windows CI configuration was changed in this step; the Windows VM
  smoke remains manual.

Verification:

- `XTRACE_OFF=1 go test ./ivy2cpp -run 'TestFindVs|TestVSInfo|TestMsvcBuildPlan|TestBuildPlanRejectsCLOnNonWindows|TestBVWidth128OnMSVCRejected' -count=1`
- `XTRACE_OFF=1 go test ./ivy2cpp -count=1`

---

## DONE 040 - `lhsTraceString` full port (number format, captures)

Created: 2026-05-23 04:28:44 UTC

(Sharpens TODO 029 and DONE 021.)

Gap:

- `assign.go:80` comment defers the full port of `emit_traced_lhs`
  to TODO 029. The current Go implementation handles
  `name=value` only.
- Python's `emit_traced_lhs` (`ivy_to_cpp.py:3142-3220`,
  approximate) walks the LHS expression tree, emitting a chain of
  `<<` operators that respects the current `number_format` global
  (hex vs decimal), recurses into destructor records (emitting
  `obj.field=...`), and emits array subscripts
  (`a[3]=...`).
- Observable consequence: trace output under
  `target=test/repl` with structured state shows the parent
  symbol's address-of representation rather than the per-field
  state, making debugging much harder.

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:3142-3220` — `emit_traced_lhs`.
- `pyivy/ivy/ivy/ivy_to_cpp.py:number_format` global definition.
- `pyivy/ivy/ivy/ivy_to_cpp.py` — uses of `number_format` for
  ostream chaining.

Go locations:

- `goivy/ivy2cpp/assign.go:80-95` — `lhsTraceString` and the
  current approximation.
- `goivy/ivy2cpp/generator.go` — needs a `NumberFormat`
  field (already has `numberFormatCache`; possibly just expose).

Conformance work:

1. Land Item 042 (context model) so the impl/expr buffer split is
   available.
2. Implement `emitTracedLhs(g *Generator, w *cppWriter, lhs Expr,
   capturedArgs map[string]string)` that mirrors the Python
   walker:
   - For `Const`: emit `"<<name<<"`.
   - For `Apply` to a destructor: recurse with `<<".field"<<...`.
   - For `Apply` to an array: emit `<<"["` then evaluate index
     expression `<<index<<"]"`.
   - Wrap numeric subexpressions with the current number format
     (hex/decimal).
3. Compute number format from the Ivy module's attribute table
   (look for `radix=16`); cache on Generator.
4. Update callers in `assign.go` and `action.go` to invoke the
   new function.

Unit testing plan:

Tests, in new `trace_test.go`:

- `TestTraceLhsRespectsHexNumberFormat` — fixture: module with
  `attribute method = radix=16`. Assert the emitted trace line
  contains `<< std::hex <<` before the value.
- `TestTraceLhsDecomposesDestructorRecord` — fixture: assignment
  to `node.field`. Assert trace contains `<< \"node.field=\"`.
- `TestTraceLhsArraySubscript` — fixture: assignment to `a[3]`.
  Assert trace contains `<< \"a[\" << 3 << \"]=\"`.
- `TestTraceLhsNestedDestructorChain` — fixture: `outer.inner.x`.
  Assert trace shows the full path.
- `TestTraceLhsNamespacedNameSuppressed` — regression for the
  existing `:` prefix suppression at `assign.go:lhsHasNamespacedName`.
- `TestTraceLhsDoesNotEmitWhenTraceOff` — assert trace lines absent
  when `Config.Trace == false`.

Status:

- Replaced the quoted-C++-lvalue approximation with an AST walker that
  emits a Python-shaped stream chain for traced assignment LHS values.
- Constants and variables now trace as source names, function
  applications trace as `name(arg,...)` with evaluated argument
  expressions, and destructor-backed field chains trace as source field
  paths such as `root.child.shade`.
- The existing namespaced-symbol suppression is preserved, so local
  helper assignments such as `loc:tmp` do not produce write traces.
- Trace lines continue to use `Generator.numberFormat()`, so
  `attribute radix = "16"` applies the same hex/showbase stream prefix
  to LHS arguments and RHS values.
- Added `trace_lhs_test.go` coverage for hex number format, function
  application arguments, destructor fields, nested destructor chains,
  namespaced local suppression, and `Trace=false`.

Verification:

- `XTRACE_OFF=1 go test ./ivy2cpp -run 'TestTraceLhs|TestAssignSimpleEmitsWriteTraceUnderTrace|TestNumberFormatHexFromRadixAttribute' -count=1`
- `XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -run 'TestTraceLhs|TestAssignSimpleEmitsWriteTraceUnderTrace|TestNumberFormatHexFromRadixAttribute' -count=1`
- `XTRACE_OFF=1 go test ./ivy2cpp -count=1`

---

## DONE 041 - `emit_value_parser` per-sort coverage

Created: 2026-05-23 04:28:44 UTC

Gap:

- Python `emit_value_parser` (referenced from the REPL boilerplate
  around `ivy_to_cpp.py:1818-1920`) emits per-sort parsers for
  every sort kind: bitvectors of any width, enums (numeric and
  named), ranges with bounds-checking, strings with escape
  handling, native sorts via user-provided parsers, and structured
  records via recursive destructor parsers.
- Go has a parser layer (`repl.go:emitEnumSortArgSpecImpls` and
  friends), but a head-to-head check against Python's table has
  not been done. There is no per-sort coverage matrix and no
  semantic-equivalence test.

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py:emit_value_parser` (around line
  1818).
- `pyivy/ivy/ivy/ivy_to_cpp.py:1830-1920` — argument-parsing
  helpers per sort.
- `pyivy/ivy/ivy/ivy_cpp_types.py:StrBV` and `IntBV` — parser
  emission for those types.

Go locations:

- `goivy/ivy2cpp/repl.go:1-651` — entire REPL emission.
- `goivy/ivy2cpp/repl.go:emitEnumSortArgSpecDecls` and
  `emitEnumSortArgSpecImpls` — enum parsers.
- `goivy/ivy2cpp/cpp_types.go:emitXBVClassDecl` — bitvector
  parser pieces.
- `goivy/ivy2cpp/destructor.go` — record parser.
- `goivy/ivy2cpp/variant.go` — variant parser.

Conformance work:

1. Walk Python's `emit_value_parser` and produce a coverage
   matrix: sort kind × Go file responsible × current status.
2. For each gap (sort kind without a Go parser), add the
   corresponding emitter. Mirror Python's exact wording for error
   messages so the REPL UX matches.
3. Verify with the canonical s-expression diff harness (CLAUDE.md
   section F): parse a test input on both Python-generated and
   Go-generated binaries, dump the resulting state via the
   `_arg<T>` template, and diff.

Unit testing plan:

Tests, in new `repl_parser_test.go`:

- `TestParserBitvectorWidthN` — table-driven for
  N ∈ {1, 2, 8, 16, 32, 64} (extended to 128, 256 after
  Item 035). Assert the parsed value round-trips.
- `TestParserEnumByName` — for each enum member name and its
  numeric position, assert both forms parse to the same value.
- `TestParserRangeBoundsRespected` — out-of-range input rejected
  with Python-identical message (compare to a captured Python
  error string in `testdata/python_errors/`).
- `TestParserStringWithEscapes` — inputs `\\n`, `\\t`, `\\\"`,
  `\\\\`; assert each unescapes correctly.
- `TestParserDestructorRecord` — nested input
  `{ field = value; field2 = value2 }`. Assert the resulting
  record has both fields populated.
- `TestParserVariantDiscriminator` — input prefixed with the
  variant tag name; assert the correct branch is constructed.
- `TestParserNativeSortRoutesToUserProvided` — fixture: a sort
  with a user-defined native parser; assert the generated code
  calls the user function.
- `TestParserCoverageMatrixMatchesPython` — meta-test: a YAML
  fixture in `testdata/parser_matrix.yaml` lists every sort kind
  the Python parser supports; assert the Go code has emission for
  each.

Reminder:

- Added a parser coverage matrix at
  `goivy/ivy2cpp/testdata/parser_matrix.yaml`, covering Python's
  primitive runtime parsers, pure bitvectors through wide `ivy_uint<N>`,
  `strbv`, `intbv`, named enums, numeric enums, ranges,
  destructor-backed records, variants, native sorts, and positional
  function parameters.
- Centralized generated parser expressions through
  `argExprForSort`/`argExprForSortBound` so action dispatch,
  `emit_value_parser`, positional parameter parsing, destructor fields,
  and variant payloads all route through the same C++ `_arg<T>` emission
  policy.
- Preserved Python's exact range/cardinality behavior: range parser
  bounds still use `csortcard`/`sort_card` semantics (`ub + 1` for
  ranges), matching the source-of-truth rather than adding a separate
  lower-bound clamp.
- Added `repl_parser_test.go` coverage for BV widths
  `{1,2,8,16,32,64,128,256}`, named and numeric enum modes, range
  rejection text, string escape parsing, nested destructor records,
  variant discriminators, and native user-provided parser routing.
- Added a captured Python-compatible range error fixture at
  `goivy/ivy2cpp/testdata/python_errors/range_out_of_bounds.txt`. The
  quoted `"argument 1"` is intentional: Python's generated C++ typedefs
  `__strlit` to `std::string`, so its `operator<<` quotes strings,
  including `out_of_bounds.txt`.

Verification:

- `env XTRACE_OFF=1 go test ./ivy2cpp -run 'TestParser|TestReplDispatchUsesPrimitiveArgForBoolRangeNatStrlit|TestArgSpecVariantBadAndGoodValuePaths|TestValueParserPrefixesLinenoOnError|TestReplMainFunctionSortedParam' -count=1`
- `env XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -run 'TestParser' -count=1`
- `env XTRACE_OFF=1 go test ./ivy2cpp -count=1`

---

## DONE 042 - Port `ivy_cpp.py` context model

Created: 2026-05-23 04:28:44 UTC

(Sharpens TODO 029 by specifying exactly which Python classes to
port and the test surface for each.)

Gap:

- Go's `writer.go` is a thin string builder. Python's `ivy_cpp.py`
  defines an entire C++ context model with class scopes,
  member-vs-local distinction, file-scope vs class-scope
  distinction, once-only globals, and lexical context managers.
- The mechanical-port mandate (CLAUDE.md section B) requires us
  to mirror the Python class structure. Per rule B.7 we may not
  add abstractions, but we also may not omit them. The Python
  context model is non-trivial and must be ported faithfully.
- Observable consequences of NOT porting:
  - Items 037 (thunk file-scope emission) and 040 (trace LHS
    decomposition) cannot be implemented cleanly.
  - Helpers that Python emits "once" (per `add_once_global`) get
    duplicated in Go, inflating the impl file.
  - Lexical class context switching (used inside variant emission
    and destructor emission) has no Go counterpart, leading to
    name-collision edge cases.

Python references:

- `pyivy/ivy/ivy/ivy_cpp.py:1-250+` — entire file.
- Specifically the classes: `CppFile`, `CppText`, `DeadCode`,
  `Context`, `CppContext`, `CppClass`, `CppClassName`, `CppArray`,
  `CppFunction`, `CppReference`, `CppVector`, `TypeDef`,
  `CppMember`, `CppLocal`, `CppScope`.
- And the module-level functions: `get_temp`, `add_global`,
  `add_once_global`, `add_impl`, `add_member`, `add_local`,
  `add_expr`, `current_classname`, `add_header`, `relname`,
  `fullname`.

Go locations:

- `goivy/ivy2cpp/writer.go:1-47` — current `cppWriter` (very
  small).
- `goivy/ivy2cpp/generator.go` — currently holds caches that
  Python uses module-level functions to manage; needs a
  `cppContext` field.

Conformance work:

1. Create `cpp_context.go` with a struct table matching the
   Python classes. Field-by-field:

   | Python class | Python attrs | Go struct | Go fields |
   | CppFile      | filename, mode, indent | CppFile | Filename, Mode string; Indent int |
   | CppText      | code list | CppText | Code []string |
   | DeadCode     | (none) | DeadCode | (empty struct) |
   | Context      | (base) | Context | (interface) |
   | CppContext   | globals, impls, members, locals, expr, classname, global_includes, impl_includes, once_globals | CppContext | Globals, Impls, Members, Locals, Expr *CppText; Classname string; GlobalIncludes, ImplIncludes []string; OnceGlobals map[string]bool |
   | CppClass     | classname, baseclass | CppClass | Classname, Baseclass string |
   | CppClassName | classname | CppClassName | Classname string |
   | CppArray     | cpptype, dims, name | CppArray | CppType CppType; Dims []int; Name string |
   | CppFunction  | cpptype, argtypes, name | CppFunction | CppType CppType; ArgTypes []CppType; Name string |
   | CppReference | cpptype, name, const_ | CppReference | CppType CppType; Name string; Const bool |
   | CppVector    | cpptype, name | CppVector | CppType CppType; Name string |
   | TypeDef      | oldtype, newname | TypeDef | OldType CppType; NewName string |
   | CppMember    | cpptype, name, static, inline | CppMember | CppType CppType; Name string; Static, Inline bool |
   | CppLocal     | (inherits CppMember) | CppLocal | (embeds CppMember) |
   | CppScope     | (none) | CppScope | (empty marker) |

2. Implement the module-level functions as methods on `*CppContext`
   (per CLAUDE.md C: no Go package-level mutable globals). Wire
   from `Generator.Ctx *CppContext`.

3. Port the `with X:` context-manager semantics. In Go, expose
   `Enter() func()` returning a cleanup closure to be `defer`'d.

4. Migrate `cppWriter` to be an adapter over `CppContext` so the
   existing call sites keep working during the transition.

Unit testing plan:

Tests, in new `cpp_context_test.go` — one suite per ported class:

- `TestCppContextHasAllPythonAttrs` — meta-test that scans
  `cpp_context.go` and asserts each Python class field maps to a
  Go field. Driven by a YAML fixture in `testdata/cpp_context_attrs.yaml`.
- `TestCppContextAddOnceGlobalDeduplicatesByContent` — call twice
  with the same string; assert only one append.
- `TestCppContextAddGlobalAlwaysAppends` — same call twice without
  dedupe; assert two appends.
- `TestCppContextAddLocalRoutedToActiveScope` — enter a class
  context, enter a nested scope, `add_local`; assert the addition
  lands in the active scope.
- `TestCppContextAddImplBypassesScope` — same setup; `add_impl`;
  assert it lands in the impl buffer regardless of scope depth.
- `TestCppContextClassNameStack` — nest two class contexts; on
  exit, assert classname pops back to the outer value.
- `TestCppContextTempCounterMonotonic` — call `get_temp` N times;
  assert names are `__temp__0`...`__temp__(N-1)`.
- `TestCppContextTempCounterSurvivesScope` — wrap two `get_temp`
  calls around a nested scope; assert counter does not reset.
- `TestDeadCodeWritePanics` — write to `DeadCode`; assert panic
  (mirrors Python `assert False`).
- `TestCppContextDeterministicOutput` — generate a fixture twice;
  assert byte-identical output (deterministic ordering).

Reminder:

- Added `cpp_context.go`, porting the Python `ivy_cpp.py` context
  model: `CppFile`, `CppText`, `DeadCode`, `Context`,
  `CppContext`, `CppClass`, `CppClassName`, `CppArray`,
  `CppFunction`, `CppReference`, `CppVector`, `TypeDef`,
  `CppMember`, `CppLocal`, and `CppScope`.
- Implemented context-style `Enter(*CppContext) func()` closures for
  class-name stacking, class member routing, member initializers,
  function-local routing, and nested local scopes.
- Ported the module-level Python helpers as `*CppContext` methods:
  add-global, add-once-global, add-impl, add-member, add-local,
  add-expr, current-classname, add-header, relname/fullname, and
  context-owned temp generation.
- Migrated `cppWriter` into an adapter over `CppText`, and wired
  `Generator.Ctx` so the header and implementation buffers are now the
  `CppContext` globals/impls buffers while existing emitter call sites
  keep their `cppWriter` API.
- Routed the native once-only memo through `CppContext.OnceGlobals`,
  matching Python's single `once_globals` set while preserving the old
  generator fallback for white-box tests.
- Added `testdata/cpp_context_attrs.yaml` and `cpp_context_test.go`
  coverage for the class/field mapping, once-only globals, ordinary
  globals, local scope routing, impl bypass, class-name stack popping,
  temp generation, `DeadCode` panic behavior, deterministic generation,
  and Python-style declarations.
- Items 037 and 040 now sit on a faithful context model instead of
  ad-hoc writer state.

Verification:

- `env XTRACE_OFF=1 go test ./ivy2cpp -run 'TestCppContext|TestDeadCode' -count=1`
- `env XTRACE_OFF=1 go test ./ivy2cpp -count=1`
- `env XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -run 'TestParser' -count=1`

---

## DONE 043 - Oracle harness concrete blueprint

Created: 2026-05-23 04:28:44 UTC

(Sharpens TODO 030 with a concrete fixture catalog, comparator,
and acceptance gate.)

Gap:

- TODO 030 calls for an oracle test suite but does not specify the
  fixture set, the comparator strictness, or the acceptance
  criterion. Without those, "is the port complete?" cannot be
  answered objectively.

Python references:

- `pyivy/ivy/ivy/ivy_to_cpp.py` (used as the oracle).
- `pyivy/ivy/ivy/ivy_cpp.py`.
- `pyivy/ivy/ivy/ivy_cpp_types.py`.

Go locations:

- No oracle file currently exists.
- Will live at `goivy/ivy2cpp/oracle_test.go` and
  `goivy/ivy2cpp/testdata/oracle/<fixture>.ivy`.

Conformance work:

1. Create `testdata/oracle/` with the following 14 fixtures
   (one Ivy file per row), each targeting `impl` unless noted:

   | Fixture                | What it exercises                          |
   | empty.ivy              | Smoke; empty module                        |
   | basic_assign.ivy       | Simple `x := 1`                            |
   | forall_assign.ivy      | Two-phase + thunk lowering                 |
   | bv_arithmetic.ivy      | Item 038 coverage (xor/shift/neg)          |
   | enum_dispatch.ivy      | Enum with constructor branches             |
   | range_bounds.ivy       | Numeric range with bounds                  |
   | destructor_record.ivy  | Record sort with two destructor fields     |
   | variant_simple.ivy     | Variant with three branches                |
   | variant_recursive.ivy  | Variant whose body references the sort     |
   | hash_thunk_assign.ivy  | Large-domain assignment forcing thunk      |
   | native_block.ivy       | `<<<` native block with antiquotes         |
   | callback_thunk.ivy     | Callback action referenced from native    |
   | progress_property.ivy  | Progress + rely on a counter               |
   | isolate_two_parts.ivy  | Two isolates exporting interface           |

2. Implement a tokenizing comparator
   (`testdata/oracle/compare_cpp.go`) that:
   - Strips C-style and C++-style comments.
   - Collapses runs of whitespace to a single space, except
     inside string literals.
   - Treats identifiers and numbers as opaque tokens.
   - Reports the first token where the two streams diverge,
     with 80 chars of left+right context.

3. The test harness `oracle_test.go`:
   - For each fixture: spawn Python `ivy_to_cpp <fixture>.ivy`
     into a temp dir; call Go `ivy2cpp.Generate(...)` and write
     the output to a parallel temp dir.
   - Compare with the tokenizing comparator.
   - Gated by `ORACLE_TEST=1` so `make test` stays fast.

4. Acceptance gate: when all 14 fixtures pass token-equality and
   compile-and-run smoke (`SLOW_CPP_TEST=1`), the audit closes
   (the prior TODO 030 also closes).

5. Until then, each fixture has a status entry in
   `testdata/oracle/STATUS.md`: PASS / EXPECTED_FAIL / SKIP. The
   harness emits a per-fixture summary at the end of the run.

Unit testing plan:

Tests, in `oracle_test.go`:

- `TestOracleFixturesExist` — sanity that every fixture in the
  catalog above is on disk.
- `TestOracleSingle/<fixture>` — one subtest per fixture using
  `t.Run`. Compares token streams; reports first divergence.
- `TestOracleCompileGo/<fixture>` (`SLOW_CPP_TEST=1`) — assert
  Go's output compiles under GCC.
- `TestOracleCompilePython/<fixture>` (`SLOW_CPP_TEST=1`) — assert
  Python's output compiles under GCC.
- `TestOracleSemanticEquivalence/<fixture>` (`SLOW_CPP_TEST=1`) —
  drive both compiled binaries with a shared input transcript
  (`testdata/oracle/<fixture>.in`); assert identical stdout.
- `TestOracleStatusFileUpToDate` — meta-test: the catalog above
  matches the rows in `STATUS.md`.

Status:

- Added the full oracle fixture catalog under
  `goivy/ivy2cpp/testdata/oracle/`, including all 14 files named in
  this item and `STATUS.md` rows for PASS / EXPECTED_FAIL / SKIP
  accounting.
- Added `goivy/ivy2cpp/oracle_compare.go` with a tokenizing C++
  comparator that strips C/C++ comments, ignores whitespace outside
  literals, keeps identifiers/numbers as tokens, and reports first
  divergence with 80-character context. Added the requested standalone
  wrapper at `goivy/ivy2cpp/testdata/oracle/compare_cpp.go`.
- Added `goivy/ivy2cpp/oracle_test.go` with:
  `TestCompareCPPTokensStripsCommentsAndWhitespace`,
  `TestCompareCPPTokensPreservesStringLiterals`,
  `TestCompareCPPTokensReportsContext`,
  `TestOracleFixturesExist`, `TestOracleSingle`,
  `TestOracleCompileGo`, `TestOracleCompilePython`,
  `TestOracleSemanticEquivalence`, and
  `TestOracleStatusFileUpToDate`.
- `TestOracleSingle`, Python compile, and semantic checks are gated by
  `ORACLE_TEST=1`; compile/run hooks are gated by `SLOW_CPP_TEST=1`.
  The current catalog starts as EXPECTED_FAIL until fixtures are
  promoted fixture-by-fixture after token parity is proven.
- The harness now uses each fixture basename as the generated class
  name, matching Python's file-basename behavior and exposing real
  token divergences rather than a constant `oracle.cpp`/`oracle.h`
  file-set mismatch. `goivy/ivy2cpp/TODO3.md` records the remaining
  whole-hog parity work after the first visible oracle divergence
  moved into the shared runtime scaffold.
- Verified with
  `env XTRACE_OFF=1 go test ./ivy2cpp -run 'TestCompareCPP|TestOracleFixturesExist|TestOracleStatusFileUpToDate' -count=1`,
  `env XTRACE_OFF=1 ORACLE_TEST=1 go test ./ivy2cpp -run TestOracleSingle -count=1`,
  `env XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -run TestOracleCompileGo -count=1`,
  and `env XTRACE_OFF=1 go test ./ivy2cpp -count=1`.

Reminder:

- [x] When this lands, rename to `## DONE 043 - …` and also
  close the prior TODO 030.

---

## DONE 044. ignored.

## DONE 045 - Sweep stale "deferred" comments

Created: 2026-05-23 04:28:44 UTC

Gap:

- Several comments describe deferred work that is actually
  partially or fully implemented. The mismatch confuses future
  audits.

Specific inventory (verified against current code):

- `solver_emit.go:65-67` — "Deferred to milestone 5 (uses
  make_thunk infrastructure)". The branch *is* implemented at
  lines 129-146; what's actually deferred is the thunk's Z3
  pairing (Item 031). Rewrite to describe accurately.
- `thunk.go:30-31` — "the z3 / gen-mode path (Python lines
  538-602) is currently stubbed". Accurate today; cross-link to
  Item 031.
- `thunk.go:122-127` — "For simplicity, this Go port treats every
  Const ... derived/function filtering is a follow-up." Cross-link
  to Item 032.
- `action_gen.go:85` — "TODO 014/026 territory". But 014 and 026
  are DONE in May 21. Rewrite to reference this audit's Item 033.

Python references:

- N/A (this is a Go-side cleanup).

Go locations:

- `goivy/ivy2cpp/solver_emit.go:65-67`.
- `goivy/ivy2cpp/thunk.go:30-31`.
- `goivy/ivy2cpp/thunk.go:122-127`.
- `goivy/ivy2cpp/action_gen.go:85`.

Conformance work:

- Rewrite each comment to reflect actual state and cross-reference
  the relevant AUDIT2 item. ~5 minutes of editing; do it together
  with whichever item lands first.

Unit testing plan:

Tests, in new `comments_test.go`:

- `TestNoStaleDeferralComments` — read the four files and grep
  for the known stale strings ("deferred to milestone 5",
  "TODO 014/026", "follow-up" without a cross-reference); assert
  absent.
- `TestEveryDeferralCommentReferencesAudit2` — for any comment
  mentioning "deferred" or "TODO", assert it includes a cross-
  reference of the form `AUDIT2 Item NNN` or `TODO_AUDIT2026...`.

Status:

- The four originally inventoried comments had already been made
  accurate by the implementations for AUDIT2 Items 031, 032, 033,
  and 034. Swept the remaining stale init-gen line in
  `goivy/ivy2cpp/solver_emit.go`, replacing "deferred to milestone 5"
  with an accurate reference to `emitHashThunkToSolver` and
  `z3_thunk::to_z3` (AUDIT2 Item 031 / DONE 034).
- Added `goivy/ivy2cpp/comments_test.go` with
  `TestNoStaleDeferralComments` and
  `TestEveryDeferralCommentReferencesAudit2`.
- Verified with
  `env XTRACE_OFF=1 go test ./ivy2cpp -run 'TestNoStaleDeferralComments|TestEveryDeferralCommentReferencesAudit2' -count=1`
  and `env XTRACE_OFF=1 go test ./ivy2cpp -count=1`.

Reminder:

- [x] When this lands, rename to `## DONE 045 - …`.

---

## DONE 046 - ignored.

---

## Suggested execution order

The dependency graph among the 16 items:

1. **Item 042** (C++ context model) — unblocks 037 (file-scope
   thunk emission) and 040 (trace LHS recursion).
2. **Item 033** (remove `recover()`) — unblocks confidence that
   `action_gen` is producing the strong path universally.
3. **Items 031, 032, 034** — the Z3 thunk family. 031 enables 034;
   032 is independent and quick.
4. **Items 035, 038** — bitvector coverage. 038 is small; 035
   touches `__int128`/MSVC and benefits from 038's tests.
5. **Item 036** — nondet completeness; touches goivy core.
6. **Item 041** — REPL parser per-sort coverage.
7. **Item 039** — Windows toolchain detection (separable;
   schedule when a Windows CI runner exists).
8. **Items 037, 040** — sequenced after Item 042 closes.
9. **Item 043** — oracle harness; the final acceptance gate.
   Validates everything else.
10. **Items 045 ** — hygiene; pair with any item above.

## Methodology and how to verify

Per-item workflow:

1. Read the Python reference at the cited line ranges in full.
2. Read the Go code at the cited line ranges in full.
3. Implement the conformance work.
4. Add the prescribed tests in the named file.
5. Run `cd ~/ivy/goivy && make test` (XTRACE_OFF=1 is default
   per Makefile).
6. Update this audit: rename `TODO NNN` → `DONE NNN`, add a
   `Status:` paragraph, link to the test names.

Per-batch (every 3-5 items):

- `ORACLE_TEST=1 make test` (after Item 043 lands).
- `SLOW_CPP_TEST=1 make test` (compile-and-run gate).
- `go test -race ./ivy2cpp -count=1` (after Item 046).
- Spot-check with the canonical s-expression diff harness
  (CLAUDE.md section F) on one fixture from each affected file.

## Cross-reference table

| This audit | Prior audit (May 21)    | Net new?       |
|------------|-------------------------|----------------|
| 031        | partial 017             | yes (Z3 path)  |
| 032        | -                       | yes            |
| 033        | contradicts DONE 014/026| yes            |
| 034        | partial 018             | yes (clarifies)|
| 035        | -                       | yes            |
| 036        | -                       | yes            |
| 037        | sharpens 029            | refinement     |
| 038        | partial 011             | yes            |
| 039        | partial 001             | yes            |
| 040        | sharpens 021 + 029      | refinement     |
| 041        | partial 027             | yes            |
| 042        | sharpens 029            | refinement     |
| 043        | sharpens 030            | refinement     |
| 045        | -                       | yes            |

## Out-of-scope / explicitly NOT in this audit

To prevent double-counting and audit drift, the following are
explicitly NOT new items:

- Anything covered by DONE 001-028 unless this audit explicitly
  contradicts the DONE status (see Item 033).
- Tests already present in `ivy2cpp_test.go` for behaviors that
  are already correct.
- Internal goivy AST/IR work outside `ivy2cpp/` — file bugs in
  the respective package's audit instead.
- Performance work — this audit is about correctness/parity.

End of AUDIT2 2026-05-23.
