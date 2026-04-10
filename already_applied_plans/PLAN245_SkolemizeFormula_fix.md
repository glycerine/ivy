# Fix AssumeAction.action_update Trace Divergence (missing ToOpenFormula call)

Created: 2026-04-09 (afternoon)

## Context

`TestOrdLive` (golden trace conformance) fails at trace line 220222 with this divergence:

```
220221  go : XTRACE: actions.AssumeAction.action_update ENTER
        py : XTRACE: actions.AssumeAction.action_update ENTER

220222  go : XTRACE: actions.AssumeAction.action_update EXIT
        py : XTRACE: ops.ToOpenFormula nFmlas=0 nDefs=0
```

Python emits `ops.ToOpenFormula nFmlas=0 nDefs=0` between ENTER and EXIT of `AssumeAction.action_update`. Go skips it entirely and exits immediately. The `nFmlas=0 nDefs=0` is the giveaway: Python is calling `to_open_formula()` on an *empty* `Clauses` object — i.e., the result of `instantiator(gts)` returning an empty Clauses. Python's source-of-truth code calls `clauses_to_formula(instantiator(gts))` *unconditionally* (no empty-guard), but Go has two independent bugs that suppress the call:

1. **Go's `ctx.Instantiator` is nil at this point** because `interp.ApplyAction` (and other `UpdateContext` constructors) does not propagate `state.Domain.Instantiator` into `ctx.Instantiator`. So `SkolemizeFormula` is invoked with a nil instantiator and the entire `if instantiator != nil` branch is skipped.

2. **Go's `SkolemizeFormula`/`DualFormula`/`dualClauses` add a non-Python guard** `if len(insts.Fmlas) > 0 { ... }` around the `clausesToFormula(insts)` call. Even with the instantiator wired up, when the instantiator returns an empty Clauses (the common case for AssumeAction whose formula references no non-EPR symbols), Go's guard skips both the `clausesToFormula` call and the `And(fmla, insts)` construction. Python has no such guard.

Either bug alone would suppress the trace; both must be fixed.

## Source-of-truth Python references

`~/ivy/pyivy/ivy/ivy/ivy_actions.py:345-358` — `AssumeAction.action_update`:
```python
def action_update(self,domain,pvars):
    if __debug__: xtracer.trace("actions.AssumeAction.action_update ENTER")
    fmla = self.args[0]
    if isinstance(fmla,ivy_ast.LabeledFormula):
        if fmla.unprovable:
            if __debug__: xtracer.trace("actions.AssumeAction.action_update EXIT")
            return ([],true_clauses(annot=EmptyAnnotation()),false_clauses(annot=EmptyAnnotation()))
        fmla = fmla.formula
    type_check(domain,fmla)
    clauses = formula_to_clauses_tseitin(skolemize_formula(fmla))   # <-- enters skolemize_formula
    clauses = unfold_definitions_clauses(clauses)
    clauses = Clauses(clauses.fmlas,clauses.defs,EmptyAnnotation())
    if __debug__: xtracer.trace("actions.AssumeAction.action_update EXIT")
    return ([],clauses,false_clauses(annot=EmptyAnnotation()))
```

`~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py:1579-1592` — `skolemize_formula` (no guard on `insts`):
```python
def skolemize_formula(fmla, skolemizer=None):
    ...
    if instantiator != None:
        gts = apps_ast(fmla)
        insts = clauses_to_formula(instantiator(gts))   # ALWAYS called
        fmla = And(fmla,insts)                          # ALWAYS conjoined
    return fmla
```

`~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py:1567-1577` — `dual_formula` (same unconditional pattern):
```python
def dual_formula(fmla, skolemizer=None):
    ...
    if instantiator != None:
        gts = apps_ast(fmla)
        insts = clauses_to_formula(instantiator(gts))   # ALWAYS called
        fmla = And(fmla,insts)                          # ALWAYS conjoined
    return fmla
```

`~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py:1554-1565` — `dual_clauses` (same unconditional pattern):
```python
def dual_clauses(clauses, skolemizer=None):
    ...
    fmla = negate(clauses_to_formula(clauses))
    if instantiator != None:
        gts = apps_clauses(clauses)
        insts = instantiator(gts)
        fmla = And(fmla,clauses_to_formula(insts))      # ALWAYS conjoined
    return formula_to_clauses(fmla)
```

`~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py:1546-1552` — `unfold_definitions_clauses` (this one *does* have a guard, and Go matches it correctly):
```python
def unfold_definitions_clauses(clauses):
    if instantiator != None:
        gts = apps_clauses(clauses)
        insts = instantiator(gts)
        if insts.fmlas:                                 # GUARDED
            clauses = and_clauses(clauses,insts)
    return clauses
```

`~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py:94-98` — `Clauses.to_open_formula` (the trace source):
```python
def to_open_formula(self):
    if __debug__:
        xtracer.trace("ops.ToOpenFormula nFmlas=%d nDefs=%d" % (len(self.fmlas), len(self.defs)))
    conjuncts = [dfn.to_constraint() for dfn in self.defs] + self.fmlas
    return And(*conjuncts)
```

`~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py:1011-1015` — `clauses_to_formula` calls `to_formula` → `to_open_formula`:
```python
def clauses_to_formula(cs):
    formula = cs.to_formula()      # to_formula calls to_open_formula → emits ops.ToOpenFormula trace
    if __debug__: xtracer.trace("ivy_logic_utils.py:1013/clauses_to_formula(): HASH canon= formula = %s" % formula.canon())
    return drop_universals(formula)
```

The instantiator is set globally by `with im.module.theory_context():` in `~/ivy/pyivy/ivy/ivy/ivy_check.py:535`, which sets `lu.instantiator = ModuleTheoryContext(non_epr)`. Go's equivalent is `cleanupTheory := mod.TheoryContext()` in `check/isolate_check.go:107`, which sets `m.Instantiator = func(...){ ... }`. So when `interp.ApplyAction` is reached during the failing trace, `state.Domain.Instantiator` IS non-nil — Go just isn't reading it.

## Files to modify

### Fix A — Remove non-Python empty-guards in module/

**`/Users/jaten/ivy/goivy/module/skolem.go`** (two functions):

1. Lines 65-71 in `DualFormula`:
   ```go
   if instantiator != nil {
       gts := il.AppsAst(fmla)
       insts := instantiator(gts)
       if len(insts.Fmlas) > 0 {                                            // DELETE GUARD
           fmla = &lg.And{Terms: []lg.Expr{fmla, clausesToFormula(insts)}}
       }
   }
   ```
   Change to (matching Python's unconditional `And(fmla, clauses_to_formula(insts))`):
   ```go
   if instantiator != nil {
       gts := il.AppsAst(fmla)
       insts := instantiator(gts)
       fmla = &lg.And{Terms: []lg.Expr{fmla, clausesToFormula(insts)}}
   }
   ```

2. Lines 102-108 in `SkolemizeFormula`: same removal of `if len(insts.Fmlas) > 0` guard.

**`/Users/jaten/ivy/goivy/module/ops.go`** (one function):

3. Lines 497-503 in `dualClauses`: same removal of `if len(insts.Fmlas) > 0` guard.

After these three edits the trace `ops.ToOpenFormula` will be emitted whenever the instantiator is non-nil, exactly matching Python's behavior. Note that `unfold_definitions_clauses` (Go: `module/skolem.go:346 UnfoldDefinitionsClauses`) DOES have an `if insts.fmlas:` guard in Python and Go's matching guard at line 352 is correct — leave it alone.

### Fix B — Plumb the instantiator from `Domain.Instantiator` into `UpdateContext`

The `UpdateContext.Instantiator` field exists at `actions/update.go:61` but no constructor populates it. Python's `lu.instantiator` is a module-level global; the equivalent in Go lives on `module.Module.Instantiator` (set by `Module.TheoryContext()` in `module/theory.go:213-216`). Every `UpdateContext` constructor that has a `Domain *module.Module` must copy `Domain.Instantiator` into `ctx.Instantiator`.

The minimum site needed for this specific failure is `interp/eval.go:151` (`interp.ApplyAction` is on the failing trace path: `art.Execute → interp.ApplyAction(EnvAction) → actions.GetUpdate → ... → AssumeAction.ActionUpdate`). However, every constructor should be fixed to keep the port faithful.

**Constructors to update:**

| File:line | Function | Action |
|---|---|---|
| `interp/eval.go:151` | `ApplyAction` | add `Instantiator: state.Domain.Instantiator,` |
| `interp/phase4.go:186` | (decompose path) | add `Instantiator: state1.Domain.Instantiator,` |
| `fragment/fragment.go:1080` | (fragment ctx) | add `Instantiator: m.Instantiator` (struct literal expansion) |
| `actions/update.go:2032` | `GetUpdateForArt` | add `Instantiator: domain.Instantiator,` |
| `actions/action.go:1404` | `WhileAction.DecomposeWithModule` | add `Instantiator: m.Instantiator` |
| `actions/match.go:362` | match-action loop | add `Instantiator: domain.Instantiator,` (verify Domain is non-nil) |
| `actions/phase3.go:434` | phase3 ctx | add `Instantiator: <module>.Instantiator,` |
| `mc/toaiger.go:122` | mc ctx | add `Instantiator: <module>.Instantiator,` |

(The test ctx in `actions/update_test.go:11` `testCtx()` uses `module.New()` whose `.Instantiator` is nil, which is fine — current tests do not exercise this path.)

### Fix C — Pass instantiator into `DualFormula` from `AssertAction.ActionUpdate`

`actions/update.go:378` currently calls:
```go
dual := module.DualFormula(fmla, nil, nil)   // <-- last arg should be ctx.Instantiator
```

Python's `AssertAction.action_update` (ivy_actions.py:391) calls `dual_formula(fmla)`, which internally consults `lu.instantiator`. Go's literal port must pass `ctx.Instantiator` as the third argument so `DualFormula` can conjoin the definition instances. Change to:
```go
dual := module.DualFormula(fmla, nil, ctx.Instantiator)
```

This is required for Python parity even though it is not directly on the current failing trace line — without it the next AssertAction trace will diverge.

## Verification

1. Build/test the affected packages first (sanity check the code compiles):
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy
   go build ./module/... ./actions/... ./interp/... ./art/... ./check/... ./fragment/... ./mc/...
   ```

2. Run the existing module unit tests that cover the touched functions (all should still pass since they exercise the non-empty-Clauses path):
   ```
   go test ./module/ -run 'TestDualFormulaInstantiator|TestSkolemizeFormulaInstantiator|TestClausesToOpenFormula|TestClausesToFormula|TestToFormulaCloseEPR' -count=1
   ```

3. Run the action unit tests that build `UpdateContext` literally:
   ```
   go test ./actions/ -run 'TestAssumeAction|TestAssertAction' -count=1
   ```

4. Re-run the failing golden test that produced `log.red`:
   ```
   go test ./parser/ -run TestOrdLive -count=1 -timeout 600s
   ```
   The new trace at line 220222 should be `ops.ToOpenFormula nFmlas=0 nDefs=0`, matching Python. The trace immediately after should be `logicutil.CloseEPR HASH canon= fmla=...` followed by `ivy_logic_utils.py:1013/clauses_to_formula(): HASH canon= formula = ...`, then continue back into the AssumeAction body and finally emit `actions.AssumeAction.action_update EXIT`. If there is a *new* divergence further down, capture it in `log.red` and write a follow-up plan.

5. Inspect `log.red` after the test runs to verify the divergence has moved past line 220222 (or, ideally, the test passes outright).

## Notes / Why this is the right fix

- This is purely a faithful-port correction: three Go locations added a non-Python `if insts.Fmlas > 0` guard, and one Go integration point (UpdateContext construction) failed to plumb a value that Python gets from a module global. Both are creative deviations from the source-of-truth Python; both must go.
- After the fix, `clausesToFormula(emptyClauses)` will produce `&lg.And{Terms: []}` (empty And), which `dropUniversals` returns unchanged. This matches Python where `clauses_to_formula(empty_clauses)` returns `And()`. The resulting `fmla = And(originalFmla, And())` is structurally the same in both languages.
- Per CLAUDE.md rule C ("NO GLOBAL VARIABLES"), the Go side stores the instantiator on `module.Module.Instantiator` (set by `Module.TheoryContext()`). Reading it from `state.Domain.Instantiator` in the UpdateContext constructors is the correct port of Python's `lu.instantiator` global.
- Per CLAUDE.md rule B.9 ("Deeply pursue the goal"): all eight UpdateContext constructors are updated, not just the one on the failing trace, so that subsequent test runs do not re-trip on a different code path.
