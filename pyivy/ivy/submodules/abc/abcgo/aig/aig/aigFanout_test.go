package aig

import "testing"

func checkFanoutLinks(t *testing.T, g *Graph, driver Edge, want int) {
	t.Helper()
	fans := g.fanouts.Handles(driver)
	if len(fans) != want {
		t.Fatal(driver, fans, want)
	}
	for _, f := range fans {
		prev, next := *Aig_FanoutPrev(g.fanouts.data, f), *Aig_FanoutNext(g.fanouts.data, f)
		if *Aig_FanoutNext(g.fanouts.data, prev) != f || *Aig_FanoutPrev(g.fanouts.data, next) != f {
			t.Fatal("broken ring", f)
		}
	}
}
func TestFanoutCircularMutation(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	x, y := g.AND(a, b), g.AND(a, c)
	Aig_ManFanoutStart(g)
	checkFanoutLinks(t, g, a, 2)
	checkFanoutLinks(t, g, b, 1)
	Aig_ObjRemoveFanout(g, a, x)
	checkFanoutLinks(t, g, a, 1)
	Aig_ObjRemoveFanout(g, a, y)
	checkFanoutLinks(t, g, a, 0)
	Aig_ObjAddFanout(g, a, y)
	Aig_ObjAddFanout(g, a, x)
	checkFanoutLinks(t, g, a, 2)
	Aig_ManFanoutStop(g)
	if g.fanouts != nil {
		t.Fatal("stop")
	}
}
func TestFanoutRebuildDeletedUnaryAndDuplicatePins(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	g.nodes = append(g.nodes, Node{Kind: ExclusiveOr, Fanins: [2]Edge{a, a.Not()}})
	x := Edge(len(g.nodes)-1) << 1
	g.nodes = append(g.nodes, Node{Kind: Output, Fanins: [2]Edge{x, InvalidEdge}})
	Aig_ManFanoutStart(g)
	checkFanoutLinks(t, g, a, 2)
	checkFanoutLinks(t, g, x, 1)
	Aig_ObjRemoveFanoutPin(g, a, x, 1)
	checkFanoutLinks(t, g, a, 1)
	Aig_ObjAddFanoutPin(g, a, x, 1)
	checkFanoutLinks(t, g, a, 2)
	g.deleted[x.NodeID()] = true
	g.nodes = g.nodes[:len(g.nodes)-1]
	Aig_ManFanoutRebuild(g)
	checkFanoutLinks(t, g, a, 0)
}
func TestFanoutCapacityAndOrder(t *testing.T) {
	f := new(FanoutIndex)
	f.add(0, 5000, 0)
	f.add(0, 5001, 1)
	fans := f.Handles(0)
	if len(fans) != 2 || fans[0] != 10000 || fans[1] != 10003 {
		t.Fatal(fans)
	}
	f.remove(0, 5000, 0)
	if got := f.Handles(0); len(got) != 1 || got[0] != 10003 {
		t.Fatal(got)
	}
	f.remove(0, 5001, 1)
	if len(f.Handles(0)) != 0 {
		t.Fatal("last removal")
	}
}
func TestFanoutDuplicateInsertionRejected(t *testing.T) {
	f := new(FanoutIndex)
	f.add(0, 1, 0)
	defer func() {
		if recover() == nil {
			t.Fatal("duplicate insertion accepted")
		}
	}()
	f.add(0, 1, 0)
}
