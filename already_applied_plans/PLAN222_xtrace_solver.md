# PLAN221: Add Comprehensive xtracer Coverage to ivy_solver.py and Go solver/

**Created:** 2026-04-07 ~18:30 UTC

## Context

The Go side panics on a Z3 Sort mismatch (`Sort mismatch at argument #1 for function (declare-fun < (Int Int) Bool) supplied sort is lclock`) during `formulaToZ3` at `solver.go:431`. The `defer/recover` in `formulaToZ3` swallows the panic and converts it to an error, allowing Go to silently skip definitions that Python translates successfully. This causes downstream divergences. PLAN220's changes to the dispatch chain were implemented but the panic persists in the golden test log.

To debug this and future solver divergences, we need comprehensive xtracer coverage on both sides. Currently only ~10 of ~90 Python functions and ~9 of ~155 Go functions have traces.

## Trace Message Convention for Solver

New convention uses **Python file:line** as the canonical reference (Python is source of truth):

```python
# Python:
if __debug__: xtracer.trace("ivy_solver.py:LINE func_name() ENTER details")

# Go (MUST emit identical string):
xtracer.Trace("ivy_solver.py:LINE func_name() ENTER details")
```

Both sides emit the SAME string so golden test comparison works. Include key argument values where useful for debugging.

## A. Python Changes: ~/ivy/pyivy/ivy/ivy/ivy_solver.py

Add `if __debug__: xtracer.trace(...)` at the top of every function body. Below is the complete list grouped by category. Line numbers are from the current Python source.

### A1. Setup/Config (6 functions)

| Line | Function | Trace Message |
|------|----------|---------------|
| 36 | `set_seed(seed)` | `"ivy_solver.py:36 set_seed() ENTER seed=%s"` |
| 43 | `set_macro_finder(truth)` | `"ivy_solver.py:43 set_macro_finder() ENTER truth=%s"` |
| 54 | `set_use_native_enums(t)` | `"ivy_solver.py:54 set_use_native_enums() ENTER t=%s"` |
| 61 | `solver_name(symbol)` | `"ivy_solver.py:61 solver_name() ENTER name=%s"` |
| 84 | `my_minus(*args)` | `"ivy_solver.py:84 my_minus() ENTER nargs=%d"` |
| 89 | `my_eq(x,y)` | `"ivy_solver.py:89 my_eq() ENTER"` |

### A2. Sort/Type Handling (8 functions)

| Line | Function | Trace Message |
|------|----------|---------------|
| 100 | `parse_array_theory(name)` | `"ivy_solver.py:100 parse_array_theory() ENTER name=%s"` |
| 108 | `sort_name_to_z3(name)` | `"ivy_solver.py:108 sort_name_to_z3() ENTER name=%s"` |
| 112 | `sorts(name)` | `"ivy_solver.py:112 sorts() ENTER name=%s"` |
| 142 | `parse_int_params(name)` | `"ivy_solver.py:142 parse_int_params() ENTER name=%s"` |
| 150 | `is_solver_sort(name)` | `"ivy_solver.py:150 is_solver_sort() ENTER name=%s"` |
| 241 | `uninterpretedsort(us)` | `"ivy_solver.py:241 uninterpretedsort() ENTER name=%s"` |
| 251 | `functionsort(fs)` | `"ivy_solver.py:251 functionsort() ENTER"` |
| 257 | `enumeratedsort(es)` | `"ivy_solver.py:257 enumeratedsort() ENTER name=%s"` |

### A3. Symbol/Relation/Function Lookup (8 functions)

| Line | Function | Trace Message |
|------|----------|---------------|
| 160 | `relations(name)` | `"ivy_solver.py:160 relations() ENTER name=%s"` |
| 175 | `bfe_to_z3(sym)` | `"ivy_solver.py:175 bfe_to_z3() ENTER sym=%s"` |
| 214 | `functions(name)` | `"ivy_solver.py:214 functions() ENTER name=%s"` |
| 225 | `is_solver_op(name)` | `"ivy_solver.py:225 is_solver_op() ENTER name=%s"` |
| 229 | `clear()` | `"ivy_solver.py:229 clear() ENTER"` |
| 267 | `symbol_to_z3(s)` | `"ivy_solver.py:267 symbol_to_z3() ENTER name=%s sort=%s"` |
| 278 | `range_sort_bounds_to_z3(itp)` | `"ivy_solver.py:278 range_sort_bounds_to_z3() ENTER"` |
| 290 | `lookup_native(thing,table,kind)` | `"ivy_solver.py:290 lookup_native() ENTER name=%s kind=%s"` |

### A4. Compatibility/Check (5 functions)

| Line | Function | Trace Message |
|------|----------|---------------|
| 327 | `check_native_compat_sym(sym)` | `"ivy_solver.py:327 check_native_compat_sym() ENTER sym=%s"` |
| 350 | `check_compat()` | `"ivy_solver.py:350 check_compat() ENTER"` |
| 358 | `sort_card(sort)` | `"ivy_solver.py:358 sort_card() ENTER sort=%s"` |
| 372 | `native_symbol(sym)` | `"ivy_solver.py:372 native_symbol() ENTER sym=%s"` |
| 377 | `apply_z3_func(pred,tup)` | `"ivy_solver.py:377 apply_z3_func() ENTER nargs=%d"` |

### A5. Term/Atom/Literal Conversion (8 functions)

| Line | Function | Trace Message |
|------|----------|---------------|
| 389 | `numeral_to_z3(num)` | `"ivy_solver.py:389 numeral_to_z3() ENTER num=%s"` |
| 412 | `enumerated_to_numeral(term)` | `"ivy_solver.py:412 enumerated_to_numeral() ENTER"` |
| 415 | `term_to_z3(term)` | `"ivy_solver.py:415 term_to_z3() ENTER type=%s name=%s"` |
| 470 | `lt_pred(sort)` | `"ivy_solver.py:470 lt_pred() ENTER sort=%s"` |
| 481 | `get_polymacs(op)` | `"ivy_solver.py:481 get_polymacs() ENTER op=%s"` |
| 484 | `atom_to_z3(atom)` | `"ivy_solver.py:484 atom_to_z3() ENTER rep=%s nargs=%d"` |
| 503 | `literal_to_z3(lit)` | `"ivy_solver.py:503 literal_to_z3() ENTER polarity=%s"` |
| 510 | `quant_constraints(vs,z3_vs)` | `"ivy_solver.py:510 quant_constraints() ENTER nvars=%d"` |

### A6. Formula Translation (5 functions, 2 already traced)

| Line | Function | Trace Message |
|------|----------|---------------|
| 524 | `forall(vs,z3_vs,z3_body)` | `"ivy_solver.py:524 forall() ENTER nvars=%d"` |
| 530 | `exists(vs,z3_vs,z3_body)` | `"ivy_solver.py:530 exists() ENTER nvars=%d"` |
| 536 | `clause_to_z3(clause)` | `"ivy_solver.py:536 clause_to_z3() ENTER nlits=%d"` |
| 546 | `conj_to_z3(cl)` | `"ivy_solver.py:546 conj_to_z3() ENTER type=%s"` |
| 551 | `type_constraints(syms)` | `"ivy_solver.py:551 type_constraints() ENTER nsyms=%d"` |
| 581 | `clauses_to_z3(clauses)` | ALREADY TRACED |
| 597 | `formula_to_z3_int(fmla)` | `"ivy_solver.py:597 formula_to_z3_int() ENTER type=%s"` |
| 646 | `formula_to_z3_closed(fmla)` | `"ivy_solver.py:646 formula_to_z3_closed() ENTER type=%s"` |
| 659 | `formula_to_z3(fmla)` | ALREADY TRACED (HASH) |

### A7. Solver Operations (9 functions)

| Line | Function | Trace Message |
|------|----------|---------------|
| 680 | `unsat_core(...)` | `"ivy_solver.py:680 unsat_core() ENTER"` |
| 709 | `binary_interpolant(...)` | `"ivy_solver.py:709 binary_interpolant() ENTER"` |
| 729 | `cube_to_z3(cube)` | `"ivy_solver.py:729 cube_to_z3() ENTER nlits=%d"` |
| 735 | `get_id(x)` | `"ivy_solver.py:735 get_id() ENTER"` |
| 738 | `check_cube(s,cube,...)` | `"ivy_solver.py:738 check_cube() ENTER"` |
| 760 | `new_solver()` | `"ivy_solver.py:760 new_solver() ENTER"` |
| 763 | `solver_add(solver,fmla)` | `"ivy_solver.py:763 solver_add() ENTER"` |
| 766 | `is_sat(s)` | `"ivy_solver.py:766 is_sat() ENTER"` |
| 769 | `add_clauses(s,clauses)` | `"ivy_solver.py:769 add_clauses() ENTER"` |

### A8. Model Operations (9 functions)

| Line | Function | Trace Message |
|------|----------|---------------|
| 774 | `get_model(s)` | `"ivy_solver.py:774 get_model() ENTER"` |
| 777 | `terms_match(tl1,tl2)` | `"ivy_solver.py:777 terms_match() ENTER"` |
| 792 | `get_arg_range(m,x)` | `"ivy_solver.py:792 get_arg_range() ENTER"` |
| 812 | `collect_numerals(z3term)` | `"ivy_solver.py:812 collect_numerals() ENTER"` |
| 821 | `from_z3_numeral(z3term,sort)` | `"ivy_solver.py:821 from_z3_numeral() ENTER sort=%s"` |
| 827 | `collect_model_values(sort,model,sym)` | `"ivy_solver.py:827 collect_model_values() ENTER sort=%s"` |
| 833 | `mine_interpreted_constants(model,vocab)` | `"ivy_solver.py:833 mine_interpreted_constants() ENTER"` |
| 844 | `enumerated_range(sort)` | `"ivy_solver.py:844 enumerated_range() ENTER"` |

### A9. HerbrandModel Methods

| Line | Method | Trace Message |
|------|--------|---------------|
| 848 | `__init__` | `"ivy_solver.py:848 HerbrandModel.__init__() ENTER"` |
| 864 | `sorted_sort_universe()` | ALREADY TRACED |

### A10. Advanced Solver Operations (10 functions)

| Line | Function | Trace Message |
|------|----------|---------------|
| 932 | `sort_from_z3(s)` | `"ivy_solver.py:932 sort_from_z3() ENTER"` |
| 935 | `constant_from_z3(sort,c)` | `"ivy_solver.py:935 constant_from_z3() ENTER sort=%s"` |
| 942 | `get_model_constant(m,t)` | `"ivy_solver.py:942 get_model_constant() ENTER"` |
| 959 | `clauses_imply(clauses1,clauses2)` | `"ivy_solver.py:959 clauses_imply() ENTER"` |
| 971 | `clauses_imply_list(...)` | `"ivy_solver.py:971 clauses_imply_list() ENTER"` |
| 1004 | `check_sequence(...)` | `"ivy_solver.py:1004 check_sequence() ENTER n=%d"` |
| 1035 | `not_clauses_to_z3(clauses)` | `"ivy_solver.py:1035 not_clauses_to_z3() ENTER"` |
| 1045 | `clauses_sat(clauses1)` | `"ivy_solver.py:1045 clauses_sat() ENTER"` |
| 1053 | `remove_duplicates_clauses(clauses)` | `"ivy_solver.py:1053 remove_duplicates_clauses() ENTER"` |
| 1058 | `clauses_case(clauses1)` | `"ivy_solver.py:1058 clauses_case() ENTER"` |

### A11. Model Simplification/Extraction (7 functions)

| Line | Function | Trace Message |
|------|----------|---------------|
| 1089 | `clause_model_simp(m,c)` | `"ivy_solver.py:1089 clause_model_simp() ENTER"` |
| 1105 | `get_model_clauses(clauses1)` | `"ivy_solver.py:1105 get_model_clauses() ENTER"` |
| 1115 | `sort_size_constraint(sort,size)` | `"ivy_solver.py:1115 sort_size_constraint() ENTER sort=%s size=%d"` |
| 1125 | `relation_size_constraint(relation,size)` | `"ivy_solver.py:1125 relation_size_constraint() ENTER size=%d"` |
| 1151 | `size_constraint(x,size)` | `"ivy_solver.py:1151 size_constraint() ENTER size=%d"` |
| 1162 | `model_if_none(clauses1,implied,model)` | `"ivy_solver.py:1162 model_if_none() ENTER"` |
| 1225 | `decide(s,atoms)` | `"ivy_solver.py:1225 decide() ENTER"` |

### A12. Complex Model Operations (8 functions)

| Line | Function | Trace Message |
|------|----------|---------------|
| 1237 | `get_small_model(...)` | `"ivy_solver.py:1237 get_small_model() ENTER shrink=%s"` |
| 1353 | `model_universe_facts(h,sort,upclose)` | `"ivy_solver.py:1353 model_universe_facts() ENTER sort=%s"` |
| 1368 | `model_facts(h,ignore,clauses1,...)` | `"ivy_solver.py:1368 model_facts() ENTER"` |
| 1399 | `numeral_assign(clauses,h)` | `"ivy_solver.py:1399 numeral_assign() ENTER"` |
| 1434 | `clauses_model_to_clauses(...)` | `"ivy_solver.py:1434 clauses_model_to_clauses() ENTER"` |
| 1458 | `bound_quantifiers_clauses(...)` | `"ivy_solver.py:1458 bound_quantifiers_clauses() ENTER"` |
| 1480 | `filter_redundant_facts(...)` | `"ivy_solver.py:1480 filter_redundant_facts() ENTER"` |
| 1509 | `clauses_model_to_diagram(...)` | `"ivy_solver.py:1509 clauses_model_to_diagram() ENTER"` |

### A13. More Model Ops + Encoding (14 functions)

| Line | Function | Trace Message |
|------|----------|---------------|
| 1579 | `relation_model_to_clauses(h,r,n)` | `"ivy_solver.py:1579 relation_model_to_clauses() ENTER"` |
| 1589 | `get_lit_facts(h,lit,res)` | `"ivy_solver.py:1589 get_lit_facts() ENTER"` |
| 1598 | `function_model_to_clauses(h,f)` | `"ivy_solver.py:1598 function_model_to_clauses() ENTER"` |
| 1615 | `clauses_imply_formula(clauses1,fmla2)` | `"ivy_solver.py:1615 clauses_imply_formula() ENTER"` |
| 1624 | `ceillog2(n)` | `"ivy_solver.py:1624 ceillog2() ENTER n=%d"` |
| 1631 | `gebin(bits,n)` | `"ivy_solver.py:1631 gebin() ENTER n=%d"` |
| 1641 | `binenc(m,n)` | `"ivy_solver.py:1641 binenc() ENTER m=%d n=%d"` |
| 1645 | `z3_function(name,sig)` | `"ivy_solver.py:1645 z3_function() ENTER name=%s"` |
| 1648 | `encode_term(t,n,sort)` | `"ivy_solver.py:1648 encode_term() ENTER sort=%s"` |
| 1678 | `encode_equality(*terms)` | `"ivy_solver.py:1678 encode_equality() ENTER nterms=%d"` |
| 1692 | `substitute(t,*m)` | `"ivy_solver.py:1692 substitute() ENTER"` |
| 1702 | `z3sort_to_sort(z3sort)` | `"ivy_solver.py:1702 z3sort_to_sort() ENTER"` |
| 1707 | `z3decl_to_symbol(z3decl)` | `"ivy_solver.py:1707 z3decl_to_symbol() ENTER"` |
| 1718 | `z3_to_formula(z3expr,vars)` | `"ivy_solver.py:1718 z3_to_formula() ENTER"` |

**Total Python additions: ~82 new trace calls** (plus 10 existing = 92 total)

## B. Go Changes: ~/go/src/github.com/glycerine/ivy/goivy/solver/

For each Go function that corresponds to a Python function, add `xtracer.Trace(...)` with the **identical message string** (same `ivy_solver.py:LINE` reference). The Go line number does NOT appear in the trace message - only the Python line number, ensuring both sides emit matching strings.

### B1. solver/encoding.go

| Go Function | Python Equivalent | Trace Message (same as Python) |
|-------------|-------------------|-------------------------------|
| `SetSeed()` | `set_seed` | `"ivy_solver.py:36 set_seed() ENTER seed=%s"` |
| `SetMacroFinder()` | `set_macro_finder` | `"ivy_solver.py:43 set_macro_finder() ENTER truth=%s"` |
| `SetUseNativeEnums()` | `set_use_native_enums` | `"ivy_solver.py:54 set_use_native_enums() ENTER t=%s"` |
| `ParseArrayTheory()` | `parse_array_theory` | `"ivy_solver.py:100 parse_array_theory() ENTER name=%s"` |
| `ParseIntParams()` | `parse_int_params` | `"ivy_solver.py:142 parse_int_params() ENTER name=%s"` |
| `IsSolverSort()` | `is_solver_sort` | `"ivy_solver.py:150 is_solver_sort() ENTER name=%s"` |
| `IsSolverOp()` | `is_solver_op` | `"ivy_solver.py:225 is_solver_op() ENTER name=%s"` |
| `NativeSymbol()` | `native_symbol` | `"ivy_solver.py:372 native_symbol() ENTER sym=%s"` |
| `LtPred()` | `lt_pred` | `"ivy_solver.py:470 lt_pred() ENTER sort=%s"` |
| `NumeralToZ3()` | `numeral_to_z3` | `"ivy_solver.py:389 numeral_to_z3() ENTER num=%s"` |
| `EnumeratedToNumeral()` | `enumerated_to_numeral` | `"ivy_solver.py:412 enumerated_to_numeral() ENTER"` |
| `CollectNumerals()` | `collect_numerals` | `"ivy_solver.py:812 collect_numerals() ENTER"` |
| `RangeSortBounds()` | `range_sort_bounds_to_z3` | `"ivy_solver.py:278 range_sort_bounds_to_z3() ENTER"` |
| `CeilLog2()` | `ceillog2` | `"ivy_solver.py:1624 ceillog2() ENTER n=%d"` |
| `BinEnc()` | `binenc` | `"ivy_solver.py:1641 binenc() ENTER m=%d n=%d"` |
| `Z3Function()` | `z3_function` | `"ivy_solver.py:1645 z3_function() ENTER name=%s"` |
| `EncodeTermZ3()` | `encode_term` | `"ivy_solver.py:1648 encode_term() ENTER sort=%s"` |
| `EncodeEqualityZ3()` | `encode_equality` | `"ivy_solver.py:1678 encode_equality() ENTER nterms=%d"` |

### B2. solver/z3convert.go

| Go Function | Python Equivalent | Trace Message |
|-------------|-------------------|---------------|
| `SolverName()` | `solver_name` | `"ivy_solver.py:61 solver_name() ENTER name=%s"` |
| `MyMinus()` | `my_minus` | `"ivy_solver.py:84 my_minus() ENTER nargs=%d"` |
| `MyEq()` | `my_eq` | `"ivy_solver.py:89 my_eq() ENTER"` |
| `SortNameToZ3()` | `sort_name_to_z3` | `"ivy_solver.py:108 sort_name_to_z3() ENTER name=%s"` |
| `LookupNative()` | `lookup_native` | `"ivy_solver.py:290 lookup_native() ENTER name=%s kind=%s"` |
| `bfeToZ3()` | `bfe_to_z3` | `"ivy_solver.py:175 bfe_to_z3() ENTER sym=%s"` |
| `RangeSortBoundsToZ3()` | `range_sort_bounds_to_z3` | `"ivy_solver.py:278 range_sort_bounds_to_z3() ENTER"` |
| `FromZ3Numeral()` | `from_z3_numeral` | `"ivy_solver.py:821 from_z3_numeral() ENTER sort=%s"` |
| `SubstituteZ3()` | `substitute` | `"ivy_solver.py:1692 substitute() ENTER"` |
| `Z3SortToSort()` | `z3sort_to_sort` | `"ivy_solver.py:1702 z3sort_to_sort() ENTER"` |
| `Z3DeclToSymbol()` | `z3decl_to_symbol` | `"ivy_solver.py:1707 z3decl_to_symbol() ENTER"` |
| `Z3ToFormula()` | `z3_to_formula` | `"ivy_solver.py:1718 z3_to_formula() ENTER"` |
| `BinaryInterpolant()` | `binary_interpolant` | `"ivy_solver.py:709 binary_interpolant() ENTER"` |
| `Gebin()` | `gebin` | `"ivy_solver.py:1631 gebin() ENTER n=%d"` |
| `SortOrder.Compare()` | `SortOrder.__lt__` | `"ivy_solver.py:800 SortOrder.compare() ENTER"` |

### B3. solver/solver.go

| Go Function | Python Equivalent | Trace Message |
|-------------|-------------------|---------------|
| `Clear()` | `clear` | `"ivy_solver.py:229 clear() ENTER"` |
| `ClausesToZ3()` | `clauses_to_z3` | ALREADY TRACED |
| `formulaToZ3()` | `formula_to_z3` | ALREADY TRACED (HASH) |
| `formulaToZ3Closed()` | `formula_to_z3_closed` | `"ivy_solver.py:646 formula_to_z3_closed() ENTER type=%s"` |
| `conjToZ3()` | `conj_to_z3` | `"ivy_solver.py:546 conj_to_z3() ENTER type=%s"` |
| `forall()` | `forall` | `"ivy_solver.py:524 forall() ENTER nvars=%d"` |
| `NotClausesToZ3()` | `not_clauses_to_z3` | `"ivy_solver.py:1035 not_clauses_to_z3() ENTER"` |
| `IsSat()` | `is_sat` | `"ivy_solver.py:766 is_sat() ENTER"` |
| `ClausesSat()` | `clauses_sat` | `"ivy_solver.py:1045 clauses_sat() ENTER"` |
| `ClausesImply()` | `clauses_imply` | `"ivy_solver.py:959 clauses_imply() ENTER"` |
| `ClausesImplyFormula()` | `clauses_imply_formula` | `"ivy_solver.py:1615 clauses_imply_formula() ENTER"` |
| `UnsatCore()` | `unsat_core` | `"ivy_solver.py:680 unsat_core() ENTER"` |
| `SortSizeConstraint()` | `sort_size_constraint` | `"ivy_solver.py:1115 sort_size_constraint() ENTER sort=%s size=%d"` |
| `RelationSizeConstraint()` | `relation_size_constraint` | `"ivy_solver.py:1125 relation_size_constraint() ENTER size=%d"` |
| `SizeConstraint()` | `size_constraint` | `"ivy_solver.py:1151 size_constraint() ENTER size=%d"` |
| `CheckSequence()` | `check_sequence` | `"ivy_solver.py:1004 check_sequence() ENTER n=%d"` |

### B4. solver/compat.go

| Go Function | Python Equivalent | Trace Message |
|-------------|-------------------|---------------|
| `CheckNativeCompatSym()` | `check_native_compat_sym` | `"ivy_solver.py:327 check_native_compat_sym() ENTER sym=%s"` |
| `CheckCompat()` | `check_compat` | `"ivy_solver.py:350 check_compat() ENTER"` |
| `TermsMatch()` | `terms_match` | `"ivy_solver.py:777 terms_match() ENTER"` |
| `GetArgRange()` | `get_arg_range` | `"ivy_solver.py:792 get_arg_range() ENTER"` |
| `ModelIfNone()` | `model_if_none` | `"ivy_solver.py:1162 model_if_none() ENTER"` |
| `CollectModelValues()` | `collect_model_values` | `"ivy_solver.py:827 collect_model_values() ENTER sort=%s"` |
| `MineInterpretedConstants()` | `mine_interpreted_constants` | `"ivy_solver.py:833 mine_interpreted_constants() ENTER"` |
| `NumeralAssign()` | `numeral_assign` | `"ivy_solver.py:1399 numeral_assign() ENTER"` |
| `GetPolymacs()` | `get_polymacs` | `"ivy_solver.py:481 get_polymacs() ENTER op=%s"` |
| `QuantConstraints()` | `quant_constraints` | `"ivy_solver.py:510 quant_constraints() ENTER nvars=%d"` |
| `TypeConstraints()` | `type_constraints` | `"ivy_solver.py:551 type_constraints() ENTER nsyms=%d"` |

### B5. solver/herbrand.go

| Go Function | Python Equivalent | Trace Message |
|-------------|-------------------|---------------|
| `NewHerbrandModel()` | `HerbrandModel.__init__` | `"ivy_solver.py:848 HerbrandModel.__init__() ENTER"` |
| `SortedSortUniverse()` | `sorted_sort_universe` | ALREADY TRACED |
| `SortCard()` | `sort_card` | `"ivy_solver.py:358 sort_card() ENTER sort=%s"` |
| `EnumeratedRange()` | `enumerated_range` | `"ivy_solver.py:844 enumerated_range() ENTER"` |
| `ModelUniverseFacts()` | `model_universe_facts` | `"ivy_solver.py:1353 model_universe_facts() ENTER sort=%s"` |
| `ModelFacts()` | `model_facts` | `"ivy_solver.py:1368 model_facts() ENTER"` |
| `RelationModelToClauses()` | `relation_model_to_clauses` | `"ivy_solver.py:1579 relation_model_to_clauses() ENTER"` |
| `FunctionModelToClauses()` | `function_model_to_clauses` | `"ivy_solver.py:1598 function_model_to_clauses() ENTER"` |
| `getLitFacts()` | `get_lit_facts` | `"ivy_solver.py:1589 get_lit_facts() ENTER"` |
| `ClausesCase()` | `clauses_case` | `"ivy_solver.py:1058 clauses_case() ENTER"` |
| `clauseModelSimp()` | `clause_model_simp` | `"ivy_solver.py:1089 clause_model_simp() ENTER"` |
| `constantFromZ3()` | `constant_from_z3` | `"ivy_solver.py:935 constant_from_z3() ENTER sort=%s"` |
| `getModelConstant()` | `get_model_constant` | `"ivy_solver.py:942 get_model_constant() ENTER"` |

### B6. solver/model.go

| Go Function | Python Equivalent | Trace Message |
|-------------|-------------------|---------------|
| `GetModelClauses()` | `get_model_clauses` | `"ivy_solver.py:1105 get_model_clauses() ENTER"` |
| `GetSmallModel()` | `get_small_model` | `"ivy_solver.py:1237 get_small_model() ENTER shrink=%s"` |
| `CubeToZ3()` | `cube_to_z3` | `"ivy_solver.py:729 cube_to_z3() ENTER nlits=%d"` |
| `LiteralToZ3()` | `literal_to_z3` | `"ivy_solver.py:503 literal_to_z3() ENTER polarity=%s"` |
| `CheckCube()` | `check_cube` | `"ivy_solver.py:738 check_cube() ENTER"` |
| `ClausesModelToClauses()` | `clauses_model_to_clauses` | `"ivy_solver.py:1434 clauses_model_to_clauses() ENTER"` |
| `FilterRedundantFacts()` | `filter_redundant_facts` | `"ivy_solver.py:1480 filter_redundant_facts() ENTER"` |

### B7. solver/clauses.go

| Go Function | Python Equivalent | Trace Message |
|-------------|-------------------|---------------|
| `ClausesImplyList()` | `clauses_imply_list` | `"ivy_solver.py:971 clauses_imply_list() ENTER"` |
| `BoundQuantifiersClauses()` | `bound_quantifiers_clauses` | `"ivy_solver.py:1458 bound_quantifiers_clauses() ENTER"` |
| `RemoveDuplicatesClauses()` | `remove_duplicates_clauses` | `"ivy_solver.py:1053 remove_duplicates_clauses() ENTER"` |
| `ClausesModelToDiagram()` | `clauses_model_to_diagram` | `"ivy_solver.py:1509 clauses_model_to_diagram() ENTER"` |
| `NewZ3Solver()` | `new_solver` | `"ivy_solver.py:760 new_solver() ENTER"` |
| `AddClauses()` | `add_clauses` | `"ivy_solver.py:769 add_clauses() ENTER"` |
| `SolverAdd()` | `solver_add` | `"ivy_solver.py:763 solver_add() ENTER"` |
| `Decide()` | `decide` | `"ivy_solver.py:1225 decide() ENTER"` |

**Total Go additions: ~82 new trace calls** (matching the Python additions)

## C. Python Functions with No Direct Go Counterpart in solver/

These Python functions are handled differently in Go (via z3bridge callbacks or inline in translate.go):

| Python Function | Go Handling | Notes |
|----------------|-------------|-------|
| `sorts(name)` py:112 | `z3bridge.TranslateSort()` via `SortLookup` callback | Dict lookup in Py → callback in Go |
| `relations(name)` py:160 | `lookupBuiltinRelation()` in z3convert.go | Dict lookup → switch statement |
| `functions(name)` py:214 | `translateBuiltinOp()` in z3bridge/translate.go | Dict lookup → switch statement |
| `uninterpretedsort(us)` py:241 | `TranslateSort()` case in z3bridge/translate.go | Inline in sort translation |
| `functionsort(fs)` py:251 | `makeFuncDecl()` in z3bridge/translate.go | Inline in decl creation |
| `enumeratedsort(es)` py:257 | `TranslateSort()` case in z3bridge/translate.go | Inline in sort translation |
| `symbol_to_z3(s)` py:267 | `translateVarOrConst()` in z3bridge/translate.go | Inline in translate |
| `apply_z3_func(pred,tup)` py:377 | `FuncDecl.Apply()` in z3bridge/quantifier.go | Direct Z3 API call |
| `term_to_z3(term)` py:415 | `Translate()` in z3bridge/translate.go | Main dispatch switch |
| `atom_to_z3(atom)` py:484 | `Translate()` Apply case + NativeLookup | Split across translate.go |
| `exists(vs,z3_vs,z3_body)` py:530 | `translateQuantifier(false,...)` in translate.go | Part of quantifier handler |
| `clause_to_z3(clause)` py:536 | Absorbed into `ClausesToZ3()` | No separate Go function |
| `get_model(s)` py:774 | `z3bridge.Solver.Model()` | Direct Z3 API |
| `get_id(x)` py:735 | `z3bridge.Expr.GetId()` | Direct Z3 API |
| `formula_to_z3_int(fmla)` py:597 | `z3bridge.Translate()` | Main dispatch |

For these, traces need to go in the z3bridge/translate.go functions that handle the equivalent logic. Add traces there too, using the same `ivy_solver.py:LINE` convention.

### z3bridge/translate.go Traces to Add

| Go Location | Python Equivalent | Trace Message |
|-------------|-------------------|---------------|
| `TranslateSort()` entry | `sorts()/uninterpretedsort()/enumeratedsort()` | `"ivy_solver.py:112 sorts() ENTER name=%s"` |
| `translateVarOrConst()` entry | `symbol_to_z3()` | `"ivy_solver.py:267 symbol_to_z3() ENTER name=%s sort=%s"` |
| `Translate()` Apply case | `term_to_z3()/atom_to_z3()` | `"ivy_solver.py:415 term_to_z3() ENTER type=%s name=%s"` |
| `translateQuantifier()` for Exists | `exists()` | `"ivy_solver.py:530 exists() ENTER nvars=%d"` |
| `Translate()` ForAll case | `forall()` wrapping | Already in solver.go |
| `Translate()` entry | `formula_to_z3_int()` | `"ivy_solver.py:597 formula_to_z3_int() ENTER type=%s"` |

## D. Bugs and Divergences to Investigate

### D1. PLAN220 panic still occurring (HIGH PRIORITY)

The Sort mismatch panic on `<` with `lclock` sort persists despite PLAN220 changes. Possible causes:
- The golden test binary wasn't rebuilt after PLAN220 edits
- The `translateBuiltinOp` comparison cases weren't fully removed
- The NativeLookup callback isn't wired for the specific code path hitting the panic

**Action:** After adding traces, rebuild and re-run golden test. The new traces on `lookup_native()`, `atom_to_z3()`, `term_to_z3()` will show the exact dispatch path when `<` is used with `lclock`, revealing whether the PLAN220 fixes are taking effect.

### D2. Go `formulaToZ3` swallows panics (MEDIUM)

`solver.go:428-434`: The `defer/recover` silently converts Z3 panics to errors. Python would crash. Go skipping definitions that Python translates successfully causes downstream divergences.

**Action:** After fixing D1 so no panic occurs, consider whether the recover should be removed or made more explicit (e.g., re-panic in debug mode).

### D3. `formula_to_z3_int` early boolean optimization (LOW)

Python checks `Definition/Eq/Iff` for boolean RHS twice - once before recursing into args, once after. Go only does it after translation via `EqFunc` callback. This is unlikely to cause issues but is a structural divergence worth noting.

## E. Execution Order

1. **Python first**: Add all ~82 xtracer.trace() calls to ivy_solver.py
2. **Go solver/ next**: Add matching ~82 xtracer.Trace() calls across the 7 .go files
3. **Go z3bridge/translate.go**: Add ~6 traces for functions handled in z3bridge
4. **Rebuild and test**: `go build ./... && go test ./solver/... && go test ./z3bridge/...`
5. **Run golden test**: `cd ~/ivy/goivy && make golden`
6. **Analyze new log.red**: The comprehensive traces will show exactly where Go and Python solver paths diverge

## F. Verification

```bash
# Build
cd ~/go/src/github.com/glycerine/ivy/goivy && go build ./...

# Unit tests
go test ./solver/... && go test ./z3bridge/... && go test ./tactics/... && go test ./actions/...

# Golden test (will produce new log.red with comprehensive solver traces)
cd ~/ivy/goivy && make golden
```

After golden test, examine log.red for:
1. Whether the `<`/`lclock` panic is gone
2. Whether Python and Go solver traces match 1:1
3. Where new divergences appear (if any)

## G. Files Modified

- `~/ivy/pyivy/ivy/ivy/ivy_solver.py` — add ~82 xtracer.trace() calls
- `~/go/src/github.com/glycerine/ivy/goivy/solver/encoding.go` — add ~18 xtracer.Trace() calls
- `~/go/src/github.com/glycerine/ivy/goivy/solver/z3convert.go` — add ~15 xtracer.Trace() calls
- `~/go/src/github.com/glycerine/ivy/goivy/solver/solver.go` — add ~14 xtracer.Trace() calls
- `~/go/src/github.com/glycerine/ivy/goivy/solver/compat.go` — add ~11 xtracer.Trace() calls
- `~/go/src/github.com/glycerine/ivy/goivy/solver/herbrand.go` — add ~13 xtracer.Trace() calls
- `~/go/src/github.com/glycerine/ivy/goivy/solver/model.go` — add ~7 xtracer.Trace() calls
- `~/go/src/github.com/glycerine/ivy/goivy/solver/clauses.go` — add ~8 xtracer.Trace() calls
- `~/go/src/github.com/glycerine/ivy/goivy/z3bridge/translate.go` — add ~6 xtracer.Trace() calls
