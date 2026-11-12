package core

import (
	"reflect"
	"testing"
)

func TestCrossManagerTransferFamiliesAndPermutation(t *testing.T) {
	source, destination := NewBDDManager(3), NewBDDManager(0)
	bdd := Cudd_Not(Cudd_bddIte(source, source.Vars[2], source.Vars[1], source.Vars[0]))
	add := Cudd_addIte(source, Cudd_addIthVar(source, 0), Cudd_addConst(source, -7), Cudd_addConst(source, 2.5))
	for _, root := range []*DdNode{bdd, add} {
		copied := source.Transfer(root, destination, nil)
		if copied == nil || !reflect.DeepEqual(reorderTruth(t, source, root), reorderTruth(t, destination, copied)) {
			t.Fatal("identity transfer changed function")
		}
		if roundtrip := destination.Transfer(copied, source, nil); roundtrip != root {
			t.Fatal("roundtrip lost canonical identity")
		}
		perm := []int{2, 0, 1}
		mapped := source.Transfer(root, destination, perm)
		inverse := []int{1, 2, 0}
		if destination.Transfer(mapped, source, inverse) != root {
			t.Fatal("permuted transfer roundtrip")
		}
	}
	if Cudd_bddTransfer(source, destination, add) != nil {
		t.Fatal("Boolean transfer accepted ADD")
	}
	if source.Transfer(bdd, destination, []int{0}) != nil || source.Transfer(bdd, destination, []int{-1, 0, 1}) != nil {
		t.Fatal("invalid transfer map accepted")
	}
	Cudd_Quit(destination)
	if source.Transfer(bdd, destination, nil) != nil {
		t.Fatal("closed destination accepted")
	}
}
