package bsat_test

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestPublicPreferredDecisionPolarity(t *testing.T) {
	for _, value := range []bool{false, true} {
		solver := bsat.NewSolver()
		variable := solver.NewVariable()
		solver.SetPolarity(variable, value)
		if solver.Solve() != bsat.Satisfiable {
			t.Fatal("unconstrained variable unsatisfiable")
		}
		got, ok := solver.Value(variable)
		if !ok || got != value {
			t.Fatalf("preferred%v got%v known%v", value, got, ok)
		}
	}
}
func TestPublicPolarityDoesNotConstrainFormula(t *testing.T) {
	solver := bsat.NewSolver()
	variable := solver.NewVariable()
	solver.SetPolarity(variable, true)
	solver.AddClause(bsat.Negative(variable))
	if solver.Solve() != bsat.Satisfiable {
		t.Fatal("phase hint overrode a clause")
	}
	value, ok := solver.Value(variable)
	if !ok || value {
		t.Fatal("forced value was ignored")
	}
	for _, variable := range []int32{-1, 1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("accepted variable%d", variable)
				}
			}()
			solver.SetPolarity(variable, false)
		}()
	}
}
