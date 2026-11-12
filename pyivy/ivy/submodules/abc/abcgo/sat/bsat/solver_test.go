package bsat

import (
	"testing"
	"time"
)

func TestPublicIncrementalSolver(t *testing.T) {
	s := NewSolver()
	a, b := s.NewVariable(), s.NewVariable()
	if !s.AddClause(Positive(a), Positive(b)) {
		t.Fatal("clause")
	}
	if s.Solve(Negative(a), Negative(b)) != Unsatisfiable || len(s.UnsatCore()) != 2 {
		t.Fatal("assumption conflict")
	}
	if s.Solve(Negative(a)) != Satisfiable {
		t.Fatal("incremental solve")
	}
	if value, ok := s.Value(b); !ok || !value {
		t.Fatal("model")
	}
	m := s.Model()
	m[b] = false
	if value, _ := s.Value(b); !value {
		t.Fatal("model alias")
	}
	s.AddClause(Negative(b))
	if s.Solve() != Satisfiable {
		t.Fatal("incremental unit")
	}
	if s.AddClause(Negative(a)) {
		t.Fatal("contradictory root unit")
	}
	if s.Solve() != Unsatisfiable {
		t.Fatal("inconsistent formula forgotten")
	}
	s.Reset()
	if s.NumVars() != 0 || s.Solve() != Satisfiable {
		t.Fatal("reset")
	}
	s.SetDeadline(time.Now().Add(-time.Second))
	if s.Solve() != Unknown {
		t.Fatal("deadline")
	}
}
func TestPublicEmptyClausePersists(t *testing.T) {
	s := NewSolver()
	if s.AddClause() || s.Solve() != Unsatisfiable || s.AddClause(Positive(0)) {
		t.Fatal("empty clause consistency")
	}
}
