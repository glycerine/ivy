package core

import "testing"

func TestCheckZeroRefCountsExternalRoots(t *testing.T) {
	dd := NewBDDManager(2)
	if got := Cudd_CheckZeroRef(dd); got != 0 {
		t.Fatalf("initial external-reference count = %d, want 0", got)
	}
	node := Cudd_bddIthVar(dd, 0)
	Cudd_Ref(node)
	if got := Cudd_CheckZeroRef(dd); got != 1 {
		t.Fatalf("count with one externally rooted node = %d, want 1", got)
	}
	Cudd_RecursiveDeref(dd, node)
	if got := Cudd_CheckZeroRef(dd); got != 0 {
		t.Fatalf("count after releasing the external root = %d, want 0", got)
	}
}
