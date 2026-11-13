package abs

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestVtaGateClauses(t *testing.T) {
	for ca := int32(0); ca < 2; ca++ {
		for cb := int32(0); cb < 2; cb++ {
			for a := int32(0); a < 2; a++ {
				for b := int32(0); b < 2; b++ {
					for v := int32(0); v < 2; v++ {
						s := bsat.NewLabeledSolver2()
						vtaSolverAddAnd(s, 1, 2, 3, ca, cb, 0, 1)
						got := s.Solve(vtaLit(1, 1-v), vtaLit(2, 1-a), vtaLit(3, 1-b)) == bsat.Satisfiable
						want := v == ((a ^ ca) & (b ^ cb))
						if got != want {
							t.Fatalf("and v=%d a=%d b=%d complements=%d,%d: got %v want %v", v, a, b, ca, cb, got, want)
						}
					}
				}
			}
		}
	}
	for complement := int32(0); complement < 2; complement++ {
		for a := int32(0); a < 2; a++ {
			for b := int32(0); b < 2; b++ {
				s := bsat.NewLabeledSolver2()
				vtaSolverAddBuffer(s, 1, 2, complement, 0, 1)
				got := s.Solve(vtaLit(1, 1-a), vtaLit(2, 1-b)) == bsat.Satisfiable
				if want := a == (b ^ complement); got != want {
					t.Fatalf("buffer a=%d b=%d complement=%d", a, b, complement)
				}
			}
		}
	}
}
