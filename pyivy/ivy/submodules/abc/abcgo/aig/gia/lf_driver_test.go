package gia

import "testing"

func TestLfDriver(t *testing.T) {
	for _, mode := range []struct {
		name        string
		min, coarse int32
	}{{"structural", 0, 0}, {"truth", 1, 0}, {"coarse", 0, 1}, {"coarse-truth", 1, 1}} {
		t.Run(mode.name, func(t *testing.T) {
			g := Gia_ManStart(12)
			a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			x := Gia_ManAppendAnd(g, a, b^1)
			y := Gia_ManAppendAnd(g, x, c)
			Gia_ManAppendCo(g, y^1)
			var pars Jf_Par_t
			Lf_ManSetDefaultPars(&pars)
			pars.NLutSize = 3
			pars.NCutNum = 4
			pars.FCutMin = mode.min
			pars.FCoarsen = mode.coarse
			out := Lf_ManPerformMapping(g, &pars)
			Gia_ManMappingVerify(out)
			for mask := 0; mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				got, err := out.Evaluate(inputs)
				want := !(inputs[0] && !inputs[1] && inputs[2])
				if err != nil || got[0] != want {
					t.Fatalf("mask %d out %v err %v", mask, got, err)
				}
			}
		})
	}
}

func TestLfCrossCut(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, a)
	Gia_ManAppendCo(g, y)
	if got := Lf_ManComputeCrossCut(g); got != 1 {
		t.Fatalf("cross cut %d", got)
	}
}
