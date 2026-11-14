package gia

import "testing"

func TestGiaMfsRoundTrip(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	pars.FCoarsen = 0
	mapped := Mf_ManPerformMapping(g, &pars)
	data := giaMfsExtract(mapped)
	result := giaMfsInsert(mapped, data, false)
	if Gia_ManCiNum(result) != 2 || Gia_ManCoNum(result) != 1 || Gia_ManLutNum(result) != 1 {
		t.Fatalf("invalid result: CI=%d CO=%d LUT=%d", Gia_ManCiNum(result), Gia_ManCoNum(result), Gia_ManLutNum(result))
	}
	Gia_ManMappingVerify(result)
}
