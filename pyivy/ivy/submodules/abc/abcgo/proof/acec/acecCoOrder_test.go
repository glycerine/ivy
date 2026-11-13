package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestAdderCoreOrdering(t *testing.T) {
	p := gia.Gia_ManStart(32)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	sum := gia.Gia_ManAppendAnd(p, a, b)
	carry := gia.Gia_ManAppendAnd(p, b, c)
	gia.Gia_ManAppendCo(p, sum)
	gia.Gia_ManAppendCo(p, carry)
	adds := vec.Vec_IntAlloc(6)
	for _, x := range []int32{a >> 1, b >> 1, c >> 1, sum >> 1, carry >> 1, 0} {
		vec.Vec_IntPush(adds, x)
	}
	var ins, outs *Vec_Int_t
	order := Gia_PolynCoreOrder(p, adds, nil, &ins, &outs)
	if vec.Vec_IntSize(order) != 1 || vec.Vec_IntEntry(order, 0) != 0 {
		t.Fatalf("order %v", vec.Vec_IntArray(order))
	}
	if vec.Vec_IntSize(outs) != 2 || vec.Vec_IntEntry(outs, 0) != sum>>1 || vec.Vec_IntEntry(outs, 1) != carry>>1 {
		t.Fatalf("outputs %v", vec.Vec_IntArray(outs))
	}
	if vec.Vec_IntSize(ins) != 3 {
		t.Fatalf("inputs %v", vec.Vec_IntArray(ins))
	}
	for i, x := range []int32{a >> 1, b >> 1, c >> 1} {
		if vec.Vec_IntEntry(ins, i) != x {
			t.Fatalf("inputs %v", vec.Vec_IntArray(ins))
		}
	}
}
func TestCollectStructuralXorPairs(t *testing.T) {
	p := gia.Gia_ManStart(32)
	gia.Gia_ManHashStart(p)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManHashXor(p, a, b)
	pairs := vec.Vec_IntAlloc(2)
	Gia_PolynCoreNonXors_rec(p, gia.Gia_ManObj(p, x>>1), pairs)
	if vec.Vec_IntSize(pairs) != 2 {
		t.Fatalf("pairs %v", vec.Vec_IntArray(pairs))
	}
	ids := map[int32]bool{vec.Vec_IntEntry(pairs, 0): true, vec.Vec_IntEntry(pairs, 1): true}
	if !ids[a>>1] || !ids[b>>1] {
		t.Fatalf("pairs %v", vec.Vec_IntArray(pairs))
	}
}
