package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestShowOrderAndAdderMapping(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	adds := vec.Vec_IntAlloc(6)
	for _, v := range []int32{a >> 1, b >> 1, 0, x >> 1, y >> 1, 0} {
		vec.Vec_IntPush(adds, v)
	}
	xors := vec.Vec_IntAlloc(0)
	mapping := Gia_ShowMapAdds(p, adds, 0, nil)
	if Gia_ShowAddOut(adds, mapping, x>>1) != y>>1 {
		t.Fatal("adder outputs not merged")
	}
	order := Gia_ShowCollectObjs(p, adds, xors, mapping, Gia_ShowMapXors(p, xors))
	if vec.Vec_IntSize(order) != 1 || vec.Vec_IntEntry(order, 0) != y>>1 || Gia_ObjLevelId(p, y>>1) != 1 {
		t.Fatalf("wrong box order or level: %v", vec.Vec_IntArray(order))
	}
	fullOnly := Gia_ShowMapAdds(p, adds, 1, nil)
	if Gia_ShowAddOut(adds, fullOnly, x>>1) != x>>1 {
		t.Fatal("half adder was not filtered")
	}
	bold := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(bold, y>>1)
	filtered := Gia_ShowMapAdds(p, adds, 0, bold)
	order = Gia_ShowCollectObjs(p, adds, xors, filtered, Gia_ShowMapXors(p, xors))
	if vec.Vec_IntSize(order) != 2 || Gia_ObjLevelId(p, y>>1) != 2 {
		t.Fatalf("wrong ordinary order: %v", vec.Vec_IntArray(order))
	}
}
