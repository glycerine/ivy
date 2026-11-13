package cec

import "testing"

func TestCec4ParametersResetDefaults(t *testing.T) {
	p := Cec_ParFra_t{fVerbose: 1, fUseCones: 1, TimeLimit: 12}
	Cec4_ManSetParams(&p)
	want := Cec_ParFra_t{jType: 2, fSatSweeping: 1, nWords: 4, nRounds: 10, nItersMax: 2000, nBTLimit: 1000000, nSatVarMax: 1000, nCallsRecycle: 500, nGenIters: 100}
	if p != want {
		t.Fatalf("defaults = %+v, want %+v", p, want)
	}
}
