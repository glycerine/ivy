# Plan: Fix SomeCondition.Children() Missing Params in Symbol Collection

**Created**: 2026-03-30 22:30

## Context

The golden test (`make golden`) diverges at line 151251 inside 3 nested `if` actions during the action-refs phase of isolate `allSyms` collection:

```
151250  go : XTRACE: actions.referencesRec ENTER type=if n_before=80
        py : XTRACE: actions.referencesRec ENTER type=if n_before=80

151251  go : XTRACE: actions.collectSymbols.add isd.rd
        py : XTRACE: actions.collectSymbols.add loc:tt
```

Go yields `isd.rd` (a symbol from the formula body) while Python yields `loc:tt` (a bound variable of a `some` quantifier). The innermost `if` has an existential condition (`if some tt:lclock. ...`).

## Root Cause

**`SomeCondition.Children()` only returns `[Fmla]`, omitting `Params` and `Index`.**

### Python path

When Python's `Action.references()` calls `symbols_ast(self.args[0])` on a compiled `Some` node:

1. `is_app(Some)` returns False (Some is not App/Const/NamedBinder)
2. Iterates `Some.args` = `[*compiled_params, compiled_fmla, ?compiled_index]`
3. For each param (`lg.Const` like `loc:tt`): `is_app(Const)` is True, yields the Const
4. For the formula: recurses, eventually yielding `isd.rd` and other symbols

**Result**: Python adds `loc:tt` (param) BEFORE `isd.rd` (formula symbol).

### Go path (BUG)

When Go's `referencesRec` default case calls `collectSymbols(IfAction.Cond)` where Cond is a `SomeCondition`:

1. Not `*lg.Const`, not `*lg.Apply` → skip
2. Iterates `SomeCondition.Children()` = `[Fmla]` only (line 410 of action.go)
3. Recurses into formula, yielding `isd.rd` first

**Result**: Go never visits the params. Skips `loc:tt` entirely, yields `isd.rd` first.

### Why this wasn't caught by PLAN138

PLAN138 fixed the same structural problem for `substitute_constants_action` by adding `AstCond` to `IfAction.Args()`. That fix routes transforms through the AST node (which has proper children). But `referencesRec` uses `ActionArgs()` → `Cond` (the SomeCondition itself), and `collectSymbols` traverses via `Children()`, which was never fixed.

## Fix

**File**: `actions/action.go:410`

Change `SomeCondition.Children()` from:

```go
func (s *SomeCondition) Children() []lg.Expr  { return []lg.Expr{s.Fmla} }
```

To:

```go
func (s *SomeCondition) Children() []lg.Expr {
	// Python: Some.args = [*params, fmla] or [*params, fmla, index]
	// Must include params so collectSymbols/symbols_ast yields bound
	// variables, matching Python's traversal of Some.args.
	result := make([]lg.Expr, 0, len(s.Params)+2)
	for _, p := range s.Params {
		result = append(result, p)
	}
	result = append(result, s.Fmla)
	if s.Index != nil {
		result = append(result, s.Index)
	}
	return result
}
```

This matches Python's `Some.args = [*params, fmla, ?index]` layout.

### Safety of the change

`Children()` is used by:
- `collectSymbols` (transforms.go:465) — **this is the bug**; fix makes it match Python
- `assignRefs` (transforms.go:396) — only called on LHS nodes of assignments, never SomeCondition
- `typeCheckRec` (phase3.go:139) — would now also check params; safe since params are valid compiled Consts
- `collectSymbolNames` (update.go:2055) — would now include params in name collection; matches Python
- `IfAction.Children()` delegates to `ActionArgs()` which returns `Cond`, not `SomeCondition.Children()` directly

No code assumes `SomeCondition.Children()` has exactly 1 element. The methods `Subactions`, `Decompose`, `GetCond`, `Sexp` all access fields directly, not via `Children()`.

## Additional XTRACE instrumentation

To make future divergences in the action-refs phase easier to pinpoint:

### 1. Add BEGIN/END markers per action in isolate loop

**File**: `isolate/isolate.go` (around line 956)

```go
for actname, act := range newActions.All() {
	if xtracer.Enabled {
		xtracer.Trace("isolate.allSyms_action_refs.BEGIN %s", actname)
	}
	sizeBefore := len(allSyms)
	actions.GetReferencesInto(act, allSyms, mod.DestructorSorts)
	if xtracer.Enabled {
		xtracer.Trace("isolate.allSyms_action_refs.END %s added=%d total=%d",
			actname, len(allSyms)-sizeBefore, len(allSyms))
	}
}
```

**File**: `~/pyivy/ivy/ivy/ivy_isolate.py` (matching Python trace in the action refs loop)

```python
for actname, action in list(new_actions.items()):
    if __debug__: xtracer.trace("isolate.allSyms_action_refs.BEGIN %s" % actname)
    size_before = len(all_syms)
    action.get_references(all_syms)
    if __debug__: xtracer.trace("isolate.allSyms_action_refs.END %s added=%d total=%d" %
        (actname, len(all_syms) - size_before, len(all_syms)))
```

### 2. Trace IfAction condition type in referencesRec

**File**: `actions/transforms.go` (inside the default case, before collectSymbols)

Add a trace when processing IfAction conditions to show whether it's a SomeCondition and what kind:

```go
default:
	for _, arg := range action.ActionArgs() {
		if _, isAct := arg.(Action); !isAct && arg != nil {
			if xtracer.Enabled {
				if ifAct, ok := action.(*IfAction); ok && arg == ifAct.Cond {
					if sc, ok := ifAct.Cond.(*SomeCondition); ok {
						xtracer.Trace("actions.referencesRec.if_cond kind=%s nparams=%d", sc.Kind, len(sc.Params))
					} else {
						xtracer.Trace("actions.referencesRec.if_cond kind=plain")
					}
				}
			}
			collectSymbols(arg, result)
		}
	}
```

**File**: `~/pyivy/ivy/ivy/ivy_actions.py` (in base `Action.references()`)

```python
def references(self, refs):
    for a in self.args:
        if not isinstance(a, Action):
            if __debug__ and isinstance(self, IfAction) and a is self.args[0]:
                from . import ivy_ast
                if isinstance(a, (ivy_ast.Some, ivy_ast.SomeMinMax)):
                    xtracer.trace("actions.referencesRec.if_cond kind=%s nparams=%d" %
                        (type(a).__name__.lower(), len(a.args) - 1))
                else:
                    xtracer.trace("actions.referencesRec.if_cond kind=plain")
            for sym in symbols_ast(a):
                if __debug__ and sym not in refs:
                    xtracer.trace("actions.collectSymbols.add %s" % str(sym))
                refs.add(sym)
```

## Files Modified

| File | Changes |
|------|---------|
| `actions/action.go:410` | Fix `SomeCondition.Children()` to return `[...Params, Fmla, ?Index]` |
| `isolate/isolate.go:956` | Add BEGIN/END markers to per-action loop |
| `~/pyivy/ivy/ivy/ivy_isolate.py` | Add matching BEGIN/END markers |
| `actions/transforms.go` | Add if_cond kind trace |
| `~/pyivy/ivy/ivy/ivy_actions.py` | Add matching if_cond kind trace |

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...` — must compile
2. `cd ~/goivy && make test` — full test suite passes
3. `cd ~/goivy && make golden` — divergence at line 151251 resolved; traces advance further
