package gia

import "testing"

func TestGiarfSimulationAndRefinement(t *testing.T) {
	g := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, a^1, b)
	z := Gia_ManAppendAnd(g, a, b^1)
	p := Gia_ManEquivStart(g, 100, 0)
	Hcd_ObjSetSim(p, a>>1, 0xaaaaaaaa)
	Hcd_ObjSetSim(p, b>>1, 0xcccccccc)
	Hcd_ManSimulateSimple(p)
	if Hcd_ObjSim(p, x>>1) != 0x88888888 || Hcd_ObjSim(p, y>>1) != 0x44444444 || Hcd_ObjSim(p, z>>1) != 0x22222222 {
		t.Fatal("wrong simulation values")
	}
	Hcd_ManClassesCreate(p)
	class := Vec_IntAlloc(3)
	for _, id := range []int32{x >> 1, y >> 1, z >> 1} {
		Vec_IntPush(class, id)
	}
	Hcd_ManClassCreate(g, class)
	if Hcd_ManClassRefineOne(p, x>>1) != 2 || Gia_ObjIsClass(g, x>>1) != 0 || Gia_ObjIsClass(g, y>>1) != 0 || Gia_ObjIsClass(g, z>>1) != 0 {
		t.Fatal("classes not refined")
	}
	if Hcd_ManCompareEqual(0xaaaa, ^uint32(0xaaaa)) != 1 || Hcd_ManCompareConst(^uint32(0)) != 1 {
		t.Fatal("complement comparisons")
	}
}
