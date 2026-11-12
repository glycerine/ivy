package gia

import "testing"

func TestBDDSupportMinimization(t *testing.T) {
	p := Gia_ManStart(16)
	Gia_ManHashAlloc(p)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManHashAnd(p, a, b)
	m := Gia_ManSuppStart(p)
	defer Gia_ManSuppStop(m)
	if got := Gia_ManSupportAnd(m, x, a^1); got != 0 {
		t.Fatalf("contradiction was not reduced: %d", got)
	}
	if Vec_IntSize(m.vCleared) != 2 {
		t.Fatalf("unexpected removed inputs: %v", Vec_IntArray(m.vCleared))
	}
	or := Gia_ManHashAnd(p, a^1, b^1) ^ 1
	if got := Gia_ManSupportAnd(m, or, a); got != a {
		t.Fatalf("absorption failed: got %d, want %d", got, a)
	}
	if got := Gia_ManSupportAnd(m, a, b); got != x {
		t.Fatalf("independent inputs changed: %d", got)
	}
}
