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
				solver := bsat.NewLabeledSolver2()
				for i := 0; i < 6; i++ {
					solver.NewVariable()
				}
				p := &Ga2_Man_t{pSat: solver, pSops: cnf.Cnf_ReadMsops()}
				Ga2_ManCnfAddDynamic(p, int32(truth), []int32{2, 4, 6, 8}, 10, 1)
				assumptions := make([]bsat.Literal, 0, 5)
				for i := 0; i < 4; i++ {
					assumptions = append(assumptions, bsat.Literal(2*(i+1)|(((inputs>>i)&1)^1)))
				}
				assumptions = append(assumptions, bsat.Literal(10|(output^1)))
				got := solver.Solve(assumptions...) == bsat.Satisfiable
				want := output == int((truth>>inputs)&1)
				if got != want {
					t.Fatalf("truth=%04x inputs=%x output=%d got=%v", truth, inputs, output, got)
				}
			}
		}
	}
}
