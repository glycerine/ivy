// Package module provides the Module type, which holds all declarations,
// axioms, actions, and other state for an Ivy module.
//
// This corresponds to Python's ivy_module.py.
package goivy

import (
	"fmt"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/glycerine/ivy/goivy/smt"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// Module holds all the definitions and declarations in an Ivy module.
type Module struct {
	Cfg *Config

	// Declarations
	AllRelations  []Expr // base and derived relations in declaration order
	Definitions   []*LabeledFormula
	LabeledAxioms []*LabeledFormula
	LabeledProps  []*LabeledFormula
	LabeledInits  []*LabeledFormula
	LabeledConjs  []*LabeledFormula // conjectures
	Assertions    []*LabeledFormula
	Postconds     map[string][]*LabeledFormula // action name → postconditions
	AssumedInvs   []*LabeledFormula            // assumed invariants

	// Relations and functions
	Relations *InsMap[NodeKey, Sort]
	Functions *InsMap[NodeKey, Sort]

	// Actions and mixins
	Actions        *InsMap[string, Action]
	Mixins         *InsMap[string, []MixinDef]
	PublicActions  *InsMap[string, bool]
	Predicates     map[string]Node
	Initializers   []NamedAction
	InitialActions []Action

	// Module structure
	Hierarchy      *InsMap[string, *InsMap[string, bool]] // parent → children
	Updates        []interface{}
	Schemata       *InsMap[string, Node]
	Theorems       map[string]Node
	Instantiations []LogicInstantiation

	// Isolates
	Isolates      map[string]*IsolateDef
	IsolateInfo   *IsolateInfo
	IsolateProofs map[string]Node
	IsolateProof  Node

	// Exports and imports
	Exports   []Exporter
	Imports   []Node
	Delegates []Delegator

	// Sorts and destructors
	DestructorSorts  map[string]Sort
	SortDestructors  *InsMap[string, []*Const]
	ConstructorSorts map[string]Sort
	SortConstructors *InsMap[string, []*Const]
	GhostSorts       map[string]bool
	SortOrder        []string
	SymbolOrder      []*Const
	Variants         map[string][]Sort // sort name → variant sorts
	Supertypes       *InsMap[string, Sort]
	FiniteSorts      map[string]bool

	// Interpretations and natives
	Interps           map[string][]Node // type name → labeled interps
	Natives           []Node
	NativeDefinitions []*LabeledFormula
	NativeTypes       map[string]*NativeType // sort name → NativeType

	// Properties and proofs
	Progress     []interface{}
	Rely         []Expr
	MixOrd       []Node
	Privates     map[string]bool
	VPrivates    map[string]bool // verified-private names (from isolate processing)
	Proofs       []ProofEntry
	Named        []NamedEntry
	Subgoals     []SubgoalEntry
	ConjActions  map[string][]string
	ConjSubgoals []*LabeledFormula

	// Parameters
	Params        []*Const
	ParamDefaults []Node // AST node (def.Rhs) or nil for "no default"; Python stores raw AST

	// Other
	Aliases        map[string]string // name → name
	BeforeExport   *InsMap[string, Action]
	Attributes     map[string]interface{}
	AttributeOrder []string
	ExtPreconds    map[string]Expr
	ConceptSpaces  []ConceptSpace

	AbstractionPredicates []interface{}

	Logics []string
	Macros map[string]*Definition // macro name → definition

	// SigMerkle is the rolling Merkle hash for compiler conformance auditing.
	// Matches Python's module-level sig_merkle in ivy_compiler.py.
	// Lives on Module so all Compiler instances share one chain per session.
	SigMerkle *MerkleState

	// CompCfg holds the per-session compiler config.
	CompCfg *CompilerConfig

	// Signature (captured at module creation time)
	Sig *Sig

	// InitCond is the initial condition clauses, computed from LabeledInits
	// and initializer actions. Corresponds to Python's module.init_cond.
	InitCond *Clauses

	// Instantiator is a function that instantiates non-EPR definitions
	// with ground terms. Set by TheoryContext. Corresponds to Python's
	// lu.instantiator / ModuleTheoryContext.__call__.
	Instantiator func(groundTerms []Expr) *Clauses

	// Name is the module name, typically the source filename without extension.
	// Corresponds to Python's module.name.
	Name string

	// Theory is the cached background theory, set by UpdateTheory.
	// Corresponds to Python's self.theory (ivy_module.py:117).
	Theory *Clauses

	// TraceHook is a diagnostic hook closure propagated from a goal's
	// LabeledFormula.TraceHook field. Mirrors Python's dynamically-attached
	// mod.trace_hook (ivy_check.py:406-407, 829-840).
	TraceHook TraceHookFn

	// AclCfg holds the loaded ACL config for unchecked property filtering.
	// Set by CheckModule from the OptUncheckedProps file, matching Python
	// ivy_acl.register_from_file (ivy_check.py:982-983).
	AclCfg *ACLConfig

	// prevModule is used by Enter/Exit for context management.
	prevModule *Module
	// oldSig is saved by Enter() and restored by Exit().
	// Corresponds to Python's self.old_sig (ivy_module.py:97).
	oldSig *Sig

	// z3SessionCache holds the Go equivalent of Python's per-module-context
	// z3_sorts/z3_predicates/z3_constants/z3_functions globals
	// (ivy_solver.py:252-260). All z3bridge.Solver instances created via
	// NewSolver(mod, ...) share this single cache, mirroring Python's
	// "z3.Solver() instances within a Module context share z3_sorts" rule.
	// Cleared by Module.Enter() (mirroring Python's clear() in __enter__).
	z3SessionCache *Z3SessionCache

	// z3SharedCtx holds the *Z3Context shared across module copies.
	// Python's _z3_check_counter is a process-global that persists across
	// all module copies; this field provides the equivalent in Go by
	// letting copied modules create fresh Z3SessionCaches that reuse the
	// same Z3Context (and its z3CheckCounter).
	//
	// Uses *z3CtxHolder (pointer to holder) so that Copy() shares the
	// same holder by pointer. When any copy's solver initializes the
	// Z3Context, all copies (past and future) see it through the holder.
	z3SharedCtx *z3CtxHolder
}

// z3CtxHolder is a shared container for the Z3Context pointer.
// All module copies from the same New() call share a single holder
// via pointer, so when any copy's solver initializes the Z3Context,
// all other copies see it. This matches Python's process-global
// _z3_check_counter which accumulates across all module copies.
type z3CtxHolder struct {
	ctx            *smt.Z3Context
	z3CheckCounter atomic.Int64
}

// NamedAction pairs a name with an action.
type NamedAction struct {
	Name   string
	Action Action
}

// ProofEntry pairs a labeled formula with a proof.
type ProofEntry struct {
	Formula *LabeledFormula
	Proof   Node
}

// NamedEntry pairs a labeled formula with a name atom.
type NamedEntry struct {
	Formula *LabeledFormula
	Name    Expr
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

// ConceptSpace pairs a label expression with a body expression for a concept space.
// Corresponds to Python's (label, body) tuples in module.concept_spaces.
type ConceptSpace struct {
	Label Expr
	Body  Expr
}

// Instantiation pairs a schema with the AST node that instantiates it.
// Python stores these as (schema, inst) tuples in module.instantiations.
type LogicInstantiation struct {
	Schema Node // the schema definition (from Module.Schemata)
	Inst   Node // the instantiation AST node
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
		Cfg:         NewConfig(),
		z3SharedCtx: &z3CtxHolder{},
	}
	m.CompCfg = NewCompilerConfig(m.Cfg)
	m.Clear()
	m.SigMerkle = &MerkleState{}
	return m
}

// NewWithSig creates a module using the given signature.
func NewWithSig(sig *Sig) *Module {
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

	// Python line 36: self.init_cond = lu.true_clauses()
	m.InitCond = TrueClauses(nil)
	m.Relations = NewInsMap[NodeKey, Sort]()
	m.Functions = NewInsMap[NodeKey, Sort]()
	m.Updates = nil
	m.Schemata = NewInsMap[string, Node]()
	m.Theorems = make(map[string]Node)

	m.Instantiations = nil
	m.ConceptSpaces = nil
	m.AbstractionPredicates = nil
	m.LabeledConjs = nil
	m.Postconds = make(map[string][]*LabeledFormula)
	m.Hierarchy = NewInsMap[string, *InsMap[string, bool]]()
	m.Actions = NewInsMap[string, Action]()
	m.Predicates = make(map[string]Node)
	m.Assertions = nil
	m.Mixins = NewInsMap[string, []MixinDef]()
	m.PublicActions = NewInsMap[string, bool]()
	m.Isolates = make(map[string]*IsolateDef)
	m.Exports = nil
	m.Imports = nil
	m.Delegates = nil
	m.Progress = nil
	m.Rely = nil
	m.MixOrd = nil

	m.DestructorSorts = make(map[string]Sort)
	m.SortDestructors = NewInsMap[string, []*Const]()
	m.ConstructorSorts = make(map[string]Sort)
	m.SortConstructors = NewInsMap[string, []*Const]()

	m.Privates = make(map[string]bool)
	m.Interps = make(map[string][]Node)
	m.Natives = nil
	m.NativeDefinitions = nil
	m.Initializers = nil
	m.InitialActions = nil
	m.Params = nil
	m.ParamDefaults = nil
	m.GhostSorts = make(map[string]bool)
	m.NativeTypes = make(map[string]*NativeType)
	m.SortOrder = nil
	m.SymbolOrder = nil
	m.Aliases = make(map[string]string)
	m.BeforeExport = NewInsMap[string, Action]()
	m.Attributes = make(map[string]interface{})
	m.AttributeOrder = nil
	m.Variants = make(map[string][]Sort)
	m.Supertypes = NewInsMap[string, Sort]()
	m.ExtPreconds = make(map[string]Expr)
	m.Proofs = nil

	m.Named = nil
	m.Subgoals = nil
	m.IsolateInfo = nil
	m.ConjActions = make(map[string][]string)
	m.ConjSubgoals = nil
	m.AssumedInvs = nil
	m.FiniteSorts = make(map[string]bool)
	m.IsolateProofs = make(map[string]Node)
	m.IsolateProof = nil

	m.Logics = nil
	if m.Cfg != nil && m.Cfg.IuCfg != nil {
		m.Sig = NewSigOn(m.Cfg.IuCfg)
	} else {
		m.Sig = NewSig()
	}
	// python does not clear macros. maybe Go should not either?
	// but clear is also used to initialize... hmm... add nil check?
	// python does not actually have macros on its module.
	//if m.Macros == nil {
	m.Macros = make(map[string]*Definition)
	//}
}

// SetAttribute stores a module attribute while preserving the first insertion
// order Python's dict gives to im.module.attributes.
func (m *Module) SetAttribute(name string, value interface{}) {
	if m.Attributes == nil {
		m.Attributes = make(map[string]interface{})
	}
	if _, ok := m.Attributes[name]; !ok {
		m.AttributeOrder = append(m.AttributeOrder, name)
	}
	m.Attributes[name] = value
}

// AttributeNames returns attribute keys in Python insertion order. Keys written
// directly to Attributes are appended in sorted order as a deterministic fallback.
func (m *Module) AttributeNames() []string {
	if len(m.Attributes) == 0 {
		return nil
	}
	names := make([]string, 0, len(m.Attributes))
	seen := make(map[string]bool, len(m.Attributes))
	for _, name := range m.AttributeOrder {
		if _, ok := m.Attributes[name]; ok && !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	if len(names) < len(m.Attributes) {
		missing := make([]string, 0, len(m.Attributes)-len(names))
		for name := range m.Attributes {
			if !seen[name] {
				missing = append(missing, name)
			}
		}
		sort.Strings(missing)
		names = append(names, missing...)
	}
	return names
}

// Copy creates a semi-shallow copy of the module.
func (m *Module) Copy() *Module {
	xtracer.Trace("module.Copy ENTER actions=%d isolates=%d", m.Actions.Len(), len(m.Isolates))
	c := New()
	*c.Cfg = *m.Cfg

	// defer after c is declared so we can report its counts
	defer func() { xtracer.Trace("module.Copy EXIT actions=%d isolates=%d", c.Actions.Len(), len(c.Isolates)) }()

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
	c.Rely = append([]Expr{}, m.Rely...)
	c.MixOrd = append([]Node{}, m.MixOrd...)
	c.Natives = append([]Node{}, m.Natives...)
	c.NativeDefinitions = append([]*LabeledFormula{}, m.NativeDefinitions...)
	c.Updates = append([]interface{}{}, m.Updates...)
	c.InitialActions = append([]Action{}, m.InitialActions...)
	c.Initializers = append([]NamedAction{}, m.Initializers...)
	c.SortOrder = append([]string{}, m.SortOrder...)
	c.Logics = append([]string{}, m.Logics...)
	c.Exports = append([]Exporter{}, m.Exports...)
	c.Imports = append([]Node{}, m.Imports...)
	c.Delegates = append([]Delegator{}, m.Delegates...)

	// Copy proofs, named, subgoals (Python copies all via __dict__ iteration)
	c.Proofs = append([]ProofEntry{}, m.Proofs...)
	c.Named = append([]NamedEntry{}, m.Named...)
	c.Subgoals = append([]SubgoalEntry{}, m.Subgoals...)
	c.ConjSubgoals = copyLFSlice(m.ConjSubgoals)
	c.ConceptSpaces = append([]ConceptSpace{}, m.ConceptSpaces...)
	c.Instantiations = append([]LogicInstantiation{}, m.Instantiations...)
	c.IsolateInfo = m.IsolateInfo
	c.IsolateProof = m.IsolateProof
	c.InitCond = m.InitCond
	c.Name = m.Name

	// Copy params
	c.Params = make([]*Const, len(m.Params))
	copy(c.Params, m.Params)
	c.ParamDefaults = append([]Node{}, m.ParamDefaults...)
	c.SymbolOrder = make([]*Const, len(m.SymbolOrder))
	copy(c.SymbolOrder, m.SymbolOrder)

	// Copy maps
	c.Postconds = copyMapLF(m.Postconds)
	c.Relations = NewInsMap[NodeKey, Sort]()
	for k, v := range m.Relations.All() {
		c.Relations.Set(k, v)
	}
	c.Functions = NewInsMap[NodeKey, Sort]()
	for k, v := range m.Functions.All() {
		c.Functions.Set(k, v)
	}
	c.Actions = NewInsMap[string, Action]()
	for k, v := range m.Actions.All() {
		c.Actions.Set(k, v)
	}
	c.Schemata = copyInsMapNode(m.Schemata)
	c.Theorems = copyMapNode(m.Theorems)
	c.Isolates = make(map[string]*IsolateDef, len(m.Isolates))
	for k, v := range m.Isolates {
		c.Isolates[k] = v
	}
	c.Predicates = copyMapNode(m.Predicates)
	c.DestructorSorts = copyMapSort(m.DestructorSorts)
	c.ConstructorSorts = copyMapSort(m.ConstructorSorts)
	c.NativeTypes = copyMapNativeType(m.NativeTypes)
	c.Aliases = copyMapStr(m.Aliases)
	c.BeforeExport = NewInsMap[string, Action]()
	for k, v := range m.BeforeExport.All() {
		c.BeforeExport.Set(k, v)
	}
	c.Attributes = copyMapIface(m.Attributes)
	c.AttributeOrder = append([]string{}, m.AttributeOrder...)
	c.PublicActions = NewInsMap[string, bool]()
	for k, v := range m.PublicActions.All() {
		c.PublicActions.Set(k, v)
	}
	c.GhostSorts = copyMapBool(m.GhostSorts)
	c.Privates = copyMapBool(m.Privates)
	c.FiniteSorts = copyMapBool(m.FiniteSorts)

	// Copy maps missing from original port (Python copies ALL via dict iteration).
	// SortDestructors: insertion-ordered map[string][]*lg.Const
	c.SortDestructors = NewInsMap[string, []*Const]()
	for k, v := range m.SortDestructors.All() {
		c.SortDestructors.Set(k, append([]*Const{}, v...))
	}
	// SortConstructors: insertion-ordered map[string][]*lg.Const
	c.SortConstructors = NewInsMap[string, []*Const]()
	for k, v := range m.SortConstructors.All() {
		c.SortConstructors.Set(k, append([]*Const{}, v...))
	}
	// Variants: map[string][]lg.Sort
	c.Variants = make(map[string][]Sort, len(m.Variants))
	for k, v := range m.Variants {
		c.Variants[k] = append([]Sort{}, v...)
	}
	// Supertypes: subtype sort name -> supertype sort.
	c.Supertypes = NewInsMap[string, Sort]()
	if m.Supertypes != nil {
		for k, v := range m.Supertypes.All() {
			c.Supertypes.Set(k, v)
		}
	}
	// ExtPreconds: map[string]lg.Expr
	c.ExtPreconds = make(map[string]Expr, len(m.ExtPreconds))
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
	// VPrivates: map[string]bool
	if m.VPrivates != nil {
		c.VPrivates = copyMapBool(m.VPrivates)
	}
	// Macros: map[string]*ast.Definition
	c.Macros = make(map[string]*Definition, len(m.Macros))
	for k, v := range m.Macros {
		c.Macros[k] = v
	}

	// Copy hierarchy
	c.Hierarchy = NewInsMap[string, *InsMap[string, bool]]()
	for k, v := range m.Hierarchy.All() {
		inner := NewInsMap[string, bool]()
		for ik, iv := range v.All() {
			inner.Set(ik, iv)
		}
		c.Hierarchy.Set(k, inner)
	}

	// Copy mixins
	c.Mixins = NewInsMap[string, []MixinDef]()
	for k, v := range m.Mixins.All() {
		c.Mixins.Set(k, append([]MixinDef{}, v...))
	}

	// Copy interps
	c.Interps = make(map[string][]Node, len(m.Interps))
	for k, v := range m.Interps {
		c.Interps[k] = append([]Node{}, v...)
	}

	// Shared per-session pointers (Python: copy.copy does shallow copy)
	c.CompCfg = m.CompCfg
	c.Instantiator = m.Instantiator
	c.Theory = m.Theory
	c.z3SharedCtx = m.z3SharedCtx

	// Copy signature (deep)
	c.Sig = m.Sig.Copy()

	return c
}

// AddToHierarchy adds a dotted name to the hierarchy tree.
func (m *Module) AddToHierarchy(name string) {
	cc := m.Cfg.IuCfg.ComposeCharacter
	if idx := strings.LastIndex(name, cc); idx >= 0 {
		pref := name[:idx]
		suff := name[idx+len(cc):]
		m.AddToHierarchy(pref)
		inner, ok := m.Hierarchy.Get2(pref)
		if !ok {
			inner = NewInsMap[string, bool]()
			m.Hierarchy.Set(pref, inner)
		}
		inner.Set(suff, true)
	} else {
		inner, ok := m.Hierarchy.Get2("this")
		if !ok {
			inner = NewInsMap[string, bool]()
			m.Hierarchy.Set("this", inner)
		}
		inner.Set(name, true)
	}
}

// AddObject adds an object name to the hierarchy.
func (m *Module) AddObject(name string) {
	if _, ok := m.Hierarchy.Get2(name); !ok {
		m.Hierarchy.Set(name, NewInsMap[string, bool]())
	}
}

// SetAction inserts or replaces an action, preserving insertion order.
// Matches Python dict semantics: new keys are appended, existing keys keep position.
func (m *Module) SetAction(name string, action Action) {
	m.Actions.Set(name, action)
}

// FindAction looks up an action by name.
func (m *Module) FindAction(name string) (Action, bool) {
	return m.Actions.Get2(name)
}

// IsVariant returns true if rsort is a variant of lsort.
func (m *Module) IsVariant(lsort, rsort Sort) bool {
	lname := IvySortName(lsort)
	variants, ok := m.Variants[lname]
	if !ok {
		return false
	}
	for _, v := range variants {
		if SortEqual(v, rsort) {
			return true
		}
	}
	return false
}

// VariantIndex returns the index of variant rsort within lsort's variants.
// Returns -1 if not found.
func (m *Module) VariantIndex(lsort, rsort Sort) int {
	lname := IvySortName(lsort)
	variants, ok := m.Variants[lname]
	if !ok {
		return -1
	}
	for i, v := range variants {
		if SortEqual(v, rsort) {
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
func (m *Module) SortCard(sort Sort) int {
	if IsFunctionSort(sort) {
		return -1
	}
	name := IvySortName(sort)
	attr := m.Cfg.IuCfg.ComposeNames(name, "cardinality")
	if val, ok := m.Attributes[attr]; ok {
		// Python: int(self.attributes[attr].rep)
		// The attribute value is an AST node. Use Sexp() for structural
		// equivalence when the value is a lg.Expr; fall back to Relname()
		// for ast.Node types, mirroring Python's .rep access.
		var rep string
		switch v := val.(type) {
		case Expr:
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
func SortCardDefault(sort Sort) int {
	if es, ok := sort.(*LogicEnumeratedSort); ok {
		return es.Card()
	}
	return -1
}

// CallGraph builds the action call graph: called → callers.
// Corresponds to Python's Module.call_graph (ivy_module.py:279-284).
func (m *Module) CallGraph() map[string][]string {
	callgraph := make(map[string][]string)
	for actname, action := range m.Actions.All() {
		for _, calledName := range action.IterCalls() {
			callgraph[calledName] = append(callgraph[calledName], actname)
		}
	}
	return callgraph
}

// SortDependencies returns sort names that the given sort depends on.
func (m *Module) SortDependencies(sortName string, withVariants bool) []string {
	if destrs, ok := m.SortDestructors.Get2(sortName); ok {
		var deps []string
		for _, destr := range destrs {
			if fs, ok := destr.CSort.(*LogicFunctionSort); ok {
				dom := fs.Domain()
				for _, d := range dom[1:] {
					deps = append(deps, IvySortName(d))
				}
				deps = append(deps, IvySortName(fs.Range()))
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
					if _, inSig := m.Sig.Sorts.Get2(rep); inSig {
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
				deps[i] = IvySortName(v)
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
	return DefaultLogics
}

// --- String representation ---

func (m *Module) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Module: %d axioms, %d conjs, %d actions, %d sorts\n",
		len(m.LabeledAxioms), len(m.LabeledConjs),
		m.Actions.Len(), m.Sig.Sorts.Len())
	return b.String()
}

// --- helper copy functions ---

func copyNodeSlice(s []Expr) []Expr {
	if s == nil {
		return nil
	}
	c := make([]Expr, len(s))
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

func copyMapSort(m map[string]Sort) map[string]Sort {
	c := make(map[string]Sort, len(m))
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

func copyMapNode(m map[string]Node) map[string]Node {
	c := make(map[string]Node, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

func copyInsMapNode(m *InsMap[string, Node]) *InsMap[string, Node] {
	c := NewInsMap[string, Node]()
	if m == nil {
		return c
	}
	for k, v := range m.All() {
		c.Set(k, v)
	}
	return c
}

func copyMapNativeType(m map[string]*NativeType) map[string]*NativeType {
	c := make(map[string]*NativeType, len(m))
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
// Corresponds to Python Module.update_conjs (ivy_module.py:268-281):
//
//	for i,cax in enumerate(mod.labeled_conjs):
//	    fmla = cax.formula
//	    csname = 'conjecture:'+ str(i)
//	    variables = list(lu.used_variables_in_order_ast(fmla))
//	    sort = il.RelationSort([v.sort for v in variables])
//	    sym = il.Symbol(csname,sort)
//	    space = ics.NamedSpace(il.Literal(0,fmla))
//	    mod.concept_spaces.append((sym(*variables),space))
func (m *Module) UpdateConjs() {
	for i, cax := range m.LabeledConjs {
		if cax == nil || cax.Formula == nil {
			continue
		}
		fmla, ok := cax.Formula.(Expr)
		if !ok {
			continue
		}
		csname := fmt.Sprintf("conjecture:%d", i)

		// Collect free variables in left-to-right traversal order with
		// first-occurrence dedup. Python uses lu.used_variables_in_order_ast
		// in ivy_module.py:277 for the same reason: deterministic, portable
		// ordering of concept-space label variables across languages.
		variables := VariablesAST(fmla)

		// Build sort: LogicRelationSort([v.sort for v in variables])
		sorts := make([]Sort, len(variables))
		for j, v := range variables {
			sorts[j] = v.VSort
		}
		symSort := LogicRelationSort(sorts)
		sym := NewConst(csname, symSort)

		// Build label: sym(*variables)
		var label Expr
		if len(variables) > 0 {
			varExprs := make([]Expr, len(variables))
			for j, v := range variables {
				varExprs[j] = v
			}
			label = MustApply(sym, varExprs...)
		} else {
			label = sym
		}

		// Build space: NamedSpace(il.Literal(0, fmla))
		space := &LogicLiteral{Polarity: 0, Atom: fmla}

		m.ConceptSpaces = append(m.ConceptSpaces, ConceptSpace{Label: label, Body: space})
	}
}

// RelationKey is the structural key Python uses for module.relations:
// lg.Const(name, sort), whose recstruct hash/equality includes both fields.
func RelationKey(name string, sort Sort) NodeKey {
	return Key(NewConst(name, sort))
}

// FunctionKey is the structural key Python uses for module.functions:
// lg.Const(name, sort), whose recstruct hash/equality includes both fields.
func FunctionKey(name string, sort Sort) NodeKey {
	return Key(NewConst(name, sort))
}
