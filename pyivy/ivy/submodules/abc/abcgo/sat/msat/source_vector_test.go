package msat

import "testing"

func TestUniqueOrderStillMovesLastEntryForDuplicate(t *testing.T) {
	v := &Msat_IntVec_t{data: []int{3, 2, 1}}
	Msat_IntVecPushUniqueOrder(v, 3, 1)
	if len(v.data) != 3 || v.data[0] != 1 || v.data[1] != 3 || v.data[2] != 2 {
		t.Fatalf("C orders last entry even when duplicate was not inserted: %v", v.data)
	}
}

func TestAdoptedAndCopiedVectorsPreserveSourceCapacity(t *testing.T) {
	backing := []int{1, 2, 99, 100}
	adopted := Msat_IntVecAllocArray(backing, 2)
	copied := Msat_IntVecAllocArrayCopy(backing, 2)
	dup := Msat_IntVecDup(adopted)
	if cap(adopted.data) != 2 || cap(copied.data) != 2 || cap(dup.data) != 2 {
		t.Fatalf("C capacities are 2: adopted=%d copied=%d dup=%d", cap(adopted.data), cap(copied.data), cap(dup.data))
	}
	Msat_IntVecPush(adopted, 3)
	if backing[2] != 99 {
		t.Fatal("C reallocates adopted array when logical capacity is reached")
	}
}
