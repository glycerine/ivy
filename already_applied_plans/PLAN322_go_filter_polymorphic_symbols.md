# PLAN: Filter polymorphic symbols from DeclBase.Defines()

Created: 2026-04-17, 17:15

## Context

`goivy_check tilelink1.ivy` fails with `LALR parse error: syntax error`. The golden test (`make tilelink`) shows the first divergence at xtrace position 29:

- Go: emits `parser.define ENTER` for the `<` relation declaration
- Python: emits `parser.declare HASH` directly (no `define` call)

Python's `RelationDecl.defines()` and `ConstantDecl.defines()` filter out names in `iu.polymorphic_symbols` (like `<`, `<=`, `+`, `*`, etc.). Go's `DeclBase.Defines()` does not filter, so it returns `<` as a defined name, causing `declare()` to call `define("<", ...)` when it shouldn't.

## Root cause

`ast/decl_ast.go` lines 258-261 — the `Atom` case in `DeclBase.Defines()`:
```go
case *Atom:
    if a.Rep != "" {
        names = append(names, a.Rep)  // no polymorphic filtering
    }
```

Python equivalent (`ivy_ast.py:996-997`):
```python
def defines(self):
    return [(c.rep,lineno(c)) for c in self.args if c.rep not in iu.polymorphic_symbols]
```

## Fix

Filter polymorphic symbols in the `Atom` and `App` cases of `DeclBase.Defines()`. The `ivyutils.PolymorphicSymbols` map already exists and `ast` already imports `ivyutils` as `iu`.

In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/ast/decl_ast.go`, lines 258-267:

```go
case *Atom:
    if a.Rep != "" {
        if _, poly := iu.PolymorphicSymbols[a.Rep]; !poly {
            names = append(names, a.Rep)
        }
    }
case *App:
    if a.Rep != nil {
        if s, ok := a.Rep.(*Symbol); ok && s.Rep != "" {
            if _, poly := iu.PolymorphicSymbols[s.Rep]; !poly {
                names = append(names, s.Rep)
            }
        }
    }
```

## Critical file

- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/ast/decl_ast.go` — filter polymorphic symbols in `DeclBase.Defines()`

## Verification

```bash
cd ~/ivy/goivy && make tilelink
```

The xtrace divergence at position 29 should be resolved. Go should skip `define` for `<` just like Python does.
