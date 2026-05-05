package goivy

import (
	"testing"
)

func TestBackgroundTheoryEmpty(t *testing.T) {
	m := New()
	theory := m.BackgroundTheory(nil)
	if theory == nil {
		t.Fatal("BackgroundTheory should not return nil")
	}
	if len(theory.Fmlas) != 0 {
		t.Errorf("empty module should have 0 theory formulas, got %d", len(theory.Fmlas))
	}
}

func TestUpdateTheorySimple(t *testing.T) {
	m := New()
	// Add a simple axiom: And() (true)
	acfg := m.Cfg.AstCfg
	m.LabeledAxioms = append(m.LabeledAxioms, acfg.NewLabeledFormula(nil, NewConst("axiom", Boolean)))

	m.UpdateTheory()

	theory := m.BackgroundTheory(nil)
	if theory == nil {
		t.Fatal("BackgroundTheory returned nil after UpdateTheory")
	}
	if len(theory.Fmlas) == 0 {
		t.Error("theory should have at least one formula from axiom")
	}
}

func TestUpdateTheoryWithTemporalAxiom(t *testing.T) {
	m := New()
	acfg := m.Cfg.AstCfg
	// A temporal axiom should be excluded from the background theory.
	trueVal := true
	temporalLF := acfg.NewLabeledFormula(nil, NewConst("axiom", Boolean))
	temporalLF.Temporal = &trueVal
	m.LabeledAxioms = append(m.LabeledAxioms, temporalLF)
	// A non-temporal axiom should be included.
	m.LabeledAxioms = append(m.LabeledAxioms, acfg.NewLabeledFormula(nil, NewConst("axiom", Boolean)))

	m.UpdateTheory()
	theory := m.BackgroundTheory(nil)
	// Should have exactly 1 formula (the non-temporal axiom).
	if len(theory.Fmlas) != 1 {
		t.Errorf("expected 1 formula (non-temporal only), got %d", len(theory.Fmlas))
	}
}

func TestAxioms(t *testing.T) {
	m := New()
	f1 := NewConst("axiom", Boolean)
	f2 := &Or{Terms: []Expr{True}}
	falseVal := false
	trueVal2 := true
	m.LabeledAxioms = []*LabeledFormula{
		{Formula: f1, Temporal: &falseVal},
		{Formula: f2, Temporal: &trueVal2},
	}

	axioms := m.Axioms()
	if len(axioms) != 1 {
		t.Errorf("expected 1 non-temporal axiom, got %d", len(axioms))
	}
	if !axioms[0].Equal(f1) {
		t.Error("axiom should be f1")
	}
}

func TestConjs(t *testing.T) {
	m := New()
	f1 := NewConst("axiom", Boolean)
	lf := &LabeledFormula{Formula: f1}
	lf.SetLineno(Location{Line: 10})
	m.LabeledConjs = []*LabeledFormula{lf}

	conjs := m.Conjs()
	if len(conjs) != 1 {
		t.Fatalf("expected 1 conjecture, got %d", len(conjs))
	}
	if conjs[0] == nil {
		t.Fatal("conjecture clause should not be nil")
	}
}

func TestGetAxiomsNoSchemata(t *testing.T) {
	m := New()
	f1 := NewConst("axiom", Boolean)
	m.LabeledAxioms = []*LabeledFormula{
		m.Cfg.AstCfg.NewLabeledFormula(nil, f1),
	}

	axioms := m.GetAxioms()
	if len(axioms) != 1 {
		t.Errorf("expected 1 axiom, got %d", len(axioms))
	}
}

func TestDropLabel(t *testing.T) {
	f := &And{}
	acfg := NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, f)
	result := DropLabel(lf)
	if !result.Equal(f) {
		t.Error("DropLabel should return the formula")
	}

	// Test with a bare node
	result2 := DropLabel(f)
	if !result2.Equal(f) {
		t.Error("DropLabel on Node should return the node itself")
	}

	// Test with nil
	result3 := DropLabel("not a formula")
	if result3 != nil {
		t.Error("DropLabel on unknown type should return nil")
	}
}

func TestVariantAxiomsEmpty(t *testing.T) {
	m := New()
	axioms := m.VariantAxioms()
	if len(axioms) != 0 {
		t.Errorf("expected 0 variant axioms, got %d", len(axioms))
	}
}

func TestVariantAxiomsWithVariants(t *testing.T) {
	m := New()
	parentSort := &UninterpretedSort{Name: "msg"}
	v1 := &UninterpretedSort{Name: "req"}
	v2 := &UninterpretedSort{Name: "resp"}

	m.Sig.AddSort(parentSort)
	m.Sig.AddSort(v1)
	m.Sig.AddSort(v2)
	m.Variants["msg"] = []Sort{v1, v2}

	axioms := m.VariantAxioms()
	if len(axioms) != 1 {
		t.Errorf("expected 1 variant exclusivity axiom, got %d", len(axioms))
	}
}

func TestTheoryContext(t *testing.T) {
	m := New()
	m.LabeledAxioms = []*LabeledFormula{
		{Formula: NewConst("axiom", Boolean)},
	}

	cleanup := m.TheoryContext()
	defer cleanup()

	// After TheoryContext, theory should be set.
	theory := m.BackgroundTheory(nil)
	if theory == nil {
		t.Fatal("BackgroundTheory should not be nil after TheoryContext")
	}
}

func TestUpdateTheoryWithDefinition(t *testing.T) {
	m := New()

	// Create a simple definition: f(X) = X
	xSort := &UninterpretedSort{Name: "t"}
	x, _ := NewVariable("X", xSort)
	fSort, _ := NewFunctionSort(xSort, xSort)
	fSym := NewConst("f", fSort)
	lhs, _ := NewApply(fSym, x)
	def := NewIvyDefinition(lhs, x)

	m.Definitions = []*LabeledFormula{
		{Formula: def},
	}

	m.UpdateTheory()
	theory := m.BackgroundTheory(nil)

	// The definition should end up in the theory's Defs, not Fmlas.
	if len(theory.Defs) != 1 {
		t.Errorf("expected 1 definition in theory, got %d", len(theory.Defs))
	}
}

func TestUpdateTheoryExtensionality(t *testing.T) {
	m := New()

	// Set up a struct with destructors.
	sSort := &UninterpretedSort{Name: "mystruct"}
	tSort := &UninterpretedSort{Name: "t"}
	dSort, _ := NewFunctionSort(sSort, tSort)
	destr := NewConst("myfield", dSort)

	m.Sig.AddSort(sSort)
	m.Sig.AddSort(tSort)
	m.Sig.AddSymbol("myfield", dSort)
	m.SortDestructors["mystruct"] = []*Const{destr}

	m.UpdateTheory()
	theory := m.BackgroundTheory(nil)

	// The extensionality axiom should be in the theory formulas.
	if len(theory.Fmlas) == 0 {
		t.Error("expected extensionality axiom in theory")
	}
}

// Verify that Clauses type is properly used.
func TestConjsReturnType(t *testing.T) {
	m := New()
	m.LabeledConjs = []*LabeledFormula{
		{Formula: &And{}},
	}
	conjs := m.Conjs()
	if len(conjs) != 1 {
		t.Fatal("expected 1 conjecture clause")
	}
	var _ *Clauses = conjs[0] // type assertion check
}
