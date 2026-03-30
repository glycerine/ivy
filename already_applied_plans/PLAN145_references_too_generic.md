# Plan: Fix allSyms Divergence — Go Over-collects Assignment LHS Symbols

**Created**: 2026-03-30 14:30

## Context

The golden test (TestOrdLive) diverges at line 150788 of XTRACE output:

```
150788  go : XTRACE: isolate.allSyms_pre_follow.sym arm.cpl_fair
        py : XTRACE: isolate.allSyms_pre_follow.sym arm_l
```

Go has the extra symbol `arm.cpl_fair` in `allSyms` that Python does not. This causes all subsequent symbols to be misaligned.

## Root Cause Analysis

### Primary Bug: `referencesRec` is Generic — Missing Specialized `references()` per Action Type

**Python** has **specialized** `references()` methods for assignment-like actions:

1. **`AssignAction.references()`** (`ivy_actions.py:496`):
   ```python
   def references(self, refs):
       refs.update(symbols_ast(self.args[1]))   # RHS only
       assign_refs(self, refs)                    # selective LHS
   ```

2. **`HavocAction.references()`** (`ivy_actions.py:675`):
   ```python
   def references(self, refs):
       assign_refs(self, refs)                    # selective LHS only
   ```

3. **`assign_refs`** (`ivy_actions.py:470`): Only adds the LHS target symbol if it's a **destructor sort**. For bare symbols (non-destructors), it processes the args (which are `[]` for a Const), adding **nothing**.

**Go's** `referencesRec` (`actions/transforms.go:297`) is completely generic — it calls `collectSymbols` on ALL `ActionArgs()` for ALL action types, including the LHS of assignments. For `AssignAction`, this means BOTH `LHS` and `RHS` get symbol-collected, adding the LHS target symbol unconditionally.

**Result**: `arm.cpl_fair` is the LHS of an assignment/havoc in `new_actions`. Go adds it to `allSyms`; Python doesn't.

### Secondary Bug #1: Missing SchemaBody Filter

**Python** (`ivy_isolate.py:1220`): `if not isinstance(y.formula, ivy_ast.SchemaBody)` — skips SchemaBody formulas.

**Go** (`isolate/isolate.go:908-912`): No such filter. Collects symbols from ALL formulas including SchemaBody. Does not affect current test (SchemaBody likely doesn't implement `lg.Expr` so the type assertion would panic if present), but is a correctness issue.

### Secondary Bug #2: Natives Collection Broken

**Python** (`ivy_isolate.py:1225-1226`): `asts.extend(tmp.args[2:])` — processes NativeDef args from index 2 onward.

**Go** (`isolate/isolate.go:924-928`): Tries `nat.(*ast.LabeledFormula)` — but natives are `*ast.NativeDef`, NOT `LabeledFormula`. The type assertion always fails, so Go collects **nothing** from natives. This under-collection may be masked by over-collection from the primary bug, but will cause divergences once the primary bug is fixed.

## Fix Plan

### Fix 1 (Primary): Specialized `referencesRec` for AssignAction and HavocAction

**File**: `actions/transforms.go`

Change `referencesRec` to dispatch on action type:

```go
func referencesRec(action Action, result map[lg.NodeKey]lg.Expr, destructorSorts map[string]lg.Sort) {
    if action == nil {
        return
    }
    switch a := action.(type) {
    case *AssignAction:
        // Python AssignAction.references: only RHS + assign_refs(LHS)
        collectSymbols(a.RHS, result)
        assignRefs(a.LHS, result, destructorSorts)
    case *HavocAction:
        // Python HavocAction.references: only assign_refs(target)
        assignRefs(a.Target, result, destructorSorts)
    default:
        // Base Action.references: process all non-Action args
        for _, arg := range action.ActionArgs() {
            if _, isAct := arg.(Action); !isAct && arg != nil {
                collectSymbols(arg, result)
            }
        }
    }
    // Python get_references: recurse into Action children
    for _, arg := range action.ActionArgs() {
        if child, ok := arg.(Action); ok {
            referencesRec(child, result, destructorSorts)
        }
    }
}
```

Add `assignRefs` function matching Python's `assign_refs` (`ivy_actions.py:470-480`):

```go
// assignRefs matches Python's assign_refs (ivy_actions.py:470-480).
// For destructor chains, adds the destructor symbol, recurses into args[0],
// and collects from remaining args. For non-destructors, processes children
// only (NOT the target symbol itself).
func assignRefs(node lg.Expr, result map[lg.NodeKey]lg.Expr, destructorSorts map[string]lg.Sort) {
    if node == nil {
        return
    }
    if app, ok := node.(*lg.Apply); ok {
        if c, ok := app.Func.(*lg.Const); ok {
            if _, isDestructor := destructorSorts[c.Name]; isDestructor {
                result[ConstSymKey(c)] = c
                if len(app.Terms) > 0 {
                    assignRefs(app.Terms[0], result, destructorSorts)
                }
                for i := 1; i < len(app.Terms); i++ {
                    collectSymbols(app.Terms[i], result)
                }
                return
            }
        }
    }
    // Not a destructor: process children only (for bare Const, Children()=nil → nothing added)
    for _, child := range node.Children() {
        collectSymbols(child, result)
    }
}
```

Update `GetReferencesInto` and `References` signatures:

```go
func GetReferencesInto(action Action, syms map[lg.NodeKey]lg.Expr, destructorSorts map[string]lg.Sort) {
    referencesRec(action, syms, destructorSorts)
}

func References(action Action, destructorSorts map[string]lg.Sort) map[lg.NodeKey]lg.Expr {
    result := make(map[lg.NodeKey]lg.Expr)
    referencesRec(action, result, destructorSorts)
    return result
}
```

### Fix 2: Update Callers of GetReferencesInto

**File**: `isolate/isolate.go`

At line 938 and 1237, pass `mod.DestructorSorts`:

```go
actions.GetReferencesInto(act, allSyms, mod.DestructorSorts)
```

### Fix 3: SchemaBody Filter

**File**: `isolate/isolate.go`, lines 908-912

Add filter matching Python's `if not isinstance(y.formula, ivy_ast.SchemaBody)`:

```go
for _, lf := range lfSlice {
    if lf.Formula != nil {
        if _, isSchema := lf.Formula.(*ast.SchemaBody); isSchema {
            continue
        }
        collectUsedSymbolNames(lf.Formula.(lg.Expr), allSyms)
    }
}
```

### Fix 4: Natives Collection

**File**: `isolate/isolate.go`, lines 923-928

Replace broken `LabeledFormula` cast with correct `NativeDef`/Node args processing:

```go
// Collect from natives — Python: asts.extend(tmp.args[2:])
for _, nat := range mod.Natives {
    args := nat.Args()
    for i := 2; i < len(args); i++ {
        if expr, ok := args[i].(lg.Expr); ok {
            collectUsedSymbolNames(expr, allSyms)
        }
    }
}
```

## Files to Modify

1. **`actions/transforms.go`** — Fixes 1: `referencesRec` specialization, `assignRefs`, signature changes
2. **`isolate/isolate.go`** — Fixes 2, 3, 4: caller updates, SchemaBody filter, natives collection

## Fix 5: Data-level XTRACE Tracing for Symbol Collection Conformance

Beyond matching call paths, we need to trace the **actual data** flowing through symbol collection to verify Go/Python conformance.

### 5a. Per-source allSyms contributions

Trace what each source contributes BEFORE merging into allSyms. This identifies which source adds unexpected symbols.

**Go** (`isolate/isolate.go`, after each collection block):

```go
// After formula collection (line ~912):
xtracer.Trace("isolate.allSyms_src_formulas n=%d", len(allSyms))

// After formal params (line ~922), compute delta:
xtracer.Trace("isolate.allSyms_src_formals n=%d", len(allSyms))

// After natives (line ~928):
xtracer.Trace("isolate.allSyms_src_natives n=%d", len(allSyms))

// After normalization (line ~934):
xtracer.Trace("isolate.allSyms_src_normalized n=%d", len(allSyms))

// After action references (line ~939):
xtracer.Trace("isolate.allSyms_src_action_refs n=%d", len(allSyms))

// After proof definitions (line ~974):
xtracer.Trace("isolate.allSyms_src_proof_defs n=%d", len(allSyms))
```

**Python** (`ivy_isolate.py`, corresponding locations):

```python
# After used_symbols_asts but before action refs:
if __debug__: xtracer.trace("isolate.allSyms_src_formulas_formals_natives n=%d" % len(all_syms))

# After action references:
if __debug__: xtracer.trace("isolate.allSyms_src_action_refs n=%d" % len(all_syms))

# After proof definitions:
if __debug__: xtracer.trace("isolate.allSyms_src_proof_defs n=%d" % len(all_syms))
```

Note: Python collects formulas+formals+natives in one `used_symbols_asts` call, so it can't easily split them. But the cumulative counts after each stage suffice to pinpoint where Go diverges.

### 5b. Per-action references() results

Trace the symbols returned by each action's `references()` call.

**Go** (inside the action references loop, after `GetReferencesInto`):

```go
for name, act := range newActions.All() {
    before := len(allSyms)
    actions.GetReferencesInto(act, allSyms, mod.DestructorSorts)
    if len(allSyms) > before {
        added := make([]string, 0)
        // Collect newly added symbols
        for k, v := range allSyms {
            // (track delta by comparing with snapshot)
        }
        xtracer.Trace("isolate.action_refs actname=%s added=%d", name, len(allSyms)-before)
    }
}
```

Actually, a simpler approach — trace per-action reference sets separately, then add:

**Go** (`actions/transforms.go`, at end of `referencesRec` or in `GetReferencesInto`):

After the top-level `GetReferencesInto` call completes for each action, trace the delta.

**Python** (`ivy_actions.py`, in `get_references`):

```python
def get_references(self, refs):
    before = len(refs)
    self.references(refs)
    for a in self.args:
        if isinstance(a, Action):
            a.get_references(refs)
    if __debug__ and len(refs) > before:
        xtracer.trace("actions.get_references type=%s added=%d total=%d" %
            (type(self).__name__, len(refs) - before, len(refs)))
```

### 5c. assignRefs results

Trace what `assignRefs` adds for each assignment/havoc LHS.

**Go** (`actions/transforms.go`, in the new `assignRefs`):

```go
func assignRefs(node lg.Expr, result map[lg.NodeKey]lg.Expr, destructorSorts map[string]lg.Sort) {
    before := len(result)
    assignRefsInner(node, result, destructorSorts)
    if len(result) > before {
        xtracer.Trace("actions.assignRefs node=%v added=%d", node, len(result)-before)
    }
}
```

**Python** (`ivy_actions.py`, in `assign_refs`):

```python
def assign_refs(self, refs):
    before = len(refs)
    def recur(n):
        ...
    recur(self.args[0])
    if __debug__ and len(refs) > before:
        xtracer.trace("actions.assign_refs added=%d" % (len(refs) - before))
```

### 5d. collectSymbols / symbols_ast per-call results (targeted)

Only add this for the specific collection points that feed allSyms (not deep recursion), to avoid trace explosion:

**Go** (in `isolate/isolate.go`, wrap `collectUsedSymbolNames` calls):

```go
func collectUsedSymbolNamesTraced(label string, node lg.Expr, syms map[lg.NodeKey]lg.Expr) {
    before := len(syms)
    collectUsedSymbolNames(node, syms)
    if len(syms) > before {
        xtracer.Trace("isolate.collectUsedSymbolNames %s added=%d total=%d", label, len(syms)-before, len(syms))
    }
}
```

## Verification

1. `go build ./...` — ensure compilation
2. `cd ~/goivy && make golden` — run golden test
3. Examine `~/goivy/log.red` — check if `arm.cpl_fair` divergence is resolved
4. The new per-source traces should confirm that cumulative counts match between Go and Python at each stage
5. If new divergences appear (likely from natives under-collection being unmasked), the per-source traces will immediately pinpoint which collection source is responsible
