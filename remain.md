
### 2.1 Explicit TODOs

| # | File:Line | Description | Status |
|---|-----------|-------------|--------|
| 8 | `end2end/conform_test.go:232` | `TODO: compile and check with Go, compare against Python` | DEFERRED — test infrastructure |

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
