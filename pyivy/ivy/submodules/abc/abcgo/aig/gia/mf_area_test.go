package gia

import "testing"

func TestMfCutReferenceArea(t *testing.T) {
	p := &Mf_Man_t{pPars: &Jf_Par_t{NLutSize: 4}, pLfObjs: make([]Mf_Obj_t, 5), iCur: 2}
	child := &Mf_Cut_t{nLeaves: 2, iFunc: 2}
	child.pLeaves[0], child.pLeaves[1] = 1, 2
	p.pLfObjs[3].iCutSet = Mf_ManSaveCuts(p, []*Mf_Cut_t{child}, 1)
	root := []int32{Mf_CutSetBoth(2, 2), 3, 4}
	if got := Mf_CutAreaDerefed(p, root); got != 2 {
		t.Fatalf("area %d", got)
	}
	if Mf_ObjMapRefNum(p, 3) != 0 || Mf_ObjMapRefNum(p, 1) != 0 {
		t.Fatal("references not restored")
	}
	if Mf_CutRef_rec(p, root) != 2 || Mf_CutAreaRefed2(p, root) != 2 {
		t.Fatal("reference area")
	}
	if Mf_ObjMapRefNum(p, 3) != 1 || Mf_ObjMapRefNum(p, 1) != 1 {
		t.Fatal("limited references not restored")
	}
	if Mf_CutDeref_rec(p, root) != 2 {
		t.Fatal("dereference area")
	}
	p.pLfObjs[1].Flow = 2
	p.pLfObjs[1].Delay = 3
	var delay int32
	if flow := Mf_CutFlow(p, []int32{Mf_CutSetBoth(2, 0), 1, 2}, &delay); flow != 3 || delay != 4 {
		t.Fatalf("flow %v delay %d", flow, delay)
	}
}
