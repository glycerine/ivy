package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
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

func TestSimulationCountUnionOfRealNodes(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	p := Ssw_SmlStart(g, 0, 1, 1)
	Ssw_SmlSimInfo(p, a)[0] = 3
	Ssw_SmlSimInfo(p, b)[0] = 5
	objects := vec.Vec_PtrAlloc(2)
	vec.Vec_PtrPush(objects, a)
	vec.Vec_PtrPush(objects, b)
	if got := Ssw_SmlNodeCountOnesRealVec(p, objects); got != 3 {
		t.Fatalf("union count %d", got)
	}
}
