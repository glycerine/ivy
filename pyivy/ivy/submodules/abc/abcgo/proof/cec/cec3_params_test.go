package cec

import "testing"

func TestCec3DefaultsResetPreviousSettings(t *testing.T) {
	params := Cec3_Par_t{nSimWords: 1, fIsMiter: 1, fUseCones: 1, fVeryVerbose: 1, fVerbose: 1}
	Cec3_SetDefaultParams(&params)
	if want := (Cec3_Par_t{nSimWords: 12, nSimRounds: 4, nItersMax: 10, nConfLimit: 1000}); params != want {
		t.Fatalf("defaults = %+v, want %+v", params, want)
	}
}
