package msat

import "testing"

func TestCancelUntilDoesNotClearQueueWhenNoLevelIsCanceled(t *testing.T) {
	p := New(1)
	Msat_QueueInsert(p.pQueue, 0)
	Msat_SolverCancelUntil(p, 0)
	if Msat_QueueReadSize(p.pQueue) != 1 {
		t.Fatal("C leaves queue intact when no level is canceled")
	}
}
func TestAssumeRequiresEmptyPropagationQueue(t *testing.T) {
	p := New(2)
	Msat_QueueInsert(p.pQueue, 0)
	defer func() {
		if recover() == nil {
			t.Fatal("C asserts empty propagation queue before assuming")
		}
	}()
	Msat_SolverAssume(p, 2)
}
