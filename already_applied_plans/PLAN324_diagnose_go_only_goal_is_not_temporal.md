# PLAN: Fix Go l2s "proof goal is not temporal" divergence on cf_pio_live.cf_liveness

Created: 2026-04-17, 21:30

## Context

Running `goivy_check ord_live.ivy` fails at temporal property `cf_pio_live.cf_liveness` (line 2521) with:
```
error: check/l2s: [2]proof goal is not temporal
```

Python `ivy_check` succeeds on the same property, producing `--- begin l2s_auto invariants ---`.

The property's proof is a `ComposeTactics` with three steps:
```ivy
proof {
    tactic skolemize;
    instantiate cfabric_pio_fair_ax;
    tactic l2s_auto2 with
        definition work_start = ...
        ...
        invariant ~cfabric.rd_pio_fair
        invariant ~cfabric.wr_pio_fair
}
```

## Analysis

The error fires at `check/l2s.go:316` when `proof.GoalConc(goal)` returns something that is NOT `*ast.TemporalModels`. The goal reaches l2s through the compose chain:

1. **`tactic skolemize`** — Skolemizes `forall T`. Adds `ConstantDecl(_T)` premise, wrapping the goal in a SchemaBody. `SkolemizeFmla` correctly handles TemporalModels (line 111: recurses into inner formula, re-wraps). Result should be `SchemaBody{[ConstantDecl(_T), TemporalModels(model, skolemized_fmla)]}`.

2. **`instantiate cfabric_pio_fair_ax`** — Parsed as `AssumeTactic` (Go grammar line 4970), dispatched to `proof/tactics.go:77 assumeTactic`. Adds the axiom as a premise. Result should be `SchemaBody{[ConstantDecl, axiom_LF, TemporalModels(...)]}`.

3. **`tactic l2s_auto2 with ...`** — Dispatched to `L2STacticAuto` → `l2sTacticInt`. Calls `GoalConc(goal)` → `sb.Conc()` → should return TemporalModels (last element).

**Static analysis shows every step should preserve TemporalModels as the conclusion.** But the error fires anyway. Without runtime tracing, I cannot determine the exact type that `conc` has at line 314.

### Key files traced

- `check/l2s.go:287-316` — l2sTacticInt, the error site
- `check/check.go:275-362` — CheckTemporals, creates TemporalModels-wrapped subgoal
- `proof/checker.go:458-483` — AdmitProposition, dispatches to ApplyProof
- `proof/checker.go:518-534` — composeProofs, iterates proof steps
- `proof/skolem.go:17-90` — SkolemizeGoal, handles TemporalModels at line 111
- `proof/tactics.go:77-127` — assumeTactic, adds premise via goalAddPrem
- `proof/goal.go:27-32` — GoalConc, unwraps SchemaBody
- `proof/goal.go:125-136` — CloneGoal, wraps in SchemaBody when prems present
- `ast/ast.go:1670-1691` — TemporalModels struct, Args(), Clone()

## Diagnostic step (Phase 1)

Run with xtracing enabled to see the actual conclusion type:

```bash
cd ~/ivy/goivy && time ./goivy_check ord_live.ivy 2>&1 | tee go.ord.log
```

The existing xtracer at `l2s.go:314` will print:
```
l2s.l2sTacticInt goalConc result type=<TYPE> (isTemporalModels=<BOOL>)
```

This reveals what type `conc` actually is. Additionally, the xtracer at `l2s.go:294` will print:
```
l2s.l2sTacticInt ENTER tactic='<NAME>' ngoals=<N> goal.Formula type=<TYPE>
```

If `goal.Formula type` is `SchemaBody`, then `GoalConc` unwraps it. If it's something else (e.g., a bare `ForAll` or `Globally`), the TemporalModels wrapper was lost upstream.

The compose proof steps also have tracing:
- `proof.composeProofs step=<I>/<N> proofType=<TYPE> goal[0].Formula type=<TYPE>`
- `proof.ApplyProof ENTER proofType=<TYPE> goal[0].Formula type=<TYPE>`

These traces will pinpoint exactly which step loses the TemporalModels.

## Fix (Phase 2 — after diagnostic)

Once we know which step strips TemporalModels, the fix will be in one of:

1. **If `skolemize` strips it**: Fix `SkolemizeFmla` or `SkolemizeGoal` to preserve the wrapper.
2. **If `instantiate`/`assumeTactic` strips it**: Fix `assumeTactic` or `goalAddPrem`.
3. **If the proof is mis-parsed**: Fix the grammar to produce the correct AST for the compose proof.
4. **If `GoalConc` misbehaves on nested SchemaBody**: Fix `SchemaBody.Conc()`.

Most likely candidates based on analysis:
- The `assumeTactic` in Go is simplified compared to Python (skips `setup_matching`, `compile_match`, `fo_match`, `witness_ast`, `close_goals`, `apply_match_goal`). For `instantiate cfabric_pio_fair_ax` with no renaming, the simplified version SHOULD work since matching is identity. But if it doesn't find the schema, it errors — which would prevent l2s from being reached.
- A possible issue with how `CloneGoal`/`MakeGoal` constructs the SchemaBody — need to verify the element ordering at runtime.

## Files to modify (TBD after diagnostic)

Likely one of:
- `proof/tactics.go` — assumeTactic
- `proof/skolem.go` — SkolemizeGoal/SkolemizeFmla
- `proof/goal.go` — CloneGoal/MakeGoal/GoalConc

## Verification

```bash
cd ~/ivy/goivy
# After fix, run without XTRACE_OFF to confirm the trace shows TemporalModels:
time ./goivy_check ord_live.ivy 2>&1 | grep "l2s.l2sTacticInt goalConc"
# Should show: type=TemporalModels (isTemporalModels=True)

# Then run with XTRACE_OFF for clean output:
XTRACE_OFF=1 time ./goivy_check ord_live.ivy
# Should proceed past cf_pio_live.cf_liveness without error
```
