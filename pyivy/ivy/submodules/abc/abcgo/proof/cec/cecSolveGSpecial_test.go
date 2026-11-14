package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestGeneralSolverModesHandleXorAndMux(t *testing.T) {
	g := gia.Gia_ManStart(40)
	g.EnableMuxes()
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendXorReal(g, a, b)
	y := gia.Gia_ManAppendXorReal(g, a, b)
	zeroX := gia.Gia_ManAppendAnd(g, x, y^1)
	m := gia.Gia_ManAppendMuxReal(g, a, b, c)
	n := gia.Gia_ManAppendMuxReal(g, a, b, c)
	zeroM := gia.Gia_ManAppendAnd(g, m, n^1)
	for mode := int32(0); mode < 3; mode++ {
		s := newCecGeneralSolver(g, &Cec_ParSat_t{SolverType: mode, NBTLimit: 1000})
		if got := s.check(zeroX); got != 1 {
			t.Fatalf("mode %d XOR contradiction = %d", mode, got)
		}
		if got := s.check(zeroM); got != 1 {
			t.Fatalf("mode %d MUX contradiction = %d", mode, got)
		}
		if got := s.check(m); got != 0 {
			t.Fatalf("mode %d MUX = %d", mode, got)
		}
	}
}
