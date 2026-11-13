package glucose2

import (
	"math/rand"
	"testing"
)

// Compare repeated solves and incremental clauses after elimination with an
// independent exhaustive oracle over the original, uneliminated formula.
func TestSimpIncrementalFrozenVariablesAgainstExhaustive(t *testing.T) {
	rng := rand.New(rand.NewSource(723489))
	for trial := 0; trial < 300; trial++ {
		const n = 6
		s := NewSimpSolver()
		s.UseAsymm = trial%3 == 0
		s.UseRCheck = trial%5 == 0
		for i := 0; i < n; i++ {
			s.NewVar(true, true)
			if i < 3 {
				s.SetFrozen(Var(i), true)
			}
		}
		var formula [][]Lit
		for i := 0; i < 3+rng.Intn(16); i++ {
			clause := make([]Lit, 1+rng.Intn(4))
			for j := range clause {
				clause[j] = MkLit(Var(rng.Intn(n)), rng.Intn(2) != 0)
			}
			formula = append(formula, append([]Lit(nil), clause...))
			s.AddClause(clause)
		}
		for step := 0; step < 8; step++ {
			var assumptions []Lit
			for i := 0; i < rng.Intn(4); i++ {
				assumptions = append(assumptions, MkLit(Var(rng.Intn(3)), rng.Intn(2) != 0))
			}
			want := bruteCNF(n, formula, assumptions)
			got := s.Solve(assumptions)
			if got != want {
				t.Fatalf("trial %d step %d status=%v want=%v clauses=%v assumptions=%v", trial, step, got, want, formula, assumptions)
			}
			if got {
				for _, clause := range formula {
					satisfied := false
					for _, lit := range clause {
						if s.ModelValue(lit) == LTrue {
							satisfied = true
							break
						}
					}
					if !satisfied {
						t.Fatalf("trial %d step %d unsatisfied clause %v model=%v", trial, step, clause, s.Model)
					}
				}
				for _, lit := range assumptions {
					if s.ModelValue(lit) != LTrue {
						t.Fatalf("trial %d step %d assumption %v model=%v", trial, step, lit, s.Model)
					}
				}
			}
			if step%2 == 1 {
				clause := []Lit{MkLit(Var(rng.Intn(3)), rng.Intn(2) != 0), MkLit(Var(rng.Intn(3)), rng.Intn(2) != 0)}
				formula = append(formula, append([]Lit(nil), clause...))
				s.AddClause(clause)
			}
		}
	}
}

func TestSimpSubstitutionAgainstExplicitEquivalence(t *testing.T) {
	rng := rand.New(rand.NewSource(44578))
	for trial := 0; trial < 300; trial++ {
		const n = 5
		s := NewSimpSolver()
		for i := 0; i < n; i++ {
			s.NewVar(true, true)
			if i > 0 {
				s.SetFrozen(Var(i), true)
			}
		}
		var formula [][]Lit
		// Binary clauses on distinct variables avoid initial unit assignments.
		for i := 0; i < 1+rng.Intn(9); i++ {
			a, b := rng.Intn(n), rng.Intn(n)
			if a == b {
				b = (b + 1) % n
			}
			clause := []Lit{MkLit(Var(a), rng.Intn(2) != 0), MkLit(Var(b), rng.Intn(2) != 0)}
			formula = append(formula, append([]Lit(nil), clause...))
			s.AddClause(clause)
		}
		replacement := MkLit(1, trial%2 != 0)
		formula = append(formula, []Lit{MkLit(0, true), replacement}, []Lit{MkLit(0, false), replacement.Not()})
		if !s.Substitute(0, replacement) {
			if bruteCNF(n, formula, nil) {
				t.Fatalf("trial %d substitution incorrectly UNSAT", trial)
			}
			continue
		}
		for mask := 0; mask < 16; mask++ {
			assumptions := make([]Lit, 4)
			for i := 1; i < n; i++ {
				assumptions[i-1] = MkLit(Var(i), mask&(1<<(i-1)) == 0)
			}
			want := bruteCNF(n, formula, assumptions)
			got := s.Solve(assumptions)
			if got != want {
				t.Fatalf("trial %d mask=%04b got %v want %v formula=%v", trial, mask, got, want, formula)
			}
			if got {
				for _, clause := range formula {
					sat := false
					for _, lit := range clause {
						if s.ModelValue(lit) == LTrue {
							sat = true
							break
						}
					}
					if !sat {
						t.Fatalf("trial %d mask=%04b model=%v violates %v", trial, mask, s.Model, clause)
					}
				}
			}
		}
	}
}
