# C6: METHOD Action Not Prepending `self` Parameter

**Created:** 2026-03-25T00:30

## Context

The Go grammar rule for `top optimpex actmeth SYMBOL optargs optreturns optactiondef` (grammar_v17.y:1166-1195) parses the `actmeth` boolean ($3) but never checks it. When `actmeth == true` (METHOD keyword), Python prepends an `App('self')` with sort `This()` to the formals list. The Go code skips this entirely. Also missing: CrashAction clone handling when adef is `*ast.CrashAction`.

## Python Source of Truth (`ivy_parser.py:2078-2099`)

```python
def p_top_optimpex_action_symbol_optargs_optreturns_eq_action(p):
    'top : top optimpex actmeth SYMBOL optargs optreturns optactiondef'
    p[0] = p[1]
    adef = p[7]
    if not hasattr(adef,'lineno'):
        adef.lineno = get_lineno(p,4)
    formals = p[5]
    if p[3]:  # actmeth is True (METHOD keyword)
        arg0 = App('self')
        arg0.sort = This()
        arg0.lineno = get_lineno(p,4)
        formals = [arg0] + formals
    if isinstance(adef, CrashAction):
        adef = adef.clone([Atom(This(), formals)])
    the_atom = Atom(p[4],[])
    the_atom.lineno = adef.lineno
    actdef = ActionDef(the_atom, adef, formals=formals, returns=p[6])
    ...
```

## Current Go Code (`grammar_v17.y:1166-1195`)

```go
| top optimpex actmeth SYMBOLx optargs optreturns optactiondef
{
    $$ = $1
    adef := $7
    // ... lineno handling ...
    theAtom := ast.NewAtom($4.Val)
    theAtom.SetLineno(lineno)
    actdef := ast.NewActionDef(theAtom, adef, $5, $6)  // $5 = formals, never modified
    actdef.SetLineno(lineno)
    decl := ast.NewActionDecl(actdef)
    decl.SetLineno(lineno)
    $$.declare(decl)
    if $2 != nil { $$.declare($2) }
}
```

**What's missing:** `$3` (actmeth bool) is never checked. No self parameter, no CrashAction clone.

## Implementation

### Change: grammar_v17.y:1166-1195

Replace the rule body to add self-parameter prepending and CrashAction handling:

```go
| top optimpex actmeth SYMBOLx optargs optreturns optactiondef
{
    xtracer.Trace("parser.p_top_optimpex_action_symbol_optargs_optreturns_eq_action ENTER (top)")
    $$ = $1
    adef := $7
    var lineno ast.Location
    if adef != nil {
        if b, ok := adef.(interface{ HasLocSet() bool }); ok && b.HasLocSet() {
            lineno = adef.GetLineno()
        }
    }
    if lineno == (ast.Location{}) {
        lineno = tokLineno(v17lex.(*v17LexAdapter), $4)
    }

    // Python: formals = p[5]
    formals := $5

    // Python: if p[3]: (actmeth is True for METHOD)
    if $3 {
        // Python: arg0 = App('self')
        // Python: arg0.sort = This()
        // Python: arg0.lineno = get_lineno(p,4)
        selfArg := ast.NewApp(ast.NewSymbol("self", nil))
        selfArg.ASort = &ast.This{}
        selfArg.SetLineno(lineno)
        // Python: formals = [arg0] + formals
        formals = append([]ast.Node{selfArg}, formals...)
    }

    // Python: if isinstance(adef, CrashAction):
    //             adef = adef.clone([Atom(This(), formals)])
    if ca, ok := adef.(*ast.CrashAction); ok {
        thisAtom := ast.NewAtom("this", formals...)
        thisAtom.SetLineno(lineno)
        adef = ca.Clone([]ast.Node{thisAtom})
    }

    theAtom := ast.NewAtom($4.Val)
    theAtom.SetLineno(lineno)
    // Python: ActionDef(theAtom, adef, formals=formals, returns=p[6])
    actdef := ast.NewActionDef(theAtom, adef, formals, $6)
    actdef.SetLineno(lineno)
    decl := ast.NewActionDecl(actdef)
    decl.SetLineno(lineno)
    $$.declare(decl)
    if $2 != nil {
        $$.declare($2)
    }
}
```

### Line-by-line Python→Go match

| Python | Go |
|---|---|
| `formals = p[5]` | `formals := $5` |
| `if p[3]:` | `if $3 {` |
| `arg0 = App('self')` | `selfArg := ast.NewApp(ast.NewSymbol("self", nil))` |
| `arg0.sort = This()` | `selfArg.ASort = &ast.This{}` |
| `arg0.lineno = get_lineno(p,4)` | `selfArg.SetLineno(lineno)` |
| `formals = [arg0] + formals` | `formals = append([]ast.Node{selfArg}, formals...)` |
| `if isinstance(adef, CrashAction):` | `if ca, ok := adef.(*ast.CrashAction); ok {` |
| `adef.clone([Atom(This(), formals)])` | `adef = ca.Clone([]ast.Node{thisAtom})` |
| `Atom(This(), formals)` | `ast.NewAtom("this", formals...)` |

### Key type notes

- `actmeth` is `%type <bval>` → Go `bool`. `$3` is directly usable in `if $3 {`.
- `App('self')` → `ast.NewApp(ast.NewSymbol("self", nil))` — App takes a Symbol as its first arg.
- `This()` as sort → `&ast.This{}` (not `ast.NewSymbol("this", nil)` — the audit suggested Symbol but Python uses the actual `This()` class).
- `Atom(This(), formals)` → `ast.NewAtom("this", formals...)` — Atom takes rep string + terms. `This().rep` is `"this"`.
- CrashAction check is **outside** the `if p[3]` block in Python (applies to both ACTION and METHOD).

## Unit Tests

New file: `lalr_full/method_self_test.go`

### Test 1: `TestMethodPrependsSelf`
Parse a METHOD declaration, verify the ActionDef's formals list starts with an App whose rep is "self" and sort is This.
```ivy
#lang 1.7
type t
method foo(x:t) = { skip }
```

### Test 2: `TestActionDoesNotPrependSelf`
Parse an ACTION declaration, verify formals does NOT contain a "self" parameter.
```ivy
#lang 1.7
type t
action bar(x:t) = { skip }
```

### Test 3: `TestMethodSelfHasThisSort`
Parse a METHOD, extract the self App, verify `.ASort` canons as `(this ...)`.

### Test 4: `TestMethodSelfIsFirstFormal`
Parse a METHOD with multiple params, verify self is first in formals, followed by the declared params.

### Test 5: `TestMethodCrashAction`
Parse `method foo = *`, verify CrashAction clone is produced with Atom("this", [self]).

### Test 6: `TestActionCrashAction`
Parse `action bar = *`, verify CrashAction is produced but without self prepend.

### Test 7: `TestMethodNoArgs`
Parse `method foo = { skip }`, verify self is the only formal.

### Test 8: `TestSelfAppConstruction`
Directly construct `App("self")` with `This{}` sort and verify canon output.

## Files to Modify

| File | Change |
|---|---|
| `lalr_full/grammar_v17.y:1166-1195` | Add self-prepend and CrashAction handling (~30 lines → ~45 lines) |
| `lalr_full/method_self_test.go` | **New file** — unit tests (~250 lines) |

## Verification

1. `cd lalr_full && go generate` — regenerate parser
2. `go build ./lalr_full/...` — compiles
3. `go test ./lalr_full/ -run TestMethod -v` — new tests pass
4. `go test ./lalr_full/ -run TestAction -v` — no-self tests pass
5. `go test ./lalr_full/...` — all existing tests still pass
6. `make golden` — verify no regression
