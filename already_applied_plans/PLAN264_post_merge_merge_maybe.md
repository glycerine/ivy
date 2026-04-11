# Plan: Merge duplicate `L2s*` named-binder constructors in `check/`

Created: 2026-04-11 16:30 UTC

## Context

`check/` currently has seven pairs of duplicate L2S named-binder constructors with subtly different signatures, left over from the recent merge of the formerly-separate `l2s/` and `ranking/` packages into `check/`:

| Lowercase (in `l2s.go`, internal) | Uppercase (in `ranking.go`, exported) |
|---|---|
| `l2sW(vs, t, label string)` | `L2sW(vs, body, label string)` |
| `l2sS(vs, t, label string)` | `L2sS(vs, body, label string)` |
| `l2sG(vs, t, environ *string)` | `L2sG(vs, body, environ string)` |
| `oldL2sG(vs, t, environ *string)` | `OldL2sG(vs, body, environ string)` |
| `l2sInit(vs, t, label string)` | `L2sInit(vs, body, label string)` |
| `l2sWhen(name, vs, t, label string)` | `L2sWhen(name, vs, body, label string)` |
| `l2sOld(vs, t, label string)` | `L2sOld(vs, body, label string)` |

Both are alive: lowercase versions are used heavily inside `l2s.go`, `l2s_shared.go`, `l2s_auto.go`; uppercase versions are used in 10 production sites in `ranking.go` + `ranking_tactic.go`, plus 7 wrapper-test functions in `ranking_test.go`. No external package imports any of them.

The label-based pairs are functionally identical (same `strPtr(label)` plumbing, only the body construction style differs: direct struct literal vs `lg.NewNamedBinder` factory). The environ-based pairs (`l2sG`/`L2sG`, `oldL2sG`/`OldL2sG`) have a **semantic difference**: the lowercase versions take `*string` and pass the pointer through unchanged; the uppercase versions take `string` and create a fresh `strPtr(environ)` per call.

The lowercase `*string` form is canonical:

1. **Python source faithfulness.** `~/ivy/pyivy/ivy/ivy/ivy_l2s.py:186-187` defines `l2s_g = lambda vs, t, environ: lg.NamedBinder('l2s_g', vs, environ, t)` — Python passes `Optional[str]` through unchanged. The lowercase Go version mirrors this; the uppercase version coerces `nil`→`&""` which is a real semantic change.
2. **Callback signature requirement.** `lu.GloballyBinderFunc` at `logicutil/logic_utils.go:943` is `func(vars, body, environ *string) *NamedBinder`. The `_l2sG` closure in `SharedStep1_ConvertTemporals` (`l2s_shared.go:80`) must satisfy this signature, so the lowercase `l2sG` cannot go away.
3. **In-tree precedent.** `l2s_shared.go:322, 334` already does `env := gprop.Environ; ...; l2sG(vs, t, env)` — passing the pointer through, exactly what we want. The uppercase ranking.go path (`ranking.go:625-628, 645-648`) extracts the value just to re-pointer-wrap it, which is the buggy outlier.
4. **Existing dedup depends on it.** `l2sGTriple.key()` at `l2s.go:165-167` uses `%v` formatting on the `*string` field, so its hash key is sensitive to pointer identity. Anything that re-wraps via `strPtr` would defeat this dedup if plumbed into `cfg.L2sGs`. (The current uppercase callers don't feed into that map, so no live bug — but the lowercase semantics are the ones the existing infrastructure already assumes.)

This plan removes the seven uppercase wrappers and routes their callers through the lowercase versions, fixing the `nil`-coercion semantics in `NewPropEvents` as a side effect.

## Approach

Five sequential edits, ordered so every intermediate state is buildable.

### Step 1 — Update production callers in `check/ranking.go`

Six call sites change:

**Lines 619-674 (`NewPropEvents`).** Currently extracts `gprop.Environ` to a local string and re-wraps. Drop the extraction blocks at lines 625-628 and 645-648, then pass the pointer through:

```go
// before
environ := ""
if gprop.Environ != nil {
    environ = *gprop.Environ
}
oldG := OldL2sG(vs, body, environ)
curG := L2sG(vs, body, environ)

// after
oldG := oldL2sG(vs, body, gprop.Environ)
curG := l2sG(vs, body, gprop.Environ)
```

Apply this to both loops (loop 1 at lines 622-639, loop 2 at lines 642-671).

**Line 688 (`WaitEvent`).** `proofLabel` is a `string`, so wrap with `strPtr`:
```go
gNegBody := l2sG(vs, negBody, strPtr(proofLabel))
```

**Line 801 (`applyWas`).** Lowercase the call:
```go
nb := l2sS(nil, expr, proofLabel)
```

**Line 808 (`applyHappened`).** Lowercase the call:
```go
nb := l2sW(nil, expr, proofLabel)
```

**Line 928 (`ConvertToInit`).** Lowercase the call:
```go
return l2sInit(nil, fmla, proofLabel)
```

After this step, build is green, the seven uppercase wrappers in `ranking.go` are dead but still defined.

### Step 2 — Update production callers in `check/ranking_tactic.go`

Two call sites:

**Line 252.** `L2sW(nil, eqRHS(workStart), "")` → `l2sW(nil, eqRHS(workStart), "")`

**Line 274.** `L2sW(progressArgs, eqRHS(workProgress), "")` → `l2sW(progressArgs, eqRHS(workProgress), "")`

(Both pass `""` for the label; the lowercase signature also takes `string`, so no `strPtr` wrap needed.)

After this step, build is green and the uppercase wrappers have zero production callers.

### Step 3 — Delete the seven wrapper-tests in `check/ranking_test.go`

Delete `TestL2sW`, `TestL2sG`, `TestOldL2sG`, `TestL2sInit`, `TestL2sWhen`, `TestL2sOld`, `TestL2sS` (lines 218-290). They are trivial constructor sanity tests for one-line wrappers we are about to remove. The lowercase constructors they would otherwise test are exercised transitively by the integration tests in the same file and by every L2S tactic run.

Keep `TestL2sD` (line 207) — `L2sD` is a `*lg.Const`-returning function, not part of this refactor, and the test does meaningful work building a `RelationSort`.

After this step, build and tests are green.

### Step 4 — Delete the seven uppercase wrappers from `check/ranking.go`

Delete lines 126-166 (the `L2sW`, `L2sG`, `OldL2sG`, `L2sInit`, `L2sWhen`, `L2sOld`, `L2sS` definitions). The intervening doc comments go too. The block ends at line 166; the next line (`168 // --- Task and Trigger types ---`) becomes the new top of the section.

After this step, build and tests are green and the duplication is gone.

### Step 5 — Verification

```sh
cd ~/ivy/goivy && go build ./...
cd ~/ivy/goivy && go vet ./check/...
cd ~/ivy/goivy && go test ./check/... ./compiler/... ./proof/... ./module/... ./tactics/... ./temporal/...
cd ~/ivy/goivy && go build ./cmd/goivy_check/...
```

Stale-symbol grep (must return only `L2sD` matches, no others):
```sh
grep -rn '\bL2sW\b\|\bL2sG\b\|\bOldL2sG\b\|\bL2sS\b\|\bL2sInit\b\|\bL2sWhen\b\|\bL2sOld\b' ~/ivy/goivy/check ~/ivy/goivy/cmd
```

External-import grep (must return nothing):
```sh
grep -rn 'check\.L2sW\|check\.L2sG\|check\.OldL2sG\|check\.L2sS\|check\.L2sInit\|check\.L2sWhen\|check\.L2sOld' ~/ivy/goivy
```

## Files to modify

| File | Change | Step |
|---|---|---|
| `check/ranking.go` | Update 6 call sites (lines 625-651, 688, 801, 808, 928); delete 7 wrapper functions (lines 126-166) | 1, 4 |
| `check/ranking_tactic.go` | Update 2 call sites (lines 252, 274) | 2 |
| `check/ranking_test.go` | Delete 7 wrapper-test functions (lines 218-290) | 3 |

No changes needed in `l2s.go`, `l2s_shared.go`, `l2s_auto.go`, or any other file. The lowercase constructors stay exactly as they are.

## Behavior change to note in commit message

The pointer-identity preservation in `NewPropEvents` is a **bug fix**, not just a refactor: previously, when `gprop.Environ == nil`, the code would coerce the value through `environ := ""` and then wrap with `strPtr(environ)`, producing a non-nil `*string` pointing at `""`. After this change, `gprop.Environ == nil` produces `Environ: nil` on the resulting `NamedBinder`, which is what Python's `Optional[str]` semantics dictate and what `l2s_shared.go` already does on the parallel code path. Any downstream consumer that distinguishes `env == nil` from `*env == ""` will now see the nil correctly. No current consumer in the tree performs that distinction (verified by reading every `.Environ` access in `check/`, `ivylogic/`, `module/`, and `temporal/`), so this is a latent-correctness fix with no observable user-facing effect today.

## Critical files

- `/Users/jaten/ivy/goivy/check/ranking.go` — primary edit site (caller updates + wrapper deletions)
- `/Users/jaten/ivy/goivy/check/ranking_tactic.go` — two `L2sW` call sites
- `/Users/jaten/ivy/goivy/check/ranking_test.go` — seven wrapper-tests to delete
- `/Users/jaten/ivy/goivy/check/l2s.go:67-100` — the canonical lowercase constructors (read-only reference; not modified)
- `/Users/jaten/ivy/goivy/check/l2s_shared.go:80-105` — the `_l2sG` closure that requires `*string` (read-only reference; explains why the lowercase signature is load-bearing)
- `/Users/jaten/ivy/goivy/logicutil/logic_utils.go:942-943` — the `GloballyBinderFunc` callback type (read-only reference)

## Out of scope (followups, not done here)

- **`rank`-prefixed helpers in `ranking.go` / `ranking_tactic.go`** (`rankApplyNB`, `rankVarsToNodes`, `rankingMakeAnd`, possibly others). They look like potential duplicates of `applyNB`/`varsToNodes`/`makeAnd` in `l2s.go`. Audit and possibly merge in a separate PR.
- **`WaitEvent`/`applyWas`/`applyHappened`/`ConvertToInit` signatures** still take `proofLabel string` and unconditionally wrap with `strPtr`. To be fully Python-faithful these should take `*string` and propagate `nil` through. Out of scope: cascading signature change would require auditing every caller of those four functions, and the current behavior is bug-compatible with the existing uppercase code.
- **`l2sGTriple.key()` `%v`-on-pointer fragility** at `l2s.go:165-167`. Two semantically-identical triples produced from different pointer origins won't currently collide in `cfg.L2sGs`. Not introduced by this plan; flagged for a separate audit.
- **`NewPropEvents` loop merge.** The function iterates `gprops` twice, constructing `oldG`/`curG` four times per property. Combinable into one loop, but ordering of pre-actions vs post-actions in the output may be load-bearing — needs separate scrutiny.
- **`L2sD`** at `ranking.go:118`. Not a duplicate (it returns `*lg.Const`, not `*lg.NamedBinder`). Stays as-is.
- **`L2sGTriple` type alias** at `l2s_shared.go:66` (`type L2sGTriple = l2sGTriple`). Required because `InstrumentationConfig` is exported and has fields of this type. Correct as-is.

## Risks and rollback

- **Risk: hidden external caller.** Mitigated by the grep in Step 5. Verified pre-plan: no `check.L2sW` etc. anywhere in the repo (excluding `already_applied_plans/` per project rules).
- **Risk: pointer-identity change breaks dedup somewhere.** Audited every reader of `.Environ` in the relevant packages. None performs pointer equality. The only reader that depends on pointer identity at all is `l2sGTriple.key()`, and the new behavior strictly improves that hashing (same input pointer → same key, where the old uppercase code would have produced different keys for equivalent inputs). No live regression.
- **Risk: tests we delete were catching something.** The seven wrapper-tests check only `nb != nil` and `nb.Name == "..."`. The first is true by construction (the wrappers can't return nil because the underlying lowercase constructors can't either). The second is true by string-literal inspection. They provide no meaningful coverage that integration tests don't already.
- **Rollback**: each step is independently revertible. Worst case is reverting the deletion in Step 4 (paste back ~40 lines from the plan or git history) and undoing the caller updates in Steps 1-2.
