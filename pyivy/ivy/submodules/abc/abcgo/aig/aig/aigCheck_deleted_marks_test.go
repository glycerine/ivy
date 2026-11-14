package aig

import "testing"

func TestMarkCheckIgnoresDeletedObjectsButRejectsLiveMarks(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	dead := g.AND(a, b)
	g.node(dead).MarkA = true
	if removed := Aig_ManCleanup(g); removed != 1 || g.ValidEdge(dead) {
		t.Fatal("expected dangling node to be deleted")
	}
	// C's Aig_ManForEachObj visits live objects, not recycled slots.
	Aig_ManCheckMarkA(g)
	g.node(a).MarkA = true
	defer func() {
		if recover() == nil {
			t.Fatal("live traversal mark was accepted")
		}
	}()
	Aig_ManCheckMarkA(g)
}
