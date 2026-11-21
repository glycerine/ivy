package mf

import (
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func addTestCut(s *cutSet, c Cut) { s.add(c, signature(c.Leaves)) }

func TestCutMergeAndSourceTieOrder(t *testing.T) {
	leaves, ok := mergeLeaves(nil, []int32{1, 3, 65}, []int32{2, 3}, 4)
	if !ok || len(leaves) != 4 || leaves[2] != 3 || leaves[3] != 65 {
		t.Fatal("ordered merge", leaves)
	}
	if _, ok := mergeLeaves(nil, []int32{1, 65}, []int32{129}, 2); ok {
		t.Fatal("signature collisions must not permit oversized cuts")
	}
	s := newCutSet(8)
	addTestCut(s, Cut{Leaves: []int32{1, 3}, Function: 4, Flow: 1, Delay: 1})
	addTestCut(s, Cut{Leaves: []int32{2, 3}, Function: 6, Flow: 1, Delay: 1})
	if s.slots[s.order[0]].cut.Function != 6 {
		t.Fatal("source stable tie rule")
	}
	addTestCut(s, Cut{Leaves: []int32{3}, Function: 2, Flow: 0})
	if s.count != 1 || s.slots[s.order[0]].cut.Function != 2 || s.slots[s.order[0]].sign != 1<<3 {
		t.Fatal("strict supersets survived")
	}
}

func TestFullCutSetOnlyCopiesRetainedCandidates(t *testing.T) {
	s := newCutSet(3)
	addTestCut(s, Cut{Leaves: []int32{1, 2}, Flow: 1})
	addTestCut(s, Cut{Leaves: []int32{3, 4}, Flow: 2})
	candidate := Cut{Leaves: []int32{5, 6}, Flow: 3}
	if allocs := testing.AllocsPerRun(20, func() { addTestCut(s, candidate) }); allocs != 0 {
		t.Fatalf("discarding a lower-ranked cut allocated %.0f times", allocs)
	}
	candidate.Flow = 1
	addTestCut(s, candidate)
	candidate.Leaves[0] = 0
	cuts := s.save()
	if len(cuts) != 2 || cuts[0].Leaves[0] != 5 || cuts[1].Leaves[0] != 1 {
		t.Fatal("equal-ranked candidate lost its order or ownership", cuts)
	}
}

func TestCutSignaturesRejectOnlyImpossibleSubsets(t *testing.T) {
	s := newCutSet(5)
	addTestCut(s, Cut{Leaves: []int32{1, 2}})
	for _, trial := range []struct {
		leaves []int32
		want   bool
	}{
		{[]int32{65, 66}, false},
		{[]int32{1, 66, 70}, false},
		{[]int32{1, 2, 66}, true},
		{[]int32{0, 1}, false},
	} {
		if got := s.contains(trial.leaves, signature(trial.leaves)); got != trial.want {
			t.Fatalf("subset check for %v got %t", trial.leaves, got)
		}
	}
	addTestCut(s, Cut{Leaves: []int32{65}})
	if s.count != 2 {
		t.Fatal("a signature collision removed distinct leaves")
	}
	for i, id := range s.order[:s.count] {
		if s.slots[id].sign != signature(s.slots[id].cut.Leaves) {
			t.Fatalf("signature lost its cut at position %d", i)
		}
	}
}

func TestSavedCutLeavesSurviveSlotReuseAndAppend(t *testing.T) {
	s := newCutSet(5)
	for i, leaves := range [][]int32{{1, 4}, {2, 5}, {1, 3}, {8}, {1}} {
		addTestCut(s, Cut{Leaves: leaves, Flow: float32(i)})
	}
	// Removing the first and third candidates leaves three live records.
	saved := s.save()
	if len(saved) != 3 {
		t.Fatalf("saved %d cuts after superset removal", len(saved))
	}
	s.count = 0
	for i := range 7 {
		addTestCut(s, Cut{Leaves: []int32{int32(20 + i)}, Flow: float32(-i)})
	}
	for i, want := range [][]int32{{2, 5}, {8}, {1}} {
		if !slices.Equal(saved[i].Leaves, want) {
			t.Fatalf("saved cut %d changed after slot reuse: %v", i, saved[i])
		}
	}
	extended := append(saved[0].Leaves, 99)
	if extended[2] != 99 || saved[1].Leaves[0] != 8 {
		t.Fatal("appending a saved cut changed its neighbor")
	}
	s.count = 0
	addTestCut(s, Cut{})
	if s.save()[0].Leaves != nil {
		t.Fatal("constant cut retained leaf storage")
	}
}
func TestTruthAlignmentAndSupportMinimization(t *testing.T) {
	// x0 XOR x1 expanded onto leaves 1 and 3 of a three-variable merged cut.
	table := []uint64{0x6666666666666666, 0x6666666666666666, 0x6666666666666666, 0x6666666666666666}
	expanded := make([]uint64, len(table))
	expandTruth(expanded, table, []int32{1, 3}, []int32{1, 2, 3}, 8, true)
	for a := 0; a < 256; a++ {
		want := uint64(1 ^ ((a & 1) ^ ((a >> 2) & 1)))
		if expanded[a>>6]>>uint(a&63)&1 != want {
			t.Fatal("expanded XOR assignment", a)
		}
	}
	minimized, leaves := minimizeSupport(expanded, []int32{1, 2, 3}, 8)
	if len(leaves) != 2 || leaves[0] != 1 || leaves[1] != 3 {
		t.Fatal("support", leaves)
	}
	for _, word := range minimized {
		if word != ^uint64(0x6666666666666666) {
			t.Fatalf("minimized truth %x", word)
		}
	}
	constant, leaves := minimizeSupport(make([]uint64, 4), []int32{1, 2}, 8)
	if len(leaves) != 0 || len(constant) != 4 {
		t.Fatal("constant support")
	}
}

func TestInternOwnsCandidateTruth(t *testing.T) {
	m := &Mapping{Options: DefaultOptions(), truthStore: vec.Vec_MemAlloc(4, 1)}
	vec.Vec_MemHashAlloc(m.truthStore, 2)
	table := []uint64{0, 0, 0, 0}
	zero := m.intern(table, 0)
	table[3] = 1 << 63
	nonzero := m.intern(table, 8)
	saved := vec.Vec_MemReadEntry(m.truthStore, int(nonzero))
	if zero == nonzero || vec.Vec_MemReadEntry(m.truthStore, int(zero))[3] != 0 || saved[3] != table[3] {
		t.Fatal("interned truths lost ownership or differ only in ignored high bits")
	}
	table[3] = 0
	for i := uint(0); i < 32; i++ {
		table[1] = uint64(1) << i
		if id := m.intern(table, 8); id != int32(i+2) {
			t.Fatalf("insertion after page growth got ID %d at %d", id, i)
		}
	}
	table[1] = 0
	if again := m.intern(table, 0); again != zero || saved[3] != 1<<63 {
		t.Fatal("reusing scratch changed a previously interned truth")
	}
}

func TestInternClauseCostAtEachTableWidth(t *testing.T) {
	const parity6 uint64 = 0x6996966996696996
	for width := 6; width <= 8; width++ {
		options := DefaultOptions()
		options.LUTSize = width
		words := 1 << (width - 6)
		m := &Mapping{Options: options, truthStore: vec.Vec_MemAlloc(words, 2)}
		vec.Vec_MemHashAlloc(m.truthStore, 4)
		for support := 6; support <= width; support++ {
			table := [4]uint64{parity6, parity6, parity6, parity6}
			if support > 6 {
				table[1], table[3] = ^parity6, ^parity6
			}
			if support == 8 {
				table[2], table[3] = ^table[2], ^table[3]
			}
			id := m.intern(table[:words], support)
			var cover [512]int32
			want := util.Abc_Tt8Cnf(table[:], int32(support), cover[:])
			if m.ClauseSizes[id] != int(want) {
				t.Fatalf("LUT width %d support %d got cost %d, cover has %d cubes", width, support, m.ClauseSizes[id], want)
			}
		}
	}
}
