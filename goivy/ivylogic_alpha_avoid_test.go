package goivy

import "testing"

// Python alpha_avoid reserves the names of set(vs) | free_variables(fmla),
// each distinct variable once; a free variable also in vs must not be
// reserved twice (which would reserve X_a too and rename a bound X to
// X_a_a instead of X_a; test/naming1.ivy).
func TestAlphaAvoidReservesEachVariableOnce(t *testing.T) {
	s := &UninterpretedSort{Name: "t"}
	x := &LogicVariable{Name: "X", VSort: s}
	p := NewConst("p", LogicRelationSort([]Sort{s, s}))
	body, _ := NewApply(p, x, &LogicVariable{Name: "X", VSort: s})
	// free X plus a bound X: forall X. p(X_free, X_bound) is awkward, so
	// use: p(X, X') & exists X. p(X, X)
	inner, _ := NewApply(p, x, x)
	fmla := &LogicAnd{Terms: []Expr{body, IvyExists([]*LogicVariable{x}, inner)}}
	vs := map[NodeKey]Expr{Key(x): x}
	got := AlphaAvoidMap(fmla, vs)
	ex := got.(*LogicAnd).Terms[1].(*LogicExists)
	if n := ex.Variables[0].Name; n != "X_a" {
		t.Fatalf("bound X renamed to %q, want X_a", n)
	}
	got2 := AlphaAvoid(fmla, []*LogicVariable{x, x})
	if n := got2.(*LogicAnd).Terms[1].(*LogicExists).Variables[0].Name; n != "X_a" {
		t.Fatalf("list form: bound X renamed to %q, want X_a", n)
	}
}
