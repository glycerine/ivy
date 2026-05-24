#
# Copyright (c) Microsoft Corporation. All Rights Reserved.
#
from . import ivy_utils as iu
from . import ivy_logic as il
from . import ivy_logic_utils as lu
from . import ivy_solver
from . import ivy_concept_space as ics
from . import ivy_ast
from . import xtracer

from collections import defaultdict
import string
import functools

################################################################################
#
# This Context object contains all the definitions in the current Ivy module
#
################################################################################

class Module(object):

    def __init__(self):
        self.clear()

    def clear(self):

        # these fields represent all the module declarations

        self.all_relations = [] # this includes both base and derived relations in declaration order
        self.definitions = []  # TODO: these are actually "derived" relations
        self.labeled_axioms = []
        self.labeled_props = []
        self.labeled_inits = []
        self.init_cond = lu.true_clauses()
        self.relations = dict()  # TODO: this is redundant, remove
        self.functions = dict()  # TODO: this is redundant, remove
        self.updates = []
        self.schemata = dict()
        self.theorems = dict()

        self.instantiations = []
        self.concept_spaces = []
        self.abstraction_predicates = []
        self.labeled_conjs = []  # conjectures
        self.postconds = defaultdict(list) # action name -> list of LabeledFormula
        self.hierarchy = defaultdict(dict)
        self.actions = {}
        self.predicates = {}
        self.assertions = []
        self.mixins = defaultdict(list)
        self.public_actions = {}
        self.isolates = {}
        self.exports = []
        self.imports = []
        self.delegates = []
        self.public_actions = {} # dict of the exported actions (insertion-ordered)
        self.progress = []  # list of progress properties
        self.rely = [] # list of rely relations
        self.mixord = [] # list of mixin order relations
        self.destructor_sorts = {}
        self.sort_destructors = defaultdict(list)
        self.constructor_sorts = {}
        self.sort_constructors = defaultdict(list)
        self.privates = set() # set of string (names of private actions)
        self.interps = defaultdict(list) # maps type names to lists of labeled interpretions
        self.natives = [] # list of NativeDef
        self.native_definitions = [] # list of definitions whose rhs is NativeExpr
        self.initializers = [] # list of name,action pairs
        self.initial_actions = [] # list of actions with formal parameters
        self.params = [] # list of symbol
        self.param_defaults = [] # list of string or None
        self.ghost_sorts = set() # set of sort names
        self.native_types = {} # map from sort names to ivy_ast.NativeType
        self.sort_order = [] # list of sorts names in order declared
        self.symbol_order = [] # list of symbols in order declared
        self.aliases = {} # map from name to name
        self.before_export = {} # map from string to action
        self.attributes = {} # map from name to atom
        self.variants = defaultdict(list) # map from sort name to list of sort
        self.supertypes = defaultdict(list) # map from subtype sort name to supertype sort
        self.ext_preconds = {} # map from action name to formula
        self.proofs = [] # list of pair (labeled formula, proof)
        self.named = [] # list of pair (labeled formula, atom)
        self.subgoals = [] # (labeled formula * labeled formula list) list
        self.isolate_info = None # IsolateInfo or None
        self.conj_actions = dict() # map from conj names to action name list
        self.conj_subgoals = None # None or labeled formula list
        self.assumed_invariants = [] # labeled_formula_list
        self.finite_sorts = set() # set of sort names
        self.isolate_proofs = {}
        self.isolate_proof = None
        self.logics = []
        self.sig = il.sig.copy() # capture the current signature

    def __enter__(self):
        global module
        self.old_module = module
        self.old_sig = il.sig
        module = self
        il.sig = self.sig
        ivy_solver.clear()   # this clears cached values, needed when changing sig
        return self

    def __exit__(self,exc_type, exc_val, exc_tb):
        global module
        module = self.old_module
        il.sig = self.old_sig
        return False # don't block any exceptions

    def get_axioms(self):
        res = self.axioms
        for n,sch in self.schemata.items():
            res += sch.formula.instances
        return res

    def background_theory(self, symbols=None):
        if hasattr(self,"theory"):
            return self.theory
        return lu.Clauses([])

    def add_to_hierarchy(self,name):
        if iu.ivy_compose_character in name:
            pref,suff = str.rsplit(name,iu.ivy_compose_character,1)
            self.add_to_hierarchy(pref)
            self.hierarchy[pref][suff] = True
        else:
            self.hierarchy['this'][name] = True

    def add_object(self,name):
        assert not isinstance(name,ivy_ast.This)
        self.hierarchy[name]

    @property
    def axioms(self):
        return [drop_label(x) for x in self.labeled_axioms if not x.temporal]

    @property
    def conjs(self):
        # This returns the list of conjectures as Clauses, without labels
        res = []
        for c in self.labeled_conjs:
            fmla = c.formula
            clauses = lu.formula_to_clauses(fmla)
            clauses.lineno = c.lineno
            res.append(clauses)
        return res

    def update_theory(self):
        theory = list(self.get_axioms())
        defs = []
        # axioms of the derived relations TODO: used only the
        # referenced ones, but we need to know abstract domain for
        # this
        for ldf in self.definitions:
            cnst = ldf.formula.to_constraint()
            if all(isinstance(p,il.Variable) for p in ldf.formula.args[0].args):
                if not isinstance(ldf.formula,il.DefinitionSchema):
#                    theory.append(ldf.formula) # TODO: make this a def?
                    ax = ldf.formula
                    if isinstance(ax.rhs(),il.Some):
                        ax = ax.to_constraint()
                        if ldf.formula.args[0].args:
                            ax = il.ForAll(ldf.formula.args[0].args,ax)
                        theory.append(ax) # TODO: make this a def?
                    else:
                        defs.append(ax)
        # extensionality axioms for structs
        for sort in sorted(self.sort_destructors):
            destrs = self.sort_destructors[sort]
            if any(d.name in self.sig.symbols for d in destrs):
                ea = il.extensionality(destrs)
                if il.is_epr(ea):
                    theory.append(ea)
        # exclusivity axioms for variants
        theory.extend(self.variant_axioms())
        self.theory = lu.Clauses(theory,defs)

    def variant_axioms(self):
        theory = []
        for sort in sorted(self.variants):
            sort_variants = self.variants[sort]
            if any(v.name in self.sig.sorts for v in sort_variants) and sort in self.sig.sorts:
                ea = il.exclusivity(self.sig.sorts[sort],sort_variants)
                theory.append(ea) # these are always in EPR
        return theory


    def theory_context(self):
        """ Set up to instiate the non-epr axioms """
        """ Return a set of clauses which represent the background theory
        restricted to the given symbols (should be like the result of used_symbols).
        """
        self.update_theory()

        non_epr = {}
        for ldf in self.definitions:
            cnst = ldf.formula.to_constraint()
            if not all(isinstance(p,il.Variable) for p in ldf.formula.args[0].args):
                non_epr[ldf.formula.defines()] = (ldf,cnst)
        return ModuleTheoryContext(non_epr)
        

    def is_variant(self,lsort,rsort):
        """ true if rsort is a variant of lsort """
        return (all(isinstance(a,il.UninterpretedSort) for a in (lsort,rsort))
                and lsort.name in self.variants and rsort in self.variants[lsort.name])

    def variant_index(self,lsort,rsort):
        """ returns the index of variant rsort of lsort """
        for idx,sort in enumerate(self.variants[lsort.name]):
            if sort == rsort:
                return idx

    # Gets an estimate of the cardinality of a sort, or None. This is
    # used for unrolling loops over the sort.

    def sort_card(self,sort):
        if il.is_function_sort(sort):
            return None
        attr = iu.compose_names(sort.name,'cardinality')
        if attr in self.attributes:
            return int(self.attributes[attr].rep)
        return sort_card(sort)


    # This makes a semi-shallow copy so we can side-effect 

    def copy(self):
        if __debug__: xtracer.trace("module.Copy ENTER actions=%d isolates=%d" % (len(self.actions), len(self.isolates)))
        m = Module()
        from copy import copy
        for x,y in self.__dict__.items():
            if x == 'sig':
                m.__dict__[x] = y.copy()
            else:
                m.__dict__[x] = copy(y)
        if __debug__: xtracer.trace("module.Copy EXIT actions=%d isolates=%d" % (len(m.actions), len(m.isolates)))
        return m

    # This removes implemented types

    def canonize_types(self):
        global sort_refinement
        with self.sig:
            sort_refinement = il.sort_refinement()
            if len(list(sort_refinement)) == 0:
                return # save time if nothing to do
            self.definitions = resort_labeled_asts(self.definitions)
            self.labeled_axioms = resort_labeled_asts(self.labeled_axioms)
            self.labeled_props = resort_labeled_asts(self.labeled_props)
            self.labeled_inits = resort_labeled_asts(self.labeled_inits)
            self.init_cond = resort_clauses(self.init_cond)
            self.concept_spaces = resort_concept_spaces(self.concept_spaces)
            self.labeled_conjs = resort_labeled_asts(self.labeled_conjs)
            self.assertions = resort_labeled_asts(self.assertions)
            self.progress = resort_asts(self.progress)
            self.initializers = resort_name_ast_pairs(self.initializers)
            self.params = resort_symbols(self.params)
            self.ghost_sorts = remove_refined_sortnames_from_set(self.ghost_sorts)
            self.sort_order = remove_refined_sortnames_from_list(self.sort_order)
            self.symbol_order = resort_symbols(self.symbol_order)
            self.aliases = resort_aliases_map(self.aliases)
            self.before_export = resort_map_any_ast(self.before_export)
            self.ext_preconds = resort_map_any_ast(self.ext_preconds)
            lu.resort_sig(sort_refinement)

            
        # Make concept spaces from the conjecture

    def update_conjs(self):
        mod = self
        for i,cax in enumerate(mod.labeled_conjs):
            fmla = cax.formula
            csname = 'conjecture:'+ str(i)
            # Use deterministic left-to-right traversal order (first-occurrence
            # dedup) instead of set hash-bucket order, so Go can produce the
            # identical concept-space label ordering for cross-language canon
            # comparison.
            variables = list(lu.used_variables_in_order_ast(fmla))
            sort = il.RelationSort([v.sort for v in variables])
            sym = il.Symbol(csname,sort)
            space = ics.NamedSpace(il.Literal(0,fmla))
            mod.concept_spaces.append((sym(*variables),space))

    def call_graph(self):
        callgraph = defaultdict(list)
        for actname,action in self.actions.items():
            for called_name in action.iter_calls():
                callgraph[called_name].append(actname)
        return callgraph

    def canon(self):
        """Return a single canonical s-expression of key module state.
        Used by sig_check to Merkle-chain module state alongside Sig.
        Matches Go's Module.Canon() in module/canon.go."""
        parts = []
        parts.append("axioms:" + _canon_lf_slice(self.labeled_axioms))
        parts.append("defs:" + _canon_lf_slice(self.definitions))
        parts.append("props:" + _canon_lf_slice(self.labeled_props))
        parts.append("inits:" + _canon_lf_slice(self.labeled_inits))
        parts.append("conjs:" + _canon_lf_slice(self.labeled_conjs))
        if hasattr(self, 'sig') and self.sig is not None:
            parts.append("sig.sorts:" + _canon_sort_map(self.sig.sorts))
            parts.append("sig.interp:" + _canon_interp_map(self.sig.interp))
        parts.append("schemata:" + _canon_schema_map(self.schemata))
        if hasattr(self, 'init_cond') and self.init_cond is not None:
            parts.append("initCond:" + self.init_cond.sexp())
        if hasattr(self, 'theory') and self.theory is not None:
            parts.append("theory:" + self.theory.sexp())
        parts.append("actions:" + _canon_action_map(self.actions))
        return "(module %s)" % " ".join(parts)

    def canon_snapshot(self, label):
        """Emit canonical s-expression snapshot of ALL module state via xtracer.
        Matches Go's Module.CanonSnapshot() in module/canon.go."""
        if not __debug__:
            return
        xtracer.trace("module.CanonSnapshot ENTER label=%s" % label)

        label += " HASH canon= "

        # Group 1: Declarations (labeled formula slices)
        xtracer.trace("module.CanonSnapshot %s labeledAxioms=%s" % (label, _canon_lf_slice(self.labeled_axioms)))
        xtracer.trace("module.CanonSnapshot %s definitions=%s" % (label, _canon_lf_slice(self.definitions)))
        xtracer.trace("module.CanonSnapshot %s labeledProps=%s" % (label, _canon_lf_slice(self.labeled_props)))
        xtracer.trace("module.CanonSnapshot %s labeledInits=%s" % (label, _canon_lf_slice(self.labeled_inits)))
        xtracer.trace("module.CanonSnapshot %s labeledConjs=%s" % (label, _canon_lf_slice(self.labeled_conjs)))
        xtracer.trace("module.CanonSnapshot %s assertions=%s" % (label, _canon_lf_slice(self.assertions)))
        xtracer.trace("module.CanonSnapshot %s assumedInvs=%s" % (label, _canon_lf_slice(self.assumed_invariants)))
        xtracer.trace("module.CanonSnapshot %s nativeDefinitions=%s" % (label, _canon_lf_slice(self.native_definitions)))
        if self.conj_subgoals is not None:
            xtracer.trace("module.CanonSnapshot %s conjSubgoals=%s" % (label, _canon_lf_slice(self.conj_subgoals)))

        # Group 2: All relations
        xtracer.trace("module.CanonSnapshot %s allRelations=%s" % (label, _canon_expr_slice(self.all_relations)))

        # Group 3: Signature
        if hasattr(self, 'sig') and self.sig is not None:
            xtracer.trace("module.CanonSnapshot %s sig.sorts=%s" % (label, _canon_sort_map(self.sig.sorts)))
            xtracer.trace("module.CanonSnapshot %s sig.interp=%s" % (label, _canon_interp_map(self.sig.interp)))

        # Group 4: Relations and functions
        xtracer.trace("module.CanonSnapshot %s relations=%s" % (label, _canon_insmap_sort(self.relations)))
        xtracer.trace("module.CanonSnapshot %s functions=%s" % (label, _canon_insmap_sort(self.functions)))

        # Group 5: Schemata, theorems, predicates
        xtracer.trace("module.CanonSnapshot %s schemata=%s" % (label, _canon_schema_map(self.schemata)))
        xtracer.trace("module.CanonSnapshot %s theorems=%s" % (label, _canon_schema_map(self.theorems)))
        xtracer.trace("module.CanonSnapshot %s predicates=%s" % (label, _canon_schema_map(self.predicates)))

        # Group 6: InitCond and Theory
        if hasattr(self, 'init_cond') and self.init_cond is not None:
            xtracer.trace("module.CanonSnapshot %s initCond=%s" % (label, self.init_cond.sexp()))
        if hasattr(self, 'theory') and self.theory is not None:
            xtracer.trace("module.CanonSnapshot %s theory=%s" % (label, self.theory.sexp()))

        # Group 7: Actions
        if hasattr(self, 'actions') and self.actions:
            act_keys = list(self.actions.keys())
            xtracer.trace("module.CanonSnapshot %s actions.keys=%d keys=%s" % (label, len(act_keys), ",".join(act_keys)))
            for name, action in self.actions.items():
                xtracer.trace("module.CanonSnapshot %s action[%s]=%s" % (label, name, action.sexp()))
        xtracer.trace("module.CanonSnapshot %s beforeExport=%s" % (label, _canon_action_map(self.before_export)))

        # Group 8: Mixins, PublicActions
        xtracer.trace("module.CanonSnapshot %s mixins=%s" % (label, _canon_insmap_mixins(self.mixins)))
        xtracer.trace("module.CanonSnapshot %s publicActions=%s" % (label, _canon_insmap_bool(self.public_actions)))

        # Group 9: Postconds
        xtracer.trace("module.CanonSnapshot %s postconds=%s" % (label, _canon_postconds_map(self.postconds)))

        # Group 10: Initializers
        xtracer.trace("module.CanonSnapshot %s initializers=%s" % (label, _canon_named_action_slice(self.initializers)))
        xtracer.trace("module.CanonSnapshot %s initialActions=%s" % (label, _canon_action_slice(self.initial_actions)))

        # Group 11: Hierarchy
        xtracer.trace("module.CanonSnapshot %s hierarchy=%s" % (label, _canon_insmap_hierarchy(self.hierarchy)))

        # Group 12: Updates, Instantiations
        xtracer.trace("module.CanonSnapshot %s updates=%s" % (label, _canon_interface_slice(self.updates)))
        xtracer.trace("module.CanonSnapshot %s instantiations=%s" % (label, _canon_instantiation_slice(self.instantiations)))

        # Group 13: Isolates
        xtracer.trace("module.CanonSnapshot %s isolates=%s" % (label, _canon_isolate_map(self.isolates)))
        if self.isolate_info is not None:
            xtracer.trace("module.CanonSnapshot %s isolateInfo=%s" % (label, _canon_isolate_info(self.isolate_info)))
        xtracer.trace("module.CanonSnapshot %s isolateProofs=%s" % (label, _canon_schema_map(self.isolate_proofs)))
        if self.isolate_proof is not None:
            xtracer.trace("module.CanonSnapshot %s isolateProof=%s" % (label, _canon_node(self.isolate_proof)))

        # Group 14: Exports, imports, delegates
        xtracer.trace("module.CanonSnapshot %s exports=%s" % (label, _canon_exporter_slice(self.exports)))
        xtracer.trace("module.CanonSnapshot %s imports=%s" % (label, _canon_node_slice(self.imports)))
        xtracer.trace("module.CanonSnapshot %s delegates=%s" % (label, _canon_delegator_slice(self.delegates)))

        # Group 15: Sorts and destructors
        xtracer.trace("module.CanonSnapshot %s destructorSorts=%s" % (label, _canon_sort_map(self.destructor_sorts)))
        xtracer.trace("module.CanonSnapshot %s sortDestructors=%s" % (label, _canon_const_slice_map(self.sort_destructors)))
        xtracer.trace("module.CanonSnapshot %s constructorSorts=%s" % (label, _canon_sort_map(self.constructor_sorts)))
        xtracer.trace("module.CanonSnapshot %s sortConstructors=%s" % (label, _canon_const_slice_map(self.sort_constructors)))
        xtracer.trace("module.CanonSnapshot %s ghostSorts=%s" % (label, _canon_bool_set(self.ghost_sorts)))
        xtracer.trace("module.CanonSnapshot %s sortOrder=%s" % (label, _canon_string_slice(self.sort_order)))
        xtracer.trace("module.CanonSnapshot %s symbolOrder=%s" % (label, _canon_const_slice(self.symbol_order)))
        xtracer.trace("module.CanonSnapshot %s variants=%s" % (label, _canon_sort_slice_map(self.variants)))
        xtracer.trace("module.CanonSnapshot %s supertypes=%s" % (label, _canon_sort_map(self.supertypes)))
        xtracer.trace("module.CanonSnapshot %s finiteSorts=%s" % (label, _canon_bool_set(self.finite_sorts)))

        # Group 16: Interpretations and natives
        xtracer.trace("module.CanonSnapshot %s interps=%s" % (label, _canon_node_map_slice(self.interps)))
        xtracer.trace("module.CanonSnapshot %s natives=%s" % (label, _canon_node_slice(self.natives)))
        xtracer.trace("module.CanonSnapshot %s nativeTypes=%s" % (label, _canon_native_type_map(self.native_types)))

        # Group 17: Properties and proofs
        xtracer.trace("module.CanonSnapshot %s progress=%s" % (label, _canon_interface_slice(self.progress)))
        xtracer.trace("module.CanonSnapshot %s rely=%s" % (label, _canon_expr_slice(self.rely)))
        xtracer.trace("module.CanonSnapshot %s mixOrd=%s" % (label, _canon_node_slice(self.mixord)))
        xtracer.trace("module.CanonSnapshot %s privates=%s" % (label, _canon_bool_set(self.privates)))
        xtracer.trace("module.CanonSnapshot %s proofs=%s" % (label, _canon_proof_entry_slice(self.proofs)))
        xtracer.trace("module.CanonSnapshot %s named=%s" % (label, _canon_named_entry_slice(self.named)))
        xtracer.trace("module.CanonSnapshot %s subgoals=%s" % (label, _canon_subgoal_entry_slice(self.subgoals)))
        xtracer.trace("module.CanonSnapshot %s conjActions=%s" % (label, _canon_string_slice_map(self.conj_actions)))

        # Group 18: Parameters
        xtracer.trace("module.CanonSnapshot %s params=%s" % (label, _canon_const_slice(self.params)))
        xtracer.trace("module.CanonSnapshot %s paramDefaults=%s" % (label, _canon_node_slice(self.param_defaults)))

        # Group 19: Other
        xtracer.trace("module.CanonSnapshot %s aliases=%s" % (label, _canon_string_map(self.aliases)))
        xtracer.trace("module.CanonSnapshot %s attributes=%s" % (label, _canon_interp_map(self.attributes)))
        xtracer.trace("module.CanonSnapshot %s extPreconds=%s" % (label, _canon_expr_map(self.ext_preconds)))
        xtracer.trace("module.CanonSnapshot %s conceptSpaces=%s" % (label, _canon_concept_space_slice(self.concept_spaces)))
        xtracer.trace("module.CanonSnapshot %s logics=%s" % (label, _canon_string_slice(self.logics)))

        xtracer.trace("module.CanonSnapshot EXIT label=%s" % label)


def _canon_lf_slice(lfs):
    """Canonical s-expression for a list of LabeledFormulas."""
    if not lfs:
        return '[]'
    parts = []
    for lf in lfs:
        if hasattr(lf, 'canon'):
            parts.append(lf.canon())
        else:
            parts.append(str(lf))
    return '[%s]' % ' '.join(parts)


def _canon_sort_map(sorts):
    """Sorted canonical representation of a sort map."""
    if not sorts:
        return '(hash)'
    parts = []
    for k in sorted(sorts.keys()):
        v = sorts[k]
        if hasattr(v, 'sexp'):
            vs = v.sexp()
        else:
            vs = str(v)
        parts.append('%s:%s' % (k, vs))
    return '(hash %s)' % ' '.join(parts)


def _canon_interp_map(interp):
    """Sorted canonical representation of the interp map."""
    if not interp:
        return '(hash)'
    parts = []
    for k in sorted(interp.keys()):
        v = interp[k]
        if isinstance(v, str):
            vs = '"%s"' % v
        elif hasattr(v, 'canon'):
            vs = v.canon()
        elif hasattr(v, 'sexp'):
            vs = v.sexp()
        else:
            vs = str(v)
        parts.append('%s:%s' % (k, vs))
    return '(hash %s)' % ' '.join(parts)


def _canon_schema_map(schemata):
    """Sorted canonical representation of the schemata map."""
    if not schemata:
        return '(hash)'
    parts = []
    for k in sorted(schemata.keys()):
        v = schemata[k]
        if hasattr(v, 'canon'):
            vs = v.canon()
        else:
            vs = str(v)
        parts.append('%s:%s' % (k, vs))
    return '(hash %s)' % ' '.join(parts)


def _canon_action_map(actions):
    """Canonical representation of the actions map in insertion order.
    Matches Go's canonActionMap() in module/canon.go.
    Format: (insMap name1:sexp1 name2:sexp2 ...)"""
    if not actions:
        return '(insMap)'
    parts = []
    for k in actions:
        v = actions[k]
        if hasattr(v, 'sexp'):
            vs = v.sexp()
        elif hasattr(v, 'canon'):
            vs = v.canon()
        else:
            vs = str(v)
        parts.append('%s:%s' % (k, vs))
    return '(insMap %s)' % ' '.join(parts)


# ---------------------------------------------------------------------------
# New generic helpers
# ---------------------------------------------------------------------------

def _canon_expr_slice(exprs):
    """Canonical s-expression for a list of logic expressions.
    Handles (symbol, arity) tuples from all_relations by extracting the symbol."""
    if not exprs:
        return '[]'
    parts = []
    for e in exprs:
        # Handle (symbol, arity) tuples from all_relations
        if isinstance(e, tuple):
            e = e[0]
        if hasattr(e, 'sexp'):
            parts.append(e.sexp())
        else:
            parts.append(str(e))
    return '[%s]' % ' '.join(parts)


def _canon_action_slice(actions):
    """Canonical s-expression for a list of actions."""
    if not actions:
        return '[]'
    parts = []
    for a in actions:
        if hasattr(a, 'sexp'):
            parts.append(a.sexp())
        elif hasattr(a, 'canon'):
            parts.append(a.canon())
        else:
            parts.append(str(a))
    return '[%s]' % ' '.join(parts)


def _canon_bool_set(s):
    """Canonical representation of a set (or map with bool values) as sorted keys."""
    if not s:
        return '(set)'
    return '(set %s)' % ' '.join(sorted(s))


def _canon_string_slice(ss):
    """Canonical form of a list of strings."""
    if not ss:
        return '[]'
    return '[%s]' % ' '.join('"%s"' % s for s in ss)


def _canon_const_slice(cs):
    """Canonical form of a list of symbols/constants."""
    if not cs:
        return '[]'
    parts = []
    for c in cs:
        if hasattr(c, 'sexp'):
            parts.append(c.sexp())
        else:
            parts.append(str(c))
    return '[%s]' % ' '.join(parts)


def _canon_node_slice(nodes):
    """Canonical form of a list of AST nodes."""
    if not nodes:
        return '[]'
    parts = []
    for n in nodes:
        parts.append(_canon_node(n))
    return '[%s]' % ' '.join(parts)


def _canon_node(n):
    """Canonical form of a single AST node."""
    if n is None:
        return 'nil'
    if hasattr(n, 'canon'):
        return n.canon()
    return str(n)


def _canon_interface_slice(items):
    """Canonical form of a list of mixed-type items."""
    if not items:
        return '[]'
    parts = []
    for item in items:
        if item is None:
            parts.append('nil')
        elif hasattr(item, 'canon'):
            parts.append(item.canon())
        elif hasattr(item, 'sexp'):
            parts.append(item.sexp())
        else:
            parts.append(str(item))
    return '[%s]' % ' '.join(parts)


# ---------------------------------------------------------------------------
# New InsMap helpers (Python dicts are insertion-ordered)
# ---------------------------------------------------------------------------

def _canon_insmap_sort(m):
    """Canonical form of a dict mapping strings to sorts (insertion order)."""
    if not m:
        return '(insMap)'
    parts = []
    for k, v in m.items():
        if hasattr(v, 'sexp'):
            vs = v.sexp()
        else:
            vs = str(v)
        parts.append('%s:%s' % (k, vs))
    return '(insMap %s)' % ' '.join(parts)


def _canon_insmap_bool(m):
    """Canonical form of a dict mapping strings to bools (insertion order, keys only)."""
    if not m:
        return '(insMap)'
    return '(insMap %s)' % ' '.join(m.keys())


def _canon_insmap_mixins(m):
    """Canonical form of a dict mapping strings to lists of mixin defs."""
    if not m:
        return '(insMap)'
    parts = []
    for k, defs in m.items():
        mparts = []
        for d in defs:
            if hasattr(d, 'canon'):
                mparts.append(d.canon())
            else:
                mparts.append(str(d))
        parts.append('%s:[%s]' % (k, ' '.join(mparts)))
    return '(insMap %s)' % ' '.join(parts)


def _canon_insmap_hierarchy(m):
    """Canonical form of a dict mapping strings to dicts of strings to bools."""
    if not m:
        return '(insMap)'
    parts = []
    for parent, children in m.items():
        parts.append('%s:%s' % (parent, _canon_insmap_bool(children)))
    return '(insMap %s)' % ' '.join(parts)


# ---------------------------------------------------------------------------
# New sorted-map helpers
# ---------------------------------------------------------------------------

def _canon_postconds_map(m):
    """Canonical form of a dict mapping strings to lists of LabeledFormulas."""
    if not m:
        return '(hash)'
    parts = []
    for k in sorted(m.keys()):
        parts.append('%s:%s' % (k, _canon_lf_slice(m[k])))
    return '(hash %s)' % ' '.join(parts)


def _canon_sort_slice_map(m):
    """Canonical form of a dict mapping strings to lists of sorts."""
    if not m:
        return '(hash)'
    parts = []
    for k in sorted(m.keys()):
        ss = []
        for s in m[k]:
            if hasattr(s, 'sexp'):
                ss.append(s.sexp())
            else:
                ss.append(str(s))
        parts.append('%s:[%s]' % (k, ' '.join(ss)))
    return '(hash %s)' % ' '.join(parts)


def _canon_const_slice_map(m):
    """Canonical form of a dict mapping strings to lists of symbols/constants."""
    if not m:
        return '(hash)'
    parts = []
    for k in sorted(m.keys()):
        cs = []
        for c in m[k]:
            if hasattr(c, 'sexp'):
                cs.append(c.sexp())
            else:
                cs.append(str(c))
        parts.append('%s:[%s]' % (k, ' '.join(cs)))
    return '(hash %s)' % ' '.join(parts)


def _canon_string_map(m):
    """Canonical form of a dict mapping strings to strings."""
    if not m:
        return '(hash)'
    parts = []
    for k in sorted(m.keys()):
        parts.append('%s:"%s"' % (k, m[k]))
    return '(hash %s)' % ' '.join(parts)


def _canon_string_slice_map(m):
    """Canonical form of a dict mapping strings to lists of strings."""
    if not m:
        return '(hash)'
    parts = []
    for k in sorted(m.keys()):
        ss = ['"%s"' % s for s in sorted(m[k])]
        parts.append('%s:[%s]' % (k, ' '.join(ss)))
    return '(hash %s)' % ' '.join(parts)


def _canon_node_map_slice(m):
    """Canonical form of a dict mapping strings to lists of AST nodes."""
    if not m:
        return '(hash)'
    parts = []
    for k in sorted(m.keys()):
        parts.append('%s:%s' % (k, _canon_node_slice(m[k])))
    return '(hash %s)' % ' '.join(parts)


def _canon_expr_map(m):
    """Canonical form of a dict mapping strings to expressions."""
    if not m:
        return '(hash)'
    parts = []
    for k in sorted(m.keys()):
        v = m[k]
        if hasattr(v, 'sexp'):
            vs = v.sexp()
        else:
            vs = str(v)
        parts.append('%s:%s' % (k, vs))
    return '(hash %s)' % ' '.join(parts)


def _canon_isolate_map(m):
    """Canonical form of a dict mapping strings to IsolateDefs."""
    if not m:
        return '(hash)'
    parts = []
    for k in sorted(m.keys()):
        v = m[k]
        if hasattr(v, 'canon'):
            vs = v.canon()
        else:
            vs = str(v)
        parts.append('%s:%s' % (k, vs))
    return '(hash %s)' % ' '.join(parts)


def _canon_native_type_map(m):
    """Canonical form of a dict mapping strings to NativeType."""
    if not m:
        return '(hash)'
    parts = []
    for k in sorted(m.keys()):
        v = m[k]
        if hasattr(v, 'canon'):
            vs = v.canon()
        else:
            vs = str(v)
        parts.append('%s:%s' % (k, vs))
    return '(hash %s)' % ' '.join(parts)


# ---------------------------------------------------------------------------
# New struct-specific helpers
# ---------------------------------------------------------------------------

def _canon_named_action_slice(nas):
    """Canonical form of a list of (name, action) pairs."""
    if not nas:
        return '[]'
    parts = []
    for name, action in nas:
        if hasattr(action, 'sexp'):
            vs = action.sexp()
        elif hasattr(action, 'canon'):
            vs = action.canon()
        else:
            vs = str(action)
        parts.append('(%s:%s)' % (name, vs))
    return '[%s]' % ' '.join(parts)


def _canon_instantiation_slice(insts):
    """Canonical form of a list of (schema, inst) pairs."""
    if not insts:
        return '[]'
    parts = []
    for schema, inst in insts:
        parts.append('(schema:%s inst:%s)' % (_canon_node(schema), _canon_node(inst)))
    return '[%s]' % ' '.join(parts)


def _canon_isolate_info(info):
    """Canonical form of an IsolateInfo object."""
    if info is None:
        return 'nil'
    impls = _canon_mixin_triple_slice(getattr(info, 'implementations', []))
    monitors = _canon_mixin_triple_slice(getattr(info, 'monitors', []))
    return '(isolateInfo impls:%s monitors:%s)' % (impls, monitors)


def _canon_mixin_triple_slice(triples):
    """Canonical form of a list of mixin triples (mixer, mixee, action)."""
    if not triples:
        return '[]'
    parts = []
    for t in triples:
        mixer = getattr(t, 'mixer', str(t[0]) if isinstance(t, tuple) else '')
        mixee = getattr(t, 'mixee', str(t[1]) if isinstance(t, tuple) else '')
        action = getattr(t, 'action', t[2] if isinstance(t, tuple) else None)
        if action is not None and hasattr(action, 'sexp'):
            av = action.sexp()
        elif action is not None and hasattr(action, 'canon'):
            av = action.canon()
        else:
            av = str(action)
        parts.append('(mixer:%s mixee:%s action:%s)' % (mixer, mixee, av))
    return '[%s]' % ' '.join(parts)


def _canon_exporter_slice(exports):
    """Canonical form of a list of export definitions."""
    if not exports:
        return '[]'
    parts = []
    for e in exports:
        if hasattr(e, 'canon'):
            parts.append(e.canon())
        else:
            parts.append(str(e))
    return '[%s]' % ' '.join(parts)


def _canon_delegator_slice(delegates):
    """Canonical form of a list of delegate definitions."""
    if not delegates:
        return '[]'
    parts = []
    for d in delegates:
        if hasattr(d, 'canon'):
            parts.append(d.canon())
        else:
            parts.append(str(d))
    return '[%s]' % ' '.join(parts)


def _canon_concept_space_slice(css):
    """Canonical form of a list of concept space (label, body) pairs."""
    if not css:
        return '[]'
    parts = []
    for cs in css:
        label, body = cs[0], cs[1]
        if hasattr(label, 'sexp'):
            ls = label.sexp()
        else:
            ls = str(label)
        if hasattr(body, 'sexp'):
            bs = body.sexp()
        else:
            bs = str(body)
        parts.append('(label:%s body:%s)' % (ls, bs))
    return '[%s]' % ' '.join(parts)


def _canon_proof_entry_slice(proofs):
    """Canonical form of a list of (formula, proof) pairs."""
    if not proofs:
        return '[]'
    parts = []
    for formula, proof in proofs:
        fc = formula.canon() if hasattr(formula, 'canon') else str(formula)
        pc = _canon_node(proof)
        parts.append('(formula:%s proof:%s)' % (fc, pc))
    return '[%s]' % ' '.join(parts)


def _canon_named_entry_slice(named):
    """Canonical form of a list of (formula, atom) pairs."""
    if not named:
        return '[]'
    parts = []
    for formula, name in named:
        fc = formula.canon() if hasattr(formula, 'canon') else str(formula)
        nc = name.sexp() if hasattr(name, 'sexp') else str(name)
        parts.append('(formula:%s name:%s)' % (fc, nc))
    return '[%s]' % ' '.join(parts)


def _canon_subgoal_entry_slice(subs):
    """Canonical form of a list of (formula, subgoals) tuples."""
    if not subs:
        return '[]'
    parts = []
    for formula, subgoals in subs:
        fc = formula.canon() if hasattr(formula, 'canon') else str(formula)
        sc = _canon_lf_slice(subgoals)
        parts.append('(formula:%s subgoals:%s)' % (fc, sc))
    return '[%s]' % ' '.join(parts)


def resort_ast(ast):
    return lu.resort_ast(ast,sort_refinement)

def resort_clauses(clauses):
    return lu.resort_clauses(clauses,sort_refinement)

def resort_asts(asts):
    return [lu.resort_ast(ast,sort_refinement) for ast in asts]

def resort_labeled_asts(asts):
    return [ast.clone([ast.args[0],lu.resort_ast(ast.args[1],sort_refinement)]) for ast in asts]

resort_concept_spaces = resort_asts

def resort_map_symbol_sort(m):
    return dict((lu.resort_symbol(sym,sort_refinement),lu.resort_sort(sort,sort_refinement))
                for sym,sort in m.items())

def resort_name_ast_pairs(pairs):
    return [(n,lu.resort_ast(a,sort_refinement)) for n,a in pairs]

def resort_symbols(symbols):
    return [lu.resort_symbol(symbol,sort_refinement) for symbol in symbols]

def remove_refined_sortnames_from_set(sorts):
    refd = set(s.name for s in sort_refinement)
    return set(n for n in sorts if n not in refd)

def remove_refined_sortnames_from_list(sorts):
    refd = set(s.name for s in sort_refinement)
    return list(n for n in sorts if n not in refd)

def resort_aliases_map(amap):
    res = dict(iter(amap.items()))
    for s1,s2 in sort_refinement.items():
        res[s1.name] = s2.name

def resort_map_any_ast(m):
    return dict((a,lu.resort_ast(b,sort_refinement)) for a,b in m.items())


module = None

def instantiate_non_epr(non_epr,ground_terms):
    theory = []
    if ground_terms != None:
        matched = set()
        for term in ground_terms:
            if term.rep in non_epr and term not in matched:
                ldf,cnst = non_epr[term.rep]
                subst = dict((v,t) for v,t in zip(ldf.formula.args[0].args,term.args)
                             if not isinstance(v,il.Variable))
                if all(lu.is_ground_ast(x) for x in list(subst.values())):
                       inst = lu.substitute_constants_ast(cnst,subst)
                       theory.append(inst)
#                iu.dbg('inst')
                matched.add(term)
    return lu.Clauses(theory)


def background_theory(symbols = None):
    return module.background_theory(symbols)

def find_action(name):
    return module.actions.get(name,None)

param_logic = iu.Parameter("complete",','.join(il.default_logics),
                           check=lambda ls: all(s in il.logics for s in ls.split(',')))

def logics():
    if module.logics:
        return module.logics
    return param_logic.get().split(',')

def drop_label(labeled_fmla):
    return labeled_fmla.formula if hasattr(labeled_fmla,'formula') else labeled_fmla

class ModuleTheoryContext(object):

    def __init__(self,non_epr):
        self.non_epr = non_epr

    def __enter__(self):
        self.old_instantiator = lu.instantiator
        lu.instantiator = self
        return self

    def __exit__(self,exc_type, exc_val, exc_tb):
        lu.instantiator = self.old_instantiator
        return False # don't block any exceptions

    def __call__(self,clauses):
        return instantiate_non_epr(self.non_epr,clauses)

    def rename(self,subst):
        new_non_epr = []
        for (df,cnst) in self.non_epr:
            dfns = df.defines()
            if dfns in subst:
                new_non_epr.append((lu.rename_ast(df,subst),lu.rename_ast(cnst,subst)))
        self.non_epr.extend(new_non_epr)

def relevant_definitions(symbols):
    dfn_map = dict((ldf.formula.defines(),ldf.formula.args[1]) for ldf in module.definitions)
    rch = set(iu.reachable(list(il.normalize_symbol(x) for x in symbols),lambda sym: lu.symbols_ilu_ast(dfn_map[sym]) if sym in dfn_map else []))
    return [ldf for ldf in module.definitions if ldf.formula.defines() in rch]
    
def sort_dependencies(mod,sortname,with_variants=True):
    if sortname in mod.sort_destructors:
        return [s.name for destr in mod.sort_destructors[sortname]
                for s in destr.sort.dom[1:] + (destr.sort.rng,)]
    if sortname in mod.native_types:
        t = mod.native_types[sortname]
        if isinstance(t,ivy_ast.NativeType):
            return [s.rep for s in t.args[1:] if s.rep in mod.sig.sorts]
    if with_variants and sortname in mod.variants:
        return [s.name for s in mod.variants[sortname]]
    return []

# Holds info about isolate for user consumption
#
# -- implementations is a list of pairs (mixer,mixee,action) for present action implementaitons
# -- monitors is a list of triples (mixer,mixee,action) for present monitors

class IsolateInfo(object):
    def __init__(self):
        self.implementations,self.monitors = [],[]
