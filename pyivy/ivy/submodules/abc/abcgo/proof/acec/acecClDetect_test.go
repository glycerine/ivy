package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCollectTopXorInputs(t *testing.T) {
	p := gia.Gia_ManStart(64)
	gia.Gia_ManHashStart(p)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	root := gia.Gia_ManHashXor(p, gia.Gia_ManHashXor(p, a, b), c)
	gia.Gia_ManAppendCo(p, root)
	got := Acec_CollectXorTops(p)
	if got == nil || vec.Vec_IntSize(got) != 4 || vec.Vec_IntEntry(got, 0) != root>>1 {
		t.Fatalf("roots %v", got)
	}
	seen := map[int32]bool{}
	for i := 1; i < 4; i++ {
		seen[vec.Vec_IntEntry(got, i)] = true
	}
	for _, x := range []int32{a, b, c} {
		if !seen[x>>1] {
			t.Errorf("missing input %d", x>>1)
		}
	}
}
func TestCollectMarkedSupport(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	root := gia.Gia_ManAppendAnd(p, a, b)
	gia.Gia_ObjSetMark0(gia.Gia_ManObj(p, a>>1), 1)
	gia.Gia_ObjSetMark0(gia.Gia_ManObj(p, b>>1), 1)
	gia.Gia_ManIncrementTravId(p)
	supp, nodes := vec.Vec_IntAlloc(2), vec.Vec_IntAlloc(1)
	Acec_DetectComputeSuppOne_rec(p, gia.Gia_ManObj(p, root>>1), supp, nodes)
	if vec.Vec_IntSize(supp) != 2 || vec.Vec_IntSize(nodes) != 1 || vec.Vec_IntEntry(nodes, 0) != root>>1 {
		t.Fatalf("support %v nodes %v", vec.Vec_IntArray(supp), vec.Vec_IntArray(nodes))
	}
}
