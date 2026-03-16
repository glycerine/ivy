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

- [ ] **ivy_alpha.py** (350 lines) — Predicate abstraction. `alpha()`, `predicate_alpha()`, `ProgressiveDomain`, `RelAlg1/2/3` classes for relational algebra on abstract states. Used in CEGAR refinement loop.

- [ ] **ivy_compose.py** (190 lines) — Compose tactic for liveness proofs. `compose_tactic()`, `create_ranking_defn()`. Needed for compositional liveness reasoning.

- [ ] **ivy_vmt.py** (289 lines) — VMT format export for model checking. `check_isolate()` using VMT, array encoding, `uf_to_array_action()`. Alternative verification backend.

### A2. IMPORTANT — Supporting infrastructure

- [ ] **ivy_logic_parser.py** (655 lines) — Standalone formula parser using PLY. Parses formulas from strings (used by `to_formula()`, `to_clause()`, etc. in ivy_logic_utils.py). The Go parser/ package handles file parsing but may not expose a string-to-formula API.

- [ ] **ivy_ev_parser.py** (443 lines) — Event trace parser. Parses event traces for counterexample display. `Events`, `Event`, `EventParser` classes with full recursive descent parser.

- [ ] **ivy_concept_space.py** (210 lines) — Concept space parser and AST. `NamedSpace`, `SumSpace`, `ProductSpace` classes, PLY grammar for concept spaces, `to_concept_space()`, `clauses_to_concept()`.

- [x] **ivy_union_find.py** (50 lines) — DONE: Ported to `unionfind/unionfind.go`. Combined v1 and v2 into single rank-based implementation.

- [x] **ivy_union_find2.py** (64 lines) — DONE: Merged into `unionfind/unionfind.go`.

- [ ] **proof.py** (176 lines) — Proof infrastructure classes. `IvyModel`, `AnalysisState`, `AnalysisSession`, `ProofGoal`, `ProofGoalStack`, `ReachabilityGraph`, `ReachabilityNode`, `ReachabilityEdge`, `AbstractState`, `ConcreteState`, `ProofManager`. These orchestrate interactive proof sessions.

- [ ] **tactics.py** (302 lines) — Interactive refinement tactics. `RemoveIfRefuted`, `RemoveGoal`, `RefineOrReverse`, `CustomRefineOrReverse`, `PathReach`, `PathReach1`, `PushDiagram`, `RecalculateFacts`, `RemoveFacts`, `ExecuteAction`, `PushNewGoal`, `CheckCover`, `Join2`, `UPDR` tactic classes. Used by the interactive UI.

- [ ] **tactics_api.py** (473 lines) — Tactics API used by UI. `forward_image()`, `backward_image()`, `refine_or_reverse()`, `implied_facts()`, `get_diagram()`, `refuted_goal()`, `push_goal()`, `top_goal()`, `remove_goal()`, `Abstractors` class. Core API for interactive proof exploration.

- [ ] **z3_utils.py** (197 lines) — Z3 utility functions for sort/symbol management.

- [ ] **logic_util.py** (335 lines) — Legacy logic utilities (partially ported to logicutil/ but not fully).

- [ ] **ivy_acl.py** (111 lines) — ACL filtering (check/nocheck theorem lists via YAML). `register_ignores()`, `register_assumes()`, `should_skip()`, `should_assume()`.

### A3. MINOR — Infrastructure / entry points

- [ ] **ivy_libs.py** (70 lines) — Library file loading.
- [ ] **ivy_init.py** (113 lines) — Initialization, parameter reading, `ivy_init()`.
- [ ] **ivy_launch.py** (221 lines) — Process launching for test/build.
- [ ] **ivy_dump.py** (60 lines) — Debug dump utilities.
- [ ] **ivy_graphviz.py** (197 lines) — Graphviz DOT output generation.
- [ ] **ivy_smtlib.py** (30 lines) — SMT-LIB format output.
- [ ] **ivy_shell.py** (12 lines) — Shell/REPL entry point.
- [ ] **ivy_lsp.py** (19 lines) — Language Server Protocol stub.
- [ ] **iupdr.py** (203 lines) — Interactive UPDR (IPython widget-based, may not be needed).
- [ ] **general.py** (12 lines) — General utilities.
- [ ] **dot_layout.py** (286 lines) — DOT graph layout algorithms.
- [ ] **concept.py** (865 lines) — Concept graph model. Partially ported to webui/concept.go but needs audit for completeness.

---

## B. Missing Functions Within Ported Modules

### B1. solver/ (vs ivy_solver.py, 1716 lines)

- [ ] `binary_interpolant(clauses2, clauses1)` — Craig interpolation. **CRITICAL** for CEGAR-based verification. Uses Z3 interpolation API.
- [x] `HerbrandModel` class (~80 lines) — DONE: Ported to `solver/herbrand.go`. Full implementation including `Sorts()`, `SortUniverse()`, `Eval()`, `EvalConstant()`, `EvalToConstant()`, `Check()`. Also added `Model.Sorts()`, `Model.SortUniverse()`, `Sort.String()` to z3bridge.
- [x] `sort_from_z3(s)` — DONE: Handled via z3bridge Sort.String() and sig lookup in HerbrandModel.
- [x] `constant_from_z3(sort, c)` — DONE: `constantFromZ3()` in solver/herbrand.go.
- [x] `get_model_constant(m, t)` — DONE: `HerbrandModel.getModelConstant()` in solver/herbrand.go.
- [x] `model_universe_facts(h, sort, upclose)` — DONE: `ModelUniverseFacts()` in solver/herbrand.go.
- [x] `model_facts(h, ignore, clauses1, upclose)` — DONE: `ModelFacts()` in solver/herbrand.go.
- [x] `relation_model_to_clauses(h, r, n)` — DONE: `RelationModelToClauses()` in solver/herbrand.go.
- [x] `function_model_to_clauses(h, f)` — DONE: `FunctionModelToClauses()` in solver/herbrand.go.
- [x] `get_lit_facts(h, lit, res)` — DONE: `getLitFacts()` in solver/herbrand.go.
- [ ] `numeral_assign(clauses, h)` — Assign numerals from model.
- [x] `clauses_case(clauses1)` — DONE: `Solver.ClausesCase()` in solver/herbrand.go.
- [ ] `clause_model_simp(m, c)` — Simplify clause using model.
- [ ] `mine_interpreted_constants(model, vocab)` — Extract interpreted constants from Z3 model.
- [x] `enumerated_range(sort)` — DONE: `Solver.EnumeratedRange()` in solver/herbrand.go.
- [ ] `collect_numerals(z3term)` — Collect numeral subterms.
- [ ] `from_z3_numeral(z3term, sort)` — Convert Z3 numeral to Ivy.
- [ ] `collect_model_values(sort, model, sym)` — Collect all model values for a symbol.
- [ ] `SortOrder` class — Ordering on sorts for model construction.
- [ ] `encode_term(t, n, sort)` — Binary encoding of terms.
- [ ] `encode_equality(*terms)` — Binary encoding of equality.
- [ ] `binenc(m, n)`, `gebin(bits, n)`, `ceillog2(n)` — Binary encoding helpers.
- [ ] `z3_function(name, sig)` — Create Z3 function declaration.
- [ ] `z3_to_formula(z3expr, vars)` — Convert Z3 expression back to Ivy formula.
- [ ] `z3sort_to_sort(z3sort)` — Convert Z3 sort to Ivy sort.
- [ ] `z3decl_to_symbol(z3decl)` — Convert Z3 func_decl to Ivy symbol.
- [ ] `terms_match(tl1, tl2)` — Check if term lists structurally match.
- [ ] `get_arg_range(m, x)` — Get argument range from model.
- [ ] `model_if_none(clauses1, implied, model)` — Get model if none provided.
- [ ] `substitute(t, *m)` — Z3 substitution wrapper.
- [ ] `set_seed(seed)` — Set Z3 random seed.
- [ ] `set_macro_finder(truth)` — Enable/disable Z3 macro finder.
- [ ] `set_use_native_enums(t)` — Enable/disable native enum sorts.
- [ ] `parse_array_theory(name)` — Parse array theory from sort name.
- [ ] `parse_int_params(name)` — Parse integer parameters from sort name.
- [ ] `is_solver_sort(name)` — Check if name is a Z3 built-in sort.
- [ ] `is_solver_op(name)` — Check if name is a Z3 built-in operation.
- [ ] `check_native_compat_sym(sym)` — Check native type compatibility.
- [ ] `check_compat()` — Check all native compatibility.
- [x] `sort_card(sort)` — DONE: `SortCard()` in solver/herbrand.go.
- [ ] `native_symbol(sym)` — Get native Z3 symbol.
- [ ] `lt_pred(sort)` — Get less-than predicate for sort.
- [ ] `get_polymacs(op)` — Get polymorphic macro.
- [ ] `numeral_to_z3(num)` — Convert numeral to Z3.
- [ ] `enumerated_to_numeral(term)` — Convert enumerated constant to numeral.
- [ ] `quant_constraints(vs, z3_vs)` — Sort constraints for quantifier variables.
- [ ] `range_sort_bounds_to_z3(itp)` — Convert range sort bounds to Z3.
- [ ] `type_constraints(syms)` — Type constraints for symbols.

### B2. logicutil/ + clauseops/ (vs ivy_logic_utils.py, 1635 lines)

- [x] `close_epr(fmla)` — DONE: `CloseEPR()` in logicutil/logic_utils.go.
- [ ] `normalize_named_binders(ast, names)` — Normalize named binder structure.
- [ ] `replace_temporals_by_named_binder_g_ast(ast, g, when)` — Replace temporal operators with named binders.
- [ ] `reduce_named_binders(ast, g)` — Reduce named binders to temporal operators.
- [ ] `replace_named_binders_ast(ast, subs)` — Substitute named binders.
- [ ] `expand_named_binders_ast(ast, fun)` — Expand named binders using function.
- [ ] `denormalize_temporal(ast)` — Reverse temporal normalization.
- [x] `resort_sort(sort, subs)` — DONE: `ResortSort()` in logicutil/logic_utils.go.
- [x] `resort_symbol(sym, subs)` — DONE: Handled via ResortAst for Const nodes.
- [x] `resort_var(sym, subs)` — DONE: Handled via ResortAst for Var nodes.
- [x] `resort_ast(ast, subs)` — DONE: `ResortAst()` in logicutil/logic_utils.go.
- [ ] `resort_sig(subs)` — Re-map all sorts in signature.
- [ ] `rename_clauses_annot_fun(annot, map)` — Rename within clause annotations.
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
- [ ] `is_vac_equality_lit(lit)` — Check if literal is vacuously true equality.
- [ ] `is_true_lit(lit)`, `is_false_lit(lit)` — Check literal truth value.
- [ ] `is_taut_lit(lit)`, `is_vac_lit(lit)` — Tautology/vacuity check.
- [x] `is_disequality_lit(lit)` — DONE: `IsDisequalityLit()` in logicutil/logic_utils.go.
- [x] `is_ground_clause(clause)` — DONE: `IsGroundLit()` in logicutil/logic_utils.go.
- [x] `is_ground_equality_lit(lit)` — DONE: `IsGroundEqualityLit()` in logicutil/logic_utils.go.
- [ ] `term_eq(t1, t2)` — Already provided by lg.Node.Equal().
- [ ] `term_lists_eq(l1, l2)` — Structural term list equality.
- [ ] `atom_eq(at1, at2)` — Already provided by lg.Node.Equal().
- [ ] `lit_eq(lit1, lit2)` — Already provided by Literal.Equal().
- [x] `swap_args_lit(lit)` — DONE: `SwapArgsLit()` in logicutil/logic_utils.go.
- [x] `eq_lit(x, y)` — DONE: `EqLit()` in logicutil/logic_utils.go.
- [x] `eq_atom(x, y)` — DONE: `EqAtom()` in logicutil/logic_utils.go.
- [x] `rel_inst(relname)` — DONE: `RelInst()` in logicutil/logic_utils.go.
- [x] `fun_inst(funname)` — DONE: `FunInst()` in logicutil/logic_utils.go.
- [x] `fun_eq_inst(funname)` — DONE: `FunEqInst()` in logicutil/logic_utils.go.
- [x] `is_relational(sym)` — DONE: `IsRelational()` in logicutil/logic_utils.go.
- [ ] `TseitinContext` class + `tseitin_encoding(f)` — Tseitin transformation for CNF.
- [x] `expand_abbrevs(f)` — DONE: `ExpandAbbrevs()` in logicutil/logic_utils.go.
- [ ] `formula_to_lit(f)` — Convert formula to literal.
- [ ] `formula_to_clause(f)` — Convert formula to clause (disjunction).
- [ ] `formula_to_cube(f)` — Convert formula to cube (conjunction of literals).
- [x] `de_morgan(f)` — DONE: `DeMorgan()` in logicutil/logic_utils.go.
- [x] `boolean_constant(x)` — DONE: `BooleanConstant()` in logicutil/logic_utils.go.
- [ ] `reduce_numerically(ast)` — Evaluate numeric expressions.
- [ ] `apply_gen_to_clauses(gen)` — Apply generalization to clauses.
- [ ] `apply_func_to_clauses(func, annot_fun)` — Apply function to clauses.
- [ ] `to_formula(s)`, `to_term(s)`, `to_clause(s)`, `to_clauses(s)`, `to_literal(s)` — Parse string to logic AST (requires ivy_logic_parser).
- [ ] `normalize_free_variables_tuple(*asts)` — Normalize free variables across tuple.

### B3. actions/ (vs ivy_actions.py, 1687 lines)

- [x] `match_annotation(action, annot, handler)` — DONE: `MatchAnnotation()` in actions/match.go. Also implemented `UniteAnnot()`, `AnnotationHandler` interface, and `AnnotBranch` type.
- [ ] `unite_annot(annot)` — Merge annotations.
- [ ] `type_check_action(action, domain, pvars)` — Type check an action.
- [ ] `apply_mixin(decl, action1, action2)` — Apply mixin before/after to action.
- [ ] `concat_actions(*actions)` — Concatenate multiple actions into sequence.
- [ ] `append_to_action(action1, action2)` — Append action2 to action1.
- [ ] `prefix_action(self, stmts)` — Prepend statements to action.
- [ ] `postfix_action(self, stmts)` — Append statements to action.
- [ ] `env_action(actname, label)` — Construct environment (external) action.
- [ ] `call_set(action_name, env)` — Compute transitive call set.
- [ ] `call_set_rec(action_name, env, res)` — Recursive helper for call_set.
- [ ] `has_code(action)` — Check if action contains executable code.
- [ ] `SubgoalAction` class — Action with subgoal annotation.
- [ ] `VarAction` class — Variable declaration action.
- [ ] `PatternBasedUpdate` class — Pattern-based state update.
- [ ] `DerivedUpdate` class — Derived relation update.
- [ ] `NamedUpdate` class — Named update.
- [ ] `AssignFieldAction` class — Field assignment.
- [ ] `NullFieldAction` class — Null field action.
- [ ] `CopyFieldAction` class — Copy field action.
- [ ] `Ranking` class — Ranking function for liveness.
- [ ] `SymExContext` class — Symbolic execution context.
- [ ] `UpdatePattern`, `UpdatePatternList` classes — Update patterns.
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
