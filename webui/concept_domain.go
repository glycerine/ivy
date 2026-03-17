// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from concept.py.
//
// This file defines the concept domain model for Ivy's concept graph:
//   - CDConcept: a quantifier-free formula with first-order free variables
//   - CDConceptSet: a named set of concept names (iterable)
//   - CDConceptDict: an ordered dict of concepts
//   - CDConceptCombiner: a formula with second-order relational free variables
//   - CDConceptDomain: set of named concepts, combiners, and combinations
//   - Various factory/helper functions
//
// Types are prefixed with CD (Concept Domain) to avoid naming conflicts
// with the rendering-layer stub types in concept.go.
package webui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/ivyutils"
	"github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/logicutil"
	"github.com/glycerine/goivy/typeinfer"
)

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

// ArityError is raised when a concept or combiner is called with the wrong
// number of arguments.
type ArityError struct {
	Expected int
	Got      int
	Detail   string
}

func (e *ArityError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("arity error: expected %d, got %d: %s", e.Expected, e.Got, e.Detail)
	}
	return fmt.Sprintf("arity error: expected %d, got %d", e.Expected, e.Got)
}

// ---------------------------------------------------------------------------
// CDConcept
// ---------------------------------------------------------------------------

// CDConcept is a quantifier-free formula with some first-order free variables.
// It represents a unary, binary, or higher-arity concept in the concept graph.
type CDConcept struct {
	Name      string
	Variables []*logic.Var
	Formula   logic.Node
}

// NewCDConcept creates a new concept, validating that all variables are
// first-order. Returns an error if validation fails.
func NewCDConcept(name string, variables []*logic.Var, formula logic.Node) (*CDConcept, error) {
	if name == "" {
		return nil, &logic.IvyError{Msg: "Concept name is empty"}
	}
	vars := make([]*logic.Var, len(variables))
	copy(vars, variables)
	for _, v := range vars {
		if !logic.FirstOrderSort(v.VSort) {
			return nil, &logic.IvyError{Msg: fmt.Sprintf("Concept variables must be first-order: %v", v)}
		}
	}
	return &CDConcept{Name: name, Variables: vars, Formula: formula}, nil
}

// MustCDConcept is like NewCDConcept but panics on error.
func MustCDConcept(name string, variables []*logic.Var, formula logic.Node) *CDConcept {
	c, err := NewCDConcept(name, variables, formula)
	if err != nil {
		panic(err)
	}
	return c
}

// Arity returns the number of variables (parameters) of the concept.
func (c *CDConcept) Arity() int { return len(c.Variables) }

// Sorts returns the sort of each variable.
func (c *CDConcept) Sorts() []logic.Sort {
	s := make([]logic.Sort, len(c.Variables))
	for i, v := range c.Variables {
		s[i] = v.VSort
	}
	return s
}

// Sort returns the sort of the first variable.
func (c *CDConcept) Sort() logic.Sort {
	if len(c.Variables) == 0 {
		return nil
	}
	return c.Variables[0].VSort
}

// Call substitutes the variables with the given terms, returning a formula.
// This corresponds to Python's Concept.__call__.
func (c *CDConcept) Call(terms ...logic.Node) (logic.Node, error) {
	if len(terms) != c.Arity() {
		return nil, &ArityError{Expected: c.Arity(), Got: len(terms)}
	}
	subs := make(map[logic.NodeKey]logic.Node)
	for i, v := range c.Variables {
		t := terms[i]
		// Concretize sorts to match the variable's sort.
		ct, err := typeinfer.ConcretizeSorts(t, v.VSort)
		if err != nil {
			ct = t // fallback: use as-is
		}
		subs[logic.Key(v)] = ct
	}
	return logicutil.Substitute(c.Formula, subs)
}

// String returns a human-readable representation.
func (c *CDConcept) String() string {
	varNames := make([]string, len(c.Variables))
	for i, v := range c.Variables {
		varNames[i] = v.String()
	}
	return fmt.Sprintf("Concept %s. %s", strings.Join(varNames, ", "), c.Formula)
}

// FormulaStr returns the formula as a string (for rendering compatibility).
func (c *CDConcept) FormulaStr() string {
	if c.Formula == nil {
		return ""
	}
	return c.Formula.String()
}

// ---------------------------------------------------------------------------
// CDConceptSet
// ---------------------------------------------------------------------------

// CDConceptSet is a named set of concept names.
type CDConceptSet struct {
	Names []string
}

// NewCDConceptSet creates a concept set from a list of names.
func NewCDConceptSet(names ...string) *CDConceptSet {
	cp := make([]string, len(names))
	copy(cp, names)
	return &CDConceptSet{Names: cp}
}

// ---------------------------------------------------------------------------
// CDConceptDict — an ordered dict of concepts (and name-lists)
// ---------------------------------------------------------------------------

// CDConceptEntry is one entry in a CDConceptDict.
type CDConceptEntry struct {
	Key     string
	Concept *CDConcept    // non-nil for real concepts
	List    []string      // non-nil for name-lists (like "nodes", "edges")
	Set     *CDConceptSet // non-nil for concept sets
}

// CDConceptDict is an ordered collection of concepts and name-lists.
// It maintains insertion order and allows O(1) lookup by name.
type CDConceptDict struct {
	entries []CDConceptEntry
	index   map[string]int // key -> position in entries
}

// NewCDConceptDict creates a new empty CDConceptDict.
func NewCDConceptDict() *CDConceptDict {
	return &CDConceptDict{
		index: make(map[string]int),
	}
}

// Len returns the number of entries.
func (d *CDConceptDict) Len() int { return len(d.entries) }

// Has returns true if the key exists.
func (d *CDConceptDict) Has(key string) bool {
	_, ok := d.index[key]
	return ok
}

// GetConcept returns the concept for the given key, or nil.
func (d *CDConceptDict) GetConcept(key string) *CDConcept {
	i, ok := d.index[key]
	if !ok {
		return nil
	}
	return d.entries[i].Concept
}

// GetList returns the name-list for the given key, or nil.
func (d *CDConceptDict) GetList(key string) []string {
	i, ok := d.index[key]
	if !ok {
		return nil
	}
	return d.entries[i].List
}

// GetSet returns the concept set for the given key, or nil.
func (d *CDConceptDict) GetSet(key string) *CDConceptSet {
	i, ok := d.index[key]
	if !ok {
		return nil
	}
	return d.entries[i].Set
}

// GetEntry returns the entry for the given key.
func (d *CDConceptDict) GetEntry(key string) (CDConceptEntry, bool) {
	i, ok := d.index[key]
	if !ok {
		return CDConceptEntry{}, false
	}
	return d.entries[i], true
}

// SetConcept adds or replaces a concept entry.
func (d *CDConceptDict) SetConcept(key string, c *CDConcept) {
	if i, ok := d.index[key]; ok {
		d.entries[i] = CDConceptEntry{Key: key, Concept: c}
	} else {
		d.index[key] = len(d.entries)
		d.entries = append(d.entries, CDConceptEntry{Key: key, Concept: c})
	}
}

// SetList adds or replaces a name-list entry.
func (d *CDConceptDict) SetList(key string, list []string) {
	cp := make([]string, len(list))
	copy(cp, list)
	if i, ok := d.index[key]; ok {
		d.entries[i] = CDConceptEntry{Key: key, List: cp}
	} else {
		d.index[key] = len(d.entries)
		d.entries = append(d.entries, CDConceptEntry{Key: key, List: cp})
	}
}

// SetSet adds or replaces a concept set entry.
func (d *CDConceptDict) SetSet(key string, s *CDConceptSet) {
	if i, ok := d.index[key]; ok {
		d.entries[i] = CDConceptEntry{Key: key, Set: s}
	} else {
		d.index[key] = len(d.entries)
		d.entries = append(d.entries, CDConceptEntry{Key: key, Set: s})
	}
}

// Delete removes an entry by key.
func (d *CDConceptDict) Delete(key string) {
	i, ok := d.index[key]
	if !ok {
		return
	}
	d.entries = append(d.entries[:i], d.entries[i+1:]...)
	delete(d.index, key)
	// Rebuild index for entries after the deleted one.
	for j := i; j < len(d.entries); j++ {
		d.index[d.entries[j].Key] = j
	}
}

// Keys returns all keys in order.
func (d *CDConceptDict) Keys() []string {
	keys := make([]string, len(d.entries))
	for i, e := range d.entries {
		keys[i] = e.Key
	}
	return keys
}

// Entries returns a copy of all entries in order.
func (d *CDConceptDict) Entries() []CDConceptEntry {
	cp := make([]CDConceptEntry, len(d.entries))
	copy(cp, d.entries)
	return cp
}

// ForEachConcept iterates over entries that are concepts.
func (d *CDConceptDict) ForEachConcept(fn func(name string, c *CDConcept)) {
	for _, e := range d.entries {
		if e.Concept != nil {
			fn(e.Key, e.Concept)
		}
	}
}

// AppendToList appends a name to an existing name-list entry.
func (d *CDConceptDict) AppendToList(key, name string) {
	i, ok := d.index[key]
	if !ok {
		d.SetList(key, []string{name})
		return
	}
	d.entries[i].List = append(d.entries[i].List, name)
}

// Copy returns a deep copy of the CDConceptDict.
func (d *CDConceptDict) Copy() *CDConceptDict {
	cp := NewCDConceptDict()
	for _, e := range d.entries {
		switch {
		case e.Concept != nil:
			cp.SetConcept(e.Key, e.Concept)
		case e.List != nil:
			cp.SetList(e.Key, e.List)
		case e.Set != nil:
			names := make([]string, len(e.Set.Names))
			copy(names, e.Set.Names)
			cp.SetSet(e.Key, &CDConceptSet{Names: names})
		default:
			// empty entry
			cp.SetList(e.Key, nil)
		}
	}
	return cp
}

// Reorder creates a new CDConceptDict with entries reordered according to keys.
// Keys not in the dict are skipped. Entries not in keys are dropped.
func (d *CDConceptDict) Reorder(keys []string) *CDConceptDict {
	cp := NewCDConceptDict()
	for _, k := range keys {
		if e, ok := d.GetEntry(k); ok {
			switch {
			case e.Concept != nil:
				cp.SetConcept(k, e.Concept)
			case e.List != nil:
				cp.SetList(k, e.List)
			case e.Set != nil:
				cp.SetSet(k, e.Set)
			}
		}
	}
	return cp
}

// ---------------------------------------------------------------------------
// CDConceptCombiner
// ---------------------------------------------------------------------------

// CDConceptCombiner is a formula with second-order free variables
// (relational, i.e. having Boolean range). The quantifier structure
// should be AE for checks to stay in EPR.
type CDConceptCombiner struct {
	Variables []*logic.Var
	Formula   logic.Node
}

// NewCDConceptCombiner creates a combiner, validating that all variables
// are relational (FunctionSort with Boolean range).
func NewCDConceptCombiner(variables []*logic.Var, formula logic.Node) (*CDConceptCombiner, error) {
	vars := make([]*logic.Var, len(variables))
	copy(vars, variables)
	for _, v := range vars {
		fs, ok := v.VSort.(*logic.FunctionSort)
		if !ok || !logic.SortEqual(fs.Range(), logic.Boolean) {
			return nil, &logic.IvyError{Msg: fmt.Sprintf("ConceptCombiner variables must be relational: %v (sort: %s)", v, v.VSort)}
		}
	}
	return &CDConceptCombiner{Variables: vars, Formula: formula}, nil
}

// MustCDConceptCombiner is like NewCDConceptCombiner but panics on error.
func MustCDConceptCombiner(variables []*logic.Var, formula logic.Node) *CDConceptCombiner {
	cc, err := NewCDConceptCombiner(variables, formula)
	if err != nil {
		panic(err)
	}
	return cc
}

// Arity returns the number of second-order variables.
func (cc *CDConceptCombiner) Arity() int { return len(cc.Variables) }

// Arities returns the arity (domain length) of each variable.
func (cc *CDConceptCombiner) Arities() []int {
	a := make([]int, len(cc.Variables))
	for i, v := range cc.Variables {
		if fs, ok := v.VSort.(*logic.FunctionSort); ok {
			a[i] = len(fs.Domain())
		}
	}
	return a
}

// Call instantiates the combiner with the given concepts.
// This corresponds to Python's ConceptCombiner.__call__, which uses
// substitute_apply.
func (cc *CDConceptCombiner) Call(concepts ...*CDConcept) (logic.Node, error) {
	if len(concepts) != cc.Arity() {
		return nil, &ArityError{Expected: cc.Arity(), Got: len(concepts)}
	}
	arities := cc.Arities()
	for i, c := range concepts {
		if c.Arity() != arities[i] {
			return nil, &ArityError{
				Expected: arities[i],
				Got:      c.Arity(),
				Detail:   fmt.Sprintf("concept %q arity mismatch at position %d", c.Name, i),
			}
		}
	}
	// Build substitution: replace each combiner variable applied to args
	// with the concept's formula applied to the same args.
	// This is substitute_apply in Python.
	result := substituteApplyNode(cc.Formula, cc.Variables, concepts)
	// Concretize sorts — replaces TopSort with inferred concrete sorts.
	// If this fails, the formula still has TopSort and cannot be sent to Z3.
	cr, err := typeinfer.ConcretizeSorts(result, nil)
	if err != nil {
		return nil, fmt.Errorf("concretize failed: %w", err)
	}
	// Verify no TopSort remains — Z3 panics on sort mismatches.
	if logic.ContainsTopSort(cr) {
		return nil, fmt.Errorf("formula still contains TopSort after concretization")
	}
	return cr, nil
}

// String returns a human-readable representation.
func (cc *CDConceptCombiner) String() string {
	varNames := make([]string, len(cc.Variables))
	for i, v := range cc.Variables {
		varNames[i] = v.String()
	}
	return fmt.Sprintf("ConceptCombiner %s. %s", strings.Join(varNames, ", "), cc.Formula)
}

// substituteApplyNode replaces applications of combiner variables with
// concept formula applications. This corresponds to Python's substitute_apply.
func substituteApplyNode(node logic.Node, variables []*logic.Var, concepts []*CDConcept) logic.Node {
	// Build a mapping from variable name -> concept (by_name matching like Python).
	varMap := make(map[string]*CDConcept)
	for i, v := range variables {
		varMap[v.Name] = concepts[i]
	}
	return substApplyRec(node, varMap)
}

func substApplyRec(node logic.Node, varMap map[string]*CDConcept) logic.Node {
	switch n := node.(type) {
	case *logic.Apply:
		// Check if function is a variable that should be replaced.
		if v, ok := n.Func.(*logic.Var); ok {
			if concept, found := varMap[v.Name]; found {
				// Replace V(args...) with concept.formula[concept.vars -> args]
				args := make([]logic.Node, len(n.Terms))
				for i, t := range n.Terms {
					args[i] = substApplyRec(t, varMap)
				}
				result, err := concept.Call(args...)
				if err != nil {
					return node // fallback
				}
				return result
			}
		}
		// Recurse on function and terms.
		newFunc := substApplyRec(n.Func, varMap)
		newTerms := make([]logic.Node, len(n.Terms))
		changed := newFunc != n.Func
		for i, t := range n.Terms {
			newTerms[i] = substApplyRec(t, varMap)
			if newTerms[i] != t {
				changed = true
			}
		}
		if !changed {
			return node
		}
		result, err := logic.NewApply(newFunc, newTerms...)
		if err != nil {
			return node
		}
		return result

	case *logic.Eq:
		t1 := substApplyRec(n.T1, varMap)
		t2 := substApplyRec(n.T2, varMap)
		if t1 == n.T1 && t2 == n.T2 {
			return node
		}
		result, err := logic.NewEq(t1, t2)
		if err != nil {
			return node
		}
		return result

	case *logic.Not:
		b := substApplyRec(n.Body, varMap)
		if b == n.Body {
			return node
		}
		result, _ := logic.NewNot(b)
		return result

	case *logic.And:
		terms := substApplySlice(n.Terms, varMap)
		if terms == nil {
			return node
		}
		result, _ := logic.NewAnd(terms...)
		return result

	case *logic.Or:
		terms := substApplySlice(n.Terms, varMap)
		if terms == nil {
			return node
		}
		result, _ := logic.NewOr(terms...)
		return result

	case *logic.Implies:
		t1 := substApplyRec(n.T1, varMap)
		t2 := substApplyRec(n.T2, varMap)
		if t1 == n.T1 && t2 == n.T2 {
			return node
		}
		result, _ := logic.NewImplies(t1, t2)
		return result

	case *logic.ForAll:
		b := substApplyRec(n.Body, varMap)
		if b == n.Body {
			return node
		}
		result, _ := logic.NewForAll(n.Variables, b)
		return result

	case *logic.Exists:
		b := substApplyRec(n.Body, varMap)
		if b == n.Body {
			return node
		}
		result, _ := logic.NewExists(n.Variables, b)
		return result

	case *logic.Var:
		// A bare variable (not applied) that's in the map:
		// This means U used as a standalone, treat as U() but concepts
		// don't support 0-arity this way. Return as-is.
		return node

	default:
		return node
	}
}

func substApplySlice(terms []logic.Node, varMap map[string]*CDConcept) []logic.Node {
	newTerms := make([]logic.Node, len(terms))
	changed := false
	for i, t := range terms {
		newTerms[i] = substApplyRec(t, varMap)
		if newTerms[i] != t {
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return newTerms
}

// ---------------------------------------------------------------------------
// CDCombinerDict — ordered dict of combiners (string keys)
// ---------------------------------------------------------------------------

// CDCombinerEntry is one entry in a CDCombinerDict.
type CDCombinerEntry struct {
	Key      string
	Combiner *CDConceptCombiner // non-nil for real combiners
	List     []string           // non-nil for name-lists (combiner groups)
}

// CDCombinerDict is an ordered collection of combiners and name-lists.
type CDCombinerDict struct {
	entries []CDCombinerEntry
	index   map[string]int
}

// NewCDCombinerDict creates a new empty CDCombinerDict.
func NewCDCombinerDict() *CDCombinerDict {
	return &CDCombinerDict{
		index: make(map[string]int),
	}
}

// SetCombiner adds or replaces a combiner entry.
func (d *CDCombinerDict) SetCombiner(key string, c *CDConceptCombiner) {
	if i, ok := d.index[key]; ok {
		d.entries[i] = CDCombinerEntry{Key: key, Combiner: c}
	} else {
		d.index[key] = len(d.entries)
		d.entries = append(d.entries, CDCombinerEntry{Key: key, Combiner: c})
	}
}

// SetList adds or replaces a list entry.
func (d *CDCombinerDict) SetList(key string, list []string) {
	cp := make([]string, len(list))
	copy(cp, list)
	if i, ok := d.index[key]; ok {
		d.entries[i] = CDCombinerEntry{Key: key, List: cp}
	} else {
		d.index[key] = len(d.entries)
		d.entries = append(d.entries, CDCombinerEntry{Key: key, List: cp})
	}
}

// GetCombiner returns the combiner for the given key, or nil.
func (d *CDCombinerDict) GetCombiner(key string) *CDConceptCombiner {
	i, ok := d.index[key]
	if !ok {
		return nil
	}
	return d.entries[i].Combiner
}

// GetList returns the list for the given key, or nil.
func (d *CDCombinerDict) GetList(key string) []string {
	i, ok := d.index[key]
	if !ok {
		return nil
	}
	return d.entries[i].List
}

// Has returns true if the key exists.
func (d *CDCombinerDict) Has(key string) bool {
	_, ok := d.index[key]
	return ok
}

// GetEntry returns the entry for the given key.
func (d *CDCombinerDict) GetEntry(key string) (CDCombinerEntry, bool) {
	i, ok := d.index[key]
	if !ok {
		return CDCombinerEntry{}, false
	}
	return d.entries[i], true
}

// Copy returns a deep copy of the CDCombinerDict.
func (d *CDCombinerDict) Copy() *CDCombinerDict {
	cp := NewCDCombinerDict()
	for _, e := range d.entries {
		switch {
		case e.Combiner != nil:
			cp.SetCombiner(e.Key, e.Combiner)
		case e.List != nil:
			cp.SetList(e.Key, e.List)
		}
	}
	return cp
}

// ---------------------------------------------------------------------------
// Combination
// ---------------------------------------------------------------------------

// Combination is a tuple describing how to combine concepts.
// Elements[0] = combination label (e.g., "node_info")
// Elements[1] = combiner class name (e.g., "node_info")
// Elements[2:] = concept set names
type Combination struct {
	Elements []string
}

// Name returns the combination name (first element).
func (c *Combination) Name() string {
	if len(c.Elements) == 0 {
		return ""
	}
	return c.Elements[0]
}

// CombinerClass returns the combiner class (second element).
func (c *Combination) CombinerClass() string {
	if len(c.Elements) < 2 {
		return ""
	}
	return c.Elements[1]
}

// ConceptSets returns the concept set names (elements[2:]).
func (c *Combination) ConceptSets() []string {
	if len(c.Elements) < 3 {
		return nil
	}
	return c.Elements[2:]
}

// NewCombination creates a combination from elements.
func NewCombination(elems ...string) Combination {
	cp := make([]string, len(elems))
	copy(cp, elems)
	return Combination{Elements: cp}
}

// ---------------------------------------------------------------------------
// Tag and Fact
// ---------------------------------------------------------------------------

// Tag is a tuple of strings identifying a specific concept fact.
// Typically: (combination_name, combiner_name, concept_name_1, ..., concept_name_k)
type Tag []string

// TagString returns the tag as a pipe-separated string for map keys.
func TagString(t Tag) string {
	return strings.Join(t, "|")
}

// TagFromString reconstructs a Tag from a pipe-separated string.
func TagFromString(s string) Tag {
	return Tag(strings.Split(s, "|"))
}

// Fact is a (Tag, formula) pair.
type Fact struct {
	Tag     Tag
	Formula logic.Node
}

// TagValue is a (Tag, bool) pair used by alpha abstraction results.
type TagValue struct {
	Tag   Tag
	Value bool
}

// ---------------------------------------------------------------------------
// CDConceptDomain
// ---------------------------------------------------------------------------

// CDConceptDomain is a set of named concepts, named concept combiners,
// and a list of combinations. It is mutable, so copy before changing.
type CDConceptDomain struct {
	Concepts     *CDConceptDict
	Combiners    *CDCombinerDict
	Combinations []Combination
}

// NewCDConceptDomain creates a new concept domain.
func NewCDConceptDomain(concepts *CDConceptDict, combiners *CDCombinerDict, combinations []Combination) *CDConceptDomain {
	if concepts == nil {
		concepts = NewCDConceptDict()
	}
	if combiners == nil {
		combiners = NewCDCombinerDict()
	}
	combCp := make([]Combination, len(combinations))
	copy(combCp, combinations)
	return &CDConceptDomain{
		Concepts:     concepts.Copy(),
		Combiners:    combiners.Copy(),
		Combinations: combCp,
	}
}

// Copy returns a deep copy of the concept domain.
func (d *CDConceptDomain) Copy() *CDConceptDomain {
	combCp := make([]Combination, len(d.Combinations))
	for i, c := range d.Combinations {
		elems := make([]string, len(c.Elements))
		copy(elems, c.Elements)
		combCp[i] = Combination{Elements: elems}
	}
	return &CDConceptDomain{
		Concepts:     d.Concepts.Copy(),
		Combiners:    d.Combiners.Copy(),
		Combinations: combCp,
	}
}

// ConceptsByArity returns the names of all concepts with the given arity.
func (d *CDConceptDomain) ConceptsByArity(a int) []string {
	var result []string
	d.Concepts.ForEachConcept(func(name string, c *CDConcept) {
		if c.Arity() == a {
			result = append(result, name)
		}
	})
	return result
}

// PossibleNodeLabels returns the names of all unary concepts that are
// not in the "nodes" list.
func (d *CDConceptDomain) PossibleNodeLabels() []string {
	nodes := make(map[string]bool)
	for _, n := range d.Concepts.GetList("nodes") {
		nodes[n] = true
	}
	var result []string
	d.Concepts.ForEachConcept(func(name string, c *CDConcept) {
		if c.Arity() == 1 && !nodes[name] {
			result = append(result, name)
		}
	})
	return result
}

// GetCombFacts computes facts for a single combination, appending to *facts.
func (d *CDConceptDomain) GetCombFacts(combinationName, combinerClass string, conceptNames [][]string, facts *[]Fact) {
	combinerNames := cdResolveName(d.Combiners, combinerClass)
	// Cartesian product of conceptNames
	for _, conceptCombo := range cartesianProduct(conceptNames) {
		concepts := make([]*CDConcept, len(conceptCombo))
		for i, n := range conceptCombo {
			concepts[i] = d.Concepts.GetConcept(n)
			if concepts[i] == nil {
				break
			}
		}
		if len(concepts) > 0 && concepts[len(concepts)-1] == nil {
			continue
		}
		for _, combinerName := range combinerNames {
			combiner := d.Combiners.GetCombiner(combinerName)
			if combiner == nil {
				continue
			}
			formula, err := combiner.Call(concepts...)
			if err != nil || formula == nil {
				// Skip ill-sorted or unconcretizable instantiations.
				continue
			}
			tag := make(Tag, 0, 2+len(conceptCombo))
			tag = append(tag, combinationName, combinerName)
			tag = append(tag, conceptCombo...)
			*facts = append(*facts, Fact{Tag: tag, Formula: formula})
		}
	}
}

// GetFacts returns all facts from all combinations.
// If projection is non-nil, it filters which concept names to include.
// Signature: projection(conceptName, categoryName) bool.
//
// Special-cases "edge_info", "node_label", and "enum" for performance.
func (d *CDConceptDomain) GetFacts(projection func(string, string) bool) []Fact {
	var facts []Fact

	// Build nodes-by-sort-name lookup.
	nodesBySortName := make(map[string][]string)
	for _, n := range d.Concepts.GetList("nodes") {
		c := d.Concepts.GetConcept(n)
		if c == nil {
			continue
		}
		sn := c.Sorts()[0].String()
		nodesBySortName[sn] = append(nodesBySortName[sn], n)
	}

	for _, combination := range d.Combinations {
		combinationName := combination.Name()

		if combinationName == "edge_info" {
			edges := d.Concepts.GetList("edges")
			for _, e := range edges {
				if projection != nil && !projection(e, "edges") {
					continue
				}
				ec := d.Concepts.GetConcept(e)
				if ec == nil || len(ec.Sorts()) < 2 {
					continue
				}
				c0 := nodesBySortName[ec.Sorts()[0].String()]
				c1 := nodesBySortName[ec.Sorts()[1].String()]
				d.GetCombFacts("edge_info", "edge_info", [][]string{{e}, c0, c1}, &facts)
			}
		} else if combinationName == "node_label" || combinationName == "enum" {
			labelCategory := ""
			if len(combination.Elements) > 3 {
				labelCategory = combination.Elements[3]
			}
			labels := d.Concepts.GetList(labelCategory)
			for _, l := range labels {
				if projection != nil && !projection(l, labelCategory) {
					continue
				}
				for _, cname := range cdResolveConceptName(d.Concepts, l) {
					lc := d.Concepts.GetConcept(cname)
					if lc == nil || len(lc.Sorts()) < 1 {
						continue
					}
					c0 := nodesBySortName[lc.Sorts()[0].String()]
					d.GetCombFacts("node_label", "node_label", [][]string{c0, {cname}}, &facts)
				}
			}
		} else {
			conceptSets := combination.ConceptSets()
			conceptNames := make([][]string, len(conceptSets))
			for i, x := range conceptSets {
				conceptNames[i] = cdResolveConceptName(d.Concepts, x)
			}
			conceptNames = cdFilterConceptNames(conceptNames, conceptSets, projection)
			d.GetCombFacts(combinationName, combination.CombinerClass(), conceptNames, &facts)
		}
	}
	return facts
}

// Split splits a concept into sub-concepts based on split_by.
func (d *CDConceptDomain) Split(concept, splitBy string) {
	c1 := d.Concepts.GetConcept(concept)
	if c1 == nil {
		return
	}
	c2Entry, c2exists := d.Concepts.GetEntry(splitBy)

	variables := c1.Variables
	var newNames []string
	var newConcepts []*CDConcept

	if c2exists && c2Entry.Set != nil {
		// splitting by a concept set
		for _, n := range c2Entry.Set.Names {
			nc := d.Concepts.GetConcept(n)
			if nc == nil {
				continue
			}
			f, err := nc.Call(nodesToSlice(variables)...)
			if err != nil {
				continue
			}
			newName := fmt.Sprintf("(%s+%s)", concept, n)
			andF, _ := logic.NewAnd(c1.Formula, f)
			newConcepts = append(newConcepts, MustCDConcept(newName, variables, andF))
			newNames = append(newNames, newName)
		}
	} else {
		// splitting by a single concept
		c2 := d.Concepts.GetConcept(splitBy)
		if c2 == nil {
			return
		}
		posF, err := c2.Call(nodesToSlice(variables)...)
		if err != nil {
			return
		}
		negF, _ := logic.NewNot(posF)
		posAnd, _ := logic.NewAnd(c1.Formula, posF)
		negAnd, _ := logic.NewAnd(c1.Formula, negF)

		posName := fmt.Sprintf("(%s+%s)", concept, splitBy)
		negName := fmt.Sprintf("(%s-%s)", concept, splitBy)

		newConcepts = append(newConcepts,
			MustCDConcept(posName, variables, posAnd),
			MustCDConcept(negName, variables, negAnd),
		)
		newNames = append(newNames, posName, negName)
	}

	// Add new concepts to the dict.
	for i, nc := range newConcepts {
		d.Concepts.SetConcept(newNames[i], nc)
	}

	// Reorder: replace concept with new names in the key order.
	oldKeys := d.Concepts.Keys()
	newKeys := cdReplaceName(oldKeys, concept, newNames)
	d.Concepts = d.Concepts.Reorder(newKeys)

	d.ReplaceConcept(concept, newNames)
}

// ReplaceConcept replaces a concept name with a list of new names
// in all name-lists and combinations.
func (d *CDConceptDomain) ReplaceConcept(concept string, newNames []string) {
	for _, e := range d.Concepts.Entries() {
		if e.Concept == nil && e.Set == nil && e.List != nil {
			d.Concepts.SetList(e.Key, cdReplaceNameInList(e.List, concept, newNames))
		}
	}
	for i, comb := range d.Combinations {
		newElems := make([]string, 0, len(comb.Elements))
		newElems = append(newElems, comb.Elements[0]) // combination name stays
		for _, x := range comb.Elements[1:] {
			newElems = append(newElems, x) // combiner class stays
		}
		// Actually, replace in elements 1: onward
		replaced := make([]string, 0, len(comb.Elements))
		replaced = append(replaced, comb.Elements[0])
		for _, x := range comb.Elements[1:] {
			repl := cdReplaceNameInList([]string{x}, concept, newNames)
			replaced = append(replaced, repl...)
		}
		d.Combinations[i] = Combination{Elements: replaced}
	}
}

// Output prints the concept domain to stdout (for debugging).
func (d *CDConceptDomain) Output() {
	fmt.Println("Concepts:")
	for _, e := range d.Concepts.Entries() {
		if e.Concept != nil {
			fmt.Printf("    %s : %s\n", e.Key, e.Concept)
		} else if e.List != nil {
			fmt.Printf("    %s : %v\n", e.Key, e.List)
		} else if e.Set != nil {
			fmt.Printf("    %s : set%v\n", e.Key, e.Set.Names)
		}
	}
	fmt.Println("Concept Combiners:")
	for _, e := range d.Combiners.entries {
		if e.Combiner != nil {
			fmt.Printf("    %s : %s\n", e.Key, e.Combiner)
		} else if e.List != nil {
			fmt.Printf("    %s : %v\n", e.Key, e.List)
		}
	}
	fmt.Println("Combinations:")
	for _, c := range d.Combinations {
		fmt.Printf("    %v\n", c.Elements)
	}
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

// nodesToSlice converts a []*logic.Var to []logic.Node.
func nodesToSlice(vars []*logic.Var) []logic.Node {
	nodes := make([]logic.Node, len(vars))
	for i, v := range vars {
		nodes[i] = v
	}
	return nodes
}

// cdUnionLists returns a list containing all elements of the input lists
// without duplicates, maintaining order.
func cdUnionLists(lists [][]string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, l := range lists {
		for _, x := range l {
			if !seen[x] {
				seen[x] = true
				result = append(result, x)
			}
		}
	}
	return result
}

// cdResolveNameCombiner resolves a name in the combiner dict.
func cdResolveName(d *CDCombinerDict, x string) []string {
	if x == "*" {
		var all []string
		for _, e := range d.entries {
			if e.Combiner != nil {
				all = append(all, e.Key)
			}
		}
		return all
	}
	e, ok := d.GetEntry(x)
	if !ok {
		return []string{x}
	}
	if e.List != nil {
		var result []string
		for _, k := range e.List {
			result = append(result, cdResolveName(d, k)...)
		}
		return cdUnionLists([][]string{result})
	}
	return []string{x}
}

// cdResolveConceptName resolves a name in the concept dict.
func cdResolveConceptName(d *CDConceptDict, x string) []string {
	if x == "*" {
		var all []string
		for _, e := range d.Entries() {
			if e.Concept != nil {
				all = append(all, e.Key)
			}
		}
		return all
	}
	e, ok := d.GetEntry(x)
	if !ok {
		return []string{x}
	}
	if e.List != nil {
		var result []string
		for _, k := range e.List {
			result = append(result, cdResolveConceptName(d, k)...)
		}
		return cdUnionLists([][]string{result})
	}
	if e.Set != nil {
		var result []string
		for _, k := range e.Set.Names {
			result = append(result, cdResolveConceptName(d, k)...)
		}
		return cdUnionLists([][]string{result})
	}
	return []string{x}
}

// cdReplaceName replaces occurrences of name with newNames in a list,
// flattening. Non-string elements are kept as-is.
func cdReplaceName(list []string, name string, newNames []string) []string {
	var result []string
	for _, x := range list {
		if x == name {
			result = append(result, newNames...)
		} else {
			result = append(result, x)
		}
	}
	return result
}

// cdReplaceNameInList replaces name with newNames in list, flattening.
func cdReplaceNameInList(list []string, name string, newNames []string) []string {
	return cdReplaceName(list, name, newNames)
}

// cdFilterConceptNames filters concept names using a projection function.
func cdFilterConceptNames(conceptNames [][]string, combination []string, projection func(string, string) bool) [][]string {
	if projection == nil {
		return conceptNames
	}
	result := make([][]string, len(conceptNames))
	for i, ns := range conceptNames {
		var filtered []string
		cat := ""
		if i < len(combination) {
			cat = combination[i]
		}
		for _, n := range ns {
			if projection(n, cat) {
				filtered = append(filtered, n)
			}
		}
		result[i] = filtered
	}
	return result
}

// cartesianProduct computes the cartesian product of string slices.
func cartesianProduct(slices [][]string) [][]string {
	if len(slices) == 0 {
		return [][]string{{}}
	}
	first := slices[0]
	rest := cartesianProduct(slices[1:])
	var result [][]string
	for _, f := range first {
		for _, r := range rest {
			combo := make([]string, 0, 1+len(r))
			combo = append(combo, f)
			combo = append(combo, r...)
			result = append(result, combo)
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// Standard combiners, combinations, factory functions
// ---------------------------------------------------------------------------

// GetStandardCombiners returns the standard set of concept combiners.
func GetStandardCombiners() *CDCombinerDict {
	T := logic.NewTopSort()
	boolSort := logic.Boolean

	unaryRelation := mustFuncSort(T, boolSort)
	binaryRelation := mustFuncSort(T, T, boolSort)

	X := mustVar("X", T)
	Y := mustVar("Y", T)
	Z := mustVar("Z", T)
	U := mustVar("U", unaryRelation)
	U1 := mustVar("U1", unaryRelation)
	U2 := mustVar("U2", unaryRelation)
	B := mustVar("B", binaryRelation)

	result := NewCDCombinerDict()

	// none: ~Exists X. U(X)
	result.SetCombiner("none", MustCDConceptCombiner(
		[]*logic.Var{U},
		mustNot(mustExists([]*logic.Var{X}, mustApplyVar(U, X))),
	))
	// at_least_one: Exists X. U(X)
	result.SetCombiner("at_least_one", MustCDConceptCombiner(
		[]*logic.Var{U},
		mustExists([]*logic.Var{X}, mustApplyVar(U, X)),
	))
	// at_most_one: ForAll X,Y. U(X) & U(Y) => X=Y
	result.SetCombiner("at_most_one", MustCDConceptCombiner(
		[]*logic.Var{U},
		mustForAll([]*logic.Var{X, Y},
			mustImplies(
				mustAnd(mustApplyVar(U, X), mustApplyVar(U, Y)),
				mustEq(X, Y),
			),
		),
	))
	// node_necessarily: ForAll X. U1(X) => U2(X)
	result.SetCombiner("node_necessarily", MustCDConceptCombiner(
		[]*logic.Var{U1, U2},
		mustForAll([]*logic.Var{X},
			mustImplies(mustApplyVar(U1, X), mustApplyVar(U2, X)),
		),
	))
	// node_necessarily_not: ForAll X. U1(X) => ~U2(X)
	result.SetCombiner("node_necessarily_not", MustCDConceptCombiner(
		[]*logic.Var{U1, U2},
		mustForAll([]*logic.Var{X},
			mustImplies(mustApplyVar(U1, X), mustNot(mustApplyVar(U2, X))),
		),
	))
	// mutually_exclusive: ForAll X,Y. ~(U1(X) & U2(Y))
	result.SetCombiner("mutually_exclusive", MustCDConceptCombiner(
		[]*logic.Var{U1, U2},
		mustForAll([]*logic.Var{X, Y},
			mustNot(mustAnd(mustApplyVar(U1, X), mustApplyVar(U2, Y))),
		),
	))
	// all_to_all: ForAll X,Y. U1(X) & U2(Y) => B(X,Y)
	result.SetCombiner("all_to_all", MustCDConceptCombiner(
		[]*logic.Var{B, U1, U2},
		mustForAll([]*logic.Var{X, Y},
			mustImplies(
				mustAnd(mustApplyVar(U1, X), mustApplyVar(U2, Y)),
				mustApplyVar(B, X, Y),
			),
		),
	))
	// none_to_none: ForAll X,Y. U1(X) & U2(Y) => ~B(X,Y)
	result.SetCombiner("none_to_none", MustCDConceptCombiner(
		[]*logic.Var{B, U1, U2},
		mustForAll([]*logic.Var{X, Y},
			mustImplies(
				mustAnd(mustApplyVar(U1, X), mustApplyVar(U2, Y)),
				mustNot(mustApplyVar(B, X, Y)),
			),
		),
	))
	// total: ForAll X. U1(X) => Exists Y. U2(Y) & B(X,Y)
	result.SetCombiner("total", MustCDConceptCombiner(
		[]*logic.Var{B, U1, U2},
		mustForAll([]*logic.Var{X},
			mustImplies(
				mustApplyVar(U1, X),
				mustExists([]*logic.Var{Y},
					mustAnd(mustApplyVar(U2, Y), mustApplyVar(B, X, Y)),
				),
			),
		),
	))
	// functional: ForAll X,Y,Z. U1(X) & U2(Y) & U2(Z) & B(X,Y) & B(X,Z) => Y=Z
	result.SetCombiner("functional", MustCDConceptCombiner(
		[]*logic.Var{B, U1, U2},
		mustForAll([]*logic.Var{X, Y, Z},
			mustImplies(
				mustAnd(
					mustApplyVar(U1, X),
					mustApplyVar(U2, Y),
					mustApplyVar(U2, Z),
					mustApplyVar(B, X, Y),
					mustApplyVar(B, X, Z),
				),
				mustEq(Y, Z),
			),
		),
	))
	// surjective: ForAll Y. U2(Y) => Exists X. U1(X) & B(X,Y)
	result.SetCombiner("surjective", MustCDConceptCombiner(
		[]*logic.Var{B, U1, U2},
		mustForAll([]*logic.Var{Y},
			mustImplies(
				mustApplyVar(U2, Y),
				mustExists([]*logic.Var{X},
					mustAnd(mustApplyVar(U1, X), mustApplyVar(B, X, Y)),
				),
			),
		),
	))
	// injective: ForAll X,Y,Z. U1(X) & U1(Y) & U2(Z) & B(X,Z) & B(Y,Z) => X=Y
	result.SetCombiner("injective", MustCDConceptCombiner(
		[]*logic.Var{B, U1, U2},
		mustForAll([]*logic.Var{X, Y, Z},
			mustImplies(
				mustAnd(
					mustApplyVar(U1, X),
					mustApplyVar(U1, Y),
					mustApplyVar(U2, Z),
					mustApplyVar(B, X, Z),
					mustApplyVar(B, Y, Z),
				),
				mustEq(X, Y),
			),
		),
	))

	// Combiner groups (name-lists).
	result.SetList("node_info", []string{"none", "at_least_one", "at_most_one"})
	result.SetList("edge_info", []string{"all_to_all", "none_to_none"})
	result.SetList("node_label", []string{"node_necessarily", "node_necessarily_not"})

	return result
}

// GetStandardCombinations returns the standard list of combinations.
func GetStandardCombinations() []Combination {
	return []Combination{
		NewCombination("node_info", "node_info", "nodes"),
		NewCombination("edge_info", "edge_info", "edges", "nodes", "nodes"),
		NewCombination("node_label", "node_label", "nodes", "node_labels"),
	}
}

// GetInitialConceptDomain creates a concept domain from a signature.
func GetInitialConceptDomain(sorts map[string]logic.Sort, symbols map[string]*logic.Const) *CDConceptDomain {
	concepts := NewCDConceptDict()

	concepts.SetList("nodes", nil)
	concepts.SetList("node_labels", nil)
	concepts.SetList("edges", nil)

	// Add sort concepts.
	sortNames := make([]string, 0, len(sorts))
	for name := range sorts {
		if name != "bool" {
			sortNames = append(sortNames, name)
		}
	}
	sort.Strings(sortNames)
	for _, name := range sortNames {
		s := sorts[name]
		X := mustVar("X", s)
		eq, _ := logic.NewEq(X, X)
		concepts.SetConcept(name, MustCDConcept(name, []*logic.Var{X}, eq))
		concepts.AppendToList("nodes", name)
	}

	// Add equality concept.
	T := logic.NewTopSort()
	XT := mustVar("X", T)
	YT := mustVar("Y", T)
	eqXY, _ := logic.NewEq(XT, YT)
	concepts.SetConcept("=", MustCDConcept("=", []*logic.Var{XT, YT}, eqXY))

	// Add concepts from symbols.
	symNames := make([]string, 0, len(symbols))
	for name := range symbols {
		symNames = append(symNames, name)
	}
	sort.Strings(symNames)
	for _, sname := range symNames {
		c := symbols[sname]
		if logic.FirstOrderSort(c.CSort) {
			// First-order constant → unary equality concept.
			X := mustVar("X", c.CSort)
			eq, _ := logic.NewEq(X, c)
			name := "=" + c.Name
			concepts.SetConcept(name, MustCDConcept(name, []*logic.Var{X}, eq))
		} else if fs, ok := c.CSort.(*logic.FunctionSort); ok {
			switch fs.Arity() {
			case 1:
				// Unary relation → node_label (e.g., "semaphore")
				X := mustVar("X", fs.Domain()[0])
				app, _ := logic.NewApply(c, X)
				concepts.SetConcept(c.Name, MustCDConcept(c.Name, []*logic.Var{X}, app))
				concepts.AppendToList("node_labels", c.Name)
			case 2:
				// Binary relation → edge (e.g., "link")
				X := mustVar("X", fs.Domain()[0])
				Y := mustVar("Y", fs.Domain()[1])
				app, _ := logic.NewApply(c, X, Y)
				concepts.SetConcept(c.Name, MustCDConcept(c.Name, []*logic.Var{X, Y}, app))
				concepts.AppendToList("edges", c.Name)
			case 3:
				// Ternary relation
				X := mustVar("X", fs.Domain()[0])
				Y := mustVar("Y", fs.Domain()[1])
				Z := mustVar("Z", fs.Domain()[2])
				app, _ := logic.NewApply(c, X, Y, Z)
				concepts.SetConcept(c.Name, MustCDConcept(c.Name, []*logic.Var{X, Y, Z}, app))
			}
		}
	}

	return NewCDConceptDomain(concepts, GetStandardCombiners(), GetStandardCombinations())
}

// GetDiagramConceptDomain creates a concept domain from a signature and diagram.
func GetDiagramConceptDomain(sorts map[string]logic.Sort, symbols []*logic.Const, diagram logic.Node) *CDConceptDomain {
	concepts := NewCDConceptDict()

	concepts.SetList("nodes", nil)
	concepts.SetList("node_labels", nil)
	concepts.SetList("edges", nil)

	// Add equality concept.
	T := logic.NewTopSort()
	XT := mustVar("X", T)
	YT := mustVar("Y", T)
	eqXY, _ := logic.NewEq(XT, YT)
	concepts.SetConcept("=", MustCDConcept("=", []*logic.Var{XT, YT}, eqXY))

	// Merge signature symbols with diagram constants.
	allConsts := make(map[string]*logic.Const)
	for _, c := range symbols {
		allConsts[c.Name] = c
	}
	if diagram != nil {
		for _, cNode := range logicutil.UsedConstants(diagram) {
			c := cNode.(*logic.Const)
			if _, exists := allConsts[c.Name]; !exists {
				allConsts[c.Name] = c
			}
		}
	}

	constNames := make([]string, 0, len(allConsts))
	for name := range allConsts {
		constNames = append(constNames, name)
	}
	sort.Strings(constNames)

	for _, sname := range constNames {
		c := allConsts[sname]
		if logic.FirstOrderSort(c.CSort) {
			X := mustVar("X", c.CSort)
			eq, _ := logic.NewEq(X, c)
			name := fmt.Sprintf("%s:%s", c.Name, c.CSort)
			concepts.SetConcept(name, MustCDConcept(name, []*logic.Var{X}, eq))
			concepts.AppendToList("nodes", name)
		} else if fs, ok := c.CSort.(*logic.FunctionSort); ok {
			switch fs.Arity() {
			case 1:
				X := mustVar("X", fs.Domain()[0])
				app, _ := logic.NewApply(c, X)
				concepts.SetConcept(c.Name, MustCDConcept(c.Name, []*logic.Var{X}, app))
			case 2:
				X := mustVar("X", fs.Domain()[0])
				Y := mustVar("Y", fs.Domain()[1])
				app, _ := logic.NewApply(c, X, Y)
				concepts.SetConcept(c.Name, MustCDConcept(c.Name, []*logic.Var{X, Y}, app))
			case 3:
				X := mustVar("X", fs.Domain()[0])
				Y := mustVar("Y", fs.Domain()[1])
				Z := mustVar("Z", fs.Domain()[2])
				app, _ := logic.NewApply(c, X, Y, Z)
				concepts.SetConcept(c.Name, MustCDConcept(c.Name, []*logic.Var{X, Y, Z}, app))
			}
		}
	}

	return NewCDConceptDomain(concepts, GetStandardCombiners(), GetStandardCombinations())
}

// UniverseElementToConceptName converts a universe element constant to its
// concept name.
func UniverseElementToConceptName(uc *logic.Const) string {
	name := uc.Name
	sortStr := uc.CSort.String()
	if !strings.Contains(name, sortStr) {
		name += ":" + sortStr
	}
	return name
}

// GetStructureConceptDomain creates a concept domain from a state with a universe.
// state is a formula, universe maps sort names to element constants,
// sig provides additional symbol information.
func GetStructureConceptDomain(
	stateFormula logic.Node,
	universe map[string][]*logic.Const,
	sigSymbols map[string]*logic.Const,
) *CDConceptDomain {
	concepts := NewCDConceptDict()

	concepts.SetList("nodes", nil)
	concepts.SetList("node_labels", nil)

	// Add equality concept.
	T := logic.NewTopSort()
	XT := mustVar("X", T)
	YT := mustVar("Y", T)
	eqXY, _ := logic.NewEq(XT, YT)
	concepts.SetConcept("=", MustCDConcept("=", []*logic.Var{XT, YT}, eqXY))

	// Add nodes for universe elements.
	var elements []*logic.Const
	elementSet := make(map[string]bool)
	for _, ucs := range universe {
		for _, uc := range ucs {
			elements = append(elements, uc)
			elementSet[uc.Name] = true
		}
	}
	sort.Slice(elements, func(i, j int) bool { return elements[i].Name < elements[j].Name })

	for _, uc := range elements {
		X := mustVar("X", uc.CSort)
		name := UniverseElementToConceptName(uc)
		eq, _ := logic.NewEq(X, uc)
		concepts.SetConcept(name, MustCDConcept(name, []*logic.Var{X}, eq))
		concepts.AppendToList("nodes", name)
	}

	// Collect all symbols from state formula and signature.
	allSymbols := make(map[string]*logic.Const)
	if stateFormula != nil {
		for _, cNode := range logicutil.UsedConstants(stateFormula) {
			c := cNode.(*logic.Const)
			if !elementSet[c.Name] {
				allSymbols[c.Name] = c
			}
		}
	}
	for name, c := range sigSymbols {
		if !elementSet[name] {
			allSymbols[name] = c
		}
	}

	symNames := make([]string, 0, len(allSymbols))
	for name := range allSymbols {
		symNames = append(symNames, name)
	}
	sort.Strings(symNames)

	for _, sname := range symNames {
		c := allSymbols[sname]

		if logic.FirstOrderSort(c.CSort) {
			X := mustVar("X", c.CSort)
			eq, _ := logic.NewEq(X, c)
			name := "=" + c.Name
			concepts.SetConcept(name, MustCDConcept(name, []*logic.Var{X}, eq))
		} else if fs, ok := c.CSort.(*logic.FunctionSort); ok {
			if logic.SortEqual(fs.Range(), logic.Boolean) {
				// Relation
				switch fs.Arity() {
				case 1:
					X := mustVar("X", fs.Domain()[0])
					app, _ := logic.NewApply(c, X)
					concepts.SetConcept(c.Name, MustCDConcept(c.Name, []*logic.Var{X}, app))
				case 2:
					X := mustVar("X", fs.Domain()[0])
					Y := mustVar("Y", fs.Domain()[1])
					app, _ := logic.NewApply(c, X, Y)
					concepts.SetConcept(c.Name, MustCDConcept(c.Name, []*logic.Var{X, Y}, app))
				case 3:
					X := mustVar("X", fs.Domain()[0])
					Y := mustVar("Y", fs.Domain()[1])
					Z := mustVar("Z", fs.Domain()[2])
					app, _ := logic.NewApply(c, X, Y, Z)
					concepts.SetConcept(c.Name, MustCDConcept(c.Name, []*logic.Var{X, Y, Z}, app))
				}
			} else {
				// Function
				switch fs.Arity() {
				case 1:
					X := mustVar("X", fs.Domain()[0])
					Y := mustVar("Y", fs.Range())
					app, _ := logic.NewApply(c, X)
					eq, _ := logic.NewEq(app, Y)
					concepts.SetConcept(c.Name, MustCDConcept(c.Name, []*logic.Var{X, Y}, eq))
				case 2:
					X := mustVar("X", fs.Domain()[0])
					Y := mustVar("Y", fs.Domain()[1])
					Z := mustVar("Z", fs.Range())
					app, _ := logic.NewApply(c, X, Y)
					eq, _ := logic.NewEq(app, Z)
					concepts.SetConcept(c.Name, MustCDConcept(c.Name, []*logic.Var{X, Y, Z}, eq))
				}
			}
		}
	}

	return NewCDConceptDomain(concepts, GetStandardCombiners(), GetStandardCombinations())
}

// GetStructureConceptAbstractValue computes abstract values for a structure
// (a state with a universe), using direct analysis of the state formula
// rather than Z3.
func GetStructureConceptAbstractValue(
	stateFormula logic.Node,
	universe map[string][]*logic.Const,
) map[string]bool {
	result := make(map[string]bool)

	// Map element constants to their concept names.
	nodes := make(map[string]string) // const.Name -> concept name
	var elements []*logic.Const
	for _, ucs := range universe {
		elements = append(elements, ucs...)
	}
	sort.Slice(elements, func(i, j int) bool { return elements[i].Name < elements[j].Name })

	for _, uc := range elements {
		name := UniverseElementToConceptName(uc)
		nodes[uc.Name] = name
		result[TagString(Tag{"node_info", "none", name})] = false
		result[TagString(Tag{"node_info", "at_least_one", name})] = true
		result[TagString(Tag{"node_info", "at_most_one", name})] = true
	}

	// Analyze state formula literals.
	andNode, ok := stateFormula.(*logic.And)
	if !ok {
		return result
	}
	for _, lit := range andNode.Terms {
		if _, ok := lit.(*logic.ForAll); ok {
			continue // universe constraint
		}

		polarity := true
		innerLit := lit
		if notNode, ok := lit.(*logic.Not); ok {
			polarity = false
			innerLit = notNode.Body
		}

		switch l := innerLit.(type) {
		case *logic.Apply:
			if fs, ok := l.Func.(*logic.Const); ok {
				fSort, ok2 := fs.CSort.(*logic.FunctionSort)
				if !ok2 {
					continue
				}
				if fSort.Arity() == 1 {
					if t0, ok := l.Terms[0].(*logic.Const); ok {
						if nodeName, ok := nodes[t0.Name]; ok {
							labelName := fs.Name
							result[TagString(Tag{"node_label", "node_necessarily", nodeName, labelName})] = polarity
							result[TagString(Tag{"node_label", "node_necessarily_not", nodeName, labelName})] = !polarity
						}
					}
				} else if fSort.Arity() == 2 {
					t0, ok0 := l.Terms[0].(*logic.Const)
					t1, ok1 := l.Terms[1].(*logic.Const)
					if ok0 && ok1 {
						sn, sok := nodes[t0.Name]
						tn, tok := nodes[t1.Name]
						if sok && tok {
							edgeName := fs.Name
							result[TagString(Tag{"edge_info", "all_to_all", edgeName, sn, tn})] = polarity
							result[TagString(Tag{"edge_info", "none_to_none", edgeName, sn, tn})] = !polarity
						}
					}
				}
			}
		case *logic.Eq:
			// Handle equality literals for functions.
			if _, ok := l.T1.(*logic.Const); ok {
				// Simple constant equality -- skip for now
			}
		}
	}

	return result
}

// GetStructureRenaming generates prettier names for universe constants
// based on topological sort of order relations.
func GetStructureRenaming(
	stateFormula logic.Node,
	universe map[string][]*logic.Const,
	orderRelations map[string]bool,
) map[string]string {
	var elements []*logic.Const
	seen := make(map[string]bool)
	for _, ucs := range universe {
		for _, uc := range ucs {
			if !seen[uc.Name] {
				seen[uc.Name] = true
				elements = append(elements, uc)
			}
		}
	}

	// Extract order from state formula.
	var order [][2]*logic.Const
	if andNode, ok := stateFormula.(*logic.And); ok {
		for _, lit := range andNode.Terms {
			if app, ok := lit.(*logic.Apply); ok {
				fc, ok2 := app.Func.(*logic.Const)
				if !ok2 {
					continue
				}
				fs, ok3 := fc.CSort.(*logic.FunctionSort)
				if !ok3 || fs.Arity() != 2 {
					continue
				}
				if !logic.SortEqual(fs.Domain()[0], fs.Domain()[1]) {
					continue
				}
				if !logic.SortEqual(fs.Range(), logic.Boolean) {
					continue
				}
				if !orderRelations[fc.Name] {
					continue
				}
				if len(app.Terms) == 2 {
					t0, ok0 := app.Terms[0].(*logic.Const)
					t1, ok1 := app.Terms[1].(*logic.Const)
					if ok0 && ok1 {
						order = append(order, [2]*logic.Const{t0, t1})
					}
				}
			}
		}
	}

	// Topological sort using the order relations.
	var orderPairs [][2]*logic.Const
	nameSet := make(map[string]*logic.Const)
	for _, elem := range elements {
		nameSet[elem.Name] = elem
	}
	for _, pair := range order {
		e0, ok0 := nameSet[pair[0].Name]
		e1, ok1 := nameSet[pair[1].Name]
		if ok0 && ok1 {
			orderPairs = append(orderPairs, [2]*logic.Const{e0, e1})
		}
	}
	sorted := ivyutils.TopologicalSort(elements, orderPairs, func(elem *logic.Const) string { return elem.Name })

	result := make(map[string]string)
	count := make(map[string]int)
	for _, c := range sorted {
		prefix := strings.ToLower(c.CSort.String())
		result[c.Name] = fmt.Sprintf("%s%d", prefix, count[prefix])
		count[prefix]++
	}
	return result
}

// ---------------------------------------------------------------------------
// Logic construction helpers (must* panic on error -- only for known-good formulas)
// ---------------------------------------------------------------------------

func mustVar(name string, s logic.Sort) *logic.Var {
	v, err := logic.NewVar(name, s)
	if err != nil {
		panic(err)
	}
	return v
}

func mustFuncSort(sorts ...logic.Sort) *logic.FunctionSort {
	fs, err := logic.NewFunctionSort(sorts...)
	if err != nil {
		panic(err)
	}
	return fs
}

func mustApplyVar(v *logic.Var, args ...logic.Node) logic.Node {
	n, err := v.Call(args...)
	if err != nil {
		panic(err)
	}
	return n
}

func mustNot(body logic.Node) logic.Node {
	n, _ := logic.NewNot(body)
	return n
}

func mustAnd(terms ...logic.Node) logic.Node {
	n, _ := logic.NewAnd(terms...)
	return n
}

func mustOr(terms ...logic.Node) logic.Node {
	n, _ := logic.NewOr(terms...)
	return n
}

func mustEq(t1, t2 logic.Node) logic.Node {
	n, _ := logic.NewEq(t1, t2)
	return n
}

func mustImplies(t1, t2 logic.Node) logic.Node {
	n, _ := logic.NewImplies(t1, t2)
	return n
}

func mustForAll(vars []*logic.Var, body logic.Node) logic.Node {
	n, _ := logic.NewForAll(vars, body)
	return n
}

func mustExists(vars []*logic.Var, body logic.Node) logic.Node {
	n, _ := logic.NewExists(vars, body)
	return n
}
