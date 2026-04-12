# Plan: Decl Clone() must not preserve Attributes or Common

Created: 2026-04-12 ~afternoon

## Context

**Problem**: Go/Python xtrace divergence at `parser.spaa EXIT` (log.red line 99-100):
- Go: `(constantDecl declArgs:[...] attributes:[spec] common:nil)`
- Python: `(constantDecl declArgs:[...] attributes:[] common:nil)`

Both enter `spaa` with `attributes:[spec]`, but Python's `AST.clone()` resets attributes to `()` and common to `None` (via `Decl.__init__`), while Go's `Clone()` explicitly preserves both fields.

**Root cause**: Python's `AST.clone()` (`ivy_ast.py:31-39`) calls `type(self)(*args)` which invokes `Decl.__init__` — this always resets `self.attributes = ()` and `self.common = None`. Go's 25+ Decl `Clone()` methods all explicitly copy `Attributes: d.Attributes, Common: d.Common`.

**Why it's safe**: Callers that need attributes after clone restore them explicitly. Python's `inst_mod` does `idecl.attributes = decl.attributes` at line 214. Go's `inst_mod` already does the same at `parser/inst_mod.go:287-291`.

## Change

**Single file**: `ast/decl_ast.go`

**Step 1** — Add comment on `DeclBase` (line 155-162) explaining the intentional omission:

```go
// NOTE: Clone methods on Decl types intentionally do NOT copy Attributes
// or Common. This matches Python's Decl.__init__, which always resets
// self.attributes=() and self.common=None on clone. Callers that need
// these fields (e.g., inst_mod) explicitly restore them after cloning.
```

**Step 2** — In every Decl Clone() method, remove `, Attributes: d.Attributes, Common: d.Common` from the DeclBase literal. This applies to all 25+ methods. The pattern:

Before:
```go
return &FooDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
```
After:
```go
return &FooDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
```

Special case — `DefinitionDecl` (line 1091-1095):
```go
// Before:
c := &DefinitionDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}, Sn: sn}
// After:
c := &DefinitionDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}, Sn: sn}
```

Delegating types (`PropertyDecl`, `IsolateObjectDecl`, etc.) inherit the fix from their parent Clone.

## Files

- `ast/decl_ast.go` — the only file to edit (42 occurrences of `Attributes: d.Attributes`)
- `parser/inst_mod.go` — confirms explicit attribute restoration at lines 287-291 (no change needed)
- `ast/rewrite.go` — AstRewrite default case calls Clone (no change needed)

## Verification

1. `go build ./...` — confirms compilation
2. `go test ./ast/...` — AST unit tests
3. `go test ./parser/... -run TestOrdLive` — the failing golden test from log.red
4. `go test ./...` — full suite
