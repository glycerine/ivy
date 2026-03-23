# Plan: Fix All 17 Grammar Variances in Go LALR Port

**Created:** 2026-03-23 21:30

## Context

We performed a production-by-production comparison of the Python Ivy grammar
(source of truth) vs the Go LALR port, producing `~/goivy/GRAMMAR_VARIANCES.md`
with 17 variances. ALL must be fixed. This is formal methods — there are no
low-priority items.

## Critical Files to Modify

1. `~/goivy/lalr_full/grammar_v17.y` — main grammar file (all 17 fixes)
2. Regenerate `~/goivy/lalr_full/grammar_v17.go` — via `goyacc`

## Existing Code to Reuse

- `ast.LowerVarStatements()` in `~/goivy/ast/lower_var.go` — already implements
  the full `lower_var_stmts` logic. The grammar stub just needs to call it.
- `ast.UnfoldSpec` struct in `~/goivy/ast/tactic.go` — already exists with
  `DefName Node` and `Renamings []Node` fields.
- `ast.SubstPrefixAtomsAst()` in `~/goivy/ast/rewrite.go` — used by LowerVarStatements.

---

## Fix 1: V-PREC-1 — Remove TOK_ISA from precedence table

**File:** `grammar_v17.y` line 483

**Current:**
```
%left   TOK_EQ TOK_LE TOK_LT TOK_GE TOK_GT TOK_PTO TOK_ISA
```

**Fix:**
```
%left   TOK_EQ TOK_LE TOK_LT TOK_GE TOK_GT TOK_PTO
```

Python does NOT put ISA in the precedence table. ISA gets its precedence from
the grammar structure (`term : term ISA atype`). Removing it from `%left` makes
Go match Python exactly.

---

## Fix 2: V-PREC-2 — Remove %right TOK_ASSIGN from precedence table

**File:** `grammar_v17.y` line 493

**Current:**
```
%right  TOK_ASSIGN
```

**Fix:** Delete line 493 entirely.

Python does NOT have ASSIGN in its precedence table. ASSIGN only appears in
action rules (`simpleact : term ASSIGN fmla`), never in expression rules. If
removing this causes goyacc shift/reduce conflicts, we can add `%prec` directives
to the specific `ASSIGN` rules instead, but removing it first and checking is the
faithful approach.

---

## Fix 3: V-TOK-1 — Keep TOK_LABEL (document as intentional design choice)

**File:** `grammar_v17.y` line 283

No code change needed. The Go approach of handling `[identifier]` via
`labelname : TOK_LB SYMBOLx TOK_RB` at grammar level (instead of lexer-level
LABEL token concatenation) is functionally equivalent. Add a comment:

```
%token <str>  TOK_LABEL   // declared for completeness; Go handles labels via
                           // labelname : TOK_LB SYMBOLx TOK_RB instead of lexer
```

---

## Fix 4: V-TOK-2 — Remove unused placeholder tokens

**File:** `grammar_v17.y` lines 340, 348

**Current:**
```
%token  TOK_VAR_KW  // not used yet, placeholder
...
%token  TOK_METHOD_KW TOK_NULL_KW TOK_SET_KW
```

**Fix:** Delete these three lines. They add noise and confusion.

---

## Fix 5: V-TERM-1 / V-EXPR-1 — Add aterm rule, fix exprterm

**File:** `grammar_v17.y`

**Add** new `aterm` nonterminal (matching Python lines 193, 198):

In `%type` declarations, add:
```
%type <node>  aterm
```

After the `appelem` rule, add:
```
aterm:
    appelem
    {
        xtracer.Trace("parser.p_aterm_aappelem ENTER (aterm)")
        $$ = $1
    }
    | aterm TOK_DOT appelem
    {
        xtracer.Trace("parser.p_aterm_aterm_dot_appelem ENTER (aterm)")
        $$ = ast.ComposeAtoms($1.(*ast.Atom), $3.(*ast.Atom))
    }
    ;
```

**Change** `exprterm` (line 4203) from:
```
exprterm:
    appelem
    { ... }
    | var
    { ... }
    ;
```
To:
```
exprterm:
    aterm
    {
        xtracer.Trace("parser.p_exprterm_aterm ENTER (exprterm)")
        $$ = $1
    }
    | var
    {
        xtracer.Trace("parser.p_exprterm_var ENTER (exprterm)")
        $$ = $1
    }
    ;
```

---

## Fix 6: V-LIT-1 — Add missing lit alternatives

**File:** `grammar_v17.y` lines 2620-2631

**Current:**
```
lit:
    atom
    { ... }
    | TOK_TILDA lit
    { ... }
    ;
```

**Add** two alternatives after `atom`:
```
lit:
    atom
    {
        xtracer.Trace("parser.p_lit_atom ENTER (lit)")
        $$ = $1
    }
    | SYMBOLx TOK_EQ SYMBOLx
    {
        xtracer.Trace("parser.p_lit_term_eq_term ENTER (lit)")
        $$ = ast.NewAtom("=", ast.NewAtom($1), ast.NewAtom($3))
    }
    | SYMBOLx TOK_TILDAEQ SYMBOLx
    {
        xtracer.Trace("parser.p_lit_term_tildaeq_term ENTER (lit)")
        $$ = &ast.Not{Body: ast.NewAtom("=", ast.NewAtom($1), ast.NewAtom($3))}
    }
    | TOK_TILDA lit
    {
        xtracer.Trace("parser.p_lit_tilda_atom ENTER (lit)")
        $$ = &ast.Not{Body: $2}
    }
    ;
```

Python wraps in `Literal(polarity, atom)` but we can use raw AST nodes. The key
is that the grammar accepts these forms — semantic details can be refined later.

---

## Fix 7: V-OPT-1 — Add PROOF keyword to optproofgroup

**File:** `grammar_v17.y` lines 3784-3795

**Current:**
```
optproofgroup:
    /* empty */
    { $$ = &ast.NoneAST{} }
    | proofgroup
    { $$ = $1 }
    ;
```

**Fix:**
```
optproofgroup:
    /* empty */
    {
        xtracer.Trace("parser.p_optproofgroup ENTER (optproofgroup)")
        $$ = nil
    }
    | TOK_PROOF proofgroup
    {
        xtracer.Trace("parser.p_optproofgroup_symbol ENTER (optproofgroup)")
        $$ = $2
    }
    ;
```

This is CRITICAL. Without `TOK_PROOF`, Go greedily consumes `{ }` blocks as
proof groups when they should be parsed as something else (e.g., a sequence body).

---

## Fix 8: V-ACT-1 — Document as intentional (no code change)

The Go approach of splitting `VAR tterm` and `VAR tterm ASSIGN fmla` into two
separate rules is functionally equivalent to Python's `VAR tterm optinit`. No
change needed — the parse results are identical.

---

## Fix 9: V-ACT-2 — Remove standalone UNPROVABLE simpleact rule

**File:** `grammar_v17.y` line 3376

**Current:**
```
    | TOK_UNPROVABLE simpleact
    {
        xtracer.Trace("parser.p_simpleact__unprovable_simpleact ENTER (simpleact)")
        $$ = &ast.Sequence{} // no-op
    }
```

**Fix:** Delete this alternative entirely.

Python does NOT have this rule. Python only allows `optunprovable` as a prefix
on `ASSERT`, `REQUIRE`, and `ENSURE` — never on arbitrary simple actions. The Go
rule is overly permissive and doesn't match the source of truth.

---

## Fix 10: V-PROOF-1 — Remove extra no-separator proofseq alternative

**File:** `grammar_v17.y` lines 3797-3813

**Current:**
```
proofseq:
    proofstep
    { ... }
    | proofseq TOK_SEMI proofstep
    { ... }
    | proofseq proofstep           ← DELETE THIS
    { ... }
    ;
```

**Fix:** Change to match Python exactly:
```
proofseq:
    proofstep
    {
        xtracer.Trace("parser.p_proofseq_proofstep ENTER (proofseq)")
        $$ = $1
    }
    | proofseq optsemi proofstep
    {
        xtracer.Trace("parser.p_proofseq_proofseq_semi_proofstep ENTER (proofseq)")
        $$ = &ast.ComposeTactics{Tactics: []ast.Node{$1, $3}}
    }
    ;
```

Python uses `proofseq optsemi proofstep` — the `optsemi` allows an optional
semicolon but still forces a grammar reduction through the `optsemi` nonterminal.
The Go version had two problems: (a) using `TOK_SEMI` instead of `optsemi`, and
(b) having an extra bare `proofseq proofstep` alternative.

---

## Fix 11: V-PROOF-3 — Change TACTIC from atype to SYMBOLx

**File:** `grammar_v17.y` line 3949

**Current:**
```
    | TOK_TACTIC atype opttacticwith optproofgroup
    {
        ...
        $$ = &ast.TacticTactic{TName: $2, Body: $3, Proof: $4}
    }
```

**Fix:**
```
    | TOK_TACTIC SYMBOLx opttacticwith optproofgroup
    {
        xtracer.Trace("parser.p_proofstep_tactic ENTER (proofstep)")
        a := ast.NewAtom($2)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = &ast.TacticTactic{TName: a, Body: $3, Proof: $4}
    }
```

Python uses bare `SYMBOL`, not `atype`. A tactic name is always a simple
identifier, never a dotted name or `this`.

---

## Fix 12: V-PROOF-4 — Change tacticwithelem DEFINITION from atype to typeddefn

**File:** `grammar_v17.y` line 3715

**Current:**
```
    | TOK_DEFINITION atype TOK_EQ fmla
    {
        ...
        $$ = ast.NewDefinition($2, $4)
    }
```

**Fix:**
```
    | TOK_DEFINITION typeddefn TOK_EQ fmla
    {
        xtracer.Trace("parser.p_tacticwithelem_fun_defn ENTER (tacticwithelem)")
        df := ast.NewDefinition(ast.AppToAtom($2), $4)
        df.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        lf := addLabel(mkLF(df), "def")
        $$ = ast.NewDerivedDecl(lf)
    }
```

Python uses `typeddefn` (which allows `defnlhs` with optional args and sort
annotation), not `atype` (just a type name). The semantic action matches
Python's `p_tacticwithelem_fun_defn` (ivy_parser.py:1525-1530).

---

## Fix 13: V-PROOF-5 — Document as intentional (no code change needed)

Go uses `labelname` (which handles both `TOK_LB SYMBOLx TOK_RB` and
`TOK_LABEL`) where Python uses bare `LABEL`. Since Go's lexer doesn't
return `TOK_LABEL`, using `labelname` is the correct equivalent. No change.

---

## Fix 14: V-PROOF-6 — Add INSTANTIATE labelname atype optrenaming WITH matches

**File:** `grammar_v17.y` — add after the existing INSTANTIATE labelname rule (~line 3923)

**Add:**
```
    | TOK_INSTANTIATE labelname atype optrenaming TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_instance_with_matches ENTER (proofstep)")
        $$ = &ast.AssumeTactic{SchemaName: $3, Ren: $4, Matches: $6}
    }
```

This is the labeled variant of `INSTANTIATE ... WITH matches` that Python has
at line 1403 but Go was missing.

---

## Fix 15: V-PROOF-7 — Add unfspec, unfspecs, renamings; fix UNFOLD rules

**File:** `grammar_v17.y`

### Step 1: Add %type declarations
```
%type <node>  unfspec
%type <nodes> unfspecs renamings
```

### Step 2: Add renamings rule (after renaming rule, ~line 3889)
```
renamings:
    /* empty */
    {
        xtracer.Trace("parser.p_renamings ENTER (renamings)")
        $$ = nil
    }
    | renamings renaming
    {
        xtracer.Trace("parser.p_renamings_renamings_renaming ENTER (renamings)")
        $$ = append($1, $2)
    }
    ;
```

### Step 3: Add unfspec and unfspecs rules (after renamings)
```
unfspec:
    callatom renamings
    {
        xtracer.Trace("parser.p_unfspec_callatom_renamings ENTER (unfspec)")
        spec := &ast.UnfoldSpec{DefName: $1, Renamings: $2}
        spec.SetLineno($1.(interface{ GetLineno() int }).(ast.Node).GetLineno())
        $$ = spec
    }
    ;

unfspecs:
    unfspec
    {
        xtracer.Trace("parser.p_unfspecs_unfspec ENTER (unfspecs)")
        $$ = []ast.Node{$1}
    }
    | unfspecs TOK_COMMA unfspec
    {
        xtracer.Trace("parser.p_unfspecs_unfspecs_unfspec ENTER (unfspecs)")
        $$ = append($1, $3)
    }
    ;
```

### Step 4: Change UNFOLD proofstep rules (lines 3993-4003)

**Current:**
```
    | TOK_UNFOLD atype TOK_WITH callatoms
    { ... }
    | TOK_UNFOLD TOK_WITH callatoms
    { ... }
```

**Fix:**
```
    | TOK_UNFOLD atype TOK_WITH unfspecs
    {
        xtracer.Trace("parser.p_proofstep_unfold_atype_with_defns ENTER (proofstep)")
        a := ast.NewAtom($2.(*ast.Symbol).Rep)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = &ast.UnfoldTactic{Premise: a, UnfSpecs: $4}
    }
    | TOK_UNFOLD TOK_WITH unfspecs
    {
        xtracer.Trace("parser.p_proofstep_unfold_with_defns ENTER (proofstep)")
        $$ = &ast.UnfoldTactic{Premise: &ast.NoneAST{}, UnfSpecs: $3}
    }
```

Note: `UnfoldTactic.UnfSpecs` field type should accept `[]ast.Node` containing
`*ast.UnfoldSpec` elements. The existing `ast/tactic.go` already has this.

---

## Fix 16: V-FUNC-1 — Wire lowerVarStmts to ast.LowerVarStatements

**File:** `grammar_v17.y` line 4277

**Current:**
```go
func lowerVarStmts(stmts []ast.Node) []ast.Node {
	xtracer.Trace("parser.lower_var_stmts ENTER")
	// TODO: implement VarAction → LocalAction transformation when tests need it
	return stmts
}
```

**Fix:**
```go
func lowerVarStmts(stmts []ast.Node) []ast.Node {
	xtracer.Trace("parser.lower_var_stmts ENTER")
	return ast.LowerVarStatements(stmts)
}
```

The full implementation already exists at `~/goivy/ast/lower_var.go`. The grammar
function just needs to delegate to it.

---

## Fix 3 (comment): V-TOK-1 — Add clarifying comment

**File:** `grammar_v17.y` line 283

Change:
```
%token <str>  TOK_LABEL
```
To:
```
%token <str>  TOK_LABEL   // Python returns LABEL from lexer; Go handles via labelname rule instead
```

---

## Execution Order

All 16 changes (14 code changes + 2 document-only) go into `grammar_v17.y` in
one pass. Then regenerate:

```bash
cd ~/goivy/lalr_full
go generate    # runs goyacc to produce grammar_v17.go
go build ./...  # verify compilation
```

After regeneration, check S/R conflict count. Current count is 314. The changes
should not significantly increase conflicts. Removing ISA from precedence and
removing ASSIGN from precedence may slightly change the count — if new conflicts
appear, resolve with `%prec` directives on the affected rules.

## Verification

```bash
cd ~/goivy && make golden
```

If `make golden` still shows the same divergence point as before (line 4872
decl count mismatch), none of these grammar fixes caused regressions. If it
advances further, even better.

Also verify that the grammar compiles without errors:
```bash
cd ~/goivy && go build ./...
```

## Summary of Changes

| Fix | Variance | Change |
|-----|----------|--------|
| 1 | V-PREC-1 | Remove TOK_ISA from `%left` line |
| 2 | V-PREC-2 | Delete `%right TOK_ASSIGN` line |
| 3 | V-TOK-1 | Add comment to TOK_LABEL |
| 4 | V-TOK-2 | Delete unused tokens (TOK_VAR_KW, TOK_METHOD_KW, TOK_NULL_KW, TOK_SET_KW) |
| 5 | V-TERM-1+V-EXPR-1 | Add `aterm` rule; change `exprterm` to use `aterm` |
| 6 | V-LIT-1 | Add `lit : SYMBOLx EQ SYMBOLx` and `SYMBOLx TILDAEQ SYMBOLx` |
| 7 | V-OPT-1 | Add `TOK_PROOF` before `proofgroup` in `optproofgroup` |
| 8 | V-ACT-1 | No change (intentional equivalent) |
| 9 | V-ACT-2 | Delete `UNPROVABLE simpleact` rule |
| 10 | V-PROOF-1 | Replace 2 proofseq alts with `proofseq optsemi proofstep` |
| 11 | V-PROOF-3 | Change TACTIC from `atype` to `SYMBOLx` |
| 12 | V-PROOF-4 | Change tacticwithelem DEFINITION from `atype` to `typeddefn` |
| 13 | V-PROOF-5 | No change (intentional equivalent) |
| 14 | V-PROOF-6 | Add `INSTANTIATE labelname atype optrenaming WITH matches` |
| 15 | V-PROOF-7 | Add `renamings`, `unfspec`, `unfspecs`; fix UNFOLD rules |
| 16 | V-FUNC-1 | Wire `lowerVarStmts` to `ast.LowerVarStatements` |

**Total: 14 code changes, 2 no-change (documented as intentional)**
