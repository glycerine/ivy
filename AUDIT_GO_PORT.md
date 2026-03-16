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

- [x] **ivy_logic_parser.py** (655 lines) — DONE: Three LALR grammars (v1.2, v1.6, v1.7) in `lalr_logicparser/` mechanically translated from the Python PLY grammar, plus a hand-written Pratt parser in `parser/expr.go` for v1.7+. The `logicparser/` package dispatches to the LALR parser for ≤v1.6 (correctly implementing the fmla/term split, right-associative ARROW, and chained-comparison rejection) and to the Pratt parser for v1.7+. Public API: `ParseFormula()`, `ParseTerm()`, `ToFormula()`, `ToTerm()`, `ToFormulaV()`, `ToTermV()`. Cross-validated with 25,000+ grammar-guided random formulas and ~7M fuzz executions with 0 mismatches between Pratt and LALR for v1.7.

- [x] **ivy_ev_parser.py** (443 lines) — DONE: Ported to `evparser/`. LALR(1) parser via goyacc (`ev_grammar.y`) with hand-written lexer. Full AST: `Event`, `Symbol`, `App`, `ListValue`, `DictValue`, `Events`. Pattern matching with `Match()`, wildcard `*`, binding variables `$x`. Event generators: `EventGen()`, `EventRevGen()`, `EventFwdGen()`. Traversal utilities: `Filter()`, `Find()`, `Bind()`, `Anchor()`. 16 tests passing.

- [x] **ivy_concept_space.py** (210 lines) — DONE: Ported to `conceptspace/`. LALR(1) parser via goyacc (`cs_grammar.y`) with hand-written lexer. AST: `NamedSpace`, `SumSpace`, `ProductSpace` with `Enumerate()` methods. `ToConceptSpace()` entry point. 9 tests passing.

- [x] **ivy_union_find.py** (50 lines) — DONE: Ported to `unionfind/unionfind.go`. Combined v1 and v2 into single rank-based implementation.

- [x] **ivy_union_find2.py** (64 lines) — DONE: Merged into `unionfind/unionfind.go`.

- [x] **proof.py** (176 lines) — DONE: Ported to `proof/proofstate.go`. All classes implemented: `ProofGoal`, `ProofGoalStack`, `ReachabilityGraph`, `ReachabilityNode`, `ReachabilityEdge`, `AbstractState`, `ConcreteState`, `ProofManager`. The proof package already had `Vocab`, `GoalConc`, `GoalPrems`, `ProofChecker`, `MatchProblem`.

- [~] **tactics.py** (302 lines) — PARTIAL: Ported to `tactics/tactics.go`. Tactic interface + concrete implementations: `RemoveIfRefuted`, `RemoveGoalTactic`, `RefineOrReverseTactic`, `UPDR`, `CheckCover`. UPDR body and full forward/backward image computation deferred pending transrel.Update integration.

- [~] **tactics_api.py** (473 lines) — PARTIAL: `TacticsContext` in `tactics/tactics.go` with `TopGoal()`, `PushGoal()`, `RemoveGoal()`, `RefutedGoal()`, `ForwardImage()`, `BackwardImage()`, `ImpliedFacts()`, `GetDiagram()`, `RefineOrReverse()`. Bodies of ForwardImage/BackwardImage/RefineOrReverse are placeholders pending transrel.Update integration.

- [x] **z3_utils.py** (197 lines) — DONE: `to_z3()` covered by z3bridge/translate.go `Translator.Translate()`. `z3_implies()` covered by `Solver.Implies()`. `z3_implies_batch()` now `Solver.ImpliesBatch()` in solver/solver.go.

- [x] **logic_util.py** (335 lines) — DONE: All functions ported to logicutil/. `used_variables`→`UsedVariables`, `free_variables`→`FreeVariables`/`FreeVariablesByName`, `bound_variables`→`BoundVariables`, `used_constants`→`UsedConstants`, `substitute`→`Substitute`, `substitute_apply`→`SubstituteApply`, `normalize_quantifiers`→`NormalizeQuantifiers`, `is_tautology_equality`→`IsTautologyEquality`, `equal_mod_alpha`→`EqualModAlpha`, `CaptureError`→`CaptureError`.

- [x] **ivy_acl.py** (111 lines) — DONE: Ported to `acl/acl.go`. Full implementation including `RegisterIgnores()`, `RegisterAssumes()`, `IsIgnored()`, `IsAssumed()`, regex support.

### A3. MINOR — Infrastructure / entry points

- [x] **ivy_libs.py** (70 lines) — DONE: Ported to `ivylibs/ivylibs.go`. `LoadSpecs()`, `SaveSpecs()`, `AddLib()`, `RemoveLib()`, `DefaultPrefix()`.
- [x] **ivy_init.py** (113 lines) — DONE: Ported to `ivyinit/ivyinit.go`. `ReadParams()`, `SourceFile()`, `IvyInit()`.
- [ ] **ivy_launch.py** (221 lines) — Process launching for test/build. Python-specific (subprocess, xterm). Will port when integration testing infrastructure is needed.
- [ ] **ivy_dump.py** (60 lines) — Debug dump (a2g→ivy converter). Python-specific (pickle). Low priority: a2g format is Python-specific.
- [ ] **ivy_graphviz.py** (197 lines) — Pydot wrapper. Go equivalent would use a DOT string builder. Low priority for core verification.
- [x] **ivy_smtlib.py** (30 lines) — DONE: `quantifiers_decidable` inlined in ivylogic/globals.go. SMT-LIB Theory/Sort types not needed (Z3 bridge handles this directly).
- [ ] **ivy_shell.py** (12 lines) — Shell env setup. Trivial (prints DYLD_LIBRARY_PATH). Low priority.
- [ ] **ivy_lsp.py** (19 lines) — LSP stub using pygls. Go equivalent would use gopls patterns. Low priority.
- [ ] **iupdr.py** (203 lines) — Interactive UPDR (IPython widget-based, may not be needed).
- [x] **general.py** (12 lines) — DONE: `IvyError` already exists in logic/error.go.
- [ ] **dot_layout.py** (286 lines) — DOT graph layout algorithms. Low priority for core verification.
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
- [x] `apply_gen_to_clauses(gen)` — DONE: Go equivalents in clauseops/ops.go: `VariablesClauses()`, `ConstantsClauses()`, `SymbolsClauses()`, `RelationsClauses()`, `FunctionsClauses()`, `AppsClauses()`, `GroundAppsClauses()`, `EqsClauses()`, `SortsClauses()`.
- [x] `apply_func_to_clauses(func, annot_fun)` — DONE: Go equivalents in clauseops/ops.go: `SubstituteClauses()`, `ResortClauses()`. Plus existing `RenameClauses()`, `SubstituteConstantsClauses()`.
- [x] `to_formula(s)`, `to_term(s)` — DONE: `ToFormula()`, `ToTerm()` in `logicparser/logicparser.go`. Version-aware via `ToFormulaV()`, `ToTermV()`. Uses LALR for ≤v1.6, Pratt for v1.7+.
- [x] `to_clause(s)`, `to_clauses(s)`, `to_literal(s)` — DONE: `Compiler.ToClause()`, `Compiler.ToClauses()`, `Compiler.ToLiteral()` in compiler/parse_helpers.go. Also `Compiler.ToFormula()`, `Compiler.ToFormulaV()`, `Compiler.ToTerm()`, `Compiler.ToTermV()` combining parse+compile+sort-inference.
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
- [x] `Schema.instances()` and other Schema methods — DONE: `Schema.Defines()`, `Schema.Instantiate()` added to actions/action.go. `Schema.Instances` field stores accumulated instantiations.
- [x] Action `update()` methods — DONE: Ported to `actions/update.go`. Implements `ActionUpdate()` for atomic actions (AssumeAction, AssertAction, AssignAction, HavocAction, SetAction, NativeAction) and `IntUpdate()` for compound actions (Sequence, ChoiceAction, EnvAction, IfAction, WhileAction, LocalAction, LetAction, CallAction, BindOldsAction, CrashAction). Top-level `GetUpdate()` hides formals and binds olds. Helper functions: `mkAssignClauses` (assignment TR with ITE for indexed updates), `dualFormula` (negate + skolemize), `skolemizeFormula`, `equivAST`. `WhileAction.Expand()` implements Floyd-Hoare loop encoding (assert invariants, havoc modset, assume invariants, conditional body). `CallAction.applyActuals()` inlines callee with formal/actual parameter binding. 24 tests passing.

### B4. cppgen/ (vs ivy_to_cpp.py, 6715 lines — LARGEST file)

The Python file has ~215 functions. The Go cppgen/ package has ~136 functions (including tests). Key potentially missing areas:

- [x] **Detailed function-by-function audit completed**: Python has 198 functions (69 emit_*); Go has 136 public functions. 67% line coverage (4525 of 6715 lines). Core expression/sort/type emission is complete.
- [ ] `emit_action_gen()` — **CRITICAL**: Generate `<action>_gen` class with preconditions, postconditions, Z3 constraint solving (~150 lines). Foundational for all action execution.
- [ ] `emit_some_action()` — **CRITICAL**: Action execution with parameter binding and constraint solving.
- [ ] `emit_repl_boilerplate3test()` — **CRITICAL**: Test harness with weighted action selection, network I/O multiplexing, reader/timer management (~200 lines).
- [ ] `emit_repl_boilerplate3server()` — Server REPL infrastructure for network agents.
- [ ] `emit_derived()` — Generate derived predicates/functions from definitions.
- [ ] `emit_constructor()` — Struct destructuring constructor initialization.
- [ ] `emit_native()` — User-provided native C++ code blocks.
- [ ] `emit_value_parser()` — REPL parameter value parsing.
- [ ] `emit_ctuple_to_solver()` — Z3 solver conversions for compound tuples.
- [ ] `emit_parameter_assignments()` — Parameter value assignments during initialization.
- [ ] Template parameter handling, tick/progress, method declarations — ~15 more functions needed for complete code generation.

### B5. isolate/ (vs ivy_isolate.py, 2022 lines)

Python has 74 functions; Go has ~43 non-test functions. Potentially missing:

- [x] Detailed isolate extraction logic — DONE: `isolate/iter.go` adds `IterIsolate()` (recursive component traversal with sub-isolate detection), `GetIsolateActions()`, `GetIsolateLFs()`, `GetIsolateConjs()`, `GetIsolatePostConjs()`, `GetIsolateExports()`, `GetIsolateMap()`. Version-aware prefix matching: `VStartsWithEqSome()`, `VStartsWithSomeRec()`, `VStartsWithEqSomeRec()`. Completeness checking: `CheckIsolateCompleteness()`. Privacy management: `SetPrivates()`. Assertion helpers: `HasAssertions()`, `HasRequires()`.
- [x] Mixin before/after merging — Already in `isolate.go`: `AddMixins()` applies before/after mixins.
- [x] Visibility/privacy computation — DONE: `SetPrivates()` in `iter.go`, `VPrivates` for version-aware privacy.
- [x] Export/import linking — DONE: `GetIsolateExports()` in `iter.go`.
- [ ] Parameter instantiation during isolation — `create_isolate()` full logic not yet ported (complex ~200-line function with import creation, mixin ordering, external action construction). Core `IsolateComponent()` exists.
- [x] **Detailed function-by-function audit completed**: 17 of 73 Python functions were already ported. Added ~15 more critical functions. Remaining: `create_isolate()` full logic, `follow_definitions()`, `get_cone()`/`get_mod_cone()`, bracket/loop/native stripping functions.

### B6. module/ (vs ivy_module.py, 412 lines)

- [x] `background_theory(symbols)` — Already in module/theory.go: `BackgroundTheory()`.
- [x] `logics()` — DONE: `GetLogics()` in module/module.go.
- [x] `ModuleTheoryContext` class — Already in module/theory.go: `TheoryContext()`.
- [x] `instantiate_non_epr(non_epr, ground_terms)` — DONE: `InstantiateNonEPR()` in module/resort.go.
- [x] `resort_ast(ast)`, `resort_clauses(clauses)`, `resort_asts(asts)` — DONE: `ResortModule()` in module/resort.go, `ResortAst()` in logicutil/logic_utils.go.
- [x] `resort_labeled_asts(asts)` — DONE: `ResortLabeledAsts()` in module/resort.go.
- [x] `resort_map_symbol_sort(m)` — DONE: `ResortMapSymbolSort()` in module/resort.go.
- [x] `resort_name_ast_pairs(pairs)` — DONE: `ResortNameAstPairs()` in module/resort.go.
- [x] `resort_symbols(symbols)` — DONE: `ResortSymbols()` in module/resort.go.
- [x] `resort_aliases_map(amap)` — DONE: `ResortAliasesMap()` in module/resort.go.
- [x] `relevant_definitions(symbols)` — Already in module/context.go: `RelevantDefinitions()`.

### B7. check/ (vs ivy_check.py, 1040 lines)

- [x] `gui_art(other_art)` — N/A (Tk-specific). Go uses web UI for visualization.
- [x] `show_counterexample(ag, state, bmc_res)` — DONE: `ShowCounterexample()` in check/check.go. Placeholder; web UI handles display.
- [x] `display_cex(msg, ag)` — DONE: `DisplayCex()` in check/check.go. Returns error with message.
- [x] `preprocess_assumed_ignored_properties()` — DONE: `PreprocessAssumedIgnoredProperties()` in check/check.go. Uses acl package for filtering.
- [~] `mc_tactic(prover, goals, proof)` — PARTIAL: `MCTactic()` placeholder in check/check.go. Full implementation needs ivy_mc integration.
- [~] `vmt_tactic(prover, goals, proof)` — PARTIAL: `VMTTactic()` placeholder in check/check.go. Full implementation needs vmt package integration.
- [x] `start()` — DONE: `Start()` in check/check.go.
- [x] `main()` — DONE: `Main()` in check/check.go.
- [x] `is_unprovable_assert(asrt)` — DONE: `IsUnprovableAssert()` in check/check.go.
- [x] `is_guarantee_mod_unprovable(asrt)` — DONE: `IsGuaranteeModUnprovable()` in check/check.go.
- [x] `is_check_mod_unprovable(lf)` — DONE: `IsCheckModUnprovable()` in check/check.go.

---

## C. Potential Logic Variances (Go exists but may differ from Python)

### C1. Action update semantics

- [x] `AssignAction.update()` — DONE: Ported in `actions/update.go`. Handles hierarchy decomposition, destructor assignments (nondeterministic + constraint), variant assignments, and standard assignment with ITE for indexed updates. Destructor handling is simplified compared to Python's nested destructor chain.

- [x] `WhileAction.update()` — DONE: `WhileAction.Expand()` + `IntUpdate()` in `actions/update.go`. Implements Floyd-Hoare encoding: assert invariants → havoc modset → assume invariants → if cond then (body + assert + assume false). Supports ranking functions with local auxiliary variable. Loop unrolling via `UnrollContext` not yet implemented (rarely needed).

- [x] `ChoiceAction.update()` and `EnvAction.update()` — DONE: `ChoiceAction.IntUpdate()` and `EnvAction.IntUpdateEnv()` in `actions/update.go`. ChoiceAction uses `JoinAction` for nondeterministic choice. EnvAction calls `GetUpdate` (with formal hiding) instead of `IntUpdate`. `set_determinize` flag not yet ported (converts binary choices to if-then-else).

- [x] `CallAction.update()` in `actions/update.go` — DONE: `CallAction.IntUpdate()` resolves callee, `applyActuals()` builds input assignments → BindOlds(callee) → output assignments, then hides formals. Simplified compared to Python: does not yet do distinct_obj_renaming to prevent capture (uses direct formal names).

### C2. Solver / Z3 integration

- [~] Python `ivy_solver.py` has extensive native Z3 type mapping: `sort_name_to_z3()`, `bfe_to_z3()`, `native_symbol()`, `symbol_to_z3()`, `lookup_native()`. The Go z3bridge/translate.go covers basic cases. Native type support for arrays and bit-vectors needs audit but basic integer, boolean, and uninterpreted sort mapping is complete. `NativeSymbol()` and `ParseArrayTheory()` already ported in solver/encoding.go.

- [~] Python solver maintains global state (`clear()`, module-level dicts for sorts/relations/functions). Go solver is instance-based. The semantic difference is intentional (Go avoids global state), but callers must pass solver instances. Most call sites are adapted.

### C3. Parser

- [x] **Dual-parser architecture:** For ≤v1.6, Go uses LALR grammars (`lalr_logicparser/v12/`, `lalr_logicparser/v16/`) mechanically translated from the Python PLY grammar, which correctly implement the fmla/term split, version-specific precedence (no precedence for ARROW/IFF in v1.6), and chained-comparison rejection. For v1.7+, Go uses a hand-written Pratt parser (`parser/expr.go`) where terms subsume formulas. The `logicparser/` package dispatches by version.

- [x] **Cross-validation:** The v1.7 hand-written parser has been cross-validated against the v1.7 LALR grammar with 25,000+ grammar-guided random formulas (grammar walker generates random valid formulas from productions), ~7M Go fuzz executions, systematic operator-pair and operator-triple exhaustion, and deep nesting tests (depth 50). Zero mismatches found.

- [x] **v1.6 bug fixes:** Four bugs in the hand-written parser's v1.6 mode were identified via LALR cross-validation and resolved by dispatching ≤v1.6 to the LALR parser: (1) ARROW left-assoc instead of right-assoc, (2) `a & b -> c` grouped as `(a&b)->c` instead of `a&(b->c)`, (3) same for IFF, (4) chained comparisons `a = b = c` accepted instead of rejected.

- [x] Python `ivy_logic_parser.py` `to_formula(s)`, `to_term(s)` — DONE: `logicparser.ToFormula()`, `logicparser.ToTerm()` with version-aware dispatch.

- [x] `to_clause(s)`, `to_clauses(s)`, `to_literal(s)` — DONE: `Compiler.ToClause()`, `Compiler.ToClauses()`, `Compiler.ToLiteral()` in compiler/parse_helpers.go.

### C4. Module context management

- [x] Python uses `__enter__`/`__exit__` with global `module` and `il.sig`. Go Module has `Enter()`/`Exit()` in module/context.go with mutex-protected global `currentModule`. `GetCurrentModule()` provides access. Thread-safe design with `moduleMu` lock.

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
- `lalr_logicparser/` — LALR(1) formula grammars for v1.2/v1.6/v1.7, generated by goyacc from Python PLY grammar translation; includes cross-validation test infrastructure with grammar-guided random formula generator and Go fuzz target
- `logicparser/` — Public formula-parsing API with version-aware dispatch (LALR for ≤v1.6, Pratt for v1.7+)
- `cmd/ivyweb/` — Web server entry point
- `pytesthelper/` — Test helper utilities

---

## F. Priority Order for Remaining Work

1. ~~**ivy_fragment.py**~~ — DONE
2. ~~**solver/ missing functions**~~ — DONE
3. ~~**actions/ match_annotation**~~ — DONE
4. ~~**ivy_logic_utils remaining functions**~~ — DONE: `to_clause`, `to_clauses`, `to_literal` in compiler/parse_helpers.go; `apply_gen_to_clauses`, `apply_func_to_clauses` equivalents in clauseops/ops.go
5. **ivy_l2s.py** — PARTIAL: skeleton exists, full monitor construction deferred pending proof infrastructure
6. ~~**ivy_auto_inst.py**~~ — DONE
7. **isolate/ missing functions** — Isolate extraction completeness, detailed audit needed
8. ~~**module/ missing functions**~~ — DONE (background_theory, resort functions)
9. **cppgen/ completeness** — Detailed audit needed (largest Python file, ~215 functions)
10. ~~**ivy_alpha.py**~~ — DONE
11. ~~**ivy_logic_parser.py**~~ — DONE (LALR + Pratt, cross-validated)
12. ~~**ivy_ev_parser.py**~~ — DONE (LALR parser via goyacc, 16 tests passing)
13. ~~**ivy_concept_space.py**~~ — DONE (LALR parser via goyacc, 9 tests passing)
14. **tactics.py / tactics_api.py** — Interactive refinement (if needed)
15. **Everything else**

### Remaining critical work (in priority order):
- ~~**Action update() methods** (B3)~~ — DONE
- **isolate/ completeness** (B5) — detailed function-by-function audit needed
- **cppgen/ completeness** (B4) — detailed audit of ~215 Python functions vs ~136 Go functions
- **C2: Solver native type support** — arrays, bit-vectors in Z3 translation
- **C4-C6: Module/Compiler/Transrel variances** — require detailed trace comparisons

---

## G. Deferred Test Fixes (must re-enable)

- [x] **bmc/bmc.go safety check**: DONE. Implemented `computeFailUpdate()` which uses `actions.GetUpdateForArt()` to compute the action's transition relation, then `transrel.ActionFailure()` to swap TR and Pre (extracting the assertion-violation condition). The BMC loop now checks this failure condition at each step using `trace.CheckFinalCond()`. This matches Python's `fail_expr`/`fail_action`/`action_failure` chain.

- [ ] **transrel/impl_test.go `TestHistorySatisfyReturnsNil`**: Renamed to `TestHistorySatisfySatReturnsModel` + `TestHistorySatisfyNilPostReturnsNil`. The original test assumed no Z3 solver was available; now Z3 is integrated and `Satisfy(True)` correctly returns a model. The new tests verify the correct behavior. No action needed unless Satisfy semantics change.
