# Plan: Refactor IsolateComponent() for Online Per-Symbol Tracing

**Created:** 2026-03-31 ~afternoon

## Context

`IsolateComponent()` in `goivy/isolate/isolate.go:266` is the Go port of `isolate_component()` in `pyivy/ivy/ivy/ivy_isolate.py:919`. The Go port has divergences from Python's execution flow in symbol collection (allSyms/allSyms2). A divergence was observed in `isolate.allSyms2.sym` traces. The goal is to add **online, per-symbol** xtracer tracing so that divergences are detected immediately when they happen -- not via batch dumps after the fact. Both sides must emit the same trace for each new symbol as it is added, in the same order, so a line-by-line diff of xtracer output pinpoints the exact first divergence.

## Divergence: allSyms2 collection ordering

**Python** (ivy_isolate.py:1363-1382) processes in this order:
1. Formulas from `[axioms, props, inits, conjs, definitions]`
2. Action bodies: `asts.extend(action for action in mod.actions.values())`
3. Params: `asts.extend(mod.params)` (if keep_destructors)
4. Action formals: separate loop over `mod.actions.values()` for `formal_params` + `formal_returns`
5. Natives: `asts.extend(tmp.args[2:])`
6. Proofs: `asts.extend(x[1] for x in mod.proofs)`
Then `all_syms = set(lu.used_symbols_asts(asts))` walks every AST via DFS, yielding symbols.

**Go** (isolate.go:1278-1335) processes DIFFERENTLY:
1. Formulas (same)
2. **Action formals AND bodies INTERLEAVED** per-action (formals first, then body)
3. Params
4. Natives
5. Proofs

Go must be restructured to match Python's exact order so that the online per-symbol traces are emitted in the same sequence.

## Implementation Steps

### Step 1: Add online tracing helpers to Go

**File:** `goivy/isolate/helpers.go`

```go
// collectSymbolsIntoTraced walks node via SymbolsIluAst and adds each symbol
// to syms. For each NEW symbol (not already in syms), emits an xtracer line.
// This is the "online" traced variant of collectSymbolsInto.
func collectSymbolsIntoTraced(label string, node lg.Expr, syms map[lg.NodeKey]lg.Expr) {
    if node == nil {
        return
    }
    for sym := range il.SymbolsIluAst(node) {
        key := lg.Key(sym)
        if _, exists := syms[key]; !exists {
            xtracer.Trace("%s.add %s", label, lg.PrettyFmla(sym))
        }
        syms[key] = sym
    }
}

// tracedSymAdd inserts a single symbol into syms, tracing if new.
func tracedSymAdd(label string, syms map[lg.NodeKey]lg.Expr, sym lg.Expr) {
    key := lg.Key(sym)
    if _, exists := syms[key]; !exists {
        xtracer.Trace("%s.add %s", label, lg.PrettyFmla(sym))
    }
    syms[key] = sym
}
```

When `xtracer.Enabled` is false (compile-time const), the Go compiler will dead-code-eliminate the traced path. Use pattern:
```go
if xtracer.Enabled {
    collectSymbolsIntoTraced(label, node, syms)
} else {
    collectSymbolsInto(node, syms)
}
```

### Step 2: Add online tracing helper to Python

**File:** `pyivy/ivy/ivy/ivy_isolate.py` (near line 36, next to `_trace_sym_set`)

```python
def _traced_add_syms(label, target_set, new_syms):
    """Add symbols to set, tracing each NEW addition immediately."""
    if __debug__:
        for sym in new_syms:
            if sym not in target_set:
                xtracer.trace("%s.add %s" % (label, str(sym)))
            target_set.add(sym)
    else:
        target_set.update(new_syms)
```

Note: `used_symbols_asts` returns symbols in DFS traversal order of the AST list. By iterating the generator (not collecting to `set()` first), we preserve the order and can trace each addition online.

### Step 3: Restructure Go allSyms2 to match Python order + online traces

**File:** `goivy/isolate/isolate.go`, lines ~1278-1335

Replace the current allSyms2 block. The label for all allSyms2 additions is `"isolate.allSyms2"` so traces read: `XTRACE: isolate.allSyms2.add <symbol>`.

```go
allSyms2 := make(map[lg.NodeKey]lg.Expr)
const as2Label = "isolate.allSyms2"

// Phase A: formulas (axioms, props, inits, conjs, definitions)
for _, lfSlice := range [][]*ast.LabeledFormula{
    mod.LabeledAxioms, mod.LabeledProps, mod.LabeledInits, mod.LabeledConjs, mod.Definitions,
} {
    for _, lf := range lfSlice {
        if lf.Formula != nil {
            if _, isSchema := lf.Formula.(*ast.SchemaBody); isSchema {
                continue
            }
            if xtracer.Enabled {
                collectSymbolsIntoTraced(as2Label, lf.Formula.(lg.Expr), allSyms2)
            } else {
                collectSymbolsInto(lf.Formula.(lg.Expr), allSyms2)
            }
        }
    }
}

// Phase B: action bodies ONLY (not formals -- Python does them separately in Phase D)
for _, act := range mod.Actions.All() {
    if xtracer.Enabled {
        collectSymbolsIntoTraced(as2Label, act, allSyms2)
    } else {
        collectSymbolsInto(act, allSyms2)
    }
}

// Phase C: params (if keep_destructors)
if isoCfg.KeepDestructors {
    for _, p := range mod.Params {
        if xtracer.Enabled {
            tracedSymAdd(as2Label, allSyms2, p)
        } else {
            allSyms2[actions.ConstSymKey(p)] = p
        }
    }
}

// Phase D: action formals (separate pass, matching Python)
for _, act := range mod.Actions.All() {
    for _, p := range act.GetFormalParams() {
        if xtracer.Enabled {
            tracedSymAdd(as2Label, allSyms2, p)
        } else {
            allSyms2[actions.ConstSymKey(p)] = p
        }
    }
    for _, r := range act.GetFormalReturns() {
        if xtracer.Enabled {
            tracedSymAdd(as2Label, allSyms2, r)
        } else {
            allSyms2[actions.ConstSymKey(r)] = r
        }
    }
}

// Phase E: natives
for _, nat := range mod.Natives {
    args := nat.Args()
    for i := 2; i < len(args); i++ {
        if expr, ok := args[i].(lg.Expr); ok {
            if xtracer.Enabled {
                collectSymbolsIntoTraced(as2Label, expr, allSyms2)
            } else {
                collectSymbolsInto(expr, allSyms2)
            }
        }
    }
}

// Phase F: proofs
for _, pe := range mod.Proofs {
    if pe.Proof != nil {
        if n, ok := pe.Proof.(lg.Expr); ok {
            if xtracer.Enabled {
                collectSymbolsIntoTraced(as2Label, n, allSyms2)
            } else {
                collectSymbolsInto(n, allSyms2)
            }
        }
    }
}
```

### Step 4: Add online traces to Python allSyms2

**File:** `pyivy/ivy/ivy/ivy_isolate.py`, lines ~1363-1382

Replace the current flat-list-then-set approach with online tracing while preserving the same collection order and semantics:

```python
# Build asts list exactly as before (same order), but trace online
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

**Critical note on Python:** The original code calls `lu.used_symbols_asts(asts)` on the entire flat list. `used_symbols_asts = gen_to_set(symbols_clause)` where `symbols_clause = apply_gen_to_list(symbols_ilu_ast)`. This means `used_symbols_asts([a,b,c])` = `set(chain(symbols_ilu_ast(a), symbols_ilu_ast(b), symbols_ilu_ast(c)))`. Our refactored code calls `used_symbols_ast(single_ast)` per-formula (which walks that one AST), yielding the same symbols in the same DFS order, just one formula at a time. The final set is identical, but now we can trace each addition online.

We use `used_symbols_ast` (singular) for single ASTs and `used_symbols_asts` (plural) for lists. Both are defined in ivy_logic_utils.py:600-601.

### Step 5: Add online traces to Go allSyms (first collection)

**File:** `goivy/isolate/isolate.go`, lines ~918-987

Add online tracing with label `"isolate.allSyms"`:

- Formula collection (line ~932): wrap with `collectSymbolsIntoTraced("isolate.allSyms", ...)`
- Formal collection (line ~940): wrap with `tracedSymAdd("isolate.allSyms", ...)`
- Native collection (line ~952): wrap with `collectSymbolsIntoTraced("isolate.allSyms", ...)`
- Action refs (line ~974): `GetReferencesInto` already traces ENTER/EXIT; additionally trace each new addition inside `referencesRec` or wrap at call site
- Proof additions (line ~1029): wrap with `tracedSymAdd("isolate.allSyms", ...)`

Keep existing phase-level `traceSymSet` calls as-is (they remain useful for summary). The online `.add` traces fire DURING collection; the existing batch dumps fire AFTER.

### Step 6: Add online traces to Python allSyms (first collection)

**File:** `pyivy/ivy/ivy/ivy_isolate.py`, lines ~1231-1260

Replace `all_syms_raw = set(lu.used_symbols_asts(formula_asts))` and `.update()` calls with `_traced_add_syms("isolate.allSyms", ...)`:

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
    if nsym not in all_syms:
        if __debug__: xtracer.trace("isolate.allSyms.add_normalized %s" % str(nsym))
    all_syms.add(nsym)
_trace_sym_set("isolate.allSyms_post_normalize", all_syms)
```

### Step 7: Verify allSyms action refs tracing matches

Both sides already trace `actions.referencesRec ENTER/EXIT`. For online per-symbol matching, we also need per-addition traces inside `referencesRec`.

**Go** (`actions/transforms.go:302`): The `collectSymbols` calls inside `referencesRec` add to `result` without tracing. Add a traced variant or pass a label through. Since `referencesRec` is used from `GetReferencesInto` (called from isolate.go), threading a label is clean:

- Add `label string` parameter to `referencesRec` and `collectSymbols` (or traced wrappers)
- At each `collectSymbols(expr, result)` call, use `collectSymbolsTraced(label, expr, result)` when `xtracer.Enabled`

**Python**: `action.get_references(all_syms)` calls `self.references(refs)` which calls `refs.update(symbols_ast(...))`. To trace online, override or patch `references()` to use `_traced_add_syms` instead of `refs.update()`. Alternatively, wrap at the call site by checking set size changes per-symbol.

This step is the most intrusive. If the allSyms divergence is in allSyms2 (as observed), this step can be deferred.

## Files to Modify

| File | Change |
|------|--------|
| `goivy/isolate/helpers.go` | Add `collectSymbolsIntoTraced()`, `tracedSymAdd()` |
| `goivy/isolate/isolate.go` | Restructure allSyms2 order (Step 3), add online traces to allSyms (Step 5) |
| `pyivy/ivy/ivy/ivy_isolate.py` | Add `_traced_add_syms()` (Step 2), restructure allSyms2 (Step 4), add online traces to allSyms (Step 6) |
| `goivy/actions/transforms.go` | (Step 7, deferrable) Add traced variants for referencesRec symbol collection |

## Verification

1. Build Go with xtracer enabled, run test that triggers `isolate_component`
2. Run same test on Python side with `__debug__` enabled
3. Diff the xtracer output line-by-line
4. The FIRST line that differs between Go and Python pinpoints exactly which symbol addition diverged and in which phase
5. No guessing, no batch comparison -- the divergence is caught at the moment it happens
