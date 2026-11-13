package msat

import "testing"

func TestUniqueOrderStillMovesLastEntryForDuplicate(t *testing.T) {
	v := &Msat_IntVec_t{data: []int{3, 2, 1}}
	Msat_IntVecPushUniqueOrder(v, 3, 1)
	if len(v.data) != 3 || v.data[0] != 1 || v.data[1] != 3 || v.data[2] != 2 {
		t.Fatalf("C orders last entry even when duplicate was not inserted: %v", v.data)
	}
}
