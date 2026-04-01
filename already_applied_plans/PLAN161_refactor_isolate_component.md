# Plan: Refactor IsolateComponent() for Online Per-Symbol Tracing

**Created:** 2026-03-31 ~afternoon

## Context

`IsolateComponent()` in `goivy/isolate/isolate.go:266` is the Go port of `isolate_component()` in `pyivy/ivy/ivy/ivy_isolate.py:919`. The Go port has divergences from Python's execution flow in symbol collection (allSyms/allSyms2). A divergence was observed in `isolate.allSyms2.sym` traces. The goal is to add **online, per-symbol** xtracer tracing so divergences are detected immediately as they happen. Both sides must emit the same trace for each new symbol as it is added, in the same order, so a line-by-line diff pinpoints the exact first divergence.

**Key constraint:** All tracing is inline in the production code, guarded by `xtracer.Enabled` (compile-time constant, zero-cost when off). No separate traced/untraced code paths or functions.

## Divergence: allSyms2 collection ordering

**Python** (ivy_isolate.py:1363-1382) processes:
1. Formulas from `[axioms, props, inits, conjs, definitions]`
2. Action bodies: `asts.extend(action for action in mod.actions.values())`
3. Params: `asts.extend(mod.params)` (if keep_destructors)
4. Action formals: separate loop for `formal_params` + `formal_returns`
5. Natives: `asts.extend(tmp.args[2:])`
6. Proofs: `asts.extend(x[1] for x in mod.proofs)`

**Go** (isolate.go:1278-1335) processes DIFFERENTLY:
1. Formulas (same)
2. **Action formals AND bodies INTERLEAVED** per-action
3. Params
4. Natives
5. Proofs

Go must match Python's exact order.

## Implementation Steps

### Step 1: Modify `collectSymbolsInto` to accept a label and trace inline

**File:** `goivy/isolate/helpers.go:405`

Change the signature to accept a `label string` and add inline tracing:

```go
func collectSymbolsInto(label string, node lg.Expr, syms map[lg.NodeKey]lg.Expr) {
    if node == nil {
        return
    }
    for sym := range il.SymbolsIluAst(node) {
        key := lg.Key(sym)
        if xtracer.Enabled {
            if _, exists := syms[key]; !exists {
                xtracer.Trace("%s.add %s", label, lg.PrettyFmla(sym))
            }
        }
        syms[key] = sym
    }
}
```

Update ALL existing callers of `collectSymbolsInto` to pass a label string (first argument). Callers in `isolate.go` already know which phase they're in, so the label is natural.

### Step 2: Add inline tracing to direct map insertions in Go

For every place in `isolate.go` where a symbol is inserted directly into allSyms or allSyms2 via `syms[key] = val`, add the inline check:

```go
key := actions.ConstSymKey(p)
if xtracer.Enabled {
    if _, exists := allSyms2[key]; !exists {
        xtracer.Trace("isolate.allSyms2.add %s", lg.PrettyFmla(p))
    }
}
allSyms2[key] = p
```

This applies to:
- Formal params/returns insertions (allSyms lines ~941-946, allSyms2)
- Definition-symbol additions (allSyms line ~1029)
- Destructor additions (allSyms line ~1062, allSyms2 line ~1329)
- Params additions (allSyms2)

### Step 3: Add `_traced_add_syms` to Python (inline in production code)

**File:** `pyivy/ivy/ivy/ivy_isolate.py` (near line 36)

```python
def _traced_add_syms(label, target_set, new_syms):
    """Add symbols to set, tracing each NEW addition inline."""
    for sym in new_syms:
        if sym not in target_set:
            if __debug__: xtracer.trace("%s.add %s" % (label, str(sym)))
        target_set.add(sym)
```

When `python -O` is used, `__debug__` is False and the trace call is eliminated at bytecode compile time. This IS the production code path -- one function, one flow.

### Step 4: Restructure Go allSyms2 to match Python order + inline traces

**File:** `goivy/isolate/isolate.go`, lines ~1278-1335

Replace the current allSyms2 block:

```go
allSyms2 := make(map[lg.NodeKey]lg.Expr)
const as2 = "isolate.allSyms2"

// Phase A: formulas (axioms, props, inits, conjs, definitions)
for _, lfSlice := range [][]*ast.LabeledFormula{
    mod.LabeledAxioms, mod.LabeledProps, mod.LabeledInits, mod.LabeledConjs, mod.Definitions,
} {
    for _, lf := range lfSlice {
        if lf.Formula != nil {
            if _, isSchema := lf.Formula.(*ast.SchemaBody); isSchema {
                continue
            }
            collectSymbolsInto(as2, lf.Formula.(lg.Expr), allSyms2)
        }
    }
}

// Phase B: action bodies ONLY (not formals -- Python does them in Phase D)
for _, act := range mod.Actions.All() {
    collectSymbolsInto(as2, act, allSyms2)
}

// Phase C: params (if keep_destructors)
if isoCfg.KeepDestructors {
    for _, p := range mod.Params {
        key := actions.ConstSymKey(p)
        if xtracer.Enabled {
            if _, exists := allSyms2[key]; !exists {
                xtracer.Trace("%s.add %s", as2, lg.PrettyFmla(p))
            }
        }
        allSyms2[key] = p
    }
}

// Phase D: action formals (separate pass, matching Python)
for _, act := range mod.Actions.All() {
    for _, p := range act.GetFormalParams() {
        key := actions.ConstSymKey(p)
        if xtracer.Enabled {
            if _, exists := allSyms2[key]; !exists {
                xtracer.Trace("%s.add %s", as2, lg.PrettyFmla(p))
            }
        }
        allSyms2[key] = p
    }
    for _, r := range act.GetFormalReturns() {
        key := actions.ConstSymKey(r)
        if xtracer.Enabled {
            if _, exists := allSyms2[key]; !exists {
                xtracer.Trace("%s.add %s", as2, lg.PrettyFmla(r))
            }
        }
        allSyms2[key] = r
    }
}

// Phase E: natives
for _, nat := range mod.Natives {
    args := nat.Args()
    for i := 2; i < len(args); i++ {
        if expr, ok := args[i].(lg.Expr); ok {
            collectSymbolsInto(as2, expr, allSyms2)
        }
    }
}

// Phase F: proofs
for _, pe := range mod.Proofs {
    if pe.Proof != nil {
        if n, ok := pe.Proof.(lg.Expr); ok {
            collectSymbolsInto(as2, n, allSyms2)
        }
    }
}
```

### Step 5: Restructure Python allSyms2 with inline online traces

**File:** `pyivy/ivy/ivy/ivy_isolate.py`, lines ~1363-1382

Replace the flat-list-then-set approach with per-formula online tracing:

```python
all_syms = set()
label = "isolate.allSyms2"

# Phase A: formulas
for x in [mod.labeled_axioms, mod.labeled_props, mod.labeled_inits, mod.labeled_conjs, mod.definitions]:
    for y in x:
        if not isinstance(y.formula, ivy_ast.SchemaBody):
            _traced_add_syms(label, all_syms, lu.used_symbols_ast(y.formula))

# Phase B: action bodies
for action in list(mod.actions.values()):
    _traced_add_syms(label, all_syms, lu.used_symbols_ast(action))

# Phase C: params
if opt_keep_destructors.get():
    _traced_add_syms(label, all_syms, lu.used_symbols_asts(mod.params))

# Phase D: action formals
for a in list(mod.actions.values()):
    _traced_add_syms(label, all_syms, lu.used_symbols_asts(a.formal_params))
    _traced_add_syms(label, all_syms, lu.used_symbols_asts(a.formal_returns))

# Phase E: natives
for tmp in mod.natives:
    _traced_add_syms(label, all_syms, lu.used_symbols_asts(tmp.args[2:]))

# Phase F: proofs
for x in mod.proofs:
    _traced_add_syms(label, all_syms, lu.used_symbols_ast(x[1]))
```

**Semantics preserved:** Python's original collected everything into `asts` then called `set(lu.used_symbols_asts(asts))`. Since `used_symbols_asts([a,b,c])` = `chain(symbols_ilu_ast(a), symbols_ilu_ast(b), symbols_ilu_ast(c))` converted to set, and our refactored code walks each AST individually adding to the same set, the final result is identical. The DFS order within each AST is preserved.

### Step 6: Add inline online traces to Go allSyms (first collection)

**File:** `goivy/isolate/isolate.go`, lines ~918-987

Update the existing `collectSymbolsInto` calls to pass `"isolate.allSyms"` as label (they currently have no label). Update direct map insertions for formals to include inline traces:

```go
// Formulas -- now passes label
collectSymbolsInto("isolate.allSyms", lf.Formula.(lg.Expr), allSyms)

// Formals -- inline trace before each insertion
for _, p := range act.GetFormalParams() {
    key := actions.ConstSymKey(p)
    if xtracer.Enabled {
        if _, exists := allSyms[key]; !exists {
            xtracer.Trace("isolate.allSyms.add %s", lg.PrettyFmla(p))
        }
    }
    allSyms[key] = p
}

// Natives -- now passes label
collectSymbolsInto("isolate.allSyms", expr, allSyms)

// Definition-symbol addition (line ~1029) -- inline trace
key := actions.ConstSymKey(c)
if xtracer.Enabled {
    if _, exists := allSyms[key]; !exists {
        xtracer.Trace("isolate.allSyms.add %s", lg.PrettyFmla(c))
    }
}
allSyms[key] = c
```

### Step 7: Add inline online traces to Python allSyms (first collection)

**File:** `pyivy/ivy/ivy/ivy_isolate.py`, lines ~1231-1260

Replace batch collection with per-formula online tracing:

```python
label = "isolate.allSyms"

# Phase 1: formulas
all_syms_raw = set()
for x in [mod.labeled_axioms, mod.labeled_props, mod.labeled_inits, mod.labeled_conjs]:
    for y in x:
        if not isinstance(y.formula, ivy_ast.SchemaBody):
            _traced_add_syms(label, all_syms_raw, lu.used_symbols_ast(y.formula))
_trace_sym_set("isolate.allSyms_post_formulas", all_syms_raw)

# Phase 2: action formals
for a in list(mod.actions.values()):
    _traced_add_syms(label, all_syms_raw, lu.used_symbols_asts(a.formal_params))
    _traced_add_syms(label, all_syms_raw, lu.used_symbols_asts(a.formal_returns))
_trace_sym_set("isolate.allSyms_post_formals", all_syms_raw)

# Phase 3: natives
for tmp in mod.natives:
    _traced_add_syms(label, all_syms_raw, lu.used_symbols_asts(tmp.args[2:]))
_trace_sym_set("isolate.allSyms_post_natives", all_syms_raw)

# Normalize
all_syms = set()
for sym in all_syms_raw:
    nsym = ivy_logic.normalize_symbol(sym)
    if __debug__:
        if nsym not in all_syms:
            xtracer.trace("isolate.allSyms.add_normalized %s" % str(nsym))
    all_syms.add(nsym)
_trace_sym_set("isolate.allSyms_post_normalize", all_syms)
```

Keep the existing `_trace_sym_set` summary traces -- they don't interfere and provide useful context. The online `.add` traces fire during collection; the summaries fire after.

### Step 8: Update callers of `collectSymbolsInto`

**File:** `goivy/isolate/isolate.go` (and any other callers)

Every call to `collectSymbolsInto(node, syms)` must be updated to `collectSymbolsInto(label, node, syms)`. Search for all callers and add the appropriate label string. Callers outside of `IsolateComponent()` (e.g., in enforce_axioms check at line ~1100) should use descriptive labels like `"isolate.droppedAxiomSyms"`.

## Files to Modify

| File | Change |
|------|--------|
| `goivy/isolate/helpers.go` | Add `label` param to `collectSymbolsInto`, inline trace on new additions |
| `goivy/isolate/isolate.go` | Restructure allSyms2 order (Step 4), inline traces on all direct insertions (Steps 2, 6), update all `collectSymbolsInto` calls with labels (Step 8) |
| `pyivy/ivy/ivy/ivy_isolate.py` | Add `_traced_add_syms` (Step 3), restructure allSyms2 (Step 5), restructure allSyms (Step 7) |

## Verification

1. Build Go with xtracer enabled, run test that triggers `isolate_component`
2. Run same test on Python with `__debug__` on (default)
3. Diff xtracer output line-by-line
4. The FIRST `.add` line that differs between Go and Python pinpoints exactly which symbol diverged, in which phase, at the moment it happened
