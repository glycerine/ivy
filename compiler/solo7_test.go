package compiler

// Solo 7 — Tests for the main IvyCompile integration (§6.3 #31):
//   - Version guard for "this" isolate
//   - ModuleTypeCheck called
//   - Action assertion checks (lineno, formal_params)
//   - Concept handler uses GetRelationSortWithTerm

import (
	"testing"

	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// ============================================================================
// Version guard for "this" isolate
// ============================================================================

// TestIvyCompile_VersionGuard_ThisIsolate_V17 verifies that version > 1.6
// creates a default "this" isolate when none exists.
func TestIvyCompile_VersionGuard_ThisIsolate_V17(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()
	mod.Sig = il.NewSig()

	err := IvyCompile(nil, mod, true)
	if err != nil {
		t.Fatalf("IvyCompile: %v", err)
	}

	if _, ok := mod.Isolates["this"]; !ok {
		t.Error("version 1.7: expected 'this' isolate to be created")
	}
}

// TestIvyCompile_VersionGuard_ThisIsolate_V16 verifies that version <= 1.6
// does NOT create a default "this" isolate.
func TestIvyCompile_VersionGuard_ThisIsolate_V16(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.6")

	mod := module.New()
	mod.Sig = il.NewSig()

	err := IvyCompile(nil, mod, true)
	if err != nil {
		t.Fatalf("IvyCompile: %v", err)
	}

	if _, ok := mod.Isolates["this"]; ok {
		t.Error("version 1.6: expected 'this' isolate NOT to be created")
	}
}

// TestIvyCompile_VersionGuard_ExistingIsolate verifies that an existing
// "this" isolate is not overwritten.
func TestIvyCompile_VersionGuard_ExistingIsolate(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()
	cfg := mod.Cfg.AstCfg
	mod.Sig = il.NewSig()
	existing := &ast.IsolateDef{
		Elems:    []ast.Node{cfg.NewAtom("this"), cfg.NewAtom("myobj")},
		WithArgs: 1,
	}
	mod.Isolates["this"] = existing

	err := IvyCompile(nil, mod, true)
	if err != nil {
		t.Fatalf("IvyCompile: %v", err)
	}

	got := mod.Isolates["this"]
	if got != existing {
		t.Error("existing 'this' isolate was overwritten")
	}
}

// ============================================================================
// TypeCheck is called
// ============================================================================

// TestIvyCompile_TypeCheckCalled verifies that IvyCompile calls
// ModuleTypeCheck on the module. We do this by adding an axiom with
// incorrect arity and checking that IvyCompile returns an error.
func TestIvyCompile_TypeCheckCalled(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()
	sig := il.NewSig()
	mod.Sig = sig

	// Add a symbol "r" with arity 2 (two args → bool)
	sortA := &lg.UninterpretedSort{Name: "a"}
	sig.AddSort(sortA)
	rSort, _ := lg.NewFunctionSort(sortA, sortA, lg.Boolean)
	sig.AddSymbol("r", rSort)
	rSym, _ := sig.FindSymbol("r", false)

	// Create an axiom that calls r with wrong arity (1 arg instead of 2).
	// Construct Apply directly to bypass NewApply's sort check.
	xVar, _ := lg.NewVariable("X", sortA)
	badApp := &lg.Apply{Func: rSym, Terms: []lg.Expr{xVar}}
	mod.LabeledAxioms = append(mod.LabeledAxioms, &ast.LabeledFormula{
		Formula: badApp,
	})

	err := IvyCompile(nil, mod, true)
	if err == nil {
		t.Error("expected type check error for wrong-arity axiom, got nil")
	}
}

// TestIvyCompile_TypeCheckValid verifies that valid axioms pass type check.
func TestIvyCompile_TypeCheckValid(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()
	sig := il.NewSig()
	mod.Sig = sig

	// Add a unary relation "p" (one arg → bool)
	sortA := &lg.UninterpretedSort{Name: "a"}
	sig.AddSort(sortA)
	pSort, _ := lg.NewFunctionSort(sortA, lg.Boolean)
	sig.AddSymbol("p", pSort)
	pSym, _ := sig.FindSymbol("p", false)

	// Create an axiom that calls p with correct arity (1 arg)
	xVar, _ := lg.NewVariable("X", sortA)
	goodApp, _ := lg.NewApply(pSym, xVar)
	mod.LabeledAxioms = append(mod.LabeledAxioms, &ast.LabeledFormula{
		Formula: goodApp,
	})

	err := IvyCompile(nil, mod, true)
	if err != nil {
		t.Fatalf("IvyCompile: unexpected error: %v", err)
	}
}

// ============================================================================
// Concept handler uses body for sort inference
// ============================================================================

// TestConcept_UsesTermForSortInference verifies that the Concept handler
// calls GetRelationSortWithTerm (not GetRelationSort), passing the body
// for joint sort inference. We verify that the Concept handler does not
// panic and properly adds the relation symbol when body is available.
func TestConcept_UsesTermForSortInference(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()
	sig := c.Sig

	// Declare sort "t" and symbols
	sortT := &lg.UninterpretedSort{Name: "t"}
	sig.AddSort(sortT)
	fSort, _ := lg.NewFunctionSort(sortT, lg.Boolean)
	sig.AddSymbol("f", fSort)
	// Declare X as a constant of sort t so compilation can resolve it
	sig.AddSymbol("X", sortT)

	// Concept: concept c(X) = f(X)
	conceptLabel := cfg.NewAtom("c")
	conceptLabel.Terms = []ast.Node{cfg.NewAtom("X")}
	body := &ast.App{
		Rep:   cfg.NewAtom("f"),
		Terms: []ast.Node{cfg.NewAtom("X")},
	}
	lf := &ast.LabeledFormula{
		Label:   conceptLabel,
		Formula: body,
	}

	ds := &DomainSetup{Compiler: c}
	err := ds.Concept(lf)
	if err != nil {
		t.Fatalf("Concept: %v", err)
	}

	// Verify the symbol "c" was added with correct sort
	sym, err := sig.FindSymbol("c", false)
	if err != nil {
		t.Fatalf("symbol 'c' not found after Concept: %v", err)
	}
	fs, ok := sym.CSort.(*lg.FunctionSort)
	if !ok {
		t.Fatalf("expected FunctionSort for c, got %T", sym.CSort)
	}
	if len(fs.Domain()) != 1 {
		t.Fatalf("expected 1 domain sort, got %d", len(fs.Domain()))
	}
	domName := fs.Domain()[0].String()
	if domName != "t" {
		t.Errorf("expected domain sort 't', got '%s'", domName)
	}

	// Verify concept space was stored
	if len(c.Module.ConceptSpaces) != 1 {
		t.Fatalf("expected 1 concept space, got %d", len(c.Module.ConceptSpaces))
	}
}
