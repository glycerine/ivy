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
