package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestOriginalIDEquivalenceSweep(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	ab := gia.Gia_ManAppendAnd(p, a, b)
	x := gia.Gia_ManAppendAnd(p, ab, c)
	bc := gia.Gia_ManAppendAnd(p, b, c)
	y := gia.Gia_ManAppendAnd(p, a, bc)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManAppendCo(p, y)
	q := Gia_ManComputeGiaEquivs(p, 1000, 0)
	if q == nil {
		t.Fatal("sweep failed")
	}
	if gia.Gia_ObjHasRepr(p, y>>1) == 0 {
		t.Fatal("equivalence was not recorded")
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != (in[0] && in[1] && in[2]) || out[1] != out[0] {
			t.Fatalf("outputs %v error %v", out, err)
		}
	}
}
