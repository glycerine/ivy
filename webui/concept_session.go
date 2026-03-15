package webui

import "fmt"

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
	// Mark in abstract value.
	cs.AbstractValue["node_info|none|"+concept] = true
	cs.Recompute()
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

// Materialize creates a concrete witness for a concept.
// Stub implementation for now.
func (cs *ConceptSession) Materialize(concept string) error {
	if _, ok := cs.Domain.Concepts[concept]; !ok {
		return fmt.Errorf("concept %q not found", concept)
	}
	cs.push()
	// Real implementation will invoke Z3 to find a witness constant.
	cs.Recompute()
	return nil
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

// RelationNames returns the names of all binary concepts (relations/edges)
// in the domain, for populating the state checkbox panel.
func (cs *ConceptSession) RelationNames() []string {
	var names []string
	for _, c := range cs.Domain.Combiners {
		if c.Source != "" && c.Target != "" {
			names = append(names, c.Name)
		}
	}
	// Also include concept names that look like binary relations
	for name, concept := range cs.Domain.Concepts {
		if concept.Arity == 2 {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		// Return domain edge names if available
		for name := range cs.Domain.Concepts {
			if name != "" {
				names = append(names, name)
			}
		}
	}
	return names
}

// Recompute recomputes the abstract value from the domain.
// Stub: real implementation will invoke concept_alpha.
func (cs *ConceptSession) Recompute() {
	// no-op stub; real implementation calls alpha(domain, formula, cache)
}
