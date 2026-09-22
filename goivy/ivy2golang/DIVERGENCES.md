# ivy2golang divergence audit

Audit target: `~/ivy/pyivy/ivy/ivy/ivy_to_cpp.py`.

Audit date: 2026-09-22.

This document records behavior gaps in `ivy2golang` relative to the Python
`ivy_to_cpp.py` generator. The goal is not to criticize local stopgaps; several
were useful while bringing up the port. The goal is to make the remaining
divergences visible and testable.

Workflow for fixing this list:

- Work items in order.
- For each item, first add a fast in-process Go unit test. The test must run
  inside the `go test` process by compiling Ivy source, running generator
  analysis, or inspecting generated source; it must not build generated Go,
  invoke `ivy2cpp`, or launch a separate tester process.
- Fix the implementation until the focused fast test and the normal
  `ivy2golang` / `cmd/ivy2golang` tests pass.
- When an item is fully fixed, put `FIXED` on the item's heading line.
- Run the full slow end-to-end verification only after every item below is
  marked fixed.

## 1. `target=test` does not use Python's solver-backed action generators

Python source behavior:

- `emit_action_gen` wraps `before_export` and `ext_preconds`, computes the
  action update, takes `tr.reverse_image`, trims/expands clauses, extracts local
  solver inputs, extracts field inputs, extracts defined parameters, adds
  relevant definitions and variant axioms, asserts the resulting formula into a
  Z3 generator, and reads back model-selected inputs before execution
  (`ivy_to_cpp.py:1211-1338`).
- The test loop constructs one `*_gen` per exported action and calls
  `generate`; only SAT generators execute (`ivy_to_cpp.py:4358-4443`).

Current Go behavior:

- `emitRandomizedActionCycles` chooses an exported action, generates random
  actual arguments, checks only `ext_preconds` plus leading assumes, and then
  calls the action directly (`ivy2golang/generator.go:2980-3067`).
- Actions containing calls use a clone/trial retry path to suppress rejected
  assumes (`ivy2golang/generator.go:3069-3107`).
- `genActionPreconditionFormulas` only sees `ExtPreconds` and leading assumes,
  not the reverse-image precondition of the whole action
  (`ivy2golang/generator.go:2820-2853`).

Risk:

- Any constraint introduced by a non-leading assume, nested call, local witness,
  assignment preimage, relevant definition, variant axiom, or field extractor can
  be missed or approximated. This can cause spurious `assumption_failed`,
  skipped enabled transitions, trace drift from `ivy2cpp`, or retry-heavy
  generated testers.

How to conform:

- Port the `emit_action_gen` planning flow into `ivy2golang` instead of growing
  more syntactic guard checks. The Go `ivy2cpp` package already has a close
  mechanical port in `ivy2cpp/action_gen.go` (`buildActionGenPlan`); use it as a
  template for shared helper extraction or a Go-output variant.
- Generated Go needs an equivalent of the generator contract: set pre-state,
  randomize solver inputs, solve the full reverse-image formula, evaluate input
  values, apply defined inputs, and execute only on SAT.
- Remove or narrow `emitTestActionAssumeGuards` and `emitTestTrialActionCall`
  once the solver-backed generator is authoritative.

Regression test:

- Add a small `target=test` spec with an exported action whose enabledness
  depends on a non-leading assume after a nested call and on a derived
  definition. As a fast test, inspect the generated Go and generator analysis to
  prove `target=test` uses per-action `generate()` methods and no longer emits
  the direct random-actuals path for exported actions.

Progress:

- 2026-09-22: Added
  `TestTargetTestNonLeadingAssumeGuardSkipsRejectedInputsFast` and fixed the
  narrow transparent-prefix case: in `target=test`, assumes after non-mutating
  assertions are now emitted as pre-action reject guards before tracing or
  executing the action. This prevents a direct runtime `assumption_failed` for
  that subset. The full reverse-image/action-generator port described above is
  still open.
- 2026-09-22: Added `TestTargetTestAssignedStateAssumeUsesPreimageFast` and
  fixed a second narrow reverse-image slice: top-level scalar state assignments
  before an assume are now substituted into the assume guard for `target=test`.
  For example, `saved := c; assume saved = green` emits a pre-action guard on
  `c = green`. This is still limited to simple top-level assignments and does
  not replace the full solver-backed action generator.
- 2026-09-22: Added
  `TestTargetTestConditionalAssumeUsesImplicationGuardFast` and fixed a third
  narrow reverse-image slice: precondition-only branches such as
  `if c = red { assume false }` now emit implication guards like
  `c = red -> false` before the action trace. Branches that can mutate state are
  still intentionally outside this limited fix.
- 2026-09-22: Added `TestTargetTestLocalAssumeUsesTrialGeneratorFast`. Actions
  whose assumptions are not statically expressible as pre-action guards now run
  on an `__ivy_clone` trial object first; rejected candidates suppress tracing
  and leave the original state untouched. This keeps local/nested runtime assume
  failures from surfacing as test failures while the full solver-backed
  generator remains open.
- 2026-09-22: Added `TestTargetTestUsesPerActionGeneratorFast` and refactored
  `target=test` output to emit one `*_generator` type per runnable exported
  action. The randomized formal generation, simple defined-input extraction,
  and syntactic/preimage guard checks now live in each action's `generate()`
  method; the main loop calls `generate()` and only traces/executes on success.
  This restores the Python-style action-generator contract boundary while the
  generator internals are still the limited Go analysis rather than the full Z3
  reverse-image plan.
- 2026-09-22: Added
  `TestTargetTestDerivedAssumeAfterTransparentCallUsesGeneratorGuardFast`.
  `target=test` precondition analysis now treats calls to precondition-only
  private actions as transparent, substitutes callee formals with actual
  arguments, and carries derived-definition assumes such as `is_green(c)` into
  the per-action generator guard. Calls that can mutate state still fall back to
  the trial path; the full solver-backed reverse-image plan remains open.
- 2026-09-22: Added
  `TestTargetTestFormalAssumeInsideIrrelevantLocalUsesGeneratorGuardFast`.
  Local-action wrappers whose collected guards do not reference the local
  symbols are now transparent to `target=test` generator precondition analysis,
  so formal-only assumes inside harmless local declarations become generator
  guards. Guards that actually depend on the local symbol still fall back to the
  runtime trial path until local solver witnesses are modeled.
- 2026-09-22: Added
  `TestTargetTestCallAssignmentPreimageUsesGeneratorGuardFast`. The
  `target=test` prefix/preimage walker now carries one substitution map through
  nested analysis and inlines simple private calls with formal-to-actual
  substitutions. This lets a callee assignment such as `saved := c` feed a
  later caller guard like `assume saved = green`, producing a generator guard on
  `c = green`. Recursive calls and unsupported callee shapes still fall back to
  the trial path.
- 2026-09-22: Added
  `TestTargetTestCallReturnPreimageUsesGeneratorGuardFast`. Simple private
  calls with return values now map callee return formals onto the actual return
  targets during `target=test` preimage analysis, so `call saved := helper(c);
  assume saved = green` produces the same generator guard on `c = green`.
  Recursive calls and complex return targets remain outside this narrow slice.
- 2026-09-22: Added
  `TestTargetTestConditionalAssignmentPreimageUsesIteGuardFast`. Conditional
  branches now collect preimage substitutions in branch-local maps and merge
  state-symbol updates back into the caller with `LogicIte`, so a later guard
  sees shapes such as `if c = red { saved := c }; assume saved = green` as a
  generator guard on `ite(c = red, c, saved) = green`. Unsupported branch
  actions still fall back to the trial path.
- 2026-09-22: Added
  `TestTargetTestRelationPointAssignmentPreimageUsesIteGuardFast`. The
  `target=test` preimage context now carries ordered point updates for
  state-function and relation cells, so `marked(c) := true; assume
  marked(green)` emits a generator guard equivalent to
  `ite(green = c, true, marked(green))`. Bulk quantified relation assignments
  and unsupported point-update shapes still fall back to the trial path.
- 2026-09-22: Added
  `TestTargetTestSetActionPreimageUsesPointUpdateGuardFast`. Relation
  `SetAction` forms such as `marked(c)` and `~marked(c)` now lower through the
  same point-update preimage recorder as explicit assignments, so source-level
  relation set statements contribute generator guards instead of forcing the
  runtime trial path.

## 2. `target=gen` action generators are syntactic guards, not solver generators

Python source behavior:

- `target=gen` uses the same `emit_action_gen` machinery as `target=test`; only
  the boilerplate/logging choices differ (`ivy_to_cpp.py:1211-1338`,
  `ivy_to_cpp.py:4564-4565`).

Current Go behavior:

- `emitGenActionGeneratorGenerate` randomizes formal parameters directly, applies
  a small defined-input extraction over leading guards, checks those guards, and
  returns `true`/`false` (`ivy2golang/generator.go:2613-2636`).
- `emitGenActionInvocations` runs generators once in sorted order
  (`ivy2golang/generator.go:2952-2964`), but the generator does not solve the
  Python reverse-image formula.

Risk:

- `target=gen` can produce different one-shot traces from `ivy2cpp` for any
  action whose enabled input is solver-discoverable but unlikely under direct
  randomization, or whose precondition is not a leading assume.

How to conform:

- Reuse the same action-generator plan as `target=test`. The only intended
  target difference should be the outer runner shape and model logging flag, not
  the definition of an enabled action.

Regression test:

- Add a tiny `target=gen` spec where the only enabled input satisfies a derived
  predicate rather than a direct leading assume. As a fast test, assert the Go
  source contains the solver-backed generator plan and not just `gen.x =
  random`.

Progress:

- 2026-09-22: Added `TestTargetGenAssignedStateAssumeUsesPreimageFast` and
  wired `target=gen` precondition collection through the same limited
  prefix/preimage assume analysis used by `target=test`. For simple top-level
  assignments before an assume, `target=gen` now emits guards on generated
  inputs (for example `gen.c == green`) instead of ignoring the assume or
  checking stale current state. Full solver-backed generator parity remains tied
  to item 1.

## 3. Initial state generation is retry/randomized, not Python's initial model

Python source behavior:

- `emit_init_gen` constructs an SMT generator from `init_cond`, axioms, and
  relevant definitions, solves it, evaluates used state symbols from the model,
  clears progress, sets `___ivy_gen`, and calls `__init`
  (`ivy_to_cpp.py:915-985`).
- `emit_one_initial_state` similarly solves initial constraints and assigns
  model values for used state symbols (`ivy_to_cpp.py:3018-3050`).

Current Go behavior:

- `emitInitialState` randomizes every non-parameter state symbol, emits ad hoc
  initial-condition assignments, then optionally retries up to 1000 times before
  calling `ivyAssume(false, "initial condition")`
  (`ivy2golang/generator.go:1205-1238`).
- The generated `target=gen` init generator just calls `__initState` and
  `__init` (`ivy2golang/generator.go:2557-2577`).

Risk:

- The generated tester may fail initial conditions that Python/C++ satisfies
  immediately with a model, or it may spend time retrying random states. This is
  exactly the class of bug that produces runtime `initial condition: error:
  assumption failed` even when `ivy2cpp` runs.

How to conform:

- Implement a Go initial-state solver path matching `emit_init_gen`: collect
  initial constraints, add relevant definitions, solve, assign used state from
  the model, randomize only unconstrained state, then call `__init`.
- The sibling `ivy2cpp/initial_state.go` already contains much of the Go-side
  analysis and model conversion logic; factor shared pieces or port the same
  algorithm into Go emission.
- If a construct cannot yet be solved or represented in generated Go, fail
  generation with a source location instead of emitting a probabilistic retry.

Regression test:

- Use a small spec with an initial axiom that determines a relation through a
  derived predicate and is unlikely to be satisfied by random retry. As a fast
  test, inspect generated initialization code and/or the generation-time model
  assignments to prove the random retry path is not used for satisfiable initial
  constraints.

Progress:

- 2026-09-22: Added
  `TestInitialAxiomThreeValueDisequalityUsesSolverModelFast`. When the existing
  initializer-action construction leaves unresolved retry formulas, generated Go
  now falls back to the same generation-time SMT model shape used by
  `ivy2cpp`: solve the initial constraints, assign concrete model values for
  used state symbols, and keep normal nondeterministic initialization for
  unconstrained state. This removes the 1000-attempt runtime retry loop for
  satisfiable constraints such as `saved != red` over a three-value enum. Full
  initial-state parity for every solver expression shape still depends on the
  model-conversion coverage of this fallback.
- 2026-09-22: Added
  `TestTargetGenInitialAxiomThreeValueDisequalityUsesSolverModelFast`. The same
  generation-time initial model fallback now applies to `target=gen` when
  finite initial constraints exist and the initializer-action analysis emits no
  actions, so satisfiable axioms are no longer silently ignored by the gen init
  generator. The existing target=gen hard failure for unenumerable quantified
  initial variables is preserved.

## 4. `modelfile` is accepted but not semantically implemented

Python source behavior:

- Command-line parsing opens `__ivy_modelfile`
  (`ivy_to_cpp.py:2813-2815`).
- `ivy_z3_gen.hpp` writes solver checks, unsat-core pruning, SAT solver state,
  and models when the file is open (`include2cpp/ivy_z3_gen.hpp:504-540`).

Current Go behavior:

- The generated Go parser accepts `modelfile` and creates the file, but the file
  handle is not stored or used by any generator logic
  (`ivy2golang/generator.go:1782-1818`).
- Existing tests only assert file creation, not contents.

Risk:

- Users get a successful run and an empty/useless model file. This hides solver
  parity bugs and breaks a documented debugging surface of the generated tester.

How to conform:

- Add generated runtime state for the model log writer and write the same classes
  of events that `ivy_z3_gen.hpp` writes once the solver-backed generators exist:
  begin check, check-sat assumptions, unsat core deletions, begin sat, and model.
- Until solver-backed generators land, either write a clear unsupported message
  to the model file or reject `modelfile` for targets whose model logging is not
  meaningful.

Regression test:

- Tighten `TestGeneratedTestMainHonorsSpecialRuntimeOptions` with fast
  source-shape coverage: generate a spec with at least one satisfiable action
  generator and assert the emitted code writes meaningful model-log content
  (`begin check` / `begin sat`, or the agreed Go equivalent) instead of merely
  creating a file.

Progress:

- 2026-09-22: Added fast source-shape coverage for `modelfile`. Until generated
  Go has solver-backed generators, the runtime
  now writes `ivy2golang: modelfile solver logging is not implemented for
  generated Go` to the requested file instead of silently creating an empty,
  misleading log. Full `begin check` / `begin sat` parity remains tied to the
  solver-backed generator work in items 1 and 2.

## 5. Native C++ blocks, actions, types, and definitions are silently weakened

Python source behavior:

- Top-level native blocks are antiquote-rendered and emitted into the generated
  C++ (`ivy_to_cpp.py:1468-1472` and the native dispatch around
  `ivy_to_cpp.py:2314-2377`).
- Native actions are emitted inline with antiquoted references
  (`ivy_to_cpp.py:4148-4165`).
- Native types participate in C++ type rendering and random/literal behavior
  throughout `ivy_to_cpp.py`.

Current Go behavior:

- Top-level native blocks are emitted as comments
  (`ivy2golang/generator.go:153-180`).
- Native types are represented as `int` placeholders
  (`ivy2golang/generator.go:182-195`).
- Native actions are Go no-ops (`ivy2golang/action.go:358-375`).
- Native definitions return zero values (`ivy2golang/definitions.go:51-69`).

Risk:

- A generated Go tester can compile and pass while omitting behavior that is
  present in the C++ tester. This is a false positive, not merely an unsupported
  feature.

How to conform:

- Decide on a Go-native replacement surface for Ivy native C++ blocks. If no
  semantics-preserving translation exists, make behavior-affecting native blocks
  a generation error with source location.
- Preserve comments only for explicitly non-behavioral native material, such as
  C++ includes, after classification.
- Use `ivy2cpp/native.go` as the reference for parsing/classification and add a
  Go-specific renderer or fatal diagnostics.

Regression test:

- Add a spec with a native action that changes state or a native definition that
  returns a non-zero value in C++. `ivy2golang` should either emit an equivalent
  Go hook and match the C++ trace, or fail generation with a diagnostic naming
  the native block line. It must not compile a no-op/zero-value tester.

Progress:

- 2026-09-22: Added fast rejection coverage for native action weakening,
  native type placeholder weakening, and native definition zero-value stubs.
  `ivy2golang` now fails generation for behavior-affecting native C++ actions,
  native type interpretations, and native expression definitions instead of
  emitting no-op or zero-value Go. Top-level native blocks remain comments, and
  the existing runtime socket-handle factory escape hatch is preserved only for
  native-only actions returning runtime handle sorts.

## 6. Unsupported action/expression paths can remain soft comments in generated Go

Python source behavior:

- Python intentionally comments out some unsupported assume/if expressions in
  emitted C++ (`ivy_to_cpp.py:3861-3875`, `ivy_to_cpp.py:3988-4005`), but many
  behavior-affecting paths raise `IvyError` during generation.

Current Go behavior:

- `unsupported` emits a generated `panic` and records an error
  (`ivy2golang/generator.go:3600-3631`).
- `softUnsupported` only writes a comment into the generated Go
  (`ivy2golang/generator.go:3633-3654`).
- Several assumption/condition emitters call `softUnsupported`
  (`ivy2golang/action.go:532-565`).

Risk:

- A generated tester may compile while dropping an assume/if condition that
  affected enabledness or safety. This is especially dangerous in `target=test`,
  where the solver generator is supposed to decide enabledness before execution.

How to conform:

- Keep a narrow compatibility list of Python-soft constructs. For everything
  else, especially exported actions, target=test/gen guard generation, and
  runtime assumptions, fail generation with a source-location diagnostic.
- Once solver-backed generation is available, unsupported precondition
  translation should be fatal before any tester source is considered valid.

Regression test:

- Add a spec with an unsupported expression inside an exported action's assume or
  if condition. Assert `Generate` returns an error with the Ivy source line rather
  than emitting a compilable Go program containing only an
  `ivy2golang: unsupported ...` comment.

Progress:

- 2026-09-22: Added `TestUnsupportedExportedAssumeAndIfConditionsAreFatal`.
  Unsupported non-compat exported action conditions now fail generation with
  source locations instead of compiling away behavior as comments. The existing
  Python/C++-compatibility path for unenumerable quantified variables remains
  soft and source-located, preserving the Hermes-style fallback behavior while
  preventing native expressions and other unsupported conditions from silently
  weakening tests.

## 7. Generated randomness ignores Python's call-stack-qualified choice labels

Python source behavior:

- The generated class stores `___ivy_stack` (`ivy_to_cpp.py:2063`).
- Calls in `target=gen/test` push and pop action unique IDs
  (`ivy_to_cpp.py:3901-3953`).
- `___ivy_choose` appends the stack IDs to the choice key before delegating to
  the generator (`ivy_to_cpp.py:2328-2335`).

Current Go behavior:

- Generated `___ivy_choose` and `___ivy_randomize` discard `name` and `id`
  entirely and draw directly from the global RNG
  (`ivy2golang/generator.go:1131-1144`).
- There is no emitted `___ivy_stack` equivalent in `ivy2golang`.

Risk:

- Even when the top-level action schedule matches C++, nested nondeterministic
  choices can drift because Python's generator keys choices by call path. That
  affects deterministic conformance and replayability.

How to conform:

- Emit an Ivy call stack for generated Go and push/pop around calls in
  `target=gen/test`.
- Route choices through the solver/generator state once that exists. If the Go
  generator remains direct-RNG for some target, include the same stack-qualified
  label in the deterministic choice stream.

Regression test:

- Use two exported actions that call the same helper containing a local nondet
  choice. With a fixed seed, compare the Go and C++ traces/state observations.
  The test should fail if the helper's choices are shared only by RNG order
  instead of stack-qualified call context.

Progress:

- 2026-09-22: Added
  `TestGeneratedChoicesUseCallStackQualifiedLabelsFast`. Generated Go now
  stores `__ivy_stack`, deep-copies it for trial clones, pushes/pops labels
  around top-level and nested action calls, builds stack-qualified choice labels,
  and mixes those labels into the direct RNG choice/randomize path. This stops
  `___ivy_choose` / `___ivy_randomize` from discarding `name` and `id`.
  Solver-generator replay parity is still part of the open items 1 and 2.

## 8. The generated test loop omits reader/timer event-loop semantics

Python source behavior:

- `emit_repl_boilerplate3test` binds readers, chooses action generators, and
  also runs a `select` loop over readers and timers. Timeouts can decrement or
  re-increment the cycle counter; background readers set `do_over`
  (`ivy_to_cpp.py:4350-4552`).

Current Go behavior:

- `emitTestMain` only creates an Ivy object, runs randomized action cycles,
  calls `__tick(0)` after each cycle, optionally sleeps, finalizes, and exits
  (`ivy2golang/generator.go:2491-2517`, `ivy2golang/generator.go:2980-3219`).
- Generated `__tick` ignores the timeout except for progress checks
  (`ivy2golang/generator.go:1671-1677`).

Risk:

- Specs that rely on imported callbacks, readers, timers, or the Python/C++
  event-loop timing contract will not be exercised in the same way. This is a
  compositional-testing behavior gap, not just a missing convenience.

How to conform:

- Port the reader/timer runtime surface needed by test target output. The Go
  event loop should preserve Python's cycle accounting: generator SAT executes,
  UNSAT decrements the cycle, empty timeout decrements unless a timer fires, and
  background reads set the do-over path.
- Add finalizer locking/order parity while doing this work; Python calls
  `ext:_finalize` under lock before the final wait/completion sequence.

Regression test:

- Add a small Ivy spec with a timer-driven exported action or imported callback
  that changes trace-visible state without a normal exported action firing. As a
  fast test, inspect the emitted event-loop code and assert it contains the
  reader/timer scheduling and cycle-accounting paths needed for parity.

Progress:

- 2026-09-22: Added fast source checks for the concrete runtime pieces present
  in `ivy2golang`: the randomized test loop now calls `ivy.__tick(sleepMs)`
  instead of hard-coding `0`, and exported `ext:_finalize` runs between
  generated `ivy.__lock()` / `ivy.__unlock()` stubs before final wait and
  `test_completed`. Full reader/timer `select` parity remains open until the
  Go runtime has generated reader/timer objects to schedule.

## 9. `before_export` analysis is only partially ported

Python source behavior:

- `emit_action_gen` substitutes `before_export[name]` before computing the
  solver precondition, but `execute` still calls the public exported action
  (`ivy_to_cpp.py:1216-1217`, `ivy_to_cpp.py:1341-1360`).

Current Go behavior:

- `actionGeneratorAnalysisAction` does select `BeforeExport` for the syntactic
  guard analysis (`ivy2golang/generator.go:2604-2611`).
- Because the Go analysis is not the reverse-image solver analysis, only leading
  assumes and simple defined inputs in `before_export` affect generation.

Risk:

- Any `before_export` behavior beyond the current syntactic subset is ignored
  for enabledness/input generation even though Python uses it to drive the full
  action generator.

How to conform:

- Fold `before_export` into the solver-backed action generator plan before
  update/reverse-image, just as Python does.

Regression test:

- Add an exported action whose public body is broad but whose `before_export`
  body constrains an input through a non-leading assume or derived definition.
  Assert the generated Go chooses the constrained input and executes the public
  action, matching `ivy2cpp`.

Progress:

- 2026-09-22: Added
  `TestTargetTestBeforeExportNonLeadingAssumeUsesPreimageFast`. The
  `before_export` action now feeds the same target=test preimage guard path for
  the simple assignment-before-assume shape, while execution still calls the
  public exported action after the trial succeeds. `target=gen` already has
  fast coverage for leading `before_export` assumes, and the item 2 fix gives it
  the same simple preimage slice. Full `before_export` reverse-image analysis is
  still part of the solver-backed generator work.

## 10. Existing parity tests should be expanded from source shape to oracle traces

Current state:

- There are useful source-shape tests and some C++ oracle comparisons in
  `ivy2golang/ivy2golang_test.go`, but several divergence-sensitive surfaces are
  only tested for source presence or file creation.

Recommended direction:

- For every fix above, add one fast in-process unit test that checks the
  generator analysis or emitted source shape.
- Prefer tiny inline specs over large external examples. Use larger regressions
  only as final end-to-end sentinels.

Short test pattern:

- Fast: `compileIvySource`, `Generate`, inspect the relevant generated function
  body with `bodyAfterMarker`, and assert the old divergent construct is absent.
- Final end-to-end verification: after all items are marked fixed, build and run
  generated testers as needed with matching `iters/runs/seed/delay`, and compare
  stdout/stderr plus any model log needed for the feature.

Progress:

- 2026-09-22: Added fast focused regressions for the fixed slices above:
  target=test per-action generators and preimage guards, target=gen preimage
  guards, initial-state SMT fallback for target=test and target=gen, native
  weakening rejection, fatal unsupported exported conditions, stack-qualified
  choices, modelfile unsupported-content markers, finalizer lock ordering, and
  before_export preimage handling. The normal `ivy2golang` and
  `cmd/ivy2golang` test run remains under five seconds; C++ oracle expansion is
  left to the final `SLOWTEST=1` pass.
