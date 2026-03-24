# Fix struct field prefix missing during module expansion (`rep:is_end` vs `rep:iter.is_end`)

**Created:** 2026-03-24 (updated)

## Context

The `make golden` test diverges at line 4606. When expanding the `sequence_iterator` module with prefix `iter`, struct field names like `is_end` and `val` should become `iter.is_end` and `iter.val`. Go doesn't prefix them:

```
Go:     (app rep:is_end terms:[] aSort:bool)
Python: (app rep:iter.is_end terms:[] aSort:bool)
```

The struct fields were recently changed from `*Atom` to `*App` nodes to better conform to Python's parse. This broke `StructSort.Defines()` which only handles `*Atom`.

## Root Cause

**Two issues, both in the ast package:**

### Issue 1: `StructSort.Defines()` doesn't handle `*App` fields

**File:** `ast/sort.go:120-130`

```go
func (s *StructSort) Defines() []string {
    for i, f := range s.Fields {
        if a, ok := f.(*Atom); ok {
            defs[i] = a.Rep        // ← only handles *Atom
        } else {
            defs[i] = fmt.Sprint(f) // ← for *App, gives "is_end:bool" not "is_end"
        }
    }
}
```

When fields are `*App` nodes, `fmt.Sprint(f)` produces `"is_end:bool"` (App.String() includes ASort). This wrong key means:
- `collectDefined()` in `inst_mod.go:361` builds a `defined` map with keys like `"is_end:bool"` instead of `"is_end"`
- `collectStatic()` in `inst_mod.go:375` has the same wrong keys
- `SubstPrefixAtomsAst` checks `ToPref["is_end"]` → false → skips prefixing

Python's `StructSort.defines()` just does `[a.rep for a in self.args]` — works for both Atom and App since both have `.rep` as a string.

### Issue 2: `AstRewrite` App case loses ASort during prefix rewriting

**File:** `ast/rewrite.go:573-579`

```go
appAtom := NewAtom(newRep, newApp.Terms...)     // ← ASort NOT copied from newApp
rewritten := rewrite.RewriteAtom(appAtom, false) // → ComposeAtoms preserves appAtom.ASort (nil)
if rewritten.Rep != newRep {
    return NewApp(NewSymbol(rewritten.Rep, nil), rewritten.Terms...) // ← ASort lost
}
```

Python's `compose_atoms` calls `copy_attributes_ast(atom, res)` which copies `.sort`. In Go, `ComposeAtoms` copies `atom.ASort`, but since `appAtom` was created without ASort, it copies nil.

Currently this doesn't manifest because the prefix isn't applied (Issue 1). After fixing Issue 1, the App would lose its ASort.

## Fix

### Fix 1: `ast/sort.go` — `StructSort.Defines()`

Add `*App` handling to extract the field name from `App.Rep`:

```go
func (s *StructSort) Defines() []string {
    defs := make([]string, len(s.Fields))
    for i, f := range s.Fields {
        if a, ok := f.(*Atom); ok {
            defs[i] = a.Rep
        } else if app, ok := f.(*App); ok {
            if sym, ok := app.Rep.(*Symbol); ok {
                defs[i] = sym.Rep
            } else {
                defs[i] = fmt.Sprint(app.Rep)
            }
        } else {
            defs[i] = fmt.Sprint(f)
        }
    }
    return defs
}
```

### Fix 2: `ast/rewrite.go` — `AstRewrite` App case

Copy ASort and Base to `appAtom` before calling `RewriteAtom`, so `ComposeAtoms` preserves them:

```go
// Convert to Atom for rewrite_atom, then convert back
appAtom := NewAtom(newRep, newApp.Terms...)
appAtom.Base = newApp.Base     // ← ADD: preserve lineno for compose_atoms
appAtom.ASort = newApp.ASort   // ← ADD: preserve sort for compose_atoms
rewritten := rewrite.RewriteAtom(appAtom, false)
if rewritten.Rep != newRep {
    result := NewApp(NewSymbol(rewritten.Rep, nil), rewritten.Terms...)
    result.Base = newApp.Base
    result.ASort = rewritten.ASort  // ← ADD: carry ASort from rewritten atom
    return result
}
return newApp
```

## Files to modify

1. **`ast/sort.go:120-130`** — `StructSort.Defines()`: add `*App` field handling
2. **`ast/rewrite.go:573-579`** — `AstRewrite` App case: preserve ASort through Atom conversion round-trip

## Verification

```bash
cd ~/goivy && make golden
```

Confirm the `rep:is_end` / `rep:iter.is_end` divergence at line 4606 is gone, and no new ASort-related divergences appear.
