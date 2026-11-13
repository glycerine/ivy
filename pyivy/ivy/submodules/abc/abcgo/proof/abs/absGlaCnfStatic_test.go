package abs

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestGlaStaticCnfAndTruthTable(t *testing.T) {
	for x := 0; x < 2; x++ {
		for y := 0; y < 2; y++ {
			for out := 0; out < 2; out++ {
				solver := bsat.NewLabeledSolver2()
				for i := 0; i < 3; i++ {
					solver.NewVariable()
				}
				zero := vec.Vec_IntAlloc(2)
				vec.Vec_IntPush(zero, 1)
				vec.Vec_IntPush(zero, 4)
				one := vec.Vec_IntAlloc(1)
				vec.Vec_IntPush(one, 10)
				Ga2_ManCnfAddStatic(solver, one, zero, []int32{0, 2}, 4, 1)
				assumptions := []bsat.Literal{bsat.Literal(x ^ 1), bsat.Literal(2 | (y ^ 1)), bsat.Literal(4 | (out ^ 1))}
				sat := solver.Solve(assumptions...) == bsat.Satisfiable
				if sat != (out == x*y) {
					t.Fatalf("x=%d y=%d out=%d sat=%v", x, y, out, sat)
				}
			}
		}
	}
}
