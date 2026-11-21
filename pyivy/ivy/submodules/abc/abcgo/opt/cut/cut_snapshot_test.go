package cut

import "testing"

func makeSnapshotList(p *Cut_Man_t) *Cut_Cut_t {
	var head *Cut_Cut_t
	for i := int32(15); i >= 0; i-- {
		c := Cut_CutCreateTriv(p, i)
		c.Next = head
		head = c
	}
	return head
}

func TestTemporaryCutListPreservesNativeCounters(t *testing.T) {
	params := DefaultParams(4)
	params.FTruth = 1
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	source := makeSnapshotList(p)
	before := p.stats
	clone := Cut_CutDupList(p, source)
	if p.stats != before {
		t.Errorf("duplicating temporary list changed counters from %+v to %+v", before, p.stats)
	}
	clone.Next.leaves[0] = 97
	clone.Next.truth[0] = 0x12345678
	if source.Next.leaves[0] != 1 || source.Next.truth[0] != 0xaaaaaaaa {
		t.Fatal("temporary copy aliases the original")
	}
	Cut_CutRecycleList(p, clone)
	if p.stats != before {
		t.Errorf("recycling temporary list changed counters from %+v to %+v", before, p.stats)
	}
	Cut_CutRecycleList(p, source)
	if p.stats.Current != 0 || p.stats.Trivial != 0 || p.stats.Recycled != 16 {
		t.Fatalf("original cuts have incorrect lifetime counts: %+v", p.stats)
	}
}

func BenchmarkCutSnapshotList(b *testing.B) {
	params := DefaultParams(4)
	params.FTruth = 1
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	source := makeSnapshotList(p)
	b.ReportAllocs()
	for b.Loop() {
		clone := Cut_CutDupList(p, source)
		if clone.Next.leaves[0] != 1 {
			b.Fatal("bad cut snapshot")
		}
		Cut_CutRecycleList(p, clone)
	}
}

func TestSnapshotBlocksPreserveCutMetadataAndWideWords(t *testing.T) {
	params := DefaultParams(7)
	params.FTruth = 1
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	var source *Cut_Cut_t
	for i := range 85 {
		c := Cut_CutAlloc(p)
		c.setLeaves([]int32{0, 1, 2, 3, 4, int32(5 + i)})
		c.Num0, c.Num1 = uint16(i), uint16(i+1)
		c.UCanon0, c.UCanon1 = uint32(i+2), uint32(i+3)
		c.FCompl, c.FSimul = i%2 == 0, i%3 == 0
		c.proof = &Cut_Proof{Node: int32(i)}
		c.truth[3] = uint32(i + 4)
		c.Next, source = source, c
	}
	copy := Cut_CutDupList(p, source)
	if Cut_CutCountList(copy) != 85 {
		t.Fatal("short multi-block snapshot")
	}
	for a, b := source, copy; a != nil; a, b = a.Next, b.Next {
		if a == b || a.Num0 != b.Num0 || a.Num1 != b.Num1 || a.UCanon0 != b.UCanon0 || a.UCanon1 != b.UCanon1 || a.FCompl != b.FCompl || a.FSimul != b.FSimul || a.proof != b.proof || a.truth[3] != b.truth[3] || a.leaves[5] != b.leaves[5] {
			t.Fatalf("copied cut lost metadata at index %d", a.Num0)
		}
		b.leaves[5], b.truth[3] = -1, 0
		if a.leaves[5] < 0 || a.truth[3] == 0 {
			t.Fatal("wide payload shared with its snapshot")
		}
	}
	Cut_CutRecycleList(p, copy)
	if p.stats.Current != 85 || p.stats.Allocated != 85 {
		t.Fatalf("multi-block snapshot changed inventory: %+v", p.stats)
	}
}
