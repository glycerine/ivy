package core

import "testing"

func TestCanonicalKeysSurviveGrowthAndPartialSweep(t *testing.T) {
	const parents, choices = 5, 16
	dd := Cudd_Init(parents+2*choices, 0, 8, 16, 0)
	defer Cudd_Quit(dd)
	type record struct {
		index      int
		t, e, node *DdNode
	}
	var kept, released []record
	for parent := range parents {
		for i := range choices {
			for j := range choices {
				high, low := dd.Vars[parents+i], Cudd_Not(dd.Vars[parents+choices+j])
				n := Cudd_BddNode(dd, parent, high, low)
				if n == nil || n.Index != uint32(parent) {
					t.Fatalf("missing canonical node at %d with children %d,%d", parent, i, j)
				}
				Cudd_Ref(n)
				r := record{parent, high, low, n}
				if (parent+i*choices+j)%7 == 0 {
					kept = append(kept, r)
				} else {
					released = append(released, r)
				}
			}
		}
	}
	base := uint32(parents + 2*choices + 4)
	if keys := Cudd_ReadKeys(dd); keys != base+parents*choices*choices {
		t.Fatalf("unique keys after growth = %d, want %d", keys, base+parents*choices*choices)
	}
	for _, r := range released {
		Cudd_RecursiveDeref(dd, r.node)
	}
	if got := Cudd_GarbageCollect(dd, true); int(got) != len(released) {
		t.Fatalf("swept %d nodes, want %d", got, len(released))
	}
	if got := Cudd_ReadKeys(dd); got != base+uint32(len(kept)) {
		t.Fatalf("%d surviving keys, want %d", got, base+uint32(len(kept)))
	}
	for _, r := range kept {
		if got := Cudd_BddNode(dd, r.index, Cudd_Not(r.t), Cudd_Not(r.e)); got != Cudd_Not(r.node) {
			t.Fatalf("lost complemented canonical handle %d", r.node.ID)
		}
	}
	for _, r := range released {
		if got := Cudd_BddNode(dd, r.index, r.t, r.e); got == nil || got == r.node {
			t.Fatalf("reinserted node %d returned stale handle %p", r.node.ID, got)
		}
	}
	if err := Cudd_CheckConsistency(dd); err != nil {
		t.Fatal(err)
	}
}

func TestUniqueProbeRetainsDistinctCollidingKeys(t *testing.T) {
	var table uniqueTable
	const count = 96
	nodes := make([]DdNode, count)
	children := make([]DdNode, count)
	other := &DdNode{ID: 21}
	// Distinct handles with equal IDs force all the pairs into one probe chain.
	key := func(i int) bddUniqueKey { return bddUniqueKey{17, &children[i], other} }
	for i := range nodes {
		table.set(key(i), &nodes[i])
	}
	for i := 0; i < count; i += 2 {
		table.remove(key(i))
	}
	for i := 1; i < count; i += 2 {
		if got := table.get(key(i)); got != &nodes[i] {
			t.Fatalf("key %d hidden behind a deleted slot", i)
		}
	}
	for i := 0; i < count; i += 2 {
		table.set(key(i), &nodes[i])
	}
	table.set(key(3), &nodes[4])
	if table.count != count || table.get(key(3)) != &nodes[4] {
		t.Fatalf("replacement changed the wrong entry or count %d", table.count)
	}
	seen := 0
	for k, node := range table.entries {
		if got := table.get(k); got != node {
			t.Fatalf("iteration disagrees with key lookup")
		}
		table.remove(k)
		seen++
	}
	if seen != count || table.count != 0 || table.get(key(1)) != nil {
		t.Fatalf("removed %d entries in one pass, %d remain", seen, table.count)
	}
}
