package gia

import "testing"

func TestGiarfEquivalenceDriver(t *testing.T) {
	for _, mini := range []int32{0, 1} {
		g := Gia_ManStart(12)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a, b)
		y := Gia_ManAppendAnd(g, a, b)
		z := Gia_ManAppendAnd(g, y, c)
		Gia_ManAppendCo(g, z)
		Gia_ComputeEquivalences(g, 1000, mini, 0)
		if Gia_ObjRepr(g, y>>1) != x>>1 || Gia_ObjProved(g, y>>1) == 0 {
			t.Fatalf("backend=%d: repr=%d proved=%d", mini, Gia_ObjRepr(g, y>>1), Gia_ObjProved(g, y>>1))
		}
	}
}
