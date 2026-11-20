package core

import (
	"github.com/glycerine/abcgo/aig/aig"
	"reflect"
	"testing"
)

func TestCutSupportMinimization(t *testing.T) {
	for truth := 0; truth < 65536; truth++ {
		c := Cut{Leaves: []aig.Edge{2, 4, 6, 8}, Truth: uint16(truth)}
		Dar_CutSuppMinimize(&c)
		for assignment := 0; assignment < 16; assignment++ {
			index := 0
			for i, e := range c.Leaves {
				index |= ((assignment >> (e.NodeID() - 1)) & 1) << i
			}
			if ((c.Truth >> index) & 1) != uint16((truth>>assignment)&1) {
				t.Fatalf("truth %04x assignment %d reduced %+v", truth, assignment, c)
			}
		}
	}
}

// These slot contents follow Dar_ObjComputeCuts' nested physical-slot order,
// dominance filtering and value-based replacement in darCut.c.
func TestCutPhysicalSlotsAndEviction(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d, e := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	x, y := g.AND(a, b), g.AND(b, c)
	z := g.AND(x, y)
	tail := g.AND(d, e)
	root := g.AND(z, tail)
	g.AddOutput(root)
	for _, tc := range []struct {
		limit  int
		leaves [][]aig.Edge
		values []int
	}{
		{2, [][]aig.Edge{{root}, {a, b, c, tail}}, []int{1001, 5}},
		{3, [][]aig.Edge{{root}, {z, tail}, {a, b, c, tail}}, []int{1001, 2, 5}},
		{4, [][]aig.Edge{{root}, {z, tail}, {a, b, c, tail}, {a, b, y, tail}}, []int{1001, 2, 5, 5}},
	} {
		slots := NewCutEnumerator(g, tc.limit).Slots(root)
		if len(slots) != tc.limit {
			t.Fatalf("limit %d: %d slots", tc.limit, len(slots))
		}
		for i, slot := range slots {
			if !slot.Used || !reflect.DeepEqual(slot.Cut.Leaves, tc.leaves[i]) || slot.Value != tc.values[i] {
				t.Fatalf("limit %d slot %d: %+v", tc.limit, i, slot)
			}
		}
	}
	slots := NewCutEnumerator(g, 10).Slots(root)
	if !slots[7].Used || slots[8].Used || slots[9].Used {
		t.Fatalf("unused physical slots: %+v", slots)
	}
	for i, x := range slots {
		if !x.Used {
			continue
		}
		for j, y := range slots {
			if i != j && y.Used && Dar_CutCheckDominance(x.Cut, y.Cut) {
				t.Fatalf("slot %d dominates slot %d", i, j)
			}
		}
	}
}
func TestCutMergeTruth(t *testing.T) {
	a := Cut{Leaves: []aig.Edge{2, 6}, Truth: 0x6666}
	b := Cut{Leaves: []aig.Edge{4, 6}, Truth: 0xeeee}
	c, ok := Dar_CutMerge(a, b)
	if !ok {
		t.Fatal("merge")
	}
	c.Truth = Dar_CutTruth(c, a, b, false, true)
	for n := 0; n < 16; n++ {
		x, y, z := n&1 != 0, n&2 != 0, n&4 != 0
		want := (x != z) && !(y || z)
		if (c.Truth&(1<<n) != 0) != want {
			t.Fatal(n, c.Truth)
		}
	}
	if _, ok := Dar_CutMerge(Cut{Leaves: []aig.Edge{2, 4, 6}}, Cut{Leaves: []aig.Edge{8, 10}}); ok {
		t.Fatal("oversized union")
	}
}
func TestCutEnumeratorTruth(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.XOR(g.AND(a, b.Not()), c)
	enumerator := NewCutEnumerator(g, 32)
	cuts := enumerator.Cuts(root)
	if len(cuts) < 2 {
		t.Fatal("no nontrivial cuts")
	}
	found := false
	for _, cut := range cuts {
		if len(cut.Leaves) != 3 || cut.Leaves[0] != a || cut.Leaves[1] != b || cut.Leaves[2] != c {
			continue
		}
		found = true
		for n := 0; n < 8; n++ {
			values := map[aig.Edge]bool{a: n&1 != 0, b: n&2 != 0, c: n&4 != 0}
			index := 0
			for i, e := range cut.Leaves {
				if values[e] {
					index |= 1 << i
				}
			}
			want := (values[a] && !values[b]) != values[c]
			if (cut.Truth&(1<<index) != 0) != want {
				t.Fatal(n, cut)
			}
		}
	}
	if !found {
		t.Fatal("missing primary-input cut")
	}
	// Returned cuts are owned by caller, not aliases of the memoized boundaries.
	original := cuts[0].Leaves[0]
	cuts[0].Leaves[0] = a
	if enumerator.Cuts(root)[0].Leaves[0] != original {
		t.Fatal("borrowed cut boundary")
	}
	again := NewCutEnumerator(g, 32).Cuts(root.Not())
	if again[0].Truth != ^NewCutEnumerator(g, 32).Cuts(root)[0].Truth {
		t.Fatal("root phase")
	}
}

func TestWholeGraphCutEnumerationDropsDanglingCones(t *testing.T) {
	for _, physical := range []bool{true, false} {
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		kept := g.AND(a, b)
		unused := g.AND(kept, c)
		tail := g.AND(unused, b.Not())
		g.AddOutput(kept)
		var contains func(aig.Edge) bool
		if physical {
			cuts, err := Dar_ManComputeCutSlots(g, 10, false)
			if err != nil {
				t.Fatal(err)
			}
			contains = func(e aig.Edge) bool { _, ok := cuts[e]; return ok }
		} else {
			cuts, err := Dar_ManComputeCuts(g, 10, false)
			if err != nil {
				t.Fatal(err)
			}
			contains = func(e aig.Edge) bool { _, ok := cuts[e]; return ok }
		}
		if contains(unused) || contains(tail) || !contains(kept) || !contains(c) ||
			!aig.Aig_ObjIsNone(g, unused) || !aig.Aig_ObjIsNone(g, tail) || g.Node(kept).Refs != 1 {
			t.Errorf("physical %v: dangling cone retained, kept gate refs %d", physical, g.Node(kept).Refs)
		}
	}
}

func TestCutMergeOwnedSortedSet(t *testing.T) {
	a := Cut{Leaves: []aig.Edge{8, 2, 8}}
	b := Cut{Leaves: []aig.Edge{6, 4, 2}}
	merged, ok := Dar_CutMerge(a, b)
	if !ok || !reflect.DeepEqual(merged.Leaves, []aig.Edge{2, 4, 6, 8}) {
		t.Fatalf("union %v, ok %v", merged.Leaves, ok)
	}
	merged.Leaves[0] = 12
	if a.Leaves[1] != 2 || b.Leaves[2] != 2 {
		t.Fatal("merged leaves alias a caller buffer")
	}
	oversized := Cut{Leaves: []aig.Edge{6, 4, 10, 12}}
	if got := testing.AllocsPerRun(100, func() {
		if _, ok := Dar_CutMerge(a, oversized); ok {
			t.Fatal("accepted six leaves")
		}
	}); got != 0 {
		t.Fatalf("rejected merge allocated %.0f objects", got)
	}
}

func TestCutStretchAllFourInputPlacements(t *testing.T) {
	all := []aig.Edge{2, 4, 6, 8}
	for mask := 0; mask < 16; mask++ {
		var leaves []aig.Edge
		for i, leaf := range all {
			if mask&(1<<i) != 0 {
				leaves = append(leaves, leaf)
			}
		}
		for value := 0; value < 1<<(1<<len(leaves)); value++ {
			have := cutStretch(Cut{leaves, uint16(value)}, all)
			for assignment := 0; assignment < 16; assignment++ {
				input := 0
				for i, leaf := range leaves {
					input |= ((assignment >> (leaf.NodeID() - 1)) & 1) << i
				}
				if (int(have)>>assignment)&1 != (value>>input)&1 {
					t.Fatalf("placement %04b function %04x assignment %d: %04x", mask, value, assignment, have)
				}
			}
		}
	}
}

func TestCachedCutsHaveIndependentBoundaries(t *testing.T) {
	g := layeredCutGraph()
	edge := g.Outputs()[0]
	m := NewCutEnumerator(g, 10)
	m.skipTruthMin = true
	want := m.Slots(edge)
	for _, physical := range []bool{true, false} {
		allocs := testing.AllocsPerRun(100, func() {
			if physical {
				m.Slots(edge)
			} else {
				m.Cuts(edge)
			}
		})
		if allocs > 2 {
			t.Errorf("physical %v: copying cached cuts allocated %.0f objects", physical, allocs)
		}
	}
	slots := m.Slots(edge)
	for i := range slots {
		// Exercise writes through the slice and append into a retained result.
		x := &slots[i].Cut
		for j := range x.Leaves {
			x.Leaves[j] = aig.InvalidEdge
		}
		x.Leaves = append(x.Leaves, 0)
		if i+1 < len(slots) && !reflect.DeepEqual(slots[i+1], want[i+1]) {
			t.Fatalf("appending to slot %d changed its neighbor", i)
		}
	}
	if !reflect.DeepEqual(m.Slots(edge), want) {
		t.Fatal("editing the caller's slots damaged the cache")
	}
}

func TestCandidatesOwnLeavesAndRespectEdgePhase(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	gate := g.NewXor(g.AND(a, b.Not()), c)
	buf := aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{gate.Not(), aig.InvalidEdge}})
	output := g.AddOutput(buf)
	m := NewCutEnumerator(g, 10)
	want := m.Slots(gate)
	for _, edge := range []aig.Edge{gate, gate.Not(), buf, buf.Not(), output, output.Not()} {
		invert := edge == gate.Not() || edge == buf || edge == output
		seen := 0
		for i, cut := range m.Candidates(edge) {
			slot := want[i]
			truth := slot.Cut.Truth
			if invert {
				truth = ^truth
			}
			if !slot.Used || truth != cut.Truth || slot.Value != int(cut.Value) ||
				!reflect.DeepEqual(slot.Cut.Leaves, cut.Leaves[:cut.NumLeaves]) {
				t.Fatalf("edge %v, slot %d: %+v, want %+v (inverted %v)", edge, i, cut, slot, invert)
			}
			cut.Leaves[0] = aig.InvalidEdge
			seen++
		}
		if seen != len(m.Cuts(gate)) {
			t.Fatal("missed a used slot")
		}
	}
	count := 0
	for range m.Candidates(gate) {
		count++
		break
	}
	if count != 1 || !reflect.DeepEqual(m.Slots(gate), want) {
		t.Fatal("early exit or owned leaves")
	}
}
