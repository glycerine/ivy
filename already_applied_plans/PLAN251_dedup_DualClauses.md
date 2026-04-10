# PLAN: Fix missing ToOpenFormula at log.red:233843 — port `dual_clauses` instantiator branch

Created: 2026-04-10 04:40 -03

## Context

Following the previous fix (now committed) that resolved the divergence at log line 233839 by making `interp.ConcretePost` delegate to a Clauses-level `actions.ComposeStateAction`, the next divergence in `~/ivy/goivy/log.red` (the freshly-regenerated file, now 3664 lines) appears at line 233843:

```
233839 ✓ ops.ToOpenFormula nFmlas=1 nDefs=0
233840 ✓ logicutil.CloseEPR fmla=(And terms:[(Or terms:[])])
233841 ✓ logicutil.CloseEPR fmla=(Or terms:[])
233842 ✓ ivy_logic_utils.py:1013/clauses_to_formula(): formula = (And terms:[(Or terms:[])])
233843 go : check.checkFcsNormalPath history path postFmlas=0 postDefs=0 axiomFmlas=20 axiomDefs=12 checkers=1
       py : ops.ToOpenFormula nFmlas=0 nDefs=0
```

Python emits an extra `ops.ToOpenFormula nFmlas=0 nDefs=0` (an EMPTY Clauses being converted to a formula) that Go skips, then jumps directly into the `checkFcsNormalPath` body.

## Root cause

Python's `dual_clauses` (`ivy_logic_utils.py:1554-1565`) has an **instantiator branch** that Go's `check.DualClauses` (`check/check.go:213-233`) omits:

```python
def dual_clauses(clauses, skolemizer=None):
    if skolemizer == None:
        skolemizer = lambda v: var_to_skolem('__',Variable(v.rep,v.sort))
    vs = used_variables_in_order_clauses(clauses)
    sksubs = dict((v.rep,skolemizer(v)) for v in vs)
    clauses = substitute_clauses(clauses,sksubs)
    fmla = negate(clauses_to_formula(clauses))           # call #1: emits the 233839 trace
    if instantiator != None:                              # ← module-level global, set by theory_context()
        gts = apps_clauses(clauses)                       # apps_clauses(Clauses([Or()])) == []
        insts = instantiator(gts)                         # ModuleTheoryContext.__call__([]) → empty Clauses
        fmla = And(fmla,clauses_to_formula(insts))        # call #2: emits 233843 trace on empty Clauses
    return formula_to_clauses(fmla)
```

The `instantiator` module-level global is set by `with im.module.theory_context():` at `ivy_check.py:535`. Inside that block, when `Checker(lg.Or())` is constructed, `dual_clauses` is invoked. Since `instantiator != None`, the `if instantiator != None:` branch fires, calling `clauses_to_formula(insts)` on the empty `insts`. This `clauses_to_formula` call invokes `cs.to_formula()` → `cs.to_open_formula()`, which emits the `ops.ToOpenFormula nFmlas=0 nDefs=0` trace.

### What's wrong in Go

Go has TWO `DualClauses` functions in different packages:

**`module/ops.go:470-506` `DualClauses` — CORRECT**: takes `(clauses, skolemizer, instantiator)` and properly invokes the instantiator branch (`f = &lg.And{Terms: []lg.Expr{f, clausesToFormula(insts)}}`). This faithfully ports Python.

**`check/check.go:213-233` `DualClauses` — BUGGY**:
```go
func DualClauses(c *module.Clauses) *module.Clauses {
    if c == nil {
        return c
    }
    vs := module.UsedVariablesOrdered(c)
    if len(vs) > 0 {
        subs := make(map[string]lg.Expr, len(vs))
        for _, v := range vs {
            subs[v.Name] = module.VarToSkolem("@", v)
        }
        c = module.SubstituteClausesByName(c, subs)
    }
    fmla := module.ClausesToFormula(c)
    negated := module.Negate(fmla)
    return module.FormulaToClauses(negated, nil)
    // ← MISSING: instantiator branch
}
```

The check-package version is a stripped-down reimplementation. It lacks the instantiator parameter entirely, so when `NewBaseChecker` (`check/check.go:96-107`) calls it via `fc = DualClauses(fc)` at line 99, Go produces a different (incorrect) result and never emits the second `clauses_to_formula(insts)` trace chain that Python does.

The Go `Module` already has the instantiator properly installed: `check/isolate_check.go:107` calls `cleanupTheory := mod.TheoryContext()` early in `CheckIsolate`. `module/theory.go:180-221 TheoryContext` sets `m.Instantiator = func(groundTerms []lg.Expr) *Clauses { return instantiateNonEPREntries(nonEPR, groundTerms) }`. So all that's needed is to wire `mod.Instantiator` through `NewBaseChecker` → `module.DualClauses`.

### Confirmation that Python's missing trace is from this `clauses_to_formula(insts)` call

Trace order in Python's first `clauses_to_formula(clauses)` call (the input is `Clauses([Or()])`, 1 fmla):
1. `cs.to_formula()` → `cs.to_open_formula()` emits `ops.ToOpenFormula nFmlas=1 nDefs=0` ← log line 233839
2. Returns `And(Or())`
3. `close_epr(And(Or()))` emits `CloseEPR fmla=(And terms:[(Or terms:[])])` ← log line 233840
4. Recurses on `Or()`, emits `CloseEPR fmla=(Or terms:[])` ← log line 233841
5. Back in `clauses_to_formula` body, emits `clauses_to_formula(): formula = (And terms:[(Or terms:[])])` ← log line 233842
6. Returns `drop_universals(And(Or()))`

Then Python enters the `if instantiator != None:` branch and calls the second `clauses_to_formula(insts)` where `insts` is an empty `Clauses([])`:
7. `cs.to_formula()` → `cs.to_open_formula()` emits `ops.ToOpenFormula nFmlas=0 nDefs=0` ← log line 233843 (Python only)
8. Subsequent CloseEPR / clauses_to_formula traces would follow but Go has already diverged.

Go does step 1-6 (matching Python through 233842) but then skips the entire instantiator branch, emitting the next trace (`check.checkFcsNormalPath`) where Python emits step 7.

## Recommended approach

Delete the duplicate `check/check.go:DualClauses` and route `NewBaseChecker` through the existing, correct `module.DualClauses`, threading `mod.Instantiator` from the module. This is the most faithful port: Python has exactly one `dual_clauses` function (in `ivy_logic_utils.py`), and the Go port should mirror that — `module/ops.go` is the correct home (it ports `ivy_logic_utils.py`).

### Step 1 — Delete `check/check.go:DualClauses` and the three direct tests

**File: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/check.go:207-233`**

Delete the function entirely. There is exactly one production caller (`NewBaseChecker` at line 99) and three test functions in `check_test.go`:
- `TestDualClausesNil` (line 172)
- `TestDualClausesEmpty` (line 179)
- `TestDualClausesSingleFormula` (line 187)

Per CLAUDE.md rule 4 (one Python file → one Go file with same base name): `dual_clauses` is in `ivy_logic_utils.py`, so its Go home is `module/ops.go` (which already has the correct port). The check-package wrapper is a duplication that should not exist.

The three `TestDualClauses*` tests should be **removed** from `check_test.go`. They are vestigial tests of the duplicate function. Equivalent coverage already exists in `module/` for the canonical `module.DualClauses`. (Verify: `grep -rn "TestDualClauses" /Users/jaten/go/src/github.com/glycerine/ivy/goivy/module/` — if no equivalent exists, port the three tests over to the `module` package; do not leave them in `check`.)

### Step 2 — Change `NewBaseChecker` to take `*module.Module` and call `module.DualClauses`

**File: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/check.go:86-107`**

Change the `BaseChecker` struct field and constructor signature:

```go
type BaseChecker struct {
    Mod        *module.Module     // was: Cfg *module.Config
    FC         *module.Clauses
    ReportPass bool
    Inverted   bool
    FailedFlag bool
}

// NewBaseChecker creates a BaseChecker for the given conjecture formula.
// If invert is true (the default), the formula is dualized for checking.
// Faithful port of Python ivy_check.py Checker.__init__ (lines 218-226).
func NewBaseChecker(mod *module.Module, conj lg.Expr, reportPass bool, invert bool) *BaseChecker {
    fc := module.FormulaToClauses(conj, nil)
    if invert {
        // Python: def witness(v): return lg.Symbol('@'+v.name, v.sort)
        //         self.fc = lut.dual_clauses(self.fc, witness)
        witness := func(v *lg.Variable) lg.Expr {
            return module.VarToSkolem("@", v)
        }
        fc = module.DualClauses(fc, witness, mod.Instantiator)
    }
    return &BaseChecker{
        Mod:        mod,
        FC:         fc,
        ReportPass: reportPass,
        Inverted:   invert,
    }
}
```

Key points:
- `mod.Instantiator` is already set by `mod.TheoryContext()` in `isolate_check.go:107` before any checker is created. For our failing path it's non-nil, so `module.DualClauses` will execute the instantiator branch and emit the missing trace chain.
- For unit-test paths where `mod.Instantiator` is nil (no theory_context wrapping), the branch is skipped — exactly matching Python's `if instantiator != None:`.
- The `witness` function uses the `@` prefix matching Python's `lg.Symbol('@'+v.name, v.sort)`, via the existing `module.VarToSkolem("@", v)` helper.
- `module.DualClauses` already handles the nil-instantiator case correctly (`module/ops.go:499-503`).

Then update all uses of `c.Cfg.X` inside the BaseChecker methods to `c.Mod.Cfg.X`:

**File: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/check.go:115-147`** — `Sat`, `Unsat`, `Fail`, `Pass` access `c.Cfg.OnlyCheckUnprovable`, `c.Cfg.Failures`, `c.Cfg.Diagnose`, `c.Cfg.OptTrace`, etc. Replace each `c.Cfg.X` with `c.Mod.Cfg.X`.

### Step 3 — Update `NewConjChecker` and `NewConjAssumer` similarly

**File: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/check.go:158-205`**

```go
func NewConjChecker(mod *module.Module, lf *ast.LabeledFormula, indent int) *ConjChecker {
    base := NewBaseChecker(mod, lf.Formula.(lg.Expr), true, true)
    return &ConjChecker{
        BaseChecker: *base,
        LF:          lf,
        Indent:      indent,
    }
}

func NewConjAssumer(mod *module.Module, lf *ast.LabeledFormula) *ConjAssumer {
    base := NewBaseChecker(mod, lf.Formula.(lg.Expr), false, false)
    return &ConjAssumer{
        BaseChecker: *base,
        LF:          lf,
    }
}
```

### Step 4 — Update production callers in `check.go`

**File: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/check.go:739, 751`**

```go
// line 739:
checkers = append(checkers, NewConjChecker(mod, c, indent))   // was: NewConjChecker(mod.Cfg, c, indent)

// line 751:
checker := NewBaseChecker(mod, &lg.Or{}, reportPass, true)    // was: NewBaseChecker(mod.Cfg, ...)
```

Both call sites already have access to `mod *module.Module`, so the change is just `mod.Cfg` → `mod`.

Search for any other production callers via:
```
grep -rn "NewBaseChecker\|NewConjChecker\|NewConjAssumer" /Users/jaten/go/src/github.com/glycerine/ivy/goivy --include="*.go" | grep -v _test.go
```
Update each to pass `mod` instead of `mod.Cfg` or `cfg`.

### Step 5 — Update test callers in `check_test.go` and `check_port_test.go`

**File: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/check_test.go`**

Replace each occurrence of `module.NewConfig()` (when used to construct a checker) with `module.New()`. The `NewConjChecker`/`NewConjAssumer`/`NewBaseChecker` calls then take the resulting `*module.Module`. Pattern:

```go
// Before:
c := NewBaseChecker(module.NewConfig(), lg.True, true, true)
// After:
c := NewBaseChecker(module.New(), lg.True, true, true)

// Before (multi-line, still using cfg later):
cfg := module.NewConfig()
c := NewBaseChecker(cfg, lg.True, false, true)
// ... uses cfg.Failures
// After:
mod := module.New()
c := NewBaseChecker(mod, lg.True, false, true)
// ... uses mod.Cfg.Failures
```

Affected tests in `check_test.go`:
- `TestBaseCheckerCreate` (line 16-27)
- `TestBaseCheckerCond` (line 29-35)
- `TestBaseCheckerPass` (line 37-46)
- `TestBaseCheckerFail` (line 48-60) — uses cfg.Failures, becomes mod.Cfg.Failures
- `TestBaseCheckerSatCallsFail` (line 62-71) — same pattern
- `TestBaseCheckerUnsatCallsPass` (line 73-79)
- `TestBaseCheckerAssume` (line 81-86)
- `TestBaseCheckerGetAnnot` (line 88-93)
- `TestBaseCheckerGetLF` (line 95-100)
- `TestConjCheckerCreate` (line 104-119) — uses cfg.AstCfg.NewLabeledFormula, becomes mod.Cfg.AstCfg.NewLabeledFormula
- `TestConjCheckerGetLF` (line 121-129) — same
- `TestConjCheckerImplementsChecker` (line 131-136) — same
- `TestConjAssumerCreate` (line 140-151) — same
- `TestConjAssumerAssume` (line 153-161) — same
- `TestConjAssumerImplementsChecker` (line 163-168) — same

**Delete the three `TestDualClauses*` tests at lines 172-204** since `check.DualClauses` no longer exists.

**File: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/check_port_test.go`** — two `NewConjChecker(module.NewConfig(), lf, 8)` / `... lf, 92)` calls at lines 83, 92. Change to `NewConjChecker(module.New(), lf, 8/92)`.

### Step 6 — Verify no other DualClauses duplicates affect the path

There are also:
- `bmc/bmc.go:234` `DualClauses(conj *module.Clauses) *module.Clauses`
- `z3bridge/solver_clauses.go:51` `DualClauses(clauses *module.Clauses) *module.Clauses`

Both are likely missing the instantiator branch too. They are **out of scope** for this fix because they are not in the `TestOrdLive` failing path at 233843. Track them for a follow-up plan to delete in favor of `module.DualClauses`. Do **not** silently regress them either — leave them untouched.

## Critical files to modify

1. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/check.go`
   - Delete `DualClauses` (lines 207-233)
   - Change `BaseChecker` struct field `Cfg *module.Config` → `Mod *module.Module`
   - Change `NewBaseChecker(cfg, ...)` → `NewBaseChecker(mod, ...)` and route through `module.DualClauses`
   - Change `NewConjChecker(cfg, ...)` and `NewConjAssumer(cfg, ...)` similarly
   - Update all `c.Cfg.X` references inside `BaseChecker` methods to `c.Mod.Cfg.X`
   - Update production call sites at lines 739, 751
2. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/check_test.go`
   - Update ~15 test calls from `module.NewConfig()` → `module.New()`
   - Delete three `TestDualClauses*` tests (lines 172-204)
3. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/check_port_test.go`
   - Update 2 test calls from `module.NewConfig()` → `module.New()`

## Existing functions to reuse (do NOT recreate)

- `module.DualClauses(clauses *Clauses, skolemizer Skolemizer, instantiator func([]lg.Expr) *Clauses) *Clauses` — at `module/ops.go:470-506` — the canonical port of Python `dual_clauses` with full instantiator support
- `module.VarToSkolem(prefix string, v *lg.Variable) lg.Expr` — produces `@<name>` constants matching Python's witness
- `module.Skolemizer` type — `func(v *lg.Variable) lg.Expr` (`module/ops.go:462`)
- `module.New() *Module` — at `module/module.go:202` — produces a fresh Module with a Cfg, suitable for tests
- `mod.Instantiator` — already populated by `mod.TheoryContext()` (called early in `check/isolate_check.go:107`)
- `module.FormulaToClauses` — already used by current `NewBaseChecker`
- `clausesToFormula` (lowercase) at `module/ops.go:513-517` — emits the `ivy_logic_utils.py:1013/clauses_to_formula()` trace

## Verification

1. **Build**: `cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go build ./...` — expect clean build after all signature changes are applied.
2. **Targeted unit tests**:
   ```
   go test ./check/... -count=1
   go test ./module/... -count=1 -run TestDualClauses
   ```
   Expect all check_test.go and check_port_test.go tests to pass after the signature updates.
3. **Failing golden test** — primary verification:
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test ./parser/... -run TestOrdLive -timeout 600s
   ```
   Expect the divergence at line 233843 to be gone. The Go side should now emit `ops.ToOpenFormula nFmlas=0 nDefs=0` matching Python at log position 233843, followed by additional matching `CloseEPR` and `clauses_to_formula` traces, and eventually both sides reaching `check.checkFcsNormalPath`.
4. **Inspect the new log**: after rerunning, look at `/Users/jaten/ivy/goivy/log.red` around line 233843. Capture the next divergence (if any) for a follow-up plan.
5. **Cross-check no regressions**:
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test ./... -count=1 -timeout 600s
   ```

## Notes / out of scope

- `bmc/bmc.go:234 DualClauses` and `z3bridge/solver_clauses.go:51 DualClauses` are duplicate stripped implementations with the same instantiator-omission bug. They are **not in the failing path** at 233843, so leave them alone here. Track for a follow-up cleanup plan to delete in favor of `module.DualClauses`.
- The same formula-vs-Clauses anti-pattern noted in the previous plan still exists in `actions/interpolant.go:43`, `tactics/tactics.go:159-170`, and `interp/helpers.go:212-241`. Out of scope.
- Per CLAUDE.md rule C: do not introduce any new package-level vars; the `Mod *module.Module` field on `BaseChecker` is per-instance state, not a global.
- Per CLAUDE.md rule 4 (one Python file → one Go file): `dual_clauses` lives in `ivy_logic_utils.py`, so the Go port lives in `module/ops.go`. Removing the duplicate from `check/check.go` aligns with this rule.
- Per CLAUDE.md rule 9: do not leave the duplicate `check.DualClauses` as a "good enough" wrapper. Delete it cleanly and route through the canonical `module.DualClauses`.
- The `BaseChecker.Mod *module.Module` field replaces the existing `Cfg *module.Config` field. Do NOT keep both — that would be redundant. Update method bodies to use `c.Mod.Cfg.X` instead of `c.Cfg.X` in the same edit.
