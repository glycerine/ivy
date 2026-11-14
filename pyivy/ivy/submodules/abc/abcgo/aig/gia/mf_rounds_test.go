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
	pars.FCoarsen = 0
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

func TestMfPerformMappingModes(t *testing.T) {
	for _, tc := range []struct{ coarse, truth, cnf int32 }{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}, {1, 1, 1}} {
		g := Gia_ManStart(12)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a, b)
		y := Gia_ManAppendAnd(g, x, c)
		Gia_ManAppendCo(g, y)
		var pars Jf_Par_t
		Mf_ManSetDefaultPars(&pars)
		pars.FCoarsen, pars.FCutMin, pars.FGenCnf = tc.coarse, tc.truth, tc.cnf
		mapped := Mf_ManPerformMapping(g, &pars)
		if Gia_ManLutNum(mapped) == 0 || Gia_ManPoNum(mapped) != 1 {
			t.Fatalf("mode %+v invalid mapping", tc)
		}
		if tc.cnf != 0 && g.pData == nil {
			t.Fatalf("mode %+v missing CNF", tc)
		}
	}
}
