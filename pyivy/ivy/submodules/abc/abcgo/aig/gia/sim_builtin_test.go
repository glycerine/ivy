package gia

import "testing"

func TestBuiltInSimulation(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManBuiltInSimStart(p, 1, 8)
	Gia_ManBuiltInData(p, a>>1)[0] = 0xaaaaaaaaaaaaaaaa
	Gia_ManBuiltInData(p, b>>1)[0] = 0xcccccccccccccccc
	x := Gia_ManAppendAnd(p, a, b^1)
	if got := Gia_ManBuiltInData(p, x>>1)[0]; got != 0x2222222222222222 {
		t.Fatalf("simulation got %x", got)
	}
	if Gia_ManBuiltInSimCheckOver(p, x, b) != 0 || Gia_ManBuiltInSimCheckEqual(p, x, x) != 1 || Gia_ManBuiltInSimCheckEqual(p, x, x^1) != 0 {
		t.Fatal("incorrect overlap or equivalence")
	}
	Gia_ManBuiltInData(p, a>>1)[0] = 0
	Gia_ManBuiltInSimResimulateCone(p, x, a)
	if Gia_ManBuiltInData(p, x>>1)[0] != 0 {
		t.Fatal("cone was not resimulated")
	}
}
