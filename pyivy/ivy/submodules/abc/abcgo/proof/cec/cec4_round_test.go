package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4SimulationRoundRefinesWithoutMerging(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a, c)
	gia.Gia_ManAppendCo(g, y)
	s := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)
	s.object(a >> 1)[0] = 2
	s.object(b >> 1)[0] = 2
	s.object(c >> 1)[0] = 2
	cec4SimulationRound(s, g, 0, false)
	if gia.Gia_ObjRepr(g, y>>1) != a>>1 {
		t.Fatalf("expected initial same-signature class")
	}
	s.object(a >> 1)[0] = 2
	s.object(b >> 1)[0] = 2
	s.object(c >> 1)[0] = 4
	cec4SimulationRound(s, g, 0, true)
	if gia.Gia_ObjRepr(g, x>>1) != a>>1 {
		t.Fatal("equal signature lost its class")
	}
	if gia.Gia_ObjRepr(g, y>>1) == a>>1 {
		t.Fatal("different signatures remained in one class")
	}
}
