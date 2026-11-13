package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_ManWindowExtractNodes turns cut signals into primary inputs and keeps
// latch pairs contained in the outlined window as registers.
func Saig_ManWindowExtractNodes(g *aig.Graph, nodes *vec.Vec_Ptr_t) *aig.Graph {
	if g == nil || nodes == nil {
		return nil
	}
	q := aig.NewGraph()
	q.Name = "wnd"
	mapping := map[aig.Edge]aig.Edge{g.Const1(): q.Const1()}
	inputs := Saig_ManWindowCollectPis(g, nodes)
	for _, item := range vec.Vec_PtrArray(inputs) {
		e := item.(aig.Edge)
		mapping[e] = q.NewInput()
	}
	var regs []aig.Edge
	for _, item := range vec.Vec_PtrArray(nodes) {
		e := item.(aig.Edge)
		if Saig_ObjIsLo(g, e) {
			mapping[e] = q.NewInput()
			regs = append(regs, e)
		}
	}
	outputs := Saig_ManWindowCollectPos(g, nodes, nil)
	for _, item := range vec.Vec_PtrArray(outputs) {
		q.AddOutput(temporCopy(q, g, mapping, item.(aig.Edge)))
	}
	for _, lo := range regs {
		li := Saig_ObjLoToLi(g, lo)
		q.AddOutput(temporCopy(q, g, mapping, aig.Aig_ObjChild0(g, li)))
	}
	aig.Aig_ManSetRegNum(q, int32(len(regs)))
	aig.Aig_ManCleanup(q)
	return q
}

func Saig_ManWindowExtract(g *aig.Graph, pivot aig.Edge, distance int32) *aig.Graph {
	if g == nil {
		return nil
	}
	return Saig_ManWindowExtractNodes(g, Saig_ManWindowOutline(g, pivot, distance))
}

func Saig_ManFindPivot(g *aig.Graph) aig.Edge {
	if g == nil {
		return aig.InvalidEdge
	}
	if Saig_ManRegNum(g) > 0 {
		return Saig_ManLo(g, Saig_ManRegNum(g)/2)
	}
	var gates []aig.Edge
	for _, e := range g.LiveEdges() {
		if aig.Aig_ObjIsNode(g, e) {
			gates = append(gates, e)
		}
	}
	if len(gates) == 0 {
		return aig.InvalidEdge
	}
	return gates[len(gates)/2]
}
