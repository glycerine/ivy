package gia

import "testing"

func TestMfSelectAndDeriveMapping(t *testing.T) {
	g := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, y)
	p := &Mf_Man_t{pGia: g, pGia0: g, pPars: &Jf_Par_t{NLutSize: 3}, pLfObjs: make([]Mf_Obj_t, g.nObjs), iCur: 2}
	xID, yID := x>>1, y>>1
	first := &Mf_Cut_t{nLeaves: 2, iFunc: 2}
	first.pLeaves[0], first.pLeaves[1] = xID, c>>1
	second := &Mf_Cut_t{nLeaves: 3, iFunc: 2}
	second.pLeaves[0], second.pLeaves[1], second.pLeaves[2] = a>>1, b>>1, c>>1
	p.pLfObjs[yID].iCutSet = Mf_ManSaveCuts(p, []*Mf_Cut_t{first, second}, 2)
	p.pLfObjs[yID].nFlowRefs = 1
	p.pLfObjs[xID].Flow = 4
	Mf_ObjComputeBestCut(p, yID)
	if got := Mf_CutSize(Mf_ObjCutBest(p, yID)); got != 3 {
		t.Fatalf("best cut size %d", got)
	}
	p.pLfObjs[yID].nMapRefs = 1
	if Mf_ManDeriveMapping(p) != g || Gia_ObjIsLut(g, yID) == 0 || Gia_ObjLutSize(g, yID) != 3 {
		t.Fatal("mapping not derived")
	}
}
