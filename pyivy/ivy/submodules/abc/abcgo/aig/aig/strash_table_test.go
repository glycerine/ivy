package aig

import (
	"math/rand/v2"
	"testing"
)

// The strash table must behave exactly like a map under random
// insert/overwrite/delete sequences, including deletions inside probe runs.
func TestStrashTableMatchesMap(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	var tab strashTable
	ref := map[edgePair]Edge{}
	for step := 0; step < 200000; step++ {
		k := edgePair{Edge(r.IntN(300)), Edge(r.IntN(300))}
		switch r.IntN(3) {
		case 0, 1:
			v := Edge(r.Uint32() >> 1)
			tab.set(k, v)
			ref[k] = v
		case 2:
			tab.del(k)
			delete(ref, k)
		}
		if tab.len() != len(ref) {
			t.Fatalf("step %d: len %d, want %d", step, tab.len(), len(ref))
		}
		probe := edgePair{Edge(r.IntN(300)), Edge(r.IntN(300))}
		got, ok := tab.get(probe)
		want, wantOK := ref[probe]
		if ok != wantOK || got != want {
			t.Fatalf("step %d: get %v = %v,%v want %v,%v", step, probe, got, ok, want, wantOK)
		}
	}
	for k, v := range ref {
		if got, ok := tab.get(k); !ok || got != v {
			t.Fatalf("lost %v", k)
		}
	}
}
