package csw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func Csw_Sweep(graph *aig.Graph, nCutsMax, nLeafMax, verbose int32) *aig.Graph {
	start := util.Abc_Clock()
	p := Csw_ManStart(graph, nCutsMax, nLeafMax, verbose)
	inputs := graph.Inputs()
	for i, ci := range p.pManRes.Inputs() {
		Csw_ObjPrepareCuts(p, ci, 1)
		Csw_ObjAddRefs(p, ci, aig.Aig_ObjRefs(graph, inputs[i]))
	}
	for _, obj := range graph.TopologicalObjects() {
		node := graph.Node(obj)
		switch node.Kind {
		case aig.Buffer:
			child := Csw_ObjChild0Equiv(p, obj)
			e := aig.Aig_ObjCreate(p.pManRes, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{child, aig.InvalidEdge}})
			Csw_ObjPrepareCuts(p, e, 1)
			Csw_ObjAddRefs(p, e, int32(node.Refs))
			Csw_ObjSetEquiv(p, obj, e)
		case aig.Output:
			Csw_ObjSetEquiv(p, obj, Csw_ObjChild0Equiv(p, obj))
		case aig.And, aig.ExclusiveOr:
			a, b := Csw_ObjChild0Equiv(p, obj), Csw_ObjChild1Equiv(p, obj)
			var e aig.Edge
			if node.Kind == aig.ExclusiveOr {
				e = p.pManRes.NewXor(a, b)
			} else {
				e = p.pManRes.AND(a, b)
			}
			trivial := int32(0)
			if node.Refs > 1 {
				trivial = 1
			}
			for {
				regular := e &^ 1
				result := Csw_ObjSweep(p, regular, trivial)
				if e.Complemented() {
					result = result.Not()
				}
				e = result
				if e&^1 == p.pManRes.Const1() || Csw_ObjCuts(p, e) != nil {
					break
				}
			}
			Csw_ObjSetEquiv(p, obj, e)
			Csw_ObjAddRefs(p, e, int32(node.Refs))
		}
	}
	for _, obj := range graph.Outputs() {
		e := Csw_ObjChild0Equiv(p, obj)
		if obj.Complemented() {
			e = e.Not()
		}
		p.pManRes.AddOutput(e)
	}
	aig.Aig_ManSetRegNum(p.pManRes, graph.RegisterCount())
	aig.Aig_ManSetConstrNum(p.pManRes, aig.Aig_ManConstrNum(graph))
	p.pManRes.Assertions = graph.Assertions
	p.pManRes.BarrierBuffers = graph.BarrierBuffers
	aig.Aig_ManCleanup(p.pManRes)
	p.timeTotal = util.Abc_Clock() - start
	p.timeOther = p.timeTotal - p.timeCuts - p.timeHash
	result := p.pManRes
	Csw_ManStop(p)
	return result
}
