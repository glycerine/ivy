package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestTreeOverlapFilters(t *testing.T) {
	p := gia.Gia_ManStart(32)
	a, b, c, d := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x, y := gia.Gia_ManAppendAnd(p, a, b), gia.Gia_ManAppendAnd(p, a, c)
	z := gia.Gia_ManAppendAnd(p, x, d)
	out0, out1 := gia.Gia_ManAppendAnd(p, z, b), gia.Gia_ManAppendAnd(p, z, c)
	adds := vec.Vec_IntAlloc(12)
	for _, v := range []int32{a >> 1, b >> 1, c >> 1, x >> 1, y >> 1, 0, z >> 1, b >> 1, c >> 1, out0 >> 1, out1 >> 1, 0} {
		vec.Vec_IntPush(adds, v)
	}
	tree := vec.Vec_IntAlloc(4)
	for _, v := range []int32{0, 0, 1, 1} {
		vec.Vec_IntPush(tree, v)
	}
	t1, t2 := vec.Vec_IntDup(tree), vec.Vec_IntDup(tree)
	Acec_TreeFilterOne(p, adds, t1)
	if vec.Vec_IntSize(t1) != 2 || vec.Vec_IntEntry(t1, 0) != 0 {
		t.Fatalf("output overlap filter %v", vec.Vec_IntArray(t1))
	}
	Acec_TreeFilterOne2(p, adds, t2)
	if vec.Vec_IntSize(t2) != 2 || vec.Vec_IntEntry(t2, 0) != 1 {
		t.Fatalf("input overlap filter %v", vec.Vec_IntArray(t2))
	}
}
