//go:build web

package webui

import (
	"encoding/json"
	"strings"
	"testing"
)

// --- WebUICyElements tests ---

func TestNewCyElements(t *testing.T) {
	g := NewWebUICyElements()
	if g == nil {
		t.Fatal("nil")
	}
	if len(g.Elements) != 0 {
		t.Errorf("elements not empty")
	}
}

func TestAddNode(t *testing.T) {
	g := NewWebUICyElements()
	g.AddNode("n1", "Node 1", []string{"state"}, "info", "longinfo", nil, "ellipse")
	if len(g.Elements) != 1 {
		t.Fatalf("len = %d", len(g.Elements))
	}
	el := g.Elements[0]
	if el.Group != "nodes" {
		t.Errorf("group = %q", el.Group)
	}
	if el.Data["label"] != "Node 1" {
		t.Errorf("label = %v", el.Data["label"])
	}
	if el.Classes != "state" {
		t.Errorf("classes = %q", el.Classes)
	}
}

func TestAddEdge(t *testing.T) {
	g := NewWebUICyElements()
	g.AddNode("a", "A", nil, "", "", nil, "ellipse")
	g.AddNode("b", "B", nil, "", "", nil, "ellipse")
	g.AddEdge("e1", "a", "b", "edge1", []string{"all_to_all"}, "info", "info")
	if len(g.Elements) != 3 {
		t.Fatalf("len = %d", len(g.Elements))
	}
	edge := g.Elements[2]
	if edge.Group != "edges" {
		t.Errorf("group = %q", edge.Group)
	}
	if edge.Data["source"] != g.NodeID["a"] {
		t.Errorf("source = %v", edge.Data["source"])
	}
	if edge.Data["target"] != g.NodeID["b"] {
		t.Errorf("target = %v", edge.Data["target"])
	}
}

func TestAddMultipleClasses(t *testing.T) {
	g := NewWebUICyElements()
	g.AddNode("x", "X", []string{"state", "bottom_state"}, "", "", nil, "ellipse")
	if g.Elements[0].Classes != "state bottom_state" {
		t.Errorf("classes = %q", g.Elements[0].Classes)
	}
}

func TestCyElementsJSON(t *testing.T) {
	g := NewWebUICyElements()
	g.AddNode("n", "N", []string{"state"}, "si", "li", nil, "ellipse")
	g.AddNode("m", "M", []string{"state"}, "si", "li", nil, "ellipse")
	g.AddEdge("e", "n", "m", "E", []string{"cover"}, "si", "li")

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	elems, ok := parsed["elements"].([]interface{})
	if !ok {
		t.Fatal("elements not array")
	}
	if len(elems) != 3 {
		t.Errorf("len = %d", len(elems))
	}
}

func TestCyElementsJSONNodeFields(t *testing.T) {
	g := NewWebUICyElements()
	g.AddNode("n", "Label", []string{"exactly_one"}, "short", "long", nil, "octagon")

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{`"group":"nodes"`, `"label":"Label"`, `"shape":"octagon"`, `"classes":"exactly_one"`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in %s", want, s)
		}
	}
}

func TestCyElementsJSONEdgeFields(t *testing.T) {
	g := NewWebUICyElements()
	g.AddNode("a", "A", nil, "", "", nil, "ellipse")
	g.AddNode("b", "B", nil, "", "", nil, "ellipse")
	g.AddEdge("rel", "a", "b", "R", []string{"none_to_none"}, "si", "li")

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{`"group":"edges"`, `"source":"n0"`, `"target":"n1"`, `"label":"R"`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in %s", want, s)
		}
	}
}

func TestNodeWidthHeuristic(t *testing.T) {
	g := NewWebUICyElements()
	g.AddNode("short", "Hi", nil, "", "", nil, "ellipse")
	w := g.Elements[0].Data["width"].(int)
	if w < 50 {
		t.Errorf("width = %d, want >= 50", w)
	}

	g2 := NewWebUICyElements()
	g2.AddNode("long", "A very long label indeed", nil, "", "", nil, "ellipse")
	w2 := g2.Elements[0].Data["width"].(int)
	if w2 <= 50 {
		t.Errorf("width for long label = %d, want > 50", w2)
	}
}

// --- RenderWebUIARG tests ---

func TestRenderARGEmpty(t *testing.T) {
	g := RenderWebUIARG(nil)
	if len(g.Elements) != 0 {
		t.Errorf("len = %d", len(g.Elements))
	}
}

func TestRenderARGNodes(t *testing.T) {
	ag := &WebUIAnalysisGraphState{
		States: []WebUIARGNode{
			{ID: 0, Label: "0", IsBottom: false, Info: "state 0"},
			{ID: 1, Label: "1", IsBottom: true, Info: "state 1"},
		},
	}
	g := RenderWebUIARG(ag)
	if len(g.Elements) != 2 {
		t.Fatalf("len = %d", len(g.Elements))
	}
	if g.Elements[0].Classes != "state" {
		t.Errorf("classes[0] = %q", g.Elements[0].Classes)
	}
	if g.Elements[1].Classes != "bottom_state" {
		t.Errorf("classes[1] = %q", g.Elements[1].Classes)
	}
}

func TestRenderARGTransitions(t *testing.T) {
	ag := &WebUIAnalysisGraphState{
		States: []WebUIARGNode{
			{ID: 0, Label: "0"},
			{ID: 1, Label: "1"},
		},
		Transitions: []WebUIARGTransition{
			{SourceID: 0, TargetID: 1, Label: "act", IsJoin: false},
		},
	}
	g := RenderWebUIARG(ag)
	// 2 nodes + 1 edge
	if len(g.Elements) != 3 {
		t.Fatalf("len = %d", len(g.Elements))
	}
	edge := g.Elements[2]
	if edge.Classes != "transition_action" {
		t.Errorf("edge classes = %q", edge.Classes)
	}
}

func TestRenderARGTransitionLabelUnescapesBraceMarkers(t *testing.T) {
	ag := &WebUIAnalysisGraphState{
		States: []WebUIARGNode{
			{ID: 0, Label: "0"},
			{ID: 1, Label: "1"},
		},
		Transitions: []WebUIARGTransition{
			{SourceID: 0, TargetID: 1, Label: "choice -[assume p]-"},
		},
	}
	g := RenderWebUIARG(ag)
	edge := g.Elements[2]
	if got := edge.Data["label"]; got != "choice {assume p}" {
		t.Fatalf("edge label = %q, want %q", got, "choice {assume p}")
	}
	if got := edge.Data["short_info"]; got != "choice {assume p}" {
		t.Fatalf("edge short_info = %q, want %q", got, "choice {assume p}")
	}
}

func TestRenderARGJoinEdge(t *testing.T) {
	ag := &WebUIAnalysisGraphState{
		States: []WebUIARGNode{
			{ID: 0, Label: "0"},
			{ID: 1, Label: "1"},
		},
		Transitions: []WebUIARGTransition{
			{SourceID: 0, TargetID: 1, Label: "join", IsJoin: true},
		},
	}
	g := RenderWebUIARG(ag)
	edge := g.Elements[2]
	if edge.Classes != "transition_join" {
		t.Errorf("edge classes = %q", edge.Classes)
	}
}

func TestRenderARGCovering(t *testing.T) {
	ag := &WebUIAnalysisGraphState{
		States: []WebUIARGNode{
			{ID: 0, Label: "0"},
			{ID: 1, Label: "1"},
		},
		Covering: []WebUIARGCover{
			{CoveredID: 0, CoveringID: 1},
		},
	}
	g := RenderWebUIARG(ag)
	// 2 nodes + 1 cover edge
	if len(g.Elements) != 3 {
		t.Fatalf("len = %d", len(g.Elements))
	}
	edge := g.Elements[2]
	if edge.Classes != "cover" {
		t.Errorf("edge classes = %q", edge.Classes)
	}
}

func TestRenderARGFullJSON(t *testing.T) {
	ag := &WebUIAnalysisGraphState{
		States: []WebUIARGNode{
			{ID: 0, Label: "init"},
			{ID: 1, Label: "step", IsBottom: true},
		},
		Transitions: []WebUIARGTransition{
			{SourceID: 0, TargetID: 1, Label: "act"},
		},
		Covering: []WebUIARGCover{
			{CoveredID: 1, CoveringID: 0},
		},
	}
	g := RenderWebUIARG(ag)
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	// Should be valid JSON with elements array.
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	elems := parsed["elements"].([]interface{})
	if len(elems) != 4 { // 2 nodes + 1 transition + 1 cover
		t.Errorf("len = %d, want 4", len(elems))
	}
}

// --- RenderProofStack tests ---

func TestRenderProofStackNil(t *testing.T) {
	g := RenderProofStack(nil)
	if len(g.Elements) != 0 {
		t.Error("expected empty")
	}
}

func TestRenderProofStackGoals(t *testing.T) {
	stack := &WebUIProofStack{
		Goals: []WebUIProofGoal{
			{ID: 0, Label: "g0", Refuted: false, Info: "goal 0", ParentID: -1},
			{ID: 1, Label: "g1", Refuted: true, Info: "goal 1", ParentID: 0},
		},
	}
	g := RenderProofStack(stack)
	// 2 nodes + 1 edge (goal 1 -> goal 0)
	if len(g.Elements) != 3 {
		t.Fatalf("len = %d", len(g.Elements))
	}
	if !strings.Contains(g.Elements[1].Classes, "refuted") {
		t.Errorf("goal 1 classes = %q, want refuted", g.Elements[1].Classes)
	}
}

// --- RenderConceptGraph tests ---

func TestRenderConceptGraphNil(t *testing.T) {
	g := RenderConceptGraph(nil, nil)
	if len(g.Elements) != 0 {
		t.Error("expected empty")
	}
}

func TestRenderConceptGraphNodes(t *testing.T) {
	cs := NewConceptSession()
	cs.Domain.Concepts["Node"] = &Concept{Name: "Node", Formula: "x:Node", Sorts: []string{"Node"}, Arity: 1}
	cs.Domain.Concepts["ID"] = &Concept{Name: "ID", Formula: "x:ID", Sorts: []string{"ID"}, Arity: 1}
	// Only sort nodes listed in Domain.Nodes become graph nodes (matching Python)
	cs.Domain.Nodes = []string{"Node", "ID"}

	g := RenderConceptGraph(cs, nil)
	if len(g.Elements) < 2 {
		t.Errorf("expected at least 2 elements, got %d", len(g.Elements))
	}
}

func TestRenderConceptGraphEdgeVisibility(t *testing.T) {
	cs := NewConceptSession()
	// Sort nodes must be in Domain.Nodes to appear as graph nodes
	cs.Domain.Concepts["A"] = &Concept{Name: "A", Formula: "a", Sorts: []string{"A"}, Arity: 1}
	cs.Domain.Concepts["B"] = &Concept{Name: "B", Formula: "b", Sorts: []string{"B"}, Arity: 1}
	cs.Domain.Nodes = []string{"A", "B"}
	// Combiners with source/target matching node names
	cs.Domain.Combiners = []*ConceptCombiner{
		{Label: "rel", Source: "A", Target: "B", Formula: "rel(a,b)"},
	}

	// With nil display: edge visible.
	g := RenderConceptGraph(cs, nil)
	edgeCount := 0
	for _, el := range g.Elements {
		if el.Group == "edges" {
			edgeCount++
		}
	}
	if edgeCount != 1 {
		t.Errorf("edge count = %d, want 1", edgeCount)
	}

	// With display hiding the edge class.
	d := NewDisplayCheckboxes()
	d.SetEdgeCheckbox("rel", "edge_unknown", false)
	g2 := RenderConceptGraph(cs, d)
	edgeCount2 := 0
	for _, el := range g2.Elements {
		if el.Group == "edges" {
			edgeCount2++
		}
	}
	if edgeCount2 != 0 {
		t.Errorf("edge count = %d, want 0 (hidden)", edgeCount2)
	}
}

func TestRenderConceptGraphUsesTransitiveOrderingAndReduction(t *testing.T) {
	cs := NewConceptSession()
	cs.Domain.Nodes = []string{"C", "B", "A"}
	for _, name := range []string{"A", "B", "C"} {
		cs.Domain.Concepts[name] = &Concept{Name: name, Formula: name, Sorts: []string{name}, Arity: 1}
	}
	cs.Domain.Combiners = []*ConceptCombiner{
		{Label: "le", Source: "A", Target: "B"},
		{Label: "le", Source: "B", Target: "C"},
		{Label: "le", Source: "A", Target: "C"},
		{Label: "le", Source: "A", Target: "A"},
	}
	cs.AbstractValue = map[string]bool{
		"edge_info|all_to_all|le|A|B": true,
		"edge_info|all_to_all|le|B|C": true,
		"edge_info|all_to_all|le|A|C": true,
		"edge_info|all_to_all|le|A|A": true,
	}
	checks := NewDisplayCheckboxes()
	checks.SetEdgeCheckbox("le", EdgeDisplayAllToAll, true)
	checks.SetEdgeCheckbox("le", EdgeDisplayTransitive, true)

	cy := RenderConceptGraph(cs, checks)
	var nodeObjs []string
	var edgePairs []string
	for _, el := range cy.Elements {
		switch el.Group {
		case "nodes":
			nodeObjs = append(nodeObjs, el.Data["obj"].(string))
			if el.Data["cluster"] != el.Data["obj"] {
				t.Fatalf("node %s cluster = %v, want same sort cluster", el.Data["obj"], el.Data["cluster"])
			}
		case "edges":
			edgePairs = append(edgePairs, el.Data["source_obj"].(string)+"->"+el.Data["target_obj"].(string))
		}
	}
	if strings.Join(nodeObjs, ",") != "A,B,C" {
		t.Fatalf("node order = %v, want A,B,C", nodeObjs)
	}
	if strings.Join(edgePairs, ",") != "A->B,B->C" {
		t.Fatalf("edge pairs = %v, want only transitive reduction cover edges", edgePairs)
	}
}

func TestConceptShapeOctagon(t *testing.T) {
	// Python ivy_graph.py get_shape always returns 'octagon'
	if s := conceptShape("Node"); s != "octagon" {
		t.Errorf("shape = %q, want octagon", s)
	}
	if s := conceptShape("__ID!foo"); s != "octagon" {
		t.Errorf("shape = %q, want octagon", s)
	}
}

// --- Style tests ---

func TestConceptStyleJSON(t *testing.T) {
	data, err := json.Marshal(ConceptStyle())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "non_existing") {
		t.Error("missing non_existing selector")
	}
}

func TestARGStyleJSON(t *testing.T) {
	data, err := json.Marshal(ARGStyle())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "bottom_state") {
		t.Error("missing bottom_state selector")
	}
}

func TestProofStyleJSON(t *testing.T) {
	data, err := json.Marshal(ProofStyle())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "refuted") {
		t.Error("missing refuted selector")
	}
}

// --- DisplayCheckboxes tests ---

func TestDisplayCheckboxesDefault(t *testing.T) {
	d := NewDisplayCheckboxes()
	// Unset edge defaults to false (not visible) in graph_model.go
	if d.EdgeVisible("anything", "all_to_all") {
		t.Error("unregistered edge should default to not visible")
	}
	if d.NodeLabelVisible("anything", "node_necessarily") {
		t.Error("unregistered label should default to not visible")
	}
}

func TestDisplayCheckboxesSetEdge(t *testing.T) {
	d := NewDisplayCheckboxes()
	d.SetEdgeCheckbox("rel", "none_to_none", false)
	if d.EdgeVisible("rel", "none_to_none") {
		t.Error("should be hidden")
	}
	d.SetEdgeCheckbox("rel", "all_to_all", true)
	if !d.EdgeVisible("rel", "all_to_all") {
		t.Error("should be visible after set")
	}
}

func TestDisplayCheckboxesSetNodeLabel(t *testing.T) {
	d := NewDisplayCheckboxes()
	d.SetNodeLabelCheckbox("is_leader", "node_maybe", false)
	if d.NodeLabelVisible("is_leader", "node_maybe") {
		t.Error("should be hidden")
	}
	d.SetNodeLabelCheckbox("is_leader", "node_necessarily", true)
	if !d.NodeLabelVisible("is_leader", "node_necessarily") {
		t.Error("should be visible after set")
	}
}

func TestDisplayCheckboxesNilReceiver(t *testing.T) {
	// A nil DisplayCheckboxes should cause a panic since it has
	// mutex fields.  Just test that NewDisplayCheckboxes works.
	d := NewDisplayCheckboxes()
	if d == nil {
		t.Error("should not be nil")
	}
}

// --- Concept session tests ---

func TestConceptSessionSplit(t *testing.T) {
	cs := NewConceptSession()
	cs.Domain.Concepts["A"] = &Concept{Name: "A", Formula: "a"}
	cs.Domain.Concepts["B"] = &Concept{Name: "B", Formula: "b"}

	if err := cs.Split("A", "B"); err != nil {
		t.Fatal(err)
	}
	if _, ok := cs.Domain.Concepts["A"]; ok {
		t.Error("A should be removed after split")
	}
	if _, ok := cs.Domain.Concepts["A+B"]; !ok {
		t.Error("A+B should exist after split")
	}
	if _, ok := cs.Domain.Concepts["A-B"]; !ok {
		t.Error("A-B should exist after split")
	}
}

func TestConceptSessionUndoSplit(t *testing.T) {
	cs := NewConceptSession()
	cs.Domain.Concepts["A"] = &Concept{Name: "A", Formula: "a"}
	cs.Domain.Concepts["B"] = &Concept{Name: "B", Formula: "b"}

	if err := cs.Split("A", "B"); err != nil {
		t.Fatal(err)
	}
	if err := cs.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, ok := cs.Domain.Concepts["A"]; !ok {
		t.Error("A should be restored after undo")
	}
}

func TestConceptSessionRemove(t *testing.T) {
	cs := NewConceptSession()
	cs.Domain.Concepts["A"] = &Concept{Name: "A", Formula: "a"}
	if err := cs.RemoveConcept("A"); err != nil {
		t.Fatal(err)
	}
	if _, ok := cs.Domain.Concepts["A"]; ok {
		t.Error("A should be removed")
	}
}

func TestConceptSessionSupposeEmpty(t *testing.T) {
	cs := NewConceptSession()
	cs.Domain.Concepts["A"] = &Concept{Name: "A", Formula: "a"}
	if err := cs.SupposeEmpty("A"); err != nil {
		t.Fatal(err)
	}
	if !cs.AbstractValue["node_info|none|A"] {
		t.Error("should be marked as none")
	}
}

// --- Fuzz test ---

func FuzzCyElementsJSON(f *testing.F) {
	f.Add("node1", "label1", "state", "ellipse")
	f.Add("", "", "", "")
	f.Add("a b c", "line1\nline2", "exactly_one at_least_one", "octagon")

	f.Fuzz(func(t *testing.T, obj, label, classes, shape string) {
		g := NewWebUICyElements()
		g.AddNode(obj, label, strings.Fields(classes), "si", "li", nil, shape)

		data, err := json.Marshal(g)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		// Must produce valid JSON.
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		elems, ok := parsed["elements"].([]interface{})
		if !ok || len(elems) != 1 {
			t.Fatal("expected 1 element")
		}
	})
}
