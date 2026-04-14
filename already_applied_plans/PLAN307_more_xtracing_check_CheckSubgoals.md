# Plan: Debug & Fix CheckSubgoals branch mismatch at xtrace line 845771

Created: 2026-04-14 ~23:50 UTC

## Context

**Previous fix** (applied, working): Moved `TheoryContext()` inside `if method != nil` block. This fixed the divergence at line 845768 — `check.CheckIsolate ENTER` now matches. Lines 845765-845770 all match.

**Current divergence**: `make golden` → TestOrdLive → line 845771, inside `module.CanonSnapshot` at the start of `CheckFragment`:
```
Go : XTRACE: module.CanonSnapshot ... labeledAxioms=[...20 spec props + def217...])]
Py : XTRACE: module.CanonSnapshot ... labeledAxioms=[...20 spec props...)]
```

Go's `LabeledAxioms` has 31 items (20 spec properties + 11 definitions def217-def227), while Python's has only 20 items (20 spec properties).

## Investigation Summary

The non-XTRACE output (not compared by the golden test, but visible in log.red) reveals a MAJOR state difference between Go and Python's inner CheckIsolate:

**Go inner CheckIsolate** (after xtrace 845768):
- Axioms: 20 spec properties + 11 definitions (def217-def227, lines 1704-1716 of ord_live.ivy)
- Properties: EMPTY
- Conjectures: EMPTY

**Python inner CheckIsolate** (after xtrace 845768):
- Axioms: 20 spec properties only
- Properties: EMPTY
- Conjectures: Many l2s-generated invariants (invar228, l2s_needed_when_start, l2s_created, etc.)

This pattern strongly indicates **Go takes the NON-TEMPORAL branch** of `CheckSubgoals` while **Python takes the TEMPORAL branch** for the same subgoal:
- Python's temporal branch: sets `LabeledConjs = model.invars` (l2s invariants), adds goal premises (including definitions) to `LabeledAxioms`
- Go's non-temporal branch: sets `LabeledConjs = nil`, doesn't add premises

## Root Cause Hypothesis

The subgoal produced by the l2s tactic contains a `*ast.TemporalModels` conclusion. In `CheckSubgoals`:

```go
conc := proof.GoalConc(goal)
tm, ok := conc.(*ast.TemporalModels)
if tm != nil {
    // temporal branch
} else {
    // non-temporal branch
}
```

Python's equivalent:
```python
conc = ivy_proof.goal_conc(goal)
if isinstance(conc, ivy_ast.TemporalModels):
    # temporal branch
```

The l2s tactic's `SharedStep12_BuildGoal` creates `newConc := acfg.NewTemporalModels(tm.Model, lg.True)` and wraps it via `proof.CloneGoal`. If there are non-temporal premises, CloneGoal wraps the conclusion in a `SchemaBody`. `GoalConc` handles SchemaBody by returning `sb.Conc()` (last element = the TemporalModels).

**Something causes Go's type assertion `conc.(*ast.TemporalModels)` to fail.** Possible causes:
1. The subgoal's formula is not a TemporalModels after proof application
2. The subgoal's SchemaBody.Conc() doesn't return the TemporalModels
3. There's a pointer/interface wrapping issue
4. A post-l2s proof tactic transforms the conclusion away from TemporalModels

Since the source of the branch mismatch couldn't be fully traced through code reading alone, **the plan requires diagnostic tracing** as the first step.

## Phase 1: Add Diagnostic Tracing

### `check/isolate_check.go` — Add tracing at CheckSubgoals branch decision

Add xtrace output right before the temporal/non-temporal branch decision to capture:

```go
// At the start of the CheckSubgoals loop, after GoalConc:
conc := proof.GoalConc(goal)
if xtracer.Enabled {
    xtracer.Trace("check.CheckSubgoals goal[%d] goalConc type=%s formula type=%s",
        goalIdx, iu.TypeName(conc), iu.TypeName(goal.Formula))
}
tm, ok := conc.(*ast.TemporalModels)
if xtracer.Enabled {
    xtracer.Trace("check.CheckSubgoals goal[%d] isTemporalModels=%v tm=%v",
        goalIdx, ok, tm != nil)
}
```

Also add at the start of CheckSubgoals:
```go
if xtracer.Enabled {
    xtracer.Trace("check.CheckSubgoals ENTER nGoals=%d nLabeledAxioms=%d method=%v",
        len(goals), len(mod.LabeledAxioms), method != nil)
}
```

### Python equivalent (for parity)

Add matching traces in `ivy_check.py:check_subgoals` at the same decision points.

## Phase 2: Build and Run Test

1. `go build ./...` — verify compilation
2. `make golden` — run golden test with new traces
3. Examine `log.red` for the new trace lines to see:
   - What type does `GoalConc(goal)` return?
   - Is it a TemporalModels?
   - What's in `mod.LabeledAxioms`?

## Phase 3: Fix Based on Diagnosis

The trace output will reveal one of:
- **Type mismatch**: GoalConc returns something unexpected → fix the type handling
- **Wrong formula type**: The subgoal formula lost its TemporalModels → fix wherever it gets stripped
- **Multiple subgoals in wrong order**: Go processes subgoals in different order → fix ordering
- **Module state pollution**: mod.LabeledAxioms was modified before CheckSubgoals → fix the mutation

## Critical Files

- `check/isolate_check.go:680-924` — CheckSubgoals (both branches)
- `check/l2s.go:281-855` — l2sTacticInt (returns the subgoal)
- `check/l2s_shared.go:792-809` — SharedStep12_BuildGoal (creates the TemporalModels conclusion)
- `proof/goal.go:27-32` — GoalConc (branch decision)
- `proof/goal.go:125-136` — CloneGoal (creates SchemaBody)
- `ivy_check.py:769-846` — Python check_subgoals (source of truth)

## Verification

1. `go build ./...` compiles clean
2. `make golden` → new trace output reveals root cause
3. After fix: `make golden` advances past line 845771
4. `go test ./...` passes
