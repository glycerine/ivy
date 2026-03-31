#!/usr/bin/env python3
"""Test that Python logic IR sexp() matches the shared golden vectors.

Also verifies canon() == sexp() for all types.
"""

import os
import re
import sys

##sys.path.insert(0, '/Users/jaten/pyivy/ivy')
from pathlib import Path

# Get the directory where THIS script is located
# .parent.parent might be needed depending on how deep your script is
base_path = Path(__file__).resolve() / "pyivy" / "ivy"

sys.path.insert(0, str(base_path))

from ivy import logic as lg
from ivy import ivy_logic as il
from ivy import ivy_logic_utils as lut
from ivy import logic_sexp

logic_sexp.install()


def parse_vectors(path):
    """Parse sexp_vectors.sexp file, returning list of (id, type, expected)."""
    with open(path) as f:
        text = f.read()

    vectors = []
    # Match (vector id:ID type:TYPE\n  expected:"EXPECTED")
    # The expected value is always double-quoted
    pattern = re.compile(
        r'\(vector\s+id:(\S+)\s+type:(\S+)\s+expected:"((?:[^"\\]|\\.)*)"\)',
        re.DOTALL
    )
    for m in pattern.finditer(text):
        vectors.append((m.group(1), m.group(2), m.group(3)))
    return vectors


# --- Common building blocks ---
S = lg.UninterpretedSort("S")
X = lg.Var("X", S)
Y = lg.Var("Y", S)
A = lg.Var("A", S)
Z = lg.Var("Z", S)
eq = lg.Eq(X, Y)
eqAZ = lg.Eq(A, Z)


def build_node(vec_id):
    """Construct the Python node for a given vector ID."""
    builders = {
        # Sorts
        "uninterp_sort_basic": lambda: S,
        "boolean_sort": lambda: lg.BooleanSort(),
        "func_sort_binary": lambda: lg.FunctionSort(S, S, lg.Boolean),
        "func_sort_unary": lambda: lg.FunctionSort(S, lg.Boolean),
        "enum_sort": lambda: lg.EnumeratedSort("Color", ["red", "green", "blue"]),
        "enum_sort_single": lambda: lg.EnumeratedSort("X", ["a"]),
        "enum_sort_empty": lambda: lg.EnumeratedSort("E", []),
        "range_sort": lambda: lg.RangeSort("idx", 0, 10),
        "top_sort_default": lambda: lg.TopSort("TopSort"),
        "top_sort_named": lambda: lg.TopSort("Alpha"),

        # Terms
        "variable": lambda: X,
        "symbol_constant": lambda: lg.Const("c", S),
        "symbol_unary_func": lambda: lg.Const("f", lg.FunctionSort(S, lg.Boolean)),
        "apply_binary": lambda: lg.Apply(
            lg.Const("leq", lg.FunctionSort(S, S, lg.Boolean)), X, Y),
        "apply_nullary": lambda: lg.Apply(
            lg.Const("f", lg.FunctionSort(lg.Boolean))),
        "apply_nested": lambda: lg.Apply(
            lg.Const("g", lg.FunctionSort(lg.Boolean, lg.Boolean)),
            lg.Apply(lg.Const("f", lg.FunctionSort(S, lg.Boolean)), X)),

        # Formulas
        "eq": lambda: eq,
        "not": lambda: lg.Not(eq),
        "and_two": lambda: lg.And(eq, eq),
        "and_empty": lambda: lg.And(),
        "or_two": lambda: lg.Or(eq, eq),
        "or_empty": lambda: lg.Or(),
        "implies": lambda: lg.Implies(eq, eq),
        "iff": lambda: lg.Iff(eq, eq),
        "ite": lambda: lg.Ite(eq, X, Y),
        "globally_nil_env": lambda: lg.Globally(None, eq),
        "globally_with_env": lambda: lg.Globally("env1", eq),
        "eventually_nil_env": lambda: lg.Eventually(None, eq),
        "eventually_with_env": lambda: lg.Eventually("env1", eq),
        "when_operator": lambda: lg.WhenOperator("when", X, eq),
        "cond": lambda: lg.Cond(X, Y),

        # Quantifiers
        "forall_single": lambda: lg.ForAll(frozenset([X]), eq),
        "forall_multi_sorted": lambda: lg.ForAll(frozenset([Z, A]), eqAZ),
        "exists": lambda: lg.Exists(frozenset([X]), eq),
        "lambda": lambda: lg.Lambda((X,), X),
        "named_binder_nil_env": lambda: lg.NamedBinder("nb", (X,), None, eq),
        "named_binder_with_env": lambda: lg.NamedBinder("nb", (X,), "e1", eq),

        # Definitions
        "definition": lambda: il.Definition(X, Y),
        "definition_schema": lambda: il.DefinitionSchema(X, Y),

        # ivylogic types
        "some_basic": lambda: il.Some(X, eq),
        "some_with_else": lambda: il.Some(X, eq, X, Y),
        "let": lambda: il.Let(il.Definition(X, Y), X),
        "literal_pos": lambda: il.Literal(1, eq),
        "literal_neg": lambda: il.Literal(0, eq),

        # Clauses
        "clauses_basic": lambda: lut.Clauses(fmlas=[eq], defs=[il.Definition(X, Y)]),
        "clauses_empty": lambda: lut.Clauses(fmlas=[], defs=[]),
    }

    if vec_id not in builders:
        return None
    return builders[vec_id]()


def test_all_vectors():
    """Test sexp() output matches golden vectors."""
    vectors_path = os.path.join(os.path.dirname(__file__), "sexp_vectors.sexp")
    vectors = parse_vectors(vectors_path)

    if not vectors:
        print("FAIL: no vectors loaded from %s" % vectors_path)
        sys.exit(1)

    failures = []
    for vec_id, vec_type, expected in vectors:
        node = build_node(vec_id)
        if node is None:
            failures.append((vec_id, "NO BUILDER", expected, ""))
            continue
        actual = node.sexp()
        if actual != expected:
            failures.append((vec_id, vec_type, expected, actual))

    if failures:
        for fid, ftype, exp, act in failures:
            print("FAIL %s (%s):" % (fid, ftype))
            print("  expected: %s" % exp)
            print("  actual:   %s" % act)
        sys.exit(1)

    print("PASS: all %d sexp vectors match" % len(vectors))


def test_canon_equals_sexp():
    """Verify canon() == sexp() for every node type."""
    vectors_path = os.path.join(os.path.dirname(__file__), "sexp_vectors.sexp")
    vectors = parse_vectors(vectors_path)

    failures = []
    for vec_id, vec_type, _ in vectors:
        node = build_node(vec_id)
        if node is None:
            continue
        s = node.sexp()
        c = node.canon()
        if s != c:
            failures.append((vec_id, s, c))

    if failures:
        for fid, s, c in failures:
            print("FAIL canon!=sexp for %s:" % fid)
            print("  sexp:  %s" % s)
            print("  canon: %s" % c)
        sys.exit(1)

    print("PASS: canon == sexp for all %d types" % len(vectors))


if __name__ == '__main__':
    test_all_vectors()
    test_canon_equals_sexp()
