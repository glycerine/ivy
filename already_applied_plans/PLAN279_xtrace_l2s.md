# Add Comprehensive xtracer/XTRACE Instrumentation to L2S & Ranking Pipelines

**Created:** 2026-04-12 ~11:30pm

## Context

Golden test `TestOrdLive` diverges at trace line 260353. Go emits `CallAction.__init__ uniqueID=700 counter=701` while Python emits `ast.LF.clone PRESERVE origid=226 counter=2306`. The existing traces only show LF.clone and CallAction.__init__ events — there's no way to tell which **pipeline stage** the divergence occurs in (which SharedStep, which modPass call). This plan adds systematic instrumentation to both Go and Python so the first divergence point can be precisely pinpointed.

**Principle:** Every trace added to Go must have an identical counterpart in Python (same format string, same data), since the golden test compares output line-by-line.

---

## Instrumentation Strategy

Three layers of instrumentation, from coarsest to finest:

1. **SharedStep boundary traces** — identify which stage diverges (low volume, highest value)
2. **modPass call traces** — identify which transform call diverges (medium volume, high value)
3. **ast.clone() level traces** — pinpoint exact clone operation (high volume, medium value)

---

## Layer 1: SharedStep Boundary Traces

Add ENTER/EXIT traces at every SharedStep call site in both l2s and ranking pipelines.

### Format

```
XTRACE: <pipeline>.SharedStep<N> ENTER nInvars=X nAsms=Y nBindings=Z nPrems=W
XTRACE: <pipeline>.SharedStep<N> EXIT nInvars=X nAsms=Y nBindings=Z nPrems=W
```

Where `<pipeline>` is `l2s` or `ranking`.

### Go Changes

**File: `check/l2s.go`** — Add ENTER/EXIT traces around each SharedStep call:
- Line 565: `SharedStep1_ConvertTemporals(cfg, model, modPass)`
- Line 593: `SharedStep3_CollectNamedBinders(cfg, model, full)`
- Line 596: `SharedBuildSaveAndWait(cfg)`
- Line 723: `SharedStep6_BuildTableau(cfg)`
- Line 728: `SharedStep7_InstrumentActions(cfg, model)`
- Line 733: `SharedStep8_PatchExports(cfg, model)`
- Line 778: `SharedStep11_ReplaceNamedBinders(cfg, model, modPass)`
- Line 786: `SharedStep12_BuildGoal(...)`

Pattern for each:
```go
xtracer.Trace("l2s.SharedStep1 ENTER nInvars=%d nAsms=%d nBindings=%d nPrems=%d",
    len(model.Invars), len(model.Asms), len(model.Bindings), len(prems))
SharedStep1_ConvertTemporals(cfg, model, modPass)
xtracer.Trace("l2s.SharedStep1 EXIT nInvars=%d nAsms=%d nBindings=%d nPrems=%d",
    len(model.Invars), len(model.Asms), len(model.Bindings), len(prems))
```

**File: `check/ranking.go`** — Same pattern around:
- Line 399: `SharedStep1_ConvertTemporals(icfg, model, modPass)`
- Line 407: `SharedStep3_CollectNamedBinders(icfg, model, false)`
- Line 410: `SharedBuildSaveAndWait(icfg)`
- Line 418: `SharedStep6_BuildTableau(icfg)`
- Line 423: `SharedStep7_InstrumentActions(icfg, model)`
- Line 428: `SharedStep8_PatchExports(icfg, model)`
- Line 470: `SharedStep11_ReplaceNamedBinders(icfg, model, modPass)`
- Line 476: `SharedStep12_BuildGoal(...)`

Use `ranking.SharedStepN` prefix instead of `l2s.SharedStepN`.

### Python Changes

**File: `ivy_l2s.py`** — Add matching traces at equivalent locations:
- Line 795: `mod_pass(replace_temporals_by_l2s_g)` ← SharedStep1 first modPass
- Line 808: `mod_pass(ilu.normalize_named_binders)` ← SharedStep1 second modPass
- Line 829-843: named_binders_conjs collection ← SharedStep3
- Line 882-897: save_state/done_waiting/reset_w ← SharedBuildSaveAndWait
- (SharedStep6, 7, 8, 11, 12 equivalents in the function body)

**File: `ivy_ranking.py`** — Same at equivalent locations:
- Line 559: `mod_pass(replace_temporals_by_l2s_g)` ← SharedStep1 first modPass
- Line 572: `mod_pass(ilu.normalize_named_binders)` ← SharedStep1 second modPass
- Line 589-602: named_binders_conjs ← SharedStep3
- (remaining SharedStep equivalents)

---

## Layer 2: modPass Call Traces with Canon() Dumps

Each modPass invocation already has an ENTER trace in l2s.go. Extend to:
1. Add ENTER trace to ranking.go modPass (currently has none)
2. Add EXIT trace to both l2s and ranking modPass
3. Add Canon() dumps of model state at ENTER and EXIT

### Format

```
XTRACE: <pipeline>.modPass ENTER transform=<name> nInvars=X nAsms=Y nBindings=Z nPrems=W
XTRACE: <pipeline>.modPass invars HASH canon=[<canon1> <canon2> ...]
XTRACE: <pipeline>.modPass EXIT transform=<name> nInvars=X nAsms=Y nBindings=Z nPrems=W
```

### Go Changes

**File: `check/l2s.go`** (modPass closure, line 526):
- Already has ENTER trace (line 534). Add transform name parameter.
- Add EXIT trace after premises loop (after line 559).
- Add Canon() dump of invars at EXIT.

**File: `check/ranking.go`** (modPass closure, line 366):
- Add ENTER trace matching l2s format (with `ranking.modPass` prefix).
- Add EXIT trace after postconds loop (after line 393).

To pass the transform name, change modPass signature to accept a name string:
```go
modPass := func(transformName string, transform func(lg.Expr) lg.Expr) {
    xtracer.Trace("l2s.modPass ENTER transform=%s nInvars=%d ...", transformName, ...)
    // ... existing body ...
    xtracer.Trace("l2s.modPass EXIT transform=%s nInvars=%d ...", transformName, ...)
}
```

Update all call sites:
- `modPass("ReplaceTemporals", cfg.ReplaceTemporals)` in SharedStep1
- `modPass("NormalizeNamedBinders", func(...) { ... })` in SharedStep1
- `modPass("ReplaceNamedBindersAst", func(...) { ... })` in SharedStep11

**File: `check/l2s_shared.go`** — Update SharedStep1 and SharedStep11 to pass transform name to modPass. Change modPass parameter type from `func(func(lg.Expr) lg.Expr)` to `func(string, func(lg.Expr) lg.Expr)`.

### Python Changes

**File: `ivy_l2s.py`** (mod_pass, line 759):
- Already has ENTER trace. Add transform name.
- Add EXIT trace.

**File: `ivy_ranking.py`** (mod_pass, line 525):
- Add ENTER and EXIT traces matching l2s format.

Change mod_pass to accept name:
```python
def mod_pass(transform, name=""):
    if __debug__:
        xtracer.trace("l2s.modPass ENTER transform=%s nInvars=%d ..." % (name, ...))
    # ... existing body ...
    if __debug__:
        xtracer.trace("l2s.modPass EXIT transform=%s nInvars=%d ..." % (name, ...))
```

Update call sites:
- `mod_pass(replace_temporals_by_l2s_g, "ReplaceTemporals")`
- `mod_pass(ilu.normalize_named_binders, "NormalizeNamedBinders")`

---

## Layer 3: ast.clone() Level Traces with Canon()/canon() Data

All Layer 3 traces include canonical s-expression dumps so the exact AST content at the divergence point is visible.

### 3a: Python AST.clone() Base Method

**File: `ivy_ast.py` line 31** — Add trace to the base `clone()` method with canon data:

```python
def clone(self, args):
    if __debug__: xtracer.trace("ast.clone ENTER type=%s HASH canon=%s" % (type(self).__name__, self.canon() if hasattr(self,'canon') else str(self)))
    res = type(self)(*args)
    if hasattr(self, 'lineno'):
        res.lineno = lineno_add_ref(self.lineno)
    if __debug__: xtracer.trace("ast.clone EXIT type=%s HASH canon=%s" % (type(self).__name__, res.canon() if hasattr(res,'canon') else str(res)))
    return res
```

**Note:** LabeledFormula.clone() (line 649) already has its own PRESERVE/FRESH traces. These will appear AFTER the base AST.clone traces since LF.clone calls AST.clone internally. Both Go and Python must produce the same sequence.

**Important:** `canon()` is available via `canon_ast.install()` which is called at startup. If `canon()` is not available on a particular type, fall back to `str()`.

### 3b: Go AST Node Clone Methods

Go doesn't have a single base clone — each type has its own. We need to add matching traces.

**Key question:** Python's `AST.clone()` fires for EVERY node type (logic expressions, actions, etc.). The Go equivalent is scattered across many types. To match, we need to add traces to all Go clone methods that are called during the l2s/ranking pipeline.

**Approach:** Rather than instrumenting every Clone() method in Go (there are dozens), add traces inside `transformAction` and the modPass loops where cloning happens. Use `Sexp()` for logic.Expr types and `Canon()` for AST types (both produce canonical s-expression strings):

**File: `check/l2s.go`** — In transformAction (line 844):
```go
func transformAction(act actions.Action, transform func(lg.Expr) lg.Expr) actions.Action {
    if act == nil { return nil }
    xtracer.Trace("transformAction ENTER type=%T", act)
    args := act.ActionArgs()
    // ... existing body ...
    result := act.ActionClone(newArgs)
    xtracer.Trace("transformAction EXIT type=%T", result)
    return result
}
```

**In modPass loops** — Add canon data when cloning LabeledFormulas:
```go
// invars loop:
xtracer.Trace("ast.clone ENTER type=LabeledFormula HASH canon=%v", inv.Canon())
model.Invars[i] = inv.Clone([]ast.Node{...}).(*ast.LabeledFormula)
xtracer.Trace("ast.clone EXIT type=LabeledFormula HASH canon=%v", model.Invars[i].Canon())
```

Same pattern for asms, prems, postconds loops.

### 3c: replace_named_binders_ast Traces with Canon Data

**Python file: `ivy_logic_utils.py` line 344:**
```python
def replace_named_binders_ast(ast, subs):
    if __debug__: xtracer.trace("ilu.replaceNamedBindersAst ENTER type=%s HASH canon=%s" % (type(ast).__name__, ast.canon() if hasattr(ast,'canon') else str(ast)))
    if is_named_binder(ast):
        result = subs.get(ast, ast)
        if __debug__: xtracer.trace("ilu.replaceNamedBindersAst EXIT type=%s found=%s" % (type(ast).__name__, str(result is not ast)))
        return result
    args = [replace_named_binders_ast(x, subs) for x in ast.args]
    if is_app(ast):
        result = subs.get(ast.rep, ast.rep)(*args)
        if __debug__: xtracer.trace("ilu.replaceNamedBindersAst EXIT type=%s app HASH canon=%s" % (type(ast).__name__, result.canon() if hasattr(result,'canon') else str(result)))
        return result
    result = ast.clone(args)
    if __debug__: xtracer.trace("ilu.replaceNamedBindersAst EXIT type=%s cloned HASH canon=%s" % (type(ast).__name__, result.canon() if hasattr(result,'canon') else str(result)))
    return result
```

**Go file: `logicutil/logic_utils.go`** — Add matching traces in `ReplaceNamedBindersAst` (line ~1126). Use `expr.Sexp()` for logic.Expr types:
```go
func ReplaceNamedBindersAst(ast logic.Expr, subs map[string]logic.Expr) logic.Expr {
    xtracer.Trace("ilu.replaceNamedBindersAst ENTER type=%T HASH canon=%s", ast, ast.Sexp())
    // ... existing body, with EXIT traces at each return path ...
}
```

### 3d: replace_temporals_by_named_binder_g_ast Traces

**Python file: `ivy_logic_utils.py` line 287:**
Add ENTER/EXIT traces at function entry and each return path, with canon data.

**Go equivalent:** The Go `ReplaceTemporals` function (built from cfg.ReplaceTemporals in l2s.go). Add matching traces with `Sexp()` data.

### 3e: normalize_named_binders Traces

**Python file: `ivy_logic_utils.py` line 255:**
Add ENTER/EXIT traces with canon data.

**Go equivalent:** `lu.NormalizeNamedBinders` in logicutil. Add matching traces with `Sexp()` data.

---

## Implementation Order (all upfront)

### Step 1: Go Pipeline Traces (l2s.go, ranking.go, l2s_shared.go)

1. Change modPass signature to `func(name string, transform func(lg.Expr) lg.Expr)` in l2s.go and ranking.go
2. Update modPass parameter type in l2s_shared.go (SharedStep1, SharedStep11)
3. Add modPass ENTER/EXIT traces with transform name + counts
4. Add ENTER/EXIT traces around all SharedStep calls in l2s.go
5. Add ENTER/EXIT traces around all SharedStep calls in ranking.go
6. Add ast.clone ENTER/EXIT + Canon() around LF Clone calls in modPass loops
7. Add transformAction ENTER/EXIT traces in l2s.go

### Step 2: Go Recursive Function Traces (logicutil)

8. Add ENTER/EXIT + Sexp() traces in ReplaceNamedBindersAst (logicutil/logic_utils.go)
9. Add ENTER/EXIT + Sexp() traces in NormalizeNamedBinders (logicutil/logic_utils.go)
10. Add ENTER/EXIT + Sexp() traces in ReplaceTemporalsByNamedBinderGAst (logicutil or check/)

### Step 3: Python Pipeline Traces (ivy_l2s.py, ivy_ranking.py)

11. Change mod_pass to accept name param in ivy_l2s.py, add EXIT trace
12. Add mod_pass ENTER/EXIT with name param in ivy_ranking.py
13. Add SharedStep boundary traces in ivy_l2s.py at equivalent locations
14. Add SharedStep boundary traces in ivy_ranking.py at equivalent locations

### Step 4: Python Recursive Function & Clone Traces

15. Add ast.clone ENTER/EXIT + canon() to AST.clone() in ivy_ast.py:31
16. Add ENTER/EXIT + canon() traces in replace_named_binders_ast (ivy_logic_utils.py:344)
17. Add ENTER/EXIT + canon() traces in replace_temporals_by_named_binder_g_ast (ivy_logic_utils.py:287)
18. Add ENTER/EXIT + canon() traces in normalize_named_binders (ivy_logic_utils.py:255)

### Step 5: Verify

19. `go build ./...` — compile clean
20. Run golden test: `cd ~/ivy/goivy/parser && go test -v -count=1 -run TestOrdLive`
21. Analyze the new divergence point — it should now include SharedStep + modPass + clone context

---

## Files to Modify

| File | Changes |
|------|---------|
| `check/l2s.go` | SharedStep boundary traces, modPass name param, transformAction traces |
| `check/ranking.go` | SharedStep boundary traces, modPass ENTER/EXIT + name param |
| `check/l2s_shared.go` | Update modPass param type to include name string |
| `logicutil/logic_utils.go` | Traces in ReplaceNamedBindersAst, NormalizeNamedBinders |
| `~/pyivy/ivy/ivy/ivy_l2s.py` | SharedStep boundary traces, mod_pass name param + EXIT trace |
| `~/pyivy/ivy/ivy/ivy_ranking.py` | SharedStep boundary traces, mod_pass ENTER/EXIT + name param |
| `~/pyivy/ivy/ivy/ivy_logic_utils.py` | Traces in replace_named_binders_ast, replace_temporals_by_named_binder_g_ast, normalize_named_binders |
| `~/pyivy/ivy/ivy/ivy_ast.py` | Trace in AST.clone() base method |

---

## Verification

1. `go build ./...` — compile clean
2. `cd ~/ivy/goivy/parser && go test -v -count=1 -run TestOrdLive` — run golden test
3. Check that all new traces match between Go and Python up to the divergence point
4. The divergence should now be pinned to a specific SharedStep + modPass call + clone operation
5. Once pinpointed, the actual bug fix becomes straightforward

---

## Volume Expectation

All 3 layers are implemented upfront per user request. Expected trace volume increase:
- **Layer 1:** ~16 SharedStep boundary traces per pipeline (negligible)
- **Layer 2:** ~6 modPass ENTER/EXIT traces per pipeline (negligible)
- **Layer 3:** Potentially thousands of per-clone traces (recursive functions clone every AST node). The golden test output will grow significantly but this is acceptable — the goal is maximum diagnostic power to pinpoint the exact divergence.

The Canon()/canon() data in traces will make lines long but each line is self-contained and diffable.
