package fraig

import "testing"

func TestNodesAreEqualHonorsEdgePhase(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
	defer Fraig_ManFree(p)
	a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
	left := Fraig_NodeAnd(p, Fraig_NodeAnd(p, a, b), c)
	right := Fraig_NodeAnd(p, a, Fraig_NodeAnd(p, b, c))
	if Fraig_NodesAreEqual(p, left, right, -1, 5) != 1 {
		t.Fatal("equivalent nodes not equal")
	}
	if Fraig_NodesAreEqual(p, left, Fraig_Not(right), -1, 5) != 0 {
		t.Fatal("complemented edge reported equal")
	}
}
