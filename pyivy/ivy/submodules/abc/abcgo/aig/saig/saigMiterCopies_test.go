package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestMiterCopyMappingsPreservePhase(t *testing.T) {
	makeGraph := func(invert bool) (*aig.Graph, aig.Edge) {
		g := aig.NewGraph()
		pi, lo := g.NewInput(), g.NewInput()
		node := g.AND(pi, lo)
		if invert {
			node = node.Not()
		}
		g.AddOutput(node)
		g.AddOutput(pi)
		aig.Aig_ManSetRegNum(g, 1)
		return g, node
	}
	a, edgeA := makeGraph(false)
	b, edgeB := makeGraph(true)
	m, ca, cb := Saig_ManCreateMiterWithCopies(a, b, 0)
	if len(ca) != len(a.Nodes()) || len(cb) != len(b.Nodes()) {
		t.Fatal("mapping dimensions")
	}
	for _, tc := range []struct {
		source *aig.Graph
		edge   aig.Edge
		copies []aig.Edge
	}{{a, edgeA, ca}, {b, edgeB, cb}} {
		id := tc.edge.NodeID()
		mapped := tc.copies[id]
		if mapped == aig.InvalidEdge {
			t.Fatal("node not mapped")
		}
		for _, inputs := range [][]bool{{false, false, false}, {true, false, false}, {true, true, false}, {true, false, true}} {
			values, err := m.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			_ = values
		}
	}
	outA := aig.Aig_ObjChild0(a, aig.Aig_ManCo(a, 0))
	outB := aig.Aig_ObjChild0(b, aig.Aig_ManCo(b, 0))
	if ca[aig.Aig_ManCo(a, 0).NodeID()] != ca[outA.NodeID()] || cb[aig.Aig_ManCo(b, 0).NodeID()] != cb[outB.NodeID()].Not() {
		t.Fatal("output copy phase lost")
	}
}
