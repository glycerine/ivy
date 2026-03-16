// Package module provides the Module type, which holds all declarations,
// axioms, actions, and other state for an Ivy module.
//
// This corresponds to Python's ivy_module.py.
package module

import (
	"fmt"
	"strings"

	co "github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
)

// Module holds all the definitions and declarations in an Ivy module.
type Module struct {
	// Declarations
	AllRelations  []lg.Node // base and derived relations in declaration order
	Definitions   []*LabeledFormula
	LabeledAxioms []*LabeledFormula
	LabeledProps  []*LabeledFormula
	LabeledInits  []*LabeledFormula
	LabeledConjs  []*LabeledFormula // conjectures
	Assertions    []*LabeledFormula
	Postconds     map[string][]*LabeledFormula // action name → postconditions
	AssumedInvs   []*LabeledFormula            // assumed invariants

	// Relations and functions
	Relations map[string]lg.Sort
	Functions map[string]lg.Sort

	// Actions and mixins
	Actions       map[string]interface{} // action name → Action (interface for now)
	Mixins        map[string][]interface{}
	PublicActions map[string]bool
	Predicates    map[string]interface{}
	Initializers  []NamedAction
	InitialActions []interface{}

	// Module structure
	Hierarchy      map[string]map[string]bool // parent → children
	Updates        []interface{}
	Schemata       map[string]interface{}
	Theorems       map[string]interface{}
	Instantiations []interface{}

	// Isolates
	Isolates     map[string]interface{}
	IsolateInfo  *IsolateInfo
	IsolateProofs map[string]interface{}
	IsolateProof interface{}

	// Exports and imports
	Exports   []interface{}
	Imports   []interface{}
	Delegates []interface{}

	// Sorts and destructors
	DestructorSorts   map[string]lg.Sort
	SortDestructors   map[string][]*lg.Const
	ConstructorSorts  map[string]lg.Sort
	SortConstructors  map[string][]*lg.Const
	GhostSorts        map[string]bool
	SortOrder         []string
	SymbolOrder       []*lg.Const
	Variants          map[string][]lg.Sort // sort name → variant sorts
	Supertypes        map[string][]lg.Sort
	FiniteSorts       map[string]bool

	// Interpretations and natives
	Interps           map[string][]interface{} // type name → labeled interps
	Natives           []interface{}
	NativeDefinitions []interface{}
	NativeTypes       map[string]interface{} // sort name → NativeType

	// Properties and proofs
	Progress       []interface{}
	Rely           []interface{}
	MixOrd         []interface{}
	Privates       map[string]bool
	Proofs         []ProofEntry
	Named          []NamedEntry
	Subgoals       []SubgoalEntry
	ConjActions    map[string][]string
	ConjSubgoals   []*LabeledFormula

	// Parameters
	Params        []*lg.Const
	ParamDefaults []string // may contain empty strings for "no default"

	// Other
	Aliases       map[string]string // name → name
	BeforeExport  map[string]interface{}
	Attributes    map[string]interface{}
	ExtPreconds   map[string]lg.Node
	ConceptSpaces []interface{}
	AbstrPreds    []interface{}
	Logics        []string

	// Signature (captured at module creation time)
	Sig *il.Sig

	// InitCond is the initial condition clauses, computed from LabeledInits
	// and initializer actions. Corresponds to Python's module.init_cond.
	InitCond *co.Clauses

	// prevModule is used by Enter/Exit for context management.
	prevModule *Module
}

// LabeledFormula is a formula with an optional label and metadata.
type LabeledFormula struct {
	Label      lg.Node // may be nil
	Formula    lg.Node
	Lineno     int
	Temporal   bool
	ID         int64
	Explicit   bool
	Assumed    bool
	Unprovable bool
}

// NamedAction pairs a name with an action.
type NamedAction struct {
	Name   string
	Action interface{}
}

// ProofEntry pairs a labeled formula with a proof.
type ProofEntry struct {
	Formula *LabeledFormula
	Proof   interface{}
}

// NamedEntry pairs a labeled formula with a name atom.
type NamedEntry struct {
	Formula *LabeledFormula
	Name    lg.Node
}

// SubgoalEntry pairs a formula with its subgoals.
type SubgoalEntry struct {
	Formula  *LabeledFormula
	Subgoals []*LabeledFormula
}

// IsolateInfo holds metadata about an isolate for user consumption.
type IsolateInfo struct {
	Implementations []MixinTriple
	Monitors        []MixinTriple
}

// MixinTriple holds mixer, mixee, and action for a mixin implementation.
type MixinTriple struct {
	Mixer  string
	Mixee  string
	Action interface{}
}

// New creates a fresh empty module with a new signature.
func New() *Module {
	m := &Module{}
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
	m.Postconds = make(map[string][]*LabeledFormula)
	m.Relations = make(map[string]lg.Sort)
	m.Functions = make(map[string]lg.Sort)
	m.Actions = make(map[string]interface{})
	m.Mixins = make(map[string][]interface{})
	m.PublicActions = make(map[string]bool)
	m.Predicates = make(map[string]interface{})
	m.Initializers = nil
	m.InitialActions = nil
	m.Hierarchy = make(map[string]map[string]bool)
	m.Updates = nil
	m.Schemata = make(map[string]interface{})
	m.Theorems = make(map[string]interface{})
	m.Instantiations = nil
	m.Isolates = make(map[string]interface{})
	m.IsolateInfo = nil
	m.IsolateProofs = make(map[string]interface{})
	m.IsolateProof = nil
	m.Exports = nil
	m.Imports = nil
	m.Delegates = nil
	m.DestructorSorts = make(map[string]lg.Sort)
	m.SortDestructors = make(map[string][]*lg.Const)
	m.ConstructorSorts = make(map[string]lg.Sort)
	m.SortConstructors = make(map[string][]*lg.Const)
	m.GhostSorts = make(map[string]bool)
	m.SortOrder = nil
	m.SymbolOrder = nil
	m.Variants = make(map[string][]lg.Sort)
	m.Supertypes = make(map[string][]lg.Sort)
	m.FiniteSorts = make(map[string]bool)
	m.Interps = make(map[string][]interface{})
	m.Natives = nil
	m.NativeDefinitions = nil
	m.NativeTypes = make(map[string]interface{})
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
	m.BeforeExport = make(map[string]interface{})
	m.Attributes = make(map[string]interface{})
	m.ExtPreconds = make(map[string]lg.Node)
	m.ConceptSpaces = nil
	m.AbstrPreds = nil
	m.Logics = nil
	m.Sig = il.NewSig()
}

// Copy creates a semi-shallow copy of the module.
func (m *Module) Copy() *Module {
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
	c.Rely = append([]interface{}{}, m.Rely...)
	c.MixOrd = append([]interface{}{}, m.MixOrd...)
	c.Natives = append([]interface{}{}, m.Natives...)
	c.NativeDefinitions = append([]interface{}{}, m.NativeDefinitions...)
	c.Initializers = append([]NamedAction{}, m.Initializers...)
	c.SortOrder = append([]string{}, m.SortOrder...)
	c.Logics = append([]string{}, m.Logics...)
	c.Exports = append([]interface{}{}, m.Exports...)
	c.Imports = append([]interface{}{}, m.Imports...)
	c.Delegates = append([]interface{}{}, m.Delegates...)

	// Copy params
	c.Params = make([]*lg.Const, len(m.Params))
	copy(c.Params, m.Params)
	c.ParamDefaults = append([]string{}, m.ParamDefaults...)
	c.SymbolOrder = make([]*lg.Const, len(m.SymbolOrder))
	copy(c.SymbolOrder, m.SymbolOrder)

	// Copy maps
	c.Postconds = copyMapLF(m.Postconds)
	c.Relations = copyMapSort(m.Relations)
	c.Functions = copyMapSort(m.Functions)
	c.Actions = copyMapIface(m.Actions)
	c.Schemata = copyMapIface(m.Schemata)
	c.Theorems = copyMapIface(m.Theorems)
	c.Isolates = copyMapIface(m.Isolates)
	c.Predicates = copyMapIface(m.Predicates)
	c.DestructorSorts = copyMapSort(m.DestructorSorts)
	c.ConstructorSorts = copyMapSort(m.ConstructorSorts)
	c.NativeTypes = copyMapIface(m.NativeTypes)
	c.Aliases = copyMapStr(m.Aliases)
	c.BeforeExport = copyMapIface(m.BeforeExport)
	c.Attributes = copyMapIface(m.Attributes)
	c.PublicActions = copyMapBool(m.PublicActions)
	c.GhostSorts = copyMapBool(m.GhostSorts)
	c.Privates = copyMapBool(m.Privates)
	c.FiniteSorts = copyMapBool(m.FiniteSorts)

	// Copy hierarchy
	c.Hierarchy = make(map[string]map[string]bool, len(m.Hierarchy))
	for k, v := range m.Hierarchy {
		c.Hierarchy[k] = copyMapBool(v)
	}

	// Copy mixins
	c.Mixins = make(map[string][]interface{}, len(m.Mixins))
	for k, v := range m.Mixins {
		c.Mixins[k] = append([]interface{}{}, v...)
	}

	// Copy interps
	c.Interps = make(map[string][]interface{}, len(m.Interps))
	for k, v := range m.Interps {
		c.Interps[k] = append([]interface{}{}, v...)
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
func (m *Module) FindAction(name string) (interface{}, bool) {
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
func (m *Module) SortCard(sort lg.Sort) int {
	if il.IsFunctionSort(sort) {
		return -1
	}
	name := il.SortName(sort)
	attr := iu.ComposeNames(name, "cardinality")
	if val, ok := m.Attributes[attr]; ok {
		// Matches Python: im.module.attributes[attr] returns a string value
		// that can be parsed as an integer cardinality bound.
		if s, ok2 := val.(string); ok2 {
			var n int
			if _, err := fmt.Sscanf(s, "%d", &n); err == nil {
				return n
			}
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
func (m *Module) CallGraph() map[string][]string {
	// Placeholder — requires action iteration which needs the Action type.
	return make(map[string][]string)
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

// GetLogics returns the logic names set for this module.
// If no logics have been set, returns the default logics (["epr"]).
func (m *Module) GetLogics() []string {
	if len(m.Logics) == 0 {
		return []string{"epr"}
	}
	return m.Logics
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

func copyNodeSlice(s []lg.Node) []lg.Node {
	if s == nil {
		return nil
	}
	c := make([]lg.Node, len(s))
	copy(c, s)
	return c
}

func copyLFSlice(s []*LabeledFormula) []*LabeledFormula {
	if s == nil {
		return nil
	}
	c := make([]*LabeledFormula, len(s))
	copy(c, s)
	return c
}

func copyMapLF(m map[string][]*LabeledFormula) map[string][]*LabeledFormula {
	c := make(map[string][]*LabeledFormula, len(m))
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
