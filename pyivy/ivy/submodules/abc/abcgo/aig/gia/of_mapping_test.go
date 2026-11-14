package gia

import "testing"

func TestOfForwardBackwardMapping(t *testing.T) {
	g := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	out := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, out)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	pars.NDelayLut1 = 10
	pars.NAreaTuner = 1
	cuts := newOfCutManager(g, &pars)
	cuts.computeCuts()
	mapper := newOfMapper(cuts)
	for i := 0; i < 3; i++ {
		mapper.forward()
		mapper.backward()
	}
	if pars.Area != 1 || pars.Delay != 10 || mapper.objects[out>>1].best == 0 || mapper.objects[x>>1].refs != 0 {
		t.Fatalf("unexpected mapping: area %d delay %d", pars.Area, pars.Delay)
	}
	cut := cuts.storage.at(mapper.objects[out>>1].best)
	if ofStoredSize(cut) != 3 {
		t.Fatal("expected three-input LUT")
	}
}
