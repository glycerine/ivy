package bsat

import "testing"

func TestRestartSearchClearsFormulaAndContradiction(t *testing.T) {
	s := NewSolver()
	defer s.Close()
	if !s.AddClause(Positive(0)) || s.AddClause(Negative(0)) {
		t.Fatal("expected contradictory units")
	}
	if s.Solve() != Unsatisfiable {
		t.Fatal("expected unsatisfiable formula")
	}
	s.RestartSearch()
	if s.NumVars() != 0 || s.Model() != nil || s.UnsatCore() != nil {
		t.Fatal("restart retained prior solve state")
	}
	if !s.AddClause(Negative(0)) || s.Solve() != Satisfiable {
		t.Fatal("restart retained old formula or contradiction")
	}
	if value, ok := s.Value(0); !ok || value {
		t.Fatalf("new model=(%v,%v)", value, ok)
	}
}
