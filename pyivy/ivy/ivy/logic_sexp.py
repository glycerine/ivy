# logic_sexp.py — Adds sexp() methods to all logic IR types.
#
# Matches Go's logic/sexp.go format exactly so that both sides produce
# identical structural identity strings for cross-language comparison.
#
# Usage: call install() once at startup (same pattern as canon_ast.py).

def _vars_sexp(variables):
    """Convert a collection of variables to sorted sexp string.
    Python ForAll/Exists use frozenset; must sort deterministically.
    Sort by (name, sort_sexp) to match Go's ordered slice."""
    sorted_vars = sorted(variables, key=lambda v: (v.name, v.sort.sexp()))
    return '[%s]' % ' '.join(v.sexp() for v in sorted_vars)


# --- Sort types ---

def _uninterpreted_sort_sexp(self):
    return '(UninterpretedSort name:%s)' % self.name

def _boolean_sort_sexp(self):
    return '(BooleanSort)'

def _function_sort_sexp(self):
    parts = ' '.join(s.sexp() for s in self.sorts)
    return '(FunctionSort sorts:[%s])' % parts

def _enumerated_sort_sexp(self):
    # NOTE: comma-separated (not space), matching Go
    return '(EnumeratedSort name:%s ext:[%s])' % (self.name, ','.join(self.extension))

def _range_sort_sexp(self):
    # Go uses BoundString() which is just str(bound)
    return '(RangeSort name:%s lb:%s ub:%s)' % (self.name, self.lb, self.ub)

def _top_sort_sexp(self):
    return '(TopSort name:%s)' % self.name


# --- Term types ---

def _var_sexp(self):
    return '(Variable name:%s sort:%s)' % (self.name, self.sort.sexp())

def _const_sexp(self):
    # Go's Symbol pre-computes sexp; format: (Symbol name:X sort:(...))
    return '(Symbol name:%s sort:%s)' % (self.name, self.sort.sexp())

def _apply_sexp(self):
    parts = ' '.join(t.sexp() for t in self.terms)
    return '(Apply func:%s terms:[%s])' % (self.func.sexp(), parts)


# --- Formula types ---

def _eq_sexp(self):
    return '(Eq t1:%s t2:%s)' % (self.t1.sexp(), self.t2.sexp())

def _not_sexp(self):
    return '(Not body:%s)' % self.body.sexp()

def _and_sexp(self):
    return '(And terms:[%s])' % ' '.join(t.sexp() for t in self.terms)

def _or_sexp(self):
    return '(Or terms:[%s])' % ' '.join(t.sexp() for t in self.terms)

def _implies_sexp(self):
    return '(Implies t1:%s t2:%s)' % (self.t1.sexp(), self.t2.sexp())

def _iff_sexp(self):
    return '(Iff t1:%s t2:%s)' % (self.t1.sexp(), self.t2.sexp())

def _ite_sexp(self):
    return '(Ite cond:%s then:%s else:%s)' % (
        self.cond.sexp(), self.t_then.sexp(), self.t_else.sexp())

def _globally_sexp(self):
    env = 'nil' if self.environ is None else self.environ
    return '(Globally environ:%s body:%s)' % (env, self.body.sexp())

def _eventually_sexp(self):
    env = 'nil' if self.environ is None else self.environ
    return '(Eventually environ:%s body:%s)' % (env, self.body.sexp())

def _when_operator_sexp(self):
    return '(WhenOperator name:%s t1:%s t2:%s)' % (
        self.name, self.t1.sexp(), self.t2.sexp())

def _cond_sexp(self):
    return '(Cond t1:%s t2:%s)' % (self.t1.sexp(), self.t2.sexp())


# --- Quantifier types (vars: key, sorted variables) ---

def _forall_sexp(self):
    return '(ForAll vars:%s body:%s)' % (_vars_sexp(self.variables), self.body.sexp())

def _exists_sexp(self):
    return '(Exists vars:%s body:%s)' % (_vars_sexp(self.variables), self.body.sexp())

def _lambda_sexp(self):
    return '(Lambda vars:%s body:%s)' % (_vars_sexp(self.variables), self.body.sexp())

def _named_binder_sexp(self):
    env = 'nil' if self.environ is None else self.environ
    return '(NamedBinder name:%s environ:%s vars:%s body:%s)' % (
        self.name, env, _vars_sexp(self.variables), self.body.sexp())


# --- Definition types (in ivy_logic.py, not logic.py) ---

def _definition_sexp(self):
    return '(Def lhs:%s rhs:%s)' % (self.args[0].sexp(), self.args[1].sexp())

def _definition_schema_sexp(self):
    return '(DefSchema lhs:%s rhs:%s)' % (self.args[0].sexp(), self.args[1].sexp())


# --- ivy_logic types: Some, Let, Literal ---
# Matches Go's ivylogic/sexp.go

def _some_sexp(self):
    params = ' '.join(p.sexp() for p in self.params())
    fmla = self.fmla().sexp()
    if_val = self.if_value()
    else_val = self.else_value()
    iv = 'nil' if if_val is None else if_val.sexp()
    ev = 'nil' if else_val is None else else_val.sexp()
    return '(Some params:[%s] fmla:%s ifVal:%s elseVal:%s)' % (params, fmla, iv, ev)

def _let_sexp(self):
    # Let.args = [def1, def2, ..., body]  (last arg is body)
    defs = ' '.join(d.sexp() for d in self.args[:-1])
    body = self.args[-1].sexp()
    return '(Let defs:[%s] body:%s)' % (defs, body)

def _literal_sexp(self):
    return '(Literal polarity:%d atom:%s)' % (self.polarity, self.atom.sexp())


# --- Clauses (in ivy_logic_utils.py) ---

def _clauses_sexp(self):
    fmlas = ' '.join(f.sexp() for f in self.fmlas)
    defs = ' '.join(d.sexp() for d in self.defs)
    return '(clauses fmlas:[%s] defs:[%s])' % (fmlas, defs)


def action_canon(a):
    """Generic canonical s-expression for any Action using name() and args.
    Format: (actionName args:[arg1.sexp() arg2.sexp() ...])
    Matches Go's module.ActionCanon()."""
    if a is None:
        return 'nil'
    arg_strs = []
    for arg in a.args:
        if arg is None:
            arg_strs.append('nil')
        elif hasattr(arg, 'sexp'):
            arg_strs.append(arg.sexp())
        elif hasattr(arg, 'canon'):
            arg_strs.append(arg.canon())
        else:
            arg_strs.append(str(arg))
    return '(%s args:[%s])' % (a.name(), ' '.join(arg_strs))


def install():
    """Monkey-patch sexp() onto all logic IR types. Call once at startup."""
    from . import logic as lg
    from . import ivy_logic as il
    from . import ivy_logic_utils as lut

    # Sort types
    lg.UninterpretedSort.sexp = _uninterpreted_sort_sexp
    lg.BooleanSort.sexp = _boolean_sort_sexp
    lg.FunctionSort.sexp = _function_sort_sexp
    lg.EnumeratedSort.sexp = _enumerated_sort_sexp
    lg.RangeSort.sexp = _range_sort_sexp
    lg.TopSort.sexp = _top_sort_sexp

    # Term types
    lg.Var.sexp = _var_sexp
    lg.Const.sexp = _const_sexp
    lg.Apply.sexp = _apply_sexp

    # Formula types
    lg.Eq.sexp = _eq_sexp
    lg.Not.sexp = _not_sexp
    lg.And.sexp = _and_sexp
    lg.Or.sexp = _or_sexp
    lg.Implies.sexp = _implies_sexp
    lg.Iff.sexp = _iff_sexp
    lg.Ite.sexp = _ite_sexp
    lg.Globally.sexp = _globally_sexp
    lg.Eventually.sexp = _eventually_sexp
    lg.WhenOperator.sexp = _when_operator_sexp
    lg.Cond.sexp = _cond_sexp

    # Quantifier types
    lg.ForAll.sexp = _forall_sexp
    lg.Exists.sexp = _exists_sexp
    lg.Lambda.sexp = _lambda_sexp
    lg.NamedBinder.sexp = _named_binder_sexp

    # Definition types (from ivy_logic, not logic)
    il.Definition.sexp = _definition_sexp
    il.DefinitionSchema.sexp = _definition_schema_sexp

    # Some, Let, Literal (from ivy_logic)
    il.Some.sexp = _some_sexp
    il.Let.sexp = _let_sexp
    il.Literal.sexp = _literal_sexp

    # Clauses
    lut.Clauses.sexp = _clauses_sexp

    # --- canon() = sexp() for all types (matching Go's logic/canon.go pattern) ---
    # Go's Canon() just wraps Sexp(), so canon = sexp.

    # Sort types
    lg.UninterpretedSort.canon = _uninterpreted_sort_sexp
    lg.BooleanSort.canon = _boolean_sort_sexp
    lg.FunctionSort.canon = _function_sort_sexp
    lg.EnumeratedSort.canon = _enumerated_sort_sexp
    lg.RangeSort.canon = _range_sort_sexp
    lg.TopSort.canon = _top_sort_sexp

    # Term types
    lg.Var.canon = _var_sexp
    lg.Const.canon = _const_sexp
    lg.Apply.canon = _apply_sexp

    # Formula types
    lg.Eq.canon = _eq_sexp
    lg.Not.canon = _not_sexp
    lg.And.canon = _and_sexp
    lg.Or.canon = _or_sexp
    lg.Implies.canon = _implies_sexp
    lg.Iff.canon = _iff_sexp
    lg.Ite.canon = _ite_sexp
    lg.Globally.canon = _globally_sexp
    lg.Eventually.canon = _eventually_sexp
    lg.WhenOperator.canon = _when_operator_sexp
    lg.Cond.canon = _cond_sexp

    # Quantifier types
    lg.ForAll.canon = _forall_sexp
    lg.Exists.canon = _exists_sexp
    lg.Lambda.canon = _lambda_sexp
    lg.NamedBinder.canon = _named_binder_sexp

    # Definition types
    il.Definition.canon = _definition_sexp
    il.DefinitionSchema.canon = _definition_schema_sexp

    # Some, Let, Literal
    il.Some.canon = _some_sexp
    il.Let.canon = _let_sexp
    il.Literal.canon = _literal_sexp

    # Clauses
    lut.Clauses.canon = _clauses_sexp
