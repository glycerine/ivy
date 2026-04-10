# Add Tracing Coverage + Conformance Fixes for int_update Overrides and Sequence

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
2. For every trace comparison point touched, audits **whether the Go behavior actually conforms to the Python behavior**, and fixes the conformance gaps that the audit surfaces — not just the trace alignment.

### Why the divergence happens (mechanics)

1. Python's `ReturnAction` (ivy_actions.py:1663) is a bare `object` subclass (not an `Action` subclass) with its own `int_update` that bypasses `Action.int_update`. So Python does not emit the base `actions.IntUpdate ENTER type=%s` trace at line 203.

2. Go's `IntUpdate` dispatch (actions/update.go:1127) has no case for `*ReturnAction`. It falls into the `default` branch (line 1189) which emits `actions.IntUpdate ENTER type=ReturnAction` and returns `NullUpdate()`. So Go emits an extra trace Python never produces.

3. The same "override int_update without emitting trace" pattern exists in Python for `DebugAction.int_update` (line 1267) and `NativeAction.int_update` (line 1286). They have not surfaced a divergence yet only because this test path does not hit them. On the Go side, both have their own `IntUpdate` methods (update.go:989 and :997) that are also silent.

4. Once we add traces to surface every `int_update` invocation, the conformance audit reveals deeper gaps in the underlying implementations (annotation types, missing logic, silent skips). These are fixed in the same plan.

## Conformance audit (per trace-comparison site)

Notation:
- **modified**: the list of modified symbols
- **TR**: transition relation clauses (formula + annotation)
- **Pre**: precondition clauses (formula + annotation)

The annotation distinction is real and load-bearing: Python code such as `ivy_interp.py:58` (`clauses.annot is None`) and `ivy_actions.py:854` (`update[1].annot is not None: update[1].annot.lineno = op.lineno`) behaves differently for `None` vs `EmptyAnnotation()` vs `ComposeAnnotation(...)`.

### Site 1 — `DebugAction.int_update`

**Python (ivy_actions.py:1267-1268):**
```python
def int_update(self,domain,pvars):
    return ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation()))
```

**Go (actions/update.go:989-991):**
```go
func (a *DebugAction) IntUpdate(ctx *UpdateContext) *Update {
    return NullUpdate()  // TR/Pre annot = nil
}
```

| Check          | Python                    | Go (current)             | Status |
|----------------|---------------------------|--------------------------|--------|
| modified       | `[]`                     | `[]`                     | ✓      |
| TR formula     | empty (tautology)         | empty (tautology)        | ✓      |
| **TR annot**   | **`EmptyAnnotation()`**   | **`nil`**                | **✗** |
| Pre formula    | `[[]]` (false)           | `[[]]` (false)           | ✓      |
| **Pre annot**  | **`EmptyAnnotation()`**   | **`nil`**                | **✗** |
| ENTER trace    | none                     | none                     | (both silent) |

**Fixes**: (a) add ENTER trace on both sides, (b) construct Update with `EmptyAnnotation{}` in Go.

### Site 2 — `NativeAction.int_update`

**Python (ivy_actions.py:1286-1287):**
```python
def int_update(self,domain,pvars):
    return ([], true_clauses(), false_clauses())   # annot defaults to None
```

**Go (actions/update.go:987-991):**
```go
// Python: NativeAction.int_update returns ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation()))  ← STALE/WRONG COMMENT
func (a *NativeAction) IntUpdate(ctx *UpdateContext) *Update {
    return NullUpdate()  // TR/Pre annot = nil
}
```

| Check          | Python                | Go (current)         | Status |
|----------------|-----------------------|----------------------|--------|
| modified       | `[]`                 | `[]`                 | ✓      |
| TR formula     | empty (tautology)     | empty (tautology)    | ✓      |
| TR annot       | `None`               | `nil`                | ✓      |
| Pre formula    | `[[]]` (false)       | `[[]]` (false)       | ✓      |
| Pre annot      | `None`               | `nil`                | ✓      |
| ENTER trace    | none                 | none                 | (both silent) |

**Fixes**: (a) add ENTER trace on both sides, (b) correct the stale comment. No implementation change needed.

### Site 3 — `ReturnAction.int_update`

**Python (ivy_actions.py:1664-1665):**
```python
def int_update(self,domain,pvars):
    return ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation()))
```

**Go (current)**: no explicit method, no switch case → falls to `default` at update.go:1189, emits `actions.IntUpdate ENTER type=ReturnAction`, returns `NullUpdate()` (nil annot).

| Check          | Python                    | Go (current)             | Status |
|----------------|---------------------------|--------------------------|--------|
| modified       | `[]`                     | `[]`                     | ✓      |
| TR formula     | empty (tautology)         | empty (tautology)        | ✓      |
| **TR annot**   | **`EmptyAnnotation()`**   | **`nil`**                | **✗** |
| Pre formula    | `[[]]` (false)           | `[[]]` (false)           | ✓      |
| **Pre annot**  | **`EmptyAnnotation()`**   | **`nil`**                | **✗** |
| **ENTER trace**| **none emitted**          | **emits ENTER**          | **✗** |

**Fixes**: (a) add ENTER trace on Python side; (b) give Go an explicit `ReturnAction.IntUpdate` method that emits the same trace and returns an Update with `EmptyAnnotation{}`; (c) add `case *ReturnAction` to the dispatch switch so it stops falling into `default`.

### Site 4 — `Sequence.int_update`

This is the calling context where the i=233572 divergence is observed. The divergence's symptom is the `ReturnAction` issue above, but auditing Sequence reveals **multiple** independent conformance gaps that the user has asked us to fix in this same plan.

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
        if xtracer.Enabled { /* compose[i] childType / childModified */ }
        result = ComposeUpdates(result, axioms, childUpdate)
        if xtracer.Enabled { /* compose[i] resultModified */ }
    }
    return result
}
```

| Check                              | Python                    | Go (current)              | Status |
|------------------------------------|---------------------------|---------------------------|--------|
| ENTER / EXIT trace                 | emitted                   | emitted (deferred)         | ✓      |
| compose[i] child trace             | emitted                   | emitted                   | ✓      |
| compose[i] result trace            | emitted                   | emitted                   | ✓      |
| **initial update annot**           | **`EmptyAnnotation()`**   | **`nil` (NullUpdate)**    | **✗** |
| child dispatch order               | args order                | Elems order               | ✓      |
| **non-Action child**               | **AttributeError (raise)**| **silent `continue`**      | **✗** |
| **lineno stamping after compose**  | **explicit assignment**   | **MISSING**               | **✗** |
| compose annotation type produced   | `ComposeAnnotation`       | `ConjAnnotation` (default) | **✗ (downstream)** |

#### Sub-gap 4a — initial update annotation

Python initializes `update` with `true_clauses(EmptyAnnotation())` for TR and `false_clauses(EmptyAnnotation())` for Pre. Go uses `NullUpdate()` which sets annot to `nil`.

Downstream code checks `annot is None` (e.g. `ivy_interp.py:58`); the two paths diverge.

#### Sub-gap 4b — `ComposeUpdates` does not pass `compose` annot_op

**Python `compose_updates` (ivy_transrel.py:319-370)** explicitly composes annotations via `annot_op = lambda x,y: x.compose(y)` passed into `and_clauses`:
```python
annot_op = lambda x,y: x.compose(y) if x is not None and y is not None else None
new_clauses = and_clauses(clauses1, rename_clauses(and_clauses(clauses2,mid_ax),map2),annot_op=annot_op)
...
temp = and_clauses(clauses1,rename_clauses(and_clauses(pre2,mid_ax),map2),annot_op=my_annot_op)
```
Effect: Python's compose_updates produces a TR whose annot is `ComposeAnnotation(...)` (or `None`).

**Go `ComposeUpdates` (actions/transrel.go:633-746)** uses `module.AndClausesTyped` which has no annot_op argument and falls back to the default Conj-based merging in `module.andClausesImpl` (module/ops.go:131-143):
```go
// Python default: annot = a.annot if annot is None else annot if a.annot is None else annot.conj(a.annot)
for _, c := range args {
    if c.Annot != nil {
        if annot == nil {
            annot = c.Annot
        } else if conjer, ok := annot.(AnnotConjoiner); ok {
            annot = conjer.ConjWith(c.Annot)
        }
    }
}
```
Effect: Go's ComposeUpdates produces a TR whose annot is `ConjAnnotation(...)` instead of `ComposeAnnotation(...)`.

This means:
- The annotation TYPE produced by Go's ComposeUpdates is wrong (Conj vs Compose).
- The Go code path is silently using Conj for everything, which is correct for top-level And but wrong for sequential composition.
- This is an *upstream* conformance gap that 4c (lineno stamping) depends on.

The compose annot_op is the same lambda twice in Python (in compose_updates and as `my_annot_op` at line 447). It must be wired into both `and_clauses` calls inside Go's ComposeUpdates.

Go already exposes `module.AndClausesWithAnnotOp` (module/ops.go:112) for exactly this case. We pass a Go-side compose annot_op closure that calls `result.Compose(other)` which already returns `*ComposeAnnotation` (annotation.go:68 and 106 and 149 etc).

#### Sub-gap 4c — lineno stamping after compose

**Python (line 854):**
```python
if hasattr(op,'lineno') and update[1].annot is not None:
    update[1].annot.lineno = op.lineno
```

This pins the source op's lineno onto the resulting TR annotation. Used downstream by counter-example reconstruction (`ivy_trace.py:329, 338`, `match_annotation`).

**Go**: missing entirely.

In Go, after composing, `result.TR.Annot` should be a `*ComposeAnnotation` (after sub-gap 4b is fixed), which has a `Lineno *ast.Location` field. We type-assert to `*ComposeAnnotation` and set `Lineno = &loc`.

`act.GetLineno()` returns the location; `act.HasLoc` indicates whether it was set (Go analog of Python's `hasattr(op,'lineno')`).

#### Sub-gap 4d — non-Action children

Python's `for i,op in enumerate(self.args)` iterates blindly; if `op.int_update` is missing it will `AttributeError`. There is **no silent skip** in Python.

Go's `if act == nil { continue }` after `unwrapToAction(child)` silently swallows non-Action children. This hides upstream construction bugs and produces a different `i` index in the trace (Python increments `i` for the bad child, Go does not).

The faithful port is to `panic` (matching Python's runtime crash) rather than skip.

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

`type(self).__name__` matches `Action.int_update` line 203 verbatim.

### Go — `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/update.go`

**Change 4 — NativeAction.IntUpdate (line 989)**: add ENTER trace, fix stale comment.
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

**Change 5 — DebugAction.IntUpdate (line 997)**: add ENTER trace, fix annotation to `EmptyAnnotation{}` to match Python.
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

**Change 6 — Add ReturnAction.IntUpdate (new method, alongside NativeAction/DebugAction)**:
```go
// --- ReturnAction ---

// IntUpdate for ReturnAction is a no-op — skips update axioms.
// Python: ReturnAction.int_update (ivy_actions.py:1664) returns
// ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation())).
// In Python, ReturnAction is a bare `object` subclass (not Action), so it
// bypasses Action.int_update; both sides emit the same ENTER trace explicitly.
func (a *ReturnAction) IntUpdate(ctx *UpdateContext) *Update {
    xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(a))
    return &Update{
        Modified: []*lg.Const{},
        TR:       module.TrueClauses(EmptyAnnotation{}),
        Pre:      module.FalseClauses(EmptyAnnotation{}),
    }
}
```

**Change 7 — IntUpdate dispatch switch (line 1133)**: add `case *ReturnAction` before `default`:
```go
case *NativeAction:
    return a.IntUpdate(ctx)
case *DebugAction:
    return a.IntUpdate(ctx)
case *ReturnAction:                       // ← NEW
    return a.IntUpdate(ctx)                // ← NEW
// ...other cases...
default:
    // Unreachable for known Action types. Defensive fallback.
    xtracer.Trace("actions.IntUpdate ENTER type=%s", ActionTypeName(action))
    return NullUpdate()
}
```

### Go — `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/transrel.go`

**Change 8 — `ComposeUpdates` must use compose annot_op for the TR-and-Pre AndClauses calls (sub-gap 4b).**

Python compose_updates passes `annot_op=lambda x,y: x.compose(y) if x is not None and y is not None else None` to two specific `and_clauses` calls (the ones producing `new_clauses` and `temp`). All other and_clauses calls inside compose_updates use the default Conj merge.

Define a package-local helper:
```go
// composeAnnotOp is the annot_op used by ComposeUpdates when composing
// TR clauses and Pre clauses, matching Python's compose_updates which
// passes annot_op = lambda x,y: x.compose(y) if x is not None and y is not None else None.
// (Same as Python's my_annot_op at ivy_transrel.py:447.)
func composeAnnotOp(annots ...interface{}) interface{} {
    if len(annots) == 0 {
        return nil
    }
    var result Annotation
    for _, a := range annots {
        if a == nil {
            return nil
        }
        ann, ok := a.(Annotation)
        if !ok {
            return nil
        }
        if result == nil {
            result = ann
        } else {
            result = result.Compose(ann)
        }
    }
    return result
}
```

In `ComposeUpdates`, replace:
```go
newTR := module.AndClausesTyped(clauses1, module.RenameClauses(module.AndClausesTyped(clauses2, midAx), map2))
```
with:
```go
newTR := module.AndClausesWithAnnotOp(composeAnnotOp, clauses1, module.RenameClauses(module.AndClausesTyped(clauses2, midAx), map2))
```

And replace:
```go
temp := module.AndClausesTyped(clauses1, module.RenameClauses(module.AndClausesTyped(pre2, midAx), map2))
```
with:
```go
temp := module.AndClausesWithAnnotOp(composeAnnotOp, clauses1, module.RenameClauses(module.AndClausesTyped(pre2, midAx), map2))
```

Note: only these two outer calls take `composeAnnotOp`. The inner `and_clauses(clauses2, mid_ax)` and `and_clauses(pre2, mid_ax)` calls remain plain `AndClausesTyped` (matching Python where those don't pass annot_op).

After this change Go's ComposeUpdates produces TR/Pre annotations of type `*ComposeAnnotation`, matching Python.

### Go — `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/update.go` (Sequence)

**Change 9 — `Sequence.IntUpdate` (line 1270)**: fix initial annotation, port lineno stamping, replace silent skip.

```go
func (s *Sequence) IntUpdate(ctx *UpdateContext) *Update {
    xtracer.Trace("actions.Sequence.int_update ENTER")
    defer xtracer.Trace("actions.Sequence.int_update EXIT")

    // Python (ivy_actions.py:841): update = ([], true_clauses(EmptyAnnotation()), false_clauses(EmptyAnnotation()))
    result := &Update{
        Modified: []*lg.Const{},
        TR:       module.TrueClauses(EmptyAnnotation{}),
        Pre:      module.FalseClauses(EmptyAnnotation{}),
    }
    axioms := ctx.BackgroundTheory()

    for i, child := range s.Elems {
        act := unwrapToAction(child)
        if act == nil {
            // Python iterates blindly and would AttributeError; faithful port panics.
            panic(fmt.Sprintf("Sequence.IntUpdate: child %d is not an Action: %T", i, child))
        }
        childUpdate := IntUpdate(act, ctx)
        if xtracer.Enabled {
            childModNames := make([]string, len(childUpdate.Modified))
            for j, m := range childUpdate.Modified {
                childModNames[j] = fmt.Sprintf("'%v'", m.Name)
            }
            sort.Strings(childModNames)
            xtracer.Trace("actions.Sequence.int_update compose[%d] childType=%s childModified=[%v]", i, ActionTypeName(act), strings.Join(childModNames, ", "))
        }
        result = ComposeUpdates(result, axioms, childUpdate)
        if xtracer.Enabled {
            resultModNames := make([]string, len(result.Modified))
            for j, m := range result.Modified {
                resultModNames[j] = fmt.Sprintf("'%v'", m.Name)
            }
            sort.Strings(resultModNames)
            xtracer.Trace("actions.Sequence.int_update compose[%d] resultModified=[%v]", i, strings.Join(resultModNames, ", "))
        }

        // Python (ivy_actions.py:854):
        //   if hasattr(op,'lineno') and update[1].annot is not None:
        //       update[1].annot.lineno = op.lineno
        // After Change 8, result.TR.Annot is *ComposeAnnotation (or nil).
        if act.HasLineno() && result.TR != nil && result.TR.Annot != nil {
            if compAnnot, ok := result.TR.Annot.(*ComposeAnnotation); ok {
                loc := act.GetLineno()
                compAnnot.Lineno = &loc
            }
        }
    }
    return result
}
```

Note: `act.HasLineno()` is the analog of Python's `hasattr(op,'lineno')`. The current `ActionBase` exposes `HasLoc bool` as a struct field but no `HasLineno()` method on the `Action` interface. Add a `HasLineno()` method on `ActionBase` (in `module/action.go`) returning `b.HasLoc`, and ensure all Action types satisfy it via embedding.

```go
// in module/action.go, alongside GetLineno/SetLineno (lines 56-57):
func (b *ActionBase) HasLineno() bool { return b.HasLoc }
```

If the `Action` interface in `module/action.go` does not already include `HasLineno()`, add it there. (Verify whether `GetLineno()` is on the interface; if so, add `HasLineno()` next to it.)

## Files to modify (summary)

1. `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_actions.py`
   — 3 one-line trace additions in DebugAction, NativeAction, ReturnAction.

2. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/update.go`
   - NativeAction.IntUpdate: +1 trace line, fix stale comment.
   - DebugAction.IntUpdate: +1 trace line, return struct with `EmptyAnnotation{}`.
   - Add new `ReturnAction.IntUpdate` method (~10 lines).
   - Add `case *ReturnAction` to dispatch switch.
   - Sequence.IntUpdate: replace `NullUpdate()` initial with `EmptyAnnotation{}` Update; replace silent skip with panic; add lineno-stamping block after `ComposeUpdates`.

3. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/transrel.go`
   - Add private helper `composeAnnotOp`.
   - In `ComposeUpdates`, switch the two outer `AndClausesTyped` calls (those that produce `newTR` and `temp`) to `module.AndClausesWithAnnotOp(composeAnnotOp, ...)`.

4. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/module/action.go`
   - Add `HasLineno()` method to `ActionBase`.
   - If the `Action` interface declares `GetLineno()`, also declare `HasLineno()`.

No tests need to be renamed or removed. Existing tests in `actions/impl_test.go` for `ComposeUpdates` may need their expected annotations updated if they assert on annotation type — to be checked during implementation.

## Post-fix trace expectation around i=233572

```
N    go : XTRACE: actions.Sequence.int_update compose[0] resultModified=[...]
N    py : XTRACE: actions.Sequence.int_update compose[0] resultModified=[...]
N+1  go : XTRACE: actions.IntUpdate ENTER type=ReturnAction
N+1  py : XTRACE: actions.IntUpdate ENTER type=ReturnAction
N+2  go : XTRACE: actions.Sequence.int_update compose[1] childType=ReturnAction childModified=[]
N+2  py : XTRACE: actions.Sequence.int_update compose[1] childType=ReturnAction childModified=[]
```

Beyond that, the next divergence (if any) will reveal the next conformance gap, exactly the way the framework is designed to work.

## Verification

1. Run the failing golden test:
   ```sh
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy
   go test ./parser/ -run TestOrdLive -v 2>&1 | tee ~/ivy/goivy/log.red.new
   ```

2. Confirm the test either passes or its divergence shifts strictly later than `233572`. The test framework's `golden_test.go:452` will print the new divergence line.

3. Spot-check around the previous divergence index to verify Go/Py traces now align 1:1.

4. Run broader regression tests:
   ```sh
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy
   go test ./...
   ```

5. **Critical conformance check**: a new divergence emitted as `actions.IntUpdate ENTER type=DebugAction` or `type=NativeAction` would mean the trace addition has surfaced a real behavior divergence in those code paths — investigate as a separate finding (this is the trace framework's *purpose*: detect them).

6. **ComposeUpdates change check**: any new divergence around `transrel.ComposeUpdates result HASH canon=` lines would signal that the `composeAnnotOp` change has changed the canonical hash in a way Python does not expect. Compare the annotation portion of the canon strings. If hashes diverge but the formula structures match, the canon may be including annot in a way the Python side does not — adjust the canon emission or the annot construction to match.

7. **ComposeUpdates unit tests**: tests in `actions/impl_test.go` (`TestComposeUpdates*`) may begin asserting `*ComposeAnnotation` instead of nil/Conj. Update test assertions as needed; this is expected as part of the conformance change.

8. **Sequence non-Action panic check**: if any test newly panics with `Sequence.IntUpdate: child N is not an Action`, that is a previously-hidden upstream bug surfacing. Investigate the upstream construction site rather than reverting the panic.

## Risks / considerations

- **Trace cost**: `if __debug__` guards in Python make new traces zero-cost in optimized mode; Go traces are guarded by `xtracer.Enabled` checks already in the code path. Negligible runtime cost.
- **Annotation type change** (Change 8): switching from `ConjAnnotation` to `ComposeAnnotation` for sequential composition is the biggest behavior change in this plan. It is the *correct* behavior per the Python source of truth, but it may shift downstream traces and counter-example reconstruction in golden tests that exercise compose-heavy code. The verification steps above explicitly look for this.
- **`HasLineno()` interface addition**: adding a method to the Action interface forces every concrete Action type to satisfy it. Since they all embed `ActionBase`, the embedded `HasLineno()` automatically satisfies the interface.
- **Sequence panic** instead of silent skip: if any production code path constructs a Sequence with non-Action children, this panic will surface a bug previously hidden by the silent skip. This is desired behavior (mechanical port faithfulness > silent papering-over).
- **DebugAction annotation change** is low risk because DebugAction is rare; only places that inspect `annot is None` would be affected.

## Out-of-scope / follow-ups (NOT in this plan)

- Audit of the other int_update trace-comparison sites (InstantiateAction, ChoiceAction, EnvAction, IfAction, WhileAction, LocalAction, LetAction, BindOldsAction, CallAction). Each should get the same conformance audit eventually but is unrelated to the current divergence.
- Audit of `ChoiceAction`, `EnvAction`, `IfAction`'s annotation handling — these also build complex annotations and likely have analogous gaps.
- Whether `compose_updates` *outside* of Sequence is called from any other Go path (currently no — `Sequence.IntUpdate` is the only non-test caller in Go).
