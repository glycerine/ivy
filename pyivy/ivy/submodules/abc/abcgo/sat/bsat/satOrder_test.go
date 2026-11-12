package bsat

import "testing"

func TestDecisionHeapBumpsAndRemoval(t *testing.T) {
	var heap []int32
	positions := make([]int32, 100)
	activity := make([]uint64, 100)
	for v := int32(0); v < 100; v++ {
		positions[v] = int32(len(heap))
		heap = append(heap, v)
		activity[v] = uint64(v)
		satOrderUpdate(heap, positions, activity, v)
	}
	activity[0] = 1000
	satOrderUpdate(heap, positions, activity, 0)
	var v int32
	heap, v = satOrderPop(heap, positions, activity)
	if v != 0 {
		t.Fatal(v)
	}
	for expected := int32(99); expected >= 1; expected-- {
		heap, v = satOrderPop(heap, positions, activity)
		if v != expected || positions[v] != -1 {
			t.Fatal(v, expected)
		}
	}
	_, v = satOrderPop(heap, positions, activity)
	if v != var_Undef {
		t.Fatal(v)
	}
}
