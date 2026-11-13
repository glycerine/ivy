package cec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestSimulationInitializesCandidateClasses(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, y)
	pars := &Cec_ParSim_t{nWords: 1, nFrames: 1, nRounds: 1, nNonRefines: 1}
	if got := Cec_ManSimulationOne(g, pars); got != 0 { t.Fatalf("unexpected simulation result %d", got) }
	if !g.HasRepresentatives() || gia.Gia_ObjRepr(g, y>>1) != x>>1 {
		t.Fatal("equivalent AND nodes were not placed in the same class")
	}
	Cec_ManSimulation(g, pars)
	if gia.Gia_ObjRepr(g, y>>1) != x>>1 { t.Fatal("repeat simulation lost an equivalent class") }
}
