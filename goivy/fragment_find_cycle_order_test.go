package goivy

import "testing"

// Python iu.cycle (ivy_utils.py) keeps each node's outgoing arcs in an
// insertion-ordered dict (m[first(arc)][arc] = None) and runs its DFS over
// them in that order, so the cycle it reports is the one reached through
// the first-inserted arc. (It used a set before, which made the reported
// cycle depend on PYTHONHASHSEED.)
func TestFindCycleFollowsArcInsertionOrder(t *testing.T) {
	a, b, c := NewUFNode(), NewUFNode(), NewUFNode()
	mk := func(from, to *UFNode) arc { return arc{from: from, to: to, fmla: True, argIdx: -1} }
	ab, ac, ba, ca := mk(a, b), mk(a, c), mk(b, a), mk(c, a)

	check := func(name string, arcs []arc, want ...arc) {
		t.Helper()
		got := (&checker{arcs: arcs}).findCycle()
		if len(got) != len(want) {
			t.Fatalf("%s: cycle %v, want %v", name, got, want)
		}
		for i := range want {
			if got[i].from != want[i].from || got[i].to != want[i].to {
				t.Fatalf("%s: cycle %v, want %v", name, got, want)
			}
		}
	}
	check("a->b first", []arc{ab, ac, ba, ba, ca}, ab, ba)
	check("a->c first", []arc{ac, ab, ba, ca}, ac, ca)
}
