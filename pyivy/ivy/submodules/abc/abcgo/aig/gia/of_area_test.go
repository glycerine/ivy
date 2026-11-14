package gia

import "testing"

func TestOfExactAreaRecoversSharedCut(t *testing.T) {
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
	mapper.forward()
	mapper.backward()
	// Start with a two-LUT implementation, leaving a one-LUT cut available.
	var two int32
	mapper.eachCut(out>>1, func(handle int32, cut []int32) {
		if ofStoredSize(cut) == 2 {
			two = handle
		}
	})
	if two == 0 {
		t.Fatal("missing two-input cut")
	}
	mapper.objects[out>>1].best = two
	mapper.objects[x>>1].refs = 1
	mapper.forward()
	mapper.backwardExact()
	if pars.Area != 1 || mapper.objects[x>>1].refs != 0 || ofStoredSize(cuts.storage.at(mapper.objects[out>>1].best)) != 3 {
		t.Fatalf("area recovery failed: %d", pars.Area)
	}
}
