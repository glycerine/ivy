# Fix ActionInterferenceCheck Divergence at xtrace line 146102+

**Created:** 2026-03-28 ~21:45 PST
**Updated:** 2026-03-28 ~23:30 PST (thorough investigation complete)

## Context

`make golden` diverges at xtrace line 146110 in `compiler.ActionInterferenceCheck`. After initial fixes (destructor unwrapping, deduplication, basic ActionOrder), the first 8 matching actions now align but the remaining actions diverge:

- **Go line 146110**: `ref.init[after93] modifies=2`
- **Python line 146110**: `isd.init[after79] modifies=3`

Full Python list (last 2): `isd.init[after79] modifies=3`, `ref.init[after93] modifies=12`
Full Go list (last 2): `ref.init[after93] modifies=2`, `ref.perform modifies=2`

## Already Applied Fixes (in current working tree)

1. **HavocAction destructor unwrapping** — `actions/transforms.go` now unwraps destructor chains for HavocAction, matching Python
2. **ActionsConfig passed to Modifies()** — `compiler/ivy_compile.go` now creates ActionsConfig with module context for proper destructor lookup
3. **Deduplication** — xtrace uses `map[lg.NodeKey]bool` set for unique symbol counting, matching Python's `set()`
4. **ActionOrder field** — `module/module.go` has `ActionOrder []string` and `SetAction()` method
5. **Most insertion sites updated** — compiler, isolate, bmc, mc, vmt files use `SetAction()`
6. **CrashAction handling** — `modifiesRec` has CrashAction case (not used in this test, but correct for future)

## Root Cause of Remaining Divergence

### Problem: `isolate/isolate.go` destroys ActionOrder via map iteration

The `IsolateComponent()` function (which runs before ActionInterferenceCheck, between ARGSetup.isolate ENTER/EXIT) iterates `mod.Actions` via `range` (random order) and rebuilds the actions map. This destroys Python's insertion-order semantics.

**Key offending code paths:**

1. **Line 482**: `for actname, act := range mod.Actions` — iterates actions in random order to decide verified/opaque/summarized. Builds `newActions` map.

2. **Line 942**: `for actname, act := range newActions` — erases unreferenced symbols in random order.

3. **Line 1111-1114**: Clears `mod.Actions` and rebuilds from `newActions` map iteration:
   ```go
   mod.Actions = make(map[string]module.Action)
   mod.ActionOrder = nil
   for name, act := range newActions {
       mod.SetAction(name, act)
   }
   ```
   This iterates `newActions` (a map) in random order, producing random ActionOrder.

**Python equivalent** iterates `mod.actions.items()` (insertion-order dict) at all these points, preserving original declaration order.

### Why counts also differ (2 vs 12, missing actions)

The CanonSnapshot does NOT capture actions — only axioms, props, inits, conjs, sig, schemata, initCond. So action trees could genuinely differ between Go and Python without being caught by prior xtrace matching.

The different modify counts (e.g., `ref.init[after93]` shows 2 in Go vs 12 in Python) and different action sets (Go has `ref.perform`, Python has `isd.init[after79]`) strongly suggest that **the action trees themselves differ**. This could be caused by:

1. **Order-dependent isolate processing** — if mixin application or cone-of-influence filtering depends on iteration order, different map orderings produce different action content
2. **A separate bug in action compilation/mixing** unrelated to ordering

Since we cannot distinguish these without fixing ordering first, the plan is:
1. Fix ordering
2. Re-run test
3. If counts still differ, investigate action tree content

## Fix Plan

### Step 1: Fix `isolate/isolate.go` — ordered iteration throughout IsolateComponent

**File:** `isolate/isolate.go`

The main loop at line 482 iterates `mod.Actions` to process each action. Replace with `mod.ActionOrder`:

```go
// Before: for actname, act := range mod.Actions {
// After:
for _, actname := range mod.ActionOrder {
    act, ok := mod.Actions[actname]
    if !ok {
        continue
    }
```

Maintain a `newActionOrder []string` alongside `newActions` to track insertion order. Every assignment `newActions[name] = ...` should append to `newActionOrder` (if not already present).

At line 1111-1114, iterate `newActionOrder` instead of `range newActions`:

```go
mod.Actions = make(map[string]module.Action)
mod.ActionOrder = nil
for _, name := range newActionOrder {
    if act, ok := newActions[name]; ok {
        mod.SetAction(name, act)
    }
}
```

### Step 2: Fix `isolate/strip.go` — preserve ActionOrder

**File:** `isolate/strip.go`

At line 709, `for name, act := range mod.Actions` iterates randomly to build `newActions`.
At line 770, `mod.Actions = newActions` replaces without updating ActionOrder.

Fix: iterate `mod.ActionOrder` at line 709. After replacing at line 770, rebuild ActionOrder.

### Step 3: Fix `isolate/isolate.go` line 942 — ordered erase

**File:** `isolate/isolate.go`

At line 942:
```go
for actname, act := range newActions {
    newActions[actname] = actions.EraseUnrefed(act, allSyms, allNames)
}
```

This iterates `newActions` randomly. Since `EraseUnrefed` is per-action (no cross-action state), order shouldn't matter for correctness, but use `newActionOrder` for consistency and future-proofing.

### Step 4: Re-run golden test

Run `make golden` to check if ordering fix resolves the divergence. If modify counts still differ, proceed to Step 5.

### Step 5: (Conditional) Investigate action tree differences

If counts still differ after ordering fix:
- Add xtrace in `ActionInterferenceCheck` to dump action types/structure for diverging actions
- Compare action trees between Go and Python to find structural differences
- Root-cause may be in mixin application, cone-of-influence, or compilation

## Files to Modify

1. `isolate/isolate.go` — ordered iteration in IsolateComponent (~3 locations)
2. `isolate/strip.go` — preserve ActionOrder during stripping
3. (Conditional) `compiler/ivy_compile.go` — additional debug xtrace

## Verification

1. `go build ./...` — must compile
2. `cd ~/goivy && make golden` — check if line 146110+ now matches
3. `go test ./isolate/... ./module/... ./compiler/...` — existing tests pass
