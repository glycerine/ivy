# Fix Structural Differences in GetCallsModsRecFull and Related Functions

**Created:** 2026-04-01 21:30

## Context

The golden trace matches through line 156174 (interfSyms trace). At line 156175, Go's `GetCallsModsRecFull`/`GetLocMods` traces diverge from Python's `get_calls_mods`/`get_loc_mods`. Rather than just fixing the trace ordering, we need to fix the structural code differences so Go's execution path exactly conforms to Python's behavior.

Current divergence:
```
156175  go : XTRACE: isolate.GetLocMods ENTER actname=rfn.abs.init[after237]
        py : XTRACE: isolate.GetCallsModsRecFull mod actname=tar_clock.impl.prev[implement16] sym=fml:y type=AssignAction
```

## Structural Differences (Go vs Python)

### Difference 1: `GetCallsModsRecFull` uses inline type switch instead of `Modifies()`

**Python** (`get_calls_mods`, ivy_isolate.py:523-527):
```python
for sub in action.iter_subactions():
    for sym in sub.modifies():
        xtracer.trace("... sym=%s type=%s" % (actname, sym, type(sub).__name__))
        if sym in interf_syms:
            amods.add(sym)
```
- Calls `sub.modifies()` on each subaction — this method walks destructor chains for AssignAction/HavocAction, traverses hierarchy for CrashAction, and returns `[]` for all other types (including SetAction).
- Returns **symbol objects**, not strings.
- **Filters at collection time**: only adds to `amods` if `sym in interf_syms`.

**Go** (`GetCallsModsRecFull`, deps.go:127-145):
```go
switch a := sub.(type) {
case *actions.AssignAction:
    if c, ok := a.LHS.(*lg.Const); ok { amods[c.Name] = true }
case *actions.HavocAction:
    if c, ok := a.Target.(*lg.Const); ok { amods[c.Name] = true }
case *actions.SetAction:
    if c, ok := a.Lit.(*lg.Const); ok { amods[c.Name] = true }
}
```
- **BUG 1**: Does NOT walk destructor chains. Just checks if LHS/Target is directly a `*lg.Const`. Python walks `while n.rep.name in destructor_sorts: n = n.args[0]`.
- **BUG 2**: Includes `SetAction`. Python's `SetAction` inherits base `modifies()` which returns `[]`.
- **BUG 3**: Missing `CrashAction`. Python's `CrashAction.modifies()` traverses hierarchy.
- **BUG 4**: Does NOT filter by `interf_syms` at collection time. Go filters after collection (deps.go:408-421). This changes the trace output.
- **BUG 5**: Does NOT pass `interf_syms` to the function at all.

**Fix**: Replace the inline type switch with a call to the existing `actions.Modifies(sub)` function (which already handles destructor walking, CrashAction hierarchy, and excludes SetAction), add `interf_syms` filtering at collection time, and pass `interf_syms` into the function. Remove the post-collection filter.

### Difference 2: `GetLocMods` recurses; Python's `get_loc_mods` does not

**Python** (`get_loc_mods`, ivy_isolate.py:590-594):
```python
def get_loc_mods(mod,actname):
    action = mod.actions[actname]
    res = [s for s in action.modifies() if s.name.startswith('fml:')]
    return res
```
- Calls `action.modifies()` on the top-level action.
- For a Sequence (the typical top-level), Python's base `Action.modifies()` returns `[]` — it does NOT recurse.
- Only AssignAction/HavocAction/CrashAction override `modifies()` to return non-empty.
- So for a top-level Sequence, this returns `[]` always.

**Go** (`GetLocMods`, helpers.go:1067-1081):
```go
func GetLocMods(mod *module.Module, actname string) []string {
    act, ok := mod.Actions.Get2(actname)
    modSet := actions.Modifies(act)
    // ... filters for "fml:" prefix
}
```
- Calls `actions.Modifies(act)` which recurses into all children of the top-level action.
- For a Sequence, Go's `modifiesRec` hits the `default` case and recurses into `.ActionArgs()` children.
- This finds AssignActions/HavocActions nested inside the Sequence.
- **BUG**: Go returns MORE symbols than Python. Python returns `[]` for Sequences.

**Fix**: Change `GetLocMods` to call a non-recursive `modifies()` that matches Python's behavior — just call the per-action `modifies` without recursing into children. Since the top-level action is typically a Sequence (which should return `[]`), this is equivalent to checking the action type and only returning modifications if it's directly an AssignAction/HavocAction/CrashAction.

### Difference 3: `interf_syms` not passed to `GetCallsModsRecFull`

**Python**: `get_calls_mods` takes `interf_syms` as its 8th parameter and filters `amods` inline:
```python
if sym in interf_syms:
    amods.add(sym)
```

**Go**: `GetCallsModsRecFull` does NOT take `interf_syms`. All modifications are collected unfiltered. Filtering happens later in `CheckInterferenceFull` (deps.go:408-421).

**Fix**: Add `interfSyms map[string]bool` parameter to `GetCallsModsRecFull` and filter at collection time, matching Python. Remove the post-collection filter from `CheckInterferenceFull`.

### Difference 4: Trace format for `type` field

**Python** traces `type(sub).__name__`:
```
isolate.GetCallsModsRecFull mod actname=X sym=Y type=AssignAction
```

**Go** traces hardcoded type strings:
```
isolate.GetCallsModsRecFull mod actname=X sym=Y type=Assign
```

The Python type name includes "Action" suffix. Go must match.

**Fix**: Change trace type strings from `Assign` → `AssignAction`, `Havoc` → `HavocAction`, `Set` → remove (SetAction has no modifies), add `CrashAction`.

### Difference 5: Mods trace uses `str(x)` in Python (symbol objects with sort info)

**Python** traces mods as `str(x)` for symbol objects. Go traces name strings from `map[string]bool`.

In `check_interference` post-loop trace:
```python
','.join(sorted(str(x) for x in mods[actname]))
```

Since `mods` contains symbol objects in Python, `str(x)` includes sort qualifiers. In Go, `mods` contains plain name strings, so there's no sort info.

**Fix**: This is a deeper issue. For now, the mods stored in Go are name-based (`map[string]bool`). Since Python stores symbol objects but traces them via `str()`, and Go stores names, the trace may still diverge on symbols with sort qualifiers. If the golden test shows mismatches here, we may need to store `*lg.Const` objects (or their PrettyFmla representations) instead of plain names. Assess after fixing the above bugs.

## Implementation Plan

### Step 1: Fix `GetCallsModsRecFull` — add `interfSyms` parameter and use `Modifies()`

**File**: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/deps.go`

1. Add `interfSyms map[string]bool` parameter to `GetCallsModsRecFull`:
```go
func GetCallsModsRecFull(
    mod *module.Module,
    summarizedActions map[string]bool,
    actname string,
    calls, mods map[string]map[string]bool,
    mixins map[string]map[string]bool,
    interfSyms map[string]bool,
    loops ...map[string][]actions.Action,
) {
```

2. Replace the inline type switch (lines 127-145) with:
```go
for _, sub := range action.IterSubactions() {
    // Collect modifications matching Python's sub.modifies()
    for _, sym := range actions.ModifiesSingle(sub) {
        xtracer.Trace("isolate.GetCallsModsRecFull mod actname=%s sym=%s type=%s",
            actname, sym.Name, actions.ActionTypeName(sub))
        if interfSyms == nil || interfSyms[sym.Name] {
            amods[sym.Name] = true
        }
    }
    // ... rest of loop (WhileAction, CallAction) unchanged
```

3. Pass `interfSyms` through recursive calls (line 160, 196):
```go
GetCallsModsRecFull(mod, summarizedActions, calledName, calls, mods, mixins, interfSyms, loops...)
```

4. Update `GetCallsModsRec` wrapper (line 86) to pass nil for interfSyms:
```go
GetCallsModsRecFull(mod, summarizedActions, actname, calls, mods, nil, nil, loops...)
```

5. Update call site in `CheckInterferenceFull` (line 392) to pass the name-based interfSyms:
```go
GetCallsModsRecFull(mod, summarizedActions, actname, calls, mods, mixinDeps, interfSyms, loops)
```

6. **Remove** the post-collection filter (lines 408-421) since filtering now happens at collection time.

### Step 2: Add `ModifiesSingle` — non-recursive single-action modifies

**File**: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/transforms.go`

Add a new function that matches Python's per-action `modifies()` semantics — handles AssignAction (with destructor walk), HavocAction (with destructor walk), CrashAction (with hierarchy walk), and returns `[]` for everything else (including SetAction and Sequence). Does NOT recurse into children:

```go
// ModifiesSingle returns the symbols modified by a single action, matching
// Python's Action.modifies() method semantics — no recursion into children.
// AssignAction and HavocAction walk destructor chains.
// CrashAction walks module hierarchy.
// All other types return nil (including SetAction and Sequence).
func ModifiesSingle(action Action, cfg ...*ActionsConfig) []*lg.Const {
    var acfg *ActionsConfig
    if len(cfg) > 0 {
        acfg = cfg[0]
    }
    switch a := action.(type) {
    case *AssignAction:
        // same destructor walk as modifiesRec AssignAction case
        ...
    case *HavocAction:
        // same destructor walk as modifiesRec HavocAction case
        ...
    case *CrashAction:
        // same hierarchy walk as modifiesRec CrashAction case
        ...
    default:
        return nil  // Python: return []
    }
}
```

This is distinct from the existing `Modifies()` which recurses. Extract the per-type logic from `modifiesRec` into `ModifiesSingle`.

### Step 3: Add `ActionTypeName` helper

**File**: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/transforms.go`

```go
// ActionTypeName returns the Python-style class name for trace matching.
func ActionTypeName(a Action) string {
    switch a.(type) {
    case *AssignAction: return "AssignAction"
    case *HavocAction:  return "HavocAction"
    case *CrashAction:  return "CrashAction"
    case *SetAction:    return "SetAction"
    case *CallAction:   return "CallAction"
    case *Sequence:     return "Sequence"
    case *IfAction:     return "IfAction"
    case *WhileAction:  return "WhileAction"
    default:            return fmt.Sprintf("%T", a)
    }
}
```

### Step 4: Fix `GetLocMods` — non-recursive to match Python

**File**: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/helpers.go`

Change from:
```go
modSet := actions.Modifies(act)  // recurses
```

To:
```go
modSet := actions.ModifiesSingle(act)  // does NOT recurse, matches Python
```

Since the top-level action is typically a Sequence, Python's `action.modifies()` returns `[]`, and now Go's `ModifiesSingle` will also return `nil` for Sequence.

### Step 5: Update all call sites

**Files that call `GetCallsModsRecFull`:**

1. `deps.go:86` — `GetCallsModsRec` wrapper: add `nil` for interfSyms
2. `deps.go:160` — recursive call in CallAction handler: pass `interfSyms` through
3. `deps.go:196` — recursive call in mixin handler: pass `interfSyms` through
4. `deps.go:392` — in `CheckInterferenceFull`: pass `interfSyms`

**Files that call `CheckInterferenceFull`:**

Already updated in previous work — no changes needed (it already accepts `interfSymsInsMap` and converts to name-based set internally).

### Step 6: Remove post-collection filter from `CheckInterferenceFull`

**File**: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/deps.go`

Remove lines 408-421 (the `if interfSyms != nil` block that filters mods after collection). This filtering now happens inside `GetCallsModsRecFull`.

## Files to Modify

1. **`/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/deps.go`** — `GetCallsModsRecFull` signature change, use `ModifiesSingle`, pass `interfSyms`, remove post-filter
2. **`/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/transforms.go`** — add `ModifiesSingle()` and `ActionTypeName()`
3. **`/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/helpers.go`** — `GetLocMods` change to use `ModifiesSingle`
4. **`/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/batch_fixes_test.go`** — update any test calls to match new signature

## Verification

```bash
cd ~/ivy/goivy && go build ./...        # must compile cleanly
cd ~/ivy/goivy && make golden            # trace should match further past 156175
cd ~/ivy/goivy/isolate && go test ./...  # existing tests pass
cd ~/ivy/goivy/actions && go test ./...  # existing tests pass
```
