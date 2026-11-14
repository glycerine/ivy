package gia

import "testing"

func TestOfDeriveMapping(t *testing.T) {
	for _, fast := range []int32{0, 1} {
		g := Gia_ManStart(10)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a, b)
		out := Gia_ManAppendAnd(g, x, c)
		Gia_ManAppendCo(g, out)
		var pars Jf_Par_t
		Mf_ManSetDefaultPars(&pars)
		pars.NLutSize = 2
		pars.NDelayLut1 = 10
		pars.NDelayLut2 = 2
		pars.NFastEdges = fast
		pars.FCutMin = 0
		cuts := newOfCutManager(g, &pars)
		cuts.computeCuts()
		mapper := newOfMapper(cuts)
		if fast == 0 {
			mapper.forward()
			mapper.backward()
		} else {
			mapper.forwardDirect()
			mapper.backwardDirect()
		}
		mapper.deriveMapping()
		Gia_ManMappingVerify(g)
		if Gia_ObjIsLut(g, out>>1) == 0 || Gia_ObjLutSize(g, out>>1) != 2 {
			t.Fatal("missing output LUT")
		}
		if fast != 0 {
			Gia_ManConvertPackingToEdges(g)
			if Gia_ObjHaveEdge(g, x>>1, out>>1) == 0 {
				t.Fatal("missing fast edge")
			}
		}
	}
}

func TestOfDriverMapsNetwork(t *testing.T) {
	for _, fast := range []int32{0, 1} {
		g := Gia_ManStart(10)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a, b)
		out := Gia_ManAppendAnd(g, x, c)
		Gia_ManAppendCo(g, out)
		var pars Jf_Par_t
		ofDefaultParameters(&pars)
		pars.NLutSize = 2
		pars.NFastEdges = fast
		result := ofMap(g, &pars)
		if result != g || Gia_ObjIsLut(g, out>>1) == 0 || pars.Area != 2 {
			t.Fatalf("fast=%d: invalid result", fast)
		}
		if fast != 0 && Gia_ObjHaveEdge(g, x>>1, out>>1) == 0 {
			t.Fatal("missing fast edge")
		}
	}
}
