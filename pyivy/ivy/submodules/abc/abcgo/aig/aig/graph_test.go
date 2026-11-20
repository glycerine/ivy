package aig

import "testing"

func TestGraphBooleanOperations(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AddOutput(g.OR(a, b))
	g.AddOutput(g.XOR(a, b))
	g.AddOutput(g.Mux(c, a, b))
	for mask := 0; mask < 8; mask++ {
		x, y, z := mask&1 != 0, mask&2 != 0, mask&4 != 0
		got, err := g.Evaluate([]bool{x, y, z})
		if err != nil {
			t.Fatal(err)
		}
		mux := y
		if z {
			mux = x
		}
		want := []bool{x && y, x || y, x != y, mux}
		for i, v := range want {
			if got[i] != v {
				t.Fatalf("assignment %d output %d: got %v want %v", mask, i, got[i], v)
			}
		}
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestGraphStrashingAndSimplification(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	n := g.AND(a, b)
	if g.AND(b, a) != n || g.AND(a, a) != a || g.AND(a, a.Not()) != g.Const0() || g.AND(a, g.Const1()) != a || g.AND(a, g.Const0()) != g.Const0() {
		t.Fatal("incorrect simplification")
	}
	if g.NodeCount() != 4 {
		t.Fatal("structural hashing did not share node")
	}
	g.AddOutput(n)
	g.AddOutput(n.Not())
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	if _, err := g.Evaluate(nil); err == nil {
		t.Fatal("accepted incorrect arity")
	}
	nodes := g.Nodes()
	nodes[0].Kind = And
	if err := g.Check(); err != nil {
		t.Fatal("Nodes exposed mutable storage", err)
	}
}
func TestGraphCheckDetectsCorruption(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	e := g.AND(a, b)
	g.nodes[e.NodeID()].Level++
	if g.Check() == nil {
		t.Fatal("accepted invalid level")
	}
}

func TestNodeSnapshotAndMissingEdges(t *testing.T) {
	p := NewGraph()
	a := p.NewInput()
	snapshot := p.Node(a)
	snapshot.Kind = And
	if p.Node(a).Kind != Input || p.ValidEdge(InvalidEdge) || Aig_Regular(InvalidEdge) != InvalidEdge {
		t.Fatal("invalid object/snapshot contract")
	}
	if Aig_ObjCopy(p, a) != InvalidEdge {
		t.Fatal("missing copy aliases constant one")
	}
	Aig_ObjSetCopy(p, a, p.Const1())
	if Aig_ObjCopy(p, a) != p.Const1() {
		t.Fatal("constant one copy lost")
	}
}

func TestTopologicalObjectsBeforeAndAfterReplacingLogic(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	last := root
	for i := 0; i < 128; i++ {
		last = g.AND(last.Not(), g.NewInput())
	}
	g.AddOutput(last)
	check := func() {
		want := g.dfs(g.LiveEdges(), false, false, false)
		got := g.TopologicalObjects()
		if len(got) != len(want) {
			t.Fatalf("got %d objects, want %d", len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("object %d got %v, want %v", i, got[i], want[i])
			}
		}
	}
	check()
	if allocs := testing.AllocsPerRun(25, func() { g.TopologicalObjects() }); allocs > 1 {
		t.Errorf("topologically numbered graph allocated %.0f objects for one returned slice", allocs)
	}
	newInput := g.NewInput()
	newFanin := g.AND(b.Not(), newInput)
	Aig_ObjReplace(g, root, g.AND(a, newFanin), 0)
	check()
}
