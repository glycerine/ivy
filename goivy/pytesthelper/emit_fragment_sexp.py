#!/usr/bin/env python3
"""Emit sexp strings for fragment canon types for cross-language comparison.

Output format: id<TAB>sexp per line.
Used by Go's TestFragmentSexpCrossLanguage in fragment/canon_test.go.
"""

import sys
import os
import types

# ---- Mock Z3 module ----
class _Z3Mock:
    def __getattr__(self, name):
        if name == '__dict__':
            return {'_to_ast_array': True, '_to_expr_ref': True}
        if name == '__file__':
            return '<z3_mock>'
        return self
    def __call__(self, *args, **kwargs):
        return self
    def __bool__(self):
        return True
    def __iter__(self):
        return iter([])
    def __contains__(self, item):
        return True

z3_mock = types.ModuleType('z3')
z3_inst = _Z3Mock()
for attr in ['_to_ast_array', '_to_expr_ref', 'z3', 'set_param', 'SolverFor',
             'BoolSort', 'IntSort', 'RealSort', 'is_and', 'is_true', 'is_false',
             'is_quantifier', 'is_const', 'is_app', 'is_not', 'is_var',
             'And', 'Or', 'Not', 'Implies', 'ForAll', 'Exists', 'Bool', 'Int',
             'Fixedpoint', 'sat', 'unsat', 'unknown', 'Solver', 'Goal',
             'Tactic', 'Then', 'AstVector', 'AstRef', 'ExprRef', 'BoolRef',
             'ArithRef', 'BitVecRef', 'ArrayRef', 'DatatypeRef', 'FuncDeclRef',
             'is_eq_ast', 'substitute', 'simplify']:
    setattr(z3_mock, attr, z3_inst)
z3_mock.__dict__['_to_ast_array'] = z3_inst
z3_mock.__dict__['_to_expr_ref'] = z3_inst
z3_mock.z3 = z3_mock
sys.modules['z3'] = z3_mock
sys.modules['ivy.z3'] = z3_mock

# Mock pattern.debug.profile
pattern_mock = types.ModuleType('pattern')
pattern_debug = types.ModuleType('pattern.debug')
pattern_profile = types.ModuleType('pattern.debug.profile')
pattern_profile.Stopwatch = lambda: None
sys.modules['pattern'] = pattern_mock
sys.modules['pattern.debug'] = pattern_debug
sys.modules['pattern.debug.profile'] = pattern_profile

# ---- Import Ivy ----
ivy_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, ivy_dir)

from ivy import logic as lg
from ivy import ivy_logic as il
from ivy import logic_sexp
import ivy.ivy_union_find2 as uf2
from ivy.ivy_union_find2 import UFNode
from ivy.canon_fragment import (
    uf_node_canon, uf_node_set_canon, strat_key_canon,
    strat_entry_canon, arc_canon, var_id_canon,
    map_fmla_res_canon, skolem_entry_canon, fmla_pair_canon,
)
from ivy.canon import node_canon

logic_sexp.install()


def emit(id_, sexp):
    print("%s\t%s" % (id_, sexp))


# ---- Reset UFNode counter for deterministic IDs ----
uf2.ufidctr = 0

# ---- Common building blocks ----
S = lg.UninterpretedSort("S")
X = lg.Var("X", S)
Y = lg.Var("Y", S)
eq = lg.Eq(X, Y)
fs = lg.FunctionSort(S, S, lg.Boolean)
sym = lg.Const("f", fs)

# UFNodes: ids will be 0, 1, 2
n0 = UFNode()  # id 0
n1 = UFNode()  # id 1
n2 = UFNode()  # id 2

# ---- UFNode vectors ----
emit("ufnode_basic", uf_node_canon(n0))
emit("ufnode_nil", uf_node_canon(None))
emit("ufnode_set_empty", uf_node_set_canon(set()))
emit("ufnode_set_three", uf_node_set_canon({n2, n0, n1}))
emit("ufnode_set_nil", uf_node_set_canon(None))

# ---- varID vectors ----
emit("varID_simple", var_id_canon(X))
emit("varID_other", var_id_canon(Y))

# ---- strat_key vectors ----
emit("strat_key_var", strat_key_canon(X))
emit("strat_key_app", strat_key_canon((sym, 2)))
# Sort equality key: Symbol('=', sort) — note Python uses il.Symbol = lg.Const
sort_eq_sym = il.Symbol('=', S)
emit("strat_key_sort", strat_key_canon(sort_eq_sym))

# ---- stratEntry vectors ----
emit("strat_entry_var", strat_entry_canon(X))
emit("strat_entry_app", strat_entry_canon((sym, 1)))
emit("strat_entry_sort", strat_entry_canon(sort_eq_sym))

# ---- arc vectors ----
emit("arc_with_idx", arc_canon((n0, n1, eq, 42, 1)))
emit("arc_no_idx", arc_canon((n0, n1, eq, 42)))
emit("arc_nil_fmla", arc_canon((n0, n1, None, 10)))
emit("arc_nil_nodes", arc_canon((None, None, eq, 5)))

# ---- mapFmlaRes vectors ----
emit("map_fmla_res_basic", map_fmla_res_canon((n0, {n1, n2})))
emit("map_fmla_res_nil", map_fmla_res_canon((None, None)))
emit("map_fmla_res_empty_uvs", map_fmla_res_canon((n1, set())))

# ---- skolemEntry vectors ----
emit("skolem_entry_basic", skolem_entry_canon((eq, X)))
emit("skolem_entry_nil", skolem_entry_canon((None, None)))

# ---- fmlaPair vectors ----
emit("fmla_pair_basic", fmla_pair_canon(eq, X, 15))
emit("fmla_pair_nil_source", fmla_pair_canon(eq, None, 7))
emit("fmla_pair_nil_fmla", fmla_pair_canon(None, None, 0))

# ---- FragmentError vector ----
# Python doesn't have FragmentError directly; emit the expected format
emit("fragment_error", '(fragmentError message:"test error message")')
