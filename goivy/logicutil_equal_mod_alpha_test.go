package goivy

import "testing"

// Python logic_util.equal_mod_alpha keys its pushable_map by the Var value
// (name and sort); goivy keyed it by *LogicVariable pointer, so a bound
// body variable that is a different object from the binder's (equal by
// value) missed the map, and exists X. p(X) did not equal itself
// (test/ded3.ivy: "conclusions do not match").
func TestEqualModAlphaKeysBoundVariablesByValue(t *testing.T) {
	ts := isolateMkSort("t")
	pSort, err := NewFunctionSort(ts, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	p := NewConst("p", pSort)
	mk := func(n string) *LogicVariable {
		v, err := NewVariable(n, ts)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	e1 := &LogicExists{Variables: []*LogicVariable{mk("X")}, Body: MustApply(p, mk("X"))}
	x2 := mk("X")
	e2 := &LogicExists{Variables: []*LogicVariable{x2}, Body: MustApply(p, x2)}
	if !EqualModAlpha(e1, e2) || !EqualModAlpha(e2, e1) {
		t.Fatal("exists X. p(X) should equal exists X. p(X)")
	}
	y := mk("Y")
	if !EqualModAlpha(e1, &LogicExists{Variables: []*LogicVariable{y}, Body: MustApply(p, y)}) {
		t.Fatal("exists X. p(X) should equal exists Y. p(Y)")
	}
	if EqualModAlpha(e1, &LogicExists{Variables: []*LogicVariable{y}, Body: MustApply(p, mk("X"))}) {
		t.Fatal("exists X. p(X) must not equal exists Y. p(X) (X free)")
	}
}
