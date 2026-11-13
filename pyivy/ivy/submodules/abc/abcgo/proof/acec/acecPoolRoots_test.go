package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestAdderCarryAndXorRoots(t *testing.T) {
	p := gia.Gia_ManStart(20)
	for i := 0; i < 10; i++ {
		gia.Gia_ManAppendCi(p)
	}
	adds := vec.Vec_IntAlloc(12)
	for _, x := range []int32{1, 2, 3, 4, 5, 0, 5, 6, 7, 8, 9, 0} {
		vec.Vec_IntPush(adds, x)
	}
	m := Acec_ManCreateCarryMap(p, adds)
	if Acec_ManCheckCarryMap(p, 9, adds, m) != 1 || Acec_ManCheckCarryMap(p, 5, adds, m) != 0 {
		t.Fatal("bad carry map")
	}
	roots := Acec_ManCollectCarryRoots(p, adds)
	if vec.Vec_IntSize(roots) != 1 || vec.Vec_IntEntry(roots, 0) != 9 {
		t.Fatalf("carry roots %v", vec.Vec_IntArray(roots))
	}
	tops := Acec_ManPoolTopMost(p, adds)
	if vec.Vec_IntSize(tops) != 1 || vec.Vec_IntEntry(tops, 0) != 1 {
		t.Fatalf("top boxes %v", vec.Vec_IntArray(tops))
	}
	xors := vec.Vec_IntAlloc(8)
	for _, x := range []int32{4, 1, 2, 3, 8, 4, 5, 6} {
		vec.Vec_IntPush(xors, x)
	}
	xr := Acec_ManCollectXorRoots(p, xors)
	if vec.Vec_IntSize(xr) != 1 || vec.Vec_IntEntry(xr, 0) != 8 {
		t.Fatalf("xor roots %v", vec.Vec_IntArray(xr))
	}
}
