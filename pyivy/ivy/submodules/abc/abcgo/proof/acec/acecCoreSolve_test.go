package acec

import (
	"testing"
)

func TestFallbackCecTransfersCounterexample(t *testing.T) {
	pars := &Acec_ParCec_t{}
	Acec_ManCecSetDefaultParams(pars)
	a, b := reviewAdderChain(-1), reviewAdderChain(0)
	if got := Acec_Solve(a, b, pars); got != 0 {
		t.Fatalf("different graphs status %d", got)
	}
	if a.CombinationalCounterexample() == nil {
		t.Fatal("missing transferred witness")
	}
	a, b = reviewAdderChain(-1), reviewAdderChain(-1)
	if got := Acec_Solve(a, b, pars); got != 1 {
		t.Fatalf("equivalent graphs status %d", got)
	}
}
