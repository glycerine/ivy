package if_

import "testing"

func TestIfifAccessorsPreserveExperimentParameters(t *testing.T) {
	p := &Ifif_Par_t{}
	lib := &If_LibLut_t_{}
	lib.LutMax = 4
	lib.fVarPinDelays = 1
	lib.pLutDelays[4][2] = 1.75
	IfifSetLibrary(p, lib)
	IfifSetParams(p, IfifParams{LUTSize: 4, Degree: 2, Cascade: 1, Verbose: 1, VeryVerbose: 1, WireDelay: .25})
	got := IfifReadParams(p)
	if got != (IfifParams{LUTSize: 4, Degree: 2, Cascade: 1, Verbose: 1, VeryVerbose: 1, WireDelay: .25}) {
		t.Fatalf("params = %+v", got)
	}
	IfifSetPinDelay(p, 2, .5)
	if IfifReadLibrary(p) != lib || IfifPinDelay(p, 2) != .5 || !If_LibLutHasVariablePinDelays(lib) || If_LibLutMaxSize(lib) != 4 || If_LibLutDelay(lib, 4, 2) != 1.75 {
		t.Fatal("IFIF library/delay accessors lost a field")
	}
}
