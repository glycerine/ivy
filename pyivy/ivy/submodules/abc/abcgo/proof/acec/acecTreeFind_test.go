package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestFindTwoAdderCarryTree(t *testing.T) {
	p := gia.Gia_ManStart(20)
	for i := 0; i < 10; i++ {
		gia.Gia_ManAppendCi(p)
	}
	adds := vec.Vec_IntAlloc(12)
	for _, x := range []int32{1, 2, 3, 4, 5, 0, 5, 6, 7, 8, 9, 0} {
		vec.Vec_IntPush(adds, x)
	}
	trees := Acec_TreeFindTrees(p, adds, nil, 0, 0)
	if vec.Vec_WecSize(trees) != 1 {
		t.Fatalf("trees %d", vec.Vec_WecSize(trees))
	}
	tree := vec.Vec_IntArray(vec.Vec_WecEntry(trees, 0))
	if len(tree) != 4 {
		t.Fatalf("tree %v", tree)
	}
	ranks := map[int32]int32{tree[0]: tree[1], tree[2]: tree[3]}
	if ranks[0] != 0 || ranks[1] != 1 {
		t.Fatalf("ranks %v", ranks)
	}
}
func TestSharedTreePointIsExcluded(t *testing.T) {
	m := vec.Vec_IntStartFull(8)
	Acec_TreeAddInOutPoint(m, 1, 0, 0)
	Acec_TreeAddInOutPoint(m, 1, 1, 0)
	Acec_TreeAddInOutPoint(m, 1, 2, 0)
	if got := vec.Vec_IntEntry(m, 2); got != -2 {
		t.Fatalf("shared point %d", got)
	}
}
