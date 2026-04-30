# Fix Tactic Closure Capturing Wrong Module

**Created:** 2026-04-30 ~10:00 UTC

## Context

XTRACE 28253287 in the Test2hrOrdLive golden test shows a divergence:
```
go : XTRACE: module.Copy ENTER actions=94 isolates=16
py : XTRACE: module.Copy ENTER actions=5 isolates=16
```

Go's module has 94 actions (the full compilation-level set) while Python's has 5 (the isolate-specific set from `create_isolate`). This happens inside `check_subgoals` during proof checking of the `rfn.abs.iso` isolate.

## Root Cause

In `check/check.go:1188-1194`, `RegisterTactics` registers the `mc` and `vmt` tactic closures capturing `mod` — the **original** module from `Start()`. This module has all 94 actions from compilation.

```go
func RegisterTactics(proofCfg *module.ProofConfig, mod *module.Module) {
    proofCfg.RegisterTactic("mc", func(pc ...) (...) {
        return MCTactic(pc, goals, p, mod)   // <-- captures original mod (94 actions)
    })
    proofCfg.RegisterTactic("vmt", func(pc ...) (...) {
        return VMTTactic(pc, goals, p, mod)  // <-- captures original mod (94 actions)
    })
```

Later, `CheckModule` processes each isolate by:
1. `isoMod := mod.Copy()` → `CreateIsolate(isolate, isoMod)` → isoMod has 5 actions
2. `CheckIsolate(isoMod, nil)` → creates `ProofChecker` with `isoMod`

When the proof checker invokes the `mc` tactic, the closure passes the original `mod` (94 actions) instead of `isoMod` (5 actions). `MCTactic` then calls `CheckSubgoals(goals, method, mod)` with the wrong module.

In Python, `mc_tactic` calls `check_subgoals()` which reads `im.module` — the global that was set to the create_isolate-modified copy by the `with im.module.copy():` context manager. So Python always uses the correct module.

The `ProofCheckerInterface` already exposes `GetModule() *Module` which returns the correct isoMod set during `NewProofChecker` creation.

## Fix

**File:** `/Users/jaten/ivy/goivy/check/check.go`, lines 1190 and 1193

Change:
```go
return MCTactic(pc, goals, p, mod)
```
to:
```go
return MCTactic(pc, goals, p, pc.GetModule())
```

And:
```go
return VMTTactic(pc, goals, p, mod)
```
to:
```go
return VMTTactic(pc, goals, p, pc.GetModule())
```

This matches the pattern already used by all other tactics (l2s, temporal, etc.) which get the module from the proof checker.

## Files

- **Edit:** `check/check.go` lines 1190, 1193 — two token changes
- **Reference (no changes):**
  - `module/proofapi.go` — defines `ProofCheckerInterface.GetModule()`
  - `proof/checker.go:130` — implements `GetModule()` returning `pc.Mod`
  - `check/isolate_check.go:68` — creates ProofChecker with the correct isoMod

## Verification

Run `cd ~/ivy/goivy && make test`. The Test2hrOrdLive golden test should pass XTRACE 28253287, with module.Copy reporting `actions=5` matching Python.
