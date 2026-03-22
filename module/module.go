// Package module provides the Module type, which holds all declarations,
// axioms, actions, and other state for an Ivy module.
//
// This corresponds to Python's ivy_module.py.
package module

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/xtracer"
)

// Module holds all the definitions and declarations in an Ivy module.
type Module struct {
	Cfg *Config // really for check. check imports module.
	//ModCfg *ModConfig

	// Declarations
	AllRelations  []lg.Expr // base and derived relations in declaration order
	Definitions   []*ast.LabeledFormula
	LabeledAxioms []*ast.LabeledFormula
	LabeledProps  []*ast.LabeledFormula
	LabeledInits  []*ast.LabeledFormula
	LabeledConjs  []*ast.LabeledFormula // conjectures
	Assertions    []*ast.LabeledFormula
	Postconds     map[string][]*ast.LabeledFormula // action name → postconditions
	AssumedInvs   []*ast.LabeledFormula            // assumed invariants

	// Relations and functions
	Relations map[string]lg.Sort
	Functions map[string]lg.Sort

	// Actions and mixins
	Actions        map[string]Action
	Mixins         map[string][]MixinDef
	PublicActions  map[string]bool
	Predicates     map[string]ast.Node
	Initializers   []NamedAction
	InitialActions []Action

	// Module structure
	Hierarchy      map[string]map[string]bool // parent → children
	Updates        []interface{}
	Schemata       map[string]ast.Node
	Theorems       map[string]ast.Node
	Instantiations []Instantiation

	// Isolates
	Isolates      map[string]*ast.IsolateDef
	IsolateInfo   *IsolateInfo
	IsolateProofs map[string]ast.Node
	IsolateProof  ast.Node

	// Exports and imports
	Exports   []Exporter
	Imports   []ast.Node
	Delegates []Delegator

	// Sorts and destructors
	DestructorSorts  map[string]lg.Sort
	SortDestructors  map[string][]*lg.Symbol
	ConstructorSorts map[string]lg.Sort
	SortConstructors map[string][]*lg.Symbol
	GhostSorts       map[string]bool
	SortOrder        []string
	SymbolOrder      []*lg.Symbol
	Variants         map[string][]lg.Sort // sort name → variant sorts
	Supertypes       map[string][]lg.Sort
	FiniteSorts      map[string]bool

	// Interpretations and natives
	Interps           map[string][]ast.Node // type name → labeled interps
	Natives           []ast.Node
	NativeDefinitions []*ast.LabeledFormula
	NativeTypes       map[string]*ast.NativeType // sort name → NativeType

	// Properties and proofs
	Progress     []interface{}
	Rely         []lg.Expr
	MixOrd       []ast.Node
	Privates     map[string]bool
	Proofs       []ProofEntry
	Named        []NamedEntry
	Subgoals     []SubgoalEntry
	ConjActions  map[string][]string
	ConjSubgoals []*ast.LabeledFormula

	// Parameters
	Params        []*lg.Symbol
	ParamDefaults []ast.Node // AST node (def.Rhs) or nil for "no default"; Python stores raw AST

	// Other
	Aliases       map[string]string // name → name
	BeforeExport  map[string]Action
	Attributes    map[string]interface{}
	ExtPreconds   map[string]lg.Expr
	ConceptSpaces []ConceptSpace

	// unused. placeholder?
	//AbstrPreds    []interface{}

	Logics []string
	Macros map[string]*ast.Definition // macro name → definition

	// CompCfg holds the per-session compiler config.
	CompCfg *CompilerConfig

	// CompileActionBodyFn is a callback to compile an AST node as an action body.
	// Set by the compiler after compilation. Used for runtime macro expansion
	// in InstantiateAction.IntUpdate. Corresponds to Python's im.compile() call
	// in InstantiateAction.int_update (ivy_actions.py:755).
	CompileActionBodyFn func(node ast.Node) (Action, error)

	// AdmitDefinitionFn is injected by the driver to call proof.ProofChecker.AdmitDefinition
	// without creating a compiler→proof import cycle. Python: prover.admit_definition(d, pmap[d.id])
	AdmitDefinitionFn func(defn *ast.LabeledFormula, proof ast.Node) error

	// Signature (captured at module creation time)
	Sig *il.Sig

	// InitCond is the initial condition clauses, computed from LabeledInits
	// and initializer actions. Corresponds to Python's module.init_cond.
	InitCond *co.Clauses

	// Instantiator is a function that instantiates non-EPR definitions
	// with ground terms. Set by TheoryContext. Corresponds to Python's
	// lu.instantiator / ModuleTheoryContext.__call__.
	Instantiator func(groundTerms []lg.Expr) *co.Clauses

	// Name is the module name, typically the source filename without extension.
	// Corresponds to Python's module.name.
	Name string

	// Theory is the cached background theory, set by UpdateTheory.
	// Corresponds to Python's self.theory (ivy_module.py:117).
	Theory *co.Clauses

	// prevModule is used by Enter/Exit for context management.
	prevModule *Module
	// oldSig is saved by Enter() and restored by Exit().
	// Corresponds to Python's self.old_sig (ivy_module.py:97).
	oldSig *il.Sig
}

// NamedAction pairs a name with an action.
type NamedAction struct {
	Name   string
	Action Action
}

// ProofEntry pairs a labeled formula with a proof.
type ProofEntry struct {
	Formula *ast.LabeledFormula
	Proof   ast.Node
}

// NamedEntry pairs a labeled formula with a name atom.
type NamedEntry struct {
	Formula *ast.LabeledFormula
	Name    lg.Expr
}

// SubgoalEntry pairs a formula with its subgoals.
type SubgoalEntry struct {
	Formula  *ast.LabeledFormula
	Subgoals []*ast.LabeledFormula
}

// IsolateInfo holds metadata about an isolate for user consumption.
type IsolateInfo struct {
	Implementations []MixinTriple
	Monitors        []MixinTriple
}

// ConceptSpace pairs a label expression with a body expression for a concept space.
// Corresponds to Python's (label, body) tuples in module.concept_spaces.
type ConceptSpace struct {
	Label lg.Expr
	Body  lg.Expr
}

// Instantiation pairs a schema with the AST node that instantiates it.
// Python stores these as (schema, inst) tuples in module.instantiations.
type Instantiation struct {
	Schema ast.Node // the schema definition (from Module.Schemata)
	Inst   ast.Node // the instantiation AST node
}

// MixinTriple holds mixer, mixee, and action for a mixin implementation.
type MixinTriple struct {
	Mixer  string
	Mixee  string
	Action Action
}

// New creates a fresh empty module with a new signature.
func New() *Module {
	m := &Module{
		Cfg: NewConfig(),
	}
	m.Clear()
	return m
}

// NewWithSig creates a module using the given signature.
func NewWithSig(sig *il.Sig) *Module {
	m := New()
	m.Sig = sig
	return m
}

// Clear resets the module to its initial empty state.
func (m *Module) Clear() {
	m.AllRelations = nil
	m.Definitions = nil
	m.LabeledAxioms = nil
	m.LabeledProps = nil
	m.LabeledInits = nil
	m.LabeledConjs = nil
	m.Assertions = nil
	m.AssumedInvs = nil
	m.Postconds = make(map[string][]*ast.LabeledFormula)
	m.Relations = make(map[string]lg.Sort)
	m.Functions = make(map[string]lg.Sort)
	m.Actions = make(map[string]Action)
	m.Mixins = make(map[string][]MixinDef)
	m.PublicActions = make(map[string]bool)
	m.Predicates = make(map[string]ast.Node)
	m.Initializers = nil
	m.InitialActions = nil
	m.Hierarchy = make(map[string]map[string]bool)
	m.Updates = nil
	m.Schemata = make(map[string]ast.Node)
	m.Theorems = make(map[string]ast.Node)
	m.Instantiations = nil
	m.Isolates = make(map[string]*ast.IsolateDef)
	m.IsolateInfo = nil
	m.IsolateProofs = make(map[string]ast.Node)
	m.IsolateProof = nil
	m.Exports = nil
	m.Imports = nil
	m.Delegates = nil
	m.DestructorSorts = make(map[string]lg.Sort)
	m.SortDestructors = make(map[string][]*lg.Symbol)
	m.ConstructorSorts = make(map[string]lg.Sort)
	m.SortConstructors = make(map[string][]*lg.Symbol)
	m.GhostSorts = make(map[string]bool)
	m.SortOrder = nil
	m.SymbolOrder = nil
	m.Variants = make(map[string][]lg.Sort)
	m.Supertypes = make(map[string][]lg.Sort)
	m.FiniteSorts = make(map[string]bool)
	m.Interps = make(map[string][]ast.Node)
	m.Natives = nil
	m.NativeDefinitions = nil
	m.NativeTypes = make(map[string]*ast.NativeType)
	m.Progress = nil
	m.Rely = nil
	m.MixOrd = nil
	m.Privates = make(map[string]bool)
	m.Proofs = nil
	m.Named = nil
	m.Subgoals = nil
	m.ConjActions = make(map[string][]string)
	m.ConjSubgoals = nil
	m.Params = nil
	m.ParamDefaults = nil
	m.Aliases = make(map[string]string)
	m.BeforeExport = make(map[string]Action)
	m.Attributes = make(map[string]interface{})
	m.ExtPreconds = make(map[string]lg.Expr)
	m.ConceptSpaces = nil
	//m.AbstrPreds = nil
	m.Logics = nil
	m.Macros = make(map[string]*ast.Definition)
	m.Sig = il.NewSig()
	// Python line 35: self.init_cond = lu.true_clauses()
	m.InitCond = co.TrueClauses(nil)
}

// Copy creates a semi-shallow copy of the module.
func (m *Module) Copy() *Module {
	xtracer.Trace("module.Copy ENTER actions=%d isolates=%d", len(m.Actions), len(m.Isolates))
	defer func() { xtracer.Trace("module.Copy EXIT") }()
	c := New()

	// Copy slices (shallow)
	c.AllRelations = copyNodeSlice(m.AllRelations)
	c.Definitions = copyLFSlice(m.Definitions)
	c.LabeledAxioms = copyLFSlice(m.LabeledAxioms)
	c.LabeledProps = copyLFSlice(m.LabeledProps)
	c.LabeledInits = copyLFSlice(m.LabeledInits)
	c.LabeledConjs = copyLFSlice(m.LabeledConjs)
	c.Assertions = copyLFSlice(m.Assertions)
	c.AssumedInvs = copyLFSlice(m.AssumedInvs)
	c.Progress = append([]interface{}{}, m.Progress...)
	c.Rely = append([]lg.Expr{}, m.Rely...)
	c.MixOrd = append([]ast.Node{}, m.MixOrd...)
	c.Natives = append([]ast.Node{}, m.Natives...)
	c.NativeDefinitions = append([]*ast.LabeledFormula{}, m.NativeDefinitions...)
	c.InitialActions = append([]Action{}, m.InitialActions...)
	c.Initializers = append([]NamedAction{}, m.Initializers...)
	c.SortOrder = append([]string{}, m.SortOrder...)
	c.Logics = append([]string{}, m.Logics...)
	c.Exports = append([]Exporter{}, m.Exports...)
	c.Imports = append([]ast.Node{}, m.Imports...)
	c.Delegates = append([]Delegator{}, m.Delegates...)

	// Copy params
	c.Params = make([]*lg.Symbol, len(m.Params))
	copy(c.Params, m.Params)
	c.ParamDefaults = append([]ast.Node{}, m.ParamDefaults...)
	c.SymbolOrder = make([]*lg.Symbol, len(m.SymbolOrder))
	copy(c.SymbolOrder, m.SymbolOrder)

	// Copy maps
	c.Postconds = copyMapLF(m.Postconds)
	c.Relations = copyMapSort(m.Relations)
	c.Functions = copyMapSort(m.Functions)
	c.Actions = copyMapAction(m.Actions)
	c.Schemata = copyMapNode(m.Schemata)
	c.Theorems = copyMapNode(m.Theorems)
	c.Isolates = make(map[string]*ast.IsolateDef, len(m.Isolates))
	for k, v := range m.Isolates {
		c.Isolates[k] = v
	}
	c.Predicates = copyMapNode(m.Predicates)
	c.DestructorSorts = copyMapSort(m.DestructorSorts)
	c.ConstructorSorts = copyMapSort(m.ConstructorSorts)
	c.NativeTypes = copyMapNativeType(m.NativeTypes)
	c.Aliases = copyMapStr(m.Aliases)
	c.BeforeExport = copyMapAction(m.BeforeExport)
	c.Attributes = copyMapIface(m.Attributes)
	c.PublicActions = copyMapBool(m.PublicActions)
	c.GhostSorts = copyMapBool(m.GhostSorts)
	c.Privates = copyMapBool(m.Privates)
	c.FiniteSorts = copyMapBool(m.FiniteSorts)

	// Copy maps missing from original port (Python copies ALL via dict iteration).
	// SortDestructors: map[string][]*lg.Symbol
	c.SortDestructors = make(map[string][]*lg.Symbol, len(m.SortDestructors))
	for k, v := range m.SortDestructors {
		c.SortDestructors[k] = append([]*lg.Symbol{}, v...)
	}
	// SortConstructors: map[string][]*lg.Symbol
	c.SortConstructors = make(map[string][]*lg.Symbol, len(m.SortConstructors))
	for k, v := range m.SortConstructors {
		c.SortConstructors[k] = append([]*lg.Symbol{}, v...)
	}
	// Variants: map[string][]lg.Sort
	c.Variants = make(map[string][]lg.Sort, len(m.Variants))
	for k, v := range m.Variants {
		c.Variants[k] = append([]lg.Sort{}, v...)
	}
	// Supertypes: map[string][]lg.Sort
	c.Supertypes = make(map[string][]lg.Sort, len(m.Supertypes))
	for k, v := range m.Supertypes {
		c.Supertypes[k] = append([]lg.Sort{}, v...)
	}
	// ExtPreconds: map[string]lg.Expr
	c.ExtPreconds = make(map[string]lg.Expr, len(m.ExtPreconds))
	for k, v := range m.ExtPreconds {
		c.ExtPreconds[k] = v
	}
	// ConjActions: map[string][]string
	c.ConjActions = make(map[string][]string, len(m.ConjActions))
	for k, v := range m.ConjActions {
		c.ConjActions[k] = append([]string{}, v...)
	}
	// IsolateProofs: map[string]ast.Node
	c.IsolateProofs = copyMapNode(m.IsolateProofs)

	// Copy hierarchy
	c.Hierarchy = make(map[string]map[string]bool, len(m.Hierarchy))
	for k, v := range m.Hierarchy {
		c.Hierarchy[k] = copyMapBool(v)
	}

	// Copy mixins
	c.Mixins = make(map[string][]MixinDef, len(m.Mixins))
	for k, v := range m.Mixins {
		c.Mixins[k] = append([]MixinDef{}, v...)
	}

	// Copy interps
	c.Interps = make(map[string][]ast.Node, len(m.Interps))
	for k, v := range m.Interps {
		c.Interps[k] = append([]ast.Node{}, v...)
	}

	// Copy signature (deep)
	c.Sig = m.Sig.Copy()

	return c
}

// AddToHierarchy adds a dotted name to the hierarchy tree.
func (m *Module) AddToHierarchy(name string) {
	if idx := strings.LastIndex(name, iu.ComposeCharacter); idx >= 0 {
		pref := name[:idx]
		suff := name[idx+len(iu.ComposeCharacter):]
		m.AddToHierarchy(pref)
		if m.Hierarchy[pref] == nil {
			m.Hierarchy[pref] = make(map[string]bool)
		}
		m.Hierarchy[pref][suff] = true
	} else {
		if m.Hierarchy["this"] == nil {
			m.Hierarchy["this"] = make(map[string]bool)
		}
		m.Hierarchy["this"][name] = true
	}
}

// AddObject adds an object name to the hierarchy.
func (m *Module) AddObject(name string) {
	if m.Hierarchy[name] == nil {
		m.Hierarchy[name] = make(map[string]bool)
	}
}

// FindAction looks up an action by name.
func (m *Module) FindAction(name string) (Action, bool) {
	a, ok := m.Actions[name]
	return a, ok
}

// IsVariant returns true if rsort is a variant of lsort.
func (m *Module) IsVariant(lsort, rsort lg.Sort) bool {
	lname := il.SortName(lsort)
	variants, ok := m.Variants[lname]
	if !ok {
		return false
	}
	for _, v := range variants {
		if lg.SortEqual(v, rsort) {
			return true
		}
	}
	return false
}

// VariantIndex returns the index of variant rsort within lsort's variants.
// Returns -1 if not found.
func (m *Module) VariantIndex(lsort, rsort lg.Sort) int {
	lname := il.SortName(lsort)
	variants, ok := m.Variants[lname]
	if !ok {
		return -1
	}
	for i, v := range variants {
		if lg.SortEqual(v, rsort) {
			return i
		}
	}
	return -1
}

// SortCard returns an estimate of the cardinality of a sort, or -1 if unknown.
// Corresponds to Python's Module.sort_card (ivy_module.py:217-223):
//
//	attr = iu.compose_names(sort.name, 'cardinality')
//	if attr in self.attributes:
//	    return int(self.attributes[attr].rep)
func (m *Module) SortCard(sort lg.Sort) int {
	if il.IsFunctionSort(sort) {
		return -1
	}
	name := il.SortName(sort)
	attr := iu.ComposeNames(name, "cardinality")
	if val, ok := m.Attributes[attr]; ok {
		// Python: int(self.attributes[attr].rep)
		// The attribute value is an AST node. Use Sexp() for structural
		// equivalence when the value is a lg.Expr; fall back to Relname()
		// for ast.Node types, mirroring Python's .rep access.
		var rep string
		switch v := val.(type) {
		case lg.Expr:
			rep = string(v.Sexp())
		case interface{ Relname() string }:
			rep = v.Relname()
		case string:
			rep = v
		default:
			return -1
		}
		var n int
		if _, err := fmt.Sscanf(rep, "%d", &n); err == nil {
			return n
		}
		return -1
	}
	return SortCardDefault(sort)
}

// SortCardDefault returns the cardinality based on sort type.
func SortCardDefault(sort lg.Sort) int {
	if es, ok := sort.(*lg.EnumeratedSort); ok {
		return es.Card()
	}
	return -1
}

// CallGraph builds the action call graph: called → callers.
// Corresponds to Python's Module.call_graph (ivy_module.py:279-284).
func (m *Module) CallGraph() map[string][]string {
	callgraph := make(map[string][]string)
	for actname, action := range m.Actions {
		for _, calledName := range action.IterCalls() {
			callgraph[calledName] = append(callgraph[calledName], actname)
		}
	}
	return callgraph
}

// SortDependencies returns sort names that the given sort depends on.
func (m *Module) SortDependencies(sortName string, withVariants bool) []string {
	if destrs, ok := m.SortDestructors[sortName]; ok {
		var deps []string
		for _, destr := range destrs {
			if fs, ok := destr.CSort.(*lg.FunctionSort); ok {
				dom := fs.Domain()
				for _, d := range dom[1:] {
					deps = append(deps, il.SortName(d))
				}
				deps = append(deps, il.SortName(fs.Range()))
			}
		}
		return deps
	}
	// NativeTypes branch (Python ivy_module.py:397-400):
	//   if sortname in mod.native_types:
	//       t = mod.native_types[sortname]
	//       if isinstance(t, ivy_ast.NativeType):
	//           return [s.rep for s in t.args[1:] if s.rep in mod.sig.sorts]
	if nt, ok := m.NativeTypes[sortName]; ok && nt != nil {
		if len(nt.Elems) > 1 {
			var deps []string
			for _, elem := range nt.Elems[1:] {
				var rep string
				switch e := elem.(type) {
				case interface{ Relname() string }:
					rep = e.Relname()
				}
				if rep != "" {
					if _, inSig := m.Sig.Sorts[rep]; inSig {
						deps = append(deps, rep)
					}
				}
			}
			return deps
		}
	}
	if withVariants {
		if vs, ok := m.Variants[sortName]; ok {
			deps := make([]string, len(vs))
			for i, v := range vs {
				deps[i] = il.SortName(v)
			}
			return deps
		}
	}
	return nil
}

// GetLogics returns the logic names for this module, implementing
// the full Python logics() fallback chain (ivy_module.py:355-358):
//  1. module.logics (per-module override, set by compiler)
//  2. Config.CompleteLogic (CLI parameter, Python: param_logic)
//  3. il.DefaultLogics (= ["epr"])
func (m *Module) GetLogics() []string {
	if len(m.Logics) > 0 {
		return m.Logics
	}
	if m.Cfg != nil && m.Cfg.CompleteLogic != "" {
		return strings.Split(m.Cfg.CompleteLogic, ",")
	}
	return il.DefaultLogics
}

// --- String representation ---

func (m *Module) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Module: %d axioms, %d conjs, %d actions, %d sorts\n",
		len(m.LabeledAxioms), len(m.LabeledConjs),
		len(m.Actions), len(m.Sig.Sorts))
	return b.String()
}

// --- helper copy functions ---

func copyNodeSlice(s []lg.Expr) []lg.Expr {
	if s == nil {
		return nil
	}
	c := make([]lg.Expr, len(s))
	copy(c, s)
	return c
}

func copyLFSlice(s []*ast.LabeledFormula) []*ast.LabeledFormula {
	if s == nil {
		return nil
	}
	c := make([]*ast.LabeledFormula, len(s))
	copy(c, s)
	return c
}

func copyMapLF(m map[string][]*ast.LabeledFormula) map[string][]*ast.LabeledFormula {
	c := make(map[string][]*ast.LabeledFormula, len(m))
	for k, v := range m {
		c[k] = copyLFSlice(v)
	}
	return c
}

func copyMapSort(m map[string]lg.Sort) map[string]lg.Sort {
	c := make(map[string]lg.Sort, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

func copyMapIface(m map[string]interface{}) map[string]interface{} {
	c := make(map[string]interface{}, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

func copyMapAction(m map[string]Action) map[string]Action {
	c := make(map[string]Action, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

func copyMapNode(m map[string]ast.Node) map[string]ast.Node {
	c := make(map[string]ast.Node, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

func copyMapNativeType(m map[string]*ast.NativeType) map[string]*ast.NativeType {
	c := make(map[string]*ast.NativeType, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

func copyMapStr(m map[string]string) map[string]string {
	c := make(map[string]string, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

func copyMapBool(m map[string]bool) map[string]bool {
	c := make(map[string]bool, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

// UpdateConjs generates concept spaces from the labeled conjectures.
// For each conjecture, it creates a named concept space suitable for
// the UI's counterexample-guided abstraction refinement loop.
//
// Corresponds to Python Module.update_conjs (ivy_module.py:268-277):
//
//	for i,cax in enumerate(mod.labeled_conjs):
//	    fmla = cax.formula
//	    csname = 'conjecture:'+ str(i)
//	    variables = list(lu.used_variables_ast(fmla))
//	    sort = il.RelationSort([v.sort for v in variables])
//	    sym = il.Symbol(csname,sort)
//	    space = ics.NamedSpace(il.Literal(0,fmla))
//	    mod.concept_spaces.append((sym(*variables),space))
func (m *Module) UpdateConjs() {
	for i, cax := range m.LabeledConjs {
		if cax == nil || cax.Formula == nil {
			continue
		}
		fmla, ok := cax.Formula.(lg.Expr)
		if !ok {
			continue
		}
		csname := fmt.Sprintf("conjecture:%d", i)

		// Collect used variables from the formula.
		varMap := co.UsedVariablesAST(fmla)
		var variables []*lg.Variable
		for _, v := range varMap {
			if vr, ok := v.(*lg.Variable); ok {
				variables = append(variables, vr)
			}
		}

		// Build sort: RelationSort([v.sort for v in variables])
		sorts := make([]lg.Sort, len(variables))
		for j, v := range variables {
			sorts[j] = v.VSort
		}
		symSort := il.RelationSort(sorts)
		sym := lg.NewSymbol(csname, symSort)

		// Build label: sym(*variables)
		var label lg.Expr
		if len(variables) > 0 {
			varExprs := make([]lg.Expr, len(variables))
			for j, v := range variables {
				varExprs[j] = v
			}
			label = &lg.Apply{Func: sym, Terms: varExprs}
		} else {
			label = sym
		}

		// Build space: NamedSpace(il.Literal(0, fmla))
		space := &il.Literal{Polarity: 0, Atom: fmla}

		m.ConceptSpaces = append(m.ConceptSpaces, ConceptSpace{Label: label, Body: space})
	}
}
