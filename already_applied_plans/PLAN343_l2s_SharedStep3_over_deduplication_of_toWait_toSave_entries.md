# Fix l2s SharedStep3 over-deduplication of toWait/toSave entries
Created: 2026-04-29 14:25:14 UTC

## Context

The 2hr golden test diverges at XTRACE 28248047. Go's `toWait` has 7 entries (indices 0-6) while Python has 8 (indices 0-7). They agree through index 6, then Go moves to `toSave[0]` while Python outputs `toWait[7]` (an `rfn.abs.done` formula).

Root cause: Go updates the `seen` set inside the l2s_g and bindings augmentation loops, deduplicating new entries against OTHER new entries. Python only checks new entries against the INITIAL set (never calls `seen.add(t)`). When two l2s_g binders produce the same negated body and that body wasn't in the initial collection, Python adds both while Go adds only the first.

## Changes

**File: `/Users/jaten/ivy/goivy/check/l2s_shared.go`**

### Fix 1: Remove `seenWait[key] = true` at line 267

The line `seenWait[key] = true` inside the l2s_g loop causes over-deduplication. Python (ivy_l2s.py:916-921) builds `seen` once and never updates it:

```python
seen = set(t for (vs,t) in named_binders_conjs['l2s_w'])
for b in ilu.named_binders_asts([ilu.normalize_named_binders(not_lf)]):
    if b.name == 'l2s_g':
        vs,t = b.variables,ilu.negate(b.body)
        if t not in seen:
            named_binders_conjs['l2s_w'].append((vs,t))
            # NO seen.add(t)
```

Delete line 267 (`seenWait[key] = true`) to match.

### Fix 2: Remove `seenSave[key] = true` at line 240

Same pattern for l2s_s augmentation. Python (ivy_l2s.py:908-915) builds `seen` once and never updates it. Delete line 240 (`seenSave[key] = true`) to match.

## Other sites examined -- no bugs

- `l2s.go:189` (`dedupeVarBodyPairs`): Correct; Python `dict.fromkeys` also fully dedupes.
- `l2s.go:1037` (`collectAllNamedBinders`): Correct; Python `set(v)` also fully dedupes.
- `l2s_auto.go:990`, `ranking.go:327`: Correct; Python `iu.unique` updates its memo set.
- `isolate_check.go:1304,1315`: Unrelated to l2s.

## Scope

Both bugs are inside the `if full {` block, which only runs for the `l2s_full` tactic. The ranking path (`full=false`) and `l2s_auto*` paths are unaffected.

## Verification

Run: `cd ~/ivy/goivy && make test` (specifically the Test2hrOrdLive test that produces the golden comparison).
