package ssw

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

func Ssw_ReportConeReductions(p *Ssw_Man_t, before, after *aig.Graph) {
	first := aig.Aig_ManDupOneOutput(before, 0, 1)
	reducedFirst := aig.Aig_ManScl(first, 1, 1, 0, -1, -1, 0, 0)
	aig.Aig_ManStop(first)
	second := aig.Aig_ManDupOneOutput(after, 0, 1)
	reducedSecond := aig.Aig_ManScl(second, 1, 1, 0, -1, -1, 0, 0)
	aig.Aig_ManStop(second)
	p.nNodesBegC = aig.Aig_ManNodeNum(reducedFirst)
	p.nNodesEndC = aig.Aig_ManNodeNum(reducedSecond)
	p.nRegsBegC = aig.Aig_ManRegNum(reducedFirst)
	p.nRegsEndC = aig.Aig_ManRegNum(reducedSecond)
	aig.Aig_ManStop(reducedFirst)
	aig.Aig_ManStop(reducedSecond)
}
func Ssw_ReportOneOutput(g *aig.Graph, edge aig.Edge) {
	if edge == aig.Aig_ManConst1(g) {
		fmt.Print("1")
	} else if edge == aig.Aig_ManConst0(g) {
		fmt.Print("0")
	} else {
		fmt.Print("X")
	}
}
func Ssw_ReportOutputs(g *aig.Graph) {
	n := saig.Saig_ManPoNum(g)
	constraints := saig.Saig_ManConstrNum(g)
	for i := int32(0); i < n; i++ {
		if i < n-constraints {
			fmt.Print("o")
		} else {
			fmt.Print("c")
		}
		Ssw_ReportOneOutput(g, aig.Aig_ObjChild0(g, aig.Aig_ManCo(g, i)))
	}
	fmt.Println()
}
func Ssw_ManUpdateEquivs(p *Ssw_Man_t, g *aig.Graph, verbose int32) {
	n := saig.Saig_ManConstrNum(g)
	pos := saig.Saig_ManPoNum(g)
	roots := g.Outputs()[pos-n : pos]
	cone := aig.Aig_ManDfsNodes(g, roots, n)
	// Go's DFS returns nodes but does not mark traversal IDs as the C DFS did.
	seen := make(map[int]bool)
	var visit func(aig.Edge)
	visit = func(edge aig.Edge) {
		id := edge.NodeID()
		if seen[id] {
			return
		}
		seen[id] = true
		node := g.Node(edge)
		switch node.Kind {
		case aig.And, aig.ExclusiveOr:
			visit(node.Fanins[0])
			visit(node.Fanins[1])
		case aig.Buffer, aig.Output:
			visit(node.Fanins[0])
		}
	}
	for _, root := range roots {
		visit(root)
	}
	var total, removed int32
	for _, obj := range g.LiveEdges() {
		if !aig.Aig_ObjIsCi(g, obj) && !aig.Aig_ObjIsNode(g, obj) {
			continue
		}
		repr := aig.Aig_ObjRepr(g, obj)
		if repr != aig.InvalidEdge {
			total++
		}
		if repr != aig.InvalidEdge && seen[obj.NodeID()] && p.pPars.FConstrs != 0 && p.pPars.FMergeFull == 0 {
			aig.Aig_ObjSetRepr(g, obj, aig.InvalidEdge)
			removed++
		}
	}
	p.nConesTotal = aig.Aig_ManCiNum(g) + aig.Aig_ManNodeNum(g)
	p.nConesConstr = int32(len(cone))
	p.nEquivsTotal = total
	p.nEquivsConstr = removed
}
