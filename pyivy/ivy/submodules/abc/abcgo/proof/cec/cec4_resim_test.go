package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4LocalResimulationUpdatesSharedCone(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, x, a)
	s := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)
	stamps := make([]int32, gia.Gia_ManObjNum(g))
	s.object(a >> 1)[0] = 2
	s.object(b >> 1)[0] = 2
	cec4ResimulateNode(s, g, y>>1, 1, stamps)
	if s.object(y >> 1)[0] != 2 || stamps[x>>1] != 1 {
		t.Fatal("initial cone not simulated")
	}
	s.object(b >> 1)[0] = 0
	cec4ResimulateNode(s, g, y>>1, 2, stamps)
	if s.object(y >> 1)[0] != 0 || stamps[x>>1] != 2 {
		t.Fatal("updated cone not resimulated")
	}
}
