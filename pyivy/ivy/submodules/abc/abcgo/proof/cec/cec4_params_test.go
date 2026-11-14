package cec

import "testing"

func TestCec4ParametersResetDefaults(t *testing.T) {
	p := Cec_ParFra_t{FVerbose: 1, FUseCones: 1, TimeLimit: 12}
	Cec4_ManSetParams(&p)
	want := Cec_ParFra_t{JType: 2, fSatSweeping: 1, NWords: 4, NRounds: 10, NItersMax: 2000, NBTLimit: 1000000, nSatVarMax: 1000, NCallsRecycle: 500, NGenIters: 100}
	if p != want {
		t.Fatalf("defaults = %+v, want %+v", p, want)
	}
}
