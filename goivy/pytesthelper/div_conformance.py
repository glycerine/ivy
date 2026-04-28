#!/usr/bin/env python3
"""
Python driver for DIV conformance tests.
Each function exercises the Python (correct) behavior for one divergence.
Output format: key=value lines, parsed by Go tests.

Usage: python3 div_conformance.py div2
"""
import sys, os
sys.path.insert(0, os.path.expanduser('~/ivy/pyivy/ivy'))

from ivy import logic as lg
from ivy import ivy_logic_utils as ilu
from ivy import logic_util as lu


def div2():
    """SubstituteByName: Lambda does NOT block substitution in Python."""
    S = lg.UninterpretedSort('S')
    X = lg.Var('X', S)
    Y = lg.Var('Y', S)
    Z = lg.Var('Z', S)
    body = lg.Eq(X, Y)
    lam = lg.Lambda([X], body)
    result = ilu.substitute_ast(lam, {'X': Z})
    print('body_t1_name=' + result.body.t1.name)


def div4():
    """SubstituteApply: Python asserts no new free variables."""
    S = lg.UninterpretedSort('S')
    X = lg.Var('X', S)
    Z = lg.Var('Z', S)
    fs = lg.FunctionSort(S, S)
    f = lg.Const('f', fs)
    app = f(X)

    def bad_sub(*terms):
        return lg.Eq(terms[0], Z)

    try:
        lu.substitute_apply(app, {f: bad_sub})
        print('assertion=no')
    except AssertionError:
        print('assertion=yes')


def div5():
    """variables_ast: Some's bound variable is excluded in Python."""
    S = lg.UninterpretedSort('S')
    X = lg.Var('X', S)
    Y = lg.Var('Y', S)
    from ivy.ivy_logic import Some
    some = Some(X, lg.Eq(X, Y))
    fvs = list(ilu.variables_ast(some))
    names = sorted(v.name for v in fvs)
    print('free_vars=' + ','.join(names))


def div7():
    """NormalizeQuantifiers: Python crashes on Lambda."""
    S = lg.UninterpretedSort('S')
    X = lg.Var('X', S)
    body = lg.Eq(X, X)
    lam = lg.Lambda([X], body)
    try:
        lu.normalize_quantifiers(lam)
        print('panics=no')
    except AssertionError:
        print('panics=yes')


def div8():
    """CloseEPR: 3-variable mixed-sort ordering test."""
    S = lg.UninterpretedSort('S')
    T = lg.UninterpretedSort('T')
    B_T = lg.Var('B', T)
    C_S = lg.Var('C', S)
    A_S = lg.Var('A', S)
    fmla = lg.Implies(lg.Eq(B_T, B_T), lg.Eq(C_S, A_S))
    result = ilu.close_epr(fmla)
    if hasattr(result, 'variables'):
        names = [v.name for v in result.variables]
        print('var_order=' + ','.join(names))
    else:
        print('var_order=none')


if __name__ == '__main__':
    if len(sys.argv) < 2:
        print('usage: div_conformance.py <div_name>')
        sys.exit(1)
    fn = globals().get(sys.argv[1])
    if fn is None:
        print('unknown function: ' + sys.argv[1])
        sys.exit(1)
    fn()
