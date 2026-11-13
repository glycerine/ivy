package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestGeneralSolverModes(t *testing.T) {
	for mode := int32(0); mode < 3; mode++ {
		t.Run(string(rune('0'+mode)), func(t *testing.T) {
			g := gia.Gia_ManStart(32)
			a := gia.Gia_ManAppendCi(g)
			b := gia.Gia_ManAppendCi(g)
			x := gia.Gia_ManAppendAnd(g, a, b)
			y := gia.Gia_ManAppendAnd(g, a^1, b)
			zero := gia.Gia_ManAppendAnd(g, x, y)
			s := newCecGeneralSolver(g, &Cec_ParSat_t{SolverType: mode, nBTLimit: 1000, nSatVarMax: 2, nCallsRecycle: 1})
			if got := s.check(zero); got != 1 {
				t.Fatalf("contradiction result %d", got)
			}
			if got := s.check(x); got != 0 {
				t.Fatalf("AND result %d", got)
			}
			if got := s.check(zero); got != 1 {
				t.Fatalf("recycled contradiction %d", got)
			}
		})
	}
}
