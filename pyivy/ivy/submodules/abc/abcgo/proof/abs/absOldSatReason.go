package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_RefManFindReason follows controlling values from the unrolled output
// to a sufficient set of frame inputs, using the original priority rule.
func Saig_RefManFindReason(p *Saig_RefMan_t) *vec.Vec_Int_t {
	g := p.pFrames
	Saig_RefManSetPhases(p, nil, 0)
	priority := make([]int, len(g.Nodes()))
	for i := range priority {
		priority[i] = -1
	}
	inputIndex := make(map[int]int, len(g.Inputs()))
	for i, e := range g.Inputs() {
		priority[e.NodeID()] = i
		inputIndex[e.NodeID()] = i
	}
	value := func(e aig.Edge) bool { return p.phase[e.NodeID()] != e.Complemented() }
	for _, e := range g.LiveEdges() {
		n := g.Node(e)
		if n.Kind != aig.And {
			continue
		}
		a, b := n.Fanins[0], n.Fanins[1]
		av, bv := value(a), value(b)
		ap, bp := priority[a.NodeID()], priority[b.NodeID()]
		switch {
		case av && bv:
			priority[e.NodeID()] = max(ap, bp)
		case !av && bv:
			priority[e.NodeID()] = ap
		case av && !bv:
			priority[e.NodeID()] = bp
		default:
			priority[e.NodeID()] = min(ap, bp)
		}
	}
	out := g.Node(g.Outputs()[0]).Fanins[0]
	if value(out) {
		panic("refinement reason requires a false output")
	}
	result := vec.Vec_IntAlloc(100)
	seen := make([]bool, len(priority))
	var visit func(aig.Edge)
	visit = func(e aig.Edge) {
		id := e.NodeID()
		if seen[id] {
			return
		}
		seen[id] = true
		if i, ok := inputIndex[id]; ok {
			vec.Vec_IntPush(result, int32(i))
			return
		}
		n := g.Node(e)
		if n.Kind != aig.And {
			panic("refinement reason expects an AND-inverter graph")
		}
		a, b := n.Fanins[0], n.Fanins[1]
		if p.phase[id] {
			visit(a)
			visit(b)
			return
		}
		av, bv := value(a), value(b)
		if !av && bv {
			visit(a)
		} else if av && !bv {
			visit(b)
		} else if priority[a.NodeID()] <= priority[b.NodeID()] {
			visit(a)
		} else {
			visit(b)
		}
	}
	if out.NodeID() != g.Const1().NodeID() {
		visit(out)
	}
	return result
}
