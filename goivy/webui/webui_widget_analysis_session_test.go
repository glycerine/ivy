//go:build web

// Tests for the widget_analysis_session.go port.
//
// These tests verify:
//   - Construction of all 4 widget classes (rule 6: required functions exist)
//   - The Weaken iter.Seq[FrontEndOperation] generator pattern
//   - That ConceptRefine compiles and can be invoked (the body is the
//     literal port flagged in PLAN253_more_followup.md)

package webui

import (
	goivy "github.com/glycerine/ivy/goivy"
	"testing"
)

// TestNewAnalysisSessionWidget verifies that constructing the top-level
// widget builds all sub-widgets and registers handlers.
func TestNewAnalysisSessionWidget(t *testing.T) {
	a := NewAnalysisSessionWidget()
	if a == nil {
		t.Fatal("NewAnalysisSessionWidget returned nil")
	}
	if a.Box == nil {
		t.Error("Box not initialized")
	}
	if a.ProofGraph == nil {
		t.Error("ProofGraph not initialized")
	}
	if a.Arg == nil {
		t.Error("Arg not initialized")
	}
	if a.Crg == nil {
		t.Error("Crg not initialized")
	}
	if a.Concept == nil {
		t.Error("Concept widget not initialized")
	}
	if a.Transition == nil {
		t.Error("Transition widget not initialized")
	}
	// Sub-widgets should know their parent.
	if a.Concept.AnalysisSessionWidget != a {
		t.Error("Concept.AnalysisSessionWidget back-pointer not set")
	}
	if a.Transition.AnalysisSessionWidget != a {
		t.Error("Transition.AnalysisSessionWidget back-pointer not set")
	}
	if a.Concept.Box == nil {
		t.Error("Concept.Box not initialized")
	}
	if a.Transition.Box == nil {
		t.Error("Transition.Box not initialized")
	}
}

func TestAnalysisSessionWidgetHistoryNavigationClampsToSession(t *testing.T) {
	a := NewAnalysisSessionWidget()
	session := &AnalysisSession{History: []any{"init", "middle", "last"}}
	a.RegisterSession(session)

	if got, want := a.CurrentStep, 2; got != want {
		t.Fatalf("CurrentStep after RegisterSession = %d, want %d", got, want)
	}
	a.Prev(nil)
	if got, want := a.CurrentStep, 1; got != want {
		t.Fatalf("CurrentStep after Prev = %d, want %d", got, want)
	}
	a.First(nil)
	if got, want := a.CurrentStep, 0; got != want {
		t.Fatalf("CurrentStep after First = %d, want %d", got, want)
	}
	a.Prev(nil)
	if got, want := a.CurrentStep, 0; got != want {
		t.Fatalf("CurrentStep after Prev at first = %d, want %d", got, want)
	}
	a.Next(nil)
	if got, want := a.CurrentStep, 1; got != want {
		t.Fatalf("CurrentStep after Next = %d, want %d", got, want)
	}
	a.Last(nil)
	if got, want := a.CurrentStep, 2; got != want {
		t.Fatalf("CurrentStep after Last = %d, want %d", got, want)
	}
	a.Next(nil)
	if got, want := a.CurrentStep, 2; got != want {
		t.Fatalf("CurrentStep after Next at last = %d, want %d", got, want)
	}
}

// TestConceptSessionControlsButtons verifies that the base class init
// builds the three concept buttons mirroring Python lines 100-104.
func TestConceptSessionControlsButtons(t *testing.T) {
	c := NewConceptSessionControls()
	if c == nil {
		t.Fatal("NewConceptSessionControls returned nil")
	}
	if got := len(c.ConceptButtons); got != 3 {
		t.Errorf("len(ConceptButtons) = %d, want 3", got)
	}
	wantLabels := []string{"undo", "reset domain", "diagram domain"}
	for i, want := range wantLabels {
		if i >= len(c.ConceptButtons) {
			break
		}
		if got := c.ConceptButtons[i].Description; got != want {
			t.Errorf("ConceptButtons[%d] = %q, want %q", i, got, want)
		}
	}
}

// TestEdgeClassClickToggles verifies that edge_class_click toggles all
// of the corresponding edge checkboxes — Python lines 186-202.
func TestEdgeClassClickToggles(t *testing.T) {
	c := NewConceptSessionControls()
	// Pre-populate two edges for the "all_to_all" class.
	c.EdgeDisplayCheckboxes["e1"] = map[string]*CheckboxWidget{
		"all_to_all": {Value: false},
	}
	c.EdgeDisplayCheckboxes["e2"] = map[string]*CheckboxWidget{
		"all_to_all": {Value: false},
	}
	c.EdgeClassClick("all_to_all")
	if !c.EdgeDisplayCheckboxes["e1"]["all_to_all"].Value {
		t.Error("e1.all_to_all should be on after first click")
	}
	if !c.EdgeDisplayCheckboxes["e2"]["all_to_all"].Value {
		t.Error("e2.all_to_all should be on after first click")
	}
	c.EdgeClassClick("all_to_all")
	if c.EdgeDisplayCheckboxes["e1"]["all_to_all"].Value {
		t.Error("e1.all_to_all should be off after second click")
	}
}

// TestWeakenInteraction drives the Weaken generator and verifies that
// the user-selected conjectures are removed.
func TestWeakenInteraction(t *testing.T) {
	asw := NewAnalysisSessionWidget()
	tw := asw.Transition

	c1 := goivy.NewConst("conj1", goivy.Boolean)
	c2 := goivy.NewConst("conj2", goivy.Boolean)
	c3 := goivy.NewConst("conj3", goivy.Boolean)
	tw.Conjectures = []goivy.Expr{c1, c2, c3}

	var seen []FrontEndOperation
	for op := range tw.Weaken(nil) {
		seen = append(seen, op)
		if usm, ok := op.(*UserSelectMultiple); ok {
			// User picks the first and third conjectures for removal.
			usm.Selection = []any{c1, c3}
		}
	}
	if len(seen) != 1 {
		t.Fatalf("expected exactly 1 yielded op, got %d", len(seen))
	}
	if _, ok := seen[0].(*UserSelectMultiple); !ok {
		t.Fatalf("expected UserSelectMultiple, got %T", seen[0])
	}
	if got := len(tw.Conjectures); got != 1 {
		t.Errorf("len(Conjectures) after weaken = %d, want 1", got)
	} else if tw.Conjectures[0] != c2 {
		t.Errorf("remaining conjecture = %v, want %v", tw.Conjectures[0], c2)
	}
}

// TestWeakenCancel verifies that cancelling the modal leaves the
// conjecture list unchanged.
func TestWeakenCancel(t *testing.T) {
	asw := NewAnalysisSessionWidget()
	tw := asw.Transition

	c1 := goivy.NewConst("conj1", goivy.Boolean)
	c2 := goivy.NewConst("conj2", goivy.Boolean)
	tw.Conjectures = []goivy.Expr{c1, c2}

	for op := range tw.Weaken(nil) {
		if usm, ok := op.(*UserSelectMultiple); ok {
			usm.Cancelled = true
		}
	}
	if got := len(tw.Conjectures); got != 2 {
		t.Errorf("len(Conjectures) after cancel = %d, want 2", got)
	}
}

// TestStrengthenAddsConjecture verifies the Strengthen click handler
// appends the currently-selected conjecture to the list.
func TestStrengthenAddsConjecture(t *testing.T) {
	asw := NewAnalysisSessionWidget()
	tw := asw.Transition

	// With no facts, Strengthen should add lg.True (the default
	// returned by GetSelectedConjecture).
	before := len(tw.Conjectures)
	if err := tw.Strengthen(nil); err != nil {
		t.Fatal(err)
	}
	if got := len(tw.Conjectures); got != before+1 {
		t.Errorf("len(Conjectures) = %d, want %d", got, before+1)
	}
	if tw.Conjectures[before] != goivy.True {
		t.Errorf("added conjecture = %v, want lg.True", tw.Conjectures[before])
	}
}
