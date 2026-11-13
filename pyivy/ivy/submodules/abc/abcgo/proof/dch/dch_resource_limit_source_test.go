package dch

import (
	"testing"

	aig "github.com/glycerine/abcgo/aig/aig"
)

// dchSweep.c marks an unresolved SAT candidate unmapped and skips every
// subsequent node with an unmapped fanin. They cannot become proved choices.
func TestConflictLimitBlocksCandidateFanouts(t *testing.T) {
	left, right := aig.NewGraph(), aig.NewGraph()
	a, b := make([]aig.Edge, 8), make([]aig.Edge, 8)
	for i := range a {
		a[i], b[i] = left.NewInput(), right.NewInput()
	}
	x := a[0]
	for _, n := range a[1:] {
		x = left.XOR(x, n)
	}
	left.AddOutput(x)
	var parity func([]aig.Edge) aig.Edge
	parity = func(v []aig.Edge) aig.Edge {
		if len(v) == 1 {
			return v[0]
		}
		mid := len(v) / 2
		return right.XOR(parity(v[:mid]), parity(v[mid:]))
	}
	right.AddOutput(parity(b))
	g := Dch_DeriveTotalAig([]*aig.Graph{left, right})
	nodes := append([]aig.Edge(nil), g.LiveEdges()...)
	extra := g.NewInput()
	for _, node := range nodes {
		if g.Node(node).Kind == aig.And {
			g.AddOutput(g.AND(node, extra))
		}
	}
	pars := DefaultParams()
	pars.ConflictLimit = 1
	manager, err := newManager(g, &pars)
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.computeEquivalences(); err != nil {
		t.Fatal(err)
	}
	if manager.stats.Unknown == 0 || manager.stats.Blocked == 0 {
		t.Fatalf("expected an unresolved proof and blocked fanouts: %+v", manager.stats)
	}
	blockedFanouts := 0
	for _, node := range nodes {
		if g.Node(node).Kind != aig.And {
			continue
		}
		if !manager.blocked[node.NodeID()] {
			continue
		}
		for _, out := range manager.fanouts[node.NodeID()] {
			child := g.Node(aig.Edge(out << 1))
			if child.Kind != aig.And {
				continue
			}
			blockedFanouts++
			if !manager.blocked[out] || manager.mapped[out] != aig.InvalidEdge {
				t.Fatalf("fanout %d of unresolved node %d was mapped", out, node.NodeID())
			}
			if _, ok := manager.proven[out]; ok {
				t.Fatalf("blocked fanout %d was published as proved", out)
			}
		}
	}
	if blockedFanouts == 0 {
		t.Fatal("fixture did not exercise a blocked AND fanout")
	}
}
