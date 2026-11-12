package gia

import "testing"

func TestMfMappingRounds(t *testing.T) {
	g := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, y)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	pars.fCoarsen = 0
	p := Mf_ManAlloc(g, &pars)
	Mf_ManComputeCuts(p)
	p.Iter = 1
	Mf_ManComputeMapping(p)
	p.Iter = 2
	p.fUseEla = 1
	Mf_ManComputeMapping(p)
	if pars.Area == 0 {
		t.Fatal("empty mapping after rounds")
	}
	mapped := Mf_ManDeriveMapping(p)
	Gia_ManMappingVerify(mapped)
}
