package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestXorRootRankingAndSmallOrders(t *testing.T) {
	p := gia.Gia_ManStart(20)
	for i := 0; i < 10; i++ {
		gia.Gia_ManAppendCi(p)
	}
	xors := vec.Vec_IntAlloc(8)
	for _, v := range []int32{4, 1, 2, 3, 8, 4, 5, 6} {
		vec.Vec_IntPush(xors, v)
	}
	roots := Acec_FindXorRoots(p, xors)
	if vec.Vec_IntSize(roots) != 1 || vec.Vec_IntEntry(roots, 0) != 8 {
		t.Fatalf("roots %v", vec.Vec_IntArray(roots))
	}
	ranks := Acec_RankTrees(p, xors, roots)
	for _, id := range []int32{1, 2, 3, 4, 5, 6, 8} {
		if vec.Vec_IntEntry(ranks, int(id)) != 0 {
			t.Fatalf("rank for %d", id)
		}
	}
	adds := vec.Vec_IntAlloc(0)
	order := Acec_OrderTreeRoots(p, adds, roots, ranks)
	if vec.Vec_IntSize(order) != 1 || vec.Vec_IntEntry(order, 0) != 8 {
		t.Fatalf("single root order %v", vec.Vec_IntArray(order))
	}
	empty := Acec_OrderTreeRoots(p, adds, vec.Vec_IntAlloc(0), ranks)
	if vec.Vec_IntSize(empty) != 0 {
		t.Fatal("nonempty empty order")
	}
}
