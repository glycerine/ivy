package bsat

import (
	"fmt"
	"math/rand"
	"testing"
)

// Compare repeated public API calls against independent exhaustive evaluation.
func TestPublicSolverIncrementalOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(20261113))
	for trial := 0; trial < 250; trial++ {
		s := NewSolver()
		for i := 0; i < 5; i++ {
			s.NewVariable()
		}
		var formula [][]lit
		var history []string
		for step := 0; step < 20; step++ {
			if step%4 == 0 {
				clause := make([]lit, 1+rng.Intn(3))
				for i := range clause {
					clause[i] = lit(rng.Intn(10))
				}
				history = append(history, fmt.Sprintf("add %v", clause))
				formula = append(formula, append([]lit(nil), clause...))
				s.AddClause(clause...)
			}
			assumptions := make([]lit, rng.Intn(4))
			for i := range assumptions {
				assumptions[i] = lit(rng.Intn(10))
			}
			history = append(history, fmt.Sprintf("solve %v", assumptions))
			want := formulaHasModel(formula, assumptions, 5)
			got := s.Solve(assumptions...)
			if (got == Satisfiable) != want {
				t.Fatalf("trial=%d step=%d clauses=%v assumptions=%v got=%v want SAT=%v history=%v", trial, step, formula, assumptions, got, want, history)
			}
			if got == Satisfiable {
				model := s.Model()
				if len(model) != 5 {
					t.Fatalf("model length %d", len(model))
				}
				value := func(l lit) bool { return model[l>>1] != (l&1 != 0) }
				for _, clause := range formula {
					ok := false
					for _, l := range clause {
						ok = ok || value(l)
					}
					if !ok {
						t.Fatalf("invalid model %v for clause %v", model, clause)
					}
				}
				for _, l := range assumptions {
					if !value(l) {
						t.Fatalf("model %v violates assumption %d", model, l)
					}
				}
			} else {
				core := s.UnsatCore()
				if len(core) > 0 {
					subset := make([]lit, len(core))
					for i, l := range core {
						subset[i] = l ^ 1
					}
					if formulaHasModel(formula, subset, 5) {
						t.Fatalf("invalid core %v", core)
					}
				}
			}
		}
	}
}

func TestRootConflictBeforeAssumptionsPersists(t *testing.T) {
	s := NewSolver()
	x, y := s.NewVariable(), s.NewVariable()
	if !s.AddClause(Positive(x), Positive(y)) || !s.AddClause(Negative(x)) || !s.AddClause(Negative(y)) {
		t.Fatal("unexpected early contradiction")
	}
	if got := s.Solve(Positive(x)); got != Unsatisfiable {
		t.Fatalf("first solve: %v", got)
	}
	if got := s.Solve(Negative(x)); got != Unsatisfiable {
		t.Fatalf("root conflict was lost: %v", got)
	}
	if got := s.Solve(); got != Unsatisfiable {
		t.Fatalf("base formula should be inconsistent: %v", got)
	}
}

func TestProofSolverRootConflictBeforeAssumptionsPersists(t *testing.T) {
	s := NewSolver2()
	x, y := s.NewVariable(), s.NewVariable()
	if !s.AddClause(Positive(x), Positive(y)) || !s.AddClause(Negative(x)) || !s.AddClause(Negative(y)) {
		t.Fatal("unexpected early contradiction")
	}
	if got := s.Solve(Positive(x)); got != Unsatisfiable {
		t.Fatalf("first solve: %v", got)
	}
	core := s.ProofCore()
	if len(core) != 3 {
		t.Fatalf("root proof core=%v, want all three clauses", core)
	}
	if got := s.Solve(Negative(x)); got != Unsatisfiable {
		t.Fatalf("root conflict was lost: %v", got)
	}
	if got := s.Solve(); got != Unsatisfiable {
		t.Fatalf("base formula should be inconsistent: %v", got)
	}
}

func TestProofSolverIncrementalOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(734916))
	for trial := 0; trial < 200; trial++ {
		s := NewSolver2()
		for i := 0; i < 4; i++ {
			s.NewVariable()
		}
		var formula [][]lit
		for step := 0; step < 16; step++ {
			if step%4 == 0 {
				clause := make([]lit, 1+rng.Intn(3))
				for i := range clause {
					clause[i] = lit(rng.Intn(8))
				}
				formula = append(formula, append([]lit(nil), clause...))
				s.AddClause(clause...)
			}
			assumptions := make([]lit, rng.Intn(3))
			for i := range assumptions {
				assumptions[i] = lit(rng.Intn(8))
			}
			want := formulaHasModel(formula, assumptions, 4)
			status := s.Solve(assumptions...)
			if (status == Satisfiable) != want {
				t.Fatalf("trial %d step %d formula %v assumptions %v result %v", trial, step, formula, assumptions, status)
			}
			if status == Satisfiable {
				model := s.Model()
				value := func(l lit) bool { return model[l>>1] != (l&1 != 0) }
				for _, clause := range formula {
					valid := false
					for _, l := range clause {
						valid = valid || value(l)
					}
					if !valid {
						t.Fatalf("invalid model %v for %v", model, clause)
					}
				}
				for _, l := range assumptions {
					if !value(l) {
						t.Fatalf("model %v violates %d", model, l)
					}
				}
			} else {
				core := s.ProofCore()
				var selected [][]lit
				for _, id := range core {
					if id < 1 || int(id) > len(formula) {
						t.Fatalf("invalid proof id %d in %v", id, core)
					}
					selected = append(selected, formula[id-1])
				}
				if formulaHasModel(selected, assumptions, 4) {
					t.Fatalf("invalid proof core %v formula=%v assumptions=%v", core, formula, assumptions)
				}
			}
		}
	}
}
