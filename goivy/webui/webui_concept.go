package webui

import (
	"strconv"
	"strings"
)

// Concept represents a single concept in the concept domain.
// A concept is a quantifier-free formula with first-order free variables.
type Concept struct {
	Name      string   `json:"name"`
	Variables []string `json:"variables"`
	Formula   string   `json:"formula"`
	Sorts     []string `json:"sorts"`
	Arity     int      `json:"arity"` // number of variables (1=unary, 2=binary/edge)
}

// ConceptCombiner represents a binary relation between concepts
// (edges in the concept graph).
type ConceptCombiner struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Source  string `json:"source"`
	Target  string `json:"target"`
	Formula string `json:"formula"`
}

// ConceptDomain holds all concepts and combiners for a session.
// Mirrors Python's ConceptDomain with categorization lists.
type ConceptDomain struct {
	Concepts  map[string]*Concept `json:"concepts"`
	Combiners []*ConceptCombiner  `json:"combiners"`
	// Categorization lists (matching Python concept.py):
	Nodes      []string `json:"nodes"`       // sort concepts (one per sort)
	Edges      []string `json:"edges"`       // binary relation concepts (arity 2)
	NodeLabels []string `json:"node_labels"` // unary relation concepts (arity 1, not sorts)
}

// NewConceptDomain creates an empty concept domain.
func NewConceptDomain() *ConceptDomain {
	return &ConceptDomain{
		Concepts: make(map[string]*Concept),
	}
}

// RelationIDs returns the concept names that need checkboxes in the state pane.
// Matches Python's Graph.relation_ids: edges + non-numeral node_labels.
func (d *ConceptDomain) RelationIDs() []string {
	defaultLabels := make(map[string]bool)
	for _, id := range d.DefaultLabelIDs() {
		defaultLabels[id] = true
	}
	var ids []string
	ids = append(ids, d.Edges...)
	for _, id := range d.NodeLabels {
		if !defaultLabels[id] {
			ids = append(ids, id)
		}
	}
	return ids
}

// DefaultLabelIDs returns node-label concepts that Python shows by default.
// Python treats numeral singleton labels specially; in the simplified web
// Concept model we recognize the common textual shape "X = <numeral>".
func (d *ConceptDomain) DefaultLabelIDs() []string {
	var ids []string
	for _, name := range d.NodeLabels {
		c := d.Concepts[name]
		if c == nil || len(c.Variables) != 1 {
			continue
		}
		parts := strings.Split(c.Formula, "=")
		if len(parts) != 2 {
			continue
		}
		rhs := strings.TrimSpace(parts[1])
		if idx := strings.Index(rhs, ":"); idx >= 0 {
			rhs = rhs[:idx]
		}
		if _, err := strconv.Atoi(rhs); err == nil {
			ids = append(ids, name)
		}
	}
	return ids
}

// Copy returns a shallow copy of the domain (concepts map is cloned).
func (d *ConceptDomain) Copy() *ConceptDomain {
	cp := &ConceptDomain{
		Concepts:   make(map[string]*Concept, len(d.Concepts)),
		Combiners:  make([]*ConceptCombiner, len(d.Combiners)),
		Nodes:      append([]string{}, d.Nodes...),
		Edges:      append([]string{}, d.Edges...),
		NodeLabels: append([]string{}, d.NodeLabels...),
	}
	for k, v := range d.Concepts {
		cp.Concepts[k] = v
	}
	copy(cp.Combiners, d.Combiners)
	return cp
}
