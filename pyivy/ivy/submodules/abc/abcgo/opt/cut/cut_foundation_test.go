package cut

import "testing"

func TestCutTruthMergeAndLists(t *testing.T) {
	params := DefaultParams(4)
	params.FTruth = 1
	p := Cut_ManStart(params)
	a, b := Cut_CutCreateTriv(p, 1), Cut_CutCreateTriv(p, 2)
	c := Cut_CutMergeTwo(p, a, b)
	if len(c.leaves) != 2 || c.leaves[0] != 1 || c.leaves[1] != 2 {
		t.Fatal(c.leaves)
	}
	Cut_TruthCompute(p, c, a, b, 0, 1)
	if c.truth[0] != 0x22222222 {
		t.Fatal(c.truth)
	}
	var list Cut_List_t
	Cut_ListStart(&list)
	Cut_ListAdd2(&list, c)
	Cut_ListAdd2(&list, a)
	Cut_ListAdd2(&list, b)
	head := Cut_ListFinish(&list)
	if head != a || head.Next != b || head.Next.Next != c {
		t.Fatal("cut ordering")
	}
	copy := Cut_CutDupList(p, head)
	copy.leaves[0] = 77
	if head.leaves[0] != 1 {
		t.Fatal("shared leaves")
	}
	Cut_CutRecycleList(p, head)
	Cut_CutRecycleList(p, copy)
	if p.stats.Current != 0 {
		t.Fatal(p.stats)
	}
	Cut_ManStop(p)
}
func TestCutDroppingAndTruthPhase(t *testing.T) {
	p := Cut_ManStart(DefaultParams(3))
	Cut_NodeSetTriv(p, 3)
	Cut_ManSetFanoutCounts(p, []int32{0, 0, 0, 2})
	Cut_NodeTryDroppingCuts(p, 3)
	if Cut_NodeReadCutsNew(p, 3) == nil {
		t.Fatal("early drop")
	}
	Cut_NodeTryDroppingCuts(p, 3)
	if Cut_NodeReadCutsNew(p, 3) != nil {
		t.Fatal("not dropped")
	}
	merged := &Cut_Cut_t{leaves: []int32{2, 4, 6}}
	child := &Cut_Cut_t{leaves: []int32{2, 6}}
	if Cut_TruthPhase(merged, child) != 5 {
		t.Fatal("phase")
	}
}
