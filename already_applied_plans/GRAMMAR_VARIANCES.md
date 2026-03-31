# Grammar Variances: Python (source of truth) vs Go (current port)

**Date:** 2026-03-23
**Python source:** `~/goivy/IVY_PYTHON_GRAMMAR.md`
**Go source:** `~/goivy/lalr_full/WIP_GO_GRAMMAR_CURRENT.md`
**Scope:** v1.7+ rules only (our target version)

Items marked with severity:
- **[CRITICAL]** — affects parsing correctness, will cause parse failures or wrong AST
- **[MODERATE]** — structural difference that may affect semantics
- **[LOW]** — cosmetic, naming, or unused-feature difference

---

## 1. PRECEDENCE TABLE VARIANCES

### V-PREC-1: Go adds TOK_ISA to relop precedence level [LOW]
- **Python:** ISA is NOT in the precedence table at all
- **Go:** `%left TOK_EQ TOK_LE TOK_LT TOK_GE TOK_GT TOK_PTO TOK_ISA` (line 483)
- **Impact:** In Python, ISA has no explicit precedence and defaults to lowest. Go gives it
  explicit left-associative precedence at the comparison level. This changes parsing of
  expressions like `x ISA t & y` — Python parses as `(x ISA t) & y` via grammar structure,
  Go does the same via precedence. Likely benign but worth noting.

### V-PREC-2: Go adds %right TOK_ASSIGN [LOW]
- **Python:** ASSIGN (`:=`) is NOT in the precedence table
- **Go:** `%right TOK_ASSIGN` (line 493)
- **Impact:** ASSIGN appears in action rules (`term ASSIGN fmla`), not in expression
  rules, so this mainly helps goyacc resolve shift/reduce conflicts. Benign.

---

## 2. TOKEN VARIANCES

### V-TOK-1: Go declares TOK_LABEL but lexer never returns it [LOW]
- **Python:** Lexer returns LABEL token for `[identifier]` patterns
- **Go:** Declares `TOK_LABEL` (line 283) but the Lex() adapter never returns it.
  Instead, `labelname` rule handles `TOK_LB SYMBOLx TOK_RB` at grammar level.
- **Impact:** Functionally equivalent. The Python LABEL token is `LB SYMBOL RB`
  concatenated by the lexer; Go splits it into three tokens.

### V-TOK-2: Go has unused placeholder tokens [LOW]
- **Go** declares but never uses: `TOK_VAR_KW`, `TOK_METHOD_KW`, `TOK_NULL_KW`, `TOK_SET_KW`
- **Impact:** None. Cleanup only.

---

## 3. SYMBOL/SUBSCRIPT RULES — MATCH

Both grammars have identical SYMBOL/subscript structure:
```
Python: SYMBOL : PRESYMBOL | SYMBOL LB SYMsubscr RB
Go:     SYMBOLx : TOK_PRESYMBOL | SYMBOLx TOK_LB SYMsubscr TOK_RB
```
SYMsubscr rules also match. LABEL rule matches (different mechanism, same effect).

Naming difference: Python uses `SYMBOL`, Go uses `SYMBOLx`. This is intentional
to avoid collision with the Go lexer's `SYMBOL` token type constant.

---

## 4. atype RULES — MATCH

Both have: `atype : SYMBOL | atype DOT SYMBOL | THIS`. Match.

---

## 5. TERM/FORMULA RULES

### V-TERM-1: Python has aterm, Go does not [MODERATE]
- **Python v1.7+** defines:
  ```
  aterm : appelem                       # line 193
  aterm : aterm DOT appelem             # line 198
  ```
  Used by `exprterm : aterm` (line 3428).
- **Go** has no `aterm` rule. Go's `exprterm` uses `appelem` directly (line 4203).
- **Impact:** Python's `exprterm` can parse dotted expressions like `foo.bar` via
  `aterm : aterm DOT appelem`. Go's `exprterm` only accepts a single `appelem`.
  This means Go cannot parse dotted concept space expressions. May affect concept
  space parsing correctness.

### V-TERM-2: term rules — MATCH
All 38+ term alternatives match between Python and Go:
- appelem, var, OLD appelem, term DOT appelem, LPAREN term RPAREN
- Arithmetic: PLUS, MINUS, TIMES, DIV
- Conditional: term IF fmla ELSE term
- Relational: EQ, LE, LT, GE, GT, PTO, TILDAEQ
- Boolean: TRUE, FALSE, TILDA, AND, OR, ARROW, IFF
- Quantifiers: FORALL/EXISTS with simplevars DOT or LPAREN vars RPAREN
- Temporal: GLOBALLY, EVENTUALLY, WHENNEXT, WHENPREV, WHENFIRST, WHENLAST
- ISA, COLON atype (sort annotation)
- Named binders: all 3 forms match

### V-TERM-3: fmla — MATCH
Both: `fmla : term`. Match.

---

## 6. lit RULES

### V-LIT-1: Go missing two lit alternatives [MODERATE]
- **Python:**
  ```
  lit : atom                             # line 377
  lit : SYMBOL EQ SYMBOL                 # line 383
  lit : SYMBOL TILDAEQ SYMBOL            # line 389
  lit : TILDA lit                        # line 395
  ```
- **Go:**
  ```
  lit : atom                             # line 2620
  lit : TILDA lit                        # line 2628
  ```
- **Missing in Go:** `lit : SYMBOL EQ SYMBOL` and `lit : SYMBOL TILDAEQ SYMBOL`
- **Impact:** `lit` is used in `simpleact : SET lit` and concept space parsing.
  The missing alternatives prevent parsing `SET x = y` and `SET x ~= y` in actions.
  However, `SET` is only a v1.0-1.1 keyword (removed in v>1.1), so this is LOW
  impact for v1.7+ unless legacy files are processed.

---

## 7. TOP-LEVEL DECLARATION RULES

### V-TOP-1: All v1.7+ top rules — MATCH
Systematic check of all `top : top ...` alternatives:
- USING, INCLUDE — match
- optexplicit opttemporal AXIOM lgprop — match
- optexplicit opttemporal PROPERTY labeledfmla optskolem optproof — match
- CONJECTURE labeledfmla — match
- optexplicit INVARIANT labeledfmla optproof — match
- UNPROVABLE INVARIANT labeledfmla optproof — match
- MODULE modulestart modcat atom optwith EQ LCB top RCB moduleend — match
- OBJECT objsym objectargs EQ LCB optdotdotdot top RCB objectend — match
- CLASS objsym objectargs EQ LCB optdotdotdot top RCB objectend — match
- SUBCLASS objsym OF atype EQ LCB optdotdotdot top RCB objectend — match
- optexplicit DEFINITION optlabel gdefn optproof — match
- SCHEMA schdefn — match
- THEOREM schdefn optproof — match
- THEOREM LABEL/labelname schdefnrhs optproof — match
- PROOF LABEL/labelname proofstep — match
- INSTANTIATE insts — match
- AUTOINSTANCE insts — match
- symdecl — match
- RELATION rels — match
- FUNCTION funs — match
- DERIVED defns — match
- optfinite optghost TYPE typesymbol (with/without EQ sort) — match
- PROGRESS defns — match
- RELY atom ARROW atom / RELY atom — match
- MIXORD callatom ARROW callatom — match
- CONCEPT cdefns — match
- UPDATE apps FROM apps upaxes — match
- MACRO atom EQ sequence — match
- optimpex actmeth SYMBOL optargs optreturns optactiondef — match
- MIXIN callatom BEFORE/AFTER callatom — match
- BEFORE atype optargs optreturns sequence — match
- AFTER atype optargs optreturns topseq — match
- AROUND atype optargs optreturns LCB actseq optsemi DOTDOTDOT actseq optsemi RCB — match
- AFTER INIT optargs topseq — match
- IMPLEMENT atype optargs optreturns topseq — match
- IMPLEMENT TYPE SYMBOL WITH SYMBOL — match
- opttrusted ISOLATE SYMBOL optargs EQ callatoms (with/without WITH callatoms) — match
- opttrusted ISOLATE SYMBOL optargs EQ LCB top RCB optwith — match
- EXTRACT objsym objectargs EQ LCB top RCB optwith — match
- EXTRACT objsym objectargs EQ callatoms — match
- EXPORT callatom — match
- IMPORT callatom — match
- DELEGATE callatoms optdelegee — match
- INTERPRET oper ARROW oper — match
- INTERPRET oper ARROW LCB term DOTS term RCB — match
- INTERPRET oper ARROW LCB SYMBOL moresymbols RCB — match
- ALIAS SYMBOL EQ callatom — match
- ATTRIBUTE callatom EQ attributeval — match
- VARIANT typesymbol OF atype (with/without EQ sort) — match
- NATIVEQUOTE — match
- SCENARIO LCB sceninit SEMI scentranss RCB — match
- specimpl LCB top RCB — match
- STATE SYMBOL EQ state_expr — match
- ASSERT SYMBOL ARROW assert_rhs — match (v1.6 compat, both have it)

All top-level declaration rules MATCH.

---

## 8. HELPER NONTERMINAL RULES

### labeledfmla, labelname, gprop, lgprop — MATCH

### V-OPT-1: optproofgroup missing PROOF keyword [CRITICAL]
- **Python:**
  ```
  optproofgroup : /* empty */              # line 1600
  optproofgroup : PROOF proofgroup         # line 1605
  ```
- **Go:**
  ```
  optproofgroup : /* empty */              # line 3784
  optproofgroup : proofgroup               # line 3790
  ```
- **Missing in Go:** The `PROOF` keyword before `proofgroup` in the non-empty alternative.
- **Impact:** Go will consume a `{ ... }` block as a proofgroup even without the `PROOF`
  keyword preceding it. This could cause incorrect parsing of declarations that use
  `optproofgroup` — e.g., `proofstep : TACTIC atype opttacticwith optproofgroup` would
  greedily consume a subsequent `{ ... }` block that isn't actually a proof.

### opttemporal, optunprovable, optexplicit, optlabel, optskolem — MATCH
### optproof — MATCH (both have empty / PROOF proofstep / PROOF LABEL proofstep)
### optsemi — MATCH

---

## 9. DEFINITION RULES

### dotsym, defnlhs, defarg, defargs, typeddefn — MATCH
### defnrhs, defn, defns, gdefn — MATCH
### somevarfmla, optin, optelse — MATCH

---

## 10. SCHEMA RULES

### schdefnrhs, schdecl, schdecls, schconc, schdefn — MATCH
All v1.7+ schema rules match exactly.

---

## 11. SYMBOL DECLARATION RULES

### symdecl, constantdecl, parameter, paramval — MATCH
All match including v1.7+ CONSTRUCTOR alternative.

---

## 12. TAPP/TTERM/TTERMS, TARGS, TSYMS — MATCH

---

## 13. TATOM/TATOMS — MATCH

---

## 14. REL/RELS, FUN/FUNS — MATCH

---

## 15. TYPE RULES

### typesymbol, optfinite, optghost, sort, names — MATCH

---

## 16. RELOP/INFIX — MATCH

---

## 17. ATOM/ATOMS, APP/APPS — MATCH

---

## 18. CALLATOM/CALLATOMS — MATCH
Both have: atom, THIS, METHOD, callatom DOT callatom.

---

## 19. MODULE/OBJECT INFRASTRUCTURE — MATCH
modulestart, moduleend, objectend, modcat, opteq, optdotdotdot, objectargs,
objsym, opttrusted, optargs, optreturns, optactualreturns, param, params,
optwith — all match.

---

## 20. LPARAM/LPARAMS — MATCH
Both have SYMBOL COLON atype and CARET SYMBOL COLON atype.

---

## 21. ACTION DEFINITION RULES — MATCH
optactiondef, topseq, optimpex, actmeth — all match.

---

## 22. SPECIMPL — MATCH

---

## 23. INSTANTIATE RULES — MATCH
insts, inst (modinst / modinst COLON modinst), modinst — all match.

---

## 24. PNAME/PNAMES — MATCH
All 7 pname alternatives match (atype, var, infix, relop, THIS, TRUE, FALSE).
pnames (empty / pname / pnames COMMA pname) — match.

---

## 25. OPER/ATTRIBUTEVAL/MORESYMBOLS/OPTDELEGEE — MATCH

---

## 26. SEQUENCE/ACTSEQ/ACTSEQREV

### actseqrev — MATCH
All 7 alternatives match exactly between Python and Go.

### actseq — MATCH (reverse of actseqrev)

### sequence — MATCH
Three alternatives: LCB RCB, LCB actseq RCB, LCB actseq SEMI RCB.

### action — MATCH
Both: simpleact | complexact.

---

## 27. SIMPLE ACTIONS

### V-ACT-1: Python VAR uses optinit, Go splits into two rules [LOW]
- **Python:**
  ```
  simpleact : VAR tterm optinit              # line 3183
  optinit   : /* empty */                    # line 3173
  optinit   : ASSIGN fmla                   # line 3178
  ```
- **Go:**
  ```
  simpleact : TOK_VAR tterm                   # line 3345
  simpleact : TOK_VAR tterm TOK_ASSIGN fmla   # line 3350
  ```
- **Impact:** Functionally equivalent. Go inlines the `optinit` alternatives.
  No `optinit` nonterminal needed.

### V-ACT-2: Go has UNPROVABLE simpleact standalone rule [LOW]
- **Go:** `simpleact : TOK_UNPROVABLE simpleact` (line 3376) — produces no-op Sequence
- **Python:** No standalone `UNPROVABLE simpleact` production. Python handles unprovable
  via `optunprovable` prefix on ASSERT/ENSURE/REQUIRE.
- **Impact:** Go accepts `unprovable <any simpleact>` while Python only accepts
  `unprovable assert/ensure/require`. Go is more permissive but silently discards.

### All other simpleact alternatives — MATCH:
- ASSUME labeledfmla
- optunprovable ASSERT labeledfmla (with/without PROOF proofstep)
- optunprovable REQUIRE labeledfmla (with/without PROOF proofstep)
- optunprovable ENSURE labeledfmla (with/without PROOF proofstep)
- term ASSIGN fmla
- termtuple ASSIGN callatom
- term ASSIGN TIMES
- CALL optactualreturns callatom / CALL callatom
- SET lit
- INSTANTIATE callatom
- DEBUG SYMBOL optdebugargs
- term (bare call, both use %prec SEMI)

---

## 28. COMPLEX ACTIONS

### All complexact alternatives — MATCH:
- sequence
- IF somefmla sequence (with/without ELSE action)
- IF TIMES sequence ELSE action
- WHILE somefmla invariants decreases sequence
- FOR tterm COMMA tterm IN fmla invariants decreases sequence
- LOCAL lparams sequence
- LET eqns sequence
- THUNK LABEL SYMBOL optargs COLON atype ASSIGN sequence

---

## 29. SOMEFMLA/BOUNDS — MATCH
All 5 somefmla alternatives and 2 bounds alternatives match.

---

## 30. INVARIANTS/DECREASES — MATCH

---

## 31. EQN/EQNS — MATCH

---

## 32. TERMTUPLE — MATCH

---

## 33. DEBUGARG/DEBUGARGS/OPTDEBUGARGS — MATCH

---

## 34. SCENARIO RULES — MATCH
sceninit, places, scentranss, scentrans, scenariomixin — all match.
Go uses a separate `scentrans` nonterminal while Python inlines into `scentranss`,
but the structure is equivalent.

---

## 35. PROOF/TACTIC RULES

### V-PROOF-1: proofseq has extra alternative in Go [MODERATE]
- **Python:**
  ```
  proofseq : proofstep                       # line 1555
  proofseq : proofseq optsemi proofstep      # line 1560
  ```
- **Go:**
  ```
  proofseq : proofstep                       # line 3797
  proofseq : proofseq TOK_SEMI proofstep     # line 3803
  proofseq : proofseq proofstep              # line 3808  ← EXTRA
  ```
- **Impact:** Go allows proof steps without any separator (not even optional semicolon).
  Python requires `optsemi` (which can be empty, but is a distinct rule causing a different
  reduction). This makes Go more permissive. May cause different parse trees for ambiguous
  proof sequences.

### V-PROOF-2: optproofgroup missing PROOF keyword (see V-OPT-1 above) [CRITICAL]

### V-PROOF-3: TACTIC uses atype vs SYMBOL [MODERATE]
- **Python:** `proofstep : TACTIC SYMBOL opttacticwith optproofgroup` (line 1265)
- **Go:** `proofstep : TOK_TACTIC atype opttacticwith optproofgroup` (line 3949)
- **Impact:** Python expects a bare SYMBOL (single identifier). Go accepts `atype`
  which includes dotted names (`foo.bar`) and `THIS`. More permissive than Python.

### V-PROOF-4: tacticwithelem DEFINITION uses atype vs typeddefn [MODERATE]
- **Python:** `tacticwithelem : DEFINITION typeddefn EQ fmla` (line 1525)
- **Go:** `tacticwithelem : TOK_DEFINITION atype TOK_EQ fmla` (line 3715)
- **Impact:** Python's `typeddefn` allows `defnlhs` (with optional args and sort), while
  Go's `atype` is just a type name. Go cannot parse `definition f(x:t) = ...` in
  tactic-with blocks.

### V-PROOF-5: INSTANTIATE LABEL ordering [LOW]
- **Python:** `proofstep : INSTANTIATE LABEL atype optrenaming` (line 1234)
- **Go:** `proofstep : TOK_INSTANTIATE labelname atype optrenaming` (line 3919)
- **Impact:** Functionally equivalent. Go uses `labelname` (which handles `[sym]` and `LABEL`),
  Python uses bare `LABEL` token.

### V-PROOF-6: INSTANTIATE LABEL ... WITH matches [LOW]
- **Python:** `proofstep : INSTANTIATE LABEL atype optrenaming WITH matches` (line 1403)
- **Go:** Not present as a separate production.
- **Impact:** Go's `TOK_INSTANTIATE labelname atype optrenaming` doesn't have a WITH variant
  with labelname. However, `TOK_INSTANTIATE atype optrenaming WITH matches` (without label)
  exists. Missing the labeled+WITH variant.

### V-PROOF-7: UNFOLD uses callatoms vs unfspecs [CRITICAL]
- **Python:**
  ```
  unfspec    : callatom renamings            # line 1425
  unfspecs   : unfspec | unfspecs COMMA unfspec  # lines 1431, 1436
  proofstep  : UNFOLD atype WITH unfspecs    # line 1442
  proofstep  : UNFOLD WITH unfspecs          # line 1451
  ```
- **Go:**
  ```
  proofstep  : TOK_UNFOLD atype TOK_WITH callatoms   # line 3993
  proofstep  : TOK_UNFOLD TOK_WITH callatoms          # line 4000
  ```
- **Missing in Go:** The `unfspec`, `unfspecs`, and `renamings` nonterminals. Python's
  UNFOLD takes unfspecs (which are callatoms with optional renamings), Go just takes
  callatoms (no renamings support).
- **Also missing in Go:** `renamings : /* empty */ | renamings renaming` (lines 1414, 1419)

### pflet, pflets, tacticwithelem (INVARIANT/TRIGGER), tacticwithlist — MATCH
### tacticwithlistchoice, opttacticwith (3 alternatives) — MATCH
### proofgroup, proofstep (APPLY/ASSUME/INSTANTIATE/etc.) — MATCH (except above)
### match, matches, renamingitem, renaminglist, optrenaming, renaming — MATCH

---

## 36. UPDATE/STATE/CONCEPT RULES — MATCH
requires, ensures, modifies, upaxes, upax, assert_rhs, state_expr,
cdefn, cdefns, expr, exprterm (except V-TERM-1), prod, sum — match.

### V-EXPR-1: exprterm uses aterm vs appelem (see V-TERM-1) [MODERATE]

---

## 37. LOC/SYMBOLS — MATCH
Both have `loc : /* empty */ | SYMBOL` and `symbols : SYMBOL | symbols COMMA SYMBOL`.

---

## 38. HELPER FUNCTION VARIANCES

### V-FUNC-1: lowerVarStmts is a stub [CRITICAL]
- **Python:** Full implementation (ivy_parser.py:2670-2697) that transforms `VarAction`
  into `LocalAction` with proper variable scoping and renaming.
- **Go:** Stub at line 4277 — `return stmts` unchanged.
- **Impact:** Variable declarations inside action sequences won't get proper scoping.
  `var x : t; x := 1; assert x = 1` won't have `x` renamed to `loc:x` and wrapped
  in a `LocalAction`.

---

## SUMMARY TABLE

| ID | Variance | Severity | Category |
|----|----------|----------|----------|
| V-PREC-1 | Go adds TOK_ISA to precedence | LOW | Precedence |
| V-PREC-2 | Go adds %right TOK_ASSIGN | LOW | Precedence |
| V-TOK-1 | TOK_LABEL declared but unused by lexer | LOW | Token |
| V-TOK-2 | Unused placeholder tokens in Go | LOW | Token |
| V-TERM-1 | Python has aterm rule, Go does not (affects exprterm) | MODERATE | Grammar |
| V-LIT-1 | Go missing lit : SYMBOL EQ SYMBOL / SYMBOL TILDAEQ SYMBOL | MODERATE | Grammar |
| V-OPT-1 | optproofgroup missing PROOF keyword in Go | CRITICAL | Grammar |
| V-ACT-1 | VAR uses optinit in Python, split rules in Go | LOW | Grammar |
| V-ACT-2 | Go has extra UNPROVABLE simpleact rule | LOW | Grammar |
| V-PROOF-1 | Go proofseq has extra no-separator alternative | MODERATE | Grammar |
| V-PROOF-3 | TACTIC uses atype (Go) vs SYMBOL (Python) | MODERATE | Grammar |
| V-PROOF-4 | tacticwithelem DEFINITION uses atype (Go) vs typeddefn (Python) | MODERATE | Grammar |
| V-PROOF-5 | INSTANTIATE LABEL ordering difference | LOW | Grammar |
| V-PROOF-6 | Missing INSTANTIATE LABEL ... WITH matches variant | LOW | Grammar |
| V-PROOF-7 | UNFOLD uses callatoms (Go) vs unfspecs (Python) — missing renamings/unfspec | CRITICAL | Grammar |
| V-EXPR-1 | exprterm uses appelem (Go) vs aterm (Python) | MODERATE | Grammar |
| V-FUNC-1 | lowerVarStmts is a stub in Go | CRITICAL | Implementation |

**Total: 17 variances**
- CRITICAL: 3 (V-OPT-1, V-PROOF-7, V-FUNC-1)
- MODERATE: 6 (V-TERM-1, V-LIT-1, V-PROOF-1, V-PROOF-3, V-PROOF-4, V-EXPR-1)
- LOW: 8 (V-PREC-1, V-PREC-2, V-TOK-1, V-TOK-2, V-ACT-1, V-ACT-2, V-PROOF-5, V-PROOF-6)

