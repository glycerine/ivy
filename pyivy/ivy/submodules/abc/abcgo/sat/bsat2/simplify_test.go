package bsat2

import (
	"math/rand"
	"testing"
)

func TestSimplifierRejectsEliminatedAssumptionsAfterPreprocessing(t *testing.T) {
	for _, turnOff := range []bool{false, true} {
		s := NewSimpSolver()
		s.NewVariable()
		s.NewVariable()
		s.SetFrozen(1, true)
		if !s.AddClause(MakeLit(0, false), MakeLit(1, false)) || !s.Eliminate(turnOff) || !s.IsEliminated(0) {
			t.Fatal("could not prepare eliminated variable")
		}
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("turnOff=%v: eliminated assumption accepted", turnOff)
				}
			}()
			s.SolveSimplified([]Lit{MakeLit(0, true)}, false, false)
		}()
		if s.SolveSimplified([]Lit{MakeLit(1, false)}, false, false) != True {
			t.Fatal("solver unusable after rejected assumption")
		}
	}
}

func TestSimplifierIncrementalAfterElimination(t *testing.T) {
	rng := rand.New(rand.NewSource(867340))
	for trial := 0; trial < 300; trial++ {
		s := NewSimpSolver()
		s.UseRedundancyCheck = trial%2 == 0
		s.UseAsymmetric = trial%3 == 0
		for v := 0; v < 5; v++ {
			s.NewVariable()
			if v != 0 {
				s.SetFrozen(Var(v), true)
			}
		}
		var formula [][]Lit
		for i := 0; i < 8; i++ {
			var clause []Lit
			for _, v := range rng.Perm(5)[:3] {
				clause = append(clause, MakeLit(Var(v), rng.Intn(2) != 0))
			}
			formula = append(formula, clause)
			s.AddClause(clause...)
		}
		s.Eliminate(trial%2 == 0)
		for step := 0; step < 8; step++ {
			if step > 0 && step%2 == 0 {
				clause := []Lit{MakeLit(Var(1+rng.Intn(4)), rng.Intn(2) != 0), MakeLit(Var(1+rng.Intn(4)), rng.Intn(2) != 0)}
				formula = append(formula, clause)
				s.AddClause(clause...)
			}
			var assumptions []Lit
			for i := 0; i < rng.Intn(4); i++ {
				assumptions = append(assumptions, MakeLit(Var(1+rng.Intn(4)), rng.Intn(2) != 0))
			}
			_, want := formulaModel(formula, assumptions, 5)
			status := s.SolveSimplified(assumptions, true, false)
			if (status == True) != want || status == Undef {
				t.Fatalf("trial=%d step=%d status=%v want=%v formula=%v assumptions=%v", trial, step, status, want, formula, assumptions)
			}
			if status == True {
				for _, clause := range formula {
					satisfied := false
					for _, l := range clause {
						if s.ModelValue(l) == True {
							satisfied = true
							break
						}
					}
					if !satisfied {
						t.Fatalf("trial=%d step=%d invalid model %v clause %v", trial, step, s.Model, clause)
					}
				}
				for _, l := range assumptions {
					if s.ModelValue(l) != True {
						t.Fatalf("trial=%d step=%d model %v violates assumption %v", trial, step, s.Model, l)
					}
				}
			}
		}
	}
}

func TestSimplifyingSolverAgainstExhaustiveEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(431))
	for trial := 0; trial < 500; trial++ {
		s := NewSimpSolver()
		if trial%5 == 0 {
			s.GarbageFraction = 0
			s.SimplificationGarbageFraction = 0
		}
		for i := 0; i < 6; i++ {
			s.NewVariable()
		}
		s.UseAsymmetric = trial%3 == 0
		s.UseRedundancyCheck = trial%4 == 0
		formula := [][]Lit{}
		for count := 0; count < 18; count++ {
			c := []Lit{}
			for _, v := range rng.Perm(6)[:2+rng.Intn(3)] {
				c = append(c, MakeLit(Var(v), rng.Intn(2) == 0))
			}
			formula = append(formula, c)
			s.AddClause(c...)
		}
		assumptions := []Lit{}
		if trial&1 != 0 {
			assumptions = append(assumptions, MakeLit(0, trial&2 != 0))
		}
		_, expected := formulaModel(formula, assumptions, 6)
		got := s.Solve(assumptions...)
		if (got == True) != expected || got == Undef {
			t.Fatal(trial, got, expected, formula, assumptions)
		}
		if got == True {
			for _, c := range formula {
				satisfied := false
				for _, l := range c {
					if s.ModelValue(l) == True {
						satisfied = true
					}
				}
				if !satisfied {
					t.Fatal("reconstructed model violates original clause", trial, c, s.Model)
				}
			}
		}
	}
}
func TestEliminationFrozenVariablesAndGarbage(t *testing.T) {
	s := NewSimpSolver()
	for i := 0; i < 4; i++ {
		s.NewVariable()
	}
	for _, c := range [][]Lit{{0, 2}, {1, 4}, {3, 6}, {5, 7}} {
		s.AddClause(c...)
	}
	s.SetFrozen(0, true)
	if !s.Eliminate(false) || s.IsEliminated(0) || s.EliminatedVariables == 0 {
		t.Fatal("elimination/frozen state")
	}
	s.GarbageCollect()
	if s.Solve(0) != True {
		t.Fatal("frozen incremental assumption")
	}
}

func TestSubstitutionExtendsOriginalModel(t *testing.T) {
	for _, negative := range []bool{false, true} {
		s := NewSimpSolver()
		s.NewVariable()
		s.NewVariable()
		s.AddClause(0, 2)
		replacement := MakeLit(1, negative)
		if !s.Substitute(0, replacement) || s.Solve() != True || s.ModelValue(0) != s.ModelValue(replacement) {
			t.Fatal(negative, s.Model)
		}
		if s.ModelValue(0) != True && s.ModelValue(2) != True {
			t.Fatal("substitution lost original clause")
		}
	}
}

func TestSimplificationFailureClearsPreviousModel(t *testing.T) {
	s := NewSimpSolver()
	s.NewVariable()
	s.SetFrozen(0, true)
	s.AddClause(0)
	if s.Solve() != True {
		t.Fatal("initial SAT")
	}
	s.AddClause(1)
	if s.Solve() != False || s.ModelValue(0) != Undef {
		t.Fatal("stale SAT model after preprocessing UNSAT", s.Model)
	}
}
