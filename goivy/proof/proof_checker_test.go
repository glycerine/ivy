package proof

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// TestAdmitPropositionWithComposeTactics tests that a property admitted with
// an empty ComposeTactics proof is added to the axioms list.
func TestAdmitPropositionWithComposeTactics(t *testing.T) {
	// Create a simple proposition: labeled formula "p" with formula True
	label := proofTestAstCfg.NewAtom("p")
	prop := proofTestAstCfg.NewLabeledFormula(label, lg.True)

	pc := NewProofChecker(nil, nil, nil, nil, nil)
	subgoals, err := pc.AdmitProposition(prop, proofTestAstCfg.NewComposeTactics(nil))
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
	lhs := lg.NewConst("f", s)
	rhs := lg.NewConst("g", s)
	def := lg.NewDefinition(lhs, rhs)
	label := proofTestAstCfg.NewAtom("mydef")
	prop := proofTestAstCfg.NewLabeledFormula(label, def)

	pc := NewProofChecker(nil, nil, nil, nil, nil)
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
	label := proofTestAstCfg.NewAtom("q")
	prop := proofTestAstCfg.NewLabeledFormula(label, lg.True)

	pc := NewProofChecker(nil, nil, nil, nil, nil)
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
	lhs := lg.NewConst("f", s)
	rhs := lg.NewConst("g", s)
	def := lg.NewDefinition(lhs, rhs)
	label := proofTestAstCfg.NewAtom("mydef")
	prop := proofTestAstCfg.NewLabeledFormula(label, def)

	pc := NewProofChecker(nil, nil, nil, nil, nil)
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
	label := proofTestAstCfg.NewAtom("p")
	prop := proofTestAstCfg.NewLabeledFormula(label, lg.True)

	pc := NewProofChecker(nil, nil, nil, nil, nil)
	axiomsBefore := len(pc.Axioms)
	subgoals, err := pc.GetSubgoals(prop, proofTestAstCfg.NewComposeTactics(nil))
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

// TestNewProofCheckerDefinitionKeyUsesDefinesName tests BUG 1:
// Definitions map should be keyed by the defined symbol name (d.formula.defines().name),
// not by the label name. Python: ivy_proof.py:53.
func TestNewProofCheckerDefinitionKeyUsesDefinesName(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "S"}
	// Definition: f = g. The defined symbol is "f", but the label is "mydef".
	lhs := lg.NewConst("f", s)
	rhs := lg.NewConst("g", s)
	def := lg.NewDefinition(lhs, rhs)
	label := proofTestAstCfg.NewAtom("mydef")
	defLF := proofTestAstCfg.NewLabeledFormula(label, def)

	pc := NewProofChecker(nil, nil, nil, []*ast.LabeledFormula{defLF}, nil)

	// Key should be "f" (the defines() name), NOT "mydef" (the label name)
	if _, ok := pc.Definitions["f"]; !ok {
		t.Fatalf("expected Definitions['f'] to exist (keyed by defines().name); keys are: %v", mapKeys(pc.Definitions))
	}
	if _, ok := pc.Definitions["mydef"]; ok {
		t.Fatal("Definitions should NOT be keyed by label name 'mydef'")
	}
}

// TestAdmitPropositionWithExistingSubgoals tests BUG 4:
// AdmitProposition should accept optional pre-computed subgoals parameter.
// Python: ivy_proof.py:98 — def admit_proposition(self, prop, proof=None, subgoals=None)
// TestAdmitPropositionWithExistingSubgoals tests BUG 4:
// AdmitProposition should accept optional pre-computed subgoals parameter.
// Python: ivy_proof.py:98 — def admit_proposition(self, prop, proof=None, subgoals=None)
func TestAdmitPropositionWithExistingSubgoals(t *testing.T) {
	label := proofTestAstCfg.NewAtom("p")
	prop := proofTestAstCfg.NewLabeledFormula(label, lg.True)

	label2 := proofTestAstCfg.NewAtom("sg1")
	sg1 := proofTestAstCfg.NewLabeledFormula(label2, lg.True)

	pc := NewProofChecker(nil, nil, nil, nil, nil)
	subgoals, err := pc.AdmitProposition(prop, proofTestAstCfg.NewComposeTactics(nil), sg1)
	if err != nil {
		t.Fatalf("AdmitProposition with existing subgoals failed: %v", err)
	}
	// When pre-supplied subgoals are given, proof applies to those, not to [prop].
	// With ComposeTactics (empty), should return the pre-supplied subgoals unchanged.
	if len(subgoals) != 1 {
		t.Fatalf("expected 1 subgoal (the pre-supplied one), got %d", len(subgoals))
	}
}

// TestGetSubgoalsRejectsDefinition tests BUG 5:
// GetSubgoals must reject definitions with an assertion error.
// Python: ivy_proof.py:129 — assert not isinstance(prop.formula, il.IvyDefinition)
func TestGetSubgoalsRejectsDefinition(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "S"}
	lhs := lg.NewConst("f", s)
	rhs := lg.NewConst("g", s)
	def := lg.NewDefinition(lhs, rhs)
	label := proofTestAstCfg.NewAtom("mydef")
	prop := proofTestAstCfg.NewLabeledFormula(label, def)

	pc := NewProofChecker(nil, nil, nil, nil, nil)
	_, err := pc.GetSubgoals(prop, proofTestAstCfg.NewComposeTactics(nil))
	if err == nil {
		t.Fatal("expected error when GetSubgoals is called with a Definition")
	}
}

// mapKeys returns the keys of a string map for debug output.
func mapKeys(m map[string]*ast.LabeledFormula) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestSetLastAxiomAndSetSchema tests the interface helper methods.
func TestSetLastAxiomAndSetSchema(t *testing.T) {
	label := proofTestAstCfg.NewAtom("p")
	prop := proofTestAstCfg.NewLabeledFormula(label, lg.True)

	pc := NewProofChecker(nil, nil, nil, nil, nil)
	// Admit a prop first
	pc.AdmitProposition(prop, proofTestAstCfg.NewComposeTactics(nil))
	if len(pc.Axioms) == 0 {
		t.Fatal("should have at least one axiom")
	}

	// SetLastAxiom
	label2 := proofTestAstCfg.NewAtom("q")
	prop2 := proofTestAstCfg.NewLabeledFormula(label2, lg.True)
	pc.SetLastAxiom(prop2)
	last := pc.Axioms[len(pc.Axioms)-1]
	if last.LabelName() != "q" {
		t.Fatalf("expected last axiom label 'q', got '%s'", last.LabelName())
	}

	// SetSchema
	pc.SetSchema("myschema", prop2)
	if got, _ := pc.Schemata.Get2("myschema"); got != prop2 {
		t.Fatal("SetSchema did not update")
	}
}
