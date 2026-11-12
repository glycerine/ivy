package core

import "testing"

func TestExactReorderSearchAndRangeLimit(t *testing.T) {
	dd := NewBDDManager(3)
	x0, x1, x2 := dd.Vars[0], dd.Vars[1], dd.Vars[2]
	f := Cudd_bddOr(dd, Cudd_bddAnd(dd, x0, x1), Cudd_bddAnd(dd, x0, x2))
	Cudd_Ref(f)
	if !Cudd_ExactReorder(dd, 0, 2) {
		t.Fatal("exact search rejected a three-variable range")
	}
	if Cudd_bddAnd(dd, f, Cudd_Not(Cudd_bddOr(dd, Cudd_bddAnd(dd, x0, x1), Cudd_bddAnd(dd, x0, x2)))) != dd.LogicZero {
		t.Fatal("reordering changed the represented function")
	}
	large := NewBDDManager(9)
	before := append([]int(nil), large.InvPerm...)
	if Cudd_ExactReorder(large, 0, 8) {
		t.Fatal("exact search should reject a range above its work bound")
	}
	for i := range before {
		if large.InvPerm[i] != before[i] {
			t.Fatal("rejected exact search changed the manager")
		}
	}
}
