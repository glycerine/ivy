package cut

import (
	"slices"
	"testing"
)

func TestSequentialDominanceFiltersBeforeCutBudget(t *testing.T) {
	params := DefaultParams(4)
	params.FSeq, params.NKeepMax = 1, 2
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	p.oldCuts[8] = testCut(p, 16<<CUT_SHIFT, 17<<CUT_SHIFT)
	left := testCut(p, 16)
	left.Next = testCut(p, 18)
	p.newCuts[0], p.newCuts[1] = left, testCut(p, 17)
	Cut_NodeComputeCutsSeq(p, 8, 0, 1, 0, 0, 0, 0, 0, -1)
	c := p.newCuts[8]
	want := []int32{17 << CUT_SHIFT, 18 << CUT_SHIFT}
	if c == nil || c.Next != nil || !slices.Equal(c.leaves, want) {
		t.Fatalf("duplicate candidate consumed the only free cut slot, got %v", c)
	}
}

func TestSequentialCutReplacesDominatedOldSuperset(t *testing.T) {
	params := DefaultParams(4)
	params.FSeq, params.NKeepMax = 1, 3
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	top := testCut(p, 8)
	large := testCut(p, 16<<CUT_SHIFT, 17<<CUT_SHIFT, 18<<CUT_SHIFT)
	top.Next = large
	p.oldCuts[8] = top
	p.newCuts[0], p.newCuts[1] = testCut(p, 16), testCut(p, 17)
	Cut_NodeComputeCutsSeq(p, 8, 0, 1, 0, 0, 0, 0, 0, -1)
	if p.oldCuts[8] != top || top.Next != nil || large.owner != nil {
		t.Error("old frontier retained a cut dominated by the new pair")
	}
	c := p.newCuts[8]
	if c == nil || !slices.Equal(c.leaves, []int32{16 << CUT_SHIFT, 17 << CUT_SHIFT}) {
		t.Fatal("missing replacement pair")
	}
}

func TestSequentialChoiceUsesExistingRootPhaseAndCurrentFrontier(t *testing.T) {
	params := DefaultParams(4)
	params.FSeq = 1
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	root := testCut(p, 8)
	root.FSimul = true
	p.oldCuts[8] = root
	p.newCuts[8] = testCut(p, 16<<CUT_SHIFT, 17<<CUT_SHIFT)
	c := Cut_NodeUnionCutsSeq(p, []int32{8}, -1, 0)
	if c == nil || !c.FCompl {
		t.Error("later iteration lost the representative's simulation phase")
	}
	p.tempCuts[0] = testCut(p, 16<<CUT_SHIFT, 17<<CUT_SHIFT, 18<<CUT_SHIFT)
	if got := Cut_NodeUnionCutsSeq(p, []int32{8}, 0, 0); got != nil || p.newCuts[8] != c {
		t.Error("temporary choice escaped dominance by the current frontier")
	}
}

func BenchmarkSequentialCutFrontier(b *testing.B) {
	params := DefaultParams(4)
	params.FSeq, params.NKeepMax = 1, 2
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	p.oldCuts[8] = testCut(p, 16<<CUT_SHIFT, 17<<CUT_SHIFT)
	a := testCut(p, 16)
	a.Next = testCut(p, 18)
	p.newCuts[0], p.newCuts[1] = a, testCut(p, 17)
	b.ReportAllocs()
	for b.Loop() {
		Cut_NodeComputeCutsSeq(p, 8, 0, 1, 0, 0, 0, 0, 0, -1)
		c := p.newCuts[8]
		if c == nil || c.Next != nil || len(c.leaves) != 2 || c.leaves[0] != 17<<CUT_SHIFT || c.leaves[1] != 18<<CUT_SHIFT {
			b.Fatal("wrong frontier")
		}
		Cut_NodeFreeCuts(p, 8)
	}
}

func TestNetworkSequentialCutIterationUsesLastFreeSlot(t *testing.T) {
	params := DefaultParams(4)
	params.NKeepMax = 3
	net := cutTestNetwork{{ID: 0, Input: true}, {ID: 1, Input: true}, {ID: 2, Fanins: [2]int32{0, 1}},
		{ID: 3, Input: true}, {ID: 4, Fanins: [2]int32{2, 3}}}
	p := Abc_NtkSeqCuts(net, params)
	defer Cut_ManStop(p)
	for c := p.newCuts[4]; c != nil; c = c.Next {
		if slices.Equal(c.leaves, []int32{0, 1 << CUT_SHIFT, 3 << CUT_SHIFT}) {
			return
		}
	}
	t.Fatal("root lacked the three-input cut after reaching a fixed point")
}

func TestFirstSequentialChoiceKeepsOldRepresentativePhase(t *testing.T) {
	params := DefaultParams(4)
	params.FSeq = 1
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	p.oldCuts[8] = testCut(p, 8)
	p.oldCuts[8].FSimul = true
	p.newCuts[8] = testCut(p, 8)
	p.newCuts[8].Next = testCut(p, 16<<CUT_SHIFT, 17<<CUT_SHIFT)
	got := Cut_NodeUnionCutsSeq(p, []int32{8}, -1, 1)
	if got == nil || got.Next == nil || !got.Next.FCompl {
		t.Fatal("current trivial cut overwrote the existing representative phase")
	}
}

func TestIndependentUnshiftedSequentialInputsKeepNativeCounts(t *testing.T) {
	params := DefaultParams(4)
	params.FSeq = 1
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	a, b := testCut(p, 2), testCut(p, 3)
	p.newCuts[2], p.newCuts[3] = a, b
	before := p.stats
	Cut_NodeComputeCutsSeq(p, 4, 2, 3, 0, 0, 0, 0, 0, -1)
	if d := p.stats.Allocated - before.Allocated; d != 1 || p.stats.Recycled != before.Recycled {
		t.Fatalf("one native merge allocated %d cuts and recycled %d", d, p.stats.Recycled-before.Recycled)
	}
	if p.newCuts[2] != a || p.newCuts[3] != b || a.leaves[0] != 2<<CUT_SHIFT || b.leaves[0] != 3<<CUT_SHIFT {
		t.Fatal("changed source cuts")
	}
}

func TestChoiceCutBudgetKeepsNativeSourceOrder(t *testing.T) {
	for _, seq := range []int32{0, 1} {
		params := DefaultParams(4)
		params.FSeq, params.NKeepMax = seq, 2
		p := Cut_ManStart(params)
		rep, other := testCut(p, 8), testCut(p, 9)
		first, later := testCut(p, 2, 3, 4), testCut(p, 6, 7)
		rep.Next, other.Next = first, later
		p.newCuts[8], p.newCuts[9] = rep, other
		var c *Cut_Cut_t
		if seq != 0 {
			c = Cut_NodeUnionCutsSeq(p, []int32{8, 9}, -1, 1)
		} else {
			c = Cut_NodeUnionCuts(p, []int32{8, 9})
		}
		if c != rep || c.Next != first || first.Next != nil || later.owner != nil {
			t.Errorf("sequential %d kept a later choice before the first small cut", seq)
		}
		Cut_ManStop(p)
	}
}

func TestShiftedIndependentInputFrontiersBorrowAndRestore(t *testing.T) {
	params := DefaultParams(4)
	params.FSeq = 1
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	a, b := testCut(p, 2), testCut(p, 3)
	p.newCuts[2], p.newCuts[3] = a, b
	before := p.stats
	Cut_NodeComputeCutsSeq(p, 4, 2, 3, 0, 0, 1, 2, 0, -1)
	c := p.newCuts[4]
	if c == nil || !slices.Equal(c.leaves, []int32{2<<CUT_SHIFT | 1, 3<<CUT_SHIFT | 2}) ||
		a.leaves[0] != 2<<CUT_SHIFT || b.leaves[0] != 3<<CUT_SHIFT || a.USign != Cut_NodeSign(a.leaves[0]) || b.USign != Cut_NodeSign(b.leaves[0]) {
		t.Fatal("shifted input restoration", a, b, c)
	}
	if d := p.stats.Allocated - before.Allocated; d != 1 || p.stats.Recycled != before.Recycled {
		t.Fatalf("shifted independent inputs counted %d allocations and %d recycles", d, p.stats.Recycled-before.Recycled)
	}
}

func TestLatchOverflowPreservesEveryEarlierLeaf(t *testing.T) {
	params := DefaultParams(4)
	params.FSeq = 1
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	a := testCut(p, 2)
	b := testCut(p, 3<<CUT_SHIFT, 4<<CUT_SHIFT|255)
	a.Next = b
	signA, signB := a.USign, b.USign
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("missing overflow")
			}
		}()
		Cut_NodeShiftCutLeaves(a, 1)
	}()
	if a.leaves[0] != 2<<CUT_SHIFT || !slices.Equal(b.leaves, []int32{3 << CUT_SHIFT, 4<<CUT_SHIFT | 255}) || a.USign != signA || b.USign != signB {
		t.Fatal("failed shift changed earlier leaves or signatures")
	}
}

func TestOccupiedSequentialDestinationPreservesFrontiers(t *testing.T) {
	for _, set := range []int32{-1, 0} {
		params := DefaultParams(4)
		params.FSeq = 1
		p := Cut_ManStart(params)
		old := testCut(p, 8)
		superset := testCut(p, 16<<CUT_SHIFT, 17<<CUT_SHIFT, 18<<CUT_SHIFT)
		old.Next = superset
		p.oldCuts[8] = old
		existing := testCut(p, 23)
		if set < 0 {
			p.newCuts[8] = existing
		} else {
			p.tempCuts[set] = existing
		}
		p.newCuts[0], p.newCuts[1] = testCut(p, 16), testCut(p, 17)
		stats := p.stats
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("occupied destination accepted")
				}
			}()
			Cut_NodeComputeCutsSeq(p, 8, 0, 1, 0, 0, 0, 0, 0, set)
		}()
		if old.Next != superset || superset.owner != p.lifetime || stats != p.stats {
			t.Errorf("occupied destination %d mutated source cuts or counters", set)
		}
		Cut_ManStop(p)
	}
}

func TestShiftedSequentialInputsCanShareASuffix(t *testing.T) {
	params := DefaultParams(4)
	params.FSeq = 1
	p := Cut_ManStart(params)
	defer Cut_ManStop(p)
	left, right, shared := testCut(p, 10), testCut(p, 11), testCut(p, 12)
	left.Next, right.Next = shared, shared
	p.newCuts[0], p.newCuts[1] = left, right
	Cut_NodeComputeCutsSeq(p, 8, 0, 1, 0, 0, 1, 2, 0, -1)
	want := []int32{12<<CUT_SHIFT | 1, 12<<CUT_SHIFT | 2}
	for c := p.newCuts[8]; c != nil; c = c.Next {
		if slices.Equal(c.leaves, want) && shared.leaves[0] == 12<<CUT_SHIFT {
			return
		}
	}
	t.Fatal("shared input suffix received both shifts")
}
