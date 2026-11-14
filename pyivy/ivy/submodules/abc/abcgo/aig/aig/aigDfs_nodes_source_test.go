package aig

import (
	"slices"
	"testing"
)

func TestDfsNodesIncludesInputsAndSourceMarks(t *testing.T) {
	p := NewGraph()
	a, b, unused := p.NewInput(), p.NewInput(), p.NewInput()
	root := p.AND(a, b)
	out := p.AddOutput(root.Not())
	old := p.travID
	got := Aig_ManDfsNodes(p, []Edge{out, a, out, unused}, 3)
	want := []Edge{a, b, root}
	if !slices.Equal(got, want) {
		t.Errorf("DFS objects = %v, want %v", got, want)
	}
	if p.travID != old+1 {
		t.Error("DFS did not start a traversal")
	}
	for _, e := range []Edge{p.Const1(), a, b, root} {
		if p.node(e).TravID != p.travID {
			t.Errorf("object %d lacks current traversal mark", e.NodeID())
		}
	}
	for _, e := range []Edge{out, unused} {
		if p.node(e).TravID == p.travID {
			t.Errorf("object %d incorrectly marked", e.NodeID())
		}
	}
	if got := Aig_ManDfsNodes(p, []Edge{p.Const1()}, 1); len(got) != 0 {
		t.Fatalf("pre-marked constant was returned: %v", got)
	}
	if p.node(a).TravID == p.travID {
		t.Fatal("new traversal retained an old input mark")
	}
}

func TestDfsNodesVisitsChoiceBeforeFanins(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	ab := p.AND(a, b)
	root := p.AND(ab, c)
	bc := p.AND(b, c)
	alternative := p.AND(a, bc)
	p.equivs = map[int]Edge{root.NodeID(): alternative}
	want := []Edge{a, b, c, bc, alternative, ab, root}
	if got := Aig_ManDfsNodes(p, []Edge{root}, 1); !slices.Equal(got, want) {
		t.Fatalf("choice DFS = %v, want %v", got, want)
	}
}
