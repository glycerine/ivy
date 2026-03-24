# Generalized fix: all `Defines()` methods need `*App` handling

**Created:** 2026-03-24

## Context

`make golden` diverges at line 4807 because `InterpretDecl.Defines()` only handles `*Atom` for Range args but they're `*App` nodes. This is the same class of bug as the `StructSort.Defines()` fix — when struct fields and other nodes changed from `*Atom` to `*App`, multiple `Defines()` methods weren't updated.

There's already a `nodeRep(n Node) string` helper at `ast/decl.go:360` that handles `*Atom`, `*App`, `*Symbol`, `*Variable`. We should use it everywhere.

## Audit of all Defines() methods

### Needs fix (Atom-only or incomplete type handling):

1. **`InterpretDecl.Defines()`** `ast/decl.go:1065` — `arg.(*Atom)` in Range args → use `nodeRep(arg)`
2. **`IsolateDecl.Defines()`** `ast/decl.go:1193` — `idef.Elems[0].(*Atom)` → use `nodeRep(idef.Elems[0])`
3. **`EnumeratedSort.Extension()`** `ast/sort.go:87` — `e.(*Symbol)` only → use `nodeRep(e)`
4. **`Definition.Defines()`** `ast/formula.go:309` — Symbol+Atom but not App → use `nodeRep(d.Lhs)`
5. **`ActionDef.Defines()`** `ast/decl.go:406` — `a.Name.(*Atom)` only → use `nodeRep(a.Name)`
6. **`TypeDef.Defines()`** `ast/decl.go:583` — Symbol+Atom but not App → use `nodeRep(t.Name)`
7. **`DeclBase.Defines()` LabeledFormula case** `ast/decl.go:214` — `la.(*Atom)` only → use `nodeRep(a.Label)`

### Already OK:
- `DeclBase.Defines()` main switch — already handles `*Atom` AND `*App` (lines 197-207)
- `StructSort.Defines()` — already fixed in prior commit
- `DerivedDecl.Defines()` — delegates to `Definition.Defines()` (fixed via #4)
- `ConstantSort/FunctionSort/RelationSort/IsolateObjectDecl.Defines()` — return nil

## Fixes

### Fix 1: `InterpretDecl.Defines()` — `ast/decl.go:1064-1077`

Replace `arg.(*Atom)` block with `nodeRep(arg)`:

```go
for _, arg := range rng.Args() {
    repStr := nodeRep(arg)
    if repStr == "" {
        continue
    }
    isDigit := true
    for _, c := range repStr {
        if c < '0' || c > '9' {
            isDigit = false
            break
        }
    }
    if !isDigit {
        res = append(res, repStr)
    }
}
```

### Fix 2: `IsolateDecl.Defines()` — `ast/decl.go:1193-1195`

Replace `a.(*Atom)` with `nodeRep`:

```go
if len(idef.Elems) > 0 {
    if rep := nodeRep(idef.Elems[0]); rep != "" {
        names = append(names, rep)
    }
}
```

### Fix 3: `EnumeratedSort.Extension()` — `ast/sort.go:85-93`

Replace `e.(*Symbol)` with `nodeRep`:

```go
for i, e := range s.Elems {
    ext[i] = nodeRep(e)
}
```

Note: `nodeRep` is in `ast/decl.go`. `EnumeratedSort` is in `ast/sort.go`. Both in `ast` package — accessible.

### Fix 4: `Definition.Defines()` — `ast/formula.go:308-316`

Replace Symbol+Atom check with `nodeRep`:

```go
func (d *Definition) Defines() string {
    return nodeRep(d.Lhs)
}
```

Note: `nodeRep` is in `ast/decl.go`, `Definition` in `ast/formula.go`. Both in `ast` package — accessible.

### Fix 5: `ActionDef.Defines()` — `ast/decl.go:405-410`

Replace Atom-only check:

```go
func (a *ActionDef) Defines() string {
    return nodeRep(a.Name)
}
```

### Fix 6: `TypeDef.Defines()` — `ast/decl.go:581-593`

Replace Symbol+Atom check:

```go
func (t *TypeDef) Defines() []string {
    var syms []string
    if rep := nodeRep(t.Name); rep != "" {
        syms = append(syms, rep)
    }
    if d, ok := t.Value.(DefinerSlice); ok {
        syms = append(syms, d.Defines()...)
    }
    return syms
}
```

### Fix 7: `DeclBase.Defines()` LabeledFormula case — `ast/decl.go:212-217`

Replace Atom-only label extraction:

```go
case *LabeledFormula:
    if a.Label != nil {
        if rep := nodeRep(a.Label); rep != "" {
            names = append(names, rep)
        }
    }
```

### Tests: `ast/ast_test.go`

Add test for `InterpretDecl.Defines()` with App Range args (the triggering bug). Existing tests already cover other methods — they'll now exercise the `nodeRep` path.

## Files to modify

1. **`ast/decl.go`** — fixes 1, 2, 5, 6, 7
2. **`ast/sort.go`** — fix 3
3. **`ast/formula.go`** — fix 4
4. **`ast/ast_test.go`** — add `TestInterpretDecl_Defines`

## Verification

```bash
cd ~/goivy/ast && go test -v -run "Defines"
cd ~/goivy && make golden
```

Confirm all Defines tests pass and divergence moves past line 4807.
