# PLAN198: module/ Second Re-Audit — Bugs in skolem.go, subsume.go

**Created:** 2026-04-04 ~23:00 UTC

## Context

A second comprehensive function-by-function re-audit of Go `module/` against Python
`ivy_logic_utils.py` found 4 bugs where Go diverges from Python, plus 1 code duplication
issue. The previous audits (PLAN196, PLAN197) focused on ops.go and clauses.go. This
audit covers skolem.go and subsume.go, which were not previously audited at this depth.

## Bugs

### Bug A: `DualFormula` missing instantiator (MEDIUM)

**File:** `module/skolem.go:46-62`
**Python** (`ivy_logic_utils.py:1536-1546`):
```python
def dual_formula(fmla, skolemizer=None):
    ...
    fmla = negate(substitute_ast(fmla,sksubs))
    if instantiator != None:
        gts = apps_ast(fmla)
        insts = clauses_to_formula(instantiator(gts))
        fmla = And(fmla,insts)
    return fmla
```
**Go:** `DualFormula(fmla, skolemizer)` — no instantiator parameter, no instantiator check.
**Impact:** Definition instances not conjoined during `DualFormula` for non-EPR theories.
**Fix:** Add `instantiator func([]lg.Expr) *Clauses` parameter. After negation, check and apply.
**Callers to update:**
- `actions/update.go:546` — `mod.DualFormula(fmla, nil)` → `mod.DualFormula(fmla, nil, nil)` (instantiator
  not needed here since `DualClauses` on the clauses path already handles it; Python's `dual_formula`
  caller in `ivy_actions.py:360` uses the global, but the global is already threaded differently in Go)

### Bug B: `SkolemizeFormula` missing instantiator (MEDIUM)

**File:** `module/skolem.go:66-89`
**Python** (`ivy_logic_utils.py:1548-1561`):
```python
def skolemize_formula(fmla, skolemizer=None):
    ...
    fmla = substitute_ast(fmla,sksubs)
    if instantiator != None:
        gts = apps_ast(fmla)
        insts = clauses_to_formula(instantiator(gts))
        fmla = And(fmla,insts)
    return fmla
```
**Go:** `SkolemizeFormula(fmla, skolemizer)` — no instantiator parameter.
**Impact:** Definition instances not conjoined during skolemization for non-EPR theories.
**Fix:** Same pattern as Bug A.
**Callers to update:**
- `actions/update.go:501` — `mod.SkolemizeFormula(fmla, nil)` → extract instantiator from ctx

### Bug C: `TaggedOrClauses` wrong implementation (MEDIUM)

**File:** `module/subsume.go:356-362`
**Python** (`ivy_logic_utils.py:1400-1407`):
```python
def tagged_or_clauses(prefix, *args):
    args = coerce_args_to_clauses(args)
    res,vs,args = or_clauses_int(UniqueRenamer('__to0',dict()),args)
    return fix_or_annot(res,vs,args)
```
**Go:** Delegates to `OrClausesTyped(args...)` which is WRONG because:
1. Uses `__ts0` prefix — Python uses `__to0`
2. Filters false branches — Python does NOT filter (passes all args to `or_clauses_int`)
3. Populates used names from symbols — Python passes empty `dict()` (no used names)
**Impact:** Incorrect disjunct indices when used with `FindTrueDisjunct` (false-branch
filtering shifts indices). Different tag variable names (`__ts0` vs `__to0`).
**Fix:** Rewrite to call `orClausesIntWithVs` directly with correct parameters.

### Bug D: `ExistsQuantClausesMap` map2-to-map1 conversion wrong (MEDIUM)

**File:** `module/subsume.go:415-472`
**Python** (`ivy_logic_utils.py:1466-1484`):
```python
# Step 3: simple mapping
for v, w in map2.items():
    for x in w:
        map1[x] = v

# Step 4: unconditionally rename ALL syms
for s in syms:
    map1[s] = rename(s, rn)
```
**Go issues:**
1. Lines 433-461: Convoluted nested loops with `_ = x` no-ops and self-mapping. Does not
   match Python's simple `map1[x] = v` logic. Missing reverse-key lookup to convert
   `lg.NodeKey` back to `*lg.Const` for map1 values.
2. Line 466: `if _, already := map1[...]; !already` — Python unconditionally overwrites.
   Go skips if already present. This means some syms may not get fresh-renamed.
**Fix:** Rewrite with correct map2→map1 conversion using a key-to-Const lookup,
and unconditional overwrite in step 4.

### Cleanup: Duplicate `dualFormula`/`skolemizeFormula` in actions/update.go

**File:** `actions/update.go:171-185, 320-339`
**Issue:** Local `dualFormula()` and `skolemizeFormula()` duplicate the logic already in
`module/skolem.go:DualFormula` and `SkolemizeFormula`. Both miss the instantiator.
**Fix:** After fixing Bugs A/B, the local functions should be removed and callers updated
to use the module-level versions. Tests `TestDualFormula` and `TestSkolemizeFormula`
should call the module-level functions. Defer this cleanup to a follow-up or include
if straightforward.

## Execution Plan

### Step 1: Bug A — Add instantiator to `DualFormula`

In `module/skolem.go`, change signature and add post-negate instantiator logic:
```go
func DualFormula(fmla lg.Expr, skolemizer func(*lg.Variable) lg.Expr,
    instantiator func([]lg.Expr) *Clauses) lg.Expr {
    // ... existing logic ...
    fmla = Negate(fmla)
    // Python: if instantiator != None: fmla = And(fmla, clauses_to_formula(insts))
    if instantiator != nil {
        gts := il.AppsAst(fmla)
        insts := instantiator(gts)
        if len(insts.Fmlas) > 0 {
            fmla = &lg.And{Terms: []lg.Expr{fmla, clausesToFormula(insts)}}
        }
    }
    return fmla
}
```
Update caller `actions/update.go:546`: `mod.DualFormula(fmla, nil, nil)`

### Step 2: Bug B — Add instantiator to `SkolemizeFormula`

Same pattern. In `module/skolem.go`:
```go
func SkolemizeFormula(fmla lg.Expr, skolemizer func(*lg.Variable) lg.Expr,
    instantiator func([]lg.Expr) *Clauses) lg.Expr {
    // ... existing logic ...
    // After substitution, before return:
    if instantiator != nil {
        gts := il.AppsAst(fmla)
        insts := instantiator(gts)
        if len(insts.Fmlas) > 0 {
            fmla = &lg.And{Terms: []lg.Expr{fmla, clausesToFormula(insts)}}
        }
    }
    return fmla
}
```
Update caller `actions/update.go:501`:
```go
var inst func([]lg.Expr) *mod.Clauses
if ctx != nil {
    inst = ctx.Instantiator
}
fmla = mod.SkolemizeFormula(fmla, nil, inst)
```

### Step 3: Bug C — Rewrite `TaggedOrClauses`

In `module/subsume.go`, replace lines 356-362:
```go
func TaggedOrClauses(prefix string, args ...*Clauses) *Clauses {
    if len(args) == 0 {
        return TrueClauses(nil)
    }
    // Python: or_clauses_int(UniqueRenamer('__to0',dict()),args)
    // Uses __to0 prefix, empty used set, no false-branch filtering
    rn := iu.NewUniqueRenamer("__to0", nil)
    res, vs, processedArgs := orClausesIntWithVs(rn, args)
    return fixOrAnnot(res, vs, processedArgs)
}
```

### Step 4: Bug D — Rewrite `ExistsQuantClausesMap`

In `module/subsume.go`, replace lines 415-472 with corrected logic:
```go
func ExistsQuantClausesMap(syms []*lg.Const, clauses *Clauses) (map[lg.NodeKey]*lg.Const, *Clauses) {
    used := collectAllUsedNames(clauses)
    symset := make(map[lg.NodeKey]bool, len(syms))
    for _, s := range syms {
        symset[lg.Key(s)] = true
    }
    map1 := make(map[lg.NodeKey]*lg.Const)
    map2 := make(map[lg.NodeKey][]lg.Expr)

    var defs []*il.Definition
    for _, df := range clauses.Defs {
        if !EqcmUpd(df.Lhs, df.Rhs, symset, map2) {
            if !EqcmUpd(df.Rhs, df.Lhs, symset, map2) {
                defs = append(defs, df)
            }
        }
    }

    // Build key→Const lookup for map2 keys (from definition args)
    keyToConst := make(map[lg.NodeKey]*lg.Const)
    for _, df := range clauses.Defs {
        if c, ok := df.Lhs.(*lg.Const); ok {
            keyToConst[lg.Key(c)] = c
        }
        if c, ok := df.Rhs.(*lg.Const); ok {
            keyToConst[lg.Key(c)] = c
        }
    }
    for _, s := range syms {
        keyToConst[lg.Key(s)] = s
    }

    // Python: for v,w in map2.items(): for x in w: map1[x] = v
    for vKey, w := range map2 {
        vConst := keyToConst[vKey]
        if vConst == nil {
            continue
        }
        for _, x := range w {
            if c, ok := x.(*lg.Const); ok {
                map1[lg.Key(c)] = vConst
            }
        }
    }

    newClauses := NewClauses(clauses.Fmlas, defs, nil)
    rn := iu.NewUniqueRenamer("__", used)
    // Python: unconditionally overwrite ALL syms
    for _, s := range syms {
        newName := rn.Rename(s.Name)
        map1[lg.Key(s)] = lg.NewConst(newName, s.CSort)
    }
    return map1, RenameClauses(newClauses, map1)
}
```

### Step 5: Add tests

In `module/port_audit_test.go`, add:
- `TestTaggedOrClausesPrefix` — verify tag variables use `__to0` prefix, no false filtering
- `TestExistsQuantClausesMapSimple` — verify basic equivalence class merging and renaming

### Step 6: Verify

```
go build ./...
go test ./module/...
go test ./actions/...
go test ./...
```

## Critical Files to Modify

- `module/skolem.go` — Bugs A, B
- `module/subsume.go` — Bugs C, D
- `actions/update.go` — caller updates for Bugs A, B

## Verification

1. `go build ./...` — compiles clean
2. `go test ./module/...` — existing + new tests pass
3. `go test ./actions/...` — existing tests pass with updated signatures
4. `go test ./...` — full suite passes
