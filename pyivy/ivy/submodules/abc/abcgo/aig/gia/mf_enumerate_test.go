package gia

import "testing"

func TestMfEnumerateAndMap(t *testing.T) {
	for _, truth := range []int32{0, 1} {
		g := Gia_ManStart(12)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a, b)
		y := Gia_ManAppendAnd(g, x, c)
		Gia_ManAppendCo(g, y)
		var pars Jf_Par_t
		Mf_ManSetDefaultPars(&pars)
		pars.FCutMin = truth
		p := Mf_ManAlloc(g, &pars)
		Mf_ObjMergeOrder(p, x>>1)
		Mf_ObjMergeOrder(p, y>>1)
		if n := Mf_ManSetMapRefs(p); n == 0 {
			t.Fatal("empty mapping")
		}
		if truth == 0 {
			Mf_ManDeriveMapping(p)
			if Gia_ObjIsLut(g, y>>1) == 0 {
				t.Fatal("missing output LUT")
			}
		}
		Mf_ManFree(p)
	}
}
