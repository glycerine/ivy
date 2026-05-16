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
	goivy "github.com/glycerine/ivy/goivy"
	"sort"
	"strings"
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
	Variables []*goivy.LogicVariable
	Formula   goivy.Expr
}

// NewCDConcept creates a new concept, validating that all variables are
// first-order. Returns an error if validation fails.
func NewCDConcept(name string, variables []*goivy.LogicVariable, formula goivy.Expr) (*CDConcept, error) {
	if name == "" {
		return nil, &goivy.IvyError{Msg: "Concept name is empty"}
	}
	vars := make([]*goivy.LogicVariable, len(variables))
	copy(vars, variables)
	for _, v := range vars {
		if !goivy.FirstOrderSort(v.VSort) {
			return nil, &goivy.IvyError{Msg: fmt.Sprintf("Concept variables must be first-order: %v", v)}
		}
	}
	return &CDConcept{Name: name, Variables: vars, Formula: formula}, nil
}

// MustCDConcept is like NewCDConcept but panics on error.
func MustCDConcept(name string, variables []*goivy.LogicVariable, formula goivy.Expr) *CDConcept {
	c, err := NewCDConcept(name, variables, formula)
	if err != nil {
		panic(err)
	}
	return c
}

// Arity returns the number of variables (parameters) of the concept.
func (c *CDConcept) Arity() int { return len(c.Variables) }

// Sorts returns the sort of each variable.
func (c *CDConcept) Sorts() []goivy.Sort {
	s := make([]goivy.Sort, len(c.Variables))
	for i, v := range c.Variables {
		s[i] = v.VSort
	}
	return s
}

// Sort returns the sort of the first variable.
func (c *CDConcept) Sort() goivy.Sort {
	if len(c.Variables) == 0 {
		return nil
	}
	return c.Variables[0].VSort
}

// Call substitutes the variables with the given terms, returning a formula.
// This corresponds to Python's Concept.__call__.
func (c *CDConcept) Call(terms ...goivy.Expr) (goivy.Expr, error) {
	if len(terms) != c.Arity() {
		return nil, &ArityError{Expected: c.Arity(), Got: len(terms)}
	}
	subs := make(map[goivy.NodeKey]goivy.Expr)
	for i, v := range c.Variables {
		t := terms[i]
		// Concretize sorts to match the variable's sort.
		ct, err := goivy.ConcretizeSorts(t, v.VSort)
		if err != nil {
			ct = t // fallback: use as-is
		}
		subs[goivy.Key(v)] = ct
	}
	return goivy.Substitute(c.Formula, subs)
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
	Variables []*goivy.LogicVariable
	Formula   goivy.Expr
}

// NewCDConceptCombiner creates a combiner, validating that all variables
// are relational (FunctionSort with Boolean range).
func NewCDConceptCombiner(variables []*goivy.LogicVariable, formula goivy.Expr) (*CDConceptCombiner, error) {
	vars := make([]*goivy.LogicVariable, len(variables))
	copy(vars, variables)
	for _, v := range vars {
		fs, ok := v.VSort.(*goivy.LogicFunctionSort)
		if !ok || !goivy.SortEqual(fs.Range(), goivy.Boolean) {
			return nil, &goivy.IvyError{Msg: fmt.Sprintf("ConceptCombiner variables must be relational: %v (sort: %s)", v, v.VSort)}
		}
	}
	return &CDConceptCombiner{Variables: vars, Formula: formula}, nil
}

// MustCDConceptCombiner is like NewCDConceptCombiner but panics on error.
func MustCDConceptCombiner(variables []*goivy.LogicVariable, formula goivy.Expr) *CDConceptCombiner {
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
		if fs, ok := v.VSort.(*goivy.LogicFunctionSort); ok {
			a[i] = len(fs.Domain())
		}
	}
	return a
}

// Call instantiates the combiner with the given concepts.
// This corresponds to Python's ConceptCombiner.__call__, which uses
// substitute_apply.
func (cc *CDConceptCombiner) Call(concepts ...*CDConcept) (goivy.Expr, error) {
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
	result, err := substituteApplyNode(cc.Formula, cc.Variables, concepts)
	if err != nil {
		return nil, err
	}
	// Concretize sorts — replaces TopSort with inferred concrete sorts.
	// If this fails, the formula still has TopSort and cannot be sent to Z3.
	cr, err := goivy.ConcretizeSorts(result, nil)
	if err != nil {
		return nil, fmt.Errorf("concretize failed: %w", err)
	}
	// Verify no TopSort remains — Z3 panics on sort mismatches.
	if goivy.ContainsTopSort(cr) {
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
func substituteApplyNode(node goivy.Expr, variables []*goivy.LogicVariable, concepts []*CDConcept) (goivy.Expr, error) {
	if node == nil {
		return nil, fmt.Errorf("concept combiner formula is nil")
	}
	// Build a mapping from variable name -> concept (by_name matching like Python).
	varMap := make(map[string]*CDConcept)
	for i, v := range variables {
		varMap[v.Name] = concepts[i]
	}
	return substApplyRec(node, varMap)
}

func substApplyRec(node goivy.Expr, varMap map[string]*CDConcept) (goivy.Expr, error) {
	if node == nil {
		return nil, fmt.Errorf("nil expression in concept combiner substitution")
	}
	switch n := node.(type) {
	case *goivy.Apply:
		// Check if function is a variable that should be replaced.
		if v, ok := n.Func.(*goivy.LogicVariable); ok {
			if concept, found := varMap[v.Name]; found {
				// Replace V(args...) with concept.formula[concept.vars -> args]
				args := make([]goivy.Expr, len(n.Terms))
				for i, t := range n.Terms {
					arg, err := substApplyRec(t, varMap)
					if err != nil {
						return nil, err
					}
					args[i] = arg
				}
				result, err := concept.Call(args...)
				if err != nil {
					return nil, fmt.Errorf("substitute concept %q: %w", concept.Name, err)
				}
				return result, nil
			}
		}
		// Recurse on function and terms.
		newFunc, err := substApplyRec(n.Func, varMap)
		if err != nil {
			return nil, err
		}
		newTerms := make([]goivy.Expr, len(n.Terms))
		for i, t := range n.Terms {
			term, err := substApplyRec(t, varMap)
			if err != nil {
				return nil, err
			}
			newTerms[i] = term
		}
		result, err := goivy.NewApply(newFunc, newTerms...)
		if err != nil {
			return nil, fmt.Errorf("rebuild apply %s: %w", n, err)
		}
		return result, nil

	case *goivy.Eq:
		t1, err := substApplyRec(n.T1, varMap)
		if err != nil {
			return nil, err
		}
		t2, err := substApplyRec(n.T2, varMap)
		if err != nil {
			return nil, err
		}
		result, err := goivy.NewEq(t1, t2)
		if err != nil {
			return nil, fmt.Errorf("rebuild equality %s: %w", n, err)
		}
		return result, nil

	case *goivy.LogicNot:
		b, err := substApplyRec(n.Body, varMap)
		if err != nil {
			return nil, err
		}
		result, err := goivy.NewNot(b)
		if err != nil {
			return nil, fmt.Errorf("rebuild negation %s: %w", n, err)
		}
		return result, nil

	case *goivy.LogicAnd:
		terms, err := substApplySlice(n.Terms, varMap)
		if err != nil {
			return nil, err
		}
		result, err := goivy.NewAnd(terms...)
		if err != nil {
			return nil, fmt.Errorf("rebuild conjunction %s: %w", n, err)
		}
		return result, nil

	case *goivy.LogicOr:
		terms, err := substApplySlice(n.Terms, varMap)
		if err != nil {
			return nil, err
		}
		result, err := goivy.NewOr(terms...)
		if err != nil {
			return nil, fmt.Errorf("rebuild disjunction %s: %w", n, err)
		}
		return result, nil

	case *goivy.LogicImplies:
		t1, err := substApplyRec(n.T1, varMap)
		if err != nil {
			return nil, err
		}
		t2, err := substApplyRec(n.T2, varMap)
		if err != nil {
			return nil, err
		}
		if t1 == n.T1 && t2 == n.T2 {
			return node, nil
		}
		result, err := goivy.NewImplies(t1, t2)
		if err != nil {
			return nil, fmt.Errorf("rebuild implication %s: %w", n, err)
		}
		return result, nil

	case *goivy.ForAll:
		b, err := substApplyRec(n.Body, varMap)
		if err != nil {
			return nil, err
		}
		if b == n.Body {
			return node, nil
		}
		result, err := goivy.NewForAll(n.Variables, b)
		if err != nil {
			return nil, fmt.Errorf("rebuild forall %s: %w", n, err)
		}
		return result, nil

	case *goivy.LogicExists:
		b, err := substApplyRec(n.Body, varMap)
		if err != nil {
			return nil, err
		}
		if b == n.Body {
			return node, nil
		}
		result, err := goivy.NewExists(n.Variables, b)
		if err != nil {
			return nil, fmt.Errorf("rebuild exists %s: %w", n, err)
		}
		return result, nil

	case *goivy.LogicVariable:
		// A bare variable (not applied) that's in the map:
		// This means U used as a standalone, treat as U() but concepts
		// don't support 0-arity this way. Return as-is.
		return node, nil

	default:
		return node, nil
	}
}

func substApplySlice(terms []goivy.Expr, varMap map[string]*CDConcept) ([]goivy.Expr, error) {
	newTerms := make([]goivy.Expr, len(terms))
	for i, t := range terms {
		term, err := substApplyRec(t, varMap)
		if err != nil {
			return nil, err
		}
		newTerms[i] = term
	}
	return newTerms, nil
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
	Formula goivy.Expr
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
func (d *CDConceptDomain) GetCombFacts(combinationName, combinerClass string, conceptNames [][]string, facts *[]Fact) error {
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
				if err == nil {
					err = fmt.Errorf("combiner returned nil formula")
				}
				return fmt.Errorf("concept fact %q/%q on %v: %w", combinationName, combinerName, conceptCombo, err)
			}
			tag := make(Tag, 0, 2+len(conceptCombo))
			tag = append(tag, combinationName, combinerName)
			tag = append(tag, conceptCombo...)
			*facts = append(*facts, Fact{Tag: tag, Formula: formula})
		}
	}
	return nil
}

// GetFacts returns all facts from all combinations.
// If projection is non-nil, it filters which concept names to include.
// Signature: projection(conceptName, categoryName) bool.
//
// Special-cases "edge_info", "node_label", and "enum" for performance.
func (d *CDConceptDomain) GetFacts(projection func(string, string) bool) ([]Fact, error) {
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
				// Match Python defaultdict auto-vivification
				s0 := ec.Sorts()[0].String()
				if _, ok := nodesBySortName[s0]; !ok {
					nodesBySortName[s0] = nil
				}
				s1 := ec.Sorts()[1].String()
				if _, ok := nodesBySortName[s1]; !ok {
					nodesBySortName[s1] = nil
				}
				c0 := nodesBySortName[s0]
				c1 := nodesBySortName[s1]
				if err := d.GetCombFacts("edge_info", "edge_info", [][]string{{e}, c0, c1}, &facts); err != nil {
					return nil, err
				}
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
					// Match Python defaultdict auto-vivification
					lcS0 := lc.Sorts()[0].String()
					if _, ok := nodesBySortName[lcS0]; !ok {
						nodesBySortName[lcS0] = nil
					}
					c0 := nodesBySortName[lcS0]
					if err := d.GetCombFacts("node_label", "node_label", [][]string{c0, {cname}}, &facts); err != nil {
						return nil, err
					}
				}
			}
		} else {
			conceptSets := combination.ConceptSets()
			conceptNames := make([][]string, len(conceptSets))
			for i, x := range conceptSets {
				conceptNames[i] = cdResolveConceptName(d.Concepts, x)
			}
			conceptNames = cdFilterConceptNames(conceptNames, conceptSets, projection)
			if err := d.GetCombFacts(combinationName, combination.CombinerClass(), conceptNames, &facts); err != nil {
				return nil, err
			}
		}
	}
	return facts, nil
}

// Split splits a concept into sub-concepts based on split_by.
func (d *CDConceptDomain) Split(concept, splitBy string) error {
	c1 := d.Concepts.GetConcept(concept)
	if c1 == nil {
		return fmt.Errorf("split %q by %q: concept not found", concept, splitBy)
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
				return fmt.Errorf("split %q by %q: %w", concept, splitBy, err)
			}
			newName := fmt.Sprintf("(%s+%s)", concept, n)
			andF, err := goivy.NewAnd(c1.Formula, f)
			if err != nil {
				return fmt.Errorf("split %q by %q: %w", concept, splitBy, err)
			}
			newConcept, err := NewCDConcept(newName, variables, andF)
			if err != nil {
				return fmt.Errorf("split %q by %q: %w", concept, splitBy, err)
			}
			newConcepts = append(newConcepts, newConcept)
			newNames = append(newNames, newName)
		}
	} else {
		// splitting by a single concept
		c2 := d.Concepts.GetConcept(splitBy)
		if c2 == nil {
			return fmt.Errorf("split %q by %q: split concept not found", concept, splitBy)
		}
		posF, err := c2.Call(nodesToSlice(variables)...)
		if err != nil {
			return fmt.Errorf("split %q by %q: %w", concept, splitBy, err)
		}
		negF, err := goivy.NewNot(posF)
		if err != nil {
			return fmt.Errorf("split %q by %q: %w", concept, splitBy, err)
		}
		posAnd, err := goivy.NewAnd(c1.Formula, posF)
		if err != nil {
			return fmt.Errorf("split %q by %q: %w", concept, splitBy, err)
		}
		negAnd, err := goivy.NewAnd(c1.Formula, negF)
		if err != nil {
			return fmt.Errorf("split %q by %q: %w", concept, splitBy, err)
		}

		posName := fmt.Sprintf("(%s+%s)", concept, splitBy)
		negName := fmt.Sprintf("(%s-%s)", concept, splitBy)

		posConcept, err := NewCDConcept(posName, variables, posAnd)
		if err != nil {
			return fmt.Errorf("split %q by %q: %w", concept, splitBy, err)
		}
		negConcept, err := NewCDConcept(negName, variables, negAnd)
		if err != nil {
			return fmt.Errorf("split %q by %q: %w", concept, splitBy, err)
		}
		newConcepts = append(newConcepts, posConcept, negConcept)
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
	return nil
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

// nodesToSlice converts a []*logic.Variable to []logic.Expr.
func nodesToSlice(vars []*goivy.LogicVariable) []goivy.Expr {
	nodes := make([]goivy.Expr, len(vars))
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
	T := goivy.NewTopSort()
	boolSort := goivy.Boolean

	unaryRelation := webuiMustFuncSort(T, boolSort)
	binaryRelation := webuiMustFuncSort(T, T, boolSort)

	X := webuiMustVar("X", T)
	Y := webuiMustVar("Y", T)
	Z := webuiMustVar("Z", T)
	U := webuiMustVar("U", unaryRelation)
	U1 := webuiMustVar("U1", unaryRelation)
	U2 := webuiMustVar("U2", unaryRelation)
	B := webuiMustVar("B", binaryRelation)

	result := NewCDCombinerDict()

	// none: ~Exists X. U(X)
	result.SetCombiner("none", MustCDConceptCombiner(
		[]*goivy.LogicVariable{U},
		mustNot(webuiMustExists([]*goivy.LogicVariable{X}, mustApplyVar(U, X))),
	))
	// at_least_one: Exists X. U(X)
	result.SetCombiner("at_least_one", MustCDConceptCombiner(
		[]*goivy.LogicVariable{U},
		webuiMustExists([]*goivy.LogicVariable{X}, mustApplyVar(U, X)),
	))
	// at_most_one: ForAll X,Y. U(X) & U(Y) => X=Y
	result.SetCombiner("at_most_one", MustCDConceptCombiner(
		[]*goivy.LogicVariable{U},
		webuiMustForAll([]*goivy.LogicVariable{X, Y},
			mustImplies(
				mustAnd(mustApplyVar(U, X), mustApplyVar(U, Y)),
				webuiMustEq(X, Y),
			),
		),
	))
	// node_necessarily: ForAll X. U1(X) => U2(X)
	result.SetCombiner("node_necessarily", MustCDConceptCombiner(
		[]*goivy.LogicVariable{U1, U2},
		webuiMustForAll([]*goivy.LogicVariable{X},
			mustImplies(mustApplyVar(U1, X), mustApplyVar(U2, X)),
		),
	))
	// node_necessarily_not: ForAll X. U1(X) => ~U2(X)
	result.SetCombiner("node_necessarily_not", MustCDConceptCombiner(
		[]*goivy.LogicVariable{U1, U2},
		webuiMustForAll([]*goivy.LogicVariable{X},
			mustImplies(mustApplyVar(U1, X), mustNot(mustApplyVar(U2, X))),
		),
	))
	// mutually_exclusive: ForAll X,Y. ~(U1(X) & U2(Y))
	result.SetCombiner("mutually_exclusive", MustCDConceptCombiner(
		[]*goivy.LogicVariable{U1, U2},
		webuiMustForAll([]*goivy.LogicVariable{X, Y},
			mustNot(mustAnd(mustApplyVar(U1, X), mustApplyVar(U2, Y))),
		),
	))
	// all_to_all: ForAll X,Y. U1(X) & U2(Y) => B(X,Y)
	result.SetCombiner("all_to_all", MustCDConceptCombiner(
		[]*goivy.LogicVariable{B, U1, U2},
		webuiMustForAll([]*goivy.LogicVariable{X, Y},
			mustImplies(
				mustAnd(mustApplyVar(U1, X), mustApplyVar(U2, Y)),
				mustApplyVar(B, X, Y),
			),
		),
	))
	// none_to_none: ForAll X,Y. U1(X) & U2(Y) => ~B(X,Y)
	result.SetCombiner("none_to_none", MustCDConceptCombiner(
		[]*goivy.LogicVariable{B, U1, U2},
		webuiMustForAll([]*goivy.LogicVariable{X, Y},
			mustImplies(
				mustAnd(mustApplyVar(U1, X), mustApplyVar(U2, Y)),
				mustNot(mustApplyVar(B, X, Y)),
			),
		),
	))
	// total: ForAll X. U1(X) => Exists Y. U2(Y) & B(X,Y)
	result.SetCombiner("total", MustCDConceptCombiner(
		[]*goivy.LogicVariable{B, U1, U2},
		webuiMustForAll([]*goivy.LogicVariable{X},
			mustImplies(
				mustApplyVar(U1, X),
				webuiMustExists([]*goivy.LogicVariable{Y},
					mustAnd(mustApplyVar(U2, Y), mustApplyVar(B, X, Y)),
				),
			),
		),
	))
	// functional: ForAll X,Y,Z. U1(X) & U2(Y) & U2(Z) & B(X,Y) & B(X,Z) => Y=Z
	result.SetCombiner("functional", MustCDConceptCombiner(
		[]*goivy.LogicVariable{B, U1, U2},
		webuiMustForAll([]*goivy.LogicVariable{X, Y, Z},
			mustImplies(
				mustAnd(
					mustApplyVar(U1, X),
					mustApplyVar(U2, Y),
					mustApplyVar(U2, Z),
					mustApplyVar(B, X, Y),
					mustApplyVar(B, X, Z),
				),
				webuiMustEq(Y, Z),
			),
		),
	))
	// surjective: ForAll Y. U2(Y) => Exists X. U1(X) & B(X,Y)
	result.SetCombiner("surjective", MustCDConceptCombiner(
		[]*goivy.LogicVariable{B, U1, U2},
		webuiMustForAll([]*goivy.LogicVariable{Y},
			mustImplies(
				mustApplyVar(U2, Y),
				webuiMustExists([]*goivy.LogicVariable{X},
					mustAnd(mustApplyVar(U1, X), mustApplyVar(B, X, Y)),
				),
			),
		),
	))
	// injective: ForAll X,Y,Z. U1(X) & U1(Y) & U2(Z) & B(X,Z) & B(Y,Z) => X=Y
	result.SetCombiner("injective", MustCDConceptCombiner(
		[]*goivy.LogicVariable{B, U1, U2},
		webuiMustForAll([]*goivy.LogicVariable{X, Y, Z},
			mustImplies(
				mustAnd(
					mustApplyVar(U1, X),
					mustApplyVar(U1, Y),
					mustApplyVar(U2, Z),
					mustApplyVar(B, X, Z),
					mustApplyVar(B, Y, Z),
				),
				webuiMustEq(X, Y),
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

// GetInitialConceptDomainE creates a concept domain from a signature.
func GetInitialConceptDomainE(sorts map[string]goivy.Sort, symbols map[string]*goivy.Const) (*CDConceptDomain, error) {
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
		X := webuiMustVar("X", s)
		eq, err := goivy.NewEq(X, X)
		if err != nil {
			return nil, fmt.Errorf("initial concept domain: sort %q: %w", name, err)
		}
		concept, err := NewCDConcept(name, []*goivy.LogicVariable{X}, eq)
		if err != nil {
			return nil, fmt.Errorf("initial concept domain: sort %q: %w", name, err)
		}
		concepts.SetConcept(name, concept)
		concepts.AppendToList("nodes", name)
	}

	// Add equality concept.
	T := goivy.NewTopSort()
	XT := webuiMustVar("X", T)
	YT := webuiMustVar("Y", T)
	eqXY, err := goivy.NewEq(XT, YT)
	if err != nil {
		return nil, fmt.Errorf("initial concept domain: equality: %w", err)
	}
	eqConcept, err := NewCDConcept("=", []*goivy.LogicVariable{XT, YT}, eqXY)
	if err != nil {
		return nil, fmt.Errorf("initial concept domain: equality: %w", err)
	}
	concepts.SetConcept("=", eqConcept)

	// Add concepts from symbols.
	symNames := make([]string, 0, len(symbols))
	for name := range symbols {
		symNames = append(symNames, name)
	}
	sort.Strings(symNames)
	for _, sname := range symNames {
		c := symbols[sname]
		if goivy.FirstOrderSort(c.CSort) {
			// First-order constant → unary equality concept.
			X := webuiMustVar("X", c.CSort)
			eq, err := goivy.NewEq(X, c)
			if err != nil {
				return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
			}
			name := "=" + c.Name
			concept, err := NewCDConcept(name, []*goivy.LogicVariable{X}, eq)
			if err != nil {
				return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
			}
			concepts.SetConcept(name, concept)
		} else if fs, ok := c.CSort.(*goivy.LogicFunctionSort); ok {
			if goivy.SortEqual(fs.Range(), goivy.Boolean) {
				switch fs.Arity() {
				case 1:
					// Unary relation → node_label (e.g., "semaphore")
					X := webuiMustVar("X", fs.Domain()[0])
					app, err := goivy.NewApply(c, X)
					if err != nil {
						return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
					}
					concept, err := NewCDConcept(c.Name, []*goivy.LogicVariable{X}, app)
					if err != nil {
						return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
					}
					concepts.SetConcept(c.Name, concept)
					concepts.AppendToList("node_labels", c.Name)
				case 2:
					// Binary relation → edge (e.g., "link")
					X := webuiMustVar("X", fs.Domain()[0])
					Y := webuiMustVar("Y", fs.Domain()[1])
					app, err := goivy.NewApply(c, X, Y)
					if err != nil {
						return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
					}
					concept, err := NewCDConcept(c.Name, []*goivy.LogicVariable{X, Y}, app)
					if err != nil {
						return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
					}
					concepts.SetConcept(c.Name, concept)
					concepts.AppendToList("edges", c.Name)
				case 3:
					// Ternary relation.
					X := webuiMustVar("X", fs.Domain()[0])
					Y := webuiMustVar("Y", fs.Domain()[1])
					Z := webuiMustVar("Z", fs.Domain()[2])
					app, err := goivy.NewApply(c, X, Y, Z)
					if err != nil {
						return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
					}
					concept, err := NewCDConcept(c.Name, []*goivy.LogicVariable{X, Y, Z}, app)
					if err != nil {
						return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
					}
					concepts.SetConcept(c.Name, concept)
				}
			} else {
				switch fs.Arity() {
				case 1:
					X := webuiMustVar("X", fs.Domain()[0])
					Y := webuiMustVar("Y", fs.Range())
					app, err := goivy.NewApply(c, X)
					if err != nil {
						return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
					}
					eq, err := goivy.NewEq(app, Y)
					if err != nil {
						return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
					}
					concept, err := NewCDConcept(c.Name, []*goivy.LogicVariable{X, Y}, eq)
					if err != nil {
						return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
					}
					concepts.SetConcept(c.Name, concept)
					concepts.AppendToList("edges", c.Name)
				case 2:
					X := webuiMustVar("X", fs.Domain()[0])
					Y := webuiMustVar("Y", fs.Domain()[1])
					Z := webuiMustVar("Z", fs.Range())
					app, err := goivy.NewApply(c, X, Y)
					if err != nil {
						return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
					}
					eq, err := goivy.NewEq(app, Z)
					if err != nil {
						return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
					}
					concept, err := NewCDConcept(c.Name, []*goivy.LogicVariable{X, Y, Z}, eq)
					if err != nil {
						return nil, fmt.Errorf("initial concept domain: symbol %q: %w", c.Name, err)
					}
					concepts.SetConcept(c.Name, concept)
				}
			}
		}
	}

	return NewCDConceptDomain(concepts, GetStandardCombiners(), GetStandardCombinations()), nil
}

// GetInitialConceptDomain creates a concept domain from a signature.
func GetInitialConceptDomain(sorts map[string]goivy.Sort, symbols map[string]*goivy.Const) *CDConceptDomain {
	domain, err := GetInitialConceptDomainE(sorts, symbols)
	if err != nil {
		panic(err)
	}
	return domain
}

// GetDiagramConceptDomainE creates a concept domain from a signature and diagram.
func GetDiagramConceptDomainE(sorts map[string]goivy.Sort, symbols []*goivy.Const, diagram goivy.Expr) (*CDConceptDomain, error) {
	concepts := NewCDConceptDict()

	concepts.SetList("nodes", nil)
	concepts.SetList("node_labels", nil)
	concepts.SetList("edges", nil)

	// Add equality concept.
	T := goivy.NewTopSort()
	XT := webuiMustVar("X", T)
	YT := webuiMustVar("Y", T)
	eqXY, err := goivy.NewEq(XT, YT)
	if err != nil {
		return nil, fmt.Errorf("diagram concept domain: equality: %w", err)
	}
	eqConcept, err := NewCDConcept("=", []*goivy.LogicVariable{XT, YT}, eqXY)
	if err != nil {
		return nil, fmt.Errorf("diagram concept domain: equality: %w", err)
	}
	concepts.SetConcept("=", eqConcept)

	// Merge signature symbols with diagram constants.
	// Uses NodeKey for structural identity, matching Python's frozenset union
	// of Const objects with structural equality.
	allConsts := make(map[goivy.NodeKey]*goivy.Const)
	for _, c := range symbols {
		allConsts[goivy.Key(c)] = c
	}
	if diagram != nil {
		for _, sym := range goivy.UsedSymbolsAst(diagram).All() {
			if c, ok := sym.(*goivy.Const); ok {
				k := goivy.Key(c)
				if _, exists := allConsts[k]; !exists {
					allConsts[k] = c
				}
			}
		}
	}

	// Sort by name for deterministic output.
	constNames := make([]string, 0, len(allConsts))
	constByName := make(map[string]*goivy.Const, len(allConsts))
	for _, c := range allConsts {
		if _, exists := constByName[c.Name]; !exists {
			constNames = append(constNames, c.Name)
		}
		constByName[c.Name] = c
	}
	sort.Strings(constNames)

	for _, sname := range constNames {
		c := constByName[sname]
		if goivy.FirstOrderSort(c.CSort) {
			X := webuiMustVar("X", c.CSort)
			eq, err := goivy.NewEq(X, c)
			if err != nil {
				return nil, fmt.Errorf("diagram concept domain: symbol %q: %w", c.Name, err)
			}
			name := fmt.Sprintf("%s:%s", c.Name, c.CSort)
			concept, err := NewCDConcept(name, []*goivy.LogicVariable{X}, eq)
			if err != nil {
				return nil, fmt.Errorf("diagram concept domain: symbol %q: %w", c.Name, err)
			}
			concepts.SetConcept(name, concept)
			concepts.AppendToList("nodes", name)
		} else if fs, ok := c.CSort.(*goivy.LogicFunctionSort); ok {
			switch fs.Arity() {
			case 1:
				X := webuiMustVar("X", fs.Domain()[0])
				app, err := goivy.NewApply(c, X)
				if err != nil {
					return nil, fmt.Errorf("diagram concept domain: symbol %q: %w", c.Name, err)
				}
				concept, err := NewCDConcept(c.Name, []*goivy.LogicVariable{X}, app)
				if err != nil {
					return nil, fmt.Errorf("diagram concept domain: symbol %q: %w", c.Name, err)
				}
				concepts.SetConcept(c.Name, concept)
			case 2:
				X := webuiMustVar("X", fs.Domain()[0])
				Y := webuiMustVar("Y", fs.Domain()[1])
				app, err := goivy.NewApply(c, X, Y)
				if err != nil {
					return nil, fmt.Errorf("diagram concept domain: symbol %q: %w", c.Name, err)
				}
				concept, err := NewCDConcept(c.Name, []*goivy.LogicVariable{X, Y}, app)
				if err != nil {
					return nil, fmt.Errorf("diagram concept domain: symbol %q: %w", c.Name, err)
				}
				concepts.SetConcept(c.Name, concept)
			case 3:
				X := webuiMustVar("X", fs.Domain()[0])
				Y := webuiMustVar("Y", fs.Domain()[1])
				Z := webuiMustVar("Z", fs.Domain()[2])
				app, err := goivy.NewApply(c, X, Y, Z)
				if err != nil {
					return nil, fmt.Errorf("diagram concept domain: symbol %q: %w", c.Name, err)
				}
				concept, err := NewCDConcept(c.Name, []*goivy.LogicVariable{X, Y, Z}, app)
				if err != nil {
					return nil, fmt.Errorf("diagram concept domain: symbol %q: %w", c.Name, err)
				}
				concepts.SetConcept(c.Name, concept)
			}
		}
	}

	return NewCDConceptDomain(concepts, GetStandardCombiners(), GetStandardCombinations()), nil
}

// GetDiagramConceptDomain creates a concept domain from a signature and diagram.
func GetDiagramConceptDomain(sorts map[string]goivy.Sort, symbols []*goivy.Const, diagram goivy.Expr) *CDConceptDomain {
	domain, err := GetDiagramConceptDomainE(sorts, symbols, diagram)
	if err != nil {
		panic(err)
	}
	return domain
}

// UniverseElementToConceptName converts a universe element constant to its
// concept name.
func UniverseElementToConceptName(uc *goivy.Const) string {
	name := uc.Name
	sortStr := uc.CSort.String()
	if !strings.Contains(name, sortStr) {
		name += ":" + sortStr
	}
	return name
}

// GetStructureConceptDomainE creates a concept domain from a state with a universe.
// state is a formula, universe maps sort names to element constants,
// sig provides additional symbol information.
func GetStructureConceptDomainE(
	stateFormula goivy.Expr,
	universe map[string][]*goivy.Const,
	sigSymbols map[string]*goivy.Const,
) (*CDConceptDomain, error) {
	concepts := NewCDConceptDict()

	concepts.SetList("nodes", nil)
	concepts.SetList("node_labels", nil)
	concepts.SetList("edges", nil)

	// Add equality concept.
	T := goivy.NewTopSort()
	XT := webuiMustVar("X", T)
	YT := webuiMustVar("Y", T)
	eqXY, err := goivy.NewEq(XT, YT)
	if err != nil {
		return nil, fmt.Errorf("structure concept domain: equality: %w", err)
	}
	eqConcept, err := NewCDConcept("=", []*goivy.LogicVariable{XT, YT}, eqXY)
	if err != nil {
		return nil, fmt.Errorf("structure concept domain: equality: %w", err)
	}
	concepts.SetConcept("=", eqConcept)

	// Add nodes for universe elements.
	var elements []*goivy.Const
	elementSet := make(map[string]bool)
	for _, ucs := range universe {
		for _, uc := range ucs {
			elements = append(elements, uc)
			elementSet[uc.Name] = true
		}
	}
	sort.Slice(elements, func(i, j int) bool { return elements[i].Name < elements[j].Name })

	for _, uc := range elements {
		X := webuiMustVar("X", uc.CSort)
		name := UniverseElementToConceptName(uc)
		eq, err := goivy.NewEq(X, uc)
		if err != nil {
			return nil, fmt.Errorf("structure concept domain: universe %q: %w", uc.Name, err)
		}
		concept, err := NewCDConcept(name, []*goivy.LogicVariable{X}, eq)
		if err != nil {
			return nil, fmt.Errorf("structure concept domain: universe %q: %w", uc.Name, err)
		}
		concepts.SetConcept(name, concept)
		concepts.AppendToList("nodes", name)

		labelName := "=" + name
		labelConcept, err := NewCDConcept(labelName, []*goivy.LogicVariable{X}, eq)
		if err != nil {
			return nil, fmt.Errorf("structure concept domain: universe %q: %w", uc.Name, err)
		}
		concepts.SetConcept(labelName, labelConcept)
		concepts.AppendToList("node_labels", labelName)
	}

	// Collect all symbols from state formula and signature.
	// Uses NodeKey for structural identity, matching Python's frozenset
	// union/difference of Const objects with structural equality.
	allSymbolsByKey := make(map[goivy.NodeKey]*goivy.Const)
	if stateFormula != nil {
		for _, sym := range goivy.UsedSymbolsAst(stateFormula).All() {
			if c, ok := sym.(*goivy.Const); ok && !elementSet[c.Name] {
				allSymbolsByKey[goivy.Key(c)] = c
			}
		}
	}
	for _, c := range sigSymbols {
		if !elementSet[c.Name] {
			allSymbolsByKey[goivy.Key(c)] = c
		}
	}

	// Build name-sorted list for deterministic iteration.
	allSymbols := make(map[string]*goivy.Const, len(allSymbolsByKey))
	for _, c := range allSymbolsByKey {
		allSymbols[c.Name] = c
	}
	symNames := make([]string, 0, len(allSymbols))
	for name := range allSymbols {
		symNames = append(symNames, name)
	}
	sort.Strings(symNames)

	for _, sname := range symNames {
		c := allSymbols[sname]

		if goivy.FirstOrderSort(c.CSort) {
			X := webuiMustVar("X", c.CSort)
			eq, err := goivy.NewEq(X, c)
			if err != nil {
				return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
			}
			name := "=" + c.Name
			concept, err := NewCDConcept(name, []*goivy.LogicVariable{X}, eq)
			if err != nil {
				return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
			}
			concepts.SetConcept(name, concept)
			concepts.AppendToList("node_labels", name)
		} else if fs, ok := c.CSort.(*goivy.LogicFunctionSort); ok {
			if goivy.SortEqual(fs.Range(), goivy.Boolean) {
				// Relation
				switch fs.Arity() {
				case 1:
					X := webuiMustVar("X", fs.Domain()[0])
					app, err := goivy.NewApply(c, X)
					if err != nil {
						return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
					}
					concept, err := NewCDConcept(c.Name, []*goivy.LogicVariable{X}, app)
					if err != nil {
						return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
					}
					concepts.SetConcept(c.Name, concept)
					concepts.AppendToList("node_labels", c.Name)
				case 2:
					X := webuiMustVar("X", fs.Domain()[0])
					Y := webuiMustVar("Y", fs.Domain()[1])
					app, err := goivy.NewApply(c, X, Y)
					if err != nil {
						return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
					}
					concept, err := NewCDConcept(c.Name, []*goivy.LogicVariable{X, Y}, app)
					if err != nil {
						return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
					}
					concepts.SetConcept(c.Name, concept)
					concepts.AppendToList("edges", c.Name)
				case 3:
					X := webuiMustVar("X", fs.Domain()[0])
					Y := webuiMustVar("Y", fs.Domain()[1])
					Z := webuiMustVar("Z", fs.Domain()[2])
					app, err := goivy.NewApply(c, X, Y, Z)
					if err != nil {
						return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
					}
					concept, err := NewCDConcept(c.Name, []*goivy.LogicVariable{X, Y, Z}, app)
					if err != nil {
						return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
					}
					concepts.SetConcept(c.Name, concept)
				}
			} else {
				// Function
				switch fs.Arity() {
				case 1:
					X := webuiMustVar("X", fs.Domain()[0])
					Y := webuiMustVar("Y", fs.Range())
					app, err := goivy.NewApply(c, X)
					if err != nil {
						return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
					}
					eq, err := goivy.NewEq(app, Y)
					if err != nil {
						return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
					}
					concept, err := NewCDConcept(c.Name, []*goivy.LogicVariable{X, Y}, eq)
					if err != nil {
						return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
					}
					concepts.SetConcept(c.Name, concept)
					concepts.AppendToList("edges", c.Name)
				case 2:
					X := webuiMustVar("X", fs.Domain()[0])
					Y := webuiMustVar("Y", fs.Domain()[1])
					Z := webuiMustVar("Z", fs.Range())
					app, err := goivy.NewApply(c, X, Y)
					if err != nil {
						return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
					}
					eq, err := goivy.NewEq(app, Z)
					if err != nil {
						return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
					}
					concept, err := NewCDConcept(c.Name, []*goivy.LogicVariable{X, Y, Z}, eq)
					if err != nil {
						return nil, fmt.Errorf("structure concept domain: symbol %q: %w", c.Name, err)
					}
					concepts.SetConcept(c.Name, concept)
				}
			}
		}
	}

	return NewCDConceptDomain(concepts, GetStandardCombiners(), GetStandardCombinations()), nil
}

// GetStructureConceptDomain creates a concept domain from a state with a universe.
func GetStructureConceptDomain(
	stateFormula goivy.Expr,
	universe map[string][]*goivy.Const,
	sigSymbols map[string]*goivy.Const,
) *CDConceptDomain {
	domain, err := GetStructureConceptDomainE(stateFormula, universe, sigSymbols)
	if err != nil {
		panic(err)
	}
	return domain
}

// GetStructureConceptAbstractValue computes abstract values for a structure
// (a state with a universe), using direct analysis of the state formula
// rather than Z3.
func GetStructureConceptAbstractValue(
	stateFormula goivy.Expr,
	universe map[string][]*goivy.Const,
) map[string]bool {
	result := make(map[string]bool)

	// Map element constants to their concept names.
	// Uses NodeKey for structural identity, matching Python's dict
	// with Const keys using structural equality.
	nodes := make(map[goivy.NodeKey]string) // Key(const) -> concept name
	var elements []*goivy.Const
	for _, ucs := range universe {
		elements = append(elements, ucs...)
	}
	sort.Slice(elements, func(i, j int) bool { return elements[i].Name < elements[j].Name })

	for _, uc := range elements {
		name := UniverseElementToConceptName(uc)
		nodes[goivy.Key(uc)] = name
		result[TagString(Tag{"node_info", "none", name})] = false
		result[TagString(Tag{"node_info", "at_least_one", name})] = true
		result[TagString(Tag{"node_info", "at_most_one", name})] = true
		result[TagString(Tag{"node_label", "node_necessarily", name, "=" + name})] = true
		result[TagString(Tag{"node_label", "node_necessarily_not", name, "=" + name})] = false
	}

	// Analyze state formula literals.
	andNode, ok := stateFormula.(*goivy.LogicAnd)
	if !ok {
		return result
	}

	// First learn aliases such as @X = 1:client. Relation facts in CTIs are
	// often written with these witness constants instead of raw universe
	// elements.
	for _, lit := range andNode.Terms {
		if _, ok := lit.(*goivy.ForAll); ok {
			continue // universe constraint
		}
		polarity, innerLit := structureLiteralPolarity(lit)
		if l, ok := innerLit.(*goivy.Eq); ok {
			markConstEqualityLabel(result, nodes, l.T1, l.T2, polarity)
			markConstEqualityLabel(result, nodes, l.T2, l.T1, polarity)
		}
	}

	for _, lit := range andNode.Terms {
		if _, ok := lit.(*goivy.ForAll); ok {
			continue // universe constraint
		}
		polarity, innerLit := structureLiteralPolarity(lit)

		switch l := innerLit.(type) {
		case *goivy.Apply:
			markStructureApplyFact(result, nodes, l, polarity)
		case *goivy.Eq:
			markBooleanApplyEquality(result, nodes, l.T1, l.T2, polarity)
			markBooleanApplyEquality(result, nodes, l.T2, l.T1, polarity)
		}
	}

	return result
}

func structureLiteralPolarity(lit goivy.Expr) (bool, goivy.Expr) {
	if notNode, ok := lit.(*goivy.LogicNot); ok {
		return false, notNode.Body
	}
	return true, lit
}

func markConstEqualityLabel(result map[string]bool, nodes map[goivy.NodeKey]string, nodeTerm, labelTerm goivy.Expr, polarity bool) {
	nodeConst, ok := nodeTerm.(*goivy.Const)
	if !ok {
		return
	}
	nodeName, ok := nodes[goivy.Key(nodeConst)]
	if !ok {
		return
	}
	labelConst, ok := labelTerm.(*goivy.Const)
	if !ok {
		return
	}
	if _, isUniverseElement := nodes[goivy.Key(labelConst)]; isUniverseElement {
		return
	}
	labelName := "=" + labelConst.Name
	result[TagString(Tag{"node_label", "node_necessarily", nodeName, labelName})] = polarity
	result[TagString(Tag{"node_label", "node_necessarily_not", nodeName, labelName})] = !polarity
	if polarity {
		nodes[goivy.Key(labelConst)] = nodeName
	}
}

func markBooleanApplyEquality(result map[string]bool, nodes map[goivy.NodeKey]string, left, right goivy.Expr, polarity bool) bool {
	app, ok := left.(*goivy.Apply)
	if !ok {
		return false
	}
	value, ok := structureBooleanValue(right)
	if !ok {
		return false
	}
	if !polarity {
		value = !value
	}
	markStructureApplyFact(result, nodes, app, value)
	return true
}

func structureBooleanValue(expr goivy.Expr) (bool, bool) {
	if goivy.IsTrue(expr) {
		return true, true
	}
	if goivy.IsFalse(expr) {
		return false, true
	}
	return false, false
}

func markStructureApplyFact(result map[string]bool, nodes map[goivy.NodeKey]string, app *goivy.Apply, polarity bool) {
	fs, ok := app.Func.(*goivy.Const)
	if !ok {
		return
	}
	fSort, ok := fs.CSort.(*goivy.LogicFunctionSort)
	if !ok {
		return
	}
	if fSort.Arity() == 1 {
		if t0, ok := app.Terms[0].(*goivy.Const); ok {
			if nodeName, ok := nodes[goivy.Key(t0)]; ok {
				labelName := fs.Name
				result[TagString(Tag{"node_label", "node_necessarily", nodeName, labelName})] = polarity
				result[TagString(Tag{"node_label", "node_necessarily_not", nodeName, labelName})] = !polarity
			}
		}
	} else if fSort.Arity() == 2 {
		t0, ok0 := app.Terms[0].(*goivy.Const)
		t1, ok1 := app.Terms[1].(*goivy.Const)
		if ok0 && ok1 {
			sn, sok := nodes[goivy.Key(t0)]
			tn, tok := nodes[goivy.Key(t1)]
			if sok && tok {
				edgeName := fs.Name
				result[TagString(Tag{"edge_info", "all_to_all", edgeName, sn, tn})] = polarity
				result[TagString(Tag{"edge_info", "none_to_none", edgeName, sn, tn})] = !polarity
			}
		}
	}
}

// GetStructureRenaming generates prettier names for universe constants
// based on topological sort of order relations.
func GetStructureRenaming(
	stateFormula goivy.Expr,
	universe map[string][]*goivy.Const,
	orderRelations map[string]bool,
) map[string]string {
	var elements []*goivy.Const
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
	var order [][2]*goivy.Const
	if andNode, ok := stateFormula.(*goivy.LogicAnd); ok {
		for _, lit := range andNode.Terms {
			if app, ok := lit.(*goivy.Apply); ok {
				fc, ok2 := app.Func.(*goivy.Const)
				if !ok2 {
					continue
				}
				fs, ok3 := fc.CSort.(*goivy.LogicFunctionSort)
				if !ok3 || fs.Arity() != 2 {
					continue
				}
				if !goivy.SortEqual(fs.Domain()[0], fs.Domain()[1]) {
					continue
				}
				if !goivy.SortEqual(fs.Range(), goivy.Boolean) {
					continue
				}
				if !orderRelations[fc.Name] {
					continue
				}
				if len(app.Terms) == 2 {
					t0, ok0 := app.Terms[0].(*goivy.Const)
					t1, ok1 := app.Terms[1].(*goivy.Const)
					if ok0 && ok1 {
						order = append(order, [2]*goivy.Const{t0, t1})
					}
				}
			}
		}
	}

	// Topological sort using the order relations.
	var orderPairs [][2]*goivy.Const
	nameSet := make(map[string]*goivy.Const)
	for _, elem := range elements {
		nameSet[elem.Name] = elem
	}
	for _, pair := range order {
		e0, ok0 := nameSet[pair[0].Name]
		e1, ok1 := nameSet[pair[1].Name]
		if ok0 && ok1 {
			orderPairs = append(orderPairs, [2]*goivy.Const{e0, e1})
		}
	}
	sorted := goivy.TopologicalSort(elements, orderPairs, func(elem *goivy.Const) string { return elem.Name })

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

func webuiMustVar(name string, s goivy.Sort) *goivy.LogicVariable {
	v, err := goivy.NewVariable(name, s)
	if err != nil {
		panic(err)
	}
	return v
}

func webuiMustFuncSort(sorts ...goivy.Sort) *goivy.LogicFunctionSort {
	fs, err := goivy.NewFunctionSort(sorts...)
	if err != nil {
		panic(err)
	}
	return fs
}

func mustApplyVar(v *goivy.LogicVariable, args ...goivy.Expr) goivy.Expr {
	n, err := v.Call(args...)
	if err != nil {
		panic(err)
	}
	return n
}

func mustNot(body goivy.Expr) goivy.Expr {
	n, err := goivy.NewNot(body)
	if err != nil {
		panic(err)
	}
	return n
}

func mustAnd(terms ...goivy.Expr) goivy.Expr {
	n, err := goivy.NewAnd(terms...)
	if err != nil {
		panic(err)
	}
	return n
}

func mustOr(terms ...goivy.Expr) goivy.Expr {
	n, err := goivy.NewOr(terms...)
	if err != nil {
		panic(err)
	}
	return n
}

func webuiMustEq(t1, t2 goivy.Expr) goivy.Expr {
	n, err := goivy.NewEq(t1, t2)
	if err != nil {
		panic(err)
	}
	return n
}

func mustImplies(t1, t2 goivy.Expr) goivy.Expr {
	n, err := goivy.NewImplies(t1, t2)
	if err != nil {
		panic(err)
	}
	return n
}

func webuiMustForAll(vars []*goivy.LogicVariable, body goivy.Expr) goivy.Expr {
	n, err := goivy.NewForAll(vars, body)
	if err != nil {
		panic(err)
	}
	return n
}

func webuiMustExists(vars []*goivy.LogicVariable, body goivy.Expr) goivy.Expr {
	n, err := goivy.NewExists(vars, body)
	if err != nil {
		panic(err)
	}
	return n
}
