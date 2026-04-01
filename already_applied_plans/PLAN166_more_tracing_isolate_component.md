# Fix allSyms2 Trace Divergence at Line 155604

**Created:** 2026-04-01 19:00

## Context

After switching allSyms/allSyms2 to insertion-ordered containers (InsMap in Go, OrderedSymSet in Python), the golden test now passes through line 155603 but diverges at 155604:

```
155602  go : XTRACE: isolate.allSyms2.add isd.wr        (match)
155603  go : XTRACE: isolate.allSyms2.add isd.rd         (match)
155604  go : XTRACE: isolate.allSyms2.add isd.rsp        ← DIVERGE
        py : XTRACE: isolate.allSyms2.add ref.lt         ← DIVERGE
```

Both symbols are eventually collected in both sides, but one side sees `isd.rsp` first while the other sees `ref.lt` first. This is a formula-level ordering or AST-structural difference exposed by the stronger insertion-order check.

## Investigation Summary (What We Ruled Out)

1. **allSyms2 phase structure** — Both Go and Python iterate the same 6 phases (A-F) in identical order. Confirmed by code comparison.
2. **Action type Children() vs .args ordering** — All action types return children in the same order. Verified for Sequence, IfAction, AssignAction, CallAction, etc.
3. **is_app / IsApp** — Both sides treat App, Const, and 0-var NamedBinder as "app". Matches.
4. **Apply.args includes func in Python but NodeArgs excludes it in Go** — Doesn't affect symbol yield order because re-processing func from args produces only duplicates (already in set).
5. **Formula list filtering** — All filter operations preserve list order (iterate and filter, never reorder). Confirmed in isolate.go lines 720-860.
6. **Action map ordering** — `mod.Actions` is InsMap (Go) / dict (Python), both insertion-ordered. Populated from same loop (`mod.Actions.All()` at line 507).

## Root Cause Hypothesis

The divergence is either:
- **(A) Formula/action ordering** — The formulas in one of the lists (axioms, props, inits, conjs, definitions) or actions are in a different order between Go and Python. This would mean two formulas exist in both sides but at different list positions.
- **(B) AST structural difference** — A single formula has different subterm ordering (e.g., `And.Terms` in different order) between Go and Python, causing depth-first traversal to yield symbols in different order.

## Plan: Diagnostic Traces to Pinpoint the Formula

### Step 1: Add phase-boundary and per-formula traces

Add matching XTRACE lines in both Go and Python **inside** the allSyms2 collection code. These traces emit BEFORE processing each item, so the test comparison will reveal exactly which formula/action is being processed at the divergence point.

#### Go: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/isolate.go`

In the allSyms2 section (lines 1343-1422), add traces at each phase boundary and before each formula/action:

```go
// Phase A (line ~1343): add before the inner loop body
if xtracer.Enabled {
    lbl := ""
    if lf.Label != nil {
        lbl = lg.ReprNode(lf.Label)
    }
    xtracer.Trace("%s.phaseA_fmla %s", as2, lbl)
}
collectSymbolsInto(as2, lf.Formula.(lg.Expr), allSyms2)

// Phase B (line ~1361): add before collectSymbolsInto
if xtracer.Enabled {
    xtracer.Trace("%s.phaseB_action %s", as2, name)
}
collectSymbolsInto(as2, act, allSyms2)

// Phase C (line ~1367): add trace
xtracer.Trace("%s.phaseC_params_start", as2)

// Phase D (line ~1382): add before each action's formals
if xtracer.Enabled {
    xtracer.Trace("%s.phaseD_formals %s", as2, name)
}

// Phase E (line ~1405): add trace
xtracer.Trace("%s.phaseE_natives_start", as2)

// Phase F (line ~1416): add trace
xtracer.Trace("%s.phaseF_proofs_start", as2)
```

Note: Phase B needs the action name. Currently the loop is `for _, act := range mod.Actions.All()` — need to capture the name: `for name, act := range mod.Actions.All()`.

#### Python: `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_isolate.py`

Add matching traces at the same locations (lines 1410-1435):

```python
# Phase A (line ~1413): add before _traced_add_syms
if __debug__: xtracer.trace("%s.phaseA_fmla %s" % (_as2_label, str(y.label)))
_traced_add_syms(_as2_label, all_syms, lu.symbols_ilu_ast(y.formula))

# Phase B (line ~1417): add before each action
for actname, action in list(mod.actions.items()):
    if __debug__: xtracer.trace("%s.phaseB_action %s" % (_as2_label, actname))
    _traced_add_syms(_as2_label, all_syms, lu.symbols_ilu_ast(action))

# Phase C: add trace
if __debug__: xtracer.trace("%s.phaseC_params_start" % _as2_label)

# Phase D: add before each action's formals
for actname, a in list(mod.actions.items()):
    if __debug__: xtracer.trace("%s.phaseD_formals %s" % (_as2_label, actname))

# Phase E: add trace
if __debug__: xtracer.trace("%s.phaseE_natives_start" % _as2_label)

# Phase F: add trace
if __debug__: xtracer.trace("%s.phaseF_proofs_start" % _as2_label)
```

**Important for Phase B Python**: Currently `for action in list(mod.actions.values())` — must change to `for actname, action in list(mod.actions.items())` to get the action name for the trace.

### Step 2: Run `make golden` and analyze output

```bash
cd ~/ivy/goivy && make golden
```

The test will now diverge at one of:
- A `phaseX_fmla`/`phaseB_action` trace — meaning the formula/action ordering differs
- An `allSyms2.add` trace between two matching `phaseA_fmla` traces — meaning the AST structure within that specific formula differs

### Step 3: Fix based on findings

#### If formula ordering differs (case A):
- The diagnostic traces will show which formula label appears in the wrong position
- Track where that formula is inserted into the list (axioms/props/etc.)
- Look for map iteration or non-deterministic ordering in the code path that populates the list
- Fix by ensuring deterministic ordering (likely sorting or using InsMap at that point)

#### If AST structure differs (case B):
- The diagnostic traces will identify the exact formula (by label)
- Use canon() comparison on that specific formula to find the structural difference
- The canon diff will show which AST node (e.g., And.Terms order) differs
- Fix the compiler/transformation step that constructs that formula

### Step 4: Remove diagnostic traces

Once the root cause is fixed, remove all the `phaseA_fmla`, `phaseB_action`, etc. diagnostic traces (they were only for investigation).

## Files to Modify

1. **`/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/isolate.go`** — lines 1343-1422 (allSyms2 section): add diagnostic traces
2. **`/Users/jaten/ivy/pyivy/ivy/ivy/ivy_isolate.py`** — lines 1410-1435 (allSyms2 section): add matching diagnostic traces

## Verification

```bash
cd ~/ivy/goivy && make golden
```

The diagnostic traces will either:
- Push the matching point further (if formula ordering matches) — the divergence will be narrowed to a specific formula
- Expose the formula ordering difference immediately — the phaseA/B trace itself will mismatch

After fixing the root cause:
```bash
cd ~/ivy/goivy && make golden     # should pass further than 155604
cd ~/ivy/goivy/isolate && go test ./...
cd ~/ivy/goivy/actions && go test ./...
```
