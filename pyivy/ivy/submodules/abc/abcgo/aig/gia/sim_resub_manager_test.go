package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSimulationResubstitutionCareAndFunction(t *testing.T) {
	g := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, y)
	g.VSimsPi = vec.Vec_WrdAlloc(3)
	for _, word := range []uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xffffffffffffffff} {
		vec.Vec_WrdPush(g.VSimsPi, word)
	}
	p := Gia_SimRsbAlloc(g)
	defer Gia_SimRsbFree(p)
	tfo := Gia_SimRsbTfo(p, x>>1, -1)
	care := Gia_SimRsbCareSet(p, x>>1, tfo)
	if care[0] != ^uint64(0) {
		t.Fatalf("care %x", care[0])
	}
	fanins := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(fanins, a>>1)
	if Gia_SimRsbResubVerify(p, x>>1, fanins) != 0 {
		t.Fatal("single input cannot determine AND")
	}
	vec.Vec_IntPush(fanins, b>>1)
	if Gia_SimRsbResubVerify(p, x>>1, fanins) != 1 {
		t.Fatal("both inputs should determine AND")
	}
}
