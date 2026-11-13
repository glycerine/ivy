package bsat

import "testing"

func TestSetPivotVariablesOwnsCopy(t *testing.T) {
	s := NewSolver()
	for i := 0; i < 3; i++ {
		s.NewVariable()
	}
	pivots := []int32{1, 2}
	s.SetPivotVariables(pivots)
	pivots[0] = 0
	got := veci_begin(&s.pivot_vars)
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("pivots %v", got)
	}
}
