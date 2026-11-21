package mf

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestCutMergeAndSourceTieOrder(t *testing.T) {
	leaves, ok := mergeLeaves(nil, []int32{1, 3, 65}, []int32{2, 3}, 4)
	if !ok || len(leaves) != 4 || leaves[2] != 3 || leaves[3] != 65 {
		t.Fatal("ordered merge", leaves)
	}
	if _, ok := mergeLeaves(nil, []int32{1, 65}, []int32{129}, 2); ok {
		t.Fatal("signature collisions must not permit oversized cuts")
	}
	cuts := []Cut{{Leaves: []int32{1, 3}, Function: 4, Flow: 1, Delay: 1}}
	cuts = addCut(cuts, Cut{Leaves: []int32{2, 3}, Function: 6, Flow: 1, Delay: 1}, 8)
	if cuts[0].Function != 6 {
		t.Fatal("source stable tie rule")
	}
	cuts = addCut(cuts, Cut{Leaves: []int32{3}, Function: 2, Flow: 0}, 8)
	if len(cuts) != 1 || cuts[0].Function != 2 {
		t.Fatal("strict supersets survived")
	}
}

func TestFullCutSetOnlyCopiesRetainedCandidates(t *testing.T) {
	cuts := []Cut{
		{Leaves: []int32{1, 2}, Flow: 1},
		{Leaves: []int32{3, 4}, Flow: 2},
	}
	candidate := Cut{Leaves: []int32{5, 6}, Flow: 3}
	if allocs := testing.AllocsPerRun(20, func() { cuts = addCut(cuts, candidate, 3) }); allocs != 0 {
		t.Fatalf("discarding a lower-ranked cut allocated %.0f times", allocs)
	}
	candidate.Flow = 1
	cuts = addCut(cuts, candidate, 3)
	candidate.Leaves[0] = 0
	if len(cuts) != 2 || cuts[0].Leaves[0] != 5 || cuts[1].Leaves[0] != 1 {
		t.Fatal("equal-ranked candidate lost its order or ownership", cuts)
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
