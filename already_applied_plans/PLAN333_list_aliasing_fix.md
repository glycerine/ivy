# Plan: fix l2s SharedStep1 nAsms divergence (l2s_auto2 on ord_live.ivy)

Created: 2026-04-19 21:45 PT

## Context

`TestOrdLive2hr` in `~/ivy/goivy/parser/golden_test.go` runs `ivy_check` and
`goivy_check` against `ivy-lang-examples/doc/examples/apple/ord_live.ivy` and
diffs their xtrace streams. All trace events up through index 2405652 match
exactly. At index 2405953 the two streams diverge for the first time:

```
2405953  go : XTRACE: l2s.SharedStep1 ENTER nInvars=36 nAsms=113 nBindings=24 nPrems=25
        py : XTRACE: l2s.SharedStep1 ENTER nInvars=36 nAsms=149 nBindings=24 nPrems=25
```

The trace fires at `pyivy/ivy/ivy/ivy_l2s.py:840` and `goivy/check/l2s.go:607`
(the start of `SharedStep1_ConvertTemporals` inside `l2s_tactic_int`). Only
`nAsms` differs. Python has **36 more** assumptions — exactly equal to
`nInvars=36`. The test is using `l2s_auto2` on the `cf_pio_live.if_liveness`
temporal property (line 2566 in `ord_live.ivy`).

Two prior `l2s.SharedStep1 ENTER` events must have matched (they don't appear
in the diff); only the third diverges. Either:
- Python accumulates state that Go discards, and the first two props happen to
  not differ in count — e.g. because the previous calls went through a path
  that didn't mutate `im.module.assumed_invariants` — or
- The specific tactic chain for `if_liveness` (`tempcase → skolemizenp →
  instantiate → l2s_auto2`) exposes a flow Go mirrors incorrectly.

## Primary hypothesis: broken slice aliasing in `NormalProgramClone`

Python's `NormalProgram.clone` at `pyivy/ivy/ivy/ivy_temporal.py:179-183` passes
`self.asms` (and `self.invars`, `self.bindings`, `self.calls`) **by
reference** to the new `NormalProgram`. `normal_program_from_module` at
lines 217-223 also assigns `asms = mod.assumed_invariants` by reference. This
means the NormalProgram that enters `l2s_tactic_int`, its `.clone([])` copy,
and `im.module.assumed_invariants` **all point to the same list object**.
When `ivy_l2s.py:159` does `model.asms.append(cloned)` for each assumed
gprop, `im.module.assumed_invariants` is mutated in place. When
`ivy_check.py:761` later does
`im.module.assumed_invariants.extend(im.module.labeled_conjs)` inside
`check_subgoals`' recursive `check_isolate`, the shared list grows again,
persisting past the `with mod:` exit.

Go's `NormalProgramClone` at `goivy/temporal/temporal.go:528-552` does:

```go
asms := make([]*ast.LabeledFormula, len(np.Asms))
copy(asms, np.Asms)
```

for all four slice fields — creating a **new backing array**. This severs the
aliasing chain. Appends to `model.Asms` never reach `mod.AssumedInvs`, so
across successive `l2s_tactic` invocations Go doesn't accumulate what Python
does. `TestNormalProgramClone` at `temporal_test.go:347-364` actively asserts
the (incorrect) non-aliasing behavior.

## Alternate hypotheses to rule out first

1. **Line 761 / `isolate_check.go:674` runs against a different-sized
   `LabeledConjs`/`labeled_conjs` because of earlier divergence in the
   module-level state.** Verify by tracing `len(mod.AssumedInvs)` and
   `len(mod.LabeledConjs)` on entry to the move at line 674 vs
   `ivy_check.py:761`.
2. **`preprocess_assumed_ignored_properties` path divergence.** Python at
   `ivy_check.py:499` extends `assumed_invariants` from `labeled_props +
   labeled_conjs` when ACL is on. Go at `check/check.go:1042-1050` mirrors
   this. This is only active with `opt_unchecked_properties` — confirm the
   test doesn't trigger it on either side.
3. **`check_subgoals` → `mod.labeled_conjs = model.invars` /
   `mod.assumed_invariants = model.asms` reference assignment.** Python at
   `ivy_check.py:789,795`; Go at `isolate_check.go:731,748`. Go's
   `withLocalMod.AssumedInvs = np.Asms` shares the (already-copied)
   `np.Asms` slice. If `NormalProgramClone` is fixed, this becomes correct
   by construction.

## Phase A — diagnostics before fix

Add xtrace events in matching Go/Python locations to pinpoint where the
36-item gap opens. Keep the label format identical so the golden diff stays
aligned. Additions per CLAUDE.md rule "Accrete xtraces, never delete".

1. `check_temporals` entry, per property:
   - Trace `check.CheckTemporals prop start label=<...> nAssumedInvs=%d nLabeledConjs=%d nLabeledProps=%d nLabeledAxioms=%d`.
   - Python: `ivy_check.py` around line 138–149.
   - Go: `check/check.go` around line 308.
2. `normal_program_from_module` exit:
   - Trace `temporal.NormalProgramFromModule EXIT nInvars=%d nAsms=%d nBindings=%d`.
   - Python: `ivy_temporal.py:223`.
   - Go: `temporal/temporal.go:396`.
3. `check_subgoals` TemporalModels branch, right before entering `with mod:`:
   - Trace `check.CheckSubgoals TemporalModels applied mod.AssumedInvs nAsms=%d nLabeledConjs=%d nInvars_from_model=%d`.
   - Python: `ivy_check.py` around line 795.
   - Go: `check/isolate_check.go` around line 748.
4. `check_isolate` just before the "move conjs to asms" step:
   - Trace `check.CheckIsolate preMove nAssumedInvs=%d nLabeledConjs=%d`.
   - Python: `ivy_check.py:761`.
   - Go: `check/isolate_check.go:673`.
5. `l2s_tactic_int` entry, immediately after `model = conc.model.clone([])`:
   - Trace `l2s.l2sTacticInt postClone nAsms=%d nInvars=%d`.
   - Python: `ivy_l2s.py:137`.
   - Go: `check/l2s.go:329`.

Run `cd ~/ivy/goivy && make test` (not `go test`, per CLAUDE.md). Inspect the
new golden diff: the first location where the two streams show different
counts is the point where Go mishandles aliasing.

## Phase B — fix

Given the expected finding (Phase A will confirm), the fix is:

### B1. Mirror Python's no-copy clone

**File: `goivy/temporal/temporal.go`** — replace `NormalProgramClone`
(lines 528–552) with a literal mirror of Python's `clone`: share all slice
and map references.

```go
func NormalProgramClone(np *NormalProgram) *NormalProgram {
    result := &NormalProgram{
        Base:     np.Base,
        Bindings: np.Bindings,
        Init:     np.Init,
        Invars:   np.Invars,
        Asms:     np.Asms,
        Calls:    np.Calls,
    }
    if np.Postconds != nil {
        result.Postconds = np.Postconds
    }
    return result
}
```

This restores the aliasing chain so appends to `model.Asms` in
`l2s_tactic_int` flow through to `im.module.AssumedInvs`.

### B2. Update the existing test to match Python semantics

**File: `goivy/temporal/temporal_test.go:347-364`** — `TestNormalProgramClone`
asserts the opposite of Python's behavior (that mutating `cloned.Calls` does
not affect `np.Calls`). Rewrite it to document and verify Python's aliasing:

- Appending to `cloned.Asms` MUST also grow `np.Asms` (same list object).
- Same for `Invars`, `Bindings`, `Calls`.

### B3. Audit other slice-copying clone paths

Search for any other `NormalProgram`-related clone that creates fresh slices
(e.g. the `(np *NormalProgram).Clone(args)` method at `temporal.go:243-244`
delegates to `NormalProgramClone`, so B1 covers it). Also audit:

- `goivy/temporal/temporal.go:589,591` (callers of `NormalProgramClone`).
- `goivy/check/l2s.go:328` (`model := temporal.NormalProgramClone(np)`).
- `goivy/check/isolate_check.go:748` (`withLocalMod.AssumedInvs = np.Asms`).
  If B1 is applied, `np.Asms` is now shared with the upstream module, so
  this assignment becomes semantically identical to Python line 795.

### B4. Verify no regression from earlier "don't skip clones" rule

Per `MEMORY.md`'s `feedback_no_changed_optimization.md`, Go must never skip a
clone that Python performs. This fix does the opposite direction — it makes
Go's clone a shallower alias, matching Python exactly. It does not skip any
clone that Python executes. Confirm by re-reading the feedback note.

## Verification

1. `cd ~/ivy/goivy && make test` — full test suite under XTRACE.
2. Inspect `~/ivy/goivy/log.golden.2hr` (or re-run `TestOrdLive2hr` via
   `cd parser && go test -v -timeout=0 -count=1 -run TestOrdLive2hr`) and
   confirm:
   - The new `SharedStep1 ENTER` at i≈2405953 shows `nAsms=149` on both
     Go and Python.
   - No new divergence appears downstream (i.e., the first divergence has
     moved past i=2405953, ideally to end-of-test).
3. Golden-compare `Canon()` output of the first post-`SharedStep1` asms on
   both sides to confirm not just count but identity of the 36 extra
   entries (they should be the invars added in an earlier iteration's
   check_subgoals recursive check_isolate).
4. Confirm `TestNormalProgramClone` still passes with the updated assertion.

## Critical files

- `/Users/jaten/ivy/goivy/temporal/temporal.go:528-552` — `NormalProgramClone` (primary fix).
- `/Users/jaten/ivy/goivy/temporal/temporal_test.go:347-364` — test update.
- `/Users/jaten/ivy/goivy/check/l2s.go:328,607` — diagnostics & caller.
- `/Users/jaten/ivy/goivy/check/check.go:308,1042-1050` — diagnostics.
- `/Users/jaten/ivy/goivy/check/isolate_check.go:673-684,748` — diagnostics & caller.
- `/Users/jaten/ivy/goivy/temporal/temporal.go:352-397` — `NormalProgramFromModule`.
- Python reference: `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_temporal.py:171-223`,
  `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_l2s.py:124-166,759-840`,
  `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_check.py:128-165,755-854`.
