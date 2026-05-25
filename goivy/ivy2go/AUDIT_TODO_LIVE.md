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

## OPEN — known residual gaps

### OPEN 055.2 — Real reverse-image precondition derivation

The Solver round-trip in `action_gen.go` Generate uses a trivial
`true` Clauses today. The full reverse-image derivation requires
running goivy's action-update analysis (`actions_transrel.go`'s
`PureStateClauses` / `StatePrecond`) to derive the action's
precondition Clauses. This is a project on its own — the Solver
wiring is in place.

### OPEN 061.1 — Read-side wiring for thunk-wrapped LHS

`assign.go` `emitAssignLarge` builds a thunk but reads of the LHS
still hit the map storage. To make this fully lazy, each LHS read
(in `expr.go` `goStorageAccess`) needs to fall back to the thunk's
`get(k)` on map miss. Requires threading thunk pointers through
the State struct.

### OPEN 062.1 — Destructor hash on hash-thunk fields

`emitDestructorHash` aggregates map fields by `len(map)` rather than
hashing each entry (avoids non-deterministic Go map iteration
order). For records used as map keys this is acceptable but lossy;
a deterministic sort + per-entry hash would tighten the hash.

### OPEN 064.1 — Type/Z3-name antiquote prefixes

`renderNativeGoTemplate` substitutes raw expression text but not
the C++ prefix flavours (`%`-type, `"`-Z3-name) ivy2cpp's
`renderNativeTemplate` supports. Go has no direct equivalents for
those (the type-name flavour could lower to `g.goType(...)`); add
them if a Go-targeting fixture needs them.
