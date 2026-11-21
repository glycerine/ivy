package mf

import (
	"math/bits"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func addTestCut(s *cutSet, c Cut) { s.add(c, signature(c.Leaves)) }

func TestGateLeafIndexAtCapacityAndAfterOverflow(t *testing.T) {
	s := newCutSet(4)
	s.index.position = make([]uint8, 66*64)
	var buffer [maxLeaves]int32
	for _, size := range []int{64, 65, 3, 0, 64} {
		var fanins [4][]Cut
		for i := size - 1; i >= 0; i-- {
			part := i % len(fanins)
			fanins[part] = append(fanins[part], Cut{Leaves: []int32{int32(i * 64)}})
		}
		s.count = 0
		s.index.reset(fanins)
		if s.index.exact != (size <= 64) {
			t.Fatalf("unexpected index mode for %d leaves", size)
		}
		if size < 3 {
			continue
		}
		last := int32((size - 1) * 64)
		if s.index.exact && s.index.mask([]int32{0, last}) != (uint64(1)|uint64(1)<<(size-1)) {
			t.Fatalf("wrong end positions at size %d", size)
		}
		first, end := []int32{0, 64}, []int32{64, last}
		sign := s.index.mask(first) | s.index.mask(end)
		if _, ok := s.candidate(&buffer, first, end, sign, 2); ok {
			t.Fatalf("accepted three distinct leaves at limit two, size %d", size)
		}
		got, ok := s.candidate(&buffer, first, end, sign, 3)
		if !ok || !slices.Equal(got, []int32{0, 64, last}) {
			t.Fatalf("size %d produced %v, accepted %t", size, got, ok)
		}
		s.add(Cut{Leaves: end}, s.index.mask(end))
		if _, ok := s.candidate(&buffer, first, end, sign, 3); ok {
			t.Fatalf("accepted a dominated union at size %d", size)
		}
		if _, ok := s.candidate(&buffer, first, nil, s.index.mask(first), 3); !ok {
			t.Fatalf("rejected distinct colliding IDs at size %d", size)
		}
	}
}

func TestDisjointCutTruthAtEveryStoredWord(t *testing.T) {
	for width := 3; width <= 8; width++ {
		words := 1 << max(0, width-6)
		m := Mapping{
			Options:      Options{LUTSize: width},
			truthStore:   vec.Vec_MemAlloc(words, 2),
			truthResults: make([]truthResult, 8),
		}
		vec.Vec_MemHashAlloc(m.truthStore, 4)
		m.intern(make([]uint64, words), 0)
		for count := 2; count <= width; count++ {
			var inputs [2]Cut
			leaves := make([]int32, count)
			for i := range leaves {
				leaves[i] = int32(7 + 64*i)
				inputs[i%2].Leaves = append(inputs[i%2].Leaves, leaves[i])
			}
			for i := range inputs {
				n := len(inputs[i].Leaves)
				table := make([]uint64, words)
				for a := 0; a < 64*words; a++ {
					if bits.OnesCount(uint(a&((1<<n)-1)))%2 != 0 {
						table[a/64] |= 1 << (a % 64)
					}
				}
				inputs[i].Function = m.intern(table, n) << 1
			}
			for _, kind := range []int{2, 3} {
				cut := m.computeTruth(inputs[:], []int32{1, 0}, kind, slices.Clone(leaves))
				if !slices.Equal(cut.Leaves, leaves) {
					t.Fatalf("width %d gate %d retained %v, want %v", width, kind, cut.Leaves, leaves)
				}
				table := vec.Vec_MemReadEntry(m.truthStore, int(cut.Function>>1))
				for a := 0; a < words*64; a++ {
					active := a & ((1 << count) - 1)
					left := bits.OnesCount(uint(active&0x55))%2 == 0
					right := bits.OnesCount(uint(active&0xaa))%2 != 0
					want := left && right
					if kind == 3 {
						want = left != right
					}
					got := (table[a/64]>>(a%64)&1)^uint64(cut.Function&1) != 0
					if got != want {
						t.Fatalf("width %d count %d gate %d assignment %d: got %t, want %t", width, count, kind, a, got, want)
					}
				}
			}
		}
	}
}

func TestCutMergeAndSourceTieOrder(t *testing.T) {
	var buffer [maxLeaves]int32
	leaves, ok := mergeLeaves(&buffer, []int32{1, 3, 65}, []int32{2, 3}, 4)
	if !ok || len(leaves) != 4 || leaves[2] != 3 || leaves[3] != 65 {
		t.Fatal("ordered merge", leaves)
	}
	if _, ok := mergeLeaves(&buffer, []int32{1, 65}, []int32{129}, 2); ok {
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
	expandTruth(expanded, table, 2, leafPositions([]int32{1, 3}, []int32{1, 2, 3}), 8, true)
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

func TestTruthOperationReusesOnlyRelativeSupport(t *testing.T) {
	for _, width := range []int{6, 8} {
		for _, objectIDs := range []bool{false, true} {
			words := 1 << (width - 6)
			m := &Mapping{
				Options:      Options{LUTSize: width, ObjectIDs: objectIDs},
				truthStore:   vec.Vec_MemAlloc(words, 2),
				truthResults: make([]truthResult, 1), // Force every distinct operation to collide.
			}
			vec.Vec_MemHashAlloc(m.truthStore, 4)
			m.intern(make([]uint64, words), 0)
			variable := make([]uint64, words)
			for i := range variable {
				variable[i] = 0xaaaaaaaaaaaaaaaa
			}
			x := m.intern(variable, 1) << 1
			for _, tc := range []struct {
				kind        int
				args, flips []int32
				keep        []int
				truth       uint8 // indexed by the two original inputs
			}{
				{4, []int32{1, 1, 0}, []int32{0, 0, 0}, []int{1}, 0b1100},
				{4, []int32{1, 1, 0}, []int32{1, 0, 1}, []int{0, 1}, 0b0110},
				{2, []int32{0, 1}, []int32{1, 0}, []int{0, 1}, 0b0100},
				{3, []int32{0, 1}, []int32{0, 1}, []int{0, 1}, 0b1001},
				{4, []int32{0, 0, 1}, []int32{1, 1, 0}, []int{0}, 0b0101},
			} {
				for _, ids := range [][2]int32{{2, 5}, {13, 47}} {
					cuts := make([]Cut, len(tc.args))
					for i, which := range tc.args {
						cuts[i] = Cut{Function: x, Leaves: []int32{ids[which]}}
					}
					merged := ids
					got := m.computeTruth(cuts, tc.flips, tc.kind, merged[:])
					keep := tc.keep
					if width == 6 && objectIDs {
						keep = []int{0, 1}
					}
					want := make([]int32, len(keep))
					for j, which := range keep {
						want[j] = ids[which]
					}
					if !slices.Equal(got.Leaves, want) {
						t.Fatalf("width %d object IDs %t gate %d with %v: leaves %v, want %v", width, objectIDs, tc.kind, ids, got.Leaves, want)
					}
					table := vec.Vec_MemReadEntry(m.truthStore, int(got.Function>>1))
					for assignment := 0; assignment < 4; assignment++ {
						index := 0
						for j, which := range keep {
							index |= (assignment >> which & 1) << j
						}
						value := uint8(table[0]>>index&1) ^ uint8(got.Function&1)
						if value != tc.truth>>assignment&1 {
							t.Fatalf("width %d object IDs %t gate %d with %v at %d: got %d", width, objectIDs, tc.kind, ids, assignment, value)
						}
					}
				}
			}
		}
	}
}

func TestTruthCacheGrowsOnMisses(t *testing.T) {
	m := &Mapping{
		Options: Options{LUTSize: 6}, truthStore: vec.Vec_MemAlloc(1, 2),
		truthResults: make([]truthResult, 2), missesBeforeGrowth: 2,
	}
	vec.Vec_MemHashAlloc(m.truthStore, 4)
	m.intern([]uint64{0}, 0)
	x := m.intern([]uint64{0xaaaaaaaaaaaaaaaa}, 1) << 1
	cuts := []Cut{{Leaves: []int32{4}, Function: x}, {Leaves: []int32{7}, Function: x}}
	compute := func(flip int32) int32 {
		return m.computeTruth(cuts, []int32{flip, 0}, 2, []int32{4, 7}).Function
	}
	positive := compute(0)
	if compute(0) != positive || len(m.truthResults) != 2 || m.missesBeforeGrowth != 1 {
		t.Fatal("a cache hit consumed the budget for growth")
	}
	negative := compute(1)
	if len(m.truthResults) != 32 || m.missesBeforeGrowth != 0 || positive == negative {
		t.Fatalf("after the second miss: %d slots, %d misses remaining, functions %d/%d", len(m.truthResults), m.missesBeforeGrowth, positive, negative)
	}
	// The operation which caused growth must have been stored in the new table.
	if !slices.ContainsFunc(m.truthResults, func(entry truthResult) bool { return entry.function == negative }) {
		t.Fatal("missing result for the operation that grew the table")
	}
	if compute(0) != positive || compute(1) != negative || len(m.truthResults) != 32 {
		t.Fatal("recomputing displaced entries changed the result or grew the cache again")
	}
}

func TestBoundedCutUnion(t *testing.T) {
	ids := []int32{1, 2, 64, 65, 128, 4111, 8000000}
	sets := make([][]int32, 1<<len(ids))
	for mask := range sets {
		for i, id := range ids {
			if mask>>i&1 != 0 {
				sets[mask] = append(sets[mask], id)
			}
		}
	}
	var buffer [maxLeaves]int32
	for left, a := range sets {
		for right, b := range sets {
			want := sets[left|right]
			for limit := 0; limit <= 8; limit++ {
				got, ok := mergeLeaves(&buffer, a, b, limit)
				if ok != (len(want) <= limit) || ok && !slices.Equal(got, want) {
					t.Fatalf("merge %v and %v at %d: got %v, %t; union %v", a, b, limit, got, ok, want)
				}
			}
		}
	}
}
