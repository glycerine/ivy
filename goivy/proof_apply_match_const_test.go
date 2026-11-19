package goivy

import "testing"

// Python apply_match_alt_rec treats a constant as an application with no
// arguments: if fmla.rep in match: return apply_fun(match_get(...), []),
// which calls a matched Lambda([], body) to get body. goivy returned the
// lambda itself, so it later appeared as a term ("lambda . X") and
// NewApply failed (test/ded1.ivy panicked).
func TestApplyMatchAltConstAppliesNullaryLambda(t *testing.T) {
	ts := isolateMkSort("t")
	p := NewConst("p", ts)
	x := NewConst("x", ts)
	lam, err := NewLambda(nil, x)
	if err != nil {
		t.Fatal(err)
	}
	got := applyMatchAltRec(map[NodeKey]Expr{Key(p): lam}, p, map[NodeKey]bool{})
	if Key(got) != Key(x) {
		t.Fatalf("got %v (%T), want x", got, got)
	}
}
