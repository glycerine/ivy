# TODO: Missing Port Logic at `get_lineno` Locations

**Created:** 2026-03-25
**Source:** Systematic comparison of Python `get_lineno(p,N)` calls vs Go `tokLineno`/`getLineno`/`nodeLineno` calls.
**Counts:** Python 331 calls, Go 254 calls. Gap: 86 missing trace calls across 75 rules.

Each missing `get_lineno` is a symptom — the surrounding logic may also be incomplete.

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

## THIRD GROUP — Lineno Only (Logic Otherwise OK)

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
