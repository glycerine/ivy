package gia

import "testing"

func TestDeriveNextsExcludesConstantCandidates(t *testing.T) {
	g := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, a, c)
	z := Gia_ManAppendAnd(g, b, c)
	g.StartRepresentatives()
	Gia_ObjSetRepr(g, x>>1, 0)
	Gia_ObjSetRepr(g, y>>1, 0)
	Gia_ObjSetRepr(g, z>>1, a>>1)
	next := Gia_ManDeriveNexts(g)
	if next[0] != 0 || next[x>>1] != 0 || next[y>>1] != 0 {
		t.Fatalf("constant candidates must not form a class: %v", next)
	}
	if next[a>>1] != z>>1 || next[z>>1] != 0 {
		t.Fatalf("nonconstant class was not linked: %v", next)
	}
}
