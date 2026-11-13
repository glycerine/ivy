package fraig

import "testing"

func TestManagerParameterBounds(t *testing.T) {
	for _, tc := range []struct{ random, dynamic, red, wantR, wantD int32 }{{1, 50000, 1, 128, 32768}, {50000, 1, 1, 32768, 128}, {32768, 32768, 0, 128, 128}} {
		params := &Fraig_Params_t{NPatsRand: tc.random, NPatsDyna: tc.dynamic, FFuncRed: tc.red}
		p := Fraig_ManCreate(params)
		if params.NPatsRand != tc.wantR || params.NPatsDyna != tc.wantD || p.nWordsRand != (tc.wantR+31)/32 || p.nWordsDyna != (tc.wantD+31)/32 {
			t.Fatalf("parameter bounds: %+v => %+v", tc, params)
		}
		Fraig_ManFree(p)
	}
}
func TestEquivalenceSkipsFailedNodesUnderSmallLimit(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	a.fFailTfo = 1
	if got := Fraig_NodeIsEquivalent(p, a, b, 10, 1); got != 0 || p.nSatFails != 1 || p.pSat != nil {
		t.Fatalf("failed-node limit: result %d, failures %d", got, p.nSatFails)
	}
}

// Fraig_ManCreate rounds pattern storage to words after clamping and copies
// the caller's settings while leaving solver allocation lazy.
func TestManagerInitialStateAndWordRounding(t *testing.T) {
	params := &Fraig_Params_t{NPatsRand: 129, NPatsDyna: 161, NBTLimit: 73, NSeconds: 8, FFuncRed: 1, FFeedBack: 1, FDist1Pats: 1, FDoSparse: 1, FChoicing: 1, FTryProve: 1, NInspLimit: 123}
	p := Fraig_ManCreate(params)
	defer Fraig_ManFree(p)
	if p.nWordsRand != 5 || p.nWordsDyna != 6 || p.nTravIds != 1 || p.nTravIds2 != 1 {
		t.Fatalf("initial dimensions: random=%d dynamic=%d trav=%d/%d", p.nWordsRand, p.nWordsDyna, p.nTravIds, p.nTravIds2)
	}
	if p.nBTLimit != 73 || p.nSeconds != 8 || p.nInspLimit != 123 || p.fFuncRed != 1 || p.fFeedBack != 1 || p.fDist1Pats != 1 || p.fDoSparse != 1 || p.fChoicing != 1 || p.fTryProve != 1 {
		t.Fatal("manager settings not copied")
	}
	if p.pSat != nil || Fraig_ManReadInputNum(p) != 0 || Fraig_ManReadOutputNum(p) != 0 || Fraig_ManReadNodeNum(p) != 1 || p.vNodes.pArray[0] != p.pConst1 {
		t.Fatal("unexpected initial manager state")
	}
	Fraig_ManCreateSolver(p)
	if p.pSat == nil || p.vVarsInt == nil || p.vVarsUsed == nil {
		t.Fatal("solver state was not initialized")
	}
}

// fraigSat.c skips a failed TFO at limits through 10, but resumes SAT above
// that threshold; a nonpositive limit bypasses the failed-TFO adjustment.
func TestFailedTFOThresholdAndUnlimitedSAT(t *testing.T) {
	for _, tc := range []struct {
		limit        int32
		calls, fails int32
	}{{10, 0, 1}, {11, 1, 1}, {0, 1, 0}, {-1, 1, 0}} {
		p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
		a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
		b.fFailTfo = 1
		result := Fraig_NodeIsEquivalent(p, a, b, tc.limit, 1)
		if result != 0 || p.nSatCalls != tc.calls || p.nSatFails != tc.fails || (p.pSat != nil) != (tc.calls != 0) {
			t.Fatalf("limit=%d result=%d calls=%d fails=%d solver=%v", tc.limit, result, p.nSatCalls, p.nSatFails, p.pSat != nil)
		}
		Fraig_ManFree(p)
	}
}
