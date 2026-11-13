package bsat

import "testing"

func TestSolveWithGlobalLimitsCombinesBudgets(t *testing.T) {
	s := NewSolver()
	s.stats.conflicts = 3
	s.stats.propagations = 4
	s.SetLimits(10, 20)
	if status := s.SolveWithGlobalLimits(nil, 8, 11); status != Satisfiable {
		t.Fatalf("status = %v", status)
	}
	if s.nConfLimit != 8 || s.nInsLimit != 11 {
		t.Fatalf("limits = %d/%d", s.nConfLimit, s.nInsLimit)
	}
	if status := s.SolveWithGlobalLimits(nil, 30, 40); status != Satisfiable {
		t.Fatalf("status = %v", status)
	}
	if s.nConfLimit != 13 || s.nInsLimit != 24 {
		t.Fatalf("per-call limits = %d/%d", s.nConfLimit, s.nInsLimit)
	}
	if s.lastStatus != Satisfiable {
		t.Fatal("last status not recorded")
	}
}
func TestSolveWithGlobalLimitsRejectsNegative(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	NewSolver().SolveWithGlobalLimits(nil, -1, 0)
}
