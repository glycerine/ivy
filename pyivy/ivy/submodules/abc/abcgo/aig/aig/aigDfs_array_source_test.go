package aig

import (
	"slices"
	"testing"
)

func TestDfsArraySourceObjectsAndMarks(t *testing.T) {
	p := NewGraph()
	a, b, unused := p.NewInput(), p.NewInput(), p.NewInput()
	root := p.AND(a, b)
	out := p.AddOutput(root.Not())
	old := p.travID
	want := []Edge{p.Const1(), a, b, root, out}
	if got := Aig_ManDfsArray(p, []Edge{out, a, out, unused}, 3); !slices.Equal(got, want) {
		t.Errorf("array DFS = %v, want %v", got, want)
	}
	if p.travID != old+1 {
		t.Error("array DFS did not start a traversal")
	}
	for _, e := range want {
		if p.node(e).TravID != p.travID {
			t.Errorf("returned object %d not marked", e.NodeID())
		}
	}
	if p.node(unused).TravID == p.travID {
		t.Error("root beyond array range was visited")
	}
	if got := Aig_ManDfsArray(p, nil, 0); !slices.Equal(got, []Edge{p.Const1()}) {
		t.Errorf("empty array DFS = %v, want constant", got)
	}
}

func TestDfsAllRecUsesCallerTraversalSeeds(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	root := p.AND(a, b)
	Aig_ManIncrementTravId(p)
	p.node(p.Const1()).TravID = p.travID
	p.node(a).TravID = p.travID
	got := []Edge{p.Const1()}
	Aig_ManDfsAll_rec(p, root, &got)
	Aig_ManDfsAll_rec(p, root, &got)
	want := []Edge{p.Const1(), b, root}
	if !slices.Equal(got, want) {
		t.Errorf("seeded recursive DFS = %v, want %v", got, want)
	}
	if p.node(b).TravID != p.travID || p.node(root).TravID != p.travID {
		t.Error("recursive DFS did not leave source traversal marks")
	}
}
