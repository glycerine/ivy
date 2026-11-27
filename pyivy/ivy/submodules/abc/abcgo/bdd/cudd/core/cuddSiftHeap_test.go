package core

import (
	"math/rand/v2"
	"testing"
)

func siftEval(s *siftHeap, e siftEdge, assign []bool) bool {
	c := siftCompl(e)
	id := siftReg(e)
	for id != 0 {
		nd := s.nodes[id]
		next := nd.e
		if assign[nd.index] {
			next = nd.t
		}
		c = c != siftCompl(next)
		id = siftReg(next)
	}
	return !c
}

// TestSiftHeapSwapsKeepFunctionsAndCanonicalSize: after random adjacent
// swaps (cuddSwapInPlace), the heap still computes the same functions, and
// the nodes reachable from the roots are exactly as many as a canonical
// BDD built for the new order has.
func TestSiftHeapSwapsKeepFunctionsAndCanonicalSize(t *testing.T) {
	r := rand.New(rand.NewPCG(5, 6))
	for trial := 0; trial < 40; trial++ {
		nv := 4 + r.IntN(6)
		dd := NewBDDManager(nv)
		var roots []*DdNode
		for k := 0; k < 3; k++ {
			f := dd.LogicZero
			for c := 0; c < 2+r.IntN(4); c++ {
				cube := dd.One
				for v := 0; v < nv; v++ {
					switch r.IntN(3) {
					case 0:
						cube = Cudd_bddAnd(dd, cube, dd.Vars[v])
					case 1:
						cube = Cudd_bddAnd(dd, cube, Cudd_Not(dd.Vars[v]))
					}
				}
				f = Cudd_bddOr(dd, f, cube)
			}
			if r.IntN(2) == 0 {
				f = Cudd_bddXor(dd, f, dd.Vars[r.IntN(nv)])
			}
			Cudd_Ref(f)
			roots = append(roots, f)
		}
		live := dd.liveReorderRoots()
		heap := reorderBuild(dd, live, append([]int(nil), dd.InvPerm...))
		s := newSiftHeap(heap)
		if s == nil {
			t.Fatal("no sift heap")
		}
		assigns := make([][]bool, 64)
		want := make([][]bool, len(s.roots))
		for a := range assigns {
			assigns[a] = make([]bool, nv)
			for v := range assigns[a] {
				assigns[a][v] = r.IntN(2) == 1
			}
		}
		for i, root := range s.roots {
			for _, as := range assigns {
				want[i] = append(want[i], siftEval(s, root, as))
			}
		}
		for step := 0; step < 30; step++ {
			s.swap(r.IntN(nv - 1))
			for i, root := range s.roots {
				for a, as := range assigns {
					if siftEval(s, root, as) != want[i][a] {
						t.Fatalf("trial %d step %d: function %d changed", trial, step, i)
					}
				}
			}
			canonical := reorderBuild(dd, live, append([]int(nil), s.invperm...))
			nonConst := 0
			for node := range canonical.nodes {
				if !Cudd_IsConstant(node) {
					nonConst++
				}
			}
			// the canonical heap also holds every projection variable
			reach := s.liveCount(append(append([]siftEdge(nil), s.roots...), siftVarEdges(s)...))
			if reach != nonConst {
				t.Fatalf("trial %d step %d: %d reachable nodes, canonical %d", trial, step, reach, nonConst)
			}
		}
	}
}

func siftVarEdges(s *siftHeap) []siftEdge {
	var e []siftEdge
	for _, id := range s.varNode {
		e = append(e, siftMk(id, false))
	}
	return e
}
