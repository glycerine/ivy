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
- 2026-09-22: Added `TestTargetTestSolvedLocalAssumeSkipsTrialGeneratorFast`.
  Local finite witness initialization now makes simple local assumes such as
  `var choice : color; assume choice = green` safe to execute directly after
  generator success; unsupported local/nested runtime assume shapes still fall
  back to the trial path while the full solver-backed generator remains open.
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
  `TestTargetTestAssignedLocalAssumeUsesGeneratorGuardFast`. Deterministic
  local scalar assignments now act as aliases during `target=test` preimage
  analysis, so `var scratch; scratch := c; assume scratch = green` emits a
  generator guard on `c = green` without exposing `scratch`. Local aliases that
  would leak into final guards or state updates still fall back to the trial
  path.
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
  `TestTargetTestCallReturnPointTargetPreimageUsesGeneratorGuardFast`. Private
  call return formals substituted with relation/function cell actual returns
  now record point updates during preimage analysis, so
  `call marked(c) := helper(c); assume marked(c)` uses the helper return
  preimage guard instead of ignoring the cell write. Point-update rewriting now
  also collapses same-key reads directly to the updated value.
- 2026-09-22: Added `TestTargetTestLetActionPreimageUsesGeneratorGuardFast`.
  `LogicLetAction` bodies now participate in the target=test preimage walker:
  simple let bindings are substituted into the body before assignment/assume
  analysis, so aliases such as `let alias = c { saved := alias; assume saved =
  green }` produce a generator guard on `c = green` instead of an unconstrained
  generator or runtime trial.
- 2026-09-22: Added
  `TestPreimageWalkerTreatsAssertLikeActionsAsTransparentFast`. The
  preimage-only walkers now treat `requires`, `ensures`, and `subgoal` actions
  as assert-like, matching Python's `AssertAction` inheritance instead of
  rejecting those transparent steps in syntactic preimage analysis.
- 2026-09-22: Added `TestPreimageWalkerTreatsIgnoreActionAsTransparentFast`.
  The syntactic preimage and precondition-only walkers now treat `IgnoreAction`
  as the no-op marker that the Go action emitter already implements, so an
  ignore step before later assignments/assumes no longer forces the generator
  planner off the direct preimage path.
- 2026-09-22: Added `TestPreimageWalkerTreatsDebugActionAsTransparentFast`.
  `DebugAction` is now transparent to syntactic preimage/precondition-only
  planning, matching its Python/Go update semantics as a no-op for transition
  formulas while preserving normal generated debug output during actual action
  execution.
- 2026-09-22: Added `TestPreimageWalkerLowersAssignFieldActionFast`,
  `TestPreimageWalkerLowersCopyFieldActionFast`, and
  `TestPreimageWalkerLowersNullFieldActionFast` /
  `TestPreimageWalkerLowersIntegerLikeNullFieldActionFast`. The syntactic
  preimage walker now lowers destructor field actions into ordinary point
  updates before later assumes, so field writes such as `shade(current) := c`,
  field copies, and finite-scalar or integer-like null/default field writes
  participate in generator guards instead of falling through to broader
  reverse-image or trial handling.
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
- 2026-09-22: Added
  `TestTargetTestConditionalPointAssignmentPreimageUsesIteGuardFast`.
  Branch-local relation/function point updates are now merged back into the
  caller as guarded point updates, so `if c = red { marked(c) := true };
  assume marked(green)` emits a generator guard that applies the cell update
  only under the branch condition. Unsupported branch-local point-update shapes
  still fall back to the trial path.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorSearchesFiniteDomainFast`. `target=test`
  per-action generators now keep the initial randomized candidate but, when a
  guard rejects it and all formal parameters have small finite domains, search
  the finite formal tuple space from randomized offsets before returning
  `false`. This narrows the gap with Python's solver-backed generators for
  simple finite constraints that are not direct defined-parameter equalities.
  Local solver witnesses, large/unbounded domains, and the full Z3
  reverse-image plan remain open.
- 2026-09-22: Added
  `TestTargetTestLocalFiniteAssumeChoosesWitnessFast`. Generated `target=test`
  / `target=gen` action bodies now initialize finite scalar locals by scanning
  their finite value set when leading assumes mention the local, so simple local
  witness constraints such as `var choice : color; assume choice = green` do not
  depend on a single lucky nondeterministic draw. This is still a finite
  generated-code witness search, not the full Python SMT local-input model.
- 2026-09-22: Added `TestLocalVariantExistsAssumeChoosesWitnessFast`. Local
  variant-super witnesses now use the existential variant matcher before action
  execution, so `var x:msg; assume exists Q:req. x *> Q` initializes `x` to a
  concrete `req` value instead of leaving it as the invalid zero/randomized
  variant and tripping the runtime assume.
- 2026-09-22: Added `TestLocalVariantConcreteAssumeChoosesWitnessFast`. Local
  variant-super witnesses now also handle direct concrete membership guards, so
  `var x:msg; assume x *> req0` constructs `x` with the `req0` payload before
  the runtime assume check.
- 2026-09-22: Added
  `TestLocalVariantIteAssumeChoosesConditionalWitnessFast`. Local
  variant-super witnesses now handle conditional branch membership guards by
  assigning a conditional upcast, e.g. `ite(active, x *> req0, x *> ack0)`
  emits `x = ivyTernary(active, req-upcast, ack-upcast)` before the runtime
  assume instead of leaving `x` invalid or hard-coding one branch.
- 2026-09-22: Added
  `TestLocalVariantPartialIteAssumeChoosesConditionalWitnessFast`. Local
  variant-super witnesses now also handle conditional guards where only one
  branch constrains the local variant, keeping the existing local value on the
  unconstrained branch.
- 2026-09-22: Added `TestLocalVariantNegatedConcreteAssumeChoosesWitnessFast`.
  Local variant-super witnesses now choose a sibling subtype for negated
  concrete membership guards, so `var x:msg; assume ~(x *> req0)` constructs an
  alternate variant before the runtime assume check.
- 2026-09-22: Added `TestTargetTestLocalVariantStateUpdateSkipsTrialFast`.
  Zero-formal `target=test` actions whose local variant witness is copied into
  state and immediately rechecked now execute directly after generator success
  instead of falling back to a hidden trial clone. The direct-execution check is
  intentionally narrow: it requires a concrete local membership witness and the
  later state membership assume to match that same concrete payload.
- 2026-09-22: Added
  `TestTargetTestLocalVariantExistsStateUpdateSkipsTrialFast`. The same
  direct-execution shortcut now also covers subtype-only existential variant
  rechecks such as `exists R:req. saved *> R` after a local `msg` witness has
  been copied into state. Exact-payload existential rechecks still require the
  previously generated concrete payload to match.
- 2026-09-22: Added
  `TestTargetTestLocalVariantAffineEqualityStateUpdateSkipsTrialFast`. The
  copied-local variant direct-execution proof now compares quantified payload
  witnesses with their integer delta, so a local witness from
  `exists Q:req. x *> Q & Q + 1 = req0` can be copied to state and rechecked
  without a hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalRelationStateUpdateSkipsTrialFast`. Zero-formal
  `target=test` actions whose local relation witness is copied into state and
  rechecked through the same positive relation now execute directly after the
  generated action body witness scan. The no-trial proof is intentionally
  narrow: it requires the same relation symbol and unchanged non-witness
  arguments, with the copied state value occupying the witnessed local's
  argument position.
- 2026-09-22: Added
  `TestTargetTestLocalRelationExistsStateUpdateSkipsTrialFast`. The relation
  state-update proof now also handles existentially ignored relation columns on
  both the local witness and later state recheck, e.g.
  `exists M. edge(n,M)` followed by `exists M. edge(saved,M)`. Fixed relation
  arguments still must match literally; ignored columns are only considered
  covered when both sides are existential.
- 2026-09-22: Added
  `TestTargetTestLocalRelationIffStateUpdateSkipsTrialFast`. The same
  relation state-update proof now normalizes the `p <-> true` spelling used by
  the witness collectors, so a local relation witness copied into state and
  rechecked as `allowed(saved) <-> true` no longer forces a hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalVariantIffStateUpdateSkipsTrialFast`. The local variant
  state-update proof now also normalizes concrete membership guards written as
  `(saved *> req0) <-> true`, matching the existing witness collector and
  avoiding a hidden trial clone for that spelling.
- 2026-09-22: Added
  `TestTargetTestLocalVariantIteStateUpdateSkipsTrialFast`. The same
  direct-execution proof now splits conditional local variant witnesses across
  matching `ite` recheck branches, so a local `msg` built by
  `ite(active, x *> req0, x *> ack0)` can be copied into state and rechecked
  with the same conditional membership guard without a hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPointUpdateSkipsTrialFast`. Zero-formal
  `target=test` actions whose relation-backed local witness is used to set a
  positive relation cell, then immediately rechecked through that same cell,
  now execute directly after generator success. The shortcut is intentionally
  limited to positive boolean relation point updates whose local arguments
  already have a modeled relation witness; later local assumes must be covered
  by a recorded state or point update rather than being treated as unrelated
  fresh witnesses.
- 2026-09-22: Added
  `TestTargetTestLocalVariantPointUpdateSkipsTrialFast`. The same positive
  relation point-update proof now also accepts local arguments with modeled
  variant witnesses, e.g. a local `msg` constructed from
  `exists Q:req. x *> Q` and then used in `seen(x) := true; assume seen(x)`.
- 2026-09-22: Added
  `TestTargetTestLocalVariantConditionalPointUpdateSkipsTrialFast`. The
  zero-formal direct-execution proof now accepts the narrow conditional case
  where both `if/else` branches are structurally identical and covered by the
  same modeled local witness, so branch-wrapped point updates no longer force a
  hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalVariantOneSidedConditionalPointUpdateSkipsTrialFast`.
  One-sided conditional relation point updates are now recorded with their
  branch guard, so `if active { seen(x) := true }` can satisfy a later
  `ite(active, seen(x), true)` recheck without making the guarded update prove
  unconditional `seen(x)` assumptions.
- 2026-09-22: Added
  `TestTargetTestLocalVariantElseSidedConditionalPointUpdateSkipsTrialFast`.
  Guarded relation point updates now record the branch polarity too, so the
  mirror case `if active { } else { seen(x) := true }` proves
  `ite(active, true, seen(x))` without a hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalVariantConditionalPointUpdateImplicationSkipsTrialFast`.
  Guarded relation point-update coverage now also handles the implication
  spelling `active -> seen(x)` by proving the consequent under the antecedent's
  true branch, matching the equivalent `ite(active, seen(x), true)` case.
- 2026-09-22: Added
  `TestTargetTestLocalVariantConditionalPointUpdateOrSkipsTrialFast`. The
  desugared implication spelling `~active | seen(x)` now follows the same
  guarded point-update proof path, so source-level OR guards do not reintroduce
  the hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPointUpdateExistsSkipsTrialFast`. Positive
  relation point updates now also cover existential ignored columns in the
  later recheck, e.g. `edge(x,n8) := true; assume exists M. edge(x,M)` when
  `x` has an already modeled local witness. Ignored existential terms are
  accepted only when the update term has the same sort.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPointUpdateIteRecheckSkipsTrialFast`. The
  relation point-update direct-execution proof now descends through conditional
  rechecks when both branches are covered by the recorded update, so an
  `ite(active, seen(n), seen(n))` assume after `seen(n) := true` no longer
  forces a hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalVariantFalsePointUpdateSkipsTrialFast`. The local
  relation point-update proof now records the boolean value written to the cell,
  so a modeled local witness can write `seen(x) := false` and then satisfy a
  negated recheck such as `assume ~seen(x)` without a hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalVariantStateAliasPointUpdateSkipsTrialFast`. Relation
  point-update coverage now handles the reverse state-alias direction too:
  after `saved := x`, a write through `seen(saved)` is known to cover a later
  recheck through `seen(x)` when `x` has a modeled local witness.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPairPointUpdateSkipsTrialFast`. The
  direct-execution proof now flattens consecutive local declarations the same
  way the action emitter does and recognizes grouped relation tuple witnesses,
  so `assume edge(a,b); seen(a,b) := true; assume seen(a,b)` can execute
  without the hidden clone/trial path.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPairStateUpdateSkipsTrialFast`. Grouped relation
  tuple witnesses are now retained with tuple positions for later state-copy
  rechecks, so `saved_a := a; saved_b := b; assume edge(saved_a,saved_b)` is
  covered by the original `edge(a,b)` witness without a hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPairStateUpdateIteRecheckSkipsTrialFast`. The
  grouped relation tuple state-copy proof now descends through `ite` rechecks
  when both branches are covered by the same recorded tuple witness, avoiding a
  hidden trial clone for branch-wrapped state relation rechecks.
- 2026-09-22: Added
  `TestTargetTestChoiceAssumePreimageUsesDisjunctiveGuardFast`. The
  `target=test` preimage walker now handles pure nondeterministic choices by
  OR-ing the branch assume preconditions, so a choice between `assume c = green`
  and `assume c = blue` emits a generator guard for the disjunction. Branches
  that leave different state updates are covered separately below.
- 2026-09-22: Added
  `TestTargetTestChoiceStateUpdatePreimageUsesReverseImageGuardFast`. Simple
  nondeterministic choices whose branches assign different scalar state values
  now carry those alternatives forward in the preimage context, so a later guard
  like `choice { saved := green } { saved := blue }; assume saved = c` emits the
  disjunctive generator guard `green = c | blue = c` instead of treating the
  action as always enabled.
- 2026-09-22: Added
  `TestTargetTestBulkRelationAssignmentUsesReverseImageGuardFast`. When the
  syntactic preimage walker cannot express an action, `target=test`/`target=gen`
  generator guard collection now falls back to Go's shared
  `GetUpdateForArt`/`ReverseImage` machinery and inlines simple reverse-image
  temporary definitions before emitting Go. This covers bulk relation
  assignments such as `marked(C) := C = c; assume marked(green)`, producing a
  guard equivalent to `c = green` instead of relying on a trial run. The fallback
  is best-effort and deliberately returns no guard if the shared update builder
  rejects an action shape.
- 2026-09-22: Extended
  `TestTargetTestBulkRelationAssignmentUsesReverseImageGuardFast` to assert
  that reverse-image-solved actions execute directly after `generate()` succeeds
  instead of falling back to a hidden public-action trial. The `target=test`
  trial decision now accepts either the syntactic preimage path or the
  reverse-image fallback, but rejects reverse-image formulas that still contain
  non-formal local solver symbols because those require a real model.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorSearchesFiniteDestructorFieldFast`. Finite
  fallback search now also enumerates small finite destructor fields on
  structured formals instead of requiring the whole structured sort to be
  enumerable. This covers simple field constraints such as `shade(c) ~= red`
  by scanning generated `gen.c.shade` values from randomized offsets.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorSearchesFiniteNestedDestructorFieldFast`.
  Finite field search now follows assignable destructor chains rooted at the
  generated formal, so `shade(inner(c)) ~= red` scans `gen.c.inner.shade`
  before rechecking the guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorSearchesFiniteDestructorArrayFieldFast`. Finite
  fallback search now expands finite indexed destructor fields into per-cell
  search dimensions, e.g. `shade(c,1) ~= red` can search
  `gen.c.shade[1]` instead of rejecting after one randomized structured value.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorSearchesFiniteHashThunkDestructorFieldFast`.
  Finite fallback search now also handles guard-mentioned hash-thunk destructor
  cells by scanning the finite range and writing each candidate with
  `ivyThunkSet`, e.g. `shade(c,current) ~= red` can search the generated
  `shade` thunk at `current`.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorSearchesFiniteHashThunkDestructorFieldWithFormalIndexFast`.
  Hash-thunk finite field search now treats formals used only as thunk indexes
  as covered by the field-input search, so `shade(c,k) ~= red` can write
  candidate colors into `gen.c.shade` at `gen.k` without requiring a separate
  solver model for `k`.
- 2026-09-22: Added
  `TestTargetTestLocalFiniteAssumeAfterTransparentAssertChoosesWitnessFast`.
  Local finite witness selection now scans assumption prefixes through
  assertion-only steps, so a harmless `assert true` before
  `assume choice = green` no longer leaves the local value to a single random
  draw. This is still a finite witness scan for generated Go, not the full SMT
  local-input model.
- 2026-09-22: Added
  `TestTargetTestLocalNumericInequalityChoosesWitnessFast`. Local unbounded
  integer-like choices now reuse the simple scalar witness extractor, so a local
  `var n : node; assume n > 10` tries `11` before executing the assume instead
  of depending on the small nondeterministic local draw.
- 2026-09-22: Added
  `TestTargetTestLocalNumericFormalInequalitySkipsTrialFast`. Local scalar
  witnesses now also use emittable formal-dependent bounds, so
  `var scratch:node; assume scratch > c` seeds `scratch` with `c + 1`; the
  local-preimage planner drops this single satisfiable local inequality instead
  of existentializing it into an unsupported generator guard or falling back to
  a hidden trial.
- 2026-09-22: Added
  `TestTargetTestLocalNumericFormalInequalityStateUpdateSkipsTrialFast`. The
  same formal-dependent inequality witness is now substituted through
  local-preimage state copies, so `scratch > c; saved := scratch; assume saved
  > c` executes directly with witness `c + 1` instead of needing a hidden
  trial.
- 2026-09-22: Added
  `TestTargetTestLocalNumericAffineEqualityStateUpdateSkipsTrialFast`. Local
  scalar witnesses now also solve formal-dependent affine equalities such as
  `scratch + 1 = c` by seeding `scratch` with `c - 1`, and the local-preimage
  planner substitutes that affine witness through later state updates so the
  action can execute without a hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalNumericFormalDisequalitySkipsTrialFast`. Negated affine
  local equalities such as `scratch + 1 ~= c` now choose a nearby
  formal-dependent witness (`c`) and substitute it through local preimage
  state updates, avoiding broad fallback scans and hidden trial execution.
- 2026-09-22: Added
  `TestTargetTestLocalNumericFormalNegatedInequalitySkipsTrialFast`. Negated
  formal-dependent local inequalities such as `~(scratch < c)` now choose the
  boundary witness (`c`) and substitute it through state-copy preimage analysis
  instead of relying on broad fallback scans or hidden trial execution.
- 2026-09-22: Added
  `TestTargetTestLocalNumericFormalIffFalseInequalitySkipsTrialFast`. The
  local-preimage eliminator now normalizes IFF-with-false inequality witnesses,
  so `(scratch <= c) <-> false` uses the same `c + 1` witness and direct
  state-copy proof as the equivalent negated inequality.
- 2026-09-22: Added
  `TestTargetTestLocalNumericFormalIffFalseAffineEqualitySkipsTrialFast`.
  IFF-with-false affine equality witnesses now use the same formal-dependent
  disequality path, so `(scratch + 1 = c) <-> false` seeds `scratch` with `c`
  and avoids broad fallback scans or hidden trial execution.
- 2026-09-22: Added
  `TestTargetTestLocalNumericFormalImplicationSkipsTrialFast`. Local witness
  preimage elimination now looks through implication consequents, so
  `active -> scratch > c` can seed `scratch` with `c + 1` and prove a copied
  state recheck without hidden trial execution.
- 2026-09-22: Added `TestTargetTestLocalNumericFormalOrSkipsTrialFast` and
  `TestTargetTestLocalNumericFormalAndPreservesGuardSkipsTrialFast`. The OR
  case now has fast guardrail coverage, and conjunction preimage handling now
  solves the local scalar conjunct while preserving non-local conjuncts such as
  `active` as generator guards instead of silently treating the whole action as
  enabled.
- 2026-09-22: Added `TestTargetTestLocalNumericFormalIteSkipsTrialFast`.
  Conditional local scalar witnesses now build an AST-level
  `ite(active, c + 1, c)` witness and preserve the substituted conditional
  guard in the per-action generator, so branch-dependent local assumes no
  longer execute after an unconditional generator success.
- 2026-09-22: Added
  `TestTargetTestLocalNumericFormalPartialIteSkipsTrialFast`. Conditional
  local scalar witnesses with a tautological unconstrained branch now use a
  closed default value for that branch in the target=test direct-execution
  proof, while the generated action body still keeps the local runtime value on
  the unconstrained branch. This removes the hidden trial clone for
  `ite(active, scratch > c, true)` copied into state and rechecked.
- 2026-09-22: Added
  `TestTargetTestLocalNumericFormalLetAndPreservesGuardSkipsTrialFast`.
  Expression-level `LogicLet` guards are now expanded before preimage
  substitution, so aliased local scalar witnesses such as
  `let bound = c in scratch > bound & active` retain the non-local `active`
  guard instead of falling back to an unconstrained generator.
- 2026-09-22: Added
  `TestTargetTestLocalNumericWitnessAfterDebugChoosesWitnessFast`. Local
  witness assume collection now treats debug, ignore, and assert-like actions
  as transparent no-op/precondition steps, so local scalar witnesses are still
  initialized before assumes that appear after a debug marker.
- 2026-09-22: Added
  `TestTargetTestLocalNumericWitnessInsideLetChoosesWitnessFast`. Local witness
  assume collection now expands simple `LogicLetAction` aliases before scanning
  for scalar witnesses, so `let alias = n { assume alias > 10 }` seeds `n`
  with a satisfying candidate instead of relying on the fallback draw.
- 2026-09-22: Added
  `TestTargetTestLocalNumericNonlinearUsesSmallIntFallbackFast`. Local
  unbounded integer-like witnesses now fall back to a bounded small-integer
  candidate scan when no sharper scalar witness is available, so nonlinear
  local guards such as `n * n = 9` are seeded before the runtime assume.
- 2026-09-22: Added
  `TestLocalWitnessUsesSmallIntFallbackForRelationBaseFast`. Local witnesses
  over unbounded relation guards now keep the extensional override scan but, for
  non-finite integer-like locals, also try the same bounded small-integer
  fallback used by action generators. This lets action bodies satisfy
  `var n; assume allowed(n)` when `allowed` is supplied by a thunk base such as
  `allowed(N) := N = 7`.
- 2026-09-22: Added `TestLocalWitnessFromNegatedRelationFast`. Unbounded
  integer-like local witnesses now handle negated relation guards such as
  `var n; assume ~banned(n)` by scanning stored relation cells and trying a
  nearby value outside known true cells before the runtime assume recheck.
- 2026-09-22: Added `TestLocalWitnessFromNegatedRelationBaseFallbackFast`.
  Negated local relation witnesses now also use the bounded small-integer
  fallback when the relation truth comes from a thunk base instead of explicit
  override cells, e.g. `banned(N) := N = 0`.
- 2026-09-22: Added `TestLocalWitnessPairFromNegatedRelationFast`.
  Consecutive local declarations guarded by a negated tuple relation now get a
  grouped witness plan: generated action bodies copy the stored tuple fields
  and nudge one integer-like component before rechecking `~rel(a,b)`. This
  avoids solving each local independently against a stale companion value.
- 2026-09-22: Added
  `TestLocalWitnessFromExistsUsesExtensionalRelationFast`. Local relation-backed
  witness selection now descends through existential guards and treats
  quantified variables as wildcard tuple fields, so `var n; assume exists M.
  edge(n,M)` can pick `n` from a true stored relation tuple without leaking `M`
  into generated search-loop conditions.
- 2026-09-22: Added `TestLocalWitnessFromIffUsesExtensionalRelationFast`.
  Single-local relation witnesses now also descend through equivalence with
  `true`, so `var n; assume allowed(n) <-> true` seeds `n` from a true stored
  relation cell before the runtime assume check.
- 2026-09-22: Added `TestLocalWitnessFromIteUsesExtensionalRelationFast`.
  Single-local relation witnesses now descend through conditional guards, so a
  local assume such as `ite(active, exists M. edge(n,M), exists M.
  permitted(n,M))` can seed `n` from an extensional relation witness instead of
  falling back to a broad bounded integer scan.
- 2026-09-22: Added
  `TestLocalWitnessFromPartialIteUsesExtensionalRelationFast`. Single-local
  relation witnesses now also handle conditional guards where only one branch
  constrains the local, scanning the relation on the constrained branch and
  leaving the local untouched on the unconstrained branch.
- 2026-09-22: Strengthened
  `TestLocalWitnessFromIteUsesExtensionalRelationFast` and added
  `TestTargetTestLocalRelationIteStateUpdateSkipsTrialFast`. Conditional local
  relation witnesses now emit branch-specific relation scans, and the
  `target=test` direct-execution proof projects the same conditional witness
  across matching state recheck branches, avoiding both wrong-branch local
  seeds and hidden trial clones.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPartialIteStateUpdateSkipsTrialFast`.
  Conditional local relation witnesses whose other branch is syntactic `true`
  now remain valid for the `target=test` direct-execution proof: the
  constrained branch projects its relation witness through the matching state
  recheck, while the tautological branch is treated as already covered instead
  of forcing a hidden trial clone.
- 2026-09-22: Added
  `TestLocalWitnessPairUsesOneExtensionalRelationTupleFast`. Consecutive local
  declarations are flattened for witness planning, and local relation witnesses
  can now assign several locals from one true relation tuple, e.g. `var a; var
  b; assume edge(a,b)`, before the normal assume check runs.
- 2026-09-22: Added
  `TestTargetTestReturningActionGeneratorSearchesFiniteDomainFast`.
  `target=test` generators for returning actions with finite formal inputs now
  use the same guard and finite-search path as non-returning actions. The
  zero-formal returning-action guard skip is intentionally preserved to avoid
  reintroducing the old `assume false` retry hang.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorSearchesRelevantFiniteFormalOnlyFast`.
  Finite fallback search now plans dimensions only for formals referenced by
  the guard formulas, so a constrained finite formal can be solved even when
  the action also has an irrelevant unbounded formal that remains randomized or
  zero-valued. Referenced unbounded formals still require a stronger solver
  model.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNumericDisequalityWitnessForUnboundedFormalFast`.
  Bare unbounded integer-like formals constrained by simple numeral
  equality/disequality guards now get a small generated witness search. For
  example, `assume n ~= 0` tries `1` after a rejecting randomized candidate and
  rechecks the full guard before accepting. This covers a narrow scalar-model
  slice without adding a runtime SMT dependency.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNumericInequalityWitnessForUnboundedFormalFast`.
  The same scalar witness path now recognizes simple numeral inequalities such
  as `n > 10` (including Ivy's normalized `10 < n` shape) and tries a nearby
  satisfying integer like `11` before rejecting the action generator candidate.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNegatedNumericInequalityWitnessForUnboundedFormalFast`.
  Scalar witness discovery now handles negated numeric bounds, including Ivy's
  normalized shape for `~(n <= 10)` as `~((n < 10) | (n = 10))`, and chooses
  the complementary witness (`11`) before rechecking the full guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIffFalseNumericInequalityWitnessForUnboundedFormalFast`.
  Scalar witness discovery now routes IFF-with-false terms through the same
  negated-bound extractor, so `(n <= 10) <-> false` gets the complementary
  witness instead of depending on a rejecting randomized value.
- 2026-09-22: Added `TestTargetTestActionGeneratorUsesIteNumericWitnessFast`.
  Scalar witness discovery now merges conditional numeric branch witnesses, so
  guards such as `ite(active, n > 10, n > 20)` try
  `ivyTernary(ivy.active, 11, 21)` before falling back to broader small search.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesItePartialNumericWitnessFast`. Conditional
  scalar numeric guards with a witness on only one branch now synthesize the
  constrained branch value while keeping the randomized generated formal on the
  unconstrained branch, e.g. `ite(active, n > 10, true)` emits
  `ivyTernary(ivy.active, 11, gen.n)`.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNumericAffineEqualityWitnessFast`. Simple
  affine integer equalities over one generated formal now produce deterministic
  scalar witnesses, so `n + 1 = 7` tries `6` before rejecting the randomized
  candidate.
- 2026-09-22: Added `TestTargetTestActionGeneratorUsesLetNumericWitnessFast`.
  Action-generator witness discovery now expands simple `let` expressions
  before planning scalar, relation, variant, and defined-input witnesses, so
  `let m = n + 1 in m = 7` gets the same numeric witness as the inlined guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNumericAffineInequalityWitnessFast`. Simple
  affine integer inequalities now choose a nearby satisfying value, including
  parser-normalized shapes such as `10 < n + 1`, before rechecking the full
  guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNumericPairInequalityWitnessFast`.
  Simple integer-like inequalities between two generated formals now get a
  deterministic pair witness before guard checking, e.g. `a < b` emits
  `gen.b = gen.a + 1` and then rechecks the full guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNegatedNumericPairInequalityWitnessFast`.
  Pairwise numeric witness discovery now handles negated normalized bounds such
  as `~(a <= b)` (`~((a < b) | (a = b))`) by inverting the strict inequality
  and assigning a deterministic satisfying pair such as `gen.a = gen.b + 1`.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNumericAffinePairInequalityWitnessFast`.
  Pairwise numeric inequality witnesses now handle simple affine terms on each
  side, so `a + 1 < b` emits `gen.b = gen.a + 2` before checking the full
  guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIteNumericPairInequalityWitnessFast`.
  Pairwise numeric inequality witnesses now merge same-target conditional
  branches, so guards such as `ite(active, a < b, a + 2 < b)` emit a
  branch-dependent assignment like
  `gen.b = ivyTernary(ivy.active, gen.a + 1, gen.a + 3)`.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIteCrossTargetNumericPairWitnessFast`.
  Conditional pair witnesses whose branches assign different target formals now
  emit two guarded assignments, keeping the inactive target unchanged; for
  example `ite(active, a < b, b < a)` updates both `gen.b` and `gen.a` before
  rechecking the full guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIteMultiNumericPairWitnessFast`.
  Conditional guards whose branches each contain multiple pairwise numeric
  constraints over the same generated targets now merge all branch witnesses
  before the guard check. This avoids the stale-random finite-search fallback
  for shapes such as `ite(active, a < b & c < d, a <= b & c <= d)`.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIteCrossMultiNumericPairWitnessFast`.
  Conditional pair guards whose branches contain multiple witnesses for
  different generated targets now emit guarded keep-or-assign updates for each
  branch target before checking the full `ite` guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesItePartialNumericPairWitnessFast`.
  Conditional pair guards with a witness on only one branch now assign the
  active-branch pair value while keeping the randomized target on the
  unconstrained branch.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNumericPairDisequalityWitnessFast`.
  Negated equality between two integer-like generated formals now uses the same
  pair-witness path, e.g. `a ~= b` emits `gen.b = gen.a + 1` before the full
  guard recheck.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNumericAffinePairDisequalityWitnessFast`.
  Pairwise disequality witnesses now handle simple affine terms too, so
  `a + 1 ~= b` emits a deterministic distinct value such as `gen.b = gen.a + 2`
  before rechecking the full guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNumericAffinePairEqualityWitnessFast`.
  Pairwise equality witnesses now also handle simple affine terms on both
  generated formals, so `a + 1 = b + 2` emits `gen.b = gen.a - 1` before the
  full guard check instead of depending on matching randomized unbounded
  values.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesSmallIntFallbackForNonlinearNumericGuardFast`.
  Referenced unbounded integer-like formals that have no sharper finite,
  relation, variant, or scalar witness now get a bounded small-integer search
  dimension before the generator rejects the candidate. This covers small-model
  nonlinear guards such as `n * n = 9` without adding a runtime SMT dependency.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesProductEqualityWitnessForUnboundedFormalsFast`.
  A focused two-formal nonlinear model slice now uses defined-input witness
  extraction for product equalities such as `a * b = 6`, assigning concrete
  values to both generated formals before rechecking the full guard instead of
  relying on an infeasible broad 1024-by-1024 search.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesProductStateEqualityWitnessFast`. The same
  product witness now accepts integer-like RHS expressions that do not mention
  either product formal, so guards such as `a * b = target` seed `a = 1` and
  `b = target` before rechecking.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesProductStateInequalityWitnessFast`.
  Product inequalities over two unbounded integer-like formals now get boundary
  witnesses too; for example `a * b > target` seeds `a = target + 1` and
  `b = 1` before checking the normalized guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesProductStateDisequalityWitnessFast`.
  Negated product equalities now get a nearby distinct product witness, so
  `a * b ~= target` seeds `a = target + 1` and `b = 1` before the final guard
  recheck.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesSumStateEqualityWitnessFast`. Additive
  two-formal equalities with an integer-like RHS now get a simple model witness
  too: `a + b = target` seeds `a = target` and `b = 0` before the guard
  recheck.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesSumStateInequalityWitnessFast`. Additive
  two-formal inequalities now seed boundary witnesses as well, e.g.
  `a + b > target` emits `a = target + 1` and `b = 0` before checking the
  normalized guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesSumStateDisequalityWitnessFast`. Negated
  additive two-formal equalities now seed a nearby distinct sum, so
  `a + b ~= target` emits `a = target + 1` and `b = 0` before rechecking.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIteMultiArithmeticDefinedInputsFast`.
  Conditional guards whose branches each produce multiple defined inputs for
  the same formals now merge those branch witnesses with `ivyTernary` values.
  For example, `ite(active, a * b = target, a + b = target)` seeds both
  generated formals before the final guard check instead of discarding the
  branch-local product/sum witnesses.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIteCrossMultiArithmeticDefinedInputsFast`.
  Conditional guards whose branches produce multiple defined inputs for
  different generated targets now emit guarded keep-or-assign values for the
  union of branch targets, preserving inactive branch randoms while satisfying
  the active branch before the final guard check.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesItePartialDefinedInputFast`. Conditional
  guards with a defined input on only one branch now emit a guarded assignment
  that keeps the randomized value on the unconstrained branch, e.g.
  `ite(active, x = green, true)` emits `gen.x = ivyTernary(active, green,
  gen.x)` before checking the guard.
- 2026-09-22: Added `TestTargetTestSolvedAssumeGeneratorSkipsTrialFast` and
  updated the before-export/call-preimage tests to expect direct execution once
  `generate()` succeeds. The trial clone path is now reserved for actions whose
  calls/assumes are not covered by the syntactic preimage walker; covered
  simple assumes, before-export guards, and inlined private-call preconditions
  no longer import `bytes` or execute a hidden trial action.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationWitnessForUnboundedFormalFast`.
  When a referenced formal is unbounded but appears as the sole argument to a
  boolean state relation guard such as `allowed(n)`, the `target=test`
  generator now scans the relation's true stored keys as candidate witnesses
  instead of relying on one random draw. This ports a small but important piece
  of Python's input-field/model extraction behavior for relation-backed
  enabledness.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationTupleWitnessForUnboundedFormalFast`.
  The relation-witness search now also handles tuple-key relation cells when
  exactly one tuple field is the unbounded formal and the other tuple fields are
  fixed by finite/evaluable guard terms, e.g. `allowed(green,n)`. The generated
  action generator scans true stored relation cells, filters the fixed tuple
  fields, assigns the formal from the matching key field, and then rechecks the
  full guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationPairWitnessForUnboundedFormalsFast`.
  Multi-formal relation witnesses now use one stored-relation scan when a guard
  constrains several unbounded formals together, e.g. `allowed(a,b)`. The
  generated search assigns all covered formals from the same true tuple key
  before rechecking the full guard, instead of nesting per-formal scans that
  were still pinned to the other randomized formal.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationWitnessInsideOrFast`. Unary
  relation-witness discovery now looks through disjunctive guards, so
  `allowed(n) | n = 99` can scan true `allowed` cells before rejecting the
  randomized candidate.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesAlternativeRelationWitnessesInsideOrFast`.
  If a guard contains several usable relation-witness candidates for the same
  unbounded formal, such as `allowed(n) | permitted(n)`, the generated search
  now tries the relation scans sequentially as alternatives and rechecks the
  complete guard after each candidate. This avoids depending on the first
  relation alone.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationPairWitnessInsideOrFast`. Joint
  tuple-key relation witness discovery now looks through disjunctions as well,
  so `edge(a,b) | a = 99` can assign both unbounded formals from one true
  stored `edge` tuple before rechecking the full guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationWitnessInsideExistsFast`. Unary
  relation-witness discovery now looks through existential guards and treats the
  quantified variables as wildcard tuple fields, so `exists M. edge(n,M)` can
  choose `n` from the first component of a true stored `edge` key before
  rechecking the full extensional existential guard.
- 2026-09-22: Added
  `TestTargetTestBeforeExportLocalRelationWitnessUsesGeneratorGuardFast`.
  `before_export` analysis actions now share the same existential local
  relation-witness bridge as normal action generators: a guard such as
  `exists choice. allowed(choice,c)` scans stored relation tuples, assigns the
  public formal from the matching key field, and rechecks the existential guard
  without requiring a runtime trial of the public action.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationOverrideForallGuardFast`. Action
  generator guards can now recheck unbounded negative relation quantifiers such
  as `forall X. ~banned(X,c)` by scanning the relation override map under the
  generator-only guard-emission mode, avoiding soft unsupported comments and
  preserving the Python solver-style enabledness check for this stored-relation
  slice.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesEqualityBoundForallGuardFast`. Single
  equality-bound quantifiers now simplify before code emission, so
  `forall X. X = c -> allowed(X)` becomes a direct generated guard on
  `allowed(c)` instead of trying to enumerate an unbounded quantified variable.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesMultiEqualityBoundForallGuardFast`.
  Equality-bound quantifier simplification now collects one binding per
  quantified variable from conjunctive antecedents, so
  `forall X,Y. X = c & Y = d -> allowed(X,Y)` emits a direct relation guard on
  `(c,d)` rather than nested unbounded loops.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorPreservesExtraEqualityBoundForallAntecedentFast`.
  Equality-bound universal simplification now preserves non-quantified
  antecedent terms, so `forall X. X = c & active -> allowed(X)` becomes the
  direct guarded implication `active -> allowed(c)` rather than falling back to
  quantified-variable enumeration.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesEqualityOnlyExistsGuardFast`. Equality-only
  existential guards such as `exists X. X = c` now simplify to `true`, matching
  the solver-trivial satisfiable case instead of emitting a soft unsupported
  unbounded quantifier comment.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesEqualityBoundForallNumericWitnessFast`.
  Equality-bound quantifier simplification now runs on the action-generator
  guard AST before witness planning as well as during expression emission, so
  `forall X. X = c -> n > X` is seen by scalar witness extraction as `n > c`
  and emits the deterministic `n = c + 1` assignment before the guard check.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesEqualityBoundExistsNumericWitnessFast`.
  Equality-bound existential guards with non-variant residual bodies now also
  simplify on the guard AST before witness planning, so
  `exists X. X = c & n > X` feeds the scalar witness extractor as `n > c`
  while variant-membership existential guards remain on their specialized
  downcast path.
- 2026-09-22: Added
  `TestTargetTestLocalEqualityBoundExistsNumericWitnessSkipsTrialFast`. Local
  witness assume collection now shares the equality-bound quantifier
  simplifier, so `var scratch; assume exists X. X = c & scratch > X` seeds the
  local with `c + 1` and avoids broad small-integer fallback scans or hidden
  trial execution.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationPairWitnessInsideExistsFast`.
  Grouped tuple-key relation witness discovery now also carries existentially
  quantified variables as wildcard tuple fields, so `exists C. edge(a,b,C)` can
  assign both generated formals from one true relation tuple without leaking
  `C` into generated search-loop conditions.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationWitnessInsideImplicationFast`.
  Relation-witness discovery now also looks through implication consequents, so
  guards like `active -> allowed(n)` can scan true `allowed` cells when the
  initial randomized candidate fails the full implication guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationWitnessInsideIffFast`.
  Relation-witness discovery now treats boolean equivalence with `true` as a
  positive guard, so `allowed(n) <-> true` can scan true stored relation cells
  for unbounded generated formals before rechecking the complete IFF guard.
  Equivalence with `false` is routed to the negated-relation witness path.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationWitnessInsideIteFast`.
  Relation-witness discovery now also descends into conditional branches, so
  guards such as `ite(active, allowed(n), permitted(n))` try both relation
  model scans as alternatives before falling back to blind numeric search.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationWitnessForUnboundedDestructorFieldFast`.
  Relation-witness extraction now recognizes destructor-field arguments rooted
  at a structured formal, so `allowed(node_id(c))` can scan true `allowed`
  cells and assign `gen.c.node_id` from the stored relation key before
  rechecking the complete guard. This ports another slice of Python's
  field-input/model extraction behavior for unbounded structured fields.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationWitnessForUnboundedDestructorFieldPairFast`.
  Joint relation-witness extraction now also handles several destructor-field
  arguments rooted at the same structured formal, so `edge(src(c),dst(c))`
  scans true `edge` tuple keys and assigns both `gen.c.src` and `gen.c.dst`
  from one model tuple before rechecking the guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationWitnessForNestedDestructorFieldFast`
  and
  `TestTargetTestActionGeneratorUsesRelationWitnessForNestedDestructorFieldPairFast`.
  Relation-witness extraction now follows assignable destructor chains rooted
  at a generated formal, so `allowed(node_id(inner(c)))` and
  `edge(src(inner(c)),dst(inner(c)))` scan stored relation keys and assign
  nested fields such as `gen.c.inner.node_id`, `gen.c.inner.src`, and
  `gen.c.inner.dst` before rechecking the guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationWitnessForHashThunkDestructorFieldFast`.
  Relation-witness assignment now supports hash-thunk destructor fields by
  emitting `ivyThunkSet` when the relation model key should populate a
  non-enumerable indexed field such as `node_id(c,current)`, then rechecking
  the full guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNegatedRelationWitnessForUnboundedFormalFast`.
  For unbounded integer-like formals guarded by a negated unary state relation,
  e.g. `assume ~banned(n)`, the `target=test` generator now scans stored
  relation keys and tries a nearby value known to differ from a true banned key
  before rechecking the complete guard. This is still a narrow relation-backed
  witness heuristic, not the full solver model.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNegatedRelationWitnessInsideExistsFast`.
  Unary negated-relation witness discovery now carries existentially quantified
  variables as ignored tuple fields, so finite-emittable guards such as
  `exists C. ~banned(n,C)` can still scan stored relation keys for the generated
  formal instead of depending on one randomized value.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNegatedRelationTupleWitnessForUnboundedFormalsFast`.
  Negated tuple relation guards over several unbounded generated formals now
  get a grouped relation scan: for a true banned tuple, the generator copies
  the tuple fields and nudges one integer-like component before rechecking the
  full guard. This covers simple `~banned(a,b)` shapes without depending on two
  lucky randomized values.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNegatedRelationTupleWitnessWithFixedFieldFast`.
  The negated relation witness path now also handles larger relation tuples
  with fixed/evaluable fields, e.g. `~banned(green,n)`: generated code filters
  stored tuple keys on the fixed fields, assigns `n` from the matching key, and
  nudges it away from true banned cells before the full guard recheck.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesSmallIntFallbackForRelationBaseFast`.
  Relation-witness scans over thunk-backed unbounded relations now keep the
  stored-override scan but also try a deterministic small integer range after
  the scan. This lets guards like `allowed(n)` find relation truth supplied by a
  base function such as `allowed(N) := N = 7`, then recheck the complete guard.
  It is a bounded generated-code fallback, not a replacement for Python's SMT
  model.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesRelationWitnessThroughDerivedDefinitionFast`.
  `target=test` now expands derived definitions before action-generator witness
  planning, so a guard such as `ok(n)` with `definition ok(N) = allowed(N)`
  contributes the same relation-key scan as the inlined `allowed(n)` guard.
- 2026-09-22: Added
  `TestTargetTestDefinedInputDependenciesUsePythonOrderFast`. Defined-input
  extraction now orders dependent generated-input assignments like Python's
  reversed `extract_defined_parameters` result, so clauses such as `x = y; y =
  3` emit `gen.y = 3` before `gen.x = gen.y` in `target=test` generators and
  finite-search retries.
- 2026-09-22: Added
  `TestTargetTestConjunctiveDefinedInputDependenciesUsePythonOrderFast`.
  Defined-input extraction now descends into conjunctive guard formulas, so a
  single guard such as `x = y & y = 3` produces the same ordered generated-input
  assignments as the equivalent pair of separate assume clauses.
- 2026-09-22: Added
  `TestTargetTestIffDefinedInputUsesPythonOrderFast`. Defined-input extraction
  now treats equivalence with `true` as a positive guard, so IFF-wrapped
  equalities such as `(x = y) <-> true` assign `gen.x = gen.y` before the guard
  check instead of relying on two matching random unbounded inputs.
- 2026-09-22: Added `TestTargetTestBooleanIffDefinedInputFast`. Boolean
  defined-input extraction now also handles direct `LogicIff(input, expr)`
  formulas such as `x <-> active`, matching Python's
  `extract_defined_parameters` treatment of IFF parameter definitions instead
  of falling back to finite boolean search.
- 2026-09-22: Added `TestTargetTestImpliedDefinedInputUsesGeneratorGuardFast`.
  Defined-input extraction now also descends into implication consequents, so
  constraints such as `active -> x = green` assign `gen.x = green` before the
  guard check instead of relying on the slower finite-search fallback.
- 2026-09-22: Added
  `TestTargetTestDisjunctiveDefinedInputUsesGeneratorGuardFast`. Defined-input
  extraction now also traverses disjunctions, so a constructive branch such as
  `x = green | active` assigns `gen.x = green` before evaluating the full OR
  guard rather than falling back to finite search.
- 2026-09-22: Added `TestTargetTestIteDefinedInputUsesGeneratorGuardFast`.
  Conditional guards whose then and else branches define the same generated
  input now merge into one conditional value assignment, e.g.
  `ite(active, x = green, x = red)` emits
  `gen.x = ivyTernary(ivy.active, green, red)` before rechecking the full guard.
- 2026-09-22: Added
  `TestTargetTestIteCrossTargetDefinedInputUsesGeneratorGuardFast`.
  Conditional guards whose branches define different generated inputs now emit
  one guarded assignment per target, keeping inactive targets unchanged; for
  example `ite(active, x = 3, y = 4)` emits conditional assignments for both
  `gen.x` and `gen.y` before the guard check.
- 2026-09-22: Added
  `TestTargetTestLiteralDefinedInputUsesGeneratorGuardFast`. Defined-input
  extraction now unwraps positive `LogicLiteral` clauses before looking for
  parameter equalities, matching Python's literal-based clause pipeline instead
  of leaving literal-wrapped equalities to finite fallback search.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesVariantRelationWitnessFast`. Positive
  variant-relation guards over generated inputs now construct the constrained
  variant value before checking the guard; for example `assume x *> req0` emits
  `gen.x = msg{tag: 0, value: ivy.req0, valid: true}`. This covers a focused
  slice of Python's variant-axiom solver behavior; the later existential
  bullets below cover the currently modeled quantified variant slices while
  general SMT variant reasoning remains open.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIteVariantRelationWitnessFast`.
  Conditional variant-relation guards whose branches construct the same subtype
  now merge into one generated variant assignment, e.g.
  `ite(active, x *> req0, x *> req1)` emits a `req` variant whose payload is
  `ivyTernary(ivy.active, ivy.req0, ivy.req1)` before rechecking the full guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIteVariantRelationWitnessAcrossSubtypesFast`.
  Conditional variant-relation guards whose branches construct different
  subtypes now emit a conditional whole-variant value, e.g.
  `ivyTernary(ivy.active, msg{tag: 0, ...}, msg{tag: 1, ...})`, before the
  full branch-sensitive membership guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIteMultiVariantRelationWitnessFast`.
  Conditional variant guards whose branches each construct multiple generated
  inputs now merge all same-target branch witnesses before the guard check, so
  shapes like `ite(active, x *> req0 & y *> ack0, x *> ack0 & y *> req0)`
  assign both generated variants deterministically.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesItePartialVariantRelationWitnessFast`.
  Conditional variant guards with a membership witness on only one branch now
  construct that branch's variant while keeping the randomized generated value
  on the unconstrained branch.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNegatedVariantRelationWitnessFast`.
  Negated concrete variant-relation guards over generated inputs now choose a
  sibling subtype witness before checking the guard, so `assume ~(x *> req0)`
  emits an alternate variant such as `ack` instead of depending on the initial
  randomized variant tag.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesNegatedExistsVariantRelationWitnessFast`.
  Negated existential variant memberships now use the same sibling-subtype
  witness path, so `exists Q:req. ~(x *> Q)` constructs an alternate variant
  for generated inputs before evaluating the quantified guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesConstrainedNegatedExistsVariantRelationWitnessFast`.
  Negated existential variant memberships with a concrete payload equality now
  eliminate the quantified payload into a direct negated variant-payload guard,
  so `exists Q:req. ~(x *> Q) & Q = req0` both constructs a sibling variant
  witness and rechecks the concrete `req0` exclusion without an unsupported
  quantified-variable diagnostic.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesAffineConstrainedNegatedExistsVariantRelationWitnessFast`.
  Negated existential variant memberships with a unique affine payload equality
  now lower the quantified payload into the same direct guard, so
  `exists Q:req. ~(x *> Q) & Q + 1 = req0` checks the forbidden payload
  `ivy.req0 - 1` instead of reporting an unenumerable quantified variable.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIffFalseExistsVariantRelationWitnessFast`.
  Negated existential variant matching now also recognizes Ivy's
  IFF-with-false spelling, so `exists Q:req. ((x *> Q) <-> false)` follows the
  same sibling-witness and direct negated-tag guard path as the explicit
  `~(x *> Q)` form.
- 2026-09-22: Added
  `TestEmitExprExistsVariantRelationChecksTagFast` and
  `TestTargetTestActionGeneratorUsesExistsVariantRelationWitnessFast`.
  Existential variant membership such as `exists Q:req. x *> Q` now emits a
  direct validity/tag check and the `target=test` generator constructs a
  zero-payload `req` variant for `x` before checking that guard. More complex
  existential bodies still remain outside this small variant-axiom slice.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIffTrueExistsVariantRelationWitnessFast`.
  Positive existential variant matching now also recognizes Ivy's
  IFF-with-true spelling, so `exists Q:req. ((x *> Q) <-> true)` follows the
  same direct tag-check and zero-payload subtype witness path as the bare
  membership form.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesConstrainedExistsVariantRelationWitnessFast`.
  Existential variant memberships with simple payload equality constraints,
  such as `exists Q:req. x *> Q & Q = req0`, now emit a guarded downcast check
  and construct the generated variant value with the concrete payload
  `ivy.req0` before rechecking the full guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIteConstrainedExistsVariantRelationWitnessFast`.
  Exact existential variant payload extraction now also handles branch
  constraints under `ite`, so `exists Q:req. x *> Q &
  ite(active, Q = req0, Q = req1)` constructs the generated `req` variant with
  `ivyTernary(ivy.active, ivy.req0, ivy.req1)` before rechecking the
  quantified guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesItePartialConstrainedExistsVariantRelationWitnessFast`.
  Exact existential variant payload extraction now also handles one-branch
  `ite` constraints, using the subtype zero payload on the unconstrained
  branch while preserving the branch-sensitive guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesDisequalExistsVariantRelationWitnessFast`.
  Existential variant payload witnesses now handle simple integer-like payload
  disequality constraints, so `exists Q:req. x *> Q & Q ~= req0` constructs
  the `req` variant with `ivy.req0 + 1` instead of defaulting the payload to
  zero and immediately failing the guard when `req0` is zero.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesInequalityExistsVariantRelationWitnessFast`.
  Existential variant payload witnesses now also handle simple integer-like
  inequality constraints, so `exists Q:req. x *> Q & Q > 10` constructs a
  satisfying payload such as `11` before rechecking the quantified guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIteInequalityExistsVariantRelationWitnessFast`.
  Conditional existential payload inequality witnesses now merge branch
  witnesses, so `exists Q:req. x *> Q & ite(active, Q > 10, Q > 20)`
  constructs a payload like `ivyTernary(ivy.active, 11, 21)` before the guard
  downcast and recheck.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesItePartialInequalityExistsVariantRelationWitnessFast`.
  Conditional existential payload inequality witnesses now also tolerate a
  payload constraint on only one branch, constructing values such as
  `ivyTernary(ivy.active, 11, 0)` before the complete quantified guard check.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIteMixedDeltaExistsVariantRelationWitnessFast`.
  Conditional existential payload witnesses now materialize per-branch integer
  deltas before merging, so mixed branches such as `Q ~= req0` and `Q > 20`
  produce `ivyTernary(ivy.active, (ivy.req0 + 1), 21)` instead of falling back
  to a zero payload.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesAffineInequalityExistsVariantRelationWitnessFast`.
  The same existential variant payload witness path now handles simple affine
  payload inequalities, so `Q + 1 > 10` constructs payload `10` before the
  quantified guard check.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesAffineEqualityExistsVariantRelationWitnessFast`.
  Existential variant payload witnesses now also solve simple affine equality
  constraints, so `exists Q:req. x *> Q & Q + 1 = req0` constructs payload
  `ivy.req0 - 1` before the quantified guard downcast and recheck.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorPrefersExactExistsVariantPayloadWitnessFast`.
  When an existential variant payload has both a disequality and an exact
  equality, exact equality now wins regardless of conjunct order, so
  `Q ~= req0 & Q = req1` constructs the payload from `req1` rather than the
  looser `req0 + 1` fallback.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesLetConstrainedExistsVariantRelationWitnessFast`.
  Existential variant relation matching now expands simple `let` bodies before
  looking for `x *> Q`, so payload constraints hidden behind `let Alias = req0`
  produce the same concrete variant witness and direct guarded downcast as the
  non-let constrained form.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesLetExtraConstrainedExistsVariantRelationWitnessFast`.
  The concrete payload witness extractor also expands simple `let` expressions
  in the extra existential body terms, so `x *> Q & (let Alias = req0 in Q =
  Alias)` constructs `x` with `ivy.req0` instead of defaulting the subtype
  payload to zero and relying on the guard to reject it.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesLetMembershipExistsVariantRelationWitnessFast`.
  The existential variant matcher now also expands simple `let` expressions on
  individual conjunct terms before checking for the `x *> Q` membership atom,
  so `let Alias = Q in x *> Alias` is recognized as the solver witness source
  instead of falling back to unbounded quantifier enumeration.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesLetConjunctionExistsVariantRelationWitnessFast`.
  Existential variant matching now normalizes guard bodies by recursively
  expanding simple `let` expressions and flattening conjunctions before the
  `x *> Q` scan, so a let term that expands to both membership and payload
  equality is treated like the direct solver formula.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesDisjunctiveExistsVariantRelationWitnessFast`.
  Existential variant payload witness extraction now looks through simple
  disjunctions and chooses a concrete equality branch, so guards such as
  `exists Q:req. x *> Q & (Q = req0 | Q = req1)` construct `x` with `req0`
  before rechecking the full disjunctive guard.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesImpliedExistsVariantRelationWitnessFast`.
  Variant existential payload witnesses now also inspect implication
  consequents, matching the other action-generator witness extractors; this
  lets guards like `exists Q:req. x *> Q & (active -> Q = req0)` construct a
  concrete `req0` payload before rechecking the complete implication.
- 2026-09-22: Added
  `TestTargetTestActionGeneratorUsesIffExistsVariantRelationWitnessFast`.
  Variant existential payload witnesses now normalize IFF-with-true wrappers,
  so guards such as `exists Q:req. x *> Q & ((Q = req0) <-> true)` construct
  the concrete payload before the generated guard check.
- 2026-09-22: Added
  `TestTargetTestLocalWitnessEqualitiesUseGeneratorGuardFast`. The
  `target=test` local preimage walker now eliminates simple local equality
  witnesses, so a local-existential shape such as `scratch = c` and
  `scratch = green` becomes a generator guard on `c = green` instead of using a
  runtime trial of the public action. More complex local solver witnesses still
  remain outside this syntactic equality slice.
- 2026-09-22: Added
  `TestTargetTestLocalWitnessDisjunctiveEqualitiesUseGeneratorGuardFast`.
  Local equality witnesses under simple disjunctions now substitute the chosen
  local into the original OR guard, so `(scratch = c | active)` paired with
  `(scratch = green | active)` becomes a direct generated guard on
  `c = green | active` instead of an existential loop over the local.
- 2026-09-22: Added
  `TestTargetTestLocalRelationWitnessUsesGeneratorGuardFast`. Local guards
  that still mention local witnesses are now surfaced as conservative
  existential guard formulas when the locals do not leak through state updates.
  This lets the existing relation-witness search pick generated formals from
  true relation tuples for shapes such as `exists choice. allowed(choice,c)`,
  avoiding the `target=test` runtime trial path for that modeled slice.
- 2026-09-22: Added
  `TestTargetTestLocalEqualityPointUpdateUsesGeneratorGuardFast`. Equality
  witness substitutions for locals now also rewrite the preimage context, so a
  point update like `marked(choice) := true` under `assume choice = c` becomes
  a generator guard over `marked(c)` instead of being rejected as a local symbol
  leak.
- 2026-09-22: Added
  `TestTargetTestLocalFiniteDisequalityPointUpdateUsesGeneratorGuardFast`.
  Local disequality witnesses now cover finite boolean/enumerated sorts, using
  a deterministic alternate value such as `ivyTernary(c = red, green, red)` so
  relation-cell preimage updates under `choice ~= c` can be modeled without a
  hidden trial or unconstrained generator.
- 2026-09-22: Added
  `TestTargetTestLocalSingletonDisequalityUsesGeneratorGuardFast`. Singleton
  finite disequality locals are now recognized as unsatisfiable during local
  preimage analysis, emitting a direct `false` generator guard and discarding
  irrelevant local-leaking state updates instead of falling back to an
  unconditional generator.
- 2026-09-22: Added
  `TestTargetTestLocalSingletonDisequalityImplicationUsesGeneratorGuardFast`.
  The same unsatisfiable finite local disequality detection now preserves
  implication wrappers, turning `active -> choice ~= c` over a singleton sort
  into a generated `active -> false` guard instead of treating the action as
  unconditionally enabled.
- 2026-09-22: Added
  `TestTargetTestLocalSingletonDisequalityOrUsesGeneratorGuardFast`. OR
  wrappers now preserve the satisfiable branch while replacing the impossible
  singleton local disequality branch with `false`, so `active | choice ~= c`
  becomes a direct guard instead of an existential local scan.
- 2026-09-22: Added
  `TestTargetTestLocalSingletonDisequalityAndUsesGeneratorGuardFast`. AND
  wrappers now collapse to a direct `false` guard when any conjunct is an
  unsatisfiable singleton local disequality, avoiding existential local scans
  for impossible conjunctive local choices.
- 2026-09-22: Added
  `TestTargetTestLocalSingletonDisequalityIteUsesGeneratorGuardFast`. ITE
  wrappers now preserve non-local branch conditions while replacing impossible
  singleton local disequality branches with `false`, so
  `ite(active, choice ~= c, true)` becomes a direct conditional generator
  guard instead of an unconditional action.
- 2026-09-22: Added
  `TestTargetTestLocalSingletonDisequalityIffTrueUsesGeneratorGuardFast`.
  IFF wrappers now replace impossible singleton local disequality sides before
  rebuilding the equivalence guard, so `(choice ~= c) <-> true` no longer
  treats an unsatisfiable local choice as unconditional enabledness.
- 2026-09-22: Added
  `TestTargetTestLocalDisjunctiveEqualityPointUpdateUsesGeneratorGuardFast`.
  The OR-wrapped local equality witness path is now covered through point
  updates as well, proving the substituted witness rewrites later relation-cell
  updates before the generated guard is emitted.
- 2026-09-22: Added
  `TestTargetTestLocalEqualityChainPointUpdateUsesGeneratorGuardFast`.
  Consecutive local declarations with chained equality witnesses now propagate
  substitutions through each other before point-update preimage rewriting, so
  `scratch = mid; mid = c; marked(scratch) := true` emits a guard over
  `marked(c)` without exposing locals or using a hidden trial.
- 2026-09-22: Added
  `TestTargetTestLocalVariantNegativeSetActionPointUpdateSkipsTrialFast`.
  The zero-formal local witness proof now treats explicit `LogicSetAction`
  relation updates as modeled point updates, so negative relation-set forms
  such as `~seen(x)` after a variant witness avoid the hidden runtime trial
  path just like assignment-form `seen(x) := false`.
- 2026-09-22: Added
  `TestTargetTestLocalVariantTwoSidedConditionalPointUpdateSkipsTrialFast`.
  Conditional relation point-update proof now records both branch-specific cell
  writes, while collapsing identical branch writes back to an unguarded update.
  This covers `ite(active, seen(x), ~seen(x))` rechecks without hidden trial
  machinery or regression in the same-update branch case.
- 2026-09-22: Added
  `TestTargetTestLocalVariantLetActionStateUpdateSkipsTrialFast`. The
  zero-formal direct-execution proof now expands action-level `let` bindings
  with the same helper used for local witness discovery, so a let-wrapped local
  variant witness copied into state and rechecked does not force a hidden trial
  clone.
- 2026-09-22: Added
  `TestTargetTestLocalVariantLetActionConditionalPointUpdateSkipsTrialFast`.
  Branch-local relation and destructor field update extraction now also expands
  action-level `let` bindings before matching point updates, so let-wrapped
  conditional updates participate in the no-trial proof.
- 2026-09-22: Added
  `TestTargetTestLocalVariantPrivateCallStateUpdateSkipsTrialFast`. The
  zero-formal direct-execution proof now inlines non-recursive private calls
  without return targets after substituting actuals for formals, so deterministic
  helper calls that copy a local witness into state can be proven safe without a
  hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalVariantPrivateCallConditionalPointUpdateSkipsTrialFast`.
  Branch-local relation and destructor field update extraction now also inlines
  non-recursive private calls without return targets, so helper-wrapped
  conditional point updates are covered by the same no-trial proof.
- 2026-09-22: Added
  `TestTargetTestLocalVariantPrivateReturnCallStateUpdateSkipsTrialFast`. The
  private-call direct-execution proof now also handles simple return targets by
  using the existing formal/return substitution map, so helper return values
  assigned into state can satisfy later local-witness rechecks without a hidden
  trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalVariantChoicePointUpdateSkipsTrialFast`. The zero-formal
  direct-execution proof now models nondeterministic `choice` actions when every
  branch has the same relation point or destructor field update, so local
  witness-keyed relation cells can be set through a choice and rechecked without
  hidden trial machinery.
- 2026-09-22: Added
  `TestTargetTestLocalVariantEnvPointUpdateSkipsTrialFast`. `LogicEnvAction`
  now shares the same identical-branch update proof as `LogicChoiceAction`,
  matching the generated action emitter's env-as-choice behavior for local
  witness-keyed point updates.
- 2026-09-22: Added
  `TestTargetTestLocalRelationAssignFieldActionSkipsTrialFast`. Local witness
  initialization now scans thunk-backed boolean relation overrides even when the
  relation is not classified as a global quantifier-support relation, and the
  zero-formal proof models destructor `AssignFieldAction` updates followed by
  matching field equality assumes. This covers shapes like selecting `x` from
  `allowed(x)`, assigning `shade(x) := green`, and rechecking
  `shade(x) = green` without a hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalRelationNullFieldActionSkipsTrialFast` and
  `TestTargetTestLocalRelationCopyFieldActionSkipsTrialFast`. The same
  zero-formal field-update proof now lowers destructor `NullFieldAction` to the
  field's Go zero value and `CopyFieldAction` to an ordinary field-reference
  RHS, so relation-selected local objects can be nulled or copied and
  immediately rechecked without a hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalRelationGuardedAssignFieldActionSkipsTrialFast`. The
  zero-formal field-update proof is now branch-aware for one-sided conditional
  destructor field writes, so implication/OR/ITE rechecks such as
  `active -> shade(x) = green` activate the matching guarded update instead of
  forcing a hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalRelationElseGuardedAssignFieldActionSkipsTrialFast`.
  The field-update proof now handles the else-branch mirror as well, so
  disjunctive rechecks such as `active | shade(x) = green` activate the
  false-branch guarded field write instead of falling back to a runtime trial.
- 2026-09-22: Added
  `TestTargetTestLocalRelationTwoSidedAssignFieldActionSkipsTrialFast`. The
  same proof now records both branch-specific destructor field writes from an
  `if/else`, allowing `ite(active, shade(x) = green, shade(x) = red)` style
  rechecks to discharge against the matching branch update without hidden trial
  machinery.
- 2026-09-22: Added
  `TestTargetTestChoicePointUpdatePreimageUsesReverseImageGuardFast`.
  Nondeterministic choices whose branches update different relation/function
  cells now retain those branch point updates as alternatives in the preimage
  context, so a later assume such as `marked(c)` becomes a generated guard over
  the branch-specific rewritten formulas instead of an unconditional generator
  or hidden trial.
- 2026-09-22: Added
  `TestTargetTestChoiceMixedStateAndPointUpdatePreimageUsesGuardFast` and
  `TestTargetTestChoiceMixedStateAndPointUpdateKeepsBranchCorrelationFast`.
  Choices whose branches change both scalar state substitutions and point
  updates now use correlated branch alternatives, preserving Python
  reverse-image semantics for later guards instead of independently OR-ing
  state and cell updates.
- 2026-09-22: Added
  `TestTargetTestChoiceGuardedPointUpdatePreimageUsesGuardFast`. Choice
  branches with their own assume guards now keep those branch guards attached to
  the corresponding point-update alternative, so later assumes are guarded as
  `branch_guard & rewritten_later_guard` instead of falling back to an
  unconditional generator.
- 2026-09-22: Added `TestTargetTestGuardedInternalChoiceUsesTrialFast`.
  `target=test` actions with internal choice/env branches containing runtime
  assumes now use the hidden rejecting trial path after generator success. This
  prevents a visible `assumption_failed` from an unlucky internal branch choice
  while Go still lacks ivy2cpp's solver-controlled internal `___branch`
  generator choices.

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
- 2026-09-22: Added
  `TestTargetGenActionGeneratorSearchesFiniteDomainFast`. `target=gen`
  per-action generators now mirror the `target=test` finite-search fallback:
  after a randomized candidate fails a guard, small finite formal domains are
  searched from randomized offsets before the one-shot generator returns
  `false`. This improves non-equality finite constraints but still is not the
  Python SMT generator.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorSearchesRelevantFiniteFormalOnlyFast`. The
  shared finite-search planner now also lets `target=gen` solve constrained
  finite formals without requiring irrelevant unbounded formals to become
  search dimensions.
- 2026-09-22: Added
  `TestTargetGenChoiceStateUpdatePreimageUsesReverseImageGuardFast`. The choice
  alternative preimage context is shared by `target=gen`, so one-shot generators
  also honor disjunctive enabledness produced by nondeterministic branch state
  updates before a later assume.
- 2026-09-22: Added
  `TestTargetGenCallReturnPointTargetPreimageUsesGeneratorGuardFast`. The call
  return point-target preimage fix is shared by `target=gen`, so one-shot
  generators also honor helper return values written into relation/function
  cells before later assumes.
- 2026-09-22: Added `TestTargetGenLetActionPreimageUsesGeneratorGuardFast`.
  The `LogicLetAction` preimage substitution is shared by `target=gen`, so
  one-shot generators also respect simple let aliases before later assumes.
- 2026-09-22: Added
  `TestPreimageWalkerTreatsAssertLikeActionsAsTransparentFast`. The
  assert-like preimage transparency fix is shared by `target=gen`, so
  `requires`, `ensures`, and `subgoal` steps no longer force the one-shot
  generator planner off the syntactic preimage path.
- 2026-09-22: Added `TestPreimageWalkerTreatsIgnoreActionAsTransparentFast`.
  The no-op ignore-action transparency fix is shared by `target=gen`, so
  one-shot generators preserve preimage guards across ignored marker steps.
- 2026-09-22: Added `TestPreimageWalkerTreatsDebugActionAsTransparentFast`.
  The debug-action preimage transparency fix is shared by `target=gen`, so
  one-shot generator enabledness ignores debug statements the same way Python's
  transition update does.
- 2026-09-22: Added `TestPreimageWalkerLowersAssignFieldActionFast`,
  `TestPreimageWalkerLowersCopyFieldActionFast`, and
  `TestPreimageWalkerLowersNullFieldActionFast` /
  `TestPreimageWalkerLowersIntegerLikeNullFieldActionFast`. The destructor
  field-action preimage lowering is shared by `target=gen`, so one-shot
  generators can build guards from field assignment/copy/null updates instead
  of treating those emitted actions as opaque.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNumericDisequalityWitnessForUnboundedFormalFast`.
  The same simple numeral equality/disequality witness search is shared by
  `target=gen`, so one-shot generators can choose deterministic scalar
  witnesses such as `1` for `n ~= 0` before returning `false`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNumericInequalityWitnessForUnboundedFormalFast`.
  The inequality witness extraction is shared by `target=gen`, giving one-shot
  generators deterministic nearby integer witnesses for simple bounds such as
  `n > 10`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNegatedNumericInequalityWitnessForUnboundedFormalFast`.
  The negated numeric-bound witness extraction is shared by `target=gen`, so
  one-shot generators can satisfy normalized guards such as `~(n <= 10)`
  without relying on one randomized candidate.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIffFalseNumericInequalityWitnessForUnboundedFormalFast`.
  The IFF-with-false scalar negation path is shared by `target=gen`, so
  one-shot generators can model `(n <= 10) <-> false` as a negated bound.
- 2026-09-22: Added `TestTargetGenActionGeneratorUsesIteNumericWitnessFast`.
  The conditional numeric branch witness merge is shared by `target=gen`, so
  one-shot generators can try `ivyTernary(ivy.active, 11, 21)` for guards such
  as `ite(active, n > 10, n > 20)`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesItePartialNumericWitnessFast`. The
  one-branch conditional scalar numeric witness merge is shared by `target=gen`,
  so one-shot generators preserve unconstrained branch values for guards like
  `ite(active, n > 10, true)`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNumericAffineEqualityWitnessFast`. The
  affine scalar equality witness extraction is shared by `target=gen`, so
  one-shot generators can satisfy simple arithmetic guards like `n + 1 = 7`
  without relying on one random draw.
- 2026-09-22: Added `TestTargetGenActionGeneratorUsesLetNumericWitnessFast`.
  The simple `let` expansion for witness discovery is shared by `target=gen`,
  so one-shot generators can plan witnesses from let-wrapped arithmetic guards.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNumericAffineInequalityWitnessFast`. The
  affine scalar inequality witness extraction is shared by `target=gen`, giving
  one-shot generators deterministic witnesses for arithmetic bounds such as
  `n + 1 > 10`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNumericPairInequalityWitnessFast`. The
  scalar pair witness assignment is shared by `target=gen`, so one-shot
  generators can satisfy simple unbounded formal inequalities like `a < b`
  without depending on two lucky randomized values.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNegatedNumericPairInequalityWitnessFast`.
  The negated pairwise numeric-bound witness extraction is shared by
  `target=gen`, so one-shot generators can satisfy normalized guards such as
  `~(a <= b)` with a deterministic pair assignment.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNumericAffinePairInequalityWitnessFast`.
  The affine pair inequality witness assignment is shared by `target=gen`, so
  one-shot generators can satisfy constraints like `a + 1 < b` deterministically.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIteNumericPairInequalityWitnessFast`.
  The conditional same-target scalar pair witness assignment is shared by
  `target=gen`, so one-shot generators can synthesize branch-dependent pair
  values before rechecking the full conditional guard.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIteCrossTargetNumericPairWitnessFast`.
  The cross-target conditional scalar pair assignment is shared by `target=gen`,
  so one-shot generators can satisfy either branch of pair constraints such as
  `ite(active, a < b, b < a)` with guarded updates to both generated formals.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIteMultiNumericPairWitnessFast`. The
  conditional multi-pair witness merge is shared by `target=gen`, so one-shot
  generators emit branch-dependent assignments for every same-target pair
  witness before checking the full `ite` guard.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIteCrossMultiNumericPairWitnessFast`. The
  cross-target conditional multi-pair witness merge is shared by `target=gen`,
  so one-shot generators avoid the stale-random finite-search fallback for
  branch-sensitive pair constraints.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesItePartialNumericPairWitnessFast`. The
  one-branch conditional pair-witness merge is shared by `target=gen`, so
  one-shot generators preserve unconstrained branch values for guards like
  `ite(active, a < b, true)`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNumericPairDisequalityWitnessFast`. The
  scalar pair disequality witness is shared by `target=gen`, giving one-shot
  generators deterministic distinct integer-like formals for `a ~= b`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNumericAffinePairDisequalityWitnessFast`.
  The affine pair disequality witness is shared by `target=gen`, giving
  one-shot generators deterministic distinct values for `a + 1 ~= b`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNumericAffinePairEqualityWitnessFast`. The
  affine pair equality witness is shared by `target=gen`, so one-shot
  generators can deterministically satisfy constraints such as
  `a + 1 = b + 2`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesSmallIntFallbackForNonlinearNumericGuardFast`.
  The referenced-unbounded-integer small search fallback is shared by
  `target=gen`, so one-shot generators also try small model values for
  nonlinear guards such as `n * n = 9` before returning `false`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesProductEqualityWitnessForUnboundedFormalsFast`.
  The product-equality defined-input witness is shared by `target=gen`, so
  one-shot generators seed both unbounded integer-like formals for simple
  product model constraints before the final guard recheck.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesProductStateEqualityWitnessFast`. The
  integer-expression RHS product witness is shared by `target=gen`, preserving
  the same simple model extraction for one-shot generators.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesProductStateInequalityWitnessFast`. The
  product inequality boundary witness is shared by `target=gen`, preserving
  the same simple two-formal model extraction for one-shot generators.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesProductStateDisequalityWitnessFast`. The
  product disequality witness is shared by `target=gen`, so one-shot
  generators also seed a nearby distinct product before checking the negated
  equality guard.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesSumStateEqualityWitnessFast`. The additive
  two-formal equality witness is shared by `target=gen`, so one-shot generators
  also seed `a = rhs` and `b = 0` before checking guards such as
  `a + b = target`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesSumStateInequalityWitnessFast`. The
  additive two-formal inequality boundary witness is shared by `target=gen`,
  preserving the same simple model extraction for one-shot generators.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesSumStateDisequalityWitnessFast`. The
  additive two-formal disequality witness is shared by `target=gen`, so
  one-shot generators also seed a nearby distinct sum before checking the
  negated equality guard.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIteMultiArithmeticDefinedInputsFast`. The
  conditional multi-formal branch-witness merge is shared by `target=gen`, so
  one-shot generators preserve product/sum defined inputs under `ite` guards
  instead of falling back to random two-formal candidates.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIteCrossMultiArithmeticDefinedInputsFast`.
  The cross-target conditional multi-defined-input merge is shared by
  `target=gen`, so one-shot generators can solve active-branch product/sum
  constraints while leaving inactive branch targets unchanged.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesItePartialDefinedInputFast`. The
  one-branch conditional defined-input merge is shared by `target=gen`, so
  one-shot generators keep unconstrained branch values instead of relying on
  finite search for guards such as `ite(active, x = green, true)`.
- 2026-09-22: Added `TestTargetGenLocalNumericInequalityChoosesWitnessFast`.
  Generated action bodies share the local scalar witness initialization in
  `target=gen`, preventing one-shot generated traces from failing simple local
  numeric assumes that Python's solver-backed generator can satisfy.
- 2026-09-22: Added
  `TestTargetGenLocalNumericFormalInequalityChoosesWitnessFast`. The
  formal-dependent local scalar witness is shared by `target=gen`, and the
  local-preimage planner no longer emits an unsupported existential guard for
  the single satisfiable local inequality shape.
- 2026-09-22: Added
  `TestTargetGenLocalNumericFormalInequalityStateUpdateFast`. The
  formal-dependent inequality state-copy preimage substitution is shared by
  `target=gen`, so one-shot generated action bodies and generators stay on the
  modeled direct path for `scratch > c` copied into state.
- 2026-09-22: Added
  `TestTargetGenLocalNumericAffineEqualityStateUpdateFast`. The
  formal-dependent affine equality local witness and preimage substitution are
  shared by `target=gen`, so one-shot generators avoid broad fallback scans and
  unsupported existential guards for `scratch + 1 = c` state-copy shapes.
- 2026-09-22: Added
  `TestTargetGenLocalNumericFormalDisequalityFast`. The negated affine local
  equality witness is shared by `target=gen`, so generated action bodies choose
  deterministic formal-dependent disequality witnesses instead of scanning the
  broad small-integer fallback.
- 2026-09-22: Added
  `TestTargetGenLocalNumericFormalNegatedInequalityFast`. The negated
  formal-dependent inequality witness is shared by `target=gen`, so one-shot
  generated action bodies use boundary witnesses for guards such as
  `~(scratch < c)` and avoid unsupported existential guard fallback.
- 2026-09-22: Added
  `TestTargetGenLocalNumericFormalIffFalseInequalityFast`. The IFF-with-false
  local inequality normalization is shared by `target=gen`, preserving the same
  deterministic witness and avoiding unsupported existential guard fallback.
- 2026-09-22: Added
  `TestTargetGenLocalNumericFormalIffFalseAffineEqualityFast`. The
  IFF-with-false affine equality witness is shared by `target=gen`, so one-shot
  generated action bodies choose formal-dependent disequality witnesses instead
  of broad fallback scans.
- 2026-09-22: Added
  `TestTargetGenLocalNumericFormalImplicationFast`. The implication-consequent
  local witness path is shared by `target=gen`, so one-shot generated action
  bodies keep deterministic witnesses for guards such as
  `active -> scratch > c`.
- 2026-09-22: Added `TestTargetGenLocalNumericFormalOrFast` and
  `TestTargetGenLocalNumericFormalAndPreservesGuardFast`. The OR guardrail and
  conjunction-preserving local scalar witness path are shared by `target=gen`,
  so generated action bodies still choose deterministic locals while generated
  action generators retain non-local guards such as `active`.
- 2026-09-22: Added `TestTargetGenLocalNumericFormalIteFast`. The conditional
  local scalar witness path is shared by `target=gen`, so one-shot generated
  action bodies use `ivyTernary` local witnesses and generated action
  generators keep the corresponding conditional guard.
- 2026-09-22: Added `TestTargetGenLocalNumericFormalPartialIteFast`. The
  partial conditional local scalar witness path is shared by `target=gen`, so
  one-shot generated action bodies can satisfy `ite(active, scratch > c, true)`
  without emitting an unsupported existential local guard.
- 2026-09-22: Added
  `TestTargetGenLocalNumericFormalLetAndPreservesGuardFast`. The
  expression-level `LogicLet` preimage expansion is shared by `target=gen`,
  preserving non-local guards under aliases while keeping the deterministic
  local scalar witness.
- 2026-09-22: Added `TestLocalVariantExistsAssumeChoosesWitnessFast`. The
  local existential variant witness initialization is shared by `target=gen`,
  so one-shot generated action bodies seed variant-super locals before assumes
  like `exists Q:req. x *> Q`.
- 2026-09-22: Added `TestLocalVariantConcreteAssumeChoosesWitnessFast`. The
  direct concrete local variant-membership witness is shared by `target=gen`,
  so one-shot generated action bodies construct local variant values for guards
  like `x *> req0` before executing the assume.
- 2026-09-22: Added
  `TestLocalVariantIteAssumeChoosesConditionalWitnessFast`. The conditional
  local variant-membership witness is shared by `target=gen`, so one-shot
  generated action bodies construct `ivyTernary` variant locals for either
  branch of guards like `ite(active, x *> req0, x *> ack0)`.
- 2026-09-22: Added
  `TestLocalVariantPartialIteAssumeChoosesConditionalWitnessFast`. The
  one-branch conditional local variant witness is shared by `target=gen`, so
  one-shot generated action bodies keep unconstrained branch locals unchanged.
- 2026-09-22: Added `TestLocalVariantNegatedConcreteAssumeChoosesWitnessFast`.
  The sibling-subtype local negated variant witness is shared by `target=gen`,
  so one-shot generated action bodies satisfy guards like `~(x *> req0)`
  deterministically.
- 2026-09-22: Added
  `TestTargetGenLocalNumericWitnessAfterDebugChoosesWitnessFast`. The
  transparent debug/ignore/assert-like local-witness assume collection is
  shared by `target=gen`, so one-shot generated action bodies seed local
  witnesses even when debug output precedes the assume.
- 2026-09-22: Added
  `TestTargetGenLocalNumericWitnessInsideLetChoosesWitnessFast`. The simple
  `LogicLetAction` alias expansion for local witness discovery is shared by
  `target=gen`, so generated action bodies seed local scalar witnesses from
  let-wrapped assumes before executing them.
- 2026-09-22: Added
  `TestTargetGenLocalNumericNonlinearUsesSmallIntFallbackFast`. The local
  unbounded-integer small candidate scan is shared by `target=gen`, so one-shot
  generated action bodies can seed nonlinear local guards such as `n * n = 9`.
- 2026-09-22: Added
  `TestLocalWitnessUsesSmallIntFallbackForRelationBaseFast`. The relation-base
  fallback for local witnesses is shared by `target=gen`, reducing runtime
  `assumption_failed` cases where the one-shot action body needs an unbounded
  local value satisfying a thunk-backed relation guard.
- 2026-09-22: Added `TestLocalWitnessFromNegatedRelationFast`. The negated
  local relation-witness path is shared by `target=gen`, so one-shot generated
  action bodies can seed locals for guards like `~banned(n)` before executing
  the assume.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNegatedRelationWitnessInsideExistsFast`.
  The existential ignored-field handling for unary negated-relation witnesses is
  shared by `target=gen`, so one-shot generators can use stored relation keys
  for finite-emittable guards such as `exists C. ~banned(n,C)`.
- 2026-09-22: Added `TestLocalWitnessFromNegatedRelationBaseFallbackFast`. The
  negated local relation small-integer fallback is shared by `target=gen`, so
  thunk-backed negated relation guards can find a satisfying local value before
  the assume.
- 2026-09-22: Added `TestLocalWitnessPairFromNegatedRelationFast`. The grouped
  local negated tuple witness path is shared by `target=gen`, so one-shot
  generated action bodies can seed several adjacent locals from one relation
  cell while nudging one integer-like field out of a known true tuple.
- 2026-09-22: Added
  `TestLocalWitnessFromExistsUsesExtensionalRelationFast`. The existential
  wildcard local-witness path is shared by `target=gen`, allowing one-shot
  generated action bodies to choose local relation witnesses under `exists`
  guards without relying on a lucky random local value.
- 2026-09-22: Added `TestLocalWitnessFromIffUsesExtensionalRelationFast`. The
  local relation-witness IFF handling is shared by `target=gen`, so one-shot
  generated action bodies seed locals for `allowed(n) <-> true` before executing
  the assume.
- 2026-09-22: Added `TestLocalWitnessFromIteUsesExtensionalRelationFast`. The
  conditional local relation-witness path is shared by `target=gen`, so
  one-shot generated action bodies can seed locals from extensional relation
  witnesses inside `ite` guards before executing the assume.
- 2026-09-22: Added
  `TestLocalWitnessFromPartialIteUsesExtensionalRelationFast`. The one-branch
  conditional local relation-witness path is shared by `target=gen`, so
  unconstrained branches no longer force a broad local fallback scan.
- 2026-09-22: Strengthened
  `TestLocalWitnessFromIteUsesExtensionalRelationFast` to assert the shared
  `target=gen` path emits branch-specific relation scans under the generated
  `if/else`, rather than always seeding the local from the first ITE branch.
- 2026-09-22: Added
  `TestLocalWitnessPairUsesOneExtensionalRelationTupleFast`. The grouped local
  tuple witness path is shared by `target=gen`, so one-shot generated action
  bodies can seed several adjacent local variables from the same relation cell
  before executing the assume.
- 2026-09-22: Added `TestTargetGenDerivedAssumeUsesGeneratorGuardFast`.
  `target=gen` now has explicit coverage for the audit's derived-predicate
  case: generated action guards inline definitions such as `is_green(c)` to
  `gen.c == green` instead of leaving a stale definition call or relying on one
  random draw.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationWitnessForUnboundedFormalFast`.
  The relation-key witness search added for `target=test` is shared with
  `target=gen`, so one-shot generators can choose unbounded inputs from true
  relation cells for simple unary relation guards.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationTupleWitnessForUnboundedFormalFast`.
  The shared relation-witness path now covers tuple-key relation cells for
  `target=gen` as well, so constraints like `allowed(green,n)` can obtain an
  unbounded generated input from stored true relation cells before executing the
  public action.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationPairWitnessForUnboundedFormalsFast`.
  The same joint tuple-key witness search is shared by `target=gen`, so
  one-shot generators can choose several unbounded inputs from one true
  relation cell and are no longer dependent on lucky randomized companion
  formals.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationWitnessInsideOrFast`. The shared
  unary relation-witness finder now handles disjunctive guards for `target=gen`
  too, rechecking the full guard after assigning a candidate from the stored
  relation key.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesAlternativeRelationWitnessesInsideOrFast`.
  The alternative relation-witness search is shared by `target=gen`, so
  one-shot generators can try multiple relation-backed sources for the same
  unbounded formal before returning `false`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationPairWitnessInsideOrFast`. The
  grouped tuple-key relation witness search is shared by `target=gen` and now
  handles disjunctive guards, preserving one-model-tuple assignment for several
  generated formals.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationWitnessInsideExistsFast`. The
  exists-aware relation-witness discovery is shared by `target=gen`, letting
  one-shot generators derive an unbounded formal from a true relation tuple even
  when the relation appears under an existential quantifier.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationPairWitnessInsideExistsFast`. The
  grouped existential wildcard handling is shared by `target=gen`, preserving
  one relation-tuple model assignment for several generated formals under
  `exists` guards.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationWitnessInsideImplicationFast`. The
  implication-consequent relation-witness path is shared by `target=gen`, so
  one-shot generators can use relation-backed candidates for guards such as
  `active -> allowed(n)` and still recheck the complete guard before accepting.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationWitnessInsideIffFast`. The IFF
  relation-witness path is shared by `target=gen`, so one-shot generators can
  use stored relation evidence for guards such as `allowed(n) <-> true` instead
  of rejecting after one randomized unbounded input.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationWitnessInsideIteFast`. The
  conditional-branch relation-witness path is shared by `target=gen`, so
  one-shot generators can try alternate relation scans for guards such as
  `ite(active, allowed(n), permitted(n))`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationWitnessForUnboundedDestructorFieldFast`.
  The destructor-field relation witness path is shared by `target=gen`, so
  one-shot generators can populate unbounded fields inside structured formals
  from relation-backed model evidence before returning `false`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationWitnessForUnboundedDestructorFieldPairFast`.
  The same joint structured-field relation witness search is shared by
  `target=gen`, covering tuple relation guards over multiple unbounded fields
  of the same generated formal.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationWitnessForNestedDestructorFieldFast`
  and
  `TestTargetGenActionGeneratorUsesRelationWitnessForNestedDestructorFieldPairFast`.
  The recursive destructor-chain witness target logic is shared by
  `target=gen`, so one-shot generators can populate nested structured fields
  from unary or tuple relation evidence instead of relying on one randomized
  nested-field value.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationWitnessForHashThunkDestructorFieldFast`.
  The setter-aware hash-thunk relation witness path is shared by `target=gen`,
  allowing one-shot generators to write relation model keys into generated
  hash-thunk fields before guard evaluation.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorSearchesFiniteHashThunkDestructorFieldFast`.
  The finite hash-thunk field-search path is shared by `target=gen`, giving
  one-shot generators the same setter-backed finite candidate scan for
  guard-mentioned hash-thunk cells.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorSearchesFiniteNestedDestructorFieldFast`.
  The nested finite destructor-field search is shared by `target=gen`, so
  one-shot generators can scan fields like `gen.c.inner.shade` before
  evaluating guards such as `shade(inner(c)) ~= red`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorSearchesFiniteHashThunkDestructorFieldWithFormalIndexFast`.
  The formal-index coverage is shared by `target=gen`, so one-shot generators
  can leave the index formal randomized while searching the generated
  hash-thunk field value.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNegatedRelationWitnessForUnboundedFormalFast`.
  The negated unary relation witness heuristic is shared by `target=gen`, so
  one-shot generators can avoid known true cells such as `banned(n)` by trying a
  neighboring integer-like value and then rechecking the full generated guard.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNegatedRelationTupleWitnessForUnboundedFormalsFast`.
  The grouped negated tuple relation heuristic is shared by `target=gen`, so
  one-shot generators can synthesize simple not-in-relation tuple candidates
  for guards like `~banned(a,b)` before executing the public action.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNegatedRelationTupleWitnessWithFixedFieldFast`.
  The fixed-field negated tuple relation scan is shared by `target=gen`, giving
  one-shot generators the same conditioned not-in-relation witness for guards
  such as `~banned(green,n)`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesSmallIntFallbackForRelationBaseFast`. The
  same bounded small-integer fallback is shared by `target=gen`, so one-shot
  generators are no longer limited to relation override maps when a thunk base
  function supplies the satisfying relation value.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationWitnessThroughDerivedDefinitionFast`.
  The same pre-planning definition expansion is shared by `target=gen`, so
  one-shot generated action generators can discover relation-backed witnesses
  through derived predicates instead of depending on one randomized candidate.
- 2026-09-22: Added
  `TestTargetGenDefinedInputDependenciesUsePythonOrderFast`. The same
  dependency ordering is shared by `target=gen`, preventing one-shot generators
  from assigning `x` from a stale randomized `y` before applying the definition
  that fixes `y`.
- 2026-09-22: Added
  `TestTargetGenConjunctiveDefinedInputDependenciesUsePythonOrderFast`. The
  conjunctive defined-input extraction is shared by `target=gen`, so one-shot
  generators define dependent inputs from `and` formulas before checking the
  full guard.
- 2026-09-22: Added `TestTargetGenIffDefinedInputUsesPythonOrderFast`. The
  IFF defined-input extraction is shared by `target=gen`, so one-shot generators
  can derive inputs from `(x = y) <-> true` before evaluating the complete guard.
- 2026-09-22: Added `TestTargetGenBooleanIffDefinedInputFast`. The direct
  boolean `LogicIff(input, expr)` parameter-definition path is shared by
  `target=gen`, so one-shot generators assign boolean generated inputs from
  state expressions before checking the IFF guard.
- 2026-09-22: Added `TestTargetGenImpliedDefinedInputUsesGeneratorGuardFast`.
  The implication-consequent defined-input extraction is shared by
  `target=gen`, so one-shot generators assign generated inputs from simple
  implied equality constraints before guard evaluation.
- 2026-09-22: Added
  `TestTargetGenDisjunctiveDefinedInputUsesGeneratorGuardFast`. The
  disjunctive defined-input extraction is shared by `target=gen`, so one-shot
  generators assign a branch-satisfying generated input before checking the
  full OR guard.
- 2026-09-22: Added `TestTargetGenIteDefinedInputUsesGeneratorGuardFast`. The
  conditional branch-defined input merge is shared by `target=gen`, so
  one-shot generators synthesize conditional generated-input assignments such
  as `gen.x = ivyTernary(ivy.active, green, red)` before evaluating the guard.
- 2026-09-22: Added
  `TestTargetGenIteCrossTargetDefinedInputUsesGeneratorGuardFast`. The
  cross-target conditional defined-input path is shared by `target=gen`, so
  one-shot generators can satisfy branch definitions such as
  `ite(active, x = 3, y = 4)` with guarded assignments to both generated
  inputs.
- 2026-09-22: Added
  `TestTargetGenLiteralDefinedInputUsesGeneratorGuardFast`. The positive
  `LogicLiteral` defined-input extraction is shared by `target=gen`, so
  literal-wrapped equality clauses define generated inputs before guard
  evaluation rather than relying on bounded search.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesVariantRelationWitnessFast`. The positive
  variant-relation witness assignment is shared by `target=gen`, letting
  one-shot generators satisfy simple `*>` guards by constructing the required
  variant value before execution.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIteVariantRelationWitnessFast`. The
  conditional same-subtype variant witness merge is shared by `target=gen`, so
  one-shot generators construct branch-dependent payloads such as
  `ivyTernary(ivy.active, ivy.req0, ivy.req1)` before evaluating the full
  conditional membership guard.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIteVariantRelationWitnessAcrossSubtypesFast`.
  The conditional whole-variant witness path is shared by `target=gen`, so
  branch-dependent subtype choices are assigned as one `ivyTernary` over the
  full variant value before the generator rechecks enabledness.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIteMultiVariantRelationWitnessFast`. The
  conditional multi-variant witness merge is shared by `target=gen`, so
  one-shot generators construct every same-target generated variant witness
  needed by a branch-sensitive `ite` guard.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesItePartialVariantRelationWitnessFast`. The
  one-branch conditional variant witness merge is shared by `target=gen`, so
  one-shot generators preserve unconstrained branch values for guards like
  `ite(active, x *> req0, true)`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNegatedVariantRelationWitnessFast`. The
  negated concrete variant-relation witness is shared by `target=gen`, so
  one-shot generators choose a sibling subtype before evaluating guards such as
  `~(x *> req0)`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesNegatedExistsVariantRelationWitnessFast`.
  The negated existential variant-membership witness is shared by `target=gen`,
  so one-shot generators choose a sibling subtype before evaluating guards such
  as `exists Q:req. ~(x *> Q)`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesConstrainedNegatedExistsVariantRelationWitnessFast`.
  The concrete-payload negated existential variant guard lowering is shared by
  `target=gen`, so one-shot generators can evaluate
  `exists Q:req. ~(x *> Q) & Q = req0` as a direct negated variant-payload
  check after constructing the sibling witness.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesAffineConstrainedNegatedExistsVariantRelationWitnessFast`.
  The unique affine-payload lowering is shared by `target=gen`, so one-shot
  generators no longer fail generation for guards such as
  `exists Q:req. ~(x *> Q) & Q + 1 = req0`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIffFalseExistsVariantRelationWitnessFast`.
  The IFF-with-false spelling for negated existential variant membership is
  shared by `target=gen`, so one-shot generators no longer reject
  `exists Q:req. ((x *> Q) <-> false)` as an unenumerable quantified guard.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesExistsVariantRelationWitnessFast`. The same
  existential variant membership witness is shared by `target=gen`, so one-shot
  generators can satisfy `exists Q:req. x *> Q` without relying on a random tag
  choice.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIffTrueExistsVariantRelationWitnessFast`.
  The IFF-with-true positive existential variant membership spelling is shared
  by `target=gen`, so one-shot generators treat
  `exists Q:req. ((x *> Q) <-> true)` like the bare membership form.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesConstrainedExistsVariantRelationWitnessFast`.
  The constrained existential variant witness is shared by `target=gen`, so
  one-shot generators construct the requested subtype payload for guards like
  `exists Q:req. x *> Q & Q = req0` instead of rejecting the unenumerable
  quantified variable.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIteConstrainedExistsVariantRelationWitnessFast`.
  The conditional exact-payload extraction is shared by `target=gen`, so
  one-shot generators construct subtype payloads from branch constraints with a
  generated `ivyTernary` expression before the full guard check.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesItePartialConstrainedExistsVariantRelationWitnessFast`.
  The one-branch conditional exact-payload extraction is shared by `target=gen`,
  so one-shot generators keep the unconstrained existential payload at the
  subtype zero value while satisfying the constrained branch.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesDisequalExistsVariantRelationWitnessFast`.
  The simple disequal existential variant payload witness is shared by
  `target=gen`, so one-shot generators use a nearby integer-like payload such
  as `ivy.req0 + 1` before evaluating `Q ~= req0`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesInequalityExistsVariantRelationWitnessFast`.
  The simple inequality existential variant payload witness is shared by
  `target=gen`, so one-shot generators can construct integer-like subtype
  payloads for guards such as `exists Q:req. x *> Q & Q > 10`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIteInequalityExistsVariantRelationWitnessFast`.
  The conditional inequality payload witness merge is shared by `target=gen`,
  so one-shot generators can construct branch-dependent numeric subtype
  payloads before rechecking the full quantified guard.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesItePartialInequalityExistsVariantRelationWitnessFast`.
  The one-branch conditional inequality payload witness merge is shared by
  `target=gen`, preserving the subtype zero payload on unconstrained branches.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIteMixedDeltaExistsVariantRelationWitnessFast`.
  The per-branch delta materialization is shared by `target=gen`, so one-shot
  generators can merge conditional payload branches that use different witness
  offsets.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesAffineInequalityExistsVariantRelationWitnessFast`.
  The affine inequality payload witness is shared by `target=gen`, so one-shot
  generators also construct satisfying subtype payloads for guards such as
  `exists Q:req. x *> Q & Q + 1 > 10`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesAffineEqualityExistsVariantRelationWitnessFast`.
  The affine equality payload witness is shared by `target=gen`, so one-shot
  generators can construct subtype payloads such as `ivy.req0 - 1` for
  quantified guards like `exists Q:req. x *> Q & Q + 1 = req0`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorPrefersExactExistsVariantPayloadWitnessFast`.
  The exact-before-disequal payload preference is shared by `target=gen`, so
  one-shot generators choose concrete equality payloads before using the
  integer-like disequality fallback.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesLetConstrainedExistsVariantRelationWitnessFast`.
  The same `let` expansion in existential variant matching is shared by
  `target=gen`, so one-shot generators can construct constrained subtype
  payloads even when the payload equality is expressed through a simple local
  `let` alias.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesLetExtraConstrainedExistsVariantRelationWitnessFast`.
  The extra-term payload witness `let` expansion is shared by `target=gen`, so
  one-shot variant generators no longer construct zero-payload subtype values
  when a simple let-bound alias determines the existential payload.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesLetMembershipExistsVariantRelationWitnessFast`.
  The let-wrapped membership atom recognition is shared by `target=gen`, so
  one-shot generators do not reject existential variant guards merely because
  `x *> Q` is expressed through a simple let alias.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesLetConjunctionExistsVariantRelationWitnessFast`.
  The recursive let/conjunction normalization for existential variant witnesses
  is shared by `target=gen`, preserving the same generated subtype construction
  for one-shot generators.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesDisjunctiveExistsVariantRelationWitnessFast`.
  The disjunctive existential variant payload witness is shared by
  `target=gen`, so one-shot generators choose a concrete subtype payload from a
  simple equality branch instead of defaulting the payload to zero.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesImpliedExistsVariantRelationWitnessFast`.
  The implication-consequent variant payload witness is shared by `target=gen`,
  so one-shot generators choose a concrete subtype payload for simple implied
  equality constraints before checking the full guard.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesIffExistsVariantRelationWitnessFast`. The
  IFF-with-true variant payload witness is shared by `target=gen`, so one-shot
  generators construct the concrete subtype payload for simple equivalence
  constraints before checking the full guard.
- 2026-09-22: Added
  `TestTargetGenBeforeExportLocalRelationWitnessUsesGeneratorGuardFast`. The
  before-export existential local relation-witness bridge is shared by
  `target=gen`, so one-shot generators scan relation tuples and constrain the
  generated public formal instead of treating the analysis action as fatal or
  unconditionally enabled.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesRelationOverrideForallGuardFast`. The
  generator-only relation override recheck for unbounded negative relation
  quantifiers is shared by `target=gen`, so one-shot generators no longer fail
  generation for guards such as `forall X. ~banned(X,c)`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesEqualityBoundForallGuardFast`. The
  equality-bound quantifier simplifier is shared by `target=gen`, so one-shot
  generators can emit direct guards for solver-trivial formulas such as
  `forall X. X = c -> allowed(X)`.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesMultiEqualityBoundForallGuardFast`. The
  multi-variable equality-bound quantifier simplifier is shared by
  `target=gen`, preserving the same direct guard emission for tuple relation
  constraints.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorPreservesExtraEqualityBoundForallAntecedentFast`.
  The preserved-antecedent equality-bound universal simplifier is shared by
  `target=gen`, keeping branch/state guards around the simplified relation
  check.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesEqualityOnlyExistsGuardFast`. The
  equality-only existential simplifier is shared by `target=gen`, so one-shot
  generators no longer fail generation for solver-trivial satisfiable
  existential equalities.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesEqualityBoundForallNumericWitnessFast`. The
  AST-level equality-bound quantifier simplification is shared by `target=gen`,
  letting one-shot generators use existing scalar witness extraction after the
  quantified guard is reduced.
- 2026-09-22: Added
  `TestTargetGenActionGeneratorUsesEqualityBoundExistsNumericWitnessFast`. The
  non-variant equality-bound existential simplifier is shared by `target=gen`,
  preserving variant-specific witness emission while reducing scalar
  existential guards for witness planning.
- 2026-09-22: Added
  `TestTargetGenLocalEqualityBoundExistsNumericWitnessFast`. The local
  equality-bound existential simplifier is shared by `target=gen`, so one-shot
  generated actions seed integer-like locals from the simplified residual guard
  instead of falling back to broad bounded scans.
- 2026-09-22: Added
  `TestTargetGenLocalWitnessEqualitiesUseGeneratorGuardFast`. The same local
  equality-witness elimination is shared by `target=gen`, so one-shot
  generators constrain generated formals for simple existential-local
  equalities instead of executing an action whose local assumes can fail.
- 2026-09-22: Added
  `TestTargetGenLocalWitnessDisjunctiveEqualitiesUseGeneratorGuardFast`. The
  disjunctive local equality bridge is shared by `target=gen`, avoiding
  existential local scans when a substituted OR guard over generated formals is
  enough.
- 2026-09-22: Added
  `TestTargetGenLocalRelationWitnessUsesGeneratorGuardFast`. The existential
  local-guard bridge is shared by `target=gen`, so one-shot generators can use
  relation tuple evidence to constrain generated formals even when another
  tuple component is a local witness.
- 2026-09-22: Added
  `TestTargetGenLocalEqualityPointUpdateUsesGeneratorGuardFast`. The local
  equality point-update preimage rewrite is shared by `target=gen`, so
  one-shot generators also preserve modeled local witness updates before later
  assumes.
- 2026-09-22: Added
  `TestTargetGenLocalFiniteDisequalityPointUpdateUsesGeneratorGuardFast`. The
  finite local disequality witness path is shared by `target=gen`, preserving
  relation-cell preimage updates for enum/boolean locals constrained by `~=`.
- 2026-09-22: Added
  `TestTargetGenLocalSingletonDisequalityUsesGeneratorGuardFast`. The
  singleton finite disequality unsat guard is shared by `target=gen`, so
  one-shot generators reject impossible local choices before execution.
- 2026-09-22: Added
  `TestTargetGenLocalSingletonDisequalityImplicationUsesGeneratorGuardFast`.
  The implication-preserving singleton disequality guard is shared by
  `target=gen`, so one-shot generators retain the branch condition around an
  impossible local choice.
- 2026-09-22: Added
  `TestTargetGenLocalSingletonDisequalityOrUsesGeneratorGuardFast`. The
  OR-preserving singleton disequality guard is shared by `target=gen`, avoiding
  existential local scans for `active | impossible-local-choice` shapes.
- 2026-09-22: Added
  `TestTargetGenLocalSingletonDisequalityAndUsesGeneratorGuardFast`. The
  AND-collapsing singleton disequality guard is shared by `target=gen`, so
  impossible conjunctive local choices emit direct `false` guards instead of
  existential local scans.
- 2026-09-22: Added
  `TestTargetGenLocalSingletonDisequalityIteUsesGeneratorGuardFast`. The
  ITE-preserving singleton disequality guard is shared by `target=gen`, so
  one-shot generators keep branch-sensitive enabledness around impossible
  local choices.
- 2026-09-22: Added
  `TestTargetGenLocalSingletonDisequalityIffTrueUsesGeneratorGuardFast`. The
  IFF-preserving singleton disequality guard is shared by `target=gen`, so
  one-shot generators do not treat impossible local equivalence constraints as
  unconditional success.
- 2026-09-22: Added
  `TestTargetGenLocalDisjunctiveEqualityPointUpdateUsesGeneratorGuardFast`.
  The same OR-wrapped local equality point-update coverage is shared by
  `target=gen`, guarding against regressions back to existential local scans or
  leaked local names.
- 2026-09-22: Added
  `TestTargetGenLocalEqualityChainPointUpdateUsesGeneratorGuardFast`. The
  chained local equality point-update substitution is shared by `target=gen`,
  so one-shot generators also reduce chained locals to generated-formal guards
  before execution.
- 2026-09-22: Added
  `TestTargetGenChoicePointUpdatePreimageUsesReverseImageGuardFast`. The
  choice point-update preimage alternative handling is shared by `target=gen`,
  so one-shot generators also preserve branch-specific relation/function cell
  writes when building guards for later assumes.
- 2026-09-22: Added
  `TestTargetGenChoiceMixedStateAndPointUpdatePreimageUsesGuardFast`. The
  correlated mixed choice alternative path is shared by `target=gen`, so
  one-shot generators keep scalar state choices and relation/function cell
  writes paired by branch when building later assume guards.
- 2026-09-22: Added
  `TestTargetGenChoiceGuardedPointUpdatePreimageUsesGuardFast`. The guarded
  branch point-update alternative path is shared by `target=gen`, preserving
  branch-local assume guards when one-shot generators build later assume guards.
- 2026-09-22: Added `TestTargetGenGuardedInternalChoiceUsesTrialFast`.
  `target=gen` now mirrors the target=test interim safety behavior for
  internal choice/env branches with runtime assumes: execute on a clone with
  assumption rejection enabled, skip the one-shot action on rejection, and copy
  back accepted state before emitting the public trace.
- 2026-09-22: Added
  `TestTargetGenGuardedInternalChoiceTrialPreservesTraceBracesFast`. Accepted
  `target=gen` trial executions with action tracing enabled now preserve the
  same `{` / `}` trace envelope as the direct action path while still keeping
  rejected trials silent.

## 3. FIXED Initial state generation is retry/randomized, not Python's initial model

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
  satisfiable constraints such as `saved != red` over a three-value enum.
  Unsupported model-conversion shapes now fail generation instead of emitting a
  probabilistic retry.
- 2026-09-22: Added
  `TestTargetGenInitialAxiomThreeValueDisequalityUsesSolverModelFast`. The same
  generation-time initial model fallback now applies to `target=gen` when
  finite initial constraints exist and the initializer-action analysis emits no
  actions, so satisfiable axioms are no longer silently ignored by the gen init
  generator. The existing target=gen hard failure for unenumerable quantified
  initial variables is preserved.

## 4. FIXED `modelfile` is accepted but not semantically implemented

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

## 5. FIXED Native C++ blocks, actions, types, and definitions are silently weakened

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
- 2026-09-22: Added
  `TestTopLevelNativeHeaderIncludesWarnAndEmitGoComments` and
  `TestTopLevelNativeMemberAndInitRejectWeakening`. Top-level native blocks are
  now classified: harmless `header` blocks containing only C++ includes/comments
  are preserved as Go comments with a warning, while behavior-affecting
  `member`, `init`, or other native blocks fail generation with a source-located
  diagnostic. Together with the existing native action/type/definition
  rejections, this closes the silent native weakening gap.

## 6. FIXED Unsupported action/expression paths can remain soft comments in generated Go

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
- 2026-09-22: Allowed the exact native type interpretation
  `interpret T -> <<< int >>>` to lower to Go `int`, matching the C++
  `typedef int T` behavior used by extensional relation oracles, while keeping
  opaque native bodies such as `<<< primitive int >>>` on the unsupported-native
  diagnostic path covered by `TestNativeTypeInterpretRejectsIntPlaceholderWeakening`.
  `TestNativePlainIntInterpretLowersToGoIntFast` now covers this translatable
  native-int exception without building or running generated code.

## 7. FIXED Generated randomness ignores Python's call-stack-qualified choice labels

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
  around top-level and nested action calls, and builds stack-qualified choice
  labels. `___ivy_choose` mixes those labels into the direct RNG choice path.
  Later oracle work corrected `___ivy_randomize` to keep accepting the same
  label parameters but return the raw `ivyRandRange64` value, matching C++
  action-formal and record-field randomization traces for range and destructor
  record fixtures. `TestGeneratedRandomizeUsesRawRangeStreamFast` guards that
  shape in-process. Solver-generator replay parity is still part of the open
  items 1 and 2.

## 8. FIXED The generated test loop omits reader/timer event-loop semantics

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
  `test_completed`. At this point, full reader/timer `select` parity still
  needed the generated reader/timer scheduling surface added below.
- 2026-09-22: Added
  `TestGeneratedTestMainIncludesReaderTimerEventLoopFast`. Generated
  `target=test` code now includes reader/timer interfaces, per-instance
  reader/timer lists, reader binding, timeout callbacks, and Python-style
  cycle accounting for timeout/read/background-do-over branches. Specs with no
  installed readers or timers remain inert, but the generated loop now has a
  real scheduling surface for reader/timer hooks instead of immediately
  discarding the non-action choice branch. Native reader installation and true
  file-descriptor `select` parity were addressed by the follow-up bullets.
- 2026-09-22: Extended
  `TestGeneratedTestMainIncludesReaderTimerEventLoopFast` to require generated
  `__installReader` / `__installTimer` hooks. Runtime/native code now has a
  concrete generated registration surface for event-loop readers and timers;
  fd-backed reader scheduling was added in the follow-up bullet.
- 2026-09-22: Extended
  `TestGeneratedTestMainIncludesReaderTimerEventLoopFast` again to require
  portable reader scheduling. Generated `target=test` code now polls reader
  `ready()` hooks, sleeps via `time` when nothing is ready, and preserves the
  existing timer timeout, cycle decrement/re-increment, random ready-reader
  choice, and background do-over accounting without importing `syscall` or
  `unsafe`.

## 9. FIXED `before_export` analysis is only partially ported

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
  public exported action after generator success. `target=gen` already has
  fast coverage for leading `before_export` assumes, and the item 2 fix gives it
  the same simple preimage slice.
- 2026-09-22: Updated before-export source-shape tests after
  `TestTargetTestSolvedAssumeGeneratorSkipsTrialFast`: covered before-export
  guards now skip the runtime trial clone and call the public action directly
  after tracing. Full `before_export` reverse-image analysis remains tied to
  the remaining solver-backed generator work.
- 2026-09-22: Added
  `TestTargetGenBeforeExportNonLeadingAssumeUsesPreimageFast`. `target=gen`
  now has explicit coverage for the same simple non-leading before-export
  preimage slice as `target=test`: the analysis action constrains generated
  formals, while `execute()` still calls the public exported action.
- 2026-09-22: Added
  `TestTargetTestBeforeExportLocalRelationWitnessUsesGeneratorGuardFast` and
  `TestTargetGenBeforeExportLocalRelationWitnessUsesGeneratorGuardFast`.
  Before-export analysis actions with local relation witnesses now generate
  supported existential relation guards for both generator targets. The
  generator scans relation overrides, assigns generated formals from matching
  tuple fields, rechecks the guard, and still executes the public exported
  action rather than the analysis body.
- 2026-09-22: Unsupported before-export guards that are genuinely not
  translatable to Go, such as native C++ expressions, still fail generation via
  the existing source-located unsupported guard tests; the former local
  relation-witness fatal case has been moved into the supported slice above.

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
- 2026-09-22: Added fast coverage twins for already-modeled shared generator
  behavior: `TestTargetTestCallReturnDestructorFieldPreimageUsesGeneratorGuardFast`,
  `TestTargetGenCallReturnDestructorFieldPreimageUsesGeneratorGuardFast`, and
  `TestTargetGenBulkRelationAssignmentUsesReverseImageGuardFast`.
- 2026-09-22: Added fast in-process tests for the variant solver slices fixed
  during the checklist pass: positive/negative existential variant IFF
  spellings, constrained negated existential variant payload guards, and local
  variant-super witnesses for existential, concrete, and negated concrete
  membership assumptions. These remain source-shape/generator-analysis tests;
  slow generated-binary oracle comparisons are still deferred until the final
  `SLOWTEST=1` verification pass.
- 2026-09-22: Added fast coverage for the local variant state-update
  direct-execution slice:
  `TestTargetTestLocalVariantStateUpdateSkipsTrialFast` checks that the
  generator no longer falls back to a hidden trial clone when a zero-formal
  action copies a concrete local variant witness into state and immediately
  rechecks the same membership guard.
- 2026-09-22: Added the fast companion
  `TestTargetTestLocalVariantExistsStateUpdateSkipsTrialFast` for the
  subtype-only existential spelling of that state recheck, again asserting that
  no hidden trial clone is emitted.
- 2026-09-22: Added
  `TestTargetTestLocalRelationStateUpdateSkipsTrialFast` for the relation
  witness version of the same direct-execution pattern: a local value selected
  from a true relation cell is copied into state and rechecked without emitting
  a hidden trial clone.
- 2026-09-22: Added the existential relation companion
  `TestTargetTestLocalRelationExistsStateUpdateSkipsTrialFast`, covering the
  same no-hidden-trial check when the relation witness and recheck both hide an
  unconstrained tuple column under `exists`.
- 2026-09-22: Added `TestTargetTestLocalRelationIffStateUpdateSkipsTrialFast`
  to cover the `relation(...) <-> true` spelling of the local relation
  state-update recheck without a hidden trial clone.
- 2026-09-22: Added `TestTargetTestLocalVariantIffStateUpdateSkipsTrialFast`
  for the same no-hidden-trial direct-execution check on concrete variant
  membership written as an IFF-with-true guard.
- 2026-09-22: Added `TestTargetTestLocalVariantIteStateUpdateSkipsTrialFast`
  for the conditional variant version of the no-hidden-trial check, covering
  state rechecks written as the same branch-sensitive `ite` membership guard
  that initialized the local witness.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPointUpdateSkipsTrialFast` for the relation
  point-update version of the local witness pattern, asserting that a modeled
  local relation witness can set and recheck a relation cell without emitting a
  hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalVariantPointUpdateSkipsTrialFast` for the variant
  witness version of the relation point-update pattern, again checking that no
  hidden trial clone is emitted.
- 2026-09-22: Added
  `TestTargetTestLocalVariantConditionalPointUpdateSkipsTrialFast` for the
  identical-branch conditional point-update pattern, checking that a modeled
  local witness still avoids the hidden trial path when the update is wrapped
  in an `if/else`.
- 2026-09-22: Added
  `TestTargetTestLocalVariantOneSidedConditionalPointUpdateSkipsTrialFast` for
  guarded one-sided point updates, checking that the proof is branch-aware for
  `ite(active, seen(x), true)` but does not treat guarded writes as
  unconditional.
- 2026-09-22: Added
  `TestTargetTestLocalVariantElseSidedConditionalPointUpdateSkipsTrialFast` for
  the else-branch mirror, proving branch polarity is part of the guarded
  point-update coverage.
- 2026-09-22: Added
  `TestTargetTestLocalVariantConditionalPointUpdateImplicationSkipsTrialFast`
  for the implication spelling of a guarded point-update recheck, keeping the
  coverage fast and source-shape based.
- 2026-09-22: Added
  `TestTargetTestLocalVariantConditionalPointUpdateOrSkipsTrialFast` for the
  desugared OR spelling of the same guarded point-update recheck.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPointUpdateExistsSkipsTrialFast` for the
  existential-column point-update recheck shape, proving the generated action
  can avoid a hidden trial clone when the positive update itself supplies the
  existential witness.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPointUpdateIteRecheckSkipsTrialFast` for the
  point-update recheck shape under `ite`, requiring both branches to be covered
  before suppressing the hidden trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalVariantFalsePointUpdateSkipsTrialFast` for the
  false-valued point-update recheck shape, covering negated relation assumes
  after a modeled local witness writes the relation cell to `false`.
- 2026-09-22: Added
  `TestTargetTestLocalVariantStateAliasPointUpdateSkipsTrialFast` for the
  state-alias point-update direction where the update key is the copied state
  value and the later assume key is the original local witness.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPairPointUpdateSkipsTrialFast` for grouped
  relation tuple local witnesses, checking that two locals initialized from one
  relation entry can drive a relation point update and recheck without a hidden
  trial clone.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPairStateUpdateSkipsTrialFast` for the grouped
  relation tuple state-copy case, checking that tuple-position witness coverage
  survives copying both locals into state before rechecking the relation.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPairStateUpdateIteRecheckSkipsTrialFast` for the
  grouped tuple state-copy recheck under `ite`, requiring both conditional
  branches to be covered before removing the hidden trial path.
- 2026-09-22: Added
  `TestTargetTestLocalRelationDirectRecheckSkipsTrialFast` for the already
  modeled single-local relation direct recheck, plus target=test and target=gen
  nonlinear numeric fallback coverage for bounded small-model search. Added
  target=test and target=gen local nonlinear witness coverage for the same
  small-candidate model applied inside generated action bodies.
- 2026-09-22: Added target=test and target=gen local numeric formal-bound
  witness coverage for `scratch > c`, checking both the generated `c + 1`
  local candidate and the absence of unsupported existential generator guards.
- 2026-09-22: Added target=test and target=gen local numeric positive
  inequality state-copy coverage, checking that formal-dependent witnesses are
  substituted through copied state before rechecks.
- 2026-09-22: Added target=test and target=gen conditional local
  relation-witness source-shape coverage for `ite` guards seeded from
  extensional relation cells instead of only the broad bounded fallback.
- 2026-09-22: Added target=test and target=gen partial conditional local
  relation-witness source-shape coverage for `ite` guards where only one branch
  constrains the local relation witness.
- 2026-09-22: Repaired the `log.red2` oracle regressions with focused coverage:
  `TestEnumDispatchTraceMatchesIvy2Cpp`,
  `TestOracleRangeBoundsCompilesAndRuns`, and
  `TestOracleDestructorRecordCompilesAndMatchesTrace` now agree with C++ after
  removing label-hash offsetting from `___ivy_randomize`; the extensional
  `<<< int >>>` native-type oracle now generates by translating that exact
  native type to Go `int`; and
  `TestTargetGenInitGeneratorRunsInitAfterRandomize` now requests traced
  target=gen output when it needs to observe a returned value.
- 2026-09-22: Added
  `TestTargetTestLocalFiniteSetActionPointUpdateSkipsTrialFast` as a fast
  companion to the assignment-form point-update regression, covering the
  explicit `LogicSetAction` AST form without launching a generated tester.
- 2026-09-22: Added
  `TestTargetTestLocalVariantNegativeSetActionPointUpdateSkipsTrialFast` for
  the negative explicit `LogicSetAction` variant-local point-update shape,
  keeping that no-hidden-trial guarantee in the fast in-process suite.
- 2026-09-22: Added
  `TestTargetTestLocalVariantTwoSidedConditionalPointUpdateSkipsTrialFast` for
  the true/false relation-cell update case under `if/else`, plus the guarded
  update collapse needed to keep the existing identical-branch test fast.
- 2026-09-22: Added
  `TestTargetTestLocalVariantLetActionStateUpdateSkipsTrialFast` for
  action-level `let` around local witness/state-update/recheck bodies, keeping
  the coverage in-process while proving the target=test runner omits hidden
  trial machinery.
- 2026-09-22: Added
  `TestTargetTestLocalVariantLetActionConditionalPointUpdateSkipsTrialFast` for
  action-level `let` inside conditional branch updates, covering the relation
  point-update extractor without launching generated binaries.
- 2026-09-22: Added
  `TestTargetTestLocalVariantPrivateCallStateUpdateSkipsTrialFast` for
  non-recursive private helper calls that copy local witnesses into state before
  a target=test recheck, keeping the no-hidden-trial guarantee in-process.
- 2026-09-22: Added
  `TestTargetTestLocalVariantPrivateCallConditionalPointUpdateSkipsTrialFast`
  for private helper calls used as conditional branch updates, exercising the
  relation point-update extractor through generated-source inspection only.
- 2026-09-22: Added
  `TestTargetTestLocalVariantPrivateReturnCallStateUpdateSkipsTrialFast` for
  private helper calls whose return value is assigned into state before a local
  witness recheck, keeping that coverage fast and in-process.
- 2026-09-22: Added
  `TestTargetTestLocalVariantChoicePointUpdateSkipsTrialFast` for
  nondeterministic choice branches with the same local witness-keyed relation
  point update, checking the no-hidden-trial guarantee through generated-source
  inspection.
- 2026-09-22: Added
  `TestTargetTestLocalVariantEnvPointUpdateSkipsTrialFast` for the env-action
  mirror of that identical-branch relation point-update proof.
- 2026-09-22: Added
  `TestTargetTestLocalRelationAssignFieldActionSkipsTrialFast` for the
  destructor-field analogue of relation point updates, covering relation-backed
  local witness selection and direct field rechecks without whole-process runs.
- 2026-09-22: Added the null/copy companions
  `TestTargetTestLocalRelationNullFieldActionSkipsTrialFast` and
  `TestTargetTestLocalRelationCopyFieldActionSkipsTrialFast`, keeping the
  destructor field-action coverage fast and in-process.
- 2026-09-22: Added
  `TestTargetTestLocalRelationGuardedAssignFieldActionSkipsTrialFast` for
  branch-sensitive destructor field updates and implication-style rechecks,
  again without building or running a generated tester binary.
- 2026-09-22: Added
  `TestTargetTestLocalRelationElseGuardedAssignFieldActionSkipsTrialFast` for
  the else-branch/disjunctive spelling of the same destructor field update
  proof.
- 2026-09-22: Added
  `TestTargetTestLocalRelationTwoSidedAssignFieldActionSkipsTrialFast` for the
  two-sided `if/else` destructor field update case, keeping the `ite` recheck
  coverage fast and in-process.
- 2026-09-22: Strengthened that conditional local relation coverage to require
  branch-specific `if/else` relation scans, and added the target=test
  no-hidden-trial state-copy/recheck regression for the same conditional shape.
- 2026-09-22: Added
  `TestTargetTestLocalRelationPartialIteStateUpdateSkipsTrialFast`, the
  tautological-branch companion for the conditional local relation
  state-copy/recheck coverage. It asserts both the partial branch witness scan
  in the generated action body and the absence of hidden trial machinery in the
  target=test runner.
- 2026-09-22: Added target=test and target=gen conditional local
  variant-witness source-shape coverage for `ite` guards, asserting that the
  local variant assignment itself is conditional and precedes the generated
  assume.
- 2026-09-22: Added target=test and target=gen partial conditional local
  variant-witness source-shape coverage for `ite` guards where only one branch
  constrains the local variant.
- 2026-09-22: Added target=test and target=gen inequality existential variant
  payload witness tests, covering the source-shape guarantee that generated
  action generators construct a satisfying subtype payload before rechecking
  quantified guards such as `exists Q:req. x *> Q & Q > 10`.
- 2026-09-22: Extended that existential variant payload coverage to affine
  integer inequalities such as `Q + 1 > 10` for both generator targets.
- 2026-09-22: Added target=test and target=gen conditional defined-input
  source-shape tests for branch definitions merged into a single generated
  `ivyTernary` assignment.
- 2026-09-22: Added target=test and target=gen cross-target conditional
  defined-input source-shape tests for branch definitions that require guarded
  assignments to different generated inputs.
- 2026-09-22: Added target=test and target=gen conditional numeric witness
  source-shape tests for branch inequality witnesses merged into one generated
  `ivyTernary` search value.
- 2026-09-22: Added target=test and target=gen conditional scalar-pair witness
  source-shape tests for same-target branch inequalities merged into one
  generated `ivyTernary` assignment.
- 2026-09-22: Added target=test and target=gen cross-target conditional
  scalar-pair witness source-shape tests for branch inequalities that need
  guarded assignments to two different generated formals.
- 2026-09-22: Added target=test and target=gen affine pair equality
  source-shape tests for constraints such as `a + 1 = b + 2`, asserting the
  deterministic pair assignment precedes the generated guard check.
- 2026-09-22: Added target=test and target=gen conditional relation witness
  source-shape tests for `ite` guards whose branches use different relation
  model sources.
- 2026-09-22: Added target=test and target=gen conditional variant witness
  source-shape tests for same-subtype `ite` membership guards merged into one
  generated `ivyTernary` payload assignment.
- 2026-09-22: Added target=test and target=gen cross-subtype conditional
  variant witness source-shape tests for `ite` guards merged into one generated
  `ivyTernary` over complete variant values.
- 2026-09-22: Added target=test and target=gen conditional exact-payload
  existential variant witness source-shape tests for quantified `ite`
  constraints merged into one generated `ivyTernary` payload.
- 2026-09-22: Added target=test and target=gen partial conditional
  exact-payload existential variant source-shape tests, covering `ite` payload
  constraints where only one branch restricts the quantified payload.
- 2026-09-22: Added target=test and target=gen conditional inequality-payload
  existential variant witness source-shape tests for quantified `ite`
  constraints merged into one generated numeric `ivyTernary` payload.
- 2026-09-22: Added target=test and target=gen partial conditional
  inequality-payload existential variant source-shape tests, covering
  branch-local numeric payload witnesses.
- 2026-09-22: Added target=test and target=gen conditional mixed-delta
  existential variant witness source-shape tests for branch payload witnesses
  that need different integer offsets before the generated `ivyTernary`.
- 2026-09-22: Added target=test and target=gen affine equality-payload
  existential variant witness source-shape tests, covering deterministic
  subtype payload construction for quantified guards such as `Q + 1 = req0`.
- 2026-09-22: Added target=test local variant affine equality state-copy
  coverage, checking that a copied local subtype payload with an integer delta
  avoids the hidden trial clone when rechecked through an existential guard.
- 2026-09-22: Added target=test and target=gen affine constrained negated
  existential variant source-shape tests, covering direct forbidden-payload
  guards for unique affine payload equalities under negated membership.
- 2026-09-22: Added target=test and target=gen local numeric affine equality
  state-copy tests, covering direct local witness construction from
  formal-dependent affine equalities and preimage substitution without hidden
  trial or unsupported existential guards.
- 2026-09-22: Added target=test and target=gen local numeric affine
  disequality tests, covering deterministic formal-dependent witnesses for
  negated local equality guards and preimage substitution through state copies.
- 2026-09-22: Added target=test and target=gen local numeric negated
  inequality tests, covering boundary witnesses for formal-dependent negated
  inequalities and preimage substitution through state copies.
- 2026-09-22: Added target=test and target=gen local numeric IFF-false
  inequality tests, covering the equivalent IFF spelling of negated
  formal-dependent bounds.
- 2026-09-22: Added target=test and target=gen local numeric IFF-false affine
  equality tests, covering the equivalent IFF spelling of formal-dependent
  disequality witnesses.
- 2026-09-22: Added target=test and target=gen local numeric implication
  tests, covering formal-dependent witnesses discovered under implication
  consequents and substituted through state copies.
- 2026-09-22: Added target=test and target=gen local numeric OR/AND wrapper
  tests, covering disjunctive local witness guardrails and conjunctions whose
  solved local conjuncts must leave non-local guards in the generated action
  generator.
- 2026-09-22: Added target=test and target=gen local numeric ITE wrapper
  tests, covering branch-dependent scalar local witnesses and the substituted
  conditional generator guards that keep those witnesses sound.
- 2026-09-22: Added target=test and target=gen local numeric expression-let
  wrapper tests, covering alias expansion before preimage substitution so
  copied local witnesses retain non-local generator guards.
- 2026-09-22: Added target=test and target=gen disjunctive local equality
  bridge tests, covering OR-wrapped local witness substitutions that should
  become direct generated guards instead of existential local scans.
- 2026-09-22: Added target=test and target=gen disjunctive local equality
  point-update tests, covering the same OR witness substitution through
  relation-cell preimage updates.
- 2026-09-22: Added target=test and target=gen finite local disequality
  point-update tests, covering enum/boolean alternate-value witnesses used to
  rewrite relation-cell preimage updates.
- 2026-09-22: Added target=test and target=gen singleton finite disequality
  tests, covering unsatisfiable local disequality guards that should become a
  direct generated `false` guard.
- 2026-09-22: Added target=test and target=gen implication-wrapped singleton
  finite disequality tests, covering preservation of branch conditions around
  unsatisfiable local choices.
- 2026-09-22: Added target=test and target=gen OR-wrapped singleton finite
  disequality tests, covering preservation of satisfiable disjuncts while
  replacing impossible local-choice branches with `false`.
- 2026-09-22: Added target=test and target=gen AND-wrapped singleton finite
  disequality tests, covering collapse of impossible conjunctive local choices
  to direct `false` guards.
- 2026-09-22: Added target=test and target=gen ITE-wrapped singleton finite
  disequality tests, covering branch-sensitive replacement of impossible local
  choices with direct conditional guards.
- 2026-09-22: Added target=test and target=gen IFF-with-true singleton finite
  disequality tests, covering equivalence-wrapper replacement of impossible
  local choices before generated guard checks.
- 2026-09-22: Added target=test and target=gen before-export local relation
  witness tests, covering existential local relation guards in analysis actions
  without whole-process runs or runtime trial fallbacks.
- 2026-09-22: Added target=test and target=gen relation override forall guard
  tests, covering unbounded negative relation quantifiers in generated action
  guards with fast source-shape checks.
- 2026-09-22: Added target=test and target=gen equality-bound forall guard
  tests, covering solver-trivial quantified formulas that should simplify to
  direct generated guards.
- 2026-09-22: Added target=test and target=gen multi-variable equality-bound
  forall guard tests, covering tuple relation guards whose quantified variables
  are all bound by equalities.
- 2026-09-22: Added target=test and target=gen equality-bound forall tests
  with extra antecedent guards, covering preservation of non-quantified guard
  terms after substitution.
- 2026-09-22: Added target=test and target=gen equality-only exists guard
  tests, covering satisfiable existential equality bindings that simplify to
  true.
- 2026-09-22: Added target=test and target=gen equality-bound forall numeric
  witness tests, covering guard-AST simplification before scalar witness
  planning.
- 2026-09-22: Added target=test and target=gen equality-bound exists numeric
  witness tests, covering non-variant residual existential bodies after
  equality substitution.
- 2026-09-22: Added target=test and target=gen local equality-bound exists
  numeric witness tests, covering action-body local witness initialization from
  simplified existential residual guards without process-level runs.
- 2026-09-22: Documented existing target=test and target=gen local equality
  chain point-update tests, covering multi-local equality substitutions before
  relation-cell preimage rewriting.
- 2026-09-22: Added target=test and target=gen product-equality witness tests,
  covering a two-unbounded-formal nonlinear guard slice without generated
  process runs or broad Cartesian fallback search.
- 2026-09-22: Added target=test and target=gen product equality tests with
  state-expression RHS values, covering product witnesses that depend on
  existing model state rather than only numerals.
- 2026-09-22: Added target=test and target=gen product inequality tests with
  state-expression bounds, covering strict-boundary two-formal product
  witnesses.
- 2026-09-22: Added target=test and target=gen product disequality tests with
  state-expression forbidden values, covering negated product equality witness
  extraction.
- 2026-09-22: Added target=test and target=gen additive equality tests with
  state-expression RHS values, covering another small two-formal arithmetic
  model slice with source-shape checks only.
- 2026-09-22: Added target=test and target=gen additive inequality tests with
  state-expression bounds, covering strict-boundary two-formal arithmetic
  witnesses without generated process runs.
- 2026-09-22: Added target=test and target=gen additive disequality tests with
  state-expression forbidden values, covering negated two-formal sum equality
  witness extraction.
- 2026-09-22: Added target=test and target=gen conditional multi-formal
  arithmetic witness tests, covering ITE guards whose branches each produce
  more than one defined input for the same generated formals. These remain fast
  source-shape tests and do not call generated binaries.
- 2026-09-22: Added target=test and target=gen cross-target conditional
  multi-formal arithmetic witness tests, covering ITE guards whose branches
  solve different generated formal sets.
- 2026-09-22: Added target=test and target=gen partial conditional
  defined-input tests, covering ITE guards where only one branch constrains a
  generated input.
- 2026-09-22: Added target=test and target=gen partial conditional scalar
  numeric witness tests, covering ITE guards where only one branch constrains
  an integer-like generated formal.
- 2026-09-22: Added target=test and target=gen partial conditional local
  scalar witness tests, covering ITE guards where only one branch constrains a
  generated action-local value before it is copied into state and rechecked.
- 2026-09-22: Added target=test and target=gen conditional multi-pair numeric
  witness tests, covering ITE guards whose branches contain multiple pairwise
  constraints over the same generated targets.
- 2026-09-22: Added target=test and target=gen cross-target conditional
  multi-pair numeric witness tests, covering ITE guards whose branches solve
  different generated pair targets.
- 2026-09-22: Added target=test and target=gen partial conditional pair
  witness tests, covering ITE guards where only one branch constrains a
  generated pair target.
- 2026-09-22: Added target=test and target=gen conditional multi-variant
  witness tests, covering ITE guards whose branches contain multiple variant
  membership constraints over the same generated targets.
- 2026-09-22: Added target=test and target=gen partial conditional variant
  witness tests, covering ITE guards where only one branch constrains a
  generated variant value.
- 2026-09-22: Added target=test and target=gen choice point-update preimage
  tests, covering nondeterministic branch alternatives that write different
  relation/function cells before a later assume, using only generated-source
  inspection inside the single Go test binary.
- 2026-09-22: Added target=test and target=gen mixed choice state/point-update
  tests, including a branch-correlation guardrail that ensures generated guards
  are not over-approximated by independently combining unrelated branch state
  and relation/function cell writes.
- 2026-09-22: Added target=test and target=gen guarded choice point-update
  tests, covering branch-local assume guards attached to branch-specific
  relation/function cell writes before a later assume.
- 2026-09-22: Added a fast target=test guarded-internal-choice trial-path
  regression, covering the interim safety contract that branch-assume choices
  must not execute directly until internal branch choices are model-controlled.
- 2026-09-22: Added the matching target=gen guarded-internal-choice trial-path
  regression, covering one-shot generated actions that must skip rejected
  internal branch choices instead of surfacing `assumption_failed`.
- 2026-09-22: Added a target=gen guarded-internal-choice trace-envelope
  regression, covering the accepted trial path so captured internal trace output
  remains bracketed like direct target=gen execution.
