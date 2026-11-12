package abc

import "github.com/glycerine/abcgo/aig/aig"

// abcStrashLocal converts one local Boolean function to the destination AIG.
func abcStrashLocal(dst *Network, node *Object) Edge {
	g := aig.NewGraph()
	args := make([]aig.Edge, len(node.fanins))
	for i := range args {
		args[i] = g.NewInput()
	}
	root := localFunctionAIG(g, node, args)
	copies := make([]Edge, len(g.Nodes()))
	copies[0] = dst.One()
	for i, e := range args {
		copies[e.NodeID()] = ntkCopiedEdge(node.fanins[i])
	}
	edge := func(e aig.Edge) Edge {
		out := copies[e.NodeID()]
		if e.Complemented() {
			out = out.Not()
		}
		return out
	}
	manager := dst.FunctionManager.(*Abc_Aig_t)
	for id, entry := range g.Nodes() {
		switch entry.Kind {
		case aig.Constant, aig.Input:
		case aig.And:
			copies[id] = Abc_AigAnd(manager, edge(entry.Fanins[0]), edge(entry.Fanins[1]))
		case aig.ExclusiveOr:
			copies[id] = Abc_AigXor(manager, edge(entry.Fanins[0]), edge(entry.Fanins[1]))
		case aig.Buffer:
			copies[id] = edge(entry.Fanins[0])
		default:
			panic("abc: unsupported local AIG node")
		}
	}
	return edge(root)
}

// Abc_NtkStrash builds a structurally hashed ABC network from local functions.
func Abc_NtkStrash(src *Network, all, cleanup, record int32) *Network {
	if src == nil || (Abc_NtkIsLogic(src) == 0 && Abc_NtkIsStrash(src) == 0) {
		panic("abc: expected logic or strashed network")
	}
	if Abc_NtkIsStrash(src) != 0 {
		return Abc_NtkDupDfs(src)
	}
	dst := Abc_NtkStartFrom(src, ABC_NTK_STRASH, ABC_FUNC_AIG)
	for _, node := range Abc_NtkDfsIter(src, all) {
		if isBarrierBuffer(node) {
			setCopy(node, ntkCopiedEdge(node.fanins[0]))
		} else {
			setCopy(node, abcStrashLocal(dst, node))
		}
	}
	Abc_NtkFinalize(src, dst)
	if cleanup != 0 {
		Abc_AigCleanup(dst.FunctionManager.(*Abc_Aig_t))
	}
	if src.Exdc != nil {
		dst.Exdc = Abc_NtkDup(src.Exdc)
	}
	return dst
}
