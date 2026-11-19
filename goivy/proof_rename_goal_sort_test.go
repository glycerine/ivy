package goivy

import "testing"

// Python rename_goal renames every symbol of goal_defns whose name is in
// rmap, sorts included (x.rename on an UninterpretedSort): "apply
// nonce<u/t>" turns type t into u and x:t into x:u. goivy's
// nodeNameStr/renameNode ignored sorts (test/ded1.ivy).
func TestRenameGoalRenamesSort(t *testing.T) {
	cfg := NewAstConfig()
	ts := &UninterpretedSort{Name: "t"}
	x := NewConst("x", ts)
	goal := cfg.NewLabeledFormula(cfg.NewAtom("nonce"), cfg.NewSchemaBody(ts, cfg.NewConstantDecl(x), True))
	ren := cfg.NewRenaming([]Node{cfg.NewDefinition(cfg.NewAtom("t"), cfg.NewAtom("u"))})
	got, err := RenameGoal(cfg, goal, ren)
	if err != nil {
		t.Fatal(err)
	}
	prems := GoalPrems(got)
	s, ok := unwrapCompiledNode(prems[0]).(*UninterpretedSort)
	if !ok || s.Name != "u" {
		t.Fatalf("sort premise = %v, want u", prems[0])
	}
	cd := unwrapCompiledNode(prems[1]).(*ConstantDecl)
	if c, ok := unwrapCompiledNode(cd.Args()[0]).(*Const); !ok || SortName(c.CSort) != "u" {
		t.Fatalf("decl = %v, want x:u", prems[1])
	}
}
