package aig

import "testing"

func TestDominatorDriverSourceTraversal(t *testing.T) {
	for _, mode := range []string{"nodes", "flops", "pis"} {
		t.Run(mode, func(t *testing.T) {
			g := NewGraph()
			a, b, lo, unused := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
			po, li := g.AND(a, lo), g.AND(b, lo)
			dangling := g.AND(a, b)
			g.AddOutput(dangling)
			co, latch := g.AddOutput(po), g.AddOutput(li)
			g.nRegs = 1
			g.travID = 7
			for i := range g.nodes {
				g.nodes[i].TravID = 5
			}
			Aig_ManFanoutStart(g)
			s := domCompute(g, 2, mode)
			wantGen, wantNodes := int32(9), 1
			if mode == "nodes" {
				wantGen, wantNodes = 8, 3
			}
			if g.travID != wantGen || s.DomNodes != wantNodes {
				t.Errorf("generation/nodes = %d/%d, want %d/%d", g.travID, s.DomNodes, wantGen, wantNodes)
			}
			if g.node(unused).TravID != 5 {
				t.Error("unused CI was visited")
			}
			if mode == "nodes" {
				for _, e := range []Edge{po, li, dangling, co, latch} {
					if g.node(e).TravID != 8 {
						t.Errorf("object %d not marked by reverse DFS", e.NodeID())
					}
				}
				if g.node(a).TravID != 5 {
					t.Error("node driver visited CI")
				}
			} else {
				for _, e := range []Edge{b, lo, li, latch} {
					if g.node(e).TravID != 9 {
						t.Errorf("object %d missing LI-cone mark", e.NodeID())
					}
				}
				if g.node(a).TravID != 5 || g.node(dangling).TravID != 8 || g.node(co).TravID != 8 {
					t.Error("non-LI state differs from reverse DFS followed by LI marking")
				}
			}
		})
	}
}

func TestDominatorDirectDriverRequiresSourceFanouts(t *testing.T) {
	g := NewGraph()
	g.AddOutput(g.AND(g.NewInput(), g.NewInput()))
	defer func() {
		if recover() == nil {
			t.Error("driver accepted missing fanout index")
		}
	}()
	Aig_ManComputeDomsNodes(g, 2)
}

func TestDominatorRejectsMissingFanoutDominators(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	live := g.AND(a, b)
	g.AND(a, c) // No dominators can be initialized for this dangling fanout.
	g.AddOutput(live)
	Aig_ManFanoutStart(g)
	defer func() {
		if recover() == nil {
			t.Error("silently merged a missing source dominator vector")
		}
	}()
	Aig_ManComputeDomsNodes(g, 2)
}

func TestFlopTfiRejectsComplementedRoot(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	Aig_ManIncrementTravId(g)
	defer func() {
		if recover() == nil {
			t.Error("accepted complemented root forbidden by C")
		}
	}()
	Aig_ManMarkFlopTfi_rec(g, a.Not())
}

func TestFlopTfiUsesBoundedTraversalGeneration(t *testing.T) {
	g := NewGraph()
	a, lo := g.NewInput(), g.NewInput()
	li := g.AddOutput(g.AND(a.Not(), lo))
	g.nRegs, g.travID = 1, (1<<30)-1
	Aig_ManMarkFlopTfi(g)
	if g.travID != 2 || !Aig_ObjIsTravIdCurrent(g, a) || !Aig_ObjIsTravIdCurrent(g, li) {
		t.Fatalf("flop TFI bypassed bounded traversal helper: generation %d", g.travID)
	}
}

func TestDominatorWrappersOwnSourceFanoutLifetime(t *testing.T) {
	for _, name := range []string{"test", "test2", "cofactoring"} {
		t.Run(name, func(t *testing.T) {
			g := NewGraph()
			a, lo := g.NewInput(), g.NewInput()
			g.AddOutput(g.AND(a, lo))
			g.nRegs = 1
			for i := 0; i < 2; i++ {
				switch name {
				case "test":
					Aig_ManComputeDomsTest(g)
				case "test2":
					Aig_ManComputeDomsTest2(g, 2)
				case "cofactoring":
					Aig_ManComputeDomsForCofactoring(g)
				}
				if g.fanouts != nil {
					t.Fatal("wrapper retained its temporary fanout index")
				}
			}
		})
	}
}
