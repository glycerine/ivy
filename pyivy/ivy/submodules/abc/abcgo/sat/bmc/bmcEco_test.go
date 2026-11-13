package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestEcoSolveDistinguishingInputs(t *testing.T) {
	makeSolver := func() *bsat.Solver {
		s := bsat.NewSolver()
		s.NewVariable()
		s.NewVariable()
		s.AddClause(bsat.Negative(0), bsat.Positive(1))
		s.AddClause(bsat.Positive(0), bsat.Negative(1))
		return s
	}
	vars := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(vars, 1)
	if got := Bmc_EcoSolve(makeSolver(), 0, vars); got != 1 {
		t.Fatalf("separable=%d", got)
	}
	if got := Bmc_EcoSolve(makeSolver(), 0, vec.Vec_IntAlloc(0)); got != 0 {
		t.Fatalf("unseparable=%d", got)
	}
}
