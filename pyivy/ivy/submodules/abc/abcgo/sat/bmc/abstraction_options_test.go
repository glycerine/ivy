package bmc

import "testing"

func TestAbstractionVerboseOption(t *testing.T) {
	var p Saig_ParBmc_t
	p.SetAbstractionVerbose(1)
	if p.AbstractionVerbose() != 1 {
		t.Fatal("verbose option was not retained")
	}
	p.SetAbstractionVerbose(0)
	if p.AbstractionVerbose() != 0 {
		t.Fatal("verbose option was not cleared")
	}
}
