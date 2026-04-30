# Fix XTRACE 212996 divergence: AddErrFlag must use DualFormula

Created: 2026-04-30 ~06:15 UTC

## Context

At XTRACE 212996, Go emits `mc.toaiger calling GetUpdate type=Sequence` while Python emits `ops.ToOpenFormula nFmlas=0 nDefs=0`. The traces match up to 212995 (`ivylogic.WithSorts.Enter nSorts=0`), so Go is missing a `ToOpenFormula` call that Python makes between `WithSorts.Enter` and `GetUpdate`.

## Root Cause

In Python's `add_err_flag` (ivy_mc.py:1021-1047), the AssertAction case uses `ilu.dual_formula(il.drop_universals(action.formula))` to compute the error condition. `dual_formula` (ivy_logic_utils.py:1616-1626) does three things:
1. Skolemizes free variables (prefix `__`)
2. Negates the formula  
3. If `lu.instantiator` is set (by `theory_context`), calls `instantiator(apps)` and conjoins the result via `clauses_to_formula(insts)` which calls `to_open_formula()` — producing the trace

Go's `AddErrFlag` (mc/toaiger.go:578-583) uses plain `&lg.Not{Body: ...}` instead of `module.DualFormula(...)`, skipping both the skolemization and the instantiator call entirely.

The call chain: `check_subgoals` → `TheoryContext()` (sets `mod.Instantiator`) → `CheckIsolate` → `ToAiger` → `AddErrFlagMod` → `AddErrFlag` for each action. By this point `mod.Instantiator` is set, so `DualFormula` will invoke it, triggering the `ToOpenFormula` trace that Python emits.

## Fix

**File: `mc/toaiger.go`**

1. Add `instantiator func([]lg.Expr) *module.Clauses` parameter to `AddErrFlag`:
   ```go
   func AddErrFlag(action actions.Action, erf *lg.Const, errConds *[]lg.Expr,
       instantiator func([]lg.Expr) *module.Clauses) actions.Action {
   ```

2. Change the AssertAction case (line 578-583) from:
   ```go
   errCond := &lg.Not{Body: il.DropUniversals(a.Formula)}
   ```
   to:
   ```go
   errCond := module.DualFormula(il.DropUniversals(a.Formula), nil, instantiator)
   ```
   This matches Python's `errcond = ilu.dual_formula(il.drop_universals(action.formula))`.

3. Thread `instantiator` through all recursive `AddErrFlag` calls in the Sequence, ChoiceAction, EnvAction, BindOldsAction, IfAction, and LocalAction cases.

4. In `AddErrFlagMod` (line 560-569), pass `mod.Instantiator` to `AddErrFlag`:
   ```go
   newAction := AddErrFlag(a, erf, errConds, mod.Instantiator)
   ```

5. In `ToAiger` (line 70), pass `mod.Instantiator` to `AddErrFlag` for the init action:
   ```go
   initAction := AddErrFlag(initSeq, erf, &errConds, mod.Instantiator)
   ```

## Files to modify

- `mc/toaiger.go` — `AddErrFlag` signature + AssertAction case + recursive calls + callers

## Existing functions to reuse

- `module.DualFormula` at `module/skolem.go:50` — already exists, matches Python's `dual_formula`
- `module.Clauses.ToOpenFormula` at `module/clauses.go:100` — called internally by `DualFormula` via `clausesToFormula`

## Verification

Run `cd ~/ivy/goivy && make test` — specifically the `TestRfnAbsIso` test should pass, with XTRACE 212996 now matching Python's `ops.ToOpenFormula nFmlas=0 nDefs=0`.
