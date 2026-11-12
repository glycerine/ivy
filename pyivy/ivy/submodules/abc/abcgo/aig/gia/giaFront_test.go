package gia

import "testing"

func TestFrontierRoundTrip(t *testing.T) {
	for mask := 0; mask < 16; mask++ {
		p := Gia_ManStart(20)
		a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendAnd(p, a^int32(mask&1), b^int32(mask>>1&1))
		y := Gia_ManAppendAnd(p, x^int32(mask>>2&1), c^int32(mask>>3&1))
		Gia_ManAppendCo(p, y)
		Gia_ManAppendCo(p, x^1)
		Gia_ManAppendCo(p, 0)
		Gia_ManAppendCi(p) // dangling CI
		if cut := Gia_ManCrossCutSimple(p); cut < 1 {
			t.Fatalf("mask %d: invalid cross-cut %d", mask, cut)
		}
		q := Gia_ManFront(p)
		Gia_ManFrontTransform(q)
		for i, want := range p.pObjs {
			got := q.pObjs[i]
			if got.iDiff0 != want.iDiff0 || got.iDiff1 != want.iDiff1 || got.fCompl0 != want.fCompl0 || got.fCompl1 != want.fCompl1 || got.fTerm != want.fTerm || got.Value != want.Value {
				t.Fatalf("mask %d object %d: got (%d %d %d %d %d %d), want (%d %d %d %d %d %d)", mask, i, got.iDiff0, got.iDiff1, got.fCompl0, got.fCompl1, got.fTerm, got.Value, want.iDiff0, want.iDiff1, want.fCompl0, want.fCompl1, want.fTerm, want.Value)
			}
		}
	}
}
