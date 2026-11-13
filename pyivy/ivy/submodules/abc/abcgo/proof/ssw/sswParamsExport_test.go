package ssw

import "testing"

func TestSetInductionDepth(t *testing.T) {
	p := new(Ssw_Pars_t)
	Ssw_ParsSetFramesK(p, 4)
	if p.nFramesK != 4 {
		t.Fatal("depth not set")
	}
}

func TestConflictLimitsAndStatistics(t *testing.T) {
	p := new(Ssw_Pars_t)
	Ssw_ManSetDefaultParamsLcorr(p)
	Ssw_ParsSetBTLimit(p, 7)
	Ssw_ParsSetBTLimitGlobal(p, 42)
	p.nIters, p.nConflicts = 3, 11
	if p.nBTLimit != 7 || p.nBTLimitGlobal != 42 || Ssw_ParsIterations(p) != 3 || Ssw_ParsConflicts(p) != 11 {
		t.Fatal("parameter accessor mismatch")
	}
}
