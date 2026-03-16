# Audit: Python Ivy → Go Port Variance Checklist

Generated: 2026-03-16

Python source: `~/pyivy/ivy/ivy/*.py`
Go source: `/Users/jaten/go/src/github.com/glycerine/goivy/`

---

## A. Entirely Missing Python Modules (no Go counterpart)

### A1. CRITICAL — Core verification logic

- [~] **ivy_l2s.py** (1522 lines) — PARTIAL: Skeleton in `l2s/l2s.go`. Entry points `L2STactic`, `L2STacticFull`, `L2STacticAuto` defined. Helper functions for temporal normalization, saved copies, l2s_g-to-Globally conversion implemented. Full monitor construction deferred pending proof infrastructure completion (requires ProofGoal, TemporalModels, compiler integration).

- [x] **ivy_fragment.py** (591 lines) — DONE: Ported to `fragment/fragment.go`. Implements stratification graph construction per Ge & de Moura, FEU fragment checking, macro maps, Skolem simulation, cycle detection. Also created `unionfind/unionfind.go` (rank-based union-find with path compression).

- [x] **ivy_auto_inst.py** (354 lines) — DONE: Ported to `autoinst/autoinst.go`. Includes `Match` class, `ApplyMatch()`, `PatternMatch()`, `TriggerMatches()`, `InstantiateAxioms()`, `MergeMatchLists()`, `Normalize()`, `TermOrd()`. Schema expansion (`expand_schemata`) requires schema infrastructure.

- [x] **ivy_alpha.py** (350 lines) — DONE: Ported to `alpha/alpha.go`. Includes `Alpha()`, `PredicateAlpha()`, `ProgressiveDomain`, `RelAlg1/2/3` for relational algebra on abstract states.

- [~] **ivy_compose.py** (190 lines) — PARTIAL: Skeleton in `compose/compose.go`. Entry point `ComposeTactic`, `RankingDef`, `CreateRankingDefn`, `ValidateRankingDef` defined. Full implementation deferred pending proof infrastructure.

- [x] **ivy_vmt.py** (289 lines) — DONE: Ported to `vmt/vmt.go`. Includes `CheckIsolateVMT()`, `WriteVMT()`, array encoding, `UFToArrayAction()`, `ArrayEncodeFormula()`.

### A2. IMPORTANT — Supporting infrastructure

- [ ] **ivy_logic_parser.py** (655 lines) — Standalone formula parser using PLY. Parses formulas from strings (used by `to_formula()`, `to_clause()`, etc. in ivy_logic_utils.py). The Go parser/ package handles file parsing but may not expose a string-to-formula API.

- [ ] **ivy_ev_parser.py** (443 lines) — Event trace parser. Parses event traces for counterexample display. `Events`, `Event`, `EventParser` classes with full recursive descent parser.

- [ ] **ivy_concept_space.py** (210 lines) — Concept space parser and AST. `NamedSpace`, `SumSpace`, `ProductSpace` classes, PLY grammar for concept spaces, `to_concept_space()`, `clauses_to_concept()`.

- [x] **ivy_union_find.py** (50 lines) — DONE: Ported to `unionfind/unionfind.go`. Combined v1 and v2 into single rank-based implementation.

- [x] **ivy_union_find2.py** (64 lines) — DONE: Merged into `unionfind/unionfind.go`.

- [x] **proof.py** (176 lines) — DONE: Ported to `proof/proofstate.go`. All classes implemented: `ProofGoal`, `ProofGoalStack`, `ReachabilityGraph`, `ReachabilityNode`, `ReachabilityEdge`, `AbstractState`, `ConcreteState`, `ProofManager`. The proof package already had `Vocab`, `GoalConc`, `GoalPrems`, `ProofChecker`, `MatchProblem`.

- [ ] **tactics.py** (302 lines) — Interactive refinement tactics. `RemoveIfRefuted`, `RemoveGoal`, `RefineOrReverse`, `CustomRefineOrReverse`, `PathReach`, `PathReach1`, `PushDiagram`, `RecalculateFacts`, `RemoveFacts`, `ExecuteAction`, `PushNewGoal`, `CheckCover`, `Join2`, `UPDR` tactic classes. Used by the interactive UI.

- [ ] **tactics_api.py** (473 lines) — Tactics API used by UI. `forward_image()`, `backward_image()`, `refine_or_reverse()`, `implied_facts()`, `get_diagram()`, `refuted_goal()`, `push_goal()`, `top_goal()`, `remove_goal()`, `Abstractors` class. Core API for interactive proof exploration.

- [ ] **z3_utils.py** (197 lines) — Z3 utility functions for sort/symbol management.

- [ ] **logic_util.py** (335 lines) — Legacy logic utilities (partially ported to logicutil/ but not fully).

- [x] **ivy_acl.py** (111 lines) — DONE: Ported to `acl/acl.go`. Full implementation including `RegisterIgnores()`, `RegisterAssumes()`, `IsIgnored()`, `IsAssumed()`, regex support.

### A3. MINOR — Infrastructure / entry points

- [ ] **ivy_libs.py** (70 lines) — Library file loading.
- [ ] **ivy_init.py** (113 lines) — Initialization, parameter reading, `ivy_init()`.
- [ ] **ivy_launch.py** (221 lines) — Process launching for test/build.
- [ ] **ivy_dump.py** (60 lines) — Debug dump utilities.
- [ ] **ivy_graphviz.py** (197 lines) — Graphviz DOT output generation.
- [x] **ivy_smtlib.py** (30 lines) — DONE: `quantifiers_decidable` inlined in ivylogic/globals.go. SMT-LIB Theory/Sort types not needed (Z3 bridge handles this directly).
- [ ] **ivy_shell.py** (12 lines) — Shell/REPL entry point.
- [ ] **ivy_lsp.py** (19 lines) — Language Server Protocol stub.
- [ ] **iupdr.py** (203 lines) — Interactive UPDR (IPython widget-based, may not be needed).
- [x] **general.py** (12 lines) — DONE: `IvyError` already exists in logic/error.go.
- [ ] **dot_layout.py** (286 lines) — DOT graph layout algorithms.
- [ ] **concept.py** (865 lines) — Concept graph model. Partially ported to webui/concept.go but needs audit for completeness.

---

## B. Missing Functions Within Ported Modules

### B1. solver/ (vs ivy_solver.py, 1716 lines)

- [x] `binary_interpolant(clauses2, clauses1)` — DONE: `BinaryInterpolant()` in solver/z3convert.go. Includes fallback error for builds without Z3 interpolation API.
- [x] `HerbrandModel` class (~80 lines) — DONE: Ported to `solver/herbrand.go`. Full implementation including `Sorts()`, `SortUniverse()`, `Eval()`, `EvalConstant()`, `EvalToConstant()`, `Check()`. Also added `Model.Sorts()`, `Model.SortUniverse()`, `Sort.String()` to z3bridge.
- [x] `sort_from_z3(s)` — DONE: Handled via z3bridge Sort.String() and sig lookup in HerbrandModel.
- [x] `constant_from_z3(sort, c)` — DONE: `constantFromZ3()` in solver/herbrand.go.
- [x] `get_model_constant(m, t)` — DONE: `HerbrandModel.getModelConstant()` in solver/herbrand.go.
- [x] `model_universe_facts(h, sort, upclose)` — DONE: `ModelUniverseFacts()` in solver/herbrand.go.
- [x] `model_facts(h, ignore, clauses1, upclose)` — DONE: `ModelFacts()` in solver/herbrand.go.
- [x] `relation_model_to_clauses(h, r, n)` — DONE: `RelationModelToClauses()` in solver/herbrand.go.
- [x] `function_model_to_clauses(h, f)` — DONE: `FunctionModelToClauses()` in solver/herbrand.go.
- [x] `get_lit_facts(h, lit, res)` — DONE: `getLitFacts()` in solver/herbrand.go.
- [x] `numeral_assign(clauses, h)` — DONE: `NumeralAssign()` in solver/compat.go.
- [x] `clauses_case(clauses1)` — DONE: `Solver.ClausesCase()` in solver/herbrand.go.
- [x] `clause_model_simp(m, c)` — DONE: `Solver.ClauseModelSimp()` in solver/compat.go.
- [x] `mine_interpreted_constants(model, vocab)` — DONE: `Solver.MineInterpretedConstants()` in solver/compat.go.
- [x] `enumerated_range(sort)` — DONE: `Solver.EnumeratedRange()` in solver/herbrand.go.
- [x] `collect_numerals(z3term)` — DONE: `CollectNumeralsRecursive()` in solver/z3convert.go.
- [x] `from_z3_numeral(z3term, sort)` — DONE: `FromZ3Numeral()` in solver/z3convert.go.
- [x] `collect_model_values(sort, model, sym)` — DONE: `CollectModelValuesZ3()` in solver/z3convert.go.
- [x] `SortOrder` class — DONE: `SortOrder` struct + `NewSortOrder()` + `Compare()` in solver/z3convert.go.
- [x] `encode_term(t, n, sort)` — DONE: `EncodeTerm()` in solver/encoding.go.
- [x] `encode_equality(*terms)` — DONE: `EncodeEquality()` in solver/encoding.go.
- [x] `binenc(m, n)`, `gebin(bits, n)`, `ceillog2(n)` — DONE: `BinEnc()`, `GetBin()`, `CeilLog2()` in solver/encoding.go.
- [x] `z3_function(name, sig)` — DONE: `Solver.Z3Function()` in solver/encoding.go.
- [x] `z3_to_formula(z3expr, vars)` — DONE: `Z3ToFormula()` in solver/z3convert.go.
- [x] `z3sort_to_sort(z3sort)` — DONE: `Z3SortToSort()` in solver/z3convert.go.
- [x] `z3decl_to_symbol(z3decl)` — DONE: `Z3DeclToSymbol()` in solver/z3convert.go.
- [x] `terms_match(tl1, tl2)` — DONE: `TermsMatch()` in solver/compat.go.
- [x] `get_arg_range(m, x)` — DONE: `Solver.GetArgRange()` in solver/compat.go.
- [x] `model_if_none(clauses1, implied, model)` — DONE: `Solver.ModelIfNone()` in solver/compat.go.
- [x] `substitute(t, *m)` — DONE: `SubstituteZ3()` in solver/z3convert.go. Wraps `Context.Substitute`.
- [x] `set_seed(seed)` — DONE: `Solver.SetSeed()` in solver/encoding.go.
- [x] `set_macro_finder(truth)` — DONE: `Solver.SetMacroFinder()` in solver/encoding.go.
- [x] `set_use_native_enums(t)` — DONE: `Solver.SetUseNativeEnums()` in solver/encoding.go.
- [x] `parse_array_theory(name)` — DONE: `ParseArrayTheory()` in solver/encoding.go.
- [x] `parse_int_params(name)` — DONE: `ParseIntParams()` in solver/encoding.go.
- [x] `is_solver_sort(name)` — DONE: `IsSolverSort()` in solver/encoding.go.
- [x] `is_solver_op(name)` — DONE: `IsSolverOp()` in solver/encoding.go.
- [x] `check_native_compat_sym(sym)` — DONE: `CheckNativeCompatSym()` in solver/compat.go.
- [x] `check_compat()` — DONE: `CheckCompat()` in solver/compat.go.
- [x] `sort_card(sort)` — DONE: `SortCard()` in solver/herbrand.go.
- [x] `native_symbol(sym)` — DONE: `NativeSymbol()` in solver/encoding.go.
- [x] `lt_pred(sort)` — DONE: `LtPred()` in solver/encoding.go.
- [x] `get_polymacs(op)` — DONE: `GetPolymacs()` in solver/compat.go.
- [x] `numeral_to_z3(num)` — DONE: `Solver.NumeralToZ3()` in solver/encoding.go.
- [x] `enumerated_to_numeral(term)` — DONE: `EnumeratedToNumeral()` in solver/encoding.go.
- [x] `quant_constraints(vs, z3_vs)` — DONE: `QuantConstraints()` in solver/compat.go.
- [x] `range_sort_bounds_to_z3(itp)` — DONE: `Solver.RangeSortBoundsToZ3()` in solver/z3convert.go.
- [x] `type_constraints(syms)` — DONE: `TypeConstraints()` in solver/compat.go.

### B2. logicutil/ + clauseops/ (vs ivy_logic_utils.py, 1635 lines)

- [x] `close_epr(fmla)` — DONE: `CloseEPR()` in logicutil/logic_utils.go.
- [x] `normalize_named_binders(ast, names)` — DONE: `NormalizeNamedBinders()` in logicutil/logic_utils.go.
- [x] `replace_temporals_by_named_binder_g_ast(ast, g, when)` — DONE: `ReplaceTemporalsByNamedBinder()` in logicutil/logic_utils.go.
- [x] `reduce_named_binders(ast, g)` — DONE: `ReduceNamedBinders()` in logicutil/logic_utils.go.
- [x] `replace_named_binders_ast(ast, subs)` — DONE: `ReplaceNamedBindersAst()` in logicutil/logic_utils.go.
- [x] `expand_named_binders_ast(ast, fun)` — DONE: `ExpandNamedBindersAst()` in logicutil/logic_utils.go.
- [x] `denormalize_temporal(ast)` — DONE: `DenormalizeTemporal()` in logicutil/logic_utils.go.
- [x] `resort_sort(sort, subs)` — DONE: `ResortSort()` in logicutil/logic_utils.go.
- [x] `resort_symbol(sym, subs)` — DONE: Handled via ResortAst for Const nodes.
- [x] `resort_var(sym, subs)` — DONE: Handled via ResortAst for Var nodes.
- [x] `resort_ast(ast, subs)` — DONE: `ResortAst()` in logicutil/logic_utils.go.
- [x] `resort_sig(subs)` — DONE: `ResortSig()` in module/resort.go.
- [x] `rename_clauses_annot_fun(annot, map)` — DONE: `RenameClausesAnnotFun()` in logicutil/logic_utils.go.
- [x] `named_binders_ast(ast)` — DONE: `NamedBindersAst()` in logicutil/logic_utils.go.
- [x] `temporals_ast(ast)` — DONE: `TemporalsAst()` in logicutil/logic_utils.go.
- [x] `sorts_ast(ast)` — DONE: `SortsAst()` in logicutil/logic_utils.go.
- [x] `relations_ast(ast)` — DONE: `RelationsAst()` in logicutil/logic_utils.go.
- [x] `functions_ast(ast)` — DONE: `FunctionsAst()` in logicutil/logic_utils.go.
- [x] `apps_ast(ast)` — DONE: `AppsAst()` in ivylogic/globals.go.
- [x] `ground_apps_ast(ast)` — DONE: `GroundAppsAst()` in logicutil/logic_utils.go.
- [x] `eqs_ast(ast)` — DONE: `EqsAst()` in logicutil/logic_utils.go.
- [x] `is_equality_lit(lit)` — DONE: `IsEqualityLit()` in logicutil/logic_utils.go.
- [x] `is_taut_equality_lit(lit)` — DONE: `IsTautEqualityLit()` in logicutil/logic_utils.go.
- [x] `is_vac_equality_lit(lit)` — DONE: `IsVacEqualityLit()` in logicutil/logic_utils.go.
- [x] `is_true_lit(lit)`, `is_false_lit(lit)` — DONE: `IsTrueLit()`, `IsFalseLit()` in logicutil/logic_utils.go.
- [x] `is_taut_lit(lit)`, `is_vac_lit(lit)` — DONE: `IsTautLit()`, `IsVacLit()` in logicutil/logic_utils.go.
- [x] `is_disequality_lit(lit)` — DONE: `IsDisequalityLit()` in logicutil/logic_utils.go.
- [x] `is_ground_clause(clause)` — DONE: `IsGroundLit()` in logicutil/logic_utils.go.
- [x] `is_ground_equality_lit(lit)` — DONE: `IsGroundEqualityLit()` in logicutil/logic_utils.go.
- [x] `term_eq(t1, t2)` — Already provided by lg.Node.Equal().
- [x] `term_lists_eq(l1, l2)` — DONE: `TermListsEq()` in logicutil/logic_utils.go.
- [x] `atom_eq(at1, at2)` — Already provided by lg.Node.Equal().
- [x] `lit_eq(lit1, lit2)` — Already provided by Literal.Equal().
- [x] `swap_args_lit(lit)` — DONE: `SwapArgsLit()` in logicutil/logic_utils.go.
- [x] `eq_lit(x, y)` — DONE: `EqLit()` in logicutil/logic_utils.go.
- [x] `eq_atom(x, y)` — DONE: `EqAtom()` in logicutil/logic_utils.go.
- [x] `rel_inst(relname)` — DONE: `RelInst()` in logicutil/logic_utils.go.
- [x] `fun_inst(funname)` — DONE: `FunInst()` in logicutil/logic_utils.go.
- [x] `fun_eq_inst(funname)` — DONE: `FunEqInst()` in logicutil/logic_utils.go.
- [x] `is_relational(sym)` — DONE: `IsRelational()` in logicutil/logic_utils.go.
- [x] `TseitinContext` class + `tseitin_encoding(f)` — DONE: `TseitinContext` + `TseitinEncoding()` in logicutil/logic_utils.go.
- [x] `expand_abbrevs(f)` — DONE: `ExpandAbbrevs()` in logicutil/logic_utils.go.
- [x] `formula_to_lit(f)` — DONE: `FormulaToLit()` in logicutil/logic_utils.go.
- [x] `formula_to_clause(f)` — DONE: `FormulaToClause()` in logicutil/logic_utils.go.
- [x] `formula_to_cube(f)` — DONE: `FormulaToClube()` in logicutil/logic_utils.go.
- [x] `de_morgan(f)` — DONE: `DeMorgan()` in logicutil/logic_utils.go.
- [x] `boolean_constant(x)` — DONE: `BooleanConstant()` in logicutil/logic_utils.go.
- [x] `reduce_numerically(ast)` — DONE: `ReduceNumerically()` in logicutil/logic_utils.go.
- [ ] `apply_gen_to_clauses(gen)` — Apply generalization to clauses.
- [ ] `apply_func_to_clauses(func, annot_fun)` — Apply function to clauses.
- [ ] `to_formula(s)`, `to_term(s)`, `to_clause(s)`, `to_clauses(s)`, `to_literal(s)` — Parse string to logic AST (requires ivy_logic_parser).
- [x] `normalize_free_variables_tuple(*asts)` — DONE: `NormalizeFreeVariablesTuple()` in logicutil/logic_utils.go.

### B3. actions/ (vs ivy_actions.py, 1687 lines)

- [x] `match_annotation(action, annot, handler)` — DONE: `MatchAnnotation()` in actions/match.go. Also implemented `UniteAnnot()`, `AnnotationHandler` interface, and `AnnotBranch` type.
- [x] `unite_annot(annot)` — DONE: `UniteAnnot()` in actions/match.go.
- [x] `type_check_action(action, domain, pvars)` — DONE: `TypeCheckAction()` in actions/extra_actions.go.
- [x] `apply_mixin(decl, action1, action2)` — DONE: `ApplyMixin()` in actions/helpers.go.
- [x] `concat_actions(*actions)` — DONE: `ConcatActions()` in actions/helpers.go.
- [x] `append_to_action(action1, action2)` — DONE: `AppendToAction()` in actions/helpers.go.
- [x] `prefix_action(self, stmts)` — DONE: `PrefixAction()` in actions/helpers.go.
- [x] `postfix_action(self, stmts)` — DONE: `PostfixAction()` in actions/helpers.go.
- [x] `env_action(actname, label)` — DONE: `BuildEnvAction()` in actions/extra_actions.go.
- [x] `call_set(action_name, env)` — DONE: `CallSet()` in actions/helpers.go.
- [x] `call_set_rec(action_name, env, res)` — DONE: `CallSetRec()` in actions/helpers.go.
- [x] `has_code(action)` — DONE: `HasCode()` in actions/helpers.go.
- [x] `SubgoalAction` class — DONE in actions/extra_actions.go.
- [x] `VarAction` class — DONE in actions/extra_actions.go.
- [x] `PatternBasedUpdate` class — DONE in actions/extra_actions.go.
- [x] `DerivedUpdate` class — DONE in actions/extra_actions.go.
- [x] `NamedUpdate` class — DONE in actions/extra_actions.go.
- [x] `AssignFieldAction` class — DONE in actions/extra_actions.go.
- [x] `NullFieldAction` class — DONE in actions/extra_actions.go.
- [x] `CopyFieldAction` class — DONE in actions/extra_actions.go.
- [x] `Ranking` class — DONE in actions/extra_actions.go.
- [x] `SymExContext` class — DONE in actions/extra_actions.go.
- [x] `UpdatePattern`, `UpdatePatternList` classes — DONE in actions/extra_actions.go.
- [ ] `Schema.instances()` and other Schema methods.
- [ ] Action `update()` methods — Each Python action type has an `update()` method that computes the transition relation. These may be partially ported but need comparison.

### B4. cppgen/ (vs ivy_to_cpp.py, 6715 lines — LARGEST file)

The Python file has ~215 functions. The Go cppgen/ package has ~136 functions (including tests). Key potentially missing areas:

- [ ] `emit_sig(impl)` — Emit type signature (noted as TODO in Go)
- [ ] `emit_randomize()` — Emit randomization code (noted as TODO in Go)
- [ ] `emit_eval_sig()` — Emit evaluation signature (noted as TODO in Go)
- [ ] Template parameter handling
- [ ] Native code interop (`emit_native`)
- [ ] Serialization/deserialization generation
- [ ] Test harness generation
- [ ] Network/UDP code generation
- [ ] Timer/callback generation
- [ ] Complete class/struct generation
- [ ] **Needs detailed function-by-function audit** — too large for first pass

### B5. isolate/ (vs ivy_isolate.py, 2022 lines)

Python has 74 functions; Go has ~43 non-test functions. Potentially missing:

- [ ] Detailed isolate extraction logic — many helper functions for computing what to include/exclude
- [ ] Mixin before/after merging
- [ ] Visibility/privacy computation
- [ ] Export/import linking
- [ ] Parameter instantiation during isolation
- [ ] **Needs detailed function-by-function audit**

### B6. module/ (vs ivy_module.py, 412 lines)

- [x] `background_theory(symbols)` — Already in module/theory.go: `BackgroundTheory()`.
- [x] `logics()` — DONE: `GetLogics()` in module/module.go.
- [x] `ModuleTheoryContext` class — Already in module/theory.go: `TheoryContext()`.
- [x] `instantiate_non_epr(non_epr, ground_terms)` — DONE: `InstantiateNonEPR()` in module/resort.go.
- [x] `resort_ast(ast)`, `resort_clauses(clauses)`, `resort_asts(asts)` — DONE: `ResortModule()` in module/resort.go, `ResortAst()` in logicutil/logic_utils.go.
- [x] `resort_labeled_asts(asts)` — DONE: `ResortLabeledAsts()` in module/resort.go.
- [x] `resort_map_symbol_sort(m)` — DONE: `ResortMapSymbolSort()` in module/resort.go.
- [ ] `resort_name_ast_pairs(pairs)` — Re-sort name/AST pairs.
- [x] `resort_symbols(symbols)` — DONE: `ResortSymbols()` in module/resort.go.
- [ ] `resort_aliases_map(amap)` — Re-sort aliases.
- [x] `relevant_definitions(symbols)` — Already in module/context.go: `RelevantDefinitions()`.

### B7. check/ (vs ivy_check.py, 1040 lines)

- [ ] `gui_art(other_art)` — GUI-aware analysis graph creation.
- [ ] `show_counterexample(ag, state, bmc_res)` — Display counterexample.
- [ ] `display_cex(msg, ag)` — Display counterexample with message.
- [ ] `preprocess_assumed_ignored_properties()` — Apply ACL filtering to properties.
- [ ] `mc_tactic(prover, goals, proof)` — Model checking tactic.
- [ ] `vmt_tactic(prover, goals, proof)` — VMT export tactic.
- [ ] `start()` — Entry point with argument parsing.
- [ ] `main()` — Main function.
- [ ] `is_unprovable_assert(asrt)` — Check if assertion is marked unprovable.
- [ ] `is_guarantee_mod_unprovable(asrt)` — Check guarantee modulo unprovable.
- [ ] `is_check_mod_unprovable(lf)` — Check modulo unprovable.

---

## C. Potential Logic Variances (Go exists but may differ from Python)

### C1. Action update semantics

- [ ] `AssignAction.update()` in Python (lines 469-574) has complex handling for destructors, variant sorts, and field assignments. Compare with Go `actions/action.go` assignment handling.

- [ ] `WhileAction.update()` in Python (lines 957-1051) has loop unrolling with configurable bound, ranking function checks, and progress property handling. ~95 lines of complex logic.

- [ ] `ChoiceAction.update()` and `EnvAction.update()` — Python has `set_determinize` flag that changes whether choices are deterministic. Verify Go handles this.

- [ ] `CallAction.update()` in Python (lines 1182-1302) — complex call resolution with mixin application, formal/actual parameter binding. ~120 lines.

### C2. Solver / Z3 integration

- [ ] Python `ivy_solver.py` has extensive native Z3 type mapping: `sort_name_to_z3()`, `bfe_to_z3()`, `native_symbol()`, `symbol_to_z3()`, `lookup_native()`. The Go z3bridge/translate.go covers basic cases but may miss native type support for arrays, bit-vectors, etc.

- [ ] Python solver maintains global state (`clear()`, module-level dicts for sorts/relations/functions). Go solver is instance-based. Verify semantics match.

### C3. Parser

- [ ] Python uses PLY (LALR parser generator) while Go uses hand-written recursive descent. Operator precedence and associativity edge cases may differ.

- [ ] Python `ivy_logic_parser.py` provides `to_formula(s)`, `to_clause(s)` etc. for parsing formulas from strings. Go has no equivalent — this is needed by many subsystems.

### C4. Module context management

- [ ] Python uses `__enter__`/`__exit__` with global `module` and `il.sig`. Go Module has `Enter()`/`Exit()` but verify all callers use it correctly.

### C5. Compiler

- [ ] Python `ivy_compiler.py` (2320 lines, 112 functions) vs Go `compiler/` (4150 lines, ~120 functions including tests). Line counts suggest comparable coverage but logic deviations possible. Needs detailed trace.

### C6. Transrel (transition relation)

- [ ] Python `ivy_transrel.py` (669 lines, 68 functions) vs Go `transrel/` (3039 lines, 70+59+55 functions including tests). Go is significantly larger. May have extra functionality or may include ported logic that's more verbose in Go.

---

## D. Intentionally Different / UI Replacement

These Python modules are Tk/Cytoscape UI-specific and are intentionally replaced by Go's `webui/` package:

- tk_ui.py, tk_graph_ui.py, tk_cy.py
- cy_elements.py, cy_render.py, cy_styles.py
- widget_analysis_session.py, widget_cy_graph.py, widget_dialog.py, widget_modal.py, widget_modal_messages.py
- concept_interactive_session.py
- ui_extensions_api.py
- ivy_ui_none.py
- token_counter.py
- client_server_example.py

---

## E. Go-Only Code (no Python equivalent)

- `gogen/` — Go code generation backend (new feature)
- `codegen/` — Generic code generation framework
- `z3bridge/` — Standalone Z3 CGo bridge (replaces Python z3 bindings)
- `cmd/ivyweb/` — Web server entry point
- `pytesthelper/` — Test helper utilities

---

## F. Priority Order for Fixing

1. **ivy_fragment.py** — Without fragment checking, Z3 may diverge on non-decidable formulas
2. **solver/ missing functions** (HerbrandModel, binary_interpolant, model extraction) — Core verification depends on these
3. **actions/ match_annotation** — Required for proof/counterexample extraction
4. **ivy_logic_utils missing functions** — Many subsystems depend on these utilities
5. **ivy_l2s.py** — Required for temporal/liveness property verification
6. **ivy_auto_inst.py** — Required for automatic proof construction
7. **isolate/ missing functions** — Isolate extraction completeness
8. **module/ missing functions** — background_theory, resort functions
9. **cppgen/ completeness** — Detailed audit needed
10. **ivy_alpha.py** — Predicate abstraction for CEGAR
11. **Everything else**
