package cgt

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/satoko"
)

type Cgt_Man_t struct {
	Params                   Cgt_Par_t
	Graph, Care, Frame, Part *aig.Graph
	Useful                   []bool
	AllGates                 Gates
	Stats                    Statistics
	frameCopies, partCopies  map[int]aig.Edge
	solver                   *satoko.Solver
	patterns                 [][]bool
	lastModel                []bool
	careCount                int
	futureInputs             map[int]aig.Edge
	encoded                  map[int]bool
}
type Manager = Cgt_Man_t

func regular(e aig.Edge) aig.Edge { return e &^ 1 }
func driver(g *aig.Graph, e aig.Edge) aig.Edge {
	if g.Node(e).Kind == aig.Output {
		return g.Node(e).Fanins[0] ^ aig.Edge(e&1)
	}
	return e
}
func layout(g *aig.Graph) (regs, pis, pos int) {
	regs = int(aig.Aig_ManRegNum(g))
	return regs, g.InputCount() - regs, g.OutputCount() - regs
}
func fanins(g *aig.Graph, e aig.Edge) []aig.Edge {
	n := g.Node(e)
	switch n.Kind {
	case aig.And, aig.ExclusiveOr:
		return []aig.Edge{n.Fanins[0], n.Fanins[1]}
	case aig.Buffer, aig.Output:
		return []aig.Edge{n.Fanins[0]}
	}
	return nil
}
func fanoutIndex(g *aig.Graph) map[int][]aig.Edge {
	index := make(map[int][]aig.Edge)
	for _, e := range g.LiveEdges() {
		for _, f := range fanins(g, e) {
			index[f.NodeID()] = append(index[f.NodeID()], e)
		}
	}
	return index
}
func cloneEdge(dst, src *aig.Graph, e aig.Edge, copies map[int]aig.Edge, input func(aig.Edge) aig.Edge) aig.Edge {
	if copy, ok := copies[e.NodeID()]; ok {
		return copy ^ aig.Edge(e&1)
	}
	n := src.Node(e)
	var copy aig.Edge
	switch n.Kind {
	case aig.Constant:
		copy = dst.Const1()
	case aig.Input:
		copy = input(regular(e))
	case aig.And:
		copy = dst.AND(cloneEdge(dst, src, n.Fanins[0], copies, input), cloneEdge(dst, src, n.Fanins[1], copies, input))
	case aig.ExclusiveOr:
		copy = dst.NewXor(cloneEdge(dst, src, n.Fanins[0], copies, input), cloneEdge(dst, src, n.Fanins[1], copies, input))
	case aig.Buffer, aig.Output:
		copy = cloneEdge(dst, src, n.Fanins[0], copies, input)
	default:
		panic("cgt: unsupported AIG node")
	}
	copies[e.NodeID()] = copy
	return copy ^ aig.Edge(e&1)
}
