package cut

import "testing"

func dominatedMergeFixture(k int32, global bool) (*Cut_Man_t, *Cut_List_t, *Cut_Cut_t, *Cut_Cut_t) {
	params := DefaultParams(k)
	params.FTruth = 1
	if global {
		params.FGlobal = 1
	}
	p := Cut_ManStart(params)
	existing := testCut(p, 1, 2)
	list := new(Cut_List_t)
	Cut_ListStart(list)
	Cut_ListAdd(list, existing)
	return p, list, testCut(p, 1, 3), testCut(p, 2, 4)
}

func TestRejectedCutMergeUsesNoOwnedStorage(t *testing.T) {
	for _, k := range []int32{4, 8, 12} {
		for _, global := range []bool{false, true} {
			p, list, a, b := dominatedMergeFixture(k, global)
			before := p.stats
			allocs := testing.AllocsPerRun(20, func() {
				Cut_NodeDoComputeCuts(p, list, 8, 0, 0, a, b, 0, 0)
			})
			if allocs != 0 {
				t.Errorf("width %d global %t allocated %.0f times for a rejected cut", k, global, allocs)
			}
			// AllocsPerRun includes one warm-up merge.
			if list.count != 1 || p.stats.Current != before.Current || p.stats.Peak != before.Current+1 ||
				p.stats.Allocated-before.Allocated != 21 || p.stats.Filtered-before.Filtered != 21 ||
				p.stats.Recycled-before.Recycled != 21 || p.stats.Trivial != before.Trivial {
				t.Fatalf("width %d global %t rejected merge accounting %+v", k, global, p.stats)
			}
			Cut_ManStop(p)
		}
	}
}

func BenchmarkRejectDominatedFourLeafCut(b *testing.B) {
	p, list, a, c := dominatedMergeFixture(4, false)
	defer Cut_ManStop(p)
	b.ReportAllocs()
	for b.Loop() {
		Cut_NodeDoComputeCuts(p, list, 8, 0, 0, a, c, 0, 0)
	}
	if list.count != 1 {
		b.Fatal("lost dominating cut")
	}
}

func TestEmptyCutListCopyOwnsItsHeads(t *testing.T) {
	p := Cut_ManStart(DefaultParams(4))
	defer Cut_ManStop(p)
	var original Cut_List_t
	Cut_ListStart(&original)
	other := original
	c := testCut(p, 1, 2)
	Cut_ListAdd(&other, c)
	if Cut_ListFinish(&other) != c || Cut_ListFinish(&original) != nil {
		t.Fatal("adding to an empty copy wrote into the original cut list")
	}
}

func TestTemporaryCutListStaysOnStack(t *testing.T) {
	p := Cut_ManStart(DefaultParams(4))
	defer Cut_ManStop(p)
	c := testCut(p, 1, 2)
	allocs := testing.AllocsPerRun(50, func() {
		var list Cut_List_t
		Cut_ListStart(&list)
		Cut_ListAdd(&list, c)
		if Cut_ListFinish(&list) != c {
			t.Fatal("temporary list")
		}
	})
	if allocs != 0 {
		t.Fatalf("one temporary list allocated %.0f times", allocs)
	}
}
