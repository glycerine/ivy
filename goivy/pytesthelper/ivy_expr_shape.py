#!/usr/bin/env python3
"""
Long-running Python oracle for cross-validating the Go LALR parser
against the real Python Ivy parser.

Communicates via stdin/stdout line protocol:

  Go -> Python:  EXPR formula_text
  Go -> Python:  ACTION action_body
  Python -> Go:  OK shape_string
  Python -> Go:  ERR error_message

  Batch mode:
  Go -> Python:  BATCH
  Go -> Python:  EXPR formula1
  Go -> Python:  EXPR formula2
  Go -> Python:  END
  Python -> Go:  OK shape1
  Python -> Go:  OK shape2

The shape_string matches the Go astShape() format in
lalr_logicparser/lalr_parser_test.go, e.g.:
  And(Atom(a),Atom(b))
  Forall([Var(X)],Implies(Atom(p,[Var(X)]),Atom(q,[Var(X)])))

Mocks Z3 and pattern.debug so it runs without heavy dependencies.
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

from ivy import ivy_ast
from ivy import ivy_utils as iu
from ivy import ivy_parser

# Suppress "It is recommended parenthesize nested -> operators" warnings.
# These fire on valid non-parenthesized formulas that we intentionally test.
_orig_warn = iu.warn
def _quiet_warn(node, msg):
    if 'parenthesize nested' in msg:
        return
    _orig_warn(node, msg)
iu.warn = _quiet_warn


# ---- AST shape formatting (matches Go astShape) ----

def flatten_and(node):
    """Flatten nested And nodes into a flat list, matching Go's flattenAnd."""
    result = []
    for t in node.args:
        if type(t).__name__ == 'And' and len(t.args) > 0:
            result.extend(flatten_and(t))
        else:
            result.append(t)
    return result

def flatten_or(node):
    """Flatten nested Or nodes into a flat list, matching Go's flattenOr."""
    result = []
    for t in node.args:
        if type(t).__name__ == 'Or' and len(t.args) > 0:
            result.extend(flatten_or(t))
        else:
            result.append(t)
    return result

def shape_list(nodes):
    return ','.join(ast_shape(n) for n in nodes)

def ast_shape(node):
    """
    Produce shape string matching Go's astShape() format.

    The format is defined in lalr_logicparser/lalr_parser_test.go.
    Symbol is normalized to Atom (matching normalizeAtomSymbol=true).
    And/Or are flattened (matching flattenAnd/flattenOr).
    """
    if node is None:
        return '<nil>'

    name = type(node).__name__

    if name == 'Symbol':
        return 'Atom(%s)' % node.rep

    if name == 'Variable':
        return 'Var(%s)' % node.rep

    if name == 'Atom':
        if not node.args:
            return 'Atom(%s)' % node.rep
        return 'Atom(%s,[%s])' % (node.rep, shape_list(node.args))

    if name == 'App':
        rep = node.rep if isinstance(node.rep, str) else str(node.rep)
        if not node.args:
            return 'App(%s,[])' % rep
        return 'App(%s,[%s])' % (rep, shape_list(node.args))

    if name == 'And':
        if not node.args:
            return 'True'
        flat = flatten_and(node)
        return 'And(%s)' % shape_list(flat)

    if name == 'Or':
        if not node.args:
            return 'False'
        flat = flatten_or(node)
        return 'Or(%s)' % shape_list(flat)

    if name == 'Not':
        return 'Not(%s)' % ast_shape(node.args[0])

    if name == 'Implies':
        return 'Implies(%s,%s)' % (ast_shape(node.args[0]), ast_shape(node.args[1]))

    if name == 'Iff':
        return 'Iff(%s,%s)' % (ast_shape(node.args[0]), ast_shape(node.args[1]))

    if name == 'Ite':
        return 'Ite(%s,%s,%s)' % (ast_shape(node.args[0]), ast_shape(node.args[1]), ast_shape(node.args[2]))

    if name in ('Forall', 'Quantifier') and hasattr(node, 'bounds'):
        # ivy_ast.Forall: .bounds is the bound variables, .args[0] is the body
        return 'Forall([%s],%s)' % (shape_list(node.bounds), ast_shape(node.args[0]))

    if name == 'Exists':
        return 'Exists([%s],%s)' % (shape_list(node.bounds), ast_shape(node.args[0]))

    if name == 'Globally':
        return 'Globally(%s)' % ast_shape(node.args[0])

    if name == 'Eventually':
        return 'Eventually(%s)' % ast_shape(node.args[0])

    if name == 'WhenOperator':
        return 'When(%s,%s,%s)' % (node.name, ast_shape(node.args[0]), ast_shape(node.args[1]))

    if name == 'Old':
        return 'Old(%s)' % ast_shape(node.args[0])

    if name == 'This':
        return 'This'

    if name == 'MethodCall':
        return 'MethodCall(%s,%s)' % (ast_shape(node.args[0]), ast_shape(node.args[1]))

    if name == 'Isa':
        return 'Isa(%s)' % shape_list(node.args)

    if name == 'NamedBinder':
        return 'NamedBinder(%s,%s)' % (node.name, ast_shape(node.body))

    if name == 'Dot':
        return 'Dot(%s,%s)' % (ast_shape(node.args[0]), ast_shape(node.args[1]))

    # --- Action and declaration types for action cross-validation ---
    if name == 'LabeledFormula':
        label = node.args[0]
        formula = node.args[1] if len(node.args) > 1 else None
        return 'LabeledFormula(%s,%s)' % (ast_shape(label), ast_shape(formula))

    if name == 'Sequence':
        return 'Sequence(%s)' % shape_list(node.args)

    if name in ('AssertAction', 'AssumeAction', 'AssignAction',
                'CallAction', 'IfAction', 'WhileAction',
                'LocalAction', 'NativeAction', 'HavocAction',
                'RequiresAction', 'EnsuresAction', 'CrashAction'):
        return '%s(%s)' % (name, shape_list(node.args))

    # Fallback: try to produce something useful
    if hasattr(node, 'args') and node.args:
        return '%s(%s)' % (name, shape_list(node.args))
    if hasattr(node, 'rep'):
        return '%s(%s)' % (name, node.rep)
    return '?(%s)' % name


# ---- Parsing helpers ----

def set_version(ver_str):
    """Set the Ivy language version for parsing."""
    iu.set_string_version(ver_str)

def parse_expr(formula_text, version='1.7'):
    """
    Parse a single formula by wrapping it in a property declaration.
    We use 'property' rather than 'axiom' because axioms reject
    temporal operators (globally, eventually), while properties allow them.
    Returns the formula AST node extracted from the declaration.
    """
    # Wrap in a minimal Ivy file so the full parser can handle it.
    # Use 'temporal property' to allow both temporal (globally, eventually)
    # and non-temporal formulas.
    wrapped = '#lang ivy%s\ntemporal property [_crossval] %s' % (version, formula_text)
    set_version(version)
    try:
        result = ivy_parser.parse(wrapped)
    except Exception as e:
        raise ValueError('parse error: %s' % e)

    # Extract the formula from the axiom declaration
    if not result.decls:
        raise ValueError('no declarations produced')

    decl = result.decls[-1]  # the axiom we added
    # AxiomDecl wraps a LabeledFormula; extract the formula
    if hasattr(decl, 'args') and decl.args:
        lf = decl.args[0]
        if hasattr(lf, 'formula'):
            return lf.formula
        if hasattr(lf, 'args') and len(lf.args) > 1:
            return lf.args[1]  # args[1] is the formula
        return lf
    raise ValueError('cannot extract formula from declaration: %s' % type(decl).__name__)

def parse_action(action_text, version='1.7'):
    """
    Parse an action body by wrapping it in an action declaration.
    Returns the action body AST node.
    """
    wrapped = '#lang ivy%s\naction _crossval = {\n%s\n}' % (version, action_text)
    set_version(version)
    try:
        result = ivy_parser.parse(wrapped)
    except Exception as e:
        raise ValueError('parse error: %s' % e)

    if not result.decls:
        raise ValueError('no declarations produced')

    decl = result.decls[-1]
    # ActionDecl has 1 arg: an ActionDef.
    # ActionDef has 2 args: [name_atom, body].
    # The body is at decl.args[0].args[1].
    if hasattr(decl, 'args') and decl.args:
        action_def = decl.args[0]
        if hasattr(action_def, 'args') and len(action_def.args) > 1:
            return action_def.args[1]
    raise ValueError('cannot extract action body from declaration: %s' % type(decl).__name__)


# ---- Stdout capture to suppress warnings during parsing ----

import io

class _StdoutCapture:
    """Context manager that redirects stdout to a buffer during parsing,
    so warnings from iu.warn() don't corrupt the line protocol.
    Captured output is forwarded to stderr on exit so it remains visible
    for debugging."""
    def __enter__(self):
        self._real = sys.stdout
        self._buf = io.StringIO()
        sys.stdout = self._buf
        return self
    def __exit__(self, *args):
        sys.stdout = self._real
        captured = self._buf.getvalue()
        if captured:
            sys.stderr.write(captured)
            sys.stderr.flush()

_capture = _StdoutCapture()


# ---- Main REPL loop ----

def main():
    # Line-buffered output for reliable pipe communication
    real_stdout = os.fdopen(sys.stdout.fileno(), 'w', buffering=1)
    sys.stdout = real_stdout

    # Signal readiness
    print('READY', flush=True)

    batch_mode = False
    batch_results = []

    for line in sys.stdin:
        line = line.rstrip('\n')
        if not line:
            continue

        if line == 'BATCH':
            batch_mode = True
            batch_results = []
            continue

        if line == 'END' and batch_mode:
            for r in batch_results:
                print(r, flush=True)
            batch_mode = False
            batch_results = []
            continue

        # Parse the command.
        # Redirect stdout to a buffer during parsing to capture any
        # warnings (e.g. iu.warn) that would corrupt the line protocol.
        if line.startswith('EXPR '):
            formula_text = line[5:]
            try:
                with _capture:
                    node = parse_expr(formula_text)
                result = 'OK %s' % ast_shape(node)
            except Exception as e:
                result = 'ERR %s' % str(e).replace('\n', ' ')
        elif line.startswith('ACTION '):
            action_text = line[7:]
            try:
                with _capture:
                    node = parse_action(action_text)
                result = 'OK %s' % ast_shape(node)
            except Exception as e:
                result = 'ERR %s' % str(e).replace('\n', ' ')
        else:
            result = 'ERR unknown command: %s' % line

        if batch_mode:
            batch_results.append(result)
        else:
            print(result, flush=True)


if __name__ == '__main__':
    main()
