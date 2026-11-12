package gia

import "testing"

func TestResubDiagnosticSimulation(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	sims, n := resubSimulate(g)
	if n != 1 || sims[x>>1][0] != (word(0xaaaaaaaaaaaaaaaa)&word(0xcccccccccccccccc)) {
		t.Fatal("incorrect simulation")
	}
	Gia_ManTryResub(g)
	if g.vFanout != nil {
		t.Fatal("fanout scratch not released")
	}
}
