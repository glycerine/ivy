package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestResubUniversalTFO(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	co := Gia_ManAppendCo(g, x)
	Gia_ManStaticFanoutStart(g)
	defer Gia_ManStaticFanoutStop(g)
	var nodes, outputs *Vec_Int_t
	if count := Gia_ManUnivTfo(g, []int32{a >> 1, b >> 1}, &nodes, &outputs); count != 4 || vec.Vec_IntSize(outputs) != 1 || vec.Vec_IntEntry(outputs, 0) != co>>1 {
		t.Fatalf("count %d outputs %v", count, vec.Vec_IntArray(outputs))
	}
}
