# Plan: Fix Variable.String() to include sort qualifier

**Created:** 2026-03-31 evening

## Context

Trace output diverges: Go prints `rfn.abs.lemma1(M)` while Python prints `rfn.abs.lemma1(M:mem_type)`. The sort qualifier on the variable argument `M` is missing in Go.

## Root Cause

**Go** — `logic/term.go:26`:
```go
func (v *Variable) String() string { return v.Name }
```
Returns only the name, ignores `VSort`.

**Python** — `ivy_ast.py:390-393`:
```python
def __repr__(self):
    res = self.rep
    res += ':' + str(self.sort)
    return res
```
Always includes `:sort`.

In Go, property labels are fully compiled to `*lg.Apply` with `*lg.Variable` terms. In Python, labels remain as `ast.Atom` with `ast.Variable` args. Python's `ast.Variable.__repr__` includes sort; Go's `lg.Variable.String()` does not.

## Fix

**File:** `goivy/logic/term.go:26`

Change `Variable.String()` to include the sort when present:

```go
func (v *Variable) String() string {
    if v.VSort != nil && !IsTopSort(v.VSort) {
        return v.Name + ":" + v.VSort.String()
    }
    return v.Name
}
```

This matches Python's behavior: Python always appends `':' + str(self.sort)`. The `!IsTopSort` guard avoids printing the generic top sort (which Python wouldn't print either since Python's Variable.sort would be a named sort, not a generic placeholder).

**Note:** Python always appends sort unconditionally. But Python's Variable objects at this point always have a real sort set (from parsing), not a placeholder. In Go, a Variable with `VSort == nil` or `VSort == TopS` has no meaningful sort info. So the guard is equivalent to Python's unconditional append.

## Verification

Re-run the golden test. The trace line should now match:
```
compiler.CheckProperties.classify label=rfn.abs.lemma1(M:mem_type) id=661 temporal=True hasPf=True
```
