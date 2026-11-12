package core

import "testing"

func TestGeneticReorderPreservesFunctionsAndRange(t *testing.T) {
	dd := NewBDDManager(4)
	x := Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[3])
	Cudd_Ref(x)
	if !Cudd_GeneticReorder(dd, 1, 2) {
		t.Fatal("genetic reordering failed")
	}
	expected := Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[3])
	if Cudd_bddXor(dd, x, expected) != dd.LogicZero {
		t.Fatal("reordering changed function semantics")
	}
	if dd.InvPerm[0] != 0 || dd.InvPerm[3] != 3 {
		t.Fatal("reordered range changed variables outside its bounds")
	}
}
