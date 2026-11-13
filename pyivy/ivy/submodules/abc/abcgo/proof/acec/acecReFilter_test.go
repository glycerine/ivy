package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRemoveTrivialHalfAdder(t *testing.T) {
	p := gia.Gia_ManStart(32)
	gia.Gia_ManHashStart(p)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManHashXor(p, a, b)
	gia.Gia_ManHashStop(p)
	o := gia.Gia_ManObj(p, x>>1)
	internal := gia.Gia_ObjFaninId0(o, x>>1)
	adds := vec.Vec_IntAlloc(6)
	for _, v := range []int32{a >> 1, b >> 1, 0, x >> 1, internal, 0} {
		vec.Vec_IntPush(adds, v)
	}
	Ree_ManRemoveTrivial(p, adds)
	if vec.Vec_IntSize(adds) != 0 {
		t.Fatalf("trivial adder retained: %v", vec.Vec_IntArray(adds))
	}
}
func TestCountFullAdders(t *testing.T) {
	v := vec.Vec_IntAlloc(12)
	for _, x := range []int32{1, 2, 0, 3, 4, 0, 1, 2, 3, 5, 6, 0} {
		vec.Vec_IntPush(v, x)
	}
	if got := Ree_ManCountFadds(v); got != 1 {
		t.Fatalf("full adders %d", got)
	}
}
