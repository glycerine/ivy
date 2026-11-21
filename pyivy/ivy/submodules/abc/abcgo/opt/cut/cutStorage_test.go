package cut

import (
	"slices"
	"testing"
)

func TestStoppedCutManagerReleasesNetworkInventory(t *testing.T) {
	p := Cut_ManStart(DefaultParams(4))
	p.network, p.networkCounts = networkInventory(cutTestNetwork{{ID: 0, Input: true}, {ID: 1, Input: true}, {ID: 2, Fanins: [2]int32{0, 1}}})
	Cut_ManStop(p)
	if p.network != nil || p.networkCounts != nil {
		t.Fatal("stopped manager retained the network inventory")
	}
}

func TestRecycledCutReleasesListAndWideBuffers(t *testing.T) {
	params := DefaultParams(8)
	params.FTruth = 1
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	a, b := Cut_CutAlloc(p), Cut_CutAlloc(p)
	a.setLeaves([]int32{1, 2, 3, 4, 5, 6})
	a.proof, a.Next = &Cut_Proof{Node: 17}, b
	snapshot := Cut_CutDupList(p, a)
	for _, head := range []*Cut_Cut_t{a, snapshot} {
		next := head.Next
		Cut_CutRecycle(p, head)
		if head.Next != nil || head.proof != nil || head.leaves != nil || head.truth != nil {
			t.Errorf("temporary %t: recycled cut still references its list, proof or buffers", head.temporary)
		}
		if next == nil || next.owner != p.lifetime || len(next.truth) != 8 {
			t.Fatal("recycling the head changed its live successor")
		}
		Cut_CutRecycle(p, next)
	}
	if stats := p.Stats(); stats.Current != 0 || stats.Recycled != 2 {
		t.Fatalf("recycling regular and snapshot cuts changed statistics: %+v", stats)
	}
}

func TestCutBlocksPreserveHandlesAndReleaseUnusedSlots(t *testing.T) {
	params := DefaultParams(4)
	params.FTruth = 1
	p := Cut_ManStart(params)
	first := Cut_CutCreateTriv(p, 17)
	expired := Cut_CutCreateTriv(p, 19)
	Cut_CutRecycle(p, expired)
	for i := range 700 {
		c := Cut_CutCreateTriv(p, int32(i))
		if c == expired || c == first || c.self != c || c.owner != p.lifetime {
			t.Fatalf("lost cut identity after allocation %d", i)
		}
		c.truth[0] = 0
		Cut_CutRecycle(p, c)
	}
	if !slices.Equal(first.leaves, []int32{17}) || first.truth[0] == 0 || expired.self != nil {
		t.Fatal("an earlier live or recycled handle changed")
	}
	if len(p.freshCuts) == 0 {
		t.Fatal("fixture should stop partway through a block")
	}
	Cut_ManStop(p)
	if p.freshCuts != nil {
		t.Fatal("stopped manager retained unused cut slots")
	}
}

func TestCutHandleRejectsForeignAndRepeatedRecycle(t *testing.T) {
	params := DefaultParams(4)
	p, other := Cut_ManStart(params), Cut_ManStart(params)
	defer Cut_ManStop(other)
	a := Cut_CutCreateTriv(p, 1)
	b := Cut_CutCreateTriv(p, 2)
	reject := func(manager *Cut_Man_t, c *Cut_Cut_t) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Fatal("invalid recycle succeeded")
			}
		}()
		Cut_CutRecycle(manager, c)
	}
	reject(other, a)
	Cut_CutRecycle(p, a)
	reject(p, a)
	if p.Stats().Current != 1 || p.Stats().Trivial != 1 || other.Stats().Recycled != 0 {
		t.Fatal("invalid recycle changed statistics")
	}
	Cut_ManStop(p)
	reject(p, b)
	Cut_ManStop(p)
}

func TestCutManagerRejectsRecycleOfValueCopy(t *testing.T) {
	p := Cut_ManStart(DefaultParams(4))
	defer Cut_ManStop(p)
	plain := Cut_CutAlloc(p)
	trivial := Cut_CutCreateTriv(p, 7)
	snapshot := Cut_CutDupList(p, trivial)
	for _, original := range []*Cut_Cut_t{plain, trivial, snapshot} {
		copied := *original
		before := p.Stats()
		func() {
			defer func() {
				if recover() == nil {
					t.Error("manager accepted the value copy of a live cut")
				}
			}()
			Cut_CutRecycle(p, &copied)
		}()
		if got := p.Stats(); got != before {
			t.Errorf("recycling a copy changed counters from %+v to %+v", before, got)
		}
		Cut_CutRecycle(p, original)
	}
	if stats := p.Stats(); stats.Current != 0 || stats.Recycled != 2 || stats.Trivial != 0 {
		t.Errorf("original handles left incorrect counts: %+v", stats)
	}
}

func TestFourLeafCutOwnsOneAllocation(t *testing.T) {
	params := DefaultParams(4)
	params.FTruth = 1
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	a, b := testCut(p, 0, 3), testCut(p, 1, 2)
	allocs := testing.AllocsPerRun(100, func() {
		c := Cut_CutMergeTwo(p, a, b)
		if c == nil || len(c.truth) != 1 || !slices.Equal(c.leaves, []int32{0, 1, 2, 3}) {
			t.Fatal("bad four-leaf merge")
		}
		Cut_CutRecycle(p, c)
	})
	if allocs > 1 {
		t.Fatalf("merging four leaves and a truth word allocated %.0f objects, want at most one", allocs)
	}
	c := Cut_CutMergeTwo(p, a, b)
	c.truth[0] = 0xa5a5a5a5
	d := Cut_CutDupList(p, c)
	d.leaves[0], d.truth[0] = 17, 19
	if a.leaves[0] != 0 || c.leaves[0] != 0 || c.truth[0] != 0xa5a5a5a5 {
		t.Fatal("cut or duplicate aliases another cut")
	}
}

func BenchmarkMergeOwnedFourLeafCut(b *testing.B) {
	params := DefaultParams(4)
	params.FTruth = 1
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	a, y := testCut(p, 0, 3), testCut(p, 1, 2)
	b.ReportAllocs()
	for b.Loop() {
		c := Cut_CutMergeTwo(p, a, y)
		if c == nil || len(c.leaves) != 4 || c.leaves[3] != 3 {
			b.Fatal("merge")
		}
		Cut_CutRecycle(p, c)
	}
}

func TestMergedCutsOwnLeavesAndTruthAtEachSupportedWidth(t *testing.T) {
	for width := CUT_SIZE_MIN; width <= CUT_SIZE_MAX; width++ {
		params := DefaultParams(width)
		params.FTruth = 1
		p := Cut_ManStart(params)
		var left, right, want []int32
		for i := range width {
			want = append(want, i)
			if i%2 == 0 {
				left = append(left, i)
			} else {
				right = append(right, i)
			}
		}
		a, b := testCut(p, left...), testCut(p, right...)
		merged := Cut_CutMergeTwo(p, a, b)
		if merged == nil || !slices.Equal(merged.leaves, want) || len(merged.truth) != int(Cut_TruthWords(width)) {
			t.Fatalf("width %d merge", width)
		}
		for i := range merged.truth {
			merged.truth[i] = uint32(i + 9)
		}
		subset := Cut_CutMergeTwo(p, merged, b)
		clone := Cut_CutDupList(p, merged)
		subset.leaves[0]++
		clone.leaves[0]++
		clone.truth[len(clone.truth)-1]++
		if !slices.Equal(merged.leaves, want) || merged.truth[len(merged.truth)-1] != uint32(len(merged.truth)+8) {
			t.Fatalf("width %d shares storage", width)
		}
		if Cut_CutMergeTwo(p, merged, testCut(p, width)) != nil {
			t.Fatalf("width %d accepted too many leaves", width)
		}
		Cut_ManStop(p)
	}
}
