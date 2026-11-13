package bsat

import (
	"testing"
	"time"
)

func TestPublicLexicographicEnumeration(t *testing.T) {
	s := NewSolver()
	for i := 0; i < 3; i++ {
		s.NewVariable()
	}
	order := []Literal{Negative(2), Negative(0), Negative(1)}
	for expected := 0; expected < 8; expected++ {
		if got := s.SolveLexicographic(order); got != Satisfiable {
			t.Fatal(expected, got)
		}
		for bit, l := range order {
			want := expected&(1<<uint(2-bit)) != 0
			if value, valid := s.Value(lit_var(l)); !valid || value != want || (lit_sign(l) == 0) != want {
				t.Fatal(expected, order, s.Model())
			}
		}
		blocking := make([]Literal, len(order))
		for i, l := range order {
			blocking[i] = lit_neg(l)
		}
		s.AddClause(blocking...)
		if s.root_level != 0 || s.trail_lim.size != 0 {
			t.Fatal("temporary assumptions leaked")
		}
	}
	if s.SolveLexicographic(order) != Unsatisfiable || s.Model() != nil {
		t.Fatal("enumeration exhaustion")
	}
}
func TestPublicLexicographicEmptyAndDeadline(t *testing.T) {
	s := NewSolver()
	if s.SolveLexicographic(nil) != Satisfiable {
		t.Fatal("empty order")
	}
	s.NewVariable()
	s.SetDeadline(time.Unix(0, 1))
	if s.SolveLexicographic([]Literal{Negative(0)}) != Unknown || s.Model() != nil {
		t.Fatal("deadline")
	}
	s.SetDeadline(time.Time{})
	if s.SolveLexicographic([]Literal{Negative(0)}) != Satisfiable {
		t.Fatal("resume")
	}
}

func TestLexicographicRootConflictPersists(t *testing.T) {
	s := NewSolver()
	x, y := s.NewVariable(), s.NewVariable()
	s.AddClause(Positive(x), Positive(y))
	s.AddClause(Negative(x))
	s.AddClause(Negative(y))
	lits := []Literal{Negative(x), Negative(y)}
	if status := s.SolveLexicographic(lits); status != Unsatisfiable {
		t.Fatalf("first solve=%v", status)
	}
	if status := s.SolveLexicographic(lits); status != Unsatisfiable {
		t.Fatalf("root conflict lost: %v", status)
	}
	if status := s.Solve(); status != Unsatisfiable {
		t.Fatalf("base formula should be inconsistent: %v", status)
	}
}
