package if_

import (
	"slices"
	"testing"
)

func TestImproveNodeUpdateResizesCutLeaves(t *testing.T) {
	p := If_ManStart(&If_Par_t{NLutSize: 4, NCutsMax: 4})
	defer If_ManStop(p)
	a, b, c := If_ManCreateCi(p), If_ManCreateCi(p), If_ManCreateCi(p)
	left := If_ManCreateAnd(p, a, b)
	root := If_ManCreateAnd(p, left, c)
	left.CutBest = If_Cut_t{nLimit: 4, nLeaves: 2, pLeaves: []int32{a.Id, b.Id}}
	root.CutBest = If_Cut_t{nLimit: 4, nLeaves: 2, pLeaves: []int32{c.Id, left.Id}}
	If_ManCreateCo(p, root)
	If_ManCreateCo(p, If_Not(root))
	If_ManMarkMapping(p)
	front := Vec_PtrAlloc(4)
	defer Vec_PtrFree(front)
	for pass := 0; pass < 6; pass++ {
		Vec_PtrClear(front)
		want := []int32{a.Id, b.Id, c.Id}
		area, leftRefs := float32(1), int32(0)
		if pass&1 == 0 {
			Vec_PtrPush(front, c)
			Vec_PtrPush(front, b)
			Vec_PtrPush(front, a)
		} else {
			Vec_PtrPush(front, left)
			Vec_PtrPush(front, c)
			want, area, leftRefs = []int32{c.Id, left.Id}, 2, 1
		}
		If_ManImproveNodeUpdate(p, root, front)
		cut := If_ObjCutBest(root)
		if !slices.Equal(cut.pLeaves, want) || cut.nLeaves != uint32(len(want)) || cut.nLimit != 4 || cut.uSign != If_ObjCutSignCompute(cut) {
			t.Fatalf("pass%d: malformed updated cut %+v, want leaves%v", pass, cut, want)
		}
		If_ManMarkMapping(p)
		if p.AreaGlo != area || root.nRefs != 2 || left.nRefs != leftRefs || a.nRefs != 1 || b.nRefs != 1 || c.nRefs != 1 {
			t.Fatalf("pass%d: area%g root/left/CI refs %d/%d/%d/%d/%d", pass, p.AreaGlo, root.nRefs, left.nRefs, a.nRefs, b.nRefs, c.nRefs)
		}
	}
}
