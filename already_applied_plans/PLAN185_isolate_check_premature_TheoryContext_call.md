# Plan: Fix premature TheoryContext call in Go CheckModule

**Created**: 2026-04-02 (replaces old ConjActions plan which is done)

## Context

`make golden` now diverges at line 157368. Go's `theory` field has content (definitions) at the `CheckFragment` CanonSnapshot, but Python's `theory` is not yet set at that point (Python skips to `actions.keys=24`).

The downstream effect: Go finds cycles in the theory (the `reportCycle` stack trace) while Python finds none — because Python doesn't have a theory yet when CheckFragment runs.

## Root Cause

**Go `CheckModule`** (`check/isolate_check.go:962-963`) wraps `CheckIsolate` in a premature `TheoryContext()` call:

```go
cleanup := isoMod.TheoryContext()   // line 962 — sets Theory!
err := CheckIsolate(isoMod, nil)    // line 963
cleanup()                            // line 964
```

**Python `check_module`** (`ivy_check.py:974`) does NOT wrap `check_isolate()` in a `theory_context()`:

```python
check_isolate()  # line 974 — no theory_context wrapper
```

Inside `check_isolate()` / `CheckIsolate()`, the flow is:
1. `check_fragment()` / `CheckFragment()` — runs first, uses CanonSnapshot
2. `with im.module.theory_context():` / `mod.TheoryContext()` — theory set HERE

So Python's theory is `None` during `check_fragment()` (correct), but Go's theory is already populated because `CheckModule` set it prematurely.

Additionally, `CheckIsolate` line 101 calls `TheoryContext()` again internally — so the `CheckModule` call is both premature AND redundant.

## Traced Call Paths

**Python** (`ivy_check.py`):
- `check_module()` line 974 → `check_isolate()` (no theory_context)
- `check_isolate()` line 525 → `ifc.check_fragment()` (theory=None)
- `check_isolate()` line 531 → `with im.module.theory_context():` (theory set here)

**Go** (`check/isolate_check.go`):
- `CheckModule()` line 962 → `isoMod.TheoryContext()` (theory set prematurely!)
- `CheckModule()` line 963 → `CheckIsolate(isoMod, nil)`
- `CheckIsolate()` line 94 → `fragment.CheckFragment(mod, false)` (theory already set — WRONG)
- `CheckIsolate()` line 101 → `mod.TheoryContext()` (redundant second call)

## Fix — 1 change

**File**: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/isolate_check.go`

Remove the `TheoryContext()` call at line 962 in `CheckModule`, since `CheckIsolate` already calls it internally at the correct point (line 101, after `CheckFragment`).

```go
// BEFORE (lines 961-964):
			// Set up theory context for the isolated module
			cleanup := isoMod.TheoryContext()
			err := CheckIsolate(isoMod, nil)
			cleanup()

// AFTER:
			err := CheckIsolate(isoMod, nil)
```

This matches Python's `check_module()` line 974 which simply calls `check_isolate()` without any `theory_context()` wrapper.

## Verification

```bash
cd ~/ivy/goivy && make test && make golden
```

The `theory` field in the CanonSnapshot at `fragment/fragment.go:922` should now be empty (matching Python), and the cycle check divergence should disappear.
