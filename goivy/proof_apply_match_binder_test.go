package goivy

import "testing"

// Python apply_match_alt_rec maps a binder's body first (args =
// [...for f in fmla.args], env unchanged) and only then binds the
// variables (with BindSymbols(env, variables)) while mapping them. goivy
// bound the variables first, so a sort mapping t -> u did not reach the
// body of forall X:t. p(X) (test/ded1.ivy).
func TestApplyMatchAltBinderMapsBodyBeforeBinding(t *testing.T) {
	ts, us := isolateMkSort("t"), isolateMkSort("u")
	pSort, _ := NewFunctionSort(ts, Boolean)
	p := NewConst("p", pSort)
	x, _ := NewVariable("X", ts)
	fa := &ForAll{Variables: []*LogicVariable{x}, Body: MustApply(p, x)}
	got := applyMatchAltRec(map[NodeKey]Expr{SortKey(ts): us}, fa, map[NodeKey]bool{})
	g, ok := got.(*ForAll)
	if !ok {
		t.Fatalf("%T", got)
	}
	app := g.Body.(*Apply)
	if SortName(g.Variables[0].VSort) != "u" || SortName(app.Terms[0].NodeSort()) != "u" {
		t.Fatalf("got %v, want forall X:u. p(X:u)", got)
	}
}
