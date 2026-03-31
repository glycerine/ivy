#!/usr/bin/env python3
"""Emit sexp strings for all logic IR types for cross-language comparison.

Output format: id<TAB>sexp per line.
Used by Go's TestSexpCrossLanguage in logic/sexp_cross_test.go.
"""

import sys
##sys.path.insert(0, '/Users/jaten/pyivy/ivy')
## portable version, we hope:
from pathlib import Path

# Get the directory where THIS script is located
# .parent.parent might be needed depending on how deep your script is
base_path = Path(__file__).resolve().parent / "pyivy" / "ivy"

sys.path.insert(0, str(base_path))

from ivy import logic as lg
from ivy import ivy_logic as il
from ivy import logic_sexp

logic_sexp.install()


def emit(id_, sexp):
    print("%s\t%s" % (id_, sexp))


# --- Common building blocks ---
S = lg.UninterpretedSort("S")
X = lg.Var("X", S)
Y = lg.Var("Y", S)
A = lg.Var("A", S)
Z = lg.Var("Z", S)
eq = lg.Eq(X, Y)
eqAZ = lg.Eq(A, Z)

# Sort types
emit("uninterp_sort_basic", S.sexp())
emit("boolean_sort", lg.BooleanSort().sexp())
emit("func_sort_binary", lg.FunctionSort(S, S, lg.Boolean).sexp())
emit("func_sort_unary", lg.FunctionSort(S, lg.Boolean).sexp())
emit("enum_sort", lg.EnumeratedSort("Color", ["red", "green", "blue"]).sexp())
emit("enum_sort_single", lg.EnumeratedSort("X", ["a"]).sexp())
emit("enum_sort_empty", lg.EnumeratedSort("E", []).sexp())
emit("range_sort", lg.RangeSort("idx", 0, 10).sexp())
emit("top_sort_default", lg.TopSort("TopSort").sexp())
emit("top_sort_named", lg.TopSort("Alpha").sexp())

# Term types
emit("variable", X.sexp())
emit("symbol_constant", lg.Const("c", S).sexp())
f_un = lg.Const("f", lg.FunctionSort(S, lg.Boolean))
emit("symbol_unary_func", f_un.sexp())

leq = lg.Const("leq", lg.FunctionSort(S, S, lg.Boolean))
emit("apply_binary", lg.Apply(leq, X, Y).sexp())

f_null = lg.Const("f", lg.FunctionSort(lg.Boolean))
emit("apply_nullary", lg.Apply(f_null).sexp())

g_sym = lg.Const("g", lg.FunctionSort(lg.Boolean, lg.Boolean))
inner = lg.Apply(f_un, X)
emit("apply_nested", lg.Apply(g_sym, inner).sexp())

# Formula types
emit("eq", eq.sexp())
emit("not", lg.Not(eq).sexp())
emit("and_two", lg.And(eq, eq).sexp())
emit("and_empty", lg.And().sexp())
emit("or_two", lg.Or(eq, eq).sexp())
emit("or_empty", lg.Or().sexp())
emit("implies", lg.Implies(eq, eq).sexp())
emit("iff", lg.Iff(eq, eq).sexp())
emit("ite", lg.Ite(eq, X, Y).sexp())
emit("globally_nil_env", lg.Globally(None, eq).sexp())
emit("globally_with_env", lg.Globally("env1", eq).sexp())
emit("eventually_nil_env", lg.Eventually(None, eq).sexp())
emit("eventually_with_env", lg.Eventually("env1", eq).sexp())
emit("when_operator", lg.WhenOperator("when", X, eq).sexp())
emit("cond", lg.Cond(X, Y).sexp())

# Quantifier types
emit("forall_single", lg.ForAll(frozenset([X]), eq).sexp())
# Multi-var: pass Z, A — should appear sorted as A, Z
emit("forall_multi_sorted", lg.ForAll(frozenset([Z, A]), eqAZ).sexp())
emit("exists", lg.Exists(frozenset([X]), eq).sexp())
emit("lambda", lg.Lambda((X,), X).sexp())
emit("named_binder_nil_env", lg.NamedBinder("nb", (X,), None, eq).sexp())
emit("named_binder_with_env", lg.NamedBinder("nb", (X,), "e1", eq).sexp())

# Definition types
emit("definition", il.Definition(X, Y).sexp())
emit("definition_schema", il.DefinitionSchema(X, Y).sexp())
