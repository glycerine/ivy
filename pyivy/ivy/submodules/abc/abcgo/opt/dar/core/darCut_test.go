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
