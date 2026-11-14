package gia

import "testing"

func TestJfMappingRounds(t *testing.T) {
	for _, mode := range []struct {
		name     string
		min, dsd int32
	}{{"structural", 0, 0}, {"truth", 1, 0}, {"dsd", 1, 1}} {
		t.Run(mode.name, func(t *testing.T) {
			g := Gia_ManStart(12)
			a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			x := Gia_ManAppendAnd(g, a, b)
			y := Gia_ManAppendAnd(g, x, c)
			Gia_ManAppendCo(g, y)
			pars := &Jf_Par_t{NLutSize: 3, NCutNum: 4, FCutMin: mode.min, FFuncDsd: mode.dsd}
			p := Jf_ManAlloc(g, pars)
			defer Jf_ManFree(p)
			p.pCutCmp = Jf_CutCompareDelay
			Jf_ManComputeCuts(p, 0)
			Jf_ManComputeRefs(p)
			if pars.Area != 1 || pars.Delay != 1 || pars.Edge != 3 {
				t.Fatalf("initial area %d delay %d edge %d", pars.Area, pars.Delay, pars.Edge)
			}
			Jf_ManPropagateFlow(p, 0)
			Jf_ManPropagateEla(p, 0)
			if pars.Area != 1 || pars.Delay != 1 || pars.Edge != 3 {
				t.Fatalf("final area %d delay %d edge %d", pars.Area, pars.Delay, pars.Edge)
			}
		})
	}
}
