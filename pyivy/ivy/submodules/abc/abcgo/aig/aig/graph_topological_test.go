package aig

import (
	"reflect"
	"testing"
)

func TestTopologicalObjectsIncludesDanglingRewrittenLogic(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	early := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{a, InvalidEdge}})
	later := g.AND(b, c)
	Aig_ObjConnect(g, early, later, InvalidEdge)
	g.AddOutput(early.Not())
	dead := g.AND(a, b)
	Aig_ManCleanup(g)
	dangling := g.AND(a, c)
	if g.valid(dead) || !g.valid(dangling) {
		t.Fatal("fixture lacks a deletion hole or dangling gate")
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	g.travID = 19
	g.node(early).TravID = 8
	g.node(a).MarkA = true
	g.node(later).MarkB = true
	before := g.Nodes()
	order := g.TopologicalObjects()
	if len(order) != g.NodeCount() {
		t.Fatalf("returned %d objects, want %d", len(order), g.NodeCount())
	}
	positions := make(map[Edge]int, len(order))
	for i, e := range order {
		if !g.valid(e) || e.Complemented() {
			t.Fatalf("invalid object %d in traversal", e)
		}
		if _, exists := positions[e]; exists {
			t.Fatalf("duplicate object %d", e)
		}
		positions[e] = i
		for _, fanin := range nodeFanins(*g.node(e)) {
			if position, seen := positions[Aig_Regular(fanin)]; !seen || position >= i {
				t.Fatalf("object %d precedes fanin %d", e, fanin)
			}
		}
	}
	if _, ok := positions[dangling]; !ok {
		t.Fatal("dangling gate omitted")
	}
	if !reflect.DeepEqual(before, g.Nodes()) || g.travID != 19 {
		t.Fatal("traversal changed graph or scratch marks")
	}
}

func TestTopologicalObjectsRejectsCycles(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{a, InvalidEdge}})
	// Deliberately corrupt the graph without invoking the mutator's checks.
	g.node(buffer).Fanins[0] = buffer
	defer func() {
		if recover() == nil {
			t.Fatal("accepted cyclic graph")
		}
	}()
	g.TopologicalObjects()
}
