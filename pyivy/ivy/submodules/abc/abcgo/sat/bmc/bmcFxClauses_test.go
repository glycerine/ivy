package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestDivisorClauses(t *testing.T) {
	for _, kind := range []Div_Type_t{DIV_AND, DIV_XOR, DIV_MUX} {
		for phase := int32(0); phase < 8; phase++ {
			raw := []int32{phase & 1, 2 | ((phase >> 1) & 1), 4 | ((phase >> 2) & 1)}
			s := bsat.NewSolver()
			for i := 0; i < 4; i++ {
				s.NewVariable()
			}
			divs := vec.Vec_IntAllocArrayCopy([]int32{int32(kind), raw[0], raw[1], raw[2]}, 4)
			Bmc_FxAddClauses(s, divs, 0, 3)
			for bits := 0; bits < 16; bits++ {
				a, b, c, z := bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0
				var want bool
				switch kind {
				case DIV_AND:
					want = (a != (raw[0]&1 != 0)) && (b != (raw[1]&1 != 0))
				case DIV_XOR:
					want = a != b
				case DIV_MUX:
					if a != (raw[0]&1 != 0) {
						want = b != (raw[1]&1 != 0)
					} else {
						want = c != (raw[2]&1 != 0)
					}
				}
				assumptions := make([]bsat.Literal, 4)
				for i := range assumptions {
					assumptions[i] = bsat.Negative(int32(i))
					if bits&(1<<i) != 0 {
						assumptions[i] = bsat.Positive(int32(i))
					}
				}
				if got := s.Solve(assumptions...) == bsat.Satisfiable; got != (z == want) {
					t.Fatalf("kind=%d phase=%d bits=%04b got=%v want=%v", kind, phase, bits, got, want)
				}
			}
		}
	}
}
