package kit

import "github.com/glycerine/abcgo/aig/hop"

type kitGiaManager interface {
	Const1() int32
	InputLiteral(int32) int32
	AndLiteral(int32, int32, bool) int32
}

func Kit_GraphToGiaInternal(manager kitGiaManager, g *Kit_Graph_t, hash int32) int32 {
	if g.fConst != 0 {
		return manager.Const1() ^ int32(g.eRoot.fCompl)
	}
	for i := g.nLeaves; i < int32(len(g.pNodes)); i++ {
		node := g.pNodes[i]
		a := g.pNodes[node.eEdge0.Node].iFunc ^ int32(node.eEdge0.fCompl)
		b := g.pNodes[node.eEdge1.Node].iFunc ^ int32(node.eEdge1.fCompl)
		node.iFunc = manager.AndLiteral(a, b, hash != 0)
	}
	return g.pNodes[g.eRoot.Node].iFunc ^ int32(g.eRoot.fCompl)
}
func Kit_GraphToGia(manager kitGiaManager, g *Kit_Graph_t, leaves []int32, hash int32) int32 {
	if leaves != nil && len(leaves) < int(g.nLeaves) {
		panic("kit: insufficient GIA leaves")
	}
	for i := int32(0); i < g.nLeaves; i++ {
		if leaves == nil {
			g.pNodes[i].iFunc = manager.InputLiteral(i)
		} else {
			g.pNodes[i].iFunc = leaves[i]
		}
	}
	return Kit_GraphToGiaInternal(manager, g, hash)
}
func Kit_TruthToGia(manager kitGiaManager, truth []uint32, nVars int32, memory *[]uint32, leaves []int32, hash int32) int32 {
	g := Kit_TruthToGraph(truth, nVars, memory)
	if g == nil {
		panic("kit: truth-to-GIA factoring failed")
	}
	return Kit_GraphToGia(manager, g, leaves, hash)
}
func Kit_TruthToGia2(manager kitGiaManager, off, on []uint32, nVars int32, memory *[]uint32, leaves []int32, hash int32) int32 {
	g := Kit_TruthToGraph2(off, on, nVars, memory)
	if g == nil {
		panic("kit: care-set truth-to-GIA factoring failed")
	}
	return Kit_GraphToGia(manager, g, leaves, hash)
}
func Kit_GraphToHopInternal(manager *hop.Hop_Man_t, g *Kit_Graph_t) *hop.Hop_Obj_t {
	if g.fConst != 0 {
		return hop.Hop_NotCond(hop.Hop_ManConst1(manager), int32(g.eRoot.fCompl))
	}
	for i := g.nLeaves; i < int32(len(g.pNodes)); i++ {
		node := g.pNodes[i]
		a, ok := g.pNodes[node.eEdge0.Node].pFunc.(*hop.Hop_Obj_t)
		if !ok {
			panic("kit: missing first Hop fanin")
		}
		b, ok := g.pNodes[node.eEdge1.Node].pFunc.(*hop.Hop_Obj_t)
		if !ok {
			panic("kit: missing second Hop fanin")
		}
		node.pFunc = hop.Hop_And(manager, hop.Hop_NotCond(a, int32(node.eEdge0.fCompl)), hop.Hop_NotCond(b, int32(node.eEdge1.fCompl)))
	}
	root, ok := g.pNodes[g.eRoot.Node].pFunc.(*hop.Hop_Obj_t)
	if !ok {
		panic("kit: missing Hop root")
	}
	return hop.Hop_NotCond(root, int32(g.eRoot.fCompl))
}
func Kit_GraphToHop(manager *hop.Hop_Man_t, g *Kit_Graph_t) *hop.Hop_Obj_t {
	for i := int32(0); i < g.nLeaves; i++ {
		g.pNodes[i].pFunc = hop.Hop_IthVar(manager, i)
	}
	return Kit_GraphToHopInternal(manager, g)
}
func Kit_TruthToHop(manager *hop.Hop_Man_t, truth []uint32, nVars int32, memory *[]uint32) *hop.Hop_Obj_t {
	g := Kit_TruthToGraph(truth, nVars, memory)
	if g == nil {
		panic("kit: truth-to-Hop factoring failed")
	}
	return Kit_GraphToHop(manager, g)
}
func Kit_CoverToHop(manager *hop.Hop_Man_t, cover []uint32, nVars int32, memory *[]uint32) *hop.Hop_Obj_t {
	if memory != nil {
		*memory = (*memory)[:0]
	}
	return Kit_GraphToHop(manager, Kit_SopFactor(cover, 0, nVars, memory))
}
