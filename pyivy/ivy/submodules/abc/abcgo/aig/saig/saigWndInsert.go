package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_ManWindowInsertNodes replaces an outlined portion of g with window.
// The window's primary input and output order must match the outline's cuts.
func Saig_ManWindowInsertNodes(g *aig.Graph, nodes *vec.Vec_Ptr_t, window *aig.Graph) *aig.Graph {
	if g == nil || nodes == nil || window == nil {
		return nil
	}
	cutInputs := Saig_ManWindowCollectPis(g, nodes)
	cutOutputs := Saig_ManWindowCollectPos(g, nodes, nil)
	if vec.Vec_PtrSize(cutInputs) != int(Saig_ManPiNum(window)) || vec.Vec_PtrSize(cutOutputs) != int(Saig_ManPoNum(window)) {
		return nil
	}
	q := aig.NewGraph()
	q.Name, q.Spec = g.Name, g.Spec
	bigCopies := map[aig.Edge]aig.Edge{g.Const1(): q.Const1()}
	smallCopies := map[aig.Edge]aig.Edge{window.Const1(): q.Const1()}
	smallInputs := make(map[aig.Edge]aig.Edge)
	bigOutputs := make(map[aig.Edge]aig.Edge)
	for i, item := range vec.Vec_PtrArray(cutInputs) {
		smallInputs[aig.Aig_ManCi(window, int32(i))] = item.(aig.Edge)
	}
	for i, item := range vec.Vec_PtrArray(cutOutputs) {
		bigOutputs[item.(aig.Edge)] = aig.Aig_ObjChild0(window, aig.Aig_ManCo(window, int32(i)))
	}
	for _, e := range g.Inputs() {
		if Saig_ObjIsPi(g, e) || !aig.Aig_ObjIsTravIdCurrent(g, e) {
			bigCopies[e] = q.NewInput()
		}
	}
	for i := int32(0); i < Saig_ManRegNum(window); i++ {
		smallCopies[Saig_ManLo(window, i)] = q.NewInput()
	}
	activeBig, activeSmall := map[aig.Edge]bool{}, map[aig.Edge]bool{}
	var copyBig, copySmall func(aig.Edge) aig.Edge
	copyBig = func(e aig.Edge) aig.Edge {
		reg := aig.Aig_Regular(e)
		if c, ok := bigCopies[reg]; ok {
			return aig.Aig_NotCond(c, int32(e&1))
		}
		if activeBig[reg] {
			panic("saig: cyclic window insertion")
		}
		activeBig[reg] = true
		var c aig.Edge
		if root, ok := bigOutputs[reg]; ok {
			c = copySmall(root)
		} else {
			n := g.Node(reg)
			switch n.Kind {
			case aig.And:
				c = q.AND(copyBig(n.Fanins[0]), copyBig(n.Fanins[1]))
			case aig.ExclusiveOr:
				c = q.NewXor(copyBig(n.Fanins[0]), copyBig(n.Fanins[1]))
			case aig.Buffer, aig.Output:
				c = copyBig(n.Fanins[0])
			default:
				panic("saig: unmapped original window input")
			}
		}
		bigCopies[reg] = c
		delete(activeBig, reg)
		return aig.Aig_NotCond(c, int32(e&1))
	}
	copySmall = func(e aig.Edge) aig.Edge {
		reg := aig.Aig_Regular(e)
		if c, ok := smallCopies[reg]; ok {
			return aig.Aig_NotCond(c, int32(e&1))
		}
		if activeSmall[reg] {
			panic("saig: cyclic window insertion")
		}
		activeSmall[reg] = true
		var c aig.Edge
		if root, ok := smallInputs[reg]; ok {
			c = copyBig(root)
		} else {
			n := window.Node(reg)
			switch n.Kind {
			case aig.And:
				c = q.AND(copySmall(n.Fanins[0]), copySmall(n.Fanins[1]))
			case aig.ExclusiveOr:
				c = q.NewXor(copySmall(n.Fanins[0]), copySmall(n.Fanins[1]))
			case aig.Buffer, aig.Output:
				c = copySmall(n.Fanins[0])
			default:
				panic("saig: unmapped replacement window input")
			}
		}
		smallCopies[reg] = c
		delete(activeSmall, reg)
		return aig.Aig_NotCond(c, int32(e&1))
	}
	for i := int32(0); i < Saig_ManPoNum(g); i++ {
		po := aig.Aig_ManCo(g, i)
		q.AddOutput(copyBig(aig.Aig_ObjChild0(g, po)))
	}
	retained := int32(0)
	for i := int32(0); i < Saig_ManRegNum(g); i++ {
		li := Saig_ManLi(g, i)
		if !aig.Aig_ObjIsTravIdCurrent(g, li) {
			q.AddOutput(copyBig(aig.Aig_ObjChild0(g, li)))
			retained++
		}
	}
	for i := int32(0); i < Saig_ManRegNum(window); i++ {
		li := Saig_ManLi(window, i)
		q.AddOutput(copySmall(aig.Aig_ObjChild0(window, li)))
	}
	aig.Aig_ManSetRegNum(q, retained+Saig_ManRegNum(window))
	aig.Aig_ManSetConstrNum(q, aig.Aig_ManConstrNum(g))
	aig.Aig_ManCleanup(q)
	return q
}
func Saig_ManWindowInsert(g *aig.Graph, pivot aig.Edge, distance int32, window *aig.Graph) *aig.Graph {
	if g == nil || window == nil {
		return nil
	}
	nodes := Saig_ManWindowOutline(g, pivot, distance)
	return Saig_ManWindowInsertNodes(g, nodes, window)
}
func Saig_ManWindowTest(g *aig.Graph) *aig.Graph {
	pivot := Saig_ManFindPivot(g)
	if pivot == aig.InvalidEdge {
		return nil
	}
	window := Saig_ManWindowExtract(g, pivot, 3)
	return Saig_ManWindowInsert(g, pivot, 3, window)
}
