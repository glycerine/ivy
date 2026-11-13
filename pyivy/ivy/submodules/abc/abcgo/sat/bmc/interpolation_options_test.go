package bmc

import "testing"

func TestInterpolationOptions(t *testing.T) {
	var params Saig_ParBmc_t
	Saig_ParBmcSetDefaultParams(&params)
	params.SetInterpolationOptions(4, 100000000, 1)
	if params.nStart != 4 || params.nConfLimit != 100000000 || params.fVerbose != 1 {
		t.Fatalf("wrong interpolation options: %+v", params)
	}
	if params.nLearnedStart != 10000 {
		t.Fatal("defaults unexpectedly changed")
	}
}
