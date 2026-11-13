package bsat

import "testing"

func TestPublicSolver2ContradictoryRootProof(t *testing.T) {
	s := NewSolver2()
	v := s.NewVariable()
	if !s.AddClause(Positive(v)) || s.AddClause(Negative(v)) {
		t.Fatal("root consistency")
	}
	if s.Solve() != Unsatisfiable || len(s.ProofCore()) != 2 {
		t.Fatal(s.ProofCore())
	}
	if _, valid := s.Value(v); valid {
		t.Fatal("UNSAT model exposed")
	}
	s.Reset()
	if s.NumVars() != 0 || s.Solve() != Satisfiable {
		t.Fatal("reset")
	}
	if s.AddClause() || s.Solve() != Unsatisfiable || len(s.ProofCore()) != 1 {
		t.Fatal("empty-clause root proof", s.ProofCore())
	}
}

func TestSolver2NewVariableInvalidatesModel(t *testing.T) {
	s := NewSolver2()
	s.NewVariable()
	if s.Solve() != Satisfiable || len(s.Model()) != 1 {
		t.Fatal("missing initial model")
	}
	s.NewVariable()
	if model := s.Model(); model != nil {
		t.Fatalf("stale model after adding variable: %v", model)
	}
	if _, ok := s.Value(0); ok {
		t.Fatal("stale value after adding variable")
	}
	if s.Solve() != Satisfiable || len(s.Model()) != 2 {
		t.Fatal("missing extended model")
	}
}
