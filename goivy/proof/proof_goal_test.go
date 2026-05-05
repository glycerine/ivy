package proof

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// Reuses proofMkSort, proofMkVar, proofMkConst, mkLF, proofTestAstCfg from proof_test.go.

// --- GoalIsDefn (Bug 1 regression) ---

func TestGoalIsDefn_ConstantDecl(t *testing.T) {
	s := proofMkSort("S")
	c := proofMkConst("f", s)
	cd := proofTestAstCfg.NewConstantDecl(c)
	if !GoalIsDefn(cd) {
		t.Error("ConstantDecl with Const arg should be a definition")
	}
}

func TestGoalIsDefn_ConstDeclLambda(t *testing.T) {
	s := proofMkSort("S")
	v := proofMkVar("X", s)
	lam, _ := lg.NewLambda([]*lg.Variable{v}, v)
	cd := proofTestAstCfg.NewConstantDecl(lam)
	if GoalIsDefn(cd) {
		t.Error("ConstantDecl with Lambda arg should NOT be a definition")
	}
}

func TestGoalIsDefn_UninterpretedSort(t *testing.T) {
	// Bug 1 regression: previously returned false.
	us := &lg.UninterpretedSort{Name: "S"}
	if !GoalIsDefn(us) {
		t.Error("UninterpretedSort should be a definition (Bug 1 fix)")
	}
}

func TestGoalIsDefn_PlainConst(t *testing.T) {
	c := proofMkConst("c", proofMkSort("S"))
	if GoalIsDefn(c) {
		t.Error("plain Const should NOT be a definition")
	}
}

func TestGoalIsDefn_ConstDeclNoArgs(t *testing.T) {
	cd := proofTestAstCfg.NewConstantDecl()
	if !GoalIsDefn(cd) {
		t.Error("ConstantDecl with no args should be a definition (empty args, no Lambda)")
	}
}

// --- GoalDefns (Bug 5 regression) ---

func TestGoalDefns_ConstPrem(t *testing.T) {
	s := proofMkSort("S")
	c := proofMkConst("f", s)
	cd := proofTestAstCfg.NewConstantDecl(c)
	conc := proofMkConst("conc", s)
	sb := proofTestAstCfg.NewSchemaBody(cd, conc)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), sb)

	defns := GoalDefns(goal)
	if len(defns) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(defns))
	}
	// The key should match the Const.
	if _, ok := defns[lg.Key(c)]; !ok {
		t.Error("expected definition keyed by Const")
	}
}

func TestGoalDefns_VariablePrem(t *testing.T) {
	// Bug 5 regression: ConstantDecl with Variable (not Const) should be excluded.
	s := proofMkSort("S")
	v := proofMkVar("X", s)
	cd := proofTestAstCfg.NewConstantDecl(v)
	conc := proofMkConst("conc", s)
	sb := proofTestAstCfg.NewSchemaBody(cd, conc)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), sb)

	defns := GoalDefns(goal)
	if len(defns) != 0 {
		t.Errorf("expected 0 definitions for Variable premise (Bug 5 fix), got %d", len(defns))
	}
}

func TestGoalDefns_UninterpSortPrem(t *testing.T) {
	us := &lg.UninterpretedSort{Name: "T"}
	conc := proofMkConst("conc", proofMkSort("S"))
	sb := proofTestAstCfg.NewSchemaBody(us, conc)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), sb)

	defns := GoalDefns(goal)
	if len(defns) != 1 {
		t.Fatalf("expected 1 definition for UninterpretedSort premise, got %d", len(defns))
	}
}

func TestGoalDefns_NoPremises(t *testing.T) {
	conc := proofMkConst("conc", proofMkSort("S"))
	goal := mkLF(proofTestAstCfg.NewAtom("g"), conc)
	defns := GoalDefns(goal)
	if len(defns) != 0 {
		t.Errorf("expected 0 definitions for no premises, got %d", len(defns))
	}
}

// --- GoalSubst (Bug 2 regression) ---

func TestGoalSubst_OK(t *testing.T) {
	s := proofMkSort("S")
	f := proofMkConst("f", s)
	g := proofMkConst("g", s)
	conc1 := proofMkConst("c1", s)
	conc2 := proofMkConst("c2", s)

	cd1 := proofTestAstCfg.NewConstantDecl(f)
	sb1 := proofTestAstCfg.NewSchemaBody(cd1, conc1)
	g1 := mkLF(proofTestAstCfg.NewAtom("g1"), sb1)

	cd2 := proofTestAstCfg.NewConstantDecl(g)
	sb2 := proofTestAstCfg.NewSchemaBody(cd2, conc2)
	g2 := mkLF(proofTestAstCfg.NewAtom("g2"), sb2)

	result, err := GoalSubst(proofTestAstCfg, g1, g2, ast.Location{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	prems := GoalPrems(result)
	if len(prems) != 2 {
		t.Errorf("expected 2 combined premises, got %d", len(prems))
	}
}

func TestGoalSubst_NameClash(t *testing.T) {
	// Bug 2 regression: overlapping defns should return error.
	s := proofMkSort("S")
	f := proofMkConst("f", s)
	conc1 := proofMkConst("c1", s)
	conc2 := proofMkConst("c2", s)

	cd1 := proofTestAstCfg.NewConstantDecl(f)
	sb1 := proofTestAstCfg.NewSchemaBody(cd1, conc1)
	g1 := mkLF(proofTestAstCfg.NewAtom("g1"), sb1)

	cd2 := proofTestAstCfg.NewConstantDecl(f) // same symbol "f"
	sb2 := proofTestAstCfg.NewSchemaBody(cd2, conc2)
	g2 := mkLF(proofTestAstCfg.NewAtom("g2"), sb2)

	_, err := GoalSubst(proofTestAstCfg, g1, g2, ast.Location{})
	if err == nil {
		t.Fatal("expected error for clashing premise names (Bug 2 fix)")
	}
}

// --- GoalVocabBound (Bug 4 regression) ---

func TestGoalVocabBound_IncludesBound(t *testing.T) {
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	body := &lg.ForAll{Variables: []*lg.Variable{x}, Body: lg.True}
	goal := mkLF(proofTestAstCfg.NewAtom("g"), body)

	v := GoalVocabBound(goal)
	found := false
	for _, vr := range v.Variables {
		if vr.Name == "X" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GoalVocabBound should include bound variable X from conclusion")
	}
}

func TestGoalVocabBound_TemporalModels(t *testing.T) {
	// Bug 4 regression: TemporalModels wrapping a ForAll.
	s := proofMkSort("S")
	x := proofMkVar("X", s)
	body := &lg.ForAll{Variables: []*lg.Variable{x}, Body: lg.True}
	tm := &ast.AstTemporalModels{Fmla: body}
	goal := mkLF(proofTestAstCfg.NewAtom("g"), tm)

	v := GoalVocabBound(goal)
	found := false
	for _, vr := range v.Variables {
		if vr.Name == "X" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GoalVocabBound should include bound variable X from TemporalModels conclusion (Bug 4 fix)")
	}
}

func TestGoalVocabBound_NilConc(t *testing.T) {
	// Non-Expr conclusion — should return same as GoalVocab.
	atom := proofTestAstCfg.NewAtom("a")
	goal := mkLF(proofTestAstCfg.NewAtom("g"), atom)
	v := GoalVocabBound(goal)
	if v == nil {
		t.Fatal("expected non-nil vocab")
	}
}

// --- ConcAsExpr / GoalConcUnwrap ---

func TestConcAsExpr_Expr(t *testing.T) {
	s := proofMkSort("S")
	c := proofMkConst("c", s)
	result := ConcAsExpr(c)
	if result != c {
		t.Error("expected Const passed through")
	}
}

func TestConcAsExpr_TemporalModels(t *testing.T) {
	s := proofMkSort("S")
	c := proofMkConst("c", s)
	tm := &ast.AstTemporalModels{Fmla: c}
	result := ConcAsExpr(tm)
	if result != c {
		t.Error("expected inner Const from TemporalModels")
	}
}

func TestConcAsExpr_NonExpr(t *testing.T) {
	atom := proofTestAstCfg.NewAtom("a")
	result := ConcAsExpr(atom)
	if result != nil {
		t.Errorf("expected nil for non-Expr, got %T", result)
	}
}

func TestGoalConcUnwrap(t *testing.T) {
	s := proofMkSort("S")
	c := proofMkConst("c", s)
	tm := &ast.AstTemporalModels{Fmla: c}
	goal := mkLF(proofTestAstCfg.NewAtom("g"), tm)
	result := GoalConcUnwrap(goal)
	if result != c {
		t.Error("expected inner Const from GoalConcUnwrap")
	}
}

// --- ApplyToConc / GoalApplyToConc ---

func TestApplyToConc_PlainExpr(t *testing.T) {
	c := proofMkConst("c", proofMkSort("S"))
	negate := func(e lg.Expr) lg.Expr { return &lg.Not{Body: e} }
	result := ApplyToConc(c, negate)
	if not, ok := result.(*lg.Not); !ok || not.Body != c {
		t.Error("expected Not{c}")
	}
}

func TestApplyToConc_TemporalModels(t *testing.T) {
	c := proofMkConst("c", proofMkSort("S"))
	tm := &ast.AstTemporalModels{Fmla: c}
	negate := func(e lg.Expr) lg.Expr { return &lg.Not{Body: e} }
	result := ApplyToConc(tm, negate)
	rtm, ok := result.(*ast.AstTemporalModels)
	if !ok {
		t.Fatalf("expected *ast.TemporalModels, got %T", result)
	}
	if _, ok := rtm.Fmla.(*lg.Not); !ok {
		t.Error("expected inner formula to be negated")
	}
}

func TestApplyToConc_NonExpr(t *testing.T) {
	atom := proofTestAstCfg.NewAtom("a")
	negate := func(e lg.Expr) lg.Expr { return &lg.Not{Body: e} }
	result := ApplyToConc(atom, negate)
	if result != atom {
		t.Error("non-Expr should pass through unchanged")
	}
}

func TestGoalApplyToConc(t *testing.T) {
	s := proofMkSort("S")
	c := proofMkConst("c", s)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), c)
	fn := func(n ast.Node) ast.Node {
		if e, ok := n.(lg.Expr); ok {
			return &lg.Not{Body: e}
		}
		return n
	}
	result := GoalApplyToConc(proofTestAstCfg, goal, fn)
	conc := GoalConc(result)
	if _, ok := conc.(*lg.Not); !ok {
		t.Errorf("expected Not conclusion, got %T", conc)
	}
}

// --- GoalAddPrem / GoalRemovePrem ---

func TestGoalAddPrem_Empty(t *testing.T) {
	s := proofMkSort("S")
	c := proofMkConst("c", s)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), c)
	prem := mkLF(proofTestAstCfg.NewAtom("p"), c)
	result := GoalAddPrem(proofTestAstCfg, goal, prem, ast.Location{})
	prems := GoalPrems(result)
	if len(prems) != 1 {
		t.Errorf("expected 1 premise, got %d", len(prems))
	}
}

func TestGoalAddPrem_Existing(t *testing.T) {
	s := proofMkSort("S")
	c := proofMkConst("c", s)
	prem1 := mkLF(proofTestAstCfg.NewAtom("p1"), c)
	sb := proofTestAstCfg.NewSchemaBody(prem1, c)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), sb)
	prem2 := mkLF(proofTestAstCfg.NewAtom("p2"), c)
	result := GoalAddPrem(proofTestAstCfg, goal, prem2, ast.Location{})
	prems := GoalPrems(result)
	if len(prems) != 2 {
		t.Errorf("expected 2 premises, got %d", len(prems))
	}
}

func TestGoalAddPrem_PreservesTemporalModelsConclusion(t *testing.T) {
	s := proofMkSort("S")
	c := proofMkConst("c", s)
	tm := proofTestAstCfg.NewTemporalModels(proofTestAstCfg.NewNoneAST(), lg.True)

	// Goal: SchemaBody{[ConstantDecl, TemporalModels]}
	cd := proofTestAstCfg.NewConstantDecl(c)
	sb := proofTestAstCfg.NewSchemaBody(cd, tm)
	goal := mkLF(proofTestAstCfg.NewAtom("goal"), sb)

	prem := mkLF(proofTestAstCfg.NewAtom("prem"), c)

	// Verify precondition
	if _, ok := GoalConc(goal).(*ast.AstTemporalModels); !ok {
		t.Fatalf("precondition: expected TemporalModels conclusion, got %T", GoalConc(goal))
	}

	result := GoalAddPrem(proofTestAstCfg, goal, prem, ast.Location{})

	// The conclusion must still be TemporalModels
	conc := GoalConc(result)
	if _, ok := conc.(*ast.AstTemporalModels); !ok {
		t.Fatalf("GoalAddPrem lost TemporalModels conclusion, got %T", conc)
	}
	// Should have 2 premises now: ConstantDecl + prem
	prems := GoalPrems(result)
	if len(prems) != 2 {
		t.Fatalf("expected 2 premises, got %d", len(prems))
	}
}

func TestGoalRemovePrem_Found(t *testing.T) {
	s := proofMkSort("S")
	c := proofMkConst("c", s)
	prem := mkLF(proofTestAstCfg.NewAtom("removeme"), c)
	sb := proofTestAstCfg.NewSchemaBody(prem, c)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), sb)

	result := GoalRemovePrem(proofTestAstCfg, goal, "removeme")
	prems := GoalPrems(result)
	if len(prems) != 0 {
		t.Errorf("expected 0 premises after removal, got %d", len(prems))
	}
}

func TestGoalRemovePrem_NotFound(t *testing.T) {
	s := proofMkSort("S")
	c := proofMkConst("c", s)
	prem := mkLF(proofTestAstCfg.NewAtom("keepme"), c)
	sb := proofTestAstCfg.NewSchemaBody(prem, c)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), sb)

	result := GoalRemovePrem(proofTestAstCfg, goal, "nonexistent")
	prems := GoalPrems(result)
	if len(prems) != 1 {
		t.Errorf("expected 1 premise unchanged, got %d", len(prems))
	}
}

// --- TrivialGoal positive case ---

func TestTrivialGoal_Positive(t *testing.T) {
	s := proofMkSort("S")
	c := proofMkConst("c", s)
	// Build inner premise goal with same conclusion as outer.
	premGoal := mkLF(proofTestAstCfg.NewAtom("prem"), c)
	sb := proofTestAstCfg.NewSchemaBody(premGoal, c)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), sb)

	if !TrivialGoal(goal) {
		t.Error("goal where premise conclusion equals outer conclusion should be trivial")
	}
}

// --- Fuzz ---

func FuzzGoalIsDefn(f *testing.F) {
	f.Add(0) // ConstantDecl
	f.Add(1) // UninterpretedSort
	f.Add(2) // plain Const
	f.Add(3) // Lambda

	f.Fuzz(func(t *testing.T, choice int) {
		s := proofMkSort("S")
		var node ast.Node
		switch choice % 4 {
		case 0:
			node = proofTestAstCfg.NewConstantDecl(proofMkConst("f", s))
		case 1:
			node = &lg.UninterpretedSort{Name: "S"}
		case 2:
			node = proofMkConst("c", s)
		case 3:
			v := proofMkVar("X", s)
			lam, _ := lg.NewLambda([]*lg.Variable{v}, v)
			node = proofTestAstCfg.NewConstantDecl(lam)
		}
		_ = GoalIsDefn(node) // must not panic
	})
}
