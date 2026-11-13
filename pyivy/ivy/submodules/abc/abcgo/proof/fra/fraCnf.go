package fra

// Incremental CNF generation from ABC's fraCnf.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

func fraLit(variable int32, negative bool) bsat.Literal {
	if negative {
		return bsat.Negative(variable)
	}
	return bsat.Positive(variable)
}
func fraClause(p *Fra_Man_t, lits ...bsat.Literal) {
	if !p.pSat.AddClause(lits...) {
		panic("fra: generated inconsistent structural CNF")
	}
}

func Fra_AddClausesMux(p *Fra_Man_t, node aig.Edge) {
	control, yes, no := aig.Aig_ObjRecognizeMux(p.pManFraig, node)
	f, i := Fra_ObjSatNum(p, node), Fra_ObjSatNum(p, control)
	t, e := Fra_ObjSatNum(p, yes), Fra_ObjSatNum(p, no)
	ct, ce := yes.Complemented(), no.Complemented()
	fraClause(p, fraLit(i, true), fraLit(t, !ct), fraLit(f, false))
	fraClause(p, fraLit(i, true), fraLit(t, ct), fraLit(f, true))
	fraClause(p, fraLit(i, false), fraLit(e, !ce), fraLit(f, false))
	fraClause(p, fraLit(i, false), fraLit(e, ce), fraLit(f, true))
	if t == e {
		return
	}
	fraClause(p, fraLit(t, ct), fraLit(e, ce), fraLit(f, true))
	fraClause(p, fraLit(t, !ct), fraLit(e, !ce), fraLit(f, false))
}

func Fra_AddClausesSuper(p *Fra_Man_t, node aig.Edge, fanins []aig.Edge) {
	f := Fra_ObjSatNum(p, node)
	long := make([]bsat.Literal, 0, len(fanins)+1)
	for _, fanin := range fanins {
		v := Fra_ObjSatNum(p, fanin)
		fraClause(p, fraLit(v, fanin.Complemented()), fraLit(f, true))
		long = append(long, fraLit(v, !fanin.Complemented()))
	}
	fraClause(p, append(long, fraLit(f, false))...)
}

func Fra_CollectSuper(graph *aig.Graph, root aig.Edge, useMuxes bool) []aig.Edge {
	if root.Complemented() || graph.Node(root).Kind != aig.And {
		panic("fra: invalid supergate root")
	}
	var result []aig.Edge
	seen := make(map[aig.Edge]bool)
	var collect func(aig.Edge, bool)
	collect = func(edge aig.Edge, first bool) {
		node := graph.Node(edge)
		if edge.Complemented() || node.Kind != aig.And || (!first && node.Refs > 1) || (useMuxes && aig.Aig_ObjIsMuxType(graph, edge)) {
			if !seen[edge] {
				seen[edge] = true
				result = append(result, edge)
			}
			return
		}
		collect(node.Fanins[0], false)
		collect(node.Fanins[1], false)
	}
	collect(root, true)
	return result
}

func Fra_ObjAddToFrontier(p *Fra_Man_t, node aig.Edge, frontier *[]aig.Edge) {
	if node.Complemented() {
		panic("fra: complemented frontier node")
	}
	// An existing CNF can assign variable zero to a real input. Its non-nil
	// fanin sentinel marks the node encoded even when the number is zero.
	if Fra_ObjSatNum(p, node) != 0 || Fra_ObjFaninVec(p, node) != nil || node.NodeID() == 0 {
		return
	}
	Fra_ObjSetSatNum(p, node, p.nSatVars)
	p.nSatVars++
	kind := p.pManFraig.Node(node).Kind
	if kind == aig.And || kind == aig.ExclusiveOr {
		*frontier = append(*frontier, node)
	}
}

func Fra_CnfNodeAddToSolver(p *Fra_Man_t, old, new aig.Edge) {
	if old == aig.InvalidEdge && new == aig.InvalidEdge {
		panic("fra: no CNF roots")
	}
	ready := func(e aig.Edge) bool { return e == aig.InvalidEdge || Fra_ObjFaninVec(p, e) != nil }
	if ready(old) && ready(new) {
		return
	}
	var frontier []aig.Edge
	if old != aig.InvalidEdge {
		Fra_ObjAddToFrontier(p, old, &frontier)
	}
	if new != aig.InvalidEdge {
		Fra_ObjAddToFrontier(p, new, &frontier)
	}
	graph := p.pManFraig
	for index := 0; index < len(frontier); index++ {
		node := frontier[index]
		n := graph.Node(node)
		var fanins []aig.Edge
		switch {
		case n.Kind == aig.ExclusiveOr:
			fanins = append(fanins, n.Fanins[:]...)
		case aig.Aig_ObjIsMuxType(graph, node):
			a, b := graph.Node(n.Fanins[0]), graph.Node(n.Fanins[1])
			seen := make(map[aig.Edge]bool)
			for _, edge := range append(a.Fanins[:], b.Fanins[:]...) {
				edge = aig.Aig_Regular(edge)
				if !seen[edge] {
					seen[edge] = true
					fanins = append(fanins, edge)
				}
			}
		default:
			fanins = Fra_CollectSuper(graph, node, true)
		}
		for _, edge := range fanins {
			Fra_ObjAddToFrontier(p, aig.Aig_Regular(edge), &frontier)
		}
		switch {
		case n.Kind == aig.ExclusiveOr:
			a, b := n.Fanins[0], n.Fanins[1]
			av, bv, f := Fra_ObjSatNum(p, a), Fra_ObjSatNum(p, b), Fra_ObjSatNum(p, node)
			for x := 0; x < 2; x++ {
				for y := 0; y < 2; y++ {
					// Exclude each assignment where the result is not x XOR y.
					wrong := (x ^ y) == 0
					fraClause(p, fraLit(av, (x != 0) != a.Complemented()), fraLit(bv, (y != 0) != b.Complemented()), fraLit(f, wrong))
				}
			}
		case aig.Aig_ObjIsMuxType(graph, node):
			Fra_AddClausesMux(p, node)
		default:
			Fra_AddClausesSuper(p, node, fanins)
		}
		Fra_ObjSetFaninVec(p, node, fanins)
	}
}
