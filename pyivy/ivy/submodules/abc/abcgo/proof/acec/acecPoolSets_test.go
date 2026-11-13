package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCarryRootGroupingThroughXor(t *testing.T) {
	p := gia.Gia_ManStart(20)
	for i := 0; i < 11; i++ {
		gia.Gia_ManAppendCi(p)
	}
	adds := vec.Vec_IntAlloc(12)
	for _, x := range []int32{1, 2, 3, 4, 5, 0, 5, 6, 7, 8, 9, 0} {
		vec.Vec_IntPush(adds, x)
	}
	xors := vec.Vec_IntAlloc(4)
	for _, x := range []int32{10, 9, 6, 7} {
		vec.Vec_IntPush(xors, x)
	}
	cmap := Acec_ManCreateCarryMap(p, adds)
	xr := Acec_ManCollectXorRoots(p, xors)
	cr := Acec_ManCollectCarryRoots(p, adds)
	sets := Acec_ManCollectCarryRootSets(p, adds, cmap, xors, xr, cr)
	if vec.Vec_WecSize(sets) != 5 {
		t.Fatalf("levels %d", vec.Vec_WecSize(sets))
	}
	if vec.Vec_IntEntry(vec.Vec_WecEntry(sets, 0), 0) != 10 || vec.Vec_IntEntry(vec.Vec_WecEntry(sets, 1), 0) != 9 {
		t.Fatalf("unexpected group XOR=%v carries=%v", vec.Vec_IntArray(vec.Vec_WecEntry(sets, 0)), vec.Vec_IntArray(vec.Vec_WecEntry(sets, 1)))
	}
}
