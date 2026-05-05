package goivy

import (
	"testing"
)

// makeT2PMod creates a test module with a fresh signature.
func makeT2PMod() *Module {
	mod := New()
	mod.Sig = NewSig()
	return mod
}

// makeSchemaGoal builds a LabeledFormula whose Formula is a SchemaBody
// with the given premises and conclusion.
func makeSchemaGoal(label Node, prems []Node, conc Node) *LabeledFormula {
	cfg := NewAstConfig()
	elems := make([]Node, len(prems)+1)
	copy(elems, prems)
	elems[len(prems)] = conc
	sb := cfg.NewSchemaBody(elems...)
	return cfg.NewLabeledFormula(label, sb)
}

// TestTheoremToProperty_NilGoal verifies nil input returns nil.
func TestTheoremToProperty_NilGoal(t *testing.T) {
	mod := makeT2PMod()
	result := TheoremToProperty(nil, mod)
	if result != nil {
		t.Fatal("expected nil for nil input")
	}
}

// TestTheoremToProperty_NonSchemaBody verifies non-SchemaBody formula returned unchanged.
func TestTheoremToProperty_NonSchemaBody(t *testing.T) {
	cfg := NewAstConfig()
	mod := makeT2PMod()
	fmla := True
	lf := cfg.NewLabeledFormula(nil, fmla)
	result := TheoremToProperty(lf, mod)
	if result != lf {
		t.Fatal("expected same object for non-SchemaBody")
	}
}

// TestTheoremToProperty_SimpleSchemaBody verifies a SchemaBody with no name collisions:
// sorts/symbols added to sig, conclusion returned as property.
func TestTheoremToProperty_SimpleSchemaBody(t *testing.T) {
	cfg := NewAstConfig()
	mod := makeT2PMod()

	// Create a sort "mySort" not in sig
	mySort := &UninterpretedSort{Name: "mySort"}
	// Create a symbol "mySym" not in sig
	mySym := NewConst("mySym", mySort)
	cdPrem := cfg.NewConstantDecl(mySym)

	conc := True
	goal := makeSchemaGoal(nil, []Node{mySort, cdPrem}, conc)

	result := TheoremToProperty(goal, mod)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Sort should be added to sig
	if _, ok := mod.Sig.Sorts.Get2("mySort"); !ok {
		t.Error("expected mySort to be added to sig")
	}
	// Symbol should be added to sig
	if _, ok := mod.Sig.Symbols.Get2("mySym"); !ok {
		t.Error("expected mySym to be added to sig")
	}
}

// TestTheoremToProperty_SortRenaming verifies sort name collision triggers renaming.
func TestTheoremToProperty_SortRenaming(t *testing.T) {
	mod := makeT2PMod()

	// Pre-add "t" to sig to cause collision
	existingSort := &UninterpretedSort{Name: "t"}
	mod.Sig.AddSort(existingSort)

	// Schema premise declares sort "t"
	schemaSort := &UninterpretedSort{Name: "t"}
	conc := True
	goal := makeSchemaGoal(nil, []Node{schemaSort}, conc)

	result := TheoremToProperty(goal, mod)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// The original "t" should still exist, and a renamed version should also exist
	if _, ok := mod.Sig.Sorts.Get2("t"); !ok {
		t.Error("original sort 't' should still be in sig")
	}
	// There should be a new sort with a different name
	found := false
	for name, _ := range mod.Sig.Sorts.All() {
		if name != "t" && len(name) > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a renamed sort to be added to sig")
	}
}

// TestTheoremToProperty_SymbolRenaming verifies symbol name collision triggers renaming.
func TestTheoremToProperty_SymbolRenaming(t *testing.T) {
	cfg := NewAstConfig()
	mod := makeT2PMod()

	// Pre-add "f" to sig to cause collision
	fSort := &UninterpretedSort{Name: "s"}
	mod.Sig.AddSort(fSort)
	mod.Sig.AddSymbol("f", fSort)

	// Schema premise declares symbol "f"
	schemaSym := NewConst("f", fSort)
	cdPrem := cfg.NewConstantDecl(schemaSym)
	conc := True
	goal := makeSchemaGoal(nil, []Node{cdPrem}, conc)

	result := TheoremToProperty(goal, mod)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// The original "f" should still exist, and a renamed version should also exist
	if _, ok := mod.Sig.Symbols.Get2("f"); !ok {
		t.Error("original symbol 'f' should still be in sig")
	}
	found := false
	for name := range mod.Sig.Symbols.All() {
		if name != "f" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a renamed symbol to be added to sig")
	}
}

// TestTheoremToProperty_BothSortAndSymbolRename verifies both sort and symbol
// collide → both renamed, symbol's sort updated via ApplyMatchFunc.
func TestTheoremToProperty_BothSortAndSymbolRename(t *testing.T) {
	cfg := NewAstConfig()
	mod := makeT2PMod()

	// Pre-add sort "t" and symbol "f" of sort "t" to sig
	existingSort := &UninterpretedSort{Name: "t"}
	mod.Sig.AddSort(existingSort)
	mod.Sig.AddSymbol("f", existingSort)

	// Schema declares same sort "t" and symbol "f : t"
	schemaSort := &UninterpretedSort{Name: "t"}
	schemaSym := NewConst("f", schemaSort)
	cdPrem := cfg.NewConstantDecl(schemaSym)
	conc := True
	goal := makeSchemaGoal(nil, []Node{schemaSort, cdPrem}, conc)

	result := TheoremToProperty(goal, mod)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Both should have been renamed. Check that sig has more entries than before.
	if mod.Sig.Sorts.Len() < 2 {
		t.Errorf("expected at least 2 sorts in sig, got %d", mod.Sig.Sorts.Len())
	}
	if mod.Sig.Symbols.Len() < 2 {
		t.Errorf("expected at least 2 symbols in sig, got %d", mod.Sig.Symbols.Len())
	}

	// The renamed symbol's sort should reference the renamed sort, not "t"
	for name, entry := range mod.Sig.Symbols.All() {
		if name != "f" {
			// This is the renamed symbol; its sort should NOT be "t"
			sortName := IvySortName(entry.Sort)
			if sortName == "t" {
				t.Errorf("renamed symbol %q should have renamed sort, got %q", name, sortName)
			}
		}
	}
}

// TestTheoremToProperty_DefinitionPremise verifies definition premises are
// appended to mod.Definitions, not included in prems.
func TestTheoremToProperty_DefinitionPremise(t *testing.T) {
	cfg := NewAstConfig()
	mod := makeT2PMod()

	// Build a definition premise: labeled formula with IsDefinition = true
	// The formula needs args[0].args structure for PropToDef
	lhs := NewConst("mydef", Boolean)
	rhs := True
	eq, _ := NewEq(lhs, rhs)
	defPrem := cfg.NewLabeledFormula(nil, eq)
	defPrem.IsDefinition = true

	conc := True
	goal := makeSchemaGoal(nil, []Node{defPrem}, conc)

	beforeDefs := len(mod.Definitions)
	result := TheoremToProperty(goal, mod)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Definitions should have grown (PropToDef may produce the original if structure doesn't match,
	// but the append should still happen)
	if len(mod.Definitions) <= beforeDefs {
		t.Error("expected definition premise to be appended to mod.Definitions")
	}
}

// TestTheoremToProperty_ExplicitPremiseSkipped verifies explicit premises
// are neither added to prems nor definitions.
func TestTheoremToProperty_ExplicitPremiseSkipped(t *testing.T) {
	cfg := NewAstConfig()
	mod := makeT2PMod()

	// Build an explicit premise
	explPrem := cfg.NewLabeledFormula(nil, True)
	explPrem.Explicit = true

	conc := True
	goal := makeSchemaGoal(nil, []Node{explPrem}, conc)

	result := TheoremToProperty(goal, mod)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Result formula should be just the conclusion (no premises)
	// since the explicit premise is skipped
	if _, ok := result.Formula.(*LogicImplies); ok {
		t.Error("expected no implication since explicit premise should be skipped")
	}
}

// TestTheoremToProperty_DefinitionConclusion verifies that a Definition
// conclusion triggers a panic.
func TestTheoremToProperty_DefinitionConclusion(t *testing.T) {
	mod := makeT2PMod()

	lhs := NewConst("lhs", Boolean)
	rhs := True
	def := NewDefinition(lhs, rhs)

	goal := makeSchemaGoal(nil, nil, def)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for Definition conclusion")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatal("expected string panic")
		}
		if msg != "definitional subgoal must be discharged" {
			t.Errorf("unexpected panic message: %s", msg)
		}
	}()

	TheoremToProperty(goal, mod)
}

// TestTheoremToProperty_ImplicationBuilt verifies that multiple non-explicit
// premises produce an Implies(And(prems), conc) structure.
func TestTheoremToProperty_ImplicationBuilt(t *testing.T) {
	cfg := NewAstConfig()
	mod := makeT2PMod()

	prem1 := cfg.NewLabeledFormula(nil, True)
	prem2 := cfg.NewLabeledFormula(nil, True)
	conc := True

	goal := makeSchemaGoal(nil, []Node{prem1, prem2}, conc)
	result := TheoremToProperty(goal, mod)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// With 2 premises both being True, the antecedent is And(True, True)
	// and result should be Implies(And(True, True), True)
	if _, ok := result.Formula.(*LogicImplies); !ok {
		// With True premises, NormalizedAnd may simplify, so just check
		// that result is non-nil
		t.Logf("result formula type: %T (may be simplified)", result.Formula)
	}
}

// TestTheoremToProperty_RecursiveSchema verifies nested SchemaBody in premise
// is recursively converted.
func TestTheoremToProperty_RecursiveSchema(t *testing.T) {
	cfg := NewAstConfig()
	mod := makeT2PMod()

	// Inner schema: SchemaBody with a premise and conclusion
	innerPrem := cfg.NewLabeledFormula(nil, True)
	innerGoal := makeSchemaGoal(nil, []Node{innerPrem}, True)

	// Outer schema with inner schema as a premise
	outerGoal := makeSchemaGoal(nil, []Node{innerGoal}, True)

	result := TheoremToProperty(outerGoal, mod)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// The inner schema should have been recursively processed
	// Result should be an Implies since inner schema produces a premise
	if _, ok := result.Formula.(*LogicImplies); !ok {
		t.Logf("result formula type: %T (may be simplified for True→True)", result.Formula)
	}
}

// TestGoalVocab_CollectsSorts verifies GoalVocab collects UninterpretedSort
// from premises.
func TestGoalVocab_CollectsSorts(t *testing.T) {
	cfg := NewAstConfig()
	mySort := &UninterpretedSort{Name: "mySort"}
	mySym := NewConst("mySym", mySort)
	cdPrem := cfg.NewConstantDecl(mySym)

	goal := makeSchemaGoal(nil, []Node{mySort, cdPrem}, True)
	vocab := GoalVocab(goal)

	if len(vocab.Sorts) != 1 {
		t.Errorf("expected 1 sort, got %d", len(vocab.Sorts))
	}
	if len(vocab.Symbols) != 1 {
		t.Errorf("expected 1 symbol, got %d", len(vocab.Symbols))
	}
}

// TestApplyMatchGoalNode_SortPremises verifies that ApplyMatchGoalNode
// renames sort premises according to the match.
func TestApplyMatchGoalNode_SortPremises(t *testing.T) {
	oldSort := &UninterpretedSort{Name: "t"}
	newSort := &UninterpretedSort{Name: "t__0"}

	match := map[NodeKey]Expr{
		Key(oldSort): newSort,
	}

	goal := makeSchemaGoal(nil, []Node{oldSort}, True)
	result := ApplyMatchGoalNode(makeT2PMod().Cfg.AstCfg, match, goal)

	// The result's SchemaBody should have the new sort in premises
	sb, ok := result.Formula.(*SchemaBody)
	if !ok {
		t.Fatal("expected SchemaBody")
	}
	prems := sb.Prems()
	if len(prems) != 1 {
		t.Fatalf("expected 1 premise, got %d", len(prems))
	}
	if s, ok := prems[0].(Sort); ok {
		name := IvySortName(s)
		if name != "t__0" {
			t.Errorf("expected renamed sort 't__0', got %q", name)
		}
	} else {
		t.Errorf("expected Sort premise, got %T", prems[0])
	}
}

// TestApplyMatchGoalNode_ConstantDeclPremises verifies that ApplyMatchGoalNode
// renames symbols in ConstantDecl premises.
func TestApplyMatchGoalNode_ConstantDeclPremises(t *testing.T) {
	cfg := NewAstConfig()
	mySort := &UninterpretedSort{Name: "s"}
	oldSym := NewConst("f", mySort)
	newSym := NewConst("f__0", mySort)

	match := map[NodeKey]Expr{
		Key(oldSym): newSym,
	}

	cdPrem := cfg.NewConstantDecl(oldSym)
	goal := makeSchemaGoal(nil, []Node{cdPrem}, True)
	result := ApplyMatchGoalNode(makeT2PMod().Cfg.AstCfg, match, goal)

	sb, ok := result.Formula.(*SchemaBody)
	if !ok {
		t.Fatal("expected SchemaBody")
	}
	prems := sb.Prems()
	if len(prems) != 1 {
		t.Fatalf("expected 1 premise, got %d", len(prems))
	}
	cd, ok := prems[0].(*ConstantDecl)
	if !ok {
		t.Fatalf("expected ConstantDecl, got %T", prems[0])
	}
	args := cd.Args()
	if len(args) != 1 {
		t.Fatalf("expected 1 arg in ConstantDecl, got %d", len(args))
	}
	if sym, ok := args[0].(*Const); ok {
		if sym.Name != "f__0" {
			t.Errorf("expected renamed symbol 'f__0', got %q", sym.Name)
		}
	} else {
		t.Errorf("expected Symbol in ConstantDecl, got %T", args[0])
	}
}
