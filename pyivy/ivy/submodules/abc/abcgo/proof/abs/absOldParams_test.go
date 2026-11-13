package abs

import "testing"

func TestLegacyAbstractionDefaultsResetPreviousValues(t *testing.T) {
	p := Gia_ParAbs_t{Algo: 3, fConstr: 1, nFramesMax: 91, fVerbose: 1, Status: 7}
	Gia_ManAbsSetDefaultParams(&p)
	if p.Algo != 0 || p.fConstr != 0 || p.fVerbose != 0 || p.nFramesMax != 10 || p.nConfMax != 10000 ||
		p.fDynamic != 1 || p.nFramesBmc != 250 || p.nConfMaxBmc != 5000 || p.nStableMax != 1000000 ||
		p.nRatio != 10 || p.nBobPar != 1000000 || p.fUseStart != 1 || p.Status != -1 || p.nFramesDone != -1 {
		t.Fatalf("unexpected default parameters: %+v", p)
	}
}
