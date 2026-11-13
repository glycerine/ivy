package bsat

import (
	"testing"
	"time"
)

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

func TestSolver2RootUnitAssumptionProofCore(t *testing.T) {
	s := NewSolver2()
	x := s.NewVariable()
	if !s.AddClause(Positive(x)) {
		t.Fatal("add unit")
	}
	if s.Solve(Negative(x)) != Unsatisfiable {
		t.Fatal("expected assumption conflict")
	}
	core := s.ProofCore()
	if len(core) != 1 || core[0] != 1 {
		t.Fatalf("root clause core=%v", core)
	}
	if s.Solve(Positive(x)) != Satisfiable {
		t.Fatal("assumption conflict poisoned formula")
	}
}

func TestSolver2DeadlineAndRecovery(t *testing.T) {
	s := NewSolver2()
	x, y := s.NewVariable(), s.NewVariable()
	s.AddClause(Positive(x), Positive(y))
	s.SetDeadline(time.Unix(0, 1))
	if got := s.Solve(); got != Unknown {
		t.Fatalf("expired deadline returned %v", got)
	}
	if s.Model() != nil || s.ProofCore() != nil {
		t.Fatal("interrupted solve exposed a result")
	}
	s.SetDeadline(time.Time{})
	if got := s.Solve(Negative(x)); got != Satisfiable {
		t.Fatalf("recovery returned %v", got)
	}
	if v, ok := s.Value(y); !ok || !v {
		t.Fatalf("recovered model y=%v valid=%v", v, ok)
	}
}
