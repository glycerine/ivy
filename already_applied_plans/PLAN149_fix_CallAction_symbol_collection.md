# Plan: Fix CallAction Symbol Collection Divergence + Add XTRACE Instrumentation

**Created**: 2026-03-31 00:15

## Context

The golden test (`make golden`) diverges at line 150847 of the xtrace output:
- Go: `isolate.allSyms_post_action_refs.sym ext:arm.issue_hook`
- Python: `isolate.allSyms_post_action_refs.sym fml:a`

The symbol sets collected during the action-refs phase of isolate processing differ between Go and Python. Root cause: Go's `referencesRec` handles `CallAction` via the `default` case, which incorrectly collects the callee name as a symbol and fails to collect symbols from actual parameters.

## Root Cause Analysis

### How Python works

Python `CallAction.args = [Atom("ext:name", [actual_params...]), ret1, ret2, ...]`

The base `Action.references()` (ivy_actions.py:287-290) iterates `self.args`:
```python
def references(self, refs):
    for a in self.args:
        if not isinstance(a, Action):
            refs.update(symbols_ast(a))
```

For `args[0]` (an Atom): `symbols_ast(Atom)` does NOT yield the Atom's rep (because `is_app(Atom)` is False — Atom is not an App/Const/NamedBinder). It only recurses into `Atom.args` (the actual parameters) and yields symbols from them.

Result: Python collects symbols from actual parameters and actual returns, but NOT the callee name.

### How Go works (BUG)

Go `CallAction.ActionArgs()` returns `[Callee(lg.Const), ...ActualReturns]`

The `default` case in `referencesRec` (transforms.go:318-327) calls `collectSymbols` on each non-Action arg:
1. `collectSymbols(Callee)` — Callee is a bare `*lg.Const` with the action name → gets ADDED to symbols (**BUG: extra symbol**)
2. `collectSymbols(ret_i)` — symbols from returns collected correctly
3. Actual parameters live on `AstCallee.Terms` — NOT in `ActionArgs()` → **NOT collected** (**BUG: missing symbols**)

### Two bugs

| # | Bug | Effect |
|---|-----|--------|
| 1 | Callee Const added as symbol | Go has extra symbols like `ext:arm.issue_hook` |
| 2 | Actual params on AstCallee.Terms not collected | Go may miss symbols from call arguments |

## Fix: Add `case *CallAction` to `referencesRec`

**File**: `actions/transforms.go:307` (inside the switch in `referencesRec`)

Add a new case before `default:`:

```go
case *CallAction:
    // Python: CallAction uses base Action.references() (no override).
    // self.args[0] is an Atom. symbols_ast(Atom) does NOT yield the
    // atom's rep (is_app(Atom) is False) — only recurses into Atom.args
    // (the actual parameters). self.args[1:] are actual returns.
    //
    // Go: Callee is a bare Const (the action name) — must NOT be added.
    // AstCallee.Terms holds the actual parameters — must be collected.
    // ActualReturns are the outputs.
    if a.AstCallee != nil {
        for _, term := range a.AstCallee.Terms {
            if expr, ok := term.(lg.Expr); ok {
                collectSymbols(expr, result)
            }
        }
    }
    for _, ret := range a.ActualReturns {
        collectSymbols(ret, result)
    }
```

This exactly matches the Python behavior:
- `symbols_ast(Atom.args[i])` → `collectSymbols(AstCallee.Terms[i])`
- `symbols_ast(self.args[1:])` → `collectSymbols(ActualReturns[i])`
- Callee name NOT collected (matches `is_app(Atom)` being False)

## XTRACE Instrumentation: Catching Divergences Earlier

The user wants comprehensive tracing so that future divergences are caught close to the source, not 150K lines downstream. Three layers of instrumentation:

### Layer 1: Per-phase allSyms tracing in isolate.go

Currently only two checkpoints: `post_normalize` and `post_action_refs`. Add per-phase traces AND per-action traces.

**File**: `isolate/isolate.go` (lines 900-952)

```go
// After formulas phase (line ~917):
if xtracer.Enabled {
    traceSymSet("isolate.allSyms_post_formulas", allSyms)
}

// After formals phase (line ~926):
if xtracer.Enabled {
    traceSymSet("isolate.allSyms_post_formals", allSyms)
}

// After natives phase (line ~935):
if xtracer.Enabled {
    traceSymSet("isolate.allSyms_post_natives", allSyms)
}

// Per-action trace in the action refs loop (line ~947-949):
for _, act := range newActions.All() {
    sizeBefore := len(allSyms)
    actions.GetReferencesInto(act, allSyms, mod.DestructorSorts)
    if xtracer.Enabled && len(allSyms) != sizeBefore {
        xtracer.Trace("isolate.allSyms_action_refs.%s added=%d total=%d",
            actionTypeName(act), len(allSyms)-sizeBefore, len(allSyms))
    }
}
```

**File**: `ivy_isolate.py` (lines 1225-1239) — matching Python traces:

```python
# After formulas:
_trace_sym_set("isolate.allSyms_post_formulas", all_syms_pre_formals)

# After formals:
_trace_sym_set("isolate.allSyms_post_formals", all_syms_pre_natives)

# After natives:
_trace_sym_set("isolate.allSyms_post_natives", all_syms_pre_normalize)

# Per-action in loop:
for action in list(new_actions.values()):
    before = len(all_syms)
    action.get_references(all_syms)
    if __debug__ and len(all_syms) != before:
        xtracer.trace("isolate.allSyms_action_refs.%s added=%d total=%d" %
            (type(action).__name__, len(all_syms) - before, len(all_syms)))
```

### Layer 2: Per-symbol addition tracing in referencesRec / symbols_ast

Trace each symbol as it's added, with the source action type and expression context.

**File**: `actions/transforms.go`

Add a `traceCollect` helper:
```go
func traceCollect(c *lg.Const, source string, result map[lg.NodeKey]lg.Expr) {
    key := ConstSymKey(c)
    if _, already := result[key]; !already {
        xtracer.Trace("actions.collectSymbols.add %s src=%s", ConstSymDisplay(c), source)
    }
    result[key] = c
}
```

Then modify `collectSymbols`, `assignRefs`, and the new CallAction case to call `traceCollect` when `xtracer.Enabled`.

**File**: `ivy_logic_utils.py` — matching Python trace in `symbols_ast`:
```python
def symbols_ast(ast, _source=""):
    if is_app(ast):
        if is_binder(ast.rep):
            for x in symbols_ast(ast.rep.body, _source):
                yield x
        else:
            if __debug__: xtracer.trace("symbols_ast.yield %s src=%s" % (str(ast.rep), _source))
            yield ast.rep
    for arg in ast.args:
        for x in symbols_ast(arg, _source):
            yield x
```

### Layer 3: Data-in / data-out tracing for key functions

**File**: `actions/transforms.go`

```go
func referencesRec(action Action, result map[lg.NodeKey]lg.Expr, ...) {
    if xtracer.Enabled {
        xtracer.Trace("actions.referencesRec ENTER type=%s n_before=%d",
            actionTypeName(action), len(result))
    }
    // ... existing logic ...
    if xtracer.Enabled {
        xtracer.Trace("actions.referencesRec EXIT type=%s n_after=%d",
            actionTypeName(action), len(result))
    }
}
```

**File**: `ivy_actions.py`

```python
def get_references(self, refs):
    if __debug__:
        xtracer.trace("actions.get_references ENTER type=%s n_before=%d" %
            (type(self).__name__, len(refs)))
    self.references(refs)
    for a in self.args:
        if isinstance(a, Action):
            a.get_references(refs)
    if __debug__:
        xtracer.trace("actions.get_references EXIT type=%s n_after=%d" %
            (type(self).__name__, len(refs)))
```

### Helper needed

**File**: `actions/transforms.go` — add `actionTypeName`:
```go
func actionTypeName(a Action) string {
    if a == nil { return "nil" }
    return reflect.TypeOf(a).Elem().Name()
}
```

Or use the existing `Name()` method on Action.

## Files Modified

| File | Changes |
|------|---------|
| `actions/transforms.go` | Add `case *CallAction` to `referencesRec`; add trace helpers; instrument `collectSymbols`, `referencesRec` |
| `isolate/isolate.go` | Add per-phase and per-action allSyms traces |
| `~/pyivy/ivy/ivy/ivy_isolate.py` | Add matching per-phase and per-action traces |
| `~/pyivy/ivy/ivy/ivy_actions.py` | Add data-in/out traces to `get_references` |
| `~/pyivy/ivy/ivy/ivy_logic_utils.py` | Add per-symbol trace to `symbols_ast` |

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...` — must compile
2. `cd ~/go/src/github.com/glycerine/goivy && go test ./actions/... ./isolate/...` — tests pass
3. `cd ~/goivy && make golden` — divergence at line 150847 resolved; traces match further
4. `cd ~/goivy && make test` — full test suite passes
