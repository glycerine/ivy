package ssc

import "testing"

// In the source counter, each candidate leaving the constant class counts
// separately, even when the displaced candidates have the same signature.
func TestClassRefinementCountsConstantCandidates(t *testing.T) {
	c := &Classes{Groups: [][]int32{{0, 1, 2, 3}, {4, 5, 6}}, Repr: make([]int32, 7), Phase: make([]bool, 7)}
	s := &Simulation{Words: 1, Values: [][]uint64{{0}, {0x55}, {0x55}, {0}, {0x33}, {0x33}, {0x66}}}
	if got := Ssc_GiaClassesRefine(c, s); got != 3 {
		t.Fatalf("refinement count = %d; want 2 displaced constants plus 1 class split", got)
	}
	if c.Repr[1] != -1 || c.Repr[2] != 1 || c.Repr[3] != 0 || c.Repr[6] != -1 {
		t.Fatalf("incorrect refined representatives: %v", c.Repr)
	}
	if got := Ssc_GiaClassesRefine(c, s); got != 0 {
		t.Fatalf("stable refinement = %d", got)
	}
}
