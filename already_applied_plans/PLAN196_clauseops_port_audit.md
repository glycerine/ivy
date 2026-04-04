# Clauseops Port Audit: Complete Go conformance to Python Clauses ecosystem

**Created:** 2026-04-04 ~12:00 UTC

## Context

The Go clauseops package ports Python's `Clauses` class and related functions from
`ivy_logic_utils.py`. A systematic function-by-function audit reveals several bugs
where Go diverges from Python's behavior. These must all be fixed before chasing
specific test divergences — getting the foundation right is prerequisite to correctness.

The prior simpIte fix and transrel trace additions (steps 1-5 from original plan) are
already applied and remain valid. This plan addresses the remaining port bugs found
by audit.

## Bugs Found

### Bug 1: `orClausesInt` uses `SimpIte` — Python uses bare `Ite`

**Python** `or_clauses_int` line 1353:
```python
defidx[s] = Definition(d.args[0], Ite(v, d.args[1], defidx[s].args[1]))
```
→ bare `Ite`, NO simplification

**Go** `orClausesInt` line 255:
```go
il.SimpIte(vs[i], d.Rhs, existing.Rhs)
```
→ uses `SimpIte` — **wrong**, should be bare `Ite`

Only `ite_clauses_int` uses `simp_ite` in Python. The previous fix over-applied
`SimpIte` to `orClausesInt` as well. The original local `simpIte` was actually
closer to correct for `orClausesInt` (only handled equal branches), but the right
fix is bare `Ite` to match Python exactly.

**File:** `clauseops/ops.go` ~line 255
**Fix:** Change `il.SimpIte(vs[i], d.Rhs, existing.Rhs)` to
`&lg.Ite{ISort: d.Rhs.NodeSort(), Cond: vs[i], Then: d.Rhs, Else: existing.Rhs}`

### Bug 2: `elimDeadDefinitions` missing skolem rename step

**Python** `elim_dead_definitions` (lines 1326-1337):
```python
captured = [sym for sym in defd if any(sym not in a.defidx for a in args)]
dead = [sym for sym in captured if not sym.is_skolem()]
to_rename = [sym for sym in captured if sym.is_skolem()]
args = [rename_symbols(rn, arg, to_rename) for arg in args]
res = [elim_definitions(a, dead) for a in args]
```

Python splits captured symbols into:
- `dead` (non-skolem) → eliminated by converting to constraints
- `to_rename` (skolem) → renamed to fresh names via `rename_symbols(rn, arg, to_rename)`

**Go** `elimDeadDefinitions` (lines 637-683):
Treats ALL captured symbols as dead — **no skolem/non-skolem split, no rename step**.

**File:** `clauseops/ops.go` ~line 637
**Fix:** Split captured into dead (non-skolem) and toRename (skolem). Rename skolems
via `RenameClauses` before eliminating dead defs.

### Bug 3: `OrClausesTyped` missing `fix_or_annot` re-wrap

**Python** `or_clauses` always returns via `fix_or_annot(res, fixed_vs, fixed_args)`
which creates a **new** `Clauses(res.fmlas, res.defs, annot)`. This re-wraps through
the constructor, re-applying `coerce_clause_to_formula` + `collect_and_list`.

Python also preserves orig_args ordering for annotation reconstruction, assigning
`Or()` (false) as the `v` for false branches.

**Go** `OrClausesTyped`:
- Returns `nonFalse[0]` directly when only 1 non-false arg — no re-wrap.
- Returns `orClausesInt` result directly — no annotation fixup.
- No `fix_or_annot` equivalent at all.

**File:** `clauseops/ops.go` ~line 190
**Fix:** Add `fixOrAnnot` function matching Python's `fix_or_annot`. Track `origArgs`
in `OrClausesTyped` and call `fixOrAnnot` at the end.

### Bug 4: `AndClauses` (untyped) annotation handling incomplete

**Go** `AndClauses` (line 57-65):
```go
// In Python, annot = annot.conj(c.annot). We just take first non-nil.
```

The comment explicitly acknowledges it only takes first non-nil rather than doing
the proper `annot.conj(c.annot)` chain. `AndClausesTyped` has the correct logic
via `AnnotConjoiner`, but the untyped `AndClauses` does not.

**File:** `clauseops/ops.go` ~line 55
**Fix:** Use the same `AnnotConjoiner` logic as `andClausesImpl`.

### Bug 5: `RenameClauses` doesn't rename annotations

**Python** `rename_clauses` is defined as:
```python
rename_clauses = apply_func_to_clauses(rename_ast, annot_fun=rename_clauses_annot_fun)
```
where `rename_clauses_annot_fun` calls `annot.rename(map)`.

**Go** `RenameClauses` calls `clauses.Apply(fn)` which passes `c.Annot` through
unchanged.

**File:** `clauseops/ops.go` ~line 501
**Fix:** After calling `Apply`, rename the annotation if it implements a `Rename`
interface.

## Plan

### Step 1: Fix `orClausesInt` def merging to use bare Ite

In `clauseops/ops.go` `orClausesInt`, change line ~255 from:
```go
il.SimpIte(vs[i], d.Rhs, existing.Rhs)
```
to:
```go
newIte, _ := lg.NewIte(vs[i], d.Rhs, existing.Rhs)
```
(or construct the Ite struct directly). This matches Python which uses bare `Ite`
in `or_clauses_int` but `simp_ite` in `ite_clauses_int`.

### Step 2: Fix `elimDeadDefinitions` to split skolem/non-skolem

Rewrite `elimDeadDefinitions` to match Python:

```go
func elimDeadDefinitions(rn *iu.UniqueRenamer, args []*Clauses) []*Clauses {
    // 1. Collect all defined symbols
    defined := make(map[lg.NodeKey]*lg.Const)
    for _, a := range args {
        for _, d := range a.Defs {
            sym := d.Defines()
            defined[lg.Key(sym)] = sym
        }
    }

    // 2. Find captured: defined somewhere but not in all args
    var captured []*lg.Const
    for key, sym := range defined {
        for _, a := range args {
            if _, ok := a.DefIdx[key]; !ok {
                captured = append(captured, sym)
                break
            }
        }
    }

    // 3. Split: non-skolem → dead (eliminate), skolem → rename
    var dead []lg.NodeKey
    toRename := make(map[lg.NodeKey]*lg.Const)
    for _, sym := range captured {
        key := lg.Key(sym)
        if isSkolem(sym) {
            toRename[key] = sym
        } else {
            dead = append(dead, key)
        }
    }

    // 4. Rename skolems to fresh names
    if len(toRename) > 0 {
        subs := make(map[lg.NodeKey]*lg.Const)
        for _, sym := range toRename {
            newName := rn.Rename(sym.Name)
            subs[lg.Key(sym)] = lg.NewConst(newName, sym.CSort)
        }
        for i, a := range args {
            args[i] = RenameClauses(a, subs)
        }
    }

    // 5. Eliminate dead definitions
    if len(dead) == 0 {
        return args
    }
    deadSet := make(map[lg.NodeKey]bool, len(dead))
    for _, s := range dead {
        deadSet[s] = true
    }
    result := make([]*Clauses, len(args))
    for i, a := range args {
        result[i] = elimDefinitions(a, deadSet)
    }
    return result
}
```

Also extract `elimDefinitions` as a separate function matching Python's
`elim_definitions`.

### Step 3: Add `fixOrAnnot` and integrate into `OrClausesTyped`

Add `fixOrAnnot` matching Python's `fix_or_annot`:

```go
func fixOrAnnot(res *Clauses, vs []lg.Expr, args []*Clauses) *Clauses {
    if len(args) == 0 {
        return res
    }
    // Build annotation from args using IteAnnotation
    annot := args[0].Annot
    for i := 1; i < len(args); i++ {
        a := args[i].Annot
        if annot == nil || a == nil {
            annot = nil
        } else if iter, ok := a.(AnnotIter); ok {
            annot = iter.Ite(vs[i], annot)
        }
    }
    return NewClauses(res.Fmlas, res.Defs, annot)
}
```

Modify `OrClausesTyped` to track `origArgs`, build `fixedVs`/`fixedArgs`, and
return via `fixOrAnnot`.

### Step 4: Fix `AndClauses` annotation handling

In the untyped `AndClauses`, replace the "take first non-nil" logic with the same
`AnnotConjoiner` pattern used in `andClausesImpl`.

### Step 5: Fix `RenameClauses` to rename annotations

Add an `AnnotRenamer` interface:
```go
type AnnotRenamer interface {
    Rename(subs map[lg.NodeKey]*lg.Const) interface{}
}
```

In `RenameClauses`, after `Apply`, check if the annotation implements `AnnotRenamer`
and rename it.

### Step 6: Add comprehensive unit tests

**File:** `clauseops/clauseops_test.go` (or new file `clauseops/port_audit_test.go`)

Tests:
1. `TestOrClausesIntBareIte`: Verify `orClausesInt` produces bare `Ite` in defs,
   not simplified Or/And. Create two clause sets with overlapping defs, verify
   the merged def uses `Ite` (not `Or`).

2. `TestElimDeadDefinitionsSkolemRename`: Create args where a skolem def exists in
   one arg but not another. Verify it gets renamed (not eliminated as a constraint).

3. `TestElimDeadDefinitionsNonSkolemElim`: Create args where a non-skolem def exists
   in one arg but not another. Verify it gets converted to a constraint formula.

4. `TestOrClausesOneFalse`: Create `OrClausesTyped(false_cls, non_false_cls)`.
   Verify result matches `non_false_cls` (no Tseitin encoding).

5. `TestOrClausesBothNonFalse`: Create two non-false clause sets. Verify Tseitin
   encoding produces `Or(v1, v2)` plus guard clauses.

6. `TestIteClauses`: Verify `IteClauses` produces `simp_ite`-simplified defs
   and adds `v = cond` definition.

7. `TestNewClausesDropUniversals`: Verify `NewClauses` strips ForAll from fmlas.
   `NewClauses([ForAll(v, Or())], nil, nil)` should have `IsFalse() == true`.

8. `TestNewClausesCollectAndList`: Verify nested And gets flattened.
   `NewClauses([And(a, And(b, c))], nil, nil)` should have `Fmlas = [a, b, c]`.

## Critical files to modify

1. `clauseops/ops.go` — Bugs 1-4 fixes
2. `clauseops/ops.go` — Bug 5 fix (RenameClauses annotation)
3. `clauseops/port_audit_test.go` — comprehensive unit tests

## Existing code to reuse

- `lg.NewIte` at `logic/formula.go` — bare Ite constructor (for Bug 1)
- `RenameClauses` at `clauseops/ops.go:501` — used by elimDeadDefinitions skolem rename
- `isSkolem` at `clauseops/clauses.go:313` — skolem detection
- `defToConstraint` at `clauseops/clauses.go:247` — def to formula conversion

## Verification

1. `go build ./...` — ensure compilation
2. `go test ./clauseops/...` — run new port audit tests + existing tests
3. `go test ./ivylogic/...` — ensure SimpIte tests still pass
4. Re-run `TestOrdLive` in `lalr_full/` to check divergence improvement
