# Plan: Fix ActionInterferenceCheck Divergence via Instrumentation

**Created:** 2026-03-29 00:15

## Context

The `make golden` test (`TestOrdLive`) compares Go and Python xtrace output line-by-line. Both match through 146,109 lines, then diverge at ActionInterferenceCheck:

```
Go  line 146110: compiler.ActionInterferenceCheck action=ref.init[after93] modifies=2
Py  line 146110: compiler.ActionInterferenceCheck action=isd.init[after79] modifies=3
```

The xtrace only prints actions with non-zero `modifies`. Full comparison:

| Position | Go | Python |
|---|---|---|
| 9 | `ref.init[after93]` modifies=2 | `isd.init[after79]` modifies=3 |
| 10 | `ref.perform` modifies=2 | `ref.init[after93]` modifies=12 |

This means:
- Go: `isd.init[after79]` has **0 modifies** (silent) — Python has **3**
- Go: `ref.perform` has **2 modifies** — Python has **0** (silent)
- Go: `ref.init[after93]` has **2 modifies** — Python has **12**

Both sides compile **identical action names** during ARGSetup (confirmed by matching xtrace at lines 126314-127702). The divergence is in the **action body structure** or the **Modifies() traversal** differing between Go and Python.

## Root Cause Hypotheses

1. **Action bodies differ** — Mixin application during ARGSetup or a post-ARGSetup phase (TypeCheckAction, etc.) modifies the action tree differently
2. **Modifies() traversal bug** — Go's `modifiesRec()` handles some action type differently than Python's `iter_subactions()`+`modifies()` pattern
3. **isDestructor() differs** — The destructor chain walk in AssignAction/HavocAction may skip/include different symbols

## Plan: Instrument Both Sides

### Step 1: Add action-map key dump at ActionInterferenceCheck (Go)

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/ivy_compile.go` (~line 1475)

Right after `ActionInterferenceCheck ENTER`, dump ALL action keys in insertion order:

```go
// Dump all action keys for comparison
{
    var allKeys []string
    for name := range mod.Actions.All() {
        allKeys = append(allKeys, name)
    }
    xtracer.Trace("compiler.ActionInterferenceCheck allKeys=%d keys=%s", len(allKeys), strings.Join(allKeys, ","))
}
```

### Step 2: Add action-map key dump at ActionInterferenceCheck (Python)

**File:** `/Users/jaten/pyivy/ivy/ivy/ivy_compiler.py` (~line 1982)

Right after `ActionInterferenceCheck ENTER`, dump ALL action keys:

```python
if __debug__:
    all_keys = list(mod.actions.keys())
    xtracer.trace("compiler.ActionInterferenceCheck allKeys=%d keys=%s" % (len(all_keys), ",".join(all_keys)))
```

### Step 3: Dump modified symbol names for each action (Go)

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/ivy_compile.go` (~line 1505)

Enhance the existing trace to also show the modified symbol names:

```go
if len(modSyms) > 0 {
    var symNames []string
    for k := range modSyms {
        symNames = append(symNames, fmt.Sprintf("%v", k))
    }
    sort.Strings(symNames)
    xtracer.Trace("compiler.ActionInterferenceCheck action=%s modifies=%d syms=%s", name, len(modSyms), strings.Join(symNames, ","))
}
```

### Step 4: Dump modified symbol names for each action (Python)

**File:** `/Users/jaten/pyivy/ivy/ivy/ivy_compiler.py` (~line 1997)

```python
if mod_syms:
    sym_names = sorted(str(s) for s in mod_syms)
    xtracer.trace("compiler.ActionInterferenceCheck action=%s modifies=%d syms=%s" % (name, len(mod_syms), ",".join(sym_names)))
```

### Step 5: Dump action body canon for divergent actions (Go)

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/ivy_compile.go` (~line 1507)

After the modifies trace, also dump the action's canonical s-expression for the specific divergent actions:

```go
if name == "isd.init[after79]" || name == "ref.init[after93]" || name == "ref.perform" {
    if canonizable, ok := act.(interface{ Canon() iu.Canonical }); ok {
        canon := string(canonizable.Canon())
        if len(canon) > 2000 { canon = canon[:2000] + "..." }
        xtracer.Trace("compiler.ActionInterferenceCheck action=%s canon=%s", name, canon)
    }
}
```

### Step 6: Dump action body canon for divergent actions (Python)

**File:** `/Users/jaten/pyivy/ivy/ivy/ivy_compiler.py` (~line 1997)

```python
if name in ("isd.init[after79]", "ref.init[after93]", "ref.perform"):
    try:
        c = actval.canon() if hasattr(actval, 'canon') else repr(actval)
        if len(c) > 2000: c = c[:2000] + "..."
        xtracer.trace("compiler.ActionInterferenceCheck action=%s canon=%s" % (name, c))
    except: pass
```

### Step 7: Dump action canon for ZERO-modifies divergent actions too

Since `isd.init[after79]` has 0 modifies in Go (so never enters the `if len(modSyms) > 0` block), add a separate trace outside that condition:

```go
// In the second loop, after computing modSyms:
if name == "isd.init[after79]" || name == "ref.init[after93]" || name == "ref.perform" {
    // Always trace these, even if modifies=0
    var symNames []string
    for k := range modSyms {
        symNames = append(symNames, fmt.Sprintf("%v", k))
    }
    sort.Strings(symNames)
    xtracer.Trace("compiler.ActionInterferenceCheck DETAIL action=%s modifies=%d syms=%s", name, len(modSyms), strings.Join(symNames, ","))
    if canonizable, ok := act.(interface{ Canon() iu.Canonical }); ok {
        canon := string(canonizable.Canon())
        if len(canon) > 2000 { canon = canon[:2000] + "..." }
        xtracer.Trace("compiler.ActionInterferenceCheck DETAIL action=%s canon=%s", name, canon)
    }
}
```

Same in Python — trace unconditionally for these three action names.

### Step 8: Build and run

```bash
cd ~/goivy && make golden
```

Review `log.red` for the new DETAIL traces and compare the action canons and modified symbol names between Go and Python.

## Critical Files

- `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/ivy_compile.go` — ActionInterferenceCheck (line 1472)
- `/Users/jaten/pyivy/ivy/ivy/ivy_compiler.py` — check_definitions (line 1979)
- `/Users/jaten/go/src/github.com/glycerine/goivy/actions/transforms.go` — `Modifies()` / `modifiesRec()` (line 115)
- `/Users/jaten/pyivy/ivy/ivy/ivy_actions.py` — `iter_subactions()` / `modifies()` (line 271)

## Verification

1. Run `cd ~/goivy && make golden`
2. Check `log.red` for the new `DETAIL` traces
3. Compare:
   - Do both sides have the same action keys? (allKeys trace)
   - For `isd.init[after79]`: why does Go have 0 modifies but Python has 3? Compare their canons.
   - For `ref.init[after93]`: why does Go have 2 modifies but Python has 12? Compare their symbol lists.
   - For `ref.perform`: why does Go have 2 modifies but Python has 0? Compare their canons.
4. The canon diff will pinpoint exactly where the action trees diverge, pointing to the upstream bug (likely in CompileAction, mixin application, or TypeCheckAction).
