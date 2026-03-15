#!/usr/bin/env python3
"""
Standalone Ivy parser AST dumper for comparison testing.

Usage: python3 ivy_ast_dump.py [--version 1.7] < input.ivy
       python3 ivy_ast_dump.py [--version 1.7] input.ivy

Outputs a canonical text representation of the AST, one line per declaration,
so the Go compiler's output can be compared against this reference.

This script mocks out Z3 and other heavy dependencies so it can run
without Z3 installed.
"""

import sys
import os
import types
import argparse

# Mock z3 module so ivy_solver doesn't crash on import
class _Z3Mock:
    """Mock Z3 module with enough stubs to let ivy_solver import."""
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

# Now import ivy
ivy_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, ivy_dir)

from ivy import ivy_ast
from ivy import ivy_lexer

# Try to import the parser — this may fail if deps are too tangled.
# If it fails, we'll use a simpler approach.
try:
    from ivy import ivy_parser
    HAS_PARSER = True
except Exception as e:
    HAS_PARSER = False
    PARSER_ERROR = str(e)


def dump_node(node, indent=0):
    """Recursively dump an AST node in canonical text format."""
    prefix = "  " * indent
    name = type(node).__name__

    if hasattr(node, 'rep'):
        rep = node.rep
    elif hasattr(node, 'name'):
        rep = str(node.name) if node.name else ''
    else:
        rep = ''

    # Get args/children
    args = []
    if hasattr(node, 'args'):
        args = list(node.args)

    line = "{}{}".format(name, (":" + rep) if rep else "")

    if not args:
        return prefix + line

    result = prefix + line
    for a in args:
        if a is None:
            result += "\n" + prefix + "  <nil>"
        elif isinstance(a, ivy_ast.AST):
            result += "\n" + dump_node(a, indent + 1)
        elif isinstance(a, (list, tuple)):
            for item in a:
                if isinstance(item, ivy_ast.AST):
                    result += "\n" + dump_node(item, indent + 1)
                else:
                    result += "\n" + prefix + "  " + repr(item)
        else:
            result += "\n" + prefix + "  " + repr(a)
    return result


def main():
    ap = argparse.ArgumentParser(description="Dump Ivy AST")
    ap.add_argument("file", nargs="?", help="Ivy source file (stdin if omitted)")
    ap.add_argument("--version", default="1.7", help="Ivy language version (default: 1.7)")
    args = ap.parse_args()

    # Read input
    if args.file:
        with open(args.file) as f:
            src = f.read()
    else:
        src = sys.stdin.read()

    # Strip #lang line
    if src.startswith('#lang'):
        src = src[src.index('\n')+1:]

    # Parse version
    parts = args.version.split('.')
    ver = [int(parts[0]), int(parts[1]) if len(parts) > 1 else 0]

    if not HAS_PARSER:
        print("ERROR: parser not available: {}".format(PARSER_ERROR), file=sys.stderr)
        sys.exit(1)

    # Set version and parse
    from ivy import ivy_utils as iu
    iu.set_string_version(args.version)
    try:
        result = ivy_parser.parse(src)
    except Exception as e:
        print("PARSE_ERROR: {}".format(e), file=sys.stderr)
        sys.exit(1)

    # Dump — result is an Ivy object with .decls containing declarations
    if hasattr(result, 'decls'):
        decls = list(result.decls)
    elif hasattr(result, 'body'):
        decls = list(result.body)
    elif hasattr(result, 'args'):
        decls = list(result.args)
    elif isinstance(result, list):
        decls = result
    else:
        decls = [result]

    # Output format: one line per declaration, showing:
    #   [index] DeclType name [details]
    # This canonical format avoids Python-specific transformations (fml: prefixes,
    # auto-labels, sort inference) so it can be compared with Go parser output.
    for i, decl in enumerate(decls):
        dtype = type(decl).__name__
        # Extract the declaration name in a stable way
        name = extract_decl_name(decl)
        print("[{}] {} {}".format(i, dtype, name).rstrip())


def extract_decl_name(decl):
    """Extract a stable name from a declaration for comparison."""
    dtype = type(decl).__name__

    if dtype == 'TypeDecl':
        # TypeDecl has args = [TypeDef] where TypeDef has a name
        for a in decl.args:
            if hasattr(a, 'name'):
                return str(a.name)
            return str(a).split('=')[0].strip()
        return ''

    if dtype in ('ActionDecl',):
        # ActionDecl has args = [ActionDef] where ActionDef.defines() gives name
        for a in decl.args:
            if hasattr(a, 'defines'):
                defs = a.defines()
                if isinstance(defs, (list, tuple)):
                    return str(defs[0]) if defs else ''
                return str(defs)
        return ''

    if dtype in ('RelationDecl', 'ConstantDecl'):
        for a in decl.args:
            if hasattr(a, 'rep'):
                # Atom: show name and arity
                rep = a.rep
                nargs = len(list(a.args)) if hasattr(a, 'args') else 0
                return "{}({})".format(rep, nargs) if nargs > 0 else rep
            return str(a)
        return ''

    if dtype in ('ConjectureDecl', 'PropertyDecl', 'AxiomDecl'):
        for a in decl.args:
            if hasattr(a, 'label') and a.label:
                return str(a.label)
            return str(a)[:60]
        return ''

    if dtype == 'ExportDecl':
        for a in decl.args:
            if hasattr(a, 'exported'):
                return str(a.exported())
            return str(a)[:40]
        return ''

    if dtype == 'MixinDecl':
        for a in decl.args:
            if hasattr(a, 'args'):
                inner = list(a.args)
                if inner:
                    return str(inner[0])[:40]
            return str(a)[:40]
        return ''

    if dtype == 'PrivateDecl':
        return 'private'

    # Fallback
    if hasattr(decl, 'defines'):
        d = decl.defines()
        if isinstance(d, (list, tuple)) and d:
            return str(d[0])
        return str(d) if d else ''

    return ''


if __name__ == '__main__':
    main()
