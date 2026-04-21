"""
Monkey-patches canon() methods onto all ivy_ast.py and ivy_actions.py classes.
Call install() once at startup to add all canon() methods.

Produces output compatible with Go goivy ast.Canon() methods (flattened format).
"""

from . import ivy_ast as ast
from . import ivy_actions as act
from .canon import (node_canon, slice_canon, string_canon, string_slice_canon,
                    bool_canon, lineno_fields, decl_fields, canon_blake3)


def install():
    """Install canon() and blake3() methods on all AST classes."""

    # --- blake3() on AST base (available to all subclasses) ---
    def _ast_blake3(self):
        """Return blake3 hash of this node's canonical s-expression.
        Matches Go ivyutils.Canonical.Blake3()."""
        return canon_blake3(self.canon())
    ast.AST.blake3 = _ast_blake3

    # --- Core AST base ---

    def _ast_canon(self):
        return '({}{})'.format(
            type(self).__name__[0].lower() + type(self).__name__[1:],
            lineno_fields(self))
    ast.AST.canon = _ast_canon

    def _noneast_canon(self):
        return '(noneAST{})'.format(lineno_fields(self))
    ast.NoneAST.canon = _noneast_canon

    # --- Symbol ---
    def _symbol_canon(self):
        return '(symbol{} rep:{} sort:{})'.format(
            lineno_fields(self), string_canon(self.rep), node_canon(self.sort))
    ast.Symbol.canon = _symbol_canon

    # --- Atom ---
    def _atom_canon(self):
        asort = getattr(self, 'sort', None)
        return '(atom{} rep:{} terms:{} aSort:{})'.format(
            lineno_fields(self), string_canon(self.rep),
            slice_canon(list(self.args)), node_canon(asort))
    ast.Atom.canon = _atom_canon

    # --- App ---
    def _app_canon(self):
        asort = getattr(self, 'sort', None)
        return '(app{} rep:{} terms:{} aSort:{})'.format(
            lineno_fields(self), node_canon(self.rep),
            slice_canon(list(self.args)), node_canon(asort))
    ast.App.canon = _app_canon

    # --- Variable ---
    def _variable_canon(self):
        return '(variable{} rep:{} vSort:{})'.format(
            lineno_fields(self), string_canon(self.rep), node_canon(self.sort))
    ast.Variable.canon = _variable_canon

    # --- This ---
    def _this_canon(self):
        return '(this{})'.format(lineno_fields(self))
    ast.This.canon = _this_canon

    # --- Old ---
    def _old_canon(self):
        return '(old{} term:{})'.format(
            lineno_fields(self), node_canon(self.args[0]))
    ast.Old.canon = _old_canon

    # --- MethodCall ---
    def _methodcall_canon(self):
        return '(methodCall{} obj:{} method:{})'.format(
            lineno_fields(self), node_canon(self.args[0]), node_canon(self.args[1]))
    ast.MethodCall.canon = _methodcall_canon

    # --- Literal ---
    def _literal_canon(self):
        return '(literal{} polarity:{} atom:{})'.format(
            lineno_fields(self), self.polarity, node_canon(self.atom))
    ast.Literal.canon = _literal_canon

    # --- Dot (Python ivy_ast.Dot, not the App-with-dot-rep convention) ---
    def _dot_canon(self):
        return '(dot{} left:{} right:{})'.format(
            lineno_fields(self),
            node_canon(self.args[0]),
            node_canon(self.args[1]))
    ast.Dot.canon = _dot_canon

    # --- Bracket (subscript a[b]) ---
    def _bracket_canon(self):
        return '(bracket{} left:{} right:{})'.format(
            lineno_fields(self),
            node_canon(self.args[0]),
            node_canon(self.args[1]))
    ast.Bracket.canon = _bracket_canon

    # --- KeyArg (App subclass with ^ prefix) ---
    # Go emits "(keyArg app:{})" with NO lineno_fields; Python mirrors that.
    def _keyarg_canon(self):
        # Use the same formatter as App.canon but wrap in keyArg.
        return '(keyArg app:{})'.format(_app_canon(self))
    ast.KeyArg.canon = _keyarg_canon

    # --- DebugItem (name=value pair) ---
    def _debugitem_canon(self):
        return '(debugItem{} name:{} value:{})'.format(
            lineno_fields(self),
            node_canon(self.args[0]),
            node_canon(self.args[1]))
    ast.DebugItem.canon = _debugitem_canon

    # --- Tuple ---
    def _tuple_canon(self):
        return '(tuple{} elems:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.Tuple.canon = _tuple_canon

    # --- Formula types ---

    def _and_canon(self):
        return '(and{} terms:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.And.canon = _and_canon

    def _or_canon(self):
        return '(or{} terms:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.Or.canon = _or_canon

    def _not_canon(self):
        return '(not{} body:{})'.format(
            lineno_fields(self), node_canon(self.args[0]))
    ast.Not.canon = _not_canon

    def _implies_canon(self):
        return '(implies{} t1:{} t2:{})'.format(
            lineno_fields(self), node_canon(self.args[0]), node_canon(self.args[1]))
    ast.Implies.canon = _implies_canon

    def _iff_canon(self):
        return '(iff{} t1:{} t2:{})'.format(
            lineno_fields(self), node_canon(self.args[0]), node_canon(self.args[1]))
    ast.Iff.canon = _iff_canon

    def _ite_canon(self):
        return '(ite{} cond:{} then:{} else:{})'.format(
            lineno_fields(self), node_canon(self.args[0]),
            node_canon(self.args[1]), node_canon(self.args[2]))
    ast.Ite.canon = _ite_canon

    def _globally_canon(self):
        return '(globally{} body:{})'.format(
            lineno_fields(self), node_canon(self.args[0]))
    ast.Globally.canon = _globally_canon

    def _eventually_canon(self):
        return '(eventually{} body:{})'.format(
            lineno_fields(self), node_canon(self.args[0]))
    ast.Eventually.canon = _eventually_canon

    def _whenoperator_canon(self):
        return '(whenOperator{} name:{} t1:{} t2:{})'.format(
            lineno_fields(self), string_canon(self.name),
            node_canon(self.args[0]), node_canon(self.args[1]))
    ast.WhenOperator.canon = _whenoperator_canon

    def _let_canon(self):
        # Let: args = [def1, def2, ..., body]
        defs = list(self.args[:-1])
        body = self.args[-1]
        return '(let{} defs:{} body:{})'.format(
            lineno_fields(self), slice_canon(defs), node_canon(body))
    ast.Let.canon = _let_canon

    def _definition_canon(self):
        return '(definition{} lhs:{} rhs:{})'.format(
            lineno_fields(self), node_canon(self.args[0]), node_canon(self.args[1]))
    ast.Definition.canon = _definition_canon

    def _definitionschema_canon(self):
        return '(definitionSchema definition:{})'.format(
            _definition_canon(self))
    ast.DefinitionSchema.canon = _definitionschema_canon

    def _forall_canon(self):
        return '(forall{} bounds:{} body:{})'.format(
            lineno_fields(self), slice_canon(self.bounds),
            node_canon(self.args[0]))
    ast.Forall.canon = _forall_canon

    def _exists_canon(self):
        return '(exists{} bounds:{} body:{})'.format(
            lineno_fields(self), slice_canon(self.bounds),
            node_canon(self.args[0]))
    ast.Exists.canon = _exists_canon

    def _isa_canon(self):
        return '(isa{} terms:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.Isa.canon = _isa_canon

    def _namedbinder_canon(self):
        return '(namedBinder{} name:{} bounds:{} body:{})'.format(
            lineno_fields(self), string_canon(self.name),
            slice_canon(self.bounds), node_canon(self.args[0]))
    ast.NamedBinder.canon = _namedbinder_canon

    def _trigger_canon(self):
        return '(trigger{} pattern:{} terms:{})'.format(
            lineno_fields(self), node_canon(self.args[0]),
            slice_canon(list(self.args[1:])))
    ast.Trigger.canon = _trigger_canon

    # --- Some/SomeMin/SomeMax/SomeExpr ---

    def _some_canon(self):
        params = list(self.args[:-1])
        fmla = self.args[-1]
        return '(some{} params:{} fmla:{})'.format(
            lineno_fields(self), slice_canon(params), node_canon(fmla))
    ast.Some.canon = _some_canon

    def _somemin_canon(self):
        params = list(self.args[:-2])
        fmla = self.args[-2]
        index = self.args[-1]
        return '(someMin{} params:{} fmla:{} index:{})'.format(
            lineno_fields(self), slice_canon(params),
            node_canon(fmla), node_canon(index))
    ast.SomeMin.canon = _somemin_canon

    def _somemax_canon(self):
        params = list(self.args[:-2])
        fmla = self.args[-2]
        index = self.args[-1]
        return '(someMax{} params:{} fmla:{} index:{})'.format(
            lineno_fields(self), slice_canon(params),
            node_canon(fmla), node_canon(index))
    ast.SomeMax.canon = _somemax_canon

    def _someexpr_canon(self):
        param = self.args[0]
        fmla = self.args[1]
        ifval = self.args[2] if len(self.args) >= 3 else None
        elseval = self.args[3] if len(self.args) >= 4 else None
        return '(someExpr{} param:{} fmla:{} ifValue:{} elseVal:{})'.format(
            lineno_fields(self), node_canon(param), node_canon(fmla),
            node_canon(ifval), node_canon(elseval))
    ast.SomeExpr.canon = _someexpr_canon

    # --- Sort types ---

    def _constantsort_canon(self):
        return '(constantSort{} elems:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.ConstantSort.canon = _constantsort_canon

    def _enumeratedsort_canon(self):
        return '(enumeratedSort{} elems:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.EnumeratedSort.canon = _enumeratedsort_canon

    def _structsort_canon(self):
        return '(structSort{} fields:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.StructSort.canon = _structsort_canon

    def _functionsort_canon(self):
        # FunctionSort: args = (*dom, rng) where rng is last
        dom = list(self.args[:-1])
        rng = self.args[-1]
        return '(functionSort{} dom:{} range:{})'.format(
            lineno_fields(self), slice_canon(dom), node_canon(rng))
    ast.FunctionSort.canon = _functionsort_canon

    def _relationsort_canon(self):
        return '(relationSort{} dom:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.RelationSort.canon = _relationsort_canon

    # --- Range ---

    def _range_canon(self):
        lo = self.args[0] if len(self.args) > 0 else None
        hi = self.args[1] if len(self.args) > 1 else None
        return '(range{} lo:{} hi:{})'.format(
            lineno_fields(self), node_canon(lo), node_canon(hi))
    ast.Range.canon = _range_canon

    # --- Declaration types ---

    def _labeledfmla_canon(self):
        label = self.args[0] if len(self.args) > 0 else None
        formula = self.args[1] if len(self.args) > 1 else None
        ##return '(labeledFormula{} label:{} formula:{} lineno:{} temporal:{} explicit:{} isDefinition:{} assumed:{} unprovable:{})'.format(
        return '(labeledFormula label:{} formula:{} id:{} temporal:{} explicit:{} isDefinition:{} assumed:{} unprovable:{})'.format(
            ##lineno_fields(self), 
            node_canon(label), node_canon(formula),
            self.id,
            ##getattr(self, 'lineno', 0) if isinstance(getattr(self, 'lineno', 0), int) else 0,
            bool_canon(self.temporal), bool_canon(self.explicit),
            bool_canon(getattr(self, 'definition', False)),
            bool_canon(self.assumed), bool_canon(self.unprovable))
    ast.LabeledFormula.canon = _labeledfmla_canon

    # --- DeclBase-only types (use decl_fields helper) ---

    def _make_decl_canon(name):
        def canon(self):
            return '({}{})'.format(name, decl_fields(self))
        return canon

    ast.ModuleDecl.canon = _make_decl_canon('moduleDecl')
    ast.MacroDecl.canon = _make_decl_canon('macroDecl')
    ast.ObjectDecl.canon = _make_decl_canon('objectDecl')
    ast.AxiomDecl.canon = _make_decl_canon('axiomDecl')
    ast.PropertyDecl.canon = _make_decl_canon('propertyDecl')
    ast.ConjectureDecl.canon = _make_decl_canon('conjectureDecl')
    ast.ProofDecl.canon = _make_decl_canon('proofDecl')
    ast.NamedDecl.canon = _make_decl_canon('namedDecl')
    ast.SchemaDecl.canon = _make_decl_canon('schemaDecl')
    ast.TheoremDecl.canon = _make_decl_canon('theoremDecl')

    # LabeledDecl is a base class; its subclasses get their own canon
    # via _make_decl_canon above.

    # --- SchemaBody ---
    def _schemabody_canon(self):
        return '(schemaBody{} elems:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.SchemaBody.canon = _schemabody_canon

    # --- Renaming ---
    def _renaming_canon(self):
        return '(renaming{} elems:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.Renaming.canon = _renaming_canon

    # --- Tactic types ---

    def _tactic_canon(self):
        return '(tactic{} elems:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.Tactic.canon = _tactic_canon

    def _schemainstantiation_canon(self):
        sname = self.args[0] if len(self.args) > 0 else None
        ren = self.args[1] if len(self.args) > 1 else None
        matches = list(self.args[2:])
        return '(schemaInstantiation{} schemaName:{} ren:{} matches:{})'.format(
            lineno_fields(self), node_canon(sname), node_canon(ren), slice_canon(matches))
    ast.SchemaInstantiation.canon = _schemainstantiation_canon

    def _assumetactic_canon(self):
        tlabel = getattr(self, 'label', None)
        sname = self.args[0] if len(self.args) > 0 else None
        ren = self.args[1] if len(self.args) > 1 else None
        matches = list(self.args[2:])
        return '(assumeTactic{} tLabel:{} schemaName:{} ren:{} matches:{})'.format(
            lineno_fields(self), node_canon(tlabel), node_canon(sname),
            node_canon(ren), slice_canon(matches))
    ast.AssumeTactic.canon = _assumetactic_canon

    def _assumeglobaltactic_canon(self):
        return '(assumeGlobalTactic assumeTactic:{})'.format(
            _assumetactic_canon(self))
    ast.AssumeGlobalTactic.canon = _assumeglobaltactic_canon

    def _unfoldspec_canon(self):
        defname = self.args[0] if len(self.args) > 0 else None
        renamings = list(self.args[1:])
        return '(unfoldSpec{} defName:{} renamings:{})'.format(
            lineno_fields(self), node_canon(defname), slice_canon(renamings))
    ast.UnfoldSpec.canon = _unfoldspec_canon

    def _unfoldtactic_canon(self):
        tlabel = self.args[0] if len(self.args) > 0 else None
        premise = self.args[1] if len(self.args) > 1 else None
        unfspecs = list(self.args[2:])
        return '(unfoldTactic{} tLabel:{} premise:{} unfSpecs:{})'.format(
            lineno_fields(self), node_canon(tlabel), node_canon(premise),
            slice_canon(unfspecs))
    ast.UnfoldTactic.canon = _unfoldtactic_canon

    def _forgettactic_canon(self):
        return '(forgetTactic{} names:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.ForgetTactic.canon = _forgettactic_canon

    def _showgoalstactic_canon(self):
        return '(showGoalsTactic{})'.format(lineno_fields(self))
    ast.ShowGoalsTactic.canon = _showgoalstactic_canon

    def _defergoaltactic_canon(self):
        return '(deferGoalTactic{} elems:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.DeferGoalTactic.canon = _defergoaltactic_canon

    def _nulltactic_canon(self):
        return '(nullTactic{})'.format(lineno_fields(self))
    ast.NullTactic.canon = _nulltactic_canon

    def _lettactic_canon(self):
        return '(letTactic{} defs:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.LetTactic.canon = _lettactic_canon

    def _witnesstactic_canon(self):
        return '(witnessTactic{} witnesses:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.WitnessTactic.canon = _witnesstactic_canon

    def _spoiltactic_canon(self):
        target = self.args[0] if len(self.args) > 0 else None
        return '(spoilTactic{} target:{})'.format(
            lineno_fields(self), node_canon(target))
    ast.SpoilTactic.canon = _spoiltactic_canon

    def _iftactic_canon(self):
        cond = self.args[0] if len(self.args) > 0 else None
        then = self.args[1] if len(self.args) > 1 else None
        else_ = self.args[2] if len(self.args) > 2 else None
        return '(ifTactic{} cond:{} then:{} else:{})'.format(
            lineno_fields(self), node_canon(cond), node_canon(then), node_canon(else_))
    ast.IfTactic.canon = _iftactic_canon

    def _propertytactic_canon(self):
        prop = self.args[0] if len(self.args) > 0 else None
        pname = self.args[1] if len(self.args) > 1 else None
        proof = self.args[2] if len(self.args) > 2 else None
        return '(propertyTactic{} prop:{} pName:{} proof:{})'.format(
            lineno_fields(self), node_canon(prop), node_canon(pname), node_canon(proof))
    ast.PropertyTactic.canon = _propertytactic_canon

    def _functiontactic_canon(self):
        return '(functionTactic{} elems:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.FunctionTactic.canon = _functiontactic_canon

    def _tactictactic_canon(self):
        tname = self.args[0] if len(self.args) > 0 else None
        body = self.args[1] if len(self.args) > 1 else None
        proof = self.args[2] if len(self.args) > 2 else None
        labels = getattr(self, 'labels', [])
        return '(tacticTactic{} tName:{} body:{} proof:{} labels:{})'.format(
            lineno_fields(self), node_canon(tname), node_canon(body),
            node_canon(proof), string_slice_canon(labels))
    ast.TacticTactic.canon = _tactictactic_canon

    def _prooftactic_canon(self):
        tlabel = self.args[0] if len(self.args) > 0 else None
        proof = self.args[1] if len(self.args) > 1 else None
        return '(proofTactic{} tLabel:{} proof:{})'.format(
            lineno_fields(self), node_canon(tlabel), node_canon(proof))
    ast.ProofTactic.canon = _prooftactic_canon

    def _tacticwith_canon(self):
        return '(tacticWith{} elems:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.TacticWith.canon = _tacticwith_canon

    def _tacticlets_canon(self):
        return '(tacticLets{} lets:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.TacticLets.canon = _tacticlets_canon

    def _composetactics_canon(self):
        return '(composeTactics{} tactics:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    ast.ComposeTactics.canon = _composetactics_canon

    # --- Remaining ivy_ast types ---

    # IsolateDef and variants
    def _isolatedef_canon(self):
        wa = getattr(self, 'with_args', 0)
        iso = getattr(self, 'is_object', False)
        trusted = isinstance(self, ast.TrustedIsolateDef) if hasattr(ast, 'TrustedIsolateDef') else False
        return '(isolateDef{} elems:{} withArgs:{} trusted:{} isObject:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)), wa,
            bool_canon(trusted), bool_canon(iso))
    ast.IsolateDef.canon = _isolatedef_canon

    if hasattr(ast, 'TrustedIsolateDef'):
        def _trustedisolatedef_canon(self):
            wa = getattr(self, 'with_args', 0)
            iso = getattr(self, 'is_object', False)
            return '(trustedIsolateDef{} elems:{} withArgs:{} trusted:{} isObject:{})'.format(
                lineno_fields(self), slice_canon(list(self.args)), wa,
                bool_canon(True), bool_canon(iso))
        ast.TrustedIsolateDef.canon = _trustedisolatedef_canon

    if hasattr(ast, 'ExtractDef'):
        def _extractdef_canon(self):
            wa = getattr(self, 'with_args', 0)
            iso = getattr(self, 'is_object', False)
            return '(extractDef{} elems:{} withArgs:{} trusted:{} isObject:{})'.format(
                lineno_fields(self), slice_canon(list(self.args)), wa,
                bool_canon(False), bool_canon(iso))
        ast.ExtractDef.canon = _extractdef_canon

    if hasattr(ast, 'ProcessDef'):
        def _processdef_canon(self):
            wa = getattr(self, 'with_args', 0)
            iso = getattr(self, 'is_object', False)
            return '(processDef{} elems:{} withArgs:{} trusted:{} isObject:{})'.format(
                lineno_fields(self), slice_canon(list(self.args)), wa,
                bool_canon(False), bool_canon(iso))
        ast.ProcessDef.canon = _processdef_canon

    # ActionDef
    def _actiondef_canon(self):
        name = self.args[0] if len(self.args) > 0 else None
        body = self.args[1] if len(self.args) > 1 else None
        fps = getattr(self, 'formal_params', [])
        frs = getattr(self, 'formal_returns', [])
        return '(actionDef{} name:{} body:{} formalParams:{} formalReturns:{})'.format(
            lineno_fields(self), node_canon(name), node_canon(body),
            slice_canon(fps), slice_canon(frs))
    ast.ActionDef.canon = _actiondef_canon

    # TypeDef
    if hasattr(ast, 'TypeDef'):
        def _typedef_canon(self):
            name = self.args[0] if len(self.args) > 0 else None
            value = self.args[1] if len(self.args) > 1 else None
            finite = getattr(self, 'finite', False)
            return '(typeDef{} name:{} value:{} finite:{})'.format(
                lineno_fields(self), node_canon(name), node_canon(value), bool_canon(finite))
        ast.TypeDef.canon = _typedef_canon

    if hasattr(ast, 'GhostTypeDef'):
        def _ghosttypedef_canon(self):
            name = self.args[0] if len(self.args) > 0 else None
            value = self.args[1] if len(self.args) > 1 else None
            finite = getattr(self, 'finite', False)
            return '(ghostTypeDef{} name:{} value:{} finite:{})'.format(
                lineno_fields(self), node_canon(name), node_canon(value), bool_canon(finite))
        ast.GhostTypeDef.canon = _ghosttypedef_canon

    # --- Remaining Decl types ---
    for name, cls in [
        ('relationDecl', getattr(ast, 'RelationDecl', None)),
        ('constantDecl', getattr(ast, 'ConstantDecl', None)),
        ('parameterDecl', getattr(ast, 'ParameterDecl', None)),
        ('destructorDecl', getattr(ast, 'DestructorDecl', None)),
        ('constructorDecl', getattr(ast, 'ConstructorDecl', None)),
        ('typeDecl', getattr(ast, 'TypeDecl', None)),
        ('variantDecl', getattr(ast, 'VariantDecl', None)),
        ('derivedDecl', getattr(ast, 'DerivedDecl', None)),
        ('definitionDecl', getattr(ast, 'DefinitionDecl', None)),
        ('progressDecl', getattr(ast, 'ProgressDecl', None)),
        ('relyDecl', getattr(ast, 'RelyDecl', None)),
        ('mixOrdDecl', getattr(ast, 'MixOrdDecl', None)),
        ('conceptDecl', getattr(ast, 'ConceptDecl', None)),
        ('initDecl', getattr(ast, 'InitDecl', None)),
        ('stateDecl', getattr(ast, 'StateDecl', None)),
        ('updateDecl', getattr(ast, 'UpdateDecl', None)),
        ('assertDecl', getattr(ast, 'AssertDecl', None)),
        ('interpretDecl', getattr(ast, 'InterpretDecl', None)),
        ('mixinDecl', getattr(ast, 'MixinDecl', None)),
        ('isolateDecl', getattr(ast, 'IsolateDecl', None)),
        ('isolateObjectDecl', getattr(ast, 'IsolateObjectDecl', None)),
        ('exportDecl', getattr(ast, 'ExportDecl', None)),
        ('importDecl', getattr(ast, 'ImportDecl', None)),
        ('privateDecl', getattr(ast, 'PrivateDecl', None)),
        ('aliasDecl', getattr(ast, 'AliasDecl', None)),
        ('delegateDecl', getattr(ast, 'DelegateDecl', None)),
        ('implementTypeDecl', getattr(ast, 'ImplementTypeDecl', None)),
        ('nativeDecl', getattr(ast, 'NativeDecl', None)),
        ('attributeDecl', getattr(ast, 'AttributeDecl', None)),
        ('instantiateDecl', getattr(ast, 'InstantiateDecl', None)),
        ('autoInstanceDecl', getattr(ast, 'AutoInstanceDecl', None)),
        ('scenarioDecl', getattr(ast, 'ScenarioDecl', None)),
        ('subclassDecl', getattr(ast, 'SubclassDecl', None)),
        ('actionDecl', getattr(ast, 'ActionDecl', None)),
        ('freshConstantDecl', getattr(ast, 'FreshConstantDecl', None)),
    ]:
        if cls is not None:
            cls.canon = _make_decl_canon(name)

    # --- Def types (Python AST subclasses held by Decl nodes) ---
    # Go emits these with their specific fields. Python previously fell
    # through to _ast_canon (empty). Match Go's format for visibility.

    def _make_empty_def_canon(name):
        def _canon(self):
            return '({}{})'.format(name, lineno_fields(self))
        return _canon

    # Empty-body Defs: Go emits "(defName{lf})" with no body fields.
    for _name, _cls_name in [
        ('attributeDef', 'AttributeDef'),
        ('delegateDef', 'DelegateDef'),
        ('importDef', 'ImportDef'),
        ('exportDef', 'ExportDef'),
        ('instantiation', 'Instantiation'),
        ('mixinAfterDef', 'MixinAfterDef'),
        ('mixinBeforeDef', 'MixinBeforeDef'),
        ('mixinImplementDef', 'MixinImplementDef'),
        ('nativeCode', 'NativeCode'),
        ('nativeType', 'NativeType'),
        ('nativeExpr', 'NativeExpr'),
        ('nativeDef', 'NativeDef'),
        ('placeList', 'PlaceList'),
        ('scenarioTransition', 'ScenarioTransition'),
        ('stateDef', 'StateDef'),
    ]:
        _cls = getattr(ast, _cls_name, None)
        if _cls is not None:
            _cls.canon = _make_empty_def_canon(_name)

    # Defs with extra structural fields:

    def _variantdef_canon(self):
        # Python VariantDef.__init__(self,name,sort): self.args = [name, sort]
        name = self.args[0] if len(self.args) > 0 else None
        vsort = self.args[1] if len(self.args) > 1 else None
        return '(variantDef{} name:{} vSort:{})'.format(
            lineno_fields(self), node_canon(name), node_canon(vsort))
    if hasattr(ast, 'VariantDef'):
        ast.VariantDef.canon = _variantdef_canon

    def _scenariodef_canon(self):
        return '(scenarioDef{} elems:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    if hasattr(ast, 'ScenarioDef'):
        ast.ScenarioDef.canon = _scenariodef_canon

    def _scenariobeforemixin_canon(self):
        # Parser: ScenarioBeforeMixin(mixer, df) → args[0]=mixer, args[1]=def
        mixer = self.args[0] if len(self.args) > 0 else None
        defn = self.args[1] if len(self.args) > 1 else None
        return '(scenarioBeforeMixin{} mixer:{} def:{})'.format(
            lineno_fields(self), node_canon(mixer), node_canon(defn))
    if hasattr(ast, 'ScenarioBeforeMixin'):
        ast.ScenarioBeforeMixin.canon = _scenariobeforemixin_canon

    def _scenarioaftermixin_canon(self):
        mixer = self.args[0] if len(self.args) > 0 else None
        defn = self.args[1] if len(self.args) > 1 else None
        return '(scenarioAfterMixin{} mixer:{} def:{})'.format(
            lineno_fields(self), node_canon(mixer), node_canon(defn))
    if hasattr(ast, 'ScenarioAfterMixin'):
        ast.ScenarioAfterMixin.canon = _scenarioaftermixin_canon

    def _privatedef_canon(self):
        return '(privateDef{} elems:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    if hasattr(ast, 'PrivateDef'):
        ast.PrivateDef.canon = _privatedef_canon

    def _implementtypedef_canon(self):
        return '(implementTypeDef{} elems:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    if hasattr(ast, 'ImplementTypeDef'):
        ast.ImplementTypeDef.canon = _implementtypedef_canon

    # --- Action types (ivy_actions.py) ---
    # Use the AST elems: format for both .sexp and .canon, matching
    # Go's ast.*Action.Canon() and actions.*Action.Sexp() (now converged).

    def _sequence_canon(self):
        return '(sequence{} stmts:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    act.Sequence.sexp = _sequence_canon
    act.Sequence.canon = _sequence_canon

    def _callaction_canon(self):
        uid = getattr(self, 'unique_id', 0)
        return '(callAction{} elems:{} uniqueID:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)), uid)
    act.CallAction.sexp = _callaction_canon
    act.CallAction.canon = _callaction_canon

    def _crashaction_canon(self):
        return '(crashAction{} declArgs:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    act.CrashAction.sexp = _crashaction_canon
    act.CrashAction.canon = _crashaction_canon

    def _thunkaction_canon(self):
        return '(thunkAction{} children:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)))
    act.ThunkAction.sexp = _thunkaction_canon
    act.ThunkAction.canon = _thunkaction_canon

    def _choiceaction_canon(self):
        uid = getattr(self, 'unique_id', 0)
        return '(choiceAction{} elems:{} uniqueID:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)), uid)
    act.ChoiceAction.sexp = _choiceaction_canon
    act.ChoiceAction.canon = _choiceaction_canon

    def _envaction_canon(self):
        uid = getattr(self, 'unique_id', 0)
        return '(envAction{} elems:{} uniqueID:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)), uid)
    act.EnvAction.sexp = _envaction_canon
    act.EnvAction.canon = _envaction_canon

    def _ifaction_canon(self):
        cond = self.args[0] if len(self.args) > 0 else None
        then = self.args[1] if len(self.args) > 1 else None
        else_ = self.args[2] if len(self.args) > 2 else None
        return '(ifAction{} cond:{} then:{} else:{})'.format(
            lineno_fields(self), node_canon(cond), node_canon(then), node_canon(else_))
    act.IfAction.sexp = _ifaction_canon
    act.IfAction.canon = _ifaction_canon

    def _localaction_canon(self):
        uid = getattr(self, 'unique_id', 0)
        return '(localAction{} elems:{} uniqueID:{})'.format(
            lineno_fields(self), slice_canon(list(self.args)), uid)
    act.LocalAction.sexp = _localaction_canon
    act.LocalAction.canon = _localaction_canon

    def _return_action_canon(self):
        return '(returnAction)'
    act.ReturnAction.sexp = _return_action_canon
    act.ReturnAction.canon = _return_action_canon

    def _ignore_action_canon(self):
        return '(ignoreAction)'
    act.IgnoreAction.sexp = _ignore_action_canon
    act.IgnoreAction.canon = _ignore_action_canon

    # Generic action: (typeName{} elems:[args])
    def _action_generic_canon(self):
        name = type(self).__name__
        name = name[0].lower() + name[1:]
        return '({}{} elems:{})'.format(
            name, lineno_fields(self), slice_canon(list(self.args)))

    # Install generic canon for action types that don't have specific methods
    for cls in [act.Action, act.AssignAction, act.HavocAction, act.SetAction,
                act.AssignFieldAction, act.NullFieldAction, act.CopyFieldAction,
                act.InstantiateAction, act.WhileAction, act.Ranking,
                act.AssumeAction, act.AssertAction, act.EnsuresAction,
                act.RequiresAction, act.SubgoalAction, act.LetAction,
                act.DebugAction, act.NativeAction, act.BindOldsAction]:
        if not hasattr(cls, 'canon') or cls.canon is _ast_canon:
            cls.canon = _action_generic_canon
            cls.sexp = _action_generic_canon

    # Schema (from ivy_actions.py)
    if hasattr(act, 'Schema'):
        def _schema_canon(self):
            defn = getattr(self, 'defn', self.args[0] if self.args else None)
            fresh = getattr(self, 'fresh', [])
            instances = getattr(self, 'instances', [])
            return '(schema{} defn:{} fresh:{} instances:{})'.format(
                lineno_fields(self), node_canon(defn),
                slice_canon(fresh), slice_canon(instances))
        act.Schema.canon = _schema_canon

    # RME
    if hasattr(act, 'RME'):
        def _rme_canon(self):
            req = self.args[0] if len(self.args) > 0 else None
            mod = list(self.args[1]) if len(self.args) > 1 and self.args[1] is not None else []
            ens = self.args[2] if len(self.args) > 2 else None
            return '(rME{} requiresFmla:{} modifiesList:{} ensuresFmla:{})'.format(
                lineno_fields(self), node_canon(req), slice_canon(mod), node_canon(ens))
        act.RME.canon = _rme_canon

    # SymbolList, UpdatePattern, UpdatePatternList, PatternBasedUpdate
    if hasattr(act, 'SymbolList'):
        def _symbollist_canon(self):
            return '(symbolList{} elems:{})'.format(
                lineno_fields(self), slice_canon(list(self.args)))
        act.SymbolList.canon = _symbollist_canon

    if hasattr(act, 'UpdatePatternList'):
        def _updatepatternlist_canon(self):
            return '(updatePatternList{} elems:{})'.format(
                lineno_fields(self), slice_canon(list(self.args)))
        act.UpdatePatternList.canon = _updatepatternlist_canon

    if hasattr(act, 'UpdatePattern'):
        def _updatepattern_canon(self):
            params = self.args[0] if len(self.args) > 0 else None
            action = self.args[1] if len(self.args) > 1 else None
            requires = self.args[2] if len(self.args) > 2 else None
            ensures = self.args[3] if len(self.args) > 3 else None
            return '(updatePattern{} params:{} action:{} requires:{} ensures:{})'.format(
                lineno_fields(self), node_canon(params), node_canon(action),
                node_canon(requires), node_canon(ensures))
        act.UpdatePattern.canon = _updatepattern_canon

    if hasattr(act, 'PatternBasedUpdate'):
        def _patternbasedupdate_canon(self):
            dfns = self.args[0] if len(self.args) > 0 else None
            deps = self.args[1] if len(self.args) > 1 else None
            patterns = self.args[2] if len(self.args) > 2 else None
            return '(patternBasedUpdate{} dfns:{} deps:{} patterns:{})'.format(
                lineno_fields(self), node_canon(dfns), node_canon(deps), node_canon(patterns))
        act.PatternBasedUpdate.canon = _patternbasedupdate_canon

    # DerivedUpdate — Python stores self.defn (ivy_logic.Definition) which has .sexp()
    if hasattr(act, 'DerivedUpdate'):
        def _derivedupdate_canon(self):
            defn_str = self.defn.sexp() if hasattr(self.defn, 'sexp') else str(self.defn)
            return '(DerivedUpdate defn:%s)' % defn_str
        act.DerivedUpdate.canon = _derivedupdate_canon

    # NamedUpdate — Python stores self.sym (string)
    if hasattr(act, 'NamedUpdate'):
        def _namedupdate_canon(self):
            return '(NamedUpdate sym:"%s")' % str(self.sym)
        act.NamedUpdate.canon = _namedupdate_canon

    # VarAction
    if hasattr(act, 'VarAction'):
        def _varaction_canon(self):
            return '(varAction{} elems:{})'.format(
                lineno_fields(self), slice_canon(list(self.args)))
        act.VarAction.canon = _varaction_canon

    # --- Temporal types (ivy_temporal.py) ---
    from . import ivy_temporal as itm

    def _actionterm_canon(self):
        return '(actionTerm{} inputs:{} outputs:{} labels:{} stmt:{})'.format(
            lineno_fields(self), slice_canon(self.inputs), slice_canon(self.outputs),
            string_slice_canon(self.labels), node_canon(self.stmt))
    itm.ActionTerm.canon = _actionterm_canon
    itm.ActionTerm.sexp = _actionterm_canon

    def _actiontermbinding_canon(self):
        return '(actionTermBinding{} name:{} action:{})'.format(
            lineno_fields(self), string_canon(self.name), node_canon(self.action))
    itm.ActionTermBinding.canon = _actiontermbinding_canon
    itm.ActionTermBinding.sexp = _actiontermbinding_canon

    # --- TemporalModels (ivy_ast.py) ---
    # Go: "(temporalModels{lf} model:{} fmla:{})" at ast/ast.go:1683.
    def _temporalmodels_canon(self):
        return '(temporalModels{} model:{} fmla:{})'.format(
            lineno_fields(self),
            node_canon(self.model),
            node_canon(self.fmla))
    if hasattr(ast, 'TemporalModels'):
        ast.TemporalModels.canon = _temporalmodels_canon

    # --- NormalProgram (ivy_temporal.py) ---
    # Go: "(normalProgram{lf} bindings:{} init:{} invars:{} asms:{} calls:{} postconds:{})"
    # postconds is a dict — sort keys for determinism.
    def _normalprogram_canon(self):
        postconds_sexp = '(hash)'
        if hasattr(self, 'postconds') and self.postconds:
            parts = []
            for k in sorted(self.postconds.keys()):
                parts.append('{}:{}'.format(
                    string_canon(k), slice_canon(list(self.postconds[k]))))
            postconds_sexp = '(hash ' + ' '.join(parts) + ')'
        return '(normalProgram{} bindings:{} init:{} invars:{} asms:{} calls:{} postconds:{})'.format(
            lineno_fields(self),
            slice_canon(list(self.bindings)),
            node_canon(self.init),
            slice_canon(list(self.invars)),
            slice_canon(list(self.asms)),
            string_slice_canon(list(self.calls)),
            postconds_sexp)
    itm.NormalProgram.canon = _normalprogram_canon

    # Install canon on fragment checker types (UFNode etc.)
    from .canon_fragment import install as install_fragment
    install_fragment()
