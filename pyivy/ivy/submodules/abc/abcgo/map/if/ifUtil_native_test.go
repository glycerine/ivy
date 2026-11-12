package if_

import "testing"

func TestIfManagerCrossCutAndReverseOrder(t *testing.T) {
	p := If_ManStart(testManagerPars())
	defer If_ManStop(p)
	ci0, ci1, ci2 := If_ManCreateCi(p), If_ManCreateCi(p), If_ManCreateCi(p)
	first := If_ManCreateAnd(p, ci0, ci1)
	second := If_ManCreateAnd(p, first, ci2)
	third := If_ManCreateAnd(p, ci0, ci2)
	if got := If_ManCrossCut(p); got != 2 {
		t.Fatalf("cross-cut size = %d, want 2", got)
	}
	for _, obj := range []*If_Obj_t{first, second, third} {
		if obj.nVisits != obj.nVisitsCopy {
			t.Fatalf("cross-cut changed visits for object %d: %d != %d", obj.Id, obj.nVisits, obj.nVisitsCopy)
		}
	}

	got := Vec_PtrArray(If_ManReverseOrder(p))
	want := []*If_Obj_t{second, third, first}
	if int32(len(got)) != If_ManObjNum(p) || len(got) < len(want) {
		t.Fatalf("reverse order length = %d, want %d objects", len(got), If_ManObjNum(p))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("reverse order[%d] = object %d, want object %d", i, got[i].(*If_Obj_t).Id, want[i].Id)
		}
	}
}

func TestIfManResetOriginalRefsAndCollectMapping(t *testing.T) {
	p := If_ManStart(testManagerPars())
	defer If_ManStop(p)
	ci0, ci1 := If_ManCreateCi(p), If_ManCreateCi(p)
	and := If_ManCreateAnd(p, ci0, ci1)
	If_ManCreateCo(p, and)
	If_ManResetOriginalRefs(p)
	if and.nRefs != 1 || ci0.nRefs != 1 || ci1.nRefs != 1 {
		t.Fatalf("reference recomputation: and=%d ci0=%d ci1=%d", and.nRefs, ci0.nRefs, ci1.nRefs)
	}
	If_CutSetup(p, &and.CutBest)
	and.CutBest.nLeaves = 2
	and.CutBest.pLeaves = []int32{ci0.Id, ci1.Id}
	if got := Vec_PtrArray(If_ManCollectMappingDirect(p)); len(got) != 1 || got[0] != and {
		t.Fatalf("direct mapped nodes = %v, want only AND node", got)
	}
	encoded := Vec_IntArray(If_ManCollectMappingInt(p))
	if ci0.nRefs != 1 || ci1.nRefs != 1 {
		t.Fatalf("mapping reference counts: ci0=%d ci1=%d, want one each", ci0.nRefs, ci1.nRefs)
	}
	want := []int32{2, ci0.Id, ci1.Id, and.Id}
	if len(encoded) != len(want) {
		t.Fatalf("encoded mapping = %v, want %v", encoded, want)
	}
	for i := range want {
		if encoded[i] != want[i] {
			t.Errorf("encoded mapping[%d] = %d, want %d", i, encoded[i], want[i])
		}
	}
}
