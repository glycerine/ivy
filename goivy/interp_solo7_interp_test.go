package goivy

// Solo 7 — Tests for ModuleTypeCheck and ModuleTypeCheckConcepts (§6.3 #31)

import (
	"testing"
)

// ============================================================================
// ModuleTypeCheck
// ============================================================================

// TestModuleTypeCheck_ArityError verifies that an axiom with wrong arity
// is caught by ModuleTypeCheck.
func TestModuleTypeCheck_ArityError(t *testing.T) {
	mod := NewModule()
	sig := NewSig()
	mod.Sig = sig

	// Add symbol "r" with arity 2
	sortA := &UninterpretedSort{Name: "a"}
	sig.AddSort(sortA)
	rSort, _ := NewFunctionSort(sortA, sortA, Boolean)
	sig.AddSymbol("r", rSort)
	rSym, _ := sig.FindSymbol("r", false)

	// Axiom calls r with 1 arg (wrong — needs 2).
	// Construct Apply directly to bypass NewApply's sort check.
	xVar, _ := NewVariable("X", sortA)
	badApp := &Apply{Func: rSym, Terms: []Expr{xVar}}
	mod.LabeledAxioms = append(mod.LabeledAxioms, mod.Cfg.AstCfg.NewLabeledFormula(nil, badApp))

	err := ModuleTypeCheck(mod)
	if err == nil {
		t.Error("expected arity error, got nil")
	}
}

// TestModuleTypeCheck_Valid verifies that valid axioms pass type check.
func TestModuleTypeCheck_Valid(t *testing.T) {
	mod := NewModule()
	sig := NewSig()
	mod.Sig = sig

	// Add symbol "p" with arity 1
	sortA := &UninterpretedSort{Name: "a"}
	sig.AddSort(sortA)
	pSort, _ := NewFunctionSort(sortA, Boolean)
	sig.AddSymbol("p", pSort)
	pSym, _ := sig.FindSymbol("p", false)

	// Axiom calls p with 1 arg (correct)
	xVar, _ := NewVariable("X", sortA)
	goodApp, _ := NewApply(pSym, xVar)
	mod.LabeledAxioms = append(mod.LabeledAxioms, mod.Cfg.AstCfg.NewLabeledFormula(nil, goodApp))

	err := ModuleTypeCheck(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestModuleTypeCheck_EmptyAxioms verifies no error with empty axiom list.
func TestModuleTypeCheck_EmptyAxioms(t *testing.T) {
	mod := NewModule()
	mod.Sig = NewSig()

	err := ModuleTypeCheck(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ============================================================================
// ModuleTypeCheckConcepts
// ============================================================================

// TestModuleTypeCheckConcepts_Empty verifies no error with no concept spaces.
func TestModuleTypeCheckConcepts_Empty(t *testing.T) {
	mod := NewModule()
	mod.Sig = NewSig()

	err := ModuleTypeCheckConcepts(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestModuleTypeCheckConcepts_Valid verifies valid concept spaces pass.
func TestModuleTypeCheckConcepts_Valid(t *testing.T) {
	mod := NewModule()
	sig := NewSig()
	mod.Sig = sig

	// Add a sort and symbol
	sortA := &UninterpretedSort{Name: "a"}
	sig.AddSort(sortA)
	pSort, _ := NewFunctionSort(sortA, Boolean)
	sig.AddSymbol("p", pSort)
	pSym, _ := sig.FindSymbol("p", false)

	// Create a concept space: (p(X), p(X))
	xVar, _ := NewVariable("X", sortA)
	pApp, _ := NewApply(pSym, xVar)

	mod.ConceptSpaces = append(mod.ConceptSpaces, ConceptSpace{Label: pApp, Body: pApp})

	err := ModuleTypeCheckConcepts(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify relations were restored (not modified)
	if mod.Relations.Len() != 0 {
		t.Errorf("expected original relations to be restored, got %d entries", mod.Relations.Len())
	}
}

// TestModuleTypeCheckConcepts_ArityError verifies that a concept space body
// with wrong arity is caught.
func TestModuleTypeCheckConcepts_ArityError(t *testing.T) {
	mod := NewModule()
	sig := NewSig()
	mod.Sig = sig

	// Add symbol "r" with arity 2
	sortA := &UninterpretedSort{Name: "a"}
	sig.AddSort(sortA)
	rSort, _ := NewFunctionSort(sortA, sortA, Boolean)
	sig.AddSymbol("r", rSort)
	rSym, _ := sig.FindSymbol("r", false)

	// Concept space body calls r with 1 arg (wrong).
	// Construct Apply directly to bypass NewApply's sort check.
	xVar, _ := NewVariable("X", sortA)
	badApp := &Apply{Func: rSym, Terms: []Expr{xVar}}
	relSym := NewConst("c", pSort(sortA))

	mod.ConceptSpaces = append(mod.ConceptSpaces, ConceptSpace{Label: relSym, Body: badApp})

	err := ModuleTypeCheckConcepts(mod)
	if err == nil {
		t.Error("expected arity error, got nil")
	}

	// Verify relations were restored even on error
	if mod.Relations.Len() != 0 {
		t.Errorf("expected original relations to be restored after error, got %d entries", mod.Relations.Len())
	}
}

// pSort is a helper to make a unary predicate sort (a -> bool).
func pSort(dom Sort) Sort {
	s, _ := NewFunctionSort(dom, Boolean)
	return s
}

// ============================================================================
// ModuleTypeCheckConcepts restores relations
// ============================================================================

// TestModuleTypeCheckConcepts_RestoresRelations verifies that the original
// relations map is restored after type checking, even when concept space
// relations are temporarily added.
func TestModuleTypeCheckConcepts_RestoresRelations(t *testing.T) {
	mod := NewModule()
	sig := NewSig()
	mod.Sig = sig

	sortA := &UninterpretedSort{Name: "a"}
	sig.AddSort(sortA)

	// Set up an existing relation in mod.Relations
	existingSort, _ := NewFunctionSort(sortA, Boolean)
	mod.Relations.Set("existing", existingSort)

	// Add a concept space with a new relation "c"
	cSort, _ := NewFunctionSort(sortA, Boolean)
	sig.AddSymbol("c", cSort)
	cSym, _ := sig.FindSymbol("c", false)
	xVar, _ := NewVariable("X", sortA)
	cApp, _ := NewApply(cSym, xVar)

	mod.ConceptSpaces = append(mod.ConceptSpaces, ConceptSpace{Label: cApp, Body: cApp})

	err := ModuleTypeCheckConcepts(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that mod.Relations still has exactly the original entry
	if mod.Relations.Len() != 1 {
		t.Errorf("expected 1 relation (original), got %d", mod.Relations.Len())
	}
	if mod.Relations.Get("existing") == nil {
		t.Error("original relation 'existing' was lost")
	}
	if mod.Relations.Get("c") != nil {
		t.Error("temporary relation 'c' leaked into mod.Relations")
	}
}
