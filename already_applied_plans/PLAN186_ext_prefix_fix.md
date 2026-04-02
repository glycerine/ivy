# Plan: Fix "ext:" prefix divergence in makeBeforeExport mixin processing

**Created**: 2026-04-02 (updated)

## Context

`make golden` diverges at line 157393 in `beforeExport` field content. The `CanonSnapshot` at `fragment/fragment.go:922` shows Go produces `rep:"ext:rfn.abs.recv2"` but Python produces `rep:"rfn.abs.recv2"` in a callAction atom inside a `beforeExport` action.

The root cause: Go's `makeBeforeExport` unconditionally prefixes all call targets with "ext:", but Python conditionally prefixes only those whose names start with something in the `verified` set.

## Root Cause Analysis

### Python (`ivy_isolate.py`)

Two distinct prefix mechanisms:

1. **`prefix_call_ext`** (line 988-989) — CONDITIONAL:
   ```python
   def prefix_call_ext(name):
       return 'ext:'+name if startswith_some(name, verified, mod) else name
   ```

2. **`ext_mod_mixin(ea)`** (line 991-992) — wraps the above:
   ```python
   def ext_mod_mixin(ea):
       return lambda mixin,m: m if startswith_some(mixin.mixer(),verified,mod) and not ea(mixin)
              else m.prefix_calls(prefix_call_ext)
   ```

In `make_before_export` (line 1100):
```python
action1 = ext_mod_mixin(all_mixins)(mixin, action1)
```
Since `all_mixins` always returns True, `not ea(mixin)` is always False, so this always applies `m.prefix_calls(prefix_call_ext)` — the **conditional** prefixer. Call targets NOT in the `verified` set are left unchanged.

### Go (`isolate/isolate.go`)

Go already has the correct `prefixCallExt` (line 474) and `extModMixin` (line 482) defined. They are used correctly in the main classification loop (line 552).

**BUT** in `makeBeforeExport` (line 617):
```go
action1 = actions.PrefixCalls(action1, "ext:")  // BUG: unconditional
```

Should be:
```go
action1 = extModMixin(allMixins)(mx, action1)  // CORRECT: conditional via prefixCallExt
```

This matches what Python does: `ext_mod_mixin(all_mixins)(mixin, action1)`.

## Fix

**File**: `isolate/isolate.go`
**Line**: 617

### Change

```go
// BEFORE (line 617):
action1 = actions.PrefixCalls(action1, "ext:")

// AFTER:
action1 = extModMixin(allMixins)(mx, action1)
```

This is a one-line fix. The `extModMixin` and `allMixins` closures are already defined earlier in the same function (lines 482, 494) and are already used in the main action loop (line 552, 568). We just need to use them here too instead of the unconditional `PrefixCalls`.

## Verification

1. `make golden` — the divergence at line 157393 in `beforeExport` should be resolved
2. The `rep:"ext:rfn.abs.recv2"` should become `rep:"rfn.abs.recv2"` matching Python
3. Check that no other calls in `makeBeforeExport` need the same fix:
   - Line 599: `actions.PrefixCalls(act, "ext:")` — this is for the main action body (not mixin), matches Python line 1091 `action.assert_to_assume(...).prefix_calls('ext:')` which also unconditionally prefixes. **No change needed here.**

## Files to modify

- `isolate/isolate.go:617` — single line change
