package gia

import "testing"

func TestMfDeriveTruthMapping(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	pars.FCutMin = 1
	p := Mf_ManAlloc(g, &pars)
	Mf_ObjMergeOrder(p, x>>1)
	Mf_ManSetMapRefs(p)
	mapped := Mf_ManDeriveMappingGia(p)
	if Gia_ManPoNum(mapped) != 1 || Gia_ManLutNum(mapped) != 1 {
		t.Fatal("missing mapped output")
	}
	Gia_ManMappingVerify(mapped)
}

func TestMfDeriveCoarseMapping(t *testing.T) {
	g := Gia_ManStart(12)
	g.EnableMuxes()
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendMuxReal(g, a, b, c)
	Gia_ManAppendCo(g, x)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	p := Mf_ManAlloc(g, &pars)
	Mf_ObjMergeOrder(p, x>>1)
	Mf_ManSetMapRefs(p)
	mapped := Mf_ManDeriveMappingCoarse(p)
	if Gia_ManPoNum(mapped) != 1 || Gia_ManLutNum(mapped) != 1 {
		t.Fatal("missing coarse LUT")
	}
	Gia_ManMappingVerify(mapped)
}
