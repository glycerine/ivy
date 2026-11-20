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

func TestCheckConsistencyRejectsBrokenCanonicalPairs(t *testing.T) {
	dd := NewBDDManager(2)
	defer Cudd_Quit(dd)
	for _, node := range []*DdNode{dd.Vars[0], dd.One} {
		original := node.inverse
		node.inverse = Cudd_Not(dd.Vars[1])
		if err := Cudd_CheckConsistency(dd); err == nil {
			t.Errorf("consistency check accepted node %d pointing to a different node's inverse", node.ID)
		}
		node.inverse = nil
		if err := Cudd_CheckConsistency(dd); err == nil {
			t.Errorf("consistency check accepted missing inverse for node %d", node.ID)
		}
		node.inverse = original
	}
	if err := Cudd_CheckConsistency(dd); err != nil {
		t.Fatalf("restored graph failed validation: %v", err)
	}
}
