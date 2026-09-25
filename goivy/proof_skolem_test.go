package goivy

import "testing"

func TestVarSubstGoalConstantDeclPassthrough(t *testing.T) {
	cfg := proofTestAstCfg
	sortS := proofMkSort("S")
	declSym := proofMkConst("declared", sortS)
	cd := cfg.NewConstantDecl(declSym)

	x := proofMkVar("X", sortS)
	replacement := proofMkConst("sk_X", sortS)
	subs := map[NodeKey]Expr{Key(x): replacement}

	got, err := varSubstGoalNode(cfg, cd, subs)
	if err != nil {
		t.Fatalf("varSubstGoalNode(ConstantDecl) returned error: %v", err)
	}
	gotCD, ok := got.(*ConstantDecl)
	if !ok {
		t.Fatalf("varSubstGoalNode(ConstantDecl) = %T, want *ConstantDecl", got)
	}
	if gotCD != cd {
		t.Fatalf("varSubstGoalNode(ConstantDecl) should return the original declaration")
	}

	goal := mkLF(cfg.NewAtom("goal"), cfg.NewSchemaBody(cd, x))
	substGoal, err := varSubstGoal(cfg, goal, subs)
	if err != nil {
		t.Fatalf("varSubstGoal with ConstantDecl premise returned error: %v", err)
	}
	prems := GoalPrems(substGoal)
	if len(prems) != 1 || prems[0] != cd {
		t.Fatalf("ConstantDecl premise was not preserved: %#v", prems)
	}
	if GoalConc(substGoal) != replacement {
		t.Fatalf("goal conclusion was not substituted: got %#v want %#v", GoalConc(substGoal), replacement)
	}
}
