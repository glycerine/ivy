# ivy2go live audit / TODO log

Created: 2026-05-25 06:06:56 UTC
Last updated: 2026-05-25 (OPEN-items pass)

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

Statuses: `OPEN`, `IN PROGRESS`, `DONE`, `PARTIAL`.

---

## DONE 001 — M0 bootstrap

Created: 2026-05-25 06:06:56 UTC

CLAUDE.md, ARCHITECTURE_TODO.md, doc.go, this file. Package compiles.

## DONE 002 — M1 skeleton

`writer.go`, `go_context.go`, `config.go`, `generator.go`, `compile.go`,
`names.go`. Empty module → Output.Files contains state.go / init.go /
runtime.go, all gofmt-clean.

## DONE 003 — M2 type emission

`go_types.go`, `types.go`, plus M2 stubs `destructor.go` / `variant.go` /
`constructors.go`. Enum / range / uninterpreted lowering works.

## DONE 004 — M3 expression emission

`expr.go`, `bv_expr.go`, `runtime.go` (helpers). emitExpr core dispatch,
emitApply with macro expansion + infix + BV path, BV operator lowering
for widths ≤ 64.

## DONE 005 — M4 action emission

`action.go`, `assign.go`, stubs `nondet.go` / `extensional.go` /
`definitions.go`. Sequence / Assert / Assume / If / While / Choice /
Call / Local / Return handled. emitActionMethods walks Mod.Actions
and emits one *State method per action.

## DONE 006 — M5 state + init + impl target + build smoke

`state.go`, `initial_state.go`, `init.go`, `build.go`. Tier 2 smoke
(SLOW_GO_TEST=1) builds the emitted impl package with `go build`.

## DONE 007 — M6 class target smoke

target=class emits a library-shaped package (no main.go, no repl.go);
Tier 2 smoke builds both a flag-only fixture and one with mixed types.

## DONE 008 — M7 REPL target

`repl.go`, `tick.go`, `vprint.go`. Simple bufio.Scanner-based REPL,
per-sort arg parsers, action-name dispatch.

## DONE 009 — M8 hash-thunk emission (runtime path)

`thunk.go`, `native_thunk.go` (stub). makeThunk emits a Go struct +
constructor + get() method per distinct large-domain assignment.

## DONE 010 — M9 test/gen target (Z3 via goivy.Solver)

`z3.go`, `solver_emit.go`, `action_gen.go`. Per-action actionGen<Name>
struct emitted for test/gen targets, holding *goivy.Solver. Tier 2
smoke builds the emitted test-target package against the in-tree
goivy via go.mod replace.

## DONE 011 — M10 native blocks + hygiene + docs

`native.go` (go / go_header / go_init / go_inline tag dispatch),
`oracle_compare.go` (stub for M11+ behavioural oracle),
`comments_test.go` (Tier 5: every TODO/DEFER must anchor an audit
item or milestone).

## DONE 050 — Quantifier emission

OPEN-pass.

`expr.go` `emitQuant` emits a Go IIFE with early-exit loops over the
quantifier variable's sort. `loopHeaderForVar` / `loopHeaderForSort`
handle bool, enum (with named-type cast), range, and integer-like
uninterpreted sorts with known cardinality. Multi-variable
quantifiers nest loops; empty-vars fall through to the body.
Verification: `TestEmitQuant_*` in `quant_test.go`.

## DONE 051 — `if some` lowering

`action.go` `emitIfSome` emits a witness/found-flag scan: declares
witnesses, runs the loop nest, sets the flag + witnesses on the
first hit, dispatches THEN inside the if-block, then dispatches
ELSE gated on !__found. `some_min` / `some_max` still defer with a
clear marker (follow-up).
Verification: `TestEmitIfSome_*` in `quant_test.go`.

## DONE 052 — Quantified-LHS assignment

`assign.go` emitAssign now routes free-var LHSes through
`emitAssignTwoPhase`: allocate a temp of the LHS storage shape,
fill it from RHS in pass 1, copy back into the LHS in pass 2. The
thunk fallback for unbounded vars stays a follow-up. The two-phase
shape avoids RHS-reads-LHS aliasing.
Verification: `TestEmitAssign_*` in `assign_test.go`.

## DONE 053 — Destructor record struct emission

`destructor.go` `emitDestructorStruct` emits a Go struct per Ivy
record sort with one exported field per destructor (scalar or
goFunctionStorageFor-lowered). An Equal method does field-by-field
compare. `emitDestructorApply` in `expr.go` lowers reads through
field access. Variants emit a tagged-union super struct; per-leaf
destructors get the record path.
Verification: `TestEmitDestructor_*` in `destructor_test.go`.

## DONE 054 — Wide BV (>64 bits) operator lowering

`bv_expr.go` `emitWideBVApply` routes all BV operators with result
width > 64 through math/big. `runtime.go` `emitBigIntHelpers`
emits the wideBV* arithmetic, shift, and mask helpers plus a
toBigInt converter that admits uint32/uint64/Uint128/*big.Int /
int. `primitiveType` now returns "" for widths > 64 so the type
path uniformly uses *big.Int. Tier 2 smoke
(`TestSmoke_BuildEmittedWideBV`) compiles a bv[96] arithmetic
fixture.

## DONE 056 — In-action native go blocks

`action.go` `emitNativeAction` reads the LogicNativeAction's code,
splits the tag line via `splitNativeGoCode`, and emits the body
verbatim into the enclosing action method when the tag starts
with "go". cpp-tagged in-action blocks are skipped silently so
mixed sources work.
Verification: `TestEmitNativeAction_*` in `native_test.go`.

## DONE 058 — Trace-LHS emission

`vprint.go` `emitTracedLHS` emits an `fmt.Fprintf(ivyTraceOut, ...)`
call describing the assignment when `Config.Trace=true`.
`numberFormat` consults module attributes for hex tracing.
`runtime.go` declares `ivyTraceOut io.Writer = os.Stdout` only
when Trace is enabled.
Verification: `TestEmitTrace_*` in `assign_test.go`.

---

## PARTIAL 055 — Real solver-driven test-gen

OPEN-pass landed the Solver round-trip: pushStateIntoSolver now
calls `goivy.Solver.IsSat` to verify state consistency at
Generate-time entry, and the runtime emission proves the
goivy.NewConst + Solver.IsSat wiring works end-to-end (the
test-target Tier 2 smoke builds and links against goivy).

What remains as **OPEN 055.1**: per-symbol state assertions and
reverse-image precondition derivation. Today input selection in
the action generators falls back to ivyChoose; to be true
solver-driven generation, each action's Generate method needs to:

  1. Build the action's reverse-image formula via
     `goivy.ModifiesSingle` + action-update analysis (mirrors
     ivy2cpp's action_gen reverse_image computation).
  2. Assert it via `Solver.Assert(...)`.
  3. Call `Solver.GetSmallModel` to find a satisfying assignment.
  4. Read back each input via `ModelResult.Eval`.

This requires substantial integration with goivy's action-analysis
pipeline (not just the solver facade) and is a project on its own.

Verification of the current Solver wiring: `TestEmit_TestTarget_*`
in `action_gen_test.go`.

---

## OPEN — known gaps still deferred

### OPEN 055.1 — Full reverse-image-driven action input synthesis

(See PARTIAL 055 above for context.)

### OPEN 057 — Behavioural oracle harness

`oracle_compare.go` is a stub. Per ARCHITECTURE_TODO.md §3.8 Tier 4,
this is the M11+ behavioural comparison of ivy2cpp- and ivy2go-
emitted binaries on shared fixtures. Explicitly post-MVP.

### OPEN 060 — `if some` min/max lowering

`action.go` emitIfSome short-circuits to a deferral marker when
the SomeCondition Kind is `some_min` or `some_max`. The full
ivy2cpp emitIfSomeMinMax shape tracks the best index across the
loop scan and dispatches THEN with the winning witness.

### OPEN 061 — Quantified-LHS thunk fallback

`assign.go` `emitAssign` routes free-var LHSes to two-phase when
loops are openable; otherwise it reports `unsupported`. The thunk
machinery from M8 + `makeThunk` should plug in here as the
last-resort path.

### OPEN 062 — Destructor Hash / Less methods

`destructor.go` emits Equal but not Hash / Less. Hash isn't
needed today because destructor structs aren't used as Go map
keys (we use tup__T1__T2 for multi-arg map indexing); Less is
useful when ivy code does sort-comparing on records and not yet
exercised by tests.

### OPEN 063 — Variant constructor + downcast emission

`variant.go` emits the super struct shape but no constructor
helpers (`NewAnimal_Cat(c)`) and no `emitVariantRelation` /
`emitDestructorApply` integration for the `*>` downcast operator.
emitExpr / emitApply still report variant operators as deferred.

### OPEN 064 — Native antiquote substitution

`native.go` `emitNativeBlocks` and `action.go` `emitNativeAction`
emit the body verbatim. ivy2cpp substitutes `$arg` references via
`renderNativeTemplate`; for parity we'd need a Go equivalent.
Today users must write antiquote-free native blocks.
