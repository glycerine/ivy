package abs

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"testing"
)

func TestGlaDynamicCnfTruthTable(t *testing.T) {
	for _, truth := range []uint32{0, 0xffff, 0x6996, 0x8e80} {
		for inputs := 0; inputs < 16; inputs++ {
			for output := 0; output < 2; output++ {
				solver := bsat.Sat_solver2_new()
				for i := 0; i < 6; i++ {
					ga2NewVar(solver)
				}
				p := &Ga2_Man_t{pSat: solver, pSops: cnf.Cnf_ReadMsops()}
				Ga2_ManCnfAddDynamic(p, int32(truth), []int32{2, 4, 6, 8}, 10, 1)
				assumptions := make([]bsat.Literal, 0, 5)
				for i := 0; i < 4; i++ {
					assumptions = append(assumptions, bsat.Literal(2*(i+1)|(((inputs>>i)&1)^1)))
				}
				assumptions = append(assumptions, bsat.Literal(10|(output^1)))
				got := ga2Solve(solver, assumptions...) == bsat.Satisfiable
				want := output == int((truth>>inputs)&1)
				if got != want {
					t.Fatalf("truth=%04x inputs=%x output=%d got=%v", truth, inputs, output, got)
				}
			}
		}
	}
}

func TestGlaDynamicCnfAcceptsTwoInputCut(t *testing.T) {
	for x := 0; x < 2; x++ {
		for y := 0; y < 2; y++ {
			for out := 0; out < 2; out++ {
				solver := bsat.Sat_solver2_new()
				p := &Ga2_Man_t{pSat: solver, pSops: cnf.Cnf_ReadMsops()}
				Ga2_ManCnfAddDynamic(p, 0x8888, []int32{2, 4}, 6, 1)
				got := ga2Solve(solver, bsat.Literal(2|(x^1)), bsat.Literal(4|(y^1)), bsat.Literal(6|(out^1))) == bsat.Satisfiable
				if got != (out == x*y) {
					t.Fatalf("%d & %d = %d: %v", x, y, out, got)
				}
			}
		}
	}
}
