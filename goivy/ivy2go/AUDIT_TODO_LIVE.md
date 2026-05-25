# ivy2go live audit / TODO log

Created: 2026-05-25 06:06:56 UTC
Last updated: 2026-05-25 (follow-up pass)

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

## DONE 001–011 — M0 through M10 (initial implementation pass)

See git history for the milestone-by-milestone landings.

## DONE 050 — Quantifier emission

`expr.go` `emitQuant` emits a Go IIFE with early-exit loops over the
quantifier variable's sort. `loopHeaderForVar` / `loopHeaderForSort`
handle bool, enum (with named-type cast), range, and integer-like
uninterpreted sorts with known cardinality.

## DONE 051 — `if some` lowering

`action.go` `emitIfSome` emits a witness/found-flag scan. THEN
dispatches inside the `if !__found && (cond)` block with witnesses
bound; ELSE dispatches gated on `!__found`.

## DONE 052 — Quantified-LHS assignment (two-phase)

`assign.go` emitAssign routes free-var LHSes through
`emitAssignTwoPhase`: temp allocation + fill + copyback.

## DONE 053 — Destructor + variant struct emission

`destructor.go` emits Go record struct + Equal method; `variant.go`
emits tagged-union super struct.

## DONE 054 — Wide BV (>64 bits) operator lowering

`bv_expr.go` `emitWideBVApply` routes all BV operators with result
width > 64 through math/big. `runtime.go` `emitBigIntHelpers` emits
the wideBV* arithmetic, shift, mask, and toBigInt helpers.

## DONE 055 — Solver-integration scaffolding

`solver_emit.go` `pushStateIntoSolver` calls `goivy.Solver.IsSat` for
end-to-end SAT-check wiring. Test-target smoke compiles against
goivy.

## DONE 055.1 — Solver-driven input synthesis

`action_gen.go` `Generate` now constructs `goivy.NewConst` symbols
for each formal param, builds a `goivy.NewClauses(nil, nil)`,
calls `g.sol.GetModelClauses(...)`, and extracts each input via
the `pickBoolOrChoose` / `pickUintOrChoose` runtime helpers
(which themselves fall back to `ivyChoose` when the model can't
supply a value). The reverse-image precondition is still trivial
(`true`) — replacing it with a real precondition only requires
changing the Clauses construction; the rest of the wiring is in
place.

## DONE 056 — In-action native go blocks

`action.go` `emitNativeAction` reads the LogicNativeAction code,
splits via `splitNativeGoCode`, and emits the body verbatim into the
enclosing action method when the tag starts with "go".

## DONE 057 — Behavioural oracle harness

`oracle_compare.go` ships `CompareGoVsCpp(goBin, cppBin, stdin)`:
runs both binaries, captures stdouts, normalises (CRLF + trailing
whitespace), and reports the first divergent line via the typed
`*OracleDiff` error. Verified by `oracle_compare_test.go` with
shell-script stand-ins for the real ivy2cpp / ivy2go binaries.
Fixture orchestration (build both sides, iterate over .in
transcripts) layers on top in a future cmd/ harness — `CompareGoVsCpp`
is the seam.

## DONE 058 — Trace-LHS emission

`vprint.go` `emitTracedLHS` writes `fmt.Fprintf(ivyTraceOut, ...)`
calls; runtime declares `ivyTraceOut` only when `Config.Trace=true`.

## DONE 060 — `if some` min/max lowering

`action.go` `emitIfSomeMinMax` declares `__found` + `__best_idx` and
per-param witnesses, scans the loop nest, updates the best index
when the cond holds and the index strictly beats the current best
(`<` for some_min, `>` for some_max). Dispatches THEN/ELSE on
`__found`.

## DONE 061 — Quantified-LHS thunk fallback

`assign.go` `emitAssignLarge` wraps the RHS in a thunk via M8's
`makeThunk` when the loop bounds aren't derivable. A residual
read-side wiring follow-up is tracked as OPEN 061.1 below — the
thunk is built but reads of the LHS still go through the map
storage.

## DONE 062 — Destructor Hash / Less methods

`destructor.go` `emitDestructorHash` + `emitDestructorLess` emit
the corresponding methods on each record struct. `runtime.go` ships
the on-demand `mixHash(uint64, any) uint64` and `lessOrd(a, b any) bool`
dispatchers (FNV-1a for primitives; delegate to `Hash()` / `Less()`
for user types).

## DONE 063 — Variant constructors + `*>` downcast

`variant.go` `emitVariantSuperStruct` emits per-leaf constructors
(`NewSuperLeaf(v) Super` or `NewSuperLeaf() Super` for plain
leaves). `expr.go` `emitVariantRelation` lowers `super *> sub` to a
tag-check + payload-equality expression.

## DONE 064 — Native antiquote substitution

`native.go` `renderNativeGoTemplate` walks the body for
backtick-delimited `` `N` `` indices and substitutes each with the
emitExpr-rendered code for params[N]. Both module-level
(`emitNativeBlocks`) and in-action (`emitNativeAction`) emission
route through it.

---

## DONE 055.2 — State-fact precondition (per-action)

`solver_emit.go` `emitSetSolver` now also emits two runtime helpers:

  - `stateFactsAsClauses(state *State) *goivy.Clauses` builds a
    Clauses asserting that every scalar bool state symbol equals
    its current value. Each scalar bool symbol contributes
    `mkBoolFact(name, value)` — the symbol itself or its negation,
    depending on the state value.
  - `mkBoolFact(name string, val bool) goivy.Expr` returns either
    `goivy.NewConst(name, goivy.Boolean)` or
    `&goivy.LogicNot{Body: …}` matching the encoding goivy.Solver
    expects.

Each actionGen's Generate now calls `stateFactsAsClauses(state)` to
seed `GetModelClauses`, so the synthesised inputs are consistent
with the current state. Function-sorted symbols + the action's
real Pre derivation are the next sub-step (OPEN 055.3 below).

## DONE 061.1 — Thunk read-side wiring

`state.go` `emitStateStruct` now declares a parallel
`__thunk_<Sym> func(K) V` field next to each hash-thunk-backed
storage field, and `emitStateGetters` emits a `(s *State) get<Sym>(k)`
helper that consults `map → thunk → zero`. `expr.go`
`goStorageAccess` routes read-context map access through that
getter; write context (LHS of assignment) keeps the raw indexed
form via the new `Generator.lhsContext` flag toggled by
`emitAssignSimple` / `emitAssignTwoPhase`. `emitAssignLarge` now
clears the map and installs the thunk on `s.__thunk_<Sym>`.

## DONE 062.1 — Deterministic hash for hash-thunk fields

`destructor.go` `emitDestructorHash` now emits a per-key sort + ordered
per-entry hash for map fields: collect keys into a slice, run
`sort.Slice` with the `lessOrd` runtime helper, then mix each
`(key, value)` pair into the FNV-1a accumulator. Avoids the
non-deterministic Go map iteration order.

## DONE 064.1 — Antiquote prefix flavours

`native.go` `renderNativeGoTemplate` now recognises two flavour
prefixes on the text preceding a `` `N` `` antiquote:

  - `%` → substitute with `g.goType(params[N].NodeSort())`
    (the Go type expression for the param's sort).
  - `"` → substitute with the bare identifier name of params[N]
    so an enclosing `"..."` stays well-formed.

Default flavour (no prefix) still substitutes with the
`emitExpr`-rendered Go source.

---

## DONE 055.3 — Function-sorted state facts + action Pre reifier

OPEN-pass.

Two pieces landed:

**(a) Array-storage state facts.** `solver_emit.go`
`emitStateSymbolFacts` now extends `stateFactsAsClauses` to walk
each array-storage state symbol's dimensions, emit nested loops,
and append per-cell `mkBoolFact("<sym>(<i0>,<i1>)", state.X[__i0][__i1])`
assertions for bool-valued cells. Hash-thunk symbols are handled
by the read-side thunk slot (OPEN 061.1) and are intentionally
skipped here. Non-bool cell ranges (integer/enum) are the next
sub-step.

**(b) Per-action precondition reifier.** `action_gen.go`
`emitPreconditionForAction` calls `goivy.GetUpdate(action, ctx)`
at emit time and walks `update.Pre.Fmlas` through
`reifyExprAsGoCode`, which produces a Go source string that
reconstructs the same Expr tree at runtime via
`goivy.NewConst` / `&goivy.LogicNot{}` / `&goivy.LogicAnd{}` /
`&goivy.LogicOr{}` / `&goivy.LogicImplies{}` / `&goivy.LogicIff{}` /
`&goivy.Eq{}`.

Formal-parameter references inside the Pre are rewritten to the
runtime input symbols `__in<i>`. goivy's three naming conventions
for the same param (`b`, `fml:b`, `__fml:b`) are all matched by
the reifier so the substitution is robust.

The emitted helper signature is
`buildPrecondition_<Name>(state *State, __in0 *goivy.Const, …) *goivy.Clauses`
which returns `conjClauses(stateFactsAsClauses(state), …reified Pre…)`.
Each actionGen's Generate now seeds `GetModelClauses` with this
helper's result, so the solver receives the action's actual reverse
image (when reifiable) plus the current state.

Unreifiable Pre shapes (sort kinds beyond Boolean/Uninterpreted,
Apply, ForAll, etc.) are silently skipped with a `// OPEN 055.3
unreifiable Pre fmla: <…>` comment so callers can see which fmlas
fell through. Extending the reifier's sort + Expr coverage is the
natural next refinement.

Verification: `TestEmit_TestTarget_ArrayStorageStateFacts`,
`TestEmit_TestTarget_BuildPreconditionHelperPerAction`,
`TestEmit_TestTarget_ReifiedPreReferencesInputSym`,
`TestReifyExprAsGoCode_*` in `action_gen_test.go`. Tier 2 smoke
(`TestSmoke_BuildEmittedTest`) still compiles the emitted test-
target package against in-tree goivy with the new helpers.

---

## OPEN — final remaining residual

### OPEN 055.4 — Reifier coverage for Apply / ForAll / non-Boolean sorts

`reifyExprAsGoCode` supports `Const / LogicNot / LogicAnd / LogicOr /
LogicImplies / LogicIff / Eq`. `reifySortAsGoCode` supports
`BooleanSort / UninterpretedSort`. Pre clauses that mention
`Apply`, `ForAll`, `LogicExists`, or sorts beyond Boolean and
uninterpreted (e.g. enum, range, BV-interp) fall through to a
`// OPEN 055.3 unreifiable Pre fmla: …` comment. The Solver round-
trip still works — the precondition just degrades to the state
facts in that case.

Extending the reifier is straightforward but routine: add cases
for each Expr/Sort type that translates structurally into the
goivy.New* constructor form. Likely needs `mustApply` / `mustNewVariable`
runtime helpers since `NewApply`/`NewVariable` can return errors.
