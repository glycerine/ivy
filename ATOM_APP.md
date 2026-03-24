# Fix `rhs:(app rep:- ...)` vs `rhs:(atom rep:"-" ...)` mismatch

**Created:** 2026-03-24

## Context

After fixing the `vSort:this` issue (done), the `make golden` diff at line 430 still shows:
- Go: `rhs:(app rep:- terms:[...])`
- Python: `rhs:(atom rep:"-" terms:[...])`

Two symptoms, one root cause:
1. **Wrong type**: `app` instead of `atom`
2. **Unquoted rep**: `rep:-` instead of `rep:"-"` (Atom.Canon uses `%q`, App.Canon uses `%v`)

## Root Cause

Python's `fmla : term` rule calls `app_to_atom(p[1])`:
```python
# /Users/jaten/pyivy/ivy/ivy/ivy_logic_parser.py:470
p[0] = app_to_atom(p[1])
```

Go's equivalent rule does NOT:
```go
// /Users/jaten/go/src/github.com/glycerine/goivy/lalr_full/grammar_v17.y:1976
$$ = $1   // missing AppToAtom call
```

## Fix

**File:** `lalr_full/grammar_v17.y` line 1976

Change:
```go
$$ = $1
```
To:
```go
$$ = ast.AppToAtom($1)
```

Then regenerate the parser: the generated `grammar_v17.go` will also need updating (same change at the corresponding location).

## Verification

Run `cd ~/goivy && make golden` and confirm the `app`/`atom` and `rep:-`/`rep:"-"` diff lines are gone.
