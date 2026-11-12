package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGiaDominatorSupport(t *testing.T) {
	g := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	co := Gia_ManAppendCo(g, y)
	Gia_ManCreateRefs(g)
	Gia_ManComputeDoms(g)
	if got := Gia_ObjDom(g, Gia_ManObj(g, x>>1)); got != y>>1 {
		t.Fatalf("dominator %d output %d", got, co>>1)
	}
	supp, refs := vec.Vec_IntAlloc(3), vec.Vec_IntAlloc(3)
	if size := Gia_NodeMffcSizeSuppRefs(g, Gia_ManObj(g, y>>1), supp, refs); size != 2 || vec.Vec_IntSize(supp) != 3 {
		t.Fatalf("size %d support %v", size, vec.Vec_IntArray(supp))
	}
	d := Gia_ManDomDerive(g, Gia_ManObj(g, y>>1), supp, 1)
	if Gia_ManPoNum(d) != 2 {
		t.Fatal("wrong cofactor count")
	}
}
