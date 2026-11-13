package fraig

import "testing"

func TestSATEquivalenceAndImplication(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{nPatsRand: 128, nPatsDyna: 128, fFuncRed: 0})
	defer Fraig_ManFree(p)
	a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
	left := Fraig_NodeAnd(p, Fraig_NodeAnd(p, a, b), c)
	right := Fraig_NodeAnd(p, a, Fraig_NodeAnd(p, b, c))
	if left == right {
		t.Fatal("expected structurally different nodes")
	}
	if Fraig_NodeIsEquivalent(p, left, right, -1, 5) != 1 {
		t.Fatal("associative nodes should be equivalent")
	}
	if Fraig_NodeIsImplication(p, left, a, -1) != 1 {
		t.Fatal("AND implies input")
	}
}
