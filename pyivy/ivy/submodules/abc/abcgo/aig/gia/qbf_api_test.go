package gia

import "testing"

func TestQbfPublicSolverAndEnumeration(t *testing.T) {
	g := Gia_ManStart(8)
	p, x := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManHashAlloc(g)
	Gia_ManAppendCo(g, Gia_ManHashOr(g, p, x))
	if got := Gia_QbfSolve(g, 1, 0, 0, 0, 0, 0, 0); got != 0 {
		t.Fatalf("solver status %d", got)
	}
	if got := Gia_ManSatEnum(g, 0, 0, 0); got != 1 {
		t.Fatalf("enumeration status %d", got)
	}
}
