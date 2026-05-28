#
# Copyright (c) Microsoft Corporation. All Rights Reserved.
#

from . import ivy_utils as iu
from . import ivy_logic as il
from . import ivy_logic_utils as lu
from . import ivy_ast as ia
from . import logic_util
from . import xtracer

class Redefinition(iu.IvyError):
    pass
class Circular(iu.IvyError):
    pass
class NoMatch(iu.IvyError):
    pass
class ProofError(iu.IvyError):
    pass
class CaptureError(iu.IvyError):
    pass



class MatchProblem(object):
    def __init__(self,schema,pat,inst,freesyms,constants,prem_matches=[]):
        self.schema,self.pat,self.inst,self.freesyms,self.constants = schema,pat,inst,set(freesyms),constants
        self.prem_matches=prem_matches
        self.revmap = dict()
    def __str__(self):
        return '{{pat:{},inst:{},freesyms:{}}}'.format(self.pat,self.inst,list(map(str,self.freesyms)))

def attrib_goals(proof,goals):
    if hasattr(proof,'lineno'):
        for g in goals:
            g.lineno = proof.lineno
    return goals

def _prop_label_name(prop):
    try:
        label = prop.label
        if label is None:
            return ''
        if hasattr(label,'relname'):
            return label.relname
        return str(label)
    except Exception:
        try:
            return prop.name
        except Exception:
            return ''

def _goal_label_trace(goal):
    try:
        label = goal.label
        if label is None:
            return ''
        return str(label)
    except Exception:
        return 'N/A'

class ProofChecker(object):
    """ This is IVY's built-in proof checker """

    def __init__(self,axioms,definitions,schemata=None):
        """ A proof checker starts with sets of axioms, definitions and schemata
    
        - axioms is a list of ivy_ast.LabeledFormula
        - definitions is a list of ivy_ast.LabeledFormula
        - schemata is a map from string names to ivy_ast.LabeledFormula

        The schemata argument is optional and is included for backward compatibility
        with ivy_mc.
        """
    
        self.axioms  = [normalize_goal(ax) for ax in axioms]
        self.definitions = dict((d.formula.defines().name,normalize_goal(d)) for d in definitions)
        self.schemata = dict()
        if schemata is not None:
            for _skey, _sval in schemata.items():
                _norm = normalize_goal(_sval)
                if __debug__: xtracer.trace("proof.ProofChecker.__init__.fromMod schemata.insert key='%s' value=%s" % (_skey, _norm.canon()))
                self.schemata[_skey] = _norm
        for ax in axioms:
            if ax.label is not None:
                if __debug__: xtracer.trace("proof.ProofChecker.__init__.axiom schemata.insert key='%s' value=%s" % (ax.name, ax.canon()))
                self.schemata[ax.name] = ax
        self.stale = set() # set of symbols that are not fresh
        for lf in axioms + definitions:
            self.stale.update(lu.used_symbols_ast(lf.formula))
        for goal in list(schemata.values()):
            vocab = goal_vocab(goal)
            self.stale.update(vocab.symbols)

    def admit_axiom(self,ax):
        self.axioms.append(normalize_goal(ax))
        if ax.label is not None:
            if __debug__: xtracer.trace("proof.ProofChecker.admit_axiom schemata.insert key='%s' value=%s" % (ax.name, ax.canon()))
            self.schemata[ax.name] = ax

    def admit_definition(self,defn,proof=None):
        """ Admits a definition if it is non-recursive or match a definition schema.
            If a proof is given it is used to match the definition to a schema, else
            default heuristic matching is used.

        - defn is an ivy_ast.LabeledFormula
        """

        if __debug__: xtracer.trace("proof.AdmitDefinition ENTER defnLabel=%s hasProof=%s" % (defn.name, proof is not None))
        defn = normalize_goal(defn)
        sym = defn.formula.defines()
        if sym.name in self.definitions:
            raise Redefinition(defn,"redefinition of {}".format(sym))
        if sym in self.stale:
            raise Circular(defn,"symbol {} defined after reference".format(sym))
        deps = list(lu.symbols_ilu_ast(defn.formula.rhs()))
        self.stale.update(deps)
        if sym in deps:
            # Recursive definitions must match a schema
            if proof is None:
                if __debug__: xtracer.trace("proof.AdmitDefinition EXIT err=noProof")
                raise NoMatch(defn,"no proof given for recursive definition")
            subgoals = self.apply_proof([defn],proof)
            if subgoals is None:
                if __debug__: xtracer.trace("proof.AdmitDefinition EXIT err=%s" % "recursive definition does not match the given schema")
                raise NoMatch(defn,"recursive definition does not match the given schema")
        else:
            subgoals = []
        self.definitions[sym.name] = defn
        if __debug__: xtracer.trace("proof.AdmitDefinition EXIT nsubgoals=%d sym=%s" % (len(subgoals), sym.name))
        return subgoals
        
    def admit_proposition(self,prop,proof=None,subgoals=None):
        """ Admits a proposition with proof.  If a proof is given it
            is used to match the definition to a schema, else default
            heuristic matching is used. If a list of subgoals is supplied, it is
            assumed that these entail prop and the proof is applied to
            the subgoals.

        - prop is an ivy_ast.LabeledFormula
        """

        if __debug__: xtracer.trace("proof.AdmitProposition ENTER propLabel=%s hasProof=%s nExistingSubgoals=%d" % (_prop_label_name(prop), proof is not None, len(subgoals or [])))
        prop = normalize_goal(prop)
        if isinstance(prop.formula,il.Definition):
            if __debug__: xtracer.trace("proof.AdmitProposition delegateToDefinition")
            return self.admit_definition(prop,proof)
        if proof is None:
            if __debug__: xtracer.trace("proof.AdmitProposition EXIT err=noProof")
            raise NoMatch(prop,"no proof given for property")
        subgoals = subgoals or [prop]
        subgoals = self.apply_proof(subgoals,proof)
        if subgoals is None:
            if __debug__: xtracer.trace("proof.AdmitProposition EXIT err=%s" % "goal does not match the given schema")
            raise NoMatch(proof,"goal does not match the given schema")
        self.axioms.append(prop)
        if prop.label is not None:
            if __debug__: xtracer.trace("proof.ProofChecker.admit_proposition schemata.insert key='%s' value=%s" % (prop.name, prop.canon()))
            self.schemata[prop.name] = prop
        vocab = goal_vocab(prop)
        self.stale.update(vocab.symbols)
        if __debug__: xtracer.trace("proof.AdmitProposition EXIT nsubgoals=%d" % len(subgoals))
        return subgoals

    def get_subgoals(self,prop,proof):
        """Return the subgoals that result from applying proof to property
            prop, but do not admit prop in the context. Note, prop may not
            be a definition.

        """
        if __debug__: xtracer.trace("proof.GetSubgoals ENTER propLabel=%s" % _prop_label_name(prop))
        assert not isinstance(prop.formula,il.Definition)
        prop = normalize_goal(prop)
        subgoals = self.apply_proof([prop],proof)
        if subgoals is None:
            if __debug__: xtracer.trace("proof.GetSubgoals EXIT err=%s" % "goal does not match the given schema")
            raise NoMatch(proof,"goal does not match the given schema")
        if __debug__: xtracer.trace("proof.GetSubgoals EXIT nsubgoals=%d" % len(subgoals))
        return subgoals
        

    def apply_proof(self,decls,proof):
        """ Apply a proof to a list of goals, producing subgoals, or None if
        the proof fails. """

        with ia.ASTContext(proof):
            if len(decls) == 0:
                return []
            if len(decls) > 0 and decls[0] is not None:
                if __debug__: xtracer.trace("proof.ApplyProof ENTER proofType=%s goal[0].Formula type=%s" % (type(proof).__name__, type(decls[0].formula).__name__ if hasattr(decls[0],'formula') else 'N/A'))
            if isinstance(proof,ia.SchemaInstantiation):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=SchemaInstantiation")
                m = self.match_schema(decls[0],proof)
                result = None if m is None else m + decls[1:]
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=SchemaInstantiation ngoals=%d" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.LetTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=LetTactic")
                result = self.let_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=LetTactic ngoals=%d err=<nil>" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.ComposeTactics):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=ComposeTactics")
                result = self.compose_proofs(decls,proof.args)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=ComposeTactics ngoals=%d err=<nil>" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.AssumeTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=AssumeTactic")
                result = self.assume_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=AssumeTactic ngoals=%d err=<nil>" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.UnfoldTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=UnfoldTactic")
                result = self.unfold_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=UnfoldTactic ngoals=%d err=<nil>" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.ForgetTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=ForgetTactic")
                result = self.forget_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=ForgetTactic ngoals=%d err=<nil>" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.ShowGoalsTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=ShowGoalsTactic")
                result = self.show_goals_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=ShowGoalsTactic ngoals=%d" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.DeferGoalTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=DeferGoalTactic")
                result = self.defer_goal_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=DeferGoalTactic ngoals=%d" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.DeferGoalTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=DeferGoalTactic")
                result = self.defer_goal_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=DeferGoalTactic ngoals=%d" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.LetTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=LetTactic")
                result = self.let_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=LetTactic ngoals=%d err=<nil>" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.IfTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=IfTactic")
                result = self.if_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=IfTactic ngoals=%d err=<nil>" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.NullTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=NullTactic")
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=NullTactic ngoals=%d" % len(decls))
                return decls
            elif isinstance(proof,ia.PropertyTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=PropertyTactic")
                result = self.property_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=PropertyTactic ngoals=%d err=<nil>" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.FunctionTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=FunctionTactic")
                result = self.function_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=FunctionTactic ngoals=%d err=<nil>" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.TacticTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=TacticTactic")
                result = self.tactic_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=TacticTactic ngoals=%d err=<nil>" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.ProofTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=ProofTactic")
                result = self.proof_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=ProofTactic ngoals=%d err=<nil>" % (len(result) if result is not None else -1))
                return result
            elif isinstance(proof,ia.WitnessTactic):
                if __debug__: xtracer.trace("proof.ApplyProof dispatch name=WitnessTactic")
                result = self.witness_tactic(decls,proof)
                if __debug__: xtracer.trace("proof.ApplyProof EXIT proofType=WitnessTactic ngoals=%d err=<nil>" % (len(result) if result is not None else -1))
                return result
            assert False,"unknown proof type {}".format(type(proof))

    def proof_tactic(self,decls,proof):
        labelStr = str(proof.label)
        if __debug__: xtracer.trace("proof.proofTactic ENTER label=%s ndecls=%d" % (labelStr, len(decls)))
        for idx,decl in enumerate(decls):
            if decl.label.rep == proof.label.rep:
                subgoals = self.apply_proof([decl],proof.proof)
                rest = decls[:idx] + decls[idx+1:] + subgoals
                if __debug__: xtracer.trace("proof.proofTactic EXIT nsubgoals=%d ndecls=%d" % (len(subgoals) if subgoals is not None else -1, len(rest) if rest is not None else -1))
                return rest
        if __debug__: xtracer.trace("proof.proofTactic EXIT err=noLabel label=%s" % labelStr)
        raise iu.IvyError(proof,'no goal with label {}'.format(proof.label))

    def tactic_tactic(self,decls,proof):
        tn = proof.tactic_name
        if len(decls) > 0 and decls[0] is not None:
            if __debug__: xtracer.trace("proof.tacticTactic name=%r goal[0].Formula type=%s" % (tn, type(decls[0].formula).__name__ if hasattr(decls[0],'formula') else 'N/A'))
        if tn not in registered_tactics:
            if __debug__: xtracer.trace("proof.tacticTactic EXIT name=%s err=unknownTactic" % tn)
            raise iu.IvyError(proof,'unknown tactic: {}'.format(tn))
        tactic = registered_tactics[tn]
        try:
            result = tactic(self,decls,proof)
        except BaseException as _tt_e:
            if __debug__: xtracer.trace("proof.tacticTactic EXIT name=%s nresult=-1 err=%s" % (tn, str(_tt_e)))
            raise
        if __debug__: xtracer.trace("proof.tacticTactic EXIT name=%s nresult=%d err=<nil>" % (tn, len(result) if result is not None else -1))
        return result

    def compose_proofs(self,decls,proofs):
        if __debug__: xtracer.trace("proof.composeProofs ENTER nproofs=%d ndecls=%d" % (len(proofs), len(decls)))
        for i,proof in enumerate(proofs):
            if len(decls) > 0 and decls[0] is not None:
                if __debug__: xtracer.trace("proof.composeProofs step=%d/%d proofType=%s goal[0].Formula type=%s" % (i, len(proofs), type(proof).__name__, type(decls[0].formula).__name__ if hasattr(decls[0],'formula') else 'N/A'))
            decls = self.apply_proof(decls,proof)
            if decls is None or len(decls) == 0:
                if __debug__: xtracer.trace("proof.composeProofs EXIT ndecls=%d step=%d" % (len(decls) if decls is not None else -1, i))
                return decls
        if __debug__: xtracer.trace("proof.composeProofs EXIT ndecls=%d" % len(decls))
        return decls

    def show_goals_tactic(self,decls,proof):
        if __debug__: xtracer.trace("proof.showGoalsTactic ENTER ndecls=%d" % len(decls))
        print()
        print('{}Proof goals:'.format(proof.lineno))
        for decl in decls:
            print()
            print('theorem ' + str(decl))
            print()
        if __debug__: xtracer.trace("proof.showGoalsTactic EXIT ndecls=%d" % len(decls))
        return decls

    def defer_goal_tactic(self,decls,proof):
        if __debug__: xtracer.trace("proof.deferGoalTactic ENTER ndecls=%d" % len(decls))
        result = decls[1:] + decls[0:1]
        if __debug__: xtracer.trace("proof.deferGoalTactic EXIT ndecls=%d" % len(result))
        return result

    def let_tactic(self,decls,proof):
        if __debug__: xtracer.trace("proof.letTactic ENTER ndecls=%d nDefs=%d" % (len(decls), len(proof.args)))
        goal = decls[0]
        vocab = goal_vocab(goal)
        defs = [compile_expr_vocab(ia.Atom('=',x.args[0],x.args[1]),vocab) for x in proof.args]
        cond = il.And(*[il.Equals(a.args[0],a.args[1]) for a in defs])
        if __debug__: xtracer.trace("proof.WrapImplies ENTER formulaType=%s" % type(decls[0].formula).__name__)
        impl = il.Implies(cond,decls[0].formula)
        if __debug__: xtracer.trace("proof.WrapImplies EXIT type=lgImplies HASH canon=%s" % (impl.canon() if hasattr(impl,'canon') else str(impl)))
        subgoal = ia.LabeledFormula(decls[0].label,impl)
        if not hasattr(decls[0],'lineno'):
            print('has no line number: {}'.format(decls[0]))
            exit(1)
        subgoal.lineno = decls[0].lineno
        if __debug__: xtracer.trace("proof.letTactic EXIT HASH canon=%s" % subgoal.canon())
        return attrib_goals(proof,[subgoal]) + decls[1:]

    def property_tactic(self,decls,proof):
        if __debug__: xtracer.trace("proof.propertyTactic ENTER ndecls=%d" % len(decls))
        goal = decls[0]
        vocab = goal_vocab(goal)
        cut = compile_expr_vocab(proof.args[0],vocab)
        cut = normalize_goal(cut)
        subgoal = goal_subst(goal,cut,cut.lineno)
        lhs = proof.args[1]
        if not isinstance(lhs,ia.NoneAST):
            fmla = il.drop_universals(cut.formula)
            if not il.is_exists(fmla) or len(fmla.variables) != 1:
                raise IvyError(proof,'property is not existential')
            evar = list(fmla.variables)[0]
            rng = evar.sort
            vmap = dict((x.name,x) for x in lu.variables_ast(fmla))
            used = set()
            args = lhs.args
            targs = []
            for a in args:
                if a.name in used:
                    raise IvyError(lhs,'repeat parameter: {}'.format(a.name))
                used.add(a.name)
                if a.name in vmap:
                    v = vmap[a.name]
                    targs.append(v)
                    if not (il.is_topsort(a.sort) or a.sort != v.sort):
                        raise IvyError(lhs,'bad sort for {}'.format(a.name))
                else:
                    if il.is_topsort(a.sort):
                        raise IvyError(lhs,'cannot infer sort for {}'.format(a.name))
                    targs.append(a)
            for x in vmap:
                if x not in used:
                    raise IvyError(lhs,'{} must be a parameter of {}'.format(x,lhs.rep))
            dom = [x.sort for x in targs]
            sym = il.Symbol(lhs.rep,il.FuncConstSort(*(dom+[rng])))
            if sym in self.stale or sym in goal_defns(goal):
                raise iu.IvyError(lhs,'{} is not fresh'.format(sym))
            term = sym(*targs) if targs else sym
            fmla = lu.substitute_ast(fmla.body,{evar.name:term})
            cut = clone_goal(cut,[],fmla)
            goal = goal_add_prem(goal,ia.ConstantDecl(sym),goal.lineno)
            if __debug__: xtracer.trace("proof.propertyTactic skolem sym=%s nTargs=%d" % (sym.name, len(targs)))
        
        subgoals = [subgoal]
        pf = proof.args[2]
        if not isinstance(pf,ia.NoneAST):
            subgoals = self.apply_proof(subgoals,pf)
            if subgoals is None:
                if __debug__: xtracer.trace("proof.propertyTactic EXIT applied=None")
                return None
        result = [goal_add_prem(goal,cut,cut.lineno)] + decls[1:] + subgoals
        if __debug__: xtracer.trace("proof.propertyTactic EXIT nresult=%d nsubgoals=%d" % (len(result), len(subgoals)))
        return result

    def function_tactic(self,decls,proof):
        if __debug__: xtracer.trace("proof.functionTactic ENTER ndecls=%d nElems=%d" % (len(decls), len(proof.args)))
        goal = decls[0]
        vocab = goal_vocab(goal)
        free = goal_free(goal)
        for df in proof.args:
            if isinstance(df,ia.ConstantDecl):
                assert False
            else:
                lf = df.args[0]
                lhs = lf.formula.args[0]
                ts = il.TopFunctionSort(len(lhs.args))
                newsym = il.Symbol(lhs.rep,ts)
                with il.WithSymbols([newsym]):
                    vars = lf.formula.args[0].args
                    fmla = ia.Forall(vars,ia.Atom('=',lf.formula.args))
                    elf = lf.clone([lf.label,fmla])
                    lf = compile_expr_vocab(elf,vocab)
                sym = lf.formula.body.args[0].rep
                deps = list(lu.symbols_ilu_ast(lf.formula.body.args[1]))
                if sym in deps:
                    raise NoMatch(lf,"no proof given for recursive definition")
                # TODO: allow proofs of recursive definitions
                cd = ia.ConstantDecl(sym)
                cd.lineno = lf.lineno
                lf.definition = True
                goal = goal_add_prem(goal,cd,lf.lineno)
                goal = goal_add_prem(goal,lf,lf.lineno)
            if sym in vocab.sorts or sym in vocab.symbols or sym in free:
                raise Redefinition(df,"redefinition of {}".format(sym))
        if __debug__: xtracer.trace("proof.functionTactic EXIT HASH canon=%s" % goal.canon())
        return [goal] + decls[1:]

    def lookup_schema(self,schemaname,decl,ast,close=False):
        if __debug__: xtracer.trace("proof.LookupSchema ENTER schemaName=%s close=%s" % (schemaname, close))
        if schemaname in self.schemata:
            schema = self.schemata[schemaname]
            if __debug__: xtracer.trace("proof.ProofChecker.LookupSchema schemata.lookup key='%s' found=true value=%s" % (schemaname, schema.canon()))
            check_schema_capture(schema,decl)
        elif schemaname in self.definitions:
            if __debug__: xtracer.trace("proof.ProofChecker.LookupSchema schemata.lookup key='%s' found=false" % schemaname)
            schema = self.definitions[schemaname]
            fmla = goal_conc(schema).to_constraint()
            fmla = il.close_formula(fmla) if close else fmla
            schema = clone_goal(schema,goal_prems(schema),fmla)
            check_schema_capture(schema,decl)
        else:
            if __debug__: xtracer.trace("proof.ProofChecker.LookupSchema schemata.lookup key='%s' found=false" % schemaname)
            premmap = dict((x.name,x) for x in goal_prem_goals(decl))
            if schemaname in premmap:
                schema = premmap[schemaname]
            else:
                if __debug__: xtracer.trace("proof.LookupSchema EXIT err=notFound schemaName=%s" % schemaname)
                raise ProofError(ast,"No property {} exists in the current context".format(schemaname))
        if __debug__: xtracer.trace("proof.LookupSchema EXIT HASH canon=%s" % schema.canon())
        return schema

    def setup_matching(self,decl,proof,allow_witness=False):
        schemaname = proof.schemaname()
        if __debug__: xtracer.trace("proof.SetupMatching ENTER schemaName=%s declLabel=%s" % (schemaname, _goal_label_trace(decl)))
        schema = self.lookup_schema(schemaname,decl,proof)
        result = self.setup_schema_matching(decl,proof,schema,allow_witness=allow_witness)
        if __debug__: xtracer.trace("proof.SetupMatching EXIT npmatch=%d" % len(result[1]))
        return result

    def setup_schema_matching(self,decl,proof,schema,allow_witness=False):
        if __debug__: xtracer.trace("proof.SetupSchemaMatching ENTER schemaLabel=%s declLabel=%s allowWitness=%s" % (_goal_label_trace(schema), _goal_label_trace(decl), allow_witness))
        if __debug__: xtracer.trace("proof.SetupSchemaMatchingRaw ENTER schemaLabel=%s declLabel=%s nmatches=%d allowWitness=%s" % (_goal_label_trace(schema), _goal_label_trace(decl), len(proof.match() or []), allow_witness))
        schema = rename_goal(schema,proof.renaming())
        schema = transform_defn_schema(schema,decl)
        prob = match_problem(schema,decl)
        prob = transform_defn_match(prob)
        if prob is None:
            if __debug__: xtracer.trace("proof.SetupSchemaMatchingRaw EXIT err=transformDefnMatchNil")
            raise NoMatch(proof,'definition does not match the given schema')
        proof_match,prob = add_prem_match(proof.match(),prob,decl,self)
        pmatch = compile_match(proof_match,prob,decl,allow_witness)
        if pmatch is None:
            if __debug__: xtracer.trace("proof.SetupSchemaMatchingRaw EXIT err=matchInconsistent nmatches=%d" % len(proof_match))
            raise ProofError(proof,'Match is inconsistent')
        if __debug__: xtracer.trace("proof.SetupSchemaMatchingRaw EXIT npmatch=%d" % len(pmatch))
        return prob, pmatch

    def assume_tactic(self,decls,proof):
        if __debug__: xtracer.trace("proof.assumeTactic ENTER ndecls=%d isGlobal=%s" % (len(decls), isinstance(proof,ia.AssumeGlobalTactic)))
        decl = decls[0]
        schemaname = proof.schemaname()
        premmap = dict((x.name,x) for x in goal_prem_goals(decl))
        if not isinstance(proof,ia.AssumeGlobalTactic) and schemaname in premmap:
            schema = premmap[schemaname]
            if isinstance(proof.label,ia.NoneAST):
                decl = goal_remove_prem(decl,schemaname)
        else:
            schema = self.lookup_schema(schemaname,decl,proof,close=False)
        schema = remove_explicit(schema)
        prob, pmatch = self.setup_schema_matching(decl,proof,schema,allow_witness=True)
        # Mirror Go's single-call isWitVar pattern (proof/tactics.go:157) —
        # call iswit once per pmatch item and partition, not twice via
        # separate comprehensions (which would double the xtrace lines).
        witness = {}
        _new_pmatch = {}
        for x,y in pmatch.items():
            if x in prob.freesyms:
                r = False
                _reason = " reason=inFreeSyms"
            else:
                r = isinstance(x, il.Variable)
                _reason = ""
            if __debug__:
                _xk = x.canon() if hasattr(x,'canon') else str(x)
                xtracer.trace("proof.isWitVar key=%s result=%s%s" % (_xk, r, _reason))
            if r:
                witness[x] = y
            else:
                _new_pmatch[x] = y
        pmatch = _new_pmatch
        if __debug__: xtracer.trace("proof.assumeTactic witnessSplit schema=%s nWitness=%d nPmatch=%d" % (schemaname, len(witness), len(pmatch)))
#        prem = make_goal(proof.lineno,fresh_label(goal_prems(decl)),[],schema)
        prem = prob.schema
        if schemaname not in premmap:
            prem = close_unmatched(prem,pmatch)
        conc = goal_conc(prem)
        conc = lu.witness_ast(True,[],witness,conc)
        if __debug__: xtracer.trace("proof.assumeTactic postWitnessAst schema=%s HASH canon=%s" % (schemaname, conc.canon() if hasattr(conc,'canon') else type(conc).__name__))
        prem = clone_goal(prem,goal_prems(prem),conc)
        prem  = apply_match_goal(pmatch,prem,apply_match_alt)
        prem = drop_supplied_prems(prem,decl,proof.match())
        if not isinstance(proof.label,ia.NoneAST):
            prem = prem.clone([proof.label,prem.formula])
        if any(prem.name == x.name for x in goal_prem_goals(decl)):
            if isinstance(proof,ia.AssumeGlobalTactic):
                prem = rename_prem_no_clash(prem,decl)
            else:
                if __debug__: xtracer.trace("proof.assumeTactic EXIT err=clash prem=%s" % prem.name)
                raise ProofError(proof,'instance name {} clashes with context'.format(prem.name))
        newGoal = goal_add_prem(decl,prem,proof.lineno)
        if __debug__: xtracer.trace("proof.assumeTactic EXIT schema=%s HASH canon=%s" % (schemaname, newGoal.canon()))
        return [newGoal] + decls[1:]

    def unfold_tactic(self,decls,proof):
        if __debug__: xtracer.trace("proof.unfoldTactic ENTER ndecls=%d nUnfSpecs=%d" % (len(decls), len(proof.unfspecs)))
        decl = decls[0]
        defns = []
        for unfspec in proof.unfspecs:
            defname = unfspec.defname
            defn = self.lookup_schema(defname,decl,proof)
            rdefs = [rename_goal(defn,rn) for rn in unfspec.renamings]
            rdefs.append(defn)
            defns.append(rdefs)
        if proof.has_premise:
            premname = proof.premname
            decl = goal_apply_to_prem(decl,premname,lambda goal: unfold_goal(goal,defns))
            if decl is None:
                if __debug__: xtracer.trace("proof.unfoldTactic EXIT err=noPremise premName=%s" % premname)
                raise ProofError(proof,'no premise {} found to unfold in'.format(premname))
        else:
            decl = goal_apply_to_conc(decl,lambda fmla: unfold_fmla(fmla,defns))
        if __debug__: xtracer.trace("proof.unfoldTactic EXIT HASH canon=%s" % decl.canon())
        return [decl] + decls[1:]

    def forget_tactic(self,decls,proof):
        if __debug__: xtracer.trace("proof.forgetTactic ENTER ndecls=%d nNames=%d" % (len(decls), len(proof.premnames)))
        decl = decls[0]
        prems = goal_prems(decl)
        prems = [p for p in prems if not (isinstance(p,ia.LabeledFormula)
                                          and p.name in proof.premnames)]
        decl = clone_goal(decl,prems,goal_conc(decl))
        if __debug__: xtracer.trace("proof.forgetTactic EXIT nkept=%d" % len(prems))
        return [decl] + decls[1:]

    def if_tactic(self,decls,proof):
        if __debug__: xtracer.trace("proof.ifTactic ENTER ndecls=%d" % len(decls))
        cond = proof.args[0]
        if __debug__: xtracer.trace("proof.WrapImplies ENTER formulaType=%s" % type(decls[0].formula).__name__)
        true_impl = il.Implies(cond,decls[0].formula)
        if __debug__: xtracer.trace("proof.WrapImplies EXIT type=lgImplies HASH canon=%s" % (true_impl.canon() if hasattr(true_impl,'canon') else str(true_impl)))
        true_goal = ia.LabeledFormula(decls[0].label,true_impl)
        true_goal.lineno = decls[0].lineno
        if __debug__: xtracer.trace("proof.WrapImplies ENTER formulaType=%s" % type(decls[0].formula).__name__)
        false_impl = il.Implies(il.Not(cond),decls[0].formula)
        if __debug__: xtracer.trace("proof.WrapImplies EXIT type=lgImplies HASH canon=%s" % (false_impl.canon() if hasattr(false_impl,'canon') else str(false_impl)))
        false_goal = ia.LabeledFormula(decls[0].label,false_impl)
        false_goal.lineno = decls[0].lineno
        result = (attrib_goals(proof.args[1],self.apply_proof([true_goal],proof.args[1])) +
                attrib_goals(proof.args[2],self.apply_proof([false_goal],proof.args[2])) +
                decls[1:])
        if __debug__: xtracer.trace("proof.ifTactic EXIT nresult=%d" % len(result))
        return result

    def match_schema(self,decl,proof):
        """ attempt to match a definition or property decl to a schema

        - decl is an ivy_ast.Definition or ivy_ast.Property
        - proof is an ivy_ast.SchemaInstantiation

        Returns a match or None
        """

        if __debug__: xtracer.trace("proof.MatchSchema ENTER goalLabel=%s HASH canon=%s" % (_goal_label_trace(decl), decl.canon()))
        if isinstance(goal_conc(decl),ia.TemporalModels):
            if __debug__: xtracer.trace("proof.MatchSchema EXIT err=temporalModels")
            raise NoMatch(proof,"goal does not match the given schema")
        prob, pmatch = self.setup_matching(decl,proof)
        apply_match_to_problem(pmatch,prob,apply_match_alt)
        if isinstance(prob.pat,ia.Tuple):
            for idx in range(len(prob.pat.args)):
                fomatch = fo_match(prob.pat.args[idx],prob.inst.args[idx],prob.freesyms,prob.constants)
                if fomatch is not None:
                    apply_match_to_problem(fomatch,prob,apply_match)
                somatch = match(prob.pat.args[idx],prob.inst.args[idx],prob.freesyms,prob.constants)
                if somatch is None:
                    if __debug__: xtracer.trace("proof.MatchSchema EXIT err=matchFailed")
                    raise NoMatch(proof,"goal does not match the given schema")
                apply_match_to_problem(somatch,prob,apply_match_alt)
        else:
            fomatch = fo_match(prob.pat,prob.inst,prob.freesyms,prob.constants)
            if fomatch is not None:
                apply_match_to_problem(fomatch,prob,apply_match)
            somatch = match(prob.pat,prob.inst,prob.freesyms,prob.constants)
            if somatch is None:
                if __debug__: xtracer.trace("proof.MatchSchema EXIT err=matchFailed")
                raise NoMatch(proof,"goal does not match the given schema")
            apply_match_to_problem(somatch,prob,apply_match_alt)
        detect_nonce_symbols(prob)
#            schema = apply_match_goal(pmatch,schema,apply_match_alt)
#            schema = apply_match_goal(fomatch,schema,apply_match)
#            schema = apply_match_goal(somatch,schema,apply_match_alt)
            # tmatch = apply_match_match(fomatch,pmatch,apply_match)
            # tmatch = apply_match_match(somatch,tmatch,apply_match_alt)
            # schema = apply_match_goal(tmatch,schema,apply_match_alt)
        result = goal_subgoals(prob.schema,decl,proof.lineno)
        if __debug__: xtracer.trace("proof.MatchSchema EXIT nsubgoals=%d" % len(result))
        return result

    def witness_tactic(self,decls,proof):
        if __debug__: xtracer.trace("proof.witnessTactic ENTER ndecls=%d nArgs=%d" % (len(decls), len(proof.args)))
        decl = decls[0]
        conc = goal_conc(decl)
        if conc is not None and hasattr(conc,'canon'):
            if __debug__: xtracer.trace("proof.witnessTactic preConc HASH canon=%s" % conc.canon())
        if ia.has_temporal(proof) and not goal_is_temporal(goal):
            if __debug__: xtracer.trace("proof.witnessTactic EXIT err=temporalInNonTemporal")
            raise iu.IvyError(proof,'temporal operator not allowed in instantiation')
        wits = compile_witness_list(proof,decls[0])
        if __debug__: xtracer.trace("proof.witnessTactic compiled nwits=%d" % len(wits))
        for wit in wits:
            if not il.is_variable(wit.args[0]):
                if __debug__: xtracer.trace("proof.witnessTactic EXIT err=lhsNotVariable")
                raise iu.IvyError(wit,'left-hand side of witness must be a variable')
        wit_map = dict((x.args[0],x.args[1]) for x in wits)
        for v, term in wit_map.items():
            if __debug__: xtracer.trace("proof.witnessTactic witnessPair key=%s rhs HASH canon=%s" % (v.canon() if hasattr(v,'canon') else str(v), term.canon() if hasattr(term,'canon') else str(term)))
        if not wit_map:
            if __debug__: xtracer.trace("proof.witnessTactic EXIT passthrough nwitness=0")
            return decls
        conc = lu.witness_ast(False,[],wit_map,conc)
        if conc is not None and hasattr(conc,'canon'):
            if __debug__: xtracer.trace("proof.witnessTactic postWitnessAst HASH canon=%s" % conc.canon())
        prems = goal_prems(decl)
        newGoal = clone_goal(decl,prems,conc)
        if __debug__: xtracer.trace("proof.witnessTactic EXIT HASH canon=%s" % newGoal.canon())
        return [newGoal] + decls[1:]
 

# A proof goal is a LabeledFormula whose body is either a Formula or a SchemaBody

def is_goal(g):
    return isinstance(g,ia.LabeledFormula)

# Get the conclusion of a goal

def goal_conc(g):
    return g.formula.conc() if isinstance(g.formula,ia.SchemaBody) else g.formula

# Get the premises of a goal

def goal_prems(g):
    return list(g.formula.prems()) if isinstance(g.formula,ia.SchemaBody) else []

# Make a goal with given label, premises (goals), conclusion (formula)

def make_goal(lineno,label,prems,conc,annot=None):
    if __debug__: xtracer.trace("proof.MakeGoal ENTER nprems=%d concType=%s" % (len(prems), type(conc).__name__))
    if isinstance(label,str):
        label = ia.Atom(label)
    res =  ia.LabeledFormula(label,ia.SchemaBody(*(prems+[conc])) if prems else conc)
    res.lineno = lineno
    if annot is not None:
        res.annot = annot
    if __debug__: xtracer.trace("proof.MakeGoal EXIT id=%d" % (res.id if hasattr(res,'id') else -1))
    return res

# Replace the premises and conclusions of a goal, keeping label and lineno
def clone_goal(goal,prems,conc):
    if __debug__: xtracer.trace("proof.CloneGoal ENTER label=%s nprems=%d concType=%s" % (_goal_label_trace(goal), len(prems), type(conc).__name__))
    result = goal.clone_with_fresh_id([goal.label,ia.SchemaBody(*(prems+[conc])) if prems else conc])
    if __debug__: xtracer.trace("proof.CloneGoal EXIT label=%s newID=%d" % (_goal_label_trace(result), result.id if hasattr(result,'id') else -1))
    return result

# Substitute a goal g2 for the conclusion of goal g1. The result has the label of g2.

def goal_subst(g1,g2,lineno):
    if __debug__: xtracer.trace("proof.GoalSubst ENTER g1Label=%s g2Label=%s" % (_goal_label_trace(g1), _goal_label_trace(g2)))
    check_name_clash(g1,g2)
    result = make_goal(lineno, g2.label, goal_prems(g1) + goal_prems(g2), goal_conc(g2))
    if __debug__: xtracer.trace("proof.GoalSubst EXIT HASH canon=%s" % result.canon())
    return result

# Substitute a sequence of subgoals in to the conclusion of the first goal

def goals_subst(goals,subgoals,lineno):
    return [goal_subst(goals[0],g,lineno) for g in subgoals] + goals[1:]

# Add a formula or schema as a premise to a goal. Make up a fresh name for it.

# Make a fresh label not used in any of a list of goals

def fresh_label(goals):
    rn = iu.UniqueRenamer(used=[x.name for x in goals])
    return ia.Atom(rn(),[])
    
# Add a premise to a goal

def goal_add_prem(goal,prem,lineno):
    if __debug__: xtracer.trace("proof.GoalAddPrem ENTER goalLabel=%s premType=%s" % (_goal_label_trace(goal), type(prem).__name__))
    result = make_goal(lineno,goal.label,goal_prems(goal) + [prem], goal_conc(goal))
    if __debug__: xtracer.trace("proof.GoalAddPrem EXIT HASH canon=%s" % result.canon())
    return result


def goal_remove_prem(goal,prem_name):
    if __debug__: xtracer.trace("proof.GoalRemovePrem ENTER goalLabel=%s premName=%s" % (_goal_label_trace(goal), prem_name))
    new_prems = [x for x in goal_prems(goal) if x.name != prem_name]
    goal = clone_goal(goal,new_prems,goal_conc(goal))
    if __debug__: xtracer.trace("proof.GoalRemovePrem EXIT nprems=%d" % len(new_prems))
    return goal

# Add a premise to a goal

def goal_prefix_prems(goal,prems,lineno):
    return make_goal(lineno,goal.label,prems + goal_prems(goal), goal_conc(goal))

# Get the symbols and types defined in the premises of a goal

def goal_defns(goal):
    res = set()
    for x in goal_prems(goal):
        if isinstance(x,ia.ConstantDecl) and isinstance(x.args[0],il.Symbol):
            res.add(x.args[0])
        elif isinstance(x,il.UninterpretedSort):
            res.add(x)
    return res

# Get all the premises of a goal that are goals

def goal_prem_goals(goal):
    return [x for x in goal_prems(goal) if isinstance(x,ia.LabeledFormula)]

# Check that there are no name clashes in a pair of goals

def check_name_clash(g1,g2):
    if __debug__: xtracer.trace("proof.CheckNameClash ENTER g1Label=%s g2Label=%s" % (g1.label, g2.label))
    d1,d2 = list(map(goal_defns,(g1,g2)))
    for s1 in d1:
        if s1 in d2:
            if __debug__: xtracer.trace("proof.CheckNameClash EXIT err=clash key=%s" % s1)
            raise ProofError(None,'premise {} of sugboal clashes with context'.format(s1))
    if __debug__: xtracer.trace("proof.CheckNameClash EXIT ok")

# A *vocabulary* consists of three lists: sorts, symbols and variables

class Vocab(object):
    def __init__(self,sorts,symbols,variables):
        self.sorts,self.symbols,self.variables = sorts,symbols,variables

# Get the vocabulary of a goal. This is the collection of sorts, symbols and
# variables that are bound in the goal. If optional argument 'bound' is true,
# include the bound variables in the conclusion.

def goal_vocab(goal,bound=False):
    if bound:
        if __debug__: xtracer.trace("proof.GoalVocabBound ENTER label=%s" % _goal_label_trace(goal))
    else:
        if __debug__: xtracer.trace("proof.GoalVocab ENTER label=%s" % _goal_label_trace(goal))
    prems = goal_prems(goal)
    conc = goal_conc(goal)
    symbols = [x.args[0] for x in prems if isinstance(x,ia.ConstantDecl)]
    sorts = [s for s in prems if isinstance(s,il.UninterpretedSort)]
    fmlas = [x.formula for x in prems if isinstance(x,ia.LabeledFormula)] + [conc]
    variables = list(lu.used_variables_asts(fmlas))
    if bound:
        conc_fmla = conc.fmla if isinstance(conc,ia.TemporalModels) else conc
        variables = variables + [x for x in logic_util.bound_variables(conc_fmla)
                                 if x not in variables]
        if __debug__: xtracer.trace("proof.GoalVocabBound EXIT nvariables=%d" % len(variables))
    else:
        if __debug__: xtracer.trace("proof.GoalVocab EXIT nsorts=%d nsymbols=%d nvariables=%d" % (len(sorts), len(symbols), len(variables)))
    return Vocab(sorts,symbols,variables)

# Check that the conclusions of two goals match

def check_concs_match(g1,g2):
    c1,c2 = list(map(goal_conc,(g1,g2)))
    if not il.equal_mod_alpha(c1,c2):
        raise ProofError(None,'conclusions do not match:\n    {}\n     {}'.format(c1,c2))

# Check that the non-proposition premises of g1 are provided by g2.

def check_premises_provided(g1,g2):
    defns = goal_defns(g2)
    for thing in goal_defns(g1):
#        syms = lu.used_symbols_ast(thing) if il.is_lambda(thing) else [thing]
        syms = [] if il.is_lambda(thing) else [thing]
        for sym in syms:
            if sym not in defns and not il.sig.contains(sym):
                raise ProofError(None,'premise "{}" does not match anything in the environment'.format(thing))

def goal_is_temporal(x):
    conc = goal_conc(x)
    return conc.temoral or isinstance(conc.formula,ia.TemporalModels)

def goal_is_defn(x):
    if isinstance(x,ia.ConstantDecl):
        return not il.is_lambda(x.args[0])
    return isinstance(x,il.UninterpretedSort)

def goal_defines(x):
    if isinstance(x,ia.ConstantDecl):
        return x.args[0]
    return x

def normalize_goal(x):
    """ normalize the subformulas of a goal, so there are only binary
    conjunctions/disjunctions and single-variable quantifiers. """
    if __debug__: xtracer.trace("proof.NormalizeGoal ENTER label=%s" % _goal_label_trace(x))
    if goal_is_defn(x):
        if __debug__: xtracer.trace("proof.NormalizeGoal EXIT passthrough=isDefn")
        return x
    if not hasattr(x,'formula'):
        print(x)
        print(type(x))
    result = clone_goal(x,list(map(normalize_goal,goal_prems(x))),il.normalize_ops(goal_conc(x)))
    if __debug__: xtracer.trace("proof.NormalizeGoal EXIT HASH canon=%s" % result.canon())
    return result

def get_unprovided_defns(g1,g2):
    defns = goal_defns(g2)
    free = goal_free(g2)
    res = []
    for prem in goal_prems(g1):
        if goal_is_defn(prem):
            sym = goal_defines(prem)
            # if sym not in defns and not il.sig.contains(sym):
            if sym not in defns and sym not in free:
                res.append(prem)
    return res

# Turn the propositional premises of a goal into a list of subgoals. The
# symbols and types in the goal must be provided by the environment.

def goal_subgoals(schema,goal,lineno):
    if __debug__: xtracer.trace("proof.GoalSubgoals ENTER schemaLabel=%s goalLabel=%s" % (_goal_label_trace(schema), _goal_label_trace(goal)))
    check_concs_match(schema,goal)
    upds = get_unprovided_defns(schema,goal)
    g = clone_goal(goal,upds,goal_conc(goal))
    goal = goal_subst(goal,g,lineno)
    gpms = goal_prem_goals(goal)
    subgoals = [goal_subst(goal,x,lineno) for x in goal_prem_goals(schema)
                if not any(goals_eq_mod_alpha(x,y) for y in gpms)]
    subgoals = [s for s in subgoals if not trivial_goal(s)]
    if __debug__: xtracer.trace("proof.GoalSubgoals EXIT nsubgoals=%d" % len(subgoals))
    return subgoals

def fmla_vocab(fmla):
    """ Get the free vocabulary of a formula, including sorts, symbols and variables """
    
    things = dict(lu.used_sorts_ast(fmla))
    things.update(lu.used_symbols_ast(fmla))
    things.update(lu.used_variables_ast(fmla))
    return things


def goal_free(goal):
    """ Get the free vocabulary of a goal, including sorts, symbols and variables """
    if __debug__: xtracer.trace("proof.GoalFree ENTER label=%s" % _goal_label_trace(goal))
    bound = set()
    def rec_fmla(fmla,res):
        for y in fmla_vocab(fmla):
            if y not in bound:
                res.add(y)
    def rec(goal,res):
        defns = goal_defns(goal)
        with il.BindSymbols(bound,defns):
            for x in goal_prem_goals(goal):
                if isinstance(x.formula,ia.SchemaBody):
                    rec(x,res)
                else:
                    rec_fmla(x.formula,res)
            rec_fmla(goal_conc(goal),res)
    res = set()
    rec(goal,res)
    if __debug__:
        _gf_items = sorted([str(x) for x in res])
        xtracer.trace("proof.GoalFree items=[%s]" % ",".join(_gf_items))
        xtracer.trace("proof.GoalFree EXIT nfree=%d" % len(res))
    return res

# Make sure the free vocabulry of the schema we are about to use is not captured
# by bindings in the goal. 

def check_schema_capture(schema,goal):
    gvocab = goal_vocab(goal)
    fvocab = goal_free(schema)
    for sym in fvocab:
        if sym in gvocab.sorts or sym in gvocab.symbols:
            raise CaptureError(None,'"{}" is captured when importing "{}"'.format(sym,schema.name))

def check_alpha_capture(goal,match):
    rev_match = dict((y,x) for x,y in match.items())
    for s in goal_free(goal).union(goal_defns(goal)):
        if s in rev_match and s not in match:
            raise CaptureError(None,'"{}" is captured by renaming "{}"'.format(s,rev_match[s]))

def check_renaming(goal,renaming):
    fwd = dict()
    rev = dict()
    for x in renaming.args:
        l,r = x.lhs().rep, x.rhs().rep
        if l in fwd:
            raise ProofError(None,'"{}" is renamed to both "{}" and "{}"'.format(l,fwd[l],r))
        if r in rev:
            raise ProofError(None,'both "{}" and "{}" are renamed to "{}"'.format(rev[r],l,r))
        fwd[l] = r
        fwd[r] = l
    return
        
def rename_goal(goal,renaming):
    if len(renaming.args) == 0:
        return goal
    check_renaming(goal,renaming)
    rmap = dict((x.lhs().rep,x.rhs().rep) for x in renaming.args)
    def rec_goal(goal):
        if not isinstance(goal,ia.LabeledFormula):
            return goal
        goal = clone_goal(goal,list(map(rec_goal,goal_prems(goal))),goal_conc(goal))
        match = dict((x,x.rename(lambda n: rmap[x.name])) for x in goal_defns(goal) if x.name in rmap)
        match = dict((x,apply_match_sym(match,y)) for x,y in match.items())
        check_alpha_capture(goal,match)
        goal = apply_match_goal(match,goal,apply_match_alt)
        goal = clone_goal(goal,goal_prems(goal),il.alpha_rename(rmap,goal_conc(goal)))
        goal = goal.rename(rmap.get(goal.name,goal.name))
        return goal
    res = rec_goal(goal)
    return res
                
            
            

# Compile an expression using a vocabulary. The expression could be a formula or a type.

def compile_expr_vocab(expr,vocab):
    if __debug__: xtracer.trace("proof.CompileExprVocab ENTER exprType=%s" % type(expr).__name__)
    with il.WithSymbols(vocab.symbols):
        with il.WithSorts(vocab.sorts):
            if isinstance(expr,ia.Atom) and expr.rep in il.sig.sorts:
                result = il.sig.sorts[expr.rep]
                if __debug__: xtracer.trace("proof.CompileExprVocab EXIT viaSort")
                return result
            with il.top_sort_as_default():
                with ia.ASTContext(expr):
                    expr = il.sort_infer_list([expr.compile()] + vocab.variables)[0]
                    if __debug__: xtracer.trace("proof.CompileExprVocab EXIT HASH canon=%s" % (expr.canon() if hasattr(expr,'canon') else str(expr)))
                    return expr


# Compile an expression using a vocabulary. The expression could be a formula or a type.

def compile_expr_vocab_ext(expr,vocab):
    with il.WithSymbols(vocab.symbols):
        with il.WithSorts(vocab.sorts):
            if isinstance(expr,ia.Atom) and expr.rep in il.sig.sorts:
                return il.sig.sorts[expr.rep]
            with il.top_sort_as_default():
                with ia.ASTContext(expr):
                    expr = expr.compile()
                    return expr


def remove_vars_match(mat,fmla):
    """ Remove the variables bindings from a match. This is used to
    prevent variable capture when applying the match to premises. Make sure free variables
    are not captured by fmla """
    res = dict((s,v) for s,v in mat.items() if il.is_ui_sort(s))
    sympairs = [(s,v) for s,v in mat.items() if il.is_constant(s)]
    symfmlas = il.rename_vars_no_clash([v for s,v in sympairs],[fmla])
    res.update((s,w) for (s,v),w in zip(sympairs,symfmlas))
    return res


def show_match(m):
    if m is None:
        print('no match')
        return 
    print('match {')
    for x,y in m.items():
        print('{} : {} |-> {}'.format(x,x.sort if hasattr(x,'sort') else 'type',y))
    print('}')
        
def match_problem(schema,decl):
    """ Creating a matching problem from a schema and a declaration """
    vocab = goal_vocab(schema)
    freesyms = set(vocab.symbols + vocab.sorts + vocab.variables)
    constants = set(v for v in goal_free(decl) if il.is_variable(v))
    return MatchProblem(schema,goal_conc(schema),goal_conc(decl),freesyms,constants)

def transform_defn_schema(schema,decl):
    """ Transform definition schema to match a definition. """
    conc = goal_conc(schema)
    decl = goal_conc(decl)
    if not(isinstance(decl,il.Definition) and isinstance(conc,il.Definition)):
        return schema
    declargs = decl.lhs().args
    concargs = conc.lhs().args
    if len(declargs) > len(concargs):
        schema = parameterize_schema([x.sort for x in declargs[:len(declargs)-len(concargs)]],schema)
    return schema

def transform_defn_match(prob):
    """ Transform a problem of matching definitions to a problem of
    matching the right-hand sides. Requires prob.inst is a definition. """

    schema, conc,decl,freesyms = prob.schema, prob.pat,prob.inst,prob.freesyms
    if not(isinstance(decl,il.Definition) and isinstance(conc,il.Definition)):
        return prob
    declsym = decl.defines()
    concsym = conc.defines()
    # dmatch = match(conc.lhs(),decl.lhs(),freesyms)
    # if dmatch is None:
    #     print "left-hand sides didn't match: {}, {}".format(conc.lhs(),decl.lhs())
    #     return None
    declargs = decl.lhs().args
    concargs = conc.lhs().args
    if len(declargs) < len(concargs):
        return None
    declrhs = decl.rhs()
    concrhs = conc.rhs()
    vmap = dict((x.name,y.resort(x.sort)) for x,y in zip(concargs,declargs))
    concrhs = lu.substitute_ast(concrhs,vmap)
    dmatch = {concsym:declsym}
    for x,y in zip(func_sorts(concsym),func_sorts(declsym)):
        if x in freesyms:
            if x in dmatch and dmatch[x] != y:
                print("lhs sorts didn't match: {}, {}".format(x,y))
                return None
            dmatch[x] = y
        else:
            if x != y:
                print("lhs sorts didn't match: {}, {}".format(x,y))
                return None
    concrhs = apply_match(dmatch,concrhs)
    freesyms = apply_match_freesyms(dmatch,freesyms)
    freesyms = [x for x in freesyms if x not in concargs]
    constants = set(x for x in prob.constants if x not in declargs)
    vvmap = dict((x,y.resort(x.sort)) for x,y in zip(concargs,declargs))
    schema = apply_match_goal(vvmap,schema,apply_match_alt)
    schema = apply_match_goal(dmatch,schema,apply_match_alt)
    return MatchProblem(schema,concrhs,declrhs,freesyms,constants)

def goal_prems_by_name(goal):
    gprems = goal_prem_goals(goal)
    return dict((p.name,p) for p in gprems)

def add_prem_match(proof_match,prob,goal,context):
    sprems = goal_prems_by_name(prob.schema)
    pats = []
    insts = []
    new_match = []
    for m in proof_match:
        lhs,rhs = m.args
        if isinstance(lhs,ia.Atom) and len(lhs.args) == 0:
            sprem = sprems.get(lhs.rep,None)
            if sprem is not None:
                if isinstance(rhs,ia.Atom) and len(rhs.args) == 0:
                    gprem = context.lookup_schema(rhs.rep,goal,rhs)
                    pats.append(sprem)
                    insts.append(gprem)
                    continue
        new_match.append(m)
    if pats:
        pat = ia.Tuple(*(pats + [prob.pat]))
        inst = ia.Tuple(*(insts + [prob.inst]))
        prob = MatchProblem(prob.schema,pat,inst,prob.freesyms,prob.constants,pats)
    return new_match,prob

def parameterize_schema(sorts,schema):
    """ Add initial parameters to all the free symbols in a schema.

    Takes a list of sorts and an ia.SchemaBody. """

    vars = make_distinct_vars(sorts,goal_conc(schema))
    match = {}
    prems = []
    for prem in goal_prems(schema):
        if isinstance(prem,ia.ConstantDecl):
            sym = prem.args[0]
            vs2 = [il.Variable('X'+str(i),y) for i,y in enumerate(sym.sort.dom)]
            sym2 = sym.resort(il.FuncConstSort(*(sorts + list(sym.sort.dom) + [sym.sort.rng])))
            match[sym] = il.Lambda(vs2,sym2(*(vars+vs2)))
            prems.append(ia.ConstantDecl(sym2))
        else:
            prems.append(prem)
    conc = apply_match(match,goal_conc(schema))
    return clone_goal(schema,prems,conc)

# A schema instantiataion has an associated list of mathces
# (following 'with').  When compiling this, the left-hand sides
# use names of constants and variables from the shema being
# instantiated, while the right-hand sides uses names from the
# current goal (and both may use names from the globla context

def compile_match_list(proof_match,left_goal,right_goal,allow_witness=False):
    def compile_match(d):
        x,y = d.lhs(),d.rhs()
        x = compile_expr_vocab(x,left_goal_vocab)
        y = compile_expr_vocab(y,right_goal_vocab)
        return ia.Definition(x,y)
    left_goal_vocab = goal_vocab(left_goal)
    right_goal_vocab = goal_vocab(right_goal)
    if allow_witness:
        left_goal_vocab.variables.extend(list(logic_util.used_variables(goal_conc(left_goal))))
    return [compile_match(d) for d in proof_match]

# A "match" is a map from symbols to lambda terms
    
def compile_one_match(lhs,rhs,freesyms,constants):
    if il.is_variable(lhs):
        return fo_match(lhs,rhs,freesyms,constants)
    if not isinstance(rhs,il.UninterpretedSort):
        rhsvs = dict((v.name,v) for v in lu.used_variables_ast(rhs))
        vmatches = [{v.sort:rhsvs[v.name].sort} for v in lu.used_variables_ast(lhs)
                    if v.name in rhsvs and v.sort in freesyms]
        vmatch = merge_matches(*vmatches)
        if vmatch is None:
            return None
        lhs = apply_match_alt(vmatch,lhs)
        newfreesyms = apply_match_freesyms(vmatch,freesyms)
        somatch = match(lhs,rhs,newfreesyms,constants)
        if somatch is None:
            return None
        somatch = compose_matches(freesyms,vmatch,somatch,vmatch)
        fmatch = merge_matches(vmatch,somatch)
        return fmatch
    else:
        return match_sort(lhs,rhs,freesyms)


def compile_match(proof_match,prob,decl,allow_witness=False):
    """ Compiles match in a proof. Only the symbols in
    freesyms may be used in the match."""

    if __debug__: xtracer.trace("proof.CompileMatchFull ENTER nProofMatch=%d declLabel=%s allowWitness=%s" % (len(proof_match), _goal_label_trace(decl), allow_witness))
    schema = prob.schema
    freesyms = prob.freesyms.copy()
    if allow_witness:
        freesyms.update(logic_util.used_variables(goal_conc(schema)))
    matches = compile_match_list(proof_match,schema,decl,allow_witness=allow_witness)
    matches = [compile_one_match(m.lhs(),m.rhs(),freesyms,prob.constants) for m in matches]
    res = merge_matches(*matches)
    if __debug__: xtracer.trace("proof.CompileMatchFull EXIT nmatches=%d nresult=%d" % (len(matches), len(res) if res else 0))
    return res
        
        
    res = dict()
    for m in proof_match:
        if il.is_app(m.lhs()):
            res[m.defines()] = il.Lambda(m.lhs().args,m.rhs())
        else:
            res[m.lhs()] = m.rhs()
    # iu.dbg('freesyms')
    # freesyms = apply_match_freesyms(res,freesyms)
    # iu.dbg('freesyms')
    # for sym in res:
    #     if sym not in freesyms:
    #         raise ProofError(proof,'{} is not a premise of schema {}'.format(repr(sym),schemaname))
    return res

def match_rhs_vars(match):
    """ Get the symbols occurring free on the right-hand side of a match """
    res = set()
    for w in list(match.values()):
        for v in w if isinstance(w,list) else [w]:
            if isinstance(v,(il.UninterpretedSort,il.EnumeratedSort)):
                res.add(v)
            else:
                res.update(fmla_vocab(v))
    return res

def is_lambda(p):
    return isinstance(p,ia.ConstantDecl) and isinstance(p.args[0],il.Lambda)

def goal_is_property(x):
    return isinstance(x,ia.LabeledFormula) and not isinstance(x.formula,ia.SchemaBody)

def goal_is_schema(x):
    return isinstance(x,ia.LabeledFormula) and isinstance(x.formula,ia.SchemaBody)

def apply_match_goal(match,x,apply_match,env = None):
    """ Apply a match to a goal """
    is_top = env is None
    if is_top:
        if __debug__: xtracer.trace("proof.ApplyMatchGoalNode ENTER label=%s nmatch=%d" % (_goal_label_trace(x), len(match)))
    env = env if env is not None else set()
    if isinstance(x,ia.LabeledFormula):
        fmla = x.formula
        if isinstance(fmla,ia.SchemaBody):
            bound = [s for s in goal_defns(x) if s not in match]
            with il.BindSymbols(env,bound):
                prems = [apply_match_goal(match,y,apply_match,env) for y in fmla.prems()]
                prems = [p for p in prems if not is_lambda(p)]
                fmla = fmla.clone(prems+[apply_match(match,fmla.conc(),env)])
        else:
            fmla = apply_match(match,fmla,env)
        # Mirror Go's CloneGoalPreserveID ENTER/EXIT wrapper (proof/goal.go:209).
        if __debug__:
            if isinstance(fmla, ia.SchemaBody):
                _cgp_nprems = len(fmla.prems())
                _cgp_conc = fmla.conc()
                _cgp_conc_type = type(_cgp_conc).__name__
            else:
                _cgp_nprems = 0
                _cgp_conc_type = type(fmla).__name__
            xtracer.trace("proof.CloneGoalPreserveID ENTER label=%s nprems=%d concType=%s id=%d" % (
                _goal_label_trace(x),
                _cgp_nprems, _cgp_conc_type,
                x.id if hasattr(x,'id') else -1))
        g = x.clone([x.label,fmla])
        if __debug__: xtracer.trace("proof.CloneGoalPreserveID EXIT label=%s" % _goal_label_trace(g))
        if is_top:
            if __debug__: xtracer.trace("proof.ApplyMatchGoalNode EXIT HASH canon=%s" % g.canon())
        return g
    if isinstance(x,(il.UninterpretedSort,il.EnumeratedSort)):
        return apply_match_sort(match,x)
    else:
        return x.clone([apply_match_func_alt(match,x.args[0],env)])

def apply_match_match(match,orig_match,apply_match):
    """ Apply a match match to match orig_match. Applying the resulting match should
    have the same effect as apply first orig_match, then match. """
    orig_match = dict((x,apply_match(match,y)) for x,y in orig_match.items())
    orig_match.update((x,y) for x,y in match.items() if x not in orig_match)
    return orig_match

def apply_match_to_problem(match,prob,apply_match):
    avoid_capture_problem(prob,match)
    prob.schema = apply_match_goal(match,prob.schema,apply_match)
    prob.pat = apply_match(match,prob.pat)
    prob.freesyms = apply_match_freesyms(match,prob.freesyms)
    prob.revmap = dict((x,y) for x,y in prob.revmap.items() if x not in match)

def rename_problem(match,prob):
    prob.schema = apply_match_goal(match,prob.schema,apply_match_alt)
    prob.pat = apply_match_alt(match,prob.pat)
    prob.freesyms = set(match.get(sym,sym) for sym in prob.freesyms)
    prob.revmap.update((y,x) for x,y in match.items())

def avoid_capture_problem(prob,match):
    """ Rename a match problem to avoid capture when applying a
    match"""
    mrv = match_rhs_vars(match)
    matchnames = set(x.name for x in match_rhs_vars(match))
    used = set(matchnames)
    used.update(x.name for x in goal_defns(prob.schema))
    used.update(v.name for v in goal_free(prob.schema) if il.is_variable(v))
    rn = iu.UniqueRenamer(used=used)
    cmatch = dict((v,v.rename(rn)) for v in prob.freesyms
                  if v.name in matchnames and v not in match)
    rename_problem(cmatch,prob)

def detect_nonce_symbols(prob):
    """ Make sure that no nonce symbols produced by
    avoid_capture_problem appear free after matching. This is done to
    avoid nonce symbols becoming visible to the user. If one of these
    remains after matching, we report the original symbol
    as clashing with the corresponding symbol in the goal."""

    for sym in list(prob.revmap.values()):
        raise CaptureError(None,'Symbol {} in schema clashes with {} in goal.\nSuggest renaming or instantiating it.'.format(sym,sym))

def trivial_goal(goal):
    """ A goal is trivial if the conclusion is equal to one of the premises modulo
    alpha conversion """
    conc = goal_conc(goal)
    for prem in goal_prem_goals(goal):
        if len(goal_prems(prem)) == 0:
            if il.equal_mod_alpha(goal_conc(prem),conc):
                return True
    return False

# Test whether two goals are equivalent mod alpha renaming
# TODO: for now just tests syntactic equality

def goals_eq_mod_alpha(x,y):
    if isinstance(x.formula,ia.SchemaBody):
        if not isinstance(y.formula,ia.SchemaBody):
            return False
        xps, yps = goal_prems(x), goal_prems(y)
        if len(xps) != len(yps):
            return False
        for xp,yp in zip(xps,yps):
            if type(xp) is not type(yp):
                return False
            if isinstance(xp,ia.LabeledFormula):
                if not goals_eq_mod_alpha(xp,yp):
                    return False
            elif isinstance(xp,ia.ConstantDecl):
                if xp.args[0] != yp.args[0]:
                    return False
            elif isinstance(xp,il.UninterpretedSort):
                if xp != yp:
                    return False
    else:
        if isinstance(y.formula,ia.SchemaBody):
            return False
    return il.equal_mod_alpha(goal_conc(x),goal_conc(y))

def apply_match(match,fmla,env = None):
    """ apply a match to a formula. 

    In effect, substitute all symbols in the match with the
    corresponding lambda terms and apply beta reduction

    Have to first alpha-rename to avoid capture of variables by binders

    """
    if __debug__: xtracer.trace("proof.ApplyMatch ENTER nmatch=%d fmlaType=%s" % (len(match), type(fmla).__name__))
    freevars = match_rhs_vars(match)
    fmla = il.alpha_avoid(fmla,freevars)
    result = apply_match_rec(match,fmla,env if env is not None else set())
    if __debug__: xtracer.trace("proof.ApplyMatch EXIT HASH canon=%s" % (result.canon() if hasattr(result,'canon') else str(result)))
    return result

def apply_match_rec(match,fmla,env):
    args = [apply_match_rec(match,f,env) for f in fmla.args]
    if il.is_app(fmla):
        if fmla.rep in match:
            func = match[fmla.rep]
            return func(*args)
        return apply_match_func(match,fmla.rep)(*args)
    if il.is_variable(fmla) and fmla in match:
        return match[fmla]
    if il.is_binder(fmla):
        with il.BindSymbols(env,fmla.variables):
            fmla = fmla.clone_binder([apply_match_rec(match,v,env) for v in fmla.variables],args[0])
        return fmla
    return fmla.clone(args)

def raise_capture(v):
    raise CaptureError(None,'symbol {} is captured in substitution'.format(v))

def match_get(match,sym,env,default=None):
    """ get the value of a symbol in a match, checking that no symbols
    are captured in env """
    val = match.get(sym,None)
    if isinstance(val,list):  # for unfolding only, may get a list of values
        save = val
        val = val[0]
        if len(save) > 1:
            del save[0]
    if val is not None:
        vocab = lu.used_symbols_ast(val)
        vocab.update(lu.variables_ast(val))
        for v in vocab:
            if v in env:
                raise_capture(v)
        return val
    return default
    

def apply_match_alt(match,fmla,env = None):
    """ apply a match to a formula.

    In effect, substitute all symbols in the match with the
    corresponding lambda terms and apply beta reduction

    If present, env is list of symbols bound in the environment.
    Substituting one of these symbols into the formula will be considered
    capture and cause CaptureError to be raised.

    """
    if __debug__: xtracer.trace("proof.ApplyMatchAlt ENTER nmatch=%d fmlaType=%s" % (len(match), type(fmla).__name__))
    freevars = list(match_rhs_vars(match))
    fmla = il.alpha_avoid(fmla,freevars)
    result = apply_match_alt_rec(match,fmla,env if env is not None else set())
    if result is not None and hasattr(result,'canon'):
        if __debug__: xtracer.trace("proof.ApplyMatchAlt EXIT HASH canon=%s" % result.canon())
    else:
        if __debug__: xtracer.trace("proof.ApplyMatchAlt EXIT resultType=%s" % type(result).__name__)
    return result


def apply_fun(fun,args):
    try:
        return fun(*args)
    except il.CaptureError as err:
        for sym in err.variables:
            raise_capture(sym)

def apply_match_alt_rec(match,fmla,env):
    args = [apply_match_alt_rec(match,f,env) for f in fmla.args]
    if il.is_app(fmla):
        if fmla.rep in match:
            return apply_fun(match_get(match,fmla.rep,env),args)
        func = apply_match_func(match,fmla.rep)
        func = match_get(match,func,env,func)
        return func(*args)
    if il.is_variable(fmla):
        if fmla in match:
            return match_get(match,fmla,env)
        fmla = il.Variable(fmla.name,apply_match_sort(match,fmla.sort))
        fmla = match_get(match,fmla,env,fmla)
        return fmla
    if il.is_binder(fmla):
        with il.BindSymbols(env,fmla.variables):
            fmla = fmla.clone_binder([apply_match_alt_rec(match,v,env) for v in fmla.variables],args[0])
        return fmla
    return fmla.clone(args)

def apply_match_func(match,func):
    sorts = func_sorts(func)
    sorts = [match.get(s,s) for s in sorts]
    return il.Symbol(func.name,sorts[0] if len(sorts) == 1 else il.FunctionSort(*sorts))

def apply_match_func_alt(match,func,env):
    if il.is_lambda(func):
        return apply_match_alt(match,func,env)
    if func in match:
        return match[func]
    func = apply_match_func(match,func)
    return match.get(func,func)

def apply_match_sym(match,sym):
    if il.is_variable(sym):
        return il.Variable(sym.name,match.get(sym.sort,sym.sort))
    return match.get(sym,sym) if isinstance(sym,il.UninterpretedSort) else apply_match_func(match,sym)

def apply_match_sort(match,sort):
    return match.get(sort,sort)

def apply_match_freesyms(match,freesyms):
    return set(apply_match_sym(match,sym) for sym in freesyms if sym not in match)

def apply_match_freesyms_alt(match,freesyms):
    msyms = [apply_match_sym(match,sym) for sym in freesyms]
    return [sym for sym in msyms if sym not in match]

def func_sorts(func):
    return list(func.sort.dom) + [func.sort.rng]

def lambda_sorts(lmbd):
    return [v.sort for v in lmbd.variables] + [lmbd.body.sort]

def term_sorts(term):
    """ Returns a list of the domain and range sorts of the head function of a term, if any """
    return func_sorts(term.rep) if il.is_app(term) else [term.sort] if il.is_variable(term) else []

def funcs_match(pat,inst,freesyms):
    psorts,isorts = list(map(func_sorts,(pat,inst)))
    res = (pat.name == inst.name and len(psorts) == len(isorts)
            and all(x == y for x,y in zip(psorts,isorts) if x not in freesyms))
    return res
    
def heads_match(pat,inst,freesyms):
    """Returns true if the heads of two terms match. This means they have
    the same top-level operator and same number of
    arguments. Quantifiers do not match anything. A function symbol matches
    if it has the same name and if it agrees on the non-free sorts in
    its type.
    """
    return (il.is_app(pat) and il.is_app(inst) and funcs_match(pat.rep,inst.rep,freesyms) and pat.rep not in freesyms
        or not il.is_app(pat) and not il.is_quantifier(pat)
           and type(pat) is type(inst) and len(pat.args) == len(inst.args))
    
def make_distinct_vars(sorts,*asts):
    vars = [il.Variable('V'+str(i),sort) for i,sort in enumerate(sorts)]
    return lu.rename_variables_distinct_asts(vars,asts)
    

def extract_terms(inst,terms):
    """ Returns a lambda term t such that t(terms) = inst and
    terms do not occur in t. vars is a list of distinct variables
    of same types as terms that are not free in inst. """

    vars = make_distinct_vars([t.sort for t in terms], inst)
    def rec(inst):
        for term,var in zip(terms,vars):
            if term == inst:
                return var
        return inst.clone(list(map(rec,inst.args)))
    return il.Lambda(vars,rec(inst))

def fo_match(pat,inst,freesyms,constants):
    """ Compute a partial first-order match. Matches free FO variables to ground terms,
    but ignores variable occurrences under free second-order symbols. """

    if __debug__: xtracer.trace("proof.FOMatch ENTER patType=%s instType=%s" % (type(pat).__name__, type(inst).__name__))
    if il.is_variable(pat):
        if pat in freesyms and all(x in constants for x in lu.variables_ast(inst)):
            res = {pat:inst}
            if pat.sort in freesyms:
                res[pat.sort] = inst.sort
                return res
            if pat.sort == inst.sort:
                return res
    if il.is_quantifier(pat) and il.is_quantifier(inst):
        with RemoveSymbols(freesyms,pat.variables):
            return fo_match(pat.body,inst.body,freesyms,constants)
    if heads_match(pat,inst,freesyms):
        matches = [fo_match(x,y,freesyms,constants) for x,y in zip(pat.args,inst.args)]
        res =  merge_matches(*matches)
        return res
    return dict()
    
            

def match(pat,inst,freesyms,constants):
    """ Match an instance to a pattern.

    A match is an assignment sigma to freesyms such
    that sigma pat =_alpha inst.

    """

    if __debug__: xtracer.trace("proof.Match ENTER patType=%s instType=%s nfreesyms=%d" % (type(pat).__name__, type(inst).__name__, len(freesyms)))
    if il.is_quantifier(pat):
        result = match_quants(pat,inst,freesyms,constants)
        if __debug__: xtracer.trace("proof.Match EXIT viaQuants success=%s" % (result is not None))
        return result
    if heads_match(pat,inst,freesyms):
        matches = [match(x,y,freesyms,constants) for x,y in zip(pat.args,inst.args)]
        matches.extend([match_sort(x,y,freesyms) for x,y in zip(term_sorts(pat),term_sorts(inst))])
        if il.is_variable(pat):
            matches.append({pat:inst})
        res = merge_matches(*matches)
        if __debug__: xtracer.trace("proof.Match EXIT viaHeads success=%s" % (res is not None))
        return res
    elif il.is_app(pat) and pat.rep in freesyms:
        B = extract_terms(inst,pat.args)
        if all(v in constants for v in lu.variables_ast(B)):
            matches = [{pat.rep:B}]
            matches.extend([match_sort(x,y,freesyms) for x,y in zip(term_sorts(pat),lambda_sorts(B))])
            res = merge_matches(*matches)
            if __debug__: xtracer.trace("proof.Match EXIT viaLambda success=%s" % (res is not None))
            return res
    if __debug__: xtracer.trace("proof.Match EXIT noMatch")
        


def match_quants(pat,inst,freesyms,constants):
    """ Match an instance to a pattern that is a quantifier.
    """

    if type(pat) is not type(inst) or len(pat.variables) != len(inst.variables):
        return None
    with AddSymbols(freesyms,pat.variables):
        matches = [match(x,y,freesyms,constants) for x,y in zip(pat.variables,inst.variables)]
        mat = merge_matches(*matches)
        if mat is not None:
            mbody = apply_match(mat,pat.body)
            bodyfreesyms = apply_match_freesyms(mat,freesyms)
            bodymat = match(mbody,inst.body,bodyfreesyms,constants)
            bodymat = compose_matches(freesyms,mat,bodymat,pat.variables)
            mat = merge_matches(mat,bodymat)
#        matches.append(match(pat.body,inst.body,freesyms,constants))
#        mat = merge_matches(*matches)
        if mat is not None:
            for x in pat.variables:
                if x in mat:
                    del mat[x]
        return mat

def compose_matches(freesyms,mat1,mat2,quants):
    if mat1 is None or mat2 is None:
        return None
    res = dict()
    for sym in freesyms:
        if sym not in quants:
            sym1 = apply_match_sym(mat1,sym)
            if sym1 in mat2:
                res[sym] = mat2[sym1]
    return res

def match_sort(pat,inst,freesyms):
    if pat in freesyms:
        return {pat:inst}
    return dict() if pat == inst else None

def merge_matches(*matches):
    if len(matches) == 0:
        return dict()
    if any(match is None for match in matches):
        return None
    res = dict(iter(matches[0].items()))
    for match2 in matches[1:]:
        for sym,lmda in match2.items():
            if sym in res:
                if not equiv_alpha(lmda,res[sym]):
                    return None
            else:
                res[sym] = lmda
    return res

def equiv_alpha(x,y):
    """check if two closed terms are equivalent module alpha
    conversion. for now, we assume the terms are closed
    """
    if x == y:
        return True
    if il.is_lambda(x) and il.is_lambda(y):
        return x.body == il.substitute(y.body,list(zip(x.variables,y.variables)))
    return False
    pass

def goal_free_vars(goal):
    if __debug__: xtracer.trace("proof.GoalFreeVars ENTER label=%s" % _goal_label_trace(goal))
    prems = goal_prems(goal)
    conc = goal_conc(goal)
    fmlas = [x.formula for x in prems if isinstance(x,ia.LabeledFormula)] + [conc]
    result = list(lu.used_variables_in_order_asts(fmlas))
    if __debug__: xtracer.trace("proof.GoalFreeVars EXIT nvars=%d" % len(result))
    return result

def var_subst_goal(goal,subst):
    """ Apply a variable substitution to a goal. """
    if __debug__: xtracer.trace("proof.varSubstGoal ENTER label=%s nsubs=%d" % (_goal_label_trace(goal), len(subst)))
    prems = [var_subst_goal(prem,subst) for prem in goal_prems(goal)]
    conc = goal_conc(goal)
    if not isinstance(conc,ia.SchemaBody):
        conc = apply_to_conc(conc,lambda x: il.substitute(x,subst))
    result = clone_goal(goal,prems,conc)
    if __debug__: xtracer.trace("proof.varSubstGoal EXIT HASH canon=%s" % result.canon())
    return result

def apply_to_conc(conc,func):
    if conc is None:
        if __debug__: xtracer.trace("proof.ApplyToConc ENTER concNil=true")
        return conc
    if __debug__: xtracer.trace("proof.ApplyToConc ENTER type=%s HASH canon=%s" % (type(conc).__name__, conc.canon() if hasattr(conc,'canon') else str(conc)))
    if isinstance(conc,ia.TemporalModels):
        result = conc.clone([func(conc.fmla)])
        if __debug__: xtracer.trace("proof.ApplyToConc EXIT type=TemporalModels HASH canon=%s" % result.canon())
        return result
    result = func(conc)
    if result is not None and hasattr(result,'canon'):
        if __debug__: xtracer.trace("proof.ApplyToConc EXIT type=Expr HASH canon=%s" % result.canon())
    else:
        if __debug__: xtracer.trace("proof.ApplyToConc EXIT type=%s passthrough" % type(result).__name__)
    return result

# Convert a goal to skolem normal form. This means the premises are in
# universal prenex form and the conclusion is in existential prenex
# form. If argument 'prenex' is false, don't convert to prenex form.

def skolemize_goal(goal,prenex=True):
    if __debug__: xtracer.trace("proof.SkolemizeGoal ENTER prenex=%s label=%s HASH canon=%s" % (prenex, _goal_label_trace(goal), goal.canon()))
    var_uniq = il.VariableUniqifier()
    vocab = goal_vocab(goal)
    used_names = set(x.name for x in vocab.symbols)
    used_names.update(x.name for x in goal_free(goal))
    renamer = iu.UniqueRenamer(used = used_names)
    skfuns = []
    if not prenex:
        variables = goal_free_vars(goal)
        if __debug__: xtracer.trace("proof.SkolemizeGoal freeVarsPass nvariables=%d" % len(variables))
        sks = [il.Symbol(renamer('_'+v.name),v.sort) for v in variables]
        for v, sk in zip(variables, sks):
            if __debug__: xtracer.trace("proof.SkolemizeGoal substitute v=%s sk=%s" % (v.name, sk.name))
        subst = list(zip(variables,sks))
        goal = var_subst_goal(goal,subst)
        skfuns += sks

    def rec(goal,pos):
        if not isinstance(goal,ia.LabeledFormula):
            return goal
        prems = [rec(prem,not pos) for prem in goal_prems(goal)]
        conc = skolemize_fmla(goal_conc(goal),pos,renamer,skfuns,prenex)
        return clone_goal(goal,prems,conc)
    goal = rec(goal,True)
    result = clone_goal(goal,[ia.ConstantDecl(s) for s in skfuns]+goal_prems(goal), goal_conc(goal))
    if __debug__: xtracer.trace("proof.SkolemizeGoal EXIT nskfuns=%d HASH canon=%s" % (len(skfuns), result.canon()))
    return result


def skolemize_fmla(fmla,pos,renamer,skfuns,prenex=True):
    if __debug__: xtracer.trace("proof.SkolemizeFmla ENTER pos=%s prenex=%s type=%s" % (pos, prenex, type(fmla).__name__))
    univs = []
    outer = []
    var_uniq = il.VariableUniqifier(used=renamer.used) # don't capture any free symbols!
    def rec( fmla,pos):
        if isinstance(fmla,il.Not):
            if __debug__: xtracer.trace("proof.SkolemizeFmla branch type=Not pos=%s" % pos)
            return fmla.clone([rec(fmla.args[0],not pos)])
        if isinstance(fmla,il.Implies):
            if __debug__: xtracer.trace("proof.SkolemizeFmla branch type=Implies pos=%s" % pos)
            return fmla.clone([
                rec(fmla.args[0],not pos),
                rec(fmla.args[1],pos),
            ])
        if isinstance(fmla,(il.And,il.Or)):
            if __debug__: xtracer.trace("proof.SkolemizeFmla branch type=%s pos=%s nterms=%d" % (type(fmla).__name__, pos, len(fmla.args)))
            return fmla.clone([rec(arg,pos) for arg in fmla.args])
        is_e = il.is_exists(fmla)
        is_a = il.is_forall(fmla)
        if is_a and pos or is_e and not pos:
            fvs = list(x for x in iu.unique(lu.variables_ast(fmla)) if x in outer)
            body = fmla.body
            if __debug__: xtracer.trace("proof.SkolemizeFmla branch type=Skolemize pos=%s isE=%s isA=%s nvars=%d varNames=%s" % (pos, is_e, is_a, len(fmla.variables), [v.name for v in fmla.variables]))
            for v in fmla.variables:
                sym = il.Symbol(renamer('_'+v.name),
                                il.FuncConstSort(*([w.sort for w in fvs] + [v.sort])))
                if __debug__: xtracer.trace("proof.SkolemizeFmla skolemize v=%s sk=%s" % (v.name, sym.name))
                term = sym(*fvs) if fvs else sym
                skfuns.append(sym)
                body =  il.substitute(body,[(v,term)])
            return rec(body,pos)
        if is_e and pos or is_a and not pos:
            body = fmla.body
            if __debug__: xtracer.trace("proof.SkolemizeFmla branch type=Universalize pos=%s isE=%s isA=%s nvars=%d" % (pos, is_e, is_a, len(fmla.variables)))
            for v in fmla.variables:
                u = var_uniq(v)
                if prenex:
                    univs.append(u)
                outer.append(u)
                body = il.substitute(body,[(v,u)])
                if __debug__: xtracer.trace("proof.SkolemizeFmla universalize v=%s u=%s" % (v.name, u.name))
            res = rec(body,pos)
            if not prenex:
                res = type(fmla)(outer[-len(fmla.variables):],res)
            for v in fmla.variables:
                outer.pop()
            return res
        if isinstance(fmla,ia.TemporalModels):
            if __debug__: xtracer.trace("proof.SkolemizeFmla branch type=TemporalModels pos=%s" % pos)
            return fmla.clone([rec(fmla.args[0],pos)])
        if __debug__: xtracer.trace("proof.SkolemizeFmla branch type=passthrough pos=%s astType=%s" % (pos, type(fmla).__name__))
        return fmla
    body = rec(fmla,pos)
    if univs:
        if __debug__: xtracer.trace("proof.SkolemizeFmla univsWrap pos=%s nunivs=%d" % (pos, len(univs)))
        quant = il.Exists if pos else il.ForAll
        if isinstance(body,ia.TemporalModels):
            body = body.clone([quant(univs,body.args[0])])
        else:
            body = quant(univs,body)
    if __debug__: xtracer.trace("proof.SkolemizeFmla EXIT nskfuns=%d" % len(skfuns))
    return body

def compile_witness_list(proof,goal):
    if __debug__: xtracer.trace("proof.CompileWitnessList ENTER nArgs=%d goalLabel=%s" % (len(proof.args), _goal_label_trace(goal)))
#    the_goal_vocab = goal_vocab(goal,get_bound_vars=True)
    the_goal_vocab = goal_vocab(goal)
    the_goal_vocab.variables.extend(list(logic_util.used_variables(goal_conc(goal))))
    result = []
    for i, d in enumerate(proof.args):
        compiled = compile_expr_vocab(d,the_goal_vocab)
        if compiled is not None:
            if __debug__: xtracer.trace("proof.CompileWitnessList arg=%d HASH canon=%s" % (i, compiled.canon() if hasattr(compiled,'canon') else str(compiled)))
            result.append(compiled)
        else:
            if __debug__: xtracer.trace("proof.CompileWitnessList arg=%d compiledNil" % i)
    if __debug__: xtracer.trace("proof.CompileWitnessList EXIT nresult=%d" % len(result))
    return result
    
def compile_with_goal_vocab(expr,goal):
#    the_goal_vocab = goal_vocab(goal,get_bound_vars=True)
    the_goal_vocab = goal_vocab(goal)
#    the_goal_vocab.variables.extend(list(logic_util.used_variables(goal_conc(goal))))
    return compile_expr_vocab_ext(expr,the_goal_vocab)

# This function compiles a defintion onto a goal, as a service to tactics
# The result for definiton f(X:t) = exp:u is to add two premises:
# - function f(X:t) : u
# - property forall X. f(X) = expr
#
# TRICKY: Operators in expr are normalized, but the forall quantifier is not normalized,
# meaning you may get multiple variables in a single quantifier. This means that natural
# deduction rules may not work as expected on the property, but it is easier for
# tactics to parse the property.

def compile_definition_goal_vocab(df,goal):
    vocab = goal_vocab(goal)
    free = goal_free(goal)
    lf = df.args[0]
    lhs = lf.formula.args[0]
    ts = il.TopFunctionSort(len(lhs.args))
    newsym = il.Symbol(lhs.rep,ts)
    with il.WithSymbols([newsym]):
        vars = lf.formula.args[0].args
        body = ia.Atom('=',lf.formula.args)
        fmla = ia.Forall(vars,body) if vars else body
        elf = lf.clone([lf.label,fmla])
        lf = compile_expr_vocab(elf,vocab)
        thing = lf.formula.body if vars else lf.formula
        thing = il.normalize_ops(thing)
        # this normalizes the body but not the quantifier
        lf = lf.clone([lf.label,lf.formula.clone([thing]) if vars else thing])
        sym = thing.args[0].rep
        deps = list(lu.symbols_ilu_ast(thing.args[1]))
        if sym in deps:
            raise NoMatch(lf,"no proof given for recursive definition")
        # TODO: allow proofs of recursive definitions
        cd = ia.ConstantDecl(sym)
        cd.lineno = lf.lineno
        lf.definition = True
        goal = goal_add_prem(goal,cd,lf.lineno)
        goal = goal_add_prem(goal,lf,lf.lineno)
        if sym in vocab.sorts or sym in vocab.symbols or sym in free:
            raise Redefinition(df,"redefinition of {}".format(sym))
        return goal

def remove_unused_definitions_goal(goal):
    prems = goal_prems(goal)
    prems = list(reversed(prems))
    conc = goal_conc(goal)
    if isinstance(conc,ia.TemporalModels):
        fmlas = conc.model.fmlas + [conc.fmla]
    else:
        fmlas = [conc]
    syms = lu.used_symbols_asts(fmlas)
    new_prems = []
    for x in prems:
        if goal_is_property(x) and x.definition:
            sym = il.drop_universals(x.formula).args[0].rep
            if sym not in syms:
                continue
        elif goal_is_defn(x):
            if goal_defines(x) not in syms:
                continue
        syms.update(lu.used_symbols_ast(x))
        new_prems.append(x)
    return clone_goal(goal,list(reversed(new_prems)),goal_conc(goal))
    
def match_from_defn(defn):
    vs = set()
    defn = defn.formula
    while isinstance(defn,il.ForAll):
        vs.update(defn.variables)
        defn = defn.body
    if il.is_eq(defn) or isinstance(defn,il.Iff):
        lhs,rhs = defn.args
        if il.is_app(lhs) & (all(x in vs for x in lhs.args) or True):
            if iu.distinct(lhs.args):
                return {lhs.rep : il.Lambda(lhs.args,rhs)}
    raise ProofError(defn,'not a definition')

def match_from_defns(defns):
    matches = [match_from_defn(d) for d in  defns]
    lhs = list(matches[0].keys())[0]
    assert all(lhs in m for m in matches)
    return {lhs:[m[lhs] for m in matches]}

def unfold_goal(goal,defns):
    for rdefs in defns:
        match = match_from_defns(rdefs)
        goal = apply_match_goal(match,goal,apply_match_alt)
    return goal

def unfold_fmla(fmla,defns):
    for rdefs in defns:
        match = match_from_defns(rdefs)
        fmla = apply_match_alt(match,fmla)
    return fmla

def goal_apply_to_prem(goal,premname,fn):
    if __debug__: xtracer.trace("proof.GoalApplyToPrem ENTER goalLabel=%s premName=%s" % (_goal_label_trace(goal), premname))
    prems = goal_prems(goal)
    premmap = dict((x.name,idx) for idx,x in enumerate(prems))
    if premname in premmap:
        idx = premmap[premname]
        prem = prems[idx]
        if not isinstance(prem,ia.LabeledFormula):
            if __debug__: xtracer.trace("proof.GoalApplyToPrem EXIT notLabeledFormula")
            return None
        result = clone_goal(goal,prems[0:idx]+[fn(prems[idx])]+prems[idx+1:],goal_conc(goal))
        if __debug__: xtracer.trace("proof.GoalApplyToPrem EXIT HASH canon=%s" % result.canon())
        return result
    if __debug__: xtracer.trace("proof.GoalApplyToPrem EXIT notFound premName=%s" % premname)
    return None

def goal_apply_to_conc(goal,fn):
    if __debug__: xtracer.trace("proof.GoalApplyToConc ENTER label=%s" % _goal_label_trace(goal))
    result = clone_goal(goal,goal_prems(goal),fn(goal_conc(goal)))
    if __debug__: xtracer.trace("proof.GoalApplyToConc EXIT HASH canon=%s" % result.canon())
    return result

# When instantiating a schema, unmatched free variables occurring only
# in the conclusion can be universally quantified.

def close_unmatched(goal,match):
    if __debug__: xtracer.trace("proof.CloseUnmatched ENTER label=%s nmatch=%d" % (_goal_label_trace(goal), len(match)))
    conc = goal_conc(goal)
    prem_vars = lu.used_variables_asts(goal_prem_goals(goal))
    conc_vars = [x for x in iu.unique(lu.variables_ast(conc))
                 if x not in match and x not in prem_vars]
    for v in reversed(conc_vars):
        conc = il.ForAll([v],conc)
    result = clone_goal(goal,goal_prems(goal),conc)
    if __debug__: xtracer.trace("proof.CloseUnmatched EXIT ntoClose=%d HASH canon=%s" % (len(conc_vars), result.canon()))
    return result

# When instantiating a schema, we drop the premises that supplied in the
# proof goal.

def drop_supplied_prems(schema,goal,proof_match):
    if __debug__: xtracer.trace("proof.DropSuppliedPrems ENTER schemaLabel=%s goalLabel=%s nMatch=%d" % (_goal_label_trace(schema), _goal_label_trace(goal), len(proof_match)))
    gprems = goal_prems_by_name(goal)
    pmap = dict()
    for m in proof_match:
        lhs,rhs = m.args
        if isinstance(lhs,ia.Atom) and len(lhs.args) == 0:
            if isinstance(rhs,ia.Atom) and len(rhs.args) == 0:
                pmap[lhs.rep] = rhs.rep
    def is_supplied(prem):
        if isinstance(prem,ia.LabeledFormula) and prem.name in pmap:
            gname = pmap[prem.name]
            return gname in gprems and goals_eq_mod_alpha(prem,gprems[gname])
        return False
    new_prems = [x for x in goal_prems(schema) if not is_supplied(x)]
    result = clone_goal(schema,new_prems,goal_conc(schema))
    if __debug__: xtracer.trace("proof.DropSuppliedPrems EXIT nnewPrems=%d" % len(new_prems))
    return result

# Remove the "explicit" tag from a goal

def remove_explicit(goal):
    if __debug__: xtracer.trace("proof.RemoveExplicit ENTER label=%s explicit=%s" % (_goal_label_trace(goal), hasattr(goal,'explicit') and goal.explicit))
    if hasattr(goal,'explicit') and goal.explicit:
        goal = goal.clone(goal.args)
        goal.explicit = False
        if __debug__: xtracer.trace("proof.RemoveExplicit EXIT cleared")
        return goal
    if __debug__: xtracer.trace("proof.RemoveExplicit EXIT passthrough")
    return goal

def rename_prem_no_clash(prem,decl):
    if __debug__: xtracer.trace("proof.RenamePremNoClash ENTER premLabel=%s declLabel=%s" % (_goal_label_trace(prem), _goal_label_trace(decl)))
    names = set(x.name for x in goal_prem_goals(decl))
    rn = iu.UniqueRenamer(used=names)
    new_name = rn(prem.name)
    result = prem.rename(new_name)
    if __debug__: xtracer.trace("proof.RenamePremNoClash EXIT newName=%s" % new_name)
    return result

class AddSymbols(object):
    """ temporarily add some symbols to a set of symbols """
    def __init__(self,symset,symlist):
        self.symset,self.symlist = symset,list(symlist)
    def __enter__(self):
        global sig
        self.saved = []
        for sym in self.symlist:
            if sym in self.symset:
                self.saved.append(sym)
                self.remove(sym)
            self.symset.add(sym)
        return self
    def __exit__(self,exc_type, exc_val, exc_tb):
        global sig
        for sym in self.symlist:
            self.symset.remove(sym)
        for sym in self.saved:
            self.symset.add(sym)
        return False # don't block any exceptions

class RemoveSymbols(object):
    """ temporarily add some symbols to a set of symbols """
    def __init__(self,symset,symlist):
        self.symset,self.symlist = symset,list(symlist)
    def __enter__(self):
        global sig
        self.saved = []
        for sym in self.symlist:
            if sym in self.symset:
                self.saved.append(sym)
                self.remove(sym)
        return self
    def __exit__(self,exc_type, exc_val, exc_tb):
        global sig
        for sym in self.saved:
            self.symset.add(sym)
        return False # don't block any exceptions

registered_tactics = dict()

def register_tactic(name,tactic):
    registered_tactics[name] = tactic
