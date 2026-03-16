# MISSING.md — goivy Port Gap Analysis

Generated: 2026-03-16

Python Ivy codebase: `~/pyivy/ivy/ivy/` (~50,600 lines across 107 .py files)
Go goivy codebase: `~/go/src/github.com/glycerine/goivy/` (~25,000 lines across 153 .go files, plus ~4,600 lines of tests)

This document catalogs every functional difference between Python Ivy and Go goivy,
organized by severity and grouped by subsystem. Each item includes a description
paragraph to guide mechanical porting.

---
INSTRUCTIONS TO CLAUDE:

This is a mechanical port from python to Go of the ivy project, /Users/jaten/pyivy/ivy. To
fill in the many critical missing pieces of this port, please follow the plan in
MISSING.md and proceed systematically and sequentially to implement the missing
functionality in the Go port (goivy; the mechanical port to go, here in
/Users/jaten/go/src/github.com/glycerine/goivy ). Do not delegate to parallel agents, as
this is alot of integration work and they will have insufficient context. Parallel agents
created the current half-finished mess, and we will not tolerate their losey-goosey slop.

Work systematically through the list in the MISSING.md and when you finish an item,
check it off on the list in the MISSING.md file, and then immediately proceed to the
next item. Do not pause and ask for guidance. Do not simplify. Do not stub out. Do the
full, deep, complex work of the port now, in a depth-first fashion. Then proceed to the
next item on the list. I will be asleep and not available, so if you cannot figure
something out, make a clear note by it in the MISSING.md, and proceed to the next
item. Do not stop until all items on the list have been either completed or marked as could
not figure it out. Start now.

---

## Table of Contents

1. [CRITICAL — Blocking Core Verification](#1-critical--blocking-core-verification)
2. [OMIT/WE WILL NOT PORT — C++ Code Generation](#2-omit--c-code-generation)
3. [HIGH — Compiler Infrastructure](#3-high--compiler-infrastructure)
4. [HIGH — Action Semantics](#4-high--action-semantics)
5. [HIGH — Isolate Extraction](#5-high--isolate-extraction)
6. [HIGH — Liveness / Temporal](#6-high--liveness--temporal)
7. [MEDIUM — Model Checking Pipeline](#7-medium--model-checking-pipeline)
8. [MEDIUM — Solver Infrastructure](#8-medium--solver-infrastructure)
9. [MEDIUM — Proof System](#9-medium--proof-system)
10. [MEDIUM — Analysis Graph & Interpretation](#10-medium--analysis-graph--interpretation)
11. [MEDIUM — Supporting Subsystems](#11-medium--supporting-subsystems)
12. [LOW — UI/Web, Secondary Backends, Tests](#12-low--uiweb-secondary-backends-tests)

---

## 1. CRITICAL — Blocking Core Verification

### [x] 1.1 check/check.go: `CheckFcsInState` is a stub (no solver call)
**Python**: `ivy_check.py:373-416`. Calls `ag.get_history(post)`, computes `itr.small_model_clauses`, performs actual Z3 satisfiability check, and reconstructs trace via `ivy_trace.Trace` + `act.match_annotation` on counterexample. **Go**: Just calls `Pass()` on every checker — no solver, no trace. This is the inner loop of all verification; without it, every check auto-passes. Port requires: wiring `solver.GetSmallModel` into the history's transition relation, building a `Trace` from the model, and calling `MatchAnnotation` for diagnosis.

### [x] 1.2 check/isolate_check.go: `CheckIsolate` body is hollow
**Python**: `ivy_check.py:560-714`. Creates an `AnalysisGraph`, adds initial state, checks `init_cond` establishes invariant, then for each external action: executes action against pre-state (conjectures), checks post-state conjectures, iterates all guarantees calling `check_safety_in_state` with `checked_assert` scoping. **Go**: Prints action names but the initialization check, action execution loop, and guarantee-checking loop are all stub comments. Port the three phases: (a) init invariant check, (b) action preservation loop, (c) guarantee assertion loop, each calling `CheckFcsInState`.

### [x] 1.3 check/isolate_check.go: `CheckModule` never calls `create_isolate`
**Python**: `ivy_check.py:940` calls `ivy_isolate.create_isolate(isolate)` which flattens module hierarchy, resolves mixins, creates the checked isolate with specification/implementation separation. **Go**: `CheckModule` proceeds directly to `CheckIsolate` without isolate preparation. Without this call, the module passed to checking has unresolved mixins, unseparated spec/impl, and unfiltered cone of influence. Port: call `isolate.CreateIsolate(isoName, mod)` early in `CheckModule`, after the current `CreateIsolate` is itself completed (see §5).

### [x] 1.4 solver/solver.go: `ClausesToZ3` missing `type_constraints`
**Python**: `ivy_solver.py:583`. After translating clauses, appends `type_constraints(used_symbols_clauses(clauses))` which generates non-negativity for `nat` sorts and bound constraints for range sorts. **Go**: Translates formulas and definitions but skips type constraints entirely. Any program using `nat` or range sorts (e.g., `type port = {0..65535}`) will produce unsound results. Port: after translating clauses to Z3, collect all symbols via `UsedSymbolsClauses`, compute per-symbol type constraints (nat: `x >= 0`, range: `lo <= x && x <= hi`), and add them to the Z3 context.

### [x] 1.5 solver/solver.go: `GetSmallModel` missing `final_cond` callback
**Python**: `ivy_solver.py:1143-1302`. The `final_cond` parameter is a callable with `start()`, `sat()`, `unsat()`, `assume()` methods that implements incremental sort-size search — the solver tries increasing universe sizes until it finds a model or proves unsat. Also handles `opt_incremental` and `opt_show_vcs`. **Go**: `GetSmallModel` calls Z3 once without sort-size iteration. Port requires: implementing the `FinalCond` interface (start/sat/unsat/assume callbacks), the incremental universe-size search loop, and wiring it into the `Check` call with proper push/pop.

### [x] 1.6 solver/model.go: `ClausesModelToClauses` and `ClausesModelToDiagram` are placeholders
**Python**: `ivy_solver.py:1340-1398`. `clauses_model_to_clauses` calls `numeral_assign` to name universe elements, then `substitute_constants_clauses` to rewrite model values into Ivy terms. `clauses_model_to_diagram` additionally calls `bound_quantifiers_clauses` for weakening, `substitute_constants_clauses`, Skolem filtering, and `upward_close`. **Go**: Both use placeholder `&lg.Eq{T1: sym, T2: sym}` instead of actual model conversion. These are needed by the CEGAR refinement loop, BMC, and trace construction. Port requires: implementing `numeral_assign`, `substitute_constants_clauses`, and the Skolem-filtering/weakening pipeline.

### [x] 1.7 solver/compat.go: `ModelIfNone` returns nil
**Python**: `ivy_solver.py:1135-1145`. Creates model with incremental sort-size search if no model given. **Go**: Returns `nil` with TODO comment. Required by `ClausesModelToClauses` and `ClausesModelToDiagram`. Port: implement the sort-size iteration loop that tries `CheckContext` with increasing cardinality constraints.

---

## 2. OMIT — C++ Code Generation. We will not port this to Go.

---

## 3. HIGH — Compiler Infrastructure

### [x] 3.1 compiler: No `ivy_compile` main entry point
**Python**: `ivy_compiler.py:2190-2254`. Runs three separate declaration interpreter passes (`IvyDomainSetup`, `IvyConjectureSetup`, `IvyARGSetup`) inside a `TopContext(collect_actions(decls))` wrapper, then runs post-processing: `create_sort_order`, `create_constructor_schemata`, `fix_constructors`, `check_definitions`, `attach_proofs`, `check_properties`, `apply_assert_proofs`, `create_conj_actions`, `handle_temporals`, `ivy_isolate.create_isolate`. **Go**: Has `ProcessDecls` (single pass) but no multi-pass design, no `TopContext`/`collect_actions` forward reference resolution, and no post-processing passes. Port requires: restructuring to three passes plus all post-processing.

### [x] 3.2 compiler: Missing action compilation for `local`, `while`, `native`, `crash`, `thunk`, `debug`, `choice`
**Python**: `ivy_compiler.py:471-782`. Each action type has a `cmpl()` method. `compile_local` handles local variable declaration with assignment inference and sort inference. `compile_while_action` handles `Some` in condition. `compile_native_action` handles backtick fields. `compile_thunk_action` creates destructor sorts. **Go**: `CompileActionBody` in action.go handles `And`, `:=`, `require`, `ensure`, `assert`, `assume`, `call`, `Ite` but is missing: `while`, `local`, `native`, `crash`, `thunk`, `debug`, `choice`. Port each missing case as a new branch in `CompileActionBody`.

### [x] 3.3 compiler: `check_definitions` validation pass
**Python**: `ivy_compiler.py:1696-1775`. Validates definitions have no cycles (via DFS), no redefinition, no interference with axioms, and uses ProofChecker for recursive definitions. **Go**: Absent. Port: implement DFS cycle detection on definition dependency graph, check for definition conflicts, and wire in proof checker.

### [x] 3.4 compiler: `check_properties` proof checking pass
**Python**: `ivy_compiler.py:1972-2053`. Uses `ivy_proof.ProofChecker` to verify each property's proof, converts theorems to properties via Skolemization, generates subgoals. **Go**: Absent. Port: iterate properties, instantiate ProofChecker, verify proofs, generate subgoals, convert proved theorems.

### [x] 3.5 compiler: `create_sort_order` topological sort
**Python**: `ivy_compiler.py:1632-1649`. Uses Tarjan SCC to order type declarations. **Go**: Absent. Port: implement Tarjan SCC on type dependency graph, reorder module's sort declarations.

### [x] 3.6 compiler: `collect_actions` / `TopContext` forward references
**Python**: `ivy_compiler.py:1565-1576`. Pre-collects all action signatures (including KeyArg handling) so that forward references resolve during compilation. **Go**: `TopCtx` field exists on `Compiler` but is never populated. Port: scan all declarations for action signatures before processing.

### [ ] 3.7 compiler: Schema/tactic compilation (10+ tactics)
**Python**: `ivy_compiler.py:912-1003`. `compile_schema_instantiation`, `compile_let_tactic`, `compile_witness_tactic`, `compile_unfold_tactic`, `compile_forget_tactic`, `compile_if_tactic`, `compile_property_tactic`, `compile_function_tactic`, `compile_tactic_tactic`, `compile_proof_tactic`. **Go**: None ported. Port each as a case in proof/tactic compilation.

### [ ] 3.8 compiler: `CompileAssign` missing tuple and variant support
**Python**: `ivy_compiler.py:528-570`. Handles tuple assignment (splitting LHS into components), variant sort inference (`is_variant` / `pto`), and full sort-unification. **Go**: Handles simple and indexed assignment but no tuple or variant. Port: add tuple destructuring and variant sort inference branches.

### [ ] 3.9 compiler: `CompileDefn` missing `SomeExpr` and `DefinitionSchema`
**Python**: `ivy_compiler.py:830-867`. Handles conditional definitions (SomeExpr), DefinitionSchema support, and variable sort substitution. **Go**: compiler.go:901-938 handles basic definitions. Port: add SomeExpr and DefinitionSchema branches.

### [ ] 3.10 compiler: Module instantiation (`inst_mod`)
**Python**: `ivy_compiler.py:1090-1210`. Full module instantiation with parameter binding, sort mapping, and recursive expansion. **Go**: `DeclInterp.ModuleD` just stores the raw node. Port: implement parameter binding and AST substitution.

---

## 4. HIGH — Action Semantics

### [x] 4.1 actions: `assert_to_assume(kinds)` method entirely missing
**Python**: `ivy_actions.py:243, 363, 374, 1014`. Base `Action.assert_to_assume` recursively clones with converted children. `AssertAction` version checks `self.kind` and converts to `AssumeAction`. `EnsuresAction` has version-dependent logic. `WhileAction` handles invariant conversion. **Go**: No method on any action type. Port: add `AssertToAssume(kinds []string) Action` interface method, implement on each action type (recursive clone with assertion→assumption conversion).

### [x] 4.2 actions: `modifies()` method missing
**Python**: `ivy_actions.py:282, 479, 648, 1099`. Returns the set of symbols modified by the action. `AssignAction.modifies()` walks destructor chains to find root symbol. `HavocAction.modifies()` similar. `CrashAction.modifies()` recursive. **Go**: Absent. Port: add `Modifies() []string` interface method, implement for each action type.

### [ ] 4.3 actions: `decompose(pre, post, fail)` signature mismatch
**Python**: `ivy_actions.py:279`. Takes `pre`, `post` state tuples for threading pre/post through decomposition. **Go**: `Decompose()` takes no arguments, returns `[][]Action`. Port: change signature to accept pre/post state parameters, implement state threading for `LocalAction` (hide_state), `CallAction` (formal/actual mapping), `WhileAction` (expand then decompose).

### [ ] 4.4 actions: `CallAction.apply_actuals` missing capture avoidance
**Python**: `ivy_actions.py:1214`. Uses `distinct_obj_renaming` to avoid variable capture, maps `old(s)` to `old(t)`, substitutes callee AST, checks sort compatibility including variant sorts. **Go**: `update.go:1218` skips capture avoidance, old-symbol handling, and callee substitution. Port: add `distinct_obj_renaming`, old-symbol mapping, and sort validation.

### [x] 4.5 actions: `SetAction.ActionUpdate` is a stub
**Python**: `ivy_actions.py:624`. Computes transition relation with new_n, sign-based polarity formulas, and equality constraints. **Go**: `update.go:727` returns trivial null update. Port: implement sign-based polarity formula construction matching Python's `set_action_update`.

### [ ] 4.6 actions: `InstantiateAction` type missing
**Python**: `ivy_actions.py:742`. Handles macro instantiation and schema resolution with its own `int_update` and `cmpl` methods. **Go**: No struct or implementation. Port: define `InstantiateAction` struct, implement `IntUpdate` and compiler integration.

### [x] 4.7 actions: Missing `references()`, `get_references()`, `erase_unrefed()`
**Python**: `ivy_actions.py:283, 298, 303`. `references` collects non-action symbol references. `get_references` recursive version. `erase_unrefed` replaces unreferenced actions with empty `Sequence()`. Needed for cone-of-influence filtering. **Go**: Absent. Port: add to Action interface.

### [x] 4.8 actions: Missing `prefix_calls()`, `drop_invariants()`, `unroll_loops()`
**Python**: `ivy_actions.py:253, 248, 258`. `prefix_calls` renames call targets (used during isolate composition). `drop_invariants` strips loop invariants. `unroll_loops` converts while loops to bounded if-then-else chains. **Go**: Absent. Port each as a method on the Action interface.

### [x] 4.9 actions: Annotation threading absent from updates
**Python**: Every `action_update` and `int_update` constructs clause sets with `EmptyAnnotation`. Annotations enable trace reconstruction from satisfying assignments. **Go**: Update functions create bare `transrel.Update` structs with no annotation fields. Port: add `Annotation` field to `transrel.Update`, thread annotations through all update construction.

### [ ] 4.10 actions: `match_annotation` CallAction inlining and WhileAction expansion
**Python**: `ivy_actions.py:1633-1651`. CallAction: resolves callee, wraps in `Sequence(IgnoreAction(), callee, ReturnAction())`, recurses. WhileAction: expands loop then recurses. **Go**: `match.go:180` calls `handler.Handle` without inlining. Port: add callee resolution and loop expansion.

### [ ] 4.11 actions: `PatternBasedUpdate.get_update_axioms` / `DerivedUpdate.get_update_axioms`
**Python**: `ivy_actions.py:151, 169`. Pattern matching against action structure, substituted precond/postcond. Derived checks if any dependency is in updated set. **Go**: Data holders with no matching logic. Port: implement pattern matching and dependency checking.

---

## 5. HIGH — Isolate Extraction

### [ ] 5.1 isolate: `isolate_component()` core function (~500 lines) heavily simplified
**Python**: `ivy_isolate.py:887-1389`. The core isolation function handling: implementation_map, impl_mixins, delegate handling, 6 different `assert_to_assume` lambda variants, verification/present action classification with `prefix_calls`, export discovery with side-effect checks, export_preconds, conjecture filtering (version-aware), init filtering, axiom filtering, property-to-axiom conversion, native filtering, initializer filtering, definition filtering, symbol filtering, sort filtering, interference checking, trust checking, `strip_isolate` call, `init_cond` computation. **Go**: `IsolateComponent()` in isolate.go:292-380 is ~90 lines capturing only outermost structure. Port: implement each subsection matching Python's flow.

### [ ] 5.2 isolate/create.go: `CreateIsolate` heavily simplified
**Python**: `ivy_isolate.py:1557-1782`. Full version includes: `check_with_parameters`, `apply_present_conjectures`, `create_imports` (~60 lines), `get_mixin_order` (topological sort), `fix_initializers`, `canonize_types`, `update_conjs`, `slv.check_compat`, `show_compiled`, bracket application, pedantic warnings. **Go**: create.go has basic mixin application and export handling but is missing all the above. Port: add each missing subsection.

### [ ] 5.3 isolate: Missing ~30 helper functions
Functions entirely absent from Go: `get_strip_binding`, `strip_natives`, `has_unsummarized_mixins`, `get_callouts_action`/`get_callouts`, `get_loc_mods`, `find_references`, `spec_ancestors`, `get_prop_dependencies`, `set_privates_prefer`, `get_private_from_attributes`, `get_props_proved_in_isolate`, `check_with_parameters`, `get_isolate_info` (full version), `follow_definitions_rec`/`follow_definitions`, `collect_relevant_destructors`, `add_extern_precond`, `get_mixin_order`/`SortOrder`, `hide_action_params`, `loop_action`, `fix_initializers`, `set_up_implementation_map`, `conj_to_assume`, `bracket_action_int`/`bracket_action`, `apply_present_conjectures`. Port each as needed by §5.1 and §5.2.

### [ ] 5.4 isolate/strip.go: `StripIsolate` missing variable substitution
**Python**: `ivy_isolate.py:341-456`. Full version handles: variable isolate parameter substitution (lines 345-352), `is_init`/`init_params`, `strip_added_symbols`, `param_defaults`, version ≤1.6 param clearing, native quote stripping. **Go**: strip.go:270-347 missing all of these. Port each missing branch.

### [ ] 5.5 isolate: `check_interference()` simplified
**Python**: `ivy_isolate.py` related functions. Tracks `impl_mixins`, `check_term`, `interf_syms`, `after_inits`, `pre_refed` symbols, `callouts` quadruples, export interference with `after_init_refs`. **Go**: deps.go:206-286 only checks basic call→mod interference. Port: add mixin interference, callout tracking, and pre-referenced symbol filtering.

---

## 6. HIGH — Liveness / Temporal

### [ ] 6.1 l2s/l2s.go: Core L2S transformation body (~1200 lines) missing
**Python**: `ivy_l2s.py` core `l2s_tactic_int` function. Extracts temporal formula, builds monitor automaton, creates Skolem constants, adds fairness constraints, instruments all actions with saved-state copies and l2s_waiting/frozen updates, replaces named binders with fresh relations, generates the safety property `~(l2s_saved & l2s_frozen & ~l2s_waiting & ~phi)`. **Go**: `l2sTacticInt` has a skeleton (steps 1-8 superficially) but the monitor construction, action instrumentation, and named binder replacement are missing. Port: implement each step of the L2S construction following Python line by line.

### [ ] 6.2 temporal/temporal.go: `invariance_tactic` missing
**Python**: `ivy_temporal.py` ~130 lines. Collects auxiliary invariants, instruments actions with `instr_stmt`, adds property events, modifies goal. This is the core tactic for proving `globally phi`. **Go**: Absent. Port: implement the invariance rule with action instrumentation.

### [ ] 6.3 ranking/ranking.go: Ranking tactic (~400 lines) is a stub
**Python**: `ivy_ranking.py` core function. Processes tasks/triggers from proof declarations, compiles formulas, builds postconditions, generates L2S ghost state, instruments model actions. **Go**: `L2STactic` returns goals unchanged. Port: implement task/trigger parsing, postcondition generation, and ghost state instrumentation.

### [ ] 6.4 l2s: Missing helper functions
`trace_hook`, `is_l2s_goal`, `get_l2s_props`, `prop_event`, `wait_event`, `desugar` ($was/$happened), `convert_to_init`, `trig_glob`, `fresh_name_binder`, `fresh_rel`, `desugar_was_happened`, `add_l2s_postconds`. Port each as needed by §6.1.

---

## 7. MEDIUM — Model Checking Pipeline

### [ ] 7.1 mc/toaiger.go: `to_aiger()` pipeline entirely missing (~310 lines)
**Python**: `ivy_mc.py:1117-1427`. The complete MC pipeline: error flag instrumentation, initialization/external action composition, invariant Skolemization, transition relation computation, 5 transformation steps (non-finite definition elimination, ITE elimination, quantifier elimination, axiom instantiation via pattern matching, propositional abstraction), state variable management, AIGER encoding. **Go**: `ToAigerStub` is 7 lines returning empty Encoder. Port requires: implementing each of the 5 transformation steps and the AIGER encoding.

### [ ] 7.2 mc/checker.go: `check_isolate` stub
**Python**: `ivy_mc.py:1684-1767`. Calls `to_aiger`, writes AAG, converts via `aigtoaig`, runs ABC, scrapes output, calls `aiger_witness_to_ivy_trace2`. **Go**: Returns error. Port: implement after §7.1.

### [ ] 7.3 mc: Missing `Encoder.eval()` — expression-to-AIGER-literal evaluation
**Python**: `ivy_mc.py:313-359`. Evaluates `il.And`, `il.Or`, `il.Not`, `il.Ite`, `il.Equals`, applications into AIGER literals. **Go**: No `eval` on Encoder. Port: implement recursive evaluation matching Python.

### [ ] 7.4 mc: Missing quantifier elimination (`Qelim.qe()`)
**Python**: `ivy_mc.py:881-914`. Recursive QE: handles forall/exists by finite instantiation, macro expansion, normalization. **Go**: `Qelim` has `Fresh`/`GetConsts` only. Port: implement `qe()` recursive elimination.

### [ ] 7.5 mc: Missing propositional abstraction (`mk_prop_abs`)
**Python**: `ivy_mc.py:1287-1318`. Replaces non-propositional atoms with fresh boolean variables, with `prev_expr` detection for state variable linkage. **Go**: `PropAbs` has map structure but no logic. Port: implement atom abstraction with `prev_expr` detection.

### [ ] 7.6 mc: Missing `mine_constants`, `expand_schemata`, `instantiate_axioms`, `to_table_lookup`
All stubs in Go. Port each following the Python implementations.

### [ ] 7.7 mc: Missing witness/trace reconstruction
**Python**: `ivy_mc.py:1445-1669`. `AigerMatchHandler`, `AigerMatchHandler2`, `aiger_witness_to_ivy_trace2` decode AIGER simulation state back to Ivy terms. **Go**: `MatchHandlerBase` has trivial eval/handle. Port: implement decode pipeline.

---

## 8. MEDIUM — Solver Infrastructure

### [ ] 8.1 solver: `lookup_native` / native interpretation infrastructure
**Python**: `ivy_solver.py:289-370`. Resolves native Z3 interpretations from `sig.interp`, handles polymorphic symbols, range sort clamped arithmetic, `bfe[...]`, `arrcst`. **Go**: `z3bridge.Translator` handles basic cases but none of the native interpretation lookup. Any Ivy program using `interpret X -> int`, `interpret X -> bv[32]`, etc. will fail. Port: implement native interpretation resolution in z3bridge or solver package.

### [ ] 8.2 solver: `term_to_z3` / `formula_to_z3_int` delegation needs verification
**Python**: `ivy_solver.py:414-647`. Direct recursive conversion handling: polymorphic symbol name mangling, native interpretation lookup, range sort clamped arithmetic, BV operations (bfe, concat), `handle_range_sorts` flag, non-Z3-enum binary encoding. **Go**: Delegates to `z3bridge.Translator.Translate()`. Need to verify z3bridge handles all these cases. Port: audit z3bridge and add missing cases.

### [ ] 8.3 solver: `UnsatCore` uses manual minimization
**Python**: `ivy_solver.py:696-710`. Uses Z3's assumption-based `check(assumptions)` + `unsat_core()` API, then calls `minimize_core` / `biased_core` from `ivy_core`. **Go**: Uses manual push/pop minimization. Port: switch to assumption-based API for efficiency.

### [ ] 8.4 solver: Missing BV operations (bfe, concat, shifts, etc.)
**Python**: `ivy_solver.py:162-222`. Maps `bvand`, `bvor`, `bvnot`, `concat`, `bfe` to Z3 lambdas. **Go**: z3convert.go has no BV operations. Port: add BV operation dispatch.

### [ ] 8.5 solver: `HerbrandModel` missing `sorted_sort_universe` and `numeral_assign`
**Python**: `ivy_solver.py:839, 1338`. `sorted_sort_universe` orders elements by the `<` relation. `numeral_assign` names universe elements respecting existing numerals. **Go**: Missing both. Port: implement ordering and numeral assignment.

### [ ] 8.6 solver: `RangeSortClampedAdd` returns placeholder
**Go**: z3convert.go:466 returns `IntVal(0)`. Also missing clamped subtract, multiply, divide. Port: implement clamped arithmetic using Z3 If/Then/Else with range bounds.

### [ ] 8.7 solver: `ClausesCase` missing unit resolution
**Python**: `ivy_solver.py:1430-1460`. Iterates with `ivy_unitres.UnitRes` and `clause_model_simp` until convergence. **Go**: herbrand.go:564 just picks disjuncts without unit propagation. Port: integrate unitres package.

---

## 9. MEDIUM — Proof System

### [ ] 9.1 proof: `MatchSchema` simplified — missing full matching pipeline
**Python**: `ivy_proof.py:380-470`. `setup_matching` builds `MatchProblem`, `fo_match` does first-order matching, then second-order matching with sort pair matching. **Go**: Uses simplified structural comparison (`EqualModAlpha` on conclusions). Port: implement `setup_matching`, `fo_match`, `compile_match` pipeline.

### [ ] 9.2 proof: Missing 7 proof tactics
`let_tactic`, `assume_tactic`, `unfold_tactic`, `if_tactic`, `property_tactic`, `function_tactic`, `witness_tactic` are all absent from Go. Port each following Python implementations.

### [ ] 9.3 proof: Missing `apply_match` with beta reduction
**Python**: `ivy_proof.py:510-540`. Full substitution with beta reduction and alpha-renaming to avoid capture. **Go**: Simplified substitution without beta reduction. Port: add beta reduction and alpha-renaming.

### [ ] 9.4 proof: Missing `detect_nonce_symbols`, `goal_nonce_definitions`
**Python**: `ivy_proof.py:600-650`. Detects fresh symbols in proofs. **Go**: Absent. Port for correct proof handling.

---

## 10. MEDIUM — Analysis Graph & Interpretation

### [ ] 10.1 art/art.go: Missing `initialize()`, `add_initial_state()`
**Python**: `ivy_art.py:120-180`. Full initialization with predicates, init_cond, initializer evaluation. **Go**: `AddState` exists but no `initialize` that sets up the initial state from module's init_cond. Port: implement initialization from module initial conditions.

### [x] 10.2 interp: `ApplyAction` uses `NullUpdate()` instead of `action.Update()`
**Python**: `ivy_interp.py:200-230`. Calls `action.update(domain, in_scope)` to compute the transition relation, then `compose_state_action` to get the post-state. **Go**: interp.go uses `NullUpdate()`. Port: wire to `actions.GetUpdate`.

### [ ] 10.3 interp: `Diagram()` returns raw clauses instead of minimal model diagram
**Python**: `ivy_interp.py:290-310`. Extracts minimal model diagram by finding satisfying assignment and computing diagram. **Go**: Returns combined clauses. Port: implement model extraction and diagram computation.

### [ ] 10.4 transrel: Missing interpolation functions
`interpolant()`, `forward_interpolant()`, `reverse_interpolant_case()`, `interpolant_case()`, `interp_from_unsat_core()` — all require solver integration. Port: implement using Z3's interpolation or Craig interpolation via unsat core.

### [ ] 10.5 transrel: `compose_state_action()` incomplete
**Python**: `ivy_transrel.py:350-400`. Composes state with action transition relation, checks precondition, returns post-state. **Go**: `ForwardImage` exists but the full precondition-checking version with `ActionFailed` is incomplete. Port: add precondition checking.

### [ ] 10.6 trace/trace.go: Trace construction from models is placeholder
`NewTraceStateFromEnv` passes nil equations, `value_to_str` only handles `*lg.Const`, `MakeVC` returns only preconditions. Port each following Python's full implementations.

---

## 11. MEDIUM — Supporting Subsystems

### [ ] 11.1 autoinst: `expand_schemata` and `match_schema_prems` missing
**Python**: `ivy_auto_inst.py:100-200`. Generator yielding match maps by recursively matching schema premises against sort constants and function symbols. **Go**: Only trigger-based matching, no schema expansion. Port: implement generator-style matching.

### [ ] 11.2 logicutil: Many Clauses utilities missing
`formula_to_clauses`, `clauses_to_formula`, `formula_to_clauses_tseitin`, `tseitin_encode`, `simplify_clauses`, `rename_clauses`, `and_clauses`, `or_clauses`, `ite_clauses`, `condition_clauses`, `negate_clauses`, `substitute_constants_clauses`, `rename_ast`, `resort_ast`. Most are in `clauseops` package but some are missing. Audit and port missing ones.

### [x] 11.3 module: Missing `init_cond`, `update_conjs()`, `call_graph()`
`init_cond` field (initialized to `lu.true_clauses()`), `update_conjs` generating concept spaces, `call_graph` building dependency graph. Port each.

### [ ] 11.4 module/theory.go: `TheoryContext.__call__()` is a no-op
**Python**: Instantiates non-EPR with ground terms. **Go**: Returns no-op cleanup. Port: implement ground-term instantiation.

### [x] 11.5 fragment: `makeFmlaPairFromAction` always returns false
Because `Action.update()` infrastructure isn't wired. Port: once action updates work (§4), wire into fragment checker.

### [ ] 11.6 typeinfer: `InsertSortVars` uses TopSort placeholders
**Go**: Cannot properly track SortVars inside FunctionSort. Port: extend FunctionSort to hold SortVar references.

### [ ] 11.7 ivyinit/ivyinit.go: `Initialize` is a placeholder
**Python**: `ivy_init.py:1-113`. Full initialization sequence: source file loading, import resolution, version detection. **Go**: Line 50 says "This function is a placeholder for the initialization sequence." Port: implement file loading and version detection.

### [ ] 11.8 webui: CTI minimization, sufficiency check, induction check stubs
`webui/ui_cti.go:328` "minimization not yet implemented", `:336` "sufficiency check not yet implemented", `:344` "induction check not yet implemented". Port from Python `ivy_ui_cti.py`.

### [ ] 11.9 webui/session.go: Redo not implemented
Line 260: `"redo not yet implemented"`. Port from Python's concept_interactive_session.

---

## 12. LOW — UI/Web, Secondary Backends, Tests

### [ ] 12.1 Python UI files not directly ported (expected — Go uses web UI instead)
Python has ~9,568 lines of Tk/widget UI code (`ivy_graph.py`, `ivy_graph_ui.py`, `ivy_ui.py`, `ivy_ui_cti.py`, `tk_cy.py`, `tk_graph_ui.py`, `tk_ui.py`, `widget_*.py`, `cy_*.py`). Go replaces these with `webui/` package. The Go webui covers session management, concept display, CTI exploration, and graph rendering, but is missing the interactive refinement operations (minimize, sufficiency check, induction check — see §11.8).

### [ ] 12.2 Dafny backend not ported
Python: `ivy_dafny_*.py` (5 files, ~1118 lines). Go: `dafnygen/dafnygen.go` exists as a stub. Low priority — rarely used.

### [ ] 12.3 Lean backend not ported
Python: `ivy_to_lean.py` (190 lines). Go: `leangen/leangen.go` exists as a stub. Low priority.

### [x] 12.4 SMT-LIB output not ported
Python: `ivy_smtlib.py` (30 lines). Missing from Go. Trivial to port.

### [ ] 12.5 Formula/term tables not ported
Python: `ivy_formulatab.py` (117 lines), `ivy_termtab.py` (117 lines). Used for hash-consing. Not critical for correctness but may be needed for performance.

### [ ] 12.6 z3_utils.py not fully ported
Python: `z3_utils.py` (197 lines). Utility functions for Z3 model inspection and simplification. Partially covered by `z3bridge/inspect.go`. Audit for gaps.

### [ ] 12.7 Concept space parser not fully ported
Python: `ivy_concept_space.py` (210 lines). Go: `conceptspace/` package exists. Verify completeness.

---

## Testing Gaps

### [ ] T.1 No integration tests that run full verify pipeline
The test suite has unit tests for individual packages but no end-to-end test that parses an .ivy file, compiles it, extracts an isolate, and runs verification. This is the most important missing test.

### [ ] T.2 No tests for action update semantics
`actions/update_test.go` exists (24 tests) but tests only `ActionUpdate` for atomic types, not `IntUpdate` for compound types (Sequence, If, While, Local, Call, Choice).

### [ ] T.3 No tests for C++ code generation output
`cppgen/cppgen_test.go` tests basic infrastructure but no golden-file tests comparing generated C++ against expected output.

### [ ] T.4 No tests for solver model extraction
No tests verify that `GetModelClauses` → `ClausesModelToClauses` → diagram construction produces correct results.

### [ ] T.5 No tests for isolate extraction
`isolate/isolate_test.go` tests basic utilities but not the full `IsolateComponent` or `CreateIsolate` pipeline.

### [ ] T.6 No tests for proof checker
`proof/proof_test.go` tests basic matching but not the full `ApplyProof` pipeline with tactics.

### [ ] T.7 No tests for L2S/temporal
No tests for the liveness-to-safety reduction or temporal proof tactics.

### [ ] T.8 No tests for model checking pipeline
No tests for AIGER encoding, quantifier elimination, or propositional abstraction.

### [ ] T.9 Missing tests for transrel forward/reverse image
`transrel/transrel_test.go` and `transrel/impl_test.go` exist but should verify forward/reverse image computation against known examples.

### [ ] T.10 No fuzz tests for parser edge cases
`lalr_logicparser/fuzz_crossval_test.go` exists for cross-validation but no fuzz testing of the main parser.

---

## Summary Statistics

| Category | Items | Estimated Lines to Port |
|----------|-------|------------------------|
| Critical — Core Verification | 7 | ~2,000 |
| High — Compiler Infrastructure | 10 | ~2,000 |
| High — Action Semantics | 11 | ~1,500 |
| High — Isolate Extraction | 5 | ~1,500 |
| High — Liveness/Temporal | 4 | ~2,000 |
| Medium — Model Checking | 7 | ~1,500 |
| Medium — Solver | 7 | ~1,000 |
| Medium — Proof System | 4 | ~800 |
| Medium — Analysis Graph | 6 | ~600 |
| Medium — Supporting | 9 | ~800 |
| Low — UI/Backends/Tests | 17 | ~2,000 |
| **Total** | **98** | **~18,200** |

We will omit all porting of C++ Code Generation.

