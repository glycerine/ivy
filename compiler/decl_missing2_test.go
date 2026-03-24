package compiler

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// --- RED tests for missing declaration handlers: items 10-14 from AUDIT18MARCH.md §6.1 ---
// These tests document expected behavior from Python's Ivy compiler.
// They should all FAIL until the corresponding handlers are implemented.

// ============================================================
// Item 10: ARGSetup.State — State declarations not handled
// Python: IvyARGSetup.state (ivy_compiler.py:1420-1421)
//   def state(self,a):
//       self.mod.predicates[a.args[0].relname] = a.args[1]
// ============================================================

// TestARGSetupState checks that state declarations populate mod.Predicates.
func TestMiss2_ARGSetupState(t *testing.T) {
	c := newTestCompiler()
	as := NewARGSetup(c)

	// state link = true
	// Python: a.args[0].relname -> "link", a.args[1] -> the body
	lhs := ast.NewAtom("link")
	rhs := ast.NewAtom("true")
	def := ast.NewDefinition(lhs, rhs)
	lf := ast.NewLabeledFormula(nil, def)
	decl := ast.NewStateDecl(lf)

	err := as.ProcessDecls([]ast.Node{decl})
	if err != nil {
		t.Fatalf("ProcessDecls(StateDecl): %v", err)
	}

	pred, ok := c.Module.Predicates["link"]
	if !ok {
		t.Fatal("expected Predicates to have entry 'link'")
	}
	if pred == nil {
		t.Fatal("expected Predicates['link'] to be non-nil")
	}
}

// TestARGSetupStateMultiple checks that multiple state decls each populate Predicates.
func TestMiss2_ARGSetupStateMultiple(t *testing.T) {
	c := newTestCompiler()
	as := NewARGSetup(c)

	// state link = true
	lf1 := ast.NewLabeledFormula(nil, ast.NewDefinition(ast.NewAtom("link"), ast.NewAtom("true")))
	decl1 := ast.NewStateDecl(lf1)

	// state conn = false
	lf2 := ast.NewLabeledFormula(nil, ast.NewDefinition(ast.NewAtom("conn"), ast.NewAtom("false")))
	decl2 := ast.NewStateDecl(lf2)

	err := as.ProcessDecls([]ast.Node{decl1, decl2})
	if err != nil {
		t.Fatalf("ProcessDecls(StateDecl x2): %v", err)
	}

	if _, ok := c.Module.Predicates["link"]; !ok {
		t.Error("expected Predicates to have entry 'link'")
	}
	if _, ok := c.Module.Predicates["conn"]; !ok {
		t.Error("expected Predicates to have entry 'conn'")
	}
}

// ============================================================
// Item 11: add_definition variable-duplication checks
// Python: ivy_compiler.py:1143-1153
//   Checks: (1) no duplicate variables on LHS
//           (2) all RHS free variables are bound on LHS
// ============================================================

// TestDefinitionDuplicateLHSVariable checks that a definition with duplicate
// LHS variables raises an error.
// Python: "Variable {} occurs twice on left-hand side of definition"
func TestMiss2_DefinitionDuplicateLHSVariable(t *testing.T) {
	c := newTestCompiler()
	d := NewDomainSetup(c)

	// Add a sort and symbols so compilation can proceed to the validation step
	boolSort := &lg.UninterpretedSort{Name: "bool"}
	c.Sig.Sorts["bool"] = boolSort

	// definition f(X, X) = body  — X appears twice on LHS
	x1 := ast.NewVariable("X", "bool")
	x2 := ast.NewVariable("X", "bool")
	lhs := ast.NewAtom("f", x1, x2)
	rhs := ast.NewAtom("true")
	def := ast.NewDefinition(lhs, rhs)
	lf := ast.NewLabeledFormula(nil, def)
	decl := ast.NewDefinitionDecl(lf)

	err := d.ProcessDecl(decl)
	if err == nil {
		t.Fatal("expected error for duplicate LHS variable, got nil")
	}
	if !strings.Contains(err.Error(), "occurs twice on left-hand side") {
		t.Errorf("expected error about 'occurs twice on left-hand side', got: %v", err)
	}
}

// TestDefinitionFreeRHSVariable checks that a definition with an unbound RHS
// variable raises an error.
// Python: "Variable {} occurs free on right-hand side of definition"
func TestMiss2_DefinitionFreeRHSVariable(t *testing.T) {
	c := newTestCompiler()
	d := NewDomainSetup(c)

	boolSort := &lg.UninterpretedSort{Name: "bool"}
	c.Sig.Sorts["bool"] = boolSort

	// definition f(X) = g(X, Y)  — Y is free on RHS but not on LHS
	x := ast.NewVariable("X", "bool")
	lhs := ast.NewAtom("f", x)
	xRef := ast.NewVariable("X", "")
	yFree := ast.NewVariable("Y", "")
	rhs := ast.NewAtom("g", xRef, yFree)
	def := ast.NewDefinition(lhs, rhs)
	lf := ast.NewLabeledFormula(nil, def)
	decl := ast.NewDefinitionDecl(lf)

	err := d.ProcessDecl(decl)
	if err == nil {
		t.Fatal("expected error for free RHS variable, got nil")
	}
	if !strings.Contains(err.Error(), "occurs free on right-hand side") {
		t.Errorf("expected error about 'occurs free on right-hand side', got: %v", err)
	}
}

// ============================================================
// Item 12: DerivedUpdate creation
// Python: derived() (line 1167) and definition() (line 1176) both append
//         DerivedUpdate(df) to self.domain.updates
// ============================================================

// TestDerivedDeclCreatesDerivedUpdate checks that a DerivedDecl causes a
// DerivedUpdate to be appended to mod.Updates.
func TestMiss2_DerivedDeclCreatesDerivedUpdate(t *testing.T) {
	c := newTestCompiler()
	d := NewDomainSetup(c)

	boolSort := &lg.UninterpretedSort{Name: "bool"}
	c.Sig.Sorts["bool"] = boolSort

	// derived foo(X:bool) = true
	x := ast.NewVariable("X", "bool")
	lhs := ast.NewAtom("foo", x)
	rhs := ast.NewAtom("true")
	def := ast.NewDefinition(lhs, rhs)
	lf := ast.NewLabeledFormula(nil, def)
	decl := ast.NewDerivedDecl(lf)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(DerivedDecl): %v", err)
	}

	if len(c.Module.Updates) == 0 {
		t.Fatal("expected Updates to have at least 1 entry after DerivedDecl, got 0")
	}

	// Type-assert last entry to *actions.DerivedUpdate
	last := c.Module.Updates[len(c.Module.Updates)-1]
	if _, ok := last.(*module.DerivedUpdate); !ok {
		t.Errorf("expected last Update to be *actions.DerivedUpdate, got %T", last)
	}
}

// TestDefinitionDeclCreatesDerivedUpdate checks that a DefinitionDecl also causes a
// DerivedUpdate to be appended to mod.Updates.
func TestMiss2_DefinitionDeclCreatesDerivedUpdate(t *testing.T) {
	c := newTestCompiler()
	d := NewDomainSetup(c)

	boolSort := &lg.UninterpretedSort{Name: "bool"}
	c.Sig.Sorts["bool"] = boolSort

	// definition bar(X:bool) = true
	x := ast.NewVariable("X", "bool")
	lhs := ast.NewAtom("bar", x)
	rhs := ast.NewAtom("true")
	def := ast.NewDefinition(lhs, rhs)
	lf := ast.NewLabeledFormula(nil, def)
	decl := ast.NewDefinitionDecl(lf)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(DefinitionDecl): %v", err)
	}

	if len(c.Module.Updates) == 0 {
		t.Fatal("expected Updates to have at least 1 entry after DefinitionDecl, got 0")
	}

	last := c.Module.Updates[len(c.Module.Updates)-1]
	if _, ok := last.(*module.DerivedUpdate); !ok {
		t.Errorf("expected last Update to be *actions.DerivedUpdate, got %T", last)
	}
}

// ============================================================
// Item 13: opt_mutax — Mutually exclusive axiom checking
// Python: ivy_compiler.py:1738-1759
//   Post-compilation check: if opt_mutax is false (default), verify that
//   no axiom's symbol dependencies are modified by actions.
// ============================================================

// TestCheckMutaxRejectsAxiomSymbolAssignment checks that CheckMutax returns
// an error when an axiom references a symbol that is modified by an action.
func TestMiss2_CheckMutaxRejectsAxiomSymbolAssignment(t *testing.T) {
	c := newTestCompiler()

	boolSort := &lg.UninterpretedSort{Name: "bool"}
	c.Sig.Sorts["bool"] = boolSort
	c.Sig.AddSymbol("s", boolSort)

	// Add an axiom that references symbol "s" — use compiled lg.Symbol so structural keys match
	sSym := &lg.Symbol{Name: "s", CSort: boolSort}
	axiomLf := ast.NewLabeledFormula(nil, sSym)
	c.Module.LabeledAxioms = append(c.Module.LabeledAxioms, axiomLf)

	// Add an action that modifies "s"
	assignAction := actions.NewAssignAction(
		&lg.Symbol{Name: "s", CSort: boolSort},
		&lg.Symbol{Name: "true"},
	)
	c.Module.Actions["test_action"] = assignAction

	// mutax disabled (false) → should error
	err := CheckMutax(c.Module, false)
	if err == nil {
		t.Fatal("expected error about immutable symbol assigned, got nil")
	}
	if !strings.Contains(err.Error(), "immutable symbol assigned") {
		t.Errorf("expected error about 'immutable symbol assigned', got: %v", err)
	}
}

// TestCheckMutaxAllowsWhenEnabled checks that CheckMutax returns no error
// when mutax is enabled (true), even if axiom symbols are modified.
func TestMiss2_CheckMutaxAllowsWhenEnabled(t *testing.T) {
	c := newTestCompiler()

	boolSort := &lg.UninterpretedSort{Name: "bool"}
	c.Sig.Sorts["bool"] = boolSort
	c.Sig.AddSymbol("s", boolSort)

	axiomAtom := ast.NewAtom("s")
	axiomLf := ast.NewLabeledFormula(nil, axiomAtom)
	c.Module.LabeledAxioms = append(c.Module.LabeledAxioms, axiomLf)

	assignAction := actions.NewAssignAction(
		&lg.Symbol{Name: "s", CSort: boolSort},
		&lg.Symbol{Name: "true"},
	)
	c.Module.Actions["test_action"] = assignAction

	// mutax enabled (true) → should NOT error
	err := CheckMutax(c.Module, true)
	if err != nil {
		t.Fatalf("expected no error when mutax enabled, got: %v", err)
	}
}

// TestCheckMutaxDefinitionLHS checks that CheckMutax catches when a
// definition's LHS symbol is modified by an action.
func TestMiss2_CheckMutaxDefinitionLHS(t *testing.T) {
	c := newTestCompiler()

	boolSort := &lg.UninterpretedSort{Name: "bool"}
	c.Sig.Sorts["bool"] = boolSort
	c.Sig.AddSymbol("f", boolSort)

	// Add a compiled definition for "f" — use lg.Definition with lg.Symbol
	// to match what the real compiler produces (Python: mod.definitions has compiled defs)
	fSym := &lg.Symbol{Name: "f", CSort: boolSort}
	falseSym := &lg.Symbol{Name: "false"}
	logicDef := lg.NewDefinition(fSym, falseSym)
	defLf := ast.NewLabeledFormula(nil, logicDef)
	c.Module.Definitions = append(c.Module.Definitions, defLf)

	// Add an action that modifies "f"
	assignAction := actions.NewAssignAction(
		&lg.Symbol{Name: "f", CSort: boolSort},
		&lg.Symbol{Name: "val"},
	)
	c.Module.Actions["test_action"] = assignAction

	err := CheckMutax(c.Module, false)
	if err == nil {
		t.Fatal("expected error about immutable symbol assigned for definition LHS, got nil")
	}
	if !strings.Contains(err.Error(), "immutable symbol assigned") {
		t.Errorf("expected error about 'immutable symbol assigned', got: %v", err)
	}
}

// ============================================================
// Item 14: compile_theory in interpret
// Python: interpret() calls compile_theory() at 3 points:
//   1. Native int type → compile_theory(self.domain, lhs, 'int')
//   2. Range → compile_theory(self.domain, lhs, interp[lhs])
//   3. Solver sort string → compile_theory(self.domain, lhs, rhs)
// ============================================================

// TestInterpretNativeIntCallsCompileTheory checks that interpreting a sort as
// native int calls CompileTheory, which should add ordering axioms.
func TestMiss2_InterpretNativeIntCallsCompileTheory(t *testing.T) {
	c := newTestCompiler()
	d := NewDomainSetup(c)

	intSort := &lg.UninterpretedSort{Name: "myint"}
	c.Sig.Sorts["myint"] = intSort

	// interpret myint = <<<int>>>  (NativeType with code "int")
	lhs := ast.NewSymbol("myint", nil)
	rhs := &ast.NativeType{Elems: []ast.Node{ast.NewAtom("int")}}
	def := ast.NewDefinition(lhs, rhs)
	lf := ast.NewLabeledFormula(nil, def)
	decl := ast.NewInterpretDecl(lf)

	beforeSchemata := len(c.Module.Schemata)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(InterpretDecl native int): %v", err)
	}

	afterSchemata := len(c.Module.Schemata)
	if afterSchemata <= beforeSchemata {
		t.Fatalf("expected Schemata to grow after native int interpretation (int theory adds rec/ind/lep schemata), before=%d after=%d",
			beforeSchemata, afterSchemata)
	}
}

// TestInterpretRangeCallsCompileTheory checks that interpreting a sort as a
// range calls CompileTheory, which should add range theory axioms.
func TestMiss2_InterpretRangeCallsCompileTheory(t *testing.T) {
	c := newTestCompiler()
	d := NewDomainSetup(c)

	myintSort := &lg.UninterpretedSort{Name: "myint"}
	c.Sig.Sorts["myint"] = myintSort

	// interpret myint = 0..100
	lhs := ast.NewSymbol("myint", nil)
	rhs := ast.NewRange(ast.NewAtom("0"), ast.NewAtom("100"))
	def := ast.NewDefinition(lhs, rhs)
	lf := ast.NewLabeledFormula(nil, def)
	decl := ast.NewInterpretDecl(lf)

	beforeSchemata := len(c.Module.Schemata)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(InterpretDecl range): %v", err)
	}

	afterSchemata := len(c.Module.Schemata)
	if afterSchemata <= beforeSchemata {
		t.Fatalf("expected Schemata to grow after range interpretation (range theory adds schemata), before=%d after=%d",
			beforeSchemata, afterSchemata)
	}
}

// TestInterpretSolverSortCallsCompileTheory checks that interpreting a sort
// as a solver sort string calls CompileTheory.
func TestMiss2_InterpretSolverSortCallsCompileTheory(t *testing.T) {
	c := newTestCompiler()
	d := NewDomainSetup(c)

	myintSort := &lg.UninterpretedSort{Name: "myint"}
	c.Sig.Sorts["myint"] = myintSort

	// interpret myint = int  (simple string RHS, known solver sort)
	lhs := ast.NewSymbol("myint", nil)
	rhs := ast.NewSymbol("int", nil)
	def := ast.NewDefinition(lhs, rhs)
	lf := ast.NewLabeledFormula(nil, def)
	decl := ast.NewInterpretDecl(lf)

	beforeSchemata := len(c.Module.Schemata)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(InterpretDecl solver sort): %v", err)
	}

	afterSchemata := len(c.Module.Schemata)
	if afterSchemata <= beforeSchemata {
		t.Fatalf("expected Schemata to grow after solver sort interpretation (theory adds schemata), before=%d after=%d",
			beforeSchemata, afterSchemata)
	}
}
