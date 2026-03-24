# LALR Grammar Audit: Go grammar_v17.y vs Python ivy_parser.py

**Date:** 2026-03-24
**Source of truth:** `~/pyivy/ivy/ivy/ivy_parser.py` and `~/pyivy/ivy/ivy/ivy_logic_parser.py`
**Go port:** `~/goivy/lalr_full/grammar_v17.y` and `~/goivy/lalr_full/inst_mod.go`

This audit identifies every discrepancy where the Go grammar diverges from the Python source of truth. Each finding includes exact file paths, line numbers, code snippets, and the fix required.

---

## CRITICAL DISCREPANCIES (affect correctness of parsed AST)

### C1. LOCAL action missing `loc:` prefix substitution

**Python** (`ivy_parser.py:3187-3195`):
```python
def p_action_local_params_lcb_action_rcb(p):
    'complexact : LOCAL lparams sequence'
    lsyms = [s.prefix('loc:') for s in p[2]]
    subst = dict((x.rep,y.rep) for x,y in zip(p[2],lsyms))
    action = subst_prefix_atoms_ast(p[3],subst,None,None)
    p[0] = LocalAction(*(lsyms+[action]))
    p[0].lineno = get_lineno(p,1)
```

**Go** (`grammar_v17.y:3991-3999`):
```go
| TOK_LOCAL lparams sequence
{
    xtracer.Trace("parser.p_action_local_params_lcb_action_rcb ENTER (complexact)")
    args := append($2, $3)
    la := ast.NewLocalAction(args...)
    la.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    $$ = la
}
```

**Problem:** Go passes through raw `lparams` and `sequence` without renaming. Python:
1. Creates `lsyms` — each lparam with `loc:` prefix via `s.prefix('loc:')`
2. Builds `subst` dict mapping original name → prefixed name
3. Applies `subst_prefix_atoms_ast(sequence, subst, nil, nil)` to rename variables in body
4. Creates `LocalAction(*(lsyms + [action]))` — prefixed params + substituted body

**Fix** — replace the Go rule with:
```go
| TOK_LOCAL lparams sequence
{
    xtracer.Trace("parser.p_action_local_params_lcb_action_rcb ENTER (complexact)")
    bounds := $2
    lsyms := make([]ast.Node, len(bounds))
    subst := make(map[string]string)
    for i, s := range bounds {
        lsyms[i] = ast.PrefixNode(s, "loc:")
        subst[ast.NodeRep(s)] = ast.NodeRep(lsyms[i])
    }
    action := ast.SubstPrefixAtomsAst($3, subst, nil, nil, nil)
    args := append(lsyms, action)
    la := ast.NewLocalAction(args...)
    la.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    $$ = la
}
```

Uses existing `ast.PrefixNode` (ast/lower_var.go:103), `ast.NodeRep` (ast/decl.go:360), `ast.SubstPrefixAtomsAst` (ast/rewrite.go:730).

---

### C2. FOR loop is a stub

**Python** (`ivy_parser.py:3040-3052`):
```python
def p_action_for_tterm_comma_tterm_in_expr_invariants_decreases_lcb_action_rcb(p):
    'complexact : FOR tterm COMMA tterm IN fmla invariants decreases sequence'
    itr = p[2]
    val = p[4]
    rng = p[6]
    invars = p[7]
    decrs = p[8]
    body = p[9]
    iend = itr.clone([App(itr.rep + '__end')])
    iend.sort = itr.sort
    stmts = [VarAction(itr, methcall(rng, App('begin'))),
             VarAction(iend, methcall(rng, App('end'))),
             VarAction(val)]
    cond = Not(Atom('=', [itr, iend]))
    assign = AssignAction(val, App(itr.rep, [Atom('*>',itr,val)]))
    incr = AssignAction(itr, methcall(itr, App('next')))
    wbody = Sequence(*(lower_var_stmts(stmts + [assign] + [body] + [incr])))
    wbody.lineno = get_lineno(p,1)
    p[0] = WhileAction(*([cond, wbody] + invars + decrs))
    p[0].lineno = get_lineno(p,1)
```

**Go** (`grammar_v17.y:3986-3990`):
```go
| TOK_FOR tterm TOK_COMMA tterm TOK_IN fmla invariants decreases sequence
{
    xtracer.Trace("parser.p_action_for_tterm_comma_tterm_in_expr_invariants_decreases_lcb_action_rcb ENTER (complexact)")
    $$ = ast.NewAtom("for", $2, $4, $6, $9)
}
```

**Problem:** Go creates a placeholder `Atom("for",...)` instead of desugaring the FOR loop. The entire Python desugaring is missing: `methcall`, `VarAction` creation, `lower_var_stmts`, `AssignAction`, `WhileAction` wrapping.

**Fix** — implement the full desugaring. Requires:
1. A `methcall(lhs, rhs)` helper matching Python `ivy_parser.py:3034-3038`:
   ```python
   def methcall(lhs,rhs):
       return compose_atoms(lhs,rhs) if isinstance(lhs,Atom) else MethodCall(lhs,rhs)
   ```
   Go equivalent: use `ast.ComposeAtoms(lhs, rhs)` for `*Atom`, else create `MethodCall` (need to add this type or use `App`).
2. The body construction with `VarAction`, `AssignAction`, `Not(Atom("=",...))`, `lower_var_stmts`, `Sequence`, `WhileAction`.
3. Clone of `itr` to create `iend` with `__end` suffix.

This is a large change. The `methcall` helper and `MethodCall` type may need to be added to `ast/`.

---

### C3. WHILE action missing `fixIfPart` on body

**Python** (`ivy_parser.py:3028-3032`):
```python
def p_action_while_somefmla_invariants_decreases_lcb_action_rcb(p):
    'complexact : WHILE somefmla invariants decreases sequence'
    p[0] = WhileAction(*([check_non_temporal(p[2]),
                          fix_if_part(p[2],p[5])] + p[3] + p[4]))
    p[0].lineno = get_lineno(p,1)
```

**Go** (`grammar_v17.y:3974-3985`):
```go
| TOK_WHILE somefmla invariants decreases sequence
{
    xtracer.Trace("parser.p_action_while_somefmla_invariants_decreases_lcb_action_rcb ENTER (complexact)")
    cond := checkNonTemporal($2)
    args := []ast.Node{cond, $5}
    args = append(args, $3...)
    args = append(args, $4...)
    w := ast.NewWhileAction(args...)
    w.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    $$ = w
}
```

**Problem:** Go passes `$5` (sequence) directly as the body. Python calls `fix_if_part(p[2], p[5])` which, when the condition is a `Some`/`SomeMin`/`SomeMax`, applies `loc:` prefix substitution on the body.

**Fix** — change `$5` to `fixIfPart($2, $5)`:
```go
    cond := checkNonTemporal($2)
    body := fixIfPart($2, $5)
    args := []ast.Node{cond, body}
```

The `fixIfPart` function already exists (grammar_v17.y:310) and already handles `Some`/`SomeMin`/`SomeMax`.

---

### C4. CLASS does not use `createObject`

**Python** (`ivy_parser.py:737-747`):
```python
def p_top_class_symbol_eq_lcb_top_rcb(p):
    'top : top CLASS objsym objectargs EQ LCB optdotdotdot top RCB objectend'
    scnst = Atom(This())
    scnst.lineno = get_lineno(p,2)
    tdfn = TypeDef(scnst, UninterpretedSort())
    tdfn.lineno = get_lineno(p,2)
    p[8].declare(TypeDecl(tdfn))
    p[8].decls = [p[8].decls[-1]] + p[8].decls[:-1]  # move TypeDecl to front
    create_object(p[0], p[3], p[4], p[8], get_lineno(p,3), p[7])
```

**Go** (`grammar_v17.y:849-869`):
```go
// Does NOT call createObject. Manually declares ObjectDecl, TypeDecl,
// and iterates decls without inst_mod or prefix substitution.
```

**Problem:** Go misses:
1. The `create_object` call which does `inst_mod` with prefix substitution
2. Moving TypeDecl to front of decls list (`p[8].decls[-1] + p[8].decls[:-1]`)
3. Proper continuation handling (optdotdotdot / `$7`)
4. The VariantDecl creation (for subclass variant)

**Fix** — rewrite the Go rule to:
1. Create `TypeDecl(TypeDef(Atom(This()), UninterpretedSort()))` with proper lineno
2. Declare it into `$8` accumulator
3. Move it to front: reorder `$8.decls` so the last element (TypeDecl) comes first
4. Call `createObject($1, $3, $4, $8, lineno, $7)`

The existing `createObject` function (grammar_v17.y:343) handles `inst_mod` and prefix substitution.

---

### C5. SUBCLASS does not use `createObject`

**Python** (`ivy_parser.py:749-761`):
```python
def p_top_subclass_symbol_eq_lcb_top_rcb(p):
    'top : top SUBCLASS objsym OF atype EQ LCB optdotdotdot top RCB objectend'
    scnst = Atom(This())
    scnst.lineno = get_lineno(p,2)
    tdfn = TypeDef(scnst, UninterpretedSort())
    tdfn.lineno = get_lineno(p,2)
    p[9].declare(TypeDecl(tdfn))
    vd = VariantDecl(Atom(p[5]),Atom(This()))
    vd.lineno = get_lineno(p,2)
    p[9].declare(vd)
    p[9].decls = [p[9].decls[-2], p[9].decls[-1]] + p[9].decls[:-2]
    create_object(p[0], p[3], [], p[9], get_lineno(p,3), p[8])
```

**Go** (`grammar_v17.y:871-885`): Does NOT call `createObject`. Misses TypeDecl, VariantDecl, decl reordering, and prefix substitution.

**Fix** — same pattern as C4: declare TypeDecl and VariantDecl into `$9`, reorder, call `createObject`. Note subclass uses empty objectargs `[]` and passes `$8` (optdotdotdot) as continuation.

---

### C6. METHOD action not prepending `self` parameter

**Python** (`ivy_parser.py:2086-2092`):
```python
if p[3]:  # actmeth is True (METHOD keyword)
    arg0 = App('self')
    arg0.sort = This()
    arg0.lineno = get_lineno(p,4)
    formals = [arg0] + formals
if isinstance(adef, CrashAction):
    adef = adef.clone([Atom(This(), formals)])
```

**Go** (`grammar_v17.y:1088-1117`): The `actmeth` value (`$3`) is parsed but its boolean value is never checked. When `actmeth == true` (METHOD), Go does NOT prepend `App("self")` with sort `This()` to the formals list. Also missing: `CrashAction` clone handling.

**Fix** — after building `formals` from `$5`, add:
```go
if isMethod {  // $3 is true for METHOD
    selfArg := ast.NewApp(ast.NewSymbol("self", nil))
    selfArg.ASort = ast.NewSymbol("this", nil)  // This()
    selfArg.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $4))
    formals = append([]ast.Node{selfArg}, formals...)
}
```

---

### C7. `IF *` creates wrong type (Ite instead of ChoiceAction)

**Python** (`ivy_parser.py:3054-3058`):
```python
def p_action_if_times_lcb_action_rcb_else_LCB_action_RCB(p):
    'complexact : IF TIMES sequence ELSE action'
    p[0] = ChoiceAction(p[3], p[5])
    p[0].lineno = get_lineno(p,1)
```

**Go** (`grammar_v17.y:3967-3973`):
```go
| TOK_IF TOK_TIMES sequence TOK_ELSE action
{
    xtracer.Trace("parser.p_action_if_times_lcb_action_rcb_else_LCB_action_RCB ENTER (complexact)")
    choice := ast.NewIte(ast.NewSymbol("*", nil), $3, $5)
    choice.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    $$ = choice
}
```

**Problem:** Go creates `Ite("*", thenPart, elsePart)` — the wrong AST type entirely. Python creates `ChoiceAction(thenPart, elsePart)`.

**Fix** — need `ChoiceAction` in the `ast` package. Check if it exists:
- `ast/ast.go`: search for `ChoiceAction` — if not present, add it matching Python's `ChoiceAction(AST)` which takes 2 args (then, else).
- Then change the grammar rule to:
```go
    choice := ast.NewChoiceAction($3, $5)
    choice.SetLineno(...)
    $$ = choice
```

If `ChoiceAction` only exists in `actions/action.go`, it needs to be ported to `ast/ast.go` (or `ast/` package) so the grammar can use it.

---

### C8. DEFINITION missing DefinitionSchema conversion for explicit

**Python** (`ivy_parser.py:1671-1683`):
```python
def p_top_definition_optlabel_gdefn_optproof(p):
    'top : top optexplicit DEFINITION optlabel gdefn optproof'
    foo = p[5]
    if p[2]:  # optexplicit is True
        foo = DefinitionSchema(*foo.args)
        foo.lineno = p[5].lineno
    lf = LabeledFormula(p[4], foo)
    lf.lineno = p[5].lineno
    d = DefinitionDecl(addlabel(lf, 'def'))
    ...
```

**Go** (`grammar_v17.y:887-899`):
```go
| top optexplicit TOK_DEFINITION optlabel gdefn optproof
{
    // ... never checks $2 (optexplicit)
    // Always uses raw $5 without converting to DefinitionSchema
}
```

**Problem:** When `optexplicit` is true, Python wraps the definition in `DefinitionSchema`. Go always uses the raw definition. `DefinitionSchema` exists in Go at `ast/formula.go:316`.

**Fix** — add after getting `$5`:
```go
gdefn := $5
if $2 != nil {  // optexplicit is set
    if def, ok := gdefn.(*ast.Definition); ok {
        ds := &ast.DefinitionSchema{Lhs: def.Lhs, Rhs: def.Rhs}
        ds.SetLineno(def.GetLineno())
        gdefn = ds
    }
}
```

---

### C9. TYPE with Range missing InterpretDecl and UninterpretedSort substitution

**Python** (`ivy_parser.py:1780-1797`):
```python
def p_top_type_symbol_eq_sort(p):
    'top : top optfinite optghost TYPE typesymbol EQ sort'
    scnst = Atom(p[5])
    scnst.lineno = get_lineno(p,3)
    defsort = UninterpretedSort() if isinstance(p[7], Range) else p[7]
    tdfn = TypeDef(scnst, defsort) if not p[3] else GhostTypeDef(scnst, defsort)
    tdfn.lineno = get_lineno(p,3)
    if p[2]:
        tdfn.finite = True
    p[0].declare(TypeDecl(tdfn))
    if isinstance(p[7], Range):
        imp = Implies(scnst, p[7])
        imp.lineno = get_lineno(p,3)
        p[0].declare(InterpretDecl(addlabel(mk_lf(imp), 'interp')))
```

**Go** (`grammar_v17.y:1014-1026`):
```go
tdfn := &ast.TypeDef{Name: scnst, Value: $7}
// ... never checks if $7 is a Range
// ... never creates InterpretDecl
// ... never substitutes UninterpretedSort
```

**Problem:** When the sort is a `Range`, Python:
1. Replaces the sort with `UninterpretedSort()` in the TypeDef
2. Creates an additional `InterpretDecl(addlabel(mk_lf(Implies(scnst, range)), 'interp'))`

Go does neither. Also, Go never checks `optghost` ($3) to use `GhostTypeDef`.

**Fix** — after creating `scnst`, add:
```go
sortNode := $7
isRange := false
if _, ok := sortNode.(*ast.Range); ok {
    isRange = true
    sortNode = &ast.UninterpretedSort{}
}
var tdfn ast.Node
if $3 != nil {  // optghost
    tdfn = &ast.GhostTypeDef{Name: scnst, Value: sortNode}
} else {
    tdfn = &ast.TypeDef{Name: scnst, Value: sortNode}
}
// ... declare TypeDecl ...
if isRange {
    imp := ast.NewImplies(scnst, $7)
    imp.SetLineno(...)
    lf := mkLF(imp)
    lf = addLabel(lf, "interp")
    $$.declare(ast.NewInterpretDecl(lf))
}
```

---

## MODERATE DISCREPANCIES

### M1. THUNK action is a stub

**Python** (`ivy_parser.py:3229-3235`):
```python
def p_action_thunk_labelname_symbol_optargs_colon_atype_assign_sequence(p):
    'complexact : THUNK labelname SYMBOL optargs COLON atype ASSIGN sequence'
    p[0] = ThunkAction(Atom(p[2][1:-1],[]), action, Atom(p[6]), p[8])
    p[0].lineno = get_lineno(p,1)
```

**Go** (`grammar_v17.y:4006-4011`):
```go
| TOK_THUNK labelname SYMBOLx optargs TOK_COLON atype TOK_ASSIGN sequence
{
    $$ = ast.NewAtom("thunk", ast.NewAtom($2.Val), ast.NewAtom($3.Val), $8)
}
```

**Problem:** Go creates a placeholder `Atom("thunk",...)`. Drops `optargs` ($4) and `atype` ($6). `ThunkAction` exists in `ast/ast.go`.

**Fix** — create proper `ThunkAction`:
```go
label := ast.NewAtom(strings.Trim($2.Val, "[]"))
action := ast.NewAtom($3.Val)
sortNode := $6  // atype
seq := $8
ta := &ast.ThunkAction{Label: label, Action: action, Sort: sortNode, Body: seq}
ta.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
$$ = ta
```

---

### M2. LET action is a stub

**Python** (`ivy_parser.py:3296-3299`):
```python
def p_action_let_eqns_lcb_action_rcb(p):
    'complexact : LET eqns sequence'
    p[0] = LetAction(*(p[2]+[p[3]]))
    p[0].lineno = get_lineno(p,1)
```

**Go** (`grammar_v17.y:4000-4005`):
```go
| TOK_LET eqns sequence
{
    $$ = ast.NewAtom("let", append($2, $3)...)
}
```

**Problem:** Go creates `Atom("let",...)`. Need `LetAction` in `ast` package.

**Fix** — add `LetAction` to `ast/ast.go` if not present. Then:
```go
args := append($2, $3)
la := ast.NewLetAction(args...)
la.SetLineno(...)
$$ = la
```

---

### M3. DECREASES missing Ranking wrapper and checkNonTemporal

**Python** (`ivy_parser.py:3016-3021`):
```python
def p_decreases_decreases_fmla(p):
    'decreases : DECREASES fmla'
    rank = Ranking(check_non_temporal(p[2]))
    rank.lineno = get_lineno(p,1)
    p[0] = [rank]
```

**Go** (`grammar_v17.y:4139-4144`):
```go
| TOK_DECREASES fmla
{
    $$ = []ast.Node{$2}
}
```

**Problem:** Missing `Ranking` wrapper and `checkNonTemporal` call.

**Fix** — add `Ranking` type to `ast` if not present. Then:
```go
fmla := checkNonTemporal($2)
rank := ast.NewRanking(fmla)
rank.SetLineno(...)
$$ = []ast.Node{rank}
```

---

### M4. TYPE with GHOST not using GhostTypeDef

**Python** (`ivy_parser.py:1774`):
```python
tdfn = (GhostTypeDef if p[3] else TypeDef)(scnst, defsort)
```

**Go** (`grammar_v17.y:1000-1026`): Always uses `TypeDef`. `GhostTypeDef` exists in `ast/decl.go` but is never used in the grammar.

**Fix** — covered in C9 fix. Also fix the uninterpreted type rule at line 1000-1012:
```go
if $2 != nil {  // optghost
    tdfn = &ast.GhostTypeDef{Name: scnst, Value: &ast.UninterpretedSort{}}
} else {
    tdfn = &ast.TypeDef{Name: scnst, Value: &ast.UninterpretedSort{}}
}
```

---

### M5. INVARIANT missing check_unprovable gate

**Python** (`ivy_parser.py:608-622`):
```python
if not lf.unprovable or check_unprovable.get():
    p[0].declare(d)
```

**Go** (`grammar_v17.y:800-801`): Comment says "we always declare for now" and always calls `$$.declare(d)`.

**Fix** — implement `check_unprovable` equivalent. This is a global flag in Python (`ivy_compiler.py`). For now, can defer if not hitting golden test divergence.

---

### M6. ASSERT/REQUIRE/ENSURE actions missing unprovable→Sequence conversion

**Python** (`ivy_parser.py:2800-2801`):
```python
if p[1] and not check_unprovable.get():
    p[0] = Sequence()  # replace action with empty sequence
```

**Go** (`grammar_v17.y:3744-3808`): Never replaces with empty Sequence.

**Fix** — same as M5, depends on `check_unprovable` global.

---

### M7. make_mixin_name format difference (dot→underscore)

**Python** (`ivy_parser.py:2585-2591`):
```python
name = atom.rep.replace(iu.ivy_compose_character, '_') + '[' + suffix + str(label_counter) + ']'
```

**Go** (`grammar_v17.y:175-179`):
```go
name := fmt.Sprintf("%s[%s%d]", atom.Rep, suffix, lalrLabelCounter)
```

**Problem:** Python replaces `.` (ivy_compose_character) with `_` in the atom rep before composing. Go does not.

**Fix**:
```go
rep := strings.ReplaceAll(atom.Rep, ".", "_")
name := fmt.Sprintf("%s[%s%d]", rep, suffix, lalrLabelCounter)
```

---

### M8. scenariomixin missing infer_action_params

**Python** (`ivy_parser.py:2593-2604`):
```python
optargs, optreturns = infer_action_params(atom.rep, p[3], p[4])
```

**Go** (`grammar_v17.y:4247-4272`): Does not call `inferActionParams`. Uses raw `$3` and `$4`.

**Fix** — call `inferActionParams` before creating the ActionDef:
```go
formals, returns := inferActionParams(atypeStr, $3, $4)
```

The function exists at grammar_v17.y:222.

---

### M9. specimpl: GLOBAL and COMMON share single slot

**Python** (`ivy_parser.py:2357-2369`): GLOBAL sets `global_attribute = "global"`, COMMON sets `common_attribute = "common"`. These are separate globals.

**Go** (`grammar_v17.y:3395-3408`): Both GLOBAL and COMMON set `lex.specialAttribute`. Nested `global { common { ... } }` would lose one attribute.

**Fix** — add separate `globalAttribute` and `commonAttribute` fields to the lexer adapter, matching Python's separate globals.

---

### M10. instMod missing lineno parameter

**Python** (`ivy_parser.py:152`):
```python
def inst_mod(ivy, module, pref, subst, vsubst, modname=None, lineno=None):
    ...
    if lineno is not None:
        set_reference_lineno(lineno)
```

**Go** (`lalr_full/inst_mod.go:135`): No `lineno` parameter. The `set_reference_lineno` functionality is missing.

**Fix** — add `lineno ast.Location` parameter to `instMod` and implement `setReferenceLineno` if needed. Some callers of `inst_mod` in Python pass a lineno (e.g., `create_object` at line 720).

---

## ALSO CRITICAL TO FIX DISCREPANCIES

### L1. eqn rule uses Definition instead of Equals

**Python** (`ivy_parser.py:3280-3283`):
```python
p[0] = Equals(App(p[1]), App(p[3]))
```

**Go** (`grammar_v17.y:4149-4154`):
```go
$$ = ast.NewDefinition(ast.NewApp(...), ast.NewApp(...))
```

**Problem:** Wrong type. Need `Equals` in `ast` package (or check if it's `Atom("=", ...)` in the logic layer).

---

### L2. debugarg rule uses Definition instead of DebugItem

**Python** (`ivy_parser.py:3240-3246`):
```python
p[0] = DebugItem(lhs, p[3])
```

**Go** (`grammar_v17.y:3904-3909`):
```go
$$ = ast.NewDefinition(ast.NewApp(...), $3)
```

**Problem:** Wrong type. Need `DebugItem` in `ast` package. The `DebugItem` type should exist (it's used in the `debug` action).

---

### L3. doInsts missing object tracking

**Python** (`ivy_parser.py:246-247`):
```python
if pref is None:
    ivy.objects.update(module.objects)
```

**Go** (`lalr_full/inst_mod.go:122-123`): Comment says "object tracking -- deferred for now".

**Fix** — add `Objects` map to `ivyAccum` and merge during instantiation.

---

## MISSING AST TYPES SUMMARY

Types used in Python grammar but missing from Go `ast` package:

| Python Type | Used In | Go Status | Priority |
|---|---|---|---|
| `ChoiceAction` | `IF * seq ELSE act` (C7) | Only in `actions/` — need in `ast/` | Critical |
| `LetAction` | `LET eqns seq` (M2) | Only in `actions/` — need in `ast/` | Moderate |
| `Ranking` | `DECREASES fmla` (M3) | Only in `actions/` — need in `ast/` | Moderate |
| `MethodCall` | `methcall()` helper (C2) | Not found anywhere | Critical (for FOR) |
| `DebugItem` | `debugarg` rule (L2) | Not found in `ast/` | Low |
| `Equals` | `eqn` rule (L1) | Use `Atom("=",...)` | Low |
| `GhostTypeDef` | TYPE with ghost (M4) | Exists in `ast/decl.go` — just unused | Moderate |
| `DefinitionSchema` | DEFINITION explicit (C8) | Exists in `ast/formula.go` — just unused | Critical |

---

## RECOMMENDED FIX ORDER

Fix order based on what `make golden` will hit first (earlier trace lines = more impact):

1. **C1** — LOCAL action `loc:` prefix (likely hit early, affects scoping)
2. **C3** — WHILE `fixIfPart` (quick one-line fix)
3. **C7** — ChoiceAction type (need new type + grammar fix)
4. **C8** — DefinitionSchema for explicit (type exists, just wire it)
5. **C9** — TYPE Range + GhostTypeDef (moderate complexity)
6. **C6** — METHOD self parameter (moderate)
7. **C4/C5** — CLASS/SUBCLASS createObject (large, interconnected)
8. **C2** — FOR loop desugaring (largest single fix)
9. **M1-M10** — Moderate fixes as encountered in golden test
10. **L1-L3** — Critical to also fix now.

---

## FILES TO MODIFY

| File | Changes |
|---|---|
| `lalr_full/grammar_v17.y` | Fix rules C1-C9, M1-M3, M7-M8, L1-L2 |
| `ast/ast.go` | Add ChoiceAction, LetAction, Ranking, DebugItem types |
| `ast/decl.go` | Verify GhostTypeDef is complete |
| `ast/formula.go` | Verify DefinitionSchema is complete |
| `lalr_full/inst_mod.go` | M10 (lineno param), L3 (object tracking) |
