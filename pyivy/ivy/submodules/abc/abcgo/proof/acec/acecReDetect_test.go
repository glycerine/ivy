package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDetectHalfAdderCutsEndToEnd(t *testing.T) {
	p := gia.Gia_ManStart(32)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	sum := gia.Gia_ManAppendXorReal(p, a, b)
	carry := gia.Gia_ManAppendAnd(p, a, b)
	gia.Gia_ManAppendCo(p, sum)
	gia.Gia_ManAppendCo(p, carry)
	var xors *Vec_Int_t
	adds := Ree_ManComputeCuts(p, &xors, 0)
	if vec.Vec_IntSize(adds) != 6 || vec.Vec_IntEntry(adds, 3) != sum>>1 || vec.Vec_IntEntry(adds, 4) != carry>>1 {
		t.Fatalf("adders %v", vec.Vec_IntArray(adds))
	}
	if vec.Vec_IntSize(xors) != 4 {
		t.Fatalf("XORs %v", vec.Vec_IntArray(xors))
	}
}
func TestDetectFullAdderCutsEndToEnd(t *testing.T) {
	p := gia.Gia_ManStart(64)
	in := []int32{gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)}
	out := make([]int32, 2)
	Acec_InsertFadd(p, in, out)
	gia.Gia_ManAppendCo(p, out[0])
	gia.Gia_ManAppendCo(p, out[1])
	adds := Ree_ManComputeCuts(p, nil, 0)
	if Ree_ManCountFadds(adds) == 0 {
		t.Fatalf("failed to detect full adder: %v", vec.Vec_IntArray(adds))
	}
}
