package gia

import "testing"

func TestMin2SupportAnd(t *testing.T) {
	p := Gia_ManStart(16)
	Gia_ManHashAlloc(p)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManHashAnd(p, a, b)
	m := Gia_Man2SuppStart(p)
	defer Gia_Man2SuppStop(m)
	if got := Gia_Man2SupportAnd(m, x, a^1); got != 0 {
		t.Fatalf("contradiction: %d", got)
	}
	if m.nSatUnsat != 1 {
		t.Fatal("SAT did not prove contradiction")
	}
	if got := Gia_Man2SupportAnd(m, a, b); got != x {
		t.Fatalf("hash lookup: %d", got)
	}
	if got := Gia_Man2SupportAnd(m, x, a); got == 0 {
		t.Fatal("lost satisfiable conjunction")
	}
}
