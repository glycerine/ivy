package msat

import "testing"

func TestBoundarySelectUsesOnlyBoundaryAndDoesNotRemove(t *testing.T) {
	p := New(2)
	q := NewBoundaryOrder(p)
	if got := q.Select(); got != MSAT_ORDER_UNKNOWN {
		t.Fatalf("empty C boundary = unknown, got %d", got)
	}
	q.present[1] = true
	q.variables = append(q.variables, 1)
	if q.Select() != 1 || q.Select() != 1 {
		t.Fatal("C selection leaves entry in boundary until assignment")
	}
}
func TestBoundaryUnassignDropsNeighborsWithNoAssignedNeighbor(t *testing.T) {
	p := New(2)
	p.vAdjacents[0].data = []int{1}
	p.vAdjacents[1].data = []int{0}
	q := NewBoundaryOrder(p)
	q.present[1] = true
	q.variables = append(q.variables, 1)
	q.Unassigned(0)
	if q.present[1] || len(q.variables) != 0 {
		t.Fatal("C removes boundary neighbors without assigned neighbors")
	}
}
