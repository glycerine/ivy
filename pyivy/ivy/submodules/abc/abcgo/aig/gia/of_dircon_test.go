package gia

import "testing"

func TestOfDirectConnectionTiming(t *testing.T) {
	g := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	out := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, out)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	pars.NLutSize = 2
	pars.NDelayLut1 = 10
	pars.NDelayLut2 = 2
	pars.NFastEdges = 1
	cuts := newOfCutManager(g, &pars)
	cuts.computeCuts()
	mapper := newOfMapper(cuts)
	mapper.forwardDirect()
	mapper.backwardDirect()
	if pars.Delay != 12 || pars.Area != 2 {
		t.Fatalf("delay %d area %d", pars.Delay, pars.Area)
	}
	cut := cuts.storage.at(mapper.objects[out>>1].best)
	found := false
	for i := range ofStoredLeaves(cut) {
		if ofStoredVar(cut, i) == x>>1 && ofStoredFlag(cut, i) == 1 {
			found = true
		}
	}
	if !found {
		t.Fatal("missing direct connection")
	}
}
