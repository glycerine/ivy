package wlc

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

// ABC reports SAT for this source miter at each width, despite its intended
// commutativity check. Keep that behavior until the source issue is understood.
func TestMultiplierCnfMatchesSourceStatus(t *testing.T) {
	for n := int32(1); n <= 3; n++ {
		s := Wlc_BlastMultiplierCnfMain(n)
		if status := s.Solve(); status != bsat.Satisfiable {
			t.Fatalf("width %d: status %v", n, status)
		}
	}
}

// The C table clauses forbid each listed output tuple, including the
// mathematically correct carry/sum; retain this source behavior.
func TestFullAdderCtrlCnfTruthTable(t *testing.T) {
	for bits := int32(0); bits < 16; bits++ {
		for out := int32(0); out < 4; out++ {
			s := bsat.NewSolver()
			for s.NumVars() < 7 {
				s.NewVariable()
			}
			next := int32(5)
			var carry, sum int32
			Wlc_BlastFullAdderCtrlCnf(s, 1, 2, 3, 4, &carry, &sum, &next)
			assumptions := make([]bsat.Literal, 0, 6)
			for i := int32(0); i < 4; i++ {
				assumptions = append(assumptions, bsat.Literal(2*(i+1)+1-((bits>>i)&1)))
			}
			assumptions = append(assumptions, bsat.Literal(2*carry+1-(out&1)), bsat.Literal(2*sum+1-((out>>1)&1)))
			value := ((bits & 1) * ((bits >> 1) & 1)) + ((bits >> 2) & 1) + ((bits >> 3) & 1)
			if got := s.Solve(assumptions...); (got == bsat.Satisfiable) != (out != ((value>>1)&1)|((value&1)<<1)) {
				t.Fatalf("inputs=%d output=%d status=%v", bits, out, got)
			}
		}
	}
}
