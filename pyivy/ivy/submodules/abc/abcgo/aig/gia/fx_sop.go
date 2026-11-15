package gia

import (
	"strings"

	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/bool/kit"
)

// Gia_ManGraphToAig converts a decomposition graph with assigned leaf signals.
func Gia_ManGraphToAig(p *Gia_Man_t, graph *dec.Dec_Graph_t) int32 {
	if dec.Dec_GraphIsConst(graph) != 0 {
		return 1 ^ dec.Dec_GraphIsComplement(graph)
	}
	read := func(index uint32) int32 {
		value, ok := dec.Dec_GraphNodeFunc(graph, int32(index)).(int32)
		if !ok || value < 0 {
			panic("gia: missing decomposition node signal")
		}
		return value
	}
	root := dec.Dec_GraphRootEdge(graph)
	if dec.Dec_GraphIsVar(graph) != 0 {
		return read(root.Node) ^ dec.Dec_GraphIsComplement(graph)
	}
	leaves := dec.Dec_GraphLeaveNum(graph)
	for i := leaves; i < leaves+dec.Dec_GraphNodeNum(graph); i++ {
		a, b := dec.Dec_GraphNode(graph, i).Edges()
		la, lb := read(a.Node), read(b.Node)
		if a.Complemented() {
			la ^= 1
		}
		if b.Complemented() {
			lb ^= 1
		}
		dec.Dec_GraphSetNodeFunc(graph, i, Gia_ManHashAnd(p, la, lb))
	}
	value := read(root.Node)
	if root.Complemented() {
		value ^= 1
	}
	return value
}

// Gia_ManSopToAig builds a hashed AIG directly from a PLA-format cover.
func Gia_ManSopToAig(p *Gia_Man_t, sop string, leaves *Vec_Int_t) int32 {
	n := kit.Kit_PlaGetVarNum(sop)
	if n < 0 || int(n) > Vec_IntSize(leaves) {
		panic("gia: invalid SOP leaf count")
	}
	if end := strings.IndexByte(sop, 0); end >= 0 {
		sop = sop[:end]
	}
	sum := int32(0)
	for _, row := range strings.Split(strings.TrimSuffix(sop, "\n"), "\n") {
		if len(row) != int(n)+2 || row[n] != ' ' {
			panic("gia: malformed SOP cube")
		}
		product := int32(1)
		for i := int32(0); i < n; i++ {
			lit := Vec_IntEntry(leaves, int(i))
			if lit < 0 {
				panic("gia: invalid SOP leaf")
			}
			switch row[i] {
			case '0':
				product = Gia_ManHashAnd(p, product, lit^1)
			case '1':
				product = Gia_ManHashAnd(p, product, lit)
			case '-':
			default:
				panic("gia: malformed SOP literal")
			}
		}
		sum = Gia_ManHashOr(p, sum, product)
	}
	if kit.Kit_PlaIsComplement(sop) != 0 {
		sum ^= 1
	}
	return sum
}

func Gia_ManFactorGraph(p *Gia_Man_t, graph *dec.Dec_Graph_t, leaves *Vec_Int_t) int32 {
	for i := int32(0); i < dec.Dec_GraphLeaveNum(graph); i++ {
		dec.Dec_GraphSetLeafFunc(graph, i, Vec_IntEntry(leaves, int(i)))
	}
	return Gia_ManGraphToAig(p, graph)
}

func Gia_ManFactorNode(p *Gia_Man_t, sop string, leaves *Vec_Int_t) int32 {
	n := kit.Kit_PlaGetVarNum(sop)
	if n == 0 {
		return 1 ^ kit.Kit_PlaIsConst0(sop)
	}
	if n < 0 || int(n) != Vec_IntSize(leaves) {
		panic("gia: invalid SOP leaf count")
	}
	if n > 2 && kit.Kit_PlaGetCubeNum(sop) > 1 {
		return Gia_ManFactorGraph(p, dec.Dec_Factor(sop), leaves)
	}
	return Gia_ManSopToAig(p, sop, leaves)
}
