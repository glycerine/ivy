package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSimulationResubstitutionCandidates(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	g.VSimsPi = vec.Vec_WrdAlloc(2)
	vec.Vec_WrdPush(g.VSimsPi, 0xaaaaaaaaaaaaaaaa)
	vec.Vec_WrdPush(g.VSimsPi, ^uint64(0))
	p := Gia_SimRsbAlloc(g)
	defer Gia_SimRsbFree(p)
	Gia_ManLevelNum(g)
	var buffers, inverters int32
	if Gia_ObjSimRsb(p, x>>1, 4, 0, &buffers, &inverters) != 1 || buffers < 1 {
		t.Fatalf("expected buffer candidate, got %d buffers and %d inverters", buffers, inverters)
	}
}
