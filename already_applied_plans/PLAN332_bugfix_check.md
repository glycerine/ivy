# Code Review: check/ and Helpers — Conformance Audit
*Created: 2026-04-17 (Thursday), 19:30 PT*

## Context

Comprehensive code review of `~/ivy/goivy/check/` (especially `l2s_hooks.go`,
`helpers.go`, `l2s.go`, `l2s_shared.go`) and helper directories against the
Python source of truth (`~/ivy/pyivy/ivy/ivy/ivy_l2s.py:1528-1702`,
`ivy_check.py:290-320`, `ivy_logic.py:313-317`). Found critical bugs in
`NewMatchHandler` and `extractJusticePredMap`, plus several format and
behavioral divergences.

Prior plan bugs (1-8) were already fixed. This plan covers NEW bugs only.

---

## Bug A (CRITICAL): `NewMatchHandler` drops bare-Const equalities

**File:** `check/helpers.go:186-208`

**Python** `ivy_logic.py:313-317`:
```python
def is_app(term):
    return (
        isinstance(term, App) or
        isinstance(term, lg.Const) or          # ← includes bare Const!
        isinstance(term, lg.NamedBinder) and len(term.variables) == 0
    )
```

Python's `MatchHandler.__init__` (`ivy_check.py:300-311`) uses `is_app(lhs)` to
decide whether to index an equality. Since `is_app(Const) == True`, bare-Const
equalities like `Eq(@V0, node_3)` ARE indexed in Python's `eqs` dict.

**Go** (`helpers.go:189-194`):
```go
case *lg.Eq:
    if app, ok := f.T1.(*lg.Apply); ok {   // ← only Apply!
        key := lg.Key(app.Func)
        h.Eqs[key] = append(h.Eqs[key], fmla)
    }
```

Go only handles `*lg.Apply`, NOT `*lg.Const`. Bare-Const equalities are silently
dropped. Same issue in the `*lg.Not` case (line 196-199) and the `default` case
(line 201-206).

**Impact:** `evalSkolemInHandler` cannot find values of individual (non-relation)
Skolem constants (e.g., `@V0 : node`). All diagnostic "Note:" messages in
`l2s_created`, `l2s_needed_when_start`, `l2s_work_preserved`,
`l2s_needed_are_frozen`, `l2s_progress_made`, and `l2s_sched_stable` may be
suppressed because `evalSkolems` returns nil.

**Fix:** In all three switch cases of `NewMatchHandler`, also handle `*lg.Const`:

```go
case *lg.Eq:
    if app, ok := f.T1.(*lg.Apply); ok {
        key := lg.Key(app.Func)
        h.Eqs[key] = append(h.Eqs[key], fmla)
    } else if cnst, ok := f.T1.(*lg.Const); ok {
        // Python: is_app(Symbol) is True; eqs[lhs.rep].append(fmla)
        key := lg.Key(cnst)
        h.Eqs[key] = append(h.Eqs[key], fmla)
    }
case *lg.Not:
    if app, ok := f.Body.(*lg.Apply); ok {
        key := lg.Key(app.Func)
        h.Eqs[key] = append(h.Eqs[key], &lg.Eq{T1: app, T2: &lg.Or{}})
    } else if cnst, ok := f.Body.(*lg.Const); ok {
        key := lg.Key(cnst)
        h.Eqs[key] = append(h.Eqs[key], &lg.Eq{T1: cnst, T2: &lg.Or{}})
    }
default:
    if app, ok := fmla.(*lg.Apply); ok {
        key := lg.Key(app.Func)
        h.Eqs[key] = append(h.Eqs[key], &lg.Eq{T1: app, T2: &lg.And{}})
    } else if cnst, ok := fmla.(*lg.Const); ok {
        key := lg.Key(cnst)
        h.Eqs[key] = append(h.Eqs[key], &lg.Eq{T1: cnst, T2: &lg.And{}})
    }
```

---

## Bug B (CRITICAL): `extractJusticePredMap` — missing subs/rsubs, wrong formula navigation

**Files:** `check/l2s_hooks.go:147-193`, `check/l2s.go:898-904`, `check/l2s_shared.go:60-66`

**Python** `ivy_l2s.py:1541-1552`:
```python
rsubs = dict((x,y) for (y,x) in subs.items())   # built at line 1531
...
justice_pred_map = dict()
for fc in fcs:
    lf = fc.lf
    if lf.name.startswith('l2s_progress_invar'):
        sfx = lf.name[len('l2s_progress_invar'):]
        gfmla = rsubs[lf.formula.args[1].rep]     # (1) resolve nonce → NamedBinder
        gargs = lf.formula.args[1].args
        jfmla = gfmla.body.args[0]                 # (2) navigate NamedBinder body
        jfmla = subs[jfmla.rep]                     # (3) resolve back to nonce
        justice_pred_map[sfx] = jfmla
```

Python uses `rsubs` (nonce → NamedBinder) and `subs` (NamedBinder → nonce) to
navigate through the substituted formula. It:
1. Gets T2 of l2s_progress_invar's Implies, which after SharedStep11 is
   `Apply(l2s_g_N, args)` — a nonce Apply
2. Uses `rsubs` to recover the original l2s_g NamedBinder (captured before
   SharedStep11, body still has NamedBinders)
3. Navigates the body: `Not(Apply(l2s_g_inner_NB, args))` → `body.args[0]`
   = inner l2s_g NamedBinder
4. Uses `subs` to get the nonce for the inner binder

**Go problems:**

1. `applyAutoDiagnosticsToHandler` does NOT receive `subs` (line 108-112):
   ```go
   func applyAutoDiagnosticsToHandler(
       handler *MatchHandler,
       fcs []Checker,
       tasks map[string]map[string]*lg.Eq,
       triggers map[string]map[string]*lg.Eq,
   )
   ```
   The closure in `l2s.go:898-904` captures `subs` but only passes it to
   `applyRenamingToHandler`, not to `applyAutoDiagnosticsToHandler`.

2. `extractJusticePred` (line 175-193) navigates the SUBSTITUTED formula
   expecting `Implies(..., Eq(Apply(...), ...))` but after SharedStep11 the
   formula is `Implies(..., Apply(nonce, args))` — T2 is Apply, not Eq.
   `extractJusticePred` always returns nil.

3. `cfg.Subs` is `map[string]string` (nonce name → binder.String()). This
   only gives the string representation, not the actual NamedBinder object
   or the subs map keyed by Sexp.

**Impact:** `justicePredMap` is always empty. `justiceMap` is never built.
Diagnostic loop 2 for l2s_progress_made ("helpful condition is eventually
enabled but not satisfied") never fires.

**Fix (multi-step):**

1. Add fields to `InstrumentationConfig` (`l2s_shared.go`):
   ```go
   // RSubs maps nonce const name → original NamedBinder (inverse of subs).
   RSubs map[string]*lg.NamedBinder
   // FullSubs maps binder.Sexp() → nonce Const (the SharedStep11 subs map).
   FullSubs map[string]lg.Expr
   ```

2. Populate them in `SharedStep11_ReplaceNamedBinders` (`l2s_shared.go:797-823`):
   ```go
   cfg.RSubs = make(map[string]*lg.NamedBinder)
   cfg.FullSubs = subs  // the local map[string]lg.Expr
   for k, binders := range namedBinders.All() {
       for i, b := range binders {
           freshName := fmt.Sprintf("%s_%d", k, i)
           subs[string(b.Sexp())] = lg.NewConst(freshName, b.NodeSort())
           cfg.Subs[freshName] = b.String()
           cfg.RSubs[freshName] = b           // NEW
       }
   }
   ```

3. Pass through closure in `l2s.go:893-904`:
   ```go
   rsubs := cfg.RSubs
   fullSubs := cfg.FullSubs
   // ...
   applyAutoDiagnosticsToHandler(handler, fcs, tasks, triggers, rsubs, fullSubs)
   ```

4. Add parameters to `applyAutoDiagnosticsToHandler` (`l2s_hooks.go:108`):
   ```go
   func applyAutoDiagnosticsToHandler(
       handler *MatchHandler,
       fcs []Checker,
       tasks map[string]map[string]*lg.Eq,
       triggers map[string]map[string]*lg.Eq,
       rsubs map[string]*lg.NamedBinder,
       fullSubs map[string]lg.Expr,
   )
   ```

5. Rewrite `extractJusticePredMap` to use rsubs/fullSubs like Python:
   ```go
   func extractJusticePredMap(fcs []Checker, rsubs map[string]*lg.NamedBinder,
       fullSubs map[string]lg.Expr) map[string]*lg.Const {
       result := make(map[string]*lg.Const)
       for _, fc := range fcs {
           fcLF := fc.GetLF()
           if fcLF == nil { continue }
           fcName := lfName(fcLF)
           if !strings.HasPrefix(fcName, "l2s_progress_invar") { continue }
           sfx := fcName[len("l2s_progress_invar"):]
           // Python: gfmla = rsubs[lf.formula.args[1].rep]
           impl, ok := fcLF.Formula.(*lg.Implies)
           if !ok { continue }
           app, ok := impl.T2.(*lg.Apply)
           if !ok { continue }
           nonceName := applyFuncName(app)
           if nonceName == "" { continue }
           gfmla, ok := rsubs[nonceName]
           if !ok || gfmla == nil { continue }
           // Python: jfmla = gfmla.body.args[0]
           notExpr, ok := gfmla.Body.(*lg.Not)
           if !ok { continue }
           innerApp, ok := notExpr.Body.(*lg.Apply)
           if !ok { continue }
           // innerApp.Func is a NamedBinder (pre-substitution)
           innerNB, ok := innerApp.Func.(*lg.NamedBinder)
           if !ok { continue }
           // Python: jfmla = subs[jfmla.rep]  (jfmla.rep = NamedBinder)
           key := string(innerNB.Sexp())
           nonceExpr, ok := fullSubs[key]
           if !ok { continue }
           if c, ok := nonceExpr.(*lg.Const); ok {
               result[sfx] = c
           }
       }
       return result
   }
   ```

6. Delete `extractJusticePred` (no longer needed).

---

## Bug C (FORMAT): `l2s_sched_exists` format string diverges

**File:** `check/l2s_hooks.go:708`

**Go:**
```go
fmt.Printf("The helpful set(s) %s have become empty, but termination has not occurred.\n",
```

**Python** (`ivy_l2s.py:1699`):
```python
print ('The helpful set(s)  {} have become empty, but termination has not occurred'.format(rank_names))
```

Differences:
1. Python has **double space** after `set(s)` — Go has single space
2. Python has **no period** at end — Go has period
3. Both produce one trailing `\n` (Python from print, Go explicit)

**Fix:** Change to:
```go
fmt.Printf("The helpful set(s)  %s have become empty, but termination has not occurred\n",
```

---

## Bug D (FORMAT): `l2s_created` second message missing trailing newline

**File:** `check/l2s_hooks.go:383`

**Go:**
```go
fmt.Println("and its argument(s) are not visited during the action execution.")
```

**Python** (`ivy_l2s.py:1568`):
```python
print ('and its argument(s) are not visited during the action execution.\n')
```

Python outputs: text + `\n` (explicit) + `\n` (print) = two trailing newlines.
Go outputs: text + `\n` (Println) = one trailing newline.

**Fix:**
```go
fmt.Printf("and its argument(s) are not visited during the action execution.\n\n")
```

---

## Bug E (BEHAVIORAL): `l2s_full` trace hook applies renaming Python doesn't

**File:** `check/l2s.go:916-924`

**Python:** `l2s_tactic_full` (ivy_l2s.py:101-104) calls `l2s_tactic` (which
sets `trace_hook = renaming_hook`) then **overwrites** with
`goals[0].trace_hook = trace_hook` (markLoopStart only, line 103). Final
trace_hook for l2s_full = markLoopStart only. No renaming.

**Go:**
```go
default:
    isFull := tacticName == "l2s_full"
    result[0].TraceHook = TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
        if subs != nil {
            applyRenamingToHandler(handler, subs)  // ← Python doesn't do this for l2s_full
        }
        if isFull {
            markLoopStart(handler)
        }
    })
```

Go applies renaming + markLoopStart for l2s_full. Python only applies markLoopStart.

**Fix:** Split the `default:` case:
```go
default:
    if tacticName == "l2s_full" {
        result[0].TraceHook = TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
            markLoopStart(handler)
        })
    } else {
        result[0].TraceHook = TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
            if subs != nil {
                applyRenamingToHandler(handler, subs)
            }
        })
    }
```

---

## Bug F (ORDER): `l2s_not_all_done`/`l2s_sched_exists` random iteration order

**File:** `check/l2s_hooks.go:687-712`

**Go:** `for sfx, task := range tasks` iterates Go map in random order.
**Python:** `for sfx in tasks` iterates dict in insertion order (Python 3.7+).

Rank names may appear in different order in diagnostic output.

**Fix:** Sort the task suffixes before iterating:

```go
case strings.HasPrefix(name, "l2s_not_all_done"):
    var sfxs []string
    for sfx := range tasks {
        sfxs = append(sfxs, sfx)
    }
    sort.Strings(sfxs)
    var rankNames []string
    for _, sfx := range sfxs {
        if tasks[sfx]["work_needed"] != nil {
            rankNames = append(rankNames, "work_needed"+sfx)
        }
    }
    // ... same for l2s_sched_exists
```

Note: Python uses insertion order, not alphabetical. But alphabetical sort
is a safe deterministic approximation since suffixes are typically `""`, `_1`,
`_2`, etc. which sort correctly alphabetically.

---

## Summary of Required Code Changes

### File: `check/helpers.go`
- **Lines 189-206:** Add `*lg.Const` handling in all three switch cases of
  `NewMatchHandler` (Bug A)

### File: `check/l2s_shared.go`
- **Lines 60-66:** Add `RSubs` and `FullSubs` fields to
  `InstrumentationConfig` (Bug B)
- **Lines 797-823:** Populate `cfg.RSubs` and `cfg.FullSubs` in
  `SharedStep11_ReplaceNamedBinders` (Bug B)

### File: `check/l2s.go`
- **Lines 893-904:** Pass `rsubs`/`fullSubs` through closure to
  `applyAutoDiagnosticsToHandler` (Bug B)
- **Lines 916-924:** Split `default:` trace hook case for l2s_full (Bug E)

### File: `check/l2s_hooks.go`
- **Lines 108-141:** Add `rsubs`/`fullSubs` parameters to
  `applyAutoDiagnosticsToHandler` (Bug B)
- **Lines 137-138:** Update `extractJusticePredMap` call signature (Bug B)
- **Lines 147-193:** Rewrite `extractJusticePredMap` and delete
  `extractJusticePred` (Bug B)
- **Line 383:** Fix trailing newline in l2s_created message (Bug D)
- **Lines 687-695:** Sort task suffixes in l2s_not_all_done (Bug F)
- **Lines 701-712:** Sort task suffixes in l2s_sched_exists; fix format
  string (Bugs C, F)

---

## Verification

```
cd ~/ivy/goivy && make test
```
