package bsat2

import (
	"math/rand"
	"testing"
)

func formulaModel(formula [][]Lit, assumptions []Lit, n int) (int, bool) {
	for model := 0; model < 1<<uint(n); model++ {
		valid := true
		for _, l := range assumptions {
			if (model&(1<<uint(l.Var())) != 0) == l.Sign() {
				valid = false
			}
		}
		for _, c := range formula {
			satisfied := false
			for _, l := range c {
				if (model&(1<<uint(l.Var())) != 0) != l.Sign() {
					satisfied = true
				}
			}
			if !satisfied {
				valid = false
				break
			}
		}
		if valid {
			return model, true
		}
	}
	return 0, false
}
func TestMiniSatCDCLAgainstExhaustiveEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(74289))
	for trial := 0; trial < 500; trial++ {
		s := NewSolver()
		s.Minimization = trial % 3
		s.PhaseSaving = trial % 3
		if trial%7 == 0 {
			s.RandomFrequency = 0.2
			s.RestartFirst = 1
		}
		for i := 0; i < 5; i++ {
			s.NewVariable()
		}
		formula := [][]Lit{}
		for count := 0; count < 20; count++ {
			c := []Lit{}
			size := 1 + rng.Intn(4)
			for _, v := range rng.Perm(5)[:size] {
				c = append(c, MakeLit(Var(v), rng.Intn(2) == 0))
			}
			formula = append(formula, c)
			s.AddClause(c...)
		}
		assumptions := []Lit{}
		if trial&1 != 0 {
			assumptions = append(assumptions, MakeLit(Var(rng.Intn(5)), rng.Intn(2) == 0))
		}
		_, expected := formulaModel(formula, assumptions, 5)
		got := s.Solve(assumptions...)
		if (got == True) != expected || got == Undef {
			t.Fatal(trial, got, expected, formula, assumptions)
		}
		if got == True {
			model := 0
			for v, value := range s.Model {
				if value == True {
					model |= 1 << uint(v)
				}
			}
			for _, c := range formula {
				satisfied := false
				for _, l := range c {
					if (model&(1<<uint(l.Var())) != 0) != l.Sign() {
						satisfied = true
					}
				}
				if !satisfied {
					t.Fatal("bad model", trial, c, s.Model)
				}
			}
		}
		_, baseSAT := formulaModel(formula, nil, 5)
		if (s.Solve() == True) != baseSAT {
			t.Fatal("assumptions leaked", trial)
		}
	}
}
func TestMiniSatGarbageCollectionRetainsReasons(t *testing.T) {
	s := NewSolver()
	s.ca.ExtraClauseField = true
	s.AddClause(0, 2)
	s.AddClause(1, 4, 6)
	s.AddClause(3, 5)
	s.AddClause(1)
	before := s.Value(2)
	s.GarbageCollect()
	if s.Value(2) != before || s.Solve() != True {
		t.Fatal("garbage collection changed formula")
	}
}

func TestResourceBudgetAndInterrupt(t *testing.T) {
	s := NewSolver()
	s.NewVariable()
	s.SetPropagationBudget(0)
	if s.SolveLimited() != Undef {
		t.Fatal("budget")
	}
	s.BudgetOff()
	s.Interrupt()
	if s.SolveLimited() != Undef {
		t.Fatal("interrupt")
	}
	s.ClearInterrupt()
	if s.Solve() != True {
		t.Fatal("resume")
	}
}
