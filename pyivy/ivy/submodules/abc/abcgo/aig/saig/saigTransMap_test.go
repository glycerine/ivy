package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestFrameMappingsLifecycle(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	Saig_ManStartMap1(g, 2)
	Saig_ManStartMap2(g, 2)
	if Saig_ManGetMap1(g, a, 1) != -1 {
		t.Fatal("map not initialized")
	}
	var frame int32
	if Saig_ManGetMap2(g, a, 1, &frame) != -1 || frame != -1 {
		t.Fatal("map pair not initialized")
	}
	Saig_ManSetMap1(g, a, 1, b)
	Saig_ManSetMap2(g, a, 1, b, 7)
	if Saig_ManGetMap1(g, a, 1) != int32(b.NodeID()) || Saig_ManGetMap2(g, a, 1, &frame) != int32(b.NodeID()) || frame != 7 {
		t.Fatal("incorrect mapping")
	}
	Saig_ManStopMap1(g)
	if Saig_ManHasMap1(g) != 0 || Saig_ManHasMap2(g) != 1 {
		t.Fatal("map lifecycle")
	}
	Saig_ManStopMap2(g)
	if g.Data != nil {
		t.Fatal("data not released")
	}
}

func TestCreateFrameMappingUsesFirstEquivalentObject(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	Saig_ManStartMap1(g, 2)
	defer Saig_ManStopMap1(g)
	frames := aig.NewGraph()
	x, y := frames.NewInput(), frames.NewInput()
	aig.Aig_ManReprStart(frames, int32(len(frames.Nodes())))
	aig.Aig_ObjCreateRepr(frames, x, y)
	Saig_ManSetMap1(g, a, 0, x)
	Saig_ManSetMap1(g, b, 1, y)
	Saig_ManCreateMapping(g, frames, 2)
	defer Saig_ManStopMap2(g)
	var frame int32
	if got := Saig_ManGetMap2(g, b, 1, &frame); got != int32(a.NodeID()) || frame != 0 {
		t.Fatalf("got object %d frame %d", got, frame)
	}
	if got := Saig_ManGetMap2(g, b, 0, &frame); got != int32(b.NodeID()) || frame != 0 {
		t.Fatalf("unmapped got object %d frame %d", got, frame)
	}
}
