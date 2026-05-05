package goivy

import (
	"testing"
)

// Reuses proofMkSort, proofMkConst, mkLF, proofTestAstCfg, mkPC from proof_test.go and
// tactics_assume_test.go.

// --- WrapImplies direct tests ---

func TestWrapImplies_LgExpr_ProducesLgImplies(t *testing.T) {
	s := proofMkSort("S")
	cond := proofMkConst("c", s)
	body := proofMkConst("p", s)

	out := WrapImplies(proofTestAstCfg, cond, body)

	imp, ok := out.(*Implies)
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
	s := proofMkSort("S")
	cond := proofMkConst("c", s)
	inner := proofMkConst("phi", s)
	model := proofTestAstCfg.NewAtom("M")
	tm := proofTestAstCfg.NewTemporalModels(model, inner)

	out := WrapImplies(proofTestAstCfg, cond, tm)

	imp, ok := out.(*AstImplies)
	if !ok {
		t.Fatalf("expected *ast.Implies for TemporalModels body, got %T", out)
	}
	// Python wraps wholesale: T2 must be the same TemporalModels, no descent.
	if imp.T2 != Node(tm) {
		t.Errorf("T2: expected same TemporalModels pointer (no descent), got %T", imp.T2)
	}
}

func TestWrapImplies_SchemaBody_ProducesAstImplies(t *testing.T) {
	s := proofMkSort("S")
	cond := proofMkConst("c", s)
	conc := proofMkConst("conc", s)
	sb := proofTestAstCfg.NewSchemaBody(conc)

	out := WrapImplies(proofTestAstCfg, cond, sb)

	imp, ok := out.(*AstImplies)
	if !ok {
		t.Fatalf("expected *ast.Implies for SchemaBody body, got %T", out)
	}
	// Python wraps wholesale: T2 must be the same SchemaBody, no descent.
	if imp.T2 != Node(sb) {
		t.Errorf("T2: expected same SchemaBody pointer (no descent), got %T", imp.T2)
	}
}

// --- ifTactic integration tests ---

// mkIfTactic builds an IfTactic proof with the given lg.Expr condition and
// no Then/Else branches (so the tactic returns the two wrapped subgoals).
func mkIfTactic(cond Expr) *IfTactic {
	return proofTestAstCfg.NewIfTactic(cond, nil, nil)
}

func TestIfTactic_LgExprGoal_WrapsWithLgImplies(t *testing.T) {
	s := proofMkSort("S")
	body := proofMkConst("p", s)
	cond := proofMkConst("c", s)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), body)

	pc := mkPC()
	result, err := pc.ifTactic([]*LabeledFormula{goal}, mkIfTactic(cond))
	if err != nil {
		t.Fatalf("ifTactic failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 subgoals, got %d", len(result))
	}
	trueImp, ok := result[0].Formula.(*Implies)
	if !ok {
		t.Fatalf("true subgoal: expected *lg.Implies, got %T", result[0].Formula)
	}
	if trueImp.T2 != body {
		t.Errorf("true subgoal T2: expected same body pointer (no descent)")
	}
	falseImp, ok := result[1].Formula.(*Implies)
	if !ok {
		t.Fatalf("false subgoal: expected *lg.Implies, got %T", result[1].Formula)
	}
	if _, isNot := falseImp.T1.(*Not); !isNot {
		t.Errorf("false subgoal T1: expected *lg.Not, got %T", falseImp.T1)
	}
	if falseImp.T2 != body {
		t.Errorf("false subgoal T2: expected same body pointer")
	}
}

func TestIfTactic_TemporalModelsGoal_WrapsWholeWithoutDescent(t *testing.T) {
	s := proofMkSort("S")
	inner := proofMkConst("phi", s)
	model := proofTestAstCfg.NewAtom("M")
	tm := proofTestAstCfg.NewTemporalModels(model, inner)
	cond := proofMkConst("c", s)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), tm)

	pc := mkPC()
	result, err := pc.ifTactic([]*LabeledFormula{goal}, mkIfTactic(cond))
	if err != nil {
		t.Fatalf("ifTactic failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 subgoals, got %d", len(result))
	}
	// Python ivy_proof.py:414: Implies(cond, TemporalModels(...))
	// NOT TemporalModels(model, Implies(cond, phi)) — that's the bug we fixed.
	trueImp, ok := result[0].Formula.(*AstImplies)
	if !ok {
		t.Fatalf("true subgoal: expected *ast.Implies (TemporalModels at top means ast.Implies fallback), got %T", result[0].Formula)
	}
	if trueImp.T2 != Node(tm) {
		t.Errorf("true subgoal T2: expected same TemporalModels pointer — if_tactic must not descend, got %T", trueImp.T2)
	}
}

func TestIfTactic_SchemaBodyGoal_WrapsWholeWithoutDescent(t *testing.T) {
	s := proofMkSort("S")
	conc := proofMkConst("conc", s)
	sb := proofTestAstCfg.NewSchemaBody(conc)
	cond := proofMkConst("c", s)
	goal := mkLF(proofTestAstCfg.NewAtom("g"), sb)

	pc := mkPC()
	result, err := pc.ifTactic([]*LabeledFormula{goal}, mkIfTactic(cond))
	if err != nil {
		t.Fatalf("ifTactic failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 subgoals, got %d", len(result))
	}
	// Python wraps the entire SchemaBody in Implies, no descent.
	trueImp, ok := result[0].Formula.(*AstImplies)
	if !ok {
		t.Fatalf("true subgoal: expected *ast.Implies for SchemaBody, got %T", result[0].Formula)
	}
	if trueImp.T2 != Node(sb) {
		t.Errorf("true subgoal T2: expected same SchemaBody pointer (no descent), got %T", trueImp.T2)
	}
}
