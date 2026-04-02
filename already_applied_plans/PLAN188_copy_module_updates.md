# Plan: Fix missing `Updates` copy in module.Copy()

**Created**: 2026-04-02 18:30

## Context

`make golden` diverges at line 157400. Go has `updates=[]` (empty) but Python has `updates=[(DerivedUpdate defn:...)]` with multiple DerivedUpdate entries at the `CheckFragment` canon snapshot (fragment/fragment.go:922).

## Root Cause

`module.Copy()` in `module/module.go` never copies the `Updates` field. It copies similar slice fields like `Progress` (line 309), `Rely` (310), `MixOrd` (311), `Natives` (312), `NativeDefinitions` (313) — but `Updates` is simply missing.

- `Updates` is defined at line 48: `Updates []interface{}`
- `Clear()` resets it at line 238: `m.Updates = nil`
- `Copy()` (lines 291-449) never sets `c.Updates`

When `CheckIsolate` calls `mod.Copy()` to create the isolated module, the copy gets `Updates = nil`. The subsequent `CheckFragment` snapshot shows `updates=[]`.

In Python, there's no explicit copy — the module object is shared, so `updates` (populated during compilation at `ivy_compiler.py:1237,1343,1354,1399`) is always present.

## Fix

**File**: `module/module.go`

Add one line after line 313 (after `NativeDefinitions` copy), following the same pattern as neighboring fields:

```go
c.Updates = append([]interface{}{}, m.Updates...)
```

## Verification

`make golden` — the `updates=[]` divergence at line 157400 should be resolved.

## Files to modify

- `module/module.go` — add 1 line in `Copy()` after line 313
