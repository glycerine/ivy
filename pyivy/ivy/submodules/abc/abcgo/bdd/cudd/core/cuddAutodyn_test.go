package core

import (
	"testing"
	"time"
)

// TestAutoReorderMakesProgressBelowMinsize: with a low threshold the
// automatic reordering can be skipped (heap below Cudd_ReduceHeap's
// minsize); the operation must still finish instead of retrying forever.
func TestAutoReorderMakesProgressBelowMinsize(t *testing.T) {
	dd := NewBDDManager(3)
	Cudd_AutodynEnable(dd, CUDD_REORDER_SYMM_SIFT)
	Cudd_SetNextReordering(dd, 1)
	Cudd_AutodynSafeBegin(dd)
	defer Cudd_AutodynSafeEnd(dd)
	done := make(chan *DdNode, 1)
	go func() { done <- Cudd_bddAnd(dd, dd.Vars[0], Cudd_bddOr(dd, dd.Vars[1], dd.Vars[2])) }()
	select {
	case r := <-done:
		if r == nil {
			t.Fatal("no result")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("automatic reordering retried forever")
	}
}
