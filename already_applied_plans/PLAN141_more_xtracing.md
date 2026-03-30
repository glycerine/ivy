# Plan: Comprehensive XTRACE Instrumentation for Phase 2 Divergence Detection

**Created**: 2026-03-30 09:15

## Context

The golden test (TestOrdLive) now PASSES through 151196 lines of matching XTRACE output with zero divergences. However, Go produces a fatal error that Python does not:

```
error: create_isolate(cf_live): call out to tar_clock.impl.next[implement15] may have visible effect on fml:y
```

This error comes from `CheckInterferenceFull` in `isolate/deps.go`. The entire interference-check pipeline and many surrounding processing phases have **zero XTRACE instrumentation**, making it impossible to determine why Go raises this error but Python does not.

### Likely Root Cause (to be confirmed by instrumentation)

During investigation, a concrete bug candidate was found: In Python's `get_calls_mods` (line 500), modifications are collected via `sub.modifies()`. The `SetAction` class **does not override** `modifies()` — it inherits `Action.modifies()` which returns `[]`. But in Go's `GetCallsModsRecFull` (line 135), `SetAction` is **explicitly included** in the mod collection:
```go
case *actions.SetAction:
    if c, ok := a.Lit.(*lg.Const); ok {
        amods[c.Name] = true
    }
```
This means Go over-reports modifications for SetAction, potentially causing the false interference error.

Additionally, Go computes `interfSyms` as `copyStringSet(allSyms2)` (all AST-referenced names), while Python computes `interf_syms = set(x for x in ivy_logic.all_symbols() if x in all_syms)` — filtering to only symbols registered in the logic module. Names like `fml:y` that appear in ASTs but aren't formal logic symbols would be in Go's set but not Python's.

The instrumentation below will confirm or refute these hypotheses and reveal any others.

## Instrumentation Plan

All traces use PascalCase labels. Every Go trace has a corresponding Python trace with the same label string.

---

### Area 1: Interference Check Core (`isolate/deps.go` + `ivy_isolate.py`)

**Priority: CRITICAL** — This is where the error is raised.

#### 1a. CheckInterferenceFull entry

**Go** (`isolate/deps.go`, `CheckInterferenceFull` entry ~line 342):
```go
xtracer.Trace("isolate.CheckInterferenceFull ENTER n_summarized=%d n_interfSyms=%d n_afterInits=%d",
    len(summarizedActions), len(interfSyms), len(afterInits))
```

**Python** (`ivy_isolate.py`, `check_interference` entry ~line 582):
```python
if __debug__: xtracer.trace("isolate.CheckInterferenceFull ENTER n_summarized=%d n_interfSyms=%d n_afterInits=%d" %
    (len(summarized_actions), len(interf_syms), len(after_inits)))
```

#### 1b. Summarized actions set

**Go** (after computing summarizedActions, before CheckInterferenceFull call):
```go
sortedSummarized := sortedKeys(summarizedActions)
xtracer.Trace("isolate.CheckInterferenceFull summarized=%s", strings.Join(sortedSummarized, ","))
```

**Python** (similarly):
```python
if __debug__: xtracer.trace("isolate.CheckInterferenceFull summarized=%s" % ','.join(sorted(summarized_actions)))
```

#### 1c. interfSyms set (sorted sample for comparison)

**Go**:
```go
sortedInterf := sortedKeys(interfSyms)
xtracer.Trace("isolate.CheckInterferenceFull interfSyms=%s", strings.Join(sortedInterf, ","))
```

**Python**:
```python
if __debug__: xtracer.trace("isolate.CheckInterferenceFull interfSyms=%s" % ','.join(sorted(str(x) for x in interf_syms)))
```

#### 1d. GetCallsModsRecFull per-action results

**Go** (inside the summarized loop, after `GetCallsModsRecFull`):
```go
for actname := range summarizedActions {
    GetCallsModsRecFull(...)
    sortedCalls := sortedKeys(calls[actname])
    sortedMods := sortedKeys(mods[actname])
    xtracer.Trace("isolate.GetCallsModsRecFull actname=%s calls=%s mods=%s",
        actname, strings.Join(sortedCalls, ","), strings.Join(sortedMods, ","))
}
```

**Python**:
```python
for actname in summarized_actions:
    get_calls_mods(...)
    if __debug__: xtracer.trace("isolate.GetCallsModsRecFull actname=%s calls=%s mods=%s" %
        (actname, ','.join(sorted(calls[actname])), ','.join(sorted(str(x) for x in mods[actname]))))
```

#### 1e. The error path — before raising the error

**Go** (line ~451, before `return fmt.Errorf("call out to...`):
```go
xtracer.Trace("isolate.CheckInterferenceFull ERROR_CALLOUT actname=%s called=%s cmods=%s",
    actname, called, strings.Join(modNames, ","))
```

**Python** (line ~613, before `raise iu.IvyError`):
```python
if __debug__: xtracer.trace("isolate.CheckInterferenceFull ERROR_CALLOUT actname=%s called=%s cmods=%s" %
    (actname, called, ','.join(sorted(map(str, cmods)))))
```

#### 1f. Non-summarized action iteration with call inspection

**Go** (line ~386):
```go
for actname, action := range newActions.All() {
    if summarizedActions[actname] { continue }
    xtracer.Trace("isolate.CheckInterferenceFull non_summarized actname=%s", actname)
    for _, sub := range action.IterSubactions() {
        ca, ok := sub.(*actions.CallAction)
        if !ok { continue }
        calledName := CanonAct(ca.CalleeName())
        xtracer.Trace("isolate.CheckInterferenceFull call_check actname=%s calledName=%s", actname, calledName)
```

**Python** (line ~594):
```python
for actname,action in new_actions.items():
    if actname not in summarized_actions:
        if __debug__: xtracer.trace("isolate.CheckInterferenceFull non_summarized actname=%s" % actname)
        for called_name in action.iter_calls():
            called_name = canon_act(called_name)
            if __debug__: xtracer.trace("isolate.CheckInterferenceFull call_check actname=%s calledName=%s" % (actname, called_name))
```

#### 1g. GetLocMods results

**Go** (after computing locmods):
```go
sortedLocs := sortedKeys(locmods[actname])
xtracer.Trace("isolate.GetLocMods actname=%s locmods=%s", actname, strings.Join(sortedLocs, ","))
```

**Python**:
```python
if __debug__: xtracer.trace("isolate.GetLocMods actname=%s locmods=%s" % (actname, ','.join(sorted(str(x) for x in locmods[actname]))))
```

---

### Area 2: Isolate Component Processing (`isolate/isolate.go` + `ivy_isolate.py`)

**Priority: HIGH** — Processing between erase_unrefed and interference check.

#### 2a. allSyms2 computation result

**Go** (after line ~1185):
```go
sortedSyms2 := sortedKeys(allSyms2)
xtracer.Trace("isolate.allSyms2 n=%d syms=%s", len(allSyms2), strings.Join(sortedSyms2, ","))
```

**Python** (after line ~1308):
```python
if __debug__: xtracer.trace("isolate.allSyms2 n=%d syms=%s" % (len(all_syms), ','.join(sorted(str(x) for x in all_syms))))
```

#### 2b. Signature filter step

**Go** (after signature filtering, ~line 1193):
```go
if mod.Sig != nil {
    remaining := make([]string, 0)
    for name := range mod.Sig.Symbols { remaining = append(remaining, name) }
    sort.Strings(remaining)
    xtracer.Trace("isolate.sig_filter_done n_remaining=%d", len(remaining))
}
```

**Python** (after line ~1318):
```python
if __debug__: xtracer.trace("isolate.sig_filter_done n_remaining=%d" % len(list(mod.sig.all_symbols())))
```

#### 2c. interfSyms after FollowDefinitions

**Go** (after line 1222):
```go
sortedInterf := sortedKeys(interfSyms)
xtracer.Trace("isolate.interfSyms_after_follow n=%d syms=%s", len(interfSyms), strings.Join(sortedInterf, ","))
```

**Python** (after line 1340):
```python
if __debug__: xtracer.trace("isolate.interfSyms_after_follow n=%d syms=%s" % (len(interf_syms), ','.join(sorted(str(x) for x in interf_syms))))
```

#### 2d. presentAfterInits / allAfterInits

**Go** (before interference check call):
```go
xtracer.Trace("isolate.presentAfterInits=%s", strings.Join(sortedStrings(presentAfterInits), ","))
xtracer.Trace("isolate.allAfterInits=%s", strings.Join(sortedKeys(allAfterInits), ","))
```

**Python**:
```python
if __debug__: xtracer.trace("isolate.presentAfterInits=%s" % ','.join(sorted(after_inits)))
if __debug__: xtracer.trace("isolate.allAfterInits=%s" % ','.join(sorted(all_after_inits)))
```

#### 2e. exported set

**Go** (where exported/publicActions is set, ~line 1136):
```go
sortedExp := make([]string, 0)
for k := range exported { sortedExp = append(sortedExp, k) }
sort.Strings(sortedExp)
xtracer.Trace("isolate.exported=%s", strings.Join(sortedExp, ","))
```

**Python** (around line 1285):
```python
if __debug__: xtracer.trace("isolate.exported=%s" % ','.join(sorted(exported)))
```

#### 2f. newActions put-in-place

**Go** (around line 1138):
```go
for name := range newActions.All() {
    xtracer.Trace("isolate.newActions_final actname=%s", name)
}
```

**Python** (around line 1287):
```python
for actname in new_actions:
    if __debug__: xtracer.trace("isolate.newActions_final actname=%s" % actname)
```

#### 2g. oldActions saved

**Go** (line 1135):
```go
for name := range oldActions.All() {
    xtracer.Trace("isolate.oldActions actname=%s", name)
}
```

**Python** (line 1283):
```python
for actname in old_actions:
    if __debug__: xtracer.trace("isolate.oldActions actname=%s" % actname)
```

#### 2h. origDefs

**Go** (after line 1067):
```go
for i, d := range origDefs {
    lbl := ""
    if d.Label != nil { lbl = fmt.Sprint(d.Label) }
    xtracer.Trace("isolate.origDefs[%d] label=%s", i, lbl)
}
```

**Python** (around line 1268):
```python
for i, d in enumerate(orig_defs):
    if __debug__: xtracer.trace("isolate.origDefs[%d] label=%s" % (i, str(d.label) if d.label else ""))
```

#### 2i. implMixins

**Go** (before interference check call):
```go
for actname, mixins := range implMixins {
    mixerNames := make([]string, 0)
    for _, m := range mixins { mixerNames = append(mixerNames, m.Mixer()) }
    sort.Strings(mixerNames)
    xtracer.Trace("isolate.implMixins actname=%s mixers=%s", actname, strings.Join(mixerNames, ","))
}
```

**Python** (correspondingly):
```python
for actname, mixins in impl_mixins.items():
    if __debug__: xtracer.trace("isolate.implMixins actname=%s mixers=%s" % (actname, ','.join(sorted(m.mixer() for m in mixins))))
```

---

### Area 3: Helper Functions (`isolate/helpers.go` + `ivy_isolate.py`)

**Priority: MEDIUM** — Called by the interference check.

#### 3a. GetCallouts entry/result

**Go** (`isolate/helpers.go` or `deps.go`, GetCallouts):
```go
xtracer.Trace("isolate.GetCallouts actname=%s", actname)
```

**Python**:
```python
if __debug__: xtracer.trace("isolate.GetCallouts actname=%s" % actname)
```

#### 3b. GetLocMods detail

**Go** (in GetLocMods function):
```go
func GetLocMods(mod *module.Module, actname string) []string {
    xtracer.Trace("isolate.GetLocMods ENTER actname=%s", actname)
    ...
}
```

**Python**:
```python
def get_loc_mods(mod, actname):
    if __debug__: xtracer.trace("isolate.GetLocMods ENTER actname=%s" % actname)
    ...
```

#### 3c. FollowDefinitions

**Go** (in FollowDefinitions):
```go
func FollowDefinitions(defs []*ast.LabeledFormula, syms map[string]bool) {
    before := len(syms)
    ... (existing logic)
    xtracer.Trace("isolate.FollowDefinitions before=%d after=%d", before, len(syms))
}
```

**Python**:
```python
def follow_definitions(ldfs, all_syms):
    before = len(all_syms)
    ... (existing logic)
    if __debug__: xtracer.trace("isolate.FollowDefinitions before=%d after=%d" % (before, len(all_syms)))
```

#### 3d. FindReferences

**Go** (in FindReferences):
```go
xtracer.Trace("isolate.FindReferences n_syms=%d n_refs=%d", len(syms), len(refs))
```

**Python**:
```python
if __debug__: xtracer.trace("isolate.FindReferences n_syms=%d n_refs=%d" % (len(syms), len(refs)))
```

---

### Area 4: Action modifies() — Structural Comparison

**Priority: HIGH** — Directly relevant to the suspected SetAction bug.

#### 4a. GetCallsModsRecFull per-subaction detail

**Go** (inside `GetCallsModsRecFull`, per subaction):
```go
for _, sub := range action.IterSubactions() {
    switch a := sub.(type) {
    case *actions.AssignAction:
        if c, ok := a.LHS.(*lg.Const); ok {
            xtracer.Trace("isolate.GetCallsModsRecFull mod actname=%s sym=%s type=Assign", actname, c.Name)
            amods[c.Name] = true
        }
    case *actions.HavocAction:
        if c, ok := a.Target.(*lg.Const); ok {
            xtracer.Trace("isolate.GetCallsModsRecFull mod actname=%s sym=%s type=Havoc", actname, c.Name)
            amods[c.Name] = true
        }
    case *actions.SetAction:
        if c, ok := a.Lit.(*lg.Const); ok {
            xtracer.Trace("isolate.GetCallsModsRecFull mod actname=%s sym=%s type=Set", actname, c.Name)
            amods[c.Name] = true
        }
    }
```

**Python** (in `get_calls_mods`, inside the subaction loop):
```python
for sub in action.iter_subactions():
    for sym in sub.modifies():
        if __debug__: xtracer.trace("isolate.GetCallsModsRecFull mod actname=%s sym=%s type=%s" % (actname, sym, type(sub).__name__))
        if sym in interf_syms:
            amods.add(sym)
```

---

### Area 5: Summarized Actions Computation

**Priority: HIGH** — Need to understand what Go vs Python consider "summarized".

#### 5a. Trace where summarizedActions is built

In `isolate/isolate.go`, find where `summarizedActions` map is populated and add:
```go
xtracer.Trace("isolate.summarizedActions_add actname=%s", actname)
```

And the corresponding Python point where `summarized_actions` is built.

---

### Area 6: Additional Uninstrumented Critical Paths

**Priority: MEDIUM** — For future divergence detection beyond current error.

#### 6a. transrel/transrel.go + ivy_transrel.py

Add entry/exit traces to:
- `SatisfyWithCond` (already partially traced in Go)
- `TransitionRelation` / `UpdateWrapper` main entry points

#### 6b. check/isolate_check.go + ivy_check.py

Add traces to:
- `CheckIsolate` inner loop body (per-conjecture checking)
- `CheckModule` per-isolate processing
- Solver invocation decision points

#### 6c. logic operations (logic/logic.go + logic.py)

Add traces to:
- Sort inference entry points used during isolate processing
- Symbol lookup/resolution

---

## Files to Modify

### Go files:
1. **`isolate/deps.go`** — Areas 1a-1g, 4a (CheckInterferenceFull, GetCallsModsRecFull, GetLocMods, GetCallouts, FindReferences)
2. **`isolate/isolate.go`** — Areas 2a-2i, 5a (allSyms2, sig filter, interfSyms, afterInits, exported, newActions, oldActions, origDefs, implMixins, summarizedActions)
3. **`isolate/helpers.go`** — Areas 3a-3d (GetCallouts, GetLocMods, FollowDefinitions, FindReferences)

### Python files:
4. **`~/pyivy/ivy/ivy/ivy_isolate.py`** — Areas 1a-1g, 2a-2i, 3a-3d, 4a, 5a (all corresponding Python trace points)

### Optional (Area 6):
5. **`transrel/transrel.go`** + **`~/pyivy/ivy/ivy/ivy_transrel.py`**
6. **`check/isolate_check.go`** + **`~/pyivy/ivy/ivy/ivy_check.py`**

## Helper Utility Needed

A `sortedKeys` helper for Go (may already exist):
```go
func sortedKeys(m map[string]bool) []string {
    keys := make([]string, 0, len(m))
    for k := range m { keys = append(keys, k) }
    sort.Strings(keys)
    return keys
}
```

## Verification

1. `go build ./...` — ensure compilation
2. `cd ~/goivy && make golden` — run the golden test
3. Examine `~/goivy/log.red` for the FIRST divergence in the new traces
4. The new traces should reveal:
   - Whether `interfSyms` differs between Go and Python
   - Whether `summarizedActions` differs
   - Whether `mods` computed by GetCallsModsRecFull differ (confirming the SetAction bug)
   - The exact data path that leads to Go's false error

## Suspected Fixes (to apply after instrumentation confirms)

1. **SetAction mods**: Remove `case *actions.SetAction` from `GetCallsModsRecFull` — Python's `SetAction.modifies()` returns `[]`
2. **interfSyms computation**: Change Go to match Python's `set(x for x in ivy_logic.all_symbols() if x in all_syms)` pattern — filtering to only registered logic symbols, not all AST-referenced names
