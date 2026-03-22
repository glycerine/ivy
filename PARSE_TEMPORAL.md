# Fix: `explicit temporal property` parse error

**Created: 2026-03-22**

## Context

Running `goivy_check isolate=cf_live ord_live.ivy` fails with:
```
parse error at 1694:14: expected 'property' after 'explicit'
```
The file contains `explicit temporal property [cf_liveness]` at line 1694. The Go parser handles `explicit property` and `temporal property` individually, but not the combination `explicit temporal property`.

## Root Cause

In Python's grammar (`ivy_parser.py:485`):
```
'top : top optexplicit opttemporal PROPERTY labeledfmla optskolem optproof'
```
Both `optexplicit` and `opttemporal` are optional modifiers before `PROPERTY`. So the grammar accepts:
- `property ...`
- `explicit property ...`
- `temporal property ...`
- `explicit temporal property ...`

Similarly for axiom (`ivy_parser.py:467`):
```
'top : top optexplicit opttemporal AXIOM lgprop'
```

And for invariant (`ivy_parser.py:518`):
```
'top : top optexplicit INVARIANT labeledfmla optproof'
```

And for definition (`ivy_parser.py:1429`):
```
'top : top optexplicit DEFINITION optlabel gdefn optproof'
```

The Go parser at `parser/decl.go:1972-1982` only handles `explicit property`:
```go
func (p *Parser) parseExplicitDecl(tok lexer.Token) ast.Node {
    p.advance()
    if p.match(lexer.PROPERTY) {
        ...
    }
    p.errorf("expected 'property' after 'explicit'")
    return nil
}
```

It doesn't check for `TEMPORAL` after `explicit`, nor for `AXIOM`, `INVARIANT`, or `DEFINITION`.

## Fix

**File: `/Users/jaten/go/src/github.com/glycerine/goivy/parser/decl.go`**

Expand `parseExplicitDecl` to handle all combinations that Python supports:

```go
func (p *Parser) parseExplicitDecl(tok lexer.Token) ast.Node {
    p.advance()

    // explicit [temporal] property ...
    // Python: 'top : top optexplicit opttemporal PROPERTY labeledfmla optskolem optproof'
    if p.match(lexer.TEMPORAL) {
        p.advance()
        if p.match(lexer.PROPERTY) {
            lf := p.parseLabeledFmla()
            lf.Explicit = true
            lf.Temporal = ast.BoolPtr(true)
            return p.setLoc(ast.NewPropertyDecl(lf), tok)
        }
        if p.match(lexer.AXIOM) {
            // explicit temporal axiom — Python: optexplicit opttemporal AXIOM lgprop
            lf := p.parseLabeledFmla()
            lf.Explicit = true
            lf.Temporal = ast.BoolPtr(true)
            return p.setLoc(ast.NewAxiomDecl(lf), tok)
        }
        p.errorf("expected 'property' or 'axiom' after 'explicit temporal'")
        return nil
    }

    // explicit property ...
    if p.match(lexer.PROPERTY) {
        lf := p.parseLabeledFmla()
        lf.Explicit = true
        return p.setLoc(ast.NewPropertyDecl(lf), tok)
    }

    // explicit axiom ...
    // Python: 'top : top optexplicit opttemporal AXIOM lgprop'
    if p.match(lexer.AXIOM) {
        lf := p.parseLabeledFmla()
        lf.Explicit = true
        return p.setLoc(ast.NewAxiomDecl(lf), tok)
    }

    // explicit invariant ...
    // Python: 'top : top optexplicit INVARIANT labeledfmla optproof'
    if p.match(lexer.INVARIANT) {
        lf := p.parseLabeledFmla()
        lf.Explicit = true
        return p.setLoc(ast.NewConjectureDecl(lf), tok)
    }

    // explicit definition ...
    // Python: 'top : top optexplicit DEFINITION optlabel gdefn optproof'
    if p.match(lexer.DEFINITION) {
        return p.parseDefinitionDeclBody(tok, true)  // pass explicit=true
    }

    p.errorf("expected 'property', 'axiom', 'invariant', or 'definition' after 'explicit'")
    return nil
}
```

Also expand `parseTemporalDecl` to handle `temporal axiom`:
```go
func (p *Parser) parseTemporalDecl(tok lexer.Token) ast.Node {
    p.advance()
    if p.match(lexer.PROPERTY) {
        lf := p.parseLabeledFmla()
        lf.Temporal = ast.BoolPtr(true)
        return p.setLoc(ast.NewPropertyDecl(lf), tok)
    }
    if p.match(lexer.AXIOM) {
        lf := p.parseLabeledFmla()
        lf.Temporal = ast.BoolPtr(true)
        return p.setLoc(ast.NewAxiomDecl(lf), tok)
    }
    p.errorf("expected 'property' or 'axiom' after 'temporal'")
    return nil
}
```

**Note**: Need to verify that `ast.NewAxiomDecl`, `ast.NewConjectureDecl` exist and handle the `Explicit`/`Temporal` fields. Also need to check if `parseLabeledFmla` handles the optskolem/optproof parts, or if those need separate handling for property vs axiom.

## Verification

1. `go build ./cmd/goivy_check/` compiles
2. `go test ./parser/...` passes
3. `./goivy_check isolate=cf_live ord_live.ivy` gets past line 1694 (may hit other parse errors further in the file)
4. Create a minimal test file with `explicit temporal property` and verify it parses
