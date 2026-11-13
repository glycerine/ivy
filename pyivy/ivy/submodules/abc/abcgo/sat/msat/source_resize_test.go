package msat

import "testing"

func TestResizeClearsOrderAndQueueAndResetsConeMask(t *testing.T) {
	p := New(2)
	p.vVarsUsed.data[0] = 0
	Msat_QueueInsert(p.pQueue, 2)
	Msat_SolverResize(p, 4)
	if len(p.pOrder.heap) != 0 {
		t.Fatalf("C resize clears order: %v", p.pOrder.heap)
	}
	if got := Msat_QueueExtract(p.pQueue); got >= 0 {
		t.Fatalf("C resize replaces queue, got %d", got)
	}
	for i, used := range p.vVarsUsed.data {
		if used != 1 {
			t.Fatalf("cone mask[%d]=%d, want 1", i, used)
		}
	}
}
