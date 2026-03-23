# UNPORTED.md — Python Functions Not Yet Ported to Go

Generated: 2026-03-17

Approximately **386 functions/classes** across 14 Python files remain unported,
out of ~1054 total (~63% ported).

# CRITICAL NOTE:

When porting from Python to Go, pay special attention to preserving
Python's structural equality semantics when they are used. This
means using lg.Key() in the Go and maps keyed by lg.NodeKey. 

---

## Summary

| Python File | Total | Ported | Missing | % Ported |
|---|---|---|---|---|
| ivy_logic_utils.py | 150 | 75 | 75 | 50% |
| ivy_transrel.py | 68 | 52 | 16 | 76% |
| ivy_actions.py | 83 | 63 | 20 | 76% |
| ivy_solver.py | 105 | 82 | 23 | 78% |
| ivy_compiler.py | 112 | ~39 | ~73 | ~35% |
| ivy_isolate.py | 74 | ~59 | ~15 | ~80% |
| ivy_check.py | 44 | 38 | 6 | 86% |
| ivy_module.py | 21 | ~15 | ~6 | ~71% |
| ivy_proof.py | 112 | 52 | 60 | 46% |
| ivy_art.py | 6 | 5 | 1 | 83% |
| ivy_interp.py | 56 | 41 | 15 | 73% |
| ivy_fragment.py | 15 | 14 | 1 | 93% |
| ivy_mc.py | 40 | 30 | 10 | 75% |
| ivy_logic.py | 168 | 103 | 65 | 61% |
| **TOTAL** | **1054** | **~668** | **~386** | **~63%** |

---

## 1. ivy_logic_utils.py -> clauseops/, logicutil/ (75 MISSING)

| Python function/class | Notes |
|---|---|
| `class LogicParseError` | Exception class |
| `coerce_clause_to_formula` | |
| `apply_gen_to_clauses` | |
| `apply_func_to_clauses` | |
| `to_formula` | |
| `to_term` | |
| `to_clause` | |
| `to_clauses` | |
| `to_literal` | |
| `substitute_ast` | Note: `SubstituteConstantsAST` exists but `substitute_ast` (variable substitution) is separate |
| `replace_temporals_by_named_binder_g_ast` | |
| `replace_named_binders_ast` | |
| `resort_symbol` | |
| `resort_var` | |
| `resort_sig` | |
| `constants_ast` | Note: `UsedConstants` exists in logicutil but `constants_ast` has slightly different semantics |
| `apps_ast` | |
| `ground_apps_ast_rec` | |
| `is_ground_clause` | |
| `term_eq` | |
| `atom_eq` | |
| `lit_eq` | |
| `condition_conj` | |
| `formula_to_cube` | |
| `formula_to_clauses_aux` | |
| `lit_to_formula` | |
| `cube_to_formula` | |
| `clause_to_formula` | |
| `canonize_clause` | |
| `trim_clauses` | |
| `rewrite_clause` | |
| `triv_fmla_to_lit` | |
| `triv_fmla_to_clause` | |
| `simplify_clause_fmla` | |
| `simplify_clause` | |
| `is_tautology` | |
| `is_tautology_fmla` | |
| `lit_in_clause` | |
| `remove_duplicates` | |
| `reduce_clauses` | |
| `term_subsume` | |
| `lit_subsume` | |
| `atom_subsume` | |
| `commute_lit` | |
| `clause_subsume_recur` | |
| `clause_subsume` | |
| `subsume` | |
| `distinct_variable_renaming` | |
| `rename_variable` | |
| `used_variable_names_ast` | |
| `variables_distinct_ast` | |
| `rename_variables_distinct_asts` | |
| `variables_distinct_list_ast` | |
| `is_individual_ast` | |
| `or_clauses2` | |
| `fix_or_annot` | |
| `elim_definitions` | |
| `rename_symbols` | |
| `debug_clauses_list` | |
| `bool_const` | |
| `tagged_or_clauses` | |
| `find_true_disjunct` | |
| `eqcm_upd` | |
| `exists_quant_clauses_map` | |
| `has_enumerated_sort` | |
| `var_to_skolem` | |
| `var_to_constant` | |
| `definition_instances` | |
| `unfold_definitions_clauses` | |
| `dual_formula` | |
| `skolemize_formula` | |
| `skolemize_ast` | |
| `witness_ast` | |
| `reskolemize_clauses` | |
| `unused_constant` | |

---

## 2. ivy_transrel.py -> transrel/ (16 MISSING)

| Python function/class | Notes |
|---|---|
| `rename` | |
| `update_frame_constraint` | |
| `symbol_frame_cond` | |
| `join` | |
| `ite` | |
| `clauses_imply_formula_cex` | |
| `implies` | |
| `implies_state` | |
| `implies_action` | |
| `my_annot_op` | |
| `clausify` | |
| `clausify_state` | |
| `remove_taut_eqs_clauses` | |
| `extract_pre_post_model` | |
| `small_model_clauses` | |
| `use_numerals` | |

---

## 3. ivy_actions.py -> actions/ (20 MISSING)

| Python function/class | Notes |
|---|---|
| `p_c_a` | |
| `class UnrollContext` | |
| `class SymbolList` | |
| `get_correct_arity` | |
| `type_check` | |
| `type_ast` | |
| `destr_asgn_val` | |
| `assign_refs` | |
| `sign` | |
| `make_field_update` | |
| `my_str` | |
| `set_determinize` | |
| `bracket_action` | |
| `class DebugAction` | |
| `entry` | |
| `class TypeCheckConext` | |
| `conj_annot` | |
| `compose_annot` | |
| `rename_annot` | |
| `ite_annot` | |

---

## 4. ivy_solver.py -> solver/, z3bridge/ (23 MISSING)

Many Python standalone functions became methods on Go's `Solver` struct.
These are the truly missing ones:

| Python function/class | Notes |
|---|---|
| `solver_name` | Polymorphic symbol naming for Z3 |
| `my_minus` | |
| `my_eq` | |
| `sort_name_to_z3` | |
| `sorts` | |
| `relations` | |
| `functions` | |
| `clear` | |
| `uninterpretedsort` | |
| `functionsort` | |
| `enumeratedsort` | |
| `symbol_to_z3` | |
| `apply_z3_func` | |
| `term_to_z3` | |
| `atom_to_z3` | |
| `forall` | |
| `exists` | |
| `clause_to_z3` | |
| `conj_to_z3` | |
| `formula_to_z3_int` | |
| `formula_to_z3_closed` | |
| `get_id` | |
| `get_model` | |
| `sort_from_z3` | |
| `constant_from_z3` | |
| `get_model_constant` | |
| `get_lit_facts` | |
| `gebin` | |
| `substitute` | Z3-level substitution |

---

## 5. ivy_compiler.py -> compiler/ (~73 MISSING)

This is the largest gap. The Go compiler package has high-level pipeline
functions and `DeclInterp`/`Compiler` methods, but many helpers are missing.

| Python function/class | Notes |
|---|---|
| `class Context` | Generic compilation context |
| `class EmptyVariableContext` | |
| `thing` | |
| `other_thing` | |
| `compile_root_args` | |
| `compile_args` | |
| `sort_infer_covariant` | |
| `sort_infer_contravariant` | |
| `old_sym` | |
| `compile_isa` | |
| `cquant` | |
| `UpdatePattern_cmpl` | |
| `ConstantDecl_cmpl` | |
| `Old_cmpl` | |
| `get_arg_sorts` | |
| `get_relation_sort` | |
| `sortify` | |
| `compile_assign_lhs` | |
| `compile_crash_action` | |
| `compile_thunk_action` | |
| `compile_debug_action` | |
| `compile_native_arg` | |
| `compile_native_symbol` | |
| `compile_native_action` | |
| `compile_native_name` | |
| `compile_native_def` | |
| `compile_native_type` | |
| `compile_schema_prem` | |
| `compile_schema_conc` | |
| `compile_schema_body` | |
| `lookup_schema` | |
| `compile_schema_instantiation` | |
| `compile_let_tactic` | |
| `compile_witness_tactic` | |
| `compile_unfold_tactic` | |
| `compile_forget_tactic` | |
| `compile_if_tactic` | |
| `compile_property_tactic` | |
| `compile_function_tactic` | |
| `compile_proof_tactic` | |
| `resolve_alias_int` | |
| `class IvyDomainSetup` | Go has partial `DomainSetup` |
| `class IvyConjectureSetup` | Go has partial `ConjSetup` |
| `check_is_action` | |
| `BalancedChoice` | |
| `get_file_version` | |
| `ivy_new` | |
| `infer_parameters` | |
| `check_instantiations` | |
| `tarjan_arcs` | |
| `get_symbol_dependencies` | |
| `prop_to_def` | |
| `reorder_props` | |
| `apply_assert_proof` | |
| `apply_assert_proofs` | |
| `check_properties` | |
| `set_verifying` | |
| `ivy_compile_theory` | |
| `ivy_compile_theory_from_string` | |
| `compile_theory` | |
| `compile_theories` | |
| `add_labels_to_proof` | |
| `add_action_label` | |
| `show_call_graph` | |
| `clear_rules` | |
| `read_module` | |
| `import_module` | |
| `ivy_load_file` | |
| `ivy_from_string` | |

---

## 6. ivy_isolate.py -> isolate/ (~15 MISSING)

| Python function/class | Notes |
|---|---|
| `startswith_some_rec` | |
| `startswith_eq_some_rec` | |
| `get_strip_params` | |
| `strip_native` | |
| `strip_natives` | |
| `has_side_effect_rec` | |
| `get_props_proved_in_isolate_orig` | |
| `follow_definitions_rec` | |
| `collect_relevant_destructors` | |
| `class SortOrder` | |
| `get_cone` | |
| `get_mod_cone` | |
| `conj_to_assume` | |
| `find_some_assertion` | |
| `find_some_call` | |

---

## 7. ivy_check.py -> check/ (6 MISSING)

| Python function/class | Notes |
|---|---|
| `gui_art` | GUI/display only |
| `usage` | CLI help text |
| `show_assertions` | Display helper |
| `print_dots` | Display helper |
| `filter_fcs` | |
| `info` | Display helper |

---

## 8. ivy_module.py -> module/ (~6 MISSING)

| Python function/class | Notes |
|---|---|
| `resort_clauses` | |
| `resort_asts` | |
| `remove_refined_sortnames_from_set` | |
| `remove_refined_sortnames_from_list` | |
| `resort_map_any_ast` | |
| `class ModuleTheoryContext` | Go has partial `TheoryContext` |

---

## 9. ivy_proof.py -> proof/ (60 MISSING)

| Python function/class | Notes |
|---|---|
| `attrib_goals` | |
| `is_goal` | |
| `goals_subst` | |
| `fresh_label` | |
| `goal_prefix_prems` | |
| `check_name_clash` | |
| `check_premises_provided` | |
| `goal_is_temporal` | |
| `goal_defines` | |
| `get_unprovided_defns` | |
| `goal_subgoals` | |
| `fmla_vocab` | |
| `check_schema_capture` | |
| `check_alpha_capture` | |
| `check_renaming` | |
| `rename_goal` | |
| `compile_expr_vocab` | |
| `compile_expr_vocab_ext` | |
| `remove_vars_match` | |
| `show_match` | |
| `transform_defn_schema` | |
| `transform_defn_match` | |
| `goal_prems_by_name` | |
| `add_prem_match` | |
| `parameterize_schema` | |
| `compile_match_list` | |
| `compile_one_match` | |
| `compile_match` | |
| `match_rhs_vars` | |
| `is_lambda` | |
| `goal_is_property` | |
| `goal_is_schema` | |
| `apply_match_match` | |
| `rename_problem` | |
| `avoid_capture_problem` | |
| `goals_eq_mod_alpha` | |
| `apply_match_rec` | |
| `raise_capture` | |
| `match_get` | |
| `apply_match_alt` | |
| `apply_fun` | |
| `apply_match_alt_rec` | |
| `apply_match_func_alt` | |
| `apply_match_sort` | |
| `apply_match_freesyms_alt` | |
| `make_distinct_vars` | |
| `var_subst_goal` | |
| `apply_to_conc` | |
| `compile_witness_list` | |
| `remove_unused_definitions_goal` | |
| `match_from_defn` | |
| `match_from_defns` | |
| `unfold_goal` | |
| `unfold_fmla` | |
| `goal_apply_to_prem` | |
| `goal_apply_to_conc` | |
| `close_unmatched` | |
| `drop_supplied_prems` | |
| `remove_explicit` | |
| `rename_prem_no_clash` | |

---

## 10. ivy_art.py -> art/ (1 MISSING)

| Python function/class | Notes |
|---|---|
| `render_rg` | Rendering/display only |

---

## 11. ivy_interp.py -> interp/ (15 MISSING)

| Python function/class | Notes |
|---|---|
| `type_check_list` | |
| `module_order` | |
| `module_skolemizer` | |
| `get_core` | |
| `reverse_join_concrete_clauses` | |
| `underapproximate_state` | |
| `states_state_expr` | |
| `decompose_action_app` | |
| `state_implies_formula` | |
| `new_history` | |
| `eval_assert_rhs` | |
| `eval_state_order` | |
| `check_state_assertion` | |
| `get_state_assertions` | |
| `universe_constraint` | |

---

## 12. ivy_fragment.py -> fragment/ (1 MISSING)

| Python function/class | Notes |
|---|---|
| `show_strat_graph` | Debug display only |

---

## 13. ivy_mc.py -> mc/ (10 MISSING)

| Python function/class | Notes |
|---|---|
| `encode_vars` | |
| `clone_normal` | |
| `uncompose_annot` | |
| `unite_annot` | |
| `class MatchHandler` | MC-specific match handler |
| `match_annotation` | |
| `checked` | |
| `badwit` | |
| `class IvyMCTrace` | |
| `aiger_witness_to_ivy_trace` | |

---

## 14. ivy_logic.py -> ivylogic/ (65 MISSING)

| Python function/class | Notes |
|---|---|
| `class top_sort_as_default` | Context manager |
| `class alpha_sort_as_default` | Context manager |
| `class AST` | Base class |
| `Atom` | Constructor function |
| `_find_sort` | |
| `find_sort` | |
| `add_sort` | |
| `find_symbol` | |
| `normalize_symbol` | |
| `add_symbol` | |
| `remove_symbol` | |
| `all_symbols` | |
| `get_sort_term` | |
| `seg_var_pat` | Note: Go has `segVarPat` in classify_ext.go (PORTED) |
| `reason` | |
| `class NotEssentiallyUninterpreted` | Exception class |
| `check_essentially_uninterpreted` | |
| `symbols_over_universals_rec` | Note: Go has this in classify_ext.go (PORTED) |
| `universal_variables_rec` | Note: Go has this in classify_ext.go (PORTED) |
| `Constant` | Constructor function |
| `quantifier_body` | |
| `binder_args` | |
| `is_epr_rec` | Note: Go has `isEPRRec` in classify_ext.go (PORTED) |
| `_eq_lit` | |
| `_neq_lit` | |
| `class PolySymsDict` | Dynamic polymorphic symbol lookup |
| `apply` | |
| `default_sort` | |
| `Equals` | Constructor function |
| `is_enumerated` | |
| `is_concretely_sorted` | |
| `ast_match_lists` | Note: Go has `astMatchLists` in util.go (PORTED) |
| `sorts` | |
| `to_str_with_var_sorts` | |
| `fmla_to_str_ambiguous` | |
| `And_ugly` | Display function |
| `Or_ugly` | Display function |
| `Not_ugly` | Display function |
| `Implies_ugly` | Display function |
| `Iff_ugly` | Display function |
| `Eq_ugly` | Display function |
| `Ite_ugly` | Display function |
| `Apply_ugly` / `app_ugly` | Display function |
| `ForAll_ugly` | Display function |
| `Exists_ugly` | Display function |
| `Lambda_ugly` | Display function |
| `NamedBinder_ugly` | Display function |
| `paren` | Display helper |
| `drop_annotations` | Note: Go has this in logic/pretty.go (PORTED) |
| `typed_symbol` | |
| `pretty_fmla` | Note: Go has `PrettyFmla` in logic/pretty.go (PORTED) |
| `is_canonical_sort` | |
| `canonize_sort` | |
| `sort_refinement` | |
| `uninterpreted_sorts` | |
| `interpreted_sorts` | |
| `is_deterministic_fmla` | |
| `sym_decl_to_str` | |
| `typed_sym_to_str` | |
| `sig_to_str` | |
| `pto` | |
| `lambda_apply` | |
| `rename_vars_no_clash` | |
| `equal_mod_alpha` | Note: Go has `EqualModAlpha` in logicutil (PORTED) |
| `alpha_rename` | |
| `normalized_and` | |

---

## Biggest Gaps by Priority

1. **ivy_logic_utils.py** (75 missing) — Subsumption, clause manipulation, skolemization, definition elimination
2. **ivy_compiler.py** (~73 missing) — Tactic compilation, schema compilation, native code, module loading
3. **ivy_logic.py** (65 missing) — Sort management, PolySymsDict, some constructor functions (many display funcs already ported via PrettyFmla)
4. **ivy_proof.py** (60 missing) — Match compilation, schema matching, goal manipulation, tactic application
5. **ivy_solver.py** (23 missing) — Z3 sort creation helpers, some translation functions
