package art

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
)

// ---------------------------------------------------------------------------
// AsCyElements tests — verify formula strings (not clause counts)
// ---------------------------------------------------------------------------

// TestAsCyElementsFormulaStrings verifies that state nodes carry actual
// formula strings in long_info, not just clause counts.
func TestAsCyElementsFormulaStrings(t *testing.T) {
	ag := testGraph()

	// Create a state with non-trivial clauses containing a named symbol.
	x := &lg.Const{Name: "x"}
	y := &lg.Const{Name: "y"}
	eq := &lg.Eq{T1: x, T2: y}
	clauses := clauseops.NewClauses([]lg.Expr{eq}, nil, nil)
	s := NewState(ag.Domain, clauses)
	ag.Add(s, nil)

	cy := ag.AsCyElements(nil)
	if cy == nil {
		t.Fatal("AsCyElements returned nil")
	}
	if len(cy.Elements) == 0 {
		t.Fatal("expected at least 1 element")
	}

	// Find the node element
	var longInfo interface{}
	for _, elem := range cy.Elements {
		if elem.Group == "nodes" {
			longInfo = elem.Data["long_info"]
			break
		}
	}

	if longInfo == nil {
		t.Fatal("node should have long_info")
	}

	// long_info should be a []string of formula strings, not a count string
	switch li := longInfo.(type) {
	case []string:
		if len(li) == 0 {
			t.Error("long_info should have at least 1 formula string")
		}
		// The formula should contain "x" and "y" (from our Eq)
		joined := strings.Join(li, " ")
		if !strings.Contains(joined, "x") || !strings.Contains(joined, "y") {
			t.Errorf("long_info should contain formula with x and y, got: %v", li)
		}
	case string:
		// Should NOT be a clause count string
		if strings.Contains(li, "clauses") {
			t.Errorf("long_info should be formula strings, not clause count: %q", li)
		}
	default:
		t.Errorf("unexpected long_info type: %T", longInfo)
	}
}

// TestAsCyElementsBottomState verifies bottom states get the right class.
func TestAsCyElementsBottomState(t *testing.T) {
	ag := testGraph()
	s := NewState(ag.Domain, falseClauses())
	ag.Add(s, nil)

	cy := ag.AsCyElements(nil)
	for _, elem := range cy.Elements {
		if elem.Group == "nodes" {
			if !strings.Contains(elem.Classes, "bottom_state") {
				t.Error("bottom state should have class 'bottom_state'")
			}
			return
		}
	}
	t.Error("no node element found")
}

// TestAsCyElementsNormalState verifies non-bottom states get the 'state' class.
func TestAsCyElementsNormalState(t *testing.T) {
	ag := testGraph()
	s := testState(ag.Domain)
	ag.Add(s, nil)

	cy := ag.AsCyElements(nil)
	for _, elem := range cy.Elements {
		if elem.Group == "nodes" {
			if !strings.Contains(elem.Classes, "state") {
				t.Error("normal state should have class 'state'")
			}
			if strings.Contains(elem.Classes, "bottom_state") {
				t.Error("normal state should NOT have class 'bottom_state'")
			}
			return
		}
	}
	t.Error("no node element found")
}

// TestAsCyElementsTransitionEdge verifies transition edges have correct classes.
func TestAsCyElementsTransitionEdge(t *testing.T) {
	ag := testGraph()
	registerAction(ag, "act")
	pre := testState(ag.Domain)
	ag.Add(pre, nil)
	post := testState(ag.Domain)
	ag.Add(post, NewActionApp("act", pre))

	cy := ag.AsCyElements(nil)
	foundEdge := false
	for _, elem := range cy.Elements {
		if elem.Group == "edges" {
			foundEdge = true
			if !strings.Contains(elem.Classes, "transition_action") {
				t.Errorf("transition edge should have class 'transition_action', got %q", elem.Classes)
			}
		}
	}
	if !foundEdge {
		t.Error("no edge element found")
	}
}

// TestAsCyElementsJoinEdge verifies join edges have correct classes.
func TestAsCyElementsJoinEdge(t *testing.T) {
	ag := testGraph()
	s1 := testState(ag.Domain)
	ag.Add(s1, nil)
	s2 := testState(ag.Domain)
	ag.Add(s2, nil)
	s3 := testState(ag.Domain)
	ag.Add(s3, NewStateJoin(s1, s2))

	cy := ag.AsCyElements(nil)
	foundJoin := false
	for _, elem := range cy.Elements {
		if elem.Group == "edges" && strings.Contains(elem.Classes, "transition_join") {
			foundJoin = true
		}
	}
	if !foundJoin {
		t.Error("join edges should have class 'transition_join'")
	}
}

// TestAsCyElementsCoverEdge verifies covering edges have correct classes.
func TestAsCyElementsCoverEdge(t *testing.T) {
	ag := testGraph()
	states := addStates(ag, 2)
	ag.Cover(states[0], states[1])

	cy := ag.AsCyElements(nil)
	foundCover := false
	for _, elem := range cy.Elements {
		if elem.Group == "edges" && strings.Contains(elem.Classes, "cover") {
			foundCover = true
			// Cover edges should have empty label
			if label, ok := elem.Data["label"].(string); ok && label != "" {
				t.Errorf("cover edge label should be empty, got %q", label)
			}
		}
	}
	if !foundCover {
		t.Error("covering edge should have class 'cover'")
	}
}

// TestAsCyElementsLabelBraceSubstitution verifies that curly braces
// in transition labels are replaced with digraphs.
func TestAsCyElementsLabelBraceSubstitution(t *testing.T) {
	ag := testGraph()
	pre := testState(ag.Domain)
	ag.Add(pre, nil)
	post := testState(ag.Domain)
	// Manually create a transition with braces in the label
	ag.States = append(ag.States, post)
	post.ID = len(ag.States) - 1
	ag.Transitions = append(ag.Transitions, Transition{
		Pre:   pre,
		Op:    nil,
		Label: "if {x} then {y}",
		Post:  post,
	})

	cy := ag.AsCyElements(nil)
	for _, elem := range cy.Elements {
		if elem.Group == "edges" {
			label := elem.Data["label"].(string)
			if strings.Contains(label, "{") || strings.Contains(label, "}") {
				t.Errorf("braces should be replaced in label, got: %q", label)
			}
			if !strings.Contains(label, "-[") || !strings.Contains(label, "]-") {
				t.Errorf("label should contain digraph replacements, got: %q", label)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// RenderRg tests — verify formula strings in text output
// ---------------------------------------------------------------------------

// TestRenderRgFormulaStrings verifies that RenderRg shows actual formulas.
func TestRenderRgFormulaStrings(t *testing.T) {
	ag := testGraph()
	x := &lg.Const{Name: "myvar"}
	y := &lg.Const{Name: "othervar"}
	eq := &lg.Eq{T1: x, T2: y}
	clauses := clauseops.NewClauses([]lg.Expr{eq}, nil, nil)
	s := NewState(ag.Domain, clauses)
	ag.Add(s, nil)

	text := RenderRg(ag)
	if strings.Contains(text, "clauses)") {
		t.Error("RenderRg should not show clause count")
	}
	if !strings.Contains(text, "myvar") || !strings.Contains(text, "othervar") {
		t.Errorf("RenderRg should show formula content, got:\n%s", text)
	}
}

// TestRenderRgBottomState verifies bottom states show "(bottom)".
func TestRenderRgBottomState(t *testing.T) {
	ag := testGraph()
	s := NewState(ag.Domain, falseClauses())
	ag.Add(s, nil)

	text := RenderRg(ag)
	if !strings.Contains(text, "(bottom)") {
		t.Errorf("RenderRg should show '(bottom)' for false states, got:\n%s", text)
	}
}

// TestRenderRgNilGraph verifies nil graph handling.
func TestRenderRgNilGraph(t *testing.T) {
	text := RenderRg(nil)
	if text != "(nil graph)" {
		t.Errorf("expected '(nil graph)', got %q", text)
	}
}
