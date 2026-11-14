package amap

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func mergeFixture(t *testing.T, maxCuts int32) (*Amap_Man_t, *Amap_Obj_t) {
	t.Helper()
	lib, _ := testAndLibrary(t)
	lib.pRules = Amap_LibLookupTableAlloc(lib.vRules, 0)
	p := &Amap_Man_t_{pLib: lib, pPars: &Amap_Par_t{NCutsMax: maxCuts}, vTemp: Vec_IntAlloc(4), ppCutsTemp: make([]*Amap_Cut_t, 8)}
	p.vObjs = Vec_PtrAlloc(4)
	a := &Amap_Obj_t_{Type: AMAP_OBJ_PI, Id: 0, nCuts: 1, pData: &Amap_Cut_t{nFans: 1, Fans: []int32{0}}}
	b := &Amap_Obj_t_{Type: AMAP_OBJ_PI, Id: 1, nCuts: 1, pData: &Amap_Cut_t{nFans: 1, Fans: []int32{2}}}
	root := &Amap_Obj_t_{Type: AMAP_OBJ_AND, Id: 2, Fan: [3]int32{0, 2}}
	for _, obj := range []*Amap_Obj_t_{a, b, root} {
		vec.Vec_PtrPush(p.vObjs, obj)
	}
	return p, root
}

func TestMergeBuildsTrivialAndImplementableCuts(t *testing.T) {
	p, root := mergeFixture(t, 8)
	Amap_ManMergeNodeCuts(p, root)
	if root.nCuts != 2 {
		t.Fatalf("cut count = %d, want trivial plus one implementation", root.nCuts)
	}
	trivial := Amap_CutHead(root)
	if trivial == nil || trivial.iMat != 0 || len(trivial.Fans) != 1 || trivial.Fans[0] != 4 {
		t.Fatalf("trivial cut = %#v", trivial)
	}
	implementation := trivial.pNext
	if implementation == nil || implementation.iMat != 1 || implementation.nFans != 2 {
		t.Fatalf("AND cut = %#v", implementation)
	}
	if implementation.Fans[0] != 0 || implementation.Fans[1] != 2 {
		t.Fatalf("AND cut fanins = %v", implementation.Fans)
	}
}

func TestMergeRespectsCutLimitAndOwnsFanStorage(t *testing.T) {
	p, root := mergeFixture(t, 3)
	single := &Amap_Cut_t{nFans: 1, Fans: []int32{0}}
	Amap_ManCutCreate(p, single, single, 1)
	Amap_ManCutCreate(p, single, single, 2)
	Amap_ManCutCreate(p, single, single, 3)
	Amap_ManCutSaveStored(p, root)
	if root.nCuts != 2 {
		t.Fatalf("source cut limit should retain at most nCutsMax-1 cuts: got %d", root.nCuts)
	}
	kept := Amap_CutHead(root).pNext
	if kept == nil || kept.iMat != 1 {
		t.Fatalf("lowest ordered implementation cut = %#v", kept)
	}
	single.Fans[0] = 99
	if kept.Fans[0] != 0 {
		t.Fatalf("stored cut aliases caller fan buffer: %v", kept.Fans)
	}
}
