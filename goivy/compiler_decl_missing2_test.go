package goivy

import (
	"strings"
	"testing"
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
	cfg := NewAstConfig()
	c := newTestCompiler()
	as := NewARGSetup(c)

	// state link = true
	// Python: a.args[0].relname -> "link", a.args[1] -> the body
	lhs := cfg.NewAtom("link")
	rhs := cfg.NewAtom("true")
	def := cfg.NewDefinition(lhs, rhs)
	lf := cfg.NewLabeledFormula(nil, def)
	decl := cfg.NewStateDecl(lf)

	err := as.ProcessDecls([]Node{decl})
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
	cfg := NewAstConfig()
	c := newTestCompiler()
	as := NewARGSetup(c)

	// state link = true
	lf1 := cfg.NewLabeledFormula(nil, cfg.NewDefinition(cfg.NewAtom("link"), cfg.NewAtom("true")))
	decl1 := cfg.NewStateDecl(lf1)

	// state conn = false
	lf2 := cfg.NewLabeledFormula(nil, cfg.NewDefinition(cfg.NewAtom("conn"), cfg.NewAtom("false")))
	decl2 := cfg.NewStateDecl(lf2)

	err := as.ProcessDecls([]Node{decl1, decl2})
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
// Python: "LogicVariable{} occurs twice on left-hand side of definition"
func TestMiss2_DefinitionDuplicateLHSVariable(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	// Add the sort and defined symbol so compilation reaches add_definition,
	// where Python performs the duplicate-variable validation.
	tSort := &UninterpretedSort{Name: "t"}
	c.Sig.Sorts.Set("t", tSort)
	fSort, err := NewFunctionSort(tSort, tSort, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Sig.AddSymbol("f", fSort); err != nil {
		t.Fatal(err)
	}

	// definition f(X, X) = body  — X appears twice on LHS
	x1 := cfg.NewVariable("X", "t")
	x2 := cfg.NewVariable("X", "t")
	lhs := cfg.NewAtom("f", x1, x2)
	rhs := cfg.NewAtom("true")
	def := cfg.NewDefinition(lhs, rhs)
	lf := cfg.NewLabeledFormula(nil, def)
	decl := cfg.NewDefinitionDecl(lf)

	err = d.ProcessDecl(decl)
	if err == nil {
		t.Fatal("expected error for duplicate LHS variable, got nil")
	}
	if !strings.Contains(err.Error(), "occurs twice on left-hand side") {
		t.Errorf("expected error about 'occurs twice on left-hand side', got: %v", err)
	}
}

// TestDefinitionFreeRHSVariable checks that a definition with an unbound RHS
// variable raises an error.
// Python: "LogicVariable{} occurs free on right-hand side of definition"
func TestMiss2_DefinitionFreeRHSVariable(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	tSort := &UninterpretedSort{Name: "t"}
	c.Sig.Sorts.Set("t", tSort)
	fSort, err := NewFunctionSort(tSort, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Sig.AddSymbol("f", fSort); err != nil {
		t.Fatal(err)
	}
	gSort, err := NewFunctionSort(tSort, tSort, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Sig.AddSymbol("g", gSort); err != nil {
		t.Fatal(err)
	}

	// definition f(X) = g(X, Y)  — Y is free on RHS but not on LHS
	x := cfg.NewVariable("X", "t")
	lhs := cfg.NewAtom("f", x)
	xRef := cfg.NewVariable("X", "")
	yFree := cfg.NewVariable("Y", "t")
	rhs := cfg.NewAtom("g", xRef, yFree)
	def := cfg.NewDefinition(lhs, rhs)
	lf := cfg.NewLabeledFormula(nil, def)
	decl := cfg.NewDefinitionDecl(lf)

	err = d.ProcessDecl(decl)
	if err == nil {
		t.Fatal("expected error for free RHS variable, got nil")
	}
	if !strings.Contains(err.Error(), "occurs free on right-hand side") {
		t.Errorf("expected error about 'occurs free on right-hand side', got: %v", err)
	}
}

func TestDefinitionDeclDoesNotPreloadPolymorphicLHSIntoSig(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)
	c.Sig.Sorts.Set("t", &UninterpretedSort{Name: "t"})

	x := cfg.NewVariable("X", "t")
	y := cfg.NewVariable("Y", "t")
	lhs := cfg.NewAtom("<", x, y)
	rhs := cfg.NewAtom("missing", x)
	def := cfg.NewDefinition(lhs, rhs)
	lf := cfg.NewLabeledFormula(nil, def)
	decl := cfg.NewDefinitionDecl(lf)

	if err := d.ProcessDecl(decl); err == nil {
		t.Fatal("expected unknown symbol error")
	}
	if _, ok := c.Sig.Symbols.Get2("<"); ok {
		t.Fatal("normal definition processing must not pre-load polymorphic '<' into sig.symbols")
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
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	boolSort := &UninterpretedSort{Name: "bool"}
	c.Sig.Sorts.Set("bool", boolSort)

	// derived foo(X:bool) = true
	x := cfg.NewVariable("X", "bool")
	lhs := cfg.NewAtom("foo", x)
	rhs := cfg.NewAtom("true")
	def := cfg.NewDefinition(lhs, rhs)
	lf := cfg.NewLabeledFormula(nil, def)
	decl := cfg.NewDerivedDecl(lf)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(DerivedDecl): %v", err)
	}

	if len(c.Module.Updates) == 0 {
		t.Fatal("expected Updates to have at least 1 entry after DerivedDecl, got 0")
	}

	// Type-assert last entry to *actions.DerivedUpdate
	last := c.Module.Updates[len(c.Module.Updates)-1]
	if _, ok := last.(*DerivedUpdate); !ok {
		t.Errorf("expected last Update to be *actions.DerivedUpdate, got %T", last)
	}
}

// TestDefinitionDeclCreatesDerivedUpdate checks that a DefinitionDecl also causes a
// DerivedUpdate to be appended to mod.Updates.
func TestMiss2_DefinitionDeclCreatesDerivedUpdate(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	boolSort := &UninterpretedSort{Name: "bool"}
	c.Sig.Sorts.Set("bool", boolSort)

	// definition bar(X:bool) = true
	x := cfg.NewVariable("X", "bool")
	lhs := cfg.NewAtom("bar", x)
	rhs := cfg.NewAtom("true")
	def := cfg.NewDefinition(lhs, rhs)
	lf := cfg.NewLabeledFormula(nil, def)
	decl := cfg.NewDefinitionDecl(lf)

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(DefinitionDecl): %v", err)
	}

	if len(c.Module.Updates) == 0 {
		t.Fatal("expected Updates to have at least 1 entry after DefinitionDecl, got 0")
	}

	last := c.Module.Updates[len(c.Module.Updates)-1]
	if _, ok := last.(*DerivedUpdate); !ok {
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
	cfg := NewAstConfig()
	c := newTestCompiler()

	boolSort := &UninterpretedSort{Name: "bool"}
	c.Sig.Sorts.Set("bool", boolSort)
	c.Sig.AddSymbol("s", boolSort)

	// Add an axiom that references symbol "s" — use compiled lg.Const so structural keys match
	sSym := &Const{Name: "s", CSort: boolSort}
	axiomLf := cfg.NewLabeledFormula(nil, sSym)
	c.Module.LabeledAxioms = append(c.Module.LabeledAxioms, axiomLf)

	// Add an action that modifies "s"
	assignAction := NewAssignAction(
		&Const{Name: "s", CSort: boolSort},
		&Const{Name: "true"},
	)
	c.Module.Actions.Set("test_action", assignAction)

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
	cfg := NewAstConfig()
	c := newTestCompiler()

	boolSort := &UninterpretedSort{Name: "bool"}
	c.Sig.Sorts.Set("bool", boolSort)
	c.Sig.AddSymbol("s", boolSort)

	axiomAtom := cfg.NewAtom("s")
	axiomLf := cfg.NewLabeledFormula(nil, axiomAtom)
	c.Module.LabeledAxioms = append(c.Module.LabeledAxioms, axiomLf)

	assignAction := NewAssignAction(
		&Const{Name: "s", CSort: boolSort},
		&Const{Name: "true"},
	)
	c.Module.Actions.Set("test_action", assignAction)

	// mutax enabled (true) → should NOT error
	err := CheckMutax(c.Module, true)
	if err != nil {
		t.Fatalf("expected no error when mutax enabled, got: %v", err)
	}
}

// TestCheckMutaxDefinitionLHS checks that CheckMutax catches when a
// definition's LHS symbol is modified by an action.
func TestMiss2_CheckMutaxDefinitionLHS(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	boolSort := &UninterpretedSort{Name: "bool"}
	c.Sig.Sorts.Set("bool", boolSort)
	c.Sig.AddSymbol("f", boolSort)

	// Add a compiled definition for "f" — use lg.Definition with lg.Const
	// to match what the real compiler produces (Python: mod.definitions has compiled defs)
	fSym := &Const{Name: "f", CSort: boolSort}
	falseSym := &Const{Name: "false"}
	logicDef := NewDefinition(fSym, falseSym)
	defLf := cfg.NewLabeledFormula(nil, logicDef)
	c.Module.Definitions = append(c.Module.Definitions, defLf)

	// Add an action that modifies "f"
	assignAction := NewAssignAction(
		&Const{Name: "f", CSort: boolSort},
		&Const{Name: "val"},
	)
	c.Module.Actions.Set("test_action", assignAction)

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
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	intSort := &UninterpretedSort{Name: "myint"}
	c.Sig.Sorts.Set("myint", intSort)

	// interpret myint = <<<int>>>  (NativeType with code "int")
	lhs := cfg.NewSymbol("myint", nil)
	rhs := cfg.NewNativeType(cfg.NewAtom("int"))
	impl := cfg.NewImplies(lhs, rhs)
	lf := cfg.NewLabeledFormula(nil, impl)
	decl := cfg.NewInterpretDecl(lf)

	beforeSchemata := c.Module.Schemata.Len()

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(InterpretDecl native int): %v", err)
	}

	afterSchemata := c.Module.Schemata.Len()
	if afterSchemata <= beforeSchemata {
		t.Fatalf("expected Schemata to grow after native int interpretation (int theory adds rec/ind/lep schemata), before=%d after=%d",
			beforeSchemata, afterSchemata)
	}
}

// TestInterpretRangeCallsCompileTheory checks that interpreting a sort as a
// range calls CompileTheory, which should add range theory axioms.
func TestMiss2_InterpretRangeCallsCompileTheory(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	myintSort := &UninterpretedSort{Name: "myint"}
	c.Sig.Sorts.Set("myint", myintSort)

	// interpret myint = 0..100
	lhs := cfg.NewSymbol("myint", nil)
	rhs := cfg.NewRange(cfg.NewAtom("0"), cfg.NewAtom("100"))
	impl := cfg.NewImplies(lhs, rhs)
	lf := cfg.NewLabeledFormula(nil, impl)
	decl := cfg.NewInterpretDecl(lf)

	beforeSchemata := c.Module.Schemata.Len()

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(InterpretDecl range): %v", err)
	}

	afterSchemata := c.Module.Schemata.Len()
	if afterSchemata <= beforeSchemata {
		t.Fatalf("expected Schemata to grow after range interpretation (range theory adds schemata), before=%d after=%d",
			beforeSchemata, afterSchemata)
	}
}

// TestInterpretSolverSortCallsCompileTheory checks that interpreting a sort
// as a solver sort string calls CompileTheory.
func TestMiss2_InterpretSolverSortCallsCompileTheory(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	d := NewDomainSetup(c)

	myintSort := &UninterpretedSort{Name: "myint"}
	c.Sig.Sorts.Set("myint", myintSort)

	// interpret myint = int  (simple string RHS, known solver sort)
	lhs := cfg.NewSymbol("myint", nil)
	rhs := cfg.NewSymbol("int", nil)
	impl := cfg.NewImplies(lhs, rhs)
	lf := cfg.NewLabeledFormula(nil, impl)
	decl := cfg.NewInterpretDecl(lf)

	beforeSchemata := c.Module.Schemata.Len()

	err := d.ProcessDecl(decl)
	if err != nil {
		t.Fatalf("ProcessDecl(InterpretDecl solver sort): %v", err)
	}

	afterSchemata := c.Module.Schemata.Len()
	if afterSchemata <= beforeSchemata {
		t.Fatalf("expected Schemata to grow after solver sort interpretation (theory adds schemata), before=%d after=%d",
			beforeSchemata, afterSchemata)
	}
}
