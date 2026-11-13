package bsat

import "testing"

func TestSimplifyPropagatesWithoutSearch(t *testing.T) {
	s := NewSolver()
	if !s.AddClause(Positive(0)) || !s.AddClause(Negative(0), Positive(1)) {
		t.Fatal("unexpected contradiction")
	}
	before := s.stats.conflicts
	if !s.Simplify() {
		t.Fatal("consistent root constraints rejected")
	}
	if s.stats.conflicts != before {
		t.Fatal("simplification performed SAT search")
	}
	if s.AddClause(Negative(1)) {
		t.Fatal("propagated implication was not retained")
	}
	if s.Simplify() {
		t.Fatal("inconsistent solver simplified successfully")
	}
}
