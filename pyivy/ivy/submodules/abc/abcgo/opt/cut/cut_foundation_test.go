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

func TestRecursiveEnumerationAreaAndManagerLifecycle(t *testing.T) {
	network := cutTestNetwork{
		{ID: 0, Input: true},
		{ID: 1, Input: true},
		{ID: 2, Fanins: [2]int32{0, 1}},
	}
	params := DefaultParams(4)
	params.FTruth = 1
	p := Cut_ManStart(params)
	p.network, p.networkCounts = networkInventory(network)
	list := Abc_NodeGetCutsRecursive(p, p.network[2], 0, 0)
	foundCut := false
	for c := list; c != nil; c = c.Next {
		if len(c.leaves) == 2 && c.leaves[0] == 0 && c.leaves[1] == 1 {
			foundCut = true
		}
	}
	if !foundCut {
		t.Fatalf("recursive enumeration did not produce cut [0 1]: %#v", list)
	}
	if p.stats.Current != p.stats.Allocated-p.stats.Recycled {
		t.Fatalf("cut allocation accounting: %+v", p.stats)
	}

	p.best[2] = testCut(p, 0, 1)
	p.best[0] = testCut(p, 0)
	p.best[1] = testCut(p, 1)
	if got := Cut_ManMappingArea_rec(p, 2); got != 1 {
		t.Fatalf("mapping area = %d, want 1", got)
	}
	if _, remains := p.best[2]; remains {
		t.Fatal("area recursion did not consume the root mapping")
	}

	Cut_ManStop(p)
	if !p.stopped || p.owned != nil || p.newCuts != nil {
		t.Fatal("manager stop did not release manager-owned state")
	}
}

func TestCutManagerRejectsInvalidAllocation(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("invalid cut parameters did not panic")
		}
	}()
	Cut_ManStart(nil)
}

func TestNetworkOracleTraversalMatchesRecordedEnumeration(t *testing.T) {
	network := cutTestNetwork{
		{ID: 0, Input: true},
		{ID: 1, Input: true},
		{ID: 2, Fanins: [2]int32{0, 1}, Compl: [2]bool{true, false}},
		{ID: 3, Fanins: [2]int32{2, 1}},
		{ID: 4, Output: true, Fanins: [2]int32{3, 0}},
	}
	params := DefaultParams(4)
	params.FTruth = 1
	params.FRecord = 1
	p := Abc_NtkCuts(network, params)
	oracle := Cut_OracleStart(p)
	Abc_NtkCutsOracle(network, oracle)
	for id, original := range p.newCuts {
		replayed := oracle.manager.newCuts[id]
		for a, b := original, replayed; a != nil && b != nil; a, b = a.Next, b.Next {
			if Cut_CutCompare(a, b) != 0 || len(a.truth) != len(b.truth) {
				t.Fatalf("node %d cut structure differs", id)
			}
			for i := range a.truth {
				if a.truth[i] != b.truth[i] {
					t.Fatalf("node %d truth word %d differs", id, i)
				}
			}
		}
		if (original == nil) != (replayed == nil) || Cut_CutCountList(original) != Cut_CutCountList(replayed) {
			t.Fatalf("node %d cut count differs", id)
		}
	}
	Cut_OracleStop(oracle)
	Cut_ManStop(p)
}
