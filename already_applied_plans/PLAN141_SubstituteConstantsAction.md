# Fix LabeledFormula conformance in SubstituteConstantsAction / ActionClone

**Created:** 2026-03-29

## Context

`make golden` diverges at line 146460:
- **Go**: `actions.apply_mixin ENTER` (second call — first apply_mixin produced no internal traces)
- **Python**: `ast.LF.clone PRESERVE origid=212 counter=1822` (clone trace from inside first apply_mixin)

## All Conformance Issues Found

### Issue 1: ActionArgs() skips LabeledFormula
Python's `AssumeAction.args[0]` IS the LabeledFormula itself. `substitute_constants_ast` traverses into it, substitutes in its children (label, formula), and clones it. Go's `AssumeAction.ActionArgs()` returns `[a.Formula]` — the unwrapped expression only. The LF is invisible to `SubstituteConstantsAction`.

### Issue 2: ActionClone copies LF by reference
When `ActionClone` IS called, it does `LF: a.LF` — a pointer copy. Python's `ast.clone(new_args)` creates a brand new AssumeAction with a freshly-cloned LabeledFormula as `args[0]`. The LF.clone triggers `LF.clone PRESERVE`.

### Issue 3: RequiresAction/EnsuresAction/SubgoalAction drop LF entirely
Their `ActionClone` methods construct a new `AssertAction{...}` without setting the `LF` field. The LF is silently lost on clone.

### Issue 4: Go only clones when something changed
Python's `substitute_constants_ast` ALWAYS clones every non-leaf node (line 1822: `res = ast.clone(new_args)`). Go's `SubstituteConstantsAction` has `if !changed { return action }` (line 228), suppressing clones (and their traces) when no substitution matched. This is a secondary conformance gap — for the immediate divergence, the formula IS changing (formals are being substituted), so ActionClone IS called.

## Fix

### Step 1: Add helper function in `actions/action.go`

```go
// cloneLFWithFormula clones a LabeledFormula with a new inner formula,
// matching Python's substitute_constants_ast behavior which traverses
// into LabeledFormula children and clones them. The clone triggers
// the LF.clone PRESERVE trace.
func cloneLFWithFormula(lf *ast.LabeledFormula, newFormula lg.Expr) *ast.LabeledFormula {
    if lf == nil {
        return nil
    }
    return lf.Clone([]ast.Node{lf.Label, newFormula}).(*ast.LabeledFormula)
}
```

### Step 2: Fix all ActionClone methods

**AssumeAction.ActionClone** (line 169):
- Change `LF: a.LF` → `LF: cloneLFWithFormula(a.LF, args[0])`

**AssertAction.ActionClone** (line 208):
- Change `LF: a.LF` → `LF: cloneLFWithFormula(a.LF, args[0])`

**RequiresAction.ActionClone** (line 235):
- Add `LF: cloneLFWithFormula(a.LF, args[0])` (currently drops LF entirely)

**EnsuresAction.ActionClone** (line 257):
- Add `LF: cloneLFWithFormula(a.LF, args[0])` (currently drops LF entirely)

**SubgoalAction.ActionClone** (line 1455):
- Add `LF: cloneLFWithFormula(a.LF, args[0])` (currently drops LF entirely)

## Files to modify

- `/Users/jaten/go/src/github.com/glycerine/goivy/actions/action.go` — helper + 5 ActionClone methods

## Verification

Run `cd ~/goivy && make golden` and confirm the test advances past line 146460.
