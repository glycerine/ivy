package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSimulationCountsRespectPrefixAndPolarity(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	p := Ssw_SmlStart(g, 1, 1, 1)
	x, y := Ssw_SmlSimInfo(p, a), Ssw_SmlSimInfo(p, b)
	x[0], x[1] = 1, 0xf
	y[0], y[1] = 0, 3
	if n := Ssw_SmlNodeNotEquWeight(p, int32(a.NodeID()), int32(b.NodeID())); n != 2 {
		t.Fatalf("different bits: %d", n)
	}
	if n := Ssw_SmlCountEqual(p, a, b); n != 30 {
		t.Fatalf("equal bits: %d", n)
	}
	if n := Ssw_SmlNodeCountOnesReal(p, a); n != 5 {
		t.Fatalf("ones: %d", n)
	}
	if n := Ssw_SmlNodeCountOnesReal(p, a.Not()); n != 59 {
		t.Fatalf("complemented ones: %d", n)
	}
	if Ssw_SmlNodeIsZero(p, a) != 0 || Ssw_SmlNodeIsZeroFrame(p, b, 0) != 1 {
		t.Fatal("wrong zero test")
	}
}
