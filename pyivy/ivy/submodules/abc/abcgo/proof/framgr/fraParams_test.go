package framgr

import "testing"

func TestDefaultParametersResetAndConfigure(t *testing.T) {
	for _, tc := range []struct {
		name                       string
		set                        func(*Fra_Par_t)
		words, frames, limit, cone int32
	}{
		{"comb", Fra_ParamsDefault, 32, 0, 100, 1},
		{"seq", Fra_ParamsDefaultSeq, 1, 1, 10000000, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := Fra_Par_t{FVerbose: 1, NMaxLevs: 99, FLatchCorr: 1}
			tc.set(&p)
			if p.NSimWords != tc.words || p.NFramesK != tc.frames || p.NBTLimitNode != tc.limit || p.FConeBias != tc.cone {
				t.Fatalf("unexpected defaults: %+v", p)
			}
			if p.FVerbose != 0 || p.NMaxLevs != 0 || p.FLatchCorr != 0 {
				t.Fatalf("stale parameters: %+v", p)
			}
			if p.DSimSatur != .005 || p.DActConeRatio != .3 || p.DActConeBumpMax != 10 || p.NBTLimitMiter != 500000 || p.MaxScore != 25 || p.FDoSparse != 1 {
				t.Fatalf("unexpected common defaults: %+v", p)
			}
		})
	}
}
