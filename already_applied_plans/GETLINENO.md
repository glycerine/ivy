# Plan: Audit Missing `get_lineno` Locations for Incomplete Port Logic

**Created:** 2026-03-25T20:00

## Context

Golden test diverges at line 61325 because Go's `p_optlabel_label` rule doesn't call `tokLineno`. But each missing `get_lineno` is a **symptom of incomplete porting** — the surrounding logic is likely also missing or divergent. Rather than blindly adding lineno calls, we audit each location thoroughly, comparing the FULL Python rule body against Go, and document ALL missing port logic in `~/goivy/TODO_MISSING_PORT.md`.

Python total: 331 `get_lineno(p,N)` calls. Go total: 232 active `tokLineno`/`getLineno` calls. Gap: ~99 across ~96 rules.

## Analysis

The 107 missing calls fall into three categories:

### Category A: Rules that EXIST in Go but need tokLineno/getLineno added (36 rules, ~42 calls)

These rules exist in `grammar_v17.y` with trace strings, but lack the `tokLineno`/`getLineno` call that Python's `get_lineno` provides. These are straightforward fixes.

**Entirely missing lineno calls (rule exists, 0 tokLineno/getLineno):**

| Go Rule | Py calls | Fix needed |
|---------|----------|------------|
| `p_optlabel_label` | 1 | Add `$$.SetLineno(tokLineno(lex, $1))` |
| `p_app_symbol` | 1 | Add tokLineno |
| `p_app_symbol_lp_terms_rp` | 1 | Add tokLineno |
| `p_app_term_infix_term` | 1 | Add tokLineno |
| `p_aterm_aterm_dot_appelem` | 1 | Add tokLineno |
| `p_callatom_callatom_dot_callatom` | 1 | Add tokLineno |
| `p_callatom_method` | 1 | Add tokLineno |
| `p_fmla_fmla_isa_atype` | 2 | Add 2 tokLineno |
| `p_inst_modinst` | 1 | Add tokLineno |
| `p_lgprop` | 1 | Add tokLineno |
| `p_lit_atom` | 1 | Add tokLineno |
| `p_lit_term_eq_term` | 1 | Add tokLineno |
| `p_lit_term_tildaeq_term` | 1 | Add tokLineno |
| `p_lit_tilda_atom` | 1 | Add tokLineno |
| `p_lparam_caret_variable_colon_symbol` | 1 | Add tokLineno |
| `p_objectend` | 1 | Add tokLineno |
| `p_optactiondef_eq_symbol` | 1 | Add tokLineno |
| `p_pname_symbol` | 1 | Add tokLineno |
| `p_proofseq_proofseq_semi_proofstep` | 1 | Add tokLineno |
| `p_proofstep_theorem` | 1 | Add tokLineno |
| `p_renaming_lt_renaminglist_gt` | 1 | Add tokLineno |
| `p_renamingitem_symbol_div_symbol` | 1 | Add tokLineno |
| `p_renamingitem_variable_div_variable` | 1 | Add tokLineno |
| `p_simpleact_debug_symbol_optdebugargs` | 2 | Add 2 tokLineno |
| `p_tacticwithelem_trigger` | 1 | Add tokLineno |
| `p_targs_lparen_tsyms_rparen` | 1 | Add tokLineno |
| `p_term_namedbinder_dollar_fmla` | 1 | Add tokLineno |
| `p_term_namedbinder_dot_fmla` | 1 | Add tokLineno |
| `p_term_namedbinder_vars_dot_term` | 2 | Add 2 tokLineno |
| `p_term_term_colon_term` | 1 | Add tokLineno |
| `p_top_around_callatom_lcb_action_rcb` | 1 | Add tokLineno |
| `p_top_assert_symbol_arrow_assert_rhs` | 1 | Add tokLineno |
| `p_top_delegate_callatom_opt` | 1 | Add tokLineno |
| `p_top_implement_type_symbol_with_symbol` | 4 | Add 4 tokLineno |
| `p_top_nativequote` | 2 | Add 2 tokLineno |
| `p_top_object_symbol_eq_lcb_top_rcb` | 1 | Add tokLineno |
| `p_top_specification_lcb_top_rcb` | 1 | Add tokLineno |
| `p_top_unprovable_invariant_labeledfmla` | 1 | Add tokLineno |

**Rules with fewer calls than Python (need additional tokLineno):**

| Go Rule | Py calls | Go calls | Missing |
|---------|----------|----------|---------|
| `p_action_call_callatom` | 2 | 1 | 1 |
| `p_action_require_proof_proofstep` | 3 | 2 | 1 |
| `p_defn_atom_fmla` | 2 | 1 | 1 |
| `p_proofstep_assume` | 2 | 1 | 1 |
| `p_proofstep_assume_with_defns` | 2 | 1 | 1 |
| `p_proofstep_instantiate` | 2 | 1 | 1 |
| `p_proofstep_instantiate_with_defns` | 2 | 1 | 1 |
| `p_proofstep_spoil_atype` | 2 | 1 | 1 |
| `p_proofstep_symbol` | 4 | 2 | 2 |
| `p_proofstep_symbol_with_defns` | 4 | 2 | 2 |
| `p_proofstep_unfold_atype_with_defns` | 2 | 1 | 1 |
| `p_proofstep_witness_pflets` | 2 | 1 | 1 |
| `p_scenariomixin_after_callatom_lcb_action_rcb` | 2 | 1 | 1 |
| `p_scenariomixin_before_callatom_lcb_action_rcb` | 2 | 1 | 1 |
| `p_term_if_fmla_else_term` | 3 | 1 | 2 |
| `p_top__top_class_symbol_objectargs_eq_lcb_optdo` | 3 | 2 | 1 |
| `p_top__top_subclass_symbol_of_atype_eq_lcb_optd` | 3 | 2 | 1 |
| `p_top_after_callatom_lcb_action_rcb` | 2 | 1 | 1 |
| `p_top_axiom_optlabel_gprop` | 2 | 1 | 1 |
| `p_top_interpret_symbol_arrow_lcb_symbol_moresymbols_rcb` | 3 | 2 | 1 |
| `p_top_optimpex_action_symbol_optargs_optreturns_eq_action` | 3 | 1 | 2 |
| `p_top_opttrusted_extract_callatom_eq_lcb_top_rcb_optwith` | 3 | 2 | 1 |
| `p_top_type_symbol_eq_sort` | 4 | 3 | 1 |
| `p_top_variant_symbol_of_atype` | 2 | 1 | 1 |
| `p_top_variant_symbol_of_symbol_eq_sort` | 2 | 1 | 1 |

### Category B: Rules that DON'T EXIST in Go grammar (35 rules, ~40 calls)

These Python grammar rules have no corresponding Go grammar rule at all. They represent **unported grammar rules**, which is a larger task than just adding lineno calls. Key groups:

- **`p_fmla_*` rules (20+)**: Python's `ivy_logic_parser.py` has separate `fmla` productions. Go unified `fmla` into `term`. These `fmla_*` rules don't fire in Go because the `term` rule handles them. The Go `term_*` equivalents already have tokLineno. **No action needed** — these are structural differences where Go's unified `term` rule already handles lineno.
- **`p_action_field_assign_*` (4)**: Field assign actions not yet ported.
- **`p_action_if_fmla_*` (2)**: If-action variants using `fmla` (Go uses `somefmla`).
- **`p_aterm_*` (2)**: Atom-term rules.
- **Other misc** (7): Various unported rules.

### Category C: Rules where Go has MORE calls than Python (8 rules)

These may be bugs (extra traces) or intentional differences. Lower priority.

## Implementation Plan

### Step 1: Fix Category A rules (the 36+25 = 61 rules needing tokLineno additions)

For each rule listed above:
1. Read the corresponding Python rule to see exactly what `get_lineno(p, N)` sets and which position N
2. Add the matching `tokLineno(v17lex.(*v17LexAdapter), $N)` or `getLineno(v17lex.(*v17LexAdapter))` call
3. Set it on the correct AST node using `.SetLineno()`

**Approach**: Process each rule one at a time:
- Find the Go rule by its xtracer.Trace string
- Find the Python rule by its `def p_` function name
- Match the `get_lineno(p, N)` call → add equivalent `tokLineno`/`getLineno` call
- The `N` in `get_lineno(p, N)` corresponds to `$N` in goyacc

### Step 2: Regenerate grammar_v17.go

```bash
cd lalr_full && go generate
```

### Step 3: Build and test

```bash
cd ~/goivy && go build ./... && go test ./lalr_full/ -run TestOrdLive -v
```

### Step 4: Investigate Category C (extra calls) if needed

After fixing missing calls, some Go rules may emit extra traces. Address these only if they cause golden test failures.

## Key Files

| File | Role |
|------|------|
| `lalr_full/grammar_v17.y` | Go grammar — add tokLineno/getLineno calls |
| `lalr_full/grammar_v17.go` | Regenerated from .y |
| `~/pyivy/ivy/ivy/ivy_parser.py` | Python reference — 252 get_lineno calls |
| `~/pyivy/ivy/ivy/ivy_logic_parser.py` | Python reference — 81 get_lineno calls |

## Helper Functions (already exist)

- `tokLineno(lex *v17LexAdapter, tok TokenInfo) ast.Location` — grammar_v17.y:411 — uses tok.Line, traces "parser.get_lineno ENTER"
- `getLineno(lex *v17LexAdapter) ast.Location` — grammar_v17.y:32 — uses lex.prevTok.Line, traces "parser.get_lineno ENTER"

## Verification

```bash
cd ~/goivy && make golden
```

Golden test should advance well past line 61325. Many of the 107 missing calls are in rules that don't fire for `ord_live.ivy`, so the test may advance significantly.

## Note on Category B

The 35 truly missing rules represent unported grammar productions. These should be tracked separately as a grammar completeness task, not conflated with the lineno fix.
