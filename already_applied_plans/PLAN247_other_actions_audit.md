# Conformance Audit + Fixes for Remaining int_update Overrides

*Created: 2026-04-09 22:10*

## Context

The previous plan fixed `Sequence.int_update`, `ReturnAction.int_update`, `DebugAction.int_update`, and `NativeAction.int_update` and shifted the `TestOrdLive` divergence from `i=233572` to `i=233622`. The plan explicitly listed the remaining `int_update` overrides as out-of-scope follow-ups:

> - Audit of the other int_update trace-comparison sites (InstantiateAction, ChoiceAction, EnvAction, IfAction, WhileAction, LocalAction, LetAction, BindOldsAction, CallAction).
> - Audit of ChoiceAction, EnvAction, IfAction's annotation handling — these also build complex annotations and likely have analogous gaps.

This plan covers exactly that follow-up. The methodology is identical: for each `int_update` site, compare the Go implementation against the Python source of truth on **modified set, TR formula, TR annot, Pre formula, Pre annot, control flow, and trace coverage**, then fix every gap surfaced by the audit.

The trace-matching framework is a **conformance detector**: silent conformance is the goal; the trace lines are the instrument. Every gap surfaced is a real behavioral divergence, not "just" a trace alignment issue.

Notation in audit tables:
- **modified**: the list of modified symbols
- **TR**: transition relation clauses (formula + annotation)
- **Pre**: precondition clauses (formula + annotation)
- ✓ = conformant, ✗ = divergence to fix, ~ = needs verification before fix

## Key Files

- Python source of truth: `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_actions.py`
- Python helpers: `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_transrel.py`, `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_logic_utils.py`
- Go action methods: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/update.go`
- Go InstantiateAction.IntUpdate: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/action.go:2144`
- Go transition helpers: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/transrel.go`
- Go module-level clause helpers: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/module/ops.go`

## Conformance audit (per site)

### Site 1 — `InstantiateAction.int_update`

**Python (ivy_actions.py:804-820):**
```python
def int_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.InstantiateAction.int_update ENTER")
    inst = self.args[0]
    if hasattr(domain,'macros'):
        im = instantiate_macro(inst,domain.macros)
        if im:
            res = im.compile().int_update(domain,pvars)
            if __debug__: xtracer.trace("actions.InstantiateAction.int_update EXIT")
            return res
    if inst.relname in domain.schemata:
        clauses = domain.schemata[inst.relname].get_instance(inst.args)
        if __debug__: xtracer.trace("actions.InstantiateAction.int_update EXIT")
        return ([],clauses, false_clauses())
    raise IvyError(inst,"instantiation of undefined: {}".format(inst.relname))
```

**Go (actions/action.go:2144-2217):** Has ENTER + deferred EXIT trace, has macros + schemata branches.

| Check                    | Python                   | Go (current)                  | Status |
|--------------------------|--------------------------|-------------------------------|--------|
| ENTER/EXIT trace         | emitted                  | emitted (deferred)            | ✓      |
| modified (schema path)   | `[]`                     | `nil` (line 2209)             | **✗** |
| TR annot (schema path)   | None (`get_instance`)    | None (`FormulaToClauses(_,nil)`) | ✓   |
| Pre annot (schema path)  | None (`false_clauses()`) | None (`FalseClauses(nil)`)    | ✓      |
| undefined-symbol fallthrough | `raise IvyError`     | `return NullUpdate()` (line 2216) | **✗** |
| `ctx.Domain == nil` guard | none (would AttributeError) | `return NullUpdate()` (line 2147-2149) | **✗** |

**Fixes:**
- (1a) Set `Modified: []*lg.Const{}` (not `nil`) on the schema-path Update at action.go:2209.
- (1b) Replace the final `return NullUpdate()` (action.go:2216) with `panic(fmt.Sprintf("instantiation of undefined: %s", instName))`.
- (1c) Remove the `ctx.Domain == nil` early-return at action.go:2147-2149; if there is no domain, panic — matches Python's AttributeError.
- (1d) Replace the `instName == ""` early-return at action.go:2193-2195 with a panic for the same reason.

### Site 2 — `ChoiceAction.int_update`

**Python (ivy_actions.py:882-894):**
```python
def int_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.ChoiceAction.int_update ENTER")
    if determinize and len(self.args) == 2:
        cond = bool_const('___branch:' + str(self.unique_id))
        ite = IfAction(Not(cond),self.args[0],self.args[1])
        if __debug__: xtracer.trace("actions.ChoiceAction.int_update EXIT")
        return ite.int_update(domain,pvars)
    result = [], false_clauses(annot=EmptyAnnotation()), false_clauses(annot=EmptyAnnotation())
    for a in self.args:
        foo = a.int_update(domain, pvars)
        result = join_action(result, foo, domain.background_theory(pvars))
    if __debug__: xtracer.trace("actions.ChoiceAction.int_update EXIT")
    return result
```

**Go (actions/update.go:1360-1383):**
```go
func (a *ChoiceAction) IntUpdate(ctx *UpdateContext) *Update {
    xtracer.Trace("actions.ChoiceAction.int_update ENTER")
    defer xtracer.Trace("actions.ChoiceAction.int_update EXIT")
    if ctx.ActCfg != nil && ctx.ActCfg.Determinize && len(a.Branches) == 2 {
        cond := module.BoolConst("___branch:" + strconv.FormatInt(a.UniqueID, 10))
        ite := NewIfAction(&lg.Not{Body: cond}, a.Branches[0], a.Branches[1])
        return ite.IntUpdate(ctx)
    }
    result := makeUpdate([]*lg.Const{}, lg.False, lg.False, nil)
    axioms := ctx.BackgroundTheory()
    for _, branch := range a.Branches {
        act := unwrapToAction(branch)
        if act == nil {
            continue
        }
        branchUpdate := IntUpdate(act, ctx)
        result = JoinAction(result, branchUpdate, axioms)
    }
    return result
}
```

| Check                       | Python                    | Go (current)                | Status |
|-----------------------------|---------------------------|-----------------------------|--------|
| ENTER/EXIT trace            | emitted                   | emitted (deferred)          | ✓      |
| determinize → IfAction      | `Not(cond)` + int_update  | `Not(cond)` + IntUpdate     | ✓      |
| modified (initial)          | `[]`                      | `[]*lg.Const{}`             | ✓      |
| **TR annot (initial)**      | **`EmptyAnnotation()`**   | **`nil`**                   | **✗** |
| **Pre annot (initial)**     | **`EmptyAnnotation()`**   | **`nil`**                   | **✗** |
| **non-Action branch**       | **AttributeError (raise)**| **silent `continue`**       | **✗** |
| join annotation pipeline    | `or_clauses` → IteAnnot   | `OrClausesTyped` → fixOrAnnot → IteAnnotation via `AnnotIteFunc` | ~ (verify; see annotation deep-dive below) |

**Fixes:**
- (2a) Replace `makeUpdate([]*lg.Const{}, lg.False, lg.False, nil)` with explicit `&Update{Modified: []*lg.Const{}, TR: module.FalseClauses(EmptyAnnotation{}), Pre: module.FalseClauses(EmptyAnnotation{})}`.
- (2b) Replace the silent `if act == nil { continue }` with `panic(fmt.Sprintf("ChoiceAction.IntUpdate: branch %d is not an Action: %T", i, branch))` (loop must be `for i, branch := range a.Branches`).
- (2c) Add per-branch trace inside the loop, mirroring Sequence's `compose[i]` traces, for finer-grained future divergence detection:
  ```go
  if xtracer.Enabled {
      modNames := sortedModNames(branchUpdate.Modified)
      xtracer.Trace("actions.ChoiceAction.int_update branch[%d] childType=%s childModified=[%v]", i, ActionTypeName(act), strings.Join(modNames, ", "))
  }
  ```
  Add the matching trace on the Python side at ivy_actions.py:891-892:
  ```python
  if __debug__:
      child_mod = [s.name for s in foo[0]] if foo[0] is not None else []
      xtracer.trace("actions.ChoiceAction.int_update branch[%d] childType=%s childModified=%s" % (i, type(a).__name__, sorted(child_mod)))
  ```
  (Loop becomes `for i,a in enumerate(self.args):` on the Python side.)

### Site 3 — `EnvAction.int_update`

**Python (ivy_actions.py:910-926):**
```python
def int_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.EnvAction.int_update ENTER")
    if determinize and len(self.args) == 2:
        cond = bool_const('___branch:' + str(self.unique_id))
        ite = IfAction(cond,self.args[0],self.args[1])
        if __debug__: xtracer.trace("actions.EnvAction.int_update EXIT")
        return ite.update(domain,pvars)
    result = [], false_clauses(annot=EmptyAnnotation()), false_clauses(annot=EmptyAnnotation())
    for a in self.args:
        foo = a.update(domain, pvars)
        result = join_action(result, foo, domain.background_theory(pvars))
    if __debug__: xtracer.trace("actions.EnvAction.int_update EXIT")
    return result
```

**Go (actions/update.go:1389-1414):** Same shape as ChoiceAction but calls `GetUpdate(act, ctx)` (matches Python `a.update(...)`) and uses positive `cond` (no `Not`) in the determinize path.

| Check                       | Python                    | Go (current)                | Status |
|-----------------------------|---------------------------|-----------------------------|--------|
| ENTER/EXIT trace            | emitted                   | emitted (deferred)          | ✓      |
| determinize → IfAction      | positive `cond` + `update`| positive cond + `GetUpdate` | ✓      |
| modified (initial)          | `[]`                      | `[]*lg.Const{}`             | ✓      |
| **TR annot (initial)**      | **`EmptyAnnotation()`**   | **`nil`**                   | **✗** |
| **Pre annot (initial)**     | **`EmptyAnnotation()`**   | **`nil`**                   | **✗** |
| **non-Action branch**       | **AttributeError (raise)**| **silent `continue`**       | **✗** |
| child dispatch              | `a.update(...)`           | `GetUpdate(act, ctx)`       | ✓      |

**Fixes:**
- (3a) Same as (2a) for the initial Update annotation: explicit `EmptyAnnotation{}`.
- (3b) Same as (2b) for the silent skip → panic.
- (3c) Same as (2c) for per-branch tracing on both sides. Trace name uses `actions.EnvAction.int_update branch[%d]`. Python loop also becomes `for i,a in enumerate(...)`.

### Site 4 — `IfAction.int_update`

**Python (ivy_actions.py:982-1015):**
```python
def int_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.IfAction.int_update ENTER")
    if used_variables_ast(self.args[0]):
        print (self)
        raise IvyError(self,'variables in "if" conditions must be explicitly quantified')
    if not isinstance(self.args[0],ivy_ast.Some):
        if not is_boolean(self.args[0]):
            raise IvyError(self,'condition must be boolean')
        branches = [self.args[1],self.args[2] if len(self.args) >= 3 else Sequence()]
        upds = [a.int_update(domain,pvars) for a in branches]
        res =  ite_action(self.args[0],upds[0],upds[1],domain.background_theory(pvars))
        if __debug__: xtracer.trace("actions.IfAction.int_update EXIT")
        return res
    if_part,else_part = (a.int_update(domain,pvars) for a in self.subactions())
    res = join_action(if_part,else_part,domain.background_theory(pvars))
    # Hack: the ite annotation comes out reversed. Fix it.
    for i in range(1,3):
        x = res[i].annot
        if isinstance(x,IteAnnotation):
            res[i].annot = IteAnnotation(Not(x.cond),x.elseb,x.thenb)
    if __debug__: xtracer.trace("actions.IfAction.int_update EXIT")
    return res
```

**Go (actions/update.go:1420-1495):** Has ENTER + deferred EXIT, free-vars panic, dispatches to `intUpdateWithSubactions` for `*SomeCondition`, otherwise calls `IteAction(cond, thenUpdate, elseUpdate, axioms)`. The subactions path performs the same `IteAnnotation` reversal as Python.

| Check                          | Python                  | Go (current)                | Status |
|--------------------------------|-------------------------|-----------------------------|--------|
| ENTER/EXIT trace               | emitted                 | emitted (deferred)          | ✓      |
| free-vars check                | `print + raise`         | `panic` (line 1428)         | ✓      |
| **`is_boolean` check**         | **raise IvyError**      | **MISSING**                 | **✗** |
| missing-else → empty Sequence  | `Sequence()`            | `NewSequence()`             | ✓      |
| simple path: `ite_action`      | called                  | `IteAction(cond, ...)`      | ✓      |
| Some path: `subactions` → join | called + reverse        | `intUpdateWithSubactions` reverses | ✓ |
| join annot reversal            | `for i in range(1,3)`   | TR + Pre handled            | ✓      |
| `IteClauses` annot pipeline    | `ite_clauses` → IteAnnot| `module.IteClauses` → ?     | ~ (verify; see annotation deep-dive below) |

**Fixes:**
- (4a) Add the boolean check after the free-vars panic: if `cond` is not boolean-sorted, panic with `"condition must be boolean"`. Use `cond.NodeSort()` and compare against `lg.Boolean`. (Skip this check on the `*SomeCondition` branch — Python only checks in the non-Some branch.)
- (4b) Add per-branch trace inside both the simple path and the subactions path, mirroring Sequence/Choice traces:
  ```go
  if xtracer.Enabled {
      xtracer.Trace("actions.IfAction.int_update then childType=%s thenModified=[%v]", ActionTypeName(thenAct), ...)
      xtracer.Trace("actions.IfAction.int_update else childType=%s elseModified=[%v]", ActionTypeName(elseAct), ...)
  }
  ```
  Add matching traces on the Python side at line 993 (between `upds = [...]` and `res = ite_action(...)`):
  ```python
  if __debug__:
      then_mod = [s.name for s in upds[0][0]] if upds[0][0] is not None else []
      else_mod = [s.name for s in upds[1][0]] if upds[1][0] is not None else []
      xtracer.trace("actions.IfAction.int_update then childType=%s thenModified=%s" % (type(branches[0]).__name__, sorted(then_mod)))
      xtracer.trace("actions.IfAction.int_update else childType=%s elseModified=%s" % (type(branches[1]).__name__, sorted(else_mod)))
  ```
  Same pattern in the Some path with `if_part`/`else_part`.

### Site 5 — `WhileAction.int_update`

**Python (ivy_actions.py:1081-1090):**
```python
def int_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.WhileAction.int_update ENTER")
    global context
    if isinstance(context,UnrollContext):
        if __debug__: xtracer.trace("actions.WhileAction.int_update EXIT")
        return self.unroll(context.card).int_update(domain,pvars)
    exp = self.expand(domain,pvars)
    res = exp.int_update(domain,pvars)
    if __debug__: xtracer.trace("actions.WhileAction.int_update EXIT")
    return res
```

**Go (actions/update.go:1502-1518):**
```go
func (a *WhileAction) IntUpdate(ctx *UpdateContext) *Update {
    xtracer.Trace("actions.WhileAction.int_update ENTER")
    defer xtracer.Trace("actions.WhileAction.int_update EXIT")
    var actCtx IActionContext
    if ctx.ActCfg != nil {
        actCtx = ctx.ActCfg.Context
    }
    if uc, ok := actCtx.(*UnrollContext); ok {
        unrolled, err := a.Unroll(uc.Card, nil)
        if err == nil {
            return IntUpdate(unrolled, ctx)
        }
    }
    expanded := a.Expand(ctx)
    return IntUpdate(expanded, ctx)
}
```

| Check                       | Python                    | Go (current)                | Status |
|-----------------------------|---------------------------|-----------------------------|--------|
| ENTER/EXIT trace            | emitted                   | emitted (deferred)          | ✓      |
| UnrollContext detection     | `isinstance`              | `actCtx.(*UnrollContext)`   | ✓      |
| Unroll call                 | `unroll(card).int_update` | `Unroll → IntUpdate`        | ✓      |
| **Unroll error path**       | **raises IvyError**       | **silent fallthrough to Expand** | **✗** |
| Expand path                 | `expand → int_update`     | `Expand → IntUpdate`        | ✓      |

**Fix:**
- (5a) Replace the silent `if err == nil { return ... }` with: if Unroll returns an error, propagate it via `panic(err.Error())`. Python's `unroll` raises `IvyError("cannot determine an iteration bound")` or `IvyError("cowardly refusing to unroll loop ...")` and does not fall back to expand.

### Site 6 — `LocalAction.int_update`

**Python (ivy_actions.py:1146-1159):**
```python
def int_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.LocalAction.int_update ENTER")
    update = self.args[-1].int_update(domain,pvars)
    if __debug__:
        body_type = type(self.args[-1]).__name__
        body_mod = [s.name for s in update[0]] if update[0] is not None else []
        xtracer.trace("actions.LocalAction.int_update bodyType=%s bodyModified=%s" % (body_type, sorted(body_mod)))
    syms = self.args[0:-1]
    res =  hide(syms,update)
    if __debug__: xtracer.trace("actions.LocalAction.int_update EXIT")
    return res
```

**Go (actions/update.go:1686-1721):**
```go
func (a *LocalAction) IntUpdate(ctx *UpdateContext) *Update {
    xtracer.Trace("actions.LocalAction.int_update ENTER")
    defer xtracer.Trace("actions.LocalAction.int_update EXIT")
    bodyAct := unwrapToAction(a.Body)
    if bodyAct == nil {
        return NullUpdate()
    }
    update := IntUpdate(bodyAct, ctx)
    if xtracer.Enabled { /* bodyType+bodyModified trace */ }
    var symsToHide []*lg.Const
    for _, local := range a.Locals {
        if c, ok := local.(*lg.Const); ok {
            symsToHide = append(symsToHide, c)
        } else {
            name := constName(local)
            if name != "" {
                symsToHide = append(symsToHide, lg.NewConst(name, lg.TopS))
            }
        }
    }
    if len(symsToHide) > 0 {
        update = Hide(symsToHide, update)
    }
    return update
}
```

| Check                       | Python                    | Go (current)                | Status |
|-----------------------------|---------------------------|-----------------------------|--------|
| ENTER/EXIT trace            | emitted                   | emitted (deferred)          | ✓      |
| body trace (mid-method)     | emitted                   | emitted                     | ✓      |
| **non-Action body**         | **AttributeError**        | **silent NullUpdate**       | **✗** |
| symbol collection           | `self.args[0:-1]` raw     | typed copy / `constName` fallback | ~ (verify) |
| **always call hide**        | **unconditional**         | **`if len > 0`**            | **✗** |

**Fixes:**
- (6a) Replace `if bodyAct == nil { return NullUpdate() }` with `panic(fmt.Sprintf("LocalAction.IntUpdate: body is not an Action: %T", a.Body))`.
- (6b) Remove the `if len(symsToHide) > 0` guard around `Hide`. Python always calls `hide(syms, update)` even with an empty list — the call is the identity but it canonicalizes the annotation pipeline. Always call `Hide(symsToHide, update)`.
- (6c) The `lg.NewConst(name, lg.TopS)` fallback for non-`*lg.Const` locals is suspicious — Python passes `self.args[0:-1]` directly to `hide` without coercion. Investigate during implementation: check what types appear in `a.Locals` in practice. If they are always `*lg.Const`, simplify the loop to `symsToHide = append(symsToHide, local.(*lg.Const))` and let a type assertion failure surface the bug. If non-Const types do appear, document why with a comment citing the construction site.

### Site 7 — `LetAction.int_update`

**Python (ivy_actions.py:1177-1185):**
```python
def int_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.LetAction.int_update ENTER")
    update = self.args[-1].int_update(domain,pvars)
    subst = dict((a.args[0].rep,a.args[1].rep) for a in self.args[0:-1])
    res = subst_action(update,subst)
    if __debug__: xtracer.trace("actions.LetAction.int_update EXIT")
    return res
```

**Go (actions/update.go:1727-1754):** Has ENTER/EXIT, returns NullUpdate on nil body, builds subst via `binding.Children() + constName`, conditionally calls `SubstAction(update, subst)`.

| Check                       | Python                    | Go (current)                | Status |
|-----------------------------|---------------------------|-----------------------------|--------|
| ENTER/EXIT trace            | emitted                   | emitted (deferred)          | ✓      |
| **non-Action body**         | **AttributeError**        | **silent NullUpdate**       | **✗** |
| subst construction          | direct `.rep`             | `Children() + constName`    | ✓ (semantically equivalent) |
| **always call subst_action**| **unconditional**         | **`if len > 0`**            | **✗** |

**Fixes:**
- (7a) Replace `if bodyAct == nil { return NullUpdate() }` with `panic(fmt.Sprintf("LetAction.IntUpdate: body is not an Action: %T", a.Body))`.
- (7b) Remove the `if len(subst) > 0` guard around `SubstAction`. Always call.

### Site 8 — `BindOldsAction.int_update`

**Python (ivy_actions.py:1294-1298):**
```python
def int_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.BindOldsAction.int_update ENTER")
    result = bind_olds_action(self.args[0].int_update(domain,pvars))
    if __debug__: xtracer.trace("actions.BindOldsAction.int_update EXIT")
    return result
```

**Go (actions/update.go:1760-1769):** Has ENTER/EXIT, returns NullUpdate on nil inner, otherwise `BindOldsUpdate(IntUpdate(inner, ctx))`.

| Check                       | Python                    | Go (current)                | Status |
|-----------------------------|---------------------------|-----------------------------|--------|
| ENTER/EXIT trace            | emitted                   | emitted (deferred)          | ✓      |
| **non-Action inner**        | **AttributeError**        | **silent NullUpdate**       | **✗** |
| `bind_olds_action` call     | direct                    | `BindOldsUpdate`            | ✓      |

**Fix:**
- (8a) Replace `if innerAct == nil { return NullUpdate() }` with `panic(fmt.Sprintf("BindOldsAction.IntUpdate: inner is not an Action: %T", a.Inner))`.

### Site 9 — `CallAction.int_update`

**Python (ivy_actions.py:1316-1378):**
```python
def get_callee(self):
    global context
    name = self.args[0].rep
    v = context.get(name)
    if not v:
        raise IvyError(self,"no value for {}".format(name))
    return v
def int_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.CallAction.int_update ENTER")
    v = self.get_callee()
    if not isinstance(v,tuple):
        if isinstance(v,Action):
            v = self.apply_actuals(domain,pvars,v)
        else:
            v = state_to_action(v.value)
    if __debug__: xtracer.trace("actions.CallAction.int_update EXIT")
    return v
```

`apply_actuals` builds `Sequence(input_asgns, BindOldsAction(callee), output_asgns)`, calls `int_update` on it, then `hide(formal_params+formal_returns, res)`.

**Go (actions/update.go:1776-1926):** Has ENTER/EXIT, looks up callee via `ctx.GetAction` then `ctx.Domain.Actions`, calls `applyActuals` which builds the same Sequence and Hides.

| Check                       | Python                    | Go (current)                | Status |
|-----------------------------|---------------------------|-----------------------------|--------|
| ENTER/EXIT trace            | emitted                   | emitted (deferred)          | ✓      |
| **callee name resolution**  | `args[0].rep`             | `constName(a.Callee)`; **silent NullUpdate on empty** | **✗** |
| **callee lookup failure**   | `raise IvyError`          | **silent NullUpdate** (lines 1799-1801) | **✗** |
| **tuple-callee path**       | `return v` directly       | **MISSING**                 | **✗** |
| **state-callee path**       | `state_to_action(v.value)`| **MISSING**                 | **✗** |
| Action-callee path          | `apply_actuals`           | `applyActuals`              | ✓      |
| param/return count check    | `raise IvyError`          | **silent NullUpdate** (1818, 1821) | **✗** |
| sort mismatch check         | `raise IvyError`          | **silent continue** (1880)  | **✗** |
| capture-avoidance renaming  | `distinct_obj_renaming`   | `distinctObjRenaming`       | ✓      |
| Sequence assembly           | `Sequence(in, BindOlds, out)` | same                    | ✓      |
| `hide(formals)` after       | unconditional             | `if len > 0` (1921)         | **✗** |

**Fixes:**
- (9a) Replace `if calleeName == "" { return NullUpdate() }` (1780-1782) with `panic(fmt.Sprintf("CallAction.IntUpdate: callee has no name: %T", a.Callee))`.
- (9b) Replace `if calleeAction == nil { return NullUpdate() }` (1799-1801) with `panic(fmt.Sprintf("CallAction.IntUpdate: no value for %s", calleeName))`. Mirrors Python's `IvyError("no value for {}")`.
- (9c) Investigate whether `ctx.Domain.Actions.Get2` can return non-`Action` values (matching Python's tuple- or state-valued context). If yes, port the `state_to_action` path; if no (i.e., Go always stores Action in `Domain.Actions`), document this as a known reduction and assert. **Default for this plan**: if `v, ok := ctx.Domain.Actions.Get2(...)` returns ok but the value is not an Action, panic with a clear "non-Action callee" message rather than silently doing nothing. Adding the full `state_to_action` port is out-of-scope unless the panic surfaces in TestOrdLive.
- (9d) Replace the param/return count silent NullUpdates (1817-1822) with panics matching Python's `IvyError("wrong number of input parameters")` / `"wrong number of output parameters"`.
- (9e) Replace the silent sort-mismatch comment-only branch (1880) with an actual panic matching Python's `IvyError("value for input parameter ... has wrong sort")`. Same for the symmetric output parameter check (which Python has at lines 1364-1369 of the Python source).
- (9f) Remove the `if len(toHide) > 0` guard around `Hide` at line 1921. Always call.

## Annotation handling deep-dive (ChoiceAction, EnvAction, IfAction)

The user explicitly called out these three for an annotation-handling audit. The shared concern: Python's `or_clauses` and `ite_clauses` build specific annotation types (`IteAnnotation` chains constructed via the renamer in `or_clauses_int`), and if Go's analogues build different types, downstream counter-example reconstruction will diverge.

### Python pipeline

- `join_action(s1, s2, axioms)` (transrel.py:285-286) → `join(s1, s2, new, axioms)` (transrel.py:196-208).
- Inside `join`: `c = or_clauses(c1, c2)` and `p = or_clauses(p1, p2)` (transrel.py:206-207). No `annot_op` is passed; `or_clauses` uses its built-in annotation reconstruction in `fix_or_annot`, which threads `IteAnnotation`s based on the Tseitin-like vs.
- `ite_action(cond, s1, s2, axioms)` (transrel.py:288-289) → `ite(cond, s1, s2, new, axioms)` (transrel.py:210-230).
- Inside `ite`: `c = ite_clauses(cond, [c1, c2])` (transrel.py:228). `ite_clauses` uses default behavior; the annotation type produced should be `IteAnnotation(cond, c1.annot, c2.annot)` (need to verify by reading `ite_clauses_int` in ivy_logic_utils.py:1391+).

### Go pipeline

- `JoinAction → joinUpdate` (transrel.go:834-869) calls `module.OrClausesTyped(c1, c2)` and `module.OrClausesTyped(p1, p2)`.
- `OrClausesTyped` (module/ops.go:215-265) routes through `fixOrAnnot` (line 264), which calls a package-level callback `AnnotIteFunc` (module/ops.go:269) that the actions package sets at init time.
- `IteAction → iteUpdate` (transrel.go:879-926) calls `module.IteClauses(cond, c1, c2)` (line 917).

### What needs verification (no code changes until verified)

1. **`AnnotIteFunc` registration**: confirm that the actions package init registers `AnnotIteFunc` and that the registered function constructs `*IteAnnotation` with the same shape Python builds (verify via `module.AnnotIteFunc` references and the fixOrAnnot loop).
2. **`module.IteClauses`**: confirm it returns a `*Clauses` with `Annot = *IteAnnotation{Cond: cond, ThenB: c1.Annot, ElseB: c2.Annot}` matching Python's `ite_clauses_int` annotation.
3. **The `fix_or_annot` reduction order**: Python's `fix_or_annot` left-folds `args[0]` with each subsequent `(args[i], vs[i])` as `IteAnnotation(vs[i], args[i].annot, accum)` — verify Go's `fixOrAnnot` performs the same fold direction with the same `vs[i]` parameter ordering.

If any of (1)-(3) is wrong, the fix is to add or correct the annotation construction. We will not write speculative fixes — only fix what verification proves wrong.

**Critical test**: After applying the fixes in Sites 2-4 above, run TestOrdLive. If the divergence shifts cleanly past the ChoiceAction/EnvAction/IfAction sites, the annotation pipeline is correct as-is. If a new divergence appears around `transrel.iteUpdate ... HASH canon=` traces or in counter-example reconstruction, dig into (1)-(3).

## Files to modify (summary)

### 1. `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_actions.py`

- **ChoiceAction.int_update** (line 882): change loop to `for i,a in enumerate(self.args):` and add per-branch `xtracer.trace("actions.ChoiceAction.int_update branch[%d] childType=%s childModified=%s" % ...)` after `result = join_action(...)`.
- **EnvAction.int_update** (line 910): same enumerate change and matching `actions.EnvAction.int_update branch[%d]` trace.
- **IfAction.int_update** (line 982): add `then`/`else` traces in both the simple-condition path (after `upds = [...]`) and the Some-path (after `if_part,else_part = (...)`).

No semantic Python changes; only trace additions.

### 2. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/action.go`

- **InstantiateAction.IntUpdate** (line 2144):
  - 2147-2149: remove `ctx.Domain == nil` early-return; replace with panic.
  - 2193-2195: remove `instName == ""` early-return; replace with panic.
  - 2208-2212: change `Modified: nil` to `Modified: []*lg.Const{}`.
  - 2216: replace `return NullUpdate()` with panic `"instantiation of undefined: %s"`.

### 3. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/update.go`

- **ChoiceAction.IntUpdate** (1360-1383):
  - 1372: replace `makeUpdate(...)` with explicit `&Update{Modified: []*lg.Const{}, TR: module.FalseClauses(EmptyAnnotation{}), Pre: module.FalseClauses(EmptyAnnotation{})}`.
  - 1374: change loop to `for i, branch := range a.Branches`.
  - 1376-1378: replace silent skip with panic.
  - After 1380: add per-branch xtracer.Trace.

- **EnvAction.IntUpdateEnv** (1389-1414): mirror ChoiceAction changes (initial annot, panic-on-non-Action, per-branch trace named `actions.EnvAction.int_update branch[%d]`).

- **IfAction.IntUpdate** (1420-1495):
  - After 1429 (free-vars panic): add `is_boolean(cond)` panic for the non-Some path. Implementation: look up `cond.NodeSort()` and panic if it is not `lg.Boolean`. Place this check inside the `if !isSome` branch (i.e., after the dispatch check at 1432) so the Some path is unaffected, mirroring Python which has the boolean check inside the non-Some branch.
  - After 1456 (`elseUpdate := IntUpdate(elseAct, ctx)`): add `then`/`else` xtracer.Trace lines.
  - In `intUpdateWithSubactions` (1464): add the same `then`/`else` traces after `elseUpdate := IntUpdate(elsePart, ctx)`.

- **WhileAction.IntUpdate** (1502-1518):
  - 1511-1514: replace silent fallthrough on Unroll error with panic.

- **LocalAction.IntUpdate** (1686-1721):
  - 1690-1692: replace silent NullUpdate with panic.
  - 1717-1719: remove `if len(symsToHide) > 0` guard; always call `Hide`.
  - 1706-1715: simplify the loop if locals are always `*lg.Const` (verify during implementation).

- **LetAction.IntUpdate** (1727-1754):
  - 1731-1733: replace silent NullUpdate with panic.
  - 1750-1752: remove `if len(subst) > 0` guard; always call `SubstAction`.

- **BindOldsAction.IntUpdate** (1760-1769):
  - 1764-1766: replace silent NullUpdate with panic.

- **CallAction.IntUpdate** (1776-1805):
  - 1780-1782: replace silent NullUpdate with panic on empty callee name.
  - 1799-1801: replace silent NullUpdate with panic on missing callee.
  - Optionally: handle non-Action stored in `Domain.Actions` via panic with descriptive message (do not port `state_to_action` unless verification surfaces the need).

- **CallAction.applyActuals** (1810-1926):
  - 1817-1822: replace silent NullUpdate count-mismatch with panic.
  - 1873-1885: replace silent sort-mismatch with panic on both input and output sides (mirror Python ivy_actions.py:1359-1369).
  - 1921-1923: remove `if len(toHide) > 0` guard; always call `Hide`.

### Optional helper

Add a small package-local helper to dedupe the per-branch trace formatting (`sortedModNames` or similar). Place it adjacent to the existing trace helpers in `actions/update.go`.

## Verification

1. Build: `cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go build ./...`
2. Run the failing golden test:
   ```sh
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy
   go test ./parser/ -run TestOrdLive -v 2>&1 | tee ~/ivy/goivy/log.red.new
   ```
3. Confirm the original `i=233572` and `i=233622` divergences are gone, OR that the divergence has shifted strictly later. The test framework's `golden_test.go:452` will print the new divergence line.
4. If the divergence shifts to a `transrel.iteUpdate ... HASH canon=` trace or to a `fix_or_annot`-related canon trace, perform the annotation deep-dive verification described above.
5. If a new divergence appears at `actions.ChoiceAction.int_update branch[%d]` or `actions.IfAction.int_update then/else`, the new traces have surfaced a deeper conformance gap — investigate as a separate finding (this is the trace framework's *purpose*).
6. Run broader regression: `go test ./...`
7. **Critical panic checks**: If any test newly panics with messages like `"InstantiateAction: instantiation of undefined"`, `"LocalAction.IntUpdate: body is not an Action"`, etc., that is a previously-hidden upstream bug surfacing. Investigate the upstream construction site rather than reverting the panic.

## Risks / considerations

- **Annotation type promotion** (Sites 2 and 3): Switching `ChoiceAction`/`EnvAction` initial annotation from `nil` to `EmptyAnnotation{}` may shift downstream traces and counter-example reconstruction in golden tests that exercise choice-heavy code. This is the *correct* behavior per Python source of truth. Verification step 3 explicitly looks for this.
- **Panics replacing silent NullUpdate**: every silent-skip-on-nil that we convert to a panic may surface a previously-hidden upstream construction bug. This is desirable (mechanical port faithfulness > silent papering-over). If any existing test passes only because of one of these silent skips, that test was masking a bug; we want it to fail loudly.
- **`is_boolean` check (Site 4)**: Python uses `is_boolean(self.args[0])`. Go's analog is `cond.NodeSort() == lg.Boolean`. There is a risk that `cond` after compilation has a different sort representation (e.g., a wrapped boolean type). If the panic fires on cases Python accepts, expand the check to handle the wrapper.
- **Unroll error path (Site 5)**: Python's `unroll` raises specifically when it can't determine an iteration bound or when the bound is too large. The Go `Unroll` error message should match closely so divergence reports are readable.
- **`state_to_action` (Site 9)**: Not porting this is a deliberate scope-limiting choice. If TestOrdLive exercises any test with a state-valued callee, the panic will surface immediately and we will port it as a follow-up.
- **Per-branch trace explosion**: adding per-branch traces in ChoiceAction/EnvAction/IfAction increases the trace volume in golden tests by O(branches per choice). This is the same trade-off Sequence already makes; trace volume is the cost of conformance detection.

## Out-of-scope / follow-ups (NOT in this plan)

- Audit of `applyActuals` deeper than the silent-NullUpdate paths (e.g., the renaming `oldOfOld` construction at update.go:1849-1851 — does it actually match Python's `subst[old(s)] = old(t)` for compound `old` symbols?).
- Audit of `Hide`, `SubstAction`, `BindOldsUpdate` annotation handling — these helpers are exercised by Sites 6-8 but their internal annotation pipelines are not part of the int_update audit.
- Porting `state_to_action` (deferred unless TestOrdLive surfaces a need).
- Audit of the `compile_action_body` callback wiring used by InstantiateAction — covered indirectly here but not exhaustively.
