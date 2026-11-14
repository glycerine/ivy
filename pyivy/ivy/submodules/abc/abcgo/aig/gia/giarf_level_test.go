package gia

import "testing"

func TestGiarfPatternPacking(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	p := Gia_ManEquivStart(g, 10, 0)
	if Gia_GiarfStorePatternTry(p.vSimInfo, p.vSimPres, 1, []int32{0, 3}, 2) != 1 {
		t.Fatal("initial pattern rejected")
	}
	if Hcd_ObjSim(p, a>>1)&2 == 0 || Hcd_ObjSim(p, b>>1)&2 != 0 {
		t.Fatal("wrong stored bits")
	}
	if Gia_GiarfStorePatternTry(p.vSimInfo, p.vSimPres, 1, []int32{1}, 1) != 0 {
		t.Fatal("conflicting pattern accepted")
	}
	if Hcd_ObjSim(p, a>>1)&2 == 0 {
		t.Fatal("conflict modified pattern")
	}
}

func TestGiarfReducedLevel(t *testing.T) {
	g := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManLevelNum(g)
	p := Gia_ManEquivStart(g, 10, 0)
	Hcd_ManClassesCreate(p)
	var roots *Vec_Ptr_t
	q := Gia_GenerateReducedLevel(g, 2, &roots)
	if Vec_PtrSize(roots) != 1 || Vec_PtrEntry(roots, 0) != Gia_ManObj(g, y>>1) || Gia_ManAndNum(q) != 2 {
		t.Fatal("reduced graph or roots are incorrect")
	}
}

func TestGiarfTemporaryRepresentativeRequiresProof(t *testing.T) {
	g := Gia_ManStart(1)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	ab := Gia_ManAppendAnd(g, a, b)
	head := Gia_ManAppendAnd(g, ab, c)
	lower, current := Gia_ManAppendAnd(g, a, b), Gia_ManAppendAnd(g, a, b)
	Gia_ManLevelNum(g)
	g.StartRepresentatives()
	Gia_ObjSetRepr(g, lower>>1, head>>1)
	Gia_ObjSetRepr(g, current>>1, head>>1)
	Gia_ObjSetNext(g, head>>1, lower>>1)
	Gia_ObjSetNext(g, lower>>1, current>>1)
	Gia_ObjSetProved(g, current>>1)
	Gia_ObjSetFailed(g, head>>1)
	if got := Gia_ObjTempRepr(g, current>>1, 3); got != nil {
		t.Fatal("selected a temporary representative without a proof")
	}
	Gia_ObjSetProved(g, lower>>1)
	if got := Gia_ObjTempRepr(g, current>>1, 3); got != Gia_ManObj(g, lower>>1) {
		t.Fatal("lost a proved lower-level representative")
	}
	Gia_ObjSetFailed(g, lower>>1)
	if got := Gia_ObjTempRepr(g, current>>1, 3); got != nil {
		t.Fatal("selected a representative whose SAT check failed")
	}
}
