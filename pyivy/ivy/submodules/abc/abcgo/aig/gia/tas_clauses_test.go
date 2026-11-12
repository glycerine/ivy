package gia

import "testing"

func TestTasClauseHandlesSurviveGrowthAndReset(t *testing.T) {
	var store Tas_Sto_t
	first := store.allocate(3)
	copy(first.literals(), []int32{2, 5, 8})
	first.setNext(1, 29)
	for range 1000 {
		store.allocate(7)
	}
	reloaded := store.clause(first.handle)
	if reloaded.next(1) != 29 || reloaded.literals()[1] != 5 {
		t.Fatal("clause handle changed after growth")
	}
	store.reset()
	second := store.allocate(1)
	if second.handle != 16 || second.literals()[0] != 0 || second.next(1) != 0 {
		t.Fatal("reset did not clear the first clause")
	}
}
