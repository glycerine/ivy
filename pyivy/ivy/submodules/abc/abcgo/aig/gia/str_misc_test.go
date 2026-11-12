package gia

import "testing"

func TestDuplicateMuxesWithoutHashing(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendXor(p, a, b)
	y := Gia_ManAppendMux(p, c, x, d)
	Gia_ManAppendCo(p, y)
	q := Gia_ManDupMuxesNoHash(p)
	if q == nil || Gia_ManVerifyTwoTruths(p, q) != 1 {
		t.Fatal("unhashed mux duplication changed function")
	}
	if Gia_ManHasDangling(q) != 0 {
		t.Fatal("duplication has dangling nodes")
	}
}
