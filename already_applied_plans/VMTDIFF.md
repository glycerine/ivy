# Analysis of Ken McMillan's `vmt-rf` Branch vs Current Ivy Checkout

**Created: 2026-03-22**

The diff was generated as `git diff vmt-rf > diff.vmt-rf`, so:
- **`-` lines** = content in `vmt-rf` (Ken McMillan's branch) — these are the NEW things on that branch
- **`+` lines** = content in the current checkout (our port target)

This analysis describes what Ken McMillan's `vmt-rf` branch adds or changes relative to the current checkout.

---

## 1. Major New Modules Added in vmt-rf

### 1.1 mypyvy Integration (`ivy_mypyvy.py` — ~966 lines, NEW)
- **What**: A full translator from Ivy specifications to the [mypyvy](https://github.com/wilcoxjay/mypyvy) format, an external tool for verifying invariants of transition systems.
- **What it does**:
  - `Translation` class handles sort/symbol/formula translation from Ivy's internal representation to mypyvy's AST.
  - Translates actions to two-state transition formulas with explicit `new()` wrappers for modified symbols.
  - Translates initializers to one-state formulas with existential quantifiers.
  - Identifies mutable vs. immutable symbols based on axiom usage.
  - Supports Skolem macro reduction (removes intermediate/Tseitin variables by inlining definitions).
  - SMT-based simplification via Z3 tactics (`ctx-simplify` or expensive `ctx-solver-simplify`).
  - Round-trips through Z3 (Ivy→Z3→Ivy) for simplification.
  - Handles second-order existentials (intermediate relations) with havoc actions.
  - `MypyvyProgram` class assembles the full mypyvy program and writes `.pyv` output.
- **New parameters**: `unfold_macros` (remove intermediary variables), `simplify` (expensive SMT simplification).
- **New tactic**: `mypyvy` registered in `ivy_check.py` — reduces temporal goals then calls `ivy_mypyvy.check_isolate`.
- **New method**: `convert_to_mypyvy` isolate method support via `opt_method` parameter.

### 1.2 mypyvy Syntax Module (`mypyvy_syntax.py` — ~1869 lines, NEW)
- **What**: Complete AST, type system, scope management, and utilities for the mypyvy language.
- **Includes**: Sort hierarchy, expression types (Bool, Int, UnaryExpr, BinaryExpr, NaryExpr, AppExpr, QuantifierExpr, IfThenElse, Let), declaration types (SortDecl, RelationDecl, FunctionDecl, ConstantDecl, DefinitionDecl, InitDecl, InvariantDecl, AxiomDecl), Scope/Binder management, sort inference, substitution (capture-avoiding), CNF conversion, quantifier relativization, faithful printer.

### 1.3 mypyvy Utilities (`mypyvy_utils.py` — ~100+ lines, NEW)
- **What**: Supporting utilities — OrderedSet, MyLogger, argument parsing (MypyvyArgs), error reporting.

### 1.4 DuoAI Integration (`ivy_duoai.py` — ~124 lines, NEW)
- **What**: Translates Ivy isolates into a format for [DuoAI](https://github.com/VeriGu/DuoAI), an AI-assisted invariant inference tool.
- **What it does**:
  - Converts enumerated sorts to uninterpreted sorts with explicit distinctness/totality axioms.
  - Inlines all call actions recursively.
  - Extracts local variables to action parameters.
  - Removes `BindOldsAction` wrappers.
  - Outputs a flat Ivy module via `ivy_printer.print_module`.
  - Renames dotted names (`.` → `_`) and colon-prefixed names for DuoAI compatibility.
- **New tactic**: `duoai` registered in `ivy_check.py`.

### 1.5 Distributed Proof Checking (`ivy_distpy.py` — ~558 lines, NEW)
- **What**: Uses the `dispy` library for distributed, parallel proof checking across multiple machines.
- **What it does**:
  - Distributes property checking at per-property granularity for non-temporal properties.
  - Distributes temporal property checking at per-isolate granularity.
  - Parallel assertion checking at isolate level.
  - Contains a near-complete copy of `check_isolate` adapted for distributed execution.
  - Has an incomplete `DistributedCheckingInstance` class stub.
- **New parameter**: `distributed` (misspelled as `distribuedt` in the code).

### 1.6 Ranking Inference (`ivy_ranking_infer.py` — ~228 lines, NEW)
- **What**: Automatic ranking function inference for liveness proofs, complementing the manual `ranking` tactic.
- **What it does**:
  - `instrument()`: Takes the temporal model and instruments all actions with a fairness predicate `.r` that tracks whether the ranking condition (`work_progress`) has occurred.
  - Adds an idle action to the model.
  - Modifies the goal to include the fairness predicate as a premise.
  - `infer()`: Extracts the `GF r -> G(p -> F q)` pattern from the property, creates definitions for `.p`, `.q`, `.r` predicates, and exports to VMT via `ivy_vmt.check_isolate` with tagged definitions.
- **New tactic**: `ranking_infer` registered alongside `ranking` in `ivy_ranking.py`.

---

## 2. Major VMT Export Enhancements (`ivy_vmt.py`)

The vmt-rf branch significantly enhances the VMT (Verification Modulo Theories) export:

### 2.1 Per-Action Boolean Guards
- Each public action gets its own boolean variable in the VMT output.
- Transition formula is structured as `(=> action_name trans_formula)` with mutual exclusion `(or (not a1) (not a2))`.
- `get_action_defs()` and `get_trans_str()` functions generate this structure.

### 2.2 Selective Array Encoding
- New `mod_set` tracking: only symbols actually modified by actions are encoded as arrays.
- `add_to_mod_set()` traverses all actions and initializers to build the modification set.
- `encode_as_array()` checks `mod_set` membership.
- `use_array_encoding` parameter allows disabling array encoding entirely.

### 2.3 Lambda/Cast Support in Array Encoding
- `encode_assign` handles parameterized assignments via lambdas: `il.Lambda([aidx], sval)` wrapped in `cast` operations.
- `is_constant_lambda`, `is_any_lambda`, `ite_into_lambda` helpers for lambda manipulation.
- `fresh_skolem()` for havoc actions converted to assignments.
- `make_ret_val()` for handling return values from call actions.

### 2.4 Havoc Action Array Encoding
- `HavocAction` is converted to an `AssignAction` with a fresh skolem.

### 2.5 Call Action Return Value Handling
- Call actions with return values are decomposed into: local variables + call + copy-back assignments.

### 2.6 Frame Clause Generation
- `frame_clauses()` computes explicit equality constraints for unmodified state variables.
- Immutable variables identified and frame clauses added to the transition.

### 2.7 Tagged Definitions for External Tools
- `tagged_dfns` parameter in `check_isolate` allows passing specially-tagged definitions (e.g., `react_p`, `react_q`, `react_r`) into the VMT output.

### 2.8 Explicit Sort Declarations
- VMT output includes `declare-sort`, `declare-datatypes` for enum sorts, and array theory functions (`ReadArr`, `WriteArr`, `ConstArr`).

### 2.9 Z3 Simplification Applied
- All VMT formulas pass through `z3.simplify()` before output.
- Uses `symbol_to_z3_full` (new in solver) for proper naming.

### 2.10 Property Herbrandization (Partially Commented)
- `herbrandize_property_vars` and `normalize_prop` functions exist for converting universally quantified properties, though partially commented out.

### 2.11 `action_to_tr` Takes Background Theory
- The function now accepts a `bgt` (background theory) parameter instead of computing it internally.

### 2.12 Assert Check Disabled (Bug?)
- `if False and checked(action)` in `add_err_flag` — the `False and` effectively disables assertion checking in VMT mode. This may be intentional (WIP) or a debugging leftover.

---

## 3. Struct/Constructor Features Added in vmt-rf

### 3.1 Constructor/Destructor Elimination (`ivy_actions.py`)
- New functions: `sort_destructors`, `fresh_constructor_args`, `is_first_order_struct`, `sort_constructor`, `elim_destructors` (~53 lines).
- `elim_destructors` replaces destructor applications with constructor arguments and equality constraints.
- `Action.update()` optionally calls `elim_destructors` when the `constructors` attribute is set.
- `use_constructors()` checks `'constructors' in ivy_module.module.attributes`.

### 3.2 `constructors` Attribute
- `ivy_compiler.py`: `"constructors"` added to `defined_attributes`.

### 3.3 `CallAction.inline()` Method (~39 lines)
- Inlines a call action by substituting actual parameters for formals, with capture avoidance.
- Used by `ivy_duoai.py` for action inlining.

### 3.4 Module Helper Functions (`ivy_module.py`)
- `is_destructor(sym)`, `is_struct_sort(sort)`, `sort_destructors(sort)`, `is_constructor(sym)`.
- Used by `flatten_structs` tactic and DuoAI.

### 3.5 `flatten_structs` Tactic (`ivy_tactics.py` — ~130 lines)
- Flattens struct types into scalar components throughout a temporal model.
- `field_types()` recursively enumerates struct fields.
- `flatten_expr()` translates expressions with destructors/constructors into flattened form.
- `flatten_action()` handles action flattening.
- `SaveSyms` helper for scoped symbol management.

### 3.6 `macro_expand` Tactic (`ivy_tactics.py` — ~60 lines)
- Expands all macro definitions inline throughout a temporal model.
- Traverses bindings, init, invariants, assumptions, and premises.

### 3.7 `NormalProgram.macros` Field
- `ivy_temporal.py`: NormalProgram gains a `macros` field for storing definitions.
- `normal_program_from_module` gains `with_definitions` parameter.
- `__str__` prints `function` lines for macros.

### 3.8 `to_module()` Function (`ivy_temporal.py`)
- Converts a proof goal back to an Ivy module, used by `ivy_ranking_infer.py`.

---

## 4. Solver Enhancements in vmt-rf (`ivy_solver.py`)

### 4.1 `symbol_to_z3_full()` — New Function
- Handles numerals and uses `solver_name()` for proper Z3 naming.
- Used by VMT export for proper `declare-fun` generation.

### 4.2 Lambda Support
- `mylambda(vs, z3_vs, z3_body)` wraps `z3.Lambda`.
- `formula_to_z3_int` handles `is_lambda` alongside quantifiers.
- `cast` term handling in `term_to_z3` — asserts sort compatibility.

### 4.3 Separate `z3_enums` Dictionary
- Enumerated sorts cached in their own dict rather than sharing `z3_sorts`.

### 4.4 Enhanced Error Reporting in `atom_to_z3`
- Wraps `apply_z3_func` in try/except to print atom details on failure.

### 4.5 SMT2 File Dump on Failure in `get_small_model`
- Writes `ivy.smt2` file and prints model on failure for debugging.

---

## 5. Language/Parser Changes in vmt-rf

### 5.1 `unprovable property` Grammar Rule
- `ivy_parser.py`: New grammar rule `p_top_unprovable_property_labeledfmla` allows `unprovable property` declarations.
- Sets `lf.unprovable = True` on the labeled formula.
- `ivy_printer.py`: Prints `unprovable` prefix.

### 5.2 `[this]` Label Support
- `ivy_logic_parser.py`: New rule `p_LABEL_LB_THIS_RB` allows `[this]` as a proof label.
- `ivy_compiler.py`: `attach_proofs` allows `lab == 'this'` for isolate proofs.

### 5.3 `is_unprovable` Checks in `ivy_check.py`
- `is_unprovable(lf)` helper with `hasattr` safety check.
- Properties filtered by `is_check_mod_unprovable` before checking.
- After checking, only non-unprovable properties become axioms.

### 5.4 `EnumeratedSort.__str__` Changed
- `logic.py`: Returns `self.name` instead of `'{' + ','.join(self.extension) + '}'`.
- Shows sort name rather than listing all enum values.

### 5.5 `is_def()` Added to `ivy_logic.py`
- Simple helper: `isinstance(expr, Definition)`.

---

## 6. Proof System Changes in vmt-rf

### 6.1 `definition_to_goal()` Added to `ivy_proof.py`
- Converts a definition to a goal by closing its formula universally.
- Used in isolate proof checking.

### 6.2 `var_subst_goal` — ConstantDecl Guard
- Returns `goal` unchanged if it's a `ConstantDecl`, preventing variable substitution on constant declarations.

### 6.3 Isolate Proof Checking — `use_context` Parameter
- `check_subgoals` gains `use_context=True` parameter.
- When `use_context=False` (used for isolate proofs), axioms and definitions are stripped from the module context.
- Comment explains: "when checking the isolate proof, we don't [use] the axioms and definitions in the module context by default. This allows the tactics to drop axioms and definitions."

### 6.4 Isolate Proof Goal Construction
- Uses `ivy_proof.make_goal` with `definition_to_goal` for definitions.
- Axioms are explicitly included as premises via `defns + axioms`.

---

## 7. Fragment Checker Changes in vmt-rf (`ivy_fragment.py`)

### 7.1 Definition-as-Axiom Heuristic Added
- `get_assumes_and_asserts` now scans `labeled_axioms` for equation-like formulas that look like definitions.
- If an axiom has the form `f(X1,...,Xn) = rhs` with distinct variable args, it's treated as a macro definition instead of an axiom.
- Tracks `defined` and `referenced` sets to avoid duplicates and detect recursion.
- **Why**: Allows the fragment checker to handle axioms that are really definitions more efficiently, treating them as macros for expansion.

---

## 8. Check Flow Changes in vmt-rf (`ivy_check.py`)

### 8.1 `opt_method` Parameter
- New parameter `method` allowing users to set the verification method globally (e.g., `convert_to_mypyvy`).
- `get_isolate_method` checks `opt_method.get()` before falling back to isolate attributes.

### 8.2 `mc_isolate` — Assumed Property Handling
- Before checking, moves assumed properties from `labeled_props` to `labeled_axioms`.
- Filters non-assumed properties to `labeled_props`.

### 8.3 `convert_postconds` Guard
- `state.update if state.update is not None else ([],None,None)` — defensive guard for None updates.

### 8.4 Print Statement
- `print ('starting ivy_check...')` at module load time.

### 8.5 `ivy_ui` Import Commented Out
- `from . import ivy_ui` commented out at the top (but still imported locally where needed).

### 8.6 Diagnostic GUI Commented Out
- `display_cex`, `check_properties`, `gui_art` — all Tkinter GUI code is commented out with `#`.

---

## 9. Other Changes in vmt-rf

### 9.1 `ivy_temporal.py` — `new_action_to_old` Enhancement
- Clones the statement and explicitly sets `formal_params` and `formal_returns`.
- Current branch just returns `act.stmt` directly.

### 9.2 `ivy_ranking.py` — `ranking_infer` Tactic Registered
- `ipr.register_tactic('ranking_infer', l2s_tactic)` added.
- `defs_needed` differs for `ranking_infer` (only requires `work_created`, `work_progress`) vs `ranking` (requires all four work_* predicates).

### 9.3 `ivy_tactics.py` Imports
- Uses `logic_util as lu` instead of `ivy_logic_utils as lu`.
- Imports `OrderedSet` from `ordered_set` package and `OrderedDict` from collections.
- These are needed by `flatten_structs`.

### 9.4 `ivy_printer.py` — `invariant` Keyword
- Uses `'invariant'` instead of `'conjecture'` for `labeled_conjs` in `print_module`.

### 9.5 Z3 Import: Bare `import z3`
- vmt-rf uses standard `import z3` throughout, not the `ivy.z3` fork.

### 9.6 `ordered-set` and `z3-solver` in `setup.py`
- Both are listed as pip dependencies (vmt-rf uses PyPI Z3, not the submodule fork).

### 9.7 Z3 Submodule: Standard Z3
- Points to `Z3Prover/z3.git` (the official Z3), not the fork.

### 9.8 `ivy_compiler.py` — `attach_proofs` `this` Label
- `elif lab in mod.isolates or lab == 'this'` — allows `this` as a proof label for isolates.

### 9.9 Test Files
- `ranking_infer1.ivy`: Test for the `ranking_infer` tactic — an unbounded queue with a liveness property.
- `unprovable1.ivy`: Has `unprovable property false` test case.
- `vcgen1.ivy`: Has semicolon after `tactic vcgen;` and includes `showgoals`.
- `frag7/8/10.ivy`: Conditional expressions on LHS of equality: `(expr if q else Z) = g(X)`.

---

## 10. Summary: What vmt-rf Adds (Ken McMillan's Work)

**Major new capabilities:**
1. **mypyvy export** — translate Ivy specs to mypyvy for external invariant checking (~2900 lines total)
2. **DuoAI integration** — export for AI-assisted invariant inference (~124 lines)
3. **Distributed checking** — parallel proof checking via dispy (~558 lines)
4. **Ranking inference** — automatic ranking function inference for liveness (~228 lines)
5. **Enhanced VMT export** — per-action booleans, selective array encoding, lambda/cast support, frame clauses, Z3 simplification, tagged definitions
6. **Struct flattening** — `flatten_structs` and `macro_expand` tactics with full constructor/destructor support
7. **`unprovable property`** — new language feature for properties expected to fail
8. **`[this]` proof labels** — allows proof labels referring to the current isolate
9. **Definition-as-axiom heuristic** in fragment checker

**Things vmt-rf does NOT have (present in current checkout only):**
- ACL system (`ivy_acl.py`) for ignoring/assuming properties
- Union-Find v2 (iterative with rank)
- Action priority ordering
- Profiling/statistics support
- `no_check_guarantees` option
- Re-enabled diagnostic GUI
- Post-mortem debugger hook
- Enhanced ranking counterexample diagnostics
- Custom Z3 fork (`ruijiefang/z3-ivy`)
- LSP server stub
- Compose tactic stub

**Impact assessment for goivy port:**
- The mypyvy/DuoAI/distpy features are research integrations that may not need porting
- The VMT enhancements (especially lambda/cast, frame clauses, per-action booleans) are significant if VMT export matters
- Struct flattening and constructor support could matter for specs using struct types
- The ranking_infer tactic is useful for automated liveness proofs
- The `unprovable property` feature and `[this]` labels are language-level additions
