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

func TestReorderingPreservesReferenceOwnership(t *testing.T) {
	for table := 1; table < 256; table++ {
		dd := NewBDDManager(3)
		root := dd.LogicZero
		Cudd_Ref(root)
		for mask := 0; mask < 8; mask++ {
			if table&(1<<mask) == 0 {
				continue
			}
			term := dd.One
			Cudd_Ref(term)
			for i, v := range dd.Vars {
				next := Cudd_bddAnd(dd, term, Cudd_NotCond(v, mask&(1<<i) == 0))
				Cudd_Ref(next)
				Cudd_RecursiveDeref(dd, term)
				term = next
			}
			next := Cudd_bddOr(dd, root, term)
			Cudd_Ref(next)
			Cudd_RecursiveDeref(dd, root)
			Cudd_RecursiveDeref(dd, term)
			root = next
		}
		if Cudd_ReduceHeap(dd, CUDD_REORDER_SYMM_SIFT, 1) == 0 {
			t.Fatal("reordering failed")
		}
		Cudd_RecursiveDeref(dd, root)
		if count := Cudd_CheckZeroRef(dd); count != 0 {
			t.Fatalf("table %d: reordering left %d external roots after releasing the sole owner", table, count)
		}
		Cudd_Quit(dd)
	}
}
