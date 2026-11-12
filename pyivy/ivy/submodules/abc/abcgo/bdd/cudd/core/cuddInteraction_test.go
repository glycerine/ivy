package core

import "testing"

func TestInteractionMatrixTracksSharedSupport(t *testing.T) {
	dd := NewBDDManager(3)
	x, y, z := Cudd_bddIthVar(dd, 0), Cudd_bddIthVar(dd, 1), Cudd_bddIthVar(dd, 2)
	if Cudd_bddAnd(dd, x, y) == nil {
		t.Fatal("failed to construct shared-support node")
	}
	Cudd_ComputeInteractions(dd)
	if !Cudd_TestInteraction(dd, 0, 1) {
		t.Fatal("support variables should interact")
	}
	if Cudd_TestInteraction(dd, 0, 2) || Cudd_TestInteraction(dd, 1, 2) {
		t.Fatal("independent variable should not interact")
	}
	Cudd_SetInteraction(dd, 0, 2)
	if !Cudd_TestInteraction(dd, 2, 0) {
		t.Fatal("interaction relation should be symmetric")
	}
	_ = z
}
