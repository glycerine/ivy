package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestSolverPhaseNormalization(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	n := gia.Gia_ManAppendAnd(g, a^1, b^1)
	gia.Gia_ManAppendCo(g, n)
	for _, flip := range []int32{0, 1} {
		pars := &Cec_ParSat_t{NBTLimit: 1000, fPolarFlip: flip}
		classic := newCecClassicSolver(g, pars)
		if classic.check(n) != 0 || classic.vars[n>>1]&1 != flip {
			t.Fatalf("classic polarity flip %d: %d", flip, classic.vars[n>>1])
		}
		pars.SolverType = 0
		general := newCecGeneralSolver(g, pars)
		if general.check(n) != 0 || general.vars[n>>1]&1 != flip {
			t.Fatalf("general polarity flip %d: %d", flip, general.vars[n>>1])
		}
	}
}
