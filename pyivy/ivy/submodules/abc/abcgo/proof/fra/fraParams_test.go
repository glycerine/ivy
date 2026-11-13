package fra

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
			p := Fra_Par_t{fVerbose: 1, nMaxLevs: 99, fLatchCorr: 1}
			tc.set(&p)
			if p.nSimWords != tc.words || p.nFramesK != tc.frames || p.nBTLimitNode != tc.limit || p.fConeBias != tc.cone {
				t.Fatalf("unexpected defaults: %+v", p)
			}
			if p.fVerbose != 0 || p.nMaxLevs != 0 || p.fLatchCorr != 0 {
				t.Fatalf("stale parameters: %+v", p)
			}
			if p.dSimSatur != .005 || p.dActConeRatio != .3 || p.dActConeBumpMax != 10 || p.nBTLimitMiter != 500000 || p.MaxScore != 25 || p.fDoSparse != 1 {
				t.Fatalf("unexpected common defaults: %+v", p)
			}
		})
	}
}
