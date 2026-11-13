package glucose2

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestDimacsRoundTripRootAssignmentsAndAssumptions(t *testing.T) {
	for _, tc := range []struct {
		name        string
		root        bool
		assumptions []bool
		want        bool
	}{
		{"contradicts true", true, []bool{true}, false},
		{"contradicts false", false, []bool{false}, false},
		{"agrees true", true, []bool{false}, true},
		{"agrees false", false, []bool{true}, true},
		{"no assumption", true, nil, true},
		{"mixed assumptions", true, []bool{false, true}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSolver()
			x := s.NewVar(false, true)
			if !s.AddClause([]Lit{MkLit(x, !tc.root)}) {
				t.Fatal("root clause")
			}
			var assumptions []Lit
			for _, neg := range tc.assumptions {
				assumptions = append(assumptions, MkLit(x, neg))
			}
			var out bytes.Buffer
			if err := s.ToDimacs(&out, assumptions); err != nil {
				t.Fatal(err)
			}
			parsed := NewSolver()
			if _, err := ParseDIMACS(bytes.NewReader(out.Bytes()), parsed); err != nil {
				t.Fatal(err)
			}
			if got := parsed.Solve(nil); got != tc.want {
				t.Fatalf("exported DIMACS satisfiable=%v want=%v; %q", got, tc.want, out.String())
			}
			if got := s.Solve(assumptions); got != tc.want {
				t.Fatalf("original satisfiable=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestDimacsRoundTripRandomRootSimplification(t *testing.T) {
	rng := rand.New(rand.NewSource(439))
	for trial := 0; trial < 300; trial++ {
		s := NewSolver()
		for i := 0; i < 4; i++ {
			s.NewVar(false, true)
		}
		for k := 0; k < 6; k++ {
			n := 1 + rng.Intn(3)
			clause := make([]Lit, n)
			for i := range clause {
				clause[i] = MkLit(Var(rng.Intn(4)), rng.Intn(2) != 0)
			}
			s.AddClause(clause)
		}
		assumptions := make([]Lit, rng.Intn(4))
		for i := range assumptions {
			assumptions[i] = MkLit(Var(rng.Intn(4)), rng.Intn(2) != 0)
		}
		var out bytes.Buffer
		if err := s.ToDimacs(&out, assumptions); err != nil {
			t.Fatal(err)
		}
		clone := NewSolver()
		if _, err := ParseDIMACS(bytes.NewReader(out.Bytes()), clone); err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		if got, want := clone.Solve(nil), s.Solve(assumptions); got != want {
			t.Fatalf("trial %d: got=%v want=%v assumptions=%v CNF=%q", trial, got, want, assumptions, out.String())
		}
	}
}
