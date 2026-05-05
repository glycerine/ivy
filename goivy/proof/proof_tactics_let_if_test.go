package proof

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// Reuses mkSort, mkConst, mkLF, testAstCfg, mkPC from proof_test.go and
// tactics_assume_test.go.

// --- WrapImplies direct tests ---

func TestWrapImplies_LgExpr_ProducesLgImplies(t *testing.T) {
	s := mkSort("S")
	cond := mkConst("c", s)
	body := mkConst("p", s)

	out := WrapImplies(testAstCfg, cond, body)

	imp, ok := out.(*lg.Implies)
	if !ok {
		t.Fatalf("expected *lg.Implies for lg.Expr body, got %T", out)
	}
	if imp.T1 != cond {
		t.Errorf("T1: expected same cond pointer, got %p vs %p", imp.T1, cond)
	}
	if imp.T2 != body {
		t.Errorf("T2: expected same body pointer, got %p vs %p", imp.T2, body)
	}
}

func TestWrapImplies_TemporalModels_ProducesAstImplies(t *testing.T) {
	s := mkSort("S")
	cond := mkConst("c", s)
	inner := mkConst("phi", s)
	model := testAstCfg.NewAtom("M")
	tm := testAstCfg.NewTemporalModels(model, inner)

	out := WrapImplies(testAstCfg, cond, tm)

	imp, ok := out.(*ast.AstImplies)
	if !ok {
		t.Fatalf("expected *ast.Implies for TemporalModels body, got %T", out)
	}
	// Python wraps wholesale: T2 must be the same TemporalModels, no descent.
	if imp.T2 != ast.Node(tm) {
		t.Errorf("T2: expected same TemporalModels pointer (no descent), got %T", imp.T2)
	}
}

func TestWrapImplies_SchemaBody_ProducesAstImplies(t *testing.T) {
	s := mkSort("S")
	cond := mkConst("c", s)
	conc := mkConst("conc", s)
	sb := testAstCfg.NewSchemaBody(conc)

	out := WrapImplies(testAstCfg, cond, sb)

	imp, ok := out.(*ast.AstImplies)
	if !ok {
		t.Fatalf("expected *ast.Implies for SchemaBody body, got %T", out)
	}
	// Python wraps wholesale: T2 must be the same SchemaBody, no descent.
	if imp.T2 != ast.Node(sb) {
		t.Errorf("T2: expected same SchemaBody pointer (no descent), got %T", imp.T2)
	}
}

// --- ifTactic integration tests ---

// mkIfTactic builds an IfTactic proof with the given lg.Expr condition and
// no Then/Else branches (so the tactic returns the two wrapped subgoals).
func mkIfTactic(cond lg.Expr) *ast.IfTactic {
	return testAstCfg.NewIfTactic(cond, nil, nil)
}

func TestIfTactic_LgExprGoal_WrapsWithLgImplies(t *testing.T) {
	s := mkSort("S")
	body := mkConst("p", s)
	cond := mkConst("c", s)
	goal := mkLF(testAstCfg.NewAtom("g"), body)

	pc := mkPC()
	result, err := pc.ifTactic([]*ast.LabeledFormula{goal}, mkIfTactic(cond))
	if err != nil {
		t.Fatalf("ifTactic failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 subgoals, got %d", len(result))
	}
	trueImp, ok := result[0].Formula.(*lg.Implies)
	if !ok {
		t.Fatalf("true subgoal: expected *lg.Implies, got %T", result[0].Formula)
	}
	if trueImp.T2 != body {
		t.Errorf("true subgoal T2: expected same body pointer (no descent)")
	}
	falseImp, ok := result[1].Formula.(*lg.Implies)
	if !ok {
		t.Fatalf("false subgoal: expected *lg.Implies, got %T", result[1].Formula)
	}
	if _, isNot := falseImp.T1.(*lg.Not); !isNot {
		t.Errorf("false subgoal T1: expected *lg.Not, got %T", falseImp.T1)
	}
	if falseImp.T2 != body {
		t.Errorf("false subgoal T2: expected same body pointer")
	}
}

func TestIfTactic_TemporalModelsGoal_WrapsWholeWithoutDescent(t *testing.T) {
	s := mkSort("S")
	inner := mkConst("phi", s)
	model := testAstCfg.NewAtom("M")
	tm := testAstCfg.NewTemporalModels(model, inner)
	cond := mkConst("c", s)
	goal := mkLF(testAstCfg.NewAtom("g"), tm)

	pc := mkPC()
	result, err := pc.ifTactic([]*ast.LabeledFormula{goal}, mkIfTactic(cond))
	if err != nil {
		t.Fatalf("ifTactic failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 subgoals, got %d", len(result))
	}
	// Python ivy_proof.py:414: Implies(cond, TemporalModels(...))
	// NOT TemporalModels(model, Implies(cond, phi)) — that's the bug we fixed.
	trueImp, ok := result[0].Formula.(*ast.AstImplies)
	if !ok {
		t.Fatalf("true subgoal: expected *ast.Implies (TemporalModels at top means ast.Implies fallback), got %T", result[0].Formula)
	}
	if trueImp.T2 != ast.Node(tm) {
		t.Errorf("true subgoal T2: expected same TemporalModels pointer — if_tactic must not descend, got %T", trueImp.T2)
	}
}

func TestIfTactic_SchemaBodyGoal_WrapsWholeWithoutDescent(t *testing.T) {
	s := mkSort("S")
	conc := mkConst("conc", s)
	sb := testAstCfg.NewSchemaBody(conc)
	cond := mkConst("c", s)
	goal := mkLF(testAstCfg.NewAtom("g"), sb)

	pc := mkPC()
	result, err := pc.ifTactic([]*ast.LabeledFormula{goal}, mkIfTactic(cond))
	if err != nil {
		t.Fatalf("ifTactic failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 subgoals, got %d", len(result))
	}
	// Python wraps the entire SchemaBody in Implies, no descent.
	trueImp, ok := result[0].Formula.(*ast.AstImplies)
	if !ok {
		t.Fatalf("true subgoal: expected *ast.Implies for SchemaBody, got %T", result[0].Formula)
	}
	if trueImp.T2 != ast.Node(sb) {
		t.Errorf("true subgoal T2: expected same SchemaBody pointer (no descent), got %T", trueImp.T2)
	}
}
