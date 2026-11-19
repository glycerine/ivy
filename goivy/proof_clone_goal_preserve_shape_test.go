package goivy

import "testing"

// Python apply_match_goal rebuilds a schema goal with
// fmla.clone(prems+[conc]): the SchemaBody stays even when no premises are
// left (unlike clone_goal). goivy's CloneGoalPreserveID collapsed it to the
// bare conclusion (test/assumetactic1.ivy).
func TestCloneGoalPreserveIDKeepsSchemaBody(t *testing.T) {
	cfg := NewAstConfig()
	c := NewConst("c", Boolean)
	goal := cfg.NewLabeledFormula(cfg.NewAtom("foo"), cfg.NewSchemaBody(NewConst("v", Boolean), c))
	got := CloneGoalPreserveID(cfg, goal, nil, c)
	sb, ok := got.Formula.(*SchemaBody)
	if !ok || len(sb.Elems) != 1 || Key(sb.Elems[0].(Expr)) != Key(c) {
		t.Fatalf("formula = %#v, want SchemaBody(c)", got.Formula)
	}
	plain := cfg.NewLabeledFormula(cfg.NewAtom("bar"), c)
	if _, ok := CloneGoalPreserveID(cfg, plain, nil, c).Formula.(*SchemaBody); ok {
		t.Fatal("a non-schema goal must stay a plain formula")
	}
}
