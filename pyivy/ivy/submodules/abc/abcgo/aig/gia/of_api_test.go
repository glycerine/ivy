package gia

import "testing"

func TestOfPublicMapper(t *testing.T) {
	g := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	out := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, out)
	var pars Jf_Par_t
	Of_ManSetDefaultPars(&pars)
	if Of_ManPerformMapping(g, &pars) != g || Gia_ObjIsLut(g, out>>1) == 0 {
		t.Fatal("public Of mapping failed")
	}
}
