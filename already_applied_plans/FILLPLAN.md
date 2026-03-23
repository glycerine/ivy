# FILLPLAN.md — Plan to Port All 386 Remaining Python Functions to Go

Generated: 2026-03-18

## Context

The goivy project is a mechanical port of Ivy (formal verification language) from Python to Go.
~63% of functions are ported (~668/1054). This plan addresses the remaining ~386 functions
across 14 Python files. The Python source at `/Users/jaten/pyivy/ivy/ivy/` is the source of truth.

**Cardinal rules:** Match Python exactly. Same names, same structure, same flow. Depth-first — implement everything now, never defer.

---

## Functions Already Ported (Verify Only)

These are marked PORTED in UNPORTED.md — verify they match Python semantics, do not re-implement:

| Python Function | Go Location |
|---|---|
| `seg_var_pat` | `ivylogic/classify_ext.go` |
| `symbols_over_universals_rec` | `ivylogic/classify_ext.go` |
| `universal_variables_rec` | `ivylogic/classify_ext.go` |
| `is_epr_rec` | `ivylogic/classify_ext.go` |
| `ast_match_lists` | `ivylogic/util.go` |
| `drop_annotations` | `logic/pretty.go` |
| `pretty_fmla` | `logic/pretty.go` |
| `equal_mod_alpha` | `logicutil/logicutil.go` |

Many solver functions listed as "missing" standalone in Python already exist as `Solver` methods or `z3bridge.Translator` methods in Go. Phase 2 requires careful verification before adding code.

---

## Phase 1: Foundation Layer — ivy_logic.py + ivy_logic_utils.py (140 functions, 6-7 sessions)

### Batch 1.1 — ivy_logic.py: Sort/Symbol Management (15 functions)
**Target:** `ivylogic/sig.go`, `ivylogic/ivylogic.go`
**Complexity:** MEDIUM

| # | Python | Go Name |
|---|---|---|
| 1 | `_find_sort` | `FindSortInternal` |
| 2 | `find_sort` | `FindSort` |
| 3 | `add_sort` | `AddSort` (module-level) |
| 4 | `find_symbol` | `FindSymbol` (module-level) |
| 5 | `normalize_symbol` | `NormalizeSymbol` |
| 6 | `add_symbol` | `AddSymbol` (module-level) |
| 7 | `remove_symbol` | `RemoveSymbol` |
| 8 | `all_symbols` | `AllSymbols` |
| 9 | `get_sort_term` | `GetSortTerm` |
| 10 | `default_sort` | `DefaultSort` |
| 11 | `sorts` | `Sorts` |
| 12 | `is_enumerated` | `IsEnumerated` |
| 13 | `is_concretely_sorted` | `IsConcretetlySorted` |
| 14 | `is_canonical_sort` | `IsCanonicalSort` |
| 15 | `canonize_sort` | `CanonizeSort` |

**Notes:** Python uses global `sig`; Go passes `*Sig` explicitly. Several may already exist as Sig methods — verify first.

### Batch 1.2 — ivy_logic.py: Constructors and Display (18 functions)
**Target:** `ivylogic/ivylogic.go`, new `ivylogic/display.go`
**Complexity:** LOW-MEDIUM

| # | Python | Go Name |
|---|---|---|
| 1 | `Atom` | `Atom` (constructor) |
| 2 | `Constant` | `Constant` (constructor) |
| 3 | `Equals` | `Equals` (constructor) |
| 4 | `apply` | `Apply` |
| 5 | `quantifier_body` | `QuantifierBody` |
| 6 | `binder_args` | `BinderArgs` |
| 7 | `_eq_lit` / `_neq_lit` | `EqLit` / `NeqLit` |
| 8-18 | `And_ugly` through `NamedBinder_ugly`, `paren` | Display functions in `display.go` |

### Batch 1.3 — ivy_logic.py: Context Managers and Misc (12 functions)
**Target:** `ivylogic/globals.go`, new `ivylogic/poly.go`
**Complexity:** MEDIUM

| # | Python | Go Name |
|---|---|---|
| 1 | `class top_sort_as_default` | `TopSortAsDefault` struct (Enter/Exit) |
| 2 | `class alpha_sort_as_default` | `AlphaSortAsDefault` struct (Enter/Exit) |
| 3 | `class AST` | verify `logic.Node` covers this |
| 4 | `class PolySymsDict` | `PolySymsDict` struct |
| 5 | `class NotEssentiallyUninterpreted` | error type |
| 6 | `check_essentially_uninterpreted` | `CheckEssentiallyUninterpreted` |
| 7 | `reason` | `Reason` |
| 8 | `sort_refinement` | `SortRefinement` |
| 9 | `uninterpreted_sorts` | `UninterpretedSorts` |
| 10 | `interpreted_sorts` | `InterpretedSorts` |
| 11 | `is_deterministic_fmla` | `IsDeterministicFmla` |
| 12 | `to_str_with_var_sorts` / `fmla_to_str_ambiguous` | display helpers |

### Batch 1.4 — ivy_logic.py: Remaining (9 functions)
**Target:** `ivylogic/ivylogic.go`, `ivylogic/display.go`
**Complexity:** LOW-MEDIUM

| # | Python | Go Name |
|---|---|---|
| 1 | `typed_symbol` | `TypedSymbol` |
| 2 | `sym_decl_to_str` | `SymDeclToStr` |
| 3 | `typed_sym_to_str` | `TypedSymToStr` |
| 4 | `sig_to_str` | `SigToStr` |
| 5 | `pto` | `Pto` |
| 6 | `lambda_apply` | `LambdaApply` |
| 7 | `rename_vars_no_clash` | `RenameVarsNoClash` |
| 8 | `alpha_rename` | `AlphaRename` |
| 9 | `normalized_and` | `NormalizedAnd` |

### Batch 1.5 — ivy_logic_utils.py: Clause/Formula Conversion (20 functions)
**Target:** `clauseops/ops.go`
**Complexity:** MEDIUM

| # | Python | Go Name |
|---|---|---|
| 1 | `coerce_clause_to_formula` | `CoerceClauseToFormula` |
| 2 | `condition_conj` | `ConditionConj` |
| 3 | `formula_to_cube` | `FormulaToCube` |
| 4 | `formula_to_clauses_aux` | `FormulaToClausesAux` |
| 5 | `lit_to_formula` | `LitToFormula` |
| 6 | `cube_to_formula` | `CubeToFormula` |
| 7 | `clause_to_formula` | `ClauseToFormula` |
| 8 | `canonize_clause` | `CanonizeClause` |
| 9 | `trim_clauses` | `TrimClauses` |
| 10 | `rewrite_clause` | `RewriteClause` |
| 11 | `triv_fmla_to_lit` | `TrivFmlaToLit` |
| 12 | `triv_fmla_to_clause` | `TrivFmlaToClause` |
| 13 | `simplify_clause_fmla` | `SimplifyClauseFmla` |
| 14 | `simplify_clause` | `SimplifyClause` |
| 15 | `is_tautology` | `IsTautology` |
| 16 | `is_tautology_fmla` | `IsTautologyFmla` |
| 17 | `lit_in_clause` | `LitInClause` |
| 18 | `remove_duplicates` | `RemoveDuplicates` |
| 19 | `reduce_clauses` | `ReduceClauses` |
| 20 | `bool_const` | `BoolConst` |

### Batch 1.6 — ivy_logic_utils.py: Substitution and AST Traversal (15 functions)
**Target:** `logicutil/logic_utils.go`, `clauseops/astutil.go`
**Complexity:** MEDIUM-HIGH

| # | Python | Go Name |
|---|---|---|
| 1 | `substitute_ast` | `SubstituteAst` (variable subst, different from `SubstituteConstantsAST`) |
| 2 | `replace_temporals_by_named_binder_g_ast` | verify existing |
| 3 | `replace_named_binders_ast` | verify existing |
| 4 | `resort_symbol` | `ResortSymbol` |
| 5 | `resort_var` | `ResortVar` |
| 6 | `resort_sig` | `ResortSig` |
| 7 | `constants_ast` | `ConstantsAst` (different semantics from `UsedConstants`) |
| 8 | `apps_ast` | verify existing `AppsAst` |
| 9 | `ground_apps_ast_rec` | verify existing |
| 10 | `is_ground_clause` | `IsGroundClause` |
| 11 | `term_eq` | `TermEq` |
| 12 | `atom_eq` | `AtomEq` |
| 13 | `lit_eq` | `LitEq` |
| 14 | `used_variable_names_ast` | `UsedVariableNamesAst` |
| 15 | `variables_distinct_ast` + `rename_variables_distinct_asts` + `variables_distinct_list_ast` | 3 functions |

### Batch 1.7 — ivy_logic_utils.py: Subsumption and Advanced Clause Ops (20 functions)
**Target:** `clauseops/ops.go`
**Complexity:** HIGH

| # | Python | Go Name |
|---|---|---|
| 1 | `term_subsume` | `TermSubsume` |
| 2 | `lit_subsume` | `LitSubsume` |
| 3 | `atom_subsume` | `AtomSubsume` |
| 4 | `commute_lit` | `CommuteLit` |
| 5 | `clause_subsume_recur` | `ClauseSubsumeRecur` |
| 6 | `clause_subsume` | `ClauseSubsume` |
| 7 | `subsume` | `Subsume` |
| 8 | `distinct_variable_renaming` | `DistinctVariableRenaming` |
| 9 | `rename_variable` | `RenameVariable` |
| 10 | `is_individual_ast` | `IsIndividualAst` |
| 11 | `or_clauses2` | `OrClauses2` |
| 12 | `fix_or_annot` | `FixOrAnnot` |
| 13 | `elim_definitions` | `ElimDefinitions` |
| 14 | `rename_symbols` | `RenameSymbols` |
| 15 | `debug_clauses_list` | `DebugClausesList` |
| 16 | `tagged_or_clauses` | `TaggedOrClauses` |
| 17 | `find_true_disjunct` | `FindTrueDisjunct` |
| 18 | `eqcm_upd` | `EqcmUpd` |
| 19 | `exists_quant_clauses_map` | `ExistsQuantClausesMap` |
| 20 | `has_enumerated_sort` | `HasEnumeratedSort` |

### Batch 1.8 — ivy_logic_utils.py: Skolemization, Definitions, Parsing (20 functions)
**Target:** `clauseops/ops.go`
**Complexity:** HIGH

| # | Python | Go Name |
|---|---|---|
| 1 | `var_to_skolem` | `VarToSkolem` |
| 2 | `var_to_constant` | `VarToConstant` |
| 3 | `definition_instances` | `DefinitionInstances` |
| 4 | `unfold_definitions_clauses` | `UnfoldDefinitionsClauses` |
| 5 | `dual_formula` | `DualFormula` |
| 6 | `skolemize_formula` | `SkolemizeFormula` (clause-level) |
| 7 | `skolemize_ast` | `SkolemizeAst` |
| 8 | `witness_ast` | `WitnessAst` |
| 9 | `reskolemize_clauses` | `ReskolemizeClauses` |
| 10 | `unused_constant` | `UnusedConstant` |
| 11 | `apply_gen_to_clauses` | `ApplyGenToClauses` |
| 12 | `apply_func_to_clauses` | `ApplyFuncToClauses` |
| 13 | `to_formula` | `ToFormula` |
| 14 | `to_term` | `ToTerm` |
| 15 | `to_clause` | `ToClause` |
| 16 | `to_clauses` | `ToClauses` |
| 17 | `to_literal` | `ToLiteral` |
| 18 | `class LogicParseError` | `LogicParseError` error type |

---

## Phase 2: Solver Layer — ivy_solver.py (26 functions, 2 sessions)

### Batch 2.1 — Sort/Symbol Translation (13 functions)
**Target:** `solver/z3convert.go`, `solver/encoding.go`
**Complexity:** MEDIUM-HIGH

| # | Python | Go Name | Notes |
|---|---|---|---|
| 1 | `solver_name` | verify `Solver.SolverName` | |
| 2 | `my_minus` | `MyMinus` | |
| 3 | `my_eq` | `MyEq` | |
| 4 | `sort_name_to_z3` | `SortNameToZ3` | |
| 5 | `sorts` | `Sorts` | |
| 6 | `relations` | `Relations` | |
| 7 | `functions` | `Functions` | |
| 8 | `clear` | `Clear` | |
| 9 | `uninterpretedsort` | `UninterpretedSort` | |
| 10 | `functionsort` | `FunctionSort` | |
| 11 | `enumeratedsort` | `EnumeratedSort` | |
| 12 | `symbol_to_z3` | `SymbolToZ3` | |
| 13 | `apply_z3_func` | `ApplyZ3Func` | |

**Notes:** Many may already be subsumed by `z3bridge.Translator`. Verify before implementing.

### Batch 2.2 — Formula Translation and Model (13 functions)
**Target:** `solver/z3convert.go`, `solver/herbrand.go`
**Complexity:** MEDIUM-HIGH

| # | Python | Go Name | Notes |
|---|---|---|---|
| 1 | `term_to_z3` | `TermToZ3` | may be in Translator |
| 2 | `atom_to_z3` | `AtomToZ3` | may be in Translator |
| 3 | `forall` | `Forall` | verify Translator |
| 4 | `exists` | `Exists` | verify Translator |
| 5 | `clause_to_z3` | `ClauseToZ3` | |
| 6 | `conj_to_z3` | `ConjToZ3` | |
| 7 | `formula_to_z3_int` | `FormulaToZ3Int` | |
| 8 | `formula_to_z3_closed` | `FormulaToZ3Closed` | |
| 9 | `get_id` | `GetId` | |
| 10 | `get_model` | verify `Solver.GetModelClauses` | |
| 11 | `sort_from_z3` | verify `Z3SortToSort` | |
| 12 | `constant_from_z3` | verify in `herbrand.go` | |
| 13 | `get_model_constant` | verify in `herbrand.go` | |
| 14 | `get_lit_facts` | verify `getLitFacts` | |
| 15 | `gebin` | `Gebin` | |
| 16 | `substitute` | verify `SubstituteZ3` | |

---

## Phase 3: Module + Actions (26 functions, 2 sessions)

### Batch 3.1 — ivy_module.py: Resort and Context (6 functions)
**Target:** `module/resort.go`, `module/context.go`
**Complexity:** MEDIUM

| # | Python | Go Name |
|---|---|---|
| 1 | `resort_clauses` | `ResortClauses` |
| 2 | `resort_asts` | `ResortAsts` |
| 3 | `remove_refined_sortnames_from_set` | `RemoveRefinedSortnamesFromSet` |
| 4 | `remove_refined_sortnames_from_list` | `RemoveRefinedSortnamesFromList` |
| 5 | `resort_map_any_ast` | `ResortMapAnyAst` |
| 6 | `class ModuleTheoryContext` | `ModuleTheoryContext` struct |

### Batch 3.2 — ivy_actions.py: Type Checking and Annotation (20 functions)
**Target:** `actions/helpers.go`, `actions/annotation.go`, new `actions/typecheck.go`
**Complexity:** MEDIUM-HIGH

| # | Python | Go Name |
|---|---|---|
| 1 | `p_c_a` | `PCA` |
| 2 | `class UnrollContext` | `UnrollContext` struct |
| 3 | `class SymbolList` | `SymbolList` struct |
| 4 | `get_correct_arity` | `GetCorrectArity` |
| 5 | `type_check` | `TypeCheck` |
| 6 | `type_ast` | `TypeAst` |
| 7 | `destr_asgn_val` | `DestrAsgnVal` |
| 8 | `assign_refs` | `AssignRefs` |
| 9 | `sign` | `Sign` |
| 10 | `make_field_update` | `MakeFieldUpdate` |
| 11 | `my_str` | `MyStr` |
| 12 | `set_determinize` | `SetDeterminize` |
| 13 | `bracket_action` | `BracketAction` |
| 14 | `class DebugAction` | `DebugAction` struct |
| 15 | `entry` | `Entry` |
| 16 | `class TypeCheckContext` | `TypeCheckContext` struct |
| 17 | `conj_annot` | `ConjAnnot` |
| 18 | `compose_annot` | `ComposeAnnot` |
| 19 | `rename_annot` | `RenameAnnot` |
| 20 | `ite_annot` | `IteAnnot` |

---

## Phase 4: Transition Relations + Interpretation (31 functions, 2 sessions)

### Batch 4.1 — ivy_transrel.py (16 functions)
**Target:** `transrel/transrel.go`
**Complexity:** MEDIUM-HIGH

| # | Python | Go Name |
|---|---|---|
| 1 | `rename` | `Rename` |
| 2 | `update_frame_constraint` | `UpdateFrameConstraint` |
| 3 | `symbol_frame_cond` | `SymbolFrameCond` |
| 4 | `join` | `Join` |
| 5 | `ite` | `Ite` |
| 6 | `clauses_imply_formula_cex` | `ClausesImplyFormulaCex` |
| 7 | `implies` | `Implies` |
| 8 | `implies_state` | `ImpliesState` |
| 9 | `implies_action` | `ImpliesAction` |
| 10 | `my_annot_op` | `MyAnnotOp` |
| 11 | `clausify` | `Clausify` |
| 12 | `clausify_state` | `ClausifyState` |
| 13 | `remove_taut_eqs_clauses` | `RemoveTautEqsClauses` |
| 14 | `extract_pre_post_model` | `ExtractPrePostModel` |
| 15 | `small_model_clauses` | `SmallModelClauses` |
| 16 | `use_numerals` | `UseNumerals` |

### Batch 4.2 — ivy_interp.py (15 functions)
**Target:** `interp/interp.go`, `interp/eval.go`
**Complexity:** MEDIUM

| # | Python | Go Name |
|---|---|---|
| 1 | `type_check_list` | `TypeCheckList` |
| 2 | `module_order` | `ModuleOrder` |
| 3 | `module_skolemizer` | `ModuleSkolemizer` |
| 4 | `get_core` | `GetCore` |
| 5 | `reverse_join_concrete_clauses` | `ReverseJoinConcreteClauses` |
| 6 | `underapproximate_state` | `UnderapproximateState` |
| 7 | `states_state_expr` | `StatesStateExpr` |
| 8 | `decompose_action_app` | `DecomposeActionApp` |
| 9 | `state_implies_formula` | `StateImpliesFormula` |
| 10 | `new_history` | `NewHistory` |
| 11 | `eval_assert_rhs` | `EvalAssertRhs` |
| 12 | `eval_state_order` | `EvalStateOrder` |
| 13 | `check_state_assertion` | `CheckStateAssertion` |
| 14 | `get_state_assertions` | `GetStateAssertions` |
| 15 | `universe_constraint` | `UniverseConstraint` |

---

## Phase 5: Proof System — ivy_proof.py (60 functions, 4 sessions)

### Batch 5.1 — Goal Utilities (15 functions)
**Target:** `proof/goal.go`
**Complexity:** MEDIUM

| # | Python | Go Name |
|---|---|---|
| 1 | `attrib_goals` | `AttribGoals` |
| 2 | `is_goal` | `IsGoal` |
| 3 | `goals_subst` | `GoalsSubst` |
| 4 | `fresh_label` | `FreshLabel` |
| 5 | `goal_prefix_prems` | `GoalPrefixPrems` |
| 6 | `check_name_clash` | `CheckNameClash` |
| 7 | `check_premises_provided` | `CheckPremisesProvided` |
| 8 | `goal_is_temporal` | `GoalIsTemporal` |
| 9 | `goal_defines` | `GoalDefines` |
| 10 | `get_unprovided_defns` | `GetUnprovidedDefns` |
| 11 | `goal_subgoals` | `GoalSubgoals` |
| 12 | `fmla_vocab` | `FmlaVocab` |
| 13 | `check_schema_capture` | `CheckSchemaCapture` |
| 14 | `check_alpha_capture` | `CheckAlphaCapture` |
| 15 | `check_renaming` | `CheckRenaming` |

### Batch 5.2 — Match Compilation (15 functions)
**Target:** `proof/matching.go`
**Complexity:** HIGH

| # | Python | Go Name |
|---|---|---|
| 1 | `compile_expr_vocab` | `CompileExprVocab` |
| 2 | `compile_expr_vocab_ext` | `CompileExprVocabExt` |
| 3 | `remove_vars_match` | `RemoveVarsMatch` |
| 4 | `show_match` | `ShowMatch` |
| 5 | `transform_defn_schema` | `TransformDefnSchema` |
| 6 | `transform_defn_match` | `TransformDefnMatch` |
| 7 | `goal_prems_by_name` | `GoalPremsByName` |
| 8 | `add_prem_match` | `AddPremMatch` |
| 9 | `parameterize_schema` | `ParameterizeSchema` |
| 10 | `compile_match_list` | `CompileMatchList` |
| 11 | `compile_one_match` | `CompileOneMatch` |
| 12 | `compile_match` | `CompileMatch` |
| 13 | `match_rhs_vars` | `MatchRhsVars` |
| 14 | `is_lambda` | `IsLambda` |
| 15 | `goal_is_property` / `goal_is_schema` | `GoalIsProperty` / `GoalIsSchema` |

### Batch 5.3 — Match Application (15 functions)
**Target:** `proof/match.go`
**Complexity:** HIGH

| # | Python | Go Name |
|---|---|---|
| 1 | `apply_match_match` | `ApplyMatchMatch` |
| 2 | `rename_problem` | `RenameProblem` |
| 3 | `avoid_capture_problem` | `AvoidCaptureProblem` |
| 4 | `goals_eq_mod_alpha` | `GoalsEqModAlpha` |
| 5 | `apply_match_rec` | `ApplyMatchRec` |
| 6 | `raise_capture` | `RaiseCapture` |
| 7 | `match_get` | `MatchGet` |
| 8 | `apply_match_alt` | `ApplyMatchAlt` |
| 9 | `apply_fun` | `ApplyFun` |
| 10 | `apply_match_alt_rec` | `ApplyMatchAltRec` |
| 11 | `apply_match_func_alt` | `ApplyMatchFuncAlt` |
| 12 | `apply_match_sort` | `ApplyMatchSort` |
| 13 | `apply_match_freesyms_alt` | `ApplyMatchFreesymsAlt` |
| 14 | `rename_goal` | `RenameGoal` |
| 15 | `make_distinct_vars` | `MakeDistinctVars` |

### Batch 5.4 — Tactics and Remaining (15 functions)
**Target:** `proof/tactics.go`, `proof/goal.go`
**Complexity:** MEDIUM-HIGH

| # | Python | Go Name |
|---|---|---|
| 1 | `var_subst_goal` | `VarSubstGoal` |
| 2 | `apply_to_conc` | `ApplyToConc` |
| 3 | `compile_witness_list` | `CompileWitnessList` |
| 4 | `remove_unused_definitions_goal` | `RemoveUnusedDefinitionsGoal` |
| 5 | `match_from_defn` | `MatchFromDefn` |
| 6 | `match_from_defns` | `MatchFromDefns` |
| 7 | `unfold_goal` | `UnfoldGoal` |
| 8 | `unfold_fmla` | `UnfoldFmla` |
| 9 | `goal_apply_to_prem` | `GoalApplyToPrem` |
| 10 | `goal_apply_to_conc` | `GoalApplyToConc` |
| 11 | `close_unmatched` | `CloseUnmatched` |
| 12 | `drop_supplied_prems` | `DropSuppliedPrems` |
| 13 | `remove_explicit` | `RemoveExplicit` |
| 14 | `rename_prem_no_clash` | `RenamePremNoClash` |

---

## Phase 6: Compiler — ivy_compiler.py (73 functions, 5-6 sessions)

### Batch 6.1 — Compilation Contexts and Sort Inference (15 functions)
**Target:** `compiler/compiler.go`, `compiler/helpers.go`
**Complexity:** MEDIUM

| # | Python | Go Name |
|---|---|---|
| 1 | `class Context` | `Context` struct (verify `ExprContext`) |
| 2 | `class EmptyVariableContext` | `EmptyVariableContext` struct |
| 3 | `thing` / `other_thing` | `Thing` / `OtherThing` |
| 4 | `compile_root_args` | `CompileRootArgs` |
| 5 | `compile_args` | verify existing |
| 6 | `sort_infer_covariant` | `SortInferCovariant` |
| 7 | `sort_infer_contravariant` | `SortInferContravariant` |
| 8 | `old_sym` | `OldSym` |
| 9 | `compile_isa` | `CompileIsa` |
| 10 | `cquant` | `Cquant` |
| 11 | `UpdatePattern_cmpl` | `UpdatePatternCmpl` |
| 12 | `ConstantDecl_cmpl` | `ConstantDeclCmpl` |
| 13 | `Old_cmpl` | `OldCmpl` |
| 14 | `get_arg_sorts` | `GetArgSorts` |
| 15 | `get_relation_sort` | `GetRelationSort` |

### Batch 6.2 — Action Compilation (12 functions)
**Target:** `compiler/action.go`
**Complexity:** MEDIUM

| # | Python | Go Name |
|---|---|---|
| 1 | `sortify` | `Sortify` |
| 2 | `compile_assign_lhs` | `CompileAssignLhs` |
| 3 | `compile_crash_action` | `CompileCrashAction` |
| 4 | `compile_thunk_action` | `CompileThunkAction` |
| 5 | `compile_debug_action` | `CompileDebugAction` |
| 6 | `compile_native_arg` | `CompileNativeArg` |
| 7 | `compile_native_symbol` | `CompileNativeSymbol` |
| 8 | `compile_native_action` | `CompileNativeAction` |
| 9 | `compile_native_name` | `CompileNativeName` |
| 10 | `compile_native_def` | `CompileNativeDef` |
| 11 | `compile_native_type` | `CompileNativeType` |
| 12 | `resolve_alias_int` | verify existing |

### Batch 6.3 — Schema and Tactic Compilation (18 functions)
**Target:** `compiler/compiler.go`
**Complexity:** HIGH

| # | Python | Go Name |
|---|---|---|
| 1 | `compile_schema_prem` | `CompileSchemaPrem` |
| 2 | `compile_schema_conc` | `CompileSchemaConc` |
| 3 | `compile_schema_body` | `CompileSchemaBody` |
| 4 | `lookup_schema` | verify existing |
| 5 | `compile_schema_instantiation` | `CompileSchemaInstantiation` |
| 6 | `compile_let_tactic` | `CompileLetTactic` |
| 7 | `compile_witness_tactic` | `CompileWitnessTactic` |
| 8 | `compile_unfold_tactic` | `CompileUnfoldTactic` |
| 9 | `compile_forget_tactic` | `CompileForgetTactic` |
| 10 | `compile_if_tactic` | `CompileIfTactic` |
| 11 | `compile_property_tactic` | `CompilePropertyTactic` |
| 12 | `compile_function_tactic` | `CompileFunctionTactic` |
| 13 | `compile_proof_tactic` | `CompileProofTactic` |
| 14 | `infer_parameters` | `InferParameters` |
| 15 | `check_instantiations` | `CheckInstantiations` |
| 16 | `tarjan_arcs` | `TarjanArcs` |
| 17 | `get_symbol_dependencies` | `GetSymbolDependencies` |
| 18 | `prop_to_def` / `reorder_props` | `PropToDef` / `ReorderProps` |

### Batch 6.4 — DomainSetup/ConjSetup Extensions (13 functions)
**Target:** `compiler/decl.go`, `compiler/ivy_compile.go`
**Complexity:** HIGH

| # | Python | Go Name |
|---|---|---|
| 1 | `class IvyDomainSetup` extensions | extend `DomainSetup` |
| 2 | `class IvyConjectureSetup` extensions | extend `ConjSetup` |
| 3 | `check_is_action` | `CheckIsAction` |
| 4 | `BalancedChoice` | `BalancedChoice` |
| 5 | `get_file_version` | `GetFileVersion` |
| 6 | `ivy_new` | `IvyNew` |
| 7 | `apply_assert_proof` | `ApplyAssertProof` |
| 8 | `apply_assert_proofs` | `ApplyAssertProofs` |
| 9 | `check_properties` | `CheckProperties` (84 lines — largest single function) |
| 10 | `set_verifying` | `SetVerifying` |
| 11 | `ivy_compile_theory` | `IvyCompileTheory` |
| 12 | `compile_theory` | `CompileTheory` |
| 13 | `compile_theories` | `CompileTheories` |

### Batch 6.5 — Module Loading and Remaining (9 functions)
**Target:** `compiler/ivy_compile.go`
**Complexity:** HIGH

| # | Python | Go Name |
|---|---|---|
| 1 | `add_labels_to_proof` | `AddLabelsToProof` |
| 2 | `add_action_label` | `AddActionLabel` |
| 3 | `show_call_graph` | `ShowCallGraph` |
| 4 | `clear_rules` | `ClearRules` |
| 5 | `read_module` | `ReadModule` |
| 6 | `import_module` | `ImportModule` |
| 7 | `ivy_load_file` | `IvyLoadFile` |
| 8 | `ivy_from_string` | `IvyFromString` |
| 9 | `ivy_compile_theory_from_string` | `IvyCompileTheoryFromString` |

---

## Phase 7: Isolate, MC, Check, Art, Fragment (32 functions, 2 sessions)

### Batch 7.1 — ivy_isolate.py (15 functions)
**Target:** `isolate/helpers.go`, `isolate/strip.go`
**Complexity:** MEDIUM

| # | Python | Go Name |
|---|---|---|
| 1 | `startswith_some_rec` | `StartsWithSomeRec` |
| 2 | `startswith_eq_some_rec` | `StartsWithEqSomeRec` |
| 3 | `get_strip_params` | `GetStripParams` |
| 4 | `strip_native` | `StripNative` |
| 5 | `strip_natives` | `StripNatives` |
| 6 | `has_side_effect_rec` | `HasSideEffectRec` |
| 7 | `get_props_proved_in_isolate_orig` | `GetPropsProvedInIsolateOrig` |
| 8 | `follow_definitions_rec` | `FollowDefinitionsRec` |
| 9 | `collect_relevant_destructors` | `CollectRelevantDestructors` |
| 10 | `class SortOrder` | `SortOrder` struct |
| 11 | `get_cone` | `GetCone` |
| 12 | `get_mod_cone` | `GetModCone` |
| 13 | `conj_to_assume` | `ConjToAssume` |
| 14 | `find_some_assertion` | `FindSomeAssertion` |
| 15 | `find_some_call` | `FindSomeCall` |

### Batch 7.2 — ivy_mc.py (10) + ivy_check.py (6) + ivy_art.py (1) + ivy_fragment.py (1)
**Target:** `mc/`, `check/`, `art/`, `fragment/`
**Complexity:** MEDIUM

**ivy_mc.py:**

| # | Python | Go Name |
|---|---|---|
| 1 | `encode_vars` | `EncodeVars` |
| 2 | `clone_normal` | `CloneNormal` |
| 3 | `uncompose_annot` | `UncomposeAnnot` |
| 4 | `unite_annot` | `UniteAnnot` |
| 5 | `class MatchHandler` | `MatchHandler` struct |
| 6 | `match_annotation` | `MatchAnnotation` |
| 7 | `checked` / `badwit` | `Checked` / `Badwit` |
| 8 | `class IvyMCTrace` | `IvyMCTrace` struct |
| 9 | `aiger_witness_to_ivy_trace` | `AigerWitnessToIvyTrace` |

**ivy_check.py:**

| # | Python | Go Name |
|---|---|---|
| 10 | `gui_art` | `GuiArt` |
| 11 | `usage` | `Usage` |
| 12 | `show_assertions` | `ShowAssertions` |
| 13 | `print_dots` | `PrintDots` |
| 14 | `filter_fcs` | `FilterFcs` |
| 15 | `info` | `Info` |

**ivy_art.py:**

| # | Python | Go Name |
|---|---|---|
| 16 | `render_rg` | `RenderRg` |

**ivy_fragment.py:**

| # | Python | Go Name |
|---|---|---|
| 17 | `show_strat_graph` | `ShowStratGraph` |

---

## Summary

| Phase | Batches | Functions | Sessions | Complexity |
|---|---|---|---|---|
| 1: Foundation (logic + utils) | 1.1–1.8 | ~140 | 6-7 | Medium-High |
| 2: Solver | 2.1–2.2 | ~26 | 2 | Medium-High |
| 3: Module + Actions | 3.1–3.2 | ~26 | 2 | Medium |
| 4: Transrel + Interp | 4.1–4.2 | ~31 | 2 | Medium-High |
| 5: Proof | 5.1–5.4 | ~60 | 4 | High |
| 6: Compiler | 6.1–6.5 | ~73 | 5-6 | High |
| 7: Isolate + MC + misc | 7.1–7.2 | ~32 | 2 | Medium |
| **TOTAL** | **24** | **~388** | **~18-20** | |

---

## Key Risks

1. **Python global state vs Go explicit state:** Python uses global `sig`, `mod`. Go passes `*Sig`, `*Module` explicitly. Each ported function must identify its global deps.

2. **Structural equality:** Python `__eq__` on AST nodes → Go `lg.Key()` and `lg.NodeKey` maps.

3. **Context managers:** Python `with` → Go Enter/Exit pattern (already established).

4. **Solver architecture:** Python standalone functions may already be in `z3bridge.Translator` or `Solver` methods. Verify before duplicating.

5. **Circular imports:** Go packages can't have circular deps. Careful placement of new functions needed.

---

## Verification

After each batch:
1. `go build ./...` — must compile clean
2. `go vet ./...` — must pass
3. Run existing tests: `go test ./packagename/...`
4. For solver-touching code: `DYLD_LIBRARY_PATH=/Users/jaten/go/src/github.com/glycerine/goivy/z3ivy/lib:$DYLD_LIBRARY_PATH go test -v ./solver/... -count=1`
5. For web UI: `make test-web`

After all phases:
- Full test suite: `go test ./... -count=1`
- Golden AST comparison tests in `compiler/golden_ast_test.go`
- Update UNPORTED.md to reflect 100% completion
