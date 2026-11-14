package bmc

import "testing"

func TestInterpolationOptions(t *testing.T) {
	var params Saig_ParBmc_t
	Saig_ParBmcSetDefaultParams(&params)
	params.SetInterpolationOptions(4, 100000000, 1)
	if params.NStart != 4 || params.NConfLimit != 100000000 || params.FVerbose != 1 {
		t.Fatalf("wrong interpolation options: %+v", params)
	}
	if params.NLearnedStart != 10000 {
		t.Fatal("defaults unexpectedly changed")
	}
}
