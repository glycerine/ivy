package core

import "testing"

func TestCheckConsistencyAcrossBDDAndZDDReorders(t *testing.T) {
	dd := NewBDDManager(4)
	x := Cudd_bddXor(dd, dd.Vars[0], dd.Vars[3])
	Cudd_Ref(x)
	if err := Cudd_CheckConsistency(dd); err != nil {
		t.Fatalf("fresh graph inconsistent: %v", err)
	}
	if Cudd_ShuffleHeap(dd, []int{3, 1, 2, 0}) != 1 {
		t.Fatal("shuffle failed")
	}
	if err := Cudd_CheckConsistency(dd); err != nil {
		t.Fatalf("reordered graph inconsistent: %v", err)
	}
}
