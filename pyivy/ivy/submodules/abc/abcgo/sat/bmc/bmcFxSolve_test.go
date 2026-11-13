package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestFxSolveAnd(t *testing.T) {
	s := bsat.NewSolver()
	for i := 0; i < 5; i++ {
		s.NewVariable()
	}
	// Variable 1 is x AND y; 4 guards blocking clauses.
	for _, cl := range [][]bsat.Literal{{bsat.Negative(1), bsat.Positive(2)}, {bsat.Negative(1), bsat.Positive(3)}, {bsat.Positive(1), bsat.Negative(2), bsat.Negative(3)}} {
		s.AddClause(cl...)
	}
	vars := vec.Vec_IntAllocArrayCopy([]int32{2, 3}, 2)
	cubes := vec.Vec_WecAlloc(2)
	var count int32
	if got := Bmc_FxSolve(s, 0, 4, vars, 0, 0, &count, cubes); got != 1 || count != 1 || vec.Vec_WecSize(cubes) != 1 {
		t.Fatalf("result=%d count=%d", got, count)
	}
	cube := vec.Vec_IntArray(vec.Vec_WecEntry(cubes, 0))
	if len(cube) != 2 || cube[0] != 0 || cube[1] != 2 {
		t.Fatalf("cube %v", cube)
	}
}
