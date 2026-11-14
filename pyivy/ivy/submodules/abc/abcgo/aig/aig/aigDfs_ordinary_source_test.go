package aig

import (
	"slices"
	"testing"
)

func TestOrdinaryDfsSourceOrderAndMarks(t *testing.T) {
	for _, preorder := range []bool{false, true} {
		for _, nodesOnly := range []int32{0, 1} {
			p := NewGraph()
			a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
			ab := p.AND(a, b)
			root := p.AND(ab, c)
			bc := p.AND(b, c)
			alternative := p.AND(a, bc)
			out := p.AddOutput(root.Not())
			p.equivs = map[int]Edge{root.NodeID(): alternative}
			want := []Edge{p.Const1(), a, b, c, bc, alternative, ab, root, out}
			if nodesOnly != 0 {
				want = []Edge{bc, alternative, ab, root}
			}
			if preorder {
				// C uses POSTORDER for the alternative, even during preorder DFS.
				want = []Edge{p.Const1(), out, root, a, b, c, bc, alternative, ab}
				if nodesOnly != 0 {
					want = []Edge{root, bc, alternative, ab}
				}
			}
			old := p.travID
			var got []Edge
			if preorder {
				got = Aig_ManDfsPreorder(p, nodesOnly)
			} else {
				got = Aig_ManDfs(p, nodesOnly)
			}
			if !slices.Equal(got, want) {
				t.Errorf("preorder=%v nodesOnly=%d: got %v, want %v", preorder, nodesOnly, got, want)
			}
			if p.travID != old+1 {
				t.Errorf("preorder=%v nodesOnly=%d: no new traversal", preorder, nodesOnly)
			}
			for _, e := range append(want, p.Const1(), a, b, c) {
				if p.node(e).TravID != p.travID {
					t.Errorf("object %d not marked", e.NodeID())
				}
			}
			if nodesOnly != 0 && p.node(out).TravID == p.travID {
				t.Error("node-only traversal marked a CO")
			}
		}
	}
}

func TestOrdinaryDfsRecRespectsSeedsAndCollectsTerminals(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	root := p.AND(a, b)
	Aig_ManIncrementTravId(p)
	p.node(p.Const1()).TravID = p.travID
	p.node(a).TravID = p.travID
	var got []Edge
	Aig_ManDfs_rec(p, root, &got)
	Aig_ManDfs_rec(p, root, &got)
	if !slices.Equal(got, []Edge{b, root}) {
		t.Errorf("seeded recursive DFS = %v, want %v", got, []Edge{b, root})
	}
}

func TestOrdinaryDfsRejectsDanglingLogicAsInC(t *testing.T) {
	for _, visit := range []func(*Graph, int32) []Edge{Aig_ManDfs, Aig_ManDfsPreorder} {
		func() {
			p := NewGraph()
			a, b := p.NewInput(), p.NewInput()
			p.AND(a, b)
			p.AddOutput(a)
			defer func() {
				if recover() == nil {
					t.Error("accepted dangling logic despite C's node-count assertion")
				}
			}()
			visit(p, 1)
		}()
	}
}
