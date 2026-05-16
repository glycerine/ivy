package webui

// Full port of ivy_graph.py — concept graph model with checkbox management,
// state, and rendering helpers.

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"sort"
	"strings"
	"sync"
)

// Edge display class constants (Python: _edge_display_classes).
const (
	EdgeDisplayAllToAll   = "all_to_all"   // "+"
	EdgeDisplayUnknown    = "edge_unknown" // "?"
	EdgeDisplayNoneToNone = "none_to_none" // "-"
	EdgeDisplayTransitive = "transitive"   // "≤"
)

// Node label display kind constants (Python: _node_label_display_checkboxes).
const (
	NodeLabelNecessarily    = "node_necessarily"
	NodeLabelMaybe          = "node_maybe"
	NodeLabelNecessarilyNot = "node_necessarily_not"
)

// EdgeDisplayCheckboxes is the ordered list of checkbox names for edges.
var EdgeDisplayCheckboxes = []string{
	EdgeDisplayAllToAll,
	EdgeDisplayUnknown,
	EdgeDisplayNoneToNone,
	EdgeDisplayTransitive,
}

// EdgeDisplayClasses is the subset used for non-transitive classification.
var EdgeDisplayClasses = []string{
	EdgeDisplayAllToAll,
	EdgeDisplayUnknown,
	EdgeDisplayNoneToNone,
}

// NodeLabelDisplayCheckboxes is the ordered list of checkbox names for node labels.
var NodeLabelDisplayCheckboxes = []string{
	NodeLabelNecessarily,
	NodeLabelMaybe,
	NodeLabelNecessarilyNot,
}

// Option wraps a boolean checkbox value (Python: class Option).
type Option struct {
	Val bool
}

// Value returns the checkbox state.
func (o *Option) Value() bool {
	return o.Val
}

// DisplayCheckboxes centralizes edge and node-label checkbox state.
type DisplayCheckboxes struct {
	mu                         sync.RWMutex
	EdgeDisplayCheckboxes      map[string]map[string]*Option // edgeName -> checkboxName -> Option
	NodeLabelDisplayCheckboxes map[string]map[string]*Option // labelName -> checkboxName -> Option
}

// NewDisplayCheckboxes creates a DisplayCheckboxes with empty maps.
func NewDisplayCheckboxes() *DisplayCheckboxes {
	return &DisplayCheckboxes{
		EdgeDisplayCheckboxes:      make(map[string]map[string]*Option),
		NodeLabelDisplayCheckboxes: make(map[string]map[string]*Option),
	}
}

// EdgeVisible returns true if an edge of the given class should be displayed.
func (dc *DisplayCheckboxes) EdgeVisible(edgeName, className string) bool {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	boxes, ok := dc.EdgeDisplayCheckboxes[edgeName]
	if !ok {
		return false
	}
	opt, ok := boxes[className]
	if !ok {
		return false
	}
	return opt.Val
}

// NodeLabelVisible returns true if a node label of the given kind should be displayed.
func (dc *DisplayCheckboxes) NodeLabelVisible(labelName, kind string) bool {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	boxes, ok := dc.NodeLabelDisplayCheckboxes[labelName]
	if !ok {
		return false
	}
	opt, ok := boxes[kind]
	if !ok {
		return false
	}
	return opt.Val
}

// EnsureEdge creates default checkboxes for an edge if not present.
func (dc *DisplayCheckboxes) EnsureEdge(edgeName string) map[string]*Option {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	boxes, ok := dc.EdgeDisplayCheckboxes[edgeName]
	if !ok {
		boxes = make(map[string]*Option, len(EdgeDisplayCheckboxes))
		for _, cb := range EdgeDisplayCheckboxes {
			boxes[cb] = &Option{Val: false}
		}
		dc.EdgeDisplayCheckboxes[edgeName] = boxes
	}
	return boxes
}

// EnsureNodeLabel creates default checkboxes for a node label if not present.
func (dc *DisplayCheckboxes) EnsureNodeLabel(labelName string) map[string]*Option {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	boxes, ok := dc.NodeLabelDisplayCheckboxes[labelName]
	if !ok {
		boxes = make(map[string]*Option, len(NodeLabelDisplayCheckboxes))
		for _, cb := range NodeLabelDisplayCheckboxes {
			boxes[cb] = &Option{Val: false}
		}
		dc.NodeLabelDisplayCheckboxes[labelName] = boxes
	}
	return boxes
}

// SetEdgeCheckbox sets a checkbox value for an edge.
func (dc *DisplayCheckboxes) SetEdgeCheckbox(edgeName, checkboxName string, val bool) {
	boxes := dc.EnsureEdge(edgeName)
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if opt, ok := boxes[checkboxName]; ok {
		opt.Val = val
	}
}

// SetNodeLabelCheckbox sets a checkbox value for a node label.
func (dc *DisplayCheckboxes) SetNodeLabelCheckbox(labelName, checkboxName string, val bool) {
	boxes := dc.EnsureNodeLabel(labelName)
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if opt, ok := boxes[checkboxName]; ok {
		opt.Val = val
	}
}

func edgeDisplayIndex(className string) int {
	for i, name := range EdgeDisplayCheckboxes {
		if name == className {
			return i
		}
	}
	return -1
}

func bareRelationName(name string) string {
	if i := strings.Index(name, "("); i >= 0 {
		return name[:i]
	}
	return name
}

// SetCheckboxClass sets a display checkbox by class name. It mirrors
// Python Graph.set_checkbox by updating both edge and node-label maps
// for the shared +/?/- columns.
func (dc *DisplayCheckboxes) SetCheckboxClass(name, className string, val bool) {
	name = bareRelationName(name)
	idx := edgeDisplayIndex(className)
	if idx < 0 {
		return
	}
	dc.SetEdgeCheckbox(name, className, val)
	if idx < len(NodeLabelDisplayCheckboxes) {
		dc.SetNodeLabelCheckbox(name, NodeLabelDisplayCheckboxes[idx], val)
	}
}

// Snapshot returns a JSON-ready copy of checkbox state.
func (dc *DisplayCheckboxes) Snapshot() *Toggles {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	out := &Toggles{
		Edges:  make(map[string]map[string]bool, len(dc.EdgeDisplayCheckboxes)),
		Labels: make(map[string]map[string]bool, len(dc.NodeLabelDisplayCheckboxes)),
	}
	for name, boxes := range dc.EdgeDisplayCheckboxes {
		out.Edges[name] = make(map[string]bool, len(boxes))
		for className, opt := range boxes {
			out.Edges[name][className] = opt.Val
		}
	}
	for name, boxes := range dc.NodeLabelDisplayCheckboxes {
		out.Labels[name] = make(map[string]bool, len(boxes))
		for className, opt := range boxes {
			out.Labels[name][className] = opt.Val
		}
	}
	return out
}

// Graph is the central concept graph model (Python: class Graph).
// It manages the concept domain, checkbox display state, and graph operations.
type Graph struct {
	mu sync.RWMutex

	// Sorts in the graph (type sort names).
	Sorts []string

	// ParentState holds a reference to the parent interp state.
	ParentState interface{}

	// ConceptSess is the interactive concept session that holds domain + abstract value.
	ConceptSess *ConceptSession

	// InteractiveSess is the full Z3-backed concept session (nil when no Z3 context).
	// Graph methods delegate to this when available for real solver-backed operations.
	InteractiveSess *ConceptInteractiveSession

	// Display state: checkbox management.
	Checks *DisplayCheckboxes

	// NewRelations tracks concepts added since last reset.
	NewRelations []string

	// State is the current formula state (clauses).
	State string

	// Concrete is the concrete state (if any).
	Concrete string

	// CyElems is the last computed Cytoscape element set.
	CyElems *WebUICyElements

	// Attributes are string tags on this graph (e.g. "backtrack_point").
	Attributes []string

	// ReverseResult is set when reverse produces a pre/post pair.
	ReverseResult []string
}

// NewGraph creates a new Graph for the given sort names (Python: Graph.__init__).
func NewGraph(sorts []string, parentState interface{}) *Graph {
	g := &Graph{
		Sorts:        sorts,
		ParentState:  parentState,
		ConceptSess:  NewConceptSession(),
		Checks:       NewDisplayCheckboxes(),
		NewRelations: nil,
		Attributes:   nil,
	}
	g.Reset()
	return g
}

// Reset reinitializes the concept domain (Python: Graph.reset).
func (g *Graph) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ConceptSess = NewConceptSession()
	g.NewRelations = nil
	// Initialize one node concept per sort.
	for _, s := range g.Sorts {
		name := "X:" + s + ".X:" + s + "=X:" + s
		g.ConceptSess.Domain.Concepts[name] = &Concept{
			Name:      name,
			Variables: []string{"X"},
			Formula:   fmt.Sprintf("X:%s = X:%s", s, s),
			Sorts:     []string{s},
		}
	}
}

// ConceptDomainRef returns the current concept domain.
func (g *Graph) ConceptDomainRef() *ConceptDomain {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.ConceptSess.Domain
}

// SortNames returns sorted sort names.
func (g *Graph) SortNames() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	names := make([]string, len(g.Sorts))
	copy(names, g.Sorts)
	sort.Strings(names)
	return names
}

// RelationIDs returns IDs of all concepts needing checkboxes (Python: Graph.relation_ids).
func (g *Graph) RelationIDs() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var ids []string
	for name := range g.ConceptSess.Domain.Concepts {
		ids = append(ids, name)
	}
	sort.Strings(ids)
	return ids
}

// NodeIDs returns IDs of all node concepts.
func (g *Graph) NodeIDs() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var ids []string
	for name, c := range g.ConceptSess.Domain.Concepts {
		if len(c.Sorts) == 1 {
			ids = append(ids, name)
		}
	}
	sort.Strings(ids)
	return ids
}

// ConceptFromID retrieves a concept by name.
func (g *Graph) ConceptFromID(id string) *Concept {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.ConceptSess.Domain.Concepts[id]
}

// ConceptLabel returns the shorthand display label for a concept (Python: Graph.concept_label).
func (g *Graph) ConceptLabel(c *Concept) string {
	if c == nil {
		return ""
	}
	// Abbreviate single-variable formulas.
	if len(c.Variables) == 1 {
		// Strip the variable prefix for readability.
		fmla := c.Formula
		v := c.Variables[0]
		fmla = strings.ReplaceAll(fmla, v+":", "")
		fmla = strings.ReplaceAll(fmla, " ", "")
		fmla = strings.ReplaceAll(fmla, "()", "")
		return fmla
	}
	return c.Formula
}

// NodeLabel returns the top-level label for a node concept (Python: Graph.node_label).
func (g *Graph) NodeLabel(c *Concept) string {
	if c == nil || len(c.Sorts) == 0 {
		return ""
	}
	return c.Sorts[0]
}

// NewRelation adds a new relation concept to the domain (Python: Graph.new_relation).
func (g *Graph) NewRelation(concept *Concept) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.ConceptSess.Domain.Concepts[concept.Name]; !exists {
		g.ConceptSess.Domain.Concepts[concept.Name] = concept
		g.NewRelations = append(g.NewRelations, concept.Name)
	}
}

// SetState sets the current clauses state (Python: Graph.set_state).
func (g *Graph) SetState(clauses string, recomp bool, clearConstraints bool, reset bool) error {
	g.mu.Lock()
	g.State = clauses
	if clearConstraints && g.InteractiveSess != nil {
		g.InteractiveSess.SupposeConstraints = nil
	}
	if reset {
		g.mu.Unlock()
		g.Reset()
		g.mu.Lock()
	}
	g.mu.Unlock()
	if recomp {
		return g.Recompute()
	}
	return nil
}

// SetConcrete stores concrete clauses (Python: Graph.set_concrete).
func (g *Graph) SetConcrete(clauses string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Concrete = clauses
}

// Recompute recomputes abstract value and cy elements (Python: Graph.recompute).
func (g *Graph) Recompute() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ConceptSess.Recompute()
	if g.InteractiveSess != nil {
		if err := g.InteractiveSess.Recompute(nil); err != nil {
			return err
		}
	}
	g.CyElems = RenderConceptGraph(g.ConceptSess, g.Checks)
	return nil
}

// SetCheckbox sets a checkbox by concept name and index (Python: Graph.set_checkbox).
// idx maps: 0=all_to_all/node_necessarily, 1=edge_unknown/node_maybe,
// 2=none_to_none/node_necessarily_not, 3=transitive (edges only).
func (g *Graph) SetCheckbox(obj string, idx int, val bool) {
	obj = bareRelationName(obj)
	if idx < len(EdgeDisplayCheckboxes) {
		g.Checks.SetEdgeCheckbox(obj, EdgeDisplayCheckboxes[idx], val)
	}
	if idx < len(NodeLabelDisplayCheckboxes) {
		g.Checks.SetNodeLabelCheckbox(obj, NodeLabelDisplayCheckboxes[idx], val)
	}
}

// SetCheckboxClass sets a checkbox by display class name.
func (g *Graph) SetCheckboxClass(obj, className string, val bool) {
	g.Checks.SetCheckboxClass(obj, className, val)
}

// GetCheckbox gets a checkbox value by concept name and index.
func (g *Graph) GetCheckbox(obj string, idx int) bool {
	if idx < len(EdgeDisplayCheckboxes) {
		return g.Checks.EdgeVisible(obj, EdgeDisplayCheckboxes[idx])
	}
	if idx < len(NodeLabelDisplayCheckboxes) {
		return g.Checks.NodeLabelVisible(obj, NodeLabelDisplayCheckboxes[idx])
	}
	return false
}

// ShowCheckboxes returns a debug string of current checkbox state.
func (g *Graph) ShowCheckboxes() string {
	g.Checks.mu.RLock()
	defer g.Checks.mu.RUnlock()
	var lines []string
	for name, boxes := range g.Checks.EdgeDisplayCheckboxes {
		var enabled []string
		for k, v := range boxes {
			if v.Val {
				enabled = append(enabled, k)
			}
		}
		sort.Strings(enabled)
		lines = append(lines, fmt.Sprintf("%s:{%s}", name, strings.Join(enabled, ",")))
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// Projection decides whether a concept class/combiner should be projected.
// Returns true if visible (Python: Graph.projection).
func (g *Graph) Projection(conceptName, conceptClass string, conceptCombiner string) bool {
	if conceptClass == "node_labels" || conceptClass == "edges" || conceptClass == "enum" {
		var boxes map[string]*Option
		if conceptClass == "edges" {
			boxes = g.Checks.EnsureEdge(conceptName)
		} else {
			boxes = g.Checks.EnsureNodeLabel(conceptName)
		}
		g.Checks.mu.RLock()
		defer g.Checks.mu.RUnlock()
		if conceptCombiner == "" {
			for k, v := range boxes {
				if k != EdgeDisplayTransitive && v.Val {
					return true
				}
			}
			return false
		}
		if opt, ok := boxes[conceptCombiner]; ok {
			return opt.Val
		}
		return false
	}
	return true
}

// Copy returns a deep copy of the graph (Python: Graph.copy).
func (g *Graph) Copy() *Graph {
	g.mu.RLock()
	defer g.mu.RUnlock()
	c := &Graph{
		Sorts:       make([]string, len(g.Sorts)),
		ParentState: g.ParentState,
		ConceptSess: &ConceptSession{
			Domain:        g.ConceptSess.Domain.Copy(),
			AbstractValue: make(map[string]bool, len(g.ConceptSess.AbstractValue)),
		},
		Checks:       NewDisplayCheckboxes(),
		NewRelations: make([]string, len(g.NewRelations)),
		State:        g.State,
		Concrete:     g.Concrete,
		CyElems:      g.CyElems,
		Attributes:   make([]string, len(g.Attributes)),
	}
	copy(c.Sorts, g.Sorts)
	copy(c.NewRelations, g.NewRelations)
	copy(c.Attributes, g.Attributes)
	for k, v := range g.ConceptSess.AbstractValue {
		c.ConceptSess.AbstractValue[k] = v
	}
	// Deep copy checkbox state.
	for name, boxes := range g.Checks.EdgeDisplayCheckboxes {
		c.Checks.EdgeDisplayCheckboxes[name] = make(map[string]*Option, len(boxes))
		for k, v := range boxes {
			c.Checks.EdgeDisplayCheckboxes[name][k] = &Option{Val: v.Val}
		}
	}
	for name, boxes := range g.Checks.NodeLabelDisplayCheckboxes {
		c.Checks.NodeLabelDisplayCheckboxes[name] = make(map[string]*Option, len(boxes))
		for k, v := range boxes {
			c.Checks.NodeLabelDisplayCheckboxes[name][k] = &Option{Val: v.Val}
		}
	}
	if g.ReverseResult != nil {
		c.ReverseResult = make([]string, len(g.ReverseResult))
		copy(c.ReverseResult, g.ReverseResult)
	}
	return c
}

// Split splits a node concept by a predicate (Python: Graph.split).
func (g *Graph) Split(nodeID, predicateID string) error {
	return g.ConceptSess.Split(nodeID, predicateID)
}

// Empty supposes a node is empty (Python: Graph.empty).
func (g *Graph) Empty(nodeID string, recompute bool) error {
	err := g.ConceptSess.SupposeEmpty(nodeID)
	if err != nil {
		return err
	}
	if recompute {
		return g.Recompute()
	}
	return nil
}

// MaterializeNode materializes a witness for a node concept (Python: Graph.materialize).
func (g *Graph) MaterializeNode(nodeID string, recompute bool) (string, error) {
	err := g.ConceptSess.Materialize(nodeID)
	if err != nil {
		return "", err
	}
	witnessName := nodeID + "_witness"
	if recompute {
		if err := g.Recompute(); err != nil {
			return "", err
		}
	}
	return witnessName, nil
}

// MaterializeEdge materializes a witness for an edge (Python: Graph.materialize_edge).
func (g *Graph) MaterializeEdge(relID, headID, tailID string, truth bool, recompute bool) ([]string, error) {
	g.mu.Lock()
	if g.InteractiveSess != nil {
		g.InteractiveSess.Push()
		headC, tailC, err := g.InteractiveSess.materializeEdge(relID, headID, tailID, truth)
		if err != nil {
			g.mu.Unlock()
			return nil, err
		}
		if recompute {
			if err := g.InteractiveSess.Recompute(nil); err != nil {
				g.mu.Unlock()
				return nil, err
			}
		}
		g.mu.Unlock()
		var witnesses []string
		if headC != nil {
			witnesses = append(witnesses, headC.Name)
		}
		if tailC != nil {
			witnesses = append(witnesses, tailC.Name)
		}
		return witnesses, nil
	}
	g.mu.Unlock()
	witnesses := []string{
		headID + "_witness",
		tailID + "_witness",
	}
	if recompute {
		if err := g.Recompute(); err != nil {
			return nil, err
		}
	}
	return witnesses, nil
}

// AddConstraints appends constraints to the concept session (Python: Graph.add_constraints).
func (g *Graph) AddConstraints(constraints []string, recompute bool) error {
	g.mu.Lock()
	if g.InteractiveSess != nil {
		for _, cs := range constraints {
			expr, err := parseConstraintString(cs)
			if err != nil {
				g.mu.Unlock()
				return fmt.Errorf("parse constraint %q: %w", cs, err)
			}
			g.InteractiveSess.Suppose(expr)
		}
		if recompute {
			if err := g.InteractiveSess.Recompute(nil); err != nil {
				g.mu.Unlock()
				return err
			}
		}
		g.mu.Unlock()
		return nil
	}
	g.mu.Unlock()
	if recompute {
		return g.Recompute()
	}
	return nil
}

// AddConstraintsExpr appends logic.Expr constraints to the interactive session.
func (g *Graph) AddConstraintsExpr(constraints []goivy.Expr, recompute bool) error {
	g.mu.Lock()
	if g.InteractiveSess != nil {
		for _, expr := range constraints {
			g.InteractiveSess.Suppose(expr)
		}
		if recompute {
			if err := g.InteractiveSess.Recompute(nil); err != nil {
				g.mu.Unlock()
				return err
			}
		}
	}
	g.mu.Unlock()
	return nil
}

// SetFacts sets the constraint facts (Python: Graph.set_facts).
func (g *Graph) SetFacts(facts []string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.InteractiveSess != nil {
		g.InteractiveSess.SupposeConstraints = nil
		for _, f := range facts {
			expr, err := parseConstraintString(f)
			if err != nil {
				return fmt.Errorf("parse fact %q: %w", f, err)
			}
			g.InteractiveSess.SupposeConstraints = append(g.InteractiveSess.SupposeConstraints, expr)
		}
	}
	return nil
}

// SetFactsExpr replaces suppose constraints with the given logic.Expr slice.
func (g *Graph) SetFactsExpr(facts []goivy.Expr) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.InteractiveSess != nil {
		g.InteractiveSess.SupposeConstraints = append([]goivy.Expr{}, facts...)
	}
}

// GetFacts gathers definite facts from the current abstract value (Python: Graph.get_facts).
func (g *Graph) GetFacts(definite bool) ([]string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.InteractiveSess != nil {
		proj := func(a, b, c string) bool { return true }
		facts, err := g.InteractiveSess.GetFacts(proj)
		if err != nil {
			return nil, err
		}
		g.InteractiveSess.SupposeConstraints = append([]goivy.Expr{}, facts...)
		result := make([]string, 0, len(facts))
		for _, f := range facts {
			result = append(result, f.String())
		}
		return result, nil
	}
	var result []string
	for k, v := range g.ConceptSess.AbstractValue {
		if v {
			result = append(result, k)
		}
	}
	return result, nil
}

// FindNodeWithLabels finds a node whose label lines match all given labels.
func (g *Graph) FindNodeWithLabels(labels []string) *Concept {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.CyElems == nil {
		return nil
	}
	for _, elem := range g.CyElems.Elements {
		if elem.Group != "nodes" {
			continue
		}
		lblStr, _ := elem.Data["label"].(string)
		nlabs := make(map[string]bool)
		for _, l := range strings.Split(lblStr, "\n") {
			nlabs[l] = true
		}
		match := true
		for _, l := range labels {
			if strings.HasPrefix(l, "!") {
				if nlabs[l[1:]] {
					match = false
					break
				}
			} else {
				if !nlabs[l] {
					match = false
					break
				}
			}
		}
		if match {
			obj, _ := elem.Data["obj"].(string)
			return g.ConceptSess.Domain.Concepts[obj]
		}
	}
	return nil
}

// FindRelationWithLabel finds a relation concept by its display label.
func (g *Graph) FindRelationWithLabel(label string) *Concept {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, c := range g.ConceptSess.Domain.Concepts {
		if g.ConceptLabel(c) == label {
			return c
		}
	}
	return nil
}

// HasAttribute checks whether the graph has a given attribute tag.
func (g *Graph) HasAttribute(attr string) bool {
	for _, a := range g.Attributes {
		if a == attr {
			return true
		}
	}
	return false
}

// RemoveAttribute removes an attribute tag from the graph.
func (g *Graph) RemoveAttribute(attr string) {
	for i, a := range g.Attributes {
		if a == attr {
			g.Attributes = append(g.Attributes[:i], g.Attributes[i+1:]...)
			return
		}
	}
}

// GraphStack provides undo/redo for concept graphs (Python: class GraphStack).
type GraphStack struct {
	Current   *Graph
	UndoStack []*Graph
	RedoStack []*Graph
}

// NewGraphStack creates a new GraphStack from a graph.
func NewGraphStack(g *Graph) *GraphStack {
	return &GraphStack{
		Current: g,
	}
}

// CanUndo returns whether undo is available.
func (gs *GraphStack) CanUndo() bool {
	return len(gs.UndoStack) > 0
}

// CanRedo returns whether redo is available.
func (gs *GraphStack) CanRedo() bool {
	return len(gs.RedoStack) > 0
}

// Undo rolls back to the most recent checkpoint.
func (gs *GraphStack) Undo() {
	if len(gs.UndoStack) == 0 {
		return
	}
	gs.RedoStack = append(gs.RedoStack, gs.Current)
	gs.Current = gs.UndoStack[len(gs.UndoStack)-1]
	gs.UndoStack = gs.UndoStack[:len(gs.UndoStack)-1]
}

// Redo undoes the most recent undo.
func (gs *GraphStack) Redo() {
	if len(gs.RedoStack) == 0 {
		return
	}
	gs.UndoStack = append(gs.UndoStack, gs.Current)
	gs.Current = gs.RedoStack[len(gs.RedoStack)-1]
	gs.RedoStack = gs.RedoStack[:len(gs.RedoStack)-1]
}

// Checkpoint records a checkpoint of the current graph.
func (gs *GraphStack) Checkpoint(setBacktrackPoint bool) {
	cp := gs.Current.Copy()
	if setBacktrackPoint {
		cp.Attributes = append(cp.Attributes, "backtrack_point")
	}
	gs.UndoStack = append(gs.UndoStack, cp)
	gs.RedoStack = nil
}

// StandardGraph creates a standard graph stack for the given sorts.
func StandardGraph(sorts []string, parentState interface{}) *GraphStack {
	g := NewGraph(sorts, parentState)
	return NewGraphStack(g)
}

// GetTransitiveReduction returns edges hidden by transitive reduction.
// edgeTuples are (edge, source, target) triples.
func GetTransitiveReduction(checks *DisplayCheckboxes, abstractValue map[string]bool, edgeTuples [][3]string) map[[3]string]bool {
	result := make(map[[3]string]bool)

	// Filter to all_to_all edges with transitive enabled.
	var filtered [][3]string
	for _, e := range edgeTuples {
		edge, src, tgt := e[0], e[1], e[2]
		if checks.EdgeVisible(edge, EdgeDisplayTransitive) {
			key := fmt.Sprintf("edge_info|all_to_all|%s|%s|%s", edge, src, tgt)
			if abstractValue[key] {
				filtered = append(filtered, e)
			}
		}
	}

	// Hide reflexive edges.
	var nonReflexive [][3]string
	for _, e := range filtered {
		if e[1] == e[2] {
			result[e] = true
		} else {
			nonReflexive = append(nonReflexive, e)
		}
	}

	// Transitive reduction: hide (x,z) if x->y and y->z.
	bySource := make(map[[2]string][]string) // (edge,source) -> []target
	for _, e := range nonReflexive {
		key := [2]string{e[0], e[1]}
		bySource[key] = append(bySource[key], e[2])
	}
	for _, e := range nonReflexive {
		key := [2]string{e[0], e[2]}
		// Match Python defaultdict auto-vivification
		if _, ok := bySource[key]; !ok {
			bySource[key] = nil
		}
		for _, z := range bySource[key] {
			result[[3]string{e[0], e[1], z}] = true
		}
	}

	return result
}

// GetShape returns the Cytoscape shape for a concept node.
func GetShape(_ string) string {
	return "octagon"
}

// NodeGT compares two node names for ordering.
func NodeGT(n1, n2 string) bool {
	return n1 > n2
}
