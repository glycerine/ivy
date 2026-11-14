package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestPatternWordSimulationAndResimulation(t *testing.T) {
	p := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b^1)
	Gia_ManAppendCo(p, x^1)
	p.VSimsPi = vec.Vec_WrdAlloc(2)
	vec.Vec_WrdPush(p.VSimsPi, 0xaaaaaaaaaaaaaaaa)
	vec.Vec_WrdPush(p.VSimsPi, 0xcccccccccccccccc)
	sims := Gia_ManSimPatSim(p)
	if vec.Vec_WrdEntry(sims, int(x>>1)) != 0x2222222222222222 {
		t.Fatal("incorrect AND simulation")
	}
	values := Gia_ManSimPatValues(p)
	if vec.Vec_WrdEntry(values, 0) != ^uint64(0x2222222222222222) {
		t.Fatal("incorrect output simulation")
	}
	ids := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(ids, a>>1)
	vec.Vec_IntPush(ids, x>>1)
	Gia_ManSimPatResim(p, ids, 1, sims)
	if vec.Vec_WrdEntry(sims, int(x>>1)) != 0x1111111111111111 {
		t.Fatal("incorrect resimulation")
	}
}
