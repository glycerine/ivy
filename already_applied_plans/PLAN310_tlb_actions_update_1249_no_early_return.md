# Fix: applyUpdateAxioms skips xtrace when numUpdates=0

Created: 2026-04-15 (plan session)

## Context

`make tlb` shows Go and Python diverge at trace line 094773. After `AssumeAction.action_update` exits, Python emits `actions.applyUpdateAxioms ENTER numUpdates=0 modNames=[]` but Go skips this trace entirely and emits the next `Sequence.int_update compose[0]` trace instead.

## Root Cause

In Go's `applyUpdateAxioms` (`actions/update.go:1249`), there is an early-return guard:

```go
if ctx.Domain == nil || len(ctx.Domain.Updates) == 0 {
    return update  // line 1251 — returns BEFORE the trace at line 1264
}
```

The xtrace at line 1258-1264 only fires after this guard. When there are 0 updates, Go never traces.

In Python (`ivy_actions.py:202-227`), the equivalent trace at line 209 is unconditional — it fires even when `len(domain.updates) == 0`, then the `for u in domain.updates` loop simply doesn't iterate.

## Fix

**File:** `actions/update.go` (single file change)

Move the xtrace block to **before** the early-return guard so it always fires, matching Python. The early return still happens afterward for the zero-update case.

Specifically, in `applyUpdateAxioms` (line 1249):

1. After the function signature (line 1249), add the xtrace block that computes `numUpdates` and `modNames` from `update.Modified`
2. Handle `ctx.Domain == nil` by using `numUpdates=0` (Python assumes domain is never None, but our nil guard is a safety net)
3. Keep the early return at line 1250-1252 as-is, but now it comes **after** the trace
4. Remove the now-duplicate trace block at lines 1258-1265

The resulting function opening:

```go
func applyUpdateAxioms(update *Update, action Action, ctx *UpdateContext) *Update {
    // Trace unconditionally, matching Python Action.int_update line 209
    if xtracer.Enabled {
        numUpdates := 0
        if ctx.Domain != nil {
            numUpdates = len(ctx.Domain.Updates)
        }
        show := make([]string, len(update.Modified))
        for i, s := range update.Modified {
            show[i] = fmt.Sprintf("'%v'", s.Name)
        }
        sort.Strings(show)
        xtracer.Trace("actions.applyUpdateAxioms ENTER numUpdates=%d modNames=[%v]", numUpdates, strings.Join(show, ", "))
    }
    if ctx.Domain == nil || len(ctx.Domain.Updates) == 0 {
        return update
    }

    modified := update.Modified
    tr := update.TR
    pre := update.Pre

    // (no trace here — moved above)
    for _, u := range ctx.Domain.Updates {
        ...
```

## Verification

```
make tlb
```

This runs `TestIvyTlbModel` which compares Go vs Python traces line-by-line. The test should now pass line 094773 (and ideally proceed further or pass entirely).
