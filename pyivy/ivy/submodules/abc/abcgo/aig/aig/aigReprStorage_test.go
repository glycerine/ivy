package aig

import "testing"

func TestRepresentativeStorageQuery(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	if Aig_ManHasReprs(g) {
		t.Fatal("new graph already has representatives")
	}
	Aig_ManReprStart(g, 10)
	Aig_ObjSetRepr(g, a, g.Const1())
	if !Aig_ManHasReprs(g) || Aig_ObjRepr(g, a) != g.Const1() {
		t.Fatal("query changed representatives")
	}
	Aig_ManReprStop(g)
	if Aig_ManHasReprs(g) {
		t.Fatal("representatives still allocated")
	}
}
