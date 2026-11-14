package core

import "testing"

func TestAverageDistanceLinksCoverAllBDDTableNodesAndChains(t *testing.T) {
	dd := Cudd_Init(3, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0)
	if dd == nil {
		t.Fatal("Cudd_Init failed")
	}
	// Force a single logical bucket to make collision-chain links observable.
	dd.bddTableSlots = []uint32{1, 1, 1}
	dd.initSlots = 1
	a := dd.node(0, dd.Vars[1], dd.Vars[2])
	b := dd.node(0, dd.Vars[1], dd.Zero)
	if a == nil || b == nil {
		t.Fatal("failed to create nodes in subtable 0")
	}
	links := Cudd_AverageDistanceLinks(dd)
	wantLinks := 2*len(dd.unique) + len(dd.constants) - 1
	perLevel := make(map[uint32]int)
	for _, node := range dd.unique {
		perLevel[node.Index]++
	}
	for _, count := range perLevel {
		wantLinks += count - 1 // each subtable has one forced bucket
	}
	if len(links) != wantLinks {
		t.Fatalf("got %d links, want %d T/E and collision-chain links", len(links), wantLinks)
	}
	nodeEntries := make(map[*DdNode]int)
	chainPairs := 0
	constantNodes := make(map[*DdNode]bool)
	for _, node := range dd.constants {
		constantNodes[node] = true
	}
	constantChainPairs := 0
	for _, pair := range links {
		if pair[0] == nil || pair[1] == nil {
			t.Fatal("nil endpoint in link list")
		}
		if _, ok := dd.unique[bddUniqueKey{pair[0].Index, pair[0].T, pair[0].E}]; ok {
			nodeEntries[pair[0]]++
			if pair[0].Index == 0 && pair[1] != Cudd_T(pair[0]) && pair[1] != Cudd_Regular(Cudd_E(pair[0])) {
				chainPairs++
			}
		}
		if constantNodes[pair[0]] && constantNodes[pair[1]] {
			constantChainPairs++
		}
	}
	for _, node := range dd.unique {
		if nodeEntries[node] < 2 {
			t.Fatalf("live BDD node %d contributes fewer than its T/E links", node.ID)
		}
	}
	if chainPairs == 0 {
		t.Fatal("expected a collision-chain link in the one-slot subtable")
	}
	if constantChainPairs != len(constantNodes)-1 {
		t.Fatalf("constant chain links = %d, want %d", constantChainPairs, len(constantNodes)-1)
	}
}
