package gia

import "testing"

func TestOfEnumerateCuts(t *testing.T) {
	for _, truth := range []int32{0, 1} {
		g := Gia_ManStart(10)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a, b)
		out := Gia_ManAppendAnd(g, x, c)
		Gia_ManAppendCo(g, out)
		var pars Jf_Par_t
		Mf_ManSetDefaultPars(&pars)
		pars.FCutMin = truth
		p := newOfCutManager(g, &pars)
		p.computeCuts()
		data := p.storage.at(p.sets[out>>1])
		if data[0] == 0 || p.delays[out>>1] == 0 {
			t.Fatal("missing cuts")
		}
		pos := 1
		found := false
		for i := int32(0); i < data[0]; i++ {
			cut := data[pos:]
			n := ofStoredSize(cut)
			if n == 2 && ofStoredVar(cut, 0) == c>>1 && ofStoredVar(cut, 1) == x>>1 {
				found = true
			}
			pos += int(n) + 4
		}
		if !found {
			t.Fatalf("truth=%d: missing cut", truth)
		}
	}
}
