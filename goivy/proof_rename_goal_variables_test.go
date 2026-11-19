package goivy

import "testing"

// Python rename_goal / check_renaming build rmap from x.lhs().rep and
// x.rhs().rep, which works for variables too ("apply elimA<Z/Y>" renames
// the bound Y). goivy accepted only Atoms, so a variable renaming was
// silently ignored (test/ded1.ivy).
func TestRenameGoalRenamesBoundVariable(t *testing.T) {
	cfg := NewAstConfig()
	ts := isolateMkSort("t")
	pSort, _ := NewFunctionSort(ts, Boolean)
	p := NewConst("p", pSort)
	y, _ := NewVariable("Y", ts)
	prem := cfg.NewLabeledFormula(cfg.NewAtom("prem"), &ForAll{Variables: []*LogicVariable{y}, Body: MustApply(p, y)})
	goal := cfg.NewLabeledFormula(cfg.NewAtom("g"), cfg.NewSchemaBody(prem, True))
	ren := cfg.NewRenaming([]Node{cfg.NewDefinition(cfg.NewVariable("Y", ""), cfg.NewVariable("Z", ""))})
	got, err := RenameGoal(cfg, goal, ren)
	if err != nil {
		t.Fatal(err)
	}
	fa, ok := GoalPrems(got)[0].(*LabeledFormula).Formula.(*ForAll)
	if !ok || fa.Variables[0].Name != "Z" {
		t.Fatalf("prem = %v, want forall Z. p(Z)", GoalPrems(got)[0])
	}
}
