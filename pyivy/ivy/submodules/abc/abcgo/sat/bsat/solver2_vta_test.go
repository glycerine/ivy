package bsat

import "testing"

func TestLabeledSolver2VtaControls(t *testing.T) {
	s := NewLabeledSolver2()
	s.SetLearnedPolicy(100, 20, 30)
	internal := s.solver.solver
	if internal.nLearntStart != 100 || internal.nLearntMax != 100 || internal.nLearntDelta != 20 || internal.nLearntRatio != 30 {
		t.Fatal("learned policy not applied")
	}
	if !s.AddClause(1, Positive(1)) || !s.Simplify() {
		t.Fatal("unit propagation failed")
	}
	s.Bookmark()
	if !s.AddClause(2, Negative(1)) { // Contradiction must remain rollbackable.
		if got := s.Solve(); got != Unsatisfiable {
			t.Fatalf("got %v", got)
		}
	}
	s.Rollback()
	if got := s.Solve(); got != Satisfiable {
		t.Fatalf("rollback got %v", got)
	}
	if s.Reductions() != internal.nDBreduces {
		t.Fatal("wrong reduction count")
	}
	a, b := s.Memory(true)
	if a < 0 || b < 0 {
		t.Fatal("negative memory estimate")
	}
}
