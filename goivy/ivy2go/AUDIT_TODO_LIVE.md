# ivy2go live audit / TODO log

Created: 2026-05-25 06:06:56 UTC
Last updated: 2026-05-25 (M10 close)

A living catalogue of known gaps between `ivy2go` and `ivy2cpp`. Items
land DONE once their tests pass; OPEN items have a known fix path and
typically a `// deferred (Mxx)` marker in source pointing here.

Format mirrors `ivy2cpp/AUDIT2_*.md`:

> ## STATUS NNN — short title
> Created: YYYY-MM-DD HH:MM:SS UTC
>
> **Gap.**
> **ivy2cpp references.** (file:line)
> **ivy2go plan.** (file:line "land here")
> **Verification.** (test name)

Statuses: `OPEN`, `IN PROGRESS`, `DONE`.

---

## DONE 001 — M0 bootstrap

Created: 2026-05-25 06:06:56 UTC

CLAUDE.md, ARCHITECTURE_TODO.md, doc.go, this file. Package compiles.

## DONE 002 — M1 skeleton

Created: 2026-05-25.

`writer.go`, `go_context.go`, `config.go`, `generator.go`, `compile.go`,
`names.go`. Empty module → Output.Files contains state.go / init.go /
runtime.go, all gofmt-clean.

## DONE 003 — M2 type emission

Created: 2026-05-25.

`go_types.go`, `types.go`, plus M2 stubs `destructor.go` / `variant.go` /
`constructors.go`. Enum / range / uninterpreted lowering works; bv
interpType parsing in place.

## DONE 004 — M3 expression emission

Created: 2026-05-25.

`expr.go`, `bv_expr.go`, `runtime.go` (helpers). emitExpr core dispatch,
emitApply with macro expansion + infix + BV path, BV operator lowering
for widths ≤ 64, Uint128 runtime stub for widths ≤ 128.

## DONE 005 — M4 action emission

Created: 2026-05-25.

`action.go`, `assign.go`, plus stubs `nondet.go` / `extensional.go` /
`definitions.go`. Sequence / Assert / Assume / If / While / Choice /
Call / Local / Return / IgnoreAction handled; scalar-LHS assignment
emits `s.X = rhs`. emitActionMethods walks Mod.Actions and emits one
*State method per action.

## DONE 006 — M5 state + init + impl target + build smoke

Created: 2026-05-25.

`state.go`, `initial_state.go`, `init.go`, `build.go`. State struct
walks Sig.Symbols; function symbols use scalar/array/map storage per
goFunctionStorageFor. NewState allocates maps; Init populates scalars
via ivyChoose. Tier 2 smoke (SLOW_GO_TEST=1) builds the emitted impl
package with `go build`.

## DONE 007 — M6 class target smoke

Created: 2026-05-25.

target=class emits the same library-shaped package (no main.go, no
repl.go); Tier 2 smoke builds both a flag-only fixture and one with
mixed types (enum + range + relation + action).

## DONE 008 — M7 REPL target

Created: 2026-05-25.

`repl.go`, `tick.go`, `vprint.go`. Simple bufio.Scanner-based REPL,
per-sort arg parsers, action-name dispatch, exit/quit support. Tier 2
smoke builds an emitted REPL binary.

## DONE 009 — M8 hash-thunk emission (runtime path)

Created: 2026-05-25.

`thunk.go`, `native_thunk.go` (stub). makeThunk emits a Go struct +
constructor + get() method per distinct large-domain assignment.
Memoization dedups identical thunks. The Z3-aware toZ3 path is
deferred to M9.1+.

## DONE 010 — M9 test/gen target (Z3 via goivy.Solver)

Created: 2026-05-25.

`z3.go`, `solver_emit.go`, `action_gen.go`. Per-action actionGen<Name>
struct emitted for test/gen targets, holding *goivy.Solver. Generated
code imports `github.com/glycerine/ivy/goivy`. Tier 2 smoke builds the
emitted test-target package against the in-tree goivy via go.mod
replace.

## DONE 011 — M10 native blocks + hygiene + docs

Created: 2026-05-25.

`native.go` (go / go_header / go_init / go_inline tag dispatch),
`oracle_compare.go` (stub for M11+ behavioural oracle),
`comments_test.go` (Tier 5: every TODO/DEFER must anchor an audit
item or milestone).

---

## OPEN — known gaps deferred to follow-up milestones

### OPEN 050 — Quantifier emission

`expr.go` `emitQuant` returns "deferred (M4/M5)". Loop-header
generation for forall / exists over finite sorts needs to mirror
ivy2cpp's `quantIterableHeader` / `loopHeaderForSortBounds`.

**Plan.** Port `quantIterableHeader`, `loopHeaderForVar`, and the
extensional-quant path from ivy2cpp/expr.go. Tests: every formula in
test_vec/oracle that exercises forall/exists.

### OPEN 051 — `if some` lowering

`action.go` `emitIf` short-circuits to `unsupported` when the cond is
a `SomeCondition`. ivy2cpp's emitIfSome handles min/max, extensional,
and variant-downcast variants.

**Plan.** Port emitIfSome shape from ivy2cpp/action.go; reuse the
quantifier helpers from OPEN 050.

### OPEN 052 — Quantified-LHS assignment

`assign.go` `emitAssign` falls through to `unsupported` when the LHS
has free variables. ivy2cpp's emitAssignTwoPhase and emitAssignLarge
handle this with two-phase loops or thunk wrapping.

**Plan.** Reuse the loop-header helpers from OPEN 050 + the thunk
machinery from `thunk.go` (DONE 009).

### OPEN 053 — Destructor + variant struct emission

`destructor.go` and `variant.go` are M2 stubs. ivy2cpp emits record
structs with Equal/Hash/Less methods and tagged-union structs.

**Plan.** Port `emitDestructorStruct`, `emitVariantSuperStruct`,
`emitVariantLeafStruct` and the field-access path through emitApply's
emitDestructorApply / emitVariantRelation.

### OPEN 054 — Wide BV (>128) operator lowering

`bv_expr.go` `emitBVApply` returns "wide-BV lowering deferred" for
widths > 64. M3 wired *big.Int helpers but operator emission still
uses uint64 / Uint128 only.

**Plan.** Per ARCHITECTURE_TODO.md §3.5.7 R5, *big.Int is acceptable
for the >128 path. Add bv_expr.go cases that route via big.Int
methods masked by bigIntMask.

### OPEN 055 — Real solver-driven test-gen

`action_gen.go` `Generate` uses ivyChoose for input selection. The
M9 skeleton constructs goivy.Solver but doesn't yet use
Solver.GetSmallModel to pick inputs satisfying the action's
reverse-image precondition.

**Plan.** In a follow-up M9.1: compute the reverse-image expression
via goivy.ModifiesSingle + action-update analysis, assert it through
Solver.Assert, call GetSmallModel, read back via the model.

### OPEN 056 — Native action emission

`action.go` `emitAction` LogicNativeAction case returns "deferred
(M10)". M10 wired free / header / init / inline native blocks but
native *actions* (inside an action body) still error out.

**Plan.** Reuse `collectNativeGoBlocks` / `splitNativeGoCode` but for
in-action blocks; emit body verbatim into the action method.

### OPEN 057 — Behavioural oracle harness

`oracle_compare.go` is a stub. Per ARCHITECTURE_TODO.md §3.8 Tier 4,
this is the M11+ behavioural comparison of ivy2cpp- and ivy2go-
emitted binaries on shared fixtures.

**Plan.** Implementation lives in M11; the seam is the
`CompareGoVsCpp` function already exported.

### OPEN 058 — Trace-LHS emission

`assign.go` `emitAssignSimple` skips the trace prelude entirely.
ivy2cpp emits `__ivy_out << "write(...)"` for any Config.Trace=true
assignment.

**Plan.** Port `emitTracedLHS` from ivy2cpp/assign.go; route through
`vprint.go` so the number-format toggle (currently a stub) drives the
hex/decimal choice.
