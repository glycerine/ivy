package bsat

import (
	"math/rand"
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

func TestLexicographicFirstModelAgainstExhaustiveOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(92318))
	for trial := 0; trial < 500; trial++ {
		s := NewSolver()
		for i := 0; i < 4; i++ {
			s.NewVariable()
		}
		var clauses [][]lit
		for k := 0; k < rng.Intn(12); k++ {
			c := make([]lit, 1+rng.Intn(3))
			for i := range c {
				c[i] = lit(rng.Intn(8))
			}
			clauses = append(clauses, c)
			s.AddClause(c...)
		}
		order := []Literal{Negative(0), Negative(1), Negative(2), Negative(3)}
		want := -1
		for mask := 0; mask < 16; mask++ {
			ass := make([]lit, 4)
			for i := range ass {
				ass[i] = Negative(int32(i))
				if mask&(1<<uint(3-i)) != 0 {
					ass[i] = Positive(int32(i))
				}
			}
			if formulaHasModel(clauses, ass, 4) {
				want = mask
				break
			}
		}
		got := s.SolveLexicographic(order)
		if want < 0 {
			if got != Unsatisfiable {
				t.Fatalf("trial=%d want UNSAT got=%v", trial, got)
			}
			continue
		}
		if got != Satisfiable {
			t.Fatalf("trial=%d got=%v want SAT", trial, got)
		}
		model := s.Model()
		for i, l := range order {
			expected := want&(1<<uint(3-i)) != 0
			if model[i] != expected || (l&1 == 0) != expected {
				t.Fatalf("trial=%d want=%04b order=%v model=%v clauses=%v", trial, want, order, model, clauses)
			}
		}
	}
}
