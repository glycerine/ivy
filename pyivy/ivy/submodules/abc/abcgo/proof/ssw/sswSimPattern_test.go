package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestDistanceOnePatterns(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	p := Ssw_SmlStart(g, 0, 1, 1)
	Ssw_SmlAssignDist1(p, []uint32{1})
	if x, y := Ssw_SmlSimInfo(p, a)[0], Ssw_SmlSimInfo(p, b)[0]; x != ^uint32(2) || y != 4 {
		t.Fatalf("patterns %x %x", x, y)
	}
}
func TestSequentialPatternLayout(t *testing.T) {
	g := aig.NewGraph()
	a, lo := g.NewInput(), g.NewInput()
	g.AddOutput(a)
	aig.Aig_ManSetRegNum(g, 1)
	p := Ssw_SmlStart(g, 0, 2, 1)
	Ssw_SmlAssignDist1(p, []uint32{5})
	x, y := Ssw_SmlSimInfo(p, a), Ssw_SmlSimInfo(p, lo)
	if x[0] != ^uint32(0) || x[1] != 0 || y[0] != ^uint32(0) {
		t.Fatalf("patterns %x %x", x, y)
	}
}
