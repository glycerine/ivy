
# Copyright (c) Microsoft Corporation. All Rights Reserved.
#
from .ivy_concept_space import NamedSpace, ProductSpace, SumSpace
from .ivy_ast import *
from .ivy_actions import AssumeAction, AssertAction, EnsuresAction, SetAction, AssignAction, VarAction, HavocAction, IfAction, AssignFieldAction, NullFieldAction, CopyFieldAction, InstantiateAction, CallAction, LocalAction, LetAction, Sequence, UpdatePattern, PatternBasedUpdate, SymbolList, UpdatePatternList, Schema, ChoiceAction, NativeAction, WhileAction, Ranking, RequiresAction, EnsuresAction, CrashAction, ThunkAction, DebugAction, check_unprovable
from .ivy_lexer import *
from . import ivy_utils as iu
from . import xtracer
from . import canon_ast as _canon_ast
from . import logic_sexp as _logic_sexp
import copy
from collections import defaultdict

_canon_ast.install()
_logic_sexp.install()


import ply.yacc as yacc
import string

if not (iu.get_numeric_version() <= [1,2]):

    if not (iu.get_numeric_version() <= [1,6]):
        precedence = (
            ('left', 'SEMI'),
            ('left', 'GLOBALLY', 'EVENTUALLY'),
            ('left', 'ARROW', 'IFF'),
            ('left', 'OR'),
            ('left', 'AND'),
            ('left', 'TILDA'),
            ('left', 'EQ','LE','LT','GE','GT','PTO'),
            ('left', 'TILDAEQ'),
            ('left', 'IF'),
            ('left', 'ELSE'),
            ('left', 'COLON'),
            ('left', 'PLUS','MINUS'),
            ('left', 'TIMES','DIV'),
            ('left', 'DOLLAR'),
            ('left', 'OLD'),
            ('left', 'DOT')
        )

    else:
        precedence = (
            ('left', 'SEMI'),
            ('left', 'GLOBALLY', 'EVENTUALLY','WHENFIRST','WHENLAST','WHENNEXT','WHENPREV'),
            ('left', 'IF'),
            ('left', 'ELSE'),
            ('left', 'OR'),
            ('left', 'AND'),
            ('left', 'TILDA'),
            ('left', 'EQ','LE','LT','GE','GT','PTO'),
            ('left', 'TILDAEQ'),
            ('left', 'COLON'),
            ('left', 'PLUS'),
            ('left', 'MINUS'),
            ('left', 'TIMES'),
            ('left', 'DIV'),
            ('left', 'DOLLAR'),
        )

else:

    # This is for versio 1.2 and older, where * is used 
    # in "concept space" descriptions, but not terms

    precedence = (
        ('left', 'SEMI'),
        ('left', 'IF'),
        ('left', 'ELSE'),
        ('left', 'OR'),
        ('left', 'AND'),
        ('left', 'PLUS'),
        ('left', 'TIMES'),
        ('left', 'DIV'),
        ('left', 'TILDA'),
        ('left', 'EQ','LE','LT','GE','GT'),
        ('left', 'TILDAEQ'),
        ('left', 'COLON'),
    )


class ParseError(Exception):
    def __init__(self,lineno,token,message):
        if __debug__: xtracer.trace("parser.__init__ ENTER")
#        print "initializing"
        self.lineno, self.token,self.message = lineno,token,message
        if iu.filename:
            self.filename = iu.filename
    def __repr__(self):
        if __debug__: xtracer.trace("parser.__repr__ ENTER")
        return ( ("{}".format(self.filename) if hasattr(self,'filename') else '')
                 + ("({})".format(self.lineno) if self.lineno != None else '')
                 + (': ' if (hasattr(self,'filename') or self.lineno != None) else '')
                 + ('error: ')
                 + ("token '{}': ".format(self.token) if self.token != None else '')
                 + self.message )
    
class Redefining(ParseError):
    def __init__(self,name,lineno,orig_lineno):
        if __debug__: xtracer.trace("parser.__init__ ENTER")
        msg = 'redefining ' + str(name)
        if orig_lineno != None:
            msg += " (from {})".format(str(orig_lineno)[:-2])
        line = lineno.line if hasattr(lineno,"line") else lineno
        super(Redefining, self).__init__(line,None,msg)

error_list = []

stack = []

def _normalize_filename(f):
    """Replace include directory path with <IVY_INCLUDE> for canonical matching."""
    if f is None:
        return f
    std_dir = iu.get_std_include_dir()
    if std_dir:
        import os.path
        base_dir = os.path.dirname(std_dir)
        if base_dir and not base_dir.endswith(os.sep):
            base_dir += os.sep
        if f.startswith(base_dir):
            return '<IVY_INCLUDE>/' + f[len(base_dir):]
    return f

def get_lineno(p,n):
    #if __debug__: xtracer.trace("parser.get_lineno ENTER")
    return iu.Location(xtracer.normalize_filename(iu.filename),p.lineno(n))

def report_error(error):
    if __debug__: xtracer.trace("parser.report_error ENTER")
#    assert False,error
    error_list.append(error)

def stack_lookup(name):
    if __debug__: xtracer.trace("parser.stack_lookup ENTER")
    for ivy in reversed(stack):
        if name in ivy.modules:
            return ivy.modules[name]
    return None


def stack_action_lookup(name,params=0):
    if __debug__: xtracer.trace("parser.stack_action_lookup ENTER")
    for ivy in reversed(stack):
        if ivy.is_module:
            break
        params += len(ivy.params)
        if name in ivy.actions:
            return ivy.actions[name],params
    return None,0

def inst_mod(ivy,module,pref,subst,vsubst,modname=None,lineno=None):
    if __debug__: xtracer.trace("parser.inst_mod ENTER name=%s" % (modname if modname else ""))
    if __debug__: xtracer.trace("ast.LF.instMod SET_FRESH cfg=global")
    set_always_clone_with_fresh_id(True)
    if pref is not None and pref.rep in vsubst:
        raise iu.IvyError(pref,'instance parameter names must differ from instance name')
    save = ivy.attributes
    ivy.attributes = tuple(x for x in ivy.attributes if x == "common")
    static = module.static.copy()
    for name,dfs in module.defined.items():
        if any((df[1] is TypeDecl) or (df[1] is DestructorDecl) for df in dfs):
            static.add(name)
    def spaa(decl,subst,pref):
        if __debug__: xtracer.trace("parser.spaa ENTER decl=%s pref=%s" % (decl.canon(), (pref.canon() if pref else "nil")))
        if modname is not None and pref is not None and isinstance(decl,ModuleDecl):
            if __debug__: xtracer.trace("parser.spaa substituting with pref.rep='%s'" % pref.rep)
            subst = subst.copy()
            p,c = iu.parent_child_name(modname)
            subst[c] = pref.rep
#        print 'instmod: {} {}'.format(pref,lineno)
        if lineno is not None:
            set_reference_lineno(lineno)
        res = subst_prefix_atoms_ast(decl,subst,pref,module.defined,static=static)
#        print 'instmod done'
        if lineno is not None:
            set_reference_lineno(None)
        if __debug__: xtracer.trace("parser.spaa EXIT res=%s" % res.canon())
        return res
    #if __debug__: xtracer.trace("parser.inst_mod.body name=%s ndecls=%d\n pref=%s" % (modname if modname else "", len(module.decls), pref))
    if __debug__: xtracer.trace("parser.inst_mod.body name=%s ndecls=%d" % (modname if modname else "", len(module.decls)))
    for di, decl in enumerate(module.decls):
        #if __debug__: xtracer.trace("parser.inst_mod.iter name=%s i=%d decl=%s\n pyClass=%s" % (modname if modname else "", di, decl.name(), type(decl).__name__))
        if __debug__: xtracer.trace("parser.inst_mod.iter name=%s i=%d decl=%s" % (modname if modname else "", di, decl.name()))
        dpref = pref.clone([]) if pref is not None and "common" in decl.attributes else pref
        dvsubst = dict() if "common" in decl.attributes else vsubst
        if isinstance(decl,AttributeDecl):
            if dvsubst:
                map1 = distinct_variable_renaming(used_variables_ast(dpref),used_variables_ast(decl))
                vpref = substitute_ast(dpref,map1)
                vvsubst = dict((x,map1[y.rep]) for x,y in dvsubst.items())
                idecl = AttributeDecl(*[x.clone([compose_atoms(vpref,x.args[0]),x.args[1]]) for x in decl.args]) if vpref is not None else decl
                idecl = substitute_constants_ast(idecl,vvsubst)
            else:
                idecl = AttributeDecl(*[x.clone([compose_atoms(dpref,x.args[0]),x.args[1]]) for x in decl.args]) if dpref is not None else decl
        elif dvsubst:
            map1 = distinct_variable_renaming(used_variables_ast(dpref),used_variables_ast(decl))
            vpref = substitute_ast(dpref,map1)
            vvsubst = dict((x,map1[y.rep]) for x,y in dvsubst.items())
            idecl = spaa(decl,subst,vpref)
            idecl = substitute_constants_ast2(idecl,vvsubst)
        else:
            idecl = spaa(decl,subst,dpref)
        if decl.common is not None:
            idecl.common = (pref.rep if decl.common == 'this' else iu.compose_names(pref.rep,decl.common)) if pref is not None else decl.common
        else:
            idecl.common = None
        if isinstance(idecl,ActionDecl):
            for foo in idecl.args:
                if not hasattr(foo.args[1],'lineno'):
                    print('no lineno: {}'.format(foo))
        idecl.attributes = decl.attributes
        #if __debug__: xtracer.trace("parser.inst_mod.declare name=%s\n pyClass=%s origClass=%s" % (idecl.name(), type(idecl).__name__, type(decl).__name__))
        if __debug__: xtracer.trace("parser.inst_mod.declare name=%s" % (idecl.name()))
        if isinstance(idecl,ObjectDecl):
            ivy.declare(idecl)
            ivy.set_object_defined(idecl.args[0].rep,module.get_object_defined(idecl.args[0].rep))
        elif isinstance(idecl,InstantiateDecl):
            old_attrs = ivy.attributes
            ivy.attributes = ivy.attributes + idecl.attributes
            do_insts(ivy,idecl.args)
            ivy.attributes = old_attrs
        else:
            ivy.declare(idecl)
    ivy.attributes = save
    if __debug__: xtracer.trace("parser.inst_mod EXIT name=%s" % (modname if modname else ""))
    if __debug__: xtracer.trace("ast.LF.instMod CLEAR_FRESH cfg=global")
    set_always_clone_with_fresh_id(False)

def do_insts(ivy,insts):
    if __debug__: xtracer.trace("parser.do_insts ENTER")
    others = []
    for instantiation in insts:
        pref, inst = instantiation.args
        defn = stack_lookup(inst.relname)
        if defn:
#            print "instantiating %s" % inst
            if pref != None:
#                ivy.define((pref.rep,inst.lineno))
                ivy.declare(ObjectDecl(pref))
            aparams = inst.args
            fparams = defn.args[0].args
            if len(aparams) != len(fparams):
                raise iu.IvyError(instantiation,"wrong number of arguments to module {}".format(inst.relname))
            subst = dict((x.rep,y.rep) for x,y in zip(fparams,aparams) if not isinstance(y,Variable))
            vsubst = dict((x.rep,y) for x,y in zip(fparams,aparams) if isinstance(y,Variable))
            pvars = set(x.rep for x in pref.args) if pref != None else set()
            for v in list(vsubst.values()):
                if v.rep not in pvars:
                    raise iu.IvyError(instantiation,"variable {} is unbound".format(v))
            module = defn.args[1]
            inst_mod(ivy,module,pref,subst,vsubst,modname=inst.relname,
                     lineno=instantiation.lineno if hasattr(instantiation,"lineno") else
                     inst.lineno if hasattr(instantiation,"lineno") else None )

            if pref is None:
                ivy.objects.update(module.objects)
        else:
            inst.lineno = instantiation.lineno
            assert hasattr(inst,"lineno")
            others.append(instantiation)
    if others:
        ivy.declare(InstantiateDecl(*others))
    if __debug__: xtracer.trace("parser.do_insts EXIT decls=%d" % len(others))

def check_non_temporal(x):
    if __debug__: xtracer.trace("parser.check_non_temporal ENTER")
    assert type(x) is not list
    if type(x) is LabeledFormula:
        check_non_temporal(x.args[1])
        return x
    elif has_temporal(x):
        report_error(IvyError(x,"non-temporal formula expected"))
    else:
        return x

special_attribute = None
parent_object = None
global_attribute = None
common_attribute = None

class Ivy(object):
    def __init__(self):
        if __debug__: xtracer.trace("parser.__init__ ENTER")
        self.decls = []
        self.defined = defaultdict(list)
        self.static = set()
        self.modules = dict()
        self.objects = dict()  # maps object names to "defined" dictionary
        self.macros = dict()
        self.actions = dict()
        self.merkle = xtracer.MerkleState()
        self.included = set()
        self.is_module = False
        self.params = []
        global special_attribute
        global global_attribute
        global common_attribute
        self.attributes = (((special_attribute,) if special_attribute is not None else ()) +
                           ((global_attribute,) if global_attribute is not None else ()) +
                           ((common_attribute,) if common_attribute is not None else ()))
        special_attribute = None
        global_attribute = None
        common_attribute = None
        # if we are the body of an attribute declaration, keep all of the enclosing attributes
        if self.attributes and stack:
            self.attributes = stack[-1].attributes + self.attributes
        # if we are a continuation object, inherent defined symbols from previous declaration
        global parent_object
        if parent_object is not None:
            parent = stack[-1]
            if parent_object == "this":
                defined = parent.defined
            else:
#                if parent_object in parent.defined:
#                    print parent.defined[parent_object]
#                print parent.defined.keys()
                defined = parent.get_object_defined(parent_object)
            if defined is not None:
                self.defined = defined
            parent_object = None
    def __repr__(self):
        return '\n'.join([repr(x) for x in self.decls])
    def canon(self):
        """Canon as a Sequence of declarations, matching Go's ast.Sequence wrapper."""
        from .canon import slice_canon, lineno_fields
        return '(sequence{} stmts:{})'.format(
            lineno_fields(self), slice_canon(list(self.decls)))
    def declare(self,decl,allow_redef=False):
        if __debug__: xtracer.trace("parser.declare ENTER")
        for df in decl.defines():
            self.define(df,allow_redef)
        for df in decl.static():
            self.static.add(df)
        decl.attributes = self.attributes + decl.attributes
        if "common" in decl.attributes:
            for df in decl.defines():
                self.static.add(df[0])
        if "common" in self.attributes and decl.common == None:
            decl.common = 'this'
        self.decls.append(decl)
        if __debug__ and hasattr(decl, 'canon'):
            canonical = decl.canon()
            leaf, root = self.merkle.add_leaf(canonical)
            if xtracer._hash_verbose:
                xtracer.trace("parser.declare HASH leaf=%s root=%s canon=%s" % (leaf, root, canonical))
            else:
                xtracer.trace("parser.declare HASH leaf=%s root=%s" % (leaf, root))
        if isinstance(decl,MacroDecl):
            for d in decl.args:
                self.macros[d.defines()] = d
        if isinstance(decl,ActionDecl):
            for d in decl.args:
                d.attributes = decl.attributes
                self.actions[d.defines()] = d
        if isinstance(decl,ModuleDecl):
            for d in decl.args:
                self.modules[d.defines()] = d
                

    def define(self,df,allow_redef=False):
        if __debug__: xtracer.trace("parser.define ENTER")
        if len(df) == 3:
            name,lineno,cls = df
        else:
            name,lineno = df
            cls = None
        for x in self.defined[name]:
            olineno,ocls = x[0],x[1]
            conflict = ((ocls is not ObjectDecl) if cls is TypeDecl 
                        else (ocls is not TypeDecl) if cls is ObjectDecl else True)
            if conflict:
                if allow_redef:
                    return
                report_error(Redefining(name,lineno,olineno))
        self.defined[name].append((lineno,cls))

    def define_type(self,df):
        if __debug__: xtracer.trace("parser.define_type ENTER")
        name,lineno = df
        if name in self.defined_types:
            report_error(Redefining(name,lineno,self.defined[name]))
        self.defined[name] = lineno

    def get_object_defined(self,name):
        if __debug__: xtracer.trace("parser.get_object_defined ENTER")
        if name in self.defined:
            x = self.defined[name][0]
            if len(x) >= 3:
                return x[2]
        return None

    def set_object_defined(self,name,defined):
        if __debug__: xtracer.trace("parser.set_object_defined ENTER")
#        print ('name: {}, defined: {}'.format(name,defined))
        if defined is not None:
            defined = defaultdict(list,((k,v.copy()) for k,v in defined.items()))
#        print 'set_object_defined: {}'.format(name)
        if name in self.defined:
#            print 'prev: {}'.format(self.defined[name])
            self.defined[name] = [(x[0],x[1],defined) for x in self.defined[name]]

    @property
    def args(self):
        if __debug__: xtracer.trace("parser.args ENTER")
        return []
    def clone(self,args):
        if __debug__: xtracer.trace("parser.clone ENTER")
        return self

    def rewrite(self,rewrite):
        if __debug__: xtracer.trace("parser.rewrite ENTER")
        if isinstance(rewrite,AstRewriteSubstPrefix):
            res = Ivy()
            inst_mod(res,self,None,rewrite.subst,dict())
            return res
        return self

def p_top(p):
    'top :'
    if __debug__: xtracer.trace("parser.p_top ENTER (top)")
    p[0] = Ivy()
    stack.append(p[0])

def p_top_using_symbol(p):
    'top : top USING SYMBOL'
    if __debug__: xtracer.trace("parser.p_top_using_symbol ENTER (top)")
    p[0] = p[1]
    pref = Atom(p[3],[])
    module = importer(p[3])
    for decl in module.decls:
        idecl = subst_prefix_atoms_ast(decl,{},pref,module.defined)
        p[0].declare(idecl)

def p_top_include_symbol(p):
    'top : top INCLUDE SYMBOL'
    if __debug__: xtracer.trace("parser.p_top_include_symbol ENTER (top)")
    p[0] = p[1]
    if not any(p[3] in m.included for m in stack):
        p[0].included.add(p[3])
        pref = Atom(p[3],[])
        pref.lineno = get_lineno(p,2)
        if __debug__: xtracer.trace("parser.include ENTER name=%s" % p[3])
        with ASTContext(pref):
            global parent_object
            parent_object = "this"
            module = importer(p[3])
            stack.pop()
        for decl in module.decls:
            p[0].declare(decl,allow_redef=True)
        p[0].included.update(module.included)
        p[0].modules.update(module.modules)
        if __debug__: xtracer.trace("parser.include EXIT name=%s decls=%d" % (p[3], len(module.decls)))

def p_labeledfmla_fmla(p):
    'labeledfmla : fmla'
    if __debug__: xtracer.trace("parser.p_labeledfmla_fmla ENTER (labeledfmla)")
    p[0] = LabeledFormula(None,p[1])
    p[0].lineno = p[1].lineno
    
def p_labeledfmla_label_fmla(p):
    'labeledfmla : LABEL fmla'
    if __debug__: xtracer.trace("parser.p_labeledfmla_label_fmla ENTER (labeledfmla)")
    p[0] = LabeledFormula(Atom(p[1][1:-1],[]),p[2])
    p[0].lineno = get_lineno(p,1)

def p_opttemporal(p):
    'opttemporal : '
    if __debug__: xtracer.trace("parser.p_opttemporal ENTER (opttemporal)")
    p[0] = None

def p_opttemporal_symbol(p):
    'opttemporal : TEMPORAL'
    if __debug__: xtracer.trace("parser.p_opttemporal_symbol ENTER (opttemporal)")
    p[0] = True

def addtemporal(lf):
    if __debug__: xtracer.trace("parser.addtemporal ENTER")
    lf.temporal = True
    return lf

def p_optunprovable(p):
    'optunprovable : '
    if __debug__: xtracer.trace("parser.p_optunprovable ENTER (optunprovable)")
    p[0] = None

def p_optunprovable_symbol(p):
    'optunprovable : UNPROVABLE'
    if __debug__: xtracer.trace("parser.p_optunprovable_symbol ENTER (optunprovable)")
    p[0] = True

def addunprovable(lf,cond):
    if __debug__: xtracer.trace("parser.addunprovable ENTER")
    if cond:
        lf.unprovable = True
    return lf

def addexplicit(lf):
    if __debug__: xtracer.trace("parser.addexplicit ENTER")
    lf.explicit = True
    return lf

label_counter = 0

def newlabel(pref):
    if __debug__: xtracer.trace("parser.newlabel ENTER")
    global label_counter
    label_counter += 1
    return Atom(pref+str(label_counter))

def mk_label(s,pref):
    if __debug__: xtracer.trace("parser.mk_label ENTER")
    return s if s is not None else newlabel(pref)
        
def addlabel(lf,pref):
    if __debug__: xtracer.trace("parser.addlabel ENTER")
    if lf.label is not None or iu.get_numeric_version() <= [1,6]:
        return lf
    res = LabeledFormula(newlabel(pref),lf.formula)
    res.lineno = lf.lineno
    return res

if iu.get_numeric_version() <= [1,6]:
    def p_top_axiom_labeledfmla(p):
        'top : top opttemporal AXIOM labeledfmla'
        if __debug__: xtracer.trace("parser.p_top_axiom_labeledfmla ENTER (top)")
        p[0] = p[1]
        lf = addlabel(p[4],'axiom')
        d = AxiomDecl(addtemporal(lf) if p[2] else check_non_temporal(lf))
        d.lineno = get_lineno(p,3)
        p[0].declare(d)
else:
    def p_gprop_fmla(p):
        'gprop : fmla'
        if __debug__: xtracer.trace("parser.p_gprop_fmla ENTER (gprop)")
        p[0] = p[1]

    def p_gprop_schdefnrhs(p):
        'gprop : schdefnrhs'
        if __debug__: xtracer.trace("parser.p_gprop_schdefnrhs ENTER (gprop)")
        p[0] = p[1]

    def p_lgprop(p):
        'lgprop : optlabel gprop'
        if __debug__: xtracer.trace("parser.p_lgprop ENTER (lgprop)")
        lf = LabeledFormula(p[1],p[2])
        lf.lineno = get_lineno(p,2)
        p[0] = lf

    def p_top_axiom_optlabel_gprop(p):
        'top : top optexplicit opttemporal AXIOM lgprop'
        if __debug__: xtracer.trace("parser.p_top_axiom_optlabel_gprop ENTER (top)")
        p[0] = p[1]
        lf = addlabel(p[5],'axiom')
        lf = addexplicit(lf) if p[2] else lf
        d = AxiomDecl(addtemporal(lf) if p[3] else check_non_temporal(lf))
        d.lineno = get_lineno(p,4)
        p[0].declare(d)

def p_optskolem(p):
    'optskolem : '
    if __debug__: xtracer.trace("parser.p_optskolem ENTER (optskolem)")
    p[0] = None

def p_optskolem_symbol(p):
    'optskolem : NAMED defnlhs'
    if __debug__: xtracer.trace("parser.p_optskolem_symbol ENTER (optskolem)")
    p[0] = p[2]
    p[0].lineno = get_lineno(p,1)

def p_top_property_labeledfmla(p):
    'top : top optexplicit opttemporal PROPERTY labeledfmla optskolem optproof'
    if __debug__: xtracer.trace("parser.p_top_property_labeledfmla ENTER (top)")
    p[0] = p[1]
    lf = addlabel(p[5],'prop')
    lf = addtemporal(lf) if p[3] else check_non_temporal(lf)
    lf = addexplicit(lf) if p[2] else lf
    d = PropertyDecl(lf)
    d.lineno = get_lineno(p,4)
    p[0].declare(d)
    if p[6] is not None:
        p[0].declare(NamedDecl(p[6]))
    if p[7] is not None:
        p[0].declare(ProofDecl(p[7]))

def p_top_conjecture_labeledfmla(p):
    'top : top CONJECTURE labeledfmla'
    if __debug__: xtracer.trace("parser.p_top_conjecture_labeledfmla ENTER (top)")
    p[0] = p[1]
    d = ConjectureDecl(addlabel(p[3],'conj'))
    d.lineno = get_lineno(p,2)
    p[0].declare(d)

def p_optexplicit(p):
    'optexplicit : '
    if __debug__: xtracer.trace("parser.p_optexplicit ENTER (optexplicit)")
    p[0] = False

def p_optexplicit_explicit(p):
    'optexplicit : EXPLICIT'
    if __debug__: xtracer.trace("parser.p_optexplicit_explicit ENTER (optexplicit)")
    p[0] = True

# from version 1.7, "invariant" replaces "conjecture"
if not iu.get_numeric_version() <= [1,6]:


    def p_top_invariant_labeledfmla(p):
        'top : top optexplicit INVARIANT labeledfmla optproof'
        if __debug__: xtracer.trace("parser.p_top_invariant_labeledfmla ENTER (top)")
        p[0] = p[1]
        lf = addlabel(p[4],'invar')
        lf.unprovable = False
        if p[2]:
            lf.explicit = True
        d = ConjectureDecl(lf)
        d.lineno = get_lineno(p,3)
        if not lf.unprovable or check_unprovable.get():
            p[0].declare(d)
            if p[5] is not None:
                p[0].declare(ProofDecl(p[5]))

    def p_top_unprovable_invariant_labeledfmla(p):
        'top : top UNPROVABLE INVARIANT labeledfmla optproof'
        if __debug__: xtracer.trace("parser.p_top_unprovable_invariant_labeledfmla ENTER (top)")
        p[0] = p[1]
        lf = addlabel(p[4],'invar')
        lf.unprovable = True
        lf.explicit = True
        d = ConjectureDecl(lf)
        d.lineno = get_lineno(p,3)
        if not lf.unprovable or check_unprovable.get():
            p[0].declare(d)
            if p[5] is not None:
                p[0].declare(ProofDecl(p[5]))

def p_modulestart(p):
    'modulestart :'
    if __debug__: xtracer.trace("parser.p_modulestart ENTER (modulestart)")
    stack[-1].is_module=True
    p[0] = None
 
def p_moduleend(p):
    'moduleend :'
    if __debug__: xtracer.trace("parser.p_moduleend ENTER (moduleend)")
    p[0] = None

def p_modcat(p):
    'modcat : '
    if __debug__: xtracer.trace("parser.p_modcat ENTER (modcat)")
    p[0] = None

def p_modcat_object(p):
    'modcat : OBJECT'
    if __debug__: xtracer.trace("parser.p_modcat_object ENTER (modcat)")
    p[0] = "object"

def p_modcat_isolate(p):
    'modcat : ISOLATE'
    if __debug__: xtracer.trace("parser.p_modcat_isolate ENTER (modcat)")
    p[0] = "isolate"    
    
def p_opteq(p):
    'opteq :'
    if __debug__: xtracer.trace("parser.p_opteq ENTER (opteq)")
    p[0] = None

def p_opteq_eq(p):
    'opteq : EQ'
    if __debug__: xtracer.trace("parser.p_opteq_eq ENTER (opteq)")
    p[0] = p[1]

def p_top_module_atom_eq_lcb_top_rcb(p):
    'top : top MODULE modulestart modcat atom optwith EQ LCB top RCB moduleend'
    if __debug__: xtracer.trace("parser.p_top_module_atom_eq_lcb_top_rcb ENTER (top)")
    p[0] = p[1]
    d = Definition(app_to_atom(p[5]),p[9])
    p[0].declare(ModuleDecl(d))
    if p[4] == "isolate":
        this = Atom(This())
        this.lineno = get_lineno(p,2)
        iso = Atom("iso",[])
        iso.lineno = get_lineno(p,2)
        d = IsolateDecl(IsolateDef(*([iso,this]+p[6])))
        d.args[0].with_args = len(p[6])
        d.args[0].lineno = get_lineno(p,2)
        d.attributes = ("common",)
        d.lineno = get_lineno(p,2)
        p[9].declare(d)
    stack.pop()
    stack[-1].is_module=False

def p_optdotdotdot(p):
    'optdotdotdot : '
    if __debug__: xtracer.trace("parser.p_optdotdotdot ENTER (optdotdotdot)")
    global parent_object
    parent_object = None
    p[0] = False

def p_optdotdotdot_dotdotdot(p):
    'optdotdotdot : DOTDOTDOT'
    if __debug__: xtracer.trace("parser.p_optdotdotdot_dotdotdot ENTER (optdotdotdot)")
    p[0] = True

def p_objectargs_optargs(p):
    'objectargs : optargs'
    if __debug__: xtracer.trace("parser.p_objectargs_optargs ENTER (objectargs)")
    p[0] = p[1]
    stack[-1].params = p[0]

def p_objectend(p):
    'objectend :'
    if __debug__: xtracer.trace("parser.p_objectend ENTER (objectend)")
    stack[-1].is_object=False
    p[0] = None

def create_object(top,name,objectargs,module,lineno=None,continuation=False):
    if __debug__: xtracer.trace("parser.create_object ENTER name=%s" % name)
    prefargs = [Variable('V'+str(idx),pr.sort) for idx,pr in enumerate(objectargs)]
    pref = Atom(name,prefargs)
    pref.lineno = lineno
#    top.define((pref.rep,get_lineno(p,2)))
    if not continuation:
        top.declare(ObjectDecl(pref))
        top.set_object_defined(name,module.defined)
    vsubst = dict((pr.rep,v) for pr,v in zip(objectargs,prefargs))
    inst_mod(top,module,pref,{},vsubst)
    # for decl in module.decls:
    #     idecl = subst_prefix_atoms_ast(decl,subst,pref,module.defined)
    #     top.declare(idecl)
    stack.pop()
    if __debug__: xtracer.trace("parser.create_object EXIT name=%s" % name)

def p_objsym(p):
    'objsym : SYMBOL'
    if __debug__: xtracer.trace("parser.p_objsym ENTER (objsym)")
    p[0] = p[1]
    global parent_object
    parent_object = p[0]

def p_top_object_symbol_eq_lcb_top_rcb(p):
    'top : top OBJECT objsym objectargs EQ LCB optdotdotdot top RCB objectend'
    if __debug__: xtracer.trace("parser.p_top_object_symbol_eq_lcb_top_rcb ENTER (top)")
    p[0] = p[1]
    create_object(p[0],p[3],p[4],p[8],get_lineno(p,3),p[7])

def p_top_class_symbol_eq_lcb_top_rcb(p):
    'top : top CLASS objsym objectargs EQ LCB optdotdotdot top RCB objectend'
    if __debug__: xtracer.trace("parser.p_top_class_symbol_eq_lcb_top_rcb ENTER (top)")
    p[0] = p[1]
    scnst = Atom(This())
    scnst.lineno = get_lineno(p,2)
    tdfn = TypeDef(scnst,UninterpretedSort())
    tdfn.lineno = get_lineno(p,2)
    p[8].declare(TypeDecl(tdfn))
    p[8].decls = [p[8].decls[-1]] + p[8].decls[:-1]
    create_object(p[0],p[3],p[4],p[8],get_lineno(p,3),p[7])

def p_top_subclass_symbol_eq_lcb_top_rcb(p):
    'top : top SUBCLASS objsym OF atype EQ LCB optdotdotdot top RCB objectend'
    if __debug__: xtracer.trace("parser.p_top_subclass_symbol_eq_lcb_top_rcb ENTER (top)")
    p[0] = p[1]
    scnst = Atom(This())
    scnst.lineno = get_lineno(p,2)
    tdfn = TypeDef(scnst,UninterpretedSort())
    tdfn.lineno = get_lineno(p,2)
    p[9].declare(TypeDecl(tdfn))
    vdfn = VariantDef(scnst,Atom(p[5]))
    p[9].declare(VariantDecl(vdfn))
    p[9].decls = p[9].decls[-2:] + p[9].decls[:-2]
    create_object(p[0],p[3],[],p[9],get_lineno(p,3),p[8])

def p_optsemi(p):
    'optsemi : '
    if __debug__: xtracer.trace("parser.p_optsemi ENTER (optsemi)")
    p[0] = None

def p_optsemi_semi(p):
    'optsemi : SEMI'
    if __debug__: xtracer.trace("parser.p_optsemi_semi ENTER (optsemi)")
    p[0] = None

def p_top_macro_atom_eq_lcb_action_rcb(p):
    'top : top MACRO atom EQ sequence'
    if __debug__: xtracer.trace("parser.p_top_macro_atom_eq_lcb_action_rcb ENTER (top)")
    p[0] = p[1]
    d = Definition(app_to_atom(p[3]),p[5])
    p[0].declare(MacroDecl(d))

def p_schdefnrhs_fmla(p):
    'schdefnrhs : fmla'
    if __debug__: xtracer.trace("parser.p_schdefnrhs_fmla ENTER (schdefnrhs)")
    p[0] = check_non_temporal(p[1])

def p_schdecl_funcdecl(p):
    'schdecl : FUNCTION funs'
    if __debug__: xtracer.trace("parser.p_schdecl_funcdecl ENTER (schdecl)")
    p[0] = p[2]

def p_schdecl_fresh_funcdecl(p):
    'schdecl : FRESH FUNCTION funs'
    if __debug__: xtracer.trace("parser.p_schdecl_fresh_funcdecl ENTER (schdecl)")
    p[0] = [FreshConstantDecl(x.args[0]) if isinstance(x,ConstantDecl) else x for x in p[3]]

def p_schdecl_indivdecl(p):
    'schdecl : INDIV funs'
    if __debug__: xtracer.trace("parser.p_schdecl_indivdecl ENTER (schdecl)")
    p[0] = p[2]

def p_schdecl_fresh_indivdecl(p):
    'schdecl : FRESH INDIV funs'
    if __debug__: xtracer.trace("parser.p_schdecl_fresh_indivdecl ENTER (schdecl)")
    p[0] = [FreshConstantDecl(x.args[0]) if isinstance(x,ConstantDecl) else x for x in p[3]]

def p_schdecl_relation_rel(p):
    'schdecl : RELATION rels'
    if __debug__: xtracer.trace("parser.p_schdecl_relation_rel ENTER (schdecl)")
    p[0] = p[2]

def p_schdecl_fresh_relation_rel(p):
    'schdecl : FRESH RELATION rels'
    if __debug__: xtracer.trace("parser.p_schdecl_fresh_relation_rel ENTER (schdecl)")
    p[0] = [FreshConstantDecl(x.args[0]) if isinstance(x,ConstantDecl) else x for x in p[3]]

def p_schdecl_typedecl(p):
    'schdecl : TYPE SYMBOL'
    if __debug__: xtracer.trace("parser.p_schdecl_typedecl ENTER (schdecl)")
    scnst = Atom(p[2])
    scnst.lineno = get_lineno(p,2)
    tdfn = TypeDef(scnst,UninterpretedSort())
    tdfn.lineno = get_lineno(p,1)
    p[0] = [tdfn]

if iu.get_numeric_version() <= [1,6]:

    def p_schdecl_propdecl(p):
        'schdecl : PROPERTY labeledfmla'
        if __debug__: xtracer.trace("parser.p_schdecl_propdecl ENTER (schdecl)")
        p[0] = [check_non_temporal(addlabel(p[2],'prop'))]

else:

    def p_schdecl_propdecl(p):
        'schdecl : optexplicit PROPERTY lgprop'
        if __debug__: xtracer.trace("parser.p_schdecl_propdecl ENTER (schdecl)")
        lf = addlabel(p[3],'prop')
        if p[1]:
            lf.explicit = True
        p[0] = [check_non_temporal(lf)]

    def p_schdecl_theorem_lgprop(p):
        'schdecl : THEOREM lgprop'
        if __debug__: xtracer.trace("parser.p_schdecl_theorem_lgprop ENTER (schdecl)")
        p[0] = [check_non_temporal(addlabel(p[2],'prop'))]

def p_schconc_defdecl(p):
    'schconc : DEFINITION defn'
    if __debug__: xtracer.trace("parser.p_schconc_defdecl ENTER (schconc)")
    p[0] = p[2]

if iu.get_numeric_version() <= [1,6]:

    def p_schconc_propdecl(p):
        'schconc : PROPERTY fmla'
        if __debug__: xtracer.trace("parser.p_schconc_propdecl ENTER (schconc)")
        p[0] = check_non_temporal(p[2])

else:

    def p_schconc_propdecl(p):
        'schconc : optexplicit PROPERTY lgprop'
        if __debug__: xtracer.trace("parser.p_schconc_propdecl ENTER (schconc)")
        fmla = p[3].formula
        if isinstance(fmla,SchemaBody):
            report_error(IvyError(fmla,"formula expected"))
        p[0] = check_non_temporal(fmla)


def p_schdecl_theorem(p):
    'schdecl :  schdefnrhs'
    if __debug__: xtracer.trace("parser.p_schdecl_theorem ENTER (schdecl)")
    lf = LabeledFormula(None,p[1])
    lf.lineno = p[1].lineno
    p[0] = [addlabel(lf,'sch')]

def p_schdecls(p):
    'schdecls :'
    if __debug__: xtracer.trace("parser.p_schdecls ENTER (schdecls)")
    p[0] = []

def p_schdecls_schdecls_schdecl(p):
    'schdecls : schdecls schdecl'
    if __debug__: xtracer.trace("parser.p_schdecls_schdecls_schdecl ENTER (schdecls)")
    p[0] = p[1]
    p[0].extend(p[2])

def p_schdefnrhs_lcb_schdecls_rcb(p):
    'schdefnrhs : LCB schdecls schconc RCB'
    if __debug__: xtracer.trace("parser.p_schdefnrhs_lcb_schdecls_rcb ENTER (schdefnrhs)")
    p[0] = SchemaBody(*(p[2]+[p[3]]))
    p[0].lineno = get_lineno(p,1)

def p_schdefn_atom_eq_fmla(p):
    'schdefn : defnlhs EQ schdefnrhs'
    if __debug__: xtracer.trace("parser.p_schdefn_atom_eq_fmla ENTER (schdefn)")
    p[0] = Definition(app_to_atom(p[1]),p[3])
    p[0].lineno = get_lineno(p,2)

def p_top_schema_defn(p):
    'top : top SCHEMA schdefn'
    if __debug__: xtracer.trace("parser.p_top_schema_defn ENTER (top)")
    p[0] = p[1]
    p[0].declare(SchemaDecl(Schema(p[3])))

def p_top_theorem_defn(p):
    'top : top THEOREM schdefn optproof'
    if __debug__: xtracer.trace("parser.p_top_theorem_defn ENTER (top)")
    p[0] = p[1]
    p[0].declare(TheoremDecl(Schema(p[3])))
    if p[4] is not None:
        p[0].declare(ProofDecl(p[4]))

def p_top_theorem_label_rhs(p):
    'top : top THEOREM LABEL schdefnrhs optproof'
    if __debug__: xtracer.trace("parser.p_top_theorem_label_rhs ENTER (top)")
    p[0] = p[1]
    label = Atom(p[3][1:-1],[])
    label.lineno = get_lineno(p,3)
    df = Definition(label,p[4])
    df.lineno = get_lineno(p,3)
    p[0].declare(TheoremDecl(Schema(df)))
    if p[5] is not None:
        p[0].declare(ProofDecl(p[5]))

def p_top_proof_label_label_proofstep(p):
    'top : top PROOF LABEL proofstep'
    if __debug__: xtracer.trace("parser.p_top_proof_label_label_proofstep ENTER (top)")
    p[0] = p[1]
    label = Atom(p[3][1:-1],[])
    label.lineno = get_lineno(p,3)
    p[0].declare(ProofDecl(LabeledFormula(label,p[4])))

def p_top_instantiate_insts(p):
    'top : top INSTANTIATE insts'
    if __debug__: xtracer.trace("parser.p_top_instantiate_insts ENTER (top)")
    p[0] = p[1]
    do_insts(p[0],p[3])

def p_top_autoinstance_insts(p):
    'top : top AUTOINSTANCE insts'
    if __debug__: xtracer.trace("parser.p_top_autoinstance_insts ENTER (top)")
    p[0] = p[1]
    p[0].declare(AutoInstanceDecl(*p[3]))

def p_insts_inst(p):
    'insts : inst'
    if __debug__: xtracer.trace("parser.p_insts_inst ENTER (insts)")
    p[0] = [p[1]]

def p_insts_insts_comma_inst(p):
    'insts : insts COMMA inst'
    if __debug__: xtracer.trace("parser.p_insts_insts_comma_inst ENTER (insts)")
    p[0] = p[1]
    p[0].append(p[3])

def p_pname_symbol(p):
    'pname : atype'
    if __debug__: xtracer.trace("parser.p_pname_symbol ENTER (pname)")
    p[0] = App(p[1])
    p[0].lineno = get_lineno(p,1)

def p_pname_var(p):
    'pname : var'
    if __debug__: xtracer.trace("parser.p_pname_var ENTER (pname)")
    p[0] = p[1]

def p_pname_infix(p):
    'pname : infix'
    if __debug__: xtracer.trace("parser.p_pname_infix ENTER (pname)")
    p[0] = App(p[1])
    p[0].lineno = get_lineno(p,1)

def p_pname_relop(p):
    'pname : relop'
    if __debug__: xtracer.trace("parser.p_pname_relop ENTER (pname)")
    p[0] = App(p[1])
    p[0].lineno = get_lineno(p,1)

def p_pname_this(p):
    'pname : THIS'
    if __debug__: xtracer.trace("parser.p_pname_this ENTER (pname)")
    p[0] = App(This())
    p[0].lineno = get_lineno(p,1)

def p_pname_true(p):
    'pname : TRUE'
    if __debug__: xtracer.trace("parser.p_pname_true ENTER (pname)")
    p[0] = Atom('true')
    p[0].lineno = get_lineno(p,1)

def p_pname_false(p):
    'pname : FALSE'
    if __debug__: xtracer.trace("parser.p_pname_false ENTER (pname)")
    p[0] = Atom('false')
    p[0].lineno = get_lineno(p,1)

def p_pnames(p):
    'pnames : '
    if __debug__: xtracer.trace("parser.p_pnames ENTER (pnames)")
    p[0] = []

def p_pnames_pname(p):
    'pnames : pname'
    if __debug__: xtracer.trace("parser.p_pnames_pname ENTER (pnames)")
    p[0] = [p[1]]

def p_pnames_pnames_pname(p):
    'pnames : pnames COMMA pname'
    if __debug__: xtracer.trace("parser.p_pnames_pnames_pname ENTER (pnames)")
    p[0] = p[1]
    p[0].append(p[3])

def p_modinst_symbol(p):
    'modinst : dotsym'
    if __debug__: xtracer.trace("parser.p_modinst_symbol ENTER (modinst)")
    p[0] = p[1]

def p_modinst_symbol_lp_pnames_rp(p):
    'modinst : dotsym LPAREN pnames RPAREN'
    if __debug__: xtracer.trace("parser.p_modinst_symbol_lp_pnames_rp ENTER (modinst)")
    p[0] = p[1]
    p[0].args = p[3]

def p_inst_modinst(p):
   'inst : modinst'
   if __debug__: xtracer.trace("parser.p_inst_modinst ENTER (inst)")
   p[0] = Instantiation(None,app_to_atom(p[1]))
   p[0].lineno = get_lineno(p,1)
   
def p_inst_atom_colon_modinst(p):
    'inst : modinst COLON modinst'
    if __debug__: xtracer.trace("parser.p_inst_atom_colon_modinst ENTER (inst)")
    p[0] = Instantiation(app_to_atom(p[1]),app_to_atom(p[3]))
    p[0].lineno = get_lineno(p,2)
    
def p_top_symdecl(p):
    'top : top symdecl'
    if __debug__: xtracer.trace("parser.p_top_symdecl ENTER (top)")
    p[0] = p[1]
    p[0].declare(p[2])

def p_symdecl_constantdecl(p):
    'symdecl : constantdecl'
    if __debug__: xtracer.trace("parser.p_symdecl_constantdecl ENTER (symdecl)")
    p[0] = p[1]

def p_symdecl_destructor_tterms(p):
    'symdecl : DESTRUCTOR tterms'
    if __debug__: xtracer.trace("parser.p_symdecl_destructor_tterms ENTER (symdecl)")
    p[0] = DestructorDecl(*p[2])
    p[0].lineno = get_lineno(p,1)

def p_symdecl_field_tterms(p):
    'symdecl : FIELD tterms'
    if __debug__: xtracer.trace("parser.p_symdecl_field_tterms ENTER (symdecl)")
    arg0 = Variable('SELF',This())
    arg0.lineno = get_lineno(p,1)
    tterms = [x.clone([arg0]+x.args) for x in p[2]]
    for x,y in zip(p[2],tterms):
        y.lineno = x.lineno
    p[0] = DestructorDecl(*tterms)
    p[0].lineno = get_lineno(p,1)

if not(iu.get_numeric_version() <= [1,6]):
    def p_symdecl_constructor_tterms(p):
        'symdecl : CONSTRUCTOR tterms'
        if __debug__: xtracer.trace("parser.p_symdecl_constructor_tterms ENTER (symdecl)")
        for t in p[2]:
            if not hasattr(t,'sort'):
                this = This()
                this.lineno = get_lineno(p,1)
                t.sort = this
        p[0] = ConstructorDecl(*p[2])
        p[0].lineno = get_lineno(p,1)

def p_constantdecl_constant_tterms(p):
    'constantdecl : INDIV tterms'
    if __debug__: xtracer.trace("parser.p_constantdecl_constant_tterms ENTER (constantdecl)")
    p[0] = ConstantDecl(*p[2])
    p[0].lineno = get_lineno(p,1)

def p_constantdecl_var_tterms(p):
    'constantdecl : VAR tterms'
    if __debug__: xtracer.trace("parser.p_constantdecl_var_tterms ENTER (constantdecl)")
    p[0] = ConstantDecl(*p[2])
    p[0].lineno = get_lineno(p,1)

def p_param_tterm(p):
    'parameter : tterm'
    if __debug__: xtracer.trace("parser.p_param_tterm ENTER (parameter)")
#    p[1].sort = p[3]
    p[0] = ParameterDecl(p[1])
    p[0].lineno = p[1].lineno

def p_paramval_true(p):
    'paramval : TRUE'
    if __debug__: xtracer.trace("parser.p_paramval_true ENTER (paramval)")
    p[0] = Atom('true')
    p[0].lineno = get_lineno(p,1)

def p_paramval_false(p):
    'paramval : FALSE'
    if __debug__: xtracer.trace("parser.p_paramval_false ENTER (paramval)")
    p[0] = Atom('false')
    p[0].lineno = get_lineno(p,1)

def p_paramval_symbol(p):
    'paramval : SYMBOL'
    if __debug__: xtracer.trace("parser.p_paramval_symbol ENTER (paramval)")
    p[0] = App(p[1])
    p[0].lineno = get_lineno(p,1)

def p_param_tterm_eq_paramval(p):
    'parameter : tterm EQ paramval'
    if __debug__: xtracer.trace("parser.p_param_tterm_eq_paramval ENTER (parameter)")
    dflt = p[3]
    df = Definition(p[1],dflt)
    df.lineno = get_lineno(p,2)
    p[0] = ParameterDecl(df)

def p_constantdecl_parameter_tterm(p):
    'constantdecl : PARAMETER parameter'
    if __debug__: xtracer.trace("parser.p_constantdecl_parameter_tterm ENTER (constantdecl)")
    p[0] = p[2]

def p_rel_defnlhs(p):
    'rel : defnlhs'
    if __debug__: xtracer.trace("parser.p_rel_defnlhs ENTER (rel)")
    p[1].sort = 'bool'
    p[0] = ConstantDecl(p[1])

def p_rel_defn(p):
    'rel : defn'
    if __debug__: xtracer.trace("parser.p_rel_defn ENTER (rel)")
    p[0] = DerivedDecl(addlabel(mk_lf(p[1]),'def'))

def p_rels_rel(p):
    'rels : rel'
    if __debug__: xtracer.trace("parser.p_rels_rel ENTER (rels)")
    p[0] = [p[1]]

def p_rels_rels_comma_rel(p):
    'rels : rels COMMA rel'
    if __debug__: xtracer.trace("parser.p_rels_rels_comma_rel ENTER (rels)")
    p[0] = p[1]
    p[0].append(p[3])

def p_top_relation_rels(p):
    'top : top RELATION rels'
    if __debug__: xtracer.trace("parser.p_top_relation_rels ENTER (top)")
    p[0] = p[1]
    for d in p[3]:
        p[0].declare(d)

def p_tatoms_tatom(p):
    'tatoms : tatom'
    if __debug__: xtracer.trace("parser.p_tatoms_tatom ENTER (tatoms)")
    p[0] = [p[1]]

def p_tatoms_tatoms_comma_tatom(p):
    'tatoms : tatoms COMMA tatom'
    if __debug__: xtracer.trace("parser.p_tatoms_tatoms_comma_tatom ENTER (tatoms)")
    p[0] = p[1]
    p[0].append(p[3])

def p_tatom_symbol(p):
    'tatom : SYMBOL'
    if __debug__: xtracer.trace("parser.p_tatom_symbol ENTER (tatom)")
    p[0] = Atom(p[1],[])
    p[0].lineno = get_lineno(p,1)

def p_tatom_symbol_targs(p):
    'tatom : SYMBOL targs'
    if __debug__: xtracer.trace("parser.p_tatom_symbol_targs ENTER (tatom)")
    p[0] = Atom(p[1],p[2])
    p[0].lineno = get_lineno(p,1)

def p_tatom_lp_symbol_relop_symbol_rp(p):
    'tatom : LPAREN var relop var RPAREN'
    if __debug__: xtracer.trace("parser.p_tatom_lp_symbol_relop_symbol_rp ENTER (tatom)")
    p[0] = Atom(p[3],[p[2],p[4]])
    p[0].lineno = get_lineno(p,3)

def p_fun_defnlhs_colon_atype(p):
    'fun : typeddefn'
    if __debug__: xtracer.trace("parser.p_fun_defnlhs_colon_atype ENTER (fun)")
#    p[1].sort = p[3]
    p[0] = ConstantDecl(p[1])
    p[0].lineno = p[1].lineno

def p_fun_defn(p):
    'fun : typeddefn EQ defnrhs'
    if __debug__: xtracer.trace("parser.p_fun_defn ENTER (fun)")
    df = Definition(app_to_atom(p[1]),p[3])
    df.lineno = get_lineno(p,2)
    p[0] = DerivedDecl(addlabel(mk_lf(df),'def'))

def p_funs_fun(p):
    'funs : fun'
    if __debug__: xtracer.trace("parser.p_funs_fun ENTER (funs)")
    p[0] = [p[1]]

def p_funs_funs_comma_fun(p):
    'funs : funs COMMA fun'
    if __debug__: xtracer.trace("parser.p_funs_funs_comma_fun ENTER (funs)")
    p[0] = p[1]
    p[0].append(p[3])

def p_top_function_tapp_colon_atype(p):
    'top : top FUNCTION funs'
    if __debug__: xtracer.trace("parser.p_top_function_tapp_colon_atype ENTER (top)")
    p[0] = p[1]
    for d in p[3]:
        p[0].declare(d)

def mk_lf(x):
    if __debug__: xtracer.trace("parser.mk_lf ENTER")
    res = LabeledFormula(None,x)
    res.lineno = x.lineno
    return res

def p_top_derived_defns(p):
    'top : top DERIVED defns'
    if __debug__: xtracer.trace("parser.p_top_derived_defns ENTER (top)")
    p[0] = p[1]
    p[0].declare(DerivedDecl(*[addlabel(mk_lf(x),'def') for x in p[3]]))

if iu.get_numeric_version() <= [1,6]:
    def p_proofstep_symbol(p):
        'proofstep : SYMBOL'
        if __debug__: xtracer.trace("parser.p_proofstep_symbol ENTER (proofstep)")
        a = Atom(p[1])
        a.lineno = get_lineno(p,1)
        p[0] = SchemaInstantiation(a,Renaming())
        p[0].lineno = get_lineno(p,1)
else:
    def p_proofstep_symbol(p):
        'proofstep : APPLY atype optrenaming'
        if __debug__: xtracer.trace("parser.p_proofstep_symbol ENTER (proofstep)")
        a = Atom(p[2])
        a.lineno = get_lineno(p,2)
        p[0] = SchemaInstantiation(a,p[3])
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_assume(p):
        'proofstep : ASSUME atype optrenaming'
        if __debug__: xtracer.trace("parser.p_proofstep_assume ENTER (proofstep)")
        a = Atom(p[2])
        a.lineno = get_lineno(p,2)
        p[0] = AssumeGlobalTactic(a,p[3])
        p[0].label = NoneAST()
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_instantiate(p):
        'proofstep : INSTANTIATE atype optrenaming'
        if __debug__: xtracer.trace("parser.p_proofstep_instantiate ENTER (proofstep)")
        a = Atom(p[2])
        a.lineno = get_lineno(p,2)
        p[0] = AssumeTactic(a,p[3])
        p[0].label = NoneAST()
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_instance(p):
        'proofstep : INSTANTIATE LABEL atype optrenaming'
        if __debug__: xtracer.trace("parser.p_proofstep_instance ENTER (proofstep)")
        a = Atom(p[3])
        a.lineno = get_lineno(p,2)
        label = Atom(p[2][1:-1],[])
        label.lineno = get_lineno(p,2)
        p[0] = AssumeTactic(a,p[4])
        p[0].label = label
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_showgoals(p):
        'proofstep : SHOWGOALS'
        if __debug__: xtracer.trace("parser.p_proofstep_showgoals ENTER (proofstep)")
        p[0] = ShowGoalsTactic()
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_defergoal(p):
        'proofstep : DEFERGOAL'
        if __debug__: xtracer.trace("parser.p_proofstep_defergoal ENTER (proofstep)")
        p[0] = DeferGoalTactic()
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_spoil_atype(p):
        'proofstep : SPOIL atype'
        if __debug__: xtracer.trace("parser.p_proofstep_spoil_atype ENTER (proofstep)")
        a = Atom(p[2])
        a.lineno = get_lineno(p,2)
        p[0] = SpoilTactic(a)
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_tactic(p):
        'proofstep : TACTIC SYMBOL opttacticwith optproofgroup'
        if __debug__: xtracer.trace("parser.p_proofstep_tactic ENTER (proofstep)")
        a = Atom(p[2])
        a.lineno = get_lineno(p,2)
        proof = p[4] if p[4] else NoneAST()
        p[0] = TacticTactic(a,p[3],proof)
        p[0].lineno = get_lineno(p,1)
    
    def p_proofstep_property(p):
        'proofstep : opttemporal PROPERTY labeledfmla optskolem optproofgroup'
        if __debug__: xtracer.trace("parser.p_proofstep_property ENTER (proofstep)")
        lf = addlabel(p[3],'prop')
        prop = addtemporal(lf) if p[1] else check_non_temporal(lf)
        name = p[4] if p[4] else NoneAST()
        proof = p[5] if p[5] else NoneAST()
        p[0] = PropertyTactic(prop,name,proof)
        p[0].lineno = get_lineno(p,2)

    def p_proofstep_function(p):
        'proofstep : FUNCTION funs'
        if __debug__: xtracer.trace("parser.p_proofstep_function ENTER (proofstep)")
        p[0] = FunctionTactic(*p[2])
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_theorem(p):
        'proofstep : THEOREM lgprop optproofgroup'
        if __debug__: xtracer.trace("parser.p_proofstep_theorem ENTER (proofstep)")
        lf = addlabel(p[2],'thm')
        proof = p[3] if p[3] else NoneAST()
        p[0] = PropertyTactic(lf,NoneAST(),proof)
        p[0].lineno = get_lineno(p,2)

    def p_proofstep_proof(p):
        'proofstep : PROOF LABEL proofgroup'
        if __debug__: xtracer.trace("parser.p_proofstep_proof ENTER (proofstep)")
        label = Atom(p[2][1:-1],[])
        label.lineno = get_lineno(p,2)
        p[0] = ProofTactic(label,p[3])
        p[0].lineno = get_lineno(p,1)

def p_match_defn(p):
    'match : defn'
    if __debug__: xtracer.trace("parser.p_match_defn ENTER (match)")
    p[0] = p[1]

def p_match_var_eq_fmla(p):
    'match : var EQ fmla'
    if __debug__: xtracer.trace("parser.p_match_var_eq_fmla ENTER (match)")
    p[0] = Definition(p[1],check_non_temporal(p[3]))
    p[0].lineno = get_lineno(p,2)

def p_matches(p):
    'matches : match'
    if __debug__: xtracer.trace("parser.p_matches ENTER (matches)")
    p[0] = [p[1]]

def p_matches_matches_comma_match(p):
    'matches : matches COMMA match'
    if __debug__: xtracer.trace("parser.p_matches_matches_comma_match ENTER (matches)")
    p[0] = p[1]
    p[0].append(p[3])

if iu.get_numeric_version() <= [1,6]:
    def p_proofstep_symbol_with_defns(p):
        'proofstep : SYMBOL WITH matches'
        if __debug__: xtracer.trace("parser.p_proofstep_symbol_with_defns ENTER (proofstep)")
        a = Atom(p[1])
        a.lineno = get_lineno(p,1)
        p[0] = SchemaInstantiation(*([a,Renaming()]+p[3]))
        p[0].lineno = get_lineno(p,1)
else:

    def p_renamingitem_variable_div_variable(p):
        'renamingitem : VARIABLE DIV VARIABLE'
        if __debug__: xtracer.trace("parser.p_renamingitem_variable_div_variable ENTER (renamingitem)")
        p[0] = Definition(Variable(p[3],universe),Variable(p[1],universe))
        p[0].lineno = get_lineno(p,2)

    def p_renamingitem_symbol_div_symbol(p):
        'renamingitem : SYMBOL DIV SYMBOL'
        if __debug__: xtracer.trace("parser.p_renamingitem_symbol_div_symbol ENTER (renamingitem)")
        p[0] = Definition(Atom(p[3],[]),Atom(p[1],[]))
        p[0].lineno = get_lineno(p,2)

    def p_renaminglist_renamingitem(p):
        'renaminglist : renamingitem'
        if __debug__: xtracer.trace("parser.p_renaminglist_renamingitem ENTER (renaminglist)")
        p[0] = [p[1]]

    def p_renaminglist_renaminglist_comma_renamingitem(p):
        'renaminglist : renaminglist COMMA renamingitem'
        if __debug__: xtracer.trace("parser.p_renaminglist_renaminglist_comma_renamingitem ENTER (renaminglist)")
        p[0] = p[1]
        p[0].append(p[3])

    def p_renaming(p):
        'optrenaming : '
        if __debug__: xtracer.trace("parser.p_renaming ENTER (optrenaming)")
        p[0] = Renaming()

    def p_renaming_lt_renaminglist_gt(p):
        'renaming : LT renaminglist GT'
        if __debug__: xtracer.trace("parser.p_renaming_lt_renaminglist_gt ENTER (renaming)")
        p[0] = Renaming(*p[2])
        p[0].lineno = get_lineno(p,1)

    def p_optrenaming_renaming(p):
        'optrenaming : renaming'
        if __debug__: xtracer.trace("parser.p_optrenaming_renaming ENTER (optrenaming)")
        p[0] = p[1]

    def p_proofstep_symbol_with_defns(p):
        'proofstep : APPLY atype optrenaming WITH matches'
        if __debug__: xtracer.trace("parser.p_proofstep_symbol_with_defns ENTER (proofstep)")
        a = Atom(p[2])
        a.lineno = get_lineno(p,2)
        p[0] = SchemaInstantiation(*([a,p[3]]+p[5]))
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_assume_with_defns(p):
        'proofstep : ASSUME atype optrenaming WITH matches'
        if __debug__: xtracer.trace("parser.p_proofstep_assume_with_defns ENTER (proofstep)")
        a = Atom(p[2])
        a.lineno = get_lineno(p,2)
        p[0] = AssumeGlobalTactic(*([a,p[3]]+p[5]))
        p[0].label = NoneAST()
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_instantiate_with_defns(p):
        'proofstep : INSTANTIATE atype optrenaming WITH matches'
        if __debug__: xtracer.trace("parser.p_proofstep_instantiate_with_defns ENTER (proofstep)")
        a = Atom(p[2])
        a.lineno = get_lineno(p,2)
        p[0] = AssumeTactic(*([a,p[3]]+p[5]))
        p[0].label = NoneAST()
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_instance_with_matches(p):
        'proofstep : INSTANTIATE LABEL atype optrenaming WITH matches'
        if __debug__: xtracer.trace("parser.p_proofstep_instance_with_matches ENTER (proofstep)")
        a = Atom(p[3])
        a.lineno = get_lineno(p,2)
        label = Atom(p[2][1:-1],[])
        label.lineno = get_lineno(p,2)
        p[0] = AssumeTactic(*([a,p[4]]+p[6]))
        p[0].label = label
        p[0].lineno = get_lineno(p,1)

    def p_renamings(p):
        'renamings : '
        if __debug__: xtracer.trace("parser.p_renamings ENTER (renamings)")
        p[0] = []

    def p_renamings_renamings_renaming(p):
        'renamings : renamings renaming'
        if __debug__: xtracer.trace("parser.p_renamings_renamings_renaming ENTER (renamings)")
        p[0] = p[1]
        p[0].append(p[2])

    def p_unfspec_callatom_renamings(p):
        'unfspec : callatom renamings'
        if __debug__: xtracer.trace("parser.p_unfspec_callatom_renamings ENTER (unfspec)")
        p[0] = UnfoldSpec(*([p[1]]+p[2]))
        p[0].lineno = p[1].lineno

    def p_unfspecs_unfspec(p):
        'unfspecs : unfspec'
        if __debug__: xtracer.trace("parser.p_unfspecs_unfspec ENTER (unfspecs)")
        p[0] = [p[1]]

    def p_unfspecs_unfspecs_unfspec(p):
        'unfspecs : unfspecs COMMA unfspec'
        if __debug__: xtracer.trace("parser.p_unfspecs_unfspecs_unfspec ENTER (unfspecs)")
        p[0] = p[1]
        p[0].append(p[3])

    def p_proofstep_unfold_atype_with_defns(p):
        'proofstep : UNFOLD atype WITH unfspecs'
        if __debug__: xtracer.trace("parser.p_proofstep_unfold_atype_with_defns ENTER (proofstep)")
        a = Atom(p[2])
        a.lineno = get_lineno(p,2)
        p[0] = UnfoldTactic(*([a]+p[4]))
        p[0].label = NoneAST()
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_unfold_with_defns(p):
        'proofstep : UNFOLD WITH unfspecs'
        if __debug__: xtracer.trace("parser.p_proofstep_unfold_with_defns ENTER (proofstep)")
        p[0] = UnfoldTactic(*([NoneAST()]+p[3]))
        p[0].label = NoneAST()
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_forget_callatoms(p):
        'proofstep : FORGET callatoms'
        if __debug__: xtracer.trace("parser.p_proofstep_forget_callatoms ENTER (proofstep)")
        p[0] = ForgetTactic(*(p[2]))
        p[0].lineno = get_lineno(p,1)

    def p_pflet_var_eq_fmla(p):
        'pflet : var EQ fmla'
        if __debug__: xtracer.trace("parser.p_pflet_var_eq_fmla ENTER (pflet)")
        p[0] = Definition(p[1],p[3])
        p[0].lineno = get_lineno(p,2)

    def p_pflets_pflet(p):
        'pflets : pflet'
        if __debug__: xtracer.trace("parser.p_pflets_pflet ENTER (pflets)")
        p[0] = [p[1]]

    def p_pflets_pflets_pflet(p):
        'pflets : pflets COMMA pflet'
        if __debug__: xtracer.trace("parser.p_pflets_pflets_pflet ENTER (pflets)")
        p[0] = p[1]
        p[0].append(p[3])
        
    def p_proofstep_let_pflets(p):
        'proofstep : LET pflets'
        if __debug__: xtracer.trace("parser.p_proofstep_let_pflets ENTER (proofstep)")
        p[0] = LetTactic(*p[2])
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_witness_pflets(p):
        'proofstep : INSTANTIATE WITH pflets'
        if __debug__: xtracer.trace("parser.p_proofstep_witness_pflets ENTER (proofstep)")
        p[0] = WitnessTactic(*p[3])
        p[0].lineno = get_lineno(p,1)

    def p_proofstep_if_fmla_proofgroup_else_proofgroup (p):
        'proofstep : IF fmla proofgroup ELSE proofgroup'
        if __debug__: xtracer.trace("parser.p_proofstep_if_fmla_proofgroup_else_proofgroup ENTER (proofstep)")
        p[0] = IfTactic(p[2],p[3],p[5])
        p[0].lineno = get_lineno(p,1)

    def p_opttacticwith(p):
        'opttacticwith : '
        if __debug__: xtracer.trace("parser.p_opttacticwith ENTER (opttacticwith)")
        p[0] = TacticWith()

    def p_opttacticwith_with_tacticwithlist(p):
        'opttacticwith : WITH tacticwithlistchoice'
        if __debug__: xtracer.trace("parser.p_opttacticwith_with_tacticwithlist ENTER (opttacticwith)")
        p[0] = p[2]
        p[0].lineno = get_lineno(p,1)

    def p_tacticwithlistchoice_tactwithlist(p):
        'tacticwithlistchoice : tacticwithlist'
        if __debug__: xtracer.trace("parser.p_tacticwithlistchoice_tactwithlist ENTER (tacticwithlistchoice)")
        p[0] = TacticWith(*p[1])

    def p_tacticwithlistchoice_pflets(p):
        'tacticwithlistchoice : pflets'
        if __debug__: xtracer.trace("parser.p_tacticwithlistchoice_pflets ENTER (tacticwithlistchoice)")
        p[0] = TacticLets(*p[1])

    def p_tacticwithelem_invariant(p):
        'tacticwithelem : INVARIANT labeledfmla'
        if __debug__: xtracer.trace("parser.p_tacticwithelem_invariant ENTER (tacticwithelem)")
        p[0] = addlabel(p[2],'invar')
        
    def p_tacticwithelem_fun_defn(p):
        'tacticwithelem : DEFINITION typeddefn EQ fmla'
        if __debug__: xtracer.trace("parser.p_tacticwithelem_fun_defn ENTER (tacticwithelem)")
        df = Definition(app_to_atom(p[2]),p[4])
        df.lineno = get_lineno(p,3)
        p[0] = DerivedDecl(addlabel(mk_lf(df),'def'))

    def p_tacticwithelem_trigger(p):
        'tacticwithelem : TRIGGER atype WITH terms'
        if __debug__: xtracer.trace("parser.p_tacticwithelem_trigger ENTER (tacticwithelem)")
        p[0] = Trigger(*([Atom(p[2])]+p[4]))
        p[0].lineno = get_lineno(p,3)

    def p_tactwithlist_tacticwithelem(p):
        'tacticwithlist : tacticwithelem'
        if __debug__: xtracer.trace("parser.p_tactwithlist_tacticwithelem ENTER (tacticwithlist)")
        p[0] = [p[1]]

    def p_tactwithlist_tactwithlist_tacticwithelem(p):
        'tacticwithlist : tacticwithlist tacticwithelem'
        if __debug__: xtracer.trace("parser.p_tactwithlist_tactwithlist_tacticwithelem ENTER (tacticwithlist)")
        p[0] = p[1]
        p[0].append(p[2])

    def p_opttacticwith_with_lcb_tacticwithlist_rcb(p):
        'opttacticwith : WITH LCB tacticwithlist RCB'
        if __debug__: xtracer.trace("parser.p_opttacticwith_with_lcb_tacticwithlist_rcb ENTER (opttacticwith)")
        p[0] = TacticWith(*p[3])
        p[0].lineno = get_lineno(p,1)

def p_proofseq_proofstep(p):
    'proofseq : proofstep'
    if __debug__: xtracer.trace("parser.p_proofseq_proofstep ENTER (proofseq)")
    p[0] = p[1]

def p_proofseq_proofseq_semi_proofstep(p):
    'proofseq : proofseq optsemi proofstep'
    if __debug__: xtracer.trace("parser.p_proofseq_proofseq_semi_proofstep ENTER (proofseq)")
    p[0] = ComposeTactics(p[1],p[3])
    p[0].lineno = get_lineno(p,2)

def p_proofstep_proofgroup(p):
    'proofstep : proofgroup'
    if __debug__: xtracer.trace("parser.p_proofstep_proofgroup ENTER (proofstep)")
    p[0] = p[1]

def p_proofgroup_lcb_proofseq_rcb(p):
    'proofgroup : LCB proofseq RCB'
    if __debug__: xtracer.trace("parser.p_proofgroup_lcb_proofseq_rcb ENTER (proofgroup)")
    p[0] = p[2]

def p_proofgroup_lcb_rcb(p):
    'proofgroup : LCB RCB'
    if __debug__: xtracer.trace("parser.p_proofgroup_lcb_rcb ENTER (proofgroup)")
    p[0] = NullTactic()
    p[0].lineno = get_lineno(p,1)

def p_optproof(p):
    'optproof :'
    if __debug__: xtracer.trace("parser.p_optproof ENTER (optproof)")
    p[0] = None

def p_optproof_symbol(p):
    'optproof : PROOF proofstep'
    if __debug__: xtracer.trace("parser.p_optproof_symbol ENTER (optproof)")
    p[0] = p[2]
    
def p_optproof_label_proofstep(p):
    'optproof : PROOF LABEL proofstep'
    if __debug__: xtracer.trace("parser.p_optproof_label_proofstep ENTER (optproof)")
    label = Atom(p[2][1:-1],[])
    label.lineno = get_lineno(p,2)
    p[0] = LabeledFormula(label,p[3])
    p[0].lineno = get_lineno(p,1)
    
def p_optproofgroup(p):
    'optproofgroup :'
    if __debug__: xtracer.trace("parser.p_optproofgroup ENTER (optproofgroup)")
    p[0] = None

def p_optproofgroup_symbol(p):
    'optproofgroup : PROOF proofgroup'
    if __debug__: xtracer.trace("parser.p_optproofgroup_symbol ENTER (optproofgroup)")
    p[0] = p[2]
    
if iu.get_numeric_version() <= [1,6]:
    def p_top_definition_defns(p):
        'top : top DEFINITION defns optproof'
        if __debug__: xtracer.trace("parser.p_top_definition_defns ENTER (top)")
        p[0] = p[1]
        p[0].declare(DefinitionDecl(*[addlabel(mk_lf(x),'def') for x in p[3]]))
        if p[4] is not None:
            p[0].declare(ProofDecl(p[4]))
else:

    def p_optlabel_label(p):
        'optlabel : LABEL'
        if __debug__: xtracer.trace("parser.p_optlabel_label ENTER (optlabel)")
        p[0] = Atom(p[1][1:-1],[])
        p[0].lineno = get_lineno(p,1)

    def p_optlabel(p):
        'optlabel : '
        if __debug__: xtracer.trace("parser.p_optlabel ENTER (optlabel)")
        p[0] = None

    def p_gdefn_defn(p):
        'gdefn : defn'
        if __debug__: xtracer.trace("parser.p_gdefn_defn ENTER (gdefn)")
        p[0] = p[1]

    def p_gdefn_lcb_defn_rcb(p):
        'gdefn : LCB defn RCB'
        if __debug__: xtracer.trace("parser.p_gdefn_lcb_defn_rcb ENTER (gdefn)")
        p[0] = DefinitionSchema(*p[2].args)
        p[0].lineno = p[2].lineno

    def p_top_definition_optlabel_gdefn_optproof(p):
        'top : top optexplicit DEFINITION optlabel gdefn optproof'
        if __debug__: xtracer.trace("parser.p_top_definition_optlabel_gdefn_optproof ENTER (top)")
        foo = p[5]
        if p[2]:
            foo = DefinitionSchema(*foo.args)
            foo.lineno = p[5].lineno
        lf = LabeledFormula(p[4],foo)
        lf.lineno = get_lineno(p,3)
        p[0] = p[1]
        p[0].declare(DefinitionDecl(addlabel(lf,'def')))
        if p[6] is not None:
            p[0].declare(ProofDecl(p[6]))


def p_top_progress_defns(p):
    'top : top PROGRESS defns'
    if __debug__: xtracer.trace("parser.p_top_progress_defns ENTER (top)")
    p[0] = p[1]
    p[0].declare(ProgressDecl(*p[3]))

def p_top_rely_atom_arrow_atom(p):
    'top : top RELY atom ARROW atom'
    if __debug__: xtracer.trace("parser.p_top_rely_atom_arrow_atom ENTER (top)")
    p[0] = p[1]
    p[0].declare(RelyDecl(Implies(p[3],p[5])))

def p_top_mixord_callatom_arrow_callatom(p):
    'top : top MIXORD callatom ARROW callatom'
    if __debug__: xtracer.trace("parser.p_top_mixord_callatom_arrow_callatom ENTER (top)")
    p[0] = p[1]
    p[0].declare(MixOrdDecl(Implies(p[3],p[5])))

def p_top_rely_atom(p):
    'top : top RELY atom'
    if __debug__: xtracer.trace("parser.p_top_rely_atom ENTER (top)")
    p[0] = p[1]
    p[0].declare(RelyDecl(p[3]))

def p_top_concept_cdefns(p):
    'top : top CONCEPT cdefns'
    if __debug__: xtracer.trace("parser.p_top_concept_cdefns ENTER (top)")
    p[0] = p[1]
    p[0].declare(ConceptDecl(*p[3]))

# init statement is banned as of version 1.7
if iu.get_numeric_version() <= [1,6]:
    def p_top_init_fmla(p):
        'top : top INIT labeledfmla'
        if __debug__: xtracer.trace("parser.p_top_init_fmla ENTER (top)")
        p[0] = p[1]
        d = InitDecl(check_non_temporal(p[3]))
        d.lineno = get_lineno(p,2)
        p[0].declare(d)


def p_top_update_terms_from_terms_upaxes(p):
    'top : top UPDATE apps FROM apps upaxes'
    if __debug__: xtracer.trace("parser.p_top_update_terms_from_terms_upaxes ENTER (top)")
    p[0] = p[1]
    dfns = [x.rep for x in p[3]]
    deps = [x.rep for x in p[5]]
    p[0].declare(UpdateDecl(PatternBasedUpdate(SymbolList(*dfns),
                                               SymbolList(*deps),
                                               UpdatePatternList(*p[6]))))

def p_optfinite(p):
    'optfinite : '
    if __debug__: xtracer.trace("parser.p_optfinite ENTER (optfinite)")
    p[0] = False

def p_optfinite_finite(p):
    'optfinite : FINITE'
    if __debug__: xtracer.trace("parser.p_optfinite_finite ENTER (optfinite)")
    p[0] = True

def p_optghost(p):
    'optghost : '
    if __debug__: xtracer.trace("parser.p_optghost ENTER (optghost)")
    p[0] = False

def p_optghost_ghost(p):
    'optghost : GHOST'
    if __debug__: xtracer.trace("parser.p_optghost_ghost ENTER (optghost)")
    p[0] = True

def p_typesymbol_symbol(p):
    'typesymbol : SYMBOL'
    if __debug__: xtracer.trace("parser.p_typesymbol_symbol ENTER (typesymbol)")
    p[0] = p[1]

def p_typesymbol_this(p):
    'typesymbol : THIS'
    if __debug__: xtracer.trace("parser.p_typesymbol_this ENTER (typesymbol)")
    p[0] = This()
    p[0].lineno = get_lineno(p,1)

def p_top_type_symbol(p):
    'top : top optfinite optghost TYPE typesymbol'
    if __debug__: xtracer.trace("parser.p_top_type_symbol ENTER (top)")
    p[0] = p[1]
    scnst = Atom(p[5])
    scnst.lineno = get_lineno(p,5)
    tdfn = (GhostTypeDef if p[3] else TypeDef)(scnst,UninterpretedSort())
    if p[2]:
        tdfn.finite = True
    tdfn.lineno = get_lineno(p,4)
    p[0].declare(TypeDecl(tdfn))

def p_top_type_symbol_eq_sort(p):
    'top : top optfinite optghost TYPE typesymbol EQ sort'
    if __debug__: xtracer.trace("parser.p_top_type_symbol_eq_sort ENTER (top)")
    p[0] = p[1]
    scnst = Atom(p[5])
    scnst.lineno = get_lineno(p,5)
    defsort = UninterpretedSort() if isinstance(p[7],Range) else p[7]
    tdfn = (GhostTypeDef if p[3] else TypeDef)(scnst,defsort)
    if p[2]:
        tdfn.finite = True
    tdfn.lineno = get_lineno(p,6)
    p[0].declare(TypeDecl(tdfn))
    if isinstance(p[7],Range):
        imp = Implies(scnst,p[7])
        imp.lineno = get_lineno(p,4)
        thing = InterpretDecl(addlabel(mk_lf(imp),'interp'))
        thing.lineno = get_lineno(p,4)
        p[0].declare(thing)

    
def p_tsyms_tsym(p):
    'tsyms : var'
    if __debug__: xtracer.trace("parser.p_tsyms_tsym ENTER (tsyms)")
    p[0] = [p[1]]

def p_tsyms_tsyms_comma_tsym(p):
    'tsyms : tsyms COMMA var'
    if __debug__: xtracer.trace("parser.p_tsyms_tsyms_comma_tsym ENTER (tsyms)")
    p[0] = p[1]
    p[0].append(p[3])

def p_targs_lparen_rparen(p):
    'targs : LPAREN RPAREN'
    if __debug__: xtracer.trace("parser.p_targs_lparen_rparen ENTER (targs)")
    p[0] = []

def p_targs_lparen_tsyms_rparen(p):
    'targs : LPAREN tsyms RPAREN'
    if __debug__: xtracer.trace("parser.p_targs_lparen_tsyms_rparen ENTER (targs)")
    p[0] = p[2]

# if iu.get_numeric_version() <= [1,5]:
#     def p_param_term_colon_symbol(p):
#         'param : SYMBOL COLON SYMBOL'
#         p[0] = App(p[1])
#         p[0].lineno = get_lineno(p,1)
#         p[0].sort = p[3]
# else:
def p_param_term_colon_symbol(p):
    'param : SYMBOL COLON SYMBOL'
    if __debug__: xtracer.trace("parser.p_param_term_colon_symbol ENTER (param)")
    p[0] = App(p[1])
    p[0].lineno = get_lineno(p,1)
    p[0].sort = p[3]


def p_params_param(p):
    'params : param'
    if __debug__: xtracer.trace("parser.p_params_param ENTER (params)")
    p[0] = [p[1]]

def p_params_params_comma_param(p):
    'params : params COMMA param'
    if __debug__: xtracer.trace("parser.p_params_params_comma_param ENTER (params)")
    p[0] = p[1]
    p[0].append(p[3])

def p_optargs(p):
    'optargs : '
    if __debug__: xtracer.trace("parser.p_optargs ENTER (optargs)")
    p[0] = []

def p_optargs_params(p):
    'optargs : LPAREN lparams RPAREN'
    if __debug__: xtracer.trace("parser.p_optargs_params ENTER (optargs)")
    p[0] = p[2]

def p_optreturns(p):
    'optreturns :'
    if __debug__: xtracer.trace("parser.p_optreturns ENTER (optreturns)")
    p[0] = []
    
def p_optreturns_tsyms(p):
    'optreturns : RETURNS LPAREN lparams RPAREN'
    if __debug__: xtracer.trace("parser.p_optreturns_tsyms ENTER (optreturns)")
    p[0] = p[3]

def p_optactualreturns(p):
    'optactualreturns :'
    if __debug__: xtracer.trace("parser.p_optactualreturns ENTER (optactualreturns)")
    p[0] = []

def p_optactualreturns_callatoms_assign(p):
    'optactualreturns : callatoms ASSIGN'
    if __debug__: xtracer.trace("parser.p_optactualreturns_callatoms_assign ENTER (optactualreturns)")
    p[0] = p[1]

def p_tapp_symbol(p):
    'tapp : SYMBOL'
    if __debug__: xtracer.trace("parser.p_tapp_symbol ENTER (tapp)")
    p[0] = App(p[1])
    p[0].lineno = get_lineno(p,1)

def p_tapp_symbol_targs(p):
    'tapp : SYMBOL targs'
    if __debug__: xtracer.trace("parser.p_tapp_symbol_targs ENTER (tapp)")
    p[0] = App(p[1],*p[2])
    p[0].lineno = get_lineno(p,1)

def p_tapp_lp_symbol_infix_symbol_rp(p):
    'tapp : LPAREN var infix var RPAREN'
    if __debug__: xtracer.trace("parser.p_tapp_lp_symbol_infix_symbol_rp ENTER (tapp)")
    p[0] = App(p[3],p[2],p[4])
    p[0].lineno = get_lineno(p,3)


def p_tterm_term(p):
    'tterm : tapp'
    if __debug__: xtracer.trace("parser.p_tterm_term ENTER (tterm)")
    p[0] = p[1]

def p_tterm_term_colon_symbol(p):
    'tterm : tapp COLON atype'
    if __debug__: xtracer.trace("parser.p_tterm_term_colon_symbol ENTER (tterm)")
    p[0] = p[1]
    p[0].sort = p[3]

def p_tterms_tterm(p):
    'tterms : tterm'
    if __debug__: xtracer.trace("parser.p_tterms_tterm ENTER (tterms)")
    p[0] = [p[1]]

def p_tterms_tterms_comma_tterm(p):
    'tterms : tterms COMMA tterm'
    if __debug__: xtracer.trace("parser.p_tterms_tterms_comma_tterm ENTER (tterms)")
    p[0] = p[1]
    p[0].append(p[3])

def p_sort_lcb_symbol_rcb(p):
    'sort : LCB SYMBOL RCB'
    if __debug__: xtracer.trace("parser.p_sort_lcb_symbol_rcb ENTER (sort)")
    p[0] = EnumeratedSort(Atom(p[2]))

def p_sort_lcb_names_rcb(p):
    'sort : LCB SYMBOL COMMA names RCB'
    if __debug__: xtracer.trace("parser.p_sort_lcb_names_rcb ENTER (sort)")
    p[0] = EnumeratedSort(*([Atom(p[2])] + [Atom(n) for n in p[4]]))

def p_sort_lcb_symbol_dots_symbol_rcb(p):
    'sort : LCB SYMBOL DOTS SYMBOL RCB'
    if __debug__: xtracer.trace("parser.p_sort_lcb_symbol_dots_symbol_rcb ENTER (sort)")
    p[0] = Range(Atom(p[2]),Atom(p[4]))
    p[0].lineno = get_lineno(p,2)

def p_sort_struct_lcb_names_rcb(p):
    'sort : STRUCT LCB tterms RCB'
    if __debug__: xtracer.trace("parser.p_sort_struct_lcb_names_rcb ENTER (sort)")
    p[0] = StructSort(*p[3])

def p_sort_struct_lcb_rcb(p):
    'sort : STRUCT LCB RCB'
    if __debug__: xtracer.trace("parser.p_sort_struct_lcb_rcb ENTER (sort)")
    p[0] = StructSort()

def p_names_symbol(p):
    'names : SYMBOL'
    if __debug__: xtracer.trace("parser.p_names_symbol ENTER (names)")
    p[0] = [p[1]]

def p_names_names_comma_symbol(p):
    'names : names COMMA SYMBOL'
    if __debug__: xtracer.trace("parser.p_names_names_comma_symbol ENTER (names)")
    p[0] = p[1]
    p[0].append(p[3])

def p_upaxes(p):
    'upaxes : '
    if __debug__: xtracer.trace("parser.p_upaxes ENTER (upaxes)")
    p[0] = []

def p_upaxes_upaxes_upax(p):
    'upaxes : upaxes upax'
    if __debug__: xtracer.trace("parser.p_upaxes_upaxes_upax ENTER (upaxes)")
    p[0] = p[1]
    p[0].append(p[2])

if True or iu.get_numeric_version() <= [1]:
    def p_upax_params_apps_in_action_arrow_ensures_fmla(p):
        'upax : PARAMS tterms IN action ARROW requires ensures'
        if __debug__: xtracer.trace("parser.p_upax_params_apps_in_action_arrow_ensures_fmla ENTER (upax)")
        p[0] = UpdatePattern(ConstantDecl(*p[2]),p[4],p[6],p[7])
else:
    def p_upax_params_apps_in_action_ensures_fmla(p):
        'upax : PARAMS tterms IN action requires ensures'
        if __debug__: xtracer.trace("parser.p_upax_params_apps_in_action_ensures_fmla ENTER (upax)")
        p[0] = UpdatePattern(ConstantDecl(*p[2]),p[4],p[5],p[6])


def p_requires(p):
    'requires : '
    if __debug__: xtracer.trace("parser.p_requires ENTER (requires)")
    p[0] = And()

def p_requires_requires_fmla(p):
    'requires : REQUIRES fmla'
    if __debug__: xtracer.trace("parser.p_requires_requires_fmla ENTER (requires)")
    p[0] = check_non_temporal(p[2])

def p_modifies(p):
    'modifies : '
    if __debug__: xtracer.trace("parser.p_modifies ENTER (modifies)")
    p[0] = None

def p_modifies_modifies_lcb_rcb(p):
    'modifies : MODIFIES LCB RCB'
    if __debug__: xtracer.trace("parser.p_modifies_modifies_lcb_rcb ENTER (modifies)")
    p[0] = []

def p_modifies_modofies_times(p):
    'modifies : MODIFIES TIMES'
    if __debug__: xtracer.trace("parser.p_modifies_modofies_times ENTER (modifies)")
    p[0] = None

def p_modifies_modifies_atoms(p):
    'modifies : MODIFIES atoms'
    if __debug__: xtracer.trace("parser.p_modifies_modifies_atoms ENTER (modifies)")
    p[0] = p[2]

# def p_ensures(p):
#     'ensures : '
#     p[0] = And()

def p_ensures_ensures_fmla(p):
    'ensures : ENSURES fmla'
    if __debug__: xtracer.trace("parser.p_ensures_ensures_fmla ENTER (ensures)")
    p[0] = check_non_temporal(p[2])

if iu.get_numeric_version() <= [1,1]:
  def p_top_action_symbol_eq_loc_action_loc(p):
    'top : top ACTION SYMBOL loc EQ sequence loc'
    if __debug__: xtracer.trace("parser.p_top_action_symbol_eq_loc_action_loc ENTER (top)")
    p[0] = p[1]
    p[0].declare(ActionDecl(ActionDef(Atom(p[3],[]),p[6])))
else:

  def p_optactiondef(p):
    'optactiondef : '
    if __debug__: xtracer.trace("parser.p_optactiondef ENTER (optactiondef)")
    p[0] = Sequence()

  def p_topseq_sequence(p):
    'topseq : sequence'
    if __debug__: xtracer.trace("parser.p_topseq_sequence ENTER (topseq)")
    p[0] = p[1]

  def p_topseq_lcb_nativequote_rcb(p):
    'topseq : LCB NATIVEQUOTE RCB'
    if __debug__: xtracer.trace("parser.p_topseq_lcb_nativequote_rcb ENTER (topseq)")
    text,bqs = parse_nativequote(p,2)
    p[0] = NativeAction(*([text] + bqs))
    p[0].lineno = get_lineno(p,2)

  def p_optactiondef_eq_topseq(p):
    'optactiondef : EQ topseq'
    if __debug__: xtracer.trace("parser.p_optactiondef_eq_topseq ENTER (optactiondef)")
    p[0] = p[2]

  def p_optactiondef_eq_symbol(p):
    'optactiondef : EQ TIMES'
    if __debug__: xtracer.trace("parser.p_optactiondef_eq_symbol ENTER (optactiondef)")
    p[0] = CrashAction()
    p[0].lineno = get_lineno(p,2)

  def p_optimpex(p):
      'optimpex : '
      if __debug__: xtracer.trace("parser.p_optimpex ENTER (optimpex)")
      p[0] = None

  def p_optimpex_export(p):
      'optimpex : EXPORT'
      if __debug__: xtracer.trace("parser.p_optimpex_export ENTER (optimpex)")
      p[0] = ExportDecl
      
  def p_optimpex_import(p):
      'optimpex : IMPORT'
      if __debug__: xtracer.trace("parser.p_optimpex_import ENTER (optimpex)")
      p[0] = ImportDecl

  def p_actmeth_action(p):
      'actmeth : ACTION'
      if __debug__: xtracer.trace("parser.p_actmeth_action ENTER (actmeth)")
      p[0] = False

  def p_actmeth_method(p):
      'actmeth : METHOD'
      if __debug__: xtracer.trace("parser.p_actmeth_method ENTER (actmeth)")
      p[0] = True

  def p_top_optimpex_action_symbol_optargs_optreturns_eq_action(p):
    'top : top optimpex actmeth SYMBOL optargs optreturns optactiondef'
    if __debug__: xtracer.trace("parser.p_top_optimpex_action_symbol_optargs_optreturns_eq_action ENTER (top)")
    p[0] = p[1]
    adef = p[7]
    if not hasattr(adef,'lineno'):
        adef.lineno = get_lineno(p,4)
    formals = p[5]
    if p[3]:
        arg0 = App('self')
        arg0.sort = This()
        arg0.lineno = get_lineno(p,4)
        formals = [arg0] + formals
    if isinstance(adef,CrashAction):
        adef = adef.clone([Atom(This(),formals)])
    the_atom = Atom(p[4],[])
    the_atom.lineno = adef.lineno
    actdef = ActionDef(the_atom,adef,formals=formals,returns=p[6])
    actdef.lineno = adef.lineno
    decl = ActionDecl(actdef)
    decl.lineno = adef.lineno
    p[0].declare(decl)
    for foo in decl.args:
        if not hasattr(foo.args[1],'lineno'):
            print('no lineno!!!: {}'.format(foo))
    if p[2]:
        if p[2] == ExportDecl:
            d = ExportDecl(ExportDef(Atom(p[4]),Atom('')))
        else:
            d = ImportDecl(ImportDef(Atom(p[4]),Atom('')))
        d.lineno = get_lineno(p,4)
        p[0].declare(d)


def handle_mixin(kind,mixer,mixee,ivy):
    if __debug__: xtracer.trace("parser.handle_mixin ENTER")
    cls = (MixinBeforeDef if kind == 'before' else MixinAfterDef if kind == 'after' else MixinImplementDef)
    m = cls(mixer,mixee)
    m.lineno = mixer.lineno
    d = MixinDecl(m)
    d.lineno = mixer.lineno
    ivy.declare(d)


def infer_action_params(actname,formals,returns):
    if __debug__: xtracer.trace("parser.infer_action_params ENTER")
    mixee,num_params = stack_action_lookup(actname)
    if not mixee:
        return formals,returns
    if ("common" in mixee.attributes) != ("common" in stack[-1].attributes):
        return formals,returns
    mformals,mreturns = mixee.formals()
    formals.extend(mformals[num_params+len(formals):])
    returns.extend(mreturns[len(returns):])
    return formals,returns

def handle_before_after(kind,atom,action,ivy,optargs=[],optreturns=[]):
    if __debug__: xtracer.trace("parser.handle_before_after ENTER")
    if atom.args:  # no args -- we get them from the matching action
        report_error(IvyError(atom,"syntax error"))
    else:
        mixer = make_mixin_name(atom,kind)
        optargs,optreturns = infer_action_params(atom.rep,optargs,optreturns)
        df = ActionDef(mixer,action,formals=optargs,returns=optreturns)
        df.lineno = atom.lineno
        ivy.declare(ActionDecl(df))
        handle_mixin(kind,mixer,atom,ivy)
    
if not (iu.get_numeric_version() <= [1,1]):
    def p_top_mixin_callatom_before_callatom(p):
        'top : top MIXIN callatom BEFORE callatom'
        if __debug__: xtracer.trace("parser.p_top_mixin_callatom_before_callatom ENTER (top)")
        p[0] = p[1]
        handle_mixin("before",p[3],p[5],p[0])
    def p_top_mixin_callatom_after_callatom(p):
        'top : top MIXIN callatom AFTER callatom'
        if __debug__: xtracer.trace("parser.p_top_mixin_callatom_after_callatom ENTER (top)")
        p[0] = p[1]
        handle_mixin("after",p[3],p[5],p[0])
    def p_top_before_callatom_lcb_action_rcb(p):
        'top : top BEFORE atype optargs optreturns sequence'
        if __debug__: xtracer.trace("parser.p_top_before_callatom_lcb_action_rcb ENTER (top)")
        p[0] = p[1]
        atom = Atom(p[3])
        atom.lineno = get_lineno(p,2)
        handle_before_after("before",atom,p[6],p[0],p[4],p[5])
    def p_top_after_callatom_lcb_action_rcb(p):
        'top : top AFTER atype optargs optreturns topseq'
        if __debug__: xtracer.trace("parser.p_top_after_callatom_lcb_action_rcb ENTER (top)")
        p[0] = p[1]
        atom = Atom(p[3])
        atom.lineno = get_lineno(p,2)
        handle_before_after("after",atom,p[6],p[0],p[4],p[5])

    if not (iu.get_numeric_version() <= [1,6]):
        def stmt_to_seq(stmts,p,n):
            if __debug__: xtracer.trace("parser.stmt_to_seq ENTER")
            stmts = lower_var_stmts(stmts)
            if len(stmts) == 1:
                return stmts[0]
            else:
                res = Sequence(*stmts)
                res.lineno = stmts[0].lineno # get_lineno(p,n)
                return res
        def p_top_around_callatom_lcb_action_rcb(p):
            'top : top AROUND atype optargs optreturns LCB actseq optsemi DOTDOTDOT actseq optsemi RCB'
            if __debug__: xtracer.trace("parser.p_top_around_callatom_lcb_action_rcb ENTER (top)")
            before = stmt_to_seq(p[7],p,7)
            after = stmt_to_seq(p[10],p,10)
            p[0] = p[1]
            atom = Atom(p[3])
            atom.lineno = get_lineno(p,2)
            handle_before_after("before",atom,before,p[0],p[4],p[5])
            handle_before_after("after",atom,after,p[0],p[4],p[5])

    def p_top_after_init_optargs_lcb_action_rcb(p):
        'top : top AFTER INIT optargs topseq'
        if __debug__: xtracer.trace("parser.p_top_after_init_optargs_lcb_action_rcb ENTER (top)")
        p[0] = p[1]
        atom = Atom("init")
        atom.lineno = get_lineno(p,2)
        handle_before_after("after",atom,p[5],p[0],p[4],[])
    def p_top_implement_callatom_lcb_action_rcb(p):
        'top : top IMPLEMENT atype optargs optreturns topseq'
        if __debug__: xtracer.trace("parser.p_top_implement_callatom_lcb_action_rcb ENTER (top)")
        p[0] = p[1]
        atom = Atom(p[3])
        atom.lineno = get_lineno(p,2)
        handle_before_after("implement",atom,p[6],p[0],p[4],p[5])
    def p_top_implement_type_symbol_with_symbol(p):
        'top : top IMPLEMENT TYPE SYMBOL WITH SYMBOL'
        if __debug__: xtracer.trace("parser.p_top_implement_type_symbol_with_symbol ENTER (top)")
        a1,a2 = Atom(p[4]),Atom(p[6])
        a1.lineno = get_lineno(p,4)
        a2.lineno = get_lineno(p,6)
        impl = ImplementTypeDef(a1,a2)
        impl.lineno = get_lineno(p,5)
        d = ImplementTypeDecl(mk_lf(impl))
        d.lineno = get_lineno(p,2)
        p[0] = p[1]
        p[0].declare(d)
    def p_opttrusted(p):
        'opttrusted :'
        if __debug__: xtracer.trace("parser.p_opttrusted ENTER (opttrusted)")
        p[0] = False
    def p_opttrusted_trusted(p):
        'opttrusted : TRUSTED'
        if __debug__: xtracer.trace("parser.p_opttrusted_trusted ENTER (opttrusted)")
        p[0] = True
    def p_top_opttrusted_isolate_callatom_eq_callatoms(p):
        'top : top opttrusted ISOLATE SYMBOL optargs EQ callatoms'
        if __debug__: xtracer.trace("parser.p_top_opttrusted_isolate_callatom_eq_callatoms ENTER (top)")
        ty = TrustedIsolateDef if p[2] else IsolateDef
        d = IsolateDecl(ty(*([Atom(p[4],p[5])] + p[7])))
        d.args[0].with_args = 0
        d.args[0].lineno = get_lineno(p,3)
        d.lineno = get_lineno(p,3)
        p[0] = p[1]
        p[0].declare(d)
    def p_top_opttrusted_isolate_callatom_eq_callatoms_with_callatoms(p):
        'top : top opttrusted ISOLATE SYMBOL optargs EQ callatoms WITH callatoms'
        if __debug__: xtracer.trace("parser.p_top_opttrusted_isolate_callatom_eq_callatoms_with_callatoms ENTER (top)")
        ty = TrustedIsolateDef if p[2] else IsolateDef
        d = IsolateDecl(ty(*([Atom(p[4],p[5])] + p[7] + p[9])))
        d.args[0].with_args = len(p[9])
        d.args[0].lineno = get_lineno(p,3)
        d.lineno = get_lineno(p,3)
        p[0] = p[1]
        p[0].declare(d)
    def p_optwith(p):
        'optwith : '
        if __debug__: xtracer.trace("parser.p_optwith ENTER (optwith)")
        p[0] = []
    def p_optwith_with_callatoms(p):
        'optwith : WITH callatoms'
        if __debug__: xtracer.trace("parser.p_optwith_with_callatoms ENTER (optwith)")
        p[0] = p[2]
    def p_top_opttrusted_isolate_callatom_eq_lcb_top_rcb_optwith(p):
        'top : top opttrusted ISOLATE SYMBOL optargs EQ LCB top RCB optwith'
        if __debug__: xtracer.trace("parser.p_top_opttrusted_isolate_callatom_eq_lcb_top_rcb_optwith ENTER (top)")
        p[0] = p[1]
        create_object(p[0],p[4],p[5],p[8],get_lineno(p,4))
        ty = TrustedIsolateDef if p[2] else IsolateDef
        df = ty(*([Atom(p[4],p[5]),Atom(p[4],p[5])]+p[10]))
        df.is_object = True
        d = IsolateObjectDecl(df)
        d.args[0].with_args = len(p[10])
        d.args[0].lineno = get_lineno(p,3)
        d.lineno = get_lineno(p,3)
        p[0].declare(d)
    def p_top_opttrusted_extract_callatom_eq_lcb_top_rcb_optwith(p):
        'top : top EXTRACT objsym objectargs EQ LCB top RCB optwith'
        if __debug__: xtracer.trace("parser.p_top_opttrusted_extract_callatom_eq_lcb_top_rcb_optwith ENTER (top)")
        p[0] = p[1]
        create_object(p[0],p[3],p[4],p[7],get_lineno(p,3))
        ty = ProcessDef
        d = IsolateObjectDecl(ty(*([Atom(p[3],p[4]),Atom(p[3],p[4])]+p[9])))
        d.args[0].with_args = len(p[9])+1
        d.args[0].lineno = get_lineno(p,2)
        d.lineno = get_lineno(p,2)
        p[0].declare(d)
    def p_top_extract_callatom_eq_callatoms(p):
        'top : top EXTRACT objsym objectargs EQ callatoms'
        if __debug__: xtracer.trace("parser.p_top_extract_callatom_eq_callatoms ENTER (top)")
        stack[-1].params = []
        global parent_object
        parent_object = None
        d = IsolateDecl(ExtractDef(*([Atom(p[3],p[4])] + p[6])))
        d.args[0].with_args = len(p[6])
        d.args[0].lineno = get_lineno(p,2)
        d.lineno = get_lineno(p,2)
        p[0] = p[1]
        p[0].declare(d)
    def p_top_export_callatom(p):
        'top : top EXPORT callatom'
        if __debug__: xtracer.trace("parser.p_top_export_callatom ENTER (top)")
        d = ExportDecl(ExportDef(p[3],Atom('')))
        d.lineno = get_lineno(p,2)
        d.args[0].lineno = d.lineno
        p[0] = p[1]
        p[0].declare(d)
    def p_top_import_callatom(p):
        'top : top IMPORT callatom'
        if __debug__: xtracer.trace("parser.p_top_import_callatom ENTER (top)")
        d = ImportDecl(ImportDef(p[3],Atom('')))
        d.lineno = get_lineno(p,2)
        d.args[0].lineno = d.lineno
        p[0] = p[1]
        p[0].declare(d)
    if iu.get_numeric_version() <= [1,6]:
        def p_top_private_callatom(p):
            'top : top PRIVATE callatom'
            if __debug__: xtracer.trace("parser.p_top_private_callatom ENTER (top)")
            d = PrivateDecl(PrivateDef(p[3]))
            d.lineno = get_lineno(p,2)
            p[0] = p[1]
            p[0].declare(d)
    def p_optdelegee(p):
        'optdelegee :'
        if __debug__: xtracer.trace("parser.p_optdelegee ENTER (optdelegee)")
        p[0] = None
    def p_optdelegee_callatom(p):
        'optdelegee : ARROW callatom'
        if __debug__: xtracer.trace("parser.p_optdelegee_callatom ENTER (optdelegee)")
        p[0] = p[2]
    def p_top_delegate_callatom_opt(p):
        'top : top DELEGATE callatoms optdelegee'
        if __debug__: xtracer.trace("parser.p_top_delegate_callatom_opt ENTER (top)")
        if p[4] is not None:
            d = DelegateDecl(*[DelegateDef(s,p[4]) for s in p[3]])
        else:
            d = DelegateDecl(*[DelegateDef(s) for s in p[3]])
        d.lineno = get_lineno(p,2)
        p[0] = p[1]
        p[0].declare(d)

if not (iu.get_numeric_version() <= [1,6]):

    def p_specimpl_specification(p):
        'specimpl : SPECIFICATION'
        if __debug__: xtracer.trace("parser.p_specimpl_specification ENTER (specimpl)")
        p[0] = p[1]
        global special_attribute
        special_attribute = "spec"

    def p_specimpl_implementation(p):
        'specimpl : IMPLEMENTATION'
        if __debug__: xtracer.trace("parser.p_specimpl_implementation ENTER (specimpl)")
        p[0] = p[1]
        global special_attribute
        special_attribute =  "impl"

    def p_specimpl_private(p):
        'specimpl : PRIVATE'
        if __debug__: xtracer.trace("parser.p_specimpl_private ENTER (specimpl)")
        p[0] = p[1]
        global special_attribute
        special_attribute =  "private"

    def p_specimpl_global(p):
        'specimpl : GLOBAL'
        if __debug__: xtracer.trace("parser.p_specimpl_global ENTER (specimpl)")
        p[0] = p[1]
        global global_attribute
        global_attribute =  "global"

    def p_specimpl_common(p):
        'specimpl : COMMON'
        if __debug__: xtracer.trace("parser.p_specimpl_common ENTER (specimpl)")
        p[0] = p[1]
        global common_attribute
        common_attribute =  "common"

    def p_top_specification_lcb_top_rcb(p):
        'top : top specimpl LCB top RCB'
        if __debug__: xtracer.trace("parser.p_top_specification_lcb_top_rcb ENTER (top)")
        p[0] = p[1]
        stack.pop()
        # Don't reapply the attributes in the parent
        temp_attr = p[0].attributes
        p[0].attributes = ()
        for decl in p[4].decls:
            p[0].declare(decl)
        p[0].attributes = temp_attr

    # def p_top_delegate_callatom(p):
    #     'top : top DELEGATE callatoms ARROW callatom'
    #     d = DelegateDecl(*[DelegateDef(s,p[5]) for s in p[3]])
    #     d.lineno = get_lineno(p,2)
    #     p[0] = p[1]
    #     p[0].declare(d)

def p_top_aliase_symbol_eq_callatom(p):
    'top : top ALIAS SYMBOL EQ callatom'
    if __debug__: xtracer.trace("parser.p_top_aliase_symbol_eq_callatom ENTER (top)")
    d = AliasDecl(Definition(Atom(p[3]),p[5]))
    d.lineno = get_lineno(p,3)
    p[0] = p[1]
    p[0].declare(d)

def p_top_state_symbol_eq_state_expr(p):
    'top : top STATE SYMBOL EQ state_expr'
    if __debug__: xtracer.trace("parser.p_top_state_symbol_eq_state_expr ENTER (top)")
    p[0] = p[1]
    p[0].declare(StateDecl(StateDef(p[3],p[5])))

def p_assert_rhs_lcb_requires_modifies_ensures_rcb(p):
    'assert_rhs : LCB requires modifies ensures RCB'
    if __debug__: xtracer.trace("parser.p_assert_rhs_lcb_requires_modifies_ensures_rcb ENTER (assert_rhs)")
    p[0] = RME(p[2],p[3],p[4])

def p_assert_rhs_fmla(p):
    'assert_rhs : fmla'
    if __debug__: xtracer.trace("parser.p_assert_rhs_fmla ENTER (assert_rhs)")
    p[0] = check_non_temporal(p[1])

if iu.get_numeric_version() <= [1,6]: 
    def p_top_assert_symbol_arrow_assert_rhs(p):
        'top : top ASSERT SYMBOL ARROW assert_rhs'
        if __debug__: xtracer.trace("parser.p_top_assert_symbol_arrow_assert_rhs ENTER (top)")
        p[0] = p[1]
        thing = Implies(Atom(p[3],[]),p[5])
        thing.lineno = get_lineno(p,4)
        p[0].declare(AssertDecl(thing))

def p_oper_symbol(p):
    'oper : atype'
    if __debug__: xtracer.trace("parser.p_oper_symbol ENTER (oper)")
    p[0] = Atom(p[1])

def p_oper_relop(p):
    'oper : relop'
    if __debug__: xtracer.trace("parser.p_oper_relop ENTER (oper)")
    p[0] = Atom(p[1])

def p_oper_infix(p):
    'oper : infix'
    if __debug__: xtracer.trace("parser.p_oper_infix ENTER (oper)")
    p[0] = Atom(p[1])

def p_oper_nativequote(p):
    'oper : NATIVEQUOTE'
    if __debug__: xtracer.trace("parser.p_oper_nativequote ENTER (oper)")
    text,bqs = parse_nativequote(p,1)
    p[0] = NativeType(*([text] + bqs))
    p[0].lineno = get_lineno(p,1)

def p_top_interpret_symbol_arrow_symbol(p):
    'top : top INTERPRET oper ARROW oper'
    if __debug__: xtracer.trace("parser.p_top_interpret_symbol_arrow_symbol ENTER (top)")
    p[0] = p[1]
    impl = Implies(p[3],p[5])
    impl.lineno = get_lineno(p,4)
    thing = InterpretDecl(addlabel(mk_lf(impl),'interp'))
    thing.lineno = get_lineno(p,4)
    p[0].declare(thing)
    
def p_top_interpret_symbol_arrow_lcb_symbol_dots_symbol_rcb(p):
    'top : top INTERPRET oper ARROW LCB term DOTS term RCB'
    if __debug__: xtracer.trace("parser.p_top_interpret_symbol_arrow_lcb_symbol_dots_symbol_rcb ENTER (top)")
    p[0] = p[1]
    imp = Implies(p[3],Range(p[6],p[8]))
    imp.lineno = get_lineno(p,4)
    thing = InterpretDecl(addlabel(mk_lf(imp),'interp'))
    thing.lineno = get_lineno(p,4)
    p[0].declare(thing)

def p_moresymbols(p):
    'moresymbols : '
    if __debug__: xtracer.trace("parser.p_moresymbols ENTER (moresymbols)")
    p[0] = []
    
def p_moresymbols_more_symbols_comma_symbol(p):
    'moresymbols : moresymbols COMMA SYMBOL'
    if __debug__: xtracer.trace("parser.p_moresymbols_more_symbols_comma_symbol ENTER (moresymbols)")
    p[0] = p[1]
    p[0].append(p[3])

def p_top_interpret_symbol_arrow_lcb_symbol_moresymbols_rcb(p):
    'top : top INTERPRET oper ARROW LCB SYMBOL moresymbols RCB'
    if __debug__: xtracer.trace("parser.p_top_interpret_symbol_arrow_lcb_symbol_moresymbols_rcb ENTER (top)")
    p[0] = p[1]
    imp = Implies(p[3],EnumeratedSort(*[Atom(n) for n in ([p[6]]+p[7])]))
    imp.lineno = get_lineno(p,4)
    thing = InterpretDecl(addlabel(mk_lf(imp),'interp'))
    thing.lineno = get_lineno(p,4)
    p[0].declare(thing)

def parse_nativequote(p,n):
    if __debug__: xtracer.trace("parser.parse_nativequote ENTER")
    string = p[n][3:-3] # drop the quotation marks
    fields = string.split('`')
    bqs = [(Atom(This()) if s == 'this' else Atom(s))  for idx,s in enumerate(fields) if idx % 2 == 1]
    text = "`".join([(s if idx % 2 == 0 else str(idx//2)) for idx,s in enumerate(fields)])
    eols = [sum(1 for c in s if c == '\n') for idx,s in enumerate(fields) if idx % 2 == 0]
    seols = 0
    loc = get_lineno(p,n)
    for idx,e in enumerate(eols[:-1]):
        seols += e
        bqs[idx].lineno = iu.Location(loc.filename,loc.line+seols)
    if len(fields) %2 != 1:
        thing = Atom("")
        thing.lineno = loc
        report_error(IvyError(thing,"unterminated back-quote"))
    return NativeCode(text),bqs

def p_top_nativequote(p):
    'top : top NATIVEQUOTE'
    if __debug__: xtracer.trace("parser.p_top_nativequote ENTER (top)")
    p[0] = p[1]
    text,bqs = parse_nativequote(p,2)
    defn = NativeDef(*([mk_label(None,'native')] + [text] + bqs))
    defn.lineno = get_lineno(p,2)
    thing = NativeDecl(defn)
    thing.lineno = get_lineno(p,2)
    p[0].declare(thing)   

def p_top_attributeval_callatom(p):
    'attributeval : callatom'
    if __debug__: xtracer.trace("parser.p_top_attributeval_callatom ENTER (attributeval)")
    p[0] = p[1]

def p_top_attributeval_true(p):
    'attributeval : TRUE'
    if __debug__: xtracer.trace("parser.p_top_attributeval_true ENTER (attributeval)")
    p[0] = Atom('true')
    p[0].lineno = get_lineno(p,1)

def p_top_attributeval_false(p):
    'attributeval : FALSE'
    if __debug__: xtracer.trace("parser.p_top_attributeval_false ENTER (attributeval)")
    p[0] = Atom('false')
    p[0].lineno = get_lineno(p,1)

def p_top_attribute_callatom_eq_attributeval(p):
    'top : top ATTRIBUTE callatom EQ attributeval'
    if __debug__: xtracer.trace("parser.p_top_attribute_callatom_eq_attributeval ENTER (top)")
    p[0] = p[1]
    defn = AttributeDef(p[3],p[5])
    defn.lineno = get_lineno(p,2)
    thing = AttributeDecl(defn)
    thing.lineno = get_lineno(p,2)
    p[0].declare(thing)   

def p_top_variant_symbol_of_atype(p):
    'top : top VARIANT typesymbol OF atype'
    if __debug__: xtracer.trace("parser.p_top_variant_symbol_of_atype ENTER (top)")
    p[0] = p[1]
    scnst = Atom(p[3])
    scnst.lineno = get_lineno(p,3)
    tdfn = TypeDef(scnst,UninterpretedSort())
    tdfn.lineno = get_lineno(p,4)
    p[0].declare(TypeDecl(tdfn))
    vdfn = VariantDef(scnst,Atom(p[5]))
    p[0].declare(VariantDecl(vdfn))

def p_top_variant_symbol_of_symbol_eq_sort(p):
    'top : top VARIANT typesymbol OF atype EQ sort'
    if __debug__: xtracer.trace("parser.p_top_variant_symbol_of_symbol_eq_sort ENTER (top)")
    p[0] = p[1]
    scnst = Atom(p[3])
    scnst.lineno = get_lineno(p,3)
    tdfn = TypeDef(scnst,p[7])
    tdfn.lineno = get_lineno(p,4)
    p[0].declare(TypeDecl(tdfn))
    vdfn = VariantDef(scnst,Atom(p[5]))
    p[0].declare(VariantDecl(vdfn))

def p_places_symbol(p):
    'places : SYMBOL'
    if __debug__: xtracer.trace("parser.p_places_symbol ENTER (places)")
    p[0] = [Atom(p[1])]
    p[0][0].lineno = get_lineno(p,1)

def p_places_places_comma_symbol(p):
    'places : places COMMA SYMBOL'
    if __debug__: xtracer.trace("parser.p_places_places_comma_symbol ENTER (places)")
    p[0] = p[1]
    p[0].append(Atom(p[3]))
    p[0][-1].lineno = get_lineno(p,3)
    
def p_sceninit_arrow_places(p):
    'sceninit : ARROW places'
    if __debug__: xtracer.trace("parser.p_sceninit_arrow_places ENTER (sceninit)")
    p[0] = PlaceList(*p[2])
    p[0].lineno = get_lineno(p,1)

def make_mixin_name(atom,suffix):
    if __debug__: xtracer.trace("parser.make_mixin_name ENTER")
    if iu.get_numeric_version() <= [1,6]:
        return atom.rename(atom.rep.replace(iu.ivy_compose_character,'_')+'[' + suffix + ']')
    global label_counter
    label_counter += 1
    return atom.rename(atom.rep.replace(iu.ivy_compose_character,'_')+'[' + suffix + str(label_counter) + ']')

def p_scenariomixin_before_callatom_lcb_action_rcb(p):
    'scenariomixin : BEFORE atype optargs optreturns sequence'
    if __debug__: xtracer.trace("parser.p_scenariomixin_before_callatom_lcb_action_rcb ENTER (scenariomixin)")
    atom = Atom(p[2])
    atom.lineno = get_lineno(p,2)
    mixer = make_mixin_name(atom,'before')
    optargs,optreturns = infer_action_params(atom.rep,p[3],p[4])
    df = ActionDef(atom,p[5],formals=optargs,returns=optreturns)
    p[0] = ScenarioBeforeMixin(mixer,df)
    p[0].lineno = get_lineno(p,1)

def p_scenariomixin_after_callatom_lcb_action_rcb(p):
    'scenariomixin : AFTER atype optargs optreturns sequence'
    if __debug__: xtracer.trace("parser.p_scenariomixin_after_callatom_lcb_action_rcb ENTER (scenariomixin)")
    atom = Atom(p[2])
    atom.lineno = get_lineno(p,2)
    mixer = make_mixin_name(atom,'after')
    optargs,optreturns = infer_action_params(atom.rep,p[3],p[4])
    optargs,optreturns = infer_action_params(atom.rep,p[3],p[4])
    df = ActionDef(atom,p[5],formals=optargs,returns=optreturns)
    p[0] = ScenarioAfterMixin(mixer,df)
    p[0].lineno = get_lineno(p,1)

def p_scentranss(p):
    'scentranss : '
    if __debug__: xtracer.trace("parser.p_scentranss ENTER (scentranss)")
    p[0] = []

def p_scentranss_scentranss_places_arrow_places_colon_scenariomixin(p):
    'scentranss : scentranss places ARROW places COLON scenariomixin'
    if __debug__: xtracer.trace("parser.p_scentranss_scentranss_places_arrow_places_colon_scenariomixin ENTER (scentranss)")
    p[0] = p[1]
    tr = ScenarioTransition(PlaceList(*p[2]),PlaceList(*p[4]),p[6])
    tr.lineno = get_lineno(p,5)
    p[0].append(tr)
    
def p_scentranss_scentranss_places_colon_scenariomixin(p):
    'scentranss : scentranss places COLON scenariomixin'
    if __debug__: xtracer.trace("parser.p_scentranss_scentranss_places_colon_scenariomixin ENTER (scentranss)")
    p[0] = p[1]
    tr = ScenarioTransition(PlaceList(*p[2]),PlaceList(),p[4])
    tr.lineno = get_lineno(p,3)
    p[0].append(tr)
    
def p_top_scenario_lcb_sceninit_semi_scentranss_rcb(p):
    'top : top SCENARIO LCB sceninit SEMI scentranss RCB'
    if __debug__: xtracer.trace("parser.p_top_scenario_lcb_sceninit_semi_scentranss_rcb ENTER (top)")
    p[0] = p[1]
    sdef = ScenarioDef(*([p[4]] + p[6]))
    sdef.lineno = get_lineno(p,2)
    p[0].declare(ScenarioDecl(sdef))

def p_loc(p):
    'loc : '
    if __debug__: xtracer.trace("parser.p_loc ENTER (loc)")
    p[0] = None

def p_loc_symbol(p):
    'loc : SYMBOL'
    if __debug__: xtracer.trace("parser.p_loc_symbol ENTER (loc)")
    p[0] = p[1]

def p_actseqrev_simpleact(p):
    'actseqrev : simpleact'
    if __debug__: xtracer.trace("parser.p_actseqrev_simpleact ENTER (actseqrev)")
    p[0] = [p[1]]

def p_actseqrev_complexact(p):
    'actseqrev : complexact'
    if __debug__: xtracer.trace("parser.p_actseqrev_complexact ENTER (actseqrev)")
    p[0] = [p[1]]

def p_actseqrev_simpact_semi_actseqrev(p):
    'actseqrev : simpleact SEMI actseqrev'
    if __debug__: xtracer.trace("parser.p_actseqrev_simpact_semi_actseqrev ENTER (actseqrev)")
    p[0] = p[3]
    p[0].append(p[1])

def p_actseqrev_simpact_semi(p):
    'actseqrev : simpleact SEMI'
    if __debug__: xtracer.trace("parser.p_actseqrev_simpact_semi ENTER (actseqrev)")
    p[0] = [p[1]]

def p_actseqrev_complexact_actseqrev(p):
    'actseqrev : complexact actseqrev'
    if __debug__: xtracer.trace("parser.p_actseqrev_complexact_actseqrev ENTER (actseqrev)")
    p[0] = p[2]
    p[0].append(p[1])

def p_actseqrev_complexact_semi_actseqrev(p):
    'actseqrev : complexact SEMI actseqrev'
    if __debug__: xtracer.trace("parser.p_actseqrev_complexact_semi_actseqrev ENTER (actseqrev)")
    p[0] = p[3]
    p[0].append(p[1])

def p_actseqrev_complexact_semi(p):
    'actseqrev : complexact SEMI'
    if __debug__: xtracer.trace("parser.p_actseqrev_complexact_semi ENTER (actseqrev)")
    p[0] = [p[1]]

def p_actseq_actseqrev(p):
    'actseq : actseqrev'
    if __debug__: xtracer.trace("parser.p_actseq_actseqrev ENTER (actseq)")
    p[0] = p[1]
    p[1].reverse()

def _lvs_canon(node):
    if hasattr(node, 'canon'):
        return node.canon()
    return type(node).__name__

def lower_var_stmts(stmts):
    if __debug__: xtracer.trace("parser.lower_var_stmts ENTER in=%d canons=[%s]" % (len(stmts), ' '.join(_lvs_canon(s) for s in stmts)))
    for idx,stmt in enumerate(stmts):
        if isinstance(stmt,VarAction):
            lhs = stmt.args[0]
            rhs = stmt.args[1] if len(stmt.args) > 1 else None
            lsym = lhs.prefix('loc:')
            subst = {lhs.rep:lsym.rep}
            lines = lower_var_stmts(stmts[idx+1:])
            lines = [subst_prefix_atoms_ast(s,subst,None,None) for s in lines]
            if rhs is not None:
                asgn = AssignAction(lsym,rhs)
                asgn.lineno = stmt.lineno
            else:
                asgn = lsym
            body = Sequence(*lines)
            body.lineno = stmt.lineno
            res = LocalAction(*[asgn,body],caller="parser.lower_var")
            res.lineno = body.lineno;
            result = stmts[:idx] + [res]
            if __debug__: xtracer.trace("parser.lower_var_stmts RETURN out=%d canons=[%s]" % (len(result), ' '.join(_lvs_canon(s) for s in result)))
            return result
        if isinstance(stmt,ThunkAction):
            name = stmt.args[1].rep
            lname = 'loc:'+name
            subst = {name:lname}
            lines = lower_var_stmts(stmts[idx+1:])
            lines = [subst_prefix_atoms_ast(s,subst,None,None) for s in lines]
            result = stmts[:idx] + [stmt.clone(stmt.args + [Sequence(*lines)])]
            if __debug__: xtracer.trace("parser.lower_var_stmts RETURN out=%d canons=[%s]" % (len(result), ' '.join(_lvs_canon(s) for s in result)))
            return result
    if __debug__: xtracer.trace("parser.lower_var_stmts RETURN out=%d canons=[%s]" % (len(stmts), ' '.join(_lvs_canon(s) for s in stmts)))
    return stmts

def p_sequence_lcb_rcb(p):
    'sequence : LCB RCB'
    if __debug__: xtracer.trace("parser.p_sequence_lcb_rcb ENTER (sequence)")
    p[0] = Sequence()
    p[0].lineno = get_lineno(p,1)

def p_sequence_lcb_actseq_rcb(p):
    'sequence : LCB actseq RCB'
    if __debug__: xtracer.trace("parser.p_sequence_lcb_actseq_rcb ENTER (sequence)")
    stmts = lower_var_stmts(p[2])
    if len(stmts) == 1:
        p[0] = stmts[0]
    else:
        p[0] = Sequence(*lower_var_stmts(stmts))
        p[0].lineno = get_lineno(p,1)

def p_sequence_lcb_actseq_semi_rcb(p):
    'sequence : LCB actseq SEMI RCB'
    if __debug__: xtracer.trace("parser.p_sequence_lcb_actseq_semi_rcb ENTER (sequence)")
    p[0] = Sequence(*lower_var_stmts(p[2]))
    p[0].lineno = get_lineno(p,1)

def p_action_sequence(p):
    'complexact : sequence'
    if __debug__: xtracer.trace("parser.p_action_sequence ENTER (complexact)")
    p[0] = p[1]

def p_action_simpleact(p):
    'action : simpleact'
    if __debug__: xtracer.trace("parser.p_action_simpleact ENTER (action)")
    p[0] = p[1]
    
def p_action_complexact(p):
    'action : complexact'
    if __debug__: xtracer.trace("parser.p_action_complexact ENTER (action)")
    p[0] = p[1]

def p_action_assume(p):
    'simpleact : ASSUME labeledfmla'
    if __debug__: xtracer.trace("parser.p_action_assume ENTER (simpleact)")
    p[0] = AssumeAction(check_non_temporal(addlabel(p[2],'asrt')))
    p[0].lineno = get_lineno(p,1)


if iu.get_numeric_version() <= [1,6]:
    def p_action_assert(p):
        'simpleact : ASSERT labeledfmla'
        if __debug__: xtracer.trace("parser.p_action_assert ENTER (simpleact)")
        p[0] = AssertAction(check_non_temporal(addlabel(p[2],'asrt')))
        p[0].lineno = get_lineno(p,1)

    def p_action_ensures(p):
        'simpleact : ENSURES labeledfmla'
        if __debug__: xtracer.trace("parser.p_action_ensures ENTER (simpleact)")
        p[0] = EnsuresAction(check_non_temporal(addlabel(p[2],'asrt')))
        p[0].lineno = get_lineno(p,1)
else:
    def p_action_assert(p):
         'simpleact : optunprovable ASSERT labeledfmla'
         if __debug__: xtracer.trace("parser.p_action_assert ENTER (simpleact)")
         p[0] = AssertAction(check_non_temporal(addlabel(p[3],'asrt')))
         addunprovable(p[0].args[0],p[1])
         if p[1] and not check_unprovable.get():
             p[0] = Sequence()
         p[0].lineno = get_lineno(p,2)

    def p_action_assert_proof_proofstep(p):
         'simpleact : optunprovable ASSERT labeledfmla PROOF proofstep'
         if __debug__: xtracer.trace("parser.p_action_assert_proof_proofstep ENTER (simpleact)")
         p[0] = AssertAction(check_non_temporal(addlabel(p[3],'asrt')),p[5])
         addunprovable(p[0].args[0],p[1])
         if p[1] and not check_unprovable.get():
             p[0] = Sequence()
         p[0].lineno = get_lineno(p,2)

    def p_action_ensure(p):
         'simpleact : optunprovable ENSURE labeledfmla'
         if __debug__: xtracer.trace("parser.p_action_ensure ENTER (simpleact)")
         p[0] = EnsuresAction(check_non_temporal(addlabel(p[3],'asrt')))
         addunprovable(p[0].args[0],p[1])
         if p[1] and not check_unprovable.get():
             p[0] = Sequence()
         p[0].lineno = get_lineno(p,2)

    def p_action_ensure_proof_proofstep(p):
         'simpleact : optunprovable ENSURE labeledfmla PROOF proofstep'
         if __debug__: xtracer.trace("parser.p_action_ensure_proof_proofstep ENTER (simpleact)")
         p[0] = EnsuresAction(check_non_temporal(addlabel(p[3],'asrt')),p[5])
         addunprovable(p[0].args[0],p[1])
         if p[1] and not check_unprovable.get():
             p[0] = Sequence()
         p[0].lineno = get_lineno(p,2)

    def p_action_require(p):
         'simpleact : optunprovable REQUIRE labeledfmla'
         if __debug__: xtracer.trace("parser.p_action_require ENTER (simpleact)")
         p[0] = RequiresAction(check_non_temporal(addlabel(p[3],'asrt')))
         addunprovable(p[0].args[0],p[1])
         if p[1] and not check_unprovable.get():
             p[0] = Sequence()
         p[0].lineno = get_lineno(p,2)

    def p_action_require_proof_proofstep(p):
         'simpleact : optunprovable REQUIRE labeledfmla PROOF proofstep'
         if __debug__: xtracer.trace("parser.p_action_require_proof_proofstep ENTER (simpleact)")
         p[0] = RequiresAction(check_non_temporal(addlabel(p[3],'asrt')),p[5])
         addunprovable(p[0].args[0],p[1])
         if p[1] and not check_unprovable.get():
             p[0] = Sequence()
         p[0].lineno = get_lineno(p,2)

    # def p_action_ensure_optproof(p):
    #     'simpleact : ENSURE fmla optproof'
    #     p[0] = EnsuresAction(*([check_non_temporal(p[2])] + ([p[3]] if p[3] is not None else [])))
    #     p[0].lineno = get_lineno(p,1)

    # def p_action_require_optproof(p):
    #     'simpleact : REQUIRE fmla optproof'
    #     p[0] = RequiresAction(*([check_non_temporal(p[2])] + ([p[3]] if p[3] is not None else [])))
    #     p[0].lineno = get_lineno(p,1)
    

def p_action_set_lit(p):
    'simpleact : SET lit'
    if __debug__: xtracer.trace("parser.p_action_set_lit ENTER (simpleact)")
    p[0] = SetAction(p[2])
    p[0].lineno = get_lineno(p,1)

def p_action_term_assign_fmla(p):
    'simpleact : term ASSIGN fmla'
    if __debug__: xtracer.trace("parser.p_action_term_assign_fmla ENTER (simpleact)")
    p[0] = AssignAction(p[1],check_non_temporal(p[3]))
    p[0].lineno = get_lineno(p,2)

def p_termtuple_lp_term_comma_terms_rp(p):
    'termtuple : LPAREN term COMMA terms RPAREN'
    if __debug__: xtracer.trace("parser.p_termtuple_lp_term_comma_terms_rp ENTER (termtuple)")
    p[0] = Tuple(*([p[2]]+p[4]))
    p[0].lineno = get_lineno(p,1)

def p_action_termtuple_assign_fmla(p):
    'simpleact : termtuple ASSIGN callatom'
    if __debug__: xtracer.trace("parser.p_action_termtuple_assign_fmla ENTER (simpleact)")
    p[0] = CallAction(*([p[3]]+list(p[1].args)))
    p[0].lineno = get_lineno(p,2)

def p_action_term_assign_times(p):
    'simpleact : term ASSIGN TIMES'
    if __debug__: xtracer.trace("parser.p_action_term_assign_times ENTER (simpleact)")
    p[0] = HavocAction(p[1])
    p[0].lineno = get_lineno(p,2)

def p_action_term(p):
    'simpleact : term'
    if __debug__: xtracer.trace("parser.p_action_term ENTER (simpleact)")
    p[0] = CallAction(p[1])
    p[0].lineno = p[1].lineno
       
    
if iu.get_numeric_version() <= [1,4]:

    def p_action_if_fmla_lcb_action_rcb(p):
        'complexact : IF fmla sequence'
        if __debug__: xtracer.trace("parser.p_action_if_fmla_lcb_action_rcb ENTER (complexact)")
        p[0] = IfAction(check_non_temporal(p[2]),p[3])
        p[0].lineno = get_lineno(p,1)

    def p_action_if_fmla_lcb_action_rcb_else_LCB_action_RCB(p):
        'complexact : IF fmla sequence ELSE action'
        if __debug__: xtracer.trace("parser.p_action_if_fmla_lcb_action_rcb_else_LCB_action_RCB ENTER (complexact)")
        p[0] = IfAction(check_non_temporal(p[2]),p[3],p[5])
        p[0].lineno = get_lineno(p,1)

else:

    def p_somefmla_fmla(p):
        'somefmla : fmla'
        if __debug__: xtracer.trace("parser.p_somefmla_fmla ENTER (somefmla)")
        p[0] = p[1]

    def p_somefmla_fmla_assign_fmla(p):
        'somefmla : fmla ASSIGN fmla'
        if __debug__: xtracer.trace("parser.p_somefmla_fmla_assign_fmla ENTER (somefmla)")
        if not (isinstance(p[1],(App,Atom)) and len(p[1].args) == 0):
            report_error(ParseError(p[2].lineno,p[2].value,"syntax error"))
        lsyms = [p[1].prefix('loc:')]
        lsyms[0].sort = p[1].sort
        subst = dict((x.rep,y.rep) for x,y in zip([p[1]],lsyms))
        fmla = App('*>',p[3],p[1])
        fmla.lineno = get_lineno(p,2)
        fmla = subst_prefix_atoms_ast(fmla,subst,None,None)
        p[0] = Some(*(lsyms+[fmla]))
        p[0].lineno = get_lineno(p,2)

    def p_bounds_params_dot(p):
        'bounds : params DOT'
        if __debug__: xtracer.trace("parser.p_bounds_params_dot ENTER (bounds)")
        p[0] = p[1]

    def p_bounds_lparen_lparams_rparen(p):
        'bounds : LPAREN lparams RPAREN'
        if __debug__: xtracer.trace("parser.p_bounds_lparen_lparams_rparen ENTER (bounds)")
        p[0] = p[2]

    def p_somefmla_some_bounds_fmla(p):
        'somefmla : SOME bounds fmla'
        if __debug__: xtracer.trace("parser.p_somefmla_some_bounds_fmla ENTER (somefmla)")
        lsyms = [s.prefix('loc:') for s in p[2]]
        subst = dict((x.rep,y.rep) for x,y in zip(p[2],lsyms))
        fmla = subst_prefix_atoms_ast(p[3],subst,None,None)
        p[0] = Some(*(lsyms+[fmla]))
        p[0].lineno = get_lineno(p,1)
    
    def p_somefmla_some_bounds_fmla_minimizing_term(p):
        'somefmla : SOME bounds fmla MINIMIZING term'
        if __debug__: xtracer.trace("parser.p_somefmla_some_bounds_fmla_minimizing_term ENTER (somefmla)")
        lsyms = [s.prefix('loc:') for s in p[2]]
        subst = dict((x.rep,y.rep) for x,y in zip(p[2],lsyms))
        fmla = subst_prefix_atoms_ast(p[3],subst,None,None)
        index = subst_prefix_atoms_ast(p[5],subst,None,None)
        p[0] = SomeMin(*(lsyms+[fmla,index]))
        p[0].lineno = get_lineno(p,1)

    def p_somefmla_some_bounds_fmla_maximizing_term(p):
        'somefmla : SOME bounds fmla MAXIMIZING term'
        if __debug__: xtracer.trace("parser.p_somefmla_some_bounds_fmla_maximizing_term ENTER (somefmla)")
        lsyms = [s.prefix('loc:') for s in p[2]]
        subst = dict((x.rep,y.rep) for x,y in zip(p[2],lsyms))
        fmla = subst_prefix_atoms_ast(p[3],subst,None,None)
        index = subst_prefix_atoms_ast(p[5],subst,None,None)
        p[0] = SomeMax(*(lsyms+[fmla,index]))
        p[0].lineno = get_lineno(p,1)

    def fix_if_part(cond,part):
        if __debug__: xtracer.trace("parser.fix_if_part ENTER")
        if isinstance(cond,Some):
            args = cond.params()
            subst = dict((x.rep[4:],x.rep) for x in args)
            part = subst_prefix_atoms_ast(part,subst,None,None)
        return part

    def p_action_if_somefmla_lcb_action_rcb(p):
        'complexact : IF somefmla sequence'
        if __debug__: xtracer.trace("parser.p_action_if_somefmla_lcb_action_rcb ENTER (complexact)")
        p[2] = check_non_temporal(p[2])
        p[0] = IfAction(p[2],fix_if_part(p[2],p[3]))
        p[0].lineno = get_lineno(p,1)

    def p_action_if_somefmla_lcb_action_rcb_else_LCB_action_RCB(p):
        'complexact : IF somefmla sequence ELSE action'
        if __debug__: xtracer.trace("parser.p_action_if_somefmla_lcb_action_rcb_else_LCB_action_RCB ENTER (complexact)")
        p[2] = check_non_temporal(p[2])
        p[0] = IfAction(p[2],fix_if_part(p[2],p[3]),p[5])
        p[0].lineno = get_lineno(p,1)
        
    def p_invariants(p):
        'invariants : '
        if __debug__: xtracer.trace("parser.p_invariants ENTER (invariants)")
        p[0] = []

    def p_invariant_invariant_fmla(p):
        'invariants : invariants INVARIANT labeledfmla'
        if __debug__: xtracer.trace("parser.p_invariant_invariant_fmla ENTER (invariants)")
        p[0] = p[1]
        inv = check_non_temporal(addlabel(p[3],'asrt'))
        a = AssertAction(inv)
        a.lineno = get_lineno(p,2)
        p[0].append(a)

    def p_invariant_invariant_fmla_proof(p):
        'invariants : invariants INVARIANT labeledfmla PROOF proofstep'
        if __debug__: xtracer.trace("parser.p_invariant_invariant_fmla_proof ENTER (invariants)")
        p[0] = p[1]
        inv = check_non_temporal(addlabel(p[3],'asrt'))
        a = AssertAction(inv,p[5])
        a.lineno = get_lineno(p,2)
        p[0].append(a)

    def p_decreases_decreases_fmla(p):
        'decreases : DECREASES fmla'
        if __debug__: xtracer.trace("parser.p_decreases_decreases_fmla ENTER (decreases)")
        rank = Ranking(check_non_temporal(p[2]))
        rank.lineno = get_lineno(p,1)
        p[0] = [rank]

    def p_decreases(p):
        'decreases : '
        if __debug__: xtracer.trace("parser.p_decreases ENTER (decreases)")
        p[0] = []

    def p_action_while_somefmla_invariants_decreases_lcb_action_rcb(p):
        'complexact : WHILE somefmla invariants decreases sequence'
        if __debug__: xtracer.trace("parser.p_action_while_somefmla_invariants_decreases_lcb_action_rcb ENTER (complexact)")
        p[0] = WhileAction(*([check_non_temporal(p[2]), fix_if_part(p[2],p[5])] + p[3] + p[4]))
        p[0].lineno = get_lineno(p,1)

    def methcall(lhs,rhs):
        if __debug__: xtracer.trace("parser.methcall ENTER")
        if (isinstance(lhs,App) or isinstance(lhs,Atom)) and len(lhs.args) == 0:
            return compose_atoms(lhs,rhs)
        return MethodCall(lhs,rhs)

    def p_action_for_tterm_comma_tterm_in_expr_invariants_decreases_lcb_action_rcb(p):
        'complexact : FOR tterm COMMA tterm IN fmla invariants decreases sequence'
        if __debug__: xtracer.trace("parser.p_action_for_tterm_comma_tterm_in_expr_invariants_decreases_lcb_action_rcb ENTER (complexact)")
        itr,val,fmla,invars,decrs,seq = p[2],p[4],check_non_temporal(p[6]),p[7],p[8],p[9]
        iend = itr.rename('loc:end')
        ln = get_lineno(p,1)
        didx = VarAction(itr,methcall(fmla,App('begin').sln(ln)).sln(ln)).sln(ln)
        dend = VarAction(iend,methcall(fmla,App('end').sln(ln)).sln(ln)).sln(ln)
        dval = VarAction(val,methcall(fmla,App('value',itr).sln(ln)).sln(ln)).sln(ln)
        incr = AssignAction(itr,methcall(itr,App('next').sln(ln)).sln(ln)).sln(ln)
        body = Sequence(*lower_var_stmts([dval,seq,incr])).sln(ln)
        loop = WhileAction(*([App('<',itr,iend).sln(ln),body] + invars + decrs)).sln(ln)
        p[0] = Sequence(*lower_var_stmts([didx,dend,loop])).sln(ln)

def p_action_if_times_lcb_action_rcb_else_LCB_action_RCB(p):
    'complexact : IF TIMES sequence ELSE action'
    if __debug__: xtracer.trace("parser.p_action_if_times_lcb_action_rcb_else_LCB_action_RCB ENTER (complexact)")
    p[0] = ChoiceAction(p[3],p[5])
    p[0].lineno = get_lineno(p,1)


if iu.get_numeric_version() <= [1,2]:
    def p_action_field_assign_term(p):
        'simpleact : term DOT SYMBOL ASSIGN term'
        if __debug__: xtracer.trace("parser.p_action_field_assign_term ENTER (simpleact)")
        p[0] = AssignFieldAction(p[1],p[3],p[5])
        p[0].lineno = get_lineno(p,4)


    def p_action_field_assign_null(p):
        'simpleact : term DOT SYMBOL ASSIGN NULL'
        if __debug__: xtracer.trace("parser.p_action_field_assign_null ENTER (simpleact)")
        p[0] = NullFieldAction(p[1],p[3])
        p[0].lineno = get_lineno(p,4)

    def p_action_field_assign_field(p):
        'simpleact : term DOT SYMBOL ASSIGN term DOT SYMBOL'
        if __debug__: xtracer.trace("parser.p_action_field_assign_field ENTER (simpleact)")
        p[0] = CopyFieldAction(p[1],p[3],p[5],p[7])
        p[0].lineno = get_lineno(p,4)

    def p_action_field_assign_false(p):
        'simpleact : term DOT SYMBOL ASSIGN FALSE'
        if __debug__: xtracer.trace("parser.p_action_field_assign_false ENTER (simpleact)")
        p[0] = NullFieldAction(p[1],p[3])
        p[0].lineno = get_lineno(p,4)

def p_action_instantiate_atom(p):
    'simpleact : INSTANTIATE callatom'
    if __debug__: xtracer.trace("parser.p_action_instantiate_atom ENTER (simpleact)")
#    p[0] = InstantiateAction(app_to_atom(p[2]))
    p[0] = InstantiateAction(p[2])
    p[0].lineno = get_lineno(p,1)

def p_callatom_atom(p):
    'callatom : atom'
    if __debug__: xtracer.trace("parser.p_callatom_atom ENTER (callatom)")
    p[0] = p[1]

if not (iu.get_numeric_version() <= [1,5]):
    def p_callatom_this(p):
        'callatom : THIS'
        if __debug__: xtracer.trace("parser.p_callatom_this ENTER (callatom)")
        p[0] = Atom(This())
        p[0].lineno = get_lineno(p,1)
        
    def p_callatom_method(p):
        'callatom : METHOD'
        if __debug__: xtracer.trace("parser.p_callatom_method ENTER (callatom)")
        p[0] = Atom('method')
        p[0].lineno = get_lineno(p,1)
        

if iu.get_numeric_version() <= [1,2]:

    def p_callatom_callatom_colon_callatom(p):
        'callatom : callatom COLON callatom'
        if __debug__: xtracer.trace("parser.p_callatom_callatom_colon_callatom ENTER (callatom)")
        p[0] = compose_atoms(p[1],p[3])
        p[0].lineno = get_lineno(p,1)

else:

    def p_callatom_callatom_dot_callatom(p):
        'callatom : callatom DOT callatom'
        if __debug__: xtracer.trace("parser.p_callatom_callatom_dot_callatom ENTER (callatom)")
        p[0] = compose_atoms(p[1],p[3])
        p[0].lineno = get_lineno(p,1)


def p_callatoms_callatom(p):
    'callatoms : callatom'
    if __debug__: xtracer.trace("parser.p_callatoms_callatom ENTER (callatoms)")
    p[0] = [p[1]]

def p_callatoms_callatoms_callatom(p):
    'callatoms : callatoms COMMA callatom'
    if __debug__: xtracer.trace("parser.p_callatoms_callatoms_callatom ENTER (callatoms)")
    p[0] = p[1]
    p[0].append(p[3])

def p_action_call_optreturns_callatom(p):
    'simpleact : CALL optactualreturns callatom'
    if __debug__: xtracer.trace("parser.p_action_call_optreturns_callatom ENTER (simpleact)")
    p[0] = CallAction(*([p[3]] + p[2]))
    p[0].lineno = get_lineno(p,1)

def p_action_call_callatom(p):
    'simpleact : CALL callatom'
    if __debug__: xtracer.trace("parser.p_action_call_callatom ENTER (simpleact)")
    p[0] = CallAction(p[2])
    p[0].lineno = get_lineno(p,1)

# def p_action_call_callatom_assign_callatom(p):
#     'simpleact : CALL callatom ASSIGN callatom'
#     p[0] = CallAction(p[4],p[2])
#     p[0].lineno = get_lineno(p,1)


def p_lparam_variable_colon_symbol(p):
    'lparam : SYMBOL COLON atype'
    if __debug__: xtracer.trace("parser.p_lparam_variable_colon_symbol ENTER (lparam)")
    p[0] = App(p[1])
    p[0].lineno = get_lineno(p,1)
    p[0].sort = p[3]

if not (iu.get_numeric_version() <= [1,6]):

    def p_lparam_caret_variable_colon_symbol(p):
        'lparam : CARET SYMBOL COLON atype'
        if __debug__: xtracer.trace("parser.p_lparam_caret_variable_colon_symbol ENTER (lparam)")
        p[0] = KeyArg(p[2])
        p[0].lineno = get_lineno(p,2)
        p[0].sort = p[4]

def p_lparams_lparam(p):
    'lparams : lparam'
    if __debug__: xtracer.trace("parser.p_lparams_lparam ENTER (lparams)")
    p[0] = [p[1]]

def p_lparams_lparams_comma_lparam(p):
    'lparams : lparams COMMA lparam'
    if __debug__: xtracer.trace("parser.p_lparams_lparams_comma_lparam ENTER (lparams)")
    p[0] = p[1]
    p[0].append(p[3])


def p_action_local_params_lcb_action_rcb(p):
    'complexact : LOCAL lparams sequence'
    if __debug__: xtracer.trace("parser.p_action_local_params_lcb_action_rcb ENTER (complexact)")
    # we rename the locals to avoid name capture
    lsyms = [s.prefix('loc:') for s in p[2]]
    subst = dict((x.rep,y.rep) for x,y in zip(p[2],lsyms))
    action = subst_prefix_atoms_ast(p[3],subst,None,None)
    p[0] = LocalAction(*(lsyms+[action]),caller="parser.local_action_rule")
    p[0].lineno = get_lineno(p,1)

if not (iu.get_numeric_version() <= [1,5]):
    def p_opttypedsym_symbol(p):
        'opttypedsym : SYMBOL'
        if __debug__: xtracer.trace("parser.p_opttypedsym_symbol ENTER (opttypedsym)")
        p[0] = App(p[1])
        p[0].lineno = get_lineno(p,1)
        p[0].sort = 'S'

    def p_opttypedsym_symbol_colon_atype(p):
        'opttypedsym : SYMBOL COLON atype'
        if __debug__: xtracer.trace("parser.p_opttypedsym_symbol_colon_atype ENTER (opttypedsym)")
        p[0] = App(p[1])
        p[0].lineno = get_lineno(p,1)
        p[0].sort = p[3]

    def p_optinit(p):
        'optinit : '
        if __debug__: xtracer.trace("parser.p_optinit ENTER (optinit)")
        p[0] = None

    def p_optinit_assign_fmla(p):
        'optinit : ASSIGN fmla'
        if __debug__: xtracer.trace("parser.p_optinit_assign_fmla ENTER (optinit)")
        p[0] = check_non_temporal(p[2])

    def p_action_var_opttypedsym_assign_fmla(p):
        'simpleact : VAR tterm optinit'
        if __debug__: xtracer.trace("parser.p_action_var_opttypedsym_assign_fmla ENTER (simpleact)")
        p[0] = VarAction(p[2],p[3]) if p[3] is not None else VarAction(p[2])
        p[0].lineno = get_lineno(p,1)

if not (iu.get_numeric_version() <= [1,6]):
    def p_action_thunk_symbol_optargs_colon_atype_assign_sequence(p):
        'complexact : THUNK LABEL SYMBOL optargs COLON atype ASSIGN sequence'
        if __debug__: xtracer.trace("parser.p_action_thunk_symbol_optargs_colon_atype_assign_sequence ENTER (complexact)")
        action = Atom(p[3],p[4])
        action.lineno = get_lineno(p,3)
        p[0] = ThunkAction(Atom(p[2][1:-1],[]),action,Atom(p[6]),p[8])
        p[0].lineno = get_lineno(p,1)



if not (iu.get_numeric_version() <= [1,6]):
    def p_debugarg_symbol_equal_fmla(p):
        'debugarg : SYMBOL EQ fmla'
        if __debug__: xtracer.trace("parser.p_debugarg_symbol_equal_fmla ENTER (debugarg)")
        lhs = App(p[1])
        lhs.lineno = get_lineno(p,1)
        p[0] = DebugItem(lhs,p[3])
        p[0].lineno = get_lineno(p,2)
    
    def p_debugargs(p):
        'debugargs : debugarg'
        if __debug__: xtracer.trace("parser.p_debugargs ENTER (debugargs)")
        p[0] = [p[1]]

    def p_debugargs_debugarg_symbol_equal_fmla(p):
        'debugargs : debugargs COMMA debugarg'
        if __debug__: xtracer.trace("parser.p_debugargs_debugarg_symbol_equal_fmla ENTER (debugargs)")
        p[0] = p[1]
        p[0].append(p[3])

    def p_optdebugargs(p):
        'optdebugargs : '
        if __debug__: xtracer.trace("parser.p_optdebugargs ENTER (optdebugargs)")
        p[0] = []
        
    def p_optdebugargs_with_debugargs(p):
        'optdebugargs : WITH debugargs'
        if __debug__: xtracer.trace("parser.p_optdebugargs_with_debugargs ENTER (optdebugargs)")
        p[0] = p[2]
        
    def p_simpleact_debug_symbol_optdebugargs(p):
        'simpleact : DEBUG SYMBOL optdebugargs'
        if __debug__: xtracer.trace("parser.p_simpleact_debug_symbol_optdebugargs ENTER (simpleact)")
        action = Atom(p[2],[])
        action.lineno = get_lineno(p,2)
        if not p[2].startswith('"'):
            report_error(IvyError(action,"expected string constant after 'debug'"))
        p[0] = DebugAction(action,*p[3])
        p[0].lineno = get_lineno(p,1)


def p_eqn_SYMBOL_EQ_SYMBOL(p):
    'eqn : SYMBOL EQ SYMBOL'
    if __debug__: xtracer.trace("parser.p_eqn_SYMBOL_EQ_SYMBOL ENTER (eqn)")
    p[0] = Equals(App(p[1]),App(p[3]))

def p_eqns_eqn(p):
    'eqns : eqn'
    if __debug__: xtracer.trace("parser.p_eqns_eqn ENTER (eqns)")
    p[0] = [p[1]]

def p_eqns_eqns_comma_eqn(p):
    'eqns : eqns COMMA eqn'
    if __debug__: xtracer.trace("parser.p_eqns_eqns_comma_eqn ENTER (eqns)")
    p[0] = p[1]
    p[0].append(p[3])

def p_action_let_eqns_lcb_action_rcb(p):
    'complexact : LET eqns sequence'
    if __debug__: xtracer.trace("parser.p_action_let_eqns_lcb_action_rcb ENTER (complexact)")
    p[0] = LetAction(*(p[2]+[p[3]]))

def p_symbols(p):
    'symbols : SYMBOL'
    if __debug__: xtracer.trace("parser.p_symbols ENTER (symbols)")
    p[0] = [p[1]]

def p_symbols_symbols_symbol(p):
    'symbols : symbols COMMA SYMBOL'
    if __debug__: xtracer.trace("parser.p_symbols_symbols_symbol ENTER (symbols)")
    p[0] = p[1]
    p[0].append(p[3])

def p_cdefns_cdefn(p):
    'cdefns : cdefn'
    if __debug__: xtracer.trace("parser.p_cdefns_cdefn ENTER (cdefns)")
    p[0] = [p[1]]

def p_cdefns_cdefns_comma_cdefn(p):
    'cdefns : cdefns COMMA cdefn'
    if __debug__: xtracer.trace("parser.p_cdefns_cdefns_comma_cdefn ENTER (cdefns)")
    p[0] = p[1]
    p[0].append(p[3])

def p_cdefn_atom_expr(p):
    'cdefn : atom EQ expr'
    if __debug__: xtracer.trace("parser.p_cdefn_atom_expr ENTER (cdefn)")
    p[0] = Definition(app_to_atom(p[1]),p[3])
    p[0].lineno = get_lineno(p,2)

def p_defns_defn(p):
    'defns : defn'
    if __debug__: xtracer.trace("parser.p_defns_defn ENTER (defns)")
    p[0] = [p[1]]

def p_defns_defns_comma_defn(p):
    'defns : defns COMMA defn'
    if __debug__: xtracer.trace("parser.p_defns_defns_comma_defn ENTER (defns)")
    p[0] = p[1]
    p[0].append(p[3])

def p_dotsym_symbol(p):
    'dotsym : SYMBOL'
    if __debug__: xtracer.trace("parser.p_dotsym_symbol ENTER (dotsym) val=%s" % p[1])
    p[0] = Atom(p[1],[])
    p[0].lineno = get_lineno(p,1)

def p_dotsym_dotsym_dot_symbol(p):
    'dotsym : dotsym DOT SYMBOL'
    if __debug__: xtracer.trace("parser.p_dotsym_dotsym_dot_symbol ENTER (dotsym)")
    p[0] = Atom(p[1].relname + iu.ivy_compose_character + p[3],[])
    p[0].lineno = p[1].lineno

def p_defnlhs_symbol(p):
    'defnlhs : dotsym'
    if __debug__: xtracer.trace("parser.p_defnlhs_symbol ENTER (defnlhs)")
    p[0] = p[1]
    
def p_defnlhs_symbol_lparen_defargs_rparen(p):
    'defnlhs : dotsym LPAREN defargs RPAREN'
    if __debug__: xtracer.trace("parser.p_defnlhs_symbol_lparen_defargs_rparen ENTER (defnlhs)")
    p[0] = p[1]
    p[0].args.extend(p[3])
    
def p_defargs_defarg(p):
    'defargs : defarg'
    if __debug__: xtracer.trace("parser.p_defargs_defarg ENTER (defargs)")
    p[0] = [p[1]]

def p_defargs_defargs_comma_defarg(p):
    'defargs : defargs COMMA defarg'
    if __debug__: xtracer.trace("parser.p_defargs_defargs_comma_defarg ENTER (defargs)")
    p[0] = p[1]
    p[0].append(p[3])

def p_defarg_lparam(p):
    'defarg : lparam'
    if __debug__: xtracer.trace("parser.p_defarg_lparam ENTER (defarg)")
    p[0] = p[1]

def p_defarg_var(p):
    'defarg : var'
    if __debug__: xtracer.trace("parser.p_defarg_var ENTER (defarg)")
    p[0] = p[1]

def p_defnlhs_lp_term_relop_term_rp(p):
    'defnlhs : LPAREN defarg relop defarg RPAREN'
    if __debug__: xtracer.trace("parser.p_defnlhs_lp_term_relop_term_rp ENTER (defnlhs)")
    p[0] = Atom(p[3],[p[2],p[4]])
    p[0].lineno = get_lineno(p,3)

def p_defnlhs_lp_term_infix_term_rp(p):
    'defnlhs : LPAREN defarg infix defarg RPAREN'
    if __debug__: xtracer.trace("parser.p_defnlhs_lp_term_infix_term_rp ENTER (defnlhs)")
    p[0] = App(p[3],[p[2],p[4]])
    p[0].lineno = get_lineno(p,3)

def p_typeddefn_defnlhs(p):
    'typeddefn : defnlhs'
    if __debug__: xtracer.trace("parser.p_typeddefn_defnlhs ENTER (typeddefn)")
    p[0] = p[1]

def p_typeddefn_defnlhs_colon_atype(p):
    'typeddefn : defnlhs COLON atype'
    if __debug__: xtracer.trace("parser.p_typeddefn_defnlhs_colon_atype ENTER (typeddefn)")
    p[0] = p[1]
    p[0].sort = p[3]

def p_defnrhs_fmla(p):
    'defnrhs : fmla'
    if __debug__: xtracer.trace("parser.p_defnrhs_fmla ENTER (defnrhs)")
    p[0] = check_non_temporal(p[1])

def p_defnrhs_somevarfmla(p):
    'defnrhs : somevarfmla'
    if __debug__: xtracer.trace("parser.p_defnrhs_somevarfmla ENTER (defnrhs)")
    p[0] = check_non_temporal(p[1])

def p_defnrhs_nativequote(p):
    'defnrhs :  NATIVEQUOTE'
    if __debug__: xtracer.trace("parser.p_defnrhs_nativequote ENTER (defnrhs)")
    text,bqs = parse_nativequote(p,1)
    p[0] = NativeExpr(*([text] + bqs))
    p[0].lineno = get_lineno(p,1)

def p_defn_atom_fmla(p):
    'defn : typeddefn EQ defnrhs'
    if __debug__: xtracer.trace("parser.p_defn_atom_fmla ENTER (defn)")
    p[0] = Definition(app_to_atom(p[1]),p[3])
    p[0].lineno = get_lineno(p,2)

# def p_defn_defnlhs_eq_(p):
#     'defn : typeddefn EQ NATIVEQUOTE'
#     text,bqs = parse_nativequote(p,3)
#     p[0] = Definition(app_to_atom(p[1]),NativeExpr(*([text] + bqs)))
#     p[0].lineno = get_lineno(p,2)

def p_optin(p):
    'optin : '
    if __debug__: xtracer.trace("parser.p_optin ENTER (optin)")
    p[0] = []

def p_optin_in_fmla(p):
    'optin : IN fmla'
    if __debug__: xtracer.trace("parser.p_optin_in_fmla ENTER (optin)")
    p[0] = [p[2]]

def p_optelse(p):
    'optelse : '
    if __debug__: xtracer.trace("parser.p_optelse ENTER (optelse)")
    p[0] = []

def p_optelse_else_fmla(p):
    'optelse : ELSE fmla'
    if __debug__: xtracer.trace("parser.p_optelse_else_fmla ENTER (optelse)")
    p[0] = [p[2]]

def p_somevarfmla_some_simplevar_dot_fmla(p):
    'somevarfmla : SOME simplevar DOT fmla optin optelse'
    if __debug__: xtracer.trace("parser.p_somevarfmla_some_simplevar_dot_fmla ENTER (somevarfmla)")
    p[0] = SomeExpr(*([p[2],p[4]]+p[5]+p[6]))
    p[0].lineno = get_lineno(p,1)

def p_expr_fmla(p):
    'expr : LCB fmla RCB'
    if __debug__: xtracer.trace("parser.p_expr_fmla ENTER (expr)")
    p[0] = NamedSpace(Literal(1,check_non_temporal(p[2])))

def p_exprterm_aterm(p):
    'exprterm : aterm'
    if __debug__: xtracer.trace("parser.p_exprterm_aterm ENTER (exprterm)")
    p[0] = p[1]

def p_exprterm_var(p):
    'exprterm : var'
    if __debug__: xtracer.trace("parser.p_exprterm_var ENTER (exprterm)")
    p[0] = p[1]

def p_expr_exprterm(p):
    'expr : exprterm'
    if __debug__: xtracer.trace("parser.p_expr_exprterm ENTER (expr)")
    p[0] = NamedSpace(Literal(1,app_to_atom(p[1])))

def p_expr_exprterm_relop_exprterm(p):
    'expr : exprterm relop exprterm'
    if __debug__: xtracer.trace("parser.p_expr_exprterm_relop_exprterm ENTER (expr)")
    p[0] = NamedSpace(Literal(1,Atom(p[2],[p[1],p[3]])))
    p[0].lineno = get_lineno(p,2)

def p_expr_exprterm_tildaeq_exprterm(p):
    'expr : exprterm TILDAEQ exprterm'
    if __debug__: xtracer.trace("parser.p_expr_exprterm_tildaeq_exprterm ENTER (expr)")
    p[0] = NamedSpace(Literal(0,Atom('=',[p[1],p[3]])))
    p[0].lineno = get_lineno(p,2)

def p_expr_tilda_atom(p):
    'expr : TILDA expr'
    if __debug__: xtracer.trace("parser.p_expr_tilda_atom ENTER (expr)")
    p[0] = NamedSpace(~p[2].lit)

# def p_expr_lit(p):
#     'expr : lit'
#     p[0] = NamedSpace(p[1])

def p_expr_lparen_expr_rparen(p):
    'expr : LPAREN expr RPAREN'
    if __debug__: xtracer.trace("parser.p_expr_lparen_expr_rparen ENTER (expr)")
    p[0] = p[2]

def p_expr_prod(p):
    'expr : prod'
    if __debug__: xtracer.trace("parser.p_expr_prod ENTER (expr)")
    p[0] = ProductSpace(p[1])
    
def p_expr_sum(p):
    'expr : sum'
    if __debug__: xtracer.trace("parser.p_expr_sum ENTER (expr)")
    p[0] = SumSpace(p[1])
    
def p_prod_expr_expr(p):
    'prod : expr TIMES expr'
    if __debug__: xtracer.trace("parser.p_prod_expr_expr ENTER (prod)")
    p[0] = [p[1],p[3]]

def p_prod_prod_expr(p):
    'prod : prod TIMES expr'
    if __debug__: xtracer.trace("parser.p_prod_prod_expr ENTER (prod)")
    p[0] = p[1]
    p[0].append(p[3]) # is this side effect OK?

def p_sum_expr_expr(p):
    'sum : expr PLUS expr'
    if __debug__: xtracer.trace("parser.p_sum_expr_expr ENTER (sum)")
    p[0] = [p[1],p[3]]

def p_sum_sum_expr(p):
    'sum : sum PLUS expr'
    if __debug__: xtracer.trace("parser.p_sum_sum_expr ENTER (sum)")
    p[0] = p[1]
    p[0].append(p[3]) # is this side effect OK?

def p_state_expr_true(p):
    'state_expr : TRUE'
    if __debug__: xtracer.trace("parser.p_state_expr_true ENTER (state_expr)")
    p[0] = And()

def p_state_expr_false(p):
    'state_expr : FALSE'
    if __debug__: xtracer.trace("parser.p_state_expr_false ENTER (state_expr)")
    p[0] = Or()

def p_state_expr_symbol(p):
    'state_expr : SYMBOL'
    if __debug__: xtracer.trace("parser.p_state_expr_symbol ENTER (state_expr)")
    p[0] = Atom(p[1],[])

def p_state_expr_symbol_lparen_state_expr_rparen(p):
    'state_expr : SYMBOL LPAREN state_expr RPAREN'
    if __debug__: xtracer.trace("parser.p_state_expr_symbol_lparen_state_expr_rparen ENTER (state_expr)")
    p[0] = Atom(p[1],[p[3]])

def p_state_expr_state_expr_or_state_expr(p):
    'state_expr : state_expr OR state_expr'
    if __debug__: xtracer.trace("parser.p_state_expr_state_expr_or_state_expr ENTER (state_expr)")
    if isinstance(p[1],Or):
        p[0] = p[1]
        p[0].args.append(p[3])
    else:
        p[0] = Or(p[1],p[3])

def p_state_expr_lcb_requires_modifies_ensures_rcb(p):
    'state_expr : LCB requires modifies ensures RCB'
    if __debug__: xtracer.trace("parser.p_state_expr_lcb_requires_modifies_ensures_rcb ENTER (state_expr)")
    p[0] = RME(p[2],p[3],p[4])
    
def p_state_expr_entry(p):
    'state_expr : ENTRY'
    if __debug__: xtracer.trace("parser.p_state_expr_entry ENTER (state_expr)")
    p[0] = RME(And(),[],And())

from .ivy_logic_parser import *

def p_error(token):
    if __debug__: xtracer.trace("parser.p_error ENTER")
    if token is not None:
        report_error(ParseError(token.lineno,token.value,"syntax error"))
    else:
        report_error(ParseError(None,None,'unexpected end of input'));
    # TEMORARY: parser goes into into infinite loop on recovery from parse errors
    # Stop here on any parse error to prevent this
    raise iu.ErrorList(error_list)

# Build the parsers
import os
tabdir = os.path.dirname(os.path.abspath(__file__))
parser = yacc.yacc(start='top',tabmodule='ivy_parsetab',errorlog=yacc.NullLogger(),outputdir=tabdir,debug=None)
#parser = yacc.yacc(start='top',tabmodule='ivy_parsetab',outputdir=tabdir,debug=None)
#parser = yacc.yacc(start='top',tabmodule='ivy_parsetab')
# formula_parser = yacc.yacc(start = 'fmla', tabmodule='ivy_formulatab')

class TypeNames(object):
    def __init__(self):
        if __debug__: xtracer.trace("parser.__init__ ENTER")
        self.namelist = []
        self.nameset = set()
    def add(self,tname):
        if __debug__: xtracer.trace("parser.add ENTER")
        if tname not in self.nameset:
            pref,refparms = iu.extract_parameters_name(tname)
            for rp in refparms:
                self.add(rp)
            self.namelist.append(tname)
            self.nameset.add(tname)

def expand_autoinstances(ivy):
    if __debug__: xtracer.trace("parser.expand_autoinstances ENTER")
    autos = defaultdict(list)
    trefs = set()
    decls = ivy.decls
    if __debug__: xtracer.trace("parser.expand_auto ENTER decls=%d" % len(decls))
    ivy.decls = []
    for decl in decls:
        if isinstance(decl,AutoInstanceDecl):
            for inst in decl.args:
                if len(inst.args) == 2:
                    pref,parms = iu.extract_parameters_name(inst.args[0].rep)
                    key = (pref,len(parms))
                    autos[key].append(inst)
        else:
            drefs = TypeNames()
            decl.get_type_names(drefs)
            for tname in drefs.namelist:
                if tname not in trefs:
                    trefs.add(tname)
                    pref,refparms = iu.extract_parameters_name(tname)
                    key = (pref,len(refparms))
                    for inst in autos[key]:
                        pref,parms = iu.extract_parameters_name(inst.args[0].rep)
                        lhs = Atom(tname,[]) 
                        subst = dict(list(zip(parms,refparms)))
                        rhs = inst.args[1].clone([Atom(subst.get(a.rep,a.rep),[]) for a in inst.args[1].args])
                        newinst = Instantiation(lhs,rhs)
                        if hasattr(decl,"lineno"):
                            newinst.lineno = decl.lineno
                        do_insts(ivy,[newinst])
            ivy.decls.append(decl)
    result = ivy.decls
    if __debug__: xtracer.trace("parser.expand_auto EXIT decls=%d" % len(result))

def parse(s,nested=False):
    if __debug__: xtracer.trace("parser.Parse ENTER")
    global error_list
    global stack
    if not nested:
        error_list = []
        stack = []
    vernum = iu.get_numeric_version()
    with LexerVersion(vernum):
        # shallow copy the parser and lexer to try for re-entrance (!!!)
        res = copy.copy(parser).parse(s,lexer=copy.copy(lexer),tracking=True)
    if not nested:
        expand_autoinstances(res)
    if error_list:
        raise iu.ErrorList(error_list)
    if __debug__: xtracer.trace("parser.Parse EXIT decls=%d" % len(res.decls))
    return res
    
def to_formula(s):
    if __debug__: xtracer.trace("parser.to_formula ENTER")
    return formula_parser.parse(s,tracking=True)

if __name__ == '__main__':
#    while True:
#       try:
#       s = raw_input('input > ')
#       except EOFError:
#           break
#       if not s: continue
       s = open('test.ivy','r').read()
       try:
           result = parse(s)
           print(result)
           print(result.defined)
       except iu.ErrorList as e:
           print(repr(e))
#       print "enum: %s" % result.enumerate(dict(),lambda x:True)

def clauses_to_concept(name,clauses):
    if __debug__: xtracer.trace("parser.clauses_to_concept ENTER")
    vars =  used_variables_clauses(clauses)
    ps = [ProductSpace([NamedSpace(~lit) for lit in clause]) for clause in clauses]
    ss = ps[0] if len(ps) == 1 else SumSpace(ps)
    return (Atom(name,vars),ss)

    
