# goivy Port Audit — 18 March 2026

Systematic function-by-function comparison of the Python Ivy source of truth
(`/Users/jaten/pyivy/ivy/ivy/`) against the Go port (`goivy`).

**Legend:**
- **MISSING** — Python function/class exists but has no Go counterpart
- **STUB** — Go function exists but body is empty, placeholder, or skeletal
- **BEHAVIORAL_DIFFERENCE** — Both exist but produce different results in some cases
- **EQUALITY_ISSUE** — Python structural `==` vs Go pointer `==` mismatch

---

## Table of Contents

1. [Critical Cross-Cutting: Equality Bugs](#1-critical-cross-cutting-equality-bugs)
2. [Critical Cross-Cutting: TODO/STUB Inventory](#2-critical-cross-cutting-todostub-inventory)
3. [ivy_logic.py → logic/, ivylogic/](#3-ivy_logicpy--logic-ivylogic)
4. [ivy_logic_utils.py → logicutil/, clauseops/](#4-ivy_logic_utilspy--logicutil-clauseops)
5. [ivy_actions.py → actions/](#5-ivy_actionspy--actions)
6. [ivy_compiler.py → compiler/](#6-ivy_compilerpy--compiler)
7. [ivy_solver.py → solver/](#7-ivy_solverpy--solver)
8. [ivy_art.py → art/](#8-ivy_artpy--art)
9. [ivy_check.py → check/](#9-ivy_checkpy--check)
10. [ivy_isolate.py → isolate/](#10-ivy_isolatepy--isolate)
11. [ivy_module.py → module/](#11-ivy_modulepy--module)
12. [ivy_utils.py → ivyutils/](#12-ivy_utilspy--ivyutils)
13. [Summary Statistics](#13-summary-statistics)

---


## 2. Critical Cross-Cutting: TODO/STUB Inventory — UPDATE: 12 OF 21 ADDRESSED

Every TODO, STUB, FIXME, and "not implemented" found in the Go codebase:

### 2.1 Explicit TODOs

| # | File:Line | Description | Status |
|---|-----------|-------------|--------|
| 8 | `end2end/conform_test.go:232` | `TODO: compile and check with Go, compare against Python` | DEFERRED — test infrastructure |

---

## 5. ivy_actions.py → actions/

### 5.1 MISSING

| # | Python Function/Class | Description |
|---|----------------------|-------------|
| 1 | `DebugAction` in `IntUpdate` dispatch | No case in type switch; may fall through to wrong default |
| 2 | `AssignFieldAction.action_update`, `NullFieldAction.action_update`, `CopyFieldAction.action_update` | No `ActionUpdate` methods — fall through to null-update |
| 3 | `make_field_update` helper | Constructs field-update `AssignAction` for struct-like fields |
| 4 | `IfAction.subactions()` | Complex method handling `Some`/`SomeMinMax` conditions |
| 5 | `IfAction.get_cond()` | Extracts effective condition (handles `Some` existential quantification) |
| 6 | `CallAction.split_returns()` | Decomposes call with returns into separate assignments |
| 7 | `CallAction.prefix_calls()` callable renamer | Only handles string prefix, missing callable-renamer case |
| 8 | `Action.iter_internal_defines()` | Yields internally defined symbols |
| 9 | `Action.get_type_names()` | Collects type names from `LocalAction` declarations |
| 10 | `Schema.get_instance()` | Full substitution+compilation. Go's `Instantiate` only appends to list. |
| 11 | `TypeCheckContext` class | `ActionContext` subclass replacing callees with null actions |
| 12 | `checked_assert`, `check_unprovable` parameters | Selective assertion checking not implemented |
| 13 | `LabeledFormula`/`unprovable` handling in `AssumeAction`/`AssertAction` | Not handled |
| 14 | `ActionContext` context manager with global `context` | Go uses function pointer instead of context stack |
| 15 | `SymExContext` | Go version has completely different semantics |

### 5.2 STUB

| # | Go Function | Issue |
|---|-------------|-------|
| 1 | `instantiateMacro` (`extra_actions.go:805-829`) | Returns nil; "full macro expansion requires AST-level rewriting" |
| 2 | `InstantiateAction.IntUpdate` | Macro path always nil; schema path simplified |
| 3 | `mkVariantAssignClauses` (`update.go:624-633`) | "Simplified: treat as regular assignment" — missing `pto` constraints |
| 4 | `destructorAssignUpdate` (`update.go:549-603`) | Missing nested destructor handling and frame conditions |
| 5 | `isDestructor` (`transforms.go:133-137`) | Always returns `false` |
| 6 | `SetAction.ActionUpdate` (`update.go:729-764`) | Missing frame condition for non-set indices |
| 7 | `WhileAction.Decompose` (`action.go:952-958`) | "Simplified: treat body as single step" |
| 8 | `ChoiceAction.IntUpdate` | Missing `determinize` check |
| 9 | `EnvAction.IntUpdateEnv` | Missing `determinize` check |

### 5.3 BEHAVIORAL_DIFFERENCE

| # | Area | Python | Go | Impact |
|---|------|--------|----|----|
| 10 | `AssertAction.action_update` | `dual_formula` → clausify → wrap with `EmptyAnnotation` | `Negate(fmla)` raw without clausification | Different CNF decomposition |
| 11 | `AssumeAction.action_update` | `formula_to_clauses_tseitin(skolemize_formula(fmla))` + `unfold_definitions_clauses` | Only `skolemizeFormula` | Missing Tseitin + definition unfolding |
| 12 | `VarAction` | `AST` subclass, NOT an `Action` | Full `Action` with `ActionBase` | Structural misrepresentation |
| 13 | `SubgoalAction` | Extends `AssertAction`, has `kind` | Separate struct, no assert semantics | Different inheritance |
| 14 | `CopyFieldAction` | 4 args (`l, lf, r, rf`) | 3 fields (`Field, Dst, Src`) | Missing second field name |
| 15 | `WhileAction.unroll` | Determines index sort, queries cardinality, guards at 100 | Fixed integer bound, no index sort | Different unrolling strategy |
| 16 | `AssignAction.action_update` | Extends partial applications, checks variables in RHS | No partial-application extension, no variable check | Missing validation |

---

## 6. ivy_compiler.py → compiler/

### 6.1 MISSING

| # | Python Function | Description |
|---|-----------------|-------------|
| 1 | `IvyDomainSetup.parameter` | Parameter declarations silently dropped |
| 2 | `IvyDomainSetup.destructor` | Standalone destructor declarations unhandled |
| 3 | `IvyDomainSetup.constructor` | Constructor declarations unhandled |
| 4 | `IvyDomainSetup.concept` | Concept space declarations not handled |
| 5 | `IvyDomainSetup.rely` | Rely declarations not handled |
| 6 | `IvyDomainSetup.mixord` | Mix-order declarations not handled |
| 7 | `IvyDomainSetup.update` | Update declarations not handled |
| 8 | `IvyDomainSetup.scenario` + `IvyARGSetup.scenario` | Scenario/state-machine declarations not handled |
| 9 | `IvyDomainSetup.implementtype` | `implementtype` declarations not handled |
| 10 | `IvyARGSetup.state` | State declarations not handled |
| 11 | `add_definition` variable-duplication checks | No validation of LHS variable uniqueness or RHS free vars |
| 12 | `DerivedUpdate` creation | Not created for derived/definition declarations |
| 13 | `opt_mutax` parameter | Mutually exclusive axiom checking not ported |
| 14 | `compile_theory` in `interpret` | Not called for int/range/solver sort interpretations |

### 6.2 STUB (Empty bodies in Go)

| # | Go Function | Python Equivalent | What It Should Do |
|---|-------------|-------------------|-------------------|
| 1 | `FixConstructors` | `fix_constructors` | Adjust constructor domain sorts from destructor sorts |
| 2 | `CreateSortOrder` | `create_sort_order` | Topological sort with Tarjan SCC, error on cycles |
| 3 | `CreateConstructorSchemata` | `create_constructor_schemata` | Constructor existence axiom schemata |
| 4 | `AttachProofs` | `attach_proofs` | Match labeled proofs to properties/conjectures |
| 5 | `CheckDefinitions` | `check_definitions` | Separate defs from props, check redefinition, detect cycles |
| 6 | `CheckPropertiesPass` | `check_properties` | Full proof checking with `prover.AdmitProposition` |
| 7 | `CreateConjActions` | `create_conj_actions` | Determine which actions must preserve each conjecture |
| 8 | `HandleTemporals` | `handle_temporals` | Label actions with isolate membership |
| 9 | `ApplyAssertProof` | Phase6 | Generate subgoals from ProofChecker |
| 10 | `InferParameters` | Phase6 | Parameter extension (extra params appended, body rewritten) |
| 11 | `TarjanArcs` | Phase6 | Only filters self-loops instead of running full Tarjan SCC |
| 12 | `GetSymbolDependencies` | Phase6 | Not transitive — only collects direct symbols |
| 13 | `TheoremToProperty` | Phase6 | Trivial copy instead of full skolemization + sort renaming |
| 14 | `compile_thunk_action` | Phase6 | Skips subtype/destructor/substitution logic |

### 6.3 BEHAVIORAL_DIFFERENCE

| # | Area | Issue |
|---|------|-------|
| 15 | `compile_action_def` | Missing `prm:` prefix substitution for action parameters |
| 16 | `compile_action_def` | Missing free-variable check in call arguments |
| 17 | `compile_local` | Missing assignment-as-declaration sort inference optimization |
| 18 | `compile_call` | Missing field-reference fallback and parameter count validation |
| 19 | `compile_if_action` | `Some`/`SomeMinMax` existential-if not handled |
| 20 | `compile_while_action` | No validation against action calls in conditions |
| 21 | `pullArgs` | Silently returns available args instead of raising IvyError |
| 22 | Ghost sorts | Not tracked via `ghost_sorts` |
| 23 | Struct destructors | Created differently (no intermediate Variable parameter) |
| 24 | `variant` | Only populates `Variants`, not `Supertypes` |
| 25 | `export`/`import_` | No `check_is_action` validation |
| 26 | `attribute` | No validation of object existence or attribute name |
| 27 | `native` | Stores raw node without `compile_native_def` |
| 28 | `mixin` | No validation that mixee is `init` or in `top_context.actions` |
| 29 | `ConjSetup` pass 2 | All cases ignore args (`_ = arg`) — no-op |
| 30 | `ARGSetup` pass 3 | Ignores exports, delegates, progress properties |
| 32 | `ExprContext.extract()` | Missing `LocalAction` wrapping for local symbols |

---

## 7. ivy_solver.py → solver/

### 7.1 MISSING

| # | Python Function | Description |
|---|-----------------|-------------|
| 1 | `clear()` | Reset global Z3 caches. No Go equivalent. |
| 2 | `uninterpretedsort()`/`functionsort()`/`enumeratedsort()` | Sort-to-Z3 conversion with caching and `z3_sorts_inv` reverse map |
| 3 | `sorts()` | BV, array, strbv, intbv, int, real, nat, strlit sort handling |
| 4 | `atom_to_z3()` | Enumerated equality encoding, polymorphic macro expansion |
| 5 | `term_to_z3()` | Variable name caching with `:sort_name` suffix convention |
| 6 | `formula_to_z3_int()` | Def True/False simplification, quantifier constraints for nat/range |
| 7 | `sort_from_z3()` | Reverse map from Z3 sorts to Ivy sorts |
| 8 | `HerbrandModel.universes()` | Returns dict from sorts to universe elements |


### 7.2 STUB

| # | Go Function | Issue |
|---|-------------|-------|
| 1 | `RangeSortBounds()` | Returns hardcoded `(0, MaxInt32)` instead of parsing real bounds |
| 2 | `EncodeTerm()` | Only handles EnumeratedSort cardinality; missing Ite, constructors, variables |
| 3 | `EncodeEquality()` | Simple Eq instead of binary encoding with `gebin()` |
| 4 | `ClausesModelToDiagram()` | **Creates trivial `X=X` equalities** — always true, doesn't capture model |
| 5 | `ClausesModelToClauses()` | Only handles 0-arity constants; missing functions, relations, post-processing |
| 6 | `CheckNativeCompatSym()` | Only checks sort compatibility, no runtime type verification |
| 7 | `bfeToZ3()` | Missing IntSort input, BV size overflow, zero-width, zero-extension |
| 8 | `SolverName()` | Returns empty string for `z3_builtins` names instead of raising error |

### 7.3 BEHAVIORAL_DIFFERENCE (Critical for Correctness)

| # | Area | Python | Go | Impact |
|---|------|--------|----|----|
| 1 | **Quantifier bound constraints** | Adds nat/range constraints INSIDE quantifier body (`ForAll(vs, Implies(constraints, body))`) | Adds constraints at clause level for symbols, NOT inside quantifiers | **Different Z3 behavior for nat/range-sorted quantified variables** |
| 2 | **`clause_model_simp`** | Returns single true literal (clause satisfied) | Keeps all true/unknown literals | Different iterative simplification |
| 3 | **`clauses_case` unit resolution** | Performs `UnitRes` propagation between rounds | Skips unit resolution entirely | Less effective simplification |
| 4 | **`model_if_none` sort size search** | All sorts at same size simultaneously | Minimizes each sort independently | May find different (larger) models |
| 5 | **`numeral_to_z3` range clamping** | Clamps: `If(val < lb, lb, If(ub < val, ub, val))` | No clamping | Out-of-range numerals not clamped |
| 6 | **`filter_redundant_facts`** | Activation literals with assumption-based checking | Push/pop with direct assertion | Different redundancy detection |
| 7 | **`my_eq`** | True/False simplification for boolean equality | Only bool→Iff dispatch | Missing simplification |
| 8 | **`gebin`** | Binary predicate encoding of enumerated sorts | Simple integer comparison `Ge(x, IntVal(bound))` | Semantically different encoding |
| 9 | **Variable naming** | Variables cached as `v.rep + ':' + v.sort.name` | Unknown convention | Possible name collisions between sorts |
| 10 | **`check_sequence`** | Reporter with start/end callbacks, early abort | No reporter, no early abort | Cannot abort sequence |
| 11 | **`sort_card`** | Handles BV sorts (`2**size`), datatypes, ranges | Only handles `EnumeratedSort.Card()` | Missing cardinality for BV/range sorts |

---

## 8. ivy_art.py → art/

### 8.1 MISSING

| # | Python Function | Description |
|---|-----------------|-------------|
| 1 | `add_initial_state(ic, abstractor)` params | Always uses `mod.InitCond`, never applies abstractor |
| 2 | `state_actions(state)` | Returns action equations for a state |
| 3 | `do_state_action(equation, abstractor)` | Evaluates state equation with abstraction |
| 4 | `recalculate_state(state, abstractor)` | Recalculates state from predecessors |
| 5 | `show_core(clause_str, state)` | Shows UNSAT core for a clause |
| 6 | `concept_graph(state, ...)` | Creates concept graph from a state |
| 7 | `make_concrete_trace(state, conc)` | Concrete trace generation |
| 8 | `decompose_edge(transition)` | Decomposes transition |
| 9 | `state_extensions(state, join)` | State extensions for refinement |
| 10 | `as_cy_elements(dot_layout)` | CyElements rendering for GUI |

### 8.2 STUB

| # | Go Function | Issue |
|---|-------------|-------|
| 1 | `FixedpointCandidate()` | Returns map without performing state join (`bottom_state` logic missing) |
| 2 | `CheckConstraints` | Stub |
| 3 | `StratifyGoals` | Stub |

### 8.3 BEHAVIORAL_DIFFERENCE

| # | Area | Issue |
|---|------|-------|
| 1 | `Add()` | No assertion checking for ActionApp validity |
| 2 | `Cover()` | Uses Z3 implication instead of `domain.order()` |
| 3 | `Delete()` | Bug: uses `state.ID + 1` after ID already set to `-1` |
| 4 | `BMC()` | Simplified SAT check instead of `history_satisfy` with path extraction |
| 5 | `CheckSafety()` | Missing expression evaluation and `IvyActionFailedError` handling |
| 6 | `AddInitialState()` | Executes each initializer separately (adds intermediate states) instead of single Sequence |
| 7 | `PostState()` | Simple `And(preFmla, trNode)` instead of `concrete_post` with proper havocing |

---

## 9. ivy_check.py → check/

### 9.1 MISSING

| # | Python Function | Description |
|---|-----------------|-------------|
| 1 | `check_properties()` with `itp.false_properties()` | No actual property checking |
| 2 | `check_conjectures(kind, msg, ag, state)` | No-op |
| 3 | `show_counterexample(ag, state, bmc_res)` | Placeholder message only |
| 4 | `gui_art(other_art)` | No GUI |
| 5 | `check_temporals()` | Stub — no temporal property verification |
| 6 | `check_subgoals(goals, method)` | Trivial stub |
| 7 | `mc_tactic`, `vmt_tactic` | Return nil, not registered as proof tactics |
| 8 | `MatchHandler` class | Stub implementations for most methods |
| 9 | `convert_postconds(state, postconds)` | Returns postconds unchanged |
| 10 | `check_fcs_in_state()` trace/diagnose path | Only basic SAT/UNSAT path |
| 11 | `start()` | Returns error "not yet fully integrated" |

### 9.2 STUB

| # | Go Function | Issue |
|---|-------------|-------|
| 1 | `CheckProperties` | Promotes properties to axioms without checking |
| 2 | `CheckConjectures` | Returns nil |
| 3 | `CheckTemporals` | Prints status, no actual verification |
| 4 | `ApplyConjProofs` | Proof branch identical to no-proof branch |
| 5 | `PreprocessAssumedIgnoredProperties` | Temporal property handling incorrect |

### 9.3 BEHAVIORAL_DIFFERENCE

| # | Area | Issue |
|---|------|-------|
| 1 | `Checker.Sat()` | Always calls `Fail()` — should call `Pass()` when `check_unprovable` is true |
| 2 | `CheckIsolate` initialization | Separate `Initialize` call vs constructor parameter |
| 3 | `CheckModule` | Copy handling may not handle all side effects |
| 4 | `check_fcs_in_state` | Different solver interaction pattern (push/pop vs `history.satisfy`) |

---

## 10. ivy_isolate.py → isolate/

### 10.1 MISSING

| # | Python Function | Description |
|---|-----------------|-------------|
| 1 | `check_isolate_completeness()` full impl | Only checks property labels, missing caller/callee assertion verification |
| 2 | `has_assertions`/`has_requires` | Fragile type assertion on `interface{}` instead of `isinstance(action, AssertAction)` |
| 3 | `strip_action()` full complexity | Missing interference checking (`modifies()`), `init_params`, `strip_binding` |
| 4 | `strip_labeled_fmla()` strip binding | No `get_strip_binding` call |
| 5 | `isolate_component()` create_imports | Import actions, out-calls, external stubs not implemented |
| 6 | `isolate_component()` `slv.check_compat()` | Native interpretation checking missing |
| 7 | `isolate_component()` `canonize_types` | Not called |

### 10.2 STUB

| # | Go Function | Issue |
|---|-------------|-------|
| 1 | `GetIsolateAttr` | Always returns `defaultVal` |
| 2 | `isExplicitOnly()` | Always returns false |
| 3 | `IsolateComponent()` ExtAction | Sets flag but doesn't create combined action |

### 10.3 BEHAVIORAL_DIFFERENCE

| # | Area | Issue |
|---|------|-------|
| 1 | Action classification | String-based kinds (`"assert"`) vs Python type references (`ia.AssertAction`) |
| 2 | `check_interference()` | Missing Ranking/WhileAction loop/termination checking |
| 3 | `strip_isolate()` | Missing version 1.6 `mod.params` clearing |
| 4 | `get_mixin_order()` | Different name extraction via `relNamer` interface vs `.relname` |
| 5 | `FixInitializers` | Missing `type_check_action` call |

---

## 11. ivy_module.py → module/

### 11.1 MISSING

| # | Python Function | Description |
|---|-----------------|-------------|
| 1 | `Module.__enter__`/`__exit__` | Does not save/restore `il.sig` or call `solver.clear()` |
| 2 | `Module.InitCond` initialization | Not initialized to `lu.true_clauses()` in `Clear()` |
| 3 | `CanonizeTypes` fields | Missing: `InitCond`, `ConceptSpaces`, `Progress`, `BeforeExport`, `lu.resort_sig` |
| 4 | `find_action` free function | Only exists as method, not package-level |
| 5 | `background_theory` free function | Only exists as method |
| 6 | `param_logic` parameter and `logics()` | Hardcodes `["epr"]` default |
| 7 | `sort_dependencies` NativeTypes branch | Skips native_types check |
| 8 | `CallGraph()` | Placeholder returning empty map |
| 9 | `UpdateConjs()` | Placeholder, no concept space creation |
| 10 | `copy()` | Missing many map copies: `SortDestructors`, `SortConstructors`, `Variants`, `Supertypes`, `ExtPreconds`, `ConjActions`, `IsolateProofs` |

### 11.2 BEHAVIORAL_DIFFERENCE

| # | Area | Issue |
|---|------|-------|
| 1 | `BackgroundTheory` | Package-level `theoryCache` with mutex (memory leak) instead of `self.theory` field |
| 2 | `SortCard` | `val.(string)` type assertion vs `self.attributes[attr].rep` |
| 3 | `VariantIndex` return | Go returns `-1`, Python returns `None` |
| 4 | `resort_aliases_map` | Go correctly returns result; Python has bug (no `return`) |
| 5 | `Exclusivity()` | Missing coverage axiom (disjunction that some variant must hold) |

---

## 12. ivy_utils.py → ivyutils/

### 12.1 MISSING (19 items — many are error/location infrastructure)

| # | Python Function/Class | Description |
|---|----------------------|-------------|
| 1 | `IvyError` exception | No Go equivalent with lineno/reference chain support |
| 2 | `IvyUndefined` exception | Subclass of IvyError |
| 3 | `ErrorList` class | No Go equivalent |
| 4 | `ErrorPrinter` context manager | No Go equivalent |
| 5 | `SourceFile` context manager | Manages global `filename` |
| 6 | `WorkingDir` context manager | No Go equivalent |
| 7 | `Location()`/`LocationTuple`/`nowhere()`/`is_nowhere()`/`lineno_str()` | Error reporting infrastructure entirely absent |
| 8 | `warn(ast, msg)` | No Go equivalent |
| 9 | `parse_with()`/`p_error()` | Parser error handling |
| 10 | Combinator functions | `apply_func_to_list`, `gen_list`, etc. |
| 11 | List utilities | `concat`, `unzip_append`, `flatten`, `union_of_list`, etc. |
| 12 | `pretty(s, max_lines)` | Code pretty-printer |
| 13 | `parse_int_subscripts(name)` | No Go equivalent |
| 14 | `distinct_obj_renaming(names1, names2)` | No Go equivalent |
| 15 | Version functions | `get_numeric_version()`, `string_version_to_numeric_version()`, `version_le()` |
| 16 | Global parameters | `use_numerals`, `use_new_ui`, `catch`, `default_ui`, `enable_debug` |
| 17 | `dbg()` | Debug printing |
| 18 | `get_default_ui_module()`/`get_default_ui_class()` | UI module functions |

### 12.2 BEHAVIORAL_DIFFERENCE

| # | Area | Python | Go |
|---|------|--------|-----|
| 1 | `SetStringVersion` compose char | version <= [1,1] uses `:` | version < 1.3 uses `"__"` — different cutoff AND different character |
| 2 | `GetStdIncludeDir` | Scans version-compatible subdirectories | `os.Stat("include")` relative to CWD — much simpler, often fails |

---

## 13. Summary Statistics

| Subsystem | Missing | Stub | Behavioral Diff | Equality Issue | Total |
|-----------|---------|------|-----------------|----------------|-------|
| **Cross-cutting TODO/stub** | — | ~~21~~ **8 remaining (13 FIXED/addressed)** | — | — | ~~21~~ **8** |
| ivy_logic | 6 | 1 | 10 | 3 | 20 |
| ivy_logic_utils | 6 | 0 | 5 | 0 | 11 |
| ivy_actions | 15 | 9 | 7 | 0 | 31 |
| ivy_compiler | 14 | 14 | 18 | 0 | 46 |
| ivy_solver | 9 | 8 | 11 | 0 | 28 |
| ivy_art | 10 | 3 | 7 | 0 | 20 |
| ivy_check | 11 | 5 | 4 | 0 | 20 |
| ivy_isolate | 7 | 3 | 5 | 0 | 15 |
| ivy_module | 10 | 0 | 5 | 0 | 15 |
| ivy_utils | 18 | 0 | 2 | 0 | 20 |
| **TOTAL** | **106** | **64** | **74** | **18** | **262** |

### Top Priority Fixes (Correctness Impact)

3. **`ClausesModelToDiagram` creates `X=X`** — trivially true, doesn't capture model (§7.2, item 4)
4. **Quantifier bound constraints** not inside quantifier body for nat/range sorts (§7.3, item 1)
5. **`ExpandAbbrevs` missing Ite case** — affects all clause generation with Ite nodes (§4.2, item 7)
6. **14 empty compiler stubs** — `FixConstructors`, `CreateSortOrder`, `CreateConstructorSchemata`, `AttachProofs`, `CheckDefinitions`, `CreateConjActions`, `HandleTemporals`, etc. (§6.2)
7. **`check/` package largely non-functional** — `CheckProperties`, `CheckConjectures`, `CheckTemporals` are no-ops (§9)
10. **`art/art.go` Delete() bug** — uses `state.ID` after setting it to `-1` (§8.3, item 3)

### Medium Priority (Behavioral Fidelity)

11. Missing `prm:` prefix substitution in `compile_action_def` (§6.3, item 15)
12. Missing unit resolution in `clauses_case` (§7.3, item 3)
13. Missing range clamping in `numeral_to_z3` (§7.3, item 5)
14. `BooleanSort` string representation `"Boolean"` vs `"bool"` (§3.2, item 10)
15. `de_morgan` not calling `expand_abbrevs` first (§4.2, item 9)
16. `SetStringVersion` compose character mismatch (§12.2, item 1)
17. `GetStdIncludeDir` too simplistic (§12.2, item 2)
18. Missing field action `ActionUpdate` methods (§5.1, items 2-3)
19. Missing `Some`/`SomeMinMax` handling in if/while compilation (§6.3, items 19-20)
20. `PostState` missing proper havocing (§8.3, item 7)
