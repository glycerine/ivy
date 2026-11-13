package abs

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestVtaConstAndConstraintClauses(t *testing.T) {
	for complement := int32(0); complement < 2; complement++ {
		for value := int32(0); value < 2; value++ {
			s := bsat.NewLabeledSolver2()
			vtaSolverAddConst(s, 1, complement, 0, 1)
			got := s.Solve(vtaLit(1, 1-value)) == bsat.Satisfiable
			if want := value != complement; got != want {
				t.Fatalf("constant value=%d complement=%d", value, complement)
			}
		}
		for v := int32(0); v < 2; v++ {
			for w := int32(0); w < 2; w++ {
				s := bsat.NewLabeledSolver2()
				vtaSolverAddConstraint(s, 1, 2, complement, 0, 1)
				got := s.Solve(vtaLit(1, 1-v), vtaLit(2, 1-w)) == bsat.Satisfiable
				if want := (v^complement) != 0 || w != 0; got != want {
					t.Fatalf("constraint v=%d w=%d complement=%d", v, w, complement)
				}
			}
		}
	}
}
