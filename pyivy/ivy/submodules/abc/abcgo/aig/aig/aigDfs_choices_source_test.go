package aig

import (
	"slices"
	"testing"
)

func TestDfsChoicesSourceTraversal(t *testing.T) {
	p := NewGraph()
	a, b, c, unused := p.NewInput(), p.NewInput(), p.NewInput(), p.NewInput()
	ab := p.AND(a, b)
	root := p.AND(ab, c)
	bc := p.AND(b, c)
	alternative := p.AND(a, bc)
	out := p.AddOutput(root)
	p.equivs = map[int]Edge{root.NodeID(): alternative}
	old := p.travID
	want := []Edge{ab, bc, alternative, root}
	if got := Aig_ManDfsChoices(p); !slices.Equal(got, want) {
		t.Errorf("choice DFS = %v, want %v", got, want)
	}
	if p.travID != old+1 {
		t.Error("choice DFS did not start a traversal")
	}
	for _, e := range append(want, p.Const1(), a, b, c, unused) {
		if p.node(e).TravID != p.travID {
			t.Errorf("object %d lacks source traversal mark", e.NodeID())
		}
	}
	if p.node(out).TravID == p.travID {
		t.Error("choice DFS marked the CO")
	}
	// The recursive entry point must respect existing traversal seeds, even
	// when they are absent from the caller's result slice.
	Aig_ManIncrementTravId(p)
	for _, e := range []Edge{p.Const1(), a, b, c, unused, ab} {
		p.node(e).TravID = p.travID
	}
	var got []Edge
	Aig_ManDfsChoices_rec(p, root, &got)
	Aig_ManDfsChoices_rec(p, root, &got)
	if !slices.Equal(got, want[1:]) {
		t.Errorf("seeded choice DFS = %v, want %v", got, want[1:])
	}
}

func TestDfsChoicesRequiresSourceEquivalenceStorage(t *testing.T) {
	p := NewGraph()
	p.AddOutput(p.NewInput())
	defer func() {
		if recover() == nil {
			t.Fatal("choice DFS accepted absent equivalence storage")
		}
	}()
	Aig_ManDfsChoices(p)
}
