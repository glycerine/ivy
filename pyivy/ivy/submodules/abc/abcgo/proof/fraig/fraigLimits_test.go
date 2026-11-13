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
