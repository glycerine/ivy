# PLAN: Integration test for Failures propagation through module copy chain

Created: 2026-04-17, 04:30

## Context

The `Failures` propagation fix (Sites 1-4) has been applied to `isolate_check.go`. The fix propagates checker failures from `isoMod.Cfg.Failures` and `withLocalMod.Cfg.Failures` back to the parent `mod.Cfg.Failures` after each `CheckIsolate(copy)` call. The user wants a proper integration test that exercises the actual checker-failure path through the module copy chain — not a placeholder that just pre-sets a variable.

## Test design

### `TestCheckModuleReportsFailures` — replace the current placeholder

Create a module with 3 false conjectures (`&lg.Or{}` = empty disjunction = False). Call `CheckModule(mod)`. This exercises the full propagation chain:

1. `CheckModule` → `isoMod := mod.Copy()` (isolate_check.go:1021)
2. `CreateIsolate("", isoMod)` — preserves conjectures on empty-isolate path
3. `CheckIsolate(isoMod)` → detects `checkedInvariants` from `LabeledConjs` (line 208-213)
4. → `CheckConjsInStateWithAG(isoMod, ag, state, 8, nil)` creates 3 `ConjChecker`s (line 1416-1418)
5. → `CheckFcsInStateWithAG` → solver finds each False conjecture SAT when inverted → calls `Fail()` 3 times
6. → `isoMod.Cfg.Failures` incremented to 3 inside the copy
7. → propagation at isolate_check.go ~line 1110: `mod.Cfg.Failures += isoMod.Cfg.Failures - failsBefore` → `mod.Cfg.Failures = 3`
8. → `CheckModule` returns error containing `"failed checks: 3"`

```go
func TestCheckModuleReportsFailures(t *testing.T) {
    mod := module.New()
    // 3 false conjectures: empty Or = False. ConjChecker inverts (dualizes)
    // each → True → trivially SAT → Fail(). This exercises the full
    // CheckModule → isoMod := mod.Copy() → CheckIsolate(isoMod) → checker
    // failure → propagation back to mod.Cfg.Failures chain.
    for i := 0; i < 3; i++ {
        lf := &ast.LabeledFormula{Formula: &lg.Or{}}
        mod.LabeledConjs = append(mod.LabeledConjs, lf)
    }

    err := CheckModule(mod)
    if err == nil {
        t.Fatal("expected error from CheckModule with false conjectures, got nil")
    }
    if !strings.Contains(err.Error(), "failed checks:") {
        t.Errorf("expected error containing 'failed checks:', got %q", err.Error())
    }
    if mod.Cfg.Failures != 3 {
        t.Errorf("expected mod.Cfg.Failures=3 (propagated from isoMod copy), got %d", mod.Cfg.Failures)
    }
}
```

### `TestCheckModuleNoFailuresReturnsNil` — keep existing

Empty module, no conjectures → no failures → nil error. Already exists, no changes.

### Remove placeholder tests

Remove `TestFailuresPropagateAcrossModuleCopy` and `TestFailuresPropagateNestedCopies` — these just tested arithmetic on pre-set variables, not actual code paths.

## Critical file

- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/check_test.go`

## Verification

`XTRACE_OFF=1 go test ./check/... -run "TestCheckModule" -v`
