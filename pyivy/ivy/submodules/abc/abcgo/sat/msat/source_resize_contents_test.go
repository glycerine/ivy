package msat

import "testing"

// The C resize reallocates the watch pointer array, creating new watch
// vectors only for grown slots; its adjacency vector is never truncated.
func TestSourceResizePreservesWatchesAndAdjacency(t *testing.T) {
	p := New(4)
	Msat_SolverClean(p, 2)
	retainedClause, removedClause := &Msat_Clause_t{Num: 11}, &Msat_Clause_t{Num: 12}
	retainedWatch, removedWatch := p.pvWatched[1], p.pvWatched[4]
	Msat_ClauseVecPush(retainedWatch, retainedClause)
	Msat_ClauseVecPush(removedWatch, removedClause)
	retainedAdj, removedAdj := p.vAdjacents[1], p.vAdjacents[2]
	Msat_IntVecPush(retainedAdj, 101)
	Msat_IntVecPush(removedAdj, 202)

	Msat_SolverResize(p, 2)
	if p.pvWatched[1] != retainedWatch || len(retainedWatch.data) != 1 || retainedWatch.data[0] != retainedClause {
		t.Fatal("shrinking must preserve surviving watch vectors and their contents")
	}
	if len(p.vAdjacents) != 4 || p.vAdjacents[1] != retainedAdj || p.vAdjacents[2] != removedAdj ||
		len(retainedAdj.data) != 1 || retainedAdj.data[0] != 101 || len(removedAdj.data) != 1 || removedAdj.data[0] != 202 {
		t.Fatal("shrinking must retain adjacency vectors and their contents")
	}

	Msat_SolverResize(p, 3)
	if p.pvWatched[1] != retainedWatch || len(retainedWatch.data) != 1 || retainedWatch.data[0] != retainedClause {
		t.Fatal("regrowing must preserve surviving watch vectors and their contents")
	}
	for _, i := range []int{4, 5} {
		if p.pvWatched[i] == nil || p.pvWatched[i] == removedWatch || len(p.pvWatched[i].data) != 0 {
			t.Fatalf("regrown watch %d must be a fresh empty vector", i)
		}
	}
	if p.vAdjacents[1] != retainedAdj || p.vAdjacents[2] != removedAdj ||
		len(retainedAdj.data) != 1 || retainedAdj.data[0] != 101 || len(removedAdj.data) != 1 || removedAdj.data[0] != 202 {
		t.Fatal("regrowing must reuse the retained adjacency vectors and their contents")
	}
}
