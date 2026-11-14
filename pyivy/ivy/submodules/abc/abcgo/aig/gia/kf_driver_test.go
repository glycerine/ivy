package gia

import "testing"

func TestKfDriver(t *testing.T) {
	for _, mode := range []struct {
		name                string
		procs, hash, simple int32
	}{{"ordered", 0, 0, 0}, {"simple", 0, 0, 1}, {"hash", 0, 1, 0}, {"workers1", 1, 0, 0}, {"workers3", 3, 0, 0}} {
		t.Run(mode.name, func(t *testing.T) {
			g := Gia_ManStart(16)
			a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			x := Gia_ManAppendAnd(g, a, b^1)
			y := Gia_ManAppendAnd(g, c, d)
			z := Gia_ManAppendAnd(g, x, y)
			Gia_ManAppendCo(g, z^1)
			var pars Jf_Par_t
			Kf_ManSetDefaultPars(&pars)
			pars.NLutSize = 4
			pars.NProcNum = mode.procs
			pars.FCutHashing = mode.hash
			pars.FCutSimple = mode.simple
			out := Kf_ManPerformMapping(g, &pars)
			Gia_ManMappingVerify(out)
			for mask := 0; mask < 16; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				got, err := out.Evaluate(inputs)
				want := !(inputs[0] && !inputs[1] && inputs[2] && inputs[3])
				if err != nil || got[0] != want {
					t.Fatalf("truth %d got %v err %v", mask, got, err)
				}
			}
			if pars.Area != 1 || pars.Edge != 4 || pars.Delay != 1 {
				t.Fatalf("area %d edge %d delay %d", pars.Area, pars.Edge, pars.Delay)
			}
		})
	}
}
