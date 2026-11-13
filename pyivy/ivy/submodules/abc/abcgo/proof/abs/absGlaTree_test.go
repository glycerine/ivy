package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGlaTreeBreakAndCollect(t *testing.T) {
	p := gia.Gia_ManStart(10)
	a := gia.Gia_ManAppendCi(p)
	b := gia.Gia_ManAppendCi(p)
	c := gia.Gia_ManAppendCi(p)
	n := gia.Gia_ManAppendAnd(p, a, b)
	root := gia.Gia_ManAppendAnd(p, n, c)
	for _, lit := range []int32{a, b, c} {
		gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(p, lit>>1), 1)
	}
	obj := gia.Gia_ManObj(p, root>>1)
	if got := Ga2_ManBreakTree_rec(p, obj, 1, 3); got != 1 {
		t.Fatalf("break returned %d", got)
	}
	nodes := vec.Vec_IntAlloc(4)
	leaves := vec.Vec_IntAlloc(4)
	Ga2_ManCollectNodes_rec(p, obj, nodes, 1)
	Ga2_ManCollectLeaves_rec(p, obj, leaves, 1)
	if vec.Vec_IntSize(nodes) != 2 || vec.Vec_IntSize(leaves) != 3 || Ga2_ManCheckNodesAnd(p, nodes) != 1 {
		t.Fatalf("nodes %d leaves %d", vec.Vec_IntSize(nodes), vec.Vec_IntSize(leaves))
	}
}
