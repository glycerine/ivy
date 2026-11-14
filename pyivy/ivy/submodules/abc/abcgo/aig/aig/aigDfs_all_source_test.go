package aig

import (
	"slices"
	"testing"
)

func TestDfsAllSourceOrderAndUnusedInputMarks(t *testing.T) {
	p := NewGraph()
	a, b, unused := p.NewInput(), p.NewInput(), p.NewInput()
	root := p.AND(a, b)
	out := p.AddOutput(root.Not())
	dead := p.AND(a, unused)
	Aig_ManCleanup(p)
	if p.valid(dead) {
		t.Fatal("fixture did not delete dangling gate")
	}
	old := p.travID
	want := []Edge{p.Const1(), a, b, root, out, unused}
	if got := Aig_ManDfsAll(p); !slices.Equal(got, want) {
		t.Errorf("DFS all = %v, want %v", got, want)
	}
	if p.travID != old+1 {
		t.Error("DFS all did not start a traversal")
	}
	for _, e := range want[:len(want)-1] {
		if p.node(e).TravID != p.travID {
			t.Errorf("reachable object %d not marked", e.NodeID())
		}
	}
	// C appends unused CIs after the output-cone walk without marking them.
	if p.node(unused).TravID == p.travID {
		t.Error("unused CI was marked unlike C")
	}
}

func TestDfsAllRejectsDanglingGatesAsInC(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AND(a, b)
	p.AddOutput(a)
	defer func() {
		if recover() == nil {
			t.Fatal("accepted dangling gate despite C's object-count assertion")
		}
	}()
	Aig_ManDfsAll(p)
}
