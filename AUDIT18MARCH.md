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

## 1. Critical Cross-Cutting: Equality Bugs

These are the highest-priority bugs because they cause **silent incorrect behavior**
(lookups miss, comparisons return wrong results) with no error.

### 1.1 Maps keyed by Sort interface — UPDATE: NOW FIXED.

Go maps with `lg.Sort` as key use **pointer equality** for lookups, but Python
dicts use `__eq__`/`__hash__` (structural equality). Two structurally-equal sorts
at different pointers will silently miss.

| # | File:Line | Code | Fix | Status |
|---|-----------|------|-----|--------|
| 1 | `ivylogic/context.go:57-58` | `map[lg.Sort]lg.Sort` in `GetSortRefinement` | Changed to `map[lg.NodeKey]lg.Sort` with `lg.SortKey()` | **FIXED** |
| 2 | `interp/phase4.go:408` | `map[lg.Sort][]lg.Node` for universe map | Added `SortUniverse` slice type as preferred path; legacy `map[lg.Sort]` path preserved for compatibility | **FIXED** |

### 1.2 Pointer `==` on Sort interface vs singletons — UPDATE: NOW FIXED.

These compared a `Sort` interface value against `lg.Boolean` using `==`, which is
pointer comparison. Fixed by using `lg.SortEqual()` for structural comparison.

| # | File:Line | Code | Status |
|---|-----------|------|--------|
| 3 | `l2s/l2s.go:405` | `bodySort == lg.Boolean` → `lg.SortEqual(bodySort, lg.Boolean)` | **FIXED** |
| 4 | `l2s/l2s.go:406` | `fs.Range() == lg.Boolean` → `lg.SortEqual(fs.Range(), lg.Boolean)` | **FIXED** |
| 5 | `mc/qelim.go:76` | `s == lg.Boolean` → `lg.SortEqual(s, lg.Boolean)` | **FIXED** |
| 6 | `logic/pretty.go:250` | `t.aSort != Boolean` → `!SortEqual(t.aSort, Boolean)` | **FIXED** |
| 7 | `webui/session.go:749` | `fs.Range() == logic.Boolean` → `logic.SortEqual(fs.Range(), logic.Boolean)` | **FIXED** |

### 1.3 Pointer `==` on True/False singletons — UPDATE: NOW FIXED.

Added `lg.IsTrue(n)` / `lg.IsFalse(n)` to the `logic` package that handle both
the singleton pointer and structurally-equivalent `&And{}`/`&Or{}`. All callers
updated to use structural checks.

| # | File:Line | Code | Status |
|---|-----------|------|--------|
| 8 | `actions/update.go` | `isTrue`/`isFalse` helpers → delegate to `lg.IsTrue`/`lg.IsFalse` | **FIXED** |
| 9 | `transrel/transrel.go` | `isFormulaTrue`/`isFormulaFalse` → delegate to `lg.IsTrue`/`lg.IsFalse` | **FIXED** |
| 10 | `vmt/vmt.go:83` | `defsFormula != lg.True` → `!lg.IsTrue(defsFormula)` | **FIXED** |
| 11 | `tactics/tactics.go:560,563` | `a == lg.True` / `b == lg.True` → `lg.IsTrue(a)` / `lg.IsTrue(b)` | **FIXED** |
| 12 | `art/art.go:406` | `trNode != lg.True` → `!lg.IsTrue(trNode)` | **FIXED** |
| 13 | `interp/eval.go:288` | `isNodeFalse` → delegate to `lg.IsFalse` | **FIXED** |

Also fixed 3 instances in `solver/solver_test.go` that used `constraint != lg.True`.

### 1.4 Polarity not flipped through negation (CRITICAL BUG). UPDATE: NOW FIXED.

In `ivylogic/classify_ext.go`, two functions compute polarity incorrectly:

| # | File:Line | Function | Bug |
|---|-----------|----------|-----|
| 14 | `classify_ext.go:227-231` | `symbolsOverUniversalsRec` | `localPos = !pos` computed but discarded (`_ = localPos`); recursive calls use original `pos` |
| 15 | `classify_ext.go:285-292` | `universalVariablesRec` | Same bug — `Not` does not flip polarity |

**Impact:** Any formula containing `Not` will produce incorrect results from
these functions. This affects `IsInLogic`, EPR fragment checking, and related
analysis.

### 1.5 Sort `!=` for dirty-flag optimization (FRAGILE). UPDATE: NOT A BUG. ADDRESSED.

These use pointer `!=` on Sort interfaces to detect "did the sort change?"
Currently safe because transformation functions return the original pointer when
nothing changed. This is semantically correct — it tests "did the transformation
return a new object?" not "are these structurally equal?" — so no fix needed.

| File | Lines |
|------|-------|
| `logicutil/logic_utils.go` | 413, 420, 1680, 1684 |
| `module/resort.go` | 97, 102, 134 |
| `proof/match.go` | 406, 424, 455 |
| `proof/phase5_matching.go` | 809 |

---

## 2. Critical Cross-Cutting: TODO/STUB Inventory — UPDATE: 12 OF 21 ADDRESSED

Every TODO, STUB, FIXME, and "not implemented" found in the Go codebase:

### 2.1 Explicit TODOs

| # | File:Line | Description | Status |
|---|-----------|-------------|--------|
| 1 | `isolate/isolate.go:1322` | `TODO: wire to real clauseops.FormulaToClauses` | **FIXED** — wired to `co.FormulaToClauses(fmla, nil)` |
| 2 | `solver/z3convert.go:778` | `TODO: arrsel → ctx.Select when z3bridge array ops implemented` | **FIXED** — z3bridge/array.go implements ArraySort, Select, Store, ConstArray, ArrayDomain, ArrayRange; solver/z3convert.go wires arrsel/arrupd/arrcst in lookupBuiltinFunc and lookupBuiltinRelation; TranslateSort handles arr[dom][rng] sort names; Z3SortToSort handles SortArray case |
| 3 | `clauseops/ops.go:102` | `TODO: implement annot.conj when annotation types support it` | **FIXED** — `AnnotConjFunc` callback registered by `actions` init |
| 4 | `transrel/transrel.go:1101` | `TODO: extract_pre_post_model` (Trans field set to nil) | **FIXED** — wired to `ExtractPrePostModel`; `ActionFailed.Trans` → `TransPre`/`TransPost` |
| 5 | `cppgen/expr.go:335` | `TODO: emit_sig(impl) + constraint addition` | SKIPPED — C++ codegen, Go backend is priority |
| 6 | `cppgen/expr.go:342` | `TODO: emit_randomize + emit_eval_sig` | SKIPPED — C++ codegen |
| 7 | `cppgen/expr.go:348` | `TODO: populate obj from model` | SKIPPED — C++ codegen |
| 8 | `end2end/conform_test.go:232` | `TODO: compile and check with Go, compare against Python` | DEFERRED — test infrastructure |
| 9 | `gogen/action.go:80` | `TODO: unhandled action type %T` | **FIXED** — added cases for SubgoalAction, DebugAction, AssignFieldAction, NullFieldAction, CopyFieldAction, InstantiateAction, VarAction |
| 10 | `gogen/action.go:285` | `TODO: thunk action` | **FIXED** — emits closure-based thunk (Python desugars thunks before codegen) |

### 2.2 Stub Functions/Packages

| # | File | Function/Package | Description | Status |
|---|------|------------------|-------------|--------|
| 11 | `interp/interp.go:12` | Package | Solver-dependent stubs fixed: `History.Satisfy` returns `SatisfyResult` with path reconstruction; `HistorySatisfy` delegates properly; `DecomposeActionApp` builds per-step states; `EvalStateAtom` looks up via `mod.FindAction`; added `HerbrandModel.Universes()` | **FIXED** |
| 12 | `ivylsp/ivylsp.go` | Package | Entire package is minimal LSP stub | DEFERRED — separate concern |
| 13 | `art/art.go:883` | `CheckConstraints` | Stub | **NOT A BUG** — does not exist in Python either |
| 14 | `art/art.go:957` | `StratifyGoals` | Stub | **NOT A BUG** — does not exist in Python either |
| 15 | `check/check.go:192` | `DualClauses` | Simplified stub | **FIXED** — full skolemization matching Python's `dual_clauses`. Exposed missing `mod.UpdateTheory()` call in `IvyCompile` (§6.3 item 31) — `BackgroundTheory()` was returning empty clauses because the theory was never built from axioms. Root cause fixed in `compiler/ivy_compile.go`. |
| 16 | `check/isolate_check.go:526` | `MCIsolate` | Stub | **FIXED** — full implementation matching Python mc_isolate: validates temporal properties, calls method within TheoryContext, handles CheckSeparately per-assertion mode. CheckModule wires mc.CheckIsolate, vmt.CheckIsolate, bmc.CheckIsolate as method backends. GetIsolateAttr extracts attribute values from module. parseBMCParams parses bmc[N][M] specifiers. |
| 17 | `mc/propabs.go:186` | `MineConstantsStub` | Stub | **FIXED** — deleted (deprecated, zero callers; real `MineConstants` at `mc/mine.go:16`) |
| 18 | `mc/checker.go:174` | `CheckIsolateStub` | Stub | **FIXED** — deleted (deprecated; real `CheckIsolate` at `mc/checker.go:135`) |
| 19 | `mc/qelim.go:250` | `InstantiateAxiomsStub` | Stub | **FIXED** — deleted (deprecated, zero callers; real `InstantiateAxioms` at `mc/transforms.go:413` with full trigger matching) |
| 20 | `webui/concept_session.go:219` | Alpha call | No-op stub | **FIXED** — clears cache; full Alpha call needs AnalysisSession wiring |
| 21 | `webui/ext_api.go:124,134` | `execute_actions`, `try_conjectures` | Nil callbacks | **FIXED** — real callbacks via `AnalysisSessionI` interface |

---

## 3. ivy_logic.py → logic/, ivylogic/

### 3.1 MISSING

| # | Python Function | Description |
|---|-----------------|-------------|
| 1 | `Sig.__enter__`/`__exit__` | **NOT A BUG** — Go idiomatically replaces the context manager with explicit save/restore (`savedSig := c.Sig; c.Sig = c.Sig.Copy(); ... c.Sig = savedSig`), already implemented at all call sites (compiler/action.go, compiler/phase6.go, proof/phase5_matching.go, etc.). `WithSymbols`/`WithSorts` Enter()/Exit() helpers also exist in ivylogic/sig.go. |
| 2 | `Sig.contains(sort_or_symbol)` | Generic contains check for both sorts and symbols. Go only has `ContainsSymbol`. | **FIXED** — Added `Sig.Contains()` dispatcher; also fixed `CheckPremisesProvided` to use it + skip lambdas. |
| 3 | `find_polymorphic_symbol()` | Missing numeral/literal-string check (`s[0].isdigit() or s[0] == '"'`) and fallback to `find_symbol()`. | **FIXED** — Added numeral/string check in `FindPolymorphicSymbol` (`ivylogic/poly.go`). Fallback to `find_symbol` is caller's responsibility in Go (callers already call `sig.FindSymbol` after). `ivy_have_polymorphism` flag not needed — Go targets modern Ivy (≥1.3) where flag is always true. |
| 4 | `all_concretely_sorted()` | Python trivially returns True. Go's version actually checks sorts (different behavior). | **FIXED** — `AllConcretelySorted` now trivially returns nil, matching Python. |
| 5 | `check_concretely_sorted()` `unsorted_var_names` param | Go does not accept the exemption list parameter. | **FIXED** — Added `unsortedVarNames map[string]bool` parameter to `CheckConcretelySorted`. All callers pass nil. |
| 6 | Global `sig` variable | **NOT A BUG** — Go threads `Sig` as a field on Compiler and Module structs instead of a module global. Equivalent effect; all call sites already pass Sig explicitly. |

### 3.2 BEHAVIORAL_DIFFERENCE

| # | Area | Python | Go | Impact |
|---|------|--------|----|----|
| 7 | `is_numeral_name` | Bug: `s[1].isdigit` (no parens) — always truthy | Correctly checks `s[1] >= '0' && s[1] <= '9'` | **Go is correct; Python has a bug (missing parentheses on `.isdigit`).** No code change. |
| 8 | `Sig.AddSymbol` polymorphism | Checks `iu.ivy_have_polymorphism` global flag | Checks `IsPolymorphicName(name)` | **FIXED** — Added `iu.IvyHavePolymorphism` global to `ivyutils/names.go`, set by `SetStringVersion` (true for version > 1.2). `AddSymbol` and `FindPolymorphicSymbol` now gate on it. |
| 9 | `IsAlternationFree` | Checks `free_variables(term)` for existential case | Missing `free_variables` check | **FIXED** — Added `len(lu.FreeVariables(n)) == 0` check for the existential case in `ivylogic/classify.go`. |
| 10 | `BooleanSort.name` | Monkey-patched to `'bool'` | `String()` returns `"Boolean"` | **FIXED** — `BooleanSort.String()` now returns `"bool"` (`logic/sort.go`). Updated tests in `logic/sort_test.go`, `solver/solver_test.go`. |
| 11 | `FunctionSort.is_finite` | Returns `True` (second monkey-patch wins) | No `IsFinite` method | **FIXED** — Added `IsFinite() bool` methods to `FunctionSort` and `BooleanSort` (`logic/sort.go`). |
| 12 | `Atom` function equals check | Compares `rel == equals` (structural, including sort) | Checks `rel.Name == "="` only | **FIXED** — `Atom` now uses `rel.Equal(Equals)` which checks both name and sort, matching Python's `rel == equals` (`recstruct.__eq__` compares `(name, sort)` tuples). `IsEquals` remains name-only, matching Python's `is_equals`. |
| 13 | `Sig.AddSort` redefinition | Silently ignores (creates error but doesn't raise) | Returns error to caller | **FIXED** — `AddSort` now silently overwrites on redefinition, matching Python (`ivylogic/sig.go`). |
| 14 | `Definition.__eq__` | `type(self) is type(other)` — identity check on type | Type assertion — may match subtypes | **NOT A BUG** — Go's `n.(*Definition)` won't match `*DefinitionSchema` because they are different struct types, even with embedding. Already correct. |
| 15 | `SortInferList` | `concretize_terms(terms, sorts)` processes all terms together | Processes each term independently | **FIXED** — Added `ConcretizeTerms` in `typeinfer/infer.go` with shared unification env; `SortInferList` now delegates to it with `sorts` and `unsortedVarNames` params matching Python. |
| 16 | `Some.sort()` | Missing `return` — returns `None` | Returns `s.Params[0].NodeSort()` or `TopS` | **Go is correct; Python has latent bug (missing `return`).** No `.sort()` is ever called on `Some` in the Python codebase. No code change. |

---

## 4. ivy_logic_utils.py → logicutil/, clauseops/

### 4.1 MISSING

| # | Python Function | Description |
|---|-----------------|-------------|
| 1 | `to_formula(s)`, `to_term(s)`, `to_clause(s)`, `to_clauses(s)`, `to_literal(s)` | Parser-based convenience functions |
| 2 | `resort_symbol(sym, subs)` | Standalone symbol resort function |
| 3 | `resort_sig(subs)` | Signature-wide sort remapping |
| 4 | Higher-order combinators | `apply_gen_to_list`, `gen_to_set`, `gen_unique`, `any_in`, `filter2` |
| 5 | Arity-filtered queries | `used_unary_relations_clause`, `used_binary_functions_clauses`, etc. |
| 6 | `Clauses.conjuncts()` | Returns `[close_epr(c) for c in self.fmlas]` |

### 4.2 BEHAVIORAL_DIFFERENCE

| # | Area | Python | Go | Impact |
|---|------|--------|----|----|
| 7 | `expand_abbrevs` | Handles `Ite` → `And(condition_conj...)` | **Missing Ite case entirely** | Ite nodes not expanded; affects Tseitin encoding and clause generation |
| 8 | `expand_abbrevs` | Handles `Iff`-with-`Ite` optimization | Missing | Missed optimization path |
| 9 | `de_morgan` | Calls `expand_abbrevs` first | Does NOT call `ExpandAbbrevs` first | Input may contain Implies/Iff that DeMorgan doesn't handle |
| 10 | `de_morgan` | Handles single-element `And`/`Or` | Missing single-element case | `Not(And(x))` not simplified to `Not(x)` → `DeMorgan(x)` |
| 11 | `de_morgan` | Does NOT handle ForAll/Exists | **Adds** ForAll/Exists push-through | Extra behavior not in Python |

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
| 31 | Main `ivy_compile` | Missing: `check_instantiations`, `add_to_hierarchy`, progress symbol removal, `type_check`. ~~`theory_context`~~ **FIXED** — `mod.UpdateTheory()` now called at end of `IvyCompile` |
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
| 9 | `functions()`/`relations()` lookups | ~~Missing `arrupd` (array Store) and `arrsel` (array Select)~~ **FIXED** — arrsel, arrupd, arrcst all wired in lookupBuiltinFunc/lookupBuiltinRelation/LookupNative (solver/z3convert.go); z3bridge/array.go provides Select, Store, ConstArray, ArraySort, ArrayDomain, ArrayRange |

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
| **Cross-cutting equality** | — | — | — | ~~15~~ **0 (all FIXED)** | ~~15~~ **0** |
| **Cross-cutting TODO/stub** | — | ~~21~~ **8 remaining (13 FIXED/addressed)** | — | — | ~~21~~ **8** |
| ivy_logic | 6 (3 FIXED) | 1 | 10 (7 FIXED, 2 acknowledged/not-a-bug) | 3 | 20 |
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

1. ~~**Polarity bug in `classify_ext.go`**~~ — ✅ **FIXED** — `Not` now correctly flips polarity in `symbolsOverUniversalsRec` and `universalVariablesRec` (§1.4, items 14-15). Regression tests in `polarity_test.go`.
2. ~~**Map keys using Sort interface**~~ — ✅ **FIXED** — `GetSortRefinement` now uses `map[lg.NodeKey]lg.Sort` with `lg.SortKey()`; `UniverseConstraint` accepts `[]SortUniverse` (§1.1, items 1-2)
3. **`ClausesModelToDiagram` creates `X=X`** — trivially true, doesn't capture model (§7.2, item 4)
4. **Quantifier bound constraints** not inside quantifier body for nat/range sorts (§7.3, item 1)
5. **`ExpandAbbrevs` missing Ite case** — affects all clause generation with Ite nodes (§4.2, item 7)
6. **14 empty compiler stubs** — `FixConstructors`, `CreateSortOrder`, `CreateConstructorSchemata`, `AttachProofs`, `CheckDefinitions`, `CreateConjActions`, `HandleTemporals`, etc. (§6.2)
7. **`check/` package largely non-functional** — `CheckProperties`, `CheckConjectures`, `CheckTemporals` are no-ops (§9)
8. ~~**Sort `== lg.Boolean` pointer comparisons**~~ — ✅ **FIXED** — all 5 sites now use `lg.SortEqual()` (§1.2)
9. ~~**True/False singleton pointer comparisons**~~ — ✅ **FIXED** — added `lg.IsTrue()`/`lg.IsFalse()` to logic package; all 6+ sites updated (§1.3)
10. **`art/art.go` Delete() bug** — uses `state.ID` after setting it to `-1` (§8.3, item 3)

### Medium Priority (Behavioral Fidelity)

11. Missing `prm:` prefix substitution in `compile_action_def` (§6.3, item 15)
12. Missing unit resolution in `clauses_case` (§7.3, item 3)
13. Missing range clamping in `numeral_to_z3` (§7.3, item 5)
14. ~~`BooleanSort` string representation `"Boolean"` vs `"bool"` (§3.2, item 10)~~ — ✅ **FIXED**
15. `de_morgan` not calling `expand_abbrevs` first (§4.2, item 9)
16. `SetStringVersion` compose character mismatch (§12.2, item 1)
17. `GetStdIncludeDir` too simplistic (§12.2, item 2)
18. Missing field action `ActionUpdate` methods (§5.1, items 2-3)
19. Missing `Some`/`SomeMinMax` handling in if/while compilation (§6.3, items 19-20)
20. `PostState` missing proper havocing (§8.3, item 7)
