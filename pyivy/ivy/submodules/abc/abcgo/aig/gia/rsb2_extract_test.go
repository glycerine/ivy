package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRsb2ExtractWindow(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	Gia_ManStaticFanoutStart(p)
	defer Gia_ManStaticFanoutStop(p)
	window := vec.Vec_IntAlloc(3)
	for _, id := range []int32{a >> 1, b >> 1, x >> 1} {
		vec.Vec_IntPush(window, id)
	}
	inputs := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(inputs, a>>1)
	vec.Vec_IntPush(inputs, b>>1)
	refs := vec.Vec_IntStart(int(p.nObjs))
	outs := Gia_RsbFindOutputs(p, window, inputs, refs)
	if vec.Vec_IntSize(outs) != 1 || vec.Vec_IntEntry(outs, 0) != x>>1 {
		t.Fatalf("outputs %v", vec.Vec_IntArray(outs))
	}
	for _, v := range vec.Vec_IntArray(refs) {
		if v != 0 {
			t.Fatal("reference scratch changed")
		}
	}
	q := Gia_RsbDeriveGiaFromWindows(p, window, inputs, outs)
	if Gia_ManVerifyTwoTruths(p, q) != 1 {
		t.Fatal("window function changed")
	}
}
