#
# Copyright (c) Microsoft Corporation. All Rights Reserved.
#

from . import ivy_module as im
from . import xtracer
from . import ivy_actions as ia
from . import ivy_logic as il
from . import ivy_transrel as tr
from . import ivy_logic_utils as ilu
from . import ivy_utils as iu
from . import ivy_art as art
from . import ivy_interp as itp
from . import ivy_theory as thy
from . import ivy_ast
from . import ivy_proof
from . import ivy_trace
from . import ivy_solver as slvr

from .z3 import Bool, simplify, is_func_decl, DatatypeSortRef, Z3_ARRAY_SORT, Z3_UNINTERPRETED_SORT
import tempfile
import subprocess
from collections import defaultdict
import itertools
import sys
import os

logfile = None
verbose = False

def checked(thing):
    return ia.checked_assert.value in ["",thing.lineno]

def action_to_tr(mod,action,method,bgt=None):
    if bgt is None:
        bgt = mod.background_theory()
    if method=="fsmc":  # if finite-state, unroll the loops
        with ia.UnrollContext(im.module.sort_card):
            if __debug__: xtracer.trace("vmt.ActionToTR calling GetUpdate (fsmc) type=%s" % type(action).__name__)
            upd = action.update(im.module,None)
    else:
        if __debug__: xtracer.trace("vmt.ActionToTR calling GetUpdate type=%s" % type(action).__name__)
        upd = action.update(im.module,None)
    stvars,trans,error = tr.add_post_axioms(upd,bgt)
    trans = ilu.and_clauses(trans,ilu.Clauses(defs=bgt.defs))
    defsyms = set(x.defines() for x in bgt.defs)
    rn = dict((tr.new(sym),tr.new(sym).prefix('__')) for sym in defsyms)
    trans = ilu.rename_clauses(trans,rn)
    error = ilu.rename_clauses(error,rn)
    stvars = [x for x in stvars if x not in defsyms]  # Remove symbols with state-dependent definitions
    return (stvars,trans,error)

def add_err_flag(action,erf,errconds):
    if isinstance(action,ia.AssertAction):
        if False and checked(action):
            if verbose:
                print("{}Model checking guarantee".format(action.lineno))
            errcond = ilu.dual_formula(il.drop_universals(action.formula))
            res = ia.AssignAction(erf,il.Or(erf,errcond))
            errconds.append(errcond)
            res.lineno = iu.nowhere()
            return res
        if isinstance(action,ia.SubgoalAction):
#            print "skipping subgoal at line {}".format(action.lineno)
            return ia.Sequence()
    if isinstance(action,ia.AssumeAction) or isinstance(action,ia.AssertAction):
        if isinstance(action,ia.AssertAction):
            if verbose:
                print("assuming assertion at line {}".format(action.lineno))
        res = ia.AssumeAction(il.Or(erf,action.formula))
        res.lineno = iu.nowhere()
        return res
    if isinstance(action,(ia.Sequence,ia.ChoiceAction,ia.EnvAction,ia.BindOldsAction)):
        return action.clone([add_err_flag(a,erf,errconds) for a in action.args])
    if isinstance(action,ia.IfAction):
        return action.clone([action.args[0]] + [add_err_flag(a,erf,errconds) for a in action.args[1:]])
    if isinstance(action,ia.LocalAction):
        return action.clone(action.args[:-1] + [add_err_flag(action.args[-1],erf,errconds)])
    return action

def add_err_flag_mod(mod,erf,errconds):
    for actname in list(mod.actions):
        action = mod.actions[actname]
        new_action = add_err_flag(action,erf,errconds)
        mod.actions[actname] = action.copy_formals(new_action)


# Take a function sort and return a corresponding array sort.
def create_array_sort(sort):
    def aname(i):
        if i == len(sort.dom):
            return (sort.rng.name,[sort.rng])
        sname,ssorts = aname(i+1)
        name = 'arr[' + sort.dom[i].name + '][' + sname + ']'
        if name not in il.sig.sorts:
            asort = il.UninterpretedSort(name)
            il.add_sort(asort)
            il.sig.interp[name] = name
        return (name,[il.sig.sorts[name]]+ssorts)
    return aname(0)

def lower_array_state(ast, arrays):
    """Lower reads after the symbolic action has captured each indexed update."""
    args = [lower_array_state(arg, arrays) for arg in ast.args]
    if args and il.is_app(ast) and not il.is_named_binder(ast):
        sorts = arrays.get(ast.rep)
        if sorts:
            value = il.Symbol(ast.rep.name, sorts[0])
            for i, arg in enumerate(args):
                select = il.Symbol('arrsel', il.FunctionSort(sorts[i], arg.sort, sorts[i+1]))
                value = select(value, arg)
            return value
    return ast.clone(args)


def has_assert(action):
    for x in action.iter_subactions():
        if isinstance(x,ia.AssertAction):
            return True
    return False

def normalize_prop(prop_z3_fm):
    if prop_z3_fm.is_forall():
        prop_sexpr = prop_z3_fm.body().sexpr()
        for i in range(prop_z3_fm.num_vars()):
            var_name = prop_z3_fm.var_name(i).split(":")[0]
            prop_sexpr = prop_sexpr.replace(f"(:var {i})", var_name)
        return prop_sexpr
    # return prop_z3_fm.sexpr()
    raise Exception("Don't support non-universally quantified properties")


def herbrandize_property_vars(prop_z3_fm):
    if prop_z3_fm.is_forall():
        var_strs = []
        trans = []
        for i in range(prop_z3_fm.num_vars()):
            var_name = prop_z3_fm.var_name(i).split(":")[0]
            var_sort = prop_z3_fm.var_sort(i).name()
            vmt_str = f"(declare-fun {var_name} () {var_sort})"
            vmt_next = f"(declare-fun new_{var_name} () {var_sort})"
            define_fun_str = var_name.replace("|", "").replace(":", "")
            define = f"(define-fun .{define_fun_str} () {var_sort} (! {var_name} :next new_{var_name}))"
            var_strs.append(vmt_str)
            var_strs.append(vmt_next)
            var_strs.append(define)
            trans.append(f"(= {var_name} new_{var_name})")
        return var_strs, trans
    raise Exception("Don't support non-universally quantified properties")

def get_dom_and_rng(sort_name):
    dom, rng = sort_name.split("-")[1:]
    if rng == "bool":
        rng = "Bool"
    return dom, rng

def frame_clauses(all_vars, mod_vars, clauses):
    mod_set = set(mod_vars)
    frames = []
    for symbol in all_vars:
        if symbol in mod_set:
            continue
        if il.is_function_sort(symbol.sort):
            args = [il.Variable(f'VMT{i}',sort) for i,sort in enumerate(symbol.sort.dom)]
            frames.append(il.forall(args,il.Equals(symbol(*args),tr.new(symbol)(*args))))
        else:
            frames.append(il.Equals(symbol,tr.new(symbol)))
    return ilu.and_clauses(ilu.Clauses(frames), clauses)

def check_isolate(method="mc",tagged_dfns=[],use_array_encoding=True):
    mod = im.module

    # Use the error flag construction to turn assertion checks into
    # an invariant check.

    erf = il.Symbol('err_flag',il.find_sort('bool'))
    errconds = []
    has_erf = any(has_assert(mod.actions[x]) for x in mod.actions)
    if has_erf:
        add_err_flag_mod(mod,erf,errconds)


    # We can't currently handle assertions in the initializers

    for (name,init) in mod.initializers:
        if has_assert(init):
            raise iu.IvyError(init,'VMT cannot handle assertions in initializers')

    # Build a single initializer action
    init_action = ia.Sequence(*[act for (nm,act) in mod.initializers])


    # get the invariant to be proved, replacing free variables with
    # skolems. First, we apply any proof tactics.

    pc = ivy_proof.ProofChecker(mod.labeled_axioms,mod.definitions,mod.schemata)
    pmap = dict((lf.id,p) for lf,p in mod.proofs)
    conjs = []
    for lf in mod.labeled_conjs:
        if not checked(lf):
#            print 'skipping {}'.format(lf)
            continue
        if verbose:
            print("{}Model checking invariant".format(lf.lineno))
        if lf.id in pmap:
            proof = pmap[lf.id]
            subgoals = pc.admit_proposition(lf,proof)
            conjs.extend(subgoals)
        else:
            conjs.append(lf)

    bgt = mod.background_theory()

    modified = {sym for action in [init_action] + list(mod.actions.values())
                for subaction in action.iter_subactions() for sym in subaction.modifies()}

    # Combine all actions nto a single action

    actions = [(x,mod.actions[x].add_label(x)) for x in sorted(mod.public_actions)]
#    action = ia.EnvAction(*ext_acts)
    if has_erf:
        actions = [(x,ia.Sequence(ia.AssignAction(erf,il.Or()).set_lineno(iu.nowhere()),act))
                   for x,act in actions]

    # Seperate out the axioms that do not have modified symbols

    have_mod = [any(y in modified for y in ilu.used_symbols_ast(x)) for x in bgt.fmlas]
    axioms = [x for x,hm in zip(bgt.fmlas,have_mod) if not hm]
    bgt = ilu.Clauses([x for x,hm in zip(bgt.fmlas,have_mod) if hm],defs=bgt.defs)

    # Convert the global action and initializer to logic. Notice we
    # use 'action_to_state' to turn the initializer TR into a state
    # predicate (i.e., it's strongest post).

    actupds = [action_to_tr(mod,action,method,bgt) for x,action in actions]
    # for a in actupds:
        # print ('a[0]: {}'.format(a[0]))
        # print ('a[1]: {}'.format(a[1]))
    init_update = action_to_tr(mod,init_action,method,bgt)
    istvars = init_update[0]
    init = tr.forward_image(mod.init_cond,bgt,init_update)
    stvars = list(iu.unique(sym for a in actupds for sym in a[0]))
    stvars = list(iu.unique(stvars+istvars))
    transs = [ilu.clauses_to_formula(frame_clauses(stvars,a[0],a[1])) for a in actupds]
    init = ilu.clauses_to_formula(init)
    if use_array_encoding:
        arrays = {}
        for symbol in stvars:
            if (il.is_function_sort(symbol.sort) and not il.is_interpreted_symbol(symbol)
                and symbol.name not in mod.destructor_sorts):
                _, sorts = create_array_sort(symbol.sort)
                arrays[symbol] = arrays[tr.new(symbol)] = sorts
        init = lower_array_state(init, arrays)
        transs = [lower_array_state(formula, arrays) for formula in transs]
        conjs = [conj.clone([conj.label,lower_array_state(conj.formula,arrays)]) for conj in conjs]
    trans = il.Or(*transs)

    sorts = []

    vmt_var_defs = []

    def add_sort(sort):
        if sort not in sorts:
            if sort.kind() == Z3_ARRAY_SORT:
                add_sort(sort.range())
                add_sort(sort.domain())
            else:
                sorts.append(sort)


    declared_symbols = set()
    enum_symbols = set()
    used_names = {str(definition.defines()) for _, definition in tagged_dfns}
    tagged_bodies = [definition.rhs() for _,definition in tagged_dfns]
    refed_syms = list(ilu.used_symbols_asts([init,trans]+[lf.formula for lf in conjs]+axioms+tagged_bodies))
    extra_refed_syms = set()
    for x in refed_syms:
        if tr.is_new(x):
            extra_refed_syms.add(tr.new_of(x))
    refed_syms = iu.unique(refed_syms+list(extra_refed_syms))
    for sym in refed_syms:
        sym = il.normalize_symbol(sym)
        if il.is_enumerated_sort(sym.sort):
            if sym in sym.sort.constructors:
                enum_symbols.add(sym)
                add_sort(sym.sort.to_z3())
                continue
        if sym not in declared_symbols and slvr.solver_name(sym) is not None and sym.name != "cast":
            declared_symbols.add(sym)
            decl = slvr.symbol_to_z3_full(sym)
            if il.is_constant(sym):
                if is_func_decl(decl):
                    for i in range(decl.arity()):
                        add_sort(decl.domain(i))
                    add_sort(decl.range())
                    if not il.is_interpreted_symbol(sym):
                        decl_sexpr = decl.sexpr()
                        used_names.add(decl.name())
                        vmt_var_defs.append(decl_sexpr)
                        # decl_sort = decl.range().sexpr()
                        # decl_dom = ' '.join(decl.domain(i).sexpr() for i in range(decl.arity()))
                        # vmt_var_defs.append(f"(declare-fun {decl_sexpr} ({decl_dom}) {decl_sort})")
                else:
                    add_sort(decl.sort())
                    if not il.is_interpreted_symbol(sym):
                        decl_sexpr = decl.sexpr()
                        decl_sort = decl.sort().sexpr()
                        used_names.add(decl.decl().name())
                        vmt_var_defs.append(f'(declare-fun {decl_sexpr} () {decl_sort})')
            else:
                vmt_var_defs.append(decl.sexpr())

    vmt_sort_defs = []
    vmt_func_defs = []
    # must declare sorts that are not built-in
    for sort in sorts:
        sort_name = sort.name()
        if sort.kind() != Z3_UNINTERPRETED_SORT and not isinstance(sort,DatatypeSortRef):
            continue
        if sort_name.startswith("Arr"):
            dom, rng = get_dom_and_rng(sort_name)
            read = f"(declare-fun Read{sort_name} ({sort_name} {dom}) {rng})"
            write = f"(declare-fun Write{sort_name} ({sort_name} {dom} {rng}) {sort_name})"
            const = f"(declare-fun Const{sort_name} ({rng}) {sort_name})"
            vmt_func_defs.append(read)
            vmt_func_defs.append(write)
            vmt_func_defs.append(const)
        if isinstance(sort,DatatypeSortRef):
            used_names.update(sort.constructor(i).name() for i in range(sort.num_constructors()))
            sort_str = f"(declare-datatypes () (({sort_name} {' '.join(str(sort.constructor(i)) for i in range(sort.num_constructors()))})))"
        else:
            sort_str = f"(declare-sort {sort.sexpr()} 0)"
        vmt_sort_defs.append(sort_str)

    def fresh_name(base):
        name = base
        suffix = 1
        while name in used_names:
            name = f'{base}${suffix}'
            suffix += 1
        used_names.add(name)
        return Bool(name).sexpr()

    all_used = ilu.used_symbols_asts([init,trans])
    immutable = [sym for sym in all_used if not il.is_function_sort(sym.sort)
                 and not sym.is_skolem() and not sym.is_numeral() and not tr.is_new(sym)
                 and "fml:" not in sym.name and "loc:" not in sym.name and 'prm:' not in sym.name and tr.new(sym) not in all_used
                 and all(sym not in upd[0] for upd in actupds)]
    background_symbols = set(ilu.used_symbols_asts(axioms))
    background_symbols.update(enum_symbols)
    immutable = [sym for sym in immutable if not sym in background_symbols]
    # print ('immutable: {}'.format(immutable))

    for sym in ilu.used_symbols_asts([init,trans]):
        if sym.is_numeral():
            continue
        if tr.is_new(sym):
            next_sym = slvr.symbol_to_z3_full(sym)
            cur_sym = slvr.symbol_to_z3_full(tr.new_of(sym))
            next_str = next_sym.sexpr()
            cur_str = cur_sym.sexpr()
            # print ('foo: {},{}'.format(cur_sym,sym))
            if is_func_decl(cur_sym):
                decl = cur_sym
                decl_sort = decl.range().sexpr()
                names = [f'arg{i}' for i in range(decl.arity())]
                params = ' '.join(f'({name} {decl.domain(i).sexpr()})' for i,name in enumerate(names))
                cur_name = cur_str.removeprefix('(declare-fun ').split(' (',1)[0]
                next_name = next_str.removeprefix('(declare-fun ').split(' (',1)[0]
                define_fun_str = fresh_name('.' + cur_name.replace("|", "").replace(":", ""))
                current = f"({cur_name} {' '.join(names)})"
                full_str = f"(define-fun {define_fun_str} ({params}) {decl_sort} (! {current} :next {next_name}))"
            else:
                sort = cur_sym.sort().sexpr()
                define_fun_str = fresh_name('.' + cur_str.replace("|", "").replace(":", ""))
                full_str = f"(define-fun {define_fun_str} () {sort} (! {cur_str} :next {next_str}))"
            vmt_var_defs.append(full_str)
        elif "fml:" in sym.name or "loc:" in sym.name or "prm:" in sym.name or sym.is_skolem() or sym in immutable:
            next_sym_str = f"new_{sym.name.replace(':', '')}"
            # next_sym_str = slvr.symbol_to_z3_full(tr.new(sym)).sexpr()
            cur_sym_str = slvr.symbol_to_z3_full(sym).sexpr()
            # cur_sym_str = f"|{sym.name}|"
            sort = slvr.symbol_to_z3(sym).sort().sexpr()
            next_fml_def = f"(declare-fun {next_sym_str} () {sort})"
            vmt_var_defs.append(next_fml_def)
            used_names.add(next_sym_str)
            define_fun_str = fresh_name('.' + cur_sym_str.replace("|", "").replace(":", ""))
            full_str = f"(define-fun {define_fun_str} () {sort} (! {cur_sym_str} :next {next_sym_str}))"
            vmt_var_defs.append(full_str)

    frame = [il.Equals(x,tr.new(x)) for x in immutable]
    # print ('frame: {}'.format(frame))

    props = []
    herbrand_trans = [simplify(slvr.formula_to_z3(x)).sexpr() for x in frame] # List[str]
    for i, lf in enumerate(conjs):
        # var_defs, herb_trans = herbrandize_property_vars(slvr.formula_to_z3(lf.formula))
        # herbrand_trans.extend(herb_trans)
        # vmt_var_defs.extend(var_defs)
        # prop_formula = normalize_prop(slvr.formula_to_z3(lf.formula))
        prop_formula = slvr.formula_to_z3(lf.formula).sexpr()
        prop_name = fresh_name(f'property_{i}')
        props.append(f'(define-fun {prop_name} () Bool (! {prop_formula} :invar-property {i}))')

    init_formula_str = simplify(slvr.formula_to_z3(init)).sexpr()
    trans_formula_str = simplify(slvr.formula_to_z3(trans)).sexpr()
    vmt_formula_defs = []
    init_name = fresh_name('init')
    init_vmt = f"(define-fun {init_name} () Bool (! {init_formula_str} :init true))"
    vmt_formula_defs.append(init_vmt)

    transition_name = fresh_name('trans')
    action_names = [fresh_name(name.split(':')[-1]) for name, _ in actions]
    for action_name in action_names:
        vmt_var_defs.append(f'(declare-fun {action_name} () Bool)')
        metadata_name = fresh_name('.' + action_name.strip('|'))
        vmt_var_defs.append(f'(define-fun {metadata_name} () Bool (! {action_name} :action 0))')

    # print ('foo : {}'.format(transs[-1]))

    trans_vmt = get_trans_str(transs, action_names, herbrand_trans, transition_name)
    vmt_formula_defs.append(trans_vmt)
    vmt_formula_defs.extend(props)

    vmt_axiom_defs = []
    for axiom in axioms:
        vmt_axiom_defs.append('(assert {})'.format(simplify(slvr.formula_to_z3(axiom)).sexpr()))

    vmt_dfns = []
    for tag,dfn in tagged_dfns:
        #print(f'dfn: {type(dfn)}')
        args = " ".join(f"({x.sexpr()} {x.sort().sexpr()})"
                        for y in dfn.lhs().args for x in [slvr.term_to_z3(y)])
        rng = slvr.term_to_z3(dfn.lhs()).sort().sexpr()
        rhs = slvr.term_to_z3(dfn.rhs()).sexpr()
        thing = f"(define-fun {dfn.defines()} ({args}) {rng} (! {rhs} :{tag} true))"
        vmt_dfns.append(thing)

    vmt_file_name = mod.name + '.vmt'
    with open(vmt_file_name, "w") as f:
        vmt_str = vmt_sort_defs + vmt_func_defs + vmt_var_defs + vmt_formula_defs + vmt_axiom_defs + vmt_dfns
        f.write("\n".join(vmt_str))

    print('output written to '+vmt_file_name)
    exit(0)


def get_trans_str(trans_fm_list, action_names, herbrand_trans, transition_name):
    vmt_list = []
    for action_name, trans_fm in zip(action_names, trans_fm_list):
        trans_str = simplify(slvr.formula_to_z3(trans_fm)).sexpr()
        full_str = f"(=> {action_name} {trans_str})"
        vmt_list.append(full_str)
    vmt_list.append(f"(or {' '.join(action_names)})" if action_names else 'false')
    for i,x in enumerate(action_names):
        for j,y in enumerate(action_names[i+1:]):
            vmt_list.append(f"(or (not {x}) (not {y}))")
    for ht in herbrand_trans:
        vmt_list.append(ht)
    vmt_str = '\n'.join(vmt_list)
    return f"(define-fun {transition_name} () Bool (! (and {vmt_str}) :trans true))"

