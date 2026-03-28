package interp

// Solo 7 — Tests for ModuleTypeCheck and ModuleTypeCheckConcepts (§6.3 #31)

import (
	"testing"

	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// ============================================================================
// ModuleTypeCheck
// ============================================================================

// TestModuleTypeCheck_ArityError verifies that an axiom with wrong arity
// is caught by ModuleTypeCheck.
func TestModuleTypeCheck_ArityError(t *testing.T) {
	mod := module.New()
	sig := il.NewSig()
	mod.Sig = sig

	// Add symbol "r" with arity 2
	sortA := &lg.UninterpretedSort{Name: "a"}
	sig.AddSort(sortA)
	rSort, _ := lg.NewFunctionSort(sortA, sortA, lg.Boolean)
	sig.AddSymbol("r", rSort)
	rSym, _ := sig.FindSymbol("r", false)

	// Axiom calls r with 1 arg (wrong — needs 2).
	// Construct Apply directly to bypass NewApply's sort check.
	xVar, _ := lg.NewVariable("X", sortA)
	badApp := &lg.Apply{Func: rSym, Terms: []lg.Expr{xVar}}
	mod.LabeledAxioms = append(mod.LabeledAxioms, mod.Cfg.AstCfg.NewLabeledFormula(nil, badApp))

	err := ModuleTypeCheck(mod)
	if err == nil {
		t.Error("expected arity error, got nil")
	}
}

// TestModuleTypeCheck_Valid verifies that valid axioms pass type check.
func TestModuleTypeCheck_Valid(t *testing.T) {
	mod := module.New()
	sig := il.NewSig()
	mod.Sig = sig

	// Add symbol "p" with arity 1
	sortA := &lg.UninterpretedSort{Name: "a"}
	sig.AddSort(sortA)
	pSort, _ := lg.NewFunctionSort(sortA, lg.Boolean)
	sig.AddSymbol("p", pSort)
	pSym, _ := sig.FindSymbol("p", false)

	// Axiom calls p with 1 arg (correct)
	xVar, _ := lg.NewVariable("X", sortA)
	goodApp, _ := lg.NewApply(pSym, xVar)
	mod.LabeledAxioms = append(mod.LabeledAxioms, mod.Cfg.AstCfg.NewLabeledFormula(nil, goodApp))

	err := ModuleTypeCheck(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestModuleTypeCheck_EmptyAxioms verifies no error with empty axiom list.
func TestModuleTypeCheck_EmptyAxioms(t *testing.T) {
	mod := module.New()
	mod.Sig = il.NewSig()

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
	mod := module.New()
	mod.Sig = il.NewSig()

	err := ModuleTypeCheckConcepts(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestModuleTypeCheckConcepts_Valid verifies valid concept spaces pass.
func TestModuleTypeCheckConcepts_Valid(t *testing.T) {
	mod := module.New()
	sig := il.NewSig()
	mod.Sig = sig

	// Add a sort and symbol
	sortA := &lg.UninterpretedSort{Name: "a"}
	sig.AddSort(sortA)
	pSort, _ := lg.NewFunctionSort(sortA, lg.Boolean)
	sig.AddSymbol("p", pSort)
	pSym, _ := sig.FindSymbol("p", false)

	// Create a concept space: (p(X), p(X))
	xVar, _ := lg.NewVariable("X", sortA)
	pApp, _ := lg.NewApply(pSym, xVar)

	mod.ConceptSpaces = append(mod.ConceptSpaces, module.ConceptSpace{Label: pApp, Body: pApp})

	err := ModuleTypeCheckConcepts(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify relations were restored (not modified)
	if len(mod.Relations) != 0 {
		t.Errorf("expected original relations to be restored, got %d entries", len(mod.Relations))
	}
}

// TestModuleTypeCheckConcepts_ArityError verifies that a concept space body
// with wrong arity is caught.
func TestModuleTypeCheckConcepts_ArityError(t *testing.T) {
	mod := module.New()
	sig := il.NewSig()
	mod.Sig = sig

	// Add symbol "r" with arity 2
	sortA := &lg.UninterpretedSort{Name: "a"}
	sig.AddSort(sortA)
	rSort, _ := lg.NewFunctionSort(sortA, sortA, lg.Boolean)
	sig.AddSymbol("r", rSort)
	rSym, _ := sig.FindSymbol("r", false)

	// Concept space body calls r with 1 arg (wrong).
	// Construct Apply directly to bypass NewApply's sort check.
	xVar, _ := lg.NewVariable("X", sortA)
	badApp := &lg.Apply{Func: rSym, Terms: []lg.Expr{xVar}}
	relSym := lg.NewSymbol("c", pSort(sortA))

	mod.ConceptSpaces = append(mod.ConceptSpaces, module.ConceptSpace{Label: relSym, Body: badApp})

	err := ModuleTypeCheckConcepts(mod)
	if err == nil {
		t.Error("expected arity error, got nil")
	}

	// Verify relations were restored even on error
	if len(mod.Relations) != 0 {
		t.Errorf("expected original relations to be restored after error, got %d entries", len(mod.Relations))
	}
}

// pSort is a helper to make a unary predicate sort (a -> bool).
func pSort(dom lg.Sort) lg.Sort {
	s, _ := lg.NewFunctionSort(dom, lg.Boolean)
	return s
}

// ============================================================================
// ModuleTypeCheckConcepts restores relations
// ============================================================================

// TestModuleTypeCheckConcepts_RestoresRelations verifies that the original
// relations map is restored after type checking, even when concept space
// relations are temporarily added.
func TestModuleTypeCheckConcepts_RestoresRelations(t *testing.T) {
	mod := module.New()
	sig := il.NewSig()
	mod.Sig = sig

	sortA := &lg.UninterpretedSort{Name: "a"}
	sig.AddSort(sortA)

	// Set up an existing relation in mod.Relations
	existingSort, _ := lg.NewFunctionSort(sortA, lg.Boolean)
	mod.Relations["existing"] = existingSort

	// Add a concept space with a new relation "c"
	cSort, _ := lg.NewFunctionSort(sortA, lg.Boolean)
	sig.AddSymbol("c", cSort)
	cSym, _ := sig.FindSymbol("c", false)
	xVar, _ := lg.NewVariable("X", sortA)
	cApp, _ := lg.NewApply(cSym, xVar)

	mod.ConceptSpaces = append(mod.ConceptSpaces, module.ConceptSpace{Label: cApp, Body: cApp})

	err := ModuleTypeCheckConcepts(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that mod.Relations still has exactly the original entry
	if len(mod.Relations) != 1 {
		t.Errorf("expected 1 relation (original), got %d", len(mod.Relations))
	}
	if mod.Relations["existing"] == nil {
		t.Error("original relation 'existing' was lost")
	}
	if mod.Relations["c"] != nil {
		t.Error("temporary relation 'c' leaked into mod.Relations")
	}
}
