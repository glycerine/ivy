# PLAN197: module/ Port Bugs — Post-Merge Re-Audit Fixes

**Created:** 2026-04-04 ~21:00 UTC

## Context

After merging clauseops/ into module/ (PLAN195) and verifying all 5 PLAN196 port audit
fixes are intact, a fresh function-by-function comparison of `module/ops.go` and
`module/clauses.go` against Python `ivy_logic_utils.py` revealed 7 new bugs where
Go diverges from Python behavior.

## Bugs

### Bug A: `iteClausesInt` drops annotation (MEDIUM)

**File:** `module/ops.go:405`
**Go:** `return NewClauses(fmlas, defs, nil)` — always nil annotation
**Python** (ivy_logic_utils.py:1373,1391-1392):
```python
a0,a1 = args[0].annot,args[1].annot
annot = None if a0 is None or a1 is None else a0.ite(v,a1)
res = Clauses(fmlas,defs,annot)
```
**Impact:** Annotations silently lost during ite_clauses operations.
**Fix:** Before the return, compute annot:
```go
var annot interface{}
a0, a1 := args[0].Annot, args[1].Annot
if a0 != nil && a1 != nil && AnnotIteFunc != nil {
    annot = AnnotIteFunc(a0, v, a1)
}
return NewClauses(fmlas, defs, annot)
```

### Bug B: `elimDeadDefinitions` uses same rename map for all args (MEDIUM)

**File:** `module/ops.go:768-776`
**Go:** Builds subs map ONCE, applies same map to ALL args.
**Python** (ivy_logic_utils.py:1334):
```python
args = [rename_symbols(rn,arg,to_rename) for arg in args]
```
Each call to `rename_symbols` calls `s.rename(rn)` fresh per arg, generating DIFFERENT
fresh names per arg (rn is stateful). This separates captured skolems across args.
**Impact:** Captured skolem definitions not properly separated across args.
Go renames `__sk -> __sk_fresh` in ALL args (shared), Python renames
`__sk -> __sk_fresh0` in arg0, `__sk -> __sk_fresh1` in arg1 (independent).
**Fix:** Move subs construction inside the per-arg loop:
```go
for i, a := range args {
    subs := make(map[lg.NodeKey]*lg.Const, len(toRename))
    for _, sym := range toRename {
        newName := rn.Rename(sym.Name)
        subs[lg.Key(sym)] = lg.NewConst(newName, sym.CSort)
    }
    args[i] = RenameClauses(a, subs)
}
```

### Bug C: `ToFormula` uses `CloseFormula` instead of `CloseEPR` (MEDIUM)

**File:** `module/clauses.go:111`
**Go:** `return il.CloseFormula(c.ToOpenFormula())` — wraps entire formula in single ForAll
**Python** (ivy_logic_utils.py:100): `return close_epr(self.to_open_formula())`
`close_epr` distributes through And — universally quantifies each conjunct separately.
**Impact:** Generated formulas have different quantifier structure.
**Fix:** Change to `return lu.CloseEPR(c.ToOpenFormula())`. The `lu` import already
exists in `clauses.go`.

### Bug D: `NegateClauses` fallback vs assert (LOW)

**File:** `module/ops.go:410-415`
**Go:** Falls back to formula conversion if `!IsUniversalFirstOrder()`.
**Python:** `assert clauses.is_universal_first_order()` — crashes.
**Impact:** Minor — silently handles invalid input instead of failing fast.
**Fix:** Replace the fallback with `panic("NegateClauses requires universal first-order clauses")`.

### Bug E: `dualClauses` missing skolemizer parameter (MEDIUM)

**File:** `module/ops.go:422`
**Go:** Hardcodes `"__"+v.Name` as the skolemization pattern.
**Python:** `dual_clauses(clauses, skolemizer=None)` — has optional parameter.
Callers with custom skolemizer: `sidecar.py:280`, `ivy_ui_cti.py:147,314,551,598`,
`ivy_bmc.py:39`, `ivy_check.py:224`, `ivy_trace.py:397`.
**Impact:** Go can't support custom skolemization strategies used by BMC, check, trace.
**Fix:** Add `skolemizer func(*lg.Variable) lg.Expr` parameter to `dualClauses` and
exported `DualClauses`. Pass `nil` for default behavior (`"__"+v.Name`).

### Bug F: `dualClauses`/`unfoldDefinitionsClauses` missing instantiator (LOW-MEDIUM)

**File:** `module/ops.go:422-439`
**Python:** Global `instantiator` (set by ivy_module.py:1009-1010) adds definition
instances during dual/unfold operations.
**Go:** No equivalent.
**Impact:** Definition instances not added during dual operations.
**Fix:** Add `Instantiator` field to `module.Config` or `OpsConfig`. Defer until
ivy_module porting is further along — the instantiator is only set by Module.__enter__.

### Bug G: `Conjuncts()` missing assert (LOW)

**File:** `module/clauses.go:116`
**Python:** `assert self.defs == []` before computing conjuncts.
**Go:** No assertion — silently proceeds even with defs present.
**Fix:** Add `if len(c.Defs) > 0 { panic("Conjuncts requires no definitions") }`.

## Execution Plan

### Step 1: Bug C — one-line fix in clauses.go

Change `module/clauses.go:111`:
```go
// before
return il.CloseFormula(c.ToOpenFormula())
// after
return lu.CloseEPR(c.ToOpenFormula())
```

### Step 2: Bug G — one-line fix in clauses.go

Add assert at `module/clauses.go:117` (inside `Conjuncts()`):
```go
if len(c.Defs) > 0 {
    panic("Conjuncts requires no definitions")
}
```

### Step 3: Bug A — fix iteClausesInt annotation in ops.go

Replace `module/ops.go:405` `return NewClauses(fmlas, defs, nil)` with:
```go
// Compute annotation matching Python: annot = None if a0 is None or a1 is None else a0.ite(v,a1)
var annot interface{}
a0, a1 := args[0].Annot, args[1].Annot
if a0 != nil && a1 != nil && AnnotIteFunc != nil {
    annot = AnnotIteFunc(a0, v, a1)
}
return NewClauses(fmlas, defs, annot)
```

### Step 4: Bug B — fix elimDeadDefinitions per-arg rename in ops.go

Replace `module/ops.go:768-776` (the single-map block) with per-arg renaming:
```go
if len(toRename) > 0 {
    for i, a := range args {
        subs := make(map[lg.NodeKey]*lg.Const, len(toRename))
        for _, sym := range toRename {
            newName := rn.Rename(sym.Name)
            subs[lg.Key(sym)] = lg.NewConst(newName, sym.CSort)
        }
        args[i] = RenameClauses(a, subs)
    }
}
```

### Step 5: Bug D — fix NegateClauses assert in ops.go

Replace `module/ops.go:411-415` fallback block with:
```go
if !clauses.IsUniversalFirstOrder() {
    panic("NegateClauses requires universal first-order clauses")
}
```

### Step 6: Bug E — add skolemizer parameter to dualClauses

Change signature of `dualClauses` and add exported `DualClauses`:
```go
type Skolemizer func(v *lg.Variable) lg.Expr

func DualClauses(clauses *Clauses, skolemizer Skolemizer) *Clauses {
    return dualClauses(clauses, skolemizer)
}

func dualClauses(clauses *Clauses, skolemizer Skolemizer) *Clauses {
    vars := UsedVariablesOrdered(clauses)
    subs := make(map[lg.NodeKey]lg.Expr, len(vars))
    for _, v := range vars {
        if skolemizer != nil {
            subs[lg.Key(v)] = skolemizer(v)
        } else {
            sk := lg.NewConst("__"+v.Name, v.VSort)
            subs[lg.Key(v)] = sk
        }
    }
    ...
}
```
Update `NegateClauses` call: `return dualClauses(clauses, nil)`.

### Step 7: Bug F — deferred (instantiator)

Not implemented now. Add TODO comment in `dualClauses`:
```go
// TODO: Python checks instantiator != None here and adds definition instances.
// Deferred until ivy_module porting provides the Instantiator callback.
```

### Step 8: Add tests for bugs A-E

In `module/port_audit_test.go`, add:
- `TestIteClausesIntAnnotation` — verify annotation is preserved when both args have non-nil annots
- `TestElimDeadDefinitionsPerArgRename` — verify each arg gets different fresh names for captured skolems
- `TestToFormulaCloseEPR` — verify And conjuncts are quantified separately
- `TestNegateClausesPanic` — verify panic on non-universal-first-order input
- `TestDualClausesCustomSkolemizer` — verify custom skolemizer is applied

### Step 9: Verify

```
go build ./...
go test ./module/...
go test ./...
```

## Critical Files to Modify

- `module/clauses.go` — Bugs C, G
- `module/ops.go` — Bugs A, B, D, E, F
- `module/port_audit_test.go` — new tests

## Verification

1. `go build ./...` — compiles clean
2. `go test ./module/...` — all existing + new tests pass
3. `go test ./...` — full suite passes
