# Plan: Add caller-site xtracer to LocalAction creation to locate uniqueID divergence

**Created**: 2026-03-28 19:24

## Context

`make golden` diverges at xtrace line 142319:
```
go : XTRACE: LocalAction.__init__ uniqueID=587
py : XTRACE: LocalAction.__init__ uniqueID=644
```

Python creates 57 more LocalAction instances than Go (644 vs 587). Rather than guessing which call sites are responsible, we'll add a **caller tag** to every LocalAction creation xtrace in both Go and Python so the traces become directly diffable.

## Approach

Change the existing xtrace from:
```
LocalAction.__init__ uniqueID=587
```
to:
```
LocalAction.__init__ uniqueID=587 caller=compiler.compile_expression
```

Each call site gets a unique `caller=` tag matching between Go and Python. Then re-run `make golden` — the first divergence in caller tags reveals exactly where Go is missing (or Python is adding) LocalAction allocations.

## Step 1: Add caller tag to Go NewLocalAction constructors

Both `ast/ast.go` and `actions/action.go` have `NewLocalAction` methods. Add a `caller string` parameter to each, and include it in the xtrace output.

**ast/ast.go** `AstConfig.NewLocalAction`:
- Change signature: `func (cfg *AstConfig) NewLocalAction(caller string, args ...Node) *LocalAction`
- Change trace: `xtracer.Trace("LocalAction.__init__ uniqueID=%d caller=%s", cfg.LocalActionCtr, caller)`

**actions/action.go** `ActionsConfig.NewLocalAction`:
- Change signature: `func (cfg *ActionsConfig) NewLocalAction(caller string, args ...lg.Expr) *LocalAction`
- Change trace: `xtracer.Trace("LocalAction.__init__ uniqueID=%d caller=%s", id, caller)`

## Step 2: Update all Go call sites with caller tags

Each call site passes a string matching its Python equivalent:

| # | Go File:Line | Caller Tag | Python Equivalent |
|---|---|---|---|
| 1 | actions/action.go:520 | `"actions.IfThenElseAction.action_update"` | ivy_actions.py:931 |
| 2 | actions/action.go:732 | `"actions.ReturnAction.action_update"` | ivy_actions.py:1354 |
| 3 | actions/match.go:522 | `"actions.match"` | ivy_actions.py:1039 |
| 4 | actions/update.go:1730 | `"actions.action_on_subgoal"` | (check Python equivalent) |
| 5 | ast/lower_var.go:70 | `"ast.lower_var"` | ivy_parser.py:2731 |
| 6 | compiler/action.go:685 | `"compiler.compile_cmpd_local"` | ivy_compiler.py:625 or :630 |
| 7 | compiler/action.go:1027 | `"compiler.compile_local_special"` | ivy_compiler.py:635 |
| 8 | compiler/action.go:1047 | `"compiler.compile_local_seq"` | ivy_compiler.py:630 |
| 9 | compiler/action.go:1093 | `"compiler.compile_local_action"` | ivy_compiler.py:689 |
| 10 | compiler/compiler.go:99 | `"compiler.compile_expression"` | ivy_compiler.py:186 |
| 11 | compiler/phase6.go:752 | `"compiler.compile_proof"` | ivy_compiler.py:865 |
| 12 | ast/ast.go:1350 (Clone) | `"ast.LocalAction.clone"` | (Python clone/copy) |
| 13 | lalr_full/grammar_v17.y:4340 | `"parser.local_action_rule"` | ivy_parser.py:3204 |
| 14 | isolate/helpers.go:1031 | `"isolate.helpers"` | ivy_isolate.py:1441 |

## Step 3: Add caller tag to Python LocalAction.__init__

In `/Users/jaten/pyivy/ivy/ivy/ivy_actions.py`, add `caller` kwarg to `LocalAction.__init__`:

```python
class LocalAction(Action):
    def __init__(self, *args, caller="unknown"):
        Action.__init__(self, *args)
        global local_action_ctr
        self.unique_id = local_action_ctr
        local_action_ctr += 1
        if __debug__:
            from . import xtracer
            xtracer.trace("LocalAction.__init__ uniqueID={} caller={}".format(self.unique_id, caller))
```

## Step 4: Update all Python call sites with matching caller tags

| # | Python File:Line | Caller Tag |
|---|---|---|
| 1 | ivy_actions.py:931 | `caller="actions.IfThenElseAction.action_update"` |
| 2 | ivy_actions.py:1039 | `caller="actions.match"` |
| 3 | ivy_actions.py:1354 | `caller="actions.ReturnAction.action_update"` |
| 4 | ivy_compiler.py:186 | `caller="compiler.compile_expression"` |
| 5 | ivy_compiler.py:625 | `caller="compiler.compile_cmpd_local"` |
| 6 | ivy_compiler.py:630 | `caller="compiler.compile_local_seq"` |
| 7 | ivy_compiler.py:635 | `caller="compiler.compile_local_special"` |
| 8 | ivy_compiler.py:689 | `caller="compiler.compile_local_action"` |
| 9 | ivy_compiler.py:865 | `caller="compiler.compile_proof"` |
| 10 | ivy_parser.py:2731 | `caller="ast.lower_var"` |
| 11 | ivy_parser.py:3204 | `caller="parser.local_action_rule"` |
| 12 | ivy_isolate.py:1441 | `caller="isolate.helpers"` |

## Step 5: Run `make golden` and analyze

Re-run `make golden`. The output will now show caller tags on every LocalAction creation. Diff the go/py traces — the first line where the caller tags diverge reveals the exact source of the 57 missing allocations.

## Verification

1. `make golden` should produce output with `caller=` tags on every `LocalAction.__init__` line
2. Compare Go vs Python caller tag sequences to identify which call site(s) produce different counts
3. Use that information to plan the actual fix

## Files to modify

**Go** (add `caller` param + update calls):
- `ast/ast.go` — NewLocalAction signature + Clone call
- `actions/action.go` — NewLocalAction signature + IfThenElse/Return calls
- `actions/match.go`
- `actions/update.go`
- `compiler/action.go` (4 sites)
- `compiler/compiler.go`
- `compiler/phase6.go`
- `ast/lower_var.go`
- `lalr_full/grammar_v17.y` (and regenerate .go)
- `isolate/helpers.go`
- Test files that call NewLocalAction (update to compile)

**Python** (add `caller` kwarg):
- `ivy/ivy_actions.py` — __init__ + 3 call sites
- `ivy/ivy_compiler.py` — 5 call sites
- `ivy/ivy_parser.py` — 2 call sites
- `ivy/ivy_isolate.py` — 1 call site
