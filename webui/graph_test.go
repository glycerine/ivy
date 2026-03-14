package webui

import (
	"strings"
	"testing"
)

// --- Graph model tests ---

func TestNewGraph(t *testing.T) {
	g := NewGraph([]string{"node", "edge_sort"}, nil)
	if g == nil {
		t.Fatal("NewGraph returned nil")
	}
	if len(g.Sorts) != 2 {
		t.Errorf("expected 2 sorts, got %d", len(g.Sorts))
	}
	if g.ConceptSess == nil {
		t.Error("ConceptSess is nil")
	}
	if g.Checks == nil {
		t.Error("Checks is nil")
	}
}

func TestGraphSortNames(t *testing.T) {
	g := NewGraph([]string{"b_sort", "a_sort"}, nil)
	names := g.SortNames()
	if len(names) != 2 || names[0] != "a_sort" || names[1] != "b_sort" {
		t.Errorf("expected sorted names [a_sort b_sort], got %v", names)
	}
}

func TestGraphReset(t *testing.T) {
	g := NewGraph([]string{"mySort"}, nil)
	g.NewRelation(&Concept{Name: "extra", Formula: "extra(X)"})
	if len(g.NewRelations) != 1 {
		t.Fatalf("expected 1 new relation, got %d", len(g.NewRelations))
	}
	g.Reset()
	if len(g.NewRelations) != 0 {
		t.Error("NewRelations should be empty after reset")
	}
}

func TestGraphNewRelation(t *testing.T) {
	g := NewGraph([]string{"s"}, nil)
	c := &Concept{Name: "r1", Formula: "r1(X,Y)", Sorts: []string{"s", "s"}}
	g.NewRelation(c)
	if g.ConceptFromID("r1") == nil {
		t.Error("concept r1 not found after NewRelation")
	}
	// Adding again should not duplicate.
	g.NewRelation(c)
	if len(g.NewRelations) != 1 {
		t.Errorf("expected 1 new relation, got %d", len(g.NewRelations))
	}
}

func TestGraphConceptLabel(t *testing.T) {
	g := NewGraph([]string{"s"}, nil)
	c := &Concept{
		Name:      "test",
		Variables: []string{"X"},
		Formula:   "X:s = a",
		Sorts:     []string{"s"},
	}
	label := g.ConceptLabel(c)
	if label == "" {
		t.Error("ConceptLabel returned empty")
	}
	// Should abbreviate by removing variable prefix.
	if strings.Contains(label, "X:") {
		t.Errorf("expected abbreviation, got %q", label)
	}
}

func TestGraphNodeLabel(t *testing.T) {
	g := NewGraph([]string{"s"}, nil)
	c := &Concept{Name: "n", Sorts: []string{"mySort"}}
	label := g.NodeLabel(c)
	if label != "mySort" {
		t.Errorf("expected 'mySort', got %q", label)
	}
}

func TestGraphConceptLabelNil(t *testing.T) {
	g := NewGraph([]string{"s"}, nil)
	if g.ConceptLabel(nil) != "" {
		t.Error("ConceptLabel(nil) should return empty")
	}
}

func TestGraphNodeLabelNil(t *testing.T) {
	g := NewGraph([]string{"s"}, nil)
	if g.NodeLabel(nil) != "" {
		t.Error("NodeLabel(nil) should return empty")
	}
}

// --- DisplayCheckboxes tests ---

func TestDisplayCheckboxesEdge(t *testing.T) {
	dc := NewDisplayCheckboxes()
	dc.SetEdgeCheckbox("rel1", EdgeDisplayAllToAll, true)
	if !dc.EdgeVisible("rel1", EdgeDisplayAllToAll) {
		t.Error("expected rel1 all_to_all to be visible")
	}
	if dc.EdgeVisible("rel1", EdgeDisplayNoneToNone) {
		t.Error("rel1 none_to_none should not be visible")
	}
}

func TestDisplayCheckboxesNodeLabel(t *testing.T) {
	dc := NewDisplayCheckboxes()
	dc.SetNodeLabelCheckbox("lbl1", NodeLabelNecessarily, true)
	if !dc.NodeLabelVisible("lbl1", NodeLabelNecessarily) {
		t.Error("expected lbl1 node_necessarily to be visible")
	}
	if dc.NodeLabelVisible("lbl1", NodeLabelMaybe) {
		t.Error("lbl1 node_maybe should not be visible")
	}
}

func TestDisplayCheckboxesEnsure(t *testing.T) {
	dc := NewDisplayCheckboxes()
	boxes := dc.EnsureEdge("newedge")
	if len(boxes) != len(EdgeDisplayCheckboxes) {
		t.Errorf("expected %d checkboxes, got %d", len(EdgeDisplayCheckboxes), len(boxes))
	}
	// Ensure again should return same map.
	boxes2 := dc.EnsureEdge("newedge")
	if len(boxes2) != len(boxes) {
		t.Error("EnsureEdge should return existing map")
	}
}

func TestDisplayCheckboxesEnsureNodeLabel(t *testing.T) {
	dc := NewDisplayCheckboxes()
	boxes := dc.EnsureNodeLabel("lbl")
	if len(boxes) != len(NodeLabelDisplayCheckboxes) {
		t.Errorf("expected %d checkboxes, got %d", len(NodeLabelDisplayCheckboxes), len(boxes))
	}
}

func TestDisplayCheckboxesUnknownEdge(t *testing.T) {
	dc := NewDisplayCheckboxes()
	if dc.EdgeVisible("nonexistent", EdgeDisplayAllToAll) {
		t.Error("nonexistent edge should not be visible")
	}
}

func TestGraphSetCheckbox(t *testing.T) {
	g := NewGraph([]string{"s"}, nil)
	g.SetCheckbox("rel", 0, true) // idx 0 = all_to_all / node_necessarily
	if !g.GetCheckbox("rel", 0) {
		t.Error("checkbox at idx 0 should be true")
	}
	g.SetCheckbox("rel", 0, false)
	if g.GetCheckbox("rel", 0) {
		t.Error("checkbox at idx 0 should be false after clear")
	}
}

func TestGraphSetCheckboxTransitive(t *testing.T) {
	g := NewGraph([]string{"s"}, nil)
	g.SetCheckbox("rel", 3, true) // idx 3 = transitive
	if !g.Checks.EdgeVisible("rel", EdgeDisplayTransitive) {
		t.Error("transitive checkbox should be visible")
	}
}

func TestGraphShowCheckboxes(t *testing.T) {
	g := NewGraph([]string{"s"}, nil)
	g.SetCheckbox("r1", 0, true)
	s := g.ShowCheckboxes()
	if !strings.Contains(s, "r1") {
		t.Errorf("ShowCheckboxes should contain r1, got %q", s)
	}
}

// --- GraphStack tests ---

func TestGraphStackUndoRedo(t *testing.T) {
	g := NewGraph([]string{"s"}, nil)
	gs := NewGraphStack(g)

	if gs.CanUndo() {
		t.Error("should not be able to undo initially")
	}

	gs.Checkpoint(false)
	gs.Current.State = "modified"

	if !gs.CanUndo() {
		t.Error("should be able to undo after checkpoint")
	}

	gs.Undo()
	if gs.Current.State == "modified" {
		t.Error("state should be reverted after undo")
	}

	gs.Redo()
	if gs.Current.State != "modified" {
		t.Error("state should be restored after redo")
	}
}

func TestGraphStackCheckpointBacktrack(t *testing.T) {
	g := NewGraph([]string{"s"}, nil)
	gs := NewGraphStack(g)
	gs.Checkpoint(true) // set backtrack point
	gs.Current.State = "step1"
	gs.Checkpoint(false)
	gs.Current.State = "step2"

	// Backtrack to the backtrack point.
	for gs.CanUndo() && !gs.Current.HasAttribute("backtrack_point") {
		gs.Undo()
	}
	if gs.Current.HasAttribute("backtrack_point") {
		gs.Current.RemoveAttribute("backtrack_point")
	}
	if gs.Current.State == "step2" {
		t.Error("should have backtracked past step2")
	}
}

func TestGraphCopy(t *testing.T) {
	g := NewGraph([]string{"s"}, nil)
	g.State = "original"
	g.Attributes = append(g.Attributes, "test_attr")
	g.SetCheckbox("rel", 0, true)

	c := g.Copy()
	if c.State != "original" {
		t.Error("copy should have same state")
	}
	if !c.HasAttribute("test_attr") {
		t.Error("copy should have test_attr")
	}
	// Modify copy, ensure original is unchanged.
	c.State = "modified"
	if g.State == "modified" {
		t.Error("modifying copy should not affect original")
	}
}

func TestGraphAttributes(t *testing.T) {
	g := NewGraph([]string{"s"}, nil)
	if g.HasAttribute("foo") {
		t.Error("should not have foo initially")
	}
	g.Attributes = append(g.Attributes, "foo")
	if !g.HasAttribute("foo") {
		t.Error("should have foo after adding")
	}
	g.RemoveAttribute("foo")
	if g.HasAttribute("foo") {
		t.Error("should not have foo after removing")
	}
}

func TestGetTransitiveReduction(t *testing.T) {
	dc := NewDisplayCheckboxes()
	dc.SetEdgeCheckbox("le", EdgeDisplayTransitive, true)
	dc.SetEdgeCheckbox("le", EdgeDisplayAllToAll, true)

	av := map[string]bool{
		"edge_info|all_to_all|le|a|b": true,
		"edge_info|all_to_all|le|b|c": true,
		"edge_info|all_to_all|le|a|c": true,
		"edge_info|all_to_all|le|a|a": true,
	}
	edges := [][3]string{
		{"le", "a", "b"},
		{"le", "b", "c"},
		{"le", "a", "c"},
		{"le", "a", "a"},
	}
	hidden := GetTransitiveReduction(dc, av, edges)
	// a->c should be hidden (transitive through b)
	if !hidden[[3]string{"le", "a", "c"}] {
		t.Error("a->c should be hidden by transitive reduction")
	}
	// a->a should be hidden (reflexive)
	if !hidden[[3]string{"le", "a", "a"}] {
		t.Error("a->a should be hidden (reflexive)")
	}
}

func TestGraphProjection(t *testing.T) {
	g := NewGraph([]string{"s"}, nil)
	// With no checkboxes enabled, projection should return false for edges.
	if g.Projection("someEdge", "edges", "") {
		t.Error("projection should be false with no checkboxes")
	}
	// Enable one checkbox.
	g.Checks.SetEdgeCheckbox("someEdge", EdgeDisplayAllToAll, true)
	if !g.Projection("someEdge", "edges", "") {
		t.Error("projection should be true with all_to_all enabled")
	}
	// With specific combiner.
	if !g.Projection("someEdge", "edges", EdgeDisplayAllToAll) {
		t.Error("projection should be true for all_to_all combiner")
	}
	if g.Projection("someEdge", "edges", EdgeDisplayNoneToNone) {
		t.Error("projection should be false for none_to_none combiner")
	}
	// Non-edge/label class should always return true.
	if !g.Projection("someEdge", "other", "") {
		t.Error("projection should be true for non-edge/label class")
	}
}

func TestStandardGraph(t *testing.T) {
	gs := StandardGraph([]string{"s1", "s2"}, nil)
	if gs == nil {
		t.Fatal("StandardGraph returned nil")
	}
	if gs.Current == nil {
		t.Fatal("Current graph is nil")
	}
	if len(gs.Current.Sorts) != 2 {
		t.Errorf("expected 2 sorts, got %d", len(gs.Current.Sorts))
	}
}

func TestGetShape(t *testing.T) {
	if GetShape("anything") != "octagon" {
		t.Error("GetShape should return octagon")
	}
}

func TestNodeGT(t *testing.T) {
	if !NodeGT("b", "a") {
		t.Error("b should be > a")
	}
	if NodeGT("a", "b") {
		t.Error("a should not be > b")
	}
}

func TestOptionValue(t *testing.T) {
	o := &Option{Val: true}
	if !o.Value() {
		t.Error("Option.Value() should be true")
	}
	o.Val = false
	if o.Value() {
		t.Error("Option.Value() should be false")
	}
}
