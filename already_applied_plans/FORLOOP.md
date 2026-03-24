# C2: FOR Loop Desugaring — Full Implementation Plan

**Created:** 2026-03-24T22:00

## Context

The Go grammar's FOR loop rule (`grammar_v17.y:3986-3990`) is a stub that creates a placeholder `Atom("for",...)`. The Python source of truth (`ivy_parser.py:3040-3052`) desugars FOR into VarAction declarations, a WhileAction loop, and Sequence wrappers with `lower_var_stmts` processing. This is audit item C2 from `LALR_AUDIT.md`.

## Python Source of Truth

```python
# ivy_parser.py:3034-3052
def methcall(lhs,rhs):
    if (isinstance(lhs,App) or isinstance(lhs,Atom)) and len(lhs.args) == 0:
        return compose_atoms(lhs,rhs)
    return MethodCall(lhs,rhs)

def p_action_for_...(p):
    'complexact : FOR tterm COMMA tterm IN fmla invariants decreases sequence'
    itr,val,fmla,invars,decrs,seq = p[2],p[4],check_non_temporal(p[6]),p[7],p[8],p[9]
    iend = itr.rename('loc:end')
    ln = get_lineno(p,1)
    didx = VarAction(itr,methcall(fmla,App('begin').sln(ln)).sln(ln)).sln(ln)
    dend = VarAction(iend,methcall(fmla,App('end').sln(ln)).sln(ln)).sln(ln)
    dval = VarAction(val,methcall(fmla,App('value',itr).sln(ln)).sln(ln)).sln(ln)
    incr = AssignAction(itr,methcall(itr,App('next').sln(ln)).sln(ln)).sln(ln)
    body = Sequence(*lower_var_stmts([dval,seq,incr])).sln(ln)
    loop = WhileAction(*([App('<',itr,iend).sln(ln),body] + invars + decrs)).sln(ln)
    p[0] = Sequence(*lower_var_stmts([didx,dend,loop])).sln(ln)
```

## Existing Go Infrastructure (already available)

All key types and functions exist — no new AST types needed:

| Component | Go Location | Status |
|---|---|---|
| `MethodCall` | `ast/ast.go:431-447` | Has `Obj`, `Method` fields |
| `VarAction` | `ast/ast.go:861-876` | `NewVarAction(args...)` |
| `AssignAction` | `ast/ast.go:831-844` | `NewAssignAction(args...)` |
| `WhileAction` | `ast/ast.go:942-955` | `NewWhileAction(args...)` |
| `Sequence` | `ast/ast.go:1087-1110` | `NewSequence(stmts...)` |
| `ComposeAtomsGeneric` | `ast/rewrite.go:234-271` | Handles both App and Atom |
| `App.Rename(s)` | `ast/ast.go:310-317` | Clones App, sets rep to s |
| `LowerVarStatements` | `ast/lower_var.go:18-98` | Full implementation exists |
| `checkNonTemporal` | `grammar_v17.y` | Already used in WHILE rule |
| `PrefixNode` | `ast/lower_var.go:105-129` | Exported |
| `NodeRep` | `ast/decl.go:365-379` | Exported |

## Implementation Steps

### Step 1: Add `methcall` helper to `lalr_full/grammar_v17.y`

Add near the other helper functions (around line 308, near `fixIfPart`):

```go
// methcall matches Python methcall(lhs, rhs) at ivy_parser.py:3034-3038.
// If lhs is an App or Atom with no args, compose; otherwise create MethodCall.
func methcall(lhs, rhs ast.Node) ast.Node {
    xtracer.Trace("parser.methcall ENTER")
    switch l := lhs.(type) {
    case *ast.App:
        if len(l.Terms) == 0 {
            return ast.ComposeAtomsGeneric(l, rhs)
        }
    case *ast.Atom:
        if len(l.Terms) == 0 {
            return ast.ComposeAtomsGeneric(l, rhs)
        }
    }
    return &ast.MethodCall{Obj: lhs, Method: rhs}
}
```

**Python line-by-line match:**
- `isinstance(lhs, App) or isinstance(lhs, Atom)` → Go type switch on `*ast.App` / `*ast.Atom`
- `len(lhs.args) == 0` → `len(l.Terms) == 0`
- `compose_atoms(lhs, rhs)` → `ast.ComposeAtomsGeneric(lhs, rhs)` (handles both App and Atom)
- `MethodCall(lhs, rhs)` → `&ast.MethodCall{Obj: lhs, Method: rhs}`

### Step 2: Add `sln` helper to `lalr_full/grammar_v17.y`

Python's `.sln(ln)` sets lineno and returns self. Go needs a helper since methods don't return `self`:

```go
// sln sets lineno on a node and returns it. Matches Python AST.sln().
func sln(n ast.Node, ln ast.Location) ast.Node {
    n.SetLineno(ln)
    return n
}
```

### Step 3: Replace FOR loop stub in `grammar_v17.y:3986-3990`

Replace:
```go
| TOK_FOR tterm TOK_COMMA tterm TOK_IN fmla invariants decreases sequence
{
    xtracer.Trace("parser.p_action_for_tterm_comma_tterm_in_expr_invariants_decreases_lcb_action_rcb ENTER (complexact)")
    $$ = ast.NewAtom("for", $2, $4, $6, $9)
}
```

With:
```go
| TOK_FOR tterm TOK_COMMA tterm TOK_IN fmla invariants decreases sequence
{
    xtracer.Trace("parser.p_action_for_tterm_comma_tterm_in_expr_invariants_decreases_lcb_action_rcb ENTER (complexact)")

    // Python: itr,val,fmla,invars,decrs,seq = p[2],p[4],check_non_temporal(p[6]),p[7],p[8],p[9]
    itr := $2                       // *ast.App from tterm
    val := $4                       // *ast.App from tterm
    fmla := checkNonTemporal($6)    // formula
    invars := $7                    // []ast.Node from invariants
    decrs := $8                     // []ast.Node from decreases
    seq := $9                       // ast.Node from sequence

    // Python: iend = itr.rename('loc:end')
    var iend ast.Node
    if itrApp, ok := itr.(*ast.App); ok {
        iend = itrApp.Rename("loc:end")
    } else if itrAtom, ok := itr.(*ast.Atom); ok {
        iend = itrAtom.Rename("loc:end")
    } else {
        iend = itr // fallback
    }

    // Python: ln = get_lineno(p,1)
    ln := tokLineno(v17lex.(*v17LexAdapter), $1)

    // Python: didx = VarAction(itr, methcall(fmla, App('begin').sln(ln)).sln(ln)).sln(ln)
    appBegin := ast.NewApp(ast.NewSymbol("begin", nil))
    appBegin.SetLineno(ln)
    mcBegin := methcall(fmla, appBegin)
    mcBegin.SetLineno(ln)
    didx := ast.NewVarAction(itr, mcBegin)
    didx.SetLineno(ln)

    // Python: dend = VarAction(iend, methcall(fmla, App('end').sln(ln)).sln(ln)).sln(ln)
    appEnd := ast.NewApp(ast.NewSymbol("end", nil))
    appEnd.SetLineno(ln)
    mcEnd := methcall(fmla, appEnd)
    mcEnd.SetLineno(ln)
    dend := ast.NewVarAction(iend, mcEnd)
    dend.SetLineno(ln)

    // Python: dval = VarAction(val, methcall(fmla, App('value', itr).sln(ln)).sln(ln)).sln(ln)
    appValue := ast.NewApp(ast.NewSymbol("value", nil), itr)
    appValue.SetLineno(ln)
    mcValue := methcall(fmla, appValue)
    mcValue.SetLineno(ln)
    dval := ast.NewVarAction(val, mcValue)
    dval.SetLineno(ln)

    // Python: incr = AssignAction(itr, methcall(itr, App('next').sln(ln)).sln(ln)).sln(ln)
    appNext := ast.NewApp(ast.NewSymbol("next", nil))
    appNext.SetLineno(ln)
    mcNext := methcall(itr, appNext)
    mcNext.SetLineno(ln)
    incr := ast.NewAssignAction(itr, mcNext)
    incr.SetLineno(ln)

    // Python: body = Sequence(*lower_var_stmts([dval, seq, incr])).sln(ln)
    bodyStmts := ast.LowerVarStatements([]ast.Node{dval, seq, incr})
    body := ast.NewSequence(bodyStmts...)
    body.SetLineno(ln)

    // Python: loop = WhileAction(*([App('<', itr, iend).sln(ln), body] + invars + decrs)).sln(ln)
    ltCond := ast.NewApp(ast.NewSymbol("<", nil), itr, iend)
    ltCond.SetLineno(ln)
    loopArgs := []ast.Node{ltCond, body}
    loopArgs = append(loopArgs, invars...)
    loopArgs = append(loopArgs, decrs...)
    loop := ast.NewWhileAction(loopArgs...)
    loop.SetLineno(ln)

    // Python: p[0] = Sequence(*lower_var_stmts([didx, dend, loop])).sln(ln)
    outerStmts := ast.LowerVarStatements([]ast.Node{didx, dend, loop})
    result := ast.NewSequence(outerStmts...)
    result.SetLineno(ln)
    $$ = result
}
```

### Step 4: Unit Tests

Create `lalr_full/for_desugar_test.go` with these test cases:

#### 4a. Test `methcall` helper

```go
func TestMethcall(t *testing.T) {
    // Case 1: App with no args → compose
    lhs := ast.NewApp(ast.NewSymbol("x", nil))
    rhs := ast.NewApp(ast.NewSymbol("begin", nil))
    result := methcall(lhs, rhs)
    // Should be composed App with rep "x.begin", not MethodCall
    if _, ok := result.(*ast.MethodCall); ok {
        t.Error("expected composed App, got MethodCall")
    }
    if ast.NodeRep(result) != "x.begin" {
        t.Errorf("expected rep 'x.begin', got %q", ast.NodeRep(result))
    }

    // Case 2: App with args → MethodCall
    lhsWithArgs := ast.NewApp(ast.NewSymbol("x", nil), ast.NewApp(ast.NewSymbol("y", nil)))
    result2 := methcall(lhsWithArgs, rhs)
    mc, ok := result2.(*ast.MethodCall)
    if !ok {
        t.Error("expected MethodCall for App with args")
    }
    if mc.Obj != lhsWithArgs || mc.Method != rhs {
        t.Error("MethodCall fields wrong")
    }

    // Case 3: Atom with no args → compose
    atomLhs := ast.NewAtom("x")
    result3 := methcall(atomLhs, rhs)
    if _, ok := result3.(*ast.MethodCall); ok {
        t.Error("expected composed node for Atom with no args, got MethodCall")
    }

    // Case 4: Atom with args → MethodCall
    atomWithArgs := ast.NewAtom("x", ast.NewAtom("y"))
    result4 := methcall(atomWithArgs, rhs)
    if _, ok := result4.(*ast.MethodCall); !ok {
        t.Error("expected MethodCall for Atom with args")
    }
}
```

#### 4b. Test App.Rename

```go
func TestAppRename(t *testing.T) {
    app := ast.NewApp(ast.NewSymbol("itr", nil))
    app.ASort = ast.NewSymbol("someSort", nil)
    renamed := app.Rename("loc:end")
    if renamed.Relname() != "loc:end" {
        t.Errorf("expected rep 'loc:end', got %q", renamed.Relname())
    }
    // Sort should be preserved
    if renamed.ASort == nil {
        t.Error("expected ASort to be preserved")
    }
}
```

#### 4c. Test full FOR loop desugaring via parser

```go
func TestForLoopDesugaring(t *testing.T) {
    // Parse a minimal FOR loop and verify the resulting AST structure
    // is a Sequence (outermost) containing the desugared form.
    //
    // Input: "for x, v in rng { skip }"
    // Expected: Sequence(lower_var_stmts([
    //   VarAction(x, x.begin),
    //   VarAction(loc:end, x.end),
    //   WhileAction(App("<", x, loc:end),
    //     Sequence(lower_var_stmts([
    //       VarAction(v, x.value(x)),
    //       skip,
    //       AssignAction(x, x.next)
    //     ]))
    //   )
    // ]))

    // Test that result is a Sequence (not an Atom("for",...))
    // Test that result contains no Atom with rep "for"
    // Test that the WhileAction is present in the desugared form
    // Test that VarAction nodes are created for itr, iend, val
    // Test that AssignAction for incr is present
}
```

#### 4d. Test canon output matches Python

```go
func TestForLoopCanon(t *testing.T) {
    // If a golden canon test file exists for FOR loops, compare.
    // Otherwise, parse a FOR loop and verify the canon output
    // has "(sequence" at the top level, contains "(whileAction",
    // contains "(varAction", and contains "(assignAction".
    // Does NOT contain "(atom rep:\"for\"".
}
```

#### 4e. Test methcall with MethodCall type produces correct canon

```go
func TestMethodCallCanon(t *testing.T) {
    mc := &ast.MethodCall{
        Obj:    ast.NewApp(ast.NewSymbol("x", nil), ast.NewApp(ast.NewSymbol("y", nil))),
        Method: ast.NewApp(ast.NewSymbol("next", nil)),
    }
    canon := string(mc.Canon())
    if !strings.Contains(canon, "methodCall") {
        t.Errorf("expected methodCall in canon, got %s", canon)
    }
}
```

## Files to Modify

| File | Change |
|---|---|
| `lalr_full/grammar_v17.y` | Add `methcall` helper (~15 lines), add `sln` helper (~4 lines), replace FOR rule (~50 lines) |
| `lalr_full/for_desugar_test.go` | **New file** — unit tests for methcall, App.Rename, FOR desugaring (~120 lines) |

## No New AST Types Needed

All types (`MethodCall`, `VarAction`, `AssignAction`, `WhileAction`, `Sequence`) already exist in `ast/ast.go`. The helper `ComposeAtomsGeneric` already exists in `ast/rewrite.go`. `App.Rename` already exists in `ast/ast.go:310`. `LowerVarStatements` already exists in `ast/lower_var.go:18`.

## Key Differences from LALR_AUDIT.md

The audit (C2) referenced an **older version** of the Python code that used `itr.clone([App(itr.rep + '__end')])`, `Not(Atom("=", ...))`, and `VarAction(val)` with no rhs. The **current Python source** (confirmed at `ivy_parser.py:3040-3052`) uses:
- `itr.rename('loc:end')` — not `__end` suffix
- `App('<', itr, iend)` — not `Not(Atom("=",...))`
- `VarAction(val, methcall(fmla, App('value', itr)))` — val has an rhs
- `.sln(ln)` chaining on every node

This plan follows the **current Python source**, not the outdated audit.

## Verification

1. `cd ~/goivy && go build ./lalr_full/...` — compiles
2. `cd ~/goivy && go test ./lalr_full/ -run TestMethcall` — unit tests pass
3. `cd ~/goivy && go test ./lalr_full/ -run TestForLoop` — integration tests pass
4. `cd ~/goivy && make golden` — verify the divergence point advances (currently at line 25733)
