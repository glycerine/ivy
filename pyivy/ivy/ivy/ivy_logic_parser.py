#
# Copyright (c) Microsoft Corporation. All Rights Reserved.
#
# This file contains parser rules for first-order formulas
from .ivy_ast import *
from . import ivy_logic_utils
from . import ivy_utils as iu
from . import xtracer

# prefer xtracer.normalize_filename() instead. Does IVY_EXAMPLES too.
# def _normalize_filename(f):
#     """Replace include directory path with <IVY_INCLUDE> for canonical matching."""
#     if f is None:
#         return f
#     std_dir = iu.get_std_include_dir()
#     if std_dir:
#         import os.path
#         base_dir = os.path.dirname(std_dir)
#         if base_dir and not base_dir.endswith(os.sep):
#             base_dir += os.sep
#         if f.startswith(base_dir):
#             return '<IVY_INCLUDE>/' + f[len(base_dir):]
#     return f

def get_lineno(p,n):
    #if __debug__: xtracer.trace("parser.get_lineno ENTER")
    return iu.Location(xtracer.normalize_filename(iu.filename), p.lineno(n))

def symbol(s):
    if __debug__: xtracer.trace("parser.symbol ENTER")
    return Variable(s,universe) if str.isupper(s[0]) else Constant(s)

def p_SYMBOL_PRESYMBOL(p):
    'SYMBOL : PRESYMBOL'
    if __debug__: xtracer.trace("parser.p_SYMBOL_PRESYMBOL ENTER (SYMBOL) val=%s" % p[1])
    p[0] = p[1]

def p_SYMBOL_SYMBOL_LB_SYMsubscr_RB(p):
    'SYMBOL : SYMBOL LB SYMsubscr RB'
    if __debug__: xtracer.trace("parser.p_SYMBOL_SYMBOL_LB_SYMsubscr_RB ENTER (SYMBOL)")
    p[0] = p[1] + p[2] + p[3] + p[4]

def p_LABEL_LB_SYMBOL_RB(p):
    'LABEL : LB SYMBOL RB'
    if __debug__: xtracer.trace("parser.p_LABEL_LB_SYMBOL_RB ENTER (LABEL)")
    p[0] = p[1] + p[2] + p[3]

def p_SYMsubscr_SYMBOL(p):
    'SYMsubscr : SYMBOL'
    if __debug__: xtracer.trace("parser.p_SYMsubscr_SYMBOL ENTER (SYMsubscr)")
    p[0] = p[1]

def p_SYMsubscr_THIS(p):
    'SYMsubscr : THIS'
    if __debug__: xtracer.trace("parser.p_SYMsubscr_THIS ENTER (SYMsubscr)")
    p[0] = 'this'

def p_SYMsubscr_SYMsubscr_dot_symbol(p):
    'SYMsubscr : SYMsubscr DOT SYMBOL'
    if __debug__: xtracer.trace("parser.p_SYMsubscr_SYMsubscr_dot_symbol ENTER (SYMsubscr)")
    p[0] = p[1] + '.' + p[3]


def p_atype_symbol(p):
    'atype : SYMBOL'
    if __debug__: xtracer.trace("parser.p_atype_symbol ENTER (atype) val=%s" % p[1])
    p[0] = p[1]

if not (iu.get_numeric_version() <= [1,2]):
    def p_atype_atype_dot_symbol(p):
        'atype : atype DOT SYMBOL'
        if __debug__: xtracer.trace("parser.p_atype_atype_dot_symbol ENTER (atype)")
        if isinstance(p[1],This):
            p[0] = p[3]
        else:
            p[0] = p[1] + '.' + p[3]
    def p_atype_this(p):
        'atype : THIS'
        if __debug__: xtracer.trace("parser.p_atype_this ENTER (atype)")
        p[0] = This()
        p[0].lineno = get_lineno(p,1)

if iu.get_numeric_version() <= [1,6]:

    def p_aterm_symbol(p):
        'aterm : SYMBOL'
        if __debug__: xtracer.trace("parser.p_aterm_symbol ENTER (aterm)")
        p[0] = App(p[1])
        p[0].lineno = get_lineno(p,1)

    def p_aterm_aterm_terms(p):
        'aterm : aterm LPAREN terms RPAREN'
        if __debug__: xtracer.trace("parser.p_aterm_aterm_terms ENTER (aterm)")
        p[0] = p[1]
        if isinstance(p[0],MethodCall):
            p[0].args[1].args.extend(p[3])
        else:
            p[0].args.extend(p[3])

    if iu.get_numeric_version() <= [1,2]:

        def p_term_term_colon_term(p):
            'aterm : aterm COLON SYMBOL'
            if __debug__: xtracer.trace("parser.p_term_term_colon_term ENTER (aterm)")
            p[0] = compose_atoms(p[1],App(p[3]))
            p[0].lineno = get_lineno(p,2)

    else:

        if iu.get_numeric_version() <= [1,6]:
            def p_term_term_dot_term(p):
                'aterm : aterm DOT SYMBOL'
                if __debug__: xtracer.trace("parser.p_term_term_dot_term ENTER (aterm)")
                p[0] = compose_atoms(p[1],App(p[3]))
                p[0].lineno = get_lineno(p,2)

else:

    def p_appelem_symbol(p):
        'appelem : SYMBOL'
        if __debug__: xtracer.trace("parser.p_appelem_symbol ENTER (appelem)")
        p[0] = App(p[1])
        p[0].lineno = get_lineno(p,1)

    def p_appelem_appelem_terms(p):
        'appelem : SYMBOL LPAREN terms RPAREN'
        if __debug__: xtracer.trace("parser.p_appelem_appelem_terms ENTER (appelem)")
        p[0] = p[1]
        p[0] = App(p[1],p[3])
        p[0].lineno = get_lineno(p,1)


def p_var_variable(p):
    'var : VARIABLE'
    if __debug__: xtracer.trace("parser.p_var_variable ENTER (var)")
    p[0] = Variable(p[1],universe)
    p[0].lineno = get_lineno(p,1)

def p_var_variable_colon_symbol(p):
    'var : VARIABLE COLON atype'
    if __debug__: xtracer.trace("parser.p_var_variable_colon_symbol ENTER (var)")
    p[0] = Variable(p[1],p[3])
    p[0].lineno = get_lineno(p,1)

def p_simplevar_variable(p):
    'simplevar : VARIABLE'
    if __debug__: xtracer.trace("parser.p_simplevar_variable ENTER (simplevar)")
    p[0] = Variable(p[1],universe)
    p[0].lineno = get_lineno(p,1)

def p_simplevar_variable_colon_symbol(p):
    'simplevar : VARIABLE COLON SYMBOL'
    if __debug__: xtracer.trace("parser.p_simplevar_variable_colon_symbol ENTER (simplevar)")
    p[0] = Variable(p[1],p[3])
    p[0].lineno = get_lineno(p,1)

if iu.get_numeric_version() <= [1,6]:

    def p_term_aterm(p):
        'term : aterm'
        if __debug__: xtracer.trace("parser.p_term_aterm ENTER (term)")
        p[0] = p[1]

    if not (iu.get_numeric_version() <= [1,6]):
        def p_term_term_dot_aterm(p):
            'aterm : term DOT aterm'
            if __debug__: xtracer.trace("parser.p_term_term_dot_aterm ENTER (aterm)")
            if isinstance(p[1],(Atom,App)):
                p[0] = compose_atoms(p[1],p[3])
            else:
                p[0] = MethodCall(p[1],p[3])
            p[0].lineno = get_lineno(p,2)

    def p_aterm_old_symbol(p):
        'term : OLD aterm'
        if __debug__: xtracer.trace("parser.p_aterm_old_symbol ENTER (term)")
        p[0] = Old(p[2])
        p[0].lineno = get_lineno(p,1)

else:

    def p_term_aappelem(p):
        'term : appelem'
        if __debug__: xtracer.trace("parser.p_term_aappelem ENTER (term)")
        p[0] = p[1]
        
    def p_term_old_aappelem(p):
        'term : OLD appelem'
        if __debug__: xtracer.trace("parser.p_term_old_aappelem ENTER (term)")
        p[0] = Old(p[2])
        p[0].lineno = get_lineno(p,1)

    def p_term_dot_appelem(p):
        'term : term DOT appelem'
        if __debug__: xtracer.trace("parser.p_term_dot_appelem ENTER (term)")
        if isinstance(p[1],(Atom,App)):
            p[0] = compose_atoms(p[1],p[3])
            p[0].lineno = get_lineno(p,2)
        elif isinstance(p[1],Old):
            t = compose_atoms(p[1].args[0],p[3])
            t.lineno = get_lineno(p,2)
            p[0] = p[1]
            p[0].args[0] = t
        else:
            p[0] = MethodCall(p[1],p[3])
            p[0].lineno = get_lineno(p,2)
        
    def p_aterm_aappelem(p):
        'aterm : appelem'
        if __debug__: xtracer.trace("parser.p_aterm_aappelem ENTER (aterm)")
        p[0] = p[1]

    def p_aterm_aterm_dot_appelem(p):
        'aterm : aterm DOT appelem'
        if __debug__: xtracer.trace("parser.p_aterm_aterm_dot_appelem ENTER (aterm)")
        p[0] = compose_atoms(p[1],p[3])
        p[0].lineno = get_lineno(p,2)
        

    def p_term_term_colon_term(p):
        'term : term COLON atype'
        if __debug__: xtracer.trace("parser.p_term_term_colon_term ENTER (term)")
        if hasattr(p[1],"sort"):
            raise iu.IvyError(p[3],"multiple sort annotations")
        p[1].sort = p[3]
        p[0] = p[1]

def p_term_var(p):
    'term : var'
    if __debug__: xtracer.trace("parser.p_term_var ENTER (term)")
    p[0] = p[1]

if not (iu.get_numeric_version() <= [1,2]):

    def p_term_term_PLUS_term(p):
        'term : term PLUS term'
        if __debug__: xtracer.trace("parser.p_term_term_PLUS_term ENTER (term)")
        p[0] = App(p[2],p[1],p[3])
        p[0].lineno = get_lineno(p,2)

    def p_term_term_MINUS_term(p):
        'term : term MINUS term'
        if __debug__: xtracer.trace("parser.p_term_term_MINUS_term ENTER (term)")
        p[0] = App(p[2],p[1],p[3])
        p[0].lineno = get_lineno(p,2)

    def p_term_term_TIMES_term(p):
        'term : term TIMES term'
        if __debug__: xtracer.trace("parser.p_term_term_TIMES_term ENTER (term)")
        p[0] = App(p[2],p[1],p[3])
        p[0].lineno = get_lineno(p,2)

    def p_term_term_DIV_term(p):
        'term : term DIV term'
        if __debug__: xtracer.trace("parser.p_term_term_DIV_term ENTER (term)")
        p[0] = App(p[2],p[1],p[3])
        p[0].lineno = get_lineno(p,2)

    def p_term_if_fmla_else_term(p):
        'term : term IF fmla ELSE term'
        if __debug__: xtracer.trace("parser.p_term_if_fmla_else_term ENTER (term)")
        p[0] = Ite(p[3],p[1],p[5])
        p[0].lineno = get_lineno(p,2)
        if isinstance(p[1],Ite) and not hasattr(p[1],"parenthesized"):
            iu.warn(p[0],"It is recommended parenthesize nested if/else operators to avoid ambiguity.")

# if not (iu.get_numeric_version() <= [1,5]):

#     def p_term_term_and_term(p):
#         'term : term AND term'
#         if isinstance(p[1],And):
#             p[0] = p[1]
#             p[0].args.append(p[3])
#         else:
#             p[0] = And(p[1],p[3])
#             p[0].lineno = get_lineno(p,2)

#     def p_term_term_or_term(p):
#         'term : term OR term'
#         if isinstance(p[1],Or):
#             p[0] = p[1]
#             p[0].args.append(p[3])
#         else:
#             p[0] = Or(p[1],p[3])
#             p[0].lineno = get_lineno(p,2)


def p_terms(p):
    'terms : '
    if __debug__: xtracer.trace("parser.p_terms ENTER (terms)")
    p[0] = []

def p_terms_term(p):
    'terms : term'
    if __debug__: xtracer.trace("parser.p_terms_term ENTER (terms)")
    p[0] = [p[1]]

def p_terms_terms_term(p):
    'terms : terms COMMA term'
    if __debug__: xtracer.trace("parser.p_terms_terms_term ENTER (terms)")
    p[0] = p[1]
    p[0].append(p[3])


def p_term_lp_term_lp(p):
    'term : LPAREN term RPAREN'
    if __debug__: xtracer.trace("parser.p_term_lp_term_lp ENTER (term)")
    p[0] = p[2]
    p[0].parenthesized = True

def p_vars_var(p):
    'vars : var'
    if __debug__: xtracer.trace("parser.p_vars_var ENTER (vars)")
    p[0] = [p[1]]

def p_vars_vars_comma_var(p):
    'vars : vars COMMA var'
    if __debug__: xtracer.trace("parser.p_vars_vars_comma_var ENTER (vars)")
    p[0] = p[1]
    p[0].append(p[3])

def p_simplevars_simplevar(p):
    'simplevars : simplevar'
    if __debug__: xtracer.trace("parser.p_simplevars_simplevar ENTER (simplevars)")
    p[0] = [p[1]]

def p_simplevars_simplevars_comma_simplevar(p):
    'simplevars : simplevars COMMA simplevar'
    if __debug__: xtracer.trace("parser.p_simplevars_simplevars_comma_simplevar ENTER (simplevars)")
    p[0] = p[1]
    p[0].append(p[3])

# apps are terms of the form symbol or symbol(term*)

def p_app_symbol(p):
    'app : SYMBOL'
    if __debug__: xtracer.trace("parser.p_app_symbol ENTER (app)")
    p[0] = App(p[1],[])
    p[0].lineno = get_lineno(p,1)

def p_app_symbol_lp_terms_rp(p):
    'app : SYMBOL LPAREN terms RPAREN'
    if __debug__: xtracer.trace("parser.p_app_symbol_lp_terms_rp ENTER (app)")
    p[0] = App(p[1],p[3])
    p[0].lineno = get_lineno(p,1)

def p_app_term_infix_term(p):
    'app : term infix term'
    if __debug__: xtracer.trace("parser.p_app_term_infix_term ENTER (app)")
    p[0] = App(p[2],p[1],p[3])
    p[0].lineno = get_lineno(p,2)


def p_apps_app(p):
    'apps : app'
    if __debug__: xtracer.trace("parser.p_apps_app ENTER (apps)")
    p[0] = [p[1]]

def p_apps_apps_app(p):
    'apps : apps COMMA app'
    if __debug__: xtracer.trace("parser.p_apps_apps_app ENTER (apps)")
    p[0] = p[1]
    p[0].append(p[3])

# atoms are formulas just of the form symbol or symbol(term*)

def p_atom_symbol(p):
    'atom : SYMBOL'
    if __debug__: xtracer.trace("parser.p_atom_symbol ENTER (atom)")
    p[0] = Atom(p[1],[])
    p[0].lineno = get_lineno(p,1)

def p_atom_symbol_lp_terms_rp(p):
    'atom : SYMBOL LPAREN terms RPAREN'
    if __debug__: xtracer.trace("parser.p_atom_symbol_lp_terms_rp ENTER (atom)")
    p[0] = Atom(p[1],p[3])
    p[0].lineno = get_lineno(p,1)

def p_atoms_atom(p):
    'atoms : atom'
    if __debug__: xtracer.trace("parser.p_atoms_atom ENTER (atoms)")
    p[0] = [p[1]]

def p_atoms_atoms_atom(p):
    'atoms : atoms COMMA atom'
    if __debug__: xtracer.trace("parser.p_atoms_atoms_atom ENTER (atoms)")
    p[0] = p[1]
    p[0].append(p[3])

# literal is an atom or its negation

def p_lit_atom(p):
    'lit : atom'
    if __debug__: xtracer.trace("parser.p_lit_atom ENTER (lit)")
    p[0] = Literal(1,p[1])
    p[0].lineno = get_lineno(p,1)

def p_lit_term_eq_term(p):
    'lit : SYMBOL EQ SYMBOL'
    if __debug__: xtracer.trace("parser.p_lit_term_eq_term ENTER (lit)")
    p[0] = Literal(1,Atom(p[2],[symbol(p[1]),symbol(p[3])]))
    p[0].lineno = get_lineno(p,2)

def p_lit_term_tildaeq_term(p):
    'lit : SYMBOL TILDAEQ SYMBOL'
    if __debug__: xtracer.trace("parser.p_lit_term_tildaeq_term ENTER (lit)")
    p[0] = Literal(0,Atom(p[2],[symbol(p[1]),symbol(p[3])]))
    p[0].lineno = get_lineno(p,2)

def p_lit_tilda_atom(p):
    'lit : TILDA lit'
    if __debug__: xtracer.trace("parser.p_lit_tilda_atom ENTER (lit)")
    p[0] = ~p[2]
    p[0].lineno = get_lineno(p,1)

def p_relop_eq(p):
    'relop : EQ'
    if __debug__: xtracer.trace("parser.p_relop_eq ENTER (relop)")
    p[0] = p[1]

def p_relop_le(p):
    'relop : LE'
    if __debug__: xtracer.trace("parser.p_relop_le ENTER (relop)")
    p[0] = p[1]

def p_relop_lt(p):
    'relop : LT'
    if __debug__: xtracer.trace("parser.p_relop_lt ENTER (relop)")
    p[0] = p[1]

def p_relop_ge(p):
    'relop : GE'
    if __debug__: xtracer.trace("parser.p_relop_ge ENTER (relop)")
    p[0] = p[1]

def p_relop_gt(p):
    'relop : GT'
    if __debug__: xtracer.trace("parser.p_relop_gt ENTER (relop)")
    p[0] = p[1]

def p_relop_pto(p):
    'relop : PTO'
    if __debug__: xtracer.trace("parser.p_relop_pto ENTER (relop)")
    p[0] = p[1]

def p_infix_plus(p):
    'infix : PLUS'
    if __debug__: xtracer.trace("parser.p_infix_plus ENTER (infix)")
    p[0] = p[1]

def p_infix_minus(p):
    'infix : MINUS'
    if __debug__: xtracer.trace("parser.p_infix_minus ENTER (infix)")
    p[0] = p[1]

def p_infix_times(p):
    'infix : TIMES'
    if __debug__: xtracer.trace("parser.p_infix_times ENTER (infix)")
    p[0] = p[1]

def p_infix_div(p):
    'infix : DIV'
    if __debug__: xtracer.trace("parser.p_infix_div ENTER (infix)")
    p[0] = p[1]

# formulas are boolean combinations of terms and equalities between terms

def p_fmla_term(p):
    'fmla : term'
    if __debug__: xtracer.trace("parser.p_fmla_term ENTER (fmla)")
    p[0] = app_to_atom(p[1])

# prior to version 1.7, formulas can't be terms!

if iu.get_numeric_version() <= [1,6]:

    def p_fmla_term_relop_term(p):
        'fmla : term relop term'
        if __debug__: xtracer.trace("parser.p_fmla_term_relop_term ENTER (fmla)")
        p[0] = Atom(p[2],[p[1],p[3]])
        p[0].lineno = get_lineno(p,2)

    def p_fmla_term_tildaeq_term(p):
        'fmla : term TILDAEQ term'
        if __debug__: xtracer.trace("parser.p_fmla_term_tildaeq_term ENTER (fmla)")
        p[0] = Not(Atom('=',[p[1],p[3]]))
        p[0].lineno = get_lineno(p,2)

    def p_fmla_lparen_fmla_rparen(p):
        'fmla : LPAREN fmla RPAREN'
        if __debug__: xtracer.trace("parser.p_fmla_lparen_fmla_rparen ENTER (fmla)")
        p[0] = p[2]
        p[0].parenthesized=True

    def p_fmla_true(p):
        'fmla : TRUE'
        if __debug__: xtracer.trace("parser.p_fmla_true ENTER (fmla)")
        p[0] = And()
        p[0].lineno = get_lineno(p,1)

    def p_fmla_false(p):
        'fmla : FALSE'
        if __debug__: xtracer.trace("parser.p_fmla_false ENTER (fmla)")
        p[0] = Or()
        p[0].lineno = get_lineno(p,1)

    def p_fmla_not_fmla(p):
        'fmla : TILDA fmla'
        if __debug__: xtracer.trace("parser.p_fmla_not_fmla ENTER (fmla)")
        p[0] = Not(p[2])
        p[0].lineno = get_lineno(p,1)

    def p_fmla_fmla_and_fmla(p):
        'fmla : fmla AND fmla'
        if __debug__: xtracer.trace("parser.p_fmla_fmla_and_fmla ENTER (fmla)")
        # if isinstance(p[1],And):
        #     p[0] = p[1]
        #     p[0].args.append(p[3])
        # else:
        p[0] = And(p[1],p[3])
        p[0].lineno = get_lineno(p,2)

    def p_fmla_fmla_or_fmla(p):
        'fmla : fmla OR fmla'
        if __debug__: xtracer.trace("parser.p_fmla_fmla_or_fmla ENTER (fmla)")
        # if isinstance(p[1],Or):
        #     p[0] = p[1]
        #     p[0].args.append(p[3])
        # else:
        p[0] = Or(p[1],p[3])
        p[0].lineno = get_lineno(p,2)

    if not (iu.get_numeric_version() <= [1]):

        def p_fmla_fmla_arrow_fmla(p):
            'fmla : fmla ARROW fmla'
            if __debug__: xtracer.trace("parser.p_fmla_fmla_arrow_fmla ENTER (fmla)")
            p[0] = Implies(p[1],p[3])
            p[0].lineno = get_lineno(p,2)

    def p_fmla_fmla_iff_fmla(p):
        'fmla : fmla IFF fmla'
        if __debug__: xtracer.trace("parser.p_fmla_fmla_iff_fmla ENTER (fmla)")
        p[0] = Iff(p[1],p[3])
        p[0].lineno = get_lineno(p,2)

    if (iu.get_numeric_version() <= [1,6]):

        def p_fmla_forall_vars_dot_fmla(p):
            'fmla : FORALL simplevars DOT fmla'
            if __debug__: xtracer.trace("parser.p_fmla_forall_vars_dot_fmla ENTER (fmla)")
            p[0] = Forall(p[2],p[4])
            p[0].lineno = get_lineno(p,1)

        def p_fmla_exists_vars_dot_fmla(p):
            'fmla : EXISTS simplevars DOT fmla'
            if __debug__: xtracer.trace("parser.p_fmla_exists_vars_dot_fmla ENTER (fmla)")
            p[0] = Exists(p[2],p[4])
            p[0].lineno = get_lineno(p,1)

    else:

        def p_fmla_forall_simplevars_dot_fmla(p):
            'fmla : FORALL simplevars DOT fmla %prec SEMI'
            if __debug__: xtracer.trace("parser.p_fmla_forall_simplevars_dot_fmla ENTER (fmla)")
            p[0] = Forall(p[2],p[4])
            p[0].lineno = get_lineno(p,1)

        def p_fmla_exists_simplevars_dot_fmla(p):
            'fmla : EXISTS simplevars DOT fmla %prec SEMI'
            if __debug__: xtracer.trace("parser.p_fmla_exists_simplevars_dot_fmla ENTER (fmla)")
            p[0] = Exists(p[2],p[4])
            p[0].lineno = get_lineno(p,1)

        def p_fmla_forall_lp_vars_lp_fmla(p):
            'fmla : FORALL LPAREN vars RPAREN fmla'
            if __debug__: xtracer.trace("parser.p_fmla_forall_lp_vars_lp_fmla ENTER (fmla)")
            p[0] = Forall(p[3],p[5])
            p[0].lineno = get_lineno(p,1)

        def p_fmla_exists_lp_vars_lp_fmla(p):
            'fmla : EXISTS LPAREN vars RPAREN fmla'
            if __debug__: xtracer.trace("parser.p_fmla_exists_lp_vars_lp_fmla ENTER (fmla)")
            p[0] = Exists(p[3],p[5])
            p[0].lineno = get_lineno(p,1)

    def p_fmla_globally_fmla(p):
        'fmla : GLOBALLY fmla'
        if __debug__: xtracer.trace("parser.p_fmla_globally_fmla ENTER (fmla)")
        p[0] = Globally(p[2])
        p[0].lineno = get_lineno(p,1)

    def p_fmla_eventually_fmla(p):
        'fmla : EVENTUALLY fmla'
        if __debug__: xtracer.trace("parser.p_fmla_eventually_fmla ENTER (fmla)")
        p[0] = Eventually(p[2])
        p[0].lineno = get_lineno(p,1)

else:

    def p_term_term_EQ_term(p):
        'term : term EQ term'
        if __debug__: xtracer.trace("parser.p_term_term_EQ_term ENTER (term)")
        p[0] = Atom(p[2],[p[1],p[3]])
        p[0].lineno = get_lineno(p,2)

    def p_term_term_LE_term(p):
        'term : term LE term'
        if __debug__: xtracer.trace("parser.p_term_term_LE_term ENTER (term)")
        p[0] = Atom(p[2],[p[1],p[3]])
        p[0].lineno = get_lineno(p,2)

    def p_term_term_LT_term(p):
        'term : term LT term'
        if __debug__: xtracer.trace("parser.p_term_term_LT_term ENTER (term)")
        p[0] = Atom(p[2],[p[1],p[3]])
        p[0].lineno = get_lineno(p,2)

    def p_term_term_GE_term(p):
        'term : term GE term'
        if __debug__: xtracer.trace("parser.p_term_term_GE_term ENTER (term)")
        p[0] = Atom(p[2],[p[1],p[3]])
        p[0].lineno = get_lineno(p,2)

    def p_term_term_GT_term(p):
        'term : term GT term'
        if __debug__: xtracer.trace("parser.p_term_term_GT_term ENTER (term)")
        p[0] = Atom(p[2],[p[1],p[3]])
        p[0].lineno = get_lineno(p,2)

    def p_term_term_PTO_term(p):
        'term : term PTO term'
        if __debug__: xtracer.trace("parser.p_term_term_PTO_term ENTER (term)")
        p[0] = Atom(p[2],[p[1],p[3]])
        p[0].lineno = get_lineno(p,2)

    def p_term_term_tildaeq_term(p):
        'term : term TILDAEQ term'
        if __debug__: xtracer.trace("parser.p_term_term_tildaeq_term ENTER (term)")
        p[0] = Not(Atom('=',[p[1],p[3]]))
        p[0].lineno = get_lineno(p,2)

    def p_term_true(p):
        'term : TRUE'
        if __debug__: xtracer.trace("parser.p_term_true ENTER (term)")
        p[0] = And()
        p[0].lineno = get_lineno(p,1)

    def p_term_false(p):
        'term : FALSE'
        if __debug__: xtracer.trace("parser.p_term_false ENTER (term)")
        p[0] = Or()
        p[0].lineno = get_lineno(p,1)

    def p_term_not_term(p):
        'term : TILDA term'
        if __debug__: xtracer.trace("parser.p_term_not_term ENTER (term)")
        p[0] = Not(p[2])
        p[0].lineno = get_lineno(p,1)

    def p_term_term_and_term(p):
        'term : term AND term'
        if __debug__: xtracer.trace("parser.p_term_term_and_term ENTER (term)")
        if isinstance(p[1],And):
            p[0] = p[1]
            p[0].args.append(p[3])
        else:
            p[0] = And(p[1],p[3])
            p[0].lineno = get_lineno(p,2)

    def p_term_term_or_term(p):
        'term : term OR term'
        if __debug__: xtracer.trace("parser.p_term_term_or_term ENTER (term)")
        if isinstance(p[1],Or):
            p[0] = p[1]
            p[0].args.append(p[3])
        else:
            p[0] = Or(p[1],p[3])
            p[0].lineno = get_lineno(p,2)

    def p_term_term_arrow_term(p):
        'term : term ARROW term'
        if __debug__: xtracer.trace("parser.p_term_term_arrow_term ENTER (term)")
        p[0] = Implies(p[1],p[3])
        p[0].lineno = get_lineno(p,2)
        if isinstance(p[1],Implies) and not hasattr(p[1],"parenthesized"):
            iu.warn(p[0],"It is recommended parenthesize nested -> operators to avoid ambiguity.")

    def p_term_term_iff_term(p):
        'term : term IFF term'
        if __debug__: xtracer.trace("parser.p_term_term_iff_term ENTER (term)")
        p[0] = Iff(p[1],p[3])
        p[0].lineno = get_lineno(p,2)

    def p_term_forall_simplevars_dot_term(p):
        'term : FORALL simplevars DOT term %prec SEMI'
        if __debug__: xtracer.trace("parser.p_term_forall_simplevars_dot_term ENTER (term)")
        p[0] = Forall(p[2],p[4])
        p[0].lineno = get_lineno(p,1)

    def p_term_exists_simplevars_dot_term(p):
        'term : EXISTS simplevars DOT term %prec SEMI'
        if __debug__: xtracer.trace("parser.p_term_exists_simplevars_dot_term ENTER (term)")
        p[0] = Exists(p[2],p[4])
        p[0].lineno = get_lineno(p,1)

    def p_term_forall_lp_vars_lp_term(p):
        'term : FORALL LPAREN vars RPAREN term'
        if __debug__: xtracer.trace("parser.p_term_forall_lp_vars_lp_term ENTER (term)")
        p[0] = Forall(p[3],p[5])
        p[0].lineno = get_lineno(p,1)

    def p_term_exists_lp_vars_lp_term(p):
        'term : EXISTS LPAREN vars RPAREN term'
        if __debug__: xtracer.trace("parser.p_term_exists_lp_vars_lp_term ENTER (term)")
        p[0] = Exists(p[3],p[5])
        p[0].lineno = get_lineno(p,1)

    def p_term_globally_term(p):
        'term : GLOBALLY term'
        if __debug__: xtracer.trace("parser.p_term_globally_term ENTER (term)")
        p[0] = Globally(p[2])
        p[0].lineno = get_lineno(p,1)

    def p_term_eventually_term(p):
        'term : EVENTUALLY term'
        if __debug__: xtracer.trace("parser.p_term_eventually_term ENTER (term)")
        p[0] = Eventually(p[2])
        p[0].lineno = get_lineno(p,1)

    def p_term_term_whennext_term(p):
        'term : term WHENNEXT term'
        if __debug__: xtracer.trace("parser.p_term_term_whennext_term ENTER (term)")
        p[0] = WhenOperator('next',p[1],p[3])
        p[0].lineno = get_lineno(p,2)

    def p_term_term_whenprev_term(p):
        'term : term WHENPREV term'
        if __debug__: xtracer.trace("parser.p_term_term_whenprev_term ENTER (term)")
        p[0] = WhenOperator('prev',p[1],p[3])
        p[0].lineno = get_lineno(p,2)

    def p_term_term_whenfirst_term(p):
        'term : term WHENFIRST term'
        if __debug__: xtracer.trace("parser.p_term_term_whenfirst_term ENTER (term)")
        p[0] = WhenOperator('first',p[1],p[3])
        p[0].lineno = get_lineno(p,2)

    def p_term_term_whenlast_term(p):
        'term : term WHENLAST term'
        if __debug__: xtracer.trace("parser.p_term_term_whenlast_term ENTER (term)")
        p[0] = WhenOperator('last',p[1],p[3])
        p[0].lineno = get_lineno(p,2)

        
def p_term_namedbinder_vars_dot_term(p):
    'term : LPAREN DOLLAR SYMBOL simplevars DOT fmla RPAREN LPAREN terms RPAREN'
    if __debug__: xtracer.trace("parser.p_term_namedbinder_vars_dot_term ENTER (term)")
    x = NamedBinder(p[3], p[4],p[6])
    x.lineno = get_lineno(p,2)
    p[0] = App(x, p[9])
    p[0].lineno = get_lineno(p,2)

def p_term_namedbinder_dot_fmla(p):
    'term : DOLLAR SYMBOL DOT fmla %prec SEMI'
    if __debug__: xtracer.trace("parser.p_term_namedbinder_dot_fmla ENTER (term)")
    p[0] = NamedBinder(p[2], [],p[4])
    p[0].lineno = get_lineno(p,1)

def p_term_namedbinder_dollar_fmla(p):
    'term : DOLLAR SYMBOL DOLLAR fmla %prec SEMI'
    if __debug__: xtracer.trace("parser.p_term_namedbinder_dollar_fmla ENTER (term)")
    p[0] = NamedBinder(p[2], [],p[4])
    p[0].lineno = get_lineno(p,1)

if not (iu.get_numeric_version() <= [1,6]):
    def p_fmla_fmla_isa_atype(p):
        'term : term ISA atype'
        if __debug__: xtracer.trace("parser.p_fmla_fmla_isa_atype ENTER (term)")
        tp = Atom(p[3],[])
        tp.lineno = get_lineno(p,2)
        p[0] = Isa(p[1],tp)
        p[0].lineno = get_lineno(p,2)
    
# TODO: should the above rules create formulas also or only for terms
