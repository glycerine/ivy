package aig

import "testing"

func TestSetInputPhase(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	Aig_ObjSetPhase(g, a, true)
	if !Aig_ObjPhase(g, a) || Aig_ObjPhaseReal(g, a.Not()) {
		t.Fatal("input phase was not preserved through complemented edge")
	}
	Aig_ObjSetPhase(g, a, false)
	if Aig_ObjPhase(g, a) || !Aig_ObjPhaseReal(g, a.Not()) {
		t.Fatal("input phase was not cleared")
	}
}
