package kit

// A manager interface keeps the computation kit independent of the AIG package:
// AIG canonicalization may itself depend on kit's decomposition algorithms.
type kitAigManager[E ~uint32] interface {
	Const1() E
	AND(E, E) E
}

func kitAigNotCond[E ~uint32](edge E, compl uint32) E {
	if compl != 0 {
		return edge ^ 1
	}
	return edge
}
func Kit_GraphToAigInternal[E ~uint32](manager kitAigManager[E], graph *Kit_Graph_t) E {
	if graph.fConst != 0 {
		return kitAigNotCond(manager.Const1(), graph.eRoot.fCompl)
	}
	for i := graph.nLeaves; i < int32(len(graph.pNodes)); i++ {
		node := graph.pNodes[i]
		a, ok := graph.pNodes[node.eEdge0.Node].pFunc.(E)
		if !ok {
			panic("kit: missing first AIG fanin")
		}
		b, ok := graph.pNodes[node.eEdge1.Node].pFunc.(E)
		if !ok {
			panic("kit: missing second AIG fanin")
		}
		a = kitAigNotCond(a, node.eEdge0.fCompl)
		b = kitAigNotCond(b, node.eEdge1.fCompl)
		node.pFunc = manager.AND(a, b)
	}
	result, ok := graph.pNodes[graph.eRoot.Node].pFunc.(E)
	if !ok {
		panic("kit: missing AIG root")
	}
	return kitAigNotCond(result, graph.eRoot.fCompl)
}
func Kit_GraphToAig[E ~uint32](manager kitAigManager[E], fanins []E, graph *Kit_Graph_t) E {
	if len(fanins) < int(graph.nLeaves) {
		panic("kit: insufficient AIG inputs")
	}
	for i := int32(0); i < graph.nLeaves; i++ {
		graph.pNodes[i].pFunc = fanins[i]
	}
	return Kit_GraphToAigInternal(manager, graph)
}
func Kit_TruthToAig[E ~uint32](manager kitAigManager[E], fanins []E, truth []uint32, nVars int32, memory *[]uint32) E {
	graph := Kit_TruthToGraph(truth, nVars, memory)
	if graph == nil {
		panic("kit: truth-table factoring failed")
	}
	return Kit_GraphToAig(manager, fanins, graph)
}
