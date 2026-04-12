package proof

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// Reuses mkSort, mkVar, mkConst, mkLF, testAstCfg from proof_test.go.

// --- GoalIsDefn (Bug 1 regression) ---

func TestGoalIsDefn_ConstantDecl(t *testing.T) {
	s := mkSort("S")
	c := mkConst("f", s)
	cd := testAstCfg.NewConstantDecl(c)
	if !GoalIsDefn(cd) {
		t.Error("ConstantDecl with Const arg should be a definition")
	}
}

func TestGoalIsDefn_ConstDeclLambda(t *testing.T) {
	s := mkSort("S")
	v := mkVar("X", s)
	lam, _ := lg.NewLambda([]*lg.Variable{v}, v)
	cd := testAstCfg.NewConstantDecl(lam)
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
	c := mkConst("c", mkSort("S"))
	if GoalIsDefn(c) {
		t.Error("plain Const should NOT be a definition")
	}
}

func TestGoalIsDefn_ConstDeclNoArgs(t *testing.T) {
	cd := testAstCfg.NewConstantDecl()
	if !GoalIsDefn(cd) {
		t.Error("ConstantDecl with no args should be a definition (empty args, no Lambda)")
	}
}

// --- GoalDefns (Bug 5 regression) ---

func TestGoalDefns_ConstPrem(t *testing.T) {
	s := mkSort("S")
	c := mkConst("f", s)
	cd := testAstCfg.NewConstantDecl(c)
	conc := mkConst("conc", s)
	sb := testAstCfg.NewSchemaBody(cd, conc)
	goal := mkLF(testAstCfg.NewAtom("g"), sb)

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
	s := mkSort("S")
	v := mkVar("X", s)
	cd := testAstCfg.NewConstantDecl(v)
	conc := mkConst("conc", s)
	sb := testAstCfg.NewSchemaBody(cd, conc)
	goal := mkLF(testAstCfg.NewAtom("g"), sb)

	defns := GoalDefns(goal)
	if len(defns) != 0 {
		t.Errorf("expected 0 definitions for Variable premise (Bug 5 fix), got %d", len(defns))
	}
}

func TestGoalDefns_UninterpSortPrem(t *testing.T) {
	us := &lg.UninterpretedSort{Name: "T"}
	conc := mkConst("conc", mkSort("S"))
	sb := testAstCfg.NewSchemaBody(us, conc)
	goal := mkLF(testAstCfg.NewAtom("g"), sb)

	defns := GoalDefns(goal)
	if len(defns) != 1 {
		t.Fatalf("expected 1 definition for UninterpretedSort premise, got %d", len(defns))
	}
}

func TestGoalDefns_NoPremises(t *testing.T) {
	conc := mkConst("conc", mkSort("S"))
	goal := mkLF(testAstCfg.NewAtom("g"), conc)
	defns := GoalDefns(goal)
	if len(defns) != 0 {
		t.Errorf("expected 0 definitions for no premises, got %d", len(defns))
	}
}

// --- GoalSubst (Bug 2 regression) ---

func TestGoalSubst_OK(t *testing.T) {
	s := mkSort("S")
	f := mkConst("f", s)
	g := mkConst("g", s)
	conc1 := mkConst("c1", s)
	conc2 := mkConst("c2", s)

	cd1 := testAstCfg.NewConstantDecl(f)
	sb1 := testAstCfg.NewSchemaBody(cd1, conc1)
	g1 := mkLF(testAstCfg.NewAtom("g1"), sb1)

	cd2 := testAstCfg.NewConstantDecl(g)
	sb2 := testAstCfg.NewSchemaBody(cd2, conc2)
	g2 := mkLF(testAstCfg.NewAtom("g2"), sb2)

	result, err := GoalSubst(testAstCfg, g1, g2, ast.Location{})
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
	s := mkSort("S")
	f := mkConst("f", s)
	conc1 := mkConst("c1", s)
	conc2 := mkConst("c2", s)

	cd1 := testAstCfg.NewConstantDecl(f)
	sb1 := testAstCfg.NewSchemaBody(cd1, conc1)
	g1 := mkLF(testAstCfg.NewAtom("g1"), sb1)

	cd2 := testAstCfg.NewConstantDecl(f) // same symbol "f"
	sb2 := testAstCfg.NewSchemaBody(cd2, conc2)
	g2 := mkLF(testAstCfg.NewAtom("g2"), sb2)

	_, err := GoalSubst(testAstCfg, g1, g2, ast.Location{})
	if err == nil {
		t.Fatal("expected error for clashing premise names (Bug 2 fix)")
	}
}

// --- GoalVocabBound (Bug 4 regression) ---

func TestGoalVocabBound_IncludesBound(t *testing.T) {
	s := mkSort("S")
	x := mkVar("X", s)
	body := &lg.ForAll{Variables: []*lg.Variable{x}, Body: lg.True}
	goal := mkLF(testAstCfg.NewAtom("g"), body)

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
	s := mkSort("S")
	x := mkVar("X", s)
	body := &lg.ForAll{Variables: []*lg.Variable{x}, Body: lg.True}
	tm := &ast.TemporalModels{Fmla: body}
	goal := mkLF(testAstCfg.NewAtom("g"), tm)

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
	atom := testAstCfg.NewAtom("a")
	goal := mkLF(testAstCfg.NewAtom("g"), atom)
	v := GoalVocabBound(goal)
	if v == nil {
		t.Fatal("expected non-nil vocab")
	}
}

// --- ConcAsExpr / GoalConcUnwrap ---

func TestConcAsExpr_Expr(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	result := ConcAsExpr(c)
	if result != c {
		t.Error("expected Const passed through")
	}
}

func TestConcAsExpr_TemporalModels(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	tm := &ast.TemporalModels{Fmla: c}
	result := ConcAsExpr(tm)
	if result != c {
		t.Error("expected inner Const from TemporalModels")
	}
}

func TestConcAsExpr_NonExpr(t *testing.T) {
	atom := testAstCfg.NewAtom("a")
	result := ConcAsExpr(atom)
	if result != nil {
		t.Errorf("expected nil for non-Expr, got %T", result)
	}
}

func TestGoalConcUnwrap(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	tm := &ast.TemporalModels{Fmla: c}
	goal := mkLF(testAstCfg.NewAtom("g"), tm)
	result := GoalConcUnwrap(goal)
	if result != c {
		t.Error("expected inner Const from GoalConcUnwrap")
	}
}

// --- ApplyToConc / GoalApplyToConc ---

func TestApplyToConc_PlainExpr(t *testing.T) {
	c := mkConst("c", mkSort("S"))
	negate := func(e lg.Expr) lg.Expr { return &lg.Not{Body: e} }
	result := ApplyToConc(c, negate)
	if not, ok := result.(*lg.Not); !ok || not.Body != c {
		t.Error("expected Not{c}")
	}
}

func TestApplyToConc_TemporalModels(t *testing.T) {
	c := mkConst("c", mkSort("S"))
	tm := &ast.TemporalModels{Fmla: c}
	negate := func(e lg.Expr) lg.Expr { return &lg.Not{Body: e} }
	result := ApplyToConc(tm, negate)
	rtm, ok := result.(*ast.TemporalModels)
	if !ok {
		t.Fatalf("expected *ast.TemporalModels, got %T", result)
	}
	if _, ok := rtm.Fmla.(*lg.Not); !ok {
		t.Error("expected inner formula to be negated")
	}
}

func TestApplyToConc_NonExpr(t *testing.T) {
	atom := testAstCfg.NewAtom("a")
	negate := func(e lg.Expr) lg.Expr { return &lg.Not{Body: e} }
	result := ApplyToConc(atom, negate)
	if result != atom {
		t.Error("non-Expr should pass through unchanged")
	}
}

func TestGoalApplyToConc(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	goal := mkLF(testAstCfg.NewAtom("g"), c)
	fn := func(n ast.Node) ast.Node {
		if e, ok := n.(lg.Expr); ok {
			return &lg.Not{Body: e}
		}
		return n
	}
	result := GoalApplyToConc(testAstCfg, goal, fn)
	conc := GoalConc(result)
	if _, ok := conc.(*lg.Not); !ok {
		t.Errorf("expected Not conclusion, got %T", conc)
	}
}

// --- GoalAddPrem / GoalRemovePrem ---

func TestGoalAddPrem_Empty(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	goal := mkLF(testAstCfg.NewAtom("g"), c)
	prem := mkLF(testAstCfg.NewAtom("p"), c)
	result := GoalAddPrem(testAstCfg, goal, prem, ast.Location{})
	prems := GoalPrems(result)
	if len(prems) != 1 {
		t.Errorf("expected 1 premise, got %d", len(prems))
	}
}

func TestGoalAddPrem_Existing(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	prem1 := mkLF(testAstCfg.NewAtom("p1"), c)
	sb := testAstCfg.NewSchemaBody(prem1, c)
	goal := mkLF(testAstCfg.NewAtom("g"), sb)
	prem2 := mkLF(testAstCfg.NewAtom("p2"), c)
	result := GoalAddPrem(testAstCfg, goal, prem2, ast.Location{})
	prems := GoalPrems(result)
	if len(prems) != 2 {
		t.Errorf("expected 2 premises, got %d", len(prems))
	}
}

func TestGoalRemovePrem_Found(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	prem := mkLF(testAstCfg.NewAtom("removeme"), c)
	sb := testAstCfg.NewSchemaBody(prem, c)
	goal := mkLF(testAstCfg.NewAtom("g"), sb)

	result := GoalRemovePrem(testAstCfg, goal, "removeme")
	prems := GoalPrems(result)
	if len(prems) != 0 {
		t.Errorf("expected 0 premises after removal, got %d", len(prems))
	}
}

func TestGoalRemovePrem_NotFound(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	prem := mkLF(testAstCfg.NewAtom("keepme"), c)
	sb := testAstCfg.NewSchemaBody(prem, c)
	goal := mkLF(testAstCfg.NewAtom("g"), sb)

	result := GoalRemovePrem(testAstCfg, goal, "nonexistent")
	prems := GoalPrems(result)
	if len(prems) != 1 {
		t.Errorf("expected 1 premise unchanged, got %d", len(prems))
	}
}

// --- TrivialGoal positive case ---

func TestTrivialGoal_Positive(t *testing.T) {
	s := mkSort("S")
	c := mkConst("c", s)
	// Build inner premise goal with same conclusion as outer.
	premGoal := mkLF(testAstCfg.NewAtom("prem"), c)
	sb := testAstCfg.NewSchemaBody(premGoal, c)
	goal := mkLF(testAstCfg.NewAtom("g"), sb)

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
		s := mkSort("S")
		var node ast.Node
		switch choice % 4 {
		case 0:
			node = testAstCfg.NewConstantDecl(mkConst("f", s))
		case 1:
			node = &lg.UninterpretedSort{Name: "S"}
		case 2:
			node = mkConst("c", s)
		case 3:
			v := mkVar("X", s)
			lam, _ := lg.NewLambda([]*lg.Variable{v}, v)
			node = testAstCfg.NewConstantDecl(lam)
		}
		_ = GoalIsDefn(node) // must not panic
	})
}
