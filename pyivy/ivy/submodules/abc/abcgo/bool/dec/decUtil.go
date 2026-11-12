package dec

// Dec_GraphDeriveTruth evaluates a C-style graph through the native Go graph
// evaluator. Temporary truth values no longer masquerade as pointer addresses.
func Dec_GraphDeriveTruth(g *Dec_Graph_t) uint32 {
	return g.Native().Truth32()
}

func (g *Dec_Graph_t) Native() *Graph {
	native := &Graph{LeafCount: int(g.nLeaves), Constant: g.fConst != 0, Root: Edge{int(g.eRoot.Node), g.eRoot.fCompl != 0}, Nodes: make([]Node, len(g.pNodes))}
	for i, n := range g.pNodes {
		native.Nodes[i] = Node{Edge{int(n.eEdge0.Node), n.eEdge0.fCompl != 0}, Edge{int(n.eEdge1.Node), n.eEdge1.fCompl != 0}}
	}
	return native
}
