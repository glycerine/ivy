package gia

import "testing"

func TestGiaMfsExtractSimple(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	pars.fCoarsen = 0
	mapped := Mf_ManPerformMapping(g, &pars)
	d := giaMfsExtract(mapped)
	if d.pis != 2 || d.pos != 1 || len(d.fanins) != 5 || len(d.fanins[3]) != 2 || d.truths[3][0]&15 != 8 || len(d.fanins[4]) != 1 {
		t.Fatalf("extracted %+v", d)
	}
}
