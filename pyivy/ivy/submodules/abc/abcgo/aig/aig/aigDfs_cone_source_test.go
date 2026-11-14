package aig

import "testing"

func coneSourceFixture() (*Graph, Edge, Edge, Edge, Edge) {
	g := NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	lower := g.AND(a, b)
	barrier := g.AND(lower, c)
	root := g.AND(barrier, d)
	return g, a, lower, barrier, root
}

func TestConeMarkSourceStopsAtExistingMarks(t *testing.T) {
	for _, clean := range []bool{false, true} {
		g, _, _, barrier, root := coneSourceFixture()
		g.node(barrier).MarkA = true
		g.copies = make(map[int]Edge)
		for _, e := range g.LiveEdges() {
			g.SetObjectData(e, "payload")
			g.copies[e.NodeID()] = e
		}
		if clean {
			Aig_ConeCleanAndMark_rec(g, root)
		} else {
			Aig_ConeMark_rec(g, root)
		}
		for _, e := range g.LiveEdges() {
			wantMark := e == root || e == barrier
			if g.node(e).MarkA != wantMark {
				t.Errorf("clean=%v: object %d mark = %v, want %v", clean, e, g.node(e).MarkA, wantMark)
			}
			_, hasCopy := g.copies[e.NodeID()]
			wantData := !(clean && e == root)
			if hasCopy != wantData || (g.ObjectData(e) != nil) != wantData {
				t.Errorf("clean=%v: object %d data/copy clearing crossed the marked barrier", clean, e)
			}
		}
	}
}

func TestConeCountSourceStopsAtExistingMarks(t *testing.T) {
	g, _, lower, barrier, root := coneSourceFixture()
	g.node(barrier).MarkA = true
	if got := Aig_ConeCountAndMark_rec(g, root); got != 1 {
		t.Errorf("new cone nodes = %d, want 1", got)
	}
	if g.node(lower).MarkA || !g.node(root).MarkA {
		t.Error("count traversal crossed existing mark or failed to mark root")
	}
}

func TestConeUnmarkSourceStopsAtUnmarkedNodes(t *testing.T) {
	g, a, lower, _, root := coneSourceFixture()
	g.node(a).MarkA = true
	g.node(lower).MarkA = true
	g.node(root).MarkA = true
	Aig_ConeUnmark_rec(g, root)
	if g.node(root).MarkA || !g.node(lower).MarkA || !g.node(a).MarkA {
		t.Fatal("unmarking crossed the unmarked barrier or modified a terminal")
	}
}

func TestDagSizeSourceConsumesConeMarks(t *testing.T) {
	g, a, lower, barrier, root := coneSourceFixture()
	g.node(barrier).MarkA = true
	g.node(a).MarkA = true
	if got := Aig_DagSize(g, root.Not()); got != 1 {
		t.Errorf("DAG size with marked barrier = %d, want 1", got)
	}
	for _, e := range []Edge{lower, barrier, root} {
		if g.node(e).MarkA {
			t.Errorf("DAG size left gate %d marked", e)
		}
	}
	if !g.node(a).MarkA {
		t.Error("DAG size cleared a terminal mark")
	}
}

func TestConeSourceStopsAtBufferAndOutput(t *testing.T) {
	g, _, _, _, root := coneSourceFixture()
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{root, InvalidEdge}})
	out := g.AddOutput(root)
	for _, e := range []Edge{buffer, out} {
		if got := Aig_DagSize(g, e); got != 0 {
			t.Errorf("non-node cone size = %d, want 0", got)
		}
		Aig_ConeMark_rec(g, e)
		if g.node(e).MarkA || g.node(root).MarkA {
			t.Error("cone marking followed a buffer or output")
		}
	}
}

func TestConeSourceCountsSharedAndXorNodesOnce(t *testing.T) {
	g := NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	shared := g.AND(a, b)
	left := g.AND(shared, c)
	right := g.NewXor(shared, d)
	root := g.AND(left, right)
	if got := Aig_ConeCountAndMark_rec(g, root); got != 4 {
		t.Fatalf("shared cone count = %d, want 4", got)
	}
	if got := Aig_ConeCountAndMark_rec(g, root); got != 0 {
		t.Fatalf("marked cone count = %d, want 0", got)
	}
	Aig_ConeUnmark_rec(g, root)
	if got := Aig_DagSize(g, root); got != 4 {
		t.Fatalf("DAG size = %d, want 4", got)
	}
	for _, e := range g.LiveEdges() {
		if g.node(e).MarkA {
			t.Fatalf("DAG size left object %d marked", e)
		}
	}
}
