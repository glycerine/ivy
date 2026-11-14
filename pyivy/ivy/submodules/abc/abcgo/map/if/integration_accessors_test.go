package if_

import "testing"

func TestSetCutCostFunction(t *testing.T) {
	pars := &If_Par_t_{}
	called := false
	pars.SetCutCostFunction(func(_ *If_Man_t, _ *If_Cut_t) int32 {
		called = true
		return 17
	})
	if pars.pFuncCost == nil {
		t.Fatal("cut cost callback was not installed")
	}
	if got := pars.pFuncCost(nil, nil); got != 17 || !called {
		t.Fatalf("cut cost callback result=%d called=%t", got, called)
	}
}

func TestSetRenodeConfigPreservesCDefaultsAndCallback(t *testing.T) {
	pars := &If_Par_t_{}
	reo := new(int)
	cost := func(_ *If_Man_t, _ *If_Cut_t) int32 { return 9 }
	pars.SetRenodeConfig(6, 128, 2, 3, 1, 1, 0, 0, 0, 1, reo, cost)
	if pars.nLutSize != 6 || pars.nCutsMax != 128 || pars.nFlowIters != 2 || pars.nAreaIters != 3 || pars.DelayTarget != -1 || pars.Epsilon != .005 || pars.fPreprocess != 1 || pars.fArea != 1 || pars.fTruth != 1 || pars.fUsePerm != 1 || pars.fUseBdds != 1 || pars.fVerbose != 1 {
		t.Fatal("renode configuration did not preserve source defaults/options")
	}
	if pars.pReoMan != reo || pars.pFuncCost == nil || pars.pFuncCost(nil, nil) != 9 {
		t.Fatal("renode manager or cost callback was not installed")
	}
}
