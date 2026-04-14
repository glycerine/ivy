# Plan: Fix `PrefixAction` early-return divergence from Python `prefix_action`

Created: 2026-04-14 ~21:30 UTC

## Context

**Divergence**: `make golden` diverges at xtrace line 734726 (in `TestOrdLive`).
- Both Go and Python enter `replaceNamedBindersAst` with an ActionTerm for binding `ext:lclock.next`
- The ActionTerm's `stmt` field already differs:
  - Go: `stmt:(sequence stmts:[(assumeAction ...)])`  — 1 level of Sequence
  - Python: `stmt:(sequence stmts:[(sequence stmts:[(sequence stmts:[(assumeAction ...)])])])` — 3 levels

**Root cause**: Go's `PrefixAction` (`actions/helpers.go:72-75`) has an early return when `stmts` is empty:
```go
if len(stmts) == 0 {
    return action
}
```

Python's `prefix_action` (`ivy_actions.py:1570-1575`) has **no** such check — it ALWAYS wraps in a `Sequence`, even with empty stmts:
```python
def prefix_action(self,stmts):
    res = Sequence(*(stmts + [self]))
    self.copy_formals(res)
    if hasattr(self,"lineno"):
        res.lineno = self.lineno
    return res
```

Note: Python's `postfix_action` DOES have the empty check (`if len(stmts) == 0: return self`), and Go's `PostfixAction` matches this — so only `PrefixAction` diverges.

**How this causes 2 extra Sequence levels**: In `SharedStep7_InstrumentActions`, `instrStmt` recursively processes actions and calls `PrefixAction`/`PostfixAction` on each. For the `ext:lclock.next` binding:

1. `instrStmt(Sequence(assumeAction))` recurses into the child:
   - `instrStmt(assumeAction)` — no modified symbols, so preEvents=[] postEvents=[]
   - Python: `prefix_action(assumeAction, [])` → `Sequence(assumeAction)` (+1 level)
   - Go: `PrefixAction(assumeAction, [])` → `assumeAction` unchanged (early return)

2. Back in the parent, clone Sequence with the transformed child:
   - Python: `Sequence(Sequence(assumeAction))` (2 levels)
   - Go: `Sequence(assumeAction)` (1 level)

3. prefix/postfix on the cloned result (also empty events):
   - Python: `prefix_action(Sequence(Sequence(assumeAction)), [])` → `Sequence(Sequence(Sequence(assumeAction)))` (+1 = 3 levels)
   - Go: `PrefixAction(Sequence(assumeAction), [])` → `Sequence(assumeAction)` (early return, stays 1 level)

4. Step 8 `ConcatActions`/`concat_actions` both flatten one level, preserving the difference.

## Fix

### Remove early return from `PrefixAction`

**File**: `actions/helpers.go:72-87`

Remove lines 73-75:
```go
if len(stmts) == 0 {
    return action
}
```

This makes Go's `PrefixAction` always wrap in a `Sequence`, matching Python's `prefix_action`.

## Critical file to modify

1. `actions/helpers.go` — PrefixAction (remove early return)

## Verification

1. `go build ./...` compiles clean
2. `make golden` — divergence advances past line 734726
