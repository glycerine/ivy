# Solo 5: Fix HandleTemporals to Conform to Python Logic

## Context

`HandleTemporals` (`compiler/ivy_compile.go:1528-1541`) labels each action with the isolate names in which it is present. It calls `isolate.GetIsolateMap()` and then sets `action.Labels` on each action in `mod.Actions`. The Python version (`ivy_compiler.py:2149-2170`) does the same but with subtle behavioral differences the Go port doesn't match.

Python source (active code, lines 2167-2169):
```python
imap = iso.get_isolate_map(mod, verified=True, present=True)
for actname, action in mod.actions.items():
    action.labels = imap[actname]
```
Key: Python's `imap` is a `defaultdict(list)`, so `imap[actname]` returns `[]` for missing keys.

Go source (lines 1531-1541):
```go
imap := isolate.GetIsolateMap(mod, true, true)
for actname, action := range mod.Actions {
    if labeler, ok := action.(interface{ SetLabels([]string) }); ok {
        if labels, ok := imap[actname]; ok {
            labeler.SetLabels(labels)
        }
    }
}
```

## Bugs Found

### Bug 1: Actions not in imap get nil Labels instead of empty slice (line 1536)
Python's `defaultdict(list)` returns `[]` for missing keys → every action gets a labels assignment. Go checks `if labels, ok := imap[actname]; ok` and **skips** actions not in imap, leaving `Labels` as nil. Downstream consumer `temporal.getLabels()` (`temporal/temporal.go:189-216`) returns nil for these, whereas Python would see `[]`. Nil vs empty-slice can change behavior in length checks or nil guards.

**Fix:** Remove the `ok` check; use `imap[actname]` directly (Go map returns nil/zero for missing keys, and `SetLabels(nil)` followed by the fix below normalizes to `[]`). Or explicitly: `labeler.SetLabels(imap[actname])`.

### Bug 2: Missing `GetLabels()` method on ActionBase
`temporal.getLabels()` (line 191) tries `act.(interface{ GetLabels() []string })` first, which always fails because `ActionBase` only has `SetLabels`. This forces the fallback through a fragile type switch that must enumerate every action type. Python just accesses `action.labels` directly.

**Fix:** Add `func (b *ActionBase) GetLabels() []string { return b.Labels }` to `actions/action.go`.

### Bug 3: Interface assertion silently skips non-Action values (line 1535)
Python sets `action.labels = ...` unconditionally on every value in `mod.actions`. Go uses `action.(interface{ SetLabels([]string) })` which silently skips anything that doesn't implement the interface. While `mod.Actions` should only contain proper Action types, this diverges from Python's behavior. If a non-Action value were present, Python would crash (AttributeError), making the bug visible. Go silently hides it.

**Fix:** Log a warning (via `pp()`) when an action doesn't implement SetLabels, to match Python's "crash loudly" behavior.

## Files to Modify

1. **`~/goivy/actions/action.go`** — Add `GetLabels()` to `ActionBase`
2. **`~/goivy/compiler/ivy_compile.go`** — Fix `HandleTemporals` imap lookup + add warning for non-labeler actions
3. **`~/goivy/compiler/handle_temporals_test.go`** — New test file

## Plan

### Step 1: Add `GetLabels()` to ActionBase (`actions/action.go:84`)

After the existing `SetLabels` method, add:
```go
func (b *ActionBase) GetLabels() []string { return b.Labels }
```

### Step 2: Fix `HandleTemporals` (`compiler/ivy_compile.go:1531-1541`)

Replace current body with:
```go
func HandleTemporals(mod *module.Module) {
    imap := isolate.GetIsolateMap(mod, true, true)
    for actname, action := range mod.Actions {
        if labeler, ok := action.(interface{ SetLabels([]string) }); ok {
            labeler.SetLabels(imap[actname])
        } else {
            pp("HandleTemporals: action %s does not support SetLabels", actname)
        }
    }
}
```

Changes:
- `imap[actname]` directly instead of `if labels, ok := imap[actname]; ok` — returns nil for missing keys, matching Python's defaultdict returning `[]`
- Added `else` branch with `pp()` warning for non-labeler actions (mirrors Python's implicit AttributeError)

### Step 3: Write tests (`compiler/handle_temporals_test.go`)

**Unit Tests:**

1. **TestHandleTemporals_EmptyActions** — No actions → no panic, no error.

2. **TestHandleTemporals_ActionGetsLabels** — One action, one isolate containing it → action.Labels == [isolateName].

3. **TestHandleTemporals_ActionNotInAnyIsolate** — One action, no isolates → action.Labels is set (nil/empty, not left uninitialized). Verifies Bug 1 fix.

4. **TestHandleTemporals_MultipleIsolates** — One action present in 2 isolates → action.Labels has both names.

5. **TestHandleTemporals_MultipleActions** — Multiple actions with different isolate memberships → each gets correct labels.

6. **TestHandleTemporals_GetLabelsWorks** — After HandleTemporals, verify `act.(interface{ GetLabels() []string }).GetLabels()` works (Bug 2 fix).

7. **TestHandleTemporals_NoIsolates** — Module with actions but empty Isolates map → all actions get nil/empty labels.

**Fuzz Test:**

8. **FuzzHandleTemporals** — Random combinations of 0-5 actions and 0-3 isolates with random membership. Verify: no panics; every action in mod.Actions that implements SetLabels has Labels set (not left as original nil from before the call); isolate names in Labels are valid isolate names.

### Step 4: Verify
Run `make test` to ensure no regressions.

## Existing utilities to reuse
- `module.New()` — `~/goivy/module/module.go:175`
- `actions.ActionBase` — `~/goivy/actions/action.go:56` — embeddable base with Labels field
- `actions.Sequence` (or any action type) — for creating test actions
- `isolate.GetIsolateMap()` — `~/goivy/isolate/iter.go:272`
- `isolate.IsolateDefInterface` — needed to make isolates visible to GetIsolateMap
- Test patterns from `~/goivy/compiler/attach_proofs_test.go` and `~/goivy/compiler/batch_f_test.go`

## Test fixture: IsolateDef for tests

`GetIsolateMap` casts isolate values to `isolate.IsolateDefInterface` (3 methods):
```go
type IsolateDefInterface interface {
    VerifiedNames() []string
    PresentNames() []string
    IsExtract() bool
}
```

`IterIsolate` walks `mod.Hierarchy` recursively from each verified/present name. For simple flat tests (no hierarchy), an isolate with `VerifiedNames() = ["actionName"]` and `present=true` will cause `imap["actionName"] = ["isolateName"]`.

Test helper struct:
```go
type testIsolateDef struct {
    verified []string
    present  []string
}
func (t *testIsolateDef) VerifiedNames() []string { return t.verified }
func (t *testIsolateDef) PresentNames() []string  { return t.present }
func (t *testIsolateDef) IsExtract() bool          { return false }
```

For test actions, use `actions.Sequence{}` (embeds ActionBase, has SetLabels via ActionBase).
