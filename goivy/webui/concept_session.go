package webui

import (
	"fmt"
	"strings"
)

// ConceptSession is the interactive concept-graph session with undo/redo.
type ConceptSession struct {
	Domain        *ConceptDomain  `json:"domain"`
	AbstractValue map[string]bool `json:"abstract_value"`
	undoStack     []*undoEntry
}

// undoEntry stores the state needed to undo one operation.
type undoEntry struct {
	Domain           *ConceptDomain
	SupposeContraints []string // formula strings
}

// NewConceptSession creates a new empty concept session.
func NewConceptSession() *ConceptSession {
	return &ConceptSession{
		Domain:        NewConceptDomain(),
		AbstractValue: make(map[string]bool),
	}
}

// push saves the current domain state for later undo.
func (cs *ConceptSession) push() {
	cs.undoStack = append(cs.undoStack, &undoEntry{
		Domain: cs.Domain.Copy(),
	})
}

// Split splits a concept by another concept (creating + and - variants).
func (cs *ConceptSession) Split(concept, splitBy string) error {
	if _, ok := cs.Domain.Concepts[concept]; !ok {
		return fmt.Errorf("concept %q not found", concept)
	}
	if _, ok := cs.Domain.Concepts[splitBy]; !ok {
		return fmt.Errorf("split-by concept %q not found", splitBy)
	}
	cs.push()

	orig := cs.Domain.Concepts[concept]
	posName := concept + "+" + splitBy
	negName := concept + "-" + splitBy

	cs.Domain.Concepts[posName] = &Concept{
		Name:      posName,
		Variables: orig.Variables,
		Formula:   fmt.Sprintf("(%s) & (%s)", orig.Formula, cs.Domain.Concepts[splitBy].Formula),
		Sorts:     orig.Sorts,
	}
	cs.Domain.Concepts[negName] = &Concept{
		Name:      negName,
		Variables: orig.Variables,
		Formula:   fmt.Sprintf("(%s) & ~(%s)", orig.Formula, cs.Domain.Concepts[splitBy].Formula),
		Sorts:     orig.Sorts,
	}
	delete(cs.Domain.Concepts, concept)
	cs.Recompute()
	return nil
}

// SupposeEmpty marks a concept as empty (supposes no elements satisfy it).
func (cs *ConceptSession) SupposeEmpty(concept string) error {
	if _, ok := cs.Domain.Concepts[concept]; !ok {
		return fmt.Errorf("concept %q not found", concept)
	}
	cs.push()
	cs.Recompute()
	// Mark in abstract value after recompute (Recompute clears the map,
	// so explicit overrides like this must come after).
	cs.AbstractValue["node_info|none|"+concept] = true
	return nil
}

// RemoveConcept deletes a concept from the domain.
func (cs *ConceptSession) RemoveConcept(concept string) error {
	if _, ok := cs.Domain.Concepts[concept]; !ok {
		return fmt.Errorf("concept %q not found", concept)
	}
	cs.push()
	delete(cs.Domain.Concepts, concept)
	// Also remove combiners referencing this concept.
	var kept []*ConceptCombiner
	for _, c := range cs.Domain.Combiners {
		if c.Source != concept && c.Target != concept {
			kept = append(kept, c)
		}
	}
	cs.Domain.Combiners = kept
	cs.Recompute()
	return nil
}

// Undo reverts the last concept operation.
func (cs *ConceptSession) Undo() error {
	if len(cs.undoStack) == 0 {
		return fmt.Errorf("nothing to undo")
	}
	entry := cs.undoStack[len(cs.undoStack)-1]
	cs.undoStack = cs.undoStack[:len(cs.undoStack)-1]
	cs.Domain = entry.Domain.Copy()
	cs.Recompute()
	return nil
}

// Materialize creates a concrete witness for a concept (Python: _materialize_node).
// Creates an equality witness concept and splits the original, matching the
// structural side of Python's ConceptInteractiveSession._materialize_node.
// For Z3-backed materialization, use ConceptInteractiveSession.MaterializeNode.
func (cs *ConceptSession) Materialize(concept string) error {
	c, ok := cs.Domain.Concepts[concept]
	if !ok {
		return fmt.Errorf("concept %q not found", concept)
	}
	cs.push()
	freshName := cs.freshConstName()
	witnessName := "=" + freshName
	cs.Domain.Concepts[witnessName] = &Concept{
		Name:      witnessName,
		Variables: c.Variables,
		Formula:   fmt.Sprintf("(X = %s)", freshName),
		Sorts:     c.Sorts,
		Arity:     c.Arity,
	}
	cs.Split(concept, witnessName)
	cs.Recompute()
	return nil
}

// freshConstName generates a unique constant name not colliding with
// existing concept names.
func (cs *ConceptSession) freshConstName() string {
	for i := 0; ; i++ {
		name := fmt.Sprintf("__c%d", i)
		if _, ok := cs.Domain.Concepts[name]; !ok {
			if _, ok2 := cs.Domain.Concepts["="+name]; !ok2 {
				return name
			}
		}
	}
}

// Reset restores the concept domain to its initial state, clearing
// all splits, supposes, and materializations.
func (cs *ConceptSession) Reset() {
	cs.push()
	cs.Domain = NewConceptDomain()
	cs.AbstractValue = make(map[string]bool)
	cs.Recompute()
}

// Diagram switches to the diagram concept domain, which shows
// concrete elements as individual nodes rather than abstract classes.
func (cs *ConceptSession) Diagram() {
	cs.push()
	// In the full implementation this calls GetDiagramConceptDomain
	// to rebuild the domain from the current state's universe.
	// For now, keep the current domain but clear the abstract value
	// so the graph re-renders.
	cs.AbstractValue = make(map[string]bool)
	cs.Recompute()
}

// RelationNames returns display names for the state checkbox panel.
// Matches Python's Graph.relation_ids: edges + node_labels (NOT sorts).
// Each name is formatted like the Python UI: "link(X,Y)" for binary, "semaphore" for unary.
func (cs *ConceptSession) RelationNames() []string {
	ids := cs.Domain.RelationIDs()
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		concept, ok := cs.Domain.Concepts[id]
		if !ok {
			continue
		}
		// Format display name matching Python's concept_label()
		display := concept.Name
		if concept.Arity >= 2 && len(concept.Variables) >= 2 {
			display = concept.Name + "(" + strings.Join(concept.Variables, ",") + ")"
		}
		names = append(names, display)
	}
	return names
}

// Splatter splits a concept node into sub-nodes, one per constant of its sort.
// Matches Python's ivy_graph.py Graph.splatter:
//   1. Collects constants of the node's sort
//   2. Creates a concept "(node+const)" for each constant: X = const & node_formula
//   3. Replaces the original node in Domain.Nodes with the new sub-nodes
//   4. Recomputes
// If constants is nil, the caller should provide the available constants.
func (cs *ConceptSession) Splatter(concept string, constants []string) error {
	c, ok := cs.Domain.Concepts[concept]
	if !ok {
		return fmt.Errorf("concept %q not found", concept)
	}
	if len(constants) == 0 {
		return fmt.Errorf("no constants available for sort of %q — need a counterexample first", concept)
	}
	cs.push()

	// Create one sub-concept per constant, matching Python:
	//   c1 = concept_from_formula(Equals(Variable('X', node.sort), cons))
	//   concepts[node.name+'.splatter'] = enum_concepts(...)
	//   domain.split(node.name, splatter_set)
	var newNames []string
	for _, constName := range constants {
		subName := fmt.Sprintf("(%s+%s)", concept, constName)
		cs.Domain.Concepts[subName] = &Concept{
			Name:      subName,
			Variables: c.Variables,
			Formula:   fmt.Sprintf("(%s) & (X = %s)", c.Formula, constName),
			Sorts:     c.Sorts,
			Arity:     c.Arity,
		}
		newNames = append(newNames, subName)
	}

	// Replace the original node with the new sub-nodes in Domain.Nodes
	var updatedNodes []string
	for _, n := range cs.Domain.Nodes {
		if n == concept {
			updatedNodes = append(updatedNodes, newNames...)
		} else {
			updatedNodes = append(updatedNodes, n)
		}
	}
	cs.Domain.Nodes = updatedNodes

	// Remove the original concept
	delete(cs.Domain.Concepts, concept)

	cs.Recompute()
	return nil
}

// Recompute recomputes the abstract value from the domain.
// Full implementation requires a CDConceptDomain and a state formula,
// which are managed by the analysis session. When those are available,
// this calls Alpha(cdDomain, stateFormula, cache, projection) and
// updates AbstractValue. Without a state formula, it resets the cache.
func (cs *ConceptSession) Recompute() {
	// Without an analysis session providing the state formula and
	// CDConceptDomain, we can only clear the abstract value cache
	// so the next query recomputes from scratch.
	cs.AbstractValue = make(map[string]bool)
}
