package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4EntryStoresZeroPatternMiterWitness(t *testing.T) {
	g := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 1)
	p := Cec_ParFra_t{nWords: 1, nItersMax: 1, fCheckMiter: 1}
	if status := Cec4_ManPerformSweeping(g, &p, nil, 1); status != 0 {
		t.Fatalf("status %d", status)
	}
	cex := g.SequentialCounterexample()
	if cex == nil || cex.IPo != 0 || cex.PData[0] != 0 {
		t.Fatalf("witness %+v", cex)
	}
}
func TestCec4EntryReturnsEquivalentReducedGraph(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, x, a)
	gia.Gia_ManAppendCo(g, y)
	p := Cec_ParFra_t{nWords: 1, nRounds: 1, nItersMax: 1, nBTLimit: 100, jType: 0}
	out := Cec4_ManSimulateTest(g, &p)
	if out == nil {
		t.Fatal("missing reduced graph")
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		want, e := g.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		got, e := out.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		if got[0] != want[0] {
			t.Fatalf("%v got %v want %v", in, got, want)
		}
	}
}
