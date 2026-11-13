package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4PreparationCapturesAllZeroWitness(t *testing.T) {
	g := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 1)
	p := Cec_ParFra_t{nWords: 1, nItersMax: 1, fCheckMiter: 1}
	_, cex := cec4PrepareSimulation(g, &p, true)
	if cex == nil || cex.IPo != 0 || cex.PData[0] != 0 {
		t.Fatalf("unexpected witness %+v", cex)
	}
}
func TestCec4PreparationInitializesClasses(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, y)
	p := Cec_ParFra_t{nWords: 1, nItersMax: 1}
	pack, cex := cec4PrepareSimulation(g, &p, true)
	if cex != nil || pack == nil || gia.Gia_ObjRepr(g, y>>1) != x>>1 {
		t.Fatalf("classes not initialized: witness %+v", cex)
	}
}
