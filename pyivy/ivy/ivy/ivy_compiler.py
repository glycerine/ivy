#
# Copyright (c) Microsoft Corporation. All Rights Reserved.
#
from .ivy_art import *
import pickle
from .ivy_interp import Interp, eval_state_facts
from functools import partial
from .ivy_concept_space import *
from .ivy_parser import parse,ConstantDecl,ActionDef,Ivy,inst_mod
from .ivy_actions import DerivedUpdate, NamedUpdate, type_check_action, type_check, SymbolList, UpdatePattern, ActionContext, LocalAction, AssignAction, CallAction, Sequence, IfAction, WhileAction, AssertAction, AssumeAction, NativeAction, ChoiceAction, CrashAction, ThunkAction, DebugAction, has_code
from .ivy_utils import IvyError
from . import ivy_logic
from . import ivy_dafny_compiler as dc
from . import ivy_solver as slv
from . import ivy_logic_utils as lu
import string
from . import ivy_ast
from . import ivy_utils as iu
from . import ivy_actions as ia
from . import ivy_alpha
from . import ivy_module as im
from . import ivy_theory as ith
from . import ivy_isolate as iso
from . import ivy_printer
from . import ivy_proof as ip
from . import xtracer
from collections import defaultdict
from tarjan import tarjan
import importlib

opt_mutax = iu.BooleanParameter("mutax",False)

# --- Sig conformance auditing via Merkle-chained Canon ---
# Matches Go's Compiler.SigCheck(). Zero cost with python -O.

from . import logic as lg

sig_merkle = xtracer.MerkleState()

def sort_name(s):
    """Match Go's il.SortName() exactly."""
    if hasattr(s, 'name'):
        return s.name
    if isinstance(s, lg.BooleanSort):
        return 'bool'
    if isinstance(s, lg.FunctionSort):
        return str(s)
    return str(s)

def sig_canon(sig):
    """Canonical s-expression of a Sig, matching Go's Sig.Canon()."""
    sort_names = sorted(sig.sorts.keys())
    sym_parts = sorted("%s:%s" % (name, sort_name(sym.sort)) for name, sym in sig.symbols.items())
    return "(sig sorts:[%s] symbols:[%s])" % (" ".join(sort_names), " ".join(sym_parts))

def sig_check(label):
    """Emit Merkle-chained HASH trace of current ivy_logic.sig + module state."""
    sig_c = sig_canon(ivy_logic.sig)
    mod_c = ""
    if im.module is not None:
        mod_c = im.module.canon()
    combined = sig_c + mod_c
    leaf, root = sig_merkle.add_leaf(combined)
    #if __debug__: xtracer.trace("compiler.SigCheck@%s HASH leaf=%s root=%s canon=%s" % (label, leaf, root, combined))
    # roots diverging, should track but go forward instead for now.
    if __debug__: xtracer.trace("compiler.SigCheck@%s HASH leaf=%s canon=%s" % (label, leaf, combined))

class IvyDeclInterp(object):
    def __call__(self,ivy):
        for decl in ivy.decls:
            with ASTContext(decl):
                n = decl.name()
#                print "decl: {} : {}".format(n,decl.lineno if hasattr(decl,'lineno') else None)
                if n == 'assert': n = '_assert' # reserved in python!
                #if __debug__: xtracer.trace("compiler.%s.dispatch name=%s\n pyDeclClass=%s hasMethod=%s" % (type(self).__name__, n, type(decl).__name__, hasattr(self, n)))
                if __debug__: xtracer.trace("compiler.%s.dispatch name=%s" % (type(self).__name__, n))
                if hasattr(self,n):
                    for x in decl.args:
                        getattr(self,n)(x)

from .ivy_ast import ASTContext

# ast compilation


ivy_ast.Variable.get_sort = lambda self: ivy_logic.find_sort(resolve_alias(self.sort.rep))

def thing(self):
    if __debug__: xtracer.trace("compiler.Thing ENTER type=%s" % type(self).__name__)
    if __debug__: xtracer.trace("compiler.CompileNode ENTER type=%s" % type(self).__name__)
    with ASTContext(self):
        result = self.cmpl()
        if __debug__: xtracer.trace("compiler.Thing return type=%s" % type(self).__name__)
        return result

# "roots" are AST objects that bind variables, such as assignment, assume, etc. 
# some roots have symbols as args instead of terms (e.g., field assign)
def compile_root_args(self):
    if __debug__: xtracer.trace("compiler.CompileRootArgs ENTER")
    result = [(find_symbol(a) if isinstance(a,str) else a.compile()) for a in self.args]
    if __debug__: xtracer.trace("compiler.CompileRootArgs return")
    return result

def other_thing(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=default type=%s" % type(self).__name__)
    if __debug__: xtracer.trace("compiler.OtherThing ENTER type=%s" % type(self).__name__)
    # we have to do sort inference on roots
    if hasattr(self,'sort_infer_root'):
        with top_sort_as_default():
            res = self.clone(compile_root_args(self))
        result = sort_infer(res)
        if __debug__: xtracer.trace("compiler.OtherThing return type=%s sort_infer_root=True" % type(self).__name__)
        return result
    else:
        result = self.clone([a.compile() for a in self.args])
        if __debug__: xtracer.trace("compiler.OtherThing return type=%s sort_infer_root=False" % type(self).__name__)
        return result

ivy_logic.AST.compile = ivy_ast.AST.compile = thing
ivy_logic.AST.cmpl = ivy_ast.AST.cmpl = other_thing

op_pairs = [
    (ivy_ast.And,ivy_logic.And),
    (ivy_ast.Or,ivy_logic.Or),
    (ivy_ast.Not,ivy_logic.Not),
    (ivy_ast.Globally,lambda x: ivy_logic.Globally("",x)),
    (ivy_ast.Eventually,lambda x: ivy_logic.Eventually("",x)),
    (ivy_ast.And,ivy_logic.And),
    (ivy_ast.Definition,ivy_logic.Definition),
    (ivy_ast.Implies,ivy_logic.Implies),
    (ivy_ast.Iff,ivy_logic.Iff),
    (ivy_ast.Ite,ivy_logic.Ite),
]

def compile_args(self):
    if __debug__: xtracer.trace("compiler.CompileArgs ENTER")
    with ReturnContext(None):
        result = [a.compile() for a in self.args]
        if __debug__: xtracer.trace("compiler.CompileArgs return")
        return result

def _make_op_cmpl(name, tc):
    def cmpl(self):
        if __debug__: xtracer.trace("compiler.CompileNode return case=%s" % name)
        if __debug__: xtracer.trace("compiler.Compile%s ENTER" % name)
        return tc(*compile_args(self))
    return cmpl

for fc, tc in op_pairs:
    fc.cmpl = _make_op_cmpl(fc.__name__, tc)

def _when_operator_cmpl(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=WhenOperator")
    if __debug__: xtracer.trace("compiler.CompileWhenOperator ENTER")
    return ivy_logic.WhenOperator(self.name, *compile_args(self))

ivy_ast.WhenOperator.cmpl = _when_operator_cmpl

class Context(object):
    def __enter__(self):
        self.old_context = globals().get(self.name,None)
        globals()[self.name] = self
        return self
    def __exit__(self,exc_type, exc_val, exc_tb):
        globals()[self.name] = self.old_context
        return False # don't block any exceptions

class ReturnContext(Context):
    def __init__(self,values,lineno=None):
        self.values = values
        self.lineno = lineno
        self.name = 'return_context'
    

class ExprContext(Context):
    """ Context Manager for compiling an expression. """
    def __init__(self,code=None,local_syms=None,lineno=None):
        self.code = code if code != None else []
        self.local_syms = local_syms if local_syms != None else []
        self.lineno = lineno
        self.name = 'expr_context'
    def extract(self):
        if __debug__: xtracer.trace("compiler.ExprContext.Extract ENTER")
        for c in self.code:
            c.lineno = self.lineno
        if len(self.code) == 1:
            return self.code[0]
        res = LocalAction(*(self.local_syms + [Sequence(*self.code)]),caller="compiler.compile_expression")
        res.lineno = self.lineno
        return res


class TopContext(Context):
    """ Context Manager for compiling an expression. """
    def __init__(self,actions):
        self.actions = actions
        self.name = 'top_context'

class EmptyVariableContext(object):
    def __init__(self):
        self.map = dict()

class VariableContext(Context):
    def __init__(self,vs):
        self.map = dict(iter(variable_context.map.items()))
        for v in vs:
            self.map[v.name] = v.sort
        self.name = 'variable_context'

return_context = None
expr_context = None
top_context = None
variable_context = EmptyVariableContext()

def pull_args(args,num,sym,top):
    if len(args) < num:
        raise IvyError(None,'not enough arguments to {}'.format(sym))
    if top and len(args) > num:
        raise IvyError(None,'too many arguments to {}'.format(sym))
    res = args[:num]
    del args[:num]
    return res


class cfrfail(Exception):
    def __init__(self,symbol_name):
        self.symbol_name = symbol_name

def compile_field_reference_rec(symbol_name,args,top=False,old=False):
    if __debug__: xtracer.trace("compiler.compile_field_reference_rec ENTER name=%s" % symbol_name)
    sym = ivy_logic.find_polymorphic_symbol(symbol_name,throw=False)
    if sym is not None:
        if __debug__: xtracer.trace("compiler.compile_field_reference_rec found name=%s sort=%s" % (symbol_name, sym.sort))
    else:
        if __debug__: xtracer.trace("compiler.compile_field_reference_rec not_found name=%s" % symbol_name)
    if sym is None:
        parent_name,child_name = iu.parent_child_name(symbol_name)
        if parent_name == 'this':
            raise cfrfail(symbol_name)
        try:
            with ReturnContext(None):
                base = compile_field_reference_rec(parent_name,args,old=old)
        except cfrfail as err:
            raise cfrfail(symbol_name if err.symbol_name in im.module.hierarchy else err.symbol_name)
        sort = base.sort
        # tricky: we first look for the method as a child of the sort.
        # if not found, we look for a sibling of the sort
        destr_name = iu.compose_names(sort.name,child_name)
        if __debug__: xtracer.trace("compiler.compile_field_reference_rec destrName=%s baseSort=%s" % (destr_name, sort.name))
        if top_context:
            _in_sig = destr_name in ivy_logic.sig.symbols
            _in_act = destr_name in top_context.actions
            if __debug__: xtracer.trace("compiler.compile_field_reference_rec destr_check name=%s inSig=%s inAct=%s" % (destr_name, str(_in_sig), str(_in_act)))
        if top_context and destr_name not in ivy_logic.sig.symbols and destr_name not in top_context.actions:
            sort_parent,sort_child = iu.parent_child_name(sort.name)
            destr_name = iu.compose_names(sort_parent,child_name)
            if __debug__: xtracer.trace("compiler.compile_field_reference_rec sibling_fallback destrName=%s" % destr_name)
        if top_context and destr_name in top_context.actions:
            if not expr_context:
                raise IvyError(None,'call to action {} not allowed outside an action'.format(destr_name))
            actparams = top_context.actions[destr_name]
            keypos = actparams[2] if len(actparams) > 2 else 0
            if __debug__: xtracer.trace("compiler.compile_field_reference_rec action_found name=%s keyPos=%d nParams=%d nArgs=%d" % (destr_name, keypos, len(actparams[0]), len(args)))
            args.insert(keypos,base)
            return field_reference_action(destr_name,args,top)
        sym = ivy_logic.find_polymorphic_symbol(destr_name)
        args.insert(0,base)
    if hasattr(sym.sort,'dom') and len(sym.sort.dom) > 0:
        args = pull_args(args,len(sym.sort.dom),sym,top)
        args = [ivy_logic.sort_infer(arg,sort) for arg,sort in zip(args,sym.sort.dom)]
        res = sym(*args)
        return res
    return old_sym(sym,old)
                           
def field_reference_action(actname,args,top):
    nformals = len(top_context.actions[actname][0])
    call = ivy_ast.Atom(actname,[])
    global field_reference_lineno
    call.lineno = field_reference_lineno
    return compile_inline_call(call,pull_args(args,nformals,actname,top),methodcall=True)

def compile_field_reference(symbol_name,args,lineno,old=False):
    if __debug__: xtracer.trace("compiler.compile_field_reference ENTER name=%s" % symbol_name)
    global field_reference_lineno
    field_reference_lineno = lineno
    try:
        return compile_field_reference_rec(symbol_name,args,top=True,old=old)
    except cfrfail as err:
        if symbol_name in ivy_logic.sig.sorts:
            raise IvyError(None,"type {} used where a function or individual symbol is expected".format(err.symbol_name))
        raise IvyError(None,"unknown symbol: {}".format(err.symbol_name))

    
def sort_infer_covariant(term,sort):
    if __debug__: xtracer.trace("compiler.sort_infer_covariant ENTER")
    try:
        return sort_infer(term,sort,True)
    except ivy_logic.Error:
        res = sort_infer(term)
        if not(res.sort == sort or im.module.is_variant(res.sort,sort)):
            raise IvyError(None,"cannot convert argument of type {} to {}".format(sort,res.sort))
        return res

def sort_infer_contravariant(term,sort):
    if __debug__: xtracer.trace("compiler.sort_infer_contravariant ENTER")
    try:
        return sort_infer(term,sort,True)
    except ivy_logic.Error:
        res = sort_infer(term)
        if not(res.sort == sort or im.module.is_variant(sort,res.sort)):
            raise IvyError(None,"cannot convert argument of type {} to {}".format(res.sort,sort))
        return res

def compile_inline_call(self,args,methodcall=False):
    if __debug__: xtracer.trace("compiler.compile_inline_call ENTER")
    rep = resolve_alias(self.rep) if isinstance(self.rep,str) else self.rep
    params,returns,keypos = top_context.actions[rep]
    if return_context is None or return_context.values is None:
        if len(returns) != 1:
            raise IvyError(self,"wrong number of return values")
            # TODO: right now we can't do anything with multiple returns
            sorts = [cmpl_sort(r.sort) for r in returns]
            ress = []
            for sort in sorts:
                res = ivy_logic.Symbol('loc:'+str(len(expr_context.local_syms)),sort)
                expr_context.local_syms.append(res)
                ress.append(res())
            act = CallAction(*([ivy_ast.Atom(rep,args)]+ress))
            act.lineno = self.lineno
            expr_context.code.append(act)
            return ivy_ast.Tuple(*ress)
        sort = cmpl_sort(returns[0].sort)
        res = ivy_logic.Symbol('loc:'+str(len(expr_context.local_syms)),sort)
        expr_context.local_syms.append(res)
        return_values = [res]
    else:
        return_values = return_context.values
        if len(returns) != len(return_values):
            raise IvyError(self,"wrong number of return values")
        return_values = [sort_infer_covariant(a,cmpl_sort(p.sort)) for a,p in zip(return_values,returns)]
    with ASTContext(self):
        if len(params) != len(args):
            # print "self: {}".format(self);
            # print "params: {}".format(map(str,params))
            # print "args: {}".format(map(str,args))
            raise iu.IvyError(self,"wrong number of input parameters (got {}, expecting {})".format(len(args),len(params)))
        args = [sort_infer_contravariant(a,cmpl_sort(p.sort)) for a,p in zip(args,params)]

    call = CallAction(*([ivy_ast.Atom(rep,args)]+return_values))
    call.lineno = self.lineno

    # Handle dispatch for method call with variants

    if methodcall and args[keypos].sort.name in im.module.variants:
#        print 'variants:{}'.format(im.module.variants)
#        print top_context.actions.keys()
        _,methodname = iu.parent_child_name(rep)
        for vsort in im.module.variants[args[keypos].sort.name]:
            vactname = iu.compose_names(vsort.name,methodname)
            if vactname not in top_context.actions:
                parent,_ = iu.parent_child_name(vsort.name)
                vactname = iu.compose_names(parent,methodname)
                if vactname not in top_context.actions or vactname == rep:
                    continue
            tmpsym = ivy_logic.Symbol('self:'+vsort.name,vsort)
            tmpargs = list(args)
            tmpargs[keypos] = tmpsym
            new_call = CallAction(*([ivy_ast.Atom(vactname,tmpargs)]+return_values))
 #           new_call.lineno = self.lineno
            call = IfAction(ivy_ast.Some(tmpsym,ivy_logic.Symbol('*>',ivy_logic.RelationSort([args[keypos].sort,vsort]))(args[keypos],tmpsym)),
                            new_call,
                            call)

    expr_context.code.append(call)
    if return_context is None or return_context.values is None:
        return res()
    return None

def old_sym(sym,old):
    return sym.prefix('old_') if old else sym

def compile_app(self,old=False):
    if __debug__: xtracer.trace("compiler.CompileNode return case=%s" % type(self).__name__)  # Atom or App
    if __debug__: xtracer.trace("compiler.compile_app ENTER old=%s" % old)
    rep = resolve_alias(self.rep) if isinstance(self.rep,str) else self.rep
    if rep == "true" or rep == "false":
        if len(self.args) > 0:
            raise IvyError(self,"{} is not a function".format(rep))
        return ivy_logic.And() if rep == "true" else ivy_logic.Or()
    if __debug__: xtracer.trace("compiler.CompileApp args loop nTerms=%d" % (len(self.args)))
    with ReturnContext(None):
        if any(isinstance(a,ivy_logic.Variable) for a in self.args):
            print("foo!: {}".format(self))
        args = [a.compile() for a in self.args]
    # handle action calls in rhs of assignment
    if expr_context and top_context and rep in top_context.actions:
        # note, we are taking 'old' of an action to be the action, since actions are immutable
        return compile_inline_call(self,args)
    sym = rep.cmpl() if isinstance(rep,ivy_ast.NamedBinder) else ivy_logic.Equals if rep == '=' else ivy_logic.find_polymorphic_symbol(rep,throw=False) 
    if sym is not ivy_logic.Equals:
        if ivy_logic.is_numeral(sym):
            if hasattr(self,'sort') and self.sort != 'S':
                sym = ivy_logic.Symbol(sym.name,variable_sort(self))
    if sym is not None:
        sym = old_sym(sym,old)
        return (sym)(*args)
    res = compile_field_reference(rep,args,self.lineno,old=old)
    return res
    
def compile_method_call(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=MethodCall")
    if __debug__: xtracer.trace("compiler.compile_method_call ENTER")
    with ReturnContext(None):
        base = self.args[0].compile()
        child_name = self.args[1].rep
        args = [a.compile() for a in self.args[1].args]
    sort = base.sort
    if ivy_logic.is_topsort(sort):
        raise IvyError(self,'cannot apply method notation to {} because its type is not inferred'.format(base))
        
    # trucky: we first look for the method as a child of the sort.
    # if not found, we look for a sibling of the sort
    destr_name = iu.compose_names(sort.name,child_name)
    if top_context and destr_name not in ivy_logic.sig.symbols and destr_name not in top_context.actions:
        sort_parent,sort_child = iu.parent_child_name(sort.name)
        destr_name = iu.compose_names(sort_parent,child_name)
    if top_context and destr_name in top_context.actions:
        if not expr_context:
            raise IvyError(None,'call to action {} not allowed outside an action'.format(destr_name))
        args.insert(0,base)
        global field_reference_lineno
        field_reference_lineno = self.lineno
        return field_reference_action(destr_name,args,True) # True means use all args
    sym = ivy_logic.find_polymorphic_symbol(destr_name)
    args.insert(0,base)
    return sym(*args)

ivy_ast.MethodCall.cmpl = compile_method_call

def cmpl_sort(sortname):
    resolved = resolve_alias(sortname)
    try:
        return ivy_logic.find_sort(resolved)
    except IvyError as e:
        if __debug__: xtracer.trace("compiler.CmplSort error\n  name=%s resolved=%s sigSorts=%s err=%s" % (sortname, resolved, list(ivy_logic.sig.sorts.keys()), e))
        raise

def cmpl_native_expr(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=NativeExpr")
    if __debug__: xtracer.trace("compiler.CompileNativeExpr ENTER")
    res = self.clone([a.compile() for a in self.args])
    res.sort = ivy_logic.TopS
    return res

ivy_ast.NativeExpr.cmpl = cmpl_native_expr

ivy_ast.App.cmpl = ivy_ast.Atom.cmpl = compile_app

def compile_isa(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=default type=Isa")
    lhs = self.args[0].compile()
    rhs = cmpl_sort(self.args[1].relname)
    vars = variables_ast(lhs)
    rn = UniqueRenamer(used=[v.name for v in vars])
    v = ivy_logic.Variable(rn('V'),rhs)
    res = ivy_logic.Exists([v],ivy_logic.pto(lhs.sort,rhs)(lhs,v))
    return res
    
ivy_ast.Isa.cmpl = compile_isa

def variable_sort(self):
    return cmpl_sort(self.sort) if isinstance(self.sort,str) else self.sort

def compile_variable(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=Variable")
    if __debug__: xtracer.trace("compiler.compile_variable ENTER")
    with ASTContext(self):
        try:
            sort = variable_sort(self)
        except IvyError as e:
            if __debug__: xtracer.trace("compiler.CompileVariable error\n  var=%s vsort=%s err=%s" % (self.rep, getattr(self,'sort',''), e))
            raise
    if ivy_logic.is_topsort(sort):
        sort = variable_context.map.get(self.rep,sort)
    return ivy_logic.Variable(self.rep,sort)

ivy_ast.Variable.cmpl = compile_variable

ivy_ast.ConstantSort.cmpl = lambda self,name: ivy_logic.ConstantSort(name)

ivy_ast.EnumeratedSort.cmpl = lambda self,name: ivy_logic.EnumeratedSort(name,self.extension)

SymbolList.cmpl = lambda self: self.clone([find_symbol(s) for s in self.symbols])

def cquant(q):
    return ivy_logic.ForAll if isinstance(q,ivy_ast.Forall) else ivy_logic.Exists

def compile_quantifier(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=%s" % type(self).__name__)  # Forall or Exists
    if __debug__: xtracer.trace("compiler.compile_quantifier ENTER")
    bounds = [ivy_logic.Variable(v.rep,variable_sort(v)) for v in self.bounds]
    with VariableContext(bounds):
        return cquant(self)(bounds,self.args[0].compile())

ivy_ast.Quantifier.cmpl = compile_quantifier

def _named_binder_cmpl(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=NamedBinder")
    if __debug__: xtracer.trace("compiler.CompileNamedBinder ENTER")
    return ivy_logic.NamedBinder(
        self.name,
        [v.compile() for v in self.bounds],
        None,
        self.args[0].compile()
    )
ivy_ast.NamedBinder.cmpl = _named_binder_cmpl

def _labeled_formula_cmpl(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=LabeledFormula")
    if __debug__: xtracer.trace("compiler.CompileLabeledFormula ENTER")
    return self.clone([None if self.label is None else self.label.clone([sortify_with_inference(x) for x in self.label.args]),
                       self.formula.compile()
                       if isinstance(self.formula,ivy_ast.SchemaBody)
                       else sortify_with_inference(self.formula)])
ivy_ast.LabeledFormula.cmpl = _labeled_formula_cmpl

ivy_ast.ProofDecl.cmpl = lambda self: self.clone([sortify_with_inference(self.args[0])]+self.args[1:]) 

# compiling update patterns is complicated because they declare constants internally
def UpdatePattern_cmpl(self):
    with ivy_logic.sig.copy():
        return ivy_ast.AST.cmpl(self)
        
UpdatePattern.cmpl = UpdatePattern_cmpl

def ConstantDecl_cmpl(self):
    return self.clone([compile_const(v,ivy_logic.sig) for v in self.args])

ConstantDecl.cmpl =  ConstantDecl_cmpl 

def Old_cmpl(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=Old")
    if __debug__: xtracer.trace("compiler.CompileOld ENTER")
#    foo = self.args[0].compile()
#    return foo.rep.prefix('old_')(*foo.args)
    return compile_app(self.args[0],old=True)

ivy_ast.Old.cmpl = Old_cmpl

def get_arg_sorts(sig,args,term=None):
    if term is not None:
        args = sortify_with_inference(ivy_ast.AST(*(args+[term]))).args[0:-1]
        return [arg.get_sort() for arg in args]
    return [arg.compile().get_sort() for arg in args]
                
def get_function_sort(sig,args,rng,term=None):
    if __debug__: xtracer.trace("compiler.GetFunctionSort ENTER")
    return ivy_logic.FunctionSort(*(get_arg_sorts(sig,args,term)+[rng])) if args else rng

def get_relation_sort(sig,args,term=None):
    return ivy_logic.RelationSort(get_arg_sorts(sig,args,term))

def sortify(ast):
    if __debug__: xtracer.trace("compiler.sortify ENTER")
    # print ("compiling : {}".format(ast))
    return ast.compile()

def sortify_with_inference(ast):
    if __debug__: xtracer.trace("compiler.sortify_with_inference ENTER")
    if __debug__: sig_check("SortifyWithInference")
#    print "compiling : {}".format(ast)
    with top_sort_as_default():
        res = ast.compile()
    with ASTContext(ast):
        res = sort_infer(res)
    return res


ivy_ast.AST.compile_with_sort_inference = sortify_with_inference

def compile_const(v,sig):
    if __debug__: xtracer.trace("compiler.CompileConst ENTER")
    with ASTContext(v):
      hasSort = hasattr(v,'sort')
      if not hasSort:
          if __debug__: xtracer.trace("compiler.CompileConst no_sort calling_default_sort HASH canon=%s" % sig_canon(ivy_logic.sig))
      rng = cmpl_sort(v.sort) if hasSort else ivy_logic.default_sort()
      sort = get_function_sort(sig,v.args,rng)
      with sig:
          return add_symbol(v.rep,sort)
    

def compile_local(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=default type=LocalAction")
    if __debug__: xtracer.trace("compiler.compile_local ENTER")
    sig = ivy_logic.sig.copy()
    with sig:
        ls = self.args[0:-1]
        if len(ls) == 1 and isinstance(ls[0],AssignAction):
            v = ls[0]
            lhs = v.args[0]
            if not hasattr(lhs,"sort"):
                lhs.sort = 'S'
            rhs = v.args[1]
            # temporarily rename lhs symbol in case it clashes with an existing symbol
#            tmp_lhs = lhs.prefix('__var_tmp:')
            tmp_lhs = lhs
            code = []
            local_syms = []
            with ExprContext(code,local_syms):
                with top_sort_as_default():
                    sym = compile_const(tmp_lhs,sig)
                    ctmp_lhs = tmp_lhs.compile()
                    crhs = rhs.compile()
            with ASTContext(self):
                if im.module.is_variant(ctmp_lhs.sort,crhs.sort):
                    teq = sort_infer(ivy_logic.pto(ctmp_lhs.sort,crhs.sort)(ctmp_lhs,crhs))
                else:
                    teq = sort_infer(Equals(ctmp_lhs,crhs))
            clhs,crhs = list(teq.args)
#            clhs = clhs.drop_prefix('__var_tmp:')
            asgn = v.clone([clhs,crhs])
            ivy_logic.remove_symbol(sym)
            if clhs.rep.name in sig.symbols:
                del sig.symbols[clhs.rep.name] # shadow the existing symbol
            ivy_logic.add_symbol(clhs.rep.name,clhs.rep.sort)
            body = sortify(self.args[-1])
            lines = body.args if isinstance(body,Sequence) else [body]
            body = Sequence(*([asgn]+lines))
            code.append(LocalAction(clhs.rep,body,caller="compiler.compile_local_special"))
            for c in code:
                c.lineno = self.lineno
            if len(code) == 1:
                return code[0]
            res = LocalAction(*(local_syms + [Sequence(*code)]),caller="compiler.compile_local_seq")
            res.lineno = self.lineno
            return res
        cls = [compile_const(v,sig) for v in ls]
        body = sortify(self.args[-1])
        res = LocalAction(*(cls+[body]),caller="compiler.compile_local_action")
        res.lineno = self.lineno
        return res

LocalAction.cmpl = compile_local

def compile_assign_lhs(a):
    res = sortify_with_inference(a)
    if not ivy_logic.is_app(res):
        raise IvyError(a,'Invalid expression on left-hand side of assignment')
    return res

def compile_assign(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=Action")  # AssignAction
    if __debug__: xtracer.trace("compiler.compile_assign ENTER")
    # rhs = self.args[1]
    # if (isinstance(rhs,ivy_ast.App) or isinstance(rhs,ivy_ast.Atom)):
    #     if top_context and rhs.rep in top_context.actions:
    #         c = CallAction(rhs,self.args[0])
    #         c.lineno = self.lineno
    #         return c.cmpl()
    code = []
    local_syms = []
    with ExprContext(code,local_syms):
        if isinstance(self.args[0],ivy_ast.Tuple):
            args = [compile_assign_lhs(a) for a in self.args]
            if not isinstance(args[1],ivy_ast.Tuple) or len(args[0].args) != len(args[1].args):
                raise IvyError(self,"wrong number of values in assignment");
            for lhs,rhs in zip(args[0].args,args[1].args):
                code.append(AssignAction(lhs,rhs))
        else:
            with top_sort_as_default():
#                args = [a.compile() for a in self.args]
                args = [self.args[0].compile()]
                if not ivy_logic.is_app(args[0]):
                    raise IvyError(self.args[0],
                                   'Invalid expression on left-hand side of assignment')
                with ReturnContext([args[0]]):
                    args.append(self.args[1].compile())
            if isinstance(args[1],ivy_ast.Tuple):
                raise IvyError(self,"wrong number of values in assignment");
            if args[1] is not None:
                asorts = [a.sort for a in args]
                with ASTContext(self):
                    if im.module.is_variant(*asorts):
                        teq = sort_infer(ivy_logic.pto(*asorts)(*args))
                    else:
                        teq = sort_infer(Equals(*args))
                args = list(teq.args)
                code.append(AssignAction(*args))
        for c in code:
            c.lineno = self.lineno
        if len(code) == 1:
            return code[0]
        res = LocalAction(*(local_syms + [Sequence(*code)]),caller="compiler.compile_cmpd_local")
        res.lineno = self.lineno
        return res

AssignAction.cmpl = compile_assign

def compile_call(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=default type=CallAction")
    if __debug__: xtracer.trace("compiler.compile_call ENTER")
    assert top_context
    ctx = ExprContext(lineno = self.lineno)
    if not hasattr(self.args[0],'rep'):
        raise IvyError(self,'call to non-action')
    name = self.args[0].rep
    if name not in top_context.actions:
        with ctx:
            with ReturnContext([a.cmpl() for a in self.args[1:]]):
                res = compile_field_reference(name,[a.compile() for a in self.args[0].args],self.lineno)
        if res is not None:
            raise IvyError(self,'call to non-action')
        res = ctx.extract()
        return res
    #    print "compiled call action: {}".format(res)
        return res

    with ctx:
        args = [a.cmpl() for a in self.args[0].args]
    params,returns,keypos = top_context.actions[name]
    if len(returns) != len(self.args) - 1:
        raise iu.IvyError(self,"wrong number of output parameters (got {}, expecting {})".format(len(self.args) - 1,len(returns)))
    if len(params) != len(args):
        raise iu.IvyError(self,"wrong number of input parameters (got {}, expecting {})".format(len(args),len(params)))
    with ASTContext(self):
        mas = [sort_infer_contravariant(a,cmpl_sort(p.sort)) for a,p in zip(args,params)]
#        print self.args
    res = CallAction(*([ivy_ast.Atom(name,mas)] + [a.cmpl() for a in self.args[1:]]))
    res.lineno = self.lineno
    ctx.code.append(res)
    res = ctx.extract()
#    print "compiled call action: {}".format(res)
    return res

CallAction.cmpl = compile_call


def compile_if_action(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=default type=IfAction")
    if __debug__: xtracer.trace("compiler.compile_if_action ENTER")
    if isinstance(self.args[0],ivy_ast.Some):
        sig = ivy_logic.sig.copy()
        with sig:
            ls = self.args[0].params()
            fmla = self.args[0].fmla()
            cls = [compile_const(v,sig) for v in ls]
            sfmla = sortify_with_inference(fmla)
            sargs = cls+[sfmla]
            if isinstance(self.args[0],ivy_ast.SomeMinMax):
                sargs.append(sortify_with_inference(self.args[0].index()))
            args = [self.args[0].clone(sargs),self.args[1].compile()]
        args += [a.compile() for a in self.args[2:]]
        return self.clone(args)
    else:
        ctx = ExprContext(lineno = self.lineno)
        with ctx:
            cond = sortify_with_inference(self.args[0])
        rest = [a.compile() for a in self.args[1:]]
        ctx.code.append(self.clone([cond]+rest))
        res = ctx.extract()
        return res

IfAction.cmpl = compile_if_action

def compile_while_action(self):
        if __debug__: xtracer.trace("compiler.CompileNode return case=default type=WhileAction")
        if __debug__: xtracer.trace("compiler.compile_while_action ENTER")
        if isinstance(self.args[0],ivy_ast.Some):
            res = compile_if_action(self.clone(self.args[:2]))
            invars = list(map(sortify_with_inference,self.args[2:]))
            return res.clone(res.args+invars)
        ctx = ExprContext(lineno = self.lineno)
        with ctx:
            cond = sortify_with_inference(self.args[0])
            invars = list(map(sortify_with_inference,self.args[2:]))
        body = self.args[1].compile()
        if ctx.code:
            raise iu.IvyError(self,'while condition may not contain action calls')
        ctx.code.append(self.clone([cond,body]+invars))
        res = ctx.extract()
        return res

WhileAction.cmpl = compile_while_action

def compile_assert_action(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=Action")  # AssertAction or AssumeAction
    if __debug__: xtracer.trace("compiler.compile_assert_action ENTER")
    ctx = ExprContext(lineno = self.lineno)
    with ctx:
        if isinstance(self.args[0],ivy_ast.LabeledFormula):
            cond = self.args[0].compile()
        else:
            cond = sortify_with_inference(self.args[0])
    if len(self.args) > 1:
        pf = self.args[1].compile()
        asrt = self.clone([cond,pf])
    else:
        asrt = self.clone([cond])
    ctx.code.append(asrt)
    res = ctx.extract()
    return res

AssertAction.cmpl = compile_assert_action
AssumeAction.cmpl = compile_assert_action

def compile_crash_action(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=default type=CrashAction")
    if __debug__: xtracer.trace("compiler.compile_crash_action ENTER")
    name = self.args[0].rep
    if isinstance(name,ivy_ast.This):
        name = 'this'
    thing = ivy_ast.Atom(name,list(map(sortify_with_inference,self.args[0].args)))
    res = self.clone([thing])
    return res

CrashAction.cmpl = compile_crash_action

def compile_thunk_action(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=default type=ThunkAction")
    if __debug__: xtracer.trace("compiler.compile_thunk_action ENTER")
    iu.dbg('self')
    global thunk_counter
    sig = ivy_logic.sig.copy()
    with sig:
        formals =  [compile_const(v,sig) for v in self.args[0].args + self.args[1].args]
        body = sortify(self.args[3])
        body.formal_params = formals
        body.formal_returns = []
    symset = set(formals)  # formal parameters don't go in the struct
    syms = []
    for sym in lu.symbols_ilu_ast(body):
        if ((sym.name.startswith('fml:') or sym.name.startswith('loc:'))
            and sym.name in ivy_logic.sig.symbols and sym not in symset):
            symset.add(sym)
            syms.append(sym)
#    subtypename = '{}[thunk{}]'.format(self.args[1].relname,thunk_counter)
    subtypename = self.args[0].relname
    thing = ivy_logic.sig.sorts
    iu.dbg('thing')
    subsort = ivy_logic.find_sort(subtypename)
    selfparam = ivy_logic.Symbol('$self',subsort)
    subs = {}
    dsyms = []
    for sym in syms:
        dsort = ivy_logic.FunctionSort(*([subsort] + sym.sort.dom + [sym.sort.rng]))
        dsym = ivy_logic.Symbol(iu.compose_names(subtypename,sym.name[4:]),dsort)
        iu.dbg('dsym')
        iu.dbg('dsym.sort')
        im.module.destructor_sorts[dsym.name] = subsort
        im.module.sort_destructors[subsort.name].append(dsym)
        subs[sym] = dsym(selfparam)
        dsyms.append(dsym)
    # insert "self" argument in last position (see comment at "def action" below)
    body.formal_params.insert(len(body.formal_params),selfparam)
    new_body = lu.substitute_constants_ast(body,subs)
    new_body.formal_params = body.formal_params
    new_body.formal_returns = body.formal_returns
    body = new_body
    iu.dbg('body.formal_params')
    subtyperun = iu.compose_names(subtypename,'run')
    im.module.actions[subtyperun] = body
    sig = ivy_logic.sig.copy()
    with sig:
        lsym = add_symbol('loc:' + self.args[1].relname,subsort)
        cont = sortify(self.args[4])
    asgns = []
    for sym,dsym in zip(syms,dsyms):
        asgns.append(AssignAction(dsym(lsym),sym))
    res = LocalAction(lsym,Sequence(*(asgns + [cont])),caller="compiler.compile_proof")
    res.lineno = self.lineno
    iu.dbg('res')
    return res

ThunkAction.cmpl = compile_thunk_action

def compile_debug_action(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=default type=DebugAction")
    if __debug__: xtracer.trace("compiler.compile_debug_action ENTER")
    ctx = ExprContext(lineno = self.lineno)
    with ctx:
        withs = [x.clone([x.args[0],sortify_with_inference(x.args[1])]) for x in self.args[1:]]
    dbg = self.clone([self.args[0]] + withs)
    ctx.code.append(dbg)
    res = ctx.extract()
    return res


DebugAction.cmpl = compile_debug_action



def compile_native_arg(arg):
    if isinstance(arg,ivy_ast.Variable):
        return sortify_with_inference(arg)
    if arg.rep in ivy_logic.sig.symbols:
        return sortify_with_inference(arg)
    res = arg.clone(list(map(sortify_with_inference,arg.args))) # handles action names
    return res.rename(resolve_alias(res.rep))


def compile_native_symbol(arg):
    name = arg.rep
    if name in ivy_logic.sig.symbols:
        sym = ivy_logic.sig.symbols[name]
        if not isinstance(sym,ivy_logic.UnionSort):
            return sym
    name = resolve_alias(name)
    if name in ivy_logic.sig.sorts:
        return ivy_logic.Variable('X',ivy_logic.sig.sorts[name])
    if ivy_logic.is_numeral_name(name):
        return ivy_logic.Symbol(name,ivy_logic.TopS)
    if name in im.module.hierarchy:
        return compile_native_name(arg)
    raise iu.IvyError(arg,'{} is not a declared symbol or type'.format(name))

def compile_native_action(self):
    if __debug__: xtracer.trace("compiler.CompileNode return case=default type=NativeAction")
    if __debug__: xtracer.trace("compiler.compile_native_action ENTER")
    fields = self.args[0].code.split('`')
    args = [self.args[0]] + [compile_native_arg(a) if not fields[i*2].endswith('"') else compile_native_symbol(a) for i,a in enumerate(self.args[1:])]
    return self.clone(args)

NativeAction.cmpl = compile_native_action

def compile_native_name(atom):
    assert all(isinstance(a,ivy_ast.Variable) and isinstance(a.sort,str) for a in atom.args)
    return ivy_ast.Atom(atom.rep,[ivy_ast.Variable(a.rep,resolve_alias(a.sort)) for a in atom.args])

def compile_native_def(self):
    fields = self.args[1].code.split('`')
    args = [compile_native_name(self.args[0]),self.args[1]] + [compile_native_arg(a) if not fields[i*2].endswith('"') else compile_native_symbol(a) for i,a in enumerate(self.args[2:])]
    return self.clone(args)

def compile_native_type(self):
    return self.clone([self.args[0]] + [x.rename(resolve_alias(x.rep)) for x in self.args[1:]])

def compile_action_def(a,sig):
    if __debug__: xtracer.trace("compiler.compile_action_def ENTER")
    sig = sig.copy()
    if not hasattr(a.args[1],'lineno'):
        print(a)
    assert hasattr(a.args[1],'lineno')
    with sig:
        with ASTContext(a.args[1]):
            params = a.args[0].args
            pformals = [v.to_const('prm:') for v in params] 
            if params:
                subst = dict((x.rep,y) for x,y in zip(params,pformals))
                a = ivy_ast.substitute_ast(a,subst)
                assert hasattr(a.args[1],'lineno')
    #            a = ivy_ast.subst_prefix_atoms_ast(a,subst,None,None)
    #            print "after: %s" % (a)
            # convert object paramaters to arguments (object-orientation!)
            formals =  [compile_const(v,sig) for v in pformals + a.formal_params]
            returns = [compile_const(v,sig) for v in a.formal_returns]
    #        print returns
            res = sortify(a.args[1])
            assert hasattr(res,'lineno'), res
            for suba in res.iter_subactions():
                if isinstance(suba,CallAction):
                    if any(lu.used_variables_ast(a) for a in suba.args[0].args):
#                        iu.dbg('a.args[0]')
#                        iu.dbg('a.formal_params')
#                        iu.dbg('suba.lineno')
#                        iu.dbg('suba')
                        raise iu.IvyError(suba,"call may not have free variables")
            res.formal_params = formals
            res.formal_returns = returns
            res.label = a.args[0].relname
            return res

def compile_defn(df):
    if __debug__: xtracer.trace("compiler.CompileDefn ENTER")
    if __debug__: xtracer.trace("compiler.CompileDefnImpl ENTER")
    if __debug__: sig_check("CompileDefnImpl.entry")
    has_consts = any(not isinstance(p,ivy_ast.Variable) for p in df.args[0].args)
    sig = ivy_logic.sig.copy() if has_consts else ivy_logic.sig
    is_schema = isinstance(df,ivy_ast.DefinitionSchema)
    with ivy_ast.ASTContext(df):
        with sig:
            if __debug__: sig_check("CompileDefnImpl.afterCopy")
            subst = dict()
            for p in df.args[0].args:
                if not isinstance(p,ivy_ast.Variable):
                    compile_const(p,sig)
                else:
                    if hasattr(p,"sort"):
                        subst[p.rep] = p.sort
            if __debug__: xtracer.trace("compiler.CompileDefnImpl rhs type=%s" % (type(df.args[1]).__name__)) ##\n  isSchema=%s lhs=%s" % (type(df.args[1]).__name__, is_schema, df.args[0]))
            if subst:
                df = df.clone([df.args[0],ivy_ast.set_variable_sorts(df.args[1],subst)])
                if __debug__: xtracer.trace("compiler.CompileDefnImpl rhs after subst type=%s" % (type(df.args[1]).__name__))
            if isinstance(df.args[1],ivy_ast.SomeExpr):
                if __debug__: xtracer.trace("compiler.CompileDefnImpl SomeExpr branch")
                ifval = df.args[1].if_value() or df.args[1].params()[0]
                elseval = df.args[1].else_value() or ifval
                eqn = ivy_ast.Forall(df.args[1].params(),
                                     ivy_ast.Atom('=',(df.args[0],ivy_ast.Ite(df.args[1].fmla(),ifval,elseval))))
                fmla = sortify_with_inference(eqn)
                args = [list(fmla.variables)[0],fmla.body.args[1].args[0]]
                if df.args[1].if_value() :
                    args.append(fmla.body.args[1].args[1])
                if df.args[1].else_value() :
                    args.append(fmla.body.args[1].args[2])
                df = ivy_logic.Definition(fmla.body.args[0],ivy_logic.Some(*args))
            else:
                if False and isinstance(df.args[1],ivy_ast.NativeExpr):
                    df = ivy_logic.Definition(sortify_with_inference(df.args[0]),df.args[1])
                    df.args[1].sort = df.args[0].sort
                else:
                    if __debug__: xtracer.trace("compiler.CompileDefnImpl standard branch rhs type=%s" % (type(df.args[1]).__name__)) ##\n  isSchema=%s lhs=%s rhs=%s" % (type(df.args[1]).__name__, is_schema, df.args[0], df.args[1]))
                    eqn = ivy_ast.Atom('=',(df.args[0],df.args[1]))
                    if __debug__: xtracer.trace("compiler.CompileDefnImpl eqAtom nTerms=%d" % (len(eqn.args)))
                    for ti,tt in enumerate(eqn.args):
                        if tt is None:
                            if __debug__: xtracer.trace("compiler.CompileDefnImpl eqAtom.Terms[%d] = nil" % ti)
                        else:
                            if __debug__: xtracer.trace("compiler.CompileDefnImpl eqAtom.Terms[%d] type=%s" % (ti, type(tt).__name__))
                    eqn = sortify_with_inference(eqn)
                    df = ivy_logic.Definition(eqn.args[0],eqn.args[1])
            if is_schema:
                df = ivy_logic.DefinitionSchema(*df.args)
            return df

def compile_schema_prem(self,sig):
    if isinstance(self,ivy_ast.ConstantDecl):
        with ivy_logic.WithSorts(list(sig.sorts.values())):
            sym = compile_const(self.args[0],sig)
        return self.clone([sym])
    elif isinstance(self,ivy_ast.DerivedDecl):
        raise IvyErr(self,'derived functions in schema premises not supported yet')
    elif isinstance(self,ivy_ast.TypeDef):
        t = ivy_logic.UninterpretedSort(self.args[0].rep)
        sig.sorts[t.name] = t
        return t
    elif isinstance(self,ivy_ast.LabeledFormula):
        with ivy_logic.WithSymbols(sig.all_symbols()):
            with ivy_logic.WithSorts(list(sig.sorts.values())):
                return self.compile()
    # elif isinstance(self,ivy_ast.SchemaBody):
    #     with ivy_logic.WithSymbols(sig.all_symbols()):
    #         with ivy_logic.WithSorts(sig.sorts.values()):
    #             return compile_schema_body(self)
    
def compile_schema_conc(self,sig):
    if __debug__: xtracer.trace("compiler.CompileSchemaConc ENTER") ## \n  concType=%s sigSorts=%s" % (type(self).__name__, list(sig.sorts.keys())))
    if __debug__: sig_check("SchemaConc.entry")
    with ivy_logic.WithSymbols(sig.all_symbols()):
        with ivy_logic.WithSorts(list(sig.sorts.values())):
            if __debug__: sig_check("SchemaConc.withCtx")
            if isinstance(self,ivy_ast.Definition):
                if __debug__: xtracer.trace("compiler.CompileSchemaConc Definition branch") ## \n  sigSorts=%s" % list(ivy_logic.sig.sorts.keys()))
                return compile_defn(self)
            return sortify_with_inference(self)

def compile_schema_body(self):
    sig = ivy_logic.Sig()
    if __debug__: xtracer.trace("compiler.CompileSchemaBody ENTER") ##  freshSig sorts=%s savedSig sorts=%s" % (list(sig.sorts.keys()), list(ivy_logic.sig.sorts.keys())))
    if __debug__: sig_check("SchemaBody.entry")
    prems = [compile_schema_prem(p,sig) for p in self.args[:-1]]
    if __debug__: sig_check("SchemaBody.afterPrems")
    res = ivy_ast.SchemaBody(*(prems+[compile_schema_conc(self.args[-1],sig)]))
    res.instances = []
    return res

ivy_ast.SchemaBody.compile = compile_schema_body    

def lookup_schema(name,proof):
    if name in im.module.schemata:
        return im.module.schemata[name]
    if name in im.module.theorems:
        return im.module.theorems[name]
    raise iu.IvyError(proof,'applied schema {} does not exist'.format(name))

def compile_schema_instantiation(self,fmla):
    return self
    schema = lookup_schema(self.schemaname(),self)
    schemasyms = [x.args[0] for x in schema.prems() if isinstance(x,ivy_ast.ConstantDecl)]
    schemasorts = [s for s in schema.prems() if isinstance(s,ivy_logic.UninterpretedSort)]
    sortmap = dict()
    pairs = []
    for d in self.match():
        x,y = d.lhs(),d.rhs()
        if isinstance(x,ivy_ast.Atom) and any((s.name == x.rep) for s in schemasorts):
            xsort = ivy_logic.UninterpretedSort(x.rep)
            if xsort in sortmap:
                raise iu.IvyError(self,'symbol {} is matched twice'.format(x))
            if y.rep not in ivy_logic.sig.sorts:
                raise iu.IvyError(self,'symbol {} is not a sort'.format(y))
            sortmap[xsort] = ivy_logic.UninterpretedSort(y.rep)
        else:
            pairs.append((x,y))
    sorted_pairs = []
    with top_sort_as_default():
        for x,y in pairs:
            with ivy_logic.WithSymbols(schemasyms):
                x,_ = ivy_logic.sort_infer_list([x.compile(),schema.conc()])
            x = ip.apply_match_alt(sortmap,x)
            y,_,_ = ivy_logic.sort_infer_list([y.compile(),fmla,x])
            sorted_pairs.append((x,y))

    return self.clone([self.args[0]]+
                      [ivy_ast.Definition(x,y) for x,y in sortmap.items()] +
                      [ivy_ast.Definition(x,y) for x,y in sorted_pairs])

last_fmla = None

ivy_ast.TacticWithMatch.compile = lambda self: compile_schema_instantiation(self,last_fmla)

def compile_let_tactic(self):
    return self
    fmla = ivy_ast.And(*[ivy_ast.Atom('=',x.args[0],x.args[1]) for x in self.args])
    thing = sortify_with_inference(fmla)
    return self.clone(thing.args)
    
ivy_ast.LetTactic.compile = compile_let_tactic

def compile_witness_tactic(self):
    return self
    
ivy_ast.WitnessTactic.compile = compile_witness_tactic

def compile_unfold_tactic(self):
    return self
    
ivy_ast.UnfoldTactic.compile = compile_unfold_tactic

def compile_forget_tactic(self):
    return self
    
ivy_ast.ForgetTactic.compile = compile_forget_tactic

def compile_if_tactic(self):
    cond = sortify_with_inference(self.args[0])
    return self.clone([cond,self.args[1].compile(),self.args[2].compile()])
    
ivy_ast.IfTactic.compile = compile_if_tactic

def compile_property_tactic(self):
#    with top_sort_as_default():
#        prop = self.args[0].compile()
    prop = self.args[0]
    name = self.args[1]
    if not isinstance(name,ivy_ast.NoneAST):
        with ivy_logic.UnsortedContext():
            args = [arg.compile() for arg in name.args]
        name = name.clone(args)
    proof = self.args[2].compile()
    return self.clone([prop,name,proof])
        
ivy_ast.PropertyTactic.compile = compile_property_tactic

def compile_function_tactic(self):
    return self
        
ivy_ast.FunctionTactic.compile = compile_function_tactic

def compile_tactic_tactic(self):
    return self.clone(self.args)
    
ivy_ast.TacticTactic.compile = compile_tactic_tactic

def compile_proof_tactic(self):
    return self.clone([self.label,self.proof.compile()])
    
ivy_ast.ProofTactic.compile = compile_proof_tactic

def compile_trigger(self):
    return self.clone([self.args[0]]+[sortify_with_inference(x) for x in self.args[1:]])

ivy_ast.Trigger.compile = compile_trigger

def resolve_alias_int(name): 
    if name in im.module.aliases:
        return im.module.aliases[name]
    parts = name.rsplit(iu.ivy_compose_character,1)
    if len(parts) == 2:
        return resolve_alias_int(parts[0]) + iu.ivy_compose_character + parts[1]
    return name

def resolve_alias(name): 
    res = resolve_alias_int(name)
    return res

defined_attributes = set(["weight","test","check","mc","bmc","method","separate","iterable","cardinality","radix","override","cppstd","libspec","macro_finder","global_parameter","complete"])

class IvyDomainSetup(IvyDeclInterp):
    def __init__(self,domain):
        self.domain = domain
    def alias(self,dfn):
        if __debug__: xtracer.trace("compiler.DomainSetup.alias ENTER")
        self.domain.aliases[dfn.defines()] = resolve_alias(dfn.args[1].rep)
    def object(self,atom):
        if __debug__: xtracer.trace("compiler.DomainSetup.object ENTER")
        self.domain.add_object(atom.rep)
    def axiom(self,ax):
        if __debug__: xtracer.trace("compiler.DomainSetup.axiom ENTER")
        cax = ax.compile()
        if isinstance(cax.formula,ivy_ast.SchemaBody):
            self.domain.schemata[cax.label.relname] = cax
        else:
            self.domain.labeled_axioms.append(cax)
    def property(self,ax):
        if __debug__: xtracer.trace("compiler.DomainSetup.property ENTER")
        lf = ax.compile()
        self.domain.labeled_props.append(lf)
        self.last_fact = lf
    def conjecture(self,c):
        if __debug__: xtracer.trace("compiler.DomainSetup.conjecture ENTER")
        self.last_fact = None
    def named(self,lhs):
        if __debug__: xtracer.trace("compiler.DomainSetup.named ENTER")
        cond = ivy_logic.drop_universals(self.last_fact.formula)
        if not ivy_logic.is_exists(cond) or len(cond.variables) != 1:
            raise IvyError(lhs,'property is not existential')
        rng = list(cond.variables)[0].sort
        vmap = dict((x.name,x) for x in lu.variables_ast(cond))
        used = set()
        with ivy_logic.UnsortedContext():
            args = [arg.compile() for arg in lhs.args]
        targs = []
        for a in args:
            if a.name in used:
                raise IvyError(lhs,'repeat parameter: {}'.format(a.name))
            used.add(a.name)
            if a.name in vmap:
                v = vmap[a.name]
                targs.append(v)
                if not (ivy_logic.is_topsort(a.sort) or a.sort != v.sort):
                    raise IvyError(lhs,'bad sort for {}'.format(a.name))
            else:
                if ivy_logic.is_topsort(a.sort):
                    raise IvyError(lhs,'cannot infer sort for {}'.format(a.name))
                targs.append(a)
        for x in vmap:
            if x not in used:
                raise IvyError(lhs,'{} must be a parameter of {}'.format(x,lhs.rep))
        dom = [x.sort for x in targs]
        sym = self.domain.sig.add_symbol(lhs.rep,ivy_logic.FuncConstSort(*(dom+[rng])))
        self.domain.named.append((self.last_fact,sym(*targs) if targs else sym))
        self.domain.updates.append(NamedUpdate(sym,cond))
    def schema(self,sch):
        if __debug__: xtracer.trace("compiler.DomainSetup.schema ENTER")
        if __debug__: sig_check("DomainSetup.schema")
        if isinstance(sch.defn.args[1],ivy_ast.SchemaBody):
            label = ivy_ast.Atom(sch.defn.defines(),[])
            ldf = ivy_ast.LabeledFormula(label,sch.defn.args[1].compile())
            ldf.lineno = sch.defn.args[1].lineno
#            self.domain.labeled_axioms.append(ldf)
            self.domain.schemata[label.relname] = ldf
        else:
            self.domain.schemata[sch.defn.defines()] = sch
    def theorem(self,sch):
        if __debug__: xtracer.trace("compiler.DomainSetup.theorem ENTER")
        if isinstance(sch.defn.args[1],ivy_ast.SchemaBody):
            label = ivy_ast.Atom(sch.defn.defines(),[])
            ldf = ivy_ast.LabeledFormula(label,sch.defn.args[1].compile())
            ldf.lineno = sch.defn.args[1].lineno
            self.domain.labeled_props.append(ldf)
            self.domain.theorems[label.relname] = ldf.formula
            self.last_fact = ldf
    def instantiate(self,instantiation):
        if __debug__: xtracer.trace("compiler.DomainSetup.instantiate ENTER")
        pref, inst = instantiation.args
        try:
            self.domain.schemata[inst.relname].instantiate(inst.args)
        except LookupError:
            raise IvyError(inst,"{} undefined in instantiation".format(inst.relname))
        except IndexError:
            raise IvyError(inst,"wrong number of parameters in instantiation")
    def relation(self,rel):
        if __debug__: xtracer.trace("compiler.DomainSetup.relation ENTER")
        with ASTContext(rel):
            sym = add_symbol(rel.relname,get_relation_sort(self.domain.sig,rel.args))
        self.domain.all_relations.append((sym,len(rel.args)))
        self.domain.relations[sym] = len(rel.args)
    def individual(self,v):
        if __debug__: xtracer.trace("compiler.DomainSetup.individual ENTER")
        sym = compile_const(v,self.domain.sig)
#        print "sym: {!r}".format(sym)
        self.domain.functions[sym] = len(v.args)
        return sym
    def parameter(self,v):
        if __debug__: xtracer.trace("compiler.DomainSetup.parameter ENTER")
        if isinstance(v,ivy_ast.Definition):
            sym = self.individual(v.args[0])
            dflt = v.args[1]
        else:
            sym = self.individual(v)
            dflt = None
        self.domain.params.append(sym)
        self.domain.param_defaults.append(dflt)
        return sym
    def destructor(self,v):
        if __debug__: xtracer.trace("compiler.DomainSetup.destructor ENTER")
        sym = self.individual(v)
        dom = sym.sort.dom
        if not dom:
            raise IvyError(v,"A destructor must have at least one parameter")
        self.domain.destructor_sorts[sym.name] = dom[0]
        self.domain.sort_destructors[dom[0].name].append(sym)
    def constructor(self,v):
        if __debug__: xtracer.trace("compiler.DomainSetup.constructor ENTER")
        sym = self.individual(v)
        rng = sym.sort.rng
        self.domain.constructor_sorts[sym.name] = rng
        sortname = rng.name
        if sortname in self.domain.sort_destructors:
            destrs = self.domain.sort_destructors[sortname]
            if not any(len(f.sort.dom) > 1 for f in destrs):
                dom = sym.sort.dom
                if len(dom) == 0 and len(destrs) > 0:
                    new_dom = [f.sort.rng for n,f in enumerate(destrs)]
                    ivy_logic.remove_symbol(sym)
                    ivy_logic.add_symbol(sym.name,ivy_logic.FunctionSort(*(new_dom + [sym.sort.rng])))
                    sym = ivy_logic.find_symbol(sym.name)
        self.domain.sort_constructors[rng.name].append(sym)



    def add_definition(self,ldf):
        defs = self.domain.native_definitions if isinstance(ldf.formula.args[1],ivy_ast.NativeExpr) else self.domain.labeled_props
        lhsvs = list(lu.variables_ast(ldf.formula.args[0]))
        for idx,v in enumerate(lhsvs):
            if v in lhsvs[idx+1:]:
                raise IvyError(ldf,"Variable {} occurs twice on left-hand side of definition".format(v))
        for v in lu.used_variables_ast(ldf.formula.args[1]):
            if v not in lhsvs:
                raise IvyError(ldf,"Variable {} occurs free on right-hand side of definition".format(v))
        defs.append(ldf)
        self.last_fact = ldf

    def derived(self,ldf):
        if __debug__: xtracer.trace("compiler.DomainSetup.derived ENTER")
        if __debug__: sig_check("DomainSetup.derived")
        try:
            label = ldf.label
            df = ldf.formula
            lhs = df.args[0]
            sym = ivy_logic.add_symbol(lhs.rep,ivy_logic.TopFunctionSort(len(lhs.args)))
            df  = compile_defn(df)
            ivy_logic.remove_symbol(sym)
            add_symbol(df.args[0].rep.name,df.args[0].rep.sort)
            self.domain.all_relations.append((sym,len(lhs.args)))
            self.domain.relations[sym] = len(lhs.args)
            self.add_definition(ldf.clone([label,df]))
            self.domain.updates.append(DerivedUpdate(df))
            self.domain.symbol_order.append(sym)
        except ValueError:
            raise IvyError(df,"definition of derived relation must be a cube")
    def definition(self,ldf):
        if __debug__: xtracer.trace("compiler.DomainSetup.definition ENTER")
        if __debug__: sig_check("DomainSetup.definition")
        label = ldf.label
        df = ldf.formula
        df = compile_defn(df)
        self.add_definition(ldf.clone([label,df]))
        self.domain.updates.append(DerivedUpdate(df))
        self.domain.symbol_order.append(df.args[0].rep)
        if not self.domain.sig.contains_symbol(df.args[0].rep):
            add_symbol(df.args[0].rep.name,df.args[0].rep.sort)
            
    def proof(self,pf):
        if __debug__: xtracer.trace("compiler.DomainSetup.proof ENTER")
        if isinstance(pf,ivy_ast.LabeledFormula):   # proof has a label
            self.domain.proofs.append((ivy_ast.LabeledFormula(pf.label,None),pf.formula.compile()))
            return
        if self.last_fact is None:
            return # this is a conjecture
        global last_fmla
        last_fmla = self.last_fact.formula
        # tricky: if proof is of a definition, move the definition to labeled_props
        # if isinstance(last_fmla,ivy_logic.Definition):
        #     defs = self.domain.definitions
        #     assert len(defs) >= 1 and defs[-1].id == self.last_fact.id
        #     self.domain.labeled_props.append(self.last_fact)
        #     defs.pop()
        self.domain.proofs.append((self.last_fact,pf.compile()))

    def progress(self,df):
        if __debug__: xtracer.trace("compiler.DomainSetup.progress ENTER")
        rel = df.args[0]
        with ASTContext(rel):
            sym = add_symbol(rel.relname,get_relation_sort(self.domain.sig,rel.args,df.args[1]))
        df = sortify_with_inference(df)
        self.domain.progress.append(df)
    def rely(self,df):
        if __debug__: xtracer.trace("compiler.DomainSetup.rely ENTER")
        df = sortify_with_inference(df)
        self.domain.rely.append(df)
    def mixord(self,df):
        if __debug__: xtracer.trace("compiler.DomainSetup.mixord ENTER")
        self.domain.mixord.append(df)
    def concept(self,c):
        if __debug__: xtracer.trace("compiler.DomainSetup.concept ENTER")
        rel = c.args[0]
        with ASTContext(c):
          add_symbol(rel.relname,get_relation_sort(self.domain.sig,rel.args,c.args[1]))
        c = sortify_with_inference(c)
        self.domain.concept_spaces.append((c.args[0],c.args[1]))
    def update(self,upd):
        if __debug__: xtracer.trace("compiler.DomainSetup.update ENTER")
        self.domain.updates.append(upd.compile())
    def type(self,typedef):
#        print "typedef {!r}".format(typedef)
        #if __debug__: xtracer.trace("compiler.DomainSetup.type ENTER\n pyClass=%s" % type(typedef).__name__)
        if __debug__: xtracer.trace("compiler.DomainSetup.type ENTER")
        if isinstance(typedef.name,ivy_ast.This):
            raise IvyError(typedef,"'type this' is only allowed in an object")
        self.domain.sort_order.append(typedef.name)
        if isinstance(typedef,ivy_ast.GhostTypeDef):
            self.domain.ghost_sorts.add(typedef.name)
        if typedef.finite:
            self.domain.finite_sorts.add(typedef.name)
        if isinstance(typedef.value,ivy_ast.StructSort):
            sort = ivy_logic.ConstantSort(typedef.name)
            if __debug__: xtracer.trace("compiler.DomainSetup.type sort=struct name=%s" % typedef.name)
            self.domain.sig.sorts[typedef.name] = sort
            # handle empty struct
            if typedef.name not in self.domain.sort_destructors:
                self.domain.sort_destructors[typedef.name] = []
            for a in typedef.value.args:
                p = a.clone([ivy_ast.Variable('V:dstr',sort.name)]+a.args)
                if not hasattr(a,'sort'):
                    raise IvyError(a,'no sort provided for field {}'.format(a))
                p.sort = a.sort
                with ASTContext(typedef):
                    with ASTContext(a):
                        self.destructor(p)
            return
        with ASTContext(typedef.value):
            sort = typedef.value.cmpl(typedef.name)
        if __debug__:
            sort_type = type(sort).__name__
            if hasattr(sort,'defines'):
                xtracer.trace("compiler.DomainSetup.type sort=%s name=%s ext=%s" % (sort_type, typedef.name, list(sort.defines())))
            else:
                xtracer.trace("compiler.DomainSetup.type sort=%s name=%s" % (sort_type, typedef.name))
        self.domain.sig.sorts[typedef.name] = sort
        for c in sort.defines(): # register the type's constructors
            sym = Symbol(c,sort)
            if __debug__: xtracer.trace("compiler.DomainSetup.type enum_constructor name=%s sort=%s" % (c, sort))
            self.domain.functions[sym] = 0
            self.domain.sig.symbols[c] = sym
            self.domain.sig.constructors.add(sym)
    def variant(self,v):
        if __debug__: xtracer.trace("compiler.DomainSetup.variant ENTER")
        for r in v.args:
            if r.rep not in self.domain.sig.sorts:
                raise IvyError(v,"undefined sort: {}".format(r.rep))
        self.domain.variants[v.args[1].rep].append(self.domain.sig.sorts[v.args[0].rep])
        self.domain.supertypes[v.args[0].rep] = self.domain.sig.sorts[v.args[1].rep]
    def implementtype(self,thing):
        if __debug__: xtracer.trace("compiler.DomainSetup.implementtype ENTER")
        v = thing.formula
        for r in v.args:
            if r.rep not in self.domain.sig.sorts:
                raise IvyError(v,"undefined sort: {}".format(r.rep))
        impd = v.implemented()
        impr = v.implementer()
        if (impd in self.domain.native_types or
            impd in self.domain.sig.interp):
            raise IvyError(v,"{} is already interpreted".format(impd))
#        if impd in ivy_logic.sort_dependencies(impr):
#            raise IvyError(v,"cannot implement type {} with {} because {} depends on {}".format(impd,impr,impr,impd))
        ivy_logic.implement_type(ivy_logic.find_sort(impd),ivy_logic.find_sort(impr))
        self.domain.interps[impd].append(thing)
        
    def interpret(self,thing):
        if __debug__: xtracer.trace("compiler.DomainSetup.interpret ENTER")
        sig = self.domain.sig
        interp = sig.interp
        lhs = resolve_alias(thing.formula.args[0].rep)
        rhs = thing.formula.args[1]
        if __debug__: xtracer.trace("compiler.DomainSetup.interpret lhs=%s rhs=%s" % (lhs, type(rhs).__name__))
        if isinstance(thing.formula.args[1],ivy_ast.NativeType):
            if __debug__: xtracer.trace("compiler.DomainSetup.interpret branch=nativeType")
            if lhs in interp or lhs in self.domain.native_types :
                raise IvyError(thing,"{} is already interpreted".format(lhs))
            self.domain.native_types[lhs] = compile_native_type(thing.formula.args[1])
            if thing.formula.args[1].args[0].code.strip() == 'int':
                if __debug__: xtracer.trace("compiler.DomainSetup.interpret branch=nativeType-int")
                compile_theory(self.domain,lhs,'int')
            if __debug__: xtracer.trace("compiler.DomainSetup.interpret return=nativeType")
            return
        rhs = thing.formula.args[1].rep
        if __debug__: xtracer.trace("compiler.DomainSetup.interpret branch=non-native rhsName=%s" % (rhs if isinstance(rhs,str) else ''))
        self.domain.interps[lhs].append(thing)
        if lhs in self.domain.native_types :
            raise IvyError(thing,"{} is already interpreted".format(lhs))
        if lhs in interp:
            if __debug__: xtracer.trace("compiler.DomainSetup.interpret branch=already-interpreted")
            if interp[lhs] != rhs:
                raise IvyError(thing,"{} is already interpreted".format(lhs))
            if __debug__: xtracer.trace("compiler.DomainSetup.interpret return=already-interpreted")
            return
        if isinstance(rhs,ivy_ast.Range):
            if __debug__: xtracer.trace("compiler.DomainSetup.interpret branch=range")
            #            interp[lhs] = ivy_logic.EnumeratedSort(lhs,["{}:{}".format(i,lhs) for i in range(int(rhs.lo),int(rhs.hi)+1)])
            if lhs not in sig.sorts:
                raise IvyError(thing,"{} is not a sort".format(lhs))
            sort = sig.sorts[lhs]
            if not isinstance(sort,ivy_logic.UninterpretedSort):
                raise IvyError(thing,"{} is already interpreted".format(lhs))
            def compile_bound(b):
                if not ivy_logic.is_numeral_name(b.rep):
                    b.sort = lhs
                    self.parameter(b)
                with top_sort_as_default():
                    res = b.compile()
                with ASTContext(thing):
                    res = sort_infer(res,sort)
                return res
            lo = compile_bound(rhs.lo)
            hi = compile_bound(rhs.hi)
            interp[lhs] = ivy_logic.RangeSort(lhs,lo,hi)
            compile_theory(self.domain,lhs,interp[lhs])
            if __debug__: xtracer.trace("compiler.DomainSetup.interpret return=range")
            return
        if isinstance(rhs,ivy_ast.EnumeratedSort):
            if __debug__: xtracer.trace("compiler.DomainSetup.interpret branch=enum")
            if lhs not in self.domain.sig.sorts:
                iu.dbg('type(lhs)')
                raise IvyError(thing,"{} is not a type".format(lhs))
            sort = ivy_logic.EnumeratedSort(lhs,[x.rep for x in rhs.args])
            interp[lhs] = sort
            for c in sort.defines(): # register the type's constructors
                sym = Symbol(c,self.domain.sig.sorts[lhs])
                self.domain.functions[sym] = 0
                self.domain.sig.symbols[c] = sym
                self.domain.sig.constructors.add(sym)
            print(interp[lhs])
            if __debug__: xtracer.trace("compiler.DomainSetup.interpret return=enum")
            return
        for x,y,z in zip([sig.sorts,sig.symbols],
                         [slv.is_solver_sort,slv.is_solver_op],
                         ['sort','symbol']):
            if lhs in x:
                if __debug__: xtracer.trace("compiler.DomainSetup.interpret branch=solver-%s" % z)
                if not y(rhs):
                    raise IvyError(thing,"{} not a native {}".format(rhs,z))
                interp[lhs] = rhs
                if z == 'sort' and isinstance(rhs,str):
                    compile_theory(self.domain,lhs,rhs)
                if __debug__: xtracer.trace("compiler.DomainSetup.interpret return=solver-%s" % z)
                return
        if __debug__: xtracer.trace("compiler.DomainSetup.interpret branch=undefined")
        raise IvyUndefined(thing,lhs)
    def scenario(self,scen):
        if __debug__: xtracer.trace("compiler.DomainSetup.scenario ENTER")
        for (s,lineno) in scen.places():
            with ASTContext(scen):
                sym = add_symbol(s,ivy_logic.RelationSort([]))
                self.domain.all_relations.append((sym,0))
                self.domain.relations[sym] = 0

    def mixin(self,m):
        if __debug__: xtracer.trace("compiler.DomainSetup.mixin ENTER")
        if m.args[1].relname  != 'init' and m.args[1].relname not in top_context.actions:
            raise IvyError(m,'unknown action: {}'.format(m.args[1].relname))

    # Here, we scan the actions for thunks to declare the thunk types
    def action(self,a):
        if __debug__: xtracer.trace("compiler.DomainSetup.action ENTER")
        counter = 0
        for action in a.args[1].iter_subactions():
            if isinstance(action,ThunkAction):
#                subtype = ivy_ast.Atom('{}[thunk{}]'.format(action.args[1].relname,counter))
                subtype = action.args[0]

                iu.dbg('action')

                counter += 1
                suptype = action.args[2]
                tdef = ivy_ast.TypeDef(subtype,ivy_ast.UninterpretedSort())
                self.type(tdef)

                vdef = ivy_ast.VariantDef(subtype,suptype)
                self.variant(vdef)

                actname = iu.compose_names(subtype.relname,'run')
                selfparam = ivy_ast.Atom('fml:$self',[])
                selfparam.sort = subtype.relname

                # tricky: the key position for the thunk is the *last* argument.
                # this is because the first N arguments might be stripped, so we
                # can't insert an argument at position zero.
                orig_args = action.args[0].args + action.args[1].args
                top_context.actions[actname] = (orig_args + [selfparam], [], len(orig_args))


            
class IvyConjectureSetup(IvyDeclInterp):
    def __init__(self,domain):
        self.domain = domain
        self.last_fact = None
    def conjecture(self,ax):
        cax = ax.compile()
        #if __debug__: xtracer.trace("compiler.ConjSetup.conjecture compiled\n pyType=%s sort=%s" % (type(cax.formula).__name__, cax.formula.sort() if hasattr(cax.formula,'sort') else '?'))
        if __debug__: xtracer.trace("compiler.ConjSetup.conjecture compiled")
        self.domain.labeled_conjs.append(cax)
        self.last_fact = cax
    def property(self,p):
        if __debug__: xtracer.trace("compiler.ConjSetup.property ENTER")
        self.last_fact = None
    def definition(self,p):
        if __debug__: xtracer.trace("compiler.ConjSetup.definition ENTER")
        self.last_fact = None
    def theorem(self,sch):
        if __debug__: xtracer.trace("compiler.ConjSetup.theorem ENTER")
        self.last_fact = None
    def proof(self,pf):
        if __debug__: xtracer.trace("compiler.ConjSetup.proof ENTER")
        if self.last_fact is None or isinstance(pf,ivy_ast.LabeledFormula):
            return # this is a not a conjecture, or proof is labeled
        global last_fmla
        last_fmla = self.last_fact.formula
        self.domain.proofs.append((self.last_fact,pf.compile()))

def check_is_action(mod,ast,name):
    if name not in mod.actions:
        raise IvyError(ast,'{} is not an action'.format(name))

class IvyARGSetup(IvyDeclInterp):
    def __init__(self,mod):
        self.mod = mod
#    def individual(self,v):
#        self.mod.pvars.append(sortify(v).rep)
    def init(self,s):
        la = s.compile()
        if __debug__: xtracer.trace("compiler.ARGSetup.init compiled sort=%s" % (la.formula.sort if hasattr(la.formula,'sort') else '?'))
        self.mod.labeled_inits.append(la)
        init_clauses = formula_to_clauses(la.formula)
        if __debug__: xtracer.trace("compiler.ARGSetup.init clauses fmlas=%d defs=%d" % (len(init_clauses.fmlas), len(init_clauses.defs)))
        im.module.init_cond = and_clauses(im.module.init_cond, init_clauses)
    def action(self,a):
        global thunk_counter
        thunk_counter = 0
        name = a.args[0].relname
        if __debug__: xtracer.trace("compiler.ARGSetup.action ENTER name=%s" % name)
        self.mod.actions[name] = compile_action_def(a,self.mod.sig)
        self.mod.public_actions[name] = True
        if __debug__: xtracer.trace("compiler.ARGSetup.action EXIT name=%s key=%s" % (name, name))
    def state(self,a):
        if __debug__: xtracer.trace("compiler.ARGSetup.state ENTER")
        self.mod.predicates[a.args[0].relname] = a.args[1]
    def mixin(self,m):
        if __debug__: xtracer.trace("compiler.ARGSetup.mixin ENTER")
#        if any(a.args for a in m.args):
#            raise IvyError(m,"mixins may not have parameters")
        self.mod.mixins[m.args[1].relname].append(m)
        if __debug__: xtracer.trace("compiler.ARGSetup.mixin mixee=%s mixer=%s kind=%s" % (m.args[1].relname, m.args[0].relname, type(m).__name__))
    def _assert(self,a):
        if __debug__: xtracer.trace("compiler.ARGSetup._assert ENTER")
        with ASTContext(a):
            self.mod.assertions.append(type(a)(a.args[0],sortify_with_inference(a.args[1])))
    def isolate(self,iso):
        args = [a.rename('this') if isinstance(a.rep,ivy_ast.This) else a for a in iso.args]
        if hasattr(iso,'is_object') and iso.is_object:
            args[0] = args[0].clone([])  # strip off parameters from isolate objects
            args[1] = args[1].clone([])
        if __debug__: xtracer.trace("compiler.ARGSetup.isolate ENTER name=%s" % iso.name())
        self.mod.isolates[iso.name()] = iso.clone(args)
        if __debug__: xtracer.trace("compiler.ARGSetup.isolate EXIT name=%s" % iso.name())
    def export(self,exp):
        if __debug__: xtracer.trace("compiler.ARGSetup.export ENTER name=%s" % exp.exported())
        check_is_action(self.mod,exp,exp.exported())
        self.mod.exports.append(exp)
        if __debug__: xtracer.trace("compiler.ARGSetup.export EXIT name=%s" % exp.exported())
    def import_(self,imp):
        if __debug__: xtracer.trace("compiler.ARGSetup.import_ ENTER")
        check_is_action(self.mod,imp,imp.imported())
        self.mod.imports.append(imp)
    def private(self,pvt):
        if __debug__: xtracer.trace("compiler.ARGSetup.private ENTER")
        self.mod.privates.add(pvt.privatized())
    def delegate(self,exp):
        if __debug__: xtracer.trace("compiler.ARGSetup.delegate ENTER")
        self.mod.delegates.append(exp)
    def native(self,native_def):
        if __debug__: xtracer.trace("compiler.ARGSetup.native ENTER")
        self.mod.natives.append(compile_native_def(native_def))
    def attribute(self,a):
        if __debug__: xtracer.trace("compiler.ARGSetup.attribute ENTER")
        lhs,rhs = a.args
#        if len(lhs.args) != 0:
#            raise IvyError(a,'attribute names may not have parameters')
        fields = lhs.rep.split(iu.ivy_compose_character)
        oname = iu.ivy_compose_character.join(fields[:-1])
        oname = 'this' if oname == '' else oname
        aname = fields[-1]
        if oname not in self.mod.actions and oname not in self.mod.hierarchy and oname != 'this' and oname not in ivy_logic.sig.sorts and oname not in ivy_logic.sig.symbols and oname not in self.mod.isolates:
            raise IvyError(a,'"{}" does not name an action, object or type'.format(oname))
        if aname not in defined_attributes:
            raise IvyError(a,'"{}" does not name a defined attribute'.format(aname))
        if aname == 'complete' and rhs.rep not in ivy_logic.logics:
            raise IvyError(a,'"{}" is not a known logic'.format(rhs))
        self.mod.attributes[lhs.rep] = rhs
    def scenario(self,scen):
        if __debug__: xtracer.trace("compiler.ARGSetup.scenario ENTER")
        init_tokens = set(p.rep for p in scen.args[0].args)
        transs_by_action = defaultdict(list)
        for tr in scen.args[1:]:
            transs_by_action[tr.args[2].args[1].args[0].rep].append(tr)
        for (place_name,lineno) in scen.places():
            sym = find_symbol(place_name)
            iname = place_name + '[init]'
            iact = AssignAction(sym,ivy_logic.And() if (place_name in init_tokens) else ivy_logic.Or())
            iact.formal_params = []
            iact.formal_returns = []
            iact.lineno = scen.lineno
            self.mod.actions[iname] = iact
            self.mixin(ivy_ast.MixinAfterDef(ivy_ast.Atom(iname),ivy_ast.Atom('init')))
        for actname,trs in transs_by_action.items():
            choices = []
            params = None
            afters = []
            for tr in trs:
                scmix = tr.args[2]
                is_after = isinstance(scmix,ivy_ast.ScenarioAfterMixin)
                df = scmix.args[1]
                body = compile_action_def(df,self.mod.sig)
                seq = []
                if not is_after:
                    for p in tr.args[0].args:
                        seq.append(AssumeAction(find_symbol(p.rep)))
                    for p in tr.args[0].args:
                        seq.append(AssignAction(find_symbol(p.rep),ivy_logic.Or()))
                    for p in tr.args[1].args:
                        seq.append(AssignAction(find_symbol(p.rep),ivy_logic.And()))
                    seq.append(body)
                    seq = Sequence(*seq)
                else:
                    for p in tr.args[0].args:
                        seq.append(AssignAction(find_symbol(p.rep),ivy_logic.Or()))
                    for p in tr.args[1].args:
                        seq.append(AssignAction(find_symbol(p.rep),ivy_logic.And()))
                    seq.append(body)
                    seq = Sequence(*seq)
                    seq = IfAction(And(*[find_symbol(p.rep) for p in tr.args[0].args]),seq)
                if params is None:
                    params = body.formal_params
                    returns = body.formal_returns
                    mixer = scmix.args[0]
                    mixee = scmix.args[1].args[0]
                else:
                    aparams = df.formal_params + df.formal_returns
                    subst = dict(list(zip(aparams,params+returns)))
                    seq = substitute_constants_ast(seq,subst)
                seq.lineno = tr.lineno
                if not is_after:
                    choices.append(seq)
                else:
                    afters.append(seq)
            if choices:
                choice = BalancedChoice(choices)
                choice.lineno = choices[0].lineno
                choice.formal_params = params
                choice.formal_returns = returns
                self.mod.actions[mixer.rep] = choice
                self.mixin(ivy_ast.MixinBeforeDef(mixer,mixee))
            if afters:
                choice = Sequence(*afters)
                choice.lineno = afters[0].lineno
                choice.formal_params = params
                choice.formal_returns = returns
                self.mod.actions[mixer.rep] = choice
                self.mixin(ivy_ast.MixinAfterDef(mixer,mixee))
                
        
def BalancedChoice(choices):
    if __debug__: xtracer.trace("compiler.BalancedChoice ENTER")
    if len(choices) == 1:
        return choices[0]
    return ChoiceAction(BalancedChoice(choices[0:len(choices)/2]),
                        BalancedChoice(choices[len(choices)/2:]))

def get_file_version(filename):
    try:
        f = open(filename,'r')
    except:
        raise IvyError(None,"not found: %s" % filename)
    header = f.readline()
    header = str.strip(header)
    if header.startswith('#lang ivy'):
        version = header[len('#lang ivy'):]
        if version.strip() != '':
            return iu.string_version_to_numeric_version(version)
    return None

def ivy_new(filename = None):
#    d = Interp()
    if filename:
        f = open(filename,'r')
        if not f:
            raise IvyError(None,"not found: %s" % filename)
        ivy_load_file(f)
    ag = AnalysisGraph()
    return ag

isolate = iu.Parameter("isolate")    

# collect actions in case of forward reference
def collect_actions(decls):
    res = {}
    for d in decls:
        if d.name() == 'action':
            for a in d.args:
                formals = a.args[0].args + a.formal_params
                keypos = 0
                for idx,p in enumerate(formals):
                    if isinstance(p,ivy_ast.KeyArg):
                        keypos = idx
                res[a.defines()] = (formals,a.formal_returns,keypos)
    return res

def infer_parameters(decls):
    if __debug__: xtracer.trace("compiler.InferParameters ENTER")
    mixees = defaultdict(list)
    actdecls = dict()
    for d in decls:
        if d.name() == "action":
            for a in d.args:
                actdecls[a.defines()] = a
    for d in decls:
        if d.name() == "mixin":
            for a in d.args:
                mixeename = a.args[1].relname
                if mixeename == "init":
                    continue
                if mixeename not in actdecls:
                    raise IvyError(a,"undefined action: {}".format(mixeename))
                mixees[a.args[0].relname].append(mixeename)
    for d in decls:
        if d.name() == "action":
            for a in d.args:
                am = mixees[a.defines()]
                if len(am) == 1 and am[0] in actdecls:
                    mixin = a.args[1]
                    mixee = actdecls[am[0]]
                    nparms = len(a.args[0].args)
                    mnparms = len(mixee.args[0].args)
                    if len(a.formal_params) + nparms > len(mixee.formal_params) + mnparms:
                        raise iu.IvyError(a.args[1],'monitor has too many input parameters for {}'.format(mixee.defines()))
                    if len(a.formal_returns) > len(mixee.formal_returns):
                        raise iu.IvyError(a.args[1],'monitor has too many output parameters for {}'.format(mixee.defines()))
                    required = mnparms - nparms
                    if len(a.formal_params) < required:
                        raise iu.IvyError(a.args[1],'monitor must supply at least {} explicit input parameters for {}'.format(required,mixee.defines()))
                    xtraps = (mixee.args[0].args+mixee.formal_params)[len(a.formal_params)+nparms:]
                    xtrars = mixee.formal_returns[len(a.formal_returns):]
                    if xtraps or xtrars:
                        a.formal_params.extend(xtraps)
                        a.formal_returns.extend(xtrars)
                        subst = dict((x.drop_prefix('fml:').rep,x.rep) for x in (xtraps + xtrars))
                        a.args[1] = ivy_ast.subst_prefix_atoms_ast(a.args[1],subst,None,None)
    if __debug__: xtracer.trace("compiler.InferParameters EXIT")

def check_instantiations(mod,decls):
    if __debug__: xtracer.trace("compiler.CheckInstantiations ENTER")
    schemata = set()
    for decl in decls.decls:
        if isinstance(decl,ivy_ast.SchemaDecl):
            for inst in decl.args:
                schemata.add(inst.defines())
    for decl in decls.decls:
        if isinstance(decl,ivy_ast.InstantiateDecl):
            for instantiation in decl.args:
                pref, inst = instantiation.args
                if inst.relname not in schemata:
                    raise IvyError(inst,"{} undefined in instantiation".format(inst.relname))
    if __debug__: xtracer.trace("compiler.CheckInstantiations EXIT")


def create_sort_order(mod):
    if __debug__: xtracer.trace("compiler.CreateSortOrder ENTER")
    arcs = [(x,s) for s in mod.sort_order for x in im.sort_dependencies(mod,s,with_variants=False)]
    # # Remove arcs for recursive types. The first variant can't be recursive.
    # variant_of = set((x.name,y) for y,l in mod.variants.iteritems() for x in l[1:])
    # arcs = [a for a in arcs if a in variant_of]
    # do nothing if sccs already sorted
    number = dict((x,i) for i,x in enumerate(mod.sort_order))
    if all(x == 'bool' or number[x] < number[y] for x,y in arcs):
        if __debug__: xtracer.trace("compiler.CreateSortOrder EXIT")
        return
    m = defaultdict(set)
    for x,y in arcs:
        m[x].add(y)
    sccs = tarjan(m)
    # remove trivial sccs
    sccs = [scc for scc in sccs if len(scc) > 1 or scc[0] in im.sort_dependencies(mod,scc[0])]
    if len(sccs) > 0:
        raise iu.IvyError(None,'these sorts form a dependency cycle: {}.'.format(','.join(sccs[0])))
    mod.sort_order = iu.topological_sort(mod.sort_order,arcs)
    if __debug__: xtracer.trace("compiler.CreateSortOrder EXIT")


def tarjan_arcs(arcs,notriv=True):
    m = defaultdict(set)
    for x,y in arcs:
        m[x].add(y)
    sccs = tarjan(m)
    if notriv:
        sccs = [scc for scc in sccs if len(scc) > 1 or scc[0] in m[scc[0]]]
    return sccs
        

def get_symbol_dependencies(mp,res,t):
    for s in lu.used_symbols_ast(t):
        if s not in res:
            res.add(s)
            if s in mp:
                get_symbol_dependencies(mp,res,mp[s])


# Attach the labeled proofs to their properties. A proof can also attach to an isolate.

def attach_proofs(mod):
    if __debug__: xtracer.trace("compiler.AttachProofs ENTER")
    m = dict((lf.label.rep,lf) for lf in mod.labeled_props if lf.label is not None)
    m.update((lf.label.rep,lf) for lf in mod.labeled_conjs if lf.label is not None)
    used = set()
    pfs = mod.proofs
    mod.proofs = []
    for pf in pfs:
        if pf[0].formula is not None:
            mod.proofs.append(pf)
            used.add(pf[0].label.rep)
    for pf in pfs:
        if pf[0].formula is None:
            lab = pf[0].label.rep
            if lab in used:
                raise IvyError
            used.add(lab)
            if lab in m:
                mod.proofs.append((m[lab],pf[1]))
            elif lab in mod.isolates:
                mod.isolate_proofs[lab] = pf[1]
            else:
                raise IvyError
    if __debug__: xtracer.trace("compiler.AttachProofs EXIT")


def check_definitions(mod):
    if __debug__: xtracer.trace("compiler.CheckDefinitions ENTER")

    # get the definitions that have no dependence on proofs

    stale = set()
    props = mod.labeled_props
    mod.labeled_props = []
    with_proofs = set()
    for lf,pf in mod.proofs:
        with_proofs.add(lf.id)
    for prop in props:
        if isinstance(prop.formula,ivy_logic.Definition):
            if prop.id not in with_proofs:
                if not any(s in stale for s in lu.used_symbols_ast(prop.formula)):
                    mod.definitions.append(prop)
                    continue
            stale.add(prop.formula.defines())
        mod.labeled_props.append(prop)
    # print "definitions:"
    # for prop in mod.definitions:
    #     print prop
    # print "props:"
    # for prop in mod.labeled_props:
    #    print prop


    defs = dict()
    def checkdef(sym,lf):
        if sym in defs:
            raise IvyError(lf,'redefinition of {}\n{} from here'.format(sym,defs[sym].lineno))
        if slv.solver_name(sym) == None:
            raise IvyError(lf,'definition of interpreted symbol {}'.format(sym))
        defs[sym] = lf
    for ldf in mod.definitions:
        checkdef(ldf.formula.defines(),ldf)
    for ldf in mod.native_definitions:
        checkdef(ldf.formula.defines(),ldf)
    for ldf,term in mod.named:
        checkdef(term.rep,ldf)

    # check that no actions interfere with axioms or definitions

    if __debug__: xtracer.trace("compiler.ActionInterferenceCheck ENTER version=%s" % iu.get_string_version())
    if iu.version_le("1.7",iu.get_string_version()):
        side_effects = dict()
        for action in list(mod.actions.values()):
            for sub in action.iter_subactions():
                for s in sub.modifies():
                    if s not in side_effects:
                        side_effects[s] = sub

        if __debug__:
            all_keys = list(mod.actions.keys())
            xtracer.trace("compiler.ActionInterferenceCheck allKeys=%d keys=%s" % (len(all_keys), ",".join(all_keys)))
            for name,actval in mod.actions.items():
                mods = list(actval.iter_subactions())
                mod_syms = set()
                for sub in mods:
                    mod_syms.update(sub.modifies())
                # Instrumentation: always trace DETAIL for divergent actions
                if mod_syms:
                    sym_names = sorted(str(s) for s in mod_syms)
                    if __debug__: xtracer.trace("compiler.ActionInterferenceCheck action=%s modifies=%d syms=%s" % (name, len(mod_syms), ",".join(sym_names)))

        mp = dict((lf.formula.defines(),lf.formula.rhs()) for lf in mod.definitions)
        if not opt_mutax.get():
            for lf in mod.labeled_axioms:
                if not lf.temporal:
                    deps = set()
                    get_symbol_dependencies(mp,deps,lf.formula)
                    for s in deps:
                        if s in side_effects:
                            if __debug__: xtracer.trace("compiler.ActionInterferenceCheck FAIL axiom immutable sym=%s" % str(s))
                            raise IvyError(side_effects[s],'immutable symbol assigned. \n{} info: symbol is used in axiom here'.format(lf.lineno))

        for lf in mod.definitions:
            s = lf.formula.lhs().rep
            if s in side_effects:
                if __debug__: xtracer.trace("compiler.ActionInterferenceCheck FAIL defn immutable key=%s" % str(s))
                raise IvyError(side_effects[s],'immutable symbol assigned. \n{} Symbol is defined here'.format(lf.lineno))
    if __debug__: xtracer.trace("compiler.ActionInterferenceCheck EXIT")
                
    arcs = [(d.formula.defines(),x)
            for d in mod.definitions
            for x in lu.symbols_ilu_ast(d.formula.args[1])]
    dmap = dict((d.formula.defines(),d) for d in mod.definitions)
    pmap = dict((lf.id,p) for lf,p in mod.proofs)
    sccs = tarjan_arcs(arcs)
    from . import ivy_proof
    prover = ivy_proof.ProofChecker(mod.labeled_axioms,[],mod.schemata)
    for scc in sccs:
        if len(scc) > 1:
            raise iu.IvyError(None,'these definitions form a dependency cycle: {}'.format(','.join(scc)))
        d = dmap[scc[0]]
        if d.id not in pmap:
            raise iu.IvyError(d,'definition of {} requires a recursion schema'.format(d.formula.defines()))
        prover.admit_definition(d,pmap[d.id])
    if __debug__: xtracer.trace("compiler.CheckDefinitions EXIT")


# Take a goal with premises and convert it to an implication. To do this, we have to skolemize
# the bound vocabulary in the goal. Also, any definitions in the proof goal are made global
# so that the fragment checker recognizes them as macros.
# TODO: possibly this should be in ivy_proof
# TODO: existential second-order quantifiers are changed to universal, which is sound
# but not complete, and not very transparent. Maybe we should throw an error.
# TODO: First-order quantifiers in shemata are handled like second-order quantifiers!

def theorem_to_property(prop,pos=True):
    if isinstance(prop.formula,ivy_ast.SchemaBody):
        vocab = ip.goal_vocab(prop)
        match = dict()
        for sort in vocab.sorts:
            if sort.name in ivy_logic.sig.sorts:
                newname = unused_name_with_base(sort.name,ivy_logic.sig.sorts)
                newsort = UninterpretedSort(newname)
                ivy_logic.add_sort(newsort)
                match[sort] = newsort
            else:
                ivy_logic.add_sort(sort)
        for sym in vocab.symbols:
            if sym.name in ivy_logic.sig.symbols:
                newname = unused_name_with_base(sym.name,ivy_logic.sig.symbols)
                newsym = ivy_logic.Symbol(newname,sym.sort)
                newsym = ip.apply_match_func(match,newsym)
                ivy_logic.add_symbol(newsym.name,newsym.sort)
                match[sym] = newsym
            else:
                ivy_logic.add_symbol(sym.name,sym.sort)
        if match:
            prop = ip.apply_match_goal(match,prop,ip.apply_match_alt)
        prems = []
        for x in prop.formula.prems():
            if isinstance(x,ivy_ast.LabeledFormula):
                if x.definition:
                    im.module.definitions.append(prop_to_def(x))
                else:
                    if not x.explicit:
                        prems.append(theorem_to_property(x,False).formula)
        conc = prop.formula.conc()
        if isinstance(conc,ivy_logic.Definition):
            raise iu.IvyError(prop,"definitional subgoal must be discharged")
        fmla = ivy_logic.Implies(ivy_logic.And(*prems),conc) if prems else conc
#        if vocab.variables:
#            fmla = ivy_logic.ForAll(vocab.variables,fmla)
        res = ivy_ast.LabeledFormula(prop.label,fmla)
        res.lineno = prop.lineno
        return res
    return prop

def prop_to_def(lf):
    return lf.clone([lf.label,ivy_logic.Definition(*lf.formula.args[0].args)])

# re-order the props so specification props comes afer other props in object

def reorder_props(mod,props):
    if __debug__: xtracer.trace("compiler.ReorderProps ENTER")
    specprops = defaultdict(list)
    iprops = []
    for prop in props:
        if iu.compose_names(prop.name,'spec') in mod.attributes:
            p,c = parent_child_name(prop.name)
            specprops[p].append(prop)
            iprops.append((prop,))  # a placeholder
        else:
            iprops.append(prop)
    rprops = []
    for prop in reversed(iprops):
        name = prop[0].name if isinstance(prop,tuple) else prop.name
        things = []
        while name != 'this':
            name,c = parent_child_name(name)
            if name in specprops:
                things.extend(specprops[name])
                del specprops[name]
        rprops.extend(reversed(things))
        if not isinstance(prop,tuple):
            rprops.append(prop)
    rprops.reverse()
    if __debug__: xtracer.trace("compiler.ReorderProps EXIT")
    return rprops


def create_constructor_schemata(mod):
    if __debug__: xtracer.trace("compiler.CreateConstructorSchemata ENTER")
    from . import ivy_proof
    for sortname,destrs in mod.sort_destructors.items():
        if any(len(f.sort.dom) > 1 for f in destrs):
            continue # TODO: higher-order constructors!
        sort = ivy_logic.find_sort(sortname)
        #sort = destrs[0].sort.dom[0]
        Y = ivy_logic.Variable('Y',sort)
        eqs = [ivy_logic.Equals(f(Y),ivy_logic.Variable('X'+str(n),f.sort.rng)) for n,f in enumerate(destrs)]
        fmla = ivy_logic.Exists([Y],ivy_logic.And(*eqs))
        name = ivy_ast.Atom(iu.compose_names(sortname,'constr'),[])
        sch = ivy_ast.SchemaBody(fmla)
        sch.lineno = None
        sch.instances = []
        goal = ivy_ast.LabeledFormula(name,sch)
        goal.lineno = None
        mod.schemata[name.relname] = goal

        for cons in mod.sort_constructors[sortname]:
            dom = cons.sort.dom
            if len(dom) != len(destrs):
                raise iu.IvyError(cons,"Constructor {} has wrong number of arguments (got {}, expecting {})".format(cons,len(dom),len(destrs)))
            for x,y in zip(dom,destrs):
                if len(y.sort.dom) != 1:
                    raise iu.IvyError(cons,"Cannot define constructor {} for type {} because field {} has higher type".format(cons,sortname,y))
                if x.name != y.sort.rng.name:
                    raise iu.IvyError(cons,"In constructor {}, argument {} has wrong type (expecting {}, got {})".format(cons,y.sort.rng,x))
            xvars = [ivy_logic.Variable('X'+str(n),f.sort.rng) for n,f in enumerate(destrs)]
            Y = cons(*xvars)
            eqs = [ivy_logic.Equals(f(Y),ivy_logic.Variable('X'+str(n),f.sort.rng)) for n,f in enumerate(destrs)]
            fmla = ivy_logic.And(*eqs)
            name = ivy_ast.Atom(iu.compose_names(cons.name,'constr'),[])
            sch = ivy_ast.SchemaBody(fmla)
            sch.lineno = None
            sch.instances = []
            goal = ivy_ast.LabeledFormula(name,sch)
            goal.lineno = None
            mod.schemata[name.relname] = goal

    for sortname,conss in mod.sort_constructors.items():
        for cons in conss:
            if sortname not in mod.sort_destructors:
                raise iu.IvyError(cons,"Cannot define constructor {} for type {} because {} is not a structure type".format(cons,sortname,sortname))
    if __debug__: xtracer.trace("compiler.CreateConstructorSchemata EXIT")


def fix_constructors(mod):
    if __debug__: xtracer.trace("compiler.FixConstructors ENTER")
    for sortname,destrs in mod.sort_destructors.items():
        if any(len(f.sort.dom) > 1 for f in destrs):
            continue # TODO: higher-order constructors!
        sort = ivy_logic.find_sort(sortname)
        #sort = destrs[0].sort.dom[0]
        Y = ivy_logic.Variable('Y',sort)

        new_cons = []
        for cons in mod.sort_constructors[sortname]:
            dom = cons.sort.dom
            if len(dom) == 0 and len(destrs) > 0:
                new_dom = [f.sort.rng for n,f in enumerate(destrs)]
                xvars = [ivy_logic.Variable('X'+str(n),f.sort.rng) for n,f in enumerate(destrs)]
                ivy_logic.remove_symbol(cons)
                ivy_logic.add_symbol(cons.name,ivy_logic.FunctionSort(*(new_dom + [cons.sort.rng])))
                cons = ivy_logic.find_symbol(cons.name)
            new_cons.append(cons)
        mod.sort_constructors[sortname] = new_cons
    if __debug__: xtracer.trace("compiler.FixConstructors EXIT")

def apply_assert_proof(prover,self,pf):
    cond = self.args[0]
    goal = cond if isinstance(cond,ivy_ast.LabeledFormula) else ivy_ast.LabeledFormula(None,cond)
    cond = ip.goal_conc(goal)
    goal.lineno = self.lineno
    subgoals = prover.get_subgoals(goal,pf)
    subgoals = list(map(theorem_to_property,subgoals))
    assm = AssumeAction(ivy_logic.close_formula(cond))
    assm.lineno = self.lineno
    sgas = [ia.SubgoalAction(sg.formula) for sg in subgoals]
    for sga in sgas:
        sga.kind = type(self)
    asrt = Sequence(*(sgas + [assm]))
    asrt.lineno = self.lineno
    for x,y in zip(asrt.args,subgoals):
        if hasattr(y,'lineno'):
            x.lineno = y.lineno
    return asrt

def apply_assert_proofs(mod,prover):
    def recur(self):
        if not isinstance(self,Action):
            return self
        if isinstance(self,AssertAction):
            if len(self.args) > 1:
                if option_verifying:
                    return apply_assert_proof(prover,self,self.args[1])
                return self.clone(self.args[:1])
            return self
        if isinstance(self,WhileAction):
            if len(self.args) > 2:
                new_invars = []
                for a in self.args[2:]:
                    r = recur(a)
                    if isinstance(r,Sequence):
                        new_invars.extend(r.args)
                    else:
                        new_invars.append(r)
                return self.clone(list(map(recur,self.args[0:2])) + new_invars)
        if isinstance(self,LocalAction):
            with ivy_logic.WithSymbols(self.args[0:-1]):
                return self.clone(list(map(recur,self.args)))
        return self.clone(list(map(recur,self.args)))
    action_idx = 0
    for actname in list(mod.actions.keys()):
        if __debug__: xtracer.trace("compiler.apply_assert_proofs action[%d]=%s" % (action_idx, actname))
        action_idx += 1
        action = mod.actions[actname]
        with ivy_logic.WithSymbols(list(set(action.formal_params+action.formal_returns))):
            mod.actions[actname] = action.copy_formals(recur(action))
    
def check_properties(mod):
    if __debug__: xtracer.trace("compiler.CheckProperties ENTER")
    props = reorder_props(mod,mod.labeled_props)
    
    

    mod.labeled_props = []
    pmap = dict((lf.id,p) for lf,p in mod.proofs)
    nmap = dict((lf.id,n) for lf,n in mod.named)

    # Give empty proofs to theorems without proofs. 
    for prop in props:
        if prop.id not in pmap and isinstance(prop.formula,ivy_ast.SchemaBody):
            emptypf = ivy_ast.ComposeTactics()
            emptypf.lineno = prop.lineno
            pmap[prop.id] = emptypf

    def named_trans(prop):
        if prop.id in nmap:
            name = nmap[prop.id]
            fmla = ivy_logic.drop_universals(prop.formula)
            v = list(fmla.variables)[0]
            fmla = fmla.body
            fmla = lu.substitute_ast(fmla,{v.name:name})
            fmla = ivy_logic.drop_universals(fmla)
            prop = prop.clone_with_fresh_id([prop.label,fmla])
        return prop
            
    from . import ivy_proof
    prover = ivy_proof.ProofChecker(mod.labeled_axioms,mod.definitions,mod.schemata)

    for prop in props:
        _plbl = str(prop.label) if prop.label else ""
        _has_pf = prop.id in pmap
        if __debug__: xtracer.trace("compiler.CheckProperties.classify label=%s id=%d temporal=%s hasPf=%s" % (_plbl, prop.id, bool(prop.temporal), bool(_has_pf)))
        if prop.temporal:
            if __debug__: xtracer.trace("compiler.CheckProperties.classify label=%s -> props (temporal)" % _plbl)
            mod.labeled_props.append(prop)
        elif prop.id in pmap:
            subgoals = prover.admit_proposition(prop,pmap[prop.id])
            if __debug__: xtracer.trace("compiler.CheckProperties.classify label=%s subgoals=%d" % (_plbl, len(subgoals)))
            if not isinstance(prop.formula,ivy_logic.Definition):
                prop = named_trans(prop)
                prover.axioms[-1] = prop
                prover.schemata[prop.name] = prop
            if len(subgoals) == 0:
                if not isinstance(prop.formula,ivy_ast.SchemaBody):
                    if isinstance(prop.formula,ivy_logic.Definition):
                        if __debug__: xtracer.trace("compiler.CheckProperties.classify label=%s -> definitions (proved, 0 subgoals, def)" % _plbl)
                        mod.definitions.append(prop)
                    else:
                        if __debug__: xtracer.trace("compiler.CheckProperties.classify label=%s -> axioms (proved, 0 subgoals)" % _plbl)
                        mod.labeled_axioms.append(prop)
                else:
                    if __debug__: xtracer.trace("compiler.CheckProperties.classify label=%s -> schemata (proved, 0 subgoals, schema)" % _plbl)
                    mod.schemata[prop.label.relname] = prop
            else:
                if __debug__: xtracer.trace("compiler.CheckProperties.classify label=%s -> props (proved, %d subgoals)" % (_plbl, len(subgoals)))
                subgoals = list(map(theorem_to_property,subgoals))
                lb = ivy_ast.Labeler()
                for g in subgoals:
                    if prop.label is None:
                        raise IvyError(prop,'Properties with subgoals must be labeled')
                    label = ia.compose_atoms(prop.label,lb())
                    mod.labeled_props.append(g.clone([label,g.formula]))
                if not isinstance(prop.formula,ivy_ast.SchemaBody):
                    if isinstance(prop.formula,ivy_logic.Definition):
                        mod.definitions.append(prop)
                    else:
                        mod.labeled_props.append(prop)
                else:
                    mod.schemata[prop.label.relname] = prop
            mod.subgoals.append((prop,subgoals))
        #     from ivy_l2s import l2s
        #     print "=================" + "\n" * 10
        #     l2s(mod, prop)
        else:
            # if isinstance(prop.formula,ivy_ast.SchemaBody):
            #     prover.schemata[prop.label.relname] = prop.formula
            #     prop = theorem_to_property(prop)
            if __debug__: xtracer.trace("compiler.CheckProperties.classify label=%s -> props (no proof)" % _plbl)
            if isinstance(prop.formula,ivy_logic.Definition):
                mod.definitions.append(prop)
            else:
                mod.labeled_props.append(prop)
                if prop.id in nmap:
                    nprop = named_trans(prop)
                    mod.labeled_props.append(nprop)
                    mod.subgoals.append((nprop,[prop]))
                    prover.admit_proposition(nprop,ivy_ast.ComposeTactics())
                else:
                    prover.admit_proposition(prop,ivy_ast.ComposeTactics())
    mod.canon_snapshot("before-apply-assert-proofs")
    apply_assert_proofs(mod,prover)
    mod.canon_snapshot("after-apply-assert-proofs")
    if __debug__: xtracer.trace("compiler.CheckProperties EXIT")

option_verifying = True
def set_verifying(tv):
    global option_verifying
    option_verifying = tv
    
def ivy_compile_theory(mod,decls,**kwargs):
    if __debug__: xtracer.trace("compiler.IvyCompileTheory ENTER")
    if __debug__: sig_check("IvyCompileTheory")
    IvyDomainSetup(mod)(decls)
    if __debug__: sig_check("IvyCompileTheory.exit")
    if __debug__: xtracer.trace("compiler.IvyCompileTheory EXIT")
    
def ivy_compile_theory_from_string(mod,theory,sortname,**kwargs):
    if __debug__: xtracer.trace("compiler.IvyCompileTheoryFromString ENTER sortname=%s" % sortname)
    import io
    sio = io.StringIO(theory)
    module = read_module(sio)
    ivy = Ivy()
    inst_mod(ivy,module,None,{'t':sortname},dict())
    ivy_compile_theory(mod,ivy,**kwargs)
    if __debug__: xtracer.trace("compiler.IvyCompileTheoryFromString EXIT")

def compile_theory(mod,sortname,theoryname,**kwargs):
    if __debug__: xtracer.trace("compiler.CompileTheory ENTER sortname=%s theoryname=%s" % (sortname, theoryname))
    theory = ith.get_theory_schemata(theoryname)
    if theory is not None:
        ivy_compile_theory_from_string(mod,theory,sortname,**kwargs)
    if __debug__: xtracer.trace("compiler.CompileTheory EXIT")

def compile_theories(mod,**kwargs):
    if __debug__: xtracer.trace("compiler.CompileTheories ENTER")
    for name,value in mod.sig.interp.items():
        if name in mod.sig.sorts and (isinstance(value,str) or isinstance(value,ivy_logic.RangeSort)):
            theory = th.get_theory_schemata(value)
            ivy_compile_theory_from_string(mod,theory,name,**kwargs)
    if __debug__: xtracer.trace("compiler.CompileTheories EXIT")

# Here, we infer for each conjecture the set of actions that must
# preserve it. A conjecture must be preserved by any action that is
# exported by its most nearly containing isolate. This is only needed
# from version 1.7, where we introduce the concept of object
# invariant.
#

def create_conj_actions(mod):
    if __debug__: xtracer.trace("compiler.CreateConjActions ENTER")
    if iu.version_le(iu.get_string_version(),"1.6"):
        if __debug__: xtracer.trace("compiler.CreateConjActions EXIT")
        return
    # for each isolate, find the exported actions
    myexports = dict()
    objects = defaultdict(list)
    cg = mod.call_graph()
    for ison,isol in mod.isolates.items():
        myexports[isol.name()] = iso.get_isolate_exports(mod,cg,isol)
        for x in isol.verified():
            objects[x.rep].append(isol)
    for conj in mod.labeled_conjs:
        lbl = conj.label.rep
        while lbl != 'this' and lbl not in objects:
            lbl,_ = iu.parent_child_name(lbl)
        if lbl == 'this':
            actions = set([exp.exported() for exp in mod.exports])
        else:
            actions = set()
            for isol in objects[lbl]:
                actions.update(myexports[isol.name()])
        mod.conj_actions[conj.label.rep] = actions
    action_isos = defaultdict(set)

    # check that object invariants are used correctly
    
    if iso.do_check_interference.get():
        for ison,actions in myexports.items():
            for action in actions:
                action_isos[action].add(ison)
        for ison,isol in mod.isolates.items():
            memo = set()
            conjs = iso.get_isolate_conjs(mod,isol,verified=False)
            exports = myexports[ison]
            roots = set(iu.reachable(exports,lambda x: cg[x]))
            for conj in conjs:
                actions = mod.conj_actions[conj.label.rep]
                for action in actions:
                    for ison1 in action_isos[action]:
                        if ison1 != ison and action not in memo:
                            memo.add(action)
                            if action in roots and action not in exports:
                                for victim in exports:
                                    if action in set(iu.reachable([victim],lambda x: cg[x])) and action != victim:
                                        raise IvyError(conj, "isolate {} depends on invariant {} which might not hold because action {} is called from within action {}, which invalidates the invariant.".format(ison,conj.label.rep,victim,action))
    if __debug__: xtracer.trace("compiler.CreateConjActions EXIT")

# This handles the modular semantics of temporal properties. Each
# temporal operator is labeled with an isolate (which could be
# 'this'). If the user does not give a label for an operator, then the
# isolate in which the property is verified is used. For now, it is an
# error if the property is verified in more than one
# isolate. Additionally, we label each action with the list of
# isolates in which is is present [1]. The modular semantics says that
# a temporal operator refers only to time instants when execution is
# *not* in the indicated isolate. In particular, this means that
# "temporal property globally p" has the same semantics as "invariant
# p".
#
# TODO: this does not handle temporal axioms.

def handle_temporals(mod):
    if __debug__: xtracer.trace("compiler.HandleTemporals ENTER")
    # imap = iso.get_isolate_map(mod,verified=True,present=False)
    # new_props = []
    # for prop in mod.labeled_props:
    #     isonames = imap[prop.name]
    #     if prop.temporal:
    #         assert len(isonames) > 0  # should at least be verified in isolate 'this'!
    #         if len(isonames) > 1:
    #             raise IvyError(prop,'Temporal property belongs to more than one isolate: {}'.format(','.join(str(x) for x in isonames)))
    #         new_props.append(prop.clone([prop.label,ivy_logic.label_temporal(prop.formula,isonames[0])]))
    #     else:
    #         new_props.append(prop)
            
    # for prop,proof in mod.proofs:
    #     if prop.temporal:
    #         isonames = imap[prop.name]
    #         add_labels_to_proof(proof,isonames)
    # mod.labeled_props = new_props
    imap = iso.get_isolate_map(mod,verified=True,present=True)
    for actname,action in mod.actions.items():
        action.labels = sorted(imap[actname]) ## jea added: sorted().
    if __debug__: xtracer.trace("compiler.HandleTemporals EXIT")

def add_labels_to_proof(proof,labels):
    if isinstance(proof,ivy_ast.ComposeTactics):
        return proof.clone([add_labels_to_proof(pf,labels) for pf in proof.args])
    if isinstance(proof,ivy_ast.IfTactic):
        return proof.clone([proof.args[0]] + [add_labels_to_proof(pf,labels) for pf in proof.args[1:]])
    if isinstance(proof,ivy_ast.TacticTactic):
        proof.labels = list(labels)
    return proof
        
        
def add_action_label(action,label):
    if not hasattr(action,'labels'):
        action.labels = []
    action.labels.append(label)

def show_call_graph(cg):
    for caller,callees in cg.items():
        print('{} -> {}'.format(caller,','.join(callees)))
    
def ivy_compile(decls,mod=None,create_isolate=True,**kwargs):
    mod = mod or im.module
    if __debug__: xtracer.trace("compiler.IvyCompile ENTER decls=%d" % len(decls.decls))
    with mod.sig:
        check_instantiations(mod,decls)
        for name in decls.defined:
            mod.add_to_hierarchy(name)
        for decl in decls.decls:
            for attribute in decl.attributes:
                for dfs in decl.defines():
                    name = dfs[0]
                    mod.attributes[iu.compose_names(name,attribute)] = decl.common if decl.common is not None and attribute == "common" else "yes"
        #        infer_parameters(decls.decls)
        if __debug__: xtracer.trace("compiler.CollectActions ENTER decls=%d" % len(decls.decls))
        _collected_actions = collect_actions(decls.decls)
        with TopContext(_collected_actions):
            if __debug__: xtracer.trace("compiler.CollectActions EXIT actions=%d" % len(_collected_actions))
            n_decls = len(decls.decls)
            tail_start = max(0, n_decls - 20)
            for i in range(tail_start, n_decls):
                if __debug__: xtracer.trace("compiler.DeclList i=%d name=%s" % (i, decls.decls[i].name()))
            if __debug__: xtracer.trace("compiler.DomainSetup ENTER")
            IvyDomainSetup(mod)(decls)
            fix_constructors(mod)
            if __debug__: xtracer.trace("compiler.DomainSetup EXIT")
            if __debug__: mod.canon_snapshot("after-domain-setup")
            if __debug__: xtracer.trace("compiler.ConjSetup ENTER")
            IvyConjectureSetup(mod)(decls)
            if __debug__: xtracer.trace("compiler.ConjSetup EXIT")
            if __debug__: mod.canon_snapshot("after-conj-setup")
            if __debug__: xtracer.trace("compiler.ARGSetup ENTER")
            if __debug__:
                action_count = sum(1 for d in decls.decls if d.name() == 'action')
                xtracer.trace("compiler.ARGSetup.ProcessDecls ENTER total_decls=%d action_decls=%d" % (len(decls.decls), action_count))
            IvyARGSetup(mod)(decls)
            if __debug__: xtracer.trace("compiler.ARGSetup EXIT mod.Mixins.Len=%d" % len(mod.mixins))
            if __debug__: mod.canon_snapshot("after-arg-setup")
        mod.macros = decls.macros
        # progress properties are not state symbols -- remove from sig
        for p in mod.progress:
            remove_symbol(p.defines())
        mod.type_check()
        # try instantiating all the actions to type check them
        for name,action in mod.actions.items():
#            print "checking: {} = {}".format(name,action)
            type_check_action(action,mod)
            if not hasattr(action,'lineno'):
                print("no lineno: {}".format(name))
            assert hasattr(action,'formal_params'), action
    
        # from version 1.7, there is always one global "isolate"
        if not iu.version_le(iu.get_string_version(),"1.6"):
            if 'this' not in mod.isolates:
                isol = ivy_ast.IsolateDef(ivy_ast.Atom('this'),ivy_ast.Atom('this'))
                isol.with_args = 0
                mod.isolates['this'] = isol
            # print "actions:"
            # for x,y in mod.actions.iteritems():
            #     print iu.pretty("action {} = {}".format(x,y))

        # find all the global roots and add them to the "with" lists of all of the isolates

        global_objects = []
        for name in im.module.attributes:
            p,c = iu.parent_child_name(name)
            if c == 'global' and p not in im.module.aliases:
                pp,pc = iu.parent_child_name(p)
                if pp == 'this' or iu.compose_names(pp,'global') not in im.module.attributes:
                    global_objects.append(ivy_ast.Atom(p,[]))
        for iso1 in list(im.module.isolates.values()): ## fixed, was iso in list...
            iso1.args += tuple(global_objects)
            iso1.with_args += len(global_objects)
                    

        create_sort_order(mod)
        create_constructor_schemata(mod)
        attach_proofs(mod)
        check_definitions(mod)
        check_properties(mod)
        create_conj_actions(mod)
        handle_temporals(mod)
        if create_isolate:
            iso.create_isolate(isolate.get(),mod,**kwargs)
            im.module.labeled_axioms.extend(im.module.labeled_props)
            im.module.theory_context().__enter__()
        if __debug__: xtracer.trace("compiler.IvyCompile EXIT mod.Mixins.Len=%d" % len(mod.mixins))


def clear_rules(modname):
    import sys
    if modname in sys.modules:
        d = sys.modules[modname].__dict__
    else:
        d = sys.modules['ivy.' + modname].__dict__
    for s in list(d):
        if s.startswith('p_'):
            del d[s]

def read_module(f,nested=False):
    #print('python side xtracer.enabled=', xtracer.enabled, '\n', file=sys.stdout, flush=True)
    if __debug__: xtracer.trace("init.ReadModule ENTER file=%s nested=%s" % (xtracer.normalize_filename(f.name) if hasattr(f,'name') else '?', str(nested)))
    from . import ivy_logic_parser
    from . import ivy_parser
    header = f.readline()
    s = '\n' + f.read() # newline at beginning to preserve line numbers
    header = str.strip(header)
    if header.startswith('#lang ivy'):
        version = header[len('#lang ivy'):]
        if version.strip() != '':
            old_version = iu.get_string_version()
            iu.set_string_version(version)
            if version != old_version:
                if nested:
                    raise IvyError(None,'#lang ivy{} expected in included file'.format(old_version)) 
    #            print "version: {}, old_version: {}".format(version,old_version)
                clear_rules('ivy_logic_parser')
                clear_rules('ivy_parser')
                importlib.reload(ivy_logic_parser)
                importlib.reload(ivy_parser)
        ivy_parser.importer = import_module
        decls = parse(s,nested)
    elif header == '//lang dafny1':
        decls = dc.parse_to_ivy(s)
    else:
        err = IvyError(None,'file must begin with "#lang ivyN.N"')
        err.lineno = Location(iu.filename,1)
#        if iu.filename:
#            err.filename = iu.filename
        raise err
    if __debug__: xtracer.trace("init.ReadModule EXIT file=%s decls=%d" % (xtracer.normalize_filename(f.name) if hasattr(f,'name') else '?', len(decls.decls) if hasattr(decls,'decls') else 0))
    return decls

def import_module(name):
    if __debug__: xtracer.trace("init.ImportModule ENTER name=%s" % name)
    fname = name + '.ivy'
    try: 
        f = open(fname,'r')
    except Exception:
        fname = os.path.join(iu.get_std_include_dir(),fname)
        try:
            f = open(fname,'r')
        except Exception:
            raise IvyError(None,"module {} not found in current directory or module path".format(name))
    with iu.SourceFile(fname):
        mod = read_module(f,nested=True)
    if __debug__: xtracer.trace("init.ImportModule EXIT name=%s" % name)
    return mod

def ivy_load_file(f,**kwargs):
    decls = read_module(f)
    ivy_compile(decls,**kwargs)

def ivy_from_string(string,**kwargs):
    import io
    sio = io.StringIO(string)
    ivy_load_file(sio,**kwargs)
    return ivy_new()
