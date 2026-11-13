package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestEnumerateTwoInputAdderCuts(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	sum := gia.Gia_ManAppendXorReal(p, a, b)
	carry := gia.Gia_ManAppendAnd(p, a, b)
	h := vec.Hash_IntManStart(4)
	data, xors := vec.Vec_IntAlloc(6), vec.Vec_IntAlloc(4)
	cuts := vec.Vec_IntAlloc(16)
	left, right := []int32{1, 1, a >> 1, 0xaa}, []int32{1, 1, b >> 1, 0xaa}
	Ree_ManCutMerge(p, sum>>1, left, right, cuts, h, data, xors)
	if vec.Vec_IntSize(xors) != 4 || vec.Vec_IntEntry(xors, 0) != sum>>1 {
		t.Fatalf("XOR cuts %v", vec.Vec_IntArray(xors))
	}
	Ree_ManCutMerge(p, carry>>1, left, right, cuts, h, data, xors)
	adds := Ree_ManDeriveAdds(h, data, 0)
	if vec.Vec_IntSize(adds) != 6 || vec.Vec_IntEntry(adds, 3) != sum>>1 || vec.Vec_IntEntry(adds, 4) != carry>>1 {
		t.Fatalf("adder records %v", vec.Vec_IntArray(adds))
	}
}
