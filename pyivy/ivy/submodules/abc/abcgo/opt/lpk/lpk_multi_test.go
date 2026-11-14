package lpk

import "testing"

// This locks down Lpk_CreateCommonOrder's score and tie semantics from
// lpkMulti.c: cofactored variables keep priority 16, ties share a priority,
// and the first remaining maximum is assigned before lower scores.
func TestLpkCreateCommonOrderSkipsCofactorAndGroupsTies(t *testing.T) {
	var table [16][16]byte
	// Variable 0 has score +3, variable 1 score -3, and variable 2 score 0.
	// Cofactor variable 0; score calculations still include its table entries,
	// matching the C loop over all nVars columns and rows.
	table[0][1] = 3
	var priorities [16]int32
	Lpk_CreateCommonOrder(table[:], []int32{0}, 1, priorities[:], 3, 0)
	want := [3]int32{16, 2, 1}
	for i := range want {
		if priorities[i] != want[i] {
			t.Errorf("priority[%d]=%d, want %d", i, priorities[i], want[i])
		}
	}
}

func TestLpkCreateCommonOrderAssignsSamePriorityToScoreTies(t *testing.T) {
	var priorities [16]int32
	Lpk_CreateCommonOrder(make([][16]byte, 3), nil, 0, priorities[:], 3, 0)
	for i := 0; i < 3; i++ {
		if priorities[i] != 1 {
			t.Errorf("priority[%d]=%d for tied zero scores, want 1", i, priorities[i])
		}
	}
}
