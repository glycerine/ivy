package gia

import "testing"

func TestLfManagerMappingRounds(t *testing.T) {
	for _, min := range []int32{0, 1} {
		g := Gia_ManStart(12)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a, b)
		y := Gia_ManAppendAnd(g, x, c)
		Gia_ManAppendCo(g, y)
		var pars Jf_Par_t
		Lf_ManSetDefaultPars(&pars)
		pars.NLutSize = 3
		pars.NCutNum = 4
		pars.FCutMin = min
		pars.FCoarsen = 0
		p := Lf_ManAlloc(g, &pars)
		for i := int32(0); i < 2; i++ {
			p.Iter = i
			Lf_ManComputeMapping(p)
			if pars.Area != 1 || pars.Edge != 3 || pars.Delay != 1 {
				t.Fatalf("min=%d iter=%d area=%d edge=%d delay=%d", min, i, pars.Area, pars.Edge, pars.Delay)
			}
		}
		p.fUseEla = 1
		p.Iter = 2
		Lf_ManComputeMapping(p)
		if pars.Area != 1 || pars.Delay != 1 {
			t.Fatalf("exact area: %+v", pars)
		}
		Lf_ManFree(p)
	}
}
