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
	for _, r := range []*Vec_Int_t{roots, vec.Vec_IntAlloc(0)} {
		func() {
			defer func() {
				if got := recover(); got != "acec: root ordering requires an XOR chain" {
					t.Errorf("root-chain panic = %v", got)
				}
			}()
			Acec_OrderTreeRoots(p, adds, r, ranks)
		}()
	}
}

// C follows the first linked rank chain and omits disconnected roots.
func TestSourceOrderTreeRootsOmitsDisconnected(t *testing.T) {
	p := gia.Gia_ManStart(8)
	adds := vec.Vec_IntAlloc(6)
	roots := vec.Vec_IntAlloc(3)
	ranks := vec.Vec_IntStartFull(3)
	for _, v := range []int32{0, 0, 1, 0, 2, 0} {
		vec.Vec_IntPush(adds, v)
	}
	for _, v := range []int32{10, 11, 12} {
		vec.Vec_IntPush(roots, v)
	}
	vec.Vec_IntWriteEntry(ranks, 1, 0)
	vec.Vec_IntWriteEntry(ranks, 2, 1)
	got := vec.Vec_IntArray(Acec_OrderTreeRoots(p, adds, roots, ranks))
	if len(got) != 2 || got[0] != 10 || got[1] != 11 {
		t.Fatalf("order %v; ABC reports [10 11]", got)
	}
}
