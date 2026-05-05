package webui

// Full port of ivy_graph_ui.py — interactive concept graph widget with
// actions for splitting, materializing, and manipulating concept nodes/edges.

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"sort"
	"strings"
	"sync"
)

// MenuDef describes a top-level menu in the UI.
type MenuDef struct {
	Type  string     `json:"type"`
	Label string     `json:"label"`
	Items []MenuItem `json:"items"`
}

// MenuItem describes a single item inside a menu.
type MenuItem struct {
	Type     string `json:"type"`
	Label    string `json:"label"`
	Action   string `json:"action"`
	Dispatch string `json:"dispatch"`
	Dialog   string `json:"dialog,omitempty"`
	Enabled  bool   `json:"enabled"`
}

// ActionEntry describes a context-menu action (label + callback key).
type ActionEntry struct {
	Label  string `json:"label"`
	Action string `json:"action"` // identifier sent to server
}

// FactSelection is the browser-facing state for one constraint/fact line.
type FactSelection struct {
	Index    int    `json:"index"`
	Text     string `json:"text"`
	Selected bool   `json:"selected"`
}

// GraphWidget manages concept-graph interactions (Python: class GraphWidget).
// It wraps a GraphStack and provides all the user-facing operations.
type GraphWidget struct {
	mu sync.Mutex

	GraphStack     *GraphStack
	Parent         interface{} // AnalysisGraphUI or similar
	UIParent       interface{} // the top-level UI for dialogs
	UpdateCallback func()

	// Selection state.
	NodeSelection map[string]bool
	EdgeSelection map[string]bool
	Mark          string // selected node for future operations

	// FactElems maps fact formulas to the graph elements they reference.
	FactElems map[string][][]string

	// Constraint selection mirrors tk_graph_ui.selected_constraints. It is
	// keyed by the current printed constraint list and defaults new lists to on.
	SelectedConstraints []bool
	constraintFactTexts []string

	// StructureRenaming maps original names to display names.
	StructureRenaming map[string]string
}

// NewGraphWidget creates a new GraphWidget around a graph stack.
func NewGraphWidget(gs *GraphStack) *GraphWidget {
	return &GraphWidget{
		GraphStack:        gs,
		NodeSelection:     make(map[string]bool),
		EdgeSelection:     make(map[string]bool),
		StructureRenaming: make(map[string]string),
	}
}

// G returns the current graph from the stack (Python: GraphWidget.g property).
func (w *GraphWidget) G() *Graph {
	return w.GraphStack.Current
}

// Menus returns the menu structure (Python: GraphWidget.menus).
func (w *GraphWidget) Menus() []MenuDef {
	return []MenuDef{
		{
			Type:  "menu",
			Label: "Action",
			Items: []MenuItem{
				{Type: "button", Label: "Undo", Action: "undo"},
				{Type: "button", Label: "Redo", Action: "redo"},
				{Type: "button", Label: "PDR step", Action: "pdr_step"},
				{Type: "button", Label: "Concrete", Action: "concrete"},
				{Type: "button", Label: "Gather", Action: "gather"},
				{Type: "button", Label: "Reverse", Action: "reverse"},
				{Type: "button", Label: "Path reach", Action: "path_reach"},
				{Type: "button", Label: "Reach", Action: "reach"},
				{Type: "button", Label: "Conjecture", Action: "conjecture"},
				{Type: "button", Label: "Backtrack", Action: "backtrack"},
				{Type: "button", Label: "Recalculate", Action: "recalculate"},
				{Type: "button", Label: "Diagram", Action: "diagram"},
				{Type: "button", Label: "Remember", Action: "remember"},
				{Type: "button", Label: "Export", Action: "export"},
			},
		},
		{
			Type:  "menu",
			Label: "View",
			Items: []MenuItem{
				{Type: "button", Label: "Add relation", Action: "add_relation"},
			},
		},
	}
}

// SetUpdateCallback sets the function called when relations change.
func (w *GraphWidget) SetUpdateCallback(cb func()) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.UpdateCallback = cb
}

// Checkpoint records a checkpoint on the graph stack.
func (w *GraphWidget) Checkpoint(setBacktrackPoint bool) {
	w.GraphStack.Checkpoint(setBacktrackPoint)
}

// Undo rolls back the graph and updates.
func (w *GraphWidget) Undo() {
	w.GraphStack.Undo()
	w.Update()
}

// Redo replays the last undo and updates.
func (w *GraphWidget) Redo() {
	w.GraphStack.Redo()
	w.Update()
}

// Backtrack undoes to the most recent backtrack point (Python: GraphWidget.backtrack).
func (w *GraphWidget) Backtrack() {
	gs := w.GraphStack
	for gs.CanUndo() && !gs.Current.HasAttribute("backtrack_point") {
		gs.Undo()
	}
	if gs.Current.HasAttribute("backtrack_point") {
		gs.Current.RemoveAttribute("backtrack_point")
	}
	w.Update()
}

// MakeConcrete adds concrete state constraints to the current state.
func (w *GraphWidget) MakeConcrete() {
	w.Checkpoint(false)
	g := w.G()
	combined := g.State + " & " + g.Concrete
	g.SetState(combined, true, false, false)
	w.Update()
}

// Gather gathers definite facts from visible relations (Python: GraphWidget.gather).
func (w *GraphWidget) Gather() {
	w.Checkpoint(false)
	g := w.G()
	g.GetFacts(true)
	w.Update()
}

// ClearElemSelection clears node and edge selections.
func (w *GraphWidget) ClearElemSelection() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.NodeSelection = make(map[string]bool)
	w.EdgeSelection = make(map[string]bool)
}

// SelectNode selects or deselects a node.
func (w *GraphWidget) SelectNode(nodeID string, selected bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if selected {
		w.NodeSelection[nodeID] = true
	} else {
		delete(w.NodeSelection, nodeID)
	}
}

// SelectEdge selects or deselects an edge (stored as "rel|src|tgt").
func (w *GraphWidget) SelectEdge(edgeKey string, selected bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if selected {
		w.EdgeSelection[edgeKey] = true
	} else {
		delete(w.EdgeSelection, edgeKey)
	}
}

// SetParentState changes the parent state, keeping checkboxes (Python: GraphWidget.set_parent_state).
func (w *GraphWidget) SetParentState(newParentState interface{}, clauses string, reset bool) {
	w.Checkpoint(false)
	g := w.G()
	g.ParentState = newParentState
	if clauses == "" {
		// Use state from parent (stub).
		clauses = g.State
	}
	g.SetState(clauses, true, false, reset)
	w.UpdateRelations()
	w.Update()
}

// AddConcept adds a concept to the graph (Python: GraphWidget.add_concept).
func (w *GraphWidget) AddConcept(c *Concept) {
	w.G().NewRelation(c)
	w.UpdateRelations()
}

// AddConceptFromString parses and adds a concept from a text string.
func (w *GraphWidget) AddConceptFromString(text string) error {
	if text == "" {
		return fmt.Errorf("empty concept text")
	}
	c := &Concept{
		Name:    text,
		Formula: text,
	}
	w.AddConcept(c)
	return nil
}

// SplitConcept splits a node by a predicate concept (Python: GraphWidget.split).
func (w *GraphWidget) SplitConcept(predicateID, nodeID string) error {
	w.Checkpoint(false)
	err := w.G().Split(nodeID, predicateID)
	if err != nil {
		return err
	}
	w.Update()
	return nil
}

// SupposeEmpty marks a node as empty (Python: GraphWidget.empty).
func (w *GraphWidget) SupposeEmpty(nodeID string) error {
	w.Checkpoint(false)
	err := w.G().Empty(nodeID, true)
	if err != nil {
		return err
	}
	w.Update()
	return nil
}

// RemoveConcept removes a concept from the domain.
func (w *GraphWidget) RemoveConcept(conceptID string) error {
	w.Checkpoint(false)
	err := w.G().ConceptSess.RemoveConcept(conceptID)
	if err != nil {
		return err
	}
	w.Update()
	return nil
}

// MaterializeNode materializes a node concept (Python: GraphWidget.materialize).
func (w *GraphWidget) MaterializeNode(nodeID string) (string, error) {
	w.Checkpoint(false)
	witness, err := w.G().MaterializeNode(nodeID, false)
	if err != nil {
		return "", err
	}
	w.UpdateRelations()
	w.Update()
	return witness, nil
}

// MaterializeEdge materializes an edge concept (Python: GraphWidget.materialize_edge).
func (w *GraphWidget) MaterializeEdge(relID, headID, tailID string, truth bool) ([]string, error) {
	w.Checkpoint(false)
	witnesses, err := w.G().MaterializeEdge(relID, headID, tailID, truth, false)
	if err != nil {
		return nil, err
	}
	w.UpdateRelations()
	w.Update()
	return witnesses, nil
}

// DematerializeEdge materializes a negative edge (Python: GraphWidget.dematerialize_edge).
func (w *GraphWidget) DematerializeEdge(relID, headID, tailID string) ([]string, error) {
	return w.MaterializeEdge(relID, headID, tailID, false)
}

// Splatter splits a node using all available constants (Python: Graph.splatter).
func (w *GraphWidget) Splatter(nodeID string) {
	w.Checkpoint(false)
	g := w.G()
	if g.InteractiveSess != nil {
		s := g.InteractiveSess
		concept := s.Domain.Concepts.GetConcept(nodeID)
		if concept == nil || concept.Arity() != 1 {
			w.Update()
			return
		}
		s.Push()
		cSort := concept.Variables[0].VSort

		seen := make(map[string]bool)
		var constants []*goivy.Const
		for _, sc := range s.SupposeConstraints {
			for _, sym := range goivy.UsedSymbolsAst(sc).All() {
				if c, ok := sym.(*goivy.Const); ok {
					if seen[c.Name] {
						continue
					}
					if goivy.SortEqual(c.CSort, cSort) || webuiIsTopSort(c.CSort) {
						constants = append(constants, c)
						seen[c.Name] = true
					}
				}
			}
		}

		if len(constants) > 0 {
			splatterName := nodeID + ".splatter"
			var eqNames []string
			for _, c := range constants {
				X := webuiMustVar("X", c.CSort)
				eq, _ := goivy.NewEq(X, c)
				eqName := "=" + c.Name
				s.Domain.Concepts.SetConcept(eqName,
					MustCDConcept(eqName, []*goivy.LogicVariable{X}, eq))
				eqNames = append(eqNames, eqName)
			}
			s.Domain.Concepts.SetSet(splatterName, NewCDConceptSet(eqNames...))
			s.Domain.Split(nodeID, splatterName)
		}
		s.Recompute(nil)
	} else {
		g.ConceptSess.Splatter(nodeID, nil)
	}
	w.Update()
}

// AddProjection adds a projection concept.
func (w *GraphWidget) AddProjection(c *Concept) {
	w.AddConcept(c)
}

// UpdateRelations notifies that relations have changed.
func (w *GraphWidget) UpdateRelations() {
	w.mu.Lock()
	cb := w.UpdateCallback
	w.mu.Unlock()
	if cb != nil {
		cb()
	}
}

// Update triggers a re-render of the graph.
func (w *GraphWidget) Update() {
	w.G().Recompute()
}

// Recalculate recalculates the current state from the parent (Python: GraphWidget.recalculate).
func (w *GraphWidget) Recalculate() {
	g := w.G()
	if g.ParentState != nil {
		if agui, ok := w.Parent.(*AnalysisGraphUI); ok && agui != nil && agui.AG != nil {
			if ps, ok := g.ParentState.(*goivy.State); ok && ps.Clauses != nil {
				clauses := ps.Clauses.ToFormula()
				if g.InteractiveSess != nil {
					g.InteractiveSess.State = clauses
					g.InteractiveSess.Recompute(nil)
				}
				g.SetState(clauses.String(), true, false, false)
			}
		}
	}
	w.Update()
}

// SetState sets the current state (Python: GraphWidget.set_state).
func (w *GraphWidget) SetState(clauses string) {
	w.Checkpoint(false)
	w.G().SetState(clauses, true, false, false)
	w.UpdateRelations()
	w.Update()
}

// SetFacts sets the current constraint facts (Python: GraphWidget.set_facts).
func (w *GraphWidget) SetFacts(facts []string) {
	w.Checkpoint(false)
	w.G().SetFacts(facts)
	w.Update()
}

// GetNodeActions returns the context-menu actions for a node.
func (w *GraphWidget) GetNodeActions(nodeID string, click string) []ActionEntry {
	if click == "left" {
		actions := []ActionEntry{
			{Label: "Select", Action: "select"},
			{Label: "Empty", Action: "empty"},
			{Label: "Materialize", Action: "materialize"},
			{Label: "Materialize edge", Action: "materialize_from_selected"},
			{Label: "Splatter", Action: "splatter"},
			{Label: "Split with...", Action: ""},
			{Label: "---", Action: ""},
		}
		actions = append(actions, w.GetNodeSplittingActions(nodeID)...)
		actions = append(actions, w.GetNodeProjectionActions(nodeID)...)
		return actions
	}
	return nil
}

// GetNodeSplittingActions returns splitting predicates for a node.
func (w *GraphWidget) GetNodeSplittingActions(nodeID string) []ActionEntry {
	g := w.G()
	node := g.ConceptFromID(nodeID)
	if node == nil || len(node.Sorts) == 0 {
		return nil
	}
	nodeSort := node.Sorts[0]

	var result []ActionEntry
	for _, c := range g.ConceptSess.Domain.Concepts {
		if len(c.Variables) == 1 && len(c.Sorts) > 0 && c.Sorts[0] == nodeSort {
			result = append(result, ActionEntry{
				Label:  g.ConceptLabel(c),
				Action: fmt.Sprintf("split:%s", c.Name),
			})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Label < result[j].Label })
	return result
}

// GetNodeProjectionActions returns projection actions for a node.
func (w *GraphWidget) GetNodeProjectionActions(nodeID string) []ActionEntry {
	// Stub: real implementation queries concept_session.get_projections.
	return nil
}

// GetEdgeActions returns the context-menu actions for an edge.
func (w *GraphWidget) GetEdgeActions(click string) []ActionEntry {
	if click != "left" {
		return nil
	}
	return []ActionEntry{
		{Label: "Empty", Action: "empty_edge"},
		{Label: "Materialize", Action: "materialize_edge"},
		{Label: "Dematerialize", Action: "dematerialize_edge"},
	}
}

// ShowRelation enables display checkboxes for a concept.
// boxes is a string of checkbox codes: "+" = all_to_all, "?" = edge_unknown,
// "-" = none_to_none, "T" = transitive.
func (w *GraphWidget) ShowRelation(concept *Concept, boxes string, value bool, update bool) {
	if concept == nil {
		return
	}
	for _, b := range boxes {
		switch b {
		case '+':
			w.G().Checks.SetEdgeCheckbox(concept.Name, EdgeDisplayAllToAll, value)
			w.G().Checks.SetNodeLabelCheckbox(concept.Name, NodeLabelNecessarily, value)
		case '?':
			w.G().Checks.SetEdgeCheckbox(concept.Name, EdgeDisplayUnknown, value)
			w.G().Checks.SetNodeLabelCheckbox(concept.Name, NodeLabelMaybe, value)
		case '-':
			w.G().Checks.SetEdgeCheckbox(concept.Name, EdgeDisplayNoneToNone, value)
			w.G().Checks.SetNodeLabelCheckbox(concept.Name, NodeLabelNecessarilyNot, value)
		case 'T':
			w.G().Checks.SetEdgeCheckbox(concept.Name, EdgeDisplayTransitive, value)
		}
	}
	if update {
		w.Update()
	}
}

// ClearEdges disables all edge display checkboxes.
func (w *GraphWidget) ClearEdges() {
	w.G().Checks.mu.Lock()
	defer w.G().Checks.mu.Unlock()
	for _, boxes := range w.G().Checks.EdgeDisplayCheckboxes {
		for _, opt := range boxes {
			opt.Val = false
		}
	}
}

// ApplyStructureRenaming replaces names in a string according to the renaming map.
func (w *GraphWidget) ApplyStructureRenaming(s string) string {
	w.mu.Lock()
	defer w.mu.Unlock()
	for orig, repl := range w.StructureRenaming {
		s = strings.ReplaceAll(s, orig, repl)
	}
	return s
}

// Node finds a node whose label lines match all given labels.
func (w *GraphWidget) Node(labels ...string) *Concept {
	return w.G().FindNodeWithLabels(labels)
}

// Relation finds a relation concept by its display label.
func (w *GraphWidget) Relation(label string) *Concept {
	return w.G().FindRelationWithLabel(label)
}

// HighlightSelectedFacts highlights graph elements corresponding to active facts.
func (w *GraphWidget) HighlightSelectedFacts() {
	w.ClearElemSelection()
	if w.FactElems == nil {
		return
	}
	for _, fact := range w.GetActiveFacts() {
		for _, elem := range w.FactElems[fact] {
			switch len(elem) {
			case 1:
				w.SelectNode(elem[0], true)
			case 3:
				w.SelectEdge(strings.Join(elem, "|"), true)
			}
		}
	}
}

// GetActiveFacts returns the currently active constraint facts.
func (w *GraphWidget) GetActiveFacts() []string {
	exprs := w.GetActiveFactExprs()
	facts := make([]string, 0, len(exprs))
	for _, expr := range exprs {
		facts = append(facts, expr.String())
	}
	return facts
}

// GetActiveFactExprs returns the selected constraint expressions.
func (w *GraphWidget) GetActiveFactExprs() []goivy.Expr {
	exprs := w.constraintExprs()
	w.mu.Lock()
	defer w.mu.Unlock()
	texts := w.syncConstraintSelectionLocked(exprs)
	active := make([]goivy.Expr, 0, len(exprs))
	for i, expr := range exprs {
		if i < len(texts) && i < len(w.SelectedConstraints) && w.SelectedConstraints[i] {
			active = append(active, expr)
		}
	}
	return active
}

// ConstraintFacts returns the current constraints with selection state.
func (w *GraphWidget) ConstraintFacts() []FactSelection {
	exprs := w.constraintExprs()
	w.mu.Lock()
	defer w.mu.Unlock()
	texts := w.syncConstraintSelectionLocked(exprs)
	facts := make([]FactSelection, 0, len(texts))
	for i, text := range texts {
		selected := i < len(w.SelectedConstraints) && w.SelectedConstraints[i]
		facts = append(facts, FactSelection{Index: i, Text: text, Selected: selected})
	}
	return facts
}

// SetFactSelected toggles one rendered constraint fact.
func (w *GraphWidget) SetFactSelected(index int, selected bool) error {
	exprs := w.constraintExprs()
	w.mu.Lock()
	defer w.mu.Unlock()
	w.syncConstraintSelectionLocked(exprs)
	if index < 0 || index >= len(w.SelectedConstraints) {
		return fmt.Errorf("fact index %d out of range", index)
	}
	w.SelectedConstraints[index] = selected
	return nil
}

func (w *GraphWidget) constraintExprs() []goivy.Expr {
	g := w.G()
	if g == nil || g.InteractiveSess == nil {
		return nil
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	return append([]goivy.Expr{}, g.InteractiveSess.SupposeConstraints...)
}

func (w *GraphWidget) syncConstraintSelectionLocked(exprs []goivy.Expr) []string {
	texts := make([]string, len(exprs))
	for i, expr := range exprs {
		texts[i] = expr.String()
	}
	if !sameStringSlice(texts, w.constraintFactTexts) {
		w.constraintFactTexts = append([]string{}, texts...)
		w.SelectedConstraints = make([]bool, len(texts))
		for i := range w.SelectedConstraints {
			w.SelectedConstraints[i] = true
		}
	}
	return texts
}

func sameStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
