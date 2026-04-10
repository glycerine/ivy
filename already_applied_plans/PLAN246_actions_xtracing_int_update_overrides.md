# Add Tracing Coverage + Conformance Audit for int_update Overrides

*Created: 2026-04-09 21:10*

## Context

The `TestOrdLive` golden test diverges at `i=233572` (see `~/ivy/goivy/log.red`):

```
233571  go : XTRACE: actions.Sequence.int_update compose[0] resultModified=[...]
233571  py : XTRACE: actions.Sequence.int_update compose[0] resultModified=[...]

233572  go : XTRACE: actions.IntUpdate ENTER type=ReturnAction           ← Go emits this
233572  py : XTRACE: actions.Sequence.int_update compose[1] childType=ReturnAction  ← Py skips straight to this
```

The trace-matching framework is a **conformance detector**: when traces diverge, it means we have found a place where Go behavior deviates from Python behavior. Silent-conformance is the real goal; the trace lines are the instrument. This plan therefore:

1. Broadens xtrace coverage so every `int_update` override on both sides emits a matching trace.
2. For every new or existing trace comparison point touched, audits **whether the Go behavior actually conforms to the Python behavior** — and calls out conformance gaps that need fixing alongside the trace additions.

### Why the divergence happens (mechanics)

1. Python's `ReturnAction` (ivy_actions.py:1663) is a bare `object` subclass (not an `Action` subclass) with its own `int_update` that bypasses `Action.int_update` — so Python does not emit the base `actions.IntUpdate ENTER type=%s` trace that `Action.int_update` emits at line 203.

2. Go's `IntUpdate` dispatch (actions/update.go:1127) has no case for `*ReturnAction`. It falls through to the `default` case (line 1189) which emits `actions.IntUpdate ENTER type=ReturnAction` and returns `NullUpdate()`. So Go emits an extra trace Python never produces.

3. The same "override int_update without emitting trace" pattern exists in Python for:
   - `DebugAction.int_update` (ivy_actions.py:1267)
   - `NativeAction.int_update` (ivy_actions.py:1286)
   These have not caused a divergence yet only because this test path does not hit them. On the Go side, `NativeAction.IntUpdate` (update.go:989) and `DebugAction.IntUpdate` (update.go:997) are also silent, which papers over the issue.

## Conformance audit (per trace-comparison site we are touching)

For each site, we check:
(a) modified list equal,
(b) TR (transition relation) equal — formula *and* annotation,
(c) Pre (precondition) equal — formula *and* annotation,
(d) any side-effects / extra logic match.

The annotation distinction is real and load-bearing: Python code such as
`ivy_interp.py:58` (`clauses.annot is None`) and `ivy_actions.py:854`
(`update[1].annot is not None: update[1].annot.lineno = op.lineno`) behaves
differently for `None` vs `EmptyAnnotation()`.

### 1. `DebugAction.int_update`

**Python (ivy_actions.py:1267-1268):**
```python
def int_update(self,domain,pvars):
    return ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation()))
```

**Go (actions/update.go:989-991):**
```go
func (a *DebugAction) IntUpdate(ctx *UpdateContext) *Update {
    return NullUpdate()  // returns TR/Pre with nil annot
}
```

| Check          | Python                          | Go (current)          | Status |
|----------------|----------------------------------|-----------------------|--------|
| modified       | `[]`                            | `[]`                  | ✓      |
| TR formula     | empty clauses (tautology)        | empty clauses (taut.) | ✓      |
| **TR annot**   | **`EmptyAnnotation()`**          | **`nil`**             | **✗ NON-CONFORMING** |
| Pre formula    | `[[]]` (false)                  | `[[]]` (false)        | ✓      |
| **Pre annot**  | **`EmptyAnnotation()`**          | **`nil`**             | **✗ NON-CONFORMING** |
| side-effects   | none                            | none                  | ✓      |

**Trace status**: neither side emits an ENTER trace. Both silent.

**Fix required**: (a) add ENTER trace on both sides, AND (b) fix Go to return `EmptyAnnotation{}` not `nil` so it conforms to Python.

### 2. `NativeAction.int_update`

**Python (ivy_actions.py:1286-1287):**
```python
def int_update(self,domain,pvars):
    return ([], true_clauses(), false_clauses())
```
Note: `true_clauses()`/`false_clauses()` with no argument → `annot=None`.

**Go (actions/update.go:987-991):**
```go
// Python: NativeAction.int_update returns ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation()))  ← WRONG COMMENT
func (a *NativeAction) IntUpdate(ctx *UpdateContext) *Update {
    return NullUpdate()  // returns TR/Pre with nil annot
}
```

| Check          | Python                          | Go (current)          | Status |
|----------------|----------------------------------|-----------------------|--------|
| modified       | `[]`                            | `[]`                  | ✓      |
| TR formula     | empty clauses (tautology)        | empty clauses (taut.) | ✓      |
| TR annot       | `None`                          | `nil`                 | ✓      |
| Pre formula    | `[[]]` (false)                  | `[[]]` (false)        | ✓      |
| Pre annot      | `None`                          | `nil`                 | ✓      |
| side-effects   | none                            | none                  | ✓      |

**Trace status**: neither side emits an ENTER trace. Both silent.

**Fix required**: (a) add ENTER trace on both sides. (b) Fix the stale Go comment — it says `EmptyAnnotation()` but Python does not use it. No implementation change needed for annotation.

### 3. `ReturnAction.int_update`

**Python (ivy_actions.py:1664-1665):**
```python
def int_update(self,domain,pvars):
    return ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation()))
```

**Go (current)**: no explicit `IntUpdate` method, no case in the dispatch switch. Falls to `default` at update.go:1189 → emits `actions.IntUpdate ENTER type=ReturnAction` and returns `NullUpdate()`.

| Check          | Python                          | Go (current)          | Status |
|----------------|----------------------------------|-----------------------|--------|
| modified       | `[]`                            | `[]`                  | ✓      |
| TR formula     | empty clauses (tautology)        | empty clauses (taut.) | ✓      |
| **TR annot**   | **`EmptyAnnotation()`**          | **`nil`**             | **✗ NON-CONFORMING** |
| Pre formula    | `[[]]` (false)                  | `[[]]` (false)        | ✓      |
| **Pre annot**  | **`EmptyAnnotation()`**          | **`nil`**             | **✗ NON-CONFORMING** |
| side-effects   | none                            | none                  | ✓      |
| **trace**      | **none emitted**                 | **emits ENTER**       | **✗ NON-CONFORMING** |

**Fix required**: (a) add ENTER trace on Python side (this makes the Go/Py traces align). (b) Give Go an explicit `ReturnAction.IntUpdate` method (replacing default fall-through) that returns `EmptyAnnotation{}` not `nil`, conforming the annotation.

### 4. `Sequence.int_update` (the calling context at i=233572)

This is an **existing** trace-comparison site, not one we're adding. But the immediate divergence happens inside it, so it's worth auditing.

**Python (ivy_actions.py:839-857):**
```python
def int_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.Sequence.int_update ENTER")
    update = ([],true_clauses(EmptyAnnotation()),false_clauses(EmptyAnnotation()))
    axioms = domain.background_theory(pvars)
    for i,op in enumerate(self.args):
        thing = op.int_update(domain,pvars);
        if __debug__:
            child_mod = [s.name for s in thing[0]] if thing[0] is not None else []
            xtracer.trace("actions.Sequence.int_update compose[%d] childType=%s childModified=%s"
                           % (i, type(op).__name__, sorted(child_mod)))
        update = compose_updates(update,axioms,thing)
        if __debug__:
            result_mod = [s.name for s in update[0]] if update[0] is not None else []
            xtracer.trace("actions.Sequence.int_update compose[%d] resultModified=%s"
                           % (i, sorted(result_mod)))
        if hasattr(op,'lineno') and update[1].annot is not None:
            update[1].annot.lineno = op.lineno
    if __debug__: xtracer.trace("actions.Sequence.int_update EXIT")
    return update
```

**Go (actions/update.go:1270-1300):**
```go
func (s *Sequence) IntUpdate(ctx *UpdateContext) *Update {
    xtracer.Trace("actions.Sequence.int_update ENTER")
    defer xtracer.Trace("actions.Sequence.int_update EXIT")
    result := NullUpdate()
    axioms := ctx.BackgroundTheory()
    for i, child := range s.Elems {
        act := unwrapToAction(child)
        if act == nil {
            continue
        }
        childUpdate := IntUpdate(act, ctx)
        if xtracer.Enabled {
            // ... emit compose[i] childType / childModified
        }
        result = ComposeUpdates(result, axioms, childUpdate)
        if xtracer.Enabled {
            // ... emit compose[i] resultModified
        }
    }
    return result
}
```

| Check                          | Python                          | Go (current)          | Status |
|--------------------------------|----------------------------------|-----------------------|--------|
| ENTER trace                    | emitted                         | emitted               | ✓      |
| EXIT trace                     | emitted                         | emitted (deferred)    | ✓      |
| compose[i] childType trace     | emitted                         | emitted               | ✓      |
| compose[i] resultModified      | emitted                         | emitted               | ✓      |
| **initial update annot**       | **`EmptyAnnotation()`**          | **`nil` (NullUpdate)**| **✗ NON-CONFORMING** |
| child dispatch order           | `op.int_update` in args order   | `IntUpdate(act)` in order | ✓  |
| non-Action child handling      | attempts `op.int_update` (raises if missing) | skips silently | ✗ DIFFERENT |
| **lineno stamping**            | **`update[1].annot.lineno = op.lineno` if annot present** | **MISSING** | **✗ NON-CONFORMING** |

**Sequence conformance gaps beyond traces:**

- **Initial update annotation**: Python uses `EmptyAnnotation()`; Go uses `NullUpdate()` which has `nil`. Any downstream code that checks `annot is None` (e.g. `ivy_interp.py:58`) will take different branches.
- **lineno stamping** (ivy_actions.py:854): Python walks each composed child and, if the composed update's TR annot exists, pins `lineno` on the annotation. Go's Sequence does not do this at all. This loses per-child lineno information in the transition relation annotation, which is important for counter-example reconstruction in `ivy_trace.py` / `match_annotation`.
- **Silent skip of non-Action children** (`if act == nil { continue }`): This is a Go-specific workaround when unwrapping `lg.Expr` to `Action` fails. Python would raise `AttributeError`, so a non-Action child in `self.args` would crash Python. If this branch is ever hit in practice the two implementations have already diverged in state-construction upstream, and we are hiding a bug.

**Fix required** (to fully conform Sequence):
(a) Initialize `result` with `EmptyAnnotation{}` on the TR/Pre clauses instead of plain `NullUpdate()`.
(b) Port the lineno-stamping logic after `ComposeUpdates`.
(c) Replace the silent `continue` with either a `panic` (match Python's crash) or an explicit check that all Sequence elements are Actions at construction time.

These Sequence fixes are **not strictly required** to fix the i=233572 divergence (that one only needs the ReturnAction changes), but they are real conformance gaps the user asked us to call out. See "Scope decisions" below.

## Scope decisions

The minimum set of changes to eliminate the i=233572 divergence is only sections 1, 2, 3 above — adding traces and fixing `EmptyAnnotation{}` for `DebugAction` and `ReturnAction`. The Sequence gaps in section 4 are real but the divergence at i=233572 does not depend on them, and touching Sequence risks introducing new divergences elsewhere.

**Recommended scope for this plan:**
- **In scope (required to fix i=233572):**
  - Add Python ENTER traces for `DebugAction`, `NativeAction`, `ReturnAction`.
  - Add Go ENTER traces for the same three.
  - Add Go `ReturnAction.IntUpdate` method and switch case.
  - Fix Go `DebugAction.IntUpdate` to use `EmptyAnnotation{}` (conformance gap).
  - Fix Go `NativeAction.IntUpdate` stale comment.
- **Out of scope for this plan (separate follow-up):**
  - Sequence initial-annotation conformance fix.
  - Sequence lineno-stamping port.
  - Sequence non-Action child handling.
  - These should be addressed in a dedicated plan because they can shift downstream trace behavior in ways that require re-running all golden tests.

If the user disagrees and wants Sequence fixed here too, we can expand scope.

## Changes

### Python — `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_actions.py`

**Change 1 — DebugAction.int_update (line 1267):**
```python
def int_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.IntUpdate ENTER type=%s" % type(self).__name__)
    return ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation()))
```

**Change 2 — NativeAction.int_update (line 1286):**
```python
def int_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.IntUpdate ENTER type=%s" % type(self).__name__)
    return ([], true_clauses(), false_clauses())
```

**Change 3 — ReturnAction.int_update (line 1664):**
```python
def int_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.IntUpdate ENTER type=%s" % type(self).__name__)
    return ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation()))
```

Uses `type(self).__name__` exactly like `Action.int_update` at line 203, emitting `"DebugAction"`, `"NativeAction"`, `"ReturnAction"`.

### Go — `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/update.go`

**Change 4 — NativeAction.IntUpdate (line 989)**: add ENTER trace; fix the comment.
```go
// --- NativeAction ---

// IntUpdate for NativeAction is a no-op — skips update axioms.
// Python: NativeAction.int_update (ivy_actions.py:1286) returns
// ([], true_clauses(), false_clauses()) — annot is None.
func (a *NativeAction) IntUpdate(ctx *UpdateContext) *Update {
    xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(a))
    return NullUpdate()
}
```
Implementation unchanged (NullUpdate's nil annot matches Python's None). Only the trace and comment change.

**Change 5 — DebugAction.IntUpdate (line 997)**: add ENTER trace; fix annotation to `EmptyAnnotation{}` to match Python.
```go
// --- DebugAction ---

// IntUpdate for DebugAction is a no-op — skips update axioms.
// Python: DebugAction.int_update (ivy_actions.py:1267) returns
// ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation())).
func (a *DebugAction) IntUpdate(ctx *UpdateContext) *Update {
    xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(a))
    return &Update{
        Modified: []*lg.Const{},
        TR:       module.TrueClauses(EmptyAnnotation{}),
        Pre:      module.FalseClauses(EmptyAnnotation{}),
    }
}
```

**Change 6 — Add ReturnAction.IntUpdate (new, alongside NativeAction/DebugAction)**:
```go
// --- ReturnAction ---

// IntUpdate for ReturnAction is a no-op — skips update axioms.
// Python: ReturnAction.int_update (ivy_actions.py:1664) returns
// ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation())).
// ReturnAction is a bare `object` subclass in Python (not a subclass of
// Action), so it bypasses Action.int_update and emits its own ENTER trace.
func (a *ReturnAction) IntUpdate(ctx *UpdateContext) *Update {
    xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(a))
    return &Update{
        Modified: []*lg.Const{},
        TR:       module.TrueClauses(EmptyAnnotation{}),
        Pre:      module.FalseClauses(EmptyAnnotation{}),
    }
}
```

**Change 7 — IntUpdate dispatch switch (line 1133)**: add `case *ReturnAction` before the default:
```go
case *NativeAction:
    return a.IntUpdate(ctx)
case *DebugAction:
    return a.IntUpdate(ctx)
case *ReturnAction:                       // ← NEW
    return a.IntUpdate(ctx)                // ← NEW
// ...existing cases...
default:
    // Unreachable for known Action types. Kept as a defensive fallback;
    // any new Action type should add its own case above. Python has no
    // equivalent fallback (unknown types would AttributeError).
    xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
    return NullUpdate()
}
```
After this change, `*ReturnAction` is handled explicitly and will not fall into `default`.

## Files to modify

- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_actions.py` — 3 one-line trace additions
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/update.go` —
  - NativeAction.IntUpdate: +1 trace line, fix stale comment
  - DebugAction.IntUpdate: +1 trace line, return struct with `EmptyAnnotation{}`
  - Add new ReturnAction.IntUpdate method
  - Add `case *ReturnAction` to dispatch switch

No other files need changes. No tests need to be renamed or removed.

## Post-fix trace expectation

At the divergence point, both sides will now emit in lockstep:

```
N    go : XTRACE: actions.Sequence.int_update compose[0] resultModified=[...]
N    py : XTRACE: actions.Sequence.int_update compose[0] resultModified=[...]
N+1  go : XTRACE: actions.IntUpdate ENTER type=ReturnAction
N+1  py : XTRACE: actions.IntUpdate ENTER type=ReturnAction
N+2  go : XTRACE: actions.Sequence.int_update compose[1] childType=ReturnAction childModified=[]
N+2  py : XTRACE: actions.Sequence.int_update compose[1] childType=ReturnAction childModified=[]
```

## Verification

1. Run the failing golden test:
   ```sh
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy
   go test ./parser/ -run TestOrdLive -v 2>&1 | tee ~/ivy/goivy/log.red.new
   ```

2. Confirm the test either passes or the divergence shifts to a strictly later line (revealing the next real issue). The test framework's output (`golden_test.go:452`) will report the new divergence line if any.

3. Spot-check the log around the previous divergence (`i=233572`) and verify go/py lines now match 1:1 around and after this index.

4. Run broader regression tests to ensure the added Python traces and the Go annotation fix do not break other golden tests:
   ```sh
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy
   go test ./...
   ```

5. If any other golden test newly fails with a divergence that mentions `actions.IntUpdate ENTER type=DebugAction` or `type=NativeAction`, the Python-side trace addition has succeeded but exposed a real behavior divergence to investigate separately — exactly the conformance-detection purpose the traces serve.

6. If a new divergence surfaces around the DebugAction or ReturnAction annotation change (i.e. downstream code that inspects `annot is None` now takes a different branch), treat that as a separately-investigatable finding: the annotation fix conforms Go *to* Python, so any new divergence is a case where Go had been relying on its nil-annot behavior and the Python reference says otherwise.

## Risks / considerations

- `if __debug__` guards make Python trace additions zero-cost in optimized mode.
- The trace strings are identical to what `Action.int_update` emits, so downstream diffing needs no updates.
- Every existing passing golden test will now see three additional trace classes (`DebugAction`, `NativeAction`, `ReturnAction`). Because the Go side gains the same traces with the same format, 1:1 lockstep is preserved.
- The `DebugAction` annotation change (`nil` → `EmptyAnnotation{}`) is the only behavioral (non-trace) change. It moves Go toward Python. Because `DebugAction` is itself infrequently used, and because it returns an empty-modified update, the chance of a downstream regression is low but nonzero — verification step 6 exists to catch it.
- Sequence conformance gaps (initial annotation, lineno stamping, non-Action handling) are **deferred to a follow-up plan**. They are documented here so they are not forgotten, but expanding this plan to fix them risks shifting many other golden tests and obscuring whether the i=233572 fix actually worked.

## Follow-up items (documented, not fixed here)

- **Sequence.int_update initial annotation**: `NullUpdate()` → `{TR: TrueClauses(EmptyAnnotation{}), Pre: FalseClauses(EmptyAnnotation{})}`.
- **Sequence.int_update lineno stamping**: port the `if hasattr(op,'lineno') and update[1].annot is not None: update[1].annot.lineno = op.lineno` loop from ivy_actions.py:854.
- **Sequence non-Action child**: replace silent `continue` with explicit handling — likely a panic to match Python's AttributeError, since a non-Action inside a Sequence is a bug.
- **Audit of other int_update trace-comparison sites** (InstantiateAction, ChoiceAction, EnvAction, IfAction, WhileAction, LocalAction, LetAction, BindOldsAction, CallAction): each should get the same conformance audit this plan performed for Debug/Native/Return/Sequence. Expected outcome: most will be clean, but some will surface subtle annotation or ordering gaps.
