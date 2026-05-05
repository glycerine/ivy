// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from concept_interactive_session.py.
//
// ConceptInteractiveSession provides the interactive concept graph session
// with push/pop undo, concept splitting, suppose constraints, materialization,
// and fact gathering. It uses the CDConceptDomain types from concept_domain.go.
package webui

import (
	"fmt"
	"strings"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	"github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/z3bridge"
)

// ConceptInteractiveSession is the full interactive concept-graph session
// with domain, state, axioms, constraints, undo stack, and alpha cache.
type ConceptInteractiveSession struct {
	Domain             *CDConceptDomain
	State              logic.Expr
	Axioms             logic.Expr
	GoalConstraints    []logic.Expr
	SupposeConstraints []logic.Expr
	UndoStack          []*cisUndoEntry
	RedoStack          []*cisUndoEntry
	Cache              map[string]bool // tag-string -> bool
	AbstractValue      []TagValue
	Widget             CISWidget // nil if no widget
	AnalysisSession    map[string]*CDConceptDomain
	Info               string
}

// cisUndoEntry stores state for one undo level.
type cisUndoEntry struct {
	Domain             *CDConceptDomain
	SupposeConstraints []logic.Expr
}

// CISWidget is an interface for rendering widgets that display the concept graph.
type CISWidget interface {
	Render()
	SetConceptSession(s *ConceptInteractiveSession)
	Projection() func(string, string) bool
}

// NewConceptInteractiveSession creates a new interactive session.
// If recompute is true, alpha abstraction is run immediately.
func NewConceptInteractiveSession(
	domain *CDConceptDomain,
	state logic.Expr,
	axioms logic.Expr,
	goalConstraints []logic.Expr,
	supposeConstraints []logic.Expr,
	widget CISWidget,
	analysisSession map[string]*CDConceptDomain,
	cache map[string]bool,
	recompute bool,
) *ConceptInteractiveSession {
	if goalConstraints == nil {
		goalConstraints = []logic.Expr{}
	}
	if supposeConstraints == nil {
		supposeConstraints = []logic.Expr{}
	}
	if analysisSession == nil {
		analysisSession = make(map[string]*CDConceptDomain)
	}

	s := &ConceptInteractiveSession{
		Domain:             domain,
		State:              state,
		Axioms:             axioms,
		GoalConstraints:    append([]logic.Expr{}, goalConstraints...),
		SupposeConstraints: append([]logic.Expr{}, supposeConstraints...),
		Widget:             widget,
		AnalysisSession:    analysisSession,
		Cache:              cache,
	}
	if widget != nil {
		widget.SetConceptSession(s)
	}
	if recompute {
		s.Recompute(nil)
	}
	return s
}

// Clone creates a copy of the session. If recompute is true, recomputes
// alpha abstraction on the copy.
func (s *ConceptInteractiveSession) Clone(recompute bool) *ConceptInteractiveSession {
	var cacheCopy map[string]bool
	if s.Cache != nil {
		cacheCopy = make(map[string]bool, len(s.Cache))
		for k, v := range s.Cache {
			cacheCopy[k] = v
		}
	}
	result := NewConceptInteractiveSession(
		s.Domain.Copy(),
		s.State,
		s.Axioms,
		append([]logic.Expr{}, s.GoalConstraints...),
		append([]logic.Expr{}, s.SupposeConstraints...),
		s.Widget,
		s.AnalysisSession,
		cacheCopy,
		recompute,
	)
	for _, u := range s.UndoStack {
		result.UndoStack = append(result.UndoStack, &cisUndoEntry{
			Domain:             u.Domain.Copy(),
			SupposeConstraints: append([]logic.Expr{}, u.SupposeConstraints...),
		})
	}
	for _, u := range s.RedoStack {
		result.RedoStack = append(result.RedoStack, &cisUndoEntry{
			Domain:             u.Domain.Copy(),
			SupposeConstraints: append([]logic.Expr{}, u.SupposeConstraints...),
		})
	}
	return result
}

// ToFormula combines state, axioms, goal constraints, and suppose constraints
// into a single formula.
func (s *ConceptInteractiveSession) ToFormula() logic.Expr {
	terms := make([]logic.Expr, 0, 2+len(s.GoalConstraints)+len(s.SupposeConstraints))
	if s.State != nil {
		terms = append(terms, s.State)
	}
	if s.Axioms != nil {
		terms = append(terms, s.Axioms)
	}
	terms = append(terms, s.GoalConstraints...)
	terms = append(terms, s.SupposeConstraints...)
	if len(terms) == 0 {
		return logic.True
	}
	result, _ := logic.NewAnd(terms...)
	return result
}

// FreshConstName returns a fresh constant name not used in any formula
// of the session.
func (s *ConceptInteractiveSession) FreshConstName(extra map[string]bool) string {
	used := make(map[string]bool)
	// Collect from formula
	formula := s.ToFormula()
	if formula != nil {
		for _, c := range il.UsedSymbolsAst(formula).All() {
			used[logic.ExprName(c)] = true
		}
	}
	// Collect from concept formulas
	s.Domain.Concepts.ForEachConcept(func(_ string, c *CDConcept) {
		for _, uc := range il.UsedSymbolsAst(c.Formula).All() {
			used[logic.ExprName(uc)] = true
		}
	})
	// Collect from extra
	for k := range extra {
		used[k] = true
	}
	// Generate fresh name
	for i := 0; ; i++ {
		name := fmt.Sprintf("__c%d", i)
		if !used[name] {
			return name
		}
	}
}

// GetProjection returns the projection function from the widget, or nil.
func (s *ConceptInteractiveSession) GetProjection() func(string, string) bool {
	if s.Widget != nil {
		return s.Widget.Projection()
	}
	return nil
}

// Recompute runs alpha abstraction on the domain with the current state.
// If projection is nil, a default all-true projection is used.
func (s *ConceptInteractiveSession) Recompute(projection func(string, string) bool) {
	if projection == nil {
		projection = func(string, string) bool { return true }
	}
	s.AbstractValue = WebUIAlpha(s.Domain, s.ToFormula(), s.Cache, projection)
	if s.Widget != nil {
		s.Widget.Render()
	}
}

// Push saves the current domain and suppose constraints for later undo.
// Clears the redo stack (new actions invalidate the redo history).
func (s *ConceptInteractiveSession) Push() {
	s.UndoStack = append(s.UndoStack, &cisUndoEntry{
		Domain:             s.Domain.Copy(),
		SupposeConstraints: append([]logic.Expr{}, s.SupposeConstraints...),
	})
	s.RedoStack = nil // new action clears redo
}

// Pop restores the domain and suppose constraints from the undo stack,
// saving the current state to the redo stack.
func (s *ConceptInteractiveSession) Pop() error {
	if len(s.UndoStack) == 0 {
		return fmt.Errorf("nothing to undo")
	}
	// Save current state to redo stack
	s.RedoStack = append(s.RedoStack, &cisUndoEntry{
		Domain:             s.Domain.Copy(),
		SupposeConstraints: append([]logic.Expr{}, s.SupposeConstraints...),
	})
	entry := s.UndoStack[len(s.UndoStack)-1]
	s.UndoStack = s.UndoStack[:len(s.UndoStack)-1]
	s.Domain = entry.Domain.Copy()
	s.SupposeConstraints = append([]logic.Expr{}, entry.SupposeConstraints...)
	return nil
}

// Undo pops the undo stack and recomputes.
func (s *ConceptInteractiveSession) Undo() error {
	if err := s.Pop(); err != nil {
		return err
	}
	s.Recompute(nil)
	return nil
}

// Redo restores the most recent undone state from the redo stack.
func (s *ConceptInteractiveSession) Redo() error {
	if len(s.RedoStack) == 0 {
		return fmt.Errorf("nothing to redo")
	}
	// Save current state to undo stack (without clearing redo)
	s.UndoStack = append(s.UndoStack, &cisUndoEntry{
		Domain:             s.Domain.Copy(),
		SupposeConstraints: append([]logic.Expr{}, s.SupposeConstraints...),
	})
	entry := s.RedoStack[len(s.RedoStack)-1]
	s.RedoStack = s.RedoStack[:len(s.RedoStack)-1]
	s.Domain = entry.Domain.Copy()
	s.SupposeConstraints = append([]logic.Expr{}, entry.SupposeConstraints...)
	s.Recompute(nil)
	return nil
}

// Split splits a concept by another concept, creating +/- variants.
func (s *ConceptInteractiveSession) Split(concept, splitBy string) {
	s.Push()
	s.Domain.Split(concept, splitBy)
	s.Recompute(nil)
}

// RemoveConcepts removes concepts from the domain.
func (s *ConceptInteractiveSession) RemoveConcepts(concepts ...string) {
	s.Push()
	for _, concept := range concepts {
		s.Domain.Concepts.Delete(concept)
		s.Domain.ReplaceConcept(concept, nil)
	}
	s.Recompute(nil)
}

// supposeEmpty adds a constraint that the concept is empty (internal, no push).
func (s *ConceptInteractiveSession) supposeEmpty(concept string) {
	c := s.Domain.Concepts.GetConcept(concept)
	if c == nil {
		return
	}
	fv := logicutil.FreeVariablesList(c.Formula)
	notF, _ := logic.NewNot(c.Formula)
	if len(fv) > 0 {
		forall, _ := logic.NewForAll(fv, notF)
		s.SupposeConstraints = append(s.SupposeConstraints, forall)
	} else {
		s.SupposeConstraints = append(s.SupposeConstraints, notF)
	}
}

// SupposeEmpty marks a concept as empty with undo support.
func (s *ConceptInteractiveSession) SupposeEmpty(concept string) {
	s.Push()
	s.supposeEmpty(concept)
	s.Recompute(nil)
}

// GetWitnesses returns constants that are witnesses for a unary concept.
// A witness c satisfies: concept(x) implies x=c.
func (s *ConceptInteractiveSession) GetWitnesses(conceptName string) []*logic.Const {
	concept := s.Domain.Concepts.GetConcept(conceptName)
	if concept == nil || concept.Arity() != 1 {
		return nil
	}
	cSort := concept.Variables[0].VSort
	if _, ok := cSort.(*logic.TopSort); ok {
		return nil
	}

	// Special case for unit sort.
	if cSort.String() == "unit" {
		return []*logic.Const{logic.NewConst("0", cSort)}
	}

	var constants []*logic.Const
	for _, sym := range il.UsedSymbolsAst(concept.Formula).All() {
		if c, ok := sym.(*logic.Const); ok {
			constants = append(constants, c)
		}
	}
	freshName := s.FreshConstName(nil)
	x := logic.NewConst(freshName, cSort)
	f, err := concept.Call(x)
	if err != nil {
		return nil
	}

	var witnesses []*logic.Const
	for _, c := range constants {
		if logic.SortEqual(c.CSort, cSort) || isTopSort(c.CSort) || isTopSort(cSort) {
			// Check if f implies x=c using Z3.
			eq, eqErr := logic.NewEq(x, c)
			if eqErr != nil {
				continue
			}
			// Use Z3 implies check.
			implies, implErr := z3Implies(f, eq)
			if implErr != nil || !implies {
				continue
			}
			witnesses = append(witnesses, c)
		}
	}
	return witnesses
}

// z3Implies checks if fmla1 implies fmla2 using the solver.
// Matches Python's concept_interactive_session.py z3_implies:
//
//	slvr = new_solver()
//	slvr.add(fmla1)
//	slvr.add(Not(fmla2))
//	return not is_sat(slvr)
func z3Implies(fmla1, fmla2 logic.Expr) (bool, error) {
	slv := z3bridge.NewSolver(nil, nil)
	return slv.Implies(fmla1, fmla2)
}

func isTopSort(s logic.Sort) bool {
	_, ok := s.(*logic.TopSort)
	return ok
}

// Suppose adds a formula to the suppose constraints (no push).
func (s *ConceptInteractiveSession) Suppose(fmla logic.Expr) {
	if logicutil.IsTautologyEquality(fmla) {
		return
	}
	s.SupposeConstraints = append(s.SupposeConstraints, fmla)
}

// materializeNode creates a concrete witness for a concept (internal).
// Returns the witness constant.
func (s *ConceptInteractiveSession) materializeNode(conceptName string) *logic.Const {
	concept := s.Domain.Concepts.GetConcept(conceptName)
	if concept == nil || concept.Arity() != 1 {
		return nil
	}
	cSort := concept.Variables[0].VSort
	if _, ok := cSort.(*logic.TopSort); ok {
		return nil
	}

	witnesses := s.GetWitnesses(conceptName)
	if len(witnesses) > 0 {
		c := witnesses[0]
		f, err := concept.Call(c)
		if err == nil {
			s.Suppose(f)
		}
		return c
	}

	// No witness found, create a fresh constant.
	freshName := s.FreshConstName(nil)
	c := logic.NewConst(freshName, cSort)

	// Add equality concept and split.
	X := mustVar("X", c.CSort)
	eq, _ := logic.NewEq(X, c)
	eqName := "=" + c.Name
	s.Domain.Concepts.SetConcept(eqName, MustCDConcept(eqName, []*logic.Variable{X}, eq))
	s.Domain.Split(conceptName, eqName)

	f, err := concept.Call(c)
	if err == nil {
		s.Suppose(f)
	}
	return c
}

// MaterializeNode creates a concrete witness for a concept with undo support.
func (s *ConceptInteractiveSession) MaterializeNode(conceptName string) {
	s.Push()
	s.materializeNode(conceptName)
	s.Recompute(nil)
}

// materializeEdge creates concrete witnesses for source and target nodes
// and supposes the edge (internal).
func (s *ConceptInteractiveSession) materializeEdge(edge, source, target string, polarity bool) (*logic.Const, *logic.Const) {
	edgeConcept := s.Domain.Concepts.GetConcept(edge)
	if edgeConcept == nil {
		return nil, nil
	}
	sourceC := s.materializeNode(source)
	if sourceC == nil {
		return nil, nil
	}
	var targetC *logic.Const
	if source == target {
		targetC = sourceC
	} else {
		targetC = s.materializeNode(target)
		if targetC == nil {
			return sourceC, nil
		}
	}
	f, err := edgeConcept.Call(sourceC, targetC)
	if err != nil {
		return sourceC, targetC
	}
	if polarity {
		s.Suppose(f)
	} else {
		notF, _ := logic.NewNot(f)
		s.Suppose(notF)
	}
	return sourceC, targetC
}

// MaterializeEdge materializes an edge with undo support.
func (s *ConceptInteractiveSession) MaterializeEdge(edge, source, target string, polarity bool) {
	s.Push()
	s.materializeEdge(edge, source, target, polarity)
	s.Recompute(nil)
}

// normalizeFacts normalizes a list of formulas by removing tautological equalities.
func normalizeFacts(facts []logic.Expr) []logic.Expr {
	if len(facts) == 0 {
		return facts
	}
	var result []logic.Expr
	for _, f := range facts {
		if !logicutil.IsTautologyEquality(f) {
			result = append(result, f)
		}
	}
	return result
}

// GetNodeFacts returns facts for a node concept used by gather.
func (s *ConceptInteractiveSession) GetNodeFacts(node string) []logic.Expr {
	av := s.abstractValueMap()
	var facts []logic.Expr

	if av[TagString(Tag{"node_info", "at_least_one", node})] {
		for _, c := range s.GetWitnesses(node) {
			concept := s.Domain.Concepts.GetConcept(node)
			if concept == nil {
				continue
			}
			f, err := concept.Call(c)
			if err == nil {
				facts = append(facts, f)
			}
			// Check node labels.
			for _, nl := range s.Domain.Concepts.GetList("node_labels") {
				if av[TagString(Tag{"node_label", "node_necessarily", node, nl})] {
					nlConcept := s.Domain.Concepts.GetConcept(nl)
					if nlConcept != nil {
						nf, err := nlConcept.Call(c)
						if err == nil {
							facts = append(facts, nf)
						}
					}
				} else if av[TagString(Tag{"node_label", "node_necessarily_not", node, nl})] {
					nlConcept := s.Domain.Concepts.GetConcept(nl)
					if nlConcept != nil {
						nf, err := nlConcept.Call(c)
						if err == nil {
							notF, _ := logic.NewNot(nf)
							facts = append(facts, notF)
						}
					}
				}
			}
			// Check enum_case labels if they exist.
			for _, nl := range s.Domain.Concepts.GetList("enum_case") {
				if av[TagString(Tag{"node_label", "node_necessarily", node, nl})] {
					nlConcept := s.Domain.Concepts.GetConcept(nl)
					if nlConcept != nil {
						nf, err := nlConcept.Call(c)
						if err == nil {
							facts = append(facts, nf)
						}
					}
				}
			}
		}
	}
	return normalizeFacts(facts)
}

// abstractValueMap converts AbstractValue to a string-keyed map for lookups.
func (s *ConceptInteractiveSession) abstractValueMap() map[string]bool {
	m := make(map[string]bool, len(s.AbstractValue))
	for _, tv := range s.AbstractValue {
		m[TagString(tv.Tag)] = tv.Value
	}
	return m
}

// getEdgeFact returns facts for a specific edge/source/target with given polarity.
func (s *ConceptInteractiveSession) getEdgeFact(edge, source, target string, polarity bool) []logic.Expr {
	var facts []logic.Expr
	edgeConcept := s.Domain.Concepts.GetConcept(edge)
	if edgeConcept == nil {
		return nil
	}
	sourceWitnesses := s.GetWitnesses(source)
	targetWitnesses := s.GetWitnesses(target)
	for _, sc := range sourceWitnesses {
		for _, tc := range targetWitnesses {
			f, err := edgeConcept.Call(sc, tc)
			if err != nil {
				continue
			}
			if polarity {
				facts = append(facts, f)
			} else {
				notF, _ := logic.NewNot(f)
				facts = append(facts, notF)
			}
		}
	}
	return normalizeFacts(facts)
}

// GetEdgeFacts returns facts for an edge used by gather.
// If filterPolarity is nil, returns both positive and negative facts.
func (s *ConceptInteractiveSession) GetEdgeFacts(edge, source, target string, filterPolarity *bool) []logic.Expr {
	av := s.abstractValueMap()
	x := strings.Join([]string{edge, source, target}, "|")

	ataKey := "edge_info|all_to_all|" + x
	ntnKey := "edge_info|none_to_none|" + x

	if _, ok := av[ataKey]; !ok {
		return nil // not a well-sorted edge
	}

	var polarity bool
	if av[ntnKey] {
		polarity = false
	} else if av[ataKey] {
		polarity = true
	} else {
		return nil // not a definite edge
	}

	if filterPolarity != nil && *filterPolarity != polarity {
		return nil
	}

	return s.getEdgeFact(edge, source, target, polarity)
}

// GetFacts returns all gathered facts.
func (s *ConceptInteractiveSession) GetFacts(projection func(string, string, string) bool) []logic.Expr {
	var facts []logic.Expr

	for _, node := range s.Domain.Concepts.GetList("nodes") {
		facts = append(facts, s.GetNodeFacts(node)...)
	}

	for _, tv := range s.AbstractValue {
		if !tv.Value {
			continue
		}
		if len(tv.Tag) < 3 {
			continue
		}
		if tv.Tag[0] == "edge_info" && (tv.Tag[1] == "none_to_none" || tv.Tag[1] == "all_to_all") {
			edgeName := tv.Tag[2]
			if len(tv.Tag) < 5 {
				continue
			}
			srcName := tv.Tag[3]
			tgtName := tv.Tag[4]
			if projection != nil && !projection(edgeName, "edges", tv.Tag[1]) {
				continue
			}
			polarity := tv.Tag[1] == "all_to_all"
			facts = append(facts, s.getEdgeFact(edgeName, srcName, tgtName, polarity)...)
		}
	}
	return facts
}

// SaveDomain saves the current domain to the analysis session.
func (s *ConceptInteractiveSession) SaveDomain(name string) {
	s.AnalysisSession[name] = s.Domain.Copy()
}

// LoadDomain loads a domain from the analysis session.
func (s *ConceptInteractiveSession) LoadDomain(name string) error {
	d, ok := s.AnalysisSession[name]
	if !ok {
		return fmt.Errorf("domain %q not found", name)
	}
	s.Push()
	s.Domain = d.Copy()
	s.Recompute(nil)
	return nil
}

// ReplaceDomain replaces the domain and suppose constraints.
func (s *ConceptInteractiveSession) ReplaceDomain(newDomain *CDConceptDomain, newSupposeConstraints []logic.Expr) {
	s.Push()
	s.Domain = newDomain.Copy()
	if newSupposeConstraints != nil {
		s.SupposeConstraints = append([]logic.Expr{}, newSupposeConstraints...)
	} else {
		s.SupposeConstraints = nil
	}
	s.Recompute(nil)
}

// NamedConcept is a (name, concept) pair returned by GetProjections.
type NamedConcept struct {
	Name    string
	Concept *CDConcept
}

// GetProjections returns all possible projections at a node.
func (s *ConceptInteractiveSession) GetProjections(node string) []NamedConcept {
	witnesses := s.GetWitnesses(node)
	if len(witnesses) == 0 {
		return nil
	}
	w := witnesses[0]
	var result []NamedConcept

	for _, tName := range s.Domain.ConceptsByArity(3) {
		tConcept := s.Domain.Concepts.GetConcept(tName)
		if tConcept == nil {
			continue
		}
		for _, v := range tConcept.Variables {
			if logic.SortEqual(v.VSort, w.CSort) {
				// Create a projected binary concept.
				var variables []*logic.Variable
				for _, x := range tConcept.Variables {
					if x != v {
						variables = append(variables, x)
					}
				}
				subs := map[logic.NodeKey]logic.Expr{logic.Key(v): w}
				formula, err := logicutil.Substitute(tConcept.Formula, subs)
				if err != nil {
					continue
				}
				name := formula.String()
				concept := MustCDConcept(name, variables, formula)
				result = append(result, NamedConcept{Name: name, Concept: concept})
			}
		}
	}
	return result
}

// AddEdge adds an edge concept to the domain with undo support.
func (s *ConceptInteractiveSession) AddEdge(name string, concept *CDConcept) {
	s.Push()
	s.Domain.Concepts.SetConcept(name, concept)
	s.Domain.Concepts.AppendToList("edges", name)
	s.Recompute(nil)
}

// AddCustomEdge adds a custom edge combination to the domain.
func (s *ConceptInteractiveSession) AddCustomEdge(edge, source, target string) {
	s.Push()
	s.Domain.Combinations = append(s.Domain.Combinations,
		NewCombination("custom_edge_info", "edge_info", edge, source, target),
	)
	s.Recompute(nil)
}

// AddCustomNodeLabel adds a custom node label combination.
func (s *ConceptInteractiveSession) AddCustomNodeLabel(node, nodeLabel string) {
	s.Push()
	s.Domain.Combinations = append(s.Domain.Combinations,
		NewCombination("custom_node_label", "node_label", node, nodeLabel),
	)
	s.Recompute(nil)
}

// Reset restores the concept domain to its initial state.
func (s *ConceptInteractiveSession) Reset(sorts map[string]logic.Sort, symbols map[string]*logic.Const) {
	s.Push()
	s.Domain = GetInitialConceptDomain(sorts, symbols)
	s.Cache = make(map[string]bool)
	s.Recompute(nil)
}

// Diagram switches to the diagram concept domain.
func (s *ConceptInteractiveSession) Diagram(sorts map[string]logic.Sort, symbols []*logic.Const, state logic.Expr) {
	s.Push()
	s.Domain = GetDiagramConceptDomain(sorts, symbols, state)
	s.Cache = make(map[string]bool)
	s.Recompute(nil)
}

// RelationNames returns display names for the state checkbox panel.
// Returns edges + node_labels, formatted for display.
func (s *ConceptInteractiveSession) RelationNames() []string {
	var names []string
	edges := s.Domain.Concepts.GetList("edges")
	nodeLabels := s.Domain.Concepts.GetList("node_labels")

	for _, id := range edges {
		c := s.Domain.Concepts.GetConcept(id)
		if c != nil && c.Arity() >= 2 {
			var varNames []string
			for _, v := range c.Variables {
				varNames = append(varNames, v.Name)
			}
			names = append(names, id+"("+strings.Join(varNames, ",")+")")
		} else {
			names = append(names, id)
		}
	}
	for _, id := range nodeLabels {
		names = append(names, id)
	}
	return names
}

// EdgeNames returns the raw edge concept IDs.
func (s *ConceptInteractiveSession) EdgeNames() []string {
	return s.Domain.Concepts.GetList("edges")
}

// NodeLabelNames returns the raw node label concept IDs.
func (s *ConceptInteractiveSession) NodeLabelNames() []string {
	return s.Domain.Concepts.GetList("node_labels")
}

// NodeNames returns the raw node (sort) concept IDs.
func (s *ConceptInteractiveSession) NodeNames() []string {
	return s.Domain.Concepts.GetList("nodes")
}
