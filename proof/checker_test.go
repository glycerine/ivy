package proof

import (
	"testing"

	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
)

// TestAdmitPropositionWithComposeTactics tests that a property admitted with
// an empty ComposeTactics proof is added to the axioms list.
func TestAdmitPropositionWithComposeTactics(t *testing.T) {
	// Create a simple proposition: labeled formula "p" with formula True
	label := ast.NewAtom("p")
	prop := ast.NewLabeledFormula(label, lg.True)

	pc := NewProofChecker(nil, nil, nil)
	subgoals, err := pc.AdmitProposition(prop, &ast.ComposeTactics{})
	if err != nil {
		t.Fatalf("AdmitProposition with ComposeTactics failed: %v", err)
	}
	// ComposeTactics with no tactics applied to [prop] returns [prop] (no subgoals eliminated)
	// The prop should be admitted to axioms regardless
	if len(pc.Axioms) == 0 {
		t.Fatal("expected prop to be admitted to axioms")
	}
	// Subgoals should be the result of applying empty compose to [prop]
	// With empty tactics list, compose returns the input goals
	if len(subgoals) != 1 {
		t.Fatalf("expected 1 subgoal, got %d", len(subgoals))
	}
}

// TestAdmitPropositionDefinitionDelegates tests that admitting a definition
// delegates to AdmitDefinition.
func TestAdmitPropositionDefinitionDelegates(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "S"}
	lhs := lg.NewSymbol("f", s)
	rhs := lg.NewSymbol("g", s)
	def := lg.NewDefinition(lhs, rhs)
	label := ast.NewAtom("mydef")
	prop := ast.NewLabeledFormula(label, def)

	pc := NewProofChecker(nil, nil, nil)
	subgoals, err := pc.AdmitProposition(prop, nil)
	if err != nil {
		t.Fatalf("AdmitProposition for definition failed: %v", err)
	}
	if len(subgoals) != 0 {
		t.Fatalf("expected 0 subgoals for non-recursive definition, got %d", len(subgoals))
	}
	if _, ok := pc.Definitions["f"]; !ok {
		t.Fatal("expected definition 'f' to be registered")
	}
}

// TestAdmitPropositionNilProof tests that a nil proof raises NoMatch.
func TestAdmitPropositionNilProof(t *testing.T) {
	label := ast.NewAtom("q")
	prop := ast.NewLabeledFormula(label, lg.True)

	pc := NewProofChecker(nil, nil, nil)
	_, err := pc.AdmitProposition(prop, nil)
	if err == nil {
		t.Fatal("expected NoMatch error for nil proof")
	}
	if _, ok := err.(*NoMatch); !ok {
		t.Fatalf("expected *NoMatch, got %T: %v", err, err)
	}
}

// TestAdmitDefinitionRedefinition tests that redefining a symbol raises Redefinition.
func TestAdmitDefinitionRedefinition(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "S"}
	lhs := lg.NewSymbol("f", s)
	rhs := lg.NewSymbol("g", s)
	def := lg.NewDefinition(lhs, rhs)
	label := ast.NewAtom("mydef")
	prop := ast.NewLabeledFormula(label, def)

	pc := NewProofChecker(nil, nil, nil)
	// First admission succeeds
	_, err := pc.AdmitDefinition(prop, nil)
	if err != nil {
		t.Fatalf("first AdmitDefinition failed: %v", err)
	}
	// Second admission should fail with Redefinition
	_, err = pc.AdmitDefinition(prop, nil)
	if err == nil {
		t.Fatal("expected Redefinition error")
	}
	if _, ok := err.(*Redefinition); !ok {
		t.Fatalf("expected *Redefinition, got %T: %v", err, err)
	}
}

// TestGetSubgoals tests that GetSubgoals returns subgoals without admitting.
func TestGetSubgoals(t *testing.T) {
	label := ast.NewAtom("p")
	prop := ast.NewLabeledFormula(label, lg.True)

	pc := NewProofChecker(nil, nil, nil)
	axiomsBefore := len(pc.Axioms)
	subgoals, err := pc.GetSubgoals(prop, &ast.ComposeTactics{})
	if err != nil {
		t.Fatalf("GetSubgoals failed: %v", err)
	}
	// Should not have added to axioms
	if len(pc.Axioms) != axiomsBefore {
		t.Fatal("GetSubgoals should not admit to axioms")
	}
	if len(subgoals) != 1 {
		t.Fatalf("expected 1 subgoal, got %d", len(subgoals))
	}
}

// TestSetLastAxiomAndSetSchema tests the interface helper methods.
func TestSetLastAxiomAndSetSchema(t *testing.T) {
	label := ast.NewAtom("p")
	prop := ast.NewLabeledFormula(label, lg.True)

	pc := NewProofChecker(nil, nil, nil)
	// Admit a prop first
	pc.AdmitProposition(prop, &ast.ComposeTactics{})
	if len(pc.Axioms) == 0 {
		t.Fatal("should have at least one axiom")
	}

	// SetLastAxiom
	label2 := ast.NewAtom("q")
	prop2 := ast.NewLabeledFormula(label2, lg.True)
	pc.SetLastAxiom(prop2)
	last := pc.Axioms[len(pc.Axioms)-1]
	if last.LabelName() != "q" {
		t.Fatalf("expected last axiom label 'q', got '%s'", last.LabelName())
	}

	// SetSchema
	pc.SetSchema("myschema", prop2)
	if pc.Schemata["myschema"] != prop2 {
		t.Fatal("SetSchema did not update")
	}
}
