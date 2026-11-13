package msat

import "testing"

func TestOrderCleanCopiesConeWithoutHeapifying(t *testing.T) {
	p := New(3)
	p.pdActivity[0], p.pdActivity[1], p.pdActivity[2] = 0, 10, 20
	Msat_OrderClean(p.pOrder, &Msat_IntVec_t{data: []int{0, 1, 2}})
	for i, want := range []int{0, 1, 2} {
		if p.pOrder.heap[i] != want {
			t.Fatalf("heap after clean = %v", p.pOrder.heap)
		}
	}
}

func TestOrderSelectAndUnassignDoNotFilterConeMask(t *testing.T) {
	p := New(1)
	p.vVarsUsed.data[0] = 0
	if got := Msat_OrderVarSelect(p.pOrder); got != 0 {
		t.Fatalf("C selects unassigned heap entry regardless of mask, got %d", got)
	}
	Msat_OrderVarUnassigned(p.pOrder, 0)
	if got := Msat_OrderVarSelect(p.pOrder); got != 0 {
		t.Fatalf("C reinserts unassigned heap entry regardless of mask, got %d", got)
	}
}
