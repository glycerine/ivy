package goivy

import "testing"

// Python apply_match_goal maps a declaration premise with
// apply_match_func_alt, which looks the symbol up in the match first
// (if func in match: return match[func]) and only then under its
// sort-mapped name. When the match also renames a sort, goivy looked up
// only the sort-mapped symbol, missed the lambda for the declared
// function, and kept a premise Python drops (test/ded1.ivy).
func TestApplyMatchGoalDeclUsesOriginalSymbolFirst(t *testing.T) {
	cfg := NewAstConfig()
	ts, us := isolateMkSort("t"), isolateMkSort("u")
	fs, _ := NewFunctionSort(ts, ts)
	f := NewConst("f", fs)
	x, _ := NewVariable("X", ts)
	lam, _ := NewLambda([]*LogicVariable{x}, x)
	conc := NewConst("c", Boolean)
	goal := cfg.NewLabeledFormula(cfg.NewAtom("g"), cfg.NewSchemaBody(cfg.NewConstantDecl(f), conc))
	match := map[NodeKey]Expr{Key(f): lam, SortKey(ts): us}
	got := applyMatchGoalNode(cfg, match, goal, false)
	if n := len(GoalPrems(got)); n != 0 {
		t.Fatalf("premises = %d (%v), want the lambda declaration dropped", n, GoalPrems(got))
	}
}
