package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4InitialClassesRespectLevelAndSimulation(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, x, a)
	out := gia.Gia_ManAppendCo(g, y)
	gia.Gia_ManLevelNum(g)
	s := newCec3Simulation(int(gia.Gia_ManObjNum(g)), 1)
	s.object(a >> 1)[0] = 2
	s.object(b >> 1)[0] = ^uint64(2)
	s.object(x >> 1)[0] = 4
	s.object(y >> 1)[0] = 2
	cec4CreateClasses(s, g, 1)
	if got := gia.Gia_ObjRepr(g, b>>1); got != a>>1 {
		t.Fatalf("complemented signature representative=%d", got)
	}
	if got := gia.Gia_ObjRepr(g, y>>1); got != gia.GIA_VOID {
		t.Fatalf("high-level node representative=%d", got)
	}
	if got := gia.Gia_ObjRepr(g, out>>1); got != gia.GIA_VOID {
		t.Fatalf("output representative=%d", got)
	}
}
