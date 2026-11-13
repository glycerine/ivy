package bsat

import "testing"

func TestStopCallbackBridge(t *testing.T) {
	s := NewSolver()
	s.NewVariable()
	var seen []int32
	s.SetStopFunc(17, func(id int32) int32 { seen = append(seen, id); return 0 })
	if s.Solve() != Satisfiable || len(seen) != 1 || seen[0] != 17 {
		t.Fatalf("callback did not receive run ID: %v", seen)
	}
	s.SetStopFunc(22, nil)
	if s.Solve() != Satisfiable || len(seen) != 1 {
		t.Fatal("callback removal failed")
	}
}
