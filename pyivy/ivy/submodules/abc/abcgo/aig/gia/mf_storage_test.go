package gia

import "testing"

func TestMfCutStorage(t *testing.T) {
	p := &Mf_Man_t{pLfObjs: make([]Mf_Obj_t, 3), iCur: 2}
	a := &Mf_Cut_t{nLeaves: 2, iFunc: 7}
	a.pLeaves[0], a.pLeaves[1] = 1, 2
	b := &Mf_Cut_t{nLeaves: 1, iFunc: 2}
	b.pLeaves[0] = 3
	id := Mf_ManSaveCuts(p, []*Mf_Cut_t{a, b}, 2)
	p.pLfObjs[2].iCutSet = id
	var cuts [3]Mf_Cut_t
	if n := Mf_ManPrepareCuts(cuts[:], p, 2, 1); n != 3 {
		t.Fatalf("count %d", n)
	}
	if cuts[0].iFunc != 7 || cuts[0].pLeaves[1] != 2 || cuts[2].pLeaves[0] != 2 {
		t.Fatalf("roundtrip %+v", cuts)
	}
	packed := Mf_ObjCutSet(p, 2)
	Mf_ObjSetBestCut(packed, 4)
	if Mf_CutIsTriv(Mf_ObjCutBest(p, 2), 3) != 1 || Mf_CutFunc(packed[3:]) != 7 {
		t.Fatalf("reorder %v", packed[:6])
	}
	p.iCur = 65534
	boundary := Mf_ManSaveCuts(p, []*Mf_Cut_t{a}, 1)
	if boundary != 65536 || len(p.vPages) != 2 {
		t.Fatalf("boundary %d pages %d", boundary, len(p.vPages))
	}
	if Mf_ObjMapRefInc(p, 1) != 0 || Mf_ObjMapRefInc(p, 1) != 1 || Mf_ObjMapRefDec(p, 1) != 1 {
		t.Fatal("reference count")
	}
}
