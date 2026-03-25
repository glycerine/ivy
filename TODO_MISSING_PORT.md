# TODO: Missing Port Logic at `get_lineno` Locations

**Created:** 2026-03-25
**Source:** Systematic comparison of Python `get_lineno(p,N)` calls vs Go `tokLineno`/`getLineno`/`nodeLineno` calls.
**Counts:** Python 331 calls, Go 254 calls. Gap: 86 missing trace calls across 75 rules.

Each missing `get_lineno` is a symptom — the surrounding logic may also be incomplete.
Entries are grouped by severity: CRITICAL (major logic gaps), MEDIUM (lineno + minor issues), LOW (lineno only), N/A (structural grammar differences).

---

## CRITICAL — Major Logic Gaps (ALL 14 FIXED 2026-03-25)

### 1. `p_top_assert_symbol_arrow_assert_rhs` — COMPLETELY EMPTY

**Python** (ivy_parser.py:2417-2423):
```python
def p_top_assert_symbol_arrow_assert_rhs(p):
    'top : top ASSERT SYMBOL ARROW assert_rhs'
    p[0] = p[1]
    thing = Implies(Atom(p[3],[]),p[5])
    thing.lineno = get_lineno(p,4)
    p[0].declare(AssertDecl(thing))
```

**Go** (grammar_v17.y:1644-1648):
```go
xtracer.Trace("parser.p_top_assert_symbol_arrow_assert_rhs ENTER (top)")
$$ = $1
// ALL LOGIC MISSING
```

**What needs to be ported:**
- Create `Implies` node from Atom($3) and $5
- Set lineno from ARROW token ($4)
- Create `AssertDecl` wrapping the Implies
- Call `$$.declare(d)`

---

### 2. `p_top_nativequote` — INCOMPLETE WITH TODO

**Python** (ivy_parser.py:2506-2515):
```python
def p_top_nativequote(p):
    'top : top NATIVEQUOTE'
    p[0] = p[1]
    text,bqs = parse_nativequote(p,2)
    defn = NativeDef(*([mk_label(None,'native')] + [text] + bqs))
    defn.lineno = get_lineno(p,2)
    thing = NativeDecl(defn)
    thing.lineno = get_lineno(p,2)
    p[0].declare(thing)
```

**Go** (grammar_v17.y:1597-1604):
```go
xtracer.Trace("parser.p_top_nativequote ENTER (top)")
$$ = $1
// TODO: Python creates NativeDef/NativeDecl with lineno here.
parseNativequote(acfg(v17lex), $2.Val, v17lex.(*v17LexAdapter))
```

**What needs to be ported:**
- Create `NativeDef` with label from `mkLabel(nil, "native")`, text, and backquotes
- Set lineno on NativeDef from $2 token
- Create `NativeDecl` wrapping the NativeDef
- Set lineno on NativeDecl from $2 token
- Call `$$.declare(thing)`

---

### 3. `p_lit_atom` — Missing Literal Wrapping

**Python** (ivy_logic_parser.py:391-395):
```python
def p_lit_atom(p):
    'lit : atom'
    p[0] = Literal(1,p[1])
    p[0].lineno = get_lineno(p,1)
```

**Go** (grammar_v17.y:3153-3157):
```go
xtracer.Trace("parser.p_lit_atom ENTER (lit)")
$$ = $1  // WRONG: returns raw atom, no Literal wrapping
```

**What needs to be ported:**
- Wrap atom in `Literal` with polarity +1: `Literal(1, atom)`
- Set lineno from $1
- Verify `Literal` type exists in Go AST (may need to be created)

---

### 4. `p_lit_term_eq_term` — Missing Literal Wrapping

**Python** (ivy_logic_parser.py:397-401):
```python
def p_lit_term_eq_term(p):
    'lit : SYMBOL EQ SYMBOL'
    p[0] = Literal(1,Atom(p[2],[symbol(p[1]),symbol(p[3])]))
    p[0].lineno = get_lineno(p,2)
```

**Go** (grammar_v17.y:3158-3162):
```go
xtracer.Trace("parser.p_lit_term_eq_term ENTER (lit)")
$$ = acfg(v17lex).NewAtom("=", acfg(v17lex).NewAtom($1.Val), acfg(v17lex).NewAtom($3.Val))
// MISSING: Literal(1, ...) wrapping and lineno
```

**What needs to be ported:**
- Wrap result in `Literal(1, ...)` with positive polarity
- Set lineno from EQ token ($2)

---

### 5. `p_lit_term_tildaeq_term` — Literal(0) vs ast.Not Mismatch

**Python** (ivy_logic_parser.py:403-407):
```python
def p_lit_term_tildaeq_term(p):
    'lit : SYMBOL TILDAEQ SYMBOL'
    p[0] = Literal(0,Atom(p[2],[symbol(p[1]),symbol(p[3])]))
    p[0].lineno = get_lineno(p,2)
```

**Go** (grammar_v17.y:3163-3167):
```go
xtracer.Trace("parser.p_lit_term_tildaeq_term ENTER (lit)")
$$ = &ast.Not{Body: acfg(v17lex).NewAtom("=", ...)}
// WRONG: uses ast.Not instead of Literal(0, ...)
```

**What needs to be ported:**
- Change from `ast.Not` to `Literal(0, ...)` — Python uses `Literal` with polarity 0 for negation
- Set lineno from TILDAEQ token ($2)
- Verify `Literal` type in Go supports polarity

---

### 6. `p_fmla_fmla_isa_atype` — Isa Construction Differs

**Python** (ivy_logic_parser.py:776-782):
```python
def p_fmla_fmla_isa_atype(p):
    'term : term ISA atype'
    tp = Atom(p[3],[])
    tp.lineno = get_lineno(p,2)
    p[0] = Isa(p[1],tp)
    p[0].lineno = get_lineno(p,2)
```

**Go** (grammar_v17.y:2130-2134):
```go
xtracer.Trace("parser.p_fmla_fmla_isa_atype ENTER (term)")
$$ = &ast.Isa{Terms: []ast.Node{$1, $3}}
// WRONG: Python wraps atype in Atom first; Go passes $3 directly
// MISSING: 2 lineno calls
```

**What needs to be ported:**
- Wrap `$3` (atype) in an Atom: `acfg(v17lex).NewAtom(atypeString)` with empty args
- Set lineno on the Atom from ISA token ($2)
- Create Isa with the term and the Atom wrapper
- Set lineno on Isa from ISA token ($2)

---

### 7. `p_term_namedbinder_vars_dot_term` — App vs Atom Type Mismatch

**Python** (ivy_logic_parser.py:755-761):
```python
def p_term_namedbinder_vars_dot_term(p):
    'term : LPAREN DOLLAR SYMBOL simplevars DOT fmla RPAREN LPAREN terms RPAREN'
    x = NamedBinder(p[3], p[4],p[6])
    x.lineno = get_lineno(p,2)
    p[0] = App(x, p[9])
    p[0].lineno = get_lineno(p,2)
```

**Go** (grammar_v17.y:2145-2150):
```go
xtracer.Trace("parser.p_term_namedbinder_vars_dot_term ENTER (term)")
binder := &ast.NamedBinder{Name: $3.Val, Bounds: $4, Body: $6}
$$ = &ast.Atom{Rep: "", Terms: append([]ast.Node{binder}, $9...)}
// WRONG: Returns Atom where Python returns App
// MISSING: 2 lineno calls
```

**What needs to be ported:**
- Change result from `ast.Atom` to `ast.App` (or `acfg.NewApp()`)
- Set lineno on NamedBinder from DOLLAR token ($2)
- Set lineno on the App from DOLLAR token ($2)

---

### 8. `p_term_term_colon_term` — Only Handles Variables

**Python** (ivy_logic_parser.py:219-225):
```python
def p_term_term_colon_term(p):
    'term : term COLON atype'
    if hasattr(p[1],"sort"):
        raise iu.IvyError(p[3],"multiple sort annotations")
    p[1].sort = p[3]
    p[0] = p[1]
```

**Go** (grammar_v17.y:2136-2143):
```go
xtracer.Trace("parser.p_term_term_colon_term ENTER (term)")
if v, ok := $1.(*ast.Variable); ok {
    v.VSort = atypeToString($3)
}
$$ = $1
// WRONG: Only handles Variables; Python handles ANY term with .sort attribute
// MISSING: error check for multiple sort annotations
// MISSING: sort annotation on non-Variable types (Atom.ASort, App.ASort, etc.)
```

**What needs to be ported:**
- Handle sort annotation on ALL term types, not just Variable
- Add error check: if sort already set, raise error
- For Atom: set ASort; for Variable: set VSort; for App: set ASort; etc.
- Note: Python stores atype object directly; Go converts to string — verify correctness

---

### 9. `p_objectend` — Missing State Mutation

**Python** (ivy_parser.py:703-707):
```python
def p_objectend(p):
    'objectend :'
    stack[-1].is_object=False
    p[0] = None
```

**Go** (grammar_v17.y:3245-3247):
```go
xtracer.Trace("parser.p_objectend ENTER (objectend)")
$$ = nil
// MISSING: equivalent of stack[-1].is_object = False
```

**What needs to be ported:**
- Set `is_object = false` on the current accumulator (equivalent of Python's `stack[-1]`)
- This affects object scope tracking during parsing

---

### 10. `p_proofstep_symbol` — Wrong Lineno Token

**Python** (ivy_parser.py:1239-1245):
```python
def p_proofstep_symbol(p):
    'proofstep : APPLY atype optrenaming'
    a = Atom(p[2])
    a.lineno = get_lineno(p,2)      # <-- position 2 (atype)
    p[0] = SchemaInstantiation(a,p[3])
    p[0].lineno = get_lineno(p,1)   # <-- position 1 (APPLY)
```

**Go** (grammar_v17.y:4825-4834):
```go
a := atypeToAtom(acfg(v17lex), $2)
a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))  // BUG: should be $2, not $1
si := &ast.SchemaInstantiation{SchemaName: a, Ren: $3}
si.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
```

**What needs to be ported:**
- Fix: `a.SetLineno(tokLineno(..., $1))` → `a.SetLineno(tokLineno(..., $2))` (atype position)
- Keep: `si.SetLineno(tokLineno(..., $1))` (APPLY position) — this is correct

---

### 11. `p_proofstep_symbol_with_defns` — Same Wrong Token Bug

**Python** (ivy_parser.py:1408-1414):
```python
a = Atom(p[2])
a.lineno = get_lineno(p,2)      # position 2
p[0] = SchemaInstantiation(...)
p[0].lineno = get_lineno(p,1)   # position 1
```

**Go** (grammar_v17.y:4835-4843):
```go
a.SetLineno(tokLineno(..., $1))  // BUG: should be $2
si.SetLineno(tokLineno(..., $1))
```

**What needs to be ported:**
- Same fix as #10: change atom lineno from $1 to $2

---

### 12. `p_simpleact_debug_symbol_optdebugargs` — Missing Validation + Lineno

**Python** (ivy_parser.py:3271-3279):
```python
def p_simpleact_debug_symbol_optdebugargs(p):
    'simpleact : DEBUG SYMBOL optdebugargs'
    action = Atom(p[2],[])
    action.lineno = get_lineno(p,2)
    if not p[2].startswith('"'):
        report_error(IvyError(action,"expected string constant after 'debug'"))
    p[0] = DebugAction(action,*p[3])
    p[0].lineno = get_lineno(p,1)
```

**Go** (grammar_v17.y:4059-4063):
```go
xtracer.Trace("parser.p_simpleact_debug_symbol_optdebugargs ENTER (simpleact)")
args := append([]ast.Node{acfg(v17lex).NewAtom($2.Val)}, $3...)
a := acfg(v17lex).NewDebugAction(args...)
$$ = a
// MISSING: lineno on action atom (from $2)
// MISSING: error check if $2.Val doesn't start with '"'
// MISSING: lineno on DebugAction (from $1)
```

**What needs to be ported:**
- Set lineno on the action Atom from $2 token
- Add error check: `if !strings.HasPrefix($2.Val, "\"")` → report error
- Set lineno on DebugAction from $1 token (DEBUG keyword)

---

### 13. `p_top_implement_type_symbol_with_symbol` — Incomplete

**Python** (ivy_parser.py:2209-2220):
```python
def p_top_implement_type_symbol_with_symbol(p):
    'top : top IMPLEMENT TYPE SYMBOL WITH SYMBOL'
    a1,a2 = Atom(p[4]),Atom(p[6])
    a1.lineno = get_lineno(p,4)
    a2.lineno = get_lineno(p,6)
    impl = ImplementTypeDef(a1,a2)
    impl.lineno = get_lineno(p,5)
    d = ImplementTypeDecl(mk_lf(impl))
    d.lineno = get_lineno(p,2)
    p[0] = p[1]
    p[0].declare(d)
```

**Go** (grammar_v17.y:1345-1353) — need to verify full extent:
```go
$$ = $1
a1 := acfg(v17lex).NewAtom($4.Val)
a2 := acfg(v17lex).NewAtom($6.Val)
impl := &ast.ImplementTypeDef{Elems: []ast.Node{a1, a2}}
// POSSIBLY INCOMPLETE — verify if mk_lf wrapping, ImplementTypeDecl, declare, and 4 lineno calls exist
```

**What needs to be ported:**
- Set lineno on a1 from $4 token
- Set lineno on a2 from $6 token
- Set lineno on impl from $5 token (WITH keyword)
- Wrap impl with `mkLF()` equivalent
- Create `ImplementTypeDecl` from the labeled formula
- Set lineno on the decl from $2 token (IMPLEMENT keyword)
- Call `$$.declare(d)`

---

### 14. `p_tacticwithelem_trigger` — Wrong Argument Type

**Python** (ivy_parser.py:1563-1567):
```python
def p_tacticwithelem_trigger(p):
    'tacticwithelem : TRIGGER atype WITH terms'
    p[0] = Trigger(*([Atom(p[2])]+p[4]))
    p[0].lineno = get_lineno(p,3)
```

**Go** (grammar_v17.y:4609-4611):
```go
xtracer.Trace("parser.p_tacticwithelem_trigger ENTER (tacticwithelem)")
$$ = &ast.Trigger{Terms: append([]ast.Node{$2}, $4...)}
// WRONG: Python wraps p[2] in Atom() first; Go passes $2 (atype node) directly
// MISSING: lineno from WITH token ($3)
```

**What needs to be ported:**
- Wrap `$2` in an Atom: `acfg(v17lex).NewAtom(atypeString)`
- Set lineno from WITH token ($3)
- Use `acfg().NewTrigger()` instead of struct init for config threading

---

## MEDIUM — Lineno + Minor Issues

### 15. `p_optlabel_label` — Missing Lineno (IMMEDIATE GOLDEN TEST FIX)

**Python** (ivy_parser.py:1651-1655):
```python
p[0] = Atom(p[1][1:-1],[])
p[0].lineno = get_lineno(p,1)
```

**Go** (grammar_v17.y:2296-2300):
```go
$$ = acfg(v17lex).NewAtom($1.Val)
// MISSING: $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
```

---

### 16. `p_optactiondef_eq_symbol` — Missing Lineno + Config Threading

**Python** (ivy_parser.py:2049-2053):
```python
p[0] = CrashAction()
p[0].lineno = get_lineno(p,2)
```

**Go** (grammar_v17.y:3461-3463):
```go
$$ = &ast.CrashAction{}
// MISSING: lineno from $2 (TIMES token)
// MINOR: should use acfg().NewCrashAction() for config threading
```

---

### 17. `p_renaming_lt_renaminglist_gt` — Missing Lineno

**Python** (ivy_parser.py:1397-1401):
```python
p[0] = Renaming(*p[2])
p[0].lineno = get_lineno(p,1)
```

**Go** (grammar_v17.y:4775-4778):
```go
r := &ast.Renaming{Elems: $2}
$$ = r
// MISSING: r.SetLineno(tokLineno(..., $1))
```

---

### 18. `p_renamingitem_symbol_div_symbol` — Missing Lineno

**Python** (ivy_parser.py:1375-1379):
```python
p[0] = Definition(Atom(p[3],[]),Atom(p[1],[]))
p[0].lineno = get_lineno(p,2)
```

**Go** (grammar_v17.y:4741-4743):
```go
$$ = acfg(v17lex).NewDefinition(acfg(v17lex).NewAtom($3.Val), acfg(v17lex).NewAtom($1.Val))
// MISSING: $$.SetLineno(tokLineno(..., $2))
```

---

### 19. `p_renamingitem_variable_div_variable` — Missing Lineno

**Python** (ivy_parser.py:1369-1373):
```python
p[0] = Definition(Variable(p[3],universe),Variable(p[1],universe))
p[0].lineno = get_lineno(p,2)
```

**Go** (grammar_v17.y:4735-4738):
```go
$$ = acfg(v17lex).NewDefinition(&ast.Variable{Rep: $3.Val, VSort: "S"}, &ast.Variable{Rep: $1.Val, VSort: "S"})
// MISSING: $$.SetLineno(tokLineno(..., $2))
// MINOR: verify "S" matches Python's `universe` value
```

---

### 20. `p_term_namedbinder_dot_fmla` — Missing Lineno + Bounds

**Python** (ivy_logic_parser.py:763-767):
```python
p[0] = NamedBinder(p[2], [],p[4])
p[0].lineno = get_lineno(p,1)
```

**Go** (grammar_v17.y:2151-2155):
```go
$$ = &ast.NamedBinder{Name: $2.Val, Body: $4}
// MISSING: Bounds: nil (or empty slice) — Python passes []
// MISSING: lineno from DOLLAR token ($1)
```

---

### 21. `p_term_namedbinder_dollar_fmla` — Missing Lineno + Bounds

**Python** (ivy_logic_parser.py:769-773):
```python
p[0] = NamedBinder(p[2], [],p[4])
p[0].lineno = get_lineno(p,1)
```

**Go** (grammar_v17.y:2156-2160):
```go
$$ = &ast.NamedBinder{Name: $2.Val, Body: $4}
// MISSING: Bounds: nil and lineno from $1
```

---

### 22. `p_top_around_callatom_lcb_action_rcb` — Missing Atom Lineno

**Python** (ivy_parser.py:2184-2193):
```python
atom = Atom(p[3])
atom.lineno = get_lineno(p,2)  # <-- this trace is missing in Go
handle_before_after("before",atom,before,p[0],p[4],p[5])
handle_before_after("after",atom,after,p[0],p[4],p[5])
```

**Go** (grammar_v17.y:1300-1322):
```go
atom := acfg(v17lex).NewAtom($3.(*ast.Symbol).Rep)
// MISSING: atom.SetLineno(tokLineno(..., $2))
// Note: Go inlines the before/after logic instead of calling handleBeforeAfter
```

---

### 23. `p_top_delegate_callatom_opt` — Missing Lineno

**Python** (ivy_parser.py:2325-2334):
```python
d.lineno = get_lineno(p,2)
p[0] = p[1]
p[0].declare(d)
```

**Go** (grammar_v17.y:1477-1488):
```go
dd := acfg(v17lex).NewDelegateDecl(args...)
$$.declare(dd)
// MISSING: dd.SetLineno(tokLineno(..., $2))
```

---

### 24. `p_top_unprovable_invariant_labeledfmla` — Missing Lineno

**Python** (ivy_parser.py:615-627):
```python
d = ConjectureDecl(lf)
d.lineno = get_lineno(p,3)
```

**Go** (grammar_v17.y:829-841):
```go
d := acfg(v17lex).NewConjectureDecl(lf)
// MISSING: d.SetLineno(tokLineno(..., $3))
```

---

### 25. `p_top_object_symbol_eq_lcb_top_rcb` — nodeLineno vs tokLineno

**Python** (ivy_parser.py:733-737):
```python
create_object(p[0],p[3],p[4],p[8],get_lineno(p,3),p[7])
```

**Go** (grammar_v17.y:877-884):
```go
createObject(acfg(v17lex), $$, pref, $4, objAccum, nodeLineno($3), $7)
// nodeLineno($3) matches get_lineno(p,3) — OK for trace matching
// But verify nodeLineno returns equivalent Location
```

---

### 26. `p_top_specification_lcb_top_rcb` — nodeLineno used (trace matches)

Go uses nodeLineno internally in attribute handling. Logic is semantically equivalent. No missing lineno trace.

---

### 27. `p_app_symbol` — Missing Lineno

**Python** (ivy_logic_parser.py:334-338):
```python
p[0] = App(p[1],[])
p[0].lineno = get_lineno(p,1)
```

**Go** (grammar_v17.y:3121-3126):
```go
$$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1.Val, nil))
// MISSING: $$.SetLineno(tokLineno(..., $1))
```

---

### 28. `p_app_symbol_lp_terms_rp` — Missing Lineno

**Python** (ivy_logic_parser.py:340-344):
```python
p[0] = App(p[1],p[3])
p[0].lineno = get_lineno(p,1)
```

**Go** (grammar_v17.y:3127-3131):
```go
$$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1.Val, nil), $3...)
// MISSING: $$.SetLineno(tokLineno(..., $1))
```

---

### 29. `p_app_term_infix_term` — Missing Lineno

**Python** (ivy_logic_parser.py:346-350):
```python
p[0] = App(p[2],p[1],p[3])
p[0].lineno = get_lineno(p,2)
```

**Go** (grammar_v17.y:3132-3136):
```go
$$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($2, nil), $1, $3)
// MISSING: $$.SetLineno — need lineno from infix operator token ($2)
```

---

### 30. `p_aterm_aterm_dot_appelem` — Missing Lineno

**Python** (ivy_logic_parser.py:212-216):
```python
p[0] = compose_atoms(p[1],p[3])
p[0].lineno = get_lineno(p,2)
```

**Go** (grammar_v17.y:1750-1754):
```go
$$ = ast.ComposeAtoms($1.(*ast.Atom), $3.(*ast.Atom))
// MISSING: $$.SetLineno(tokLineno(..., $2)) — DOT token
```

---

### 31. `p_lparam_caret_variable_colon_symbol` — Missing Lineno

**Python** (ivy_parser.py — has get_lineno):
```python
# Creates a variable with caret prefix and sort annotation
p[0].lineno = get_lineno(p, ...)
```

**Go** — exists with trace but no lineno call.

---

### 32. `p_targs_lparen_tsyms_rparen` — Missing Lineno

**Python** (ivy_logic_parser.py):
```python
p[0].lineno = get_lineno(p,1)
```

**Go** — exists with trace but no lineno call.

---

### 33. `p_lit_tilda_atom` — Missing Lineno

**Python** (ivy_logic_parser.py:409-413):
```python
p[0] = ~p[2]
p[0].lineno = get_lineno(p,1)
```

**Go** (grammar_v17.y:3168-3172):
```go
$$ = &ast.Not{Body: $2}
// MISSING: $$.SetLineno(tokLineno(..., $1)) — TILDA token
// Note: Python uses ~ operator (Literal.__invert__) vs Go's ast.Not
```

---

### 34. `p_inst_modinst` — Missing Lineno

**Python** has `get_lineno` call. Go exists but no lineno set.

---

### 35. `p_callatom_method` — Missing Lineno

**Python** has `get_lineno` call. Go exists but no lineno set.

---

### 36. `p_callatom_callatom_dot_callatom` — nodeLineno used (trace matches)

Go uses `nodeLineno($1)` which produces the trace. No missing trace.

---

### 37. `p_proofstep_theorem` — nodeLineno used (trace matches)

Go uses `nodeLineno($2)`. No missing trace.

---

## LOW — Lineno Only (Logic Otherwise OK)

### 38-47. Proofstep rules with nodeLineno pattern

These rules use `nodeLineno()` for child node lineno and `tokLineno()` for the tactic lineno. Python uses `get_lineno` for both. The nodeLineno produces the trace, so these are NOT missing traces — but the lineno source differs slightly (node's stored lineno vs token position):

- `p_proofstep_assume` — 2 py, 1 tokLineno + 1 nodeLineno
- `p_proofstep_assume_with_defns` — same pattern
- `p_proofstep_instantiate` — same pattern
- `p_proofstep_instantiate_with_defns` — same pattern
- `p_proofstep_spoil_atype` — same pattern
- `p_proofstep_unfold_atype_with_defns` — same pattern
- `p_scenariomixin_before_callatom_lcb_action_rcb` — same pattern
- `p_scenariomixin_after_callatom_lcb_action_rcb` — same pattern
- `p_top_variant_symbol_of_atype` — same pattern
- `p_top_variant_symbol_of_symbol_eq_sort` — same pattern

### 48-57. Rules in "fewer calls" category that actually match

After including nodeLineno, these rules have equivalent trace counts:
- `p_top__top_class_symbol_objectargs_eq_lcb_optdo` — nodeLineno covers the gap
- `p_top__top_subclass_symbol_of_atype_eq_lcb_optd` — same
- `p_top_after_callatom_lcb_action_rcb` — same
- `p_top_type_symbol_eq_sort` — same
- `p_top_opttrusted_extract_callatom_eq_lcb_top_rcb_optwith` — same
- `p_defn_atom_fmla` — matches (1 py, 1 go)
- `p_action_require_proof_proofstep` — matches (Go has 2)
- `p_action_call_callatom` — matches (1 py, 1 go)
- `p_top_interpret_symbol_arrow_lcb_symbol_moresymbols_rcb` — matches (2 py, 2 go)
- `p_top_axiom_optlabel_gprop` — matches (1 py, 1 go)

### 58-60. Remaining fewer-calls

- `p_proofstep_witness_pflets` — py=2 go=1 — needs investigation
- `p_term_if_fmla_else_term` — py=3 go=1 — Python has nested-if warning check, not a lineno issue
- `p_top_optimpex_action_symbol_optargs_optreturns_eq_action` — py=3 go=1 — Go consolidates into lineno var

---

## NOT APPLICABLE — Structural Grammar Differences

These Python `p_fmla_*` rules don't exist in Go because Go unified `fmla` into `term`. The Go `p_term_*` equivalents already have lineno calls. No action needed:

- `p_fmla_true` → `p_term_true` (already has tokLineno)
- `p_fmla_false` → `p_term_false` (already has tokLineno)
- `p_fmla_not_fmla` → `p_term_not_term` (already has tokLineno)
- `p_fmla_fmla_and_fmla` → `p_term_term_and_term` (already has tokLineno)
- `p_fmla_fmla_or_fmla` → `p_term_term_or_term` (already has tokLineno)
- `p_fmla_fmla_arrow_fmla` → `p_term_term_arrow_term` (already has tokLineno)
- `p_fmla_fmla_iff_fmla` → `p_term_term_iff_term` (already has tokLineno)
- `p_fmla_forall_simplevars_dot_fmla` → `p_term_forall_simplevars_dot_term` (already has tokLineno)
- `p_fmla_exists_simplevars_dot_fmla` → `p_term_exists_simplevars_dot_term` (already has tokLineno)
- `p_fmla_forall_lp_vars_lp_fmla` → `p_term_forall_lp_vars_lp_term` (already has tokLineno)
- `p_fmla_exists_lp_vars_lp_fmla` → `p_term_exists_lp_vars_lp_term` (already has tokLineno)
- `p_fmla_globally_fmla` → `p_term_globally_term` (already has tokLineno)
- `p_fmla_eventually_fmla` → `p_term_eventually_term` (already has tokLineno)
- `p_fmla_term_relop_term` → handled by `p_term_term_EQ_term` etc.
- `p_fmla_term_tildaeq_term` → `p_term_term_tildaeq_term` (already has tokLineno)
- `p_fmla_exists_vars_dot_fmla` → no direct equivalent (uses simplevars variant)
- `p_fmla_forall_vars_dot_fmla` → no direct equivalent (uses simplevars variant)

## UNPORTED RULES — No Go Grammar Rule Exists

These Python rules have no corresponding Go grammar rule at all (not just missing lineno):

- `p_action_ensures` — action 'ensures' clause
- `p_action_field_assign_false` — field assignment = false
- `p_action_field_assign_field` — field assignment = field
- `p_action_field_assign_null` — field assignment = null
- `p_action_field_assign_term` — field assignment = term
- `p_action_if_fmla_lcb_action_rcb` — if-action with fmla (Go uses somefmla)
- `p_action_if_fmla_lcb_action_rcb_else_LCB_action_RCB` — if-else with fmla
- `p_aterm_old_symbol` — old(symbol) in aterm context
- `p_aterm_symbol` — bare symbol as aterm
- `p_callatom_callatom_colon_callatom` — callatom:callatom syntax
- `p_opttypedsym_symbol` — optional typed symbol
- `p_opttypedsym_symbol_colon_atype` — optional typed symbol with sort
- `p_term_term_dot_aterm` — term.aterm
- `p_term_term_dot_term` — term.term
- `p_top_init_fmla` — init formula declaration
- `p_top_private_callatom` — private declaration
